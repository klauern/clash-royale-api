package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/klauer/clash-royale-api/go/pkg/fuzzstorage"
	"github.com/klauer/clash-royale-api/go/pkg/leaderboard"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v3"
)

type stageCanceler struct {
	mu     sync.Mutex
	cancel context.CancelFunc
}

func (sc *stageCanceler) Set(cancel context.CancelFunc) {
	sc.mu.Lock()
	sc.cancel = cancel
	sc.mu.Unlock()
}

func (sc *stageCanceler) Clear() {
	sc.Set(nil)
}

func (sc *stageCanceler) Cancel() {
	sc.mu.Lock()
	cancel := sc.cancel
	sc.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// deckFuzzCommand is the action function for the deck fuzz command
func deckFuzzCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")
	count := cmd.Int("count")
	workers := cmd.Int("workers")
	// Auto-detect CPU count if workers is at default value
	if workers == 1 {
		workers = runtime.NumCPU()
		fprintf(os.Stderr, "Auto-detected %d CPU cores, using %d workers\n", runtime.NumCPU(), workers)
	}
	includeCards := cmd.StringSlice("include-cards")
	excludeCards := cmd.StringSlice("exclude-cards")
	includeFromSaved := cmd.Int("include-from-saved")
	fromSaved := cmd.Int("from-saved")
	resumeFrom := cmd.Int("resume-from")
	basedOn := cmd.String("based-on")
	minElixir := cmd.Float64("min-elixir")
	maxElixir := cmd.Float64("max-elixir")
	minOverall := cmd.Float64("min-overall")
	minSynergy := cmd.Float64("min-synergy")
	top := cmd.Int("top")
	sortBy := cmd.String("sort-by")
	format := cmd.String("format")
	outputDir := cmd.String("output-dir")
	verbose := cmd.Bool("verbose")
	fromAnalysis := cmd.Bool("from-analysis")
	apiToken := cmd.String("api-token")
	storagePath := cmd.String("storage")
	saveTop := cmd.Bool("save-top")
	synergyPairs := cmd.Bool("synergy-pairs")
	evolutionCentric := cmd.Bool("evolution-centric")
	minEvoCards := cmd.Int("min-evo-cards")
	minEvoLevel := cmd.Int("min-evo-level")
	evoWeight := cmd.Float64("evo-weight")

	var interrupted atomic.Bool
	var canceler stageCanceler

	interrupts := make(chan os.Signal, 2)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupts)

	go func() {
		<-interrupts
		if interrupted.CompareAndSwap(false, true) {
			fprintln(os.Stderr, "\nInterrupt received; stopping current stage and saving partial results (press Ctrl+C again to exit immediately)")
			canceler.Cancel()
		}
		<-interrupts
		fprintln(os.Stderr, "\nSecond interrupt received; exiting immediately.")
		os.Exit(130)
	}()

	// Validate flags
	if playerTag == "" && !fromAnalysis {
		return fmt.Errorf("--tag is required (or use --from-analysis for offline mode)")
	}

	if minOverall < 0 || minOverall > 10 {
		return fmt.Errorf("--min-overall must be between 0 and 10")
	}

	if minSynergy < 0 || minSynergy > 10 {
		return fmt.Errorf("--min-synergy must be between 0 and 10")
	}

	if top < 1 {
		return fmt.Errorf("--top must be at least 1")
	}

	// Validate sort-by field
	validSortFields := map[string]bool{
		"overall":     true,
		"attack":      true,
		"defense":     true,
		"synergy":     true,
		"versatility": true,
		"elixir":      true,
	}
	if !validSortFields[sortBy] {
		return fmt.Errorf("invalid --sort-by value: %s (must be one of: overall, attack, defense, synergy, versatility, elixir)", sortBy)
	}

	// Validate format
	validFormats := map[string]bool{
		"summary":  true,
		"json":     true,
		"csv":      true,
		"detailed": true,
	}
	if !validFormats[format] {
		return fmt.Errorf("invalid --format value: %s (must be one of: summary, json, csv, detailed)", format)
	}

	var player *clashroyale.Player
	var playerName string
	var err error

	// Load player data
	if fromAnalysis {
		// Load from existing analysis file
		analysisFile := cmd.String("analysis-file")
		analysisDir := cmd.String("analysis-dir")

		if analysisFile == "" && analysisDir == "" {
			return fmt.Errorf("--analysis-file or --analysis-dir required when using --from-analysis")
		}

		player, playerName, err = loadPlayerFromAnalysis(analysisFile, analysisDir, playerTag)
		if err != nil {
			return fmt.Errorf("failed to load player from analysis: %w", err)
		}
	} else {
		// Load from API
		if apiToken == "" {
			apiToken = os.Getenv("CLASH_ROYALE_API_TOKEN")
		}
		if apiToken == "" {
			return fmt.Errorf("--api-token or CLASH_ROYALE_API_TOKEN environment variable required")
		}

		client := clashroyale.NewClient(apiToken)
		cleanTag := strings.TrimPrefix(playerTag, "#")

		player, err = client.GetPlayer(cleanTag)
		if err != nil {
			return fmt.Errorf("failed to fetch player: %w", err)
		}
		playerName = player.Name
	}

	if verbose {
		fprintf(os.Stderr, "Loaded player: %s (%s)\n", playerName, player.Tag)
		fprintf(os.Stderr, "Cards available: %d\n", len(player.Cards))
	}

	// Initialize fuzzer configuration
	fuzzerCfg := &deck.FuzzingConfig{
		Count:             count,
		Workers:           workers,
		IncludeCards:      includeCards,
		ExcludeCards:      excludeCards,
		MinAvgElixir:      minElixir,
		MaxAvgElixir:      maxElixir,
		MinOverallScore:   minOverall,
		MinSynergyScore:   minSynergy,
		SynergyFirst:      synergyPairs,
		EvolutionCentric:  evolutionCentric,
		MinEvolutionCards: minEvoCards,
		MinEvoLevel:       minEvoLevel,
		EvoWeight:         evoWeight,
	}

	// Handle --include-from-saved: extract cards from saved top decks
	if includeFromSaved > 0 {
		savedCards, err := loadCardsFromSavedDecks(includeFromSaved, verbose)
		if err != nil {
			return fmt.Errorf("failed to load cards from saved decks: %w", err)
		}
		// Merge with existing include cards (avoiding duplicates)
		fuzzerCfg.IncludeCards = mergeUniqueCards(fuzzerCfg.IncludeCards, savedCards)
		if verbose && len(savedCards) > 0 {
			fprintf(os.Stderr, "Included %d cards from saved top decks\n", len(savedCards))
		}
	}

	// Set seed for reproducibility if specified
	if seed := cmd.Int("seed"); seed != 0 {
		fuzzerCfg.Seed = int64(seed)
	}

	// Create fuzzer
	fuzzer, err := deck.NewDeckFuzzer(player, fuzzerCfg)
	if err != nil {
		return fmt.Errorf("failed to create fuzzer: %w", err)
	}

	// Runtime estimation for large batches
	const sampleSize = 1000
	var estimate *runtimeEstimate
	if count > sampleSize && verbose {
		estimate, err = estimateRuntime(fuzzer, count, sampleSize)
		if err != nil {
			fprintf(os.Stderr, "Warning: could not estimate runtime: %v\n", err)
			// Continue anyway, estimation is optional
		} else {
			formattedTime := formatDuration(estimate.totalSeconds)
			decksPerSec := int(estimate.decksPerSec)
			fprintf(os.Stderr, "Estimated time: ~%s for %d decks (~%d decks/sec)\n",
				formattedTime, count, decksPerSec)

			confirmed, err := confirmAction("Continue? (y/n) ")
			if err != nil {
				return fmt.Errorf("confirmation failed: %w", err)
			}
			if !confirmed {
				fprintln(os.Stderr, "Aborted.")
				return nil
			}
		}
	}

	if verbose {
		fprintf(os.Stderr, "\nStarting deck fuzzing...\n")
		fprintf(os.Stderr, "Configuration:\n")
		fprintf(os.Stderr, "  Count: %d\n", count)
		fprintf(os.Stderr, "  Workers: %d\n", workers)
		if synergyPairs {
			fprintf(os.Stderr, "  Mode: synergy-first (4 pairs)\n")
		}
		if evolutionCentric {
			fprintf(os.Stderr, "  Mode: evolution-centric (min %d evo cards, level %d+)\n", minEvoCards, minEvoLevel)
		}
		if len(includeCards) > 0 {
			fprintf(os.Stderr, "  Include cards: %s\n", strings.Join(includeCards, ", "))
		}
		if len(excludeCards) > 0 {
			fprintf(os.Stderr, "  Exclude cards: %s\n", strings.Join(excludeCards, ", "))
		}
		if basedOn != "" {
			fprintf(os.Stderr, "  Based on deck: %s\n", basedOn)
		}
		fprintf(os.Stderr, "  Elixir range: %.1f - %.1f\n", minElixir, maxElixir)
		fprintf(os.Stderr, "  Min overall score: %.1f\n", minOverall)
		fprintf(os.Stderr, "  Min synergy score: %.1f\n", minSynergy)
		fprintf(os.Stderr, "\n")
	}

	// Generate decks
	startTime := time.Now()

	// Start progress reporter for generation
	var generationDone sync.WaitGroup
	stopProgress := make(chan struct{})
	if verbose {
		generationDone.Add(1)
		go func() {
			defer generationDone.Done()
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()

			lastCount := 0
			startTime := time.Now()

			for {
				select {
				case <-stopProgress:
					return
				case <-ticker.C:
					stats := fuzzer.GetStats()
					currentCount := stats.Generated
					elapsed := time.Since(startTime)

					// Calculate rate
					rate := float64(currentCount) / elapsed.Seconds()

					// Only print if progress has been made
					if currentCount > lastCount {
						eta := time.Duration(float64(count-currentCount)/rate) * time.Second
						fprintf(os.Stderr, "\rGenerating... %d/%d decks (%.1f decks/sec, ETA: %v) ",
							currentCount, count, rate, eta.Round(time.Second))
						lastCount = currentCount
					}
				}
			}
		}()
	}

	// Handle --resume-from: load top N saved decks as initial seed population
	var seedDecks [][]string
	if resumeFrom > 0 && !interrupted.Load() {
		savedDecks, err := loadSavedDecksForSeeding(resumeFrom, player, verbose)
		if err != nil {
			return fmt.Errorf("failed to load saved decks for resume: %w", err)
		}
		if len(savedDecks) > 0 {
			seedDecks = savedDecks
			if verbose {
				fprintf(os.Stderr, "Loaded %d saved decks as initial seed population\n", len(savedDecks))
			}
		}
	}

	generationCtx, cancelGeneration := context.WithCancel(ctx)
	canceler.Set(cancelGeneration)

	var generatedDecks [][]string
	if workers > 1 {
		generatedDecks, err = fuzzer.GenerateDecksParallelWithContext(generationCtx)
	} else {
		generatedDecks, err = fuzzer.GenerateDecksWithContext(generationCtx, count)
	}

	// Prepend seed decks to generated decks
	if len(seedDecks) > 0 {
		generatedDecks = append(seedDecks, generatedDecks...)
	}
	canceler.Clear()
	cancelGeneration()
	if err != nil && !(interrupted.Load() && errors.Is(err, context.Canceled)) {
		return fmt.Errorf("failed to generate decks: %w", err)
	}

	// Handle --from-saved: add mutations of saved decks
	if fromSaved > 0 && !interrupted.Load() {
		savedDecks, err := loadSavedDecksForSeeding(fromSaved, player, verbose)
		if err != nil {
			return fmt.Errorf("failed to load saved decks for seeding: %w", err)
		}
		if len(savedDecks) > 0 {
			mutations := generateDeckMutations(savedDecks, player, count, verbose)
			generatedDecks = append(generatedDecks, mutations...)
			if verbose {
				fprintf(os.Stderr, "Added %d mutations from %d saved decks\n", len(mutations), len(savedDecks))
			}
		}
	}

	// Handle --based-on: load a specific deck and generate variations
	if basedOn != "" && !interrupted.Load() {
		baseDeck, err := loadDeckFromStorage(basedOn, verbose)
		if err != nil {
			return fmt.Errorf("failed to load deck from storage: %w", err)
		}
		variations := generateVariations(baseDeck, player, count, verbose)
		if len(variations) > 0 {
			generatedDecks = append(generatedDecks, variations...)
			if verbose {
				fprintf(os.Stderr, "Added %d variations based on deck: %s\n", len(variations), strings.Join(baseDeck, ", "))
			}
		}
	}

	// Stop progress reporter
	close(stopProgress)
	generationDone.Wait()
	fprintln(os.Stderr) // New line after progress

	generationTime := time.Since(startTime)

	stats := fuzzer.GetStats()

	if verbose {
		fprintf(os.Stderr, "\nGenerated %d decks in %v (%.1f decks/sec)\n",
			len(generatedDecks), generationTime.Round(time.Millisecond),
			float64(len(generatedDecks))/generationTime.Seconds())
		fprintf(os.Stderr, "Success: %d, Failed: %d\n", stats.Success, stats.Failed)
		if stats.SkippedElixir > 0 {
			fprintf(os.Stderr, "Skipped (elixir): %d\n", stats.SkippedElixir)
		}
		fprintf(os.Stderr, "\n")
	}

	if len(generatedDecks) == 0 {
		if interrupted.Load() {
			fprintln(os.Stderr, "\nInterrupted before any decks were generated.")
			return nil
		}
		return fmt.Errorf("no decks were successfully generated")
	}

	// Evaluate decks
	if verbose {
		fprintf(os.Stderr, "Evaluating %d decks with %d workers...\n", len(generatedDecks), workers)
	}

	evaluationCtx, cancelEvaluation := context.WithCancel(ctx)
	canceler.Set(cancelEvaluation)
	evaluationResults, evalErr := evaluateGeneratedDecks(
		evaluationCtx,
		generatedDecks,
		player,
		playerTag,
		storagePath,
		workers,
		verbose,
	)
	canceler.Clear()
	cancelEvaluation()
	if evalErr != nil && !(interrupted.Load() && errors.Is(evalErr, context.Canceled)) {
		return fmt.Errorf("failed to evaluate decks: %w", evalErr)
	}
	if len(evaluationResults) == 0 {
		if interrupted.Load() {
			fprintln(os.Stderr, "\nInterrupted before any decks were evaluated.")
			return nil
		}
		return fmt.Errorf("no decks were evaluated")
	}

	// Filter by score thresholds
	filteredResults := filterResultsByScore(evaluationResults, minOverall, minSynergy, verbose)

	if len(filteredResults) == 0 {
		return fmt.Errorf("no decks passed the score filters (min-overall: %.1f, min-synergy: %.1f)", minOverall, minSynergy)
	}

	if verbose {
		fprintf(os.Stderr, "%d decks passed score filters\n", len(filteredResults))
	}

	// Deduplicate results (remove identical decks)
	dedupedResults := deduplicateResults(filteredResults)
	if verbose {
		fprintf(os.Stderr, "Removed %d duplicate decks, %d unique decks remaining\n", len(filteredResults)-len(dedupedResults), len(dedupedResults))
	}

	// Sort results
	sortFuzzingResults(dedupedResults, sortBy)

	// Get top N results
	topResults := getTopResults(dedupedResults, top)

	// Format and output results
	if err := formatFuzzingResults(topResults, format, playerName, playerTag, fuzzerCfg, generationTime, &stats, len(dedupedResults)); err != nil {
		return fmt.Errorf("failed to format results: %w", err)
	}

	// Save to file if output-dir specified
	if outputDir != "" {
		if err := saveResultsToFile(topResults, outputDir, format, playerTag); err != nil {
			return fmt.Errorf("failed to save results: %w", err)
		}
		if verbose {
			fprintf(os.Stderr, "\nResults saved to %s\n", outputDir)
		}
	}

	// Save top decks to persistent storage if requested
	if saveTop {
		if err := saveTopDecksToStorage(topResults, verbose); err != nil {
			return fmt.Errorf("failed to save top decks to storage: %w", err)
		}
	}

	return nil
}

