package main

import (
	"context"
	"encoding/csv"
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
	"github.com/klauer/clash-royale-api/go/pkg/recommend"
	"github.com/urfave/cli/v3"
)

//nolint:gocognit,gocyclo,funlen // Supports offline/online execution branches; full split tracked in clash-royale-api-sb3q.
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
			printf("Generating recommendations from offline analysis for player %s\n", tag)
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
				printf("Loaded analysis from: %s\n", analysisFile)
			}
		} else {
			// Load latest analysis for player tag
			loadedAnalysis, err = builder.LoadLatestAnalysis(tag, analysisDir)
			if err != nil {
				return fmt.Errorf("failed to load analysis for player %s from %s: %w", tag, analysisDir, err)
			}
			if verbose {
				printf("Loaded latest analysis from: %s\n", analysisDir)
			}
		}

		deckCardAnalysis = *loadedAnalysis
		playerTag = tag
		playerName = tag // Use tag as name in offline mode
	} else {
		// ONLINE MODE: Fetch from API
		if apiToken == "" {
			return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag. Use --from-analysis for offline mode")
		}

		client := clashroyale.NewClient(apiToken)

		if verbose {
			printf("Generating recommendations for player %s\n", tag)
		}

		// Get player information
		player, err := client.GetPlayerWithContext(ctx, tag)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}

		playerName = player.Name
		playerTag = player.Tag

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

		// Convert analysis.CardAnalysis to deck.CardAnalysis
		deckCardAnalysis = deck.CardAnalysis{
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
		printf("\nExported to CSV: %s\n", getRecommendationsCSVPath(dataDir, playerTag))
	}

	return nil
}

// displayRecommendations displays deck recommendations in a formatted table
func displayRecommendations(result *recommend.RecommendationResult, verbose bool) {
	printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	printf("║                    DECK RECOMMENDATIONS                            ║\n")
	printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	printf("Player: %s (%s)\n", result.PlayerName, result.PlayerTag)
	if result.TopArchetype != "" {
		printf("Top Archetype Match: %s\n", result.TopArchetype)
	}
	printf("Generated: %s\n", result.GeneratedAt)
	printf("\n")

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

	printf("Deck #%d [%s - %s]\n", rank, rec.ArchetypeName, typeLabel)
	printf("─────────────────────────────────────────────────────\n")
	printf("Compatibility: %.1f%% | Synergy: %.1f%% | Overall: %.1f%%\n",
		rec.CompatibilityScore, rec.SynergyScore, rec.OverallScore)
	printf("Avg Elixir: %.2f\n", rec.Deck.AvgElixir)

	// Display cards in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "Card\t\tLevel\tElixir\n")
	for _, card := range rec.Deck.DeckDetail {
		fprintf(w, "%s\t\t%d\t%d\n", card.Name, card.Level, card.Elixir)
	}
	flushWriter(w)

	// Display evolution slots if any
	if len(rec.Deck.EvolutionSlots) > 0 {
		printf("Evolution Slots: %s\n", strings.Join(rec.Deck.EvolutionSlots, ", "))
	}

	// Display reasons
	if len(rec.Reasons) > 0 {
		printf("\nWhy Recommended:\n")
		for _, reason := range rec.Reasons {
			printf("  • %s\n", reason)
		}
	}

	if verbose {
		// Display upgrade cost if available
		if rec.UpgradeCost.CardsNeeded > 0 {
			printf("\nUpgrade needed: %d cards, %d gold\n",
				rec.UpgradeCost.CardsNeeded, rec.UpgradeCost.GoldNeeded)
		}

		// Display notes
		if len(rec.Deck.Notes) > 0 {
			printf("\nNotes:\n")
			for _, note := range rec.Deck.Notes {
				printf("  • %s\n", note)
			}
		}
	}
}

