package deck

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuilder_BuildDeckFromAnalysis(t *testing.T) {
	builder := NewBuilder("testdata")

	// Create test analysis data
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Hog Rider": {
				Level:    8,
				MaxLevel: 13,
				Rarity:   "Rare",
				Elixir:   4,
			},
			"Fireball": {
				Level:    7,
				MaxLevel: 11,
				Rarity:   "Rare",
				Elixir:   4,
			},
			"Zap": {
				Level:    11,
				MaxLevel: 13,
				Rarity:   "Common",
				Elixir:   2,
			},
			"Cannon": {
				Level:    11,
				MaxLevel: 13,
				Rarity:   "Common",
				Elixir:   3,
			},
			"Archers": {
				Level:    10,
				MaxLevel: 13,
				Rarity:   "Common",
				Elixir:   3,
			},
			"Knight": {
				Level:    11,
				MaxLevel: 13,
				Rarity:   "Common",
				Elixir:   3,
			},
			"Skeletons": {
				Level:    11,
				MaxLevel: 13,
				Rarity:   "Common",
				Elixir:   1,
			},
			"Valkyrie": {
				Level:    7,
				MaxLevel: 11,
				Rarity:   "Rare",
				Elixir:   4,
			},
			"Baby Dragon": {
				Level:    5,
				MaxLevel: 11,
				Rarity:   "Epic",
				Elixir:   4,
			},
		},
		AnalysisTime: "2024-01-15T10:30:00Z",
	}

	deck, err := builder.BuildDeckFromAnalysis(analysis)
	if err != nil {
		t.Fatalf("Failed to build deck: %v", err)
	}

	// Validate deck structure
	if len(deck.Deck) != 8 {
		t.Errorf("Expected 8 cards in deck, got %d", len(deck.Deck))
	}

	if len(deck.DeckDetail) != 8 {
		t.Errorf("Expected 8 card details, got %d", len(deck.DeckDetail))
	}

	if deck.AvgElixir <= 0 || deck.AvgElixir > 10 {
		t.Errorf("Invalid average elixir: %f", deck.AvgElixir)
	}

	// Validate card details
	for i, cardName := range deck.Deck {
		detail := deck.DeckDetail[i]
		if detail.Name != cardName {
			t.Errorf("Card name mismatch at index %d: %s != %s", i, cardName, detail.Name)
		}
		if detail.Level <= 0 || detail.Level > detail.MaxLevel {
			t.Errorf("Invalid level for %s: %d/%d", detail.Name, detail.Level, detail.MaxLevel)
		}
	}

	// Test that we have different roles represented
	roles := make(map[string]bool)
	for _, detail := range deck.DeckDetail {
		if detail.Role != "" {
			roles[detail.Role] = true
		}
	}

	// Should have at least some variety in roles
	if len(roles) < 2 {
		t.Logf("Note: Limited role variety in test deck: %v", roles)
	}

	t.Logf("Generated deck: %v", deck.Deck)
	t.Logf("Average elixir: %.2f", deck.AvgElixir)
	t.Logf("Notes: %v", deck.Notes)
}

func TestBuilder_LoadAndSaveDeck(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "deck_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	builder := NewBuilder(tempDir)

	// Create test analysis
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Hog Rider": {Level: 8, MaxLevel: 13, Rarity: "Rare", Elixir: 4},
			"Fireball":  {Level: 7, MaxLevel: 11, Rarity: "Rare", Elixir: 4},
			"Zap":       {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 2},
			"Cannon":    {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 3},
			"Archers":   {Level: 10, MaxLevel: 13, Rarity: "Common", Elixir: 3},
			"Knight":    {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 3},
			"Skeletons": {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 1},
			"Valkyrie":  {Level: 7, MaxLevel: 11, Rarity: "Rare", Elixir: 4},
		},
	}

	// Build deck
	deck, err := builder.BuildDeckFromAnalysis(analysis)
	if err != nil {
		t.Fatalf("Failed to build deck: %v", err)
	}

	// Save analysis to file
	analysisData, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal analysis: %v", err)
	}

	analysisPath := filepath.Join(tempDir, "test_analysis.json")
	if err := os.WriteFile(analysisPath, analysisData, 0644); err != nil {
		t.Fatalf("Failed to write analysis file: %v", err)
	}

	// Test loading from file
	loadedDeck, err := builder.BuildDeckFromFile(analysisPath)
	if err != nil {
		t.Fatalf("Failed to build deck from file: %v", err)
	}

	// Compare decks
	if len(loadedDeck.Deck) != len(deck.Deck) {
		t.Errorf("Deck size mismatch: %d vs %d", len(loadedDeck.Deck), len(deck.Deck))
	}

	// Test saving deck
	deckPath, err := builder.SaveDeck(deck, tempDir, "#TEST123")
	if err != nil {
		t.Fatalf("Failed to save deck: %v", err)
	}

	// Verify deck file exists
	if _, err := os.Stat(deckPath); os.IsNotExist(err) {
		t.Errorf("Deck file was not created: %s", deckPath)
	}

	// Verify filename format
	expectedPattern := "*_deck_TEST123.json"
	matched, _ := filepath.Match(expectedPattern, filepath.Base(deckPath))
	if !matched {
		t.Errorf("Deck filename doesn't match expected pattern: %s", filepath.Base(deckPath))
	}
}

