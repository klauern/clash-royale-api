# Player Context Documentation

The Clash Royale API provides player context functionality for context-aware deck evaluation, enabling personalized analysis based on a player's arena level, card collection, card levels, and evolution status.

## Overview

Player context transforms deck evaluation from generic recommendations to personalized analysis by considering:

- **Arena Progress**: Card unlock validation based on current arena
- **Card Collection**: Which cards the player owns and their levels
- **Evolution Status**: Which evolutions are unlocked and progress toward next evolution
- **Level-Based Analysis**: Upgrade gap calculation and ladder readiness assessment
- **Missing Card Detection**: Arena-aware suggestions for替代 cards

When player context is provided via the `--tag` flag, deck evaluation becomes context-aware, filtering recommendations to only playable cards and providing personalized insights.

## Usage

### Basic Player Context

```bash
# Fetch player data from API and build context-aware deck
./bin/cr-api deck build --tag '#PLAYERTAG'

# Analyze with player context
./bin/cr-api analyze --tag '#PLAYERTAG'
```

### What-If Analysis with Player Context

```bash
# Simulate upgrades with player context
./bin/cr-api what-if --tag '#PLAYERTAG' --upgrade "Archers:15" --show-decks
```

### Manual Arena Override

```bash
# Simulate deck building for a specific arena (useful for theory-crafting)
./bin/cr-api deck build --tag '#PLAYERTAG' --arena 5
```

**Note**: The `--arena` flag is primarily used for theory-crafting and analysis. When omitted, the player's current arena is fetched from the API.

## Configuration

### Environment Variables

```env
# Default player tag for commands (optional)
DEFAULT_PLAYER_TAG=#PLAYERTAG
```

### CLI Flags

| Flag | Description | Required |
|------|-------------|----------|
| `--tag <TAG>` | Player tag to fetch from API | Yes |
| `--arena <ID>` | Override arena level (0-15) | No |

### Evolution Tracking

```bash
# Environment variable
export UNLOCKED_EVOLUTIONS="Archers,Knight,Valkyrie"

# CLI flag
./bin/cr-api deck build --tag '#PLAYERTAG' --unlocked-evolutions "Archers,Knight"
```

See [EVOLUTION.md](EVOLUTION.md) for complete evolution system documentation.

## Features

### Arena-Aware Card Unlock Validation

The system validates whether cards are unlocked based on the player's arena level. Cards unlock at specific arenas, and recommendations respect these restrictions.

**Unlock Arena Examples**:

| Card | Unlock Arena |
|------|--------------|
| Knight, Archers, Giants | Training Camp (0) |
| Spear Goblins | Arena 1 |
| Hog Rider | Arena 2 |
| Royal Giant | Arena 3 |
| Princess | Arena 4 |
| Mega Knight | Arena 11 |
| Skeleton King | Arena 15 |

**Validation Behavior**:
- If `ArenaID == 0`: No restrictions (training mode)
- If `ArenaID > 0`: Card must unlock at or below player's current arena
- Missing cards are marked with their unlock arena and locked status

### Collection-Based Playability Scoring

Deck evaluation scores cards based on ownership and level:

- **Owned Cards**: Scored based on level, rarity, and combat stats
- **Missing Cards**: Excluded from deck building, marked with替代 suggestions
- **Level Ratio**: Normalized score comparing current level to max level

**Playability Formula**:
```
LevelRatio = CurrentLevel / MaxLevel
CardScore = BaseScore × LevelRatio × RarityMultiplier
```

### Level-Based Ladder Analysis

The system calculates upgrade gaps and ladder readiness:

**Upgrade Gap Calculation**:
```
TotalGap = Σ (MaxLevel - CurrentLevel) for all deck cards
```

**Use Cases**:
- Identify how many upgrades are needed to max a deck
- Prioritize gold spending on high-impact upgrades
- Compare multiple decks by upgrade cost

