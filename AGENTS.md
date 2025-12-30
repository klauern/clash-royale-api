# AGENTS.md

**Note**: This project uses [bd (beads)](https://github.com/steveyegge/beads)
for issue tracking. Use `bd` commands instead of markdown TODOs.

This file provides guidance to AI agents (Claude Code, etc.) when working with this repository.

## Repository Overview

This is a Go-only Clash Royale API client and analysis tool. The Go implementation provides comprehensive functionality for analyzing player data, building decks, tracking event performance, and exporting data.

Key features:
- Uses the official Clash Royale API (developer.clashroyale.com)
- Implements rate limiting (1 req/sec default)
- Clean architecture: API client → Data models → Analysis → Export
- Stores data in the `data/` directory
- Follows beads task management workflow
- Type-safe with comprehensive error handling

## Project Architecture

### Go Implementation

**Structure:**
- `cmd/cr-api/main.go` - Main CLI application using `urfave/cli/v3`
- `pkg/clashroyale/` - API client library (independent implementation)
- `pkg/events/` - Event deck tracking and analysis
- `pkg/deck/` - Intelligent deck building with role-based selection
- `pkg/analysis/` - Card collection analysis and upgrade priorities
- `internal/exporter/csv/` - CSV export functionality
- `internal/storage/` - Data persistence and organization
- `internal/utils/` - Common utilities and helpers

**Current Status:**
- Production-ready with comprehensive testing
- High performance with type safety
- Complete feature set with comprehensive functionality

**Go Architecture Patterns:**
- Clean package structure (`pkg/` for libraries, `internal/` for internals)
- Type-safe enums for card roles and constants
- Interface-based design for testability and extensibility
- Comprehensive error types with specific error codes
- Builder pattern for deck construction

## Development Commands

### Task Runner (Recommended)

Install: `./scripts/install-task.sh` or from https://taskfile.dev

```bash
task              # Show all tasks
task setup        # Set up env (.env, deps, build)
task build        # Build binaries
task run -- '#TAG'        # Analyze player
task run-with-save -- '#TAG'    # Analyze + save JSON
task export-csv -- '#TAG'       # Export to CSV
task scan-events -- '#TAG'      # Scan event decks
task test         # Run tests with coverage
task lint         # Run golangci-lint
task snapshot     # Test release locally
task release      # Create release (requires GITHUB_TOKEN)
```

### Direct CLI Usage

Build: `cd go && go build -o bin/cr-api ./cmd/cr-api`

```bash
./bin/cr-api player --tag <TAG> [--chests] [--save] [--export-csv]
./bin/cr-api cards [--export-csv]
./bin/cr-api analyze --tag <TAG> [--save] [--export-csv]
./bin/cr-api deck build --tag <TAG> [--combat-stats-weight 0.25] [--disable-combat-stats]
./bin/cr-api events scan --tag <TAG>
./bin/cr-api playstyle --tag <TAG> [--recommend-decks] [--save]

cd go && go test ./...              # Run all tests
cd go && go test ./pkg/deck/... -v  # Test specific package
```

## Release Process

1. **Ensure tests pass**: `task test && task lint`
2. **Test locally**: `task snapshot && ls dist/`
3. **Create and push tag**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0: Description"
   git push origin v1.0.0
   ```
4. **GitHub Actions** builds all platforms and publishes to GitHub Releases

**Versioning**: Semantic versioning (Major.Minor.Patch)
- `feat:` → Features, `fix:` → Bug Fixes, `perf:`/`refactor:` → Improvements
- `test:` → Tests, `chore:`/`docs:`/`ci:` → Excluded from changelog

**Manual release** (if needed): `export GITHUB_TOKEN=... && task release`

## Configuration

**⚠️ Security**: Copy `.env.example` to `.env` and add your actual values. Never commit `.env` to version control.

**Required Environment Variables (.env):**
```env
CLASH_ROYALE_API_TOKEN=your_token_here
```

**Optional Configuration:**
```env
DEFAULT_PLAYER_TAG=#TAG          # Allows running tasks without arguments
DATA_DIR=./data                  # Data storage location
REQUEST_DELAY=1                  # Seconds between API requests
MAX_RETRIES=3                    # API retry attempts
CSV_DIR=./data/csv               # CSV export directory
COMBAT_STATS_WEIGHT=0.25         # Combat stats weight for deck building (0.0-1.0)
```

**Configuration Priority:**
1. CLI arguments (highest)
2. Environment variables
3. Default values (lowest)

## Evolution Tracking

The deck builder tracks unlocked evolutions (API doesn't provide this). Configure in `.env`:

```env
UNLOCKED_EVOLUTIONS="Archers,Knight,Musketeer"
```

**CLI override**: `./bin/cr-api deck build --tag <TAG> --unlocked-evolutions "Archers,Bomber" --evolution-slots 2`

**How it works:**
1. **Bonus scaling**: Level-scaled bonus via formula `0.25 * (level/maxLevel)^1.5 * (1 + 0.2*(maxEvoLevel-1))`
2. **Role overrides**: Evolved cards may change role (e.g., Valkyrie: Cycle → Support)
3. **Slot priority**: Top 2 evolved cards (by role priority + score) get evolution slots

**Recommendations**: `./bin/cr-api evolutions recommend --tag <TAG> [--top 5] [--verbose]`

See [EVOLUTION.md](EVOLUTION.md) for complete documentation including:
- Evolution level mechanics (evolutionLevel vs maxEvolutionLevel)
- Shard management and inventory tracking
- Evolution recommendation scoring factors
- Role override details and upgrade priorities

## Combat Stats Integration

The deck builder blends traditional scoring (level, rarity, cost) with combat stats (DPS/elixir, HP/elixir, role effectiveness):

`finalScore = (traditional × (1-weight)) + (combat × weight)`

**Configuration**: Set `COMBAT_STATS_WEIGHT` in `.env` (default: 0.25, range: 0.0-1.0)

**CLI options**:
```bash
./bin/cr-api deck build --tag <TAG>                              # 25% weight (default)
./bin/cr-api deck build --tag <TAG> --combat-stats-weight 0.6   # 60% weight
./bin/cr-api deck build --tag <TAG> --disable-combat-stats       # 0% weight (traditional only)
./bin/cr-api deck build --from-analysis <file>                   # Offline mode
```

**Weight guidance**:
- **0.5-0.8**: Prioritize statistically strong cards (theory-crafting)
- **0.25** (default): Balanced, recommended for most
- **0.0-0.2**: Focus on highest-level cards (ladder pushing)
```

## Testing

```bash
# Go
task test                        # Run all tests
task test-go                     # Run all Go tests
task test-go-coverage           # Run with coverage report
cd go && go test ./pkg/deck/... -v  # Test specific package with verbose output

# Integration
cd go && go test -tags=integration ./...  # Full integration tests (requires API token)
```

## Data Storage Structure

**Note**: The `data/` directory is gitignored and contains local-only artifacts (cached API responses, analysis results, CSV exports). It is excluded from version control.

```
data/
├── static/          # Static game data (cards, etc.)
├── players/         # Player profiles (JSON)
├── analysis/        # Collection analysis results
├── csv/
│   ├── players/     # Player CSV exports
│   ├── reference/   # Card database
│   └── analysis/    # Analytical reports
└── event_decks/     # Event deck tracking data
```

Files are timestamped: `YYYYMMDD_HHMMSS_type_PLAYERTAG.json`

## Scripts and Automation

Located in `scripts/`:
- `install-task.sh` - Automated Task installer
- `run_all_tasks.sh` - Execute all Taskfile tasks with detailed logging

## Dependencies

**Go:**
- `urfave/cli/v3` - CLI framework
- `go.uber.org/ratelimit` - Rate limiting
- Go 1.22+ required

## Development Workflow

1. **Setup**: Run `task setup` to configure environment
2. **Development**: Use `bd ready` to find tasks, claim with `bd update`
3. **Testing**: Run `task test` or `task test-go` for comprehensive testing
4. **Code Quality**: Use `task lint` and `task format` for Go code
5. **Data Management**: All data persists in `data/` directory
6. **Task Completion**: Close tasks with `bd close`, then `bd sync --from-main` and commit

## API Rate Limiting

The Go implementation respects API limits:
- Default: 1 request/second
- Automatic retry with exponential backoff
- Configurable via `REQUEST_DELAY` environment variable

## Issue Tracking with bd (beads)

**IMPORTANT**: This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

### Why bd?

- Dependency-aware: Track blockers and relationships between issues
- Git-friendly: Auto-syncs to JSONL for version control
- Agent-optimized: JSON output, ready work detection, discovered-from links
- Prevents duplicate tracking systems and confusion

### Quick Start

**Check for ready work:**
```bash
bd ready --json
```

**Create new issues:**
```bash
bd create "Issue title" -t bug|feature|task -p 0-4 --json
bd create "Issue title" -p 1 --deps discovered-from:bd-123 --json
bd create "Subtask" --parent <epic-id> --json  # Hierarchical subtask (gets ID like epic-id.1)
```

**Claim and update:**
```bash
bd update bd-42 --status in_progress --json
bd update bd-42 --priority 1 --json
```

**Complete work:**
```bash
bd close bd-42 --reason "Completed" --json
```

### Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks
- `chore` - Maintenance (dependencies, tooling)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

### Workflow for AI Agents

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task**: `bd update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" -p 1 --deps discovered-from:<parent-id>`
5. **Complete**: `bd close <id> --reason "Done"`
6. **Commit together**: Always commit the `.beads/issues.jsonl` file together with the code changes so issue state stays in sync with code state

### Writing Self-Contained Issues

Issues must be fully self-contained - readable without any external context (plans, chat history, etc.). A future session should understand the issue completely from its description alone.

**Required elements:**
- **Summary**: What and why in 1-2 sentences
- **Files to modify**: Exact paths (with line numbers if relevant)
- **Implementation steps**: Numbered, specific actions
- **Example**: Show before → after transformation when applicable

**Optional but helpful:**
- Edge cases or gotchas to watch for
- Test references (point to test files or test_data examples)
- Dependencies on other issues

**Bad example:**
```
Implement the refactoring from the plan
```

**Good example:**
```
Add timeout parameter to fetchUser() in src/api/users.ts

1. Add optional timeout param (default 5000ms)
2. Pass to underlying fetch() call
3. Update tests in src/api/users.test.ts

Example: fetchUser(id) → fetchUser(id, { timeout: 3000 })
Depends on: bd-abc123 (fetch wrapper refactor)
```

### Dependencies: Think "Needs", Not "Before"

`bd dep add X Y` = "X needs Y" = Y blocks X

**TRAP**: Temporal words ("Phase 1", "before", "first") invert your thinking!
```
WRONG: "Phase 1 before Phase 2" → bd dep add phase1 phase2
RIGHT: "Phase 2 needs Phase 1" → bd dep add phase2 phase1
```
**Verify**: `bd blocked` - tasks blocked by prerequisites, not dependents.

### Auto-Sync

bd automatically syncs with git:
- Exports to `.beads/issues.jsonl` after changes (5s debounce)
- Imports from JSONL when newer (e.g., after `git pull`)
- No manual export/import needed!

### GitHub Copilot Integration

If using GitHub Copilot, also create `.github/copilot-instructions.md` for automatic instruction loading.
Run `bd onboard` to get the content, or see step 2 of the onboard instructions.

### MCP Server (Recommended)

If using Claude or MCP-compatible clients, install the beads MCP server:

```bash
pip install beads-mcp
```

Add to MCP config (e.g., `~/.config/claude/config.json`):
```json
{
  "beads": {
    "command": "beads-mcp",
    "args": []
  }
}
```

Then use `mcp__beads__*` functions instead of CLI commands.

### Managing AI-Generated Planning Documents

AI assistants often create planning and design documents during development:
- PLAN.md, IMPLEMENTATION.md, ARCHITECTURE.md
- DESIGN.md, CODEBASE_SUMMARY.md, INTEGRATION_PLAN.md
- TESTING_GUIDE.md, TECHNICAL_DESIGN.md, and similar files

**Best Practice: Use a dedicated directory for these ephemeral files**

**Recommended approach:**
- Create a `history/` directory in the project root
- Store ALL AI-generated planning/design docs in `history/`
- Keep the repository root clean and focused on permanent project files
- Only access `history/` when explicitly asked to review past planning

**Example .gitignore entry (optional):**
```
# AI planning documents (ephemeral)
history/
```

**Benefits:**
- ✅ Clean repository root
- ✅ Clear separation between ephemeral and permanent documentation
- ✅ Easy to exclude from version control if desired
- ✅ Preserves planning history for archeological research
- ✅ Reduces noise when browsing the project

### CLI Help

Run `bd <command> --help` to see all available flags for any command.
For example: `bd create --help` shows `--parent`, `--deps`, `--assignee`, etc.

### Important Rules

- ✅ Use bd for ALL task tracking
- ✅ Always use `--json` flag for programmatic use
- ✅ Link discovered work with discovered-from dependencies
- ✅ Check `bd ready` before asking "what should I work on?"
- ✅ Store AI planning docs in `history/` directory
- ✅ Run `bd <cmd> --help` to discover available flags
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT use external issue trackers
- ❌ Do NOT duplicate tracking systems
- ❌ Do NOT clutter repo root with planning documents

For more details, see README.md and QUICKSTART.md.
