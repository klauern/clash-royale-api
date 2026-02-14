package research

import (
	"testing"
)

func TestValidateConstraintsRejectsDuplicateAndMissingWincon(t *testing.T) {
	cards := testDeck()
	cards[0] = cards[1]
	report := ValidateConstraints(cards)
	if report.IsValid() {
		t.Fatalf("expected invalid deck")
	}
	if len(report.Violations) == 0 {
		t.Fatalf("expected at least one violation")
	}
}

func TestValidateConstraintsAcceptsValidDeck(t *testing.T) {
	report := ValidateConstraints(testDeck())
	if !report.IsValid() {
		t.Fatalf("expected valid deck, got: %v", report.Violations)
	}
}
