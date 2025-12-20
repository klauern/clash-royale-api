package archetypes

import (
	"math"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// CalculateUpgradeCosts calculates upgrade requirements for all cards in a deck.
// It returns per-card upgrades, total cards needed, and total gold needed.
func CalculateUpgradeCosts(
	recommendation *deck.DeckRecommendation,
	targetLevel int,
) ([]CardUpgrade, int, int) {
	upgrades := make([]CardUpgrade, len(recommendation.DeckDetail))
	totalCards := 0
	totalGold := 0

	for i, card := range recommendation.DeckDetail {
		upgrade := CardUpgrade{
			CardName:     card.Name,
			CurrentLevel: card.Level,
			TargetLevel:  targetLevel,
			Rarity:       card.Rarity,
			LevelGap:     targetLevel - card.Level,
		}

		// Calculate cards and gold needed from current to target level
		if card.Level < targetLevel {
			upgrade.CardsNeeded = calculateCardsToLevel(
				card.Level,
				targetLevel,
				card.Rarity,
			)
			upgrade.GoldNeeded = deck.CalculateGoldNeeded(
				card.Level,
				targetLevel,
				card.Rarity,
			)
		}

		upgrades[i] = upgrade
		totalCards += upgrade.CardsNeeded
		totalGold += upgrade.GoldNeeded
	}

	return upgrades, totalCards, totalGold
}

// CalculateDistanceMetric computes how far a deck is from ideal competitive state.
// Returns 0.0 (perfect - all cards at target level) to 1.0 (very far from target).
//
// The metric uses weighted averaging where more important card roles (like win
// conditions) are weighted more heavily than less critical roles (like cycle cards).
func CalculateDistanceMetric(
	recommendation *deck.DeckRecommendation,
	targetLevel int,
) float64 {
	if len(recommendation.DeckDetail) == 0 {
		return 1.0
	}

	// Calculate weighted level gap
	totalWeightedGap := 0.0
	totalWeight := 0.0

	for _, card := range recommendation.DeckDetail {
		weight := getRoleWeight(card.Role)
		levelGap := float64(targetLevel - card.Level)
		if levelGap < 0 {
			levelGap = 0 // Already above target
		}

		// Normalize by max possible gap (14 levels max)
		normalizedGap := levelGap / 14.0

		totalWeightedGap += normalizedGap * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	// Return normalized distance (0-1 range)
	distance := totalWeightedGap / totalWeight
	return math.Min(distance, 1.0)
}

// getRoleWeight returns importance weight for a card role.
// Win conditions are most critical (2.0x), cycle cards least critical (0.8x).
func getRoleWeight(roleStr string) float64 {
	role := deck.CardRole(roleStr)
	weights := map[deck.CardRole]float64{
		deck.RoleWinCondition: 2.0, // Most important - primary damage dealer
		deck.RoleSpellBig:     1.5, // High value spells
		deck.RoleSpellSmall:   1.5, // Essential utility spells
		deck.RoleSupport:      1.2, // Supporting offense/defense
		deck.RoleBuilding:     1.0, // Defensive structures
		deck.RoleCycle:        0.8, // Least critical - cycle cards
	}

	if weight, exists := weights[role]; exists {
		return weight
	}
	return 1.0 // Default weight for unknown roles
}

// calculateCardsToLevel calculates total cards needed from currentLevel to targetLevel.
// Uses the existing upgrade calculator from the analysis package.
func calculateCardsToLevel(currentLevel, targetLevel int, rarity string) int {
	if currentLevel >= targetLevel {
		return 0
	}

	total := 0
	for level := currentLevel; level < targetLevel; level++ {
		total += analysis.CalculateCardsNeeded(level, rarity)
	}
	return total
}

// calculateAvgLevel calculates average card level for a deck.
// Used by the analyzer to determine current deck strength.
func calculateAvgLevel(cards []deck.CardDetail) float64 {
	if len(cards) == 0 {
		return 0
	}

	total := 0
	for _, card := range cards {
		total += card.Level
	}

	return float64(total) / float64(len(cards))
}
