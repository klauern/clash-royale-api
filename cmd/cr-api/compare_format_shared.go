package main

import (
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

type comparisonCategoryWinner struct {
	CategoryName string
	DeckName     string
	Score        evaluation.CategoryScore
}

type comparisonAnalysisSection struct {
	Label   string
	Score   float64
	Rating  evaluation.Rating
	Details []string
}

type comparisonSynergySummary struct {
	PairCount      int
	Coverage       float64
	AverageSynergy float64
}

func getEvaluationCategories() []struct {
	name string
	get  func(evaluation.EvaluationResult) evaluation.CategoryScore
} {
	return []struct {
		name string
		get  func(evaluation.EvaluationResult) evaluation.CategoryScore
	}{
		{"Attack", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.Attack }},
		{"Defense", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.Defense }},
		{"Synergy", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.Synergy }},
		{"Versatility", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.Versatility }},
		{"F2P Friendly", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.F2PFriendly }},
		{"Playability", func(r evaluation.EvaluationResult) evaluation.CategoryScore { return r.Playability }},
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatStarsDisplay(count int) string {
	filled := "★"
	empty := "☆"

	stars := strings.Repeat(filled, count)
	stars += strings.Repeat(empty, 3-count)

	return stars
}

func calculateStars(score float64) int {
	switch {
	case score >= 9:
		return 3
	case score >= 7:
		return 2
	case score >= 5:
		return 1
	default:
		return 0
	}
}

func collectCategoryWinners(vm comparisonViewModel) []comparisonCategoryWinner {
	winners := make([]comparisonCategoryWinner, 0, len(vm.Categories))
	for _, category := range vm.Categories {
		if category.BestDeckIndex < 0 || category.BestDeckIndex >= len(vm.Decks) || category.BestDeckIndex >= len(category.Scores) {
			continue
		}

		winners = append(winners, comparisonCategoryWinner{
			CategoryName: category.Name,
			DeckName:     vm.Decks[category.BestDeckIndex].Name,
			Score:        category.Scores[category.BestDeckIndex],
		})
	}

	return winners
}

func writeDeckCardGrid(sb *strings.Builder, cards []string, cardWidth int) {
	for i, card := range cards {
		sb.WriteString(padRight(card, cardWidth))
		if (i+1)%4 == 0 {
			sb.WriteString("\n")
		}
	}
	if len(cards)%4 != 0 {
		sb.WriteString("\n")
	}
}

func prefixCards(cards []string, prefix string) []string {
	prefixed := make([]string, len(cards))
	for i, card := range cards {
		prefixed[i] = prefix + card
	}
	return prefixed
}

func collectAnalysisSections(r evaluation.EvaluationResult) []comparisonAnalysisSection {
	return []comparisonAnalysisSection{
		{
			Label:   "Defense",
			Score:   r.DefenseAnalysis.Score,
			Rating:  r.DefenseAnalysis.Rating,
			Details: r.DefenseAnalysis.Details,
		},
		{
			Label:   "Attack",
			Score:   r.AttackAnalysis.Score,
			Rating:  r.AttackAnalysis.Rating,
			Details: r.AttackAnalysis.Details,
		},
	}
}

func getSynergySummary(r evaluation.EvaluationResult) (comparisonSynergySummary, bool) {
	if r.SynergyMatrix.PairCount <= 0 {
		return comparisonSynergySummary{}, false
	}

	return comparisonSynergySummary{
		PairCount:      r.SynergyMatrix.PairCount,
		Coverage:       r.SynergyMatrix.SynergyCoverage,
		AverageSynergy: r.SynergyMatrix.AverageSynergy,
	}, true
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
