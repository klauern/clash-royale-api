package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/comparison"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/urfave/cli/v3"
)

func deckCompareAlgorithmsCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	strategiesStr := cmd.String("strategies")
	outputFile := cmd.String("output")
	format := cmd.String("format")
	significance := cmd.Float64("significance")
	winThreshold := cmd.Float64("win-threshold")
	dataDir := cmd.String("data-dir")

	// Step 1: Configure combat stats
	if err := configureCombatStats(cmd); err != nil {
		return err
	}

	// Step 2: Parse strategies
	strategies, err := parseStrategies(strategiesStr)
	if err != nil {
		return err
	}

	printComparisonHeader()

	// Step 3: Load player card analysis
	builder := deck.NewBuilder(dataDir)
	playerData, err := loadPlayerCardAnalysis(cmd, builder, tag)
	if err != nil {
		return fmt.Errorf("failed to analyze player cards: %w", err)
	}

	printComparisonPlayerInfo(tag, playerData, strategies)

	// Step 4-5: Run comparison
	result, err := runAlgorithmComparison(tag, playerData.CardAnalysis, strategies, significance, winThreshold)
	if err != nil {
		return err
	}

	// Step 6-7: Format and output
	return formatAndOutputComparison(result, format, outputFile)
}

// parseStrategies parses a comma-separated list of strategy strings
func parseStrategies(strategiesStr string) ([]deck.Strategy, error) {
	strategiesStr = strings.ToLower(strings.TrimSpace(strategiesStr))

	// Handle "all" keyword
	if strategiesStr == deckStrategyAll {
		return []deck.Strategy{
			deck.StrategyBalanced,
			deck.StrategyAggro,
			deck.StrategyControl,
			deck.StrategyCycle,
			deck.StrategySplash,
			deck.StrategySpell,
		}, nil
	}

	var strategies []deck.Strategy
	strategyList := strings.Split(strategiesStr, ",")

	for _, s := range strategyList {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		parsedStrategy, err := deck.ParseStrategy(s)
		if err != nil {
			return nil, fmt.Errorf("invalid strategy '%s': %w", s, err)
		}
		strategies = append(strategies, parsedStrategy)
	}

	if len(strategies) == 0 {
		return nil, fmt.Errorf("no valid strategies specified")
	}

	return strategies, nil
}

// suitePlayerData holds player data loaded for suite operations
type suitePlayerData struct {
	CardAnalysis deck.CardAnalysis
	PlayerName   string
	PlayerTag    string
}

//nolint:unused // Reserved for phased suite refactor tracked in beads.
type suiteDeckInfo struct {
	Strategy  string   `json:"strategy"`
	Variation int      `json:"variation"`
	Cards     []string `json:"cards"`
	AvgElixir float64  `json:"avg_elixir"`
	FilePath  string   `json:"file_path"`
}

//nolint:unused // Reserved for phased suite refactor tracked in beads.
type suiteEvalResult struct {
	Name      string                      `json:"name"`
	Strategy  string                      `json:"strategy"`
	Deck      []string                    `json:"deck"`
	Result    evaluation.EvaluationResult `json:"Result"`
	FilePath  string                      `json:"FilePath"`
	Evaluated string                      `json:"Evaluated"`
	Duration  int64                       `json:"Duration"`
}

// loadSuitePlayerData loads player data for suite commands (online or offline)
func loadSuitePlayerDataFromAnalysis(builder *deck.Builder, tag, dataDir string, verbose bool) (*suitePlayerData, error) {
	if verbose {
		printf("Building deck suite from offline analysis for player %s\n", tag)
	}
	analysisDir := filepath.Join(dataDir, "analysis")
	loadedAnalysis, err := builder.LoadLatestAnalysis(tag, analysisDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load analysis for player %s from %s: %w", tag, analysisDir, err)
	}
	return &suitePlayerData{
		CardAnalysis: *loadedAnalysis,
		PlayerName:   tag,
		PlayerTag:    tag,
	}, nil
}

