//nolint:staticcheck // Formatter output construction intentionally uses string builder composition.
package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

//nolint:staticcheck // Keep explicit builder writes for formatter readability and parity.
func formatComparisonTable(names []string, results []evaluation.EvaluationResult, verbose, showWinRate bool) string {
	var sb strings.Builder
	model := buildComparisonRenderModel(names, results)

	sb.WriteString("╔════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                       DECK COMPARISON                                ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════════════╝\n\n")

	formatTableOverviewSection(&sb, model)
	formatTableCategoryScoresSection(&sb, model)
	formatTableBestInCategorySection(&sb, model)
	if showWinRate {
		formatTableWinRateSection(&sb, model)
	}
	formatTableDeckCompositionSection(&sb, model)

	if verbose {
		formatTableVerboseAnalysisSection(&sb, model)
	}

	return sb.String()
}

func formatTableOverviewSection(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("📊 OVERVIEW\n")
	sb.WriteString("════════════\n\n")
	sb.WriteString(fmt.Sprintf("%-15s", "Deck"))
	for _, deck := range model.Decks {
		sb.WriteString(fmt.Sprintf(" | %-20s", truncate(deck.Name, 20)))
	}
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", 15+23*len(model.Decks)) + "\n")

	sb.WriteString(fmt.Sprintf("%-15s", "Overall Score"))
	for _, deck := range model.Decks {
		r := deck.Result
		stars := calculateStars(r.OverallScore)
		rating := formatStarsDisplay(stars)
		sb.WriteString(fmt.Sprintf(" | %.2f %s %-13s", r.OverallScore, rating, r.OverallRating))
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("%-15s", "Avg Elixir"))
	for _, deck := range model.Decks {
		r := deck.Result
		sb.WriteString(fmt.Sprintf(" | %-20.2f", r.AvgElixir))
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("%-15s", "Archetype"))
	for _, deck := range model.Decks {
		r := deck.Result
		sb.WriteString(fmt.Sprintf(" | %-20s", r.DetectedArchetype))
	}
	sb.WriteString("\n\n")
}

func formatTableCategoryScoresSection(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("📈 CATEGORY SCORES\n")
	sb.WriteString("══════════════════\n\n")
	sb.WriteString(fmt.Sprintf("%-15s", "Category"))
	for _, deck := range model.Decks {
		sb.WriteString(fmt.Sprintf(" | %-20s", truncate(deck.Name, 20)))
	}
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", 15+23*len(model.Decks)) + "\n")

	for _, category := range model.Categories {
		sb.WriteString(fmt.Sprintf("%-15s", category.Name))
		for _, score := range category.Scores {
			rating := formatStarsDisplay(score.Stars)
			sb.WriteString(fmt.Sprintf(" | %.1f %s %-14s", score.Score, rating, score.Rating))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

func formatTableBestInCategorySection(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("🏆 BEST IN CATEGORY\n")
	sb.WriteString("══════════════════\n\n")

	bestOverall := model.Decks[model.BestOverallIdx]
	sb.WriteString(fmt.Sprintf("%-15s: %s (%.2f)\n", "Overall", bestOverall.Name, bestOverall.Result.OverallScore))

	for _, category := range model.Categories {
		winner := model.Decks[category.WinnerIdx]
		bestScore := category.Scores[category.WinnerIdx].Score
		sb.WriteString(fmt.Sprintf("%-15s: %s (%.2f)\n", category.Name, winner.Name, bestScore))
	}
	sb.WriteString("\n")
}

func formatTableWinRateSection(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("🎯 PREDICTED WIN RATE (score-based estimate)\n")
	sb.WriteString("═════════════════════════════════════════════\n\n")
	for _, deck := range model.Decks {
		r := deck.Result
		predicted := estimateWinRateFromScore(r.OverallScore)
		sb.WriteString(fmt.Sprintf("%-20s: %5.1f%%\n", truncate(deck.Name, 20), predicted))
	}
	sb.WriteString("\n")
}

func estimateWinRateFromScore(score float64) float64 {
	estimated := 50.0 + (score-5.0)*4.0
	return math.Max(30.0, math.Min(70.0, estimated))
}

func formatTableDeckCompositionSection(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("🃏 DECK COMPOSITION\n")
	sb.WriteString("══════════════════\n\n")
	for _, deck := range model.Decks {
		r := deck.Result
		sb.WriteString(fmt.Sprintf("%s:\n", deck.Name))
		for j, card := range r.Deck {
			if j > 0 && j%4 == 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("  %-18s", card))
		}
		sb.WriteString("\n\n")
	}
}

func formatTableVerboseAnalysisSection(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("📋 DETAILED ANALYSIS\n")
	sb.WriteString("════════════════════\n\n")

	for _, deck := range model.Decks {
		r := deck.Result
		fmt.Fprintf(sb, "═══ %s ═══\n\n", deck.Name)

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
			fmt.Fprintf(sb, "Synergy: %d pairs found (%.1f%% coverage)\n", r.SynergyMatrix.PairCount, r.SynergyMatrix.SynergyCoverage)
			sb.WriteString("\n")
		}
	}
}
