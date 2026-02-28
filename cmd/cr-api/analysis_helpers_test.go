package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestLoadOfflineAnalysisFromFlags_UsesAnalysisFileOverride(t *testing.T) {
	t.Parallel()

	analysisPath := writeDeckAnalysisFixture(
		t,
		t.TempDir(),
		"analysis_file.json",
		deck.CardAnalysis{
			CardLevels: map[string]deck.CardLevelData{
				"Hog Rider": {Level: 14, MaxLevel: 14},
			},
			PlayerTag: "#FROM_FILE",
		},
		time.Now(),
	)

	builder := deck.NewBuilder(t.TempDir())
	result, err := loadOfflineAnalysisFromFlags(builder, "#FALLBACK", "", "", analysisPath, false)
	if err != nil {
		t.Fatalf("loadOfflineAnalysisFromFlags() error = %v", err)
	}

	if result.PlayerName != "#FALLBACK" {
		t.Fatalf("PlayerName = %q, want fallback %q", result.PlayerName, "#FALLBACK")
	}
	if result.PlayerTag != "#FROM_FILE" {
		t.Fatalf("PlayerTag = %q, want %q", result.PlayerTag, "#FROM_FILE")
	}
}

func TestLoadOfflineAnalysisFromFlags_UsesLatestFromDefaultAnalysisDir(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	analysisDir := filepath.Join(dataDir, "analysis")
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	playerTag := "#P2Y8Q"
	oldTime := time.Now().Add(-time.Hour)
	newTime := time.Now()

	writeDeckAnalysisFixture(
		t,
		analysisDir,
		"20260220_analysis_P2Y8Q.json",
		deck.CardAnalysis{PlayerName: "Older", PlayerTag: playerTag, CardLevels: map[string]deck.CardLevelData{}},
		oldTime,
	)
	writeDeckAnalysisFixture(
		t,
		analysisDir,
		"20260221_analysis_P2Y8Q.json",
		deck.CardAnalysis{PlayerName: "Newest", PlayerTag: playerTag, CardLevels: map[string]deck.CardLevelData{}},
		newTime,
	)

	builder := deck.NewBuilder(dataDir)
	result, err := loadOfflineAnalysisFromFlags(builder, playerTag, dataDir, "", "", false)
	if err != nil {
		t.Fatalf("loadOfflineAnalysisFromFlags() error = %v", err)
	}
	if result.PlayerName != "Newest" {
		t.Fatalf("PlayerName = %q, want %q", result.PlayerName, "Newest")
	}
}

func TestLoadOfflineAnalysisFromFlags_TrimmedTagWorksForLatestLookup(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	analysisDir := filepath.Join(dataDir, "analysis")
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	writeDeckAnalysisFixture(
		t,
		analysisDir,
		"20260221_analysis_P2Y8Q.json",
		deck.CardAnalysis{PlayerName: "Trimmed", PlayerTag: "#P2Y8Q", CardLevels: map[string]deck.CardLevelData{}},
		time.Now(),
	)

	builder := deck.NewBuilder(dataDir)
	result, err := loadOfflineAnalysisFromFlags(builder, "  #P2Y8Q  ", dataDir, "", "", false)
	if err != nil {
		t.Fatalf("loadOfflineAnalysisFromFlags() error = %v", err)
	}
	if result.PlayerName != "Trimmed" {
		t.Fatalf("PlayerName = %q, want %q", result.PlayerName, "Trimmed")
	}
}

func TestLoadOfflineAnalysisFromFlags_RequiresTagForAnalysisDirMode(t *testing.T) {
	t.Parallel()

	builder := deck.NewBuilder(t.TempDir())
	_, err := loadOfflineAnalysisFromFlags(builder, "", "", "analysis", "", false)
	if err == nil {
		t.Fatal("expected error when tag is missing for analysis-dir mode")
	}
	if !strings.Contains(err.Error(), "--tag is required") {
		t.Fatalf("error = %q, want message containing --tag is required", err.Error())
	}
}

func writeDeckAnalysisFixture(
	t *testing.T,
	dir string,
	fileName string,
	analysis deck.CardAnalysis,
	modTime time.Time,
) string {
	t.Helper()

	path := filepath.Join(dir, fileName)
	data, err := json.Marshal(analysis)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("Chtimes failed: %v", err)
	}

	return path
}
