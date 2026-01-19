package events

import (
	"testing"
	"time"
)

// Helper function to create a test event deck
func createTestEventDeck(cards []string, wins int, losses int) EventDeck {
	deckCards := make([]CardInDeck, len(cards))
	totalElixir := 0
	for i, cardName := range cards {
		elixir := 3 // Default elixir cost
		deckCards[i] = CardInDeck{
			Name:       cardName,
			ID:         i,
			Level:      11,
			MaxLevel:   14,
			Rarity:     "Common",
			ElixirCost: elixir,
		}
		totalElixir += elixir
	}

	deck := EventDeck{
		EventID:   "test-event-123",
		PlayerTag: "#TEST123",
		EventName: "Test Challenge",
		EventType: EventTypeChallenge,
		StartTime: time.Now(),
		Deck: Deck{
			Cards:     deckCards,
			AvgElixir: float64(totalElixir) / float64(len(cards)),
		},
		Performance: EventPerformance{
			Wins:    wins,
			Losses:  losses,
			WinRate: float64(wins) / float64(wins+losses),
		},
	}

	return deck
}

func TestSuggestCardConstraints_EmptyDecks(t *testing.T) {
	decks := []EventDeck{}
	threshold := 50.0

	suggestions := SuggestCardConstraints(decks, threshold)

	if len(suggestions) != 0 {
		t.Errorf("Expected 0 suggestions for empty decks, got %d", len(suggestions))
	}
}

func TestSuggestCardConstraints_ThresholdFiltering(t *testing.T) {
	tests := []struct {
		name           string
		threshold      float64
		expectedCount  int
		expectedCards  []string
	}{
		{
			name:          "0% threshold - all cards",
			threshold:     0.0,
			expectedCount: 10, // Top 10 cards (limited by analyzeCardUsage)
			expectedCards: []string{"Skeleton Dragons", "Tornado", "Ice Wizard"}, // Only verify top cards
		},
		{
			name:          "30% threshold",
			threshold:     30.0,
			expectedCount: 3, // Cards in 2+ of 5 decks (40%+)
			expectedCards: []string{"Skeleton Dragons", "Tornado", "Ice Wizard"},
		},
		{
			name:          "50% threshold",
			threshold:     50.0,
			expectedCount: 2, // Cards in 3+ of 5 decks (60%+)
			expectedCards: []string{"Skeleton Dragons", "Tornado"},
		},
		{
			name:          "100% threshold",
			threshold:     100.0,
			expectedCount: 0, // No cards in all decks
			expectedCards: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create 5 test decks with varying card overlap
			decks := []EventDeck{
				// Deck 1: Skeleton Dragons, Tornado, Ice Wizard
				createTestEventDeck([]string{"Skeleton Dragons", "Tornado", "Ice Wizard", "Knight", "Fireball", "Zap", "Goblin Barrel", "Princess"}, 10, 2),
				// Deck 2: Skeleton Dragons, Tornado, Ice Wizard (high overlap)
				createTestEventDeck([]string{"Skeleton Dragons", "Tornado", "Ice Wizard", "Valkyrie", "Arrows", "Log", "Miner", "Bats"}, 9, 3),
				// Deck 3: Skeleton Dragons, Tornado (medium overlap)
				createTestEventDeck([]string{"Skeleton Dragons", "Tornado", "Musketeer", "Mini PEKKA", "Rocket", "Freeze", "Skeletons", "Ice Spirit"}, 8, 4),
				// Deck 4: Different cards
				createTestEventDeck([]string{"Hog Rider", "Executioner", "Mega Knight", "Electro Wizard", "Lightning", "Mirror", "Clone", "Heal Spirit"}, 7, 5),
				// Deck 5: Different cards
				createTestEventDeck([]string{"Giant", "Witch", "Balloon", "Lava Hound", "Baby Dragon", "Inferno Dragon", "Rage", "Elixir Collector"}, 6, 6),
			}

			suggestions := SuggestCardConstraints(decks, tt.threshold)

			if len(suggestions) != tt.expectedCount {
				t.Errorf("Expected %d suggestions, got %d", tt.expectedCount, len(suggestions))
			}

			// Verify expected cards are present (for non-zero thresholds)
			if tt.expectedCount > 0 {
				foundCards := make(map[string]bool)
				for _, suggestion := range suggestions {
					foundCards[suggestion.CardName] = true
				}

				for _, expectedCard := range tt.expectedCards {
					if !foundCards[expectedCard] {
						t.Errorf("Expected card '%s' not found in suggestions", expectedCard)
					}
				}
			}
		})
	}
}

