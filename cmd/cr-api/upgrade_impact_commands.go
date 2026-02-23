package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/storage"
	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/urfave/cli/v3"
)

// addUpgradeImpactCommands adds upgrade impact analysis commands to the CLI
func addUpgradeImpactCommands() *cli.Command {
	return &cli.Command{
		Name:    "upgrade-impact",
		Aliases: []string{"ui"},
		Usage:   "Analyze which card upgrades have the biggest impact on deck viability",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "tag",
				Aliases:  []string{"p"},
				Usage:    "Player tag (without #)",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "top",
				Value: 10,
				Usage: "Number of top impact cards to show",
			},
			&cli.Float64Flag{
				Name:  "viability-threshold",
				Value: 0.75,
				Usage: "Minimum deck score to be considered viable (0.0-1.0)",
			},
			&cli.BoolFlag{
				Name:  "include-max-level",
				Usage: "Include already maxed cards in analysis",
			},
			&cli.StringSliceFlag{
				Name:  "focus-rarities",
				Usage: "Filter to specific rarities (Common, Rare, Epic, Legendary, Champion)",
			},
			&cli.StringSliceFlag{
				Name:  "exclude-cards",
				Usage: "Cards to exclude from analysis",
			},
			&cli.BoolFlag{
				Name:  "show-all",
				Usage: "Show full analysis including all cards",
			},
			&cli.BoolFlag{
				Name:  "show-unlock-tree",
				Usage: "Show archetype unlock tree",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output in JSON format",
			},
			&cli.BoolFlag{
				Name:  "save",
				Usage: "Save analysis to file",
			},
			&cli.BoolFlag{
				Name:  "use-combat-stats",
				Usage: "Include combat stats (DPS/HP) in impact scoring",
			},
			&cli.StringFlag{
				Name:  "archetypes-file",
				Usage: "Path to custom archetypes JSON file (uses embedded defaults if empty)",
			},
		},
		Action: upgradeImpactCommand,
	}
}

func upgradeImpactCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	topN := cmd.Int("top")
	viabilityThreshold := cmd.Float64("viability-threshold")
	includeMaxLevel := cmd.Bool("include-max-level")
	focusRarities := cmd.StringSlice("focus-rarities")
	excludeCards := cmd.StringSlice("exclude-cards")
	showAll := cmd.Bool("show-all")
	showUnlockTree := cmd.Bool("show-unlock-tree")
	jsonOutput := cmd.Bool("json")
	saveData := cmd.Bool("save")
	useCombatStats := cmd.Bool("use-combat-stats")
	archetypesFile := cmd.String("archetypes-file")
	verbose := cmd.Bool("verbose")
	dataDir := cmd.String("data-dir")

	client, err := requireAPIClient(cmd, apiClientOptions{})
	if err != nil {
		return err
	}

	if verbose {
		printf("Analyzing upgrade impact for player %s...\n", tag)
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

	// Perform card collection analysis first
	analysisOptions := analysis.DefaultAnalysisOptions()
	cardAnalysis, err := analysis.AnalyzeCardCollection(player, analysisOptions)
	if err != nil {
		return fmt.Errorf("failed to analyze card collection: %w", err)
	}

	// Configure upgrade impact options
	impactOptions := analysis.UpgradeImpactOptions{
		ViabilityThreshold: viabilityThreshold,
		TopN:               topN,
		IncludeMaxLevel:    includeMaxLevel,
		FocusRarities:      focusRarities,
		ExcludeCards:       excludeCards,
		UseCombatStats:     useCombatStats,
	}

	// Create analyzer and run analysis (now returns error)
	analyzer, err := analysis.NewUpgradeImpactAnalyzer(dataDir, archetypesFile, impactOptions)
	if err != nil {
		return fmt.Errorf("failed to create analyzer: %w", err)
	}
	impactAnalysis, err := analyzer.AnalyzeUpgradeImpact(cardAnalysis)
	if err != nil {
		return fmt.Errorf("failed to analyze upgrade impact: %w", err)
	}

	// Output results
	if jsonOutput {
		return outputUpgradeImpactJSON(impactAnalysis)
	}

	displayUpgradeImpactAnalysis(impactAnalysis, showAll, showUnlockTree)

	// Save if requested
	if saveData {
		if savedPath, err := saveUpgradeImpactAnalysis(dataDir, impactAnalysis); err != nil {
			printf("Warning: Failed to save analysis: %v\n", err)
		} else {
			printf("\nAnalysis saved to: %s\n", savedPath)
		}
	}

	return nil
}

