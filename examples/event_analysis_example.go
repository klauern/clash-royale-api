//go:build example
// +build example

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/events"
)

func main() {
	// This example demonstrates how to use the event analysis system
	// to analyze performance in Clash Royale events (challenges, tournaments, etc.)

	// Set your API token
	apiToken := os.Getenv("CLASH_ROYALE_API_TOKEN")
	if apiToken == "" {
		fmt.Println("Please set CLASH_ROYALE_API_TOKEN environment variable")
		fmt.Println("Example: export CLASH_ROYALE_API_TOKEN=your_token_here")
		os.Exit(1)
	}

	// Player tag to analyze (without #)
	playerTag := "#LQLPQGV8Q" // Example player tag

	// Create API client
	client := clashroyale.NewClient(apiToken)

	// Get player battle logs
	fmt.Printf("Fetching battle logs for player %s...\n", playerTag)
	battleLog, err := client.GetPlayerBattleLog(playerTag)
	if err != nil {
		log.Fatalf("Failed to get battle logs: %v", err)
	}

	battles := []clashroyale.Battle(*battleLog)
	fmt.Printf("Found %d battles in battle log\n", len(battles))

	// Create event manager and scan for event decks
	fmt.Println("\nScanning battle logs for event decks...")
	dataDir := "./data"
	manager := events.NewManager(dataDir)

	eventDecks, err := manager.ImportFromBattleLogs(battles, playerTag)
	if err != nil {
		log.Fatalf("Failed to import event decks: %v", err)
	}

	fmt.Printf("Found %d event decks\n", len(eventDecks))

	if len(eventDecks) == 0 {
		fmt.Println("No event decks found. This player may not have participated in recent events.")
		fmt.Println("Try with a different player tag or check back after they participate in challenges.")
		return
	}

	// Display basic event summary
	fmt.Println("\n=== Event Summary ===")
	eventTypes := make(map[events.EventType]int)
	totalBattles := 0
	totalWins := 0

	for _, deck := range eventDecks {
		eventTypes[deck.EventType]++
		totalBattles += deck.Performance.TotalBattles()
		totalWins += deck.Performance.Wins
	}

	for eventType, count := range eventTypes {
		fmt.Printf("%s: %d events\n", eventType, count)
	}

	if totalBattles > 0 {
		overallWinRate := float64(totalWins) / float64(totalBattles) * 100
		fmt.Printf("\nOverall Performance: %d wins / %d battles (%.1f%% win rate)\n",
			totalWins, totalBattles, overallWinRate)
	}

	// Perform comprehensive analysis
	fmt.Println("\n=== Performing Event Analysis ===")
	analysisOptions := events.DefaultAnalysisOptions()
	analysisOptions.MinBattlesForTopDecks = 3
	analysisOptions.LimitTopDecks = 5

	analysis := events.AnalyzeEventDecks(eventDecks, analysisOptions)

	// Display analysis results
	fmt.Printf("\nAnalysis generated at: %s\n", analysis.AnalysisTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Total decks analyzed: %d\n\n", analysis.TotalDecks)

	// Performance summary
	fmt.Println("📊 Performance Summary:")
	fmt.Printf("Total Battles: %d\n", analysis.Summary.TotalBattles)
	fmt.Printf("Total Wins: %d\n", analysis.Summary.TotalWins)
	fmt.Printf("Total Losses: %d\n", analysis.Summary.TotalLosses)
	fmt.Printf("Overall Win Rate: %.1f%%\n", analysis.Summary.OverallWinRate*100)
	fmt.Printf("Average Crowns per Battle: %.1f\n", analysis.Summary.AvgCrownsPerBattle)
	fmt.Printf("Average Deck Elixir: %.1f\n", analysis.Summary.AvgDeckElixir)

	// Card usage analysis
	fmt.Println("\n🎴 Card Usage Analysis:")
	fmt.Printf("Total Unique Cards Used: %d\n", analysis.CardAnalysis.TotalUniqueCards)

	if len(analysis.CardAnalysis.MostUsedCards) > 0 {
		fmt.Println("\nMost Used Cards:")
		for i, card := range analysis.CardAnalysis.MostUsedCards[:5] {
			fmt.Printf("%d. %s (used %d times)\n", i+1, card.CardName, card.Count)
		}
	}

	if len(analysis.CardAnalysis.HighestWinRateCards) > 0 {
		fmt.Println("\nHighest Win Rate Cards:")
		for i, card := range analysis.CardAnalysis.HighestWinRateCards[:5] {
			fmt.Printf("%d. %s (%.1f%% win rate)\n", i+1, card.CardName, card.WinRate*100)
		}
	}

	// Elixir analysis
	fmt.Println("\n⚡ Elixir Cost Analysis:")
	fmt.Printf("Low Elixir Decks (< 3.5): %d decks (%.1f%% win rate)\n",
		analysis.ElixirAnalysis.LowElixir.DeckCount,
		analysis.ElixirAnalysis.LowElixir.AvgWinRate*100)
	fmt.Printf("Mid Elixir Decks (3.5-4.4): %d decks (%.1f%% win rate)\n",
		analysis.ElixirAnalysis.MidElixir.DeckCount,
		analysis.ElixirAnalysis.MidElixir.AvgWinRate*100)
	fmt.Printf("High Elixir Decks (> 4.4): %d decks (%.1f%% win rate)\n",
		analysis.ElixirAnalysis.HighElixir.DeckCount,
		analysis.ElixirAnalysis.HighElixir.AvgWinRate*100)

	// Event type breakdown
	if len(analysis.EventBreakdown) > 0 {
		fmt.Println("\n🏆 Performance by Event Type:")
		for eventType, stats := range analysis.EventBreakdown {
			winRate := float64(0)
			if stats.Wins+stats.Losses > 0 {
				winRate = float64(stats.Wins) / float64(stats.Wins+stats.Losses) * 100
			}
			fmt.Printf("%s: %d events, %d wins, %d losses (%.1f%% win rate)\n",
				eventType, stats.Count, stats.Wins, stats.Losses, winRate)
		}
	}

	// Top performing decks
	if len(analysis.TopDecks) > 0 {
		fmt.Println("\n⭐ Top Performing Decks:")
		for i, deck := range analysis.TopDecks {
			fmt.Printf("%d. %s (%s)\n", i+1, deck.EventName, deck.EventType)
			fmt.Printf("   Record: %s | Win Rate: %.1f%% | Avg Elixir: %.1f\n",
				deck.Record, deck.WinRate*100, deck.AvgElixir)
			fmt.Printf("   Deck: %s\n", formatDeckList(deck.Deck))
		}
	}

	// Example: Filter and analyze specific event type
	fmt.Println("\n=== Analyzing Only Challenges ===")
	challengeType := events.EventTypeChallenge
	var challengeDecks []events.EventDeck
	for _, deck := range eventDecks {
		if deck.EventType == challengeType {
			challengeDecks = append(challengeDecks, deck)
		}
	}

	if len(challengeDecks) > 0 {
		challengeAnalysis := events.AnalyzeEventDecks(challengeDecks, analysisOptions)
		fmt.Printf("Challenge Performance: %d wins / %d battles (%.1f%% win rate)\n",
			challengeAnalysis.Summary.TotalWins,
			challengeAnalysis.Summary.TotalBattles,
			challengeAnalysis.Summary.OverallWinRate*100)
	} else {
		fmt.Println("No challenge events found")
	}

	// Example: Export event decks to different formats
	fmt.Println("\n=== Export Examples ===")

	// Create exporter for CSV format
	csvOptions := events.DefaultExportOptions()
	csvOptions.Format = events.FormatCSV
	csvOptions.OutputDir = "./exports"
	csvOptions.MinBattles = 3
	csvOptions.GroupByEvent = true

	csvExporter := events.NewExporter(csvOptions)
	collection := &events.EventDeckCollection{
		PlayerTag:   playerTag,
		Decks:       eventDecks,
		LastUpdated: time.Now(),
	}

	if err := csvExporter.Export(collection); err != nil {
		log.Printf("Failed to export to CSV: %v", err)
	} else {
		fmt.Println("✅ Event data exported to CSV format in ./exports/")
	}

	// Create exporter for JSON format
	jsonOptions := events.DefaultExportOptions()
	jsonOptions.Format = events.FormatJSON
	jsonOptions.OutputDir = "./exports"
	jsonOptions.IncludeBattles = true

	jsonExporter := events.NewExporter(jsonOptions)
	if err := jsonExporter.Export(collection); err != nil {
		log.Printf("Failed to export to JSON: %v", err)
	} else {
		fmt.Println("✅ Event data exported to JSON format in ./exports/")
	}

	// Create exporter for deck list format
	deckListOptions := events.DefaultExportOptions()
	deckListOptions.Format = events.FormatDeckList
	deckListOptions.OutputDir = "./exports"
	deckListOptions.MinBattles = 5
	deckListOptions.GroupByEvent = true

	deckListExporter := events.NewExporter(deckListOptions)
	if err := deckListExporter.Export(collection); err != nil {
		log.Printf("Failed to export to deck list: %v", err)
	} else {
		fmt.Println("✅ Event data exported to deck list format in ./exports/")
	}

	fmt.Println("\n🎉 Event analysis example completed!")
	fmt.Println("\nThis example demonstrates:")
	fmt.Println("- Scanning battle logs for event decks")
	fmt.Println("- Performing comprehensive event analysis")
	fmt.Println("- Analyzing card usage patterns")
	fmt.Println("- Identifying top performing decks")
	fmt.Println("- Exporting data in multiple formats")
}

// Helper function to format deck list nicely
func formatDeckList(cards []string) string {
	if len(cards) == 0 {
		return "No cards"
	}

	result := ""
	for i, card := range cards {
		if i > 0 {
			result += ", "
		}
		result += card
	}
	return result
}
