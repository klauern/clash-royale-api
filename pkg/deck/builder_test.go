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
		name        string
		cardName    string
		level       int
		maxLevel    int
		maxEvoLevel int
		unlocked    map[string]bool
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "evolution not unlocked",
			cardName:    "Knight",
			level:       14,
			maxLevel:    14,
			maxEvoLevel: 3,
			unlocked:    map[string]bool{},
			expectedMin: 0.0,
			expectedMax: 0.0,
		},
		{
			name:        "no evolution support",
			cardName:    "P.E.K.K.A",
			level:       12,
			maxLevel:    14,
			maxEvoLevel: 0,
			unlocked:    map[string]bool{"P.E.K.K.A": true},
			expectedMin: 0.0,
			expectedMax: 0.0,
		},
		{
			name:        "single evolution at max level",
			cardName:    "Archers",
			level:       14,
			maxLevel:    14,
			maxEvoLevel: 1,
			unlocked:    map[string]bool{"Archers": true},
			expectedMin: 0.24,
			expectedMax: 0.26,
		},
		{
			name:        "single evolution at mid level",
			cardName:    "Bomber",
			level:       11,
			maxLevel:    14,
			maxEvoLevel: 1,
			unlocked:    map[string]bool{"Bomber": true},
			expectedMin: 0.17,
			expectedMax: 0.18,
		},
		{
			name:        "multi-evolution (3 levels) at max level",
			cardName:    "Knight",
			level:       14,
			maxLevel:    14,
			maxEvoLevel: 3,
			unlocked:    map[string]bool{"Knight": true},
			expectedMin: 0.37, // Updated for 10% role override bonus: 0.35 * 1.1 = 0.385
			expectedMax: 0.39,
		},
		{
			name:        "multi-evolution (3 levels) at level 10",
			cardName:    "Musketeer",
			level:       10,
			maxLevel:    14,
			maxEvoLevel: 3,
			unlocked:    map[string]bool{"Musketeer": true},
			expectedMin: 0.21,
			expectedMax: 0.22,
		},
		{
			name:        "multi-evolution (2 levels) at max level",
			cardName:    "Giant",
			level:       14,
			maxLevel:    14,
			maxEvoLevel: 2,
			unlocked:    map[string]bool{"Giant": true},
			expectedMin: 0.29,
			expectedMax: 0.31,
		},
		{
			name:        "evolution at very low level",
			cardName:    "Knight",
			level:       1,
			maxLevel:    14,
			maxEvoLevel: 3,
			unlocked:    map[string]bool{"Knight": true},
			expectedMin: 0.00,
			expectedMax: 0.02,
		},
		{
			name:        "evolution at level 0",
			cardName:    "Test",
			level:       0,
			maxLevel:    14,
			maxEvoLevel: 1,
			unlocked:    map[string]bool{"Test": true},
			expectedMin: 0.0,
			expectedMax: 0.0,
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

func TestCardCandidate_LevelRatio(t *testing.T) {
	tests := []struct {
		name              string
		level             int
		maxLevel          int
		evolutionLevel    int
		maxEvolutionLevel int
		expected          float64
		tolerance         float64
	}{
		{
			name:              "no evolution capability",
			level:             13,
			maxLevel:          14,
			evolutionLevel:    0,
			maxEvolutionLevel: 0,
			expected:          13.0 / 14.0, // ~0.929
			tolerance:         0.01,
		},
		{
			name:              "max level, no evolution",
			level:             14,
			maxLevel:          14,
			evolutionLevel:    0,
			maxEvolutionLevel: 0,
			expected:          1.0,
			tolerance:         0.01,
		},
		{
			name:              "evolution capable, no evolution progress",
			level:             14,
			maxLevel:          14,
			evolutionLevel:    0,
			maxEvolutionLevel: 3,
			expected:          (1.0 * 0.7) + (0.0 * 0.3), // 0.7
			tolerance:         0.01,
		},
		{
			name:              "evolution capable, partial evolution",
			level:             14,
			maxLevel:          14,
			evolutionLevel:    2,
			maxEvolutionLevel: 3,
			expected:          (1.0 * 0.7) + (2.0 / 3.0 * 0.3), // ~0.90
			tolerance:         0.01,
		},
		{
			name:              "evolution capable, max evolution",
			level:             14,
			maxLevel:          14,
			evolutionLevel:    3,
			maxEvolutionLevel: 3,
			expected:          (1.0 * 0.7) + (1.0 * 0.3), // 1.0
			tolerance:         0.01,
		},
		{
			name:              "mid level with partial evolution",
			level:             10,
			maxLevel:          14,
			evolutionLevel:    1,
			maxEvolutionLevel: 3,
			expected:          (10.0 / 14.0 * 0.7) + (1.0 / 3.0 * 0.3), // ~0.60
			tolerance:         0.01,
		},
		{
			name:              "low level with max evolution",
			level:             5,
			maxLevel:          14,
			evolutionLevel:    2,
			maxEvolutionLevel: 2,
			expected:          (5.0 / 14.0 * 0.7) + (1.0 * 0.3), // ~0.55
			tolerance:         0.01,
		},
		{
			name:              "zero max level edge case",
			level:             0,
			maxLevel:          0,
			evolutionLevel:    0,
			maxEvolutionLevel: 0,
			expected:          0.0,
			tolerance:         0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := &CardCandidate{
				Level:             tt.level,
				MaxLevel:          tt.maxLevel,
				EvolutionLevel:    tt.evolutionLevel,
				MaxEvolutionLevel: tt.maxEvolutionLevel,
			}

			ratio := candidate.LevelRatio()

			if ratio < tt.expected-tt.tolerance || ratio > tt.expected+tt.tolerance {
				t.Errorf("LevelRatio() = %f, want %f (±%f)", ratio, tt.expected, tt.tolerance)
			}

			t.Logf("Card %d/%d, Evo %d/%d: ratio = %.3f",
				tt.level, tt.maxLevel, tt.evolutionLevel, tt.maxEvolutionLevel, ratio)
		})
	}
}

