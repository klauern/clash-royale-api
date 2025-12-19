package deck

import (
	"testing"
)

func TestCalculateGoldNeeded(t *testing.T) {
	tests := []struct {
		name         string
		currentLevel int
		targetLevel  int
		rarity       string
		expected     int
	}{
		{"Common 1 to 2", 1, 2, "Common", 5},
		{"Common 1 to 14", 1, 14, "Common", 185625}, // Sum of all upgrades
		{"Rare 3 to 4", 3, 4, "Rare", 50},
		{"Legendary 9 to 14", 9, 14, "Legendary", 175000},
		{"Already at target", 10, 10, "Common", 0},
		{"Above target", 12, 10, "Common", 0},
		{"Invalid rarity", 1, 5, "Invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateGoldNeeded(tt.currentLevel, tt.targetLevel, tt.rarity)
			if result != tt.expected {
				t.Errorf("CalculateGoldNeeded(%d, %d, %s) = %d, want %d",
					tt.currentLevel, tt.targetLevel, tt.rarity, result, tt.expected)
			}
		})
	}
}

func TestCalculateTotalGoldToMax(t *testing.T) {
	tests := []struct {
		name         string
		currentLevel int
		rarity       string
		expected     int
	}{
		{"Common from 1", 1, "Common", 185625},
		{"Rare from 3", 3, "Rare", 185600},
		{"Epic from 6", 6, "Epic", 184400},
		{"Legendary from 9", 9, "Legendary", 175000},
		{"Champion from 11", 11, "Champion", 150000},
		{"Already at max", 14, "Common", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTotalGoldToMax(tt.currentLevel, tt.rarity)
			if result != tt.expected {
				t.Errorf("CalculateTotalGoldToMax(%d, %s) = %d, want %d",
					tt.currentLevel, tt.rarity, result, tt.expected)
			}
		})
	}
}

func TestNewProjection(t *testing.T) {
	// Create test deck
	roleWinCond := RoleWinCondition
	roleSupport := RoleSupport

	deck := []*CardCandidate{
		{Name: "Hog Rider", Level: 11, MaxLevel: 14, Rarity: "Rare", Role: &roleWinCond, Score: 0.8},
		{Name: "Ice Spirit", Level: 10, MaxLevel: 14, Rarity: "Common", Role: &roleSupport, Score: 0.6},
		{Name: "Fireball", Level: 9, MaxLevel: 14, Rarity: "Rare", Role: &roleSupport, Score: 0.7},
	}

	tests := []struct {
		name   string
		policy TargetLevelPolicy
	}{
		{"Max All Policy", PolicyMaxAll},
		{"Match Highest Policy", PolicyMatchHighest},
		{"Tournament Policy", PolicyTournament},
		{"Budget Policy", PolicyBudget},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projection := NewProjection(deck, tt.policy, nil)

			if projection == nil {
				t.Fatal("NewProjection returned nil")
			}

			if projection.Deck == nil || len(projection.Deck) != 3 {
				t.Error("Projection deck is invalid")
			}

			if projection.CurrentScore <= 0 {
				t.Error("Current score should be > 0")
			}

			if projection.UpgradePath == nil {
				t.Error("Upgrade path should not be nil")
			}

			if tt.policy == PolicyBudget && projection.UpgradePath.TotalGold > 100000 {
				t.Error("Budget policy should have lower upgrade cost")
			}
		})
	}
}

func TestDetermineTargetLevels(t *testing.T) {
	deck := []*CardCandidate{
		{Name: "Card1", Level: 10, MaxLevel: 14},
		{Name: "Card2", Level: 12, MaxLevel: 14},
		{Name: "Card3", Level: 9, MaxLevel: 14},
	}

	t.Run("PolicyMaxAll", func(t *testing.T) {
		targets := determineTargetLevels(deck, PolicyMaxAll, nil)
		for name, level := range targets {
			if level != 14 {
				t.Errorf("Card %s should target level 14, got %d", name, level)
			}
		}
	})

	t.Run("PolicyMatchHighest", func(t *testing.T) {
		targets := determineTargetLevels(deck, PolicyMatchHighest, nil)
		for name, level := range targets {
			if level != 12 {
				t.Errorf("Card %s should target level 12 (highest), got %d", name, level)
			}
		}
	})

	t.Run("PolicyTournament", func(t *testing.T) {
		targets := determineTargetLevels(deck, PolicyTournament, nil)
		for name, level := range targets {
			if level != 11 {
				t.Errorf("Card %s should target level 11 (tournament), got %d", name, level)
			}
		}
	})

	t.Run("PolicyBudget", func(t *testing.T) {
		targets := determineTargetLevels(deck, PolicyBudget, nil)
		if targets["Card1"] != 11 {
			t.Errorf("Card1 should target 11 (10+1), got %d", targets["Card1"])
		}
		if targets["Card2"] != 13 {
			t.Errorf("Card2 should target 13 (12+1), got %d", targets["Card2"])
		}
	})

	t.Run("PolicyCustom", func(t *testing.T) {
		customLevels := map[string]int{
			"Card1": 13,
			"Card2": 14,
			"Card3": 11,
		}
		targets := determineTargetLevels(deck, PolicyCustom, customLevels)
		for name, expectedLevel := range customLevels {
			if targets[name] != expectedLevel {
				t.Errorf("Card %s should target custom level %d, got %d", name, expectedLevel, targets[name])
			}
		}
	})
}

