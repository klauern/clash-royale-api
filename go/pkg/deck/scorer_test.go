package deck

import (
	"math"
	"testing"
)

// TestScoreCard validates the scoring algorithm matches expected Python behavior
func TestScoreCard(t *testing.T) {
	tests := []struct {
		name       string
		level      int
		maxLevel   int
		rarity     string
		elixir     int
		role       *CardRole
		wantScoreMin float64
		wantScoreMax float64
	}{
		{
			name:         "High level common with role",
			level:        11,
			maxLevel:     14,
			rarity:       "Common",
			elixir:       3,
			role:         rolePtr(RoleCycle),
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
			role:         rolePtr(RoleWinCondition),
			wantScoreMin: 0.69,
			wantScoreMax: 0.70,
		},
		{
			name:         "Champion with role",
			level:        12,
			maxLevel:     14,
			rarity:       "Champion",
			elixir:       3,
			role:         rolePtr(RoleWinCondition),
			wantScoreMin: 1.43,
			wantScoreMax: 1.44,
		},
		{
			name:         "Mid level rare optimal elixir",
			level:        8,
			maxLevel:     14,
			rarity:       "Rare",
			elixir:       3,
			role:         rolePtr(RoleSupport),
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
		Role:     rolePtr(RoleCycle),
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
	role := rolePtr(RoleWinCondition)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ScoreCard(10, 14, "Epic", 4, role)
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
