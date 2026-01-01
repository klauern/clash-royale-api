// Package scoring provides implementations of the Scorer interface for
// various card scoring algorithms.
package scoring

import (
	"math"

	"github.com/klauer/clash-royale-api/go/internal/config"
)

// BaseScorer implements the Scorer interface with traditional card scoring.
// It considers level ratio, rarity boost, elixir cost, role bonus, and
// evolution level. This scorer provides a solid baseline for deck building
// and can be combined with other scorers via CompositeScorer.
//
// # Scoring Formula
//
//	score = (levelRatio × levelWeight × rarityBoost) +
//	       (elixirWeight × elixirWeightFactor) +
//	       roleBonus +
//	       evolutionBonus
//
// Where:
//   - levelRatio: curve-based level progression (0.0 to 1.0)
//   - rarityBoost: multiplier based on card rarity (1.0 to 1.2)
//   - elixirWeight: penalizes cards far from optimal elixir cost (0.0 to 1.0)
//   - roleBonus: fixed bonus for cards with defined roles (0.0 or 0.05)
//   - evolutionBonus: proportional to evolution progress (0.0 to ~0.15)
//
// Default weights:
//   - levelWeightFactor: 1.2
//   - elixirWeightFactor: 0.15
//   - roleBonusValue: 0.05
//   - evolutionBonusWeight: 0.15
type BaseScorer struct {
	// levelCurve provides curve-based level calculation.
	// If nil, falls back to linear level ratio.
	levelCurve LevelCurve

	// rarityWeights defines the boost multiplier for each rarity.
	// Rarer cards get higher boost since they're harder to level.
	rarityWeights map[string]float64

	// levelWeightFactor determines how much card level impacts the score.
	// Higher values prioritize well-leveled cards.
	levelWeightFactor float64

	// elixirOptimal is the ideal elixir cost for cards (default 3.0).
	elixirOptimal float64

	// elixirWeightFactor determines how much elixir efficiency impacts the score.
	// Higher values penalize high-cost cards more heavily.
	elixirWeightFactor float64

	// roleBonusValue is the fixed bonus for cards with defined roles.
	// Cards without roles get no bonus.
	roleBonusValue float64

	// evolutionBonusWeight is the maximum bonus for fully evolved cards.
	// Scaled by evolution progress: evolutionBonusWeight × (evolutionLevel / maxEvolutionLevel)
	evolutionBonusWeight float64
}

// BaseScorerConfig configures a BaseScorer with custom weights and parameters.
type BaseScorerConfig struct {
	// LevelCurve provides curve-based level calculation.
	// If nil, linear level ratio is used.
	LevelCurve LevelCurve

	// RarityWeights defines rarity boost multipliers.
	// If nil, uses default weights.
	RarityWeights map[string]float64

	// LevelWeightFactor determines level importance (default 1.2).
	LevelWeightFactor float64

	// ElixirOptimal is the target elixir cost (default 3.0).
	ElixirOptimal float64

	// ElixirWeightFactor determines elixir efficiency importance (default 0.15).
	ElixirWeightFactor float64

	// RoleBonusValue is the fixed bonus for cards with roles (default 0.05).
	RoleBonusValue float64

	// EvolutionBonusWeight is the max evolution bonus (default 0.15).
	EvolutionBonusWeight float64
}

// DefaultRarityWeights returns the standard rarity weight map.
// Rarer cards get higher boost since they're harder to level up.
// This function delegates to the centralized config package.
func DefaultRarityWeights() map[string]float64 {
	return map[string]float64{
		"Common":    config.GetRarityWeight("Common"),
		"Rare":      config.GetRarityWeight("Rare"),
		"Epic":      config.GetRarityWeight("Epic"),
		"Legendary": config.GetRarityWeight("Legendary"),
		"Champion":  config.GetRarityWeight("Champion"),
	}
}

// NewBaseScorer creates a new BaseScorer with default parameters.
//
// Example:
//
//	scorer := NewBaseScorer(levelCurve)
//	score := scorer.Score(candidate, config)
func NewBaseScorer(levelCurve LevelCurve) *BaseScorer {
	return &BaseScorer{
		levelCurve:           levelCurve,
		rarityWeights:        DefaultRarityWeights(),
		levelWeightFactor:    1.2,
		elixirOptimal:        3.0,
		elixirWeightFactor:   0.15,
		roleBonusValue:       0.05,
		evolutionBonusWeight: 0.15,
	}
}

// NewBaseScorerWithConfig creates a new BaseScorer with custom configuration.
//
// Example:
//
//	config := BaseScorerConfig{
//	    LevelCurve: levelCurve,
//	    LevelWeightFactor: 1.5,  // Prioritize level more
//	    ElixirWeightFactor: 0.2,  // Penalize high cost more
//	}
//	scorer := NewBaseScorerWithConfig(config)
func NewBaseScorerWithConfig(config BaseScorerConfig) *BaseScorer {
	rarityWeights := config.RarityWeights
	if rarityWeights == nil {
		rarityWeights = DefaultRarityWeights()
	}

	levelWeightFactor := config.LevelWeightFactor
	if levelWeightFactor == 0 {
		levelWeightFactor = 1.2
	}

	elixirOptimal := config.ElixirOptimal
	if elixirOptimal == 0 {
		elixirOptimal = 3.0
	}

	elixirWeightFactor := config.ElixirWeightFactor
	if elixirWeightFactor == 0 {
		elixirWeightFactor = 0.15
	}

	roleBonusValue := config.RoleBonusValue
	if roleBonusValue == 0 {
		roleBonusValue = 0.05
	}

	evolutionBonusWeight := config.EvolutionBonusWeight
	if evolutionBonusWeight == 0 {
		evolutionBonusWeight = 0.15
	}

	return &BaseScorer{
		levelCurve:           config.LevelCurve,
		rarityWeights:        rarityWeights,
		levelWeightFactor:    levelWeightFactor,
		elixirOptimal:        elixirOptimal,
		elixirWeightFactor:   elixirWeightFactor,
		roleBonusValue:       roleBonusValue,
		evolutionBonusWeight: evolutionBonusWeight,
	}
}

