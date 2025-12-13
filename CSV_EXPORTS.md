# CSV Exports Documentation

The Clash Royale API provides comprehensive CSV export functionality for player data, card collections, battle logs, event performance, and analysis results. This system allows for easy data analysis, spreadsheet integration, and external processing.

## Overview

CSV exports are available in both Python (complete implementation) and Go (implementation in progress). The export system supports:

- Player profiles and statistics
- Card collections and upgrade information
- Battle logs and match history
- Event deck performance data
- Card collection analysis results
- Custom CSV formats and filtering options

## Directory Structure

Exported CSV files are organized in a structured directory hierarchy:

```
data/csv/
├── players/         # Player profile exports
├── clans/          # Clan information exports
├── analysis/       # Collection analysis exports
├── reference/      # Static game data (cards, arenas, etc.)
├── events/         # Event deck performance exports
└── decks/          # Deck recommendation exports
```

## Export Types

### 1. Player Data Export

**Filename**: `players.csv`

**Fields**:
| Column | Description | Example |
|--------|-------------|---------|
| Tag | Player tag | `#PLAYERTAG` |
| Name | Player name | `ProPlayer123` |
| Experience Level | Current experience level | `85` |
| Experience Points | Total XP earned | `2456789` |
| Trophies | Current trophies | `4521` |
| Best Trophies | All-time best | `5123` |
| Wins | Total wins | `12543` |
| Losses | Total losses | `10456` |
| Total Battles | Battles played | `23456` |
| Three Crown Wins | 3-crown victories | `3456` |
| Challenge Wins | Challenge victories | `234` |
| Tournament Wins | Tournament victories | `12` |
| Role | Clan role | `member`, `elder`, `coLeader`, `leader` |
| Clan Tag | Current clan tag | `#CLANTAG` |
| Clan Name | Current clan name | `Elite Squad` |
| Arena ID | Current arena ID | `54000000` |
| Arena Name | Current arena name | `Legendary Arena` |
| League ID | Current league ID | `29000000` |
| League Name | Current league name | `Legendary League` |
| Donations | Current season donations | `2345` |
| Star Points | Star points earned | `12345` |
| Created At | Account creation date | `2016-04-04 12:34:56` |

### 2. Card Reference Export

**Filename**: `cards.csv`

**Fields**:
| Column | Description | Example |
|--------|-------------|---------|
| ID | Card ID number | `26000000` |
| Name | Card name | `Hog Rider` |
| Elixir Cost | Elixir cost | `4` |
| Type | Card type | `troop`, `spell`, `building` |
| Rarity | Card rarity | `Common`, `Rare`, `Epic`, `Legendary`, `Champion` |
| Max Level | Maximum level | `13` |
| Max Evolution Level | Maximum evolution level | `3` |
| Description | Card description | `Fast melee unit...` |

### 3. Card Collection Export

**Filename**: `card_collection_[TIMESTAMP]_[PLAYER].csv`

**Fields**:
| Column | Description | Example |
|--------|-------------|---------|
| Card Name | Card name | `Hog Rider` |
| Level | Current level | `8` |
| Max Level | Maximum possible level | `13` |
| Count | Cards owned | `45` |
| Rarity | Card rarity | `Rare` |
| Elixir Cost | Elixir cost | `4` |
| Upgrade Cost | Gold needed for upgrade | `500` |
| Cards to Max | Cards needed for max level | `460` |
| Gold to Max | Total gold to max level | `96000` |
| Priority | Upgrade priority score | `8.5` |
| Role | Strategic role | `win_condition` |
| Evolution Level | Current evolution level | `1` |
| Evolution Possible | If evolution is possible | `true` |

### 4. Battle Log Export

**Filename**: `battles_[TIMESTAMP]_[PLAYER].csv`

**Fields**:
| Column | Description | Example |
|--------|-------------|---------|
| Battle Time | UTC timestamp | `2024-01-15 10:30:45` |
| Opponent Tag | Opponent player tag | `#OPPONENT` |
| Opponent Name | Opponent name | `RivalPlayer` |
| Result | Battle result | `win` or `loss` |
| Team Crowns | Your crowns earned | `3` |
| Opponent Crowns | Opponent crowns | `0` |
| Trophy Change | Trophy delta | `+28` |
| Battle Mode | Game mode | `Ladder 1v1` |
| Arena | Arena name | `Legendary Arena` |
| Deck Hash | Hash of your deck | `abc123...` |
| Opponent Deck Hash | Hash of opponent deck | `def456...` |
| Is Event | If battle was in event | `false` |
| Event Type | Type of event | `challenge` (if applicable) |

