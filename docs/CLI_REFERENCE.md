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
