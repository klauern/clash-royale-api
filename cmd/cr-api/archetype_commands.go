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

	"github.com/klauer/clash-royale-api/go/internal/exporter/csv"
	"github.com/klauer/clash-royale-api/go/pkg/archetypes"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/urfave/cli/v3"
)

// addArchetypeCommands adds archetype analysis commands to the CLI
func addArchetypeCommands() *cli.Command {
	return &cli.Command{
		Name:  "archetypes",
		Usage: "Analyze deck archetypes and upgrade costs across different playstyles",
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
		fmt.Printf("%s (%.1f elixir)\n", strings.Title(string(arch.Archetype)), arch.AvgElixir)
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
		return "None"
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
		return "None"
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
	if err := os.MkdirAll(dir, 0755); err != nil {
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

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// convertPlayerToAnalysis converts player data to CardAnalysis format
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
