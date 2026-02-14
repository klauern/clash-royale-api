package research

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func testPool() []deck.CardCandidate {
	cards := []deck.CardCandidate{
		{Name: "Hog Rider", Level: 14, MaxLevel: 15, Elixir: 4, Role: testRolePtr(deck.RoleWinCondition), Stats: &clashroyale.CombatStats{DamagePerSecond: 150, Targets: "Buildings"}},
		{Name: "Royal Giant", Level: 14, MaxLevel: 15, Elixir: 6, Role: testRolePtr(deck.RoleWinCondition), Stats: &clashroyale.CombatStats{DamagePerSecond: 180, Targets: "Buildings"}},
		{Name: "Mini P.E.K.K.A", Level: 13, MaxLevel: 15, Elixir: 4, Role: testRolePtr(deck.RoleSupport), Stats: &clashroyale.CombatStats{DamagePerSecond: 330, Targets: "Ground"}},
		{Name: "Musketeer", Level: 13, MaxLevel: 15, Elixir: 4, Role: testRolePtr(deck.RoleSupport), Stats: &clashroyale.CombatStats{DamagePerSecond: 181, Targets: "Air & Ground"}},
		{Name: "Baby Dragon", Level: 13, MaxLevel: 15, Elixir: 4, Role: testRolePtr(deck.RoleSupport), Stats: &clashroyale.CombatStats{Radius: 1.0, Targets: "Air & Ground"}},
		{Name: "Fireball", Level: 13, MaxLevel: 15, Elixir: 4, Role: testRolePtr(deck.RoleSpellBig), Stats: &clashroyale.CombatStats{Radius: 2.5}},
		{Name: "Poison", Level: 13, MaxLevel: 15, Elixir: 4, Role: testRolePtr(deck.RoleSpellBig), Stats: &clashroyale.CombatStats{Radius: 3.0}},
		{Name: "The Log", Level: 13, MaxLevel: 15, Elixir: 2, Role: testRolePtr(deck.RoleSpellSmall), Stats: &clashroyale.CombatStats{Radius: 1.8}},
		{Name: "Zap", Level: 13, MaxLevel: 15, Elixir: 2, Role: testRolePtr(deck.RoleSpellSmall), Stats: &clashroyale.CombatStats{Targets: "Air & Ground"}},
		{Name: "Cannon", Level: 13, MaxLevel: 15, Elixir: 3, Role: testRolePtr(deck.RoleBuilding), Stats: &clashroyale.CombatStats{DamagePerSecond: 140, Targets: "Ground"}},
		{Name: "Skeletons", Level: 14, MaxLevel: 15, Elixir: 1, Role: testRolePtr(deck.RoleCycle), Stats: &clashroyale.CombatStats{Targets: "Ground"}},
		{Name: "Ice Spirit", Level: 14, MaxLevel: 15, Elixir: 1, Role: testRolePtr(deck.RoleCycle), Stats: &clashroyale.CombatStats{Targets: "Ground"}},
	}
	return cards
}

func TestRoleFirstMethodDeterministicWithSeed(t *testing.T) {
	method := RoleFirstMethod{}
	cfg := MethodConfig{Seed: 7, TopN: 1}
	pool := testPool()

	first, err := method.Build(pool, cfg)
	if err != nil {
		t.Fatalf("first build failed: %v", err)
	}
	second, err := method.Build(pool, cfg)
	if err != nil {
		t.Fatalf("second build failed: %v", err)
	}
	if len(first.Deck) != len(second.Deck) {
		t.Fatalf("deck length mismatch: %d vs %d", len(first.Deck), len(second.Deck))
	}
	for i := range first.Deck {
		if first.Deck[i] != second.Deck[i] {
			t.Fatalf("non-deterministic deck at %d: %s vs %s", i, first.Deck[i], second.Deck[i])
		}
	}
}

func TestConstraintMethodDeterministicWithSeed(t *testing.T) {
	method := ConstraintMethod{}
	cfg := MethodConfig{Seed: 17, TopN: 1}
	pool := testPool()

	first, err := method.Build(pool, cfg)
	if err != nil {
		t.Fatalf("first build failed: %v", err)
	}
	second, err := method.Build(pool, cfg)
	if err != nil {
		t.Fatalf("second build failed: %v", err)
	}
	for i := range first.Deck {
		if first.Deck[i] != second.Deck[i] {
			t.Fatalf("non-deterministic deck at %d: %s vs %s", i, first.Deck[i], second.Deck[i])
		}
	}
}
