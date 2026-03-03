package main

import (
	"strings"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

func TestSortEvaluationResultsUsesTypedAdapters(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		verifyTop func(*testing.T, string)
	}{
		{
			name: "eval batch results",
			verifyTop: func(t *testing.T, sortBy string) {
				t.Helper()
				results := testEvalBatchResults()
				sortEvaluationResults(results, sortBy)
				if got := results[0].Name; got != "Fast Deck" {
					t.Fatalf("sortEvaluationResults(%s) first = %q, want %q", sortBy, got, "Fast Deck")
				}
			},
		},
		{
			name: "compare batch results",
			verifyTop: func(t *testing.T, sortBy string) {
				t.Helper()
				results := testCompareBatchResults()
				sortEvaluationResults(results, sortBy)
				if got := results[0].Name; got != "Fast Deck" {
					t.Fatalf("sortEvaluationResults(%s) first = %q, want %q", sortBy, got, "Fast Deck")
				}
			},
		},
		{
			name: "suite batch results",
			verifyTop: func(t *testing.T, sortBy string) {
				t.Helper()
				results := testSuiteBatchResults()
				sortEvaluationResults(results, sortBy)
				if got := results[0].Name; got != "Fast Deck" {
					t.Fatalf("sortEvaluationResults(%s) first = %q, want %q", sortBy, got, "Fast Deck")
				}
			},
		},
		{
			name: "detailed formatter path",
			verifyTop: func(t *testing.T, sortBy string) {
				t.Helper()
				results := testEvalBatchResults()
				sortEvaluationResults(results, sortBy)

				got := formatEvaluationBatchDetailed(results, "Player One", "#TAG")
				firstIdx := strings.Index(got, "DECK #1: Fast Deck")
				secondIdx := strings.Index(got, "DECK #2: Heavy Deck")
				if firstIdx == -1 || secondIdx == -1 {
					t.Fatalf("detailed output missing expected deck headings\n%s", got)
				}
				if firstIdx > secondIdx {
					t.Fatalf("detailed output order incorrect for %s\n%s", sortBy, got)
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		for _, sortBy := range []string{"overall", "elixir"} {
			sortBy := sortBy
			t.Run(tc.name+"/"+sortBy, func(t *testing.T) {
				t.Parallel()
				tc.verifyTop(t, sortBy)
			})
		}
	}
}

func TestFormatEvaluationBatchSummaryOutput(t *testing.T) {
	results := []evalBatchResult{
		{
			Name:     "Fast Cycle Pressure Deck Name",
			Strategy: "cycle",
			Deck:     []string{"Hog Rider", "Ice Spirit"},
			Result:   testBatchEvaluationResult(8.4, 7.1, 6.8, 7.9, 7.6, 8.2, 2.8, evaluation.ArchetypeCycle),
		},
		{
			Name:     "Heavy Beatdown",
			Strategy: "beatdown",
			Deck:     []string{"Giant", "Night Witch"},
			Result:   testBatchEvaluationResult(6.2, 8.0, 7.5, 7.1, 6.8, 5.9, 3.8, evaluation.ArchetypeBeatdown),
		},
	}

	got := formatEvaluationBatchSummary(results, 2, 2*time.Second, "overall", "Player One", "#TAG")

	wantLines := []string{
		"Player: Player One (#TAG)",
		"Total Decks: 2 | Evaluated: 2 | Sorted by: overall",
		"Total Time: 2s | Avg: 1s",
		"│   1 │ Fast Cycle Pressure Deck ... │   8.40  │   7.10 │   6.80 │   7.90 │ cycle        │",
		"│   2 │ Heavy Beatdown               │   6.20  │   8.00 │   7.50 │   7.10 │ beatdown     │",
	}

	for _, want := range wantLines {
		if !strings.Contains(got, want) {
			t.Fatalf("summary output missing line %q\nfull output:\n%s", want, got)
		}
	}
}

func TestFormatEvaluationBatchCSVOutput(t *testing.T) {
	t.Parallel()

	results := []evalBatchResult{
		{
			Name:     "Fast Deck",
			Strategy: "cycle",
			Deck:     []string{"Hog Rider", "Ice Spirit", "The Log"},
			Result:   testBatchEvaluationResult(8.4, 7.1, 6.8, 7.9, 7.6, 8.2, 2.8, evaluation.ArchetypeCycle),
		},
	}

	got := formatEvaluationBatchCSV(results)
	want := strings.Join([]string{
		"Rank,Name,Strategy,Overall,Attack,Defense,Synergy,Versatility,F2P,Playability,Archetype,Avg_Elixir,Deck",
		"1,Fast Deck,cycle,8.40,7.10,6.80,7.90,7.60,8.20,6.50,cycle,2.80,\"Hog Rider - Ice Spirit - The Log\"",
		"",
	}, "\n")

	if got != want {
		t.Fatalf("formatEvaluationBatchCSV mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestFormatEvaluationBatchSkipsUnnamedRowsWithoutRankGaps(t *testing.T) {
	t.Parallel()

	results := []evalBatchResult{
		{
			Name:     "Fast Deck",
			Strategy: "cycle",
			Deck:     []string{"Hog Rider"},
			Result:   testBatchEvaluationResult(8.4, 7.1, 6.8, 7.9, 7.6, 8.2, 2.8, evaluation.ArchetypeCycle),
		},
		{
			Name:     "",
			Strategy: "skip",
			Deck:     []string{"Knight"},
			Result:   testBatchEvaluationResult(7.4, 7.0, 7.0, 7.0, 7.0, 7.0, 3.2, evaluation.ArchetypeControl),
		},
		{
			Name:     "Heavy Deck",
			Strategy: "beatdown",
			Deck:     []string{"Giant"},
			Result:   testBatchEvaluationResult(6.2, 8.0, 7.5, 7.1, 6.8, 5.9, 3.8, evaluation.ArchetypeBeatdown),
		},
	}

	summary := formatEvaluationBatchSummary(results, 3, 3*time.Second, "overall", "Player One", "#TAG")
	if !strings.Contains(summary, "│   2 │ Heavy Deck") {
		t.Fatalf("summary output should use contiguous ranks\n%s", summary)
	}

	csvOutput := formatEvaluationBatchCSV(results)
	if !strings.Contains(csvOutput, "\n2,Heavy Deck,beatdown,") {
		t.Fatalf("csv output should use contiguous ranks\n%s", csvOutput)
	}

	detailed := formatEvaluationBatchDetailed(results, "Player One", "#TAG")
	if !strings.Contains(detailed, "DECK #2: Heavy Deck") {
		t.Fatalf("detailed output should use contiguous ranks\n%s", detailed)
	}
}

func testEvalBatchResults() []evalBatchResult {
	return []evalBatchResult{
		{
			Name:     "Heavy Deck",
			Strategy: "beatdown",
			Deck:     []string{"Giant"},
			Result:   testBatchEvaluationResult(6.2, 8.0, 7.5, 7.1, 6.8, 5.9, 3.8, evaluation.ArchetypeBeatdown),
		},
		{
			Name:     "Fast Deck",
			Strategy: "cycle",
			Deck:     []string{"Hog Rider"},
			Result:   testBatchEvaluationResult(8.4, 7.1, 6.8, 7.9, 7.6, 8.2, 2.8, evaluation.ArchetypeCycle),
		},
	}
}

func testCompareBatchResults() []batchEvalResult {
	return []batchEvalResult{
		{
			Name:     "Heavy Deck",
			Strategy: "beatdown",
			Deck:     []string{"Giant"},
			Result:   testBatchEvaluationResult(6.2, 8.0, 7.5, 7.1, 6.8, 5.9, 3.8, evaluation.ArchetypeBeatdown),
		},
		{
			Name:     "Fast Deck",
			Strategy: "cycle",
			Deck:     []string{"Hog Rider"},
			Result:   testBatchEvaluationResult(8.4, 7.1, 6.8, 7.9, 7.6, 8.2, 2.8, evaluation.ArchetypeCycle),
		},
	}
}

func testSuiteBatchResults() []suiteEvalResult {
	return []suiteEvalResult{
		{
			Name:     "Heavy Deck",
			Strategy: "beatdown",
			Deck:     []string{"Giant"},
			Result:   testBatchEvaluationResult(6.2, 8.0, 7.5, 7.1, 6.8, 5.9, 3.8, evaluation.ArchetypeBeatdown),
		},
		{
			Name:     "Fast Deck",
			Strategy: "cycle",
			Deck:     []string{"Hog Rider"},
			Result:   testBatchEvaluationResult(8.4, 7.1, 6.8, 7.9, 7.6, 8.2, 2.8, evaluation.ArchetypeCycle),
		},
	}
}

func testBatchEvaluationResult(overall, attack, defense, synergy, versatility, f2p, elixir float64, archetype evaluation.Archetype) evaluation.EvaluationResult {
	return evaluation.EvaluationResult{
		OverallScore:      overall,
		Attack:            evaluation.CategoryScore{Score: attack},
		Defense:           evaluation.CategoryScore{Score: defense},
		Synergy:           evaluation.CategoryScore{Score: synergy},
		Versatility:       evaluation.CategoryScore{Score: versatility},
		F2PFriendly:       evaluation.CategoryScore{Score: f2p},
		Playability:       evaluation.CategoryScore{Score: 6.5},
		AvgElixir:         elixir,
		DetectedArchetype: archetype,
	}
}
