# Clash Royale API Data Collector

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![GitHub Releases](https://img.shields.io/github/v/release/klauern/clash-royale-api)](https://github.com/klauern/clash-royale-api/releases)

A comprehensive Go tool for collecting, analyzing, and tracking Clash Royale card data, player statistics, event decks, and intelligent deck building using the official Clash Royale API.

## Quickstart

```bash
# 1. Get your API token from https://developer.clashroyale.com/
# 2. Download the latest release for your platform
# 3. Configure
cp .env.example .env
# Edit .env and add your API token

# 4. Run
./cr-api analyze --tag YOUR_PLAYER_TAG
```

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
├── cmd/                      # Command-line applications
│   └── cr-api/              # Main CLI application
├── pkg/                      # Go libraries
│   ├── clashroyale/          # API client
│   ├── analysis/             # Collection analysis & playstyle
│   ├── deck/                 # Deck building algorithms
│   └── events/               # Event deck tracking
├── internal/                 # Internal packages
│   ├── exporter/             # CSV export
│   └── storage/              # Data persistence
├── bin/                      # Built binaries
├── data/                    # Data storage
│   ├── evolution_shards.json # Evolution shard inventory
│   ├── static/              # Static game data (cards.json cache)
│   ├── players/             # Player profiles
│   ├── analysis/            # Collection analysis
│   ├── csv/                 # CSV exports
│   └── event_decks/         # Event deck tracking
├── scripts/                 # Utility scripts
├── LICENSE                  # MIT License
└── README.md                # This file
```

## Security Notice

**⚠️ Important**: Never commit your `.env` file or API tokens to version control.

The `.env` file contains sensitive credentials and is **gitignored** by design. This repository includes `.env.example` as a safe template with placeholder values.

**Local-only data**: The `data/` directory stores cached API responses, analysis results, and local artifacts. It is excluded from version control to keep the repository clean and avoid committing large generated files.

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

3. **Install** binary to your PATH:
   ```bash
   # Unix/macOS (system-wide)
   sudo mv cr-api /usr/local/bin/

   # Unix/macOS (user-only)
   mv cr-api ~/.local/bin/

   # Windows: Add directory to PATH via System Properties
   ```

4. **(Optional) Install shell completions**:
   ```bash
   # Bash
   cp completions/cr-api.bash ~/.local/share/bash-completion/completions/cr-api

   # Zsh
   mkdir -p ~/.zsh/completions
   cp completions/cr-api.zsh ~/.zsh/completions/

   # Fish
   cp completions/cr-api.fish ~/.config/fish/completions/
   ```

5. **Verify installation**:
   ```bash
   cr-api --version
   ```

### Build from Source

Requires Go 1.22+ and Task runner:

```bash
# Clone the repository
git clone https://github.com/klauern/clash-royale-api.git
cd clash-royale-api

# Build binaries
task build

# Binaries will be in bin/
./bin/cr-api --version
```

## Quick Start (Go)

### 1. Get Your API Token

1. Visit [developer.clashroyale.com](https://developer.clashroyale.com/#/)
2. Create a developer account and verify your email
3. Generate a new API key
4. Copy your API token

### 2. Configure the Project

Copy the example environment file and add your API token:

```bash
# Copy the template to create your local .env file
cp .env.example .env

# Edit .env and add your actual API token
nano .env  # or use your preferred editor
```

Edit `.env` in the project root and add your API token:

```env
CLASH_ROYALE_API_TOKEN=your_actual_api_token_here
DEFAULT_PLAYER_TAG=#your_player_tag_here  # Optional
```

**⚠️ Remember**: The `.env` file is gitignored and should never be committed to version control. Only `.env.example` (which contains safe placeholders) is tracked in the repository.

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
task build
# or: go build -o bin/cr-api ./cmd/cr-api

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

# Evolution shard inventory
./bin/cr-api evolutions shards list
./bin/cr-api evolutions shards list --card "Archers"
./bin/cr-api evolutions shards set --card "Archers" --count 3
./bin/cr-api cards                                  # Refresh cached card database for validation

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

Evolution shard commands validate card names against the cached card database at `data/static/cards.json`. Run `cr-api cards` to refresh the cache (or pass `--api-token` to fetch on demand).

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

    "github.com/klauer/clash-royale-api/pkg/clashroyale"
    "github.com/klauer/clash-royale-api/pkg/analysis"
    "github.com/klauer/clash-royale-api/pkg/deck"
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

- `github.com/klauer/clash-royale-api/pkg/clashroyale`: API client
- `github.com/klauer/clash-royale-api/pkg/analysis`: Collection analysis
- `github.com/klauer/clash-royale-api/pkg/deck`: Deck building algorithms
- `github.com/klauer/clash-royale-api/pkg/events`: Event deck tracking

#### Go Modules

Import the Go modules in your project:

```bash
go get github.com/klauer/clash-royale-api/pkg/clashroyale
go get github.com/klauer/clash-royale-api/pkg/analysis
go get github.com/klauer/clash-royale-api/pkg/deck
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

## Upgrade Projections and Wildcards

**Important**: The Clash Royale API does not provide wildcard counts. You must manually enter your wildcard inventory for upgrade projections.

### Wildcard Configuration

Wildcards are special items that can substitute any card of a specific rarity. They are tracked separately from regular cards.

To use upgrade projections with wildcards, create an `upgrade_plan.json` file:

```json
{
  "wildcards": {
    "Common": 1200,
    "Rare": 400,
    "Epic": 60,
    "Legendary": 8,
    "Champion": 3
  },
  "upgrades": [
    {
      "card": "Hog Rider",
      "target_level": 11
    },
    {
      "card": "Fireball",
      "target_level": 10
    }
  ]
}
```

**Rarity values** (case-insensitive): `Common`, `Rare`, `Epic`, `Legendary`, `Champion`

### Running Upgrade Projections

```bash
# 1. Generate an analysis file
./bin/cr-api analyze --tag PLAYER_TAG --save

# 2. Apply upgrade plan with wildcards
./scripts/project_deck_projection.py \
  --analysis ./data/analysis/ANALYSIS_FILE.json \
  --plan ./data/upgrade_plan.json \
  --tag PLAYER_TAG

# 3. Build deck from projected analysis (offline mode)
./bin/cr-api deck build --tag PLAYER_TAG --from-analysis PROJECTED_FILE.json
```

### Options

- `--unbounded`: Skip wildcard affordability checks (allows any upgrades)
- `--dry-run`: Show planned changes without writing output
- `--output`: Specify output path for projected analysis

**See** `data/upgrade_plan_example.json` for a complete template.

## Environment Variable Configuration

The project uses a `.env` file in the project root for configuration. All CLI arguments can be configured via environment variables, allowing you to run tasks without repeatedly typing arguments.

**⚠️ Security**: Copy `.env.example` to `.env` and add your actual values. The `.env` file is gitignored and should never be committed. Only `.env.example` (with safe placeholders) is tracked in the repository.

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

## Testing

The project uses Go's built-in testing framework with support for both unit and integration tests.

### Unit Tests

Unit tests run in CI and require no external dependencies:

```bash
# Run unit tests
task test-go
# or: go test -v -race -tags='!integration' ./...

# Run with coverage
task test-go
# Coverage report generated at coverage.html
```

### Integration Tests

Integration tests connect to the live Clash Royale API and require a valid API token:

```bash
# Run integration tests (requires CLASH_ROYALE_API_TOKEN in .env)
task test-integration
# or: go test -v -race -tags=integration ./...
```

**Note**: Integration tests are excluded from CI due to Clash Royale API IP restrictions. See [CI/CD](#ci--cd-limitations) below for details.

### Test Structure

- **Unit tests** (`*_test.go`): Test individual functions and packages in isolation
- **Integration tests** (`integration_test.go`, `integration_evolution_test.go`): Test end-to-end flows with live API

Integration tests are built with the `//go:build integration` tag, ensuring they only run when explicitly enabled.

## CI/CD Limitations

### Clash Royale API IP Restrictions

The official Clash Royale API has significant limitations that affect CI/CD integration:

**IP Whitelisting Requirements:**
- **Maximum 5 IP addresses** per API key
- **Static IPs only** - must be pre-allowlisted at [developer.clashroyale.com](https://developer.clashroyale.com/)
- **No "allow all" option** - dynamic IPs cannot be used

**Rate Limits (Undocumented):**
- Exact limits are not publicly disclosed by Supercell
- Community reports suggest **10-20 requests/second** may trigger blocking
- Exceeding limits can result in **24-hour IP bans**

**GitHub Actions Compatibility:**
| Runner Type | Static IP | Viable for CR API |
|-------------|-----------|-------------------|
| Standard (`ubuntu-latest`) | ❌ Dynamic | No |
| Standard (`windows-latest`) | ❌ Dynamic | No |
| Standard (`macos-latest`) | ❌ Dynamic | No |
| Larger Runners (Enterprise) | ✅ Configurable | Yes (requires Enterprise) |
| Self-Hosted | ✅ Your IP | Yes |

### Recommended Approaches

**Option 1: Manual Integration Tests (Current Approach)** ✅
```bash
# Run locally before pushing changes
task test-integration
```
- **Pros**: Simple, no additional infrastructure, free
- **Cons**: Manual step, relies on developer diligence

**Option 2: Self-Hosted GitHub Actions Runner** ⭐ Recommended for Teams
```yaml
# Deploy a self-hosted runner with a static IP
# Allowlist the runner's IP at developer.clashroyale.com
```
- **Pros**: Automated integration tests, full control
- **Cons**: Requires server maintenance ($5-20/month for VPS)

**Option 3: GitHub Enterprise Larger Runners** 💰
- **Pros**: Managed service, static IPs available
- **Cons**: Requires GitHub Enterprise Cloud (~$21/user/month)

**Option 4: Proxy Service** 🔀
- Use services that provide static IP forwarding
- **Cons**: Additional dependency, cost, potential latency

### Current Strategy

This project uses **Option 1** (manual testing) for public CI:
- CI runs unit tests only (no API calls)
- Integration tests run locally via `task test-integration`
- This keeps the repository public and CI free without API costs

**Sources:**
- [About GitHub's IP addresses](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-githubs-ip-addresses)
- [Larger runners reference](https://docs.github.com/en/actions/reference/runners/larger-runners)
- [GitHub Actions IP discussion](https://github.com/orgs/community/discussions/26442)
- [Clash Royale Developer Portal](https://developer.clashroyale.com/)

## Releases

This project uses [GoReleaser](https://goreleaser.com/) for automated releases.

### Release Process

Releases are created automatically when you push a version tag:

```bash
# 1. Ensure all changes are committed and tests pass
task test && task lint

# 2. Create and push a semantic version tag
git tag -a v2.0.1 -m "Release v2.0.1: Bug fixes and improvements"
git push origin v2.0.1
```

GitHub Actions will:
- Build binaries for Linux, macOS, and Windows (amd64/arm64)
- Generate a changelog from conventional commits
- Create a GitHub Release with all artifacts
- Generate shell completion files

### Version Tags

- Format: `vX.Y.Z` (Semantic Versioning)
- `X` = Major (breaking changes)
- `Y` = Minor (new features)
- `Z` = Patch (bug fixes)

### Commit Prefixes

| Prefix | Section |
|--------|---------|
| `feat:` | Features |
| `fix:` | Bug Fixes |
| `perf:`, `refactor:` | Improvements |

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

## Troubleshooting

### "API token not found" Error

**Problem**: The tool reports that `CLASH_ROYALE_API_TOKEN` is not configured.

**Solution**:
```bash
# Copy the example environment file
cp .env.example .env

# Edit .env and add your actual API token from developer.clashroyale.com
nano .env
```

### "Invalid API token" Error

**Problem**: API returns 403 or 401 errors.

**Solutions**:
1. Verify your token is correct (no extra spaces)
2. Check that your IP is allowlisted at [developer.clashroyale.com](https://developer.clashroyale.com/)
3. Ensure the token hasn't expired

### Rate Limiting Issues

**Problem**: Requests are failing or getting blocked.

**Solutions**:
1. Increase `REQUEST_DELAY` in `.env` (default: 1 second)
2. Reduce the frequency of requests
3. Check if you've exceeded the undocumented rate limits (~10-20 req/sec)

### Build Failures

**Problem**: `go build` fails with module errors.

**Solution**:
```bash
# Download dependencies
go mod download

# Verify go.mod is consistent
go mod tidy

# Try building again
go build -o bin/cr-api ./cmd/cr-api
```

### Permission Denied on Binary

**Problem**: `./bin/cr-api` fails with "permission denied".

**Solution**:
```bash
chmod +x bin/cr-api
```

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
