package deck

import (
	"path/filepath"
	"testing"

	"github.com/klauer/clash-royale-api/go/internal/storage"
)

func TestSuiteFilenames(t *testing.T) {
	t.Parallel()

	deckName := SuiteDeckFilename("20260227_010203", "balanced", 2, "#abC123")
	if deckName != "20260227_010203_deck_balanced_var2_abC123.json" {
		t.Fatalf("unexpected deck filename: %s", deckName)
	}

	summaryName := SuiteSummaryFilename("20260227_010203", "#abC123")
	if summaryName != "20260227_010203_deck_suite_summary_abC123.json" {
		t.Fatalf("unexpected summary filename: %s", summaryName)
	}
}

func TestWriteSuiteDeckAndSummary(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	deckPath := filepath.Join(tempDir, "deck.json")
	summaryPath := filepath.Join(tempDir, "summary.json")

	recommendation := &DeckRecommendation{
		Deck:      []string{"Hog Rider", "Fireball"},
		AvgElixir: 3.5,
	}
	if err := WriteSuiteDeck(deckPath, recommendation); err != nil {
		t.Fatalf("WriteSuiteDeck() error = %v", err)
	}

	var deckPayload SuiteDeckPayload
	if err := storage.ReadJSON(deckPath, &deckPayload); err != nil {
		t.Fatalf("failed to load deck payload: %v", err)
	}
	if len(deckPayload.Deck) != 2 || deckPayload.AvgElixir != 3.5 {
		t.Fatalf("unexpected deck payload: %+v", deckPayload)
	}

	summary := NewSuiteSummary(
		"2026-02-27T00:00:00Z",
		"Player",
		"#TAG123",
		SuiteBuildInfo{
			TotalDecks:     4,
			Successful:     3,
			Failed:         1,
			Strategies:     2,
			Variations:     2,
			GenerationTime: "1s",
		},
		[]SuiteDeckSummary{{
			Strategy:  "balanced",
			Variation: 1,
			Cards:     []string{"Hog Rider"},
			AvgElixir: 3.0,
			FilePath:  "deck.json",
		}},
	)

	if err := WriteSuiteSummary(summaryPath, summary); err != nil {
		t.Fatalf("WriteSuiteSummary() error = %v", err)
	}

	var loadedSummary SuiteSummary
	if err := storage.ReadJSON(summaryPath, &loadedSummary); err != nil {
		t.Fatalf("failed to load summary payload: %v", err)
	}
	if loadedSummary.Version != "1.0.0" {
		t.Fatalf("unexpected summary version: %s", loadedSummary.Version)
	}
	if loadedSummary.Player.Tag != "#TAG123" {
		t.Fatalf("unexpected player tag: %s", loadedSummary.Player.Tag)
	}
	if len(loadedSummary.Decks) != 1 {
		t.Fatalf("unexpected deck count: %d", len(loadedSummary.Decks))
	}
}

func TestWriteSuiteDeckNilInput(t *testing.T) {
	t.Parallel()

	deckPath := filepath.Join(t.TempDir(), "deck.json")
	if err := WriteSuiteDeck(deckPath, nil); err == nil {
		t.Fatal("expected WriteSuiteDeck to reject nil recommendation")
	}
}
