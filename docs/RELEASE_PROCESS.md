# Release Process

Complete guide for creating releases of the Clash Royale API tool.

## Prerequisites

1. All tests pass: `task test && task lint`
2. Changes are committed and pushed to main
3. You have push access to the repository
4. GitHub Actions is enabled

## Release Workflow

### 1. Prepare for Release

Ensure tests pass and code is clean:
```bash
task test && task lint
```

### 2. Test Release Locally

Test the release process without actually publishing:
```bash
task snapshot
ls dist/
```

This builds binaries for all platforms without creating a git tag or triggering GitHub Actions.

### 3. Create and Push Version Tag

Create an annotated tag with semantic version:
```bash
git tag -a v1.0.0 -m "Release v1.0.0: Description of changes"
git push origin v1.0.0
```

### 4. GitHub Actions Build

Pushing the tag triggers GitHub Actions automatically:
- Builds binaries for Linux, macOS, and Windows
- Runs all tests
- Creates GitHub Release with binaries attached
- Generates changelog from commit messages

The release will be available at: https://github.com/klauern/clash-royale-api/releases

## Semantic Versioning

Follow Semantic Versioning (Major.Minor.Patch):

**Major** (X.0.0): Breaking changes
- Remove or significantly change features
- API changes that require user action

**Minor** (0.X.0): New features
- Add new functionality
- Add new CLI commands
- Add new data exports

**Patch** (0.0.X): Bug fixes
- Fix bugs without changing features
- Update dependencies
- Small improvements

## Commit Message Conventions

Use conventional commits for automatic changelog generation:

| Type | Description | Included in Changelog |
|------|-------------|----------------------|
| `feat:` | New features | ✅ Yes |
| `fix:` | Bug fixes | ✅ Yes |
| `perf:` | Performance improvements | ✅ Yes |
| `refactor:` | Code refactoring | ✅ Yes |
| `test:` | Test changes | ❌ No |
| `chore:` | Maintenance tasks | ❌ No |
| `docs:` | Documentation changes | ❌ No |
| `ci:` | CI/CD changes | ❌ No |

**Examples:**
```
feat: add what-if analysis for card upgrades
fix: resolve panic when player has no cards
perf: optimize deck building algorithm
refactor: simplify error handling in API client
chore: update dependencies
docs: add CLI reference documentation
test: add integration tests for event scanning
```

## Manual Release (If Needed)

If GitHub Actions fails or you need to create a release manually:

```bash
# Build for all platforms
export GITHUB_TOKEN=your_token
task release
```

This requires `GITHUB_TOKEN` with repo scope to create releases.

## Post-Release Checklist

- [ ] Verify release on GitHub Releases page
- [ ] Test downloaded binaries on each platform
- [ ] Update documentation if needed
- [ ] Announce release (if applicable)

## Rollback Procedure

If a release has critical issues:

1. **Delete the tag and release:**
   ```bash
   git tag -d v1.0.0
   git push origin :refs/tags/v1.0.0
   ```

2. **Create a new patch release:**
   ```bash
   git tag -a v1.0.1 -m "Release v1.0.1: Hotfix for critical bug"
   git push origin v1.0.1
   ```

3. **Yank the release on GitHub** (delete the release, keeping the tag in git history if needed for references)

## Configuration

Release settings in `.goreleaser.yml`:
- Build targets (linux, darwin, windows)
- Binary naming
- Checksum generation
- Homebrew formula creation
- GitHub Release creation

See [AGENTS.md](../AGENTS.md) for development workflow.
