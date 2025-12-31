package scoring

import (
	"math"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// CardRole is an alias for deck.CardRole for convenience
type CardRole = deck.CardRole

const (
	RoleWinCondition CardRole = deck.RoleWinCondition
	RoleBuilding     CardRole = deck.RoleBuilding
	RoleSupport      CardRole = deck.RoleSupport
	RoleSpellBig     CardRole = deck.RoleSpellBig
	RoleSpellSmall   CardRole = deck.RoleSpellSmall
	RoleCycle        CardRole = deck.RoleCycle
)

func TestNewCombatScorer(t *testing.T) {
	scorer := NewCombatScorer(nil)

	if scorer == nil {
		t.Fatal("NewCombatScorer returned nil")
	}

	// Verify default values
	if scorer.dpsWeight != 0.4 {
		t.Errorf("expected dpsWeight 0.4, got %f", scorer.dpsWeight)
	}
	if scorer.hpWeight != 0.4 {
		t.Errorf("expected hpWeight 0.4, got %f", scorer.hpWeight)
	}
	if scorer.roleWeight != 0.2 {
		t.Errorf("expected roleWeight 0.2, got %f", scorer.roleWeight)
	}
	if scorer.dpsNormalizationThreshold != 50.0 {
		t.Errorf("expected dpsNormalizationThreshold 50.0, got %f", scorer.dpsNormalizationThreshold)
	}
	if scorer.hpNormalizationThreshold != 400.0 {
		t.Errorf("expected hpNormalizationThreshold 400.0, got %f", scorer.hpNormalizationThreshold)
	}
}

func TestNewCombatScorerWithConfig(t *testing.T) {
	config := CombatScorerConfig{
		StatsRegistry:             nil,
		DPSWeight:                 0.5,
		HPWeight:                  0.3,
		RoleWeight:                0.2,
		DPSNormalizationThreshold: 60.0,
		HPNormalizationThreshold:  500.0,
	}

	scorer := NewCombatScorerWithConfig(config)

	if scorer.dpsWeight != 0.5 {
		t.Errorf("expected dpsWeight 0.5, got %f", scorer.dpsWeight)
	}
	if scorer.hpWeight != 0.3 {
		t.Errorf("expected hpWeight 0.3, got %f", scorer.hpWeight)
	}
	if scorer.dpsNormalizationThreshold != 60.0 {
		t.Errorf("expected dpsNormalizationThreshold 60.0, got %f", scorer.dpsNormalizationThreshold)
	}
}

func TestCombatScorer_Score_NoStatsRegistry(t *testing.T) {
	scorer := NewCombatScorer(nil)
	config := DefaultScoringConfig()

	role := RoleSupport
	candidate := CardCandidate{
		Name:   "Archers",
		Elixir: 3,
		Role:   &role,
	}

	score := scorer.Score(candidate, config)

	if score != 0.0 {
		t.Errorf("expected score 0.0 with no stats registry, got %f", score)
	}
}

func TestCombatScorer_Score_NoStatsForCard(t *testing.T) {
	// Create a mock registry that returns nil for all cards
	registry := &clashroyale.CardStatsRegistry{}

	scorer := NewCombatScorer(registry)
	config := DefaultScoringConfig()

	role := RoleSupport
	candidate := CardCandidate{
		Name:   "Unknown Card",
		Elixir: 3,
		Role:   &role,
	}

	score := scorer.Score(candidate, config)

	if score != 0.0 {
		t.Errorf("expected score 0.0 when stats not available, got %f", score)
	}
}

func TestCombatScorer_Score_ConfigStatsRegistry(t *testing.T) {
	scorer := NewCombatScorer(nil) // No default registry

	// Mock registry - in real tests would use actual stats
	registry := &clashroyale.CardStatsRegistry{}

	config := DefaultScoringConfig()
	config.CardStats = registry

	role := RoleSupport
	candidate := CardCandidate{
		Name:   "Test Card",
		Elixir: 3,
		Role:   &role,
	}

	// Score should be 0 since mock registry returns nil stats
	score := scorer.Score(candidate, config)

	if score != 0.0 {
		t.Errorf("expected score 0.0 with nil stats from registry, got %f", score)
	}
}

func TestCombatScorer_RoleToString(t *testing.T) {
	scorer := NewCombatScorer(nil)

	tests := []struct {
		role        *CardRole
		expectedStr string
	}{
		{func() *CardRole { r := RoleWinCondition; return &r }(), "wincondition"},
		{func() *CardRole { r := RoleBuilding; return &r }(), "building"},
		{func() *CardRole { r := RoleSupport; return &r }(), "support"},
		{func() *CardRole { r := RoleSpellBig; return &r }(), "spell"},
		{func() *CardRole { r := RoleSpellSmall; return &r }(), "spell"},
		{func() *CardRole { r := RoleCycle; return &r }(), "cycle"},
		{nil, ""},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := scorer.roleToString(tt.role)
			if result != tt.expectedStr {
				t.Errorf("roleToString(%v): expected %s, got %s", tt.role, tt.expectedStr, result)
			}
		})
	}
}

