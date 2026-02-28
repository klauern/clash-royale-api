package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/klauer/clash-royale-api/go/internal/playertag"
	"github.com/klauer/clash-royale-api/go/pkg/leaderboard"
	"github.com/urfave/cli/v3"
)

// addLeaderboardCommands returns the leaderboard command group
func addLeaderboardCommands() *cli.Command {
	return &cli.Command{
		Name:  "leaderboard",
		Usage: "View and manage the deck leaderboard",
		Commands: []*cli.Command{
			{
				Name:  "show",
				Usage: "Display top decks from the leaderboard",
				Flags: []cli.Flag{
					playerTagFlag(true),
					&cli.IntFlag{
						Name:    "top",
						Aliases: []string{"n"},
						Value:   10,
						Usage:   "Number of top decks to display",
					},
					&cli.StringFlag{
						Name:  "format",
						Value: "summary",
						Usage: "Output format: summary, detailed, json, csv",
					},
					&cli.StringFlag{
						Name:  "output",
						Usage: "Output file path (optional, prints to stdout if not specified)",
					},
				},
				Action: leaderboardShowCommand,
			},
			{
				Name:  "filter",
				Usage: "Query leaderboard with filters",
				Flags: []cli.Flag{
					playerTagFlag(true),
					&cli.StringFlag{
						Name:  "archetype",
						Usage: "Filter by archetype (e.g., beatdown, control, cycle)",
					},
					&cli.Float64Flag{
						Name:  "min-score",
						Usage: "Minimum overall score (0-10)",
					},
					&cli.Float64Flag{
						Name:  "max-score",
						Usage: "Maximum overall score (0-10)",
					},
					&cli.Float64Flag{
						Name:  "min-elixir",
						Usage: "Minimum average elixir",
					},
					&cli.Float64Flag{
						Name:  "max-elixir",
						Usage: "Maximum average elixir",
					},
					&cli.StringFlag{
						Name:  "strategy",
						Usage: "Filter by generation strategy",
					},
					&cli.StringSliceFlag{
						Name:  "require-all",
						Usage: "Decks must contain ALL of these cards (comma-separated)",
					},
					&cli.StringSliceFlag{
						Name:  "require-any",
						Usage: "Decks must contain ANY of these cards (comma-separated)",
					},
					&cli.StringSliceFlag{
						Name:  "exclude",
						Usage: "Exclude decks containing ANY of these cards (comma-separated)",
					},
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"n"},
						Value:   10,
						Usage:   "Maximum number of results",
					},
					&cli.IntFlag{
						Name:  "offset",
						Value: 0,
						Usage: "Number of results to skip (for pagination)",
					},
					&cli.StringFlag{
						Name:  "sort-by",
						Value: "overall_score",
						Usage: "Sort field: overall_score, attack_score, defense_score, synergy_score, avg_elixir",
					},
					&cli.StringFlag{
						Name:  "order",
						Value: "desc",
						Usage: "Sort order: asc, desc",
					},
					&cli.StringFlag{
						Name:  "format",
						Value: "summary",
						Usage: "Output format: summary, detailed, json, csv",
					},
					&cli.StringFlag{
						Name:  "output",
						Usage: "Output file path (optional, prints to stdout if not specified)",
					},
				},
				Action: leaderboardFilterCommand,
			},
			{
				Name:  "export",
				Usage: "Export leaderboard to file",
				Flags: []cli.Flag{
					playerTagFlag(true),
					&cli.StringFlag{
						Name:  "format",
						Value: "json",
						Usage: "Export format: json, csv",
					},
					&cli.StringFlag{
						Name:  "output",
						Usage: "Output file path (required)",
					},
				},
				Action: leaderboardExportCommand,
			},
			{
				Name:  "stats",
				Usage: "Display leaderboard statistics",
				Flags: []cli.Flag{
					playerTagFlag(true),
					&cli.BoolFlag{
						Name:  "archetypes",
						Usage: "Show archetype distribution",
					},
				},
				Action: leaderboardStatsCommand,
			},
			{
				Name:  "clear",
				Usage: "Clear all decks from the leaderboard",
				Flags: []cli.Flag{
					playerTagFlag(true),
					&cli.BoolFlag{
						Name:    "confirm",
						Aliases: []string{"y"},
						Usage:   "Skip confirmation prompt",
					},
				},
				Action: leaderboardClearCommand,
			},
		},
	}
}

func leaderboardShowCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")
	top := cmd.Int("top")
	format := cmd.String("format")
	outputPath := cmd.String("output")

	storage, err := leaderboard.NewStorage(playerTag)
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	decks, err := storage.GetTopN(top)
	if err != nil {
		return fmt.Errorf("failed to query decks: %w", err)
	}

	if len(decks) == 0 {
		printf("No decks found in leaderboard for player #%s\n", playerTag)
		printf("Run 'cr-api deck discover run --tag %s' to populate the leaderboard\n", playerTag)
		return nil
	}

	// Prepare output
	var output string
	switch format {
	case "json":
		data, err := json.MarshalIndent(decks, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		output = string(data)
	case "csv":
		output, err = formatDecksAsCSV(decks)
		if err != nil {
			return err
		}
	case "detailed":
		output = formatDecksDetailed(decks)
	default: // summary
		output = formatDecksSummary(decks)
	}

	// Write output
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(output), 0o644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		printf("Output written to: %s\n", outputPath)
	} else {
		printf("%s\n", output)
	}

	return nil
}

func leaderboardFilterCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")

	storage, err := leaderboard.NewStorage(playerTag)
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	// Build query options
	opts := leaderboard.QueryOptions{
		Limit:           cmd.Int("limit"),
		Offset:          cmd.Int("offset"),
		Archetype:       cmd.String("archetype"),
		Strategy:        cmd.String("strategy"),
		MinScore:        cmd.Float64("min-score"),
		MaxScore:        cmd.Float64("max-score"),
		MinAvgElixir:    cmd.Float64("min-elixir"),
		MaxAvgElixir:    cmd.Float64("max-elixir"),
		SortBy:          cmd.String("sort-by"),
		SortOrder:       cmd.String("order"),
		RequireAllCards: cmd.StringSlice("require-all"),
		RequireAnyCards: cmd.StringSlice("require-any"),
		ExcludeCards:    cmd.StringSlice("exclude"),
	}

	decks, err := storage.Query(opts)
	if err != nil {
		return fmt.Errorf("failed to query decks: %w", err)
	}

	if len(decks) == 0 {
		printf("No decks found matching the filters\n")
		return nil
	}

	format := cmd.String("format")
	outputPath := cmd.String("output")

	var output string
	switch format {
	case "json":
		data, err := json.MarshalIndent(decks, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		output = string(data)
	case "csv":
		output, err = formatDecksAsCSV(decks)
		if err != nil {
			return err
		}
	case "detailed":
		output = formatDecksDetailed(decks)
	default: // summary
		output = formatDecksSummary(decks)
	}

	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(output), 0o644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		printf("Output written to: %s\n", outputPath)
	} else {
		printf("%s\n", output)
	}

	return nil
}

func leaderboardExportCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")
	format := cmd.String("format")
	outputPath := cmd.String("output")

	if outputPath == "" {
		return fmt.Errorf("--output is required for export")
	}

	storage, err := leaderboard.NewStorage(playerTag)
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	// Get all decks
	opts := leaderboard.QueryOptions{
		Limit:     0, // No limit
		SortBy:    "overall_score",
		SortOrder: "desc",
	}
	decks, err := storage.Query(opts)
	if err != nil {
		return fmt.Errorf("failed to query decks: %w", err)
	}

	if len(decks) == 0 {
		return fmt.Errorf("no decks to export")
	}

	// Determine output format and write
	switch format {
	case "csv":
		csvData, err := formatDecksAsCSV(decks)
		if err != nil {
			return err
		}
		if err := os.WriteFile(outputPath, []byte(csvData), 0o644); err != nil {
			return fmt.Errorf("failed to write CSV: %w", err)
		}
	default: // json
		data, err := json.MarshalIndent(decks, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		if err := os.WriteFile(outputPath, data, 0o644); err != nil {
			return fmt.Errorf("failed to write JSON: %w", err)
		}
	}

	printf("Exported %d decks to %s\n", len(decks), outputPath)
	return nil
}

func leaderboardStatsCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")
	showArchetypes := cmd.Bool("archetypes")

	storage, err := leaderboard.NewStorage(playerTag)
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	stats, err := storage.GetStats()
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	// Display header
	printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	printf("║                      LEADERBOARD STATS                             ║\n")
	printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	printf("Player: #%s\n", playerTag)
	printf("Total Decks Stored: %d\n", stats.TotalUniqueDecks)
	if stats.TotalDecksEvaluated > 0 {
		printf("Total Evaluated: %d\n", stats.TotalDecksEvaluated)
	}
	printf("Last Updated: %s\n\n", stats.LastUpdated.Format("2006-01-02 15:04:05"))

	if stats.TotalUniqueDecks > 0 {
		printf("Score Statistics:\n")
		printf("  Top Score: %.2f\n", stats.TopScore)
		printf("  Avg Score: %.2f\n\n", stats.AvgScore)
	}

	// Get database file size
	homeDir, err := os.UserHomeDir()
	if err == nil {
		sanitizedTag, sanitizeErr := playertag.Sanitize(playerTag)
		if sanitizeErr == nil {
			dbPath := filepath.Join(homeDir, ".cr-api", "leaderboards", fmt.Sprintf("%s.db", sanitizedTag))
			if info, err := os.Stat(dbPath); err == nil {
				printf("Database Size: %.2f MB\n", float64(info.Size())/(1024*1024))
			}
		}
	}

	// Show archetype distribution if requested
	if showArchetypes && stats.TotalUniqueDecks > 0 {
		printf("\nArchetype Distribution:\n")
		printf("════════════════════════\n")

		// Get all decks to calculate distribution
		allDecks, err := storage.Query(leaderboard.QueryOptions{Limit: 0})
		if err == nil {
			archetypeCounts := make(map[string]int)
			for _, deck := range allDecks {
				archetypeCounts[deck.Archetype]++
			}

			// Sort by count
			type archCount struct {
				archetype string
				count     int
				pct       float64
			}
			sorted := make([]archCount, 0, len(archetypeCounts))
			for arch, count := range archetypeCounts {
				sorted = append(sorted, archCount{
					archetype: arch,
					count:     count,
					pct:       float64(count) / float64(len(allDecks)) * 100,
				})
			}
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].count > sorted[j].count
			})

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fprintf(w, "Archetype\tCount\tPercentage\n")
			fprintf(w, "---------\t-----\t----------\n")
			for _, ac := range sorted {
				fprintf(w, "%s\t%d\t%.1f%%\n", ac.archetype, ac.count, ac.pct)
			}
			flushWriter(w)
		}
	}

	return nil
}

func leaderboardClearCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")
	confirm := cmd.Bool("confirm")

	storage, err := leaderboard.NewStorage(playerTag)
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	// Get current count
	count, err := storage.Count()
	if err != nil {
		return fmt.Errorf("failed to get deck count: %w", err)
	}

	if count == 0 {
		printf("Leaderboard is already empty for player #%s\n", playerTag)
		return nil
	}

	// Confirm unless --confirm flag
	if !confirm {
		printf("This will delete all %d decks from the leaderboard for player #%s\n", count, playerTag)
		printf("Are you sure? (y/N): ")
		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			printf("Operation canceled\n")
			return nil
		}
	}

	// Clear all decks
	if err := storage.Clear(); err != nil {
		return fmt.Errorf("failed to clear leaderboard: %w", err)
	}

	printf("Cleared %d decks from leaderboard for player #%s\n", count, playerTag)
	return nil
}

