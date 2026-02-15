package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/urfave/cli/v3"
)

const (
	compareFormatTable    = "table"
	compareFormatHuman    = "human"
	compareFormatJSON     = "json"
	compareFormatCSV      = "csv"
	compareFormatMarkdown = "markdown"
	compareFormatMD       = "md"
)

// addCompareCommands adds deck comparison subcommands to the CLI
func addCompareCommands() *cli.Command {
	return &cli.Command{
		Name:  "compare",
		Usage: "Compare multiple decks side-by-side with detailed analysis",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "decks",
				Aliases:  []string{"d"},
				Usage:    "Decks to compare (format: \"card1-card2-...-card8\", can specify multiple)",
				Required: false,
			},
			&cli.StringSliceFlag{
				Name:  "names",
				Usage: "Custom names for each deck (optional, defaults to Deck #1, Deck #2, etc.)",
			},
			&cli.StringSliceFlag{
				Name:    "from-evaluations",
				Aliases: []string{"e"},
				Usage:   "Load decks from evaluation batch results (JSON files from 'deck evaluate-batch')",
			},
			&cli.IntFlag{
				Name:  "auto-select-top",
				Usage: "Automatically select and compare top N decks by score from evaluation results",
				Value: 0,
			},
			&cli.StringFlag{
				Name:  "format",
				Value: "table",
				Usage: "Output format: table, json, csv, markdown",
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Output file path (optional, prints to stdout if not specified)",
			},
			&cli.StringFlag{
				Name:  "report-output",
				Usage: "Generate comprehensive markdown report to file",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show detailed comparison including all analysis sections",
			},
			&cli.BoolFlag{
				Name:  "winrate",
				Usage: "Show predicted win rate comparison (requires meta data)",
			},
		},
		Action: deckCompareCommand,
	}
}

// batchEvalResult matches the structure from deck evaluate-batch
type batchEvalResult struct {
	Name      string                      `json:"name"`
	Strategy  string                      `json:"strategy"`
	Deck      []string                    `json:"deck"`
	Result    evaluation.EvaluationResult `json:"Result"`
	FilePath  string                      `json:"FilePath"`
	Evaluated string                      `json:"Evaluated"`
	Duration  int64                       `json:"Duration"`
}

// deckCompareCommand compares multiple decks side-by-side
func deckCompareCommand(ctx context.Context, cmd *cli.Command) error {
	deckStrings := cmd.StringSlice("decks")
	names := cmd.StringSlice("names")
	evalFiles := cmd.StringSlice("from-evaluations")
	autoSelectTop := cmd.Int("auto-select-top")
	format := cmd.String("format")
	outputFile := cmd.String("output")
	reportOutput := cmd.String("report-output")
	verbose := cmd.Bool("verbose")
	showWinRate := cmd.Bool("winrate")

	if err := validateComparisonInputs(deckStrings, evalFiles); err != nil {
		return err
	}

	deckNames, results, err := loadDecksForComparison(evalFiles, deckStrings, names, autoSelectTop)
	if err != nil {
		return err
	}

	if err := validateLoadedDecks(results); err != nil {
		return err
	}

	formattedOutput, err := formatDeckComparisonOutput(format, deckNames, results, verbose, showWinRate)
	if err != nil {
		return err
	}

	if err := writeComparisonOutput(outputFile, formattedOutput); err != nil {
		return err
	}

	if reportOutput != "" {
		report := generateComparisonReport(deckNames, results)
		if err := os.WriteFile(reportOutput, []byte(report), 0o644); err != nil {
			return fmt.Errorf("failed to write report: %w", err)
		}
		printf("Comprehensive report saved to: %s\n", reportOutput)
	}

	return nil
}

// validateComparisonInputs validates input sources for comparison command
func validateComparisonInputs(deckStrings, evalFiles []string) error {
	if len(deckStrings) == 0 && len(evalFiles) == 0 {
		return fmt.Errorf("must provide either --decks or --from-evaluations")
	}

	if len(deckStrings) > 0 && len(evalFiles) > 0 {
		return fmt.Errorf("cannot use both --decks and --from-evaluations")
	}

	return nil
}

