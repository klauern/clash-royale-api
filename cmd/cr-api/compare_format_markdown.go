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
	sb.WriteString(fmt.Sprintf("*Comparing %d decks*\n\n", len(model.Names)))

	formatMarkdownOverviewSection(&sb, names, results)
	formatMarkdownCategoryScoresSectionWithModel(&sb, model)
	formatMarkdownBestInCategorySectionWithModel(&sb, model)
	formatMarkdownDeckCompositionsSection(&sb, names, results)

	if verbose {
		formatMarkdownVerboseAnalysisSection(&sb, names, results)
	}

	return sb.String()
}

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

func formatMarkdownCategoryScoresSection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	formatMarkdownCategoryScoresSectionWithModel(sb, buildComparisonRenderModel(names, results))
}

func formatMarkdownCategoryScoresSectionWithModel(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("## Category Scores\n\n")
	sb.WriteString("| Category | ")
	for _, name := range model.Names {
		sb.WriteString(fmt.Sprintf("%s | ", name))
	}
	sb.WriteString("\n|----------|")
	for range model.Names {
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

func formatMarkdownBestInCategorySection(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	formatMarkdownBestInCategorySectionWithModel(sb, buildComparisonRenderModel(names, results))
}

func formatMarkdownBestInCategorySectionWithModel(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("## 🏆 Best in Category\n\n")
	for _, category := range model.Categories {
		bestScore := category.Scores[category.BestDeckIdx].Score
		sb.WriteString(fmt.Sprintf("- **%s**: %s (%.2f)\n", category.Name, category.BestDeckName, bestScore))
	}
	sb.WriteString("\n")
}

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
			fmt.Fprintf(sb, "**Synergy**: %d pairs found (%.1f%% coverage)\n\n", r.SynergyMatrix.PairCount, r.SynergyMatrix.SynergyCoverage)
		}
	}
}
