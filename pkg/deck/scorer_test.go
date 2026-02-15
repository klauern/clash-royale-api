package deck

import (
	"math"
	"os"
	"testing"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// TestScoreCard validates the scoring algorithm matches expected Python behavior
func TestScoreCard(t *testing.T) {
	tests := []struct {
		name         string
		level        int
		maxLevel     int
		rarity       string
		elixir       int
		role         *CardRole
		wantScoreMin float64
		wantScoreMax float64
	}{
		{
			name:         "High level common with role",
			level:        11,
			maxLevel:     14,
			rarity:       "Common",
			elixir:       3,
			role:         new(RoleCycle),
			wantScoreMin: 1.14, // (11/14 * 1.2 * 1.0) + (1.0 * 0.15) + 0.05
			wantScoreMax: 1.15,
		},
		{
			name:         "Max level legendary no role",
			level:        14,
			maxLevel:     14,
			rarity:       "Legendary",
			elixir:       4,
			role:         nil,
			wantScoreMin: 1.51, // (1.0 * 1.2 * 1.15) + (0.89 * 0.15)
			wantScoreMax: 1.52,
		},
		{
			name:         "Low level epic high cost",
			level:        6,
			maxLevel:     14,
			rarity:       "Epic",
			elixir:       7,
			role:         new(RoleWinCondition),
			wantScoreMin: 0.69,
			wantScoreMax: 0.70,
		},
		{
			name:         "Champion with role",
			level:        12,
			maxLevel:     14,
			rarity:       "Champion",
			elixir:       3,
			role:         new(RoleWinCondition),
			wantScoreMin: 1.43,
			wantScoreMax: 1.44,
		},
		{
			name:         "Mid level rare optimal elixir",
			level:        8,
			maxLevel:     14,
			rarity:       "Rare",
			elixir:       3,
			role:         new(RoleSupport),
			wantScoreMin: 0.92,
			wantScoreMax: 0.94,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := ScoreCard(tt.level, tt.maxLevel, tt.rarity, tt.elixir, tt.role)

			if score < tt.wantScoreMin || score > tt.wantScoreMax {
				t.Errorf("ScoreCard() = %v, want between %v and %v",
					score, tt.wantScoreMin, tt.wantScoreMax)
			}
		})
	}
}

// TestScoreCardZeroMaxLevel tests edge case of zero max level
func TestScoreCardZeroMaxLevel(t *testing.T) {
	score := ScoreCard(5, 0, "Common", 3, nil)
	// Should not panic, level ratio should be 0
	if score < 0 || score > 0.2 {
		t.Errorf("ScoreCard with zero maxLevel = %v, want small non-negative value", score)
	}
}

// TestScoreCardUnknownRarity tests fallback for unknown rarity
func TestScoreCardUnknownRarity(t *testing.T) {
	score := ScoreCard(10, 14, "Unknown", 3, nil)
	// Should default to Common (1.0 boost)
	if score < 0.9 || score > 1.1 {
		t.Errorf("ScoreCard with unknown rarity = %v, want around 1.0", score)
	}
}

// TestScoreCardCandidate tests scoring a CardCandidate
func TestScoreCardCandidate(t *testing.T) {
	candidate := CardCandidate{
		Name:     "Fire Spirit",
		Level:    10,
		MaxLevel: 14,
		Rarity:   "Common",
		Elixir:   1,
		Role:     new(RoleCycle),
	}

	score := ScoreCardCandidate(&candidate)

	// Verify score was set on candidate
	if candidate.Score != score {
		t.Errorf("ScoreCardCandidate did not update candidate.Score")
	}

	// Verify score is reasonable
	if score < 0.8 || score > 1.2 {
		t.Errorf("ScoreCardCandidate score = %v, want between 0.8 and 1.2", score)
	}
}

// TestSortByScore tests sorting candidates by score
func TestSortByScore(t *testing.T) {
	candidates := []CardCandidate{
		{Name: "Low", Score: 0.5},
		{Name: "High", Score: 1.5},
		{Name: "Mid", Score: 1.0},
	}

	SortByScore(candidates)

	// Should be sorted descending
	if candidates[0].Name != "High" {
		t.Errorf("SortByScore first = %v, want High", candidates[0].Name)
	}
	if candidates[2].Name != "Low" {
		t.Errorf("SortByScore last = %v, want Low", candidates[2].Name)
	}
}

