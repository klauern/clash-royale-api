#!/usr/bin/env bash
# Close a PR-backed Bead only after GitHub confirms that its PR merged.

set -euo pipefail

usage() {
    echo "usage: $0 <bead-id> <pr-number>" >&2
    exit 2
}

[[ $# -eq 2 ]] || usage

bead_id="${1#clash-royale-api-}"
pr_number="$2"

if [[ -z "$bead_id" || ! "$bead_id" =~ ^[[:alnum:]-]+$ ]]; then
    echo "error: invalid Bead ID: $1" >&2
    exit 2
fi

if [[ ! "$pr_number" =~ ^[0-9]+$ ]]; then
    echo "error: invalid PR number: $pr_number" >&2
    exit 2
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
cd "$repo_root"

gh_bin="${GH_BIN:-gh}"
bd_bin="${BD_BIN:-bd}"
full_bead_id="clash-royale-api-$bead_id"

if ! pr_json="$($gh_bin pr view "$pr_number" --json number,state,mergedAt,body,url)"; then
    echo "error: unable to inspect PR #$pr_number" >&2
    exit 1
fi

pr_state="$(jq -r 'if type == "array" then .[0].state else .state end // empty' <<<"$pr_json")"
merged_at="$(jq -r 'if type == "array" then .[0].mergedAt else .mergedAt end // empty' <<<"$pr_json")"
pr_url="$(jq -r 'if type == "array" then .[0].url else .url end // empty' <<<"$pr_json")"

if [[ -z "$merged_at" ]]; then
    echo "error: PR #$pr_number is not merged (state=$pr_state)" >&2
    exit 1
fi

if ! jq -e --arg bead "$full_bead_id" \
    'if type == "array" then .[0].body else .body end // "" | contains($bead)' \
    <<<"$pr_json" >/dev/null; then
    echo "error: PR #$pr_number does not reference $full_bead_id" >&2
    exit 1
fi

if ! bead_json="$(BD_DOLT_MODE=embedded "$bd_bin" show "$full_bead_id" --json)"; then
    echo "error: unable to inspect $full_bead_id" >&2
    exit 1
fi

bead_status="$(jq -r 'if type == "array" then .[0].status else .status end // empty' <<<"$bead_json")"
if [[ -z "$bead_status" ]]; then
    echo "error: $full_bead_id has no readable status" >&2
    exit 1
fi

if [[ "$bead_status" == "closed" ]]; then
    echo "$full_bead_id is already closed; verified PR #$pr_number merged at $merged_at"
    exit 0
fi

reason="Completed by merged PR #$pr_number"
if [[ -n "$pr_url" ]]; then
    reason="$reason ($pr_url)"
fi

BD_DOLT_MODE=embedded "$bd_bin" close "$full_bead_id" --reason "$reason" --json
