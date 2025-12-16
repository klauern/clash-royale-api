//go:build example
// +build example

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func main() {
	// This example demonstrates how to use the deck building system
	// to create optimized decks based on a player's card collection.

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

	// Get player data
	fmt.Printf("Fetching data for player %s...\n", playerTag)
	player, err := client.GetPlayer(playerTag)
	if err != nil {
		log.Fatalf("Failed to get player data: %v", err)
	}

	fmt.Printf("Successfully retrieved data for %s (%s)\n", player.Name, player.Tag)
	fmt.Printf("Trophy count: %d | Level: %d\n", player.Trophies, player.ExpLevel)
	fmt.Printf("Total cards: %d | Cards in collection: %d\n", len(player.Cards), len(player.Cards))

	// Analyze card collection
	fmt.Println("\nAnalyzing card collection...")
	analysisOptions := analysis.DefaultAnalysisOptions()
	analysisOptions.TopN = 20 // Show top 20 upgrade priorities
	analysisOptions.MinPriorityScore = 20

	cardAnalysis, err := analysis.AnalyzeCardCollection(player, analysisOptions)
	if err != nil {
		log.Fatalf("Failed to analyze card collection: %v", err)
	}

	// Display collection summary
	fmt.Printf("\n=== Collection Summary ===\n")
	fmt.Printf("Total Cards: %d\n", cardAnalysis.TotalCards)
	fmt.Printf("Max Level Cards: %d\n", len(cardAnalysis.MaxLevelCards))
	fmt.Printf("Average Card Level: %.1f\n", cardAnalysis.Summary.AvgCardLevel)
	fmt.Printf("Collection Completion: %.1f%%\n", cardAnalysis.Summary.CompletionPercent)

	// Display top upgrade priorities
	if len(cardAnalysis.UpgradePriority) > 0 {
		fmt.Printf("\n=== Top 10 Upgrade Priorities ===\n")
		for i, priority := range cardAnalysis.UpgradePriority[:10] {
			fmt.Printf("%d. %s (%s) - Priority: %.0f (%s)\n",
				i+1, priority.CardName, priority.Rarity, priority.PriorityScore, priority.Priority)
			fmt.Printf("   Level: %d/%d | Need: %d cards\n",
				priority.CurrentLevel, priority.MaxLevel, priority.CardsNeeded)
			fmt.Printf("   Reasons: %v\n", priority.Reasons)
		}
	}

	// Build optimized deck using the deck builder
	fmt.Println("\n=== Building Optimized Deck ===")
	dataDir := "./data"
	builder := deck.NewBuilder(dataDir)

	// Convert analysis to deck builder format
	deckAnalysis := deck.ConvertAnalysisForDeckBuilding(cardAnalysis)

	// Build deck from analysis
	deckRecommendation, err := builder.BuildDeckFromAnalysis(deckAnalysis)
	if err != nil {
		log.Fatalf("Failed to build deck: %v", err)
	}

	fmt.Printf("Recommended Deck\n")
	fmt.Printf("Average Elixir: %.2f\n", deckRecommendation.AvgElixir)
	if deckRecommendation.AnalysisTime != "" {
		fmt.Printf("Analysis Time: %s\n", deckRecommendation.AnalysisTime)
	}

	fmt.Println("\nCards:")
	for i, card := range deckRecommendation.Deck {
		fmt.Printf("%d. %s\n", i+1, card)
	}

	fmt.Println("\nDetailed Card Information:")
	for _, detail := range deckRecommendation.DeckDetail {
		roleIcon := ""
		switch detail.Role {
		case "win_conditions":
			roleIcon = "🎯"
		case "buildings":
			roleIcon = "🏰"
		case "spells_big":
			roleIcon = "💥"
		case "spells_small":
			roleIcon = "⚡"
		case "support":
			roleIcon = "🤺"
		case "cycle":
			roleIcon = "🔄"
		}

		fmt.Printf("%s %s (Lv.%d) - %s - %.1f elixir - Score: %.3f\n",
			roleIcon, detail.Name, detail.Level, detail.Rarity,
			float64(detail.Elixir), detail.Score)
	}

	// Print strategic notes
	if len(deckRecommendation.Notes) > 0 {
		fmt.Println("\n💡 Strategic Notes:")
		for _, note := range deckRecommendation.Notes {
			fmt.Printf("• %s\n", note)
		}
	}

	fmt.Println("\nExample completed successfully!")
	fmt.Println("This demonstrates how to:")
	fmt.Println("- Analyze a player's card collection")
	fmt.Println("- Generate upgrade priorities")
	fmt.Println("- Build optimized decks for different game modes")
}