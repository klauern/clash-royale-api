package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/klauer/clash-royale-api/go/pkg/leaderboard"
	"github.com/urfave/cli/v3"
)

func validateEvaluateFlags(deckString, fromAnalysis, playerTag, apiToken string, showUpgradeImpact bool) error {
	// Validation: Must provide either --deck or --from-analysis
	if deckString == "" && fromAnalysis == "" {
		return fmt.Errorf("must provide either --deck or --from-analysis")
	}

	if deckString != "" && fromAnalysis != "" {
		return fmt.Errorf("cannot use both --deck and --from-analysis")
	}

	// Validate upgrade impact requirements
	if showUpgradeImpact && playerTag == "" {
		return fmt.Errorf("--show-upgrade-impact requires --tag to fetch player card levels")
	}

	if showUpgradeImpact && apiToken == "" {
		return fmt.Errorf("--show-upgrade-impact requires API token (set CLASH_ROYALE_API_TOKEN or use --api-token)")
	}

	return nil
}

// loadDeckCardsFromInput loads deck cards from either deck string or analysis file
func loadDeckCardsFromInput(deckString, fromAnalysis string) ([]string, error) {
	var deckCardNames []string
	if deckString != "" {
		// Parse deck string (cards separated by dashes)
		deckCardNames = parseDeckString(deckString)
		if len(deckCardNames) != 8 {
			return nil, fmt.Errorf("deck must contain exactly 8 cards, got %d", len(deckCardNames))
		}
	} else {
		// Load deck from analysis file
		loadedCards, err := loadDeckFromAnalysis(fromAnalysis)
		if err != nil {
			return nil, fmt.Errorf("failed to load deck from analysis: %w", err)
		}
		deckCardNames = loadedCards
	}
	return deckCardNames, nil
}

// fetchPlayerContextIfNeeded fetches player context from API if tag and token are provided
func fetchPlayerContextIfNeeded(playerTag, apiToken string, verbose bool) *evaluation.PlayerContext {
	if playerTag == "" || apiToken == "" {
		return nil
	}

	if verbose {
		printf("Fetching player data for context-aware evaluation...\n")
	}

	client := clashroyale.NewClient(apiToken)
	player, err := client.GetPlayer(playerTag)
	if err != nil {
		// Log warning but continue with evaluation without context
		fprintf(os.Stderr, "Warning: Failed to fetch player data: %v\n", err)
		fprintf(os.Stderr, "Continuing with evaluation without player context.\n")
		return nil
	}

	if verbose {
		printf("Player context loaded: %s (%s), Arena: %s\n",
			player.Name, player.Tag, player.Arena.Name)
	}

	return evaluation.NewPlayerContextFromPlayer(player)
}

// persistEvaluationResult saves evaluation result to storage if player tag is provided
//
//nolint:gocyclo // Error-path branching required for storage fallbacks.
func persistEvaluationResult(result *evaluation.EvaluationResult, playerTag string, verbose bool) error {
	if playerTag == "" {
		return nil
	}

	storage, err := leaderboard.NewStorage(playerTag)
	if err != nil {
		if verbose {
			fprintf(os.Stderr, "Warning: failed to initialize storage: %v\n", err)
		}
		return err
	}
	defer func() {
		if err := storage.Close(); err != nil {
			fprintf(os.Stderr, "Warning: failed to close storage: %v\n", err)
		}
	}()

	entry := &leaderboard.DeckEntry{
		Cards:             result.Deck,
		OverallScore:      result.OverallScore,
		AttackScore:       result.Attack.Score,
		DefenseScore:      result.Defense.Score,
		SynergyScore:      result.Synergy.Score,
		VersatilityScore:  result.Versatility.Score,
		F2PScore:          result.F2PFriendly.Score,
		PlayabilityScore:  result.Playability.Score,
		Archetype:         string(result.DetectedArchetype),
		ArchetypeConf:     result.ArchetypeConfidence,
		Strategy:          "", // Single evaluations don't have a strategy
		AvgElixir:         result.AvgElixir,
		EvaluatedAt:       time.Now(),
		PlayerTag:         playerTag,
		EvaluationVersion: "1.0.0",
	}

	deckID, isNew, err := storage.InsertDeck(entry)
	if err != nil {
		if verbose {
			fprintf(os.Stderr, "Warning: failed to save deck to storage: %v\n", err)
		}
		return err
	}

	if _, err := storage.RecalculateStats(); err != nil && verbose {
		fprintf(os.Stderr, "Warning: failed to recalculate stats: %v\n", err)
	}

	if verbose {
		if isNew {
			printf("Saved deck to storage (ID: %d) at: %s\n", deckID, storage.GetDBPath())
		} else {
			printf("Updated existing deck in storage (ID: %d)\n", deckID)
		}
	}

	return nil
}