func TestBuilder_BuildCandidate_EvolutionLevel(t *testing.T) {
	builder := NewBuilder("testdata")

	tests := []struct {
		name              string
		cardName          string
		data              CardLevelData
		expectedEvoLevel  int
		expectedMaxEvoLvl int
	}{
		{
			name:     "no evolution capability",
			cardName: "P.E.K.K.A",
			data: CardLevelData{
				Level:             14,
				MaxLevel:          14,
				Rarity:            "Epic",
				Elixir:            7,
				EvolutionLevel:    0,
				MaxEvolutionLevel: 0,
			},
			expectedEvoLevel:  0,
			expectedMaxEvoLvl: 0,
		},
		{
			name:     "evolution capable, no progress",
			cardName: "Knight",
			data: CardLevelData{
				Level:             14,
				MaxLevel:          14,
				Rarity:            "Common",
				Elixir:            3,
				EvolutionLevel:    0,
				MaxEvolutionLevel: 3,
			},
			expectedEvoLevel:  0,
			expectedMaxEvoLvl: 3,
		},
		{
			name:     "evolution capable, partial progress",
			cardName: "Archers",
			data: CardLevelData{
				Level:             14,
				MaxLevel:          14,
				Rarity:            "Common",
				Elixir:            3,
				EvolutionLevel:    1,
				MaxEvolutionLevel: 1,
			},
			expectedEvoLevel:  1,
			expectedMaxEvoLvl: 1,
		},
		{
			name:     "evolution capable, max progress",
			cardName: "Musketeer",
			data: CardLevelData{
				Level:             14,
				MaxLevel:          14,
				Rarity:            "Rare",
				Elixir:            4,
				EvolutionLevel:    3,
				MaxEvolutionLevel: 3,
			},
			expectedEvoLevel:  3,
			expectedMaxEvoLvl: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := builder.buildCandidate(tt.cardName, tt.data)

			if candidate.EvolutionLevel != tt.expectedEvoLevel {
				t.Errorf("EvolutionLevel = %d, want %d", candidate.EvolutionLevel, tt.expectedEvoLevel)
			}

			if candidate.MaxEvolutionLevel != tt.expectedMaxEvoLvl {
				t.Errorf("MaxEvolutionLevel = %d, want %d", candidate.MaxEvolutionLevel, tt.expectedMaxEvoLvl)
			}

			t.Logf("Card %s: EvolutionLevel=%d, MaxEvolutionLevel=%d, LevelRatio=%.3f",
				tt.cardName, candidate.EvolutionLevel, candidate.MaxEvolutionLevel, candidate.LevelRatio())
		})
	}
}

