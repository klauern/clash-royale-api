package deck

import (
	"testing"
)

// TestNewEvolutionRecommender tests creating a new recommender.
func TestNewEvolutionRecommender(t *testing.T) {
	shards := map[string]int{"Knight": 5, "Archers": 10}
	unlocked := []string{"Musketeer"}

	recommender := NewEvolutionRecommender(shards, unlocked)

	if recommender == nil {
		t.Fatal("NewEvolutionRecommender returned nil")
	}
	if recommender.shardsPerEvolution != 10 {
		t.Errorf("shardsPerEvolution = %d, want 10", recommender.shardsPerEvolution)
	}
	if !recommender.unlockedEvolutions["Musketeer"] {
		t.Error("Musketeer should be in unlocked evolutions")
	}
}

// TestEvolutionRecommender_Recommend tests generating recommendations.
func TestEvolutionRecommender_Recommend(t *testing.T) {
	shardSource := map[string]int{
		"Knight":    5,  // Halfway to first evolution
		"Archers":   10, // Ready for evolution
		"Musketeer": 0,  // No shards
		"Hog Rider": 2,  // Low shard count
	}
	unlocked := []string{"Valkyrie"} // Valkyrie already unlocked

	recommender := NewEvolutionRecommender(shardSource, unlocked)

	supportRole := RoleSupport
	winConRole := RoleWinCondition

	candidates := []CardCandidate{
		{
			Name:              "Knight",
			Level:             14,
			MaxLevel:          14,
			Rarity:            "Rare",
			Elixir:            3,
			Role:              &supportRole,
			EvolutionLevel:    0,
			MaxEvolutionLevel: 3,
		},
		{
			Name:              "Archers",
			Level:             11,
			MaxLevel:          14,
			Rarity:            "Common",
			Elixir:            3,
			Role:              &supportRole,
			EvolutionLevel:    0,
			MaxEvolutionLevel: 1,
		},
		{
			Name:              "Musketeer",
			Level:             14,
			MaxLevel:          14,
			Rarity:            "Rare",
			Elixir:            4,
			Role:              &supportRole,
			EvolutionLevel:    0,
			MaxEvolutionLevel: 3,
		},
		{
			Name:              "Hog Rider",
			Level:             14,
			MaxLevel:          14,
			Rarity:            "Legendary",
			Elixir:            4,
			Role:              &winConRole,
			EvolutionLevel:    0,
			MaxEvolutionLevel: 3,
		},
		{
			Name:              "Valkyrie",
			Level:             14,
			MaxLevel:          14,
			Rarity:            "Epic",
			Elixir:            4,
			Role:              &supportRole,
			EvolutionLevel:    0,
			MaxEvolutionLevel: 3,
		},
		{
			Name:     "Giant",
			Level:    14,
			MaxLevel: 14,
			Rarity:   "Rare",
			Elixir:   5,
			Role:     &winConRole,
			// No evolution capability
			EvolutionLevel:    0,
			MaxEvolutionLevel: 0,
		},
	}

	recommendations := recommender.Recommend(candidates, 10)

	// Should have 4 recommendations (excludes Valkyrie/unlocked and Giant/no evo)
	if len(recommendations) != 4 {
		t.Errorf("Recommend() returned %d items, want 4", len(recommendations))
	}

	// Check that Valkyrie is not in recommendations (already unlocked)
	for _, rec := range recommendations {
		if rec.CardName == "Valkyrie" {
			t.Error("Valkyrie should not be recommended (already unlocked)")
		}
		if rec.CardName == "Giant" {
			t.Error("Giant should not be recommended (no evolution capability)")
		}
	}

	// Recommendations should be sorted by score (descending)
	for i := 1; i < len(recommendations); i++ {
		if recommendations[i-1].RecommendationScore < recommendations[i].RecommendationScore {
			t.Errorf("Recommendations not sorted by score: [%d].Score=%f < [%d].Score=%f",
				i-1, recommendations[i-1].RecommendationScore,
				i, recommendations[i].RecommendationScore)
		}
	}
}

// TestEvolutionRecommender_TopN tests limiting recommendations count.
func TestEvolutionRecommender_TopN(t *testing.T) {
	recommender := NewEvolutionRecommender(map[string]int{"Knight": 10}, nil)

	candidates := make([]CardCandidate, 5)
	supportRole := RoleSupport
	for i := range candidates {
		candidates[i] = CardCandidate{
			Name:              string(rune('A' + i)),
			Level:             14,
			MaxLevel:          14,
			Rarity:            "Common",
			Elixir:            3,
			Role:              &supportRole,
			EvolutionLevel:    0,
			MaxEvolutionLevel: 1,
		}
	}

	// Request top 3
	recommendations := recommender.Recommend(candidates, 3)
	if len(recommendations) != 3 {
		t.Errorf("TopN=3 returned %d items, want 3", len(recommendations))
	}

	// Request all (topN=0 or > len)
	recommendations = recommender.Recommend(candidates, 0)
	if len(recommendations) != 5 {
		t.Errorf("TopN=0 returned %d items, want 5", len(recommendations))
	}
}

