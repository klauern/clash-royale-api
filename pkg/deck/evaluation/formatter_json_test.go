package evaluation

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestFormatJSON(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()
	deckCards := []deck.CardCandidate{
		makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
		makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
		makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
		makeCard("Zap", deck.RoleSpellSmall, 11, 11, "Common", 2),
		makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
		makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
		makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
		makeCard("Ice Golem", deck.RoleCycle, 11, 11, "Rare", 2),
	}

	result := Evaluate(deckCards, synergyDB, nil)

	jsonOutput, err := FormatJSON(&result)
	if err != nil {
		t.Fatalf("FormatJSON() returned error: %v", err)
	}

	if jsonOutput == "" {
		t.Errorf("FormatJSON() returned empty string")
	}

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(jsonOutput), &parsed); err != nil {
		t.Errorf("FormatJSON() returned invalid JSON: %v", err)
	}

	// Check for expected top-level keys
	expectedKeys := []string{"version", "timestamp", "evaluation"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("FormatJSON() missing top-level key %q", key)
		}
	}

	// Check evaluation structure
	eval, ok := parsed["evaluation"].(map[string]any)
	if !ok {
		t.Fatal("FormatJSON() evaluation is not a map")
	}

	expectedEvalKeys := []string{"deck", "overall", "archetype", "category_scores"}
	for _, key := range expectedEvalKeys {
		if _, ok := eval[key]; !ok {
			t.Errorf("FormatJSON() evaluation missing key %q", key)
		}
	}
}

func TestCategoryScoreToMap(t *testing.T) {
	score := CategoryScore{
		Score:      7.5,
		Rating:     "Good",
		Stars:      3,
		Assessment: "Solid performance",
	}

	result := categoryScoreToMap(score)

	// Check all expected fields are present
	expectedFields := []string{"score", "rating", "assessment", "stars"}
	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			t.Errorf("categoryScoreToMap() missing field %q", field)
		}
	}

	// Check values
	if result["score"] != 7.5 {
		t.Errorf("categoryScoreToMap() score = %v, want 7.5", result["score"])
	}
	// Rating is a custom type, compare as string
	ratingStr := string(result["rating"].(Rating))
	if ratingStr != "Good" {
		t.Errorf("categoryScoreToMap() rating = %q, want 'Good'", ratingStr)
	}
}

func TestAnalysisSectionToMap(t *testing.T) {
	section := AnalysisSection{
		Title:   "Defense Analysis",
		Summary: "Good defense",
		Details: []string{"Strong anti-air", "Good buildings"},
		Score:   8.0,
		Rating:  "Strong",
	}

	result := analysisSectionToMap(section)

	// Check all expected fields are present
	expectedFields := []string{"title", "summary", "details", "score", "rating"}
	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			t.Errorf("analysisSectionToMap() missing field %q", field)
		}
	}

	// Check values
	if result["title"] != "Defense Analysis" {
		t.Errorf("analysisSectionToMap() title = %q, want 'Defense Analysis'", result["title"])
	}

	// Check details is array
	details, ok := result["details"].([]string)
	if !ok {
		t.Errorf("analysisSectionToMap() details is not []string")
	} else if len(details) != 2 {
		t.Errorf("analysisSectionToMap() details length = %d, want 2", len(details))
	}
}

func TestSynergyMatrixToMap(t *testing.T) {
	matrix := SynergyMatrix{
		TotalScore:       7.5,
		AverageSynergy:   0.65,
		PairCount:        5,
		MaxPossiblePairs: 28,
		SynergyCoverage:  17.86,
		Pairs: []deck.SynergyPair{
			{Card1: "Hog Rider", Card2: "Ice Golem", Score: 0.8, Description: "Good cycle"},
		},
	}

	result := synergyMatrixToMap(matrix)

	// Check all expected fields are present
	expectedFields := []string{"total_score", "average_synergy", "pair_count", "max_possible_pairs", "synergy_coverage", "pairs"}
	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			t.Errorf("synergyMatrixToMap() missing field %q", field)
		}
	}

	// Check values
	if result["total_score"] != 7.5 {
		t.Errorf("synergyMatrixToMap() total_score = %v, want 7.5", result["total_score"])
	}

	// Check pairs is array
	pairs, ok := result["pairs"].([]map[string]any)
	if !ok {
		t.Errorf("synergyMatrixToMap() pairs is not []map[string]interface{}")
	} else if len(pairs) != 1 {
		t.Errorf("synergyMatrixToMap() pairs length = %d, want 1", len(pairs))
	} else {
		// Check first pair has required fields
		if _, ok := pairs[0]["card1"]; !ok {
			t.Errorf("synergyMatrixToMap() pair[0] missing card1")
		}
	}
}

func TestGenerateRecommendationsJSON(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()
	deckCards := []deck.CardCandidate{
		makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
		makeCard("Archers", deck.RoleSupport, 11, 11, "Common", 3),
	}

	result := Evaluate(deckCards, synergyDB, nil)

	recs := generateRecommendationsJSON(&result)

	// Should return an array (even if empty)
	if recs == nil {
		t.Errorf("generateRecommendationsJSON() returned nil")
	}

	// If there are recommendations, check their structure
	if len(recs) > 0 {
		for i, rec := range recs {
			expectedFields := []string{"priority", "title", "description"}
			for _, field := range expectedFields {
				if _, ok := rec[field]; !ok {
					t.Errorf("generateRecommendationsJSON() rec[%d] missing field %q", i, field)
				}
			}
		}
	}
}

func TestFormatJSONErrorHandling(t *testing.T) {
	// Test that FormatJSON handles errors gracefully
	// For now, we just verify it returns output for a valid evaluation
	synergyDB := deck.NewSynergyDatabase()
	deckCards := []deck.CardCandidate{
		makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
	}

	result := Evaluate(deckCards, synergyDB, nil)

	jsonOutput, err := FormatJSON(&result)
	if err != nil {
		t.Errorf("FormatJSON() returned unexpected error: %v", err)
	}

	if !strings.HasPrefix(jsonOutput, "{") {
		t.Errorf("FormatJSON() output doesn't start with '{', got: %s", jsonOutput[:min(10, len(jsonOutput))])
	}
}
