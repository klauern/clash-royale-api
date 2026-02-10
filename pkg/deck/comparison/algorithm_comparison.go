// Package comparison provides framework for comparing deck building algorithms
// (V1 vs V2) on quality metrics, archetype purity, meta viability, and user satisfaction.
package comparison

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

// Constants for algorithm comparison results
const (
	WinnerV1  = "v1"
	WinnerV2  = "v2"
	WinnerTie = "tie"

	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
	ConfidenceLow    = "low"
)

// ComparisonDimensions represents the axes for algorithm comparison
type ComparisonDimensions struct {
	DeckQuality      bool // Which produces better decks?
	ArchetypePurity  bool // Which follows strategies better?
	MetaViability    bool // Which decks would work on ladder?
	UserSatisfaction bool // Which addresses user complaints?
}

// AlgorithmComparisonResult contains comprehensive comparison results between V1 and V2 algorithms
type AlgorithmComparisonResult struct {
	// Metadata
	PlayerTag  string   `json:"player_tag"`
	Timestamp  string   `json:"timestamp"`
	Strategies []string `json:"strategies_tested"`
	TotalDecks int      `json:"total_decks_compared"`

	// Overall summary
	Summary ComparisonSummary `json:"summary"`

	// Per-strategy results
	StrategyResults map[string]StrategyComparison `json:"strategy_results"`

	// Detailed metrics
	MetricBreakdown MetricBreakdown `json:"metric_breakdown"`

	// Recommendations
	Recommendations Recommendations `json:"recommendations"`
}

// ComparisonSummary provides high-level comparison summary
type ComparisonSummary struct {
	Winner            string  `json:"winning_algorithm"` // "v1", "v2", or "tie"
	V1AverageScore    float64 `json:"v1_average_score"`
	V2AverageScore    float64 `json:"v2_average_score"`
	Improvement       float64 `json:"improvement_percent"`   // ((v2 - v1) / v1) * 100
	SignificantWins   int     `json:"v2_significant_wins"`   // v2 wins by >10%
	SignificantLosses int     `json:"v2_significant_losses"` // v2 loses by >10%
}

// StrategyComparison compares V1 vs V2 for a specific strategy
type StrategyComparison struct {
	StrategyName string `json:"strategy"`

	// Deck scores
	V1AverageDeckScore float64 `json:"v1_average_deck_score"`
	V2AverageDeckScore float64 `json:"v2_average_deck_score"`

	// Metric breakdowns
	Metrics StrategyMetrics `json:"metrics"`

	// Per-deck comparisons
	DeckComparisons []DeckComparison `json:"deck_comparisons"`

	// Winner for this strategy
	Winner string `json:"winner"` // "v1", "v2", or "tie"
}

// StrategyMetrics contains detailed metric comparison for a strategy
type StrategyMetrics struct {
	SynergyScore          MetricComparison `json:"synergy_score"`
	CounterCoverage       MetricComparison `json:"counter_coverage"`
	ArchetypeCoherence    MetricComparison `json:"archetype_coherence"`
	DefensiveCapability   MetricComparison `json:"defensive_capability"`
	ElixirRangeAdherence  MetricComparison `json:"elixir_range_adherence"`
	CardLevelDistribution MetricComparison `json:"card_level_distribution"`
}

// MetricComparison compares a single metric between V1 and V2
type MetricComparison struct {
	V1Average   float64 `json:"v1_average"`
	V2Average   float64 `json:"v2_average"`
	Improvement float64 `json:"improvement_percent"`
	Winner      string  `json:"winner"`
	Significant bool    `json:"significant"` // Difference >5%
}

// DeckComparison compares a single deck built by both algorithms
type DeckComparison struct {
	DeckName string `json:"deck_name"`

	// V1 results
	V1Deck  []string `json:"v1_deck"`
	V1Score float64  `json:"v1_score"`

	// V2 results
	V2Deck  []string `json:"v2_deck"`
	V2Score float64  `json:"v2_score"`

	// V2 evaluation metrics (from quality test suite)
	V2Evaluation EvaluationMetrics `json:"v2_evaluation"`

	// Comparison
	ScoreDifference float64 `json:"score_difference"`
	Winner          string  `json:"winner"` // "v1", "v2", or "tie"
}

