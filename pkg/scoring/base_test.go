package scoring

import (
	"math"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// mockLevelCurve is a test double for LevelCurve interface
type mockLevelCurve struct {
	getRelativeLevelRatioFunc func(cardName string, level, maxLevel int) float64
}

func (m *mockLevelCurve) GetRelativeLevelRatio(cardName string, level, maxLevel int) float64 {
	if m.getRelativeLevelRatioFunc != nil {
		return m.getRelativeLevelRatioFunc(cardName, level, maxLevel)
	}
	return float64(level) / float64(maxLevel)
}

func TestNewBaseScorer(t *testing.T) {
	scorer := NewBaseScorer(nil)

	if scorer == nil {
		t.Fatal("NewBaseScorer returned nil")
	}

	// Verify default values
	if scorer.levelWeightFactor != 1.2 {
		t.Errorf("expected levelWeightFactor 1.2, got %f", scorer.levelWeightFactor)
	}
	if scorer.elixirOptimal != 3.0 {
		t.Errorf("expected elixirOptimal 3.0, got %f", scorer.elixirOptimal)
	}
	if scorer.elixirWeightFactor != 0.15 {
		t.Errorf("expected elixirWeightFactor 0.15, got %f", scorer.elixirWeightFactor)
	}
	if scorer.roleBonusValue != 0.05 {
		t.Errorf("expected roleBonusValue 0.05, got %f", scorer.roleBonusValue)
	}
	if scorer.evolutionBonusWeight != 0.15 {
		t.Errorf("expected evolutionBonusWeight 0.15, got %f", scorer.evolutionBonusWeight)
	}
}

func TestNewBaseScorerWithConfig(t *testing.T) {
	config := BaseScorerConfig{
		LevelCurve:           &mockLevelCurve{},
		LevelWeightFactor:    1.5,
		ElixirOptimal:        4.0,
		ElixirWeightFactor:   0.2,
		RoleBonusValue:       0.1,
		EvolutionBonusWeight: 0.2,
	}

	scorer := NewBaseScorerWithConfig(config)

	if scorer.levelWeightFactor != 1.5 {
		t.Errorf("expected levelWeightFactor 1.5, got %f", scorer.levelWeightFactor)
	}
	if scorer.elixirOptimal != 4.0 {
		t.Errorf("expected elixirOptimal 4.0, got %f", scorer.elixirOptimal)
	}
}

func TestDefaultRarityWeights(t *testing.T) {
	weights := DefaultRarityWeights()

	expectedWeights := map[string]float64{
		"Common":    1.0,
		"Rare":      1.05,
		"Epic":      1.1,
		"Legendary": 1.15,
		"Champion":  1.2,
	}

	for rarity, expected := range expectedWeights {
		if actual, exists := weights[rarity]; !exists {
			t.Errorf("missing rarity: %s", rarity)
		} else if actual != expected {
			t.Errorf("rarity %s: expected %f, got %f", rarity, expected, actual)
		}
	}
}

func TestBaseScorer_Score_BasicCard(t *testing.T) {
	scorer := NewBaseScorer(nil)
	config := DefaultScoringConfig()

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:              "Archers",
		Level:             9,
		MaxLevel:          14,
		Rarity:            "Common",
		Elixir:            3,
		Role:              &role,
		EvolutionLevel:    0,
		MaxEvolutionLevel: 0,
	}

	score := scorer.Score(candidate, config)

	// Score should be positive and reasonable
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}

	// With level 9/14 (~0.64), rarity 1.0, elixir 3 (optimal), role bonus 0.05
	// Expected: (0.64 * 1.2 * 1.0) + (1.0 * 0.15) + 0.05 = ~0.97
	if score < 0.8 || score > 1.2 {
		t.Errorf("score %f outside expected range [0.8, 1.2]", score)
	}
}

func TestBaseScorer_Score_MaxLevelCard(t *testing.T) {
	scorer := NewBaseScorer(nil)
	config := DefaultScoringConfig()

	role := deck.RoleWinCondition
	candidate := CardCandidate{
		Name:              "Mega Knight",
		Level:             14,
		MaxLevel:          14,
		Rarity:            "Legendary",
		Elixir:            7,
		Role:              &role,
		EvolutionLevel:    0,
		MaxEvolutionLevel: 0,
	}

	score := scorer.Score(candidate, config)

	// Max level legendary should score very high
	if score < 1.0 {
		t.Errorf("max level legendary should score high, got %f", score)
	}
}

