package research

import (
	"fmt"
	"sort"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// PlayerInput holds runtime player context for the benchmark.
type PlayerInput struct {
	Tag        string
	Name       string
	Candidates []deck.CardCandidate
}

// BenchmarkRunner executes method comparisons.
type BenchmarkRunner struct {
	Builder *deck.Builder
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}

// Run executes all configured methods across all players.
//
//nolint:gocyclo,funlen // Benchmark pipeline is intentionally linear and explicit.
func (r *BenchmarkRunner) Run(config BenchmarkConfig, players []PlayerInput) (*BenchmarkReport, error) {
	if len(players) == 0 {
		return nil, fmt.Errorf("no players provided")
	}
	report := &BenchmarkReport{
		Timestamp:  time.Now().UTC(),
		Seed:       config.Seed,
		Methods:    config.Methods,
		PlayerRuns: make([]PlayerResult, 0, len(players)),
	}
	constraints, err := resolveConstraintConfig(config.Constraints)
	if err != nil {
		return nil, fmt.Errorf("invalid benchmark constraints: %w", err)
	}

	aggScores := make(map[string][]float64)
	aggRuntime := make(map[string][]float64)
	aggViolations := make(map[string]int)

	for idx, p := range players {
		playerRes := PlayerResult{
			PlayerTag:  p.Tag,
			PlayerName: p.Name,
			MethodRuns: make([]DeckResult, 0, len(config.Methods)),
		}
		for _, methodName := range config.Methods {
			method, err := methodByName(methodName, r.Builder)
			if err != nil {
				return nil, err
			}
			res, err := method.Build(p.Candidates, MethodConfig{
				Seed:        config.Seed + int64(idx),
				TopN:        config.TopN,
				DataDir:     config.DataDir,
				Constraints: &constraints,
			})
			if err != nil {
				return nil, fmt.Errorf("%s failed for %s: %w", methodName, p.Tag, err)
			}
			playerRes.MethodRuns = append(playerRes.MethodRuns, res)
			aggScores[res.Method] = append(aggScores[res.Method], res.Metrics.Composite)
			aggRuntime[res.Method] = append(aggRuntime[res.Method], float64(res.Metrics.RuntimeMs))
			aggViolations[res.Method] += len(res.Metrics.ConstraintViolations)
		}

		sort.Slice(playerRes.MethodRuns, func(i, j int) bool {
			return playerRes.MethodRuns[i].Metrics.Composite > playerRes.MethodRuns[j].Metrics.Composite
		})
		if len(playerRes.MethodRuns) > 0 {
			playerRes.Winner = playerRes.MethodRuns[0].Method
			playerRes.WinnerScore = playerRes.MethodRuns[0].Metrics.Composite
		}
		report.PlayerRuns = append(report.PlayerRuns, playerRes)
	}

	aggs := make([]BenchmarkAggregate, 0, len(config.Methods))
	for _, method := range config.Methods {
		scores := aggScores[method]
		runtimes := aggRuntime[method]
		if len(scores) == 0 {
			continue
		}
		totalScore := 0.0
		totalRuntime := 0.0
		for _, s := range scores {
			totalScore += s
		}
		for _, rt := range runtimes {
			totalRuntime += rt
		}
		aggs = append(aggs, BenchmarkAggregate{
			Method:                   method,
			Runs:                     len(scores),
			MeanComposite:            totalScore / float64(len(scores)),
			MedianComposite:          median(scores),
			MeanRuntimeMs:            totalRuntime / float64(len(runtimes)),
			ConstraintViolationCount: aggViolations[method],
		})
	}
	sort.Slice(aggs, func(i, j int) bool { return aggs[i].MeanComposite > aggs[j].MeanComposite })
	report.Aggregates = aggs
	report.Recommendations = recommendNextSteps(aggs)

	return report, nil
}

func recommendNextSteps(aggs []BenchmarkAggregate) []string {
	if len(aggs) == 0 {
		return []string{"No method results available."}
	}
	recs := []string{
		fmt.Sprintf("Top method by mean composite: %s", aggs[0].Method),
	}
	if len(aggs) > 1 {
		recs = append(recs, fmt.Sprintf("Second-best method: %s", aggs[1].Method))
	}
	for _, a := range aggs {
		if a.ConstraintViolationCount > 0 {
			recs = append(recs, fmt.Sprintf("Method %s needs constraint hardening (%d violations).", a.Method, a.ConstraintViolationCount))
		}
	}
	recs = append(recs, "Recommended next method to implement from remaining set: Counter-centric (lowest implementation risk with current coverage primitives).")
	return recs
}
