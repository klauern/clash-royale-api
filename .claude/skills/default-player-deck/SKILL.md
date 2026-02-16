# Default Player Deck Skill

## Invocation

Use this skill when the user requests a deck for their default player account without specifying a player tag.

**Trigger phrases:**
- "build me a deck"
- "make me a deck"
- "create a deck for my account"
- "deck build"
- "quick deck"
- "build a cycle deck"
- "make an aggro deck"
- "build me a control deck"
- "create a splash deck"

## What This Skill Does

Builds an optimized deck for the default player configured in `.env` using the `cr-api` CLI tool:

1. Reads `DEFAULT_PLAYER_TAG` from `.env` file
2. Verifies the player exists and fetches their profile
3. Builds a single optimized deck based on the requested strategy
4. Provides upgrade recommendations and ideal deck preview

## Prerequisites

The `.env` file must contain:
```bash
DEFAULT_PLAYER_TAG="#YOUR_PLAYER_TAG"
```

## Parameters

### Optional Strategy Selection
- `--strategy`: Deck strategy (default: "balanced")
  - Available: balanced, aggro, control, cycle, splash, spell, synergy-first

### Deck Constraints
- `--min-elixir`: Minimum average elixir (default: 2.5)
- `--max-elixir`: Maximum average elixir (default: 4.5)
- `--include-cards`: Cards that must be in the deck (comma-separated)
- `--exclude-cards`: Cards to exclude (comma-separated)

### Evolution Support
- Uses `UNLOCKED_EVOLUTIONS` from `.env` automatically
- Evolution slots allocated based on deck strategy

### Analysis Options
- `--enable-synergy`: Enable synergy-based scoring (recommended)
- `--ideal-deck`: Show deck with upgrade recommendations applied

## Workflow

### 1. Verify prerequisites
```bash
# Check .env has DEFAULT_PLAYER_TAG
grep DEFAULT_PLAYER_TAG .env

# Build the binary if needed
task build
```

### 2. Verify player data
```bash
./bin/cr-api player get --tag "<PLAYER_TAG>"
```

### 3. Build optimized deck
```bash
./bin/cr-api deck build \
  --tag "<PLAYER_TAG>" \
  --strategy <STRATEGY> \
  --enable-synergy \
  [additional options]
```

### 4. Show ideal deck (optional)
```bash
./bin/cr-api deck build \
  --tag "<PLAYER_TAG>" \
  --strategy <STRATEGY> \
  --enable-synergy \
  --ideal-deck
```

## Response Format

After building the deck, provide:

### 1. Player Profile Summary
- Name, trophies, arena
- Card collection summary
- Unlocked evolutions from `.env`

### 2. Deck Recommendation
- 8-card deck composition
- Average elixir cost
- Strategy archetype

### 3. Strategy Notes
- Why these cards work together
- Key synergies and combos
- Win condition explanation
- Defensive strategy

### 4. Upgrade Recommendations
- Top 3-5 cards to upgrade
- Gold costs
- Expected impact on deck performance

### 5. Ideal Deck Preview (if requested)
- Deck with all recommended upgrades applied
- Comparison to current levels

## Examples

### Basic Deck Build
```bash
./bin/cr-api deck build --tag "#R8QGUQRCV" --strategy balanced --enable-synergy
```

### Cycle Deck with Constraints
```bash
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy cycle \
  --max-elixir 3.5 \
  --enable-synergy
```

### Include Specific Cards
```bash
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy control \
  --include-cards "Hog Rider,Fireball" \
  --enable-synergy
```

### Show Ideal Deck
```bash
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy aggro \
  --enable-synergy \
  --ideal-deck
```

## Strategy Descriptions

- **balanced**: General-purpose deck with mix of offense and defense
- **aggro**: Fast-paced, high-pressure deck for quick wins
- **control**: Defensive play with strong counter-push potential
- **cycle**: Ultra-fast rotation for constant pressure and chip damage
- **splash**: Area damage focus to counter swarm decks
- **spell**: Spell-cycle win condition with spell damage
- **synergy-first**: Prioritizes card interactions over raw stats

## Error Handling

If `DEFAULT_PLAYER_TAG` is not set:
1. Inform the user the variable is missing
2. Provide instructions to add it to `.env`:
   ```bash
   echo 'DEFAULT_PLAYER_TAG="#YOUR_TAG"' >> .env
   ```
3. Offer to use the full `deck-analysis` skill with explicit tag instead

If player tag is invalid:
1. Show the error from the API
2. Verify the tag format (should include #)
3. Suggest checking the tag in-game

## Notes

- This skill builds a single optimized deck rather than a full analysis suite
- Faster than comprehensive analysis for quick deck building
- Uses same scoring and evaluation as the full deck-analysis skill
- Evolution configuration is automatically read from `.env`
- For comprehensive multi-deck analysis, use the `deck-analysis` skill instead
