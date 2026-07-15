package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/playertag"
	"github.com/klauer/clash-royale-api/go/internal/storage"
)

const artifactTimestampLayout = "20060102_150405"

type taggedJSONArtifactOptions struct {
	subdir      string
	fileStem    string
	timestamped bool
}

type taggedTextArtifactOptions struct {
	subdir      string
	fileStem    string
	extension   string
	timestamped bool
	saveMessage string
}

type timestampedJSONArtifactOptions struct {
	subdir    string
	fileStem  string
	timestamp time.Time
}

func saveTaggedJSONArtifact(dataDir, playerTag string, payload any, opts taggedJSONArtifactOptions) (string, error) {
	path, err := buildTaggedArtifactPath(dataDir, playerTag, opts.subdir, opts.fileStem, "json", opts.timestamped)
	if err != nil {
		return "", err
	}
	if err := storage.WriteJSON(path, payload); err != nil {
		return "", err
	}
	return path, nil
}

func saveTaggedTextArtifact(dataDir, playerTag, content string, opts taggedTextArtifactOptions) (string, error) {
	path, err := buildTaggedArtifactPath(dataDir, playerTag, opts.subdir, opts.fileStem, opts.extension, opts.timestamped)
	if err != nil {
		return "", err
	}
	if err := writeTextOutput(content, path, textOutputOptions{saveMessage: opts.saveMessage}); err != nil {
		return "", err
	}
	return path, nil
}

func buildTaggedArtifactPath(dataDir, playerTag, subdir, fileStem, extension string, timestamped bool) (string, error) {
	sanitizedTag, err := playertag.Sanitize(playerTag)
	if err != nil {
		return "", fmt.Errorf("failed to sanitize player tag %q: %w", playerTag, err)
	}

	filename := fmt.Sprintf("%s_%s.%s", fileStem, sanitizedTag, extension)
	if timestamped {
		filename = fmt.Sprintf("%s_%s_%s.%s", time.Now().Format(artifactTimestampLayout), fileStem, sanitizedTag, extension)
	}

	return filepath.Join(dataDir, subdir, filename), nil
}

func saveTimestampedJSONArtifact(dataDir string, payload any, opts timestampedJSONArtifactOptions) (string, error) {
	timestamp := opts.timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	filename := fmt.Sprintf("%s_%s.json", opts.fileStem, timestamp.Format(artifactTimestampLayout))
	path := filepath.Join(dataDir, opts.subdir, filename)
	if err := storage.WriteJSON(path, payload); err != nil {
		return "", err
	}
	return path, nil
}
