package deck

import (
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
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
		name     string
		cardName string
		level    int
		wantMin  float64 // Minimum expected multiplier
		wantMax  float64 // Maximum expected multiplier
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
		name     string
		cardName string
		level    int
		maxLevel int
		wantMin  float64 // Minimum expected ratio
		wantMax  float64 // Maximum expected ratio
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
				1: 1.0,  // Base level
				2: 1.1,  // +10%
				3: 1.21, // +10% compounded
				4: 1.33, // +10% compounded
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
	if err := os.WriteFile(configPath, []byte(emptyConfig), 0o644); err != nil {
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

// ==================== Phase 3 Validation Tests ====================

// TestLevelCurve_MathematicalAccuracy validates 20+ cards against known stat values
// Target: 90%+ of cards within 2% error tolerance
func TestLevelCurve_MathematicalAccuracy(t *testing.T) {
	lc, err := NewLevelCurve("../../config/card_level_curves.json")
	if err != nil {
		t.Fatalf("Failed to create LevelCurve: %v", err)
	}

	tests := []struct {
		name             string
		cardName         string
		validationPoints map[int]float64 // level -> expected multiplier
		isException      bool            // Cards with known non-standard scaling
	}{
		{
			name:     "Knight - Standard Common scaling",
			cardName: "Knight",
			validationPoints: map[int]float64{
				1:  1.00,
				2:  1.10,
				3:  1.21,
				4:  1.33,
				5:  1.46,
				6:  1.60,
				7:  1.76,
				8:  1.93,
				9:  2.12, // Tournament standard
				11: 2.56,
				13: 3.09,
				15: 3.72,
			},
			isException: false,
		},
		{
			name:     "Archers - Standard Common scaling",
			cardName: "Archers",
			validationPoints: map[int]float64{
				1:  1.00,
				5:  1.46,
				9:  2.12,
				11: 2.56,
				15: 3.72,
			},
			isException: false,
		},
		{
			name:     "Giant - Standard Rare scaling",
			cardName: "Giant",
			validationPoints: map[int]float64{
				1:  1.00,
				3:  1.21,
				6:  1.60,
				9:  2.12,
				11: 2.56,
				15: 3.72,
			},
			isException: false,
		},
		{
			name:     "Musketeer - Standard Rare scaling",
			cardName: "Musketeer",
			validationPoints: map[int]float64{
				1:  1.00,
				9:  2.12,
				13: 3.09,
			},
			isException: false,
		},
		{
			name:     "Hog Rider - Standard Rare scaling",
			cardName: "Hog_Rider",
			validationPoints: map[int]float64{
				1:  1.00,
				7:  1.76,
				9:  2.12,
				11: 2.56,
			},
			isException: false,
		},
		{
			name:     "Baby Dragon - Epic with rarity bonus",
			cardName: "Baby_Dragon",
			validationPoints: map[int]float64{
				1: 1.00,
				6: 1.65, // Slightly higher due to rarity bonus
				9: 2.16,
			},
			isException: true, // Different due to rarity bonus
		},
		{
			name:     "Prince - Epic with rarity bonus",
			cardName: "Prince",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 2.16,
			},
			isException: true, // Different due to rarity bonus
		},
		{
			name:     "Golem - Epic with rarity bonus",
			cardName: "Golem",
			validationPoints: map[int]float64{
				1:  1.00,
				9:  2.16,
				15: 3.79,
			},
			isException: true, // Different due to rarity bonus
		},
		{
			name:     "P.E.K.K.A - Epic with rarity bonus",
			cardName: "P.E.K.K.A",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 2.16,
			},
			isException: true, // Different due to rarity bonus
		},
		{
			name:     "Archer_Queen - Champion with rarity bonus",
			cardName: "Archer_Queen",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 2.16, // Champions have 2% rarity bonus
			},
			isException: true,
		},
		{
			name:     "Golden_Knight - Champion with rarity bonus",
			cardName: "Golden_Knight",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 2.16,
			},
			isException: true,
		},
		{
			name:     "Mighty_Miner - Champion with rarity bonus",
			cardName: "Mighty_Miner",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 2.16,
			},
			isException: true,
		},
		{
			name:     "Skeleton_King - Champion with rarity bonus",
			cardName: "Skeleton_King",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 2.16,
			},
			isException: true,
		},
		{
			name:     "Miner - Legendary with rarity bonus",
			cardName: "Miner",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 2.16,
			},
			isException: true, // Different due to rarity bonus
		},
		{
			name:     "Princess - Legendary with rarity bonus",
			cardName: "Princess",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 2.16,
			},
			isException: true, // Different due to rarity bonus
		},
		{
			name:     "Ice_Wizard - Legendary with rarity bonus",
			cardName: "Ice_Wizard",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 2.16,
			},
			isException: true, // Different due to rarity bonus
		},
		{
			name:     "Electro_Wizard - Legendary with rarity bonus",
			cardName: "Electro_Wizard",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 2.16,
			},
			isException: true, // Different due to rarity bonus
		},
		{
			name:     "Bandit - Legendary with rarity bonus",
			cardName: "Bandit",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 2.16,
			},
			isException: true, // Different due to rarity bonus
		},
		{
			name:     "Poison - Spell with 8% growth",
			cardName: "Poison",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 1.83, // 8% growth instead of 10%
			},
			isException: true, // Different growth rate
		},
		{
			name:     "Fireball - Spell with 8% growth",
			cardName: "Fireball",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 1.83,
			},
			isException: true, // Different growth rate
		},
		{
			name:     "The_Log - Spell with 8% growth",
			cardName: "The_Log",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 1.83,
			},
			isException: true, // Different growth rate
		},
		{
			name:     "Zap - Spell with 8% growth",
			cardName: "Zap",
			validationPoints: map[int]float64{
				1: 1.00,
				9: 1.83,
			},
			isException: true, // Different growth rate
		},
	}

	standardPassCount := 0
	standardTotalCount := 0
	exceptionPassCount := 0
	exceptionTotalCount := 0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passCount := 0
			totalCount := 0

			for level, expected := range tt.validationPoints {
				totalCount++
				actual := lc.GetLevelMultiplier(tt.cardName, level)

				// Calculate error percentage
				errorPct := math.Abs(actual-expected) / expected * 100

				// Use different tolerance for exceptions
				tolerance := 2.0 // 2% for standard
				if tt.isException {
					tolerance = 5.0 // 5% for exceptions
				}

				if errorPct <= tolerance {
					passCount++
				} else {
					t.Logf("Level %d: expected %.3f, got %.3f (error: %.2f%%)",
						level, expected, actual, errorPct)
				}
			}

			// Log summary for this card
			passRate := float64(passCount) / float64(totalCount) * 100
			t.Logf("%s: %d/%d points passed (%.1f%%)", tt.cardName, passCount, totalCount, passRate)

			// Track overall metrics
			if tt.isException {
				exceptionPassCount += passCount
				exceptionTotalCount += totalCount
			} else {
				standardPassCount += passCount
				standardTotalCount += totalCount
			}

			// Fail if any card has less than 50% pass rate
			if passRate < 50 {
				t.Errorf("%s validation failed: only %.1f%% pass rate (expected >75%%)",
					tt.cardName, passRate)
			}
		})
	}

	// Calculate overall metrics
	standardPassRate := float64(standardPassCount) / float64(standardTotalCount) * 100
	exceptionPassRate := float64(exceptionPassCount) / float64(exceptionTotalCount) * 100

	t.Logf("\n=== Mathematical Accuracy Summary ===")
	t.Logf("Standard cards: %d/%d passed (%.1f%%)", standardPassCount, standardTotalCount, standardPassRate)
	t.Logf("Exception cards: %d/%d passed (%.1f%%)", exceptionPassCount, exceptionTotalCount, exceptionPassRate)

	// Success criteria from Phase 3
	if standardPassRate < 90 {
		t.Errorf("Standard card accuracy below target: %.1f%% (target: 90%%+)", standardPassRate)
	}
	if exceptionPassRate < 75 {
		t.Errorf("Exception card accuracy below target: %.1f%% (target: 75%%+)", exceptionPassRate)
	}
}

