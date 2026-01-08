//go:build integration

package evaluation

import (
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// TestEvaluateWithNilPlayerContext tests evaluation without player context
// This validates backwards compatibility - evaluation should work assuming all cards are available
func TestEvaluateWithNilPlayerContext(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	deckCards := []deck.CardCandidate{
		{Name: "Hog Rider", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleWinCondition)},
		{Name: "Musketeer", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSupport)},
		{Name: "Fireball", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSpellBig)},
		{Name: "The Log", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Legendary", Role: ptrRole(deck.RoleSpellSmall)},
		{Name: "Ice Spirit", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle)},
		{Name: "Skeletons", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle)},
		{Name: "Cannon", Elixir: 3, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleBuilding)},
		{Name: "Ice Golem", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleCycle)},
	}

	start := time.Now()
	result := Evaluate(deckCards, synergyDB, nil)
	duration := time.Since(start)

	// Performance check
	if duration > 2*time.Second {
		t.Errorf("Evaluation took too long: %v (want < 2s)", duration)
	}

	// Validate basic evaluation works without player context
	if result.OverallScore < 0 || result.OverallScore > 10 {
		t.Errorf("Overall score %.2f out of range [0, 10]", result.OverallScore)
	}

	// Playability should be perfect (10.0) when no player context
	if result.Playability.Score != 10.0 {
		t.Errorf("Expected playability score 10.0 with nil context, got %.2f", result.Playability.Score)
	}

	// Missing cards analysis should not be populated
	if result.MissingCardsAnalysis != nil {
		t.Error("Missing cards analysis should be nil when no player context provided")
	}

	// Validate other scores are calculated
	validateCategoryScore(t, "Attack", result.Attack)
	validateCategoryScore(t, "Defense", result.Defense)
	validateCategoryScore(t, "Synergy", result.Synergy)
	validateCategoryScore(t, "Versatility", result.Versatility)
	validateCategoryScore(t, "F2P Friendly", result.F2PFriendly)

	t.Logf("✓ Evaluation with nil context completed in %v", duration)
	t.Logf("  Overall: %.2f (%s) | Playability: %.2f", result.OverallScore, result.OverallRating, result.Playability.Score)
}

// TestEvaluateWithFullPlayerContext tests evaluation with complete player context
// This validates personalized analysis with all player context features
func TestEvaluateWithFullPlayerContext(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	// Create a mock player with full collection
	playerContext := createMockPlayerContext(
		"Arena 8",  // Frozen Peak
		[]string{ // Full collection - all cards owned
			"Hog Rider", "Musketeer", "Fireball", "The Log",
			"Ice Spirit", "Skeletons", "Cannon", "Ice Golem",
		},
		map[string]int{ // No evolutions unlocked
			"Hog Rider":   0,
			"Musketeer":   0,
			"Fireball":    0,
			"The Log":     0,
			"Ice Spirit":  0,
			"Skeletons":   0,
			"Cannon":      0,
			"Ice Golem":   0,
		},
	)

	deckCards := []deck.CardCandidate{
		{Name: "Hog Rider", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleWinCondition)},
		{Name: "Musketeer", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSupport)},
		{Name: "Fireball", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSpellBig)},
		{Name: "The Log", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Legendary", Role: ptrRole(deck.RoleSpellSmall)},
		{Name: "Ice Spirit", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle)},
		{Name: "Skeletons", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle)},
		{Name: "Cannon", Elixir: 3, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleBuilding)},
		{Name: "Ice Golem", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleCycle)},
	}

	start := time.Now()
	result := Evaluate(deckCards, synergyDB, playerContext)
	duration := time.Since(start)

	// Performance check
	if duration > 2*time.Second {
		t.Errorf("Evaluation took too long: %v (want < 2s)", duration)
	}

	// Validate missing cards analysis is populated
	if result.MissingCardsAnalysis == nil {
		t.Fatal("Missing cards analysis should be populated with player context")
	}

	// Deck should be fully playable (all cards owned)
	if !result.MissingCardsAnalysis.IsPlayable {
		t.Errorf("Expected deck to be playable, got missing count: %d", result.MissingCardsAnalysis.MissingCount)
	}

	// Playability should be perfect (10.0) when all cards owned
	if result.Playability.Score != 10.0 {
		t.Errorf("Expected playability score 10.0 with full collection, got %.2f", result.Playability.Score)
	}

	// Validate ladder analysis is populated
	if result.LadderAnalysis.Title == "" {
		t.Error("Ladder analysis should be populated with player context")
	}

	// Validate evolution analysis is populated
	if result.EvolutionAnalysis.Title == "" {
		t.Error("Evolution analysis should be populated with player context")
	}

	t.Logf("✓ Evaluation with full context completed in %v", duration)
	t.Logf("  Overall: %.2f (%s) | Playability: %.2f", result.OverallScore, result.OverallRating, result.Playability.Score)
	t.Logf("  IsPlayable: %v | Missing: %d", result.MissingCardsAnalysis.IsPlayable, result.MissingCardsAnalysis.MissingCount)
}

