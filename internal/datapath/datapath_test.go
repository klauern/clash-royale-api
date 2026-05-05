package datapath

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathsUseCrAPIHome(t *testing.T) {
	tmp := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Setenv("HOME", tmp)
	defer os.Setenv("HOME", originalHome)

	appDir, err := AppDir()
	if err != nil {
		t.Fatalf("AppDir() failed: %v", err)
	}
	if appDir != filepath.Join(tmp, ".cr-api") {
		t.Fatalf("unexpected app dir: %s", appDir)
	}

	leaderboardsDir, err := LeaderboardsDir()
	if err != nil {
		t.Fatalf("LeaderboardsDir() failed: %v", err)
	}
	if leaderboardsDir != filepath.Join(tmp, ".cr-api", "leaderboards") {
		t.Fatalf("unexpected leaderboards dir: %s", leaderboardsDir)
	}

	leaderboardDB, err := LeaderboardDBPath("ABC123")
	if err != nil {
		t.Fatalf("LeaderboardDBPath() failed: %v", err)
	}
	if leaderboardDB != filepath.Join(tmp, ".cr-api", "leaderboards", "ABC123.db") {
		t.Fatalf("unexpected leaderboard db path: %s", leaderboardDB)
	}

	fuzzDB, err := FuzzStorageDBPath("fuzz_top_decks.db")
	if err != nil {
		t.Fatalf("FuzzStorageDBPath() failed: %v", err)
	}
	if fuzzDB != filepath.Join(tmp, ".cr-api", "fuzz_top_decks.db") {
		t.Fatalf("unexpected fuzz db path: %s", fuzzDB)
	}
}
