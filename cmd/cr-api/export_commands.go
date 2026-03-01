package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/exporter/csv"
	"github.com/klauer/clash-royale-api/go/internal/storage"
	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/events"
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

func exportTargetFile(exportDir, filename string) string {
	return filepath.Join(exportDir, filename)
}

// extractEventBattles filters battles to include only non-ladder/training events
func extractEventBattles(battles []clashroyale.Battle) []clashroyale.Battle {
	return events.FilterEventBattles(battles)
}

func timestampValue(enabled bool) string {
	if !enabled {
		return ""
	}
	return time.Now().Format("20060102_150405")
}

func appendTimestampToFilename(path, timestamp string) string {
	if timestamp == "" {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	return base + "_" + timestamp + ext
}

func applyTimestampToExport(path, timestamp string) (string, error) {
	if timestamp == "" {
		return path, nil
	}
	target := appendTimestampToFilename(path, timestamp)
	if err := storage.MoveFile(path, target); err != nil {
		return "", fmt.Errorf("failed to timestamp export %s: %w", path, err)
	}
	return target, nil
}

func exportPlayerType(exportType, dataDir, exportDir string, player *clashroyale.Player) error {
	exportSummary := func() error {
		exporter := csv.NewPlayerExporter()
		return exportWithFeedback(func(ddir string, d any) error { return exporter.Export(ddir, d) }, dataDir, player, "player summary", exportTargetFile(exportDir, exporter.Filename()))
	}
	exportCards := func() error {
		exporter := csv.NewPlayerCardsExporter()
		return exportWithFeedback(func(ddir string, d any) error { return exporter.Export(ddir, d) }, dataDir, player, "player cards", exportTargetFile(exportDir, exporter.Filename()))
	}

	switch exportType {
	case "summary":
		return exportSummary()
	case "cards":
		return exportCards()
	case "all":
		if err := exportSummary(); err != nil {
			return err
		}
		return exportCards()
	default:
		return nil
	}
}

func exportPlayerCommand() *cli.Command {
	return &cli.Command{
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
			pathBuilder := storage.NewPathBuilder(dataDir)
			exportDir := pathBuilder.GetCSVPlayersDir()
			if err := storage.EnsureDirectory(exportDir); err != nil {
				return fmt.Errorf("failed to create export directory: %w", err)
			}

			// Export based on requested types
			for _, exportType := range types {
				if err := exportPlayerType(exportType, dataDir, exportDir, player); err != nil {
					return err
				}
			}

			printf("Successfully exported player data for %s\n", player.Name)
			return nil
		},
	}
}

func exportCardsCommand() *cli.Command {
	return &cli.Command{
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
			pathBuilder := storage.NewPathBuilder(dataDir)
			exportDir := pathBuilder.GetCSVReferenceDir()
			if err := storage.EnsureDirectory(exportDir); err != nil {
				return fmt.Errorf("failed to create export directory: %w", err)
			}

			// Export cards
			exporter := csv.NewCardsExporter()
			if err := exportWithFeedback(func(ddir string, d any) error { return exporter.Export(ddir, d) }, dataDir, cards, fmt.Sprintf("%d cards", len(cards)), exportTargetFile(exportDir, exporter.Filename())); err != nil {
				return err
			}

			return nil
		},
	}
}

func exportAnalysisCommand() *cli.Command {
	return &cli.Command{
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

			// Analyze collection with default options
			options := analysis.DefaultAnalysisOptions()
			result, err := analysis.AnalyzeCardCollection(player, options)
			if err != nil {
				return fmt.Errorf("failed to analyze collection: %w", err)
			}

			// Create export directory
			pathBuilder := storage.NewPathBuilder(dataDir)
			exportDir := pathBuilder.GetCSVAnalysisDir()
			if err := storage.EnsureDirectory(exportDir); err != nil {
				return fmt.Errorf("failed to create export directory: %w", err)
			}

			// Export analysis
			exporter := csv.NewAnalysisExporter()
			if err := exportWithFeedback(func(ddir string, d any) error { return exporter.Export(ddir, d) }, dataDir, result, "collection analysis", exportTargetFile(exportDir, exporter.Filename())); err != nil {
				return err
			}

			return nil
		},
	}
}

func exportBattlesCommand() *cli.Command {
	return &cli.Command{
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
			pathBuilder := storage.NewPathBuilder(dataDir)
			exportDir := filepath.Join(pathBuilder.GetCSVDir(), storage.CSVBattlesSubdir)
			if err := storage.EnsureDirectory(exportDir); err != nil {
				return fmt.Errorf("failed to create export directory: %w", err)
			}

			// Export battles
			exporter := csv.NewBattleLogExporter()
			if err := exportWithFeedback(func(ddir string, d any) error { return exporter.Export(ddir, d) }, dataDir, battles, fmt.Sprintf("%d battles", len(battles)), exportTargetFile(exportDir, exporter.Filename())); err != nil {
				return err
			}

			return nil
		},
	}
}

