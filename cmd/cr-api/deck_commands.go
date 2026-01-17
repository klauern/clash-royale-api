package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/budget"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/klauer/clash-royale-api/go/pkg/leaderboard"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
	"github.com/klauer/clash-royale-api/go/pkg/recommend"
	"github.com/urfave/cli/v3"
)

// addDeckCommands adds deck-related subcommands to the CLI
func addDeckCommands() *cli.Command {
	return &cli.Command{
		Name:  "deck",
		Usage: "Deck building and analysis commands",
		Commands: []*cli.Command{
			{
				Name:  "evaluate",
				Usage: "Evaluate a deck with comprehensive analysis and scoring",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "deck",
						Usage: "Deck string (8 cards separated by dashes, e.g., Knight-Archers-Fireball-...)",
					},
					&cli.StringFlag{
						Name:    "tag",
						Aliases: []string{"p"},
						Usage:   "Player tag (without #) for card level context and upgrade impact analysis",
					},
					&cli.StringFlag{
						Name:  "from-analysis",
						Usage: "Load deck from analysis JSON file",
					},
					&cli.IntFlag{
						Name:  "arena",
						Value: 0,
						Usage: "Arena level for card unlock context (0 = no restriction)",
					},
					&cli.StringFlag{
						Name:  "format",
						Value: "human",
						Usage: "Output format: human, json, csv, detailed",
					},
					&cli.StringFlag{
						Name:  "output",
						Usage: "Output file path (optional, prints to stdout if not specified)",
					},
					&cli.BoolFlag{
						Name:  "show-upgrade-impact",
						Usage: "Show upgrade impact analysis and recommendations (requires --tag)",
					},
					&cli.IntFlag{
						Name:  "top-upgrades",
						Value: 5,
						Usage: "Number of top upgrades to show in upgrade impact analysis",
					},
				},
				Action: deckEvaluateCommand,
			},
			{
				Name:  "build",
				Usage: "Build an optimized deck based on player's card collection",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "strategy",
						Aliases: []string{"s"},
						Value:   "balanced",
						Usage:   "Deck building strategy: balanced, aggro, control, cycle, splash, spell, all",
					},
					&cli.Float64Flag{
						Name:  "min-elixir",
						Value: 2.5,
						Usage: "Minimum average elixir for the deck",
					},
					&cli.Float64Flag{
						Name:  "max-elixir",
						Value: 4.5,
						Usage: "Maximum average elixir for the deck",
					},
					&cli.StringSliceFlag{
						Name:  "include-cards",
						Usage: "Specific cards to include in the deck (by name)",
					},
					&cli.StringSliceFlag{
						Name:  "exclude-cards",
						Usage: "Cards to exclude from the deck (by name)",
					},
					&cli.IntFlag{
						Name:  "min-level",
						Value: 1,
						Usage: "Minimum card level to consider",
					},
					&cli.BoolFlag{
						Name:  "prioritize-upgrades",
						Usage: "Prioritize cards that can be upgraded soon",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export deck analysis to CSV",
					},
					&cli.BoolFlag{
						Name:  "save",
						Usage: "Save deck to file",
					},
					&cli.StringFlag{
						Name:  "unlocked-evolutions",
						Usage: "Comma-separated list of cards with unlocked evolutions (overrides UNLOCKED_EVOLUTIONS env var)",
					},
					&cli.IntFlag{
						Name:  "evolution-slots",
						Value: 2,
						Usage: "Number of evolution slots available (default 2)",
					},
					&cli.Float64Flag{
						Name:  "combat-stats-weight",
						Value: 0.25,
						Usage: "Weight for combat stats in scoring (0.0-1.0, where 0=disabled, 0.25=default, 1.0=combat-only)",
					},
					&cli.BoolFlag{
						Name:  "disable-combat-stats",
						Usage: "Disable combat stats completely (use traditional scoring only)",
					},
					&cli.BoolFlag{
						Name:    "from-analysis",
						Aliases: []string{"a"},
						Usage:   "Enable offline mode: load analysis from JSON file instead of fetching from API",
					},
					&cli.StringFlag{
						Name:  "analysis-dir",
						Usage: "Directory containing analysis JSON files (default: data/analysis)",
					},
					&cli.StringFlag{
						Name:  "analysis-file",
						Usage: "Specific analysis file path (overrides --analysis-dir lookup)",
					},
					&cli.BoolFlag{
						Name:  "enable-synergy",
						Usage: "Enable synergy-based card selection (considers card interactions and combos)",
					},
					&cli.Float64Flag{
						Name:  "synergy-weight",
						Value: 0.15,
						Usage: "Weight for synergy scoring (0.0-1.0, default 0.15 = 15%)",
					},
					&cli.BoolFlag{
						Name:  "no-suggest-upgrades",
						Usage: "Disable upgrade recommendations for the built deck (recommendations are shown by default)",
					},
					&cli.IntFlag{
						Name:  "upgrade-count",
						Value: 5,
						Usage: "Number of upgrade recommendations to show (default 5)",
					},
					&cli.BoolFlag{
						Name:  "ideal-deck",
						Usage: "Show ideal deck composition after applying recommended upgrades",
					},
				},
				Action: deckBuildCommand,
			},
			{
				Name:  "build-suite",
				Usage: "Build multiple deck variations in one invocation for systematic analysis",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "strategies",
						Aliases: []string{"s"},
						Value:   "balanced",
						Usage:   "Comma-separated list of strategies or 'all': balanced,aggro,control,cycle,splash,spell",
					},
					&cli.IntFlag{
						Name:  "variations",
						Value: 1,
						Usage: "Number of variations per strategy (default 1)",
					},
					&cli.StringFlag{
						Name:  "output-dir",
						Usage: "Output directory for deck files (default: data/decks/)",
					},
					&cli.BoolFlag{
						Name:  "from-analysis",
						Usage: "Use offline mode with pre-analyzed player data",
					},
					&cli.Float64Flag{
						Name:  "min-elixir",
						Value: 2.5,
						Usage: "Minimum average elixir for the deck",
					},
					&cli.Float64Flag{
						Name:  "max-elixir",
						Value: 4.5,
						Usage: "Maximum average elixir for the deck",
					},
					&cli.StringSliceFlag{
						Name:  "include-cards",
						Usage: "Cards that must be included in all decks",
					},
					&cli.StringSliceFlag{
						Name:  "exclude-cards",
						Usage: "Cards that must be excluded from all decks",
					},
					&cli.BoolFlag{
						Name:  "save",
						Value: true,
						Usage: "Save individual deck files and summary JSON (default: true)",
					},
				},
				Action: deckBuildSuiteCommand,
			},
			{
				Name:  "evaluate-batch",
				Usage: "Evaluate multiple decks from a suite or directory",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "from-suite",
						Usage: "Path to deck suite summary JSON file (from build-suite command)",
					},
					&cli.StringFlag{
						Name:  "deck-dir",
						Usage: "Directory containing deck JSON files",
					},
					&cli.StringFlag{
						Name:    "tag",
						Aliases: []string{"p"},
						Usage:   "Player tag (without #) for context-aware evaluation",
					},
					&cli.StringFlag{
						Name:  "format",
						Value: "summary",
						Usage: "Output format: summary, json, csv, detailed",
					},
					&cli.StringFlag{
						Name:  "output-dir",
						Usage: "Output directory for evaluation files (default: prints to stdout)",
					},
					&cli.StringFlag{
						Name:  "sort-by",
						Value: "overall",
						Usage: "Sort results by: overall, attack, defense, synergy, versatility, f2p, playability, elixir",
					},
					&cli.BoolFlag{
						Name:  "top-only",
						Usage: "Show only top N decks",
					},
					&cli.IntFlag{
						Name:  "top-n",
						Value: 10,
						Usage: "Number of top decks to show (with --top-only)",
					},
					&cli.BoolFlag{
						Name:  "filter-archetype",
						Usage: "Filter by specific archetype (use with --archetype)",
					},
					&cli.StringFlag{
						Name:  "archetype",
						Usage: "Archetype to filter by (e.g., beatdown, control, cycle)",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show detailed progress information",
					},
					&cli.BoolFlag{
						Name:  "timing",
						Usage: "Show timing information for each deck evaluation",
					},
					&cli.BoolFlag{
						Name:  "save-aggregated",
						Value: true,
						Usage: "Save aggregated results to output-dir (default: true)",
					},
				},
				Action: deckEvaluateBatchCommand,
			},
			{
				Name:  "analyze-suite",
				Usage: "Build deck variations, evaluate all decks, compare top performers, and generate comprehensive report",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "strategies",
						Aliases: []string{"s"},
						Value:   "all",
						Usage:   "Deck building strategies (comma-separated or 'all'): balanced, aggro, control, cycle, splash, spell",
					},
					&cli.IntFlag{
						Name:  "variations",
						Value: 1,
						Usage: "Number of variations per strategy",
					},
					&cli.StringFlag{
						Name:  "output-dir",
						Value: "data/analysis",
						Usage: "Base output directory for all analysis results",
					},
					&cli.IntFlag{
						Name:  "top-n",
						Value: 5,
						Usage: "Number of top decks to compare in final report",
					},
					&cli.BoolFlag{
						Name:  "from-analysis",
						Usage: "Use offline mode (load from existing analysis files instead of API)",
					},
					&cli.Float64Flag{
						Name:  "min-elixir",
						Value: 2.5,
						Usage: "Minimum average elixir for decks",
					},
					&cli.Float64Flag{
						Name:  "max-elixir",
						Value: 4.5,
						Usage: "Maximum average elixir for decks",
					},
					&cli.StringSliceFlag{
						Name:  "include-cards",
						Usage: "Cards that must be included in all decks",
					},
					&cli.StringSliceFlag{
						Name:  "exclude-cards",
						Usage: "Cards that must be excluded from all decks",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show detailed progress information",
					},
				},
				Action: deckAnalyzeSuiteCommand,
			},
			{
				Name:  "war",
				Usage: "Build a 4-deck war set with no repeated cards",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "deck-count",
						Value: 4,
						Usage: "Number of decks to build (default 4)",
					},
					&cli.StringFlag{
						Name:  "unlocked-evolutions",
						Usage: "Comma-separated list of cards with unlocked evolutions (overrides UNLOCKED_EVOLUTIONS env var)",
					},
					&cli.IntFlag{
						Name:  "evolution-slots",
						Value: 2,
						Usage: "Number of evolution slots available (default 2)",
					},
					&cli.Float64Flag{
						Name:  "combat-stats-weight",
						Value: 0.25,
						Usage: "Weight for combat stats in scoring (0.0-1.0, where 0=disabled, 0.25=default, 1.0=combat-only)",
					},
					&cli.BoolFlag{
						Name:  "disable-combat-stats",
						Usage: "Disable combat stats completely (use traditional scoring only)",
					},
				},
				Action: deckWarCommand,
			},
			{
				Name:  "analyze",
				Usage: "Analyze a deck's strengths and weaknesses",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "cards",
						Aliases:  []string{"c"},
						Usage:    "8 card names for the deck",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export analysis to CSV",
					},
				},
				Action: deckAnalyzeCommand,
			},
			{
				Name:  "optimize",
				Usage: "Optimize an existing deck with available cards",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:    "cards",
						Aliases: []string{"c"},
						Usage:   "Current 8-card deck to optimize (if not provided, uses player's current deck from API)",
					},
					&cli.IntFlag{
						Name:  "max-changes",
						Value: 4,
						Usage: "Maximum number of cards to change",
					},
					&cli.BoolFlag{
						Name:  "keep-win-con",
						Usage: "Keep win condition cards unchanged",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export optimization results to CSV",
					},
					&cli.BoolFlag{
						Name:  "verbose",
						Usage: "Enable verbose output",
					},
				},
				Action: deckOptimizeCommand,
			},
			{
				Name:  "recommend",
				Usage: "Get deck recommendations based on meta analysis",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "arena",
						Usage: "Filter by arena name",
					},
					&cli.StringFlag{
						Name:  "league",
						Usage: "Filter by league name",
					},
					&cli.IntFlag{
						Name:  "limit",
						Value: 5,
						Usage: "Number of recommendations to return",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export recommendations to CSV",
					},
					&cli.BoolFlag{
						Name:    "from-analysis",
						Aliases: []string{"a"},
						Usage:   "Enable offline mode: load analysis from JSON file instead of fetching from API",
					},
					&cli.StringFlag{
						Name:  "analysis-dir",
						Usage: "Directory containing analysis JSON files (default: data/analysis)",
					},
					&cli.StringFlag{
						Name:  "analysis-file",
						Usage: "Specific analysis file path (overrides --analysis-dir lookup)",
					},
				},
				Action: deckRecommendCommand,
			},
			{
				Name:  "mulligan",
				Usage: "Generate mulligan guide (opening hand strategy) for a deck",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "cards",
						Aliases:  []string{"c"},
						Usage:    "8 card names for the deck to analyze",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "deck-name",
						Usage: "Custom name for the deck (optional)",
					},
					&cli.BoolFlag{
						Name:  "save",
						Usage: "Save mulligan guide to file",
					},
					&cli.BoolFlag{
						Name:  "json",
						Usage: "Output guide in JSON format",
					},
				},
				Action: deckMulliganCommand,
			},
			{
				Name:  "budget",
				Usage: "Find budget-optimized decks with minimal upgrade investment",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "max-cards",
						Value: 0,
						Usage: "Maximum cards needed for upgrades (0 = no limit)",
					},
					&cli.IntFlag{
						Name:  "max-gold",
						Value: 0,
						Usage: "Maximum gold needed for upgrades (0 = no limit)",
					},
					&cli.Float64Flag{
						Name:  "target-level",
						Value: 12.0,
						Usage: "Target average card level for viability",
					},
					&cli.StringFlag{
						Name:  "sort-by",
						Value: "roi",
						Usage: "Sort results by: roi, cost_efficiency, total_cards, total_gold, current_score, projected_score",
					},
					&cli.IntFlag{
						Name:  "top-n",
						Value: 10,
						Usage: "Number of top decks to display",
					},
					&cli.BoolFlag{
						Name:  "include-variations",
						Usage: "Generate and analyze deck variations",
					},
					&cli.IntFlag{
						Name:  "max-variations",
						Value: 5,
						Usage: "Maximum number of deck variations to generate",
					},
					&cli.BoolFlag{
						Name:  "quick-wins",
						Usage: "Show only quick-win decks (1-2 upgrades away)",
					},
					&cli.BoolFlag{
						Name:  "ready-only",
						Usage: "Show only decks that are already competitive",
					},
					&cli.StringFlag{
						Name:  "unlocked-evolutions",
						Usage: "Comma-separated list of cards with unlocked evolutions",
					},
					&cli.IntFlag{
						Name:  "evolution-slots",
						Value: 2,
						Usage: "Number of evolution slots available",
					},
					&cli.BoolFlag{
						Name:  "json",
						Usage: "Output results in JSON format",
					},
					&cli.BoolFlag{
						Name:  "save",
						Usage: "Save results to file",
					},
				},
				Action: deckBudgetCommand,
			},
			{
				Name:  "possible-count",
				Usage: "Calculate the number of possible deck combinations from available cards",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #) - calculates decks possible with player's cards",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "format",
						Value: "human",
						Usage: "Output format: human, json, csv",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show detailed breakdown by role and elixir range",
					},
					&cli.StringFlag{
						Name:  "output",
						Usage: "Save output to file (optional, prints to stdout if not specified)",
					},
				},
				Action: deckPossibleCountCommand,
			},
			{
				Name:  "fuzz",
				Usage: "Generate and evaluate random deck combinations using Monte Carlo sampling",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #) for card collection context",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "count",
						Value: 1000,
						Usage: "Number of random decks to generate and evaluate",
					},
					&cli.IntFlag{
						Name:  "workers",
						Value: 1,
						Usage: "Number of parallel workers for deck generation",
					},
					&cli.StringSliceFlag{
						Name:  "include-cards",
						Usage: "Cards that must be included in every generated deck",
					},
					&cli.StringSliceFlag{
						Name:  "exclude-cards",
						Usage: "Cards that must be excluded from all generated decks",
					},
					&cli.Float64Flag{
						Name:  "min-elixir",
						Value: 0.0,
						Usage: "Minimum average elixir for generated decks",
					},
					&cli.Float64Flag{
						Name:  "max-elixir",
						Value: 10.0,
						Usage: "Maximum average elixir for generated decks",
					},
					&cli.Float64Flag{
						Name:  "min-overall",
						Value: 0.0,
						Usage: "Minimum overall score to include in results (0.0-10.0)",
					},
					&cli.Float64Flag{
						Name:  "min-synergy",
						Value: 0.0,
						Usage: "Minimum synergy score to include in results (0.0-10.0)",
					},
					&cli.IntFlag{
						Name:  "top",
						Value: 10,
						Usage: "Number of top decks to display in results",
					},
					&cli.StringFlag{
						Name:  "sort-by",
						Value: "overall",
						Usage: "Sort results by: overall, attack, defense, synergy, versatility, elixir",
					},
					&cli.StringFlag{
						Name:  "format",
						Value: "summary",
						Usage: "Output format: summary, json, csv, detailed",
					},
					&cli.StringFlag{
						Name:  "output-dir",
						Usage: "Directory to save results (default: stdout only)",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show detailed progress information",
					},
					&cli.BoolFlag{
						Name:  "from-analysis",
						Usage: "Load player data from existing analysis file (offline mode)",
					},
					&cli.StringFlag{
						Name:  "analysis-file",
						Usage: "Path to specific analysis file (for --from-analysis)",
					},
					&cli.StringFlag{
						Name:  "analysis-dir",
						Usage: "Directory containing analysis files (for --from-analysis)",
					},
					&cli.IntFlag{
						Name:  "seed",
						Value: 0,
						Usage: "Random seed for reproducibility (0 = random)",
					},
					&cli.StringFlag{
						Name:  "storage",
						Usage: "Path to persistent storage database for saving evaluated decks",
					},
					&cli.BoolFlag{
						Name:  "save-top",
						Usage: "Save top decks to persistent storage for reuse in subsequent fuzz runs",
					},
				},
				Action: deckFuzzCommand,
			},
		},
	}
}

func deckBuildCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	strategy := cmd.String("strategy")
	minElixir := cmd.Float64("min-elixir")
	maxElixir := cmd.Float64("max-elixir")
	saveData := cmd.Bool("save")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	dataDir := cmd.String("data-dir")
	combatStatsWeight := cmd.Float64("combat-stats-weight")
	disableCombatStats := cmd.Bool("disable-combat-stats")
	includeCards := cmd.StringSlice("include-cards")
	excludeCards := cmd.StringSlice("exclude-cards")

	// Upgrade recommendations flags
	noSuggestUpgrades := cmd.Bool("no-suggest-upgrades")
	upgradeCount := cmd.Int("upgrade-count")
	idealDeck := cmd.Bool("ideal-deck")

	// Offline mode flags
	fromAnalysis := cmd.Bool("from-analysis")
	analysisDir := cmd.String("analysis-dir")
	analysisFile := cmd.String("analysis-file")

	// Configure combat stats weight
	if disableCombatStats {
		os.Setenv("COMBAT_STATS_WEIGHT", "0")
		if verbose {
			fmt.Printf("Combat stats disabled (using traditional scoring only)\n")
		}
	} else if combatStatsWeight >= 0 && combatStatsWeight <= 1.0 {
		os.Setenv("COMBAT_STATS_WEIGHT", fmt.Sprintf("%.2f", combatStatsWeight))
		if verbose {
			fmt.Printf("Combat stats weight set to: %.2f\n", combatStatsWeight)
		}
	}

	// Create deck builder
	builder := deck.NewBuilder(dataDir)

	// Override unlocked evolutions if CLI flag provided
	if unlockedEvos := cmd.String("unlocked-evolutions"); unlockedEvos != "" {
		builder.SetUnlockedEvolutions(strings.Split(unlockedEvos, ","))
	}

	// Override evolution slot limit if provided
	if slots := cmd.Int("evolution-slots"); slots > 0 {
		builder.SetEvolutionSlotLimit(slots)
	}

	// Set include/exclude card filters if provided
	if len(includeCards) > 0 {
		builder.SetIncludeCards(includeCards)
	}
	if len(excludeCards) > 0 {
		builder.SetExcludeCards(excludeCards)
	}

	// Parse and set strategy if provided (skip for "all" which is handled later)
	if strategy != "" && strings.ToLower(strings.TrimSpace(strategy)) != "all" {
		parsedStrategy, err := deck.ParseStrategy(strategy)
		if err != nil {
			return fmt.Errorf("invalid strategy: %w", err)
		}
		if err := builder.SetStrategy(parsedStrategy); err != nil {
			return fmt.Errorf("failed to set strategy: %w", err)
		}
		if verbose {
			fmt.Printf("Using deck building strategy: %s\n", parsedStrategy)
		}
	}

	// Configure synergy system if enabled
	if enableSynergy := cmd.Bool("enable-synergy"); enableSynergy {
		builder.SetSynergyEnabled(true)
		if synergyWeight := cmd.Float64("synergy-weight"); synergyWeight > 0 {
			builder.SetSynergyWeight(synergyWeight)
		}
		if verbose {
			fmt.Printf("Synergy scoring enabled (weight: %.2f)\n", cmd.Float64("synergy-weight"))
		}
	}

	var deckCardAnalysis deck.CardAnalysis
	var playerName, playerTag string

	if fromAnalysis {
		// OFFLINE MODE: Load from existing analysis JSON
		if verbose {
			fmt.Printf("Building deck from offline analysis for player %s\n", tag)
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
				return fmt.Errorf("failed to load analysis file %s: %w", analysisFile, err)
			}
			if verbose {
				fmt.Printf("Loaded analysis from: %s\n", analysisFile)
			}
		} else {
			// Load latest analysis for player tag
			loadedAnalysis, err = builder.LoadLatestAnalysis(tag, analysisDir)
			if err != nil {
				return fmt.Errorf("failed to load analysis for player %s from %s: %w", tag, analysisDir, err)
			}
			if verbose {
				fmt.Printf("Loaded latest analysis from: %s\n", analysisDir)
			}
		}

		deckCardAnalysis = *loadedAnalysis
		playerTag = tag
		playerName = tag // Use tag as name in offline mode
	} else {
		// ONLINE MODE: Fetch from API
		if apiToken == "" {
			return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag. Use --from-analysis for offline mode.")
		}

		client := clashroyale.NewClient(apiToken)

		if verbose {
			fmt.Printf("Building deck for player %s\n", tag)
		}

		// Get player information
		player, err := client.GetPlayer(tag)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}

		playerName = player.Name
		playerTag = player.Tag

		if verbose {
			fmt.Printf("Player: %s (%s)\n", player.Name, player.Tag)
			fmt.Printf("Analyzing %d cards...\n", len(player.Cards))
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
		}

		for cardName, cardInfo := range cardAnalysis.CardLevels {
			deckCardAnalysis.CardLevels[cardName] = deck.CardLevelData{
				Level:             cardInfo.Level,
				MaxLevel:          cardInfo.MaxLevel,
				Rarity:            cardInfo.Rarity,
				Elixir:            cardInfo.Elixir,
				MaxEvolutionLevel: cardInfo.MaxEvolutionLevel,
			}
		}
	}

	// Apply exclude filter (works for both modes)
	excludeMap := make(map[string]bool)
	for _, card := range excludeCards {
		trimmed := strings.TrimSpace(card)
		if trimmed != "" {
			excludeMap[strings.ToLower(trimmed)] = true
		}
	}

	if len(excludeMap) > 0 {
		filteredLevels := make(map[string]deck.CardLevelData)
		for cardName, cardInfo := range deckCardAnalysis.CardLevels {
			if !excludeMap[strings.ToLower(cardName)] {
				filteredLevels[cardName] = cardInfo
			}
		}
		deckCardAnalysis.CardLevels = filteredLevels
	}

	// Handle --strategy all
	if strings.ToLower(strings.TrimSpace(strategy)) == "all" {
		return buildAllStrategies(ctx, cmd, builder, deckCardAnalysis, playerName, playerTag)
	}

	// Build deck from analysis
	deckRec, err := builder.BuildDeckFromAnalysis(deckCardAnalysis)
	if err != nil {
		return fmt.Errorf("failed to build deck: %w", err)
	}

	// Validate elixir constraints
	if deckRec.AvgElixir < minElixir || deckRec.AvgElixir > maxElixir {
		fmt.Printf("\n⚠ Warning: Deck average elixir (%.2f) is outside requested range (%.1f-%.1f)\n",
			deckRec.AvgElixir, minElixir, maxElixir)
	}

	// Display deck recommendation
	displayDeckRecommendationOffline(deckRec, playerName, playerTag)

	// Display upgrade recommendations by default (unless disabled)
	var upgrades *deck.UpgradeRecommendations
	if !noSuggestUpgrades {
		fmt.Printf("\n")
		var err error
		upgrades, err = builder.GetUpgradeRecommendations(deckCardAnalysis, deckRec, upgradeCount)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: Failed to generate upgrade recommendations: %v\n", err)
			}
		} else {
			displayUpgradeRecommendations(upgrades)
		}
	}

	// Show ideal deck with recommended upgrades applied
	if idealDeck && upgrades != nil && len(upgrades.Recommendations) > 0 {
		fmt.Printf("\n")
		fmt.Printf("============================================================================\n")
		fmt.Printf("                        IDEAL DECK (WITH UPGRADES)\n")
		fmt.Printf("============================================================================\n\n")

		// Create a copy of the card analysis with simulated upgrades
		idealAnalysis := deck.CardAnalysis{
			CardLevels:   make(map[string]deck.CardLevelData),
			AnalysisTime: deckCardAnalysis.AnalysisTime,
		}

		// Copy all card levels
		for cardName, cardData := range deckCardAnalysis.CardLevels {
			idealAnalysis.CardLevels[cardName] = cardData
		}

		// Apply upgrades
		fmt.Printf("Simulating upgrades:\n")
		for _, rec := range upgrades.Recommendations {
			if cardData, exists := idealAnalysis.CardLevels[rec.CardName]; exists {
				oldLevel := cardData.Level
				cardData.Level = rec.TargetLevel
				idealAnalysis.CardLevels[rec.CardName] = cardData
				fmt.Printf("  • %s: Level %d → %d\n", rec.CardName, oldLevel, rec.TargetLevel)
			}
		}
		fmt.Printf("\n")

		// Build ideal deck with upgraded cards
		idealDeckRec, err := builder.BuildDeckFromAnalysis(idealAnalysis)
		if err != nil {
			fmt.Printf("Warning: Failed to build ideal deck: %v\n", err)
		} else {
			displayDeckRecommendationOffline(idealDeckRec, playerName, playerTag)

			// Show comparison
			fmt.Printf("\n")
			fmt.Printf("Comparison:\n")
			fmt.Printf("  Current Deck:  %.2f avg elixir\n", deckRec.AvgElixir)
			fmt.Printf("  Ideal Deck:    %.2f avg elixir\n", idealDeckRec.AvgElixir)

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
				fmt.Printf("\n  Deck Changes:\n")
				if len(removedCards) > 0 {
					fmt.Printf("    Removed: %s\n", strings.Join(removedCards, ", "))
				}
				if len(addedCards) > 0 {
					fmt.Printf("    Added:   %s\n", strings.Join(addedCards, ", "))
				}
			} else {
				fmt.Printf("\n  Deck composition remains the same (upgrades strengthen existing cards)\n")
			}
		}
	}

	// Save deck if requested
	if saveData {
		if verbose {
			fmt.Printf("\nSaving deck to: %s\n", dataDir)
		}
		deckPath, err := builder.SaveDeck(deckRec, "", playerTag)
		if err != nil {
			fmt.Printf("Warning: Failed to save deck: %v\n", err)
		} else {
			fmt.Printf("\nDeck saved to: %s\n", deckPath)
		}
	}

	return nil
}

func deckBuildSuiteCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	strategiesStr := cmd.String("strategies")
	variations := cmd.Int("variations")
	outputDir := cmd.String("output-dir")
	fromAnalysis := cmd.Bool("from-analysis")
	saveData := cmd.Bool("save")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	dataDir := cmd.String("data-dir")
	minElixir := cmd.Float64("min-elixir")
	maxElixir := cmd.Float64("max-elixir")
	includeCards := cmd.StringSlice("include-cards")
	excludeCards := cmd.StringSlice("exclude-cards")

	// Determine output directory
	if outputDir == "" {
		outputDir = filepath.Join(dataDir, "decks")
	}

	// Parse strategies
	var strategies []deck.Strategy
	strategiesStr = strings.ToLower(strings.TrimSpace(strategiesStr))
	if strategiesStr == "all" {
		strategies = []deck.Strategy{
			deck.StrategyBalanced,
			deck.StrategyAggro,
			deck.StrategyControl,
			deck.StrategyCycle,
			deck.StrategySplash,
			deck.StrategySpell,
		}
	} else {
		strategyList := strings.Split(strategiesStr, ",")
		for _, s := range strategyList {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			parsedStrategy, err := deck.ParseStrategy(s)
			if err != nil {
				return fmt.Errorf("invalid strategy '%s': %w", s, err)
			}
			strategies = append(strategies, parsedStrategy)
		}
	}

	if len(strategies) == 0 {
		return fmt.Errorf("no valid strategies specified")
	}

	if verbose {
		fmt.Printf("Building deck suite with %d strategies x %d variations = %d total decks\n",
			len(strategies), variations, len(strategies)*variations)
	}

	// Create deck builder
	builder := deck.NewBuilder(dataDir)

	// Set include/exclude card filters if provided
	if len(includeCards) > 0 {
		builder.SetIncludeCards(includeCards)
	}
	if len(excludeCards) > 0 {
		builder.SetExcludeCards(excludeCards)
	}

	// Load player data
	var deckCardAnalysis deck.CardAnalysis
	var playerName, playerTag string

	if fromAnalysis {
		// OFFLINE MODE: Load from existing analysis JSON
		if verbose {
			fmt.Printf("Building deck suite from offline analysis for player %s\n", tag)
		}

		analysisDir := filepath.Join(dataDir, "analysis")
		loadedAnalysis, err := builder.LoadLatestAnalysis(tag, analysisDir)
		if err != nil {
			return fmt.Errorf("failed to load analysis for player %s from %s: %w", tag, analysisDir, err)
		}

		deckCardAnalysis = *loadedAnalysis
		playerTag = tag
		playerName = tag
	} else {
		// ONLINE MODE: Fetch from API
		if apiToken == "" {
			return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag. Use --from-analysis for offline mode.")
		}

		client := clashroyale.NewClient(apiToken)

		if verbose {
			fmt.Printf("Building deck suite for player %s\n", tag)
		}

		// Get player information
		player, err := client.GetPlayer(tag)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}

		playerName = player.Name
		playerTag = player.Tag

		if verbose {
			fmt.Printf("Player: %s (%s)\n", player.Name, player.Tag)
			fmt.Printf("Analyzing %d cards...\n", len(player.Cards))
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
		}

		for cardName, cardInfo := range cardAnalysis.CardLevels {
			deckCardAnalysis.CardLevels[cardName] = deck.CardLevelData{
				Level:             cardInfo.Level,
				MaxLevel:          cardInfo.MaxLevel,
				Rarity:            cardInfo.Rarity,
				Elixir:            cardInfo.Elixir,
				MaxEvolutionLevel: cardInfo.MaxEvolutionLevel,
			}
		}
	}

	// Apply exclude filter
	if len(excludeCards) > 0 {
		excludeMap := make(map[string]bool)
		for _, card := range excludeCards {
			trimmed := strings.TrimSpace(card)
			if trimmed != "" {
				excludeMap[strings.ToLower(trimmed)] = true
			}
		}

		filteredLevels := make(map[string]deck.CardLevelData)
		for cardName, cardInfo := range deckCardAnalysis.CardLevels {
			if !excludeMap[strings.ToLower(cardName)] {
				filteredLevels[cardName] = cardInfo
			}
		}
		deckCardAnalysis.CardLevels = filteredLevels
	}

	// Build decks for all strategy x variation combinations
	type deckResult struct {
		Strategy   string
		Variation  int
		Deck       *deck.DeckRecommendation
		FilePath   string
		BuildError error
	}

	startTime := time.Now()
	results := []deckResult{}

	fmt.Printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                    DECK BUILD SUITE                                ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")
	fmt.Printf("Player: %s (%s)\n", playerName, playerTag)
	fmt.Printf("Output: %s\n\n", outputDir)

	// Build decks for each strategy
	for _, strategy := range strategies {
		if verbose {
			fmt.Printf("Building decks for strategy: %s\n", strategy)
		}

		for v := 1; v <= variations; v++ {
			// Create a new builder for this deck
			deckBuilder := deck.NewBuilder(dataDir)

			// Copy configuration
			if len(includeCards) > 0 {
				deckBuilder.SetIncludeCards(includeCards)
			}
			if len(excludeCards) > 0 {
				deckBuilder.SetExcludeCards(excludeCards)
			}

			// Set strategy
			if err := deckBuilder.SetStrategy(strategy); err != nil {
				results = append(results, deckResult{
					Strategy:   string(strategy),
					Variation:  v,
					BuildError: err,
				})
				fmt.Printf("  ⚠ Variation %d: Failed to set strategy: %v\n", v, err)
				continue
			}

			// Build deck
			deckRec, err := deckBuilder.BuildDeckFromAnalysis(deckCardAnalysis)
			if err != nil {
				results = append(results, deckResult{
					Strategy:   string(strategy),
					Variation:  v,
					BuildError: err,
				})
				fmt.Printf("  ⚠ Variation %d: Failed to build deck: %v\n", v, err)
				continue
			}

			// Validate elixir constraints
			if deckRec.AvgElixir < minElixir || deckRec.AvgElixir > maxElixir {
				if verbose {
					fmt.Printf("  ⚠ Variation %d: Deck average elixir (%.2f) outside range (%.1f-%.1f)\n",
						v, deckRec.AvgElixir, minElixir, maxElixir)
				}
			}

			// Save deck file if requested
			var filePath string
			if saveData {
				timestamp := time.Now().Format("20060102_150405")
				filename := fmt.Sprintf("%s_deck_%s_var%d_%s.json", timestamp, strategy, v, playerTag)
				filePath = filepath.Join(outputDir, filename)

				// Save using builder
				savedPath, err := deckBuilder.SaveDeck(deckRec, outputDir, fmt.Sprintf("%s_var%d_%s", strategy, v, playerTag))
				if err != nil {
					if verbose {
						fmt.Printf("  ⚠ Variation %d: Failed to save deck: %v\n", v, err)
					}
				} else {
					filePath = savedPath
				}
			}

			results = append(results, deckResult{
				Strategy:  string(strategy),
				Variation: v,
				Deck:      deckRec,
				FilePath:  filePath,
			})

			fmt.Printf("  ✓ %s variation %d: %.2f avg elixir, %d cards\n",
				strategy, v, deckRec.AvgElixir, len(deckRec.Deck))
		}
	}

	totalTime := time.Since(startTime)

	// Display summary
	fmt.Printf("\n")
	fmt.Printf("════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("                           SUMMARY\n")
	fmt.Printf("════════════════════════════════════════════════════════════════════\n\n")

	successful := 0
	failed := 0
	for _, r := range results {
		if r.BuildError == nil {
			successful++
		} else {
			failed++
		}
	}

	fmt.Printf("Total decks:     %d\n", len(results))
	fmt.Printf("Successful:      %d\n", successful)
	fmt.Printf("Failed:          %d\n", failed)
	fmt.Printf("Build time:      %v\n", totalTime)
	fmt.Printf("Avg per deck:    %v\n\n", totalTime/time.Duration(len(results)))

	// Save summary JSON if requested
	if saveData && successful > 0 {
		timestamp := time.Now().Format("20060102_150405")
		summaryFilename := fmt.Sprintf("%s_deck_suite_summary_%s.json", timestamp, playerTag)
		summaryPath := filepath.Join(outputDir, summaryFilename)

		// Build summary structure
		summary := map[string]interface{}{
			"version":   "1.0.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"player": map[string]string{
				"name": playerName,
				"tag":  playerTag,
			},
			"build_info": map[string]interface{}{
				"total_decks":     len(results),
				"successful":      successful,
				"failed":          failed,
				"strategies":      len(strategies),
				"variations":      variations,
				"generation_time": totalTime.String(),
			},
			"decks": []map[string]interface{}{},
		}

		// Add individual deck summaries
		decks := []map[string]interface{}{}
		for _, r := range results {
			if r.Deck != nil {
				deckSummary := map[string]interface{}{
					"strategy":   r.Strategy,
					"variation":  r.Variation,
					"cards":      r.Deck.Deck,
					"avg_elixir": r.Deck.AvgElixir,
					"file_path":  r.FilePath,
				}
				decks = append(decks, deckSummary)
			}
		}
		summary["decks"] = decks

		// Write summary JSON
		summaryJSON, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			fmt.Printf("Warning: Failed to marshal summary JSON: %v\n", err)
		} else {
			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				fmt.Printf("Warning: Failed to create output directory: %v\n", err)
			} else {
				if err := os.WriteFile(summaryPath, summaryJSON, 0o644); err != nil {
					fmt.Printf("Warning: Failed to write summary file: %v\n", err)
				} else {
					fmt.Printf("Summary saved to: %s\n", summaryPath)
				}
			}
		}
	}

	return nil
}

// deckEvaluateBatchCommand evaluates multiple decks from a suite or directory
func deckEvaluateBatchCommand(ctx context.Context, cmd *cli.Command) error {
	fromSuite := cmd.String("from-suite")
	deckDir := cmd.String("deck-dir")
	playerTag := cmd.String("tag")
	format := cmd.String("format")
	outputDir := cmd.String("output-dir")
	sortBy := cmd.String("sort-by")
	topOnly := cmd.Bool("top-only")
	topN := cmd.Int("top-n")
	filterArchetype := cmd.Bool("filter-archetype")
	archetypeFilter := cmd.String("archetype")
	verbose := cmd.Bool("verbose")
	showTiming := cmd.Bool("timing")
	saveAggregated := cmd.Bool("save-aggregated")

	// Validation: Must provide either --from-suite or --deck-dir
	if fromSuite == "" && deckDir == "" {
		return fmt.Errorf("must provide either --from-suite or --deck-dir")
	}

	if fromSuite != "" && deckDir != "" {
		return fmt.Errorf("cannot use both --from-suite and --deck-dir")
	}

	// Load deck data
	type deckInfo struct {
		Name     string
		Cards    []string
		Strategy string
		FilePath string
	}

	var decks []deckInfo
	var playerName string

	if fromSuite != "" {
		// Load from suite summary JSON
		data, err := os.ReadFile(fromSuite)
		if err != nil {
			return fmt.Errorf("failed to read suite file: %w", err)
		}

		var suiteData map[string]interface{}
		if err := json.Unmarshal(data, &suiteData); err != nil {
			return fmt.Errorf("failed to parse suite JSON: %w", err)
		}

		// Extract player info
		if playerInfo, ok := suiteData["player"].(map[string]interface{}); ok {
			if name, ok := playerInfo["name"].(string); ok {
				playerName = name
			}
			if tag, ok := playerInfo["tag"].(string); ok && playerTag == "" {
				playerTag = tag
			}
		}

		// Extract decks
		if decksList, ok := suiteData["decks"].([]interface{}); ok {
			for i, d := range decksList {
				deckMap, ok := d.(map[string]interface{})
				if !ok {
					continue
				}

				var cards []string
				if cardsList, ok := deckMap["cards"].([]interface{}); ok {
					for _, c := range cardsList {
						if cardStr, ok := c.(string); ok {
							cards = append(cards, cardStr)
						}
					}
				}

				strategy := "unknown"
				if s, ok := deckMap["strategy"].(string); ok {
					strategy = s
				}

				variation := 0
				if v, ok := deckMap["variation"].(float64); ok {
					variation = int(v)
				}

				filePath := ""
				if fp, ok := deckMap["file_path"].(string); ok {
					filePath = fp
				}

				name := fmt.Sprintf("Deck #%d (%s v%d)", i+1, strategy, variation)
				decks = append(decks, deckInfo{
					Name:     name,
					Cards:    cards,
					Strategy: strategy,
					FilePath: filePath,
				})
			}
		}

		if verbose {
			fmt.Printf("Loaded %d decks from suite: %s\n", len(decks), fromSuite)
		}
	} else {
		// Load from directory
		entries, err := os.ReadDir(deckDir)
		if err != nil {
			return fmt.Errorf("failed to read deck directory: %w", err)
		}

		for i, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}

			deckPath := filepath.Join(deckDir, entry.Name())
			data, err := os.ReadFile(deckPath)
			if err != nil {
				if verbose {
					fmt.Printf("Warning: Failed to read %s: %v\n", entry.Name(), err)
				}
				continue
			}

			var deckData map[string]interface{}
			if err := json.Unmarshal(data, &deckData); err != nil {
				if verbose {
					fmt.Printf("Warning: Failed to parse %s: %v\n", entry.Name(), err)
				}
				continue
			}

			// Try to extract cards from various possible formats
			var cards []string
			if deckMap, ok := deckData["deck"].([]interface{}); ok {
				for _, c := range deckMap {
					if cardStr, ok := c.(string); ok {
						cards = append(cards, cardStr)
					}
				}
			} else if cardsList, ok := deckData["cards"].([]interface{}); ok {
				for _, c := range cardsList {
					if cardStr, ok := c.(string); ok {
						cards = append(cards, cardStr)
					}
				}
			}

			if len(cards) != 8 {
				if verbose {
					fmt.Printf("Warning: Skipping %s (expected 8 cards, got %d)\n", entry.Name(), len(cards))
				}
				continue
			}

			name := strings.TrimSuffix(entry.Name(), ".json")
			decks = append(decks, deckInfo{
				Name:     name,
				Cards:    cards,
				FilePath: deckPath,
			})

			if i == 0 {
				// Try to extract player info from first deck
				if rec, ok := deckData["recommendation"].(map[string]interface{}); ok {
					if pname, ok := rec["player_name"].(string); ok {
						playerName = pname
					}
				}
			}
		}

		if verbose {
			fmt.Printf("Loaded %d decks from directory: %s\n", len(decks), deckDir)
		}
	}

	if len(decks) == 0 {
		return fmt.Errorf("no decks found to evaluate")
	}

	// Load player context if tag provided
	var playerContext *evaluation.PlayerContext
	if playerTag != "" {
		apiToken := cmd.String("api-token")
		if apiToken == "" {
			apiToken = os.Getenv("CLASH_ROYALE_API_TOKEN")
		}

		if apiToken != "" {
			client := clashroyale.NewClient(apiToken)
			player, err := client.GetPlayer(playerTag)
			if err != nil {
				if verbose {
					fmt.Printf("Warning: Failed to load player data: %v\n", err)
					fmt.Println("Continuing with generic evaluation (no player context)")
				}
			} else {
				playerContext = evaluation.NewPlayerContextFromPlayer(player)
				if playerName == "" && player != nil {
					playerName = player.Name
				}
				if verbose {
					fmt.Printf("Loaded player context for %s (%s)\n", player.Name, playerTag)
				}
			}
		}
	}

	if verbose {
		fmt.Printf("Evaluating %d decks...\n", len(decks))
	}

	// Create synergy database (shared for all evaluations)
	synergyDB := deck.NewSynergyDatabase()

	// Initialize persistent storage if player tag is available
	var storage *leaderboard.Storage
	var storageErr error
	if playerTag != "" {
		storage, storageErr = leaderboard.NewStorage(playerTag)
		if storageErr != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to initialize storage: %v\n", storageErr)
			}
		} else {
			defer func() {
				if err := storage.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to close storage: %v\n", err)
				}
			}()
			if verbose {
				fmt.Printf("Initialized persistent storage at: %s\n", storage.GetDBPath())
			}
		}
	}

	// Evaluate all decks
	type batchEvalResult struct {
		Name      string
		Strategy  string
		Deck      []string
		Result    evaluation.EvaluationResult
		FilePath  string
		Evaluated time.Time
		Duration  time.Duration
	}

	results := make([]batchEvalResult, 0, len(decks))
	startTime := time.Now()

	for i, deckData := range decks {
		deckStart := time.Now()

		if len(deckData.Cards) != 8 {
			if verbose {
				fmt.Printf("  [%d/%d] Skipping %s: invalid card count (%d)\n",
					i+1, len(decks), deckData.Name, len(deckData.Cards))
			}
			continue
		}

		deckCards := convertToCardCandidates(deckData.Cards)
		result := evaluation.Evaluate(deckCards, synergyDB, playerContext)

		elapsed := time.Since(deckStart)

		results = append(results, batchEvalResult{
			Name:      deckData.Name,
			Strategy:  deckData.Strategy,
			Deck:      result.Deck,
			Result:    result,
			FilePath:  deckData.FilePath,
			Evaluated: deckStart,
			Duration:  elapsed,
		})

		// Save to persistent storage if available
		if storage != nil {
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
				Strategy:          deckData.Strategy,
				AvgElixir:         result.AvgElixir,
				EvaluatedAt:       deckStart,
				PlayerTag:         playerTag,
				EvaluationVersion: "1.0.0",
			}

			_, isNew, err := storage.InsertDeck(entry)
			if err != nil && verbose {
				fmt.Fprintf(os.Stderr, "  Warning: failed to save deck to storage: %v\n", err)
			} else if verbose && !isNew {
				fmt.Printf("  (deck already in storage, updated)\n")
			}
		}

		if verbose {
			fmt.Printf("  [%d/%d] %s: %.2f (%s) - %s\n",
				i+1, len(decks), deckData.Name, result.OverallScore,
				result.OverallRating, result.DetectedArchetype)
		}
	}

	totalTime := time.Since(startTime)

	if verbose || showTiming {
		fmt.Printf("\nBatch evaluation completed in %v\n", totalTime)
		if len(results) > 0 {
			fmt.Printf("Average time per deck: %v\n", totalTime/time.Duration(len(results)))
		}
	}

	if len(results) == 0 {
		return fmt.Errorf("no decks were successfully evaluated")
	}

	// Recalculate and update statistics after all insertions
	if storage != nil {
		stats, err := storage.RecalculateStats()
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to recalculate storage stats: %v\n", err)
			}
		} else if verbose {
			fmt.Printf("\nStorage statistics updated:\n")
			fmt.Printf("  Total decks evaluated: %d\n", stats.TotalDecksEvaluated)
			fmt.Printf("  Unique decks: %d\n", stats.TotalUniqueDecks)
			fmt.Printf("  Top score: %.2f\n", stats.TopScore)
			fmt.Printf("  Average score: %.2f\n", stats.AvgScore)
		}
	}

	// Sort results
	sortEvaluationResults(results, sortBy)

	// Filter by archetype if requested
	if filterArchetype && archetypeFilter != "" {
		filtered := make([]batchEvalResult, 0)
		for _, r := range results {
			if strings.EqualFold(string(r.Result.DetectedArchetype), archetypeFilter) {
				filtered = append(filtered, r)
			}
		}
		results = filtered
		if len(results) == 0 {
			fmt.Printf("No decks found matching archetype: %s\n", archetypeFilter)
			return nil
		}
	}

	// Apply top filter
	if topOnly && len(results) > topN {
		results = results[:topN]
	}

	// Format and output results
	var output string
	switch strings.ToLower(format) {
	case "summary", "human":
		output = formatEvaluationBatchSummary(results, len(decks), totalTime, sortBy, playerName, playerTag)
	case "json":
		jsonData := map[string]interface{}{
			"version":   "1.0.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"player": map[string]string{
				"name": playerName,
				"tag":  playerTag,
			},
			"evaluation_info": map[string]interface{}{
				"total_decks":     len(decks),
				"evaluated":       len(results),
				"sort_by":         sortBy,
				"evaluation_time": totalTime.String(),
			},
			"results": results,
		}
		jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		output = string(jsonBytes)
	case "csv":
		output = formatEvaluationBatchCSV(results)
	case "detailed":
		output = formatEvaluationBatchDetailed(results, playerName, playerTag)
	default:
		return fmt.Errorf("unknown format: %s (supported: summary, json, csv, detailed)", format)
	}

	// Write output
	if outputDir != "" && saveAggregated {
		timestamp := time.Now().Format("20060102_150405")
		var filename string
		switch strings.ToLower(format) {
		case "json":
			filename = fmt.Sprintf("%s_deck_evaluations_%s.json", timestamp, playerTag)
		case "csv":
			filename = fmt.Sprintf("%s_deck_evaluations_%s.csv", timestamp, playerTag)
		default:
			filename = fmt.Sprintf("%s_deck_evaluations_%s.txt", timestamp, playerTag)
		}
		outputPath := filepath.Join(outputDir, filename)

		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if err := os.WriteFile(outputPath, []byte(output), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		fmt.Printf("\nEvaluation results saved to: %s\n", outputPath)
	} else {
		fmt.Print(output)
	}

	return nil
}

