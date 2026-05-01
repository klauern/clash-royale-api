#!/usr/bin/env bash
# Detect stale beads references in source comments.
#
# Compares clash-royale-api-* IDs found in cmd/, internal/, pkg/ against the
# live IDs in .beads/issues.jsonl. Exits 1 if any unknown ID is referenced.

set -euo pipefail

cd "$(dirname "$0")/.."

if [[ ! -f .beads/issues.jsonl ]]; then
    echo "error: .beads/issues.jsonl not found" >&2
    exit 2
fi

valid_ids="$(grep -oE '"id"[[:space:]]*:[[:space:]]*"clash-royale-api-[a-z0-9]+"' .beads/issues.jsonl \
    | sed -E 's/.*"(clash-royale-api-[a-z0-9]+)".*/\1/' \
    | sort -u)"

if [[ -z "$valid_ids" ]]; then
    echo "error: no beads IDs found in .beads/issues.jsonl" >&2
    exit 2
fi

stale_found=0
while IFS=: read -r file line id; do
    if ! grep -qx "$id" <<<"$valid_ids"; then
        echo "stale: $file:$line -> $id"
        stale_found=1
    fi
done < <(grep -rEon 'clash-royale-api-[a-z0-9]+' cmd/ internal/ pkg/ 2>/dev/null || true)

if [[ "$stale_found" -ne 0 ]]; then
    echo
    echo "Stale beads references found. Update or remove them."
    echo "Valid IDs:"
    echo "$valid_ids" | sed 's/^/  /'
    exit 1
fi

echo "no stale beads references"