// EvaluationMetrics contains the metrics from the evaluation package
type EvaluationMetrics struct {
	OverallScore        float64              `json:"overall_score"`
	AttackScore         float64              `json:"attack_score"`
	DefenseScore        float64              `json:"defense_score"`
	SynergyScore        float64              `json:"synergy_score"`
	VersatilityScore    float64              `json:"versatility_score"`
	Archetype           evaluation.Archetype `json:"archetype"`
	ArchetypeConfidence float64              `json:"archetype_confidence"`
	SynergyPairs        int                  `json:"synergy_pairs"`
}

// MetricBreakdown provides aggregate metrics across all strategies
type MetricBreakdown struct {
	SynergyScore          AggregateMetric `json:"synergy_score"`
	CounterCoverage       AggregateMetric `json:"counter_coverage"`
	ArchetypeCoherence    AggregateMetric `json:"archetype_coherence"`
	DefensiveCapability   AggregateMetric `json:"defensive_capability"`
	ElixirRangeAdherence  AggregateMetric `json:"elixir_range_adherence"`
	CardLevelDistribution AggregateMetric `json:"card_level_distribution"`
}

// AggregateMetric provides statistical summary for a metric
type AggregateMetric struct {
	V1Mean      float64 `json:"v1_mean"`
	V2Mean      float64 `json:"v2_mean"`
	V1StdDev    float64 `json:"v1_stddev"`
	V2StdDev    float64 `json:"v2_stddev"`
	Improvement float64 `json:"improvement_percent"`
	PValue      float64 `json:"p_value"` // Simplified statistical significance
	Winner      string  `json:"winner"`
}

// Recommendations provides actionable recommendations based on comparison
type Recommendations struct {
	AlgorithmCutover string   `json:"recommended_algorithm"`
	Confidence       string   `json:"confidence"` // "high", "medium", "low"
	Reasoning        []string `json:"reasoning"`
	NextSteps        []string `json:"next_steps"`
	KnownIssues      []string `json:"known_issues"`
}

// ComparisonConfig configures the algorithm comparison
type ComparisonConfig struct {
	PlayerTag             string          `json:"player_tag"`
	Strategies            []deck.Strategy `json:"strategies"`
	SignificanceThreshold float64         `json:"significance_threshold"` // Default: 0.05 (5%)
	WinThreshold          float64         `json:"win_threshold"`          // Default: 0.10 (10%)
}

// DefaultComparisonConfig returns default comparison configuration
func DefaultComparisonConfig() ComparisonConfig {
	return ComparisonConfig{
		Strategies: []deck.Strategy{
			deck.StrategyBalanced,
			deck.StrategyCycle,
			deck.StrategyControl,
			deck.StrategyAggro,
			deck.StrategySplash,
		},
		SignificanceThreshold: 0.05,
		WinThreshold:          0.10,
	}
}