### 5. Event Performance Export

**Filename**: `event_performance_[TIMESTAMP]_[PLAYER].csv`

**Fields**:
| Column | Description | Example |
|--------|-------------|---------|
| Event ID | Unique event identifier | `2024_01_15_hog_challenge` |
| Event Name | Event display name | `Hog Rider Challenge` |
| Event Type | Type of event | `challenge` |
| Start Time | Event start timestamp | `2024-01-15 10:30:00` |
| End Time | Event end timestamp | `2024-01-15 11:45:00` |
| Wins | Total wins | `9` |
| Losses | Total losses | `3` |
| Win Rate | Win percentage | `0.75` |
| Max Wins | Maximum possible wins | `12` |
| Best Streak | Best win streak | `4` |
| Current Streak | Current streak | `0` |
| Progress | Event status | `eliminated` |
| Deck Used | 8-card deck list | `Hog Rider...` |
| Avg Elixir | Deck average elixir | `3.5` |

### 6. Deck Recommendation Export

**Filename**: `deck_recommendation_[TIMESTAMP]_[PLAYER].csv`

**Fields**:
| Column | Description | Example |
|--------|-------------|---------|
| Card Name | Recommended card | `Hog Rider` |
| Level | Current level | `8` |
| Max Level | Maximum level | `13` |
| Rarity | Card rarity | `Rare` |
| Elixir Cost | Elixir cost | `4` |
| Role | Strategic role | `win_condition` |
| Score | Algorithm score | `8.75` |
| Priority | Upgrade priority | `high` |
| Notes | Card-specific notes | `Core win condition` |

## Usage Examples

### Python Implementation

```python
from clash_royale_api import ClashRoyaleAPI

async def main():
    async with ClashRoyaleAPI() as api:
        # Export all data types
        await api.export_to_csv(
            player_tag="#PLAYERTAG",
            csv_types=["all"],
            output_dir="custom_csv_dir"
        )

        # Export specific types
        await api.export_to_csv(
            player_tag="#PLAYERTAG",
            csv_types=["player", "cards", "battles"]
        )

# Run with task
task export-all '#PLAYERTAG'
task export-csv '#PLAYERTAG' --csv-types player,cards,battles
```

### Go Implementation

```go
package main

import (
    "context"
    "log"
    "github.com/klauer/clash-royale-api/go/internal/exporter/csv"
    "github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

func main() {
    // Create API client
    client := clashroyale.NewClient()

    // Get player data
    player, err := client.GetPlayer(context.Background(), "#PLAYERTAG")
    if err != nil {
        log.Fatal(err)
    }

    // Create player exporter
    playerExporter := csv.NewPlayerExporter()
    err = playerExporter.Export("data", player)
    if err != nil {
        log.Fatal(err)
    }

    // Get all cards
    cards, err := client.GetAllCards(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // Create cards exporter
    cardsExporter := csv.NewCardsExporter()
    err = cardsExporter.Export("data", cards)
    if err != nil {
        log.Fatal(err)
    }
}

// CLI usage (planned)
./go/bin/cr-api export player '#PLAYERTAG'
./go/bin/cr-api export cards
./go/bin/cr-api export all '#PLAYERTAG' --output-dir custom/path
```

## CSV Format Standards

