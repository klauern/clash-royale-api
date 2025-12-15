# Clash Royale API Data Collector

A comprehensive Python tool for collecting, analyzing, and tracking Clash Royale card data, player statistics, and wild card information using the official Clash Royale API.

## Features

- 🎴 **Complete Card Database**: Access all Clash Royale cards with detailed statistics
- 👤 **Player Profile Analysis**: Comprehensive player data including card collections
- 🃏 **Wild Card Tracking**: Monitor wild card inventory and usage
- 📊 **Collection Analysis**: Detailed statistics on card levels, rarities, and upgrade needs
- 💾 **Data Persistence**: Save and track historical data over time
- 🔄 **Rate Limiting**: Built-in rate limiting to respect API limits
- 📈 **Upgrade Priority**: Intelligent analysis of which cards to upgrade next

## Project Structure

```
clash-royale-api/
├── .env.example               # Example configuration
├── config/                    # Configuration directory
├── src/
│   └── clash_royale_api/     # Python package
│       ├── __init__.py       # Package initialization
│       ├── api.py            # Main API client and analysis tools
│       └── cli.py            # Command line interface
├── data/
│   ├── static/               # Static game data (cards, etc.)
│   ├── players/              # Player profiles and data
│   └── analysis/             # Collection analysis results
├── tests/                    # Test files
├── docs/                     # Additional documentation
├── pyproject.toml           # Project configuration and dependencies
├── LICENSE                  # MIT License
└── README.md                # This file
```

## Setup

### 1. Get Your API Token

