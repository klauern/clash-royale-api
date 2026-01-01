package deck

import (
	"strings"
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

// TestValidateStrategyConfig tests the validateStrategyConfig function
func TestValidateStrategyConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    StrategyConfig
		shouldErr bool
		errMsg    string
	}{
		{
			name: "Valid config",
			config: StrategyConfig{
				TargetElixirMin: 3.0,
				TargetElixirMax: 3.5,
			},
			shouldErr: false,
		},
		{
			name: "Min below range",
			config: StrategyConfig{
				TargetElixirMin: -0.5,
				TargetElixirMax: 3.5,
			},
			shouldErr: true,
			errMsg:    "target_elixir_min must be between 0 and 10",
		},
		{
			name: "Min above range",
			config: StrategyConfig{
				TargetElixirMin: 11.0,
				TargetElixirMax: 3.5,
			},
			shouldErr: true,
			errMsg:    "target_elixir_min must be between 0 and 10",
		},
		{
			name: "Max below range",
			config: StrategyConfig{
				TargetElixirMin: 2.0,
				TargetElixirMax: -0.5,
			},
			shouldErr: true,
			errMsg:    "target_elixir_max must be between 0 and 10",
		},
		{
			name: "Max above range",
			config: StrategyConfig{
				TargetElixirMin: 2.0,
				TargetElixirMax: 11.0,
			},
			shouldErr: true,
			errMsg:    "target_elixir_max must be between 0 and 10",
		},
		{
			name: "Min greater than max",
			config: StrategyConfig{
				TargetElixirMin: 4.0,
				TargetElixirMax: 3.0,
			},
			shouldErr: true,
			errMsg:    "target_elixir_min",
		},
		{
			name: "Min equals max (valid)",
			config: StrategyConfig{
				TargetElixirMin: 3.0,
				TargetElixirMax: 3.0,
			},
			shouldErr: false,
		},
		{
			name: "Boundary values (valid)",
			config: StrategyConfig{
				TargetElixirMin: 0.0,
				TargetElixirMax: 10.0,
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStrategyConfig(&tt.config)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("validateStrategyConfig() expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateStrategyConfig() error = %v, want it to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateStrategyConfig() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestGetStrategyScaling tests the GetStrategyScaling function
func TestGetStrategyScaling(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected float64
	}{
		{"No env set", "", 1.0}, // Default is 1.0
		{"Zero value", "0", 0.0},
		{"Half value", "0.5", 0.5},
		{"Default value", "1.0", 1.0},
		{"Max value", "2.0", 2.0},
		{"Above max (clamped)", "3.0", 2.0},
		{"Negative (clamped)", "-1.0", 0.0},
		{"Invalid (uses default)", "invalid", 1.0},
		{"Partial invalid (uses default)", "1.5abc", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env variable for this test
			if tt.envValue != "" {
				t.Setenv("STRATEGY_BONUS_SCALE", tt.envValue)
			} else {
				// Clear the env var for the "no env set" test
				// Note: os.Unsetenv would be used here, but t.Setenv with empty string works too
			}

			result := GetStrategyScaling()
			if result != tt.expected {
				t.Errorf("GetStrategyScaling() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestScoreCardWithStrategy tests strategy-aware scoring
func TestScoreCardWithStrategy(t *testing.T) {
	supportRole := RoleSupport
	winConRole := RoleWinCondition

	tests := []struct {
		name          string
		card          CardCandidate
		role          *CardRole
		strategy      Strategy
		expectedMin   float64
		expectedMax   float64
	}{
		{
			name: "High level card with aggro strategy",
			card: CardCandidate{
				Name:     "Hog Rider",
				Level:    14,
				MaxLevel: 14,
				Rarity:   "Legendary",
				Elixir:   4,
			},
			role:        &winConRole,
			strategy:    StrategyAggro,
			expectedMin: 1.3, // Win condition gets bonus in aggro
			expectedMax: 2.5,
		},
		{
			name: "Support card with cycle strategy",
			card: CardCandidate{
				Name:     "Skeletons",
				Level:    14,
				MaxLevel: 14,
				Rarity:   "Common",
				Elixir:   1,
			},
			role:        &supportRole,
			strategy:    StrategyCycle,
			expectedMin: 0.8,
			expectedMax: 1.5,
		},
		{
			name: "Card without stats (uses fallback)",
			card: CardCandidate{
				Name:     "Test Card",
				Level:    10,
				MaxLevel: 14,
				Rarity:   "Rare",
				Elixir:   3,
				Stats:    nil,
			},
			role:        &supportRole,
			strategy:    StrategyBalanced,
			expectedMin: 0.7,
			expectedMax: 1.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategyConfig := GetStrategyConfig(tt.strategy)
			score := ScoreCardWithStrategy(&tt.card, tt.role, strategyConfig, nil)

			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("ScoreCardWithStrategy() = %v, want range [%v, %v]",
					score, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

// TestReloadStrategyConfig tests reloading strategy configuration
func TestReloadStrategyConfig(t *testing.T) {
	// Get original config
	originalConfig := GetStrategyConfig(StrategyBalanced)

	// Reload the config
	err := ReloadStrategyConfig()
	if err != nil {
		t.Fatalf("ReloadStrategyConfig failed: %v", err)
	}

	// Get reloaded config
	reloadedConfig := GetStrategyConfig(StrategyBalanced)

	// Values should be similar (might be exactly the same if file unchanged)
	if reloadedConfig.TargetElixirMin != originalConfig.TargetElixirMin {
		t.Logf("TargetElixirMin changed from %v to %v", originalConfig.TargetElixirMin, reloadedConfig.TargetElixirMin)
	}
}
