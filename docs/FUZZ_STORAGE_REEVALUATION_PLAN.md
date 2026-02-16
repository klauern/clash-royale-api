# Fuzz Storage Re-Evaluation Plan

## Goal
Enable repeatable re-scoring of stored fuzz decks when scoring logic changes and when player-specific context is available, without losing historical provenance.

## Current State
- Fuzz storage (`pkg/fuzzstorage`) stores one mutable row per unique deck hash.
- `deck fuzz update` re-evaluates decks and overwrites scores in place.
- Storage has no explicit schema support for:
  - scoring/version provenance,
  - player-level context metadata,
  - selective re-evaluation scheduling.

## Key Requirements
1. Re-evaluate decks with and without player context.
2. Preserve provenance for each score update.
3. Support targeted updates (subset by version/context/archetype).
4. Keep CLI behavior simple for default usage.

## Data Model Changes
### 1) Extend `top_decks`
Add columns:
- `score_version TEXT NOT NULL DEFAULT 'v1'`
- `context_mode TEXT NOT NULL DEFAULT 'global'`
  - values: `global`, `player`
- `context_player_tag TEXT`
- `updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`

Keep existing `evaluated_at` as original evaluation timestamp for compatibility.

### 2) Add `re_eval_runs`
New table to capture run provenance:
- `id INTEGER PRIMARY KEY AUTOINCREMENT`
- `started_at DATETIME NOT NULL`
- `finished_at DATETIME`
- `trigger TEXT NOT NULL` (manual, migration, scheduled)
- `score_version_from TEXT`
- `score_version_to TEXT NOT NULL`
- `context_mode TEXT NOT NULL`
- `context_player_tag TEXT`
- `filters_json TEXT` (serialized CLI filters)
- `updated_rows INTEGER NOT NULL DEFAULT 0`
- `notes TEXT`

### 3) Optional audit table (phase 2)
If full history is needed later, add `top_deck_score_history` and append pre-update snapshots.

## Update Selection Strategy
`deck fuzz update` should select rows by:
- score version (`--from-version`, default any older than current)
- context (`--tag` implies player-context evaluation)
- existing filters (`--top`, `--archetype`, `--min-score`, etc.)
- optional force flag (`--force`) to re-score rows already at target version.

## CLI Behavior
### Default (`global`)
`cr-api deck fuzz update`
- Uses global scoring mode.
- Re-scores rows with outdated score versions.
- Writes `score_version`, `context_mode=global`, `context_player_tag=NULL`.

### Player-aware
`cr-api deck fuzz update --tag <TAG>`
- Loads player data/context.
- Re-scores with level-aware logic.
- Writes `context_mode=player`, `context_player_tag=<TAG>`.

### Proposed new flags
- `--score-version <version>`: target output version (defaults to current evaluator version)
- `--from-version <version>`: only rows from a source version
- `--force`: re-score even when already on target version
- `--dry-run`: show affected row counts only
- `--run-note <text>`: annotate `re_eval_runs`

## Migration Path
1. Add schema migration helpers in `pkg/fuzzstorage` (idempotent ALTER + CREATE TABLE).
2. Backfill existing rows:
   - `score_version='legacy'`
   - `context_mode='global'`
   - `updated_at=evaluated_at` when possible.
3. Ship update command changes with compatibility fallback for pre-migration DBs.
4. Add tests for migration + selection + provenance row writes.

## Rollout Phases
1. Phase 1: Schema + metadata writes + `--dry-run`.
2. Phase 2: Version-based selective update and run logging.
3. Phase 3: Optional history table and rollback tooling.

## Validation
- Unit tests:
  - migration/backfill idempotence,
  - run logging,
  - selection correctness by version/context.
- CLI tests:
  - `--dry-run` output,
  - player-context writes.
- Performance checks:
  - update throughput,
  - DB growth after repeated re-evaluations.