1. Visit [developer.clashroyale.com](https://developer.clashroyale.com/#/)
2. Create a developer account and verify your email
3. Generate a new API key
4. Copy your API token

### 2. Configure the Project

1. Run the setup task (creates .env file automatically):
   ```bash
   task setup
   ```

2. Edit `.env` in the project root and add your API token:
   ```env
   CLASH_ROYALE_API_TOKEN=your_api_token_here
   DEFAULT_PLAYER_TAG=#your_player_tag_here  # Optional, allows running tasks without arguments
   ```

### 3. Install Dependencies

This project uses **uv** for dependency management:

```bash
# Install all dependencies
uv sync

# Install with optional dependencies for development
uv sync --extra dev

# Install with all extras (dev + analysis + export)
uv sync --extra all
```

## Usage

### Using Task (Recommended)

First, install Task from https://taskfile.dev/installation/ or run:
```bash
curl -sL https://taskfile.dev/install.sh | sh
```

Then use these common commands:

```bash
# Show all available tasks
task

# Set up the project
task setup

# Analyze a player (works without arguments if DEFAULT_PLAYER_TAG is set in .env)
task run -- #PLAYER_TAG  # Optional: specify player tag
task run               # Uses DEFAULT_PLAYER_TAG from .env

# Save results to JSON
task run-with-save -- #PLAYER_TAG
task run-with-save

# Export to CSV
task export-csv -- #PLAYER_TAG
task export-csv

# Export all data types
task export-all -- #PLAYER_TAG
task export-all

# Event deck operations
task scan-events          # Uses DEFAULT_PLAYER_TAG from .env
task export-events        # Uses DEFAULT_PLAYER_TAG from .env
task analyze-events       # Uses DEFAULT_PLAYER_TAG from .env

# Development tasks
task lint          # Run linting
task format        # Format code
task clean         # Clean generated files
task status        # Show project status
```

### Direct CLI Usage

```bash
# Using uv (recommended)
uv run python src/clash_royale_api/cli.py --player #PLAYERTAG --save

# Basic player analysis
uv run python src/clash_royale_api/cli.py --player #PLAYERTAG

# Export to CSV
uv run python src/clash_royale_api/cli.py --player #PLAYERTAG --export-csv

# Save results
uv run python src/clash_royale_api/cli.py --player #PLAYERTAG --save --format json

# Auto-build a 1v1 ladder deck (uses latest saved analysis if present)
uv run python src/clash_royale_api/cli.py --player #PLAYERTAG --build-ladder-deck --save-deck
uv run python src/clash_royale_api/cli.py --player #PLAYERTAG --build-ladder-deck --analysis-file data/analysis/20251208_174559_analysis_R8QGUQRCV.json
```

### Python API

```python
import asyncio
from clash_royale_api import ClashRoyaleAPI

async def main():
    # Use async context manager for proper resource management
    async with ClashRoyaleAPI() as api:
        # Get all available cards
        cards = await api.get_all_cards()
        print(f"Total cards: {len(cards['items'])}")

        # Get player information with card needs
        player_tag = "#ABC123"  # Replace with actual player tag
        player_info = await api.get_player_info(player_tag, include_card_needs=True)

        # Get complete player profile with analysis
        profile = await api.get_complete_player_profile(player_tag, include_card_analysis=True)

        # Analyze card collection
        analysis = await api.analyze_card_collection(player_tag)
        print(f"Cards needing upgrade: {len(analysis['cards_needing_upgrade'])}")

# Run the async main function
asyncio.run(main())
```

### Key Methods

- `get_all_cards()`: Fetch all available cards
- `get_player_info(player_tag)`: Get comprehensive player data
- `get_player_upcoming_chests(player_tag)`: Get upcoming chests
- `get_player_card_collection(player_tag)`: Get detailed card collection
- `analyze_card_collection(player_tag)`: Analyze collection with upgrade priorities
- `get_complete_player_profile(player_tag)`: Get everything in one call

## Data Structure

### Player Card Collection
Each card in your collection includes:
- `name`: Card name
- `level`: Current level
- `count`: Number of cards owned
- `rarity`: Card rarity (Common, Rare, Epic, Legendary, Champion)
- `max_level`: Maximum achievable level

### Analysis Output
The analysis provides:
- Total card count
- Rarity breakdown
- Max level cards
- Priority list for upgrades
- Cards needed for next levels

## Environment Variable Configuration

The project uses a `.env` file in the project root for configuration. All CLI arguments can be configured via environment variables, allowing you to run tasks without repeatedly typing arguments.

### Required Configuration
```env
# API Token (required)
CLASH_ROYALE_API_TOKEN=your_api_token_here
```

### Optional Configuration
```env
# Default Player Tag (allows running tasks without arguments)
DEFAULT_PLAYER_TAG=#PLAYERTAG

# API Configuration
API_BASE_URL=https://api.clashroyale.com/v1
REQUEST_DELAY=1  # Seconds between requests
MAX_RETRIES=3

# Data Storage
DATA_DIR=./data
EXPORT_FORMAT=json  # json, csv, or both

# CSV Export Settings
CSV_DIR=./data/csv  # Output directory for CSV files
CSV_TYPES=all  # Types to export: all,player,cards,battles,chests
BATTLE_LIMIT=100  # Number of recent battles to export

# Event Analysis Settings
DAYS_BACK=7  # Number of days to scan for event decks
EVENT_OUTPUT=event_decks  # Output filename for event decks

# Output Format (for CLI display)
OUTPUT_FORMAT=table  # json, table
```

### Priority System
The CLI uses the following priority for configuration:
1. **CLI arguments** (highest priority)
2. **Environment variables**
3. **Default values** (lowest priority)

This means you can set defaults in `.env` and override them when needed with CLI arguments.

## Rate Limiting

The client includes built-in rate limiting to respect the API's limits:
- Default: 1 second between requests
- Automatic retry with exponential backoff
- Configurable delay and retry limits

## Examples

### Track Multiple Players

```python
players = ["#PLAYER1", "#PLAYER2", "#PLAYER3"]

for player_tag in players:
    try:
        analysis = api.analyze_card_collection(player_tag)
        print(f"{player_tag}: {len(analysis['max_level_cards'])} max level cards")
    except Exception as e:
        print(f"Error analyzing {player_tag}: {e}")
```

### Save Historical Data

```python
import datetime
import json

# Save player data with timestamp
player_data = api.get_player_info(player_tag)
timestamp = datetime.datetime.now().isoformat()

historical_data = {
    "timestamp": timestamp,
    "player_data": player_data
}

# Save to historical tracking file
with open(f"data/history/{player_tag}_{timestamp}.json", "w") as f:
    json.dump(historical_data, f, indent=2)
```

## API Endpoints Used

- `GET /cards` - All available cards
- `GET /players/{tag}` - Player information
- `GET /players/{tag}/upcomingchests` - Upcoming chests
- `GET /players/{tag}/chestcycle` - Chest cycle

## Error Handling

The client includes comprehensive error handling:
- Invalid API tokens
- Rate limiting
- Network issues
- Invalid player tags
- API downtime

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add your improvements
4. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

This project is for educational and personal use. Please respect Supercell's Terms of Service when using the Clash Royale API.

## Support

- [Official API Documentation](https://developer.clashroyale.com/#/documentation)
- [Clash Royale API Discord](https://discord.gg/clashroyale)
- Issues: Create an issue in this repository

## Changelog

### v1.0.0
- Initial release
- Complete card database access
- Player profile analysis
- Wild card tracking
- Upgrade priority analysis
- Rate limiting and error handling
