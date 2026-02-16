package deck

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// createBuilderTestCandidates creates a set of test candidates for builder tests.
func createBuilderTestCandidates() []*CardCandidate {
	winCon := RoleWinCondition
	spellBig := RoleSpellBig
	spellSmall := RoleSpellSmall
	support := RoleSupport
	building := RoleBuilding
	cycle := RoleCycle

	return []*CardCandidate{
		// Win conditions
		{
			Name: "Hog Rider", Elixir: 4, Role: &winCon, Level: 14, MaxLevel: 14, Rarity: "Rare",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 200, Hitpoints: 1600, Targets: "Buildings"},
		},
		{
			Name: "Giant", Elixir: 5, Role: &winCon, Level: 13, MaxLevel: 14, Rarity: "Rare",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 140, Hitpoints: 4000, Targets: "Buildings"},
		},
		{
			Name: "Balloon", Elixir: 5, Role: &winCon, Level: 14, MaxLevel: 14, Rarity: "Epic",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 200, Hitpoints: 1800, Targets: "Buildings"},
		},

		// Big spells
		{
			Name: "Fireball", Elixir: 4, Role: &spellBig, Level: 14, MaxLevel: 14, Rarity: "Rare",
			Stats: &clashroyale.CombatStats{Radius: 2.5},
		},
		{
			Name: "Poison", Elixir: 4, Role: &spellBig, Level: 13, MaxLevel: 14, Rarity: "Epic",
			Stats: &clashroyale.CombatStats{Radius: 3.5},
		},

		// Small spells
		{
			Name: "Zap", Elixir: 2, Role: &spellSmall, Level: 14, MaxLevel: 14, Rarity: "Common",
			Stats: &clashroyale.CombatStats{Radius: 2.5},
		},
		{
			Name: "The Log", Elixir: 2, Role: &spellSmall, Level: 14, MaxLevel: 14, Rarity: "Legendary",
			Stats: &clashroyale.CombatStats{Radius: 3.9},
		},
		{
			Name: "Arrows", Elixir: 3, Role: &spellSmall, Level: 14, MaxLevel: 14, Rarity: "Common",
			Stats: &clashroyale.CombatStats{Radius: 4.0},
		},

		// Support with air targeting
		{
			Name: "Musketeer", Elixir: 4, Role: &support, Level: 14, MaxLevel: 14, Rarity: "Rare",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 180, Hitpoints: 700, Targets: "Air & Ground", Range: 6.0},
		},
		{
			Name: "Electro Wizard", Elixir: 4, Role: &support, Level: 13, MaxLevel: 14, Rarity: "Legendary",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 120, Hitpoints: 600, Targets: "Air & Ground", Range: 5.0},
		},
		{
			Name: "Baby Dragon", Elixir: 4, Role: &support, Level: 14, MaxLevel: 14, Rarity: "Epic",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 130, Hitpoints: 1000, Targets: "Air & Ground", Radius: 1.0},
		},
		{
			Name: "Mega Minion", Elixir: 3, Role: &support, Level: 14, MaxLevel: 14, Rarity: "Rare",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 170, Hitpoints: 700, Targets: "Air & Ground"},
		},

		// Ground-only support
		{
			Name: "Valkyrie", Elixir: 4, Role: &support, Level: 14, MaxLevel: 14, Rarity: "Rare",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 150, Hitpoints: 1700, Targets: "Ground", Radius: 1.2},
		},
		{
			Name: "Mini P.E.K.K.A", Elixir: 4, Role: &support, Level: 14, MaxLevel: 14, Rarity: "Rare",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 300, Hitpoints: 1200, Targets: "Ground"},
		},
		{
			Name: "Dark Prince", Elixir: 4, Role: &support, Level: 13, MaxLevel: 14, Rarity: "Epic",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 180, Hitpoints: 1000, Targets: "Ground", Radius: 1.0},
		},

		// Buildings
		{
			Name: "Cannon", Elixir: 3, Role: &building, Level: 14, MaxLevel: 14, Rarity: "Common",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 160, Hitpoints: 900, Targets: "Ground"},
		},
		{
			Name: "Tesla", Elixir: 4, Role: &building, Level: 14, MaxLevel: 14, Rarity: "Common",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 170, Hitpoints: 1000, Targets: "Air & Ground"},
		},
		{
			Name: "Inferno Tower", Elixir: 5, Role: &building, Level: 13, MaxLevel: 14, Rarity: "Rare",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 90, Hitpoints: 1400, Targets: "Air & Ground"},
		},

		// Cycle cards
		{
			Name: "Skeletons", Elixir: 1, Role: &cycle, Level: 14, MaxLevel: 14, Rarity: "Common",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 100, Hitpoints: 100, Targets: "Ground"},
		},
		{
			Name: "Ice Spirit", Elixir: 1, Role: &cycle, Level: 14, MaxLevel: 14, Rarity: "Common",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 40, Hitpoints: 200, Targets: "Air & Ground"},
		},
		{
			Name: "Ice Golem", Elixir: 2, Role: &cycle, Level: 14, MaxLevel: 14, Rarity: "Rare",
			Stats: &clashroyale.CombatStats{DamagePerSecond: 0, Hitpoints: 1000, Targets: "Buildings"},
		},
	}
}

