package evaluation

import (
	"slices"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// Helper function to create a test PlayerContext
func makeTestPlayerContext() *PlayerContext {
	return &PlayerContext{
		Arena:     &clashroyale.Arena{ID: 8, Name: "Frozen Peak"},
		ArenaID:   8,
		ArenaName: "Frozen Peak",
		Collection: map[string]CardLevelInfo{
			"Giant": {
				Level:             11,
				MaxLevel:          14,
				EvolutionLevel:    0,
				MaxEvolutionLevel: 0,
				Rarity:            "Rare",
				Count:             12,
			},
			"Musketeer": {
				Level:             11,
				MaxLevel:          14,
				EvolutionLevel:    1,
				MaxEvolutionLevel: 2,
				Rarity:            "Rare",
				Count:             8,
			},
			"Fireball": {
				Level:             11,
				MaxLevel:          14,
				EvolutionLevel:    0,
				MaxEvolutionLevel: 0,
				Rarity:            "Rare",
				Count:             15,
			},
			"Zap": {
				Level:             13,
				MaxLevel:          14,
				EvolutionLevel:    0,
				MaxEvolutionLevel: 0,
				Rarity:            "Common",
				Count:             32,
			},
			"Hog Rider": {
				Level:             9,
				MaxLevel:          14,
				EvolutionLevel:    0,
				MaxEvolutionLevel: 1,
				Rarity:            "Rare",
				Count:             3,
			},
			"Log": {
				Level:             5,
				MaxLevel:          8,
				EvolutionLevel:    0,
				MaxEvolutionLevel: 0,
				Rarity:            "Legendary",
				Count:             1,
			},
		},
		UnlockedEvolutions: map[string]bool{
			"Musketeer": true,
		},
		PlayerTag:  "#ABC123",
		PlayerName: "TestPlayer",
	}
}

// Helper function to create a context with empty collection
func makeEmptyPlayerContext() *PlayerContext {
	return &PlayerContext{
		Arena:              &clashroyale.Arena{ID: 5, Name: "Spell Valley"},
		ArenaID:            5,
		ArenaName:          "Spell Valley",
		Collection:         map[string]CardLevelInfo{},
		UnlockedEvolutions: map[string]bool{},
		PlayerTag:          "#DEF456",
		PlayerName:         "EmptyPlayer",
	}
}

func TestPlayerContext_GetCardLevel(t *testing.T) {
	tests := []struct {
		name          string
		ctx           *PlayerContext
		cardName      string
		expectedLevel int
	}{
		{
			name:          "Existing card - Giant",
			ctx:           makeTestPlayerContext(),
			cardName:      "Giant",
			expectedLevel: 11,
		},
		{
			name:          "Existing card - Zap",
			ctx:           makeTestPlayerContext(),
			cardName:      "Zap",
			expectedLevel: 13,
		},
		{
			name:          "Non-existent card",
			ctx:           makeTestPlayerContext(),
			cardName:      "Mega Knight",
			expectedLevel: 0,
		},
		{
			name:          "Empty collection",
			ctx:           makeEmptyPlayerContext(),
			cardName:      "Giant",
			expectedLevel: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := tt.ctx.GetCardLevel(tt.cardName)
			if level != tt.expectedLevel {
				t.Errorf("GetCardLevel(%q) = %d, want %d", tt.cardName, level, tt.expectedLevel)
			}
		})
	}
}

func TestPlayerContext_HasCard(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *PlayerContext
		cardName string
		expected bool
	}{
		{
			name:     "Existing card",
			ctx:      makeTestPlayerContext(),
			cardName: "Giant",
			expected: true,
		},
		{
			name:     "Non-existent card",
			ctx:      makeTestPlayerContext(),
			cardName: "Mega Knight",
			expected: false,
		},
		{
			name:     "Empty collection",
			ctx:      makeEmptyPlayerContext(),
			cardName: "Giant",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			has := tt.ctx.HasCard(tt.cardName)
			if has != tt.expected {
				t.Errorf("HasCard(%q) = %v, want %v", tt.cardName, has, tt.expected)
			}
		})
	}
}

