package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/klauer/clash-royale-api/go/internal/exporter/csv"
	"github.com/klauer/clash-royale-api/go/internal/storage"
	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/urfave/cli/v3"
)

// Version information (set via ldflags during build)
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	// Get default paths
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	defaultDataDir := filepath.Join(homeDir, ".cr-api")

	// Export manager will be created per command as needed

	// Create the CLI app
	cmd := &cli.Command{
		Name:    "cr-api",
		Usage:   "Clash Royale API client and analysis tool",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildTime),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "api-token",
				Aliases: []string{"t"},
				Usage:   "Clash Royale API token",
				Sources: cli.EnvVars("CLASH_ROYALE_API_TOKEN"),
			},
			&cli.StringFlag{
				Name:    "data-dir",
				Aliases: []string{"d"},
				Value:   defaultDataDir,
				Usage:   "Data storage directory",
				Sources: cli.EnvVars("DATA_DIR"),
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose logging",
			},
		},
		Commands: []*cli.Command{
			addArchetypeCommands(),
			addDeckCommands(),
			addEvolutionCommands(),
			addEventCommands(),
			addExportCommands(),
			addUpgradeImpactCommands(),
			addWhatIfCommands(),
			addOnboardCommand(),
			addCompareCommands(),
			addPlayerCommand(),
			addCardsCommand(),
			addAnalyzeCommand(),
			addPlaystyleCommand(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func playerCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	showChests := cmd.Bool("chests")
	saveData := cmd.Bool("save")
	exportCSV := cmd.Bool("export-csv")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")

	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	client := clashroyale.NewClient(apiToken)

	if verbose {
		printf("Getting player data for tag: %s\n", tag)
	}

	// Get player information
	player, err := client.GetPlayerWithContext(ctx, tag)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Display player info
	displayPlayerInfo(player)

	// Get and display chest cycle if requested
	if showChests {
		if verbose {
			printf("\nFetching upcoming chests...\n")
		}
		chests, err := client.GetPlayerUpcomingChestsWithContext(ctx, tag)
		if err != nil {
			printf("Warning: Failed to get chests: %v\n", err)
		} else {
			displayUpcomingChests(chests)
		}
	}

	// Save player data if requested
	if saveData {
		dataDir := cmd.String("data-dir")
		if verbose {
			printf("\nSaving player data to: %s\n", dataDir)
		}
		if err := savePlayerData(dataDir, player); err != nil {
			printf("Warning: Failed to save player data: %v\n", err)
		} else {
			printf("Player data saved to: %s/players/%s.json\n", dataDir, player.Tag)
		}
	}

	// Export to CSV if requested
	if exportCSV {
		dataDir := cmd.String("data-dir")
		if verbose {
			printf("\nExporting player data to CSV...\n")
		}
		playerExporter := csv.NewPlayerExporter()
		if err := playerExporter.Export(dataDir, player); err != nil {
			printf("Warning: Failed to export player data: %v\n", err)
		} else {
			printf("Player data exported to CSV\n")
		}
	}

	return nil
}

func displayPlayerInfo(p *clashroyale.Player) {
	printf("\nPlayer Information:\n")
	printf("==================\n")
	printf("Name: %s (%s)\n", p.Name, p.Tag)
	printf("Level: %d (Experience: %d/%d)\n", p.ExpLevel, p.ExpPoints, p.Experience)
	printf("Trophies: %d (Best: %d)\n", p.Trophies, p.BestTrophies)
	printf("Arena: %s\n", p.Arena.Name)
	printf("League: %s\n", p.League.Name)

	if p.Clan != nil {
		printf("\nClan Information:\n")
		printf("Clan: %s (%s)\n", p.Clan.Name, p.Clan.Tag)
		printf("Role: %s\n", p.Role)
		printf("Clan Trophies: %d\n", p.Clan.ClanScore)
	}

	printf("\nBattle Statistics:\n")
	if p.Wins+p.Losses > 0 {
		printf("Wins: %d | Losses: %d | Win Rate: %.1f%%\n",
			p.Wins, p.Losses,
			float64(p.Wins)/float64(p.Wins+p.Losses)*100)
	} else {
		printf("Wins: %d | Losses: %d\n", p.Wins, p.Losses)
	}
	printf("3-Crown Wins: %d\n", p.ThreeCrownWins)
	printf("Total Battles: %d\n", p.BattleCount)

	printf("\nCard Collection:\n")
	printf("Total Cards: %d\n", len(p.Cards))
	printf("Star Points: %d\n", p.StarPoints)
}

func displayUpcomingChests(chests *clashroyale.ChestCycle) {
	printf("\n\nUpcoming Chests:\n")
	printf("================\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "Slot\tChest Name\n")
	fprintf(w, "----\t----------\n")

	for i, chest := range chests.Items {
		fprintf(w, "%d\t%s\n", chest.Index+1, chest.Name)
		if i >= 9 { // Show first 10 chests
			break
		}
	}

	flushWriter(w)
}

func savePlayerData(dataDir string, p *clashroyale.Player) error {
	// Create data directory if it doesn't exist
	playersDir := filepath.Join(dataDir, "players")
	if err := os.MkdirAll(playersDir, 0o755); err != nil {
		return fmt.Errorf("failed to create players directory: %w", err)
	}

	// Save as JSON
	filename := filepath.Join(playersDir, fmt.Sprintf("%s.json", p.Tag))
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal player data: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("failed to write player file: %w", err)
	}

	return nil
}

func cardsCommand(ctx context.Context, cmd *cli.Command) error {
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	exportCSV := cmd.Bool("export-csv")
	dataDir := cmd.String("data-dir")

	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	client := clashroyale.NewClient(apiToken)

	if verbose {
		printf("Fetching card database...\n")
	}

	cards, err := client.GetCardsWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cards: %w", err)
	}

	if err := cacheStaticCards(dataDir, cards); err != nil && verbose {
		printf("Warning: Failed to cache card database: %v\n", err)
	}

	// Always display cards unless only exporting
	if !exportCSV {
		displayCards(cards.Items)
	}

	// Export to CSV if requested
	if exportCSV {
		if verbose {
			printf("\nExporting card database to CSV...\n")
		}
		cardsExporter := csv.NewCardsExporter()
		if err := cardsExporter.Export(dataDir, cards.Items); err != nil {
			printf("Warning: Failed to export cards: %v\n", err)
		} else {
			printf("Card database exported to CSV\n")
		}
	}

	return nil
}

