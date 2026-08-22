package csv

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

func TestBattleLogExport_StableSchema(t *testing.T) {
	tempDir := t.TempDir()
	battles := []clashroyale.Battle{
		{
			Type:    "PvP",
			UTCDate: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			Team: []clashroyale.BattleTeam{
				{Tag: "#P", Name: "Player", StartingTrophies: 1000, TrophyChange: 30, Crowns: 3},
			},
			Opponent: []clashroyale.BattleTeam{
				{Tag: "#O", Name: "Opponent", StartingTrophies: 1100, TrophyChange: -30, Crowns: 1},
			},
		},
		{
			Type:        "PvP",
			UTCDate:     time.Date(2026, 1, 1, 13, 0, 0, 0, time.UTC),
			DeckAverage: 4,
			Deck: []clashroyale.Card{
				{Name: "Knight", Level: 11},
			},
			Team: []clashroyale.BattleTeam{
				{Tag: "#P", Name: "Player", StartingTrophies: 1030, TrophyChange: -10, Crowns: 1},
			},
			Opponent: []clashroyale.BattleTeam{
				{Tag: "#O", Name: "Opponent", StartingTrophies: 1070, TrophyChange: 10, Crowns: 2},
			},
		},
	}

	if err := NewBattleLogExporter().Export(tempDir, battles); err != nil {
		t.Fatalf("battle log export failed: %v", err)
	}

	records := readCSVFile(t, filepath.Join(tempDir, "csv", "battles", "battle_log.csv"))
	if len(records) != 3 {
		t.Fatalf("row count = %d, want 3", len(records))
	}
	if got := len(records[0]); got != len(records[1]) || got != len(records[2]) {
		t.Fatalf("inconsistent csv row width: header=%d row1=%d row2=%d", len(records[0]), len(records[1]), len(records[2]))
	}
	if records[0][len(records[0])-1] != "Deck Cards" {
		t.Fatalf("last header = %q, want %q", records[0][len(records[0])-1], "Deck Cards")
	}
}

func readCSVFile(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s failed: %v", path, err)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read csv %s failed: %v", path, err)
	}
	return records
}