// CompareAlgorithms runs comprehensive comparison between V1 and V2 algorithms
func CompareAlgorithms(playerTag string, cardAnalysis deck.CardAnalysis, config ComparisonConfig) (*AlgorithmComparisonResult, error) {
	if config.PlayerTag == "" {
		config.PlayerTag = playerTag
	}

	result := &AlgorithmComparisonResult{
		PlayerTag:       config.PlayerTag,
		Timestamp:       time.Now().Format(time.RFC3339),
		Strategies:      strategyNames(config.Strategies),
		StrategyResults: make(map[string]StrategyComparison),
	}

	// Build synergy database for V2
	synergyDB := deck.NewSynergyDatabase()

	// Compare each strategy
	totalV1Score := 0.0
	totalV2Score := 0.0
	significantWins := 0
	significantLosses := 0
	totalDecks := 0

	for _, strategy := range config.Strategies {
		strategyResult, err := compareStrategy(strategy, cardAnalysis, synergyDB, config)
		if err != nil {
			return nil, fmt.Errorf("failed to compare strategy %s: %w", strategy, err)
		}

		result.StrategyResults[strategy.String()] = *strategyResult
		totalV1Score += strategyResult.V1AverageDeckScore
		totalV2Score += strategyResult.V2AverageDeckScore
		significantWins += countSignificant(strategyResult, WinnerV2, config.WinThreshold)
		significantLosses += countSignificant(strategyResult, WinnerV1, config.WinThreshold)
		totalDecks += len(strategyResult.DeckComparisons)
	}

	result.TotalDecks = totalDecks

	// Calculate summary
	if len(config.Strategies) > 0 {
		result.Summary.V1AverageScore = totalV1Score / float64(len(config.Strategies))
		result.Summary.V2AverageScore = totalV2Score / float64(len(config.Strategies))
	}

	if result.Summary.V1AverageScore > 0 {
		result.Summary.Improvement = ((result.Summary.V2AverageScore - result.Summary.V1AverageScore) / result.Summary.V1AverageScore) * 100
	}

	result.Summary.SignificantWins = significantWins
	result.Summary.SignificantLosses = significantLosses

	// Determine winner
	if result.Summary.V2AverageScore > result.Summary.V1AverageScore*(1+config.SignificanceThreshold) {
		result.Summary.Winner = WinnerV2
	} else if result.Summary.V1AverageScore > result.Summary.V2AverageScore*(1+config.SignificanceThreshold) {
		result.Summary.Winner = WinnerV1
	} else {
		result.Summary.Winner = WinnerTie
	}

	// Calculate metric breakdown
	result.MetricBreakdown = calculateMetricBreakdown(result.StrategyResults, config)

	// Generate recommendations
	result.Recommendations = generateRecommendations(result, config)

	return result, nil
}

// compareStrategy compares V1 vs V2 for a single strategy
func compareStrategy(strategy deck.Strategy, cardAnalysis deck.CardAnalysis, synergyDB *deck.SynergyDatabase, config ComparisonConfig) (*StrategyComparison, error) {
	result := &StrategyComparison{
		StrategyName: strategy.String(),
	}

	// Build both algorithm versions
	recV1, recV2, err := buildDeckWithAlgorithm(strategy, cardAnalysis)
	if err != nil {
		return nil, err
	}

	// Calculate scores
	v1Score := calculateAverageCardScore(recV1.DeckDetail)
	v2Cards := convertToCardCandidates(recV2.DeckDetail)
	v2Result := deck.ScoreDeckV2(v2Cards, strategy, synergyDB)

	result.V1AverageDeckScore = v1Score
	result.V2AverageDeckScore = v2Result.FinalScore

	// Evaluate both decks with quality test suite
	v1DeckCards := convertToCardCandidates(recV1.DeckDetail)
	v1Eval := evaluation.Evaluate(v1DeckCards, synergyDB, nil)
	v2Eval := evaluation.Evaluate(v2Cards, synergyDB, nil)

	// Create deck comparison
	deckComp := createDeckComparison(strategy, recV1, recV2, v1Score, v2Result, v2Eval, config)
	result.DeckComparisons = []DeckComparison{deckComp}

	// Calculate metrics
	result.Metrics = calculateStrategyMetrics(v1Eval, v2Eval, v2Result, config)

	// Determine strategy winner
	result.Winner = determineWinner(result.V2AverageDeckScore, result.V1AverageDeckScore, config.SignificanceThreshold)

	return result, nil
}

