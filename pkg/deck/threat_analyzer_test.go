// Package deck provides tests for threat analyzer functionality
package deck

import (
	"testing"
)

func TestNewThreatAnalyzer(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	if analyzer == nil {
		t.Fatal("NewThreatAnalyzer returned nil")
	}

	if analyzer.matrix != matrix {
		t.Error("Analyzer matrix not set correctly")
	}

	if len(analyzer.metaThreats) == 0 {
		t.Error("Meta threats not loaded")
	}
}

func TestGetDefaultMetaThreats(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	threats := analyzer.GetMetaThreats()

	if len(threats) == 0 {
		t.Fatal("No meta threats loaded")
	}

	// Check for expected threats
	threatNames := make(map[string]bool)
	for _, threat := range threats {
		threatNames[threat.Name] = true
	}

	expectedThreats := []string{"Mega Knight", "Balloon", "Graveyard", "Hog Rider", "Golem", "Lava Hound"}
	for _, name := range expectedThreats {
		if !threatNames[name] {
			t.Errorf("Expected threat not found: %s", name)
		}
	}
}

func TestAnalyzeDeck_EmptyDeck(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	report := analyzer.AnalyzeDeck([]string{})

	if report.OverallDefensiveScore != 0.0 {
		t.Errorf("Expected score 0.0 for empty deck, got %f", report.OverallDefensiveScore)
	}

	if len(report.CriticalGaps) == 0 {
		t.Error("Expected critical gaps for empty deck")
	}
}

func TestAnalyzeDeck_BalancedDeck(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	// Well-balanced deck
	balancedDeck := []string{
		"Musketeer", "Inferno Tower", "P.E.K.K.A", "Valkyrie", "The Log", "Zap", "Hog Rider", "Fireball",
	}

	report := analyzer.AnalyzeDeck(balancedDeck)

	if report.OverallDefensiveScore < 0.1 {
		t.Errorf("Expected non-zero score for balanced deck, got %f", report.OverallDefensiveScore)
	}

	if len(report.MetaThreatMatches) == 0 {
		t.Error("Expected threat matches to be analyzed")
	}

	// Note: With default limited threat data, there will be more gaps
	// The JSON file has comprehensive data for production use
}

func TestAnalyzeDeck_NoAirDefense(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	// Deck with no air defense
	noAirDeck := []string{
		"P.E.K.K.A", "Mini P.E.K.K.A", "Valkyrie", "The Log", "Zap", "Knight", "Golem", "Hog Rider",
	}

	report := analyzer.AnalyzeDeck(noAirDeck)

	// Should have critical gaps for air threats
	hasAirGap := false
	for _, gap := range report.CriticalGaps {
		if threatContains(gap, "Balloon") || threatContains(gap, "Lava Hound") || threatContains(gap, "air") {
			hasAirGap = true
			break
		}
	}
	if !hasAirGap {
		t.Error("Expected air threats to be flagged as critical gaps")
	}
}

func TestAnalyzeThreat(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	// Test with good counters
	goodDeck := []string{"Inferno Tower", "P.E.K.K.A", "Musketeer", "Baby Dragon"}
	threat := ThreatDefinition{
		Name:          "Mega Knight",
		Type:          ThreatTypeTank,
		Description:   "High HP jumping tank",
		MetaRelevance: 0.95,
	}

	match := analyzer.analyzeThreat(goodDeck, threat)

	if !match.CanCounter {
		t.Error("Deck should be able to counter Mega Knight")
	}

	if match.Effectiveness < 0.8 {
		t.Errorf("Expected high effectiveness, got %f", match.Effectiveness)
	}

	if len(match.CounterCards) == 0 {
		t.Error("Expected counter cards to be identified")
	}
}

