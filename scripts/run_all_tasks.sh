#!/bin/bash
set +e  # Continue on task failures, but we'll handle errors manually

# Script to run all Taskfile tasks and build debug list
# Based on plan: async-coalescing-frost.md

# Configuration
LOG_DIR="./task_execution_logs"
DEBUG_LIST="${LOG_DIR}/debug_list.txt"
mkdir -p "${LOG_DIR}"

# Initialize debug list
echo "=== Task Execution Debug List ===" > "${DEBUG_LIST}"
echo "Generated: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "${DEBUG_LIST}"
echo "" >> "${DEBUG_LIST}"

# Helper function to execute a task and capture results
execute_task() {
    local task_name="$1"
    local task_args="$2"
    local start_time=$(date +%s)

    echo "[$(date +'%H:%M:%S')] Starting: $task_name" | tee -a "${DEBUG_LIST}"
    echo "  Command: task $task_name $task_args" >> "${DEBUG_LIST}"

    # Execute with output capture
    task $task_name $task_args > "${LOG_DIR}/${task_name}.stdout.log" 2> "${LOG_DIR}/${task_name}.stderr.log"
    local exit_code=$?
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    # Determine status
    if [ $exit_code -eq 0 ]; then
        local status="SUCCESS"
    else
        local status="FAILURE"
    fi

    # Count output lines
    local stdout_lines=$(wc -l < "${LOG_DIR}/${task_name}.stdout.log" 2>/dev/null || echo "0")
    local stderr_lines=$(wc -l < "${LOG_DIR}/${task_name}.stderr.log" 2>/dev/null || echo "0")

    # Log to debug list
    echo "[$status] $task_name (${duration}s)" >> "${DEBUG_LIST}"
    echo "  Exit code: $exit_code" >> "${DEBUG_LIST}"
    echo "  Output: $stdout_lines lines to stdout, $stderr_lines errors" >> "${DEBUG_LIST}"

    # Add error details if any
    if [ $stderr_lines -gt 0 ]; then
        echo "  Last error: $(tail -1 "${LOG_DIR}/${task_name}.stderr.log" 2>/dev/null || echo "Error log not readable")" >> "${DEBUG_LIST}"
    fi
    echo "" >> "${DEBUG_LIST}"

    return $exit_code
}

# Check prerequisites
echo "Checking prerequisites..." | tee -a "${DEBUG_LIST}"

# 1. Taskfile exists
if [ ! -f "Taskfile.yml" ]; then
    echo "❌ Taskfile.yml not found" | tee -a "${DEBUG_LIST}"
    exit 1
fi
echo "✅ Taskfile.yml exists" | tee -a "${DEBUG_LIST}"

# 2. Task binary available
if ! command -v task &> /dev/null; then
    echo "❌ 'task' binary not found. Install from https://taskfile.dev/" | tee -a "${DEBUG_LIST}"
    exit 1
fi
echo "✅ 'task' binary available" | tee -a "${DEBUG_LIST}"

# 3. .env file exists
if [ ! -f ".env" ]; then
    echo "❌ .env file not found" | tee -a "${DEBUG_LIST}"
    exit 1
fi
echo "✅ .env file exists" | tee -a "${DEBUG_LIST}"

# 4. API token configured
if ! grep -q "CLASH_ROYALE_API_TOKEN=" .env; then
    echo "❌ CLASH_ROYALE_API_TOKEN not configured in .env" | tee -a "${DEBUG_LIST}"
    exit 1
fi
echo "✅ API token configured in .env" | tee -a "${DEBUG_LIST}"

# 5. Go binaries exist
if [ ! -x "go/bin/cr-api" ] && [ ! -x "bin/cr-api" ]; then
    echo "⚠️  cr-api binary not found or not executable. Some tasks may fail." | tee -a "${DEBUG_LIST}"
    echo "   Run 'task build' to build it." | tee -a "${DEBUG_LIST}"
fi
echo "✅ Go binaries check completed" | tee -a "${DEBUG_LIST}"

echo "" >> "${DEBUG_LIST}"
echo "=== Starting Task Execution ===" >> "${DEBUG_LIST}"
echo "" >> "${DEBUG_LIST}"

# Export player tag from .env or use default
if grep -q "DEFAULT_PLAYER_TAG=" .env; then
    DEFAULT_PLAYER_TAG=$(grep "DEFAULT_PLAYER_TAG=" .env | cut -d'=' -f2- | tr -d '"' | tr -d "'")
    echo "Using DEFAULT_PLAYER_TAG from .env: $DEFAULT_PLAYER_TAG" | tee -a "${DEBUG_LIST}"
else
    DEFAULT_PLAYER_TAG="#PLAYERTAG"
    echo "Using default player tag: $DEFAULT_PLAYER_TAG" | tee -a "${DEBUG_LIST}"
