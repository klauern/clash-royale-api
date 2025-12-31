// Package scoring provides a unified interface for card scoring algorithms.
// It supports multiple scoring dimensions including base card stats, combat effectiveness,
// evolution bonuses, and card synergies. The interface is designed for extensibility,
// allowing new scoring algorithms to be added without modifying existing code.
//
// # Architecture Overview
//
// The scoring system is built around the Scorer interface, which defines a single
// Score method for calculating card scores. Scorer implementations can be:
//   - Simple: BaseScorer for traditional level/rarity/elixir scoring
//   - Advanced: CombatScorer for stats-based scoring
//   - Contextual: EvolutionScorer for evolution-aware bonuses
//   - Composite: CompositeScorer combining multiple scorers with weights
//
// # Basic Usage
//
//	baseScorer := NewBaseScorer(levelCurve, rarityWeights)
//	score := baseScorer.Score(candidate, config)
//
// # Composite Scoring
//
//	composite := NewCompositeScorer([]WeightedScorer{
//	    {Scorer: baseScorer, Weight: 0.70},
//	    {Scorer: combatScorer, Weight: 0.25},
//	    {Scorer: evolutionScorer, Weight: 0.05},
//	})
//	score := composite.Score(candidate, config)
//
// # Extending with Custom Scorers
//
//	type MyCustomScorer struct{}
//
//	func (s *MyCustomScorer) Score(candidate CardCandidate, config ScoringConfig) float64 {
//	    // Your custom scoring logic
//	    return 0.0
//	}
package scoring

