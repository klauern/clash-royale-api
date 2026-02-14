package config

import (
	"testing"
)

func TestNormalizeRarity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Standard cases
		{"Common lowercase", "common", "Common"},
		{"Rare lowercase", "rare", "Rare"},
		{"Epic lowercase", "epic", "Epic"},
		{"Legendary lowercase", "legendary", "Legendary"},
		{"Champion lowercase", "champion", "Champion"},

		// TitleCase (already normalized)
		{"Common TitleCase", "Common", "Common"},
		{"Rare TitleCase", "Rare", "Rare"},
		{"Epic TitleCase", "Epic", "Epic"},
		{"Legendary TitleCase", "Legendary", "Legendary"},
		{"Champion TitleCase", "Champion", "Champion"},

		// UPPERCASE
		{"Common UPPERCASE", "COMMON", "Common"},
		{"Rare UPPERCASE", "RARE", "Rare"},
		{"Epic UPPERCASE", "EPIC", "Epic"},
		{"Legendary UPPERCASE", "LEGENDARY", "Legendary"},
		{"Champion UPPERCASE", "CHAMPION", "Champion"},

		// With whitespace
		{"Common with spaces", "  common  ", "Common"},
		{"Rare with tabs", "\trare\t", "Rare"},
		{"Epic with newlines", "\nepic\n", "Epic"},

		// Mixed case
		{"Common MixedCase", "CoMmOn", "Common"},
		{"Legendary MixedCase", "LeGeNdArY", "Legendary"},

		// Edge cases
		{"Empty string", "", ""},
		{"Whitespace only", "   ", ""},
		{"Unknown rarity", "mythic", "Mythic"},
		{"Unknown with spaces", "  ultra rare  ", "Ultra Rare"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeRarity(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeRarity(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetRarityWeight(t *testing.T) {
	tests := []struct {
		name     string
		rarity   string
		expected float64
	}{
		// Standard rarities
		{"Common", "Common", 1.0},
		{"Rare", "Rare", 1.05},
		{"Epic", "Epic", 1.1},
		{"Legendary", "Legendary", 1.15},
		{"Champion", "Champion", 1.2},

		// Case variations
		{"common lowercase", "common", 1.0},
		{"RARE uppercase", "RARE", 1.05},
		{"ePiC mixed", "ePiC", 1.1},

		// With whitespace
		{"Common with spaces", "  Common  ", 1.0},
		{"Rare with tabs", "\tRare\t", 1.05},

		// Unknown rarities (should return default 1.0)
		{"Unknown rarity", "Mythic", 1.0},
		{"Empty string", "", 1.0},
		{"Invalid", "invalid", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRarityWeight(tt.rarity)
			if result != tt.expected {
				t.Errorf("GetRarityWeight(%q) = %.2f, want %.2f", tt.rarity, result, tt.expected)
			}
		})
	}
}

func TestGetRarityPriorityScore(t *testing.T) {
	tests := []struct {
		name     string
		rarity   string
		expected float64
	}{
		// Standard rarities (0-80 scale)
		{"Common", "Common", 0.0},
		{"Rare", "Rare", 20.0},
		{"Epic", "Epic", 40.0},
		{"Legendary", "Legendary", 60.0},
		{"Champion", "Champion", 80.0},

		// Case variations
		{"common lowercase", "common", 0.0},
		{"RARE uppercase", "RARE", 20.0},
		{"LegEndaRy mixed", "LegEndaRy", 60.0},

		// Unknown rarities (should return 0.0 = lowest priority)
		{"Unknown rarity", "Mythic", 0.0},
		{"Empty string", "", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRarityPriorityScore(tt.rarity)
			if result != tt.expected {
				t.Errorf("GetRarityPriorityScore(%q) = %.1f, want %.1f", tt.rarity, result, tt.expected)
			}
		})
	}
}

func TestGetRarityPriorityBonus(t *testing.T) {
	tests := []struct {
		name     string
		rarity   string
		expected float64
	}{
		// Standard rarities (1.0-2.5 scale, adjusted to prevent over-prioritization)
		{"Common", "Common", 1.0},
		{"Rare", "Rare", 1.3},
		{"Epic", "Epic", 1.7},
		{"Legendary", "Legendary", 2.2},
		{"Champion", "Champion", 2.5},

		// Case variations
		{"common lowercase", "common", 1.0},
		{"EPIC uppercase", "EPIC", 1.7},

		// Unknown rarities (should return 1.0 = neutral)
		{"Unknown rarity", "Mythic", 1.0},
		{"Empty string", "", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRarityPriorityBonus(tt.rarity)
			if result != tt.expected {
				t.Errorf("GetRarityPriorityBonus(%q) = %.1f, want %.1f", tt.rarity, result, tt.expected)
			}
		})
	}
}