// formatEvaluationResult formats evaluation result according to the specified format
func formatEvaluationResult(result *evaluation.EvaluationResult, format string) (string, error) {
	var formattedOutput string
	var err error

	switch strings.ToLower(format) {
	case "human":
		formattedOutput = evaluation.FormatHuman(result)
	case "json":
		formattedOutput, err = evaluation.FormatJSON(result)
		if err != nil {
			return "", fmt.Errorf("failed to format JSON: %w", err)
		}
	case "csv":
		formattedOutput = evaluation.FormatCSV(result)
	case "detailed":
		formattedOutput = evaluation.FormatDetailed(result)
	default:
		return "", fmt.Errorf("unknown format: %s (supported: human, json, csv, detailed)", format)
	}

	return formattedOutput, nil
}

// writeEvaluationOutput writes formatted output to file or stdout
func writeEvaluationOutput(formattedOutput, outputFile string, verbose bool) error {
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(formattedOutput), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		if verbose {
			printf("Evaluation saved to: %s\n", outputFile)
		}
	} else {
		fmt.Print(formattedOutput)
	}
	return nil
}

// performUpgradeAnalysisIfRequested performs optional upgrade impact analysis
func performUpgradeAnalysisIfRequested(showUpgradeImpact bool, format string, deckCardNames []string, playerTag string, topUpgrades int, apiToken, dataDir string, verbose bool) error {
	if !showUpgradeImpact {
		return nil
	}

	// Only for human output format (not applicable to JSON/CSV)
	if format == "human" || format == "detailed" {
		if err := performDeckUpgradeImpactAnalysis(deckCardNames, playerTag, topUpgrades, apiToken, dataDir, verbose); err != nil {
			// Log error but don't fail the entire command
			fprintf(os.Stderr, "\nWarning: Failed to perform upgrade impact analysis: %v\n", err)
		}
	} else if verbose {
		fprintf(os.Stderr, "\nNote: Upgrade impact analysis only available for human and detailed output formats\n")
	}
	return nil
}

// deckEvaluateCommand evaluates a deck with comprehensive analysis and scoring
func deckEvaluateCommand(ctx context.Context, cmd *cli.Command) error {
	deckString := cmd.String("deck")
	playerTag := cmd.String("tag")
	fromAnalysis := cmd.String("from-analysis")
	_ = cmd.Int("arena") // TODO: Use for arena-specific analysis in future tasks
	format := cmd.String("format")
	outputFile := cmd.String("output")
	showUpgradeImpact := cmd.Bool("show-upgrade-impact")
	topUpgrades := cmd.Int("top-upgrades")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	dataDir := cmd.String("data-dir")

	// Validate flags
	if err := validateEvaluateFlags(deckString, fromAnalysis, playerTag, apiToken, showUpgradeImpact); err != nil {
		return err
	}

	// Load deck cards
	deckCardNames, err := loadDeckCardsFromInput(deckString, fromAnalysis)
	if err != nil {
		return err
	}

	if verbose {
		printf("Evaluating deck: %v\n", deckCardNames)
		printf("Output format: %s\n", format)
	}

	// Convert card names to CardCandidates and create synergy database
	deckCards := convertToCardCandidates(deckCardNames)
	synergyDB := deck.NewSynergyDatabase()

	// Fetch player context if available
	playerContext := fetchPlayerContextIfNeeded(playerTag, apiToken, verbose)

	// Evaluate the deck
	result := evaluation.Evaluate(deckCards, synergyDB, playerContext)

	// Save to persistent storage.
	if err := persistEvaluationResult(&result, playerTag, verbose); err != nil && verbose {
		fprintf(os.Stderr, "warning: failed to persist evaluation result: %v\n", err)
	}

	// Format output
	formattedOutput, err := formatEvaluationResult(&result, format)
	if err != nil {
		return err
	}

	// Write output
	if err := writeEvaluationOutput(formattedOutput, outputFile, verbose); err != nil {
		return err
	}

	// Perform upgrade analysis if requested
	return performUpgradeAnalysisIfRequested(showUpgradeImpact, format, deckCardNames, playerTag, topUpgrades, apiToken, dataDir, verbose)
}

