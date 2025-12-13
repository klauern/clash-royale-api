//go:build example

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func main() {
	// Create a new deck builder
	builder := deck.NewBuilder("data")

	// Example analysis data - in real usage this would come from the card analyzer
	analysis := deck.CardAnalysis{
		CardLevels: map[string]deck.CardLevelData{
			"Hog Rider": {
				Level:    8,
				MaxLevel: 13,
				Rarity:   "Rare",
				Elixir:   4,
			},
			"Fireball": {
				Level:    7,
				MaxLevel: 11,
				Rarity:   "Rare",
				Elixir:   4,
			},
			"Zap": {
				Level:    11,
				MaxLevel: 13,
				Rarity:   "Common",
				Elixir:   2,
			},
			"Cannon": {
				Level:    11,
				MaxLevel: 13,
				Rarity:   "Common",
				Elixir:   3,
			},
			"Archers": {
				Level:    10,
				MaxLevel: 13,
				Rarity:   "Common",
				Elixir:   3,
			},
			"Knight": {
				Level:    11,
				MaxLevel: 13,
				Rarity:   "Common",
				Elixir:   3,
			},
			"Skeletons": {
				Level:    11,
				MaxLevel: 13,
				Rarity:   "Common",
				Elixir:   1,
			},
			"Valkyrie": {
				Level:    7,
				MaxLevel: 11,
				Rarity:   "Rare",
				Elixir:   4,
			},
			"Baby Dragon": {
				Level:    5,
				MaxLevel: 11,
				Rarity:   "Epic",
				Elixir:   4,
			},
			"Musketeer": {
				Level:    6,
				MaxLevel: 13,
				Rarity:   "Rare",
				Elixir:   4,
			},
			"Wizard": {
				Level:    4,
				MaxLevel: 11,
				Rarity:   "Epic",
				Elixir:   5,
			},
			"Goblin Barrel": {
				Level:    2,
				MaxLevel: 13,
				Rarity:   "Rare",
				Elixir:   3,
			},
		},
		AnalysisTime: "2024-01-15T10:30:00Z",
	}

	// Build a deck from the analysis
	fmt.Println("🏗️  Building optimized deck from card collection...")
	deckRecommendation, err := builder.BuildDeckFromAnalysis(analysis)
	if err != nil {
		log.Fatalf("Failed to build deck: %v", err)
	}

	// Print the recommended deck
	fmt.Println("\n🎯 Recommended Deck:")
	fmt.Println("===============")
	for i, card := range deckRecommendation.Deck {
		fmt.Printf("%d. %s\n", i+1, card)
	}

	// Print detailed information
	fmt.Println("\n📊 Deck Details:")
	fmt.Println("===============")
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

	// Print summary
	fmt.Printf("\n📈 Summary:\n")
	fmt.Printf("Average Elixir: %.2f\n", deckRecommendation.AvgElixir)
	if deckRecommendation.AnalysisTime != "" {
		fmt.Printf("Analysis Time: %s\n", deckRecommendation.AnalysisTime)
	}

	// Print strategic notes
	if len(deckRecommendation.Notes) > 0 {
		fmt.Println("\n💡 Strategic Notes:")
		for _, note := range deckRecommendation.Notes {
			fmt.Printf("• %s\n", note)
		}
	}

	// Save the deck to disk
	deckPath, err := builder.SaveDeck(deckRecommendation, "", "#EXAMPLE123")
	if err != nil {
		log.Printf("Warning: Failed to save deck: %v", err)
	} else {
		fmt.Printf("\n💾 Deck saved to: %s\n", deckPath)
	}

	// Example of loading from a file (if analysis file exists)
	fmt.Println("\n🔍 Example: Loading deck from saved analysis...")

	// Save the analysis first for demo purposes
	analysisData, _ := json.MarshalIndent(analysis, "", "  ")
	analysisPath := "example_analysis.json"
	err = os.WriteFile(analysisPath, analysisData, 0644)
	if err != nil {
		log.Printf("Warning: Could not save analysis file: %v", err)
	} else {
		defer os.Remove(analysisPath) // Clean up

		// Load deck from file
		deckFromFile, err := builder.BuildDeckFromFile(analysisPath)
		if err != nil {
			log.Printf("Warning: Failed to load deck from file: %v", err)
		} else {
			fmt.Printf("✅ Successfully loaded deck from file: %v\n", deckFromFile.Deck)
		}
	}

	fmt.Println("\n🎉 Deck building example completed!")
}