func TestPlayerContext_HasEvolution(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *PlayerContext
		cardName string
		expected bool
	}{
		{
			name:     "Card with evolution unlocked",
			ctx:      makeTestPlayerContext(),
			cardName: "Musketeer",
			expected: true,
		},
		{
			name:     "Card without evolution",
			ctx:      makeTestPlayerContext(),
			cardName: "Giant",
			expected: false,
		},
		{
			name:     "Non-existent card",
			ctx:      makeTestPlayerContext(),
			cardName: "Mega Knight",
			expected: false,
		},
		{
			name:     "Empty collection",
			ctx:      makeEmptyPlayerContext(),
			cardName: "Musketeer",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			has := tt.ctx.HasEvolution(tt.cardName)
			if has != tt.expected {
				t.Errorf("HasEvolution(%q) = %v, want %v", tt.cardName, has, tt.expected)
			}
		})
	}
}

func TestPlayerContext_GetAverageLevel(t *testing.T) {
	tests := []struct {
		name        string
		ctx         *PlayerContext
		cardNames   []string
		expectedAvg float64
	}{
		{
			name:        "All cards exist",
			ctx:         makeTestPlayerContext(),
			cardNames:   []string{"Giant", "Musketeer", "Fireball"},
			expectedAvg: 11.0,
		},
		{
			name:        "Mixed existing and non-existing",
			ctx:         makeTestPlayerContext(),
			cardNames:   []string{"Giant", "Musketeer", "Mega Knight"},
			expectedAvg: 11.0,
		},
		{
			name:        "Different levels",
			ctx:         makeTestPlayerContext(),
			cardNames:   []string{"Giant", "Zap", "Log"},
			expectedAvg: (11.0 + 13.0 + 5.0) / 3.0,
		},
		{
			name:        "No cards exist",
			ctx:         makeTestPlayerContext(),
			cardNames:   []string{"Mega Knight", "Lava Hound", "Miner"},
			expectedAvg: 0.0,
		},
		{
			name:        "Empty card list",
			ctx:         makeTestPlayerContext(),
			cardNames:   []string{},
			expectedAvg: 0.0,
		},
		{
			name:        "Empty collection",
			ctx:         makeEmptyPlayerContext(),
			cardNames:   []string{"Giant", "Musketeer"},
			expectedAvg: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avg := tt.ctx.GetAverageLevel(tt.cardNames)
			if avg != tt.expectedAvg {
				t.Errorf("GetAverageLevel(%v) = %v, want %v", tt.cardNames, avg, tt.expectedAvg)
			}
		})
	}
}

func TestPlayerContext_IsCardUnlockedInArena(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *PlayerContext
		cardName string
		expected bool
	}{
		{
			name:     "Training Camp card (Arena 0)",
			ctx:      makeTestPlayerContext(),
			cardName: "Giant",
			expected: true,
		},
		{
			name:     "Arena 1 card, player in Arena 8",
			ctx:      makeTestPlayerContext(),
			cardName: "Spear Goblins",
			expected: true,
		},
		{
			name:     "Arena 8 card, player in Arena 8",
			ctx:      makeTestPlayerContext(),
			cardName: "Electro Wizard",
			expected: true,
		},
		{
			name:     "Arena 10 card, player in Arena 8",
			ctx:      makeTestPlayerContext(),
			cardName: "Royal Ghost",
			expected: false,
		},
		{
			name: "Arena ID 0 (no restrictions)",
			ctx: &PlayerContext{
				ArenaID:    0,
				Collection: makeTestPlayerContext().Collection,
			},
			cardName: "Royal Ghost",
			expected: true,
		},
		{
			name:     "Unknown card defaults to Arena 0",
			ctx:      makeTestPlayerContext(),
			cardName: "Unknown Card",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unlocked := tt.ctx.IsCardUnlockedInArena(tt.cardName)
			if unlocked != tt.expected {
				t.Errorf("IsCardUnlockedInArena(%q) = %v, want %v", tt.cardName, unlocked, tt.expected)
			}
		})
	}
}

