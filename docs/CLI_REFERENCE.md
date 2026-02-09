# CLI Reference

Complete reference for all `cr-api` CLI commands and task runner commands.

## Task Runner Commands (Recommended)

Install: `./scripts/install-task.sh` or from https://taskfile.dev

```bash
task              # Show all tasks
task setup        # Set up env (.env, deps, build)
task build        # Build binaries
task run -- '#TAG'        # Analyze player
task run-with-save -- '#TAG'    # Analyze + save JSON
task export-csv -- '#TAG'       # Export to CSV
task scan-events -- '#TAG'      # Scan event decks
task test         # Run tests with coverage
task lint         # Run golangci-lint
task snapshot     # Test release locally
task release      # Create release (requires GITHUB_TOKEN)
```

## Direct CLI Usage

Build: `cd go && go build -o bin/cr-api ./cmd/cr-api`

### Player Commands

```bash
./bin/cr-api player --tag <TAG> [--chests] [--save] [--export-csv]
./bin/cr-api cards [--export-csv]
./bin/cr-api analyze --tag <TAG> [--save] [--export-csv]
```

### Deck Building

```bash
./bin/cr-api deck build --tag <TAG> [--combat-stats-weight 0.25] [--disable-combat-stats]
./bin/cr-api deck build --tag <TAG> --strategy cycle --verbose
./bin/cr-api deck build --tag <TAG> --enable-synergy --synergy-weight 0.25
```

**Strategies**: `balanced` (default), `aggro`, `control`, `cycle`, `splash`, `spell`

### Deck Evaluation with Player Context

The `deck evaluate` command supports player context flags that enhance evaluation accuracy:

```bash
# Basic evaluation (no player context)
./bin/cr-api deck evaluate --deck "Knight-Archers-Fireball-Musketeer-Hog Rider-Ice Spirit-Cannon-Log"

# Evaluation with player context (fetches card levels from API)
./bin/cr-api deck evaluate --deck "Knight-Archers-Fireball-Musketeer-Hog Rider-Ice Spirit-Cannon-Log" --tag PLAYER_TAG

# Evaluation with arena context (manual arena level)
./bin/cr-api deck evaluate --deck "Knight-Archers-Fireball-Musketeer-Hog Rider-Ice Spirit-Cannon-Log" --arena 10

# Combined: player context + upgrade impact analysis
./bin/cr-api deck evaluate --deck "Knight-Archers-Fireball-Musketeer-Hog Rider-Ice Spirit-Cannon-Log" --tag PLAYER_TAG --show-upgrade-impact
```

**Player Context Flags:**
- `--tag <PLAYER_TAG>` - Fetches player data from API for:
  - Card level information (exact levels, not just defaults)
  - Arena level for unlock validation
  - Evolution status
  - Collection-aware playability scoring
- `--arena <arena_id>` - Sets arena level for unlock validation:
  - `0` = No arena restrictions (training camp mode)
  - `1-14` = Specific arena level (1=Training Camp, 14=Champion)
  - Useful for evaluating decks at different progression stages

**What Player Context Changes:**

Without context (basic evaluation):
```
Deck Score: 75.2
Playability: 100% (assumes all cards available)
Average Level: N/A (uses default levels)
```

With `--tag PLAYER_TAG` (context-aware evaluation):
```
Deck Score: 82.7
Playability: 87.5% (7/8 cards owned, 1 locked)
Average Level: 11.3/14 (uses actual card levels)
Upgrade Gap: 12 levels below max
Missing: Miner (locked - unlocks in Arena 10)
```

**Example Output Comparison:**

```bash
# Without context
$ ./bin/cr-api deck evaluate --deck "Hog Rider-Fireball-Ice Spirit-Log-Musketeer-Cannon-Knight-Archers"

Deck Evaluation
═══════════════
Deck Score: 78.5
Playability: 100.0%
Win Condition: Hog Rider
Average Elixir: 3.1
```

```bash
# With player context
$ ./bin/cr-api deck evaluate --deck "Hog Rider-Fireball-Ice Spirit-Log-Musketeer-Cannon-Knight-Archers" --tag PLAYER_TAG

Player Context Loaded: PlayerName (PLAYER_TAG), Arena: Arena 12

Deck Evaluation (Context-Aware)
═══════════════════════════════
Deck Score: 84.2 (with level bonuses)
Playability: 87.5% (7/8 cards playable)
Win Condition: Hog Rider (Level 12/14)
Average Elixir: 3.1
Average Card Level: 11.1/14

Card Levels
═══════════
Hog Rider      12/14  ✓  Owned
Fireball       11/14  ✓  Owned
Ice Spirit      9/14  ○  Owned (underleveled)
Log            13/14  ✓  Owned
Musketeer      12/14  ✓  Owned
Cannon         11/14  ✓  Owned
Knight          8/14  ○  Owned (underleveled)
Archers         7/14  ○  Owned (underleveled)

Missing Cards
═══════════════
None - All cards in collection

Upgrade Impact
═══════════════
Top 5 Upgrades:
1. Knight: 8→13 (+2.3 score, +4500 gold)
2. Archers: 7→12 (+2.1 score, +4000 gold)
3. Ice Spirit: 9→14 (+1.8 score, +2000 gold)
...
```

