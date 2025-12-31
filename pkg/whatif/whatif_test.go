package whatif

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestParseCardUpgrade(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		want    CardUpgrade
		wantErr bool
	}{
		{
			name: "valid format with to level only",
			spec: "Knight:15",
			want: CardUpgrade{
				CardName:  "Knight",
				FromLevel: 0,
				ToLevel:   15,
				GoldCost:  0,
			},
			wantErr: false,
		},
		{
			name: "valid format with from and to levels",
			spec: "Archers:9:15",
			want: CardUpgrade{
				CardName:  "Archers",
				FromLevel: 9,
				ToLevel:   15,
				GoldCost:  0,
			},
			wantErr: false,
		},
		{
			name:    "invalid format - missing level",
			spec:    "Knight",
			wantErr: true,
		},
		{
			name:    "invalid format - to level less than from level",
			spec:    "Knight:15:9",
			wantErr: true,
		},
		{
			name:    "invalid format - negative from level",
			spec:    "Knight:-1:15",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCardUpgrade(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCardUpgrade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.CardName != tt.want.CardName || got.FromLevel != tt.want.FromLevel || got.ToLevel != tt.want.ToLevel {
					t.Errorf("ParseCardUpgrade() = %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func TestCalculateUpgradeCost(t *testing.T) {
	tests := []struct {
		name      string
		fromLevel int
		toLevel   int
		rarity    string
		want      int
	}{
		{
			name:      "common single level",
			fromLevel: 9,
			toLevel:   10,
			rarity:    "Common",
			want:      100,
		},
		{
			name:      "common multiple levels",
			fromLevel: 9,
			toLevel:   15,
			rarity:    "Common",
			want:      600,
		},
		{
			name:      "rare single level",
			fromLevel: 9,
			toLevel:   10,
			rarity:    "Rare",
			want:      1000,
		},
		{
			name:      "epic multiple levels",
			fromLevel: 5,
			toLevel:   10,
			rarity:    "Epic",
			want:      15000,
		},
		{
			name:      "legendary single level",
			fromLevel: 11,
			toLevel:   12,
			rarity:    "Legendary",
			want:      40000,
		},
		{
			name:      "champion multiple levels",
			fromLevel: 11,
			toLevel:   15,
			rarity:    "Champion",
			want:      200000,
		},
		{
			name:      "unknown rarity defaults to 1000",
			fromLevel: 9,
			toLevel:   10,
			rarity:    "Unknown",
			want:      1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateUpgradeCost(tt.fromLevel, tt.toLevel, tt.rarity)
			if got != tt.want {
				t.Errorf("calculateUpgradeCost() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestApplyUpgradesToCardLevels(t *testing.T) {
	analyzer := &WhatIfAnalyzer{}

	baseCardLevels := map[string]deck.CardLevelData{
		"Knight": {
			Level:    9,
			MaxLevel: 16,
			Rarity:   "Common",
			Elixir:   3,
		},
		"Archers": {
			Level:             10,
			MaxLevel:          16,
			Rarity:            "Common",
			Elixir:            3,
			EvolutionLevel:    1,
			MaxEvolutionLevel: 3,
		},
		"Fireball": {
			Level:    11,
			MaxLevel: 16,
			Rarity:   "Rare",
			Elixir:   4,
		},
	}

	t.Run("infer from level and calculate gold cost", func(t *testing.T) {
		upgrades := []CardUpgrade{
			{CardName: "Knight", FromLevel: 0, ToLevel: 15, GoldCost: 0},
		}

		modified := analyzer.applyUpgradesToCardLevels(baseCardLevels, upgrades)

		// Check that the level was updated
		if modified["Knight"].Level != 15 {
			t.Errorf("Expected Knight level to be 15, got %d", modified["Knight"].Level)
		}

		// Check that FromLevel was inferred
		if upgrades[0].FromLevel != 9 {
			t.Errorf("Expected FromLevel to be inferred as 9, got %d", upgrades[0].FromLevel)
		}

		// Check that GoldCost was calculated (9→15 = 6 levels * 100 = 600)
		if upgrades[0].GoldCost != 600 {
			t.Errorf("Expected GoldCost to be 600, got %d", upgrades[0].GoldCost)
		}
	})

	t.Run("explicit from level and gold cost", func(t *testing.T) {
		upgrades := []CardUpgrade{
			{CardName: "Fireball", FromLevel: 11, ToLevel: 14, GoldCost: 0},
		}

		modified := analyzer.applyUpgradesToCardLevels(baseCardLevels, upgrades)

		// Check that the level was updated
		if modified["Fireball"].Level != 14 {
			t.Errorf("Expected Fireball level to be 14, got %d", modified["Fireball"].Level)
		}

		// Check that FromLevel was not changed
		if upgrades[0].FromLevel != 11 {
			t.Errorf("Expected FromLevel to remain 11, got %d", upgrades[0].FromLevel)
		}

		// Check that GoldCost was calculated (11→14 = 3 levels * 1000 = 3000)
		if upgrades[0].GoldCost != 3000 {
			t.Errorf("Expected GoldCost to be 3000, got %d", upgrades[0].GoldCost)
		}
	})

	t.Run("preserve card properties", func(t *testing.T) {
		upgrades := []CardUpgrade{
			{CardName: "Archers", FromLevel: 0, ToLevel: 15, GoldCost: 0},
		}

		modified := analyzer.applyUpgradesToCardLevels(baseCardLevels, upgrades)

		card := modified["Archers"]
		if card.MaxLevel != 16 {
			t.Errorf("Expected MaxLevel to be preserved as 16, got %d", card.MaxLevel)
		}
		if card.Rarity != "Common" {
			t.Errorf("Expected Rarity to be preserved as Common, got %s", card.Rarity)
		}
		if card.Elixir != 3 {
			t.Errorf("Expected Elixir to be preserved as 3, got %d", card.Elixir)
		}
		if card.EvolutionLevel != 1 {
			t.Errorf("Expected EvolutionLevel to be preserved as 1, got %d", card.EvolutionLevel)
		}
		if card.MaxEvolutionLevel != 3 {
			t.Errorf("Expected MaxEvolutionLevel to be preserved as 3, got %d", card.MaxEvolutionLevel)
		}
	})

	t.Run("multiple upgrades", func(t *testing.T) {
		upgrades := []CardUpgrade{
			{CardName: "Knight", FromLevel: 0, ToLevel: 15, GoldCost: 0},
			{CardName: "Archers", FromLevel: 0, ToLevel: 14, GoldCost: 0},
		}

		modified := analyzer.applyUpgradesToCardLevels(baseCardLevels, upgrades)

		if modified["Knight"].Level != 15 {
			t.Errorf("Expected Knight level to be 15, got %d", modified["Knight"].Level)
		}
		if modified["Archers"].Level != 14 {
			t.Errorf("Expected Archers level to be 14, got %d", modified["Archers"].Level)
		}

		// Check gold costs
		if upgrades[0].GoldCost != 600 {
			t.Errorf("Expected Knight GoldCost to be 600, got %d", upgrades[0].GoldCost)
		}
		if upgrades[1].GoldCost != 400 {
			t.Errorf("Expected Archers GoldCost to be 400, got %d", upgrades[1].GoldCost)
		}
	})
}

func TestCalculateImpact(t *testing.T) {
	analyzer := &WhatIfAnalyzer{}

	originalDeck := &deck.DeckRecommendation{
		Deck: []string{"Knight", "Archers", "Fireball", "Goblin Barrel", "Baby Dragon", "Arrows", "Goblin Hut", "Goblin Gang"},
		DeckDetail: []deck.CardDetail{
			{Name: "Knight", Score: 0.8},
			{Name: "Archers", Score: 1.0},
			{Name: "Fireball", Score: 0.9},
			{Name: "Goblin Barrel", Score: 1.0},
			{Name: "Baby Dragon", Score: 0.9},
			{Name: "Arrows", Score: 0.9},
			{Name: "Goblin Hut", Score: 0.9},
			{Name: "Goblin Gang", Score: 1.0},
		},
	}

	t.Run("positive score delta", func(t *testing.T) {
		simulatedDeck := &deck.DeckRecommendation{
			Deck: []string{"Knight", "Archers", "Fireball", "Goblin Barrel", "Baby Dragon", "Arrows", "Goblin Hut", "Goblin Gang"},
			DeckDetail: []deck.CardDetail{
				{Name: "Knight", Score: 1.3},
				{Name: "Archers", Score: 1.4},
				{Name: "Fireball", Score: 0.9},
				{Name: "Goblin Barrel", Score: 1.0},
				{Name: "Baby Dragon", Score: 0.9},
				{Name: "Arrows", Score: 0.9},
				{Name: "Goblin Hut", Score: 0.9},
				{Name: "Goblin Gang", Score: 1.0},
			},
		}

		upgrades := []CardUpgrade{
			{CardName: "Knight", FromLevel: 9, ToLevel: 15, GoldCost: 600},
			{CardName: "Archers", FromLevel: 10, ToLevel: 15, GoldCost: 500},
		}

		impact := analyzer.calculateImpact(originalDeck, simulatedDeck, upgrades)

		// Original score: 7.4, Simulated score: 8.3, Delta: 0.9
		if impact.DeckScoreDelta < 0.8 || impact.DeckScoreDelta > 1.0 {
			t.Errorf("Expected DeckScoreDelta around 0.9, got %.2f", impact.DeckScoreDelta)
		}

		// Viability improvement should be around 12% (0.9 / 7.4 * 100)
		if impact.ViabilityImprovement < 11.0 || impact.ViabilityImprovement > 13.0 {
			t.Errorf("Expected ViabilityImprovement around 12%%, got %.1f%%", impact.ViabilityImprovement)
		}
	})

	t.Run("deck composition changes", func(t *testing.T) {
		simulatedDeck := &deck.DeckRecommendation{
			Deck: []string{"Knight", "Archers", "Fireball", "Goblin Barrel", "Baby Dragon", "Zap", "Cannon", "Goblin Gang"},
			DeckDetail: []deck.CardDetail{
				{Name: "Knight", Score: 1.3},
				{Name: "Archers", Score: 1.4},
				{Name: "Fireball", Score: 0.9},
				{Name: "Goblin Barrel", Score: 1.0},
				{Name: "Baby Dragon", Score: 0.9},
				{Name: "Zap", Score: 1.0},
				{Name: "Cannon", Score: 1.0},
				{Name: "Goblin Gang", Score: 1.0},
			},
		}

		upgrades := []CardUpgrade{}

		impact := analyzer.calculateImpact(originalDeck, simulatedDeck, upgrades)

		// Check new cards
		if len(impact.NewCardsInDeck) != 2 {
			t.Errorf("Expected 2 new cards, got %d", len(impact.NewCardsInDeck))
		}

		// Check removed cards
		if len(impact.RemovedCards) != 2 {
			t.Errorf("Expected 2 removed cards, got %d", len(impact.RemovedCards))
		}
	})
}

func TestGenerateRecommendation(t *testing.T) {
	analyzer := &WhatIfAnalyzer{}

	tests := []struct {
		name                 string
		scoreDelta           float64
		viabilityImprovement float64
		totalGold            int
		wantSubstring        string
	}{
		{
			name:                 "negative delta",
			scoreDelta:           -0.5,
			viabilityImprovement: -6.0,
			totalGold:            1000,
			wantSubstring:        "not recommended",
		},
		{
			name:                 "high improvement",
			scoreDelta:           0.9,
			viabilityImprovement: 12.0,
			totalGold:            1100,
			wantSubstring:        "Highly recommended",
		},
		{
			name:                 "moderate improvement",
			scoreDelta:           0.5,
			viabilityImprovement: 7.0,
			totalGold:            600,
			wantSubstring:        "Recommended",
		},
		{
			name:                 "minor improvement",
			scoreDelta:           0.2,
			viabilityImprovement: 3.0,
			totalGold:            500,
			wantSubstring:        "Minor improvement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impact := &SimulationImpact{
				DeckScoreDelta:       tt.scoreDelta,
				ViabilityImprovement: tt.viabilityImprovement,
			}

			upgrades := []CardUpgrade{
				{GoldCost: tt.totalGold},
			}

			recommendation := analyzer.generateRecommendation(impact, upgrades)

			if len(recommendation) == 0 {
				t.Errorf("Expected non-empty recommendation")
			}

			// Check that the recommendation contains the expected substring (case-insensitive)
			if !contains(recommendation, tt.wantSubstring) {
				t.Errorf("Expected recommendation to contain '%s', got: %s", tt.wantSubstring, recommendation)
			}
		})
	}
}

func TestGenerateScenarioName(t *testing.T) {
	tests := []struct {
		name     string
		upgrades []CardUpgrade
		want     string
	}{
		{
			name: "single upgrade",
			upgrades: []CardUpgrade{
				{CardName: "Knight", ToLevel: 15},
			},
			want: "Upgrade Knight to Lv15",
		},
		{
			name: "multiple upgrades",
			upgrades: []CardUpgrade{
				{CardName: "Knight", ToLevel: 15},
				{CardName: "Archers", ToLevel: 14},
			},
			want: "Upgrade 2 cards: Knight, Archers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateScenarioName(tt.upgrades)
			if got != tt.want {
				t.Errorf("generateScenarioName() = %s, want %s", got, tt.want)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
