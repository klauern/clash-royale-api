package csv

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/klauer/clash-royale-api/go/internal/storage"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// NewBattleLogExporter creates a new battle log CSV exporter
func NewBattleLogExporter() *CSVExporter {
	return NewCSVExporter(
		"battle_log.csv",
		battleLogHeaders,
		battleLogExport,
	)
}

// battleLogHeaders returns the CSV headers for battle log data
func battleLogHeaders() []string {
	return []string{
		"Timestamp",
		"Battle Type",
		"Player Tag",
		"Player Name",
		"Player Starting Trophies",
		"Player Trophy Change",
		"Player Crowns",
		"Opponent Tag",
		"Opponent Name",
		"Opponent Starting Trophies",
		"Opponent Trophy Change",
		"Opponent Crowns",
		"Is Ladder Tournament",
		"Team Size",
		"Deck Average Elixir",
		"Deck Link",
		"Not Counted",
		"Deck Cards",
	}
}

// battleLogExport exports battle log data to CSV.
func battleLogExport(dataDir string, data any) error {
	battles, ok := data.([]clashroyale.Battle)
	if !ok {
		return csvTypeMismatchError(reflect.TypeOf([]clashroyale.Battle(nil)), data)
	}

	rows := makeBattleLogRows(battles)
	return writeCSVRows(dataDir, storage.CSVBattlesSubdir, "battle_log.csv", battleLogHeaders(), rows)
}

func makeBattleLogRows(battles []clashroyale.Battle) [][]string {
	rows := make([][]string, 0, len(battles))
	for _, battle := range battles {
		row, ok := battleLogRow(battle)
		if !ok {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

func battleLogRow(battle clashroyale.Battle) ([]string, bool) {
	if len(battle.Team) == 0 || len(battle.Opponent) == 0 {
		return nil, false
	}
	player := battle.Team[0]
	opponent := battle.Opponent[0]
	row := []string{
		battle.UTCDate.Format("2006-01-02 15:04:05"),
		battle.Type,
		player.Tag,
		player.Name,
		strconv.Itoa(player.StartingTrophies),
		strconv.Itoa(player.TrophyChange),
		strconv.Itoa(player.Crowns),
		opponent.Tag,
		opponent.Name,
		strconv.Itoa(opponent.StartingTrophies),
		strconv.Itoa(opponent.TrophyChange),
		strconv.Itoa(opponent.Crowns),
		strconv.FormatBool(battle.IsLadderTournament),
		strconv.Itoa(len(battle.Team)),
		formatPositiveIntOrEmpty(battle.DeckAverage),
		battle.GameMode.DeckLink,
		strconv.FormatBool(battle.GameMode.NotCounted),
		formatDeckCards(battle.Deck),
	}
	return row, true
}

func formatDeckCards(cards []clashroyale.Card) string {
	if len(cards) == 0 {
		return ""
	}
	cardNames := make([]string, len(cards))
	for i, card := range cards {
		cardNames[i] = fmt.Sprintf("%s (Lv.%d)", card.Name, card.Level)
	}
	return fmt.Sprintf("%v", cardNames)
}

func formatPositiveIntOrEmpty(value int) string {
	if value <= 0 {
		return ""
	}
	return strconv.Itoa(value)
}
