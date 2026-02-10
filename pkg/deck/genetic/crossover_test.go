// Package genetic provides genetic algorithm-based deck optimization
// using the eaopt library for evolutionary deck generation.
package genetic

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestDeckGenomeCrossoverExecution(t *testing.T) {
	candidates := createMockCandidates(20)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	parent1Cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	parent2Cards := []string{"Card8", "Card9", "Card10", "Card11", "Card12", "Card13", "Card14", "Card15"}

	parent1, err := NewDeckGenomeFromCards(parent1Cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() parent1 failed: %v", err)
	}

	parent2, err := NewDeckGenomeFromCards(parent2Cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() parent2 failed: %v", err)
	}

	offspring, err := parent1.Crossover(parent2)
	if err != nil {
		t.Errorf("Crossover() error = %v", err)
	}

	if offspring == nil {
		t.Fatal("Crossover() returned nil offspring")
	}

	offspringDeck, ok := offspring.(*DeckGenome)
	if !ok {
		t.Fatal("Crossover() did not return *DeckGenome")
	}

	// Validate offspring has 8 cards
	if len(offspringDeck.Cards) != 8 {
		t.Errorf("Crossover() produced %d cards, want 8", len(offspringDeck.Cards))
	}

	// Validate all cards are unique
	cardSet := make(map[string]bool)
	for _, card := range offspringDeck.Cards {
		if cardSet[card] {
			t.Errorf("Crossover() produced duplicate card: %v", card)
		}
		cardSet[card] = true
	}

	// All offspring cards should be valid candidates
	candidateMap := make(map[string]bool)
	for _, candidate := range candidates {
		candidateMap[candidate.Name] = true
	}

	for _, card := range offspringDeck.Cards {
		if !candidateMap[card] {
			t.Errorf("Crossover() produced invalid candidate card: %v", card)
		}
	}
}

func TestCrossoverInvalidType(t *testing.T) {
	candidates := createMockCandidates(10)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	parent1, err := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)
	if err != nil {
		t.Fatalf("NewDeckGenomeFromCards() failed: %v", err)
	}

	// Try crossover with wrong type
	_, err = parent1.Crossover("not a genome")
	if err == nil {
		t.Error("Crossover() with invalid type should error")
	}
	if !contains(err.Error(), "DeckGenome") {
		t.Errorf("Crossover() error should mention DeckGenome, got: %v", err)
	}
}

func TestCrossoverStrategies(t *testing.T) {
	// Test that all crossover strategies produce valid offspring
	candidates := createMockCandidates(25)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	parent1Cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	parent2Cards := []string{"Card10", "Card11", "Card12", "Card13", "Card14", "Card15", "Card16", "Card17"}

	parent1, _ := NewDeckGenomeFromCards(parent1Cards, candidates, strategy, &cfg)
	parent2, _ := NewDeckGenomeFromCards(parent2Cards, candidates, strategy, &cfg)

	// Run many crossovers to exercise all strategy paths
	for i := 0; i < 100; i++ {
		offspring, err := parent1.Crossover(parent2)
		if err != nil {
			t.Errorf("Crossover() iteration %d error = %v", i, err)
			continue
		}

		offspringDeck := offspring.(*DeckGenome)

		// Validate deck size
		if len(offspringDeck.Cards) != 8 {
			t.Errorf("Crossover() iteration %d produced %d cards", i, len(offspringDeck.Cards))
		}

		// Validate uniqueness
		cardSet := make(map[string]bool)
		for _, card := range offspringDeck.Cards {
			cardSet[card] = true
		}
		if len(cardSet) != 8 {
			t.Errorf("Crossover() iteration %d produced non-unique cards", i)
		}
	}
}

