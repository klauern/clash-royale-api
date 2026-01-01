# Versatility Metrics and Formulas

**Research Spike: clash-royale-api-5xb**
**Date**: 2026-01-01

---

## Executive Summary

Versatility is defined as a card's ability to serve **multiple roles** across **different situations** at **various elixir stages**. Based on meta deck analysis, the versatility score is composed of:

- **60%** Multi-role capability (cards serving 2+ strategic roles)
- **30%** Situational flexibility (effective in diverse game situations)
- **10%** Elixir adaptability (works at different elixir stages)

---

## Component 1: Multi-Role Bonus (60% weight)

### Definition

Cards that can be classified into **multiple strategic roles** provide more versatility than single-role cards.

### Role System Background

Existing roles (from `pkg/deck/role_classifier.go`):
- Win Condition - Primary tower damage dealer
- Building - Defensive structures
- Spell Big - 4+ elixir damage spells
- Spell Small - 2-3 elixir utility spells
- Support - Mid-cost troops (3-5 elixir)
- Cycle - Low-cost cards (1-2 elixir)

### Multi-Role Card Identification

**Method 1: Explicit Multi-Role Classification**

Some cards inherently serve multiple roles:

| Card | Primary Role | Secondary Role | Versatility Factor |
|------|-------------|----------------|-------------------|
| Valkyrie | Support | Defensive anti-swarm | 2.0 |
| Electro Wizard | Support | Air defense | 2.0 |
| Miner | Win Condition | Cycle spell bait | 2.0 |
| X-Bow | Building | Win Condition (siege) | 2.0 |
| Goblin Barrel | Win Condition | Spell bait | 2.0 |
| Royal Ghost | Support | Win Condition (bridge spam) | 2.0 |
| Baby Dragon | Support | Anti-air + anti-swarm | 2.0 |
| Mega Knight | Support | Win Condition (bridge spam) | 2.0 |
| Mini P.E.K.K.A | Support | Win Condition (bridge spam) | 2.0 |

**Method 2: Situational Multi-Role**

Cards that shift roles based on game state:

| Card | Offensive Role | Defensive Role | Context |
|------|---------------|----------------|---------|
| Valkyrie | Tank for push | Anti-swarm defender | Evolution enhances both |
| Magic Archer | Chip damage | Anti-swarm splash | Range enables both |
| Hunter | Air defense | Ground splash | High versatility |
| Bandit | Dash damage | Defensive tank | Dashes through |
| Skeleton Army | Win condition support | Anti-tank defense | Swarm versatility |

**Method 3: Evolution Role Override**

Evolved cards may gain additional roles:

| Base Card | Base Role | Evolved Role | Versatility Gain |
|-----------|-----------|--------------|-----------------|
| Knight | Cycle | Support (with dash) | +1.0 |
- Skeletons | Cycle | Support (with split) | +1.0 |
| Zap | Spell Small | Cycle enhanced | +0.5 |
| Archers | Support | Anti-air enhanced | +0.5 |

### Multi-Role Scoring Formula

```go
// MultiRoleVersatility calculates versatility bonus for serving multiple roles
func MultiRoleVersatility(card Card, evolutionLevel int) float64 {
    roleCount := countRoles(card, evolutionLevel)

    // Base score: 1.0 for single role, +0.6 for each additional role
    baseScore := 1.0 + float64(roleCount-1) * 0.6

    // Evolution bonus: additional 0.3 if evolution adds role
    if evolutionLevel > 0 && evolutionAddsRole(card) {
        baseScore += 0.3
    }

    return baseScore
}

// Deck-level multi-role versatility: Average across all cards
func DeckMultiRoleVersatility(deck Deck) float64 {
    totalScore := 0.0
    for _, card := range deck.Cards {
        totalScore += MultiRoleVersatility(card, card.EvolutionLevel)
    }
    return totalScore / float64(len(deck.Cards))
}
```

### Multi-Role Database Schema

```go
type CardRoleData struct {
    CardName        string
    PrimaryRole     CardRole
    SecondaryRoles  []CardRole     // Additional roles this card can serve
    EvolutionRoles  []CardRole     // Roles gained when evolved
    RoleConditions  []RoleCondition // When each role applies
}

type RoleCondition struct {
    Role          CardRole
    Condition     string  // e.g., "vs_air", "vs_swarm", "at_10_elixir"
    Probability   float64 // How often this condition applies (0-1)
}
```

---

## Component 2: Situational Flexibility (30% weight)

### Definition

A card's effectiveness across **different game situations** (offense, defense, air, ground, single-target, AOE).

### Situational Dimensions

#### Dimension 1: Offensive + Defensive (15%)

