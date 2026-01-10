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
