package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
	"github.com/urfave/cli/v3"
)

// addDeckCommands adds deck-related subcommands to the CLI
func addDeckCommands() *cli.Command {
	return &cli.Command{
		Name:  "deck",
		Usage: "Deck building and analysis commands",
		Commands: []*cli.Command{
			{
				Name:  "build",
				Usage: "Build an optimized deck based on player's card collection",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "strategy",
						Aliases: []string{"s"},
						Value:   "balanced",
						Usage:   "Deck building strategy: balanced, aggro, control, cycle, splash, spell",
					},
					&cli.Float64Flag{
						Name:  "min-elixir",
						Value: 2.5,
						Usage: "Minimum average elixir for the deck",
					},
					&cli.Float64Flag{
						Name:  "max-elixir",
						Value: 4.5,
						Usage: "Maximum average elixir for the deck",
					},
					&cli.StringSliceFlag{
						Name:  "include-cards",
						Usage: "Specific cards to include in the deck (by name)",
					},
					&cli.StringSliceFlag{
						Name:  "exclude-cards",
						Usage: "Cards to exclude from the deck (by name)",
					},
					&cli.IntFlag{
						Name:  "min-level",
						Value: 1,
						Usage: "Minimum card level to consider",
					},
					&cli.BoolFlag{
						Name:  "prioritize-upgrades",
						Usage: "Prioritize cards that can be upgraded soon",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export deck analysis to CSV",
					},
					&cli.BoolFlag{
						Name:  "save",
						Usage: "Save deck to file",
					},
					&cli.StringFlag{
						Name:  "unlocked-evolutions",
						Usage: "Comma-separated list of cards with unlocked evolutions (overrides UNLOCKED_EVOLUTIONS env var)",
					},
					&cli.IntFlag{
						Name:  "evolution-slots",
						Value: 2,
						Usage: "Number of evolution slots available (default 2)",
					},
					&cli.Float64Flag{
						Name:  "combat-stats-weight",
						Value: 0.25,
						Usage: "Weight for combat stats in scoring (0.0-1.0, where 0=disabled, 0.25=default, 1.0=combat-only)",
					},
					&cli.BoolFlag{
						Name:  "disable-combat-stats",
						Usage: "Disable combat stats completely (use traditional scoring only)",
					},
				},
				Action: deckBuildCommand,
			},
			{
				Name:  "war",
				Usage: "Build a 4-deck war set with no repeated cards",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "deck-count",
						Value: 4,
						Usage: "Number of decks to build (default 4)",
					},
					&cli.StringFlag{
						Name:  "unlocked-evolutions",
						Usage: "Comma-separated list of cards with unlocked evolutions (overrides UNLOCKED_EVOLUTIONS env var)",
					},
					&cli.IntFlag{
						Name:  "evolution-slots",
						Value: 2,
						Usage: "Number of evolution slots available (default 2)",
					},
					&cli.Float64Flag{
						Name:  "combat-stats-weight",
						Value: 0.25,
						Usage: "Weight for combat stats in scoring (0.0-1.0, where 0=disabled, 0.25=default, 1.0=combat-only)",
					},
					&cli.BoolFlag{
						Name:  "disable-combat-stats",
						Usage: "Disable combat stats completely (use traditional scoring only)",
					},
				},
				Action: deckWarCommand,
			},
			{
				Name:  "analyze",
				Usage: "Analyze a deck's strengths and weaknesses",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "cards",
						Aliases:  []string{"c"},
						Usage:    "8 card names for the deck",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export analysis to CSV",
					},
				},
				Action: deckAnalyzeCommand,
			},
			{
				Name:  "optimize",
				Usage: "Optimize an existing deck with available cards",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:     "cards",
						Aliases:  []string{"c"},
						Usage:    "Current 8-card deck to optimize",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "max-changes",
						Value: 4,
						Usage: "Maximum number of cards to change",
					},
					&cli.BoolFlag{
						Name:  "keep-win-con",
						Usage: "Keep win condition cards unchanged",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export optimization results to CSV",
					},
				},
				Action: deckOptimizeCommand,
			},
			{
				Name:  "recommend",
				Usage: "Get deck recommendations based on meta analysis",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "arena",
						Usage: "Filter by arena name",
					},
					&cli.StringFlag{
						Name:  "league",
						Usage: "Filter by league name",
					},
					&cli.IntFlag{
						Name:  "limit",
						Value: 5,
						Usage: "Number of recommendations to return",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export recommendations to CSV",
					},
				},
				Action: deckRecommendCommand,
			},
			{
				Name:  "mulligan",
				Usage: "Generate mulligan guide (opening hand strategy) for a deck",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "cards",
						Aliases:  []string{"c"},
						Usage:    "8 card names for the deck to analyze",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "deck-name",
						Usage: "Custom name for the deck (optional)",
					},
					&cli.BoolFlag{
						Name:  "save",
						Usage: "Save mulligan guide to file",
					},
					&cli.BoolFlag{
						Name:  "json",
						Usage: "Output guide in JSON format",
					},
				},
				Action: deckMulliganCommand,
			},
		},
	}
}

func deckBuildCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	strategy := cmd.String("strategy")
	minElixir := cmd.Float64("min-elixir")
	maxElixir := cmd.Float64("max-elixir")
	saveData := cmd.Bool("save")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	dataDir := cmd.String("data-dir")
	combatStatsWeight := cmd.Float64("combat-stats-weight")
	disableCombatStats := cmd.Bool("disable-combat-stats")
	excludeCards := cmd.StringSlice("exclude-cards")

	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	// Configure combat stats weight
	if disableCombatStats {
		os.Setenv("COMBAT_STATS_WEIGHT", "0")
		if verbose {
			fmt.Printf("Combat stats disabled (using traditional scoring only)\n")
		}
	} else if combatStatsWeight >= 0 && combatStatsWeight <= 1.0 {
		os.Setenv("COMBAT_STATS_WEIGHT", fmt.Sprintf("%.2f", combatStatsWeight))
		if verbose {
			fmt.Printf("Combat stats weight set to: %.2f\n", combatStatsWeight)
		}
	}

	client := clashroyale.NewClient(apiToken)

	if verbose {
		fmt.Printf("Building deck for player %s with strategy: %s\n", tag, strategy)
	}

	// Get player information
	player, err := client.GetPlayer(tag)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	if verbose {
		fmt.Printf("Player: %s (%s)\n", player.Name, player.Tag)
		fmt.Printf("Analyzing %d cards...\n", len(player.Cards))
	}

	// Perform card collection analysis
	analysisOptions := analysis.DefaultAnalysisOptions()
	cardAnalysis, err := analysis.AnalyzeCardCollection(player, analysisOptions)
	if err != nil {
		return fmt.Errorf("failed to analyze card collection: %w", err)
	}

	// Create deck builder
	builder := deck.NewBuilder(dataDir)

	// Override unlocked evolutions if CLI flag provided
	if unlockedEvos := cmd.String("unlocked-evolutions"); unlockedEvos != "" {
		builder.SetUnlockedEvolutions(strings.Split(unlockedEvos, ","))
	}

	// Override evolution slot limit if provided
	if slots := cmd.Int("evolution-slots"); slots > 0 {
		builder.SetEvolutionSlotLimit(slots)
	}

	// Convert analysis.CardAnalysis to deck.CardAnalysis
	deckCardAnalysis := deck.CardAnalysis{
		CardLevels:   make(map[string]deck.CardLevelData),
		AnalysisTime: cardAnalysis.AnalysisTime.Format(time.RFC3339),
	}

	excludeMap := make(map[string]bool)
	for _, card := range excludeCards {
		trimmed := strings.TrimSpace(card)
		if trimmed != "" {
			excludeMap[strings.ToLower(trimmed)] = true
		}
	}

	for cardName, cardInfo := range cardAnalysis.CardLevels {
		if excludeMap[strings.ToLower(cardName)] {
			continue
		}
		deckCardAnalysis.CardLevels[cardName] = deck.CardLevelData{
			Level:             cardInfo.Level,
			MaxLevel:          cardInfo.MaxLevel,
			Rarity:            cardInfo.Rarity,
			Elixir:            cardInfo.Elixir,
			MaxEvolutionLevel: cardInfo.MaxEvolutionLevel,
		}
	}

	// Build deck from analysis
	deckRec, err := builder.BuildDeckFromAnalysis(deckCardAnalysis)
	if err != nil {
		return fmt.Errorf("failed to build deck: %w", err)
	}

	// Validate elixir constraints
	if deckRec.AvgElixir < minElixir || deckRec.AvgElixir > maxElixir {
		fmt.Printf("\n⚠ Warning: Deck average elixir (%.2f) is outside requested range (%.1f-%.1f)\n",
			deckRec.AvgElixir, minElixir, maxElixir)
	}

	// Display deck recommendation
	displayDeckRecommendation(deckRec, player)

	// Save deck if requested
	if saveData {
		if verbose {
			fmt.Printf("\nSaving deck to: %s\n", dataDir)
		}
		deckPath, err := builder.SaveDeck(deckRec, "", player.Tag)
		if err != nil {
			fmt.Printf("Warning: Failed to save deck: %v\n", err)
		} else {
			fmt.Printf("\nDeck saved to: %s\n", deckPath)
		}
	}

	return nil
}