func TestBaseScorer_Score_WithEvolution(t *testing.T) {
	scorer := NewBaseScorer(nil)
	config := DefaultScoringConfig()

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:              "Archers",
		Level:             9,
		MaxLevel:          14,
		Rarity:            "Common",
		Elixir:            3,
		Role:              &role,
		EvolutionLevel:    2,
		MaxEvolutionLevel: 3,
	}

	scoreWithoutEvo := scorer.Score(CardCandidate{
		Name:              "Archers",
		Level:             9,
		MaxLevel:          14,
		Rarity:            "Common",
		Elixir:            3,
		Role:              &role,
		EvolutionLevel:    0,
		MaxEvolutionLevel: 3,
	}, config)

	scoreWithEvo := scorer.Score(candidate, config)

	// Score with evolution should be higher
	if scoreWithEvo <= scoreWithoutEvo {
		t.Errorf("evolved card should score higher: without=%f, with=%f",
			scoreWithoutEvo, scoreWithEvo)
	}

	// Evolution bonus should be approximately 0.15 * (2/3) = 0.10
	evolutionDelta := scoreWithEvo - scoreWithoutEvo
	expectedDelta := 0.15 * (2.0 / 3.0)

	if math.Abs(evolutionDelta-expectedDelta) > 0.01 {
		t.Errorf("evolution delta: expected ~%f, got %f", expectedDelta, evolutionDelta)
	}
}

func TestBaseScorer_Score_LevelCurve(t *testing.T) {
	// Mock curve that gives higher ratio for lower levels
	curve := &mockLevelCurve{
		getRelativeLevelRatioFunc: func(cardName string, level, maxLevel int) float64 {
			// Curve gives 0.8 for level 9/14 instead of linear 0.64
			if level == 9 && maxLevel == 14 {
				return 0.8
			}
			return float64(level) / float64(maxLevel)
		},
	}

	scorer := NewBaseScorer(curve)
	config := DefaultScoringConfig()

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:     "Archers",
		Level:    9,
		MaxLevel: 14,
		Rarity:   "Common",
		Elixir:   3,
		Role:     &role,
	}

	scoreWithCurve := scorer.Score(candidate, config)

	// Create scorer without curve for comparison
	scorerNoCurve := NewBaseScorer(nil)
	scoreNoCurve := scorerNoCurve.Score(candidate, config)

	// Score with curve should be higher due to boosted level ratio
	if scoreWithCurve <= scoreNoCurve {
		t.Errorf("curve-based score should be higher: noCurve=%f, curve=%f",
			scoreNoCurve, scoreWithCurve)
	}
}

func TestBaseScorer_Score_NoRole(t *testing.T) {
	scorer := NewBaseScorer(nil)
	config := DefaultScoringConfig()

	candidateWithRole := CardCandidate{
		Name:     "Archers",
		Level:    9,
		MaxLevel: 14,
		Rarity:   "Common",
		Elixir:   3,
		Role:     func() *deck.CardRole { r := deck.RoleSupport; return &r }(),
	}

	candidateWithoutRole := CardCandidate{
		Name:     "Some Card",
		Level:    9,
		MaxLevel: 14,
		Rarity:   "Common",
		Elixir:   3,
		Role:     nil,
	}

	scoreWithRole := scorer.Score(candidateWithRole, config)
	scoreWithoutRole := scorer.Score(candidateWithoutRole, config)

	// Score with role should be higher by roleBonusValue
	delta := scoreWithRole - scoreWithoutRole
	expectedDelta := scorer.roleBonusValue

	if math.Abs(delta-expectedDelta) > 0.001 {
		t.Errorf("role bonus delta: expected %f, got %f", expectedDelta, delta)
	}
}

