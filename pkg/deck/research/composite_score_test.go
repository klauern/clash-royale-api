package research

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func testRolePtr(r deck.CardRole) *deck.CardRole { return &r }

func testDeck() []deck.CardCandidate {
	return []deck.CardCandidate{
		{Name: "Hog Rider", Level: 14, MaxLevel: 15, Elixir: 4, Role: testRolePtr(deck.RoleWinCondition), Stats: &clashroyale.CombatStats{DamagePerSecond: 150, Targets: "Buildings"}},
		{Name: "Musketeer", Level: 13, MaxLevel: 15, Elixir: 4, Role: testRolePtr(deck.RoleSupport), Stats: &clashroyale.CombatStats{DamagePerSecond: 181, Targets: "Air & Ground"}},
		{Name: "Fireball", Level: 13, MaxLevel: 15, Elixir: 4, Role: testRolePtr(deck.RoleSpellBig), Stats: &clashroyale.CombatStats{Radius: 2.5}},
		{Name: "The Log", Level: 13, MaxLevel: 15, Elixir: 2, Role: testRolePtr(deck.RoleSpellSmall), Stats: &clashroyale.CombatStats{Radius: 1.8}},
		{Name: "Cannon", Level: 13, MaxLevel: 15, Elixir: 3, Role: testRolePtr(deck.RoleBuilding), Stats: &clashroyale.CombatStats{DamagePerSecond: 140, Targets: "Ground"}},
		{Name: "Knight", Level: 13, MaxLevel: 15, Elixir: 3, Role: testRolePtr(deck.RoleSupport), Stats: &clashroyale.CombatStats{DamagePerSecond: 160, Targets: "Ground"}},
		{Name: "Skeletons", Level: 14, MaxLevel: 15, Elixir: 1, Role: testRolePtr(deck.RoleCycle), Stats: &clashroyale.CombatStats{Targets: "Ground"}},
		{Name: "Ice Spirit", Level: 14, MaxLevel: 15, Elixir: 1, Role: testRolePtr(deck.RoleCycle), Stats: &clashroyale.CombatStats{Targets: "Air & Ground"}},
	}
}

func TestScoreDeckCompositeBounds(t *testing.T) {
	cards := testDeck()
	cfg := DefaultConstraintConfig()
	metrics := ScoreDeckComposite(cards, deck.NewSynergyDatabase(), cfg)

	if metrics.Composite < 0 || metrics.Composite > 1 {
		t.Fatalf("composite out of range: %f", metrics.Composite)
	}
	if metrics.Synergy < 0 || metrics.Synergy > 1 {
		t.Fatalf("synergy out of range: %f", metrics.Synergy)
	}
	if metrics.Coverage < 0 || metrics.Coverage > 1 {
		t.Fatalf("coverage out of range: %f", metrics.Coverage)
	}
}

func TestScoreDeckCompositeHasNoConstraintViolationsForValidDeck(t *testing.T) {
	cards := testDeck()
	cfg := DefaultConstraintConfig()
	metrics := ScoreDeckComposite(cards, deck.NewSynergyDatabase(), cfg)
	if len(metrics.ConstraintViolations) != 0 {
		t.Fatalf("expected no violations, got %v", metrics.ConstraintViolations)
	}
}
