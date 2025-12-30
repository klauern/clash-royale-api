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
	"github.com/klauer/clash-royale-api/go/pkg/budget"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
	"github.com/klauer/clash-royale-api/go/pkg/recommend"
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
					&cli.BoolFlag{
						Name:    "from-analysis",
						Aliases: []string{"a"},
						Usage:   "Enable offline mode: load analysis from JSON file instead of fetching from API",
					},
					&cli.StringFlag{
						Name:  "analysis-dir",
						Usage: "Directory containing analysis JSON files (default: data/analysis)",
					},
					&cli.StringFlag{
						Name:  "analysis-file",
						Usage: "Specific analysis file path (overrides --analysis-dir lookup)",
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
					&cli.BoolFlag{
						Name:    "from-analysis",
						Aliases: []string{"a"},
						Usage:   "Enable offline mode: load analysis from JSON file instead of fetching from API",
					},
					&cli.StringFlag{
						Name:  "analysis-dir",
						Usage: "Directory containing analysis JSON files (default: data/analysis)",
					},
					&cli.StringFlag{
						Name:  "analysis-file",
						Usage: "Specific analysis file path (overrides --analysis-dir lookup)",
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
			{
				Name:  "budget",
				Usage: "Find budget-optimized decks with minimal upgrade investment",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "max-cards",
						Value: 0,
						Usage: "Maximum cards needed for upgrades (0 = no limit)",
					},
					&cli.IntFlag{
						Name:  "max-gold",
						Value: 0,
						Usage: "Maximum gold needed for upgrades (0 = no limit)",
					},
					&cli.Float64Flag{
						Name:  "target-level",
						Value: 12.0,
						Usage: "Target average card level for viability",
					},
					&cli.StringFlag{
						Name:  "sort-by",
						Value: "roi",
						Usage: "Sort results by: roi, cost_efficiency, total_cards, total_gold, current_score, projected_score",
					},
					&cli.IntFlag{
						Name:  "top-n",
						Value: 10,
						Usage: "Number of top decks to display",
					},
					&cli.BoolFlag{
						Name:  "include-variations",
						Usage: "Generate and analyze deck variations",
					},
					&cli.IntFlag{
						Name:  "max-variations",
						Value: 5,
						Usage: "Maximum number of deck variations to generate",
					},
					&cli.BoolFlag{
						Name:  "quick-wins",
						Usage: "Show only quick-win decks (1-2 upgrades away)",
					},
					&cli.BoolFlag{
						Name:  "ready-only",
						Usage: "Show only decks that are already competitive",
					},
					&cli.StringFlag{
						Name:  "unlocked-evolutions",
						Usage: "Comma-separated list of cards with unlocked evolutions",
					},
					&cli.IntFlag{
						Name:  "evolution-slots",
						Value: 2,
						Usage: "Number of evolution slots available",
					},
					&cli.BoolFlag{
						Name:  "json",
						Usage: "Output results in JSON format",
					},
					&cli.BoolFlag{
						Name:  "save",
						Usage: "Save results to file",
					},
				},
				Action: deckBudgetCommand,
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

	// Offline mode flags
	fromAnalysis := cmd.Bool("from-analysis")
	analysisDir := cmd.String("analysis-dir")
	analysisFile := cmd.String("analysis-file")

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

	// Parse and set strategy if provided
	if strategy != "" {
		parsedStrategy, err := deck.ParseStrategy(strategy)
		if err != nil {
			return fmt.Errorf("invalid strategy: %w", err)
		}
		if err := builder.SetStrategy(parsedStrategy); err != nil {
			return fmt.Errorf("failed to set strategy: %w", err)
		}
		if verbose {
			fmt.Printf("Using deck building strategy: %s\n", parsedStrategy)
		}
	}

	var deckCardAnalysis deck.CardAnalysis
	var playerName, playerTag string

	if fromAnalysis {
		// OFFLINE MODE: Load from existing analysis JSON
		if verbose {
			fmt.Printf("Building deck from offline analysis for player %s\n", tag)
		}

		// Default analysis dir to data/analysis if not specified
		if analysisDir == "" {
			analysisDir = filepath.Join(dataDir, "analysis")
		}

		var loadedAnalysis *deck.CardAnalysis
		var err error

		if analysisFile != "" {
			// Load from explicit file path
			loadedAnalysis, err = builder.LoadAnalysis(analysisFile)
			if err != nil {
				return fmt.Errorf("failed to load analysis file %s: %w", analysisFile, err)
			}
			if verbose {
				fmt.Printf("Loaded analysis from: %s\n", analysisFile)
			}
		} else {
			// Load latest analysis for player tag
			loadedAnalysis, err = builder.LoadLatestAnalysis(tag, analysisDir)
			if err != nil {
				return fmt.Errorf("failed to load analysis for player %s from %s: %w", tag, analysisDir, err)
			}
			if verbose {
				fmt.Printf("Loaded latest analysis from: %s\n", analysisDir)
			}
		}

		deckCardAnalysis = *loadedAnalysis
		playerTag = tag
		playerName = tag // Use tag as name in offline mode
	} else {
		// ONLINE MODE: Fetch from API
		if apiToken == "" {
			return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag. Use --from-analysis for offline mode.")
		}

		client := clashroyale.NewClient(apiToken)

		if verbose {
			fmt.Printf("Building deck for player %s\n", tag)
		}

		// Get player information
		player, err := client.GetPlayer(tag)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}

		playerName = player.Name
		playerTag = player.Tag

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

		// Convert analysis.CardAnalysis to deck.CardAnalysis
		deckCardAnalysis = deck.CardAnalysis{
			CardLevels:   make(map[string]deck.CardLevelData),
			AnalysisTime: cardAnalysis.AnalysisTime.Format(time.RFC3339),
		}

		for cardName, cardInfo := range cardAnalysis.CardLevels {
			deckCardAnalysis.CardLevels[cardName] = deck.CardLevelData{
				Level:             cardInfo.Level,
				MaxLevel:          cardInfo.MaxLevel,
				Rarity:            cardInfo.Rarity,
				Elixir:            cardInfo.Elixir,
				MaxEvolutionLevel: cardInfo.MaxEvolutionLevel,
			}
		}
	}

	// Apply exclude filter (works for both modes)
	excludeMap := make(map[string]bool)
	for _, card := range excludeCards {
		trimmed := strings.TrimSpace(card)
		if trimmed != "" {
			excludeMap[strings.ToLower(trimmed)] = true
		}
	}

	if len(excludeMap) > 0 {
		filteredLevels := make(map[string]deck.CardLevelData)
		for cardName, cardInfo := range deckCardAnalysis.CardLevels {
			if !excludeMap[strings.ToLower(cardName)] {
				filteredLevels[cardName] = cardInfo
			}
		}
		deckCardAnalysis.CardLevels = filteredLevels
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
	displayDeckRecommendationOffline(deckRec, playerName, playerTag)

	// Save deck if requested
	if saveData {
		if verbose {
			fmt.Printf("\nSaving deck to: %s\n", dataDir)
		}
		deckPath, err := builder.SaveDeck(deckRec, "", playerTag)
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
	arena := cmd.String("arena")
	league := cmd.String("league")
	exportCSV := cmd.Bool("export-csv")
	apiToken := cmd.String("api-token")
	dataDir := cmd.String("data-dir")
	verbose := cmd.Bool("verbose")

	// Offline mode flags
	fromAnalysis := cmd.Bool("from-analysis")
	analysisDir := cmd.String("analysis-dir")
	analysisFile := cmd.String("analysis-file")

	var deckCardAnalysis deck.CardAnalysis
	var playerName, playerTag string

	if fromAnalysis {
		// OFFLINE MODE: Load from existing analysis JSON
		if verbose {
			fmt.Printf("Generating recommendations from offline analysis for player %s\n", tag)
		}

		// Default analysis dir to data/analysis if not specified
		if analysisDir == "" {
			analysisDir = filepath.Join(dataDir, "analysis")
		}

		builder := deck.NewBuilder(dataDir)
		var loadedAnalysis *deck.CardAnalysis
		var err error

		if analysisFile != "" {
			// Load from explicit file path
			loadedAnalysis, err = builder.LoadAnalysis(analysisFile)
			if err != nil {
				return fmt.Errorf("failed to load analysis file %s: %w", analysisFile, err)
			}
			if verbose {
				fmt.Printf("Loaded analysis from: %s\n", analysisFile)
			}
		} else {
			// Load latest analysis for player tag
			loadedAnalysis, err = builder.LoadLatestAnalysis(tag, analysisDir)
			if err != nil {
				return fmt.Errorf("failed to load analysis for player %s from %s: %w", tag, analysisDir, err)
			}
			if verbose {
				fmt.Printf("Loaded latest analysis from: %s\n", analysisDir)
			}
		}

		deckCardAnalysis = *loadedAnalysis
		playerTag = tag
		playerName = tag // Use tag as name in offline mode
	} else {
		// ONLINE MODE: Fetch from API
		if apiToken == "" {
			return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag. Use --from-analysis for offline mode.")
		}

		client := clashroyale.NewClient(apiToken)

		if verbose {
			fmt.Printf("Generating recommendations for player %s\n", tag)
		}

		// Get player information
		player, err := client.GetPlayer(tag)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}

		playerName = player.Name
		playerTag = player.Tag

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

		// Convert analysis.CardAnalysis to deck.CardAnalysis
		deckCardAnalysis = deck.CardAnalysis{
			CardLevels:   make(map[string]deck.CardLevelData),
			AnalysisTime: cardAnalysis.AnalysisTime.Format(time.RFC3339),
		}

		for cardName, cardInfo := range cardAnalysis.CardLevels {
			deckCardAnalysis.CardLevels[cardName] = deck.CardLevelData{
				Level:             cardInfo.Level,
				MaxLevel:          cardInfo.MaxLevel,
				Rarity:            cardInfo.Rarity,
				Elixir:            cardInfo.Elixir,
				EvolutionLevel:    cardInfo.EvolutionLevel,
				MaxEvolutionLevel: cardInfo.MaxEvolutionLevel,
			}
		}
	}

	// Create recommender with options
	options := recommend.DefaultOptions()
	options.Limit = limit
	options.Arena = arena
	options.League = league

	recommender := recommend.NewRecommender(dataDir, options)

	// Generate recommendations
	result, err := recommender.GenerateRecommendations(playerTag, playerName, deckCardAnalysis)
	if err != nil {
		return fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// Display results
	displayRecommendations(result, verbose)

	// Export to CSV if requested
	if exportCSV {
		if err := exportRecommendationsToCSV(dataDir, result); err != nil {
			return fmt.Errorf("failed to export CSV: %w", err)
		}
		fmt.Printf("\nExported to CSV: %s\n", getRecommendationsCSVPath(dataDir, playerTag))
	}

	return nil
}

// displayRecommendations displays deck recommendations in a formatted table
func displayRecommendations(result *recommend.RecommendationResult, verbose bool) {
	fmt.Printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              DECK RECOMMENDATIONS                                  ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Player: %s (%s)\n", result.PlayerName, result.PlayerTag)
	if result.TopArchetype != "" {
		fmt.Printf("Top Archetype Match: %s\n", result.TopArchetype)
	}
	fmt.Printf("Generated: %s\n", result.GeneratedAt)
	fmt.Printf("\n")

	if len(result.Recommendations) == 0 {
		fmt.Println("No recommendations found. Your card collection may be too limited.")
		return
	}

	for i, rec := range result.Recommendations {
		displaySingleRecommendation(i+1, rec, verbose)
		fmt.Println()
	}
}

// displaySingleRecommendation displays a single deck recommendation
func displaySingleRecommendation(rank int, rec *recommend.DeckRecommendation, verbose bool) {
	typeLabel := "Meta"
	if rec.Type == recommend.TypeCustomVariation {
		typeLabel = "Custom"
	}

	fmt.Printf("Deck #%d [%s - %s]\n", rank, rec.ArchetypeName, typeLabel)
	fmt.Printf("─────────────────────────────────────────────────────\n")
	fmt.Printf("Compatibility: %.1f%% | Synergy: %.1f%% | Overall: %.1f%%\n",
		rec.CompatibilityScore, rec.SynergyScore, rec.OverallScore)
	fmt.Printf("Avg Elixir: %.2f\n", rec.Deck.AvgElixir)

	// Display cards in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Card\t\tLevel\tElixir\n")
	for _, card := range rec.Deck.DeckDetail {
		fmt.Fprintf(w, "%s\t\t%d\t%d\n", card.Name, card.Level, card.Elixir)
	}
	w.Flush()

	// Display evolution slots if any
	if len(rec.Deck.EvolutionSlots) > 0 {
		fmt.Printf("Evolution Slots: %s\n", strings.Join(rec.Deck.EvolutionSlots, ", "))
	}

	// Display reasons
	if len(rec.Reasons) > 0 {
		fmt.Printf("\nWhy Recommended:\n")
		for _, reason := range rec.Reasons {
			fmt.Printf("  • %s\n", reason)
		}
	}

	if verbose {
		// Display upgrade cost if available
		if rec.UpgradeCost.CardsNeeded > 0 {
			fmt.Printf("\nUpgrade needed: %d cards, %d gold\n",
				rec.UpgradeCost.CardsNeeded, rec.UpgradeCost.GoldNeeded)
		}

		// Display notes
		if len(rec.Deck.Notes) > 0 {
			fmt.Printf("\nNotes:\n")
			for _, note := range rec.Deck.Notes {
				fmt.Printf("  • %s\n", note)
			}
		}
	}
}

// exportRecommendationsToCSV exports recommendations to CSV file
func exportRecommendationsToCSV(dataDir string, result *recommend.RecommendationResult) error {
	csvDir := filepath.Join(dataDir, "csv")
	if err := os.MkdirAll(csvDir, 0755); err != nil {
		return fmt.Errorf("failed to create CSV directory: %w", err)
	}

	csvPath := getRecommendationsCSVPath(dataDir, result.PlayerTag)

	file, err := os.Create(csvPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	// Write header
	header := []string{
		"Rank", "Archetype", "Type", "Compatibility", "Synergy", "Overall",
		"AvgElixir", "Cards", "EvolutionSlots", "Reasons",
	}
	if _, err := file.WriteString(strings.Join(header, ",") + "\n"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write rows
	for i, rec := range result.Recommendations {
		cardsStr := strings.Join(rec.Deck.Deck, ";")
		evoSlotsStr := strings.Join(rec.Deck.EvolutionSlots, ";")
		reasonsStr := strings.Join(rec.Reasons, "; ")

		row := []string{
			strconv.Itoa(i + 1),
			rec.ArchetypeName,
			string(rec.Type),
			fmt.Sprintf("%.1f", rec.CompatibilityScore),
			fmt.Sprintf("%.1f", rec.SynergyScore),
			fmt.Sprintf("%.1f", rec.OverallScore),
			fmt.Sprintf("%.2f", rec.Deck.AvgElixir),
			cardsStr,
			evoSlotsStr,
			reasonsStr,
		}
		if _, err := file.WriteString(strings.Join(row, ",") + "\n"); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// getRecommendationsCSVPath returns the CSV file path for recommendations
func getRecommendationsCSVPath(dataDir, playerTag string) string {
	return filepath.Join(dataDir, "csv", fmt.Sprintf("recommendations_%s.csv", playerTag))
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

// displayDeckRecommendationOffline displays a formatted deck recommendation without full player object
// Used for offline mode where we only have player name and tag as strings
func displayDeckRecommendationOffline(rec *deck.DeckRecommendation, playerName, playerTag string) {
	fmt.Printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              RECOMMENDED 1v1 LADDER DECK                           ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Player: %s (%s)\n", playerName, playerTag)
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
	fmt.Fprintf(w, "#\tCard\tLevel\t\tElixir\tRole\n")
	fmt.Fprintf(w, "─\t────\t─────\t\t──────\t────\n")

	for i, card := range rec.DeckDetail {
		evoBadge := deck.FormatEvolutionBadge(card.EvolutionLevel)
		levelStr := fmt.Sprintf("%d/%d", card.Level, card.MaxLevel)
		if evoBadge != "" {
			levelStr = fmt.Sprintf("%s (%s)", levelStr, evoBadge)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%s\n",
			i+1,
			card.Name,
			levelStr,
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

func deckBudgetCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	maxCards := cmd.Int("max-cards")
	maxGold := cmd.Int("max-gold")
	targetLevel := cmd.Float64("target-level")
	sortBy := cmd.String("sort-by")
	topN := cmd.Int("top-n")
	includeVariations := cmd.Bool("include-variations")
	maxVariations := cmd.Int("max-variations")
	quickWinsOnly := cmd.Bool("quick-wins")
	readyOnly := cmd.Bool("ready-only")
	jsonOutput := cmd.Bool("json")
	saveData := cmd.Bool("save")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	dataDir := cmd.String("data-dir")

	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	client := clashroyale.NewClient(apiToken)

	if verbose {
		fmt.Printf("Finding budget-optimized decks for player %s\n", tag)
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

	// Create budget finder options
	options := budget.BudgetFinderOptions{
		MaxCardsNeeded:      maxCards,
		MaxGoldNeeded:       maxGold,
		TargetAverageLevel:  targetLevel,
		QuickWinMaxUpgrades: 2,
		QuickWinMaxCards:    1000,
		SortBy:              parseSortCriteria(sortBy),
		TopN:                topN,
		IncludeVariations:   includeVariations,
		MaxVariations:       maxVariations,
	}

	// Create budget finder
	finder := budget.NewFinder(dataDir, options)

	// Override unlocked evolutions if CLI flag provided
	if unlockedEvos := cmd.String("unlocked-evolutions"); unlockedEvos != "" {
		finder.SetUnlockedEvolutions(strings.Split(unlockedEvos, ","))
	}

	// Override evolution slot limit if provided
	if slots := cmd.Int("evolution-slots"); slots > 0 {
		finder.SetEvolutionSlotLimit(slots)
	}

	// Convert analysis.CardAnalysis to deck.CardAnalysis
	deckCardAnalysis := deck.CardAnalysis{
		CardLevels:   make(map[string]deck.CardLevelData),
		AnalysisTime: cardAnalysis.AnalysisTime.Format(time.RFC3339),
	}

	for cardName, cardInfo := range cardAnalysis.CardLevels {
		deckCardAnalysis.CardLevels[cardName] = deck.CardLevelData{
			Level:             cardInfo.Level,
			MaxLevel:          cardInfo.MaxLevel,
			Rarity:            cardInfo.Rarity,
			Elixir:            cardInfo.Elixir,
			MaxEvolutionLevel: cardInfo.MaxEvolutionLevel,
		}
	}

	// Find optimal decks
	result, err := finder.FindOptimalDecks(deckCardAnalysis, player.Tag, player.Name)
	if err != nil {
		return fmt.Errorf("failed to find optimal decks: %w", err)
	}

	// Filter results if requested
	if quickWinsOnly {
		result.AllDecks = result.QuickWins
	} else if readyOnly {
		result.AllDecks = result.ReadyDecks
	}

	// Output results
	if jsonOutput {
		return outputBudgetResultJSON(result)
	}

	displayBudgetResult(result, player, options)

	// Save results if requested
	if saveData {
		if verbose {
			fmt.Printf("\nSaving budget analysis to: %s\n", dataDir)
		}
		if err := saveBudgetResult(dataDir, result); err != nil {
			fmt.Printf("Warning: Failed to save budget analysis: %v\n", err)
		} else {
			fmt.Printf("\nBudget analysis saved to file\n")
		}
	}

	return nil
}

// parseSortCriteria converts string to SortCriteria
func parseSortCriteria(s string) budget.SortCriteria {
	switch strings.ToLower(s) {
	case "roi":
		return budget.SortByROI
	case "cost_efficiency":
		return budget.SortByCostEfficiency
	case "total_cards":
		return budget.SortByTotalCards
	case "total_gold":
		return budget.SortByTotalGold
	case "current_score":
		return budget.SortByCurrentScore
	case "projected_score":
		return budget.SortByProjectedScore
	default:
		return budget.SortByROI
	}
}

// displayBudgetResult displays budget analysis results in a formatted way
func displayBudgetResult(result *budget.BudgetFinderResult, player *clashroyale.Player, options budget.BudgetFinderOptions) {
	fmt.Printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              BUDGET-OPTIMIZED DECK FINDER                          ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Player: %s (%s)\n", result.PlayerName, result.PlayerTag)
	fmt.Printf("Average Card Level: %.2f\n\n", result.Summary.PlayerAverageLevel)

	// Display summary
	fmt.Printf("Summary:\n")
	fmt.Printf("════════\n")
	fmt.Printf("Total Decks Analyzed:    %d\n", result.Summary.TotalDecksAnalyzed)
	fmt.Printf("Ready Decks:             %d\n", result.Summary.ReadyDeckCount)
	fmt.Printf("Quick Win Decks:         %d\n", result.Summary.QuickWinCount)
	fmt.Printf("Best ROI:                %.4f\n", result.Summary.BestROI)
	fmt.Printf("Lowest Cards Needed:     %d\n", result.Summary.LowestCardsNeeded)
	fmt.Printf("\n")

	// Display quick wins if available
	if len(result.QuickWins) > 0 {
		fmt.Printf("Quick Wins (1-2 upgrades away):\n")
		fmt.Printf("════════════════════════════════\n")
		for i, analysis := range result.QuickWins {
			if i >= 3 {
				break // Show top 3 quick wins
			}
			displayBudgetDeckSummary(i+1, analysis)
		}
		fmt.Printf("\n")
	}

	// Display all decks
	if len(result.AllDecks) > 0 {
		fmt.Printf("Top Decks (sorted by %s):\n", options.SortBy)
		fmt.Printf("═════════════════════════════════════\n\n")

		for i, analysis := range result.AllDecks {
			displayBudgetDeckDetail(i+1, analysis)
		}
	} else {
		fmt.Printf("No decks found matching criteria.\n")
	}
}

// displayBudgetDeckSummary displays a brief deck summary
func displayBudgetDeckSummary(rank int, analysis *budget.DeckBudgetAnalysis) {
	if analysis.Deck == nil {
		return
	}

	cards := make([]string, 0, len(analysis.Deck.DeckDetail))
	for _, card := range analysis.Deck.DeckDetail {
		cards = append(cards, card.Name)
	}

	fmt.Printf("#%d: %s\n", rank, strings.Join(cards[:min(3, len(cards))], ", ")+"...")
	fmt.Printf("    Cards Needed: %d | Gold: %d | ROI: %.4f\n",
		analysis.TotalCardsNeeded, analysis.TotalGoldNeeded, analysis.ROI)
}

// displayBudgetDeckDetail displays detailed deck information
func displayBudgetDeckDetail(rank int, analysis *budget.DeckBudgetAnalysis) {
	if analysis.Deck == nil {
		return
	}

	fmt.Printf("Deck #%d [%s]\n", rank, analysis.BudgetCategory)
	fmt.Printf("─────────────────────────────────────\n")

	// Deck cards table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Card\tLevel\t\tElixir\tRole\n")
	fmt.Fprintf(w, "────\t─────\t\t──────\t────\n")

	for _, card := range analysis.Deck.DeckDetail {
		evoBadge := deck.FormatEvolutionBadge(card.EvolutionLevel)
		levelStr := fmt.Sprintf("%d/%d", card.Level, card.MaxLevel)
		if evoBadge != "" {
			levelStr = fmt.Sprintf("%s (%s)", levelStr, evoBadge)
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
			card.Name,
			levelStr,
			card.Elixir,
			card.Role)
	}
	w.Flush()

	fmt.Printf("\n")
	fmt.Printf("Average Elixir: %.2f\n", analysis.Deck.AvgElixir)
	fmt.Printf("Current Score: %.4f | Projected Score: %.4f\n",
		analysis.CurrentScore, analysis.ProjectedScore)
	fmt.Printf("Cards Needed: %d | Gold Needed: %d\n",
		analysis.TotalCardsNeeded, analysis.TotalGoldNeeded)
	fmt.Printf("ROI: %.4f | Cost Efficiency: %.4f\n",
		analysis.ROI, analysis.CostEfficiency)

	// Display upgrade priorities if there are upgrades needed
	if len(analysis.CardUpgrades) > 0 {
		fmt.Printf("\nUpgrade Priorities:\n")
		for i, upgrade := range analysis.CardUpgrades {
			if i >= 3 {
				fmt.Printf("  ... and %d more\n", len(analysis.CardUpgrades)-3)
				break
			}
			fmt.Printf("  %d. %s: Level %d -> %d (%d cards, %d gold)\n",
				i+1, upgrade.CardName, upgrade.CurrentLevel, upgrade.TargetLevel,
				upgrade.CardsNeeded, upgrade.GoldNeeded)
		}
	}

	fmt.Printf("\n")
}

// outputBudgetResultJSON outputs budget analysis in JSON format
func outputBudgetResultJSON(result *budget.BudgetFinderResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal budget result: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// saveBudgetResult saves budget analysis to a JSON file
func saveBudgetResult(dataDir string, result *budget.BudgetFinderResult) error {
	// Create budget directory if it doesn't exist
	budgetDir := filepath.Join(dataDir, "budget")
	if err := os.MkdirAll(budgetDir, 0755); err != nil {
		return fmt.Errorf("failed to create budget directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	cleanTag := strings.TrimPrefix(result.PlayerTag, "#")
	filename := filepath.Join(budgetDir, fmt.Sprintf("%s_budget_%s.json", timestamp, cleanTag))

	// Save as JSON
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal budget result: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write budget file: %w", err)
	}

	fmt.Printf("Budget analysis saved to: %s\n", filename)
	return nil
}
