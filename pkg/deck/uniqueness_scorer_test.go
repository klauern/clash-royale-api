package deck

import (
	"testing"
)

func TestCardPopularity_GetPopularity(t *testing.T) {
	cp := NewCardPopularity()

	// Test known card
	popularity := cp.GetPopularity("Zap")
	if popularity <= 0 || popularity > 1.0 {
		t.Errorf("Expected Zap popularity to be between 0 and 1, got %f", popularity)
	}

	// Test unknown card returns default
	unknownPopularity := cp.GetPopularity("UnknownCard")
	if unknownPopularity != 0.5 {
		t.Errorf("Expected unknown card popularity to be 0.5, got %f", unknownPopularity)
	}
}

func TestCardPopularity_GetUniquenessScore(t *testing.T) {
	cp := NewCardPopularity()

	// Very common card should have low uniqueness
	zapUniqueness := cp.GetUniquenessScore("Zap")
	if zapUniqueness >= 0.5 {
		t.Errorf("Expected Zap to have low uniqueness (< 0.5), got %f", zapUniqueness)
	}

	// Rare card should have high uniqueness
	healUniqueness := cp.GetUniquenessScore("Heal")
	if healUniqueness <= 0.5 {
		t.Errorf("Expected Heal to have high uniqueness (> 0.5), got %f", healUniqueness)
	}
}

func TestUniquenessScorer_ScoreDeck(t *testing.T) {
	config := UniquenessConfig{
		Enabled:                true,
		Weight:                 0.2,
		MinUniquenessThreshold: 0.3,
		UseGeometricMean:       false,
	}
	scorer := NewUniquenessScorer(config)

	// Test with a mix of common and unique cards
	deck := []string{"Zap", "Hog Rider", "Heal", "Clone", "Knight", "Archers", "Golem", "Musketeer"}
	score := scorer.ScoreDeck(deck)

	if score < 0 || score > 1.0 {
		t.Errorf("Expected score to be between 0 and 1, got %f", score)
	}

	// Score should be positive with some unique cards
	if score == 0 {
		t.Error("Expected non-zero score for deck with some unique cards")
	}
}

func TestUniquenessScorer_ScoreDeck_Disabled(t *testing.T) {
	config := UniquenessConfig{
		Enabled: false,
		Weight:  0.2,
	}
	scorer := NewUniquenessScorer(config)

	deck := []string{"Zap", "Hog Rider", "Heal", "Clone"}
	score := scorer.ScoreDeck(deck)

	if score != 0 {
		t.Errorf("Expected score to be 0 when disabled, got %f", score)
	}
}

func TestUniquenessScorer_ScoreDeckWithDetails(t *testing.T) {
	config := UniquenessConfig{
		Enabled:                true,
		Weight:                 0.2,
		MinUniquenessThreshold: 0.3,
		UseGeometricMean:       false,
	}
	scorer := NewUniquenessScorer(config)

	deck := []string{"Zap", "Hog Rider", "Heal", "Clone", "Knight", "Archers", "Golem", "Musketeer"}
	result := scorer.ScoreDeckWithDetails(deck)

	if result.FinalScore < 0 || result.FinalScore > 1.0 {
		t.Errorf("Expected FinalScore to be between 0 and 1, got %f", result.FinalScore)
	}

	if result.WeightedScore != result.FinalScore*config.Weight {
		t.Errorf("Expected WeightedScore to be FinalScore * Weight, got %f (expected %f)",
			result.WeightedScore, result.FinalScore*config.Weight)
	}

	if len(result.CardUniqueness) != len(deck) {
		t.Errorf("Expected CardUniqueness to have %d entries, got %d", len(deck), len(result.CardUniqueness))
	}

	// Check that most/least unique cards are identified
	if result.MostUniqueCard == "" {
		t.Error("Expected MostUniqueCard to be set")
	}
	if result.LeastUniqueCard == "" {
		t.Error("Expected LeastUniqueCard to be set")
	}
}

func TestUniquenessScorer_GeometricMean(t *testing.T) {
	config := UniquenessConfig{
		Enabled:                true,
		Weight:                 0.2,
		MinUniquenessThreshold: 0.0, // No threshold so all cards count
		UseGeometricMean:       true,
	}
	scorer := NewUniquenessScorer(config)

	// Test with all cards that have some uniqueness (above 0)
	// Zap has popularity ~0.95, so uniqueness ~0.05
	deck := []string{"Zap", "Heal", "Clone", "Rage", "Mirror", "Guardian", "Knight", "Archers"}
	score := scorer.ScoreDeck(deck)

	// Score should be positive since all cards have some uniqueness
	if score <= 0 {
		t.Errorf("Expected geometric mean to be positive, got %f", score)
	}

	// Geometric mean should be less than or equal to arithmetic mean
	config.UseGeometricMean = false
	arithmeticScorer := NewUniquenessScorer(config)
	arithmeticScore := arithmeticScorer.ScoreDeck(deck)

	if score > arithmeticScore {
		t.Errorf("Expected geometric mean (%f) to be <= arithmetic mean (%f)", score, arithmeticScore)
	}
}

