package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/klauer/clash-royale-api/go/pkg/events"
	"github.com/klauer/clash-royale-api/go/pkg/leaderboard"
	"github.com/urfave/cli/v3"
)

func runPhase0CardConstraints(tag, dataDir string, suggestConstraints bool, constraintThreshold float64, topN int, verbose bool) error {
	if !suggestConstraints {
		return nil
	}
	fmt.Println("💡 PHASE 0: Analyzing top event decks for card constraint suggestions...")
	fmt.Println("─────────────────────────────────────────────────────────────────────")

	// Initialize event manager
	eventManager := events.NewManager(dataDir)

	// Load player's event deck collection
	collection, err := eventManager.GetCollection(tag)
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not load event decks: %v\n", err)
		fmt.Println("Continuing without constraint suggestions...")
		fmt.Println()
		return nil
	}
	if len(collection.Decks) == 0 {
		fmt.Println("⚠️  No event decks found for this player.")
		fmt.Println("Play some challenges or tournaments to build event deck history.")
		fmt.Println()
		return nil
	}

	// Get top N decks by win rate
	minBattles := 3 // Minimum battles to qualify
	topDecks := collection.GetBestDecksByWinRate(minBattles, topN)

	if len(topDecks) == 0 {
		fmt.Printf("⚠️  No event decks found with at least %d battles.\n", minBattles)
		fmt.Println()
		return nil
	}

	// Analyze and suggest constraints
	suggestions := events.SuggestCardConstraints(topDecks, constraintThreshold)

	if len(suggestions) == 0 {
		fmt.Printf("No cards meet the %.0f%% threshold in top %d decks.\n", constraintThreshold, len(topDecks))
		fmt.Println()
		return nil
	}

	fmt.Printf("\n=== Card Constraint Suggestions (from top %d decks) ===\n", len(topDecks))
	for _, suggestion := range suggestions {
		fmt.Printf("%d/%d decks (%.0f%%) contain %s\n",
			suggestion.Appearances,
			suggestion.TotalDecks,
			suggestion.Percentage,
			suggestion.CardName)
	}
	fmt.Println()

	// Generate example command with suggested constraints
	fmt.Println("To apply these constraints, re-run with:")
	cmdExample := fmt.Sprintf("  cr-api deck analyze-suite --tag %s", tag)
	for _, suggestion := range suggestions {
		cmdExample += fmt.Sprintf(" --include-cards \"%s\"", suggestion.CardName)
	}
	fmt.Println(cmdExample)
	fmt.Println()
	return nil
}

