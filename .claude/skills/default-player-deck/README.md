# Default Player Deck Skill

Quick deck building for your primary Clash Royale account.

## Overview

This skill builds optimized decks using the `DEFAULT_PLAYER_TAG` from your `.env` file, eliminating the need to specify a player tag every time you want a new deck.

## Setup

1. Ensure your `.env` file has your player tag:
   ```bash
   DEFAULT_PLAYER_TAG="#YOUR_PLAYER_TAG"
   ```

2. Build the CLI binary:
   ```bash
   task build
   ```

## Usage

Simply ask for a deck:

```
"Build me a deck"
"Make me a cycle deck"
"Create a control deck for my account"
"Build me a deck with Hog Rider"
```

## Available Strategies

| Strategy | Description | Avg Elixir |
|----------|-------------|------------|
| balanced | General-purpose ladder deck | 3.0-4.0 |
| aggro | Fast, aggressive pressure | 2.5-3.5 |
| control | Defensive, counter-push | 3.5-4.5 |
| cycle | Ultra-fast rotation | 2.3-3.0 |
| splash | Anti-swarm area damage | 3.0-4.0 |
| spell | Spell-cycle win condition | 3.5-4.5 |
| synergy-first | Prioritizes card combos | varies |

## Parameters

### Strategy Selection
- `--strategy <name>` - Choose deck archetype

### Deck Constraints
- `--min-elixir <float>` - Minimum average elixir (default: 2.5)
- `--max-elixir <float>` - Maximum average elixir (default: 4.5)
- `--include-cards <list>` - Cards to include (comma-separated)
- `--exclude-cards <list>` - Cards to exclude (comma-separated)

### Options
- `--enable-synergy` - Enable synergy scoring (recommended)
- `--ideal-deck` - Show deck with upgrades applied

## Evolution Support

Evolutions are automatically configured from your `.env` file:

```bash
UNLOCKED_EVOLUTIONS="Archers,Knight,Musketeer"
```

The skill will prioritize using your unlocked evolutions in deck builds.

## Examples

### Quick Balanced Deck
```
"Build me a deck"
```

### Specific Strategy
```
"Build me a cycle deck"
"Make me an aggro deck"
```

### With Card Constraints
```
"Build me a deck with Hog Rider and Fireball"
"Create a control deck without building cards"
```

### With Elixir Constraints
```
"Build me a deck with max elixir 3.5"
"Make me a heavy deck with min elixir 4.0"
```

### Show Ideal Version
```
"Build me a deck and show the ideal version"
```

## Output

The skill provides:

1. **Player Profile** - Your trophies, arena, and collection summary
2. **Deck Composition** - 8 cards with elixir average
3. **Strategy Notes** - How to play the deck
4. **Upgrade Recommendations** - Best cards to upgrade
5. **Ideal Deck Preview** - Deck at recommended levels (if requested)

## Troubleshooting

### "DEFAULT_PLAYER_TAG not set"
Add your player tag to `.env`:
```bash
echo 'DEFAULT_PLAYER_TAG="#YOUR_TAG"' >> .env
```

### "Player not found"
- Verify your tag includes the `#` symbol
- Check your tag in the Clash Royale app
- Ensure the API token is valid

### Need Full Analysis?
For comprehensive multi-deck analysis with comparisons, use the `deck-analysis` skill instead and provide your player tag explicitly.

## Differences from deck-analysis Skill

| Feature | default-player-deck | deck-analysis |
|---------|---------------------|---------------|
| Player tag | From `.env` | Must specify |
| Decks generated | 1 optimized deck | 6-18+ decks |
| Analysis depth | Quick build | Comprehensive |
| Best for | Quick deck needs | Full meta analysis |
| Time | Fast | Thorough |