// deckAnalyzeSuiteCommand orchestrates the full deck analysis workflow:
// (1) Build multiple deck variations using build-suite logic
// (2) Evaluate all built decks using evaluate-batch logic
// (3) Compare top performers using compare logic
// (4) Generate comprehensive markdown report
func deckAnalyzeSuiteCommand(ctx context.Context, cmd *cli.Command) error {
	// Extract flags
	tag := cmd.String("tag")
	strategiesStr := cmd.String("strategies")
	variations := cmd.Int("variations")
	outputDir := cmd.String("output-dir")
	topN := cmd.Int("top-n")
	fromAnalysis := cmd.Bool("from-analysis")
	minElixir := cmd.Float64("min-elixir")
	maxElixir := cmd.Float64("max-elixir")
	includeCards := cmd.StringSlice("include-cards")
	excludeCards := cmd.StringSlice("exclude-cards")
	verbose := cmd.Bool("verbose")

	// Get global flags
	apiToken := cmd.String("api-token")
	dataDir := cmd.String("data-dir")

	if dataDir == "" {
		dataDir = "data"
	}

	// Create timestamp for consistent file naming across all phases
	timestamp := time.Now().Format("20060102_150405")

	fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║           DECK ANALYSIS SUITE - COMPREHENSIVE WORKFLOW            ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ========================================================================
	// PHASE 1: Build deck variations
	// ========================================================================
	fmt.Println("📦 PHASE 1: Building deck variations...")
	fmt.Println("─────────────────────────────────────────────────────────────────────")

	decksDir := filepath.Join(outputDir, "decks")
	if err := os.MkdirAll(decksDir, 0o755); err != nil {
		return fmt.Errorf("failed to create decks directory: %w", err)
	}

	// Parse strategies
	var strategies []deck.Strategy
	strategiesStr = strings.ToLower(strings.TrimSpace(strategiesStr))
	if strategiesStr == "all" {
		strategies = []deck.Strategy{
			deck.StrategyBalanced,
			deck.StrategyAggro,
			deck.StrategyControl,
			deck.StrategyCycle,
			deck.StrategySplash,
			deck.StrategySpell,
		}
	} else {
		strategyList := strings.Split(strategiesStr, ",")
		for _, s := range strategyList {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			parsedStrategy, err := deck.ParseStrategy(s)
			if err != nil {
				return fmt.Errorf("invalid strategy '%s': %w", s, err)
			}
			strategies = append(strategies, parsedStrategy)
		}
	}

	if len(strategies) == 0 {
		return fmt.Errorf("no valid strategies specified")
	}

	if verbose {
		fmt.Printf("Strategies: %v\n", strategies)
		fmt.Printf("Variations per strategy: %d\n", variations)
		fmt.Printf("Total decks to build: %d\n", len(strategies)*variations)
	}

	// Build decks using build-suite logic
	builder := deck.NewBuilder(dataDir)

	// Load card analysis (online or offline)
	var deckCardAnalysis deck.CardAnalysis
	var playerName, playerTag string

	if fromAnalysis {
		analysisDir := filepath.Join(dataDir, "analysis")
		loadedAnalysis, err := builder.LoadLatestAnalysis(tag, analysisDir)
		if err != nil {
			return fmt.Errorf("failed to load analysis: %w", err)
		}
		deckCardAnalysis = *loadedAnalysis
		playerName = tag
		playerTag = tag
		if verbose {
			fmt.Printf("✓ Loaded analysis from: %s\n", analysisDir)
		}
	} else {
		if apiToken == "" {
			return fmt.Errorf("API token required (use --api-token or CLASH_ROYALE_API_TOKEN env var)")
		}
		client := clashroyale.NewClient(apiToken)
		player, err := client.GetPlayer(tag)
		if err != nil {
			return fmt.Errorf("failed to fetch player data: %w", err)
		}

		playerName = player.Name
		playerTag = player.Tag

		// Analyze card collection
		options := analysis.AnalysisOptions{}
		cardAnalysis, err := analysis.AnalyzeCardCollection(player, options)
		if err != nil {
			return fmt.Errorf("failed to analyze cards: %w", err)
		}

		// Convert to deck.CardAnalysis format
		deckCardAnalysis = deck.CardAnalysis{
			CardLevels:   make(map[string]deck.CardLevelData),
			AnalysisTime: cardAnalysis.AnalysisTime.Format(time.RFC3339),
		}

		for cardName, cardInfo := range cardAnalysis.CardLevels {
			deckCardAnalysis.CardLevels[cardName] = deck.CardLevelData{
				Level:             cardInfo.Level,
				MaxLevel:          cardInfo.MaxLevel,
				Rarity:            cardInfo.Rarity,
				Elixir:            cardInfo.Elixir,
				MaxEvolutionLevel: cardInfo.MaxEvolutionLevel,
			}
		}

		if verbose {
			fmt.Printf("✓ Fetched player data for: %s (%s)\n", playerName, playerTag)
		}
	}

	// Build all deck variations
	type deckInfo struct {
		Strategy  string   `json:"strategy"`
		Variation int      `json:"variation"`
		Cards     []string `json:"cards"`
		AvgElixir float64  `json:"avg_elixir"`
		FilePath  string   `json:"file_path"`
	}

	var builtDecks []deckInfo
	_ = minElixir // Reserved for future use in validation
	_ = maxElixir // Reserved for future use in validation
	successCount := 0
	failCount := 0
	buildStart := time.Now()

	for _, strategy := range strategies {
		for v := 1; v <= variations; v++ {
			deckBuilder := deck.NewBuilder(dataDir)

			// Apply configuration
			if len(includeCards) > 0 {
				deckBuilder.SetIncludeCards(includeCards)
			}
			if len(excludeCards) > 0 {
				deckBuilder.SetExcludeCards(excludeCards)
			}

			// Set strategy
			if err := deckBuilder.SetStrategy(strategy); err != nil {
				fmt.Printf("  ✗ Failed to set strategy %s variation %d: %v\n", strategy, v, err)
				failCount++
				continue
			}

			// Build deck
			deckRec, err := deckBuilder.BuildDeckFromAnalysis(deckCardAnalysis)
			if err != nil {
				fmt.Printf("  ✗ Failed to build %s variation %d: %v\n", strategy, v, err)
				failCount++
				continue
			}

			// Save deck to file
			deckFileName := fmt.Sprintf("%s_deck_%s_var%d_%s.json", timestamp, strategy, v, strings.TrimPrefix(playerTag, "#"))
			deckFilePath := filepath.Join(decksDir, deckFileName)

			deckData := map[string]interface{}{
				"deck":           deckRec.Deck,
				"avg_elixir":     deckRec.AvgElixir,
				"recommendation": deckRec,
			}

			deckJSON, err := json.MarshalIndent(deckData, "", "  ")
			if err != nil {
				fmt.Printf("  ✗ Failed to marshal deck %s variation %d: %v\n", strategy, v, err)
				failCount++
				continue
			}

			if err := os.WriteFile(deckFilePath, deckJSON, 0o644); err != nil {
				fmt.Printf("  ✗ Failed to save deck %s variation %d: %v\n", strategy, v, err)
				failCount++
				continue
			}

			builtDecks = append(builtDecks, deckInfo{
				Strategy:  string(strategy),
				Variation: v,
				Cards:     deckRec.Deck,
				AvgElixir: deckRec.AvgElixir,
				FilePath:  deckFilePath,
			})

			successCount++
			if verbose {
				fmt.Printf("  ✓ Built %s variation %d (%.2f avg elixir)\n", strategy, v, deckRec.AvgElixir)
			}
		}
	}

	buildDuration := time.Since(buildStart)

	// Save suite summary
	suiteFileName := fmt.Sprintf("%s_deck_suite_summary_%s.json", timestamp, strings.TrimPrefix(playerTag, "#"))
	suiteSummaryPath := filepath.Join(decksDir, suiteFileName)

	suiteData := map[string]interface{}{
		"version":   "1.0.0",
		"timestamp": timestamp,
		"player": map[string]string{
			"name": playerName,
			"tag":  playerTag,
		},
		"build_info": map[string]interface{}{
			"total_decks":     len(strategies) * variations,
			"successful":      successCount,
			"failed":          failCount,
			"strategies":      len(strategies),
			"variations":      variations,
			"generation_time": buildDuration.String(),
		},
		"decks": builtDecks,
	}

	suiteJSON, err := json.MarshalIndent(suiteData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal suite summary: %w", err)
	}

	if err := os.WriteFile(suiteSummaryPath, suiteJSON, 0o644); err != nil {
		return fmt.Errorf("failed to save suite summary: %w", err)
	}

	fmt.Println()
	fmt.Printf("✓ Built %d/%d decks successfully in %s\n", successCount, successCount+failCount, buildDuration.Round(time.Millisecond))
	fmt.Printf("  Suite summary: %s\n", suiteSummaryPath)
	fmt.Println()

	if successCount == 0 {
		return fmt.Errorf("no decks were built successfully")
	}

	// ========================================================================
	// PHASE 2: Evaluate all decks
	// ========================================================================
	fmt.Println("📊 PHASE 2: Evaluating all decks...")
	fmt.Println("─────────────────────────────────────────────────────────────────────")

	evaluationsDir := filepath.Join(outputDir, "evaluations")
	if err := os.MkdirAll(evaluationsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create evaluations directory: %w", err)
	}

	// Load player context if available
	var playerContext *evaluation.PlayerContext
	if !fromAnalysis && apiToken != "" {
		client := clashroyale.NewClient(apiToken)
		player, err := client.GetPlayer(tag)
		if err == nil {
			playerContext = evaluation.NewPlayerContextFromPlayer(player)
		}
	}

	// Create shared synergy database
	synergyDB := deck.NewSynergyDatabase()

	// Initialize persistent storage if player tag is available
	var storage *leaderboard.Storage
	var storageErr error
	if tag != "" {
		storage, storageErr = leaderboard.NewStorage(tag)
		if storageErr != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to initialize storage: %v\n", storageErr)
			}
		} else {
			defer func() {
				if err := storage.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to close storage: %v\n", err)
				}
			}()
			if verbose {
				fmt.Printf("Initialized persistent storage at: %s\n", storage.GetDBPath())
			}
		}
	}

	// Evaluate each deck
	type evalResult struct {
		Name      string                      `json:"name"`
		Strategy  string                      `json:"strategy"`
		Deck      []string                    `json:"deck"`
		Result    evaluation.EvaluationResult `json:"Result"`
		FilePath  string                      `json:"FilePath"`
		Evaluated string                      `json:"Evaluated"`
		Duration  int64                       `json:"Duration"`
	}

	var results []evalResult
	evalStart := time.Now()

	for _, deckInf := range builtDecks {
		deckStart := time.Now()

		// Convert card names to CardCandidate
		var deckCards []deck.CardCandidate
		for _, cardName := range deckInf.Cards {
			deckCards = append(deckCards, deck.CardCandidate{
				Name: cardName,
				// Use defaults for evaluation
				Rarity: inferRarity(cardName),
				Elixir: config.GetCardElixir(cardName, 0),
				Role:   inferRole(cardName),
				Stats:  inferStats(cardName),
			})
		}

		// Evaluate
		deckEvalResult := evaluation.Evaluate(deckCards, synergyDB, playerContext)
		evalDuration := time.Since(deckStart)

		deckName := fmt.Sprintf("Deck #%d (%s v%d)", len(results)+1, deckInf.Strategy, deckInf.Variation)

		results = append(results, evalResult{
			Name:      deckName,
			Strategy:  deckInf.Strategy,
			Deck:      deckInf.Cards,
			Result:    deckEvalResult,
			FilePath:  deckInf.FilePath,
			Evaluated: time.Now().Format(time.RFC3339),
			Duration:  evalDuration.Nanoseconds(),
		})

		// Save to persistent storage if available
		if storage != nil {
			entry := &leaderboard.DeckEntry{
				Cards:             deckInf.Cards,
				OverallScore:      deckEvalResult.OverallScore,
				AttackScore:       deckEvalResult.Attack.Score,
				DefenseScore:      deckEvalResult.Defense.Score,
				SynergyScore:      deckEvalResult.Synergy.Score,
				VersatilityScore:  deckEvalResult.Versatility.Score,
				F2PScore:          deckEvalResult.F2PFriendly.Score,
				PlayabilityScore:  deckEvalResult.Playability.Score,
				Archetype:         string(deckEvalResult.DetectedArchetype),
				ArchetypeConf:     deckEvalResult.ArchetypeConfidence,
				Strategy:          deckInf.Strategy,
				AvgElixir:         deckEvalResult.AvgElixir,
				EvaluatedAt:       deckStart,
				PlayerTag:         tag,
				EvaluationVersion: "1.0.0",
			}

			_, _, err := storage.InsertDeck(entry)
			if err != nil && verbose {
				fmt.Fprintf(os.Stderr, "  Warning: failed to save deck to storage: %v\n", err)
			}
		}

		if verbose {
			fmt.Printf("  ✓ Evaluated %s: %.2f overall (%.0fms)\n", deckName, deckEvalResult.OverallScore, float64(evalDuration.Nanoseconds())/1e6)
		}
	}

	evalDuration := time.Since(evalStart)

	// Recalculate and update statistics after all insertions
	if storage != nil {
		_, err := storage.RecalculateStats()
		if err != nil && verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to recalculate storage stats: %v\n", err)
		}
	}

	// Sort by overall score
	sortEvaluationResults(results, "overall")

	// Save evaluation results
	evalFileName := fmt.Sprintf("%s_deck_evaluations_%s.json", timestamp, strings.TrimPrefix(playerTag, "#"))
	evalFilePath := filepath.Join(evaluationsDir, evalFileName)

	evalData := map[string]interface{}{
		"version":   "1.0.0",
		"timestamp": timestamp,
		"player": map[string]string{
			"name": playerName,
			"tag":  playerTag,
		},
		"evaluation_info": map[string]interface{}{
			"total_decks":     len(results),
			"evaluated":       len(results),
			"sort_by":         "overall",
			"evaluation_time": evalDuration.String(),
		},
		"results": results,
	}

	evalJSON, err := json.MarshalIndent(evalData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal evaluation results: %w", err)
	}

	if err := os.WriteFile(evalFilePath, evalJSON, 0o644); err != nil {
		return fmt.Errorf("failed to save evaluation results: %w", err)
	}

	fmt.Println()
	fmt.Printf("✓ Evaluated %d decks in %s\n", len(results), evalDuration.Round(time.Millisecond))
	fmt.Printf("  Evaluation results: %s\n", evalFilePath)
	fmt.Println()

	// ========================================================================
	// PHASE 3: Compare top performers
	// ========================================================================
	fmt.Println("🏆 PHASE 3: Comparing top performers...")
	fmt.Println("─────────────────────────────────────────────────────────────────────")

	// Select top N decks
	compareCount := topN
	if compareCount > len(results) {
		compareCount = len(results)
	}
	if compareCount < 2 {
		compareCount = 2
	}
	if compareCount > 5 {
		compareCount = 5
	}

	topResults := results[:compareCount]

	// Extract names and evaluation results for comparison
	var deckNames []string
	var evalResults []evaluation.EvaluationResult

	for _, res := range topResults {
		deckNames = append(deckNames, res.Name)
		evalResults = append(evalResults, res.Result)
	}

	// Generate comparison report
	reportsDir := filepath.Join(outputDir, "reports")
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	reportFileName := fmt.Sprintf("%s_deck_analysis_report_%s.md", timestamp, strings.TrimPrefix(playerTag, "#"))
	reportFilePath := filepath.Join(reportsDir, reportFileName)

	// Generate comprehensive markdown report
	reportContent := generateComparisonReport(deckNames, evalResults)

	if err := os.WriteFile(reportFilePath, []byte(reportContent), 0o644); err != nil {
		return fmt.Errorf("failed to save comparison report: %w", err)
	}

	fmt.Printf("✓ Generated comparison report for top %d decks\n", compareCount)
	fmt.Printf("  Report: %s\n", reportFilePath)
	fmt.Println()

	// ========================================================================
	// SUMMARY
	// ========================================================================
	fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                      ANALYSIS SUITE COMPLETE                       ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Player: %s (%s)\n", playerName, playerTag)
	fmt.Printf("Decks built: %d\n", successCount)
	fmt.Printf("Decks evaluated: %d\n", len(results))
	fmt.Printf("Top performers compared: %d\n", compareCount)
	fmt.Println()
	fmt.Println("📂 Output files:")
	fmt.Printf("  • Suite summary:  %s\n", suiteSummaryPath)
	fmt.Printf("  • Evaluations:    %s\n", evalFilePath)
	fmt.Printf("  • Final report:   %s\n", reportFilePath)
	fmt.Println()

	if len(results) > 0 {
		fmt.Println("🥇 Top 3 decks:")
		for i := 0; i < 3 && i < len(results); i++ {
			medal := []string{"🥇", "🥈", "🥉"}[i]
			res := results[i]
			fmt.Printf("  %s %s: %.2f (%.2f avg elixir, %s archetype)\n",
				medal, res.Name, res.Result.OverallScore, res.Result.AvgElixir,
				res.Result.DetectedArchetype)
		}
	}

	return nil
}