**Arena Validation Examples:**

```bash
# Check if deck is playable at Arena 5
$ ./bin/cr-api deck evaluate --deck "Giant-Witch-Skeleton Army-Musketeer-Fireball-Zap-Ice Golem-Archers" --arena 5

Arena Validation (Arena 5: Spell Valley)
═══════════════════════════════════════
✓ Giant (unlocked Arena 1)
✓ Witch (unlocked Arena 2)
✗ Skeleton Army (locked - unlocks Arena 8)
✓ Musketeer (unlocked Arena 4)
✓ Fireball (unlocked Arena 1)
✓ Zap (unlocked Arena 1)
✗ Ice Golem (locked - unlocks Arena 6)
✓ Archers (unlocked Arena 1)

Result: 6/8 cards playable (75%)
Missing: Skeleton Army (Arena 8), Ice Golem (Arena 6)
```

**Troubleshooting Player Context:**

| Issue | Solution |
|-------|----------|
| `Failed to fetch player data` | Check `CLASH_ROYALE_API_TOKEN` in `.env` and verify player tag format (without `#`) |
| `Player context not loaded` | Ensure tag is correct and player profile is public |
| `Evaluation without player context` | Warning message appears when API fetch fails; evaluation continues with default values |
| `Arena level mismatch` | Use `--arena` flag to manually set arena if auto-detection is incorrect |

**Best Practices:**
1. Always use `--tag` for accurate deck evaluation with real card levels
2. Use `--show-upgrade-impact` to identify priority upgrades
3. Use `--arena` when evaluating decks for specific progression stages
4. Combine with `--format json` for programmatic analysis
5. Check playability percentage before committing to a deck build

### Batch Deck Building and Evaluation

Build multiple deck variations systematically and evaluate them in batch:

```bash
# Build multiple deck variations with different strategies
./bin/cr-api deck build-suite --tag <TAG> --strategies balanced,aggro,cycle --variations 3

# Build all strategies with 2 variations each
./bin/cr-api deck build-suite --tag <TAG> --strategies all --variations 2 --output-dir data/decks

# Evaluate all decks from a build-suite run
./bin/cr-api deck evaluate-batch --from-suite data/decks/20240110_120000_deck_suite_summary_TAG.json

# Evaluate decks from a directory
./bin/cr-api deck evaluate-batch --deck-dir data/decks --format csv --output-dir data/evaluations

# Filter and sort evaluations
./bin/cr-api deck evaluate-batch --from-suite data/decks/suite.json --sort-by attack --top-only --top-n 5

# Evaluate with player context
./bin/cr-api deck evaluate-batch --from-suite data/decks/suite.json --tag TAG --verbose
```

**build-suite Flags:**
- `--strategies <list>` - Comma-separated strategies or 'all' (default: balanced)
  - Available: balanced, aggro, control, cycle, splash, spell, all
- `--variations <n>` - Number of variations per strategy (default: 1)
- `--output-dir <dir>` - Output directory for deck files (default: data/decks/)
- `--save` - Save individual deck files and summary JSON (default: true)
- `--from-analysis` - Use offline mode with pre-analyzed player data
- `--min-elixir`, `--max-elixir` - Deck elixir constraints
- `--include-cards`, `--exclude-cards` - Card filters

**evaluate-batch Flags:**
- `--from-suite <file>` - Load decks from build-suite summary JSON
- `--deck-dir <dir>` - Load decks from directory of JSON files
- `--tag <TAG>` - Player tag for context-aware evaluation
- `--format <format>` - Output format: summary, json, csv, detailed (default: summary)
- `--output-dir <dir>` - Save results to directory
- `--sort-by <criteria>` - Sort by: overall, attack, defense, synergy, versatility, f2p, playability, elixir
- `--top-only` - Show only top N decks
- `--top-n <n>` - Number of top decks (default: 10)
- `--filter-archetype` - Filter by archetype (use with --archetype)
- `--archetype <type>` - Archetype to filter (e.g., beatdown, control, cycle)
- `--verbose` - Show detailed progress
- `--timing` - Show timing information