func TestGetCardUniquenessTier(t *testing.T) {
	tests := []struct {
		uniqueness float64
		expected   string
	}{
		{0.9, "Very Unique"},
		{0.7, "Unique"},
		{0.5, "Moderate"},
		{0.3, "Common"},
		{0.1, "Very Common"},
		{0.0, "Very Common"},
	}

	for _, test := range tests {
		tier := GetCardUniquenessTier(test.uniqueness)
		if tier != test.expected {
			t.Errorf("GetCardUniquenessTier(%f) = %s, expected %s", test.uniqueness, tier, test.expected)
		}
	}
}

func TestCardPopularity_GetMostCommonCards(t *testing.T) {
	cp := NewCardPopularity()

	// Get top 5 most common cards
	common := cp.GetMostCommonCards(5)

	if len(common) != 5 {
		t.Errorf("Expected 5 common cards, got %d", len(common))
	}

	// Verify they are sorted by popularity descending
	for i := 1; i < len(common); i++ {
		if common[i].Popularity > common[i-1].Popularity {
			t.Error("Expected common cards to be sorted by popularity descending")
		}
	}
}

func TestCardPopularity_GetMostUniqueCards(t *testing.T) {
	cp := NewCardPopularity()

	// Get top 5 most unique cards
	unique := cp.GetMostUniqueCards(5)

	if len(unique) != 5 {
		t.Errorf("Expected 5 unique cards, got %d", len(unique))
	}

	// Verify they are sorted by popularity ascending
	for i := 1; i < len(unique); i++ {
		if unique[i].Popularity < unique[i-1].Popularity {
			t.Error("Expected unique cards to be sorted by popularity ascending")
		}
	}

	// Verify uniqueness is calculated correctly
	for _, card := range unique {
		expectedUniqueness := 1.0 - card.Popularity
		if card.Uniqueness != expectedUniqueness {
			t.Errorf("Expected uniqueness %f for %s, got %f", expectedUniqueness, card.CardName, card.Uniqueness)
		}
	}
}

func TestDefaultUniquenessConfig(t *testing.T) {
	config := DefaultUniquenessConfig()

	if config.Enabled {
		t.Error("Expected uniqueness to be disabled by default")
	}

	if config.Weight != 0.15 {
		t.Errorf("Expected default weight to be 0.15, got %f", config.Weight)
	}

	if config.MinUniquenessThreshold != 0.5 {
		t.Errorf("Expected default threshold to be 0.5, got %f", config.MinUniquenessThreshold)
	}

	if config.UseGeometricMean {
		t.Error("Expected geometric mean to be disabled by default")
	}
}

func TestGlobalUniquenessScorer(t *testing.T) {
	// Test that global scorer exists and can be enabled/disabled
	if GlobalUniquenessScorer == nil {
		t.Fatal("GlobalUniquenessScorer should not be nil")
	}

	// Enable scoring
	EnableUniquenessScoring(0.2)
	if !GlobalUniquenessScorer.config.Enabled {
		t.Error("Expected global scorer to be enabled")
	}
	if GlobalUniquenessScorer.config.Weight != 0.2 {
		t.Errorf("Expected weight to be 0.2, got %f", GlobalUniquenessScorer.config.Weight)
	}

	// Disable scoring
	DisableUniquenessScoring()
	if GlobalUniquenessScorer.config.Enabled {
		t.Error("Expected global scorer to be disabled")
	}
}

func TestCardPopularity_SetPopularity(t *testing.T) {
	cp := NewCardPopularity()

	// Set a new popularity value
	cp.SetPopularity("TestCard", 0.75)

	// Verify it was set
	popularity := cp.GetPopularity("TestCard")
	if popularity != 0.75 {
		t.Errorf("Expected popularity 0.75, got %f", popularity)
	}

	// Test clamping - values should be clamped to 0-1 range
	cp.SetPopularity("TestCard2", 1.5)
	if cp.GetPopularity("TestCard2") != 1.0 {
		t.Errorf("Expected popularity to be clamped to 1.0, got %f", cp.GetPopularity("TestCard2"))
	}

	cp.SetPopularity("TestCard3", -0.5)
	if cp.GetPopularity("TestCard3") != 0.0 {
		t.Errorf("Expected popularity to be clamped to 0.0, got %f", cp.GetPopularity("TestCard3"))
	}
}
