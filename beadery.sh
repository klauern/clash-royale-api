#!/usr/bin/env bash
set -euo pipefail

MAX_TURNS="${MAX_TURNS:-12}"

while true; do
  echo "=== Calling /bd-work-now (max turns: ${MAX_TURNS}) ==="

  # Use the built-in skill that handles bead selection, work, and commit
  zai -p \
    --max-turns "${MAX_TURNS}" \
    "/bd-work-now"

  # Check if there are any ready beads left
  if ! bd ready 2>&1 | grep -q "Ready work"; then
    echo "No more ready beads. Exiting."
    exit 0
  fi
done
