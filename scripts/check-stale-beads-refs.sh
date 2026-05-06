#!/usr/bin/env bash
#
# check-stale-beads-refs.sh
# Scans the repo for clash-royale-api-<id> references and verifies that each
# referenced ID still exists as an open or closed beads issue. Exits non-zero
# (with a list of stale references) when one of them no longer resolves —
# useful as a pre-PR/CI guard against nolint comments and design notes that
# rot when an issue is renamed, superseded, or deleted.
#
# Usage:
#   scripts/check-stale-beads-refs.sh           # scan and report
#   scripts/check-stale-beads-refs.sh --strict  # exit non-zero on stale refs

set -euo pipefail

cd "$(dirname "$0")/.."

if ! command -v bd >/dev/null 2>&1; then
  echo "bd CLI not found in PATH; skipping stale-beads-ref check" >&2
  exit 0
fi

# Collect referenced IDs from source code and markdown — skip the issues.jsonl
# file itself (it's bd's own export and naturally lists every ID).
# Strip rg's "file:" prefix so the comm comparison sees just the IDs.
referenced=$(rg -o --no-heading -N -I 'clash-royale-api-[a-z0-9]+' \
  --glob '!.beads/issues.jsonl' \
  --glob '!.beads/backup/**' \
  --glob '!.git/**' \
  -- . 2>/dev/null | sort -u || true)

if [[ -z "$referenced" ]]; then
  echo "No clash-royale-api-* references found."
  exit 0
fi

# Pull the live issue list once.
known=$(BD_DOLT_MODE=embedded bd list --status=all --json 2>/dev/null \
  | python3 -c 'import json,sys; [print(i["id"]) for i in json.load(sys.stdin)]' \
  | sort -u)

stale=$(comm -23 <(printf '%s\n' "$referenced") <(printf '%s\n' "$known"))

if [[ -z "$stale" ]]; then
  echo "All beads references resolve."
  exit 0
fi

echo "Stale beads references found (no matching issue in bd):"
while IFS= read -r id; do
  echo "  $id"
  rg -n --no-heading "$id" \
    --glob '!.beads/issues.jsonl' \
    --glob '!.beads/backup/**' \
    --glob '!.git/**' \
    -- . | sed 's/^/    /'
done <<< "$stale"

if [[ "${1:-}" == "--strict" ]]; then
  exit 1
fi
exit 0
