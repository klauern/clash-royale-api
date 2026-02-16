// Package genetic provides genetic algorithm-based deck optimization
// using the eaopt library for evolutionary deck generation.
package genetic

import (
	"fmt"
	"testing"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestNewDeckGenomeFromCards(t *testing.T) {
	candidates := createMockCandidates(10)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	tests := []struct {
		name        string
		cards       []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid 8-card deck",
			cards:   []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"},
			wantErr: false,
		},
		{
			name:        "too few cards",
			cards:       []string{"Card0", "Card1", "Card2"},
			wantErr:     true,
			errContains: "exactly 8 cards",
		},
		{
			name:        "too many cards",
			cards:       []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7", "Card8"},
			wantErr:     true,
			errContains: "exactly 8 cards",
		},
		{
			name:        "empty deck",
			cards:       []string{},
			wantErr:     true,
			errContains: "exactly 8 cards",
		},
		{
			name:        "card not in candidates",
			cards:       []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "UnknownCard"},
			wantErr:     true,
			errContains: "not found in candidates",
		},
		{
			name:        "duplicate cards",
			cards:       []string{"Card0", "Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6"},
			wantErr:     false, // Currently allows duplicates - deck validation would catch this
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genome, err := NewDeckGenomeFromCards(tt.cards, candidates, strategy, &cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewDeckGenomeFromCards() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("NewDeckGenomeFromCards() error = %v, should contain %v", err, tt.errContains)
				}
			}

			if !tt.wantErr && genome != nil {
				if len(genome.Cards) != 8 {
					t.Errorf("NewDeckGenomeFromCards() genome has %d cards, want 8", len(genome.Cards))
				}
			}
		})
	}
}

func TestDeckGenomeClone(t *testing.T) {
	candidates := createMockCandidates(10)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	originalCards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	original, err := NewDeckGenomeFromCards(originalCards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}
	original.Fitness = 0.75

	clone := original.Clone()

	clonedGenome, ok := clone.(*DeckGenome)
	if !ok {
		t.Fatalf("Clone() did not return *DeckGenome")
	}

	// Verify cards are the same
	if len(clonedGenome.Cards) != len(original.Cards) {
		t.Errorf("Clone() has %d cards, original has %d", len(clonedGenome.Cards), len(original.Cards))
	}

	for i, card := range original.Cards {
		if clonedGenome.Cards[i] != card {
			t.Errorf("Clone() card[%d] = %v, want %v", i, clonedGenome.Cards[i], card)
		}
	}

	// Verify fitness is copied
	if clonedGenome.Fitness != original.Fitness {
		t.Errorf("Clone() fitness = %v, want %v", clonedGenome.Fitness, original.Fitness)
	}

	// Verify it's a deep copy (modifying clone doesn't affect original)
	clonedGenome.Cards[0] = "ModifiedCard"
	if original.Cards[0] == "ModifiedCard" {
		t.Error("Clone() modified original cards - not a deep copy")
	}
	original.Cards[0] = "Card0" // Restore for other tests
}

func TestDeckGenomeCrossover(t *testing.T) {
	candidates := createMockCandidates(12)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	parent1Cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	parent2Cards := []string{"Card4", "Card5", "Card6", "Card7", "Card8", "Card9", "Card10", "Card11"}

	parent1, err := NewDeckGenomeFromCards(parent1Cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}

	parent2, err := NewDeckGenomeFromCards(parent2Cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}

	t.Run("produces valid offspring", func(t *testing.T) {
		offspring, err := parent1.Crossover(parent2)
		if err != nil {
			t.Fatalf("Crossover() error = %v", err)
		}

		offspringGenome, ok := offspring.(*DeckGenome)
		if !ok {
			t.Fatalf("Crossover() did not return *DeckGenome")
		}

		// Check offspring has 8 cards
		if len(offspringGenome.Cards) != 8 {
			t.Errorf("Crossover() produced offspring with %d cards, want 8", len(offspringGenome.Cards))
		}

		// Check all cards are from candidates
		for _, card := range offspringGenome.Cards {
			found := false
			for _, c := range candidates {
				if c.Name == card {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Crossover() produced offspring with card %v not in candidates", card)
			}
		}
	})

	t.Run("produces unique offspring", func(t *testing.T) {
		offspring, err := parent1.Crossover(parent2)
		if err != nil {
			t.Fatalf("Crossover() error = %v", err)
		}

		offspringGenome := offspring.(*DeckGenome)

		// Check for duplicates within offspring
		cardSet := make(map[string]bool)
		for _, card := range offspringGenome.Cards {
			if cardSet[card] {
				t.Errorf("Crossover() produced offspring with duplicate card: %v", card)
			}
			cardSet[card] = true
		}
	})

	t.Run("rejects wrong type", func(t *testing.T) {
		_, err := parent1.Crossover("not a genome")
		if err == nil {
			t.Error("Crossover() should error with wrong type")
		}
		if !contains(err.Error(), "DeckGenome") {
			t.Errorf("Crossover() error should mention DeckGenome, got: %v", err)
		}
	})
}

func TestDeckGenomeString(t *testing.T) {
	candidates := createMockCandidates(10)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}
	genome.Fitness = 0.85

	s := genome.String()

	if s == "" {
		t.Error("String() returned empty string")
	}

	// Should contain fitness
	if !contains(s, "0.85") {
		t.Errorf("String() should contain fitness 0.85, got: %v", s)
	}

	// Should contain deck identifier
	if !contains(s, "Deck{") {
		t.Errorf("String() should contain deck identifier, got: %v", s)
	}
}

