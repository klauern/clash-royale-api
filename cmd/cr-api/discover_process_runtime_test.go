package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDiscoverPID(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	pidPath := filepath.Join(tmp, "discover.pid")
	if err := os.WriteFile(pidPath, []byte("12345"), 0o644); err != nil {
		t.Fatalf("write pid file: %v", err)
	}

	pid, err := readDiscoverPID(pidPath)
	if err != nil {
		t.Fatalf("readDiscoverPID returned error: %v", err)
	}
	if pid != 12345 {
		t.Fatalf("expected pid=12345, got %d", pid)
	}
}

func TestReadDiscoverPIDInvalid(t *testing.T) {
	t.Parallel()

	for _, tc := range []string{"not-a-pid", "123abc", "0", "-1"} {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			t.Parallel()

			tmp := t.TempDir()
			pidPath := filepath.Join(tmp, "discover.pid")
			if err := os.WriteFile(pidPath, []byte(tc), 0o644); err != nil {
				t.Fatalf("write pid file: %v", err)
			}

			if _, err := readDiscoverPID(pidPath); err == nil {
				t.Fatalf("expected parse error for invalid pid file")
			}
		})
	}
}

func TestCheckAndCleanupStaleDiscoverPIDRemovesStalePID(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	pidPath := filepath.Join(tmp, "discover.pid")
	if err := os.WriteFile(pidPath, []byte("999999"), 0o644); err != nil {
		t.Fatalf("write pid file: %v", err)
	}

	pid, running, err := checkAndCleanupStaleDiscoverPID(pidPath)
	if err != nil {
		t.Fatalf("checkAndCleanupStaleDiscoverPID returned error: %v", err)
	}
	if running {
		t.Fatalf("expected stale pid to be reported as not running")
	}
	if pid != 999999 {
		t.Fatalf("expected pid=999999, got %d", pid)
	}
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Fatalf("expected stale pid file to be removed, stat err=%v", err)
	}
}