// runPhase1BuildDeckVariations builds deck variations for the analysis suite
//
//nolint:unused,funlen,gocognit,gocyclo // Large orchestration retained pending modularization task clash-royale-api-1g1r.
func runPhase1BuildDeckVariations(tag, strategiesStr, outputDir string, variations, topN int, includeCards, excludeCards []string, verbose bool, apiToken, dataDir string, fromAnalysis bool, minElixir, maxElixir float64, timestamp string) ([]suiteDeckInfo, *suitePlayerData, int, int, string, error) {
	decksDir := filepath.Join(outputDir, "decks")
	if err := os.MkdirAll(decksDir, 0o755); err != nil {
		return nil, nil, 0, 0, "", fmt.Errorf("failed to create decks directory: %w", err)
	}

	// Parse strategies
	strategies, err := parseStrategies(strategiesStr)
	if err != nil {
		return nil, nil, 0, 0, "", err
	}

	if verbose {
		printf("Strategies: %v\n", strategies)
		printf("Variations per strategy: %d\n", variations)
		printf("Total decks to build: %d\n", len(strategies)*variations)
	}

	// Build decks using build-suite logic
	builder := deck.NewBuilder(dataDir)

	// Load player data
	playerData, err := loadSuitePlayerData(builder, tag, apiToken, dataDir, fromAnalysis, verbose)
	if err != nil {
		return nil, nil, 0, 0, "", err
	}

	var builtDecks []suiteDeckInfo
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
				printf("  ✗ Failed to set strategy %s variation %d: %v\n", strategy, v, err)
				failCount++
				continue
			}

			// Build deck
			deckRec, err := deckBuilder.BuildDeckFromAnalysis(playerData.CardAnalysis)
			if err != nil {
				printf("  ✗ Failed to build %s variation %d: %v\n", strategy, v, err)
				failCount++
				continue
			}

			// Save deck to file
			deckFileName := fmt.Sprintf("%s_deck_%s_var%d_%s.json", timestamp, strategy, v, strings.TrimPrefix(playerData.PlayerTag, "#"))
			deckFilePath := filepath.Join(decksDir, deckFileName)

			deckData := map[string]interface{}{
				"deck":           deckRec.Deck,
				"avg_elixir":     deckRec.AvgElixir,
				"recommendation": deckRec,
			}

			deckJSON, err := json.MarshalIndent(deckData, "", "  ")
			if err != nil {
				printf("  ✗ Failed to marshal deck %s variation %d: %v\n", strategy, v, err)
				failCount++
				continue
			}

			if err := os.WriteFile(deckFilePath, deckJSON, 0o644); err != nil {
				printf("  ✗ Failed to save deck %s variation %d: %v\n", strategy, v, err)
				failCount++
				continue
			}

			builtDecks = append(builtDecks, suiteDeckInfo{
				Strategy:  string(strategy),
				Variation: v,
				Cards:     deckRec.Deck,
				AvgElixir: deckRec.AvgElixir,
				FilePath:  deckFilePath,
			})

			successCount++
			if verbose {
				printf("  ✓ Built %s variation %d (%.2f avg elixir)\n", strategy, v, deckRec.AvgElixir)
			}
		}
	}

	buildDuration := time.Since(buildStart)

	// Save suite summary
	suiteFileName := fmt.Sprintf("%s_deck_suite_summary_%s.json", timestamp, strings.TrimPrefix(playerData.PlayerTag, "#"))
	suiteSummaryPath := filepath.Join(decksDir, suiteFileName)

	suiteData := map[string]interface{}{
		"version":   "1.0.0",
		"timestamp": timestamp,
		"player": map[string]string{
			"name": playerData.PlayerName,
			"tag":  playerData.PlayerTag,
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
		return nil, nil, 0, 0, "", fmt.Errorf("failed to marshal suite summary: %w", err)
	}

	if err := os.WriteFile(suiteSummaryPath, suiteJSON, 0o644); err != nil {
		return nil, nil, 0, 0, "", fmt.Errorf("failed to save suite summary: %w", err)
	}

	fmt.Println()
	printf("✓ Built %d/%d decks successfully in %s\n", successCount, successCount+failCount, buildDuration.Round(time.Millisecond))
	printf("  Suite summary: %s\n", suiteSummaryPath)
	fmt.Println()

	if successCount == 0 {
		return nil, nil, 0, 0, "", fmt.Errorf("no decks were built successfully")
	}

	return builtDecks, playerData, successCount, failCount, suiteSummaryPath, nil
}

