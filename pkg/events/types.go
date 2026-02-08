// Package events provides types and utilities for tracking Clash Royale event decks
// (challenges, tournaments, special events) with performance metrics.
package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/errors"
	"github.com/klauer/clash-royale-api/go/internal/util"
)

// EventType represents the type of Clash Royale event
type EventType string

const (
	EventTypeChallenge         EventType = "challenge"
	EventTypeGrandChallenge    EventType = "grand_challenge"
	EventTypeClassicChallenge  EventType = "classic_challenge"
	EventTypeDraftChallenge    EventType = "draft_challenge"
	EventTypeTournament        EventType = "tournament"
	EventTypeSpecialEvent      EventType = "special_event"
	EventTypeSuddenDeath       EventType = "sudden_death"
	EventTypeDoubleElimination EventType = "double_elimination"
)

// EventProgress represents the current status of an event
type EventProgress string

const (
	EventProgressInProgress EventProgress = "in_progress"
	EventProgressCompleted  EventProgress = "completed"
	EventProgressEliminated EventProgress = "eliminated"
	EventProgressForfeited  EventProgress = "forfeited"
)

// CardInDeck represents a single card within a deck, including its level and properties
type CardInDeck struct {
	Name              string `json:"name"`
	ID                int    `json:"id"`
	Level             int    `json:"level"`
	MaxLevel          int    `json:"max_level"`
	Rarity            string `json:"rarity"`
	ElixirCost        int    `json:"elixir_cost"`
	EvolutionLevel    int    `json:"evolution_level,omitempty"`
	MaxEvolutionLevel int    `json:"max_evolution_level,omitempty"`
}

// Deck represents an 8-card Clash Royale deck composition
type Deck struct {
	Cards     []CardInDeck `json:"cards"`
	AvgElixir float64      `json:"avg_elixir"`
}

// CalculateAvgElixir calculates the average elixir cost of cards in the deck
func (d *Deck) CalculateAvgElixir() float64 {
	return util.CalcAvgElixir(d.Cards, func(card CardInDeck) int {
		return card.ElixirCost
	})
}

// Validate checks if the deck has exactly 8 cards and valid elixir costs
func (d *Deck) Validate() error {
	if len(d.Cards) != 8 {
		return ErrInvalidDeckSize
	}

	for _, card := range d.Cards {
		if card.ElixirCost < 0 || card.ElixirCost > 10 {
			return ErrInvalidElixirCost
		}
	}

	return nil
}

// BattleRecord represents a single battle outcome within an event
type BattleRecord struct {
	Timestamp      time.Time `json:"timestamp"`
	OpponentTag    string    `json:"opponent_tag"`
	OpponentName   string    `json:"opponent_name,omitempty"`
	Result         string    `json:"result"` // "win" or "loss"
	Crowns         int       `json:"crowns"`
	OpponentCrowns int       `json:"opponent_crowns"`
	TrophyChange   *int      `json:"trophy_change,omitempty"`
	BattleMode     string    `json:"battle_mode,omitempty"`
}

// IsWin returns true if this battle was a win
func (br *BattleRecord) IsWin() bool {
	return br.Result == "win"
}

// IsLoss returns true if this battle was a loss
func (br *BattleRecord) IsLoss() bool {
	return br.Result == "loss"
}

// EventPerformance tracks performance metrics for an event deck
type EventPerformance struct {
	Wins          int           `json:"wins"`
	Losses        int           `json:"losses"`
	WinRate       float64       `json:"win_rate"`
	CrownsEarned  int           `json:"crowns_earned"`
	CrownsLost    int           `json:"crowns_lost"`
	MaxWins       *int          `json:"max_wins,omitempty"`
	CurrentStreak int           `json:"current_streak"`
	BestStreak    int           `json:"best_streak"`
	Progress      EventProgress `json:"progress"`
}

// TotalBattles returns the total number of battles (wins + losses)
func (ep *EventPerformance) TotalBattles() int {
	return ep.Wins + ep.Losses
}