### Encoding and Delimiters
- **Encoding**: UTF-8
- **Delimiter**: Comma (`,`)
- **Line Ending**: CRLF (`\r\n`) for Windows compatibility
- **Quote Character**: Double quote (`"`)
- **Escape Character**: Backslash (`\`)

### Date/Time Format
All timestamps use ISO 8601 format: `YYYY-MM-DD HH:MM:SS`
- Timezone: UTC
- 24-hour format
- No milliseconds

### Number Formatting
- **Integers**: No formatting, plain digits
- **Decimals**: Dot (`.`) as decimal separator
- **Percentages**: Decimal fraction (0.75 = 75%)
- **Currencies**: Integer values (100 = 100 gold)

### Text Fields
- **Empty Values**: Empty string (`""`)
- **Null Values**: Empty string (`""`)
- **Quotes**: Escaped with backslash (`\"`)
- **Newlines**: Replaced with space
- **Commas**: Field is quoted if contains comma

## Advanced Features

### Custom Export Formats

```python
# Custom CSV with specific fields
await api.export_to_csv(
    player_tag="#PLAYERTAG",
    csv_types=["custom"],
    custom_fields=["name", "trophies", "wins", "losses"],
    custom_filename="my_custom_export.csv"
)
```

### Filtering and Selection

```python
# Export battles from date range
await api.export_battles(
    player_tag="#PLAYERTAG",
    from_date="2024-01-01",
    to_date="2024-01-31",
    filter_mode="ladder"  # Only ladder battles
)

# Export only completed events
await api.export_events(
    player_tag="#PLAYERTAG",
    event_status="completed",
    min_battles=5
)
```

### Data Validation

All exports include automatic validation:

1. **Required Fields**: All required fields are present
2. **Data Types**: Values match expected types
3. **Ranges**: Numbers within valid ranges
4. **Formatting**: Dates and times properly formatted
5. **Encoding**: UTF-8 encoding validation

## CLI Commands (Planned)

```bash
# Export all data for a player
./go/bin/cr-api export all '#PLAYERTAG'

# Export specific data types
./go/bin/cr-api export player '#PLAYERTAG'
./go/bin/cr-api export cards
./go/bin/cr-api export battles '#PLAYERTAG'
./go/bin/cr-api export events '#PLAYERTAG'

# Custom output directory
./go/bin/cr-api export all '#PLAYERTAG' --output-dir /path/to/exports

# Date range filtering
./go/bin/cr-api export battles '#PLAYERTAG' --from 2024-01-01 --to 2024-01-31

# Event filtering
./go/bin/cr-api export events '#PLAYERTAG' --event-type challenge --min-wins 5
```

## Integration Examples

### Excel/Google Sheets

1. Import CSV using "Import from file" option
2. Use first row as headers
3. Set delimiter to comma
4. Encoding: UTF-8
5. Date columns: Format as Date/Time

### Pandas (Python)

```python
import pandas as pd

# Load player data
df = pd.read_csv('data/csv/players/player_20240115_123456.csv')

# Filter high-level players
high_level = df[df['Experience Level'] >= 100]

# Calculate win rate
df['Win Rate'] = df['Wins'] / (df['Wins'] + df['Losses'])

# Export to Excel
df.to_excel('players.xlsx', index=False)
```

### R Analysis

```r
# Load CSV files
player_data <- read.csv('data/csv/players/player_20240115_123456.csv')
battle_data <- read.csv('data/csv/battles/battles_20240115_123456.csv')

# Merge data
merged <- merge(player_data, battle_data, by='Tag')

# Analyze performance by arena
library(dplyr)
arena_stats <- merged %>%
  group_by(Arena.Name) %>%
  summarise(
    Avg_Win_Rate = mean(Result == 'win'),
    Total_Battles = n()
  )
```

## Performance Considerations

1. **Large Datasets**: Use streaming for large exports
2. **Memory Usage**: Process in chunks for memory efficiency
3. **Disk Space**: Monitor available disk space for exports
4. **Concurrent Exports**: Avoid simultaneous exports for same player
5. **API Limits**: Respect rate limits during data collection

## Error Handling

Common export errors and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| File permission denied | Directory not writable | Check directory permissions |
| Invalid player tag | Tag doesn't exist | Verify player tag format |
| Rate limit exceeded | Too many API requests | Add delays between requests |
| Out of memory | Large dataset | Process data in chunks |
| Encoding issues | Special characters | Ensure UTF-8 encoding |

## Best Practices

1. **Regular Backups**: Export data regularly for backup
2. **Compression**: Compress old CSV files to save space
3. **Validation**: Validate CSV files after export
4. **Documentation**: Document custom export formats
5. **Version Control**: Track CSV schema changes
6. **Privacy**: Remove or anonymize sensitive data when sharing
7. **Retention**: Establish data retention policies