package analysis

import (
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"strings"
	"testing"
)

// TestAnalyzeCardCollection_Basic tests basic card collection analysis
func TestAnalyzeCardCollection_Basic(t *testing.T) {
	player := &clashroyale.Player{
		Tag:  "#TEST123",
		Name: "TestPlayer",
		Cards: []clashroyale.Card{
			{
				ID:       26000000, // Knight
				Name:     "Knight",
				Level:    11,
				MaxLevel: 14,
				Count:    500,
				Rarity:   "Common",
			},
			{
				ID:       26000001, // Archers
				Name:     "Archers",
				Level:    12,
				MaxLevel: 14,
				Count:    1200,
				Rarity:   "Common",
			},
			{
				ID:       28000000, // Mega Knight
				Name:     "Mega Knight",
				Level:    13,
				MaxLevel: 14,
				Count:    10,
				Rarity:   "Legendary",
			},
		},
	}

	options := DefaultAnalysisOptions()
	analysis, err := AnalyzeCardCollection(player, options)
	if err != nil {
		t.Fatalf("AnalyzeCardCollection failed: %v", err)
	}
	if analysis == nil {
		t.Fatal("AnalyzeCardCollection returned nil analysis")
	}

	// Basic validation
	if analysis.PlayerTag != "#TEST123" {
		t.Errorf("PlayerTag = %v, want #TEST123", analysis.PlayerTag)
	}
	if analysis.PlayerName != "TestPlayer" {
		t.Errorf("PlayerName = %v, want TestPlayer", analysis.PlayerName)
	}
	if analysis.TotalCards != 3 {
		t.Errorf("TotalCards = %v, want 3", analysis.TotalCards)
	}
	if len(analysis.CardLevels) != 3 {
		t.Errorf("CardLevels length = %v, want 3", len(analysis.CardLevels))
	}
	if len(analysis.RarityBreakdown) != 2 { // Common and Legendary
		t.Errorf("RarityBreakdown length = %v, want 2", len(analysis.RarityBreakdown))
	}
	if len(analysis.UpgradePriority) == 0 {
		t.Error("UpgradePriority should not be empty")
	}
	if analysis.Summary.TotalCards != 3 {
		t.Errorf("Summary.TotalCards = %v, want 3", analysis.Summary.TotalCards)
	}

	// Verify card levels map
	knightInfo, ok := analysis.CardLevels["Knight"]
	if !ok {
		t.Fatal("Knight should be in CardLevels")
	}
	if knightInfo.Level != 11 {
		t.Errorf("Knight Level = %v, want 11", knightInfo.Level)
	}
	if knightInfo.MaxLevel != 14 {
		t.Errorf("Knight MaxLevel = %v, want 14", knightInfo.MaxLevel)
	}
	if knightInfo.Rarity != "Common" {
		t.Errorf("Knight Rarity = %v, want Common", knightInfo.Rarity)
	}
	if knightInfo.CardCount != 500 {
		t.Errorf("Knight CardCount = %v, want 500", knightInfo.CardCount)
	}
	if knightInfo.IsMaxLevel {
		t.Error("Knight IsMaxLevel = true, want false")
	}

	// Verify rarity breakdown
	commonStats, ok := analysis.RarityBreakdown["Common"]
	if !ok {
		t.Fatal("Common rarity should be in breakdown")
	}
	if commonStats.TotalCards != 2 {
		t.Errorf("Common TotalCards = %v, want 2", commonStats.TotalCards)
	}
	if commonStats.MaxLevelCards != 0 {
		t.Errorf("Common MaxLevelCards = %v, want 0", commonStats.MaxLevelCards)
	}

	legendaryStats, ok := analysis.RarityBreakdown["Legendary"]
	if !ok {
		t.Fatal("Legendary rarity should be in breakdown")
	}
	if legendaryStats.TotalCards != 1 {
		t.Errorf("Legendary TotalCards = %v, want 1", legendaryStats.TotalCards)
	}
	if legendaryStats.MaxLevelCards != 0 {
		t.Errorf("Legendary MaxLevelCards = %v, want 0", legendaryStats.MaxLevelCards)
	}

	// Verify summary calculations - average level should be (11+12+13)/3 = 12
	wantAvgLevel := 12.0
	if analysis.Summary.AvgCardLevel != wantAvgLevel {
		t.Errorf("Summary.AvgCardLevel = %v, want %v", analysis.Summary.AvgCardLevel, wantAvgLevel)
	}
}