// TestFormatRecommendations tests formatting recommendations.
func TestFormatRecommendations(t *testing.T) {
	recs := []EvolutionRecommendation{
		{
			CardName:            "Knight",
			CurrentShards:       5,
			ShardsNeeded:        10,
			CompletionPercent:   50.0,
			CardLevel:           14,
			MaxLevel:            14,
			LevelRatio:          1.0,
			Role:                "Support",
			EvolutionMaxLevel:   3,
			RecommendationScore: 75.0,
			Reasons:             []string{"High level", "Halfway to shards"},
		},
	}

	output := FormatRecommendations(recs, true)
	if output == "" {
		t.Error("FormatRecommendations returned empty string")
	}

	// Check that key info is included
	if !contains(output, "Knight") {
		t.Error("Output should contain card name")
	}
	if !contains(output, "5/10") {
		t.Error("Output should contain shard progress")
	}
	if !contains(output, "Support") {
		t.Error("Output should contain role")
	}
	if !contains(output, "75.0") {
		t.Error("Output should contain score")
	}
}

// TestFormatRecommendations_NoRecs tests formatting with no recommendations.
func TestFormatRecommendations_NoRecs(t *testing.T) {
	output := FormatRecommendations([]EvolutionRecommendation{}, true)
	if !contains(output, "No evolution recommendations") {
		t.Errorf("Expected 'no recommendations' message, got: %s", output)
	}
}

// TestRolePriorityScore tests role-based scoring.
func TestRolePriorityScore(t *testing.T) {
	recommender := NewEvolutionRecommender(nil, nil)

	tests := []struct {
		role CardRole
		min  float64
		max  float64
	}{
		{RoleWinCondition, 19, 21},
		{RoleSupport, 14, 16},
		{RoleSpellBig, 11, 13},
		{RoleSpellSmall, 7, 9},
		{RoleBuilding, 9, 11},
		{RoleCycle, 4, 6},
	}

	for _, tt := range tests {
		t.Run(tt.role.String(), func(t *testing.T) {
			score := recommender.rolePriorityScore(tt.role)
			if score < tt.min || score > tt.max {
				t.Errorf("rolePriorityScore(%v) = %v, want between %v and %v",
					tt.role, score, tt.min, tt.max)
			}
		})
	}
}

// TestEvaluateCandidate tests candidate evaluation logic.
func TestEvaluateCandidate(t *testing.T) {
	shardSource := map[string]int{"Knight": 5}
	recommender := NewEvolutionRecommender(shardSource, nil)

	supportRole := RoleSupport
	candidate := CardCandidate{
		Name:              "Knight",
		Level:             14,
		MaxLevel:          14,
		Rarity:            "Rare",
		Elixir:            3,
		Role:              &supportRole,
		EvolutionLevel:    0,
		MaxEvolutionLevel: 3,
	}

	rec := recommender.evaluateCandidate(candidate)

	if rec == nil {
		t.Fatal("evaluateCandidate returned nil")
	}

	if rec.CardName != "Knight" {
		t.Errorf("CardName = %v, want Knight", rec.CardName)
	}
	if rec.CurrentShards != 5 {
		t.Errorf("CurrentShards = %v, want 5", rec.CurrentShards)
	}
	if rec.ShardsNeeded != 10 {
		t.Errorf("ShardsNeeded = %v, want 10", rec.ShardsNeeded)
	}
	if rec.CardLevel != 14 {
		t.Errorf("CardLevel = %v, want 14", rec.CardLevel)
	}
	if rec.MaxLevel != 14 {
		t.Errorf("MaxLevel = %v, want 14", rec.MaxLevel)
	}
	if rec.EvolutionMaxLevel != 3 {
		t.Errorf("EvolutionMaxLevel = %v, want 3", rec.EvolutionMaxLevel)
	}

	// Score should be positive
	if rec.RecommendationScore <= 0 {
		t.Errorf("RecommendationScore = %v, want > 0", rec.RecommendationScore)
	}

	// Should have some reasons
	if len(rec.Reasons) == 0 {
		t.Error("Expected some reasons to be generated")
	}
}

// TestSetShardsPerEvolution tests custom shard count.
func TestSetShardsPerEvolution(t *testing.T) {
	recommender := NewEvolutionRecommender(nil, nil)
	recommender.SetShardsPerEvolution(5)

	if recommender.shardsPerEvolution != 5 {
		t.Errorf("shardsPerEvolution = %d, want 5", recommender.shardsPerEvolution)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr ||
		(len(s) > len(substr) && findSubstring(s[1:], substr)))
}
