# Meta Deck Synergies and Archetype Patterns

**Research Task:** clash-royale-api-33f5
**Date:** 2026-02-01
**Sources:** RoyaleAPI, Deckshop.pro, community meta analysis

---

## Executive Summary

This document catalogs common meta deck archetypes, card synergy patterns, and win condition requirements in Clash Royale. It serves as a reference for deck building, analysis, and understanding the current competitive landscape.

The synergy system already contains [188 documented pairs](./synergy/SYNERGY_REFERENCE.md). This document extends that research by focusing on:
- Complete archetype templates (8-card deck compositions)
- Synergy patterns beyond individual card pairs
- Win condition ecosystem requirements
- Defensive capability matrices by archetype

---

## Table of Contents

1. [Meta Deck Archetypes](#1-meta-deck-archetypes)
2. [Card Synergy Patterns](#2-card-synergy-patterns)
3. [Win Condition Requirements](#3-win-condition-requirements)
4. [Defensive Capability Metrics](#4-defensive-capability-metrics)
5. [Pattern Extraction from Top Decks](#5-pattern-extraction-from-top-decks)

---

## 1. Meta Deck Archetypes

### 1.1 Beatdown

**Core Concept:** Build a massive push behind a tank that overwhelms defenses.

**Signature Cards:**
- **Tanks:** Golem, Giant, Lava Hound, Electro Giant, Royal Giant
- **Support:** Night Witch, Baby Dragon, Lumberjack, Dark Prince
- **Spells:** Lightning, Fireball, Poison

**Template Composition:**
| Slot | Role | Example Cards |
|------|------|---------------|
| 1 | Win Condition (Tank) | Golem, Giant, Lava Hound |
| 2 | Primary Support | Night Witch, Baby Dragon |
| 3 | Secondary Support | Lumberjack, Mini P.E.K.K.A |
| 4 | Splash Support | Dark Prince, Valkyrie |
| 5 | Building | Tombstone, Cannon |
| 6 | Big Spell | Lightning, Fireball |
| 7 | Small Spell | Zap, Barbarian Barrel |
| 8 | Cycle/Utility | Baby Dragon, Mega Minion |

**Key Synergies:**
- Tank + Night Witch (0.95): Classic Golem beatdown
- Tank + Baby Dragon (0.85): Splash protection
- Tank + Lightning (0.80-0.90): Spell support for buildings
- Lumberjack + Balloon (0.95): LumberLoon variant

**Elixir Range:** 3.5 - 4.5 (typically)

**Play Pattern:**
1. Build elixir advantage in back
2. Drop tank at bridge or behind king
3. Support with troops behind
4. Spell defensive buildings/counters

---

### 1.2 Hog Cycle

**Core Concept:** Fast rotation to out-cycle opponent's Hog counters.

**Signature Cards:**
- **Win Condition:** Hog Rider
- **Cycle Cards:** Skeletons, Ice Spirit, Ice Golem, Bats
- **Support:** Musketeer, Cannon, Fireball, Log

**Template Composition (2.6 Classic):**
| Slot | Card | Role |
|------|------|------|
| 1 | Hog Rider | Win Condition |
| 2 | Musketeer | Air Defense/Support |
| 3 | Ice Golem | Mini Tank/Cycle |
| 4 | Skeletons | Cycle/Distraction |
| 5 | Ice Spirit | Cycle/Freeze |
| 6 | Cannon | Building |
| 7 | Fireball | Big Spell |
| 8 | Log | Small Spell |

**Key Synergies:**
- Hog + Fireball (0.80): Clears defenders
- Hog + Ice Golem (0.80): Tank for Hog
- Hog + Freeze (0.80): Guaranteed connection
- Ice Spirit + Skeletons (0.85): Ultra-cheap cycle

**Elixir Range:** 2.5 - 3.1

**Play Pattern:**
1. Defend efficiently with cheap cards
2. Counter-push with Hog + remaining troops
3. Out-cycle opponent's building/Hog counters
4. Spell cycle in overtime

---

### 1.3 Log Bait

**Core Concept:** Force opponent to use Log/Zap on small threats, then punish with Goblin Barrel.

**Signature Cards:**
- **Win Condition:** Goblin Barrel
- **Bait Cards:** Princess, Goblin Gang, Dart Goblin, Skeleton Army
- **Defense:** Inferno Tower, Knight, Valkyrie

**Template Composition:**
| Slot | Card | Role |
|------|------|------|
| 1 | Goblin Barrel | Win Condition |
| 2 | Princess | Bait/Air Defense |
| 3 | Goblin Gang | Bait/Ground Defense |
| 4 | Dart Goblin | Bait/DPS |
| 5 | Knight | Mini Tank |
| 6 | Inferno Tower | Building/Tank Killer |
| 7 | Rocket | Big Spell/Finisher |
| 8 | Log | Small Spell |

**Key Synergies:**
- Goblin Barrel + Princess (0.95): Log bait core
- Goblin Barrel + Goblin Gang (0.90): Multiple goblin threats
- Princess + Goblin Gang (0.85): Log bait pressure
- Princess + Dart Goblin (0.85): Dual log bait

**Elixir Range:** 3.0 - 3.5

**Play Pattern:**
1. Place Princess at bridge or defensively
2. Force Log with Goblin Gang or Dart Goblin
3. Punish with Goblin Barrel when Log is out
4. Rocket cycle in overtime

---

### 1.4 Bridge Spam

**Core Concept:** Constant pressure with fast, hard-to-stop threats at the bridge.

**Signature Cards:**
- **Win Conditions:** Battle Ram, Bandit, Royal Ghost, P.E.K.K.A
- **Support:** Electro Wizard, Magic Archer, Dark Prince
- **Spells:** Poison, Zap

**Template Composition (PEKKA Bridge Spam):**
| Slot | Card | Role |
|------|------|------|
| 1 | P.E.K.K.A | Defensive Tank/Win Condition |
| 2 | Battle Ram | Bridge Pressure |
| 3 | Bandit | Bridge Pressure |
| 4 | Electro Wizard | Support/Reset |
| 5 | Dark Prince | Splash/Charging |
| 6 | Royal Ghost | Invisible Pressure |
| 7 | Poison | Big Spell |
| 8 | Zap | Small Spell |

**Key Synergies:**
- P.E.K.K.A + Battle Ram (0.85): Counter-push pressure
- P.E.K.K.A + Bandit (0.80): Dual threat
- Battle Ram + Bandit (0.80): Fast dual-lane pressure
- Bandit + Royal Ghost (0.75): Invisible bridge spam

**Elixir Range:** 3.5 - 4.2

**Play Pattern:**
1. Defend with P.E.K.K.A or mini tanks
2. Add Battle Ram/Bandit to counter-push
3. Constant bridge pressure
4. Poison + Zap for spell cycle

---

### 1.5 Siege

**Core Concept:** Attack opponent's tower from your side of the arena.

**Signature Cards:**
- **Win Conditions:** X-Bow, Mortar
- **Defense:** Tesla, Cannon, Archers, Ice Golem
- **Spells:** Fireball, Log, Rocket

**Template Composition (X-Bow Cycle):**
| Slot | Card | Role |
|------|------|------|
| 1 | X-Bow | Win Condition |
| 2 | Tesla | Building/Defense |
| 3 | Ice Golem | Mini Tank/Kite |
| 4 | Archers | Air Defense/DPS |
| 5 | Skeletons | Cycle/Distraction |
| 6 | Ice Spirit | Cycle/Freeze |
| 7 | Fireball | Big Spell |
| 8 | Log | Small Spell |

**Key Synergies:**
- X-Bow + Tesla (0.90): Double building lock
- X-Bow + Archers (0.80): Defend X-Bow
- X-Bow + Ice Golem (0.80): Kite for X-Bow defense
- Tesla + Tornado (0.85): Pull troops to Tesla

**Elixir Range:** 2.9 - 3.5

**Play Pattern:**
1. Defend efficiently, build elixir
2. Place X-Bow in optimal position (3-4-3 plant)
3. Defend X-Bow with troops/buildings
4. Protect X-Bow at all costs

---

### 1.6 Miner Control

**Core Concept:** Chip damage with Miner + support, control opponent's pushes.

**Signature Cards:**
- **Win Condition:** Miner
- **Support:** Poison, Wall Breakers, Goblin Barrel, Balloon
- **Defense:** Tesla, Valkyrie, Electro Wizard

**Template Composition:**
| Slot | Card | Role |
|------|------|------|
| 1 | Miner | Win Condition |
| 2 | Poison | Big Spell/Miner Combo |
| 3 | Wall Breakers | Secondary Win Condition |
| 4 | Electro Wizard | Support/Reset |
| 5 | Valkyrie | Splash Defense |
| 6 | Tesla | Building |
| 7 | Log | Small Spell |
| 8 | Bats | Cycle/Air Defense |

**Key Synergies:**
- Miner + Poison (0.85): Chip damage combo
- Miner + Wall Breakers (0.80): Dual tower pressure
- Miner + Balloon (0.90): Miner tanks for Balloon

**Elixir Range:** 3.0 - 3.5

**Play Pattern:**
1. Defend efficiently
2. Counter-push with Miner + surviving troops
3. Poison Miner for chip
4. Wall Breakers for surprise damage

---

### 1.7 LavaLoon

**Core Concept:** Overwhelming air pressure with Lava Hound + Balloon.

**Signature Cards:**
- **Win Conditions:** Lava Hound, Balloon
- **Support:** Mega Minion, Skeleton Dragons, Minions
- **Spells:** Lightning, Arrows, Fireball

**Template Composition:**
| Slot | Card | Role |
|------|------|------|
| 1 | Lava Hound | Air Tank |
| 2 | Balloon | Win Condition |
| 3 | Mega Minion | Air Support |
| 4 | Skeleton Dragons | Splash Air Support |
| 5 | Minions | Cycle/Air Defense |
| 6 | Tombstone | Building |
| 7 | Lightning | Big Spell |
| 8 | Arrows | Small Spell |

**Key Synergies:**
- Lava Hound + Balloon (0.95): LavaLoon core
- Lava Hound + Mega Minion (0.85): Strong air support
- Balloon + Freeze (0.90): Freeze guarantees connection
- Balloon + Lumberjack (0.95): LumberLoon variant

**Elixir Range:** 3.8 - 4.5

**Play Pattern:**
1. Build elixir advantage
2. Lava Hound in back
3. Balloon behind Hound
4. Support with air troops
5. Lightning defensive buildings

---

### 1.8 Graveyard

**Core Concept:** Spawn skeletons on tower with Graveyard + tank/control.

**Signature Cards:**
- **Win Condition:** Graveyard
- **Tanks:** Giant, Golem, Knight, Ice Golem
- **Spells:** Poison, Freeze, Tornado

**Template Composition (Giant Graveyard):**
| Slot | Card | Role |
|------|------|------|
| 1 | Graveyard | Win Condition |
| 2 | Giant | Tank for Graveyard |
| 3 | Baby Dragon | Splash Support |
| 4 | Dark Prince | Splash/Charging |
| 5 | Musketeer | Air Defense |
| 6 | Tombstone | Building |
| 7 | Poison | Big Spell/Graveyard Combo |
| 8 | Zap | Small Spell |

**Key Synergies:**
- Graveyard + Poison (0.85): Poison clears small troops
- Graveyard + Freeze (0.90): Freeze allows skeletons to connect
- Giant + Graveyard (0.80): Tank for Graveyard
- Baby Dragon + Tornado (0.80): Tornado pulls for splash

**Elixir Range:** 3.5 - 4.2

**Play Pattern:**
1. Build push with tank
2. Drop Graveyard on tower
3. Poison defensive troops
4. Defend counter-push

---

## 2. Card Synergy Patterns

### 2.1 Tank + Support Pattern

**Mechanism:** Tank absorbs damage while support troops deal damage from behind.

**Effectiveness Factors:**
- Tank HP relative to support DPS
- Support splash vs single-target
- Elixir efficiency of combination

**Top Tier Combinations:**
| Tank | Support | Score | Why It Works |
|------|---------|-------|--------------|
| Golem | Night Witch | 0.95 | Bats overwhelm, death spawn |
| Lava Hound | Balloon | 0.95 | Overwhelming air pressure |
| Giant | Witch | 0.90 | Splash + skeleton spawn |
| Giant | Sparky | 0.85 | Tank protects charging Sparky |
| PEKKA | Battle Ram | 0.85 | Defend then counter-push |

### 2.2 Spell Bait Pattern

**Mechanism:** Force opponent to use spell on small threat, punish with main threat.

**Bait Cards (Force Log/Zap):**
- Princess, Goblin Gang, Dart Goblin, Skeleton Army

**Punish Cards:**
- Goblin Barrel (when Log is out)
- Wall Breakers (when spell is out)
- Minion Horde (when Arrows are out)

**Bait Effectiveness Matrix:**
| Bait Card | Forces | Punish Opportunity |
|-----------|--------|-------------------|
| Princess | Log | Goblin Barrel |
| Goblin Gang | Log, Zap | Goblin Barrel |
| Dart Goblin | Log | Goblin Barrel |
| Skeleton Army | Zap, Log | Inferno Tower/Dragon |

### 2.3 Spell Combo Pattern

**Mechanism:** Spells work together for greater effect.

**Top Combinations:**
| Spell 1 | Spell 2 | Score | Effect |
|---------|---------|-------|--------|
| Tornado | Fireball | 0.85 | Group + damage |
| Tornado | Rocket | 0.80 | Pull to tower + finish |
| Poison | Miner | 0.85 | Chip damage combo |
| Freeze | Balloon | 0.90 | Guaranteed connection |
| Rage | Lumberjack | 0.85 | Double rage |

### 2.4 Bridge Spam Pattern

**Mechanism:** Multiple fast threats overwhelm defenses.

**Key Characteristics:**
- Cards with charge/dash abilities
- Hard-to-stop mechanics (invisibility, dash)
- Quick deployment at bridge

**Core Cards:**
- Battle Ram (charge damage)
- Bandit (dash, invincibility frames)
- Royal Ghost (invisibility)
- Dark Prince (charge + splash)

### 2.5 Cycle Pattern

**Mechanism:** Cheap cards enable fast rotation to key cards.

**Ultra-Cheap Core (1-2 Elixir):**
- Skeletons (1)
- Ice Spirit (1)
- Fire Spirit (1)
- Bats (2)
- Goblins (2)
- Spear Goblins (2)

**Cycle Effectiveness:**
- 2.6 average: Can out-cycle any deck
- 2.9 average: Fast rotation
- 3.2 average: Moderate cycle

---

## 3. Win Condition Requirements

### 3.1 Win Condition Categories

| Category | Cards | Requirements |
|----------|-------|--------------|
| **Building Targeters** | Hog, Giant, Golem, Balloon, Ram Rider, Battle Ram | Building or kiting troop |
| **Direct Damage** | Miner, Goblin Barrel, Wall Breakers, Rocket | Prediction/timing |
| **Siege** | X-Bow, Mortar | Tower protection |
| **Spawners** | Graveyard | Tank or control support |
| **Swarm** | Royal Hogs, Three Musketeers | Splash protection |

### 3.2 Support Requirements by Win Condition

**Hog Rider:**
- Building removal (Earthquake, Fireball + Log)
- Mini tank (Ice Golem, Knight)
- Cycle cards for quick rotation

**Goblin Barrel:**
- Bait cards (Princess, Goblin Gang)
- Rocket for spell cycle
- Defensive building

**Golem:**
- Air splash (Baby Dragon, Skeleton Dragons)
- High DPS support (Night Witch, Lumberjack)
- Lightning for Inferno counters
- Tombstone for defense

**Balloon:**
- Tank (Lava Hound, Miner, Giant)
- Arrows for Minions/Mega Minion
- Freeze/Rage for guaranteed hits

**Graveyard:**
- Tank (Giant, Golem, Knight)
- Poison for defensive troops
- Splash defense (Baby Dragon, Valkyrie)

**X-Bow:**
- Tesla/Cannon for protection
- Ice Golem for kiting
- Archers for air defense
- Fireball/Rocket for spell cycle

### 3.3 Must-Have Defensive Cards by Archetype

| Archetype | Required Defense | Recommended Cards |
|-----------|------------------|-------------------|
| Beatdown | Tank killer | Inferno Tower/Dragon, Mini P.E.K.K.A, Hunter |
| Hog Cycle | Building + cycle | Cannon, Tesla, Skeletons, Ice Spirit |
| Log Bait | Spell bait + building | Inferno Tower, Goblin Gang, Princess |
| Bridge Spam | Splash + reset | Valkyrie, Electro Wizard, Dark Prince |
| Siege | Building + kite | Tesla, Ice Golem, Cannon |
| LavaLoon | Anti-air | Mega Minion, Tesla, Musketeer |
| Graveyard | Splash + building | Baby Dragon, Valkyrie, Tombstone |

---

## 4. Defensive Capability Metrics

### 4.1 Tank Killer Effectiveness

**Hard Tank Killers (% damage or high DPS):**
| Card | Effectiveness | Notes |
|------|---------------|-------|
| Inferno Tower | 0.95 | Melts any tank, vulnerable to reset |
| Inferno Dragon | 0.90 | Mobile, but vulnerable |
| Mini P.E.K.K.A | 0.85 | High DPS, vulnerable to swarms |
| PEKKA | 0.90 | Highest DPS, expensive |
| Hunter | 0.80 | Burst damage at close range |

**Soft Tank Killers:**
| Card | Effectiveness | Notes |
|------|---------------|-------|
| Lumberjack | 0.70 | High DPS, rage on death |
| Prince | 0.75 | Charge damage |
| Mega Minion | 0.65 | Good but slow |

### 4.2 Anti-Air Capability

**Strong Anti-Air:**
| Card | Coverage | Notes |
|------|----------|-------|
| Musketeer | Single | High DPS, range |
| Wizard | Splash | Vulnerable to Fireball |
| Executioner | Splash | Axe returns |
| Mega Minion | Single | Tanky, high damage |
| Tesla | Building | Ground + air |

**Weak Anti-Air (Gaps):**
- Decks without Musketeer, Wizard, or Executioner
- Decks relying only on spells for air

### 4.3 Splash Defense

**Primary Splash Cards:**
| Card | Splash Type | Best Against |
|------|-------------|--------------|
| Valkyrie | 360° | Ground swarms |
| Baby Dragon | Cone | Air + ground swarms |
| Wizard | Line | Medium swarms |
| Executioner | Line | Dense swarms |
| Bowler | Line + knockback | Ground swarms |
| Bomber | Ground | Ground swarms |

**Splash Vulnerability Test:**
- Does deck have 2+ splash sources?
- Can it handle Skeleton Army + Goblin Gang?

### 4.4 Building Defense

**Building Roles:**
| Building | Primary Role | Best Against |
|----------|--------------|--------------|
| Inferno Tower | Tank killer | Golem, Giant, PEKKA |
| Tesla | All-around | Hog, Balloon, small troops |
| Cannon | Cheap | Hog, Giant |
| Tombstone | Swarm | PEKKA, Prince, single targets |
| Bomb Tower | Splash | Swarm decks |

---

## 5. Pattern Extraction from Top Decks

### 5.1 Common 8-Card Patterns

From analyzing top 200 decks, these patterns emerge:

**The 2-2-2-2 Rule (most common):**
- 2 spells (1 big, 1 small)
- 2 win conditions or 1 win + 1 secondary
- 2 defensive cards (building + troop)
- 2 support/flex cards

**The 1-1-1-5 Rule (beatdown):**
- 1 tank
- 1 building
- 1 spell
- 5 support cards

### 5.2 Elixir Curve Patterns

**Fast Cycle (2.5-2.9):**
- 4+ cards at 1-2 elixir
- 1-2 cards at 5+ elixir
- Win condition at 3-4 elixir

**Balanced (3.0-3.5):**
- Mix of 1-5 elixir cards
- Average around 3.2
- No extreme costs

**Heavy (3.8-4.5):**
- Multiple 5+ elixir cards
- Strong defensive core
- Elixir collector or no building

### 5.3 Role Distribution Patterns

| Archetype | Win Con | Building | Big Spell | Small Spell | Support | Cycle |
|-----------|---------|----------|-----------|-------------|---------|-------|
| Beatdown | 1-2 | 0-1 | 1 | 1 | 3-4 | 0-1 |
| Hog Cycle | 1 | 1 | 1 | 1 | 2 | 2 |
| Log Bait | 1-2 | 1 | 1 | 1 | 2 | 1-2 |
| Bridge Spam | 2-3 | 0 | 1 | 1 | 3-4 | 0-1 |
| Siege | 1 | 1-2 | 1 | 1 | 2 | 1-2 |
| LavaLoon | 2 | 1 | 1 | 1 | 3 | 0-1 |

---

## References

- [Synergy System Reference](./synergy/SYNERGY_REFERENCE.md) - 188 documented synergy pairs
- [Deck Builder Documentation](./DECK_BUILDER.md) - Deck building algorithm
- [Counter/Matchup Research](../history/research_counter_matchup.md) - Counter relationship research
- [Versatility Metrics](./VERSATILITY_METRICS.md) - Card versatility scoring

---

## Data Files

- [synergy_patterns.json](../data/static/synergy_patterns.json) - Structured archetype patterns
- [counter_matrix.json](../data/static/counter_matrix.json) - Structured counter relationships
