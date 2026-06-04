package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func readDiscoverPID(pidFile string) (int, error) {
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read PID file: %w", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID file: %w", err)
	}
	if pid <= 0 {
		return 0, fmt.Errorf("invalid PID %d", pid)
	}

	return pid, nil
}

func discoverProcessAlive(pid int) (bool, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, fmt.Errorf("failed to find process: %w", err)
	}

	err = process.Signal(syscall.Signal(0))
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, syscall.EPERM):
		return true, nil
	case errors.Is(err, syscall.ESRCH), errors.Is(err, os.ErrProcessDone):
		return false, nil
	default:
		return false, fmt.Errorf("failed to signal process: %w", err)
	}
}

func checkAndCleanupStaleDiscoverPID(pidFile string) (int, bool, error) {
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return 0, false, nil
	}

	pid, err := readDiscoverPID(pidFile)
	if err != nil {
		return 0, false, err
	}

	alive, err := discoverProcessAlive(pid)
	if err != nil {
		return 0, false, err
	}
	if alive {
		return pid, true, nil
	}

	if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
		return 0, false, fmt.Errorf("failed to remove stale PID file: %w", err)
	}
	return pid, false, nil
}
