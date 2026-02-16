# Algorithm Comparison

## Overview

The Clash Royale API now includes a comprehensive algorithm comparison framework that evaluates the **V1** (legacy) vs **V2** (synergy-enhanced) deck building algorithms across multiple quality metrics.

## Quick Start

```bash
# Compare algorithms on a real player
cr-api deck compare-algorithms --tag R8QGUQRCV

# Save comparison report to file
cr-api deck compare-algorithms --tag R8QGUQRCV --output comparison.md

# Export as JSON for further analysis
cr-api deck compare-algorithms --tag R8QGUQRCV --format json --output comparison.json

# Compare only specific strategies
cr-api deck compare-algorithms --tag R8QGUQRCV --strategies balanced,cycle,control

# Adjust significance thresholds
cr-api deck compare-algorithms --tag R8QGUQRCV --significance 0.03 --win-threshold 0.15
```

## Comparison Dimensions

The framework evaluates algorithms across four key dimensions:

### 1. Deck Quality
Which produces better decks overall? Measured by:
- Overall quality score (0-10 scale)
- Synergy detection and card interactions
- Counter coverage against common threats
- Archetype coherence and strategic fit

### 2. Archetype Purity
Which follows intended strategies better? Measured by:
- Archetype confidence scores
- Win condition clarity
- Elixir curve adherence
- Role distribution balance

### 3. Meta Viability
Which decks would perform better on ladder? Measured by:
- Attack potential (win conditions + spell damage)
- Defensive capability (anti-air, buildings, support)
- Versatility (role diversity, elixir variety)
- F2P friendliness (upgrade accessibility)

### 4. User Satisfaction
Which addresses user complaints better? Measured by:
- Card availability (playability score)
- Evolution optimization
- Synergistic combos detected
- Coherent strategy execution

## Metrics

The comparison framework tracks these metrics for each algorithm:

| Metric | Description | V1 Behavior | V2 Behavior |
|--------|-------------|-------------|-------------|
| **Synergy Score** | Card pair interactions | No detection | 188-pair database analysis |
| **Counter Coverage** | Defensive capability | Basic | WASTED framework analysis |
| **Archetype Coherence** | Strategic fit | Loose | Strict archetype enforcement |
| **Defensive Capability** | Anti-air, buildings | Implicit | Explicit scoring |
| **Elixir Range Adherence** | Curve fit to strategy | 15% weight | 25% weight (increased) |
| **Card Level Distribution** | F2P accessibility | High weight | Reduced weight (60% vs 120%) |

## Output Formats

### Markdown Report

```bash
cr-api deck compare-algorithms --tag R8QGUQRCV --format markdown
```

Generates a formatted markdown report with:
- Executive summary (winner, improvement %)
- Metric breakdown tables
- Per-strategy comparisons
- Recommendations with confidence levels

### JSON Export

```bash
cr-api deck compare-algorithms --tag R8QGUQRCV --format json --output comparison.json
```

Generates structured JSON with:
- Full comparison results
- Per-deck analysis
- Metric breakdowns
- Statistical significance data

## Example Output

### Markdown Summary

```markdown
# Algorithm Comparison Report

**Player Tag**: R8QGUQRCV
**Date**: 2026-02-01T12:00:00Z
**Strategies Tested**: 5

---

## Executive Summary

### Winner: **V2**

| Metric | Value |
|--------|-------|
| V1 Average Score | 0.6234 |
| V2 Average Score | 0.7121 |
| Improvement | 14.2% |
| V2 Significant Wins | 4 |
| V2 Significant Losses | 0 |

---

## Metric Breakdown

| Metric | V1 Mean | V2 Mean | Improvement | Winner |
|--------|---------|---------|-------------|--------|
| Synergy Score | 4.20 | 6.80 | 61.9% | V2 |
| Counter Coverage | 6.50 | 7.20 | 10.8% | V2 |
| Archetype Coherence | 6.00 | 7.50 | 25.0% | V2 |
```

### JSON Structure

```json
{
  "player_tag": "R8QGUQRCV",
  "timestamp": "2026-02-01T12:00:00Z",
  "summary": {
    "winning_algorithm": "v2",
    "v1_average_score": 0.6234,
    "v2_average_score": 0.7121,
    "improvement_percent": 14.2
  },
  "metric_breakdown": {
    "synergy_score": {
      "v1_mean": 4.20,
      "v2_mean": 6.80,
      "winner": "v2"
    }
  },
  "recommendations": {
    "recommended_algorithm": "v2",
    "confidence": "high",
    "reasoning": [
      "V2 shows 14.2% improvement over V1",
      "V2 significantly improves synergy detection",
      "V2 produces more coherent archetypes"
    ]
  }
}
```

