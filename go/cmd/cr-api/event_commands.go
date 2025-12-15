package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/exporter/csv"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/events"
	"github.com/urfave/cli/v3"
)

// addEventCommands adds event-related subcommands to the CLI
func addEventCommands() *cli.Command {
	return &cli.Command{
		Name:  "events",
		Usage: "Event deck tracking and analysis commands",
		Commands: []*cli.Command{
			{
				Name:  "scan",
				Usage: "Scan player's battle log for event decks",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "days",
						Value: 7,
						Usage: "Number of days to scan back",
					},
					&cli.StringSliceFlag{
						Name:  "event-types",
						Usage: "Specific event types to scan: challenge, tournament, special_event",
					},
					&cli.BoolFlag{
						Name:  "save",
						Usage: "Save event decks to file",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export event data to CSV",
					},
				},
				Action: eventScanCommand,
			},
			{
				Name:  "list",
				Usage: "List tracked event decks for a player",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "event-type",
						Usage: "Filter by event type",
					},
					&cli.IntFlag{
						Name:  "days",
						Usage: "Filter to recent events (in days)",
					},
					&cli.IntFlag{
						Name:  "min-battles",
						Value: 1,
						Usage: "Minimum battle count",
					},
					&cli.StringFlag{
						Name:  "sort-by",
						Value: "date",
						Usage: "Sort by: date, wins, win_rate",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export event list to CSV",
					},
				},
				Action: eventListCommand,
			},
			{
				Name:  "analyze",
				Usage: "Analyze event deck performance",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "event-id",
						Usage: "Specific event ID to analyze",
					},
					&cli.StringFlag{
						Name:  "event-type",
						Usage: "Analyze all events of this type",
					},
					&cli.IntFlag{
						Name:  "min-battles",
						Value: 5,
						Usage: "Minimum battles for meaningful analysis",
					},
					&cli.BoolFlag{
						Name:  "include-decks",
						Usage: "Include individual deck analysis",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export analysis to CSV",
					},
				},
				Action: eventAnalyzeCommand,
			},
			{
				Name:  "compare",
				Usage: "Compare performance between different event types or decks",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:  "event-types",
						Usage: "Event types to compare (e.g., challenge,grand_challenge)",
					},
					&cli.StringSliceFlag{
						Name:  "event-ids",
						Usage: "Specific event IDs to compare",
					},
					&cli.StringFlag{
						Name:  "metric",
						Value: "win_rate",
						Usage: "Comparison metric: win_rate, avg_crowns, win_streak",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export comparison to CSV",
					},
				},
				Action: eventCompareCommand,
			},
			{
				Name:  "deck-stats",
				Usage: "Show statistics for cards used in events",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "event-type",
						Usage: "Filter by event type",
					},
					&cli.IntFlag{
						Name:  "days",
						Value: 30,
						Usage: "Analyze events from last N days",
					},
					&cli.IntFlag{
						Name:  "min-usage",
						Value: 3,
						Usage: "Minimum times card must be used",
					},
					&cli.BoolFlag{
						Name:  "show-archetypes",
						Usage: "Show deck archetype analysis",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export card stats to CSV",
					},
				},
				Action: eventDeckStatsCommand,
			},
		},
	}
}

func eventScanCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	days := cmd.Int("days")
	eventTypes := cmd.StringSlice("event-types")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	exportCSV := cmd.Bool("export-csv")
	dataDir := cmd.String("data-dir")

	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	if verbose {
		fmt.Printf("Scanning battle log for player %s (last %d days)\n", tag, days)
		if len(eventTypes) > 0 {
			fmt.Printf("Event types to scan: %v\n", eventTypes)
		}
	}

	// Create API client
	client := clashroyale.NewClient(apiToken)

	// Get battle logs
	battleLog, err := client.GetPlayerBattleLog(tag)
	if err != nil {
		return fmt.Errorf("failed to get battle logs: %w", err)
	}
	battles := []clashroyale.Battle(*battleLog)

	// Filter battles by days
	if days > 0 {
		cutoff := time.Now().AddDate(0, 0, -days)
		filtered := make([]clashroyale.Battle, 0, len(battles))
		for _, battle := range battles {
			if !battle.UTCDate.Before(cutoff) {
				filtered = append(filtered, battle)
			}
		}
		battles = filtered
	}

	if verbose {
		fmt.Printf("Found %d recent battles to scan\n", len(battles))
	}

	// Create manager and parse event decks
	manager := events.NewManager(dataDir)
	importedDecks, err := manager.ImportFromBattleLogs(battles, tag)
	if err != nil {
		return fmt.Errorf("failed to import event decks: %w", err)
	}

	// Filter by event types if specified
	filteredDecks := importedDecks
	if len(eventTypes) > 0 {
		allowedTypes := make(map[events.EventType]bool)
		for _, et := range eventTypes {
			allowedTypes[events.EventType(et)] = true
		}
		filtered := make([]events.EventDeck, 0, len(importedDecks))
		for _, deck := range importedDecks {
			if allowedTypes[deck.EventType] {
				filtered = append(filtered, deck)
			}
		}
		filteredDecks = filtered
	}

	if verbose {
		fmt.Printf("Successfully imported %d event decks (after filtering: %d)\n", len(importedDecks), len(filteredDecks))
		if len(filteredDecks) > 0 {
			collection := &events.EventDeckCollection{
				PlayerTag: tag,
				Decks:     filteredDecks,
			}
			displayEventSummary(collection)
		} else {
			fmt.Println("No event decks matched the criteria")
		}
	}

	// Export to CSV if requested
	if exportCSV && len(filteredDecks) > 0 {
		collection := &events.EventDeckCollection{
			PlayerTag: tag,
			Decks:     filteredDecks,
		}
		exporter := csv.NewEventDeckExporter()
		if err := exporter.Export(dataDir, collection); err != nil {
			return fmt.Errorf("failed to export event decks to CSV: %w", err)
		}
		if verbose {
			fmt.Println("Event decks exported to CSV")
		}
	}

	return nil
}

func eventListCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	eventType := cmd.String("event-type")
	days := cmd.Int("days")
	minBattles := cmd.Int("min-battles")
	sortBy := cmd.String("sort-by")
	exportCSV := cmd.Bool("export-csv")
	apiToken := cmd.String("api-token")

	if apiToken == "" {
		return fmt.Errorf("API token is required")
	}

	// Load existing event deck collection
	dataDir := cmd.String("data-dir")
	collection, err := loadEventDeckCollection(dataDir, tag)
	if err != nil {
		return fmt.Errorf("failed to load event decks: %w", err)
	}

	// Filter events
	filtered := filterEventDecks(collection, eventType, days, minBattles)

	// Sort events
	sortEventDecks(filtered, sortBy)

	// Display events
	displayEventList(filtered)

	// Export to CSV if requested
	if exportCSV {
		// Create a filtered collection for export
		exportCollection := &events.EventDeckCollection{
			PlayerTag:   tag,
			Decks:       filtered,
			LastUpdated: collection.LastUpdated,
		}

		eventExporter := csv.NewEventDeckExporter()
		if err := eventExporter.Export(dataDir, exportCollection); err != nil {
			return fmt.Errorf("failed to export event list: %w", err)
		}
		fmt.Printf("Event list exported to CSV\n")
	}

	return nil
}

func eventAnalyzeCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	eventID := cmd.String("event-id")
	eventType := cmd.String("event-type")
	apiToken := cmd.String("api-token")

	if apiToken == "" {
		return fmt.Errorf("API token is required")
	}

	if eventID == "" && eventType == "" {
		return fmt.Errorf("either --event-id or --event-type must be specified")
	}

	// Load event deck collection
	dataDir := cmd.String("data-dir")
	collection, err := loadEventDeckCollection(dataDir, tag)
	if err != nil {
		return fmt.Errorf("failed to load event decks: %w", err)
	}

	// For now, just display basic statistics
	displayEventStats(collection)

	return nil
}

func eventCompareCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	apiToken := cmd.String("api-token")

	if apiToken == "" {
		return fmt.Errorf("API token is required")
	}

	// Load event deck collection
	dataDir := cmd.String("data-dir")
	collection, err := loadEventDeckCollection(dataDir, tag)
	if err != nil {
		return fmt.Errorf("failed to load event decks: %w", err)
	}

	// For now, just display basic statistics
	displayEventStats(collection)

	return nil
}

func eventDeckStatsCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	apiToken := cmd.String("api-token")

	if apiToken == "" {
		return fmt.Errorf("API token is required")
	}

	// Load event deck collection
	dataDir := cmd.String("data-dir")
	collection, err := loadEventDeckCollection(dataDir, tag)
	if err != nil {
		return fmt.Errorf("failed to load event decks: %w", err)
	}

	// For now, just display basic statistics
	displayEventStats(collection)

	return nil
}

