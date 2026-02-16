package main

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/fuzzstorage"
)

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
	for i := range 3 {
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

func TestLimitArchetypeRepetition(t *testing.T) {
	input := []FuzzingResult{
		{Deck: []string{"1"}, Archetype: "cycle"},
		{Deck: []string{"2"}, Archetype: "cycle"},
		{Deck: []string{"3"}, Archetype: "beatdown"},
		{Deck: []string{"4"}, Archetype: "cycle"},
		{Deck: []string{"5"}, Archetype: "beatdown"},
	}

	decks := make([]fuzzstorage.DeckEntry, 0, len(input))
	for _, r := range input {
		decks = append(decks, fuzzstorage.DeckEntry{
			Cards:     r.Deck,
			Archetype: r.Archetype,
		})
	}

	filtered := limitArchetypeRepetition(decks, 1)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 decks after limiting archetypes, got %d", len(filtered))
	}

	counts := map[string]int{}
	for _, deck := range filtered {
		counts[deck.Archetype]++
	}
	if counts["cycle"] != 1 {
		t.Fatalf("expected 1 cycle deck, got %d", counts["cycle"])
	}
	if counts["beatdown"] != 1 {
		t.Fatalf("expected 1 beatdown deck, got %d", counts["beatdown"])
	}
}

func TestFormatScoreTransition(t *testing.T) {
	t.Run("without theoretical value", func(t *testing.T) {
		got := formatScoreTransition(nil, 123, 4.2, func(entry fuzzstorage.DeckEntry) float64 {
			return entry.OverallScore
		})
		if got != "4.20" {
			t.Fatalf("formatScoreTransition() = %q, want %q", got, "4.20")
		}
	})

	t.Run("with theoretical value", func(t *testing.T) {
		theoreticalByID := map[int]fuzzstorage.DeckEntry{
			123: {
				ID:           123,
				OverallScore: 9.0,
			},
		}
		got := formatScoreTransition(theoreticalByID, 123, 4.2, func(entry fuzzstorage.DeckEntry) float64 {
			return entry.OverallScore
		})
		if got != "9.00->4.20" {
			t.Fatalf("formatScoreTransition() = %q, want %q", got, "9.00->4.20")
		}
	})
}

func TestSelectGAFitnessEvaluator(t *testing.T) {
	t.Run("archetype-free mode uses composite fitness evaluator", func(t *testing.T) {
		evaluator, mode := selectGAFitnessEvaluator(false)
		if evaluator == nil {
			t.Fatal("expected archetype-free mode to return evaluator")
		}
		if mode != gaFitnessModeArchetypeFree {
			t.Fatalf("mode = %q, want %q", mode, gaFitnessModeArchetypeFree)
		}

		score, err := evaluator(testCompositeDeckCandidates())
		if err != nil {
			t.Fatalf("unexpected evaluator error: %v", err)
		}
		if score < 0 || score > 10 {
			t.Fatalf("score = %v, want in range [0,10]", score)
		}
	})

	t.Run("legacy mode uses built-in evaluator", func(t *testing.T) {
		evaluator, mode := selectGAFitnessEvaluator(true)
		if evaluator != nil {
			t.Fatal("expected legacy mode to use built-in evaluator")
		}
		if mode != gaFitnessModeLegacy {
			t.Fatalf("mode = %q, want %q", mode, gaFitnessModeLegacy)
		}
	})
}

func testCompositeDeckCandidates() []deck.CardCandidate {
	return []deck.CardCandidate{
		{Name: "Hog Rider", Elixir: 4, Role: ptrRole(deck.RoleWinCondition), Level: 11, MaxLevel: 14},
		{Name: "Fireball", Elixir: 4, Role: ptrRole(deck.RoleSpellBig), Level: 11, MaxLevel: 14},
		{Name: "Zap", Elixir: 2, Role: ptrRole(deck.RoleSpellSmall), Level: 11, MaxLevel: 14},
		{
			Name:   "Musketeer",
			Elixir: 4,
			Role:   ptrRole(deck.RoleSupport),
			Level:  11, MaxLevel: 14,
			Stats: &clashroyale.CombatStats{Targets: "Air & Ground", DamagePerSecond: 181},
		},
		{
			Name:   "Mini P.E.K.K.A",
			Elixir: 4,
			Role:   ptrRole(deck.RoleSupport),
			Level:  11, MaxLevel: 14,
			Stats: &clashroyale.CombatStats{Targets: "Ground", DamagePerSecond: 325},
		},
		{
			Name:   "Valkyrie",
			Elixir: 4,
			Role:   ptrRole(deck.RoleSupport),
			Level:  11, MaxLevel: 14,
			Stats: &clashroyale.CombatStats{Targets: "Ground", Radius: 1.2},
		},
		{Name: "Skeletons", Elixir: 1, Role: ptrRole(deck.RoleCycle), Level: 11, MaxLevel: 14},
		{
			Name:   "Archers",
			Elixir: 3,
			Role:   ptrRole(deck.RoleSupport),
			Level:  11, MaxLevel: 14,
			Stats: &clashroyale.CombatStats{Targets: "Air & Ground", DamagePerSecond: 108},
		},
	}
}

func ptrRole(role deck.CardRole) *deck.CardRole {
	return &role
}
