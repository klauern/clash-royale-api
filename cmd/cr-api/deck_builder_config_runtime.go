package main

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/fuzzstorage"
	"github.com/urfave/cli/v3"
)

func configureCombatStats(cmd *cli.Command) error {
	combatStatsWeight := cmd.Float64("combat-stats-weight")
	disableCombatStats := cmd.Bool("disable-combat-stats")
	verbose := cmd.Bool("verbose")

	if disableCombatStats {
		setEnv("COMBAT_STATS_WEIGHT", "0")
		if verbose {
			printf("Combat stats disabled (using traditional scoring only)\n")
		}
	} else if cmd.IsSet("combat-stats-weight") && combatStatsWeight >= 0 && combatStatsWeight <= 1.0 {
		setEnv("COMBAT_STATS_WEIGHT", fmt.Sprintf("%.2f", combatStatsWeight))
		if verbose {
			printf("Combat stats weight set to: %.2f\n", combatStatsWeight)
		}
	}
	return nil
}

// configureDeckBuilder sets up the deck builder with evolutions, filters, strategy, and synergy
func configureDeckBuilder(cmd *cli.Command, dataDir, strategy string) (*deck.Builder, error) {
	includeCards := cmd.StringSlice("include-cards")
	excludeCards := cmd.StringSlice("exclude-cards")
	verbose := cmd.Bool("verbose")

	builder := deck.NewBuilder(dataDir)

	if err := configureEvolutions(cmd, builder); err != nil {
		return nil, err
	}

	configureCardFilters(builder, includeCards, excludeCards)

	if err := configureStrategy(cmd, builder, strategy, verbose); err != nil {
		return nil, err
	}

	configureSynergy(cmd, builder, verbose)
	configureUniqueness(cmd, builder, verbose)
	configureArchetypeAvoidance(cmd, builder, verbose)

	return builder, nil
}

// configureEvolutions sets up evolution overrides from CLI flags
func configureEvolutions(cmd *cli.Command, builder *deck.Builder) error {
	if unlockedEvos := unlockedEvolutionsFromCommand(cmd); len(unlockedEvos) > 0 {
		builder.SetUnlockedEvolutions(unlockedEvos)
	}

	if slots := cmd.Int("evolution-slots"); slots > 0 {
		builder.SetEvolutionSlotLimit(slots)
	}

	return nil
}

// configureCardFilters sets up include/exclude card filters
func configureCardFilters(builder *deck.Builder, includeCards, excludeCards []string) {
	if len(includeCards) > 0 {
		builder.SetIncludeCards(includeCards)
	}
	if len(excludeCards) > 0 {
		builder.SetExcludeCards(excludeCards)
	}
}

// configureStrategy sets up the deck building strategy if provided
func configureStrategy(cmd *cli.Command, builder *deck.Builder, strategy string, verbose bool) error {
	if strategy == "" || strings.ToLower(strings.TrimSpace(strategy)) == deckStrategyAll {
		return nil
	}

	parsedStrategy, err := deck.ParseStrategy(strategy)
	if err != nil {
		return fmt.Errorf("invalid strategy: %w", err)
	}
	if err := builder.SetStrategy(parsedStrategy); err != nil {
		return fmt.Errorf("failed to set strategy: %w", err)
	}
	if verbose {
		printf("Using deck building strategy: %s\n", parsedStrategy)
	}

	return nil
}

// configureSynergy sets up the synergy system if enabled
func configureSynergy(cmd *cli.Command, builder *deck.Builder, verbose bool) {
	if !cmd.Bool("enable-synergy") {
		return
	}

	builder.SetSynergyEnabled(true)

	if synergyWeight := cmd.Float64("synergy-weight"); synergyWeight > 0 {
		builder.SetSynergyWeight(synergyWeight)
	}

	if verbose {
		printf("Synergy scoring enabled (weight: %.2f)\n", cmd.Float64("synergy-weight"))
	}
}