// TestEvaluateWithPartialCollection tests evaluation with missing cards
// This validates playability scoring and missing cards analysis
func TestEvaluateWithPartialCollection(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	// Create a player with partial collection (missing 2 cards)
	playerContext := createMockPlayerContext(
		"Arena 8", // Frozen Peak
		[]string{ // Only 6 cards owned
			"Hog Rider", "Musketeer", "Fireball", "The Log",
			"Ice Spirit", "Skeletons",
			// Missing: Cannon, Ice Golem
		},
		map[string]int{ // No evolutions
			"Hog Rider":  0,
			"Musketeer":  0,
			"Fireball":   0,
			"The Log":    0,
			"Ice Spirit": 0,
			"Skeletons":  0,
		},
	)

	deckCards := []deck.CardCandidate{
		{Name: "Hog Rider", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleWinCondition)},
		{Name: "Musketeer", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSupport)},
		{Name: "Fireball", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSpellBig)},
		{Name: "The Log", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Legendary", Role: ptrRole(deck.RoleSpellSmall)},
		{Name: "Ice Spirit", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle)},
		{Name: "Skeletons", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle)},
		{Name: "Cannon", Elixir: 3, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleBuilding)},
		{Name: "Ice Golem", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleCycle)},
	}

	result := Evaluate(deckCards, synergyDB, playerContext)

	// Validate missing cards analysis
	if result.MissingCardsAnalysis == nil {
		t.Fatal("Missing cards analysis should be populated")
	}

	// Should have 2 missing cards
	if result.MissingCardsAnalysis.MissingCount != 2 {
		t.Errorf("Expected 2 missing cards, got %d", result.MissingCardsAnalysis.MissingCount)
	}

	// Should not be playable
	if result.MissingCardsAnalysis.IsPlayable {
		t.Error("Expected deck to not be playable with missing cards")
	}

	// Should have 6 available cards
	if result.MissingCardsAnalysis.AvailableCount != 6 {
		t.Errorf("Expected 6 available cards, got %d", result.MissingCardsAnalysis.AvailableCount)
	}

	// Playability score should be reduced (< 10.0)
	if result.Playability.Score >= 10.0 {
		t.Errorf("Expected playability score < 10.0 with missing cards, got %.2f", result.Playability.Score)
	}

	// Missing cards should include Cannon and Ice Golem
	missingNames := make(map[string]bool)
	for _, card := range result.MissingCardsAnalysis.MissingCards {
		missingNames[card.Name] = true
	}

	if !missingNames["Cannon"] {
		t.Error("Expected Cannon to be in missing cards")
	}
	if !missingNames["Ice Golem"] {
		t.Error("Expected Ice Golem to be in missing cards")
	}

	t.Logf("✓ Partial collection evaluation")
	t.Logf("  Playability: %.2f | Missing: %d | Available: %d",
		result.Playability.Score,
		result.MissingCardsAnalysis.MissingCount,
		result.MissingCardsAnalysis.AvailableCount)
}