// TestFilterByMinScore tests filtering by minimum score
func TestFilterByMinScore(t *testing.T) {
	candidates := []CardCandidate{
		{Name: "A", Score: 0.5},
		{Name: "B", Score: 1.0},
		{Name: "C", Score: 1.5},
		{Name: "D", Score: 0.8},
	}

	filtered := FilterByMinScore(candidates, 0.9)

	if len(filtered) != 2 {
		t.Errorf("FilterByMinScore returned %v cards, want 2", len(filtered))
	}

	// Check that all filtered cards have score >= 0.9
	for _, card := range filtered {
		if card.Score < 0.9 {
			t.Errorf("FilterByMinScore included card with score %v < 0.9", card.Score)
		}
	}
}

// TestGetTopN tests getting top N candidates
func TestGetTopN(t *testing.T) {
	candidates := []CardCandidate{
		{Name: "A", Score: 0.5},
		{Name: "B", Score: 1.5},
		{Name: "C", Score: 1.0},
		{Name: "D", Score: 0.8},
	}

	top2 := GetTopN(candidates, 2)

	if len(top2) != 2 {
		t.Errorf("GetTopN(2) returned %v cards, want 2", len(top2))
	}

	// Should be sorted and contain top 2 scores
	if top2[0].Name != "B" {
		t.Errorf("GetTopN first = %v, want B (score 1.5)", top2[0].Name)
	}
	if top2[1].Name != "C" {
		t.Errorf("GetTopN second = %v, want C (score 1.0)", top2[1].Name)
	}
}

// TestFilterByRole tests filtering by card role
func TestFilterByRole(t *testing.T) {
	winCon := RoleWinCondition
	cycle := RoleCycle

	candidates := []CardCandidate{
		{Name: "Hog", Role: &winCon},
		{Name: "Skeletons", Role: &cycle},
		{Name: "Giant", Role: &winCon},
		{Name: "No Role", Role: nil},
	}

	filtered := FilterByRole(candidates, RoleWinCondition)

	if len(filtered) != 2 {
		t.Errorf("FilterByRole(WinCondition) returned %v cards, want 2", len(filtered))
	}

	for _, card := range filtered {
		if card.Role == nil || *card.Role != RoleWinCondition {
			t.Errorf("FilterByRole included card %v without WinCondition role", card.Name)
		}
	}
}

// TestFilterByElixirRange tests filtering by elixir cost range
func TestFilterByElixirRange(t *testing.T) {
	candidates := []CardCandidate{
		{Name: "Skeletons", Elixir: 1},
		{Name: "Zap", Elixir: 2},
		{Name: "Knight", Elixir: 3},
		{Name: "Fireball", Elixir: 4},
		{Name: "Golem", Elixir: 8},
	}

	// Filter for 2-4 elixir range
	filtered := FilterByElixirRange(candidates, 2, 4)

	if len(filtered) != 3 {
		t.Errorf("FilterByElixirRange(2, 4) returned %v cards, want 3", len(filtered))
	}

	for _, card := range filtered {
		if card.Elixir < 2 || card.Elixir > 4 {
			t.Errorf("FilterByElixirRange included card %v with elixir %v outside range 2-4",
				card.Name, card.Elixir)
		}
	}
}

// TestCalculateAvgElixir tests average elixir calculation
func TestCalculateAvgElixir(t *testing.T) {
	candidates := []CardCandidate{
		{Elixir: 1},
		{Elixir: 2},
		{Elixir: 3},
		{Elixir: 4},
	}

	avg := CalculateAvgElixir(candidates)
	expected := 2.5

	if math.Abs(avg-expected) > 0.01 {
		t.Errorf("CalculateAvgElixir = %v, want %v", avg, expected)
	}
}

// TestCalculateAvgElixirEmpty tests average with empty slice
func TestCalculateAvgElixirEmpty(t *testing.T) {
	avg := CalculateAvgElixir([]CardCandidate{})
	if avg != 0.0 {
		t.Errorf("CalculateAvgElixir on empty slice = %v, want 0.0", avg)
	}
}

