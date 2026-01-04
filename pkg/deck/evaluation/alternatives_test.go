package evaluation

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestGenerateAlternatives(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	tests := []struct {
		name                   string
		deckCards              []deck.CardCandidate
		maxSuggestions         int
		playerCards            map[string]bool
		expectNil              bool
		minSuggestions         int
		expectedMaxSuggestions int
	}{
		{
			name: "Hog Cycle deck with alternatives",
			deckCards: []deck.CardCandidate{
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
				makeCard("Zap", deck.RoleSpellSmall, 11, 11, "Common", 2),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Ice Golem", deck.RoleCycle, 11, 11, "Rare", 2),
			},
			maxSuggestions:         5,
			playerCards:            nil,
			expectNil:              false,
			minSuggestions:         0,
			expectedMaxSuggestions: 5,
		},
		{
			name:                   "Empty deck",
			deckCards:              []deck.CardCandidate{},
			maxSuggestions:         5,
			playerCards:            nil,
			expectNil:              false,
			minSuggestions:         0,
			expectedMaxSuggestions: 0,
		},
		{
			name: "Zero max suggestions (defaults to 5)",
			deckCards: []deck.CardCandidate{
				makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
				makeCard("Archers", deck.RoleSupport, 11, 11, "Common", 3),
			},
			maxSuggestions:         0,
			playerCards:            nil,
			expectNil:              false,
			minSuggestions:         0,
			expectedMaxSuggestions: 5,
		},
		{
			name: "With player card restrictions",
			deckCards: []deck.CardCandidate{
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
			},
			maxSuggestions: 3,
			playerCards: map[string]bool{
				"Poison":    true,
				"Rocket":    true,
				"Lightning": true,
			},
			expectNil:              false,
			minSuggestions:         0,
			expectedMaxSuggestions: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateAlternatives(tt.deckCards, synergyDB, tt.maxSuggestions, tt.playerCards)

			if result == nil {
				if !tt.expectNil {
					t.Errorf("GenerateAlternatives() returned nil, expected non-nil")
				}
				return
			}

			if tt.expectNil {
				t.Errorf("GenerateAlternatives() returned non-nil, expected nil")
				return
			}

			// Check structure
			if result.OriginalDeck == nil {
				t.Errorf("GenerateAlternatives() OriginalDeck is nil")
			}

			if len(result.OriginalDeck) != len(tt.deckCards) {
				t.Errorf("GenerateAlternatives() OriginalDeck length = %d, want %d",
					len(result.OriginalDeck), len(tt.deckCards))
			}

			if result.OriginalScore == 0 && len(tt.deckCards) > 0 {
				t.Errorf("GenerateAlternatives() OriginalScore = 0, want > 0")
			}

			if result.Suggestions == nil {
				t.Errorf("GenerateAlternatives() Suggestions is nil")
			}

			if len(result.Suggestions) < tt.minSuggestions || len(result.Suggestions) > tt.expectedMaxSuggestions {
				t.Errorf("GenerateAlternatives() suggestions count = %d, want between %d and %d",
					len(result.Suggestions), tt.minSuggestions, tt.expectedMaxSuggestions)
			}

			// Check top suggestion
			if len(result.Suggestions) > 0 {
				if result.TopSuggestion == nil {
					t.Errorf("GenerateAlternatives() TopSuggestion is nil when suggestions exist")
				} else {
					// Verify top suggestion has required fields
					if result.TopSuggestion.OriginalCard == "" {
						t.Errorf("GenerateAlternatives() TopSuggestion.OriginalCard is empty")
					}
					if result.TopSuggestion.ReplacementCard == "" {
						t.Errorf("GenerateAlternatives() TopSuggestion.ReplacementCard is empty")
					}
				}
			}

			// Verify each suggestion
			for i, suggestion := range result.Suggestions {
				if suggestion.OriginalCard == "" {
					t.Errorf("GenerateAlternatives() suggestion[%d].OriginalCard is empty", i)
				}
				if suggestion.ReplacementCard == "" {
					t.Errorf("GenerateAlternatives() suggestion[%d].ReplacementCard is empty", i)
				}
				if len(suggestion.Deck) != len(tt.deckCards) {
					t.Errorf("GenerateAlternatives() suggestion[%d].Deck length = %d, want %d",
						i, len(suggestion.Deck), len(tt.deckCards))
				}
				if suggestion.ScoreDelta <= 0 {
					t.Errorf("GenerateAlternatives() suggestion[%d].ScoreDelta = %f, want > 0",
						i, suggestion.ScoreDelta)
				}
			}
		})
	}
}

