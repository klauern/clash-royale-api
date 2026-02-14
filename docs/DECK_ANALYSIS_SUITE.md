# Deck Analysis Suite Documentation

Comprehensive guide to the deck analysis suite commands for systematic deck building, batch evaluation, and competitive analysis.

## Overview

The deck analysis suite provides:
- **Batch deck building** with multiple strategies and variations
- **Bulk evaluation** of decks with player context and scoring
- **Side-by-side comparison** with detailed analysis reports
- **Unified workflow** combining all phases in a single command
- **Export capabilities** for external analysis and sharing

This suite eliminates manual scaffolding for multi-deck analysis workflows, enabling systematic deck discovery and optimization.

## Core Concepts

### Build-Suite: Systematic Deck Generation

The `build-suite` command generates multiple deck variations across different strategies in a single invocation:

- **Strategies**: Different deck-building approaches (balanced, aggro, control, cycle, splash, spell)
- **Variations**: Multiple attempts per strategy to explore the possibility space
- **Suite Summary**: Aggregated metadata and file paths for all generated decks
- **Player Context**: Uses player card levels, evolutions, and collection for realistic builds

**Key Feature**: Saves both individual deck JSON files and a suite summary for batch processing.

### Evaluate-Batch: Bulk Deck Scoring

The `evaluate-batch` command evaluates multiple decks in one operation:

- **Input Sources**: Load from suite summary or directory of deck JSON files
- **Scoring Categories**: Attack, Defense, Synergy, Versatility, F2P-Friendly, Playability
- **Sorting & Filtering**: Rank by any category, filter by archetype, show top N
- **Context-Aware**: Incorporates player card levels for accurate F2P and playability scores
- **Multiple Formats**: Output as summary tables, JSON, CSV, or detailed text

**Key Feature**: Aggregates results with sort/filter options for quick deck discovery.

### Compare: Head-to-Head Analysis

The `compare` command provides side-by-side deck comparison:

- **Direct Comparison**: Compare up to 5 decks by deck string
- **Auto-Selection**: Load top N decks from evaluation results
- **Rich Formats**: Table (ASCII), Markdown (reports), JSON, CSV
- **Category Winners**: Highlights which deck excels in each category
- **Detailed Reports**: Comprehensive markdown with strengths/weaknesses per deck

**Key Feature**: Generates publication-ready markdown reports with recommendations.

### Analyze-Suite: Unified Workflow

The `analyze-suite` command chains build-suite → evaluate-batch → compare:

- **Three Phases**: Build, evaluate, compare in one command
- **Automatic Flow**: Seamlessly passes data between phases
- **Organized Output**: Structured directory with decks/, evaluations/, reports/
- **Top N Selection**: Automatically selects best performers for comparison report
- **Verbose Progress**: Track each phase with timing and statistics

**Key Feature**: End-to-end workflow from player tag to comprehensive analysis report.

## Commands

### build-suite

Build multiple deck variations with different strategies.

```bash
# Build all strategies with 2 variations each
./bin/cr-api deck build-suite --tag <TAG> --strategies all --variations 2

# Specific strategies with constraints
./bin/cr-api deck build-suite \
  --tag <TAG> \
  --strategies balanced,aggro,cycle \
  --variations 3 \
  --min-elixir 2.8 \
  --max-elixir 4.0 \
  --output-dir data/my-decks

# Must include specific cards (your highest-level cards)
./bin/cr-api deck build-suite \
  --tag <TAG> \
  --strategies all \
  --variations 2 \
  --include-cards "Hog Rider,Log,Fireball"

# Exclude certain cards (cards you don't like playing)
./bin/cr-api deck build-suite \
  --tag <TAG> \
  --strategies all \
  --variations 1 \
  --exclude-cards "Elite Barbarians,Royal Giant,Mega Knight"

# Offline mode (no API call, uses existing player data)
./bin/cr-api deck build-suite \
  --tag <TAG> \
  --from-analysis \
  --strategies balanced \
  --variations 5
```

