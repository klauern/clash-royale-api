package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/storage"
)

const artifactTimestampLayout = "20060102_150405"

type taggedJSONArtifactOptions struct {
	subdir      string
	fileStem    string
	timestamped bool
}

func saveTaggedJSONArtifact(dataDir, playerTag string, payload any, opts taggedJSONArtifactOptions) (string, error) {
	sanitizedTag, err := storage.SanitizePlayerTag(playerTag)
	if err != nil {
		return "", fmt.Errorf("failed to sanitize player tag %q: %w", playerTag, err)
	}

	filename := fmt.Sprintf("%s_%s.json", opts.fileStem, sanitizedTag)
	if opts.timestamped {
		filename = fmt.Sprintf("%s_%s_%s.json", time.Now().Format(artifactTimestampLayout), opts.fileStem, sanitizedTag)
	}

	path := filepath.Join(dataDir, opts.subdir, filename)
	if err := storage.WriteJSON(path, payload); err != nil {
		return "", err
	}

	return path, nil
}