// buildDeckWithAlgorithm builds decks using both V1 and V2 algorithms
func buildDeckWithAlgorithm(strategy deck.Strategy, cardAnalysis deck.CardAnalysis) (*deck.DeckRecommendation, *deck.DeckRecommendation, error) {
	// Build deck using V1 algorithm (synergy disabled = V1 behavior)
	builderV1 := deck.NewBuilder("")
	if err := builderV1.SetStrategy(strategy); err != nil {
		return nil, nil, fmt.Errorf("V1 set strategy failed: %w", err)
	}
	recV1, err := builderV1.BuildDeckFromAnalysis(cardAnalysis)
	if err != nil {
		return nil, nil, fmt.Errorf("V1 build failed: %w", err)
	}

	// Build deck using V2 algorithm (with synergy enabled)
	builderV2 := deck.NewBuilder("")
	if err := builderV2.SetStrategy(strategy); err != nil {
		return nil, nil, fmt.Errorf("V2 set strategy failed: %w", err)
	}
	builderV2.SetSynergyEnabled(true)
	recV2, err := builderV2.BuildDeckFromAnalysis(cardAnalysis)
	if err != nil {
		return nil, nil, fmt.Errorf("V2 build failed: %w", err)
	}

	return recV1, recV2, nil
}

// createDeckComparison creates a DeckComparison object from build results
func createDeckComparison(strategy deck.Strategy, recV1, recV2 *deck.DeckRecommendation, v1Score float64, v2Result deck.ScorerV2Result, v2Eval evaluation.EvaluationResult, config ComparisonConfig) DeckComparison {
	deckComp := DeckComparison{
		DeckName:        fmt.Sprintf("%s Deck", strategy),
		V1Deck:          recV1.Deck,
		V1Score:         v1Score,
		V2Deck:          recV2.Deck,
		V2Score:         v2Result.FinalScore,
		V2Evaluation:    formatDeckEvaluation(v2Eval),
		ScoreDifference: v2Result.FinalScore - v1Score,
		Winner:          determineWinner(v2Result.FinalScore, v1Score, config.WinThreshold),
	}
	return deckComp
}

// formatDeckEvaluation formats evaluation results into EvaluationMetrics
func formatDeckEvaluation(eval evaluation.EvaluationResult) EvaluationMetrics {
	return EvaluationMetrics{
		OverallScore:        eval.OverallScore,
		AttackScore:         eval.Attack.Score,
		DefenseScore:        eval.Defense.Score,
		SynergyScore:        eval.Synergy.Score,
		VersatilityScore:    eval.Versatility.Score,
		Archetype:           eval.DetectedArchetype,
		ArchetypeConfidence: eval.ArchetypeConfidence,
		SynergyPairs:        eval.SynergyMatrix.PairCount,
	}
}

// determineWinner compares two scores using a threshold to determine the winner
func determineWinner(v2Score, v1Score, threshold float64) string {
	if v2Score > v1Score*(1+threshold) {
		return WinnerV2
	} else if v1Score > v2Score*(1+threshold) {
		return WinnerV1
	}
	return WinnerTie
}