**Example Workflow:**
```bash
# 1. Build deck suite with all strategies, 3 variations each
./bin/cr-api deck build-suite --tag R8QGUQRCV --strategies all --variations 3 --output-dir data/my-decks

# Output:
#   data/my-decks/20240110_120000_deck_suite_summary_R8QGUQRCV.json
#   data/my-decks/20240110_120000_deck_balanced_var1_R8QGUQRCV.json
#   data/my-decks/20240110_120000_deck_balanced_var2_R8QGUQRCV.json
#   ... (18 deck files total for 6 strategies × 3 variations)

# 2. Evaluate all decks from suite with player context
./bin/cr-api deck evaluate-batch \
  --from-suite data/my-decks/20240110_120000_deck_suite_summary_R8QGUQRCV.json \
  --tag R8QGUQRCV \
  --sort-by overall \
  --format csv \
  --output-dir data/evaluations \
  --verbose

# Output:
#   data/evaluations/20240110_130000_deck_evaluations_R8QGUQRCV.csv
#   (Ranked list of all 18 decks with scores)

# 3. Find top 5 aggressive decks
./bin/cr-api deck evaluate-batch \
  --from-suite data/my-decks/20240110_120000_deck_suite_summary_R8QGUQRCV.json \
  --filter-archetype --archetype bridge \
  --sort-by attack \
  --top-only --top-n 5
```

### Deck Comparison and Analysis Reports

Compare multiple decks side-by-side with detailed analysis and generate comprehensive reports:

```bash
# Compare decks directly (max 5 decks)
./bin/cr-api deck compare \
  --decks "Knight-Archers-Fireball-Musketeer-Hog Rider-Ice Spirit-Cannon-Log" \
  --decks "Giant-Witch-Skeleton Army-Musketeer-Fireball-Zap-Ice Golem-Archers" \
  --names "Hog Cycle" --names "Giant Beatdown" \
  --format table

# Compare from evaluation results with auto-selection
./bin/cr-api deck compare \
  --from-evaluations data/evaluations/20240110_deck_evaluations_TAG.json \
  --auto-select-top 5 \
  --format markdown \
  --report-output data/reports/comparison_report.md

# Generate JSON comparison for programmatic analysis
./bin/cr-api deck compare \
  --from-evaluations data/evaluations/20240110_deck_evaluations_TAG.json \
  --auto-select-top 3 \
  --format json \
  --output data/reports/comparison.json

# Detailed comparison with win rate predictions
./bin/cr-api deck compare \
  --decks "Deck1" --decks "Deck2" --decks "Deck3" \
  --format table \
  --verbose \
  --winrate
```

**compare Flags:**
- `--decks <deck>` - Deck string (format: "Card1-Card2-...-Card8"), can specify multiple times
- `--names <name>` - Custom name for each deck (optional, corresponds to --decks order)
- `--from-evaluations <file>` - Load from evaluation batch results JSON (alternative to --decks)
- `--auto-select-top <n>` - Auto-select top N decks by overall score (requires --from-evaluations)
- `--format <format>` - Output format: table, json, csv, markdown/md (default: table)
- `--output <file>` - Output file path (default: stdout)
- `--report-output <file>` - Generate comprehensive markdown report to file
- `--verbose` - Show detailed comparison with strengths/weaknesses per deck
- `--winrate` - Show predicted win rate comparison

**Deck Format:** `"Card1-Card2-Card3-Card4-Card5-Card6-Card7-Card8"` (8 cards, hyphen-separated)

**Comparison Outputs:**

*Table Format (default):*
```
╔═══════════════════════════════════════════════════════════════╗
║                      DECK COMPARISON                          ║
╠═══════════════════════════════════════════════════════════════╣

OVERVIEW
┌─────────────┬─────────┬────────────┬──────────────┐
│ Deck        │ Score   │ Avg Elixir │ Archetype    │
├─────────────┼─────────┼────────────┼──────────────┤
│ Hog Cycle   │ 8.5 ⭐  │ 3.1        │ Cycle        │
│ Giant Beat  │ 8.2 ⭐  │ 4.0        │ Beatdown     │
└─────────────┴─────────┴────────────┴──────────────┘

CATEGORY SCORES
┌──────────────┬───────────┬─────────────┐
│ Category     │ Hog Cycle │ Giant Beat  │
├──────────────┼───────────┼─────────────┤
│ Attack       │ 8.2 ⭐⭐  │ 8.7 ⭐⭐⭐   │
│ Defense      │ 8.8 ⭐⭐⭐ │ 7.9 ⭐⭐     │
│ Synergy      │ 7.9 ⭐⭐  │ 8.5 ⭐⭐⭐   │
│ Versatility  │ 8.4 ⭐⭐⭐ │ 7.7 ⭐⭐     │
│ F2P Friendly │ 8.0 ⭐⭐  │ 6.5 ⭐⭐     │
└──────────────┴───────────┴─────────────┘

BEST IN CATEGORY
🥇 Attack: Giant Beat (8.7)
🥇 Defense: Hog Cycle (8.8)
🥇 Synergy: Giant Beat (8.5)
```

