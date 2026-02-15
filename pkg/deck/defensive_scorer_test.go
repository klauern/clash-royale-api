// Package deck provides tests for defensive scorer functionality
package deck

import (
	"slices"
	"testing"
)

func TestNewDefensiveScorer(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)

	if scorer == nil {
		t.Fatal("NewDefensiveScorer returned nil")
	}

	if scorer.matrix != matrix {
		t.Error("Scorer matrix not set correctly")
	}
}

func TestCalculateDefensiveCoverage_EmptyDeck(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)

	report := scorer.CalculateDefensiveCoverage([]string{})

	if report.OverallScore != 0.0 {
		t.Errorf("Expected score 0.0 for empty deck, got %f", report.OverallScore)
	}

	if len(report.CoverageGaps) == 0 {
		t.Error("Expected coverage gaps for empty deck")
	}
}

func TestCalculateDefensiveCoverage_BalancedDeck(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)

	// A well-balanced deck with good coverage
	balancedDeck := []string{
		"Musketeer",     // Air defense
		"Inferno Tower", // Tank killer, air defense, building
		"P.E.K.K.A",     // Tank killer
		"Valkyrie",      // Splash defense
		"The Log",       // Swarm clear
		"Zap",           // Swarm clear
		"Hog Rider",     // Win condition
		"Fireball",      // Big spell
	}

	report := scorer.CalculateDefensiveCoverage(balancedDeck)

	if report.OverallScore < 0.7 {
		t.Errorf("Expected high score for balanced deck, got %f", report.OverallScore)
	}

	if len(report.CoverageGaps) > 2 {
		t.Errorf("Expected minimal gaps for balanced deck, got %d: %v", len(report.CoverageGaps), report.CoverageGaps)
	}
}

func TestCalculateDefensiveCoverage_NoAirDefense(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)

	// Deck with no air defense
	noAirDeck := []string{
		"P.E.K.K.A", "Mini P.E.K.K.A", "Valkyrie", "The Log", "Zap", "Knight", "Golem", "Baby Dragon",
	}

	_ = scorer.CalculateDefensiveCoverage(noAirDeck)

	// Baby Dragon does have air defense, so let's create a truly no-air deck
	noAirDeck2 := []string{
		"P.E.K.K.A", "Mini P.E.K.K.A", "Valkyrie", "The Log", "Zap", "Knight", "Golem", "Hog Rider",
	}

	report2 := scorer.CalculateDefensiveCoverage(noAirDeck2)

	// Should have low air defense score
	airScore := report2.CategoryScores[CounterAirDefense]
	if airScore >= 0.6 {
		t.Errorf("Expected low air defense score, got %f", airScore)
	}

	// Should flag air defense as a gap
	hasAirGap := slices.Contains(report2.CoverageGaps, "Insufficient air defense")
	if !hasAirGap {
		t.Error("Expected air defense gap to be flagged")
	}
}

func TestCalculateDefensiveCoverage_NoSplash(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)

	// Deck with no splash damage
	noSplashDeck := []string{
		"Musketeer", "Inferno Tower", "Hog Rider", "Knight", "The Log", "Zap", "Ice Spirit", "Skeletons",
	}

	report := scorer.CalculateDefensiveCoverage(noSplashDeck)

	// Should have low splash score
	splashScore := report.CategoryScores[CounterSplashDefense]
	if splashScore >= 0.6 {
		t.Errorf("Expected low splash score, got %f", splashScore)
	}

	// Should flag splash as a gap
	hasSplashGap := slices.Contains(report.CoverageGaps, "No splash damage")
	if !hasSplashGap {
		t.Error("Expected splash gap to be flagged")
	}
}

func TestCalculateDefensiveCoverage_NoSwarmSpell(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)

	// Deck with no swarm spells (Valkyrie removed since she's in swarm_clear category)
	noSwarmDeck := []string{
		"P.E.K.K.A", "Musketeer", "Hog Rider", "Knight", "Golem", "Baby Dragon", "Fireball", "Mini P.E.K.K.A",
	}

	report := scorer.CalculateDefensiveCoverage(noSwarmDeck)

	// Should flag swarm spell as a gap (no Log, Zap, Arrows, or Valkyrie)
	hasSwarmGap := slices.Contains(report.CoverageGaps, "No swarm spell")
	if !hasSwarmGap {
		t.Error("Expected swarm spell gap to be flagged")
	}
}

func TestScoreCategory(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)

	// Test air defense scoring
	perfectAirDeck := []string{"Musketeer", "Electro Wizard", "Inferno Tower", "Baby Dragon"}
	score := scorer.scoreCategory(perfectAirDeck, CounterAirDefense, 2, 3)
	if score != 1.0 {
		t.Errorf("Expected perfect score, got %f", score)
	}

	// Below minimum
	badAirDeck := []string{"Hog Rider", "Knight"}
	score = scorer.scoreCategory(badAirDeck, CounterAirDefense, 2, 3)
	if score >= 0.6 {
		t.Errorf("Expected score below 0.6 for below-minimum, got %f", score)
	}
}