// runPhase2EvaluateAllDecks evaluates all built decks for the analysis suite
//
//nolint:unused,funlen,gocognit,gocyclo // Large orchestration retained pending modularization task clash-royale-api-1g1r.
func runPhase2EvaluateAllDecks(builtDecks []suiteDeckInfo, playerData *suitePlayerData, outputDir, tag, apiToken string, fromAnalysis, verbose bool, timestamp string) ([]suiteEvalResult, string, error) {
	evaluationsDir := filepath.Join(outputDir, "evaluations")
	if err := os.MkdirAll(evaluationsDir, 0o755); err != nil {
		return nil, "", fmt.Errorf("failed to create evaluations directory: %w", err)
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
				fprintf(os.Stderr, "Warning: failed to initialize storage: %v\n", storageErr)
			}
		} else {
			defer func() {
				if err := storage.Close(); err != nil {
					fprintf(os.Stderr, "Warning: failed to close storage: %v\n", err)
				}
			}()
			if verbose {
				printf("Initialized persistent storage at: %s\n", storage.GetDBPath())
			}
		}
	}

	var results []suiteEvalResult
	evalStart := time.Now()

	for _, deckInf := range builtDecks {
		deckStart := time.Now()

		var deckCards []deck.CardCandidate
		if deckInf.FilePath != "" {
			candidates, ok, err := loadDeckCandidatesFromFile(deckInf.FilePath)
			if err != nil && verbose {
				printf("Warning: Failed to load deck details from %s: %v\n", deckInf.FilePath, err)
			}
			if ok {
				deckCards = candidates
			}
		}
		if len(deckCards) == 0 {
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
		}

		// Evaluate
		deckEvalResult := evaluation.Evaluate(deckCards, synergyDB, playerContext)
		evalDuration := time.Since(deckStart)

		deckName := fmt.Sprintf("Deck #%d (%s v%d)", len(results)+1, deckInf.Strategy, deckInf.Variation)

		results = append(results, suiteEvalResult{
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
				fprintf(os.Stderr, "  Warning: failed to save deck to storage: %v\n", err)
			}
		}

		if verbose {
			printf("  ✓ Evaluated %s: %.2f overall (%.0fms)\n", deckName, deckEvalResult.OverallScore, float64(evalDuration.Nanoseconds())/1e6)
		}
	}

	evalDuration := time.Since(evalStart)

	// Recalculate and update statistics after all insertions
	if storage != nil {
		_, err := storage.RecalculateStats()
		if err != nil && verbose {
			fprintf(os.Stderr, "Warning: failed to recalculate storage stats: %v\n", err)
		}
	}

	// Sort by overall score
	sortEvaluationResults(results, "overall")

	// Save evaluation results
	evalFileName := fmt.Sprintf("%s_deck_evaluations_%s.json", timestamp, strings.TrimPrefix(playerData.PlayerTag, "#"))
	evalFilePath := filepath.Join(evaluationsDir, evalFileName)

	evalData := map[string]interface{}{
		"version":   "1.0.0",
		"timestamp": timestamp,
		"player": map[string]string{
			"name": playerData.PlayerName,
			"tag":  playerData.PlayerTag,
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
		return nil, "", fmt.Errorf("failed to marshal evaluation results: %w", err)
	}

	if err := os.WriteFile(evalFilePath, evalJSON, 0o644); err != nil {
		return nil, "", fmt.Errorf("failed to save evaluation results: %w", err)
	}

	fmt.Println()
	printf("✓ Evaluated %d decks in %s\n", len(results), evalDuration.Round(time.Millisecond))
	printf("  Evaluation results: %s\n", evalFilePath)
	fmt.Println()

	return results, evalFilePath, nil
}

// runPhase3CompareTopPerformers compares top decks and generates final report
//
//nolint:unused,funlen // Reserved for phased suite refactor tracked in beads.
func runPhase3CompareTopPerformers(results []suiteEvalResult, topN int, outputDir, timestamp string, playerData *suitePlayerData, successCount int, suiteSummaryPath, evalFilePath string) (string, error) {
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
		return "", fmt.Errorf("failed to create reports directory: %w", err)
	}

	reportFileName := fmt.Sprintf("%s_deck_analysis_report_%s.md", timestamp, strings.TrimPrefix(playerData.PlayerTag, "#"))
	reportFilePath := filepath.Join(reportsDir, reportFileName)

	// Generate comprehensive markdown report
	reportContent := generateComparisonReport(deckNames, evalResults)

	if err := os.WriteFile(reportFilePath, []byte(reportContent), 0o644); err != nil {
		return "", fmt.Errorf("failed to save comparison report: %w", err)
	}

	printf("✓ Generated comparison report for top %d decks\n", compareCount)
	printf("  Report: %s\n", reportFilePath)
	fmt.Println()

	// ========================================================================
	// SUMMARY
	// ========================================================================
	fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                      ANALYSIS SUITE COMPLETE                       ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	printf("Player: %s (%s)\n", playerData.PlayerName, playerData.PlayerTag)
	printf("Decks built: %d\n", successCount)
	printf("Decks evaluated: %d\n", len(results))
	printf("Top performers compared: %d\n", compareCount)
	fmt.Println()
	fmt.Println("📂 Output files:")
	printf("  • Suite summary:  %s\n", suiteSummaryPath)
	printf("  • Evaluations:    %s\n", evalFilePath)
	printf("  • Final report:   %s\n", reportFilePath)
	fmt.Println()

	if len(results) > 0 {
		fmt.Println("🥇 Top 3 decks:")
		for i := 0; i < 3 && i < len(results); i++ {
			medal := []string{"🥇", "🥈", "🥉"}[i]
			res := results[i]
			printf("  %s %s: %.2f (%.2f avg elixir, %s archetype)\n",
				medal, res.Name, res.Result.OverallScore, res.Result.AvgElixir,
				res.Result.DetectedArchetype)
		}
	}

	return reportFilePath, nil
}