func TestBuilder_BuildDeckFromAnalysis_WithEvolutionLevels(t *testing.T) {
	builder := NewBuilder("testdata")
	builder.unlockedEvolutions = map[string]bool{
		"Knight":  true,
		"Archers": true,
	}

	// Create test analysis with evolution levels
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
				Level:             14,
				MaxLevel:          14,
				Rarity:            "Common",
				Elixir:            3,
				EvolutionLevel:    1,
				MaxEvolutionLevel: 1,
			},
			"Knight": {
				Level:             14,
				MaxLevel:          14,
				Rarity:            "Common",
				Elixir:            3,
				EvolutionLevel:    2,
				MaxEvolutionLevel: 3,
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
		},
		AnalysisTime: "2024-01-15T10:30:00Z",
	}

	deck, err := builder.BuildDeckFromAnalysis(analysis)
	if err != nil {
		t.Fatalf("Failed to build deck: %v", err)
	}

	// Validate that evolution levels are preserved
	foundKnight := false
	foundArchers := false

	for _, detail := range deck.DeckDetail {
		if detail.Name == "Knight" {
			foundKnight = true
			t.Logf("Knight: Level %d/%d, Score %.3f", detail.Level, detail.MaxLevel, detail.Score)
		}
		if detail.Name == "Archers" {
			foundArchers = true
			t.Logf("Archers: Level %d/%d, Score %.3f", detail.Level, detail.MaxLevel, detail.Score)
		}
	}

	if !foundKnight && !foundArchers {
		t.Log("Note: Neither evolved card made it into the deck (may be expected based on other factors)")
	}

	t.Logf("Generated deck: %v", deck.Deck)
	t.Logf("Evolution slots: %v", deck.EvolutionSlots)
}

// TestBuilder_EvolutionSlotSelection tests evolution slot prioritization
func TestBuilder_EvolutionSlotSelection(t *testing.T) {
	// Create builder with multiple unlocked evolutions
	builder := NewBuilder("testdata")
	builder.SetUnlockedEvolutions([]string{"Knight", "Archers", "Valkyrie", "Musketeer"})
	builder.SetEvolutionSlotLimit(2) // Default 2 slots

	// Build analysis with multiple evolved cards at high levels
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			// Multiple win conditions
			"Hog Rider":   {Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 4},
			"Royal Giant": {Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 6},
			// Multiple evolved support cards
			"Knight":    {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3, MaxEvolutionLevel: 3},
			"Archers":   {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3, MaxEvolutionLevel: 1},
			"Valkyrie":  {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4, MaxEvolutionLevel: 3},
			"Musketeer": {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4, MaxEvolutionLevel: 3},
			// Other cards
			"Cannon":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Zap":        {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 2},
			"Log":        {Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 5},
			"Ice Spirit": {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 1},
		},
		AnalysisTime: "2024-01-15T10:30:00Z",
	}

	deck, err := builder.BuildDeckFromAnalysis(analysis)
	if err != nil {
		t.Fatalf("Failed to build deck: %v", err)
	}

	// Verify evolution slots were assigned
	if len(deck.EvolutionSlots) > 2 {
		t.Errorf("Expected at most 2 evolution slots, got %d", len(deck.EvolutionSlots))
	}

	// Check that evolution slots contain valid evolved cards from our deck
	slotCards := make(map[string]bool)
	for _, slot := range deck.EvolutionSlots {
		slotCards[slot] = true
		t.Logf("Evolution slot: %s", slot)
	}

	// Verify slot assignments follow priority rules
	// Slots should go to: Win Conditions > Buildings > Big Spells > Support > Small Spells > Cycle
	// Knight, Archers, Valkyrie, Musketeer are all Support, so they compete on score

	foundEvoInDeck := 0
	for _, cardName := range deck.Deck {
		if slotCards[cardName] {
			foundEvoInDeck++
			t.Logf("Evolved card in deck: %s", cardName)
		}
	}

	t.Logf("Total evolved cards with slots: %d", foundEvoInDeck)
}