func configureUniqueness(cmd *cli.Command, builder *deck.Builder, verbose bool) {
	if !cmd.Bool("prefer-unique") {
		return
	}

	builder.SetUniquenessEnabled(true)

	if uniquenessWeight := cmd.Float64("uniqueness-weight"); uniquenessWeight > 0 {
		builder.SetUniquenessWeight(uniquenessWeight)
	}

	if verbose {
		printf("Uniqueness/anti-meta scoring enabled (weight: %.2f)\n", cmd.Float64("uniqueness-weight"))
	}
}

// configureArchetypeAvoidance sets archetypes to avoid during deck building
func configureArchetypeAvoidance(cmd *cli.Command, builder *deck.Builder, verbose bool) {
	archetypes := cmd.StringSlice("avoid-archetype")
	if len(archetypes) == 0 {
		return
	}

	builder.SetAvoidArchetypes(archetypes)

	if verbose {
		printf("Avoiding archetypes: %v\n", archetypes)
	}
}

// configureFuzzIntegration sets up fuzz storage integration if enabled
func configureFuzzIntegration(cmd *cli.Command, builder *deck.Builder) error {
	fuzzStoragePath := cmd.String("fuzz-storage")
	if fuzzStoragePath == "" {
		return nil
	}

	verbose := cmd.Bool("verbose")
	fuzzIntegration := deck.NewFuzzIntegration()

	// Set custom weight if provided
	if fuzzWeight := cmd.Float64("fuzz-weight"); fuzzWeight > 0 {
		fuzzIntegration.SetWeight(fuzzWeight)
	}

	// Open storage and analyze
	storage, err := fuzzstorage.NewStorage(fuzzStoragePath)
	if err != nil {
		return fmt.Errorf("failed to open fuzz storage: %w", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			// Log but don't fail - error during cleanup is not critical
			fmt.Fprintf(os.Stderr, "warning: failed to close fuzz storage: %v\n", err)
		}
	}()

	deckLimit := cmd.Int("fuzz-deck-limit")
	if deckLimit <= 0 {
		deckLimit = 100
	}

	if err := fuzzIntegration.AnalyzeFromStorage(storage, deckLimit); err != nil {
		return fmt.Errorf("failed to analyze fuzz results: %w", err)
	}

	if fuzzIntegration.HasStats() {
		builder.SetFuzzIntegration(fuzzIntegration)
		if verbose {
			printf("Fuzz integration enabled with %d card stats (weight: %.2f)\n",
				fuzzIntegration.StatsCount(), fuzzIntegration.GetWeight())
		}
	} else if verbose {
		printf("No fuzz statistics available in storage\n")
	}

	return nil
}

// playerDataLoadResult contains the result of loading player data
type playerDataLoadResult struct {
	CardAnalysis deck.CardAnalysis
	PlayerName   string
	PlayerTag    string
}

// loadPlayerCardAnalysis loads player card data from offline analysis or API
func loadPlayerCardAnalysis(ctx context.Context, cmd *cli.Command, builder *deck.Builder, tag string) (*playerDataLoadResult, error) {
	fromAnalysis := cmd.Bool("from-analysis")
	analysisDir := cmd.String("analysis-dir")
	analysisFile := cmd.String("analysis-file")
	dataDir := cmd.String("data-dir")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")

	if fromAnalysis {
		return loadPlayerDataOffline(builder, tag, analysisDir, analysisFile, dataDir, verbose)
	}

	// ONLINE MODE
	return loadPlayerDataOnline(ctx, builder, tag, apiToken, verbose)
}

// loadPlayerDataOffline loads player data from pre-analyzed JSON files
func loadPlayerDataOffline(builder *deck.Builder, tag, analysisDir, analysisFile, dataDir string, verbose bool) (*playerDataLoadResult, error) {
	if verbose {
		printf("Building deck from offline analysis for player %s\n", tag)
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
			return nil, fmt.Errorf("failed to load analysis file %s: %w", analysisFile, err)
		}
		if verbose {
			printf("Loaded analysis from: %s\n", analysisFile)
		}
	} else {
		// Load latest analysis for player tag
		loadedAnalysis, err = builder.LoadLatestAnalysis(tag, analysisDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load analysis for player %s from %s: %w", tag, analysisDir, err)
		}
		if verbose {
			printf("Loaded latest analysis from: %s\n", analysisDir)
		}
	}

	// Use player name from analysis if available, fallback to tag
	playerName := loadedAnalysis.PlayerName
	if playerName == "" {
		playerName = tag
	}

	return &playerDataLoadResult{
		CardAnalysis: *loadedAnalysis,
		PlayerName:   playerName,
		PlayerTag:    tag,
	}, nil
}

