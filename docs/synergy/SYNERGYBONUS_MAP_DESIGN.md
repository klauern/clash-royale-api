# SynergyBonus Map Design

## Overview

The current synergy system provides raw synergy scores (0.0-1.0) for card pairs. For deck evaluation, we need to extend this with **categorized synergy ratings** that separate synergies by their strategic role (attack, defense, support).

## Current State

### Existing Implementation

```go
// Current: Raw synergy score only
func (db *SynergyDatabase) GetSynergy(card1, card2 string) float64

// Usage in builder
candidate.Score += calculateSynergyScore(candidate.Name, currentDeck) * synergyWeight

// Result: Single score, no categorization
```

**Limitation**: The synergy score alone doesn't indicate whether the synergy helps offensively (attack), defensively, or provides utility/counter capabilities.

## Proposed Enhancement: synergyBonus Structure

### Design Requirements

Based on the task description and deck evaluation needs, we need:

1. **Per-pair categorization**: Each synergy pair gets ratings for attack/defense/support
2. **Strategic context**: Ratings help evaluate deck's offensive/defensive capabilities
3. **Integration ready**: Easy to consume by evaluation scoring engine
4. **Backward compatible**: Doesn't break existing synergy system

### synergyBonus Map Structure

```go
type SynergyBonus struct {
    Card1       string
    Card2       string
    SynergyType SynergyCategory  // Existing
    Score       float64         // Existing: 0.0-1.0

    // NEW: Categorized ratings
    AttackRating      float64  // 0.0-1.0: Offensive contribution
    DefenseRating     float64  // 0.0-1.0: Defensive contribution
    SupportRating     float64  // 0.0-1.0: Utility/counter contribution
}
```

### Bonus Map Storage

```go
type SynergyDatabase struct {
    Pairs          []SynergyPair                // Existing
    Categories     map[SynergyCategory][]SynergyPair  // Existing
    BonusRatings   map[string]map[string]*SynergyBonus  // NEW: card1 -> card2 -> Bonus
}
```

**Key Structure**: Nested map for O(1) lookups in both directions
- `bonusRatings["Giant"]["Witch"] = &SynergyBonus{...}`
- `bonusRatings["Witch"]["Giant"] = &SynergyBonus{...}` (same object)

## Categorization Framework: Attack vs Defense vs Support

### Rating Definitions

For each synergy pair, we assign three ratings that sum to the total synergy score:

#### Attack Rating (0.0-1.0)
Measures offensive contribution - how much this pair helps take towers or apply pressure.

**High Attack Examples**:
- Giant + Witch: `0.65/0.15/0.10` (Powerful push)
- Lava Hound + Balloon: `0.85/0.05/0.05` (Overwhelming air)
- Hog Rider + Fireball: `0.70/0.05/0.25` (Spell cycle)

**Medium Attack Examples**:
- X-Bow + Tesla: `0.40/0.45/0.15` (Siege + defense)
- Miner + Poison: `0.55/0.10/0.35` (Chip damage)

**Low Attack Examples**:
- Cannon + Ice Spirit: `0.05/0.80/0.15` (Pure defense)
- Tesla + Tornado: `0.10/0.75/0.15` (Defensive combo)

#### Defense Rating (0.0-1.0)
Measures defensive contribution - how well this pair stops enemy pushes.

**High Defense Examples**:
- Inferno Tower + Zap: `0.05/0.85/0.10` (Tank killer + reset)
- Tornado + Executioner: `0.15/0.70/0.15` (Area denial)
- Mega Knight + Bats: `0.20/0.65/0.15` (Counter-push)

**Medium Defense Examples**:
- PEKKA + Electro Wizard: `0.40/0.45/0.15` (Offensive + defensive)
- Giant + Musketeer: `0.50/0.30/0.20` (Push + air defense)

#### Support Rating (0.0-1.0)
Measures utility/counter contribution - cycle ability, spell synergy, versatility.