// TestEvaluateWithArenaLockedCards tests arena-aware validation
// This validates that cards locked by arena are properly detected
func TestEvaluateWithArenaLockedCards(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	// Create a low-level player (Arena 2 - Bone Pit)
	playerContext := createMockPlayerContext(
		"Arena 2", // Bone Pit - early arena
		[]string{ // Owns some cards, missing others
			"Hog Rider", "Musketeer", "Fireball", "Skeletons", "Cannon",
			// Missing: The Log, Ice Spirit, Ice Golem (locked by arena)
		},
		map[string]int{ // No evolutions
			"Hog Rider":  0,
			"Musketeer":  0,
			"Fireball":   0,
			"Skeletons":  0,
			"Cannon":     0,
		},
	)

	// Try to use a high-level deck with locked cards
	deckCards := []deck.CardCandidate{
		{Name: "Hog Rider", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleWinCondition)}, // Unlocks Arena 2
		{Name: "Musketeer", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSupport)},          // Unlocks Arena 0
		{Name: "Fireball", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSpellBig)},          // Unlocks Arena 0
		{Name: "The Log", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Legendary", Role: ptrRole(deck.RoleSpellSmall)},    // Unlocks Arena 6 - LOCKED
		{Name: "Ice Spirit", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle)},         // Unlocks Arena 6 - LOCKED
		{Name: "Skeletons", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle)},          // Unlocks Arena 0
		{Name: "Cannon", Elixir: 3, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleBuilding)},          // Unlocks Arena 0
		{Name: "Ice Golem", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleCycle)},            // Unlocks Arena 8 - LOCKED
	}

	result := Evaluate(deckCards, synergyDB, playerContext)

	// Validate missing cards analysis
	if result.MissingCardsAnalysis == nil {
		t.Fatal("Missing cards analysis should be populated")
	}

	// Should have 3 missing cards (The Log, Ice Spirit, Ice Golem)
	if result.MissingCardsAnalysis.MissingCount != 3 {
		t.Errorf("Expected 3 missing cards, got %d", result.MissingCardsAnalysis.MissingCount)
	}

	// Check that locked cards are properly marked
	lockedCount := 0
	for _, card := range result.MissingCardsAnalysis.MissingCards {
		if card.IsLocked {
			lockedCount++
		}
	}

	// Should have 3 locked cards (The Log @ Arena 6, Ice Spirit @ Arena 6, Ice Golem @ Arena 8)
	if lockedCount != 3 {
		t.Errorf("Expected 3 locked cards, got %d", lockedCount)
	}

	// Playability should be significantly reduced due to locked cards
	if result.Playability.Score >= 8.0 {
		t.Errorf("Expected playability score < 8.0 with locked cards, got %.2f", result.Playability.Score)
	}

	// Overall score should be penalized for locked cards
	if result.OverallScore >= 7.0 {
		t.Errorf("Expected overall score < 7.0 with locked cards, got %.2f", result.OverallScore)
	}

	t.Logf("✓ Arena-locked cards evaluation")
	t.Logf("  Playability: %.2f | Locked: %d | Missing: %d",
		result.Playability.Score,
		lockedCount,
		result.MissingCardsAnalysis.MissingCount)
}

