package evaluation

import (
	"strings"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestBuildDefenseAnalysis_WarnsWhenNoResetRetargetCoverage(t *testing.T) {
	deckCards := []deck.CardCandidate{
		makeCard("Royal Giant", deck.RoleWinCondition, 11, 11, "Common", 6),
		makeCard("Witch", deck.RoleSupport, 11, 11, "Epic", 5),
		makeCard("Mini P.E.K.K.A", deck.RoleSupport, 11, 11, "Rare", 4),
		makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
		makeCard("Dart Goblin", deck.RoleSupport, 11, 11, "Rare", 3),
		makeCard("Hunter", deck.RoleSupport, 11, 11, "Epic", 4),
		makeCard("The Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
		makeCard("Executioner", deck.RoleSupport, 11, 11, "Epic", 5),
	}

	result := BuildDefenseAnalysis(deckCards)

	joinedDetails := strings.Join(result.Details, " ")
	if !strings.Contains(joinedDetails, "No reset/retarget tools") {
		t.Fatalf("expected reset/retarget warning in defense details, got: %v", result.Details)
	}
}

func TestBuildDefenseAnalysis_ShowsResetRetargetTools(t *testing.T) {
	deckCards := []deck.CardCandidate{
		makeCard("Royal Giant", deck.RoleWinCondition, 11, 11, "Common", 6),
		makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
		makeCard("Zap", deck.RoleSpellSmall, 11, 11, "Common", 2),
		makeCard("Electro Wizard", deck.RoleSupport, 11, 11, "Legendary", 4),
		makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
		makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
		makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
		makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
	}

	result := BuildDefenseAnalysis(deckCards)

	joinedDetails := strings.Join(result.Details, " ")
	if !strings.Contains(joinedDetails, "Reset/retarget tools") {
		t.Fatalf("expected reset/retarget tools detail, got: %v", result.Details)
	}
	if strings.Contains(joinedDetails, "No reset/retarget tools") {
		t.Fatalf("did not expect reset/retarget warning when deck has reset tools: %v", result.Details)
	}
}
