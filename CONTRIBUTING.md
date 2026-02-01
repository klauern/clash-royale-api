# Contributing

Thanks for your interest in contributing!

## Quick Start

1. Fork and clone the repo
2. Run `task setup` (or manually: `cp .env.example .env`, add your API token, `go mod download`)
3. Run `task test` to verify everything works

## Issue Tracking with bd (beads)

This project uses [bd (beads)](https://github.com/steveyegge/beads) for issue tracking. Do NOT use markdown TODOs or external trackers.

### Workflow

```bash
# Find ready work
bd ready

# Create new issues
bd create "Issue title" -t bug|feature|task -p 0-4

# Claim and work on an issue
bd update bd-42 --status in_progress

# Complete work
bd close bd-42 --reason "Implemented feature X"
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

### Best Practices

- Write self-contained issues with clear implementation steps
- Link discovered work: `bd create "Found bug" --deps discovered-from:bd-123`
- Always commit `.beads/issues.jsonl` with code changes

See CLAUDE.md for complete bd documentation.

## Code Quality

Before submitting:

```bash
task format        # Format code (gofmt)
task lint          # Run golangci-lint (all issues)
task lint-quality  # Run golangci-lint (new violations only, same as CI)
task test          # Run tests
```

### CI Lint Policy

**CI now hard-fails on new lint violations.** The linter is configured to only report issues introduced in your branch (compared to `origin/main`). This means:

- Existing violations won't block your PR
- New violations you introduce will cause CI to fail
- Use `task lint-quality` locally to preview CI behavior

To fix failing CI lint:
1. Run `task lint-quality` to see new violations
2. Fix the reported issues
3. Commit and push the fixes

## Pull Requests

1. Create a feature branch from `main`
2. Make focused, atomic commits using [Conventional Commits](https://www.conventionalcommits.org/)
   - `feat: add deck export`
   - `fix: handle empty card list`
3. Ensure tests pass and lint is clean
4. Open a PR with a clear description

## Release Process

This project uses GitHub Actions for automated releases with GoReleaser.

### Creating a Release

1. **Ensure quality**: `task test && task lint`
2. **Test locally**: `task snapshot && ls dist/`
3. **Create and push tag**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0: Description"
   git push origin v1.0.0
   ```
4. **GitHub Actions** automatically builds all platforms and publishes to GitHub Releases

### Versioning

- Follow [Semantic Versioning](https://semver.org/): Major.Minor.Patch
- `feat:` commits → Minor version bump
- `fix:` commits → Patch version bump
- Breaking changes → Major version bump

## Documentation Standards

- **CLAUDE.md**: Project overview, architecture, workflow guidance
- **README.md**: User-facing documentation, quick start, features
- **Code comments**: Explain "why", not "what"
- **CSV exports**: Document field definitions in CSV_EXPORTS.md
- **Planning docs**: Store in `history/` directory (ephemeral)

## Questions?

Open an issue or start a discussion.
