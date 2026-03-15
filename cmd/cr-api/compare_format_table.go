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

	sb.WriteString("╔════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                       DECK COMPARISON                                ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════════════╝\n\n")

	formatTableOverviewSection(&sb, names, results)
	formatTableCategoryScoresSection(&sb, names, results)
	formatTableBestInCategorySection(&sb, names, results)
	if showWinRate {
		formatTableWinRateSection(&sb, names, results)
	}
	formatTableDeckCompositionSection(&sb, names, results)

	if verbose {
		formatTableVerboseAnalysisSection(&sb, names, results)
	}

	return sb.String()
}

func formatTableOverviewSection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("📊 OVERVIEW\n")
	sb.WriteString("════════════\n\n")
	sb.WriteString(fmt.Sprintf("%-15s", "Deck"))
	for _, name := range names {
		sb.WriteString(fmt.Sprintf(" | %-20s", truncate(name, 20)))
	}
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", 15+23*len(names)) + "\n")

	sb.WriteString(fmt.Sprintf("%-15s", "Overall Score"))
	for _, r := range results {
		stars := calculateStars(r.OverallScore)
		rating := formatStarsDisplay(stars)
		sb.WriteString(fmt.Sprintf(" | %.2f %s %-13s", r.OverallScore, rating, r.OverallRating))
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("%-15s", "Avg Elixir"))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf(" | %-20.2f", r.AvgElixir))
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("%-15s", "Archetype"))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf(" | %-20s", r.DetectedArchetype))
	}
	sb.WriteString("\n\n")
}

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

func formatTableBestInCategorySection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("🏆 BEST IN CATEGORY\n")
	sb.WriteString("══════════════════\n\n")

	bestOverallIdx := findBestOverallDeck(results)
	sb.WriteString(fmt.Sprintf("%-15s: %s (%.2f)\n", "Overall", names[bestOverallIdx], results[bestOverallIdx].OverallScore))

	categories := getEvaluationCategories()

	for _, cat := range categories {
		bestIdx := findBestDeckIndex(results, cat.get)
		bestScore := cat.get(results[bestIdx]).Score
		sb.WriteString(fmt.Sprintf("%-15s: %s (%.2f)\n", cat.name, names[bestIdx], bestScore))
	}
	sb.WriteString("\n")
}

func formatTableWinRateSection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("🎯 PREDICTED WIN RATE (score-based estimate)\n")
	sb.WriteString("═════════════════════════════════════════════\n\n")
	for i, r := range results {
		predicted := estimateWinRateFromScore(r.OverallScore)
		sb.WriteString(fmt.Sprintf("%-20s: %5.1f%%\n", truncate(names[i], 20), predicted))
	}
	sb.WriteString("\n")
}

func estimateWinRateFromScore(score float64) float64 {
	estimated := 50.0 + (score-5.0)*4.0
	return math.Max(30.0, math.Min(70.0, estimated))
}

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
			fmt.Fprintf(sb, "Synergy: %d pairs found (%.1f%% coverage)\n", r.SynergyMatrix.PairCount, r.SynergyMatrix.SynergyCoverage)
			sb.WriteString("\n")
		}
	}
}