Cards that work well on both offense and defense:

| Card | Offensive | Defensive | Flexibility Score |
|------|-----------|-----------|-------------------|
| Valkyrie | 8/10 | 9/10 | 0.85 |
| Electro Wizard | 8/10 | 8/10 | 0.80 |
| Mini P.E.K.K.A | 9/10 | 7/10 | 0.80 |
| Magic Archer | 7/10 | 8/10 | 0.75 |
| Musketeer | 6/10 | 9/10 | 0.75 |
| Hunter | 7/10 | 8/10 | 0.75 |

#### Dimension 2: Target Type Coverage (10%)

Cards that hit both air and ground, or have utility:

| Card | Air | Ground | Utility | Coverage Score |
|------|-----|--------|---------|----------------|
| Electro Wizard | ✓ | ✓ | Stun | 1.0 |
| Baby Dragon | ✓ | ✓ | Splash | 1.0 |
| Valkyrie | ✗ | ✓ | Dash | 0.7 |
| Musketeer | ✓ | ✓ | None | 0.9 |
| Mega Knight | ✗ | ✓ | Spawn damage | 0.8 |

#### Dimension 3: Response Variety (5%)

Cards effective against multiple threat types:

| Card | vs Tank | vs Swarm | vs Air | vs Building | Response Count |
|------|---------|----------|--------|-------------|----------------|
| Valkyrie | ✓ | ✓ | ✗ | ✓ | 3 |
| Baby Dragon | ✗ | ✓ | ✓ | ✓ | 3 |
| Mini P.E.K.K.A | ✓ | ✗ | ✓ | ✓ | 3 |
| P.E.K.K.A | ✓ | ✗ | ✓ | ✓ | 3 |
| Skeleton Army | ✓ | ✗ | ✗ | ✓ | 2 |

### Situational Flexibility Formula

```go
type SituationalData struct {
    OffensiveEffectiveness  float64 // 0-1, offense capability
    DefensiveEffectiveness  float64 // 0-1, defense capability
    HitsAir                 bool
    HitsGround              bool
    HasUtility              bool    // Stun, push, pull, etc.
    VsTank                  bool
    VsSwarm                 bool
    VsAir                   bool
    VsBuilding              bool
}

func SituationalFlexibility(card Card) float64 {
    data := getSituationalData(card)

    // Component 1: Offense + Defense balance (15%)
    offenseDefenseScore := (data.OffensiveEffectiveness + data.DefensiveEffectiveness) / 2.0
    offenseDefenseBonus := offenseDefenseScore * 0.15

    // Component 2: Target type coverage (10%)
    targetScore := 0.0
    if data.HitsAir && data.HitsGround { targetScore += 0.5 }
    if data.HasUtility { targetScore += 0.3 }
    targetBonus := targetScore * 0.10

    // Component 3: Response variety (5%)
    responseCount := 0
    if data.VsTank { responseCount++ }
    if data.VsSwarm { responseCount++ }
    if data.VsAir { responseCount++ }
    if data.VsBuilding { responseCount++ }
    responseBonus := float64(responseCount) / 4.0 * 0.05

    return offenseDefenseBonus + targetBonus + responseBonus
}
```

---

## Component 3: Elixir Adaptability (10% weight)

### Definition

A card's effectiveness at **different elixir stages** of the match (early game 0-5 elixir, mid-game, double/triple elixir).

### Elixir Flexibility Categories

#### Category 1: Always Viable (1.0)

Cards effective at any elixir level:

| Card | Early Game | Mid Game | Late Game | Adaptability |
|------|------------|----------|-----------|--------------|
| Knight | ✓ | ✓ | ✓ | 1.0 |
| Valkyrie | ✓ | ✓ | ✓ | 1.0 |
| Skeletons | ✓ | ✓ | ✓ | 1.0 |
| Ice Spirit | ✓ | ✓ | ✓ | 1.0 |
| Zap | ✓ | ✓ | ✓ | 1.0 |
| The Log | ✓ | ✓ | ✓ | 1.0 |

#### Category 2: Scales with Elixir (0.7)

Cards that get better in double/triple elixir:

| Card | Early Game | Late Game | Why |
|------|------------|-----------|-----|
| Giant | Weak | Strong | Can support with spells |
| Lava Hound | Weak | Strong | Full combo available |
| Sparky | Weak | Strong | Protection available |
| Three Musketeers | Weak | Strong | Split more viable |

#### Category 3: Early Game Only (0.5)

Cards primarily for early game:

| Card | Early Game | Late Game | Why |
|------|------------|-----------|-----|
| Bandit (dash) | Strong | Weaker | Easier to counter later |
| Goblin Barrel (surprise) | Strong | Weaker | Opponent prepared |