func TestPriorityScoreVsBonusDifferentScales(t *testing.T) {
	// Verify that PriorityScore and PriorityBonus are on different scales
	// and serve different purposes (they should NOT be identical)

	rarities := []string{"Common", "Rare", "Epic", "Legendary", "Champion"}

	for _, rarity := range rarities {
		score := GetRarityPriorityScore(rarity)
		bonus := GetRarityPriorityBonus(rarity)

		// Score should be on 0-80 scale
		if score < 0 || score > 80 {
			t.Errorf("GetRarityPriorityScore(%q) = %.1f, expected to be in range [0, 80]", rarity, score)
		}

		// Bonus should be on 1.0-2.5 scale
		if bonus < 1.0 || bonus > 2.5 {
			t.Errorf("GetRarityPriorityBonus(%q) = %.1f, expected to be in range [1.0, 2.5]", rarity, bonus)
		}

		// They should be different values (different purposes)
		if score == bonus {
			t.Errorf("For %q, PriorityScore (%.1f) should not equal PriorityBonus (%.1f) - different scales", rarity, score, bonus)
		}
	}
}

func TestGetMaxLevel(t *testing.T) {
	tests := []struct {
		name     string
		rarity   string
		expected int
	}{
		// Standard rarities (all 16 currently)
		{"Common", "Common", 16},
		{"Rare", "Rare", 16},
		{"Epic", "Epic", 16},
		{"Legendary", "Legendary", 16},
		{"Champion", "Champion", 16},

		// Case variations
		{"common lowercase", "common", 16},
		{"RARE uppercase", "RARE", 16},

		// Unknown rarities (should return 0 to signal invalid)
		{"Unknown rarity", "Mythic", 0},
		{"Empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMaxLevel(tt.rarity)
			if result != tt.expected {
				t.Errorf("GetMaxLevel(%q) = %d, want %d", tt.rarity, result, tt.expected)
			}
		})
	}
}

func TestGetStartingLevel(t *testing.T) {
	tests := []struct {
		name     string
		rarity   string
		expected int
	}{
		// Standard rarities (progressive starting levels)
		{"Common", "Common", 1},
		{"Rare", "Rare", 3},
		{"Epic", "Epic", 6},
		{"Legendary", "Legendary", 9},
		{"Champion", "Champion", 11},

		// Case variations
		{"common lowercase", "common", 1},
		{"LEGENDARY uppercase", "LEGENDARY", 9},
		{"ChAmPiOn mixed", "ChAmPiOn", 11},

		// Unknown rarities (should return 0 to signal invalid)
		{"Unknown rarity", "Mythic", 0},
		{"Empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStartingLevel(tt.rarity)
			if result != tt.expected {
				t.Errorf("GetStartingLevel(%q) = %d, want %d", tt.rarity, result, tt.expected)
			}
		})
	}
}

