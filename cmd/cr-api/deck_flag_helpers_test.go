package main

import (
	"testing"

	"github.com/urfave/cli/v3"
)

var sharedBuilderFlagNames = []string{
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
}

func TestDeckSharedBuilderFlagsContainExpectedNames(t *testing.T) {
	for _, tt := range []struct {
		name    string
		command *cli.Command
	}{
		{name: "addDeckBuildCommand", command: addDeckBuildCommand()},
		{name: "addDeckBuildSuiteCommand", command: addDeckBuildSuiteCommand()},
		{name: "addDeckAnalyzeSuiteCommand", command: addDeckAnalyzeSuiteCommand()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assertCommandHasFlags(t, tt.name, tt.command, sharedBuilderFlagNames)
		})
	}
}

func TestDeckWarCommandContainsSharedEvolutionAndCombatFlags(t *testing.T) {
	command := addDeckWarCommand()

	assertCommandHasFlags(t, "addDeckWarCommand", command, []string{
		unlockedEvolutionsFlagName,
		"evolution-slots",
		"combat-stats-weight",
		"disable-combat-stats",
	})
}

func assertCommandHasFlags(t *testing.T, commandName string, command *cli.Command, names []string) {
	t.Helper()

	declared := commandFlagSet(command)
	for _, name := range names {
		if _, ok := declared[name]; !ok {
			t.Fatalf("%s() missing shared flag %q", commandName, name)
		}
	}
}
