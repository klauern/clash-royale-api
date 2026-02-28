package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/leaderboard"
	"github.com/urfave/cli/v3"
)

const (
	storageFormatSummary = "summary"
	storageFormatJSON    = "json"
)

// addStorageCommands returns the storage command group.
//
//nolint:funlen // Command list intentionally explicit for discoverability.
func addStorageCommands() *cli.Command {
	return &cli.Command{
		Name:  "storage",
		Usage: "Manage persistent deck storage",
		Commands: []*cli.Command{
			{
				Name:  "stats",
				Usage: "Show storage database size and deck counts by archetype",
				Flags: []cli.Flag{
					playerTagFlag(true),
					&cli.StringFlag{
						Name:  "format",
						Value: storageFormatSummary,
						Usage: "Output format: summary, json",
					},
				},
				Action: storageStatsCommand,
			},
			{
				Name:  "purge",
				Usage: "Delete all saved decks for a player",
				Flags: []cli.Flag{
					playerTagFlag(true),
					&cli.BoolFlag{Name: "confirm", Aliases: []string{"y"}, Usage: "Skip confirmation prompt"},
				},
				Action: storagePurgeCommand,
			},
			{
				Name:  "cleanup",
				Usage: "Remove low-score and/or old decks",
				Flags: []cli.Flag{
					playerTagFlag(true),
					&cli.Float64Flag{Name: "min-score", Usage: "Delete decks with overall score below this threshold"},
					&cli.IntFlag{Name: "older-than-days", Usage: "Delete decks evaluated more than N days ago"},
					&cli.StringFlag{Name: "archetype", Usage: "Restrict cleanup to a single archetype"},
					&cli.BoolFlag{Name: "dry-run", Usage: "Show matching deck count without deleting"},
					&cli.BoolFlag{Name: "confirm", Aliases: []string{"y"}, Usage: "Skip confirmation prompt"},
				},
				Action: storageCleanupCommand,
			},
			{
				Name:  "prune",
				Usage: "Keep only top N decks per archetype",
				Flags: []cli.Flag{
					playerTagFlag(true),
					&cli.IntFlag{Name: "keep", Value: 50, Usage: "Number of top decks to keep per archetype"},
					&cli.BoolFlag{Name: "dry-run", Usage: "Show decks that would be deleted without deleting"},
					&cli.BoolFlag{Name: "confirm", Aliases: []string{"y"}, Usage: "Skip confirmation prompt"},
				},
				Action: storagePruneCommand,
			},
			{
				Name:  "vacuum",
				Usage: "Compact the SQLite database file",
				Flags: []cli.Flag{
					playerTagFlag(true),
				},
				Action: storageVacuumCommand,
			},
			{
				Name:  "export",
				Usage: "Export storage decks for backup",
				Flags: []cli.Flag{
					playerTagFlag(true),
					&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "Output path for backup JSON", Required: true},
					&cli.StringFlag{Name: "format", Value: storageFormatJSON, Usage: "Export format (json)"},
				},
				Action: storageExportCommand,
			},
			{
				Name:  "import",
				Usage: "Import decks from backup",
				Flags: []cli.Flag{
					playerTagFlag(true),
					&cli.StringFlag{Name: "input", Aliases: []string{"i"}, Usage: "Backup file to import", Required: true},
					&cli.BoolFlag{Name: "confirm", Aliases: []string{"y"}, Usage: "Skip confirmation prompt"},
				},
				Action: storageImportCommand,
			},
		},
	}
}

type storageStatsOutput struct {
	PlayerTag       string                       `json:"player_tag"`
	DBPath          string                       `json:"db_path"`
	DBSizeBytes     int64                        `json:"db_size_bytes"`
	DBSizeHuman     string                       `json:"db_size_human"`
	TotalDecks      int                          `json:"total_decks"`
	ArchetypeCounts []leaderboard.ArchetypeCount `json:"archetype_counts"`
}

func storageStatsCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")
	format := strings.ToLower(cmd.String("format"))

	storage, err := leaderboard.NewStorage(playerTag)
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	totalDecks, err := storage.Count()
	if err != nil {
		return fmt.Errorf("failed to count decks: %w", err)
	}

	archetypeCounts, err := storage.GetArchetypeCounts()
	if err != nil {
		return fmt.Errorf("failed to get archetype counts: %w", err)
	}

	dbPath := storage.GetDBPath()
	dbInfo, err := os.Stat(dbPath)
	if err != nil {
		return fmt.Errorf("failed to read database size: %w", err)
	}

	output := storageStatsOutput{
		PlayerTag:       strings.TrimPrefix(playerTag, "#"),
		DBPath:          dbPath,
		DBSizeBytes:     dbInfo.Size(),
		DBSizeHuman:     humanReadableBytes(dbInfo.Size()),
		TotalDecks:      totalDecks,
		ArchetypeCounts: archetypeCounts,
	}

	switch format {
	case storageFormatSummary:
		return printStorageStatsSummary(output)
	case storageFormatJSON:
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		printf("%s\n", string(data))
		return nil
	default:
		return fmt.Errorf("invalid format %q (valid: summary, json)", format)
	}
}

