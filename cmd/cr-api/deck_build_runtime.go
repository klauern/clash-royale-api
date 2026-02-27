package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/urfave/cli/v3"
)

func deckBuildCommand(ctx context.Context, cmd *cli.Command) error {
	flags := parseDeckBuildFlags(cmd)

	if err := configureCombatStats(cmd); err != nil {
		return err
	}

	builder, err := configureDeckBuilder(cmd, flags.DataDir, flags.Strategy)
	if err != nil {
		return err
	}

	if err := configureFuzzIntegration(cmd, builder); err != nil {
		return err
	}

	playerData, err := loadPlayerCardAnalysis(ctx, cmd, builder, flags.Tag)
	if err != nil {
		return err
	}
	overrides, err := parseBoostedCardLevels(flags.BoostedCardLevels)
	if err != nil {
		return err
	}
	applyBoostedLevelsToCardAnalysis(&playerData.CardAnalysis, overrides)

	applyExcludeFilter(&playerData.CardAnalysis, flags.ExcludeCards)

	if strings.ToLower(strings.TrimSpace(flags.Strategy)) == deckStrategyAll {
		return buildAllStrategies(ctx, cmd, builder, playerData.CardAnalysis, playerData.PlayerName, playerData.PlayerTag)
	}

	deckRec, err := builder.BuildDeckFromAnalysis(playerData.CardAnalysis)
	if err != nil {
		return fmt.Errorf("failed to build deck: %w", err)
	}

	validateElixirConstraints(deckRec, flags.MinElixir, flags.MaxElixir)
	displayDeckRecommendationOffline(deckRec, playerData.PlayerName, playerData.PlayerTag)

	upgrades := displayUpgradeRecommendationsIfEnabled(
		cmd,
		builder,
		playerData.CardAnalysis,
		deckRec,
		flags.NoSuggestUpgrades,
		flags.UpgradeCount,
	)

	if flags.IdealDeck {
		displayIdealDeck(cmd, builder, playerData.CardAnalysis, deckRec, playerData.PlayerName, playerData.PlayerTag, upgrades)
	}

	if err := saveDeckIfRequested(cmd, builder, deckRec, playerData.PlayerTag, flags.DataDir); err != nil {
		return err
	}

	return nil
}

type deckBuildFlags struct {
	Tag               string
	Strategy          string
	MinElixir         float64
	MaxElixir         float64
	DataDir           string
	ExcludeCards      []string
	BoostedCardLevels []string
	NoSuggestUpgrades bool
	UpgradeCount      int
	IdealDeck         bool
}

func parseDeckBuildFlags(cmd *cli.Command) deckBuildFlags {
	return deckBuildFlags{
		Tag:               cmd.String("tag"),
		Strategy:          cmd.String("strategy"),
		MinElixir:         cmd.Float64("min-elixir"),
		MaxElixir:         cmd.Float64("max-elixir"),
		DataDir:           cmd.String("data-dir"),
		ExcludeCards:      cmd.StringSlice("exclude-cards"),
		BoostedCardLevels: cmd.StringSlice("boosted-card-level"),
		NoSuggestUpgrades: cmd.Bool("no-suggest-upgrades"),
		UpgradeCount:      cmd.Int("upgrade-count"),
		IdealDeck:         cmd.Bool("ideal-deck"),
	}
}

func buildSuiteDeckVariation(
	cmd *cli.Command,
	dataDir string,
	strategy string,
	cardAnalysis deck.CardAnalysis,
) (*deck.DeckRecommendation, error) {
	deckBuilder, err := configureDeckBuilder(cmd, dataDir, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to configure builder: %w", err)
	}
	if err := configureFuzzIntegration(cmd, deckBuilder); err != nil {
		return nil, fmt.Errorf("failed to configure fuzz integration: %w", err)
	}

	deckRec, err := deckBuilder.BuildDeckFromAnalysis(cardAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to build deck: %w", err)
	}
	return deckRec, nil
}

