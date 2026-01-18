package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/urfave/cli/v3"
)

const (
	batchFormatSummary  = "summary"
	batchFormatHuman    = "human"
	batchFormatJSON     = "json"
	batchFormatCSV      = "csv"
	batchFormatDetailed = "detailed"

	batchSortOverall     = "overall"
	batchSortAttack      = "attack"
	batchSortDefense     = "defense"
	batchSortSynergy     = "synergy"
	batchSortVersatility = "versatility"
	batchSortElixir      = "elixir"
)

// addBatchCommands adds batch evaluation subcommands to the CLI
func addBatchCommands() *cli.Command {
	return &cli.Command{
		Name:  "batch",
		Usage: "Evaluate multiple decks in batch with performance optimization",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "file",
				Usage: "Input file containing decks to evaluate (CSV or JSON)",
			},
			&cli.StringSliceFlag{
				Name:    "decks",
				Aliases: []string{"d"},
				Usage:   "Decks to evaluate directly (format: \"card1-card2-...-card8\", can specify multiple)",
			},
			&cli.StringFlag{
				Name:  "format",
				Value: batchFormatSummary,
				Usage: "Output format: summary, json, csv, detailed",
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Output file path (optional, prints to stdout if not specified)",
			},
			&cli.StringFlag{
				Name:  "sort-by",
				Value: batchSortOverall,
				Usage: "Sort results by: overall, attack, defense, synergy, versatility, elixir",
			},
			&cli.BoolFlag{
				Name:  "top-only",
				Usage: "Show only top 10 decks",
			},
			&cli.IntFlag{
				Name:  "top-n",
				Value: 10,
				Usage: "Number of top decks to show (with --top-only)",
			},
			&cli.BoolFlag{
				Name:  "filter-archetype",
				Usage: "Filter by specific archetype (use with --archetype)",
			},
			&cli.StringFlag{
				Name:  "archetype",
				Usage: "Archetype to filter by (e.g., beatdown, control, cycle)",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show detailed progress information",
			},
			&cli.BoolFlag{
				Name:  "timing",
				Usage: "Show timing information for each deck evaluation",
			},
		},
		Action: deckBatchCommand,
	}
}

