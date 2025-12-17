// Package analysis provides upgrade calculation utilities for Clash Royale cards.
// Based on official Clash Royale card progression system.
package analysis

// Upgrade costs define how many cards are needed to upgrade from each level
// Maps: rarity -> currentLevel -> cardsNeeded
// Note: Level 14 is max, so there are no upgrade costs for level 14
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
		12: 5000,
		13: 10000,
	},
	"Rare": {
		3:  2,
		4:  4,
		5:  10,
		6:  20,
		7:  50,
		8:  100,
		9:  200,
		10: 400,
		11: 800,
		12: 1000,
		13: 2000,
	},
	"Epic": {
		6:  2,
		7:  4,
		8:  10,
		9:  20,
		10: 50,
		11: 100,
		12: 200,
		13: 400,
	},
	"Legendary": {
		9:  2,
		10: 4,
		11: 10,
		12: 20,
		13: 40,
	},
	"Champion": {
		11: 2,
		12: 4,
		13: 10,
	},
}

// Max levels for each rarity
var maxLevels = map[string]int{
	"Common":    14,
	"Rare":      14,
	"Epic":      14,
	"Legendary": 14,
	"Champion":  14,
}

// Starting levels for each rarity (when first unlocked)
var startingLevels = map[string]int{
	"Common":    1,
	"Rare":      3,
	"Epic":      6,
	"Legendary": 9,
	"Champion":  11,
}

// Total cards per rarity in the game (approximate counts - needs verification)
// TODO: Replace with actual counts from card database or fetch from API
var totalCardsPerRarity = map[string]int{
	"Common":    30,
	"Rare":      30,
	"Epic":      20,
	"Legendary": 15,
	"Champion":  5,
}

// CalculateCardsNeeded returns how many cards are needed to upgrade from currentLevel
// Returns 0 if already at max level or invalid rarity
func CalculateCardsNeeded(currentLevel int, rarity string) int {
	costs, exists := upgradeCosts[rarity]
	if !exists {
		return 0
	}

	cardsNeeded, exists := costs[currentLevel]
	if !exists {
		return 0 // Either max level or invalid level for this rarity
	}

	return cardsNeeded
}

// GetMaxLevel returns the maximum level for a given rarity
func GetMaxLevel(rarity string) int {
	if maxLevel, exists := maxLevels[rarity]; exists {
		return maxLevel
	}
	return 14 // Default to 14 if unknown
}

// GetStartingLevel returns the initial level when a card is first unlocked
func GetStartingLevel(rarity string) int {
	if startLevel, exists := startingLevels[rarity]; exists {
		return startLevel
	}
	return 1 // Default to 1 if unknown
}

// IsMaxLevel checks if a card is at maximum level for its rarity
func IsMaxLevel(currentLevel int, rarity string) bool {
	return currentLevel >= GetMaxLevel(rarity)
}

