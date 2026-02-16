package evaluation

import (
	"slices"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestIdentifyMissingCards(t *testing.T) {
	tests := []struct {
		name              string
		deckCards         []deck.CardCandidate
		playerCollection  map[string]bool
		playerArena       int
		expectPlayable    bool
		expectedMissing   int
		expectedAvailable int
	}{
		{
			name: "All cards owned",
			deckCards: []deck.CardCandidate{
				makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
			},
			playerCollection: map[string]bool{
				"Knight":    true,
				"Hog Rider": true,
				"Fireball":  true,
			},
			playerArena:       10,
			expectPlayable:    true,
			expectedMissing:   0,
			expectedAvailable: 3,
		},
		{
			name: "Some cards missing",
			deckCards: []deck.CardCandidate{
				makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
				makeCard("Mega Knight", deck.RoleWinCondition, 11, 14, "Legendary", 7),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
			},
			playerCollection: map[string]bool{
				"Knight":   true,
				"Fireball": true,
				// Mega Knight not owned
			},
			playerArena:       10,
			expectPlayable:    false,
			expectedMissing:   1,
			expectedAvailable: 2,
		},
		{
			name: "Card locked by arena",
			deckCards: []deck.CardCandidate{
				makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
				makeCard("Mega Knight", deck.RoleWinCondition, 11, 14, "Legendary", 7),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
			},
			playerCollection: map[string]bool{
				"Knight": true,
				// Mega Knight and Fireball not owned (Mega Knight also not unlocked)
			},
			playerArena:       5,
			expectPlayable:    false,
			expectedMissing:   2,
			expectedAvailable: 1,
		},
		{
			name: "Nil player collection means all cards missing",
			deckCards: []deck.CardCandidate{
				makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
			},
			playerCollection:  nil,
			playerArena:       0,
			expectPlayable:    false,
			expectedMissing:   2,
			expectedAvailable: 0,
		},
		{
			name:              "Empty deck",
			deckCards:         []deck.CardCandidate{},
			playerCollection:  map[string]bool{},
			playerArena:       0,
			expectPlayable:    true, // Empty deck has no missing cards
			expectedMissing:   0,
			expectedAvailable: 0,
		},
		{
			name: "Multiple missing cards with alternatives",
			deckCards: []deck.CardCandidate{
				makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
				makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Mega Knight", deck.RoleWinCondition, 11, 14, "Legendary", 7),
			},
			playerCollection: map[string]bool{
				// Player owns Ice Golem, Dark Prince (alternatives to Knight/Valkyrie)
				"Ice Golem":   true,
				"Dark Prince": true,
			},
			playerArena:       10,
			expectPlayable:    false,
			expectedMissing:   3,
			expectedAvailable: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IdentifyMissingCards(tt.deckCards, tt.playerCollection, tt.playerArena)

			if result == nil {
				t.Fatal("IdentifyMissingCards() returned nil")
			}

			// Check playable status
			if result.IsPlayable != tt.expectPlayable {
				t.Errorf("IdentifyMissingCards() IsPlayable = %v, want %v",
					result.IsPlayable, tt.expectPlayable)
			}

			// Check missing count
			if result.MissingCount != tt.expectedMissing {
				t.Errorf("IdentifyMissingCards() MissingCount = %d, want %d",
					result.MissingCount, tt.expectedMissing)
			}

			// Check available count
			if result.AvailableCount != tt.expectedAvailable {
				t.Errorf("IdentifyMissingCards() AvailableCount = %d, want %d",
					result.AvailableCount, tt.expectedAvailable)
			}

			// Check deck length
			if len(result.Deck) != len(tt.deckCards) {
				t.Errorf("IdentifyMissingCards() Deck length = %d, want %d",
					len(result.Deck), len(tt.deckCards))
			}

			// Check missing cards length
			if len(result.MissingCards) != tt.expectedMissing {
				t.Errorf("IdentifyMissingCards() MissingCards length = %d, want %d",
					len(result.MissingCards), tt.expectedMissing)
			}

			// Verify structure of missing cards
			for i, missing := range result.MissingCards {
				if missing.Name == "" {
					t.Errorf("IdentifyMissingCards() MissingCards[%d].Name is empty", i)
				}
				if missing.Rarity == "" {
					t.Errorf("IdentifyMissingCards() MissingCards[%d].Rarity is empty", i)
				}
				if missing.UnlockArenaName == "" {
					t.Errorf("IdentifyMissingCards() MissingCards[%d].UnlockArenaName is empty", i)
				}
			}

			// Check that cards are sorted by unlock arena
			for i := 1; i < len(result.MissingCards); i++ {
				if result.MissingCards[i-1].UnlockArena > result.MissingCards[i].UnlockArena {
					t.Errorf("IdentifyMissingCards() MissingCards not sorted by unlock arena: "+
						"[%d].UnlockArena=%d > [%d].UnlockArena=%d",
						i-1, result.MissingCards[i-1].UnlockArena,
						i, result.MissingCards[i].UnlockArena)
				}
			}
		})
	}
}