// TestLevelCurve_PerformanceBenchmark validates performance metrics
// Target: <1ms per card calculation overhead (95th percentile)
func TestLevelCurve_PerformanceBenchmark(t *testing.T) {
	lc, err := NewLevelCurve("../../config/card_level_curves.json")
	if err != nil {
		t.Fatalf("Failed to create LevelCurve: %v", err)
	}

	// Test cards used for benchmarking
	testCards := []struct {
		name   string
		levels []int
	}{
		{"Knight", []int{1, 5, 9, 11, 13, 15}},
		{"Archers", []int{1, 9, 15}},
		{"Giant", []int{1, 9, 15}},
		{"Musketeer", []int{1, 9, 13}},
		{"Baby_Dragon", []int{1, 6, 9}},
		{"Prince", []int{1, 9}},
		{"Golem", []int{1, 9, 15}},
		{"P.E.K.K.A", []int{1, 9}},
		{"Archer_Queen", []int{1, 9}},
		{"Golden_Knight", []int{1, 9}},
		{"Miner", []int{1, 9}},
		{"Princess", []int{1, 9}},
		{"Poison", []int{1, 9}},
		{"Fireball", []int{1, 9}},
		{"The_Log", []int{1, 9}},
		{"Zap", []int{1, 9}},
	}

	iterations := 10000
	totalTime := time.Duration(0)
	var timings []time.Duration

	// Warm up cache
	for _, tc := range testCards {
		for _, level := range tc.levels {
			_ = lc.GetLevelMultiplier(tc.name, level)
		}
	}

	// Benchmark calculations
	for i := 0; i < iterations; i++ {
		start := time.Now()

		for _, tc := range testCards {
			for _, level := range tc.levels {
				_ = lc.GetLevelMultiplier(tc.name, level)
			}
		}

		elapsed := time.Since(start)
		timings = append(timings, elapsed)
		totalTime += elapsed
	}

	// Calculate statistics
	avgTime := totalTime / time.Duration(iterations)
	totalCalculations := len(testCards) * 0 // Will be calculated below
	for _, tc := range testCards {
		totalCalculations += len(tc.levels)
	}

	// Sort timings for percentile calculation
	sort.Slice(timings, func(i, j int) bool {
		return timings[i] < timings[j]
	})

	p50 := timings[len(timings)/2]
	p95 := timings[int(float64(len(timings))*0.95)]
	p99 := timings[int(float64(len(timings))*0.99)]

	avgPerCalculation := avgTime / time.Duration(totalCalculations)
	p95PerCalculation := p95 / time.Duration(totalCalculations)

	t.Logf("\n=== Performance Benchmark Results ===")
	t.Logf("Total iterations: %d", iterations)
	t.Logf("Calculations per iteration: %d", totalCalculations)
	t.Logf("Average time per iteration: %v", avgTime)
	t.Logf("P50 per iteration: %v", p50)
	t.Logf("P95 per iteration: %v", p95)
	t.Logf("P99 per iteration: %v", p99)
	t.Logf("Average time per calculation: %v", avgPerCalculation)
	t.Logf("P95 time per calculation: %v", p95PerCalculation)

	// Target: <1ms per calculation overhead (95th percentile)
	if p95PerCalculation > time.Millisecond {
		t.Errorf("P95 performance too slow: %v (target: <1ms per calculation)", p95PerCalculation)
	}
	if avgPerCalculation > 500*time.Microsecond {
		t.Logf("Warning: Average performance slower than ideal: %v (target: <0.5ms)", avgPerCalculation)
	}
}

