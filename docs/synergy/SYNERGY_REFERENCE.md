# Synergy System Reference

## Overview

The Clash Royale API synergy system contains **188 carefully crafted synergy pairs** that represent the most effective card combinations in the current meta. Each pair includes a synergy score from 0.0 to 1.0 and a detailed description of why the cards work well together.

## Synergy Categories

The system categorizes synergies into 7 distinct types:

| Category | Count | Description |
|----------|-------|-------------|
| Tank Support | 29 | Tanks paired with support troops that protect them or benefit from their tanking |
| Win Condition | 47 | Cards that directly contribute to taking towers and winning the game |
| Defensive | 14 | Combinations that excel at defending against pushes |
| Bait | 16 | Cards that force opponents to waste spells, enabling other plays |
| Spell Combo | 17 | Spells and troops that work together for maximum effect |
| Bridge Spam | 13 | Aggressive, lane-pressuring combinations |
| Cycle | 11 | Cheap cards that enable fast rotation and out-cycling |

**Total: 188 pairs covering ~60 unique cards**

## Synergy Scoring System

All synergies are scored on a scale of **0.0 to 1.0**:

- **0.95-1.0**: Iconic meta combos (e.g., Lava Hound + Balloon, Goblin Barrel + Princess, Lumberjack + Balloon)
- **0.85-0.94**: Strong synergies that define archetypes (e.g., Golem + Night Witch, Tornado + Executioner)
- **0.75-0.84**: Good synergies that are common in competitive play
- **0.70-0.74**: Solid complementary pairs that work well together

## Synergy Pairs by Category

### 1. Tank Support (29 pairs)

Tank + Support synergies represent combinations where a high-health tank (Giant, Golem, Lava Hound, PEKKA, Mega Knight, Electro Giant) is paired with support troops that benefit from the tanking or protect the tank.

| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| Giant | Witch | 0.90 | Witch supports Giant with splash damage and spawns |
| Giant | Sparky | 0.85 | Giant tanks while Sparky deals massive damage |
| Giant | Musketeer | 0.80 | Musketeer provides ranged support behind Giant |
| Giant | Dark Prince | 0.80 | Dark Prince provides splash support and charging damage |
| Giant | Mini P.E.K.K.A | 0.75 | Mini PEKKA defends then supports Giant counter-push |
| Golem | Night Witch | 0.95 | Classic Golem beatdown synergy |
| Golem | Baby Dragon | 0.85 | Baby Dragon provides splash support |
| Golem | Lumberjack | 0.90 | Lumberjack provides rage and fast clearing |
| Lava Hound | Mega Minion | 0.85 | Mega Minion provides strong air support |
| Lava Hound | Skeleton Dragons | 0.80 | Skeleton Dragons provide splash air support |
| Mega Knight | Bats | 0.75 | Bats provide fast swarm defense |
| Mega Knight | Inferno Dragon | 0.80 | Inferno Dragon handles tanks while MK defends |
| Mega Knight | Minions | 0.75 | Minions provide air support for MK |
| Mega Knight | Electro Wizard | 0.85 | E-Wiz provides reset and ranged support |
| Mega Knight | Goblin Gang | 0.70 | Goblin Gang provides defensive bait value |
| Electro Giant | Tornado | 0.90 | Tornado groups enemies for E-Giant zaps |
| Electro Giant | Heal Spirit | 0.80 | Heal Spirit sustains E-Giant push |
| Electro Giant | Mother Witch | 0.85 | Mother Witch converts swarms to hogs |
| Electro Giant | Dark Prince | 0.80 | Dark Prince provides splash and charging support |
| P.E.K.K.A | Electro Wizard | 0.85 | E-Wiz provides reset and support for PEKKA |
| P.E.K.K.A | Magic Archer | 0.80 | Magic Archer provides ranged piercing support |
| P.E.K.K.A | Dark Prince | 0.80 | Dark Prince provides splash support |

#### Additional Tank Support Pairs (19 more)
- PEKKA + Battle Ram (0.85)
- PEKKA + Bandit (0.80)
- Ram Rider + PEKKA (0.80)
- Ram Rider + Mega Knight (0.75)
- Fireball synergies with Giant (0.80), Golem (0.85), Lava Hound (0.80), PEKKA (0.80)
- Lightning synergies with Lava Hound (0.80), PEKKA (0.85), Royal Giant (0.90), Electro Giant (0.85)
- Support troop pairs: Mega Minion + Bats (0.75), Electro Wizard + Mega Minion (0.75), etc.

### 2. Win Condition (47 pairs)

Win Condition synergies represent the core cards that directly contribute to taking towers and winning the game. These are the most important synergies in deck building.