func TestBuilder_ScoreCard(t *testing.T) {
	builder := NewBuilder("testdata")

	// Create role variables to take addresses of
	supportRole := RoleSupport
	winConditionRole := RoleWinCondition

	tests := []struct {
		name     string
		level    int
		maxLevel int
		rarity   string
		elixir   int
		role     *CardRole
		expected float64
	}{
		{
			name:     "Max Level Common",
			level:    13,
			maxLevel: 13,
			rarity:   "Common",
			elixir:   3,
			role:     &supportRole,
			expected: 1.0*1.2*1.0 + (1.0-0)*0.15 + 0.05, // ~1.4
		},
		{
			name:     "Low Level Legendary",
			level:    1,
			maxLevel: 5,
			rarity:   "Legendary",
			elixir:   5,
			role:     &winConditionRole,
			expected: 0.2*1.2*1.15 + (1.0-2/9)*0.15 + 0.05, // ~0.53
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			score := builder.scoreCard("TestCard", test.level, test.maxLevel, test.rarity, test.elixir, test.role, 0)
			if score < test.expected-0.1 || score > test.expected+0.1 {
				t.Errorf("Score %f not close to expected %f", score, test.expected)
			}
		})
	}
}

func TestBuilder_PickBest(t *testing.T) {
	builder := NewBuilder("testdata")

	// Create role variables to take addresses of
	winConditionRole := RoleWinCondition
	spellBigRole := RoleSpellBig
	buildingRole := RoleBuilding

	candidates := []*CardCandidate{
		{Name: "Hog Rider", Score: 2.5, Role: &winConditionRole},
		{Name: "Giant", Score: 2.1, Role: &winConditionRole},
		{Name: "Fireball", Score: 1.8, Role: &spellBigRole},
		{Name: "Cannon", Score: 1.5, Role: &buildingRole},
	}

	used := make(map[string]bool)

	// Test picking win condition
	winCondition := builder.pickBest(RoleWinCondition, candidates, used)
	if winCondition == nil {
		t.Error("Expected to find win condition")
	}

	if winCondition.Name != "Hog Rider" {
		t.Errorf("Expected Hog Rider, got %s", winCondition.Name)
	}

	used["Hog Rider"] = true

	// Test picking building
	building := builder.pickBest(RoleBuilding, candidates, used)
	if building == nil {
		t.Error("Expected to find building")
	}

	if building.Name != "Cannon" {
		t.Errorf("Expected Cannon, got %s", building.Name)
	}
}

func TestBuilder_ResolveElixir(t *testing.T) {
	builder := NewBuilder("testdata")

	tests := []struct {
		name     string
		data     CardLevelData
		expected int
	}{
		{
			name:     "Elixir in data",
			data:     CardLevelData{Elixir: 5},
			expected: 5,
		},
		{
			name:     "Fallback elixir",
			data:     CardLevelData{},
			expected: 4, // Default fallback
		},
		{
			name:     "Known card fallback",
			data:     CardLevelData{},
			expected: 4, // Hog Rider fallback
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.name == "Known card fallback" {
				result := builder.resolveElixir("Hog Rider", test.data)
				if result != test.expected {
					t.Errorf("Expected %d, got %d", test.expected, result)
				}
			} else {
				result := builder.resolveElixir("Unknown Card", test.data)
				if result != test.expected {
					t.Errorf("Expected %d, got %d", test.expected, result)
				}
			}
		})
	}
}

func TestBuilder_InferRole(t *testing.T) {
	builder := NewBuilder("testdata")

	// Create role variables to take addresses of
	winConditionRole := RoleWinCondition
	buildingRole := RoleBuilding

	tests := []struct {
		name     string
		cardName string
		expected *CardRole
	}{
		{
			name:     "Win condition",
			cardName: "Hog Rider",
			expected: &winConditionRole,
		},
		{
			name:     "Building",
			cardName: "Cannon",
			expected: &buildingRole,
		},
		{
			name:     "Unknown card",
			cardName: "Unknown Card",
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			role := builder.inferRole(test.cardName)
			if test.expected == nil && role != nil {
				t.Errorf("Expected nil role, got %v", role)
			} else if test.expected != nil && role == nil {
				t.Errorf("Expected role %v, got nil", *test.expected)
			} else if test.expected != nil && role != nil && *role != *test.expected {
				t.Errorf("Expected role %v, got %v", *test.expected, *role)
			}
		})
	}
}

// Benchmark tests
func BenchmarkBuilder_BuildDeckFromAnalysis(b *testing.B) {
	builder := NewBuilder("testdata")

	// Create test analysis with many cards
	analysis := CardAnalysis{
		CardLevels: make(map[string]CardLevelData),
	}

	// Add many cards to simulate real scenario
	cards := []string{
		"Hog Rider", "Fireball", "Zap", "Cannon", "Archers", "Knight",
		"Skeletons", "Valkyrie", "Baby Dragon", "Musketeer", "Wizard",
		"Minions", "Goblin Barrel", "Poison", "Log", "Ice Spirit",
		"Bats", "Spear Goblins", "Royal Giant", "Giant", "P.E.K.K.A",
	}

	for _, card := range cards {
		analysis.CardLevels[card] = CardLevelData{
			Level:    8,
			MaxLevel: 13,
			Rarity:   "Rare",
			Elixir:   3 + len(strings.Split(card, ""))%5, // Variable elixir
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := builder.BuildDeckFromAnalysis(analysis)
		if err != nil {
			b.Fatalf("Failed to build deck: %v", err)
		}
	}
}
