// Package deck provides tests for archetype coherence scoring
package deck

import (
	"testing"
)

// Helper to create a card candidate
func makeCard(name string, elixir int, role CardRole) CardCandidate {
	return CardCandidate{
		Name:   name,
		Elixir: elixir,
		Role:   &role,
	}
}

// TestLoadCoherenceScorer tests loading the coherence scorer configuration
func TestLoadCoherenceScorer(t *testing.T) {
	// Load with embedded defaults
	scorer, err := LoadCoherenceScorer("")
	if err != nil {
		t.Fatalf("Failed to load coherence scorer: %v", err)
	}

	if scorer == nil {
		t.Fatal("Scorer should not be nil")
	}

	if len(scorer.config.Archetypes) == 0 {
		t.Error("Expected at least one archetype")
	}
}

// TestCoherenceHogCycle tests that a pure Hog Cycle deck scores high on coherence
func TestCoherenceHogCycle(t *testing.T) {
	scorer, _ := LoadCoherenceScorer("")

	// Pure Hog 2.6 Cycle deck
	cards := []CardCandidate{
		makeCard("Hog Rider", 4, RoleWinCondition),
		makeCard("Musketeer", 4, RoleSupport),
		makeCard("Ice Golem", 2, RoleSupport),
		makeCard("Cannon", 3, RoleBuilding),
		makeCard("Skeletons", 1, RoleCycle),
		makeCard("Ice Spirit", 1, RoleCycle),
		makeCard("Fireball", 4, RoleSpellBig),
		makeCard("Log", 2, RoleSpellSmall),
	}

	result := scorer.AnalyzeCoherence(cards, StrategyCycle)

	// Should detect as cycle archetype
	if result.PrimaryArchetype != "cycle" {
		t.Errorf("Expected primary archetype 'cycle', got '%s'", result.PrimaryArchetype)
	}

	// Should have high coherence (at least 0.7)
	if result.CoherenceScore < 0.7 {
		t.Errorf("Expected high coherence score >= 0.7, got %.2f", result.CoherenceScore)
	}

	// Should have multiple cycle cards
	if result.CycleCardCount < 2 {
		t.Errorf("Expected at least 2 cycle cards, got %d", result.CycleCardCount)
	}

	// Elixir should be in cycle range
	if result.AverageElixir < 2.4 || result.AverageElixir > 3.2 {
		t.Errorf("Expected average elixir in cycle range (2.4-3.2), got %.1f", result.AverageElixir)
	}

	// Should have no major violations
	majorViolations := 0
	for _, v := range result.Violations {
		if v.Severity > 0.2 {
			majorViolations++
		}
	}
	if majorViolations > 0 {
		t.Errorf("Expected no major violations for pure Hog Cycle, got %d", majorViolations)
	}
}

// TestCoherenceBeatdownPlusCycle tests that mixing beatdown and cycle deck scores low
func TestCoherenceBeatdownPlusCycle(t *testing.T) {
	scorer, _ := LoadCoherenceScorer("")

	// Mixed beatdown + cycle deck (Golem + Hog Rider - bad combination)
	cards := []CardCandidate{
		makeCard("Golem", 8, RoleWinCondition),
		makeCard("Hog Rider", 4, RoleWinCondition),
		makeCard("Night Witch", 4, RoleSupport),
		makeCard("Baby Dragon", 4, RoleSupport),
		makeCard("Skeletons", 1, RoleCycle),
		makeCard("Ice Spirit", 1, RoleCycle),
		makeCard("Lightning", 6, RoleSpellBig),
		makeCard("Tombstone", 3, RoleBuilding),
	}

	result := scorer.AnalyzeCoherence(cards, StrategyBalanced)

	// Should detect anti-synergy violation
	foundConflict := false
	for _, v := range result.Violations {
		if v.Type == "anti_synergy" && v.Severity >= 0.25 {
			foundConflict = true
			break
		}
	}

	if !foundConflict {
		t.Error("Expected anti-synergy violation for beatdown+cycle mix")
	}

	// Coherence should be penalized
	if result.CoherenceScore > 0.7 {
		t.Errorf("Expected penalized coherence score <= 0.7 for conflicting deck, got %.2f", result.CoherenceScore)
	}
}

