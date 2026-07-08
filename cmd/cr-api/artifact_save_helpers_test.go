package main

import (
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/storage"
)

func TestSaveTaggedJSONArtifactTimestamped(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	payload := map[string]string{"status": "ok"}

	path, err := saveTaggedJSONArtifact(dir, "#TAG123", payload, taggedJSONArtifactOptions{
		subdir:      "budget",
		fileStem:    "budget",
		timestamped: true,
	})
	if err != nil {
		t.Fatalf("saveTaggedJSONArtifact() error = %v", err)
	}

	pattern := regexp.MustCompile(`budget/\d{8}_\d{6}_budget_TAG123\.json$`)
	if !pattern.MatchString(filepath.ToSlash(path)) {
		t.Fatalf("saveTaggedJSONArtifact() path = %q, want timestamped budget path", path)
	}

	var got map[string]string
	if err := storage.ReadJSON(path, &got); err != nil {
		t.Fatalf("storage.ReadJSON() error = %v", err)
	}
	if got["status"] != "ok" {
		t.Fatalf("saved payload status = %q, want ok", got["status"])
	}
}

func TestSaveTaggedJSONArtifactStableName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	payload := map[string]string{"mode": "playstyle"}

	path, err := saveTaggedJSONArtifact(dir, "#PLYR99", payload, taggedJSONArtifactOptions{
		subdir:   storage.AnalysisDir,
		fileStem: "playstyle",
	})
	if err != nil {
		t.Fatalf("saveTaggedJSONArtifact() error = %v", err)
	}

	want := filepath.Join(dir, storage.AnalysisDir, "playstyle_PLYR99.json")
	if path != want {
		t.Fatalf("saveTaggedJSONArtifact() path = %q, want %q", path, want)
	}

	var got map[string]string
	if err := storage.ReadJSON(path, &got); err != nil {
		t.Fatalf("storage.ReadJSON() error = %v", err)
	}
	if got["mode"] != "playstyle" {
		t.Fatalf("saved payload mode = %q, want playstyle", got["mode"])
	}
}

func TestSaveTimestampedJSONArtifactUsesProvidedTimestamp(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	payload := map[string]string{"scenario": "demo"}
	path, err := saveTimestampedJSONArtifact(dir, payload, timestampedJSONArtifactOptions{
		subdir:    "whatif",
		fileStem:  "scenario",
		timestamp: time.Date(2026, time.July, 8, 5, 4, 3, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("saveTimestampedJSONArtifact() error = %v", err)
	}

	want := filepath.Join(dir, "whatif", "scenario_20260708_050403.json")
	if path != want {
		t.Fatalf("saveTimestampedJSONArtifact() path = %q, want %q", path, want)
	}

	var got map[string]string
	if err := storage.ReadJSON(path, &got); err != nil {
		t.Fatalf("storage.ReadJSON() error = %v", err)
	}
	if got["scenario"] != "demo" {
		t.Fatalf("saved payload scenario = %q, want demo", got["scenario"])
	}
}
