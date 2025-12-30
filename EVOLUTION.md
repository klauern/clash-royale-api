# Evolution System Documentation

This document describes the Clash Royale evolution system as implemented in this API client and analysis tool.

## Overview

Clash Royale evolutions are card enhancements that unlock at King Level 7 and provide significant power boosts. This tool tracks evolution progress, recommends optimal evolution paths, and integrates evolution data into deck building and upgrade priority calculations.

## Key Concepts

### Evolution Levels

Cards have two evolution level values:
- **EvolutionLevel (0-3)**: Current evolution stage
- **MaxEvolutionLevel (0-3)**: Maximum possible evolutions for this card

### Unlocked Evolutions

The Clash Royale API doesn't provide evolution unlock status. Configure manually:

```bash
# Environment variable
export UNLOCKED_EVOLUTIONS="Archers,Knight,Valkyrie"

# CLI flag
./bin/cr-api deck build --tag PLAYER_TAG --unlocked-evolutions "Archers,Knight"
```

## Evolution Role Overrides

Some cards change their strategic role when evolved:

| Card | Base Role | Evolved Role |
|------|-----------|--------------|
| Valkyrie | Cycle | Support |
| Knight | Cycle | Support |
| Royal Giant | Support | Win Condition |
| Barbarian | Cycle | Support |

Cards with role overrides receive bonus scoring when their evolution is unlocked.

## Evolution Slot Priority

When a deck contains 3+ evolved cards, only a limited number (default: 2) can use evolution slots simultaneously. Slots are assigned by role priority:

1. Win Conditions
2. Buildings
3. Big Spells
4. Support
5. Small Spells
6. Cycle

```bash
# Configure slot limit (default: 2)
./bin/cr-api deck build --tag PLAYER_TAG --evolution-slots 3
```

## Evolution Recommendations

The `evolutions recommend` command analyzes your card collection to recommend optimal evolution upgrades.

```bash
# Basic recommendations
./bin/cr-api evolutions recommend --tag PLAYER_TAG

# Top 5 recommendations with detailed reasons
./bin/cr-api evolutions recommend --tag PLAYER_TAG --top 5 --verbose
```

Recommendations consider card level, shard progress, role priority, and multi-evolution potential.

### Shard Management

```bash
# List current shard inventory
./bin/cr-api evolutions shards list

# Set shard count for a card
./bin/cr-api evolutions shards set --card "Archers" --count 50
```

Shard data is stored in `data/evolution_shards.json`.

## CSV Export

Evolution data is included in CSV exports:

- **Player Cards**: `Evolution Level`, `Max Evolution Level`
- **Event Decks**: Cards display evolution level (e.g., "Hog Rider Lv.11 Evo.2")

## Deck Display

When displaying decks, evolution slots are marked with the `⚡` badge:

```
Deck 1 - Hog Cycle
#  Card           Level           Elixir  Role
1  Hog Rider      14/14           4       Win Condition
2  Valkyrie       14/14 (⚡)      4       Support
3  Ice Spirit     14/14           1       Cycle

Notes:
• Evolution slots: Valkyrie, Knight
```
