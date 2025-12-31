package scoring

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestNewSynergyScorer(t *testing.T) {
	db := deck.NewSynergyDatabase()
	scorer := NewSynergyScorer(NewDeckSynergyDatabase(db))

	if scorer == nil {
		t.Fatal("NewSynergyScorer returned nil")
	}

	// Verify default values
	if scorer.synergyWeight != 0.15 {
		t.Errorf("expected synergyWeight 0.15, got %f", scorer.synergyWeight)
	}
}

func TestNewSynergyScorerWithConfig(t *testing.T) {
	db := deck.NewSynergyDatabase()
	config := SynergyScorerConfig{
		SynergyDatabase: NewDeckSynergyDatabase(db),
		SynergyWeight:   0.25,
	}

	scorer := NewSynergyScorerWithConfig(config)

	if scorer.synergyWeight != 0.25 {
		t.Errorf("expected synergyWeight 0.25, got %f", scorer.synergyWeight)
	}
}

func TestSynergyScorer_Score_NoSynergyDatabase(t *testing.T) {
	scorer := NewSynergyScorer(nil)
	config := DefaultScoringConfig()

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name: "Archers",
		Role: &role,
	}

	score := scorer.Score(candidate, config)

	if score != 0.0 {
		t.Errorf("expected score 0.0 with no synergy database, got %f", score)
	}
}

func TestSynergyScorer_Score_EmptyDeck(t *testing.T) {
	db := deck.NewSynergyDatabase()
	scorer := NewSynergyScorer(NewDeckSynergyDatabase(db))
	config := DefaultScoringConfig()
	config.CurrentDeck = []CardCandidate{} // Empty deck

	role := deck.RoleSupport
	candidate := CardCandidate{
		Name: "Archers",
		Role: &role,
	}

	score := scorer.Score(candidate, config)

	if score != 0.0 {
		t.Errorf("expected score 0.0 with empty deck, got %f", score)
	}
}

func TestSynergyScorer_Score_WithSynergy(t *testing.T) {
	db := deck.NewSynergyDatabase()
	scorer := NewSynergyScorer(NewDeckSynergyDatabase(db))
	config := DefaultScoringConfig()

	// Set up a deck with Giant
	role := deck.RoleWinCondition
	currentDeck := []CardCandidate{
		{Name: "Giant", Role: &role},
	}
	config.CurrentDeck = currentDeck

	// Test Witch - has synergy with Giant
	candidate := CardCandidate{
		Name: "Witch",
		Role: &role,
	}

	score := scorer.Score(candidate, config)

	// Witch + Giant should have positive synergy
	if score <= 0.0 {
		t.Errorf("expected positive synergy score for Witch+Giant, got %f", score)
	}
}

func TestSynergyScorer_Score_NoSynergy(t *testing.T) {
	db := deck.NewSynergyDatabase()
	scorer := NewSynergyScorer(NewDeckSynergyDatabase(db))
	config := DefaultScoringConfig()

	// Set up a deck with Giant
	role := deck.RoleWinCondition
	currentDeck := []CardCandidate{
		{Name: "Giant", Role: &role},
	}
	config.CurrentDeck = currentDeck

	// Test a card with no known synergy with Giant
	// Using a made-up card name that won't be in the synergy database
	candidate := CardCandidate{
		Name: "Unknown Card",
		Role: &role,
	}

	score := scorer.Score(candidate, config)

	if score != 0.0 {
		t.Errorf("expected score 0.0 with no synergy, got %f", score)
	}
}

func TestSynergyScorer_Score_MultipleDeckCards(t *testing.T) {
	db := deck.NewSynergyDatabase()
	scorer := NewSynergyScorer(NewDeckSynergyDatabase(db))
	config := DefaultScoringConfig()

	// Set up a deck with multiple cards
	roleWin := deck.RoleWinCondition
	roleSupport := deck.RoleSupport
	currentDeck := []CardCandidate{
		{Name: "Giant", Role: &roleWin},
		{Name: "Witch", Role: &roleSupport},
	}
	config.CurrentDeck = currentDeck

	// Test Sparky - has synergy with Giant (tank + splash)
	candidate := CardCandidate{
		Name: "Sparky",
		Role: &roleWin,
	}

	score := scorer.Score(candidate, config)

	// Sparky should have synergy with at least one card
	if score <= 0.0 {
		t.Errorf("expected positive synergy score for Sparky, got %f", score)
	}
}

