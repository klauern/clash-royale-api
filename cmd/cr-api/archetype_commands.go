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

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/klauer/clash-royale-api/go/internal/exporter/csv"
	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/archetypes"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/urfave/cli/v3"
)

const noCardsLabel = "None"

// addArchetypeCommands adds archetype analysis commands to the CLI
func addArchetypeCommands() *cli.Command {
	return &cli.Command{
		Name:  "archetypes",
		Usage: "Analyze deck archetypes and upgrade costs across different playstyles",
		Commands: []*cli.Command{
			addArchetypeVarietyCommand(),
			addArchetypeDetectCommand(),
		},
	}
}

// addArchetypeVarietyCommand adds the variety analysis command (original behavior)
func addArchetypeVarietyCommand() *cli.Command {
	return &cli.Command{
		Name:  "variety",
		Usage: "Analyze archetype variety across 8 predefined playstyles",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "tag",
				Aliases:  []string{"p"},
				Usage:    "Player tag (without #)",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "target-level",
				Value: 11,
				Usage: "Target competitive card level for archetype viability",
			},
			&cli.StringFlag{
				Name:  "sort-by",
				Value: "distance",
				Usage: "Sort results by: distance (viability), cards_needed (investment), avg_level (current strength)",
			},
			&cli.BoolFlag{
				Name:  "export-csv",
				Usage: "Export archetype analysis to CSV",
			},
			&cli.BoolFlag{
				Name:  "save",
				Usage: "Save analysis to JSON file",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Show detailed deck composition and upgrade information per archetype",
			},
		},
		Action: analyzeArchetypesCommand,
	}
}