// calculateStrategyMetrics computes detailed metrics for strategy comparison
func calculateStrategyMetrics(v1Eval, v2Eval evaluation.EvaluationResult, v2Result deck.ScorerV2Result, config ComparisonConfig) StrategyMetrics {
	metrics := StrategyMetrics{}

	// Synergy score
	metrics.SynergyScore = MetricComparison{
		V1Average:   v1Eval.Synergy.Score,
		V2Average:   v2Eval.Synergy.Score,
		Improvement: ((v2Eval.Synergy.Score - v1Eval.Synergy.Score) / maxf(v1Eval.Synergy.Score, 0.01)) * 100,
		Winner:      winnerForValue(v2Eval.Synergy.Score, v1Eval.Synergy.Score, config.SignificanceThreshold),
		Significant: math.Abs(v2Eval.Synergy.Score-v1Eval.Synergy.Score) > (v1Eval.Synergy.Score * config.SignificanceThreshold),
	}

	// Counter coverage (from defense score as proxy)
	metrics.CounterCoverage = MetricComparison{
		V1Average:   v1Eval.Defense.Score,
		V2Average:   v2Eval.Defense.Score,
		Improvement: ((v2Eval.Defense.Score - v1Eval.Defense.Score) / maxf(v1Eval.Defense.Score, 0.01)) * 100,
		Winner:      winnerForValue(v2Eval.Defense.Score, v1Eval.Defense.Score, config.SignificanceThreshold),
		Significant: math.Abs(v2Eval.Defense.Score-v1Eval.Defense.Score) > (v1Eval.Defense.Score * config.SignificanceThreshold),
	}

	// Archetype coherence
	metrics.ArchetypeCoherence = MetricComparison{
		V1Average:   v1Eval.ArchetypeConfidence * 10,
		V2Average:   v2Eval.ArchetypeConfidence * 10,
		Improvement: ((v2Eval.ArchetypeConfidence - v1Eval.ArchetypeConfidence) / maxf(v1Eval.ArchetypeConfidence, 0.01)) * 100,
		Winner:      winnerForValue(v2Eval.ArchetypeConfidence, v1Eval.ArchetypeConfidence, config.SignificanceThreshold),
		Significant: math.Abs(v2Eval.ArchetypeConfidence-v1Eval.ArchetypeConfidence) > config.SignificanceThreshold,
	}

	// Defensive capability
	metrics.DefensiveCapability = MetricComparison{
		V1Average:   v1Eval.Defense.Score,
		V2Average:   v2Eval.Defense.Score,
		Improvement: ((v2Eval.Defense.Score - v1Eval.Defense.Score) / maxf(v1Eval.Defense.Score, 0.01)) * 100,
		Winner:      winnerForValue(v2Eval.Defense.Score, v1Eval.Defense.Score, config.SignificanceThreshold),
		Significant: math.Abs(v2Eval.Defense.Score-v1Eval.Defense.Score) > (v1Eval.Defense.Score * config.SignificanceThreshold),
	}

	// Elixir range adherence (from V2 result)
	v1ElixirScore := calculateElixirFitScore(v1Eval.AvgElixir, deck.StrategyBalanced)
	metrics.ElixirRangeAdherence = MetricComparison{
		V1Average:   v1ElixirScore * 10,
		V2Average:   v2Result.ElixirFitScore * 10,
		Improvement: ((v2Result.ElixirFitScore - v1ElixirScore) / maxf(v1ElixirScore, 0.01)) * 100,
		Winner:      winnerForValue(v2Result.ElixirFitScore, v1ElixirScore, config.SignificanceThreshold),
		Significant: math.Abs(v2Result.ElixirFitScore-v1ElixirScore) > config.SignificanceThreshold,
	}

	// Card level distribution (F2P score as proxy)
	metrics.CardLevelDistribution = MetricComparison{
		V1Average:   v1Eval.F2PFriendly.Score,
		V2Average:   v2Eval.F2PFriendly.Score,
		Improvement: ((v2Eval.F2PFriendly.Score - v1Eval.F2PFriendly.Score) / maxf(v1Eval.F2PFriendly.Score, 0.01)) * 100,
		Winner:      winnerForValue(v2Eval.F2PFriendly.Score, v1Eval.F2PFriendly.Score, config.SignificanceThreshold),
		Significant: math.Abs(v2Eval.F2PFriendly.Score-v1Eval.F2PFriendly.Score) > (v1Eval.F2PFriendly.Score * config.SignificanceThreshold),
	}

	return metrics
}