func TestSuggestCardConstraints_PercentageCalculation(t *testing.T) {
	// Create 10 decks where "Skeleton Dragons" appears in 6 (60%)
	decks := make([]EventDeck, 10)
	for i := 0; i < 10; i++ {
		var cards []string
		if i < 6 {
			// First 6 decks contain Skeleton Dragons
			cards = []string{"Skeleton Dragons", "Knight", "Fireball", "Zap", "Tornado", "Ice Wizard", "Musketeer", "Mini PEKKA"}
		} else {
			// Last 4 decks don't contain Skeleton Dragons
			cards = []string{"Hog Rider", "Knight", "Fireball", "Zap", "Tornado", "Ice Wizard", "Musketeer", "Mini PEKKA"}
		}
		decks[i] = createTestEventDeck(cards, 10, 2)
	}

	threshold := 50.0
	suggestions := SuggestCardConstraints(decks, threshold)

	// Find Skeleton Dragons in suggestions
	var skeletonDragons *CardConstraintSuggestion
	for i := range suggestions {
		if suggestions[i].CardName == "Skeleton Dragons" {
			skeletonDragons = &suggestions[i]
			break
		}
	}

	if skeletonDragons == nil {
		t.Fatal("Expected to find Skeleton Dragons in suggestions")
	}

	if skeletonDragons.Appearances != 6 {
		t.Errorf("Expected 6 appearances, got %d", skeletonDragons.Appearances)
	}

	if skeletonDragons.TotalDecks != 10 {
		t.Errorf("Expected 10 total decks, got %d", skeletonDragons.TotalDecks)
	}

	expectedPercentage := 60.0
	if skeletonDragons.Percentage != expectedPercentage {
		t.Errorf("Expected %.1f%% percentage, got %.1f%%", expectedPercentage, skeletonDragons.Percentage)
	}
}

func TestSuggestCardConstraints_SortedByPercentage(t *testing.T) {
	// Create decks with varying card frequencies to ensure proper sorting
	// High freq (5/5 = 100%), Medium freq (3/5 = 60%), Low freq (1/5 = 20%)
	decks := []EventDeck{
		createTestEventDeck([]string{"HighFreqCard", "MediumFreqCard", "LowFreqCard", "Unique1A", "Unique1B", "Unique1C", "Unique1D", "Unique1E"}, 10, 2),
		createTestEventDeck([]string{"HighFreqCard", "MediumFreqCard", "Unique2A", "Unique2B", "Unique2C", "Unique2D", "Unique2E", "Unique2F"}, 9, 3),
		createTestEventDeck([]string{"HighFreqCard", "MediumFreqCard", "Unique3A", "Unique3B", "Unique3C", "Unique3D", "Unique3E", "Unique3F"}, 8, 4),
		createTestEventDeck([]string{"HighFreqCard", "Unique4A", "Unique4B", "Unique4C", "Unique4D", "Unique4E", "Unique4F", "Unique4G"}, 7, 5),
		createTestEventDeck([]string{"HighFreqCard", "Unique5A", "Unique5B", "Unique5C", "Unique5D", "Unique5E", "Unique5F", "Unique5G"}, 6, 6),
	}

	threshold := 0.0 // Get all cards
	suggestions := SuggestCardConstraints(decks, threshold)

	// Verify suggestions are sorted by percentage descending
	for i := 0; i < len(suggestions)-1; i++ {
		if suggestions[i].Percentage < suggestions[i+1].Percentage {
			t.Errorf("Suggestions not sorted correctly: %s (%.1f%%) should be before %s (%.1f%%)",
				suggestions[i].CardName, suggestions[i].Percentage,
				suggestions[i+1].CardName, suggestions[i+1].Percentage)
		}
	}

	// Verify HighFreqCard appears with 100% frequency
	foundHighFreq := false
	for _, suggestion := range suggestions {
		if suggestion.CardName == "HighFreqCard" {
			foundHighFreq = true
			if suggestion.Percentage != 100.0 {
				t.Errorf("Expected HighFreqCard to have 100%% frequency, got %.1f%%", suggestion.Percentage)
			}
			break
		}
	}
	if !foundHighFreq {
		t.Error("Expected to find HighFreqCard in suggestions")
	}

	// Verify MediumFreqCard appears with 60% frequency
	foundMediumFreq := false
	for _, suggestion := range suggestions {
		if suggestion.CardName == "MediumFreqCard" {
			foundMediumFreq = true
			if suggestion.Percentage != 60.0 {
				t.Errorf("Expected MediumFreqCard to have 60%% frequency, got %.1f%%", suggestion.Percentage)
			}
			break
		}
	}
	if !foundMediumFreq {
		t.Error("Expected to find MediumFreqCard in suggestions")
	}
}

func TestSuggestCardConstraints_NoSharedCards(t *testing.T) {
	// Create decks with completely different cards
	decks := []EventDeck{
		createTestEventDeck([]string{"Card 1A", "Card 1B", "Card 1C", "Card 1D", "Card 1E", "Card 1F", "Card 1G", "Card 1H"}, 10, 2),
		createTestEventDeck([]string{"Card 2A", "Card 2B", "Card 2C", "Card 2D", "Card 2E", "Card 2F", "Card 2G", "Card 2H"}, 9, 3),
		createTestEventDeck([]string{"Card 3A", "Card 3B", "Card 3C", "Card 3D", "Card 3E", "Card 3F", "Card 3G", "Card 3H"}, 8, 4),
	}

	threshold := 50.0
	suggestions := SuggestCardConstraints(decks, threshold)

	if len(suggestions) != 0 {
		t.Errorf("Expected 0 suggestions when decks share no cards with threshold 50%%, got %d", len(suggestions))
	}
}
