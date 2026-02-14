# Code Quality Baseline

**Generated:** 2026-01-28
**Tool:** golangci-lint v2.6.2
**Status:** ✅ Zero Issues

## Summary

The codebase is currently in excellent health with **zero linting issues** across all enabled linters. This baseline was established after configuring the comprehensive linter suite with warning-level thresholds.

## Enabled Linters

### Architecture & Code Organization
- **depguard** - Prevents `cmd/` from importing `internal/` packages
- **interfacebloat** - Limits interfaces to 10 methods max
- **ireturn** - Checks interface return types

### Code Quality
- **govet** - Go vet static analysis
- **staticcheck** - Advanced static analysis
- **errcheck** - Error checking (with type-assertions and blank checks)
- **unused** - Detects unused code
- **ineffassign** - Detects ineffectual assignments

### Style & Formatting
- **revive** - Fast, configurable linter (exported rules disabled)
- **gocritic** - Provides diagnostics that check for bugs, performance, and style issues
- **misspell** - Spell checking (US locale)
- **gofumpt** - Stricter gofmt
- **goimports** - Import management

### Complexity & Duplication
- **dupl** - Code duplication detection (threshold: 100)
- **gocyclo** - Cyclomatic complexity (min: 10)
- **gocognit** - Cognitive complexity (min: 20)
- **funlen** - Function length limits (60 lines / 50 statements)
- **goconst** - Repeated strings that could be constants

## Configuration

### Settings

| Linter | Setting | Value |
|--------|---------|-------|
| dupl | threshold | 100 |
| goconst | min-len | 2 |
| goconst | min-occurrences | 3 |
| gocyclo | min-complexity | 10 |
| gocognit | min-complexity | 20 |
| funlen | lines | 60 |
| funlen | statements | 50 |
| interfacebloat | max | 10 |

### DepGuard Rules

```yaml
cmd-internal:
  files:
    - cmd/**
  deny:
    - pkg: github.com/klauern/clash-royale-api/go/internal
      desc: cmd should not import internal packages
```

### Test Exclusions

All linters are disabled for test files (`_test.go`) to allow flexibility in test code.

### Issue Reporting

- `new-from-rev: origin/main` - Only report new issues
- `max-issues-per-linter: 0` - Unlimited issues per linter
- `max-same-issues: 0` - Unlimited duplicate issues

## Violation Categories

### Security
- **Status:** ✅ None detected
- **Linters:** govet, staticcheck, gocritic

### Duplication
- **Status:** ✅ None detected
- **Linters:** dupl (threshold: 100)

### Complexity
- **Status:** ✅ None detected
- **Linters:** gocyclo (10+), gocognit (20+), funlen (60/50)

### Architecture
- **Status:** ✅ None detected
- **Linters:** depguard, interfacebloat (max 10 methods), ireturn

### Error Handling
- **Status:** ✅ None detected
- **Linters:** errcheck (with type-assertions and blank)

## Reports

Full reports are available in the `reports/` directory:

- `baseline.json` - Machine-readable JSON report
- `baseline.html` - Human-readable HTML report

## Next Steps

This baseline provides a foundation for:

1. **CI Integration** - Add linters to CI with soft-fail mode ([clash-royale-api-4nga](https://github.com/klauern/clash-royale-api/issues/4nga))
2. **Continuous Monitoring** - Track new issues as they're introduced
3. **Refactoring Roadmap** - Prioritize fixes if violations increase
4. **Quality Gates** - Consider stricter thresholds in the future

## Regenerating This Baseline

To regenerate baseline reports:

```bash
# JSON report
golangci-lint run --output.json.path reports/baseline.json

# HTML report
golangci-lint run --output.html.path reports/baseline.html

# Update this document
# Review the reports and update the Violation Categories section
```

## Notes

- The project uses a "new-from-rev" strategy, meaning only new issues are reported
- Test files are excluded from linting to maintain test flexibility
- Thresholds are set to generous values to avoid noise while catching real issues
- All linters run with a 5-minute timeout