*Markdown Report Format:*
- Executive summary with recommended deck
- Overall rankings (🥇🥈🥉)
- Detailed score comparison table
- Category champions
- Complete deck compositions with card roles
- Per-deck analysis (strengths, areas for improvement, recommendations)

**Example Workflow with Reports:**
```bash
# 1. Build and evaluate deck suite (see previous sections)
./bin/cr-api deck build-suite --tag TAG --strategies all --variations 2
./bin/cr-api deck evaluate-batch --from-suite data/decks/suite_TAG.json --tag TAG

# 2. Compare top 5 performers with comprehensive report
./bin/cr-api deck compare \
  --from-evaluations data/evaluations/evaluations_TAG.json \
  --auto-select-top 5 \
  --format markdown \
  --report-output data/reports/top5_comparison.md \
  --verbose

# Output: data/reports/top5_comparison.md
# - Executive summary with deck recommendation
# - Rankings and category winners
# - Detailed analysis for each deck
# - Usage recommendations
```

### Unified Deck Analysis Suite

The `analyze-suite` command combines building, evaluation, and comparison into a single unified workflow:

```bash
# Full analysis suite: build all strategies, evaluate all decks, compare top 5
./bin/cr-api deck analyze-suite --tag <TAG> --strategies all --variations 2

# Custom workflow: specific strategies, 3 variations, compare top 3
./bin/cr-api deck analyze-suite \
  --tag <TAG> \
  --strategies balanced,aggro,cycle \
  --variations 3 \
  --top-n 3 \
  --output-dir data/analysis

# Offline mode with existing player data
./bin/cr-api deck analyze-suite \
  --tag <TAG> \
  --from-analysis \
  --strategies all \
  --variations 1 \
  --top-n 5

# With deck constraints
./bin/cr-api deck analyze-suite \
  --tag <TAG> \
  --strategies all \
  --variations 2 \
  --min-elixir 2.8 \
  --max-elixir 4.2 \
  --include-cards "Hog Rider,Log" \
  --exclude-cards "Elite Barbarians,Royal Giant" \
  --verbose
```

**analyze-suite Flags:**
- `--tag <TAG>` - Player tag (required)
- `--strategies <list>` - Comma-separated strategies or 'all' (default: all)
  - Available: balanced, aggro, control, cycle, splash, spell, all
- `--variations <n>` - Number of variations per strategy (default: 1)
- `--output-dir <dir>` - Base output directory for all results (default: data/analysis)
- `--top-n <n>` - Number of top decks to compare in final report (default: 5, max: 5)
- `--from-analysis` - Use offline mode with pre-analyzed player data
- `--min-elixir <float>` - Minimum average elixir for decks (default: 2.5)
- `--max-elixir <float>` - Maximum average elixir for decks (default: 4.5)
- `--include-cards <cards>` - Cards that must be included (comma-separated)
- `--exclude-cards <cards>` - Cards that must be excluded (comma-separated)
- `--verbose` - Show detailed progress information

**Workflow Phases:**

The command executes three phases automatically:

1. **Phase 1: Building Deck Variations**
   - Generates decks for each strategy × variations
   - Saves individual deck JSON files
   - Creates suite summary JSON

2. **Phase 2: Evaluating All Decks**
   - Evaluates each deck with player context
   - Generates comprehensive scores for all categories
   - Saves evaluation results JSON

3. **Phase 3: Comparing Top Performers**
   - Selects top N decks (sorted by overall score)
   - Generates detailed comparison markdown report
   - Includes recommendations and analysis

**Output Directory Structure:**
```
data/analysis/
├── decks/
│   ├── 20240110_120000_deck_balanced_var1_TAG.json
│   ├── 20240110_120000_deck_aggro_var1_TAG.json
│   ├── ... (individual deck files)
│   └── 20240110_120000_deck_suite_summary_TAG.json
├── evaluations/
│   └── 20240110_120100_deck_evaluations_TAG.json
└── reports/
    └── 20240110_120200_deck_analysis_report_TAG.md
```

**Example Complete Workflow:**
```bash
# Run comprehensive analysis with all strategies, 3 variations each
./bin/cr-api deck analyze-suite --tag R8QGUQRCV --strategies all --variations 3 --verbose

# Console Output (abbreviated):
# 🏗️  Phase 1/3: Building deck variations...
# ✓ Built 18 decks (6 strategies × 3 variations)
# 📁 Saved: data/analysis/decks/20240110_deck_suite_summary_R8QGUQRCV.json
#
# 📊 Phase 2/3: Evaluating all decks...
# ✓ Evaluated 18 decks in 12.3s
# 📁 Saved: data/analysis/evaluations/20240110_deck_evaluations_R8QGUQRCV.json
#
# 🔍 Phase 3/3: Comparing top 5 performers...
# ✓ Generated comparison report
# 📁 Saved: data/analysis/reports/20240110_deck_analysis_report_R8QGUQRCV.md
#
# ✅ Analysis complete!
#    Total decks: 18
#    Top performer: Cycle variation 2 (Score: 8.9)
#    Full report: data/analysis/reports/20240110_deck_analysis_report_R8QGUQRCV.md

# Review the generated markdown report
cat data/analysis/reports/20240110_120200_deck_analysis_report_R8QGUQRCV.md
```

