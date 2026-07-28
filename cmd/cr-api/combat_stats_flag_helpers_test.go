package main

import (
	"testing"
)

func TestCombatStatsFlagsExposeExpectedDefaults(t *testing.T) {
	t.Parallel()

	if combatStatsWeightFlag.Name != combatStatsWeightFlagName {
		t.Fatalf("weight flag name mismatch: got %q, want %q", combatStatsWeightFlag.Name, combatStatsWeightFlagName)
	}
	if combatStatsWeightFlag.Value != defaultCombatStatsWeight {
		t.Fatalf("weight flag default mismatch: got %.2f, want %.2f", combatStatsWeightFlag.Value, defaultCombatStatsWeight)
	}

	if disableCombatStatsFlag.Name != disableCombatStatsFlagName {
		t.Fatalf("disable flag name mismatch: got %q, want %q", disableCombatStatsFlag.Name, disableCombatStatsFlagName)
	}
}