// deckAnalyzeSuiteCommand orchestrates the full deck analysis workflow:
// (1) Build multiple deck variations using build-suite logic
// (2) Evaluate all built decks using evaluate-batch logic
// (3) Compare top performers using compare logic
// (4) Generate comprehensive markdown report
//
//nolint:funlen,gocognit,gocyclo,gocritic,dupl // Legacy orchestration path pending extraction in clash-royale-api-1g1r.
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
	suggestConstraints := cmd.Bool("suggest-constraints")
	constraintThreshold := cmd.Float64("constraint-threshold")

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
	// PHASE 0 (Optional): Card Constraint Suggestions
	// ========================================================================
	if suggestConstraints {
		if err := runPhase0CardConstraints(tag, dataDir, suggestConstraints, constraintThreshold, topN, verbose); err != nil {
			return err
		}
	}

	// ========================================================================
	// PHASE 1: Build deck variations
	// ========================================================================
	fmt.Println("📦 PHASE 1: Building deck variations...")
	fmt.Println("─────────────────────────────────────────────────────────────────────")

	builtDecks, playerData, successCount, _, suiteSummaryPath, err := runPhase1BuildDeckVariations(
		tag, strategiesStr, outputDir, variations, topN, includeCards, excludeCards, verbose,
		apiToken, dataDir, fromAnalysis, minElixir, maxElixir, timestamp,
	)
	if err != nil {
		return err
	}

	// ========================================================================
	// PHASE 2: Evaluate all decks
	// ========================================================================
	fmt.Println("📊 PHASE 2: Evaluating all decks...")
	fmt.Println("─────────────────────────────────────────────────────────────────────")

	results, evalFilePath, err := runPhase2EvaluateAllDecks(
		builtDecks, playerData, outputDir, tag, apiToken, fromAnalysis, verbose, timestamp,
	)
	if err != nil {
		return err
	}

	// ========================================================================
	// PHASE 3: Compare top performers
	// ========================================================================
	fmt.Println("🏆 PHASE 3: Comparing top performers...")
	fmt.Println("─────────────────────────────────────────────────────────────────────")

	reportFilePath, err := runPhase3CompareTopPerformers(
		results, topN, outputDir, timestamp, playerData, successCount, suiteSummaryPath, evalFilePath,
	)
	if err != nil {
		return err
	}

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

	// ========================================================================
	// SUMMARY
	// ========================================================================
	fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                      ANALYSIS SUITE COMPLETE                       ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	printf("Player: %s (%s)\n", playerData.PlayerName, playerData.PlayerTag)
	printf("Decks built: %d\n", successCount)
	printf("Decks evaluated: %d\n", len(results))
	printf("Top performers compared: %d\n", compareCount)
	fmt.Println()
	fmt.Println("📂 Output files:")
	printf("  • Suite summary:  %s\n", suiteSummaryPath)
	printf("  • Evaluations:    %s\n", evalFilePath)
	printf("  • Final report:   %s\n", reportFilePath)
	fmt.Println()

	if len(results) > 0 {
		fmt.Println("🥇 Top 3 decks:")
		for i := 0; i < 3 && i < len(results); i++ {
			medal := []string{"🥇", "🥈", "🥉"}[i]
			res := results[i]
			printf("  %s %s: %.2f (%.2f avg elixir, %s archetype)\n",
				medal, res.Name, res.Result.OverallScore, res.Result.AvgElixir,
				res.Result.DetectedArchetype)
		}
	}

	return nil
}
