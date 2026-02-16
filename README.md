# Clash Royale API CLI

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![GitHub Releases](https://img.shields.io/github/v/release/klauern/clash-royale-api)](https://github.com/klauern/clash-royale-api/releases)

A Go CLI for collecting Clash Royale player/card data, building and evaluating decks, tracking events, and exporting results from the official Clash Royale API.

## Features

- Player profile, collection, and playstyle analysis
- Deck building, evaluation, comparison, and batch suite workflows
- Evolution-aware deck recommendations and upgrade impact analysis
- Event deck tracking and CSV export support
- Built-in API rate limiting and retry handling

## Quick Start (5 minutes)

### Prerequisites

- Go 1.24+
- [Task](https://taskfile.dev)
- Clash Royale API token from [developer.clashroyale.com](https://developer.clashroyale.com/)

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

**Binary Releases** (Recommended): Download from [Releases page](https://github.com/klauern/clash-royale-api/releases).

**Build from Source** (Go 1.26+):
```bash
git clone https://github.com/klauern/clash-royale-api.git
cd clash-royale-api && task build
```

## CLI Usage

```bash
cp .env.example .env
# Edit .env and set CLASH_ROYALE_API_TOKEN
```

### Build

```bash
task build
```

### First run

```bash
./bin/cr-api analyze --tag <PLAYER_TAG>
```

Use player tags without `#` in CLI flags.

## Common Workflows

```bash
# 1) Player profile
./bin/cr-api player --tag <PLAYER_TAG> --chests

# 2) Collection analysis
./bin/cr-api analyze --tag <PLAYER_TAG> --save

# 3) Build one deck
./bin/cr-api deck build --tag <PLAYER_TAG> --strategy balanced

# 4) Build/evaluate/compare a suite
./bin/cr-api deck build-suite --tag <PLAYER_TAG> --strategies all --variations 2
./bin/cr-api deck evaluate-batch --from-suite data/decks/<suite-file>.json --tag <PLAYER_TAG>
./bin/cr-api compare --from-evaluations data/evaluations/<eval-file>.json --auto-select-top 5

# 5) Scan events
./bin/cr-api events scan --tag <PLAYER_TAG>
```

For full flags and command surface, see [docs/CLI_REFERENCE.md](docs/CLI_REFERENCE.md).

## Taskfile Shortcuts

```bash
task setup
task run -- '#TAG'
task test
task lint
```

Use `task --list` to see all available tasks.

## Configuration

Required:

- `CLASH_ROYALE_API_TOKEN` - Clash Royale API token

Common optional settings:

- `DEFAULT_PLAYER_TAG`
- `DATA_DIR`
- `REQUEST_DELAY`
- `MAX_RETRIES`
- `UNLOCKED_EVOLUTIONS`

Configuration precedence:

1. CLI flags
2. Environment variables
3. Built-in defaults

See [`.env.example`](.env.example) for the full environment configuration.

## Data & Security

- Never commit `.env` or API tokens.
- The CLI default data directory is `~/.cr-api` (from `--data-dir` default in code).
- Task workflows commonly pass `--data-dir data`, so outputs are written to the repo-local `data/` directory during task-based usage.

## Testing & CI Reality

```bash
task test
task test-integration
```

- CI runs unit tests.
- Integration tests require live API access and are run manually because Clash Royale API keys require IP allowlisting.

## Documentation Index

- [CLI Reference](docs/CLI_REFERENCE.md)
- [Deck Builder](docs/DECK_BUILDER.md)
- [Deck Analysis Suite](docs/DECK_ANALYSIS_SUITE.md)
- [Evolution System](docs/EVOLUTION.md)
- [Event Tracking](docs/EVENT_TRACKING.md)
- [CSV Exports](docs/CSV_EXPORTS.md)
- [Testing](docs/TESTING.md)
- [Release Process](docs/RELEASE_PROCESS.md)

## Troubleshooting

- API auth errors: confirm `CLASH_ROYALE_API_TOKEN` and API key IP allowlist settings.
- Tag errors: pass `--tag` without `#`.
- Rate limiting/timeouts: increase `REQUEST_DELAY` in `.env`.
- Build issues: run `go mod download` and `task build`.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT. See [LICENSE](LICENSE).

## Support

- [Official API Documentation](https://developer.clashroyale.com/#/documentation)
- [Clash Royale API Discord](https://discord.gg/clashroyale)
- [GitHub Issues](https://github.com/klauern/clash-royale-api/issues)
