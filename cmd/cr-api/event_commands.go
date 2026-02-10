package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/exporter/csv"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/events"
	"github.com/urfave/cli/v3"
)

// eventTagFlag returns the standard player tag flag for event commands
func eventTagFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "tag",
		Aliases:  []string{"p"},
		Usage:    "Player tag (without #)",
		Required: true,
	}
}

// eventExportCSVFlag returns the standard export-csv flag
func eventExportCSVFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:  "export-csv",
		Usage: "Export to CSV",
	}
}

// addEventScanCommand creates the events scan subcommand
func addEventScanCommand() *cli.Command {
	return &cli.Command{
		Name:  "scan",
		Usage: "Scan player's battle log for event decks",
		Flags: []cli.Flag{
			eventTagFlag(),
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
			eventExportCSVFlag(),
		},
		Action: eventScanCommand,
	}
}

// addEventListCommand creates the events list subcommand
func addEventListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List tracked event decks for a player",
		Flags: []cli.Flag{
			eventTagFlag(),
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
			eventExportCSVFlag(),
		},
		Action: eventListCommand,
	}
}

// addEventAnalyzeCommand creates the events analyze subcommand
func addEventAnalyzeCommand() *cli.Command {
	return &cli.Command{
		Name:  "analyze",
		Usage: "Analyze event deck performance",
		Flags: []cli.Flag{
			eventTagFlag(),
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
			eventExportCSVFlag(),
		},
		Action: eventAnalyzeCommand,
	}
}

// addEventCompareCommand creates the events compare subcommand
func addEventCompareCommand() *cli.Command {
	return &cli.Command{
		Name:  "compare",
		Usage: "Compare performance between different event types or decks",
		Flags: []cli.Flag{
			eventTagFlag(),
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
			eventExportCSVFlag(),
		},
		Action: eventCompareCommand,
	}
}

// addEventDeckStatsCommand creates the events deck-stats subcommand
func addEventDeckStatsCommand() *cli.Command {
	return &cli.Command{
		Name:  "deck-stats",
		Usage: "Show statistics for cards used in events",
		Flags: []cli.Flag{
			eventTagFlag(),
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
			eventExportCSVFlag(),
		},
		Action: eventDeckStatsCommand,
	}
}

// addEventCommands adds event-related subcommands to the CLI
func addEventCommands() *cli.Command {
	return &cli.Command{
		Name:  "events",
		Usage: "Event deck tracking and analysis commands",
		Commands: []*cli.Command{
			addEventScanCommand(),
			addEventListCommand(),
			addEventAnalyzeCommand(),
			addEventCompareCommand(),
			addEventDeckStatsCommand(),
		},
	}
}