func deckAnalyzeCommand(ctx context.Context, cmd *cli.Command) error {
	cardNames := cmd.StringSlice("cards")

	if len(cardNames) != 8 {
		return fmt.Errorf("exactly 8 cards are required for deck analysis")
	}

	fmt.Printf("Analyzing deck with cards: %v\n", cardNames)
	fmt.Println("Note: Full deck analysis not yet implemented")

	return nil
}

func deckOptimizeCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	cardNames := cmd.StringSlice("cards")
	apiToken := cmd.String("api-token")
	dataDir := cmd.String("data-dir")
	verbose := cmd.Bool("verbose")
	_ = cmd.Int("max-changes")   // TODO: implement max-changes filtering
	_ = cmd.Bool("keep-win-con") // TODO: implement win condition preservation
	exportCSV := cmd.Bool("export-csv")

	if apiToken == "" {
		return fmt.Errorf("API token is required")
	}

	client := clashroyale.NewClient(apiToken)

	// If no cards provided, fetch player's current deck from API
	if len(cardNames) == 0 {
		if verbose {
			fmt.Printf("Fetching player data for tag: %s\n", tag)
		}

		player, err := client.GetPlayer(tag)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}

		if len(player.CurrentDeck) == 0 {
			return fmt.Errorf("player %s has no current deck configured", tag)
		}

		if len(player.CurrentDeck) != 8 {
			return fmt.Errorf("player's current deck has %d cards, expected 8", len(player.CurrentDeck))
		}

		// Extract card names from CurrentDeck
		cardNames = make([]string, len(player.CurrentDeck))
		for i, card := range player.CurrentDeck {
			cardNames[i] = card.Name
		}

		if verbose {
			fmt.Printf("Using player's current deck: %v\n", cardNames)
		}
	} else if len(cardNames) != 8 {
		return fmt.Errorf("exactly 8 cards are required for optimization")
	}

	// Fetch player data for context
	player, err := client.GetPlayer(tag)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Convert card names to CardCandidates
	deckCards := convertToCardCandidates(cardNames)

	// Load synergy database
	synergyDB := deck.NewSynergyDatabase()

	// Create player context
	playerContext := evaluation.NewPlayerContextFromPlayer(player)

	// Evaluate current deck
	if verbose {
		fmt.Println("Evaluating current deck...")
	}
	currentResult := evaluation.Evaluate(deckCards, synergyDB, playerContext)

	// Convert player cards to map for GenerateAlternatives
	playerCardMap := make(map[string]bool)
	for _, card := range player.Cards {
		playerCardMap[card.Name] = true
	}

	// Generate alternative suggestions
	if verbose {
		fmt.Println("Generating optimization suggestions...")
	}
	alternatives := evaluation.GenerateAlternatives(deckCards, synergyDB, 10, playerCardMap)

	// Display current deck analysis
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    DECK OPTIMIZATION REPORT                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("🃏 Current Deck: %s\n", strings.Join(cardNames, " • "))
	fmt.Printf("📊 Average Elixir: %.2f\n", currentResult.AvgElixir)
	fmt.Printf("🎯 Archetype: %s (%.0f%% confidence)\n",
		strings.Title(string(currentResult.DetectedArchetype)),
		currentResult.ArchetypeConfidence*100)
	fmt.Println()
	fmt.Printf("⭐ Current Overall Score: %.1f/10 - %s\n",
		currentResult.OverallScore,
		currentResult.OverallRating)
	fmt.Println()

	// Display current category scores
	fmt.Println("Current Category Scores:")
	fmt.Printf("  ⚔️  Attack:        %s  %.1f/10 - %s\n",
		formatStars(currentResult.Attack.Stars),
		currentResult.Attack.Score,
		currentResult.Attack.Rating)
	fmt.Printf("  🛡️  Defense:       %s  %.1f/10 - %s\n",
		formatStars(currentResult.Defense.Stars),
		currentResult.Defense.Score,
		currentResult.Defense.Rating)
	fmt.Printf("  🔗 Synergy:       %s  %.1f/10 - %s\n",
		formatStars(currentResult.Synergy.Stars),
		currentResult.Synergy.Score,
		currentResult.Synergy.Rating)
	fmt.Printf("  🎭 Versatility:   %s  %.1f/10 - %s\n",
		formatStars(currentResult.Versatility.Stars),
		currentResult.Versatility.Score,
		currentResult.Versatility.Rating)
	fmt.Printf("  💰 F2P Friendly:  %s  %.1f/10 - %s\n",
		formatStars(currentResult.F2PFriendly.Stars),
		currentResult.F2PFriendly.Score,
		currentResult.F2PFriendly.Rating)
	fmt.Println()

	// Display optimization suggestions
	if len(alternatives.Suggestions) == 0 {
		fmt.Println("✨ Your deck is already well-optimized! No better alternatives found.")
		return nil
	}

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("                OPTIMIZATION SUGGESTIONS (%d found)\n", len(alternatives.Suggestions))
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Display top suggestions
	displayCount := len(alternatives.Suggestions)
	if displayCount > 5 {
		displayCount = 5
	}

	for i, alt := range alternatives.Suggestions[:displayCount] {
		fmt.Printf("Suggestion #%d: %s\n", i+1, alt.Impact)
		fmt.Println("───────────────────────────────────────────────────────────────")
		fmt.Printf("  Replace: %s  →  %s\n", alt.OriginalCard, alt.ReplacementCard)
		fmt.Printf("  Score Improvement: %.1f → %.1f (+%.1f)\n",
			alt.OriginalScore, alt.NewScore, alt.ScoreDelta)
		fmt.Printf("  Rationale: %s\n", alt.Rationale)
		fmt.Printf("  New Deck: %s\n", strings.Join(alt.Deck, " • "))
		fmt.Println()
	}

	// CSV export if requested
	if exportCSV {
		csvPath := filepath.Join(dataDir, fmt.Sprintf("deck-optimize-%s-%d.csv", tag, time.Now().Unix()))
		if err := exportOptimizationCSV(csvPath, tag, cardNames, currentResult, *alternatives); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to export CSV: %v\n", err)
		} else {
			fmt.Printf("✓ Optimization results exported to: %s\n", csvPath)
		}
	}

	return nil
}

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
			fmt.Printf("Generating recommendations from offline analysis for player %s\n", tag)
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
				fmt.Printf("Loaded analysis from: %s\n", analysisFile)
			}
		} else {
			// Load latest analysis for player tag
			loadedAnalysis, err = builder.LoadLatestAnalysis(tag, analysisDir)
			if err != nil {
				return fmt.Errorf("failed to load analysis for player %s from %s: %w", tag, analysisDir, err)
			}
			if verbose {
				fmt.Printf("Loaded latest analysis from: %s\n", analysisDir)
			}
		}

		deckCardAnalysis = *loadedAnalysis
		playerTag = tag
		playerName = tag // Use tag as name in offline mode
	} else {
		// ONLINE MODE: Fetch from API
		if apiToken == "" {
			return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag. Use --from-analysis for offline mode.")
		}

		client := clashroyale.NewClient(apiToken)

		if verbose {
			fmt.Printf("Generating recommendations for player %s\n", tag)
		}

		// Get player information
		player, err := client.GetPlayer(tag)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}

		playerName = player.Name
		playerTag = player.Tag

		if verbose {
			fmt.Printf("Player: %s (%s)\n", player.Name, player.Tag)
			fmt.Printf("Analyzing %d cards...\n", len(player.Cards))
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
		fmt.Printf("\nExported to CSV: %s\n", getRecommendationsCSVPath(dataDir, playerTag))
	}

	return nil
}

