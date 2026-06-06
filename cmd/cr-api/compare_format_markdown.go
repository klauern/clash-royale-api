//nolint:staticcheck // Formatter output construction intentionally uses string builder composition.
package main

import (
	"fmt"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

//nolint:staticcheck // Keep explicit builder writes for formatter readability and parity.
func formatComparisonMarkdown(names []string, results []evaluation.EvaluationResult, verbose bool) string {
	var sb strings.Builder
	vm := buildComparisonViewModel(names, results)

	sb.WriteString("# Deck Comparison\n\n")
	sb.WriteString(fmt.Sprintf("*Comparing %d decks*\n\n", len(vm.Decks)))

	formatMarkdownOverviewSection(&sb, vm)
	formatMarkdownCategoryScoresSection(&sb, vm)
	formatMarkdownBestInCategorySection(&sb, vm)
	formatMarkdownDeckCompositionsSection(&sb, vm)

	if verbose {
		formatMarkdownVerboseAnalysisSection(&sb, vm)
	}

	return sb.String()
}

func formatMarkdownOverviewSection(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("## Overview\n\n")
	sb.WriteString("| Deck | Overall Score | Rating | Avg Elixir | Archetype |\n")
	sb.WriteString("|------|--------------|--------|------------|------------|\n")

	for _, deck := range vm.Decks {
		r := deck.Result
		sb.WriteString(fmt.Sprintf("| %s | %.2f %s | %s | %.2f | %s |\n",
			deck.Name, r.OverallScore, deck.OverallStars, r.OverallRating, r.AvgElixir, r.DetectedArchetype))
	}
	sb.WriteString("\n")
}

func formatMarkdownCategoryScoresSection(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("## Category Scores\n\n")
	sb.WriteString("| Category | ")
	for _, deck := range vm.Decks {
		sb.WriteString(fmt.Sprintf("%s | ", deck.Name))
	}
	sb.WriteString("\n|----------|")
	for range vm.Decks {
		sb.WriteString("---------|")
	}
	sb.WriteString("\n")

	for _, category := range vm.Categories {
		sb.WriteString(fmt.Sprintf("| **%s** | ", category.Name))
		for _, score := range category.Scores {
			sb.WriteString(fmt.Sprintf("%.1f %s | ", score.Score, formatStarsDisplay(score.Stars)))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

func formatMarkdownBestInCategorySection(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("## 🏆 Best in Category\n\n")
	for _, winner := range collectCategoryWinners(vm) {
		sb.WriteString(fmt.Sprintf("- **%s**: %s (%.2f)\n", winner.CategoryName, winner.DeckName, winner.Score.Score))
	}
	sb.WriteString("\n")
}

func formatMarkdownDeckCompositionsSection(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("## Deck Compositions\n\n")
	for _, deck := range vm.Decks {
		sb.WriteString(fmt.Sprintf("### %s\n\n", deck.Name))
		sb.WriteString("```\n")
		writeDeckCardGrid(sb, deck.Result.Deck, 18)
		sb.WriteString("```\n\n")
	}
}

func formatMarkdownVerboseAnalysisSection(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("## Detailed Analysis\n\n")
	for _, deck := range vm.Decks {
		r := deck.Result
		fmt.Fprintf(sb, "### %s\n\n", deck.Name)

		for _, section := range collectAnalysisSections(r) {
			fmt.Fprintf(sb, "**%s** (%.1f/10.0): %s\n\n", section.Label, section.Score, section.Rating)
			for _, detail := range section.Details {
				fmt.Fprintf(sb, "- %s\n", detail)
			}
			sb.WriteString("\n")
		}

		if synergy, ok := getSynergySummary(r); ok {
			fmt.Fprintf(sb, "**Synergy**: %d pairs found (%.1f%% coverage)\n\n", synergy.PairCount, synergy.Coverage)
		}
	}
}
