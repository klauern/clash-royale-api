package events

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// Parser parses battle logs to identify event decks and extract performance data
type Parser struct {
	// Event battle modes that indicate special events
	eventBattleModes map[string]EventType

	// Event name patterns to detect special events
	eventPatterns map[string]string
}

// NewParser creates a new battle log parser
func NewParser() *Parser {
	return &Parser{
		eventBattleModes: map[string]EventType{
			"challenge":          EventTypeChallenge,
			"grand challenge":    EventTypeGrandChallenge,
			"classic challenge":  EventTypeClassicChallenge,
			"draft challenge":    EventTypeDraftChallenge,
			"tournament":         EventTypeTournament,
			"special event":      EventTypeSpecialEvent,
			"sudden death":       EventTypeSuddenDeath,
			"double elimination": EventTypeDoubleElimination,
		},
		eventPatterns: map[string]string{
			"lava":        "Lava Hound Challenge",
			"hog":         "Hog Rider Challenge",
			"mortar":      "Mortar Challenge",
			"graveyard":   "Graveyard Challenge",
			"ram rage":    "Ram Rage Challenge",
			"sparky":      "Sparky Challenge",
			"electro":     "Electro Challenge",
			"skeleton":    "Skeleton Army Challenge",
			"bandit":      "Bandit Challenge",
			"night witch": "Night Witch Challenge",
			"royale":      "Clash Royale Championship",
			"worlds":      "World Championship",
			"ccgs":        "Clash Championship Series",
		},
	}
}

// ParseBattleLogs parses battle logs and extracts event decks
func (p *Parser) ParseBattleLogs(battleLogs []clashroyale.Battle, playerTag string) ([]EventDeck, error) {
	// Group battles by potential events
	eventGroups := p.groupBattlesByEvent(battleLogs, playerTag)

	// Convert groups to EventDeck objects
	eventDecks := make([]EventDeck, 0, len(eventGroups))
	for _, group := range eventGroups {
		eventDeck, err := p.createEventDeck(group, playerTag)
		if err != nil {
			// Log error but continue processing other events
			continue
		}
		if eventDeck != nil {
			eventDecks = append(eventDecks, *eventDeck)
		}
	}

	return eventDecks, nil
}

// eventGroup represents a group of battles that belong to the same event
type eventGroup struct {
	eventType  EventType
	eventName  string
	battleMode string
	deckCards  []string
	battles    []clashroyale.Battle
	startTime  time.Time
}

// groupBattlesByEvent groups battles into events based on timing and mode
func (p *Parser) groupBattlesByEvent(battleLogs []clashroyale.Battle, playerTag string) []eventGroup {
	// Filter and sort event battles by time
	eventBattles := make([]clashroyale.Battle, 0)
	for _, battle := range battleLogs {
		if p.isEventBattle(battle) {
			eventBattles = append(eventBattles, battle)
		}
	}

	// Sort by battle time (oldest first)
	sort.Slice(eventBattles, func(i, j int) bool {
		return eventBattles[i].UTCDate.Before(eventBattles[j].UTCDate)
	})

	if len(eventBattles) == 0 {
		return []eventGroup{}
	}

	// Group battles into events
	groups := make([]eventGroup, 0)
	var currentGroup *eventGroup

	for _, battle := range eventBattles {
		if p.isNewEvent(battle, currentGroup) {
			// Save previous group if exists
			if currentGroup != nil {
				groups = append(groups, *currentGroup)
			}

			// Start new group
			eventData := p.extractEventData(battle)
			currentGroup = &eventGroup{
				eventType:  eventData.eventType,
				eventName:  eventData.eventName,
				battleMode: eventData.battleMode,
				deckCards:  eventData.deckCards,
				battles:    []clashroyale.Battle{battle},
				startTime:  battle.UTCDate,
			}
		} else {
			// Continue current group
			currentGroup.battles = append(currentGroup.battles, battle)
		}
	}

	// Save last group
	if currentGroup != nil {
		groups = append(groups, *currentGroup)
	}

	return groups
}

