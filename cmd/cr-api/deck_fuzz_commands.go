package main

import (
	"github.com/klauer/clash-royale-api/go/pkg/deck/genetic"
	"github.com/urfave/cli/v3"
)

// basicFlags returns the basic fuzzing parameters
func basicFlags() []cli.Flag {
	return []cli.Flag{
		playerTagFlagWithUsage(true, "Player tag (without #) for card collection context"),
		&cli.StringFlag{
			Name:  "mode",
			Value: "random",
			Usage: "Fuzzing mode: random or genetic",
		},
		&cli.IntFlag{
			Name:  "count",
			Value: 1000,
			Usage: "Number of random decks to generate and evaluate",
		},
		&cli.IntFlag{
			Name:  "workers",
			Value: 1,
			Usage: "Number of parallel workers for deck generation",
		},
	}
}

// cardConstraintFlags returns flags for card inclusion/exclusion constraints
func cardConstraintFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "include-cards",
			Usage: "Cards that must be included in every generated deck",
		},
		&cli.StringSliceFlag{
			Name:  "exclude-cards",
			Usage: "Cards that must be excluded from all generated decks",
		},
	}
}

// savedDeckFlags returns flags for saved deck integration
func savedDeckFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:  "include-from-saved",
			Usage: "Include top N decks from saved storage as starting points",
		},
		&cli.IntFlag{
			Name:  "from-saved",
			Usage: "Use saved top decks as seeds (generates mutations of saved decks)",
		},
		&cli.IntFlag{
			Name:  "resume-from",
			Usage: "Load top N saved decks as initial seed population (before random generation)",
		},
		&cli.StringFlag{
			Name:  "based-on",
			Usage: "Deck name or ID from saved storage to use as template for variations",
		},
	}
}

// scoreFilterFlags returns flags for score filtering and sorting
func scoreFilterFlags() []cli.Flag {
	return []cli.Flag{
		&cli.Float64Flag{
			Name:  "min-elixir",
			Value: 0.0,
			Usage: "Minimum average elixir for generated decks",
		},
		&cli.Float64Flag{
			Name:  "max-elixir",
			Value: 10.0,
			Usage: "Maximum average elixir for generated decks",
		},
		&cli.Float64Flag{
			Name:  "min-overall",
			Value: 0.0,
			Usage: "Minimum overall score to include in results (0.0-10.0)",
		},
		&cli.Float64Flag{
			Name:  "min-synergy",
			Value: 0.0,
			Usage: "Minimum synergy score to include in results (0.0-10.0)",
		},
		&cli.IntFlag{
			Name:  "top",
			Value: 10,
			Usage: "Number of top decks to display in results",
		},
		&cli.StringFlag{
			Name:  "sort-by",
			Value: "overall",
			Usage: "Sort results by: overall, attack, defense, synergy, versatility, elixir",
		},
		&cli.StringFlag{
			Name:  "format",
			Value: "summary",
			Usage: "Output format: summary, json, csv, detailed",
		},
	}
}

// outputFlags returns flags for output and storage configuration
func outputFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "output-dir",
			Usage: "Directory to save results (default: stdout only)",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Show detailed progress information",
		},
		&cli.BoolFlag{
			Name:  "from-analysis",
			Usage: "Load player data from existing analysis file (offline mode)",
		},
		&cli.StringFlag{
			Name:  "analysis-file",
			Usage: "Path to specific analysis file (for --from-analysis)",
		},
		&cli.StringFlag{
			Name:  "analysis-dir",
			Usage: "Directory containing analysis files (for --from-analysis)",
		},
		&cli.IntFlag{
			Name:  "seed",
			Value: 0,
			Usage: "Random seed for reproducibility (0 = random)",
		},
	}
}

// geneticAlgorithmFlags returns flags for genetic algorithm configuration
func geneticAlgorithmFlags(gaDefaults genetic.GeneticConfig) []cli.Flag {
	flags := []cli.Flag{}
	flags = append(flags, geneticAlgorithmBasicFlags(gaDefaults)...)
	flags = append(flags, geneticAlgorithmConvergenceFlags(gaDefaults)...)
	flags = append(flags, geneticAlgorithmIslandFlags(gaDefaults)...)
	return flags
}

