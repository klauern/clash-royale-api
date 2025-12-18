package analysis

import (
	"testing"
)

// TestCalculateCardsNeeded tests upgrade cost calculations
func TestCalculateCardsNeeded(t *testing.T) {
	tests := []struct {
		name         string
		currentLevel int
		rarity       string
		want         int
	}{
		// Common progression
		{"Common level 1", 1, "Common", 2},
		{"Common level 5", 5, "Common", 50},
		{"Common level 10", 10, "Common", 1000},
		{"Common level 13", 13, "Common", 10000},
		{"Common max level", 14, "Common", 0},

		// Rare progression
		{"Rare level 3", 3, "Rare", 2},
		{"Rare level 7", 7, "Rare", 50},
		{"Rare level 11", 11, "Rare", 800},
		{"Rare max level", 14, "Rare", 0},

		// Epic progression
		{"Epic level 6", 6, "Epic", 2},
		{"Epic level 10", 10, "Epic", 50},
		{"Epic level 13", 13, "Epic", 400},
		{"Epic max level", 14, "Epic", 0},

		// Legendary progression
		{"Legendary level 9", 9, "Legendary", 2},
		{"Legendary level 12", 12, "Legendary", 20},
		{"Legendary max level", 14, "Legendary", 0},

		// Champion progression
		{"Champion level 11", 11, "Champion", 2},
		{"Champion level 13", 13, "Champion", 10},
		{"Champion max level", 14, "Champion", 0},

		// Unknown rarity
		{"Unknown rarity", 5, "Unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateCardsNeeded(tt.currentLevel, tt.rarity)
			if got != tt.want {
				t.Errorf("CalculateCardsNeeded(%v, %v) = %v, want %v",
					tt.currentLevel, tt.rarity, got, tt.want)
			}
		})
	}
}

// TestGetMaxLevel tests max level retrieval
func TestGetMaxLevel(t *testing.T) {
	tests := []struct {
		rarity string
		want   int
	}{
		{"Common", 14},
		{"Rare", 14},
		{"Epic", 14},
		{"Legendary", 14},
		{"Champion", 14},
		{"Unknown", 14}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.rarity, func(t *testing.T) {
			got := GetMaxLevel(tt.rarity)
			if got != tt.want {
				t.Errorf("GetMaxLevel(%v) = %v, want %v", tt.rarity, got, tt.want)
			}
		})
	}
}

// TestGetStartingLevel tests starting level retrieval
func TestGetStartingLevel(t *testing.T) {
	tests := []struct {
		rarity string
		want   int
	}{
		{"Common", 1},
		{"Rare", 3},
		{"Epic", 6},
		{"Legendary", 9},
		{"Champion", 11},
		{"Unknown", 1}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.rarity, func(t *testing.T) {
			got := GetStartingLevel(tt.rarity)
			if got != tt.want {
				t.Errorf("GetStartingLevel(%v) = %v, want %v", tt.rarity, got, tt.want)
			}
		})
	}
}

// TestIsMaxLevel tests max level detection
func TestIsMaxLevel(t *testing.T) {
	tests := []struct {
		name         string
		currentLevel int
		rarity       string
		want         bool
	}{
		{"Common at max", 14, "Common", true},
		{"Common below max", 13, "Common", false},
		{"Rare at max", 14, "Rare", true},
		{"Epic at max", 14, "Epic", true},
		{"Legendary below max", 13, "Legendary", false},
		{"Champion at max", 14, "Champion", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMaxLevel(tt.currentLevel, tt.rarity)
			if got != tt.want {
				t.Errorf("IsMaxLevel(%v, %v) = %v, want %v",
					tt.currentLevel, tt.rarity, got, tt.want)
			}
		})
	}
}

// TestCalculateTotalCardsToMax tests total cards needed calculation
func TestCalculateTotalCardsToMax(t *testing.T) {
	tests := []struct {
		name         string
		currentLevel int
		rarity       string
		want         int
	}{
		{
			name:         "Common 1 to max",
			currentLevel: 1,
			rarity:       "Common",
			// 2+4+10+20+50+100+200+400+800+1000+2000+5000+10000 = 19586
			want: 19586,
		},
		{
			name:         "Common 13 to max",
			currentLevel: 13,
			rarity:       "Common",
			want:         10000,
		},
		{
			name:         "Common already max",
			currentLevel: 14,
			rarity:       "Common",
			want:         0,
		},
		{
			name:         "Rare 10 to max",
			currentLevel: 10,
			rarity:       "Rare",
			// 400+800+1000+2000 = 4200
			want: 4200,
		},
		{
			name:         "Epic 12 to max",
			currentLevel: 12,
			rarity:       "Epic",
			// 200+400 = 600
			want: 600,
		},
		{
			name:         "Legendary 9 to max",
			currentLevel: 9,
			rarity:       "Legendary",
			// 2+4+10+20+40 = 76
			want: 76,
		},
		{
			name:         "Champion 11 to max",
			currentLevel: 11,
			rarity:       "Champion",
			// 2+4+10 = 16
			want: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTotalCardsToMax(tt.currentLevel, tt.rarity)
			if got != tt.want {
				t.Errorf("CalculateTotalCardsToMax(%v, %v) = %v, want %v",
					tt.currentLevel, tt.rarity, got, tt.want)
			}
		})
	}
}

