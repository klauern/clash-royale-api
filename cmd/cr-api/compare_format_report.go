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
	model := buildComparisonRenderModel(names, results)

	sb.WriteString("# Deck Comparison Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated**: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Decks Compared**: %d\n\n", len(names)))
	sb.WriteString("---\n\n")

	formatReportExecutiveSummary(&sb, model)
	formatReportDetailedScoreComparison(&sb, model)
	formatReportCategoryChampions(&sb, model)
	formatReportDeckDetails(&sb, model)
	formatReportRecommendations(&sb, model)

	return sb.String()
}

func formatReportExecutiveSummary(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("## Executive Summary\n\n")

	bestDeck := model.Decks[model.BestOverallIdx]
	sb.WriteString(fmt.Sprintf("### 🏆 Recommended Deck: **%s**\n\n", bestDeck.Name))
	sb.WriteString(fmt.Sprintf("- **Overall Score**: %.2f/10.0 (%s)\n", bestDeck.Result.OverallScore, bestDeck.Result.OverallRating))
	sb.WriteString(fmt.Sprintf("- **Archetype**: %s\n", bestDeck.Result.DetectedArchetype))
	sb.WriteString(fmt.Sprintf("- **Average Elixir**: %.2f\n\n", bestDeck.Result.AvgElixir))

	formatReportRankings(sb, model)
	sb.WriteString("\n---\n\n")
}

func formatReportRankings(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("### Overall Rankings\n\n")
	indices := make([]int, len(model.Decks))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool {
		return model.Decks[indices[i]].Result.OverallScore > model.Decks[indices[j]].Result.OverallScore
	})

	for rank, idx := range indices {
		emoji := getRankingEmoji(rank + 1)
		deck := model.Decks[idx]
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

func formatReportDetailedScoreComparison(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("## Detailed Score Comparison\n\n")
	sb.WriteString("| Metric | ")
	for _, deck := range model.Decks {
		sb.WriteString(fmt.Sprintf("%s | ", truncate(deck.Name, 18)))
	}
	sb.WriteString("\n|--------|")
	for range model.Decks {
		sb.WriteString("--------------------|")
	}
	sb.WriteString("\n")

	sb.WriteString("| **Overall** | ")
	for _, deck := range model.Decks {
		r := deck.Result
		sb.WriteString(fmt.Sprintf("%.2f %s | ", r.OverallScore, formatStarsDisplay(calculateStars(r.OverallScore))))
	}
	sb.WriteString("\n")

	for _, category := range model.Categories {
		sb.WriteString(fmt.Sprintf("| %s | ", category.Name))
		for _, score := range category.Scores {
			sb.WriteString(fmt.Sprintf("%.1f %s | ", score.Score, formatStarsDisplay(score.Stars)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("| Avg Elixir | ")
	for _, deck := range model.Decks {
		r := deck.Result
		sb.WriteString(fmt.Sprintf("%.2f | ", r.AvgElixir))
	}
	sb.WriteString("\n")

	sb.WriteString("| Archetype | ")
	for _, deck := range model.Decks {
		r := deck.Result
		sb.WriteString(fmt.Sprintf("%s | ", r.DetectedArchetype))
	}
	sb.WriteString("\n\n---\n\n")
}

func formatReportCategoryChampions(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("## Category Champions\n\n")
	for _, category := range model.Categories {
		winner := model.Decks[category.WinnerIdx]
		winningScore := category.Scores[category.WinnerIdx]
		sb.WriteString(fmt.Sprintf("### 🏆 Best %s: **%s**\n\n", category.Name, winner.Name))
		sb.WriteString(fmt.Sprintf("- **Score**: %.1f/10.0 (%s)\n", winningScore.Score, winningScore.Rating))
		sb.WriteString(fmt.Sprintf("- **Assessment**: %s\n\n", winningScore.Assessment))
	}

	sb.WriteString("---\n\n")
}

func formatReportDeckDetails(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("## Deck Details\n\n")
	for i, deck := range model.Decks {
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

func formatReportRecommendations(sb *strings.Builder, model comparisonRenderModel) {
	sb.WriteString("## Recommendations\n\n")
	bestDeck := model.Decks[model.BestOverallIdx]
	sb.WriteString(fmt.Sprintf("Based on the analysis, **%s** is the strongest deck overall with a score of %.2f/10.0.\n\n", bestDeck.Name, bestDeck.Result.OverallScore))

	sb.WriteString("### When to Use Each Deck\n\n")
	for _, deck := range model.Decks {
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