func TestCalculateUpgradePath(t *testing.T) {
	roleSupport := RoleSupport

	deck := []*CardCandidate{
		{Name: "Ice Spirit", Level: 10, MaxLevel: 14, Rarity: "Common", Role: &roleSupport, Score: 0.6},
		{Name: "Hog Rider", Level: 11, MaxLevel: 14, Rarity: "Rare", Role: &roleSupport, Score: 0.8},
	}

	projection := NewProjection(deck, PolicyMaxAll, nil)

	if projection.UpgradePath == nil {
		t.Fatal("Upgrade path should not be nil")
	}

	path := projection.UpgradePath

	if len(path.CardUpgrades) != 2 {
		t.Errorf("Expected 2 card upgrades, got %d", len(path.CardUpgrades))
	}

	if path.TotalGold <= 0 {
		t.Error("Total gold should be > 0")
	}

	if path.TotalCards <= 0 {
		t.Error("Total cards should be > 0")
	}

	// Verify specific card upgrade
	for _, upgrade := range path.CardUpgrades {
		if upgrade.TargetLevel != 14 {
			t.Errorf("Card %s should target level 14, got %d", upgrade.CardName, upgrade.TargetLevel)
		}
		if upgrade.CurrentLevel >= upgrade.TargetLevel {
			t.Errorf("Card %s current level %d >= target %d", upgrade.CardName, upgrade.CurrentLevel, upgrade.TargetLevel)
		}
	}
}

func TestScoreAtLevel(t *testing.T) {
	roleSupport := RoleSupport

	deck := []*CardCandidate{
		{Name: "Card1", Level: 10, MaxLevel: 14, Rarity: "Common", Role: &roleSupport, Score: 0.5},
		{Name: "Card2", Level: 11, MaxLevel: 14, Rarity: "Rare", Role: &roleSupport, Score: 0.6},
	}

	projection := NewProjection(deck, PolicyMaxAll, nil)

	score11 := projection.ScoreAtLevel(11)
	score14 := projection.ScoreAtLevel(14)

	if score11 <= 0 {
		t.Error("Score at level 11 should be > 0")
	}

	if score14 <= 0 {
		t.Error("Score at level 14 should be > 0")
	}

	// Note: Without actual scoring algorithm integration, scores might be the same
	// This test just ensures the function executes without error
}

func TestSimulateUpgrade(t *testing.T) {
	roleSupport := RoleSupport

	deck := []*CardCandidate{
		{Name: "Ice Spirit", Level: 10, MaxLevel: 14, Rarity: "Common", Role: &roleSupport, Score: 0.6},
		{Name: "Hog Rider", Level: 11, MaxLevel: 14, Rarity: "Rare", Role: &roleSupport, Score: 0.8},
	}

	projection := NewProjection(deck, PolicyMatchHighest, nil)
	simulatedProjection := projection.SimulateUpgrade("Ice Spirit", 11)

	if simulatedProjection == nil {
		t.Fatal("Simulated projection should not be nil")
	}

	// Verify that the simulated deck has upgraded Ice Spirit
	iceSpirit := simulatedProjection.Deck[0]
	if iceSpirit.Name == "Ice Spirit" && iceSpirit.Level != 11 {
		t.Errorf("Ice Spirit should be upgraded to 11, got %d", iceSpirit.Level)
	}

	// Verify that upgrade costs changed
	if projection.UpgradePath.TotalGold <= simulatedProjection.UpgradePath.TotalGold {
		t.Error("Simulated projection should have lower upgrade cost after simulated upgrade")
	}
}

func TestCompareProjections(t *testing.T) {
	roleSupport := RoleSupport

	deckA := []*CardCandidate{
		{Name: "Ice Spirit", Level: 10, MaxLevel: 14, Rarity: "Common", Role: &roleSupport, Score: 0.6},
		{Name: "Hog Rider", Level: 11, MaxLevel: 14, Rarity: "Rare", Role: &roleSupport, Score: 0.8},
	}

	deckB := []*CardCandidate{
		{Name: "Giant", Level: 9, MaxLevel: 14, Rarity: "Rare", Role: &roleSupport, Score: 0.7},
		{Name: "Witch", Level: 10, MaxLevel: 14, Rarity: "Epic", Role: &roleSupport, Score: 0.75},
	}

	projectionA := NewProjection(deckA, PolicyMaxAll, nil)
	projectionB := NewProjection(deckB, PolicyMaxAll, nil)

	comparison := CompareProjections(projectionA, projectionB)

	if comparison == nil {
		t.Fatal("Comparison should not be nil")
	}

	if comparison.DeckA == "" || comparison.DeckB == "" {
		t.Error("Deck names should not be empty")
	}

	if comparison.Recommendation == "" {
		t.Error("Recommendation should not be empty")
	}
}

func TestNilInputs(t *testing.T) {
	t.Run("NewProjection with nil deck", func(t *testing.T) {
		projection := NewProjection(nil, PolicyMaxAll, nil)
		if projection != nil {
			t.Error("Should return nil for nil deck")
		}
	})

	t.Run("CompareProjections with nil", func(t *testing.T) {
		roleSupport := RoleSupport
		deck := []*CardCandidate{
			{Name: "Card1", Level: 10, MaxLevel: 14, Rarity: "Common", Role: &roleSupport, Score: 0.5},
		}
		projection := NewProjection(deck, PolicyMaxAll, nil)

		comparison := CompareProjections(nil, projection)
		if comparison != nil {
			t.Error("Should return nil when first projection is nil")
		}

		comparison = CompareProjections(projection, nil)
		if comparison != nil {
			t.Error("Should return nil when second projection is nil")
		}
	})
}