// TestExcludeCards tests excluding specific cards by name
func TestExcludeCards(t *testing.T) {
	candidates := []CardCandidate{
		{Name: "Hog Rider"},
		{Name: "Royal Giant"},
		{Name: "Goblin Barrel"},
		{Name: "Cannon"},
	}

	excluded := ExcludeCards(candidates, []string{"Hog Rider", "Cannon"})

	if len(excluded) != 2 {
		t.Errorf("ExcludeCards returned %v cards, want 2", len(excluded))
	}

	for _, card := range excluded {
		if card.Name == "Hog Rider" || card.Name == "Cannon" {
			t.Errorf("ExcludeCards did not exclude %v", card.Name)
		}
	}
}

// BenchmarkScoreCard benchmarks the scoring algorithm
func BenchmarkScoreCard(b *testing.B) {
	role := new(RoleWinCondition)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ScoreCard(10, 14, "Epic", 4, role)
	}
}

// Test ScoreCardWithCombat functionality
func TestScoreCardWithCombat(t *testing.T) {
	// Create role variables
	winConRole := RoleWinCondition
	supportRole := RoleSupport

	// Test with default combat weight
	t.Run("Default combat weight", func(t *testing.T) {
		stats := &clashroyale.CombatStats{
			Hitpoints:       4000, // Higher HP to favor win condition
			DamagePerSecond: 150,  // Higher DPS
			Range:           6,
			Targets:         "Air & Ground",
			Speed:           "Fast",
		}

		score := ScoreCardWithCombat(11, 14, "Legendary", 5, &winConRole, stats)

		// Score should be reasonable for good combat stats
		if score <= 0 {
			t.Errorf("Combat score should be positive, got %v", score)
		}

		// Combat-enhanced scoring should produce a meaningful score
		if score < 0.5 || score > 2.0 {
			t.Errorf("Combat score %v seems unreasonable", score)
		}
	})

	// Test with no combat stats
	t.Run("No combat stats", func(t *testing.T) {
		score := ScoreCardWithCombat(10, 14, "Common", 3, &supportRole, nil)
		baseScore := ScoreCard(10, 14, "Common", 3, &supportRole)

		// Should equal base score when no combat stats available
		if score != baseScore {
			t.Errorf("Score with no stats %v should equal base score %v", score, baseScore)
		}
	})

	// Test with combat weight disabled
	t.Run("Combat weight disabled", func(t *testing.T) {
		os.Setenv("COMBAT_STATS_WEIGHT", "0")
		defer os.Unsetenv("COMBAT_STATS_WEIGHT")

		stats := &clashroyale.CombatStats{
			Hitpoints: 2000,
			Damage:    500,
		}

		score := ScoreCardWithCombat(10, 14, "Rare", 5, &winConRole, stats)
		baseScore := ScoreCard(10, 14, "Rare", 5, &winConRole)

		// Should equal base score when combat weight is 0
		if score != baseScore {
			t.Errorf("Score with disabled weight %v should equal base score %v", score, baseScore)
		}
	})
}

func TestScoreCardCandidateWithCombat(t *testing.T) {
	supportRole := RoleSupport
	candidate := &CardCandidate{
		Name:     "Test Card",
		Level:    10,
		MaxLevel: 14,
		Rarity:   "Rare",
		Elixir:   4,
		Role:     &supportRole,
		Stats: &clashroyale.CombatStats{
			Hitpoints:       1500,
			DamagePerSecond: 80,
			Range:           6,
			Targets:         "Air & Ground",
		},
	}

	originalScore := candidate.Score
	score := ScoreCardCandidateWithCombat(candidate)

	// Verify score was updated on candidate
	if candidate.Score != score {
		t.Error("ScoreCardCandidateWithCombat did not update candidate.Score")
	}

	// Verify score is reasonable
	if score <= 0 {
		t.Errorf("Combat-enhanced score should be positive, got %v", score)
	}

	// Score should be higher than original (which was 0) for good combat stats
	if score <= originalScore {
		t.Errorf("Combat-enhanced score %v should be > original score %v for good combat stats", score, originalScore)
	}
}

