package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/budget"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/urfave/cli/v3"
)

//nolint:gocyclo,funlen // CLI orchestration path retained pending modular decomposition in clash-royale-api-sb3q.
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
		printf("Finding budget-optimized decks for player %s\n", tag)
	}

	// Get player information
	player, err := client.GetPlayer(tag)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	if verbose {
		printf("Player: %s (%s)\n", player.Name, player.Tag)
		printf("Analyzing %d cards...\n", len(player.Cards))
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
	deckCardAnalysis := convertToDeckCardAnalysis(cardAnalysis, player)

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
			printf("\nSaving budget analysis to: %s\n", dataDir)
		}
		if err := saveBudgetResult(dataDir, result); err != nil {
			printf("Warning: Failed to save budget analysis: %v\n", err)
		} else {
			printf("\nBudget analysis saved to file\n")
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
	printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	printf("║              BUDGET-OPTIMIZED DECK FINDER                          ║\n")
	printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	printf("Player: %s (%s)\n", result.PlayerName, result.PlayerTag)
	printf("Average Card Level: %.2f\n\n", result.Summary.PlayerAverageLevel)

	// Display summary
	printf("Summary:\n")
	printf("════════\n")
	printf("Total Decks Analyzed:    %d\n", result.Summary.TotalDecksAnalyzed)
	printf("Ready Decks:             %d\n", result.Summary.ReadyDeckCount)
	printf("Quick Win Decks:         %d\n", result.Summary.QuickWinCount)
	printf("Best ROI:                %.4f\n", result.Summary.BestROI)
	printf("Lowest Cards Needed:     %d\n", result.Summary.LowestCardsNeeded)
	printf("\n")

	// Display quick wins if available
	if len(result.QuickWins) > 0 {
		printf("Quick Wins (1-2 upgrades away):\n")
		printf("════════════════════════════════\n")
		for i, analysis := range result.QuickWins {
			if i >= 3 {
				break // Show top 3 quick wins
			}
			displayBudgetDeckSummary(i+1, analysis)
		}
		printf("\n")
	}

	// Display all decks
	if len(result.AllDecks) > 0 {
		printf("Top Decks (sorted by %s):\n", options.SortBy)
		printf("═════════════════════════════════════\n\n")

		for i, analysis := range result.AllDecks {
			displayBudgetDeckDetail(i+1, analysis)
		}
	} else {
		printf("No decks found matching criteria.\n")
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

	printf("#%d: %s\n", rank, strings.Join(cards[:min(3, len(cards))], ", ")+"...")
	printf("    Cards Needed: %d | Gold: %d | ROI: %.4f\n",
		analysis.TotalCardsNeeded, analysis.TotalGoldNeeded, analysis.ROI)
}

// displayBudgetDeckDetail displays detailed deck information
func displayBudgetDeckDetail(rank int, analysis *budget.DeckBudgetAnalysis) {
	if analysis.Deck == nil {
		return
	}

	printf("Deck #%d [%s]\n", rank, analysis.BudgetCategory)
	printf("─────────────────────────────────────\n")

	// Deck cards table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "Card\tLevel\t\tElixir\tRole\n")
	fprintf(w, "────\t─────\t\t──────\t────\n")

	for _, card := range analysis.Deck.DeckDetail {
		evoBadge := deck.FormatEvolutionBadge(card.EvolutionLevel)
		levelStr := fmt.Sprintf("%d/%d", card.Level, card.MaxLevel)
		if evoBadge != "" {
			levelStr = fmt.Sprintf("%s (%s)", levelStr, evoBadge)
		}
		fprintf(w, "%s\t%s\t%d\t%s\n",
			card.Name,
			levelStr,
			card.Elixir,
			card.Role)
	}
	flushWriter(w)

	printf("\n")
	printf("Average Elixir: %.2f\n", analysis.Deck.AvgElixir)
	printf("Current Score: %.4f | Projected Score: %.4f\n",
		analysis.CurrentScore, analysis.ProjectedScore)
	printf("Cards Needed: %d | Gold Needed: %d\n",
		analysis.TotalCardsNeeded, analysis.TotalGoldNeeded)
	printf("ROI: %.4f | Cost Efficiency: %.4f\n",
		analysis.ROI, analysis.CostEfficiency)

	// Display upgrade priorities if there are upgrades needed
	if len(analysis.CardUpgrades) > 0 {
		printf("\nUpgrade Priorities:\n")
		for i, upgrade := range analysis.CardUpgrades {
			if i >= 3 {
				printf("  ... and %d more\n", len(analysis.CardUpgrades)-3)
				break
			}
			printf("  %d. %s: Level %d -> %d (%d cards, %d gold)\n",
				i+1, upgrade.CardName, upgrade.CurrentLevel, upgrade.TargetLevel,
				upgrade.CardsNeeded, upgrade.GoldNeeded)
		}
	}

	printf("\n")
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
	if err := os.MkdirAll(budgetDir, 0o755); err != nil {
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

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("failed to write budget file: %w", err)
	}

	printf("Budget analysis saved to: %s\n", filename)
	return nil
}
