# Integration Design: Redundancy & Versatility Scoring

**Research Spike: clash-royale-api-5xb**
**Date**: 2026-01-01

---

## Overview

This document describes how to integrate redundancy penalties and versatility bonuses into the existing Clash Royale API codebase, specifically for the **ao8 epic** (Deck Evaluation System).

---

## Architecture Integration

### Current Architecture

The existing codebase has:
- `pkg/deck/role_classifier.go` - Role classification (6 roles)
- `pkg/deck/synergy.go` - Synergy database (188 pairs)
- `pkg/deck/builder.go` - Deck builder with scoring
- `pkg/deck/scorer.go` - Card scoring algorithms
- `pkg/deck/strategy_config.go` - Strategy configuration (JSON-based)
- `pkg/scoring/interfaces.go` - Scorer interface

### New Components

| Component | File | Purpose |
|-----------|------|---------|
| RedundancyScorer | `pkg/deck/redundancy_scorer.go` | Calculate redundancy penalties |
| VersatilityScorer | `pkg/deck/versatility_scorer.go` | Calculate versatility bonuses |
| CardVersatilityDatabase | `data/card_versatility.json` | Pre-computed versatility data |
| RedundancyAnalyzer | `pkg/deck/redundancy_analyzer.go` | Combined analysis |

---

## Integration Points

### 1. Extend `StrategyConfig` (`pkg/deck/strategy_config.go`)

```go
// StrategyConfig defines deck-building strategy parameters
type StrategyConfig struct {
    // ... existing fields ...

    // Redundancy tolerance per archetype (dynamic thresholds)
    RedundancyTolerance map[Archetype]RoleTolerance `json:"redundancy_tolerance"`

    // Versatility scoring weights (sum to 1.0)
    VersatilityWeights struct {
        MultiRole   float64 `json:"multi_role"`    // Default: 0.6
        Situational float64 `json:"situational"`   // Default: 0.3
        ElixirFlex  float64 `json:"elixir_flex"`   // Default: 0.1
    } `json:"versatility_weights"`
}

// RoleTolerance defines acceptable card counts per role
type RoleTolerance struct {
    WinCondition int `json:"win_condition"`
    Building     int `json:"building"`
    SpellBig     int `json:"spell_big"`
    SpellSmall   int `json:"spell_small"`
    Support      int `json:"support"`
    Cycle        int `json:"cycle"`
}
```

**JSON Example** (`config/strategies.json`):
```json
{
  "strategies": {
    "balanced": {
      "name": "Balanced",
      "redundancy_tolerance": {
        "win_condition": 2,
        "building": 1,
        "spell_big": 2,
        "spell_small": 3,
        "support": 4,
        "cycle": 3
      },
      "versatility_weights": {
        "multi_role": 0.6,
        "situational": 0.3,
        "elixir_flex": 0.1
      }
    },
    "beatdown": {
      "name": "Beatdown",
      "redundancy_tolerance": {
        "win_condition": 2,
        "building": 1,
        "spell_big": 2,
        "spell_small": 3,
        "support": 4,
        "cycle": 2
      },
      "versatility_weights": {
        "multi_role": 0.5,  // Beatdown values focused roles over versatility
        "situational": 0.3,
        "elixir_flex": 0.2
      }
    }
  }
}
```

### 2. Extend `Scorer` Interface (`pkg/scoring/interfaces.go`)

```go
// Scorer calculates scores for cards and decks
type Scorer interface {
    ScoreCard(card Card, deck Deck, config ScoringConfig) float64
}

// RedundancyScorer implements Scorer for redundancy penalties
type RedundancyScorer interface {
    Scorer
    AnalyzeRedundancy(deck Deck, archetype Archetype) (*RedundancyReport, error)
}

// VersatilityScorer implements Scorer for versatility bonuses
type VersatilityScorer interface {
    Scorer
    CalculateCardVersatility(card Card) (*CardVersatilityScore, error)
    CalculateDeckVersatility(deck Deck) (*DeckVersatilityReport, error)
}
```

### 3. Create `DeckEvaluation` Type (`pkg/deck/evaluation/types.go`)