func TestStartingLevelProgression(t *testing.T) {
	// Verify that starting levels are in ascending order by rarity
	rarities := GetAllRarities()
	previousLevel := 0

	for _, rarity := range rarities {
		level := GetStartingLevel(rarity)
		if level <= previousLevel {
			t.Errorf("Starting level progression broken: %s (%d) should be > previous (%d)", rarity, level, previousLevel)
		}
		previousLevel = level
	}
}

func TestGetAllRarities(t *testing.T) {
	rarities := GetAllRarities()

	// Should have exactly 5 rarities
	if len(rarities) != 5 {
		t.Errorf("GetAllRarities() returned %d rarities, want 5", len(rarities))
	}

	// Should contain all expected rarities in order
	expected := []string{"Common", "Rare", "Epic", "Legendary", "Champion"}
	for i, exp := range expected {
		if i >= len(rarities) {
			t.Errorf("Missing rarity at index %d: %s", i, exp)
			continue
		}
		if rarities[i] != exp {
			t.Errorf("GetAllRarities()[%d] = %q, want %q", i, rarities[i], exp)
		}
	}

	// Verify no duplicates
	seen := make(map[string]bool)
	for _, rarity := range rarities {
		if seen[rarity] {
			t.Errorf("Duplicate rarity found: %s", rarity)
		}
		seen[rarity] = true
	}
}

func TestAllRaritiesHaveCompleteData(t *testing.T) {
	// Verify that all rarities returned by GetAllRarities() have complete data
	rarities := GetAllRarities()

	for _, rarity := range rarities {
		// Check weight
		weight := GetRarityWeight(rarity)
		if weight == 0 {
			t.Errorf("%s has invalid weight (0)", rarity)
		}

		// Check priority score
		score := GetRarityPriorityScore(rarity)
		if score < 0 {
			t.Errorf("%s has invalid priority score (%f)", rarity, score)
		}

		// Check priority bonus
		bonus := GetRarityPriorityBonus(rarity)
		if bonus == 0 {
			t.Errorf("%s has invalid priority bonus (0)", rarity)
		}

		// Check max level
		maxLevel := GetMaxLevel(rarity)
		if maxLevel == 0 {
			t.Errorf("%s has invalid max level (0)", rarity)
		}

		// Check starting level
		startLevel := GetStartingLevel(rarity)
		if startLevel == 0 {
			t.Errorf("%s has invalid starting level (0)", rarity)
		}

		// Verify starting level < max level
		if startLevel >= maxLevel {
			t.Errorf("%s starting level (%d) >= max level (%d)", rarity, startLevel, maxLevel)
		}
	}
}

func TestLookupCardRarity(t *testing.T) {
	tests := []struct {
		name       string
		cardName   string
		wantRarity string
		wantFound  bool
	}{
		{name: "Legendary canonical", cardName: "The Log", wantRarity: "Legendary", wantFound: true},
		{name: "Legendary alias", cardName: "Log", wantRarity: "Legendary", wantFound: true},
		{name: "Champion", cardName: "Little Prince", wantRarity: "Champion", wantFound: true},
			{name: "Epic punctuation variant", cardName: "P.E.K.K.A", wantRarity: "Epic", wantFound: true},
			{name: "Rare punctuation variant", cardName: "Mini P.E.K.K.A", wantRarity: "Rare", wantFound: true},
			{name: "Rare three musketeers", cardName: "Three Musketeers", wantRarity: "Rare", wantFound: true},
			{name: "Unknown card", cardName: "Unknown Card", wantRarity: "", wantFound: false},
		}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRarity, gotFound := LookupCardRarity(tt.cardName)
			if gotFound != tt.wantFound {
				t.Fatalf("LookupCardRarity(%q) found = %v, want %v", tt.cardName, gotFound, tt.wantFound)
			}
			if gotRarity != tt.wantRarity {
				t.Fatalf("LookupCardRarity(%q) rarity = %q, want %q", tt.cardName, gotRarity, tt.wantRarity)
			}
		})
	}
}