func displayCards(cards []clashroyale.Card) {
	printf("\nCard Database:\n")
	printf("=============\n")
	printf("Total Cards: %d\n\n", len(cards))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "Name\tRarity\tElixir\tType\n")
	fprintf(w, "----\t------\t------\t----\n")

	for _, card := range cards {
		fprintf(w, "%s\t%s\t%d\t%s\n",
			card.Name,
			card.Rarity,
			card.ElixirCost,
			card.Type)
	}

	flushWriter(w)
}

func analyzeCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	saveData := cmd.Bool("save")
	exportCSV := cmd.Bool("export-csv")

	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	// Build analysis options from CLI flags
	options := analysis.AnalysisOptions{
		IncludeMaxLevel:   cmd.Bool("include-max-level"),
		MinPriorityScore:  cmd.Float64("min-priority-score"),
		FocusRarities:     cmd.StringSlice("focus-rarities"),
		ExcludeCards:      cmd.StringSlice("exclude-cards"),
		PrioritizeWinCons: cmd.Bool("prioritize-win-cons"),
		TopN:              cmd.Int("top-n"),
	}

	client := clashroyale.NewClient(apiToken)

	if verbose {
		printf("Analyzing card collection for tag: %s\n", tag)
	}

	// Get player information
	player, err := client.GetPlayerWithContext(ctx, tag)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	if verbose {
		printf("Player: %s (%s)\n", player.Name, player.Tag)
		printf("Analyzing %d cards...\n", len(player.Cards))
	}

	// Perform analysis
	cardAnalysis, err := analysis.AnalyzeCardCollection(player, options)
	if err != nil {
		return fmt.Errorf("failed to analyze card collection: %w", err)
	}

	// Display analysis results
	displayAnalysis(cardAnalysis)

	// Save analysis if requested
	if saveData {
		dataDir := cmd.String("data-dir")
		if verbose {
			printf("\nSaving analysis to: %s\n", dataDir)
		}
		pb := storage.NewPathBuilder(dataDir)
		analysisPath, pathErr := pb.GetAnalysisFilePath(cardAnalysis.PlayerTag)
		if pathErr != nil {
			printf("Warning: failed to build analysis path for player tag %q: %v\n", cardAnalysis.PlayerTag, pathErr)
			analysisPath = pb.GetAnalysisDir()
		}
		if err := saveAnalysisData(dataDir, cardAnalysis); err != nil {
			printf("Warning: Failed to save analysis: %v\n", err)
		} else {
			printf("Analysis saved to: %s\n", analysisPath)
		}
	}

	// Export to CSV if requested
	if exportCSV {
		dataDir := cmd.String("data-dir")
		if verbose {
			printf("\nExporting analysis to CSV...\n")
		}
		analysisExporter := csv.NewAnalysisExporter()
		if err := analysisExporter.Export(dataDir, cardAnalysis); err != nil {
			printf("Warning: Failed to export analysis: %v\n", err)
		} else {
			printf("Analysis exported to CSV\n")
		}
	}

	return nil
}