func TestGetCardUnlockArena(t *testing.T) {
	tests := []struct {
		name          string
		cardName      string
		expectedArena int
	}{
		// Training Camp (Arena 0)
		{name: "Knight", cardName: "Knight", expectedArena: 0},
		{name: "Archers", cardName: "Archers", expectedArena: 0},
		{name: "Giant", cardName: "Giant", expectedArena: 0},
		{name: "P.E.K.K.A", cardName: "P.E.K.K.A", expectedArena: 0},
		{name: "Musketeer", cardName: "Musketeer", expectedArena: 0},
		{name: "Baby Dragon", cardName: "Baby Dragon", expectedArena: 0},
		{name: "Fireball", cardName: "Fireball", expectedArena: 0},
		{name: "Zap", cardName: "Zap", expectedArena: 0},
		{name: "Cannon", cardName: "Cannon", expectedArena: 0},

		// Arena 1
		{name: "Spear Goblins", cardName: "Spear Goblins", expectedArena: 1},
		{name: "Giant Skeleton", cardName: "Giant Skeleton", expectedArena: 1},
		{name: "Tombstone", cardName: "Tombstone", expectedArena: 1},

		// Arena 2
		{name: "Hog Rider", cardName: "Hog Rider", expectedArena: 2},
		{name: "Minion Horde", cardName: "Minion Horde", expectedArena: 2},

		// Arena 3
		{name: "Ice Wizard", cardName: "Ice Wizard", expectedArena: 3},
		{name: "Royal Giant", cardName: "Royal Giant", expectedArena: 3},
		{name: "Rocket", cardName: "Rocket", expectedArena: 3},
		{name: "Goblin Barrel", cardName: "Goblin Barrel", expectedArena: 3},

		// Arena 4
		{name: "Princess", cardName: "Princess", expectedArena: 4},
		{name: "Dark Prince", cardName: "Dark Prince", expectedArena: 4},
		{name: "Ice Spirit", cardName: "Ice Spirit", expectedArena: 6},

		// Arena 5
		{name: "Lava Hound", cardName: "Lava Hound", expectedArena: 5},
		{name: "Poison", cardName: "Poison", expectedArena: 5},

		// Arena 6
		{name: "Miner", cardName: "Miner", expectedArena: 6},
		{name: "The Log", cardName: "The Log", expectedArena: 6},

		// Arena 7
		{name: "Lumberjack", cardName: "Lumberjack", expectedArena: 7},
		{name: "Battle Ram", cardName: "Battle Ram", expectedArena: 7},
		{name: "Tornado", cardName: "Tornado", expectedArena: 7},

		// Arena 8
		{name: "Ice Golem", cardName: "Ice Golem", expectedArena: 8},
		{name: "Mega Minion", cardName: "Mega Minion", expectedArena: 8},
		{name: "Goblin Gang", cardName: "Goblin Gang", expectedArena: 8},
		{name: "Electro Wizard", cardName: "Electro Wizard", expectedArena: 8},

		// Arena 9
		{name: "Elite Barbarians", cardName: "Elite Barbarians", expectedArena: 9},
		{name: "Hunter", cardName: "Hunter", expectedArena: 9},
		{name: "Executioner", cardName: "Executioner", expectedArena: 9},
		{name: "Bandit", cardName: "Bandit", expectedArena: 9},

		// Arena 10
		{name: "Royal Recruits", cardName: "Royal Recruits", expectedArena: 10},
		{name: "Night Witch", cardName: "Night Witch", expectedArena: 10},
		{name: "Bats", cardName: "Bats", expectedArena: 10},
		{name: "Royal Ghost", cardName: "Royal Ghost", expectedArena: 10},

		// Arena 11
		{name: "Ram Rider", cardName: "Ram Rider", expectedArena: 11},
		{name: "Mega Knight", cardName: "Mega Knight", expectedArena: 11},

		// Arena 12
		{name: "Flying Machine", cardName: "Flying Machine", expectedArena: 12},
		{name: "Royal Hogs", cardName: "Royal Hogs", expectedArena: 12},
		{name: "Heal Spirit", cardName: "Heal Spirit", expectedArena: 12},

		// Arena 13+
		{name: "Fisherman", cardName: "Fisherman", expectedArena: 13},
		{name: "Magic Archer", cardName: "Magic Archer", expectedArena: 13},
		{name: "Electro Dragon", cardName: "Electro Dragon", expectedArena: 13},
		{name: "Giant Snowball", cardName: "Giant Snowball", expectedArena: 13},

		// Arena 14+
		{name: "Mighty Miner", cardName: "Mighty Miner", expectedArena: 14},
		{name: "Elixir Golem", cardName: "Elixir Golem", expectedArena: 14},
		{name: "Battle Healer", cardName: "Battle Healer", expectedArena: 14},

		// Arena 15+
		{name: "Skeleton King", cardName: "Skeleton King", expectedArena: 15},
		{name: "Archer Queen", cardName: "Archer Queen", expectedArena: 15},
		{name: "Mother Witch", cardName: "Mother Witch", expectedArena: 15},
		{name: "Electro Giant", cardName: "Electro Giant", expectedArena: 15},

		// Unknown card defaults to 0
		{name: "Unknown card", cardName: "Unknown Card", expectedArena: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCardUnlockArena(tt.cardName)
			if result != tt.expectedArena {
				t.Errorf("getCardUnlockArena(%q) = %d, want %d",
					tt.cardName, result, tt.expectedArena)
			}
		})
	}
}