func TestUniformCrossover(t *testing.T) {
	candidates := createMockCandidates(20)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	parent1Cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	parent2Cards := []string{"Card8", "Card9", "Card10", "Card11", "Card12", "Card13", "Card14", "Card15"}

	parent1, _ := NewDeckGenomeFromCards(parent1Cards, candidates, strategy, &cfg)
	parent2, _ := NewDeckGenomeFromCards(parent2Cards, candidates, strategy, &cfg)

	offspring := parent1.uniformCrossover(parent2)

	// Should have 8 cards (though may have duplicates before repair)
	if len(offspring) != 8 {
		t.Errorf("uniformCrossover() produced %d cards, want 8", len(offspring))
	}

	// All cards should come from parent pools
	parentPool := make(map[string]bool)
	for _, card := range parent1.Cards {
		parentPool[card] = true
	}
	for _, card := range parent2.Cards {
		parentPool[card] = true
	}

	for i, card := range offspring {
		if !parentPool[card] {
			t.Errorf("uniformCrossover() card[%d] = %v not from parent pools", i, card)
		}
	}
}

func TestRolePreservingCrossover(t *testing.T) {
	candidates := createMockCandidates(20)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	parent1Cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	parent2Cards := []string{"Card8", "Card9", "Card10", "Card11", "Card12", "Card13", "Card14", "Card15"}

	parent1, _ := NewDeckGenomeFromCards(parent1Cards, candidates, strategy, &cfg)
	parent2, _ := NewDeckGenomeFromCards(parent2Cards, candidates, strategy, &cfg)

	offspring := parent1.rolePreservingCrossover(parent2)

	// Should produce some cards
	if len(offspring) == 0 {
		t.Error("rolePreservingCrossover() produced empty offspring")
	}

	// All cards should come from parent pools
	parentPool := make(map[string]bool)
	for _, card := range parent1.Cards {
		parentPool[card] = true
	}
	for _, card := range parent2.Cards {
		parentPool[card] = true
	}

	for i, card := range offspring {
		if !parentPool[card] {
			t.Errorf("rolePreservingCrossover() card[%d] = %v not from parent pools", i, card)
		}
	}
}

func TestSynergyAwareCrossover(t *testing.T) {
	candidates := createMockCandidates(20)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	parent1Cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	parent2Cards := []string{"Card8", "Card9", "Card10", "Card11", "Card12", "Card13", "Card14", "Card15"}

	parent1, _ := NewDeckGenomeFromCards(parent1Cards, candidates, strategy, &cfg)
	parent2, _ := NewDeckGenomeFromCards(parent2Cards, candidates, strategy, &cfg)

	offspring := parent1.synergyAwareCrossover(parent2)

	// Should produce up to 8 cards
	if len(offspring) > 8 {
		t.Errorf("synergyAwareCrossover() produced %d cards, want <= 8", len(offspring))
	}

	// All cards should come from parent pools
	parentPool := make(map[string]bool)
	for _, card := range parent1.Cards {
		parentPool[card] = true
	}
	for _, card := range parent2.Cards {
		parentPool[card] = true
	}

	for i, card := range offspring {
		if !parentPool[card] {
			t.Errorf("synergyAwareCrossover() card[%d] = %v not from parent pools", i, card)
		}
	}

	// Cards in synergy offspring should be unique
	cardSet := make(map[string]bool)
	for _, card := range offspring {
		if cardSet[card] {
			t.Errorf("synergyAwareCrossover() produced duplicate card: %v", card)
		}
		cardSet[card] = true
	}
}

func TestRepairDeck(t *testing.T) {
	candidates := createMockCandidates(20)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, _ := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)

	tests := []struct {
		name       string
		brokenDeck []string
		wantSize   int
		wantUnique bool
	}{
		{
			name:       "already valid deck",
			brokenDeck: []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"},
			wantSize:   8,
			wantUnique: true,
		},
		{
			name:       "deck with duplicates",
			brokenDeck: []string{"Card0", "Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6"},
			wantSize:   8,
			wantUnique: true,
		},
		{
			name:       "deck too short",
			brokenDeck: []string{"Card0", "Card1", "Card2"},
			wantSize:   8,
			wantUnique: true,
		},
		{
			name:       "deck too long",
			brokenDeck: []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7", "Card8", "Card9"},
			wantSize:   8,
			wantUnique: true,
		},
		{
			name:       "empty deck",
			brokenDeck: []string{},
			wantSize:   8,
			wantUnique: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repaired := genome.repairDeck(tt.brokenDeck)

			if len(repaired) != tt.wantSize {
				t.Errorf("repairDeck() size = %d, want %d", len(repaired), tt.wantSize)
			}

			if tt.wantUnique {
				cardSet := make(map[string]bool)
				for _, card := range repaired {
					if cardSet[card] {
						t.Errorf("repairDeck() produced duplicate card: %v", card)
					}
					cardSet[card] = true
				}
			}

			// All repaired cards should be valid candidates
			candidateMap := make(map[string]bool)
			for _, candidate := range candidates {
				candidateMap[candidate.Name] = true
			}

			for _, card := range repaired {
				if !candidateMap[card] {
					t.Errorf("repairDeck() produced invalid candidate: %v", card)
				}
			}
		})
	}
}