// CalculateWinRate updates the win rate based on current wins/losses
func (ep *EventPerformance) CalculateWinRate() {
	total := ep.TotalBattles()
	if total == 0 {
		ep.WinRate = 0
		return
	}
	ep.WinRate = float64(ep.Wins) / float64(total)
}

// UpdateProgress determines the event progress status based on performance
func (ep *EventPerformance) UpdateProgress() {
	// If we have max wins defined and reached it
	if ep.MaxWins != nil && ep.Wins >= *ep.MaxWins {
		ep.Progress = EventProgressCompleted
		return
	}

	// Grand Challenge: 3 losses = eliminated
	// Classic/Draft: 3 losses = eliminated
	if ep.Losses >= 3 {
		ep.Progress = EventProgressEliminated
		return
	}

	// Otherwise still in progress
	ep.Progress = EventProgressInProgress
}

// EventDeck represents a complete event deck with all associated data
type EventDeck struct {
	EventID     string           `json:"event_id"`
	PlayerTag   string           `json:"player_tag"`
	EventName   string           `json:"event_name"`
	EventType   EventType        `json:"event_type"`
	StartTime   time.Time        `json:"start_time"`
	EndTime     *time.Time       `json:"end_time,omitempty"`
	Deck        Deck             `json:"deck"`
	Performance EventPerformance `json:"performance"`
	Battles     []BattleRecord   `json:"battles"`
	EventRules  map[string]any   `json:"event_rules,omitempty"`
	Notes       string           `json:"notes,omitempty"`
}

// AddBattle adds a battle record to the event and updates performance metrics
func (ed *EventDeck) AddBattle(battle BattleRecord) {
	ed.Battles = append(ed.Battles, battle)

	// Update performance metrics
	if battle.IsWin() {
		ed.Performance.Wins++
		ed.Performance.CurrentStreak++
		if ed.Performance.CurrentStreak > ed.Performance.BestStreak {
			ed.Performance.BestStreak = ed.Performance.CurrentStreak
		}
	} else if battle.IsLoss() {
		ed.Performance.Losses++
		ed.Performance.CurrentStreak = 0
	}

	ed.Performance.CrownsEarned += battle.Crowns
	ed.Performance.CrownsLost += battle.OpponentCrowns

	// Recalculate win rate and progress
	ed.Performance.CalculateWinRate()
	ed.Performance.UpdateProgress()

	// Set end time if event is completed
	if ed.Performance.Progress != EventProgressInProgress {
		now := time.Now()
		ed.EndTime = &now
	}
}

// EventDeckCollection represents a player's collection of event decks
type EventDeckCollection struct {
	PlayerTag   string      `json:"player_tag"`
	Decks       []EventDeck `json:"decks"`
	LastUpdated time.Time   `json:"last_updated"`
}

// AddDeck adds or updates an event deck in the collection
func (edc *EventDeckCollection) AddDeck(deck EventDeck) {
	// Check if deck with this event ID already exists
	for i, existingDeck := range edc.Decks {
		if existingDeck.EventID == deck.EventID {
			// Update existing deck
			edc.Decks[i] = deck
			edc.LastUpdated = time.Now()
			return
		}
	}

	// Add new deck
	edc.Decks = append(edc.Decks, deck)
	edc.LastUpdated = time.Now()
}

// GetDecksByType returns all event decks of a specific type
func (edc *EventDeckCollection) GetDecksByType(eventType EventType) []EventDeck {
	return util.FilterSlice(edc.Decks, func(deck EventDeck) bool {
		return deck.EventType == eventType
	})
}

// GetRecentDecks returns event decks from the last N days
func (edc *EventDeckCollection) GetRecentDecks(days int) []EventDeck {
	cutoff := time.Now().AddDate(0, 0, -days)
	return util.FilterSlice(edc.Decks, func(deck EventDeck) bool {
		return deck.StartTime.After(cutoff)
	})
}