func TestGetArenaName(t *testing.T) {
	tests := []struct {
		name     string
		arenaNum int
		expected string
	}{
		{name: "Training Camp", arenaNum: 0, expected: "Training Camp"},
		{name: "Goblin Stadium", arenaNum: 1, expected: "Goblin Stadium"},
		{name: "Bone Pit", arenaNum: 2, expected: "Bone Pit"},
		{name: "Barbarian Bowl", arenaNum: 3, expected: "Barbarian Bowl"},
		{name: "P.E.K.K.A's Playhouse", arenaNum: 4, expected: "P.E.K.K.A's Playhouse"},
		{name: "Spell Valley", arenaNum: 5, expected: "Spell Valley"},
		{name: "Builder's Workshop", arenaNum: 6, expected: "Builder's Workshop"},
		{name: "Royal Arena", arenaNum: 7, expected: "Royal Arena"},
		{name: "Frozen Peak", arenaNum: 8, expected: "Frozen Peak"},
		{name: "Jungle Arena", arenaNum: 9, expected: "Jungle Arena"},
		{name: "Hog Mountain", arenaNum: 10, expected: "Hog Mountain"},
		{name: "Electro Valley", arenaNum: 11, expected: "Electro Valley"},
		{name: "Spooky Town", arenaNum: 12, expected: "Spooky Town"},
		{name: "Rascal's Hideout", arenaNum: 13, expected: "Rascal's Hideout"},
		{name: "Serenity Peak", arenaNum: 14, expected: "Serenity Peak"},
		{name: "Legendary Arena", arenaNum: 15, expected: "Legendary Arena"},
		{name: "Unknown arena", arenaNum: 20, expected: "Arena 20"},
		{name: "Invalid arena", arenaNum: -1, expected: "Arena -1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getArenaName(tt.arenaNum)
			if result != tt.expected {
				t.Errorf("getArenaName(%d) = %q, want %q", tt.arenaNum, result, tt.expected)
			}
		})
	}
}

