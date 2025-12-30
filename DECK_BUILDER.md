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

## Deck Building Strategies

The deck builder supports multiple strategies that adjust card selection, role priorities, and elixir targeting to match different playstyles.

### Available Strategies

Use the `--strategy` flag to specify a deck building strategy:

```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --strategy cycle
```

#### 1. Balanced (Default)

**Target Elixir**: 3.0-3.5
**Best For**: General ladder climbing, when you want a well-rounded deck

Neutral strategy with no role preferences. Selects highest-level cards across all roles for a versatile, competitive deck.

**Example**:
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --strategy balanced --verbose
# Average Elixir: ~3.2
# Composition: Standard (1 win condition, 1 building, 2 spells, 2 support, 2 cycle)
```

#### 2. Aggro

**Target Elixir**: 3.5-4.0
**Best For**: Aggressive playstyle, when you have overleveled win conditions

Prioritizes win conditions (1.3x multiplier) and support troops (1.1x) for constant offensive pressure.

**When to Use**:
- Your win conditions are significantly higher level than other cards
- You prefer offensive, high-pressure gameplay
- Enemy has weak defenses

**Example**:
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --strategy aggro --verbose
# Average Elixir: ~3.7
# Deck includes: High-level Hog Rider, strong support troops
```

#### 3. Control

**Target Elixir**: 3.5-4.2
**Best For**: Defensive playstyle, building-heavy strategies

Prioritizes defensive buildings (1.3x) and big spells (1.2x) for controlling the battlefield and making positive elixir trades.

**When to Use**:
- Your buildings are well-leveled
- You prefer defensive counterplay
- Meta favors building-centric strategies

**Example**:
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --strategy control --verbose
# Average Elixir: ~3.9
# Deck includes: Inferno Tower, Cannon, Fireball
```

#### 4. Cycle

**Target Elixir**: 2.5-3.0
**Best For**: Fast rotation, when you have many cheap cards

Strongly favors cycle cards (1.4x) and heavily penalizes high-cost cards (>4 elixir) to maintain fast deck rotation.

**When to Use**:
- You have many level-14 cycle cards (Skeletons, Ice Spirit, etc.)
- You want to out-cycle opponent's counters
- Fast-paced gameplay is your preference

**Special Mechanics**:
- Heavy penalty for cards >4 elixir (-0.3 per elixir over target)
- Small spells boosted (1.1x), big spells penalized (0.8x)

**Example**:
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --strategy cycle --verbose
# Average Elixir: ~2.8
# Deck includes: Skeletons, Ice Spirit, Knight, Hog Rider (fast rotation)
```

#### 5. Splash

**Target Elixir**: 3.2-3.8
**Best For**: AoE-focused gameplay, anti-swarm

Prioritizes support troops with area damage (1.3x) for handling swarm-heavy opponents.

**When to Use**:
- Meta is swarm-heavy (Goblin Gang, Skeleton Army, etc.)
- Your AoE troops (Baby Dragon, Wizard, Valkyrie) are high level
- You want strong defense against grouped troops

**Example**:
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --strategy splash --verbose
# Average Elixir: ~3.5
# Deck includes: Baby Dragon, Valkyrie, Wizard
```

#### 6. Spell

**Target Elixir**: 3.2-3.8
**Best For**: Spell cycling, chip damage strategies

Heavily prioritizes big spells (1.5x) and small spells (1.2x) with forced composition: **2 big spells, 0 buildings, 1 small spell**.

**Composition Override**:
- Always includes 2 big spells (Fireball, Rocket, Lightning, etc.)
- Never includes buildings (overridden to 0)
- Focuses on spell-based win conditions

**When to Use**:
- Your spells are significantly overleveled
- Opponent relies on defensive buildings
- You want spell cycle as a win condition

**Example**:
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --strategy spell --verbose
# Average Elixir: ~3.4
# Deck includes: Rocket, Fireball, Zap (spell-focused composition)
# Note: No buildings in deck due to composition override
```

### Strategy Selection Guide

| Scenario | Recommended Strategy |
|----------|---------------------|
| Balanced card levels, general climbing | **Balanced** |
| Overleveled win conditions | **Aggro** |
| Overleveled buildings/big spells | **Control** |
| Many maxed cycle cards | **Cycle** |
| Swarm-heavy meta | **Splash** |
| Overleveled spells, want chip damage | **Spell** |

### Technical Details

**Role Multipliers**: Each strategy applies multipliers to card roles during scoring:
- Values > 1.0 increase preference for that role
- Values < 1.0 decrease preference
- Balanced strategy uses 1.0 for all roles

**Elixir Targeting**: Strategies penalize cards outside their target elixir range:
- Standard penalty: -0.15 per elixir outside range
- Cycle strategy penalty: -0.3 per elixir for cards >4 elixir

**Composition Overrides**: Spell strategy forces specific role counts, overriding the default 1-1-1-1-2-2 distribution.

## Best Practices

1. Rebuild decks regularly as card levels change
2. Test new decks thoroughly before committing
3. Higher-level cards receive higher scores
4. Use `--strategy` to match your preferred playstyle
5. Combine strategies with `--unlocked-evolutions` for optimal results
6. Use `--verbose` to see which strategy is active