// isEventBattle checks if a battle is part of an event
func (p *Parser) isEventBattle(battle clashroyale.Battle) bool {
	// Check battle mode for event keywords
	modeLower := strings.ToLower(battle.GameMode.Name)
	for mode := range p.eventBattleModes {
		if strings.Contains(modeLower, mode) {
			return true
		}
	}

	// Check if it's a ladder tournament (regular ladder battles have this as false)
	// IsLadderTournament being true means it's NOT a regular ladder match
	if battle.IsLadderTournament {
		return true
	}

	return false
}

// eventData represents extracted event information
type eventData struct {
	eventType  EventType
	eventName  string
	battleMode string
	deckCards  []string
}

// extractEventData extracts event information from a battle
func (p *Parser) extractEventData(battle clashroyale.Battle) eventData {
	modeName := battle.GameMode.Name
	modeLower := strings.ToLower(modeName)

	// Determine event type - check more specific patterns first
	eventType := EventTypeChallenge // Default
	// Order matters: check more specific patterns before generic ones
	specificPatterns := []struct {
		pattern   string
		eventType EventType
	}{
		{"grand challenge", EventTypeGrandChallenge},
		{"classic challenge", EventTypeClassicChallenge},
		{"draft challenge", EventTypeDraftChallenge},
		{"double elimination", EventTypeDoubleElimination},
		{"sudden death", EventTypeSuddenDeath},
		{"special event", EventTypeSpecialEvent},
		{"tournament", EventTypeTournament},
		{"challenge", EventTypeChallenge}, // Generic - must be last
	}

	for _, sp := range specificPatterns {
		if strings.Contains(modeLower, sp.pattern) {
			eventType = sp.eventType
			break
		}
	}

	// Try to detect specific event name
	eventName := modeName
	for pattern, name := range p.eventPatterns {
		if strings.Contains(modeLower, pattern) {
			eventName = name
			break
		}
	}

	// Extract deck cards
	deckCards := p.extractDeckFromBattle(battle)

	return eventData{
		eventType:  eventType,
		eventName:  eventName,
		battleMode: modeName,
		deckCards:  deckCards,
	}
}

// extractDeckFromBattle extracts card names from battle team data
func (p *Parser) extractDeckFromBattle(battle clashroyale.Battle) []string {
	if len(battle.Team) == 0 {
		return []string{}
	}

	// Get first player's cards (should be the player we're analyzing)
	cards := battle.Team[0].Cards

	cardNames := make([]string, 0, len(cards))
	for _, card := range cards {
		cardNames = append(cardNames, card.Name)
	}

	// Sort for consistent comparison
	sort.Strings(cardNames)

	return cardNames
}

// isNewEvent determines if this battle starts a new event
func (p *Parser) isNewEvent(battle clashroyale.Battle, currentGroup *eventGroup) bool {
	if currentGroup == nil {
		return true
	}

	// Check time gap from last battle (events typically don't have large gaps)
	if len(currentGroup.battles) > 0 {
		lastBattle := currentGroup.battles[len(currentGroup.battles)-1]
		if battle.UTCDate.Sub(lastBattle.UTCDate) > 1*time.Hour {
			return true
		}
	}

	// Check if event type changed
	newEventData := p.extractEventData(battle)
	if newEventData.eventType != currentGroup.eventType {
		return true
	}

	// Check deck composition (significant change might indicate new event)
	if !p.isSameDeck(currentGroup.deckCards, newEventData.deckCards) {
		return true
	}

	return false
}

// isSameDeck checks if two decks have the same cards (already sorted)
func (p *Parser) isSameDeck(deck1, deck2 []string) bool {
	if len(deck1) != len(deck2) {
		return false
	}

	for i := range deck1 {
		if deck1[i] != deck2[i] {
			return false
		}
	}

	return true
}

