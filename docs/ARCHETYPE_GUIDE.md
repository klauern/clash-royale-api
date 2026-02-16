# Archetype Guide

**Understanding Deck Playstyles in Clash Royale**

This guide explains deck archetypes - strategic playstyles defined by card composition, elixir profile, and win conditions.

---

## Table of Contents

1. [What is an Archetype?](#what-is-an-archetype)
2. [Primary Archetypes](#primary-archetypes)
3. [Hybrid Archetypes](#hybrid-archetypes)
4. [Choosing Your Archetype](#choosing-your-archetype)
5. [Archetype Matchups](#archetype-matchups)

---

## What is an Archetype?

An **archetype** is a strategic deck category defined by:
- **Elixir range**: Average cost and curve shape
- **Win condition**: Primary tower damage source
- **Play pattern**: Offensive vs defensive focus
- **Card composition**: Role distribution

### Why Archetypes Matter

| Benefit | Impact |
|---------|--------|
| **Predictability** | Know what your deck can do |
| **Consistency** | Cards work toward the same goal |
| **Matchup knowledge** | Understand favorable/unfavorable matchups |
| **Mastery focus** | Learn one style deeply |

---

## Primary Archetypes

### 1. Beatdown

**Elixir Range:** 3.5-4.5 (Heavy)
**Win Condition:** Golem, Lava Hound, P.E.K.K.A, Giant, Electro Giant

#### Strategy
Build a massive push in back → defend until double elixir → overwhelm opponent with unstoppable push.

#### Core Components
- **1 Heavy Tank** (Golem, Lava Hound, Electro Giant)
- **2-3 Support Troops** (Night Witch, Lumberjack, Baby Dragon)
- **1 Building** (Tombstone, Elixir Collector)
- **1 Big Spell** (Lightning, Fireball, Poison)
- **1 Small Spell** (Zap, Log, Arrows)
- **1-2 Cycle Cards** (Ice Spirit, Skeletons)

#### Strengths
- Overwhelms defenses with sheer power
- Death value creates second wave
- One successful push wins game

#### Weaknesses
- Vulnerable to rush attacks (Hog, X-Bow)
- Requires patience and timing
- Punished by cycle decks

#### Example Deck
```
Golem | Night Witch | Baby Dragon | Mega Minion
Tombstone | Lightning | Zap | Ice Spirit
```

#### CLI Usage
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --strategy balanced --archetype beatdown
```

---

### 2. Cycle

**Elixir Range:** 2.4-3.0 (Fast)
**Win Condition:** Hog Rider, Royal Giant, Goblin Barrel

#### Strategy
Defend efficiently → counter-push → cycle back before opponent → repeat constantly.

#### Core Components
- **1 Fast Win Condition** (Hog Rider, Royal Giant, Goblin Barrel)
- **1-2 Cycle Cards** (Ice Spirit, Skeletons, Bats)
- **1 Defensive Building** (Cannon, Tesla)
- **1 Big Spell** (Fireball, Poison)
- **1 Small Spell** (Zap, Log, Arrows)
- **2 Support Troops** (Musketeer, Knight, Archers)

#### Strengths
- Always has win condition available
- Out-cycles opponent's counters
- Punishes slow decks

#### Weaknesses
- Weak to splash damage
- Struggles against heavy tanks
- Requires constant activity

#### Example Deck (Hog Cycle)
```
Hog Rider | Musketeer | Ice Golem | Cannon
Fireball | Zap | Skeletons | Ice Spirit
```

#### CLI Usage
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --strategy cycle
```

---

### 3. Control

**Elixir Range:** 3.0-4.0 (Medium-High)
**Win Condition:** Medium troops + spell chip, Miner chip

#### Strategy
Defend with positive elixir trades → chip damage over time → win through efficiency.

#### Core Components
- **1-2 Defensive Buildings** (Inferno Tower, Tesla, Bomb Tower)
- **2 Big Spells** (Fireball, Poison, Rocket)
- **1 Small Spell** (Zap, Log, Giant Snowball)
- **2-3 Defensive Troops** ( Valkyrie, Mini P.E.K.K.A, Executioner)
- **1 Win Condition** (Miner, Graveyard, X-Bow)

#### Strengths
- Positive elixir trades
- Counters almost any push
- Spell damage adds up

#### Weaknesses
- Weak to bait decks
- Can be out-cycled
- Limited offensive pressure

#### Example Deck
```
Inferno Tower | Valkyrie | Executioner | Musketeer
Fireball | Poison | Miner | Zap
```

#### CLI Usage
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --strategy control
```

---

### 4. Siege

**Elixir Range:** 3.0-4.0 (Medium)
**Win Condition:** X-Bow, Mortar

#### Strategy
Protect siege building → lock onto tower → defend until timer → win by tower destruction.

#### Core Components
- **1 Siege Building** (X-Bow, Mortar)
- **2 Defensive Buildings** (Tesla, Cannon, Furnace)
- **1 Big Spell** (Fireball, Rocket)
- **1 Small Spell** (Log, Zap, Arrows)
- **2-3 Support Troops** (Archers, Ice Wizard, Electro Wizard)
- **0-1 Cycle Card** (Ice Spirit)

#### Strengths
- Forces opponent to play defense
- Controls bridge area
- Wins time efficiently

#### Weaknesses
- Hard-countered by heavy tanks
- Requires precise placement
- Punished by rush attacks

#### Example Deck
```
X-Bow | Tesla | Cannon | Ice Wizard
Archers | Fireball | Log | Ice Spirit
```

#### CLI Usage
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --archetype siege
```

---

### 5. Spell Bait

**Elixir Range:** 2.6-3.2 (Fast-Medium)
**Win Condition:** Goblin Barrel, Graveyard

#### Strategy
Deploy bait cards → opponent uses spell → punish with real win condition → repeat.

#### Core Components
- **1 Primary Win Condition** (Goblin Barrel, Graveyard)
- **2-3 Bait Cards** (Princess, Goblin Gang, Dart Goblin)
- **1 Defensive Building** (Cannon, Tombstone)
- **1-2 Small Spells** (Zap, Log, Arrows)
- **1 Big Spell** (Fireball, Poison)
- **1-2 Support Troops** (Ice Wizard, Musketeer)

#### Strengths
- Punishes spell-heavy opponents
- Unpredictable win conditions
- Fast cycling

#### Weaknesses
- Fails against spell-cycle decks
- Requires bluffing skill
- Bait cards must be threatened

#### Example Deck (Log Bait)
```
Goblin Barrel | Princess | Goblin Gang | Dart Goblin
Cannon | Fireball | Ice Spirit | Log
```

#### CLI Usage
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --archetype bait
```

---

### 6. Bridge Spam

**Elixir Range:** 3.2-4.2 (Medium-High)
**Win Condition:** P.E.K.K.A, Battle Ram, Royal Giant, Bandit

#### Strategy
Deploy medium-cost troops at bridge → overwhelm with multiple threats → constant pressure.

#### Core Components
- **2-3 Bridge Spam Troops** (P.E.K.K.A, Battle Ram, Bandit, Royal Giant)
- **1 Big Spell** (Fireball, Poison, Lightning)
- **1 Small Spell** (Zap, Log)
- **1 Building** (Tesla, Goblin Cage)
- **1-2 Support Troops** (Electro Wizard, Mega Minion)
- **0-1 Cycle Card** (Ice Spirit)

#### Strengths
- Overwhelms single-target defenses
- Constant bridge pressure
- Flexible win conditions

#### Weaknesses
- Weak to splash damage
- Punished by swarm cards
- High elixir cost

#### Example Deck
```
P.E.K.K.A | Battle Ram | Bandit | Electro Wizard
Tesla | Fireball | Zap | Ice Spirit
```

#### CLI Usage
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --archetype bridge-spam
```

---

## Hybrid Archetypes

### LavaLoon (Beatdown + Air)

**Elixir Range:** 3.8-4.3

Hybrid of beatdown and air tactics. Lava Hound tanks tower → Balloon arrives with Hound → Tornado repositions defenders.

**Core:** Lava Hound, Balloon, Tornado, Ice Wizard

### Control Beatdown

**Elixir Range:** 3.2-3.8

Uses control tactics until double elixir, then transitions to beatdown pushes.

**Core:** P.E.K.K.A, Electro Wizard, defensive building, spells

### Cycle Siege

**Elixir Range:** 2.8-3.4

Fast cycle deck that occasionally locks X-Bow/Mortar for chip damage.

**Core:** X-Bow, fast cycle cards, defensive building

---

## Choosing Your Archetype

### Based on Card Levels

| Your Best Cards | Recommended Archetype |
|-----------------|----------------------|
| High-level tanks (Golem, Giant) | **Beatdown** |
| High-level Hog/Royal Giant | **Cycle** |
| High-level buildings | **Control** or **Siege** |
| High-level spells | **Spell Bait** or **Control** |
| Balanced mid-cost cards | **Bridge Spam** |

### Based on Playstyle

| Your Preference | Recommended Archetype |
|-----------------|----------------------|
| Patient, big pushes | **Beatdown** |
| Fast, constant action | **Cycle** |
| Defensive, efficient | **Control** |
| Strategic, positioning | **Siege** |
| Unpredictable, baiting | **Spell Bait** |
| Aggressive, overwhelming | **Bridge Spam** |

### Based on Trophy Range

| Trophy Range | Favored Archetypes |
|--------------|-------------------|
| Low (0-3000) | Beatdown, Cycle (simpler to execute) |
| Mid (3000-5000) | All archetypes viable |
| High (5000+) | Cycle, Control, Spell Bait (skill-based) |

---

## Archetype Matchups

### The Triangle Theory

**Beatdown > Siege > Control > Beatdown**

| Matchup | Advantage | Why |
|---------|-----------|-----|
| Beatdown vs Siege | Beatdown | Tanks absorb siege damage, reach tower |
| Siege vs Control | Siege | Out-ranges defensive buildings |
| Control vs Beatdown | Control | Positive trades vs slow pushes |
| Cycle vs Beatdown | Cycle | Out-cycles, opposite lane pressure |
| Beatdown vs Cycle | Beatdown | Overwhelms light defense |

### Detailed Matchup Table

| Defender | vs Beatdown | vs Cycle | vs Control | vs Siege | vs Bait | vs Bridge Spam |
|----------|-------------|----------|------------|----------|---------|----------------|
| **Beatdown** | Even | Lose (cycle) | Lose (control) | Win | Even | Even |
| **Cycle** | Win | Even | Lose | Win | Lose | Win |
| **Control** | Win | Win | Even | Lose | Win | Lose |
| **Siege** | Lose | Lose | Win | Even | Even | Lose |
| **Bait** | Even | Win | Lose | Even | Even | Win |
| **Bridge Spam** | Even | Lose | Win | Win | Lose | Even |

### Matchup Strategies

#### Beatdown vs Cycle
- **Play defensively until double elixir**
- Use cycle cards to scout opponent's counters
- One big push in OT wins

#### Cycle vs Beatdown
- **Constant opposite lane pressure**
- Don't let them build a push
- Out-cycle their win condition

#### Control vs Siege
- **Don't let siege lock**
- Pressure opposite lane
- Save big spells for siege building

#### Siege vs Control
- **Play siege when they have low elixir**
- Protect with buildings
- Timer is your friend

---

## Detecting Your Archetype

The deck builder can detect your deck's archetype:

```bash
# Analyze a deck
./bin/cr-api deck evaluate --deck "Hog,Ice Golem,Musketeer,Log,Zap,Cannon,Skeletons,Ice Spirit" --verbose

# Build for specific archetype
./bin/cr-api deck build --tag '#PLAYERTAG' --strategy cycle
```

The evaluation reports:
- Primary archetype match percentage
- Secondary archetype matches
- Archetype coherence score

---

## Tuning Your Archetype

### Adjusting Elixir Curve

| Want Faster | Want Slower |
|-------------|-------------|
| Add 1-2 cost cards | Add 5+ cost cards |
| Remove high-cost cards | Remove cycle cards |
| Target 2.6-2.8 avg | Target 3.5-4.0 avg |

### Adjusting Playstyle

| Want More Aggression | Want More Defense |
|---------------------|-------------------|
| Add second win condition | Add defensive building |
| Remove defensive building | Add big spell |
| Lower elixir curve | Raise elixir curve |

---

## References

- [DECK_BUILDER.md](./DECK_BUILDER.md) - Building decks by archetype
- [SYNERGY_GUIDE.md](./SYNERGY_GUIDE.md) - Card combinations for archetypes
- [COUNTER_RELATIONSHIPS.md](./COUNTER_RELATIONSHIPS.md) - Archetype matchups
- [DECK_STRATEGIES.md](./DECK_STRATEGIES.md) - Strategy templates
- [IMPROVED_SCORING_DESIGN.md](./IMPROVED_SCORING_DESIGN.md) - Archetype scoring algorithm
