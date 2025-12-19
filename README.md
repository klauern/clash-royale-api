# Clash Royale API Data Collector

A comprehensive Go tool for collecting, analyzing, and tracking Clash Royale card data, player statistics, event decks, and intelligent deck building using the official Clash Royale API.

## Features

- 🎴 **Complete Card Database**: Access all Clash Royale cards with detailed statistics
- 👤 **Player Profile Analysis**: Comprehensive player data including card collections
- 🏗️ **Intelligent Deck Building**: AI-powered deck recommendations based on your collection
- 📊 **Collection Analysis**: Detailed statistics on card levels, rarities, and upgrade priorities
- 🎮 **Playstyle Analysis**: Analyze player's playstyle and get personalized deck recommendations
- 🃏 **Event Deck Tracking**: Monitor and analyze performance in special events
- 💾 **Data Persistence**: Save and track historical data over time
- 📈 **CSV Export**: Export player data, card collections, and event statistics
- 🔄 **Rate Limiting**: Built-in rate limiting to respect API limits
- ⚡ **High Performance**: Go implementation provides superior performance and type safety

## Project Structure

```
clash-royale-api/
├── .env.example               # Example configuration
├── Taskfile.yml              # Task runner configuration
├── go/                       # Go implementation
│   ├── cmd/
│   │   ├── cr-api/          # Main CLI application
│   │   └── deckbuilder/     # Standalone deck builder
│   ├── pkg/                 # Go libraries
│   │   ├── clashroyale/     # API client
│   │   ├── analysis/        # Collection analysis & playstyle
│   │   ├── deck/            # Deck building algorithms
│   │   └── events/          # Event deck tracking
│   ├── internal/            # Internal packages
│   │   ├── exporter/        # CSV export
│   │   └── storage/         # Data persistence
│   └── bin/                 # Built binaries
├── data/                    # Data storage
│   ├── static/              # Static game data
│   ├── players/             # Player profiles
│   ├── analysis/            # Collection analysis
│   ├── csv/                 # CSV exports
│   └── event_decks/         # Event deck tracking
├── scripts/                 # Utility scripts
├── LICENSE                  # MIT License
└── README.md                # This file
```

## Installation

### Binary Releases (Recommended)

