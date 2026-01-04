package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/urfave/cli/v3"
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
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:  "names",
				Usage: "Custom names for each deck (optional, defaults to Deck #1, Deck #2, etc.)",
			},
			&cli.StringFlag{
				Name:  "format",
				Value: "table",
				Usage: "Output format: table, json, csv",
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Output file path (optional, prints to stdout if not specified)",
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

// deckCompareCommand compares multiple decks side-by-side
func deckCompareCommand(ctx context.Context, cmd *cli.Command) error {
	deckStrings := cmd.StringSlice("decks")
	names := cmd.StringSlice("names")
	format := cmd.String("format")
	outputFile := cmd.String("output")
	verbose := cmd.Bool("verbose")
	showWinRate := cmd.Bool("winrate")

	if len(deckStrings) < 2 {
		return fmt.Errorf("at least 2 decks are required for comparison, got %d", len(deckStrings))
	}

	if len(deckStrings) > 5 {
		return fmt.Errorf("maximum 5 decks can be compared at once, got %d", len(deckStrings))
	}

	// Parse all decks
	deckCards := make([][]deck.CardCandidate, len(deckStrings))
	deckNames := make([]string, len(deckStrings))

	for i, deckStr := range deckStrings {
		cardNames := parseDeckString(deckStr)
		if len(cardNames) != 8 {
			return fmt.Errorf("deck #%d must contain exactly 8 cards, got %d", i+1, len(cardNames))
		}
		deckCards[i] = convertToCardCandidates(cardNames)

		// Assign name
		if i < len(names) && names[i] != "" {
			deckNames[i] = names[i]
		} else {
			deckNames[i] = fmt.Sprintf("Deck #%d", i+1)
		}
	}

	// Create synergy database
	synergyDB := deck.NewSynergyDatabase()

	// Evaluate all decks
	results := make([]evaluation.EvaluationResult, len(deckCards))
	for i, cards := range deckCards {
		result := evaluation.Evaluate(cards, synergyDB)
		results[i] = result
	}

	// Format output based on requested format
	var formattedOutput string
	var err error
	switch strings.ToLower(format) {
	case "table", "human":
		formattedOutput = formatComparisonTable(deckNames, results, verbose, showWinRate)
	case "json":
		formattedOutput, err = formatComparisonJSON(deckNames, results)
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
	case "csv":
		formattedOutput = formatComparisonCSV(deckNames, results)
	default:
		return fmt.Errorf("unknown format: %s (supported: table, json, csv)", format)
	}

	// Output to file or stdout
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(formattedOutput), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Comparison saved to: %s\n", outputFile)
	} else {
		fmt.Print(formattedOutput)
	}

	return nil
}

// formatComparisonTable formats deck comparison as a table
func formatComparisonTable(names []string, results []evaluation.EvaluationResult, verbose, showWinRate bool) string {
	var sb strings.Builder

	sb.WriteString("╔════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                       DECK COMPARISON                                ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════════════╝\n\n")

	// Overview table
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
		// Calculate stars based on overall score (0-3 scale)
		stars := 0
		if r.OverallScore >= 9 {
			stars = 3
		} else if r.OverallScore >= 7 {
			stars = 2
		} else if r.OverallScore >= 5 {
			stars = 1
		}
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

	// Category scores comparison
	sb.WriteString("📈 CATEGORY SCORES\n")
	sb.WriteString("══════════════════\n\n")
	sb.WriteString(fmt.Sprintf("%-15s", "Category"))
	for _, name := range names {
		sb.WriteString(fmt.Sprintf(" | %-20s", truncate(name, 20)))
	}
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", 15+23*len(names)) + "\n")

	categories := []struct {
		name  string
		score evaluation.CategoryScore
	}{
		{"Attack", results[0].Attack},
		{"Defense", results[0].Defense},
		{"Synergy", results[0].Synergy},
		{"Versatility", results[0].Versatility},
		{"F2P Friendly", results[0].F2PFriendly},
	}

	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf("%-15s", cat.name))
		for _, r := range results {
			var score evaluation.CategoryScore
			switch cat.name {
			case "Attack":
				score = r.Attack
			case "Defense":
				score = r.Defense
			case "Synergy":
				score = r.Synergy
			case "Versatility":
				score = r.Versatility
			case "F2P Friendly":
				score = r.F2PFriendly
			}
			rating := formatStarsDisplay(score.Stars)
			sb.WriteString(fmt.Sprintf(" | %.1f %s %-14s", score.Score, rating, score.Rating))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Best deck per category
	sb.WriteString("🏆 BEST IN CATEGORY\n")
	sb.WriteString("══════════════════\n\n")

	bestCategories := []struct {
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

	for _, cat := range bestCategories {
		bestIdx := 0
		bestScore := -1.0
		for i, r := range results {
			score := cat.get(r).Score
			if score > bestScore {
				bestScore = score
				bestIdx = i
			}
		}
		sb.WriteString(fmt.Sprintf("%-15s: %s (%.2f)\n", cat.name, names[bestIdx], bestScore))
	}
	sb.WriteString("\n")

	// Deck lists
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

	// Verbose analysis
	if verbose {
		sb.WriteString("📋 DETAILED ANALYSIS\n")
		sb.WriteString("════════════════════\n\n")

		for i, r := range results {
			sb.WriteString(fmt.Sprintf("═══ %s ═══\n\n", names[i]))

			// Defense analysis
			sb.WriteString(fmt.Sprintf("Defense (%.1f/10.0): %s\n", r.DefenseAnalysis.Score, r.DefenseAnalysis.Rating))
			for _, detail := range r.DefenseAnalysis.Details {
				sb.WriteString(fmt.Sprintf("  • %s\n", detail))
			}
			sb.WriteString("\n")

			// Attack analysis
			sb.WriteString(fmt.Sprintf("Attack (%.1f/10.0): %s\n", r.AttackAnalysis.Score, r.AttackAnalysis.Rating))
			for _, detail := range r.AttackAnalysis.Details {
				sb.WriteString(fmt.Sprintf("  • %s\n", detail))
			}
			sb.WriteString("\n")

			// Synergy matrix
			if r.SynergyMatrix.PairCount > 0 {
				sb.WriteString(fmt.Sprintf("Synergy: %d pairs found (%.1f%% coverage)\n",
					r.SynergyMatrix.PairCount, r.SynergyMatrix.SynergyCoverage))
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

// formatComparisonJSON formats deck comparison as JSON
func formatComparisonJSON(names []string, results []evaluation.EvaluationResult) (string, error) {
	comparison := map[string]interface{}{
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
