package deck

import (
	"testing"
)

func TestNewSynergyDatabase(t *testing.T) {
	db := NewSynergyDatabase()

	if db == nil {
		t.Fatal("NewSynergyDatabase returned nil")
	}

	if len(db.Pairs) == 0 {
		t.Error("Synergy database should have pairs")
	}

	if len(db.Categories) == 0 {
		t.Error("Synergy database should have categories")
	}

	// Verify categories are populated
	if len(db.Categories[SynergyTankSupport]) == 0 {
		t.Error("Tank support category should have synergies")
	}

	if len(db.Categories[SynergyBait]) == 0 {
		t.Error("Bait category should have synergies")
	}
}

func TestGetSynergy(t *testing.T) {
	db := NewSynergyDatabase()

	tests := []struct {
		name     string
		card1    string
		card2    string
		expected float64
	}{
		{"Giant + Witch", "Giant", "Witch", 0.9},
		{"Reversed order", "Witch", "Giant", 0.9},
		{"Goblin Barrel + Princess", "Goblin Barrel", "Princess", 0.95},
		{"No synergy", "Giant", "Unknown Card", 0.0},
		{"Self synergy", "Giant", "Giant", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.GetSynergy(tt.card1, tt.card2)
			if result != tt.expected {
				t.Errorf("GetSynergy(%s, %s) = %.2f, want %.2f", tt.card1, tt.card2, result, tt.expected)
			}
		})
	}
}

func TestGetSynergyPair(t *testing.T) {
	db := NewSynergyDatabase()

	tests := []struct {
		name       string
		card1      string
		card2      string
		shouldFind bool
	}{
		{"Existing pair", "Giant", "Witch", true},
		{"Reversed order", "Witch", "Giant", true},
		{"Non-existent", "Giant", "Unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair := db.GetSynergyPair(tt.card1, tt.card2)
			if tt.shouldFind && pair == nil {
				t.Error("Expected to find synergy pair")
			}
			if !tt.shouldFind && pair != nil {
				t.Error("Expected not to find synergy pair")
			}
		})
	}
}

func TestAnalyzeDeckSynergy(t *testing.T) {
	db := NewSynergyDatabase()

	tests := []struct {
		name          string
		deck          []string
		expectScore   bool
		expectSynergy int // Minimum expected synergies
	}{
		{
			name:          "Log bait deck",
			deck:          []string{"Goblin Barrel", "Princess", "Goblin Gang", "Log", "Cannon", "Ice Spirit", "Knight", "Rocket"},
			expectScore:   true,
			expectSynergy: 2, // At least Goblin Barrel + Princess, Goblin Barrel + Goblin Gang
		},
		{
			name:          "Giant beatdown",
			deck:          []string{"Giant", "Witch", "Musketeer", "Fireball", "Zap", "Cannon", "Ice Spirit", "Skeletons"},
			expectScore:   true,
			expectSynergy: 2, // Giant + Witch, Giant + Musketeer
		},
		{
			name:          "No synergies",
			deck:          []string{"Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7", "Card8"},
			expectScore:   false,
			expectSynergy: 0,
		},
		{
			name:          "Empty deck",
			deck:          []string{},
			expectScore:   false,
			expectSynergy: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := db.AnalyzeDeckSynergy(tt.deck)

			if analysis == nil {
				t.Fatal("AnalyzeDeckSynergy returned nil")
			}

			if tt.expectScore && analysis.TotalScore == 0 {
				t.Error("Expected non-zero synergy score")
			}

			if !tt.expectScore && analysis.TotalScore > 0 {
				t.Error("Expected zero synergy score")
			}

			if len(analysis.TopSynergies) < tt.expectSynergy {
				t.Errorf("Expected at least %d synergies, got %d", tt.expectSynergy, len(analysis.TopSynergies))
			}

			// Verify category scores
			if tt.expectScore && len(analysis.CategoryScores) == 0 {
				t.Error("Expected category scores to be populated")
			}
		})
	}
}

func TestAnalyzeDeckSynergy_TopSynergies(t *testing.T) {
	db := NewSynergyDatabase()

	// Deck with multiple strong synergies
	deck := []string{"Giant", "Witch", "Golem", "Night Witch", "Lava Hound", "Balloon", "Princess", "Goblin Barrel"}

	analysis := db.AnalyzeDeckSynergy(deck)

	if len(analysis.TopSynergies) == 0 {
		t.Error("Expected top synergies to be populated")
	}

	// Verify synergies are sorted by score (descending)
	for i := 0; i < len(analysis.TopSynergies)-1; i++ {
		if analysis.TopSynergies[i].Score < analysis.TopSynergies[i+1].Score {
			t.Error("Top synergies should be sorted by score descending")
		}
	}

	// Verify top synergies contain high-scoring pairs
	foundLavaLoon := false
	for _, syn := range analysis.TopSynergies {
		if (syn.Card1 == "Lava Hound" && syn.Card2 == "Balloon") ||
			(syn.Card1 == "Balloon" && syn.Card2 == "Lava Hound") {
			foundLavaLoon = true
			if syn.Score != 0.95 {
				t.Errorf("LavaLoon synergy should be 0.95, got %.2f", syn.Score)
			}
		}
	}

	if !foundLavaLoon {
		t.Error("Expected to find LavaLoon synergy in top synergies")
	}
}

