package main

import "testing"

func TestGetElixirBucket(t *testing.T) {
	testCases := []struct {
		name      string
		avgElixir float64
		expected  string
	}{
		{name: "low below threshold", avgElixir: 3.29, expected: elixirBucketLow},
		{name: "medium at low threshold", avgElixir: 3.3, expected: elixirBucketMedium},
		{name: "medium upper boundary", avgElixir: 4.0, expected: elixirBucketMedium},
		{name: "high above medium threshold", avgElixir: 4.01, expected: elixirBucketHigh},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := getElixirBucket(tc.avgElixir)
			if got != tc.expected {
				t.Fatalf("getElixirBucket(%.2f) = %q, want %q", tc.avgElixir, got, tc.expected)
			}
		})
	}
}

func TestEnsureElixirBucketDistribution_CoversTopBuckets(t *testing.T) {
	results := []FuzzingResult{
		{Deck: []string{"L1"}, OverallScore: 9.9, AvgElixir: 2.8}, // low
		{Deck: []string{"L2"}, OverallScore: 9.8, AvgElixir: 3.1}, // low
		{Deck: []string{"M1"}, OverallScore: 9.7, AvgElixir: 3.5}, // medium
		{Deck: []string{"M2"}, OverallScore: 9.6, AvgElixir: 3.8}, // medium
		{Deck: []string{"H1"}, OverallScore: 9.5, AvgElixir: 4.4}, // high
	}

	reordered := ensureElixirBucketDistribution(results, 3, false)
	if len(reordered) != len(results) {
		t.Fatalf("reordered length = %d, want %d", len(reordered), len(results))
	}

	topBuckets := map[string]bool{}
	for i := 0; i < 3; i++ {
		topBuckets[getElixirBucket(reordered[i].AvgElixir)] = true
	}

	if !topBuckets[elixirBucketLow] || !topBuckets[elixirBucketMedium] || !topBuckets[elixirBucketHigh] {
		t.Fatalf("top 3 should include low/medium/high buckets, got %#v", topBuckets)
	}
}

func TestEnsureElixirBucketDistribution_MissingBucketFallsBackToScoreOrder(t *testing.T) {
	results := []FuzzingResult{
		{Deck: []string{"M1"}, OverallScore: 9.9, AvgElixir: 3.6}, // medium
		{Deck: []string{"L1"}, OverallScore: 9.8, AvgElixir: 2.9}, // low
		{Deck: []string{"M2"}, OverallScore: 9.7, AvgElixir: 3.7}, // medium
		{Deck: []string{"L2"}, OverallScore: 9.6, AvgElixir: 3.0}, // low
	}

	reordered := ensureElixirBucketDistribution(results, 4, false)
	if len(reordered) != len(results) {
		t.Fatalf("reordered length = %d, want %d", len(reordered), len(results))
	}

	if reordered[0].Deck[0] != "L1" {
		t.Fatalf("first result should come from low bucket, got %s", reordered[0].Deck[0])
	}
	if reordered[1].Deck[0] != "M1" {
		t.Fatalf("second result should come from medium bucket, got %s", reordered[1].Deck[0])
	}
}
