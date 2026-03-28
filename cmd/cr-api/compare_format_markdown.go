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
	model := buildComparisonRenderModel(names, results)

	sb.WriteString("# Deck Comparison\n\n")
	sb.WriteString(fmt.Sprintf("*Comparing %d decks*\n\n", len(names)))

	formatMarkdownOverviewSection(&sb, model)
	formatMarkdownCategoryScoresSection(&sb, model)
	formatMarkdownBestInCategorySection(&sb, model)
	formatMarkdownDeckCompositionsSection(&sb, model)

	if verbose {
		formatMarkdownVerboseAnalysisSection(&sb, model)
	}

	return sb.String()
}

func formatMarkdownOverviewSection(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("## Overview\n\n")
	sb.WriteString("| Deck | Overall Score | Rating | Avg Elixir | Archetype |\n")
	sb.WriteString("|------|--------------|--------|------------|------------|\n")

	for _, deck := range model.Decks {
		r := deck.Result
		stars := formatStarsDisplay(calculateStars(r.OverallScore))
		sb.WriteString(fmt.Sprintf("| %s | %.2f %s | %s | %.2f | %s |\n",
			deck.Name, r.OverallScore, stars, r.OverallRating, r.AvgElixir, r.DetectedArchetype))
	}
	sb.WriteString("\n")
}

func formatMarkdownCategoryScoresSection(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("## Category Scores\n\n")
	sb.WriteString("| Category | ")
	for _, deck := range model.Decks {
		sb.WriteString(fmt.Sprintf("%s | ", deck.Name))
	}
	sb.WriteString("\n|----------|")
	for range model.Decks {
		sb.WriteString("---------|")
	}
	sb.WriteString("\n")

	for _, category := range model.Categories {
		sb.WriteString(fmt.Sprintf("| **%s** | ", category.Name))
		for _, score := range category.Scores {
			stars := formatStarsDisplay(score.Stars)
			sb.WriteString(fmt.Sprintf("%.1f %s | ", score.Score, stars))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
}

func formatMarkdownBestInCategorySection(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("## 🏆 Best in Category\n\n")
	for _, category := range model.Categories {
		winner := model.Decks[category.WinnerIdx]
		bestScore := category.Scores[category.WinnerIdx].Score
		sb.WriteString(fmt.Sprintf("- **%s**: %s (%.2f)\n", category.Name, winner.Name, bestScore))
	}
	sb.WriteString("\n")
}

func formatMarkdownDeckCompositionsSection(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("## Deck Compositions\n\n")
	for _, deck := range model.Decks {
		sb.WriteString(fmt.Sprintf("### %s\n\n", deck.Name))
		sb.WriteString("```\n")
		for j, card := range deck.Result.Deck {
			sb.WriteString(fmt.Sprintf("%-18s", card))
			if (j+1)%4 == 0 {
				sb.WriteString("\n")
			}
		}
		sb.WriteString("```\n\n")
	}
}

func formatMarkdownVerboseAnalysisSection(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("## Detailed Analysis\n\n")
	for _, deck := range model.Decks {
		r := deck.Result
		fmt.Fprintf(sb, "### %s\n\n", deck.Name)

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
			fmt.Fprintf(sb, "**Synergy**: %d pairs found (%.1f%% coverage)\n\n", r.SynergyMatrix.PairCount, r.SynergyMatrix.SynergyCoverage)
		}
	}
}
