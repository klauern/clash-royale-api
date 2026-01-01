package config

import "testing"

func TestGetUpgradeCost(t *testing.T) {
	tests := []struct {
		name         string
		currentLevel int
		rarity       string
		expectedCost int
	}{
		// Common rarity
		{"Common level 1", 1, "Common", 2},
		{"Common level 10", 10, "Common", 1000},
		{"Common level 15", 15, "Common", 5500},
		{"Common at max", 16, "Common", 0},
		{"Common lowercase", 10, "common", 1000},

		// Rare rarity
		{"Rare level 10", 10, "Rare", 300},
		{"Rare level 15", 15, "Rare", 1000},
		{"Rare at max", 16, "Rare", 0},

		// Epic rarity
		{"Epic level 10", 10, "Epic", 50},
		{"Epic level 15", 15, "Epic", 140},
		{"Epic at max", 16, "Epic", 0},

		// Legendary rarity
		{"Legendary level 10", 10, "Legendary", 4},
		{"Legendary level 15", 15, "Legendary", 15},
		{"Legendary at max", 16, "Legendary", 0},

		// Champion rarity
		{"Champion level 12", 12, "Champion", 4},
		{"Champion level 15", 15, "Champion", 12},
		{"Champion at max", 16, "Champion", 0},

		// Invalid cases
		{"Invalid level negative", -1, "Common", 0},
		{"Invalid level too high", 20, "Common", 0},
		{"Invalid rarity", 10, "InvalidRarity", 0},
		{"Empty rarity", 10, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUpgradeCost(tt.currentLevel, tt.rarity)
			if result != tt.expectedCost {
				t.Errorf("GetUpgradeCost(%d, %q) = %d, want %d",
					tt.currentLevel, tt.rarity, result, tt.expectedCost)
			}
		})
	}
}

func TestGetGoldCost(t *testing.T) {
	tests := []struct {
		name         string
		currentLevel int
		rarity       string
		expectedCost int
	}{
		// Common rarity
		{"Common level 1", 1, "Common", 5},
		{"Common level 10", 10, "Common", 20000},
		{"Common level 13", 13, "Common", 100000},
		{"Common lowercase", 10, "common", 20000},

		// Rare rarity
		{"Rare level 3", 3, "Rare", 50},
		{"Rare level 10", 10, "Rare", 20000},
		{"Rare level 13", 13, "Rare", 100000},
		{"Rare level 1 (no data)", 1, "Rare", 0},

		// Epic rarity
		{"Epic level 6", 6, "Epic", 400},
		{"Epic level 10", 10, "Epic", 20000},
		{"Epic level 13", 13, "Epic", 100000},
		{"Epic level 1 (no data)", 1, "Epic", 0},

		// Legendary rarity
		{"Legendary level 9", 9, "Legendary", 5000},
		{"Legendary level 11", 11, "Legendary", 50000},
		{"Legendary level 13", 13, "Legendary", 100000},
		{"Legendary level 1 (no data)", 1, "Legendary", 0},

		// Champion rarity
		{"Champion level 11", 11, "Champion", 50000},
		{"Champion level 13", 13, "Champion", 100000},
		{"Champion level 1 (no data)", 1, "Champion", 0},

		// Invalid cases
		{"Invalid level negative", -1, "Common", 0},
		{"Invalid level 14 (no data yet)", 14, "Common", 0},
		{"Invalid rarity", 10, "InvalidRarity", 0},
		{"Empty rarity", 10, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetGoldCost(tt.currentLevel, tt.rarity)
			if result != tt.expectedCost {
				t.Errorf("GetGoldCost(%d, %q) = %d, want %d",
					tt.currentLevel, tt.rarity, result, tt.expectedCost)
			}
		})
	}
}

