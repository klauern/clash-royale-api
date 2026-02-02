package main

import (
	"github.com/urfave/cli/v3"
)

// addDeckEvaluateCommand adds the deck evaluate command
func addDeckEvaluateCommand() *cli.Command {
	return &cli.Command{
		Name:  "evaluate",
		Usage: "Evaluate a deck with comprehensive analysis and scoring",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "deck",
				Usage: "Deck string (8 cards separated by dashes, e.g., Knight-Archers-Fireball-...)",
			},
			&cli.StringFlag{
				Name:    "tag",
				Aliases: []string{"p"},
				Usage:   "Player tag (without #) for card level context and upgrade impact analysis",
			},
			&cli.StringFlag{
				Name:  "from-analysis",
				Usage: "Load deck from analysis JSON file",
			},
			&cli.IntFlag{
				Name:  "arena",
				Value: 0,
				Usage: "Arena level for card unlock context (0 = no restriction)",
			},
			&cli.StringFlag{
				Name:  "format",
				Value: "human",
				Usage: "Output format: human, json, csv, detailed",
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Output file path (optional, prints to stdout if not specified)",
			},
			&cli.BoolFlag{
				Name:  "show-upgrade-impact",
				Usage: "Show upgrade impact analysis and recommendations (requires --tag)",
			},
			&cli.IntFlag{
				Name:  "top-upgrades",
				Value: 5,
				Usage: "Number of top upgrades to show in upgrade impact analysis",
			},
		},
		Action: deckEvaluateCommand,
	}
}

// addDeckEvaluateBatchCommand adds the deck evaluate-batch command
func addDeckEvaluateBatchCommand() *cli.Command {
	return &cli.Command{
		Name:  "evaluate-batch",
		Usage: "Evaluate multiple decks from a suite or directory",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "from-suite",
				Usage: "Path to deck suite summary JSON file (from build-suite command)",
			},
			&cli.StringFlag{
				Name:  "deck-dir",
				Usage: "Directory containing deck JSON files",
			},
			&cli.StringFlag{
				Name:    "tag",
				Aliases: []string{"p"},
				Usage:   "Player tag (without #) for context-aware evaluation",
			},
			&cli.StringFlag{
				Name:  "format",
				Value: "summary",
				Usage: "Output format: summary, json, csv, detailed",
			},
			&cli.StringFlag{
				Name:  "output-dir",
				Usage: "Output directory for evaluation files (default: prints to stdout)",
			},
			&cli.StringFlag{
				Name:  "sort-by",
				Value: "overall",
				Usage: "Sort results by: overall, attack, defense, synergy, versatility, f2p, playability, elixir",
			},
			&cli.BoolFlag{
				Name:  "top-only",
				Usage: "Show only top N decks",
			},
			&cli.IntFlag{
				Name:  "top-n",
				Value: 10,
				Usage: "Number of top decks to show (with --top-only)",
			},
			&cli.BoolFlag{
				Name:  "filter-archetype",
				Usage: "Filter by specific archetype (use with --archetype)",
			},
			&cli.StringFlag{
				Name:  "archetype",
				Usage: "Archetype to filter by (e.g., beatdown, control, cycle)",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show detailed progress information",
			},
			&cli.BoolFlag{
				Name:  "timing",
				Usage: "Show timing information for each deck evaluation",
			},
			&cli.BoolFlag{
				Name:  "save-aggregated",
				Value: true,
				Usage: "Save aggregated results to output-dir (default: true)",
			},
		},
		Action: deckEvaluateBatchCommand,
	}
}