**Example**:
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --show-upgrade-impact
```

Output:
```
Upgrade Impact Analysis
=======================
Deck: Hog Cycle
Total Upgrade Gap: 24 levels
Highest Priority Upgrades:
1. Hog Rider (Lv.12 → Lv.14): 2000 gold
2. Ice Spirit (Lv.11 → Lv.14): 1500 gold
```

### Evolution Integration

Player context tracks evolution progress and integrates it into deck building:

**Tracked Evolution Data**:
- `EvolutionLevel`: Current evolution stage (0-3)
- `MaxEvolutionLevel`: Maximum possible evolutions
- `Count`: Number of evolution shards owned

**Evolution Bonuses**:
- Evolved cards receive role-based scoring bonuses
- Some cards change roles when evolved (e.g., Valkyrie: Cycle → Support)
- Evolution slot priority determines which evolutions are active

**Example**:
```bash
./bin/cr-api deck build --tag '#PLAYERTAG' --unlocked-evolutions "Archers,Knight" --verbose
```

### Missing Card Identification

When a deck contains cards the player doesn't own, the system provides:

**Missing Card Information**:
- Card name and rarity
- Unlock arena and arena name
- Locked status (whether the card is available at current arena)
- Alternative cards the player owns

**Example Output**:
```
Missing Cards Analysis
======================
Deck Status: 6/8 cards available

1. Electro Dragon (Epic)
   Unlocks: Rascal's Hideout (Arena 13)
   Status: ✓ Unlocked - Available in chests and shop
   Alternatives: Baby Dragon, Mega Minion

2. Mega Knight (Legendary)
   Unlocks: Electro Valley (Arena 11)
   Status: 🔒 LOCKED - Progress to Arena 11 to unlock
   Alternatives: P.E.K.K.A, Mini P.E.K.K.A
```

## Examples

### Example 1: Basic Deck Building with Player Context

```bash
./bin/cr-api deck build --tag '#ABC123'
```

**Without Player Context**:
- Recommends best cards regardless of ownership
- May include cards the player doesn't own
- Generic scoring without level consideration

**With Player Context**:
- Only recommends cards the player owns
- Scores cards based on player's levels
- Respects arena unlock restrictions
- Provides personalized viability score

### Example 2: Arena Override for Theory-Crafting

```bash
# Build a deck as if the player were in Arena 8
./bin/cr-api deck build --tag '#ABC123' --arena 8
```

**Use Cases**:
- Plan future decks before unlocking cards
- Compare deck viability at different arenas
- Test strategies without being current arena limited

### Example 3: Upgrade Impact Analysis

```bash
./bin/cr-api what-if --tag '#ABC123' --upgrade "Hog Rider:14" --show-decks
```

**Output**:
```
============================================================================
                        WHAT-IF ANALYSIS
============================================================================

