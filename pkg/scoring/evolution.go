// Package scoring provides implementations of the Scorer interface for
// various card scoring algorithms.
package scoring

import (
	"math"
)

// EvolutionScorer implements the Scorer interface for evolution-aware scoring.
// It provides bonus scores for cards based on their evolution level and the
// player's unlocked evolutions. This scorer supports both flat (linear) and
// level-scaled bonus calculations.
//
// # Scoring Formula (Flat Mode)
//
//	evolutionBonus = baseBonus × (evolutionLevel / maxEvolutionLevel)
//
// # Scoring Formula (Level-Scaled Mode)
//
//	levelRatio = cardLevel / maxCardLevel
//	scaledRatio = levelRatio^1.5
//	evoMultiplier = 1 + 0.2 × (maxEvolutionLevel - 1)
//	evolutionBonus = baseBonus × scaledRatio × evoMultiplier
//
// The level-scaled mode rewards higher-level cards more and accounts for
// multi-evolution cards (e.g., Knight with evo level 3).
//
// Default weights:
//   - baseBonus: 0.25 (maximum bonus for fully evolved, max-level card)
//   - levelScalingExponent: 1.5 (for level-scaled calculation)
//   - multiEvoMultiplier: 0.2 (bonus per additional evolution level)
type EvolutionScorer struct {
	// unlockedEvolutions tracks which evolutions are available to the player.
	// Maps card name to true if evolution is unlocked.
	unlockedEvolutions map[string]bool

	// baseBonus is the maximum bonus for a fully evolved card at max level.
	// Default 0.25.
	baseBonus float64

	// levelScalingExponent determines how strongly card level affects the bonus.
	// Higher values reward high-level cards more.
	// Default 1.5.
	levelScalingExponent float64

	// multiEvoMultiplier is the bonus multiplier for multi-evolution cards.
	// Applied as: 1 + multiEvoMultiplier × (maxEvolutionLevel - 1)
	// Default 0.2.
	multiEvoMultiplier float64

	// levelScaled enables level-scaled bonus calculation.
	// If false, uses flat (linear) bonus calculation.
	// Default true.
	levelScaled bool
}

// EvolutionScorerConfig configures an EvolutionScorer with custom parameters.
type EvolutionScorerConfig struct {
	// UnlockedEvolutions tracks which evolutions are available to the player.
	// Maps card name to true if evolution is unlocked.
	UnlockedEvolutions map[string]bool

	// BaseBonus is the maximum bonus for a fully evolved card (default 0.25).
	BaseBonus float64

	// LevelScalingExponent determines level impact (default 1.5).
	LevelScalingExponent float64

	// MultiEvoMultiplier is the bonus per additional evolution level (default 0.2).
	MultiEvoMultiplier float64

	// LevelScaled enables level-scaled calculation (default true).
	LevelScaled bool
}

// NewEvolutionScorer creates a new EvolutionScorer with default parameters.
//
// Example:
//
//	unlocked := map[string]bool{"Archers": true, "Knight": true}
//	scorer := NewEvolutionScorer(unlocked)
//	score := scorer.Score(candidate, config)
func NewEvolutionScorer(unlockedEvolutions map[string]bool) *EvolutionScorer {
	return &EvolutionScorer{
		unlockedEvolutions:   unlockedEvolutions,
		baseBonus:            0.25,
		levelScalingExponent: 1.5,
		multiEvoMultiplier:   0.2,
		levelScaled:          true,
	}
}

// NewEvolutionScorerWithConfig creates a new EvolutionScorer with custom configuration.
//
// Example:
//
//	config := EvolutionScorerConfig{
//	    UnlockedEvolutions: unlocked,
//	    BaseBonus: 0.3,  // Higher bonus
//	    LevelScaled: false,  // Use flat calculation
//	}
//	scorer := NewEvolutionScorerWithConfig(config)
func NewEvolutionScorerWithConfig(config EvolutionScorerConfig) *EvolutionScorer {
	baseBonus := config.BaseBonus
	if baseBonus == 0 {
		baseBonus = 0.25
	}

	levelScalingExponent := config.LevelScalingExponent
	if levelScalingExponent == 0 {
		levelScalingExponent = 1.5
	}

	multiEvoMultiplier := config.MultiEvoMultiplier
	if multiEvoMultiplier == 0 {
		multiEvoMultiplier = 0.2
	}

	return &EvolutionScorer{
		unlockedEvolutions:   config.UnlockedEvolutions,
		baseBonus:            baseBonus,
		levelScalingExponent: levelScalingExponent,
		multiEvoMultiplier:   multiEvoMultiplier,
		levelScaled:          config.LevelScaled,
	}
}

// Score calculates the evolution bonus for a card candidate.
//
// The bonus depends on:
//   - Whether the evolution is unlocked for this card
//   - The card's current evolution level
//   - The card's level (for level-scaled mode)
//   - The maximum evolution level (for multi-evolution cards)
//
// Returns 0 if:
//   - Evolution is not unlocked for this card
//   - Card has no evolution capability (maxEvolutionLevel == 0)
//   - Card has no evolution progress (evolutionLevel == 0)
//
// Returns a score typically in the range 0.0 to ~0.3, higher is better.
func (s *EvolutionScorer) Score(candidate CardCandidate, config ScoringConfig) float64 {
	// Use config's unlocked evolutions if provided, otherwise use scorer's default
	unlockedEvolutions := config.UnlockedEvolutions
	if unlockedEvolutions == nil {
		unlockedEvolutions = s.unlockedEvolutions
	}

	// Check if evolution is unlocked for this card
	if !s.isEvolutionUnlocked(candidate.Name, unlockedEvolutions) {
		return 0.0
	}

	// Check if card has evolution capability
	if candidate.MaxEvolutionLevel == 0 {
		return 0.0
	}

	// Check if card has evolution progress
	if candidate.EvolutionLevel == 0 {
		return 0.0
	}

	if s.levelScaled {
		return s.calculateLevelScaledBonus(candidate)
	}
	return s.calculateFlatBonus(candidate)
}