// deckBatchCommand evaluates multiple decks in batch
func deckBatchCommand(ctx context.Context, cmd *cli.Command) error {
	inputFile := cmd.String("file")
	decksFlag := cmd.StringSlice("decks")
	format := cmd.String("format")
	outputFile := cmd.String("output")
	sortBy := cmd.String("sort-by")
	topOnly := cmd.Bool("top-only")
	topN := cmd.Int("top-n")
	filterArchetype := cmd.Bool("filter-archetype")
	archetypeFilter := cmd.String("archetype")
	verbose := cmd.Bool("verbose")
	showTiming := cmd.Bool("timing")

	// Validation: Must provide either --file or --decks
	if inputFile == "" && len(decksFlag) == 0 {
		return fmt.Errorf("must provide either --file or --decks")
	}

	if inputFile != "" && len(decksFlag) > 0 {
		return fmt.Errorf("cannot use both --file and --decks")
	}

	// Parse decks
	var deckStrings []string
	var deckNames []string

	if inputFile != "" {
		// Load decks from file
		loadedDecks, loadedNames, err := loadDecksFromFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to load decks from file: %w", err)
		}
		deckStrings = loadedDecks
		deckNames = loadedNames

		if verbose {
			printf("Loaded %d decks from %s\n", len(deckStrings), inputFile)
		}
	} else {
		// Use decks from command line
		deckStrings = decksFlag
		for i := range deckStrings {
			deckNames = append(deckNames, fmt.Sprintf("Deck #%d", i+1))
		}
	}

	if len(deckStrings) == 0 {
		return fmt.Errorf("no decks to evaluate")
	}

	if verbose {
		printf("Evaluating %d decks...\n", len(deckStrings))
	}

	// Create synergy database (shared for all evaluations)
	synergyDB := deck.NewSynergyDatabase()

	// Evaluate all decks with timing
	results := make([]BatchEvaluationResult, len(deckStrings))
	startTime := time.Now()

	for i, deckStr := range deckStrings {
		deckStart := time.Now()

		cardNames := parseDeckString(deckStr)
		if len(cardNames) != 8 {
			return fmt.Errorf("deck #%d (%s) must contain exactly 8 cards, got %d",
				i+1, deckNames[i], len(cardNames))
		}

		deckCards := convertToCardCandidates(cardNames)
		result := evaluation.Evaluate(deckCards, synergyDB, nil)

		elapsed := time.Since(deckStart)

		results[i] = BatchEvaluationResult{
			Name:      deckNames[i],
			Deck:      result.Deck,
			Result:    result,
			Evaluated: deckStart,
			Duration:  elapsed,
		}

		if verbose {
			printf("  [%d/%d] %s: %.2f (%s)\n",
				i+1, len(deckStrings), deckNames[i], result.OverallScore, result.OverallRating)
		}
	}

	totalTime := time.Since(startTime)

	if verbose || showTiming {
		printf("\nBatch evaluation completed in %v\n", totalTime)
		printf("Average time per deck: %v\n", totalTime/time.Duration(len(deckStrings)))
	}

	// Sort results
	sortResults(results, sortBy)

	// Filter by archetype if requested
	if filterArchetype && archetypeFilter != "" {
		results = filterByArchetype(results, archetypeFilter)
		if len(results) == 0 {
			printf("No decks found matching archetype: %s\n", archetypeFilter)
			return nil
		}
	}

	// Apply top filter
	if topOnly && len(results) > topN {
		results = results[:topN]
	}

	// Format output based on requested format
	var formattedOutput string
	var err error
	batchResult := &BatchResult{
		TotalDecks:    len(deckStrings),
		FilteredDecks: len(results),
		SortBy:        sortBy,
		TotalTime:     totalTime,
		AvgTime:       totalTime / time.Duration(len(deckStrings)),
		Results:       results,
	}

	switch strings.ToLower(format) {
	case batchFormatSummary, batchFormatHuman:
		formattedOutput = formatBatchSummary(batchResult, verbose)
	case batchFormatJSON:
		formattedOutput, err = formatBatchJSON(batchResult)
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
	case batchFormatCSV:
		formattedOutput = formatBatchCSV(batchResult)
	case batchFormatDetailed:
		formattedOutput = formatBatchDetailed(batchResult)
	default:
		return fmt.Errorf("unknown format: %s (supported: summary, json, csv, detailed)", format)
	}

	// Output to file or stdout
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(formattedOutput), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		printf("Batch evaluation saved to: %s\n", outputFile)
	} else {
		fmt.Print(formattedOutput)
	}

	return nil
}

// BatchEvaluationResult represents a single deck evaluation with metadata
type BatchEvaluationResult struct {
	Name      string
	Deck      []string
	Result    evaluation.EvaluationResult
	Evaluated time.Time
	Duration  time.Duration
}

// BatchResult represents the complete batch evaluation results
type BatchResult struct {
	TotalDecks    int
	FilteredDecks int
	SortBy        string
	TotalTime     time.Duration
	AvgTime       time.Duration
	Results       []BatchEvaluationResult
}

// loadDecksFromFile loads decks from a CSV or JSON file
func loadDecksFromFile(filePath string) ([]string, []string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".csv":
		return loadDecksFromCSV(filePath)
	case ".json":
		return loadDecksFromJSON(filePath)
	default:
		// Try to detect format by content
		decks, names, err := loadDecksFromJSON(filePath)
		if err != nil {
			decks, names, err = loadDecksFromCSV(filePath)
			if err != nil {
				return nil, nil, fmt.Errorf("unsupported file format: %s", ext)
			}
		}
		return decks, names, nil
	}
}

