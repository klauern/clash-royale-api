# AGENTS.md

## What This Is

Go CLI for analyzing Clash Royale player data, building decks, tracking events, and exporting data.

**Tech Stack:** Go 1.26, Official Clash Royale API, Rate-limited client (1 req/sec)
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

### Claude Skills (local to this repo)
- [deck-analysis](.claude/skills/deck-analysis/) - Comprehensive deck building and analysis workflow

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

<!-- bv-agent-instructions-v1 -->

---

## Beads Workflow Integration

This project uses [beads_viewer](https://github.com/Dicklesworthstone/beads_viewer) for issue tracking. Issues are stored in `.beads/` and tracked in git.

### Essential Commands

```bash
# View issues (launches TUI - avoid in automated sessions)
bv

# CLI commands for agents (use these instead)
bd ready              # Show issues ready to work (no blockers)
bd list --status=open # All open issues
bd show <id>          # Full issue details with dependencies
bd create --title="..." --type=task --priority=2
bd update <id> --status=in_progress
bd close <id> --reason="Completed"
bd close <id1> <id2>  # Close multiple issues at once
bd sync               # Commit and push changes
```

### Workflow Pattern

1. **Start**: Run `bd ready` to find actionable work
2. **Claim**: Use `bd update <id> --status=in_progress`
3. **Work**: Implement the task
4. **Complete**: Use `bd close <id>`
5. **Sync**: Always run `bd sync` at session end

### Key Concepts

- **Dependencies**: Issues can block other issues. `bd ready` shows only unblocked work.
- **Priority**: P0=critical, P1=high, P2=medium, P3=low, P4=backlog (use numbers, not words)
- **Types**: task, bug, feature, epic, question, docs
- **Blocking**: `bd dep add <issue> <depends-on>` to add dependencies

### Session Protocol

**Before ending any session, run this checklist:**

```bash
git status              # Check what changed
git add <files>         # Stage code changes
bd sync                 # Commit beads changes
git commit -m "..."     # Commit code
bd sync                 # Commit any new beads changes
git push                # Push to remote
```

### Best Practices

- Check `bd ready` at session start to find available work
- Update status as you work (in_progress → closed)
- Create new issues with `bd create` when you discover tasks
- Use descriptive titles and set appropriate priority/type
- Always `bd sync` before ending session

<!-- end-bv-agent-instructions -->

<!-- BEGIN BEADS INTEGRATION -->
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
bd create "Issue title" --description="Detailed context" -t bug|feature|task -p 0-4 --json
bd create "Issue title" --description="What this issue is about" -p 1 --deps discovered-from:bd-123 --json
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
   - `bd create "Found bug" --description="Details about what was found" -p 1 --deps discovered-from:<parent-id>`
5. **Complete**: `bd close <id> --reason "Done"`

### Auto-Sync

bd automatically syncs with git:

- Exports to `.beads/issues.jsonl` after changes (5s debounce)
- Imports from JSONL when newer (e.g., after `git pull`)
- No manual export/import needed!

### Important Rules

- ✅ Use bd for ALL task tracking
- ✅ Always use `--json` flag for programmatic use
- ✅ Link discovered work with `discovered-from` dependencies
- ✅ Check `bd ready` before asking "what should I work on?"
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT use external issue trackers
- ❌ Do NOT duplicate tracking systems

For more details, see README.md and docs/QUICKSTART.md.

<!-- END BEADS INTEGRATION -->

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