func displayAnalysisHeader(a *analysis.CardAnalysis) {
	printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	printf("║                   CARD COLLECTION ANALYSIS                         ║\n")
	printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	printf("Player: %s (%s)\n", a.PlayerName, a.PlayerTag)
	printf("Analysis Time: %s\n\n", a.AnalysisTime.Format("2006-01-02 15:04:05"))
}

func displayAnalysisSummary(a *analysis.CardAnalysis) {
	// Display summary
	printf("Summary:\n")
	printf("════════\n")
	printf("Total Cards:        %d\n", a.Summary.TotalCards)
	printf("Max Level Cards:    %d (%.1f%%)\n", a.Summary.MaxLevelCards, a.Summary.CompletionPercent)
	printf("Average Level:      %.2f\n", a.Summary.AvgCardLevel)
	printf("Ready to Upgrade:   %d\n", a.Summary.UpgradableCards)

	// Calculate cards near max from rarity breakdown
	cardsNearMax := 0
	for _, stats := range a.RarityBreakdown {
		cardsNearMax += stats.CardsNearMax
	}
	printf("Near Max (1-2 lvl): %d\n", cardsNearMax)
	printf("\n")
}

func displayRarityBreakdown(a *analysis.CardAnalysis) {
	// Display rarity breakdown
	if len(a.RarityBreakdown) > 0 {
		printf("Rarity Breakdown:\n")
		printf("═════════════════\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fprintf(w, "Rarity\tTotal\tMax Lvl\tAvg Lvl\tReady\tNear Max\n")
		fprintf(w, "──────\t─────\t───────\t───────\t─────\t────────\n")

		// Display in order: Common, Rare, Epic, Legendary, Champion
		order := []string{"Common", "Rare", "Epic", "Legendary", "Champion"}
		for _, rarity := range order {
			if stats, ok := a.RarityBreakdown[rarity]; ok {
				fprintf(w, "%s\t%d\t%d\t%.1f\t%d\t%d\n",
					rarity,
					stats.TotalCards,
					stats.MaxLevelCards,
					stats.AvgLevel,
					stats.CardsReadyUpgrade,
					stats.CardsNearMax)
			}
		}
		flushWriter(w)
		printf("\n")
	}
}

func displayUpgradePriorities(a *analysis.CardAnalysis) {
	// Display upgrade priorities
	if len(a.UpgradePriority) > 0 {
		printf("Upgrade Priorities:\n")
		printf("═══════════════════\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fprintf(w, "Card\tRarity\tLevel\tOwned\tNeeded\tScore\tPriority\tReasons\n")
		fprintf(w, "────\t──────\t─────\t─────\t──────\t─────\t────────\t───────\n")

		for _, priority := range a.UpgradePriority {
			reasons := ""
			if len(priority.Reasons) > 0 {
				reasons = priority.Reasons[0]
				if len(priority.Reasons) > 1 {
					reasons += fmt.Sprintf(" +%d", len(priority.Reasons)-1)
				}
			}

			fprintf(w, "%s\t%s\t%d/%d\t%d\t%d\t%.1f\t%s\t%s\n",
				priority.CardName,
				priority.Rarity,
				priority.CurrentLevel,
				priority.MaxLevel,
				priority.CardsOwned,
				priority.CardsNeeded,
				priority.PriorityScore,
				priority.Priority,
				reasons)
		}
		flushWriter(w)
	} else {
		printf("No upgrade priorities found.\n")
	}
}

func displayAnalysis(a *analysis.CardAnalysis) {
	displayAnalysisHeader(a)
	displayAnalysisSummary(a)
	displayRarityBreakdown(a)
	displayUpgradePriorities(a)
}

func saveAnalysisData(dataDir string, a *analysis.CardAnalysis) error {
	// Use storage.PathBuilder for consistent file naming
	pb := storage.NewPathBuilder(dataDir)

	// Ensure analysis directory exists
	if err := os.MkdirAll(pb.GetAnalysisDir(), 0o755); err != nil {
		return fmt.Errorf("failed to create analysis directory: %w", err)
	}

	// Get standardized file path with timestamp
	filename, err := pb.GetAnalysisFilePath(a.PlayerTag)
	if err != nil {
		return fmt.Errorf("failed to sanitize player tag %q: %w", a.PlayerTag, err)
	}

	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analysis data: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("failed to write analysis file: %w", err)
	}

	return nil
}

func playstyleCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	recommendDecks := cmd.Bool("recommend-decks")
	saveData := cmd.Bool("save")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	dataDir := cmd.String("data-dir")

	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	client := clashroyale.NewClient(apiToken)

	if verbose {
		printf("Analyzing playstyle for tag: %s\n", tag)
	}

	// Get player information
	player, err := client.GetPlayerWithContext(ctx, tag)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	if verbose {
		printf("Player: %s (%s)\n", player.Name, player.Tag)
		printf("Analyzing playstyle based on %d battles...\n", player.BattleCount)
	}

	// Perform playstyle analysis
	playstyleAnalysis, err := analysis.AnalyzePlaystyle(player)
	if err != nil {
		return fmt.Errorf("failed to analyze playstyle: %w", err)
	}

	// Display playstyle analysis
	displayPlaystyleAnalysis(playstyleAnalysis)

	// Get deck recommendations if requested
	var recommendations *analysis.DeckRecommendationResult
	if recommendDecks {
		if verbose {
			printf("\nGenerating deck recommendations...\n")
		}
		recommendations, err = analysis.RecommendDecks(playstyleAnalysis, dataDir)
		if err != nil {
			printf("Warning: Failed to generate deck recommendations: %v\n", err)
		} else {
			displayDeckRecommendations(recommendations)
		}
	}

	// Save analysis if requested
	if saveData {
		if verbose {
			printf("\nSaving playstyle analysis to: %s\n", dataDir)
		}

		// Save playstyle analysis
		if err := savePlaystyleData(dataDir, playstyleAnalysis, recommendations); err != nil {
			printf("Warning: Failed to save playstyle analysis: %v\n", err)
		} else {
			printf("Playstyle analysis saved to: %s/analysis/playstyle_%s.json\n", dataDir, playstyleAnalysis.PlayerTag)
		}
	}

	return nil
}

