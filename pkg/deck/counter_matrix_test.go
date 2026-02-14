// Package deck provides tests for counter matrix functionality
package deck

import (
	"testing"
)

func TestNewCounterMatrix(t *testing.T) {
	matrix := NewCounterMatrix()

	if matrix == nil {
		t.Fatal("NewCounterMatrix returned nil")
	}

	if matrix.threatCounters == nil {
		t.Error("threatCounters map not initialized")
	}

	if matrix.counterCategories == nil {
		t.Error("counterCategories map not initialized")
	}

	if matrix.cardCapabilities == nil {
		t.Error("cardCapabilities map not initialized")
	}
}

func TestNewCounterMatrixWithDefaults(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()

	// Check that default threats are loaded
	mkCounters := matrix.GetCountersForThreat("Mega Knight")
	if len(mkCounters) == 0 {
		t.Error("No default counters loaded for Mega Knight")
	}

	// Check specific counter
	hasInfernoTower := false
	for _, counter := range mkCounters {
		if counter.Card == "Inferno Tower" {
			hasInfernoTower = true
			if counter.Effectiveness != 1.0 {
				t.Errorf("Inferno Tower effectiveness should be 1.0, got %f", counter.Effectiveness)
			}
		}
	}
	if !hasInfernoTower {
		t.Error("Inferno Tower not found as Mega Knight counter")
	}

	// Check counter categories
	airCards := matrix.GetCardsInCategory(CounterAirDefense)
	if len(airCards) == 0 {
		t.Error("No air defense cards loaded")
	}
}

func TestGetCounterEffectiveness(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()

	// Test existing counter
	effectiveness := matrix.GetCounterEffectiveness("Mega Knight", "Inferno Tower")
	if effectiveness != 1.0 {
		t.Errorf("Expected effectiveness 1.0, got %f", effectiveness)
	}

	// Test non-existing counter
	effectiveness = matrix.GetCounterEffectiveness("Mega Knight", "Skeletons")
	if effectiveness != 0.0 {
		t.Errorf("Expected effectiveness 0.0 for non-counter, got %f", effectiveness)
	}

	// Test non-existing threat
	effectiveness = matrix.GetCounterEffectiveness("Unknown Threat", "Inferno Tower")
	if effectiveness != 0.0 {
		t.Errorf("Expected effectiveness 0.0 for non-existing threat, got %f", effectiveness)
	}
}

func TestHasCapability(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()

	// Test air defense capability
	if !matrix.HasCapability("Musketeer", CounterAirDefense) {
		t.Error("Musketeer should have air defense capability")
	}

	// Test splash defense capability
	if !matrix.HasCapability("Baby Dragon", CounterSplashDefense) {
		t.Error("Baby Dragon should have splash defense capability")
	}

	// Test non-existing capability
	if matrix.HasCapability("Skeletons", CounterAirDefense) {
		t.Error("Skeletons should not have air defense capability")
	}
}

func TestCountCardsWithCapability(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()

	deck := []string{"Musketeer", "Electro Wizard", "Inferno Tower", "Baby Dragon", "Valkyrie"}

	// Count air defense
	airCount := matrix.CountCardsWithCapability(deck, CounterAirDefense)
	expectedAir := 3 // Musketeer, E-Wiz, Inferno Tower
	if airCount != expectedAir {
		t.Errorf("Expected %d air defense cards, got %d", expectedAir, airCount)
	}

	// Count splash
	splashCount := matrix.CountCardsWithCapability(deck, CounterSplashDefense)
	expectedSplash := 2 // Baby Dragon, Valkyrie
	if splashCount != expectedSplash {
		t.Errorf("Expected %d splash cards, got %d", expectedSplash, splashCount)
	}
}

