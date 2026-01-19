# Deck Analysis Skill

## Invocation

Use this skill when the user requests deck building, deck analysis, or deck recommendations for Clash Royale.

**Trigger phrases:**
- "build me a deck"
- "analyze decks for player X"
- "recommend decks"
- "best deck for my cards"
- "deck suggestions"
- "1v1 ladder deck"

## What This Skill Does

Generates comprehensive deck recommendations using the `cr-api` CLI tool by:
1. Building multiple deck variations across different strategies
2. Evaluating decks based on attack, defense, synergy, versatility, and F2P friendliness
3. Providing targeted recommendations based on player's card collection and levels
4. Optionally incorporating unlocked evolutions

## Parameters

### Required
- `--tag` or `-p`: Player tag (with or without #)

### Optional Strategy Selection
- `--strategies`: Comma-separated strategies or "all" (default: "all")
  - Available: balanced, aggro, control, cycle, splash, spell
- `--variations`: Number of variations per strategy (default: 1)

### Deck Constraints
- `--min-elixir`: Minimum average elixir (default: 2.5)
- `--max-elixir`: Maximum average elixir (default: 4.5)
- `--include-cards`: Cards that must be in the deck
- `--exclude-cards`: Cards to exclude from the deck

### Evolution Support
- `--unlocked-evolutions`: Comma-separated list of cards with unlocked evolutions
- `--evolution-slots`: Number of evolution slots (default: 2)

### Analysis Options
- `--enable-synergy`: Enable synergy-based scoring
- `--top-n`: Number of top decks to compare (default: 5)
- `--verbose`: Show detailed progress

## Workflow

### 1. Build the binary
```bash
task build
```

### 2. Verify player data
```bash
./bin/cr-api player get --tag "<PLAYER_TAG>"
```

### 3. Run comprehensive analysis
```bash
./bin/cr-api deck analyze-suite \
  --tag "<PLAYER_TAG>" \
  --strategies all \
  --variations 3 \
  --verbose
```

### 4. Build targeted decks (if specific cards requested)
```bash
./bin/cr-api deck build \
  --tag "<PLAYER_TAG>" \
  --strategy <STRATEGY> \
  --include-cards "<CARD_NAME>" \
  --enable-synergy
```

## Output

The skill generates:

### Files
- **Deck suite summary**: `data/analysis/decks/<timestamp>_deck_suite_summary_<TAG>.json`
- **Evaluations**: `data/analysis/evaluations/<timestamp>_deck_evaluations_<TAG>.json`
- **Analysis report**: `data/analysis/reports/<timestamp>_deck_analysis_report_<TAG>.md`

### Console Output
- Top 3 ranked decks with overall scores
- Detailed deck composition (8 cards per deck)
- Strategic notes and archetype identification
- Upgrade recommendations with gold costs

## Response Format

After running the analysis, provide:

1. **Player Profile Summary**
   - Trophy count, arena, win rate
   - Card collection size
   - Unlocked evolutions

2. **Top Deck Recommendations**
   - Top 3-5 decks with scores
   - Card composition and average elixir
   - Strengths and weaknesses
   - When to use each deck

3. **Targeted Recommendations** (if evolutions specified)
   - Decks incorporating unlocked evolutions
   - Strategy-specific builds (cycle, control, etc.)

4. **Upgrade Priority**
   - Top 5 cards to upgrade for each recommended deck
   - Gold costs and expected score improvements

5. **File Locations**
   - Paths to generated reports and data files

## Examples

### Basic Analysis
```bash
./bin/cr-api deck analyze-suite --tag "#R8QGUQRCV"
```

### With Evolution Focus
```bash
./bin/cr-api deck analyze-suite \
  --tag "#R8QGUQRCV" \
  --strategies all \
  --variations 3 \
  --verbose

./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy cycle \
  --include-cards "Archers" \
  --enable-synergy
```

### Constrained Analysis
```bash
./bin/cr-api deck analyze-suite \
  --tag "#R8QGUQRCV" \
  --strategies "cycle,control,aggro" \
  --max-elixir 3.5 \
  --variations 2
```

## Common Strategies

- **balanced**: General-purpose ladder climbing
- **aggro**: Fast, aggressive wins with high elixir efficiency
- **control**: Defensive play with late-game win conditions
- **cycle**: Ultra-fast rotation, chip damage strategy
- **splash**: Anti-swarm with area damage focus
- **spell**: Spell-cycling win conditions

## Notes

- Analysis generates 6-18+ decks depending on variation settings
- All decks are optimized for player's current card levels
- F2P friendliness considers upgrade costs and rarity distribution
- Synergy scoring evaluates card interactions and combos
- Evolution slots are automatically allocated to specified cards
