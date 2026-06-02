package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestDiscoverPlayerTagFromValue(t *testing.T) {
	t.Parallel()

	got, err := discoverPlayerTagFromValue("#P2ABC")
	if err != nil {
		t.Fatalf("discoverPlayerTagFromValue() error = %v", err)
	}

	if got.input != "#P2ABC" {
		t.Fatalf("input = %q, want %q", got.input, "#P2ABC")
	}
	if got.sanitized != "P2ABC" {
		t.Fatalf("sanitized = %q, want %q", got.sanitized, "P2ABC")
	}
	if got.canonical != "#P2ABC" {
		t.Fatalf("canonical = %q, want %q", got.canonical, "#P2ABC")
	}
}

func TestLoadDiscoverCheckpointState(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	checkpointPath := discoverCheckpointPath("P2ABC")
	if err := os.MkdirAll(filepath.Dir(checkpointPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	want := deck.DiscoveryCheckpoint{
		GeneratorCheckpoint: &deck.GeneratorCheckpoint{
			Strategy:  deck.StrategySmartSample,
			Position:  7,
			Generated: 9,
		},
		PlayerTag: "#P2ABC",
		Strategy:  deck.StrategySmartSample,
		Timestamp: time.Unix(1234, 0),
	}
	if err := deck.SaveDiscoveryCheckpoint(checkpointPath, want); err != nil {
		t.Fatalf("SaveDiscoveryCheckpoint() error = %v", err)
	}

	got, err := loadDiscoverCheckpointState("P2ABC", "missing %s")
	if err != nil {
		t.Fatalf("loadDiscoverCheckpointState() error = %v", err)
	}

	if got.tag.sanitized != "P2ABC" {
		t.Fatalf("tag.sanitized = %q, want %q", got.tag.sanitized, "P2ABC")
	}
	if got.tag.canonical != "#P2ABC" {
		t.Fatalf("tag.canonical = %q, want %q", got.tag.canonical, "#P2ABC")
	}
	if got.checkpointPath != checkpointPath {
		t.Fatalf("checkpointPath = %q, want %q", got.checkpointPath, checkpointPath)
	}
	if got.checkpoint.PlayerTag != want.PlayerTag {
		t.Fatalf("checkpoint.PlayerTag = %q, want %q", got.checkpoint.PlayerTag, want.PlayerTag)
	}
}

func TestLoadDiscoverCheckpointStateNoCheckpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := loadDiscoverCheckpointState("P2ABC", "missing player %s")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "missing player P2ABC" {
		t.Fatalf("error = %q, want %q", err.Error(), "missing player P2ABC")
	}
}

func TestLoadDiscoverCheckpointStateInvalidCheckpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	checkpointPath := discoverCheckpointPath("P2ABC")
	if err := os.MkdirAll(filepath.Dir(checkpointPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(checkpointPath, []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := loadDiscoverCheckpointState("P2ABC", "missing %s")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, deck.ErrInvalidCheckpoint) {
		t.Fatalf("expected ErrInvalidCheckpoint, got %v", err)
	}
}