// calculateMetricBreakdown aggregates metrics across all strategies
func calculateMetricBreakdown(strategyResults map[string]StrategyComparison, config ComparisonConfig) MetricBreakdown {
	breakdown := MetricBreakdown{}

	// Collect values for each metric
	var v1Synergy, v2Synergy []float64
	var v1Counter, v2Counter []float64
	var v1Archetype, v2Archetype []float64
	var v1Defense, v2Defense []float64
	var v1Elixir, v2Elixir []float64
	var v1F2P, v2F2P []float64

	for _, sr := range strategyResults {
		v1Synergy = append(v1Synergy, sr.Metrics.SynergyScore.V1Average)
		v2Synergy = append(v2Synergy, sr.Metrics.SynergyScore.V2Average)
		v1Counter = append(v1Counter, sr.Metrics.CounterCoverage.V1Average)
		v2Counter = append(v2Counter, sr.Metrics.CounterCoverage.V2Average)
		v1Archetype = append(v1Archetype, sr.Metrics.ArchetypeCoherence.V1Average)
		v2Archetype = append(v2Archetype, sr.Metrics.ArchetypeCoherence.V2Average)
		v1Defense = append(v1Defense, sr.Metrics.DefensiveCapability.V1Average)
		v2Defense = append(v2Defense, sr.Metrics.DefensiveCapability.V2Average)
		v1Elixir = append(v1Elixir, sr.Metrics.ElixirRangeAdherence.V1Average)
		v2Elixir = append(v2Elixir, sr.Metrics.ElixirRangeAdherence.V2Average)
		v1F2P = append(v1F2P, sr.Metrics.CardLevelDistribution.V1Average)
		v2F2P = append(v2F2P, sr.Metrics.CardLevelDistribution.V2Average)
	}

	breakdown.SynergyScore = calculateAggregateMetric(v1Synergy, v2Synergy, config)
	breakdown.CounterCoverage = calculateAggregateMetric(v1Counter, v2Counter, config)
	breakdown.ArchetypeCoherence = calculateAggregateMetric(v1Archetype, v2Archetype, config)
	breakdown.DefensiveCapability = calculateAggregateMetric(v1Defense, v2Defense, config)
	breakdown.ElixirRangeAdherence = calculateAggregateMetric(v1Elixir, v2Elixir, config)
	breakdown.CardLevelDistribution = calculateAggregateMetric(v1F2P, v2F2P, config)

	return breakdown
}

// calculateAggregateMetric computes aggregate statistics for a metric
func calculateAggregateMetric(v1, v2 []float64, config ComparisonConfig) AggregateMetric {
	am := AggregateMetric{}

	am.V1Mean = mean(v1)
	am.V2Mean = mean(v2)
	am.V1StdDev = stdDev(v1, am.V1Mean)
	am.V2StdDev = stdDev(v2, am.V2Mean)

	if am.V1Mean > 0 {
		am.Improvement = ((am.V2Mean - am.V1Mean) / am.V1Mean) * 100
	}

	am.Winner = winnerForValue(am.V2Mean, am.V1Mean, config.SignificanceThreshold)

	// Simplified p-value (would use proper t-test in production)
	am.PValue = calculateSimplifiedPValue(am.V2Mean, am.V1Mean, am.V2StdDev, am.V1StdDev, len(v1))

	return am
}

// generateRecommendations creates recommendations based on comparison results
func generateRecommendations(result *AlgorithmComparisonResult, config ComparisonConfig) Recommendations {
	recs := Recommendations{
		Reasoning:   []string{},
		NextSteps:   []string{},
		KnownIssues: []string{},
	}

	// Determine recommendation
	if result.Summary.Winner == WinnerV2 {
		recs.AlgorithmCutover = WinnerV2
		if result.Summary.Improvement > 20 {
			recs.Confidence = ConfidenceHigh
			recs.Reasoning = append(recs.Reasoning, fmt.Sprintf("V2 shows %.1f%% improvement over V1", result.Summary.Improvement))
		} else {
			recs.Confidence = ConfidenceMedium
			recs.Reasoning = append(recs.Reasoning, fmt.Sprintf("V2 shows modest %.1f%% improvement", result.Summary.Improvement))
		}
	} else if result.Summary.Winner == WinnerV1 {
		recs.AlgorithmCutover = WinnerV1
		recs.Confidence = ConfidenceHigh
		recs.Reasoning = append(recs.Reasoning, "V1 still outperforms V2")
		recs.KnownIssues = append(recs.KnownIssues, "V2 algorithm needs refinement")
	} else {
		recs.AlgorithmCutover = WinnerV1
		recs.Confidence = ConfidenceLow
		recs.Reasoning = append(recs.Reasoning, "No significant difference between V1 and V2")
	}

	// Add metric-specific reasoning
	if result.MetricBreakdown.SynergyScore.Winner == WinnerV2 {
		recs.Reasoning = append(recs.Reasoning, "V2 significantly improves synergy detection")
	}
	if result.MetricBreakdown.ArchetypeCoherence.Winner == WinnerV2 {
		recs.Reasoning = append(recs.Reasoning, "V2 produces more coherent archetypes")
	}
	if result.MetricBreakdown.CounterCoverage.Winner == WinnerV1 {
		recs.KnownIssues = append(recs.KnownIssues, "V2 counter coverage needs improvement")
	}

	// Next steps
	recs.NextSteps = append(recs.NextSteps, "Run comparison on larger player sample")
	recs.NextSteps = append(recs.NextSteps, "Test with meta deck fixtures")
	recs.NextSteps = append(recs.NextSteps, "Gather user feedback on recommended decks")

	if recs.AlgorithmCutover == WinnerV2 && recs.Confidence == ConfidenceHigh {
		recs.NextSteps = append(recs.NextSteps, "Proceed with V2 algorithm cutover")
	} else {
		recs.NextSteps = append(recs.NextSteps, "Continue refining V2 algorithm")
	}

	return recs
}