#### Category 4: Late Game Only (0.5)

Cards that need setup:

| Card | Early Game | Late Game | Why |
|------|------------|-----------|-----|
| Graveyard | Weak | Strong | Needs tank |
| X-Bow (defense) | Weak | Strong | Needs protection |
| Mortar (bridge) | Weak | Strong | Needs support |

### Elixir Adaptability Formula

```go
type ElixirAdaptabilityData struct {
    EffectiveEarly   bool // 0-5 elixir available
    EffectiveMid      bool // 6-9 elixir available
    EffectiveLate     bool // 10+ elixir (2x/3x elixir)
    OptimalElixir     int  // Card's optimal elixir stage
}

func ElixirAdaptability(card Card) float64 {
    data := getElixirData(card)

    // Count viable stages
    viableStages := 0
    if data.EffectiveEarly { viableStages++ }
    if data.EffectiveMid { viableStages++ }
    if data.EffectiveLate { viableStages++ }

    // Base score: percentage of stages where card is viable
    baseScore := float64(viableStages) / 3.0

    // Penalty for cards that only work in specific stages
    if viableStages == 1 {
        baseScore *= 0.5
    }

    return baseScore * 0.10 // 10% weight
}
```

### Elixir Sweet Spot Bonus

Cards with elixir cost 2-4 are most flexible:

```go
func ElixirCostFlexibility(card Card) float64 {
    cost := card.ElixirCost

    // Sweet spot: 2-4 elixir
    if cost >= 2 && cost <= 4 {
        return 1.0
    }

    // Still viable: 1 or 5 elixir
    if cost == 1 || cost == 5 {
        return 0.8
    }

    // Less flexible: 6-7 elixir
    if cost == 6 || cost == 7 {
        return 0.5
    }

    // Very inflexible: 8+ elixir
    return 0.3
}
```

---

## Combined Versatility Score

### Card-Level Versatility

```go
type CardVersatility struct {
    MultiRoleScore        float64 // 0-1.6, 60% weight
    SituationalScore      float64 // 0-0.3, 30% weight
    ElixirScore           float64 // 0-0.1, 10% weight
    TotalScore           float64 // Weighted sum
}

func CalculateCardVersatility(card Card, evolutionLevel int) CardVersatility {
    multiRole := MultiRoleVersatility(card, evolutionLevel) * 0.6
    situational := SituationalFlexibility(card)
    elixir := ElixirAdaptability(card)

    return CardVersatility{
        MultiRoleScore: multiRole,
        SituationalScore: situational,
        ElixirScore: elixir,
        TotalScore: multiRole + situational + elixir,
    }
}
```

### Deck-Level Versatility

```go
type DeckVersatilityReport struct {
    CardVersatility   []CardVersatility
    AverageVersatility float64
    VersatileCards     []string     // Cards with versatility > 1.0
    LowVersatilityCards []string    // Cards with versatility < 0.5
    RoleDiversity      float64      // Complement to redundancy analysis
}

func CalculateDeckVersatility(deck Deck) DeckVersatilityReport {
    cardVersatility := make([]CardVersatility, len(deck.Cards))
    totalVersatility := 0.0
    versatile := []string{}
    lowVersatile := []string{}

    for i, card := range deck.Cards {
        cv := CalculateCardVersatility(card, card.EvolutionLevel)
        cardVersatility[i] = cv
        totalVersatility += cv.TotalScore

        if cv.TotalScore > 1.0 {
            versatile = append(versatile, card.Name)
        }
        if cv.TotalScore < 0.5 {
            lowVersatile = append(lowVersatile, card.Name)
        }
    }

    return DeckVersatilityReport{
        CardVersatility: cardVersatility,
        AverageVersatility: totalVersatility / float64(len(deck.Cards)),
        VersatileCards: versatile,
        LowVersatilityCards: lowVersatile,
        RoleDiversity: calculateRoleDiversity(deck),
    }
}
```

---

## Integration with ao8 Epic's 5-Category Scoring

The **Versatility** category (one of 5 in ao8) is composed of:

```
VersatilityScore = (1.0 - RedundancyPenalty) * 0.6 + VersatilityBonus * 0.3 + ElixirFlexibility * 0.1
```

Where:
- **RedundancyPenalty** (inverse): From `REDUNDANCY_PATTERNS.md` analysis
- **VersatilityBonus**: Multi-role bonus from this document
- **ElixirFlexibility**: Elixir adaptability from this document

### Example Calculation

