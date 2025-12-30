package deck

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewLevelCurve tests the creation of a new LevelCurve instance
func TestNewLevelCurve(t *testing.T) {
	tests := []struct {
		name      string
		useConfig bool
		wantErr   bool
	}{
		{
			name:      "With valid config file",
			useConfig: true,
			wantErr:   false,
		},
		{
			name:      "Without config file (defaults)",
			useConfig: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configPath string
			if tt.useConfig {
				configPath = "../../config/card_level_curves.json"
			} else {
				// Use non-existent path to trigger default behavior
				configPath = "/tmp/non_existent_config.json"
			}

			lc, err := NewLevelCurve(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLevelCurve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && lc == nil {
				t.Error("NewLevelCurve() returned nil without error")
			}
		})
	}
}

// TestLevelCurve_GetLevelMultiplier tests the GetLevelMultiplier method
func TestLevelCurve_GetLevelMultiplier(t *testing.T) {
	lc, err := NewLevelCurve("../../config/card_level_curves.json")
	if err != nil {
		t.Fatalf("Failed to create LevelCurve: %v", err)
	}

	tests := []struct {
		name        string
		cardName    string
		level       int
		wantMin     float64 // Minimum expected multiplier
		wantMax     float64 // Maximum expected multiplier
	}{
		{
			name:     "Knight level 1",
			cardName: "Knight",
			level:    1,
			wantMin:  0.9,
			wantMax:  1.1,
		},
		{
			name:     "Knight level 9 (tournament standard)",
			cardName: "Knight",
			level:    9,
			wantMin:  2.0,
			wantMax:  2.2,
		},
		{
			name:     "Knight level 15 (max)",
			cardName: "Knight",
			level:    15,
			wantMin:  3.5,
			wantMax:  4.0,
		},
		{
			name:     "Archer Queen level 1 (champion)",
			cardName: "Archer_Queen",
			level:    1,
			wantMin:  0.9,
			wantMax:  1.1,
		},
		{
			name:     "Archer Queen level 9 (champion)",
			cardName: "Archer_Queen",
			level:    9,
			wantMin:  2.1,
			wantMax:  2.3,
		},
		{
			name:     "Rage spell with overrides",
			cardName: "Rage",
			level:    9,
			wantMin:  1.58,
			wantMax:  1.60,
		},
		{
			name:     "Non-existent card uses default",
			cardName: "NonExistentCard",
			level:    9,
			wantMin:  2.0,
			wantMax:  2.2,
		},
		{
			name:     "Level 0 returns 0",
			cardName: "Knight",
			level:    0,
			wantMin:  0.0,
			wantMax:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lc.GetLevelMultiplier(tt.cardName, tt.level)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("GetLevelMultiplier() = %v, want range [%v, %v]",
					got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestLevelCurve_GetRelativeLevelRatio tests the GetRelativeLevelRatio method
func TestLevelCurve_GetRelativeLevelRatio(t *testing.T) {
	lc, err := NewLevelCurve("../../config/card_level_curves.json")
	if err != nil {
		t.Fatalf("Failed to create LevelCurve: %v", err)
	}

	tests := []struct {
		name      string
		cardName  string
		level     int
		maxLevel  int
		wantMin   float64 // Minimum expected ratio
		wantMax   float64 // Maximum expected ratio
	}{
		{
			name:     "Knight level 1 of 15",
			cardName: "Knight",
			level:    1,
			maxLevel: 15,
			wantMin:  0.25,
			wantMax:  0.30,
		},
		{
			name:     "Knight level 9 of 15",
			cardName: "Knight",
			level:    9,
			maxLevel: 15,
			wantMin:  0.55,
			wantMax:  0.65,
		},
		{
			name:     "Knight level 15 of 15",
			cardName: "Knight",
			level:    15,
			maxLevel: 15,
			wantMin:  0.95,
			wantMax:  1.05,
		},
		{
			name:     "Archer Queen level 1 of 9",
			cardName: "Archer_Queen",
			level:    1,
			maxLevel: 9,
			wantMin:  0.40,
			wantMax:  0.50,
		},
		{
			name:     "Archer Queen level 9 of 9",
			cardName: "Archer_Queen",
			level:    9,
			maxLevel: 9,
			wantMin:  0.95,
			wantMax:  1.05,
		},
		{
			name:     "maxLevel 0 returns 0",
			cardName: "Knight",
			level:    9,
			maxLevel: 0,
			wantMin:  0.0,
			wantMax:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lc.GetRelativeLevelRatio(tt.cardName, tt.level, tt.maxLevel)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("GetRelativeLevelRatio() = %v, want range [%v, %v]",
					got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestLevelCurve_ValidateCard tests card validation against known stats
func TestLevelCurve_ValidateCard(t *testing.T) {
	lc, err := NewLevelCurve("../../config/card_level_curves.json")
	if err != nil {
		t.Fatalf("Failed to create LevelCurve: %v", err)
	}

	tests := []struct {
		name             string
		cardName         string
		validationPoints map[int]float64
		expectError      bool
	}{
		{
			name:     "Knight standard curve validation",
			cardName: "Knight",
			validationPoints: map[int]float64{
				1: 1.0,   // Base level
				2: 1.1,   // +10%
				3: 1.21,  // +10% compounded
				4: 1.33,  // +10% compounded
				5: 1.46,
				6: 1.60,
				7: 1.76,
				8: 1.93,
				9: 2.12, // Tournament standard
			},
			expectError: false,
		},
		{
			name:     "Rage spell with overrides",
			cardName: "Rage",
			validationPoints: map[int]float64{
				1: 1.00,
				2: 1.07,
				3: 1.14,
				9: 1.59,
			},
			expectError: false,
		},
		{
			name:     "Invalid validation (should fail tolerance)",
			cardName: "Knight",
			validationPoints: map[int]float64{
				9: 5.0, // Way off from expected ~2.12
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := lc.ValidateCard(tt.cardName, tt.validationPoints)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateCard() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestLevelCurve_ExponentialGrowth tests that the curve follows exponential pattern
func TestLevelCurve_ExponentialGrowth(t *testing.T) {
	lc, err := NewLevelCurve("../../config/card_level_curves.json")
	if err != nil {
		t.Fatalf("Failed to create LevelCurve: %v", err)
	}

	// Test that each level is approximately 10% higher than previous
	prevMultiplier := lc.GetLevelMultiplier("Knight", 1)
	for level := 2; level <= 10; level++ {
		currentMultiplier := lc.GetLevelMultiplier("Knight", level)
		growthRate := (currentMultiplier - prevMultiplier) / prevMultiplier

		// Should be approximately 10% growth (0.10)
		// Allow some tolerance for floating point precision
		if growthRate < 0.09 || growthRate > 0.11 {
			t.Errorf("Level %d growth rate = %.3f, expected ~0.10", level, growthRate)
		}

		prevMultiplier = currentMultiplier
	}
}

// TestLevelCurve_ConfigCaching tests that card configurations are properly cached
func TestLevelCurve_ConfigCaching(t *testing.T) {
	lc, err := NewLevelCurve("../../config/card_level_curves.json")
	if err != nil {
		t.Fatalf("Failed to create LevelCurve: %v", err)
	}

	// Access same card multiple times
	for i := 0; i < 5; i++ {
		multiplier := lc.GetLevelMultiplier("Knight", 9)
		expectedMin := 2.0
		expectedMax := 2.2

		if multiplier < expectedMin || multiplier > expectedMax {
			t.Errorf("Iteration %d: GetLevelMultiplier() = %v, want range [%v, %v]",
				i, multiplier, expectedMin, expectedMax)
		}
	}
}

// TestLevelCurve_DefaultConfigFallback tests fallback to default configuration
func TestLevelCurve_DefaultConfigFallback(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "empty_config.json")

	// Create empty config
	emptyConfig := `{"cardLevelCurves": {}}`
	if err := os.WriteFile(configPath, []byte(emptyConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	lc, err := NewLevelCurve(configPath)
	if err != nil {
		t.Fatalf("Failed to create LevelCurve: %v", err)
	}

	// Should fall back to defaults for unknown cards
	multiplier := lc.GetLevelMultiplier("UnknownCard", 9)
	expectedMin := 2.0
	expectedMax := 2.2

	if multiplier < expectedMin || multiplier > expectedMax {
		t.Errorf("GetLevelMultiplier() for unknown card = %v, want range [%v, %v]",
			multiplier, expectedMin, expectedMax)
	}
}

// TestLevelCurve_RarityBonus tests that rarity bonus is applied correctly
func TestLevelCurve_RarityBonus(t *testing.T) {
	lc, err := NewLevelCurve("../../config/card_level_curves.json")
	if err != nil {
		t.Fatalf("Failed to create LevelCurve: %v", err)
	}

	// Champions should have higher growth due to rarity bonus
	knightMultiplier := lc.GetLevelMultiplier("Knight", 9)
	queenMultiplier := lc.GetLevelMultiplier("Archer_Queen", 9)

	// Archer Queen should have higher multiplier than Knight at same level
	if queenMultiplier <= knightMultiplier {
		t.Errorf("Archer_Queen multiplier (%.3f) should be > Knight multiplier (%.3f)",
			queenMultiplier, knightMultiplier)
	}
}

// TestLevelCurve_MissingConfigFile tests behavior with missing config file
func TestLevelCurve_MissingConfigFile(t *testing.T) {
	lc, err := NewLevelCurve("/non/existent/path/config.json")
	if err != nil {
		t.Fatalf("Failed to create LevelCurve with missing config: %v", err)
	}

	// Should work with default configuration
	multiplier := lc.GetLevelMultiplier("AnyCard", 9)
	expectedMin := 2.0
	expectedMax := 2.2

	if multiplier < expectedMin || multiplier > expectedMax {
		t.Errorf("GetLevelMultiplier() with missing config = %v, want range [%v, %v]",
			multiplier, expectedMin, expectedMax)
	}
}

// BenchmarkLevelCurve_GetLevelMultiplier benchmarks the level multiplier calculation
func BenchmarkLevelCurve_GetLevelMultiplier(b *testing.B) {
	lc, err := NewLevelCurve("../../config/card_level_curves.json")
	if err != nil {
		b.Fatalf("Failed to create LevelCurve: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lc.GetLevelMultiplier("Knight", 9)
	}
}

// BenchmarkLevelCurve_GetRelativeLevelRatio benchmarks the relative level ratio calculation
func BenchmarkLevelCurve_GetRelativeLevelRatio(b *testing.B) {
	lc, err := NewLevelCurve("../../config/card_level_curves.json")
	if err != nil {
		b.Fatalf("Failed to create LevelCurve: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lc.GetRelativeLevelRatio("Knight", 9, 15)
	}
}