func TestRepairDeckWithParents(t *testing.T) {
	candidates := createMockCandidates(20)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	parent1Cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	parent2Cards := []string{"Card8", "Card9", "Card10", "Card11", "Card12", "Card13", "Card14", "Card15"}

	parent1, _ := NewDeckGenomeFromCards(parent1Cards, candidates, strategy, &cfg)
	parent2, _ := NewDeckGenomeFromCards(parent2Cards, candidates, strategy, &cfg)

	// Broken deck with only 3 cards
	brokenDeck := []string{"Card0", "Card1", "Card2"}

	repaired := parent1.repairDeck(brokenDeck, parent2)

	// Should fill to 8 cards using parents
	if len(repaired) != 8 {
		t.Errorf("repairDeck() with parents size = %d, want 8", len(repaired))
	}

	// Should prioritize cards from broken deck and parents
	hasOriginal := 0
	for _, card := range brokenDeck {
		for _, repairedCard := range repaired {
			if card == repairedCard {
				hasOriginal++
				break
			}
		}
	}

	if hasOriginal != len(brokenDeck) {
		t.Errorf("repairDeck() lost %d original cards", len(brokenDeck)-hasOriginal)
	}

	// Verify uniqueness
	cardSet := make(map[string]bool)
	for _, card := range repaired {
		if cardSet[card] {
			t.Errorf("repairDeck() with parents produced duplicate: %v", card)
		}
		cardSet[card] = true
	}
}

func TestEnsureWinCondition(t *testing.T) {
	candidates := createMockCandidates(20)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, _ := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)

	candidateMap := genome.candidateMap()
	used := make(map[string]bool)
	for _, card := range cards {
		used[card] = true
	}

	// Test ensureWinCondition
	result := genome.ensureWinCondition(cards, used, candidateMap)

	// Should return valid deck
	if len(result) != 8 {
		t.Errorf("ensureWinCondition() size = %d, want 8", len(result))
	}
}

func TestCandidateMap(t *testing.T) {
	candidates := createMockCandidates(15)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, _ := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)

	candidateMap := genome.candidateMap()

	// Should have all candidates
	if len(candidateMap) != len(candidates) {
		t.Errorf("candidateMap() size = %d, want %d", len(candidateMap), len(candidates))
	}

	// Verify all candidates are in map
	for _, candidate := range candidates {
		if _, ok := candidateMap[candidate.Name]; !ok {
			t.Errorf("candidateMap() missing candidate: %v", candidate.Name)
		}
	}
}

func TestCardsByRole(t *testing.T) {
	candidates := createMockCandidates(15)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, _ := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)

	candidateMap := genome.candidateMap()
	byRole := genome.cardsByRole(cards, candidateMap)

	// Should have at least one role
	if len(byRole) == 0 {
		t.Error("cardsByRole() returned empty map")
	}

	// All cards should be accounted for
	totalCards := 0
	for _, roleCards := range byRole {
		totalCards += len(roleCards)
	}

	if totalCards != len(cards) {
		t.Errorf("cardsByRole() total cards = %d, want %d", totalCards, len(cards))
	}

	// Verify no duplicates within roles
	for role, roleCards := range byRole {
		cardSet := make(map[string]bool)
		for _, card := range roleCards {
			if cardSet[card] {
				t.Errorf("cardsByRole() role %v has duplicate card: %v", role, card)
			}
			cardSet[card] = true
		}
	}
}

