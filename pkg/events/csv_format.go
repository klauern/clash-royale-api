package events

import (
	"fmt"
	"strings"
)

// EventDeckCSVHeaders returns the canonical CSV headers for event deck exports.
func EventDeckCSVHeaders() []string {
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

// EventDeckCSVRow formats an event deck as a canonical CSV row.
func EventDeckCSVRow(deck EventDeck) []string {
	cardNames := make([]string, len(deck.Deck.Cards))
	for i, card := range deck.Deck.Cards {
		cardNames[i] = formatDeckCard(card)
	}

	endTime := ""
	if deck.EndTime != nil {
		endTime = deck.EndTime.Format("2006-01-02 15:04:05")
	}

	maxWins := ""
	if deck.Performance.MaxWins != nil {
		maxWins = fmt.Sprintf("%d", *deck.Performance.MaxWins)
	}

	return []string{
		deck.EventID,
		deck.PlayerTag,
		deck.EventName,
		string(deck.EventType),
		deck.StartTime.Format("2006-01-02 15:04:05"),
		endTime,
		strings.Join(cardNames, " | "),
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
}

// EventTypeSeparatorCSVRow returns a row marker for grouped event exports.
func EventTypeSeparatorCSVRow(eventType EventType) []string {
	row := make([]string, len(EventDeckCSVHeaders()))
	row[0] = fmt.Sprintf("# Event Type: %s", eventType)
	return row
}

func formatDeckCard(card CardInDeck) string {
	if card.EvolutionLevel > 0 {
		return fmt.Sprintf("%s (Lv.%d Evo.%d)", card.Name, card.Level, card.EvolutionLevel)
	}

	return fmt.Sprintf("%s (Lv.%d)", card.Name, card.Level)
}
