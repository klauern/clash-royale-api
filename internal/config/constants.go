// Package config provides centralized game configuration constants and mappings.
// This file consolidates scoring-related constants that were previously scattered
// across multiple packages (deck, recommend).
package config

// Card Scoring Constants
// These constants control how cards are scored for deck building

const (
	// LevelWeightFactor amplifies the level ratio's contribution to card score
	// Higher values make card level more important in deck building
	// Formula impact: levelRatio * LevelWeightFactor * rarityBoost
	LevelWeightFactor = 1.2

	// RoleBonusValue is the flat bonus added when a card has a defined strategic role
	// Cards with roles (win condition, support, etc.) get this small boost
	RoleBonusValue = 0.05
)

// Combat Stats Integration Constants
// These constants control how combat statistics are weighted in card scoring

const (
	// DefaultCombatWeight is the default weight for combat stats in scoring (0.0-1.0)
	// 0.25 means 25% combat stats, 75% base scoring (level, rarity, elixir)
	// Can be overridden via COMBAT_STATS_WEIGHT environment variable
	DefaultCombatWeight = 0.25

	// CombatDPSWeight is the percentage of combat score from DPS efficiency
	// 40% of combat score comes from damage-per-second per elixir
	CombatDPSWeight = 0.4

	// CombatHPWeight is the percentage of combat score from HP efficiency
	// 40% of combat score comes from hit-points per elixir
	CombatHPWeight = 0.4

	// CombatRoleWeight is the percentage of combat score from role effectiveness
	// 20% of combat score comes from how well stats match the card's role
	CombatRoleWeight = 0.2
)

// Evolution Bonus Constants
// These constants control how evolution levels affect card scoring

const (
	// EvolutionBonusWeight is the base weight for evolution level bonus
	// Maximum bonus at full evolution: EvolutionBonusWeight (0.15 = up to +0.15 score)
	// This is the linear weight used in new evolution scoring (ScoreCardWithEvolution)
	EvolutionBonusWeight = 0.15

	// EvolutionBaseBonus is the base bonus for unlocked evolution cards
	// Used in level-scaled evolution bonus calculation in deck builder
	// Formula: baseBonus * (level/maxLevel)^1.5 * (1 + 0.2*(maxEvoLevel-1))
	EvolutionBaseBonus = 0.25
)

// Deck Recommendation Scoring Weights
// These constants control how different factors contribute to overall deck recommendation score

const (
	// RecommendationWeightCompatibility is the weight for card level compatibility (60%)
	// Card levels matter most for ladder viability
	RecommendationWeightCompatibility = 0.60

	// RecommendationWeightSynergy is the weight for card pair synergy (25%)
	// Synergy improves win rate but levels matter more
	RecommendationWeightSynergy = 0.25

	// RecommendationWeightArchetypeFit is the weight for archetype matching (15%)
	// Following proven patterns has value but less than card strength
	RecommendationWeightArchetypeFit = 0.15
)