// formatDecksSummary formats decks in a compact summary view
func formatDecksSummary(decks []leaderboard.DeckEntry) string {
	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)

	fprintf(w, "Rank\tScore\tArchetype\tElixir\tDeck\n")
	fprintf(w, "----\t-----\t---------\t------\t----\n")

	for i, deck := range decks {
		cardsStr := strings.Join(deck.Cards, ", ")
		if len(cardsStr) > 50 {
			cardsStr = cardsStr[:50] + "..."
		}
		fprintf(w, "%d\t%.2f\t%s\t%.1f\t%s\n",
			i+1, deck.OverallScore, deck.Archetype, deck.AvgElixir, cardsStr)
	}

	flushWriter(w)
	return sb.String()
}

// formatDecksDetailed formats decks with full details
func formatDecksDetailed(decks []leaderboard.DeckEntry) string {
	var sb strings.Builder

	for i, deck := range decks {
		sb.WriteString("════════════════════════════════════════════════════════════════════════════════\n")
		fprintf(&sb, "Deck #%d - Score: %.2f\n", i+1, deck.OverallScore)
		sb.WriteString("════════════════════════════════════════════════════════════════════════════════\n")
		fprintf(&sb, "Archetype: %s (%.1f%% confidence)\n", deck.Archetype, deck.ArchetypeConf*100)
		fprintf(&sb, "Average Elixir: %.1f\n", deck.AvgElixir)
		fprintf(&sb, "Strategy: %s\n", deck.Strategy)
		sb.WriteString("\n")
		sb.WriteString("Category Scores:\n")
		fprintf(&sb, "  Attack:       %.2f\n", deck.AttackScore)
		fprintf(&sb, "  Defense:      %.2f\n", deck.DefenseScore)
		fprintf(&sb, "  Synergy:      %.2f\n", deck.SynergyScore)
		fprintf(&sb, "  Versatility:  %.2f\n", deck.VersatilityScore)
		fprintf(&sb, "  F2P:          %.2f\n", deck.F2PScore)
		fprintf(&sb, "  Playability:  %.2f\n", deck.PlayabilityScore)
		sb.WriteString("\n")
		sb.WriteString("Cards:\n")
		for _, card := range deck.Cards {
			fprintf(&sb, "  • %s\n", card)
		}
		sb.WriteString("\n")
		fprintf(&sb, "Evaluated: %s\n", deck.EvaluatedAt.Format("2006-01-02 15:04:05"))
		if deck.PlayerTag != "" {
			fprintf(&sb, "Player: #%s\n", deck.PlayerTag)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatDecksAsCSV formats decks as CSV
func formatDecksAsCSV(decks []leaderboard.DeckEntry) (string, error) {
	var sb strings.Builder
	header := []string{
		"Rank", "OverallScore", "AttackScore", "DefenseScore", "SynergyScore",
		"VersatilityScore", "F2PScore", "PlayabilityScore",
		"Archetype", "ArchetypeConf", "AvgElixir", "Strategy", "Cards",
	}
	rows := make([][]string, 0, len(decks))
	for i, deck := range decks {
		cardsStr := strings.Join(deck.Cards, "; ")
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%.2f", deck.OverallScore),
			fmt.Sprintf("%.2f", deck.AttackScore),
			fmt.Sprintf("%.2f", deck.DefenseScore),
			fmt.Sprintf("%.2f", deck.SynergyScore),
			fmt.Sprintf("%.2f", deck.VersatilityScore),
			fmt.Sprintf("%.2f", deck.F2PScore),
			fmt.Sprintf("%.2f", deck.PlayabilityScore),
			deck.Archetype,
			fmt.Sprintf("%.2f", deck.ArchetypeConf),
			fmt.Sprintf("%.1f", deck.AvgElixir),
			deck.Strategy,
			cardsStr,
		})
	}
	if err := writeCSVDocument(&sb, header, rows); err != nil {
		return "", err
	}

	return sb.String(), nil
}
