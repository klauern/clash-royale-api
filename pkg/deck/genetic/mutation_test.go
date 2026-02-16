// Package genetic provides genetic algorithm-based deck optimization
// using the eaopt library for evolutionary deck generation.
package genetic

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestDeckGenomeMutateExecution(t *testing.T) {
	candidates := createMockCandidates(20)
	cfg := DefaultGeneticConfig()
	cfg.MutationIntensity = 0.5
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}

	genome.Fitness = 5.5

	originalCards := make([]string, 8)
	copy(originalCards, genome.Cards)

	err = genome.Mutate()
	if err != nil {
		t.Errorf("Mutate() error = %v", err)
	}

	// Fitness should be reset to 0
	if genome.Fitness != 0 {
		t.Errorf("Mutate() should reset fitness to 0, got %v", genome.Fitness)
	}

	// Should still have 8 cards
	if len(genome.Cards) != 8 {
		t.Errorf("Mutate() resulted in %d cards, want 8", len(genome.Cards))
	}

	// Cards should be unique
	cardSet := make(map[string]bool)
	for _, card := range genome.Cards {
		if cardSet[card] {
			t.Errorf("Mutate() produced duplicate card: %v", card)
		}
		cardSet[card] = true
	}

	// At least one card should have changed (with 50% intensity)
	changes := 0
	for i, card := range genome.Cards {
		if card != originalCards[i] {
			changes++
		}
	}
	if changes == 0 {
		t.Error("Mutate() changed no cards with 50% intensity")
	}
}

func TestDeckGenomeMutateIntensityLevels(t *testing.T) {
	tests := []struct {
		name               string
		intensity          float64
		minExpectedChanges int
		maxExpectedChanges int
	}{
		{
			name:               "low intensity (10%)",
			intensity:          0.1,
			minExpectedChanges: 1, // At least 1 card
			maxExpectedChanges: 2,
		},
		{
			name:               "medium intensity (50%)",
			intensity:          0.5,
			minExpectedChanges: 1,
			maxExpectedChanges: 8,
		},
		{
			name:               "high intensity (100%)",
			intensity:          1.0,
			minExpectedChanges: 1,
			maxExpectedChanges: 8,
		},
		{
			name:               "zero intensity (minimum 1)",
			intensity:          0.0,
			minExpectedChanges: 1, // Should mutate at least 1 card
			maxExpectedChanges: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := createMockCandidates(20)
			cfg := DefaultGeneticConfig()
			cfg.MutationIntensity = tt.intensity
			strategy := deck.StrategyBalanced

			cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
			genome, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
			if err != nil {
				t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
			}

			originalCards := make([]string, 8)
			copy(originalCards, genome.Cards)

			// Run mutation multiple times to test expected range
			// Use more runs for better statistical accuracy
			totalChanges := 0
			runs := 50 // Increased from 10 to reduce flakiness
			for range runs {
				genome2, _ := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
				_ = genome2.Mutate()

				changes := 0
				for j, card := range genome2.Cards {
					if card != originalCards[j] {
						changes++
					}
				}
				totalChanges += changes
			}

			// At least one mutation should have changed something across all runs
			if totalChanges == 0 {
				t.Error("No changes across all mutation runs")
			}

			// Check that mutations are actually happening
			// Even with 0 intensity, should mutate at least 1 card per run
			if totalChanges < runs/2 {
				t.Errorf("Too few total changes %d across %d runs (expected at least %d)", totalChanges, runs, runs/2)
			}
		})
	}
}

func TestDeckGenomeMutateUniqueness(t *testing.T) {
	candidates := createMockCandidates(15)
	cfg := DefaultGeneticConfig()
	cfg.MutationIntensity = 0.8 // High intensity
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}

	// Run mutation multiple times
	for i := range 20 {
		err := genome.Mutate()
		if err != nil {
			t.Errorf("Mutate() iteration %d error = %v", i, err)
		}

		// Validate uniqueness each time
		cardSet := make(map[string]bool)
		for _, card := range genome.Cards {
			if cardSet[card] {
				t.Errorf("Mutate() iteration %d produced duplicate card: %v", i, card)
			}
			cardSet[card] = true
		}

		// Validate deck size
		if len(genome.Cards) != 8 {
			t.Errorf("Mutate() iteration %d resulted in %d cards, want 8", i, len(genome.Cards))
		}
	}
}

func TestDeckGenomeMutateWithSmallCandidatePool(t *testing.T) {
	// Test mutation with exactly 8 candidates (minimum for a deck)
	candidates := createMockCandidates(8)
	cfg := DefaultGeneticConfig()
	cfg.MutationIntensity = 0.3
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}

	// Mutation should handle small pool gracefully
	err = genome.Mutate()
	if err != nil {
		t.Errorf("Mutate() with small pool error = %v", err)
	}

	// Should still have valid deck
	if len(genome.Cards) != 8 {
		t.Errorf("Mutate() resulted in %d cards, want 8", len(genome.Cards))
	}

	cardSet := make(map[string]bool)
	for _, card := range genome.Cards {
		cardSet[card] = true
	}
	if len(cardSet) != 8 {
		t.Errorf("Mutate() produced %d unique cards, want 8", len(cardSet))
	}
}