// FuzzingResult represents a single fuzzing result with deck and evaluation
type FuzzingResult struct {
	Deck                []string
	OverallScore        float64
	AttackScore         float64
	DefenseScore        float64
	SynergyScore        float64
	VersatilityScore    float64
	AvgElixir           float64
	Archetype           string
	ArchetypeConfidence float64
	EvaluatedAt         time.Time
}

// evaluateGeneratedDecks evaluates a list of generated decks
func evaluateGeneratedDecks(
	ctx context.Context,
	decks [][]string,
	player *clashroyale.Player,
	playerTag string,
	storagePath string,
	workers int,
	verbose bool,
) ([]FuzzingResult, error) {
	// Create player context if player tag provided (shared, read-only)
	var playerContext *evaluation.PlayerContext
	if playerTag != "" && player != nil {
		playerContext = evaluation.NewPlayerContextFromPlayer(player)
	}

	var storage *leaderboard.Storage
	var storageErr error
	if storagePath != "" {
		storage, storageErr = leaderboard.NewStorage(storagePath)
		if storageErr != nil && verbose {
			fprintf(os.Stderr, "Warning: failed to open storage: %v\n", storageErr)
		}
		if storage != nil {
			defer closeFile(storage)
		}
	}

	// Use parallel evaluation if workers > 1
	if workers > 1 {
		return evaluateDecksParallel(ctx, decks, player, playerTag, playerContext, storage, workers, verbose)
	}

	// Sequential evaluation (original behavior)
	return evaluateDecksSequential(ctx, decks, player, playerTag, playerContext, storage, verbose)
}

