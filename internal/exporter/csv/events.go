package csv

import (
	"fmt"

	"github.com/klauer/clash-royale-api/go/internal/storage"
	"github.com/klauer/clash-royale-api/go/pkg/events"
)

// NewEventDeckExporter creates a new event deck CSV exporter
func NewEventDeckExporter() *CSVExporter {
	return NewCSVExporter(
		"event_decks.csv",
		eventDeckHeaders,
		eventDeckExport,
	)
}

// eventDeckHeaders returns the CSV headers for event deck data
func eventDeckHeaders() []string {
	return events.EventDeckCSVHeaders()
}

// eventDeckExport exports event deck data to CSV
func eventDeckExport(dataDir string, data any) error {
	collection, ok := data.(*events.EventDeckCollection)
	if !ok {
		return fmt.Errorf("expected EventDeckCollection type, got %T", data)
	}

	// Prepare CSV rows
	var rows [][]string

	for _, deck := range collection.Decks {
		rows = append(rows, events.EventDeckCSVRow(deck))
	}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "event_decks.csv"}
	filePath := exporter.csvFilePath(dataDir, storage.CSVEventsSubdir)
	return exporter.writeCSV(filePath, eventDeckHeaders(), rows)
}

// NewEventBattlesExporter creates a new event battles CSV exporter
func NewEventBattlesExporter() *CSVExporter {
	return NewCSVExporter(
		"event_battles.csv",
		eventBattlesHeaders,
		eventBattlesExport,
	)
}

// eventBattlesHeaders returns the CSV headers for event battle data
func eventBattlesHeaders() []string {
	return []string{
		"Event ID",
		"Player Tag",
		"Battle Timestamp",
		"Opponent Tag",
		"Opponent Name",
		"Result",
		"Player Crowns",
		"Opponent Crowns",
		"Trophy Change",
		"Battle Mode",
		"Player Deck Hash",
		"Opponent Deck Hash",
		"Player Deck",
		"Opponent Deck",
	}
}

// eventBattlesExport exports event battle data to CSV
func eventBattlesExport(dataDir string, data any) error {
	collection, ok := data.(*events.EventDeckCollection)
	if !ok {
		return fmt.Errorf("expected EventDeckCollection type, got %T", data)
	}

	// Prepare CSV rows
	var rows [][]string

	for _, deck := range collection.Decks {
		for _, battle := range deck.Battles {
			// Format trophy change
			trophyChange := ""
			if battle.TrophyChange != nil {
				trophyChange = fmt.Sprintf("%d", *battle.TrophyChange)
			}

			row := []string{
				deck.EventID,
				deck.PlayerTag,
				battle.Timestamp.Format("2006-01-02 15:04:05"),
				battle.OpponentTag,
				battle.OpponentName,
				battle.Result,
				fmt.Sprintf("%d", battle.Crowns),
				fmt.Sprintf("%d", battle.OpponentCrowns),
				trophyChange,
				battle.BattleMode,
				battle.PlayerDeckHash,
				battle.OpponentDeckHash,
				fmt.Sprintf("%v", battle.PlayerDeck),
				fmt.Sprintf("%v", battle.OpponentDeck),
			}
			rows = append(rows, row)
		}
	}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "event_battles.csv"}
	filePath := exporter.csvFilePath(dataDir, storage.CSVEventsSubdir)
	return exporter.writeCSV(filePath, eventBattlesHeaders(), rows)
}