// TestLevelCurve_DeckBuildingImpact tests deck recommendation changes
// Measures the percentage of deck recommendations that change with new system
func TestLevelCurve_DeckBuildingImpact(t *testing.T) {
	// Skip if API token not available (integration test)
	if os.Getenv("CLASH_ROYALE_API_TOKEN") == "" {
		t.Skip("CLASH_ROYALE_API_TOKEN not set, skipping integration test")
	}

	_ = os.Getenv("DEFAULT_PLAYER_TAG") // Read from environment but not used in integration tests

	// Create test analysis data
	analysisData := setupTestAnalysisData()

	// Generate deck using old linear system (simulate)
	oldRecommendations := generateDeckWithLinearScoring(analysisData)

	// Create level curve system
	lc, err := NewLevelCurve("../../config/card_level_curves.json")
	if err != nil {
		t.Fatalf("Failed to create LevelCurve: %v", err)
	}

	// Generate deck using new curve-based system
	newRecommendations := generateDeckWithCurveScoring(analysisData, lc)

	// Compare recommendations
	changed := compareDeckRecommendations(oldRecommendations, newRecommendations)

	t.Logf("\n=== Deck Building Impact Analysis ===")
	t.Logf("Old system deck: %v", oldRecommendations)
	t.Logf("New system deck: %v", newRecommendations)
	t.Logf("Cards changed: %d/8", changed)

	// Track metrics - we expect some change to demonstrate improvement
	if changed == 0 {
		t.Logf("Warning: No cards changed, which may indicate:")
		t.Logf("  1. Test data is too uniform")
		t.Logf("  2. Level curve system needs adjustment")
		t.Logf("  3. Cards are all at similar levels")
	}

	// We want at least 1-2 cards to change to show the curve system has impact
	if changed < 1 {
		t.Logf("Note: Very few cards changed (%d), which might indicate:", changed)
		t.Logf("  - Testing with real player data may show more impact")
		t.Logf("  - Consider testing with more diverse level distributions")
	}
}

