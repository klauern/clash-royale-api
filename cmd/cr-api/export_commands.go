package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/klauer/clash-royale-api/go/internal/exporter/csv"
	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/urfave/cli/v3"
)

// Helper functions for export commands

// validateAPIToken validates and returns the API token from the command context
func validateAPIToken(cmd *cli.Command) (string, error) {
	apiToken := cmd.String("api-token")
	if apiToken == "" {
		return "", fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}
	return apiToken, nil
}

// createClient creates a new Clash Royale API client
func createClient(apiToken string) *clashroyale.Client {
	return clashroyale.NewClient(apiToken)
}

// ensureDir creates a directory with consistent error handling
func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

// exportFunc is a function that performs an export operation
type exportFunc func(dataDir string, data any) error

// exportWithFeedback wraps an export call with consistent error handling and success message
func exportWithFeedback(fn exportFunc, dataDir string, data any, description, targetFile string) error {
	if err := fn(dataDir, data); err != nil {
		return fmt.Errorf("failed to export %s: %w", description, err)
	}
	printf("✓ Exported %s to %s\n", description, targetFile)
	return nil
}

// extractEventBattles filters battles to include only non-ladder/training events
func extractEventBattles(battles []clashroyale.Battle) []clashroyale.Battle {
	var eventBattles []clashroyale.Battle
	for _, battle := range battles {
		if battle.GameMode.Name != "Ladder" && battle.GameMode.Name != "Training Camp" &&
			battle.GameMode.Name != "1v1" && battle.GameMode.Name != "2v2" {
			eventBattles = append(eventBattles, battle)
		}
	}
	return eventBattles
}

