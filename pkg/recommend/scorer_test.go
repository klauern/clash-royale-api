package recommend

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
)

// TestNewScorer tests the scorer constructor
func TestNewScorer(t *testing.T) {
	scorer := NewScorer()
	if scorer == nil {
		t.Fatal("NewScorer returned nil")
	}
	if scorer.synergyDB == nil {
		t.Error("Scorer synergyDB should not be nil")
	}
}

// TestCalculateCompatibility_Perfect tests compatibility with perfect card levels
func TestCalculateCompatibility_Perfect(t *testing.T) {
	scorer := NewScorer()

	deckDetail := []deck.CardDetail{
		{Name: "Knight", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
		{Name: "Archers", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
		{Name: "Fireball", Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
		{Name: "Giant", Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
		{Name: "Mega Minion", Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 3},
		{Name: "Baby Dragon", Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 4},
		{Name: "Electro Wizard", Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 4},
		{Name: "Skeleton Army", Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 3},
	}

	playerCards := map[string]deck.CardLevelData{
		"Knight":         {Level: 14, MaxLevel: 14, Rarity: "Common"},
		"Archers":        {Level: 14, MaxLevel: 14, Rarity: "Common"},
		"Fireball":       {Level: 14, MaxLevel: 14, Rarity: "Rare"},
		"Giant":          {Level: 14, MaxLevel: 14, Rarity: "Rare"},
		"Mega Minion":    {Level: 14, MaxLevel: 14, Rarity: "Rare"},
		"Baby Dragon":    {Level: 14, MaxLevel: 14, Rarity: "Epic"},
		"Electro Wizard": {Level: 14, MaxLevel: 14, Rarity: "Legendary"},
		"Skeleton Army":  {Level: 14, MaxLevel: 14, Rarity: "Epic"},
	}

	score := scorer.CalculateCompatibility(deckDetail, playerCards)

	// Perfect match should be close to 100 (may be slightly higher due to rarity weights)
	// The actual score depends on configured rarity weights
	if score < 100.0 || score > 110.0 {
		t.Errorf("CalculateCompatibility with perfect levels = %.2f, want between 100.00 and 110.00", score)
	}
}

// TestCalculateCompatibility_PartialMatch tests compatibility with mixed card levels
func TestCalculateCompatibility_PartialMatch(t *testing.T) {
	scorer := NewScorer()

	deckDetail := []deck.CardDetail{
		{Name: "Knight", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
		{Name: "Archers", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
		{Name: "Fireball", Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
		{Name: "Giant", Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
	}

	playerCards := map[string]deck.CardLevelData{
		"Knight":   {Level: 14, MaxLevel: 14, Rarity: "Common"}, // 100% level
		"Archers":  {Level: 7, MaxLevel: 14, Rarity: "Common"},  // 50% level
		"Fireball": {Level: 10, MaxLevel: 14, Rarity: "Rare"},   // ~71% level
		"Giant":    {Level: 12, MaxLevel: 14, Rarity: "Rare"},   // ~86% level
	}

	score := scorer.CalculateCompatibility(deckDetail, playerCards)

	// Score should be between 0 and 100
	if score < 0 || score > 100 {
		t.Errorf("CalculateCompatibility score = %.2f, should be between 0 and 100", score)
	}

	// Should be less than perfect (100)
	if score >= 100 {
		t.Errorf("CalculateCompatibility with partial levels = %.2f, should be less than 100", score)
	}

	// Should be greater than 0 since we have some cards
	if score <= 0 {
		t.Errorf("CalculateCompatibility with some cards = %.2f, should be greater than 0", score)
	}
}

// TestCalculateCompatibility_MissingCards tests compatibility when player lacks cards
func TestCalculateCompatibility_MissingCards(t *testing.T) {
	scorer := NewScorer()

	deckDetail := []deck.CardDetail{
		{Name: "Knight", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
		{Name: "Archers", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
		{Name: "Fireball", Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
		{Name: "Giant", Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
	}

	playerCards := map[string]deck.CardLevelData{
		"Knight":  {Level: 14, MaxLevel: 14, Rarity: "Common"},
		"Archers": {Level: 14, MaxLevel: 14, Rarity: "Common"},
		// Fireball and Giant are missing
	}

	score := scorer.CalculateCompatibility(deckDetail, playerCards)

	// Should be 50% since we have 2 out of 4 cards maxed
	expectedScore := 50.0
	if score != expectedScore {
		t.Errorf("CalculateCompatibility with 2/4 cards = %.2f, want %.2f", score, expectedScore)
	}
}

// TestCalculateCompatibility_NoCards tests compatibility when player has no cards
func TestCalculateCompatibility_NoCards(t *testing.T) {
	scorer := NewScorer()

	deckDetail := []deck.CardDetail{
		{Name: "Knight", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
		{Name: "Archers", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
	}

	playerCards := map[string]deck.CardLevelData{}

	score := scorer.CalculateCompatibility(deckDetail, playerCards)

	// No cards = 0 score
	if score != 0.0 {
		t.Errorf("CalculateCompatibility with no player cards = %.2f, want 0.00", score)
	}
}

// TestCalculateCompatibility_EmptyDeck tests compatibility with empty deck
func TestCalculateCompatibility_EmptyDeck(t *testing.T) {
	scorer := NewScorer()

	deckDetail := []deck.CardDetail{}
	playerCards := map[string]deck.CardLevelData{
		"Knight": {Level: 14, MaxLevel: 14, Rarity: "Common"},
	}

	score := scorer.CalculateCompatibility(deckDetail, playerCards)

	// Empty deck = 0 score
	if score != 0.0 {
		t.Errorf("CalculateCompatibility with empty deck = %.2f, want 0.00", score)
	}
}

// TestCalculateCompatibility_RarityWeights tests that rarity affects scoring
func TestCalculateCompatibility_RarityWeights(t *testing.T) {
	scorer := NewScorer()

	// Test with common card at 50% level
	deckDetailCommon := []deck.CardDetail{
		{Name: "Knight", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
	}
	playerCardsCommon := map[string]deck.CardLevelData{
		"Knight": {Level: 7, MaxLevel: 14, Rarity: "Common"},
	}
	scoreCommon := scorer.CalculateCompatibility(deckDetailCommon, playerCardsCommon)

	// Test with legendary card at 50% level
	deckDetailLegendary := []deck.CardDetail{
		{Name: "Mega Knight", Level: 14, MaxLevel: 14, Rarity: "Legendary", Elixir: 7},
	}
	playerCardsLegendary := map[string]deck.CardLevelData{
		"Mega Knight": {Level: 7, MaxLevel: 14, Rarity: "Legendary"},
	}
	scoreLegendary := scorer.CalculateCompatibility(deckDetailLegendary, playerCardsLegendary)

	// Both should be valid scores
	if scoreCommon < 0 || scoreCommon > 100 {
		t.Errorf("Common card score = %.2f, should be between 0 and 100", scoreCommon)
	}
	if scoreLegendary < 0 || scoreLegendary > 100 {
		t.Errorf("Legendary card score = %.2f, should be between 0 and 100", scoreLegendary)
	}

	// The scores will differ based on rarity weights configured in config
	// Just verify both are reasonable
	if scoreCommon == 0 && scoreLegendary == 0 {
		t.Error("Both scores are 0, rarity weights may not be working")
	}
}

// TestCalculateSynergy tests basic synergy calculation
func TestCalculateSynergy(t *testing.T) {
	scorer := NewScorer()

	// Test with a typical deck
	deckNames := []string{
		"Hog Rider",
		"Valkyrie",
		"Musketeer",
		"Fireball",
		"Zap",
		"Ice Spirit",
		"Cannon",
		"Skeletons",
	}

	score := scorer.CalculateSynergy(deckNames)

	// Score should be between 0 and 100
	if score < 0 || score > 100 {
		t.Errorf("CalculateSynergy score = %.2f, should be between 0 and 100", score)
	}
}

// TestCalculateSynergy_EmptyDeck tests synergy with empty deck
func TestCalculateSynergy_EmptyDeck(t *testing.T) {
	scorer := NewScorer()

	deckNames := []string{}
	score := scorer.CalculateSynergy(deckNames)

	// Empty deck should have 0 synergy
	if score != 0.0 {
		t.Errorf("CalculateSynergy with empty deck = %.2f, want 0.00", score)
	}
}

// TestCalculateOverallScore tests overall score calculation with known weights
func TestCalculateOverallScore(t *testing.T) {
	scorer := NewScorer()

	tests := []struct {
		name          string
		compatibility float64
		synergy       float64
		archetypeFit  float64
		expectedMin   float64
		expectedMax   float64
	}{
		{
			name:          "Perfect scores",
			compatibility: 100.0,
			synergy:       100.0,
			archetypeFit:  100.0,
			expectedMin:   99.0,
			expectedMax:   100.0,
		},
		{
			name:          "Zero scores",
			compatibility: 0.0,
			synergy:       0.0,
			archetypeFit:  0.0,
			expectedMin:   0.0,
			expectedMax:   0.0,
		},
		{
			name:          "Mixed scores",
			compatibility: 80.0,
			synergy:       60.0,
			archetypeFit:  90.0,
			expectedMin:   50.0,
			expectedMax:   90.0,
		},
		{
			name:          "High compatibility dominates",
			compatibility: 100.0,
			synergy:       0.0,
			archetypeFit:  0.0,
			expectedMin:   50.0, // 60% of 100 = 60
			expectedMax:   70.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.CalculateOverallScore(tt.compatibility, tt.synergy, tt.archetypeFit)

			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("CalculateOverallScore(%.1f, %.1f, %.1f) = %.2f, want between %.2f and %.2f",
					tt.compatibility, tt.synergy, tt.archetypeFit, score, tt.expectedMin, tt.expectedMax)
			}

			// Overall score should always be between 0 and 100
			if score < 0 || score > 100 {
				t.Errorf("CalculateOverallScore = %.2f, should be between 0 and 100", score)
			}
		})
	}
}

// TestGenerateReasons tests reason generation for different score levels
func TestGenerateReasons(t *testing.T) {
	scorer := NewScorer()

	tests := []struct {
		name               string
		compatScore        float64
		synergyScore       float64
		overallScore       float64
		archetype          mulligan.Archetype
		avgElixir          float64
		recType            RecommendationType
		expectedContains   []string
		expectedNotContain []string
	}{
		{
			name:             "Excellent compatibility",
			compatScore:      85.0,
			synergyScore:     75.0,
			overallScore:     80.0,
			archetype:        "cycle",
			avgElixir:        2.9,
			recType:          TypeArchetypeMatch,
			expectedContains: []string{"Excellent card level match", "Excellent card synergies", "Ultra-low elixir"},
		},
		{
			name:             "Strong compatibility",
			compatScore:      65.0,
			synergyScore:     55.0,
			overallScore:     60.0,
			archetype:        "beatdown",
			avgElixir:        4.2,
			recType:          TypeArchetypeMatch,
			expectedContains: []string{"Strong card levels", "Good synergy", "High elixir beatdown"},
		},
		{
			name:             "Decent compatibility",
			compatScore:      45.0,
			synergyScore:     40.0,
			overallScore:     42.0,
			archetype:        "bait",
			avgElixir:        3.5,
			recType:          TypeArchetypeMatch,
			expectedContains: []string{"Decent card levels", "Bait archetype"},
		},
		{
			name:               "Low compatibility",
			compatScore:        25.0,
			synergyScore:       30.0,
			overallScore:       27.0,
			archetype:          "control",
			avgElixir:          3.8,
			recType:            TypeArchetypeMatch,
			expectedContains:   []string{"Consider upgrading", "Control archetype"},
			expectedNotContain: []string{"Excellent", "Strong"},
		},
		{
			name:             "Custom variation",
			compatScore:      70.0,
			synergyScore:     60.0,
			overallScore:     67.0,
			archetype:        "cycle",
			avgElixir:        3.2,
			recType:          TypeCustomVariation,
			expectedContains: []string{"Custom variation optimized", "Strong card levels"},
		},
		{
			name:             "Low elixir cycle",
			compatScore:      75.0,
			synergyScore:     65.0,
			overallScore:     72.0,
			archetype:        "cycle",
			avgElixir:        3.3,
			recType:          TypeArchetypeMatch,
			expectedContains: []string{"Low elixir cost supports cycle"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &DeckRecommendation{
				Deck: &deck.DeckRecommendation{
					AvgElixir: tt.avgElixir,
				},
				Archetype:          tt.archetype,
				ArchetypeName:      string(tt.archetype),
				CompatibilityScore: tt.compatScore,
				SynergyScore:       tt.synergyScore,
				OverallScore:       tt.overallScore,
				Type:               tt.recType,
			}

			reasons := scorer.GenerateReasons(rec)

			if len(reasons) == 0 {
				t.Error("GenerateReasons returned empty reasons list")
			}

			// Check for expected strings
			for _, expected := range tt.expectedContains {
				found := false
				for _, reason := range reasons {
					if contains(reason, expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected reason containing '%s', got reasons: %v", expected, reasons)
				}
			}

			// Check that certain strings are NOT present
			for _, notExpected := range tt.expectedNotContain {
				for _, reason := range reasons {
					if contains(reason, notExpected) {
						t.Errorf("Did not expect reason containing '%s', but found: %s", notExpected, reason)
					}
				}
			}
		})
	}
}

// TestGenerateReasons_AllArchetypes tests that all archetypes generate appropriate reasons
func TestGenerateReasons_AllArchetypes(t *testing.T) {
	scorer := NewScorer()

	archetypes := []mulligan.Archetype{
		"cycle",
		"beatdown",
		"bait",
		"control",
		"siege",
		"bridge_spam",
		"graveyard",
		"miner_control",
	}

	for _, archetype := range archetypes {
		t.Run(string(archetype), func(t *testing.T) {
			rec := &DeckRecommendation{
				Deck: &deck.DeckRecommendation{
					AvgElixir: 3.5,
				},
				Archetype:          archetype,
				ArchetypeName:      string(archetype),
				CompatibilityScore: 70.0,
				SynergyScore:       60.0,
				OverallScore:       65.0,
				Type:               TypeArchetypeMatch,
			}

			reasons := scorer.GenerateReasons(rec)

			if len(reasons) == 0 {
				t.Errorf("Archetype %s generated no reasons", archetype)
			}

			// At minimum, should have compatibility and type-based reason
			if len(reasons) < 2 {
				t.Errorf("Archetype %s generated only %d reasons, expected at least 2", archetype, len(reasons))
			}
		})
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(substr) > 0 && (s[:len(substr)] == substr ||
			(len(s) > len(substr) && contains(s[1:], substr))))
}