func TestSynergyGraphBuilder_Build(t *testing.T) {
	candidates := createBuilderTestCandidates()
	synergyDB := NewSynergyDatabase()

	config := DefaultBuilderConfig(candidates, synergyDB)
	builder := NewSynergyGraphBuilder(config)

	deck, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if len(deck) != 8 {
		t.Errorf("expected 8 cards, got %d", len(deck))
	}

	// Check for uniqueness
	seen := make(map[string]bool)
	for _, card := range deck {
		if seen[card] {
			t.Errorf("duplicate card in deck: %s", card)
		}
		seen[card] = true
	}

	t.Logf("Synergy Graph Deck: %v", deck)
}

func TestSynergyGraphBuilder_Name(t *testing.T) {
	builder := NewSynergyGraphBuilder(BuilderConfig{})
	if builder.Name() != "synergy_graph" {
		t.Errorf("expected name 'synergy_graph', got %s", builder.Name())
	}
}

func TestConstraintSatisfactionBuilder_Build(t *testing.T) {
	candidates := createBuilderTestCandidates()
	synergyDB := NewSynergyDatabase()

	config := DefaultBuilderConfig(candidates, synergyDB)
	builder := NewConstraintSatisfactionBuilder(config)

	deck, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if len(deck) != 8 {
		t.Errorf("expected 8 cards, got %d", len(deck))
	}

	// Verify constraints are met
	hasWinCon := false
	hasSpell := false
	for _, cardName := range deck {
		for _, c := range candidates {
			if c.Name == cardName && c.Role != nil {
				if *c.Role == RoleWinCondition {
					hasWinCon = true
				}
				if *c.Role == RoleSpellBig || *c.Role == RoleSpellSmall {
					hasSpell = true
				}
			}
		}
	}

	if !hasWinCon {
		t.Error("deck missing win condition")
	}
	if !hasSpell {
		t.Error("deck missing spell")
	}

	t.Logf("Constraint Satisfaction Deck: %v", deck)
}

func TestRoleFirstBuilder_Build(t *testing.T) {
	candidates := createBuilderTestCandidates()
	synergyDB := NewSynergyDatabase()

	config := DefaultBuilderConfig(candidates, synergyDB)
	builder := NewRoleFirstBuilder(config)

	deck, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if len(deck) != 8 {
		t.Errorf("expected 8 cards, got %d", len(deck))
	}

	t.Logf("Role First Deck: %v", deck)
}

func TestCounterCentricBuilder_Build(t *testing.T) {
	candidates := createBuilderTestCandidates()
	synergyDB := NewSynergyDatabase()
	counterMatrix := NewCounterMatrixWithDefaults()

	config := DefaultBuilderConfig(candidates, synergyDB)
	config.CounterMatrix = counterMatrix
	builder := NewCounterCentricBuilder(config)

	deck, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if len(deck) != 8 {
		t.Errorf("expected 8 cards, got %d", len(deck))
	}

	t.Logf("Counter Centric Deck: %v", deck)
}

func TestMetaLearningBuilder_Build(t *testing.T) {
	candidates := createBuilderTestCandidates()
	synergyDB := NewSynergyDatabase()

	// Create sample decks for learning
	sampleDecks := [][]string{
		{"Hog Rider", "Musketeer", "Fireball", "Zap", "Ice Spirit", "Skeletons", "Cannon", "Ice Golem"},
		{"Hog Rider", "Valkyrie", "Fireball", "The Log", "Ice Spirit", "Musketeer", "Tesla", "Skeletons"},
		{"Giant", "Musketeer", "Fireball", "Zap", "Mega Minion", "Mini P.E.K.K.A", "Cannon", "Ice Spirit"},
	}

	coOccurrence := NewCoOccurrenceMatrix()
	coOccurrence.LearnFromDecks(sampleDecks)

	config := DefaultBuilderConfig(candidates, synergyDB)
	builder := NewMetaLearningBuilder(config, coOccurrence)

	deck, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	if len(deck) != 8 {
		t.Errorf("expected 8 cards, got %d", len(deck))
	}

	t.Logf("Meta Learning Deck: %v", deck)
}

func TestArchetypeFreeScorer_Score(t *testing.T) {
	candidates := createBuilderTestCandidates()
	synergyDB := NewSynergyDatabase()
	counterMatrix := NewCounterMatrixWithDefaults()

	scorer := NewArchetypeFreeScorer(synergyDB, counterMatrix)

	// Create a deck from candidates
	deck := []CardCandidate{}
	for i := 0; i < 8 && i < len(candidates); i++ {
		deck = append(deck, *candidates[i])
	}

	score := scorer.Score(deck)

	if score < 0 || score > 1 {
		t.Errorf("score out of range [0,1]: %f", score)
	}

	t.Logf("Archetype-Free Score: %.3f", score)
}

