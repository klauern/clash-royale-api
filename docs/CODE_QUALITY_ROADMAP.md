# Code Quality Roadmap

**Generated:** 2026-01-28
**Baseline:** [CODE_QUALITY_BASELINE.md](CODE_QUALITY_BASELINE.md)
**Status:** ✅ Zero Violations

## Overview

This roadmap prioritizes code quality improvements based on the baseline analysis. With **zero current violations**, the focus is on continuous monitoring, maintaining quality standards, and proactive improvements before technical debt accumulates.

## Current Status

| Category | Status | Priority | Action Required |
|----------|--------|----------|-----------------|
| Security (errcheck) | ✅ None | P0 | Monitoring only |
| Duplication | ✅ None | P1 | Monitoring only |
| Complexity | ✅ None | P2 | Monitoring only |
| Architecture | ✅ None | P2 | Monitoring only |
| Error Handling | ✅ None | P0 | Monitoring only |

## Quick Wins (P0 - Immediate)

### 1. CI Integration with Soft-Fail
**Issue:** [clash-royale-api-4nga](https://github.com/klauern/clash-royale-api/issues/4nga)

Add golangci-lint to CI pipeline with soft-fail mode to catch regressions early without blocking merges.

**Implementation:**
- Add golangci-lint step to GitHub Actions
- Use `--new-from-rev=HEAD^` to report only new issues
- Exit code 0 even with violations (soft-fail)
- Comment PR results for visibility

**Effort:** 1 hour | **Impact:** High | **Risk:** Low

### 2. Pre-Commit Hook
Add a pre-commit hook to run linters before each commit, providing immediate feedback.

**Implementation:**
```bash
# .git/hooks/pre-commit
task lint
```

**Effort:** 15 minutes | **Impact:** Medium | **Risk:** Low

## Monitoring & Maintenance (P1 - Short Term)

### 3. Monthly Baseline Regeneration
Regenerate baseline reports monthly to track code quality trends.

**Process:**
```bash
golangci-lint run --output.json.path reports/baseline.json
golangci-lint run --output.html.path reports/baseline.html
```

**Schedule:** First Monday of each month

### 4. Threshold Review
Review linter thresholds quarterly to ensure they remain appropriate as the codebase grows.

**Current thresholds to monitor:**
- `dupl`: 100 tokens (consider lowering to 80)
- `gocyclo`: 10 complexity (consider lowering to 8)
- `gocognit`: 20 cognitive (consider lowering to 15)
- `funlen`: 60 lines / 50 statements (current is good)

**Review cadence:** Quarterly

## Proactive Improvements (P2 - Medium Term)

### 5. Enhanced Linter Coverage
Enable additional linters that are currently disabled:

| Linter | Purpose | Complexity |
|--------|---------|------------|
| `gochecknoglobals` | Detect global variables | Low |
| `gochecknoinits` | Detect init() usage | Low |
| `nakedret` | Detect naked returns | Low |
| `prealloc` | Suggest slice pre-allocation | Medium |
| `unparam` | Detect unused parameters | Medium |

**Effort:** 2 hours | **Impact:** Medium | **Risk:** Low (fix first, then enforce)

### 6. Test Coverage Thresholds
Add coverage tracking and establish minimum coverage thresholds.

**Implementation:**
- Add `go test -coverprofile=coverage.out` to CI
- Set minimum coverage: 70% for new code, 80% for critical paths
- Track coverage trends over time

**Effort:** 1 hour | **Impact:** High | **Risk:** Low

### 7. Dependency Scanning
Add `gosec` security scanning to detect potential security issues.

**Current status:** gosec is available but not strictly enforced

**Effort:** 30 minutes | **Impact:** High | **Risk:** Low

## Major Refactoring (P3 - Long Term)

### 8. Architecture Review
While no violations exist, conduct a periodic architecture review to identify:
- Package boundary clarity
- Interface bloat prevention
- Circular dependency risks

**Tools:** `depguard`, `interfacebloat`, `ireturn` (already enabled)

**Review cadence:** Bi-annually

### 9. Performance Profiling
Add performance benchmarking and profiling to catch regressions early.

**Implementation:**
- Add benchmark tests for hot paths
- Use `go test -bench` in CI
- Track benchmark trends

**Effort:** 4 hours | **Impact:** Medium | **Risk:** Low

## Response Plan: When Violations Appear

If future baseline reports show violations, follow this prioritization:

### P0 - Security & Error Handling
**Immediate action required (within 1 week)**
- `errcheck` violations - Unchecked errors
- `gosec` security issues
- `staticcheck` bugs

### P1 - Code Duplication
**Fix within 2 weeks**
- `dupl` violations above threshold
- Refactor shared code into functions/packages

### P2 - Complexity
**Fix within 1 month**
- `gocyclo`, `gocognit`, `funlen` violations
- Break down complex functions
- Consider extracting types/functions

### P3 - Architecture
**Fix within 2 months**
- `depguard` violations - Architectural boundary issues
- `interfacebloat` - Too many interface methods
- `ireturn` - Interface return type issues

## Implementation Order

```
┌─────────────────────────────────────────────────────────────┐
│ Week 1: Quick Wins                                          │
│   1. CI Integration (soft-fail)                             │
│   2. Pre-commit hook                                        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ Month 1: Monitoring                                        │
│   3. Monthly baseline regeneration                          │
│   4. Coverage tracking                                      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ Quarter 1: Proactive Improvements                          │
│   5. Enhanced linter coverage                               │
│   6. Dependency scanning                                    │
│   7. Threshold review                                       │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ Ongoing: Major Refactoring                                 │
│   8. Architecture reviews (bi-annual)                       │
│   9. Performance profiling (quarterly)                      │
└─────────────────────────────────────────────────────────────┘
```

## Tracking

Track progress in the beads issue tracker:
- Create issues for each roadmap item
- Link related issues with dependencies
- Mark completed items with roadmap version

## Resources

- **Baseline:** [CODE_QUALITY_BASELINE.md](CODE_QUALITY_BASELINE.md)
- **Testing:** [TESTING.md](TESTING.md)
- **CI Issue:** [clash-royale-api-4nga](https://github.com/klauern/clash-royale-api/issues/4nga)

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-01-28 | Initial roadmap based on zero-violation baseline |