**High Support Examples**:
- Goblin Barrel + Princess: `0.25/0.20/0.55` (Bait pressure)
- Ice Spirit + Log: `0.15/0.25/0.60` (Cycle + control)
- Three Musketeers + Battle Ram: `0.40/0.20/0.40` (Split pressure)

### Rating Distribution Guidelines

**Total Score Distribution**: All three ratings sum to the total synergy score.

```
SynergyPair.Score = AttackRating + DefenseRating + SupportRating
```

**Typical Distributions by Category**:

#### Tank Support Category
```
Attack:    0.50-0.70 (Primary - enables pushes)
Defense:   0.20-0.30 (Secondary - defensive utility)
Support:   0.10-0.20 (Tertiary - minor utility)
```

Example: Giant + Witch
- Attack: 0.65 (Witch enables Giant push)
- Defense: 0.15 (Can defend after counter-push)
- Support: 0.10 (Minor cycle value)
- Total: 0.90

#### Bait Category
```
Attack:    0.20-0.40 (Enables attacks when spells are baited)
Defense:   0.15-0.25 (Some defensive utility)
Support:   0.40-0.60 (Primary - spell forcing)
```

Example: Goblin Barrel + Princess
- Attack: 0.25 (Can deal tower damage)
- Defense: 0.20 (Princess defense)
- Support: 0.50 (Forces Log/Fireball usage)
- Total: 0.95

#### Defensive Category
```
Attack:    0.05-0.15 (Minimal offensive contribution)
Defense:   0.70-0.85 (Primary - stops pushes)
Support:   0.10-0.20 (Some utility)
```

Example: Cannon + Ice Spirit
- Attack: 0.05 (Can be used in counter-push)
- Defense: 0.80 (Excellent defensive combo)
- Support: 0.15 (Cheap cycle value)
- Total: 0.80

#### Cycle Category
```
Attack:    0.10-0.20 (Minimal)
Defense:   0.15-0.25 (Moderate)
Support:   0.60-0.75 (Primary - enables faster cycles)
```

Example: Ice Spirit + Skeletons
- Attack: 0.15 (Can support pushes)
- Defense: 0.20 (Kiting, distraction)
- Support: 0.75 (Ultra-cheap cycle)
- Total: 0.85

#### Spell Combo Category
```
Attack:    0.60-0.80 (Primary - spell damage)
Defense:   0.10-0.20 (Some defensive utility)
Support:   0.10-0.20 (Minor utility)
```

Example: Tornado + Fireball
- Attack: 0.70 (Crown tower damage)
- Defense: 0.15 (Can defend swarms)
- Support: 0.15 (Versatility)
- Total: 0.85

## Implementation Example: Categorizing 188 Pairs

### Sample Pair Analysis

Let's categorize some representative pairs:

#### 1. Giant + Witch (Tank Support, Score: 0.90)
```
Rationale:
- Attack: 0.65 (Witch provides splash support for Giant push, enables tower damage)
- Defense: 0.15 (Witch can defend swarms, minimal overall defensive contribution)
- Support: 0.10 (Some cycle value, but primarily offensive)

Final: {Attack: 0.65, Defense: 0.15, Support: 0.10, Total: 0.90}
```

#### 2. Goblin Barrel + Princess (Bait, Score: 0.95)
```
Rationale:
- Attack: 0.25 (Both can deal tower damage when spells are baited)
- Defense: 0.20 (Princess provides defensive splash)
- Support: 0.50 (Primary purpose is forcing Log/Fireball usage)

Final: {Attack: 0.25, Defense: 0.20, Support: 0.50, Total: 0.95}
```

#### 3. Cannon + Ice Spirit (Defensive, Score: 0.80)
```
Rationale:
- Attack: 0.05 (Rarely used offensively)
- Defense: 0.80 (Excellent at stopping ground pushes)
- Support: 0.15 (Cheap cycle, kiting utility)

Final: {Attack: 0.05, Defense: 0.80, Support: 0.15, Total: 0.80}
```

