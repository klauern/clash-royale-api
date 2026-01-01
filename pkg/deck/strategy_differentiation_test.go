package deck

import (
	"testing"
)

// TestBuilder_StrategyArchetypeAffinity tests that archetype affinity bonuses
// help on-archetype cards compete with higher-level off-archetype cards
func TestBuilder_StrategyArchetypeAffinity(t *testing.T) {
	// Create analysis with cards at different levels to test archetype affinity
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			// Win conditions: Hog Rider is aggro-archetype but lower level
			"Hog Rider":     {Level: 8, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Goblin Barrel": {Level: 10, MaxLevel: 14, Rarity: "Epic", Elixir: 3},
			"Mortar":        {Level: 9, MaxLevel: 14, Rarity: "Common", Elixir: 4},
			// Buildings
			"Cannon":        {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Inferno Tower": {Level: 8, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Goblin Hut":    {Level: 11, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			// Spells
			"Fireball": {Level: 10, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Poison":   {Level: 9, MaxLevel: 14, Rarity: "Epic", Elixir: 4},
			"Zap":      {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 2},
			"Arrows":   {Level: 9, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			// Support
			"Archers":     {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Musketeer":   {Level: 10, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Baby Dragon": {Level: 10, MaxLevel: 14, Rarity: "Epic", Elixir: 4},
			"Ice Wizard":  {Level: 7, MaxLevel: 14, Rarity: "Legendary", Elixir: 3},
			"Mini PEKKA":  {Level: 8, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			// Cycle
			"Knight":      {Level: 11, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Skeletons":   {Level: 9, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Ice Spirit":  {Level: 9, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Goblin Gang": {Level: 11, MaxLevel: 14, Rarity: "Common", Elixir: 3},
		},
	}

	// Test Aggro strategy: Should prefer Hog Rider (aggro-archetype) despite being 2 levels lower
	t.Run("Aggro prefers archetype cards", func(t *testing.T) {
		builder := NewBuilder("testdata")
		err := builder.SetStrategy(StrategyAggro)
		if err != nil {
			t.Fatalf("SetStrategy failed: %v", err)
		}

		deck, err := builder.BuildDeckFromAnalysis(analysis)
		if err != nil {
			t.Fatalf("BuildDeckFromAnalysis failed: %v", err)
		}

		// Verify Hog Rider is selected (archetype affinity should help it overcome level difference)
		hasHogRider := false
		for _, card := range deck.Deck {
			if card == "Hog Rider" {
				hasHogRider = true
				break
			}
		}

		if !hasHogRider {
			t.Errorf("Aggro deck should include Hog Rider (archetype card) despite being 2 levels lower than Goblin Barrel")
			t.Logf("Deck: %v", deck.Deck)
		}
	})

	// Test Control strategy: Should prefer defensive cards
	t.Run("Control prefers defensive archetype", func(t *testing.T) {
		builder := NewBuilder("testdata")
		err := builder.SetStrategy(StrategyControl)
		if err != nil {
			t.Fatalf("SetStrategy failed: %v", err)
		}

		deck, err := builder.BuildDeckFromAnalysis(analysis)
		if err != nil {
			t.Fatalf("BuildDeckFromAnalysis failed: %v", err)
		}

		// Verify Mortar or Poison are selected (control archetypes)
		hasMortar := false
		hasPoison := false
		for _, card := range deck.Deck {
			if card == "Mortar" {
				hasMortar = true
			}
			if card == "Poison" {
				hasPoison = true
			}
		}

		if !hasMortar && !hasPoison {
			t.Errorf("Control deck should include defensive archetypes like Mortar or Poison")
			t.Logf("Deck: %v", deck.Deck)
		}
	})

	// Test Cycle strategy: Should prefer low-cost archetype cards
	t.Run("Cycle prefers low-cost archetype", func(t *testing.T) {
		builder := NewBuilder("testdata")
		err := builder.SetStrategy(StrategyCycle)
		if err != nil {
			t.Fatalf("SetStrategy failed: %v", err)
		}

		deck, err := builder.BuildDeckFromAnalysis(analysis)
		if err != nil {
			t.Fatalf("BuildDeckFromAnalysis failed: %v", err)
		}

		// Verify low average elixir (cycle should be fast)
		if deck.AvgElixir > 3.0 {
			t.Errorf("Cycle deck should have average elixir <= 3.0, got %.2f", deck.AvgElixir)
		}

		// Verify archetype cycle cards are preferred (Skeletons, Ice Spirit, etc.)
		hasArchetypeCycle := false
		for _, card := range deck.Deck {
			if card == "Skeletons" || card == "Ice Spirit" || card == "Archers" || card == "Knight" {
				hasArchetypeCycle = true
				break
			}
		}

		if !hasArchetypeCycle {
			t.Errorf("Cycle deck should include archetype cycle cards")
			t.Logf("Deck: %v", deck.Deck)
		}
	})
}

// TestBuilder_StrategyDifferentiation tests that different strategies produce
// meaningfully different deck compositions when given the same card collection
func TestBuilder_StrategyDifferentiation(t *testing.T) {
	// Create a balanced card collection
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			// Win conditions
			"Hog Rider":     {Level: 10, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Royal Giant":   {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 6},
			"Goblin Barrel": {Level: 10, MaxLevel: 14, Rarity: "Epic", Elixir: 3},
			"Mortar":        {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 4},
			// Buildings
			"Cannon":        {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Inferno Tower": {Level: 10, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Bomb Tower":    {Level: 10, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			// Spells
			"Fireball":  {Level: 10, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Poison":    {Level: 10, MaxLevel: 14, Rarity: "Epic", Elixir: 4},
			"Lightning": {Level: 10, MaxLevel: 14, Rarity: "Epic", Elixir: 6},
			"Zap":       {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 2},
			"Log":       {Level: 10, MaxLevel: 14, Rarity: "Legendary", Elixir: 2},
			"Arrows":    {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			// Support
			"Archers":     {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Musketeer":   {Level: 10, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Wizard":      {Level: 10, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Valkyrie":    {Level: 10, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Baby Dragon": {Level: 10, MaxLevel: 14, Rarity: "Epic", Elixir: 4},
			// Cycle
			"Knight":        {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Skeletons":     {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Ice Spirit":    {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Goblin Gang":   {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Spear Goblins": {Level: 10, MaxLevel: 14, Rarity: "Common", Elixir: 2},
		},
	}

	// Build decks with different strategies
	strategies := []Strategy{StrategyBalanced, StrategyAggro, StrategyControl, StrategyCycle}
	decks := make(map[Strategy][]string)

	for _, strategy := range strategies {
		builder := NewBuilder("testdata")
		err := builder.SetStrategy(strategy)
		if err != nil {
			t.Fatalf("SetStrategy(%v) failed: %v", strategy, err)
		}

		deck, err := builder.BuildDeckFromAnalysis(analysis)
		if err != nil {
			t.Fatalf("BuildDeckFromAnalysis for %v failed: %v", strategy, err)
		}

		decks[strategy] = deck.Deck
	}

	// Verify that strategies produce different decks
	// At least some cards should differ between strategies
	comparisons := []struct {
		strategy1 Strategy
		strategy2 Strategy
	}{
		{StrategyAggro, StrategyControl},
		{StrategyAggro, StrategyCycle},
		{StrategyControl, StrategyCycle},
	}

	for _, comp := range comparisons {
		deck1 := decks[comp.strategy1]
		deck2 := decks[comp.strategy2]

		// Count differences
		differences := 0
		for _, card1 := range deck1 {
			found := false
			for _, card2 := range deck2 {
				if card1 == card2 {
					found = true
					break
				}
			}
			if !found {
				differences++
			}
		}

		// Expect at least 2 different cards between strategies
		if differences < 2 {
			t.Errorf("Strategies %v and %v should produce more different decks (only %d differences)",
				comp.strategy1, comp.strategy2, differences)
			t.Logf("%v deck: %v", comp.strategy1, deck1)
			t.Logf("%v deck: %v", comp.strategy2, deck2)
		}
	}
}
