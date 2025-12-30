# Deck Builder Documentation

The Clash Royale API includes an intelligent deck building system that creates optimized 8-card decks based on a player's card collection and levels.

## Overview

The deck builder analyzes player data to recommend balanced decks with proper card roles and synergies. It considers card levels, rarity, elixir cost, and strategic roles.

## Card Roles

Each card is assigned a strategic role:

| Role | Description | Examples |
|------|-------------|----------|
| **Win Condition** | Primary tower-damaging cards | Hog Rider, Royal Giant, Giant, P.E.K.K.A |
| **Building** | Defensive structures | Cannon, Inferno Tower, Tombstone |
| **Big Spell** | High-elixir spells | Fireball, Poison, Lightning, Rocket |
| **Small Spell** | Low-elixir utility spells | Zap, Arrows, Log, Giant Snowball |
| **Support** | Mid-cost versatile troops | Archers, Musketeer, Wizard, Valkyrie |
| **Cycle** | Low-cost cycling cards | Skeletons, Ice Spirit, Knight, Bats |

## Usage

```bash
# Build a deck from player's latest analysis
./bin/cr-api deck build --tag '#PLAYERTAG'

# Build with custom evolution settings
./bin/cr-api deck build --tag '#PLAYERTAG' --unlocked-evolutions "Archers,Knight"

# Build from offline analysis file
./bin/cr-api deck build --from-analysis 'path/to/analysis.json'
```

## Deck Building Strategy

### Role Distribution

The algorithm aims for optimal balance:

1. 1 Win Condition
2. 1 Building
3. 1 Big Spell
4. 1 Small Spell
5. 2 Support Cards
6. 2 Cycle Cards

### Strategic Notes

The deck builder provides contextual advice:

- **High Elixir** (> 3.8): "Play patiently and build pushes"
- **Low Elixir** (< 2.8): "Pressure often and out-cycle counters"
- **No Building**: "Play troops high to kite"
- **No Spell**: "Beware of swarm matchups"

## Best Practices

1. Rebuild decks regularly as card levels change
2. Test new decks thoroughly before committing
3. Higher-level cards receive higher scores