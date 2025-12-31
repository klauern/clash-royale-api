package deck

import (
	"testing"
)

func TestParseStrategy(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  Strategy
		shouldErr bool
	}{
		{"Balanced lowercase", "balanced", StrategyBalanced, false},
		{"Aggro uppercase", "AGGRO", StrategyAggro, false},
		{"Control mixed case", "Control", StrategyControl, false},
		{"Cycle with spaces", "  cycle  ", StrategyCycle, false},
		{"Splash normal", "splash", StrategySplash, false},
		{"Spell uppercase", "SPELL", StrategySpell, false},
		{"Invalid strategy", "invalid", "", true},
		{"Empty string", "", "", true},
		{"Random text", "random", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseStrategy(tt.input)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("ParseStrategy(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseStrategy(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ParseStrategy(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestStrategyValidate(t *testing.T) {
	tests := []struct {
		name      string
		strategy  Strategy
		shouldErr bool
	}{
		{"Valid balanced", StrategyBalanced, false},
		{"Valid aggro", StrategyAggro, false},
		{"Valid control", StrategyControl, false},
		{"Valid cycle", StrategyCycle, false},
		{"Valid splash", StrategySplash, false},
		{"Valid spell", StrategySpell, false},
		{"Invalid strategy", Strategy("invalid"), true},
		{"Empty strategy", Strategy(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.strategy.Validate()

			if tt.shouldErr && err == nil {
				t.Errorf("%v.Validate() expected error, got nil", tt.strategy)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("%v.Validate() unexpected error: %v", tt.strategy, err)
			}
		})
	}
}

func TestGetStrategyConfig(t *testing.T) {
	tests := []struct {
		name                string
		strategy            Strategy
		expectedElixirMin   float64
		expectedElixirMax   float64
		expectedMultipliers map[CardRole]float64
		hasOverrides        bool
	}{
		{
			name:              "Balanced strategy",
			strategy:          StrategyBalanced,
			expectedElixirMin: 3.0,
			expectedElixirMax: 3.5,
			expectedMultipliers: map[CardRole]float64{
				RoleWinCondition: 1.0,
				RoleBuilding:     1.0,
				RoleSpellBig:     1.0,
				RoleSpellSmall:   1.0,
				RoleSupport:      1.0,
				RoleCycle:        1.0,
			},
			hasOverrides: false,
		},
		{
			name:              "Aggro strategy",
			strategy:          StrategyAggro,
			expectedElixirMin: 3.5,
			expectedElixirMax: 4.0,
			expectedMultipliers: map[CardRole]float64{
				RoleWinCondition: 2.0,
				RoleSupport:      1.2,
				RoleBuilding:     0.3,
			},
			hasOverrides: true,
		},
		{
			name:              "Control strategy",
			strategy:          StrategyControl,
			expectedElixirMin: 3.5,
			expectedElixirMax: 4.2,
			expectedMultipliers: map[CardRole]float64{
				RoleBuilding:     2.0,
				RoleSpellBig:     1.5,
				RoleSpellSmall:   0.3,
				RoleWinCondition: 0.5,
				RoleCycle:        0.5,
			},
			hasOverrides: true,
		},
		{
			name:              "Cycle strategy",
			strategy:          StrategyCycle,
			expectedElixirMin: 2.5,
			expectedElixirMax: 3.0,
			expectedMultipliers: map[CardRole]float64{
				RoleCycle:      2.0,
				RoleSpellSmall: 1.2,
				RoleSpellBig:   0.3,
			},
			hasOverrides: true,
		},
		{
			name:              "Splash strategy",
			strategy:          StrategySplash,
			expectedElixirMin: 3.2,
			expectedElixirMax: 3.8,
			expectedMultipliers: map[CardRole]float64{
				RoleSupport:  2.0,
				RoleSpellBig: 1.2,
				RoleCycle:    0.5,
			},
			hasOverrides: true,
		},
		{
			name:              "Spell strategy",
			strategy:          StrategySpell,
			expectedElixirMin: 3.2,
			expectedElixirMax: 3.8,
			expectedMultipliers: map[CardRole]float64{
				RoleSpellBig:   2.0,
				RoleSpellSmall: 1.5,
				RoleBuilding:   0.1,
			},
			hasOverrides: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetStrategyConfig(tt.strategy)

			// Check elixir range
			if config.TargetElixirMin != tt.expectedElixirMin {
				t.Errorf("TargetElixirMin = %v, want %v", config.TargetElixirMin, tt.expectedElixirMin)
			}
			if config.TargetElixirMax != tt.expectedElixirMax {
				t.Errorf("TargetElixirMax = %v, want %v", config.TargetElixirMax, tt.expectedElixirMax)
			}

			// Check role multipliers
			for role, expected := range tt.expectedMultipliers {
				actual := config.RoleMultipliers[role]
				if actual != expected {
					t.Errorf("RoleMultipliers[%v] = %v, want %v", role, actual, expected)
				}
			}

			// Check overrides
			if tt.hasOverrides {
				if config.CompositionOverrides == nil {
					t.Error("Expected CompositionOverrides, got nil")
				}
			} else {
				if config.CompositionOverrides != nil {
					t.Error("Expected no CompositionOverrides, got non-nil")
				}
			}
		})
	}
}

func TestStrategyString(t *testing.T) {
	tests := []struct {
		strategy Strategy
		expected string
	}{
		{StrategyBalanced, "balanced"},
		{StrategyAggro, "aggro"},
		{StrategyControl, "control"},
		{StrategyCycle, "cycle"},
		{StrategySplash, "splash"},
		{StrategySpell, "spell"},
	}

	for _, tt := range tests {
		t.Run(string(tt.strategy), func(t *testing.T) {
			result := tt.strategy.String()
			if result != tt.expected {
				t.Errorf("%v.String() = %v, want %v", tt.strategy, result, tt.expected)
			}
		})
	}
}
