package research

import (
	"fmt"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// ConstraintReport contains hard-constraint checks for a deck.
type ConstraintReport struct {
	Violations []string
}

func (r ConstraintReport) IsValid() bool {
	return len(r.Violations) == 0
}

// ValidateConstraints enforces the phase-1 hard constraints.
//
//nolint:gocyclo // Explicit rule checks keep constraints auditable.
func ValidateConstraints(cards []deck.CardCandidate) ConstraintReport {
	violations := make([]string, 0)
	if len(cards) != 8 {
		violations = append(violations, fmt.Sprintf("deck size must be 8, got %d", len(cards)))
	}

	seen := make(map[string]bool)
	winCons := 0
	spells := 0
	airDefense := 0
	tankKillers := 0

	for _, c := range cards {
		if seen[c.Name] {
			violations = append(violations, fmt.Sprintf("duplicate card: %s", c.Name))
		}
		seen[c.Name] = true

		if c.Role != nil && *c.Role == deck.RoleWinCondition {
			winCons++
		}
		if c.Role != nil && (*c.Role == deck.RoleSpellBig || *c.Role == deck.RoleSpellSmall) {
			spells++
		}
		if canTargetAir(c) {
			airDefense++
		}
		if isTankKiller(c) {
			tankKillers++
		}
	}

	if winCons < 1 {
		violations = append(violations, "must include at least 1 win condition")
	}
	if spells < 1 {
		violations = append(violations, "must include at least 1 spell")
	}
	if airDefense < 2 {
		violations = append(violations, "must include at least 2 air-defense cards")
	}
	if tankKillers < 1 {
		violations = append(violations, "must include at least 1 tank-killer")
	}

	return ConstraintReport{Violations: violations}
}