func TestPlayerContext_CalculateUpgradeGap(t *testing.T) {
	tests := []struct {
		name        string
		ctx         *PlayerContext
		deckCards   []deck.CardCandidate
		expectedGap int
	}{
		{
			name: "Maxed out cards",
			ctx:  makeTestPlayerContext(),
			deckCards: []deck.CardCandidate{
				{Name: "Zap", Level: 13, MaxLevel: 14},
			},
			expectedGap: 1, // 14 - 13
		},
		{
			name: "Multiple cards with gaps",
			ctx:  makeTestPlayerContext(),
			deckCards: []deck.CardCandidate{
				{Name: "Giant", Level: 11, MaxLevel: 14},
				{Name: "Musketeer", Level: 11, MaxLevel: 14},
				{Name: "Zap", Level: 13, MaxLevel: 14},
			},
			expectedGap: 7, // (14-11) + (14-11) + (14-13) = 3 + 3 + 1
		},
		{
			name: "Card not in collection",
			ctx:  makeTestPlayerContext(),
			deckCards: []deck.CardCandidate{
				{Name: "Giant", Level: 11, MaxLevel: 14},
				{Name: "Mega Knight", Level: 1, MaxLevel: 14},
			},
			expectedGap: 3, // Only Giant counts (14-11)
		},
		{
			name:        "Empty deck",
			ctx:         makeTestPlayerContext(),
			deckCards:   []deck.CardCandidate{},
			expectedGap: 0,
		},
		{
			name:        "Empty collection",
			ctx:         makeEmptyPlayerContext(),
			deckCards:   []deck.CardCandidate{{Name: "Giant", Level: 1, MaxLevel: 14}},
			expectedGap: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gap := tt.ctx.CalculateUpgradeGap(tt.deckCards)
			if gap != tt.expectedGap {
				t.Errorf("CalculateUpgradeGap() = %d, want %d", gap, tt.expectedGap)
			}
		})
	}
}

func TestPlayerContext_GetRarity(t *testing.T) {
	tests := []struct {
		name           string
		ctx            *PlayerContext
		cardName       string
		expectedRarity string
	}{
		{
			name:           "Common rarity",
			ctx:            makeTestPlayerContext(),
			cardName:       "Zap",
			expectedRarity: "Common",
		},
		{
			name:           "Rare rarity",
			ctx:            makeTestPlayerContext(),
			cardName:       "Giant",
			expectedRarity: "Rare",
		},
		{
			name:           "Legendary rarity",
			ctx:            makeTestPlayerContext(),
			cardName:       "Log",
			expectedRarity: "Legendary",
		},
		{
			name:           "Non-existent card",
			ctx:            makeTestPlayerContext(),
			cardName:       "Mega Knight",
			expectedRarity: "",
		},
		{
			name:           "Empty collection",
			ctx:            makeEmptyPlayerContext(),
			cardName:       "Giant",
			expectedRarity: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rarity := tt.ctx.GetRarity(tt.cardName)
			if rarity != tt.expectedRarity {
				t.Errorf("GetRarity(%q) = %q, want %q", tt.cardName, rarity, tt.expectedRarity)
			}
		})
	}
}

func TestPlayerContext_IsEvolutionAvailable(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *PlayerContext
		cardName string
		expected bool
	}{
		{
			name:     "Card with evolution potential",
			ctx:      makeTestPlayerContext(),
			cardName: "Musketeer",
			expected: true,
		},
		{
			name:     "Card without evolution",
			ctx:      makeTestPlayerContext(),
			cardName: "Giant",
			expected: false,
		},
		{
			name:     "Non-existent card",
			ctx:      makeTestPlayerContext(),
			cardName: "Mega Knight",
			expected: false,
		},
		{
			name:     "Empty collection",
			ctx:      makeEmptyPlayerContext(),
			cardName: "Musketeer",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			available := tt.ctx.IsEvolutionAvailable(tt.cardName)
			if available != tt.expected {
				t.Errorf("IsEvolutionAvailable(%q) = %v, want %v", tt.cardName, available, tt.expected)
			}
		})
	}
}

func TestPlayerContext_CanEvolve(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *PlayerContext
		cardName string
		expected bool
	}{
		{
			name:     "Card ready to evolve (level 11, count 8)",
			ctx:      makeTestPlayerContext(),
			cardName: "Musketeer",
			expected: true,
		},
		{
			name:     "Card not high enough level (level 9, count 3)",
			ctx:      makeTestPlayerContext(),
			cardName: "Hog Rider",
			expected: false,
		},
		{
			name: "Card high level but low count (level 11, count 3)",
			ctx: &PlayerContext{
				Collection: map[string]CardLevelInfo{
					"Test Card": {Level: 11, Count: 3},
				},
			},
			cardName: "Test Card",
			expected: false,
		},
		{
			name: "Card low level but high count (level 9, count 10)",
			ctx: &PlayerContext{
				Collection: map[string]CardLevelInfo{
					"Test Card": {Level: 9, Count: 10},
				},
			},
			cardName: "Test Card",
			expected: false,
		},
		{
			name:     "Non-existent card",
			ctx:      makeTestPlayerContext(),
			cardName: "Mega Knight",
			expected: false,
		},
		{
			name:     "Empty collection",
			ctx:      makeEmptyPlayerContext(),
			cardName: "Musketeer",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canEvolve := tt.ctx.CanEvolve(tt.cardName)
			if canEvolve != tt.expected {
				t.Errorf("CanEvolve(%q) = %v, want %v", tt.cardName, canEvolve, tt.expected)
			}
		})
	}
}

