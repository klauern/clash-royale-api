package main

import (
	"github.com/urfave/cli/v3"
)

// addDeckWarCommand adds the deck war command
func addDeckWarCommand() *cli.Command {
	return &cli.Command{
		Name:  "war",
		Usage: "Build a 4-deck war set with no repeated cards",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "tag",
				Aliases:  []string{"p"},
				Usage:    "Player tag (without #)",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "deck-count",
				Value: 4,
				Usage: "Number of decks to build (default 4)",
			},
			unlockedEvolutionsFlag(),
			&cli.IntFlag{
				Name:  "evolution-slots",
				Value: 2,
				Usage: "Number of evolution slots available (default 2)",
			},
			&cli.Float64Flag{
				Name:  "combat-stats-weight",
				Value: 0.25,
				Usage: "Weight for combat stats in scoring (0.0-1.0, where 0=disabled, 0.25=default, 1.0=combat-only)",
			},
			&cli.BoolFlag{
				Name:  "disable-combat-stats",
				Usage: "Disable combat stats completely (use traditional scoring only)",
			},
		},
		Action: deckWarCommand,
	}
}

// addDeckMulliganCommand adds the deck mulligan command
func addDeckMulliganCommand() *cli.Command {
	return &cli.Command{
		Name:  "mulligan",
		Usage: "Generate mulligan guide (opening hand strategy) for a deck",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "cards",
				Aliases:  []string{"c"},
				Usage:    "8 card names for the deck to analyze",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "deck-name",
				Usage: "Custom name for the deck (optional)",
			},
			&cli.BoolFlag{
				Name:  "save",
				Usage: "Save mulligan guide to file",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output guide in JSON format",
			},
		},
		Action: deckMulliganCommand,
	}
}

// addDeckBudgetCommand adds the deck budget command
func addDeckBudgetCommand() *cli.Command {
	return &cli.Command{
		Name:  "budget",
		Usage: "Find budget-optimized decks with minimal upgrade investment",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "tag", Aliases: []string{"p"}, Usage: "Player tag (without #)", Required: true},
			&cli.IntFlag{Name: "max-cards", Value: 0, Usage: "Maximum cards needed for upgrades (0 = no limit)"},
			&cli.IntFlag{Name: "max-gold", Value: 0, Usage: "Maximum gold needed for upgrades (0 = no limit)"},
			&cli.Float64Flag{Name: "target-level", Value: 12.0, Usage: "Target average card level for viability"},
			&cli.StringFlag{Name: "sort-by", Value: "roi", Usage: "Sort results by: roi, cost_efficiency, total_cards, total_gold, current_score, projected_score"},
			&cli.IntFlag{Name: "top-n", Value: 10, Usage: "Number of top decks to display"},
			&cli.BoolFlag{Name: "include-variations", Usage: "Generate and analyze deck variations"},
			&cli.IntFlag{Name: "max-variations", Value: 5, Usage: "Maximum number of deck variations to generate"},
			&cli.BoolFlag{Name: "quick-wins", Usage: "Show only quick-win decks (1-2 upgrades away)"},
			&cli.BoolFlag{Name: "ready-only", Usage: "Show only decks that are already competitive"},
			unlockedEvolutionsFlag(),
			&cli.IntFlag{Name: "evolution-slots", Value: 2, Usage: "Number of evolution slots available"},
			&cli.BoolFlag{Name: "json", Usage: "Output results in JSON format"},
			&cli.BoolFlag{Name: "save", Usage: "Save results to file"},
		},
		Action: deckBudgetCommand,
	}
}

// addDeckPossibleCountCommand adds the deck possible-count command
func addDeckPossibleCountCommand() *cli.Command {
	return &cli.Command{
		Name:  "possible-count",
		Usage: "Calculate the number of possible deck combinations from available cards",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "tag",
				Aliases:  []string{"p"},
				Usage:    "Player tag (without #) - calculates decks possible with player's cards",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "format",
				Value: "human",
				Usage: "Output format: human, json, csv",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show detailed breakdown by role and elixir range",
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Save output to file (optional, prints to stdout if not specified)",
			},
		},
		Action: deckPossibleCountCommand,
	}
}

