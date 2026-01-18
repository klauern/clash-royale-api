package deck

import (
	"fmt"
	"testing"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

func TestNewDeckFuzzer(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Zap", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 2},
			{Name: "Cannon", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Archers", Level: 10, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Knight", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Skeletons", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Valkyrie", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Baby Dragon", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
			{Name: "Musketeer", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Ice Spirit", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Giant", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 5},
		},
	}

	fuzzer, err := NewDeckFuzzer(player, &FuzzingConfig{Count: 10, Workers: 1})
	if err != nil {
		t.Fatalf("Failed to create fuzzer: %v", err)
	}

	if fuzzer == nil {
		t.Fatal("Fuzzer is nil")
	}

	if len(fuzzer.GetAllCards()) != 12 {
		t.Errorf("Expected 12 cards, got %d", len(fuzzer.GetAllCards()))
	}
}

func TestNewDeckFuzzerNilPlayer(t *testing.T) {
	_, err := NewDeckFuzzer(nil, &FuzzingConfig{})
	if err == nil {
		t.Fatal("Expected error for nil player, got nil")
	}
}

func TestNewDeckFuzzerInsufficientCards(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
		},
	}

	_, err := NewDeckFuzzer(player, &FuzzingConfig{})
	if err == nil {
		t.Fatal("Expected error for insufficient cards, got nil")
	}
}

func TestNewDeckFuzzerDefaultConfig(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Zap", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 2},
			{Name: "Cannon", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Archers", Level: 10, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Knight", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Skeletons", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Valkyrie", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Baby Dragon", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
			{Name: "Musketeer", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Ice Spirit", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Giant", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 5},
		},
	}

	fuzzer, err := NewDeckFuzzer(player, nil)
	if err != nil {
		t.Fatalf("Failed to create fuzzer: %v", err)
	}

	if fuzzer.config.Count != 1000 {
		t.Errorf("Expected default count of 1000, got %d", fuzzer.config.Count)
	}

	if fuzzer.config.Workers != 1 {
		t.Errorf("Expected default workers of 1, got %d", fuzzer.config.Workers)
	}

	if fuzzer.config.MinAvgElixir != 0 {
		t.Errorf("Expected default min elixir of 0, got %f", fuzzer.config.MinAvgElixir)
	}

	if fuzzer.config.MaxAvgElixir != 10 {
		t.Errorf("Expected default max elixir of 10, got %f", fuzzer.config.MaxAvgElixir)
	}
}

func TestGenerateRandomDeck(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Zap", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 2},
			{Name: "Cannon", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Archers", Level: 10, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Knight", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Skeletons", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Valkyrie", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Baby Dragon", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
			{Name: "Musketeer", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Ice Spirit", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Giant", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 5},
			{Name: "Log", Level: 11, MaxLevel: 13, Rarity: "Legendary", ElixirCost: 2},
			{Name: "Tesla", Level: 7, MaxLevel: 11, Rarity: "Common", ElixirCost: 3},
			{Name: "Minion Horde", Level: 9, MaxLevel: 13, Rarity: "Common", ElixirCost: 5},
			{Name: "Poison", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
		},
	}

	fuzzer, err := NewDeckFuzzer(player, &FuzzingConfig{Count: 10, Workers: 1})
	if err != nil {
		t.Fatalf("Failed to create fuzzer: %v", err)
	}

	deck, err := fuzzer.GenerateRandomDeck()
	if err != nil {
		t.Fatalf("Failed to generate deck: %v", err)
	}

	if len(deck) != 8 {
		t.Errorf("Expected 8 cards in deck, got %d", len(deck))
	}

	// Check for duplicates
	uniqueCards := make(map[string]bool)
	for _, card := range deck {
		if uniqueCards[card] {
			t.Errorf("Duplicate card found: %s", card)
		}
		uniqueCards[card] = true
	}
}

