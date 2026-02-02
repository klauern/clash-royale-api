package main

import (
	"github.com/urfave/cli/v3"
)

// addDeckAnalyzeCommand adds the deck analyze command
func addDeckAnalyzeCommand() *cli.Command {
	return &cli.Command{
		Name:  "analyze",
		Usage: "Analyze deck strengths, weaknesses, and archetype classification",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "deck",
				Usage:    "Deck string (8 cards separated by dashes)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "format",
				Value: "human",
				Usage: "Output format: human, json",
			},
		},
		Action: deckAnalyzeCommand,
	}
}

// addDeckOptimizeCommand adds the deck optimize command
func addDeckOptimizeCommand() *cli.Command {
	return &cli.Command{
		Name:  "optimize",
		Usage: "Optimize an existing deck by suggesting card replacements",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "deck",
				Usage:    "Current deck string (8 cards separated by dashes)",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "tag",
				Aliases: []string{"p"},
				Usage:   "Player tag (without #) for card level context",
			},
			&cli.IntFlag{
				Name:  "suggestions",
				Value: 3,
				Usage: "Number of optimization suggestions to show",
			},
			&cli.StringFlag{
				Name:  "focus",
				Value: "balanced",
				Usage: "Optimization focus: balanced, attack, defense, synergy",
			},
			&cli.BoolFlag{
				Name:  "export-csv",
				Usage: "Export optimization suggestions to CSV",
			},
		},
		Action: deckOptimizeCommand,
	}
}

// addDeckRecommendCommand adds the deck recommend command
func addDeckRecommendCommand() *cli.Command {
	return &cli.Command{
		Name:  "recommend",
		Usage: "Get meta-based deck recommendations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "tag",
				Aliases: []string{"p"},
				Usage:   "Player tag (without #) for personalized recommendations",
			},
			&cli.StringFlag{
				Name:  "archetype",
				Usage: "Preferred archetype (beatdown, control, cycle, siege, bait, hybrid)",
			},
			&cli.IntFlag{
				Name:  "count",
				Value: 5,
				Usage: "Number of recommendations to show",
			},
			&cli.BoolFlag{
				Name:  "include-unowned",
				Usage: "Include cards not in player's collection",
			},
			&cli.BoolFlag{
				Name:  "export-csv",
				Usage: "Export recommendations to CSV",
			},
		},
		Action: deckRecommendCommand,
	}
}
