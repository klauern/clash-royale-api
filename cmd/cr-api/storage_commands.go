package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/klauer/clash-royale-api/go/pkg/leaderboard"
	"github.com/urfave/cli/v3"
)

// addStorageCommands returns the storage command group.
func addStorageCommands() *cli.Command {
	return &cli.Command{
		Name:  "storage",
		Usage: "Manage persistent deck storage",
		Commands: []*cli.Command{
			{
				Name:  "stats",
				Usage: "Show storage database size and deck counts by archetype",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "format",
						Value: batchFormatSummary,
						Usage: "Output format: summary, json",
					},
				},
				Action: storageStatsCommand,
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
	case batchFormatSummary:
		return printStorageStatsSummary(output)
	case "json":
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
		pct := float64(c.Count) * 100 / float64(stats.TotalDecks)
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