// TestAnalyzeCardCollection_EmptyCards tests edge case with empty card collection
func TestAnalyzeCardCollection_EmptyCards(t *testing.T) {
	player := &clashroyale.Player{
		Tag:   "#EMPTY",
		Name:  "EmptyPlayer",
		Cards: []clashroyale.Card{},
	}

	options := DefaultAnalysisOptions()
	analysis, err := AnalyzeCardCollection(player, options)
	if err != nil {
		t.Fatalf("AnalyzeCardCollection failed: %v", err)
	}
	if analysis == nil {
		t.Fatal("AnalyzeCardCollection returned nil analysis")
	}

	if analysis.PlayerTag != "#EMPTY" {
		t.Errorf("PlayerTag = %v, want #EMPTY", analysis.PlayerTag)
	}
	if analysis.TotalCards != 0 {
		t.Errorf("TotalCards = %v, want 0", analysis.TotalCards)
	}
	if len(analysis.CardLevels) != 0 {
		t.Errorf("CardLevels length = %v, want 0", len(analysis.CardLevels))
	}
	if len(analysis.RarityBreakdown) != 0 {
		t.Errorf("RarityBreakdown length = %v, want 0", len(analysis.RarityBreakdown))
	}
	if len(analysis.UpgradePriority) != 0 {
		t.Errorf("UpgradePriority length = %v, want 0", len(analysis.UpgradePriority))
	}
	if analysis.Summary.TotalCards != 0 {
		t.Errorf("Summary.TotalCards = %v, want 0", analysis.Summary.TotalCards)
	}
	if analysis.Summary.MaxLevelCards != 0 {
		t.Errorf("Summary.MaxLevelCards = %v, want 0", analysis.Summary.MaxLevelCards)
	}
}

// TestAnalyzeCardCollection_NilPlayer tests error handling for nil player
func TestAnalyzeCardCollection_NilPlayer(t *testing.T) {
	options := DefaultAnalysisOptions()
	analysis, err := AnalyzeCardCollection(nil, options)
	if err == nil {
		t.Error("AnalyzeCardCollection should return error for nil player")
	}
	if analysis != nil {
		t.Error("AnalyzeCardCollection should return nil analysis for nil player")
	}
	if !strings.Contains(err.Error(), "player cannot be nil") {
		t.Errorf("Error message should contain 'player cannot be nil', got: %v", err)
	}
}

// TestAnalyzeCardCollection_MaxLevelCards tests analysis with max level cards
func TestAnalyzeCardCollection_MaxLevelCards(t *testing.T) {
	player := &clashroyale.Player{
		Tag:  "#MAXLEVEL",
		Name: "MaxLevelPlayer",
		Cards: []clashroyale.Card{
			{
				ID:       26000000,
				Name:     "Knight",
				Level:    14,
				MaxLevel: 14,
				Count:    5000,
				Rarity:   "Common",
			},
			{
				ID:       28000000,
				Name:     "Mega Knight",
				Level:    14,
				MaxLevel: 14,
				Count:    20,
				Rarity:   "Legendary",
			},
		},
	}

	options := DefaultAnalysisOptions()
	options.IncludeMaxLevel = true
	analysis, err := AnalyzeCardCollection(player, options)
	if err != nil {
		t.Fatalf("AnalyzeCardCollection failed: %v", err)
	}

	if analysis.TotalCards != 2 {
		t.Errorf("TotalCards = %v, want 2", analysis.TotalCards)
	}
	if analysis.Summary.MaxLevelCards != 2 {
		t.Errorf("Summary.MaxLevelCards = %v, want 2", analysis.Summary.MaxLevelCards)
	}
	if analysis.Summary.CompletionPercent != 100.0 {
		t.Errorf("Summary.CompletionPercent = %v, want 100.0", analysis.Summary.CompletionPercent)
	}

	// With IncludeMaxLevel=true, upgrade priorities should be empty
	// because all cards are max level
	if len(analysis.UpgradePriority) != 0 {
		t.Errorf("UpgradePriority length = %v, want 0 (all max level)", len(analysis.UpgradePriority))
	}
}