func TestPlayerContext_GetEvolutionProgress(t *testing.T) {
	tests := []struct {
		name              string
		ctx               *PlayerContext
		cardName          string
		expectedCurrLvl   int
		expectedMaxLvl    int
		expectedCurrCount int
		expectedReqCount  int
	}{
		{
			name:              "Card with evolution",
			ctx:               makeTestPlayerContext(),
			cardName:          "Musketeer",
			expectedCurrLvl:   1,
			expectedMaxLvl:    2,
			expectedCurrCount: 8,
			expectedReqCount:  5,
		},
		{
			name:              "Card without evolution",
			ctx:               makeTestPlayerContext(),
			cardName:          "Giant",
			expectedCurrLvl:   0,
			expectedMaxLvl:    0,
			expectedCurrCount: 12,
			expectedReqCount:  5,
		},
		{
			name:              "Non-existent card",
			ctx:               makeTestPlayerContext(),
			cardName:          "Mega Knight",
			expectedCurrLvl:   0,
			expectedMaxLvl:    0,
			expectedCurrCount: 0,
			expectedReqCount:  5,
		},
		{
			name:              "Empty collection",
			ctx:               makeEmptyPlayerContext(),
			cardName:          "Musketeer",
			expectedCurrLvl:   0,
			expectedMaxLvl:    0,
			expectedCurrCount: 0,
			expectedReqCount:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currLvl, maxLvl, currCount, reqCount := tt.ctx.GetEvolutionProgress(tt.cardName)
			if currLvl != tt.expectedCurrLvl {
				t.Errorf("GetEvolutionProgress(%q) current level = %d, want %d", tt.cardName, currLvl, tt.expectedCurrLvl)
			}
			if maxLvl != tt.expectedMaxLvl {
				t.Errorf("GetEvolutionProgress(%q) max level = %d, want %d", tt.cardName, maxLvl, tt.expectedMaxLvl)
			}
			if currCount != tt.expectedCurrCount {
				t.Errorf("GetEvolutionProgress(%q) current count = %d, want %d", tt.cardName, currCount, tt.expectedCurrCount)
			}
			if reqCount != tt.expectedReqCount {
				t.Errorf("GetEvolutionProgress(%q) required count = %d, want %d", tt.cardName, reqCount, tt.expectedReqCount)
			}
		})
	}
}

func TestPlayerContext_GetUnlockedEvolutionCards(t *testing.T) {
	tests := []struct {
		name          string
		ctx           *PlayerContext
		expectedCount int
		expectedCards []string
	}{
		{
			name:          "One evolved card",
			ctx:           makeTestPlayerContext(),
			expectedCount: 1,
			expectedCards: []string{"Musketeer"},
		},
		{
			name:          "No evolved cards",
			ctx:           makeEmptyPlayerContext(),
			expectedCount: 0,
			expectedCards: []string{},
		},
		{
			name: "Multiple evolved cards",
			ctx: &PlayerContext{
				UnlockedEvolutions: map[string]bool{
					"Musketeer": true,
					"Giant":     true,
					"Hog Rider": true,
				},
			},
			expectedCount: 3,
			expectedCards: []string{"Musketeer", "Giant", "Hog Rider"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards := tt.ctx.GetUnlockedEvolutionCards()
			if len(cards) != tt.expectedCount {
				t.Errorf("GetUnlockedEvolutionCards() length = %d, want %d", len(cards), tt.expectedCount)
			}
			// Check that expected cards are present
			for _, expected := range tt.expectedCards {
				found := slices.Contains(cards, expected)
				if !found {
					t.Errorf("GetUnlockedEvolutionCards() missing expected card %q", expected)
				}
			}
		})
	}
}

