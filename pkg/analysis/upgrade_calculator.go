// Package analysis provides upgrade calculation utilities for Clash Royale cards.
// Based on official Clash Royale card progression system.
package analysis

import "strings"

// CardInfo interface for card data
// This allows the package to work without importing the clashroyale package directly
type CardInfo interface {
	GetRarity() string
}

// cardAdapter adapts different card types to CardInfo interface
type cardAdapter struct {
	rarity string
}

func (c cardAdapter) GetRarity() string {
	return c.rarity
}

// NormalizeRarity ensures rarity strings are in TitleCase for consistent map lookups
func NormalizeRarity(rarity string) string {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "common":
		return "Common"
	case "rare":
		return "Rare"
	case "epic":
		return "Epic"
	case "legendary":
		return "Legendary"
	case "champion":
		return "Champion"
	default:
		// Return original if no match, or could capitalize first letter
		if len(rarity) == 0 {
			return rarity
		}
		return strings.Title(strings.ToLower(rarity))
	}
}

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

// Max levels for each rarity
var maxLevels = map[string]int{
	"Common":    16,
	"Rare":      16,
	"Epic":      16,
	"Legendary": 16,
	"Champion":  16,
}

// Starting levels for each rarity (when first unlocked)

var startingLevels = map[string]int{

	"Common": 1,

	"Rare": 3,

	"Epic": 6,

	"Legendary": 9,

	"Champion": 11,
}

// NewCardAdapter creates a CardInfo from a rarity string
// This can be used when converting from external card types
func NewCardAdapter(rarity string) CardInfo {
	return cardAdapter{rarity: rarity}
}

// CalculateCardsNeeded returns how many cards are needed to upgrade from currentLevel

// Returns 0 if already at max level, below starting level, or invalid rarity

func CalculateCardsNeeded(currentLevel int, rarity string) int {

	rarity = NormalizeRarity(rarity)

	// Check if already at max level

	maxLevel := GetMaxLevel(rarity)

	if currentLevel >= maxLevel {

		return 0 // Valid: already at max level

	}

	// Check if level is below starting level for this rarity

	startingLevel := GetStartingLevel(rarity)

	if currentLevel < startingLevel {

		return 0 // Valid: card not unlocked yet

	}

	costs, exists := upgradeCosts[rarity]

	if !exists {

		return 0 // Invalid rarity

	}

	cardsNeeded, exists := costs[currentLevel]

	if !exists {

		// If below max but no entry, return 2 as a safe default for low levels

		if currentLevel < maxLevel {

			return 2

		}

		return 0

	}

	return cardsNeeded

}

// GetMaxLevel returns the maximum level for a given rarity

func GetMaxLevel(rarity string) int {

	rarity = NormalizeRarity(rarity)

	if maxLevel, exists := maxLevels[rarity]; exists {

		return maxLevel

	}

	return 16 // Default to 16 if unknown

}

// GetStartingLevel returns the initial level when a card is first unlocked

func GetStartingLevel(rarity string) int {

	rarity = NormalizeRarity(rarity)

	if startLevel, exists := startingLevels[rarity]; exists {

		return startLevel

	}

	return 1 // Default to 1 if unknown

}

// IsMaxLevel checks if a card is at maximum level for its rarity

func IsMaxLevel(currentLevel int, rarity string) bool {

	rarity = NormalizeRarity(rarity)

	return currentLevel >= GetMaxLevel(rarity)

}

// CalculateTotalCardsToMax calculates total cards needed from current level to max

func CalculateTotalCardsToMax(currentLevel int, rarity string) int {

	rarity = NormalizeRarity(rarity)

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

		} else {

			// If below max but no entry, assume 2 for low levels

			total += 2

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
	CardName string `json:"card_name"`

	Rarity string `json:"rarity"`

	ElixirCost int `json:"elixir_cost,omitempty"`

	CurrentLevel int `json:"current_level"`

	MaxLevel int `json:"max_level"`

	EvolutionLevel int `json:"evolution_level,omitempty"`

	IsMaxLevel bool `json:"is_max_level"`

	CardsOwned int `json:"cards_owned"`

	CardsToNextLevel int `json:"cards_to_next_level"`

	CardsRemaining int `json:"cards_remaining"`

	ProgressPercent float64 `json:"progress_percent"`

	CanUpgradeNow bool `json:"can_upgrade_now"`

	TotalToMax int `json:"total_to_max"`

	LevelsToMax int `json:"levels_to_max"`

	MaxEvolutionLevel int `json:"max_evolution_level,omitempty"`
}

// CalculateUpgradeInfo creates a complete UpgradeInfo for a card

func CalculateUpgradeInfo(cardName string, rarity string, elixirCost int, currentLevel int, cardsOwned int, evolutionLevel int, maxEvolutionLevel int, apiMaxLevel int) UpgradeInfo {

	rarity = NormalizeRarity(rarity)

	maxLevel := apiMaxLevel

	if maxLevel == 0 {

		maxLevel = GetMaxLevel(rarity)

	}

	isMaxLevel := currentLevel >= maxLevel

	var cardsNeeded int

	if isMaxLevel {

		cardsNeeded = 0

	} else {

		// Use a local copy of GetMaxLevel/CalculateCardsNeeded logic that respects our dynamic maxLevel

		cardsNeeded = CalculateCardsNeeded(currentLevel, rarity)

		// If CalculateCardsNeeded thinks it's not max but our dynamic maxLevel says it is, override

		if currentLevel >= maxLevel {

			cardsNeeded = 0

		}

	}

	totalToMax := CalculateTotalCardsToMax(currentLevel, rarity)

	if currentLevel >= maxLevel {

		totalToMax = 0

	}

	cardsRemaining := cardsNeeded - cardsOwned

	if cardsRemaining < 0 {

		cardsRemaining = 0

	}

	progress := CalculateUpgradeProgress(cardsOwned, cardsNeeded)

	canUpgrade := cardsNeeded > 0 && cardsOwned >= cardsNeeded && !isMaxLevel

	return UpgradeInfo{

		CardName: cardName,

		Rarity: rarity,

		ElixirCost: elixirCost,

		CurrentLevel: currentLevel,

		MaxLevel: maxLevel,

		EvolutionLevel: evolutionLevel,

		IsMaxLevel: isMaxLevel,

		CardsOwned: cardsOwned,

		CardsToNextLevel: cardsNeeded,

		CardsRemaining: cardsRemaining,

		ProgressPercent: progress,

		CanUpgradeNow: canUpgrade,

		TotalToMax: totalToMax,

		LevelsToMax: maxLevel - currentLevel,

		MaxEvolutionLevel: maxEvolutionLevel,
	}

}