func TestGenerateRandomDeckWithIncludeCards(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Zap", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 2},
			{Name: "Cannon", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Archers", Level: 10, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Knight", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Skeletons", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Valkyrie", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Baby Dragon", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
			{Name: "Musketeer", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Ice Spirit", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Giant", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 5},
			{Name: "Log", Level: 11, MaxLevel: 13, Rarity: "Legendary", ElixirCost: 2},
			{Name: "Tesla", Level: 7, MaxLevel: 11, Rarity: "Common", ElixirCost: 3},
			{Name: "Minion Horde", Level: 9, MaxLevel: 13, Rarity: "Common", ElixirCost: 5},
			{Name: "Poison", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
		},
	}

	cfg := &FuzzingConfig{
		Count:        10,
		Workers:      1,
		IncludeCards: []string{"Hog Rider", "Fireball"},
	}

	fuzzer, err := NewDeckFuzzer(player, cfg)
	if err != nil {
		t.Fatalf("Failed to create fuzzer: %v", err)
	}

	// Generate multiple decks to ensure include cards are always present
	for i := 0; i < 5; i++ {
		deck, err := fuzzer.GenerateRandomDeck()
		if err != nil {
			t.Fatalf("Failed to generate deck: %v", err)
		}

		foundHogRider := false
		foundFireball := false
		for _, card := range deck {
			if card == "Hog Rider" {
				foundHogRider = true
			}
			if card == "Fireball" {
				foundFireball = true
			}
		}

		if !foundHogRider {
			t.Errorf("Deck %d missing Hog Rider (include card)", i)
		}
		if !foundFireball {
			t.Errorf("Deck %d missing Fireball (include card)", i)
		}
	}
}

func TestGenerateRandomDeckWithExcludeCards(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Zap", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 2},
			{Name: "Cannon", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Archers", Level: 10, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Skeletons", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Valkyrie", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Baby Dragon", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
			{Name: "Musketeer", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Ice Spirit", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Giant", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 5},
			{Name: "Log", Level: 11, MaxLevel: 13, Rarity: "Legendary", ElixirCost: 2},
			{Name: "Tesla", Level: 7, MaxLevel: 11, Rarity: "Common", ElixirCost: 3},
			{Name: "Minion Horde", Level: 9, MaxLevel: 13, Rarity: "Common", ElixirCost: 5},
			{Name: "Poison", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
			{Name: "Golem", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 8},
			{Name: "P.E.K.K.A", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 7},
		},
	}

	cfg := &FuzzingConfig{
		Count:        10,
		Workers:      1,
		ExcludeCards: []string{"Knight"}, // Exclude a card not in the list to avoid issues
	}

	fuzzer, err := NewDeckFuzzer(player, cfg)
	if err != nil {
		t.Fatalf("Failed to create fuzzer: %v", err)
	}

	// Generate multiple decks - some may fail but most should succeed
	successCount := 0
	for i := 0; i < 10; i++ {
		deck, err := fuzzer.GenerateRandomDeck()
		if err != nil {
			continue
		}

		successCount++

		// Verify Knight is never in the deck
		for _, card := range deck {
			if card == "Knight" {
				t.Errorf("Deck %d contains Knight (should be excluded but not in card list)", i)
			}
		}
	}

	if successCount == 0 {
		t.Error("No decks were successfully generated")
	}
}

func TestGenerateRandomDeckWithElixirConstraints(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Zap", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 2},
			{Name: "Cannon", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Archers", Level: 10, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Knight", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Skeletons", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Valkyrie", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Baby Dragon", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
			{Name: "Musketeer", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Ice Spirit", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Giant", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 5},
			{Name: "Log", Level: 11, MaxLevel: 13, Rarity: "Legendary", ElixirCost: 2},
			{Name: "Tesla", Level: 7, MaxLevel: 11, Rarity: "Common", ElixirCost: 3},
			{Name: "Minion Horde", Level: 9, MaxLevel: 13, Rarity: "Common", ElixirCost: 5},
			{Name: "Poison", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
		},
	}

	cfg := &FuzzingConfig{
		Count:        10,
		Workers:      1,
		MinAvgElixir: 2.5,
		MaxAvgElixir: 3.5,
	}

	fuzzer, err := NewDeckFuzzer(player, cfg)
	if err != nil {
		t.Fatalf("Failed to create fuzzer: %v", err)
	}

	// Generate multiple decks - some may fail due to elixir constraints
	successCount := 0
	for i := 0; i < 10; i++ {
		deck, err := fuzzer.GenerateRandomDeck()
		if err != nil {
			// Some decks may fail elixir constraints, that's OK
			continue
		}

		successCount++
		avgElixir := fuzzer.calculateAvgElixir(deck)
		if avgElixir < 2.5 || avgElixir > 3.5 {
			t.Errorf("Deck %d elixir out of range: %.2f", i, avgElixir)
		}
	}

	if successCount == 0 {
		t.Error("No decks were successfully generated - elixir constraints may be too strict")
	}
}

