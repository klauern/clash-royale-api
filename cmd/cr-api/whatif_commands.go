package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/whatif"
	"github.com/urfave/cli/v3"
)

// addWhatIfCommands adds what-if analysis commands to the CLI
func addWhatIfCommands() *cli.Command {
	return &cli.Command{
		Name:    "what-if",
		Aliases: []string{"wi"},
		Usage:   "Simulate deck changes with upgraded cards ('what-if' analysis)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "tag",
				Aliases:  []string{"p"},
				Usage:    "Player tag (without #)",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "upgrade",
				Aliases:  []string{"u"},
				Usage:    "Card upgrades to simulate (format: CardName:ToLevel or CardName:FromLevel:ToLevel)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "from-analysis",
				Usage: "Path to existing analysis file (optional, skips API fetch if provided)",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output in JSON format",
			},
			&cli.BoolFlag{
				Name:  "save",
				Usage: "Save scenario to file",
			},
			&cli.StringFlag{
				Name:  "strategy",
				Usage: "Deck building strategy (balanced, aggro, control, cycle, splash, spell)",
				Value: "balanced",
			},
			&cli.BoolFlag{
				Name:  "show-decks",
				Usage: "Show both original and simulated deck compositions",
			},
		},
		Action: whatIfCommand,
	}
}

func whatIfCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	upgradesSpec := cmd.StringSlice("upgrade")
	fromAnalysis := cmd.String("from-analysis")
	jsonOutput := cmd.Bool("json")
	saveData := cmd.Bool("save")
	strategy := cmd.String("strategy")
	showDecks := cmd.Bool("show-decks")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	dataDir := cmd.String("data-dir")

	// Get card levels either from file or from API
	var cardLevels map[string]deck.CardLevelData
	var playerName string

	if fromAnalysis != "" {
		// Load from existing analysis file
		if verbose {
			fmt.Printf("Loading analysis from: %s\n", fromAnalysis)
		}
		cardAnalysis, err := loadCardAnalysis(fromAnalysis)
		if err != nil {
			return fmt.Errorf("failed to load analysis file: %w", err)
		}
		cardLevels = convertCardAnalysisToCardLevels(cardAnalysis)
		playerName = cardAnalysis.PlayerName
	} else {
		// Fetch from API
		if apiToken == "" {
			return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag, or provide --from-analysis")
		}

		client := clashroyale.NewClient(apiToken)

		if verbose {
			fmt.Printf("Fetching player data for tag: %s\n", tag)
		}

		player, err := client.GetPlayer(tag)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}

		playerName = player.Name
		cardLevels = convertPlayerToCardLevels(player)
	}

	// Parse upgrade specifications
	upgrades := make([]whatif.CardUpgrade, 0, len(upgradesSpec))
	for _, spec := range upgradesSpec {
		upgrade, err := whatif.ParseCardUpgrade(spec)
		if err != nil {
			return fmt.Errorf("failed to parse upgrade spec '%s': %w", spec, err)
		}
		upgrades = append(upgrades, upgrade)
	}

	if verbose {
		fmt.Printf("Analyzing %d upgrade scenario(s)...\n", len(upgrades))
		for _, u := range upgrades {
			fmt.Printf("  - %s: Lv%d -> Lv%d\n", u.CardName, u.FromLevel, u.ToLevel)
		}
	}

	// Create deck builder with configured strategy
	builder := deck.NewBuilder(dataDir)
	if strategy != "" {
		if err := builder.SetStrategy(deck.Strategy(strategy)); err != nil {
			return fmt.Errorf("invalid strategy '%s': %w", strategy, err)
		}
	}

	// Create what-if analyzer and run analysis
	analyzer := whatif.NewWhatIfAnalyzer(builder)
	scenario, err := analyzer.AnalyzeUpgradePath(cardLevels, upgrades)
	if err != nil {
		return fmt.Errorf("failed to analyze upgrade path: %w", err)
	}

	// Add player info to scenario
	if playerName != "" {
		scenario.Description = fmt.Sprintf("What-if analysis for %s (%s)", playerName, tag)
	}

	// Output results
	if jsonOutput {
		return outputWhatIfJSON(scenario)
	}

	displayWhatIfScenario(scenario, showDecks)

	// Save if requested
	if saveData {
		if err := saveWhatIfScenario(dataDir, scenario); err != nil {
			fmt.Printf("Warning: Failed to save scenario: %v\n", err)
		} else {
			fmt.Printf("\nScenario saved to: %s/whatif/%s_%s.json\n", dataDir, tag, time.Now().Format("20060102_150405"))
		}
	}

	return nil
}