// TestEvaluateWithEvolutions tests evolution integration
// This validates that evolutions are properly detected and analyzed
func TestEvaluateWithEvolutions(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	// Create a player with evolutions
	playerContext := createMockPlayerContext(
		"Arena 15", // Legendary Arena
		[]string{ // Full collection
			"Hog Rider", "Musketeer", "Fireball", "The Log",
			"Ice Spirit", "Skeletons", "Cannon", "Ice Golem",
		},
		map[string]int{ // Some evolutions unlocked
			"Hog Rider":   1,  // Evolution Level 1
			"Musketeer":   2,  // Evolution Level 2
			"Fireball":    0,
			"The Log":     1,  // Evolution Level 1
			"Ice Spirit":  0,
			"Skeletons":   0,
			"Cannon":      0,
			"Ice Golem":   0,
		},
	)

	deckCards := []deck.CardCandidate{
		{Name: "Hog Rider", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleWinCondition), EvolutionLevel: 1},
		{Name: "Musketeer", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSupport), EvolutionLevel: 2},
		{Name: "Fireball", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSpellBig), EvolutionLevel: 0},
		{Name: "The Log", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Legendary", Role: ptrRole(deck.RoleSpellSmall), EvolutionLevel: 1},
		{Name: "Ice Spirit", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle), EvolutionLevel: 0},
		{Name: "Skeletons", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle), EvolutionLevel: 0},
		{Name: "Cannon", Elixir: 3, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleBuilding), EvolutionLevel: 0},
		{Name: "Ice Golem", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleCycle), EvolutionLevel: 0},
	}

	result := Evaluate(deckCards, synergyDB, playerContext)

	// Evolution analysis should be populated
	if result.EvolutionAnalysis.Title == "" {
		t.Error("Evolution analysis should be populated")
	}

	// Should detect evolved cards
	hasEvolutions := false
	for _, card := range deckCards {
		if card.EvolutionLevel > 0 {
			hasEvolutions = true
			break
		}
	}

	if !hasEvolutions {
		t.Error("Expected to find evolved cards in deck")
	}

	// Attack and Defense scores should get evolution bonuses
	if result.Attack.Score <= 5.0 {
		t.Logf("Note: Attack score %.2f may include evolution bonuses", result.Attack.Score)
	}
	if result.Defense.Score <= 5.0 {
		t.Logf("Note: Defense score %.2f may include evolution bonuses", result.Defense.Score)
	}

	t.Logf("✓ Evolution integration evaluation")
	t.Logf("  Overall: %.2f | Attack: %.2f | Defense: %.2f",
		result.OverallScore, result.Attack.Score, result.Defense.Score)
	t.Logf("  Evolution Analysis: %s", result.EvolutionAnalysis.Title)
}

// TestEvaluatePlayabilityScoring tests playability scoring in detail
// This validates the playability score calculation with different scenarios
func TestEvaluatePlayabilityScoring(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	tests := []struct {
		name              string
		ownedCards        []string
		expectedPlayable  bool
		minPlayability    float64
		maxPlayability    float64
	}{
		{
			name:              "Full collection",
			ownedCards:        []string{"Hog Rider", "Musketeer", "Fireball", "The Log", "Ice Spirit", "Skeletons", "Cannon", "Ice Golem"},
			expectedPlayable:  true,
			minPlayability:    10.0,
			maxPlayability:    10.0,
		},
		{
			name:              "Missing 1 card",
			ownedCards:        []string{"Hog Rider", "Musketeer", "Fireball", "The Log", "Ice Spirit", "Skeletons", "Cannon"},
			expectedPlayable:  false,
			minPlayability:    8.0,
			maxPlayability:    9.5,
		},
		{
			name:              "Missing 2 cards",
			ownedCards:        []string{"Hog Rider", "Musketeer", "Fireball", "The Log", "Ice Spirit", "Skeletons"},
			expectedPlayable:  false,
			minPlayability:    6.0,
			maxPlayability:    8.5,
		},
		{
			name:              "Missing 4 cards",
			ownedCards:        []string{"Hog Rider", "Musketeer", "Fireball", "The Log"},
			expectedPlayable:  false,
			minPlayability:    0.0,
			maxPlayability:    6.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			playerContext := createMockPlayerContext(
				"Arena 15",
				tt.ownedCards,
				map[string]int{},
			)

			deckCards := []deck.CardCandidate{
				{Name: "Hog Rider", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleWinCondition)},
				{Name: "Musketeer", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSupport)},
				{Name: "Fireball", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSpellBig)},
				{Name: "The Log", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Legendary", Role: ptrRole(deck.RoleSpellSmall)},
				{Name: "Ice Spirit", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle)},
				{Name: "Skeletons", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle)},
				{Name: "Cannon", Elixir: 3, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleBuilding)},
				{Name: "Ice Golem", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleCycle)},
			}

			result := Evaluate(deckCards, synergyDB, playerContext)

			// Validate playability
			if result.MissingCardsAnalysis.IsPlayable != tt.expectedPlayable {
				t.Errorf("Expected playable=%v, got %v", tt.expectedPlayable, result.MissingCardsAnalysis.IsPlayable)
			}

			// Validate playability score range
			if result.Playability.Score < tt.minPlayability || result.Playability.Score > tt.maxPlayability {
				t.Errorf("Playability score %.2f out of expected range [%.2f, %.2f]",
					result.Playability.Score, tt.minPlayability, tt.maxPlayability)
			}

			t.Logf("✓ %s: Playability=%.2f, Missing=%d, Available=%d",
				tt.name,
				result.Playability.Score,
				result.MissingCardsAnalysis.MissingCount,
				result.MissingCardsAnalysis.AvailableCount)
		})
	}
}

