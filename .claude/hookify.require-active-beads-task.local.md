---
name: require-active-beads-task
enabled: true
event: file
conditions:
  - field: file_path
    operator: regex_match
    pattern: \.(go|ts|tsx|js|jsx|py|rb|java|rs|c|cpp|h)$
  - field: new_text
    operator: regex_match
    pattern: (func |function |class |def |impl |struct )
action: warn
---

🚨 **WARNING: Significant code changes without active beads task**

You're writing substantial code (functions, classes, structs) but there's no active beads task.

**Required workflow:**
1. **Find or create a task:**
   ```bash
   bd ready              # Find available work
   bd create --title "Your task" --type task --priority 2
   ```

2. **Claim the task:**
   ```bash
   bd update <task-id> --status in_progress
   ```

3. **Then proceed with code changes**

**Why this matters:**
- Tracks all work in beads for visibility
- Links code changes to requirements/issues
- Helps with context recovery across sessions
- Prevents orphaned work

**For quick fixes:**
If this is trivial (typo, formatting), create a quick task:
```bash
bd create --title "Fix typo in X" --type task --priority 3
bd update <id> --status in_progress
```

**Current status:**
Run `bd list --status=in_progress` to see if you have an active task.