func TestCalculateTotalCardsToMax(t *testing.T) {
	tests := []struct {
		name          string
		currentLevel  int
		rarity        string
		expectedTotal int
	}{
		// Common rarity
		{"Common from level 1", 1, "Common", 2 + 4 + 10 + 20 + 50 + 100 + 200 + 400 + 800 + 1000 + 2000 + 3000 + 2500 + 3500 + 5500},
		{"Common from level 10", 10, "Common", 1000 + 2000 + 3000 + 2500 + 3500 + 5500},
		{"Common from level 15", 15, "Common", 5500},
		{"Common at max", 16, "Common", 0},
		{"Common lowercase", 10, "common", 1000 + 2000 + 3000 + 2500 + 3500 + 5500},

		// Rare rarity
		{"Rare from level 10", 10, "Rare", 300 + 400 + 400 + 550 + 750 + 1000},
		{"Rare from level 15", 15, "Rare", 1000},
		{"Rare at max", 16, "Rare", 0},

		// Epic rarity
		{"Epic from level 10", 10, "Epic", 50 + 30 + 40 + 70 + 100 + 140},
		{"Epic from level 15", 15, "Epic", 140},
		{"Epic at max", 16, "Epic", 0},

		// Legendary rarity
		{"Legendary from level 10", 10, "Legendary", 4 + 10 + 20 + 10 + 12 + 15},
		{"Legendary from level 15", 15, "Legendary", 15},
		{"Legendary at max", 16, "Legendary", 0},

		// Champion rarity
		{"Champion from level 11", 11, "Champion", 2 + 4 + 8 + 10 + 12},
		{"Champion from level 15", 15, "Champion", 12},
		{"Champion at max", 16, "Champion", 0},

		// Invalid cases
		{"Invalid level negative", -1, "Common", 0},
		{"Invalid level too high", 20, "Common", 0},
		{"Invalid rarity", 10, "InvalidRarity", 0},
		{"Empty rarity", 10, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTotalCardsToMax(tt.currentLevel, tt.rarity)
			if result != tt.expectedTotal {
				t.Errorf("CalculateTotalCardsToMax(%d, %q) = %d, want %d",
					tt.currentLevel, tt.rarity, result, tt.expectedTotal)
			}
		})
	}
}

func TestUpgradeCostConsistency(t *testing.T) {
	// Verify that upgrade costs exist for all rarities from their starting level to max-1
	rarities := GetAllRarities()

	for _, rarity := range rarities {
		t.Run(rarity, func(t *testing.T) {
			startLevel := GetStartingLevel(rarity)
			maxLevel := GetMaxLevel(rarity)

			for level := startLevel; level < maxLevel; level++ {
				cost := GetUpgradeCost(level, rarity)
				if cost == 0 {
					// Allow level 1 to have 0 cost for non-Common rarities (they start at higher levels)
					if level == 1 && startLevel > 1 {
						continue
					}
					t.Errorf("Expected non-zero upgrade cost for %s at level %d", rarity, level)
				}
			}

			// Max level should have 0 upgrade cost
			maxCost := GetUpgradeCost(maxLevel, rarity)
			if maxCost != 0 {
				t.Errorf("Expected zero upgrade cost for %s at max level %d, got %d",
					rarity, maxLevel, maxCost)
			}
		})
	}
}

func TestGoldCostDataAvailability(t *testing.T) {
	// Document the current state of gold cost data availability
	// Gold costs are currently available up to level 13

	testCases := []struct {
		rarity     string
		startLevel int
		endLevel   int // Last level with known gold cost
	}{
		{"Common", 1, 13},
		{"Rare", 3, 13},
		{"Epic", 6, 13},
		{"Legendary", 9, 13},
		{"Champion", 11, 13},
	}

	for _, tc := range testCases {
		t.Run(tc.rarity, func(t *testing.T) {
			// Verify data exists from start to end level
			for level := tc.startLevel; level <= tc.endLevel; level++ {
				cost := GetGoldCost(level, tc.rarity)
				if cost == 0 {
					t.Errorf("Expected gold cost data for %s at level %d", tc.rarity, level)
				}
			}

			// Verify no data exists beyond end level (up to level 15)
			for level := tc.endLevel + 1; level <= 15; level++ {
				cost := GetGoldCost(level, tc.rarity)
				if cost != 0 {
					t.Errorf("Unexpected gold cost data for %s at level %d: got %d, want 0",
						tc.rarity, level, cost)
				}
			}
		})
	}
}
