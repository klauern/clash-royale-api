# Event Tracking Documentation

The Clash Royale API includes a comprehensive event tracking system that monitors and analyzes player performance in special events, challenges, and tournaments. This system parses battle logs to identify event decks and tracks detailed performance metrics.

## Overview

The event tracking system is implemented in `go/pkg/events/` and provides:

- Automatic detection of event battles from player battle logs
- Performance tracking (wins, losses, win rates, streaks)
- Event deck storage and retrieval
- Analysis of event-specific deck usage
- Support for multiple event types (challenges, tournaments, special events)

## Supported Event Types

### Challenge Events

| Event Type | Description | Max Losses | Typical Rewards |
|------------|-------------|------------|-----------------|
| **Grand Challenge** | Premium 12-win challenge | 3 | 2000+ cards, 500+ gold |
| **Classic Challenge** | Standard 12-win challenge | 3 | 1000+ cards, 200+ gold |
| **Draft Challenge** | Draft pick format challenge | 3 | Varies by draft |
| **Special Challenge** | Limited-time unique challenges | Varies | Varies by event |

### Tournament Events

| Event Type | Description | Format |
|------------|-------------|--------|
| **Tournament** | Custom tournaments created by players | Swiss, Round Robin, or Elimination |
| **Special Event** | Official Supercell special events | Varies (usually 2-3 losses) |
| **Sudden Death** | Single elimination format | 1 loss |
| **Double Elimination** | Two chances before elimination | 2 losses |

## Core Components

### 1. Event Parser (`parser.go`)

The parser scans battle logs to identify event battles and extract relevant data:

```go
// Create a new parser
parser := events.NewParser()

// Parse battle logs for event decks
eventDecks, err := parser.ParseBattleLogs(battles, playerTag)
```

**Detection Logic:**
- Battle mode analysis (identifies challenge/tournament modes)
- Event name pattern matching (e.g., "Lava Hound Challenge")
- Timestamp clustering to group battles into events
- Deck hash generation to identify consistent deck usage

### 2. Event Manager (`manager.go`)

The manager handles storage, retrieval, and organization of event deck data:

```go
// Create a new event manager
manager := events.NewManager("data")

// Store an event deck
err := manager.StoreEventDeck(eventDeck)

// Retrieve player's event deck collection
collection, err := manager.GetPlayerEventDecks("#PLAYERTAG")

// Get specific event deck
eventDeck, err := manager.GetEventDeck(eventID, playerTag)
```

**Directory Structure:**
```
data/event_decks/
├── PLAYER_TAG/
│   ├── challenges/      # Challenge event decks
│   ├── tournaments/     # Tournament decks
│   ├── special_events/  # Special event decks
│   └── aggregated/      # Processed analytics
```

## Data Models

### EventDeck

The core data structure representing a complete event deck performance:

```json
{
  "event_id": "2024_01_15_hog_rider_challenge_abc123",
  "player_tag": "#PLAYERTAG",
  "event_name": "Hog Rider Challenge",
  "event_type": "challenge",
  "start_time": "2024-01-15T10:30:00Z",
  "end_time": "2024-01-15T11:45:00Z",
  "deck": {
    "cards": [
      {
        "name": "Hog Rider",
        "id": 26000000,
        "level": 8,
        "max_level": 13,
        "rarity": "Rare",
        "elixir_cost": 4
      }
    ],
    "avg_elixir": 3.5
  },
  "performance": {
    "wins": 9,
    "losses": 3,
    "win_rate": 0.75,
    "crowns_earned": 24,
    "crowns_lost": 15,
    "max_wins": 12,
    "current_streak": 0,
    "best_streak": 4,
    "progress": "eliminated"
  },
  "battles": [...],
  "event_rules": {
    "card_level_cap": 9,
    "special_rules": "Hog Rider always in starting hand"
  }
}
```

### BattleRecord

Individual battle data within an event:

```json
{
  "timestamp": "2024-01-15T10:35:00Z",
  "opponent_tag": "#OPPONENT",
  "opponent_name": "ProPlayer123",
  "result": "win",
  "crowns": 3,
  "opponent_crowns": 0,
  "trophy_change": 20,
  "battle_mode": "challenge"
}
```

## Usage Examples

### 1. Parse and Store Event Decks

```go
package main

import (
    "context"
    "log"
    "github.com/klauer/clash-royale-api/go/pkg/clashroyale"
    "github.com/klauer/clash-royale-api/go/pkg/events"
)

func main() {
    // Initialize API client and event manager
    client := clashroyale.NewClient()
    manager := events.NewManager("data")
    parser := events.NewParser()

    // Get player battle logs
    playerTag := "#PLAYERTAG"
    battles, err := client.GetPlayerBattleLogs(context.Background(), playerTag)
    if err != nil {
        log.Fatal(err)
    }

    // Parse battle logs for event decks
    eventDecks, err := parser.ParseBattleLogs(battles, playerTag)
    if err != nil {
        log.Fatal(err)
    }

    // Store detected event decks
    for _, eventDeck := range eventDecks {
        err := manager.StoreEventDeck(eventDeck)
        if err != nil {
            log.Printf("Failed to store event deck: %v", err)
            continue
        }
        log.Printf("Stored event deck: %s", eventDeck.EventName)
    }
}
```