Download pre-built binaries from the [Releases page](https://github.com/klauern/clash-royale-api/releases):

1. **Download** the appropriate archive for your platform:
   - Linux (amd64): `clash-royale-api_vX.X.X_linux_amd64.tar.gz`
   - Linux (arm64): `clash-royale-api_vX.X.X_linux_arm64.tar.gz`
   - macOS (Intel): `clash-royale-api_vX.X.X_darwin_amd64.tar.gz`
   - macOS (Apple Silicon): `clash-royale-api_vX.X.X_darwin_arm64.tar.gz`
   - Windows: `clash-royale-api_vX.X.X_windows_amd64.zip`

2. **Extract** the archive:
   ```bash
   # Unix/macOS
   tar -xzf clash-royale-api_vX.X.X_*.tar.gz

   # Windows (PowerShell)
   Expand-Archive clash-royale-api_vX.X.X_windows_amd64.zip
   ```

3. **Install** binaries to your PATH:
   ```bash
   # Unix/macOS (system-wide)
   sudo mv cr-api deckbuilder /usr/local/bin/

   # Unix/macOS (user-only)
   mv cr-api deckbuilder ~/.local/bin/

   # Windows: Add directory to PATH via System Properties
   ```

4. **(Optional) Install shell completions**:
   ```bash
   # Bash
   cp completions/cr-api.bash ~/.local/share/bash-completion/completions/cr-api
   cp completions/deckbuilder.bash ~/.local/share/bash-completion/completions/deckbuilder

   # Zsh
   mkdir -p ~/.zsh/completions
   cp completions/*.zsh ~/.zsh/completions/

   # Fish
   cp completions/*.fish ~/.config/fish/completions/
   ```

5. **Verify installation**:
   ```bash
   cr-api --version
   deckbuilder --version
   ```

### Build from Source

Requires Go 1.22+ and Task runner:

```bash
# Clone the repository
git clone https://github.com/klauern/clash-royale-api.git
cd clash-royale-api

# Build binaries
task build

# Binaries will be in go/bin/
./go/bin/cr-api --version
```

## Quick Start (Go)

### 1. Get Your API Token

1. Visit [developer.clashroyale.com](https://developer.clashroyale.com/#/)
2. Create a developer account and verify your email
3. Generate a new API key
4. Copy your API token

### 2. Configure the Project

Run the setup task (creates .env file automatically):

```bash
task setup
```

Edit `.env` in the project root and add your API token:

```env
CLASH_ROYALE_API_TOKEN=your_api_token_here
DEFAULT_PLAYER_TAG=#your_player_tag_here  # Optional
```

### 3. Build and Run

```bash
# Build Go binaries
task build-go

# Analyze a player
./bin/cr-api analyze --tag PLAYER_TAG

# Build a deck
./bin/cr-api deck build --tag PLAYER_TAG

# Export to CSV
./bin/cr-api player --tag PLAYER_TAG --export-csv
```

## CLI Usage

### Main Commands

```bash
# Build binaries (first time only)
task build-go
# or: cd go && go build -o bin/cr-api ./cmd/cr-api

# Player information
./bin/cr-api player --tag PLAYER_TAG
./bin/cr-api player --tag PLAYER_TAG --chests      # Include upcoming chests
./bin/cr-api player --tag PLAYER_TAG --save        # Save to JSON
./bin/cr-api player --tag PLAYER_TAG --export-csv  # Export to CSV

# Card database
./bin/cr-api cards                                   # List all cards
./bin/cr-api cards --export-csv                      # Export card database

# Collection analysis
./bin/cr-api analyze --tag PLAYER_TAG               # Full collection analysis
./bin/cr-api analyze --tag PLAYER_TAG --save        # Save analysis
./bin/cr-api analyze --tag PLAYER_TAG --export-csv  # Export analysis

# Deck building
./bin/cr-api deck build --tag PLAYER_TAG             # Build optimized deck
./bin/cr-api deck build --tag PLAYER_TAG --verbose   # With detailed reasoning

# Event tracking
./bin/cr-api events scan --tag PLAYER_TAG            # Scan battle log for events
./bin/cr-api events scan --tag PLAYER_TAG --days 30 --save
./bin/cr-api events list --tag PLAYER_TAG            # List tracked event decks
./bin/cr-api events analyze --tag PLAYER_TAG         # Analyze event performance

# Export utilities
./bin/cr-api export player --tag PLAYER_TAG          # Export player data
./bin/cr-api export cards                            # Export card database
./bin/cr-api export analysis --tag PLAYER_TAG        # Export collection analysis
```

### Deck Building Options

```bash
# Basic deck building
./bin/cr-api deck build --tag PLAYER_TAG

# With evolution settings
./bin/cr-api deck build --tag PLAYER_TAG --unlocked-evolutions "Archers,Knight"
./bin/cr-api deck build --tag PLAYER_TAG --evolution-slots 3

# Different strategies
./bin/cr-api deck build --tag PLAYER_TAG --strategy balanced    # Default
./bin/cr-api deck build --tag PLAYER_TAG --strategy aggressive
./bin/cr-api deck build --tag PLAYER_TAG --strategy control
```

### Task Runner (Recommended)

```bash
# Show all available tasks
task

# Common tasks
task setup              # Configure environment
task build-go           # Build Go binaries
task test-go            # Run Go tests

# Player analysis (uses DEFAULT_PLAYER_TAG from .env if no tag provided)
task run -- #PLAYER_TAG
task run-with-save -- #PLAYER_TAG
task export-csv -- #PLAYER_TAG

# Development
task lint               # Run linting
task format             # Format code
task clean              # Clean generated files
```

## Go API

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/klauer/clash-royale-api/go/pkg/clashroyale"
    "github.com/klauer/clash-royale-api/go/pkg/analysis"
    "github.com/klauer/clash-royale-api/go/pkg/deck"
)

func main() {
    // Create API client
    client := clashroyale.NewClient("your-api-token")

    // Get player information
    player, err := client.GetPlayer("#ABC123")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Player: %s (%s)\n", player.Name, player.Tag)
    fmt.Printf("Level: %d, Trophies: %d\n", player.ExpLevel, player.Trophies)

    // Analyze card collection
    analysisOptions := analysis.DefaultAnalysisOptions()
    cardAnalysis, err := analysis.AnalyzeCardCollection(player, analysisOptions)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Total cards: %d\n", cardAnalysis.TotalCards)
    fmt.Printf("Max level cards: %d\n", len(cardAnalysis.MaxLevelCards))

    // Build a deck
    builder := deck.NewBuilder("data")
    deckRec, err := builder.BuildDeckFromAnalysis(deck.CardAnalysis{
        CardLevels: convertCardLevels(cardAnalysis.CardLevels),
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Recommended deck (%.1f elixir):\n", deckRec.AvgElixir)
    for i, card := range deckRec.Deck {
        fmt.Printf("%d. %s (%d/%d)\n", i+1, card.Name, card.Level, card.MaxLevel)
    }
}

func convertCardLevels(levels map[string]analysis.CardLevelInfo) map[string]deck.CardLevelData {
    result := make(map[string]deck.CardLevelData)
    for name, info := range levels {
        result[name] = deck.CardLevelData{
            Level:             info.Level,
            MaxLevel:          info.MaxLevel,
            Rarity:            info.Rarity,
            Elixir:            info.Elixir,
            MaxEvolutionLevel: info.MaxEvolutionLevel,
        }
    }
    return result
}
```

#### Key Go Packages

- `github.com/klauer/clash-royale-api/go/pkg/clashroyale`: API client
- `github.com/klauer/clash-royale-api/go/pkg/analysis`: Collection analysis
- `github.com/klauer/clash-royale-api/go/pkg/deck`: Deck building algorithms
- `github.com/klauer/clash-royale-api/go/pkg/events`: Event deck tracking

#### Go Modules

Import the Go modules in your project:

```bash
go get github.com/klauer/clash-royale-api/go/pkg/clashroyale
go get github.com/klauer/clash-royale-api/go/pkg/analysis
go get github.com/klauer/clash-royale-api/go/pkg/deck
```

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

### Evolution Support

The Go implementation includes evolution tracking for enhanced deck building:

**Configuration (.env):**

```env
# Cards with unlocked evolutions (comma-separated)
UNLOCKED_EVOLUTIONS="Archers,Knight,Musketeer"

# Override via CLI:
./bin/cr-api deck build --tag PLAYER_TAG --unlocked-evolutions "Archers,Bomber"
```

**How it works:**

1. Cards with unlocked evolutions receive a level-scaled bonus
2. Evolution slots are prioritized (default: 2 slots)
3. Deck builder shows which cards use evolution slots in output

**Example Output:**

```
Strategic Notes:
• Evolution slots: Archers, Knight
```

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

### v2.0.0 (Current)

- **Go Implementation**: Complete Go implementation with feature parity to Python
- **Intelligent Deck Building**: AI-powered deck recommendations based on card collection
- **Evolution Support**: Track unlocked evolutions and optimize deck building
- **Event Deck Tracking**: Monitor and analyze performance in special events
- **CSV Export**: Comprehensive export functionality for all data types
- **Enhanced Performance**: Significant performance improvements with Go implementation
- **Type Safety**: Full type safety and comprehensive error handling
- **Combat Statistics**: Advanced card statistics including DPS, HP, and effectiveness metrics

### v1.0.0 (Python - Legacy)

- Initial Python release
- Complete card database access
- Player profile analysis
- Wild card tracking
- Upgrade priority analysis
- Rate limiting and error handling