// geneticAlgorithmBasicFlags returns basic genetic algorithm configuration flags
func geneticAlgorithmBasicFlags(gaDefaults genetic.GeneticConfig) []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:  "ga-population",
			Value: gaDefaults.PopulationSize,
			Usage: "Genetic algorithm population size",
		},
		&cli.IntFlag{
			Name:  "ga-generations",
			Value: gaDefaults.Generations,
			Usage: "Genetic algorithm generation count",
		},
		&cli.Float64Flag{
			Name:  "ga-mutation-rate",
			Value: gaDefaults.MutationRate,
			Usage: "Genetic algorithm mutation rate (0.0-1.0)",
		},
		&cli.Float64Flag{
			Name:  "ga-crossover-rate",
			Value: gaDefaults.CrossoverRate,
			Usage: "Genetic algorithm crossover rate (0.0-1.0)",
		},
		&cli.Float64Flag{
			Name:  "ga-mutation-intensity",
			Value: gaDefaults.MutationIntensity,
			Usage: "Genetic algorithm mutation intensity (0.0-1.0)",
		},
		&cli.IntFlag{
			Name:  "ga-elite-count",
			Value: gaDefaults.EliteCount,
			Usage: "Genetic algorithm elite count per generation",
		},
		&cli.IntFlag{
			Name:  "ga-tournament-size",
			Value: gaDefaults.TournamentSize,
			Usage: "Genetic algorithm tournament size",
		},
		&cli.BoolFlag{
			Name:  "ga-parallel-eval",
			Value: gaDefaults.ParallelEvaluations,
			Usage: "Enable parallel evaluation for genetic algorithm",
		},
	}
}

// geneticAlgorithmConvergenceFlags returns genetic algorithm convergence and early stopping flags
func geneticAlgorithmConvergenceFlags(gaDefaults genetic.GeneticConfig) []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:  "ga-convergence-generations",
			Value: gaDefaults.ConvergenceGenerations,
			Usage: "Genetic algorithm early stop after N generations without improvement (0 = disabled)",
		},
		&cli.Float64Flag{
			Name:  "ga-target-fitness",
			Value: gaDefaults.TargetFitness,
			Usage: "Genetic algorithm fitness target for early stop (0 = disabled)",
		},
	}
}

// geneticAlgorithmIslandFlags returns genetic algorithm island model flags
func geneticAlgorithmIslandFlags(gaDefaults genetic.GeneticConfig) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  "ga-island-model",
			Value: gaDefaults.IslandModel,
			Usage: "Enable island model for genetic algorithm",
		},
		&cli.IntFlag{
			Name:  "ga-island-count",
			Value: gaDefaults.IslandCount,
			Usage: "Number of islands when island model is enabled",
		},
		&cli.IntFlag{
			Name:  "ga-migration-interval",
			Value: gaDefaults.MigrationInterval,
			Usage: "Generations between island migrations",
		},
		&cli.IntFlag{
			Name:  "ga-migration-size",
			Value: gaDefaults.MigrationSize,
			Usage: "Number of migrants per island migration",
		},
		&cli.BoolFlag{
			Name:  "ga-use-archetypes",
			Value: gaDefaults.UseArchetypes,
			Usage: "Use legacy archetype-aware GA fitness objective (default uses archetype-free composite objective)",
		},
	}
}

// storageFlags returns flags for persistent storage configuration
func storageFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "storage",
			Usage: "Path to persistent storage database for saving evaluated decks",
		},
		&cli.BoolFlag{
			Name:  "save-top",
			Usage: "Save top decks to persistent storage for reuse in subsequent fuzz runs",
		},
		&cli.IntFlag{
			Name:  "analyze-top",
			Usage: "analyze top N saved decks and suggest card constraints based on frequency",
			Value: 0,
		},
		&cli.Float64Flag{
			Name:  "analyze-threshold",
			Usage: "minimum percentage threshold for card suggestions (0-100)",
			Value: 30.0,
		},
	}
}

// evolutionFlags returns flags for evolution-centric deck generation
func evolutionFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  "synergy-pairs",
			Usage: "Generate decks from 4 synergy pairs instead of role-based composition",
		},
		&cli.BoolFlag{
			Name:  "evolution-centric",
			Usage: "Generate decks focused on evolution-eligible cards (default: 3+ evo cards)",
		},
		&cli.IntFlag{
			Name:  "min-evo-cards",
			Usage: "Minimum number of evolution-eligible cards in deck (default: 3)",
			Value: 3,
		},
		&cli.IntFlag{
			Name:  "min-evo-level",
			Usage: "Minimum evolution level for cards to prioritize (default: 1)",
			Value: 1,
		},
		&cli.Float64Flag{
			Name:  "evo-weight",
			Usage: "Weight for evolution scoring in card selection (default: 0.3)",
			Value: 0.3,
		},
		&cli.IntFlag{
			Name:  "mutation-intensity",
			Usage: "Number of cards to swap during deck mutations (1-5, default: 2). Higher values create more diverse decks.",
			Value: 2,
		},
	}
}

