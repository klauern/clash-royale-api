# Counter Relationships and Matchup Analysis

**Research Task:** clash-royale-api-33f5
**Date:** 2026-02-01
**Sources:** Deckshop.pro, community guides, competitive analysis

---

## Executive Summary

This document maps counter relationships in Clash Royale - what beats what. Understanding counters is essential for deck building, match prediction, and vulnerability analysis.

Counter types:
- **Hard Counter:** Completely shuts down the opponent's card
- **Soft Counter:** Favorable matchup but allows some damage or requires skill
- **Conditional Counter:** Depends on levels, timing, or support

---

## Table of Contents

1. [Counter Categories](#1-counter-categories)
2. [Critical Counter Pairs](#2-critical-counter-pairs)
3. [Win Condition Counters](#3-win-condition-counters)
4. [Spell Interactions](#4-spell-interactions)
5. [Archetype Matchups](#5-archetype-matchups)
6. [Vulnerability Detection](#6-vulnerability-detection)

---

## 1. Counter Categories

### 1.1 Building Counters

Buildings counter building-targeting troops by pulling them away from towers.

**Inferno Tower Counters (Hard):**
| Target | Effectiveness | Notes |
|--------|---------------|-------|
| Golem | 0.95 | Melts tank efficiently |
| Giant | 0.90 | Classic counter |
| Lava Hound | 0.90 | Melts before tower damage |
| P.E.K.K.A | 0.95 | Fast melt |
| Mega Knight | 0.85 | Effective but MK can jump |
| Royal Giant | 0.80 | Range disadvantage |

**Tesla Counters:**
| Target | Effectiveness | Notes |
|--------|---------------|-------|
| Hog Rider | 0.85 | Reliable counter |
| Balloon | 0.80 | Hits air |
| Battle Ram | 0.80 | Stops charge |

**Cannon Counters:**
| Target | Effectiveness | Notes |
|--------|---------------|-------|
| Hog Rider | 0.80 | Cheap counter |
| Giant | 0.75 | Good for cost |
| Royal Giant | 0.70 | Range disadvantage |

**Building Vulnerabilities:**
| Building | Countered By |
|----------|--------------|
| Inferno Tower | Lightning, Electro Wizard, Zap (reset), Earthquake |
| Tesla | Rocket, Earthquake, Heavy spells |
| Cannon | Fireball, Heavy spells |
| Tombstone | Poison, Splash |

### 1.2 Spell Counters

Spells counter troops based on damage and area.

**Fireball Counters:**
| Target | Effectiveness | Level Dependent |
|--------|---------------|-----------------|
| Wizard | 0.85 | Yes - needs +1 level |
| Musketeer | 0.85 | Yes - needs +1 level |
| Witch | 0.80 | Partial |
| Three Musketeers | 0.90 | No - hits all three |
| Flying Machine | 0.80 | Yes |
| Barbarians | 0.75 | No |

**Lightning Counters:**
| Target | Effectiveness | Notes |
|--------|---------------|-------|
| Inferno Tower | 0.95 | Hard counter |
| Inferno Dragon | 0.90 | Hard counter |
| Three Musketeers | 0.95 | Hits all three |
| Wizard | 0.90 | No level dependence |
| Musketeer | 0.90 | No level dependence |
| Electro Wizard | 0.90 | Hard counter |

**Zap/Log Counters:**
| Target | Spell | Effectiveness |
|--------|-------|---------------|
| Skeleton Army | Zap | 0.90 |
| Minion Horde | Zap | 0.70 | Needs +2 levels |
| Goblin Barrel | Zap | 0.60 | Timing dependent |
| Goblin Barrel | Log | 0.85 |
| Princess | Log | 0.95 |
| Dart Goblin | Log | 0.90 |
| Goblin Gang | Log | 0.85 |

**Arrows Counters:**
| Target | Effectiveness |
|--------|---------------|
| Minion Horde | 0.95 |
| Skeleton Army | 0.90 |
| Goblin Gang | 0.85 |
| Princess | 0.90 |
| Evolved Firecracker | 0.70 | Cycles back |

**Rocket Counters:**
| Target | Effectiveness |
|--------|---------------|
| Elixir Collector | 0.80 | +2 elixir trade |
| Sparky | 0.90 |
| X-Bow | 0.85 |
| Mortar | 0.80 |
| Tower | 0.90 | Direct damage |

**Earthquake Counters:**
| Target | Effectiveness |
|--------|---------------|
| All Buildings | 0.90 |
| Elixir Collector | 0.85 |
| Tombstone | 0.95 |

### 1.3 Air Defense Counters

Cards that counter air threats.

**Hard Anti-Air:**
| Defender | Targets | Effectiveness |
|----------|---------|---------------|
| Musketeer | Balloon, Lava Hound, Baby Dragon | 0.80 |
| Wizard | Minion Horde, Balloon | 0.85 |
| Executioner | All air | 0.85 |
| Tesla | All air | 0.80 |
| Inferno Tower | Lava Hound, Balloon | 0.90 |

**Soft Anti-Air:**
| Defender | Targets | Effectiveness |
|----------|---------|---------------|
| Mega Minion | Balloon, Baby Dragon | 0.75 |
| Archers | Balloon, support | 0.70 |
| Ice Wizard | Slows air | 0.60 |
| Bats | Balloon, Inferno Dragon | 0.65 |

### 1.4 Swarm Counters

Splash damage and spells counter swarm troops.

**Splash Counters:**
| Splash Card | Counters | Effectiveness |
|-------------|----------|---------------|
| Valkyrie | Skeleton Army, Goblin Gang, Guards | 0.90 |
| Baby Dragon | Minion Horde, Bats, swarms | 0.85 |
| Wizard | Medium swarms | 0.80 |
| Executioner | Dense swarms | 0.85 |
| Bowler | Ground swarms | 0.85 |
| Bomber | Ground swarms | 0.80 |
| Dark Prince | Ground swarms | 0.80 |

### 1.5 Tank Killer Counters

High DPS cards counter tanks.

**Tank Killers:**
| Card | Counters | Effectiveness |
|------|----------|---------------|
| Mini P.E.K.K.A | Hog, Giant, Golem, Royal Giant | 0.85 |
| P.E.K.K.A | Golem, Mega Knight, Giant | 0.90 |
| Inferno Dragon | Any tank | 0.90 |
| Hunter | Golem, Giant, Royal Giant | 0.80 |
| Prince | Giant, Hog | 0.75 |

---

## 2. Critical Counter Pairs

### 2.1 Hard Counters (Complete Shutdown)

| Counter | Target | Effectiveness | Scenario |
|---------|--------|---------------|----------|
| Inferno Tower | Golem | 0.95 | Tank melt |
| Arrows | Minion Horde | 0.95 | Complete elimination |
| Zap | Skeleton Army | 0.90 | Complete elimination |
| Log | Princess | 0.95 | Complete elimination |
| Fireball | Three Musketeers | 0.95 | All three eliminated |
| Lightning | Inferno Tower | 0.95 | Building destroyed |
| Lightning | Three Musketeers | 0.95 | All three eliminated |
| Rocket | Sparky | 0.90 | Direct elimination |
| Tornado | Hog Rider | 0.85 | Pull to King Tower |
| Freeze | Inferno Tower | 0.85 | Beam reset |
| Electro Wizard | Inferno Tower | 0.85 | Permanent reset |
| Electro Wizard | Sparky | 0.95 | Charge reset |

### 2.2 Soft Counters (Favorable Matchup)

| Counter | Target | Effectiveness | Notes |
|---------|--------|---------------|-------|
| Musketeer | Baby Dragon | 0.70 | Wins 1v1 |
| Musketeer | Balloon | 0.75 | High DPS |
| Mini P.E.K.K.A | Hog Rider | 0.70 | If not spelled |
| Mini P.E.K.K.A | Giant | 0.75 | High DPS |
| Valkyrie | Goblin Gang | 0.80 | 360° splash |
| Valkyrie | Skeleton Army | 0.85 | Complete clear |
| Cannon | Hog Rider | 0.70 | Cheap counter |
| Tesla | Balloon | 0.70 | Air targeting |
| Mega Minion | Baby Dragon | 0.65 | Wins 1v1 |
| Barbarians | Hog Rider | 0.70 | If not spelled |

### 2.3 Counter-to-Counter (Meta Layer)

| Counter | Target | Effectiveness | Notes |
|---------|--------|---------------|-------|
| Lightning | Inferno Tower | 0.95 | Counters the counter |
| Electro Wizard | Inferno Tower | 0.80 | Reset beam |
| Zap | Inferno Dragon | 0.70 | Reset beam |
| Zap | Skeleton Army | 0.90 | Counters Hog counter |
| Poison | Graveyard counters | 0.70 | Clears small troops |
| Earthquake | Buildings | 0.90 | Counters building decks |
| Freeze | Inferno Tower | 0.85 | Counters tank counter |

---

## 3. Win Condition Counters

### 3.1 Hog Rider Counters

| Counter | Type | Effectiveness |
|---------|------|---------------|
| Cannon | Building | 0.80 |
| Tesla | Building | 0.85 |
| Inferno Tower | Building | 0.75 |
| Tombstone | Building | 0.70 |
| Mini P.E.K.K.A | Troop | 0.70 |
| Barbarians | Troop | 0.70 |
| Goblin Gang | Troop | 0.65 |
| Skeleton Army | Troop | 0.60 | Spell vulnerable |
| Tornado | Spell | 0.75 | King activation |
| Mega Knight | Troop | 0.75 |

### 3.2 Goblin Barrel Counters

| Counter | Type | Effectiveness |
|---------|------|---------------|
| Log | Spell | 0.90 |
| Barbarian Barrel | Spell | 0.85 |
| Zap | Spell | 0.70 | Timing dependent |
| Arrows | Spell | 0.85 |
| Electro Wizard | Troop | 0.80 | Split damage |
| Goblin Gang | Troop | 0.70 |
| Valkyrie | Troop | 0.85 | If timed right |

### 3.3 Balloon Counters

| Counter | Type | Effectiveness |
|---------|------|---------------|
| Tesla | Building | 0.80 |
| Inferno Tower | Building | 0.90 |
| Musketeer | Troop | 0.80 |
| Mega Minion | Troop | 0.75 |
| Wizard | Troop | 0.75 |
| Bats | Troop | 0.70 |
| E-Wiz | Troop | 0.75 | Reset + damage |

### 3.4 Golem Counters

| Counter | Type | Effectiveness |
|---------|------|---------------|
| Inferno Tower | Building | 0.95 |
| Inferno Dragon | Troop | 0.90 |
| Mini P.E.K.K.A | Troop | 0.80 |
| PEKKA | Troop | 0.90 |
| Hunter | Troop | 0.80 |
| Tornado | Spell | 0.70 | Separates support |

### 3.5 Graveyard Counters

| Counter | Type | Effectiveness |
|---------|------|---------------|
| Poison | Spell | 0.90 |
| Valkyrie | Troop | 0.85 |
| Archers | Troop | 0.80 |
| Executioner | Troop | 0.85 |
| Baby Dragon | Troop | 0.75 |
| Wizard | Troop | 0.80 |

### 3.6 X-Bow Counters

| Counter | Type | Effectiveness |
|---------|------|---------------|
| Rocket | Spell | 0.90 |
| Earthquake | Spell | 0.85 |
| Giant | Troop | 0.80 | Tank it |
| Golem | Troop | 0.85 | Tank it |
| Hog Rider | Win Con | 0.75 | Opposite lane pressure |
| Miner | Win Con | 0.70 | Chip while defending |

### 3.7 Royal Giant Counters

| Counter | Type | Effectiveness |
|---------|------|---------------|
| Inferno Tower | Building | 0.85 |
| PEKKA | Troop | 0.90 |
| Mini P.E.K.K.A | Troop | 0.80 |
| Hunter | Troop | 0.85 |
| Barbarians | Troop | 0.75 |
| E-Wiz | Troop | 0.70 |

### 3.8 Lava Hound Counters

| Counter | Type | Effectiveness |
|---------|------|---------------|
| Inferno Tower | Building | 0.90 |
| Inferno Dragon | Troop | 0.90 |
| Musketeer | Troop | 0.80 |
| Wizard | Troop | 0.80 |
| Executioner | Troop | 0.85 |
| Tesla | Building | 0.75 |

---

## 4. Spell Interactions

### 4.1 Level-Dependent Interactions

Critical breakpoints where level matters:

| Spell | Target | Requirement | Outcome |
|-------|--------|-------------|---------|
| Fireball | Wizard/Musketeer | +1 level | One-shot kill |
| Fireball | Flying Machine | +1 level | One-shot kill |
| Zap | Goblins | Equal level | One-shot kill |
| Zap | Minions | +2 levels | One-shot kill |
| Arrows | Minions | Equal level | One-shot kill |
| Arrows | Princess | Equal level | One-shot kill |
| Log | Princess | Equal level | One-shot kill |
| Log | Dart Goblin | Equal level | One-shot kill |

### 4.2 Spell Combinations

| Combo | Effect | Effectiveness |
|-------|--------|---------------|
| Fireball + Zap | Kills Wizard, Musketeer | 0.90 |
| Fireball + Log | Kills Witch, Wizard | 0.85 |
| Poison + Zap | Clears Graveyard counters | 0.80 |
| Lightning + Log | Kills medium troops | 0.75 |
| Rocket + Log | Kills most support | 0.85 |

---

## 5. Archetype Matchups

### 5.1 The Triangle Theory

**Beatdown > Siege > Control > Beatdown**

| Matchup | Advantage | Reason |
|---------|-----------|--------|
| Beatdown vs Siege | Beatdown | Tanks absorb siege damage |
| Siege vs Control | Siege | Out-ranges defensive buildings |
| Control vs Beatdown | Control | Positive trades against big pushes |
| Cycle vs Beatdown | Cycle | Out-cycles slow deck |
| Beatdown vs Cycle | Beatdown | Overwhelms light defense |

### 5.2 Archetype Vulnerabilities

| Archetype | Weak Against | Strong Against |
|-----------|--------------|----------------|
| Beatdown | Inferno, Cycle, PEKKA | Siege, Spell Bait |
| Hog Cycle | Splash, Spell Cycle | Beatdown |
| Log Bait | Multiple Spells | Single Spell decks |
| Bridge Spam | Swarms, Splash | No swarm counters |
| Siege | Beatdown, Heavy tanks | Control, Cycle |
| LavaLoon | Anti-air heavy | Anti-air light |
| Graveyard | Splash-heavy | Splash-light |

### 5.3 Speed Matchups

| Speed | Beats | Loses To |
|-------|-------|----------|
| Fast Cycle | Slow Beatdown | Aggro Spam |
| Medium Control | Fast Cycle | Heavy Beatdown |
| Slow Beatdown | Medium Control | Fast Cycle |

---

## 6. Vulnerability Detection

### 6.1 WASTED Framework

Minimum requirements for a viable deck:

| Letter | Requirement | Test |
|--------|-------------|------|
| W | Win Condition | Has dedicated tower damage card |
| A | Air Counter | Has 2+ air-targeting cards |
| S | Splash | Has 2+ splash damage sources |
| T | Tank Killer | Has high DPS or % damage |
| E | Elixir Management | Has pump or efficient cycle |
| D | Defense | Has building or defensive troops |

### 6.2 Common Vulnerabilities

**Weak to Air:**
- Missing: Musketeer, Wizard, Executioner, Mega Minion, Tesla
- Vulnerable to: Lava Hound, Balloon, Minion Horde

**Weak to Splash:**
- Missing: Valkyrie, Baby Dragon, Wizard, Bomber, splash spells
- Vulnerable to: Skeleton Army, Goblin Gang, Graveyard

**Weak to Tanks:**
- Missing: Inferno Tower, Inferno Dragon, Mini P.E.K.K.A, PEKKA, Hunter
- Vulnerable to: Golem, Giant, Mega Knight, Lava Hound

**Weak to Building-Targeters:**
- Missing: Cannon, Tesla, Inferno Tower, Tombstone
- Vulnerable to: Hog Rider, Royal Giant, Giant, Balloon

**Weak to Spell Bait:**
- Missing: 2+ cheap spells (Zap, Log, Arrows, Barbarian Barrel)
- Vulnerable to: Goblin Barrel, Princess, Goblin Gang

**Weak to Graveyard:**
- Missing: Poison, Valkyrie, Archers, splash at tower
- Vulnerable to: Graveyard + Freeze/Poison

### 6.3 Coverage Scoring

Calculate deck coverage (0-100):

```
Coverage = (WinCon ? 20 : 0) +
           (AirDefense ? 20 : 0) +
           (Splash ? 20 : 0) +
           (TankKiller ? 20 : 0) +
           (Defense ? 20 : 0)
```

| Score | Rating |
|-------|--------|
| 100 | Perfect coverage |
| 80-99 | Good coverage |
| 60-79 | Moderate vulnerabilities |
| 40-59 | Significant vulnerabilities |
| <40 | Critical vulnerabilities |

---

## Data Files

- [counter_matrix.json](../data/static/counter_matrix.json) - Structured counter relationships
- [synergy_patterns.json](../data/static/synergy_patterns.json) - Archetype patterns

## References

- [Counter/Matchup Research](../history/research_counter_matchup.md) - Original research spike
- [Deck Synergies](./DECK_SYNERGIES.md) - Archetype patterns
- [Synergy Reference](./synergy/SYNERGY_REFERENCE.md) - Card pair synergies