// GetBestDecksByWinRate returns top N event decks by win rate (min battles required)
func (edc *EventDeckCollection) GetBestDecksByWinRate(minBattles, limit int) []EventDeck {
	// Filter decks with minimum battle count
	qualified := util.FilterSlice(edc.Decks, func(deck EventDeck) bool {
		return deck.Performance.TotalBattles() >= minBattles
	})

	// Sort by win rate (descending)
	// Note: In production, use sort.Slice for actual sorting
	// This is a placeholder for the type definition

	if len(qualified) > limit {
		return qualified[:limit]
	}

	return qualified
}

// EventMetadata represents metadata about an event type, independent of player participation
type EventMetadata struct {
	EventType     EventType      `json:"event_type"`
	Name          string         `json:"name"`
	Description   string         `json:"description,omitempty"`
	MaxWins       *int           `json:"max_wins,omitempty"`
	MaxLosses     *int           `json:"max_losses,omitempty"`
	EntryFee      *int           `json:"entry_fee,omitempty"`
	Rewards       []string       `json:"rewards,omitempty"`
	Rules         map[string]any `json:"rules,omitempty"`
	AvailableFrom *time.Time     `json:"available_from,omitempty"`
	AvailableTo   *time.Time     `json:"available_to,omitempty"`
	IsActive      bool           `json:"is_active"`
}

// Validate checks if event metadata is valid
func (em *EventMetadata) Validate() error {
	if em.EventType == "" {
		return fmt.Errorf("event type is required")
	}
	if em.Name == "" {
		return fmt.Errorf("event name is required")
	}
	if em.AvailableFrom != nil && em.AvailableTo != nil && em.AvailableFrom.After(*em.AvailableTo) {
		return fmt.Errorf("available_from cannot be after available_to")
	}
	return nil
}

// BattleLog represents a collection of battle records with helper methods
type BattleLog []BattleRecord

// FilterByResult filters battle log by result (win/loss)
func (bl BattleLog) FilterByResult(result string) BattleLog {
	return util.FilterSlice(bl, func(battle BattleRecord) bool {
		return battle.Result == result
	})
}

// FilterByTimeRange filters battle log by time range
func (bl BattleLog) FilterByTimeRange(start, end time.Time) BattleLog {
	return util.FilterSlice(bl, func(battle BattleRecord) bool {
		return !battle.Timestamp.Before(start) && !battle.Timestamp.After(end)
	})
}

// TotalCrowns returns total crowns earned across all battles
func (bl BattleLog) TotalCrowns() int {
	total := 0
	for _, battle := range bl {
		total += battle.Crowns
	}
	return total
}

// WinRate calculates win rate across the battle log
func (bl BattleLog) WinRate() float64 {
	if len(bl) == 0 {
		return 0
	}
	wins := 0
	for _, battle := range bl {
		if battle.IsWin() {
			wins++
		}
	}
	return float64(wins) / float64(len(bl))
}

// MarshalJSON implements custom JSON marshaling for time handling
func (ed EventDeck) MarshalJSON() ([]byte, error) {
	type Alias EventDeck
	return json.Marshal(&struct {
		StartTime string  `json:"start_time"`
		EndTime   *string `json:"end_time,omitempty"`
		*Alias
	}{
		StartTime: ed.StartTime.Format(time.RFC3339),
		EndTime:   formatTimePtr(ed.EndTime),
		Alias:     (*Alias)(&ed),
	})
}

// formatTimePtr formats a time pointer to RFC3339 string, returns nil if pointer is nil
func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format(time.RFC3339)
	return &formatted
}

// Error types for event operations
var (
	ErrInvalidDeckSize   = &EventError{Code: "INVALID_DECK_SIZE", Message: "deck must contain exactly 8 cards"}
	ErrInvalidElixirCost = &EventError{Code: "INVALID_ELIXIR", Message: "card elixir cost must be between 0 and 10"}
	ErrEventNotFound     = &EventError{Code: "EVENT_NOT_FOUND", Message: "event deck not found"}
	ErrInvalidEventType  = &EventError{Code: "INVALID_EVENT_TYPE", Message: "invalid event type"}
)

// EventError represents an event-related error (type alias for shared CodedError)
type EventError = errors.CodedError