// Helper functions for displaying data
func displayEventSummary(collection *events.EventDeckCollection) {
	fmt.Printf("\nEvent Summary:\n")
	fmt.Printf("==============\n")
	fmt.Printf("Total Events: %d\n", len(collection.Decks))

	// Count by event type
	typeCounts := make(map[events.EventType]int)
	for _, deck := range collection.Decks {
		typeCounts[deck.EventType]++
	}

	fmt.Printf("\nBy Type:\n")
	for eventType, count := range typeCounts {
		fmt.Printf("  %s: %d\n", eventType, count)
	}

	// Calculate overall stats
	totalBattles := 0
	totalWins := 0
	totalCrowns := 0
	bestStreak := 0

	for _, deck := range collection.Decks {
		totalBattles += deck.Performance.TotalBattles()
		totalWins += deck.Performance.Wins
		totalCrowns += deck.Performance.CrownsEarned
		if deck.Performance.BestStreak > bestStreak {
			bestStreak = deck.Performance.BestStreak
		}
	}

	overallWinRate := float64(0)
	if totalBattles > 0 {
		overallWinRate = float64(totalWins) / float64(totalBattles)
	}

	fmt.Printf("\nOverall Performance:\n")
	fmt.Printf("  Total Battles: %d\n", totalBattles)
	fmt.Printf("  Total Wins: %d\n", totalWins)
	fmt.Printf("  Overall Win Rate: %.1f%%\n", overallWinRate*100)
	fmt.Printf("  Total Crowns Earned: %d\n", totalCrowns)
	fmt.Printf("  Best Win Streak: %d\n", bestStreak)
}

func displayEventList(decks []events.EventDeck) {
	fmt.Printf("\nEvent Decks:\n")
	fmt.Printf("============\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Event Name\tType\tBattles\tWins\tWin Rate\tBest Streak\n")
	fmt.Fprintf(w, "-----------\t----\t--------\t----\t--------\t-----------\n")

	for _, deck := range decks {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%.1f%%\t%d\n",
			deck.EventName,
			deck.EventType,
			deck.Performance.TotalBattles(),
			deck.Performance.Wins,
			deck.Performance.WinRate*100,
			deck.Performance.BestStreak)
	}

	w.Flush()
}

// Simplified display functions for now
func displayEventStats(collection *events.EventDeckCollection) {
	fmt.Printf("\nEvent Statistics:\n")
	fmt.Printf("==================\n")
	fmt.Printf("Total Events: %d\n", len(collection.Decks))

	// Count by event type
	typeCounts := make(map[string]int)
	for _, deck := range collection.Decks {
		typeCounts[string(deck.EventType)]++
	}

	fmt.Printf("\nBy Type:\n")
	for eventType, count := range typeCounts {
		fmt.Printf("  %s: %d\n", eventType, count)
	}
}

func displayIndividualDeck(deck *events.EventDeck) {
	fmt.Printf("\n%s (%s)\n", deck.EventName, deck.EventType)
	fmt.Printf("Battles: %d | Wins: %d | Win Rate: %.1f%%\n",
		deck.Performance.TotalBattles(),
		deck.Performance.Wins,
		deck.Performance.WinRate*100)

	fmt.Printf("Deck: ")
	for i, card := range deck.Deck.Cards {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%s", card.Name)
	}
	fmt.Printf("\n")
}

// Helper functions for saving/loading data
func saveEventDeckCollection(dataDir string, collection *events.EventDeckCollection) error {
	eventDir := filepath.Join(dataDir, "event_decks")
	if err := os.MkdirAll(eventDir, 0755); err != nil {
		return fmt.Errorf("failed to create event_decks directory: %w", err)
	}

	// In a real implementation, marshal collection to JSON
	return nil
}

func loadEventDeckCollection(dataDir, playerTag string) (*events.EventDeckCollection, error) {
	// In a real implementation, unmarshal from JSON from:
	// filepath.Join(dataDir, "event_decks", fmt.Sprintf("%s.json", playerTag))

	// In a real implementation, unmarshal from JSON
	// For now, return empty collection
	return &events.EventDeckCollection{
		PlayerTag: playerTag,
		Decks:     []events.EventDeck{},
	}, nil
}

func filterEventDecks(collection *events.EventDeckCollection, eventType string, days, minBattles int) []events.EventDeck {
	var filtered []events.EventDeck

	for _, deck := range collection.Decks {
		// Filter by event type
		if eventType != "" && string(deck.EventType) != eventType {
			continue
		}

		// Filter by days
		if days > 0 {
			cutoff := deck.StartTime.AddDate(0, 0, -days)
			if deck.StartTime.Before(cutoff) {
				continue
			}
		}

		// Filter by minimum battles
		if deck.Performance.TotalBattles() < minBattles {
			continue
		}

		filtered = append(filtered, deck)
	}

	return filtered
}

func sortEventDecks(decks []events.EventDeck, sortBy string) {
	// In a real implementation, sort decks based on sortBy criteria
	// For now, decks are already in chronological order
}
