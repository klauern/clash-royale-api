package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
	"github.com/klauer/clash-royale-api/go/pkg/recommend"
)

func TestSanitizePathComponent(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "../deck", want: "deck"},
		{in: "  Hog Rider / Cycle  ", want: "Hog_Rider___Cycle"},
		{in: "", want: "deck"},
	}

	for _, tt := range tests {
		got := sanitizePathComponent(tt.in)
		if got != tt.want {
			t.Fatalf("sanitizePathComponent(%q)=%q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSaveMulliganGuideSanitizesDeckName(t *testing.T) {
	baseDir := t.TempDir()
	guide := &mulligan.MulliganGuide{
		DeckName:    "../escape",
		DeckCards:   []string{"Hog Rider"},
		Archetype:   mulligan.ArchetypeCycle,
		GeneratedAt: time.Date(2026, 2, 14, 10, 0, 0, 0, time.UTC),
	}

	if err := saveMulliganGuide(baseDir, guide); err != nil {
		t.Fatalf("saveMulliganGuide returned error: %v", err)
	}

	matches, err := filepath.Glob(filepath.Join(baseDir, "mulligan", "*.json"))
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 output file, found %d", len(matches))
	}

	base := filepath.Base(matches[0])
	if strings.Contains(base, "..") || strings.Contains(base, "/") || strings.Contains(base, "\\") {
		t.Fatalf("unsafe output filename produced: %q", base)
	}
	if !strings.HasPrefix(base, "escape_") {
		t.Fatalf("expected sanitized filename prefix 'escape_', got %q", base)
	}

	if _, err := os.Stat(filepath.Join(baseDir, "escape_20260214_100000.json")); !os.IsNotExist(err) {
		t.Fatalf("unexpected file found outside mulligan directory")
	}
}

func TestSaveRecommendationsCSVEscapesSpecialCharacters(t *testing.T) {
	dataDir := t.TempDir()
	result := &recommend.RecommendationResult{
		PlayerTag: "PTEST",
		Recommendations: []*recommend.DeckRecommendation{
			{
				Deck: &deck.DeckRecommendation{
					Deck:           []string{"Hog Rider", "Fireball, Deluxe"},
					EvolutionSlots: []string{"Evo: Archers"},
					AvgElixir:      3.50,
				},
				ArchetypeName:      `Bridge Spam, "Control"`,
				Type:               recommend.TypeCustomVariation,
				CompatibilityScore: 91.2,
				SynergyScore:       89.7,
				OverallScore:       90.1,
				Reasons:            []string{`Use "bait", then punish`, "High pressure, low risk"},
			},
		},
	}

	if err := exportRecommendationsToCSV(dataDir, result); err != nil {
		t.Fatalf("exportRecommendationsToCSV returned error: %v", err)
	}

	f, err := os.Open(getRecommendationsCSVPath(dataDir, result.PlayerTag))
	if err != nil {
		t.Fatalf("failed to open CSV output: %v", err)
	}
	defer closeFile(f)

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("failed to parse generated CSV: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row, got %d rows", len(records))
	}
	if got := len(records[1]); got != 10 {
		t.Fatalf("expected 10 CSV columns, got %d", got)
	}
	if records[1][1] != `Bridge Spam, "Control"` {
		t.Fatalf("unexpected archetype cell: %q", records[1][1])
	}
	if records[1][9] != `Use "bait", then punish; High pressure, low risk` {
		t.Fatalf("unexpected reasons cell: %q", records[1][9])
	}
}

func TestDeriveLeagueLabel(t *testing.T) {
	t.Run("uses league name when present", func(t *testing.T) {
		p := &clashroyale.Player{
			League:       clashroyale.League{Name: "Challenger I"},
			BestTrophies: 7000,
		}
		if got := deriveLeagueLabel(p); got != "Challenger I" {
			t.Errorf("got %q, want %q", got, "Challenger I")
		}
	})

	t.Run("falls back to best-trophies bracket when league empty", func(t *testing.T) {
		// Reproduces the post-season-reset shape from R8QGUQRCV.
		p := &clashroyale.Player{
			League:       clashroyale.League{Name: ""},
			BestTrophies: 6500,
		}
		got := deriveLeagueLabel(p)
		if !strings.Contains(got, "6500") || !strings.Contains(got, "Path of Legends") {
			t.Errorf("got %q, want best-trophies bracket label", got)
		}
	})

	t.Run("returns empty for nil player", func(t *testing.T) {
		if got := deriveLeagueLabel(nil); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

func TestBestTrophiesBracketLabel(t *testing.T) {
	cases := []struct {
		trophies int
		contains string
	}{
		{0, ""},
		{2000, "pre-Path of Legends"},
		{5000, "lower"},
		{7000, "mid"},
		{9500, "upper"},
		{11000, "Ultimate Champion"},
	}
	for _, tc := range cases {
		got := bestTrophiesBracketLabel(tc.trophies)
		if tc.contains == "" {
			if got != "" {
				t.Errorf("trophies=%d: got %q, want empty", tc.trophies, got)
			}
			continue
		}
		if !strings.Contains(got, tc.contains) {
			t.Errorf("trophies=%d: got %q, want substring %q", tc.trophies, got, tc.contains)
		}
	}
}