#### Hog Rider (3 pairs)
| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| Hog Rider | Fireball | 0.80 | Fireball clears defenders for Hog |
| Hog Rider | Earthquake | 0.85 | Earthquake destroys buildings for Hog |
| Hog Rider | Freeze | 0.80 | Freeze guarantees Hog tower damage |

#### Royal Giant (3 pairs)
| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| Royal Giant | Fisherman | 0.85 | Fisherman activates King Tower for RG |
| Royal Giant | Lightning | 0.90 | Lightning clears defensive buildings |
| Royal Giant | Hunter | 0.75 | Hunter provides defensive synergy |

#### X-Bow (3 pairs)
| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| X-Bow | Tesla | 0.90 | Double building lock |
| X-Bow | Archers | 0.80 | Archers defend X-Bow |
| X-Bow | Ice Golem | 0.80 | Ice Golem kites for X-Bow defense |

#### Mortar (4 pairs)
| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| Mortar | Cannon | 0.85 | Mortar + defensive building |
| Mortar | Knight | 0.80 | Knight tanks and defends for Mortar |
| Mortar | Archers | 0.75 | Archers support Mortar defense |
| Mortar | Skeletons | 0.70 | Skeletons cycle and defend |

#### Miner (4 pairs)
| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| Miner | Balloon | 0.90 | Miner tanks for Balloon |
| Miner | Goblin Barrel | 0.75 | Dual win condition pressure |
| Miner | Wall Breakers | 0.80 | Dual tower pressure |
| Miner | Skeleton Barrel | 0.75 | Dual air pressure |

#### Balloon & Lava Hound (2 pairs)
| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| Lava Hound | Balloon | 0.95 | LavaLoon: overwhelming air pressure |
| Freeze | Balloon | 0.90 | Freeze guarantees Balloon connection |

#### Three Musketeers (2 pairs)
| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| Three Musketeers | Battle Ram | 0.90 | 3M split with Battle Ram pressure |
| Three Musketeers | Ice Golem | 0.80 | Ice Golem tanks for 3M split |

#### Sparky (3 pairs)
| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| Sparky | Giant | 0.90 | Giant tanks while Sparky charges |
| Sparky | Goblin Giant | 0.90 | Goblin Giant tanks with spear support |
| Sparky | Tornado | 0.85 | Tornado groups enemies for Sparky |

#### Additional Win Condition Pairs (26 more)
- Wall Breakers + Giant (0.75)
- Royal Hogs + Earthquake (0.85), Royal Hogs + Fisherman (0.75)
- Hog Rider + Valkyrie (0.80), Hog Rider + Ice Golem (0.80), Hog Rider + Musketeer (0.75)
- Various spell combos like Poison + Miner (0.85), Rage + Balloon (0.85), etc.

### 3. Defensive (14 pairs)

Defensive synergies create strong defensive combinations that can stop enemy pushes efficiently.

| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| Cannon | Ice Spirit | 0.80 | Cheap defensive combo |
| Cannon | Knight | 0.80 | Knight + Cannon cheap defense |
| Tesla | Ice Spirit | 0.75 | Tesla + Ice Spirit kiting |
| Tesla | Tornado | 0.85 | Tornado pulls troops to Tesla |
| Inferno Tower | Zap | 0.85 | Zap resets for Inferno Tower |
| Inferno Tower | Tornado | 0.90 | Tornado pulls tanks to Inferno |
| Inferno Dragon | Zap | 0.80 | Zap protects Inferno Dragon beam |
| Bomb Tower | Valkyrie | 0.75 | Dual splash defensive combo |
| Goblin Cage | Guards | 0.70 | Defensive troops chain |
| Mega Minion | Bats | 0.75 | Air defense combo |
| Musketeer | Ice Spirit | 0.75 | Musketeer + freeze for air defense |
| Hunter | Tornado | 0.85 | Tornado groups for Hunter burst |
| Electro Wizard | Mega Minion | 0.75 | E-Wiz reset + air defense |

### 4. Bait (16 pairs)

Bait synergies are designed to force the opponent to waste key spells, then punish them when those spells are unavailable.

| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| Goblin Barrel | Princess | 0.95 | Log bait: Princess baits log for Goblin Barrel |
| Goblin Barrel | Goblin Gang | 0.90 | Multiple goblin threats overwhelm spells |
| Goblin Barrel | Dart Goblin | 0.85 | Dart Goblin baits small spells |
| Goblin Barrel | Skeleton Army | 0.85 | Swarm bait forces spell usage |
| Goblin Barrel | Inferno Tower | 0.75 | Building bait punishes spell usage |
| Skeleton Barrel | Goblin Barrel | 0.80 | Double barrel pressure |
| Princess | Goblin Gang | 0.85 | Log bait pressure |
| Princess | Dart Goblin | 0.85 | Dual log bait threats |
| Graveyard | Skeleton Army | 0.75 | Skeleton flood overwhelms single spells |
| Graveyard | Tombstone | 0.80 | Continuous skeleton pressure |
| Skeleton Army | Goblin Gang | 0.80 | Dual swarm bait |
| Bats | Minions | 0.75 | Zap bait flying swarms |
| Spear Goblins | Goblins | 0.70 | Small spell bait pressure |
| Goblin Hut | Furnace | 0.75 | Building spam forces spell usage |
| X-Bow | Tesla | 0.90 | Double building bait and defense |