func TestSynergyScorer_CalculateSynergyScore(t *testing.T) {
	db := deck.NewSynergyDatabase()
	scorer := NewSynergyScorer(NewDeckSynergyDatabase(db))

	// Set up a deck with Giant
	role := deck.RoleWinCondition
	currentDeck := []CardCandidate{
		{Name: "Giant", Role: &role},
	}

	// Test Witch - has synergy with Giant
	score := scorer.CalculateSynergyScore("Witch", currentDeck)

	if score <= 0.0 {
		t.Errorf("expected positive synergy score for Witch+Giant, got %f", score)
	}
}

func TestSynergyScorer_CalculateSynergyScore_EmptyDeck(t *testing.T) {
	db := deck.NewSynergyDatabase()
	scorer := NewSynergyScorer(NewDeckSynergyDatabase(db))

	score := scorer.CalculateSynergyScore("Witch", []CardCandidate{})

	if score != 0.0 {
		t.Errorf("expected score 0.0 with empty deck, got %f", score)
	}
}

func TestSynergyScorer_SetSynergyDatabase(t *testing.T) {
	scorer := NewSynergyScorer(nil)

	db := deck.NewSynergyDatabase()
	scorer.SetSynergyDatabase(NewDeckSynergyDatabase(db))

	retrieved := scorer.GetSynergyDatabase()
	if retrieved == nil {
		t.Errorf("expected non-nil synergy database after set")
	}
}

func TestSynergyScorer_GetSynergyDatabase(t *testing.T) {
	db := deck.NewSynergyDatabase()
	scorer := NewSynergyScorer(NewDeckSynergyDatabase(db))

	retrieved := scorer.GetSynergyDatabase()
	if retrieved == nil {
		t.Errorf("expected non-nil synergy database")
	}
}

func TestSynergyScorer_SetSynergyWeight(t *testing.T) {
	scorer := NewSynergyScorer(nil)

	scorer.SetSynergyWeight(0.3)

	if scorer.GetSynergyWeight() != 0.3 {
		t.Errorf("expected synergyWeight 0.3, got %f", scorer.GetSynergyWeight())
	}
}

func TestSynergyScorer_GetSynergyWeight(t *testing.T) {
	scorer := NewSynergyScorer(nil)

	weight := scorer.GetSynergyWeight()
	if weight != 0.15 {
		t.Errorf("expected synergyWeight 0.15, got %f", weight)
	}
}

func TestSynergyScorerConfig_ZeroDefaults(t *testing.T) {
	// Test that zero values in config get replaced with defaults
	config := SynergyScorerConfig{
		SynergyWeight: 0, // Should default to 0.15
	}

	scorer := NewSynergyScorerWithConfig(config)

	if scorer.synergyWeight != 0.15 {
		t.Errorf("expected default synergyWeight 0.15, got %f", scorer.synergyWeight)
	}
}

func TestDeckSynergyDatabase_GetSynergy(t *testing.T) {
	db := deck.NewSynergyDatabase()
	adapter := NewDeckSynergyDatabase(db)

	// Giant + Witch is a known synergy (0.9)
	score := adapter.GetSynergy("Giant", "Witch")

	if score != 0.9 {
		t.Errorf("expected synergy 0.9 for Giant+Witch, got %f", score)
	}

	// Reverse order should give same result
	scoreReverse := adapter.GetSynergy("Witch", "Giant")

	if scoreReverse != 0.9 {
		t.Errorf("expected synergy 0.9 for Witch+Giant, got %f", scoreReverse)
	}
}

func TestDeckSynergyDatabase_GetSynergy_NoSynergy(t *testing.T) {
	db := deck.NewSynergyDatabase()
	adapter := NewDeckSynergyDatabase(db)

	// Random cards with no known synergy
	score := adapter.GetSynergy("Card1", "Card2")

	if score != 0.0 {
		t.Errorf("expected synergy 0.0 for unknown cards, got %f", score)
	}
}