func TestBaseScorer_Score_ElixirEfficiency(t *testing.T) {
	scorer := NewBaseScorer(nil)
	config := DefaultScoringConfig()

	// Create candidates with different elixir costs
	candidates := []struct {
		name   string
		elixir int
	}{
		{"Low Elixir", 1},
		{"Optimal Elixir", 3},
		{"High Elixir", 7},
	}

	scores := make(map[string]float64)
	for _, c := range candidates {
		role := deck.RoleCycle
		candidate := CardCandidate{
			Name:              c.name,
			Level:             9,
			MaxLevel:          14,
			Rarity:            "Common",
			Elixir:            c.elixir,
			Role:              &role,
			EvolutionLevel:    0,
			MaxEvolutionLevel: 0,
		}
		scores[c.name] = scorer.Score(candidate, config)
	}

	// Optimal elixir (3) should score highest
	if scores["Optimal Elixir"] <= scores["Low Elixir"] {
		t.Errorf("optimal elixir should score higher than low: optimal=%f, low=%f",
			scores["Optimal Elixir"], scores["Low Elixir"])
	}

	if scores["Optimal Elixir"] <= scores["High Elixir"] {
		t.Errorf("optimal elixir should score higher than high: optimal=%f, high=%f",
			scores["Optimal Elixir"], scores["High Elixir"])
	}
}

func TestBaseScorer_Score_RarityBoost(t *testing.T) {
	scorer := NewBaseScorer(nil)
	config := DefaultScoringConfig()

	rarities := []string{"Common", "Rare", "Epic", "Legendary", "Champion"}
	scores := make(map[string]float64)

	for _, rarity := range rarities {
		role := deck.RoleSupport
		candidate := CardCandidate{
			Name:     "Card",
			Level:    9,
			MaxLevel: 14,
			Rarity:   rarity,
			Elixir:   3,
			Role:     &role,
		}
		scores[rarity] = scorer.Score(candidate, config)
	}

	// Each increasing rarity should have equal or higher score
	if scores["Common"] > scores["Rare"] {
		t.Errorf("Rare should score >= Common: common=%f, rare=%f",
			scores["Common"], scores["Rare"])
	}

	if scores["Rare"] > scores["Epic"] {
		t.Errorf("Epic should score >= Rare: rare=%f, epic=%f",
			scores["Rare"], scores["Epic"])
	}

	if scores["Epic"] > scores["Legendary"] {
		t.Errorf("Legendary should score >= Epic: epic=%f, legendary=%f",
			scores["Epic"], scores["Legendary"])
	}

	if scores["Legendary"] > scores["Champion"] {
		t.Errorf("Champion should score >= Legendary: legendary=%f, champion=%f",
			scores["Legendary"], scores["Champion"])
	}
}

func TestBaseScorer_Score_UnknownRarity(t *testing.T) {
	scorer := NewBaseScorer(nil)
	config := DefaultScoringConfig()

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:     "Unknown Card",
		Level:    9,
		MaxLevel: 14,
		Rarity:   "UnknownRarity",
		Elixir:   3,
		Role:     &role,
	}

	// Should not panic and should treat unknown rarity as Common (1.0)
	score := scorer.Score(candidate, config)

	if score <= 0 {
		t.Errorf("unknown rarity should still yield positive score, got %f", score)
	}
}

func TestBaseScorer_Score_ZeroMaxLevel(t *testing.T) {
	scorer := NewBaseScorer(nil)
	config := DefaultScoringConfig()

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:     "Invalid Card",
		Level:    9,
		MaxLevel: 0, // Invalid
		Rarity:   "Common",
		Elixir:   3,
		Role:     &role,
	}

	// Should not panic and should return 0 or minimal score
	score := scorer.Score(candidate, config)

	if score < 0 {
		t.Errorf("score should not be negative, got %f", score)
	}

	// With max level 0, level ratio is 0, so score should be low
	// (only elixir weight + role bonus)
	if score > 0.5 {
		t.Errorf("score with maxLevel=0 should be low, got %f", score)
	}
}

