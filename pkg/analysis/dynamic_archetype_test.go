package analysis

import "testing"

func TestEstimateGoldCost_UsesConfiguredValues(t *testing.T) {
	card := CardLevelInfo{
		Level:  10,
		Rarity: rarityCommon,
	}

	got := estimateGoldCost(card)
	if got != 20000 {
		t.Fatalf("estimateGoldCost(common level 10) = %d, want 20000", got)
	}
}

func TestEstimateGoldCost_FallsBackToNearestConfiguredLevel(t *testing.T) {
	card := CardLevelInfo{
		Level:  15,
		Rarity: rarityCommon,
	}

	// No direct level-15 gold entry; fallback should use nearest known (level 13 -> 100000).
	got := estimateGoldCost(card)
	if got != 100000 {
		t.Fatalf("estimateGoldCost(common level 15) = %d, want 100000", got)
	}
}

func TestEstimateGoldCost_UnknownRarityFallback(t *testing.T) {
	card := CardLevelInfo{
		Level:  10,
		Rarity: "Mythic",
	}

	got := estimateGoldCost(card)
	if got != 5000 {
		t.Fatalf("estimateGoldCost(unknown rarity) = %d, want 5000", got)
	}
}
