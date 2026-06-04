package main

import (
	"github.com/klauer/clash-royale-api/go/pkg/deck/genetic"
	"github.com/urfave/cli/v3"
)

// addDeckBuildCommand adds the deck build command
func addDeckBuildCommand() *cli.Command {
	_ = genetic.DefaultGeneticConfig()
	flags := []cli.Flag{
		playerTagFlag(true),
		&cli.StringFlag{Name: strategyFlagName, Aliases: []string{"s"}, Value: optimizeFocusBalanced, Usage: "Deck building strategy: balanced, aggro, control, cycle, splash, spell, synergy-first, all"},
		&cli.Float64Flag{Name: minElixirFlagName, Value: 2.5, Usage: "Minimum average elixir for the deck"},
		&cli.Float64Flag{Name: maxElixirFlagName, Value: 4.5, Usage: "Maximum average elixir for the deck"},
		&cli.StringSliceFlag{Name: includeCardsFlagName, Usage: "Specific cards to include in the deck (by name)"},
		&cli.StringSliceFlag{Name: excludeCardsFlagName, Usage: "Cards to exclude from the deck (by name)"},
		&cli.IntFlag{Name: "min-level", Value: 1, Usage: "Minimum card level to consider"},
		&cli.BoolFlag{Name: "prioritize-upgrades", Usage: "Prioritize cards that can be upgraded soon"},
		&cli.BoolFlag{Name: exportCSVFlagName, Usage: "Export deck analysis to CSV"},
		&cli.BoolFlag{Name: saveFlagName, Usage: "Save deck to file"},
	}
	flags = append(flags, deckSharedBuilderFlags()...)
	flags = append(flags,
		&cli.BoolFlag{Name: fromAnalysisFlagName, Aliases: []string{"a"}, Usage: "Enable offline mode: load analysis from JSON file instead of fetching from API"},
		&cli.StringFlag{Name: analysisDirFlagName, Usage: "Directory containing analysis JSON files (default: data/analysis)"},
		&cli.StringFlag{Name: analysisFileFlagName, Usage: "Specific analysis file path (overrides --analysis-dir lookup)"},
		&cli.BoolFlag{Name: "no-suggest-upgrades", Usage: "Disable upgrade recommendations for the built deck (recommendations are shown by default)"},
		&cli.IntFlag{Name: "upgrade-count", Value: 5, Usage: "Number of upgrade recommendations to show (default 5)"},
		&cli.BoolFlag{Name: "ideal-deck", Usage: "Show ideal deck composition after applying recommended upgrades"},
	)
	return &cli.Command{
		Name:   "build",
		Usage:  "Build an optimized deck based on player's card collection",
		Flags:  flags,
		Action: deckBuildCommand,
	}
}

//nolint:funlen // Command flag matrix is intentionally explicit for CLI discoverability.
func addDeckBuildSuiteCommand() *cli.Command {
	flags := []cli.Flag{
		playerTagFlag(true),
		&cli.StringFlag{
			Name:    strategiesFlagName,
			Aliases: []string{"s"},
			Value:   optimizeFocusBalanced,
			Usage:   "Comma-separated list of strategies or 'all': balanced,aggro,control,cycle,splash,spell",
		},
		&cli.IntFlag{
			Name:  "variations",
			Value: 1,
			Usage: "Number of variations per strategy (default 1)",
		},
		&cli.StringFlag{
			Name:  outputDirFlagName,
			Usage: "Output directory for deck files (default: data/decks/)",
		},
		&cli.BoolFlag{
			Name:  fromAnalysisFlagName,
			Usage: "Use offline mode with pre-analyzed player data",
		},
		&cli.Float64Flag{
			Name:  minElixirFlagName,
			Value: 2.5,
			Usage: "Minimum average elixir for the deck",
		},
		&cli.Float64Flag{
			Name:  maxElixirFlagName,
			Value: 4.5,
			Usage: "Maximum average elixir for the deck",
		},
		&cli.StringSliceFlag{
			Name:  includeCardsFlagName,
			Usage: "Cards that must be included in all decks",
		},
		&cli.StringSliceFlag{
			Name:  excludeCardsFlagName,
			Usage: "Cards that must be excluded from all decks",
		},
	}
	flags = append(flags, deckSharedBuilderFlags()...)
	flags = append(flags, &cli.BoolFlag{
		Name:  saveFlagName,
		Value: true,
		Usage: "Save individual deck files and summary JSON (default: true)",
	})
	return &cli.Command{
		Name:   "build-suite",
		Usage:  "Build multiple deck variations in one invocation for systematic analysis",
		Flags:  flags,
		Action: deckBuildSuiteCommand,
	}
}

// addDeckAnalyzeSuiteCommand adds the deck analyze-suite command
func addDeckAnalyzeSuiteCommand() *cli.Command {
	flags := []cli.Flag{
		playerTagFlag(true),
		&cli.StringFlag{Name: strategiesFlagName, Aliases: []string{"s"}, Value: deckStrategyAll, Usage: "Deck building strategies (comma-separated or 'all'): balanced, aggro, control, cycle, splash, spell"},
		&cli.IntFlag{Name: "variations", Value: 1, Usage: "Number of variations per strategy"},
		&cli.StringFlag{Name: outputDirFlagName, Value: "data/analysis", Usage: "Base output directory for all analysis results"},
		&cli.IntFlag{Name: topNFlagName, Value: 5, Usage: "Number of top decks to compare in final report"},
		&cli.BoolFlag{Name: fromAnalysisFlagName, Usage: "Use offline mode (load from existing analysis files instead of API)"},
		&cli.Float64Flag{Name: minElixirFlagName, Value: 2.5, Usage: "Minimum average elixir for decks"},
		&cli.Float64Flag{Name: maxElixirFlagName, Value: 4.5, Usage: "Maximum average elixir for decks"},
		&cli.StringSliceFlag{Name: includeCardsFlagName, Usage: "Cards that must be included in all decks"},
		&cli.StringSliceFlag{Name: excludeCardsFlagName, Usage: "Cards that must be excluded from all decks"},
	}
	flags = append(flags, deckSharedBuilderFlags()...)
	flags = append(flags,
		&cli.BoolFlag{Name: "suggest-constraints", Usage: "analyze top N decks and suggest card constraints based on frequency", Value: false},
		&cli.Float64Flag{Name: "constraint-threshold", Usage: "minimum percentage threshold for card suggestions (0-100)", Value: 50.0},
	)
	return &cli.Command{
		Name:   "analyze-suite",
		Usage:  "Build deck variations, evaluate all decks, compare top performers, and generate comprehensive report",
		Flags:  flags,
		Action: deckAnalyzeSuiteCommand,
	}
}
