package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/klauer/clash-royale-api/go/internal/exporter/csv"
	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/urfave/cli/v3"
)

// addExportCommands adds export-related subcommands to the CLI
func addExportCommands() *cli.Command {
	return &cli.Command{
		Name:  "export",
		Usage: "Export various data types to CSV format",
		Commands: []*cli.Command{
			{
				Name:  "player",
				Usage: "Export player data to CSV",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:  "types",
						Value: []string{"summary"},
						Usage: "Export types: summary,cards,all",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					tag := cmd.String("tag")
					types := cmd.StringSlice("types")
					apiToken := cmd.String("api-token")
					dataDir := cmd.String("data-dir")

					if apiToken == "" {
						return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
					}

					client := clashroyale.NewClient(apiToken)

					// Get player information
					player, err := client.GetPlayer(tag)
					if err != nil {
						return fmt.Errorf("failed to get player data: %w", err)
					}

					// Create export directory
					exportDir := filepath.Join(dataDir, "csv", "players")
					if err := os.MkdirAll(exportDir, 0o755); err != nil {
						return fmt.Errorf("failed to create export directory: %w", err)
					}

					// Export based on requested types
					for _, exportType := range types {
						switch exportType {
						case "summary", "all":
							exporter := csv.NewPlayerExporter()
							if err := exporter.Export(dataDir, player); err != nil {
								return fmt.Errorf("failed to export player summary: %w", err)
							}
							fmt.Printf("✓ Exported player summary to %s\n", filepath.Join(exportDir, "players.csv"))

						case "cards":
							exporter := csv.NewPlayerCardsExporter()
							if err := exporter.Export(dataDir, player); err != nil {
								return fmt.Errorf("failed to export player cards: %w", err)
							}
							fmt.Printf("✓ Exported player cards to %s\n", filepath.Join(exportDir, "player_cards.csv"))
						}
					}

					fmt.Printf("Successfully exported player data for %s\n", player.Name)
					return nil
				},
			},
			{
				Name:  "cards",
				Usage: "Export card database to CSV",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					apiToken := cmd.String("api-token")
					dataDir := cmd.String("data-dir")

					if apiToken == "" {
						return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
					}

					client := clashroyale.NewClient(apiToken)

					// Get all cards
					cardList, err := client.GetCards()
					if err != nil {
						return fmt.Errorf("failed to get card database: %w", err)
					}
					cards := cardList.Items

					// Create export directory
					exportDir := filepath.Join(dataDir, "csv", "reference")
					if err := os.MkdirAll(exportDir, 0o755); err != nil {
						return fmt.Errorf("failed to create export directory: %w", err)
					}

					// Export cards
					exporter := csv.NewCardsExporter()
					if err := exporter.Export(dataDir, cardList); err != nil {
						return fmt.Errorf("failed to export cards: %w", err)
					}

					fmt.Printf("✓ Exported %d cards to %s\n", len(cards), filepath.Join(exportDir, "cards.csv"))
					return nil
				},
			},
			{
				Name:  "analysis",
				Usage: "Export card collection analysis to CSV",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					tag := cmd.String("tag")
					apiToken := cmd.String("api-token")
					dataDir := cmd.String("data-dir")

					if apiToken == "" {
						return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
					}

					client := clashroyale.NewClient(apiToken)

					// Get player information and analyze
					player, err := client.GetPlayer(tag)
					if err != nil {
						return fmt.Errorf("failed to get player data: %w", err)
					}

					// Get all cards for analysis context
					cardList, err := client.GetCards()
					if err != nil {
						return fmt.Errorf("failed to get card database: %w", err)
					}
					_ = cardList.Items // Cards reference for future use

					// Analyze collection with default options
					options := analysis.DefaultAnalysisOptions()
					result, err := analysis.AnalyzeCardCollection(player, options)
					if err != nil {
						return fmt.Errorf("failed to analyze collection: %w", err)
					}

					// Create export directory
					exportDir := filepath.Join(dataDir, "csv", "analysis")
					if err := os.MkdirAll(exportDir, 0o755); err != nil {
						return fmt.Errorf("failed to create export directory: %w", err)
					}

					// Export analysis
					exporter := csv.NewAnalysisExporter()
					if err := exporter.Export(dataDir, result); err != nil {
						return fmt.Errorf("failed to export analysis: %w", err)
					}

					fmt.Printf("✓ Exported collection analysis to %s\n", filepath.Join(exportDir, "analysis.csv"))
					return nil
				},
			},
			{
				Name:  "battles",
				Usage: "Export battle log to CSV",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "limit",
						Value: 0,
						Usage: "Maximum number of battles to export (0 for all)",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					tag := cmd.String("tag")
					apiToken := cmd.String("api-token")
					dataDir := cmd.String("data-dir")
					limit := cmd.Int("limit")

					if apiToken == "" {
						return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
					}

					client := clashroyale.NewClient(apiToken)

					// Get battle log
					battleLog, err := client.GetPlayerBattleLog(tag)
					if err != nil {
						return fmt.Errorf("failed to get battle log: %w", err)
					}
					battles := []clashroyale.Battle(*battleLog)
					if limit > 0 && limit < len(battles) {
						battles = battles[:limit]
					}

					// Create export directory
					exportDir := filepath.Join(dataDir, "csv", "battles")
					if err := os.MkdirAll(exportDir, 0o755); err != nil {
						return fmt.Errorf("failed to create export directory: %w", err)
					}

					// Export battles
					exporter := csv.NewBattleLogExporter()
					if err := exporter.Export(dataDir, battles); err != nil {
						return fmt.Errorf("failed to export battles: %w", err)
					}

					fmt.Printf("✓ Exported %d battles to %s\n", len(battles), filepath.Join(exportDir, "battles.csv"))
					return nil
				},
			},
			{
				Name:  "events",
				Usage: "Export event deck data to CSV",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					tag := cmd.String("tag")
					apiToken := cmd.String("api-token")
					dataDir := cmd.String("data-dir")

					if apiToken == "" {
						return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
					}

					client := clashroyale.NewClient(apiToken)

					// Get player's battle logs to extract event data
					battleLog, err := client.GetPlayerBattleLog(tag)
					if err != nil {
						return fmt.Errorf("failed to get battle log: %w", err)
					}
					battles := []clashroyale.Battle(*battleLog)

					// Extract event battles from battle log
					var eventBattles []clashroyale.Battle
					for _, battle := range battles {
						// Consider any battle from special events as an event battle
						if battle.GameMode.Name != "Ladder" && battle.GameMode.Name != "Training Camp" &&
							battle.GameMode.Name != "1v1" && battle.GameMode.Name != "2v2" {
							eventBattles = append(eventBattles, battle)
						}
					}

					// Create export directory
					exportDir := filepath.Join(dataDir, "csv", "events")
					if err := os.MkdirAll(exportDir, 0o755); err != nil {
						return fmt.Errorf("failed to create export directory: %w", err)
					}

					// Export event battles
					exporter := csv.NewEventBattlesExporter()
					if err := exporter.Export(dataDir, eventBattles); err != nil {
						return fmt.Errorf("failed to export events: %w", err)
					}

					fmt.Printf("✓ Exported %d event battles to %s\n", len(eventBattles), filepath.Join(exportDir, "events.csv"))
					return nil
				},
			},
			{
				Name:  "all",
				Usage: "Export all available data for a player",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "timestamp",
						Usage: "Include timestamp in exported filenames",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					tag := cmd.String("tag")
					apiToken := cmd.String("api-token")
					dataDir := cmd.String("data-dir")

					if apiToken == "" {
						return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
					}

					client := clashroyale.NewClient(apiToken)

					fmt.Printf("Starting full export for player %s...\n\n", tag)

					// Get player data
					player, err := client.GetPlayer(tag)
					if err != nil {
						return fmt.Errorf("failed to get player data: %w", err)
					}

					// Get all cards for analysis
					cardList, err := client.GetCards()
					if err != nil {
						return fmt.Errorf("failed to get card database: %w", err)
					}
					_ = cardList.Items // Cards reference for future use

					// Get battle log
					battleLog, err := client.GetPlayerBattleLog(tag)
					if err != nil {
						return fmt.Errorf("failed to get battle log: %w", err)
					}
					battles := []clashroyale.Battle(*battleLog)

					// 1. Export player summary and cards
					fmt.Println("1. Exporting player data...")
					playerExportDir := filepath.Join(dataDir, "csv", "players")
					if err := os.MkdirAll(playerExportDir, 0o755); err != nil {
						return fmt.Errorf("failed to create player export directory: %w", err)
					}

					playerExporter := csv.NewPlayerExporter()
					if err := playerExporter.Export(dataDir, player); err != nil {
						return fmt.Errorf("failed to export player summary: %w", err)
					}
					fmt.Printf("   ✓ Player summary: %s\n", filepath.Join(playerExportDir, "players.csv"))

					playerCardsExporter := csv.NewPlayerCardsExporter()
					if err := playerCardsExporter.Export(dataDir, player); err != nil {
						return fmt.Errorf("failed to export player cards: %w", err)
					}
					fmt.Printf("   ✓ Player cards: %s\n", filepath.Join(playerExportDir, "player_cards.csv"))

					// Note: Current deck is included in the main player export

					// 2. Export analysis
					fmt.Println("\n2. Analyzing collection...")
					options := analysis.DefaultAnalysisOptions()
					analysisResult, err := analysis.AnalyzeCardCollection(player, options)
					if err != nil {
						return fmt.Errorf("failed to analyze collection: %w", err)
					}

					analysisExportDir := filepath.Join(dataDir, "csv", "analysis")
					if err := os.MkdirAll(analysisExportDir, 0o755); err != nil {
						return fmt.Errorf("failed to create analysis export directory: %w", err)
					}

					analysisExporter := csv.NewAnalysisExporter()
					if err := analysisExporter.Export(dataDir, analysisResult); err != nil {
						return fmt.Errorf("failed to export analysis: %w", err)
					}
					fmt.Printf("   ✓ Collection analysis: %s\n", filepath.Join(analysisExportDir, "analysis.csv"))

					// 3. Export battles
					fmt.Println("\n3. Exporting battle log...")
					battleExportDir := filepath.Join(dataDir, "csv", "battles")
					if err := os.MkdirAll(battleExportDir, 0o755); err != nil {
						return fmt.Errorf("failed to create battle export directory: %w", err)
					}

					battleExporter := csv.NewBattleLogExporter()
					if err := battleExporter.Export(dataDir, battles); err != nil {
						return fmt.Errorf("failed to export battles: %w", err)
					}
					fmt.Printf("   ✓ Battles (%d records): %s\n", len(battles), filepath.Join(battleExportDir, "battles.csv"))

					// 4. Export event battles from battle log
					fmt.Println("\n4. Extracting event battles...")
					var eventBattles []clashroyale.Battle
					for _, battle := range battles {
						// Consider any battle from special events as an event battle
						if battle.GameMode.Name != "Ladder" && battle.GameMode.Name != "Training Camp" &&
							battle.GameMode.Name != "1v1" && battle.GameMode.Name != "2v2" {
							eventBattles = append(eventBattles, battle)
						}
					}

					if len(eventBattles) > 0 {
						eventExportDir := filepath.Join(dataDir, "csv", "events")
						if err := os.MkdirAll(eventExportDir, 0o755); err != nil {
							return fmt.Errorf("failed to create event export directory: %w", err)
						}

						eventExporter := csv.NewEventBattlesExporter()
						if err := eventExporter.Export(dataDir, eventBattles); err != nil {
							return fmt.Errorf("failed to export events: %w", err)
						}
						fmt.Printf("   ✓ Event battles (%d records): %s\n", len(eventBattles), filepath.Join(eventExportDir, "events.csv"))
					} else {
						fmt.Println("   ℹ No event battles found in recent battles")
					}

					// 5. Export card database (only if not already exists)
					fmt.Println("\n5. Exporting card database...")
					cardExportDir := filepath.Join(dataDir, "csv", "reference")
					if err := os.MkdirAll(cardExportDir, 0o755); err != nil {
						return fmt.Errorf("failed to create card export directory: %w", err)
					}

					cardExporter := csv.NewCardsExporter()
					if err := cardExporter.Export(dataDir, cardList); err != nil {
						return fmt.Errorf("failed to export cards: %w", err)
					}
					fmt.Printf("   ✓ Card database: %s\n", filepath.Join(cardExportDir, "cards.csv"))

					// Summary
					fmt.Printf("\n✅ Export complete for %s!\n", player.Name)
					fmt.Printf("   Player: %s (%d trophies)\n", player.Name, player.Trophies)
					fmt.Printf("   Cards: %d collected\n", len(player.Cards))
					fmt.Printf("   Battles: %d recent games\n", len(battles))
					fmt.Printf("   Location: %s\n", filepath.Join(dataDir, "csv"))

					return nil
				},
			},
		},
	}
}