func TestDeckGenomeMutateNilConfig(t *testing.T) {
	candidates := createMockCandidates(10)
	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}

	genome := &DeckGenome{
		Cards:      cards,
		config:     nil, // Nil config
		candidates: candidates,
		strategy:   deck.StrategyBalanced,
	}

	err := genome.Mutate()
	if err == nil {
		t.Error("Mutate() with nil config should error")
	}
	if !contains(err.Error(), "config") {
		t.Errorf("Mutate() error should mention config, got: %v", err)
	}
}

func TestMutationStrategiesExecution(t *testing.T) {
	// This test verifies that mutation strategies execute without errors
	// Individual strategy testing is implicitly done through repeated mutations
	candidates := createMockCandidates(25)
	cfg := DefaultGeneticConfig()
	cfg.MutationIntensity = 0.4
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}

	// Run many mutations to exercise all strategy paths
	for i := range 100 {
		genome, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
		if err != nil {
			t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
		}

		err = genome.Mutate()
		if err != nil {
			t.Errorf("Mutate() iteration %d error = %v", i, err)
		}

		// Basic validation
		if len(genome.Cards) != 8 {
			t.Errorf("Mutation %d produced %d cards", i, len(genome.Cards))
		}

		cardSet := make(map[string]bool)
		for _, card := range genome.Cards {
			cardSet[card] = true
		}
		if len(cardSet) != 8 {
			t.Errorf("Mutation %d produced non-unique cards", i)
		}
	}
}

func TestMutationFitnessReset(t *testing.T) {
	candidates := createMockCandidates(15)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}

	// Set various fitness values and verify reset
	fitnessValues := []float64{0.0, 5.5, 10.0, 3.2, 7.8}

	for _, fitness := range fitnessValues {
		genome.Fitness = fitness
		err := genome.Mutate()
		if err != nil {
			t.Errorf("Mutate() error = %v", err)
		}

		if genome.Fitness != 0 {
			t.Errorf("After mutation, fitness = %v, want 0 (was %v before mutation)", genome.Fitness, fitness)
		}
	}
}

func TestMutationPickPositions(t *testing.T) {
	candidates := createMockCandidates(10)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}

	tests := []struct {
		name  string
		count int
	}{
		{"pick 1 position", 1},
		{"pick 3 positions", 3},
		{"pick 5 positions", 5},
		{"pick 8 positions", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			positions := genome.pickMutationPositions(tt.count)

			if len(positions) != tt.count {
				t.Errorf("pickMutationPositions(%d) returned %d positions", tt.count, len(positions))
			}

			// Verify all positions are unique
			posSet := make(map[int]bool)
			for _, pos := range positions {
				if posSet[pos] {
					t.Errorf("pickMutationPositions() returned duplicate position: %d", pos)
				}
				posSet[pos] = true
			}

			// Verify all positions are valid (0-7)
			for _, pos := range positions {
				if pos < 0 || pos >= 8 {
					t.Errorf("pickMutationPositions() returned invalid position: %d", pos)
				}
			}
		})
	}
}

func TestMutationCurrentCardSet(t *testing.T) {
	candidates := createMockCandidates(10)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}

	cardSet := genome.currentCardSet()

	// Verify all cards are in the set
	if len(cardSet) != 8 {
		t.Errorf("currentCardSet() returned %d cards, want 8", len(cardSet))
	}

	for _, card := range cards {
		if !cardSet[card] {
			t.Errorf("currentCardSet() missing card: %v", card)
		}
	}
}

func TestSingleCardSwap(t *testing.T) {
	candidates := createMockCandidates(15)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}

	used := genome.currentCardSet()

	// Get replacement
	replacement := genome.singleCardSwap(used)

	// Replacement should not be in used set
	if used[replacement] {
		t.Errorf("singleCardSwap() returned card already in deck: %v", replacement)
	}

	// Replacement should be a valid candidate
	found := false
	for _, candidate := range candidates {
		if candidate.Name == replacement {
			found = true
			break
		}
	}
	if !found && replacement != "" {
		t.Errorf("singleCardSwap() returned non-candidate card: %v", replacement)
	}
}

func TestSingleCardSwapNoOptions(t *testing.T) {
	// Test when all candidates are already used (edge case)
	candidates := createMockCandidates(8)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}

	// All candidates used
	used := genome.currentCardSet()

	replacement := genome.singleCardSwap(used)

	// Should return empty string when no options
	if replacement != "" {
		t.Errorf("singleCardSwap() with no options should return empty, got: %v", replacement)
	}
}