// loadPlayerDataOnline fetches and analyzes player data from the API
//
//nolint:dupl // Shared API loading refactor tracked under clash-royale-api-sg50.
func loadPlayerDataOnline(ctx context.Context, builder *deck.Builder, tag, apiToken string, verbose bool) (*playerDataLoadResult, error) {
	client, err := requireAPIClientFromToken(apiToken, apiClientOptions{offlineAllowed: true})
	if err != nil {
		return nil, err
	}

	if verbose {
		printf("Building deck for player %s\n", tag)
	}

	// Get player information
	player, err := client.GetPlayerWithContext(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	if verbose {
		printf("Player: %s (%s)\n", player.Name, player.Tag)
		printf("Analyzing %d cards...\n", len(player.Cards))
	}

	// Perform card collection analysis
	analysisOptions := analysis.DefaultAnalysisOptions()
	cardAnalysis, err := analysis.AnalyzeCardCollection(player, analysisOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze card collection: %w", err)
	}

	deckCardAnalysis := convertToDeckCardAnalysis(cardAnalysis, player)

	return &playerDataLoadResult{
		CardAnalysis: deckCardAnalysis,
		PlayerName:   player.Name,
		PlayerTag:    player.Tag,
	}, nil
}

// applyExcludeFilter filters out excluded cards from the card analysis
func applyExcludeFilter(cardAnalysis *deck.CardAnalysis, excludeCards []string) {
	if len(excludeCards) == 0 {
		return
	}

	excludeMap := make(map[string]bool)
	for _, card := range excludeCards {
		trimmed := strings.TrimSpace(card)
		if trimmed != "" {
			excludeMap[strings.ToLower(trimmed)] = true
		}
	}

	filteredLevels := make(map[string]deck.CardLevelData)
	for cardName, cardInfo := range cardAnalysis.CardLevels {
		if !excludeMap[strings.ToLower(cardName)] {
			filteredLevels[cardName] = cardInfo
		}
	}
	cardAnalysis.CardLevels = filteredLevels
}

// displayIdealDeck shows the deck with recommended upgrades applied
//
//nolint:funlen,gocognit,gocyclo // Presentation logic slated for extraction to dedicated display package.
func displayIdealDeck(cmd *cli.Command, builder *deck.Builder, cardAnalysis deck.CardAnalysis, deckRec *deck.DeckRecommendation, playerName, playerTag string, upgrades *deck.UpgradeRecommendations) {
	if upgrades == nil || len(upgrades.Recommendations) == 0 {
		return
	}

	verbose := cmd.Bool("verbose")

	printf("\n")
	printf("============================================================================\n")
	printf("                        IDEAL DECK (WITH UPGRADES)\n")
	printf("============================================================================\n\n")

	// Create a copy of the card analysis with simulated upgrades
	idealAnalysis := deck.CardAnalysis{
		CardLevels:   make(map[string]deck.CardLevelData),
		AnalysisTime: cardAnalysis.AnalysisTime,
	}

	// Copy all card levels
	maps.Copy(idealAnalysis.CardLevels, cardAnalysis.CardLevels)

	// Apply upgrades
	printf("Simulating upgrades:\n")
	for _, rec := range upgrades.Recommendations {
		if cardData, exists := idealAnalysis.CardLevels[rec.CardName]; exists {
			oldLevel := cardData.Level
			cardData.Level = rec.TargetLevel
			idealAnalysis.CardLevels[rec.CardName] = cardData
			printf("  • %s: Level %d → %d\n", rec.CardName, oldLevel, rec.TargetLevel)
		}
	}
	printf("\n")

	// Build ideal deck with upgraded cards
	idealDeckRec, err := builder.BuildDeckFromAnalysis(idealAnalysis)
	if err != nil {
		if verbose {
			printf("Warning: Failed to build ideal deck: %v\n", err)
		}
		return
	}

	displayDeckRecommendationOffline(idealDeckRec, playerName, playerTag)

	// Show comparison
	printf("\n")
	printf("Comparison:\n")
	printf("  Current Deck:  %.2f avg elixir\n", deckRec.AvgElixir)
	printf("  Ideal Deck:    %.2f avg elixir\n", idealDeckRec.AvgElixir)

	// Show cards that changed
	currentCards := make(map[string]bool)
	for _, card := range deckRec.Deck {
		currentCards[card] = true
	}

	idealCards := make(map[string]bool)
	for _, card := range idealDeckRec.Deck {
		idealCards[card] = true
	}

	addedCards := []string{}
	removedCards := []string{}

	for card := range idealCards {
		if !currentCards[card] {
			addedCards = append(addedCards, card)
		}
	}

	for card := range currentCards {
		if !idealCards[card] {
			removedCards = append(removedCards, card)
		}
	}

	if len(addedCards) > 0 || len(removedCards) > 0 {
		printf("\n  Deck Changes:\n")
		if len(removedCards) > 0 {
			printf("    Removed: %s\n", strings.Join(removedCards, ", "))
		}
		if len(addedCards) > 0 {
			printf("    Added:   %s\n", strings.Join(addedCards, ", "))
		}
	} else {
		printf("\n  Deck composition remains the same (upgrades strengthen existing cards)\n")
	}
}

// saveDeckIfRequested saves the deck to disk if the save flag is set
func saveDeckIfRequested(cmd *cli.Command, builder *deck.Builder, deckRec *deck.DeckRecommendation, playerTag, dataDir string) error {
	saveData := cmd.Bool("save")
	verbose := cmd.Bool("verbose")

	if !saveData {
		return nil
	}

	if verbose {
		printf("\nSaving deck to: %s\n", dataDir)
	}

	deckPath, err := builder.SaveDeck(deckRec, "", playerTag)
	if err != nil {
		printf("Warning: Failed to save deck: %v\n", err)
		return nil // Don't fail the whole command for save errors
	}

	printf("\nDeck saved to: %s\n", deckPath)
	return nil
}

func displayDeckRecommendationOffline(rec *deck.DeckRecommendation, playerName, playerTag string) {
	printDeckBuilderHeader("RECOMMENDED 1v1 LADDER DECK")

	printf("Player: %s (%s)\n", playerName, playerTag)
	printf("Average Elixir: %.2f\n", rec.AvgElixir)

	// Display combat stats information if available
	if combatWeight := os.Getenv("COMBAT_STATS_WEIGHT"); combatWeight != "" {
		if combatWeight == "0" {
			printf("Scoring: Traditional only (combat stats disabled)\n")
		} else {
			printf("Scoring: %.0f%% traditional, %.0f%% combat stats\n",
				(1-parseFloatOrZero(combatWeight))*100,
				parseFloatOrZero(combatWeight)*100)
		}
	}
	printf("\n")

	// Display deck cards in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "#\tCard\tLevel\t\tElixir\tRole\n")
	fprintf(w, "─\t────\t─────\t\t──────\t────\n")

	for i, card := range rec.DeckDetail {
		evoBadge := deck.FormatEvolutionBadge(card.EvolutionLevel)
		levelStr := fmt.Sprintf("%d/%d", card.Level, card.MaxLevel)
		if evoBadge != "" {
			levelStr = fmt.Sprintf("%s (%s)", levelStr, evoBadge)
		}
		fprintf(w, "%d\t%s\t%s\t%d\t%s\n",
			i+1,
			card.Name,
			levelStr,
			card.Elixir,
			card.Role)
	}
	flushWriter(w)

	// Display strategic notes
	if len(rec.Notes) > 0 {
		printf("\nStrategic Notes:\n")
		printf("════════════════\n")
		for _, note := range rec.Notes {
			printf("• %s\n", note)
		}
	}
}

// displayUpgradeRecommendations displays upgrade recommendations in a formatted table
func displayUpgradeRecommendations(upgrades *deck.UpgradeRecommendations) {
	printDeckBuilderHeader("UPGRADE RECOMMENDATIONS")

	if len(upgrades.Recommendations) == 0 {
		fmt.Println("No upgrade recommendations - all cards are at max level!")
		return
	}

	printf("Total Gold Needed: %d\n\n", upgrades.TotalGoldNeeded)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "#\tCard\t\tLevel\t\tRarity\t\tImpact\tGold\t\tValue/1k\n")
	fprintf(w, "─\t────\t\t─────\t\t──────\t\t──────\t────\t\t────────\n")

	for i, rec := range upgrades.Recommendations {
		goldDisplay := fmt.Sprintf("%dk", rec.GoldCost/1000)
		if rec.GoldCost < 1000 {
			goldDisplay = fmt.Sprintf("%d", rec.GoldCost)
		}

		fprintf(w, "%d\t%s\t\t%d->%d\t\t%s\t\t%.1f\t%s\t\t%.2f\n",
			i+1,
			rec.CardName,
			rec.CurrentLevel,
			rec.TargetLevel,
			rec.Rarity,
			rec.ImpactScore,
			goldDisplay,
			rec.ValuePerGold,
		)
	}
	flushWriter(w)

	// Display reasons
	printf("\nWhy These Upgrades:\n")
	printf("──────────────────\n")
	for i, rec := range upgrades.Recommendations {
		if i >= 3 {
			printf("... and %d more\n", len(upgrades.Recommendations)-3)
			break
		}
		printf("%d. %s: %s\n", i+1, rec.CardName, rec.Reason)
	}
}

