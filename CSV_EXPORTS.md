# CSV Exports Documentation

The Clash Royale API provides comprehensive CSV export functionality for player data, card collections, event performance, and analysis results. This system allows for easy data analysis, spreadsheet integration, and external processing.

## Overview

CSV exports are available in the Go implementation. The export system supports:

- Player profiles and statistics
- Card collections and upgrade information
- Event deck performance data
- Card collection analysis results
- Custom CSV formats and filtering options

## Directory Structure

Exported CSV files are organized in a structured directory hierarchy:

```
data/csv/
├── players/         # Player profile exports
├── analysis/        # Collection analysis exports
├── reference/       # Static game data (cards, arenas, etc.)
└── events/          # Event deck performance exports
```

## Export Types

### 1. Player Data Export

**Filename**: `player_[TIMESTAMP]_[PLAYER].csv`

**Fields**:

| Column | Description | Example |
|--------|-------------|---------|
| Tag | Player tag | `#PLAYERTAG` |
| Name | Player name | `ProPlayer123` |
| Experience Level | Current experience level | `85` |
| Trophies | Current trophies | `4521` |
| Best Trophies | All-time best | `5123` |
| Wins | Total wins | `12543` |
| Losses | Total losses | `10456` |
| Win Rate | Win percentage | `0.543` |
| Clan Tag | Current clan tag | `#CLANTAG` |
| Clan Name | Current clan name | `Elite Squad` |
| Arena Name | Current arena name | `Legendary Arena` |

### 2. Card Reference Export

**Filename**: `card_database_[TIMESTAMP].csv`

**Fields**:

| Column | Description | Example |
|--------|-------------|---------|
| ID | Card ID number | `26000000` |
| Name | Card name | `Hog Rider` |
| Elixir Cost | Elixir cost | `4` |
| Type | Card type | `troop`, `spell`, `building` |
| Rarity | Card rarity | `Common`, `Rare`, `Epic`, `Legendary` |
| Max Level | Maximum level | `13` |
| Description | Card description | `Fast melee unit...` |

### 3. Card Collection Export

**Filename**: `cards_[TIMESTAMP]_[PLAYER].csv`

**Fields**:

| Column | Description | Example |
|--------|-------------|---------|
| Card Name | Card name | `Hog Rider` |
| Level | Current level | `8` |
| Max Level | Maximum possible level | `13` |
| Count | Cards owned | `45` |
| Rarity | Card rarity | `Rare` |
| Elixir Cost | Elixir cost | `4` |

### 4. Event Performance Export

**Filename**: `events_[TIMESTAMP]_[PLAYER].csv`

**Fields**:

| Column | Description | Example |
|--------|-------------|---------|
| Event Name | Event display name | `Hog Rider Challenge` |
| Event Type | Type of event | `challenge` |
| Wins | Total wins | `9` |
| Losses | Total losses | `3` |
| Win Rate | Win percentage | `0.75` |
| Deck | 8-card deck list | `Hog Rider, Fireball...` |

### 5. Analysis Export

**Filename**: `analysis_[TIMESTAMP]_[PLAYER].csv`

**Fields**:

| Column | Description | Example |
|--------|-------------|---------|
| Card Name | Card name | `Hog Rider` |
| Level | Current level | `8` |
| Max Level | Maximum level | `13` |
| Rarity | Card rarity | `Rare` |
| Priority | Upgrade priority score | `8.5` |
| Role | Strategic role | `win_condition` |

## Usage Examples

### Go CLI Usage

```bash
# Export player data to CSV
./go/bin/cr-api player --tag '#PLAYERTAG' --export-csv

# Export card database to CSV
./go/bin/cr-api cards --export-csv

# Export analysis to CSV
./go/bin/cr-api analyze --tag '#PLAYERTAG' --export-csv

# Export all data
./go/bin/cr-api export all '#PLAYERTAG'

# Export specific data types
./go/bin/cr-api export player '#PLAYERTAG'
./go/bin/cr-api export cards
./go/bin/cr-api export events '#PLAYERTAG'
```

### Task Runner Usage

```bash
# Export with default player tag
task export-csv -- '#PLAYERTAG'
task export-all -- '#PLAYERTAG'

# Scan and export events
task scan-events -- '#PLAYERTAG'
task export-events -- '#PLAYERTAG'
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

- **Empty Values**: Empty string
- **Null Values**: Empty string
- **Quotes**: Escaped with backslash
- **Newlines**: Replaced with space
- **Commas**: Field is quoted if contains comma

## Integration Examples

### Excel/Google Sheets

Import CSV files using "Import from file" with these settings:
- Delimiter: comma
- Encoding: UTF-8
- First row as headers

### Python/Pandas

```python
import pandas as pd
df = pd.read_csv('data/csv/players/player_20240115_123456.csv')
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
| Encoding issues | Special characters | Ensure UTF-8 encoding |

## Best Practices

1. **Regular Backups**: Export data regularly for backup
2. **Compression**: Compress old CSV files to save space
3. **Validation**: Validate CSV files after export
4. **Documentation**: Document custom export formats
5. **Privacy**: Remove or anonymize sensitive data when sharing
6. **Retention**: Establish data retention policies