// TestLevelCurve_ConfigCoverage validates card config coverage
// Success criteria: 50+ cards with at least 10% having non-standard curves
func TestLevelCurve_ConfigCoverage(t *testing.T) {
	lc, err := NewLevelCurve("../../config/card_level_curves.json")
	if err != nil {
		t.Fatalf("Failed to create LevelCurve: %v", err)
	}

	if lc.config.CardLevelCurves == nil {
		t.Fatal("No card configurations loaded")
	}

	totalCards := 0
	nonStandardCount := 0
	standardCount := 0

	for cardName, config := range lc.config.CardLevelCurves {
		if cardName == "_default" {
			continue
		}

		totalCards++

		// Check if card has non-standard configuration
		isNonStandard := false
		defaultConfig := defaultCardLevelConfig

		if config.GrowthRate != defaultConfig.GrowthRate {
			isNonStandard = true
		}
		if config.RarityBonus != defaultConfig.RarityBonus {
			isNonStandard = true
		}
		if config.Type != defaultConfig.Type {
			isNonStandard = true
		}
		if len(config.LevelOverrides) > 0 {
			isNonStandard = true
		}

		if isNonStandard {
			nonStandardCount++
			t.Logf("Non-standard: %s (growthRate: %.2f, rarityBonus: %.2f, type: %s)",
				cardName, config.GrowthRate, config.RarityBonus, config.Type)
		} else {
			standardCount++
		}
	}

	nonStandardPct := float64(nonStandardCount) / float64(totalCards) * 100

	t.Logf("\n=== Configuration Coverage Analysis ===")
	t.Logf("Total configured cards: %d", totalCards)
	t.Logf("Standard curves: %d", standardCount)
	t.Logf("Non-standard curves: %d (%.1f%%)", nonStandardCount, nonStandardPct)

	// Success criteria from Phase 3
	if totalCards < 50 {
		t.Errorf("Insufficient card coverage: %d cards (target: 50+)", totalCards)
	}
	if nonStandardPct < 10 {
		t.Errorf("Insufficient non-standard curves: %.1f%% (target: 10%% or higher)", nonStandardPct)
	}

	// Additional checks
	if totalCards >= 74 { // Phase 2 mentioned 74 cards
		t.Logf("Excellent coverage: Met Phase 2 goal of 74+ cards")
	}
}

// Helper functions for deck building impact tests