// loadDecksFromCSV loads decks from a CSV file
func loadDecksFromCSV(filePath string) ([]string, []string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer closeFile(file)

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	if len(records) == 0 {
		return nil, nil, fmt.Errorf("CSV file is empty")
	}

	decks := make([]string, 0, len(records))
	names := make([]string, 0, len(records))

	// Determine if first row is header
	startIdx := 0
	if len(records) > 0 {
		// Check if first row looks like a header (non-numeric content)
		firstDeck := strings.Join(records[0], ",")
		if !containsCardNames(firstDeck) {
			startIdx = 1 // Skip header
		}
	}

	for i := startIdx; i < len(records); i++ {
		record := records[i]
		if len(record) == 0 {
			continue
		}

		// First column is name (optional), rest are cards
		name := fmt.Sprintf("Deck #%d", len(decks)+1)
		var deckStr string

		if len(record) > 1 && record[0] != "" && !looksLikeCardName(record[0]) {
			name = record[0]
			deckStr = strings.Join(record[1:], ",")
		} else {
			deckStr = strings.Join(record, ",")
		}

		// Convert comma-separated to dash-separated
		deckStr = strings.ReplaceAll(deckStr, ",", "-")
		deckStr = strings.ReplaceAll(deckStr, " ", "")

		decks = append(decks, deckStr)
		names = append(names, name)
	}

	return decks, names, nil
}

// loadDecksFromJSON loads decks from a JSON file
func loadDecksFromJSON(filePath string) ([]string, []string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}

	var inputData struct {
		Decks []struct {
			Name  string   `json:"name"`
			Cards []string `json:"cards"`
		} `json:"decks"`
	}

	if err := json.Unmarshal(data, &inputData); err != nil {
		return nil, nil, err
	}

	if len(inputData.Decks) == 0 {
		return nil, nil, fmt.Errorf("JSON file contains no decks")
	}

	decks := make([]string, len(inputData.Decks))
	names := make([]string, len(inputData.Decks))

	for i, deck := range inputData.Decks {
		if len(deck.Cards) != 8 {
			return nil, nil, fmt.Errorf("deck #%d must contain exactly 8 cards, got %d", i+1, len(deck.Cards))
		}

		decks[i] = strings.Join(deck.Cards, "-")
		if deck.Name != "" {
			names[i] = deck.Name
		} else {
			names[i] = fmt.Sprintf("Deck #%d", i+1)
		}
	}

	return decks, names, nil
}