func displayPlaystyleAnalysis(p *analysis.PlaystyleAnalysis) {
	printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	printf("║                    PLAYSTYLE ANALYSIS                             ║\n")
	printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	printf("Player: %s (%s)\n", p.PlayerName, p.PlayerTag)
	printf("Analysis Time: %s\n\n", p.AnalysisTime.Format("2006-01-02 15:04:05"))

	// Display statistics
	printf("Overall Statistics:\n")
	printf("═══════════════════\n")
	printf("Total Battles:     %d\n", p.TotalBattles)
	printf("Record:            %dW - %dL\n", p.Wins, p.Losses)
	printf("Win Rate:          %.1f%%\n", p.WinRate)
	printf("Three-Crown Wins:  %d (%.1f%% of wins)\n\n", p.ThreeCrownWins, p.ThreeCrownRate)

	// Display playstyle profile
	printf("Playstyle Profile:\n")
	printf("═══════════════════\n")
	printf("Aggression Level:  %s\n", p.AggressionLevel)
	printf("Consistency:       %s\n", p.Consistency)
	printf("Current Deck Style: %s\n", p.DeckStyle)
	if p.CurrentWinCondition != "" {
		printf("Current Win Condition: %s\n", p.CurrentWinCondition)
		printf("Current Average Elixir: %.1f\n", p.CurrentDeckAvgElixir)
	}
	printf("Deck Elixir Distribution: %s\n", p.DeckElixirDistribution)
	printf("\n")

	// Display traits
	printf("Key Traits:\n")
	printf("════════════\n")
	for _, trait := range p.PlaystyleTraits {
		printf("• %s\n", trait)
	}
	printf("\n")

	// Display current deck if available
	if len(p.CurrentDeckCards) > 0 {
		printf("Current Deck Cards:\n")
		printf("═══════════════════\n")
		for _, card := range p.CurrentDeckCards {
			printf("• %s\n", card)
		}
		printf("\n")
	}
}