// loadDecksForComparison loads decks from either evaluation files or CLI strings
func loadDecksForComparison(evalFiles, deckStrings, names []string, autoSelectTop int) ([]string, []evaluation.EvaluationResult, error) {
	if len(evalFiles) > 0 {
		deckNames, results, err := loadDecksFromEvaluations(evalFiles, autoSelectTop)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load evaluations: %w", err)
		}
		return deckNames, results, nil
	}

	// Parse decks from command line
	if len(deckStrings) < 2 {
		return nil, nil, fmt.Errorf("at least 2 decks are required for comparison, got %d", len(deckStrings))
	}

	if len(deckStrings) > 5 {
		return nil, nil, fmt.Errorf("maximum 5 decks can be compared at once, got %d", len(deckStrings))
	}

	deckCards := make([][]deck.CardCandidate, len(deckStrings))
	deckNames := make([]string, len(deckStrings))

	for i, deckStr := range deckStrings {
		cardNames := parseDeckString(deckStr)
		if len(cardNames) != 8 {
			return nil, nil, fmt.Errorf("deck #%d must contain exactly 8 cards, got %d", i+1, len(cardNames))
		}
		deckCards[i] = convertToCardCandidates(cardNames)

		if i < len(names) && names[i] != "" {
			deckNames[i] = names[i]
		} else {
			deckNames[i] = fmt.Sprintf("Deck #%d", i+1)
		}
	}

	synergyDB := deck.NewSynergyDatabase()
	results := make([]evaluation.EvaluationResult, len(deckCards))
	for i, cards := range deckCards {
		results[i] = evaluation.Evaluate(cards, synergyDB, nil)
	}

	return deckNames, results, nil
}

// validateLoadedDecks validates the loaded deck count
func validateLoadedDecks(results []evaluation.EvaluationResult) error {
	if len(results) < 2 {
		return fmt.Errorf("at least 2 decks are required for comparison, got %d", len(results))
	}

	if len(results) > 5 {
		return fmt.Errorf("maximum 5 decks can be compared at once, got %d (use --auto-select-top to limit)", len(results))
	}

	return nil
}

// formatComparisonOutput formats comparison output based on format type
func formatDeckComparisonOutput(format string, deckNames []string, results []evaluation.EvaluationResult, verbose, showWinRate bool) (string, error) {
	switch strings.ToLower(format) {
	case compareFormatTable, compareFormatHuman:
		return formatComparisonTable(deckNames, results, verbose, showWinRate), nil
	case compareFormatJSON:
		output, err := formatComparisonJSON(deckNames, results)
		if err != nil {
			return "", fmt.Errorf("failed to format JSON: %w", err)
		}
		return output, nil
	case compareFormatCSV:
		return formatComparisonCSV(deckNames, results), nil
	case compareFormatMarkdown, compareFormatMD:
		return formatComparisonMarkdown(deckNames, results, verbose), nil
	default:
		return "", fmt.Errorf("unknown format: %s (supported: table, json, csv, markdown)", format)
	}
}

// writeComparisonOutput writes output to file or stdout
func writeComparisonOutput(outputFile, content string) error {
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		printf("Comparison saved to: %s\n", outputFile)
	} else {
		fmt.Print(content)
	}
	return nil
}

// formatComparisonTable formats deck comparison as a table
func formatComparisonTable(names []string, results []evaluation.EvaluationResult, verbose, showWinRate bool) string {
	var sb strings.Builder

	sb.WriteString("╔════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                       DECK COMPARISON                                ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════════════╝\n\n")

	formatTableOverviewSection(&sb, names, results)
	formatTableCategoryScoresSection(&sb, names, results)
	formatTableBestInCategorySection(&sb, names, results)
	formatTableDeckCompositionSection(&sb, names, results)

	if verbose {
		formatTableVerboseAnalysisSection(&sb, names, results)
	}

	return sb.String()
}