func TestFindSynergyPairs(t *testing.T) {
	candidates := createMockCandidates(10)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	genome, _ := NewDeckGenomeFromCards(cards, candidates, strategy, &cfg)

	db := deck.NewSynergyDatabase()
	pairs := genome.findSynergyPairs(cards, db)

	// Pairs might be nil or empty if no high synergy found (threshold 0.8)
	// This is acceptable - the function returns nil for empty slice
	t.Logf("Found %d synergy pairs (nil is OK)", len(pairs))

	// Verify pair structure if any exist
	for i, pair := range pairs {
		if pair.Card1 == "" || pair.Card2 == "" {
			t.Errorf("findSynergyPairs() pair[%d] has empty card", i)
		}
		if pair.Score < synergyCrossoverThreshold {
			t.Errorf("findSynergyPairs() pair[%d] score %v < threshold %v",
				i, pair.Score, synergyCrossoverThreshold)
		}
	}
}

func TestCrossoverPreservesConfig(t *testing.T) {
	candidates := createMockCandidates(15)
	cfg := GeneticConfig{
		PopulationSize:    100,
		Generations:       50,
		MutationRate:      0.15,
		CrossoverRate:     0.85,
		MutationIntensity: 0.4,
	}
	strategy := deck.StrategyBalanced

	parent1Cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	parent2Cards := []string{"Card7", "Card8", "Card9", "Card10", "Card11", "Card12", "Card13", "Card14"}

	parent1, _ := NewDeckGenomeFromCards(parent1Cards, candidates, strategy, &cfg)
	parent2, _ := NewDeckGenomeFromCards(parent2Cards, candidates, strategy, &cfg)

	offspring, err := parent1.Crossover(parent2)
	if err != nil {
		t.Fatalf("Crossover() error = %v", err)
	}

	offspringDeck := offspring.(*DeckGenome)

	// Offspring should inherit parent1's config
	if offspringDeck.config == nil {
		t.Error("Crossover() offspring has nil config")
	}
	if offspringDeck.config.MutationRate != cfg.MutationRate {
		t.Errorf("Crossover() offspring MutationRate = %v, want %v",
			offspringDeck.config.MutationRate, cfg.MutationRate)
	}

	// Offspring should inherit parent1's strategy
	if offspringDeck.strategy != strategy {
		t.Errorf("Crossover() offspring strategy = %v, want %v",
			offspringDeck.strategy, strategy)
	}

	// Offspring should have same candidates reference
	if len(offspringDeck.candidates) != len(candidates) {
		t.Errorf("Crossover() offspring candidates = %d, want %d",
			len(offspringDeck.candidates), len(candidates))
	}
}

func TestCrossoverOffspringFromParents(t *testing.T) {
	// Verify that offspring cards come from parent pools
	candidates := createMockCandidates(20)
	cfg := DefaultGeneticConfig()
	strategy := deck.StrategyBalanced

	parent1Cards := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	parent2Cards := []string{"Card10", "Card11", "Card12", "Card13", "Card14", "Card15", "Card16", "Card17"}

	parent1, _ := NewDeckGenomeFromCards(parent1Cards, candidates, strategy, &cfg)
	parent2, _ := NewDeckGenomeFromCards(parent2Cards, candidates, strategy, &cfg)

	// Build parent pool
	parentPool := make(map[string]bool)
	for _, card := range parent1.Cards {
		parentPool[card] = true
	}
	for _, card := range parent2.Cards {
		parentPool[card] = true
	}

	// Run multiple crossovers
	for i := 0; i < 50; i++ {
		offspring, err := parent1.Crossover(parent2)
		if err != nil {
			t.Errorf("Crossover() iteration %d error = %v", i, err)
			continue
		}

		offspringDeck := offspring.(*DeckGenome)

		// Count how many cards came from parents
		fromParents := 0
		for _, card := range offspringDeck.Cards {
			if parentPool[card] {
				fromParents++
			}
		}

		// Most cards should come from parents (repair may add new cards)
		if fromParents == 0 {
			t.Errorf("Crossover() iteration %d: no cards from parents", i)
		}
	}
}
