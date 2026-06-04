package main

import "testing"

func TestDeckSharedBuilderFlagsContainExpectedNames(t *testing.T) {
	command := addDeckBuildCommand()
	declared := commandFlagSet(command)

	for _, name := range []string{
		unlockedEvolutionsFlagName,
		"evolution-slots",
		"combat-stats-weight",
		"disable-combat-stats",
		"enable-synergy",
		"synergy-weight",
		"prefer-unique",
		"uniqueness-weight",
		"avoid-archetype",
		"fuzz-storage",
		"fuzz-weight",
		"fuzz-deck-limit",
		boostedCardLevelFlagName,
	} {
		if _, ok := declared[name]; !ok {
			t.Fatalf("addDeckBuildCommand() missing shared flag %q", name)
		}
	}
}

func TestDeckWarCommandContainsSharedEvolutionAndCombatFlags(t *testing.T) {
	command := addDeckWarCommand()
	declared := commandFlagSet(command)

	for _, name := range []string{
		unlockedEvolutionsFlagName,
		"evolution-slots",
		"combat-stats-weight",
		"disable-combat-stats",
	} {
		if _, ok := declared[name]; !ok {
			t.Fatalf("addDeckWarCommand() missing shared flag %q", name)
		}
	}
}
