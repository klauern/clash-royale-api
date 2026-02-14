package archetypes

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
)

func TestGetArchetypeConstraintsReturnsDefensiveCopies(t *testing.T) {
	first := GetArchetypeConstraints()
	beatdown := first[mulligan.ArchetypeBeatdown]
	beatdown.RequiredRoles[deck.RoleSupport] = 99
	beatdown.PreferredCards[0] = "Mutated Card"
	beatdown.ExcludedCards[0] = "Mutated Exclusion"
	first[mulligan.ArchetypeBeatdown] = beatdown

	second := GetArchetypeConstraints()
	got := second[mulligan.ArchetypeBeatdown]

	if got.RequiredRoles[deck.RoleSupport] == 99 {
		t.Fatalf("required role mutation leaked into shared constraints")
	}
	if got.PreferredCards[0] == "Mutated Card" {
		t.Fatalf("preferred card mutation leaked into shared constraints")
	}
	if got.ExcludedCards[0] == "Mutated Exclusion" {
		t.Fatalf("excluded card mutation leaked into shared constraints")
	}
}