**Best Practices for Deck Analysis Suite:**
1. Start with `--strategies all --variations 2` for comprehensive baseline analysis
2. Use `--verbose` to monitor progress for large batch operations
3. Use `--top-n 3` for focused comparison when you have many variations
4. Apply `--min-elixir` and `--max-elixir` to constrain analysis to your playstyle
5. Use `--include-cards` to ensure specific cards (like your highest-level cards) are always included
6. Review the markdown report for detailed recommendations before committing to a deck
7. Re-run with `--from-analysis` for instant offline experimentation with different parameters

See [DECK_ANALYSIS_SUITE.md](DECK_ANALYSIS_SUITE.md) for comprehensive guide and advanced workflows.

### Deck Optimization Commands

#### Deck Fuzzing (Monte Carlo and Genetic Algorithm)

Generate and evaluate random decks to discover optimal combinations:

**Monte Carlo Fuzzing** (Random Sampling):
```bash
# Generate 1000 random decks, show top 10
./bin/cr-api deck fuzz --tag <TAG>

# With constraints
./bin/cr-api deck fuzz --tag <TAG> \
  --count 5000 \
  --include-cards "Royal Giant" \
  --max-elixir 3.5 \
  --top 20
```

**Genetic Algorithm** (Evolutionary Optimization):
```bash
# Optimize decks using genetic algorithm
./bin/cr-api deck fuzz --mode genetic --tag <TAG>

# Quick optimization
./bin/cr-api deck fuzz --mode genetic --tag <TAG> \
  --ga-population 50 \
  --ga-generations 30

# Thorough search
./bin/cr-api deck fuzz --mode genetic --tag <TAG> \
  --ga-population 200 \
  --ga-generations 300 \
  --ga-elite-count 20

# Island model for diversity
./bin/cr-api deck fuzz --mode genetic --tag <TAG> \
  --ga-island-model \
  --ga-island-count 4 \
  --ga-migration-interval 10
```

**General Fuzz Flags:**
- `--tag <TAG>` - Player tag (required)
- `--mode <mode>` - Fuzzing mode: `random` (default) or `genetic`
- `--count <n>` - Number of random decks (Monte Carlo only, default: 1000)
- `--top <n>` - Number of top decks to display (default: 10)
- `--sort-by <criteria>` - Sort by: overall, attack, defense, synergy, versatility, elixir
- `--format <fmt>` - Output format: summary, json, csv, detailed
- `--output-dir <dir>` - Directory to save results
- `--verbose` - Show detailed progress

**Monte Carlo Flags:**
- `--workers <n>` - Parallel workers (default: 1)
- `--include-cards <cards>` - Cards that must be in every deck
- `--exclude-cards <cards>` - Cards to exclude from all decks
- `--min-elixir <float>` - Minimum average elixir
- `--max-elixir <float>` - Maximum average elixir
- `--seed <n>` - Random seed (0 = random)

**Genetic Algorithm Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ga-population` | int | 100 | Population size per generation |
| `--ga-generations` | int | 100 | Maximum number of generations |
| `--ga-mutation-rate` | float | 0.2 | Probability of mutation (0.0-1.0) |
| `--ga-crossover-rate` | float | 0.7 | Probability of crossover (0.0-1.0) |
| `--ga-mutation-intensity` | float | 0.3 | Mutation intensity: cards changed (0.0-1.0) |
| `--ga-elite-count` | int | 10 | Best decks to preserve per generation |
| `--ga-tournament-size` | int | 5 | Tournament selection size |
| `--ga-parallel-eval` | bool | false | Enable parallel fitness evaluation |
| `--ga-convergence-generations` | int | 0 | Stop if no improvement for N generations (0=off) |
| `--ga-target-fitness` | float | 0.0 | Stop when fitness reaches this value (0=off) |
| `--ga-island-model` | bool | false | Enable island model (parallel populations) |
| `--ga-island-count` | int | 4 | Number of islands |
| `--ga-migration-interval` | int | 10 | Generations between migrations |
| `--ga-migration-size` | int | 5 | Decks to migrate per interval |

See [DECK_FUZZING.md](DECK_FUZZING.md) for Monte Carlo fuzzing details and [GENETIC_FUZZING.md](GENETIC_FUZZING.md) for genetic algorithm documentation.

#### Deck Mulligan Guide (Opening Hand Strategy)

Generate mulligan guides that recommend opening hand strategies for your deck:

```bash
# Generate mulligan guide for a deck
./bin/cr-api deck mulligan --cards Knight Archers Fireball Musketeer "Hog Rider" "Ice Spirit" Cannon Log

