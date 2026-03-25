package main

import (
	"math"
	"sort"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

type comparisonDeckView struct {
	Name            string
	TruncatedName20 string
	TruncatedName18 string
	Result          evaluation.EvaluationResult
	OverallStars    string
	PredictedWinPct float64
}

func estimateWinRateFromScore(score float64) float64 {
	estimated := 50.0 + (score-5.0)*4.0
	return math.Max(30.0, math.Min(70.0, estimated))
}

type comparisonCategoryView struct {
	Name          string
	Scores        []evaluation.CategoryScore
	BestDeckIndex int
}

type comparisonViewModel struct {
	Decks            []comparisonDeckView
	Categories       []comparisonCategoryView
	BestOverallIndex int
	RankedDecks      []int
}

func buildComparisonViewModel(names []string, results []evaluation.EvaluationResult) comparisonViewModel {
	vm := comparisonViewModel{
		Decks:       make([]comparisonDeckView, len(names)),
		Categories:  make([]comparisonCategoryView, 0, len(getEvaluationCategories())),
		RankedDecks: make([]int, len(results)),
	}

	bestOverallIdx := 0
	bestOverallScore := -1.0
	for i := range names {
		r := results[i]
		vm.Decks[i] = comparisonDeckView{
			Name:            names[i],
			TruncatedName20: truncate(names[i], 20),
			TruncatedName18: truncate(names[i], 18),
			Result:          r,
			OverallStars:    formatStarsDisplay(calculateStars(r.OverallScore)),
			PredictedWinPct: estimateWinRateFromScore(r.OverallScore),
		}

		vm.RankedDecks[i] = i
		if r.OverallScore > bestOverallScore {
			bestOverallScore = r.OverallScore
			bestOverallIdx = i
		}
	}
	vm.BestOverallIndex = bestOverallIdx

	for _, category := range getEvaluationCategories() {
		view := comparisonCategoryView{
			Name:          category.name,
			Scores:        make([]evaluation.CategoryScore, len(results)),
			BestDeckIndex: 0,
		}
		bestCategoryScore := -1.0
		for i, result := range results {
			score := category.get(result)
			view.Scores[i] = score
			if score.Score > bestCategoryScore {
				bestCategoryScore = score.Score
				view.BestDeckIndex = i
			}
		}
		vm.Categories = append(vm.Categories, view)
	}

	sort.Slice(vm.RankedDecks, func(i, j int) bool {
		return vm.Decks[vm.RankedDecks[i]].Result.OverallScore > vm.Decks[vm.RankedDecks[j]].Result.OverallScore
	})

	return vm
}