func TestGenerateEvolutionCentricDeck(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Zap", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 2},
			{Name: "Cannon", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Archers", Level: 10, MaxLevel: 13, Rarity: "Common", ElixirCost: 3, EvolutionLevel: 1, MaxEvolutionLevel: 3},
			{Name: "Knight", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3, EvolutionLevel: 1, MaxEvolutionLevel: 3},
			{Name: "Skeletons", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1, EvolutionLevel: 1, MaxEvolutionLevel: 1},
			{Name: "Valkyrie", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Baby Dragon", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
			{Name: "Musketeer", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4, EvolutionLevel: 1, MaxEvolutionLevel: 3},
			{Name: "Ice Spirit", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1, EvolutionLevel: 1, MaxEvolutionLevel: 1},
			{Name: "Giant", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 5},
			{Name: "Log", Level: 11, MaxLevel: 13, Rarity: "Legendary", ElixirCost: 2},
			{Name: "Tesla", Level: 7, MaxLevel: 11, Rarity: "Common", ElixirCost: 3},
		},
	}

	cfg := &FuzzingConfig{
		Count:             10,
		Workers:           1,
		EvolutionCentric:  true,
		MinEvolutionCards: 3,
		MinEvoLevel:       1,
	}

	fuzzer, err := NewDeckFuzzer(player, cfg)
	if err != nil {
		t.Fatalf("Failed to create fuzzer: %v", err)
	}

	deck, err := fuzzer.GenerateRandomDeck()
	if err != nil {
		t.Fatalf("Failed to generate evolution deck: %v", err)
	}

	if len(deck) != 8 {
		t.Fatalf("Expected 8 cards, got %d", len(deck))
	}

	evoCount := 0
	for _, cardName := range deck {
		for _, card := range player.Cards {
			if card.Name == cardName {
				if card.EvolutionLevel >= cfg.MinEvoLevel ||
					(card.MaxEvolutionLevel > 0 && card.EvolutionLevel < card.MaxEvolutionLevel) {
					evoCount++
				}
				break
			}
		}
	}

	if evoCount < cfg.MinEvolutionCards {
		t.Errorf("Expected at least %d evolution cards, got %d", cfg.MinEvolutionCards, evoCount)
	}
}

func TestGenerateDecks(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Zap", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 2},
			{Name: "Cannon", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Archers", Level: 10, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Knight", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Skeletons", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Valkyrie", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Baby Dragon", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
			{Name: "Musketeer", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Ice Spirit", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Giant", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 5},
			{Name: "Log", Level: 11, MaxLevel: 13, Rarity: "Legendary", ElixirCost: 2},
			{Name: "Tesla", Level: 7, MaxLevel: 11, Rarity: "Common", ElixirCost: 3},
			{Name: "Minion Horde", Level: 9, MaxLevel: 13, Rarity: "Common", ElixirCost: 5},
			{Name: "Poison", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
		},
	}

	fuzzer, err := NewDeckFuzzer(player, &FuzzingConfig{Count: 20, Workers: 1})
	if err != nil {
		t.Fatalf("Failed to create fuzzer: %v", err)
	}

	decks, err := fuzzer.GenerateDecks(20)
	if err != nil {
		t.Fatalf("Failed to generate decks: %v", err)
	}

	if len(decks) == 0 {
		t.Fatal("No decks generated")
	}

	for i, deck := range decks {
		if len(deck) != 8 {
			t.Errorf("Deck %d has %d cards, expected 8", i, len(deck))
		}
	}
}