#### 4. Lava Hound + Balloon (Win Condition, Score: 0.95)
```
Rationale:
- Attack: 0.85 (Overwhelming air push, primary win condition)
- Defense: 0.05 (Minimal defensive utility)
- Support: 0.05 (Minor cycle value)

Final: {Attack: 0.85, Defense: 0.05, Support: 0.05, Total: 0.95}
```

#### 5. Hog Rider + Fireball (Spell Combo, Score: 0.80)
```
Rationale:
- Attack: 0.70 (Fireball clears defenders, enables Hog connections)
- Defense: 0.05 (Can defend swarms if desperate)
- Support: 0.25 (Versatile spell, tower damage)

Final: {Attack: 0.70, Defense: 0.05, Support: 0.25, Total: 0.80}
```

### Bulk Categorization Guidelines

For efficiently categorizing all 188 pairs, use these heuristics:

#### Rule 1: By Category (Default Distribution)

```go
categoryRatings := map[SynergyCategory]RatingDist{
    SynergyTankSupport:  {Attack: 0.65, Defense: 0.20, Support: 0.15},
    SynergyBait:         {Attack: 0.25, Defense: 0.20, Support: 0.55},
    SynergySpellCombo:   {Attack: 0.70, Defense: 0.10, Support: 0.20},
    SynergyWinCondition: {Attack: 0.75, Defense: 0.10, Support: 0.15},
    SynergyDefensive:    {Attack: 0.05, Defense: 0.80, Support: 0.15},
    SynergyCycle:        {Attack: 0.15, Defense: 0.20, Support: 0.65},
    SynergyBridgeSpam:   {Attack: 0.70, Defense: 0.15, Support: 0.15},
}
```

#### Rule 2: Adjust for Specific Card Roles

**If card is pure win condition**: Increase Attack by 0.05-0.10
**If card is pure defense**: Increase Defense by 0.05-0.10
**If card is cycle/utility**: Increase Support by 0.05-0.10

Example: Fireball is a spell (dab utility), but also attack
- Base (SpellCombo): 0.70/0.10/0.20
- Fireball adjusts: 0.70/0.05/0.25 (more utility than pure attack)

#### Rule 3: Cross-Category Pairs

Some pairs fit multiple categories - use weighted averages:

Example: Princess + Goblin Barrel
- Bait base: 0.25/0.20/0.55
- Spell combo: 0.70/0.10/0.20
- Combined: 0.40/0.15/0.45
- Final (score 0.95): Normalize to 0.35/0.15/0.50

## Database Migration Strategy

### Option 1: Extend Existing Structure (Recommended)

Add new fields to existing pairs:

```go
type SynergyPair struct {
    Card1       string          `json:"card1"`
    Card2       string          `json:"card2"`
    SynergyType SynergyCategory `json:"synergy_type"`
    Score       float64         `json:"score"`
    Description string          `json:"description"`

    // NEW: Categorized ratings
    AttackRating  float64 `json:"attack_rating"`
    DefenseRating float64 `json:"defense_rating"`
    SupportRating float64 `json:"support_rating"`
}
```

**Pros**:
- Single source of truth
- Atomic updates
- Simpler queries

**Cons**:
- Larger memory footprint
- Schema migration needed

**Migration**:
```go
// Backfill example
for _, pair := range existingPairs {
    pair.AttackRating, pair.DefenseRating, pair.SupportRating =
        categorizePair(pair)
}
```

### Option 2: Separate Bonus Database

Keep raw scores in SynergyPair, add separate bonus map:

```go
type SynergyBonusDatabase struct {
    Ratings map[string]map[string]*SynergyBonus
}

type SynergyBonus struct {
    AttackRating  float64
    DefenseRating float64
    SupportRating float64
}
```

**Pros**:
- Backward compatible
- Can be loaded lazily
- Smaller base memory footprint

**Cons**:
- Two sources to maintain
- Join operations needed