```go
// DeckEvaluation provides comprehensive deck analysis for the ao8 epic
type DeckEvaluation struct {
    // Basic info
    Deck      Deck
    Archetype Archetype

    // 5-category scores (ao8 epic)
    AttackScore      float64 // 0-10
    DefenseScore     float64 // 0-10
    SynergyScore     float64 // 0-10 (existing from pkg/deck/synergy.go)
    VersatilityScore float64 // 0-10 (NEW - from this research)
    F2PScore         float64 // 0-10

    // Detailed analysis
    RedundancyAnalysis *RedundancyReport
    VersatilityReport  *DeckVersatilityReport
}

// EvaluateDeck performs comprehensive deck evaluation
func EvaluateDeck(deck Deck) (*DeckEvaluation, error) {
    // Detect archetype
    archetype := DetectArchetype(deck)

    // Calculate redundancy
    redundancyScorer := NewRedundancyScorer()
    redundancyReport, _ := redundancyScorer.AnalyzeRedundancy(deck, archetype)

    // Calculate versatility
    versatilityScorer := NewVersatilityScorer()
    versatilityReport, _ := versatilityScorer.CalculateDeckVersatility(deck)

    // Calculate 5-category scores
    attackScore := calculateAttackScore(deck)
    defenseScore := calculateDefenseScore(deck)
    synergyScore := calculateSynergyScore(deck) // Existing
    versatilityScore := calculateVersatilityForAo8(deck, redundancyReport.OverallPenalty)
    f2pScore := calculateF2PScore(deck)

    return &DeckEvaluation{
        Deck:               deck,
        Archetype:          archetype,
        AttackScore:        attackScore,
        DefenseScore:       defenseScore,
        SynergyScore:       synergyScore,
        VersatilityScore:   versatilityScore,
        F2PScore:          f2pScore,
        RedundancyAnalysis: redundancyReport,
        VersatilityReport:  versatilityReport,
    }, nil
}
```

---

## Mapping to ao8's 5-Category System

### The 5 Categories

The ao8 epic defines 5 scoring categories for DeckShop.pro-style evaluation:

| Category | Description | Weight | Source |
|----------|-------------|--------|--------|
| **Attack** | Win condition strength, offensive pressure | - | Existing combat stats |
| **Defense** | Defensive capabilities, counter potential | - | Existing combat stats |
| **Synergy** | Card pair interactions | - | Existing `pkg/deck/synergy.go` |
| **Versatility** | Role diversity, multi-role cards, flexibility | - | **NEW (this research)** |
| **F2P Score** | Free-to-play accessibility, card availability | - | Existing rarity data |

### Versatility Category Composition

```
VersatilityScore (0-10) = (1.0 - redundancyPenalty) * 6.0 + versatilityBonus * 3.0 + elixirFlexibility * 1.0
```

**Components:**
- **60%** Role Diversity (inverse of redundancy penalty)
- **30%** Multi-role Cards (versatility bonus)
- **10%** Elixir Flexibility

**Scaling:**
- 0.0-1.0 internal score → 0-10 star rating
- 0-3 stars: Poor versatility
- 4-6 stars: Average versatility
- 7-8 stars: Good versatility
- 9-10 stars: Excellent versatility

### Example Calculation

**Deck: Log Bait**
```
Cards: Goblin Barrel, Princess, Goblin Gang, Knight, Ice Spirit,
       Inferno Tower, Rocket, The Log

Redundancy Penalty: 0.1 (synergistic redundancy in bait)
Versatility Bonus: 0.486 (average card versatility)
Elixir Flexibility: 0.08

VersatilityScore = (1.0 - 0.1) * 6.0 + 0.486 * 3.0 + 0.08 * 1.0
                 = 5.4 + 1.458 + 0.08
                 = 6.938 / 10.0
                 = 6.9 stars (Good versatility)
```

---

## Data Structures

### Card Versatility Database (`data/card_versatility.json`)

```json
{
  "version": "1.0",
  "last_updated": "2026-01-01",
  "cards": {
    "Valkyrie": {
      "primary_role": "support",
      "secondary_roles": ["support"],
      "evolution_roles": ["support"],
      "multi_role_score": 1.3,
      "offensive_effectiveness": 0.8,
      "defensive_effectiveness": 0.9,
      "hits_air": false,
      "hits_ground": true,
      "has_utility": true,
      "vs_tank": true,
      "vs_swarm": true,
      "vs_air": false,
      "vs_building": true,
      "situational_score": 0.25,
      "effective_early": true,
      "effective_mid": true,
      "effective_late": true,
      "elixir_score": 0.10,
      "total_versatility": 1.65
    },
    "Electro Wizard": {
      "primary_role": "support",
      "secondary_roles": ["support"],
      "evolution_roles": [],
      "multi_role_score": 1.5,
      "offensive_effectiveness": 0.8,
      "defensive_effectiveness": 0.8,
      "hits_air": true,
      "hits_ground": true,
      "has_utility": true,
      "vs_tank": false,
      "vs_swarm": false,
      "vs_air": true,
      "vs_building": true,
      "situational_score": 0.30,
      "effective_early": true,
      "effective_mid": true,
      "effective_late": true,
      "elixir_score": 0.08,
      "total_versatility": 1.88
    }
  }
}
```

### Archetype Pattern Database (`data/archetype_patterns.json`)