// moveExportFile moves a file from source to destination, falling back to copy+remove if rename fails
func moveExportFile(src, dst string) error {
	if err := os.Rename(src, dst); err != nil {
		// If rename fails, try copy
		content, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read source file: %w", err)
		}
		if err := os.WriteFile(dst, content, 0o644); err != nil {
			return fmt.Errorf("failed to write destination file: %w", err)
		}
		if err := os.Remove(src); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to remove source file: %w", err)
		}
	}
	return nil
}

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
					dataDir := cmd.String("data-dir")

					apiToken, err := validateAPIToken(cmd)
					if err != nil {
						return err
					}

					client := createClient(apiToken)

					// Get player information
					player, err := client.GetPlayerWithContext(ctx, tag)
					if err != nil {
						return fmt.Errorf("failed to get player data: %w", err)
					}

					// Create export directory
					exportDir := filepath.Join(dataDir, "csv", "players")
					if err := ensureDir(exportDir); err != nil {
						return fmt.Errorf("failed to create export directory: %w", err)
					}

					// Export based on requested types
					for _, exportType := range types {
						switch exportType {
						case "summary", "all":
							exporter := csv.NewPlayerExporter()
							if err := exportWithFeedback(func(ddir string, d any) error { return exporter.Export(ddir, d) }, dataDir, player, "player summary", filepath.Join(exportDir, "players.csv")); err != nil {
								return err
							}

						case "cards":
							exporter := csv.NewPlayerCardsExporter()
							if err := exportWithFeedback(func(ddir string, d any) error { return exporter.Export(ddir, d) }, dataDir, player, "player cards", filepath.Join(exportDir, "player_cards.csv")); err != nil {
								return err
							}
						}
					}

					printf("Successfully exported player data for %s\n", player.Name)
					return nil
				},
			},
			{
				Name:  "cards",
				Usage: "Export card database to CSV",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					dataDir := cmd.String("data-dir")

					apiToken, err := validateAPIToken(cmd)
					if err != nil {
						return err
					}

					client := createClient(apiToken)

					// Get all cards
					cardList, err := client.GetCardsWithContext(ctx)
					if err != nil {
						return fmt.Errorf("failed to get card database: %w", err)
					}
					cards := cardList.Items

					// Create export directory
					exportDir := filepath.Join(dataDir, "csv", "reference")
					if err := ensureDir(exportDir); err != nil {
						return fmt.Errorf("failed to create export directory: %w", err)
					}

					// Export cards
					exporter := csv.NewCardsExporter()
					if err := exportWithFeedback(func(ddir string, d any) error { return exporter.Export(ddir, d) }, dataDir, cards, fmt.Sprintf("%d cards", len(cards)), filepath.Join(exportDir, "cards.csv")); err != nil {
						return err
					}

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
					dataDir := cmd.String("data-dir")

					apiToken, err := validateAPIToken(cmd)
					if err != nil {
						return err
					}

					client := createClient(apiToken)

					// Get player information and analyze
					player, err := client.GetPlayerWithContext(ctx, tag)
					if err != nil {
						return fmt.Errorf("failed to get player data: %w", err)
					}

					// Get all cards for analysis context
					cardList, err := client.GetCardsWithContext(ctx)
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
					if err := ensureDir(exportDir); err != nil {
						return fmt.Errorf("failed to create export directory: %w", err)
					}

					// Export analysis
					exporter := csv.NewAnalysisExporter()
					if err := exportWithFeedback(func(ddir string, d any) error { return exporter.Export(ddir, d) }, dataDir, result, "collection analysis", filepath.Join(exportDir, "analysis.csv")); err != nil {
						return err
					}

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
					dataDir := cmd.String("data-dir")
					limit := cmd.Int("limit")

					apiToken, err := validateAPIToken(cmd)
					if err != nil {
						return err
					}

					client := createClient(apiToken)

					// Get battle log
					battleLog, err := client.GetPlayerBattleLogWithContext(ctx, tag)
					if err != nil {
						return fmt.Errorf("failed to get battle log: %w", err)
					}
					battles := []clashroyale.Battle(*battleLog)
					if limit > 0 && limit < len(battles) {
						battles = battles[:limit]
					}

					// Create export directory
					exportDir := filepath.Join(dataDir, "csv", "battles")
					if err := ensureDir(exportDir); err != nil {
						return fmt.Errorf("failed to create export directory: %w", err)
					}

					// Export battles
					exporter := csv.NewBattleLogExporter()
					if err := exportWithFeedback(func(ddir string, d any) error { return exporter.Export(ddir, d) }, dataDir, battles, fmt.Sprintf("%d battles", len(battles)), filepath.Join(exportDir, "battles.csv")); err != nil {
						return err
					}

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
					dataDir := cmd.String("data-dir")

					apiToken, err := validateAPIToken(cmd)
					if err != nil {
						return err
					}

					client := createClient(apiToken)

					// Get player's battle logs to extract event data
					battleLog, err := client.GetPlayerBattleLogWithContext(ctx, tag)
					if err != nil {
						return fmt.Errorf("failed to get battle log: %w", err)
					}
					battles := []clashroyale.Battle(*battleLog)

					// Extract event battles from battle log
					eventBattles := extractEventBattles(battles)

					// Create export directory
					exportDir := filepath.Join(dataDir, "csv", "events")
					if err := ensureDir(exportDir); err != nil {
						return fmt.Errorf("failed to create export directory: %w", err)
					}

					// Export event battles using battle log exporter
					exporter := csv.NewBattleLogExporter()
					if err := exporter.Export(dataDir, eventBattles); err != nil {
						return fmt.Errorf("failed to export event battles: %w", err)
					}

					// Move the output file to events directory
					battlesFile := filepath.Join(dataDir, "csv", "battles", "battle_log.csv")
					eventsFile := filepath.Join(exportDir, "event_battles.csv")
					if err := moveExportFile(battlesFile, eventsFile); err != nil {
						return err
					}

					printf("✓ Exported %d event battles to %s\n", len(eventBattles), eventsFile)
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
					dataDir := cmd.String("data-dir")

					apiToken, err := validateAPIToken(cmd)
					if err != nil {
						return err
					}

					client := createClient(apiToken)

					printf("Starting full export for player %s...\n\n", tag)

					// Get player data
					player, err := client.GetPlayerWithContext(ctx, tag)
					if err != nil {
						return fmt.Errorf("failed to get player data: %w", err)
					}

					// Get all cards for analysis
					cardList, err := client.GetCardsWithContext(ctx)
					if err != nil {
						return fmt.Errorf("failed to get card database: %w", err)
					}

					// Get battle log
					battleLog, err := client.GetPlayerBattleLogWithContext(ctx, tag)
					if err != nil {
						return fmt.Errorf("failed to get battle log: %w", err)
					}
					battles := []clashroyale.Battle(*battleLog)

					// 1. Export player summary and cards
					fmt.Println("1. Exporting player data...")
					playerExportDir := filepath.Join(dataDir, "csv", "players")
					if err := ensureDir(playerExportDir); err != nil {
						return fmt.Errorf("failed to create player export directory: %w", err)
					}

					playerExporter := csv.NewPlayerExporter()
					if err := exportWithFeedback(func(ddir string, d any) error { return playerExporter.Export(ddir, d) }, dataDir, player, "player summary", filepath.Join(playerExportDir, "players.csv")); err != nil {
						return err
					}

					playerCardsExporter := csv.NewPlayerCardsExporter()
					if err := exportWithFeedback(func(ddir string, d any) error { return playerCardsExporter.Export(ddir, d) }, dataDir, player, "player cards", filepath.Join(playerExportDir, "player_cards.csv")); err != nil {
						return err
					}

					// Note: Current deck is included in the main player export

					// 2. Export analysis
					fmt.Println("\n2. Analyzing collection...")
					options := analysis.DefaultAnalysisOptions()
					analysisResult, err := analysis.AnalyzeCardCollection(player, options)
					if err != nil {
						return fmt.Errorf("failed to analyze collection: %w", err)
					}

					analysisExportDir := filepath.Join(dataDir, "csv", "analysis")
					if err := ensureDir(analysisExportDir); err != nil {
						return fmt.Errorf("failed to create analysis export directory: %w", err)
					}

					analysisExporter := csv.NewAnalysisExporter()
					if err := exportWithFeedback(func(ddir string, d any) error { return analysisExporter.Export(ddir, d) }, dataDir, analysisResult, "collection analysis", filepath.Join(analysisExportDir, "analysis.csv")); err != nil {
						return err
					}

					// 3. Export battles
					fmt.Println("\n3. Exporting battle log...")
					battleExportDir := filepath.Join(dataDir, "csv", "battles")
					if err := ensureDir(battleExportDir); err != nil {
						return fmt.Errorf("failed to create battle export directory: %w", err)
					}

					battleExporter := csv.NewBattleLogExporter()
					if err := exportWithFeedback(func(ddir string, d any) error { return battleExporter.Export(ddir, d) }, dataDir, battles, fmt.Sprintf("battles (%d records)", len(battles)), filepath.Join(battleExportDir, "battles.csv")); err != nil {
						return err
					}

					// 4. Export event battles from battle log
					fmt.Println("\n4. Extracting event battles...")
					eventBattles := extractEventBattles(battles)

					if len(eventBattles) > 0 {
						eventExportDir := filepath.Join(dataDir, "csv", "events")
						if err := ensureDir(eventExportDir); err != nil {
							return fmt.Errorf("failed to create event export directory: %w", err)
						}

						// Use battle log exporter for event battles (simplified export)
						eventExporter := csv.NewBattleLogExporter()
						if err := eventExporter.Export(dataDir, eventBattles); err != nil {
							return fmt.Errorf("failed to export event battles: %w", err)
						}

						// Move the output file to events directory
						battlesFile := filepath.Join(dataDir, "csv", "battles", "battle_log.csv")
						eventsFile := filepath.Join(eventExportDir, "event_battles.csv")
						if err := moveExportFile(battlesFile, eventsFile); err != nil {
							return err
						}
						printf("   ✓ Event battles (%d records): %s\n", len(eventBattles), eventsFile)
					} else {
						fmt.Println("   ℹ No event battles found in recent battles")
					}

					// 5. Export card database
					fmt.Println("\n5. Exporting card database...")
					cardExportDir := filepath.Join(dataDir, "csv", "reference")
					if err := ensureDir(cardExportDir); err != nil {
						return fmt.Errorf("failed to create card export directory: %w", err)
					}

					cardExporter := csv.NewCardsExporter()
					if err := exportWithFeedback(func(ddir string, d any) error { return cardExporter.Export(ddir, d) }, dataDir, cardList.Items, "card database", filepath.Join(cardExportDir, "cards.csv")); err != nil {
						return err
					}

					// Summary
					printf("\n✅ Export complete for %s!\n", player.Name)
					printf("   Player: %s (%d trophies)\n", player.Name, player.Trophies)
					printf("   Cards: %d collected\n", len(player.Cards))
					printf("   Battles: %d recent games\n", len(battles))
					printf("   Location: %s\n", filepath.Join(dataDir, "csv"))

					return nil
				},
			},
		},
	}
}