// formatTableOverviewSection formats the overview section of comparison table
func formatTableOverviewSection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("📊 OVERVIEW\n")
	sb.WriteString("════════════\n\n")
	sb.WriteString(fmt.Sprintf("%-15s", "Deck"))
	for _, name := range names {
		sb.WriteString(fmt.Sprintf(" | %-20s", truncate(name, 20)))
	}
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", 15+23*len(names)) + "\n")

	// Overall scores
	sb.WriteString(fmt.Sprintf("%-15s", "Overall Score"))
	for _, r := range results {
		stars := calculateStars(r.OverallScore)
		rating := formatStarsDisplay(stars)
		sb.WriteString(fmt.Sprintf(" | %.2f %s %-13s", r.OverallScore, rating, r.OverallRating))
	}
	sb.WriteString("\n")

	// Average elixir
	sb.WriteString(fmt.Sprintf("%-15s", "Avg Elixir"))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf(" | %-20.2f", r.AvgElixir))
	}
	sb.WriteString("\n")

	// Archetype
	sb.WriteString(fmt.Sprintf("%-15s", "Archetype"))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf(" | %-20s", r.DetectedArchetype))
	}
	sb.WriteString("\n\n")
}

// formatTableCategoryScoresSection formats category scores section
func formatTableCategoryScoresSection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("📈 CATEGORY SCORES\n")
	sb.WriteString("══════════════════\n\n")
	sb.WriteString(fmt.Sprintf("%-15s", "Category"))
	for _, name := range names {
		sb.WriteString(fmt.Sprintf(" | %-20s", truncate(name, 20)))
	}
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", 15+23*len(names)) + "\n")

	categories := getEvaluationCategories()

	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf("%-15s", cat.name))
		for _, r := range results {
			score := cat.get(r)
			rating := formatStarsDisplay(score.Stars)
			sb.WriteString(fmt.Sprintf(" | %.1f %s %-14s", score.Score, rating, score.Rating))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

// formatTableBestInCategorySection formats best in category section
func formatTableBestInCategorySection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("🏆 BEST IN CATEGORY\n")
	sb.WriteString("══════════════════\n\n")

	categories := []struct {
		name string
		get  func(evaluation.EvaluationResult) evaluation.CategoryScore
	}{
		{"Overall", func(r evaluation.EvaluationResult) evaluation.CategoryScore {
			return evaluation.CategoryScore{Score: r.OverallScore, Rating: r.OverallRating}
		}},
		{"Attack", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.Attack }},
		{"Defense", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.Defense }},
		{"Synergy", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.Synergy }},
		{"Versatility", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.Versatility }},
	}

	for _, cat := range categories {
		bestIdx := findBestDeckIndex(results, cat.get)
		bestScore := cat.get(results[bestIdx]).Score
		sb.WriteString(fmt.Sprintf("%-15s: %s (%.2f)\n", cat.name, names[bestIdx], bestScore))
	}
	sb.WriteString("\n")
}

// formatTableDeckCompositionSection formats deck composition section
func formatTableDeckCompositionSection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("🃏 DECK COMPOSITION\n")
	sb.WriteString("══════════════════\n\n")
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("%s:\n", names[i]))
		for j, card := range r.Deck {
			if j > 0 && j%4 == 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("  %-18s", card))
		}
		sb.WriteString("\n\n")
	}
}

// formatTableVerboseAnalysisSection formats verbose analysis section
func formatTableVerboseAnalysisSection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("📋 DETAILED ANALYSIS\n")
	sb.WriteString("════════════════════\n\n")

	for i, r := range results {
		fmt.Fprintf(sb, "═══ %s ═══\n\n", names[i])

		fmt.Fprintf(sb, "Defense (%.1f/10.0): %s\n", r.DefenseAnalysis.Score, r.DefenseAnalysis.Rating)
		for _, detail := range r.DefenseAnalysis.Details {
			fmt.Fprintf(sb, "  • %s\n", detail)
		}
		sb.WriteString("\n")

		fmt.Fprintf(sb, "Attack (%.1f/10.0): %s\n", r.AttackAnalysis.Score, r.AttackAnalysis.Rating)
		for _, detail := range r.AttackAnalysis.Details {
			fmt.Fprintf(sb, "  • %s\n", detail)
		}
		sb.WriteString("\n")

		if r.SynergyMatrix.PairCount > 0 {
			fmt.Fprintf(sb, "Synergy: %d pairs found (%.1f%% coverage)\n",
				r.SynergyMatrix.PairCount, r.SynergyMatrix.SynergyCoverage)
			sb.WriteString("\n")
		}
	}
}