// calculateFlatBonus calculates evolution bonus using linear progression.
// Formula: baseBonus × (evolutionLevel / maxEvolutionLevel)
func (s *EvolutionScorer) calculateFlatBonus(candidate CardCandidate) float64 {
	evolutionRatio := float64(candidate.EvolutionLevel) / float64(candidate.MaxEvolutionLevel)

	// Clamp ratio to valid range
	if evolutionRatio > 1.0 {
		evolutionRatio = 1.0
	}

	return s.baseBonus * evolutionRatio
}

// calculateLevelScaledBonus calculates evolution bonus using level-scaled progression.
// Formula: baseBonus × (level/maxLevel)^1.5 × (1 + 0.2 × (maxEvoLevel - 1))
//
// This rewards higher-level cards more and accounts for multi-evolution cards.
func (s *EvolutionScorer) calculateLevelScaledBonus(candidate CardCandidate) float64 {
	// Calculate level ratio
	if candidate.MaxLevel == 0 {
		return 0.0
	}

	levelRatio := float64(candidate.Level) / float64(candidate.MaxLevel)
	scaledRatio := math.Pow(levelRatio, s.levelScalingExponent)

	// Bonus multiplier for multi-evolution cards
	evoMultiplier := 1.0 + (s.multiEvoMultiplier * float64(candidate.MaxEvolutionLevel-1))

	bonus := s.baseBonus * scaledRatio * evoMultiplier

	// Apply evolution progress ratio
	evolutionRatio := float64(candidate.EvolutionLevel) / float64(candidate.MaxEvolutionLevel)
	if evolutionRatio > 1.0 {
		evolutionRatio = 1.0
	}

	return bonus * evolutionRatio
}

// isEvolutionUnlocked checks if an evolution is unlocked for a card.
func (s *EvolutionScorer) isEvolutionUnlocked(cardName string, unlocked map[string]bool) bool {
	if unlocked == nil {
		return false
	}
	return unlocked[cardName]
}

// SetUnlockedEvolutions updates the unlocked evolutions map.
// This allows runtime modification of unlocked evolutions.
func (s *EvolutionScorer) SetUnlockedEvolutions(unlocked map[string]bool) {
	s.unlockedEvolutions = unlocked
}

// GetUnlockedEvolutions returns the current unlocked evolutions map.
func (s *EvolutionScorer) GetUnlockedEvolutions() map[string]bool {
	return s.unlockedEvolutions
}

// SetLevelScaled enables or disables level-scaled bonus calculation.
func (s *EvolutionScorer) SetLevelScaled(levelScaled bool) {
	s.levelScaled = levelScaled
}

// IsLevelScaled returns whether level-scaled calculation is enabled.
func (s *EvolutionScorer) IsLevelScaled() bool {
	return s.levelScaled
}

// SetBaseBonus updates the base bonus value.
func (s *EvolutionScorer) SetBaseBonus(bonus float64) {
	s.baseBonus = bonus
}

// GetBaseBonus returns the current base bonus.
func (s *EvolutionScorer) GetBaseBonus() float64 {
	return s.baseBonus
}

// SetLevelScalingExponent updates the level scaling exponent.
func (s *EvolutionScorer) SetLevelScalingExponent(exponent float64) {
	s.levelScalingExponent = exponent
}

// GetLevelScalingExponent returns the current level scaling exponent.
func (s *EvolutionScorer) GetLevelScalingExponent() float64 {
	return s.levelScalingExponent
}

// HasEvolutionOverride checks if a card has an evolution-specific role override.
// These cards have their strategic role changed by evolution, making them
// more valuable in deck building.
//
// This is determined by checking against known evolution overrides in the
// deck package.
func HasEvolutionOverride(cardName string, evolutionLevel int) bool {
	// This would integrate with deck.evolutionOverrides from the original code
	// For now, return a placeholder
	// TODO(clash-royale-api-42nu): Integrate with deck.HasEvolutionOverride when available
	return false
}

// CalculateEvolutionBonusForCard is a convenience function for calculating
// evolution bonus for a single card without creating a scorer instance.
//
// Uses flat bonus calculation by default.
func CalculateEvolutionBonusForCard(
	cardName string,
	cardLevel, maxCardLevel int,
	evolutionLevel, maxEvolutionLevel int,
	unlockedEvolutions map[string]bool,
) float64 {
	candidate := CardCandidate{
		Name:              cardName,
		Level:             cardLevel,
		MaxLevel:          maxCardLevel,
		EvolutionLevel:    evolutionLevel,
		MaxEvolutionLevel: maxEvolutionLevel,
	}

	scorer := NewEvolutionScorer(unlockedEvolutions)
	scorer.SetLevelScaled(false) // Use flat calculation

	config := DefaultScoringConfig()
	return scorer.Score(candidate, config)
}