func storagePurgeCommand(ctx context.Context, cmd *cli.Command) error {
	storage, err := leaderboard.NewStorage(cmd.String("tag"))
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	count, err := storage.Count()
	if err != nil {
		return fmt.Errorf("failed to count decks: %w", err)
	}
	if count == 0 {
		printf("Storage is already empty\n")
		return nil
	}

	if !cmd.Bool("confirm") {
		message := fmt.Sprintf("This will permanently delete %d deck(s). Continue? (y/N): ", count)
		confirmed, err := confirmStorageAction(message)
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		if !confirmed {
			printf("Operation canceled\n")
			return nil
		}
	}

	if err := storage.Clear(); err != nil {
		return fmt.Errorf("failed to purge decks: %w", err)
	}
	if _, err := storage.RecalculateStats(); err != nil {
		return fmt.Errorf("failed to recalculate stats: %w", err)
	}
	printf("Purged %d deck(s)\n", count)
	return nil
}

//nolint:gocyclo // Multiple validation and action branches are intentional.
func storageCleanupCommand(ctx context.Context, cmd *cli.Command) error {
	opts, err := buildCleanupOptions(cmd)
	if err != nil {
		return err
	}

	storage, err := leaderboard.NewStorage(cmd.String("tag"))
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	matchingCount, err := countMatchingCleanupDecks(storage, opts)
	if err != nil {
		return err
	}
	if matchingCount == 0 {
		printf("No decks matched cleanup filters\n")
		return nil
	}
	if cmd.Bool("dry-run") {
		printf("Would delete %d deck(s)\n", matchingCount)
		return nil
	}

	if !cmd.Bool("confirm") {
		message := fmt.Sprintf("Delete %d matching deck(s)? (y/N): ", matchingCount)
		confirmed, err := confirmStorageAction(message)
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		if !confirmed {
			printf("Operation canceled\n")
			return nil
		}
	}

	deleted, err := storage.Cleanup(opts)
	if err != nil {
		return err
	}
	if _, err := storage.RecalculateStats(); err != nil {
		return fmt.Errorf("failed to recalculate stats: %w", err)
	}

	printf("Deleted %d deck(s)\n", deleted)
	return nil
}

//nolint:gocyclo // Multiple validation and action branches are intentional.
func storagePruneCommand(ctx context.Context, cmd *cli.Command) error {
	keep := cmd.Int("keep")
	if keep < 1 {
		return fmt.Errorf("--keep must be >= 1")
	}

	storage, err := leaderboard.NewStorage(cmd.String("tag"))
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	before, err := storage.Count()
	if err != nil {
		return fmt.Errorf("failed to count decks: %w", err)
	}
	if before == 0 {
		printf("Storage is empty\n")
		return nil
	}

	toDelete, err := calculatePruneDeleteCount(storage, keep)
	if err != nil {
		return err
	}

	if toDelete == 0 {
		printf("No decks to prune (all archetypes already <= %d)\n", keep)
		return nil
	}
	if cmd.Bool("dry-run") {
		printf("Would delete %d deck(s), keeping top %d per archetype\n", toDelete, keep)
		return nil
	}

	if !cmd.Bool("confirm") {
		message := fmt.Sprintf("Delete %d deck(s) and keep top %d per archetype? (y/N): ", toDelete, keep)
		confirmed, err := confirmStorageAction(message)
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		if !confirmed {
			printf("Operation canceled\n")
			return nil
		}
	}

	deleted, err := storage.PruneTopNPerArchetype(keep)
	if err != nil {
		return err
	}
	if _, err := storage.RecalculateStats(); err != nil {
		return fmt.Errorf("failed to recalculate stats: %w", err)
	}

	printf("Pruned %d deck(s), kept top %d per archetype\n", deleted, keep)
	return nil
}

func countMatchingCleanupDecks(storage *leaderboard.Storage, opts leaderboard.CleanupOptions) (int, error) {
	queryOpts := leaderboard.QueryOptions{}
	if opts.Archetype != "" {
		queryOpts.Archetype = opts.Archetype
	}

	allDecks, err := storage.Query(queryOpts)
	if err != nil {
		return 0, fmt.Errorf("failed to query decks for cleanup: %w", err)
	}

	matchingCount := 0
	for _, deck := range allDecks {
		if opts.MinScore > 0 && deck.OverallScore >= opts.MinScore {
			continue
		}
		if !opts.OlderThan.IsZero() && !deck.EvaluatedAt.Before(opts.OlderThan) {
			continue
		}
		matchingCount++
	}
	return matchingCount, nil
}

func calculatePruneDeleteCount(storage *leaderboard.Storage, keep int) (int, error) {
	counts, err := storage.GetArchetypeCounts()
	if err != nil {
		return 0, fmt.Errorf("failed to get archetype counts: %w", err)
	}

	toDelete := 0
	for _, c := range counts {
		if c.Count > keep {
			toDelete += c.Count - keep
		}
	}
	return toDelete, nil
}