// findBestDeckIndex finds the index of the best deck using a score getter function
func findBestDeckIndex(results []evaluation.EvaluationResult, getScore func(evaluation.EvaluationResult) evaluation.CategoryScore) int {
	bestIdx := 0
	bestScore := -1.0
	for i, r := range results {
		score := getScore(r).Score
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	return bestIdx
}

// formatComparisonJSON formats deck comparison as JSON
func formatComparisonJSON(names []string, results []evaluation.EvaluationResult) (string, error) {
	comparison := map[string]any{
		"decks":   names,
		"results": results,
	}

	data, err := json.MarshalIndent(comparison, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data) + "\n", nil
}

// formatComparisonCSV formats deck comparison as CSV
func formatComparisonCSV(names []string, results []evaluation.EvaluationResult) string {
	var sb strings.Builder

	// Header
	sb.WriteString("Deck")
	for _, name := range names {
		sb.WriteString(fmt.Sprintf(",%s", name))
	}
	sb.WriteString("\n")

	// Overall score
	sb.WriteString("Overall Score")
	for _, r := range results {
		sb.WriteString(fmt.Sprintf(",%.2f", r.OverallScore))
	}
	sb.WriteString("\n")

	// Average elixir
	sb.WriteString("Avg Elixir")
	for _, r := range results {
		sb.WriteString(fmt.Sprintf(",%.2f", r.AvgElixir))
	}
	sb.WriteString("\n")

	// Category scores
	categories := []string{"Attack", "Defense", "Synergy", "Versatility", "F2P Friendly"}
	for _, cat := range categories {
		sb.WriteString(cat)
		for _, r := range results {
			var score float64
			switch cat {
			case "Attack":
				score = r.Attack.Score
			case "Defense":
				score = r.Defense.Score
			case "Synergy":
				score = r.Synergy.Score
			case "Versatility":
				score = r.Versatility.Score
			case "F2P Friendly":
				score = r.F2PFriendly.Score
			}
			sb.WriteString(fmt.Sprintf(",%.1f", score))
		}
		sb.WriteString("\n")
	}

	// Archetypes
	sb.WriteString("Archetype")
	for _, r := range results {
		sb.WriteString(fmt.Sprintf(",%s", r.DetectedArchetype))
	}
	sb.WriteString("\n")

	return sb.String()
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatStarsDisplay converts star count (0-3) to visual star representation
func formatStarsDisplay(count int) string {
	filled := "★"
	empty := "☆"

	stars := strings.Repeat(filled, count)
	stars += strings.Repeat(empty, 3-count)

	return stars
}

// loadDecksFromEvaluations loads decks from evaluation batch result JSON files
func loadDecksFromEvaluations(evalFiles []string, autoSelectTop int) ([]string, []evaluation.EvaluationResult, error) {
	type batchResultFile struct {
		Version        string `json:"version"`
		Timestamp      string `json:"timestamp"`
		EvaluationInfo struct {
			TotalDecks int    `json:"total_decks"`
			Evaluated  int    `json:"evaluated"`
			SortBy     string `json:"sort_by"`
		} `json:"evaluation_info"`
		Results []batchEvalResult `json:"results"`
	}

	var allResults []batchEvalResult

	// Load and merge all evaluation files
	for _, evalFile := range evalFiles {
		// Support glob patterns
		matches, err := filepath.Glob(evalFile)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid pattern %s: %w", evalFile, err)
		}

		if len(matches) == 0 {
			// Try as literal file path
			matches = []string{evalFile}
		}

		for _, file := range matches {
			data, err := os.ReadFile(file)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read %s: %w", file, err)
			}

			var batchData batchResultFile
			if err := json.Unmarshal(data, &batchData); err != nil {
				return nil, nil, fmt.Errorf("failed to parse %s: %w", file, err)
			}

			allResults = append(allResults, batchData.Results...)
		}
	}

	if len(allResults) == 0 {
		return nil, nil, fmt.Errorf("no decks found in evaluation files")
	}

	// Sort by overall score (descending)
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Result.OverallScore > allResults[j].Result.OverallScore
	})

	// Apply top selection if requested
	if autoSelectTop > 0 && len(allResults) > autoSelectTop {
		allResults = allResults[:autoSelectTop]
	}

	// Extract names and results
	names := make([]string, len(allResults))
	results := make([]evaluation.EvaluationResult, len(allResults))

	for i, r := range allResults {
		names[i] = r.Name
		results[i] = r.Result
	}

	return names, results, nil
}