func TestAnalyzeThreat_NoCounters(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	// Test with no counters
	badDeck := []string{"Skeletons", "Goblins", "Spear Goblins"}
	threat := ThreatDefinition{
		Name:          "Mega Knight",
		Type:          ThreatTypeTank,
		Description:   "High HP jumping tank",
		MetaRelevance: 0.95,
	}

	match := analyzer.analyzeThreat(badDeck, threat)

	if match.CanCounter {
		t.Error("Deck should not be able to counter Mega Knight")
	}

	if match.Effectiveness != 0.0 {
		t.Errorf("Expected zero effectiveness, got %f", match.Effectiveness)
	}
}

func TestCalculateOverallScore(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	// Create test matches
	matches := []ThreatMatch{
		{
			Threat:        ThreatDefinition{Name: "Threat1", MetaRelevance: 1.0},
			Effectiveness: 1.0,
		},
		{
			Threat:        ThreatDefinition{Name: "Threat2", MetaRelevance: 0.5},
			Effectiveness: 0.5,
		},
	}

	score := analyzer.calculateOverallScore(matches)

	// Weighted: (1.0*1.0 + 0.5*0.5) / 1.5 = 1.25/1.5 ≈ 0.833
	if score < 0.8 || score > 0.9 {
		t.Errorf("Expected score around 0.83, got %f", score)
	}
}

func TestIdentifyCriticalGaps(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	// Create matches with gaps
	matches := []ThreatMatch{
		{
			Threat:        ThreatDefinition{Name: "Balloon", MetaRelevance: 0.85, Type: ThreatTypeAir},
			CanCounter:    false,
			Effectiveness: 0.0,
		},
		{
			Threat:        ThreatDefinition{Name: "Mega Knight", MetaRelevance: 0.95, Type: ThreatTypeTank},
			CanCounter:    true,
			Effectiveness: 0.9,
		},
		{
			Threat:        ThreatDefinition{Name: "Hog Rider", MetaRelevance: 0.9, Type: ThreatTypeWinCondition},
			CanCounter:    false,
			Effectiveness: 0.3,
		},
	}

	gaps := analyzer.identifyCriticalGaps(matches)

	// Should identify Balloon and Hog Rider as gaps
	if len(gaps) < 2 {
		t.Errorf("Expected at least 2 critical gaps, got %d: %v", len(gaps), gaps)
	}

	// Check that high-meta threat with weak counter is flagged
	hasHogWeak := false
	for _, gap := range gaps {
		if threatContains(gap, "Hog Rider") {
			hasHogWeak = true
			break
		}
	}
	if !hasHogWeak {
		t.Error("Expected Hog Rider with weak counter to be flagged")
	}
}

func TestIdentifyStrongDefenses(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	// Create matches with strong defenses
	matches := []ThreatMatch{
		{
			Threat:        ThreatDefinition{Name: "Balloon", MetaRelevance: 0.85, Type: ThreatTypeAir},
			CanCounter:    true,
			Effectiveness: 0.9,
		},
		{
			Threat:        ThreatDefinition{Name: "Mega Knight", MetaRelevance: 0.95, Type: ThreatTypeTank},
			CanCounter:    true,
			Effectiveness: 0.85,
		},
		{
			Threat:        ThreatDefinition{Name: "Minor Threat", MetaRelevance: 0.5, Type: ThreatTypeWinCondition},
			CanCounter:    true,
			Effectiveness: 1.0,
		},
	}

	strong := analyzer.identifyStrongDefenses(matches)

	// Should include Balloon and Mega Knight (high meta + high effectiveness)
	// Should exclude Minor Threat (low meta relevance even if perfect counter)
	if len(strong) != 2 {
		t.Errorf("Expected 2 strong defenses, got %d", len(strong))
	}

	// Check that they're sorted by effectiveness
	if len(strong) >= 2 && strong[0].Effectiveness < strong[1].Effectiveness {
		t.Error("Strong defenses should be sorted by effectiveness")
	}
}