// createEventDeck creates an EventDeck object from grouped battles
func (p *Parser) createEventDeck(group eventGroup, playerTag string) (*EventDeck, error) {
	if len(group.battles) == 0 {
		return nil, fmt.Errorf("no battles in event group")
	}

	// Extract detailed card information from first battle
	firstBattle := group.battles[0]
	if len(firstBattle.Team) == 0 {
		return nil, fmt.Errorf("no team data found in battle")
	}

	playerData := firstBattle.Team[0]
	cardsData := playerData.Cards
	if len(cardsData) != 8 {
		return nil, fmt.Errorf("expected 8 cards, found %d", len(cardsData))
	}

	// Create CardInDeck objects
	deckCards := make([]CardInDeck, 0, 8)
	totalElixir := 0
	for _, cardData := range cardsData {
		card := CardInDeck{
			Name:       cardData.Name,
			ID:         cardData.ID,
			Level:      cardData.Level,
			MaxLevel:   cardData.MaxLevel,
			Rarity:     cardData.Rarity,
			ElixirCost: cardData.ElixirCost,
		}
		deckCards = append(deckCards, card)
		totalElixir += card.ElixirCost
	}

	// Create deck object
	deck := Deck{
		Cards:     deckCards,
		AvgElixir: float64(totalElixir) / 8.0,
	}

	// Validate deck
	if err := deck.Validate(); err != nil {
		return nil, fmt.Errorf("invalid deck: %w", err)
	}

	// Create event performance object
	performance := EventPerformance{
		Progress: EventProgressInProgress,
	}

	// Generate event ID
	eventID := p.generateEventID(group)

	// Create EventDeck
	eventDeck := EventDeck{
		EventID:     eventID,
		PlayerTag:   playerTag,
		EventName:   group.eventName,
		EventType:   group.eventType,
		StartTime:   group.startTime,
		Deck:        deck,
		Performance: performance,
		Battles:     make([]BattleRecord, 0, len(group.battles)),
		EventRules: map[string]any{
			"battle_mode": group.battleMode,
			"is_ladder":   false,
		},
	}

	// Add all battles
	for _, battle := range group.battles {
		battleRecord := p.createBattleRecord(battle, playerTag)
		if battleRecord != nil {
			eventDeck.AddBattle(*battleRecord)
		}
	}

	// Set max_wins based on event type
	switch group.eventType {
	case EventTypeGrandChallenge:
		maxWins := 10
		eventDeck.Performance.MaxWins = &maxWins
	case EventTypeClassicChallenge, EventTypeDraftChallenge:
		maxWins := 12
		eventDeck.Performance.MaxWins = &maxWins
	default:
		maxWins := 20 // Default for special events
		eventDeck.Performance.MaxWins = &maxWins
	}

	// Update progress
	eventDeck.Performance.UpdateProgress()

	return &eventDeck, nil
}

// createBattleRecord creates a BattleRecord from battle data
func (p *Parser) createBattleRecord(battle clashroyale.Battle, playerTag string) *BattleRecord {
	if len(battle.Team) == 0 || len(battle.Opponent) == 0 {
		return nil
	}

	// Determine win/loss
	teamCrowns := battle.Team[0].Crowns
	opponentCrowns := battle.Opponent[0].Crowns
	result := "win"
	if teamCrowns < opponentCrowns {
		result = "loss"
	} else if teamCrowns == opponentCrowns {
		result = "draw"
	}

	// Get trophy change
	var trophyChange *int
	if battle.Team[0].TrophyChange != 0 {
		tc := battle.Team[0].TrophyChange
		trophyChange = &tc
	}

	return &BattleRecord{
		Timestamp:      battle.UTCDate,
		OpponentTag:    battle.Opponent[0].Tag,
		OpponentName:   battle.Opponent[0].Name,
		Result:         result,
		Crowns:         teamCrowns,
		OpponentCrowns: opponentCrowns,
		TrophyChange:   trophyChange,
		BattleMode:     battle.GameMode.Name,
	}
}

// generateEventID generates a unique event ID
func (p *Parser) generateEventID(group eventGroup) string {
	// Normalize event name
	eventName := strings.ToLower(group.eventName)
	eventName = strings.ReplaceAll(eventName, " ", "_")

	// Format date
	dateStr := group.startTime.Format("20060102")

	// Create hash of deck cards
	deckStr := strings.Join(group.deckCards, ",")
	hash := sha256.Sum256([]byte(deckStr))
	deckHash := hex.EncodeToString(hash[:])[:8]

	return fmt.Sprintf("%s_%s_%s", eventName, dateStr, deckHash)
}