// formatComparisonMarkdown formats deck comparison as markdown
func formatComparisonMarkdown(names []string, results []evaluation.EvaluationResult, verbose bool) string {
	var sb strings.Builder

	sb.WriteString("# Deck Comparison\n\n")
	sb.WriteString(fmt.Sprintf("*Comparing %d decks*\n\n", len(names)))

	formatMarkdownOverviewSection(&sb, names, results)
	formatMarkdownCategoryScoresSection(&sb, names, results)
	formatMarkdownBestInCategorySection(&sb, names, results)
	formatMarkdownDeckCompositionsSection(&sb, names, results)

	if verbose {
		formatMarkdownVerboseAnalysisSection(&sb, names, results)
	}

	return sb.String()
}

// formatMarkdownOverviewSection formats overview table in markdown
func formatMarkdownOverviewSection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("## Overview\n\n")
	sb.WriteString("| Deck | Overall Score | Rating | Avg Elixir | Archetype |\n")
	sb.WriteString("|------|--------------|--------|------------|------------|\n")

	for i, name := range names {
		r := results[i]
		stars := formatStarsDisplay(calculateStars(r.OverallScore))
		sb.WriteString(fmt.Sprintf("| %s | %.2f %s | %s | %.2f | %s |\n",
			name, r.OverallScore, stars, r.OverallRating, r.AvgElixir, r.DetectedArchetype))
	}
	sb.WriteString("\n")
}

// formatMarkdownCategoryScoresSection formats category scores in markdown
func formatMarkdownCategoryScoresSection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("## Category Scores\n\n")
	sb.WriteString("| Category | ")
	for _, name := range names {
		sb.WriteString(fmt.Sprintf("%s | ", name))
	}
	sb.WriteString("\n|----------|")
	for range names {
		sb.WriteString("---------|")
	}
	sb.WriteString("\n")

	categories := getEvaluationCategories()

	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf("| **%s** | ", cat.name))
		for _, r := range results {
			score := cat.get(r)
			stars := formatStarsDisplay(score.Stars)
			sb.WriteString(fmt.Sprintf("%.1f %s | ", score.Score, stars))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

// formatMarkdownBestInCategorySection formats best in category in markdown
func formatMarkdownBestInCategorySection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("## 🏆 Best in Category\n\n")
	categories := getEvaluationCategories()

	for _, cat := range categories {
		bestIdx := findBestDeckIndex(results, cat.get)
		bestScore := cat.get(results[bestIdx]).Score
		sb.WriteString(fmt.Sprintf("- **%s**: %s (%.2f)\n", cat.name, names[bestIdx], bestScore))
	}
	sb.WriteString("\n")
}

// formatMarkdownDeckCompositionsSection formats deck compositions in markdown
func formatMarkdownDeckCompositionsSection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("## Deck Compositions\n\n")
	for i, name := range names {
		sb.WriteString(fmt.Sprintf("### %s\n\n", name))
		sb.WriteString("```\n")
		for j, card := range results[i].Deck {
			sb.WriteString(fmt.Sprintf("%-18s", card))
			if (j+1)%4 == 0 {
				sb.WriteString("\n")
			}
		}
		sb.WriteString("```\n\n")
	}
}

