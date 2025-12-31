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
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	dataDir := cmd.String("data-dir")

	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	client := clashroyale.NewClient(apiToken)

	if verbose {
		fmt.Printf("Analyzing upgrade impact for player %s...\n", tag)
	}

	// Get player information
	player, err := client.GetPlayer(tag)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	if verbose {
		fmt.Printf("Player: %s (%s)\n", player.Name, player.Tag)
		fmt.Printf("Analyzing %d cards...\n", len(player.Cards))
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
		if err := saveUpgradeImpactAnalysis(dataDir, impactAnalysis); err != nil {
			fmt.Printf("Warning: Failed to save analysis: %v\n", err)
		} else {
			fmt.Printf("\nAnalysis saved to: %s/analysis/upgrade_impact_%s.json\n", dataDir, impactAnalysis.PlayerTag)
		}
	}

	return nil
}

func displayUpgradeImpactAnalysis(impactAnalysis *analysis.UpgradeImpactAnalysis, showAll, showUnlockTree bool) {
	fmt.Printf("\n")
	fmt.Printf("============================================================================\n")
	fmt.Printf("                    UPGRADE IMPACT ANALYSIS                                 \n")
	fmt.Printf("============================================================================\n\n")

	fmt.Printf("Player: %s (%s)\n", impactAnalysis.PlayerName, impactAnalysis.PlayerTag)
	fmt.Printf("Analysis Time: %s\n\n", impactAnalysis.AnalysisTime.Format("2006-01-02 15:04:05"))

	// Summary
	fmt.Printf("Summary\n")
	fmt.Printf("-------\n")
	fmt.Printf("Cards Analyzed:     %d\n", impactAnalysis.Summary.TotalCardsAnalyzed)
	fmt.Printf("Key Cards Found:    %d\n", impactAnalysis.Summary.KeyCardsIdentified)
	fmt.Printf("Average Impact:     %.2f\n", impactAnalysis.Summary.AvgImpactScore)
	fmt.Printf("Max Impact Score:   %.2f\n", impactAnalysis.Summary.MaxImpactScore)
	fmt.Printf("Viable Deck Count:  %d\n", impactAnalysis.Summary.TotalViableDecks)
	fmt.Printf("Potential Unlocks:  %d\n\n", impactAnalysis.Summary.PotentialUnlocks)

	// Top Impact Cards
	fmt.Printf("Most Impactful Upgrades\n")
	fmt.Printf("-----------------------\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "#\tCard\tLevel\tRarity\tImpact\tGold\tValue/1k\tUnlocks\n")
	fmt.Fprintf(w, "-\t----\t-----\t------\t------\t----\t--------\t-------\n")

	impacts := impactAnalysis.TopImpacts
	if showAll && len(impactAnalysis.CardImpacts) > len(impacts) {
		impacts = impactAnalysis.CardImpacts
	}

	for i, impact := range impacts {
		keyMarker := ""
		if impact.IsKeyCard {
			keyMarker = " [KEY]"
		}

		goldDisplay := fmt.Sprintf("%dk", impact.GoldCost/1000)
		if impact.GoldCost < 1000 {
			goldDisplay = fmt.Sprintf("%d", impact.GoldCost)
		}

		fmt.Fprintf(w, "%d\t%s%s\t%d->%d\t%s\t%.1f\t%s\t%.2f\t%d\n",
			i+1,
			impact.CardName,
			keyMarker,
			impact.CurrentLevel,
			impact.UpgradedLevel,
			impact.Rarity,
			impact.ImpactScore,
			goldDisplay,
			impact.ValuePerGold,
			impact.UnlockPotential,
		)

		// Show top 10 only in compact mode
		if !showAll && i >= 9 {
			break
		}
	}
	w.Flush()

	// Key Cards Section
	if len(impactAnalysis.KeyCards) > 0 {
		fmt.Printf("\nKey Cards (Unlock Multiple Archetypes)\n")
		fmt.Printf("--------------------------------------\n")

		for i, keyCard := range impactAnalysis.KeyCards {
			if i >= 5 {
				fmt.Printf("   ... and %d more\n", len(impactAnalysis.KeyCards)-5)
				break
			}
			fmt.Printf("  %s (Level %d, %s)\n", keyCard.CardName, keyCard.CurrentLevel, keyCard.Rarity)
			if len(keyCard.UnlockedArchetypes) > 0 {
				fmt.Printf("    Unlocks: %v\n", keyCard.UnlockedArchetypes)
			}
			fmt.Printf("    Potential Deck Unlocks: %d\n", keyCard.DeckUnlockCount)
		}
	}

	// Unlock Tree Section
	if showUnlockTree && len(impactAnalysis.UnlockTree) > 0 {
		fmt.Printf("\nArchetype Unlock Tree\n")
		fmt.Printf("---------------------\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "Archetype\tStatus\tPriority Upgrade\tEst. Gold\n")
		fmt.Fprintf(w, "---------\t------\t----------------\t---------\n")

		for _, unlock := range impactAnalysis.UnlockTree {
			statusSymbol := "?"
			switch unlock.CurrentViability {
			case "viable":
				statusSymbol = "[OK]"
			case "marginal":
				statusSymbol = "[~]"
			case "blocked":
				statusSymbol = "[X]"
			}

			goldStr := "-"
			if unlock.EstimatedGold > 0 {
				goldStr = fmt.Sprintf("%dk", unlock.EstimatedGold/1000)
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				unlock.Archetype,
				statusSymbol,
				unlock.PriorityUpgrade,
				goldStr,
			)
		}
		w.Flush()
	}

	// Detailed deck impact for top cards
	if showAll && len(impactAnalysis.TopImpacts) > 0 {
		fmt.Printf("\nDetailed Deck Impact (Top 3 Upgrades)\n")
		fmt.Printf("-------------------------------------\n")

		for i, impact := range impactAnalysis.TopImpacts {
			if i >= 3 {
				break
			}

			fmt.Printf("\n%d. %s (%s, Level %d -> %d)\n",
				i+1, impact.CardName, impact.Rarity, impact.CurrentLevel, impact.UpgradedLevel)
			fmt.Printf("   Impact Score: %.2f | Gold Cost: %d | Value: %.2f per 1k gold\n",
				impact.ImpactScore, impact.GoldCost, impact.ValuePerGold)

			if len(impact.AffectedDecks) > 0 {
				fmt.Printf("   Affected Decks:\n")
				for _, deck := range impact.AffectedDecks {
					scoreChange := ""
					if deck.ScoreDelta > 0 {
						scoreChange = fmt.Sprintf("+%.2f", deck.ScoreDelta)
					} else {
						scoreChange = fmt.Sprintf("%.2f", deck.ScoreDelta)
					}

					viableMarker := ""
					if deck.BecomesViable {
						viableMarker = " [UNLOCKS!]"
					}

					fmt.Printf("     - %s: %.2f -> %.2f (%s)%s\n",
						deck.DeckName, deck.CurrentScore, deck.ProjectedScore, scoreChange, viableMarker)
				}
			}

			if len(impact.UnlocksArchetypes) > 0 {
				fmt.Printf("   Unlocks Archetypes: %v\n", impact.UnlocksArchetypes)
			}
		}
	}

	// Recommendations
	fmt.Printf("\nRecommendations\n")
	fmt.Printf("---------------\n")
	if len(impactAnalysis.TopImpacts) > 0 {
		top := impactAnalysis.TopImpacts[0]
		fmt.Printf("Best upgrade: %s (Level %d -> %d)\n", top.CardName, top.CurrentLevel, top.UpgradedLevel)
		fmt.Printf("  Impact: %.1f points | Gold: %d | Unlocks %d deck(s)\n",
			top.ImpactScore, top.GoldCost, top.UnlockPotential)

		if len(impactAnalysis.TopImpacts) > 1 {
			second := impactAnalysis.TopImpacts[1]
			if second.ValuePerGold > top.ValuePerGold*1.5 {
				fmt.Printf("\nValue alternative: %s (%.2f impact per 1k gold vs %.2f)\n",
					second.CardName, second.ValuePerGold, top.ValuePerGold)
			}
		}
	}
}

func outputUpgradeImpactJSON(impactAnalysis *analysis.UpgradeImpactAnalysis) error {
	data, err := json.MarshalIndent(impactAnalysis, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func saveUpgradeImpactAnalysis(dataDir string, impactAnalysis *analysis.UpgradeImpactAnalysis) error {
	// Create analysis directory if it doesn't exist
	analysisDir := filepath.Join(dataDir, "analysis")
	if err := os.MkdirAll(analysisDir, 0755); err != nil {
		return fmt.Errorf("failed to create analysis directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(analysisDir, fmt.Sprintf("upgrade_impact_%s_%s.json", impactAnalysis.PlayerTag, timestamp))

	// Save as JSON
	data, err := json.MarshalIndent(impactAnalysis, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write analysis file: %w", err)
	}

	return nil
}
