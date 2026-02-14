package research

import "testing"

func TestConstraintConfigValidateRejectsInvalidHardBounds(t *testing.T) {
	cfg := DefaultConstraintConfig()
	cfg.Hard.MinAirDefense = 9
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for hard bound")
	}
}

func TestConstraintConfigValidateRejectsNegativeSoftWeight(t *testing.T) {
	cfg := DefaultConstraintConfig()
	cfg.Soft.RoleFit = -0.1
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for negative soft weight")
	}
}

func TestConstraintConfigValidateRejectsZeroWeightSum(t *testing.T) {
	cfg := DefaultConstraintConfig()
	cfg.Soft = SoftWeights{}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for zero soft-weight sum")
	}
}

func TestConstraintConfigNormalizesSoftWeights(t *testing.T) {
	cfg := DefaultConstraintConfig()
	cfg.Soft = SoftWeights{
		Synergy:     3,
		Coverage:    2,
		RoleFit:     2,
		ElixirFit:   2,
		CardQuality: 1,
	}
	weights := cfg.normalizedSoftWeights()
	sum := weights.Synergy + weights.Coverage + weights.RoleFit + weights.ElixirFit + weights.CardQuality
	if sum < 0.999999 || sum > 1.000001 {
		t.Fatalf("expected normalized sum=1, got %f", sum)
	}
}