// TestAnalyzeCardCollection_WithOptions tests analysis with various options
func TestAnalyzeCardCollection_WithOptions(t *testing.T) {
	player := &clashroyale.Player{
		Tag:  "#OPTIONS",
		Name: "OptionsPlayer",
		Cards: []clashroyale.Card{
			{ID: 1, Name: "Common1", Level: 11, MaxLevel: 14, Count: 500, Rarity: "Common"},
			{ID: 2, Name: "Common2", Level: 12, MaxLevel: 14, Count: 600, Rarity: "Common"},
			{ID: 3, Name: "Rare1", Level: 8, MaxLevel: 14, Count: 100, Rarity: "Rare"},
			{ID: 4, Name: "Epic1", Level: 6, MaxLevel: 14, Count: 20, Rarity: "Epic"},
			{ID: 5, Name: "Legendary1", Level: 9, MaxLevel: 14, Count: 5, Rarity: "Legendary"},
		},
	}

	t.Run("FocusRarities", func(t *testing.T) {
		options := DefaultAnalysisOptions()
		options.FocusRarities = []string{"Common", "Rare"}
		analysis, err := AnalyzeCardCollection(player, options)
		if err != nil {
			t.Fatalf("AnalyzeCardCollection failed: %v", err)
		}

		// Should only have Common and Rare in breakdown
		if len(analysis.RarityBreakdown) != 2 {
			t.Errorf("RarityBreakdown length = %v, want 2", len(analysis.RarityBreakdown))
		}
		if _, ok := analysis.RarityBreakdown["Common"]; !ok {
			t.Error("RarityBreakdown should contain Common")
		}
		if _, ok := analysis.RarityBreakdown["Rare"]; !ok {
			t.Error("RarityBreakdown should contain Rare")
		}
		if _, ok := analysis.RarityBreakdown["Epic"]; ok {
			t.Error("RarityBreakdown should not contain Epic")
		}
		if _, ok := analysis.RarityBreakdown["Legendary"]; ok {
			t.Error("RarityBreakdown should not contain Legendary")
		}

		// Upgrade priorities should only include Common and Rare cards
		for _, priority := range analysis.UpgradePriority {
			if priority.Rarity != "Common" && priority.Rarity != "Rare" {
				t.Errorf("UpgradePriority contains unexpected rarity: %v", priority.Rarity)
			}
		}
	})

	t.Run("ExcludeCards", func(t *testing.T) {
		options := DefaultAnalysisOptions()
		options.ExcludeCards = []string{"Common1", "Epic1"}
		analysis, err := AnalyzeCardCollection(player, options)
		if err != nil {
			t.Fatalf("AnalyzeCardCollection failed: %v", err)
		}

		// Excluded cards should not appear in upgrade priorities
		for _, priority := range analysis.UpgradePriority {
			if priority.CardName == "Common1" {
				t.Error("Excluded card Common1 appears in upgrade priorities")
			}
			if priority.CardName == "Epic1" {
				t.Error("Excluded card Epic1 appears in upgrade priorities")
			}
		}
	})

	t.Run("TopN", func(t *testing.T) {
		options := DefaultAnalysisOptions()
		options.TopN = 2
		analysis, err := AnalyzeCardCollection(player, options)
		if err != nil {
			t.Fatalf("AnalyzeCardCollection failed: %v", err)
		}

		// Should have at most 2 upgrade priorities
		if len(analysis.UpgradePriority) > 2 {
			t.Errorf("UpgradePriority length = %v, want <= 2", len(analysis.UpgradePriority))
		}
	})

	t.Run("MinPriorityScore", func(t *testing.T) {
		options := DefaultAnalysisOptions()
		options.MinPriorityScore = 80.0 // High threshold
		analysis, err := AnalyzeCardCollection(player, options)
		if err != nil {
			t.Fatalf("AnalyzeCardCollection failed: %v", err)
		}

		// All priorities should have score >= 80
		for _, priority := range analysis.UpgradePriority {
			if priority.PriorityScore < 80.0 {
				t.Errorf("PriorityScore = %v, want >= 80.0", priority.PriorityScore)
			}
		}
	})
}

// TestCardAnalysis_Validate tests validation of CardAnalysis
func TestCardAnalysis_Validate(t *testing.T) {
	tests := []struct {
		name        string
		analysis    *CardAnalysis
		expectError bool
	}{
		{
			name: "Valid analysis",
			analysis: &CardAnalysis{
				PlayerTag:  "#VALID",
				TotalCards: 2,
				CardLevels: map[string]CardLevelInfo{
					"Card1": {Name: "Card1"},
					"Card2": {Name: "Card2"},
				},
			},
			expectError: false,
		},
		{
			name: "Missing player tag",
			analysis: &CardAnalysis{
				PlayerTag:  "",
				TotalCards: 1,
				CardLevels: map[string]CardLevelInfo{"Card1": {Name: "Card1"}},
			},
			expectError: true,
		},
		{
			name: "Negative card count",
			analysis: &CardAnalysis{
				PlayerTag:  "#TAG",
				TotalCards: -1,
				CardLevels: map[string]CardLevelInfo{"Card1": {Name: "Card1"}},
			},
			expectError: true,
		},
		{
			name: "Card count mismatch",
			analysis: &CardAnalysis{
				PlayerTag:  "#TAG",
				TotalCards: 2,
				CardLevels: map[string]CardLevelInfo{"Card1": {Name: "Card1"}},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.analysis.Validate()
			hasError := err != nil
			if hasError != tt.expectError {
				t.Errorf("Validate() error = %v, expectError = %v", err, tt.expectError)
			}
		})
	}
}