func deckAnalyzeCommand(ctx context.Context, cmd *cli.Command) error {
	cardNames := cmd.StringSlice("cards")

	if len(cardNames) != 8 {
		return fmt.Errorf("exactly 8 cards are required for deck analysis")
	}

	fmt.Printf("Analyzing deck with cards: %v\n", cardNames)
	fmt.Println("Note: Full deck analysis not yet implemented")

	return nil
}

func deckOptimizeCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	cardNames := cmd.StringSlice("cards")
	apiToken := cmd.String("api-token")

	if apiToken == "" {
		return fmt.Errorf("API token is required")
	}

	if len(cardNames) != 8 {
		return fmt.Errorf("exactly 8 cards are required for optimization")
	}

	fmt.Printf("Optimizing deck for player %s\n", tag)
	fmt.Printf("Current deck: %v\n", cardNames)
	fmt.Println("Note: Deck optimization not yet implemented")

	return nil
}

func deckRecommendCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	limit := cmd.Int("limit")
	apiToken := cmd.String("api-token")

	if apiToken == "" {
		return fmt.Errorf("API token is required")
	}

	fmt.Printf("Getting deck recommendations for player %s\n", tag)
	fmt.Printf("Limit: %d recommendations\n", limit)
	fmt.Println("Note: Deck recommendations not yet implemented")

	return nil
}

func deckMulliganCommand(ctx context.Context, cmd *cli.Command) error {
	cardNames := cmd.StringSlice("cards")
	deckName := cmd.String("deck-name")
	saveData := cmd.Bool("save")
	jsonOutput := cmd.Bool("json")
	dataDir := cmd.String("data-dir")
	verbose := cmd.Bool("verbose")

	if len(cardNames) != 8 {
		return fmt.Errorf("exactly 8 cards are required for mulligan analysis")
	}

	// Generate default deck name if not provided
	if deckName == "" {
		deckName = "Custom Deck"
	}

	if verbose {
		fmt.Printf("Generating mulligan guide for deck: %s\n", deckName)
		fmt.Printf("Cards: %v\n", cardNames)
	}

	// Create mulligan generator
	generator := mulligan.NewGenerator()

	// Generate the mulligan guide
	guide, err := generator.GenerateGuide(cardNames, deckName)
	if err != nil {
		return fmt.Errorf("failed to generate mulligan guide: %w", err)
	}

	// Output the guide
	if jsonOutput {
		return outputMulliganGuideJSON(guide)
	} else {
		displayMulliganGuide(guide)
	}

	// Save guide if requested
	if saveData {
		if verbose {
			fmt.Printf("\nSaving mulligan guide to: %s\n", dataDir)
		}
		if err := saveMulliganGuide(dataDir, guide); err != nil {
			fmt.Printf("Warning: Failed to save mulligan guide: %v\n", err)
		} else {
			fmt.Printf("\nMulligan guide saved to file\n")
		}
	}

	return nil
}

