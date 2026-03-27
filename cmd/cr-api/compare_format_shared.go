package main

import (
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

type comparisonCategoryResult struct {
	Name         string
	Scores       []evaluation.CategoryScore
	BestDeckIdx  int
	BestDeckName string
}

type comparisonRenderModel struct {
	Names         []string
	Results       []evaluation.EvaluationResult
	BestOverallIx int
	Categories    []comparisonCategoryResult
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

func buildComparisonRenderModel(names []string, results []evaluation.EvaluationResult) comparisonRenderModel {
	model := comparisonRenderModel{
		Names:         names,
		Results:       results,
		BestOverallIx: findBestOverallDeck(results),
		Categories:    make([]comparisonCategoryResult, 0, len(getEvaluationCategories())),
	}

	for _, category := range getEvaluationCategories() {
		bestIdx := findBestDeckIndex(results, category.get)
		scores := make([]evaluation.CategoryScore, len(results))
		for i, result := range results {
			scores[i] = category.get(result)
		}

		model.Categories = append(model.Categories, comparisonCategoryResult{
			Name:         category.name,
			Scores:       scores,
			BestDeckIdx:  bestIdx,
			BestDeckName: names[bestIdx],
		})
	}

	return model
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