// formatMarkdownVerboseAnalysisSection formats verbose analysis in markdown
func formatMarkdownVerboseAnalysisSection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("## Detailed Analysis\n\n")
	for i, name := range names {
		r := results[i]
		fmt.Fprintf(sb, "### %s\n\n", name)

		fmt.Fprintf(sb, "**Defense** (%.1f/10.0): %s\n\n", r.DefenseAnalysis.Score, r.DefenseAnalysis.Rating)
		for _, detail := range r.DefenseAnalysis.Details {
			fmt.Fprintf(sb, "- %s\n", detail)
		}
		sb.WriteString("\n")

		fmt.Fprintf(sb, "**Attack** (%.1f/10.0): %s\n\n", r.AttackAnalysis.Score, r.AttackAnalysis.Rating)
		for _, detail := range r.AttackAnalysis.Details {
			fmt.Fprintf(sb, "- %s\n", detail)
		}
		sb.WriteString("\n")

		if r.SynergyMatrix.PairCount > 0 {
			fmt.Fprintf(sb, "**Synergy**: %d pairs found (%.1f%% coverage)\n\n",
				r.SynergyMatrix.PairCount, r.SynergyMatrix.SynergyCoverage)
		}
	}
}

// getEvaluationCategories returns the standard evaluation categories
func getEvaluationCategories() []struct {
	name string
	get  func(evaluation.EvaluationResult) evaluation.CategoryScore
} {
	return []struct {
		name string
		get  func(evaluation.EvaluationResult) evaluation.CategoryScore
	}{
		{"Attack", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.Attack }},
		{"Defense", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.Defense }},
		{"Synergy", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.Synergy }},
		{"Versatility", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.Versatility }},
		{"F2P Friendly", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.F2PFriendly }},
	}
}

// generateComparisonReport generates a comprehensive markdown report
func generateComparisonReport(names []string, results []evaluation.EvaluationResult) string {
	var sb strings.Builder

	// Title and metadata
	sb.WriteString("# Deck Comparison Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated**: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Decks Compared**: %d\n\n", len(names)))
	sb.WriteString("---\n\n")

	bestIdx := findBestOverallDeck(results)
	formatReportExecutiveSummary(&sb, names, results, bestIdx)
	formatReportDetailedScoreComparison(&sb, names, results)
	formatReportCategoryChampions(&sb, names, results)
	formatReportDeckDetails(&sb, names, results)
	formatReportRecommendations(&sb, names, results, bestIdx)

	return sb.String()
}

// findBestOverallDeck finds the index of the best overall deck
func findBestOverallDeck(results []evaluation.EvaluationResult) int {
	bestIdx := 0
	bestScore := -1.0
	for i, r := range results {
		if r.OverallScore > bestScore {
			bestScore = r.OverallScore
			bestIdx = i
		}
	}
	return bestIdx
}

// formatReportExecutiveSummary formats the executive summary section
func formatReportExecutiveSummary(sb *strings.Builder, names []string, results []evaluation.EvaluationResult, bestIdx int) {
	sb.WriteString("## Executive Summary\n\n")

	sb.WriteString(fmt.Sprintf("### 🏆 Recommended Deck: **%s**\n\n", names[bestIdx]))
	sb.WriteString(fmt.Sprintf("- **Overall Score**: %.2f/10.0 (%s)\n",
		results[bestIdx].OverallScore, results[bestIdx].OverallRating))
	sb.WriteString(fmt.Sprintf("- **Archetype**: %s\n", results[bestIdx].DetectedArchetype))
	sb.WriteString(fmt.Sprintf("- **Average Elixir**: %.2f\n\n", results[bestIdx].AvgElixir))

	formatReportRankings(sb, names, results)
	sb.WriteString("\n---\n\n")
}