// addDeckCompareAlgorithmsCommand adds the deck compare-algorithms command
func addDeckCompareAlgorithmsCommand() *cli.Command {
	return &cli.Command{
		Name:  "compare-algorithms",
		Usage: "Compare V1 vs V2 deck building algorithms on quality metrics",
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
				Value:   "balanced,cycle,control,aggro,splash",
				Usage:   "Strategies to compare (comma-separated): balanced, cycle, control, aggro, splash, spell, synergy",
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Output file path for comparison report (optional, prints to stdout if not specified)",
			},
			&cli.StringFlag{
				Name:  "format",
				Value: "markdown",
				Usage: "Output format: markdown, json",
			},
			&cli.Float64Flag{
				Name:  "significance",
				Value: 0.05,
				Usage: "Significance threshold for determining winner (default: 0.05 = 5%)",
			},
			&cli.Float64Flag{
				Name:  "win-threshold",
				Value: 0.10,
				Usage: "Win threshold for significant wins/losses (default: 0.10 = 10%)",
			},
		},
		Action: deckCompareAlgorithmsCommand,
	}
}

// addDeckResearchEvalCommand adds the deck research-eval command.
//
//nolint:funlen // Flag list is intentionally explicit for discoverability.
func addDeckResearchEvalCommand() *cli.Command {
	return &cli.Command{
		Name:  "research-eval",
		Usage: "Benchmark archetype-free deck-building methods against baseline",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "tags",
				Usage: "Player tags to benchmark (without #). If omitted, uses phase-1 defaults",
			},
			&cli.StringFlag{
				Name:  "methods",
				Value: "baseline,genetic,constraint,role-first",
				Usage: "Comma-separated methods: baseline, genetic, constraint, role-first",
			},
			&cli.IntFlag{
				Name:  "seed",
				Value: 42,
				Usage: "Random seed for reproducible benchmarking",
			},
			&cli.IntFlag{
				Name:  "top",
				Value: 1,
				Usage: "Method-specific top-N setting (reserved for future multi-deck outputs)",
			},
			&cli.StringFlag{
				Name:  "output-dir",
				Value: "reports/research-eval",
				Usage: "Directory for benchmark.json and benchmark.md",
			},
			&cli.StringFlag{
				Name:  "data-dir",
				Value: "data",
				Usage: "Directory containing card metadata (cards_stats.json)",
			},
			&cli.IntFlag{
				Name:  "min-wincons",
				Value: 1,
				Usage: "Minimum win conditions required by hard constraints",
			},
			&cli.IntFlag{
				Name:  "min-spells",
				Value: 1,
				Usage: "Minimum spells required by hard constraints",
			},
			&cli.IntFlag{
				Name:  "min-air",
				Value: 2,
				Usage: "Minimum air-defense cards required by hard constraints",
			},
			&cli.IntFlag{
				Name:  "min-tank-killers",
				Value: 1,
				Usage: "Minimum tank-killers required by hard constraints",
			},
			&cli.Float64Flag{
				Name:  "weight-synergy",
				Value: 0.30,
				Usage: "Soft-objective weight for synergy score",
			},
			&cli.Float64Flag{
				Name:  "weight-coverage",
				Value: 0.25,
				Usage: "Soft-objective weight for coverage score",
			},
			&cli.Float64Flag{
				Name:  "weight-role-fit",
				Value: 0.20,
				Usage: "Soft-objective weight for role-fit score",
			},
			&cli.Float64Flag{
				Name:  "weight-elixir-fit",
				Value: 0.15,
				Usage: "Soft-objective weight for elixir-fit score",
			},
			&cli.Float64Flag{
				Name:  "weight-card-quality",
				Value: 0.10,
				Usage: "Soft-objective weight for card-quality score",
			},
			&cli.StringFlag{
				Name:    "api-token",
				Aliases: []string{"t"},
				Usage:   "Clash Royale API token",
			},
		},
		Action: deckResearchEvalCommand,
	}
}
