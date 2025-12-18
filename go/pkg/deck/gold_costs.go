// Package deck provides gold upgrade cost calculations for Clash Royale cards.
// Based on official Clash Royale card progression system.
package deck

// Gold upgrade costs define how much gold is needed to upgrade from each level
// Maps: rarity -> currentLevel -> goldNeeded
// Note: Level 14 is max, so there are no upgrade costs for level 14
var goldCosts = map[string]map[int]int{
	"Common": {
		1:  5,
		2:  20,
		3:  50,
		4:  150,
		5:  400,
		6:  1000,
		7:  2000,
		8:  4000,
		9:  8000,
		10: 20000,
		11: 50000,
		12: 100000,
	},
	"Rare": {
		3:  50,
		4:  150,
		5:  400,
		6:  1000,
		7:  2000,
		8:  4000,
		9:  8000,
		10: 20000,
		11: 50000,
		12: 100000,
	},
	"Epic": {
		6:  400,
		7:  2000,
		8:  4000,
		9:  8000,
		10: 20000,
		11: 50000,
		12: 100000,
	},
	"Legendary": {
		9:  5000,
		10: 20000,
		11: 50000,
		12: 100000,
	},
	"Champion": {
		11: 50000,
		12: 100000,
	},
}

// CalculateGoldNeeded returns how much gold is needed to upgrade from currentLevel to targetLevel
// Returns 0 if already at target level, invalid rarity, or invalid level range
func CalculateGoldNeeded(currentLevel, targetLevel int, rarity string) int {
	if currentLevel >= targetLevel {
		return 0
	}

	costs, exists := goldCosts[rarity]
	if !exists {
		return 0
	}

	totalGold := 0
	for level := currentLevel; level < targetLevel; level++ {
		if goldNeeded, exists := costs[level]; exists {
			totalGold += goldNeeded
		}
	}

	return totalGold
}

// CalculateTotalGoldToMax returns total gold needed from current level to max level
func CalculateTotalGoldToMax(currentLevel int, rarity string) int {
	maxLevel := 14 // Max level for all rarities
	return CalculateGoldNeeded(currentLevel, maxLevel, rarity)
}

// GetGoldForSingleUpgrade returns gold needed for just the next upgrade
func GetGoldForSingleUpgrade(currentLevel int, rarity string) int {
	costs, exists := goldCosts[rarity]
	if !exists {
		return 0
	}

	goldNeeded, exists := costs[currentLevel]
	if !exists {
		return 0 // Either max level or invalid level for this rarity
	}

	return goldNeeded
}