// performDeckUpgradeImpactAnalysis performs upgrade impact analysis for a specific deck
// It fetches the player's card levels and shows which deck card upgrades would have the most impact
func performDeckUpgradeImpactAnalysis(deckCardNames []string, playerTag string, topN int, apiToken, dataDir string, verbose bool) error {
	// Create client to fetch player data
	client := clashroyale.NewClient(apiToken)

	if verbose {
		printf("\nFetching player data for upgrade impact analysis...\n")
	}

	// Get player information
	player, err := client.GetPlayer(playerTag)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	if verbose {
		printf("Player: %s (%s)\n", player.Name, player.Tag)
		printf("Analyzing deck: %v\n", deckCardNames)
	}

	// Perform card collection analysis
	analysisOptions := analysis.DefaultAnalysisOptions()
	cardAnalysis, err := analysis.AnalyzeCardCollection(player, analysisOptions)
	if err != nil {
		return fmt.Errorf("failed to analyze card collection: %w", err)
	}

	// Convert analysis.CardAnalysis to deck.CardAnalysis
	deckCardAnalysis := deck.CardAnalysis{
		CardLevels:   make(map[string]deck.CardLevelData),
		AnalysisTime: cardAnalysis.AnalysisTime.Format(time.RFC3339),
		PlayerName:   player.Name,
		PlayerTag:    player.Tag,
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

	// Create a deck builder to build the current deck
	builder := deck.NewBuilder(dataDir)

	// Find which deck cards can be upgraded and calculate their impact
	upgradeImpacts := calculateDeckCardUpgrades(deckCardNames, deckCardAnalysis, builder)

	// Sort by impact score (highest first)
	sortUpgradeImpactsByScore(upgradeImpacts)

	// Display the upgrade impact analysis
	displayDeckUpgradeImpactAnalysis(deckCardNames, upgradeImpacts, topN, player)

	return nil
}

// DeckCardUpgrade represents a potential upgrade for a card in the deck
type DeckCardUpgrade struct {
	CardName       string
	CurrentLevel   int
	TargetLevel    int
	MaxLevel       int
	Rarity         string
	ImpactScore    float64
	GoldCost       int
	CardsNeeded    int
	Reason         string
	IsKeyUpgrade   bool
	UnlocksNewDeck bool
}

// calculateDeckCardUpgrades calculates upgrade impacts for cards in the deck
func calculateDeckCardUpgrades(deckCardNames []string, cardAnalysis deck.CardAnalysis, builder *deck.Builder) []DeckCardUpgrade {
	impacts := make([]DeckCardUpgrade, 0, len(deckCardNames))

	for _, cardName := range deckCardNames {
		cardData, exists := cardAnalysis.CardLevels[cardName]
		if !exists {
			// Player doesn't have this card
			continue
		}

		// Skip if already at max level
		if cardData.Level >= cardData.MaxLevel {
			continue
		}

		// Calculate potential upgrade (typically +1 level)
		targetLevel := cardData.Level + 1
		if targetLevel > cardData.MaxLevel {
			targetLevel = cardData.MaxLevel
		}

		// Calculate gold cost and cards needed for this upgrade
		goldCost := calculateUpgradeGoldCost(cardData.Rarity, cardData.Level, targetLevel)
		cardsNeeded := calculateUpgradeCardsNeeded(cardData.Rarity, cardData.Level, targetLevel)

		// Calculate impact score (simplified - based on rarity and level gap)
		// Higher impact for upgrading win conditions and key cards
		baseImpact := calculateBaseImpact(cardData.Rarity, targetLevel)
		levelGap := float64(targetLevel - cardData.Level)
		impactScore := baseImpact * levelGap

		// Determine if this is a key upgrade
		isKeyUpgrade := cardData.Rarity == rarityLegendary || cardData.Rarity == rarityChampion

		// Generate reason
		reason := fmt.Sprintf("Upgrade %s from level %d to %d (%s)", cardName, cardData.Level, targetLevel, cardData.Rarity)

		impacts = append(impacts, DeckCardUpgrade{
			CardName:       cardName,
			CurrentLevel:   cardData.Level,
			TargetLevel:    targetLevel,
			MaxLevel:       cardData.MaxLevel,
			Rarity:         cardData.Rarity,
			ImpactScore:    impactScore,
			GoldCost:       goldCost,
			CardsNeeded:    cardsNeeded,
			Reason:         reason,
			IsKeyUpgrade:   isKeyUpgrade,
			UnlocksNewDeck: false, // TODO: Could analyze if this unlocks new archetypes
		})
	}

	return impacts
}

// calculateBaseImpact calculates the base impact score for an upgrade
func calculateBaseImpact(rarity string, level int) float64 {
	// Higher rarity = higher base impact
	// Higher level = slightly diminishing returns
	rarityMultiplier := 1.0
	switch rarity {
	case rarityCommon:
		rarityMultiplier = 1.0
	case rarityRare:
		rarityMultiplier = 2.0
	case rarityEpic:
		rarityMultiplier = 4.0
	case rarityLegendary:
		rarityMultiplier = 8.0
	case rarityChampion:
		rarityMultiplier = 10.0
	}

	// Slight diminishing returns at higher levels
	levelModifier := 1.0
	if level > 13 {
		levelModifier = 0.8
	} else if level > 11 {
		levelModifier = 0.9
	}

	return 10.0 * rarityMultiplier * levelModifier
}

// calculateUpgradeGoldCost estimates the gold cost for an upgrade
// This is a simplified calculation - actual costs vary by specific card
func calculateUpgradeGoldCost(rarity string, fromLevel, toLevel int) int {
	// Simplified gold cost calculation
	baseCost := 0
	switch rarity {
	case "Common":
		baseCost = 100
	case "Rare":
		baseCost = 400
	case "Epic":
		baseCost = 1000
	case "Legendary":
		baseCost = 4000
	case "Champion":
		baseCost = 5000
	}

	// Cost increases with level
	levelMultiplier := 1 << uint(fromLevel-1) // Doubles each level
	return baseCost * levelMultiplier * (toLevel - fromLevel)
}

// calculateUpgradeCardsNeeded estimates the number of cards needed for an upgrade
func calculateUpgradeCardsNeeded(rarity string, fromLevel, toLevel int) int {
	// Simplified card cost calculation
	baseCards := 2
	switch rarity {
	case "Common":
		baseCards = 2
	case "Rare":
		baseCards = 2
	case "Epic":
		baseCards = 2
	case "Legendary":
		baseCards = 1
	case "Champion":
		baseCards = 1
	}

	// Cards needed increase with level
	levelMultiplier := 1 << uint(fromLevel-1) // Doubles each level
	return baseCards * levelMultiplier * (toLevel - fromLevel)
}

// sortUpgradeImpactsByScore sorts upgrade impacts by score (highest first)
func sortUpgradeImpactsByScore(impacts []DeckCardUpgrade) {
	sort.Slice(impacts, func(i, j int) bool {
		return impacts[i].ImpactScore > impacts[j].ImpactScore
	})
}

// displayDeckUpgradeImpactAnalysis displays the upgrade impact analysis for deck cards
func displayDeckUpgradeImpactAnalysis(deckCardNames []string, impacts []DeckCardUpgrade, topN int, player *clashroyale.Player) {
	printf("\n")
	printf("╔════════════════════════════════════════════════════════════════════╗\n")
	printf("║                    UPGRADE IMPACT ANALYSIS                          ║\n")
	printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	printf("Player: %s (%s)\n", player.Name, player.Tag)
	printf("Deck: %v\n\n", deckCardNames)

	if len(impacts) == 0 {
		printf("✨ All deck cards are already at max level!\n")
		return
	}

	// Limit to top N
	displayCount := topN
	if displayCount > len(impacts) {
		displayCount = len(impacts)
	}

	// Calculate total costs
	totalGold := 0
	totalCards := 0
	for i := 0; i < displayCount; i++ {
		totalGold += impacts[i].GoldCost
		totalCards += impacts[i].CardsNeeded
	}

	printf("Summary:\n")
	printf("════════\n")
	printf("Upgradable Cards: %d\n", len(impacts))
	printf("Top %d Upgrades: %d gold total\n\n", displayCount, totalGold)

	printf("Most Impactful Upgrades:\n")
	printf("════════════════════════\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "#\tCard\tLevel\t\tRarity\t\tImpact\tGold\t\tCards\n")
	fprintf(w, "─\t────\t─────\t\t──────\t\t──────\t────\t\t─────\n")

	for i := 0; i < displayCount; i++ {
		upgrade := impacts[i]
		keyMarker := ""
		if upgrade.IsKeyUpgrade {
			keyMarker = " ⭐"
		}

		goldDisplay := formatGoldCost(upgrade.GoldCost)
		fprintf(w, "%d\t%s%s\t%d->%d\t\t%s\t\t%.1f\t%s\t\t%d\n",
			i+1,
			upgrade.CardName,
			keyMarker,
			upgrade.CurrentLevel,
			upgrade.TargetLevel,
			upgrade.Rarity,
			upgrade.ImpactScore,
			goldDisplay,
			upgrade.CardsNeeded,
		)
	}
	flushWriter(w)

	printf("\n")
	printf("💡 Tip: Focus on upgrading cards with the highest impact score first.\n")
	printf("   Win conditions and Legendary/Champion cards typically provide the best ROI.\n")
}

// formatGoldCost formats a gold cost for display
func formatGoldCost(gold int) string {
	if gold >= 1000 {
		return fmt.Sprintf("%dk", gold/1000)
	}
	return fmt.Sprintf("%d", gold)
}

// convertToCardCandidates converts card names to CardCandidate structs with inferred data
// For evaluation purposes, we create cards with reasonable defaults based on card names
func convertToCardCandidates(cardNames []string) []deck.CardCandidate {
	deckCards := make([]deck.CardCandidate, 0, len(cardNames))

	for _, name := range cardNames {
		// Create a CardCandidate with inferred properties
		candidate := deck.CardCandidate{
			Name:     name,
			Level:    11, // Default level
			MaxLevel: 15, // Default max level
			Rarity:   inferRarity(name),
			Elixir:   config.GetCardElixir(name, 0),
			Role:     inferRole(name),
			Stats:    inferStats(name),
		}
		deckCards = append(deckCards, candidate)
	}

	return deckCards
}

// inferRarity infers card rarity from card name
func inferRarity(name string) string {
	// This is a simplified version - in reality, you'd look this up from a database
	// For now, we'll use common as default
	return "Common"
}

// inferRole infers card role from card name
func inferRole(name string) *deck.CardRole {
	lowercaseName := strings.ToLower(name)

	// Win conditions
	if strings.Contains(lowercaseName, "hog") ||
		strings.Contains(lowercaseName, "balloon") ||
		strings.Contains(lowercaseName, "giant") ||
		strings.Contains(lowercaseName, "golem") ||
		strings.Contains(lowercaseName, "graveyard") {
		role := deck.RoleWinCondition
		return &role
	}

	// Spells (big)
	if strings.Contains(lowercaseName, "fireball") ||
		strings.Contains(lowercaseName, "lightning") ||
		strings.Contains(lowercaseName, "rocket") {
		role := deck.RoleSpellBig
		return &role
	}

	// Buildings
	if strings.Contains(lowercaseName, "tesla") ||
		strings.Contains(lowercaseName, "cannon") ||
		strings.Contains(lowercaseName, "inferno tower") {
		role := deck.RoleBuilding
		return &role
	}

	// Support troops
	if strings.Contains(lowercaseName, "wizard") ||
		strings.Contains(lowercaseName, "witch") ||
		strings.Contains(lowercaseName, "musketeer") {
		role := deck.RoleSupport
		return &role
	}

	// Default to support
	role := deck.RoleSupport
	return &role
}

// inferStats infers basic combat stats from card name
func inferStats(name string) *clashroyale.CombatStats {
	// For evaluation purposes, we use simplified stats
	// In a full implementation, this would come from the API
	return &clashroyale.CombatStats{
		Targets:         "Air & Ground", // Default to versatile
		DamagePerSecond: 100,            // Default DPS
		Hitpoints:       1000,           // Default HP
		HitSpeed:        1.5,            // Default hit speed
		Range:           5.0,            // Default range
	}
}

// parseDeckString parses a deck string into individual card names
func parseDeckString(deckStr string) []string {
	// Split by dash and trim whitespace
	parts := strings.Split(deckStr, "-")
	cards := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			cards = append(cards, trimmed)
		}
	}

	return cards
}