func displayDeckRecommendations(r *analysis.DeckRecommendationResult) {
	if r.Recommended == nil {
		printf("\nNo deck recommendations available.\n")
		return
	}

	printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	printf("║                    RECOMMENDED DECK                               ║\n")
	printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	topDeck := r.Recommended.Deck
	printf("Deck: %s\n", topDeck.DeckName)
	printf("Win Condition: %s\n", topDeck.WinCondition)
	printf("Average Elixir: %.1f\n", topDeck.AverageElixir)
	printf("Match Score: %d/100\n", r.Recommended.Score)
	printf("Compatibility: %s\n\n", r.Recommended.Compatibility)

	printf("Why this deck:\n")
	for _, reason := range r.Recommended.Reasons {
		printf("✓ %s\n", reason)
	}
	printf("\n")

	if len(topDeck.DeckDetail) > 0 {
		printf("Cards:\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fprintf(w, "  Card                 Level       Elixir\n")
		fprintf(w, "  ----                 -----       ------\n")
		for _, card := range topDeck.DeckDetail {
			levelPct := float64(card.Level) / float64(card.MaxLevel) * 100
			indicator := "✓"
			if levelPct <= 60 {
				indicator = "○"
			}
			levelStr := fmt.Sprintf("%2d/%-2d", card.Level, card.MaxLevel)
			if evoBadge := deck.FormatEvolutionBadge(card.EvolutionLevel); evoBadge != "" {
				levelStr = fmt.Sprintf("%s (%s)", levelStr, evoBadge)
			}
			fprintf(w, "  %s%-20s %-11s %2d\n", indicator, card.Name, levelStr, card.Elixir)
		}
		flushWriter(w)
		printf("\n")
	}

	printf("Strategy: %s\n\n", topDeck.Strategy)

	// Show other options
	if len(r.AllScores) > 1 {
		printf("╔════════════════════════════════════════════════════════════════════╗\n")
		printf("║                    OTHER DECK OPTIONS                              ║\n")
		printf("╚════════════════════════════════════════════════════════════════════╝\n\n")
		for i, rankedDeck := range r.AllScores[1:] {
			deck := rankedDeck.Deck
			printf("#%d: %s\n", i+2, deck.DeckName)
			printf("    Score: %d/100\n", rankedDeck.Score)
			printf("    Compatibility: %s\n", rankedDeck.Compatibility)
			printf("    Average Elixir: %.1f\n", deck.AverageElixir)
			if len(rankedDeck.Reasons) > 0 {
				printf("    Top reason: %s\n", rankedDeck.Reasons[0])
			}
			printf("\n")
		}
	}
}

func savePlaystyleData(dataDir string, p *analysis.PlaystyleAnalysis, r *analysis.DeckRecommendationResult) error {
	// Create analysis directory if it doesn't exist
	analysisDir := filepath.Join(dataDir, "analysis")
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		return fmt.Errorf("failed to create analysis directory: %w", err)
	}

	// Prepare data for saving
	saveData := struct {
		PlaystyleAnalysis   *analysis.PlaystyleAnalysis        `json:"playstyle_analysis"`
		DeckRecommendations *analysis.DeckRecommendationResult `json:"deck_recommendations,omitempty"`
	}{
		PlaystyleAnalysis: p,
	}

	if r != nil {
		saveData.DeckRecommendations = r
	}

	// Save as JSON
	filename := filepath.Join(analysisDir, fmt.Sprintf("playstyle_%s.json", p.PlayerTag))
	data, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal playstyle data: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("failed to write playstyle file: %w", err)
	}

	return nil
}

