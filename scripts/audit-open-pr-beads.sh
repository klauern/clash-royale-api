#!/usr/bin/env bash
# Audit open implementation PRs against their Beads lifecycle state.

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
cd "$repo_root"

gh_bin="${GH_BIN:-gh}"
bd_bin="${BD_BIN:-bd}"

if ! prs_json="$($gh_bin pr list --state open --limit 1000 --json number,url,body,title)"; then
    echo "error: unable to list open PRs" >&2
    exit 1
fi

rows="$(jq -r '.[] | [.number, .url, ((.body // "") | [scan("clash-royale-api-[[:alnum:]]+")] | unique | join(","))] | @tsv' <<<"$prs_json")"
failures=0

while IFS=$'\t' read -r pr_number pr_url bead_refs; do
    [[ -n "$pr_number" ]] || continue

    if [[ -z "$bead_refs" ]]; then
        echo "error: PR #$pr_number has no Bead reference ($pr_url)" >&2
        failures=$((failures + 1))
        continue
    fi

    IFS=',' read -r -a refs <<<"$bead_refs"
    if [[ "${#refs[@]}" -ne 1 ]]; then
        echo "error: PR #$pr_number has multiple Bead references: $bead_refs" >&2
        failures=$((failures + 1))
        continue
    fi

    full_bead_id="${refs[0]}"
    if [[ ! "$full_bead_id" =~ ^clash-royale-api-[[:alnum:]-]+$ ]]; then
        echo "error: PR #$pr_number has malformed Bead reference: $full_bead_id" >&2
        failures=$((failures + 1))
        continue
    fi

    if ! bead_json="$(BD_DOLT_MODE=embedded "$bd_bin" show "$full_bead_id" --json 2>/dev/null)"; then
        echo "error: PR #$pr_number references missing Bead $full_bead_id" >&2
        failures=$((failures + 1))
        continue
    fi

    bead_status="$(jq -r 'if type == "array" then .[0].status else .status end // empty' <<<"$bead_json")"
    if [[ -z "$bead_status" ]]; then
        echo "error: PR #$pr_number references Bead $full_bead_id with no readable status" >&2
        failures=$((failures + 1))
        continue
    fi

    if [[ "$bead_status" == "closed" ]]; then
        echo "error: open PR #$pr_number is tied to closed Bead $full_bead_id" >&2
        failures=$((failures + 1))
        continue
    fi

    echo "ok: PR #$pr_number -> $full_bead_id ($bead_status)"
done <<<"$rows"

if [[ "$failures" -ne 0 ]]; then
    echo "open PR/Bead audit failed: $failures problem(s)" >&2
    exit 1
fi

echo "open PR/Bead audit passed"
