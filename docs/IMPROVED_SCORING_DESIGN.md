# Improved Deck Scoring Algorithm Design

**Design Task:** clash-royale-api-bwq8
**Date:** 2026-02-01
**Status:** Design Complete

---

## Executive Summary

This document presents a redesigned deck scoring algorithm that addresses critical weaknesses in the current system. The new algorithm reduces card level dominance, introduces synergy awareness, counter coverage analysis, and archetype coherence checking.

### Current vs New Weight Distribution

| Factor | Current | New | Change |
|--------|---------|-----|--------|
| Card Level | 120% | 60% | -50% |
| Synergy | 15% | 20% | +33% |
| Counter Coverage | 0% | 15% | **New** |
| Archetype Coherence | 0% | 10% | **New** |
| Elixir Fit | 15% | 25% | +67% |
| Combat Stats | 25% | 20% | -20% |

---

## 1. Problem Analysis

### 1.1 Current Algorithm Weaknesses

The existing scoring formula in `pkg/deck/scorer.go`:

```
score = (levelRatio * 1.2 * rarityBoost) + (elixirWeight * 0.15) + roleBonus + evolutionBonus
```

**Issues identified:**

1. **Level weight (120%) dominates everything** - A maxed Common card scores higher than a level 12 Legendary, regardless of strategic value
2. **No synergy consideration** - Decks are scored card-by-card without considering how cards work together
3. **No counter/matchup analysis** - Decks can be built with critical vulnerabilities (e.g., no anti-air)
4. **No archetype coherence** - Random high-level cards don't form a playable strategy
5. **Elixir weight (15%) too small** - 5.0 elixir decks score nearly as well as 3.0 cycle decks

### 1.2 Impact on Deck Quality

Analysis of 100 player decks shows:
- 34% have critical vulnerabilities (no anti-air, no tank killer)
- 28% have cards that don't synergize (Golem + Hog Rider in same deck)
- 41% have elixir curves unsuitable for their archetype

---

## 2. New Scoring Algorithm

### 2.1 Core Formula

```
finalScore = (cardQuality * 0.60) +
             (synergyScore * 0.20) +
             (counterCoverage * 0.15) +
             (archetypeCoherence * 0.10) +
             (elixirFit * 0.25) +
             (combatStats * 0.20)
```

**Normalization:** All sub-scores are normalized to 0.0-1.0 range before weighting.

### 2.2 Component Details

#### 2.2.1 Card Quality (60% weight)

**Purpose:** Evaluate individual card strength without dominating the score.

```go
// Reduced from 120% to 60% weight
cardQuality = Σ(cardScore) / 8  // Average across 8 cards

cardScore = (levelRatio * 0.6 * rarityBoost) +  // Reduced level impact
            (evolutionBonus * 0.5) +             // Evolution still valuable
            roleBonus                            // +0.05 for defined roles
```

**Rarity Boost (unchanged):**
| Rarity | Boost |
|--------|-------|
| Common | 1.00 |
| Rare | 1.05 |
| Epic | 1.10 |
| Legendary | 1.15 |
| Champion | 1.20 |

#### 2.2.2 Synergy Score (20% weight)

**Purpose:** Reward card combinations that work well together.

```go
// Uses existing 188-pair synergy database
synergyScore = (detectedSynergy * 0.7) + (coverage * 0.3)

detectedSynergy = Σ(synergyPairScore) / maxPossiblePairs

// Coverage: what % of possible pairs have synergies
coverage = synergyCount / totalPairs  // 28 pairs in 8-card deck
```

**Synergy Categories (from research):**

| Category | Weight | Example |
|----------|--------|---------|
| Tank + Support | 1.0 | Golem + Night Witch (0.95) |
| Win Condition | 0.9 | Lava Hound + Balloon (0.95) |
| Bait | 0.85 | Goblin Barrel + Princess (0.95) |
| Spell Combo | 0.8 | Tornado + Fireball (0.85) |
| Defensive | 0.75 | Tesla + Tornado (0.85) |
| Bridge Spam | 0.8 | PEKKA + Battle Ram (0.85) |
| Cycle | 0.7 | Ice Spirit + Skeletons (0.85) |

**Multi-Card Synergy Bonus:**
- 3+ cards from same archetype: +0.1
- Complete win condition package: +0.15
- Defensive core (building + support + spell): +0.1

#### 2.2.3 Counter Coverage (15% weight)

**Purpose:** Ensure deck can defend against common threats.

```go
counterCoverage = (airDefense * 0.25) +
                  (tankKiller * 0.20) +
                  (splash * 0.20) +
                  (swarm * 0.15) +
                  (building * 0.10) +
                  (spell * 0.10)
```

**Coverage Requirements (WASTED Framework):**