## Programmatic Usage

### Go Library

```go
import "github.com/klauer/clash-royale-api/go/pkg/deck/comparison"

// Create default config
config := comparison.DefaultComparisonConfig()
config.PlayerTag = "R8QGUQRCV"
config.Strategies = []deck.Strategy{
    deck.StrategyBalanced,
    deck.StrategyCycle,
}

// Load card analysis (from API or file)
cardAnalysis := loadPlayerAnalysis("R8QGUQRCV")

// Run comparison
result, err := comparison.CompareAlgorithms("R8QGUQRCV", cardAnalysis, config)
if err != nil {
    log.Fatal(err)
}

// Check winner
if result.Summary.Winner == "v2" {
    fmt.Printf("V2 wins by %.1f%%\n", result.Summary.Improvement)
}

// Export report
markdown := result.ExportMarkdown()
json, _ := result.ExportJSON()
```

## Interpretation Guide

### Confidence Levels

| Confidence | Criteria | Recommendation |
|------------|-----------|----------------|
| **High** | Improvement > 20%, V2 wins most metrics | Proceed with V2 cutover |
| **Medium** | Improvement 5-20%, mixed metrics | Test with more players |
| **Low** | Improvement < 5%, or V1 wins | Keep V1, refine V2 |

### Significant Wins/Losses

A "significant win" occurs when one algorithm outperforms the other by more than the `--win-threshold` (default: 10%).

- **V2 Significant Wins**: V2 builds measurably better decks
- **V2 Significant Losses**: V2 regresses in quality (investigate metrics)

### Statistical Significance

The framework uses simplified statistical testing. Key metrics:
- **p-value < 0.05**: Statistically significant difference
- **p-value >= 0.05**: Difference may be due to chance

For production use, consider implementing proper t-tests or bootstrap sampling.

## Testing the Comparison

### Test on Meta Deck Fixtures

```bash
# Run quality tests
task test

# Test meta deck recognition
cr-api deck evaluate --deck "Golem-NightWitch-BabyDragon-Tornado-Lightning-MegaMinion-ElixirCollector-Lumberjack"
```

### Test on Real Player Data

```bash
# Test with multiple players
for tag in "R8QGUQRCV" "2P0GYQJ" "8VCGL8CG"; do
    cr-api deck compare-algorithms --tag $tag --output "comparison_${tag}.md"
done
```

### Test Specific Strategies

```bash
# Compare only cycle strategies
cr-api deck compare-algorithms --tag R8QGUQRCV --strategies cycle

# Compare beatdown vs control
cr-api deck compare-algorithms --tag R8QGUQRCV --strategies aggro,control
```

## Troubleshooting

### V2 Performs Worse Than V1

**Symptoms**: V2 loses significant comparisons, lower scores

**Potential Causes**:
1. Synergy database not loading properly
2. Strategy configuration mismatch
3. Elixir curve thresholds too strict

**Solutions**:
```bash
# Check synergy database is loaded
cr-api deck evaluate --deck "..." --verbose

# Relax elixir constraints
# Edit pkg/deck/scorer_v2.go: StrategyElixirProfiles

# Check strategy weights
cr-api deck build --tag R8QGUQRCV --strategy balanced --verbose
```

### Inconsistent Results

**Symptoms**: Results vary significantly between runs

**Potential Causes**:
1. Random seed in deck building (genetic algorithm)
2. Fuzz integration variability
3. Card level differences

**Solutions**:
- Run multiple comparisons and aggregate results
- Use `--variations` flag to test multiple decks per strategy
- Ensure consistent card analysis data

## Future Enhancements

### Planned Features

1. **Win Rate Prediction**: Integrate meta data to predict actual win rates
2. **User Feedback Collection**: Track user satisfaction with recommended decks
3. **Longitudinal Studies**: Track performance over time as users upgrade cards
4. **A/B Testing**: Deploy both algorithms and measure real-world usage

### Contributing

To improve the comparison framework:

1. Add new metrics to `pkg/deck/comparison/algorithm_comparison.go`
2. Enhance statistical testing in `calculateSimplifiedPValue`
3. Improve recommendations logic in `generateRecommendations`
4. Add visualization output formats (HTML, charts)

## References

- [Deck Analysis Suite](DECK_ANALYSIS_SUITE.md) - Quality metrics and testing
- [Deck Builder](DECK_BUILDER.md) - Algorithm design and implementation
- [Scoring V2](pkg/deck/scorer_v2.go) - Improved scoring implementation
- [Synergy Database](pkg/deck/synergy.go) - 188-pair synergy definitions