func TestBuildRationale(t *testing.T) {
	tests := []struct {
		name         string
		original     deck.CardCandidate
		replacement  deck.CardCandidate
		originalEval EvaluationResult
		newEval      EvaluationResult
		minLength    int
	}{
		{
			name:        "Attack improvement",
			original:    makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
			replacement: makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
			originalEval: EvaluationResult{
				Attack:      CategoryScore{Score: 5.0},
				Defense:     CategoryScore{Score: 6.0},
				Synergy:     CategoryScore{Score: 5.0},
				Versatility: CategoryScore{Score: 5.0},
			},
			newEval: EvaluationResult{
				Attack:      CategoryScore{Score: 7.0}, // Improved
				Defense:     CategoryScore{Score: 6.0},
				Synergy:     CategoryScore{Score: 5.0},
				Versatility: CategoryScore{Score: 5.0},
			},
			minLength: 20,
		},
		{
			name:        "Multiple improvements",
			original:    makeCard("Zap", deck.RoleSpellSmall, 11, 11, "Common", 2),
			replacement: makeCard("The Log", deck.RoleSpellSmall, 11, 14, "Legendary", 2),
			originalEval: EvaluationResult{
				Attack:      CategoryScore{Score: 5.0},
				Defense:     CategoryScore{Score: 5.0},
				Synergy:     CategoryScore{Score: 5.0},
				Versatility: CategoryScore{Score: 5.0},
			},
			newEval: EvaluationResult{
				Attack:      CategoryScore{Score: 6.0},
				Defense:     CategoryScore{Score: 6.0},
				Synergy:     CategoryScore{Score: 7.0},
				Versatility: CategoryScore{Score: 6.0},
			},
			minLength: 40,
		},
		{
			name:        "No specific improvements",
			original:    makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
			replacement: makeCard("Ice Golem", deck.RoleCycle, 11, 11, "Rare", 2),
			originalEval: EvaluationResult{
				Attack:      CategoryScore{Score: 5.0},
				Defense:     CategoryScore{Score: 6.0},
				Synergy:     CategoryScore{Score: 5.0},
				Versatility: CategoryScore{Score: 5.0},
			},
			newEval: EvaluationResult{
				Attack:      CategoryScore{Score: 5.0},
				Defense:     CategoryScore{Score: 6.0},
				Synergy:     CategoryScore{Score: 5.0},
				Versatility: CategoryScore{Score: 5.0},
			},
			minLength: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rationale := buildRationale(tt.original, tt.replacement, tt.originalEval, tt.newEval)

			if len(rationale) < tt.minLength {
				t.Errorf("buildRationale() length = %d, want >= %d", len(rationale), tt.minLength)
			}

			if rationale == "" {
				t.Errorf("buildRationale() returned empty string")
			}
		})
	}
}

