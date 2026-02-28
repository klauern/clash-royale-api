package deck

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauer/clash-royale-api/go/internal/storage"
)

// DefaultDiscoveryCheckpointDir returns the default discovery checkpoint directory.
func DefaultDiscoveryCheckpointDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".cr-api", "discover")
}

// DiscoveryCheckpointPath returns the checkpoint file path for a player tag.
func DiscoveryCheckpointPath(checkpointDir, playerTag string) string {
	return filepath.Join(checkpointDir, fmt.Sprintf("%s.json", playerTag))
}

// SaveDiscoveryCheckpoint writes the checkpoint JSON to disk.
func SaveDiscoveryCheckpoint(path string, checkpoint DiscoveryCheckpoint) error {
	if err := storage.WriteJSON(path, checkpoint); err != nil {
		return fmt.Errorf("failed to write checkpoint: %w", err)
	}
	return nil
}

// LoadDiscoveryCheckpoint reads and validates a discovery checkpoint from disk.
func LoadDiscoveryCheckpoint(path string) (DiscoveryCheckpoint, error) {
	var checkpoint DiscoveryCheckpoint
	if err := storage.ReadJSON(path, &checkpoint); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return DiscoveryCheckpoint{}, ErrNoCheckpoint
		}
		if isInvalidCheckpointError(err) {
			return DiscoveryCheckpoint{}, ErrInvalidCheckpoint
		}
		return DiscoveryCheckpoint{}, fmt.Errorf("failed to read checkpoint: %w", err)
	}

	if checkpoint.GeneratorCheckpoint == nil {
		return DiscoveryCheckpoint{}, ErrInvalidCheckpoint
	}

	return checkpoint, nil
}

func isInvalidCheckpointError(err error) bool {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var invalidUnmarshalErr *json.InvalidUnmarshalError

	// storage.ReadJSON wraps json.Unmarshal errors in a file-specific message, so
	// we retain this substring check until storage exposes a sentinel or typed error.
	return errors.As(err, &syntaxErr) ||
		errors.As(err, &typeErr) ||
		errors.As(err, &invalidUnmarshalErr) ||
		strings.Contains(err.Error(), "failed to unmarshal JSON")
}
