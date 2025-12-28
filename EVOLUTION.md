# Evolution System Documentation

This document describes the Clash Royale evolution system as implemented in this API client and analysis tool.

## Overview

Clash Royale evolutions are card enhancements that unlock at King Level 7 and provide significant power boosts. This tool tracks evolution progress, recommends optimal evolution paths, and integrates evolution data into deck building and upgrade priority calculations.

## Key Concepts

### Evolution Levels

- **EvolutionLevel (0-3)**: Current evolution stage of a card
  - `0`: Not evolved
  - `1`: First evolution unlocked
  - `2`: Second evolution unlocked (multi-evolution cards)
  - `3`: Third evolution unlocked (rare, multi-evolution cards)

- **MaxEvolutionLevel (0-3)**: Maximum possible evolution stages for this card
  - `0`: Card cannot evolve
  - `1`: Standard single-evolution card
  - `2-3`: Multi-evolution card (rare)

### Shards and Progress

- **Evolution Shards**: Currency required to unlock evolutions
- **Shard Requirements**: Vary by card rarity and evolution stage
- **Progress Tracking**: Shard counts are stored locally (not provided by API)

### Unlocked Evolutions

The Clash Royale API doesn't provide evolution unlock status or shard counts. You must configure this manually:

```bash
# Environment variable
export UNLOCKED_EVOLUTIONS="Archers,Knight,Valkyrie"

# CLI flag
./bin/cr-api deck build --tag PLAYER_TAG --unlocked-evolutions "Archers,Knight"
```

## Evolution Role Overrides

Some cards change their strategic role when evolved. The deck builder accounts for this:

| Card | Base Role | Evolved Role | Reason |
|------|-----------|--------------|--------|
| Valkyrie | Cycle | Support | Whirlwind pull makes it a control support card |
| Knight | Cycle | Support | Clone on death increases defensive value |
| Royal Giant | Support | Win Condition | Anti-pushback improves win condition reliability |
| Barbarian | Cycle | Support | 3 spawned barbarians act as support swarm |
| Witch | Support | Support (enhanced) | Faster skeleton spawn enhances support value |
| Golem | Win Condition | Win Condition (enhanced) | Golemites spawn on death strengthens win condition |

**Impact**: Cards with evolution role overrides receive a 10% bonus to their evolution score when the evolution is unlocked.

## Evolution Slot Priority

When a deck contains 3+ evolved cards, only a limited number (default: 2) can use evolution slots simultaneously. Slots are assigned by priority:

1. **Win Conditions** (Hog Rider, Royal Giant, etc.)
2. **Buildings** (Cannon, Inferno Tower, etc.)
3. **Big Spells** (Fireball, Poison, etc.)
4. **Support** (Musketeer, Valkyrie, etc.)
5. **Small Spells** (Zap, Log, etc.)
6. **Cycle** (Knight, Ice Spirit, etc)

Within the same role tier, the card with the higher score receives the slot.

```bash
# Configure slot limit (default: 2)
./bin/cr-api deck build --tag PLAYER_TAG --evolution-slots 3
```

## Evolution Bonus Scoring

Cards with unlocked evolutions receive a scoring bonus:

```
base_bonus = 0.25 * (level/maxLevel)^1.5 * (1 + 0.2*(maxEvoLevel-1))
```

**Examples**:
- Level 10/14, Evo 1: +0.15
- Level 14/14, Evo 1: +0.25
- Level 14/14, Evo 3: +0.35

Additional 10% bonus applies to cards with evolution role overrides.

## Evolution Recommendations

The `evolutions recommend` command analyzes your card collection and shard inventory to recommend optimal evolution upgrades.

### Usage

```bash
# Basic recommendations
./bin/cr-api evolutions recommend --tag PLAYER_TAG

# Top 5 recommendations with detailed reasons
./bin/cr-api evolutions recommend --tag PLAYER_TAG --top 5 --verbose

# Override unlocked evolutions
./bin/cr-api evolutions recommend --tag PLAYER_TAG --unlocked-evolutions "Archers,Knight"
```

### Scoring Factors

1. **Level Ratio** (0-40 points): Higher level cards score better
   - Formula: `(level/maxLevel) * 40`

2. **Shard Progress** (0-30 points): Cards closer to evolution threshold
   - Formula: `(shards/required) * 30`

3. **Role Priority** (0-20 points): Strategic importance
   - Win Conditions: 20
   - Buildings: 16
   - Big Spells: 12
   - Support: 8
   - Small Spells: 4
   - Cycle: 0

4. **Multi-Evolution Bonus** (0-10 points): Cards with 2+ evolution stages

### Shard Management

```bash
# List current shard inventory
./bin/cr-api evolutions shards list

# Set shard count for a card
./bin/cr-api evolutions shards set --card "Archers" --count 50

# Filter by card name
./bin/cr-api evolutions shards list --card "Knight"
```

Shard data is stored in `data/evolution_shards.json`.

## Upgrade Priorities

Evolution-capable cards receive bonus priority in upgrade calculations:

- **Base bonus**: +10 points for any evolution-capable card
- **Level bonus**: Up to +20 points based on `(level/maxLevel) * 20`
- **Progress bonus**: Up to +5 points for partial evolution progress

This ensures you're guided to upgrade cards that will benefit most from evolution.

## CSV Export

Evolution data is included in CSV exports:

**Player Cards** (`players/`):
- `Evolution Level`: Current evolution stage (0-3)
- `Max Evolution Level`: Maximum possible evolutions (0-3)

**Analysis** (`analysis/`):
- Same evolution fields for card level analysis

**Event Decks** (`events/`):
- Cards show evolution level: `"Hog Rider (Lv.11 Evo.2)"`

## Deck Display

When displaying decks, evolution information is shown:

```
Deck 1 - Hog Cycle
#  Card           Level           Elixir  Role
1  Hog Rider      14/14           4       Win Condition
2  Valkyrie       14/14 (⚡)      4       Support
3  Ice Spirit     14/14           1       Cycle

Notes:
• Evolution slots: Valkyrie, Knight
```

The `⚡` badge indicates a card is using an evolution slot.

## API Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `evolutionLevel` | int | Current evolution stage (0-3) |
| `maxEvolutionLevel` | int | Maximum possible evolutions (0-3) |
| `requiredForEvolution` | int | Shards needed for next evolution |

## Testing

Comprehensive test coverage for evolution features:

- `pkg/deck/role_classifier_test.go`: Evolution role classification
- `pkg/deck/builder_test.go`: Evolution slot selection tests
- `pkg/deck/evolution_recommender_test.go`: Recommendation algorithm
- `pkg/analysis/upgrade_calculator_test.go`: Evolution bonus in upgrade priorities

Run tests:
```bash
go test ./pkg/deck/... -v -run Evolution
go test ./pkg/analysis/... -v -run Evolution
```