// TestCalculateUpgradeProgress tests upgrade progress percentage
func TestCalculateUpgradeProgress(t *testing.T) {
	tests := []struct {
		name        string
		cardsOwned  int
		cardsNeeded int
		want        float64
	}{
		{"0% progress", 0, 100, 0.0},
		{"50% progress", 50, 100, 50.0},
		{"100% progress", 100, 100, 100.0},
		{"Over 100%", 150, 100, 100.0}, // Capped at 100
		{"Zero needed", 50, 0, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateUpgradeProgress(tt.cardsOwned, tt.cardsNeeded)
			if got != tt.want {
				t.Errorf("CalculateUpgradeProgress(%v, %v) = %v, want %v",
					tt.cardsOwned, tt.cardsNeeded, got, tt.want)
			}
		})
	}
}

// TestCalculateUpgradeInfo tests complete upgrade info generation
func TestCalculateUpgradeInfo(t *testing.T) {
	info := CalculateUpgradeInfo("Fire Spirit", "Common", 1, 10, 500, 0)

	// Verify basic fields
	if info.CardName != "Fire Spirit" {
		t.Errorf("CardName = %v, want Fire Spirit", info.CardName)
	}
	if info.Rarity != "Common" {
		t.Errorf("Rarity = %v, want Common", info.Rarity)
	}
	if info.CurrentLevel != 10 {
		t.Errorf("CurrentLevel = %v, want 10", info.CurrentLevel)
	}
	if info.MaxLevel != 14 {
		t.Errorf("MaxLevel = %v, want 14", info.MaxLevel)
	}

	// Verify calculated fields
	if info.IsMaxLevel {
		t.Error("IsMaxLevel = true, want false")
	}
	if info.CardsOwned != 500 {
		t.Errorf("CardsOwned = %v, want 500", info.CardsOwned)
	}
	if info.CardsToNextLevel != 1000 {
		t.Errorf("CardsToNextLevel = %v, want 1000 (Common level 10)", info.CardsToNextLevel)
	}
	if info.ProgressPercent != 50.0 {
		t.Errorf("ProgressPercent = %v, want 50.0", info.ProgressPercent)
	}
	if info.CanUpgradeNow {
		t.Error("CanUpgradeNow = true, want false (only 500/1000 cards)")
	}
	if info.LevelsToMax != 4 {
		t.Errorf("LevelsToMax = %v, want 4 (10->14)", info.LevelsToMax)
	}

	// Test max level card
	maxInfo := CalculateUpgradeInfo("Max Card", "Legendary", 3, 14, 100, 0)
	if !maxInfo.IsMaxLevel {
		t.Error("Max level card: IsMaxLevel = false, want true")
	}
	if maxInfo.CanUpgradeNow {
		t.Error("Max level card: CanUpgradeNow = true, want false")
	}
	if maxInfo.CardsToNextLevel != 0 {
		t.Errorf("Max level card: CardsToNextLevel = %v, want 0", maxInfo.CardsToNextLevel)
	}
}

// TestCalculateRarityStats tests rarity statistics calculation
func TestCalculateRarityStats(t *testing.T) {
	cards := []UpgradeInfo{
		{Rarity: "Common", CurrentLevel: 10, IsMaxLevel: false, CanUpgradeNow: true, ProgressPercent: 50.0, TotalToMax: 100},
		{Rarity: "Common", CurrentLevel: 14, IsMaxLevel: true, CanUpgradeNow: false, ProgressPercent: 100.0, TotalToMax: 0},
		{Rarity: "Common", CurrentLevel: 12, IsMaxLevel: false, CanUpgradeNow: false, ProgressPercent: 20.0, TotalToMax: 50},
		{Rarity: "Rare", CurrentLevel: 8, IsMaxLevel: false, CanUpgradeNow: true, ProgressPercent: 75.0, TotalToMax: 200},
	}

	stats := CalculateRarityStats(cards, "Common")

	if stats.Rarity != "Common" {
		t.Errorf("Rarity = %v, want Common", stats.Rarity)
	}
	if stats.TotalCards != 3 {
		t.Errorf("TotalCards = %v, want 3", stats.TotalCards)
	}
	if stats.MaxLevelCards != 1 {
		t.Errorf("MaxLevelCards = %v, want 1", stats.MaxLevelCards)
	}
	if stats.UpgradableCards != 1 {
		t.Errorf("UpgradableCards = %v, want 1", stats.UpgradableCards)
	}

	// Average level: (10+14+12)/3 = 12
	if stats.AvgLevel != 12.0 {
		t.Errorf("AvgLevel = %v, want 12.0", stats.AvgLevel)
	}

	// Average progress: (50+100+20)/3 = 56.67
	expectedAvgProgress := (50.0 + 100.0 + 20.0) / 3.0
	if stats.AvgProgressPercent != expectedAvgProgress {
		t.Errorf("AvgProgressPercent = %v, want %v", stats.AvgProgressPercent, expectedAvgProgress)
	}

	// Total cards needed: 100+0+50 = 150
	if stats.TotalCardsNeeded != 150 {
		t.Errorf("TotalCardsNeeded = %v, want 150", stats.TotalCardsNeeded)
	}

	// Completion: 1/3 = 33.33%
	expectedCompletion := (1.0 / 3.0) * 100.0
	// Allow small floating point differences
	if stats.CompletionPercent < expectedCompletion-0.01 || stats.CompletionPercent > expectedCompletion+0.01 {
		t.Errorf("CompletionPercent = %v, want ~%v", stats.CompletionPercent, expectedCompletion)
	}
}