func TestScoreAllCandidatesWithCombat(t *testing.T) {
	winConRole := RoleWinCondition
	cycleRole := RoleCycle
	supportRole := RoleSupport

	candidates := []CardCandidate{
		{
			Name:     "Strong Card",
			Level:    11,
			MaxLevel: 14,
			Rarity:   "Epic",
			Elixir:   5,
			Role:     &winConRole,
			Stats: &clashroyale.CombatStats{
				Hitpoints:       3000,
				DamagePerSecond: 120,
				Range:           4,
				Targets:         "Ground",
				Speed:           "Slow",
			},
		},
		{
			Name:     "Weak Card",
			Level:    5,
			MaxLevel: 14,
			Rarity:   "Common",
			Elixir:   2,
			Role:     &cycleRole,
			Stats: &clashroyale.CombatStats{
				Hitpoints: 300,
				Damage:    30,
				Speed:     "Fast",
			},
		},
		{
			Name:     "No Stats Card",
			Level:    8,
			MaxLevel: 14,
			Rarity:   "Rare",
			Elixir:   3,
			Role:     &supportRole,
			Stats:    nil,
		},
	}

	// Store original scores
	originalScores := make([]float64, len(candidates))
	for i, candidate := range candidates {
		originalScores[i] = candidate.Score
	}

	// Score with combat
	ScoreAllCandidatesWithCombat(candidates)

	// Verify all candidates have scores
	for i, candidate := range candidates {
		if candidate.Score <= 0 {
			t.Errorf("Candidate %v has invalid score after combat scoring: %v", candidate.Name, candidate.Score)
		}

		// Verify scores were updated
		if candidate.Score == originalScores[i] && candidate.Stats != nil {
			t.Logf("Note: %v score unchanged after combat scoring: %v", candidate.Name, candidate.Score)
		}
	}
}

func TestGetCombatWeight(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		expectedValue float64
	}{
		{
			name:          "Default weight",
			envValue:      "",
			expectedValue: 0.25,
		},
		{
			name:          "Custom weight",
			envValue:      "0.5",
			expectedValue: 0.5,
		},
		{
			name:          "Zero weight",
			envValue:      "0",
			expectedValue: 0,
		},
		{
			name:          "Weight above 1 (clamped)",
			envValue:      "1.5",
			expectedValue: 1,
		},
		{
			name:          "Negative weight (clamped)",
			envValue:      "-0.5",
			expectedValue: 0,
		},
		{
			name:          "Invalid weight (uses default)",
			envValue:      "invalid",
			expectedValue: 0.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("COMBAT_STATS_WEIGHT", tt.envValue)
				defer os.Unsetenv("COMBAT_STATS_WEIGHT")
			} else {
				os.Unsetenv("COMBAT_STATS_WEIGHT")
			}

			weight := getCombatWeight()
			if weight != tt.expectedValue {
				t.Errorf("getCombatWeight() = %v, want %v", weight, tt.expectedValue)
			}
		})
	}
}

// BenchmarkSortByScore benchmarks sorting candidates
func BenchmarkSortByScore(b *testing.B) {
	candidates := make([]CardCandidate, 100)
	for i := range candidates {
		candidates[i] = CardCandidate{
			Name:  "Card",
			Score: float64(i) * 0.01,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SortByScore(candidates)
	}
}

// TestCalculateEvolutionLevelBonus tests the evolution level bonus calculation
func TestCalculateEvolutionLevelBonus(t *testing.T) {
	tests := []struct {
		name              string
		evolutionLevel    int
		maxEvolutionLevel int
		wantBonus         float64
	}{
		{
			name:              "No evolution capability",
			evolutionLevel:    0,
			maxEvolutionLevel: 0,
			wantBonus:         0.0,
		},
		{
			name:              "Has evolution but not evolved",
			evolutionLevel:    0,
			maxEvolutionLevel: 3,
			wantBonus:         0.0,
		},
		{
			name:              "Half evolved (1/2)",
			evolutionLevel:    1,
			maxEvolutionLevel: 2,
			wantBonus:         0.075, // 0.15 * (1/2) = 0.075
		},
		{
			name:              "Fully evolved (2/2)",
			evolutionLevel:    2,
			maxEvolutionLevel: 2,
			wantBonus:         0.15, // 0.15 * (2/2) = 0.15
		},
		{
			name:              "Partially evolved (1/3)",
			evolutionLevel:    1,
			maxEvolutionLevel: 3,
			wantBonus:         0.05, // 0.15 * (1/3) = 0.05
		},
		{
			name:              "Two-thirds evolved (2/3)",
			evolutionLevel:    2,
			maxEvolutionLevel: 3,
			wantBonus:         0.10, // 0.15 * (2/3) = 0.10
		},
		{
			name:              "Fully evolved (3/3)",
			evolutionLevel:    3,
			maxEvolutionLevel: 3,
			wantBonus:         0.15, // 0.15 * (3/3) = 0.15
		},
		{
			name:              "Single evolution level fully evolved (1/1)",
			evolutionLevel:    1,
			maxEvolutionLevel: 1,
			wantBonus:         0.15, // 0.15 * (1/1) = 0.15
		},
		{
			name:              "Negative evolution level treated as zero",
			evolutionLevel:    -1,
			maxEvolutionLevel: 3,
			wantBonus:         0.0,
		},
		{
			name:              "Negative max evolution level treated as zero",
			evolutionLevel:    1,
			maxEvolutionLevel: -1,
			wantBonus:         0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonus := calculateEvolutionLevelBonus(tt.evolutionLevel, tt.maxEvolutionLevel)

			if math.Abs(bonus-tt.wantBonus) > 0.001 {
				t.Errorf("calculateEvolutionLevelBonus(%d, %d) = %v, want %v",
					tt.evolutionLevel, tt.maxEvolutionLevel, bonus, tt.wantBonus)
			}
		})
	}
}