// validateElixirConstraints checks if deck elixir is within requested range
func validateElixirConstraints(deckRec *deck.DeckRecommendation, minElixir, maxElixir float64) {
	if deckRec.AvgElixir < minElixir || deckRec.AvgElixir > maxElixir {
		printf("\n⚠ Warning: Deck average elixir (%.2f) is outside requested range (%.1f-%.1f)\n",
			deckRec.AvgElixir, minElixir, maxElixir)
	}
}

// displayUpgradeRecommendationsIfEnabled displays upgrade recommendations if not disabled
func displayUpgradeRecommendationsIfEnabled(
	cmd *cli.Command,
	builder *deck.Builder,
	cardAnalysis deck.CardAnalysis,
	deckRec *deck.DeckRecommendation,
	noSuggestUpgrades bool,
	upgradeCount int,
) *deck.UpgradeRecommendations {
	if noSuggestUpgrades {
		return nil
	}

	printf("\n")
	upgrades, err := builder.GetUpgradeRecommendations(cardAnalysis, deckRec, upgradeCount)
	if err != nil {
		if cmd.Bool("verbose") {
			printf("Warning: Failed to generate upgrade recommendations: %v\n", err)
		}
		return nil
	}

	displayUpgradeRecommendations(upgrades)
	return upgrades
}

