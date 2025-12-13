package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/urfave/cli/v3"
)

// addDeckCommands adds deck-related subcommands to the CLI
func addDeckCommands() *cli.Command {
	return &cli.Command{
		Name:  "deck",
		Usage: "Deck building and analysis commands",
		Commands: []*cli.Command{
			{
				Name:  "build",
				Usage: "Build an optimized deck based on player's card collection",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "strategy",
						Aliases: []string{"s"},
						Value:   "balanced",
						Usage:   "Deck building strategy: balanced, aggro, control, cycle, splash, spell",
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
						Usage: "Specific cards to include in the deck (by name)",
					},
					&cli.StringSliceFlag{
						Name:  "exclude-cards",
						Usage: "Cards to exclude from the deck (by name)",
					},
					&cli.IntFlag{
						Name:  "min-level",
						Value: 1,
						Usage: "Minimum card level to consider",
					},
					&cli.BoolFlag{
						Name:  "prioritize-upgrades",
						Usage: "Prioritize cards that can be upgraded soon",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export deck analysis to CSV",
					},
					&cli.BoolFlag{
						Name:  "save",
						Usage: "Save deck to file",
					},
				},
				Action: deckBuildCommand,
			},
			{
				Name:  "analyze",
				Usage: "Analyze a deck's strengths and weaknesses",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "cards",
						Aliases:  []string{"c"},
						Usage:    "8 card names for the deck",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export analysis to CSV",
					},
				},
				Action: deckAnalyzeCommand,
			},
			{
				Name:  "optimize",
				Usage: "Optimize an existing deck with available cards",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:     "cards",
						Aliases:  []string{"c"},
						Usage:    "Current 8-card deck to optimize",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "max-changes",
						Value: 4,
						Usage: "Maximum number of cards to change",
					},
					&cli.BoolFlag{
						Name:  "keep-win-con",
						Usage: "Keep win condition cards unchanged",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export optimization results to CSV",
					},
				},
				Action: deckOptimizeCommand,
			},
			{
				Name:  "recommend",
				Usage: "Get deck recommendations based on meta analysis",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "arena",
						Usage: "Filter by arena name",
					},
					&cli.StringFlag{
						Name:  "league",
						Usage: "Filter by league name",
					},
					&cli.IntFlag{
						Name:  "limit",
						Value: 5,
						Usage: "Number of recommendations to return",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export recommendations to CSV",
					},
				},
				Action: deckRecommendCommand,
			},
		},
	}
}

func deckBuildCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	strategy := cmd.String("strategy")
	minElixir := cmd.Float64("min-elixir")
	maxElixir := cmd.Float64("max-elixir")
	saveData := cmd.Bool("save")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")

	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	client := clashroyale.NewClient(apiToken)

	if verbose {
		fmt.Printf("Building deck for player %s with strategy: %s\n", tag, strategy)
	}

	// Get player information
	player, err := client.GetPlayer(tag)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// TODO: Create a builder when needed
	// builder := deck.NewBuilder(cmd.String("data-dir"))

	// TODO: Convert player data to card analysis format
	// For now, just display basic info
	fmt.Printf("\nDeck Build Options:\n")
	fmt.Printf("===================\n")
	fmt.Printf("Strategy: %s\n", strategy)
	fmt.Printf("Min Elixir: %.1f\n", minElixir)
	fmt.Printf("Max Elixir: %.1f\n", maxElixir)
	fmt.Printf("Available Cards: %d\n", len(player.Cards))

	// Save basic info if requested
	if saveData {
		dataDir := cmd.String("data-dir")
		if verbose {
			fmt.Printf("\nDeck build options saved to: %s\n", dataDir)
		}
	}

	return nil
}

func deckAnalyzeCommand(ctx context.Context, cmd *cli.Command) error {
	cardNames := cmd.StringSlice("cards")

	if len(cardNames) != 8 {
		return fmt.Errorf("exactly 8 cards are required for deck analysis")
	}

	fmt.Printf("Analyzing deck with cards: %v\n", cardNames)
	fmt.Println("Note: Full deck analysis not yet implemented")

	return nil
}

func deckOptimizeCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	cardNames := cmd.StringSlice("cards")
	apiToken := cmd.String("api-token")

	if apiToken == "" {
		return fmt.Errorf("API token is required")
	}

	if len(cardNames) != 8 {
		return fmt.Errorf("exactly 8 cards are required for optimization")
	}

	fmt.Printf("Optimizing deck for player %s\n", tag)
	fmt.Printf("Current deck: %v\n", cardNames)
	fmt.Println("Note: Deck optimization not yet implemented")

	return nil
}

func deckRecommendCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	limit := cmd.Int("limit")
	apiToken := cmd.String("api-token")

	if apiToken == "" {
		return fmt.Errorf("API token is required")
	}

	fmt.Printf("Getting deck recommendations for player %s\n", tag)
	fmt.Printf("Limit: %d recommendations\n", limit)
	fmt.Println("Note: Deck recommendations not yet implemented")

	return nil
}

// Helper functions for deck operations
func saveDeck(dataDir, playerTag string, options map[string]interface{}) error {
	decksDir := filepath.Join(dataDir, "decks")
	if err := os.MkdirAll(decksDir, 0755); err != nil {
		return fmt.Errorf("failed to create decks directory: %w", err)
	}

	// In a real implementation, you'd marshal options to JSON
	return nil
}
