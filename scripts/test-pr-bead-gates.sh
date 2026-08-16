#!/usr/bin/env bash
# Black-box tests for the PR/Bead lifecycle guard scripts.

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
close_guard="$script_dir/close-bead-after-merge.sh"
audit="$script_dir/audit-open-pr-beads.sh"
tmp_dir="$(mktemp -d /tmp/pr-bead-gates.XXXXXX)"
trap 'rm -rf "$tmp_dir"' EXIT

gh_mock="$tmp_dir/gh"
bd_mock="$tmp_dir/bd"
gh_pr_fixture="$tmp_dir/gh-pr.json"
gh_list_fixture="$tmp_dir/gh-list.json"
bd_fixture="$tmp_dir/bd.json"
bd_close_log="$tmp_dir/bd-close.log"

printf '%s\n' '#!/usr/bin/env bash' \
    'set -euo pipefail' \
    'if [[ "$1 $2" == "pr view" ]]; then cat "$GH_PR_FIXTURE"; exit 0; fi' \
    'if [[ "$1 $2" == "pr list" ]]; then cat "$GH_LIST_FIXTURE"; exit 0; fi' \
    'echo "unexpected gh invocation: $*" >&2; exit 2' > "$gh_mock"

printf '%s\n' '#!/usr/bin/env bash' \
    'set -euo pipefail' \
    'if [[ "$1" == "show" ]]; then cat "$BD_FIXTURE"; exit 0; fi' \
    'if [[ "$1" == "close" ]]; then printf "%s\\n" "$*" > "$BD_CLOSE_LOG"; exit 0; fi' \
    'echo "unexpected bd invocation: $*" >&2; exit 2' > "$bd_mock"
chmod +x "$gh_mock" "$bd_mock"

expect_failure() {
    if "$@" >/dev/null 2>&1; then
        echo "expected failure: $*" >&2
        exit 1
    fi
}

export GH_BIN="$gh_mock"
export BD_BIN="$bd_mock"
export GH_PR_FIXTURE="$gh_pr_fixture"
export GH_LIST_FIXTURE="$gh_list_fixture"
export BD_FIXTURE="$bd_fixture"
export BD_CLOSE_LOG="$bd_close_log"

# Merged PR closes an open Bead.
printf '%s\n' '{"number":123,"state":"MERGED","mergedAt":"2026-08-16T15:00:00Z","body":"Closes clash-royale-api-test","url":"https://github.com/example/repo/pull/123"}' > "$gh_pr_fixture"
printf '%s\n' '[{"status":"open"}]' > "$bd_fixture"
"$close_guard" test 123 >/dev/null
grep -q 'clash-royale-api-test' "$bd_close_log"

# A second run is idempotent after the Bead is already closed.
printf '%s\n' '[{"status":"closed"}]' > "$bd_fixture"
"$close_guard" test 123 >/dev/null

# Open, closed-unmerged, and mismatched PRs are refused.
printf '%s\n' '{"number":123,"state":"OPEN","mergedAt":null,"body":"Closes clash-royale-api-test","url":"https://github.com/example/repo/pull/123"}' > "$gh_pr_fixture"
expect_failure "$close_guard" test 123
printf '%s\n' '{"number":123,"state":"CLOSED","mergedAt":null,"body":"Closes clash-royale-api-test","url":"https://github.com/example/repo/pull/123"}' > "$gh_pr_fixture"
expect_failure "$close_guard" test 123
printf '%s\n' '{"number":123,"state":"MERGED","mergedAt":"2026-08-16T15:00:00Z","body":"Closes clash-royale-api-other","url":"https://github.com/example/repo/pull/123"}' > "$gh_pr_fixture"
expect_failure "$close_guard" test 123

# Audit accepts linked non-closed Beads.
printf '%s\n' '[{"number":123,"url":"https://github.com/example/repo/pull/123","body":"Closes clash-royale-api-test","title":"test"},{"number":124,"url":"https://github.com/example/repo/pull/124","body":"Tracks clash-royale-api-other","title":"other"}]' > "$gh_list_fixture"
printf '%s\n' '[{"status":"open"}]' > "$bd_fixture"
"$audit" >/dev/null

# Audit rejects missing, duplicate, and closed Bead links.
printf '%s\n' '[{"number":123,"url":"https://github.com/example/repo/pull/123","body":"No task link","title":"missing"}]' > "$gh_list_fixture"
expect_failure "$audit"
printf '%s\n' '[{"number":123,"url":"https://github.com/example/repo/pull/123","body":"clash-royale-api-test clash-royale-api-other","title":"duplicate"}]' > "$gh_list_fixture"
expect_failure "$audit"
printf '%s\n' '[{"number":123,"url":"https://github.com/example/repo/pull/123","body":"clash-royale-api-test","title":"closed"}]' > "$gh_list_fixture"
printf '%s\n' '[{"status":"closed"}]' > "$bd_fixture"
expect_failure "$audit"

echo "PR/Bead guard tests passed"