func TestGetMatchForThreat(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	deck := []string{"Inferno Tower", "P.E.K.K.A"}

	// Test existing threat
	match := analyzer.GetMatchForThreat(deck, "Mega Knight")
	if match == nil {
		t.Error("Expected match for Mega Knight")
		return
	}

	if match.Threat.Name != "Mega Knight" {
		t.Errorf("Expected threat name Mega Knight, got %s", match.Threat.Name)
	}

	// Test non-existing threat
	match = analyzer.GetMatchForThreat(deck, "Unknown Threat")
	if match != nil {
		t.Error("Expected nil for non-existing threat")
	}
}

func TestGetThreatsByType(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	// Get air threats
	airThreats := analyzer.GetThreatsByType(ThreatTypeAir)

	if len(airThreats) == 0 {
		t.Error("Expected air threats to be found")
	}

	// Check that all returned threats are air type
	for _, threat := range airThreats {
		if threat.Type != ThreatTypeAir {
			t.Errorf("Expected air threat, got %s", threat.Type)
		}
	}
}

func TestAddCustomThreat(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	customThreat := ThreatDefinition{
		Name:          "Custom Threat",
		Type:          ThreatTypeWinCondition,
		Description:   "A custom threat for testing",
		MetaRelevance: 0.7,
	}

	analyzer.AddCustomThreat(customThreat)

	threats := analyzer.GetMetaThreats()

	// Check that custom threat was added
	found := false
	for _, threat := range threats {
		if threat.Name == "Custom Threat" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Custom threat was not added")
	}
}

func TestSetMetaThreats(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	customThreats := []ThreatDefinition{
		{Name: "Threat1", Type: ThreatTypeTank, Description: "Test", MetaRelevance: 0.8},
		{Name: "Threat2", Type: ThreatTypeAir, Description: "Test", MetaRelevance: 0.7},
	}

	analyzer.SetMetaThreats(customThreats)

	threats := analyzer.GetMetaThreats()

	if len(threats) != 2 {
		t.Errorf("Expected 2 threats, got %d", len(threats))
	}
}

func TestThreatBreakdown(t *testing.T) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)

	deck := []string{"Musketeer", "Inferno Tower", "P.E.K.K.A", "Valkyrie", "The Log", "Zap", "Hog Rider", "Fireball"}

	report := analyzer.AnalyzeDeck(deck)

	if len(report.ThreatBreakdown) == 0 {
		t.Error("Expected threat breakdown to be populated")
	}

	// Check that threat types are counted
	hasTankThreats := false
	hasAirThreats := false
	for threatType, count := range report.ThreatBreakdown {
		if threatType == ThreatTypeTank && count > 0 {
			hasTankThreats = true
		}
		if threatType == ThreatTypeAir && count > 0 {
			hasAirThreats = true
		}
	}

	if !hasTankThreats {
		t.Error("Expected tank threats in breakdown")
	}
	if !hasAirThreats {
		t.Error("Expected air threats in breakdown")
	}
}

// Helper function to check if string contains substring (local to avoid conflicts)
func threatContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && threatContainsHelper(s, substr)))
}

func threatContainsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func BenchmarkAnalyzeDeck(b *testing.B) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)
	deck := []string{"Musketeer", "Inferno Tower", "P.E.K.K.A", "Valkyrie", "The Log", "Zap", "Hog Rider", "Fireball"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeDeck(deck)
	}
}

func BenchmarkAnalyzeThreat(b *testing.B) {
	matrix := NewCounterMatrixWithDefaults()
	analyzer := NewThreatAnalyzer(matrix)
	deck := []string{"Musketeer", "Inferno Tower", "P.E.K.K.A", "Valkyrie"}
	threat := ThreatDefinition{
		Name:          "Mega Knight",
		Type:          ThreatTypeTank,
		Description:   "High HP jumping tank",
		MetaRelevance: 0.95,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.analyzeThreat(deck, threat)
	}
}