// exportRecommendationsToCSV exports recommendations to CSV file
func exportRecommendationsToCSV(dataDir string, result *recommend.RecommendationResult) error {
	csvDir := filepath.Join(dataDir, "csv")
	if err := os.MkdirAll(csvDir, 0o755); err != nil {
		return fmt.Errorf("failed to create CSV directory: %w", err)
	}

	csvPath := getRecommendationsCSVPath(dataDir, result.PlayerTag)

	file, err := os.Create(csvPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer closeFile(file)

	writer := csv.NewWriter(file)

	// Write header
	header := []string{
		"Rank", "Archetype", "Type", "Compatibility", "Synergy", "Overall",
		"AvgElixir", "Cards", "EvolutionSlots", "Reasons",
	}
	if err := writer.Write(header); err != nil {
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
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV data: %w", err)
	}

	return nil
}

// getRecommendationsCSVPath returns the CSV file path for recommendations
func getRecommendationsCSVPath(dataDir, playerTag string) string {
	return filepath.Join(dataDir, "csv", fmt.Sprintf("recommendations_%s.csv", sanitizePathComponent(playerTag)))
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
		printf("Generating mulligan guide for deck: %s\n", deckName)
		printf("Cards: %v\n", cardNames)
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
			printf("\nSaving mulligan guide to: %s\n", dataDir)
		}
		if err := saveMulliganGuide(dataDir, guide); err != nil {
			printf("Warning: Failed to save mulligan guide: %v\n", err)
		} else {
			printf("\nMulligan guide saved to file\n")
		}
	}

	return nil
}

// displayMulliganGuide displays a formatted mulligan guide
func displayMulliganGuide(guide *mulligan.MulliganGuide) {
	printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	printf("║                 MULLIGAN GUIDE - OPENING PLAYS                    ║\n")
	printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	printf("Deck: %s (%s)\n", guide.DeckName, guide.Archetype.String())
	printf("Generated: %s\n\n", guide.GeneratedAt.Format("2006-01-02 15:04:05"))

	printf("📋 General Principles:\n")
	for _, principle := range guide.GeneralPrinciples {
		printf("   • %s\n", principle)
	}
	fmt.Println()

	printf("🃏 Deck Composition:\n")
	printf("   Cards: %s\n", strings.Join(guide.DeckCards, ", "))
	fmt.Println()

	if len(guide.IdealOpenings) > 0 {
		printf("✅ Ideal Opening Cards:\n")
		for _, opening := range guide.IdealOpenings {
			printf("   ✓ %s\n", opening)
		}
		fmt.Println()
	}

	if len(guide.NeverOpenWith) > 0 {
		printf("❌ Never Open With:\n")
		for _, never := range guide.NeverOpenWith {
			printf("   ✗ %s\n", never)
		}
		fmt.Println()
	}

	printf("🎮 Matchup-Specific Openings:\n\n")

	for i, matchup := range guide.Matchups {
		printf("%d. VS %s\n", i+1, matchup.OpponentType)
		printf("   ▶ Opening Play: %s\n", matchup.OpeningPlay)
		printf("   ▶ Why: %s\n", matchup.Reason)
		printf("   ▶ Backup: %s\n", matchup.Backup)
		printf("   ▶ Key Cards: %s\n", strings.Join(matchup.KeyCards, ", "))
		printf("   ▶ Danger Level: %s\n", matchup.DangerLevel)
		fmt.Println()
	}
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

//nolint:gocyclo // Filename sanitization intentionally handles multiple character classes explicitly.
func sanitizePathComponent(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "deck"
	}

	var b strings.Builder
	for _, r := range value {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
		if isAlphaNum || r == '-' || r == '_' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}

	sanitized := strings.Trim(b.String(), "._-")
	if sanitized == "" {
		return "deck"
	}
	return sanitized
}

// saveMulliganGuide saves the mulligan guide to a JSON file
func saveMulliganGuide(dataDir string, guide *mulligan.MulliganGuide) error {
	// Create mulligan directory if it doesn't exist
	mulliganDir := filepath.Join(dataDir, "mulligan")
	if err := os.MkdirAll(mulliganDir, 0o755); err != nil {
		return fmt.Errorf("failed to create mulligan directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := guide.GeneratedAt.Format("20060102_150405")
	safeDeckName := sanitizePathComponent(guide.DeckName)
	filename := filepath.Join(mulliganDir, fmt.Sprintf("%s_%s.json", safeDeckName, timestamp))

	// Save as JSON
	data, err := json.MarshalIndent(guide, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mulligan guide: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("failed to write mulligan guide file: %w", err)
	}

	return nil
}