// TestScoreCardWithEvolution tests the evolution-aware scoring function
func TestScoreCardWithEvolution(t *testing.T) {
	tests := []struct {
		name              string
		level             int
		maxLevel          int
		rarity            string
		elixir            int
		role              *CardRole
		evolutionLevel    int
		maxEvolutionLevel int
		wantScoreMin      float64
		wantScoreMax      float64
	}{
		{
			name:              "Card without evolution (same as ScoreCard)",
			level:             11,
			maxLevel:          14,
			rarity:            "Common",
			elixir:            3,
			role:              new(RoleCycle),
			evolutionLevel:    0,
			maxEvolutionLevel: 0,
			wantScoreMin:      1.14, // Should match TestScoreCard result
			wantScoreMax:      1.15,
		},
		{
			name:              "Card with evolution capability but not evolved",
			level:             11,
			maxLevel:          14,
			rarity:            "Common",
			elixir:            3,
			role:              new(RoleCycle),
			evolutionLevel:    0,
			maxEvolutionLevel: 3,
			wantScoreMin:      1.14, // Same as without evolution
			wantScoreMax:      1.15,
		},
		{
			name:              "Card fully evolved",
			level:             11,
			maxLevel:          14,
			rarity:            "Common",
			elixir:            3,
			role:              new(RoleCycle),
			evolutionLevel:    3,
			maxEvolutionLevel: 3,
			wantScoreMin:      1.29, // 1.14 + 0.15 = 1.29
			wantScoreMax:      1.30,
		},
		{
			name:              "Card partially evolved (1/3)",
			level:             11,
			maxLevel:          14,
			rarity:            "Common",
			elixir:            3,
			role:              new(RoleCycle),
			evolutionLevel:    1,
			maxEvolutionLevel: 3,
			wantScoreMin:      1.19, // 1.14 + 0.05 = 1.19
			wantScoreMax:      1.20,
		},
		{
			name:              "Max level legendary fully evolved",
			level:             14,
			maxLevel:          14,
			rarity:            "Legendary",
			elixir:            4,
			role:              nil,
			evolutionLevel:    2,
			maxEvolutionLevel: 2,
			wantScoreMin:      1.66, // 1.51 + 0.15 = 1.66
			wantScoreMax:      1.67,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := ScoreCardWithEvolution(
				tt.level, tt.maxLevel, tt.rarity, tt.elixir, tt.role,
				tt.evolutionLevel, tt.maxEvolutionLevel,
			)

			if score < tt.wantScoreMin || score > tt.wantScoreMax {
				t.Errorf("ScoreCardWithEvolution() = %v, want between %v and %v",
					score, tt.wantScoreMin, tt.wantScoreMax)
			}
		})
	}
}