// RarityUpgradeStats contains aggregate statistics for a rarity

type RarityUpgradeStats struct {
	Rarity string `json:"rarity"`

	TotalCards int `json:"total_cards"`

	MaxLevelCards int `json:"max_level_cards"`

	UpgradableCards int `json:"upgradable_cards"`

	AvgLevel float64 `json:"avg_level"`

	AvgProgressPercent float64 `json:"avg_progress_percent"`

	TotalCardsNeeded int `json:"total_cards_needed"` // Total cards needed for all upgrades

	CompletionPercent float64 `json:"completion_percent"` // % of cards at max level

}

// CalculateRarityStats computes aggregate statistics for cards of a specific rarity

func CalculateRarityStats(cards []UpgradeInfo, rarity string) RarityUpgradeStats {

	rarity = NormalizeRarity(rarity)

	filtered := make([]UpgradeInfo, 0)

	for _, card := range cards {

		if NormalizeRarity(card.Rarity) == rarity {

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

		Rarity: rarity,

		TotalCards: cardCount,

		MaxLevelCards: maxLevelCount,

		UpgradableCards: upgradableCount,

		AvgLevel: float64(totalLevel) / float64(cardCount),

		AvgProgressPercent: totalProgress / float64(cardCount),

		TotalCardsNeeded: totalNeeded,

		CompletionPercent: (float64(maxLevelCount) / float64(cardCount)) * 100.0,
	}

}

// CalculatePriorityScore computes an upgrade priority score (0-100)

// Higher score = higher priority for upgrading

// Factors:

// - Proximity to next level (50% weight)

// - Current level ratio (30% weight)

// - Rarity boost (20% weight)

// - Evolution capability (bonus 10-30 points)

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

		"Common": 0.0,

		"Rare": 20.0,

		"Epic": 40.0,

		"Legendary": 60.0,

		"Champion": 80.0,
	}

	rarityScore, exists := rarityScores[NormalizeRarity(info.Rarity)]

	if !exists {

		rarityScore = 0.0

	}

	// Weighted combination (base score without evolution)
	priorityScore := (proximityScore * 0.5) + (levelScore * 0.3) + (rarityScore * 0.2)

	// Evolution bonus: cards with evolution capability get higher priority
	if info.MaxEvolutionLevel > 0 {
		// Evolution bonus based on proximity to max level (evolution is only useful at max)
		evolutionBonus := 0.0

		// Base bonus for having evolution capability
		evolutionBonus += 10.0

		// Additional bonus based on level ratio (more useful when closer to max)
		evolutionBonus += levelRatio * 20.0

		// Bonus for evolution progress (partially evolved cards get slight boost)
		if info.EvolutionLevel > 0 {
			evoRatio := float64(info.EvolutionLevel) / float64(info.MaxEvolutionLevel)
			evolutionBonus += evoRatio * 5.0 // Up to 5 extra points
		}

		priorityScore += evolutionBonus
	}

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

// CardCountConfig provides immutable card count configuration
// This replaces the global mutable totalCardsPerRarity map for better testability
type CardCountConfig struct {
	cardCounts map[string]int
}

// NewCardCountConfig creates a config from actual card data
// Counts cards by rarity and applies defaults for rarities with zero cards
func NewCardCountConfig(cards []CardInfo) *CardCountConfig {
	counts := make(map[string]int)

	// Count cards by rarity
	for _, card := range cards {
		rarity := NormalizeRarity(card.GetRarity())
		if rarity != "" {
			counts[rarity]++
		}
	}

	// Apply defaults for missing rarities (fallback to game defaults)
	defaults := map[string]int{
		"Common":    19,
		"Rare":      20,
		"Epic":      12,
		"Legendary": 10,
		"Champion":  6,
	}

	for rarity, defaultVal := range defaults {
		if counts[rarity] == 0 {
			counts[rarity] = defaultVal
		}
	}

	return &CardCountConfig{cardCounts: counts}
}

// DefaultCardCountConfig returns a config with game default card counts
// Use this when actual card data is not available
func DefaultCardCountConfig() *CardCountConfig {
	return &CardCountConfig{
		cardCounts: map[string]int{
			"Common":    19,
			"Rare":      20,
			"Epic":      12,
			"Legendary": 10,
			"Champion":  6,
		},
	}
}

// GetTotalCards returns the total number of cards for a given rarity
// This method is thread-safe for concurrent read access
func (c *CardCountConfig) GetTotalCards(rarity string) int {
	if c == nil {
		return 0
	}

	normalized := NormalizeRarity(rarity)
	if count, ok := c.cardCounts[normalized]; ok {
		return count
	}

	return 0
}