func TestBaseScorer_GetRarityBoost(t *testing.T) {
	scorer := NewBaseScorer(nil)

	tests := []struct {
		rarity        string
		expectedBoost float64
	}{
		{"Common", 1.0},
		{"Rare", 1.05},
		{"Epic", 1.1},
		{"Legendary", 1.15},
		{"Champion", 1.2},
		{"Unknown", 1.0}, // Should default to Common
	}

	for _, tt := range tests {
		t.Run(tt.rarity, func(t *testing.T) {
			boost := scorer.getRarityBoost(tt.rarity)
			if boost != tt.expectedBoost {
				t.Errorf("rarity %s: expected boost %f, got %f",
					tt.rarity, tt.expectedBoost, boost)
			}
		})
	}
}

func TestBaseScorer_CalculateElixirWeight(t *testing.T) {
	scorer := NewBaseScorer(nil)

	tests := []struct {
		elixir           int
		expectedRangeMin float64
		expectedRangeMax float64
	}{
		{1, 0.75, 0.8}, // Far from optimal: 1 - (2/9) ≈ 0.78
		{2, 0.85, 0.9}, // Below optimal: 1 - (1/9) ≈ 0.89
		{3, 1.0, 1.0},  // Optimal: 1 - (0/9) = 1.0
		{4, 0.85, 0.9}, // Above optimal: 1 - (1/9) ≈ 0.89
		{5, 0.75, 0.8}, // High: 1 - (2/9) ≈ 0.78
		{7, 0.5, 0.6},  // Very high: 1 - (4/9) ≈ 0.56
	}

	for _, tt := range tests {
		t.Run(elixirStr(tt.elixir), func(t *testing.T) {
			weight := scorer.calculateElixirWeight(tt.elixir)
			if weight < tt.expectedRangeMin || weight > tt.expectedRangeMax {
				t.Errorf("elixir %d: weight %f outside range [%f, %f]",
					tt.elixir, weight, tt.expectedRangeMin, tt.expectedRangeMax)
			}
		})
	}
}

func elixirStr(e int) string {
	switch e {
	case 1:
		return "Low"
	case 2:
		return "BelowOptimal"
	case 3:
		return "Optimal"
	case 4:
		return "AboveOptimal"
	case 5:
		return "High"
	case 7:
		return "VeryHigh"
	default:
		return "Unknown"
	}
}

func TestBaseScorer_CalculateEvolutionBonus(t *testing.T) {
	scorer := NewBaseScorer(nil)

	tests := []struct {
		name              string
		evolutionLevel    int
		maxEvolutionLevel int
		expectedBonus     float64
	}{
		{"No Evolution", 0, 3, 0.0},
		{"No Evolution Capability", 1, 0, 0.0},
		{"Partial Evolution 1/3", 1, 3, 0.15 * (1.0 / 3.0)},
		{"Partial Evolution 2/3", 2, 3, 0.15 * (2.0 / 3.0)},
		{"Full Evolution", 3, 3, 0.15},
		{"Over Max", 5, 3, 0.15}, // Should clamp to 1.0 ratio
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonus := scorer.calculateEvolutionBonus(tt.evolutionLevel, tt.maxEvolutionLevel)
			if math.Abs(bonus-tt.expectedBonus) > 0.001 {
				t.Errorf("%s: expected bonus %f, got %f",
					tt.name, tt.expectedBonus, bonus)
			}
		})
	}
}

func TestBaseScorer_CalculateLevelRatio_Linear(t *testing.T) {
	scorer := NewBaseScorer(nil) // No level curve provided

	tests := []struct {
		level    int
		maxLevel int
		expected float64
	}{
		{0, 14, 0.0},
		{7, 14, 0.5},
		{14, 14, 1.0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			ratio := scorer.calculateLevelRatio("Card", tt.level, tt.maxLevel, nil)
			if ratio != tt.expected {
				t.Errorf("level %d/%d: expected ratio %f, got %f",
					tt.level, tt.maxLevel, tt.expected, ratio)
			}
		})
	}
}