func TestAnalyzeCommonThreats(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)

	// Deck with good counters
	goodDeck := []string{"Inferno Tower", "P.E.K.K.A", "Musketeer", "Baby Dragon", "Valkyrie", "The Log", "Zap", "Hog Rider"}
	threatAnalysis := scorer.analyzeCommonThreats(goodDeck)

	if len(threatAnalysis) == 0 {
		t.Error("Expected threat analysis results")
	}

	// Should analyze common threats like Mega Knight, Balloon, etc.
	threatNames := make(map[string]bool)
	for _, analysis := range threatAnalysis {
		threatNames[analysis.ThreatName] = true
	}

	expectedThreats := []string{"Mega Knight", "Balloon", "Graveyard", "Hog Rider"}
	for _, threat := range expectedThreats {
		if !threatNames[threat] {
			t.Errorf("Expected analysis for threat: %s", threat)
		}
	}
}

func TestIdentifyGaps(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)

	// Create scores with gaps
	scores := map[CounterCategory]float64{
		CounterAirDefense:    0.3, // Below 0.6
		CounterTankKillers:   1.0,
		CounterSplashDefense: 0.4, // Below 0.6
		CounterSwarmClear:    0.8,
		CounterBuildings:     0.3, // Below 0.4
	}

	gaps := scorer.identifyGaps([]string{}, scores)

	if len(gaps) != 3 {
		t.Errorf("Expected 3 gaps, got %d: %v", len(gaps), gaps)
	}
}

func TestGenerateRecommendations(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)
	const (
		knightCard   = "Knight"
		hogRiderCard = "Hog Rider"
	)

	gaps := []string{"Insufficient air defense", "No splash damage"}
	deckCards := []string{hogRiderCard, knightCard, "Golem", "P.E.K.K.A"}

	recommendations := scorer.generateRecommendations(deckCards, gaps)

	if len(recommendations) == 0 {
		t.Error("Expected recommendations for gaps")
	}

	// Should recommend air defenders and splash cards not in deck
	for _, rec := range recommendations {
		if rec == hogRiderCard || rec == knightCard || rec == "Golem" || rec == "P.E.K.K.A" {
			t.Errorf("Should not recommend cards already in deck: %s", rec)
		}
	}
}

func TestGetScoreForCategory(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)

	deck := []string{"Musketeer", "Electro Wizard", "Inferno Tower", "Baby Dragon", "Valkyrie", "The Log", "Zap", "Hog Rider"}

	// Test each category
	categories := []CounterCategory{
		CounterAirDefense,
		CounterTankKillers,
		CounterSplashDefense,
		CounterSwarmClear,
		CounterBuildings,
	}

	for _, category := range categories {
		score := scorer.GetScoreForCategory(deck, category)
		if score < 0.0 || score > 1.0 {
			t.Errorf("Score out of range for %s: %f", category, score)
		}
	}
}

// Integration test: Deck with no splash defense scores low against swarm
func TestDeckWithNoSplashScoresLow(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)

	// Deck with no splash defense
	noSplashDeck := []string{
		"Musketeer", "Inferno Tower", "Hog Rider", "Knight", "The Log", "Zap", "Ice Spirit", "Skeletons",
	}

	report := scorer.CalculateDefensiveCoverage(noSplashDeck)

	splashScore := report.CategoryScores[CounterSplashDefense]
	if splashScore >= 0.6 {
		t.Errorf("Deck with no splash should score low on splash defense, got %f", splashScore)
	}
}

// Integration test: Deck with no tank killer scores low
func TestDeckWithNoTankKillerScoresLow(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)

	// Deck with no tank killers
	noTankKillerDeck := []string{
		"Musketeer", "Valkyrie", "Hog Rider", "Knight", "The Log", "Zap", "Ice Spirit", "Fireball",
	}

	report := scorer.CalculateDefensiveCoverage(noTankKillerDeck)

	tankScore := report.CategoryScores[CounterTankKillers]
	if tankScore >= 0.6 {
		t.Errorf("Deck with no tank killer should score low on tank killer score, got %f", tankScore)
	}
}

// Integration test: Balanced deck scores high
func TestBalancedDeckScoresHigh(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)

	// Well-balanced deck
	balancedDeck := []string{
		"Musketeer", "Inferno Tower", "P.E.K.K.A", "Valkyrie", "The Log", "Zap", "Hog Rider", "Fireball",
	}

	report := scorer.CalculateDefensiveCoverage(balancedDeck)

	if report.OverallScore < 0.7 {
		t.Errorf("Balanced deck should score high, got %f", report.OverallScore)
	}

	// Should have strong counters identified
	if len(report.StrongCounters) == 0 {
		t.Error("Expected strong counters to be identified")
	}
}

func BenchmarkCalculateDefensiveCoverage(b *testing.B) {
	matrix := NewCounterMatrixWithDefaults()
	scorer := NewDefensiveScorer(matrix)
	deck := []string{"Musketeer", "Inferno Tower", "P.E.K.K.A", "Valkyrie", "The Log", "Zap", "Hog Rider", "Fireball"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scorer.CalculateDefensiveCoverage(deck)
	}
}
