package events

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExporter_ExportCSV_GroupByEventWritesSeparatorRows(t *testing.T) {
	tempDir := t.TempDir()
	now := time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC)

	collection := &EventDeckCollection{
		PlayerTag: "#TEST",
		Decks: []EventDeck{
			{
				EventID:   "event-challenge",
				PlayerTag: "#TEST",
				EventName: "Challenge",
				EventType: EventTypeChallenge,
				StartTime: now,
				Deck: Deck{
					Cards: []CardInDeck{{Name: "Knight", Level: 11}},
				},
				Performance: EventPerformance{Wins: 1, Losses: 0},
			},
			{
				EventID:   "event-tournament",
				PlayerTag: "#TEST",
				EventName: "Tournament",
				EventType: EventTypeTournament,
				StartTime: now.Add(time.Hour),
				Deck: Deck{
					Cards: []CardInDeck{{Name: "Archers", Level: 10}},
				},
				Performance: EventPerformance{Wins: 2, Losses: 1},
			},
		},
	}

	exporter := NewExporter(ExportOptions{
		Format:       FormatCSV,
		OutputDir:    tempDir,
		GroupByEvent: true,
	})
	if err := exporter.Export(collection); err != nil {
		t.Fatalf("export failed: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(tempDir, "event_decks_*.csv"))
	if err != nil || len(files) != 1 {
		t.Fatalf("expected one CSV export file, got %d (%v)", len(files), err)
	}

	file, err := os.Open(files[0])
	if err != nil {
		t.Fatalf("failed to open export file: %v", err)
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV records: %v", err)
	}

	separatorRows := 0
	for _, row := range records {
		if len(row) > 0 && strings.HasPrefix(row[0], "# Event Type: ") {
			separatorRows++
		}
	}

	if separatorRows != 2 {
		t.Fatalf("expected 2 separator rows, got %d", separatorRows)
	}
}