// ExportJSON exports comparison result as JSON
func (r *AlgorithmComparisonResult) ExportJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ExportMarkdown exports comparison result as Markdown report
func (r *AlgorithmComparisonResult) ExportMarkdown() string {
	var sb strings.Builder

	// Header
	sb.WriteString("# Algorithm Comparison Report\n\n")
	sb.WriteString(fmt.Sprintf("**Player Tag**: %s\n", r.PlayerTag))
	sb.WriteString(fmt.Sprintf("**Date**: %s\n", r.Timestamp))
	sb.WriteString(fmt.Sprintf("**Strategies Tested**: %d\n\n", len(r.Strategies)))
	sb.WriteString("---\n\n")

	// Summary
	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString(fmt.Sprintf("### Winner: **%s**\n\n", strings.ToUpper(r.Summary.Winner)))
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| V1 Average Score | %.4f |\n", r.Summary.V1AverageScore))
	sb.WriteString(fmt.Sprintf("| V2 Average Score | %.4f |\n", r.Summary.V2AverageScore))
	sb.WriteString(fmt.Sprintf("| Improvement | %.1f%% |\n", r.Summary.Improvement))
	sb.WriteString(fmt.Sprintf("| V2 Significant Wins | %d |\n", r.Summary.SignificantWins))
	sb.WriteString(fmt.Sprintf("| V2 Significant Losses | %d |\n\n", r.Summary.SignificantLosses))
	sb.WriteString("---\n\n")

	// Metric Breakdown
	sb.WriteString("## Metric Breakdown\n\n")
	sb.WriteString("| Metric | V1 Mean | V2 Mean | Improvement | Winner |\n")
	sb.WriteString("|--------|---------|---------|-------------|--------|\n")
	sb.WriteString(fmt.Sprintf("| Synergy Score | %.2f | %.2f | %.1f%% | %s |\n",
		r.MetricBreakdown.SynergyScore.V1Mean,
		r.MetricBreakdown.SynergyScore.V2Mean,
		r.MetricBreakdown.SynergyScore.Improvement,
		strings.ToUpper(r.MetricBreakdown.SynergyScore.Winner)))
	sb.WriteString(fmt.Sprintf("| Counter Coverage | %.2f | %.2f | %.1f%% | %s |\n",
		r.MetricBreakdown.CounterCoverage.V1Mean,
		r.MetricBreakdown.CounterCoverage.V2Mean,
		r.MetricBreakdown.CounterCoverage.Improvement,
		strings.ToUpper(r.MetricBreakdown.CounterCoverage.Winner)))
	sb.WriteString(fmt.Sprintf("| Archetype Coherence | %.2f | %.2f | %.1f%% | %s |\n\n",
		r.MetricBreakdown.ArchetypeCoherence.V1Mean,
		r.MetricBreakdown.ArchetypeCoherence.V2Mean,
		r.MetricBreakdown.ArchetypeCoherence.Improvement,
		strings.ToUpper(r.MetricBreakdown.ArchetypeCoherence.Winner)))
	sb.WriteString("---\n\n")

	// Recommendations
	sb.WriteString("## Recommendations\n\n")
	sb.WriteString(fmt.Sprintf("### Recommended Algorithm: **%s**\n\n", strings.ToUpper(r.Recommendations.AlgorithmCutover)))
	sb.WriteString(fmt.Sprintf("**Confidence**: %s\n\n", strings.ToUpper(r.Recommendations.Confidence)))
	sb.WriteString("**Reasoning**:\n")
	for _, r := range r.Recommendations.Reasoning {
		sb.WriteString(fmt.Sprintf("- %s\n", r))
	}
	sb.WriteString("\n")
	if len(r.Recommendations.KnownIssues) > 0 {
		sb.WriteString("**Known Issues**:\n")
		for _, i := range r.Recommendations.KnownIssues {
			sb.WriteString(fmt.Sprintf("- %s\n", i))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("**Next Steps**:\n")
	for _, s := range r.Recommendations.NextSteps {
		sb.WriteString(fmt.Sprintf("- %s\n", s))
	}

	return sb.String()
}

// Helper functions

func strategyNames(strategies []deck.Strategy) []string {
	names := make([]string, len(strategies))
	for i, s := range strategies {
		names[i] = s.String()
	}
	return names
}

func calculateAverageCardScore(cards []deck.CardDetail) float64 {
	if len(cards) == 0 {
		return 0
	}
	total := 0.0
	for _, card := range cards {
		total += card.Score
	}
	return total / float64(len(cards))
}

func convertToCardCandidates(details []deck.CardDetail) []deck.CardCandidate {
	candidates := make([]deck.CardCandidate, len(details))
	for i, d := range details {
		role := deck.CardRole(d.Role)
		candidates[i] = deck.CardCandidate{
			Name:              d.Name,
			Level:             d.Level,
			MaxLevel:          d.MaxLevel,
			Rarity:            d.Rarity,
			Elixir:            d.Elixir,
			Role:              &role,
			EvolutionLevel:    d.EvolutionLevel,
			MaxEvolutionLevel: d.MaxEvolutionLevel,
		}
	}
	return candidates
}

func countSignificant(sr *StrategyComparison, winner string, threshold float64) int {
	count := 0
	for _, dc := range sr.DeckComparisons {
		if dc.Winner == winner {
			if math.Abs(dc.ScoreDifference) > threshold {
				count++
			}
		}
	}
	return count
}

func calculateElixirFitScore(avgElixir float64, strategy deck.Strategy) float64 {
	// Simplified elixir fit calculation
	profile, exists := deck.StrategyElixirProfiles[strategy]
	if !exists {
		profile = deck.StrategyElixirProfiles[deck.StrategyBalanced]
	}

	if avgElixir >= profile.Min && avgElixir <= profile.Max {
		return 1.0
	}
	if avgElixir < profile.Min {
		return 1.0 - ((profile.Min - avgElixir) / 2.0)
	}
	return 1.0 - ((avgElixir - profile.Max) / 2.0)
}

func winnerForValue(v2, v1, threshold float64) string {
	if v2 > v1*(1+threshold) {
		return WinnerV2
	} else if v1 > v2*(1+threshold) {
		return WinnerV1
	}
	return WinnerTie
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stdDev(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	sumSqDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSqDiff += diff * diff
	}
	return math.Sqrt(sumSqDiff / float64(len(values)-1))
}

func calculateSimplifiedPValue(mean2, mean1, sd2, sd1 float64, n int) float64 {
	// Very simplified - would use proper t-test in production
	diff := mean2 - mean1
	pooledSD := (sd1 + sd2) / 2
	if pooledSD == 0 {
		return 1.0
	}
	z := diff / (pooledSD / math.Sqrt(float64(n)))
	if math.Abs(z) > 1.96 {
		return 0.05 // Significant
	}
	return 0.10 // Not significant
}
