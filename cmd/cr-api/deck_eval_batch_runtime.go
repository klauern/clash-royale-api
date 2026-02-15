package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/klauer/clash-royale-api/go/pkg/leaderboard"
	"github.com/urfave/cli/v3"
)

func parseEvalBatchFlags(cmd *cli.Command) (*evalBatchFlags, error) {
	flags := &evalBatchFlags{
		FromSuite:       cmd.String("from-suite"),
		DeckDir:         cmd.String("deck-dir"),
		PlayerTag:       cmd.String("tag"),
		Format:          cmd.String("format"),
		OutputDir:       cmd.String("output-dir"),
		SortBy:          cmd.String("sort-by"),
		TopOnly:         cmd.Bool("top-only"),
		TopN:            cmd.Int("top-n"),
		FilterArchetype: cmd.Bool("filter-archetype"),
		ArchetypeFilter: cmd.String("archetype"),
		Verbose:         cmd.Bool("verbose"),
		ShowTiming:      cmd.Bool("timing"),
		SaveAggregated:  cmd.Bool("save-aggregated"),
	}

	if flags.FromSuite == "" && flags.DeckDir == "" {
		return nil, fmt.Errorf("must provide either --from-suite or --deck-dir")
	}

	if flags.FromSuite != "" && flags.DeckDir != "" {
		return nil, fmt.Errorf("cannot use both --from-suite and --deck-dir")
	}

	return flags, nil
}

// loadEvalDecks loads decks from either a suite file or directory
func loadEvalDecks(fromSuite, deckDir string, verbose bool) ([]evalDeckInfo, string, string, error) {
	if fromSuite != "" {
		return loadEvalDecksFromSuite(fromSuite, verbose)
	}
	return loadEvalDecksFromDirectory(deckDir, verbose)
}

// loadEvalDecksFromSuite loads deck information from a suite summary JSON file
func loadEvalDecksFromSuite(fromSuite string, verbose bool) ([]evalDeckInfo, string, string, error) {
	data, err := os.ReadFile(fromSuite)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to read suite file: %w", err)
	}

	var suiteData map[string]any
	if err := json.Unmarshal(data, &suiteData); err != nil {
		return nil, "", "", fmt.Errorf("failed to parse suite JSON: %w", err)
	}

	playerName, playerTag := extractSuitePlayerInfo(suiteData)
	decks := extractSuiteDecks(suiteData)

	if verbose {
		printf("Loaded %d decks from suite: %s\n", len(decks), fromSuite)
	}

	return decks, playerName, playerTag, nil
}

// extractSuitePlayerInfo extracts player name and tag from suite data
func extractSuitePlayerInfo(suiteData map[string]any) (string, string) {
	playerInfo, ok := suiteData["player"].(map[string]any)
	if !ok {
		return "", ""
	}

	var playerName, playerTag string
	if name, ok := playerInfo["name"].(string); ok {
		playerName = name
	}
	if tag, ok := playerInfo["tag"].(string); ok {
		playerTag = tag
	}
	return playerName, playerTag
}

// extractSuiteDecks extracts deck information from suite data
func extractSuiteDecks(suiteData map[string]any) []evalDeckInfo {
	decksList, ok := suiteData["decks"].([]any)
	if !ok {
		return nil
	}

	var decks []evalDeckInfo
	for i, d := range decksList {
		deckMap, ok := d.(map[string]any)
		if !ok {
			continue
		}
		decks = append(decks, parseSuiteDeckEntry(deckMap, i))
	}
	return decks
}

// parseSuiteDeckEntry parses a single deck entry from suite data
func parseSuiteDeckEntry(deckMap map[string]any, index int) evalDeckInfo {
	cards := extractStringSlice(deckMap["cards"])
	strategy := extractString(deckMap["strategy"], "unknown")
	variation := extractInt(deckMap["variation"])
	filePath := extractString(deckMap["file_path"], "")

	name := fmt.Sprintf("Deck #%d (%s v%d)", index+1, strategy, variation)
	return evalDeckInfo{
		Name:     name,
		Cards:    cards,
		Strategy: strategy,
		FilePath: filePath,
	}
}