func TestBaseScorer_CalculateLevelRatio_WithCurve(t *testing.T) {
	curve := &mockLevelCurve{
		getRelativeLevelRatioFunc: func(cardName string, level, maxLevel int) float64 {
			// Mock curve returns double the linear ratio
			linear := float64(level) / float64(maxLevel)
			return math.Min(linear*2, 1.0)
		},
	}

	scorer := NewBaseScorer(curve)

	// Test with curve
	ratioWithCurve := scorer.calculateLevelRatio("TestCard", 7, 14, curve)
	ratioNoCurve := scorer.calculateLevelRatio("TestCard", 7, 14, nil)

	// Curve should give higher ratio
	if ratioWithCurve <= ratioNoCurve {
		t.Errorf("curve should give higher ratio: curve=%f, noCurve=%f",
			ratioWithCurve, ratioNoCurve)
	}
}

func TestBaseScorer_SetLevelCurve(t *testing.T) {
	scorer := NewBaseScorer(nil)

	if scorer.levelCurve != nil {
		t.Errorf("expected nil level curve, got %v", scorer.levelCurve)
	}

	curve := &mockLevelCurve{}
	scorer.SetLevelCurve(curve)

	if scorer.levelCurve != curve {
		t.Errorf("level curve not set correctly")
	}
}

func TestBaseScorer_GetLevelCurve(t *testing.T) {
	curve := &mockLevelCurve{}
	scorer := NewBaseScorer(curve)

	retrieved := scorer.GetLevelCurve()
	if retrieved != curve {
		t.Errorf("GetLevelCurve returned different curve")
	}
}

func TestBaseScorer_SetRarityWeights(t *testing.T) {
	scorer := NewBaseScorer(nil)

	customWeights := map[string]float64{
		"Common": 1.0,
		"Rare":   2.0, // Custom higher weight
	}

	scorer.SetRarityWeights(customWeights)

	boost := scorer.getRarityBoost("Rare")
	if boost != 2.0 {
		t.Errorf("expected custom weight 2.0, got %f", boost)
	}
}

func TestBaseScorer_GetRarityWeights(t *testing.T) {
	scorer := NewBaseScorer(nil)

	weights := scorer.GetRarityWeights()
	if weights == nil {
		t.Errorf("expected non-nil weights")
	}

	if len(weights) != 5 {
		t.Errorf("expected 5 rarity weights, got %d", len(weights))
	}
}

func TestBaseScorer_Score_ConfigLevelCurve(t *testing.T) {
	scorer := NewBaseScorer(nil)

	curve := &mockLevelCurve{
		getRelativeLevelRatioFunc: func(cardName string, level, maxLevel int) float64 {
			return 0.9 // Fixed high ratio
		},
	}

	config := DefaultScoringConfig()
	config.LevelCurve = curve

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name:     "TestCard",
		Level:    5,
		MaxLevel: 14,
		Rarity:   "Common",
		Elixir:   3,
		Role:     &role,
	}

	score := scorer.Score(candidate, config)

	// Score should use config's level curve (0.9 ratio) instead of linear (0.35)
	// Expected: (0.9 * 1.2 * 1.0) + (1.0 * 0.15) + 0.05 = ~1.28
	if score < 1.2 {
		t.Errorf("score with config level curve should be high, got %f", score)
	}
}

func TestBaseScorerConfig_ZeroDefaults(t *testing.T) {
	// Test that zero values in config get replaced with defaults
	config := BaseScorerConfig{
		// All zero/nil values
	}

	scorer := NewBaseScorerWithConfig(config)

	// Should use default values
	if scorer.levelWeightFactor != 1.2 {
		t.Errorf("expected default levelWeightFactor 1.2, got %f", scorer.levelWeightFactor)
	}
	if scorer.elixirOptimal != 3.0 {
		t.Errorf("expected default elixirOptimal 3.0, got %f", scorer.elixirOptimal)
	}
	if scorer.elixirWeightFactor != 0.15 {
		t.Errorf("expected default elixirWeightFactor 0.15, got %f", scorer.elixirWeightFactor)
	}
	if scorer.roleBonusValue != 0.05 {
		t.Errorf("expected default roleBonusValue 0.05, got %f", scorer.roleBonusValue)
	}
	if scorer.evolutionBonusWeight != 0.15 {
		t.Errorf("expected default evolutionBonusWeight 0.15, got %f", scorer.evolutionBonusWeight)
	}
	if scorer.rarityWeights == nil {
		t.Errorf("expected default rarity weights, got nil")
	}
}
