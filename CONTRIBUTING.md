# Contributing to Clash Royale API

Thank you for your interest in contributing! This document provides guidelines for contributing to this project.

## Development Setup

### Prerequisites

- Go 1.22 or later
- [Task](https://taskfile.dev/) (optional but recommended)
- A Clash Royale API token from [developer.clashroyale.com](https://developer.clashroyale.com)

### Getting Started

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/clash-royale-api.git
   cd clash-royale-api
   ```

2. **Set up your environment**
   ```bash
   # Using Task (recommended)
   task setup

   # Or manually
   cp .env.example .env
   # Edit .env and add your API token
   go mod download
   go build -o bin/cr-api ./cmd/cr-api
   ```

3. **Verify your setup**
   ```bash
   task test
   # Or: go test -short ./...
   ```

## Code Style

### Formatting

- All Go code must be formatted with `gofmt`
- Run `task format` or `go fmt ./...` before committing

### Linting

- Code must pass `golangci-lint`
- Run `task lint` to check for issues
- The CI pipeline will fail if linting errors are present

### Best Practices

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Write clear, self-documenting code with meaningful names
- Keep functions focused and reasonably sized
- Add comments for non-obvious logic

## Testing

### Running Tests

```bash
# Run all unit tests
task test

# Run tests with coverage
task test-go-coverage

# Run specific package tests
go test -v ./pkg/deck/...
```

### Writing Tests

- All new features should include tests
- Bug fixes should include a regression test
- Aim for meaningful coverage of edge cases
- Integration tests (requiring API token) should use the `integration` build tag

### Test Data

- Use the `test_data/` directory for test fixtures
- Never commit real player data or API tokens in tests

## Pull Request Process

### Before Submitting

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Keep commits focused and atomic
   - Write clear commit messages using [Conventional Commits](https://www.conventionalcommits.org/)
   - Examples: `feat: add deck export`, `fix: handle empty card list`, `docs: update README`

3. **Ensure quality**
   ```bash
   task test    # All tests pass
   task lint    # No lint errors
   task format  # Code is formatted
   ```

### Submitting

1. Push your branch and create a Pull Request
2. Fill out the PR template completely
3. Link any related issues
4. Wait for CI checks to pass
5. Respond to review feedback promptly

### Review Process

- PRs require at least one approving review
- Address all review comments or explain why you disagree
- Keep the PR focused; large changes should be split into smaller PRs
- Squash commits if requested for cleaner history

## Issue Guidelines

### Reporting Bugs

- Search existing issues first to avoid duplicates
- Use the bug report template
- Include:
  - Steps to reproduce
  - Expected vs actual behavior
  - Go version and OS
  - Relevant error messages or logs

### Requesting Features

- Check if the feature has already been requested
- Use the feature request template
- Explain the use case and benefits
- Consider if you'd be willing to implement it

## Project Structure

```
clash-royale-api/
├── cmd/
│   ├── cr-api/          # Main CLI application
│   └── deckbuilder/     # Standalone deck builder
├── pkg/
│   ├── clashroyale/     # API client library
│   ├── deck/            # Deck building logic
│   ├── events/          # Event tracking
│   └── analysis/        # Card analysis
├── internal/
│   ├── exporter/        # Export functionality
│   ├── storage/         # Data persistence
│   └── utils/           # Common utilities
└── data/                # Local data (gitignored)
```

## Questions?

- Open a [Discussion](https://github.com/klauern/clash-royale-api/discussions) for general questions
- Check existing issues for similar problems
- Read the README and documentation first

Thank you for contributing!
