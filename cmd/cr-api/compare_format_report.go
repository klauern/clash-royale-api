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
	vm := buildComparisonViewModel(names, results)

	sb.WriteString("# Deck Comparison Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated**: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Decks Compared**: %d\n\n", len(vm.Decks)))
	sb.WriteString("---\n\n")

	formatReportExecutiveSummary(&sb, vm)
	formatReportDetailedScoreComparison(&sb, vm)
	formatReportCategoryChampions(&sb, vm)
	formatReportDeckDetails(&sb, vm)
	formatReportRecommendations(&sb, vm)

	return sb.String()
}

func formatReportExecutiveSummary(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("## Executive Summary\n\n")

	best := vm.Decks[vm.BestOverallIndex]
	sb.WriteString(fmt.Sprintf("### 🏆 Recommended Deck: **%s**\n\n", best.Name))
	sb.WriteString(fmt.Sprintf("- **Overall Score**: %.2f/10.0 (%s)\n", best.Result.OverallScore, best.Result.OverallRating))
	sb.WriteString(fmt.Sprintf("- **Archetype**: %s\n", best.Result.DetectedArchetype))
	sb.WriteString(fmt.Sprintf("- **Average Elixir**: %.2f\n\n", best.Result.AvgElixir))

	formatReportRankings(sb, vm)
	sb.WriteString("\n---\n\n")
}

func formatReportRankings(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("### Overall Rankings\n\n")

	for rank, idx := range vm.RankedDecks {
		emoji := getRankingEmoji(rank + 1)
		deck := vm.Decks[idx]
		sb.WriteString(fmt.Sprintf("%s **%s** - %.2f/10.0 (%s)\n", emoji, deck.Name, deck.Result.OverallScore, deck.Result.OverallRating))
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

func formatReportDetailedScoreComparison(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("## Detailed Score Comparison\n\n")
	sb.WriteString("| Metric | ")
	for _, deck := range vm.Decks {
		sb.WriteString(fmt.Sprintf("%s | ", deck.TruncatedName18))
	}
	sb.WriteString("\n|--------|")
	for range vm.Decks {
		sb.WriteString("--------------------|")
	}
	sb.WriteString("\n")

	sb.WriteString("| **Overall** | ")
	for _, deck := range vm.Decks {
		sb.WriteString(fmt.Sprintf("%.2f %s | ", deck.Result.OverallScore, deck.OverallStars))
	}
	sb.WriteString("\n")

	for _, category := range vm.Categories {
		sb.WriteString(fmt.Sprintf("| %s | ", category.Name))
		for _, score := range category.Scores {
			sb.WriteString(fmt.Sprintf("%.1f %s | ", score.Score, formatStarsDisplay(score.Stars)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("| Avg Elixir | ")
	for _, deck := range vm.Decks {
		sb.WriteString(fmt.Sprintf("%.2f | ", deck.Result.AvgElixir))
	}
	sb.WriteString("\n")

	sb.WriteString("| Archetype | ")
	for _, deck := range vm.Decks {
		sb.WriteString(fmt.Sprintf("%s | ", deck.Result.DetectedArchetype))
	}
	sb.WriteString("\n\n---\n\n")
}

func formatReportCategoryChampions(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("## Category Champions\n\n")
	for _, category := range vm.Categories {
		best := vm.Decks[category.BestDeckIndex]
		bestScore := category.Scores[category.BestDeckIndex]
		sb.WriteString(fmt.Sprintf("### 🏆 Best %s: **%s**\n\n", category.Name, best.Name))
		sb.WriteString(fmt.Sprintf("- **Score**: %.1f/10.0 (%s)\n", bestScore.Score, bestScore.Rating))
		sb.WriteString(fmt.Sprintf("- **Assessment**: %s\n\n", bestScore.Assessment))
	}

	sb.WriteString("---\n\n")
}

func formatReportDeckDetails(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("## Deck Details\n\n")
	for i, deck := range vm.Decks {
		r := deck.Result
		sb.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, deck.Name))

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

func formatReportRecommendations(sb *strings.Builder, vm comparisonViewModel) {
	sb.WriteString("## Recommendations\n\n")
	best := vm.Decks[vm.BestOverallIndex]
	sb.WriteString(fmt.Sprintf("Based on the analysis, **%s** is the strongest deck overall with a score of %.2f/10.0.\n\n", best.Name, best.Result.OverallScore))

	sb.WriteString("### When to Use Each Deck\n\n")
	for _, deck := range vm.Decks {
		r := deck.Result
		sb.WriteString(fmt.Sprintf("**%s** (%s archetype):\n", deck.Name, r.DetectedArchetype))

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
