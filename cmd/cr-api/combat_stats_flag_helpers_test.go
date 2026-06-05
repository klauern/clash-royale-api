package main

import (
	"testing"

	"github.com/urfave/cli/v3"
)

func TestCombatStatsFlagsExposeExpectedDefaults(t *testing.T) {
	t.Parallel()

	flags := combatStatsFlags()
	if len(flags) != 2 {
		t.Fatalf("combatStatsFlags length mismatch: got %d, want 2", len(flags))
	}

	weightFlag, ok := flags[0].(*cli.Float64Flag)
	if !ok {
		t.Fatalf("combatStatsFlags[0] type mismatch: got %T", flags[0])
	}
	if weightFlag.Name != combatStatsWeightFlagName {
		t.Fatalf("weight flag name mismatch: got %q, want %q", weightFlag.Name, combatStatsWeightFlagName)
	}
	if weightFlag.Value != defaultCombatStatsWeight {
		t.Fatalf("weight flag default mismatch: got %.2f, want %.2f", weightFlag.Value, defaultCombatStatsWeight)
	}

	disableFlag, ok := flags[1].(*cli.BoolFlag)
	if !ok {
		t.Fatalf("combatStatsFlags[1] type mismatch: got %T", flags[1])
	}
	if disableFlag.Name != disableCombatStatsFlagName {
		t.Fatalf("disable flag name mismatch: got %q, want %q", disableFlag.Name, disableCombatStatsFlagName)
	}
}
