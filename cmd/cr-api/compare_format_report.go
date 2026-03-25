//nolint:staticcheck // Formatter output construction intentionally uses string builder composition.
package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

//nolint:staticcheck // Keep explicit builder writes for formatter readability and parity.
func generateComparisonReport(names []string, results []evaluation.EvaluationResult) string {
	var sb strings.Builder

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

func formatReportExecutiveSummary(sb *strings.Builder, names []string, results []evaluation.EvaluationResult, bestIdx int) {
	sb.WriteString("## Executive Summary\n\n")

	sb.WriteString(fmt.Sprintf("### 🏆 Recommended Deck: **%s**\n\n", names[bestIdx]))
	sb.WriteString(fmt.Sprintf("- **Overall Score**: %.2f/10.0 (%s)\n", results[bestIdx].OverallScore, results[bestIdx].OverallRating))
	sb.WriteString(fmt.Sprintf("- **Archetype**: %s\n", results[bestIdx].DetectedArchetype))
	sb.WriteString(fmt.Sprintf("- **Average Elixir**: %.2f\n\n", results[bestIdx].AvgElixir))

	formatReportRankings(sb, names, results)
	sb.WriteString("\n---\n\n")
}

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
		sb.WriteString(fmt.Sprintf("%s **%s** - %.2f/10.0 (%s)\n", emoji, names[idx], results[idx].OverallScore, results[idx].OverallRating))
	}
}

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

func formatReportCategoryChampions(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("## Category Champions\n\n")
	for _, winner := range computeCategoryWinners(names, results) {
		sb.WriteString(fmt.Sprintf("### 🏆 Best %s: **%s**\n\n", winner.categoryName, winner.deckName))
		sb.WriteString(fmt.Sprintf("- **Score**: %.1f/10.0 (%s)\n", winner.score.Score, winner.score.Rating))
		sb.WriteString(fmt.Sprintf("- **Assessment**: %s\n\n", winner.score.Assessment))
	}

	sb.WriteString("---\n\n")
}

func formatReportDeckDetails(sb *strings.Builder, names []string, results []evaluation.EvaluationResult) {
	sb.WriteString("## Deck Details\n\n")
	for i, name := range names {
		r := results[i]
		sb.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, name))

		sb.WriteString("**Cards**:\n```\n")
		for j, card := range r.Deck {
			sb.WriteString(fmt.Sprintf("%-20s", card))
			if (j+1)%4 == 0 {
				sb.WriteString("\n")
			}
		}
		sb.WriteString("```\n\n")

		sb.WriteString("**Key Statistics**:\n")
		sb.WriteString(fmt.Sprintf("- Overall Score: %.2f/10.0 (%s)\n", r.OverallScore, r.OverallRating))
		sb.WriteString(fmt.Sprintf("- Archetype: %s (%.0f%% confidence)\n", r.DetectedArchetype, r.ArchetypeConfidence*100))
		sb.WriteString(fmt.Sprintf("- Average Elixir: %.2f\n\n", r.AvgElixir))

		formatDeckStrengthsAndWeaknesses(sb, r)
		formatDeckAnalysis(sb, r)

		sb.WriteString("---\n\n")
	}
}

func formatDeckStrengthsAndWeaknesses(sb *strings.Builder, r evaluation.EvaluationResult) {
	sb.WriteString("**Strengths**:\n")
	categoryEntries := getEvaluationCategories()
	strengths := make([]struct {
		name  string
		score evaluation.CategoryScore
	}, 0, len(categoryEntries))
	for _, cat := range categoryEntries {
		strengths = append(strengths, struct {
			name  string
			score evaluation.CategoryScore
		}{
			name:  cat.name,
			score: cat.get(r),
		})
	}

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
	for i := len(strengths) - 1; i >= max(0, len(strengths)-3); i-- {
		s := strengths[i]
		if s.score.Score < 7.0 {
			fmt.Fprintf(sb, "- %s: %.1f/10.0 - %s\n", s.name, s.score.Score, s.score.Assessment)
		}
	}
	sb.WriteString("\n")
}

func formatDeckAnalysis(sb *strings.Builder, r evaluation.EvaluationResult) {
	if len(r.DefenseAnalysis.Details) > 0 {
		fmt.Fprintf(sb, "**Defense Analysis** (%.1f/10.0):\n", r.DefenseAnalysis.Score)
		for _, detail := range r.DefenseAnalysis.Details {
			fmt.Fprintf(sb, "- %s\n", detail)
		}
		sb.WriteString("\n")
	}

	if len(r.AttackAnalysis.Details) > 0 {
		fmt.Fprintf(sb, "**Attack Analysis** (%.1f/10.0):\n", r.AttackAnalysis.Score)
		for _, detail := range r.AttackAnalysis.Details {
			fmt.Fprintf(sb, "- %s\n", detail)
		}
		sb.WriteString("\n")
	}

	if r.SynergyMatrix.PairCount > 0 {
		fmt.Fprintf(sb, "**Synergy**: %d card pairs found (%.1f%% coverage, avg synergy: %.2f)\n\n", r.SynergyMatrix.PairCount, r.SynergyMatrix.SynergyCoverage, r.SynergyMatrix.AverageSynergy)
	}
}

func formatReportRecommendations(sb *strings.Builder, names []string, results []evaluation.EvaluationResult, bestIdx int) {
	sb.WriteString("## Recommendations\n\n")
	sb.WriteString(fmt.Sprintf("Based on the analysis, **%s** is the strongest deck overall with a score of %.2f/10.0.\n\n", names[bestIdx], results[bestIdx].OverallScore))

	sb.WriteString("### When to Use Each Deck\n\n")
	for i, name := range names {
		r := results[i]
		sb.WriteString(fmt.Sprintf("**%s** (%s archetype):\n", name, r.DetectedArchetype))

		switch {
		case r.Attack.Score > r.Defense.Score:
			sb.WriteString("- Best for: Aggressive playstyle, ladder pushing\n")
		case r.Defense.Score > r.Attack.Score:
			sb.WriteString("- Best for: Defensive counter-attacks, conservative play\n")
		default:
			sb.WriteString("- Best for: Balanced matchups, versatile gameplay\n")
		}

		if r.F2PFriendly.Score >= 8.0 {
			sb.WriteString("- Excellent for F2P players\n")
		}

		sb.WriteString("\n")
	}
}