func TestCoOccurrenceMatrix_LearnFromDecks(t *testing.T) {
	matrix := NewCoOccurrenceMatrix()

	// Sample decks where Hog Rider always appears with Fireball
	decks := [][]string{
		{"Hog Rider", "Fireball", "Card3", "Card4", "Card5", "Card6", "Card7", "Card8"},
		{"Hog Rider", "Fireball", "Card9", "Card10", "Card11", "Card12", "Card13", "Card14"},
		{"Other", "Fireball", "Card3", "Card4", "Card5", "Card6", "Card7", "Card8"},
	}

	matrix.LearnFromDecks(decks)

	// P(Fireball | Hog Rider) should be 1.0 (both decks with Hog have Fireball)
	prob := matrix.GetProbability("Hog Rider", "Fireball")
	if prob != 1.0 {
		t.Errorf("expected P(Fireball|Hog Rider) = 1.0, got %f", prob)
	}

	// P(Hog Rider | Fireball) should be 2/3 (2 of 3 Fireball decks have Hog)
	prob = matrix.GetProbability("Fireball", "Hog Rider")
	expected := 2.0 / 3.0
	if prob < expected-0.01 || prob > expected+0.01 {
		t.Errorf("expected P(Hog Rider|Fireball) = %.3f, got %.3f", expected, prob)
	}
}

func TestCreateBuilder(t *testing.T) {
	candidates := createBuilderTestCandidates()
	synergyDB := NewSynergyDatabase()
	config := DefaultBuilderConfig(candidates, synergyDB)
	config.CounterMatrix = NewCounterMatrixWithDefaults()

	testCases := []struct {
		builderType BuilderType
		expectError bool
	}{
		{BuilderSynergyGraph, false},
		{BuilderConstraintSatisfaction, false},
		{BuilderRoleFirst, false},
		{BuilderCounterCentric, false},
		{BuilderMetaLearning, false},
		{BuilderType("unknown"), true},
	}

	for _, tc := range testCases {
		builder, err := CreateBuilder(tc.builderType, config)
		if tc.expectError && err == nil {
			t.Errorf("expected error for builder type %s", tc.builderType)
		}
		if !tc.expectError && err != nil {
			t.Errorf("unexpected error for builder type %s: %v", tc.builderType, err)
		}
		if !tc.expectError && builder == nil {
			t.Errorf("expected builder for type %s", tc.builderType)
		}
	}
}

func TestBuildMultiple(t *testing.T) {
	candidates := createBuilderTestCandidates()
	synergyDB := NewSynergyDatabase()
	config := DefaultBuilderConfig(candidates, synergyDB)
	config.CounterMatrix = NewCounterMatrixWithDefaults()

	results := BuildMultiple(config,
		BuilderSynergyGraph,
		BuilderConstraintSatisfaction,
		BuilderRoleFirst,
		BuilderCounterCentric,
	)

	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}

	for bt, deck := range results {
		if len(deck) != 8 {
			t.Errorf("builder %s produced deck with %d cards", bt, len(deck))
		}
		t.Logf("%s: %v", bt, deck)
	}
}

func TestSynergyGraphBuilder_InsufficientCandidates(t *testing.T) {
	// Only 5 candidates - not enough for an 8-card deck
	candidates := createBuilderTestCandidates()[:5]
	synergyDB := NewSynergyDatabase()
	config := DefaultBuilderConfig(candidates, synergyDB)
	builder := NewSynergyGraphBuilder(config)

	_, err := builder.Build()
	if err == nil {
		t.Error("expected error for insufficient candidates")
	}
}

func TestConstraintSatisfactionBuilder_MissingDependency(t *testing.T) {
	config := BuilderConfig{
		Candidates: nil,
	}
	builder := NewConstraintSatisfactionBuilder(config)

	_, err := builder.Build()
	if err == nil {
		t.Error("expected error for nil candidates")
	}
}

func BenchmarkSynergyGraphBuilder(b *testing.B) {
	candidates := createBuilderTestCandidates()
	synergyDB := NewSynergyDatabase()
	config := DefaultBuilderConfig(candidates, synergyDB)
	builder := NewSynergyGraphBuilder(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = builder.Build()
	}
}

func BenchmarkConstraintSatisfactionBuilder(b *testing.B) {
	candidates := createBuilderTestCandidates()
	synergyDB := NewSynergyDatabase()
	config := DefaultBuilderConfig(candidates, synergyDB)
	builder := NewConstraintSatisfactionBuilder(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = builder.Build()
	}
}

func BenchmarkRoleFirstBuilder(b *testing.B) {
	candidates := createBuilderTestCandidates()
	synergyDB := NewSynergyDatabase()
	config := DefaultBuilderConfig(candidates, synergyDB)
	builder := NewRoleFirstBuilder(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = builder.Build()
	}
}

func BenchmarkCounterCentricBuilder(b *testing.B) {
	candidates := createBuilderTestCandidates()
	synergyDB := NewSynergyDatabase()
	counterMatrix := NewCounterMatrixWithDefaults()
	config := DefaultBuilderConfig(candidates, synergyDB)
	config.CounterMatrix = counterMatrix
	builder := NewCounterCentricBuilder(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = builder.Build()
	}
}
