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
			playerTagFlag(true),
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

	// Load card levels and player info
	cardLevels, playerName, err := loadCardLevelsForWhatIf(ctx, fromAnalysis, tag, apiToken, verbose)
	if err != nil {
		return err
	}

	// Parse upgrade specifications
	upgrades, err := parseUpgradeSpecs(upgradesSpec, verbose)
	if err != nil {
		return err
	}

	// Run what-if analysis
	scenario, err := runWhatIfAnalysis(cardLevels, upgrades, strategy, dataDir, playerName, tag)
	if err != nil {
		return err
	}

	// Output results
	if err := outputWhatIfResults(scenario, jsonOutput, showDecks, saveData, dataDir, tag); err != nil {
		return err
	}

	return nil
}

// loadCardLevelsForWhatIf loads card level data from file or API
func loadCardLevelsForWhatIf(ctx context.Context, fromAnalysis, tag, apiToken string, verbose bool) (map[string]deck.CardLevelData, string, error) {
	if fromAnalysis != "" {
		return loadCardLevelsFromFile(fromAnalysis, verbose)
	}
	return loadCardLevelsFromAPI(ctx, tag, apiToken, verbose)
}

// loadCardLevelsFromFile loads card levels from an analysis file
func loadCardLevelsFromFile(filePath string, verbose bool) (map[string]deck.CardLevelData, string, error) {
	if verbose {
		printf("Loading analysis from: %s\n", filePath)
	}
	cardAnalysis, err := loadCardAnalysis(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load analysis file: %w", err)
	}
	cardLevels := convertCardAnalysisToCardLevels(cardAnalysis)
	return cardLevels, cardAnalysis.PlayerName, nil
}

// loadCardLevelsFromAPI fetches card levels from the Clash Royale API
func loadCardLevelsFromAPI(ctx context.Context, tag, apiToken string, verbose bool) (map[string]deck.CardLevelData, string, error) {
	client, err := requireAPIClientFromToken(apiToken, apiClientOptions{
		offlineHint: ", or provide --from-analysis",
	})
	if err != nil {
		return nil, "", err
	}

	if verbose {
		printf("Fetching player data for tag: %s\n", tag)
	}

	player, err := client.GetPlayerWithContext(ctx, tag)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get player: %w", err)
	}

	cardLevels := convertPlayerToCardLevels(player)
	return cardLevels, player.Name, nil
}

// parseUpgradeSpecs parses upgrade specifications from command-line arguments
func parseUpgradeSpecs(upgradesSpec []string, verbose bool) ([]whatif.CardUpgrade, error) {
	upgrades := make([]whatif.CardUpgrade, 0, len(upgradesSpec))
	for _, spec := range upgradesSpec {
		upgrade, err := whatif.ParseCardUpgrade(spec)
		if err != nil {
			return nil, fmt.Errorf("failed to parse upgrade spec '%s': %w", spec, err)
		}
		upgrades = append(upgrades, upgrade)
	}

	if verbose {
		printf("Analyzing %d upgrade scenario(s)...\n", len(upgrades))
		for _, u := range upgrades {
			printf("  - %s: Lv%d -> Lv%d\n", u.CardName, u.FromLevel, u.ToLevel)
		}
	}

	return upgrades, nil
}

// runWhatIfAnalysis executes the what-if analysis with the given parameters
func runWhatIfAnalysis(cardLevels map[string]deck.CardLevelData, upgrades []whatif.CardUpgrade, strategy, dataDir, playerName, tag string) (*whatif.WhatIfScenario, error) {
	builder := deck.NewBuilder(dataDir)
	if strategy != "" {
		if err := builder.SetStrategy(deck.Strategy(strategy)); err != nil {
			return nil, fmt.Errorf("invalid strategy '%s': %w", strategy, err)
		}
	}

	analyzer := whatif.NewWhatIfAnalyzer(builder)
	scenario, err := analyzer.AnalyzeUpgradePath(cardLevels, upgrades)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze upgrade path: %w", err)
	}

	if playerName != "" {
		scenario.Description = fmt.Sprintf("What-if analysis for %s (%s)", playerName, tag)
	}

	return scenario, nil
}

