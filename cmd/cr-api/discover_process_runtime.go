package main

import (
	"fmt"
	"os"
	"syscall"
)

func readDiscoverPID(pidFile string) (int, error) {
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read PID file: %w", err)
	}

	var pid int
	if _, err := fmt.Sscanf(string(pidData), "%d", &pid); err != nil {
		return 0, fmt.Errorf("failed to parse PID file: %w", err)
	}

	return pid, nil
}

func discoverProcessAlive(pid int) (bool, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, fmt.Errorf("failed to find process: %w", err)
	}

	return process.Signal(syscall.Signal(0)) == nil, nil
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
