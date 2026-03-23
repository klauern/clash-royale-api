package main

import (
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

type categoryWinner struct {
	name     string
	score    evaluation.CategoryScore
	deckName string
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

func findBestDeckIndex(results []evaluation.EvaluationResult, getScore func(evaluation.EvaluationResult) evaluation.CategoryScore) int {
	bestIdx := 0
	bestScore := -1.0
	for i, r := range results {
		score := getScore(r).Score
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	return bestIdx
}

func findBestOverallDeck(results []evaluation.EvaluationResult) int {
	bestIdx := 0
	bestScore := -1.0
	for i, r := range results {
		if r.OverallScore > bestScore {
			bestScore = r.OverallScore
			bestIdx = i
		}
	}
	return bestIdx
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

func computeCategoryWinners(names []string, results []evaluation.EvaluationResult) []categoryWinner {
	categories := getEvaluationCategories()
	winners := make([]categoryWinner, 0, len(categories))

	for _, cat := range categories {
		bestIdx := findBestDeckIndex(results, cat.get)
		winners = append(winners, categoryWinner{
			name:     cat.name,
			score:    cat.get(results[bestIdx]),
			deckName: names[bestIdx],
		})
	}

	return winners
}