fi
export DEFAULT_PLAYER_TAG

# Phase 1: Setup & Prerequisites (Non-destructive)
echo "--- Phase 1: Setup & Prerequisites ---" | tee -a "${DEBUG_LIST}"
execute_task "default" ""
execute_task "status" ""
execute_task "help-api" ""

# Phase 2: Build & Installation
echo "--- Phase 2: Build & Installation ---" | tee -a "${DEBUG_LIST}"
execute_task "install" ""
execute_task "build-go" ""

# Phase 3: Analysis Tasks (Require API token & player tag)
echo "--- Phase 3: Analysis Tasks ---" | tee -a "${DEBUG_LIST}"
execute_task "run" "-- ${DEFAULT_PLAYER_TAG}"
execute_task "run-with-save" "-- ${DEFAULT_PLAYER_TAG}"
execute_task "export-csv" "-- ${DEFAULT_PLAYER_TAG}"
execute_task "export-all" "-- ${DEFAULT_PLAYER_TAG}"
execute_task "build-deck" "-- ${DEFAULT_PLAYER_TAG}"

# Phase 4: Event Tasks
echo "--- Phase 4: Event Tasks ---" | tee -a "${DEBUG_LIST}"
execute_task "scan-events" "-- ${DEFAULT_PLAYER_TAG}"
execute_task "export-events" "-- ${DEFAULT_PLAYER_TAG}"
execute_task "export-decks" "-- ${DEFAULT_PLAYER_TAG}"
execute_task "analyze-events" "-- ${DEFAULT_PLAYER_TAG}"
execute_task "sync-events" "-- ${DEFAULT_PLAYER_TAG}"

# Phase 5: Development & Testing
echo "--- Phase 5: Development & Testing ---" | tee -a "${DEBUG_LIST}"
execute_task "test" ""
execute_task "test-go" ""
execute_task "lint" ""
execute_task "format" ""
# SKIP: clean task (destructive, per user request)
echo "[SKIPPED] clean" >> "${DEBUG_LIST}"
echo "  Reason: User requested skip to preserve data" >> "${DEBUG_LIST}"
echo "" >> "${DEBUG_LIST}"

# Phase 6: Special Tasks
echo "--- Phase 6: Special Tasks ---" | tee -a "${DEBUG_LIST}"
execute_task "setup" ""

# Handle dev task with timeout (special handling)
echo "[$(date +'%H:%M:%S')] Starting: dev (with 30s timeout)" | tee -a "${DEBUG_LIST}"
echo "  Command: task dev" >> "${DEBUG_LIST}"
task dev > "${LOG_DIR}/dev.stdout.log" 2> "${LOG_DIR}/dev.stderr.log" &
DEV_PID=$!
sleep 30
if kill -0 $DEV_PID 2>/dev/null; then
    kill $DEV_PID 2>/dev/null
    echo "[TIMEOUT] dev (30s)" >> "${DEBUG_LIST}"
    echo "  Note: Killed after 30-second timeout" >> "${DEBUG_LIST}"
else
    wait $DEV_PID
    exit_code=$?
    if [ $exit_code -eq 0 ]; then
        echo "[SUCCESS] dev (30s)" >> "${DEBUG_LIST}"
    else
        echo "[FAILURE] dev (30s)" >> "${DEBUG_LIST}"
    fi
    echo "  Exit code: $exit_code" >> "${DEBUG_LIST}"
fi
echo "" >> "${DEBUG_LIST}"

# Final summary
echo "=== Execution Complete ===" >> "${DEBUG_LIST}"
echo "Logs available in: ${LOG_DIR}/" >> "${DEBUG_LIST}"
echo "" >> "${DEBUG_LIST}"

# Generate summary statistics
success_count=$(grep -c "^\[SUCCESS\]" "${DEBUG_LIST}" || true)
failure_count=$(grep -c "^\[FAILURE\]" "${DEBUG_LIST}" || true)
timeout_count=$(grep -c "^\[TIMEOUT\]" "${DEBUG_LIST}" || true)
skipped_count=$(grep -c "^\[SKIPPED\]" "${DEBUG_LIST}" || true)

echo "=== Summary Statistics ===" >> "${DEBUG_LIST}"
echo "Successful: $success_count" >> "${DEBUG_LIST}"
echo "Failed: $failure_count" >> "${DEBUG_LIST}"
echo "Timeout: $timeout_count" >> "${DEBUG_LIST}"
echo "Skipped: $skipped_count" >> "${DEBUG_LIST}"
echo "Total tasks attempted: $((success_count + failure_count + timeout_count))" >> "${DEBUG_LIST}"

echo "" >> "${DEBUG_LIST}"
echo "Debug list saved to: ${DEBUG_LIST}"
echo "Individual task logs in: ${LOG_DIR}/"