func exportEventsCommand() *cli.Command {
	return &cli.Command{
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
			pathBuilder := storage.NewPathBuilder(dataDir)
			exportDir := pathBuilder.GetCSVEventsDir()
			if err := storage.EnsureDirectory(exportDir); err != nil {
				return fmt.Errorf("failed to create export directory: %w", err)
			}

			// Export event battles using battle log exporter
			exporter := csv.NewBattleLogExporter()
			if err := exporter.Export(dataDir, eventBattles); err != nil {
				return fmt.Errorf("failed to export event battles: %w", err)
			}

			// Move the output file to events directory
			battlesFile := filepath.Join(pathBuilder.GetCSVDir(), storage.CSVBattlesSubdir, "battle_log.csv")
			eventsFile := filepath.Join(exportDir, "event_battles.csv")
			if err := storage.MoveFile(battlesFile, eventsFile); err != nil {
				return err
			}

			printf("✓ Exported %d event battles to %s\n", len(eventBattles), eventsFile)
			return nil
		},
	}
}

type exportAllData struct {
	player   *clashroyale.Player
	cardList *clashroyale.CardList
	battles  []clashroyale.Battle
}

func loadExportAllData(ctx context.Context, client *clashroyale.Client, tag string) (exportAllData, error) {
	player, err := client.GetPlayerWithContext(ctx, tag)
	if err != nil {
		return exportAllData{}, fmt.Errorf("failed to get player data: %w", err)
	}

	cardList, err := client.GetCardsWithContext(ctx)
	if err != nil {
		return exportAllData{}, fmt.Errorf("failed to get card database: %w", err)
	}

	battleLog, err := client.GetPlayerBattleLogWithContext(ctx, tag)
	if err != nil {
		return exportAllData{}, fmt.Errorf("failed to get battle log: %w", err)
	}
	battles := []clashroyale.Battle(*battleLog)

	return exportAllData{
		player:   player,
		cardList: cardList,
		battles:  battles,
	}, nil
}

func exportAllPlayerData(dataDir string, player *clashroyale.Player, timestamp string) error {
	fmt.Println("1. Exporting player data...")
	pathBuilder := storage.NewPathBuilder(dataDir)
	playerExportDir := pathBuilder.GetCSVPlayersDir()
	if err := storage.EnsureDirectory(playerExportDir); err != nil {
		return fmt.Errorf("failed to create player export directory: %w", err)
	}

	playerExporter := csv.NewPlayerExporter()
	playerSummaryFile := exportTargetFile(playerExportDir, playerExporter.Filename())
	if err := exportWithFeedback(func(ddir string, d any) error { return playerExporter.Export(ddir, d) }, dataDir, player, "player summary", playerSummaryFile); err != nil {
		return err
	}
	timestampedSummaryFile, err := applyTimestampToExport(playerSummaryFile, timestamp)
	if err != nil {
		return err
	}

	playerCardsExporter := csv.NewPlayerCardsExporter()
	playerCardsFile := exportTargetFile(playerExportDir, playerCardsExporter.Filename())
	if err := exportWithFeedback(func(ddir string, d any) error { return playerCardsExporter.Export(ddir, d) }, dataDir, player, "player cards", playerCardsFile); err != nil {
		return err
	}
	timestampedCardsFile, err := applyTimestampToExport(playerCardsFile, timestamp)
	if err != nil {
		return err
	}
	if timestamp != "" {
		printf("   ✓ Timestamped player exports:\n")
		printf("     - %s\n", timestampedSummaryFile)
		printf("     - %s\n", timestampedCardsFile)
	}

	return nil
}

func exportAllAnalysisData(dataDir string, player *clashroyale.Player, timestamp string) error {
	fmt.Println("\n2. Analyzing collection...")
	options := analysis.DefaultAnalysisOptions()
	analysisResult, err := analysis.AnalyzeCardCollection(player, options)
	if err != nil {
		return fmt.Errorf("failed to analyze collection: %w", err)
	}

	pathBuilder := storage.NewPathBuilder(dataDir)
	analysisExportDir := pathBuilder.GetCSVAnalysisDir()
	if err := storage.EnsureDirectory(analysisExportDir); err != nil {
		return fmt.Errorf("failed to create analysis export directory: %w", err)
	}

	analysisExporter := csv.NewAnalysisExporter()
	analysisFile := exportTargetFile(analysisExportDir, analysisExporter.Filename())
	if err := exportWithFeedback(func(ddir string, d any) error { return analysisExporter.Export(ddir, d) }, dataDir, analysisResult, "collection analysis", analysisFile); err != nil {
		return err
	}
	timestampedAnalysisFile, err := applyTimestampToExport(analysisFile, timestamp)
	if err != nil {
		return err
	}
	if timestamp != "" {
		printf("   ✓ Timestamped analysis export: %s\n", timestampedAnalysisFile)
	}

	return nil
}