### 2. Analyze Event Performance

```go
// Retrieve player's event decks
collection, err := manager.GetPlayerEventDecks("#PLAYERTAG")
if err != nil {
    log.Fatal(err)
}

// Get best performing decks
bestDecks := collection.GetBestDecksByWinRate(5, 10) // Min 5 battles, top 10
for _, deck := range bestDecks {
    fmt.Printf("Deck: %s\n", deck.EventName)
    fmt.Printf("Win Rate: %.1f%% (%d-%d)\n",
        deck.Performance.WinRate*100,
        deck.Performance.Wins,
        deck.Performance.Losses)
    fmt.Printf("Best Streak: %d wins\n", deck.Performance.BestStreak)
    fmt.Println("---")
}
```

### 3. Track Recent Performance

```go
// Get event decks from last 30 days
recentDecks := collection.GetRecentDecks(30)

// Calculate overall performance
totalWins := 0
totalLosses := 0
for _, deck := range recentDecks {
    totalWins += deck.Performance.Wins
    totalLosses += deck.Performance.Losses
}

overallWinRate := float64(totalWins) / float64(totalWins+totalLosses)
fmt.Printf("30-day Event Performance: %.1f%% win rate\n", overallWinRate*100)
```

## Event Detection Algorithm

### 1. Battle Mode Classification

The parser uses battle mode information to identify events:

- Standard 1v1 battles are filtered out
- Challenge modes are classified by type (Grand, Classic, Draft)
- Tournament modes are identified by custom game settings
- Special events are detected via name patterns

### 2. Time Window Analysis

Battles are grouped into events using:

- **Temporal Proximity**: Battles within a reasonable time window
- **Deck Consistency**: Same or very similar 8-card composition
- **Win Streaks**: Sequential wins indicating active event run

### 3. Event Type Inference

```go
// Example event patterns detected
eventPatterns := map[string]string{
    "lava":          "Lava Hound Challenge",
    "hog":           "Hog Rider Challenge",
    "mortar":        "Mortar Challenge",
    "graveyard":     "Graveyard Challenge",
    "ram rage":      "Ram Rage Challenge",
    "sparky":        "Sparky Challenge",
    "electro":       "Electro Challenge",
    "skeleton":      "Skeleton Army Challenge",
    "bandit":        "Bandit Challenge",
    "night witch":   "Night Witch Challenge",
    "royale":        "Clash Royale Championship",
    "worlds":        "World Championship",
    "ccgs":          "Clash Championship Series",
}
```

## Performance Metrics

### Calculated Metrics

1. **Win Rate**: `Wins / (Wins + Losses)`
2. **Crown Differential**: `Crowns Earned - Crowns Lost`
3. **Average Crowns per Battle**: `Crowns Earned / Total Battles`
4. **Streak Analysis**: Current and best win streaks
5. **Event Progress**: In progress, completed, eliminated

### Progress Tracking

Different event types have unique completion criteria:

- **Grand/Classic/Draft Challenges**: Max 3 losses
- **Tournaments**: Varies by format
- **Special Events**: Usually 2-3 losses
- **Sudden Death**: 1 loss elimination

## CLI Integration (Planned)

Event tracking commands will be integrated in Week 3:

```bash
# Scan for new event decks
./go/bin/cr-api events scan '#PLAYERTAG'

# Show event performance summary
./go/bin/cr-api events summary '#PLAYERTAG' --days 30

# Export event data to CSV
./go/bin/cr-api events export '#PLAYERTAG' --format csv

# Analyze best performing decks
./go/bin/cr-api events analyze '#PLAYERTAG' --min-battles 5
```

## Future Enhancements

Planned features include:

1. **Meta Analysis**: Aggregate event deck trends across all players
2. **Opponent Tracking**: Track frequently faced opponents in events
3. **Deck Synergy Analysis**: Identify optimal card combinations for events
4. **Prediction Model**: Suggest deck changes based on historical performance
5. **Real-time Updates**: Live event progress tracking during active challenges
6. **Event Recommendations**: Suggest events based on player's deck collection

## Best Practices

1. **Regular Scanning**: Scan battle logs regularly to capture all events
2. **Backup Data**: Event data is valuable - maintain regular backups
3. **Privacy**: Respect opponent privacy when analyzing shared data
4. **Rate Limits**: When analyzing multiple players, respect API rate limits
5. **Data Validation**: Always validate event deck data before processing

## Error Handling

The system provides specific error types:

- `ErrInvalidDeckSize`: Deck doesn't contain exactly 8 cards
- `ErrInvalidElixirCost`: Card elixir cost outside valid range
- `ErrEventNotFound`: Event deck not found in storage
- `ErrInvalidEventType`: Unsupported event type provided