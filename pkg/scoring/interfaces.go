// Package scoring provides card scoring algorithms for base card stats,
// combat effectiveness, evolution bonuses, and card synergies.
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