// formatReportRankings formats the rankings with medal emojis
func formatReportRankings(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("### Overall Rankings\n\n")
	indices := make([]int, len(names))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool {
		return results[indices[i]].OverallScore > results[indices[j]].OverallScore
	})

	for rank, idx := range indices {
		emoji := getRankingEmoji(rank + 1)
		sb.WriteString(fmt.Sprintf("%s **%s** - %.2f/10.0 (%s)\n",
			emoji, names[idx], results[idx].OverallScore, results[idx].OverallRating))
	}
}

// getRankingEmoji returns the appropriate emoji for a rank
func getRankingEmoji(rank int) string {
	switch rank {
	case 1:
		return "🥇"
	case 2:
		return "🥈"
	case 3:
		return "🥉"
	default:
		return fmt.Sprintf("%d.", rank)
	}
}

// formatReportDetailedScoreComparison formats detailed score comparison table
func formatReportDetailedScoreComparison(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("## Detailed Score Comparison\n\n")
	sb.WriteString("| Metric | ")
	for _, name := range names {
		sb.WriteString(fmt.Sprintf("%s | ", truncate(name, 18)))
	}
	sb.WriteString("\n|--------|")
	for range names {
		sb.WriteString("--------------------|")
	}
	sb.WriteString("\n")

	// Overall
	sb.WriteString("| **Overall** | ")
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("%.2f %s | ", r.OverallScore, formatStarsDisplay(calculateStars(r.OverallScore))))
	}
	sb.WriteString("\n")

	categories := getEvaluationCategories()

	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf("| %s | ", cat.name))
		for _, r := range results {
			score := cat.get(r)
			sb.WriteString(fmt.Sprintf("%.1f %s | ", score.Score, formatStarsDisplay(score.Stars)))
		}
		sb.WriteString("\n")
	}

	// Elixir and archetype
	sb.WriteString("| Avg Elixir | ")
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("%.2f | ", r.AvgElixir))
	}
	sb.WriteString("\n")

	sb.WriteString("| Archetype | ")
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("%s | ", r.DetectedArchetype))
	}
	sb.WriteString("\n\n---\n\n")
}

// formatReportCategoryChampions formats category champions section
func formatReportCategoryChampions(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("## Category Champions\n\n")
	categories := getEvaluationCategories()

	for _, cat := range categories {
		bestIdx := findBestDeckIndex(results, cat.get)
		sb.WriteString(fmt.Sprintf("### 🏆 Best %s: **%s**\n\n", cat.name, names[bestIdx]))
		sb.WriteString(fmt.Sprintf("- **Score**: %.1f/10.0 (%s)\n",
			cat.get(results[bestIdx]).Score, cat.get(results[bestIdx]).Rating))
		sb.WriteString(fmt.Sprintf("- **Assessment**: %s\n\n", cat.get(results[bestIdx]).Assessment))
	}

	sb.WriteString("---\n\n")
}

// formatReportDeckDetails formats detailed deck information
func formatReportDeckDetails(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("## Deck Details\n\n")
	for i, name := range names {
		r := results[i]
		sb.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, name))

		// Deck composition
		sb.WriteString("**Cards**:\n```\n")
		for j, card := range r.Deck {
			sb.WriteString(fmt.Sprintf("%-20s", card))
			if (j+1)%4 == 0 {
				sb.WriteString("\n")
			}
		}
		sb.WriteString("```\n\n")

		// Key stats
		sb.WriteString("**Key Statistics**:\n")
		sb.WriteString(fmt.Sprintf("- Overall Score: %.2f/10.0 (%s)\n", r.OverallScore, r.OverallRating))
		sb.WriteString(fmt.Sprintf("- Archetype: %s (%.0f%% confidence)\n", r.DetectedArchetype, r.ArchetypeConfidence*100))
		sb.WriteString(fmt.Sprintf("- Average Elixir: %.2f\n\n", r.AvgElixir))

		formatDeckStrengthsAndWeaknesses(sb, r)
		formatDeckAnalysis(sb, r)

		sb.WriteString("---\n\n")
	}
}