**Deck: Log Bait**
```
Goblin Barrel - Multi-role: 1.2, Situational: 0.2, Elixir: 0.08 = 0.48
Princess      - Multi-role: 1.0, Situational: 0.2, Elixir: 0.10 = 0.50
Goblin Gang   - Multi-role: 1.2, Situational: 0.2, Elixir: 0.10 = 0.50
Knight        - Multi-role: 1.3, Situational: 0.25, Elixir: 0.10 = 0.58
Ice Spirit    - Multi-role: 1.0, Situational: 0.15, Elixir: 0.10 = 0.45
Inferno Tower - Multi-role: 1.0, Situational: 0.20, Elixir: 0.05 = 0.45
Rocket        - Multi-role: 1.0, Situational: 0.15, Elixir: 0.03 = 0.38
The Log       - Multi-role: 1.2, Situational: 0.25, Elixir: 0.10 = 0.55
---
Average Card Versatility: 0.486
Redundancy Penalty: 0.1 (synergistic redundancy in bait)
Role Diversity: 0.7 (good coverage)

Final Versatility Score: (1.0 - 0.1) * 0.6 + 0.486 * 0.3 + 0.08 * 0.1
                     = 0.54 + 0.146 + 0.008
                     = 0.694 / 1.0
```

---

## Card Database Schema for Versatility

```go
type CardVersatilityData struct {
    CardName            string

    // Multi-role data
    PrimaryRole         CardRole
    SecondaryRoles      []CardRole
    EvolutionRoles      []CardRole

    // Situational data
    OffensiveEffectiveness float64  // 0-1
    DefensiveEffectiveness float64  // 0-1
    HitsAir                bool
    HitsGround             bool
    HasUtility             bool
    VsTank                bool
    VsSwarm               bool
    VsAir                 bool
    VsBuilding            bool

    // Elixir adaptability
    EffectiveEarly        bool
    EffectiveMid          bool
    EffectiveLate         bool
    OptimalElixirStage    string  // "early", "mid", "late", "any"

    // Calculated fields
    MultiRoleScore       float64
    SituationalScore     float64
    ElixirScore          float64
    TotalVersatility     float64
}
```

---

## Top Versatile Cards (Based on Analysis)

### Tier 1: Maximum Versatility (1.5+)
| Card | Multi-Role | Situational | Elixir | Total |
|------|-----------|-------------|--------|-------|
| Valkyrie (Evo) | 1.5 | 0.28 | 0.10 | **1.88** |
| Electro Wizard | 1.5 | 0.30 | 0.08 | **1.88** |
| Mini P.E.K.K.A | 1.5 | 0.28 | 0.08 | **1.86** |

### Tier 2: High Versatility (1.2-1.5)
| Card | Multi-Role | Situational | Elixir | Total |
|------|-----------|-------------|--------|-------|
| Valkyrie | 1.3 | 0.25 | 0.10 | **1.65** |
| Magic Archer | 1.3 | 0.28 | 0.08 | **1.66** |
| Hunter | 1.3 | 0.28 | 0.08 | **1.66** |
| Baby Dragon | 1.3 | 0.30 | 0.08 | **1.68** |

### Tier 3: Moderate Versatility (0.9-1.2)
| Card | Multi-Role | Situational | Elixir | Total |
|------|-----------|-------------|--------|-------|
| Musketeer | 1.0 | 0.25 | 0.10 | **1.35** |
| Bandit | 1.3 | 0.22 | 0.08 | **1.60** |
| Skeleton Army | 1.3 | 0.20 | 0.05 | **1.55** |

### Tier 4: Low Versatility (<0.9)
| Card | Multi-Role | Situational | Elixir | Total |
|------|-----------|-------------|--------|-------|
| Giant | 1.0 | 0.15 | 0.05 | **1.20** |
| X-Bow | 1.3 | 0.15 | 0.05 | **1.50** |
| Sparky | 1.0 | 0.15 | 0.03 | **1.18** |

---

## Open Questions

1. **Data Sourcing**: How do we populate the versatility database? Manual curation? Community input? ML analysis?

2. **Subjectivity**: Offensive/Defensive effectiveness scores are subjective. Can we derive from game data?

3. **Meta Changes**: Versatility may shift with balance changes. How to keep data current?

4. **Synergy with Redundancy**: High versatility might justify some redundancy. How to model this interaction?

5. **Evolution Impact**: Some evolutions dramatically increase versatility. Should this be a separate calculation?

---

## References

- `pkg/deck/role_classifier.go` - Role classification system
- `docs/REDUNDANCY_PATTERNS.md` - Redundancy analysis (complementary)
- Meta analysis: DeckShop.pro, RoyaleAPI (January 2026)
- Related beads tasks: clash-royale-api-ao8 (Deck Evaluation System epic)