import (
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// CardCandidate represents a card being evaluated for scoring.
// It is intentionally re-exported from the deck package to avoid circular dependencies.
// The scoring package operates on this type without importing deck directly.
type CardCandidate = deck.CardCandidate

// LevelCurve defines the interface for card level curve calculations.
// This allows the scoring system to work with curve-based level calculations
// rather than simple linear ratios.
type LevelCurve interface {
	// GetRelativeLevelRatio returns the level ratio compared to max level for a card.
	// This replaces the simple linear ratio (level / maxLevel) with card-specific
	// exponential curves based on community research.
	GetRelativeLevelRatio(cardName string, level, maxLevel int) float64
}

// Scorer is the core interface for card scoring algorithms.
//
// A Scorer calculates a score for a CardCandidate based on various factors:
// - Card level and rarity
// - Elixir cost
// - Combat statistics (DPS, HP)
// - Evolution level
// - Card synergies
//
// The score is a floating-point value where higher is better. The interpretation
// of the score depends on the specific implementation, but typically scores
// fall in the range of 0.0 to 2.0.
type Scorer interface {
	// Score calculates the score for a card candidate based on the provided config.
	//
	// The config parameter allows scorers to access contextual information like:
	// - Currently selected deck cards (for synergy scoring)
	// - Strategy preferences (role multipliers, elixir targeting)
	// - Level curve data (for non-linear level calculations)
	// - Card stats registry (for combat stats)
	//
	// Implementations should be thread-safe and handle nil/missing data gracefully.
	// If required data is not available, return a neutral score (0.0) rather than
	// causing a panic or error.
	Score(candidate CardCandidate, config ScoringConfig) float64
}

// ScoringConfig provides contextual information for scoring operations.
// It allows scorers to access external data without tight coupling to specific types.
type ScoringConfig struct {
	// LevelCurve provides curve-based level calculation for more accurate scoring.
	// If nil, scorers fall back to linear level ratio calculation.
	LevelCurve LevelCurve

	// CardStats provides combat statistics for stat-based scoring.
	// If nil, combat stat scoring is skipped.
	CardStats *clashroyale.CardStatsRegistry

	// CurrentDeck contains cards already selected for the deck.
	// Used for synergy scoring to identify card pair interactions.
	CurrentDeck []CardCandidate

	// RoleMultipliers defines strategy-specific role bonuses.
	// Maps CardRole to multiplier (e.g., 1.2 for 20% bonus).
	// Nil map means no role multipliers applied.
	RoleMultipliers map[deck.CardRole]float64

	// TargetElixirMin defines the minimum elixir cost for the strategy.
	// Used to penalize cards outside the optimal elixir range.
	TargetElixirMin float64

	// TargetElixirMax defines the maximum elixir cost for the strategy.
	TargetElixirMax float64

	// UnlockedEvolutions tracks which evolutions are available to the player.
	// Maps card name to true if evolution is unlocked.
	UnlockedEvolutions map[string]bool

	// SynergyDatabase provides card pair synergy data.
	// If nil, synergy scoring returns 0.
	SynergyDatabase SynergyDatabase

	// ElixirAdjustmentFunc is an optional function for custom elixir penalty logic.
	// If provided, it overrides the default elixir adjustment calculation.
	// Receives card elixir cost and the ScoringConfig, returns score adjustment.
	ElixirAdjustmentFunc func(elixir int, config ScoringConfig) float64
}

// SynergyDatabase defines the interface for card synergy data access.
// This allows the scoring system to query card pair synergies without
// coupling to the specific deck.SynergyDatabase implementation.
type SynergyDatabase interface {
	// GetSynergy returns the synergy score between two cards.
	// Returns 0.0 if no synergy exists.
	// Synergy scores typically range from 0.0 to 1.0.
	GetSynergy(card1, card2 string) float64

	// AnalyzeDeckSynergy returns comprehensive synergy analysis for a deck.
	// This method is optional for basic synergy scoring but provides
	// detailed analysis when needed.
	AnalyzeDeckSynergy(deckNames []string) *deck.DeckSynergyAnalysis
}

// WeightedScorer combines a Scorer with a weight for composite scoring.
// The weight determines the relative importance of each scorer in the
// composite calculation. Weights should typically sum to 1.0, but this is
// not enforced - the composite scorer normalizes weights automatically.
type WeightedScorer struct {
	// Scorer is the underlying scoring implementation.
	Scorer Scorer

	// Weight determines the relative importance of this scorer.
	// Higher values give more influence to this scorer's output.
	// Weight is applied as: score * weight
	Weight float64
}

// CompositeScorer combines multiple scorers with configurable weights.
// It enables flexible scoring by composing different scoring dimensions.
//
// Example: Combine base scoring (70%), combat stats (25%), and evolution (5%)
//
//	composite := &CompositeScorer{
//	    Scorers: []WeightedScorer{
//	        {Scorer: baseScorer, Weight: 0.70},
//	        {Scorer: combatScorer, Weight: 0.25},
//	        {Scorer: evolutionScorer, Weight: 0.05},
//	    },
//	}
type CompositeScorer struct {
	// Scorers is the list of weighted scorers to combine.
	Scorers []WeightedScorer

	// NormalizationMode determines how weights are applied.
	// If true, weights are normalized to sum to 1.0 before scoring.
	// If false, weights are applied as-is (useful for fixed-weight composites).
	NormalizationMode bool
}

// Score calculates a composite score by combining all scorer outputs.
// The final score is calculated as:
//
//	finalScore = Σ(scorer.Score(candidate, config) * weight)
//
// If NormalizationMode is true, weights are normalized to sum to 1.0 first.
// If a scorer returns a negative score, it is clamped to 0.0.
func (c *CompositeScorer) Score(candidate CardCandidate, config ScoringConfig) float64 {
	if len(c.Scorers) == 0 {
		return 0.0
	}

	totalWeight := 0.0
	for _, ws := range c.Scorers {
		totalWeight += ws.Weight
	}

	// Normalize weights if enabled
	normalizeWeight := c.NormalizationMode && totalWeight > 0

	finalScore := 0.0
	for _, ws := range c.Scorers {
		if ws.Scorer == nil {
			continue
		}

		score := ws.Scorer.Score(candidate, config)
		if score < 0 {
			score = 0 // Clamp negative scores
		}

		weight := ws.Weight
		if normalizeWeight {
			weight = weight / totalWeight
		}

		finalScore += score * weight
	}

	return finalScore
}

// AddScorer adds a new weighted scorer to the composite.
// Returns the composite for method chaining.
func (c *CompositeScorer) AddScorer(scorer Scorer, weight float64) *CompositeScorer {
	c.Scorers = append(c.Scorers, WeightedScorer{
		Scorer: scorer,
		Weight: weight,
	})
	return c
}

// RemoveScorer removes a scorer at the specified index.
// Returns the composite for method chaining.
func (c *CompositeScorer) RemoveScorer(index int) *CompositeScorer {
	if index >= 0 && index < len(c.Scorers) {
		c.Scorers = append(c.Scorers[:index], c.Scorers[index+1:]...)
	}
	return c
}

// SetWeight updates the weight of a scorer at the specified index.
// Returns the composite for method chaining.
func (c *CompositeScorer) SetWeight(index int, weight float64) *CompositeScorer {
	if index >= 0 && index < len(c.Scorers) {
		c.Scorers[index].Weight = weight
	}
	return c
}

// NewCompositeScorer creates a new composite scorer with the provided scorers.
// By default, normalization mode is enabled (weights sum to 1.0).
func NewCompositeScorer(scorers []WeightedScorer) *CompositeScorer {
	return &CompositeScorer{
		Scorers:           scorers,
		NormalizationMode: true,
	}
}

// NewCompositeScorerWithNormalization creates a composite scorer with explicit
// normalization mode control.
func NewCompositeScorerWithNormalization(scorers []WeightedScorer, normalize bool) *CompositeScorer {
	return &CompositeScorer{
		Scorers:           scorers,
		NormalizationMode: normalize,
	}
}

// DefaultScoringConfig returns a ScoringConfig with sensible defaults.
// All optional fields are nil or zero, allowing scorers to use their
// built-in fallback behavior.
func DefaultScoringConfig() ScoringConfig {
	return ScoringConfig{
		LevelCurve:           nil,
		CardStats:            nil,
		CurrentDeck:          nil,
		RoleMultipliers:      nil,
		TargetElixirMin:      0,
		TargetElixirMax:      10, // Effectively no upper limit
		UnlockedEvolutions:   nil,
		SynergyDatabase:      nil,
		ElixirAdjustmentFunc: nil,
	}
}

// ScorerFunc is an adapter that allows ordinary functions to be used as Scorer.
// This is useful for simple scoring logic without defining a full type.
//
// Example:
//
//	scorer := ScorerFunc(func(candidate CardCandidate, config ScoringConfig) float64 {
//	    return float64(candidate.Level) / float64(candidate.MaxLevel)
//	})
type ScorerFunc func(candidate CardCandidate, config ScoringConfig) float64

// Score implements the Scorer interface for ScorerFunc.
func (f ScorerFunc) Score(candidate CardCandidate, config ScoringConfig) float64 {
	return f(candidate, config)
}
