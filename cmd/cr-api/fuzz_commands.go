package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/klauer/clash-royale-api/go/pkg/leaderboard"
	"github.com/urfave/cli/v3"
)

// deckFuzzCommand is the action function for the deck fuzz command
func deckFuzzCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")
	count := cmd.Int("count")
	workers := cmd.Int("workers")
	includeCards := cmd.StringSlice("include-cards")
	excludeCards := cmd.StringSlice("exclude-cards")
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
		fmt.Fprintf(os.Stderr, "Loaded player: %s (%s)\n", playerName, player.Tag)
		fmt.Fprintf(os.Stderr, "Cards available: %d\n", len(player.Cards))
	}

	// Initialize fuzzer configuration
	fuzzerCfg := &deck.FuzzingConfig{
		Count:           count,
		Workers:         workers,
		IncludeCards:    includeCards,
		ExcludeCards:    excludeCards,
		MinAvgElixir:    minElixir,
		MaxAvgElixir:    maxElixir,
		MinOverallScore: minOverall,
		MinSynergyScore: minSynergy,
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

	if verbose {
		fmt.Fprintf(os.Stderr, "\nStarting deck fuzzing...\n")
		fmt.Fprintf(os.Stderr, "Configuration:\n")
		fmt.Fprintf(os.Stderr, "  Count: %d\n", count)
		fmt.Fprintf(os.Stderr, "  Workers: %d\n", workers)
		if len(includeCards) > 0 {
			fmt.Fprintf(os.Stderr, "  Include cards: %s\n", strings.Join(includeCards, ", "))
		}
		if len(excludeCards) > 0 {
			fmt.Fprintf(os.Stderr, "  Exclude cards: %s\n", strings.Join(excludeCards, ", "))
		}
		fmt.Fprintf(os.Stderr, "  Elixir range: %.1f - %.1f\n", minElixir, maxElixir)
		fmt.Fprintf(os.Stderr, "  Min overall score: %.1f\n", minOverall)
		fmt.Fprintf(os.Stderr, "  Min synergy score: %.1f\n", minSynergy)
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Generate decks
	startTime := time.Now()

	var generatedDecks [][]string
	if workers > 1 {
		generatedDecks, err = fuzzer.GenerateDecksParallel()
	} else {
		generatedDecks, err = fuzzer.GenerateDecks(count)
	}

	if err != nil {
		return fmt.Errorf("failed to generate decks: %w", err)
	}

	generationTime := time.Since(startTime)

	stats := fuzzer.GetStats()

	if verbose {
		fmt.Fprintf(os.Stderr, "Generated %d decks in %v\n", len(generatedDecks), generationTime)
		fmt.Fprintf(os.Stderr, "Success: %d, Failed: %d\n", stats.Success, stats.Failed)
		if stats.SkippedElixir > 0 {
			fmt.Fprintf(os.Stderr, "Skipped (elixir): %d\n", stats.SkippedElixir)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	if len(generatedDecks) == 0 {
		return fmt.Errorf("no decks were successfully generated")
	}

	// Evaluate decks
	if verbose {
		fmt.Fprintf(os.Stderr, "Evaluating %d decks...\n", len(generatedDecks))
	}

	evaluationResults := evaluateGeneratedDecks(
		generatedDecks,
		player,
		playerTag,
		storagePath,
		verbose,
	)

	// Filter by score thresholds
	filteredResults := filterResultsByScore(evaluationResults, minOverall, minSynergy, verbose)

	if len(filteredResults) == 0 {
		return fmt.Errorf("no decks passed the score filters (min-overall: %.1f, min-synergy: %.1f)", minOverall, minSynergy)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "%d decks passed score filters\n", len(filteredResults))
	}

	// Sort results
	sortFuzzingResults(filteredResults, sortBy)

	// Get top N results
	topResults := getTopResults(filteredResults, top)

	// Format and output results
	if err := formatFuzzingResults(topResults, format, playerName, playerTag, fuzzerCfg, generationTime, &stats, len(filteredResults)); err != nil {
		return fmt.Errorf("failed to format results: %w", err)
	}

	// Save to file if output-dir specified
	if outputDir != "" {
		if err := saveResultsToFile(topResults, outputDir, format, playerTag); err != nil {
			return fmt.Errorf("failed to save results: %w", err)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "\nResults saved to %s\n", outputDir)
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
	decks [][]string,
	player *clashroyale.Player,
	playerTag string,
	storagePath string,
	verbose bool,
) []FuzzingResult {
	results := make([]FuzzingResult, 0, len(decks))

	// Create synergy database once
	synergyDB := deck.NewSynergyDatabase()

	// Create player context if player tag provided
	var playerContext *evaluation.PlayerContext
	if playerTag != "" && player != nil {
		playerContext = evaluation.NewPlayerContextFromPlayer(player)
	}

	var storage *leaderboard.Storage
	var storageErr error
	if storagePath != "" {
		storage, storageErr = leaderboard.NewStorage(storagePath)
		if storageErr != nil && verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to open storage: %v\n", storageErr)
		}
		if storage != nil {
			defer storage.Close()
		}
	}

	// Evaluate each deck
	for i, deckCards := range decks {
		if verbose && (i+1)%100 == 0 {
			fmt.Fprintf(os.Stderr, "  Evaluated %d/%d decks...\n", i+1, len(decks))
		}

		// Convert deck strings to CardCandidates
		candidates := convertDeckToCandidates(deckCards, player)

		// Run evaluation
		evalResult := evaluation.Evaluate(candidates, synergyDB, playerContext)

		result := FuzzingResult{
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

		results = append(results, result)

		// Save to persistent storage if available
		if storage != nil {
			entry := &leaderboard.DeckEntry{
				Cards:             deckCards,
				OverallScore:      evalResult.OverallScore,
				AttackScore:       evalResult.Attack.Score,
				DefenseScore:      evalResult.Defense.Score,
				SynergyScore:      evalResult.Synergy.Score,
				VersatilityScore:  evalResult.Versatility.Score,
				F2PScore:          evalResult.F2PFriendly.Score,
				PlayabilityScore:  evalResult.Playability.Score,
				Archetype:         string(evalResult.DetectedArchetype),
				ArchetypeConf:     evalResult.ArchetypeConfidence,
				AvgElixir:         evalResult.AvgElixir,
				EvaluatedAt:       result.EvaluatedAt,
				PlayerTag:         playerTag,
				EvaluationVersion: "1.0.0",
			}
			storage.InsertDeck(entry)
		}
	}

	return results
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
func filterResultsByScore(results []FuzzingResult, minOverall, minSynergy float64, verbose bool) []FuzzingResult {
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
	fmt.Printf("\nDeck Fuzzing Results for %s (%s)\n", playerName, playerTag)
	fmt.Printf("Generated %d random decks in %v\n", stats.Generated, generationTime.Round(time.Millisecond))
	fmt.Printf("Configuration:\n")

	if len(fuzzerConfig.IncludeCards) > 0 {
		fmt.Printf("  Include cards: %s\n", strings.Join(fuzzerConfig.IncludeCards, ", "))
	}
	if len(fuzzerConfig.ExcludeCards) > 0 {
		fmt.Printf("  Exclude cards: %s\n", strings.Join(fuzzerConfig.ExcludeCards, ", "))
	}
	fmt.Printf("  Elixir range: %.1f - %.1f\n", fuzzerConfig.MinAvgElixir, fuzzerConfig.MaxAvgElixir)
	if fuzzerConfig.MinOverallScore > 0 {
		fmt.Printf("  Min overall score: %.1f\n", fuzzerConfig.MinOverallScore)
	}
	if fuzzerConfig.MinSynergyScore > 0 {
		fmt.Printf("  Min synergy score: %.1f\n", fuzzerConfig.MinSynergyScore)
	}

	fmt.Printf("\nTop %d Decks (from %d decks passing filters):\n\n", len(results), totalFiltered)

	// Print table header
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, "Rank\tDeck\tOverall\tAttack\tDefense\tSynergy\tElixir")

	// Print each deck
	for i, result := range results {
		deckStr := formatDeckString(result.Deck, 40)
		fmt.Fprintf(w, "%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\n",
			i+1,
			deckStr,
			result.OverallScore,
			result.AttackScore,
			result.DefenseScore,
			result.SynergyScore,
			result.AvgElixir,
		)
	}

	w.Flush()

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
	output := map[string]interface{}{
		"player_name":             playerName,
		"player_tag":              playerTag,
		"generated":               stats.Generated,
		"success":                 stats.Success,
		"failed":                  stats.Failed,
		"filtered":                totalFiltered,
		"returned":                len(results),
		"generation_time_seconds": generationTime.Seconds(),
		"config": map[string]interface{}{
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
	defer w.Flush()

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

	return nil
}

// formatResultsDetailed outputs results in detailed format with full evaluation
func formatResultsDetailed(
	results []FuzzingResult,
	playerName string,
	playerTag string,
) error {
	fmt.Printf("\nDeck Fuzzing Results for %s (%s)\n", playerName, playerTag)
	fmt.Printf("\nTop %d Decks:\n\n", len(results))

	for i, result := range results {
		fmt.Printf("=== Deck %d ===\n", i+1)
		fmt.Printf("Cards: %s\n", strings.Join(result.Deck, ", "))
		fmt.Printf("Overall: %.2f | Attack: %.2f | Defense: %.2f | Synergy: %.2f | Versatility: %.2f\n",
			result.OverallScore, result.AttackScore, result.DefenseScore, result.SynergyScore, result.VersatilityScore)
		fmt.Printf("Avg Elixir: %.2f | Archetype: %s (%.0f%% confidence)\n",
			result.AvgElixir, result.Archetype, result.ArchetypeConfidence*100)
		fmt.Printf("Evaluated: %s\n\n", result.EvaluatedAt.Format(time.RFC3339))
	}

	return nil
}

// formatDeckString formats a deck as a truncated string
func formatDeckString(deck []string, maxLen int) string {
	deckStr := strings.Join(deck, ", ")
	if len(deckStr) <= maxLen {
		return deckStr
	}
	return deckStr[:maxLen-3] + "..."
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
	defer file.Close()

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

// printFuzzingProgress prints real-time progress during fuzzing
func printFuzzingProgress(generated, total int, startTime time.Time) {
	if total == 0 {
		return
	}

	elapsed := time.Since(startTime)
	remaining := time.Duration(float64(elapsed) / float64(generated) * float64(total-generated))

	percent := float64(generated) / float64(total) * 100
	fmt.Fprintf(os.Stderr, "\rProgress: %d/%d (%.1f%%) | Elapsed: %v | ETA: %v",
		generated, total, percent, elapsed.Round(time.Second), remaining.Round(time.Second))
}