func TestJoinCardNames(t *testing.T) {
	tests := []struct {
		name     string
		cards    []string
		expected string
	}{
		{
			name:     "Empty",
			cards:    []string{},
			expected: "",
		},
		{
			name:     "Single card",
			cards:    []string{"Knight"},
			expected: "Knight",
		},
		{
			name:     "Two cards",
			cards:    []string{"Knight", "Valkyrie"},
			expected: "Knight, Valkyrie",
		},
		{
			name:     "Three cards",
			cards:    []string{"Knight", "Valkyrie", "Ice Golem"},
			expected: "Knight, Valkyrie, Ice Golem",
		},
		{
			name:     "Many cards",
			cards:    []string{"A", "B", "C", "D"},
			expected: "A, B, C, D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinCardNames(tt.cards)
			if result != tt.expected {
				t.Errorf("joinCardNames() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatMissingCardsReport(t *testing.T) {
	tests := []struct {
		name              string
		analysis          *MissingCardsAnalysis
		expectContains    []string
		expectNotContains []string
	}{
		{
			name: "Playable deck",
			analysis: &MissingCardsAnalysis{
				Deck:           []string{"Knight", "Hog Rider", "Fireball"},
				MissingCards:   []MissingCard{},
				MissingCount:   0,
				AvailableCount: 3,
				IsPlayable:     true,
			},
			expectContains: []string{"✓", "All cards", "available"},
		},
		{
			name: "Missing cards",
			analysis: &MissingCardsAnalysis{
				Deck: []string{"Knight", "Mega Knight", "Fireball"},
				MissingCards: []MissingCard{
					{
						Name:             "Mega Knight",
						Rarity:           "Legendary",
						UnlockArena:      11,
						UnlockArenaName:  "Electro Valley",
						IsLocked:         false,
						AlternativeCards: []string{"P.E.K.K.A", "Golem"},
					},
				},
				MissingCount:   1,
				AvailableCount: 2,
				IsPlayable:     false,
			},
			expectContains: []string{
				"Missing Cards Analysis",
				"2/3",
				"Mega Knight",
				"Legendary",
				"Electro Valley",
				"Unlocked",
				"P.E.K.K.A",
			},
		},
		{
			name: "Locked card",
			analysis: &MissingCardsAnalysis{
				Deck: []string{"Knight", "Mega Knight", "Fireball"},
				MissingCards: []MissingCard{
					{
						Name:             "Mega Knight",
						Rarity:           "Legendary",
						UnlockArena:      11,
						UnlockArenaName:  "Electro Valley",
						IsLocked:         true, // Player is in Arena 5
						AlternativeCards: []string{},
					},
				},
				MissingCount:   1,
				AvailableCount: 1,
				IsPlayable:     false,
			},
			expectContains: []string{
				"🔒",
				"LOCKED",
				"Progress to Arena",
			},
		},
		{
			name: "No alternatives",
			analysis: &MissingCardsAnalysis{
				Deck: []string{"Knight", "Mega Knight", "Fireball"},
				MissingCards: []MissingCard{
					{
						Name:             "Mega Knight",
						Rarity:           "Legendary",
						UnlockArena:      11,
						UnlockArenaName:  "Electro Valley",
						IsLocked:         false,
						AlternativeCards: []string{},
					},
				},
				MissingCount:   1,
				AvailableCount: 2,
				IsPlayable:     false,
			},
			expectContains: []string{"None found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatMissingCardsReport(tt.analysis)

			if result == "" {
				t.Errorf("FormatMissingCardsReport() returned empty string")
			}

			for _, expected := range tt.expectContains {
				if !contains(result, expected) {
					t.Errorf("FormatMissingCardsReport() missing expected string %q", expected)
				}
			}

			for _, notExpected := range tt.expectNotContains {
				if contains(result, notExpected) {
					t.Errorf("FormatMissingCardsReport() contains unexpected string %q", notExpected)
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestMissingCardStruct(t *testing.T) {
	// Test that MissingCard can be properly constructed
	missing := MissingCard{
		Name:             "Mega Knight",
		Rarity:           "Legendary",
		UnlockArena:      11,
		UnlockArenaName:  "Electro Valley",
		AlternativeCards: []string{"P.E.K.K.A", "Golem"},
		IsLocked:         false,
	}

	if missing.Name != "Mega Knight" {
		t.Errorf("MissingCard.Name = %q, want %q", missing.Name, "Mega Knight")
	}
	if missing.UnlockArena != 11 {
		t.Errorf("MissingCard.UnlockArena = %d, want %d", missing.UnlockArena, 11)
	}
	if len(missing.AlternativeCards) != 2 {
		t.Errorf("MissingCard.AlternativeCards length = %d, want %d", len(missing.AlternativeCards), 2)
	}
}

func TestMissingCardsAnalysisStruct(t *testing.T) {
	// Test that MissingCardsAnalysis can be properly constructed
	analysis := MissingCardsAnalysis{
		Deck: []string{"Knight", "Hog Rider", "Fireball"},
		MissingCards: []MissingCard{
			{
				Name:   "Mega Knight",
				Rarity: "Legendary",
			},
		},
		MissingCount:   1,
		AvailableCount: 2,
		IsPlayable:     false,
		SuggestedReplacements: map[string][]string{
			"Mega Knight": {"P.E.K.K.A", "Golem"},
		},
	}

	if len(analysis.Deck) != 3 {
		t.Errorf("MissingCardsAnalysis.Deck length = %d, want %d", len(analysis.Deck), 3)
	}
	if analysis.MissingCount != 1 {
		t.Errorf("MissingCardsAnalysis.MissingCount = %d, want %d", analysis.MissingCount, 1)
	}
	if len(analysis.SuggestedReplacements) != 1 {
		t.Errorf("MissingCardsAnalysis.SuggestedReplacements length = %d, want %d",
			len(analysis.SuggestedReplacements), 1)
	}
}

func TestFindOwnedAlternatives(t *testing.T) {
	missingCard := makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3)

	tests := []struct {
		name                 string
		missingCard          deck.CardCandidate
		playerCollection     map[string]bool
		minAlternatives      int
		maxAlternatives      int
		expectedAlternatives []string
	}{
		{
			name:        "Player owns some alternatives",
			missingCard: missingCard,
			playerCollection: map[string]bool{
				"Valkyrie":    true,
				"Ice Golem":   true,
				"Dark Prince": false, // Not owned
			},
			minAlternatives:      2,
			maxAlternatives:      2,
			expectedAlternatives: []string{"Valkyrie", "Ice Golem"},
		},
		{
			name:        "Player owns no alternatives",
			missingCard: missingCard,
			playerCollection: map[string]bool{
				"Hog Rider": true,
				"Fireball":  true,
			},
			minAlternatives:      0,
			maxAlternatives:      0,
			expectedAlternatives: []string{},
		},
		{
			name:                 "Nil player collection",
			missingCard:          missingCard,
			playerCollection:     nil,
			minAlternatives:      0,
			maxAlternatives:      0,
			expectedAlternatives: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findOwnedAlternatives(tt.missingCard, tt.playerCollection)

			if len(result) < tt.minAlternatives || len(result) > tt.maxAlternatives {
				t.Errorf("findOwnedAlternatives() returned %d alternatives, want between %d and %d",
					len(result), tt.minAlternatives, tt.maxAlternatives)
			}

			// Check expected alternatives are present
			if len(tt.expectedAlternatives) > 0 {
				for _, expected := range tt.expectedAlternatives {
					found := slices.Contains(result, expected)
					if !found {
						t.Errorf("findOwnedAlternatives() missing expected alternative %q", expected)
					}
				}
			}
		})
	}
}