# Save to file
./bin/cr-api deck mulligan --cards Knight Archers Fireball Musketeer "Hog Rider" "Ice Spirit" Cannon Log --save

# Output as JSON
./bin/cr-api deck mulligan --cards Knight Archers Fireball Musketeer "Hog Rider" "Ice Spirit" Cannon Log --json

# With custom deck name
./bin/cr-api deck mulligan --cards Knight Archers Fireball Musketeer "Hog Rider" "Ice Spirit" Cannon Log --deck-name "Hog Cycle v3"
```

**Mulligan Flags:**
- `--cards <card>` - 8 card names (required, can specify multiple times)
- `--deck-name <name>` - Custom name for the deck (optional)
- `--save` - Save guide to file
- `--json` - Output in JSON format instead of human-readable table

#### Deck War Builder

Build 4-deck war sets with zero card overlap for clan war:

```bash
# Build war deck set for your collection
./bin/cr-api deck war --tag <TAG>

# Build 5-deck war set instead of default 4
./bin/cr-api deck war --tag <TAG> --deck-count 5

# With custom evolution settings
./bin/cr-api deck war --tag <TAG> --unlocked-evolutions "Archers,Knight,Musketeer" --evolution-slots 3

# Disable combat stats weighting
./bin/cr-api deck war --tag <TAG> --disable-combat-stats
```

**War Deck Flags:**
- `--tag <TAG>` - Player tag (required)
- `--deck-count <n>` - Number of decks to build (default: 4)
- `--unlocked-evolutions <cards>` - Comma-separated evolution cards
- `--evolution-slots <n>` - Number of evolution slots available (default: 2)
- `--combat-stats-weight <float>` - Weight for combat stats (0.0-1.0, default: 0.25)
- `--disable-combat-stats` - Use traditional scoring only

#### Deck Budget Finder

Find budget-optimized decks that maximize impact with minimal upgrade investment:

```bash
# Find decks requiring minimal upgrades
./bin/cr-api deck budget --tag <TAG> --sort-by roi

# Find quick-win decks (1-2 upgrades away from viability)
./bin/cr-api deck budget --tag <TAG> --quick-wins --ready-only

# Find decks requiring specific gold/cards
./bin/cr-api deck budget --tag <TAG> --max-gold 10000 --max-cards 50

# With deck constraints
./bin/cr-api deck budget --tag <TAG> \
  --include-cards "Hog Rider,Log" \
  --max-variations 5 \
  --target-level 12.0

# Generate deck variations for different budgets
./bin/cr-api deck budget --tag <TAG> \
  --include-variations \
  --max-variations 3 \
  --sort-by cost_efficiency
```

**Budget Finder Flags:**
- `--tag <TAG>` - Player tag (required)
- `--max-cards <n>` - Maximum cards needed for upgrades (0 = unlimited)
- `--max-gold <n>` - Maximum gold needed for upgrades (0 = unlimited)
- `--target-level <float>` - Target average card level (default: 12.0)
- `--sort-by <criteria>` - Sort by: roi, cost_efficiency, total_cards, total_gold, current_score, projected_score
- `--top-n <n>` - Number of top decks to display (default: 10)
- `--include-variations` - Generate and analyze deck variations
- `--max-variations <n>` - Maximum variations per base deck (default: 5)
- `--quick-wins` - Show only quick-win decks (1-2 upgrades)
- `--ready-only` - Show only already-competitive decks
- `--json` - Output as JSON
- `--save` - Save results to file

#### Deck Possible Count

Calculate realistic deck combination possibilities from your card collection:

```bash
# Calculate possible deck combinations
./bin/cr-api deck possible-count --tag <TAG>

# Show detailed breakdown by role and archetype
./bin/cr-api deck possible-count --tag <TAG> --verbose

# Output as JSON
./bin/cr-api deck possible-count --tag <TAG> --format json

# Export to file
./bin/cr-api deck possible-count --tag <TAG> --format csv --output stats.csv
```

**Possible Count Flags:**
- `--tag <TAG>` - Player tag (required)
- `--format <format>` - Output format: human, json, csv (default: human)
- `--verbose` - Show detailed breakdown by role/archetype
- `--output <file>` - Save output to file

### Archetype Analysis

#### Dynamic Archetype Detection

Detect deck archetypes dynamically with confidence scoring:

```bash
# Analyze archetypes in player's collection
./bin/cr-api archetype detect --tag <TAG>

# Show detailed strategy recommendations per archetype
./bin/cr-api archetype detect --tag <TAG> --show-strategies --show-upgrades

# Verbose output with full archetype details
./bin/cr-api archetype detect --tag <TAG> --verbose

