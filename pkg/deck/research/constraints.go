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
func ValidateConstraints(cards []deck.CardCandidate, cfg ConstraintConfig) ConstraintReport {
	violations := make([]string, 0)
	hard := cfg.Hard

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

	if winCons < hard.MinWinConditions {
		violations = append(violations, fmt.Sprintf("must include at least %d win condition(s)", hard.MinWinConditions))
	}
	if spells < hard.MinSpells {
		violations = append(violations, fmt.Sprintf("must include at least %d spell(s)", hard.MinSpells))
	}
	if airDefense < hard.MinAirDefense {
		violations = append(violations, fmt.Sprintf("must include at least %d air-defense card(s)", hard.MinAirDefense))
	}
	if tankKillers < hard.MinTankKillers {
		violations = append(violations, fmt.Sprintf("must include at least %d tank-killer(s)", hard.MinTankKillers))
	}

	return ConstraintReport{Violations: violations}
}
