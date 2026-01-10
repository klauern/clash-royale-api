# AGENTS.md

## What This Is

Go CLI for analyzing Clash Royale player data, building decks, tracking events, and exporting data.

**Tech Stack:** Go 1.24, Official Clash Royale API, Rate-limited client (1 req/sec)
**Architecture:** Clean separation - API client → Data models → Analysis → Export

## Quick Start

```bash
task setup        # Configure .env (requires CLASH_ROYALE_API_TOKEN)
task build        # Build cr-api binary
task test         # Run all tests
task lint         # Run linters
```

**Required:** Copy `.env.example` to `.env` and add token from [developer.clashroyale.com](https://developer.clashroyale.com/)

## Development Workflow

1. Find work: `bd ready`
2. Claim: `bd update <id> --status=in_progress`
3. Implement: Edit code, run `task test` and `task lint`
4. Complete: `bd close <id>` and commit with `.beads/issues.jsonl`

See [.cursor/rules/beads.mdc](.cursor/rules/beads.mdc) for detailed beads workflow.

## Deep Dives

### Documentation
| Topic | Document |
|--------|----------|
| CLI commands | [docs/CLI_REFERENCE.md](docs/CLI_REFERENCE.md) |
| Testing | [docs/TESTING.md](docs/TESTING.md) |
| Release process | [docs/RELEASE_PROCESS.md](docs/RELEASE_PROCESS.md) |
| Deck building | [docs/DECK_BUILDER.md](docs/DECK_BUILDER.md) |
| Deck analysis suite | [docs/DECK_ANALYSIS_SUITE.md](docs/DECK_ANALYSIS_SUITE.md) |
| Evolution mechanics | [docs/EVOLUTION.md](docs/EVOLUTION.md) |
| Event tracking | [docs/EVENT_TRACKING.md](docs/EVENT_TRACKING.md) |
| CSV exports | [docs/CSV_EXPORTS.md](docs/CSV_EXPORTS.md) |

### Cursor Rules (auto-loaded by context)
- [.cursor/rules/go.mdc](.cursor/rules/go.mdc) - Go conventions (auto-loads for `*.go` files)
- [.cursor/rules/testing.mdc](.cursor/rules/testing.mdc) - Testing patterns
- [.cursor/rules/beads.mdc](.cursor/rules/beads.mdc) - Beads workflow (always loaded)

## Writing Good Issues

- **Summary:** What and why (1-2 sentences)
- **Files:** Exact paths with line numbers
- **Steps:** Numbered implementation actions
- **Example:** Before → after transformation

**Dependencies:** `bd dep add X Y` means "X needs Y" (Y blocks X)

## Getting Help

- `task` - Show all available tasks
- `bd <command> --help` - Show bd command options
- `./bin/cr-api --help` - Show CLI options
