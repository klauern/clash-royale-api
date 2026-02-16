package main

import "testing"

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