# Save analysis results
./bin/cr-api archetype detect --tag <TAG> --save
```

**Archetype Detect Flags:**
- `--tag <TAG>` - Player tag (required)
- `--show-strategies` - Display strategy recommendations
- `--show-upgrades` - Display upgrade recommendations per archetype
- `--verbose` - Show detailed information
- `--save` - Save results to file

#### Archetype Variety Analysis

Analyze variety and synergy within supported archetypes:

```bash
# Analyze archetype variety in collection
./bin/cr-api archetype variety --tag <TAG>

# Show specific archetype details
./bin/cr-api archetype variety --tag <TAG> --archetype cycle

# Full analysis with recommendations
./bin/cr-api archetype variety --tag <TAG> --verbose --show-recommendations
```

**Archetype Variety Flags:**
- `--tag <TAG>` - Player tag (required)
- `--archetype <type>` - Analyze specific archetype
- `--verbose` - Show detailed analysis
- `--show-recommendations` - Display building recommendations

### Evolution System

#### Recommend Evolution Paths

Get personalized evolution upgrade recommendations:

```bash
# Recommend top 5 evolution upgrades
./bin/cr-api evolution recommend --tag <TAG> --top 5

# Show all evolution options with analysis
./bin/cr-api evolution recommend --tag <TAG> --verbose

# Check specific evolution card
./bin/cr-api evolution shards list --tag <TAG>

# Update evolution shard inventory
./bin/cr-api evolution shards set --tag <TAG> --card "Archers" --shards 50
```

**Evolution Flags:**
- `--tag <TAG>` - Player tag (required)
- `--top <n>` - Number of recommendations (default: 5)
- `--verbose` - Show detailed evolution analysis

**Shards Subcommands:**
- `shards list` - Show current evolution shard inventory
- `shards set` - Update shards for a specific card

### Event Tracking

```bash
./bin/cr-api events scan --tag <TAG>
./bin/cr-api playstyle --tag <TAG> [--recommend-decks] [--save]
```

### What-If Analysis

Simulate the impact of upgrading specific cards on deck composition and viability.

```bash
# Basic upgrade simulation
./bin/cr-api what-if --tag <TAG> --upgrade "Archers:15" --upgrade "Knight:15" --show-decks

# Offline mode with existing analysis file
./bin/cr-api what-if --tag <TAG> --from-analysis data/analysis/<file>.json \
  --upgrade "Archers:9:15" --save --json

# Specify different deck building strategy
./bin/cr-api what-if --tag <TAG> --upgrade "Goblin Barrel:14" --strategy aggro --show-decks
```

**Upgrade Format:**
- `CardName:ToLevel` - Upgrades from current level to specified level
- `CardName:FromLevel:ToLevel` - Explicit from→to upgrade path

**What-If Flags:**
- `--from-analysis <file>` - Use cached analysis (offline mode, no API call)
- `--show-decks` - Display full deck compositions before/after
- `--save` - Save scenario to `data/whatif/`
- `--json` - Output in JSON format
- `--strategy <name>` - Deck building strategy

**What-If Output:**
- Upgrade costs (gold per card)
- Deck score delta
- Viability improvement percentage
- New/removed cards in recommended deck
- Recommendation (highly recommended / recommended / minor improvement / not recommended)

**Example Output:**
```
============================================================================
                        WHAT-IF ANALYSIS
============================================================================