**Flags:**
- `--tag <TAG>` - Player tag (required, without #)
- `--strategies <list>` - Strategies: balanced, aggro, control, cycle, splash, spell, all (default: balanced)
- `--variations <n>` - Variations per strategy (default: 1)
- `--output-dir <dir>` - Output directory (default: data/decks/)
- `--save` - Save deck files and summary JSON (default: true)
- `--from-analysis` - Offline mode with cached player data
- `--min-elixir <float>` - Minimum average elixir (default: 2.5)
- `--max-elixir <float>` - Maximum average elixir (default: 4.5)
- `--include-cards <cards>` - Must-include cards (comma-separated)
- `--exclude-cards <cards>` - Must-exclude cards (comma-separated)

**Output:**
- Individual deck files: `{timestamp}_deck_{strategy}_var{N}_{tag}.json`
- Suite summary: `{timestamp}_deck_suite_summary_{tag}.json`
- Console: Progress and statistics

### evaluate-batch

Evaluate multiple decks with comprehensive scoring.

```bash
# Evaluate from suite summary
./bin/cr-api deck evaluate-batch \
  --from-suite data/decks/20240110_120000_deck_suite_summary_TAG.json

# Evaluate from directory with player context
./bin/cr-api deck evaluate-batch \
  --deck-dir data/decks \
  --tag <TAG> \
  --verbose

# Sort by attack, show top 5
./bin/cr-api deck evaluate-batch \
  --from-suite data/decks/suite_TAG.json \
  --sort-by attack \
  --top-only \
  --top-n 5

# Filter by archetype
./bin/cr-api deck evaluate-batch \
  --from-suite data/decks/suite_TAG.json \
  --filter-archetype \
  --archetype cycle \
  --sort-by versatility

# Export to CSV
./bin/cr-api deck evaluate-batch \
  --from-suite data/decks/suite_TAG.json \
  --tag <TAG> \
  --format csv \
  --output-dir data/evaluations \
  --timing
```

**Flags:**
- `--from-suite <file>` - Load from suite summary JSON (mutually exclusive with --deck-dir)
- `--deck-dir <dir>` - Load from directory of deck JSON files
- `--tag <TAG>` - Player tag for context-aware evaluation
- `--format <format>` - Output format: summary, json, csv, detailed (default: summary)
- `--output-dir <dir>` - Save results to directory
- `--sort-by <criteria>` - Sort by: overall, attack, defense, synergy, versatility, f2p, playability, elixir
- `--top-only` - Show only top N decks
- `--top-n <n>` - Number of top decks (default: 10)
- `--filter-archetype` - Enable archetype filtering
- `--archetype <type>` - Archetype: beatdown, control, cycle, bridge, siege, bait, graveyard, miner, hybrid
- `--verbose` - Show detailed progress
- `--timing` - Show timing per deck
- `--save-aggregated` - Save results to output-dir (default: true)

**Output:**
- Evaluation results: `{timestamp}_deck_evaluations_{tag}.json` (if --output-dir specified)
- Console: Sorted/filtered deck rankings with scores

**Sort Criteria:**
- `overall` - Overall deck score (weighted average of all categories)
- `attack` - Attack capability
- `defense` - Defense capability
- `synergy` - Card synergies and combos
- `versatility` - Adaptability across matchups
- `f2p` / `f2p-friendly` - F2P friendliness (card rarity/levels)
- `playability` - How playable/intuitive the deck is
- `elixir` - Average elixir cost (ascending)

### compare

Compare multiple decks side-by-side with detailed analysis.

```bash
# Compare two decks directly
./bin/cr-api deck compare \
  --decks "Knight-Archers-Fireball-Musketeer-Hog Rider-Ice Spirit-Cannon-Log" \
  --decks "Giant-Witch-Skeleton Army-Musketeer-Fireball-Zap-Ice Golem-Archers" \
  --names "Hog Cycle" \
  --names "Giant Beatdown" \
  --format table

# Auto-select top 5 from evaluations and generate report
./bin/cr-api deck compare \
  --from-evaluations data/evaluations/20240110_deck_evaluations_TAG.json \
  --auto-select-top 5 \
  --format markdown \
  --report-output data/reports/top5_comparison.md \
  --verbose

# Compare top 3 as JSON for external analysis
./bin/cr-api deck compare \
  --from-evaluations data/evaluations/evaluations_TAG.json \
  --auto-select-top 3 \
  --format json \
  --output data/analysis.json

# Detailed comparison with win rate predictions
./bin/cr-api deck compare \
  --decks "Deck1" \
  --decks "Deck2" \
  --decks "Deck3" \
  --format table \
  --verbose \
  --winrate
```

**Flags:**
- `--decks <deck>` - Deck string (Card1-Card2-...-Card8), specify multiple times (max 5)
- `--names <name>` - Custom name for each deck (optional, in order)
- `--from-evaluations <file>` - Load from evaluation batch results (alternative to --decks)
- `--auto-select-top <n>` - Auto-select top N by score (requires --from-evaluations)
- `--format <format>` - Output format: table, json, csv, markdown/md (default: table)
- `--output <file>` - Output file path (default: stdout)
- `--report-output <file>` - Generate comprehensive markdown report
- `--verbose` - Show detailed comparison with strengths/weaknesses
- `--winrate` - Show predicted win rate comparison

**Deck String Format:** `"Card1-Card2-Card3-Card4-Card5-Card6-Card7-Card8"` (exactly 8 cards, hyphen-separated)

**Output:**
- Table format: ASCII table with emojis, category scores, winners
- Markdown format: Comprehensive report with executive summary, rankings, detailed analysis
- JSON format: Structured comparison data for programmatic use
- CSV format: Tabular data for spreadsheet analysis

### analyze-suite

Unified workflow: build, evaluate, compare in one command.

```bash
# Full analysis with all strategies, 2 variations, top 5 comparison
./bin/cr-api deck analyze-suite --tag <TAG> --strategies all --variations 2

# Focused analysis: specific strategies, 3 variations, top 3 comparison
./bin/cr-api deck analyze-suite \
  --tag <TAG> \
  --strategies balanced,aggro,cycle \
  --variations 3 \
  --top-n 3 \
  --output-dir data/analysis \
  --verbose

# Constrained analysis with card filters
./bin/cr-api deck analyze-suite \
  --tag <TAG> \
  --strategies all \
  --variations 2 \
  --min-elixir 2.8 \
  --max-elixir 4.2 \
  --include-cards "Hog Rider,Log" \
  --exclude-cards "Elite Barbarians,Royal Giant" \
  --top-n 5

# Offline mode (instant, no API calls)
./bin/cr-api deck analyze-suite \
  --tag <TAG> \
  --from-analysis \
  --strategies all \
  --variations 1 \
  --top-n 5
```

**Flags:**
- `--tag <TAG>` - Player tag (required, without #)
- `--strategies <list>` - Strategies: balanced, aggro, control, cycle, splash, spell, all (default: all)
- `--variations <n>` - Variations per strategy (default: 1)
- `--output-dir <dir>` - Base output directory (default: data/analysis)
- `--top-n <n>` - Top decks for comparison report (default: 5, max: 5)
- `--from-analysis` - Offline mode with cached player data
- `--min-elixir <float>` - Minimum average elixir (default: 2.5)
- `--max-elixir <float>` - Maximum average elixir (default: 4.5)
- `--include-cards <cards>` - Must-include cards (comma-separated)
- `--exclude-cards <cards>` - Must-exclude cards (comma-separated)
- `--verbose` - Show detailed progress

**Workflow:**

**Phase 1: Building Deck Variations**
- Generates strategy × variations decks
- Applies elixir and card constraints
- Saves individual JSON files + suite summary

**Phase 2: Evaluating All Decks**
- Evaluates each deck with player context
- Scores all categories (attack, defense, synergy, etc.)
- Saves evaluation results JSON

**Phase 3: Comparing Top Performers**
- Selects top N decks by overall score
- Generates comprehensive markdown comparison report
- Saves report with recommendations

**Output Structure:**
```
data/analysis/
├── decks/
│   ├── {timestamp}_deck_balanced_var1_{tag}.json
│   ├── {timestamp}_deck_aggro_var1_{tag}.json
│   ├── ... (individual deck files)
│   └── {timestamp}_deck_suite_summary_{tag}.json
├── evaluations/
│   └── {timestamp}_deck_evaluations_{tag}.json
└── reports/
    └── {timestamp}_deck_analysis_report_{tag}.md
```

## Evaluation Scoring System

All deck evaluations produce scores across six categories, each scored 0-10:

| Category | Description | What It Measures |
|----------|-------------|------------------|
| **Attack** | Offensive capability | Win condition viability, damage potential, push strength |
| **Defense** | Defensive capability | Survivability, counter potential, building defenses, reset/retarget coverage |
| **Synergy** | Card interactions | Combo potential, card pair synergies, archetype coherence |
| **Versatility** | Adaptability | Performance across different matchups and metas |
| **F2P Friendly** | Accessibility | Card rarity distribution, upgrade requirements |
| **Playability** | Ease of use | Intuitive gameplay, skill floor, collection availability |

**Overall Score:** Weighted average of all categories (0-10)

**Rating Scale:**
- 9.0-10.0: Godly!
- 8.0-8.9: Amazing
- 7.0-7.9: Great
- 6.0-6.9: Good
- 5.0-5.9: Decent
- 4.0-4.9: Mediocre
- 3.0-3.9: Poor
- 2.0-2.9: Bad
- 1.0-1.9: Terrible
- 0.0-0.9: Awful

## Detected Archetypes

The evaluation system automatically detects deck archetypes:

| Archetype | Characteristics | Example Cards |
|-----------|----------------|---------------|
| **Beatdown** | Heavy tank-based, high elixir | Golem, Giant, Lava Hound |
| **Control** | Defensive, spell control | X-Bow, Tesla, Tornado |
| **Cycle** | Fast-cycling, low elixir | Ice Spirit, Skeletons, Hog Rider |
| **Bridge** | Aggressive bridge spam | Battle Ram, Bandit, Royal Ghost |
| **Siege** | Building-based | X-Bow, Mortar, Princess |
| **Bait** | Spell bait mechanics | Goblin Barrel, Princess, Skeleton Army |
| **Graveyard** | Graveyard-focused | Graveyard, Freeze, Poison |
| **Miner** | Miner chip damage | Miner, Poison, Wallbreakers |
| **Hybrid** | Multiple win conditions | Various |
| **Unknown** | Doesn't fit archetypes | Various |

Archetype detection helps filter and compare decks with similar playstyles.

## Advanced Workflows

### Workflow 1: Find Your Best Deck

**Goal:** Discover which deck archetype suits your collection best.

```bash
# Step 1: Build comprehensive suite (all strategies, 3 variations)
./bin/cr-api deck build-suite \
  --tag <TAG> \
  --strategies all \
  --variations 3 \
  --output-dir data/my-analysis

# Step 2: Evaluate with player context and sort by overall score
./bin/cr-api deck evaluate-batch \
  --from-suite data/my-analysis/decks/*_deck_suite_summary_*.json \
  --tag <TAG> \
  --sort-by overall \
  --format csv \
  --output-dir data/my-analysis/evaluations \
  --verbose

# Step 3: Compare top 5 performers with detailed report
./bin/cr-api deck compare \
  --from-evaluations data/my-analysis/evaluations/*_deck_evaluations_*.json \
  --auto-select-top 5 \
  --format markdown \
  --report-output data/my-analysis/top5_report.md \
  --verbose

# Step 4: Review the report and pick your deck
cat data/my-analysis/top5_report.md
```

**OR use unified workflow:**
```bash
./bin/cr-api deck analyze-suite \
  --tag <TAG> \
  --strategies all \
  --variations 3 \
  --top-n 5 \
  --output-dir data/my-analysis \
  --verbose
```

### Workflow 2: Optimize for F2P

**Goal:** Find decks that maximize effectiveness with your current card levels.

```bash
# Build suite with F2P focus (include your highest-level cards)
./bin/cr-api deck build-suite \
  --tag <TAG> \
  --strategies all \
  --variations 2 \
  --include-cards "Knight,Archers,Fireball" \
  --output-dir data/f2p-decks

# Evaluate and sort by F2P friendliness
./bin/cr-api deck evaluate-batch \
  --from-suite data/f2p-decks/*_deck_suite_summary_*.json \
  --tag <TAG> \
  --sort-by f2p \
  --top-only \
  --top-n 10

# Compare top 3 F2P-friendly decks
./bin/cr-api deck compare \
  --from-evaluations data/f2p-decks/*_deck_evaluations_*.json \
  --auto-select-top 3 \
  --format markdown \
  --report-output data/f2p-decks/f2p_comparison.md
```

### Workflow 3: Explore Specific Archetype

**Goal:** Find the best cycle deck variations for your collection.

```bash
# Build cycle-focused variations
./bin/cr-api deck build-suite \
  --tag <TAG> \
  --strategies cycle \
  --variations 5 \
  --min-elixir 2.5 \
  --max-elixir 3.5 \
  --output-dir data/cycle-decks

# Evaluate and filter by cycle archetype
./bin/cr-api deck evaluate-batch \
  --from-suite data/cycle-decks/*_deck_suite_summary_*.json \
  --tag <TAG> \
  --filter-archetype \
  --archetype cycle \
  --sort-by versatility \
  --format json \
  --output-dir data/cycle-decks/evaluations

# Compare all cycle variations
./bin/cr-api deck compare \
  --from-evaluations data/cycle-decks/evaluations/*_deck_evaluations_*.json \
  --auto-select-top 5 \
  --format table \
  --verbose
```

### Workflow 4: Meta Analysis and Adaptation

**Goal:** Analyze multiple archetypes to understand matchup dynamics.

```bash
# Build representative decks for different archetypes
./bin/cr-api deck build-suite \
  --tag <TAG> \
  --strategies balanced,aggro,control,cycle \
  --variations 2 \
  --output-dir data/meta-analysis

# Evaluate with detailed metrics
./bin/cr-api deck evaluate-batch \
  --from-suite data/meta-analysis/*_deck_suite_summary_*.json \
  --tag <TAG> \
  --format detailed \
  --output-dir data/meta-analysis/evaluations \
  --timing \
  --verbose

# Compare top performers across archetypes
./bin/cr-api deck compare \
  --from-evaluations data/meta-analysis/evaluations/*_deck_evaluations_*.json \
  --auto-select-top 5 \
  --format markdown \
  --report-output data/meta-analysis/meta_report.md \
  --winrate \
  --verbose
```

### Workflow 5: Offline Experimentation

**Goal:** Rapidly test different deck constraints without API calls.

```bash
# Step 1: Build initial suite online (fetches player data once)
./bin/cr-api deck build-suite \
  --tag <TAG> \
  --strategies all \
  --variations 1 \
  --output-dir data/offline-base

# Step 2: Experiment offline with different constraints
./bin/cr-api deck build-suite \
  --tag <TAG> \
  --from-analysis \
  --strategies cycle,aggro \
  --variations 5 \
  --min-elixir 2.5 \
  --max-elixir 3.2 \
  --output-dir data/experiment1

./bin/cr-api deck build-suite \
  --tag <TAG> \
  --from-analysis \
  --strategies balanced,control \
  --variations 5 \
  --min-elixir 3.5 \
  --max-elixir 4.5 \
  --output-dir data/experiment2

# Step 3: Evaluate both experiments
./bin/cr-api deck evaluate-batch --deck-dir data/experiment1 --tag <TAG>
./bin/cr-api deck evaluate-batch --deck-dir data/experiment2 --tag <TAG>
```

## Best Practices

1. **Start broad, then narrow**: Use `--strategies all --variations 2` for baseline, then focus on top archetypes
2. **Use player context**: Always include `--tag` in evaluate-batch for accurate F2P and playability scores
3. **Constrain elixir range**: Use `--min-elixir` and `--max-elixir` to match your playstyle (e.g., 2.5-3.5 for cycle)
4. **Include high-level cards**: Use `--include-cards` to ensure your strongest cards are in all decks
5. **Leverage offline mode**: Use `--from-analysis` for rapid iteration after initial API fetch
6. **Save comprehensive reports**: Use `--format markdown --report-output` for shareable analysis
7. **Monitor timing**: Use `--timing --verbose` for large batch operations to track progress
8. **Filter by archetype**: Use `--filter-archetype` to focus analysis on specific playstyles
9. **Compare strategically**: Use `--auto-select-top 3` for focused comparison vs `--top-n 5` for broader view
10. **Iterate systematically**: Start with analyze-suite for end-to-end flow, then use individual commands for refinement

## Integration with Other Tools

### Export to Spreadsheet Analysis

```bash
# Generate CSV evaluation results
./bin/cr-api deck evaluate-batch \
  --from-suite data/decks/suite_TAG.json \
  --tag <TAG> \
  --format csv \
  --output-dir data/exports

# Open in Excel/Google Sheets for custom analysis
# - Pivot tables for archetype comparison
# - Charts for score distributions
# - Conditional formatting for top performers
```

### Combine with What-If Analysis

```bash
# Step 1: Find your best deck with analyze-suite
./bin/cr-api deck analyze-suite --tag <TAG> --strategies all --variations 2

# Step 2: Extract top deck from report
# (e.g., Cycle variation 1 with cards: Knight, Archers, Hog Rider, ...)

# Step 3: Simulate upgrades for that specific deck
./bin/cr-api what-if \
  --tag <TAG> \
  --upgrade "Knight:15" \
  --upgrade "Archers:15" \
  --show-decks \
  --save
```

### Track Analysis Over Time

```bash
# Create timestamped analysis directories
DATE=$(date +%Y%m%d)
./bin/cr-api deck analyze-suite \
  --tag <TAG> \
  --strategies all \
  --variations 2 \
  --output-dir data/analysis_${DATE} \
  --verbose

# Compare reports across weeks/months to track collection growth
# - data/analysis_20240101/
# - data/analysis_20240108/
# - data/analysis_20240115/
```

## Troubleshooting

| Issue | Cause | Solution |
|-------|-------|----------|
| "No decks generated" | Player collection too limited or constraints too strict | Relax elixir constraints, reduce --variations, or remove --include-cards |
| "Failed to fetch player data" | Invalid API token or player tag | Check `.env` for CLASH_ROYALE_API_TOKEN, verify tag format (no #) |
| "Suite summary not found" | Incorrect file path or --save=false | Verify file path, ensure --save is true (default) |
| Evaluation shows 0% playability | Player data not loaded | Add `--tag <TAG>` to evaluate-batch for context-aware evaluation |
| Compare shows "too many decks" | More than 5 decks selected | Use `--auto-select-top 5` (max) or reduce deck count |
| Slow batch evaluation | Large suite + API calls | Use `--from-analysis` for offline mode, or reduce --variations |
| Missing markdown report | --report-output not specified | Add `--report-output <file>` to compare command |
| Archetype filter shows no results | No decks match archetype | Remove filter or use `--archetype unknown` for unclassified decks |

## Related Documentation

- [CLI_REFERENCE.md](CLI_REFERENCE.md) - Complete CLI command reference
- [DECK_BUILDER.md](DECK_BUILDER.md) - Deck building algorithm details
- [CSV_EXPORTS.md](CSV_EXPORTS.md) - CSV export formats and integration
- [TESTING.md](TESTING.md) - Testing deck analysis suite features
