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

## Documentation

- **[Deck Building](DECK_BUILDER.md)** - Algorithm details and Go API examples
- **[Evolution System](EVOLUTION.md)** - Evolution mechanics and configuration
- **[Event Tracking](EVENT_TRACKING.md)** - Event deck analysis
- **[CSV Exports](CSV_EXPORTS.md)** - Export functionality
- **[Deck Strategies](DECK_STRATEGIES.md)** - Playstyle analysis and recommendations

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

**Binary Releases** (Recommended): Download from [Releases page](https://github.com/klauern/clash-royale-api/releases) for Linux, macOS, or Windows. Extract and move `cr-api` to your PATH.

**Build from Source**: Requires Go 1.22+
```bash
git clone https://github.com/klauern/clash-royale-api.git
cd clash-royale-api
task build
```

## Quick Start

1. Get API token from [developer.clashroyale.com](https://developer.clashroyale.com)
2. Create `.env` file: `cp .env.example .env` and add your token
3. Build: `task build`
4. Run: `./bin/cr-api analyze --tag YOUR_TAG`

## CLI Usage

```bash
# Common commands (use --help for all options)
./bin/cr-api player --tag PLAYER_TAG [--chests] [--save] [--export-csv]
./bin/cr-api analyze --tag PLAYER_TAG [--save] [--export-csv]
./bin/cr-api deck build --tag PLAYER_TAG [--strategy STRATEGY] [--verbose]
./bin/cr-api events scan --tag PLAYER_TAG
./bin/cr-api cards [--export-csv]

# Deck building strategies: balanced (default), aggro, control, cycle, splash, spell
./bin/cr-api deck build --tag PLAYER_TAG --strategy cycle --verbose

# Task runner (recommended)
task                     # Show all tasks
task run -- #PLAYER_TAG  # Analyze player
task test                # Run tests
```

See feature-specific documentation for detailed command options:
- [DECK_BUILDER.md](DECK_BUILDER.md) - Deck building strategies and options
- [EVENT_TRACKING.md](EVENT_TRACKING.md) - Event scanning and analysis
- [CSV_EXPORTS.md](CSV_EXPORTS.md) - Export formats and options
- [EVOLUTION.md](EVOLUTION.md) - Evolution shard management

## Using as a Go Library

For complete Go API examples, integration patterns, and package documentation, see [DECK_BUILDER.md](DECK_BUILDER.md).

## Data Structure

Card collections include name, level, rarity, count, and max_level. Analysis provides upgrade priorities, rarity breakdowns, and max-level card tracking. For evolution mechanics and configuration, see [EVOLUTION.md](EVOLUTION.md).

## Configuration

**⚠️ Security**: Copy `.env.example` to `.env` and add your actual values. Never commit `.env` to version control.

**Required**: `CLASH_ROYALE_API_TOKEN` (get from [developer.clashroyale.com](https://developer.clashroyale.com/))

**Optional**: `DEFAULT_PLAYER_TAG`, `DATA_DIR`, `REQUEST_DELAY`, and more. See `.env.example` for all options.

**Priority**: CLI arguments override environment variables, which override defaults.

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

```bash
task test              # Run unit tests with coverage
task test-integration  # Run integration tests (requires API token)
```

Unit tests run in CI. Integration tests connect to the live API and are excluded from CI due to IP restrictions (see CI/CD section below). See [AGENTS.md](AGENTS.md) for complete testing details.

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

### Current Strategy

CI runs unit tests only. Integration tests require manual execution (`task test-integration`) due to IP restrictions. Options for automation: self-hosted runners (static IP), GitHub Enterprise larger runners, or proxy services.

**Sources:**
- [About GitHub's IP addresses](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-githubs-ip-addresses)
- [Larger runners reference](https://docs.github.com/en/actions/reference/runners/larger-runners)
- [GitHub Actions IP discussion](https://github.com/orgs/community/discussions/26442)
- [Clash Royale Developer Portal](https://developer.clashroyale.com/)

## Releases

Releases are automated via GitHub Actions when pushing version tags (`vX.Y.Z`). Binaries are built for Linux, macOS, and Windows. See [Releases](https://github.com/klauern/clash-royale-api/releases) for downloads.

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

**API token errors**: Copy `.env.example` to `.env` and add your token. Verify your IP is allowlisted at [developer.clashroyale.com](https://developer.clashroyale.com/).

**Rate limiting**: Increase `REQUEST_DELAY` in `.env` or reduce request frequency.

**Build failures**: Run `go mod download && go mod tidy`

**Permission denied**: Run `chmod +x bin/cr-api`

## Changelog

See [GitHub Releases](https://github.com/klauern/clash-royale-api/releases) for version history and release notes.
