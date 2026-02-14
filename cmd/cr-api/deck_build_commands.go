package main

import (
	"github.com/klauer/clash-royale-api/go/pkg/deck/genetic"
	"github.com/urfave/cli/v3"
)

// addDeckBuildCommand adds the deck build command
func addDeckBuildCommand() *cli.Command {
	_ = genetic.DefaultGeneticConfig()
	return &cli.Command{
		Name:  "build",
		Usage: "Build an optimized deck based on player's card collection",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "tag", Aliases: []string{"p"}, Usage: "Player tag (without #)", Required: true},
			&cli.StringFlag{Name: "strategy", Aliases: []string{"s"}, Value: "balanced", Usage: "Deck building strategy: balanced, aggro, control, cycle, splash, spell, synergy-first, all"},
			&cli.Float64Flag{Name: "min-elixir", Value: 2.5, Usage: "Minimum average elixir for the deck"},
			&cli.Float64Flag{Name: "max-elixir", Value: 4.5, Usage: "Maximum average elixir for the deck"},
			&cli.StringSliceFlag{Name: "include-cards", Usage: "Specific cards to include in the deck (by name)"},
			&cli.StringSliceFlag{Name: "exclude-cards", Usage: "Cards to exclude from the deck (by name)"},
			&cli.IntFlag{Name: "min-level", Value: 1, Usage: "Minimum card level to consider"},
			&cli.BoolFlag{Name: "prioritize-upgrades", Usage: "Prioritize cards that can be upgraded soon"},
			&cli.BoolFlag{Name: "export-csv", Usage: "Export deck analysis to CSV"},
			&cli.BoolFlag{Name: "save", Usage: "Save deck to file"},
			&cli.StringFlag{Name: "unlocked-evolutions", Usage: "Comma-separated list of cards with unlocked evolutions (overrides UNLOCKED_EVOLUTIONS env var)"},
			&cli.IntFlag{Name: "evolution-slots", Value: 2, Usage: "Number of evolution slots available (default 2)"},
			&cli.Float64Flag{Name: "combat-stats-weight", Value: 0.25, Usage: "Weight for combat stats in scoring (0.0-1.0, where 0=disabled, 0.25=default, 1.0=combat-only)"},
			&cli.BoolFlag{Name: "disable-combat-stats", Usage: "Disable combat stats completely (use traditional scoring only)"},
			&cli.BoolFlag{Name: "from-analysis", Aliases: []string{"a"}, Usage: "Enable offline mode: load analysis from JSON file instead of fetching from API"},
			&cli.StringFlag{Name: "analysis-dir", Usage: "Directory containing analysis JSON files (default: data/analysis)"},
			&cli.StringFlag{Name: "analysis-file", Usage: "Specific analysis file path (overrides --analysis-dir lookup)"},
			&cli.BoolFlag{Name: "enable-synergy", Usage: "Enable synergy-based card selection (considers card interactions and combos)"},
			&cli.Float64Flag{Name: "synergy-weight", Value: 0.15, Usage: "Weight for synergy scoring (0.0-1.0, default 0.15 = 15%)"},
			&cli.BoolFlag{Name: "prefer-unique", Usage: "Enable uniqueness/anti-meta scoring (prefers less common cards)"},
			&cli.Float64Flag{Name: "uniqueness-weight", Value: 0.2, Usage: "Weight for uniqueness scoring (0.0-0.3, default 0.2 = 20%)"},
			&cli.StringSliceFlag{Name: "avoid-archetype", Usage: "Archetypes to avoid when building decks (e.g., beatdown, cycle, control, siege, bridge_spam, bait, spawndeck, midrange)"},
			&cli.BoolFlag{Name: "no-suggest-upgrades", Usage: "Disable upgrade recommendations for the built deck (recommendations are shown by default)"},
			&cli.IntFlag{Name: "upgrade-count", Value: 5, Usage: "Number of upgrade recommendations to show (default 5)"},
			&cli.BoolFlag{Name: "ideal-deck", Usage: "Show ideal deck composition after applying recommended upgrades"},
			&cli.StringFlag{Name: "fuzz-storage", Usage: "Path to fuzz storage database for data-driven card scoring (default: ~/.cr-api/fuzz_top_decks.db)"},
			&cli.Float64Flag{Name: "fuzz-weight", Value: 0.10, Usage: "Weight for fuzz-based card scoring (0.0-1.0, default 0.10 = 10%)"},
			&cli.IntFlag{Name: "fuzz-deck-limit", Value: 100, Usage: "Number of top fuzz decks to analyze for card stats (default 100)"},
		},
		Action: deckBuildCommand,
	}
}