func TestCombatScorer_NormalizeDPS(t *testing.T) {
	scorer := NewCombatScorer(nil) // Default threshold 50.0

	tests := []struct {
		dpsPerElixir     float64
		expectedRangeMin float64
		expectedRangeMax float64
	}{
		{0.0, 0.0, 0.0},
		{25.0, 0.49, 0.51}, // Half of threshold
		{50.0, 0.99, 1.0},  // At threshold
		{100.0, 0.99, 1.0}, // Double threshold (clamped to 1.0)
		{200.0, 0.99, 1.0}, // Very high (clamped to 1.0)
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := scorer.normalizeDPS(tt.dpsPerElixir)
			if result < tt.expectedRangeMin || result > tt.expectedRangeMax {
				t.Errorf("normalizeDPS(%.1f): result %f outside range [%f, %f]",
					tt.dpsPerElixir, result, tt.expectedRangeMin, tt.expectedRangeMax)
			}
		})
	}
}

func TestCombatScorer_NormalizeHP(t *testing.T) {
	scorer := NewCombatScorer(nil) // Default threshold 400.0

	tests := []struct {
		hpPerElixir      float64
		expectedRangeMin float64
		expectedRangeMax float64
	}{
		{0.0, 0.0, 0.0},
		{200.0, 0.49, 0.51}, // Half of threshold
		{400.0, 0.99, 1.0},  // At threshold
		{800.0, 0.99, 1.0},  // Double threshold (clamped to 1.0)
		{1600.0, 0.99, 1.0}, // Very high (clamped to 1.0)
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := scorer.normalizeHP(tt.hpPerElixir)
			if result < tt.expectedRangeMin || result > tt.expectedRangeMax {
				t.Errorf("normalizeHP(%.1f): result %f outside range [%f, %f]",
					tt.hpPerElixir, result, tt.expectedRangeMin, tt.expectedRangeMax)
			}
		})
	}
}

func TestCombatScorer_SetStatsRegistry(t *testing.T) {
	scorer := NewCombatScorer(nil)

	if scorer.statsRegistry != nil {
		t.Errorf("expected nil stats registry, got %v", scorer.statsRegistry)
	}

	registry := &clashroyale.CardStatsRegistry{}
	scorer.SetStatsRegistry(registry)

	if scorer.statsRegistry != registry {
		t.Errorf("stats registry not set correctly")
	}
}

func TestCombatScorer_GetStatsRegistry(t *testing.T) {
	registry := &clashroyale.CardStatsRegistry{}
	scorer := NewCombatScorer(registry)

	retrieved := scorer.GetStatsRegistry()
	if retrieved != registry {
		t.Errorf("GetStatsRegistry returned different registry")
	}
}

func TestCombatScorer_SetWeights(t *testing.T) {
	scorer := NewCombatScorer(nil)

	scorer.SetWeights(0.5, 0.3, 0.2)

	if scorer.dpsWeight != 0.5 {
		t.Errorf("expected dpsWeight 0.5, got %f", scorer.dpsWeight)
	}
	if scorer.hpWeight != 0.3 {
		t.Errorf("expected hpWeight 0.3, got %f", scorer.hpWeight)
	}
	if scorer.roleWeight != 0.2 {
		t.Errorf("expected roleWeight 0.2, got %f", scorer.roleWeight)
	}
}

func TestCombatScorer_GetWeights(t *testing.T) {
	scorer := NewCombatScorer(nil)

	dps, hp, role := scorer.GetWeights()

	if dps != 0.4 {
		t.Errorf("expected dps weight 0.4, got %f", dps)
	}
	if hp != 0.4 {
		t.Errorf("expected hp weight 0.4, got %f", hp)
	}
	if role != 0.2 {
		t.Errorf("expected role weight 0.2, got %f", role)
	}
}

