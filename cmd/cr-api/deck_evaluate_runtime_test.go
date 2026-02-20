package main

import (
	"strings"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestCalculateUpgradeGoldCostHandlesNonPositiveFromLevel(t *testing.T) {
	got := calculateUpgradeGoldCost(rarityCommon, 0, 1)
	if got != 100 {
		t.Fatalf("calculateUpgradeGoldCost(common, 0, 1)=%d, want 100", got)
	}
}

func TestCalculateUpgradeCardsNeededHandlesNonPositiveFromLevel(t *testing.T) {
	got := calculateUpgradeCardsNeeded(rarityEpic, -3, 1)
	if got != 8 {
		t.Fatalf("calculateUpgradeCardsNeeded(epic, -3, 1)=%d, want 8", got)
	}
}

func TestCalculateUpgradeCostsReturnZeroWhenNoUpgrade(t *testing.T) {
	if got := calculateUpgradeGoldCost(rarityRare, 10, 10); got != 0 {
		t.Fatalf("calculateUpgradeGoldCost(rare, 10, 10)=%d, want 0", got)
	}
	if got := calculateUpgradeCardsNeeded(rarityRare, 10, 9); got != 0 {
		t.Fatalf("calculateUpgradeCardsNeeded(rare, 10, 9)=%d, want 0", got)
	}
}

func TestUpgradeUnlocksNewArchetype(t *testing.T) {
	if !upgradeUnlocksNewArchetype("Hog Rider", deck.CardLevelData{Level: 10, Elixir: 4, Rarity: rarityRare}, 11) {
		t.Fatal("expected win condition crossing level 11 to unlock new deck options")
	}

	if !upgradeUnlocksNewArchetype("Princess", deck.CardLevelData{Level: 11, Elixir: 3, Rarity: rarityLegendary}, 12) {
		t.Fatal("expected legendary crossing level 12 to unlock new deck options")
	}

	if upgradeUnlocksNewArchetype("Knight", deck.CardLevelData{Level: 10, Elixir: 3, Rarity: rarityCommon}, 11) {
		t.Fatal("did not expect non-win-condition common upgrade to unlock new deck options")
	}
}

func TestCalculateDeckCardUpgrades_SetsUnlocksNewDeck(t *testing.T) {
	analysis := deck.CardAnalysis{
		CardLevels: map[string]deck.CardLevelData{
			"Hog Rider": {Level: 10, MaxLevel: 15, Rarity: rarityRare, Elixir: 4},
		},
	}

	impacts := calculateDeckCardUpgrades([]string{"Hog Rider"}, analysis)
	if len(impacts) != 1 {
		t.Fatalf("expected one upgrade impact, got %d", len(impacts))
	}
	if !impacts[0].UnlocksNewDeck {
		t.Fatal("expected UnlocksNewDeck to be true for Hog Rider level 10->11")
	}
	if !strings.Contains(impacts[0].Reason, "unlock additional archetype options") {
		t.Fatalf("expected reason to mention unlock impact, got %q", impacts[0].Reason)
	}
}