// outputWhatIfResults handles output formatting and optional saving
func outputWhatIfResults(scenario *whatif.WhatIfScenario, jsonOutput, showDecks, saveData bool, dataDir, tag string) error {
	if jsonOutput {
		return outputWhatIfJSON(scenario)
	}

	displayWhatIfScenario(scenario, showDecks)

	if saveData {
		if err := saveWhatIfScenario(dataDir, scenario); err != nil {
			printf("Warning: Failed to save scenario: %v\n", err)
		} else {
			printf("\nScenario saved to: %s/whatif/%s_%s.json\n", dataDir, tag, time.Now().Format("20060102_150405"))
		}
	}

	return nil
}

func displayWhatIfScenario(scenario *whatif.WhatIfScenario, showDecks bool) {
	printf("\n")
	printf("============================================================================\n")
	printf("                        WHAT-IF ANALYSIS                                    \n")
	printf("============================================================================\n\n")

	printf("Scenario: %s\n", scenario.Name)
	if scenario.Description != "" {
		printf("%s\n", scenario.Description)
	}
	printf("\n")

	// Upgrades section
	printf("Upgrades Simulated\n")
	printf("-------------------\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "Card\tFrom\tTo\tGold\n")
	fprintf(w, "----\t----\t--\t----\n")
	for _, u := range scenario.Upgrades {
		fprintf(w, "%s\t%d\t%d\t%d\n", u.CardName, u.FromLevel, u.ToLevel, u.GoldCost)
	}
	flushWriter(w)
	printf("\n")

	printf("Total Gold Cost: %d\n", scenario.TotalGold)
	printf("\n")

	// Impact section
	printf("Impact Analysis\n")
	printf("---------------\n")
	printf("Deck Score Delta:     %+f\n", scenario.Impact.DeckScoreDelta)
	printf("Viability Change:     %+.1f%%\n", scenario.Impact.ViabilityImprovement)

	if len(scenario.Impact.NewCardsInDeck) > 0 {
		printf("New Cards in Deck:     %s\n", formatCardList(scenario.Impact.NewCardsInDeck))
	}
	if len(scenario.Impact.RemovedCards) > 0 {
		printf("Removed from Deck:     %s\n", formatCardList(scenario.Impact.RemovedCards))
	}
	printf("\n")

	// Recommendation
	printf("Recommendation\n")
	printf("-------------\n")
	printf("%s\n", scenario.Impact.Recommendation)
	printf("\n")

	// Show decks if requested
	if showDecks && scenario.OriginalDeck != nil && scenario.SimulatedDeck != nil {
		printf("Deck Comparison\n")
		printf("===============\n")

		printf("\nOriginal Deck (Score: %.3f, Avg Elixir: %.1f)\n",
			calculateDeckScore(scenario.OriginalDeck), scenario.OriginalDeck.AvgElixir)
		displayDeck(scenario.OriginalDeck)

		printf("\nSimulated Deck (Score: %.3f, Avg Elixir: %.1f)\n",
			calculateDeckScore(scenario.SimulatedDeck), scenario.SimulatedDeck.AvgElixir)
		displayDeck(scenario.SimulatedDeck)
	}
}

func displayDeck(deck *deck.DeckRecommendation) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "  Card                 Level       Role        Score     Elixir\n")
	fprintf(w, "  ----                 -----       ----        -----     ------\n")

	for _, card := range deck.DeckDetail {
		levelStr := fmt.Sprintf("%d/%d", card.Level, card.MaxLevel)
		fprintf(w, "  %-20s %-11s %-10s %+.3f    %d\n",
			card.Name, levelStr, card.Role, card.Score, card.Elixir)
	}
	flushWriter(w)

	if len(deck.EvolutionSlots) > 0 {
		printf("  Evolution Slots: %s\n", formatCardList(deck.EvolutionSlots))
	}

	if len(deck.Notes) > 0 {
		printf("\n  Notes:\n")
		for _, note := range deck.Notes {
			printf("  • %s\n", note)
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
	if err := os.MkdirAll(whatifDir, 0o755); err != nil {
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

	if err := os.WriteFile(filename, data, 0o644); err != nil {
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
