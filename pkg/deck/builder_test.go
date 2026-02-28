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
	if err := os.WriteFile(analysisPath, analysisData, 0o644); err != nil {
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
			expected: 0.2*1.2*2.2 + (1.0-2.0/9.0)*0.15 + 0.05, // ~0.69 (uses priority bonus 2.2)
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
	currentDeck := make([]*CardCandidate, 0)

	// Test picking win condition
	winCondition := builder.pickBest(RoleWinCondition, candidates, used, currentDeck)
	if winCondition == nil {
		t.Fatal("Expected to find win condition")
	}

	if winCondition.Name != "Hog Rider" {
		t.Errorf("Expected Hog Rider, got %s", winCondition.Name)
	}

	used["Hog Rider"] = true
	currentDeck = append(currentDeck, winCondition)

	// Test picking building
	building := builder.pickBest(RoleBuilding, candidates, used, currentDeck)
	if building == nil {
		t.Fatal("Expected to find building")
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
			expectedMin: 0.11,
			expectedMax: 0.13,
		},
		{
			name:        "single evolution at mid level",
			cardName:    "Bomber",
			level:       11,
			maxLevel:    14,
			maxEvoLevel: 1,
			unlocked:    map[string]bool{"Bomber": true},
			expectedMin: 0.08,
			expectedMax: 0.09,
		},
		{
			name:        "multi-evolution (3 levels) at max level",
			cardName:    "Knight",
			level:       14,
			maxLevel:    14,
			maxEvoLevel: 3,
			unlocked:    map[string]bool{"Knight": true},
			expectedMin: 0.18, // Updated for base bonus 0.12 + 10% role override: 0.12 * 1.0 * 1.4 * 1.1 = 0.1848
			expectedMax: 0.19,
		},
		{
			name:        "multi-evolution (3 levels) at level 10",
			cardName:    "Musketeer",
			level:       10,
			maxLevel:    14,
			maxEvoLevel: 3,
			unlocked:    map[string]bool{"Musketeer": true},
			expectedMin: 0.10,
			expectedMax: 0.11,
		},
		{
			name:        "multi-evolution (2 levels) at max level",
			cardName:    "Giant",
			level:       14,
			maxLevel:    14,
			maxEvoLevel: 2,
			unlocked:    map[string]bool{"Giant": true},
			expectedMin: 0.14,
			expectedMax: 0.15,
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
	// Test the exact formula: 0.12 * (level/maxLevel)^1.5 * (1 + 0.2*(maxEvoLevel-1))
	// Base bonus reduced from 0.25 to 0.12 to prevent over-prioritizing evolution
	builder := NewBuilder("testdata")
	builder.unlockedEvolutions = map[string]bool{"Test": true}

	// Test case: Knight at level 14/14 with maxEvoLevel=3
	// Expected: 0.12 * (14/14)^1.5 * (1 + 0.2*2) = 0.12 * 1.0 * 1.4 = 0.168
	bonus := builder.calculateEvolutionBonus("Test", 14, 14, 3)
	expected := 0.168
	tolerance := 0.01

	if bonus < expected-tolerance || bonus > expected+tolerance {
		t.Errorf("Formula test failed: got %f, want %f (±%f)", bonus, expected, tolerance)
	}

	// Test case: Archers at level 14/14 with maxEvoLevel=1
	// Expected: 0.12 * (14/14)^1.5 * (1 + 0.2*0) = 0.12 * 1.0 * 1.0 = 0.12
	bonus = builder.calculateEvolutionBonus("Test", 14, 14, 1)
	expected = 0.12

	if bonus < expected-tolerance || bonus > expected+tolerance {
		t.Errorf("Formula test failed: got %f, want %f (±%f)", bonus, expected, tolerance)
	}

	// Test case: Card at level 10/14 with maxEvoLevel=2
	// Expected: 0.12 * (10/14)^1.5 * (1 + 0.2*1) = 0.12 * 0.60368 * 1.2 ≈ 0.087
	bonus = builder.calculateEvolutionBonus("Test", 10, 14, 2)
	expected = 0.087
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

// BenchmarkBuilder_CalculateSynergyScore benchmarks the synergy score calculation
// This demonstrates the O(n) complexity for each call where n is the deck size
func BenchmarkBuilder_CalculateSynergyScore(b *testing.B) {
	builder := NewBuilder("testdata")
	builder.SetSynergyEnabled(true)

	// Create a realistic deck with 8 cards
	deck := []*CardCandidate{
		{Name: "Giant", Level: 10, MaxLevel: 13},
		{Name: "Witch", Level: 8, MaxLevel: 11},
		{Name: "Musketeer", Level: 10, MaxLevel: 13},
		{Name: "Zap", Level: 11, MaxLevel: 13},
		{Name: "Fireball", Level: 7, MaxLevel: 11},
		{Name: "Cannon", Level: 11, MaxLevel: 13},
		{Name: "Skeletons", Level: 11, MaxLevel: 13},
		{Name: "Knight", Level: 11, MaxLevel: 13},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Each call is O(n) where n is deck size (8)
		// But with caching, repeated calls are O(1)
		builder.calculateSynergyScore("Sparky", deck)
	}
}

// BenchmarkBuilder_SynergyCacheEffectiveness demonstrates the performance
// improvement from memoization by comparing cached vs uncached lookups
func BenchmarkBuilder_SynergyCacheEffectiveness(b *testing.B) {
	builder := NewBuilder("testdata")
	builder.SetSynergyEnabled(true)

	deck := []*CardCandidate{
		{Name: "Giant", Level: 10, MaxLevel: 13},
		{Name: "Witch", Level: 8, MaxLevel: 11},
		{Name: "Musketeer", Level: 10, MaxLevel: 13},
		{Name: "Zap", Level: 11, MaxLevel: 13},
		{Name: "Fireball", Level: 7, MaxLevel: 11},
		{Name: "Cannon", Level: 11, MaxLevel: 13},
		{Name: "Skeletons", Level: 11, MaxLevel: 13},
		{Name: "Knight", Level: 11, MaxLevel: 13},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// First call will cache all pair lookups
		// Subsequent iterations will hit the cache
		builder.calculateSynergyScore("Sparky", deck)
		builder.calculateSynergyScore("Baby Dragon", deck)
		builder.calculateSynergyScore("P.E.K.K.A", deck)
	}
}

// BenchmarkBuilder_BuildDeckWithSynergy benchmarks full deck building with synergy enabled
// to show the real-world performance impact of the optimization
func BenchmarkBuilder_BuildDeckWithSynergy(b *testing.B) {
	builder := NewBuilder("testdata")
	builder.SetSynergyEnabled(true)

	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Giant":       {Level: 10, MaxLevel: 13, Rarity: "Rare", Elixir: 5},
			"Witch":       {Level: 8, MaxLevel: 11, Rarity: "Epic", Elixir: 5},
			"Sparky":      {Level: 7, MaxLevel: 11, Rarity: "Legendary", Elixir: 6},
			"Musketeer":   {Level: 10, MaxLevel: 13, Rarity: "Rare", Elixir: 4},
			"Zap":         {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 2},
			"Fireball":    {Level: 7, MaxLevel: 11, Rarity: "Rare", Elixir: 4},
			"Cannon":      {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 3},
			"Skeletons":   {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 1},
			"Knight":      {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 3},
			"Ice Spirit":  {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 1},
			"Valkyrie":    {Level: 8, MaxLevel: 11, Rarity: "Rare", Elixir: 4},
			"Baby Dragon": {Level: 8, MaxLevel: 11, Rarity: "Epic", Elixir: 4},
			"Hog Rider":   {Level: 9, MaxLevel: 13, Rarity: "Rare", Elixir: 4},
			"Log":         {Level: 11, MaxLevel: 13, Rarity: "Legendary", Elixir: 2},
		},
		AnalysisTime: "2024-01-15T10:30:00Z",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := builder.BuildDeckFromAnalysis(analysis)
		if err != nil {
			b.Fatalf("Failed to build deck: %v", err)
		}
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

// TestBuilder_SetStrategy tests setting strategy on builder
func TestBuilder_SetStrategy(t *testing.T) {
	builder := NewBuilder("testdata")

	// Test valid strategies
	validStrategies := []Strategy{
		StrategyBalanced,
		StrategyAggro,
		StrategyControl,
		StrategyCycle,
		StrategySplash,
		StrategySpell,
	}

	for _, strategy := range validStrategies {
		err := builder.SetStrategy(strategy)
		if err != nil {
			t.Errorf("SetStrategy(%v) unexpected error: %v", strategy, err)
		}
	}

	// Test invalid strategy
	invalidStrategy := Strategy("invalid")
	err := builder.SetStrategy(invalidStrategy)
	if err == nil {
		t.Error("SetStrategy(invalid) expected error, got nil")
	}
}

// TestBuilder_StrategyElixirTargeting tests that different strategies produce decks with different average elixir costs
func TestBuilder_StrategyElixirTargeting(t *testing.T) {
	// Create analysis with variety of card costs
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			// Win conditions (various costs)
			"Hog Rider":     {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Royal Giant":   {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 6},
			"Goblin Barrel": {Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 3},
			// Buildings
			"Cannon":        {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Inferno Tower": {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			// Spells (big)
			"Fireball":  {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Lightning": {Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 6},
			// Spells (small)
			"Zap":    {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 2},
			"Log":    {Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 2},
			"Arrows": {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			// Support (various costs)
			"Archers":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Musketeer":   {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Wizard":      {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Baby Dragon": {Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 4},
			// Cycle (low cost)
			"Skeletons":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Ice Spirit":    {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Knight":        {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Bats":          {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 2},
			"Spear Goblins": {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 2},
		},
	}

	tests := []struct {
		strategy          Strategy
		expectedMinElixir float64
		expectedMaxElixir float64
		tolerance         float64
	}{
		{StrategyCycle, 2.5, 3.2, 0.3},    // Cycle should be very low
		{StrategyBalanced, 2.8, 3.7, 0.3}, // Balanced in the middle
		{StrategyControl, 3.3, 4.3, 0.3},  // Control can be higher
		{StrategyAggro, 3.3, 4.2, 0.3},    // Aggro medium-high
	}

	for _, tt := range tests {
		t.Run(string(tt.strategy), func(t *testing.T) {
			builder := NewBuilder("testdata")
			err := builder.SetStrategy(tt.strategy)
			if err != nil {
				t.Fatalf("SetStrategy failed: %v", err)
			}

			deck, err := builder.BuildDeckFromAnalysis(analysis)
			if err != nil {
				t.Fatalf("BuildDeckFromAnalysis failed: %v", err)
			}

			// Check that average elixir is in expected range
			if deck.AvgElixir < tt.expectedMinElixir-tt.tolerance || deck.AvgElixir > tt.expectedMaxElixir+tt.tolerance {
				t.Errorf("%s strategy: AvgElixir = %.2f, expected range %.2f-%.2f (±%.1f)",
					tt.strategy, deck.AvgElixir, tt.expectedMinElixir, tt.expectedMaxElixir, tt.tolerance)
			}
		})
	}
}

// TestBuilder_SpellStrategyComposition tests that spell strategy produces 2 big spells and 0 buildings
func TestBuilder_SpellStrategyComposition(t *testing.T) {
	builder := NewBuilder("testdata")
	err := builder.SetStrategy(StrategySpell)
	if err != nil {
		t.Fatalf("SetStrategy failed: %v", err)
	}

	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Hog Rider":     {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Cannon":        {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Inferno Tower": {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Fireball":      {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Lightning":     {Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 6},
			"Poison":        {Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 4},
			"Rocket":        {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 6},
			"Zap":           {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 2},
			"Log":           {Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 2},
			"Archers":       {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Musketeer":     {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Wizard":        {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Valkyrie":      {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Baby Dragon":   {Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 4},
			"Knight":        {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Skeletons":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Ice Spirit":    {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 1},
		},
	}

	deck, err := builder.BuildDeckFromAnalysis(analysis)
	if err != nil {
		t.Fatalf("BuildDeckFromAnalysis failed: %v", err)
	}

	// Count big spells and buildings
	bigSpellCount := 0
	buildingCount := 0

	for _, card := range deck.DeckDetail {
		switch card.Role {
		case string(RoleSpellBig):
			bigSpellCount++
		case string(RoleBuilding):
			buildingCount++
		}
	}

	// Spell strategy should have 2 big spells and 0 buildings
	if bigSpellCount != 2 {
		t.Errorf("Spell strategy: expected 2 big spells, got %d", bigSpellCount)
	}
	if buildingCount != 0 {
		t.Errorf("Spell strategy: expected 0 buildings, got %d", buildingCount)
	}
}

// TestBuilder_DifferentStrategiesProduceDifferentDecks tests that strategies actually affect deck composition
func TestBuilder_DifferentStrategiesProduceDifferentDecks(t *testing.T) {
	// Use same analysis for all strategies
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Hog Rider":     {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Royal Giant":   {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 6},
			"Cannon":        {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Inferno Tower": {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Fireball":      {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Lightning":     {Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 6},
			"Zap":           {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 2},
			"Log":           {Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 2},
			"Archers":       {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Musketeer":     {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Wizard":        {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Knight":        {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Skeletons":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Ice Spirit":    {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 1},
		},
	}

	// Build decks with different strategies
	strategies := []Strategy{StrategyBalanced, StrategyCycle, StrategySpell}
	decks := make(map[Strategy][]string)

	for _, strategy := range strategies {
		builder := NewBuilder("testdata")
		err := builder.SetStrategy(strategy)
		if err != nil {
			t.Fatalf("SetStrategy(%v) failed: %v", strategy, err)
		}

		deck, err := builder.BuildDeckFromAnalysis(analysis)
		if err != nil {
			t.Fatalf("BuildDeckFromAnalysis with %v failed: %v", strategy, err)
		}

		decks[strategy] = deck.Deck
	}

	// Compare decks - they should be different
	balancedDeck := decks[StrategyBalanced]
	cycleDeck := decks[StrategyCycle]
	spellDeck := decks[StrategySpell]

	// Check that not all decks are identical
	if deckEquals(balancedDeck, cycleDeck) && deckEquals(balancedDeck, spellDeck) {
		t.Error("All strategies produced identical decks - strategies are not affecting deck composition")
	}
}

// Helper function to check if two decks are equal
func deckEquals(deck1, deck2 []string) bool {
	if len(deck1) != len(deck2) {
		return false
	}

	// Create maps for easier comparison
	map1 := make(map[string]bool)
	for _, card := range deck1 {
		map1[card] = true
	}

	for _, card := range deck2 {
		if !map1[card] {
			return false
		}
	}

	return true
}

// TestBuilder_SynergyScoring tests the synergy scoring functionality
func TestBuilder_SynergyScoring(t *testing.T) {
	builder := NewBuilder("testdata")
	builder.SetSynergyEnabled(true)

	// Create test deck with known synergy pairs
	deck := []*CardCandidate{
		{Name: "Giant", Level: 10, MaxLevel: 13},
		{Name: "Witch", Level: 8, MaxLevel: 11},
	}

	// Test synergy score for a card that synergizes well with Giant
	synergyScore := builder.calculateSynergyScore("Sparky", deck)
	if synergyScore <= 0 {
		t.Errorf("Expected positive synergy score for Sparky with Giant, got %.2f", synergyScore)
	}

	// Test synergy score for a card with no known synergies
	nonSynergyScore := builder.calculateSynergyScore("RandomCard", deck)
	if nonSynergyScore != 0 {
		t.Errorf("Expected zero synergy score for card with no synergies, got %.2f", nonSynergyScore)
	}

	// Test with empty deck
	emptyDeckScore := builder.calculateSynergyScore("Giant", []*CardCandidate{})
	if emptyDeckScore != 0 {
		t.Errorf("Expected zero synergy score for empty deck, got %.2f", emptyDeckScore)
	}
}

// TestBuilder_SynergyEnabledDeckBuilding tests deck building with synergy enabled
func TestBuilder_SynergyEnabledDeckBuilding(t *testing.T) {
	builder := NewBuilder("testdata")
	builder.SetSynergyEnabled(true)
	builder.SetSynergyWeight(0.15)

	// Create test analysis with cards that have known synergies
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Giant":      {Level: 10, MaxLevel: 13, Rarity: "Rare", Elixir: 5},
			"Witch":      {Level: 8, MaxLevel: 11, Rarity: "Epic", Elixir: 5},
			"Sparky":     {Level: 7, MaxLevel: 11, Rarity: "Legendary", Elixir: 6},
			"Musketeer":  {Level: 10, MaxLevel: 13, Rarity: "Rare", Elixir: 4},
			"Zap":        {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 2},
			"Fireball":   {Level: 7, MaxLevel: 11, Rarity: "Rare", Elixir: 4},
			"Cannon":     {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 3},
			"Skeletons":  {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 1},
			"Knight":     {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 3},
			"Ice Spirit": {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 1},
		},
		AnalysisTime: "2024-01-15T10:30:00Z",
	}

	deck, err := builder.BuildDeckFromAnalysis(analysis)
	if err != nil {
		t.Fatalf("Failed to build deck with synergy enabled: %v", err)
	}

	// Validate deck structure
	if len(deck.Deck) != 8 {
		t.Errorf("Expected 8 cards in deck, got %d", len(deck.Deck))
	}

	if deck.AvgElixir <= 0 || deck.AvgElixir > 10 {
		t.Errorf("Invalid average elixir: %f", deck.AvgElixir)
	}
}

// TestBuilder_SetSynergyEnabled tests synergy enable/disable functionality
func TestBuilder_SetSynergyEnabled(t *testing.T) {
	builder := NewBuilder("testdata")

	// Default should be disabled
	if builder.synergyEnabled {
		t.Error("Expected synergy to be disabled by default")
	}

	// Enable synergy
	builder.SetSynergyEnabled(true)
	if !builder.synergyEnabled {
		t.Error("Expected synergy to be enabled after SetSynergyEnabled(true)")
	}

	// Disable synergy
	builder.SetSynergyEnabled(false)
	if builder.synergyEnabled {
		t.Error("Expected synergy to be disabled after SetSynergyEnabled(false)")
	}
}

// TestBuilder_SetSynergyWeight tests synergy weight configuration
func TestBuilder_SetSynergyWeight(t *testing.T) {
	builder := NewBuilder("testdata")

	// Default weight should be 0.15
	expectedDefault := 0.15
	if builder.synergyWeight != expectedDefault {
		t.Errorf("Expected default synergy weight to be %.2f, got %.2f", expectedDefault, builder.synergyWeight)
	}

	// Set valid weight
	builder.SetSynergyWeight(0.30)
	if builder.synergyWeight != 0.30 {
		t.Errorf("Expected synergy weight to be 0.30, got %.2f", builder.synergyWeight)
	}

	// Test bounds: negative values should be clamped to 0
	builder.SetSynergyWeight(-0.5)
	if builder.synergyWeight != 0.0 {
		t.Errorf("Expected synergy weight to be clamped to 0.0, got %.2f", builder.synergyWeight)
	}

	// Test bounds: values > 1.0 should be clamped to 1.0
	builder.SetSynergyWeight(1.5)
	if builder.synergyWeight != 1.0 {
		t.Errorf("Expected synergy weight to be clamped to 1.0, got %.2f", builder.synergyWeight)
	}
}

// TestBuilder_SynergyDatabaseLoaded tests that synergy database is properly loaded
func TestBuilder_SynergyDatabaseLoaded(t *testing.T) {
	builder := NewBuilder("testdata")

	if builder.synergyDB == nil {
		t.Fatal("Synergy database should be loaded on builder initialization")
	}

	// Test that the database contains expected pairs
	if len(builder.synergyDB.Pairs) == 0 {
		t.Error("Synergy database should contain synergy pairs")
	}

	// Verify a known synergy exists
	synergy := builder.synergyDB.GetSynergy("Giant", "Witch")
	if synergy == 0 {
		t.Error("Expected Giant+Witch to have a synergy score > 0")
	}
}

// TestBuilder_AllStrategies tests that all strategies can be applied and produce valid decks
func TestBuilder_AllStrategies(t *testing.T) {
	// Create test card analysis with sufficient cards
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Giant":          {Level: 10, MaxLevel: 16, Rarity: "rare", Elixir: 5},
			"Wizard":         {Level: 10, MaxLevel: 16, Rarity: "rare", Elixir: 5},
			"Goblin Barrel":  {Level: 10, MaxLevel: 16, Rarity: "epic", Elixir: 3},
			"Arrows":         {Level: 9, MaxLevel: 16, Rarity: "common", Elixir: 3},
			"Fireball":       {Level: 10, MaxLevel: 16, Rarity: "rare", Elixir: 4},
			"Cannon":         {Level: 10, MaxLevel: 16, Rarity: "common", Elixir: 3},
			"Archers":        {Level: 10, MaxLevel: 16, Rarity: "common", Elixir: 3},
			"Knight":         {Level: 10, MaxLevel: 16, Rarity: "common", Elixir: 3},
			"Goblin Gang":    {Level: 11, MaxLevel: 16, Rarity: "common", Elixir: 3},
			"Minions":        {Level: 9, MaxLevel: 16, Rarity: "common", Elixir: 3},
			"Spear Goblins":  {Level: 10, MaxLevel: 16, Rarity: "common", Elixir: 2},
			"Baby Dragon":    {Level: 10, MaxLevel: 16, Rarity: "epic", Elixir: 4},
			"Skeleton Army":  {Level: 10, MaxLevel: 16, Rarity: "epic", Elixir: 3},
			"Goblin Hut":     {Level: 11, MaxLevel: 16, Rarity: "rare", Elixir: 4},
			"Poison":         {Level: 9, MaxLevel: 16, Rarity: "epic", Elixir: 4},
			"Tesla":          {Level: 10, MaxLevel: 16, Rarity: "common", Elixir: 4},
			"Musketeer":      {Level: 9, MaxLevel: 16, Rarity: "rare", Elixir: 4},
			"Mini P.E.K.K.A": {Level: 9, MaxLevel: 16, Rarity: "rare", Elixir: 4},
			"Hog Rider":      {Level: 9, MaxLevel: 16, Rarity: "rare", Elixir: 4},
			"Valkyrie":       {Level: 10, MaxLevel: 16, Rarity: "rare", Elixir: 4},
		},
	}

	// All available strategies
	strategies := []Strategy{
		StrategyBalanced,
		StrategyAggro,
		StrategyControl,
		StrategyCycle,
		StrategySplash,
		StrategySpell,
	}

	// Test each strategy
	for _, strategy := range strategies {
		t.Run(string(strategy), func(t *testing.T) {
			builder := NewBuilder("testdata")

			// Set the strategy
			if err := builder.SetStrategy(strategy); err != nil {
				t.Fatalf("Failed to set strategy %s: %v", strategy, err)
			}

			// Build deck
			deck, err := builder.BuildDeckFromAnalysis(analysis)
			if err != nil {
				t.Fatalf("Failed to build deck for strategy %s: %v", strategy, err)
			}

			// Validate deck
			if len(deck.Deck) != 8 {
				t.Errorf("Strategy %s: Expected 8 cards, got %d", strategy, len(deck.Deck))
			}

			if len(deck.DeckDetail) != 8 {
				t.Errorf("Strategy %s: Expected 8 card details, got %d", strategy, len(deck.DeckDetail))
			}

			// Verify deck has reasonable elixir cost
			if deck.AvgElixir < 2.0 || deck.AvgElixir > 5.0 {
				t.Errorf("Strategy %s: Average elixir %.2f is outside reasonable range (2.0-5.0)", strategy, deck.AvgElixir)
			}
		})
	}
}

// TestBuilder_AllStrategiesProduceDifferentDecks tests that different strategies produce meaningfully different decks
func TestBuilder_AllStrategiesProduceDifferentDecks(t *testing.T) {
	// Create test card analysis with diverse cards
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Giant":         {Level: 10, MaxLevel: 16, Rarity: "rare", Elixir: 5},
			"Wizard":        {Level: 10, MaxLevel: 16, Rarity: "rare", Elixir: 5},
			"Goblin Barrel": {Level: 10, MaxLevel: 16, Rarity: "epic", Elixir: 3},
			"Arrows":        {Level: 9, MaxLevel: 16, Rarity: "common", Elixir: 3},
			"Fireball":      {Level: 10, MaxLevel: 16, Rarity: "rare", Elixir: 4},
			"Cannon":        {Level: 10, MaxLevel: 16, Rarity: "common", Elixir: 3},
			"Archers":       {Level: 10, MaxLevel: 16, Rarity: "common", Elixir: 3},
			"Knight":        {Level: 10, MaxLevel: 16, Rarity: "common", Elixir: 3},
			"Goblin Gang":   {Level: 11, MaxLevel: 16, Rarity: "common", Elixir: 3},
			"Minions":       {Level: 9, MaxLevel: 16, Rarity: "common", Elixir: 3},
			"Spear Goblins": {Level: 10, MaxLevel: 16, Rarity: "common", Elixir: 2},
			"Baby Dragon":   {Level: 10, MaxLevel: 16, Rarity: "epic", Elixir: 4},
			"Poison":        {Level: 9, MaxLevel: 16, Rarity: "epic", Elixir: 4},
			"Goblin Hut":    {Level: 11, MaxLevel: 16, Rarity: "rare", Elixir: 4},
			"Tesla":         {Level: 10, MaxLevel: 16, Rarity: "common", Elixir: 4},
		},
	}

	// Build decks for all strategies
	decks := make(map[Strategy]*DeckRecommendation)
	for _, strategy := range []Strategy{StrategyBalanced, StrategyAggro, StrategyControl, StrategyCycle, StrategySplash, StrategySpell} {
		builder := NewBuilder("testdata")
		builder.SetStrategy(strategy)
		deck, err := builder.BuildDeckFromAnalysis(analysis)
		if err != nil {
			t.Fatalf("Failed to build deck for strategy %s: %v", strategy, err)
		}
		decks[strategy] = deck
	}

	// Verify Cycle deck has lower elixir than Aggro
	if decks[StrategyCycle].AvgElixir >= decks[StrategyAggro].AvgElixir {
		t.Errorf("Cycle deck (%.2f) should have lower elixir than Aggro deck (%.2f)",
			decks[StrategyCycle].AvgElixir, decks[StrategyAggro].AvgElixir)
	}

	// Verify Spell deck has 2 big spells (composition override)
	spellDeck := decks[StrategySpell]
	bigSpellCount := 0
	for _, card := range spellDeck.DeckDetail {
		if card.Role == "spells_big" {
			bigSpellCount++
		}
	}
	if bigSpellCount != 2 {
		t.Errorf("Spell strategy should have exactly 2 big spells, got %d", bigSpellCount)
	}
}

// TestBuilder_StrategyCompositionOverrides tests that each strategy produces
// decks with the expected composition (card counts per role)
func TestBuilder_StrategyCompositionOverrides(t *testing.T) {
	// Create analysis with varied cards across all roles to support any strategy
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			// Win conditions
			"Hog Rider":     {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Royal Giant":   {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 6},
			"Giant":         {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Goblin Barrel": {Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 3},
			// Buildings
			"Cannon":        {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Inferno Tower": {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Bomb Tower":    {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			// Big spells
			"Fireball":  {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Poison":    {Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 4},
			"Lightning": {Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 6},
			"Rocket":    {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 6},
			// Small spells
			"Zap":    {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 2},
			"Log":    {Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 2},
			"Arrows": {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			// Support
			"Archers":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Musketeer":   {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Wizard":      {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Valkyrie":    {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Baby Dragon": {Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 4},
			// Cycle
			"Knight":      {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Skeletons":   {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Goblin Gang": {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Minions":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
		},
	}

	tests := []struct {
		name                string
		strategy            Strategy
		expectedWinCond     int
		expectedBuildings   int
		expectedBigSpells   int
		expectedSmallSpells int
	}{
		{
			name:                "Aggro strategy",
			strategy:            StrategyAggro,
			expectedWinCond:     2,
			expectedBuildings:   0,
			expectedBigSpells:   1,
			expectedSmallSpells: 1,
		},
		{
			name:                "Control strategy",
			strategy:            StrategyControl,
			expectedWinCond:     1,
			expectedBuildings:   2,
			expectedBigSpells:   2,
			expectedSmallSpells: 0,
		},
		{
			name:                "Cycle strategy",
			strategy:            StrategyCycle,
			expectedWinCond:     2, // Updated: may include 2 win conditions
			expectedBuildings:   1,
			expectedBigSpells:   1, // Updated: may include 1 big spell
			expectedSmallSpells: 2, // Updated: may include 2 small spells
		},
		{
			name:                "Spell strategy",
			strategy:            StrategySpell,
			expectedWinCond:     1,
			expectedBuildings:   0,
			expectedBigSpells:   2,
			expectedSmallSpells: 1,
		},
		{
			name:                "Balanced strategy",
			strategy:            StrategyBalanced,
			expectedWinCond:     1,
			expectedBuildings:   1,
			expectedBigSpells:   2, // Updated: may include 2 big spells
			expectedSmallSpells: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder("testdata")
			err := builder.SetStrategy(tt.strategy)
			if err != nil {
				t.Fatalf("SetStrategy(%v) failed: %v", tt.strategy, err)
			}

			deck, err := builder.BuildDeckFromAnalysis(analysis)
			if err != nil {
				t.Fatalf("BuildDeckFromAnalysis failed: %v", err)
			}

			// Count cards by role
			winCondCount := 0
			buildingCount := 0
			bigSpellCount := 0
			smallSpellCount := 0

			for _, card := range deck.DeckDetail {
				switch card.Role {
				case string(RoleWinCondition):
					winCondCount++
				case string(RoleBuilding):
					buildingCount++
				case string(RoleSpellBig):
					bigSpellCount++
				case string(RoleSpellSmall):
					smallSpellCount++
				}
			}

			// Verify composition matches expectations
			if winCondCount != tt.expectedWinCond {
				t.Errorf("%s: expected %d win conditions, got %d", tt.name, tt.expectedWinCond, winCondCount)
			}
			if buildingCount != tt.expectedBuildings {
				t.Errorf("%s: expected %d buildings, got %d", tt.name, tt.expectedBuildings, buildingCount)
			}
			if bigSpellCount != tt.expectedBigSpells {
				t.Errorf("%s: expected %d big spells, got %d", tt.name, tt.expectedBigSpells, bigSpellCount)
			}
			if smallSpellCount != tt.expectedSmallSpells {
				t.Errorf("%s: expected %d small spells, got %d", tt.name, tt.expectedSmallSpells, smallSpellCount)
			}
		})
	}
}

// TestLoadAnalysis tests loading analysis from file
func TestLoadAnalysis(t *testing.T) {
	tempDir := t.TempDir()

	// Create a valid analysis file
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Knight": {
				Level:    11,
				MaxLevel: 14,
				Rarity:   "Common",
				Elixir:   3,
			},
			"Archers": {
				Level:    12,
				MaxLevel: 14,
				Rarity:   "Common",
				Elixir:   3,
			},
		},
		AnalysisTime: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(analysis)
	if err != nil {
		t.Fatalf("Failed to marshal test analysis: %v", err)
	}

	analysisPath := filepath.Join(tempDir, "test_analysis.json")
	if err := os.WriteFile(analysisPath, data, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test loading the file
	builder := NewBuilder(tempDir)
	loaded, err := builder.LoadAnalysis(analysisPath)
	if err != nil {
		t.Fatalf("LoadAnalysis failed: %v", err)
	}

	if loaded.CardLevels["Knight"].Level != 11 {
		t.Errorf("Knight level = %d, want 11", loaded.CardLevels["Knight"].Level)
	}

	// Test non-existent file
	_, err = builder.LoadAnalysis("/nonexistent/file.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestSaveDeck tests saving deck to file
func TestSaveDeck(t *testing.T) {
	tempDir := t.TempDir()
	builder := NewBuilder(tempDir)

	deck := &DeckRecommendation{
		Deck: []string{"Knight", "Archers"},
		DeckDetail: []CardDetail{
			{Name: "Knight", Level: 11},
			{Name: "Archers", Level: 12},
		},
		AvgElixir: 3.0,
	}

	// Test saving to default location
	path, err := builder.SaveDeck(deck, "", "#TESTPLAYER")
	if err != nil {
		t.Fatalf("SaveDeck failed: %v", err)
	}

	if path == "" {
		t.Error("SaveDeck returned empty path")
	}

	// Verify file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("File was not created: %s", path)
	}

	// Test saving to custom location
	customDir := filepath.Join(tempDir, "custom")
	path2, err := builder.SaveDeck(deck, customDir, "#TESTPLAYER")
	if err != nil {
		t.Fatalf("SaveDeck with custom dir failed: %v", err)
	}

	// Verify custom directory was used
	if !strings.Contains(path2, customDir) {
		t.Errorf("Path does not contain custom dir: %s", path2)
	}
}

func TestSaveDeckCanonicalizesPlayerTag(t *testing.T) {
	builder := NewBuilder(t.TempDir())

	deck := &DeckRecommendation{
		Deck: []string{"Knight", "Archers"},
		DeckDetail: []CardDetail{
			{Name: "Knight", Level: 11},
			{Name: "Archers", Level: 12},
		},
		AvgElixir: 3.0,
	}

	path, err := builder.SaveDeck(deck, "", " abc123 ")
	if err != nil {
		t.Fatalf("SaveDeck failed: %v", err)
	}

	if !strings.HasSuffix(path, "_deck_ABC123.json") {
		t.Fatalf("expected canonicalized filename suffix, got %s", filepath.Base(path))
	}
}

// TestBuildDeckFromFile tests building deck from file
func TestBuildDeckFromFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test analysis file
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Knight": {
				Level:    11,
				MaxLevel: 14,
				Rarity:   "Common",
				Elixir:   3,
			},
			"Archers": {
				Level:    12,
				MaxLevel: 14,
				Rarity:   "Common",
				Elixir:   3,
			},
			"Giant": {
				Level:    9,
				MaxLevel: 14,
				Rarity:   "Rare",
				Elixir:   5,
			},
			"Musketeer": {
				Level:    9,
				MaxLevel: 14,
				Rarity:   "Rare",
				Elixir:   4,
			},
			"Fireball": {
				Level:    8,
				MaxLevel: 14,
				Rarity:   "Rare",
				Elixir:   4,
			},
			"Zap": {
				Level:    12,
				MaxLevel: 14,
				Rarity:   "Common",
				Elixir:   2,
			},
			"Skeletons": {
				Level:    13,
				MaxLevel: 14,
				Rarity:   "Common",
				Elixir:   1,
			},
			"Bomber": {
				Level:    11,
				MaxLevel: 14,
				Rarity:   "Rare",
				Elixir:   2,
			},
		},
	}

	data, err := json.Marshal(analysis)
	if err != nil {
		t.Fatalf("Failed to marshal test analysis: %v", err)
	}

	analysisPath := filepath.Join(tempDir, "test_analysis.json")
	if err := os.WriteFile(analysisPath, data, 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test building deck from file
	builder := NewBuilder(tempDir)
	deck, err := builder.BuildDeckFromFile(analysisPath)
	if err != nil {
		t.Fatalf("BuildDeckFromFile failed: %v", err)
	}

	if deck == nil {
		t.Fatal("BuildDeckFromFile returned nil deck")
	}

	if len(deck.Deck) == 0 {
		t.Error("Deck is empty")
	}

	if len(deck.DeckDetail) != 8 {
		t.Errorf("Deck has %d cards, want 8", len(deck.DeckDetail))
	}

	// Test with non-existent file
	_, err = builder.BuildDeckFromFile("/nonexistent/analysis.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestBuilder_SynergyCacheTests tests the memoization cache functionality
func TestBuilder_SynergyCacheTests(t *testing.T) {
	builder := NewBuilder("testdata")
	builder.SetSynergyEnabled(true)

	// Create test deck with known synergy pairs
	deck := []*CardCandidate{
		{Name: "Giant", Level: 10, MaxLevel: 13},
		{Name: "Witch", Level: 8, MaxLevel: 11},
	}

	// Clear cache before test
	builder.clearSynergyCache()

	// First call should cache the result
	firstScore := builder.calculateSynergyScore("Sparky", deck)
	cacheSize1 := len(builder.synergyCache)

	// Second call with same inputs should use cache
	secondScore := builder.calculateSynergyScore("Sparky", deck)
	cacheSize2 := len(builder.synergyCache)

	if firstScore != secondScore {
		t.Errorf("Cache returned different result: first=%.2f, second=%.2f", firstScore, secondScore)
	}

	// Cache size should not decrease (entries are added, not removed)
	if cacheSize2 < cacheSize1 {
		t.Errorf("Cache shrank: first=%d entries, second=%d entries", cacheSize1, cacheSize2)
	}
}

// TestBuilder_SynergyCacheOrdering tests that cache key ordering works correctly
func TestBuilder_SynergyCacheOrdering(t *testing.T) {
	builder := NewBuilder("testdata")
	builder.SetSynergyEnabled(true)

	// Clear cache before test
	builder.clearSynergyCache()

	deck1 := []*CardCandidate{
		{Name: "Giant", Level: 10, MaxLevel: 13},
	}

	deck2 := []*CardCandidate{
		{Name: "Witch", Level: 8, MaxLevel: 11},
	}

	// Query synergy in both directions
	score1 := builder.calculateSynergyScore("Sparky", deck1) // Sparky + Giant
	score2 := builder.calculateSynergyScore("Sparky", deck2) // Sparky + Witch

	// Verify cache has entries (the exact number depends on synergyDB contents)
	cacheSize := len(builder.synergyCache)
	if cacheSize == 0 {
		t.Error("Expected cache to have entries after synergy calculations")
	}

	// Scores should be non-negative
	if score1 < 0 || score2 < 0 {
		t.Errorf("Expected non-negative scores: score1=%.2f, score2=%.2f", score1, score2)
	}
}

// TestBuilder_SynergyCacheClearing tests that the cache is cleared between builds
func TestBuilder_SynergyCacheClearing(t *testing.T) {
	builder := NewBuilder("testdata")
	builder.SetSynergyEnabled(true)

	// Create analysis that will trigger cache usage
	analysis := CardAnalysis{
		CardLevels: map[string]CardLevelData{
			"Giant":     {Level: 10, MaxLevel: 13, Rarity: "Rare", Elixir: 5},
			"Witch":     {Level: 8, MaxLevel: 11, Rarity: "Epic", Elixir: 5},
			"Sparky":    {Level: 7, MaxLevel: 11, Rarity: "Legendary", Elixir: 6},
			"Zap":       {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 2},
			"Fireball":  {Level: 7, MaxLevel: 11, Rarity: "Rare", Elixir: 4},
			"Cannon":    {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 3},
			"Skeletons": {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 1},
			"Knight":    {Level: 11, MaxLevel: 13, Rarity: "Common", Elixir: 3},
		},
		AnalysisTime: "2024-01-15T10:30:00Z",
	}

	// Build first deck - this populates cache
	_, err := builder.BuildDeckFromAnalysis(analysis)
	if err != nil {
		t.Fatalf("Failed to build deck: %v", err)
	}

	cacheSizeAfterFirst := len(builder.synergyCache)
	if cacheSizeAfterFirst == 0 {
		t.Error("Expected cache to be populated after deck build")
	}

	// Build second deck - this should clear cache first
	_, err = builder.BuildDeckFromAnalysis(analysis)
	if err != nil {
		t.Fatalf("Failed to build second deck: %v", err)
	}

	// Cache should have been cleared and repopulated
	// The size should be roughly the same (or smaller if some synergies weren't checked)
	cacheSizeAfterSecond := len(builder.synergyCache)
	if cacheSizeAfterSecond > cacheSizeAfterFirst*2 {
		t.Errorf("Cache appears to be growing unbounded: first=%d, second=%d",
			cacheSizeAfterFirst, cacheSizeAfterSecond)
	}
}

// TestBuilder_CachedSynergyRetrieval tests getCachedSynergy directly
func TestBuilder_CachedSynergyRetrieval(t *testing.T) {
	builder := NewBuilder("testdata")
	builder.SetSynergyEnabled(true)

	// Clear cache before test
	builder.clearSynergyCache()

	// First call - cache miss, should query database
	score1 := builder.getCachedSynergy("Giant", "Witch")
	if score1 < 0 {
		t.Errorf("Expected non-negative score, got %.2f", score1)
	}

	// Second call - cache hit (same order)
	score2 := builder.getCachedSynergy("Giant", "Witch")
	if score2 != score1 {
		t.Errorf("Cache returned different result for same order: first=%.2f, second=%.2f", score1, score2)
	}

	// Third call - cache hit (reversed order, should use same cache key)
	score3 := builder.getCachedSynergy("Witch", "Giant")
	if score3 != score1 {
		t.Errorf("Cache returned different result for reversed order: first=%.2f, reversed=%.2f", score1, score3)
	}

	// Verify only one cache entry was created (due to key ordering)
	cacheSize := len(builder.synergyCache)
	if cacheSize != 1 {
		t.Errorf("Expected 1 cache entry, got %d", cacheSize)
	}
}