// Battle filtering helpers
func filterBattlesByDays(battles []clashroyale.Battle, days int) []clashroyale.Battle {
	if days <= 0 {
		return battles
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	filtered := make([]clashroyale.Battle, 0, len(battles))
	for _, battle := range battles {
		if !battle.UTCDate.Before(cutoff) {
			filtered = append(filtered, battle)
		}
	}
	return filtered
}

func filterEventDecksByEventTypes(decks []events.EventDeck, eventTypes []string) []events.EventDeck {
	if len(eventTypes) == 0 {
		return decks
	}

	allowedTypes := make(map[events.EventType]bool)
	for _, et := range eventTypes {
		allowedTypes[events.EventType(et)] = true
	}

	filtered := make([]events.EventDeck, 0, len(decks))
	for _, deck := range decks {
		if allowedTypes[deck.EventType] {
			filtered = append(filtered, deck)
		}
	}
	return filtered
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
		printf("Scanning battle log for player %s (last %d days)\n", tag, days)
		if len(eventTypes) > 0 {
			printf("Event types to scan: %v\n", eventTypes)
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
	battles = filterBattlesByDays(battles, days)

	if verbose {
		printf("Found %d recent battles to scan\n", len(battles))
	}

	// Create manager and parse event decks
	manager := events.NewManager(dataDir)
	importedDecks, err := manager.ImportFromBattleLogs(battles, tag)
	if err != nil {
		return fmt.Errorf("failed to import event decks: %w", err)
	}

	// Filter by event types if specified
	filteredDecks := filterEventDecksByEventTypes(importedDecks, eventTypes)

	if verbose {
		printf("Successfully imported %d event decks (after filtering: %d)\n", len(importedDecks), len(filteredDecks))
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
		printf("Event list exported to CSV\n")
	}

	return nil
}

// Event filtering helpers
func filterEventDecksByID(decks []events.EventDeck, eventID string) []events.EventDeck {
	filtered := make([]events.EventDeck, 0)
	for _, deck := range decks {
		if deck.EventID == eventID {
			filtered = append(filtered, deck)
		}
	}
	return filtered
}

func filterEventDecksByType(decks []events.EventDeck, eventType string) []events.EventDeck {
	filtered := make([]events.EventDeck, 0)
	for _, deck := range decks {
		if string(deck.EventType) == eventType {
			filtered = append(filtered, deck)
		}
	}
	return filtered
}

func filterEventDecksByMinBattles(decks []events.EventDeck, minBattles int) []events.EventDeck {
	filtered := make([]events.EventDeck, 0)
	for _, deck := range decks {
		if deck.Performance.TotalBattles() >= minBattles {
			filtered = append(filtered, deck)
		}
	}
	return filtered
}

func applyEventDeckFilters(decks []events.EventDeck, eventID, eventType string, minBattles int) []events.EventDeck {
	filtered := decks

	if eventID != "" {
		filtered = filterEventDecksByID(filtered, eventID)
	} else if eventType != "" {
		filtered = filterEventDecksByType(filtered, eventType)
	}

	filtered = filterEventDecksByMinBattles(filtered, minBattles)
	return filtered
}

func eventAnalyzeCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	eventID := cmd.String("event-id")
	eventType := cmd.String("event-type")
	minBattles := cmd.Int("min-battles")
	includeDecks := cmd.Bool("include-decks")
	exportCSV := cmd.Bool("export-csv")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")

	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	if eventID == "" && eventType == "" {
		return fmt.Errorf("either --event-id or --event-type must be specified")
	}

	if verbose {
		printf("Analyzing event data for player %s\n", tag)
		if eventID != "" {
			printf("Event ID: %s\n", eventID)
		} else {
			printf("Event Type: %s\n", eventType)
		}
		printf("Minimum battles: %d\n", minBattles)
	}

	// Load event deck collection
	dataDir := cmd.String("data-dir")
	collection, err := loadEventDeckCollection(dataDir, tag)
	if err != nil {
		return fmt.Errorf("failed to load event decks: %w", err)
	}

	if len(collection.Decks) == 0 {
		printf("No event decks found for player %s\n", tag)
		printf("Try running 'cr-api events scan --tag %s' first to collect event data\n", tag)
		return nil
	}

	// Filter decks based on criteria
	filteredDecks := applyEventDeckFilters(collection.Decks, eventID, eventType, minBattles)

	if len(filteredDecks) == 0 {
		printf("No event decks match the specified criteria\n")
		return nil
	}

	if verbose {
		printf("Analyzing %d event decks\n", len(filteredDecks))
	}

	// Configure analysis options
	analysisOptions := events.DefaultAnalysisOptions()
	analysisOptions.MinBattlesForTopDecks = minBattles

	if eventID != "" {
		// For single event analysis, show all decks
		analysisOptions.LimitTopDecks = 10
	}

	// Perform analysis
	analysis := events.AnalyzeEventDecks(filteredDecks, analysisOptions)

	// Display comprehensive analysis
	displayComprehensiveAnalysis(analysis, includeDecks)

	// Export to CSV if requested
	if exportCSV {
		if err := exportAnalysisToCSV(dataDir, analysis); err != nil {
			return fmt.Errorf("failed to export analysis to CSV: %w", err)
		}
		if verbose {
			fmt.Println("Analysis exported to CSV")
		}
	}

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
	printf("\nEvent Summary:\n")
	printf("==============\n")
	printf("Total Events: %d\n", len(collection.Decks))

	// Count by event type
	typeCounts := make(map[events.EventType]int)
	for _, deck := range collection.Decks {
		typeCounts[deck.EventType]++
	}

	printf("\nBy Type:\n")
	for eventType, count := range typeCounts {
		printf("  %s: %d\n", eventType, count)
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

	printf("\nOverall Performance:\n")
	printf("  Total Battles: %d\n", totalBattles)
	printf("  Total Wins: %d\n", totalWins)
	printf("  Overall Win Rate: %.1f%%\n", overallWinRate*100)
	printf("  Total Crowns Earned: %d\n", totalCrowns)
	printf("  Best Win Streak: %d\n", bestStreak)
}

func displayEventList(decks []events.EventDeck) {
	printf("\nEvent Decks:\n")
	printf("============\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "Event Name\tType\tBattles\tWins\tWin Rate\tBest Streak\n")
	fprintf(w, "-----------\t----\t--------\t----\t--------\t-----------\n")

	for _, deck := range decks {
		fprintf(w, "%s\t%s\t%d\t%d\t%.1f%%\t%d\n",
			deck.EventName,
			deck.EventType,
			deck.Performance.TotalBattles(),
			deck.Performance.Wins,
			deck.Performance.WinRate*100,
			deck.Performance.BestStreak)
	}

	flushWriter(w)
}

// Simplified display functions for now
func displayEventStats(collection *events.EventDeckCollection) {
	printf("\nEvent Statistics:\n")
	printf("==================\n")
	printf("Total Events: %d\n", len(collection.Decks))

	// Count by event type
	typeCounts := make(map[string]int)
	for _, deck := range collection.Decks {
		typeCounts[string(deck.EventType)]++
	}

	printf("\nBy Type:\n")
	for eventType, count := range typeCounts {
		printf("  %s: %d\n", eventType, count)
	}
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

// displayComprehensiveAnalysis displays detailed event analysis results
func displayComprehensiveAnalysis(analysis *events.EventAnalysis, includeDecks bool) {
	printf("\n=== Event Analysis for %s ===\n", analysis.PlayerTag)
	printf("Generated: %s\n", analysis.AnalysisTime.Format("2006-01-02 15:04:05"))
	printf("Total Decks Analyzed: %d\n\n", analysis.TotalDecks)

	// Summary Statistics
	printf("📊 Overall Performance Summary:\n")
	printf("==============================\n")
	printf("Total Battles: %d\n", analysis.Summary.TotalBattles)
	printf("Total Wins: %d\n", analysis.Summary.TotalWins)
	printf("Total Losses: %d\n", analysis.Summary.TotalLosses)
	printf("Overall Win Rate: %.1f%%\n", analysis.Summary.OverallWinRate*100)
	printf("Average Crowns per Battle: %.1f\n", analysis.Summary.AvgCrownsPerBattle)
	printf("Average Deck Elixir: %.1f\n\n", analysis.Summary.AvgDeckElixir)

	// Card Analysis
	printf("🎴 Card Usage Analysis:\n")
	printf("=======================\n")
	printf("Total Unique Cards Used: %d\n\n", analysis.CardAnalysis.TotalUniqueCards)

	if len(analysis.CardAnalysis.MostUsedCards) > 0 {
		printf("Most Used Cards:\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fprintf(w, "Card\tTimes Used\n")
		fprintf(w, "----\t----------\n")
		for _, card := range analysis.CardAnalysis.MostUsedCards {
			fprintf(w, "%s\t%d\n", card.CardName, card.Count)
		}
		flushWriter(w)
		fmt.Println()
	}

	if len(analysis.CardAnalysis.HighestWinRateCards) > 0 {
		printf("Highest Win Rate Cards (3+ battles):\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fprintf(w, "Card\tWin Rate\n")
		fprintf(w, "----\t--------\n")
		for _, card := range analysis.CardAnalysis.HighestWinRateCards {
			fprintf(w, "%s\t%.1f%%\n", card.CardName, card.WinRate*100)
		}
		flushWriter(w)
		fmt.Println()
	}

	// Elixir Analysis
	printf("⚡ Elixir Cost Analysis:\n")
	printf("========================\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "Elixir Range\tDecks\tAvg Win Rate\n")
	fprintf(w, "-------------\t-----\t------------\n")
	fprintf(w, "%s\t%d\t%.1f%%\n", analysis.ElixirAnalysis.LowElixir.Range,
		analysis.ElixirAnalysis.LowElixir.DeckCount,
		analysis.ElixirAnalysis.LowElixir.AvgWinRate*100)
	fprintf(w, "%s\t%d\t%.1f%%\n", analysis.ElixirAnalysis.MidElixir.Range,
		analysis.ElixirAnalysis.MidElixir.DeckCount,
		analysis.ElixirAnalysis.MidElixir.AvgWinRate*100)
	fprintf(w, "%s\t%d\t%.1f%%\n", analysis.ElixirAnalysis.HighElixir.Range,
		analysis.ElixirAnalysis.HighElixir.DeckCount,
		analysis.ElixirAnalysis.HighElixir.AvgWinRate*100)
	flushWriter(w)
	fmt.Println()

	// Event Breakdown
	if len(analysis.EventBreakdown) > 0 {
		printf("🏆 Performance by Event Type:\n")
		printf("===============================\n")
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fprintf(w, "Event Type\tEvents\tWins\tLosses\tWin Rate\n")
		fprintf(w, "-----------\t------\t----\t------\t--------\n")
		for eventType, stats := range analysis.EventBreakdown {
			winRate := float64(0)
			if stats.Wins+stats.Losses > 0 {
				winRate = float64(stats.Wins) / float64(stats.Wins+stats.Losses)
			}
			fprintf(w, "%s\t%d\t%d\t%d\t%.1f%%\n",
				eventType, stats.Count, stats.Wins, stats.Losses, winRate*100)
		}
		flushWriter(w)
		fmt.Println()
	}

	// Top Performing Decks
	if len(analysis.TopDecks) > 0 {
		printf("⭐ Top Performing Decks:\n")
		printf("========================\n")
		for i, deck := range analysis.TopDecks {
			printf("%d. %s (%s)\n", i+1, deck.EventName, deck.EventType)
			printf("   Record: %s | Win Rate: %.1f%% | Avg Elixir: %.1f\n",
				deck.Record, deck.WinRate*100, deck.AvgElixir)
			printf("   Deck: %s\n", strings.Join(deck.Deck, ", "))
			fmt.Println()
		}
	}

	if includeDecks && analysis.TotalDecks > 0 {
		printf("📋 Individual Deck Details:\n")
		printf("=============================\n")
		printf("Run 'cr-api events list --tag %s' for detailed deck information\n", analysis.PlayerTag)
	}
}

// exportAnalysisToCSV exports event analysis to CSV format
func exportAnalysisToCSV(dataDir string, analysis *events.EventAnalysis) error {
	analysisDir := filepath.Join(dataDir, "csv", "analysis")
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		return fmt.Errorf("failed to create analysis directory: %w", err)
	}

	// Create summary CSV
	summaryFile := filepath.Join(analysisDir, fmt.Sprintf("event_analysis_%s_%s.csv",
		analysis.PlayerTag, analysis.AnalysisTime.Format("20060102_150405")))

	file, err := os.Create(summaryFile)
	if err != nil {
		return fmt.Errorf("failed to create analysis CSV: %w", err)
	}
	defer closeFile(file)

	// Write analysis summary
	fprintf(file, "Event Analysis Summary\n")
	fprintf(file, "Player Tag,%s\n", analysis.PlayerTag)
	fprintf(file, "Analysis Time,%s\n", analysis.AnalysisTime.Format("2006-01-02 15:04:05"))
	fprintf(file, "Total Decks,%d\n", analysis.TotalDecks)
	fprintf(file, "\nPerformance Summary\n")
	fprintf(file, "Total Battles,%d\n", analysis.Summary.TotalBattles)
	fprintf(file, "Total Wins,%d\n", analysis.Summary.TotalWins)
	fprintf(file, "Total Losses,%d\n", analysis.Summary.TotalLosses)
	fprintf(file, "Overall Win Rate,%.2f\n", analysis.Summary.OverallWinRate)
	fprintf(file, "Average Crowns per Battle,%.2f\n", analysis.Summary.AvgCrownsPerBattle)
	fprintf(file, "Average Deck Elixir,%.1f\n", analysis.Summary.AvgDeckElixir)

	// Write card usage
	if len(analysis.CardAnalysis.MostUsedCards) > 0 {
		fprintf(file, "\nMost Used Cards\n")
		fprintf(file, "Card,Times Used\n")
		for _, card := range analysis.CardAnalysis.MostUsedCards {
			fprintf(file, "%s,%d\n", card.CardName, card.Count)
		}
	}

	// Write highest win rate cards
	if len(analysis.CardAnalysis.HighestWinRateCards) > 0 {
		fprintf(file, "\nHighest Win Rate Cards\n")
		fprintf(file, "Card,Win Rate\n")
		for _, card := range analysis.CardAnalysis.HighestWinRateCards {
			fprintf(file, "%s,%.2f\n", card.CardName, card.WinRate)
		}
	}

	// Write event breakdown
	if len(analysis.EventBreakdown) > 0 {
		fprintf(file, "\nEvent Type Performance\n")
		fprintf(file, "Event Type,Count,Wins,Losses,Win Rate\n")
		for eventType, stats := range analysis.EventBreakdown {
			winRate := float64(0)
			if stats.Wins+stats.Losses > 0 {
				winRate = float64(stats.Wins) / float64(stats.Wins+stats.Losses)
			}
			fprintf(file, "%s,%d,%d,%d,%.2f\n",
				eventType, stats.Count, stats.Wins, stats.Losses, winRate)
		}
	}

	// Write top decks
	if len(analysis.TopDecks) > 0 {
		fprintf(file, "\nTop Performing Decks\n")
		fprintf(file, "Rank,Event Name,Event Type,Win Rate,Record,Avg Elixir,Deck\n")
		for i, deck := range analysis.TopDecks {
			deckStr := strings.Join(deck.Deck, "|")
			fprintf(file, "%d,%s,%s,%.2f,%s,%.1f,%s\n",
				i+1, deck.EventName, deck.EventType, deck.WinRate,
				deck.Record, deck.AvgElixir, deckStr)
		}
	}

	return nil
}
