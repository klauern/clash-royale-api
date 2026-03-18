//nolint:staticcheck // Formatter output construction intentionally uses string builder composition.
package main

import (
	"fmt"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

//nolint:staticcheck // Keep explicit builder writes for formatter readability and parity.
func formatComparisonTable(names []string, results []evaluation.EvaluationResult, verbose, showWinRate bool) string {
	var sb strings.Builder
	vm := buildComparisonViewModel(names, results)

	sb.WriteString("╔════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                       DECK COMPARISON                                ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════════════╝\n\n")

	formatTableOverviewSection(&sb, vm)
	formatTableCategoryScoresSection(&sb, vm)
	formatTableBestInCategorySection(&sb, vm)
	if showWinRate {
		formatTableWinRateSection(&sb, vm)
	}
	formatTableDeckCompositionSection(&sb, vm)

	if verbose {
		formatTableVerboseAnalysisSection(&sb, vm)
	}

	return sb.String()
}

func formatTableOverviewSection(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("📊 OVERVIEW\n")
	sb.WriteString("════════════\n\n")
	sb.WriteString(fmt.Sprintf("%-15s", "Deck"))
	for _, deck := range vm.Decks {
		sb.WriteString(fmt.Sprintf(" | %-20s", deck.TruncatedName20))
	}
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", 15+23*len(vm.Decks)) + "\n")

	sb.WriteString(fmt.Sprintf("%-15s", "Overall Score"))
	for _, deck := range vm.Decks {
		sb.WriteString(fmt.Sprintf(" | %.2f %s %-13s", deck.Result.OverallScore, deck.OverallStars, deck.Result.OverallRating))
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("%-15s", "Avg Elixir"))
	for _, deck := range vm.Decks {
		sb.WriteString(fmt.Sprintf(" | %-20.2f", deck.Result.AvgElixir))
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("%-15s", "Archetype"))
	for _, deck := range vm.Decks {
		sb.WriteString(fmt.Sprintf(" | %-20s", deck.Result.DetectedArchetype))
	}
	sb.WriteString("\n\n")
}

func formatTableCategoryScoresSection(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("📈 CATEGORY SCORES\n")
	sb.WriteString("══════════════════\n\n")
	sb.WriteString(fmt.Sprintf("%-15s", "Category"))
	for _, deck := range vm.Decks {
		sb.WriteString(fmt.Sprintf(" | %-20s", deck.TruncatedName20))
	}
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", 15+23*len(vm.Decks)) + "\n")

	for _, category := range vm.Categories {
		sb.WriteString(fmt.Sprintf("%-15s", category.Name))
		for _, score := range category.Scores {
			rating := formatStarsDisplay(score.Stars)
			sb.WriteString(fmt.Sprintf(" | %.1f %s %-14s", score.Score, rating, score.Rating))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

func formatTableBestInCategorySection(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("🏆 BEST IN CATEGORY\n")
	sb.WriteString("══════════════════\n\n")

	bestOverall := vm.Decks[vm.BestOverallIndex]
	sb.WriteString(fmt.Sprintf("%-15s: %s (%.2f)\n", "Overall", bestOverall.Name, bestOverall.Result.OverallScore))

	for _, category := range vm.Categories {
		best := vm.Decks[category.BestDeckIndex]
		bestScore := category.Scores[category.BestDeckIndex].Score
		sb.WriteString(fmt.Sprintf("%-15s: %s (%.2f)\n", category.Name, best.Name, bestScore))
	}
	sb.WriteString("\n")
}

func formatTableWinRateSection(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("🎯 PREDICTED WIN RATE (score-based estimate)\n")
	sb.WriteString("═════════════════════════════════════════════\n\n")
	for _, deck := range vm.Decks {
		sb.WriteString(fmt.Sprintf("%-20s: %5.1f%%\n", deck.TruncatedName20, deck.PredictedWinPct))
	}
	sb.WriteString("\n")
}

func formatTableDeckCompositionSection(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("🃏 DECK COMPOSITION\n")
	sb.WriteString("══════════════════\n\n")
	for _, deck := range vm.Decks {
		sb.WriteString(fmt.Sprintf("%s:\n", deck.Name))
		for j, card := range deck.Result.Deck {
			if j > 0 && j%4 == 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("  %-18s", card))
		}
		sb.WriteString("\n\n")
	}
}

func formatTableVerboseAnalysisSection(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("📋 DETAILED ANALYSIS\n")
	sb.WriteString("════════════════════\n\n")

	for _, deck := range vm.Decks {
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