// parseFloatOrZero parses a float from a string, returning 0 if parsing fails.
func parseFloatOrZero(s string) float64 {
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}
	return 0
}

// buildAllStrategies builds decks using all available strategies and displays them for comparison
func buildAllStrategies(ctx context.Context, cmd *cli.Command, builder *deck.Builder, cardAnalysis deck.CardAnalysis, playerName, playerTag string) error {
	verbose := cmd.Bool("verbose")
	excludeCards := cmd.StringSlice("exclude-cards")

	strategies := getAllDeckStrategies()
	displayAllStrategiesHeader(playerName, playerTag)

	filteredAnalysis := applyCardExclusions(cardAnalysis, excludeCards)

	for i, strategy := range strategies {
		strategyBuilder, err := createStrategyBuilder(cmd)
		if err != nil {
			printf("⚠ Failed to configure strategy builder: %v\n\n", err)
			continue
		}
		if err := strategyBuilder.SetStrategy(strategy); err != nil {
			printf("⚠ Failed to set strategy %s: %v\n\n", strategy, err)
			continue
		}

		deckRec, err := strategyBuilder.BuildDeckFromAnalysis(filteredAnalysis)
		if err != nil {
			printf("⚠ Failed to build deck for strategy %s: %v\n\n", strategy, err)
			continue
		}

		displayStrategyDeck(i+1, strategy, deckRec, verbose)
	}

	return nil
}

