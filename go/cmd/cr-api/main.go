package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/klauer/clash-royale-api/go/internal/exporter/csv"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/urfave/cli/v3"
)

func main() {
	// Get default paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	defaultDataDir := filepath.Join(homeDir, ".cr-api")

	// Export manager will be created per command as needed

	// Create the CLI app
	cmd := &cli.Command{
		Name:  "cr-api",
		Usage: "Clash Royale API client and analysis tool",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "api-token",
				Aliases: []string{"t"},
				Usage:   "Clash Royale API token",
				Sources: cli.EnvVars("CLASH_ROYALE_API_TOKEN"),
			},
			&cli.StringFlag{
				Name:    "data-dir",
				Aliases: []string{"d"},
				Value:   defaultDataDir,
				Usage:   "Data storage directory",
				Sources: cli.EnvVars("DATA_DIR"),
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose logging",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "player",
				Usage: "Get player information",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "chests",
						Usage:   "Show upcoming chests",
					},
					&cli.BoolFlag{
						Name:    "save",
						Usage:   "Save player data to file",
					},
					&cli.BoolFlag{
						Name:    "export-csv",
						Usage:   "Export player data to CSV",
					},
				},
				Action: playerCommand,
			},
			{
				Name:  "cards",
				Usage: "Get card database",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "export-csv",
						Usage:   "Export card database to CSV",
					},
				},
				Action: cardsCommand,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func playerCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	showChests := cmd.Bool("chests")
	saveData := cmd.Bool("save")
	exportCSV := cmd.Bool("export-csv")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")

	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	client := clashroyale.NewClient(apiToken)

	if verbose {
		fmt.Printf("Getting player data for tag: %s\n", tag)
	}

	// Get player information
	player, err := client.GetPlayer(tag)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Display player info
	displayPlayerInfo(player)

	// Get and display chest cycle if requested
	if showChests {
		if verbose {
			fmt.Printf("\nFetching upcoming chests...\n")
		}
		chests, err := client.GetPlayerUpcomingChests(tag)
		if err != nil {
			fmt.Printf("Warning: Failed to get chests: %v\n", err)
		} else {
			displayUpcomingChests(chests)
		}
	}

	// Save player data if requested
	if saveData {
		dataDir := cmd.String("data-dir")
		if verbose {
			fmt.Printf("\nSaving player data to: %s\n", dataDir)
		}
		if err := savePlayerData(dataDir, player); err != nil {
			fmt.Printf("Warning: Failed to save player data: %v\n", err)
		} else {
			fmt.Printf("Player data saved to: %s/players/%s.json\n", dataDir, player.Tag)
		}
	}

	// Export to CSV if requested
	if exportCSV {
		dataDir := cmd.String("data-dir")
		if verbose {
			fmt.Printf("\nExporting player data to CSV...\n")
		}
		playerExporter := csv.NewPlayerExporter()
		if err := playerExporter.Export(dataDir, player); err != nil {
			fmt.Printf("Warning: Failed to export player data: %v\n", err)
		} else {
			fmt.Printf("Player data exported to CSV\n")
		}
	}

	return nil
}

func displayPlayerInfo(p *clashroyale.Player) {
	fmt.Printf("\nPlayer Information:\n")
	fmt.Printf("==================\n")
	fmt.Printf("Name: %s (%s)\n", p.Name, p.Tag)
	fmt.Printf("Level: %d (Experience: %d/%d)\n", p.ExpLevel, p.ExpPoints, p.Experience)
	fmt.Printf("Trophies: %d (Best: %d)\n", p.Trophies, p.BestTrophies)
	fmt.Printf("Arena: %s\n", p.Arena.Name)
	fmt.Printf("League: %s\n", p.League.Name)

	if p.Clan != nil {
		fmt.Printf("\nClan Information:\n")
		fmt.Printf("Clan: %s (%s)\n", p.Clan.Name, p.Clan.Tag)
		fmt.Printf("Role: %s\n", p.Role)
		fmt.Printf("Clan Trophies: %d\n", p.Clan.ClanScore)
	}

	fmt.Printf("\nBattle Statistics:\n")
	if p.Wins+p.Losses > 0 {
		fmt.Printf("Wins: %d | Losses: %d | Win Rate: %.1f%%\n",
			p.Wins, p.Losses,
			float64(p.Wins)/float64(p.Wins+p.Losses)*100)
	} else {
		fmt.Printf("Wins: %d | Losses: %d\n", p.Wins, p.Losses)
	}
	fmt.Printf("3-Crown Wins: %d\n", p.ThreeCrownWins)
	fmt.Printf("Total Battles: %d\n", p.BattleCount)

	fmt.Printf("\nCard Collection:\n")
	fmt.Printf("Total Cards: %d\n", len(p.Cards))
	fmt.Printf("Star Points: %d\n", p.StarPoints)
}

func displayUpcomingChests(chests *clashroyale.ChestCycle) {
	fmt.Printf("\n\nUpcoming Chests:\n")
	fmt.Printf("================\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Slot\tChest Name\n")
	fmt.Fprintf(w, "----\t----------\n")

	for i, chest := range chests.Items {
		fmt.Fprintf(w, "%d\t%s\n", chest.Index+1, chest.Name)
		if i >= 9 { // Show first 10 chests
			break
		}
	}

	w.Flush()
}

func savePlayerData(dataDir string, p *clashroyale.Player) error {
	// Create data directory if it doesn't exist
	playersDir := filepath.Join(dataDir, "players")
	if err := os.MkdirAll(playersDir, 0755); err != nil {
		return fmt.Errorf("failed to create players directory: %w", err)
	}

	// Save as JSON
	filename := filepath.Join(playersDir, fmt.Sprintf("%s.json", p.Tag))
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal player data: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write player file: %w", err)
	}

	return nil
}

func cardsCommand(ctx context.Context, cmd *cli.Command) error {
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	exportCSV := cmd.Bool("export-csv")

	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	client := clashroyale.NewClient(apiToken)

	if verbose {
		fmt.Printf("Fetching card database...\n")
	}

	cards, err := client.GetCards()
	if err != nil {
		return fmt.Errorf("failed to get cards: %w", err)
	}

	// Always display cards unless only exporting
	if !exportCSV {
		displayCards(*cards)
	}

	// Export to CSV if requested
	if exportCSV {
		dataDir := cmd.String("data-dir")
		if verbose {
			fmt.Printf("\nExporting card database to CSV...\n")
		}
		cardsExporter := csv.NewCardsExporter()
		if err := cardsExporter.Export(dataDir, *cards); err != nil {
			fmt.Printf("Warning: Failed to export cards: %v\n", err)
		} else {
			fmt.Printf("Card database exported to CSV\n")
		}
	}

	return nil
}

func displayCards(cards []clashroyale.Card) {
	fmt.Printf("\nCard Database:\n")
	fmt.Printf("=============\n")
	fmt.Printf("Total Cards: %d\n\n", len(cards))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Name\tRarity\tElixir\tType\n")
	fmt.Fprintf(w, "----\t------\t------\t----\n")

	for _, card := range cards {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
			card.Name,
			card.Rarity,
			card.ElixirCost,
			card.Type)
	}

	w.Flush()
}