// TestScoreCardCandidateWithEvolution tests that ScoreCardCandidate uses evolution data
func TestScoreCardCandidateWithEvolution(t *testing.T) {
	// Card without evolution
	candidateNoEvo := CardCandidate{
		Name:              "Knight",
		Level:             11,
		MaxLevel:          14,
		Rarity:            "Common",
		Elixir:            3,
		Role:              new(RoleCycle),
		EvolutionLevel:    0,
		MaxEvolutionLevel: 0,
	}

	// Card with full evolution
	candidateFullEvo := CardCandidate{
		Name:              "Knight",
		Level:             11,
		MaxLevel:          14,
		Rarity:            "Common",
		Elixir:            3,
		Role:              new(RoleCycle),
		EvolutionLevel:    3,
		MaxEvolutionLevel: 3,
	}

	scoreNoEvo := ScoreCardCandidate(&candidateNoEvo)
	scoreFullEvo := ScoreCardCandidate(&candidateFullEvo)

	// Verify scores were updated on candidates
	if candidateNoEvo.Score != scoreNoEvo {
		t.Error("ScoreCardCandidate did not update candidateNoEvo.Score")
	}
	if candidateFullEvo.Score != scoreFullEvo {
		t.Error("ScoreCardCandidate did not update candidateFullEvo.Score")
	}

	// Full evolution should score higher than no evolution
	if scoreFullEvo <= scoreNoEvo {
		t.Errorf("Fully evolved card score (%v) should be higher than non-evolved (%v)",
			scoreFullEvo, scoreNoEvo)
	}

	// The difference should be approximately the evolution bonus weight (0.15)
	expectedDiff := config.EvolutionBonusWeight
	actualDiff := scoreFullEvo - scoreNoEvo
	if math.Abs(actualDiff-expectedDiff) > 0.01 {
		t.Errorf("Score difference (%v) should be approximately %v", actualDiff, expectedDiff)
	}
}

// TestScoreCardWithCombatAndEvolution tests the combined combat + evolution scoring
func TestScoreCardWithCombatAndEvolution(t *testing.T) {
	winConRole := RoleWinCondition

	stats := &clashroyale.CombatStats{
		Hitpoints:       4000,
		DamagePerSecond: 150,
		Range:           6,
		Targets:         "Air & Ground",
		Speed:           "Fast",
	}

	// Test without evolution
	scoreNoEvo := ScoreCardWithCombatAndEvolution(
		11, 14, "Legendary", 5, &winConRole, stats,
		0, 0,
	)

	// Test with full evolution
	scoreFullEvo := ScoreCardWithCombatAndEvolution(
		11, 14, "Legendary", 5, &winConRole, stats,
		3, 3,
	)

	// Full evolution should score higher
	if scoreFullEvo <= scoreNoEvo {
		t.Errorf("Evolved score (%v) should be higher than non-evolved (%v)",
			scoreFullEvo, scoreNoEvo)
	}

	// Verify scores are reasonable
	if scoreNoEvo <= 0 || scoreNoEvo > 2.0 {
		t.Errorf("Score without evolution (%v) seems unreasonable", scoreNoEvo)
	}
	if scoreFullEvo <= 0 || scoreFullEvo > 2.5 {
		t.Errorf("Score with evolution (%v) seems unreasonable", scoreFullEvo)
	}
}

// TestScoreCardCandidateWithCombatIncludesEvolution tests that combat scoring includes evolution
func TestScoreCardCandidateWithCombatIncludesEvolution(t *testing.T) {
	supportRole := RoleSupport

	// Card without evolution
	candidateNoEvo := &CardCandidate{
		Name:              "Test Card",
		Level:             10,
		MaxLevel:          14,
		Rarity:            "Rare",
		Elixir:            4,
		Role:              &supportRole,
		EvolutionLevel:    0,
		MaxEvolutionLevel: 0,
		Stats: &clashroyale.CombatStats{
			Hitpoints:       1500,
			DamagePerSecond: 80,
			Range:           6,
			Targets:         "Air & Ground",
		},
	}

	// Card with full evolution
	candidateFullEvo := &CardCandidate{
		Name:              "Test Card",
		Level:             10,
		MaxLevel:          14,
		Rarity:            "Rare",
		Elixir:            4,
		Role:              &supportRole,
		EvolutionLevel:    2,
		MaxEvolutionLevel: 2,
		Stats: &clashroyale.CombatStats{
			Hitpoints:       1500,
			DamagePerSecond: 80,
			Range:           6,
			Targets:         "Air & Ground",
		},
	}

	scoreNoEvo := ScoreCardCandidateWithCombat(candidateNoEvo)
	scoreFullEvo := ScoreCardCandidateWithCombat(candidateFullEvo)

	// Verify scores were updated
	if candidateNoEvo.Score != scoreNoEvo {
		t.Error("ScoreCardCandidateWithCombat did not update candidate.Score")
	}
	if candidateFullEvo.Score != scoreFullEvo {
		t.Error("ScoreCardCandidateWithCombat did not update candidate.Score")
	}

	// Full evolution should score higher
	if scoreFullEvo <= scoreNoEvo {
		t.Errorf("Evolved card score (%v) should be higher than non-evolved (%v)",
			scoreFullEvo, scoreNoEvo)
	}
}

