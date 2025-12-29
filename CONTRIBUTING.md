# Contributing

Thanks for your interest in contributing!

## Quick Start

1. Fork and clone the repo
2. Run `task setup` (or manually: `cp .env.example .env`, add your API token, `go mod download`)
3. Run `task test` to verify everything works

## Code Quality

Before submitting:

```bash
task format   # Format code (gofmt)
task lint     # Run golangci-lint
task test     # Run tests
```

## Pull Requests

1. Create a feature branch from `main`
2. Make focused, atomic commits using [Conventional Commits](https://www.conventionalcommits.org/)
   - `feat: add deck export`
   - `fix: handle empty card list`
3. Ensure tests pass and lint is clean
4. Open a PR with a clear description

## Questions?

Open an issue or start a discussion.