// displayRecommendations displays deck recommendations in a formatted table
func displayRecommendations(result *recommend.RecommendationResult, verbose bool) {
	fmt.Printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              DECK RECOMMENDATIONS                                  ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Player: %s (%s)\n", result.PlayerName, result.PlayerTag)
	if result.TopArchetype != "" {
		fmt.Printf("Top Archetype Match: %s\n", result.TopArchetype)
	}
	fmt.Printf("Generated: %s\n", result.GeneratedAt)
	fmt.Printf("\n")

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

	fmt.Printf("Deck #%d [%s - %s]\n", rank, rec.ArchetypeName, typeLabel)
	fmt.Printf("─────────────────────────────────────────────────────\n")
	fmt.Printf("Compatibility: %.1f%% | Synergy: %.1f%% | Overall: %.1f%%\n",
		rec.CompatibilityScore, rec.SynergyScore, rec.OverallScore)
	fmt.Printf("Avg Elixir: %.2f\n", rec.Deck.AvgElixir)

	// Display cards in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Card\t\tLevel\tElixir\n")
	for _, card := range rec.Deck.DeckDetail {
		fmt.Fprintf(w, "%s\t\t%d\t%d\n", card.Name, card.Level, card.Elixir)
	}
	w.Flush()

	// Display evolution slots if any
	if len(rec.Deck.EvolutionSlots) > 0 {
		fmt.Printf("Evolution Slots: %s\n", strings.Join(rec.Deck.EvolutionSlots, ", "))
	}

	// Display reasons
	if len(rec.Reasons) > 0 {
		fmt.Printf("\nWhy Recommended:\n")
		for _, reason := range rec.Reasons {
			fmt.Printf("  • %s\n", reason)
		}
	}

	if verbose {
		// Display upgrade cost if available
		if rec.UpgradeCost.CardsNeeded > 0 {
			fmt.Printf("\nUpgrade needed: %d cards, %d gold\n",
				rec.UpgradeCost.CardsNeeded, rec.UpgradeCost.GoldNeeded)
		}

		// Display notes
		if len(rec.Deck.Notes) > 0 {
			fmt.Printf("\nNotes:\n")
			for _, note := range rec.Deck.Notes {
				fmt.Printf("  • %s\n", note)
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
	defer file.Close()

	// Write header
	header := []string{
		"Rank", "Archetype", "Type", "Compatibility", "Synergy", "Overall",
		"AvgElixir", "Cards", "EvolutionSlots", "Reasons",
	}
	if _, err := file.WriteString(strings.Join(header, ",") + "\n"); err != nil {
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
		if _, err := file.WriteString(strings.Join(row, ",") + "\n"); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// getRecommendationsCSVPath returns the CSV file path for recommendations
func getRecommendationsCSVPath(dataDir, playerTag string) string {
	return filepath.Join(dataDir, "csv", fmt.Sprintf("recommendations_%s.csv", playerTag))
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
		fmt.Printf("Generating mulligan guide for deck: %s\n", deckName)
		fmt.Printf("Cards: %v\n", cardNames)
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
			fmt.Printf("\nSaving mulligan guide to: %s\n", dataDir)
		}
		if err := saveMulliganGuide(dataDir, guide); err != nil {
			fmt.Printf("Warning: Failed to save mulligan guide: %v\n", err)
		} else {
			fmt.Printf("\nMulligan guide saved to file\n")
		}
	}

	return nil
}

// displayMulliganGuide displays a formatted mulligan guide
func displayMulliganGuide(guide *mulligan.MulliganGuide) {
	fmt.Printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                    MULLIGAN GUIDE - OPENING PLAYS               ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Deck: %s (%s)\n", guide.DeckName, guide.Archetype.String())
	fmt.Printf("Generated: %s\n\n", guide.GeneratedAt.Format("2006-01-02 15:04:05"))

	fmt.Printf("📋 General Principles:\n")
	for _, principle := range guide.GeneralPrinciples {
		fmt.Printf("   • %s\n", principle)
	}
	fmt.Println()

	fmt.Printf("🃏 Deck Composition:\n")
	fmt.Printf("   Cards: %s\n", strings.Join(guide.DeckCards, ", "))
	fmt.Println()

	if len(guide.IdealOpenings) > 0 {
		fmt.Printf("✅ Ideal Opening Cards:\n")
		for _, opening := range guide.IdealOpenings {
			fmt.Printf("   ✓ %s\n", opening)
		}
		fmt.Println()
	}

	if len(guide.NeverOpenWith) > 0 {
		fmt.Printf("❌ Never Open With:\n")
		for _, never := range guide.NeverOpenWith {
			fmt.Printf("   ✗ %s\n", never)
		}
		fmt.Println()
	}

	fmt.Printf("🎮 Matchup-Specific Openings:\n\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for i, matchup := range guide.Matchups {
		fmt.Printf("%d. VS %s\n", i+1, matchup.OpponentType)
		fmt.Printf("   ▶ Opening Play: %s\n", matchup.OpeningPlay)
		fmt.Printf("   ▶ Why: %s\n", matchup.Reason)
		fmt.Printf("   ▶ Backup: %s\n", matchup.Backup)
		fmt.Printf("   ▶ Key Cards: %s\n", strings.Join(matchup.KeyCards, ", "))
		fmt.Printf("   ▶ Danger Level: %s\n", matchup.DangerLevel)
		fmt.Println()
	}
	w.Flush()
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

// saveMulliganGuide saves the mulligan guide to a JSON file
func saveMulliganGuide(dataDir string, guide *mulligan.MulliganGuide) error {
	// Create mulligan directory if it doesn't exist
	mulliganDir := filepath.Join(dataDir, "mulligan")
	if err := os.MkdirAll(mulliganDir, 0o755); err != nil {
		return fmt.Errorf("failed to create mulligan directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := guide.GeneratedAt.Format("20060102_150405")
	filename := filepath.Join(mulliganDir, fmt.Sprintf("%s_%s.json", guide.DeckName, timestamp))

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

// displayDeckRecommendationOffline displays a formatted deck recommendation without full player object
// Used for offline mode where we only have player name and tag as strings
func displayDeckRecommendationOffline(rec *deck.DeckRecommendation, playerName, playerTag string) {
	fmt.Printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              RECOMMENDED 1v1 LADDER DECK                           ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Player: %s (%s)\n", playerName, playerTag)
	fmt.Printf("Average Elixir: %.2f\n", rec.AvgElixir)

	// Display combat stats information if available
	if combatWeight := os.Getenv("COMBAT_STATS_WEIGHT"); combatWeight != "" {
		if combatWeight == "0" {
			fmt.Printf("Scoring: Traditional only (combat stats disabled)\n")
		} else {
			fmt.Printf("Scoring: %.0f%% traditional, %.0f%% combat stats\n",
				(1-mustParseFloat(combatWeight))*100,
				mustParseFloat(combatWeight)*100)
		}
	}
	fmt.Printf("\n")

	// Display deck cards in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "#\tCard\tLevel\t\tElixir\tRole\n")
	fmt.Fprintf(w, "─\t────\t─────\t\t──────\t────\n")

	for i, card := range rec.DeckDetail {
		evoBadge := deck.FormatEvolutionBadge(card.EvolutionLevel)
		levelStr := fmt.Sprintf("%d/%d", card.Level, card.MaxLevel)
		if evoBadge != "" {
			levelStr = fmt.Sprintf("%s (%s)", levelStr, evoBadge)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%s\n",
			i+1,
			card.Name,
			levelStr,
			card.Elixir,
			card.Role)
	}
	w.Flush()

	// Display strategic notes
	if len(rec.Notes) > 0 {
		fmt.Printf("\nStrategic Notes:\n")
		fmt.Printf("════════════════\n")
		for _, note := range rec.Notes {
			fmt.Printf("• %s\n", note)
		}
	}
}

// displayUpgradeRecommendations displays upgrade recommendations in a formatted table
func displayUpgradeRecommendations(upgrades *deck.UpgradeRecommendations) {
	fmt.Printf("╔════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              UPGRADE RECOMMENDATIONS                                ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	if len(upgrades.Recommendations) == 0 {
		fmt.Println("No upgrade recommendations - all cards are at max level!")
		return
	}

	fmt.Printf("Total Gold Needed: %d\n\n", upgrades.TotalGoldNeeded)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "#\tCard\t\tLevel\t\tRarity\t\tImpact\tGold\t\tValue/1k\n")
	fmt.Fprintf(w, "─\t────\t\t─────\t\t──────\t\t──────\t────\t\t────────\n")

	for i, rec := range upgrades.Recommendations {
		goldDisplay := fmt.Sprintf("%dk", rec.GoldCost/1000)
		if rec.GoldCost < 1000 {
			goldDisplay = fmt.Sprintf("%d", rec.GoldCost)
		}

		fmt.Fprintf(w, "%d\t%s\t\t%d->%d\t\t%s\t\t%.1f\t%s\t\t%.2f\n",
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
	w.Flush()

	// Display reasons
	fmt.Printf("\nWhy These Upgrades:\n")
	fmt.Printf("──────────────────\n")
	for i, rec := range upgrades.Recommendations {
		if i >= 3 {
			fmt.Printf("... and %d more\n", len(upgrades.Recommendations)-3)
			break
		}
		fmt.Printf("%d. %s: %s\n", i+1, rec.CardName, rec.Reason)
	}
}

// mustParseFloat parses a float from a string, returning 0 if parsing fails
func mustParseFloat(s string) float64 {
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}
	return 0
}

// buildAllStrategies builds decks using all available strategies and displays them for comparison
func buildAllStrategies(ctx context.Context, cmd *cli.Command, builder *deck.Builder, cardAnalysis deck.CardAnalysis, playerName, playerTag string) error {
	verbose := cmd.Bool("verbose")
	excludeCards := cmd.StringSlice("exclude-cards")

	// All available strategies
	strategies := []deck.Strategy{
		deck.StrategyBalanced,
		deck.StrategyAggro,
		deck.StrategyControl,
		deck.StrategyCycle,
		deck.StrategySplash,
		deck.StrategySpell,
	}

	fmt.Printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              ALL DECK BUILDING STRATEGIES                          ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Player: %s (%s)\n\n", playerName, playerTag)

	// Build a deck for each strategy
	for i, strategy := range strategies {
		// Create a new builder for this strategy
		strategyBuilder := deck.NewBuilder(cmd.String("data-dir"))

		// Copy builder configuration
		if unlockedEvos := cmd.String("unlocked-evolutions"); unlockedEvos != "" {
			strategyBuilder.SetUnlockedEvolutions(strings.Split(unlockedEvos, ","))
		}
		if slots := cmd.Int("evolution-slots"); slots > 0 {
			strategyBuilder.SetEvolutionSlotLimit(slots)
		}
		if enableSynergy := cmd.Bool("enable-synergy"); enableSynergy {
			strategyBuilder.SetSynergyEnabled(true)
			if synergyWeight := cmd.Float64("synergy-weight"); synergyWeight > 0 {
				strategyBuilder.SetSynergyWeight(synergyWeight)
			}
		}

		// Set the strategy
		if err := strategyBuilder.SetStrategy(strategy); err != nil {
			fmt.Printf("⚠ Failed to set strategy %s: %v\n\n", strategy, err)
			continue
		}

		// Apply exclude filter
		filteredAnalysis := cardAnalysis
		if len(excludeCards) > 0 {
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
			filteredAnalysis.CardLevels = filteredLevels
		}

		// Build deck
		deckRec, err := strategyBuilder.BuildDeckFromAnalysis(filteredAnalysis)
		if err != nil {
			fmt.Printf("⚠ Failed to build deck for strategy %s: %v\n\n", strategy, err)
			continue
		}

		// Display the deck with strategy label
		displayStrategyDeck(i+1, strategy, deckRec, verbose)
	}

	return nil
}

// displayStrategyDeck displays a single deck with its strategy label
func displayStrategyDeck(rank int, strategy deck.Strategy, rec *deck.DeckRecommendation, verbose bool) {
	fmt.Printf("Strategy #%d: %s\n", rank, strings.ToUpper(string(strategy)))
	fmt.Printf("═══════════════════════════════════════════════════════════════════\n")
	fmt.Printf("Average Elixir: %.2f\n\n", rec.AvgElixir)

	// Display deck cards in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "#\tCard\tLevel\t\tElixir\tRole\n")
	fmt.Fprintf(w, "─\t────\t─────\t\t──────\t────\n")

	for i, card := range rec.DeckDetail {
		evoBadge := deck.FormatEvolutionBadge(card.EvolutionLevel)
		levelStr := fmt.Sprintf("%d/%d", card.Level, card.MaxLevel)
		if evoBadge != "" {
			levelStr = fmt.Sprintf("%s (%s)", levelStr, evoBadge)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%s\n",
			i+1,
			card.Name,
			levelStr,
			card.Elixir,
			card.Role)
	}
	w.Flush()

	// Display strategic notes if verbose
	if verbose && len(rec.Notes) > 0 {
		fmt.Printf("\nStrategic Notes:\n")
		for _, note := range rec.Notes {
			fmt.Printf("• %s\n", note)
		}
	}

	fmt.Printf("\n")
}

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
		fmt.Printf("Finding budget-optimized decks for player %s\n", tag)
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
	deckCardAnalysis := deck.CardAnalysis{
		CardLevels:   make(map[string]deck.CardLevelData),
		AnalysisTime: cardAnalysis.AnalysisTime.Format(time.RFC3339),
	}

	for cardName, cardInfo := range cardAnalysis.CardLevels {
		deckCardAnalysis.CardLevels[cardName] = deck.CardLevelData{
			Level:             cardInfo.Level,
			MaxLevel:          cardInfo.MaxLevel,
			Rarity:            cardInfo.Rarity,
			Elixir:            cardInfo.Elixir,
			MaxEvolutionLevel: cardInfo.MaxEvolutionLevel,
		}
	}

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
			fmt.Printf("\nSaving budget analysis to: %s\n", dataDir)
		}
		if err := saveBudgetResult(dataDir, result); err != nil {
			fmt.Printf("Warning: Failed to save budget analysis: %v\n", err)
		} else {
			fmt.Printf("\nBudget analysis saved to file\n")
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
	fmt.Printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              BUDGET-OPTIMIZED DECK FINDER                          ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Player: %s (%s)\n", result.PlayerName, result.PlayerTag)
	fmt.Printf("Average Card Level: %.2f\n\n", result.Summary.PlayerAverageLevel)

	// Display summary
	fmt.Printf("Summary:\n")
	fmt.Printf("════════\n")
	fmt.Printf("Total Decks Analyzed:    %d\n", result.Summary.TotalDecksAnalyzed)
	fmt.Printf("Ready Decks:             %d\n", result.Summary.ReadyDeckCount)
	fmt.Printf("Quick Win Decks:         %d\n", result.Summary.QuickWinCount)
	fmt.Printf("Best ROI:                %.4f\n", result.Summary.BestROI)
	fmt.Printf("Lowest Cards Needed:     %d\n", result.Summary.LowestCardsNeeded)
	fmt.Printf("\n")

	// Display quick wins if available
	if len(result.QuickWins) > 0 {
		fmt.Printf("Quick Wins (1-2 upgrades away):\n")
		fmt.Printf("════════════════════════════════\n")
		for i, analysis := range result.QuickWins {
			if i >= 3 {
				break // Show top 3 quick wins
			}
			displayBudgetDeckSummary(i+1, analysis)
		}
		fmt.Printf("\n")
	}

	// Display all decks
	if len(result.AllDecks) > 0 {
		fmt.Printf("Top Decks (sorted by %s):\n", options.SortBy)
		fmt.Printf("═════════════════════════════════════\n\n")

		for i, analysis := range result.AllDecks {
			displayBudgetDeckDetail(i+1, analysis)
		}
	} else {
		fmt.Printf("No decks found matching criteria.\n")
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

	fmt.Printf("#%d: %s\n", rank, strings.Join(cards[:min(3, len(cards))], ", ")+"...")
	fmt.Printf("    Cards Needed: %d | Gold: %d | ROI: %.4f\n",
		analysis.TotalCardsNeeded, analysis.TotalGoldNeeded, analysis.ROI)
}

// displayBudgetDeckDetail displays detailed deck information
func displayBudgetDeckDetail(rank int, analysis *budget.DeckBudgetAnalysis) {
	if analysis.Deck == nil {
		return
	}

	fmt.Printf("Deck #%d [%s]\n", rank, analysis.BudgetCategory)
	fmt.Printf("─────────────────────────────────────\n")

	// Deck cards table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Card\tLevel\t\tElixir\tRole\n")
	fmt.Fprintf(w, "────\t─────\t\t──────\t────\n")

	for _, card := range analysis.Deck.DeckDetail {
		evoBadge := deck.FormatEvolutionBadge(card.EvolutionLevel)
		levelStr := fmt.Sprintf("%d/%d", card.Level, card.MaxLevel)
		if evoBadge != "" {
			levelStr = fmt.Sprintf("%s (%s)", levelStr, evoBadge)
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
			card.Name,
			levelStr,
			card.Elixir,
			card.Role)
	}
	w.Flush()

	fmt.Printf("\n")
	fmt.Printf("Average Elixir: %.2f\n", analysis.Deck.AvgElixir)
	fmt.Printf("Current Score: %.4f | Projected Score: %.4f\n",
		analysis.CurrentScore, analysis.ProjectedScore)
	fmt.Printf("Cards Needed: %d | Gold Needed: %d\n",
		analysis.TotalCardsNeeded, analysis.TotalGoldNeeded)
	fmt.Printf("ROI: %.4f | Cost Efficiency: %.4f\n",
		analysis.ROI, analysis.CostEfficiency)

	// Display upgrade priorities if there are upgrades needed
	if len(analysis.CardUpgrades) > 0 {
		fmt.Printf("\nUpgrade Priorities:\n")
		for i, upgrade := range analysis.CardUpgrades {
			if i >= 3 {
				fmt.Printf("  ... and %d more\n", len(analysis.CardUpgrades)-3)
				break
			}
			fmt.Printf("  %d. %s: Level %d -> %d (%d cards, %d gold)\n",
				i+1, upgrade.CardName, upgrade.CurrentLevel, upgrade.TargetLevel,
				upgrade.CardsNeeded, upgrade.GoldNeeded)
		}
	}

	fmt.Printf("\n")
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

	fmt.Printf("Budget analysis saved to: %s\n", filename)
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

	// Parse deck cards
	var deckCardNames []string
	if deckString != "" {
		// Parse deck string (cards separated by dashes)
		deckCardNames = parseDeckString(deckString)
		if len(deckCardNames) != 8 {
			return fmt.Errorf("deck must contain exactly 8 cards, got %d", len(deckCardNames))
		}
	} else {
		// Load deck from analysis file
		loadedCards, err := loadDeckFromAnalysis(fromAnalysis)
		if err != nil {
			return fmt.Errorf("failed to load deck from analysis: %w", err)
		}
		deckCardNames = loadedCards
	}

	if verbose {
		fmt.Printf("Evaluating deck: %v\n", deckCardNames)
		fmt.Printf("Output format: %s\n", format)
	}

	// Convert card names to CardCandidates with default card data
	deckCards := convertToCardCandidates(deckCardNames)

	// Create synergy database
	synergyDB := deck.NewSynergyDatabase()

	// Fetch player context if --tag is provided
	var playerContext *evaluation.PlayerContext
	if playerTag != "" && apiToken != "" {
		if verbose {
			fmt.Printf("Fetching player data for context-aware evaluation...\n")
		}

		client := clashroyale.NewClient(apiToken)
		player, err := client.GetPlayer(playerTag)
		if err != nil {
			// Log warning but continue with evaluation without context
			fmt.Fprintf(os.Stderr, "Warning: Failed to fetch player data: %v\n", err)
			fmt.Fprintf(os.Stderr, "Continuing with evaluation without player context.\n")
		} else {
			playerContext = evaluation.NewPlayerContextFromPlayer(player)
			if verbose {
				fmt.Printf("Player context loaded: %s (%s), Arena: %s\n",
					player.Name, player.Tag, player.Arena.Name)
			}
		}
	}

	// Evaluate the deck
	result := evaluation.Evaluate(deckCards, synergyDB, playerContext)

	// Save to persistent storage if player tag is available
	if playerTag != "" {
		storage, err := leaderboard.NewStorage(playerTag)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to initialize storage: %v\n", err)
			}
		} else {
			defer func() {
				if err := storage.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to close storage: %v\n", err)
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
					fmt.Fprintf(os.Stderr, "Warning: failed to save deck to storage: %v\n", err)
				}
			} else {
				if _, err := storage.RecalculateStats(); err != nil && verbose {
					fmt.Fprintf(os.Stderr, "Warning: failed to recalculate stats: %v\n", err)
				}
				if verbose {
					if isNew {
						fmt.Printf("Saved deck to storage (ID: %d) at: %s\n", deckID, storage.GetDBPath())
					} else {
						fmt.Printf("Updated existing deck in storage (ID: %d)\n", deckID)
					}
				}
			}
		}
	}

	// Format output based on requested format
	var formattedOutput string
	var err error
	switch strings.ToLower(format) {
	case "human":
		formattedOutput = evaluation.FormatHuman(&result)
	case "json":
		formattedOutput, err = evaluation.FormatJSON(&result)
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
	case "csv":
		formattedOutput = evaluation.FormatCSV(&result)
	case "detailed":
		formattedOutput = evaluation.FormatDetailed(&result)
	default:
		return fmt.Errorf("unknown format: %s (supported: human, json, csv, detailed)", format)
	}

	// Output to file or stdout
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(formattedOutput), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		if verbose {
			fmt.Printf("Evaluation saved to: %s\n", outputFile)
		}
	} else {
		fmt.Print(formattedOutput)
	}

	// Perform upgrade impact analysis if requested
	if showUpgradeImpact {
		// Only for human output format (not applicable to JSON/CSV)
		if format == "human" || format == "detailed" {
			if err := performDeckUpgradeImpactAnalysis(deckCardNames, playerTag, topUpgrades, apiToken, dataDir, verbose); err != nil {
				// Log error but don't fail the entire command
				fmt.Fprintf(os.Stderr, "\nWarning: Failed to perform upgrade impact analysis: %v\n", err)
			}
		} else if verbose {
			fmt.Fprintf(os.Stderr, "\nNote: Upgrade impact analysis only available for human and detailed output formats\n")
		}
	}

	return nil
}

// performDeckUpgradeImpactAnalysis performs upgrade impact analysis for a specific deck
// It fetches the player's card levels and shows which deck card upgrades would have the most impact
func performDeckUpgradeImpactAnalysis(deckCardNames []string, playerTag string, topN int, apiToken, dataDir string, verbose bool) error {
	// Create client to fetch player data
	client := clashroyale.NewClient(apiToken)

	if verbose {
		fmt.Printf("\nFetching player data for upgrade impact analysis...\n")
	}

	// Get player information
	player, err := client.GetPlayer(playerTag)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	if verbose {
		fmt.Printf("Player: %s (%s)\n", player.Name, player.Tag)
		fmt.Printf("Analyzing deck: %v\n", deckCardNames)
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
		isKeyUpgrade := cardData.Rarity == "Legendary" || cardData.Rarity == "Champion"

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
	case "Common":
		rarityMultiplier = 1.0
	case "Rare":
		rarityMultiplier = 2.0
	case "Epic":
		rarityMultiplier = 4.0
	case "Legendary":
		rarityMultiplier = 8.0
	case "Champion":
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
	// Simple bubble sort (for small lists)
	n := len(impacts)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if impacts[j].ImpactScore < impacts[j+1].ImpactScore {
				impacts[j], impacts[j+1] = impacts[j+1], impacts[j]
			}
		}
	}
}

// displayDeckUpgradeImpactAnalysis displays the upgrade impact analysis for deck cards
func displayDeckUpgradeImpactAnalysis(deckCardNames []string, impacts []DeckCardUpgrade, topN int, player *clashroyale.Player) {
	fmt.Printf("\n")
	fmt.Printf("╔════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                    UPGRADE IMPACT ANALYSIS                          ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Player: %s (%s)\n", player.Name, player.Tag)
	fmt.Printf("Deck: %v\n\n", deckCardNames)

	if len(impacts) == 0 {
		fmt.Printf("✨ All deck cards are already at max level!\n")
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

	fmt.Printf("Summary:\n")
	fmt.Printf("════════\n")
	fmt.Printf("Upgradable Cards: %d\n", len(impacts))
	fmt.Printf("Top %d Upgrades: %d gold total\n\n", displayCount, totalGold)

	fmt.Printf("Most Impactful Upgrades:\n")
	fmt.Printf("════════════════════════\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "#\tCard\tLevel\t\tRarity\t\tImpact\tGold\t\tCards\n")
	fmt.Fprintf(w, "─\t────\t─────\t\t──────\t\t──────\t────\t\t─────\n")

	for i := 0; i < displayCount; i++ {
		upgrade := impacts[i]
		keyMarker := ""
		if upgrade.IsKeyUpgrade {
			keyMarker = " ⭐"
		}

		goldDisplay := formatGoldCost(upgrade.GoldCost)
		fmt.Fprintf(w, "%d\t%s%s\t%d->%d\t\t%s\t\t%.1f\t%s\t\t%d\n",
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
	w.Flush()

	fmt.Printf("\n")
	fmt.Printf("💡 Tip: Focus on upgrading cards with the highest impact score first.\n")
	fmt.Printf("   Win conditions and Legendary/Champion cards typically provide the best ROI.\n")
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

// exportOptimizationCSV exports optimization results to a CSV file
func exportOptimizationCSV(
	path string,
	playerTag string,
	currentDeck []string,
	currentResult evaluation.EvaluationResult,
	alternatives evaluation.AlternativeSuggestions,
) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintln(file, "# DECK OPTIMIZATION RESULTS")
	fmt.Fprintf(file, "Player Tag,%s\n", playerTag)
	fmt.Fprintf(file, "Current Deck,%s\n", strings.Join(currentDeck, ";"))
	fmt.Fprintf(file, "Current Score,%.2f\n", currentResult.OverallScore)
	fmt.Fprintf(file, "Archetype,%s\n", currentResult.DetectedArchetype)
	fmt.Fprintln(file, "")

	// Write category scores
	fmt.Fprintln(file, "# CURRENT CATEGORY SCORES")
	fmt.Fprintln(file, "Category,Score,Rating,Stars")
	fmt.Fprintf(file, "Attack,%.2f,%s,%d\n",
		currentResult.Attack.Score,
		currentResult.Attack.Rating,
		currentResult.Attack.Stars)
	fmt.Fprintf(file, "Defense,%.2f,%s,%d\n",
		currentResult.Defense.Score,
		currentResult.Defense.Rating,
		currentResult.Defense.Stars)
	fmt.Fprintf(file, "Synergy,%.2f,%s,%d\n",
		currentResult.Synergy.Score,
		currentResult.Synergy.Rating,
		currentResult.Synergy.Stars)
	fmt.Fprintf(file, "Versatility,%.2f,%s,%d\n",
		currentResult.Versatility.Score,
		currentResult.Versatility.Rating,
		currentResult.Versatility.Stars)
	fmt.Fprintf(file, "F2P Friendly,%.2f,%s,%d\n",
		currentResult.F2PFriendly.Score,
		currentResult.F2PFriendly.Rating,
		currentResult.F2PFriendly.Stars)
	fmt.Fprintln(file, "")

	// Write optimization suggestions
	fmt.Fprintln(file, "# OPTIMIZATION SUGGESTIONS")
	fmt.Fprintln(file, "Rank,Original Card,Replacement Card,Score Before,Score After,Improvement,Impact,Rationale,New Deck")
	for i, alt := range alternatives.Suggestions {
		fmt.Fprintf(file, "%d,%s,%s,%.2f,%.2f,%.2f,%s,\"%s\",%s\n",
			i+1,
			alt.OriginalCard,
			alt.ReplacementCard,
			alt.OriginalScore,
			alt.NewScore,
			alt.ScoreDelta,
			alt.Impact,
			alt.Rationale,
			strings.Join(alt.Deck, ";"))
	}

	return nil
}

// sortEvaluationResults sorts batch evaluation results by the specified criteria
func sortEvaluationResults[T any](results []T, sortBy string) {
	type resultInterface interface {
		GetResult() evaluation.EvaluationResult
	}

	// Type assertion helper
	getResult := func(r T) evaluation.EvaluationResult {
		switch v := any(r).(type) {
		case struct {
			Name      string
			Strategy  string
			Deck      []string
			Result    evaluation.EvaluationResult
			FilePath  string
			Evaluated time.Time
			Duration  time.Duration
		}:
			return v.Result
		default:
			return evaluation.EvaluationResult{}
		}
	}

	// Sort based on criteria
	switch strings.ToLower(sortBy) {
	case "attack":
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if getResult(results[i]).Attack.Score < getResult(results[j]).Attack.Score {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	case "defense":
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if getResult(results[i]).Defense.Score < getResult(results[j]).Defense.Score {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	case "synergy":
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if getResult(results[i]).Synergy.Score < getResult(results[j]).Synergy.Score {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	case "versatility":
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if getResult(results[i]).Versatility.Score < getResult(results[j]).Versatility.Score {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	case "f2p", "f2p-friendly":
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if getResult(results[i]).F2PFriendly.Score < getResult(results[j]).F2PFriendly.Score {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	case "playability":
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if getResult(results[i]).Playability.Score < getResult(results[j]).Playability.Score {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	case "elixir":
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if getResult(results[i]).AvgElixir > getResult(results[j]).AvgElixir {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	default: // "overall"
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if getResult(results[i]).OverallScore < getResult(results[j]).OverallScore {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	}
}

// formatEvaluationBatchSummary formats batch evaluation results as a human-readable summary
func formatEvaluationBatchSummary[T any](results []T, totalDecks int, totalTime time.Duration, sortBy, playerName, playerTag string) string {
	// Type extraction helpers using reflection
	getName := func(r T) string {
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Struct {
			if field := v.FieldByName("Name"); field.IsValid() && field.Kind() == reflect.String {
				return field.String()
			}
		}
		return ""
	}
	getResult := func(r T) evaluation.EvaluationResult {
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Struct {
			if field := v.FieldByName("Result"); field.IsValid() {
				if result, ok := field.Interface().(evaluation.EvaluationResult); ok {
					return result
				}
			}
		}
		return evaluation.EvaluationResult{}
	}

	var buf strings.Builder

	buf.WriteString("╔═══════════════════════════════════════════════════════════════════════════════╗\n")
	buf.WriteString("║                        BATCH DECK EVALUATION RESULTS                          ║\n")
	buf.WriteString("╚═══════════════════════════════════════════════════════════════════════════════╝\n\n")

	if playerName != "" || playerTag != "" {
		buf.WriteString(fmt.Sprintf("Player: %s (%s)\n", playerName, playerTag))
	}
	buf.WriteString(fmt.Sprintf("Total Decks: %d | Evaluated: %d | Sorted by: %s\n", totalDecks, len(results), sortBy))
	buf.WriteString(fmt.Sprintf("Total Time: %v | Avg: %v\n\n", totalTime, totalTime/time.Duration(max(len(results), 1))))

	buf.WriteString("┌─────┬──────────────────────────────┬─────────┬────────┬────────┬────────┬──────────────┐\n")
	buf.WriteString("│ Rank│ Deck Name                    │ Overall │ Attack │ Defense│ Synergy│ Archetype    │\n")
	buf.WriteString("├─────┼──────────────────────────────┼─────────┼────────┼────────┼────────┼──────────────┤\n")

	for i, r := range results {
		name := getName(r)
		if name == "" {
			continue
		}
		if len(name) > 28 {
			name = name[:25] + "..."
		}
		result := getResult(r)
		archetype := string(result.DetectedArchetype)
		if len(archetype) > 12 {
			archetype = archetype[:9] + "..."
		}
		buf.WriteString(fmt.Sprintf("│ %3d │ %-28s │  %5.2f  │  %5.2f │  %5.2f │  %5.2f │ %-12s │\n",
			i+1, name,
			result.OverallScore,
			result.Attack.Score,
			result.Defense.Score,
			result.Synergy.Score,
			archetype))
	}

	buf.WriteString("└─────┴──────────────────────────────┴─────────┴────────┴────────┴────────┴──────────────┘\n")

	return buf.String()
}

// formatEvaluationBatchCSV formats batch evaluation results as CSV
func formatEvaluationBatchCSV[T any](results []T) string {
	// Type extraction helpers using reflection
	getName := func(r T) string {
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Struct {
			if field := v.FieldByName("Name"); field.IsValid() && field.Kind() == reflect.String {
				return field.String()
			}
		}
		return ""
	}
	getStrategy := func(r T) string {
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Struct {
			if field := v.FieldByName("Strategy"); field.IsValid() && field.Kind() == reflect.String {
				return field.String()
			}
		}
		return ""
	}
	getDeck := func(r T) []string {
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Struct {
			if field := v.FieldByName("Deck"); field.IsValid() && field.Kind() == reflect.Slice {
				if deck, ok := field.Interface().([]string); ok {
					return deck
				}
			}
		}
		return nil
	}
	getResult := func(r T) evaluation.EvaluationResult {
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Struct {
			if field := v.FieldByName("Result"); field.IsValid() {
				if result, ok := field.Interface().(evaluation.EvaluationResult); ok {
					return result
				}
			}
		}
		return evaluation.EvaluationResult{}
	}

	var buf strings.Builder

	// Header
	buf.WriteString("Rank,Name,Strategy,Overall,Attack,Defense,Synergy,Versatility,F2P,Playability,Archetype,Avg_Elixir,Deck\n")

	// Data rows
	for i, r := range results {
		name := getName(r)
		if name == "" {
			continue
		}
		strategy := getStrategy(r)
		deck := getDeck(r)
		result := getResult(r)
		buf.WriteString(fmt.Sprintf("%d,%s,%s,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%s,%.2f,\"%s\"\n",
			i+1,
			name,
			strategy,
			result.OverallScore,
			result.Attack.Score,
			result.Defense.Score,
			result.Synergy.Score,
			result.Versatility.Score,
			result.F2PFriendly.Score,
			result.Playability.Score,
			result.DetectedArchetype,
			result.AvgElixir,
			strings.Join(deck, " - ")))
	}

	return buf.String()
}

// formatEvaluationBatchDetailed formats batch evaluation results with detailed analysis
func formatEvaluationBatchDetailed[T any](results []T, playerName, playerTag string) string {
	// Type extraction helpers using reflection
	getName := func(r T) string {
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Struct {
			if field := v.FieldByName("Name"); field.IsValid() && field.Kind() == reflect.String {
				return field.String()
			}
		}
		return ""
	}
	getStrategy := func(r T) string {
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Struct {
			if field := v.FieldByName("Strategy"); field.IsValid() && field.Kind() == reflect.String {
				return field.String()
			}
		}
		return ""
	}
	getDeck := func(r T) []string {
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Struct {
			if field := v.FieldByName("Deck"); field.IsValid() && field.Kind() == reflect.Slice {
				if deck, ok := field.Interface().([]string); ok {
					return deck
				}
			}
		}
		return nil
	}
	getResult := func(r T) evaluation.EvaluationResult {
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Struct {
			if field := v.FieldByName("Result"); field.IsValid() {
				if result, ok := field.Interface().(evaluation.EvaluationResult); ok {
					return result
				}
			}
		}
		return evaluation.EvaluationResult{}
	}

	var buf strings.Builder

	buf.WriteString("╔═══════════════════════════════════════════════════════════════════════════════╗\n")
	buf.WriteString("║                    DETAILED BATCH EVALUATION RESULTS                          ║\n")
	buf.WriteString("╚═══════════════════════════════════════════════════════════════════════════════╝\n\n")

	if playerName != "" || playerTag != "" {
		buf.WriteString(fmt.Sprintf("Player: %s (%s)\n\n", playerName, playerTag))
	}

	for i, r := range results {
		name := getName(r)
		if name == "" {
			continue
		}
		strategy := getStrategy(r)
		deck := getDeck(r)
		result := getResult(r)

		buf.WriteString(fmt.Sprintf("═══════════════════ DECK #%d: %s ═══════════════════\n\n", i+1, name))

		if strategy != "" && strategy != "unknown" {
			buf.WriteString(fmt.Sprintf("Strategy: %s\n", strategy))
		}

		buf.WriteString(fmt.Sprintf("Deck: %s\n", strings.Join(deck, " - ")))
		buf.WriteString(fmt.Sprintf("Avg Elixir: %.2f\n", result.AvgElixir))
		buf.WriteString(fmt.Sprintf("Archetype: %s (%.1f%% confidence)\n\n", result.DetectedArchetype, result.ArchetypeConfidence*100))

		// Category scores
		buf.WriteString("SCORES:\n")
		buf.WriteString(fmt.Sprintf("  Overall:     %.2f (%s)\n", result.OverallScore, result.OverallRating))
		buf.WriteString(fmt.Sprintf("  Attack:      %.2f (%s)\n", result.Attack.Score, result.Attack.Rating))
		buf.WriteString(fmt.Sprintf("  Defense:     %.2f (%s)\n", result.Defense.Score, result.Defense.Rating))
		buf.WriteString(fmt.Sprintf("  Synergy:     %.2f (%s)\n", result.Synergy.Score, result.Synergy.Rating))
		buf.WriteString(fmt.Sprintf("  Versatility: %.2f (%s)\n", result.Versatility.Score, result.Versatility.Rating))
		buf.WriteString(fmt.Sprintf("  F2P:         %.2f (%s)\n", result.F2PFriendly.Score, result.F2PFriendly.Rating))
		buf.WriteString(fmt.Sprintf("  Playability: %.2f (%s)\n\n", result.Playability.Score, result.Playability.Rating))

		// Key assessments
		if result.Attack.Assessment != "" {
			buf.WriteString(fmt.Sprintf("Attack: %s\n", result.Attack.Assessment))
		}
		if result.Defense.Assessment != "" {
			buf.WriteString(fmt.Sprintf("Defense: %s\n", result.Defense.Assessment))
		}
		if result.Synergy.Assessment != "" {
			buf.WriteString(fmt.Sprintf("Synergy: %s\n", result.Synergy.Assessment))
		}

		buf.WriteString("\n" + strings.Repeat("─", 80) + "\n\n")
	}

	return buf.String()
}

func deckPossibleCountCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	format := cmd.String("format")
	verbose := cmd.Bool("verbose")
	outputFile := cmd.String("output")

	// Get API token
	apiToken := cmd.String("api-token")
	if apiToken == "" {
		return fmt.Errorf("API token required (set CLASH_ROYALE_API_TOKEN or use --api-token)")
	}

	// Fetch player data
	client := clashroyale.NewClient(apiToken)
	player, err := client.GetPlayer(tag)
	if err != nil {
		return fmt.Errorf("failed to get player data: %w", err)
	}

	// Create deck space calculator
	calculator, err := deck.NewDeckSpaceCalculator(player)
	if err != nil {
		return fmt.Errorf("failed to create calculator: %w", err)
	}

	// Calculate statistics
	stats := calculator.CalculateStats()

	// Format output
	var output string
	switch strings.ToLower(format) {
	case "json":
		output, err = formatPossibleCountJSON(player, stats)
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
	case "csv":
		output = formatPossibleCountCSV(player, stats, verbose)
	case "human":
		fallthrough
	default:
		output = formatPossibleCountHuman(player, stats, verbose)
	}

	// Output to file or stdout
	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(output), 0o644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Results saved to: %s\n", outputFile)
	} else {
		fmt.Print(output)
	}

	return nil
}

func formatPossibleCountHuman(player *clashroyale.Player, stats *deck.DeckSpaceStats, verbose bool) string {
	var buf strings.Builder

	buf.WriteString("╔════════════════════════════════════════════════════════════════════════╗\n")
	buf.WriteString("║               DECK COMBINATION CALCULATOR                              ║\n")
	buf.WriteString("╚════════════════════════════════════════════════════════════════════════╝\n\n")

	buf.WriteString(fmt.Sprintf("Player: %s (Tag: %s)\n", player.Name, player.Tag))
	buf.WriteString(fmt.Sprintf("Total Cards: %d\n\n", stats.TotalCards))

	buf.WriteString("═══ POSSIBLE DECK COMBINATIONS ═══\n\n")

	// Total combinations
	buf.WriteString(fmt.Sprintf("Total Unconstrained:  %s (%s)\n",
		stats.TotalCombinations.String(),
		deck.FormatLargeNumber(stats.TotalCombinations)))

	buf.WriteString(fmt.Sprintf("Valid (With Roles):   %s (%s)\n\n",
		stats.ValidCombinations.String(),
		deck.FormatLargeNumber(stats.ValidCombinations)))

	if verbose {
		// Cards by role
		buf.WriteString("═══ CARDS BY ROLE ═══\n\n")
		w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "Role\tCount\n")
		fmt.Fprintf(w, "────\t─────\n")

		// Print in a specific order
		roles := []deck.CardRole{
			deck.RoleWinCondition,
			deck.RoleBuilding,
			deck.RoleSpellBig,
			deck.RoleSpellSmall,
			deck.RoleSupport,
			deck.RoleCycle,
		}
		roleLabels := map[deck.CardRole]string{
			deck.RoleWinCondition: "Win Condition",
			deck.RoleBuilding:     "Building",
			deck.RoleSpellBig:     "Big Spell",
			deck.RoleSpellSmall:   "Small Spell",
			deck.RoleSupport:      "Support",
			deck.RoleCycle:        "Cycle",
		}

		for _, role := range roles {
			count := stats.CardsByRole[role]
			label := roleLabels[role]
			fmt.Fprintf(w, "%s\t%d\n", label, count)
		}
		w.Flush()
		buf.WriteString("\n")

		// Combinations by elixir range
		if len(stats.ByElixirRange) > 0 {
			buf.WriteString("═══ ESTIMATED BY ELIXIR RANGE ═══\n\n")
			w2 := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w2, "Range\tCombinations\n")
			fmt.Fprintf(w2, "─────\t────────────\n")

			for _, elixirRange := range deck.StandardElixirRanges {
				if count, exists := stats.ByElixirRange[elixirRange.Label]; exists {
					fmt.Fprintf(w2, "%s\t%s\n",
						elixirRange.Label,
						deck.FormatLargeNumber(count))
				}
			}
			w2.Flush()
			buf.WriteString("\n")
		}

		// Combinations by archetype
		if len(stats.ByArchetype) > 0 {
			buf.WriteString("═══ ESTIMATED BY ARCHETYPE ═══\n\n")
			w3 := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w3, "Archetype\tCombinations\n")
			fmt.Fprintf(w3, "─────────\t────────────\n")

			archetypes := []string{"Beatdown", "Control", "Cycle", "Siege", "Bridge Spam", "Bait"}
			for _, archetype := range archetypes {
				if count, exists := stats.ByArchetype[archetype]; exists {
					fmt.Fprintf(w3, "%s\t%s\n",
						archetype,
						deck.FormatLargeNumber(count))
				}
			}
			w3.Flush()
			buf.WriteString("\n")
		}
	}

	buf.WriteString("Note: Estimates for elixir ranges and archetypes are approximations.\n")
	buf.WriteString("Default deck composition: 1 win condition, 1 building, 1 big spell,\n")
	buf.WriteString("1 small spell, 2 support, 2 cycle.\n")

	return buf.String()
}

func formatPossibleCountJSON(player *clashroyale.Player, stats *deck.DeckSpaceStats) (string, error) {
	output := map[string]any{
		"player": map[string]string{
			"tag":  player.Tag,
			"name": player.Name,
		},
		"total_cards":              stats.TotalCards,
		"total_combinations":       stats.TotalCombinations.String(),
		"valid_combinations":       stats.ValidCombinations.String(),
		"total_combinations_human": deck.FormatLargeNumber(stats.TotalCombinations),
		"valid_combinations_human": deck.FormatLargeNumber(stats.ValidCombinations),
		"cards_by_role":            stats.CardsByRole,
	}

	// Add elixir ranges
	elixirRanges := make(map[string]string)
	for label, count := range stats.ByElixirRange {
		elixirRanges[label] = deck.FormatLargeNumber(count)
	}
	output["by_elixir_range"] = elixirRanges

	// Add archetypes
	archetypes := make(map[string]string)
	for archetype, count := range stats.ByArchetype {
		archetypes[archetype] = deck.FormatLargeNumber(count)
	}
	output["by_archetype"] = archetypes

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func formatPossibleCountCSV(player *clashroyale.Player, stats *deck.DeckSpaceStats, verbose bool) string {
	var buf strings.Builder

	// Header
	buf.WriteString("Metric,Value\n")

	// Basic info
	buf.WriteString(fmt.Sprintf("Player Name,%s\n", player.Name))
	buf.WriteString(fmt.Sprintf("Player Tag,%s\n", player.Tag))
	buf.WriteString(fmt.Sprintf("Total Cards,%d\n", stats.TotalCards))
	buf.WriteString(fmt.Sprintf("Total Combinations,%s\n", stats.TotalCombinations.String()))
	buf.WriteString(fmt.Sprintf("Valid Combinations,%s\n", stats.ValidCombinations.String()))
	buf.WriteString(fmt.Sprintf("Total Combinations (Formatted),%s\n", deck.FormatLargeNumber(stats.TotalCombinations)))
	buf.WriteString(fmt.Sprintf("Valid Combinations (Formatted),%s\n\n", deck.FormatLargeNumber(stats.ValidCombinations)))

	if verbose {
		// Cards by role
		buf.WriteString("Role,Card Count\n")
		roles := []deck.CardRole{
			deck.RoleWinCondition,
			deck.RoleBuilding,
			deck.RoleSpellBig,
			deck.RoleSpellSmall,
			deck.RoleSupport,
			deck.RoleCycle,
		}
		roleLabels := map[deck.CardRole]string{
			deck.RoleWinCondition: "Win Condition",
			deck.RoleBuilding:     "Building",
			deck.RoleSpellBig:     "Big Spell",
			deck.RoleSpellSmall:   "Small Spell",
			deck.RoleSupport:      "Support",
			deck.RoleCycle:        "Cycle",
		}

		for _, role := range roles {
			count := stats.CardsByRole[role]
			label := roleLabels[role]
			buf.WriteString(fmt.Sprintf("%s,%d\n", label, count))
		}
	}

	return buf.String()
}