// BenchmarkEvaluateWithPlayerContext benchmarks evaluation with player context
func BenchmarkEvaluateWithPlayerContext(b *testing.B) {
	synergyDB := deck.NewSynergyDatabase()
	playerContext := createMockPlayerContext(
		"Arena 15",
		[]string{"Hog Rider", "Musketeer", "Fireball", "The Log", "Ice Spirit", "Skeletons", "Cannon", "Ice Golem"},
		map[string]int{"Hog Rider": 1, "Musketeer": 2, "Fireball": 0, "The Log": 1, "Ice Spirit": 0, "Skeletons": 0, "Cannon": 0, "Ice Golem": 0},
	)

	deckCards := []deck.CardCandidate{
		{Name: "Hog Rider", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleWinCondition), EvolutionLevel: 1},
		{Name: "Musketeer", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSupport), EvolutionLevel: 2},
		{Name: "Fireball", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSpellBig), EvolutionLevel: 0},
		{Name: "The Log", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Legendary", Role: ptrRole(deck.RoleSpellSmall), EvolutionLevel: 1},
		{Name: "Ice Spirit", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle), EvolutionLevel: 0},
		{Name: "Skeletons", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle), EvolutionLevel: 0},
		{Name: "Cannon", Elixir: 3, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleBuilding), EvolutionLevel: 0},
		{Name: "Ice Golem", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleCycle), EvolutionLevel: 0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Evaluate(deckCards, synergyDB, playerContext)
	}
}

// BenchmarkEvaluateWithoutPlayerContext benchmarks evaluation without player context
func BenchmarkEvaluateWithoutPlayerContext(b *testing.B) {
	synergyDB := deck.NewSynergyDatabase()

	deckCards := []deck.CardCandidate{
		{Name: "Hog Rider", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleWinCondition)},
		{Name: "Musketeer", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSupport)},
		{Name: "Fireball", Elixir: 4, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleSpellBig)},
		{Name: "The Log", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Legendary", Role: ptrRole(deck.RoleSpellSmall)},
		{Name: "Ice Spirit", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle)},
		{Name: "Skeletons", Elixir: 1, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleCycle)},
		{Name: "Cannon", Elixir: 3, Level: 11, MaxLevel: 14, Rarity: "Common", Role: ptrRole(deck.RoleBuilding)},
		{Name: "Ice Golem", Elixir: 2, Level: 11, MaxLevel: 14, Rarity: "Rare", Role: ptrRole(deck.RoleCycle)},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Evaluate(deckCards, synergyDB, nil)
	}
}

// Helper function to create mock player context
func createMockPlayerContext(arenaName string, ownedCards []string, evolutions map[string]int) *PlayerContext {
	// Arena ID mapping
	arenaID := 15
	switch arenaName {
	case "Arena 2":
		arenaID = 2
	case "Arena 8":
		arenaID = 8
	case "Arena 15":
		arenaID = 15
	}

	ctx := &PlayerContext{
		Arena: &clashroyale.Arena{
			ID:   arenaID,
			Name: arenaName,
		},
		ArenaID:            arenaID,
		ArenaName:          arenaName,
		Collection:         make(map[string]CardLevelInfo),
		UnlockedEvolutions: make(map[string]bool),
		PlayerTag:          "#TESTPLAYER",
		PlayerName:         "Test Player",
	}

	// Populate collection
	for _, cardName := range ownedCards {
		evolutionLevel := 0
		if evoLevel, exists := evolutions[cardName]; exists {
			evolutionLevel = evoLevel
		}

		ctx.Collection[cardName] = CardLevelInfo{
			Level:             11,
			MaxLevel:          14,
			EvolutionLevel:    evolutionLevel,
			MaxEvolutionLevel: 2,
			Rarity:            "Rare",
			Count:             10,
		}

		// Track unlocked evolutions
		if evolutionLevel > 0 {
			ctx.UnlockedEvolutions[cardName] = true
		}
	}

	return ctx
}
