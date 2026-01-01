# AGENTS.md

This file provides guidance to AI agents (Claude Code, etc.) when working with this repository.

## Project Overview

Go-only Clash Royale API client for analyzing player data, building decks, tracking events, and exporting data.

**Key features:**
- Official API integration with rate limiting (1 req/sec default)
- Clean architecture: API client → Data models → Analysis → Export
- Type-safe with comprehensive error handling
- Beads task management workflow

## Repository Structure

```
clash-royale-api/
├── cmd/cr-api/          # Main CLI application
├── pkg/
│   ├── clashroyale/     # API client
│   ├── deck/            # Deck building algorithms
│   ├── analysis/        # Collection analysis
│   └── events/          # Event deck tracking
├── internal/
│   ├── exporter/csv/    # CSV export
│   └── storage/         # Data persistence
├── data/                # Local-only (gitignored)
└── docs/                # Feature documentation
```

## Quick Start

```bash
task setup        # Configure environment (.env)
task build        # Build binaries
task test         # Run tests
```

**Required:** Copy `.env.example` to `.env` and add `CLASH_ROYALE_API_TOKEN` from [developer.clashroyale.com](https://developer.clashroyale.com/)

## Development Workflow

1. **Find work**: `bd ready` shows unblocked issues
2. **Claim task**: `bd update <id> --status in_progress`
3. **Implement**: Edit code, run `task test` and `task lint`
4. **Complete**: `bd close <id>` and commit changes with `.beads/issues.jsonl`

## Task-Specific Documentation

This file contains only universally applicable information. For task-specific details, refer to:

| Topic | Document |
|--------|----------|
| Complete CLI commands | [docs/CLI_REFERENCE.md](docs/CLI_REFERENCE.md) |
| Testing strategies | [docs/TESTING.md](docs/TESTING.md) |
| Release process | [docs/RELEASE_PROCESS.md](docs/RELEASE_PROCESS.md) |
| Deck building algorithms | [docs/DECK_BUILDER.md](docs/DECK_BUILDER.md) |
| Evolution mechanics | [docs/EVOLUTION.md](docs/EVOLUTION.md) |
| Event tracking | [docs/EVENT_TRACKING.md](docs/EVENT_TRACKING.md) |
| CSV exports | [docs/CSV_EXPORTS.md](docs/CSV_EXPORTS.md) |

**Tip:** When working on a specific task, read the relevant documentation first for detailed context.

## Beads Issue Tracking

This project uses [bd (beads)](https://github.com/steveyegge/beads) for ALL task tracking.

**Why bd?**
- Dependency-aware: Tracks blockers and relationships
- Git-friendly: Auto-syncs to `.beads/issues.jsonl`
- Agent-optimized: JSON output, ready work detection

**Essential commands:**
```bash
bd ready                           # Find unblocked work
bd create "Title" -t task -p 1     # Create issue
bd update <id> --status in_progress # Claim work
bd close <id> --reason "Done"      # Complete work
```

**Issue types:** `bug`, `feature`, `task`, `epic`, `chore`
**Priorities:** `0` (critical) → `4` (backlog)

**Writing self-contained issues:**
- Summary: What and why (1-2 sentences)
- Files: Exact paths with line numbers
- Steps: Numbered implementation actions
- Example: Before → after transformation

**Dependencies:** `bd dep add X Y` means "X needs Y" (Y blocks X)
**Verify:** `bd blocked` shows blocked tasks

## Code Conventions

- Use `task` runner instead of direct `make` or `go` commands
- Use `fd` instead of `find`, `rg` instead of `grep`
- Use `gh` for GitHub operations
- All data persists in `data/` directory (gitignored)
- Store AI planning docs in `history/` directory

## Architecture Patterns

- **Clean package structure:** `pkg/` (libraries), `internal/` (internals)
- **Type-safe enums:** Card roles, constants
- **Interface-based design:** Testability and extensibility
- **Builder pattern:** Deck construction
- **Error types:** Specific error codes with context

## Common Pitfalls

- ❌ Don't commit `.env` or `data/` (gitignored)
- ❌ Don't use markdown TODOs (use bd instead)
- ❌ Don't skip `task lint` before committing
- ❌ Don't forget to commit `.beads/issues.jsonl` with code changes
- ✅ Always run tests before claiming completion
- ✅ Check for broken links after moving documentation

## Getting Help

- `task` - Show all available tasks
- `bd <command> --help` - Show bd command options
- `./bin/cr-api --help` - Show CLI options
- `bd ready` - Find work to do
