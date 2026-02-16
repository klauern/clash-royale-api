package research

import (
	"testing"
)

func TestValidateConstraintsRejectsDuplicateAndMissingWincon(t *testing.T) {
	cards := testDeck()
	cards[0] = cards[1]
	report := ValidateConstraints(cards, DefaultConstraintConfig())
	if report.IsValid() {
		t.Fatalf("expected invalid deck")
	}
	if len(report.Violations) == 0 {
		t.Fatalf("expected at least one violation")
	}
}

func TestValidateConstraintsAcceptsValidDeck(t *testing.T) {
	report := ValidateConstraints(testDeck(), DefaultConstraintConfig())
	if !report.IsValid() {
		t.Fatalf("expected valid deck, got: %v", report.Violations)
	}
}

func TestValidateConstraintsUsesCustomHardMinimums(t *testing.T) {
	cfg := DefaultConstraintConfig()
	cfg.Hard.MinSpells = 3
	report := ValidateConstraints(testDeck(), cfg)
	if report.IsValid() {
		t.Fatalf("expected invalid deck with min_spells=3")
	}
}