// loadDeckFromAnalysis loads a deck from an analysis JSON file
func loadDeckFromAnalysis(filePath string) ([]string, error) {
	// Read the analysis file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read analysis file: %w", err)
	}

	// Parse JSON to extract deck cards
	var analysisData map[string]interface{}
	if err := json.Unmarshal(data, &analysisData); err != nil {
		return nil, fmt.Errorf("failed to parse analysis JSON: %w", err)
	}

	// Extract deck cards from analysis
	// Assuming the analysis file has a "current_deck" or "deck" field
	deckField, ok := analysisData["current_deck"]
	if !ok {
		deckField, ok = analysisData["deck"]
		if !ok {
			return nil, fmt.Errorf("analysis file does not contain 'current_deck' or 'deck' field")
		}
	}

	// Convert to string array
	deckArray, ok := deckField.([]interface{})
	if !ok {
		return nil, fmt.Errorf("deck field is not an array")
	}

	cards := make([]string, 0, len(deckArray))
	for _, card := range deckArray {
		cardStr, ok := card.(string)
		if !ok {
			return nil, fmt.Errorf("deck contains non-string card")
		}
		cards = append(cards, cardStr)
	}

	if len(cards) != 8 {
		return nil, fmt.Errorf("deck must contain exactly 8 cards, got %d", len(cards))
	}

	return cards, nil
}

