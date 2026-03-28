package main

import (
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

type evaluationCategory struct {
	name string
	get  func(evaluation.EvaluationResult) evaluation.CategoryScore
}

type comparisonDeckRender struct {
	Name   string
	Result evaluation.EvaluationResult
}

type comparisonCategoryRender struct {
	Name      string
	Scores    []evaluation.CategoryScore
	WinnerIdx int
}

type comparisonRenderModel struct {
	Decks          []comparisonDeckRender
	Categories     []comparisonCategoryRender
	BestOverallIdx int
}

func getEvaluationCategories() []evaluationCategory {
	return []evaluationCategory{
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

func buildComparisonRenderModel(names []string, results []evaluation.EvaluationResult) comparisonRenderModel {
	decks := make([]comparisonDeckRender, len(results))
	for i, result := range results {
		decks[i] = comparisonDeckRender{
			Name:   names[i],
			Result: result,
		}
	}

	categoryDefs := getEvaluationCategories()
	categories := make([]comparisonCategoryRender, 0, len(categoryDefs))
	for _, category := range categoryDefs {
		categoryScores := make([]evaluation.CategoryScore, len(results))
		bestIdx := 0
		bestScore := -1.0
		for i, result := range results {
			score := category.get(result)
			categoryScores[i] = score
			if score.Score > bestScore {
				bestScore = score.Score
				bestIdx = i
			}
		}

		categories = append(categories, comparisonCategoryRender{
			Name:      category.name,
			Scores:    categoryScores,
			WinnerIdx: bestIdx,
		})
	}

	return comparisonRenderModel{
		Decks:          decks,
		Categories:     categories,
		BestOverallIdx: findBestOverallDeck(results),
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