// displayMulliganGuide displays a formatted mulligan guide
func displayMulliganGuide(guide *mulligan.MulliganGuide) {
	fmt.Printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                    MULLIGAN GUIDE - OPENING PLAYS               ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Deck: %s (%s)\n", guide.DeckName, guide.Archetype.String())
	fmt.Printf("Generated: %s\n\n", guide.GeneratedAt.Format("2006-01-02 15:04:05"))

	fmt.Printf("📋 General Principles:\n")
	for _, principle := range guide.GeneralPrinciples {
		fmt.Printf("   • %s\n", principle)
	}
	fmt.Println()

	fmt.Printf("🃏 Deck Composition:\n")
	fmt.Printf("   Cards: %s\n", strings.Join(guide.DeckCards, ", "))
	fmt.Println()

	if len(guide.IdealOpenings) > 0 {
		fmt.Printf("✅ Ideal Opening Cards:\n")
		for _, opening := range guide.IdealOpenings {
			fmt.Printf("   ✓ %s\n", opening)
		}
		fmt.Println()
	}

	if len(guide.NeverOpenWith) > 0 {
		fmt.Printf("❌ Never Open With:\n")
		for _, never := range guide.NeverOpenWith {
			fmt.Printf("   ✗ %s\n", never)
		}
		fmt.Println()
	}

	fmt.Printf("🎮 Matchup-Specific Openings:\n\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for i, matchup := range guide.Matchups {
		fmt.Printf("%d. VS %s\n", i+1, matchup.OpponentType)
		fmt.Printf("   ▶ Opening Play: %s\n", matchup.OpeningPlay)
		fmt.Printf("   ▶ Why: %s\n", matchup.Reason)
		fmt.Printf("   ▶ Backup: %s\n", matchup.Backup)
		fmt.Printf("   ▶ Key Cards: %s\n", strings.Join(matchup.KeyCards, ", "))
		fmt.Printf("   ▶ Danger Level: %s\n", matchup.DangerLevel)
		fmt.Println()
	}
	w.Flush()
}

// outputMulliganGuideJSON outputs the guide in JSON format
func outputMulliganGuideJSON(guide *mulligan.MulliganGuide) error {
	data, err := json.MarshalIndent(guide, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mulligan guide: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// saveMulliganGuide saves the mulligan guide to a JSON file
func saveMulliganGuide(dataDir string, guide *mulligan.MulliganGuide) error {
	// Create mulligan directory if it doesn't exist
	mulliganDir := filepath.Join(dataDir, "mulligan")
	if err := os.MkdirAll(mulliganDir, 0755); err != nil {
		return fmt.Errorf("failed to create mulligan directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := guide.GeneratedAt.Format("20060102_150405")
	filename := filepath.Join(mulliganDir, fmt.Sprintf("%s_%s.json", guide.DeckName, timestamp))

	// Save as JSON
	data, err := json.MarshalIndent(guide, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mulligan guide: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write mulligan guide file: %w", err)
	}

	return nil
}

// displayDeckRecommendation displays a formatted deck recommendation
func displayDeckRecommendation(rec *deck.DeckRecommendation, player *clashroyale.Player) {
	fmt.Printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              RECOMMENDED 1v1 LADDER DECK                           ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Player: %s (%s)\n", player.Name, player.Tag)
	fmt.Printf("Average Elixir: %.2f\n", rec.AvgElixir)

	// Display combat stats information if available
	if combatWeight := os.Getenv("COMBAT_STATS_WEIGHT"); combatWeight != "" {
		if combatWeight == "0" {
			fmt.Printf("Scoring: Traditional only (combat stats disabled)\n")
		} else {
			fmt.Printf("Scoring: %.0f%% traditional, %.0f%% combat stats\n",
				(1-mustParseFloat(combatWeight))*100,
				mustParseFloat(combatWeight)*100)
		}
	}
	fmt.Printf("\n")

	// Display deck cards in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "#\tCard\tLevel\tElixir\tRole\n")
	fmt.Fprintf(w, "─\t────\t─────\t──────\t────\n")

	for i, card := range rec.DeckDetail {
		fmt.Fprintf(w, "%d\t%s\t%d/%d\t%d\t%s\n",
			i+1,
			card.Name,
			card.Level,
			card.MaxLevel,
			card.Elixir,
			card.Role)
	}
	w.Flush()

	// Display strategic notes
	if len(rec.Notes) > 0 {
		fmt.Printf("\nStrategic Notes:\n")
		fmt.Printf("════════════════\n")
		for _, note := range rec.Notes {
			fmt.Printf("• %s\n", note)
		}
	}
}

// mustParseFloat parses a float from a string, returning 0 if parsing fails
func mustParseFloat(s string) float64 {
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}
	return 0
}

// Helper functions for deck operations
func saveDeck(dataDir, playerTag string, options map[string]interface{}) error {
	decksDir := filepath.Join(dataDir, "decks")
	if err := os.MkdirAll(decksDir, 0755); err != nil {
		return fmt.Errorf("failed to create decks directory: %w", err)
	}

	// In a real implementation, you'd marshal options to JSON
	return nil
}