type deckFilePayload struct {
	Deck       []string          `json:"deck"`
	DeckDetail []deck.CardDetail `json:"deck_detail"`
}

func buildCandidatesFromDetails(details []deck.CardDetail) []deck.CardCandidate {
	deckCards := make([]deck.CardCandidate, 0, len(details))
	for _, detail := range details {
		role := inferRole(detail.Name)
		if detail.Role != "" {
			parsedRole := deck.CardRole(detail.Role)
			role = &parsedRole
		}

		rarity := detail.Rarity
		if rarity == "" {
			rarity = inferRarity(detail.Name)
		}

		deckCards = append(deckCards, deck.CardCandidate{
			Name:              detail.Name,
			Level:             detail.Level,
			MaxLevel:          detail.MaxLevel,
			Rarity:            rarity,
			Elixir:            config.GetCardElixir(detail.Name, detail.Elixir),
			Role:              role,
			EvolutionLevel:    detail.EvolutionLevel,
			MaxEvolutionLevel: detail.MaxEvolutionLevel,
			Stats:             inferStats(detail.Name),
		})
	}

	return deckCards
}

func loadDeckCandidatesFromFile(filePath string) ([]deck.CardCandidate, bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, false, err
	}

	var payload deckFilePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, false, err
	}

	if len(payload.DeckDetail) != 8 {
		return nil, false, nil
	}

	return buildCandidatesFromDetails(payload.DeckDetail), true, nil
}

// formatStars formats a star rating as visual stars
func formatStars(stars int) string {
	const filledStar = "★"
	const emptyStar = "☆"

	result := ""
	for i := 0; i < 3; i++ {
		if i < stars {
			result += filledStar
		} else {
			result += emptyStar
		}
	}
	return result
}