// TestCoherenceSiegeVsBeatdown tests siege vs beatdown anti-synergy
func TestCoherenceSiegeVsBeatdown(t *testing.T) {
	scorer, _ := LoadCoherenceScorer("")

	// Siege + beatdown mix (X-Bow + Golem - bad combination)
	cards := []CardCandidate{
		makeCard("X-Bow", 6, RoleWinCondition),
		makeCard("Golem", 8, RoleWinCondition),
		makeCard("Tesla", 4, RoleBuilding),
		makeCard("Night Witch", 4, RoleSupport),
		makeCard("Skeletons", 1, RoleCycle),
		makeCard("Ice Spirit", 1, RoleCycle),
		makeCard("Fireball", 4, RoleSpellBig),
		makeCard("Log", 2, RoleSpellSmall),
	}

	result := scorer.AnalyzeCoherence(cards, StrategyBalanced)

	// Should detect siege vs beatdown conflict
	foundConflict := false
	for _, v := range result.Violations {
		if v.Type == "anti_synergy" {
			// Check if the conflict mentions siege/beatdown
			for _, card := range v.Cards {
				if card == "X-Bow" || card == "Golem" {
					foundConflict = true
					break
				}
			}
		}
		if foundConflict {
			break
		}
	}

	if !foundConflict {
		t.Error("Expected anti-synergy violation for siege+beatdown mix")
	}
}

// TestCoherenceTooManyBuildings tests too many buildings penalty
func TestCoherenceTooManyBuildings(t *testing.T) {
	scorer, _ := LoadCoherenceScorer("")

	// Deck with 4 buildings (threshold is 3, so 4 should trigger violation)
	cards := []CardCandidate{
		makeCard("X-Bow", 6, RoleWinCondition),
		makeCard("Tesla", 4, RoleBuilding),
		makeCard("Cannon", 3, RoleBuilding),
		makeCard("Bomb Tower", 4, RoleBuilding),
		makeCard("Furnace", 4, RoleBuilding), // 4th building
		makeCard("Skeletons", 1, RoleCycle),
		makeCard("Fireball", 4, RoleSpellBig),
		makeCard("Log", 2, RoleSpellSmall),
	}

	result := scorer.AnalyzeCoherence(cards, StrategyBalanced)

	// Should have too many buildings violation
	foundViolation := false
	for _, v := range result.Violations {
		if v.Type == "composition" && v.Message == "More than 2 buildings makes deck too passive" {
			foundViolation = true
			break
		}
	}

	if !foundViolation {
		t.Errorf("Expected composition violation for too many buildings (got %d buildings)", result.BuildingCount)
	}
}

// TestCoherenceNoWinCondition tests no win condition penalty
func TestCoherenceNoWinCondition(t *testing.T) {
	scorer, _ := LoadCoherenceScorer("")

	// Deck with no clear win condition
	cards := []CardCandidate{
		makeCard("Knight", 3, RoleSupport),
		makeCard("Valkyrie", 4, RoleSupport),
		makeCard("Baby Dragon", 4, RoleSupport),
		makeCard("Musketeer", 4, RoleSupport),
		makeCard("Skeletons", 1, RoleCycle),
		makeCard("Ice Spirit", 1, RoleCycle),
		makeCard("Fireball", 4, RoleSpellBig),
		makeCard("Log", 2, RoleSpellSmall),
	}

	result := scorer.AnalyzeCoherence(cards, StrategyBalanced)

	// Should have no win condition violation
	foundViolation := false
	for _, v := range result.Violations {
		if v.Type == "composition" && v.Message == "Deck needs a clear tower-damage win condition" {
			foundViolation = true
			break
		}
	}

	if !foundViolation {
		t.Error("Expected composition violation for no win condition")
	}
}

// TestCoherenceBaitDeck tests that a proper bait deck scores well
func TestCoherenceBaitDeck(t *testing.T) {
	scorer, _ := LoadCoherenceScorer("")

	// Log Bait deck
	cards := []CardCandidate{
		makeCard("Goblin Barrel", 3, RoleWinCondition),
		makeCard("Princess", 3, RoleSupport),
		makeCard("Goblin Gang", 3, RoleSupport),
		makeCard("Knight", 3, RoleSupport),
		makeCard("Ice Spirit", 1, RoleCycle),
		makeCard("Log", 2, RoleSpellSmall),
		makeCard("Rocket", 6, RoleSpellBig),
		makeCard("Inferno Tower", 5, RoleBuilding),
	}

	result := scorer.AnalyzeCoherence(cards, StrategyBalanced)

	// Should detect as bait archetype
	if result.PrimaryArchetype != "bait" {
		t.Errorf("Expected primary archetype 'bait', got '%s'", result.PrimaryArchetype)
	}

	// Should have decent coherence (at least 0.6)
	if result.CoherenceScore < 0.6 {
		t.Errorf("Expected decent coherence score >= 0.6 for bait deck, got %.2f", result.CoherenceScore)
	}

	// Should have multiple bait cards
	if result.BaitCardCount < 2 {
		t.Errorf("Expected at least 2 bait cards, got %d", result.BaitCardCount)
	}
}

