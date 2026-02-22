package events

import (
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatal("NewParser returned nil")
	}
	if len(parser.eventBattleModes) == 0 {
		t.Error("Parser should have event battle modes configured")
	}
	if len(parser.eventPatterns) == 0 {
		t.Error("Parser should have event patterns configured")
	}
}

func TestParser_IsEventBattle(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		battle   clashroyale.Battle
		expected bool
	}{
		{
			name: "Grand Challenge",
			battle: clashroyale.Battle{
				GameMode: clashroyale.GameMode{
					Name: "Grand Challenge",
				},
				IsLadderTournament: false,
			},
			expected: true,
		},
		{
			name: "Classic Challenge",
			battle: clashroyale.Battle{
				GameMode: clashroyale.GameMode{
					Name: "Classic Challenge",
				},
				IsLadderTournament: false,
			},
			expected: true,
		},
		{
			name: "Tournament",
			battle: clashroyale.Battle{
				GameMode: clashroyale.GameMode{
					Name: "Tournament",
				},
				IsLadderTournament: false,
			},
			expected: true,
		},
		{
			name: "Ladder Battle",
			battle: clashroyale.Battle{
				GameMode: clashroyale.GameMode{
					Name: "Ladder",
				},
				IsLadderTournament: false,
			},
			expected: false,
		},
		{
			name: "Non-ladder special mode",
			battle: clashroyale.Battle{
				GameMode: clashroyale.GameMode{
					Name: "Special Event Mode",
				},
				IsLadderTournament: true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isEventBattle(tt.battle)
			if result != tt.expected {
				t.Errorf("isEventBattle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFilterEventBattles(t *testing.T) {
	battles := []clashroyale.Battle{
		{GameMode: clashroyale.GameMode{Name: "Ladder"}},
		{GameMode: clashroyale.GameMode{Name: "Grand Challenge"}},
		{GameMode: clashroyale.GameMode{Name: "Tournament"}},
		{GameMode: clashroyale.GameMode{Name: "1v1"}},
		{GameMode: clashroyale.GameMode{Name: "Random"}, IsLadderTournament: true},
	}

	filtered := FilterEventBattles(battles)
	if len(filtered) != 3 {
		t.Fatalf("FilterEventBattles returned %d battles, want 3", len(filtered))
	}

	for _, battle := range filtered {
		if !IsEventBattle(battle) {
			t.Fatalf("filtered battle should be an event battle: %+v", battle)
		}
	}
}

func TestParser_ExtractEventData(t *testing.T) {
	parser := NewParser()

	battle := clashroyale.Battle{
		GameMode: clashroyale.GameMode{
			Name: "Grand Challenge",
		},
		Team: []clashroyale.BattleTeam{
			{
				Cards: []clashroyale.Card{
					{Name: "Knight"},
					{Name: "Archers"},
					{Name: "Fireball"},
					{Name: "Zap"},
					{Name: "Mini P.E.K.K.A"},
					{Name: "Musketeer"},
					{Name: "Giant"},
					{Name: "Valkyrie"},
				},
			},
		},
	}

	data := parser.extractEventData(battle)

	if data.eventType != EventTypeGrandChallenge {
		t.Errorf("eventType = %v, want %v", data.eventType, EventTypeGrandChallenge)
	}
	if data.eventName != "Grand Challenge" {
		t.Errorf("eventName = %v, want Grand Challenge", data.eventName)
	}
	if data.battleMode != "Grand Challenge" {
		t.Errorf("battleMode = %v, want Grand Challenge", data.battleMode)
	}
	if len(data.deckCards) != 8 {
		t.Errorf("deckCards length = %d, want 8", len(data.deckCards))
	}
}

func TestParser_ExtractDeckFromBattle(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		battle   clashroyale.Battle
		expected int
	}{
		{
			name: "8 cards",
			battle: clashroyale.Battle{
				Team: []clashroyale.BattleTeam{
					{
						Cards: []clashroyale.Card{
							{Name: "Knight"},
							{Name: "Archers"},
							{Name: "Fireball"},
							{Name: "Zap"},
							{Name: "Mini P.E.K.K.A"},
							{Name: "Musketeer"},
							{Name: "Giant"},
							{Name: "Valkyrie"},
						},
					},
				},
			},
			expected: 8,
		},
		{
			name: "No team data",
			battle: clashroyale.Battle{
				Team: []clashroyale.BattleTeam{},
			},
			expected: 0,
		},
		{
			name: "Empty cards",
			battle: clashroyale.Battle{
				Team: []clashroyale.BattleTeam{
					{Cards: []clashroyale.Card{}},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.extractDeckFromBattle(tt.battle)
			if len(result) != tt.expected {
				t.Errorf("extractDeckFromBattle() returned %d cards, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestParser_IsSameDeck(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		deck1    []string
		deck2    []string
		expected bool
	}{
		{
			name:     "Same decks",
			deck1:    []string{"Archers", "Fireball", "Giant", "Knight", "Mini P.E.K.K.A", "Musketeer", "Valkyrie", "Zap"},
			deck2:    []string{"Archers", "Fireball", "Giant", "Knight", "Mini P.E.K.K.A", "Musketeer", "Valkyrie", "Zap"},
			expected: true,
		},
		{
			name:     "Different decks",
			deck1:    []string{"Archers", "Fireball", "Giant", "Knight", "Mini P.E.K.K.A", "Musketeer", "Valkyrie", "Zap"},
			deck2:    []string{"Archers", "Fireball", "Giant", "Knight", "Mini P.E.K.K.A", "Musketeer", "Valkyrie", "Log"},
			expected: false,
		},
		{
			name:     "Different lengths",
			deck1:    []string{"Archers", "Fireball", "Giant", "Knight"},
			deck2:    []string{"Archers", "Fireball", "Giant", "Knight", "Mini P.E.K.K.A", "Musketeer", "Valkyrie", "Zap"},
			expected: false,
		},
		{
			name:     "Empty decks",
			deck1:    []string{},
			deck2:    []string{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isSameDeck(tt.deck1, tt.deck2)
			if result != tt.expected {
				t.Errorf("isSameDeck() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParser_IsNewEvent(t *testing.T) {
	parser := NewParser()
	baseTime := time.Now()

	tests := []struct {
		name         string
		battle       clashroyale.Battle
		currentGroup *eventGroup
		expected     bool
	}{
		{
			name: "No current group",
			battle: clashroyale.Battle{
				UTCDate:  baseTime,
				GameMode: clashroyale.GameMode{Name: "Grand Challenge"},
			},
			currentGroup: nil,
			expected:     true,
		},
		{
			name: "Time gap > 1 hour",
			battle: clashroyale.Battle{
				UTCDate:  baseTime.Add(2 * time.Hour),
				GameMode: clashroyale.GameMode{Name: "Grand Challenge"},
			},
			currentGroup: &eventGroup{
				eventType: EventTypeGrandChallenge,
				startTime: baseTime,
				deckCards: []string{"Knight", "Archers"},
				battles:   []clashroyale.Battle{{UTCDate: baseTime}},
			},
			expected: true,
		},
		{
			name: "Different event type",
			battle: clashroyale.Battle{
				UTCDate:  baseTime.Add(10 * time.Minute),
				GameMode: clashroyale.GameMode{Name: "Classic Challenge"},
				Team: []clashroyale.BattleTeam{
					{Cards: []clashroyale.Card{{Name: "Knight"}, {Name: "Archers"}}},
				},
			},
			currentGroup: &eventGroup{
				eventType: EventTypeGrandChallenge,
				startTime: baseTime,
				deckCards: []string{"Archers", "Knight"},
				battles:   []clashroyale.Battle{{UTCDate: baseTime}},
			},
			expected: true,
		},
		{
			name: "Different deck",
			battle: clashroyale.Battle{
				UTCDate:  baseTime.Add(10 * time.Minute),
				GameMode: clashroyale.GameMode{Name: "Grand Challenge"},
				Team: []clashroyale.BattleTeam{
					{Cards: []clashroyale.Card{{Name: "Giant"}, {Name: "Wizard"}}},
				},
			},
			currentGroup: &eventGroup{
				eventType: EventTypeGrandChallenge,
				startTime: baseTime,
				deckCards: []string{"Archers", "Knight"},
				battles:   []clashroyale.Battle{{UTCDate: baseTime}},
			},
			expected: true,
		},
		{
			name: "Same event continues",
			battle: clashroyale.Battle{
				UTCDate:  baseTime.Add(10 * time.Minute),
				GameMode: clashroyale.GameMode{Name: "Grand Challenge"},
				Team: []clashroyale.BattleTeam{
					{Cards: []clashroyale.Card{{Name: "Archers"}, {Name: "Knight"}}},
				},
			},
			currentGroup: &eventGroup{
				eventType: EventTypeGrandChallenge,
				startTime: baseTime,
				deckCards: []string{"Archers", "Knight"},
				battles:   []clashroyale.Battle{{UTCDate: baseTime}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isNewEvent(tt.battle, tt.currentGroup)
			if result != tt.expected {
				t.Errorf("isNewEvent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParser_CreateBattleRecord(t *testing.T) {
	parser := NewParser()
	battleTime := time.Now()
	trophyChange := 10
	playerCards := []clashroyale.Card{
		{Name: "Hog Rider"},
		{Name: "Earthquake"},
		{Name: "The Log"},
		{Name: "Firecracker"},
		{Name: "Cannon"},
		{Name: "Skeletons"},
		{Name: "Ice Spirit"},
		{Name: "Knight"},
	}
	opponentCards := []clashroyale.Card{
		{Name: "Golem"},
		{Name: "Night Witch"},
		{Name: "Baby Dragon"},
		{Name: "Lightning"},
		{Name: "Tornado"},
		{Name: "Barbarian Barrel"},
		{Name: "Electro Dragon"},
		{Name: "Lumberjack"},
	}

	tests := []struct {
		name    string
		battle  clashroyale.Battle
		wantNil bool
	}{
		{
			name: "Valid win battle",
			battle: clashroyale.Battle{
				UTCDate:  battleTime,
				GameMode: clashroyale.GameMode{Name: "Grand Challenge"},
				Team: []clashroyale.BattleTeam{
					{
						Tag:          "#PLAYER",
						Name:         "TestPlayer",
						Crowns:       3,
						TrophyChange: trophyChange,
						Cards:        playerCards,
					},
				},
				Opponent: []clashroyale.BattleTeam{
					{
						Tag:    "#OPPONENT",
						Name:   "OpponentPlayer",
						Crowns: 1,
						Cards:  opponentCards,
					},
				},
			},
			wantNil: false,
		},
		{
			name: "Valid loss battle",
			battle: clashroyale.Battle{
				UTCDate:  battleTime,
				GameMode: clashroyale.GameMode{Name: "Grand Challenge"},
				Team: []clashroyale.BattleTeam{
					{
						Tag:    "#PLAYER",
						Name:   "TestPlayer",
						Crowns: 1,
						Cards:  playerCards,
					},
				},
				Opponent: []clashroyale.BattleTeam{
					{
						Tag:    "#OPPONENT",
						Name:   "OpponentPlayer",
						Crowns: 3,
						Cards:  opponentCards,
					},
				},
			},
			wantNil: false,
		},
		{
			name: "No team data",
			battle: clashroyale.Battle{
				UTCDate:  battleTime,
				GameMode: clashroyale.GameMode{Name: "Grand Challenge"},
				Team:     []clashroyale.BattleTeam{},
				Opponent: []clashroyale.BattleTeam{
					{Tag: "#OPPONENT", Name: "Opponent", Crowns: 1},
				},
			},
			wantNil: true,
		},
		{
			name: "No opponent data",
			battle: clashroyale.Battle{
				UTCDate:  battleTime,
				GameMode: clashroyale.GameMode{Name: "Grand Challenge"},
				Team: []clashroyale.BattleTeam{
					{Tag: "#PLAYER", Name: "Player", Crowns: 3},
				},
				Opponent: []clashroyale.BattleTeam{},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.createBattleRecord(tt.battle, "#PLAYER")
			if (result == nil) != tt.wantNil {
				t.Errorf("createBattleRecord() nil = %v, wantNil = %v", result == nil, tt.wantNil)
			}

			if !tt.wantNil && result != nil {
				// Verify basic fields
				if result.Timestamp != battleTime {
					t.Errorf("Timestamp = %v, want %v", result.Timestamp, battleTime)
				}
				if result.BattleMode != tt.battle.GameMode.Name {
					t.Errorf("BattleMode = %v, want %v", result.BattleMode, tt.battle.GameMode.Name)
				}
				if result.Crowns != tt.battle.Team[0].Crowns {
					t.Errorf("Crowns = %v, want %v", result.Crowns, tt.battle.Team[0].Crowns)
				}
				if result.OpponentCrowns != tt.battle.Opponent[0].Crowns {
					t.Errorf("OpponentCrowns = %v, want %v", result.OpponentCrowns, tt.battle.Opponent[0].Crowns)
				}
				if len(tt.battle.Team[0].Cards) > 0 {
					if len(result.PlayerDeck) != len(tt.battle.Team[0].Cards) {
						t.Errorf("PlayerDeck length = %d, want %d", len(result.PlayerDeck), len(tt.battle.Team[0].Cards))
					}
					if result.PlayerDeckHash == "" {
						t.Error("PlayerDeckHash should not be empty when player cards are present")
					}
				}
				if len(tt.battle.Opponent[0].Cards) > 0 {
					if len(result.OpponentDeck) != len(tt.battle.Opponent[0].Cards) {
						t.Errorf("OpponentDeck length = %d, want %d", len(result.OpponentDeck), len(tt.battle.Opponent[0].Cards))
					}
					if result.OpponentDeckHash == "" {
						t.Error("OpponentDeckHash should not be empty when opponent cards are present")
					}
				}

				// Check win/loss
				expectedResult := "win"
				if tt.battle.Team[0].Crowns < tt.battle.Opponent[0].Crowns {
					expectedResult = "loss"
				} else if tt.battle.Team[0].Crowns == tt.battle.Opponent[0].Crowns {
					expectedResult = "draw"
				}
				if result.Result != expectedResult {
					t.Errorf("Result = %v, want %v", result.Result, expectedResult)
				}
			}
		})
	}
}

func TestParser_CreateBattleRecord_DeckHashNormalization(t *testing.T) {
	parser := NewParser()
	battleTime := time.Now()

	battleA := clashroyale.Battle{
		UTCDate:  battleTime,
		GameMode: clashroyale.GameMode{Name: "Classic Challenge"},
		Team: []clashroyale.BattleTeam{
			{
				Tag:    "#PLAYER",
				Crowns: 1,
				Cards: []clashroyale.Card{
					{Name: "Knight"},
					{Name: "Archers"},
				},
			},
		},
		Opponent: []clashroyale.BattleTeam{
			{
				Tag:    "#OPP",
				Crowns: 0,
				Cards: []clashroyale.Card{
					{Name: "Zap"},
					{Name: "Fireball"},
				},
			},
		},
	}

	battleB := clashroyale.Battle{
		UTCDate:  battleTime,
		GameMode: clashroyale.GameMode{Name: "Classic Challenge"},
		Team: []clashroyale.BattleTeam{
			{
				Tag:    "#PLAYER",
				Crowns: 1,
				Cards: []clashroyale.Card{
					{Name: "Archers"},
					{Name: "Knight"},
				},
			},
		},
		Opponent: []clashroyale.BattleTeam{
			{
				Tag:    "#OPP",
				Crowns: 0,
				Cards: []clashroyale.Card{
					{Name: "Fireball"},
					{Name: "Zap"},
				},
			},
		},
	}

	recordA := parser.createBattleRecord(battleA, "#PLAYER")
	recordB := parser.createBattleRecord(battleB, "#PLAYER")

	if recordA == nil || recordB == nil {
		t.Fatal("expected non-nil records")
	}

	if recordA.PlayerDeckHash != recordB.PlayerDeckHash {
		t.Errorf("player deck hashes differ for same cards in different order: %s vs %s", recordA.PlayerDeckHash, recordB.PlayerDeckHash)
	}

	if recordA.OpponentDeckHash != recordB.OpponentDeckHash {
		t.Errorf("opponent deck hashes differ for same cards in different order: %s vs %s", recordA.OpponentDeckHash, recordB.OpponentDeckHash)
	}
}

func TestParser_GenerateEventID(t *testing.T) {
	parser := NewParser()
	baseTime := time.Date(2024, 12, 11, 10, 30, 0, 0, time.UTC)

	group := eventGroup{
		eventName: "Grand Challenge",
		startTime: baseTime,
		deckCards: []string{"Archers", "Fireball", "Giant", "Knight", "Mini P.E.K.K.A", "Musketeer", "Valkyrie", "Zap"},
	}

	eventID := parser.generateEventID(group)

	// Check format: eventname_YYYYMMDD_hash
	if len(eventID) == 0 {
		t.Error("generateEventID() returned empty string")
	}

	// Should contain event name (normalized)
	if !contains(eventID, "grand_challenge") && !contains(eventID, "challenge") {
		t.Errorf("Event ID should contain normalized event name, got: %s", eventID)
	}

	// Should contain date
	if !contains(eventID, "20241211") {
		t.Errorf("Event ID should contain date 20241211, got: %s", eventID)
	}

	// Generating with same inputs should produce same ID
	eventID2 := parser.generateEventID(group)
	if eventID != eventID2 {
		t.Errorf("generateEventID() not deterministic: %s != %s", eventID, eventID2)
	}

	// Different deck should produce different ID
	group2 := group
	group2.deckCards = []string{"Different", "Deck", "Cards", "Here", "Test", "A", "B", "C"}
	eventID3 := parser.generateEventID(group2)
	if eventID == eventID3 {
		t.Error("Different decks should produce different event IDs")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParser_GroupBattlesByEvent(t *testing.T) {
	parser := NewParser()
	baseTime := time.Now()

	battles := []clashroyale.Battle{
		// Event 1: Grand Challenge (3 battles)
		{
			UTCDate:            baseTime,
			GameMode:           clashroyale.GameMode{Name: "Grand Challenge"},
			IsLadderTournament: false,
			Team: []clashroyale.BattleTeam{
				{Cards: []clashroyale.Card{{Name: "Knight"}, {Name: "Archers"}, {Name: "Fireball"}, {Name: "Zap"}, {Name: "Giant"}, {Name: "Musketeer"}, {Name: "Valkyrie"}, {Name: "Mini P.E.K.K.A"}}},
			},
		},
		{
			UTCDate:            baseTime.Add(5 * time.Minute),
			GameMode:           clashroyale.GameMode{Name: "Grand Challenge"},
			IsLadderTournament: false,
			Team: []clashroyale.BattleTeam{
				{Cards: []clashroyale.Card{{Name: "Knight"}, {Name: "Archers"}, {Name: "Fireball"}, {Name: "Zap"}, {Name: "Giant"}, {Name: "Musketeer"}, {Name: "Valkyrie"}, {Name: "Mini P.E.K.K.A"}}},
			},
		},
		{
			UTCDate:            baseTime.Add(10 * time.Minute),
			GameMode:           clashroyale.GameMode{Name: "Grand Challenge"},
			IsLadderTournament: false,
			Team: []clashroyale.BattleTeam{
				{Cards: []clashroyale.Card{{Name: "Knight"}, {Name: "Archers"}, {Name: "Fireball"}, {Name: "Zap"}, {Name: "Giant"}, {Name: "Musketeer"}, {Name: "Valkyrie"}, {Name: "Mini P.E.K.K.A"}}},
			},
		},
		// Ladder battle (should be filtered out)
		{
			UTCDate:            baseTime.Add(15 * time.Minute),
			GameMode:           clashroyale.GameMode{Name: "Ladder"},
			IsLadderTournament: false,
		},
		// Event 2: Classic Challenge with different deck (2 battles)
		{
			UTCDate:            baseTime.Add(2 * time.Hour),
			GameMode:           clashroyale.GameMode{Name: "Classic Challenge"},
			IsLadderTournament: false,
			Team: []clashroyale.BattleTeam{
				{Cards: []clashroyale.Card{{Name: "Hog Rider"}, {Name: "Goblins"}, {Name: "Rocket"}, {Name: "Log"}, {Name: "Ice Spirit"}, {Name: "Cannon"}, {Name: "Fireball"}, {Name: "Skeletons"}}},
			},
		},
		{
			UTCDate:            baseTime.Add(2*time.Hour + 5*time.Minute),
			GameMode:           clashroyale.GameMode{Name: "Classic Challenge"},
			IsLadderTournament: false,
			Team: []clashroyale.BattleTeam{
				{Cards: []clashroyale.Card{{Name: "Hog Rider"}, {Name: "Goblins"}, {Name: "Rocket"}, {Name: "Log"}, {Name: "Ice Spirit"}, {Name: "Cannon"}, {Name: "Fireball"}, {Name: "Skeletons"}}},
			},
		},
	}

	groups := parser.groupBattlesByEvent(battles, "#PLAYER")

	// Should have 2 event groups (Grand Challenge and Classic Challenge)
	if len(groups) != 2 {
		t.Fatalf("Expected 2 event groups, got %d", len(groups))
	}

	// First group should have 3 battles
	if len(groups[0].battles) != 3 {
		t.Errorf("First group should have 3 battles, got %d", len(groups[0].battles))
	}
	if groups[0].eventType != EventTypeGrandChallenge {
		t.Errorf("First group should be Grand Challenge, got %v", groups[0].eventType)
	}

	// Second group should have 2 battles
	if len(groups[1].battles) != 2 {
		t.Errorf("Second group should have 2 battles, got %d", len(groups[1].battles))
	}
	if groups[1].eventType != EventTypeClassicChallenge {
		t.Errorf("Second group should be Classic Challenge, got %v", groups[1].eventType)
	}
}