// archetypeFlags returns flags for archetype-based deck generation
func archetypeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "archetypes",
			Usage: "Force generation from specific archetypes (comma-separated: beatdown,control,cycle,bridge,siege,bait,graveyard,miner,hybrid)",
		},
		&cli.IntFlag{
			Name:  "refine",
			Value: 1,
			Usage: "Number of refinement rounds. Each round uses top decks from previous round as seeds, with progressively tighter constraints",
		},
		&cli.Float64Flag{
			Name:  "uniqueness-weight",
			Value: 0.0,
			Usage: "Weight for card uniqueness scoring (0.0-0.3, default: 0.0 = disabled). Higher values prefer less common/anti-meta cards",
		},
		&cli.BoolFlag{
			Name:  "ensure-archetypes",
			Usage: "Ensure generated decks cover all archetypes (beatdown, control, cycle, bridge, siege, bait, graveyard, miner)",
		},
		&cli.BoolFlag{
			Name:  "ensure-elixir-buckets",
			Usage: "Ensure top decks are spread across low/medium/high average elixir buckets",
		},
	}
}

// advancedFlags returns flags for advanced deck generation and storage options
func advancedFlags() []cli.Flag {
	flags := []cli.Flag{}
	flags = append(flags, storageFlags()...)
	flags = append(flags, evolutionFlags()...)
	flags = append(flags, archetypeFlags()...)
	return flags
}

// addDeckFuzzCommand adds the fuzz command with subcommands
func addDeckFuzzCommand() *cli.Command {
	gaDefaults := genetic.DefaultGeneticConfig()
	flags := []cli.Flag{}
	flags = append(flags, basicFlags()...)
	flags = append(flags, cardConstraintFlags()...)
	flags = append(flags, savedDeckFlags()...)
	flags = append(flags, scoreFilterFlags()...)
	flags = append(flags, outputFlags()...)
	flags = append(flags, geneticAlgorithmFlags(gaDefaults)...)
	flags = append(flags, advancedFlags()...)
	return &cli.Command{
		Name:  "fuzz",
		Usage: "Generate and evaluate random deck combinations using Monte Carlo sampling",
		Commands: []*cli.Command{
			addDeckFuzzListCommand(),
			addDeckFuzzUpdateCommand(),
		},
		Flags:  flags,
		Action: deckFuzzCommand,
	}
}

// addDeckFuzzListCommand adds the fuzz list subcommand
func addDeckFuzzListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List saved top decks from storage",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "top",
				Value: 10,
				Usage: "Number of top decks to display",
			},
			&cli.StringFlag{
				Name:  "archetype",
				Usage: "Filter by archetype",
			},
			&cli.Float64Flag{
				Name:  "min-score",
				Usage: "Minimum overall score",
			},
			&cli.Float64Flag{
				Name:  "max-score",
				Usage: "Maximum overall score",
			},
			&cli.Float64Flag{
				Name:  "min-elixir",
				Usage: "Minimum average elixir",
			},
			&cli.Float64Flag{
				Name:  "max-elixir",
				Usage: "Maximum average elixir",
			},
			&cli.IntFlag{
				Name:  "max-same-archetype",
				Usage: "Maximum decks per archetype in returned results (0 = unlimited)",
				Value: 0,
			},
			&cli.StringFlag{
				Name:  "format",
				Value: "summary",
				Usage: "Output format: summary, json, csv, detailed",
			},
			playerTagFlagWithUsage(false, "Player tag (without #) to re-evaluate saved decks with your card levels"),
			&cli.StringFlag{
				Name:  "api-token",
				Usage: "Clash Royale API token (defaults to CLASH_ROYALE_API_TOKEN env var)",
			},
			&cli.IntFlag{
				Name:  "workers",
				Value: 1,
				Usage: "Number of parallel workers for player-specific re-evaluation",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show detailed progress information",
			},
		},
		Action: deckFuzzListCommand,
	}
}

// addDeckFuzzUpdateCommand adds the fuzz update subcommand
func addDeckFuzzUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Re-evaluate saved decks with current scoring and update storage",
		Flags: []cli.Flag{
			playerTagFlagWithUsage(false, "Player tag (without #) to apply level-aware scoring"),
			&cli.IntFlag{
				Name:  "top",
				Value: 0,
				Usage: "Maximum number of decks to update (0 = all)",
			},
			&cli.StringFlag{
				Name:  "archetype",
				Usage: "Filter by archetype",
			},
			&cli.Float64Flag{
				Name:  "min-score",
				Usage: "Minimum overall score",
			},
			&cli.Float64Flag{
				Name:  "max-score",
				Usage: "Maximum overall score",
			},
			&cli.Float64Flag{
				Name:  "min-elixir",
				Usage: "Minimum average elixir",
			},
			&cli.Float64Flag{
				Name:  "max-elixir",
				Usage: "Maximum average elixir",
			},
			&cli.IntFlag{
				Name:  "workers",
				Value: 1,
				Usage: "Number of parallel workers for re-evaluation",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show detailed progress information",
			},
		},
		Action: deckFuzzUpdateCommand,
	}
}
