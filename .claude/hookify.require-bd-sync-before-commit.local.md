---
name: require-bd-sync-before-commit
enabled: true
event: bash
pattern: git\s+commit
action: warn
---

🚨 **WARNING: Git commit without bd sync**

You must run `bd sync` before committing to ensure the beads database is synchronized with your code changes.

**Required workflow:**
1. ✅ Make code changes
2. ⚠️ **Run `bd sync` first** (you're here)
3. ✅ Then `git add` and `git commit`

**Why this matters:**
- Beads database changes must be committed alongside code
- Prevents desync between issues.jsonl and actual code state
- Ensures team sees your task status updates

**To proceed:**
```bash
bd sync
git add .
git commit -m "your message"
```
