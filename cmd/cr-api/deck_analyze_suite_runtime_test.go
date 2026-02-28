package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestCalculateCompareCount(t *testing.T) {
	tests := []struct {
		name        string
		resultCount int
		topN        int
		wantCount   int
		wantErr     string
	}{
		{
			name:        "requires two results",
			resultCount: 1,
			topN:        5,
			wantErr:     "need at least 2 evaluated decks to compare, got 1",
		},
		{
			name:        "caps to available result count",
			resultCount: 3,
			topN:        5,
			wantCount:   3,
		},
		{
			name:        "minimum compare count",
			resultCount: 4,
			topN:        1,
			wantCount:   2,
		},
		{
			name:        "maximum compare count",
			resultCount: 10,
			topN:        8,
			wantCount:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateCompareCount(tt.resultCount, tt.topN)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantCount {
				t.Fatalf("calculateCompareCount(%d, %d)=%d, want %d", tt.resultCount, tt.topN, got, tt.wantCount)
			}
		})
	}
}

func TestSaveSuiteDeckFile(t *testing.T) {
	t.Run("returns saved path on success", func(t *testing.T) {
		outputDir := t.TempDir()
		recommendation := &deck.DeckRecommendation{
			Deck:      []string{"Knight", "Archers"},
			AvgElixir: 3,
		}

		path, err := saveSuiteDeckFile(outputDir, "balanced", 1, "#TEST123", recommendation)
		if err != nil {
			t.Fatalf("saveSuiteDeckFile() error = %v", err)
		}
		if path == "" {
			t.Fatal("expected non-empty saved path")
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected saved file to exist: %v", err)
		}
	})

	t.Run("does not leak path on write failure", func(t *testing.T) {
		parent := t.TempDir()
		outputPath := filepath.Join(parent, "not-a-directory")
		if err := os.WriteFile(outputPath, []byte("x"), 0o644); err != nil {
			t.Fatalf("write sentinel file: %v", err)
		}

		recommendation := &deck.DeckRecommendation{
			Deck:      []string{"Knight", "Archers"},
			AvgElixir: 3,
		}

		path, err := saveSuiteDeckFile(outputPath, "balanced", 1, "#TEST123", recommendation)
		if err == nil {
			t.Fatal("expected saveSuiteDeckFile() to fail")
		}
		if path != "" {
			t.Fatalf("expected empty path on failure, got %q", path)
		}
	})
}