func TestPlayerContext_GetEvolvableCards(t *testing.T) {
	tests := []struct {
		name          string
		ctx           *PlayerContext
		expectedCount int
	}{
		{
			name:          "One evolvable card",
			ctx:           makeTestPlayerContext(),
			expectedCount: 1, // Hog Rider has MaxEvolutionLevel 1 but EvolutionLevel 0
		},
		{
			name:          "No evolvable cards",
			ctx:           makeEmptyPlayerContext(),
			expectedCount: 0,
		},
		{
			name: "Mixed evolvable and evolved",
			ctx: &PlayerContext{
				Collection: map[string]CardLevelInfo{
					"Card1": {MaxEvolutionLevel: 1, EvolutionLevel: 0},
					"Card2": {MaxEvolutionLevel: 1, EvolutionLevel: 1},
					"Card3": {MaxEvolutionLevel: 0, EvolutionLevel: 0},
				},
			},
			expectedCount: 1, // Only Card1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards := tt.ctx.GetEvolvableCards()
			if len(cards) != tt.expectedCount {
				t.Errorf("GetEvolvableCards() length = %d, want %d", len(cards), tt.expectedCount)
			}
		})
	}
}

func TestPlayerContext_NewPlayerContextFromPlayer(t *testing.T) {
	tests := []struct {
		name    string
		player  *clashroyale.Player
		wantNil bool
		verify  func(*testing.T, *PlayerContext)
	}{
		{
			name:    "Nil player",
			player:  nil,
			wantNil: true,
		},
		{
			name: "Player with cards",
			player: &clashroyale.Player{
				Tag:  "#TEST123",
				Name: "TestPlayer",
				Arena: clashroyale.Arena{
					ID:   8,
					Name: "Frozen Peak",
				},
				Cards: []clashroyale.Card{
					{
						Name:              "Giant",
						Level:             11,
						MaxLevel:          14,
						EvolutionLevel:    0,
						MaxEvolutionLevel: 0,
						Rarity:            "Rare",
						Count:             12,
					},
					{
						Name:              "Musketeer",
						Level:             11,
						MaxLevel:          14,
						EvolutionLevel:    1,
						MaxEvolutionLevel: 2,
						Rarity:            "Rare",
						Count:             8,
					},
				},
			},
			wantNil: false,
			verify: func(t *testing.T, ctx *PlayerContext) {
				if ctx.PlayerTag != "#TEST123" {
					t.Errorf("PlayerTag = %q, want #TEST123", ctx.PlayerTag)
				}
				if ctx.PlayerName != "TestPlayer" {
					t.Errorf("PlayerName = %q, want TestPlayer", ctx.PlayerName)
				}
				if ctx.ArenaID != 8 {
					t.Errorf("ArenaID = %d, want 8", ctx.ArenaID)
				}
				if ctx.ArenaName != "Frozen Peak" {
					t.Errorf("ArenaName = %q, want Frozen Peak", ctx.ArenaName)
				}
				if len(ctx.Collection) != 2 {
					t.Errorf("Collection length = %d, want 2", len(ctx.Collection))
				}
				if len(ctx.UnlockedEvolutions) != 1 {
					t.Errorf("UnlockedEvolutions length = %d, want 1", len(ctx.UnlockedEvolutions))
				}
				// Check Musketeer evolution is tracked
				if !ctx.UnlockedEvolutions["Musketeer"] {
					t.Error("Musketeer evolution not tracked")
				}
			},
		},
		{
			name: "Player with no cards",
			player: &clashroyale.Player{
				Tag:  "#EMPTY",
				Name: "EmptyPlayer",
				Arena: clashroyale.Arena{
					ID:   0,
					Name: "Training Camp",
				},
				Cards: []clashroyale.Card{},
			},
			wantNil: false,
			verify: func(t *testing.T, ctx *PlayerContext) {
				if len(ctx.Collection) != 0 {
					t.Errorf("Collection length = %d, want 0", len(ctx.Collection))
				}
				if len(ctx.UnlockedEvolutions) != 0 {
					t.Errorf("UnlockedEvolutions length = %d, want 0", len(ctx.UnlockedEvolutions))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewPlayerContextFromPlayer(tt.player)

			if tt.wantNil {
				if ctx != nil {
					t.Errorf("NewPlayerContextFromPlayer() = %v, want nil", ctx)
				}
				return
			}

			if ctx == nil {
				t.Fatal("NewPlayerContextFromPlayer() returned nil, want non-nil")
			}

			if tt.verify != nil {
				tt.verify(t, ctx)
			}
		})
	}
}