func addDeckBuildSuiteCommand() *cli.Command {
	return &cli.Command{
		Name:  "build-suite",
		Usage: "Build multiple deck variations in one invocation for systematic analysis",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "tag",
				Aliases:  []string{"p"},
				Usage:    "Player tag (without #)",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "strategies",
				Aliases: []string{"s"},
				Value:   "balanced",
				Usage:   "Comma-separated list of strategies or 'all': balanced,aggro,control,cycle,splash,spell",
			},
			&cli.IntFlag{
				Name:  "variations",
				Value: 1,
				Usage: "Number of variations per strategy (default 1)",
			},
			&cli.StringFlag{
				Name:  "output-dir",
				Usage: "Output directory for deck files (default: data/decks/)",
			},
			&cli.BoolFlag{
				Name:  "from-analysis",
				Usage: "Use offline mode with pre-analyzed player data",
			},
			&cli.Float64Flag{
				Name:  "min-elixir",
				Value: 2.5,
				Usage: "Minimum average elixir for the deck",
			},
			&cli.Float64Flag{
				Name:  "max-elixir",
				Value: 4.5,
				Usage: "Maximum average elixir for the deck",
			},
			&cli.StringSliceFlag{
				Name:  "include-cards",
				Usage: "Cards that must be included in all decks",
			},
				&cli.StringSliceFlag{
					Name:  "exclude-cards",
					Usage: "Cards that must be excluded from all decks",
				},
				&cli.StringFlag{Name: "unlocked-evolutions", Usage: "Comma-separated list of cards with unlocked evolutions (overrides UNLOCKED_EVOLUTIONS env var)"},
				&cli.IntFlag{Name: "evolution-slots", Value: 2, Usage: "Number of evolution slots available (default 2)"},
				&cli.Float64Flag{Name: "combat-stats-weight", Value: 0.25, Usage: "Weight for combat stats in scoring (0.0-1.0, where 0=disabled, 0.25=default, 1.0=combat-only)"},
				&cli.BoolFlag{Name: "disable-combat-stats", Usage: "Disable combat stats completely (use traditional scoring only)"},
				&cli.BoolFlag{Name: "enable-synergy", Usage: "Enable synergy-based card selection (considers card interactions and combos)"},
				&cli.Float64Flag{Name: "synergy-weight", Value: 0.15, Usage: "Weight for synergy scoring (0.0-1.0, default 0.15 = 15%)"},
				&cli.BoolFlag{Name: "prefer-unique", Usage: "Enable uniqueness/anti-meta scoring (prefers less common cards)"},
				&cli.Float64Flag{Name: "uniqueness-weight", Value: 0.2, Usage: "Weight for uniqueness scoring (0.0-0.3, default 0.2 = 20%)"},
				&cli.StringSliceFlag{Name: "avoid-archetype", Usage: "Archetypes to avoid when building decks (e.g., beatdown, cycle, control, siege, bridge_spam, bait, spawndeck, midrange)"},
				&cli.StringFlag{Name: "fuzz-storage", Usage: "Path to fuzz storage database for data-driven card scoring (default: ~/.cr-api/fuzz_top_decks.db)"},
				&cli.Float64Flag{Name: "fuzz-weight", Value: 0.10, Usage: "Weight for fuzz-based card scoring (0.0-1.0, default 0.10 = 10%)"},
				&cli.IntFlag{Name: "fuzz-deck-limit", Value: 100, Usage: "Number of top fuzz decks to analyze for card stats (default 100)"},
				&cli.BoolFlag{
					Name:  "save",
					Value: true,
				Usage: "Save individual deck files and summary JSON (default: true)",
			},
		},
		Action: deckBuildSuiteCommand,
	}
}

