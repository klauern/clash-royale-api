package scoring

import (
	"math"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestNewEvolutionScorer(t *testing.T) {
	unlocked := map[string]bool{"Archers": true, "Knight": true}
	scorer := NewEvolutionScorer(unlocked)

	if scorer == nil {
		t.Fatal("NewEvolutionScorer returned nil")
	}

	// Verify default values
	if scorer.baseBonus != 0.25 {
		t.Errorf("expected baseBonus 0.25, got %f", scorer.baseBonus)
	}
	if scorer.levelScalingExponent != 1.5 {
		t.Errorf("expected levelScalingExponent 1.5, got %f", scorer.levelScalingExponent)
	}
	if scorer.multiEvoMultiplier != 0.2 {
		t.Errorf("expected multiEvoMultiplier 0.2, got %f", scorer.multiEvoMultiplier)
	}
	if !scorer.levelScaled {
		t.Errorf("expected levelScaled true, got %v", scorer.levelScaled)
	}
}

func TestNewEvolutionScorerWithConfig(t *testing.T) {
	unlocked := map[string]bool{"Archers": true}
	config := EvolutionScorerConfig{
		UnlockedEvolutions:   unlocked,
		BaseBonus:            0.3,
		LevelScalingExponent: 2.0,
		MultiEvoMultiplier:   0.25,
		LevelScaled:          false,
	}

	scorer := NewEvolutionScorerWithConfig(config)

	if scorer.baseBonus != 0.3 {
		t.Errorf("expected baseBonus 0.3, got %f", scorer.baseBonus)
	}
	if scorer.levelScalingExponent != 2.0 {
		t.Errorf("expected levelScalingExponent 2.0, got %f", scorer.levelScalingExponent)
	}
	if scorer.levelScaled {
		t.Errorf("expected levelScaled false, got %v", scorer.levelScaled)
	}
}

func TestEvolutionScorer_Score_NoUnlockedEvolutions(t *testing.T) {
	scorer := NewEvolutionScorer(nil)
	config := DefaultScoringConfig()

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:              "Archers",
		Level:             9,
		MaxLevel:          14,
		EvolutionLevel:    2,
		MaxEvolutionLevel: 3,
		Role:              &role,
	}

	score := scorer.Score(candidate, config)

	if score != 0.0 {
		t.Errorf("expected score 0.0 with no unlocked evolutions, got %f", score)
	}
}

func TestEvolutionScorer_Score_EvolutionNotUnlocked(t *testing.T) {
	unlocked := map[string]bool{"Knight": true} // Archers not unlocked
	scorer := NewEvolutionScorer(unlocked)
	config := DefaultScoringConfig()

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:              "Archers",
		Level:             9,
		MaxLevel:          14,
		EvolutionLevel:    2,
		MaxEvolutionLevel: 3,
		Role:              &role,
	}

	score := scorer.Score(candidate, config)

	if score != 0.0 {
		t.Errorf("expected score 0.0 when evolution not unlocked, got %f", score)
	}
}

func TestEvolutionScorer_Score_EvolutionUnlocked(t *testing.T) {
	unlocked := map[string]bool{"Archers": true}
	scorer := NewEvolutionScorer(unlocked)
	config := DefaultScoringConfig()

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:              "Archers",
		Level:             9,
		MaxLevel:          14,
		EvolutionLevel:    2,
		MaxEvolutionLevel: 3,
		Role:              &role,
	}

	score := scorer.Score(candidate, config)

	if score <= 0.0 {
		t.Errorf("expected positive score with unlocked evolution, got %f", score)
	}
}

func TestEvolutionScorer_Score_NoEvolutionCapability(t *testing.T) {
	unlocked := map[string]bool{"Archers": true}
	scorer := NewEvolutionScorer(unlocked)
	config := DefaultScoringConfig()

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:              "Archers",
		Level:             9,
		MaxLevel:          14,
		EvolutionLevel:    0,
		MaxEvolutionLevel: 0, // No evolution capability
		Role:              &role,
	}

	score := scorer.Score(candidate, config)

	if score != 0.0 {
		t.Errorf("expected score 0.0 with no evolution capability, got %f", score)
	}
}

func TestEvolutionScorer_Score_NoEvolutionProgress(t *testing.T) {
	unlocked := map[string]bool{"Archers": true}
	scorer := NewEvolutionScorer(unlocked)
	config := DefaultScoringConfig()

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:              "Archers",
		Level:             9,
		MaxLevel:          14,
		EvolutionLevel:    0, // No evolution progress
		MaxEvolutionLevel: 3,
		Role:              &role,
	}

	score := scorer.Score(candidate, config)

	if score != 0.0 {
		t.Errorf("expected score 0.0 with no evolution progress, got %f", score)
	}
}