//nolint:gocognit,gocyclo,funlen // Legacy suite command path; phased extraction follows in clash-royale-api-sb3q.
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
	excludeCards := cmd.StringSlice("exclude-cards")
	boostedCardLevels := cmd.StringSlice("boosted-card-level")

	// Determine output directory
	if outputDir == "" {
		outputDir = filepath.Join(dataDir, "decks")
	}

	// Configure combat stats behavior for suite builds
	if err := configureCombatStats(cmd); err != nil {
		return err
	}

	// Parse strategies
	strategies, err := parseStrategies(strategiesStr)
	if err != nil {
		return err
	}

	if verbose {
		printf("Building deck suite with %d strategies x %d variations = %d total decks\n",
			len(strategies), variations, len(strategies)*variations)
	}

	// Create configured builder and load player data
	builder, err := configureDeckBuilder(cmd, dataDir, "")
	if err != nil {
		return err
	}

	// Load player data
	playerData, err := loadSuitePlayerData(ctx, builder, tag, apiToken, dataDir, fromAnalysis, verbose)
	if err != nil {
		return err
	}
	overrides, err := parseBoostedCardLevels(boostedCardLevels)
	if err != nil {
		return err
	}
	applyBoostedLevelsToCardAnalysis(&playerData.CardAnalysis, overrides)

	// Apply exclude filter
	applyExcludeFilter(&playerData.CardAnalysis, excludeCards)

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

	printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	printf("║                    DECK BUILD SUITE                                ║\n")
	printf("╚════════════════════════════════════════════════════════════════════╝\n\n")
	printf("Player: %s (%s)\n", playerData.PlayerName, playerData.PlayerTag)
	printf("Output: %s\n\n", outputDir)

	// Build decks for each strategy
	for _, strategy := range strategies {
		if verbose {
			printf("Building decks for strategy: %s\n", strategy)
		}

		for v := 1; v <= variations; v++ {
			deckRec, err := buildSuiteDeckVariation(cmd, dataDir, string(strategy), playerData.CardAnalysis)
			if err != nil {
				results = append(results, deckResult{
					Strategy:   string(strategy),
					Variation:  v,
					BuildError: err,
				})
				printf("  ⚠ Variation %d: Failed to build deck: %v\n", v, err)
				continue
			}

			// Validate elixir constraints
			if deckRec.AvgElixir < minElixir || deckRec.AvgElixir > maxElixir {
				if verbose {
					printf("  ⚠ Variation %d: Deck average elixir (%.2f) outside range (%.1f-%.1f)\n",
						v, deckRec.AvgElixir, minElixir, maxElixir)
				}
			}

			// Save deck file if requested
			var filePath string
			if saveData {
				timestamp := time.Now().Format("20060102_150405")
				filePath = filepath.Join(outputDir, deck.SuiteDeckFilename(timestamp, string(strategy), v, playerData.PlayerTag))

				// Save with a plain builder instance; build-specific config is not required for serialization.
				saveBuilder := deck.NewBuilder(dataDir)
				savedPath, err := saveBuilder.SaveDeck(deckRec, outputDir, fmt.Sprintf("%s_var%d_%s", strategy, v, playerData.PlayerTag))
				if err != nil {
					if verbose {
						printf("  ⚠ Variation %d: Failed to save deck: %v\n", v, err)
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

			printf("  ✓ %s variation %d: %.2f avg elixir, %d cards\n",
				strategy, v, deckRec.AvgElixir, len(deckRec.Deck))
		}
	}

	totalTime := time.Since(startTime)

	// Display summary
	printf("\n")
	printf("════════════════════════════════════════════════════════════════════\n")
	printf("                           SUMMARY\n")
	printf("════════════════════════════════════════════════════════════════════\n\n")

	successful := 0
	failed := 0
	for _, r := range results {
		if r.BuildError == nil {
			successful++
		} else {
			failed++
		}
	}

	printf("Total decks:     %d\n", len(results))
	printf("Successful:      %d\n", successful)
	printf("Failed:          %d\n", failed)
	printf("Build time:      %v\n", totalTime)
	if len(results) > 0 {
		printf("Avg per deck:    %v\n\n", totalTime/time.Duration(len(results)))
	} else {
		printf("Avg per deck:    n/a (no generated decks)\n\n")
	}

	// Save summary JSON if requested
	if saveData && successful > 0 {
		timestamp := time.Now().Format("20060102_150405")
		summaryPath := filepath.Join(outputDir, deck.SuiteSummaryFilename(timestamp, playerData.PlayerTag))
		summaries := make([]deck.SuiteDeckSummary, 0, len(results))
		for _, result := range results {
			if result.Deck == nil {
				continue
			}
			summaries = append(summaries, deck.SuiteDeckSummary{
				Strategy:  result.Strategy,
				Variation: result.Variation,
				Cards:     result.Deck.Deck,
				AvgElixir: result.Deck.AvgElixir,
				FilePath:  result.FilePath,
			})
		}
		summary := deck.NewSuiteSummary(
			time.Now().UTC().Format(time.RFC3339),
			playerData.PlayerName,
			playerData.PlayerTag,
			deck.SuiteBuildInfo{
				TotalDecks:     len(results),
				Successful:     successful,
				Failed:         failed,
				Strategies:     len(strategies),
				Variations:     variations,
				GenerationTime: totalTime.String(),
			},
			summaries,
		)
		if err := deck.WriteSuiteSummary(summaryPath, summary); err != nil {
			printf("Warning: Failed to write summary file: %v\n", err)
		} else {
			printf("Summary saved to: %s\n", summaryPath)
		}
	}

	return nil
}

// deckEvaluateBatchCommand evaluates multiple decks from a suite or directory
// evalBatchFlags holds parsed CLI flags for batch evaluation
type evalBatchFlags struct {
	FromSuite       string
	DeckDir         string
	PlayerTag       string
	Format          string
	OutputDir       string
	SortBy          string
	TopOnly         bool
	TopN            int
	FilterArchetype bool
	ArchetypeFilter string
	Verbose         bool
	ShowTiming      bool
	SaveAggregated  bool
}

// evalDeckInfo holds information about a deck to evaluate
type evalDeckInfo struct {
	Name     string
	Cards    []string
	Strategy string
	FilePath string
}

// evalBatchResult holds the result of evaluating a single deck
type evalBatchResult struct {
	Name      string
	Strategy  string
	Deck      []string
	Result    evaluation.EvaluationResult
	FilePath  string
	Evaluated time.Time
	Duration  time.Duration
}

// parseEvalBatchFlags extracts and validates CLI flags for batch evaluation