// addDeckAnalyzeSuiteCommand adds the deck analyze-suite command
func addDeckAnalyzeSuiteCommand() *cli.Command {
	return &cli.Command{
		Name:  "analyze-suite",
		Usage: "Build deck variations, evaluate all decks, compare top performers, and generate comprehensive report",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "tag", Aliases: []string{"p"}, Usage: "Player tag (without #)", Required: true},
			&cli.StringFlag{Name: "strategies", Aliases: []string{"s"}, Value: deckStrategyAll, Usage: "Deck building strategies (comma-separated or 'all'): balanced, aggro, control, cycle, splash, spell"},
			&cli.IntFlag{Name: "variations", Value: 1, Usage: "Number of variations per strategy"},
			&cli.StringFlag{Name: "output-dir", Value: "data/analysis", Usage: "Base output directory for all analysis results"},
			&cli.IntFlag{Name: "top-n", Value: 5, Usage: "Number of top decks to compare in final report"},
			&cli.BoolFlag{Name: "from-analysis", Usage: "Use offline mode (load from existing analysis files instead of API)"},
				&cli.Float64Flag{Name: "min-elixir", Value: 2.5, Usage: "Minimum average elixir for decks"},
				&cli.Float64Flag{Name: "max-elixir", Value: 4.5, Usage: "Maximum average elixir for decks"},
				&cli.StringSliceFlag{Name: "include-cards", Usage: "Cards that must be included in all decks"},
				&cli.StringSliceFlag{Name: "exclude-cards", Usage: "Cards that must be excluded from all decks"},
				&cli.StringFlag{Name: "unlocked-evolutions", Usage: "Comma-separated list of cards with unlocked evolutions (overrides UNLOCKED_EVOLUTIONS env var)"},
				&cli.IntFlag{Name: "evolution-slots", Value: 2, Usage: "Number of evolution slots available (default 2)"},
				&cli.Float64Flag{Name: "combat-stats-weight", Value: 0.25, Usage: "Weight for combat stats in scoring (0.0-1.0, where 0=disabled, 0.25=default, 1.0=combat-only)"},
				&cli.BoolFlag{Name: "disable-combat-stats", Usage: "Disable combat stats completely (use traditional scoring only)"},
				&cli.BoolFlag{Name: "enable-synergy", Usage: "Enable synergy-based card selection (considers card interactions and combos)"},
				&cli.Float64Flag{Name: "synergy-weight", Value: 0.15, Usage: "Weight for synergy scoring (0.0-1.0, default 0.15 = 15%)"},
				&cli.BoolFlag{Name: "prefer-unique", Usage: "Enable uniqueness/anti-meta scoring (prefers less common cards)"},
				&cli.Float64Flag{Name: "uniqueness-weight", Value: 0.2, Usage: "Weight for uniqueness scoring (0.0-0.3, default 0.2 = 20%)"},
				&cli.StringSliceFlag{Name: "avoid-archetype", Usage: "Archetypes to avoid when building decks (e.g., beatdown, cycle, control, siege, bridge_spam, bait, spawndeck, midrange)"},
				&cli.StringFlag{Name: "fuzz-storage", Usage: "Path to fuzz storage database for data-driven card scoring (default: ~/.cr-api/fuzz_top_decks.db)"},
				&cli.Float64Flag{Name: "fuzz-weight", Value: 0.10, Usage: "Weight for fuzz-based card scoring (0.0-1.0, default 0.10 = 10%)"},
				&cli.IntFlag{Name: "fuzz-deck-limit", Value: 100, Usage: "Number of top fuzz decks to analyze for card stats (default 100)"},
				&cli.BoolFlag{Name: "suggest-constraints", Usage: "analyze top N decks and suggest card constraints based on frequency", Value: false},
				&cli.Float64Flag{Name: "constraint-threshold", Usage: "minimum percentage threshold for card suggestions (0-100)", Value: 50.0},
			},
		Action: deckAnalyzeSuiteCommand,
	}
}