func TestJoinImprovements(t *testing.T) {
	tests := []struct {
		name         string
		improvements []string
		expected     string
	}{
		{
			name:         "Empty",
			improvements: []string{},
			expected:     "",
		},
		{
			name:         "Single item",
			improvements: []string{"stronger attack"},
			expected:     "stronger attack",
		},
		{
			name:         "Two items",
			improvements: []string{"stronger attack", "better defense"},
			expected:     "stronger attack and better defense",
		},
		{
			name:         "Three items",
			improvements: []string{"stronger attack", "better defense", "improved synergy"},
			expected:     "stronger attack, better defenseand improved synergy",
		},
		{
			name:         "Four items",
			improvements: []string{"attack", "defense", "synergy", "versatility"},
			expected:     "attack, defense, synergyand versatility",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinImprovements(tt.improvements)
			if result != tt.expected {
				t.Errorf("joinImprovements() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDetermineImpact(t *testing.T) {
	tests := []struct {
		name     string
		delta    float64
		expected string
	}{
		{
			name:     "Major positive",
			delta:    2.5,
			expected: "Major Improvement",
		},
		{
			name:     "Major negative",
			delta:    -2.5,
			expected: "Major Improvement",
		},
		{
			name:     "Significant positive",
			delta:    1.2,
			expected: "Significant Improvement",
		},
		{
			name:     "Significant negative",
			delta:    -1.2,
			expected: "Significant Improvement",
		},
		{
			name:     "Moderate positive",
			delta:    0.6,
			expected: "Moderate Improvement",
		},
		{
			name:     "Moderate negative",
			delta:    -0.6,
			expected: "Moderate Improvement",
		},
		{
			name:     "Minor positive",
			delta:    0.3,
			expected: "Minor Improvement",
		},
		{
			name:     "Minor negative",
			delta:    -0.3,
			expected: "Minor Improvement",
		},
		{
			name:     "Zero",
			delta:    0.0,
			expected: "Minor Improvement",
		},
		{
			name:     "Edge case major",
			delta:    2.0,
			expected: "Major Improvement",
		},
		{
			name:     "Edge case significant",
			delta:    1.0,
			expected: "Significant Improvement",
		},
		{
			name:     "Edge case moderate",
			delta:    0.5,
			expected: "Moderate Improvement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineImpact(tt.delta)
			if result != tt.expected {
				t.Errorf("determineImpact(%f) = %q, want %q", tt.delta, result, tt.expected)
			}
		})
	}
}

func TestInferElixirForCard(t *testing.T) {
	tests := []struct {
		name     string
		cardName string
		expected int
	}{
		{"Knight", "Knight", 3},
		{"Valkyrie", "Valkyrie", 4},
		{"Hog Rider", "Hog Rider", 4},
		{"Fireball", "Fireball", 4},
		{"Zap", "Zap", 2},
		{"The Log", "The Log", 2},
		{"Rocket", "Rocket", 6},
		{"Lightning", "Lightning", 6},
		{"Musketeer", "Musketeer", 4},
		{"Ice Spirit", "Ice Spirit", 1},
		{"Golem", "Golem", 8},
		{"Lava Hound", "Lava Hound", 7},
		{"P.E.K.K.A", "P.E.K.K.A", 7},
		{"Mega Knight", "Mega Knight", 7},
		{"Electro Wizard", "Electro Wizard", 4},
		{"Unknown card", "Unknown Card XYZ", 3}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferElixirForCard(tt.cardName)
			if result != tt.expected {
				t.Errorf("inferElixirForCard(%q) = %d, want %d", tt.cardName, result, tt.expected)
			}
		})
	}
}

func TestGetSimilarCards(t *testing.T) {
	tests := []struct {
		name        string
		card        deck.CardCandidate
		minCount    int
		maxCount    int
		expectNames []string
	}{
		{
			name:        "Knight has known alternatives",
			card:        makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
			minCount:    3,
			maxCount:    3,
			expectNames: []string{"Valkyrie", "Ice Golem", "Dark Prince"},
		},
		{
			name:        "Hog Rider alternatives",
			card:        makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
			minCount:    3,
			maxCount:    3,
			expectNames: []string{"Ram Rider", "Battle Ram", "Royal Hogs"},
		},
		{
			name:     "Unknown card has no alternatives",
			card:     makeCard("Unknown Card", deck.RoleSupport, 11, 11, "Common", 3),
			minCount: 0,
			maxCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSimilarCards(tt.card)

			if len(result) < tt.minCount || len(result) > tt.maxCount {
				t.Errorf("getSimilarCards() returned %d cards, want between %d and %d",
					len(result), tt.minCount, tt.maxCount)
			}

			// Check expected names are present
			if len(tt.expectNames) > 0 {
				resultNames := make([]string, len(result))
				for i, card := range result {
					resultNames[i] = card.Name
				}

				for _, expectedName := range tt.expectNames {
					found := false
					for _, name := range resultNames {
						if name == expectedName {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("getSimilarCards() missing expected card %q", expectedName)
					}
				}
			}

			// Check that cards preserve the role
			for _, card := range result {
				if card.Role == nil {
					t.Errorf("getSimilarCards() returned card with nil Role")
				}
			}
		})
	}
}

func TestFindReplacements(t *testing.T) {
	deckCards := []deck.CardCandidate{
		makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
		makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
		makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
		makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
	}

	tests := []struct {
		name         string
		originalCard deck.CardCandidate
		playerCards  map[string]bool
		minCount     int
		maxCount     int
		exclude      []string // Cards that should NOT be in replacements
	}{
		{
			name:         "Knight replacements without player restrictions",
			originalCard: makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
			playerCards:  nil,
			minCount:     3,
			maxCount:     3,
			exclude:      []string{"Knight", "Hog Rider", "Fireball", "Musketeer"},
		},
		{
			name:         "Knight replacements with player restrictions",
			originalCard: makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
			playerCards: map[string]bool{
				"Valkyrie":    true,
				"Ice Golem":   false,
				"Dark Prince": true,
			},
			minCount: 2,
			maxCount: 2,
			exclude:  []string{"Knight", "Hog Rider", "Fireball", "Musketeer", "Ice Golem"},
		},
		{
			name:         "Unknown card has no replacements",
			originalCard: makeCard("Unknown", deck.RoleSupport, 11, 11, "Common", 3),
			playerCards:  nil,
			minCount:     0,
			maxCount:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findReplacements(tt.originalCard, deckCards, tt.playerCards)

			if len(result) < tt.minCount || len(result) > tt.maxCount {
				t.Errorf("findReplacements() returned %d cards, want between %d and %d",
					len(result), tt.minCount, tt.maxCount)
			}

			// Check excluded cards are not present
			resultNames := make([]string, len(result))
			for i, card := range result {
				resultNames[i] = card.Name
			}

			for _, excluded := range tt.exclude {
				for _, name := range resultNames {
					if name == excluded {
						t.Errorf("findReplacements() returned excluded card %q", excluded)
					}
				}
			}
		})
	}
}