func TestAnalyzeDeckSynergy_MissingSynergies(t *testing.T) {
	db := NewSynergyDatabase()

	// Deck with some cards that have no synergies
	deck := []string{"Giant", "Witch", "UnknownCard1", "UnknownCard2"}

	analysis := db.AnalyzeDeckSynergy(deck)

	if len(analysis.MissingSynergies) != 2 {
		t.Errorf("Expected 2 cards with missing synergies, got %d", len(analysis.MissingSynergies))
	}

	// Verify the missing cards are correct
	expectedMissing := map[string]bool{"UnknownCard1": true, "UnknownCard2": true}
	for _, card := range analysis.MissingSynergies {
		if !expectedMissing[card] {
			t.Errorf("Unexpected card in missing synergies: %s", card)
		}
	}
}

func TestSuggestSynergyCards(t *testing.T) {
	db := NewSynergyDatabase()

	roleSupport := RoleSupport

	currentDeck := []string{"Giant", "Musketeer"}
	available := []*CardCandidate{
		{Name: "Witch", Role: &roleSupport, Score: 0.8},
		{Name: "Sparky", Role: &roleSupport, Score: 0.75},
		{Name: "Unknown Card", Role: &roleSupport, Score: 0.7},
		{Name: "Giant", Role: &roleSupport, Score: 0.9}, // Already in deck
	}

	recommendations := db.SuggestSynergyCards(currentDeck, available)

	if len(recommendations) == 0 {
		t.Error("Expected synergy recommendations")
	}

	// Verify Giant is not recommended (already in deck)
	for _, rec := range recommendations {
		if rec.CardName == "Giant" {
			t.Error("Should not recommend cards already in deck")
		}
	}

	// Verify Witch is recommended (high synergy with Giant)
	foundWitch := false
	for _, rec := range recommendations {
		if rec.CardName == "Witch" {
			foundWitch = true
			if rec.SynergyScore == 0 {
				t.Error("Witch should have non-zero synergy score")
			}
			if len(rec.Synergies) == 0 {
				t.Error("Witch recommendation should include synergy details")
			}
		}
	}

	if !foundWitch {
		t.Error("Expected Witch to be recommended (synergizes with Giant)")
	}

	// Verify recommendations are sorted by score
	for i := 0; i < len(recommendations)-1; i++ {
		if recommendations[i].SynergyScore < recommendations[i+1].SynergyScore {
			t.Error("Recommendations should be sorted by synergy score descending")
		}
	}
}

func TestSuggestSynergyCards_EmptyInputs(t *testing.T) {
	db := NewSynergyDatabase()

	roleSupport := RoleSupport

	t.Run("Empty deck", func(t *testing.T) {
		available := []*CardCandidate{
			{Name: "Witch", Role: &roleSupport, Score: 0.8},
		}
		recommendations := db.SuggestSynergyCards([]string{}, available)
		if recommendations != nil {
			t.Error("Expected nil for empty deck")
		}
	})

	t.Run("Empty available", func(t *testing.T) {
		currentDeck := []string{"Giant"}
		recommendations := db.SuggestSynergyCards(currentDeck, []*CardCandidate{})
		if recommendations != nil {
			t.Error("Expected nil for empty available cards")
		}
	})
}

func TestGetCategoryDescription(t *testing.T) {
	tests := []struct {
		category SynergyCategory
		expected string
	}{
		{SynergyTankSupport, "Tank + Support"},
		{SynergyBait, "Spell Bait"},
		{SynergySpellCombo, "Spell Combo"},
		{SynergyWinCondition, "Win Condition"},
		{SynergyDefensive, "Defensive"},
		{SynergyCycle, "Cycle"},
		{SynergyBridgeSpam, "Bridge Spam"},
		{"UnknownCategory", "UnknownCategory"}, // Unknown category returns itself
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			result := GetCategoryDescription(tt.category)
			if result != tt.expected {
				t.Errorf("GetCategoryDescription(%s) = %s, want %s", tt.category, result, tt.expected)
			}
		})
	}
}

func TestSynergyCategories(t *testing.T) {
	db := NewSynergyDatabase()

	// Verify each category has at least one synergy
	expectedCategories := []SynergyCategory{
		SynergyTankSupport,
		SynergyBait,
		SynergySpellCombo,
		SynergyWinCondition,
		SynergyDefensive,
		SynergyCycle,
		SynergyBridgeSpam,
	}

	for _, cat := range expectedCategories {
		if len(db.Categories[cat]) == 0 {
			t.Errorf("Category %s should have at least one synergy", cat)
		}
	}
}