// containsCardNames checks if a string contains card names
func containsCardNames(s string) bool {
	cardKeywords := []string{"knight", "archer", "giant", "pekka", "musketeer", "hog", "miner"}
	lower := strings.ToLower(s)
	for _, keyword := range cardKeywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

// looksLikeCardName checks if a string looks like a card name
func looksLikeCardName(s string) bool {
	// Card names are typically 1-3 words, not numbers
	if _, err := parseCardCount(s); err == nil {
		return false // It's a number, not a name
	}
	return true
}

// parseCardCount tries to parse a string as a card count (for header detection)
func parseCardCount(s string) (int, error) {
	var count int
	_, err := fmt.Sscanf(s, "%d", &count)
	return count, err
}

// sortResults sorts batch results by the specified criteria
func sortResults(results []BatchEvaluationResult, sortBy string) {
	switch strings.ToLower(sortBy) {
	case batchSortOverall:
		// Already sorted by overall by default
	case batchSortAttack:
		// Sort by attack score (descending)
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[j].Result.Attack.Score > results[i].Result.Attack.Score {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	case batchSortDefense:
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[j].Result.Defense.Score > results[i].Result.Defense.Score {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	case batchSortSynergy:
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[j].Result.Synergy.Score > results[i].Result.Synergy.Score {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	case batchSortVersatility:
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[j].Result.Versatility.Score > results[i].Result.Versatility.Score {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	case batchSortElixir:
		// Sort by elixir (ascending)
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[j].Result.AvgElixir < results[i].Result.AvgElixir {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	}

	// Apply overall score as secondary sort for all categories
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Result.OverallScore > results[i].Result.OverallScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// filterByArchetype filters results by archetype
func filterByArchetype(results []BatchEvaluationResult, archetype string) []BatchEvaluationResult {
	filtered := make([]BatchEvaluationResult, 0)
	targetArchetype := evaluation.Archetype(strings.ToLower(archetype))

	for _, r := range results {
		if r.Result.DetectedArchetype == targetArchetype {
			filtered = append(filtered, r)
		}
	}

	return filtered
}

// formatBatchSummary formats batch results as a summary table
func formatBatchSummary(batch *BatchResult, verbose bool) string {
	var sb strings.Builder

	sb.WriteString("╔════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                       BATCH EVALUATION                              ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════════════╝\n\n")

	sb.WriteString(fmt.Sprintf("Total Decks Evaluated: %d\n", batch.TotalDecks))
	if batch.FilteredDecks != batch.TotalDecks {
		sb.WriteString(fmt.Sprintf("Decks After Filtering: %d\n", batch.FilteredDecks))
	}
	sb.WriteString(fmt.Sprintf("Sort By: %s\n", batch.SortBy))
	sb.WriteString(fmt.Sprintf("Total Time: %v\n", batch.TotalTime))
	sb.WriteString(fmt.Sprintf("Average Time: %v\n\n", batch.AvgTime))

	sb.WriteString("📊 TOP DECKS\n")
	sb.WriteString("════════════\n\n")

	sb.WriteString(fmt.Sprintf("%-4s %-25s %-10s %-10s %-10s %-10s %-8s %s\n",
		"Rank", "Deck Name", "Overall", "Attack", "Defense", "Synergy", "Elixir", "Archetype"))
	sb.WriteString(strings.Repeat("─", 100) + "\n")

	for i, r := range batch.Results {
		sb.WriteString(fmt.Sprintf("%-4d %-25s %-10.2f %-10.1f %-10.1f %-10.1f %-8.2f %s\n",
			i+1,
			truncate(r.Name, 25),
			r.Result.OverallScore,
			r.Result.Attack.Score,
			r.Result.Defense.Score,
			r.Result.Synergy.Score,
			r.Result.AvgElixir,
			r.Result.DetectedArchetype))
	}

	if verbose {
		sb.WriteString("\n⏱ TIMING INFORMATION\n")
		sb.WriteString("══════════════════\n\n")
		for _, r := range batch.Results {
			sb.WriteString(fmt.Sprintf("%-30s: %v\n", r.Name, r.Duration))
		}
	}

	return sb.String()
}

// formatBatchJSON formats batch results as JSON
func formatBatchJSON(batch *BatchResult) (string, error) {
	data, err := json.MarshalIndent(batch, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

// formatBatchCSV formats batch results as CSV
func formatBatchCSV(batch *BatchResult) string {
	var sb strings.Builder

	// Header
	sb.WriteString("Rank,Name,Overall,Attack,Defense,Synergy,Versatility,F2P,AvgElixir,Archetype\n")

	// Rows
	for i, r := range batch.Results {
		sb.WriteString(fmt.Sprintf("%d,\"%s\",%.2f,%.1f,%.1f,%.1f,%.1f,%.1f,%.2f,%s\n",
			i+1,
			r.Name,
			r.Result.OverallScore,
			r.Result.Attack.Score,
			r.Result.Defense.Score,
			r.Result.Synergy.Score,
			r.Result.Versatility.Score,
			r.Result.F2PFriendly.Score,
			r.Result.AvgElixir,
			r.Result.DetectedArchetype))
	}

	return sb.String()
}

// formatBatchDetailed formats batch results with full details
func formatBatchDetailed(batch *BatchResult) string {
	var sb strings.Builder

	sb.WriteString("╔════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                 DETAILED BATCH EVALUATION                          ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════════════╝\n\n")

	sb.WriteString(fmt.Sprintf("Total Decks Evaluated: %d\n", batch.TotalDecks))
	sb.WriteString(fmt.Sprintf("Total Time: %v\n\n", batch.TotalTime))

	for i, r := range batch.Results {
		sb.WriteString(fmt.Sprintf("═══ Deck #%d: %s ═══\n\n", i+1, r.Name))
		sb.WriteString(evaluation.FormatDetailed(&r.Result))
		sb.WriteString("\n")
	}

	return sb.String()
}