func TestAlternativeDeckStruct(t *testing.T) {
	// Test that AlternativeDeck can be properly constructed
	alt := AlternativeDeck{
		OriginalCard:    "Knight",
		ReplacementCard: "Valkyrie",
		Rationale:       "Better defense and splash damage",
		Deck:            []string{"Hog Rider", "Valkyrie", "Fireball", "Zap", "Ice Spirit", "Skeletons", "Cannon", "Ice Golem"},
		OriginalScore:   6.5,
		NewScore:        7.2,
		ScoreDelta:      0.7,
		Impact:          "Moderate Improvement",
	}

	if alt.OriginalCard != "Knight" {
		t.Errorf("AlternativeDeck.OriginalCard = %q, want %q", alt.OriginalCard, "Knight")
	}
	if alt.ScoreDelta != 0.7 {
		t.Errorf("AlternativeDeck.ScoreDelta = %f, want %f", alt.ScoreDelta, 0.7)
	}
	if len(alt.Deck) != 8 {
		t.Errorf("AlternativeDeck.Deck length = %d, want %d", len(alt.Deck), 8)
	}
}

func TestAlternativeSuggestionsStruct(t *testing.T) {
	// Test that AlternativeSuggestions can be properly constructed
	suggestions := AlternativeSuggestions{
		OriginalDeck:  []string{"Knight", "Hog Rider", "Fireball", "Zap", "Ice Spirit", "Skeletons", "Cannon", "Musketeer"},
		OriginalScore: 6.5,
		Suggestions: []AlternativeDeck{
			{
				OriginalCard:    "Knight",
				ReplacementCard: "Valkyrie",
				ScoreDelta:      0.7,
			},
		},
		TopSuggestion: &AlternativeDeck{
			OriginalCard:    "Knight",
			ReplacementCard: "Valkyrie",
			ScoreDelta:      0.7,
		},
	}

	if len(suggestions.OriginalDeck) != 8 {
		t.Errorf("AlternativeSuggestions.OriginalDeck length = %d, want %d", len(suggestions.OriginalDeck), 8)
	}
	if len(suggestions.Suggestions) != 1 {
		t.Errorf("AlternativeSuggestions.Suggestions length = %d, want %d", len(suggestions.Suggestions), 1)
	}
	if suggestions.TopSuggestion == nil {
		t.Errorf("AlternativeSuggestions.TopSuggestion is nil")
	}
}