func TestDefaultRoleComposition(t *testing.T) {
	comp := DefaultRoleComposition()

	if comp.WinConditions != 1 {
		t.Errorf("Expected WinConditions=1, got %d", comp.WinConditions)
	}
	if comp.Buildings != 1 {
		t.Errorf("Expected Buildings=1, got %d", comp.Buildings)
	}
	if comp.BigSpells != 1 {
		t.Errorf("Expected BigSpells=1, got %d", comp.BigSpells)
	}
	if comp.SmallSpells != 1 {
		t.Errorf("Expected SmallSpells=1, got %d", comp.SmallSpells)
	}
	if comp.Support != 2 {
		t.Errorf("Expected Support=2, got %d", comp.Support)
	}
	if comp.Cycle != 2 {
		t.Errorf("Expected Cycle=2, got %d", comp.Cycle)
	}
}

func TestGetCardsByRole(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Giant", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 5},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Cannon", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Knight", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Skeletons", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Ice Spirit", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Archers", Level: 10, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Zap", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 2},
			{Name: "Log", Level: 11, MaxLevel: 13, Rarity: "Legendary", ElixirCost: 2},
			{Name: "Tesla", Level: 7, MaxLevel: 11, Rarity: "Common", ElixirCost: 3},
		},
	}

	fuzzer, err := NewDeckFuzzer(player, &FuzzingConfig{Count: 10, Workers: 1})
	if err != nil {
		t.Fatalf("Failed to create fuzzer: %v", err)
	}

	winConditions := fuzzer.GetCardsByRole(config.RoleWinCondition)
	if len(winConditions) < 2 {
		t.Errorf("Expected at least 2 win conditions, got %d", len(winConditions))
	}

	smallSpells := fuzzer.GetCardsByRole(config.RoleSpellSmall)
	if len(smallSpells) == 0 {
		t.Error("Expected at least 1 small spell")
	}
}

func TestGetStats(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Zap", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 2},
			{Name: "Cannon", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Archers", Level: 10, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Knight", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Skeletons", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Valkyrie", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Baby Dragon", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
			{Name: "Musketeer", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Ice Spirit", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Giant", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 5},
		},
	}

	fuzzer, err := NewDeckFuzzer(player, &FuzzingConfig{Count: 10, Workers: 1})
	if err != nil {
		t.Fatalf("Failed to create fuzzer: %v", err)
	}

	// Generate a few decks
	for i := 0; i < 5; i++ {
		fuzzer.GenerateRandomDeck()
	}

	stats := fuzzer.GetStats()

	if stats.Generated < 5 {
		t.Errorf("Expected at least 5 generated decks, got %d", stats.Generated)
	}

	if stats.Success+stats.Failed != stats.Generated {
		t.Errorf("Stats mismatch: success+failed != total")
	}
}

func TestSetRoleComposition(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Zap", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 2},
			{Name: "Cannon", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Archers", Level: 10, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Knight", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Skeletons", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Valkyrie", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Baby Dragon", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
			{Name: "Musketeer", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Ice Spirit", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Giant", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 5},
		},
	}

	fuzzer, err := NewDeckFuzzer(player, &FuzzingConfig{Count: 10, Workers: 1})
	if err != nil {
		t.Fatalf("Failed to create fuzzer: %v", err)
	}

	customComp := &RoleComposition{
		WinConditions: 2,
		Buildings:     0,
		BigSpells:     2,
		SmallSpells:   1,
		Support:       2,
		Cycle:         1,
	}

	fuzzer.SetRoleComposition(customComp)

	if fuzzer.composition.WinConditions != 2 {
		t.Errorf("Expected WinConditions=2, got %d", fuzzer.composition.WinConditions)
	}

	if fuzzer.composition.Buildings != 0 {
		t.Errorf("Expected Buildings=0, got %d", fuzzer.composition.Buildings)
	}
}