func TestResetRetargetCapability(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()

	deck := []string{"Hog Rider", "Musketeer", "Zap", "Electro Wizard", "Cannon"}
	count := matrix.CountCardsWithCapability(deck, CounterResetRetarget)
	if count != 2 {
		t.Fatalf("expected 2 reset/retarget cards, got %d", count)
	}

	matched := matrix.GetDeckCardsWithCapability(deck, CounterResetRetarget)
	if len(matched) != 2 {
		t.Fatalf("expected 2 matched reset/retarget cards, got %d", len(matched))
	}
	if matched[0] != "Zap" || matched[1] != "Electro Wizard" {
		t.Fatalf("unexpected matched cards: %v", matched)
	}
}

func TestAnalyzeThreatCoverage(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()

	// Test deck with good counters
	goodDeck := []string{"Inferno Tower", "P.E.K.K.A", "Musketeer", "Baby Dragon", "Valkyrie", "The Log", "Tornado", "Hog Rider"}
	coverage := matrix.AnalyzeThreatCoverage(goodDeck, "Mega Knight")

	if !coverage.CanCounter {
		t.Error("Deck should be able to counter Mega Knight")
	}

	if coverage.Effectiveness < 0.8 {
		t.Errorf("Expected high effectiveness, got %f", coverage.Effectiveness)
	}

	if len(coverage.DeckCounters) == 0 {
		t.Error("Expected deck counters to be populated")
	}

	// Test deck with no counters
	badDeck := []string{"Skeletons", "Goblins", "Spear Goblins", "Ice Spirit", "Fire Spirit"}
	coverage = matrix.AnalyzeThreatCoverage(badDeck, "Mega Knight")

	if coverage.CanCounter {
		t.Error("Deck should not be able to counter Mega Knight")
	}

	if coverage.Effectiveness != 0.0 {
		t.Errorf("Expected zero effectiveness, got %f", coverage.Effectiveness)
	}
}

func TestAnalyzeThreatCoverageBalloon(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()

	// Deck with air defense
	airDeck := []string{"Musketeer", "Inferno Tower", "Electro Wizard"}
	coverage := matrix.AnalyzeThreatCoverage(airDeck, "Balloon")

	if !coverage.CanCounter {
		t.Error("Deck with air defense should counter Balloon")
	}

	// Deck without air defense
	noAirDeck := []string{"Knight", "Valkyrie", "Golem", "P.E.K.K.A"}
	coverage = matrix.AnalyzeThreatCoverage(noAirDeck, "Balloon")

	if coverage.CanCounter {
		t.Error("Deck without air defense should not counter Balloon")
	}

	if coverage.Suggestion == "" {
		t.Error("Should provide suggestion when no counters available")
	}
}

func TestThreatCoverageReason(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()

	// Excellent counter
	excellentDeck := []string{"Inferno Tower", "Inferno Dragon"}
	coverage := matrix.AnalyzeThreatCoverage(excellentDeck, "Mega Knight")

	if coverage.Reason == "" {
		t.Error("Reason should not be empty")
	}

	// No counter
	badDeck := []string{"Skeletons"}
	coverage = matrix.AnalyzeThreatCoverage(badDeck, "Balloon")

	if coverage.Reason == "" {
		t.Error("Reason should be provided even for bad coverage")
	}
}

func BenchmarkAnalyzeThreatCoverage(b *testing.B) {
	matrix := NewCounterMatrixWithDefaults()
	deck := []string{"Inferno Tower", "P.E.K.K.A", "Musketeer", "Baby Dragon", "Valkyrie", "The Log", "Zap", "Hog Rider"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matrix.AnalyzeThreatCoverage(deck, "Mega Knight")
	}
}

func BenchmarkGetCounterEffectiveness(b *testing.B) {
	matrix := NewCounterMatrixWithDefaults()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matrix.GetCounterEffectiveness("Mega Knight", "Inferno Tower")
	}
}

func BenchmarkCountCardsWithCapability(b *testing.B) {
	matrix := NewCounterMatrixWithDefaults()
	deck := []string{"Musketeer", "Electro Wizard", "Inferno Tower", "Baby Dragon", "Valkyrie", "The Log", "Zap", "Hog Rider"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matrix.CountCardsWithCapability(deck, CounterAirDefense)
	}
}
