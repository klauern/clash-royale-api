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

func TestBuilder_CalculateEvolutionBonus(t *testing.T) {
	tests := []struct {
		name         string
		cardName     string
		level        int
		maxLevel     int
		maxEvoLevel  int
		unlocked     map[string]bool
		expectedMin  float64
		expectedMax  float64
	}{
		{
			name:         "evolution not unlocked",
			cardName:     "Knight",
			level:        14,
			maxLevel:     14,
			maxEvoLevel:  3,
			unlocked:     map[string]bool{},
			expectedMin:  0.0,
			expectedMax:  0.0,
		},
		{
			name:         "no evolution support",
			cardName:     "P.E.K.K.A",
			level:        12,
			maxLevel:     14,
			maxEvoLevel:  0,
			unlocked:     map[string]bool{"P.E.K.K.A": true},
			expectedMin:  0.0,
			expectedMax:  0.0,
		},
		{
			name:         "single evolution at max level",
			cardName:     "Archers",
			level:        14,
			maxLevel:     14,
			maxEvoLevel:  1,
			unlocked:     map[string]bool{"Archers": true},
			expectedMin:  0.24,
			expectedMax:  0.26,
		},
		{
			name:         "single evolution at mid level",
			cardName:     "Bomber",
			level:        11,
			maxLevel:     14,
			maxEvoLevel:  1,
			unlocked:     map[string]bool{"Bomber": true},
			expectedMin:  0.17,
			expectedMax:  0.18,
		},
		{
			name:         "multi-evolution (3 levels) at max level",
			cardName:     "Knight",
			level:        14,
			maxLevel:     14,
			maxEvoLevel:  3,
			unlocked:     map[string]bool{"Knight": true},
			expectedMin:  0.34,
			expectedMax:  0.36,
		},
		{
			name:         "multi-evolution (3 levels) at level 10",
			cardName:     "Musketeer",
			level:        10,
			maxLevel:     14,
			maxEvoLevel:  3,
			unlocked:     map[string]bool{"Musketeer": true},
			expectedMin:  0.21,
			expectedMax:  0.22,
		},
		{
			name:         "multi-evolution (2 levels) at max level",
			cardName:     "Giant",
			level:        14,
			maxLevel:     14,
			maxEvoLevel:  2,
			unlocked:     map[string]bool{"Giant": true},
			expectedMin:  0.29,
			expectedMax:  0.31,
		},
		{
			name:         "evolution at very low level",
			cardName:     "Knight",
			level:        1,
			maxLevel:     14,
			maxEvoLevel:  3,
			unlocked:     map[string]bool{"Knight": true},
			expectedMin:  0.00,
			expectedMax:  0.02,
		},
		{
			name:         "evolution at level 0",
			cardName:     "Test",
			level:        0,
			maxLevel:     14,
			maxEvoLevel:  1,
			unlocked:     map[string]bool{"Test": true},
			expectedMin:  0.0,
			expectedMax:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder("testdata")
			builder.unlockedEvolutions = tt.unlocked

			bonus := builder.calculateEvolutionBonus(tt.cardName, tt.level, tt.maxLevel, tt.maxEvoLevel)

			if bonus < tt.expectedMin || bonus > tt.expectedMax {
				t.Errorf("Evolution bonus = %f, want range [%f, %f]",
					bonus, tt.expectedMin, tt.expectedMax)
			}

			// Verify bonus is non-zero when evolution is unlocked
			if tt.unlocked[tt.cardName] && tt.maxEvoLevel > 0 && tt.level > 0 {
				if bonus <= 0 {
					t.Errorf("Expected positive bonus for unlocked evolution, got %f", bonus)
				}
			}
		})
	}
}