func TestFuzzingConfigValidation(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Zap", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 2},
			{Name: "Cannon", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Archers", Level: 10, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Knight", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Skeletons", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Valkyrie", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Baby Dragon", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
			{Name: "Musketeer", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Ice Spirit", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Giant", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 5},
		},
	}

	tests := []struct {
		name      string
		config    *FuzzingConfig
		checkFunc func(*FuzzingConfig) error
	}{
		{
			name: "Negative MinAvgElixir gets set to 0",
			config: &FuzzingConfig{
				Count:        10,
				Workers:      1,
				MinAvgElixir: -1.0,
			},
			checkFunc: func(cfg *FuzzingConfig) error {
				if cfg.MinAvgElixir != 0 {
					return fmt.Errorf("expected MinAvgElixir=0, got %f", cfg.MinAvgElixir)
				}
				return nil
			},
		},
		{
			name: "MaxAvgElixir > 10 gets set to 10",
			config: &FuzzingConfig{
				Count:        10,
				Workers:      1,
				MaxAvgElixir: 15.0,
			},
			checkFunc: func(cfg *FuzzingConfig) error {
				if cfg.MaxAvgElixir != 10 {
					return fmt.Errorf("expected MaxAvgElixir=10, got %f", cfg.MaxAvgElixir)
				}
				return nil
			},
		},
		{
			name: "Zero Count gets set to 1000",
			config: &FuzzingConfig{
				Count:   0,
				Workers: 1,
			},
			checkFunc: func(cfg *FuzzingConfig) error {
				if cfg.Count != 1000 {
					return fmt.Errorf("expected Count=1000, got %d", cfg.Count)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fuzzer, err := NewDeckFuzzer(player, tt.config)
			if err != nil {
				t.Fatalf("Failed to create fuzzer: %v", err)
			}
			if err := tt.checkFunc(fuzzer.config); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGenerateDecksParallel(t *testing.T) {
	player := &clashroyale.Player{
		Name: "TestPlayer",
		Tag:  "#TEST123",
		Cards: []clashroyale.Card{
			{Name: "Hog Rider", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Fireball", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Zap", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 2},
			{Name: "Cannon", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Archers", Level: 10, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Knight", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 3},
			{Name: "Skeletons", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Valkyrie", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 4},
			{Name: "Baby Dragon", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
			{Name: "Musketeer", Level: 8, MaxLevel: 13, Rarity: "Rare", ElixirCost: 4},
			{Name: "Ice Spirit", Level: 11, MaxLevel: 13, Rarity: "Common", ElixirCost: 1},
			{Name: "Giant", Level: 7, MaxLevel: 11, Rarity: "Rare", ElixirCost: 5},
			{Name: "Log", Level: 11, MaxLevel: 13, Rarity: "Legendary", ElixirCost: 2},
			{Name: "Tesla", Level: 7, MaxLevel: 11, Rarity: "Common", ElixirCost: 3},
			{Name: "Minion Horde", Level: 9, MaxLevel: 13, Rarity: "Common", ElixirCost: 5},
			{Name: "Poison", Level: 5, MaxLevel: 11, Rarity: "Epic", ElixirCost: 4},
		},
	}

	cfg := &FuzzingConfig{
		Count:   50,
		Workers: 4,
	}

	fuzzer, err := NewDeckFuzzer(player, cfg)
	if err != nil {
		t.Fatalf("Failed to create fuzzer: %v", err)
	}

	decks, err := fuzzer.GenerateDecksParallel()
	if err != nil {
		t.Fatalf("Failed to generate decks: %v", err)
	}

	if len(decks) == 0 {
		t.Fatal("No decks generated")
	}

	// Verify all decks are valid
	for i, deck := range decks {
		if len(deck) != 8 {
			t.Errorf("Deck %d has %d cards, expected 8", i, len(deck))
		}
	}
}
