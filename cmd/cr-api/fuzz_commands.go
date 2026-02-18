package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"slices"
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
	"github.com/klauer/clash-royale-api/go/pkg/deck/genetic"
	"github.com/klauer/clash-royale-api/go/pkg/deck/research"
	"github.com/klauer/clash-royale-api/go/pkg/fuzzstorage"
	"github.com/klauer/clash-royale-api/go/pkg/leaderboard"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v3"
)

type stageCanceler struct {
	mu     sync.Mutex
	cancel context.CancelFunc
}

const (
	gaFitnessModeLegacy        = "legacy-evaluation"
	gaFitnessModeArchetypeFree = "archetype-free-composite"
)

func selectGAFitnessEvaluator(useArchetypes bool) (func([]deck.CardCandidate) (float64, error), string) {
	if useArchetypes {
		return nil, gaFitnessModeLegacy
	}

	constraints := research.DefaultConstraintConfig()
	synergyDB := deck.NewSynergyDatabase()
	return func(deckCards []deck.CardCandidate) (float64, error) {
		metrics := research.ScoreDeckComposite(deckCards, synergyDB, constraints)
		return metrics.Composite * 10.0, nil
	}, gaFitnessModeArchetypeFree
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
	mutationIntensity := cmd.Int("mutation-intensity")
	archetypes := cmd.StringSlice("archetypes")
	refineRounds := cmd.Int("refine")
	uniquenessWeight := cmd.Float64("uniqueness-weight")
	ensureArchetypes := cmd.Bool("ensure-archetypes")
	ensureElixirBuckets := cmd.Bool("ensure-elixir-buckets")
	mode := strings.ToLower(cmd.String("mode"))
	gaPopulation := cmd.Int("ga-population")
	gaGenerations := cmd.Int("ga-generations")
	gaMutationRate := cmd.Float64("ga-mutation-rate")
	gaCrossoverRate := cmd.Float64("ga-crossover-rate")
	gaMutationIntensity := cmd.Float64("ga-mutation-intensity")
	gaEliteCount := cmd.Int("ga-elite-count")
	gaTournamentSize := cmd.Int("ga-tournament-size")
	gaParallelEval := cmd.Bool("ga-parallel-eval")
	gaConvergenceGenerations := cmd.Int("ga-convergence-generations")
	gaTargetFitness := cmd.Float64("ga-target-fitness")
	gaIslandModel := cmd.Bool("ga-island-model")
	gaIslandCount := cmd.Int("ga-island-count")
	gaMigrationInterval := cmd.Int("ga-migration-interval")
	gaMigrationSize := cmd.Int("ga-migration-size")
	gaUseArchetypes := cmd.Bool("ga-use-archetypes")

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
	if mode == "" {
		mode = "random"
	}
	if mode != "random" && mode != "genetic" {
		return fmt.Errorf("invalid --mode value: %s (must be random or genetic)", mode)
	}

	// Validate archetypes
	validArchetypes := map[string]bool{
		"beatdown":  true,
		"control":   true,
		"cycle":     true,
		"bridge":    true,
		"siege":     true,
		"bait":      true,
		"graveyard": true,
		"miner":     true,
		"hybrid":    true,
	}
	for _, arch := range archetypes {
		arch = strings.ToLower(strings.TrimSpace(arch))
		if !validArchetypes[arch] {
			return fmt.Errorf("invalid archetype '%s' (must be one of: beatdown, control, cycle, bridge, siege, bait, graveyard, miner, hybrid)", arch)
		}
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

		player, err = client.GetPlayerWithContext(ctx, cleanTag)
		if err != nil {
			return fmt.Errorf("failed to fetch player: %w", err)
		}
		playerName = player.Name
	}

	if verbose {
		fprintf(os.Stderr, "Loaded player: %s (%s)\n", playerName, player.Tag)
		fprintf(os.Stderr, "Cards available: %d\n", len(player.Cards))
	}

	// Normalize archetypes to lowercase
	normalizedArchetypes := make([]string, 0, len(archetypes))
	for _, arch := range archetypes {
		normalizedArchetypes = append(normalizedArchetypes, strings.ToLower(strings.TrimSpace(arch)))
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
		MutationIntensity: mutationIntensity,
		ArchetypeFilter:   normalizedArchetypes,
		UniquenessWeight:  uniquenessWeight,
		EnsureArchetypes:  ensureArchetypes,
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

	seed := cmd.Int("seed")
	if seed != 0 {
		fuzzerCfg.Seed = int64(seed)
	}

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

	if mode == "genetic" && verbose {
		if synergyPairs {
			fprintf(os.Stderr, "Warning: --synergy-pairs is ignored in genetic mode\n")
		}
		if evolutionCentric {
			fprintf(os.Stderr, "Warning: --evolution-centric is ignored in genetic mode\n")
		}
	}

	var generatedDecks [][]string
	var generationTime time.Duration
	var stats deck.FuzzingStats

	if mode == "genetic" {
		if verbose {
			fprintf(os.Stderr, "\nStarting deck fuzzing (genetic mode)...\n")
			if refineRounds > 1 {
				fprintf(os.Stderr, "Running %d refinement rounds\n", refineRounds)
			}
		}

		candidates, err := buildGeneticCandidates(player, includeCards, excludeCards)
		if err != nil {
			return err
		}
		fitnessEvaluator, gaFitnessMode := selectGAFitnessEvaluator(gaUseArchetypes)
		if verbose {
			fprintf(os.Stderr, "GA objective: %s\n", gaFitnessMode)
		}

		// Store initial seed decks for first round
		initialSeedDecks := filterDecksByIncludeExclude(seedDecks, includeCards, excludeCards)

		// Use saved decks, mutations, and variations as seed population for first round.
		if fromSaved > 0 && !interrupted.Load() {
			savedDecks, err := loadSavedDecksForSeeding(fromSaved, player, verbose)
			if err != nil {
				return fmt.Errorf("failed to load saved decks for seeding: %w", err)
			}
			if len(savedDecks) > 0 {
				mutations := generateDeckMutations(savedDecks, player, count, fuzzerCfg.MutationIntensity, verbose)
				mutations = filterDecksByIncludeExclude(mutations, includeCards, excludeCards)
				initialSeedDecks = append(initialSeedDecks, mutations...)
			}
		}

		if basedOn != "" && !interrupted.Load() {
			baseDeck, err := loadDeckFromStorage(basedOn, verbose)
			if err != nil {
				return fmt.Errorf("failed to load deck from storage: %w", err)
			}
			variations := generateVariations(baseDeck, player, count, fuzzerCfg.MutationIntensity, verbose)
			if len(variations) > 0 {
				variations = filterDecksByIncludeExclude(variations, includeCards, excludeCards)
				initialSeedDecks = append(initialSeedDecks, variations...)
			}
		}

		// Iterative refinement loop
		currentSeedDecks := initialSeedDecks
		var allRoundResults [][]*genetic.DeckGenome
		var totalTime time.Duration

		for round := 1; round <= refineRounds; round++ {
			if interrupted.Load() {
				break
			}

			if verbose && refineRounds > 1 {
				fprintf(os.Stderr, "\n--- Refinement Round %d/%d ---\n", round, refineRounds)
			}

			gaConfig := genetic.DefaultGeneticConfig()
			gaConfig.PopulationSize = gaPopulation
			gaConfig.Generations = gaGenerations
			gaConfig.CrossoverRate = gaCrossoverRate
			gaConfig.TournamentSize = gaTournamentSize
			gaConfig.ParallelEvaluations = gaParallelEval
			gaConfig.ConvergenceGenerations = gaConvergenceGenerations
			gaConfig.TargetFitness = gaTargetFitness
			gaConfig.IslandModel = gaIslandModel
			gaConfig.IslandCount = gaIslandCount
			gaConfig.MigrationInterval = gaMigrationInterval
			gaConfig.MigrationSize = gaMigrationSize
			gaConfig.UseArchetypes = gaUseArchetypes

			// Progressive refinement: adjust parameters each round
			if round == 1 {
				// First round: use user-specified parameters
				gaConfig.MutationRate = gaMutationRate
				gaConfig.MutationIntensity = gaMutationIntensity
				gaConfig.EliteCount = gaEliteCount
			} else {
				// Subsequent rounds: reduce exploration, increase exploitation
				// Gradually reduce mutation rate (min 0.02)
				mutationRate := gaMutationRate * math.Pow(0.7, float64(round-1))
				if mutationRate < 0.02 {
					mutationRate = 0.02
				}
				gaConfig.MutationRate = mutationRate
				// Gradually reduce mutation intensity (min 0.1)
				mutationIntensity := gaMutationIntensity * math.Pow(0.7, float64(round-1))
				if mutationIntensity < 0.1 {
					mutationIntensity = 0.1
				}
				gaConfig.MutationIntensity = mutationIntensity
				// Gradually increase elite count (max 20% of population)
				eliteCount := gaEliteCount + round - 1
				maxElite := gaPopulation / 5
				if eliteCount > maxElite {
					eliteCount = maxElite
				}
				gaConfig.EliteCount = eliteCount
			}

			// Use seed decks from previous round
			if len(currentSeedDecks) > 0 {
				gaConfig.SeedPopulation = currentSeedDecks
			}

			optimizer, err := genetic.NewGeneticOptimizer(candidates, deck.StrategyBalanced, &gaConfig)
			if err != nil {
				return fmt.Errorf("failed to create genetic optimizer: %w", err)
			}
			optimizer.FitnessFunc = fitnessEvaluator
			if seed != 0 {
				optimizer.RNG = rand.New(rand.NewSource(int64(seed) + int64(round)))
			}
			if verbose {
				startTime := time.Now()
				totalGens := gaGenerations
				totalPop := gaPopulation
				optimizer.Progress = func(progress genetic.GeneticProgress) {
					gens := int(progress.Generation)
					elapsed := time.Since(startTime)
					etaStr := "?"
					if gens > 0 {
						rate := float64(gens) / elapsed.Seconds()
						remaining := max(totalGens-gens, 0)
						if rate > 0 {
							etaStr = formatDurationFloor(float64(remaining) / rate)
						}
					}
					evalsDone := int64(gens) * int64(totalPop)
					if refineRounds > 1 {
						fprintf(
							os.Stderr,
							"\rRound %d: GA gen %d/%d | evals ~%d | best %.2f | avg %.2f | elapsed %s | eta %s",
							round,
							progress.Generation,
							totalGens,
							evalsDone,
							progress.BestFitness,
							progress.AvgFitness,
							formatDurationFloor(elapsed.Seconds()),
							etaStr,
						)
					} else {
						fprintf(
							os.Stderr,
							"\rGA gen %d/%d | evals ~%d | best %.2f | avg %.2f | elapsed %s | eta %s",
							progress.Generation,
							totalGens,
							evalsDone,
							progress.BestFitness,
							progress.AvgFitness,
							formatDurationFloor(elapsed.Seconds()),
							etaStr,
						)
					}
				}
			}

			startTime := time.Now()
			result, err := optimizer.Optimize()
			if verbose {
				fprintln(os.Stderr)
			}
			if err != nil {
				return fmt.Errorf("failed to optimize decks in round %d: %w", round, err)
			}
			roundTime := result.Duration
			if roundTime == 0 {
				roundTime = time.Since(startTime)
			}
			totalTime += roundTime

			// Store results from this round
			allRoundResults = append(allRoundResults, result.HallOfFame)

			// Prepare seed decks for next round: use top decks from this round
			if round < refineRounds {
				topCount := min(gaPopulation/4, len(result.HallOfFame))
				if topCount == 0 && len(result.HallOfFame) > 0 {
					topCount = len(result.HallOfFame)
				}
				currentSeedDecks = make([][]string, 0, topCount)
				for i := 0; i < topCount && i < len(result.HallOfFame); i++ {
					if result.HallOfFame[i] == nil {
						continue
					}
					clone := make([]string, len(result.HallOfFame[i].Cards))
					copy(clone, result.HallOfFame[i].Cards)
					currentSeedDecks = append(currentSeedDecks, clone)
				}
				if verbose && len(currentSeedDecks) > 0 {
					fprintf(os.Stderr, "Round %d complete. Using top %d decks as seeds for next round\n", round, len(currentSeedDecks))
				}
			}
		}

		generationTime = totalTime

		// Combine results from all rounds, preferring later rounds
		generatedDecks = make([][]string, 0)
		seenDecks := make(map[string]bool)

		// Add decks from later rounds first (they're more refined)
		for i := len(allRoundResults) - 1; i >= 0; i-- {
			for _, genome := range allRoundResults[i] {
				if genome == nil {
					continue
				}
				deckKey := strings.Join(genome.Cards, ",")
				if seenDecks[deckKey] {
					continue
				}
				seenDecks[deckKey] = true
				clone := make([]string, len(genome.Cards))
				copy(clone, genome.Cards)
				generatedDecks = append(generatedDecks, clone)
			}
		}

		generatedDecks = filterDecksByIncludeExclude(generatedDecks, includeCards, excludeCards)
		stats.Generated = len(generatedDecks)
		stats.Success = len(generatedDecks)
	} else {
		// Create fuzzer
		fuzzer, err := deck.NewDeckFuzzer(player, fuzzerCfg)
		if err != nil {
			return fmt.Errorf("failed to create fuzzer: %w", err)
		}

		// Runtime estimation for large batches
		const sampleSize = 1000
		var estimate *runtimeEstimate
		if count > sampleSize {
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
			fprintf(os.Stderr, "  Mode: random\n")
			fprintf(os.Stderr, "  Count: %d\n", count)
			fprintf(os.Stderr, "  Workers: %d\n", workers)
			if synergyPairs {
				fprintf(os.Stderr, "  Mode: synergy-first (4 pairs)\n")
			}
			if evolutionCentric {
				fprintf(os.Stderr, "  Mode: evolution-centric (min %d evo cards, level %d+)\n", minEvoCards, minEvoLevel)
			}
			if len(normalizedArchetypes) > 0 {
				fprintf(os.Stderr, "  Archetype filter: %s\n", strings.Join(normalizedArchetypes, ", "))
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
			generationDone.Go(func() {
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
			})
		}

		generationCtx, cancelGeneration := context.WithCancel(ctx)
		canceler.Set(cancelGeneration)

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

		// Stop progress reporter
		close(stopProgress)
		generationDone.Wait()
		fprintln(os.Stderr) // New line after progress

		generationTime = time.Since(startTime)
		stats = fuzzer.GetStats()
	}

	if mode != "genetic" {
		// Handle --from-saved: add mutations of saved decks
		if fromSaved > 0 && !interrupted.Load() {
			savedDecks, err := loadSavedDecksForSeeding(fromSaved, player, verbose)
			if err != nil {
				return fmt.Errorf("failed to load saved decks for seeding: %w", err)
			}
			if len(savedDecks) > 0 {
				mutations := generateDeckMutations(savedDecks, player, count, fuzzerCfg.MutationIntensity, verbose)
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
			variations := generateVariations(baseDeck, player, count, fuzzerCfg.MutationIntensity, verbose)
			if len(variations) > 0 {
				generatedDecks = append(generatedDecks, variations...)
				if verbose {
					fprintf(os.Stderr, "Added %d variations based on deck: %s\n", len(variations), strings.Join(baseDeck, ", "))
				}
			}
		}
	}

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

	// Filter by archetype if specified
	archetypeFilteredResults := filterResultsByArchetype(filteredResults, normalizedArchetypes, verbose)

	if len(archetypeFilteredResults) == 0 && len(normalizedArchetypes) > 0 {
		return fmt.Errorf("no decks matched the specified archetypes: %s", strings.Join(normalizedArchetypes, ", "))
	}

	if len(normalizedArchetypes) > 0 {
		if verbose {
			fprintf(os.Stderr, "%d decks passed archetype filter (%s)\n", len(archetypeFilteredResults), strings.Join(normalizedArchetypes, ", "))
		}
		filteredResults = archetypeFilteredResults
	}

	// Deduplicate results (remove identical decks)
	dedupedResults := deduplicateResults(filteredResults)
	if verbose {
		fprintf(os.Stderr, "Removed %d duplicate decks, %d unique decks remaining\n", len(filteredResults)-len(dedupedResults), len(dedupedResults))
	}

	// Sort results
	sortFuzzingResults(dedupedResults, sortBy)

	// Ensure archetype coverage if requested
	if ensureArchetypes && mode != "genetic" {
		dedupedResults = ensureArchetypeCoverage(dedupedResults, top, verbose)
	}
	if ensureElixirBuckets {
		dedupedResults = ensureElixirBucketDistribution(dedupedResults, top, verbose)
	}

	// Get top N results
	topResults := getTopResults(dedupedResults, top)

	// Format and output results
	if err := formatFuzzingResults(topResults, format, playerName, playerTag, fuzzerCfg, mode, generationTime, &stats, len(dedupedResults)); err != nil {
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
	ContextualScore     float64
	LadderScore         float64
	NormalizedScore     float64
	DeckLevelRatio      float64
	NormalizationFactor float64
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
		wg.Go(func() {

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
		})
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

	contextualScore := evalResult.OverallScore
	ladderScore := 0.0
	normalizedScore := evalResult.OverallScore
	deckLevelRatio := 1.0
	normalizationFactor := 1.0
	if evalResult.OverallBreakdown != nil {
		contextualScore = evalResult.OverallBreakdown.ContextualScore
		ladderScore = evalResult.OverallBreakdown.LadderScore
		normalizedScore = evalResult.OverallBreakdown.NormalizedScore
		deckLevelRatio = evalResult.OverallBreakdown.DeckLevelRatio
		normalizationFactor = evalResult.OverallBreakdown.NormalizationFactor
	}

	return FuzzingResult{
		Deck:                deckCards,
		OverallScore:        evalResult.OverallScore,
		ContextualScore:     contextualScore,
		LadderScore:         ladderScore,
		NormalizedScore:     normalizedScore,
		DeckLevelRatio:      deckLevelRatio,
		NormalizationFactor: normalizationFactor,
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

func buildGeneticCandidates(player *clashroyale.Player, includeCards, excludeCards []string) ([]*deck.CardCandidate, error) {
	if player == nil {
		return nil, fmt.Errorf("player cannot be nil")
	}

	excludeMap := make(map[string]bool)
	for _, card := range excludeCards {
		cardName := strings.TrimSpace(card)
		if cardName == "" {
			continue
		}
		excludeMap[cardName] = true
	}

	includeMap := make(map[string]bool)
	for _, card := range includeCards {
		cardName := strings.TrimSpace(card)
		if cardName == "" {
			continue
		}
		if excludeMap[cardName] {
			return nil, fmt.Errorf("card %q is both included and excluded", cardName)
		}
		includeMap[cardName] = true
	}

	candidates := make([]*deck.CardCandidate, 0, len(player.Cards))
	for _, card := range player.Cards {
		cardName := strings.TrimSpace(card.Name)
		if excludeMap[cardName] {
			continue
		}

		role := config.GetCardRoleWithEvolution(cardName, card.EvolutionLevel)
		candidate := &deck.CardCandidate{
			Name:              cardName,
			Level:             card.Level,
			MaxLevel:          card.MaxLevel,
			Rarity:            card.Rarity,
			Elixir:            config.GetCardElixir(cardName, card.ElixirCost),
			Role:              &role,
			EvolutionLevel:    card.EvolutionLevel,
			MaxEvolutionLevel: card.MaxEvolutionLevel,
		}
		candidates = append(candidates, candidate)
		delete(includeMap, cardName)
	}

	if len(includeMap) > 0 {
		missing := make([]string, 0, len(includeMap))
		for cardName := range includeMap {
			missing = append(missing, cardName)
		}
		sort.Strings(missing)
		return nil, fmt.Errorf("include cards not in player collection: %s", strings.Join(missing, ", "))
	}
	if len(candidates) < 8 {
		return nil, fmt.Errorf("insufficient candidates: need at least 8 cards, got %d", len(candidates))
	}

	return candidates, nil
}

func filterDecksByIncludeExclude(decks [][]string, includeCards, excludeCards []string) [][]string {
	if len(decks) == 0 {
		return decks
	}

	includeMap := make(map[string]bool)
	for _, card := range includeCards {
		cardName := strings.TrimSpace(card)
		if cardName == "" {
			continue
		}
		includeMap[cardName] = true
	}

	excludeMap := make(map[string]bool)
	for _, card := range excludeCards {
		cardName := strings.TrimSpace(card)
		if cardName == "" {
			continue
		}
		excludeMap[cardName] = true
	}

	if len(includeMap) == 0 && len(excludeMap) == 0 {
		return decks
	}

	filtered := make([][]string, 0, len(decks))
	for _, deckCards := range decks {
		if !deckContainsAll(deckCards, includeMap) {
			continue
		}
		if deckContainsAny(deckCards, excludeMap) {
			continue
		}
		filtered = append(filtered, deckCards)
	}

	return filtered
}

func deckContainsAll(deckCards []string, required map[string]bool) bool {
	if len(required) == 0 {
		return true
	}
	seen := make(map[string]bool, len(deckCards))
	for _, card := range deckCards {
		seen[card] = true
	}
	for card := range required {
		if !seen[card] {
			return false
		}
	}
	return true
}

func deckContainsAny(deckCards []string, excluded map[string]bool) bool {
	if len(excluded) == 0 {
		return false
	}
	for _, card := range deckCards {
		if excluded[card] {
			return true
		}
	}
	return false
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

// filterResultsByArchetype filters results to only include decks matching specified archetypes
func filterResultsByArchetype(results []FuzzingResult, archetypes []string, _ bool) []FuzzingResult {
	if len(archetypes) == 0 {
		return results
	}

	filtered := make([]FuzzingResult, 0, len(results))
	archetypeSet := make(map[string]bool, len(archetypes))
	for _, arch := range archetypes {
		archetypeSet[arch] = true
	}

	for _, result := range results {
		if archetypeSet[result.Archetype] {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// ensureArchetypeCoverage ensures the top results include at least one deck from each archetype.
// It reorders results to guarantee archetype diversity while preserving score-based ranking as much as possible.
//
//nolint:funlen,gocognit,gocyclo // Selection heuristics intentionally explicit for readability.
func ensureArchetypeCoverage(results []FuzzingResult, top int, verbose bool) []FuzzingResult {
	if len(results) == 0 {
		return results
	}

	// Define all archetypes we want to cover
	allArchetypes := []string{"beatdown", "control", "cycle", "bridge", "siege", "bait", "graveyard", "miner", "hybrid", "unknown"}

	// Group results by archetype
	archetypeGroups := make(map[string][]FuzzingResult)
	for _, arch := range allArchetypes {
		archetypeGroups[arch] = make([]FuzzingResult, 0)
	}

	for _, result := range results {
		arch := result.Archetype
		if _, exists := archetypeGroups[arch]; !exists {
			arch = "unknown"
		}
		archetypeGroups[arch] = append(archetypeGroups[arch], result)
	}

	// Count how many archetypes have at least one deck
	coveredArchetypes := 0
	for _, arch := range allArchetypes {
		if len(archetypeGroups[arch]) > 0 {
			coveredArchetypes++
		}
	}

	if verbose {
		fprintf(os.Stderr, "Archetype coverage: %d/%d archetypes represented in results\n", coveredArchetypes, len(allArchetypes))
	}

	// Build the final result list with archetype diversity
	// Strategy: Round-robin selection from each archetype group, taking the best deck from each
	// archetype in turn, until we've filled the top N slots or exhausted all decks
	finalResults := make([]FuzzingResult, 0, len(results))
	usedDecks := make(map[string]bool) // Track used decks by their key

	// First pass: ensure at least one from each archetype that has decks
	for _, arch := range allArchetypes {
		group := archetypeGroups[arch]
		if len(group) == 0 {
			continue
		}

		// Find the first unused deck from this archetype
		for _, result := range group {
			key := deckKeyForResult(result)
			if !usedDecks[key] {
				finalResults = append(finalResults, result)
				usedDecks[key] = true
				break
			}
		}
	}

	// Second pass: fill remaining slots with the best remaining decks from any archetype
	for _, result := range results {
		key := deckKeyForResult(result)
		if usedDecks[key] {
			continue
		}
		finalResults = append(finalResults, result)
		usedDecks[key] = true
	}

	if verbose {
		// Count how many archetypes are represented in the top N
		topN := min(top, len(finalResults))
		topArchetypes := make(map[string]int)
		for i := range topN {
			topArchetypes[finalResults[i].Archetype]++
		}
		fprintf(os.Stderr, "Top %d decks include %d different archetypes: ", topN, len(topArchetypes))
		first := true
		for arch, count := range topArchetypes {
			if !first {
				fprintf(os.Stderr, ", ")
			}
			fprintf(os.Stderr, "%s=%d", arch, count)
			first = false
		}
		fprintln(os.Stderr)
	}

	return finalResults
}

const (
	elixirBucketLow    = "low"
	elixirBucketMedium = "medium"
	elixirBucketHigh   = "high"
)

func getElixirBucket(avgElixir float64) string {
	switch {
	case avgElixir < 3.3:
		return elixirBucketLow
	case avgElixir <= 4.0:
		return elixirBucketMedium
	default:
		return elixirBucketHigh
	}
}

// ensureElixirBucketDistribution ensures top results include decks from low/medium/high elixir buckets.
//
//nolint:gocyclo,funlen // Bucket balancing logic is branch-heavy by design.
func ensureElixirBucketDistribution(results []FuzzingResult, top int, verbose bool) []FuzzingResult {
	if len(results) == 0 {
		return results
	}

	bucketOrder := []string{elixirBucketLow, elixirBucketMedium, elixirBucketHigh}
	bucketGroups := make(map[string][]FuzzingResult, len(bucketOrder))
	for _, bucket := range bucketOrder {
		bucketGroups[bucket] = make([]FuzzingResult, 0)
	}

	for _, result := range results {
		bucket := getElixirBucket(result.AvgElixir)
		bucketGroups[bucket] = append(bucketGroups[bucket], result)
	}

	if verbose {
		fprintf(os.Stderr, "Elixir bucket distribution in candidates: low=%d, medium=%d, high=%d\n",
			len(bucketGroups[elixirBucketLow]), len(bucketGroups[elixirBucketMedium]), len(bucketGroups[elixirBucketHigh]))
	}

	finalResults := make([]FuzzingResult, 0, len(results))
	usedDecks := make(map[string]bool, len(results))

	// First pass: ensure at least one from each bucket that has decks.
	for _, bucket := range bucketOrder {
		group := bucketGroups[bucket]
		if len(group) == 0 {
			continue
		}
		for _, result := range group {
			key := deckKeyForResult(result)
			if !usedDecks[key] {
				finalResults = append(finalResults, result)
				usedDecks[key] = true
				break
			}
		}
	}

	// Second pass: fill remaining slots with best remaining decks.
	for _, result := range results {
		key := deckKeyForResult(result)
		if usedDecks[key] {
			continue
		}
		finalResults = append(finalResults, result)
		usedDecks[key] = true
	}

	if verbose {
		topN := min(top, len(finalResults))
		topBuckets := map[string]int{
			elixirBucketLow:    0,
			elixirBucketMedium: 0,
			elixirBucketHigh:   0,
		}
		for i := range topN {
			bucket := getElixirBucket(finalResults[i].AvgElixir)
			topBuckets[bucket]++
		}
		fprintf(os.Stderr, "Top %d decks elixir buckets: low=%d, medium=%d, high=%d\n",
			topN, topBuckets[elixirBucketLow], topBuckets[elixirBucketMedium], topBuckets[elixirBucketHigh])
	}

	return finalResults
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
	mode string,
	generationTime time.Duration,
	stats *deck.FuzzingStats,
	totalFiltered int,
) error {
	switch format {
	case "json":
		return formatResultsJSON(results, playerName, playerTag, fuzzerConfig, mode, generationTime, stats, totalFiltered)
	case "csv":
		return formatResultsCSV(results)
	case "detailed":
		return formatResultsDetailed(results, playerName, playerTag)
	default:
		return formatResultsSummary(results, playerName, playerTag, fuzzerConfig, mode, generationTime, stats, totalFiltered)
	}
}

// formatResultsSummary outputs results in summary format
func formatResultsSummary(
	results []FuzzingResult,
	playerName string,
	playerTag string,
	fuzzerConfig *deck.FuzzingConfig,
	mode string,
	generationTime time.Duration,
	stats *deck.FuzzingStats,
	totalFiltered int,
) error {
	printf("\nDeck Fuzzing Results for %s (%s)\n", playerName, playerTag)
	printf("Generated %d decks in %v\n", stats.Generated, generationTime.Round(time.Millisecond))
	printf("Configuration:\n")
	if mode != "" {
		printf("  Mode: %s\n", mode)
	}

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
	fprintln(w, "Rank\tDeck\tOverall\tLadder\tNorm\tAttack\tDefense\tSynergy\tElixir")

	// Print each deck with all 8 cards
	for i, result := range results {
		// Format deck with all cards (no truncation)
		deckStr := strings.Join(result.Deck, ", ")

		// If deck string is very long, use multi-line format
		if len(deckStr) > 50 {
			// First line: Rank, first 4 cards, scores
			firstLine := strings.Join(result.Deck[:4], ", ")
			fprintf(w, "%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\n",
				i+1,
				firstLine+",",
				result.OverallScore,
				result.LadderScore,
				result.NormalizedScore,
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
			fprintf(w, "%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\n",
				i+1,
				deckStr,
				result.OverallScore,
				result.LadderScore,
				result.NormalizedScore,
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
	mode string,
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
			"mode":              mode,
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
	header := []string{"Rank", "Deck", "Overall", "Contextual", "Ladder", "Normalized", "LevelRatio", "NormFactor", "Attack", "Defense", "Synergy", "Versatility", "AvgElixir", "Archetype"}
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
			fmt.Sprintf("%.2f", result.ContextualScore),
			fmt.Sprintf("%.2f", result.LadderScore),
			fmt.Sprintf("%.2f", result.NormalizedScore),
			fmt.Sprintf("%.3f", result.DeckLevelRatio),
			fmt.Sprintf("%.3f", result.NormalizationFactor),
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
		printf("Contextual: %.2f | Ladder: %.2f | Normalized: %.2f\n",
			result.ContextualScore, result.LadderScore, result.NormalizedScore)
		printf("Level Ratio: %.3f | Normalization Factor: %.3f\n",
			result.DeckLevelRatio, result.NormalizationFactor)
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
		return formatResultsJSON(results, cleanTag, playerTag, config, "unknown", 0, stats, len(results))
	case "csv":
		return formatResultsCSV(results)
	default:
		return formatResultsSummary(results, cleanTag, playerTag, &deck.FuzzingConfig{}, "unknown", 0, &deck.FuzzingStats{}, len(results))
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
	maxSameArchetype := cmd.Int("max-same-archetype")
	format := cmd.String("format")
	playerTag := cmd.String("tag")
	workers := cmd.Int("workers")
	verbose := cmd.Bool("verbose")

	if workers == 1 && playerTag != "" {
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
	histogram, err := storage.ArchetypeHistogram(queryOpts)
	if err != nil {
		return fmt.Errorf("failed to query archetype histogram: %w", err)
	}

	var theoreticalByID map[int]fuzzstorage.DeckEntry
	if playerTag != "" && len(decks) > 0 {
		apiToken, err := requireAPIToken(cmd, apiTokenRequirement{
			Reason: "to load player context",
		})
		if err != nil {
			return err
		}

		client := clashroyale.NewClient(apiToken)
		cleanTag := strings.TrimPrefix(playerTag, "#")
		player, playerErr := client.GetPlayerWithContext(ctx, cleanTag)
		if playerErr != nil {
			return fmt.Errorf("failed to load player data for %s: %w", playerTag, playerErr)
		}
		playerContext := evaluation.NewPlayerContextFromPlayer(player)

		theoreticalByID = make(map[int]fuzzstorage.DeckEntry, len(decks))
		for _, deck := range decks {
			theoreticalByID[deck.ID] = deck
		}

		decks = reevaluateStoredDecks(decks, player, player.Tag, playerContext, workers, verbose)
		sort.Slice(decks, func(i, j int) bool {
			return decks[i].OverallScore > decks[j].OverallScore
		})

		if maxSameArchetype > 0 {
			decks = limitArchetypeRepetition(decks, maxSameArchetype)
		}

		if verbose {
			printf("Loaded player context for %s (%s)\n", player.Name, player.Tag)
		}
	}
	if maxSameArchetype > 0 {
		decks = limitArchetypeRepetition(decks, maxSameArchetype)
	}

	total, _ := storage.Count()
	dbPath := storage.GetDBPath()

	fprintf(os.Stderr, "Top decks from: %s\n", dbPath)
	fprintf(os.Stderr, "Showing %d of %d total decks\n\n", len(decks), total)

	// Format output
	switch format {
	case "json":
		return formatListResultsJSON(decks, dbPath, total, histogram, theoreticalByID)
	case "csv":
		return formatListResultsCSV(decks, theoreticalByID)
	case "detailed":
		return formatListResultsDetailed(decks, dbPath, total, histogram, theoreticalByID)
	default:
		return formatListResultsSummary(decks, dbPath, total, histogram, theoreticalByID)
	}
}

// deckFuzzUpdateCommand re-evaluates saved decks with current scoring and updates storage.
func deckFuzzUpdateCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")
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

	var player *clashroyale.Player
	var playerContext *evaluation.PlayerContext
	if playerTag != "" {
		apiToken, err := requireAPIToken(cmd, apiTokenRequirement{
			Reason: "to load player context",
		})
		if err != nil {
			return err
		}
		client := clashroyale.NewClient(apiToken)
		var err error
		player, err = client.GetPlayerWithContext(ctx, playerTag)
		if err != nil {
			return fmt.Errorf("failed to load player data for %s: %w", playerTag, err)
		}
		playerContext = evaluation.NewPlayerContextFromPlayer(player)
		if verbose {
			printf("Loaded player context for %s (%s)\n", player.Name, playerTag)
		}
	}

	start := time.Now()
	updatedEntries := reevaluateStoredDecks(entries, player, playerTag, playerContext, workers, verbose)

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

func formatScoreTransition(
	theoreticalByID map[int]fuzzstorage.DeckEntry,
	deckID int,
	current float64,
	extract func(fuzzstorage.DeckEntry) float64,
) string {
	if theoreticalByID == nil {
		return fmt.Sprintf("%.2f", current)
	}
	theoretical, ok := theoreticalByID[deckID]
	if !ok {
		return fmt.Sprintf("%.2f", current)
	}
	return fmt.Sprintf("%.2f->%.2f", extract(theoretical), current)
}

func reevaluateStoredDecks(entries []fuzzstorage.DeckEntry, player *clashroyale.Player, playerTag string, playerContext *evaluation.PlayerContext, workers int, verbose bool) []fuzzstorage.DeckEntry {
	if workers <= 1 {
		return reevaluateStoredDecksSequential(entries, player, playerTag, playerContext, verbose)
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
		wg.Go(func() {
			synergyDB := deck.NewSynergyDatabase()

			for work := range workChan {
				result := evaluateSingleDeck(work.entry.Cards, player, playerTag, synergyDB, playerContext)
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
		})
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

func reevaluateStoredDecksSequential(entries []fuzzstorage.DeckEntry, player *clashroyale.Player, playerTag string, playerContext *evaluation.PlayerContext, verbose bool) []fuzzstorage.DeckEntry {
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
		result := evaluateSingleDeck(entry.Cards, player, playerTag, synergyDB, playerContext)
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
func formatListResultsSummary(
	decks []fuzzstorage.DeckEntry,
	dbPath string,
	total int,
	histogram map[string]int,
	theoreticalByID map[int]fuzzstorage.DeckEntry,
) error {
	printf("Saved Top Decks\n")
	printf("Database: %s\n", dbPath)
	printf("Total decks: %d\n\n", total)

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, ' ', 0)
	fprintln(w, "Rank\tDeck\tOverall\tAttack\tDefense\tSynergy\tElixir\tArchetype")

	for i, deck := range decks {
		deckStr := strings.Join(deck.Cards, ", ")
		overall := formatScoreTransition(theoreticalByID, deck.ID, deck.OverallScore, func(entry fuzzstorage.DeckEntry) float64 { return entry.OverallScore })
		attack := formatScoreTransition(theoreticalByID, deck.ID, deck.AttackScore, func(entry fuzzstorage.DeckEntry) float64 { return entry.AttackScore })
		defense := formatScoreTransition(theoreticalByID, deck.ID, deck.DefenseScore, func(entry fuzzstorage.DeckEntry) float64 { return entry.DefenseScore })
		synergy := formatScoreTransition(theoreticalByID, deck.ID, deck.SynergyScore, func(entry fuzzstorage.DeckEntry) float64 { return entry.SynergyScore })
		if len(deckStr) > 50 {
			firstLine := strings.Join(deck.Cards[:4], ", ")
			fprintf(w, "%d\t%s,\t%s\t%s\t%s\t%s\t%.2f\t%s\n",
				i+1, firstLine, overall, attack, defense, synergy, deck.AvgElixir, deck.Archetype)
			secondLine := strings.Join(deck.Cards[4:], ", ")
			fprintf(w, "\t%s\n", secondLine)
		} else {
			fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%.2f\t%s\n",
				i+1, deckStr, overall, attack, defense, synergy, deck.AvgElixir, deck.Archetype)
		}
	}

	flushWriter(w)

	if len(histogram) > 0 {
		printf("\nArchetype Histogram (matching query):\n")
		printArchetypeHistogram(histogram)
	}
	return nil
}

// formatListResultsJSON formats list results in JSON format
func formatListResultsJSON(
	decks []fuzzstorage.DeckEntry,
	dbPath string,
	total int,
	histogram map[string]int,
	theoreticalByID map[int]fuzzstorage.DeckEntry,
) error {
	results := make([]map[string]any, 0, len(decks))
	for _, deck := range decks {
		result := map[string]any{
			"id":                deck.ID,
			"cards":             deck.Cards,
			"overall_score":     deck.OverallScore,
			"attack_score":      deck.AttackScore,
			"defense_score":     deck.DefenseScore,
			"synergy_score":     deck.SynergyScore,
			"versatility_score": deck.VersatilityScore,
			"avg_elixir":        deck.AvgElixir,
			"archetype":         deck.Archetype,
			"archetype_conf":    deck.ArchetypeConf,
			"evaluated_at":      deck.EvaluatedAt,
		}
		if theoreticalByID != nil {
			if theoretical, ok := theoreticalByID[deck.ID]; ok {
				result["stored_overall_score"] = theoretical.OverallScore
				result["stored_attack_score"] = theoretical.AttackScore
				result["stored_defense_score"] = theoretical.DefenseScore
				result["stored_synergy_score"] = theoretical.SynergyScore
			}
		}
		results = append(results, result)
	}

	output := map[string]any{
		"database":            dbPath,
		"total":               total,
		"returned":            len(decks),
		"results":             results,
		"archetype_histogram": histogram,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// formatListResultsCSV formats list results in CSV format
func formatListResultsCSV(decks []fuzzstorage.DeckEntry, theoreticalByID map[int]fuzzstorage.DeckEntry) error {
	w := csv.NewWriter(os.Stdout)

	header := []string{"Rank", "Deck", "Overall", "Attack", "Defense", "Synergy", "Versatility", "AvgElixir", "Archetype"}
	if theoreticalByID != nil {
		header = []string{
			"Rank", "Deck",
			"StoredOverall", "PlayerOverall",
			"StoredAttack", "PlayerAttack",
			"StoredDefense", "PlayerDefense",
			"StoredSynergy", "PlayerSynergy",
			"Versatility", "AvgElixir", "Archetype",
		}
	}
	if err := w.Write(header); err != nil {
		return err
	}

	for i, deck := range decks {
		deckStr := strings.Join(deck.Cards, ", ")
		row := []string{
			strconv.Itoa(i + 1),
			deckStr,
		}
		if theoreticalByID != nil {
			theoretical := theoreticalByID[deck.ID]
			row = append(row,
				fmt.Sprintf("%.2f", theoretical.OverallScore),
				fmt.Sprintf("%.2f", deck.OverallScore),
				fmt.Sprintf("%.2f", theoretical.AttackScore),
				fmt.Sprintf("%.2f", deck.AttackScore),
				fmt.Sprintf("%.2f", theoretical.DefenseScore),
				fmt.Sprintf("%.2f", deck.DefenseScore),
				fmt.Sprintf("%.2f", theoretical.SynergyScore),
				fmt.Sprintf("%.2f", deck.SynergyScore),
				fmt.Sprintf("%.2f", deck.VersatilityScore),
				fmt.Sprintf("%.2f", deck.AvgElixir),
				deck.Archetype,
			)
		} else {
			row = append(row,
				fmt.Sprintf("%.2f", deck.OverallScore),
				fmt.Sprintf("%.2f", deck.AttackScore),
				fmt.Sprintf("%.2f", deck.DefenseScore),
				fmt.Sprintf("%.2f", deck.SynergyScore),
				fmt.Sprintf("%.2f", deck.VersatilityScore),
				fmt.Sprintf("%.2f", deck.AvgElixir),
				deck.Archetype,
			)
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
func formatListResultsDetailed(
	decks []fuzzstorage.DeckEntry,
	dbPath string,
	total int,
	histogram map[string]int,
	theoreticalByID map[int]fuzzstorage.DeckEntry,
) error {
	printf("Saved Top Decks\n")
	printf("Database: %s\n", dbPath)
	printf("Total decks: %d\n\n", total)

	for i, deck := range decks {
		printf("=== Deck %d ===\n", i+1)
		printf("Cards: %s\n", strings.Join(deck.Cards, ", "))
		if theoreticalByID != nil {
			if theoretical, ok := theoreticalByID[deck.ID]; ok {
				printf("Overall: %.2f -> %.2f | Attack: %.2f -> %.2f | Defense: %.2f -> %.2f | Synergy: %.2f -> %.2f | Versatility: %.2f\n",
					theoretical.OverallScore, deck.OverallScore,
					theoretical.AttackScore, deck.AttackScore,
					theoretical.DefenseScore, deck.DefenseScore,
					theoretical.SynergyScore, deck.SynergyScore,
					deck.VersatilityScore,
				)
			} else {
				printf("Overall: %.2f | Attack: %.2f | Defense: %.2f | Synergy: %.2f | Versatility: %.2f\n",
					deck.OverallScore, deck.AttackScore, deck.DefenseScore, deck.SynergyScore, deck.VersatilityScore)
			}
		} else {
			printf("Overall: %.2f | Attack: %.2f | Defense: %.2f | Synergy: %.2f | Versatility: %.2f\n",
				deck.OverallScore, deck.AttackScore, deck.DefenseScore, deck.SynergyScore, deck.VersatilityScore)
		}
		printf("Avg Elixir: %.2f | Archetype: %s (%.0f%% confidence)\n",
			deck.AvgElixir, deck.Archetype, deck.ArchetypeConf*100)
		printf("Evaluated: %s\n\n", deck.EvaluatedAt.Format(time.RFC3339))
	}

	if len(histogram) > 0 {
		printf("Archetype Histogram (matching query):\n")
		printArchetypeHistogram(histogram)
	}

	return nil
}

func printArchetypeHistogram(histogram map[string]int) {
	type entry struct {
		archetype string
		count     int
	}

	entries := make([]entry, 0, len(histogram))
	for archetype, count := range histogram {
		entries = append(entries, entry{archetype: archetype, count: count})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count == entries[j].count {
			return entries[i].archetype < entries[j].archetype
		}
		return entries[i].count > entries[j].count
	})

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, ' ', 0)
	fprintln(w, "Archetype\tCount")
	fprintln(w, "---------\t-----")
	for _, e := range entries {
		fprintf(w, "%s\t%d\n", e.archetype, e.count)
	}
	flushWriter(w)
}

func limitArchetypeRepetition(decks []fuzzstorage.DeckEntry, maxPerArchetype int) []fuzzstorage.DeckEntry {
	if maxPerArchetype <= 0 {
		return decks
	}

	counts := make(map[string]int, len(decks))
	filtered := make([]fuzzstorage.DeckEntry, 0, len(decks))
	for _, deck := range decks {
		if counts[deck.Archetype] >= maxPerArchetype {
			continue
		}
		counts[deck.Archetype]++
		filtered = append(filtered, deck)
	}
	return filtered
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
//
//nolint:gocognit,gocyclo // Mutation pipeline uses explicit branching for reproducibility.
func generateDeckMutations(savedDecks [][]string, player *clashroyale.Player, count, mutationIntensity int, verbose bool) [][]string {
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
	mutationsPerDeck := max(count/len(savedDecks), 1)

	for _, deck := range savedDecks {
		for i := 0; i < mutationsPerDeck; i++ {
			// Create mutation by swapping 1-2 random cards
			mutation := make([]string, len(deck))
			copy(mutation, deck)

			// Swap cards based on mutation intensity
			numSwaps := 1 + (i % mutationIntensity) // Vary from 1 to mutationIntensity
			for range numSwaps {
				// Find cards to swap
				swapIdx := i % len(mutation)

				// Find a replacement card
				for _, card := range player.Cards {
					// Skip if card is already in deck
					alreadyInDeck := slices.Contains(mutation, card.Name)
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
//
//nolint:gocyclo // Variation generation includes multiple guarded mutation paths.
func generateVariations(baseDeck []string, player *clashroyale.Player, count, mutationIntensity int, verbose bool) [][]string {
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
	for i := range count {
		variation := make([]string, len(baseDeck))
		copy(variation, baseDeck)

		// Number of cards to swap (1-mutationIntensity, varying across variations)
		numSwaps := 1 + (i % mutationIntensity)

		// Swap random positions with available cards
		for j := range numSwaps {
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

func formatDurationFloor(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", int(seconds))
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
