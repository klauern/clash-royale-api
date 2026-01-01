# Testing Guide

Complete testing documentation for the Clash Royale API project.

## Test Commands

### Task Runner (Recommended)

```bash
task test              # Run all tests
task test-go           # Run all Go tests
task test-go-coverage  # Run with coverage report
```

### Direct Go Commands

```bash
cd go && go test ./...              # Run all tests
cd go && go test ./pkg/deck/... -v  # Test specific package with verbose output
cd go && go test -tags=integration ./...  # Full integration tests (requires API token)
```

## Test Types

### Unit Tests

Unit tests run without external dependencies and are fast. They cover:
- Package-level functionality
- Business logic
- Data structures
- Error handling

Unit tests run automatically in CI.

### Integration Tests

Integration tests connect to the live Clash Royale API and require:
- Valid API token in `.env`
- IP allowlisted at [developer.clashroyale.com](https://developer.clashroyale.com/)
- Active internet connection

Integration tests are excluded from CI due to IP restrictions (see below).

## CI/CD Limitations

The Clash Royale API requires static IP whitelisting (maximum 5 IPs per API key). GitHub Actions standard runners use dynamic IPs and cannot access the live API.

**Current strategy:**
- CI runs unit tests only
- Integration tests require manual execution via `task test-integration`

**Automation options:**
- Self-hosted runners with static IP
- GitHub Enterprise larger runners with configurable static IPs
- Proxy services with static IP addresses

## Coverage

To view coverage report:
```bash
task test-go-coverage
# Or
cd go && go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Writing Tests

### Test File Location

Place test files in the same package as the code they test:
```
pkg/deck/builder.go      # Implementation
pkg/deck/builder_test.go # Tests
```

### Test Naming

Use descriptive names that indicate what is being tested:
```go
func TestDeckBuilder_BuildsBalancedDeck(t *testing.T) {
    // Test implementation
}
```

### Test Organization

Use table-driven tests for multiple scenarios:
```go
func TestCardScore(t *testing.T) {
    tests := []struct {
        name     string
        card     Card
        expected float64
    }{
        {"common card level 1", Card{Level: 1, Rarity: "common"}, 1.0},
        {"rare card level 1", Card{Level: 1, Rarity: "rare"}, 2.0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateScore(tt.card)
            if result != tt.expected {
                t.Errorf("expected %f, got %f", tt.expected, result)
            }
        })
    }
}
```

## Test Data

Use `testdata` directory for test fixtures:
```
pkg/deck/testdata/
├── player_1.json
├── player_2.json
└── expected_decks.json
```

## Mocking

For API-dependent code, use interfaces and mocks:
```go
// In production code
type APIClient interface {
    GetPlayer(tag string) (*Player, error)
}

// In tests
type MockAPIClient struct {
    players map[string]*Player
}

func (m *MockAPIClient) GetPlayer(tag string) (*Player, error) {
    return m.players[tag], nil
}
```

## Linting

Run linter before committing:
```bash
task lint
# Or
cd go && golangci-lint run
```

See [AGENTS.md](../AGENTS.md) for development workflow.
