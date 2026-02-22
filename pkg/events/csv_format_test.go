package events

import (
	"strings"
	"testing"
	"time"
)

func TestEventDeckCSVRow_FormatsDeckCardsAndMaxWins(t *testing.T) {
	now := time.Date(2026, 2, 20, 12, 34, 56, 0, time.UTC)
	end := now.Add(30 * time.Minute)
	maxWins := 12

	deck := EventDeck{
		EventID:   "event-123",
		PlayerTag: "#TEST",
		EventName: "Grand Challenge",
		EventType: EventTypeGrandChallenge,
		StartTime: now,
		EndTime:   &end,
		Deck: Deck{
			Cards: []CardInDeck{
				{Name: "Knight", Level: 11},
				{Name: "Archers", Level: 10, EvolutionLevel: 2},
			},
			AvgElixir: 3.4,
		},
		Performance: EventPerformance{
			Wins:          8,
			Losses:        3,
			WinRate:       8.0 / 11.0,
			CurrentStreak: 2,
			BestStreak:    4,
			CrownsEarned:  20,
			CrownsLost:    11,
			Progress:      EventProgressInProgress,
			MaxWins:       &maxWins,
		},
		Notes: "note",
	}

	row := EventDeckCSVRow(deck)
	headers := EventDeckCSVHeaders()
	if len(row) != len(headers) {
		t.Fatalf("row has %d columns, want %d", len(row), len(headers))
	}

	deckCards := row[6]
	if !strings.Contains(deckCards, "Knight (Lv.11)") {
		t.Fatalf("deck cards missing standard level formatting: %q", deckCards)
	}
	if !strings.Contains(deckCards, "Archers (Lv.10 Evo.2)") {
		t.Fatalf("deck cards missing evolution formatting: %q", deckCards)
	}

	if row[17] != "12" {
		t.Fatalf("max wins = %q, want 12", row[17])
	}
}