func TestDeckSynergyDatabase_GetSynergy_NilDB(t *testing.T) {
	adapter := NewDeckSynergyDatabase(nil)

	score := adapter.GetSynergy("Giant", "Witch")

	if score != 0.0 {
		t.Errorf("expected synergy 0.0 with nil DB, got %f", score)
	}
}

func TestDeckSynergyDatabase_AnalyzeDeckSynergy(t *testing.T) {
	db := deck.NewSynergyDatabase()
	adapter := NewDeckSynergyDatabase(db)

	// Deck with known synergies
	deckNames := []string{"Giant", "Witch", "Mega Knight"}

	analysis := adapter.AnalyzeDeckSynergy(deckNames)

	if analysis == nil {
		t.Fatal("expected non-nil analysis")
	}

	// Should have at least one synergy
	if len(analysis.TopSynergies) == 0 {
		t.Error("expected at least one top synergy")
	}

	// Total score should be positive
	if analysis.TotalScore <= 0 {
		t.Errorf("expected positive total score, got %f", analysis.TotalScore)
	}
}

func TestDeckSynergyDatabase_AnalyzeDeckSynergy_NilDB(t *testing.T) {
	adapter := NewDeckSynergyDatabase(nil)

	analysis := adapter.AnalyzeDeckSynergy([]string{"Giant", "Witch"})

	if analysis == nil {
		t.Fatal("expected non-nil analysis")
	}

	// With nil DB, should return empty analysis
	if analysis.TotalScore != 0 {
		t.Errorf("expected total score 0.0 with nil DB, got %f", analysis.TotalScore)
	}
}

func TestSynergyScorer_ConfigSynergyDatabase(t *testing.T) {
	db := deck.NewSynergyDatabase()
	scorer := NewSynergyScorer(nil) // No default DB

	config := DefaultScoringConfig()
	config.SynergyDatabase = NewDeckSynergyDatabase(db)

	role := deck.RoleWinCondition
	currentDeck := []CardCandidate{
		{Name: "Giant", Role: &role},
	}
	config.CurrentDeck = currentDeck

	candidate := CardCandidate{
		Name: "Witch",
		Role: &role,
	}

	score := scorer.Score(candidate, config)

	if score <= 0.0 {
		t.Errorf("expected positive score with config DB, got %f", score)
	}
}

func TestSynergyScorer_Symmetry(t *testing.T) {
	// Test that synergy scoring is symmetric
	db := deck.NewSynergyDatabase()
	scorer := NewSynergyScorer(NewDeckSynergyDatabase(db))
	config := DefaultScoringConfig()

	role := deck.RoleWinCondition

	// Test 1: Deck has Giant, candidate is Witch
	currentDeck1 := []CardCandidate{{Name: "Giant", Role: &role}}
	config.CurrentDeck = currentDeck1
	candidate := CardCandidate{Name: "Witch", Role: &role}
	score1 := scorer.Score(candidate, config)

	// Test 2: Deck has Witch, candidate is Giant
	currentDeck2 := []CardCandidate{{Name: "Witch", Role: &role}}
	config.CurrentDeck = currentDeck2
	candidate = CardCandidate{Name: "Giant", Role: &role}
	score2 := scorer.Score(candidate, config)

	// Scores should be equal (synergy is bidirectional)
	if score1 != score2 {
		t.Errorf("synergy should be symmetric: score1=%f, score2=%f", score1, score2)
	}
}

func TestNewDeckSynergyDatabase(t *testing.T) {
	db := deck.NewSynergyDatabase()
	adapter := NewDeckSynergyDatabase(db)

	if adapter == nil {
		t.Fatal("NewDeckSynergyDatabase returned nil")
	}

	if adapter.db != db {
		t.Error("synergy database not set correctly")
	}
}

func TestSynergyScorer_ThreadSafety(t *testing.T) {
	db := deck.NewSynergyDatabase()
	scorer := NewSynergyScorer(NewDeckSynergyDatabase(db))

	// This test just verifies that concurrent access doesn't cause data races
	// It should be run with -race flag for full verification
	done := make(chan bool)

	go func() {
		scorer.GetSynergyDatabase()
		done <- true
	}()

	go func() {
		scorer.GetSynergyWeight()
		done <- true
	}()

	<-done
	<-done
	// If we get here without panic/deadlock, basic concurrent access works
}