// Formatting helpers for upgrade impact display
func formatGoldDisplay(gold int) string {
	if gold < 1000 {
		return fmt.Sprintf("%d", gold)
	}
	return fmt.Sprintf("%dk", gold/1000)
}

func getUnlockStatusSymbol(viability string) string {
	switch viability {
	case "viable":
		return "[OK]"
	case "marginal":
		return "[~]"
	case "blocked":
		return "[X]"
	default:
		return "?"
	}
}

func getKeyCardMarker(isKeyCard bool) string {
	if isKeyCard {
		return " [KEY]"
	}
	return ""
}

func formatScoreChange(delta float64) string {
	if delta > 0 {
		return fmt.Sprintf("+%.2f", delta)
	}
	return fmt.Sprintf("%.2f", delta)
}

// Section display functions
func displayUpgradeImpactHeader(analysis *analysis.UpgradeImpactAnalysis) {
	printf("\n")
	printf("============================================================================\n")
	printf("                    UPGRADE IMPACT ANALYSIS                                 \n")
	printf("============================================================================\n\n")
	printf("Player: %s (%s)\n", analysis.PlayerName, analysis.PlayerTag)
	printf("Analysis Time: %s\n\n", analysis.AnalysisTime.Format("2006-01-02 15:04:05"))
}

func displayUpgradeImpactSummary(summary analysis.ImpactSummary) {
	printf("Summary\n")
	printf("-------\n")
	printf("Cards Analyzed:     %d\n", summary.TotalCardsAnalyzed)
	printf("Key Cards Found:    %d\n", summary.KeyCardsIdentified)
	printf("Average Impact:     %.2f\n", summary.AvgImpactScore)
	printf("Max Impact Score:   %.2f\n", summary.MaxImpactScore)
	printf("Viable Deck Count:  %d\n", summary.TotalViableDecks)
	printf("Potential Unlocks:  %d\n\n", summary.PotentialUnlocks)
}