**Migration**:
```go
// Load bonuses from separate file
bonuses := loadSynergyBonusesFromFile("synergy_bonuses.json")
for card1, card2Map := range bonuses {
    for card2, bonus := range card2Map {
        db.BonusRatings[card1][card2] = bonus
    }
}
```

### Recommended: Hybrid Approach

Use Option 1 (extend `SynergyPair`) for new system, provide backward compatibility:

```go
// New API for evaluation
func (db *SynergyDatabase) GetSynergyBonus(card1, card2 string) *SynergyBonus

// Legacy API (still works)
func (db *SynergyDatabase) GetSynergy(card1, card2 string) float64 {
    bonus := db.GetSynergyBonus(card1, card2)
    return bonus.Score()
}

// Helper on SynergyBonus
func (b *SynergyBonus) Score() float64 {
    return b.AttackRating + b.DefenseRating + b.SupportRating
}
```

## API Design for Evaluation Integration

### New Methods for Bonus Access

```go
// Get complete bonus information
type SynergyDatabase struct {
    // ... existing fields ...
}

func (db *SynergyDatabase) GetSynergyBonus(card1, card2 string) *SynergyBonus {
    // Order keys for consistent lookup
    first, second := card1, card2
    if card2 < card1 {
        first, second = card2, card1
    }

    if bonusMap, exists := db.BonusRatings[first]; exists {
        if bonus, exists := bonusMap[second]; exists {
            return bonus
        }
    }

    // Return zero bonus if not found
    return &SynergyBonus{
        AttackRating:  0.0,
        DefenseRating: 0.0,
        SupportRating: 0.0,
    }
}

// Batch operation for deck analysis
func (db *SynergyDatabase) GetDeckSynergyBonuses(deck []string) []SynergyBonus {
    bonuses := make([]SynergyBonus, 0)

    for i := 0; i < len(deck); i++ {
        for j := i + 1; j < len(deck); j++ {
            if bonus := db.GetSynergyBonus(deck[i], deck[j]); bonus.TotalScore() > 0 {
                bonuses = append(bonuses, *bonus)
            }
        }
    }

    return bonuses
}
```

### Object-Oriented Bonus Structure

```go
type SynergyBonus struct {
    Card1       string
    Card2       string
    Ratings     RatingBreakdown
}

type RatingBreakdown struct {
    Attack  float64
    Defense float64
    Support float64
}

// Convenience methods
func (rb RatingBreakdown) Total() float64 {
    return rb.Attack + rb.Defense + rb.Support
}

func (rb RatingBreakdown) AttackPercent() float64 {
    return rb.Attack / rb.Total() * 100
}

func (rb RatingBreakdown) DefensePercent() float64 {
    return rb.Defense / rb.Total() * 100
}

func (rb RatingBreakdown) SupportPercent() float64 {
    return rb.Support / rb.Total() * 100
}
```

## Integration with Evaluation Scoring

### Deck Evaluation Example

```go
// In evaluation/scorer.go

type DeckEvaluation struct {
    AttackScore    float64
    DefenseScore   float64
    SynergyScore   float64
    Versatility    float64
    Overall        float64
}

func (s *SynergyScorer) Score(deck []string) *DeckEvaluation {
    bonuses := s.synergyDB.GetDeckSynergyBonuses(deck)

    // Sum up all ratings
    totalAttack := 0.0
    totalDefense := 0.0
    totalSupport := 0.0
    count := 0

    for _, bonus := range bonuses {
        totalAttack += bonus.Ratings.Attack
        totalDefense += bonus.Ratings.Defense
        totalSupport += bonus.Ratings.Support
        count++
    }

    // Normalize to 0-100 scale
    maxPossible := float64(len(deck)*7) * 3.0  // Approximate

    eval := &DeckEvaluation{
        AttackScore:  (totalAttack / maxPossible) * 100,
        DefenseScore: (totalDefense / maxPossible) * 100,
        SynergyScore: (totalSupport / maxPossible) * 100,  // Support contributes to synergy
    }

    // Weighted average for overall (per design spec)
    eval.Overall = (eval.AttackScore*0.25 +
                     eval.DefenseScore*0.25 +
                     eval.SynergyScore*0.20 +
                     eval.Versatility*0.20 +
                     eval.F2P*0.10)

    return eval
}
```