// formatDeckStrengthsAndWeaknesses formats strengths and weaknesses for a deck
func formatDeckStrengthsAndWeaknesses(sb *strings.Builder, r evaluation.EvaluationResult) {
	sb.WriteString("**Strengths**:\n")
	strengths := []struct {
		name  string
		score evaluation.CategoryScore
	}{
		{"Attack", r.Attack},
		{"Defense", r.Defense},
		{"Synergy", r.Synergy},
		{"Versatility", r.Versatility},
		{"F2P Friendly", r.F2PFriendly},
	}

	// Sort by score descending
	sort.Slice(strengths, func(i, j int) bool {
		return strengths[i].score.Score > strengths[j].score.Score
	})

	for _, s := range strengths[:min(3, len(strengths))] {
		if s.score.Score >= 7.0 {
			fmt.Fprintf(sb, "- %s: %.1f/10.0 - %s\n", s.name, s.score.Score, s.score.Assessment)
		}
	}
	sb.WriteString("\n")

	sb.WriteString("**Areas for Improvement**:\n")
	// Reverse for weaknesses
	for i := len(strengths) - 1; i >= max(0, len(strengths)-3); i-- {
		s := strengths[i]
		if s.score.Score < 7.0 {
			fmt.Fprintf(sb, "- %s: %.1f/10.0 - %s\n", s.name, s.score.Score, s.score.Assessment)
		}
	}
	sb.WriteString("\n")
}

// formatDeckAnalysis formats defense, attack, and synergy analysis
func formatDeckAnalysis(sb *strings.Builder, r evaluation.EvaluationResult) {
	// Defense analysis
	if len(r.DefenseAnalysis.Details) > 0 {
		fmt.Fprintf(sb, "**Defense Analysis** (%.1f/10.0):\n", r.DefenseAnalysis.Score)
		for _, detail := range r.DefenseAnalysis.Details {
			fmt.Fprintf(sb, "- %s\n", detail)
		}
		sb.WriteString("\n")
	}

	// Attack analysis
	if len(r.AttackAnalysis.Details) > 0 {
		fmt.Fprintf(sb, "**Attack Analysis** (%.1f/10.0):\n", r.AttackAnalysis.Score)
		for _, detail := range r.AttackAnalysis.Details {
			fmt.Fprintf(sb, "- %s\n", detail)
		}
		sb.WriteString("\n")
	}

	// Synergy information
	if r.SynergyMatrix.PairCount > 0 {
		fmt.Fprintf(sb, "**Synergy**: %d card pairs found (%.1f%% coverage, avg synergy: %.2f)\n\n",
			r.SynergyMatrix.PairCount, r.SynergyMatrix.SynergyCoverage, r.SynergyMatrix.AverageSynergy)
	}
}

// formatReportRecommendations formats the recommendations section
func formatReportRecommendations(sb *strings.Builder, names []string, results []evaluation.EvaluationResult, bestIdx int) {
	sb.WriteString("## Recommendations\n\n")
	sb.WriteString(fmt.Sprintf("Based on the analysis, **%s** is the strongest deck overall with a score of %.2f/10.0.\n\n",
		names[bestIdx], results[bestIdx].OverallScore))

	sb.WriteString("### When to Use Each Deck\n\n")
	for i, name := range names {
		r := results[i]
		sb.WriteString(fmt.Sprintf("**%s** (%s archetype):\n", name, r.DetectedArchetype))

		// Recommend based on archetype and strengths
		if r.Attack.Score > r.Defense.Score {
			sb.WriteString("- Best for: Aggressive playstyle, ladder pushing\n")
		} else if r.Defense.Score > r.Attack.Score {
			sb.WriteString("- Best for: Defensive counter-attacks, conservative play\n")
		} else {
			sb.WriteString("- Best for: Balanced matchups, versatile gameplay\n")
		}

		if r.F2PFriendly.Score >= 8.0 {
			sb.WriteString("- Excellent for F2P players\n")
		}

		sb.WriteString("\n")
	}
}

// calculateStars converts overall score to star count (0-3)
func calculateStars(score float64) int {
	if score >= 9 {
		return 3
	} else if score >= 7 {
		return 2
	} else if score >= 5 {
		return 1
	}
	return 0
}