//nolint:dupl // Shared API loading refactor tracked under clash-royale-api-sg50.
func loadSuitePlayerDataFromAPI(builder *deck.Builder, tag, apiToken string, verbose bool) (*suitePlayerData, error) {
	if apiToken == "" {
		return nil, fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag. Use --from-analysis for offline mode")
	}
	client := clashroyale.NewClient(apiToken)
	if verbose {
		printf("Building deck suite for player %s\n", tag)
	}
	player, err := client.GetPlayer(tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	if verbose {
		printf("Player: %s (%s)\n", player.Name, player.Tag)
		printf("Analyzing %d cards...\n", len(player.Cards))
	}
	analysisOptions := analysis.DefaultAnalysisOptions()
	cardAnalysis, err := analysis.AnalyzeCardCollection(player, analysisOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze card collection: %w", err)
	}
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
			MaxEvolutionLevel: cardInfo.MaxEvolutionLevel,
		}
	}
	return &suitePlayerData{
		CardAnalysis: deckCardAnalysis,
		PlayerName:   player.Name,
		PlayerTag:    player.Tag,
	}, nil
}

func loadSuitePlayerData(builder *deck.Builder, tag, apiToken, dataDir string, fromAnalysis, verbose bool) (*suitePlayerData, error) {
	if fromAnalysis {
		return loadSuitePlayerDataFromAnalysis(builder, tag, dataDir, verbose)
	}
	return loadSuitePlayerDataFromAPI(builder, tag, apiToken, verbose)
}

// printComparisonHeader prints the algorithm comparison header
func printComparisonHeader() {
	printf("╔════════════════════════════════════════════════════════════════════╗\n")
	printf("║              ALGORITHM COMPARISON: V1 vs V2                         ║\n")
	printf("╚════════════════════════════════════════════════════════════════════╝\n\n")
}

// printComparisonPlayerInfo prints player and strategy information
func printComparisonPlayerInfo(tag string, playerData *playerDataLoadResult, strategies []deck.Strategy) {
	printf("Player: %s (%s)\n", playerData.PlayerName, tag)
	strategyNames := make([]string, len(strategies))
	for i, s := range strategies {
		strategyNames[i] = s.String()
	}
	printf("Strategies: %s\n\n", strings.Join(strategyNames, ", "))
}

// runAlgorithmComparison runs the algorithm comparison with the given configuration
func runAlgorithmComparison(
	tag string,
	cardAnalysis deck.CardAnalysis,
	strategies []deck.Strategy,
	significance, winThreshold float64,
) (*comparison.AlgorithmComparisonResult, error) {
	config := comparison.DefaultComparisonConfig()
	config.PlayerTag = tag
	config.Strategies = strategies
	config.SignificanceThreshold = significance
	config.WinThreshold = winThreshold

	result, err := comparison.CompareAlgorithms(tag, cardAnalysis, config)
	if err != nil {
		return nil, fmt.Errorf("comparison failed: %w", err)
	}

	return result, nil
}

// formatAndOutputComparison formats the comparison result and outputs it
func formatAndOutputComparison(result *comparison.AlgorithmComparisonResult, format, outputFile string) error {
	output, err := formatComparisonOutput(result, format)
	if err != nil {
		return err
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		printf("Comparison report saved to: %s\n", outputFile)
	} else {
		fmt.Print(output)
	}

	return nil
}

// formatComparisonOutput formats the comparison result based on the specified format
func formatComparisonOutput(result *comparison.AlgorithmComparisonResult, format string) (string, error) {
	switch strings.ToLower(format) {
	case compareFormatJSON:
		output, err := result.ExportJSON()
		if err != nil {
			return "", fmt.Errorf("failed to export JSON: %w", err)
		}
		return output, nil
	case compareFormatMarkdown, compareFormatMD:
		return result.ExportMarkdown(), nil
	default:
		return "", fmt.Errorf("unknown format: %s (supported: json, markdown)", format)
	}
}