func displayTopImpactCards(impacts []analysis.CardUpgradeImpact, showAll bool) {
	printf("Most Impactful Upgrades\n")
	printf("-----------------------\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "#\tCard\tLevel\tRarity\tImpact\tGold\tValue/1k\tUnlocks\n")
	fprintf(w, "-\t----\t-----\t------\t------\t----\t--------\t-------\n")

	for i, impact := range impacts {
		fprintf(w, "%d\t%s%s\t%d->%d\t%s\t%.1f\t%s\t%.2f\t%d\n",
			i+1,
			impact.CardName,
			getKeyCardMarker(impact.IsKeyCard),
			impact.CurrentLevel,
			impact.UpgradedLevel,
			impact.Rarity,
			impact.ImpactScore,
			formatGoldDisplay(impact.GoldCost),
			impact.ValuePerGold,
			impact.UnlockPotential,
		)

		if !showAll && i >= 9 {
			break
		}
	}
	flushWriter(w)
}

func displayKeyCards(keyCards []analysis.KeyCardInfo) {
	printf("\nKey Cards (Unlock Multiple Archetypes)\n")
	printf("--------------------------------------\n")

	for i, keyCard := range keyCards {
		if i >= 5 {
			printf("   ... and %d more\n", len(keyCards)-5)
			break
		}
		printf("  %s (Level %d, %s)\n", keyCard.CardName, keyCard.CurrentLevel, keyCard.Rarity)
		if len(keyCard.UnlockedArchetypes) > 0 {
			printf("    Unlocks: %v\n", keyCard.UnlockedArchetypes)
		}
		printf("    Potential Deck Unlocks: %d\n", keyCard.DeckUnlockCount)
	}
}

func displayArchetypeUnlockTree(unlocks []analysis.ArchetypeUnlockInfo) {
	printf("\nArchetype Unlock Tree\n")
	printf("---------------------\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fprintf(w, "Archetype\tStatus\tPriority Upgrade\tEst. Gold\n")
	fprintf(w, "---------\t------\t----------------\t---------\n")

	for _, unlock := range unlocks {
		goldStr := "-"
		if unlock.EstimatedGold > 0 {
			goldStr = formatGoldDisplay(unlock.EstimatedGold)
		}

		fprintf(w, "%s\t%s\t%s\t%s\n",
			unlock.Archetype,
			getUnlockStatusSymbol(unlock.CurrentViability),
			unlock.PriorityUpgrade,
			goldStr,
		)
	}
	flushWriter(w)
}

func displayDetailedDeckImpact(topImpacts []analysis.CardUpgradeImpact) {
	printf("\nDetailed Deck Impact (Top 3 Upgrades)\n")
	printf("-------------------------------------\n")

	for i, impact := range topImpacts {
		if i >= 3 {
			break
		}

		printf("\n%d. %s (%s, Level %d -> %d)\n",
			i+1, impact.CardName, impact.Rarity, impact.CurrentLevel, impact.UpgradedLevel)
		printf("   Impact Score: %.2f | Gold Cost: %d | Value: %.2f per 1k gold\n",
			impact.ImpactScore, impact.GoldCost, impact.ValuePerGold)

		if len(impact.AffectedDecks) > 0 {
			printf("   Affected Decks:\n")
			for _, deck := range impact.AffectedDecks {
				viableMarker := ""
				if deck.BecomesViable {
					viableMarker = " [UNLOCKS!]"
				}

				printf("     - %s: %.2f -> %.2f (%s)%s\n",
					deck.DeckName, deck.CurrentScore, deck.ProjectedScore,
					formatScoreChange(deck.ScoreDelta), viableMarker)
			}
		}

		if len(impact.UnlocksArchetypes) > 0 {
			printf("   Unlocks Archetypes: %v\n", impact.UnlocksArchetypes)
		}
	}
}

func displayUpgradeImpactRecommendations(topImpacts []analysis.CardUpgradeImpact) {
	printf("\nRecommendations\n")
	printf("---------------\n")
	if len(topImpacts) == 0 {
		return
	}

	top := topImpacts[0]
	printf("Best upgrade: %s (Level %d -> %d)\n", top.CardName, top.CurrentLevel, top.UpgradedLevel)
	printf("  Impact: %.1f points | Gold: %d | Unlocks %d deck(s)\n",
		top.ImpactScore, top.GoldCost, top.UnlockPotential)

	if len(topImpacts) > 1 {
		second := topImpacts[1]
		if second.ValuePerGold > top.ValuePerGold*1.5 {
			printf("\nValue alternative: %s (%.2f impact per 1k gold vs %.2f)\n",
				second.CardName, second.ValuePerGold, top.ValuePerGold)
		}
	}
}

func displayUpgradeImpactAnalysis(impactAnalysis *analysis.UpgradeImpactAnalysis, showAll, showUnlockTree bool) {
	displayUpgradeImpactHeader(impactAnalysis)
	displayUpgradeImpactSummary(impactAnalysis.Summary)

	impacts := impactAnalysis.TopImpacts
	if showAll && len(impactAnalysis.CardImpacts) > len(impacts) {
		impacts = impactAnalysis.CardImpacts
	}
	displayTopImpactCards(impacts, showAll)

	if len(impactAnalysis.KeyCards) > 0 {
		displayKeyCards(impactAnalysis.KeyCards)
	}

	if showUnlockTree && len(impactAnalysis.UnlockTree) > 0 {
		displayArchetypeUnlockTree(impactAnalysis.UnlockTree)
	}

	if showAll && len(impactAnalysis.TopImpacts) > 0 {
		displayDetailedDeckImpact(impactAnalysis.TopImpacts)
	}

	displayUpgradeImpactRecommendations(impactAnalysis.TopImpacts)
}

// outputUpgradeImpactJSON prints upgrade impact analysis in pretty JSON format.
// outputUpgradeImpactJSON prints upgrade impact analysis in pretty JSON format.
func outputUpgradeImpactJSON(impactAnalysis *analysis.UpgradeImpactAnalysis) error {
	data, err := json.MarshalIndent(impactAnalysis, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// saveUpgradeImpactAnalysis writes upgrade impact analysis and returns the final file path.
func saveUpgradeImpactAnalysis(dataDir string, impactAnalysis *analysis.UpgradeImpactAnalysis) (string, error) {
	analysisDir := filepath.Join(dataDir, "analysis")

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(analysisDir, fmt.Sprintf("upgrade_impact_%s_%s.json", impactAnalysis.PlayerTag, timestamp))

	if err := storage.WriteJSON(filename, impactAnalysis); err != nil {
		return "", fmt.Errorf("failed to write analysis file: %w", err)
	}

	return filename, nil
}