```json
{
  "archetypes": {
    "beatdown": {
      "win_condition_range": [1, 2],
      "avg_elixir_range": [3.8, 4.5],
      "redundancy_multiplier": 1.0,
      "versatility_weights": {
        "multi_role": 0.5,
        "situational": 0.3,
        "elixir_flex": 0.2
      }
    },
    "bait": {
      "win_condition_range": [1, 2],
      "avg_elixir_range": [3.0, 3.5],
      "redundancy_multiplier": 0.3,
      "versatility_weights": {
        "multi_role": 0.6,
        "situational": 0.3,
        "elixir_flex": 0.1
      }
    }
  }
}
```

---

## CLI Integration

### New Command: `./bin/cr-api deck evaluate`

```bash
# Evaluate a deck string
./bin/cr-api deck evaluate "giant;witch;electrowizard;musketeer;fireball;zap;tombstone;icespirit"

# Evaluate from player data
./bin/cr-api deck evaluate --player TAG

# Output formats
./bin/cr-api deck evaluate "DECK" --format json
./bin/cr-api deck evaluate "DECK" --format csv
./bin/cr-api deck evaluate "DECK" --format human
```

### Example Output (Human)

```
Deck Evaluation
═══════════════

Deck: Giant, Witch, Electro Wizard, Musketeer, Fireball, Zap, Tombstone, Ice Spirit
Archetype: Beatdown

Category Scores (0-10 stars)
────────────────────────────
Attack:      ██████████  7.2 ★★★★★★★★
Defense:     ████████░░  6.5 ★★★★★★★
Synergy:     ████████░░  6.8 ★★★★★★★
Versatility: ████████░░  6.4 ★★★★★★★
F2P Score:   ██████████  8.5 ★★★★★★★★

Redundancy Analysis
──────────────────
✓ No redundancy issues detected

Versatility Analysis
────────────────────
Versatile Cards: Electro Wizard (1.88), Valkyrie (1.65)
Low Versatility: Giant (1.20)

Recommendations
───────────────
• Consider swapping Giant for a more versatile win condition
• Electro Wizard provides excellent air defense + stun utility
```

---

## Performance Considerations

### Caching Strategy

1. **Card Versatility Data**: Load once at startup (in-memory cache)
2. **Synergy Calculations**: Already memoized in `pkg/deck/synergy.go`
3. **Archetype Detection**: O(n) - acceptable for 8 cards
4. **Redundancy Analysis**: O(n) - acceptable for 8 cards

### Performance Targets

- Single deck evaluation: < 100ms
- Batch evaluation (100 decks): < 5 seconds
- Memory overhead: < 50MB for card database

---

## Testing Strategy

### Unit Tests

- `TestRedundancyScorer_*`: Test redundancy detection for each archetype
- `TestVersatilityScorer_*`: Test versatility calculations
- `TestArchetypeDetection_*`: Test archetype identification
- `TestAo8Scoring_*`: Test 5-category score calculation

### Integration Tests

- `TestDeckEvaluation_*`: Test full evaluation pipeline
- `TestCLICommands_*`: Test `deck evaluate` command

### Validation Tests

- `TestMetaDeckAnalysis_*`: Validate against real meta decks
- `TestVersatilityCorrelation_*`: Correlate versatility scores with win rates

---

## Migration Path

### Phase 1: Research Prototypes (Current)
- `pkg/deck/redundancy_scorer.go` - Research skeleton
- `pkg/deck/versatility_scorer.go` - Research skeleton
- Documentation only

### Phase 2: Production Implementation (Future - ao8 epic)
- Refactor prototypes into production code
- Populate card versatility database
- Add to `pkg/deck/evaluation/` package
- Integrate with deck builder

### Phase 3: CLI Integration (Future - ao8 epic)
- Add `deck evaluate` command
- Add `--format` flags for output
- Integrate with player data

---

## Open Questions

1. **Database Population**: How to initially populate the card versatility database?
   - Option A: Manual curation (time-intensive but accurate)
   - Option B: Community crowdsourcing (scalable but needs validation)
   - Option C: ML analysis (requires training data)

2. **Meta Updates**: How often to update versatility scores based on balance changes?
   - Recommendation: Quarterly reviews, with hotfixes for major balance changes

3. **Archetype Confidence**: What if archetype detection is uncertain?
   - Recommendation: Apply weighted average of multiple archetype tolerances

4. **Subjectivity**: How to handle subjective effectiveness scores?
   - Recommendation: Derive from game statistics where possible, use expert curation where not

---

## References

- `docs/REDUNDANCY_PATTERNS.md` - Redundancy patterns and thresholds
- `docs/VERSATILITY_METRICS.md` - Versatility metrics and formulas
- `pkg/deck/role_classifier.go` - Existing role classification
- `pkg/deck/synergy.go` - Existing synergy system (pattern reference)
- `pkg/deck/strategy_config.go` - Existing strategy configuration
- Related beads tasks: clash-royale-api-ao8 (Deck Evaluation System epic)
