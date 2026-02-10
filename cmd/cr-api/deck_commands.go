package main

import "github.com/urfave/cli/v3"

const (
	deckStrategyAll = "all"

	rarityCommon    = "Common"
	rarityRare      = "Rare"
	rarityEpic      = "Epic"
	rarityLegendary = "Legendary"
	rarityChampion  = "Champion"
)

// addDeckCommands adds deck-related subcommands to the CLI
func addDeckCommands() *cli.Command {
	return &cli.Command{
		Name:  "deck",
		Usage: "Deck building and analysis commands",
		Commands: []*cli.Command{
			addDeckEvaluateCommand(),
			addDeckBuildCommand(),
			addDeckBuildSuiteCommand(),
			addDeckEvaluateBatchCommand(),
			addDeckAnalyzeSuiteCommand(),
			addDeckWarCommand(),
			addDeckAnalyzeCommand(),
			addDeckOptimizeCommand(),
			addDeckRecommendCommand(),
			addDeckMulliganCommand(),
			addDeckBudgetCommand(),
			addDeckPossibleCountCommand(),
			addDeckFuzzCommand(),
			addDeckCompareAlgorithmsCommand(),
			addDiscoverCommands(),
			addLeaderboardCommands(),
			addStorageCommands(),
		},
	}
}