// evaluateDecksSequential evaluates decks sequentially (original implementation)
func evaluateDecksSequential(
	ctx context.Context,
	decks [][]string,
	player *clashroyale.Player,
	playerTag string,
	playerContext *evaluation.PlayerContext,
	storage *leaderboard.Storage,
	verbose bool,
) ([]FuzzingResult, error) {
	results := make([]FuzzingResult, 0, len(decks))

	// Create synergy database once for sequential use
	synergyDB := deck.NewSynergyDatabase()

	// Create progress bar if verbose
	var bar *progressbar.ProgressBar
	if verbose {
		bar = progressbar.NewOptions(len(decks),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("decks"),
			progressbar.OptionOnCompletion(func() {
				fprintln(os.Stderr)
			}),
		)
	}

	// Evaluate each deck
	for _, deckCards := range decks {
		if err := ctx.Err(); err != nil {
			return results, err
		}
		result := evaluateSingleDeck(deckCards, player, playerTag, synergyDB, playerContext)
		results = append(results, result)

		// Save to persistent storage if available
		if storage != nil {
			saveDeckToStorage(result, playerTag, storage)
		}

		if verbose && bar != nil {
			if err := bar.Add(1); err != nil {
				return results, err
			}
		}
	}

	if err := ctx.Err(); err != nil {
		return results, err
	}

	return results, nil
}

// evaluateDecksParallel evaluates decks using parallel workers
func evaluateDecksParallel(
	ctx context.Context,
	decks [][]string,
	player *clashroyale.Player,
	playerTag string,
	playerContext *evaluation.PlayerContext,
	storage *leaderboard.Storage,
	workers int,
	verbose bool,
) ([]FuzzingResult, error) {
	results := make([]FuzzingResult, 0, len(decks))
	var wg sync.WaitGroup

	// Create work channel
	workChan := make(chan []string, len(decks))
	resultChan := make(chan FuzzingResult, len(decks))

	// Create progress bar if verbose
	var bar *progressbar.ProgressBar
	if verbose {
		bar = progressbar.NewOptions(len(decks),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("decks"),
			progressbar.OptionOnCompletion(func() {
				fprintln(os.Stderr)
			}),
		)
	}

	// Start workers
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each worker gets its own synergy database to avoid concurrent access
			synergyDB := deck.NewSynergyDatabase()

			for {
				select {
				case <-ctx.Done():
					return
				case deckCards, ok := <-workChan:
					if !ok {
						return
					}
					// Evaluate deck and send to result channel
					result := evaluateSingleDeck(deckCards, player, playerTag, synergyDB, playerContext)
					select {
					case <-ctx.Done():
						return
					case resultChan <- result:
					}
				}
			}
		}()
	}

	// Send work
	go func() {
		for _, deck := range decks {
			select {
			case <-ctx.Done():
				close(workChan)
				return
			case workChan <- deck:
			}
		}
		close(workChan)
	}()

	// Close result channel when workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results and update progress bar
	for result := range resultChan {
		results = append(results, result)

		if verbose && bar != nil {
			if err := bar.Add(1); err != nil {
				return results, err
			}
		}
	}

	// Save all results to storage after collection (storage may not be thread-safe)
	if storage != nil {
		for _, result := range results {
			saveDeckToStorage(result, playerTag, storage)
		}
	}

	if err := ctx.Err(); err != nil {
		return results, err
	}

	return results, nil
}