func TestBuilder_CalculateEvolutionBonus_FormulaVerification(t *testing.T) {
	// Test the exact formula: 0.25 * (level/maxLevel)^1.5 * (1 + 0.2*(maxEvoLevel-1))
	builder := NewBuilder("testdata")
	builder.unlockedEvolutions = map[string]bool{"Test": true}

	// Test case: Knight at level 14/14 with maxEvoLevel=3
	// Expected: 0.25 * (14/14)^1.5 * (1 + 0.2*2) = 0.25 * 1.0 * 1.4 = 0.35
	bonus := builder.calculateEvolutionBonus("Test", 14, 14, 3)
	expected := 0.35
	tolerance := 0.01

	if bonus < expected-tolerance || bonus > expected+tolerance {
		t.Errorf("Formula test failed: got %f, want %f (±%f)", bonus, expected, tolerance)
	}

	// Test case: Archers at level 14/14 with maxEvoLevel=1
	// Expected: 0.25 * (14/14)^1.5 * (1 + 0.2*0) = 0.25 * 1.0 * 1.0 = 0.25
	bonus = builder.calculateEvolutionBonus("Test", 14, 14, 1)
	expected = 0.25

	if bonus < expected-tolerance || bonus > expected+tolerance {
		t.Errorf("Formula test failed: got %f, want %f (±%f)", bonus, expected, tolerance)
	}

	// Test case: Card at level 10/14 with maxEvoLevel=2
	// Expected: 0.25 * (10/14)^1.5 * (1 + 0.2*1) = 0.25 * 0.60368 * 1.2 ≈ 0.181
	bonus = builder.calculateEvolutionBonus("Test", 10, 14, 2)
	expected = 0.181
	tolerance = 0.01

	if bonus < expected-tolerance || bonus > expected+tolerance {
		t.Errorf("Formula test failed: got %f, want %f (±%f)", bonus, expected, tolerance)
	}
}

func TestBuilder_CalculateEvolutionBonus_LevelScaling(t *testing.T) {
	builder := NewBuilder("testdata")
	builder.unlockedEvolutions = map[string]bool{"Test": true}

	// Test that bonus increases with level for the same maxLevel
	levels := []int{1, 5, 10, 12, 14}
	var prevBonus float64

	for i, level := range levels {
		bonus := builder.calculateEvolutionBonus("Test", level, 14, 1)

		if i > 0 && bonus <= prevBonus {
			t.Errorf("Expected bonus to increase with level: level %d bonus %f <= level %d bonus %f",
				level, bonus, levels[i-1], prevBonus)
		}

		prevBonus = bonus
		t.Logf("Level %d/14: bonus = %f", level, bonus)
	}
}

func TestBuilder_CalculateEvolutionBonus_EvolutionLevelScaling(t *testing.T) {
	builder := NewBuilder("testdata")
	builder.unlockedEvolutions = map[string]bool{"Test": true}

	// Test that bonus increases with maxEvolutionLevel for the same card level
	evoLevels := []int{1, 2, 3}
	var prevBonus float64

	for i, evoLevel := range evoLevels {
		bonus := builder.calculateEvolutionBonus("Test", 14, 14, evoLevel)

		if i > 0 && bonus <= prevBonus {
			t.Errorf("Expected bonus to increase with evo level: evo %d bonus %f <= evo %d bonus %f",
				evoLevel, bonus, evoLevels[i-1], prevBonus)
		}

		prevBonus = bonus
		t.Logf("MaxEvoLevel %d: bonus = %f", evoLevel, bonus)
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

func BenchmarkBuilder_CalculateEvolutionBonus(b *testing.B) {
	builder := NewBuilder("testdata")
	builder.unlockedEvolutions = map[string]bool{"Knight": true, "Archers": true, "Giant": true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.calculateEvolutionBonus("Knight", 14, 14, 3)
		builder.calculateEvolutionBonus("Archers", 13, 14, 1)
		builder.calculateEvolutionBonus("Giant", 12, 14, 2)
		builder.calculateEvolutionBonus("P.E.K.K.A", 14, 14, 0)
	}
}
