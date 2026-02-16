package csv

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/events"
)

func TestEventBattlesHeaders_IncludeMatchupColumns(t *testing.T) {
	headers := eventBattlesHeaders()
	expected := []string{
		"Player Deck Hash",
		"Opponent Deck Hash",
		"Player Deck",
		"Opponent Deck",
	}

	for _, col := range expected {
		found := slices.Contains(headers, col)
		if !found {
			t.Errorf("expected header %q to be present", col)
		}
	}
}

func TestEventBattlesExport_IncludesMatchupData(t *testing.T) {
	tempDir := t.TempDir()
	now := time.Now()

	collection := &events.EventDeckCollection{
		PlayerTag: "#TEST123",
		Decks: []events.EventDeck{
			{
				EventID:   "event-1",
				PlayerTag: "#TEST123",
				Battles: []events.BattleRecord{
					{
						Timestamp:        now,
						OpponentTag:      "#OPP",
						OpponentName:     "Opponent",
						Result:           "win",
						Crowns:           3,
						OpponentCrowns:   1,
						BattleMode:       "Grand Challenge",
						PlayerDeckHash:   "player-hash",
						OpponentDeckHash: "opp-hash",
						PlayerDeck:       []string{"Knight", "Archers"},
						OpponentDeck:     []string{"Golem", "Baby Dragon"},
					},
				},
			},
		},
	}

	exporter := NewEventBattlesExporter()
	if err := exporter.Export(tempDir, collection); err != nil {
		t.Fatalf("export failed: %v", err)
	}

	filePath := filepath.Join(tempDir, "csv", "events", "event_battles.csv")
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open exported file: %v", err)
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("failed to read exported csv: %v", err)
	}

	if len(records) < 2 {
		t.Fatalf("expected at least header + one row, got %d", len(records))
	}

	header := records[0]
	row := records[1]
	indexOf := func(name string) int {
		for i, h := range header {
			if h == name {
				return i
			}
		}
		return -1
	}

	playerHashIdx := indexOf("Player Deck Hash")
	opponentHashIdx := indexOf("Opponent Deck Hash")
	playerDeckIdx := indexOf("Player Deck")
	opponentDeckIdx := indexOf("Opponent Deck")
	if playerHashIdx < 0 || opponentHashIdx < 0 || playerDeckIdx < 0 || opponentDeckIdx < 0 {
		t.Fatalf("expected matchup columns missing from headers: %v", header)
	}

	if row[playerHashIdx] != "player-hash" {
		t.Errorf("player hash = %q, want %q", row[playerHashIdx], "player-hash")
	}
	if row[opponentHashIdx] != "opp-hash" {
		t.Errorf("opponent hash = %q, want %q", row[opponentHashIdx], "opp-hash")
	}
	if row[playerDeckIdx] == "" {
		t.Error("player deck column should not be empty")
	}
	if row[opponentDeckIdx] == "" {
		t.Error("opponent deck column should not be empty")
	}
}