### Detailed Synergy Report Generation

```go
func GenerateSynergyReport(deck []string, db *SynergyDatabase) string {
    bonuses := db.GetDeckSynergyBonuses(deck)

    // Aggregate by category
    categoryTotals := make(map[SynergyCategory]*RatingBreakdown)

    for _, bonus := range bonuses {
        cat := bonus.SynergyType
        if _, exists := categoryTotals[cat]; !exists {
            categoryTotals[cat] = &RatingBreakdown{}
        }

        categoryTotals[cat].Attack += bonus.Ratings.Attack
        categoryTotals[cat].Defense += bonus.Ratings.Defense
        categoryTotals[cat].Support += bonus.Ratings.Support
    }

    // Generate report
    var report strings.Builder
    report.WriteString("Synergy Analysis Report\n")
    report.WriteString("======================\n\n")

    for cat, totals := range categoryTotals {
        total := totals.Total()
        report.WriteString(fmt.Sprintf("%s: %.2f total\n", cat, total))
        report.WriteString(fmt.Sprintf("  Attack: %.1f%%\n", totals.AttackPercent()))
        report.WriteString(fmt.Sprintf("  Defense: %.1f%%\n", totals.DefensePercent()))
        report.WriteString(fmt.Sprintf("  Support: %.1f%%\n", totals.SupportPercent()))
        report.WriteString("\n")
    }

    return report.String()
}
```

## Migration Path

### Phase 1: Database Extension (Week 1)
- [ ] Add `AttackRating`, `DefenseRating`, `SupportRating` to `SynergyPair`
- [ ] Update JSON serialization
- [ ] Create categorization function
- [ ] Backfill 188 pairs with automatic categorization
- [ ] Test backward compatibility

### Phase 2: API Implementation (Week 2)
- [ ] Implement `GetSynergyBonus()` method
- [ ] Add deck-level analysis functions
- [ ] Create rating aggregation utilities
- [ ] Write comprehensive tests
- [ ] Benchmark performance

### Phase 3: Integration (Week 3)
- [ ] Integrate with evaluation package
- [ ] Implement scoring functions using categorized ratings
- [ ] Generate sample reports
- [ ] Validate against expert evaluations
- [ ] Tune rating distributions

### Phase 4: Production Ready (Week 4)
- [ ] Update all 188 pairs with manually reviewed ratings
- [ ] Add validation to ensure ratings sum correctly
- [ ] Create rating adjustment tools
- [ ] Document rating guidelines
- [ ] Deploy to production

## Benefits of This Design

### For Deck Evaluation

1. **Granular Scoring**: Separate attack/defense/support contributions
2. **Archetype Mapping**: Identify deck archetype by rating distribution
3. **Gap Analysis**: Spot missing offensive or defensive synergies
4. **Recommendation Engine**: Suggest cards that fill specific needs

### For Users

1. **Better Insights**: Understand why a deck works (or doesn't)
2. **Strategic Guidance**: See if deck lacks offense or defense
3. **Improvement Path**: Clear actions to improve deck balance
4. **Meta Understanding**: Learn synergy patterns

### For Development

1. **Extensible**: Easy to add new rating dimensions
2. **Testable**: Each rating can be validated independently
3. **Maintainable**: Clear categorization rules
4. **Performant**: O(1) lookups with caching

## Summary

The `synergyBonus` map structure extends the existing synergy system with categorized ratings (attack/defense/support) that enable comprehensive deck evaluation. The design:

✅ Provides granular scoring for evaluation engine
✅ Maintains backward compatibility
✅ Uses proven O(1) caching
✅ Enables detailed synergy reports
✅ Supports strategic deck analysis

**Next Steps**: Begin categorizing the 188 pairs following the guidelines in this document, starting with Priority 1 pairs (meta-relevant combinations).