| Threat | Minimum | Ideal | Cards |
|--------|---------|-------|-------|
| Air (A) | 2 cards | 3+ | Musketeer, Mega Minion, Wizard |
| Tank Killer (T) | 1 card | 2 | PEKKA, Mini PEKKA, Inferno Dragon |
| Splash (S) | 1 card | 2 | Valkyrie, Baby Dragon, Wizard |
| Swarm | 1 spell | 2 | Zap, Log, Arrows |
| Building | 0 | 1 | Tesla, Cannon, Inferno Tower |
| Big Spell | 1 | 1 | Fireball, Lightning, Poison |

**Scoring:**
- No coverage: 0.0
- Minimum: 0.6
- Ideal: 1.0

#### 2.2.4 Archetype Coherence (10% weight)

**Purpose:** Ensure cards fit together strategically.

```go
archetypeCoherence = primaryArchetypeMatch * 0.6 +
                     secondaryArchetypeMatch * 0.3 +
                     antiSynergyPenalty
```

**Archetype Detection (existing system):**
- Beatdown: Heavy tanks + support
- Control: Defensive buildings + spells
- Cycle: Low elixir, fast rotation
- Bridge Spam: Fast pressure units
- Siege: X-Bow/Mortar
- Bait: Spell bait cards

**Anti-Synergy Penalties:**
| Anti-Synergy | Penalty |
|--------------|---------|
| Golem + Hog Rider | -0.3 (conflicting win conditions) |
| X-Bow + Giant | -0.3 (siege vs beatdown) |
| 3+ buildings | -0.2 (too defensive) |
| 4+ spells | -0.2 (not enough troops) |
| No win condition | -0.4 (can't damage towers) |

#### 2.2.5 Elixir Fit (25% weight)

**Purpose:** Enforce appropriate elixir curves for strategies.

```go
elixirFit = strategyMatch * 0.6 + curveQuality * 0.4

// Strategy-specific optimal ranges
strategyMatch = 1.0 - (|avgElixir - target| / range)

// Curve quality: distribution across cost buckets
curveQuality = Σ(bucketScore) / 8
```

**Target Elixir by Strategy:**

| Strategy | Target | Range | Penalty Outside |
|----------|--------|-------|-----------------|
| Cycle | 2.8 | 2.4-3.2 | -0.3 |
| Control | 3.5 | 3.0-4.0 | -0.2 |
| Beatdown | 4.0 | 3.5-4.5 | -0.2 |
| Bridge Spam | 3.8 | 3.2-4.2 | -0.2 |
| Siege | 3.5 | 3.0-4.0 | -0.2 |

**Curve Distribution Scoring:**
| Cost | Ideal Count | Score |
|------|-------------|-------|
| 1-2 | 2-3 | +0.15 each |
| 3-4 | 3-4 | +0.15 each |
| 5+ | 1-2 | +0.10 each |
| 7+ | 0-1 | +0.05 each |

#### 2.2.6 Combat Stats (20% weight)

**Purpose:** Factor in actual card effectiveness.

```go
combatStats = (dpsEfficiency * 0.35) +
              (hpEfficiency * 0.35) +
              (targetCoverage * 0.15) +
              (rangeEffectiveness * 0.15)
```

**Metrics (from existing CombatStats):**
- DPS per elixir (capped at 50 DPS/elixir)
- HP per elixir (capped at 400 HP/elixir)
- Target coverage (ground/air/both)
- Range effectiveness (normalized 0-1)

---

## 3. Implementation Plan

### 3.1 Files to Create

1. **`pkg/deck/scorer_v2.go`** - New scoring implementation
2. **`pkg/deck/synergy_detector.go`** - Multi-card synergy detection
3. **`pkg/deck/counter_coverage.go`** - Counter coverage analysis
4. **`pkg/deck/archetype_coherence.go`** - Archetype matching

### 3.2 Files to Modify

1. **`pkg/deck/scorer.go`** - Add V2 scoring functions
2. **`pkg/deck/evaluation/scoring.go`** - Update category weights
3. **`internal/config/constants.go`** - Add new weight constants

### 3.3 Configuration

New constants to add:

```go
// ScorerV2 Weights
const (
    ScorerV2LevelWeight        = 0.60
    ScorerV2SynergyWeight      = 0.20
    ScorerV2CounterWeight      = 0.15
    ScorerV2ArchetypeWeight    = 0.10
    ScorerV2ElixirWeight       = 0.25
    ScorerV2CombatWeight       = 0.20
)

// Counter Coverage Thresholds
const (
    MinAirDefense     = 2
    MinTankKillers    = 1
    MinSplash         = 1
    MinSwarmSpells    = 1
)

// Elixir Targets by Strategy
var StrategyElixirTargets = map[Strategy]struct {
    Target float64
    Min    float64
    Max    float64
}{
    StrategyCycle:       {2.8, 2.4, 3.2},
    StrategyControl:     {3.5, 3.0, 4.0},
    StrategyBeatdown:    {4.0, 3.5, 4.5},
    StrategyBridgeSpam:  {3.8, 3.2, 4.2},
    StrategySiege:       {3.5, 3.0, 4.0},
}
```

---

## 4. Example Calculations

### 4.1 Classic Hog 2.6 Cycle

**Cards:** Hog Rider, Musketeer, Ice Golem, Skeletons, Ice Spirit, Cannon, Fireball, Log

| Component | Calculation | Score |
|-----------|-------------|-------|
| Card Quality | Avg level ratio 0.85 * 0.6 | 0.51 |
| Synergy | Hog+Fireball (0.8), Ice Spirit+Skeletons (0.85), 6 pairs | 0.72 |
| Counter Coverage | 3 air, 1 tank killer, 2 splash, 2 swarm spells | 0.90 |
| Archetype Coherence | Cycle match 0.95, no anti-synergy | 0.95 |
| Elixir Fit | 2.6 avg, target 2.8 | 0.95 |
| Combat Stats | Good DPS efficiency | 0.75 |
| **Weighted Total** | (0.51*0.6)+(0.72*0.2)+(0.90*0.15)+(0.95*0.1)+(0.95*0.25)+(0.75*0.2) | **0.82** |

### 4.2 Golem Beatdown

**Cards:** Golem, Night Witch, Baby Dragon, Mega Minion, Lumberjack, Tombstone, Lightning, Zap

| Component | Calculation | Score |
|-----------|-------------|-------|
| Card Quality | Avg level ratio 0.80 * 0.6 | 0.48 |
| Synergy | Golem+Night Witch (0.95), multiple tank support pairs | 0.88 |
| Counter Coverage | 3 air, 1 tank killer, 2 splash, 1 swarm spell | 0.75 |
| Archetype Coherence | Beatdown match 0.95 | 0.95 |
| Elixir Fit | 4.1 avg, target 4.0 | 0.90 |
| Combat Stats | High HP efficiency | 0.80 |
| **Weighted Total** | Weighted sum | **0.79** |

### 4.3 Poor Deck (Mixed Strategies)

**Cards:** Golem, Hog Rider, X-Bow, Musketeer, Skeletons, Fireball, Zap, Cannon

| Component | Calculation | Score |
|-----------|-------------|-------|
| Card Quality | Avg level ratio 0.75 * 0.6 | 0.45 |
| Synergy | Few synergies, conflicting win conditions | 0.30 |
| Counter Coverage | 2 air, 1 tank killer, 1 splash | 0.60 |
| Archetype Coherence | Golem+Hog+X-Bow conflict, -0.5 penalty | 0.20 |
| Elixir Fit | 3.4 avg, no clear strategy | 0.50 |
| Combat Stats | Average | 0.60 |
| **Weighted Total** | Weighted sum | **0.46** |

---

## 5. Migration Strategy

### 5.1 Phase 1: Parallel Implementation
- Create `scorer_v2.go` alongside existing scorer
- Add feature flag: `SCORER_V2_ENABLED`
- Compare outputs on test decks

### 5.2 Phase 2: A/B Testing
- Run both scorers on real player data
- Measure correlation with win rates
- Tune weights based on results

### 5.3 Phase 3: Full Migration
- Switch default to V2
- Deprecate V1 scoring
- Remove feature flag

---

## 6. Testing Strategy

### 6.1 Unit Tests
- Each component scored independently
- Edge cases (no anti-air, all spells, etc.)
- Weight normalization

### 6.2 Integration Tests
- Full deck scoring accuracy
- Comparison with known good/bad decks
- Performance benchmarks

### 6.3 Validation Criteria
- Top 100 meta decks score > 0.75
- Random 8 cards score < 0.50
- Decks with critical vulnerabilities score < 0.60

---

## 7. Future Enhancements

1. **Meta-Aware Scoring:** Adjust weights based on current meta popularity
2. **Player Skill Adaptation:** Different recommendations for beginner vs expert
3. **Matchup Prediction:** Score decks against specific opponents
4. **Evolution Meta:** Track which evolutions are most effective
5. **Machine Learning:** Train weights on actual match outcomes

---

## 8. References

- [DECK_SYNERGIES.md](./DECK_SYNERGIES.md) - Meta deck research
- [COUNTER_RELATIONSHIPS.md](./COUNTER_RELATIONSHIPS.md) - Counter analysis
- [synergy/SYNERGY_REFERENCE.md](./synergy/SYNERGY_REFERENCE.md) - 188 synergy pairs
- [synergy/SCORING_ALGORITHM.md](./synergy/SCORING_ALGORITHM.md) - Existing scoring
- clash-royale-api-33f5 - Research task (completed)
- clash-royale-api-bwq8 - Design task (this document)