### 5. Spell Combo (17 pairs)

Spell combinations that work together to clear defenses or enable win conditions.

| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| Tornado | Fireball | 0.85 | Tornado groups troops for Fireball |
| Tornado | Rocket | 0.80 | Tornado + Rocket tower finish |
| Tornado | Executioner | 0.90 | Tornado pulls troops into Executioner's axe |
| Tornado | Bowler | 0.80 | Tornado + Bowler knockback combo |
| Tornado | Ice Wizard | 0.80 | Tornado groups for Ice Wizard slow |
| Tornado | Baby Dragon | 0.80 | Tornado pulls troops for Baby Dragon splash |
| Graveyard | Freeze | 0.90 | Freeze allows Graveyard skeletons to connect |
| Graveyard | Poison | 0.85 | Poison clears small troops from Graveyard |
| Poison | Miner | 0.85 | Poison + Miner chip damage combo |
| Earthquake | Royal Giant | 0.85 | Earthquake removes buildings for RG |
| Earthquake | Miner | 0.80 | Earthquake clears buildings for Miner |
| Freeze | Balloon | 0.90 | Freeze guarantees Balloon connection |
| Rage | Lumberjack | 0.85 | Double rage acceleration |
| Rage | Balloon | 0.85 | Rage accelerates Balloon to tower |
| Rage | Elite Barbarians | 0.80 | Rage boosts E-Barbs speed and DPS |
| Lumberjack | Balloon | 0.95 | LumberLoon: Rage boost for Balloon |
| Princess | Goblin Barrel | 0.95 | Log bait combo (also in Bait category) |

### 6. Bridge Spam (13 pairs)

Bridge spam focuses on fast, lane-pressuring combinations that overwhelm opponents with threats.

| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| P.E.K.K.A | Battle Ram | 0.85 | PEKKA Bridge Spam pressure |
| P.E.K.K.A | Bandit | 0.80 | Bandit supports PEKKA counterpush |
| Battle Ram | Bandit | 0.80 | Fast dual-lane pressure |
| Battle Ram | Minions | 0.75 | Air support for Battle Ram push |
| Battle Ram | Dark Prince | 0.85 | Dual charge pressure |
| Bandit | Royal Ghost | 0.75 | Invisible bridge spam |
| Bandit | Magic Archer | 0.75 | Bandit dash with Magic Archer support |
| Bandit | Electro Wizard | 0.75 | E-Wiz support for Bandit |
| Royal Ghost | Dark Prince | 0.75 | Dual invisible pressure |
| Royal Ghost | Minions | 0.70 | Air support for Ghost push |
| Lumberjack | Balloon | 0.95 | LumberLoon: Rage boost for Balloon |
| Ram Rider | PEKKA | 0.80 | PEKKA supports Ram Rider push |
| Ram Rider | Mega Knight | 0.75 | MK defends then Ram counterpush |

### 7. Cycle (11 pairs)

Cycle synergies with cheap cards that enable fast rotation and out-cycling opponents.

| Card 1 | Card 2 | Score | Description |
|--------|--------|-------|-------------|
| Ice Spirit | Skeletons | 0.85 | Ultra-cheap cycle combo |
| Ice Spirit | Fire Spirit | 0.80 | Cheap spirit cycle |
| Ice Spirit | Spear Goblins | 0.75 | Fast cycle defensive combo |
| Ice Spirit | Bats | 0.75 | Ultra-cheap air cycle |
| Ice Spirit | Log | 0.80 | Cheap cycle and control |
| Skeletons | Goblins | 0.80 | Fast cycle swarm combo |
| Skeletons | Ice Golem | 0.80 | Cheap cycle tank |
| Skeletons | Log | 0.75 | Cycle and clear combo |
| Fire Spirit | Heal Spirit | 0.75 | Dual spirit cycle |
| Fire Spirit | Goblins | 0.70 | Fast rotation combo |
| Heal Spirit | Skeletons | 0.75 | Ultra-fast cycle |

## Coverage Analysis

### Cards Covered

The synergy system covers approximately **60 unique cards** across all rarity types:

**Tanks:** Giant, Golem, Lava Hound, PEKKA, Mega Knight, Electro Giant, Giant Skeleton, Royal Giant, Goblin Giant

**Win Conditions:** Hog Rider, Miner, Balloon, Graveyard, Three Musketeers, X-Bow, Battle Ram, Mortar, Wall Breakers, Royal Hogs, Ram Rider, Elite Barbarians, Sparky, Skeleton Barrel, Miner

**Support Troops:** Witch, Night Witch, Sparky, Musketeer, Dark Prince, Mini PEKKA, Lumberjack, Baby Dragon, Skeleton Dragons, Electro Wizard, Magic Archer, Inferno Dragon, Mega Minion, Bats, Minions, Guards, Valkyrie, Knight, Ice Golem, Hunter, Spear Goblins, Goblins, Goblin Gang, Dart Goblin, Princess, Firecracker, Mother Witch, Fisherman

**Spells:** Fireball, Lightning, Earthquake, Freeze, Tornado, Rocket, Poison, Rage, Zap, Log

**Buildings:** Cannon, Tesla, Inferno Tower, Bomb Tower, Goblin Cage, X-Bow, Mortar, Furnace, Goblin Hut, Tombstone

**Cycle Cards:** Ice Spirit, Fire Spirit, Heal Spirit, Skeletons, Spear Goblins, Log, Skeleton Army, Ice Golem, Bats, Goblin Gang

### Missing Cards

The following cards are **not currently covered** by any synergy pairs:

**Newer Additions:** Silent Monk, Soul Stealer, Reaper, Demon Prince, Super Mini P.E.K.K.A

**Underrepresented:** Archers (only 2 pairs), Barbarians, Royal Recruits, Elite Barbarians, Battle Healer

**Spells Not Covered:** Heal, Clone, Mirror, Golden Knight

**Buildings Not Covered:** Elixir Collector

## Usage in Code

### Key Functions

```go
// Get synergy score between two cards (0.0-1.0)
func (db *SynergyDatabase) GetSynergy(card1, card2 string) float64

// Get full synergy pair details
func (db *SynergyDatabase) GetSynergyPair(card1, card2 string) *SynergyPair

// Analyze overall deck synergy
func (db *SynergyDatabase) AnalyzeDeckSynergy(deck []string) *DeckSynergyAnalysis

// Get card suggestions based on synergies
func (db *SynergyDatabase) SuggestSynergyCards(currentDeck []string, available []*CardCandidate) []*SynergyRecommendation
```

### In Builder

The synergy system is integrated into the deck builder via:

- `calculateSynergyScore(cardName string, deck []*CardCandidate) float64` - Average synergy with existing deck
- `synergyEnabled bool` - Toggle synergy scoring (default: false)
- `synergyWeight float64` - Weight multiplier (default: 0.15)
- `synergyCache map[string]float64` - O(1) cached lookups

### DeckSynergyAnalysis Output

```go
type DeckSynergyAnalysis struct {
    TotalScore       float64                 // 0-100 (normalized)
    AverageScore     float64                 // 0.0-1.0 avg per pair
    TopSynergies     []SynergyPair           // Best 5 pairs
    MissingSynergies []string                // Cards with no synergies
    CategoryScores   map[SynergyCategory]int // Count by category
}
```

## Integration Recommendations

### For Deck Evaluation

The synergy system provides a foundation for deck evaluation with these key advantages:

1. **Quantified Relationships**: Each pair has a measurable synergy score
2. **Categorized Logic**: 7 categories map to evaluation aspects (Attack, Defense, Countering)
3. **Coverage**: Good coverage of meta-relevant cards and combinations
4. **Tested**: Comprehensive test suite ensures reliability

### Missing for Full Evaluation

To support complete deck evaluation, the synergy system needs:

1. **Attack/Defense/Counter Ratings**: Current system has only raw scores, no categorization by role
2. **More Card Coverage**: Missing newer cards and underrepresented cards
3. **Synergy Matrix Generation**: Function to generate full 8x8 matrices for deck reports
4. **Individual Card Analysis**: Break down which cards contribute most to synergy

### Recommended Next Steps

1. **Categorize 188 pairs** by attack/defense/counter roles
2. **Add missing cards** and synergies (prioritize newer additions)
3. **Create synergyBonus map** structure for evaluation integration
4. **Implement matrix generation** for comprehensive deck reports
5. **Integrate with redundancy scorer** (not mutually exclusive with synergy)

## Maintenance Notes

This synergy database should be updated:
- When new cards are released (add relevant synergies)
- When meta shifts significantly (adjust scores)
- When balance changes affect interactions
- Based on competitive match data

**Current Version:** Based on pre-Demon Prince meta
**Last Updated:** 2026-01-01