Scenario: Upgrade 1 card: Hog Rider
What-if analysis for Player (#ABC123)

Upgrades Simulated
-------------------
Card       From  To  Gold
----       ----  --  ----
Hog Rider  12    14  2000

Total Gold Cost: 2000

Impact Analysis
---------------
Deck Score Delta:     +1.250000
Viability Change:     +18.5%

Before:
Deck: Hog Cycle
1. Hog Rider (Lv.12), 2. Ice Spirit (Lv.11), 3. Cannon (Lv.12)...

After:
Deck: Hog Cycle
1. Hog Rider (Lv.14), 2. Ice Spirit (Lv.11), 3. Cannon (Lv.12)...

Recommendation
-------------
Highly recommended! This upgrade (2000 gold) significantly improves your deck viability by 18.5%.
```

### Example 4: Evolution-Aware Deck Building

```bash
./bin/cr-api deck build --tag '#ABC123' --unlocked-evolutions "Archers,Knight,Valkyrie" --strategy aggro
```

**Evolution Integration**:
- Valkyrie promoted from Cycle to Support role
- Evolution bonuses applied to scoring
- Slot allocation prioritizes win conditions

## Integration with Other Features

### Deck Building Strategies

Player context enhances all deck building strategies:

```bash
# Balanced strategy with player context
./bin/cr-api deck build --tag '#ABC123' --strategy balanced

# Cycle strategy filters out high-cost cards the player doesn't own
./bin/cr-api deck build --tag '#ABC123' --strategy cycle

# Aggro strategy prioritizes overleveled win conditions
./bin/cr-api deck build --tag '#ABC123' --strategy aggro
```

See [DECK_BUILDER.md](DECK_BUILDER.md) for complete strategy documentation.

### What-If Analysis

Player context enables accurate upgrade simulation:

```bash
# Offline mode using existing analysis
./bin/cr-api what-if --tag '#ABC123' --from-analysis data/analysis/player.json \
  --upgrade "Archers:9:15" --save --json
```

### CSV Exports

Player context data is included in CSV exports:

- **Player Cards**: Level, max level, evolution level, count
- **Decks**: Playability status, missing cards, upgrade gaps
- **Analysis**: Average level, upgrade cost, viability score

See [CSV_EXPORTS.md](CSV_EXPORTS.md) for export format documentation.

## Technical Details

### PlayerContext Structure

```go
type PlayerContext struct {
    // Arena information
    Arena     *clashroyale.Arena
    ArenaID   int
    ArenaName string

    // Card collection: map of card name -> level data
    Collection map[string]CardLevelInfo

    // Evolution data: which cards have evolutions unlocked
    UnlockedEvolutions map[string]bool

    // Player metadata
    PlayerTag  string
    PlayerName string
}
```

### CardLevelInfo Structure

```go
type CardLevelInfo struct {
    Level             int
    MaxLevel          int
    EvolutionLevel    int
    MaxEvolutionLevel int
    Rarity            string
    Count             int
}
```

### Key Methods

| Method | Description |
|--------|-------------|
| `HasCard(cardName string) bool` | Check if player owns a card |
| `GetCardLevel(cardName string) int` | Get card level (0 if not owned) |
| `HasEvolution(cardName string) bool` | Check if evolution is unlocked |
| `IsCardUnlockedInArena(cardName string) bool` | Validate arena unlock status |
| `CalculateUpgradeGap(deckCards []CardCandidate) int` | Calculate total levels to max |
| `GetAverageLevel(cardNames []string) float64` | Get average level of card list |

## Best Practices

1. **Always use `--tag` for personalized recommendations**: Generic deck building doesn't consider ownership or levels
2. **Check missing cards before committing to a deck**: Alternative suggestions can save gold
3. **Use upgrade impact analysis before spending gold**: Prioritize high-impact upgrades
4. **Combine with evolution tracking**: Evolved cards receive significant bonuses
5. **Leverage arena override for planning**: Test future decks before unlocking cards
6. **Regularly rebuild decks as card levels change**: What's optimal today may not be optimal tomorrow

## Troubleshooting

### Common Issues

**Issue**: "Card not found in collection"
- **Solution**: Verify the player tag is correct and the API token is valid

**Issue**: "All cards showing as locked"
- **Solution**: Check that arena data is being fetched correctly (API token permissions)

**Issue**: "Evolution bonuses not applied"
- **Solution**: Ensure `--unlocked-evolutions` flag or `UNLOCKED_EVOLUTIONS` env var is set

**Issue**: "Deck recommendations include cards I don't own"
- **Solution**: Use `--tag` flag to enable player context, not `--from-analysis` without context

## Related Documentation

- [CLI_REFERENCE.md](CLI_REFERENCE.md) - Complete command reference
- [DECK_BUILDER.md](DECK_BUILDER.md) - Deck building strategies and algorithms
- [EVOLUTION.md](EVOLUTION.md) - Evolution system configuration
- [CSV_EXPORTS.md](CSV_EXPORTS.md) - Data export formats
- [TESTING.md](TESTING.md) - Testing player context functionality
