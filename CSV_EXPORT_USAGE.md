# Clash Royale CSV Export Tool Usage Guide

## Overview

The Clash Royale CSV Export Tool allows you to export player data, card collections, battle logs, and other information from the Clash Royale API into structured CSV files. These CSV files can be easily imported into Excel, Google Sheets, or analyzed with data analysis tools.

## Setup

1. **Ensure your API token is configured**:
   Copy `.env.example` to `.env` and add your Clash Royale API token:
   ```
   CLASH_ROYALE_API_TOKEN=your_api_token_here
   ```

2. **Install dependencies**:
   ```bash
   uv pip install -e .
   ```

## Usage Examples

### Command Line Interface

#### Export all data for a player:
```bash
# Export all available data types
uv run src/clash_royale_api/cli.py -p "#YOUR_PLAYER_TAG" --export-csv

# Export specific data types only
uv run src/clash_royale_api/cli.py -p "#YOUR_PLAYER_TAG" --export-csv --csv-types player,cards

# Export with custom output directory
uv run src/clash_royale_api/cli.py -p "#YOUR_PLAYER_TAG" --export-csv --csv-dir ~/my_csv_exports

# Export limited battle history
uv run src/clash_royale_api/cli.py -p "#YOUR_PLAYER_TAG" --export-csv --csv-types battles --battle-limit 50
```

#### Export card database:
```bash
# Export the complete card database (reference file)
uv run src/clash_royale_api/cli.py --cards --export-csv
```

### Programmatic Usage

```python
import asyncio
from clash_royale_api import ClashRoyaleAPI, CSVExporter

async def export_data():
    # Initialize API client
    api = ClashRoyaleAPI()

    # Method 1: Use built-in CSV export methods
    await api.export_player_info_csv("#PLAYER_TAG")
    await api.export_card_collection_csv("#PLAYER_TAG")
    await api.export_battle_log_csv("#PLAYER_TAG", limit=100)
    await api.export_chest_cycle_csv("#PLAYER_TAG")

    # Method 2: Use CSV exporter directly
    exporter = CSVExporter(api, output_dir="./my_exports")

    # Export specific data
    await exporter.export_player_info_csv("#PLAYER_TAG")

    # Export all data at once
    exported_files = await exporter.export_all_data_csv("#PLAYER_TAG")
    print(f"Exported {len(exported_files)} files")

    # Export card database
    await exporter.export_card_database_csv()

# Run the export
asyncio.run(export_data())
```

## Exported CSV Files

### 1. Player Information (`player_info_[tag]_[date].csv`)
Contains comprehensive player profile data:
- `player_tag` - Unique player identifier
- `name` - Player name
- `exp_level` - Experience level
- `trophies` - Current trophy count
- `best_trophies` - Highest trophies achieved
- `wins`, `losses`, `battle_count` - Battle statistics
- `win_rate` - Percentage of wins
- `clan_tag`, `clan_name`, `clan_role` - Clan information
- `arena_name`, `arena_id` - Current arena
- `current_deck_1` through `current_deck_8` - Cards in current deck
- `deck_avg_elixir` - Average elixir cost of current deck
- `fetch_time` - When data was retrieved

### 2. Card Collection (`card_collection_[tag]_[date].csv`)
Detailed card collection information:
- `player_tag` - Player identifier
- `card_name` - Name of the card
- `level` - Current level
- `count` - Number of cards owned
- `max_level` - Maximum possible level
- `rarity` - Card rarity (Common, Rare, Epic, Legendary)
- `elixir_cost` - Elixir cost to play
- `card_type` - Type (Troop, Spell, Building)
- `arena_required` - Arena needed to unlock
- `is_max_level` - Whether card is at max level

### 3. Battle Log (`battle_log_[tag]_[date].csv`)
Recent battle history:
- `battle_time_utc` - Battle timestamp
- `game_mode` - Type of battle (Ladder, Tournament, etc.)
- `team_crowns`, `opponent_crowns` - Crowns earned
- `result` - Win or loss
- `trophy_change` - Trophy gain/loss
- `opponent_name`, `opponent_tag`, `opponent_trophies` - Opponent info
- `team_card_1` through `team_card_8` - Cards used with levels
- `opponent_card_1` through `opponent_card_8` - Opponent's cards (if available)

### 4. Chest Cycle (`chest_cycle_[tag]_[date].csv`)
Upcoming chest information:
- `chest_position` - Position in chest cycle
- `chest_name` - Type of chest
- `time_to_open_hours` - Hours until chest opens
- `chest_rarity` - Classified rarity

### 5. Card Database (`card_database_[date].csv`)
Complete reference of all cards:
- `card_id` - Unique card identifier
- `name` - Card name
- `rarity`, `type`, `elixir_cost` - Basic properties
- `arena_required` - Arena to unlock
- `max_level` - Maximum level
- `description` - Card description

## File Organization

CSV files are organized in subdirectories:
```
data/csv/
├── players/          # Player-specific exports
├── clans/            # Clan information (when implemented)
├── analysis/         # Analytical reports
└── reference/        # Reference data like card database
```

## Tips for Data Analysis

1. **Import into Excel/Google Sheets**:
   - Open CSV files directly
   - Use PivotTables for analysis
   - Create charts for visualization

2. **Python Analysis**:
   ```python
   import pandas as pd

   # Load player's card collection
   cards = pd.read_csv('data/csv/players/card_collection_TAG_date.csv')

   # Filter cards needing upgrades
   upgrade_needed = cards[cards['is_max_level'] == False]

   # Group by rarity
   rarity_counts = cards['rarity'].value_counts()
   ```

3. **Track Progress Over Time**:
   - Export data regularly
   - Use timestamps to track changes
   - Compare trophy progression
   - Monitor card collection growth

## Error Handling

The tool includes robust error handling:
- Invalid player tags are caught and reported
- API rate limits are respected
- Failed exports are logged but don't stop other exports
- Missing data fields are filled with empty values

## Performance Considerations

- Battle logs can be large; use `--battle-limit` to restrict
- Multiple exports are processed sequentially to respect API limits
- Card database is cached after first export for subsequent uses

## Troubleshooting

### Common Issues:
1. **"API token not found"** - Check your `.env` file configuration
2. **"Invalid player tag"** - Ensure tag starts with `#` and is correct
3. **"Rate limit exceeded"** - Wait a moment between large requests
4. **Empty CSV files** - API might be returning test data in some environments

### Getting Help:
- Check the API documentation for valid endpoints
- Verify your API token has the required permissions
- Ensure the player is active and data is publicly available

## Future Enhancements

Planned features include:
- Clan statistics export
- Tournament results export
- Historical trend analysis
- Custom CSV templates
- Scheduled exports