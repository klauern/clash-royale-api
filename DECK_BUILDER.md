# Deck Builder Documentation

The Clash Royale API includes an intelligent deck building system that creates optimized 8-card decks based on a player's card collection and levels. This system analyzes player data to recommend balanced decks with proper card roles and synergies.

## Overview

The deck builder is implemented in `go/pkg/deck/` and uses a sophisticated algorithm that:

- Analyzes a player's card collection and levels
- Assigns strategic roles to each card
- Scores cards based on level, rarity, and elixir cost
- Builds balanced decks with proper role distribution
- Provides strategic notes about the recommended deck

## Core Concepts

### Card Roles

Each card in Clash Royale is assigned a strategic role that defines its function in a deck:

| Role | Description | Examples |
|------|-------------|----------|
| **Win Condition** | Primary tower-damaging cards that form your main offensive strategy | Hog Rider, Royal Giant, Giant, P.E.K.K.A, Goblin Barrel |
| **Building** | Defensive structures that protect your towers and provide ground control | Cannon, Inferno Tower, Tombstone, Bomb Tower |
| **Big Spell** | High-elixir spells for clearing pushes or dealing major damage | Fireball, Poison, Lightning, Rocket |
| **Small Spell** | Low-elixir utility spells for cycle and small threats | Zap, Arrows, Log, Giant Snowball |
| **Support** | Mid-cost troops that support your win condition and provide versatility | Archers, Musketeer, Wizard, Valkyrie, Baby Dragon |
| **Cycle** | Low-cost cards that enable fast cycling and cheap elixir plays | Skeletons, Ice Spirit, Knight, Bats, Spear Goblins |

### Scoring Algorithm

The deck builder scores each card based on:

1. **Level Ratio** (`Level / MaxLevel`): How leveled the card is (40% weight)
2. **Rarity Boost**: Higher rarity cards get bonus points (15% weight)
   - Common: 1.0x
   - Rare: 1.05x
   - Epic: 1.1x
   - Legendary: 1.15x
   - Champion: 1.2x
3. **Elixir Weight**: Slightly favors cheaper cards for better cycling (15% weight)
4. **Role Bonus**: Cards with defined roles get a small bonus (5% weight)

## Usage

### Programmatically

```go
package main

import (
    "fmt"
    "log"
    "github.com/klauer/clash-royale-api/go/pkg/deck"
)

func main() {
    // Create a new deck builder
    builder := deck.NewBuilder("data")

    // Load card analysis from file
    analysis, err := builder.LoadLatestAnalysis("#PLAYERTAG", "data/analysis")
    if err != nil {
        log.Fatal(err)
    }

    // Build a deck from the analysis
    recommendation, err := builder.BuildDeckFromAnalysis(*analysis)
    if err != nil {
        log.Fatal(err)
    }

    // Print the deck
    fmt.Printf("Recommended Deck (Avg Elixir: %.1f):\n", recommendation.AvgElixir)
    for i, card := range recommendation.DeckDetail {
        fmt.Printf("%d. %s (Lvl %d, %d elixir) - %s\n",
            i+1, card.Name, card.Level, card.Elixir, card.Role)
    }

    // Print strategic notes
    if len(recommendation.Notes) > 0 {
        fmt.Println("\nStrategic Notes:")
        for _, note := range recommendation.Notes {
            fmt.Printf("- %s\n", note)
        }
    }

    // Save the deck recommendation
    savedPath, err := builder.SaveDeck(recommendation, "data/decks", "#PLAYERTAG")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("\nDeck saved to: %s\n", savedPath)
}
```

### Building a Deck from Card Analysis Data

The deck builder expects analysis data in this format:

```json
{
  "card_levels": {
    "Hog Rider": {
      "level": 8,
      "max_level": 13,
      "rarity": "Rare",
      "elixir": 4
    },
    "Fireball": {
      "level": 7,
      "max_level": 11,
      "rarity": "Rare",
      "elixir": 4
    }
  },
  "analysis_time": "2024-01-15T10:30:00Z"
}
```

## Deck Building Strategy

### Role Distribution

The algorithm aims for this optimal role distribution:

1. **1 Win Condition** - Your main offensive threat
2. **1 Building** - Defensive structure and ground control
3. **1 Big Spell** - High damage spell for clearing pushes
4. **1 Small Spell** - Cheap utility spell for cycling
5. **2 Support Cards** - Versatile troops that support your strategy
6. **2 Cycle Cards** - Low-cost cards for fast cycling

### Fallback Logic

If the ideal card for a role isn't available in the player's collection:

- **No Win Condition**: The builder selects the highest power cards instead and adds a note
- **Missing Roles**: The builder fills remaining slots with the highest-scoring available cards
- **Insufficient Cards**: Returns the best possible deck with available cards

### Strategic Notes

The deck builder provides contextual advice based on the deck composition:

- **High Elixir** (> 3.8): "Play patiently and build pushes"
- **Low Elixir** (< 2.8): "Pressure often and out-cycle counters"
- **No Building**: "Play troops high to kite"
- **No Spell**: "Beware of swarm matchups"

## Implementation Details

### Core Types

- `CardCandidate`: A card being considered with its metadata and score
- `DeckRecommendation`: Complete 8-card deck with details and notes
- `CardRole`: Enum defining strategic roles
- `CardAnalysis`: Input format for player's card collection

### Key Methods

```go
// Main entry point for building decks
func (b *Builder) BuildDeckFromAnalysis(analysis CardAnalysis) (*DeckRecommendation, error)

// Load analysis from file
func (b *Builder) LoadAnalysis(path string) (*CardAnalysis, error)

// Load most recent analysis for a player
func (b *Builder) LoadLatestAnalysis(playerTag, analysisDir string) (*CardAnalysis, error)

// Save deck recommendation
func (b *Builder) SaveDeck(deck *DeckRecommendation, outputDir, playerTag string) (string, error)
```

### Error Handling

The deck builder provides specific error types:

- `ErrInvalidDeckSize`: Deck doesn't contain exactly 8 cards
- `ErrInvalidAvgElixir`: Average elixir is outside valid range (0-10)
- `ErrNoWinCondition`: Deck lacks a win condition card
- `ErrInsufficientCards`: Not enough cards available for building

## Testing

The deck builder includes comprehensive tests:

```bash
# Run all deck package tests
cd go && go test ./pkg/deck/...

# Run specific tests with coverage
cd go && go test ./pkg/deck/... -v -cover
```

## Integration with CLI

While the deck builder is fully implemented, CLI integration is planned for Week 3. The planned command structure:

```bash
# Build a deck from player's latest analysis
./go/bin/cr-api deck build '#PLAYERTAG'

# Build a deck from specific analysis file
./go/bin/cr-api deck build --from-file 'path/to/analysis.json'

# Save deck to custom location
./go/bin/cr-api deck build '#PLAYERTAG' --output-dir 'custom/decks'
```

## Future Enhancements

Planned improvements include:

1. **Meta Awareness**: Consider current meta card win rates
2. **Synergy Scoring**: Bonus points for card synergies (e.g., Hog + Freeze)
3. **Player Preferences**: Allow role weight customization
4. **Multiple Deck Options**: Generate and rank multiple deck variations
5. **Performance Tracking**: Track deck performance against battle logs

## Best Practices

1. **Regular Updates**: Rebuild decks regularly as card levels change
2. **Consider Playstyle**: The algorithm builds balanced decks but adjust to your preferences
3. **Test Thoroughly**: Play several matches before committing to a new deck
4. **Watch the Meta**: Card effectiveness changes with game updates
5. **Keep Cards Leveled**: Higher-level cards receive higher scores in the algorithm