// evaluateSingleDeck evaluates a single deck and returns the result
func evaluateSingleDeck(
	deckCards []string,
	player *clashroyale.Player,
	playerTag string,
	synergyDB *deck.SynergyDatabase,
	playerContext *evaluation.PlayerContext,
) FuzzingResult {
	// Convert deck strings to CardCandidates
	candidates := convertDeckToCandidates(deckCards, player)

	// Run evaluation
	evalResult := evaluation.Evaluate(candidates, synergyDB, playerContext)

	return FuzzingResult{
		Deck:                deckCards,
		OverallScore:        evalResult.OverallScore,
		AttackScore:         evalResult.Attack.Score,
		DefenseScore:        evalResult.Defense.Score,
		SynergyScore:        evalResult.Synergy.Score,
		VersatilityScore:    evalResult.Versatility.Score,
		AvgElixir:           evalResult.AvgElixir,
		Archetype:           string(evalResult.DetectedArchetype),
		ArchetypeConfidence: evalResult.ArchetypeConfidence,
		EvaluatedAt:         time.Now(),
	}
}

// saveDeckToStorage saves a deck evaluation result to persistent storage
func saveDeckToStorage(result FuzzingResult, _ string, storage *leaderboard.Storage) {
	// Reconstruct evalResult for storage (we only store what we need)
	entry := &leaderboard.DeckEntry{
		Cards:             result.Deck,
		OverallScore:      result.OverallScore,
		AttackScore:       result.AttackScore,
		DefenseScore:      result.DefenseScore,
		SynergyScore:      result.SynergyScore,
		VersatilityScore:  result.VersatilityScore,
		F2PScore:          0,
		PlayabilityScore:  0,
		Archetype:         result.Archetype,
		ArchetypeConf:     result.ArchetypeConfidence,
		AvgElixir:         result.AvgElixir,
		EvaluatedAt:       result.EvaluatedAt,
		PlayerTag:         "",
		EvaluationVersion: "1.0.0",
	}
	if _, _, err := storage.InsertDeck(entry); err != nil {
		fprintf(os.Stderr, "Warning: failed to store deck: %v\n", err)
	}
}

// convertDeckToCandidates converts a deck of card names to CardCandidates
func convertDeckToCandidates(deckCards []string, player *clashroyale.Player) []deck.CardCandidate {
	candidates := make([]deck.CardCandidate, 0, len(deckCards))

	// Build a map of player cards for quick lookup
	playerCardsMap := make(map[string]*clashroyale.Card)
	if player != nil {
		for i := range player.Cards {
			playerCardsMap[player.Cards[i].Name] = &player.Cards[i]
		}
	}

	for _, cardName := range deckCards {
		var candidate deck.CardCandidate
		var role config.CardRole

		// Try to get card info from player's cards first
		if playerCard, exists := playerCardsMap[cardName]; exists {
			role = config.GetCardRoleWithEvolution(cardName, playerCard.EvolutionLevel)
			candidate = deck.CardCandidate{
				Name:              cardName,
				Level:             playerCard.Level,
				MaxLevel:          playerCard.MaxLevel,
				Rarity:            playerCard.Rarity,
				Elixir:            playerCard.ElixirCost,
				Role:              &role,
				EvolutionLevel:    playerCard.EvolutionLevel,
				MaxEvolutionLevel: playerCard.MaxEvolutionLevel,
			}
		} else {
			// Card not in player's collection, use defaults
			role = config.GetCardRole(cardName)
			candidate = deck.CardCandidate{
				Name:     cardName,
				Level:    11,
				MaxLevel: 15,
				Rarity:   "Common",
				Elixir:   config.GetCardElixir(cardName, 0),
				Role:     &role,
			}
		}

		candidates = append(candidates, candidate)
	}

	return candidates
}

