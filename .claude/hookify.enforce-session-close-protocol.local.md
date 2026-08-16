---
name: enforce-session-close-protocol
enabled: true
event: stop
pattern: .*
action: block
---

# 🚨 SESSION CLOSE PROTOCOL 🚨

**CRITICAL**: Before saying "done" or "complete", you MUST run this checklist:

```
[ ] 1. git status              (check what changed)
[ ] 2. git add <files>         (stage code changes)
[ ] 3. bd sync                 (commit beads changes)
[ ] 4. git commit -m "..."     (commit code)
[ ] 5. bd sync                 (commit any new beads changes)
[ ] 6. git push                (push to remote)
```

**NEVER skip this.** Work is not done until pushed.

## Why Each Step Matters

1. **git status** - Verify you know what changed
2. **git add** - Stage the files you modified
3. **bd sync** - Sync beads database with remote (catches any task updates)
4. **git commit** - Create atomic commit with your changes
5. **bd sync** - Sync again (commit step may have updated beads state)
6. **git push** - Push to remote so work isn't lost

## Verification

Before stopping, confirm:
- ✅ `git status` shows "working tree clean"
- ✅ `git log -1` shows your commit
- ✅ `git push` succeeded (no errors)
- ✅ All beads tasks properly closed/synced

## If You Forgot Something

**A linked PR has merged and the Bead still needs completion:**
```bash
bash scripts/close-bead-after-merge.sh <bead-id> <pr-number>
bd sync
git add .beads/
git commit -m "chore(beads): complete merged task <id>"
git push
```

Do not close a PR-backed task before GitHub reports a non-null `mergedAt`.
Passing tests, committing code, opening a PR, or green CI is not completion.

**Forgot to commit a file:**
```bash
git add <file>
git commit -m "fix: add forgotten file"
git push
```

**Only after completing this checklist may you stop.**