// TestEvolutionBonusClampedToMax tests that evolution ratio is clamped to 1.0
func TestEvolutionBonusClampedToMax(t *testing.T) {
	// Evolution level exceeds max (edge case that shouldn't happen in practice)
	bonus := calculateEvolutionLevelBonus(5, 3)

	// Should be clamped to max bonus
	expectedMax := config.EvolutionBonusWeight
	if bonus > expectedMax+0.001 {
		t.Errorf("Evolution bonus (%v) should be clamped to max (%v)", bonus, expectedMax)
	}
}

// BenchmarkScoreCardWithEvolution benchmarks the evolution-aware scoring
func BenchmarkScoreCardWithEvolution(b *testing.B) {
	role := new(RoleWinCondition)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ScoreCardWithEvolution(10, 14, "Epic", 4, role, 2, 3)
	}
}

// BenchmarkCalculateEvolutionLevelBonus benchmarks the evolution bonus calculation
func BenchmarkCalculateEvolutionLevelBonus(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateEvolutionLevelBonus(2, 3)
	}
}

// TestRoleToString tests the roleToString helper function
func TestRoleToString(t *testing.T) {
	tests := []struct {
		name     string
		role     *CardRole
		expected string
	}{
		{"nil role", nil, ""},
		{"Win Condition", new(RoleWinCondition), "wincondition"},
		{"Building", new(RoleBuilding), "building"},
		{"Support", new(RoleSupport), "support"},
		{"Spell Big", new(RoleSpellBig), "spell"},
		{"Spell Small", new(RoleSpellSmall), "spell"}, // Both map to "spell"
		{"Cycle", new(RoleCycle), "cycle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roleToString(tt.role)
			if result != tt.expected {
				t.Errorf("roleToString(%v) = %v, want %v", tt.role, result, tt.expected)
			}
		})
	}
}

// TestGetTopN_EdgeCases tests edge cases for GetTopN
func TestGetTopN_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		candidates  []CardCandidate
		n           int
		expectedLen int
	}{
		{
			name:        "Empty slice",
			candidates:  []CardCandidate{},
			n:           5,
			expectedLen: 0,
		},
		{
			name: "N larger than slice",
			candidates: []CardCandidate{
				{Name: "A", Score: 1.0},
				{Name: "B", Score: 0.5},
			},
			n:           10,
			expectedLen: 2,
		},
		{
			name: "N is zero",
			candidates: []CardCandidate{
				{Name: "A", Score: 1.0},
			},
			n:           0,
			expectedLen: 0,
		},
		{
			name: "All same scores",
			candidates: []CardCandidate{
				{Name: "A", Score: 1.0},
				{Name: "B", Score: 1.0},
				{Name: "C", Score: 1.0},
			},
			n:           2,
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTopN(tt.candidates, tt.n)
			if len(result) != tt.expectedLen {
				t.Errorf("GetTopN() returned %d items, want %d", len(result), tt.expectedLen)
			}
		})
	}
}

// TestGetTopN_AllScoresNegative tests that negative scores are handled correctly
func TestGetTopN_AllScoresNegative(t *testing.T) {
	candidates := []CardCandidate{
		{Name: "A", Score: -1.5},
		{Name: "B", Score: -0.5},
		{Name: "C", Score: -2.0},
	}

	top2 := GetTopN(candidates, 2)

	if len(top2) != 2 {
		t.Errorf("GetTopN with negative scores returned %d items, want 2", len(top2))
	}

	// Should still sort correctly (least negative first)
	if top2[0].Score != -0.5 {
		t.Errorf("Expected highest score -0.5, got %v", top2[0].Score)
	}
}
