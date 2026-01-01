# Event Tracking Documentation

The Clash Royale API includes an event tracking system that monitors and analyzes player performance in special events, challenges, and tournaments.

## Overview

The event tracking system provides:

- Automatic detection of event battles from player battle logs
- Performance tracking (wins, losses, win rates, streaks)
- Event deck storage and retrieval
- Support for challenges, tournaments, and special events

## Supported Event Types

### Challenge Events

| Event Type | Description | Max Losses |
|------------|-------------|------------|
| **Grand Challenge** | Premium 12-win challenge | 3 |
| **Classic Challenge** | Standard 12-win challenge | 3 |
| **Draft Challenge** | Draft pick format | 3 |
| **Special Challenge** | Limited-time challenges | Varies |

### Tournament Events

| Event Type | Description | Format |
|------------|-------------|--------|
| **Tournament** | Custom player tournaments | Swiss, Round Robin, Elimination |
| **Special Event** | Official Supercell events | Usually 2-3 losses |
| **Sudden Death** | Single elimination | 1 loss |

## Data Storage

Event decks are stored in:
```
data/event_decks/
├── PLAYER_TAG/
│   ├── challenges/
│   ├── tournaments/
│   └── special_events/
```

## Event Data

Each event deck includes:

- Event metadata (name, type, timestamps)
- Deck composition (8 cards with levels)
- Performance metrics (wins, losses, win rate, streaks)
- Individual battle records
- Event-specific rules

## Usage

```bash
# Scan for new event decks
./bin/cr-api events scan --tag '#PLAYERTAG'

# Show event performance summary
./bin/cr-api events summary --tag '#PLAYERTAG' --days 30

# Export event data to CSV
./bin/cr-api events export --tag '#PLAYERTAG' --format csv
```

## Event Detection

The system identifies events by:

- Battle mode analysis (challenges vs tournaments)
- Event name pattern matching
- Timestamp clustering to group battles
- Deck consistency tracking

## Performance Metrics

Tracked metrics include:

- **Win Rate**: Wins / (Wins + Losses)
- **Crown Differential**: Crowns Earned - Crowns Lost
- **Streak Analysis**: Current and best win streaks
- **Event Progress**: In progress, completed, or eliminated

## Best Practices

1. Scan battle logs regularly to capture all events
2. Maintain regular backups of event data
3. Respect API rate limits when analyzing multiple players