// addPlayerCommand creates the player command
func addPlayerCommand() *cli.Command {
	return &cli.Command{
		Name:  "player",
		Usage: "Get player information",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "tag",
				Aliases:  []string{"p"},
				Usage:    "Player tag (without #)",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "chests",
				Usage: "Show upcoming chests",
			},
			&cli.BoolFlag{
				Name:  "save",
				Usage: "Save player data to file",
			},
			&cli.BoolFlag{
				Name:  "export-csv",
				Usage: "Export player data to CSV",
			},
		},
		Action: playerCommand,
	}
}

// addCardsCommand creates the cards command
func addCardsCommand() *cli.Command {
	return &cli.Command{
		Name:  "cards",
		Usage: "Get card database",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "export-csv",
				Usage: "Export card database to CSV",
			},
		},
		Action: cardsCommand,
	}
}

// addAnalyzeCommand creates the analyze command
func addAnalyzeCommand() *cli.Command {
	return &cli.Command{
		Name:  "analyze",
		Usage: "Analyze player card collection and upgrade priorities",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "tag",
				Aliases:  []string{"p"},
				Usage:    "Player tag (without #)",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "include-max-level",
				Usage: "Include max level cards in analysis",
			},
			&cli.Float64Flag{
				Name:  "min-priority-score",
				Value: 30.0,
				Usage: "Minimum priority score for upgrade recommendations",
			},
			&cli.StringSliceFlag{
				Name:  "focus-rarities",
				Usage: "Focus on specific rarities (Common, Rare, Epic, Legendary, Champion)",
			},
			&cli.StringSliceFlag{
				Name:  "exclude-cards",
				Usage: "Exclude specific cards from recommendations",
			},
			&cli.BoolFlag{
				Name:  "prioritize-win-cons",
				Value: true,
				Usage: "Boost priority for win condition cards",
			},
			&cli.IntFlag{
				Name:  "top-n",
				Value: 15,
				Usage: "Show top N upgrade priorities",
			},
			&cli.BoolFlag{
				Name:  "save",
				Usage: "Save analysis to JSON file",
			},
			&cli.BoolFlag{
				Name:  "export-csv",
				Usage: "Export analysis to CSV",
			},
		},
		Action: analyzeCommand,
	}
}

// addPlaystyleCommand creates the playstyle command
func addPlaystyleCommand() *cli.Command {
	return &cli.Command{
		Name:  "playstyle",
		Usage: "Analyze player's playstyle and recommend decks",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "tag",
				Aliases:  []string{"p"},
				Usage:    "Player tag (without #)",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "recommend-decks",
				Usage: "Include deck recommendations based on playstyle",
			},
			&cli.BoolFlag{
				Name:  "save",
				Usage: "Save analysis to JSON file",
			},
		},
		Action: playstyleCommand,
	}
}