func TestEvolutionScorer_FlatBonusCalculation(t *testing.T) {
	unlocked := map[string]bool{"Archers": true}
	scorer := NewEvolutionScorer(unlocked)
	scorer.SetLevelScaled(false) // Use flat calculation
	config := DefaultScoringConfig()

	role := deck.RoleSupport

	tests := []struct {
		name           string
		evolutionLevel int
		maxEvoLevel    int
		expectedMin    float64
		expectedMax    float64
	}{
		{"1/3 evolved", 1, 3, 0.083, 0.084}, // 0.25 * (1/3) ≈ 0.083
		{"2/3 evolved", 2, 3, 0.166, 0.167}, // 0.25 * (2/3) ≈ 0.166
		{"3/3 evolved", 3, 3, 0.249, 0.251}, // 0.25 * (3/3) = 0.25
		{"1/1 evolved", 1, 1, 0.249, 0.251}, // 0.25 * (1/1) = 0.25
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := CardCandidate{
				Name:              "Archers",
				Level:             9,
				MaxLevel:          14,
				EvolutionLevel:    tt.evolutionLevel,
				MaxEvolutionLevel: tt.maxEvoLevel,
				Role:              &role,
			}

			score := scorer.Score(candidate, config)
			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("%s: score %f outside expected range [%f, %f]",
					tt.name, score, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestEvolutionScorer_LevelScaledBonusCalculation(t *testing.T) {
	unlocked := map[string]bool{"Archers": true}
	scorer := NewEvolutionScorer(unlocked)
	scorer.SetLevelScaled(true) // Use level-scaled calculation
	config := DefaultScoringConfig()

	role := deck.RoleSupport

	tests := []struct {
		name           string
		level          int
		maxLevel       int
		evolutionLevel int
		maxEvoLevel    int
		expectedMin    float64
		expectedMax    float64
	}{
		{"Low level card", 5, 14, 2, 3, 0.03, 0.05},
		{"Mid level card", 9, 14, 2, 3, 0.11, 0.13},
		{"High level card", 13, 14, 2, 3, 0.17, 0.23},
		{"Max level card", 14, 14, 2, 3, 0.20, 0.28},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := CardCandidate{
				Name:              "Archers",
				Level:             tt.level,
				MaxLevel:          tt.maxLevel,
				EvolutionLevel:    tt.evolutionLevel,
				MaxEvolutionLevel: tt.maxEvoLevel,
				Role:              &role,
			}

			score := scorer.Score(candidate, config)
			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("%s: score %f outside expected range [%f, %f]",
					tt.name, score, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestEvolutionScorer_ConfigUnlockedEvolutions(t *testing.T) {
	scorer := NewEvolutionScorer(nil) // No default unlocked evolutions

	config := DefaultScoringConfig()
	config.UnlockedEvolutions = map[string]bool{"Archers": true}

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:              "Archers",
		Level:             9,
		MaxLevel:          14,
		EvolutionLevel:    2,
		MaxEvolutionLevel: 3,
		Role:              &role,
	}

	score := scorer.Score(candidate, config)

	if score <= 0.0 {
		t.Errorf("expected positive score with config unlocked evolutions, got %f", score)
	}
}

func TestEvolutionScorer_SetUnlockedEvolutions(t *testing.T) {
	scorer := NewEvolutionScorer(nil)

	if scorer.unlockedEvolutions != nil {
		t.Errorf("expected nil unlocked evolutions, got %v", scorer.unlockedEvolutions)
	}

	unlocked := map[string]bool{"Archers": true}
	scorer.SetUnlockedEvolutions(unlocked)

	if scorer.unlockedEvolutions == nil || !scorer.unlockedEvolutions["Archers"] {
		t.Errorf("unlocked evolutions not set correctly")
	}
}

func TestEvolutionScorer_GetUnlockedEvolutions(t *testing.T) {
	unlocked := map[string]bool{"Archers": true, "Knight": true}
	scorer := NewEvolutionScorer(unlocked)

	retrieved := scorer.GetUnlockedEvolutions()
	if len(retrieved) != len(unlocked) {
		t.Errorf("GetUnlockedEvolutions returned different map")
	}
	if !retrieved["Archers"] || !retrieved["Knight"] {
		t.Errorf("GetUnlockedEvolutions missing expected entries")
	}
}

func TestEvolutionScorer_SetLevelScaled(t *testing.T) {
	scorer := NewEvolutionScorer(nil)

	scorer.SetLevelScaled(false)
	if scorer.levelScaled {
		t.Errorf("expected levelScaled false, got %v", scorer.levelScaled)
	}

	scorer.SetLevelScaled(true)
	if !scorer.levelScaled {
		t.Errorf("expected levelScaled true, got %v", scorer.levelScaled)
	}
}

func TestEvolutionScorer_IsLevelScaled(t *testing.T) {
	scorer := NewEvolutionScorer(nil)

	if !scorer.IsLevelScaled() {
		t.Errorf("expected levelScaled true by default")
	}

	scorer.SetLevelScaled(false)
	if scorer.IsLevelScaled() {
		t.Errorf("expected levelScaled false after setting")
	}
}

func TestEvolutionScorer_SetBaseBonus(t *testing.T) {
	scorer := NewEvolutionScorer(nil)

	scorer.SetBaseBonus(0.5)

	if scorer.baseBonus != 0.5 {
		t.Errorf("expected baseBonus 0.5, got %f", scorer.baseBonus)
	}
}

func TestEvolutionScorer_GetBaseBonus(t *testing.T) {
	scorer := NewEvolutionScorer(nil)

	bonus := scorer.GetBaseBonus()
	if bonus != 0.25 {
		t.Errorf("expected baseBonus 0.25, got %f", bonus)
	}
}

func TestEvolutionScorer_SetLevelScalingExponent(t *testing.T) {
	scorer := NewEvolutionScorer(nil)

	scorer.SetLevelScalingExponent(2.0)

	if scorer.levelScalingExponent != 2.0 {
		t.Errorf("expected levelScalingExponent 2.0, got %f", scorer.levelScalingExponent)
	}
}

func TestEvolutionScorer_GetLevelScalingExponent(t *testing.T) {
	scorer := NewEvolutionScorer(nil)

	exponent := scorer.GetLevelScalingExponent()
	if exponent != 1.5 {
		t.Errorf("expected levelScalingExponent 1.5, got %f", exponent)
	}
}

func TestEvolutionScorer_MultiEvolutionBonus(t *testing.T) {
	unlocked := map[string]bool{"Knight": true}
	scorer := NewEvolutionScorer(unlocked)
	config := DefaultScoringConfig()

	role := deck.RoleSupport

	// Compare cards at same evolution progress level (fully evolved)
	// Knight with 3 evolution levels (at evo 3/3) vs single evolution (at evo 1/1)
	candidate1Evo := CardCandidate{
		Name:              "Knight",
		Level:             14,
		MaxLevel:          14,
		EvolutionLevel:    1,
		MaxEvolutionLevel: 1,
		Role:              &role,
	}

	candidate3Evo := CardCandidate{
		Name:              "Knight",
		Level:             14,
		MaxLevel:          14,
		EvolutionLevel:    3, // Fully evolved (3/3)
		MaxEvolutionLevel: 3,
		Role:              &role,
	}

	score1Evo := scorer.Score(candidate1Evo, config)
	score3Evo := scorer.Score(candidate3Evo, config)

	// 3-evo card (fully evolved) should have higher score due to evoMultiplier
	if score3Evo <= score1Evo {
		t.Errorf("3-evo card should score higher: 1evo=%f, 3evo=%f", score1Evo, score3Evo)
	}
}

func TestEvolutionScorerConfig_ZeroDefaults(t *testing.T) {
	// Test that zero values in config get replaced with defaults
	config := EvolutionScorerConfig{
		// All zero/nil values
	}

	scorer := NewEvolutionScorerWithConfig(config)

	// Should use default values
	if scorer.baseBonus != 0.25 {
		t.Errorf("expected default baseBonus 0.25, got %f", scorer.baseBonus)
	}
	if scorer.levelScalingExponent != 1.5 {
		t.Errorf("expected default levelScalingExponent 1.5, got %f", scorer.levelScalingExponent)
	}
	if scorer.multiEvoMultiplier != 0.2 {
		t.Errorf("expected default multiEvoMultiplier 0.2, got %f", scorer.multiEvoMultiplier)
	}
}

func TestEvolutionScorer_ZeroMaxLevel(t *testing.T) {
	unlocked := map[string]bool{"Archers": true}
	scorer := NewEvolutionScorer(unlocked)
	config := DefaultScoringConfig()

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:              "Archers",
		Level:             9,
		MaxLevel:          0, // Invalid
		EvolutionLevel:    2,
		MaxEvolutionLevel: 3,
		Role:              &role,
	}

	score := scorer.Score(candidate, config)

	// Should return 0 for invalid max level in level-scaled mode
	if score != 0.0 {
		t.Errorf("expected score 0.0 with maxLevel=0, got %f", score)
	}
}

func TestCalculateEvolutionBonusForCard(t *testing.T) {
	unlocked := map[string]bool{"Archers": true}

	bonus := CalculateEvolutionBonusForCard(
		"Archers",
		9, 14, // level, maxLevel
		2, 3, // evolutionLevel, maxEvolutionLevel
		unlocked,
	)

	if bonus <= 0 {
		t.Errorf("expected positive bonus, got %f", bonus)
	}

	// Should use flat calculation: 0.25 * (2/3) ≈ 0.166
	expected := 0.25 * (2.0 / 3.0)
	if math.Abs(bonus-expected) > 0.01 {
		t.Errorf("expected bonus ~%f, got %f", expected, bonus)
	}
}

func TestEvolutionScorer_OverMaxEvolution(t *testing.T) {
	unlocked := map[string]bool{"Archers": true}
	scorer := NewEvolutionScorer(unlocked)
	scorer.SetLevelScaled(false)
	config := DefaultScoringConfig()

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:              "Archers",
		Level:             14,
		MaxLevel:          14,
		EvolutionLevel:    5, // More than max
		MaxEvolutionLevel: 3,
		Role:              &role,
	}

	score := scorer.Score(candidate, config)

	// Should clamp to max bonus (0.25)
	if score > 0.251 {
		t.Errorf("score should be clamped to 0.25, got %f", score)
	}
}
