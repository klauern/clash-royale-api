// Package config provides centralized game configuration constants and mappings.
// This file consolidates upgrade cost data that was previously duplicated
// across multiple packages (deck, analysis).
package config

// Upgrade costs define how many cards are needed to upgrade from each level.
// Maps: rarity -> currentLevel -> cardsNeeded
// Note: Level 16 is max, so there are no upgrade costs for level 16
var upgradeCosts = map[string]map[int]int{
	"Common": {
		1:  2,
		2:  4,
		3:  10,
		4:  20,
		5:  50,
		6:  100,
		7:  200,
		8:  400,
		9:  800,
		10: 1000,
		11: 2000,
		12: 3000,
		13: 2500, // Updated 2025
		14: 3500, // Updated 2025
		15: 5500, // Updated 2025
	},
	"Rare": {
		1:  2, // Fallback for low levels
		2:  2,
		3:  2,
		4:  4,
		5:  10,
		6:  20,
		7:  50,
		8:  100,
		9:  200,
		10: 300,  // Updated 2025
		11: 400,  // Updated 2025
		12: 400,  // Updated 2025
		13: 550,  // Updated 2025
		14: 750,  // Updated 2025
		15: 1000, // Updated 2025
	},
	"Epic": {
		1:  2, // Fallback for low levels
		2:  2,
		3:  2,
		4:  2,
		5:  2,
		6:  2,
		7:  4,
		8:  10,
		9:  20,
		10: 50,
		11: 30,  // Updated 2025
		12: 40,  // Updated 2025
		13: 70,  // Updated 2025
		14: 100, // Updated 2025
		15: 140, // Updated 2025
	},
	"Legendary": {
		1:  2, // Fallback for low levels
		2:  2,
		3:  2,
		4:  2,
		5:  2,
		6:  2,
		7:  2,
		8:  2,
		9:  2,
		10: 4,
		11: 10,
		12: 20,
		13: 10, // Updated 2025
		14: 12, // Updated 2025
		15: 15, // Updated 2025
	},
	"Champion": {
		1:  2, // Fallback for low levels
		11: 2,
		12: 4,
		13: 8,  // Updated 2025
		14: 10, // Updated 2025
		15: 12, // Updated 2025
	},
}

// Gold costs define how much gold is needed to upgrade from each level.
// Maps: rarity -> currentLevel -> goldCost
// Note: Data currently available up to level 13
var goldCosts = map[string]map[int]int{
	"Common": {
		1: 5, 2: 20, 3: 50, 4: 150, 5: 400, 6: 1000, 7: 2000,
		8: 4000, 9: 8000, 10: 20000, 11: 50000, 12: 100000, 13: 100000,
	},
	"Rare": {
		3: 50, 4: 150, 5: 400, 6: 1000, 7: 2000,
		8: 4000, 9: 8000, 10: 20000, 11: 50000, 12: 100000, 13: 100000,
	},
	"Epic": {
		6: 400, 7: 2000, 8: 4000, 9: 8000, 10: 20000, 11: 50000, 12: 100000, 13: 100000,
	},
	"Legendary": {
		9: 5000, 10: 20000, 11: 50000, 12: 100000, 13: 100000,
	},
	"Champion": {
		11: 50000, 12: 100000, 13: 100000,
	},
}

// GetUpgradeCost returns the number of cards needed to upgrade from a specific level.
// Returns 0 if the rarity/level combination is invalid or if already at max level.
//
// Parameters:
//   - currentLevel: The current level of the card (1-15)
//   - rarity: The rarity of the card (Common, Rare, Epic, Legendary, Champion)
//
// Example:
//   - GetUpgradeCost(10, "Common") returns 1000 (cards needed to go from 10→11)
//   - GetUpgradeCost(15, "Common") returns 5500 (cards needed to go from 15→16)
//   - GetUpgradeCost(16, "Common") returns 0 (already at max)
func GetUpgradeCost(currentLevel int, rarity string) int {
	normalized := NormalizeRarity(rarity)
	costs, exists := upgradeCosts[normalized]
	if !exists {
		return 0
	}

	cardsNeeded, exists := costs[currentLevel]
	if !exists {
		return 0
	}

	return cardsNeeded
}

// GetGoldCost returns the amount of gold needed to upgrade from a specific level.
// Returns 0 if the rarity/level combination is invalid or data not available.
//
// Parameters:
//   - currentLevel: The current level of the card
//   - rarity: The rarity of the card (Common, Rare, Epic, Legendary, Champion)
//
// Example:
//   - GetGoldCost(10, "Common") returns 20000 (gold needed to go from 10→11)
//   - GetGoldCost(1, "Legendary") returns 0 (no data for level 1 legendary)
//
// Note: Gold cost data is currently available up to level 13. Returns 0 for higher levels.
func GetGoldCost(currentLevel int, rarity string) int {
	normalized := NormalizeRarity(rarity)
	costs, exists := goldCosts[normalized]
	if !exists {
		return 0
	}

	goldNeeded, exists := costs[currentLevel]
	if !exists {
		return 0
	}

	return goldNeeded
}

// CalculateTotalCardsToMax calculates the total number of cards needed to reach
// max level from a given current level.
//
// Parameters:
//   - currentLevel: The current level of the card
//   - rarity: The rarity of the card (Common, Rare, Epic, Legendary, Champion)
//
// Returns:
//   - Total cards needed to reach max level (sum of all upgrade costs)
//   - Returns 0 if already at max level or invalid rarity
//
// Example:
//   - For a Common card at level 10, calculates: costs[10] + costs[11] + ... + costs[15]
//   - For a card already at max level (16), returns 0
func CalculateTotalCardsToMax(currentLevel int, rarity string) int {
	normalized := NormalizeRarity(rarity)
	maxLevel := GetMaxLevel(normalized)

	// Validate level range
	if currentLevel < 1 || currentLevel >= maxLevel {
		return 0
	}

	costs, exists := upgradeCosts[normalized]
	if !exists {
		return 0
	}

	total := 0
	for level := currentLevel; level < maxLevel; level++ {
		if cardsNeeded, exists := costs[level]; exists {
			total += cardsNeeded
		}
	}

	return total
}