// extractStringSlice extracts a string slice from an any value
func extractStringSlice(v any) []string {
	list, ok := v.([]any)
	if !ok {
		return nil
	}
	var result []string
	for _, item := range list {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// extractString extracts a string value with a default
func extractString(v any, defaultVal string) string {
	if s, ok := v.(string); ok {
		return s
	}
	return defaultVal
}

// extractInt extracts an int value from a float64
func extractInt(v any) int {
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}

// loadEvalDecksFromDirectory loads deck information from JSON files in a directory
func loadEvalDecksFromDirectory(deckDir string, verbose bool) ([]evalDeckInfo, string, string, error) {
	entries, err := os.ReadDir(deckDir)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to read deck directory: %w", err)
	}

	var decks []evalDeckInfo
	var playerName string

	for i, entry := range entries {
		if !isJSONFile(entry) {
			continue
		}

		deck, name := loadDeckFromFile(entry, deckDir, verbose, i == 0)
		if deck == nil {
			continue
		}

		decks = append(decks, *deck)
		if i == 0 {
			playerName = name
		}
	}

	if verbose {
		printf("Loaded %d decks from directory: %s\n", len(decks), deckDir)
	}

	return decks, playerName, "", nil
}

// isJSONFile checks if a directory entry is a JSON file
func isJSONFile(entry os.DirEntry) bool {
	return !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json")
}

// loadDeckFromFile loads a single deck from a JSON file
func loadDeckFromFile(entry os.DirEntry, deckDir string, verbose, extractPlayer bool) (*evalDeckInfo, string) {
	deckPath := filepath.Join(deckDir, entry.Name())
	data, err := os.ReadFile(deckPath)
	if err != nil {
		if verbose {
			printf("Warning: Failed to read %s: %v\n", entry.Name(), err)
		}
		return nil, ""
	}

	var deckData map[string]any
	if err := json.Unmarshal(data, &deckData); err != nil {
		if verbose {
			printf("Warning: Failed to parse %s: %v\n", entry.Name(), err)
		}
		return nil, ""
	}

	cards := extractEvalCardsFromDeckData(deckData)
	if len(cards) != 8 {
		if verbose {
			printf("Warning: Skipping %s (expected 8 cards, got %d)\n", entry.Name(), len(cards))
		}
		return nil, ""
	}

	name := strings.TrimSuffix(entry.Name(), ".json")
	deck := &evalDeckInfo{
		Name:     name,
		Cards:    cards,
		FilePath: deckPath,
	}

	var playerName string
	if extractPlayer {
		playerName = extractEvalPlayerNameFromDeckData(deckData)
	}

	return deck, playerName
}

// extractEvalCardsFromDeckData extracts card names from deck data in various formats
func extractEvalCardsFromDeckData(deckData map[string]any) []string {
	var cards []string

	if deckMap, ok := deckData["deck"].([]any); ok {
		for _, c := range deckMap {
			if cardStr, ok := c.(string); ok {
				cards = append(cards, cardStr)
			}
		}
	} else if cardsList, ok := deckData["cards"].([]any); ok {
		for _, c := range cardsList {
			if cardStr, ok := c.(string); ok {
				cards = append(cards, cardStr)
			}
		}
	}

	return cards
}

// extractEvalPlayerNameFromDeckData extracts player name from deck data if available
func extractEvalPlayerNameFromDeckData(deckData map[string]any) string {
	if rec, ok := deckData["recommendation"].(map[string]any); ok {
		if pname, ok := rec["player_name"].(string); ok {
			return pname
		}
	}
	return ""
}

// loadEvalPlayerContext loads player context from API if tag and token are available
func loadEvalPlayerContext(playerTag, apiToken string, verbose bool) (*evaluation.PlayerContext, string, error) {
	if playerTag == "" || apiToken == "" {
		return nil, "", nil
	}

	client := clashroyale.NewClient(apiToken)
	player, err := client.GetPlayer(playerTag)
	if err != nil {
		if verbose {
			printf("Warning: Failed to load player data: %v\n", err)
			fmt.Println("Continuing with generic evaluation (no player context)")
		}
		return nil, "", nil
	}

	playerContext := evaluation.NewPlayerContextFromPlayer(player)
	if verbose {
		printf("Loaded player context for %s (%s)\n", player.Name, playerTag)
	}

	return playerContext, player.Name, nil
}

// initEvalStorage creates persistent storage for evaluation results
func initEvalStorage(playerTag string, verbose bool) (*leaderboard.Storage, error) {
	if playerTag == "" {
		return nil, nil
	}

	storage, err := leaderboard.NewStorage(playerTag)
	if err != nil {
		return nil, err
	}

	if verbose {
		printf("Initialized persistent storage at: %s\n", storage.GetDBPath())
	}

	return storage, nil
}

// runEvalDecksBatch evaluates all decks and returns results
func runEvalDecksBatch(
	decks []evalDeckInfo,
	synergyDB *deck.SynergyDatabase,
	playerContext *evaluation.PlayerContext,
	storage *leaderboard.Storage,
	playerTag string,
	verbose bool,
) ([]evalBatchResult, time.Duration, error) {
	results := make([]evalBatchResult, 0, len(decks))
	startTime := time.Now()

	for i, deckData := range decks {
		result := evalSingleDeck(i, deckData, len(decks), synergyDB, playerContext, storage, playerTag, verbose)
		if result != nil {
			results = append(results, *result)
		}
	}

	return results, time.Since(startTime), nil
}

// evalSingleDeck evaluates a single deck and saves to storage if available
func evalSingleDeck(
	index int,
	deckData evalDeckInfo,
	totalDecks int,
	synergyDB *deck.SynergyDatabase,
	playerContext *evaluation.PlayerContext,
	storage *leaderboard.Storage,
	playerTag string,
	verbose bool,
) *evalBatchResult {
	deckStart := time.Now()

	if len(deckData.Cards) != 8 {
		if verbose {
			printf("  [%d/%d] Skipping %s: invalid card count (%d)\n",
				index+1, totalDecks, deckData.Name, len(deckData.Cards))
		}
		return nil
	}

	deckCards := loadEvalDeckCards(deckData, verbose)
	result := evaluation.Evaluate(deckCards, synergyDB, playerContext)
	elapsed := time.Since(deckStart)

	if storage != nil {
		saveEvalDeckToStorage(storage, result, deckData, deckStart, playerTag, verbose)
	}

	if verbose {
		printf("  [%d/%d] %s: %.2f (%s) - %s\n",
			index+1, totalDecks, deckData.Name, result.OverallScore,
			result.OverallRating, result.DetectedArchetype)
	}

	return &evalBatchResult{
		Name:      deckData.Name,
		Strategy:  deckData.Strategy,
		Deck:      result.Deck,
		Result:    result,
		FilePath:  deckData.FilePath,
		Evaluated: deckStart,
		Duration:  elapsed,
	}
}

// loadEvalDeckCards loads card candidates from file or converts from names
func loadEvalDeckCards(deckData evalDeckInfo, verbose bool) []deck.CardCandidate {
	if deckData.FilePath != "" {
		candidates, ok, err := loadDeckCandidatesFromFile(deckData.FilePath)
		if err != nil && verbose {
			printf("Warning: Failed to load deck details from %s: %v\n", deckData.FilePath, err)
		}
		if ok {
			return candidates
		}
	}
	return convertToCardCandidates(deckData.Cards)
}

// saveEvalDeckToStorage saves a deck evaluation result to persistent storage
func saveEvalDeckToStorage(
	storage *leaderboard.Storage,
	result evaluation.EvaluationResult,
	deckData evalDeckInfo,
	evaluatedAt time.Time,
	playerTag string,
	verbose bool,
) {
	entry := &leaderboard.DeckEntry{
		Cards:             result.Deck,
		OverallScore:      result.OverallScore,
		AttackScore:       result.Attack.Score,
		DefenseScore:      result.Defense.Score,
		SynergyScore:      result.Synergy.Score,
		VersatilityScore:  result.Versatility.Score,
		F2PScore:          result.F2PFriendly.Score,
		PlayabilityScore:  result.Playability.Score,
		Archetype:         string(result.DetectedArchetype),
		ArchetypeConf:     result.ArchetypeConfidence,
		Strategy:          deckData.Strategy,
		AvgElixir:         result.AvgElixir,
		EvaluatedAt:       evaluatedAt,
		PlayerTag:         playerTag,
		EvaluationVersion: "1.0.0",
	}

	_, isNew, err := storage.InsertDeck(entry)
	if err != nil && verbose {
		fprintf(os.Stderr, "  Warning: failed to save deck to storage: %v\n", err)
	} else if verbose && !isNew {
		printf("  (deck already in storage, updated)\n")
	}
}

// printEvalTimingSummary prints evaluation timing information
func printEvalTimingSummary(results []evalBatchResult, totalTime time.Duration, showTiming, verbose bool) {
	if !verbose && !showTiming {
		return
	}
	printf("\nBatch evaluation completed in %v\n", totalTime)
	if len(results) > 0 {
		printf("Average time per deck: %v\n", totalTime/time.Duration(len(results)))
	}
}

// updateEvalStorageStats recalculates and displays storage statistics
func updateEvalStorageStats(storage *leaderboard.Storage, verbose bool) {
	if storage == nil {
		return
	}

	stats, err := storage.RecalculateStats()
	if err != nil {
		if verbose {
			fprintf(os.Stderr, "Warning: failed to recalculate storage stats: %v\n", err)
		}
		return
	}

	if verbose {
		printf("\nStorage statistics updated:\n")
		printf("  Total decks evaluated: %d\n", stats.TotalDecksEvaluated)
		printf("  Unique decks: %d\n", stats.TotalUniqueDecks)
		printf("  Top score: %.2f\n", stats.TopScore)
		printf("  Average score: %.2f\n", stats.AvgScore)
	}
}

// processEvalBatchResults sorts, filters, and limits evaluation results
func processEvalBatchResults(
	results []evalBatchResult,
	sortBy, archetypeFilter string,
	topOnly bool,
	topN int,
	verbose bool,
) []evalBatchResult {
	sortEvaluationResults(results, sortBy)
	results = filterEvalResultsByArchetype(results, archetypeFilter, verbose)
	results = applyEvalTopNFilter(results, topOnly, topN)
	return results
}

// filterEvalResultsByArchetype filters results by archetype if filter is specified
func filterEvalResultsByArchetype(
	results []evalBatchResult,
	archetypeFilter string,
	verbose bool,
) []evalBatchResult {
	if archetypeFilter == "" {
		return results
	}

	filtered := make([]evalBatchResult, 0)
	for _, r := range results {
		if strings.EqualFold(string(r.Result.DetectedArchetype), archetypeFilter) {
			filtered = append(filtered, r)
		}
	}

	if len(filtered) == 0 && verbose {
		printf("No decks found matching archetype: %s\n", archetypeFilter)
	}

	return filtered
}

// applyEvalTopNFilter limits results to top N if requested
func applyEvalTopNFilter(results []evalBatchResult, topOnly bool, topN int) []evalBatchResult {
	if !topOnly || len(results) <= topN {
		return results
	}
	return results[:topN]
}

// formatEvalBatchResults formats evaluation results according to the specified format
func formatEvalBatchResults(
	results []evalBatchResult,
	format, sortBy, playerName, playerTag string,
	totalDecks int,
	totalTime time.Duration,
) (string, error) {
	switch strings.ToLower(format) {
	case "summary", compareFormatHuman:
		return formatEvaluationBatchSummary(results, totalDecks, totalTime, sortBy, playerName, playerTag), nil
	case batchFormatJSON:
		return formatEvalBatchResultsAsJSON(results, playerName, playerTag, totalDecks, sortBy, totalTime)
	case batchFormatCSV:
		return formatEvaluationBatchCSV(results), nil
	case batchFormatDetailed:
		return formatEvaluationBatchDetailed(results, playerName, playerTag), nil
	default:
		return "", fmt.Errorf("unknown format: %s (supported: summary, json, csv, detailed)", format)
	}
}

// formatEvalBatchResultsAsJSON formats evaluation results as JSON
func formatEvalBatchResultsAsJSON(
	results []evalBatchResult,
	playerName, playerTag string,
	totalDecks int,
	sortBy string,
	totalTime time.Duration,
) (string, error) {
	jsonData := map[string]any{
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"player": map[string]string{
			"name": playerName,
			"tag":  playerTag,
		},
		"evaluation_info": map[string]any{
			"total_decks":     totalDecks,
			"evaluated":       len(results),
			"sort_by":         sortBy,
			"evaluation_time": totalTime.String(),
		},
		"results": results,
	}

	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// writeEvalBatchOutput writes evaluation output to file or stdout
func writeEvalBatchOutput(output, outputDir, format, playerTag string, saveAggregated bool) error {
	if outputDir == "" || !saveAggregated {
		fmt.Print(output)
		return nil
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := buildEvalOutputFilename(timestamp, format, playerTag)
	outputPath := filepath.Join(outputDir, filename)

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(output), 0o644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	printf("\nEvaluation results saved to: %s\n", outputPath)
	return nil
}

// buildEvalOutputFilename constructs the output filename based on format
func buildEvalOutputFilename(timestamp, format, playerTag string) string {
	extension := "txt"
	switch format {
	case batchFormatJSON:
		extension = batchFormatJSON
	case batchFormatCSV:
		extension = batchFormatCSV
	}
	return fmt.Sprintf("%s_deck_evaluations_%s.%s", timestamp, playerTag, extension)
}

// evalBatchSetup holds the initialized resources for batch evaluation
type evalBatchSetup struct {
	Decks         []evalDeckInfo
	PlayerName    string
	PlayerTag     string
	PlayerContext *evaluation.PlayerContext
	Storage       *leaderboard.Storage
	SynergyDB     *deck.SynergyDatabase
}

// setupEvalBatch prepares all resources needed for batch evaluation
func setupEvalBatch(cmd *cli.Command, flags *evalBatchFlags) (*evalBatchSetup, error) {
	decks, playerName, loadedTag, err := loadEvalDecks(flags.FromSuite, flags.DeckDir, flags.Verbose)
	if err != nil {
		return nil, err
	}
	if len(decks) == 0 {
		return nil, fmt.Errorf("no decks found to evaluate")
	}

	playerTag := flags.PlayerTag
	if playerTag == "" {
		playerTag = loadedTag
	}

	apiToken := cmd.String("api-token")
	if apiToken == "" {
		apiToken = os.Getenv("CLASH_ROYALE_API_TOKEN")
	}

	playerContext, loadedName, err := loadEvalPlayerContext(playerTag, apiToken, flags.Verbose)
	if err != nil {
		return nil, err
	}
	if loadedName != "" {
		playerName = loadedName
	}

	storage, err := initEvalStorage(playerTag, flags.Verbose)
	if err != nil && flags.Verbose {
		fprintf(os.Stderr, "Warning: failed to initialize storage: %v\n", err)
	}

	if flags.Verbose {
		printf("Evaluating %d decks...\n", len(decks))
	}

	return &evalBatchSetup{
		Decks:         decks,
		PlayerName:    playerName,
		PlayerTag:     playerTag,
		PlayerContext: playerContext,
		Storage:       storage,
		SynergyDB:     deck.NewSynergyDatabase(),
	}, nil
}

// cleanupEvalBatch closes storage and prints stats
func cleanupEvalBatch(storage *leaderboard.Storage, verbose bool) {
	if storage != nil {
		updateEvalStorageStats(storage, verbose)
		if err := storage.Close(); err != nil {
			fprintf(os.Stderr, "Warning: failed to close storage: %v\n", err)
		}
	}
}

func deckEvaluateBatchCommand(ctx context.Context, cmd *cli.Command) error {
	flags, err := parseEvalBatchFlags(cmd)
	if err != nil {
		return err
	}

	setup, err := setupEvalBatch(cmd, flags)
	if err != nil {
		return err
	}
	defer cleanupEvalBatch(setup.Storage, flags.Verbose)

	results, totalTime, err := runEvalDecksBatch(setup.Decks, setup.SynergyDB, setup.PlayerContext, setup.Storage, setup.PlayerTag, flags.Verbose)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		return fmt.Errorf("no decks were successfully evaluated")
	}

	printEvalTimingSummary(results, totalTime, flags.ShowTiming, flags.Verbose)

	processed := processEvalBatchResults(results, flags.SortBy, flags.ArchetypeFilter, flags.TopOnly, flags.TopN, flags.Verbose)
	if len(processed) == 0 && flags.ArchetypeFilter != "" {
		return nil
	}

	output, err := formatEvalBatchResults(processed, flags.Format, flags.SortBy, setup.PlayerName, setup.PlayerTag, len(setup.Decks), totalTime)
	if err != nil {
		return err
	}

	return writeEvalBatchOutput(output, flags.OutputDir, flags.Format, setup.PlayerTag, flags.SaveAggregated)
}

// runPhase0CardConstraints runs the optional card constraint suggestion phase
//
//nolint:unused // Reserved for phased suite refactor tracked in beads.