func displayWhatIfScenario(scenario *whatif.WhatIfScenario, showDecks bool) {
	fmt.Printf("\n")
	fmt.Printf("============================================================================\n")
	fmt.Printf("                        WHAT-IF ANALYSIS                                    \n")
	fmt.Printf("============================================================================\n\n")

	fmt.Printf("Scenario: %s\n", scenario.Name)
	if scenario.Description != "" {
		fmt.Printf("%s\n", scenario.Description)
	}
	fmt.Printf("\n")

	// Upgrades section
	fmt.Printf("Upgrades Simulated\n")
	fmt.Printf("-------------------\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Card\tFrom\tTo\tGold\n")
	fmt.Fprintf(w, "----\t----\t--\t----\n")
	for _, u := range scenario.Upgrades {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\n", u.CardName, u.FromLevel, u.ToLevel, u.GoldCost)
	}
	w.Flush()
	fmt.Printf("\n")

	fmt.Printf("Total Gold Cost: %d\n", scenario.TotalGold)
	fmt.Printf("\n")

	// Impact section
	fmt.Printf("Impact Analysis\n")
	fmt.Printf("---------------\n")
	fmt.Printf("Deck Score Delta:     %+f\n", scenario.Impact.DeckScoreDelta)
	fmt.Printf("Viability Change:     %+.1f%%\n", scenario.Impact.ViabilityImprovement)

	if len(scenario.Impact.NewCardsInDeck) > 0 {
		fmt.Printf("New Cards in Deck:     %s\n", formatCardList(scenario.Impact.NewCardsInDeck))
	}
	if len(scenario.Impact.RemovedCards) > 0 {
		fmt.Printf("Removed from Deck:     %s\n", formatCardList(scenario.Impact.RemovedCards))
	}
	fmt.Printf("\n")

	// Recommendation
	fmt.Printf("Recommendation\n")
	fmt.Printf("-------------\n")
	fmt.Printf("%s\n", scenario.Impact.Recommendation)
	fmt.Printf("\n")

	// Show decks if requested
	if showDecks && scenario.OriginalDeck != nil && scenario.SimulatedDeck != nil {
		fmt.Printf("Deck Comparison\n")
		fmt.Printf("===============\n")

		fmt.Printf("\nOriginal Deck (Score: %.3f, Avg Elixir: %.1f)\n",
			calculateDeckScore(scenario.OriginalDeck), scenario.OriginalDeck.AvgElixir)
		displayDeck(scenario.OriginalDeck)

		fmt.Printf("\nSimulated Deck (Score: %.3f, Avg Elixir: %.1f)\n",
			calculateDeckScore(scenario.SimulatedDeck), scenario.SimulatedDeck.AvgElixir)
		displayDeck(scenario.SimulatedDeck)
	}
}

func displayDeck(deck *deck.DeckRecommendation) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Card                 Level       Role        Score     Elixir\n")
	fmt.Fprintf(w, "  ----                 -----       ----        -----     ------\n")

	for _, card := range deck.DeckDetail {
		levelStr := fmt.Sprintf("%d/%d", card.Level, card.MaxLevel)
		fmt.Fprintf(w, "  %-20s %-11s %-10s %+.3f    %d\n",
			card.Name, levelStr, card.Role, card.Score, card.Elixir)
	}
	w.Flush()

	if len(deck.EvolutionSlots) > 0 {
		fmt.Printf("  Evolution Slots: %s\n", formatCardList(deck.EvolutionSlots))
	}

	if len(deck.Notes) > 0 {
		fmt.Printf("\n  Notes:\n")
		for _, note := range deck.Notes {
			fmt.Printf("  • %s\n", note)
		}
	}
}

func calculateDeckScore(deck *deck.DeckRecommendation) float64 {
	if deck == nil || len(deck.DeckDetail) == 0 {
		return 0
	}
	total := 0.0
	for _, card := range deck.DeckDetail {
		total += card.Score
	}
	return total
}

func formatCardList(cards []string) string {
	if len(cards) == 0 {
		return "None"
	}
	if len(cards) == 1 {
		return cards[0]
	}
	if len(cards) <= 3 {
		return fmt.Sprintf("%s, %s",
			fmt.Sprintf("%v", cards[:len(cards)-1]),
			cards[len(cards)-1])
	}
	return fmt.Sprintf("%s, and %d more",
		fmt.Sprintf("%v", cards[:2]),
		len(cards)-2)
}

func outputWhatIfJSON(scenario *whatif.WhatIfScenario) error {
	data, err := json.MarshalIndent(scenario, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scenario: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func saveWhatIfScenario(dataDir string, scenario *whatif.WhatIfScenario) error {
	// Create whatif directory if it doesn't exist
	whatifDir := filepath.Join(dataDir, "whatif")
	if err := os.MkdirAll(whatifDir, 0755); err != nil {
		return fmt.Errorf("failed to create whatif directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(whatifDir, fmt.Sprintf("scenario_%s.json", timestamp))

	// Save as JSON
	data, err := json.MarshalIndent(scenario, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal scenario: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write scenario file: %w", err)
	}

	return nil
}

func loadCardAnalysis(path string) (*analysis.CardAnalysis, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cardAnalysis analysis.CardAnalysis
	if err := json.Unmarshal(data, &cardAnalysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis JSON: %w", err)
	}

	return &cardAnalysis, nil
}

func convertCardAnalysisToCardLevels(cardAnalysis *analysis.CardAnalysis) map[string]deck.CardLevelData {
	cardLevels := make(map[string]deck.CardLevelData)
	for name, info := range cardAnalysis.CardLevels {
		cardLevels[name] = deck.CardLevelData{
			Level:             info.Level,
			MaxLevel:          info.MaxLevel,
			Rarity:            info.Rarity,
			Elixir:            info.Elixir,
			EvolutionLevel:    info.EvolutionLevel,
			MaxEvolutionLevel: info.MaxEvolutionLevel,
			ScoreBoost:        0,
		}
	}
	return cardLevels
}

func convertPlayerToCardLevels(player *clashroyale.Player) map[string]deck.CardLevelData {
	cardLevels := make(map[string]deck.CardLevelData)
	for _, card := range player.Cards {
		cardLevels[card.Name] = deck.CardLevelData{
			Level:             card.Level,
			MaxLevel:          card.MaxLevel,
			Rarity:            card.Rarity,
			Elixir:            card.ElixirCost,
			EvolutionLevel:    card.EvolutionLevel,
			MaxEvolutionLevel: card.MaxEvolutionLevel,
			ScoreBoost:        0,
		}
	}
	return cardLevels
}