// CalculateTotalCardsToMax calculates total cards needed from current level to max
func CalculateTotalCardsToMax(currentLevel int, rarity string) int {
	maxLevel := GetMaxLevel(rarity)
	if currentLevel >= maxLevel {
		return 0
	}

	costs, exists := upgradeCosts[rarity]
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

// CalculateUpgradeProgress calculates upgrade progress as a percentage (0-100)
func CalculateUpgradeProgress(cardsOwned, cardsNeeded int) float64 {
	if cardsNeeded == 0 {
		return 100.0
	}

	progress := (float64(cardsOwned) / float64(cardsNeeded)) * 100.0
	if progress > 100.0 {
		return 100.0
	}

	return progress
}

// UpgradeInfo contains complete upgrade information for a card
type UpgradeInfo struct {
	CardName         string  `json:"card_name"`
	Rarity           string  `json:"rarity"`
	ElixirCost       int     `json:"elixir_cost,omitempty"`
	CurrentLevel     int     `json:"current_level"`
	MaxLevel         int     `json:"max_level"`
	IsMaxLevel       bool    `json:"is_max_level"`
	CardsOwned       int     `json:"cards_owned"`
	CardsToNextLevel int     `json:"cards_to_next_level"`
	ProgressPercent  float64 `json:"progress_percent"`
	CanUpgradeNow    bool    `json:"can_upgrade_now"`
	TotalToMax       int     `json:"total_to_max"`
	LevelsToMax      int     `json:"levels_to_max"`
	MaxEvolutionLevel int    `json:"max_evolution_level,omitempty"`
}

// CalculateUpgradeInfo creates a complete UpgradeInfo for a card
func CalculateUpgradeInfo(cardName string, rarity string, elixirCost int, currentLevel int, cardsOwned int, maxEvolutionLevel int) UpgradeInfo {
	maxLevel := GetMaxLevel(rarity)
	isMaxLevel := IsMaxLevel(currentLevel, rarity)
	cardsNeeded := CalculateCardsNeeded(currentLevel, rarity)
	totalToMax := CalculateTotalCardsToMax(currentLevel, rarity)
	progress := CalculateUpgradeProgress(cardsOwned, cardsNeeded)
	canUpgrade := cardsOwned >= cardsNeeded && !isMaxLevel

	return UpgradeInfo{
		CardName:          cardName,
		Rarity:            rarity,
		ElixirCost:        elixirCost,
		CurrentLevel:      currentLevel,
		MaxLevel:          maxLevel,
		IsMaxLevel:        isMaxLevel,
		CardsOwned:        cardsOwned,
		CardsToNextLevel:  cardsNeeded,
		ProgressPercent:   progress,
		CanUpgradeNow:     canUpgrade,
		TotalToMax:        totalToMax,
		LevelsToMax:       maxLevel - currentLevel,
		MaxEvolutionLevel: maxEvolutionLevel,
	}
}

// RarityUpgradeStats contains aggregate statistics for a rarity
type RarityUpgradeStats struct {
	Rarity             string  `json:"rarity"`
	TotalCards         int     `json:"total_cards"`
	MaxLevelCards      int     `json:"max_level_cards"`
	UpgradableCards    int     `json:"upgradable_cards"`
	AvgLevel           float64 `json:"avg_level"`
	AvgProgressPercent float64 `json:"avg_progress_percent"`
	TotalCardsNeeded   int     `json:"total_cards_needed"` // Total cards needed for all upgrades
	CompletionPercent  float64 `json:"completion_percent"` // % of cards at max level
}

// CalculateRarityStats computes aggregate statistics for cards of a specific rarity
func CalculateRarityStats(cards []UpgradeInfo, rarity string) RarityUpgradeStats {
	filtered := make([]UpgradeInfo, 0)
	for _, card := range cards {
		if card.Rarity == rarity {
			filtered = append(filtered, card)
		}
	}

	if len(filtered) == 0 {
		return RarityUpgradeStats{Rarity: rarity}
	}

	totalLevel := 0
	totalProgress := 0.0
	maxLevelCount := 0
	upgradableCount := 0
	totalNeeded := 0

	for _, card := range filtered {
		totalLevel += card.CurrentLevel
		totalProgress += card.ProgressPercent

		if card.IsMaxLevel {
			maxLevelCount++
		}

		if card.CanUpgradeNow {
			upgradableCount++
		}

		totalNeeded += card.TotalToMax
	}

	cardCount := len(filtered)
	return RarityUpgradeStats{
		Rarity:             rarity,
		TotalCards:         cardCount,
		MaxLevelCards:      maxLevelCount,
		UpgradableCards:    upgradableCount,
		AvgLevel:           float64(totalLevel) / float64(cardCount),
		AvgProgressPercent: totalProgress / float64(cardCount),
		TotalCardsNeeded:   totalNeeded,
		CompletionPercent:  (float64(maxLevelCount) / float64(cardCount)) * 100.0,
	}
}

// CalculatePriorityScore computes an upgrade priority score (0-100)
// Higher score = higher priority for upgrading
// Factors:
// - Proximity to next level (50% weight)
// - Current level ratio (30% weight)
// - Rarity boost (20% weight)
func CalculatePriorityScore(info UpgradeInfo) float64 {
	// Already max level = 0 priority
	if info.IsMaxLevel {
		return 0.0
	}

	// Proximity to next level (0-100, higher if closer to upgrade)
	proximityScore := info.ProgressPercent

	// Level ratio (cards at higher levels are better to upgrade)
	levelRatio := float64(info.CurrentLevel) / float64(info.MaxLevel)
	levelScore := levelRatio * 100.0

	// Rarity boost (prioritize harder-to-get cards)
	rarityScores := map[string]float64{
		"Common":    0.0,
		"Rare":      20.0,
		"Epic":      40.0,
		"Legendary": 60.0,
		"Champion":  80.0,
	}

	rarityScore, exists := rarityScores[info.Rarity]
	if !exists {
		rarityScore = 0.0
	}

	// Weighted combination
	priorityScore := (proximityScore * 0.5) + (levelScore * 0.3) + (rarityScore * 0.2)

	// Boost if can upgrade immediately
	if info.CanUpgradeNow {
		priorityScore *= 1.2 // 20% boost
		if priorityScore > 100.0 {
			priorityScore = 100.0
		}
	}

	return priorityScore
}

// GetUpgradePriorities returns a sorted list of cards by upgrade priority
func GetUpgradePriorities(cards []UpgradeInfo, minScore float64, topN int) []UpgradeInfo {
	// Filter by minimum score and exclude max level
	filtered := make([]UpgradeInfo, 0)
	for _, card := range cards {
		score := CalculatePriorityScore(card)
		if score >= minScore && !card.IsMaxLevel {
			filtered = append(filtered, card)
		}
	}

	// Sort by priority score (descending)
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			scoreI := CalculatePriorityScore(filtered[i])
			scoreJ := CalculatePriorityScore(filtered[j])
			if scoreJ > scoreI {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// Return top N
	if topN > 0 && topN < len(filtered) {
		return filtered[:topN]
	}

	return filtered
}