// filterResultsByScore filters results by minimum score thresholds
func filterResultsByScore(results []FuzzingResult, minOverall, minSynergy float64, _ bool) []FuzzingResult {
	filtered := make([]FuzzingResult, 0, len(results))

	for _, result := range results {
		passesOverall := result.OverallScore >= minOverall
		passesSynergy := result.SynergyScore >= minSynergy

		if passesOverall && passesSynergy {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// deduplicateResults removes duplicate decks based on card composition
// Keeps the first occurrence (highest score after sorting)
func deduplicateResults(results []FuzzingResult) []FuzzingResult {
	seen := make(map[string]bool)
	deduped := make([]FuzzingResult, 0, len(results))

	for _, result := range results {
		// Create a canonical key by sorting card names
		deckKey := deckKeyForResult(result)
		if !seen[deckKey] {
			seen[deckKey] = true
			deduped = append(deduped, result)
		}
	}

	return deduped
}

// deckKeyForResult creates a unique key for a deck based on sorted card names
func deckKeyForResult(result FuzzingResult) string {
	cards := make([]string, len(result.Deck))
	copy(cards, result.Deck)
	sort.Strings(cards)
	return strings.Join(cards, "|")
}

// sortFuzzingResults sorts fuzzing results by the specified field
func sortFuzzingResults(results []FuzzingResult, sortBy string) {
	sort.Slice(results, func(i, j int) bool {
		var iValue, jValue float64

		switch sortBy {
		case "overall":
			iValue = results[i].OverallScore
			jValue = results[j].OverallScore
		case "attack":
			iValue = results[i].AttackScore
			jValue = results[j].AttackScore
		case "defense":
			iValue = results[i].DefenseScore
			jValue = results[j].DefenseScore
		case "synergy":
			iValue = results[i].SynergyScore
			jValue = results[j].SynergyScore
		case "versatility":
			iValue = results[i].VersatilityScore
			jValue = results[j].VersatilityScore
		case "elixir":
			// For elixir, sort ascending (lower is better)
			return results[i].AvgElixir < results[j].AvgElixir
		default:
			iValue = results[i].OverallScore
			jValue = results[j].OverallScore
		}

		return iValue > jValue // Descending order (higher is better)
	})
}

// getTopResults returns the top N results
func getTopResults(results []FuzzingResult, top int) []FuzzingResult {
	if len(results) <= top {
		return results
	}
	return results[:top]
}

// formatFuzzingResults formats and outputs fuzzing results
func formatFuzzingResults(
	results []FuzzingResult,
	format string,
	playerName string,
	playerTag string,
	fuzzerConfig *deck.FuzzingConfig,
	generationTime time.Duration,
	stats *deck.FuzzingStats,
	totalFiltered int,
) error {
	switch format {
	case "json":
		return formatResultsJSON(results, playerName, playerTag, fuzzerConfig, generationTime, stats, totalFiltered)
	case "csv":
		return formatResultsCSV(results)
	case "detailed":
		return formatResultsDetailed(results, playerName, playerTag)
	default:
		return formatResultsSummary(results, playerName, playerTag, fuzzerConfig, generationTime, stats, totalFiltered)
	}
}

// formatResultsSummary outputs results in summary format
func formatResultsSummary(
	results []FuzzingResult,
	playerName string,
	playerTag string,
	fuzzerConfig *deck.FuzzingConfig,
	generationTime time.Duration,
	stats *deck.FuzzingStats,
	totalFiltered int,
) error {
	printf("\nDeck Fuzzing Results for %s (%s)\n", playerName, playerTag)
	printf("Generated %d random decks in %v\n", stats.Generated, generationTime.Round(time.Millisecond))
	printf("Configuration:\n")

	if len(fuzzerConfig.IncludeCards) > 0 {
		printf("  Include cards: %s\n", strings.Join(fuzzerConfig.IncludeCards, ", "))
	}
	if len(fuzzerConfig.ExcludeCards) > 0 {
		printf("  Exclude cards: %s\n", strings.Join(fuzzerConfig.ExcludeCards, ", "))
	}
	printf("  Elixir range: %.1f - %.1f\n", fuzzerConfig.MinAvgElixir, fuzzerConfig.MaxAvgElixir)
	if fuzzerConfig.MinOverallScore > 0 {
		printf("  Min overall score: %.1f\n", fuzzerConfig.MinOverallScore)
	}
	if fuzzerConfig.MinSynergyScore > 0 {
		printf("  Min synergy score: %.1f\n", fuzzerConfig.MinSynergyScore)
	}

	printf("\nTop %d Decks (from %d decks passing filters):\n\n", len(results), totalFiltered)

	// Print table header with multi-line deck display
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, ' ', 0)
	fprintln(w, "Rank\tDeck\tOverall\tAttack\tDefense\tSynergy\tElixir")

	// Print each deck with all 8 cards
	for i, result := range results {
		// Format deck with all cards (no truncation)
		deckStr := strings.Join(result.Deck, ", ")

		// If deck string is very long, use multi-line format
		if len(deckStr) > 50 {
			// First line: Rank, first 4 cards, scores
			firstLine := strings.Join(result.Deck[:4], ", ")
			fprintf(w, "%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\n",
				i+1,
				firstLine+",",
				result.OverallScore,
				result.AttackScore,
				result.DefenseScore,
				result.SynergyScore,
				result.AvgElixir,
			)

			// Second line: continuation with remaining cards
			secondLine := strings.Join(result.Deck[4:], ", ")
			fprintf(w, "\t%s\n", secondLine)
		} else {
			// Single line format for shorter deck strings
			fprintf(w, "%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\n",
				i+1,
				deckStr,
				result.OverallScore,
				result.AttackScore,
				result.DefenseScore,
				result.SynergyScore,
				result.AvgElixir,
			)
		}
	}

	flushWriter(w)

	return nil
}

// formatResultsJSON outputs results in JSON format
func formatResultsJSON(
	results []FuzzingResult,
	playerName string,
	playerTag string,
	fuzzerConfig *deck.FuzzingConfig,
	generationTime time.Duration,
	stats *deck.FuzzingStats,
	totalFiltered int,
) error {
	output := map[string]any{
		"player_name":             playerName,
		"player_tag":              playerTag,
		"generated":               stats.Generated,
		"success":                 stats.Success,
		"failed":                  stats.Failed,
		"filtered":                totalFiltered,
		"returned":                len(results),
		"generation_time_seconds": generationTime.Seconds(),
		"config": map[string]any{
			"count":             fuzzerConfig.Count,
			"workers":           fuzzerConfig.Workers,
			"include_cards":     fuzzerConfig.IncludeCards,
			"exclude_cards":     fuzzerConfig.ExcludeCards,
			"min_avg_elixir":    fuzzerConfig.MinAvgElixir,
			"max_avg_elixir":    fuzzerConfig.MaxAvgElixir,
			"min_overall_score": fuzzerConfig.MinOverallScore,
			"min_synergy_score": fuzzerConfig.MinSynergyScore,
		},
		"results": results,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// formatResultsCSV outputs results in CSV format
func formatResultsCSV(results []FuzzingResult) error {
	w := csv.NewWriter(os.Stdout)

	// Write header
	header := []string{"Rank", "Deck", "Overall", "Attack", "Defense", "Synergy", "Versatility", "AvgElixir", "Archetype"}
	if err := w.Write(header); err != nil {
		return err
	}

	// Write rows
	for i, result := range results {
		deckStr := strings.Join(result.Deck, ", ")
		row := []string{
			strconv.Itoa(i + 1),
			deckStr,
			fmt.Sprintf("%.2f", result.OverallScore),
			fmt.Sprintf("%.2f", result.AttackScore),
			fmt.Sprintf("%.2f", result.DefenseScore),
			fmt.Sprintf("%.2f", result.SynergyScore),
			fmt.Sprintf("%.2f", result.VersatilityScore),
			fmt.Sprintf("%.2f", result.AvgElixir),
			result.Archetype,
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	return nil
}

// formatResultsDetailed outputs results in detailed format with full evaluation
func formatResultsDetailed(
	results []FuzzingResult,
	playerName string,
	playerTag string,
) error {
	printf("\nDeck Fuzzing Results for %s (%s)\n", playerName, playerTag)
	printf("\nTop %d Decks:\n\n", len(results))

	for i, result := range results {
		printf("=== Deck %d ===\n", i+1)
		printf("Cards: %s\n", strings.Join(result.Deck, ", "))
		printf("Overall: %.2f | Attack: %.2f | Defense: %.2f | Synergy: %.2f | Versatility: %.2f\n",
			result.OverallScore, result.AttackScore, result.DefenseScore, result.SynergyScore, result.VersatilityScore)
		printf("Avg Elixir: %.2f | Archetype: %s (%.0f%% confidence)\n",
			result.AvgElixir, result.Archetype, result.ArchetypeConfidence*100)
		printf("Evaluated: %s\n\n", result.EvaluatedAt.Format(time.RFC3339))
	}

	return nil
}

// saveResultsToFile saves results to a file in the specified format
func saveResultsToFile(results []FuzzingResult, outputDir, format, playerTag string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102_150405")
	cleanTag := strings.TrimPrefix(playerTag, "#")
	var filename string

	switch format {
	case "json":
		filename = fmt.Sprintf("fuzz_%s_%s.json", cleanTag, timestamp)
	case "csv":
		filename = fmt.Sprintf("fuzz_%s_%s.csv", cleanTag, timestamp)
	default:
		filename = fmt.Sprintf("fuzz_%s_%s.txt", cleanTag, timestamp)
	}

	outputPath := filepath.Join(outputDir, filename)

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer closeFile(file)

	// Redirect stdout to file for formatting
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	os.Stdout = file
	os.Stderr = file
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Format results to file
	switch format {
	case "json":
		config := &deck.FuzzingConfig{}
		stats := &deck.FuzzingStats{}
		return formatResultsJSON(results, cleanTag, playerTag, config, 0, stats, len(results))
	case "csv":
		return formatResultsCSV(results)
	default:
		return formatResultsSummary(results, cleanTag, playerTag, &deck.FuzzingConfig{}, 0, &deck.FuzzingStats{}, len(results))
	}
}

// loadPlayerFromAnalysis loads player data from an existing analysis file
func loadPlayerFromAnalysis(analysisFile, analysisDir, playerTag string) (*clashroyale.Player, string, error) {
	var analysisPath string

	if analysisFile != "" {
		analysisPath = analysisFile
	} else {
		// Find latest analysis file for player
		cleanTag := strings.TrimPrefix(playerTag, "#")
		pattern := fmt.Sprintf("*analysis*%s.json", cleanTag)

		matches, err := filepath.Glob(filepath.Join(analysisDir, pattern))
		if err != nil {
			return nil, "", fmt.Errorf("failed to glob analysis files: %w", err)
		}

		if len(matches) == 0 {
			return nil, "", fmt.Errorf("no analysis files found for player %s", playerTag)
		}

		// Sort by modification time (newest first)
		sort.Slice(matches, func(i, j int) bool {
			infoI, _ := os.Stat(matches[i])
			infoJ, _ := os.Stat(matches[j])
			return infoI.ModTime().After(infoJ.ModTime())
		})

		analysisPath = matches[0]
	}

	// Load analysis data
	data, err := os.ReadFile(analysisPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read analysis file: %w", err)
	}

	var cardAnalysis analysis.CardAnalysis
	if err := json.Unmarshal(data, &cardAnalysis); err != nil {
		return nil, "", fmt.Errorf("failed to parse analysis JSON: %w", err)
	}

	// Convert analysis to player object
	player := &clashroyale.Player{
		Name:  cardAnalysis.PlayerName,
		Tag:   cardAnalysis.PlayerTag,
		Cards: make([]clashroyale.Card, 0, len(cardAnalysis.CardLevels)),
	}

	for cardName, cardData := range cardAnalysis.CardLevels {
		card := clashroyale.Card{
			Name:              cardName,
			Level:             cardData.Level,
			MaxLevel:          cardData.MaxLevel,
			Rarity:            cardData.Rarity,
			ElixirCost:        cardData.Elixir,
			EvolutionLevel:    cardData.EvolutionLevel,
			MaxEvolutionLevel: cardData.MaxEvolutionLevel,
		}
		player.Cards = append(player.Cards, card)
	}

	return player, cardAnalysis.PlayerName, nil
}

// saveTopDecksToStorage saves the top fuzzing results to persistent storage
func saveTopDecksToStorage(results []FuzzingResult, verbose bool) error {
	storage, err := fuzzstorage.NewStorage("")
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	// Convert FuzzingResult to fuzzstorage.DeckEntry
	entries := make([]fuzzstorage.DeckEntry, len(results))
	for i, result := range results {
		entries[i] = fuzzstorage.DeckEntry{
			Cards:            result.Deck,
			OverallScore:     result.OverallScore,
			AttackScore:      result.AttackScore,
			DefenseScore:     result.DefenseScore,
			SynergyScore:     result.SynergyScore,
			VersatilityScore: result.VersatilityScore,
			AvgElixir:        result.AvgElixir,
			Archetype:        result.Archetype,
			ArchetypeConf:    result.ArchetypeConfidence,
			EvaluatedAt:      result.EvaluatedAt,
		}
	}

	saved, err := storage.SaveTopDecks(entries)
	if err != nil {
		return fmt.Errorf("failed to save decks: %w", err)
	}

	total, _ := storage.Count()
	dbPath := storage.GetDBPath()

	if verbose {
		fprintf(os.Stderr, "\nTop decks saved to storage: %s\n", dbPath)
		fprintf(os.Stderr, "  New decks saved: %d\n", saved)
		fprintf(os.Stderr, "  Total decks in storage: %d\n", total)
	}

	return nil
}

// deckFuzzListCommand lists saved top decks from storage
func deckFuzzListCommand(ctx context.Context, cmd *cli.Command) error {
	top := cmd.Int("top")
	archetype := cmd.String("archetype")
	minScore := cmd.Float64("min-score")
	maxScore := cmd.Float64("max-score")
	minElixir := cmd.Float64("min-elixir")
	maxElixir := cmd.Float64("max-elixir")
	format := cmd.String("format")

	storage, err := fuzzstorage.NewStorage("")
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	// Build query options
	queryOpts := fuzzstorage.QueryOptions{
		Limit: top,
	}

	if archetype != "" {
		queryOpts.Archetype = archetype
	}
	if minScore > 0 {
		queryOpts.MinScore = minScore
	}
	if maxScore > 0 {
		queryOpts.MaxScore = maxScore
	}
	if minElixir > 0 {
		queryOpts.MinAvgElixir = minElixir
	}
	if maxElixir > 0 {
		queryOpts.MaxAvgElixir = maxElixir
	}

	// Query decks
	decks, err := storage.Query(queryOpts)
	if err != nil {
		return fmt.Errorf("failed to query decks: %w", err)
	}

	total, _ := storage.Count()
	dbPath := storage.GetDBPath()

	fprintf(os.Stderr, "Top decks from: %s\n", dbPath)
	fprintf(os.Stderr, "Showing %d of %d total decks\n\n", len(decks), total)

	// Format output
	switch format {
	case "json":
		return formatListResultsJSON(decks, dbPath, total)
	case "csv":
		return formatListResultsCSV(decks)
	case "detailed":
		return formatListResultsDetailed(decks, dbPath, total)
	default:
		return formatListResultsSummary(decks, dbPath, total)
	}
}

// deckFuzzUpdateCommand re-evaluates saved decks with current scoring and updates storage.
func deckFuzzUpdateCommand(ctx context.Context, cmd *cli.Command) error {
	top := cmd.Int("top")
	archetype := cmd.String("archetype")
	minScore := cmd.Float64("min-score")
	maxScore := cmd.Float64("max-score")
	minElixir := cmd.Float64("min-elixir")
	maxElixir := cmd.Float64("max-elixir")
	workers := cmd.Int("workers")
	verbose := cmd.Bool("verbose")

	if workers == 1 {
		workers = runtime.NumCPU()
		if verbose {
			fprintf(os.Stderr, "Auto-detected %d CPU cores, using %d workers\n", runtime.NumCPU(), workers)
		}
	}

	storage, err := fuzzstorage.NewStorage("")
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	queryOpts := fuzzstorage.QueryOptions{
		Limit: top,
	}
	if archetype != "" {
		queryOpts.Archetype = archetype
	}
	if minScore > 0 {
		queryOpts.MinScore = minScore
	}
	if maxScore > 0 {
		queryOpts.MaxScore = maxScore
	}
	if minElixir > 0 {
		queryOpts.MinAvgElixir = minElixir
	}
	if maxElixir > 0 {
		queryOpts.MaxAvgElixir = maxElixir
	}

	entries, err := storage.Query(queryOpts)
	if err != nil {
		return fmt.Errorf("failed to query decks: %w", err)
	}
	if len(entries) == 0 {
		fmt.Println("No decks found for update.")
		return nil
	}

	start := time.Now()
	updatedEntries := reevaluateStoredDecks(entries, workers, verbose)

	updated := 0
	for i := range updatedEntries {
		if err := storage.UpdateDeck(&updatedEntries[i]); err != nil {
			return fmt.Errorf("failed to update deck %d: %w", updatedEntries[i].ID, err)
		}
		updated++
	}

	if verbose {
		fprintf(os.Stderr, "Updated %d decks in %v\n", updated, time.Since(start).Round(time.Millisecond))
		fprintf(os.Stderr, "Database: %s\n", storage.GetDBPath())
	}

	printf("Updated %d saved decks\n", updated)
	return nil
}

type storedDeckWork struct {
	index int
	entry fuzzstorage.DeckEntry
}

type storedDeckResult struct {
	index int
	entry fuzzstorage.DeckEntry
}

func reevaluateStoredDecks(entries []fuzzstorage.DeckEntry, workers int, verbose bool) []fuzzstorage.DeckEntry {
	if workers <= 1 {
		return reevaluateStoredDecksSequential(entries, verbose)
	}

	results := make([]fuzzstorage.DeckEntry, len(entries))
	workChan := make(chan storedDeckWork, len(entries))
	resultChan := make(chan storedDeckResult, len(entries))
	var wg sync.WaitGroup

	var bar *progressbar.ProgressBar
	if verbose {
		bar = progressbar.NewOptions(len(entries),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("decks"),
			progressbar.OptionOnCompletion(func() {
				fprintln(os.Stderr)
			}),
		)
	}

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			synergyDB := deck.NewSynergyDatabase()

			for work := range workChan {
				result := evaluateSingleDeck(work.entry.Cards, nil, "", synergyDB, nil)
				updated := work.entry
				updated.OverallScore = result.OverallScore
				updated.AttackScore = result.AttackScore
				updated.DefenseScore = result.DefenseScore
				updated.SynergyScore = result.SynergyScore
				updated.VersatilityScore = result.VersatilityScore
				updated.AvgElixir = result.AvgElixir
				updated.Archetype = result.Archetype
				updated.ArchetypeConf = result.ArchetypeConfidence
				updated.EvaluatedAt = result.EvaluatedAt
				resultChan <- storedDeckResult{index: work.index, entry: updated}
			}
		}()
	}

	for i, entry := range entries {
		workChan <- storedDeckWork{index: i, entry: entry}
	}
	close(workChan)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		results[result.index] = result.entry
		if verbose && bar != nil {
			if err := bar.Add(1); err != nil {
				fprintf(os.Stderr, "Warning: progress update failed: %v\n", err)
			}
		}
	}

	return results
}

func reevaluateStoredDecksSequential(entries []fuzzstorage.DeckEntry, verbose bool) []fuzzstorage.DeckEntry {
	results := make([]fuzzstorage.DeckEntry, len(entries))
	synergyDB := deck.NewSynergyDatabase()

	var bar *progressbar.ProgressBar
	if verbose {
		bar = progressbar.NewOptions(len(entries),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("decks"),
			progressbar.OptionOnCompletion(func() {
				fprintln(os.Stderr)
			}),
		)
	}

	for i, entry := range entries {
		result := evaluateSingleDeck(entry.Cards, nil, "", synergyDB, nil)
		entry.OverallScore = result.OverallScore
		entry.AttackScore = result.AttackScore
		entry.DefenseScore = result.DefenseScore
		entry.SynergyScore = result.SynergyScore
		entry.VersatilityScore = result.VersatilityScore
		entry.AvgElixir = result.AvgElixir
		entry.Archetype = result.Archetype
		entry.ArchetypeConf = result.ArchetypeConfidence
		entry.EvaluatedAt = result.EvaluatedAt
		results[i] = entry

		if verbose && bar != nil {
			if err := bar.Add(1); err != nil {
				fprintf(os.Stderr, "Warning: progress update failed: %v\n", err)
			}
		}
	}

	return results
}

// formatListResultsSummary formats list results in summary format
func formatListResultsSummary(decks []fuzzstorage.DeckEntry, dbPath string, total int) error {
	printf("Saved Top Decks\n")
	printf("Database: %s\n", dbPath)
	printf("Total decks: %d\n\n", total)

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, ' ', 0)
	fprintln(w, "Rank\tDeck\tOverall\tAttack\tDefense\tSynergy\tElixir\tArchetype")

	for i, deck := range decks {
		deckStr := strings.Join(deck.Cards, ", ")
		if len(deckStr) > 50 {
			firstLine := strings.Join(deck.Cards[:4], ", ")
			fprintf(w, "%d\t%s,\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%s\n",
				i+1, firstLine, deck.OverallScore, deck.AttackScore, deck.DefenseScore,
				deck.SynergyScore, deck.AvgElixir, deck.Archetype)
			secondLine := strings.Join(deck.Cards[4:], ", ")
			fprintf(w, "\t%s\n", secondLine)
		} else {
			fprintf(w, "%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%s\n",
				i+1, deckStr, deck.OverallScore, deck.AttackScore, deck.DefenseScore,
				deck.SynergyScore, deck.AvgElixir, deck.Archetype)
		}
	}

	flushWriter(w)
	return nil
}

// formatListResultsJSON formats list results in JSON format
func formatListResultsJSON(decks []fuzzstorage.DeckEntry, dbPath string, total int) error {
	output := map[string]any{
		"database": dbPath,
		"total":    total,
		"returned": len(decks),
		"results":  decks,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// formatListResultsCSV formats list results in CSV format
func formatListResultsCSV(decks []fuzzstorage.DeckEntry) error {
	w := csv.NewWriter(os.Stdout)

	header := []string{"Rank", "Deck", "Overall", "Attack", "Defense", "Synergy", "Versatility", "AvgElixir", "Archetype"}
	if err := w.Write(header); err != nil {
		return err
	}

	for i, deck := range decks {
		deckStr := strings.Join(deck.Cards, ", ")
		row := []string{
			strconv.Itoa(i + 1),
			deckStr,
			fmt.Sprintf("%.2f", deck.OverallScore),
			fmt.Sprintf("%.2f", deck.AttackScore),
			fmt.Sprintf("%.2f", deck.DefenseScore),
			fmt.Sprintf("%.2f", deck.SynergyScore),
			fmt.Sprintf("%.2f", deck.VersatilityScore),
			fmt.Sprintf("%.2f", deck.AvgElixir),
			deck.Archetype,
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	return nil
}

// formatListResultsDetailed formats list results in detailed format
func formatListResultsDetailed(decks []fuzzstorage.DeckEntry, dbPath string, total int) error {
	printf("Saved Top Decks\n")
	printf("Database: %s\n", dbPath)
	printf("Total decks: %d\n\n", total)

	for i, deck := range decks {
		printf("=== Deck %d ===\n", i+1)
		printf("Cards: %s\n", strings.Join(deck.Cards, ", "))
		printf("Overall: %.2f | Attack: %.2f | Defense: %.2f | Synergy: %.2f | Versatility: %.2f\n",
			deck.OverallScore, deck.AttackScore, deck.DefenseScore, deck.SynergyScore, deck.VersatilityScore)
		printf("Avg Elixir: %.2f | Archetype: %s (%.0f%% confidence)\n",
			deck.AvgElixir, deck.Archetype, deck.ArchetypeConf*100)
		printf("Evaluated: %s\n\n", deck.EvaluatedAt.Format(time.RFC3339))
	}

	return nil
}

// loadCardsFromSavedDecks loads unique cards from top N saved decks
func loadCardsFromSavedDecks(n int, _ bool) ([]string, error) {
	storage, err := fuzzstorage.NewStorage("")
	if err != nil {
		return nil, err
	}
	defer closeFile(storage)

	decks, err := storage.GetTopN(n)
	if err != nil {
		return nil, err
	}

	// Extract unique cards
	cardMap := make(map[string]bool)
	for _, deck := range decks {
		for _, card := range deck.Cards {
			cardMap[card] = true
		}
	}

	cards := make([]string, 0, len(cardMap))
	for card := range cardMap {
		cards = append(cards, card)
	}

	return cards, nil
}

// mergeUniqueCards merges two card slices, removing duplicates
func mergeUniqueCards(base, additional []string) []string {
	cardMap := make(map[string]bool)

	// Add base cards
	for _, card := range base {
		cardMap[card] = true
	}

	// Add additional cards
	for _, card := range additional {
		cardMap[card] = true
	}

	result := make([]string, 0, len(cardMap))
	for card := range cardMap {
		result = append(result, card)
	}

	return result
}

// loadSavedDecksForSeeding loads top N saved decks for use as mutation seeds
func loadSavedDecksForSeeding(n int, _ *clashroyale.Player, verbose bool) ([][]string, error) {
	storage, err := fuzzstorage.NewStorage("")
	if err != nil {
		return nil, err
	}
	defer closeFile(storage)

	entries, err := storage.GetTopN(n)
	if err != nil {
		return nil, err
	}

	if verbose {
		fprintf(os.Stderr, "Loaded %d saved decks for seeding\n", len(entries))
	}

	// Convert to deck slices
	decks := make([][]string, len(entries))
	for i, entry := range entries {
		decks[i] = entry.Cards
	}

	return decks, nil
}

// generateDeckMutations generates mutations of saved decks by swapping cards
func generateDeckMutations(savedDecks [][]string, player *clashroyale.Player, count int, verbose bool) [][]string {
	if player == nil || len(player.Cards) == 0 {
		if verbose {
			fprintf(os.Stderr, "No player cards available for mutations\n")
		}
		return nil
	}

	// Build available cards map
	availableCards := make(map[string]bool)
	for _, card := range player.Cards {
		availableCards[card.Name] = true
	}

	mutations := make([][]string, 0)
	mutationsPerDeck := count / len(savedDecks)
	if mutationsPerDeck < 1 {
		mutationsPerDeck = 1
	}

	for _, deck := range savedDecks {
		for i := 0; i < mutationsPerDeck; i++ {
			// Create mutation by swapping 1-2 random cards
			mutation := make([]string, len(deck))
			copy(mutation, deck)

			// Swap 1-2 cards
			numSwaps := 1 + (i % 2) // Alternate between 1 and 2 swaps
			for range numSwaps {
				// Find cards to swap
				swapIdx := i % len(mutation)

				// Find a replacement card
				for _, card := range player.Cards {
					// Skip if card is already in deck
					alreadyInDeck := false
					for _, existing := range mutation {
						if existing == card.Name {
							alreadyInDeck = true
							break
						}
					}
					if !alreadyInDeck {
						mutation[swapIdx] = card.Name
						break
					}
				}
			}

			mutations = append(mutations, mutation)
		}
	}

	return mutations
}

// loadDeckFromStorage loads a specific deck from storage by ID or name
func loadDeckFromStorage(deckRef string, verbose bool) ([]string, error) {
	storage, err := fuzzstorage.NewStorage("")
	if err != nil {
		return nil, fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	// Try to parse as integer ID
	var deckID int
	if _, err := fmt.Sscanf(deckRef, "%d", &deckID); err == nil {
		// Query by ID using the database directly
		entries, err := storage.Query(fuzzstorage.QueryOptions{
			Limit: 1000, // Get all decks to find by ID
		})
		if err != nil {
			return nil, fmt.Errorf("failed to query storage: %w", err)
		}

		for _, entry := range entries {
			if entry.ID == deckID {
				if verbose {
					fprintf(os.Stderr, "Loaded deck by ID %d: %s\n", deckID, strings.Join(entry.Cards, ", "))
				}
				return entry.Cards, nil
			}
		}
		return nil, fmt.Errorf("no deck found with ID %d", deckID)
	}

	// Try to find by matching deck cards (partial match)
	entries, err := storage.Query(fuzzstorage.QueryOptions{
		Limit: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query storage: %w", err)
	}

	// Try to find deck that matches the reference (could be card names or partial deck)
	deckRefLower := strings.ToLower(deckRef)
	for _, entry := range entries {
		deckStr := strings.ToLower(strings.Join(entry.Cards, " "))
		if strings.Contains(deckStr, deckRefLower) {
			if verbose {
				fprintf(os.Stderr, "Loaded matching deck: %s\n", strings.Join(entry.Cards, ", "))
			}
			return entry.Cards, nil
		}
	}

	return nil, fmt.Errorf("no deck found matching '%s'", deckRef)
}

// generateVariations generates variations of a base deck by swapping some cards
func generateVariations(baseDeck []string, player *clashroyale.Player, count int, verbose bool) [][]string {
	if player == nil || len(player.Cards) == 0 {
		if verbose {
			fprintf(os.Stderr, "No player cards available for variations\n")
		}
		return nil
	}

	// Build available cards map (excluding cards already in base deck)
	availableCards := make([]string, 0)
	baseDeckMap := make(map[string]bool)
	for _, card := range baseDeck {
		baseDeckMap[card] = true
	}

	for _, card := range player.Cards {
		if !baseDeckMap[card.Name] {
			availableCards = append(availableCards, card.Name)
		}
	}

	if len(availableCards) == 0 {
		if verbose {
			fprintf(os.Stderr, "No additional cards available for variations\n")
		}
		return nil
	}

	variations := make([][]string, 0, count)

	// Generate variations by swapping 1-3 cards
	for i := 0; i < count; i++ {
		variation := make([]string, len(baseDeck))
		copy(variation, baseDeck)

		// Number of cards to swap (1-3, varying across variations)
		numSwaps := 1 + (i % 3)

		// Swap random positions with available cards
		for j := 0; j < numSwaps; j++ {
			// Pick a random position to swap
			swapIdx := j % len(variation)

			// Pick a random replacement card
			if len(availableCards) > 0 {
				replacementIdx := (i + j) % len(availableCards)
				variation[swapIdx] = availableCards[replacementIdx]
			}
		}

		variations = append(variations, variation)
	}

	if verbose {
		fprintf(os.Stderr, "Generated %d variations of base deck\n", len(variations))
	}

	return variations
}

// runtimeEstimate holds the estimated runtime information
type runtimeEstimate struct {
	totalSeconds float64
	decksPerSec  float64
}

// estimateRuntime generates a sample of decks and extrapolates the runtime
func estimateRuntime(fuzzer *deck.DeckFuzzer, targetCount, sampleSize int) (*runtimeEstimate, error) {
	fprintf(os.Stderr, "Measuring generation rate from %d deck sample...\n", sampleSize)

	sampleDecks, sampleDuration, err := fuzzer.GenerateSampleDecks(sampleSize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sample: %w", err)
	}

	if len(sampleDecks) == 0 {
		return nil, fmt.Errorf("no decks generated in sample")
	}

	decksPerSec := float64(len(sampleDecks)) / sampleDuration.Seconds()
	totalSeconds := float64(targetCount) / decksPerSec

	return &runtimeEstimate{
		totalSeconds: totalSeconds,
		decksPerSec:  decksPerSec,
	}, nil
}

// formatDuration formats a duration in seconds to a human-readable string
func formatDuration(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.0fs", seconds)
	}
	minutes := int(seconds / 60)
	secs := int(seconds) % 60
	if secs == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dm %ds", minutes, secs)
}

// confirmAction prompts the user to confirm before proceeding
func confirmAction(prompt string) (bool, error) {
	fprintf(os.Stderr, "%s", prompt)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false, err
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

// printFuzzingProgress prints real-time progress during fuzzing