Scenario: Upgrade 2 cards: Archers, Knight
What-if analysis for Player (#TAG)

Upgrades Simulated
-------------------
Card     From  To  Gold
----     ----  --  ----
Archers  10    15  500
Knight   9     15  600

Total Gold Cost: 1100

Impact Analysis
---------------
Deck Score Delta:     +0.888000
Viability Change:     +12.0%

Recommendation
-------------
Highly recommended! These upgrades (1100 gold) significantly improve your deck viability by 12.0%.
```

### Evolution Management

```bash
./bin/cr-api evolutions recommend --tag <TAG> [--top 5] [--verbose]
```

See [EVOLUTION.md](EVOLUTION.md) for evolution mechanics and configuration.

### Testing Commands

```bash
cd go && go test ./...              # Run all tests
cd go && go test ./pkg/deck/... -v  # Test specific package with verbose output
cd go && go test -tags=integration ./...  # Full integration tests (requires API token)
```

See [TESTING.md](TESTING.md) for complete testing documentation.

## Configuration Options

**Required**: `CLASH_ROYALE_API_TOKEN` in `.env`

**Optional**:
```env
DEFAULT_PLAYER_TAG=#TAG          # Allows running tasks without arguments
DATA_DIR=./data                  # Data storage location
REQUEST_DELAY=1                  # Seconds between API requests
MAX_RETRIES=3                    # API retry attempts
CSV_DIR=./data/csv               # CSV export directory
COMBAT_STATS_WEIGHT=0.25         # Combat stats weight for deck building (0.0-1.0)
UNLOCKED_EVOLUTIONS="Archers,Knight,Musketeer"  # Evolution tracking
```

**Configuration Priority:**
1. CLI arguments (highest)
2. Environment variables
3. Default values (lowest)

## Deck Building Options

### Combat Stats Integration

The deck builder blends traditional scoring (level, rarity, cost) with combat stats (DPS/elixir, HP/elixir):

`finalScore = (traditional × (1-weight)) + (combat × weight)`

```bash
./bin/cr-api deck build --tag <TAG>                              # 25% weight (default)
./bin/cr-api deck build --tag <TAG> --combat-stats-weight 0.6   # 60% weight
./bin/cr-api deck build --tag <TAG> --disable-combat-stats       # 0% weight (traditional only)
```

**Weight guidance:**
- **0.5-0.8**: Prioritize statistically strong cards (theory-crafting)
- **0.25** (default): Balanced, recommended for most
- **0.0-0.2**: Focus on highest-level cards (ladder pushing)

### Synergy Scoring

Enable optional synergy system that considers card interactions and combos.

```bash
./bin/cr-api deck build --tag <TAG> --enable-synergy                      # 15% weight (default)
./bin/cr-api deck build --tag <TAG> --enable-synergy --synergy-weight 0.25  # 25% weight
```

**Weight guidance:**
- **0.10-0.15** (default): Subtle synergy influence, card levels still matter most
- **0.20-0.30**: Moderate synergy emphasis, good balance for most players
- **0.40-0.50**: Strong synergy focus, prioritizes combos over individual card strength

**Synergy categories:**
- **Tank + Support**: Giant/Witch, Golem/Night Witch, Lava Hound/Balloon
- **Spell Combos**: Hog/Fireball, Tornado/Executioner, Freeze/Balloon
- **Bait**: Goblin Barrel/Princess, Log bait variations
- **Win Conditions**: Hog/Ice Golem, X-Bow/Tesla, Miner/Poison
- **Defensive**: Cannon/Ice Spirit, Inferno Tower/Tornado, Tesla defense
- **Cycle**: Ice Spirit/Skeletons, fast rotation combos
- **Bridge Spam**: PEKKA/Battle Ram, Bandit/Royal Ghost

See [DECK_BUILDER.md](DECK_BUILDER.md) for algorithm details and Go API examples.

### Deck Discovery & Leaderboard

Discover optimal deck combinations through systematic exploration and persistent storage.

```bash
# Start a discovery session
cr-api deck discover start --tag <TAG> --strategy smart-sample --sample-size 5000

# Resume from checkpoint
cr-api deck discover resume --tag <TAG>

# Check discovery status
cr-api deck discover status --tag <TAG>

# View detailed statistics
cr-api deck discover stats --tag <TAG>

# Stop background discovery
cr-api deck discover stop --tag <TAG>
```

**Discovery Strategies:**
- `smart-sample` (default): Intelligently prioritizes high-level cards and synergies
- `random-sample`: Unbiased random exploration
- `archetype-focused`: Deep dive into specific archetype
- `exhaustive`: Evaluate all combinations (warning: very slow for large collections)

**Discovery Flags:**
- `--strategy <type>` - Sampling strategy (default: smart-sample)
- `--sample-size <n>` - Number of decks to generate (default: 1000)
- `--limit <n>` - Maximum decks to evaluate (0 = unlimited)
- `--verbose`, `-v` - Show detailed progress
- `--background` - Run as daemon process

```bash
# View leaderboard
cr-api deck leaderboard show --tag <TAG> --top 20

# Filter leaderboard
cr-api deck leaderboard filter --tag <TAG> --archetype cycle --min-elixir 2.5 --max-elixir 3.0

# View statistics
cr-api deck leaderboard stats --tag <TAG> --archetypes

# View storage statistics (DB size + archetype distribution)
cr-api deck storage stats --tag <TAG>

# Export leaderboard
cr-api deck leaderboard export --tag <TAG> --format csv --output decks.csv
```

**Leaderboard Flags:**
- `--top <n>` - Number of results (default: 10)
- `--format <fmt>` - Output: summary, detailed, json, csv
- `--archetype <type>` - Filter by archetype
- `--min-score`, `--max-score` - Score range filter (0-10)
- `--min-elixir`, `--max-elixir` - Elixir range filter
- `--sort-by <field>` - Sort field: overall_score, attack_score, defense_score, etc.
- `--order <order>` - Sort order: asc, desc (default: desc)
- `--require-all <cards>` - Decks must contain ALL cards
- `--require-any <cards>` - Decks must contain ANY cards
- `--exclude <cards>` - Exclude decks with ANY cards

See [DECK_DISCOVERY.md](DECK_DISCOVERY.md) for complete documentation.