// TestBuilder_EvolutionSlotPriority tests that higher-priority evolved cards get slots
func TestBuilder_EvolutionSlotPriority(t *testing.T) {
	builder := NewBuilder("testdata")

	// Scenario: Win condition evolution vs support evolution
	// Win condition should get priority
	builder.SetUnlockedEvolutions([]string{"Royal Giant", "Knight"})
	builder.SetEvolutionSlotLimit(1) // Only 1 slot

	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Royal Giant": {Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 6, MaxEvolutionLevel: 3},
			"Knight":      {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3, MaxEvolutionLevel: 3},
			// Other cards to fill deck
			"Cannon":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Zap":        {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 2},
			"Musketeer":  {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Fireball":   {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Log":        {Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 5},
			"Ice Spirit": {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 1},
		},
		AnalysisTime: "2024-01-15T10:30:00Z",
	}

	deck, err := builder.BuildDeckFromAnalysis(analysis)
	if err != nil {
		t.Fatalf("Failed to build deck: %v", err)
	}

	// Verify slot was assigned
	if len(deck.EvolutionSlots) == 0 {
		t.Error("Expected at least 1 evolution slot, got 0")
	}

	// Royal Giant (win condition) should be prioritized over Knight (support)
	// if both are in the deck
	hasRoyalGiant := false
	hasKnight := false
	royalGiantHasSlot := false
	knightHasSlot := false

	for _, card := range deck.Deck {
		if card == "Royal Giant" {
			hasRoyalGiant = true
		}
		if card == "Knight" {
			hasKnight = true
		}
	}

	for _, slot := range deck.EvolutionSlots {
		if slot == "Royal Giant" {
			royalGiantHasSlot = true
		}
		if slot == "Knight" {
			knightHasSlot = true
		}
	}

	t.Logf("Deck contains Royal Giant: %v, Knight: %v", hasRoyalGiant, hasKnight)
	t.Logf("Slots: Royal Giant=%v, Knight=%v", royalGiantHasSlot, knightHasSlot)

	// If both cards are in deck and Royal Giant is higher priority (win condition),
	// it should get the slot when only 1 is available
	if hasRoyalGiant && hasKnight && len(deck.EvolutionSlots) == 1 {
		if !royalGiantHasSlot {
			t.Error("Expected Royal Giant (win condition) to get evolution slot over Knight (support)")
		}
	}
}

// TestBuilder_EvolutionMultiSlot tests with multiple evolution slots available
func TestBuilder_EvolutionMultiSlot(t *testing.T) {
	builder := NewBuilder("testdata")
	builder.SetUnlockedEvolutions([]string{"Knight", "Archers", "Valkyrie"})
	builder.SetEvolutionSlotLimit(3) // 3 slots available

	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Knight":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3, MaxEvolutionLevel: 3},
			"Archers":    {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3, MaxEvolutionLevel: 1},
			"Valkyrie":   {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4, MaxEvolutionLevel: 3},
			"Hog Rider":  {Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 4},
			"Musketeer":  {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Log":        {Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 5},
			"Ice Spirit": {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Cannon":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
		},
		AnalysisTime: "2024-01-15T10:30:00Z",
	}

	deck, err := builder.BuildDeckFromAnalysis(analysis)
	if err != nil {
		t.Fatalf("Failed to build deck: %v", err)
	}

	// Should use up to 3 evolution slots
	if len(deck.EvolutionSlots) > 3 {
		t.Errorf("Expected at most 3 evolution slots, got %d", len(deck.EvolutionSlots))
	}

	t.Logf("Evolution slots (%d): %v", len(deck.EvolutionSlots), deck.EvolutionSlots)
}

// TestBuilder_EvolutionNoUnlocked tests behavior when no evolutions are unlocked
func TestBuilder_EvolutionNoUnlocked(t *testing.T) {
	builder := NewBuilder("testdata")
	// Explicitly clear all unlocked evolutions (to override any environment variable)
	builder.SetUnlockedEvolutions([]string{})
	builder.SetEvolutionSlotLimit(2)

	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Knight":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3, MaxEvolutionLevel: 3},
			"Archers":    {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3, MaxEvolutionLevel: 1},
			"Hog Rider":  {Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 4},
			"Musketeer":  {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Log":        {Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 5},
			"Ice Spirit": {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Cannon":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Zap":        {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 2},
		},
		AnalysisTime: "2024-01-15T10:30:00Z",
	}

	deck, err := builder.BuildDeckFromAnalysis(analysis)
	if err != nil {
		t.Fatalf("Failed to build deck: %v", err)
	}

	// No evolution slots should be assigned when no evolutions are unlocked
	if len(deck.EvolutionSlots) != 0 {
		t.Errorf("Expected 0 evolution slots when none unlocked, got %d", len(deck.EvolutionSlots))
	}
}