// Score calculates the base score for a card candidate.
//
// The score combines multiple factors:
//   - Level ratio (curve-based if LevelCurve is provided, else linear)
//   - Rarity boost (rarer cards get higher multiplier)
//   - Elixir efficiency (cards around 3-4 elixir score better)
//   - Role bonus (cards with defined roles get small bonus)
//   - Evolution bonus (evolved cards get additional score)
//
// Returns a score typically in the range 0.0 to ~1.65, higher is better.
func (s *BaseScorer) Score(candidate CardCandidate, config ScoringConfig) float64 {
	// Use config's level curve if provided, otherwise use scorer's default
	levelCurve := config.LevelCurve
	if levelCurve == nil {
		levelCurve = s.levelCurve
	}

	// Calculate level ratio (curve-based or linear)
	levelRatio := s.calculateLevelRatio(candidate.Name, candidate.Level, candidate.MaxLevel, levelCurve)

	// Get rarity boost multiplier
	rarityBoost := s.getRarityBoost(candidate.Rarity)

	// Calculate elixir efficiency weight
	elixirWeight := s.calculateElixirWeight(candidate.Elixir)

	// Calculate role bonus
	roleBonus := 0.0
	if candidate.Role != nil {
		roleBonus = s.roleBonusValue
	}

	// Calculate evolution bonus
	evolutionBonus := s.calculateEvolutionBonus(candidate.EvolutionLevel, candidate.MaxEvolutionLevel)

	// Combine all factors
	score := (levelRatio * s.levelWeightFactor * rarityBoost) +
		(elixirWeight * s.elixirWeightFactor) +
		roleBonus +
		evolutionBonus

	return score
}

// calculateLevelRatio calculates the level ratio for a card.
// Uses curve-based calculation when levelCurve is available and card name is provided.
// Falls back to linear calculation for backward compatibility.
func (s *BaseScorer) calculateLevelRatio(cardName string, level, maxLevel int, levelCurve LevelCurve) float64 {
	if maxLevel <= 0 {
		return 0.0
	}

	// Use curve-based calculation if available and card name provided
	if levelCurve != nil && cardName != "" {
		return levelCurve.GetRelativeLevelRatio(cardName, level, maxLevel)
	}

	// Fall back to linear calculation
	return float64(level) / float64(maxLevel)
}

// getRarityBoost returns the rarity multiplier for a card.
// Returns 1.0 (Common) if rarity is unknown.
func (s *BaseScorer) getRarityBoost(rarity string) float64 {
	if boost, exists := s.rarityWeights[rarity]; exists {
		return boost
	}
	return 1.0 // Default to Common
}

// calculateElixirWeight calculates elixir efficiency score.
// Penalizes cards far from optimal elixir cost (default 3).
// Returns a value between 0.0 and ~1.0.
func (s *BaseScorer) calculateElixirWeight(elixir int) float64 {
	elixirDiff := math.Abs(float64(elixir) - s.elixirOptimal)
	return 1.0 - (elixirDiff / 9.0) // 9 is max meaningful diff
}

// calculateEvolutionBonus calculates the evolution level bonus for a card.
// The bonus is proportional to the evolution progress (evolutionLevel/maxEvolutionLevel).
//
// Formula: evolutionBonusWeight × (evolutionLevel / maxEvolutionLevel)
//
// Returns 0 if card has no evolution capability (maxEvolutionLevel == 0) or no evolution progress.
func (s *BaseScorer) calculateEvolutionBonus(evolutionLevel, maxEvolutionLevel int) float64 {
	if maxEvolutionLevel <= 0 || evolutionLevel <= 0 {
		return 0.0
	}

	// Calculate evolution ratio (0.0 to 1.0)
	evolutionRatio := float64(evolutionLevel) / float64(maxEvolutionLevel)

	// Clamp ratio to valid range
	if evolutionRatio > 1.0 {
		evolutionRatio = 1.0
	}

	// Apply evolution bonus weight
	return s.evolutionBonusWeight * evolutionRatio
}

// SetLevelCurve updates the level curve used for scoring.
// This allows runtime modification of the curve.
func (s *BaseScorer) SetLevelCurve(levelCurve LevelCurve) {
	s.levelCurve = levelCurve
}

// GetLevelCurve returns the current level curve.
func (s *BaseScorer) GetLevelCurve() LevelCurve {
	return s.levelCurve
}

// SetRarityWeights updates the rarity weights map.
// This allows runtime modification of rarity bonuses.
func (s *BaseScorer) SetRarityWeights(weights map[string]float64) {
	if weights != nil {
		s.rarityWeights = weights
	}
}

// GetRarityWeights returns the current rarity weights.
func (s *BaseScorer) GetRarityWeights() map[string]float64 {
	return s.rarityWeights
}
