package deck

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDiscoveryCheckpoint_NoCheckpoint(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing.json")
	_, err := LoadDiscoveryCheckpoint(path)
	if !errors.Is(err, ErrNoCheckpoint) {
		t.Fatalf("expected ErrNoCheckpoint, got %v", err)
	}
}

func TestSaveAndLoadDiscoveryCheckpoint(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "checkpoint.json")
	input := DiscoveryCheckpoint{
		GeneratorCheckpoint: &GeneratorCheckpoint{
			Generated: 10,
			State:     map[string]any{"remaining": 20},
			Strategy:  StrategyRandomSample,
		},
		Stats: DiscoveryStats{
			Evaluated: 10,
			Stored:    3,
			AvgScore:  76.5,
		},
		Timestamp: time.Now().UTC().Truncate(time.Second),
		PlayerTag: "TEST123",
		Strategy:  StrategyRandomSample,
	}

	if err := SaveDiscoveryCheckpoint(path, input); err != nil {
		t.Fatalf("SaveDiscoveryCheckpoint() error = %v", err)
	}

	got, err := LoadDiscoveryCheckpoint(path)
	if err != nil {
		t.Fatalf("LoadDiscoveryCheckpoint() error = %v", err)
	}
	if got.PlayerTag != input.PlayerTag {
		t.Fatalf("player tag = %q, want %q", got.PlayerTag, input.PlayerTag)
	}
	if got.GeneratorCheckpoint == nil {
		t.Fatal("expected generator checkpoint to be present")
	}
	if got.GeneratorCheckpoint.Generated != input.GeneratorCheckpoint.Generated {
		t.Fatalf("generated = %d, want %d", got.GeneratorCheckpoint.Generated, input.GeneratorCheckpoint.Generated)
	}
}

func TestLoadDiscoveryCheckpoint_InvalidJSON(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "checkpoint.json")
	if err := os.WriteFile(path, []byte("{broken"), 0o644); err != nil {
		t.Fatalf("failed to write invalid checkpoint: %v", err)
	}

	_, err := LoadDiscoveryCheckpoint(path)
	if !errors.Is(err, ErrInvalidCheckpoint) {
		t.Fatalf("expected ErrInvalidCheckpoint, got %v", err)
	}
}