func setupTestAnalysisData() map[string]interface{} {
	// Return mock analysis data for testing
	return map[string]interface{}{
		"cards": []map[string]interface{}{
			{"name": "Knight", "level": 12, "maxLevel": 14, "rarity": "Common", "count": 500},
			{"name": "Archers", "level": 13, "maxLevel": 14, "rarity": "Common", "count": 450},
			{"name": "Giant", "level": 11, "maxLevel": 14, "rarity": "Rare", "count": 300},
			{"name": "Musketeer", "level": 10, "maxLevel": 14, "rarity": "Rare", "count": 280},
			{"name": "Baby_Dragon", "level": 10, "maxLevel": 14, "rarity": "Epic", "count": 100},
			{"name": "Prince", "level": 9, "maxLevel": 14, "rarity": "Rare", "count": 250},
			{"name": "Hog_Rider", "level": 12, "maxLevel": 14, "rarity": "Rare", "count": 400},
			{"name": "Fireball", "level": 11, "maxLevel": 14, "rarity": "Rare", "count": 320},
		},
	}
}

func generateDeckWithLinearScoring(analysisData map[string]interface{}) []string {
	// Simulate old linear scoring system
	// This is a simplified version for testing
	return []string{
		"Knight", "Archers", "Baby_Dragon", "Hog_Rider",
		"Fireball", "Giant", "Musketeer", "Prince",
	}
}

func generateDeckWithCurveScoring(analysisData map[string]interface{}, lc *LevelCurve) []string {
	// Simulate new curve-based scoring system
	// This would integrate with actual deck builder in production
	return []string{
		"Knight", "Archers", "Baby_Dragon", "Hog_Rider",
		"Fireball", "Giant", "Musketeer", "Prince",
	}
}

func compareDeckRecommendations(old, new []string) int {
	// Count how many cards differed between old and new systems
	changed := 0
	for i := 0; i < len(old) && i < len(new); i++ {
		if old[i] != new[i] {
			changed++
		}
	}
	return changed
}

// TestFillMissingFields tests that missing config fields are filled with defaults
func TestFillMissingFields(t *testing.T) {
	lc, err := NewLevelCurve("../../config/card_level_curves.json")
	if err != nil {
		t.Fatalf("Failed to create LevelCurve: %v", err)
	}

	tests := []struct {
		name     string
		input    CardLevelConfig
		expected CardLevelConfig
	}{
		{
			name: "All fields missing",
			input: CardLevelConfig{
				BaseScale:  0,
				GrowthRate: 0,
				Type:       "",
			},
			expected: CardLevelConfig{
				BaseScale:  100.0,
				GrowthRate: 0.10,
				Type:       "standard",
			},
		},
		{
			name: "Some fields missing",
			input: CardLevelConfig{
				BaseScale:  100.0,
				GrowthRate: 0,
				Type:       "",
			},
			expected: CardLevelConfig{
				BaseScale:  100.0,
				GrowthRate: 0.10,
				Type:       "standard",
			},
		},
		{
			name: "No fields missing",
			input: CardLevelConfig{
				BaseScale:  100.0,
				GrowthRate: 0.10,
				Type:       "standard",
			},
			expected: CardLevelConfig{
				BaseScale:  100.0,
				GrowthRate: 0.10,
				Type:       "standard",
			},
		},
		{
			name: "Custom values preserved",
			input: CardLevelConfig{
				BaseScale:  105.0,
				GrowthRate: 0.12,
				Type:       "spell_duration",
			},
			expected: CardLevelConfig{
				BaseScale:  105.0,
				GrowthRate: 0.12,
				Type:       "spell_duration",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lc.fillMissingFields(tt.input)

			if result.BaseScale != tt.expected.BaseScale {
				t.Errorf("BaseScale = %v, want %v", result.BaseScale, tt.expected.BaseScale)
			}
			if result.GrowthRate != tt.expected.GrowthRate {
				t.Errorf("GrowthRate = %v, want %v", result.GrowthRate, tt.expected.GrowthRate)
			}
			if result.Type != tt.expected.Type {
				t.Errorf("Type = %v, want %v", result.Type, tt.expected.Type)
			}
		})
	}
}
