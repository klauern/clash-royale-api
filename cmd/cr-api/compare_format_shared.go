package main

import (
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

type compareCategoryWinner struct {
	label      string
	deckName   string
	score      float64
	rating     string
	assessment string
}

type compareAnalysisSection struct {
	label   string
	score   float64
	rating  string
	details []string
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

func buildBestInCategoryEntries(names []string, results []evaluation.EvaluationResult, includeOverall bool) []compareCategoryWinner {
	entries := make([]compareCategoryWinner, 0, len(getEvaluationCategories())+1)
	if includeOverall {
		bestOverallIdx := findBestOverallDeck(results)
		entries = append(entries, compareCategoryWinner{
			label:    "Overall",
			deckName: names[bestOverallIdx],
			score:    results[bestOverallIdx].OverallScore,
			rating:   string(results[bestOverallIdx].OverallRating),
		})
	}

	for _, cat := range getEvaluationCategories() {
		bestIdx := findBestDeckIndex(results, cat.get)
		bestScore := cat.get(results[bestIdx])
		entries = append(entries, compareCategoryWinner{
			label:      cat.name,
			deckName:   names[bestIdx],
			score:      bestScore.Score,
			rating:     string(bestScore.Rating),
			assessment: bestScore.Assessment,
		})
	}

	return entries
}

func groupDeckCards(deck []string, cardsPerRow int) [][]string {
	if cardsPerRow <= 0 {
		return nil
	}
	rows := make([][]string, 0, (len(deck)+cardsPerRow-1)/cardsPerRow)
	for i := 0; i < len(deck); i += cardsPerRow {
		end := min(i+cardsPerRow, len(deck))
		rows = append(rows, deck[i:end])
	}
	return rows
}

func buildCoreAnalysisSections(result evaluation.EvaluationResult) []compareAnalysisSection {
	return []compareAnalysisSection{
		{
			label:   "Defense",
			score:   result.DefenseAnalysis.Score,
			rating:  string(result.DefenseAnalysis.Rating),
			details: result.DefenseAnalysis.Details,
		},
		{
			label:   "Attack",
			score:   result.AttackAnalysis.Score,
			rating:  string(result.AttackAnalysis.Rating),
			details: result.AttackAnalysis.Details,
		},
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