// TestCalculatePriorityScore tests priority scoring algorithm
func TestCalculatePriorityScore(t *testing.T) {
	tests := []struct {
		name    string
		info    UpgradeInfo
		wantMin float64
		wantMax float64
	}{
		{
			name: "Max level card",
			info: UpgradeInfo{
				Rarity:          "Common",
				CurrentLevel:    14,
				MaxLevel:        14,
				IsMaxLevel:      true,
				ProgressPercent: 100.0,
			},
			wantMin: 0.0,
			wantMax: 0.0,
		},
		{
			name: "Ready to upgrade common",
			info: UpgradeInfo{
				Rarity:          "Common",
				CurrentLevel:    10,
				MaxLevel:        14,
				IsMaxLevel:      false,
				CanUpgradeNow:   true,
				ProgressPercent: 100.0,
			},
			wantMin: 60.0, // High score, boosted by ready status
			wantMax: 100.0,
		},
		{
			name: "High level legendary nearly ready",
			info: UpgradeInfo{
				Rarity:          "Legendary",
				CurrentLevel:    13,
				MaxLevel:        14,
				IsMaxLevel:      false,
				CanUpgradeNow:   false,
				ProgressPercent: 80.0,
			},
			wantMin: 70.0, // High priority: high level, rare, nearly ready
			wantMax: 100.0,
		},
		{
			name: "Low progress common",
			info: UpgradeInfo{
				Rarity:          "Common",
				CurrentLevel:    5,
				MaxLevel:        14,
				IsMaxLevel:      false,
				CanUpgradeNow:   false,
				ProgressPercent: 10.0,
			},
			wantMin: 0.0,
			wantMax: 30.0, // Low priority: low level, low progress
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculatePriorityScore(tt.info)

			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("CalculatePriorityScore() = %v, want between %v and %v",
					score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestGetUpgradePriorities tests filtering and sorting priorities
func TestGetUpgradePriorities(t *testing.T) {
	cards := []UpgradeInfo{
		{CardName: "A", Rarity: "Common", CurrentLevel: 14, MaxLevel: 14, IsMaxLevel: true, ProgressPercent: 100.0},
		{CardName: "B", Rarity: "Legendary", CurrentLevel: 13, MaxLevel: 14, IsMaxLevel: false, CanUpgradeNow: true, ProgressPercent: 90.0},
		{CardName: "C", Rarity: "Common", CurrentLevel: 5, MaxLevel: 14, IsMaxLevel: false, ProgressPercent: 5.0},
		{CardName: "D", Rarity: "Epic", CurrentLevel: 10, MaxLevel: 14, IsMaxLevel: false, CanUpgradeNow: false, ProgressPercent: 60.0},
	}

	// Get top 2 with min score 30
	priorities := GetUpgradePriorities(cards, 30.0, 2)

	// Should exclude A (max level) and C (low score)
	// Should include B and D, sorted by priority
	if len(priorities) > 2 {
		t.Errorf("GetUpgradePriorities returned %v cards, want <= 2", len(priorities))
	}

	// First should be B (legendary, high progress, ready to upgrade)
	if len(priorities) > 0 && priorities[0].CardName != "B" {
		t.Errorf("First priority = %v, want B", priorities[0].CardName)
	}
}

// BenchmarkCalculateCardsNeeded benchmarks upgrade cost lookup
func BenchmarkCalculateCardsNeeded(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateCardsNeeded(10, "Common")
	}
}

// BenchmarkCalculateTotalCardsToMax benchmarks total cards calculation
func BenchmarkCalculateTotalCardsToMax(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateTotalCardsToMax(1, "Common")
	}
}

// BenchmarkCalculatePriorityScore benchmarks priority scoring
func BenchmarkCalculatePriorityScore(b *testing.B) {
	info := UpgradeInfo{
		Rarity:          "Legendary",
		CurrentLevel:    13,
		MaxLevel:        14,
		IsMaxLevel:      false,
		CanUpgradeNow:   true,
		ProgressPercent: 85.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculatePriorityScore(info)
	}
}