func analyzeArchetypesCommand(ctx context.Context, cmd *cli.Command) error {
	// Get CLI flags
	playerTag := cmd.String("tag")
	targetLevel := cmd.Int("target-level")
	sortBy := cmd.String("sort-by")
	exportCSV := cmd.Bool("export-csv")
	saveJSON := cmd.Bool("save")
	verbose := cmd.Bool("verbose")
	apiToken := cmd.String("api-token")
	dataDir := cmd.String("data-dir")

	// Validate API token
	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	// Validate target level
	if targetLevel < 1 || targetLevel > 14 {
		return fmt.Errorf("target level must be between 1 and 14")
	}

	// Validate sort-by option
	validSortBy := map[string]archetypes.SortBy{
		"distance":     archetypes.SortByDistance,
		"cards_needed": archetypes.SortByCardsNeeded,
		"avg_level":    archetypes.SortByAvgLevel,
	}
	sortOption, ok := validSortBy[sortBy]
	if !ok {
		return fmt.Errorf("invalid sort-by option: %s (must be distance, cards_needed, or avg_level)", sortBy)
	}

	// Initialize API client
	client := clashroyale.NewClient(apiToken)

	if verbose {
		fmt.Printf("Fetching player data for tag: %s\n", playerTag)
	}

	// Fetch player data
	player, err := client.GetPlayer(playerTag)
	if err != nil {
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	if verbose {
		fmt.Printf("Player: %s (%s)\n", player.Name, player.Tag)
		fmt.Printf("Building archetype decks from %d cards...\n\n", len(player.Cards))
	}

	// Load or generate card analysis
	// Try to load existing analysis first
	builder := deck.NewBuilder(dataDir)
	analysis, err := builder.LoadLatestAnalysis(playerTag, filepath.Join(dataDir, "analysis"))
	if err != nil {
		if verbose {
			fmt.Printf("No existing analysis found, generating from player data...\n")
		}
		// Convert player cards to analysis format
		analysis = convertPlayerToAnalysis(player)
	}

	// Perform archetype analysis
	analyzer := archetypes.NewAnalyzer(dataDir)
	result, err := analyzer.AnalyzeArchetypes(
		player.Tag,
		player.Name,
		*analysis,
		targetLevel,
	)
	if err != nil {
		return fmt.Errorf("archetype analysis failed: %w", err)
	}

	// Sort results
	result.SortBy(sortOption)

	// Display results
	displayArchetypeComparison(result, verbose)

	// Save JSON if requested
	if saveJSON {
		if err := saveArchetypeAnalysis(result, dataDir); err != nil {
			fmt.Printf("\nWarning: Failed to save analysis: %v\n", err)
		} else {
			fmt.Printf("\n✓ Analysis saved to: %s/archetypes/\n", dataDir)
		}
	}

	// Export CSV if requested
	if exportCSV {
		exporter := csv.NewArchetypeExporter()
		if err := exporter.Export(dataDir, result); err != nil {
			fmt.Printf("\nWarning: CSV export failed: %v\n", err)
		} else {
			fmt.Printf("\n✓ CSV exported to: %s/csv/archetypes/\n", dataDir)
		}

		// Also export detailed upgrade breakdown
		detailsExporter := csv.NewArchetypeDetailsExporter()
		if err := detailsExporter.Export(dataDir, result); err != nil {
			fmt.Printf("Warning: CSV details export failed: %v\n", err)
		}
	}

	return nil
}

// displayArchetypeComparison shows formatted comparison table
func displayArchetypeComparison(result *archetypes.ArchetypeAnalysisResult, verbose bool) {
	// Use tabwriter with proper spacing for aligned columns
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)

	// Header - print directly to stdout (not through tabwriter)
	fmt.Printf("\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("  ARCHETYPE VARIETY ANALYSIS - %s (##%s)\n", result.PlayerName, strings.TrimPrefix(result.PlayerTag, "#"))
	fmt.Printf("  Target Level: %d\n", result.TargetLevel)
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("\n")

	// Table headers through tabwriter for proper alignment
	fmt.Fprintf(w, "Archetype\tAvg Elixir\tCurrent Lvl\tCards Needed\tGold Needed\tGems Needed\tDistance\n")
	fmt.Fprintf(w, "────────\t──────────\t────────────\t─────────────\t────────────\t────────────\t──────────\t\n")

	// Data rows - table only
	for _, arch := range result.Archetypes {
		fmt.Fprintf(w, "%s\t%.1f\t%.1f\t%s\t%s\t%s\t%.2f\n",
			arch.Archetype,
			arch.AvgElixir,
			arch.CurrentAvgLevel,
			formatNumber(arch.CardsNeeded),
			formatNumber(arch.GoldNeeded),
			formatNumber(arch.GemsNeeded),
			arch.DistanceMetric,
		)
	}

	// Flush the tabwriter to print the table
	w.Flush()

	// Deck compositions section
	fmt.Printf("\n")
	fmt.Printf("Archetype Deck Breakdown:\n")
	fmt.Printf("─────────────────────────\n")

	for _, arch := range result.Archetypes {
		fmt.Printf("\n")
		fmt.Printf("%s (%.1f elixir)\n", cases.Title(language.English).String(string(arch.Archetype)), arch.AvgElixir)
		fmt.Printf("├─ Win Condition: %s\n", findCardByRole(arch.DeckDetail, string(deck.RoleWinCondition)))
		fmt.Printf("├─ Spells: %s\n", findSpells(arch.DeckDetail))
		fmt.Printf("├─ Support: %s\n", findCardsByRole(arch.DeckDetail, string(deck.RoleSupport)))
		fmt.Printf("├─ Cycle: %s\n", findCardsByRole(arch.DeckDetail, string(deck.RoleCycle)))
		fmt.Printf("└─ Building: %s\n", findCardByRole(arch.DeckDetail, string(deck.RoleBuilding)))
	}

	// Footer with interpretation guide
	fmt.Printf("\n")
	fmt.Printf("Interpretation Guide:\n")
	fmt.Printf("  • Distance: 0.00 (perfect match) to 1.00 (far from target)\n")
	fmt.Printf("  • Lower distance = more viable archetype for your collection\n")
	fmt.Printf("  • Cards/Gold/Gems Needed = upgrade investment to reach target level\n")
	fmt.Printf("  • Gem cost: ~1 gem per 17 gold (based on shop conversion rates)\n")
	fmt.Printf("\n")
}

// Helper functions for card role breakdown
func findCardByRole(cards []deck.CardDetail, role string) string {
	var matches []string
	for _, card := range cards {
		if card.Role == role {
			matches = append(matches, card.Name)
		}
	}

	if len(matches) == 0 {
		return noCardsLabel
	}
	return strings.Join(matches, ", ")
}

func findCardsByRole(cards []deck.CardDetail, role string) string {
	var matches []string
	for _, card := range cards {
		if card.Role == role {
			matches = append(matches, card.Name)
		}
	}

	if len(matches) == 0 {
		return noCardsLabel
	}
	return strings.Join(matches, ", ")
}

func findSpells(cards []deck.CardDetail) string {
	var bigSpells []string
	var smallSpells []string

	for _, card := range cards {
		if card.Role == string(deck.RoleSpellBig) {
			bigSpells = append(bigSpells, card.Name)
		} else if card.Role == string(deck.RoleSpellSmall) {
			smallSpells = append(smallSpells, card.Name)
		}
	}

	allSpells := append(bigSpells, smallSpells...)
	if len(allSpells) == 0 {
		return "None"
	}
	return strings.Join(allSpells, ", ")
}

// formatNumber formats numbers with commas for readability
func formatNumber(n int) string {
	if n == 0 {
		return "0"
	}

	// Convert to string and add commas
	s := fmt.Sprintf("%d", n)
	var result strings.Builder
	for i, digit := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}
	return result.String()
}

// saveArchetypeAnalysis saves analysis to JSON file
func saveArchetypeAnalysis(result *archetypes.ArchetypeAnalysisResult, dataDir string) error {
	dir := filepath.Join(dataDir, "archetypes")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create archetypes directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	cleanTag := strings.TrimPrefix(result.PlayerTag, "#")
	filename := fmt.Sprintf("%s_archetype_analysis_%s.json", timestamp, cleanTag)
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// convertPlayerToAnalysis converts player data to CardAnalysis format (for deck package)
func convertPlayerToAnalysis(player *clashroyale.Player) *deck.CardAnalysis {
	analysis := &deck.CardAnalysis{
		CardLevels:   make(map[string]deck.CardLevelData),
		AnalysisTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	for _, card := range player.Cards {
		analysis.CardLevels[card.Name] = deck.CardLevelData{
			Level:             card.Level,
			MaxLevel:          card.MaxLevel,
			Rarity:            card.Rarity,
			Elixir:            card.ElixirCost,
			MaxEvolutionLevel: card.MaxEvolutionLevel,
		}
	}

	return analysis
}

// convertPlayerToAnalysisPackage converts player data to analysis.CardAnalysis format
func convertPlayerToAnalysisPackage(player *clashroyale.Player) *analysis.CardAnalysis {
	cardAnalysis := &analysis.CardAnalysis{
		PlayerTag:    player.Tag,
		PlayerName:   player.Name,
		AnalysisTime: time.Now(),
		TotalCards:   len(player.Cards),
		CardLevels:   make(map[string]analysis.CardLevelInfo),
	}

	for _, card := range player.Cards {
		cardAnalysis.CardLevels[card.Name] = analysis.CardLevelInfo{
			Name:              card.Name,
			ID:                card.ID,
			Level:             card.Level,
			MaxLevel:          card.MaxLevel,
			EvolutionLevel:    card.EvolutionLevel,
			Rarity:            card.Rarity,
			Elixir:            card.ElixirCost,
			CardCount:         card.Count,
			CardsToNext:       0, // Not available from API
			IsMaxLevel:        card.Level >= card.MaxLevel,
			MaxEvolutionLevel: card.MaxEvolutionLevel,
		}
	}

	return cardAnalysis
}

// addArchetypeDetectCommand adds the dynamic archetype detection command
func addArchetypeDetectCommand() *cli.Command {
	return &cli.Command{
		Name:  "detect",
		Usage: "Dynamically detect viable archetypes based on player's card collection",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "tag",
				Aliases:  []string{"p"},
				Usage:    "Player tag (without #)",
				Required: true,
			},
			&cli.Float64Flag{
				Name:  "min-viability",
				Value: 0,
				Usage: "Minimum viability score to display (0-100)",
			},
			&cli.BoolFlag{
				Name:  "show-strategies",
				Value: true,
				Usage: "Display recommended deck building strategies",
			},
			&cli.BoolFlag{
				Name:  "show-upgrades",
				Value: true,
				Usage: "Display upgrade recommendations",
			},
			&cli.BoolFlag{
				Name:  "save",
				Usage: "Save analysis to JSON file",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output in JSON format",
			},
			&cli.StringFlag{
				Name:  "archetypes-file",
				Usage: "Path to custom archetypes JSON file (optional, uses embedded defaults)",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Show detailed analysis information",
			},
		},
		Action: detectArchetypesCommand,
	}
}

func detectArchetypesCommand(ctx context.Context, cmd *cli.Command) error {
	// Get CLI flags
	playerTag := cmd.String("tag")
	minViability := cmd.Float64("min-viability")
	showStrategies := cmd.Bool("show-strategies")
	showUpgrades := cmd.Bool("show-upgrades")
	saveJSON := cmd.Bool("save")
	jsonOutput := cmd.Bool("json")
	archetypesFile := cmd.String("archetypes-file")
	verbose := cmd.Bool("verbose")
	apiToken := cmd.String("api-token")
	dataDir := cmd.String("data-dir")

	// Validate API token
	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	// Validate min-viability
	if minViability < 0 || minViability > 100 {
		return fmt.Errorf("min-viability must be between 0 and 100")
	}

	// Initialize API client
	client := clashroyale.NewClient(apiToken)

	if verbose {
		fmt.Printf("Fetching player data for tag: %s\n", playerTag)
	}

	// Fetch player data
	player, err := client.GetPlayer(playerTag)
	if err != nil {
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	if verbose {
		fmt.Printf("Player: %s (%s)\n", player.Name, player.Tag)
		fmt.Printf("Analyzing %d cards across archetype templates...\n\n", len(player.Cards))
	}

	// Convert player data to analysis format
	cardAnalysis := convertPlayerToAnalysisPackage(player)

	// Create synergy database and strategy provider adapters
	synergyDB := deck.NewSynergyDatabase()
	strategyProvider := &deckStrategyProvider{}

	// Create dynamic archetype detector
	detector, err := analysis.NewDynamicArchetypeDetector(dataDir, archetypesFile, &synergyDBAdapter{db: synergyDB}, strategyProvider)
	if err != nil {
		return fmt.Errorf("failed to create archetype detector: %w", err)
	}

	// Configure detection options
	options := analysis.DetectionOptions{
		MinViability:         minViability,
		IncludeStrategies:    showStrategies,
		IncludeUpgrades:      showUpgrades,
		TopUpgradesPerArch:   3,
		TopCrossArchUpgrades: 10,
		ExcludeCards:         []string{},
	}

	// Detect archetypes
	result, err := detector.DetectArchetypes(cardAnalysis, options)
	if err != nil {
		return fmt.Errorf("archetype detection failed: %w", err)
	}

	// Output results
	if jsonOutput {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal analysis: %w", err)
		}
		fmt.Println(string(data))
	} else {
		displayDetectedArchetypes(result, showStrategies, showUpgrades, verbose)
	}

	// Save JSON if requested
	if saveJSON {
		if err := saveDynamicArchetypeAnalysis(result, dataDir); err != nil {
			fmt.Printf("\nWarning: Failed to save analysis: %v\n", err)
		} else {
			fmt.Printf("\n✓ Analysis saved to: %s/archetype_analysis/\n", dataDir)
		}
	}

	return nil
}

// displayDetectedArchetypes shows formatted archetype detection results
func displayDetectedArchetypes(result *analysis.DynamicArchetypeAnalysis, showStrategies, showUpgrades, verbose bool) {
	// Header
	fmt.Printf("\n")
	fmt.Printf("════════════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("               DYNAMIC ARCHETYPE DETECTION\n")
	fmt.Printf("════════════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("Player: %s (%s)\n", result.PlayerName, result.PlayerTag)
	fmt.Printf("Analysis Time: %s\n", result.AnalysisTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("════════════════════════════════════════════════════════════════════════════════\n\n")

	// Display archetypes by tier
	tiers := []struct {
		name       string
		archetypes []string
		symbol     string
	}{
		{"OPTIMAL ARCHETYPES (90-100)", result.OptimalArchetypes, "🏆"},
		{"COMPETITIVE ARCHETYPES (75-89)", result.CompetitiveArchetypes, "⚔️"},
		{"PLAYABLE ARCHETYPES (60-74)", result.PlayableArchetypes, "✓"},
		{"BLOCKED ARCHETYPES (<60)", result.BlockedArchetypes, "✗"},
	}

	for _, tier := range tiers {
		if len(tier.archetypes) == 0 {
			continue
		}

		fmt.Printf("%s %s\n", tier.symbol, tier.name)
		fmt.Printf("────────────────────────────────────────────────────────────────────────────────\n")

		// Create tabwriter for aligned columns
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)
		fmt.Fprintf(w, "Archetype\tScore\tWin Con Lvl\tSupport Avg\tSynergy\n")

		// Find and display each archetype in this tier
		for _, archName := range tier.archetypes {
			for _, arch := range result.DetectedArchetypes {
				if arch.Name == archName {
					fmt.Fprintf(w, "%s\t%.1f\t%d/%d\t%.1f\t%.1f\n",
						arch.Name,
						arch.ViabilityScore,
						arch.WinConditionLevel,
						arch.WinConditionMax,
						arch.SupportCardsAvg,
						arch.SynergyScore,
					)

					if showStrategies && len(arch.RecommendedStrategies) > 0 {
						fmt.Fprintf(w, "  └─ Strategies:\t%s\t\t\t\n", formatStrategies(arch.RecommendedStrategies))
					}

					break
				}
			}
		}
		w.Flush()
		fmt.Printf("\n")
	}

	// Display top cross-archetype upgrades
	if showUpgrades && len(result.TopUpgradeImpacts) > 0 {
		fmt.Printf("═══════════════════════════════════════════════════════════════════════════════\n")
		fmt.Printf("TOP UPGRADE RECOMMENDATIONS (Cross-Archetype Impact)\n")
		fmt.Printf("═══════════════════════════════════════════════════════════════════════════════\n\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)
		fmt.Fprintf(w, "Card\tLevel\tGold\tArchetypes Affected\tTotal Impact\tUnlocks\n")
		fmt.Fprintf(w, "────\t─────\t────\t───────────────────\t────────────\t───────\n")

		for i, upgrade := range result.TopUpgradeImpacts {
			if i >= 10 {
				break // Limit to top 10
			}
			fmt.Fprintf(w, "%s\t%d\t%s\t%d\t+%.1f\t%d\n",
				upgrade.CardName,
				upgrade.CurrentLevel,
				formatNumber(upgrade.GoldCost),
				len(upgrade.AffectedArchetypes),
				upgrade.TotalViabilityGain,
				upgrade.ArchetypesUnlocked,
			)
		}
		w.Flush()
		fmt.Printf("\n")
	}

	// Footer
	fmt.Printf("Interpretation Guide:\n")
	fmt.Printf("  • Viability Score: 0-100 based on card levels, synergies, and completeness\n")
	fmt.Printf("  • Optimal (90+): Tournament-ready, high-level cards\n")
	fmt.Printf("  • Competitive (75-89): Ladder-viable, solid for ranked play\n")
	fmt.Printf("  • Playable (60-74): Functional but needs upgrades\n")
	fmt.Printf("  • Blocked (<60): Missing cards or severely underleveled\n")
	fmt.Printf("\n")
}

// formatStrategies formats strategy recommendations for display
func formatStrategies(strategies []analysis.StrategyRecommendation) string {
	if len(strategies) == 0 {
		return "None"
	}

	var parts []string
	for i, strategy := range strategies {
		if i >= 3 {
			break // Limit to top 3
		}
		parts = append(parts, fmt.Sprintf("%s (%.0f%%)", strategy.Strategy, strategy.CompatibilityScore))
	}

	return strings.Join(parts, ", ")
}

// saveDynamicArchetypeAnalysis saves analysis to JSON file
func saveDynamicArchetypeAnalysis(result *analysis.DynamicArchetypeAnalysis, dataDir string) error {
	dir := filepath.Join(dataDir, "archetype_analysis")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	cleanTag := strings.TrimPrefix(result.PlayerTag, "#")
	filename := fmt.Sprintf("%s_dynamic_archetype_analysis_%s.json", timestamp, cleanTag)
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Adapter types to bridge deck package to analysis interfaces

// synergyDBAdapter adapts deck.SynergyDatabase to analysis.SynergyDB interface
type synergyDBAdapter struct {
	db *deck.SynergyDatabase
}

func (a *synergyDBAdapter) GetSynergy(card1, card2 string) float64 {
	return a.db.GetSynergy(card1, card2)
}

func (a *synergyDBAdapter) GetSynergyPair(card1, card2 string) *analysis.SynergyPair {
	pair := a.db.GetSynergyPair(card1, card2)
	if pair == nil {
		return nil
	}
	return &analysis.SynergyPair{
		Card1:       pair.Card1,
		Card2:       pair.Card2,
		SynergyType: string(pair.SynergyType),
		Score:       pair.Score,
		Description: pair.Description,
	}
}

// deckStrategyProvider adapts deck package strategy functions to analysis.StrategyProvider interface
type deckStrategyProvider struct{}

func (p *deckStrategyProvider) GetConfig(strategy analysis.Strategy) analysis.StrategyConfig {
	// Convert analysis.Strategy to deck.Strategy
	deckStrategy := deck.Strategy(strategy)
	config := deck.GetStrategyConfig(deckStrategy)

	return analysis.StrategyConfig{
		TargetElixirMin:   config.TargetElixirMin,
		TargetElixirMax:   config.TargetElixirMax,
		ArchetypeAffinity: config.ArchetypeAffinity,
	}
}

func (p *deckStrategyProvider) GetAllStrategies() []analysis.Strategy {
	return []analysis.Strategy{
		analysis.Strategy(deck.StrategyBalanced),
		analysis.Strategy(deck.StrategyAggro),
		analysis.Strategy(deck.StrategyControl),
		analysis.Strategy(deck.StrategyCycle),
		analysis.Strategy(deck.StrategySplash),
		analysis.Strategy(deck.StrategySpell),
	}
}