// TestCoherenceBridgeSpam tests bridge spam deck detection
func TestCoherenceBridgeSpam(t *testing.T) {
	scorer, _ := LoadCoherenceScorer("")

	// PEKKA Bridge Spam deck - using many fast threats
	cards := []CardCandidate{
		makeCard("P.E.K.K.A", 7, RoleWinCondition),
		makeCard("Battle Ram", 5, RoleWinCondition),
		makeCard("Bandit", 3, RoleWinCondition),
		makeCard("Royal Ghost", 3, RoleWinCondition), // 4th fast threat
		makeCard("Electro Wizard", 4, RoleSupport),
		makeCard("Dark Prince", 4, RoleSupport),
		makeCard("Poison", 4, RoleSpellBig),
		makeCard("Zap", 2, RoleSpellSmall),
	}

	result := scorer.AnalyzeCoherence(cards, StrategyAggro)

	// Should detect as bridge_spam archetype (or beatdown which is similar)
	// Note: PEKKA is in both beatdown and bridge_spam win conditions
	if result.PrimaryArchetype != "bridge_spam" && result.PrimaryArchetype != "beatdown" {
		t.Errorf("Expected primary archetype 'bridge_spam' or 'beatdown', got '%s'", result.PrimaryArchetype)
	}

	// Should have multiple fast threats (3+)
	if result.FastThreatCount < 3 {
		t.Errorf("Expected at least 3 fast threats, got %d", result.FastThreatCount)
	}

	// Should have decent coherence
	if result.CoherenceScore < 0.6 {
		t.Errorf("Expected decent coherence score >= 0.6, got %.2f", result.CoherenceScore)
	}
}

// TestGetCoherenceScore tests the quick score function
func TestGetCoherenceScore(t *testing.T) {
	scorer, _ := LoadCoherenceScorer("")

	cards := []CardCandidate{
		makeCard("Hog Rider", 4, RoleWinCondition),
		makeCard("Musketeer", 4, RoleSupport),
		makeCard("Ice Golem", 2, RoleSupport),
		makeCard("Cannon", 3, RoleBuilding),
		makeCard("Skeletons", 1, RoleCycle),
		makeCard("Ice Spirit", 1, RoleCycle),
		makeCard("Fireball", 4, RoleSpellBig),
		makeCard("Log", 2, RoleSpellSmall),
	}

	score := scorer.GetCoherenceScore(cards, StrategyCycle)

	if score < 0.0 || score > 1.0 {
		t.Errorf("Coherence score should be between 0.0 and 1.0, got %.2f", score)
	}

	if score < 0.7 {
		t.Errorf("Expected high coherence score >= 0.7 for good Hog Cycle, got %.2f", score)
	}
}

// TestCoherenceArchetypeRequirementsElixir tests elixir range validation
func TestCoherenceArchetypeRequirementsElixir(t *testing.T) {
	scorer, _ := LoadCoherenceScorer("")

	// High elixir deck being used as cycle strategy (should elicit warning)
	cards := []CardCandidate{
		makeCard("Golem", 8, RoleWinCondition),
		makeCard("Night Witch", 4, RoleSupport),
		makeCard("Baby Dragon", 4, RoleSupport),
		makeCard("Lumberjack", 4, RoleSupport),
		makeCard("Lightning", 6, RoleSpellBig),
		makeCard("Tornado", 3, RoleSpellSmall),
		makeCard("Tombstone", 3, RoleBuilding),
		makeCard("Skeletons", 1, RoleCycle),
	}

	result := scorer.AnalyzeCoherence(cards, StrategyCycle)

	// Should detect elixir mismatch
	foundElixirIssue := false
	for _, v := range result.Violations {
		if v.Type == "elixir" {
			foundElixirIssue = true
			break
		}
	}

	if !foundElixirIssue && result.AverageElixir > 4.0 {
		t.Errorf("Expected elixir violation for %.1f avg elixir with cycle strategy", result.AverageElixir)
	}
}
