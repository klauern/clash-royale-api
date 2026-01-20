---
name: require-tests-or-beads-task
enabled: true
event: file
action: warn
---

## ⚠️ Testing Required for New Code

You are adding new code without tests or proper documentation.

### Requirements

When adding new code (packages, functions, or significant features), you must do **ONE** of the following:

1. **Write tests now** - Add `_test.go` files alongside your implementation
2. **Document deferred tests** - Use this pattern in a test file:

```go
t.Run("note: <what> deferred", func(t *testing.T) {
    // Clear explanation of what's deferred and why
    t.Skip("<reason> (task <beads-issue-id>)")
})
```

3. **Create a beads task** - Run `bd create --title "Add tests for <feature>" --type task` and link it as a dependency

### What This Checks

This warning triggers when you:
- Add new `.go` files without corresponding `_test.go` files
- Add TODO/FIXME comments without a beads task reference (e.g., "task clash-xxx")
- Skip tests without proper documentation or task references

### Examples

**Good** - Tests included:
```
feat(genetic): add eaopt dependency and genetic package scaffold
- pkg/deck/genetic/config.go
- pkg/deck/genetic/config_test.go (comprehensive validation tests)
- pkg/deck/genetic/genome.go
- pkg/deck/genetic/genome_test.go (unit tests for all methods)
```

**Good** - Tests properly deferred:
```go
// In genome_test.go
t.Run("note: full implementation deferred", func(t *testing.T) {
    t.Skip("Full fitness calculation depends on evaluation system integration (task clash-royale-api-hj9j.2)")
})
```

**Good** - TODO with task reference:
```go
// TODO: Implement fitness calculation using deck scoring system (task clash-royale-api-hj9j.2)
```

**Bad** - No tests, no documentation:
```
feat(pkg): add new feature
- pkg/newfeature/code.go (no tests added!)
```

**Bad** - Orphaned TODO:
```go
// TODO: Implement this later (no task ID referenced!)
```

### Why This Matters

From the conversation history:
- The genetic package was added without tests
- User had to explicitly ask "do we have tests to verify things work?"
- Tests were then added in a follow-up commit 7 minutes later
- This reactive pattern wastes time and should be avoided

### Beads Workflow Integration

When deferring tests:
1. Create a beads task: `bd create --title "Add tests for <feature>" --type task --priority 2`
2. Note the task ID (e.g., `clash-royale-api-xxxx`)
3. Reference it in your `t.Skip()` message or `// TODO` comment
4. Link dependencies: `bd dep add <current-task> <test-task>`
5. Don't close the feature task until tests are done OR properly deferred
