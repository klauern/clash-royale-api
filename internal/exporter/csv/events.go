package csv

import (
	"fmt"
	"path/filepath"

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
	return []string{
		"Event ID",
		"Player Tag",
		"Event Name",
		"Event Type",
		"Start Time",
		"End Time",
		"Deck Cards",
		"Deck Average Elixir",
		"Total Battles",
		"Wins",
		"Losses",
		"Win Rate",
		"Current Streak",
		"Best Streak",
		"Crowns Earned",
		"Crowns Lost",
		"Event Progress",
		"Max Wins",
		"Notes",
	}
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
		// Format deck cards as a string, including evolution level if present
		cardNames := make([]string, len(deck.Deck.Cards))
		for i, card := range deck.Deck.Cards {
			if card.EvolutionLevel > 0 {
				cardNames[i] = fmt.Sprintf("%s (Lv.%d Evo.%d)", card.Name, card.Level, card.EvolutionLevel)
			} else {
				cardNames[i] = fmt.Sprintf("%s (Lv.%d)", card.Name, card.Level)
			}
		}

		// Format end time
		endTime := ""
		if deck.EndTime != nil {
			endTime = deck.EndTime.Format("2006-01-02 15:04:05")
		}

		// Format max wins
		maxWins := ""
		if deck.Performance.MaxWins != nil {
			maxWins = fmt.Sprintf("%d", *deck.Performance.MaxWins)
		}

		row := []string{
			deck.EventID,
			deck.PlayerTag,
			deck.EventName,
			string(deck.EventType),
			deck.StartTime.Format("2006-01-02 15:04:05"),
			endTime,
			fmt.Sprintf("%v", cardNames),
			fmt.Sprintf("%.1f", deck.Deck.AvgElixir),
			fmt.Sprintf("%d", deck.Performance.TotalBattles()),
			fmt.Sprintf("%d", deck.Performance.Wins),
			fmt.Sprintf("%d", deck.Performance.Losses),
			fmt.Sprintf("%.2f", deck.Performance.WinRate),
			fmt.Sprintf("%d", deck.Performance.CurrentStreak),
			fmt.Sprintf("%d", deck.Performance.BestStreak),
			fmt.Sprintf("%d", deck.Performance.CrownsEarned),
			fmt.Sprintf("%d", deck.Performance.CrownsLost),
			string(deck.Performance.Progress),
			maxWins,
			deck.Notes,
		}
		rows = append(rows, row)
	}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "event_decks.csv"}
	filePath := filepath.Join(dataDir, "csv", "events", exporter.FilenameBase)
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
	filePath := filepath.Join(dataDir, "csv", "events", exporter.FilenameBase)
	return exporter.writeCSV(filePath, eventBattlesHeaders(), rows)
}
