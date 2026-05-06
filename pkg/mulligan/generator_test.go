package mulligan

import (
	"slices"
	"testing"
)

func TestSharedArchetypeSignals(t *testing.T) {
	t.Run("beatdown heavy win conditions", func(t *testing.T) {
		if !hasHeavyWinCondition([]string{"Golem"}) {
			t.Fatal("expected Golem to match beatdown heavy signal")
		}
		if hasHeavyWinCondition([]string{"Hog Rider"}) {
			t.Fatal("did not expect Hog Rider to match beatdown heavy signal")
		}
	})

	t.Run("siege buildings", func(t *testing.T) {
		if !hasSiegeBuilding([]string{"Xbow"}) {
			t.Fatal("expected Xbow to match siege signal")
		}
		if hasSiegeBuilding([]string{"Inferno Tower"}) {
			t.Fatal("did not expect Inferno Tower to match siege signal")
		}
	})

	t.Run("bridge spam win conditions", func(t *testing.T) {
		if !hasBridgeSpamWinCondition([]string{"Battle Ram"}) {
			t.Fatal("expected Battle Ram to match bridge-spam signal")
		}
		if !hasBridgeSpamWinCondition([]string{"Hog Rider"}) {
			t.Fatal("expected Hog Rider to match bridge-spam signal")
		}
	})
}

// TestGenerateGuide_KeyCardsPopulated reproduces clash-royale-api-gmk where
// every matchup block printed empty Key Cards because unknown cards
// defaulted to RoleSupport and RoleSupport was dropped from analyzeDeck's
// categorization switch.
func TestGenerateGuide_KeyCardsPopulated(t *testing.T) {
	gen := NewGenerator()
	cards := []string{
		"Witch", "Golden Knight", "Balloon", "Dark Prince",
		"Skeleton Dragons", "Minion Horde", "Bowler", "Tornado",
	}
	guide, err := gen.GenerateGuide(cards, "test")
	if err != nil {
		t.Fatalf("GenerateGuide: %v", err)
	}
	if len(guide.Matchups) != 6 {
		t.Fatalf("Matchups: want 6, got %d", len(guide.Matchups))
	}
	for i, m := range guide.Matchups {
		if len(m.KeyCards) == 0 {
			t.Errorf("matchup %d (%q): KeyCards must not be empty", i, m.OpponentType)
		}
	}
}

// TestMapConfigRoleToMulligan_FallbackClassifier verifies the
// internal/config role bridge places cards into the right buckets when the
// mulligan-local cards.json has no entry.
func TestMapConfigRoleToMulligan_FallbackClassifier(t *testing.T) {
	cases := []struct {
		card     string
		expected CardRole
	}{
		{"Hog Rider", RoleWinCondition},
		{"Balloon", RoleWinCondition},
		{"Inferno Tower", RoleBuilding},
		{"Fireball", RoleSpell},
		{"Tornado", RoleSpell},
		{"Skeletons", RoleCycle},
		{"Witch", RoleSupport},
		{"Bowler", RoleSupport},
	}
	for _, tc := range cases {
		gen := NewGenerator()
		delete(gen.cardDatabase, tc.card)
		analysis := gen.analyzeDeck([]string{tc.card})
		if got := classifyForTest(analysis, tc.card); got != tc.expected {
			t.Errorf("%s: want bucket %q, got %q", tc.card, tc.expected, got)
		}
	}
}

func classifyForTest(a DeckAnalysis, card string) CardRole {
	switch {
	case slices.Contains(a.WinConditions, card):
		return RoleWinCondition
	case slices.Contains(a.Buildings, card):
		return RoleBuilding
	case slices.Contains(a.DefensiveCards, card):
		return RoleSupport
	case slices.Contains(a.CycleCards, card):
		return RoleCycle
	case slices.Contains(a.Spells, card):
		return RoleSpell
	}
	return ""
}