// getAllDeckStrategies returns all available deck building strategies
func getAllDeckStrategies() []deck.Strategy {
	return []deck.Strategy{
		deck.StrategyBalanced,
		deck.StrategyAggro,
		deck.StrategyControl,
		deck.StrategyCycle,
		deck.StrategySplash,
		deck.StrategySpell,
	}
}

// displayAllStrategiesHeader prints the header for all-strategies display
func displayAllStrategiesHeader(playerName, playerTag string) {
	printDeckBuilderHeader("ALL DECK BUILDING STRATEGIES")
	printf("Player: %s (%s)\n\n", playerName, playerTag)
}

const deckBuilderHeaderWidth = 68

func printDeckBuilderHeader(title string) {
	printf("\n╔%s╗\n", strings.Repeat("═", deckBuilderHeaderWidth))
	printf("║%s║\n", padHeaderTitle(title, deckBuilderHeaderWidth))
	printf("╚%s╝\n\n", strings.Repeat("═", deckBuilderHeaderWidth))
}

func padHeaderTitle(title string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(title) >= width {
		return title[:width]
	}
	padding := width - len(title)
	left := padding / 2
	right := padding - left
	return strings.Repeat(" ", left) + title + strings.Repeat(" ", right)
}

// createStrategyBuilder creates a new builder with configuration from command
func createStrategyBuilder(cmd *cli.Command) (*deck.Builder, error) {
	builder, err := configureDeckBuilder(cmd, cmd.String("data-dir"), "")
	if err != nil {
		return nil, err
	}
	if err := configureFuzzIntegration(cmd, builder); err != nil {
		return nil, err
	}
	return builder, nil
}