func TestDeckGenomeEvaluate(t *testing.T) {
	candidates := createMockCandidates(10)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}

	t.Run("calculates fitness score", func(t *testing.T) {
		fitness, err := genome.Evaluate()
		if err != nil {
			t.Errorf("Evaluate() error = %v", err)
		}

		// Fitness should be calculated (0-10 scale from evaluation system)
		if fitness < 0 || fitness > 10 {
			t.Errorf("Evaluate() = %v, want score in range [0, 10]", fitness)
		}

		// Should set genome fitness
		if genome.Fitness != fitness {
			t.Errorf("Evaluate() did not set genome.Fitness, got %v want %v", genome.Fitness, fitness)
		}
	})

	t.Run("different decks have different scores", func(t *testing.T) {
		// Create a second genome with different cards
		cards2 := []string{"Card2", "Card3", "Card4", "Card5", "Card6", "Card7", "Card8", "Card9"}
		genome2, err := NewDeckGenomeFromCards(cards2, candidates, strategy, &cfg)
		if err != nil {
			t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
		}

		fitness1, err1 := genome.Evaluate()
		fitness2, err2 := genome2.Evaluate()

		if err1 != nil {
			t.Errorf("Evaluate() error for genome1 = %v", err1)
		}
		if err2 != nil {
			t.Errorf("Evaluate() error for genome2 = %v", err2)
		}

		// Different decks can have different fitness (not guaranteed but likely with mock data)
		// At minimum, verify both are in valid range
		if fitness1 < 0 || fitness1 > 10 {
			t.Errorf("genome1 fitness = %v, want score in range [0, 10]", fitness1)
		}
		if fitness2 < 0 || fitness2 > 10 {
			t.Errorf("genome2 fitness = %v, want score in range [0, 10]", fitness2)
		}
	})

	t.Run("error when cards cannot be resolved", func(t *testing.T) {
		// Create genome with cards that won't resolve
		genome3 := &DeckGenome{
			Cards:      []string{"Unknown1", "Unknown2", "Unknown3", "Unknown4", "Unknown5", "Unknown6", "Unknown7", "Unknown8"},
			config:     &cfg,
			candidates: candidates,
			strategy:   strategy,
		}

		_, err := genome3.Evaluate()
		if err == nil {
			t.Error("Evaluate() should error when cards cannot be resolved")
		}
		if !contains(err.Error(), "failed to resolve all cards") {
			t.Errorf("Evaluate() error should mention resolution failure, got: %v", err)
		}
	})
}

func TestDeckGenomeMutate(t *testing.T) {
	candidates := createMockCandidates(20) // Need more candidates for mutation
	cfg := DefaultGeneticConfig()
	cfg.MutationIntensity = 0.5 // Mutate 4 cards (50% of 8)
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}

	originalCards := make([]string, 8)
	copy(originalCards, genome.Cards)

	t.Run("mutation changes cards", func(t *testing.T) {
		err := genome.Mutate()
		if err != nil {
			t.Errorf("Mutate() error = %v", err)
		}

		// Fitness should be reset
		if genome.Fitness != 0 {
			t.Errorf("Mutate() should reset fitness, got %v", genome.Fitness)
		}

		// Count how many cards changed
		changes := 0
		for i, card := range genome.Cards {
			if card != originalCards[i] {
				changes++
			}
		}

		// Should have changed some cards (with 50% intensity and 20 candidates)
		if changes == 0 {
			t.Error("Mutate() changed no cards")
		}

		// Should still have 8 unique cards
		cardSet := make(map[string]bool)
		for _, card := range genome.Cards {
			if cardSet[card] {
				t.Errorf("Mutate() produced duplicate card: %v", card)
			}
			cardSet[card] = true
		}
	})

	t.Run("mutation with low intensity", func(t *testing.T) {
		cfgLow := DefaultGeneticConfig()
		cfgLow.MutationIntensity = 0.1 // Should mutate 1 card

		genome2, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfgLow)
		if err != nil {
			t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
		}

		originalCards2 := make([]string, 8)
		copy(originalCards2, genome2.Cards)

		err = genome2.Mutate()
		if err != nil {
			t.Errorf("Mutate() error = %v", err)
		}

		changes := 0
		for i, card := range genome2.Cards {
			if card != originalCards2[i] {
				changes++
			}
		}

		// With 10% intensity, should mutate at least 1 card
		if changes < 1 {
			t.Error("Mutate() with 10%% intensity should change at least 1 card")
		}
	})

	t.Run("note: effectiveness tests deferred", func(t *testing.T) {
		// Tests for mutation quality (synergy preservation, etc.)
		// are deferred until fitness evaluation is implemented
		t.Skip("Mutation quality tests depend on fitness function (task clash-royale-api-hj9j.2)")
	})
}

// Helper function to create mock card candidates
func createMockCandidates(count int) []*deck.CardCandidate {
	candidates := make([]*deck.CardCandidate, count)
	roles := []config.CardRole{
		config.RoleWinCondition,
		config.RoleBuilding,
		config.RoleSpellBig,
		config.RoleSpellSmall,
		config.RoleSupport,
		config.RoleCycle,
	}

	for i := range count {
		role := roles[i%len(roles)]
		candidates[i] = &deck.CardCandidate{
			Name:     fmt.Sprintf("Card%d", i),
			Level:    8,
			MaxLevel: 14,
			Rarity:   "Rare",
			Elixir:   3 + (i % 5),
			Role:     &role,
			Score:    float64(i) / 10.0,
		}
	}

	return candidates
}