func storageVacuumCommand(ctx context.Context, cmd *cli.Command) error {
	storage, err := leaderboard.NewStorage(cmd.String("tag"))
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	beforeSize := int64(0)
	if info, err := os.Stat(storage.GetDBPath()); err == nil {
		beforeSize = info.Size()
	}

	if err := storage.Vacuum(); err != nil {
		return err
	}

	afterSize := int64(0)
	if info, err := os.Stat(storage.GetDBPath()); err == nil {
		afterSize = info.Size()
	}

	freed := max(beforeSize-afterSize, 0)
	printf("Vacuum complete. Size: %s -> %s (freed %s)\n", humanReadableBytes(beforeSize), humanReadableBytes(afterSize), humanReadableBytes(freed))
	return nil
}

func storageExportCommand(ctx context.Context, cmd *cli.Command) error {
	format := strings.ToLower(cmd.String("format"))
	if format != storageFormatJSON {
		return fmt.Errorf("invalid format %q (only json is supported)", format)
	}

	storage, err := leaderboard.NewStorage(cmd.String("tag"))
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	count, err := storage.ExportJSON(cmd.String("output"))
	if err != nil {
		return err
	}

	printf("Exported %d deck(s) to %s\n", count, cmd.String("output"))
	return nil
}

func storageImportCommand(ctx context.Context, cmd *cli.Command) error {
	input := cmd.String("input")
	if _, err := os.Stat(input); err != nil {
		return fmt.Errorf("invalid input file: %w", err)
	}

	storage, err := leaderboard.NewStorage(cmd.String("tag"))
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	if !cmd.Bool("confirm") {
		confirmed, err := confirmStorageAction(fmt.Sprintf("Import decks from %s? (y/N): ", input))
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		if !confirmed {
			printf("Operation canceled\n")
			return nil
		}
	}

	inserted, updated, err := storage.ImportJSON(input)
	if err != nil {
		return err
	}
	if _, err := storage.RecalculateStats(); err != nil {
		return fmt.Errorf("failed to recalculate stats: %w", err)
	}

	printf("Imported decks from %s (new: %d, updated: %d)\n", input, inserted, updated)
	return nil
}

func buildCleanupOptions(cmd *cli.Command) (leaderboard.CleanupOptions, error) {
	minScore := cmd.Float64("min-score")
	if minScore < 0 || minScore > 10 {
		return leaderboard.CleanupOptions{}, fmt.Errorf("--min-score must be between 0 and 10")
	}

	olderThanDays := cmd.Int("older-than-days")
	if olderThanDays < 0 {
		return leaderboard.CleanupOptions{}, fmt.Errorf("--older-than-days must be >= 0")
	}

	opts := leaderboard.CleanupOptions{
		MinScore:  minScore,
		Archetype: strings.TrimSpace(cmd.String("archetype")),
	}
	if olderThanDays > 0 {
		opts.OlderThan = time.Now().AddDate(0, 0, -olderThanDays)
	}
	if opts.MinScore == 0 && opts.OlderThan.IsZero() && opts.Archetype == "" {
		return leaderboard.CleanupOptions{}, fmt.Errorf("at least one filter required (--min-score, --older-than-days, or --archetype)")
	}

	return opts, nil
}

func confirmStorageAction(prompt string) (bool, error) {
	fprintf(os.Stdout, "%s", prompt)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(response), "y"), nil
}

func printStorageStatsSummary(stats storageStatsOutput) error {
	printf("\nStorage Statistics for #%s\n", stats.PlayerTag)
	printf("═══════════════════════════════════════════\n")
	printf("Database: %s\n", stats.DBPath)
	printf("Database Size: %s (%d bytes)\n", stats.DBSizeHuman, stats.DBSizeBytes)
	printf("Total Decks: %d\n", stats.TotalDecks)

	if len(stats.ArchetypeCounts) == 0 {
		printf("\nDeck Counts by Archetype: no decks found\n")
		return nil
	}

	sort.SliceStable(stats.ArchetypeCounts, func(i, j int) bool {
		if stats.ArchetypeCounts[i].Count == stats.ArchetypeCounts[j].Count {
			return stats.ArchetypeCounts[i].Archetype < stats.ArchetypeCounts[j].Archetype
		}
		return stats.ArchetypeCounts[i].Count > stats.ArchetypeCounts[j].Count
	})

	printf("\nDeck Counts by Archetype:\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "Archetype\tCount\tPercentage\n")
	fprintf(w, "---------\t-----\t----------\n")
	for _, c := range stats.ArchetypeCounts {
		pct := 0.0
		if stats.TotalDecks > 0 {
			pct = float64(c.Count) * 100 / float64(stats.TotalDecks)
		}
		fprintf(w, "%s\t%d\t%.1f%%\n", c.Archetype, c.Count, pct)
	}
	flushWriter(w)

	return nil
}

func humanReadableBytes(sizeBytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(sizeBytes)
	unitIndex := 0
	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}
	if unitIndex == 0 {
		return fmt.Sprintf("%d %s", sizeBytes, units[unitIndex])
	}
	return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}
