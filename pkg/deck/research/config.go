package research

import "fmt"

// HardConstraints define required deck composition checks.
type HardConstraints struct {
	MinWinConditions int `json:"min_win_conditions"`
	MinSpells        int `json:"min_spells"`
	MinAirDefense    int `json:"min_air_defense"`
	MinTankKillers   int `json:"min_tank_killers"`
}

// SoftWeights define objective weights for composite score components.
type SoftWeights struct {
	Synergy     float64 `json:"synergy"`
	Coverage    float64 `json:"coverage"`
	RoleFit     float64 `json:"role_fit"`
	ElixirFit   float64 `json:"elixir_fit"`
	CardQuality float64 `json:"card_quality"`
}

// ConstraintConfig controls hard requirements and soft objective weighting.
type ConstraintConfig struct {
	Hard HardConstraints `json:"hard"`
	Soft SoftWeights     `json:"soft"`
}

func defaultHardConstraints() HardConstraints {
	return HardConstraints{
		MinWinConditions: 1,
		MinSpells:        1,
		MinAirDefense:    2,
		MinTankKillers:   1,
	}
}

func defaultSoftWeights() SoftWeights {
	return SoftWeights{
		Synergy:     0.30,
		Coverage:    0.25,
		RoleFit:     0.20,
		ElixirFit:   0.15,
		CardQuality: 0.10,
	}
}

// DefaultConstraintConfig returns the phase-1 benchmark defaults.
func DefaultConstraintConfig() ConstraintConfig {
	return ConstraintConfig{
		Hard: defaultHardConstraints(),
		Soft: defaultSoftWeights(),
	}
}

func (c ConstraintConfig) normalizedSoftWeights() SoftWeights {
	sum := c.Soft.Synergy + c.Soft.Coverage + c.Soft.RoleFit + c.Soft.ElixirFit + c.Soft.CardQuality
	if sum <= 0 {
		return defaultSoftWeights()
	}
	return SoftWeights{
		Synergy:     c.Soft.Synergy / sum,
		Coverage:    c.Soft.Coverage / sum,
		RoleFit:     c.Soft.RoleFit / sum,
		ElixirFit:   c.Soft.ElixirFit / sum,
		CardQuality: c.Soft.CardQuality / sum,
	}
}

// Validate verifies hard bounds and soft-weight values.
//
//nolint:gocyclo // Explicit checks provide actionable per-field validation errors.
func (c ConstraintConfig) Validate() error {
	h := c.Hard
	if h.MinWinConditions < 0 || h.MinWinConditions > 8 {
		return fmt.Errorf("hard.min_win_conditions must be in [0,8], got %d", h.MinWinConditions)
	}
	if h.MinSpells < 0 || h.MinSpells > 8 {
		return fmt.Errorf("hard.min_spells must be in [0,8], got %d", h.MinSpells)
	}
	if h.MinAirDefense < 0 || h.MinAirDefense > 8 {
		return fmt.Errorf("hard.min_air_defense must be in [0,8], got %d", h.MinAirDefense)
	}
	if h.MinTankKillers < 0 || h.MinTankKillers > 8 {
		return fmt.Errorf("hard.min_tank_killers must be in [0,8], got %d", h.MinTankKillers)
	}

	s := c.Soft
	if s.Synergy < 0 || s.Coverage < 0 || s.RoleFit < 0 || s.ElixirFit < 0 || s.CardQuality < 0 {
		return fmt.Errorf("soft weights must be non-negative")
	}
	sum := s.Synergy + s.Coverage + s.RoleFit + s.ElixirFit + s.CardQuality
	if sum <= 0 {
		return fmt.Errorf("soft weights must sum to > 0")
	}
	return nil
}

func resolveConstraintConfig(cfg *ConstraintConfig) (ConstraintConfig, error) {
	if cfg == nil {
		def := DefaultConstraintConfig()
		return def, nil
	}
	if err := cfg.Validate(); err != nil {
		return ConstraintConfig{}, err
	}
	return *cfg, nil
}