func TestCombatScorer_SetNormalizationThresholds(t *testing.T) {
	scorer := NewCombatScorer(nil)

	scorer.SetNormalizationThresholds(60.0, 500.0)

	if scorer.dpsNormalizationThreshold != 60.0 {
		t.Errorf("expected dpsNormalizationThreshold 60.0, got %f", scorer.dpsNormalizationThreshold)
	}
	if scorer.hpNormalizationThreshold != 500.0 {
		t.Errorf("expected hpNormalizationThreshold 500.0, got %f", scorer.hpNormalizationThreshold)
	}
}

func TestCombatScorer_GetNormalizationThresholds(t *testing.T) {
	scorer := NewCombatScorer(nil)

	dps, hp := scorer.GetNormalizationThresholds()

	if dps != 50.0 {
		t.Errorf("expected dps threshold 50.0, got %f", dps)
	}
	if hp != 400.0 {
		t.Errorf("expected hp threshold 400.0, got %f", hp)
	}
}

func TestCombatScorerConfig_ZeroDefaults(t *testing.T) {
	// Test that zero values in config get replaced with defaults
	config := CombatScorerConfig{
		// All zero/nil values
	}

	scorer := NewCombatScorerWithConfig(config)

	// Should use default values
	if scorer.dpsWeight != 0.4 {
		t.Errorf("expected default dpsWeight 0.4, got %f", scorer.dpsWeight)
	}
	if scorer.hpWeight != 0.4 {
		t.Errorf("expected default hpWeight 0.4, got %f", scorer.hpWeight)
	}
	if scorer.roleWeight != 0.2 {
		t.Errorf("expected default roleWeight 0.2, got %f", scorer.roleWeight)
	}
	if scorer.dpsNormalizationThreshold != 50.0 {
		t.Errorf("expected default dpsNormalizationThreshold 50.0, got %f", scorer.dpsNormalizationThreshold)
	}
	if scorer.hpNormalizationThreshold != 400.0 {
		t.Errorf("expected default hpNormalizationThreshold 400.0, got %f", scorer.hpNormalizationThreshold)
	}
}

func TestCombatScorer_CustomNormalizationThresholds(t *testing.T) {
	config := CombatScorerConfig{
		DPSNormalizationThreshold: 100.0, // Higher threshold
		HPNormalizationThreshold:  800.0, // Higher threshold
	}

	scorer := NewCombatScorerWithConfig(config)

	// Test normalization with custom thresholds
	dpsScore := scorer.normalizeDPS(50.0) // 50/100 = 0.5
	if math.Abs(dpsScore-0.5) > 0.01 {
		t.Errorf("with threshold 100, DPS 50 should normalize to ~0.5, got %f", dpsScore)
	}

	hpScore := scorer.normalizeHP(400.0) // 400/800 = 0.5
	if math.Abs(hpScore-0.5) > 0.01 {
		t.Errorf("with threshold 800, HP 400 should normalize to ~0.5, got %f", hpScore)
	}
}

func TestCombatScorer_WeightSum(t *testing.T) {
	// Verify that custom weights are applied correctly
	config := CombatScorerConfig{
		DPSWeight:  0.6,
		HPWeight:   0.3,
		RoleWeight: 0.1,
	}

	scorer := NewCombatScorerWithConfig(config)

	dps, hp, role := scorer.GetWeights()

	total := dps + hp + role
	if math.Abs(total-1.0) > 0.001 {
		t.Errorf("weights should sum to 1.0, got %f", total)
	}
}

func TestCombatScorer_NilRole(t *testing.T) {
	scorer := NewCombatScorer(nil)

	roleStr := scorer.roleToString(nil)
	if roleStr != "" {
		t.Errorf("expected empty string for nil role, got %s", roleStr)
	}
}

func TestCombatScorer_SpellRoles(t *testing.T) {
	scorer := NewCombatScorer(nil)

	// Both spell types should map to "spell"
	bigSpell := RoleSpellBig
	smallSpell := RoleSpellSmall

	bigResult := scorer.roleToString(&bigSpell)
	smallResult := scorer.roleToString(&smallSpell)

	if bigResult != "spell" {
		t.Errorf("expected 'spell' for big spell, got %s", bigResult)
	}

	if smallResult != "spell" {
		t.Errorf("expected 'spell' for small spell, got %s", smallResult)
	}

	// Both should be the same
	if bigResult != smallResult {
		t.Errorf("both spell types should map to same string: big=%s, small=%s", bigResult, smallResult)
	}
}
