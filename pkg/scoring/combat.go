// Package scoring provides implementations of the Scorer interface for
// various card scoring algorithms.
package scoring

import (
	"math"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// CardRole is an alias for deck.CardRole for convenience
type CardRole = deck.CardRole

// CombatScorer implements the Scorer interface using combat statistics.
// It evaluates card effectiveness based on DPS per elixir, HP per elixir,
// and role-specific effectiveness. This scorer provides stat-based evaluation
// that complements traditional level-based scoring.
//
// # Scoring Formula
//
//	combatScore = (dpsNormalized × dpsWeight) +
//	              (hpNormalized × hpWeight) +
//	              (roleEffectiveness × roleWeight)
//
// Where:
//   - dpsNormalized: DPS efficiency normalized to 0-1 range
//   - hpNormalized: HP efficiency normalized to 0-1 range
//   - roleEffectiveness: Role-specific effectiveness from stats registry
//
// Default weights:
//   - dpsWeight: 0.4 (40% of score from DPS)
//   - hpWeight: 0.4 (40% of score from HP)
//   - roleWeight: 0.2 (20% of score from role effectiveness)
//
// Normalization thresholds:
//   - Excellent DPS: ~50 DPS/elixir
//   - Excellent HP: ~400 HP/elixir
type CombatScorer struct {
	// statsRegistry provides combat statistics for cards.
	// If nil, scorer returns 0 for all cards.
	statsRegistry *clashroyale.CardStatsRegistry

	// dpsWeight determines how much DPS efficiency impacts the score.
	// Default 0.4 (40% of combat score).
	dpsWeight float64

	// hpWeight determines how much HP efficiency impacts the score.
	// Default 0.4 (40% of combat score).
	hpWeight float64

	// roleWeight determines how much role effectiveness impacts the score.
	// Default 0.2 (20% of combat score).
	roleWeight float64

	// dpsNormalizationThreshold is the DPS/elixir value considered excellent.
	// Used to normalize DPS scores to 0-1 range.
	// Default 50.0.
	dpsNormalizationThreshold float64

	// hpNormalizationThreshold is the HP/elixir value considered excellent.
	// Used to normalize HP scores to 0-1 range.
	// Default 400.0.
	hpNormalizationThreshold float64
}

// CombatScorerConfig configures a CombatScorer with custom weights and parameters.
type CombatScorerConfig struct {
	// StatsRegistry provides combat statistics for cards.
	// If nil, combat scoring returns 0.
	StatsRegistry *clashroyale.CardStatsRegistry

	// DPSWeight determines DPS importance (default 0.4).
	DPSWeight float64

	// HPWeight determines HP importance (default 0.4).
	HPWeight float64

	// RoleWeight determines role effectiveness importance (default 0.2).
	RoleWeight float64

	// DPSNormalizationThreshold sets the excellent DPS/elixir benchmark (default 50.0).
	DPSNormalizationThreshold float64

	// HPNormalizationThreshold sets the excellent HP/elixir benchmark (default 400.0).
	HPNormalizationThreshold float64
}

// NewCombatScorer creates a new CombatScorer with default parameters.
//
// Example:
//
//	scorer := NewCombatScorer(statsRegistry)
//	score := scorer.Score(candidate, config)
func NewCombatScorer(statsRegistry *clashroyale.CardStatsRegistry) *CombatScorer {
	return &CombatScorer{
		statsRegistry:             statsRegistry,
		dpsWeight:                 0.4,
		hpWeight:                  0.4,
		roleWeight:                0.2,
		dpsNormalizationThreshold: 50.0,
		hpNormalizationThreshold:  400.0,
	}
}

// NewCombatScorerWithConfig creates a new CombatScorer with custom configuration.
//
// Example:
//
//	config := CombatScorerConfig{
//	    StatsRegistry: statsRegistry,
//	    DPSWeight: 0.5,  // Prioritize DPS more
//	    HPWeight: 0.3,
//	    RoleWeight: 0.2,
//	}
//	scorer := NewCombatScorerWithConfig(config)
func NewCombatScorerWithConfig(config CombatScorerConfig) *CombatScorer {
	dpsWeight := config.DPSWeight
	if dpsWeight == 0 {
		dpsWeight = 0.4
	}

	hpWeight := config.HPWeight
	if hpWeight == 0 {
		hpWeight = 0.4
	}

	roleWeight := config.RoleWeight
	if roleWeight == 0 {
		roleWeight = 0.2
	}

	dpsThreshold := config.DPSNormalizationThreshold
	if dpsThreshold == 0 {
		dpsThreshold = 50.0
	}

	hpThreshold := config.HPNormalizationThreshold
	if hpThreshold == 0 {
		hpThreshold = 400.0
	}

	return &CombatScorer{
		statsRegistry:             config.StatsRegistry,
		dpsWeight:                 dpsWeight,
		hpWeight:                  hpWeight,
		roleWeight:                roleWeight,
		dpsNormalizationThreshold: dpsThreshold,
		hpNormalizationThreshold:  hpThreshold,
	}
}

// Score calculates the combat score for a card candidate.
//
// The score combines three factors:
//   - DPS efficiency (damage per second per elixir)
//   - HP efficiency (hit points per elixir)
//   - Role-specific effectiveness (how good the card is at its role)
//
// Returns 0 if combat stats are not available.
// Returns a score in the range 0.0 to 1.0, higher is better.
func (s *CombatScorer) Score(candidate CardCandidate, config ScoringConfig) float64 {
	// Use config's stats registry if provided, otherwise use scorer's default
	statsRegistry := config.CardStats
	if statsRegistry == nil {
		statsRegistry = s.statsRegistry
	}

	if statsRegistry == nil {
		return 0.0 // No stats available
	}

	// Get stats for the card
	stats := statsRegistry.GetStats(candidate.Name)
	if stats == nil {
		return 0.0 // No stats for this card
	}

	// Get role string for effectiveness calculation
	roleStr := s.roleToString(candidate.Role)

	// Calculate DPS efficiency
	dpsEfficiency := stats.DPSPerElixir(candidate.Elixir)

	// Calculate HP efficiency
	hpEfficiency := stats.HPPerElixir(candidate.Elixir)

	// Get role-specific effectiveness
	roleEffectiveness := stats.RoleSpecificEffectiveness(roleStr)

	// Normalize to 0-1 range
	dpsNormalized := s.normalizeDPS(dpsEfficiency)
	hpNormalized := s.normalizeHP(hpEfficiency)

	// Combine with weights
	combatScore := (dpsNormalized * s.dpsWeight) +
		(hpNormalized * s.hpWeight) +
		(roleEffectiveness * s.roleWeight)

	// Clamp to 0-1 range
	return math.Max(0, math.Min(1, combatScore))
}

// roleToString converts CardRole to string for stats lookup.
func (s *CombatScorer) roleToString(role *CardRole) string {
	if role == nil {
		return ""
	}

	// CardRole is an alias to deck.CardRole
	switch *role {
	case "win_conditions":
		return "wincondition"
	case "buildings":
		return "building"
	case "support":
		return "support"
	case "spells_big":
		return "spell"
	case "spells_small":
		return "spell" // Treat both spell types as "spell"
	case "cycle":
		return "cycle"
	default:
		return ""
	}
}

// normalizeDPS normalizes DPS/elixir to 0-1 range.
// Uses the normalization threshold as the "excellent" benchmark.
func (s *CombatScorer) normalizeDPS(dpsPerElixir float64) float64 {
	normalized := dpsPerElixir / s.dpsNormalizationThreshold
	return math.Min(normalized, 1.0)
}

// normalizeHP normalizes HP/elixir to 0-1 range.
// Uses the normalization threshold as the "excellent" benchmark.
func (s *CombatScorer) normalizeHP(hpPerElixir float64) float64 {
	normalized := hpPerElixir / s.hpNormalizationThreshold
	return math.Min(normalized, 1.0)
}

// SetStatsRegistry updates the stats registry used for scoring.
func (s *CombatScorer) SetStatsRegistry(registry *clashroyale.CardStatsRegistry) {
	s.statsRegistry = registry
}

// GetStatsRegistry returns the current stats registry.
func (s *CombatScorer) GetStatsRegistry() *clashroyale.CardStatsRegistry {
	return s.statsRegistry
}

// SetWeights updates all scoring weights at once.
func (s *CombatScorer) SetWeights(dps, hp, role float64) {
	s.dpsWeight = dps
	s.hpWeight = hp
	s.roleWeight = role
}

// GetWeights returns the current scoring weights.
func (s *CombatScorer) GetWeights() (dps, hp, role float64) {
	return s.dpsWeight, s.hpWeight, s.roleWeight
}

// SetNormalizationThresholds updates the normalization thresholds.
func (s *CombatScorer) SetNormalizationThresholds(dps, hp float64) {
	s.dpsNormalizationThreshold = dps
	s.hpNormalizationThreshold = hp
}

// GetNormalizationThresholds returns the current normalization thresholds.
func (s *CombatScorer) GetNormalizationThresholds() (dps, hp float64) {
	return s.dpsNormalizationThreshold, s.hpNormalizationThreshold
}
