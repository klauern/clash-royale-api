package main

import (
	"testing"
)

func TestCombatStatsFlagsExposeExpectedDefaults(t *testing.T) {
	t.Parallel()

	if combatStatsWeightFlag.Name != "combat-stats-weight" {
		t.Fatalf("weight flag name mismatch: got %q, want %q", combatStatsWeightFlag.Name, combatStatsWeightFlagName)
	}
	if combatStatsWeightFlag.Value != 0.25 {
		t.Fatalf("weight flag default mismatch: got %.2f, want %.2f", combatStatsWeightFlag.Value, defaultCombatStatsWeight)
	}

	if disableCombatStatsFlag.Name != "disable-combat-stats" {
		t.Fatalf("disable flag name mismatch: got %q, want %q", disableCombatStatsFlag.Name, disableCombatStatsFlagName)
	}
}