func exportAllBattleData(dataDir string, battles []clashroyale.Battle, timestamp string) error {
	fmt.Println("\n3. Exporting battle log...")
	pathBuilder := storage.NewPathBuilder(dataDir)
	battleExportDir := filepath.Join(pathBuilder.GetCSVDir(), storage.CSVBattlesSubdir)
	if err := storage.EnsureDirectory(battleExportDir); err != nil {
		return fmt.Errorf("failed to create battle export directory: %w", err)
	}

	battleExporter := csv.NewBattleLogExporter()
	battleLogFile := exportTargetFile(battleExportDir, battleExporter.Filename())
	if err := exportWithFeedback(func(ddir string, d any) error { return battleExporter.Export(ddir, d) }, dataDir, battles, fmt.Sprintf("battles (%d records)", len(battles)), battleLogFile); err != nil {
		return err
	}
	timestampedBattleFile, err := applyTimestampToExport(battleLogFile, timestamp)
	if err != nil {
		return err
	}
	if timestamp != "" {
		printf("   ✓ Timestamped battle export: %s\n", timestampedBattleFile)
	}

	return nil
}

func exportAllEventData(dataDir string, battles []clashroyale.Battle, timestamp string) error {
	fmt.Println("\n4. Extracting event battles...")
	eventBattles := extractEventBattles(battles)

	if len(eventBattles) == 0 {
		fmt.Println("   ℹ No event battles found in recent battles")
		return nil
	}

	pathBuilder := storage.NewPathBuilder(dataDir)
	eventExportDir := pathBuilder.GetCSVEventsDir()
	if err := storage.EnsureDirectory(eventExportDir); err != nil {
		return fmt.Errorf("failed to create event export directory: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "cr-events-export-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary export directory: %w", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
			printf("warning: failed to remove temp directory %s: %v\n", tempDir, cleanupErr)
		}
	}()

	eventExporter := csv.NewBattleLogExporter()
	if err := eventExporter.Export(tempDir, eventBattles); err != nil {
		return fmt.Errorf("failed to export event battles: %w", err)
	}

	battlesFile := filepath.Join(storage.NewPathBuilder(tempDir).GetCSVDir(), storage.CSVBattlesSubdir, "battle_log.csv")
	eventsFile := appendTimestampToFilename(filepath.Join(eventExportDir, "event_battles.csv"), timestamp)
	if err := storage.MoveFile(battlesFile, eventsFile); err != nil {
		return err
	}
	printf("   ✓ Event battles (%d records): %s\n", len(eventBattles), eventsFile)
	return nil
}

func exportAllCardDatabase(dataDir string, cardList *clashroyale.CardList, timestamp string) error {
	fmt.Println("\n5. Exporting card database...")
	pathBuilder := storage.NewPathBuilder(dataDir)
	cardExportDir := pathBuilder.GetCSVReferenceDir()
	if err := storage.EnsureDirectory(cardExportDir); err != nil {
		return fmt.Errorf("failed to create card export directory: %w", err)
	}

	cardExporter := csv.NewCardsExporter()
	cardFile := exportTargetFile(cardExportDir, cardExporter.Filename())
	if err := exportWithFeedback(func(ddir string, d any) error { return cardExporter.Export(ddir, d) }, dataDir, cardList.Items, "card database", cardFile); err != nil {
		return err
	}
	timestampedCardFile, err := applyTimestampToExport(cardFile, timestamp)
	if err != nil {
		return err
	}
	if timestamp != "" {
		printf("   ✓ Timestamped card export: %s\n", timestampedCardFile)
	}

	return nil
}

func printExportAllSummary(dataDir string, player *clashroyale.Player, battles []clashroyale.Battle) {
	printf("\n✅ Export complete for %s!\n", player.Name)
	printf("   Player: %s (%d trophies)\n", player.Name, player.Trophies)
	printf("   Cards: %d collected\n", len(player.Cards))
	printf("   Battles: %d recent games\n", len(battles))
	printf("   Location: %s\n", storage.NewPathBuilder(dataDir).GetCSVDir())
}

func exportAllCommand() *cli.Command {
	return &cli.Command{
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
			timestamp := timestampValue(cmd.Bool("timestamp"))

			apiToken, err := validateAPIToken(cmd)
			if err != nil {
				return err
			}

			client := createClient(apiToken)

			printf("Starting full export for player %s...\n\n", tag)

			exportData, err := loadExportAllData(ctx, client, tag)
			if err != nil {
				return err
			}

			if err := exportAllPlayerData(dataDir, exportData.player, timestamp); err != nil {
				return err
			}

			if err := exportAllAnalysisData(dataDir, exportData.player, timestamp); err != nil {
				return err
			}

			if err := exportAllBattleData(dataDir, exportData.battles, timestamp); err != nil {
				return err
			}

			if err := exportAllEventData(dataDir, exportData.battles, timestamp); err != nil {
				return err
			}

			if err := exportAllCardDatabase(dataDir, exportData.cardList, timestamp); err != nil {
				return err
			}

			printExportAllSummary(dataDir, exportData.player, exportData.battles)
			return nil
		},
	}
}

// addExportCommands adds export-related subcommands to the CLI
func addExportCommands() *cli.Command {
	return &cli.Command{
		Name:  "export",
		Usage: "Export various data types to CSV format",
		Commands: []*cli.Command{
			exportPlayerCommand(),
			exportCardsCommand(),
			exportAnalysisCommand(),
			exportBattlesCommand(),
			exportEventsCommand(),
			exportAllCommand(),
		},
	}
}