// buildCardExclusionMap creates a map of cards to exclude (case-insensitive)
func buildCardExclusionMap(excludeCards []string) map[string]bool {
	excludeMap := make(map[string]bool)
	for _, card := range excludeCards {
		trimmed := strings.TrimSpace(card)
		if trimmed != "" {
			excludeMap[strings.ToLower(trimmed)] = true
		}
	}
	return excludeMap
}

// applyCardExclusions filters out excluded cards from card analysis
func applyCardExclusions(cardAnalysis deck.CardAnalysis, excludeCards []string) deck.CardAnalysis {
	if len(excludeCards) == 0 {
		return cardAnalysis
	}

	excludeMap := buildCardExclusionMap(excludeCards)
	filteredLevels := make(map[string]deck.CardLevelData)

	for cardName, cardInfo := range cardAnalysis.CardLevels {
		if !excludeMap[strings.ToLower(cardName)] {
			filteredLevels[cardName] = cardInfo
		}
	}

	filteredAnalysis := cardAnalysis
	filteredAnalysis.CardLevels = filteredLevels
	return filteredAnalysis
}

// displayStrategyDeck displays a single deck with its strategy label
func displayStrategyDeck(rank int, strategy deck.Strategy, rec *deck.DeckRecommendation, verbose bool) {
	printf("Strategy #%d: %s\n", rank, strings.ToUpper(string(strategy)))
	printf("═══════════════════════════════════════════════════════════════════\n")
	printf("Average Elixir: %.2f\n\n", rec.AvgElixir)

	// Display deck cards in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "#\tCard\tLevel\t\tElixir\tRole\n")
	fprintf(w, "─\t────\t─────\t\t──────\t────\n")

	for i, card := range rec.DeckDetail {
		evoBadge := deck.FormatEvolutionBadge(card.EvolutionLevel)
		levelStr := fmt.Sprintf("%d/%d", card.Level, card.MaxLevel)
		if evoBadge != "" {
			levelStr = fmt.Sprintf("%s (%s)", levelStr, evoBadge)
		}
		fprintf(w, "%d\t%s\t%s\t%d\t%s\n",
			i+1,
			card.Name,
			levelStr,
			card.Elixir,
			card.Role)
	}
	flushWriter(w)

	// Display strategic notes if verbose
	if verbose && len(rec.Notes) > 0 {
		printf("\nStrategic Notes:\n")
		for _, note := range rec.Notes {
			printf("• %s\n", note)
		}
	}

	printf("\n")
}

// convertToDeckCardAnalysis converts analysis.CardAnalysis to deck.CardAnalysis.
func convertToDeckCardAnalysis(cardAnalysis *analysis.CardAnalysis, player *clashroyale.Player) deck.CardAnalysis {
	result := deck.CardAnalysis{
		CardLevels:   make(map[string]deck.CardLevelData),
		AnalysisTime: cardAnalysis.AnalysisTime.Format(time.RFC3339),
		PlayerName:   player.Name,
		PlayerTag:    player.Tag,
	}
	for cardName, cardInfo := range cardAnalysis.CardLevels {
		result.CardLevels[cardName] = deck.CardLevelData{
			Level:             cardInfo.Level,
			MaxLevel:          cardInfo.MaxLevel,
			Rarity:            cardInfo.Rarity,
			Elixir:            cardInfo.Elixir,
			EvolutionLevel:    cardInfo.EvolutionLevel,
			MaxEvolutionLevel: cardInfo.MaxEvolutionLevel,
		}
	}
	return result
}
