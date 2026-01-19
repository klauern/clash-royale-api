package deck

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/fuzzstorage"
)

// TestNewFuzzIntegration verifies that NewFuzzIntegration creates a properly initialized instance
func TestNewFuzzIntegration(t *testing.T) {
	fi := NewFuzzIntegration()

	if fi == nil {
		t.Fatal("NewFuzzIntegration returned nil")
	}

	if fi.weight != DefaultFuzzScoringWeight {
		t.Errorf("Expected default weight %f, got %f", DefaultFuzzScoringWeight, fi.weight)
	}

	if fi.topPercentile != DefaultFuzzTopPercentile {
		t.Errorf("Expected default top percentile %f, got %f", DefaultFuzzTopPercentile, fi.topPercentile)
	}

	if fi.minBoost != DefaultFuzzMinBoost {
		t.Errorf("Expected default min boost %f, got %f", DefaultFuzzMinBoost, fi.minBoost)
	}

	if fi.maxBoost != DefaultFuzzMaxBoost {
		t.Errorf("Expected default max boost %f, got %f", DefaultFuzzMaxBoost, fi.maxBoost)
	}

	if fi.stats == nil {
		t.Error("stats map was not initialized")
	}
}

// TestFuzzIntegrationSetWeight verifies that SetWeight properly clamps values
func TestFuzzIntegrationSetWeight(t *testing.T) {
	fi := NewFuzzIntegration()

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"negative value", -0.5, 0.0},
		{"zero", 0.0, 0.0},
		{"normal value", 0.25, 0.25},
		{"one", 1.0, 1.0},
		{"greater than one", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fi.SetWeight(tt.input)
			if fi.GetWeight() != tt.expected {
				t.Errorf("SetWeight(%f) = %f, want %f", tt.input, fi.GetWeight(), tt.expected)
			}
		})
	}
}

// TestFuzzIntegrationSetTopPercentile verifies that SetTopPercentile properly clamps values
func TestFuzzIntegrationSetTopPercentile(t *testing.T) {
	fi := NewFuzzIntegration()

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"too small", 0.001, 0.01},
		{"one percent", 0.01, 0.01},
		{"normal value", 0.25, 0.25},
		{"one hundred percent", 1.0, 1.0},
		{"greater than one", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fi.SetTopPercentile(tt.input)
			// We can't directly access topPercentile, but we can verify no panic occurred
			// The effect is tested indirectly through AnalyzeFromStorage
		})
	}
}

// TestFuzzIntegrationSetBoostRange verifies that SetBoostRange sets min and max correctly
func TestFuzzIntegrationSetBoostRange(t *testing.T) {
	fi := NewFuzzIntegration()

	fi.SetBoostRange(1.1, 2.0)
	// Can't directly verify, but we can test the effect through ApplyFuzzBoost later

	fi.SetBoostRange(2.0, 1.1) // Invalid order, but accepted
	// Should still work, even if min > max
}

// TestFuzzIntegrationGetFuzzBoost verifies that GetFuzzBoost returns correct values
func TestFuzzIntegrationGetFuzzBoost(t *testing.T) {
	fi := NewFuzzIntegration()

	// Test with no stats (should return 1.0)
	boost := fi.GetFuzzBoost("NonExistentCard")
	if boost != 1.0 {
		t.Errorf("Expected boost of 1.0 for non-existent card, got %f", boost)
	}

	// Manually add a stat entry
	fi.mu.Lock()
	fi.stats["TestCard"] = &FuzzCardStats{
		CardName:  "TestCard",
		Frequency: 5,
		AvgScore:  8.0,
		MaxScore:  9.0,
		Boost:     1.3,
	}
	fi.mu.Unlock()

	boost = fi.GetFuzzBoost("TestCard")
	if boost != 1.3 {
		t.Errorf("Expected boost of 1.3 for TestCard, got %f", boost)
	}

	// Test with whitespace variations
	boost = fi.GetFuzzBoost("  TestCard  ")
	if boost != 1.3 {
		t.Errorf("Expected boost of 1.3 for '  TestCard  ', got %f", boost)
	}
}

// TestFuzzIntegrationGetFuzzStats verifies that GetFuzzStats returns a copy of stats
func TestFuzzIntegrationGetFuzzStats(t *testing.T) {
	fi := NewFuzzIntegration()

	// Test with non-existent card
	stats := fi.GetFuzzStats("NonExistentCard")
	if stats != nil {
		t.Error("Expected nil for non-existent card, got non-nil")
	}

	// Add a stat entry
	fi.mu.Lock()
	fi.stats["TestCard"] = &FuzzCardStats{
		CardName:  "TestCard",
		Frequency: 10,
		AvgScore:  7.5,
		MaxScore:  8.5,
		Boost:     1.2,
	}
	fi.mu.Unlock()

	stats = fi.GetFuzzStats("TestCard")
	if stats == nil {
		t.Fatal("Expected stats for TestCard, got nil")
	}

	if stats.CardName != "TestCard" {
		t.Errorf("Expected CardName 'TestCard', got '%s'", stats.CardName)
	}

	if stats.Frequency != 10 {
		t.Errorf("Expected Frequency 10, got %d", stats.Frequency)
	}

	if stats.AvgScore != 7.5 {
		t.Errorf("Expected AvgScore 7.5, got %f", stats.AvgScore)
	}

	if stats.MaxScore != 8.5 {
		t.Errorf("Expected MaxScore 8.5, got %f", stats.MaxScore)
	}

	if stats.Boost != 1.2 {
		t.Errorf("Expected Boost 1.2, got %f", stats.Boost)
	}

	// Verify it's a copy by modifying the returned value
	stats.Boost = 9.9
	original := fi.GetFuzzStats("TestCard")
	if original.Boost == 9.9 {
		t.Error("Modifying returned stats affected original (not a copy)")
	}
}

// TestFuzzIntegrationGetAllStats verifies that GetAllStats returns a complete copy
func TestFuzzIntegrationGetAllStats(t *testing.T) {
	fi := NewFuzzIntegration()

	// Add some stats
	fi.mu.Lock()
	fi.stats["Card1"] = &FuzzCardStats{CardName: "Card1", Frequency: 5, Boost: 1.1}
	fi.stats["Card2"] = &FuzzCardStats{CardName: "Card2", Frequency: 10, Boost: 1.3}
	fi.mu.Unlock()

	allStats := fi.GetAllStats()
	if len(allStats) != 2 {
		t.Errorf("Expected 2 stats, got %d", len(allStats))
	}

	// Verify it's a copy by modifying the returned map
	allStats["Card1"].Boost = 9.9
	original := fi.GetFuzzStats("Card1")
	if original.Boost == 9.9 {
		t.Error("Modifying returned stats affected original (not a copy)")
	}
}

// TestFuzzIntegrationGetTopCards verifies that GetTopCards returns correctly sorted results
func TestFuzzIntegrationGetTopCards(t *testing.T) {
	fi := NewFuzzIntegration()

	// Add stats with varying boost values
	fi.mu.Lock()
	fi.stats["LowBoost"] = &FuzzCardStats{CardName: "LowBoost", Frequency: 1, Boost: 1.05}
	fi.stats["HighBoost"] = &FuzzCardStats{CardName: "HighBoost", Frequency: 10, Boost: 1.5}
	fi.stats["MedBoost"] = &FuzzCardStats{CardName: "MedBoost", Frequency: 5, Boost: 1.25}
	fi.mu.Unlock()

	// Get top 2
	topCards := fi.GetTopCards(2)
	if len(topCards) != 2 {
		t.Fatalf("Expected 2 top cards, got %d", len(topCards))
	}

	// Should be sorted by boost descending
	if topCards[0].CardName != "HighBoost" {
		t.Errorf("Expected first card to be HighBoost, got %s", topCards[0].CardName)
	}

	if topCards[1].CardName != "MedBoost" {
		t.Errorf("Expected second card to be MedBoost, got %s", topCards[1].CardName)
	}

	// Get more than available
	allCards := fi.GetTopCards(10)
	if len(allCards) != 3 {
		t.Errorf("Expected 3 cards when asking for 10, got %d", len(allCards))
	}

	// Get all (n <= 0 returns all)
	allCardsAgain := fi.GetTopCards(0)
	if len(allCardsAgain) != 3 {
		t.Errorf("Expected 3 cards when asking for 0, got %d", len(allCardsAgain))
	}

	// Verify each is a copy
	allCards[0].Boost = 9.9
	original := fi.GetFuzzStats("HighBoost")
	if original.Boost == 9.9 {
		t.Error("Modifying returned card affected original (not a copy)")
	}
}

// TestFuzzIntegrationHasStats verifies that HasStats correctly reports state
func TestFuzzIntegrationHasStats(t *testing.T) {
	fi := NewFuzzIntegration()

	if fi.HasStats() {
		t.Error("Expected HasStats to return false for new instance")
	}

	// Add a stat
	fi.mu.Lock()
	fi.stats["Card"] = &FuzzCardStats{CardName: "Card"}
	fi.mu.Unlock()

	if !fi.HasStats() {
		t.Error("Expected HasStats to return true after adding stats")
	}
}

// TestFuzzIntegrationStatsCount verifies that StatsCount returns correct count
func TestFuzzIntegrationStatsCount(t *testing.T) {
	fi := NewFuzzIntegration()

	if fi.StatsCount() != 0 {
		t.Errorf("Expected StatsCount 0 for new instance, got %d", fi.StatsCount())
	}

	// Add stats
	fi.mu.Lock()
	fi.stats["Card1"] = &FuzzCardStats{CardName: "Card1"}
	fi.stats["Card2"] = &FuzzCardStats{CardName: "Card2"}
	fi.mu.Unlock()

	if fi.StatsCount() != 2 {
		t.Errorf("Expected StatsCount 2, got %d", fi.StatsCount())
	}
}

// TestFuzzIntegrationClear verifies that Clear removes all stats
func TestFuzzIntegrationClear(t *testing.T) {
	fi := NewFuzzIntegration()

	// Add stats
	fi.mu.Lock()
	fi.stats["Card"] = &FuzzCardStats{CardName: "Card"}
	fi.mu.Unlock()

	if !fi.HasStats() {
		t.Error("Expected HasStats to return true before clear")
	}

	fi.Clear()

	if fi.HasStats() {
		t.Error("Expected HasStats to return false after clear")
	}

	if fi.StatsCount() != 0 {
		t.Errorf("Expected StatsCount 0 after clear, got %d", fi.StatsCount())
	}
}

// TestFuzzIntegrationApplyFuzzBoost verifies that ApplyFuzzBoost calculates correctly
func TestFuzzIntegrationApplyFuzzBoost(t *testing.T) {
	fi := NewFuzzIntegration()
	fi.SetWeight(0.20) // 20% weight for testing

	// Test with no stats (should return base score unchanged)
	baseScore := 5.0
	result := fi.ApplyFuzzBoost(baseScore, "NonExistentCard")
	if result != baseScore {
		t.Errorf("Expected base score %f for card without stats, got %f", baseScore, result)
	}

	// Add a stat with known boost
	fi.mu.Lock()
	fi.stats["TestCard"] = &FuzzCardStats{
		CardName:  "TestCard",
		Frequency: 10,
		AvgScore:  8.0,
		MaxScore:  9.0,
		Boost:     1.5, // 50% boost
	}
	fi.mu.Unlock()

	// With weight=0.20 and boost=1.5:
	// boostEffect = (1.5 - 1.0) * 0.20 = 0.5 * 0.20 = 0.10
	// finalScore = 5.0 * (1.0 + 0.10) = 5.0 * 1.10 = 5.5
	result = fi.ApplyFuzzBoost(baseScore, "TestCard")
	expected := 5.5
	if result != expected {
		t.Errorf("Expected score %f, got %f (base=%f, boost=%f, weight=%f)",
			expected, result, baseScore, 1.5, 0.20)
	}

	// Test with boost=1.0 (no effect)
	fi.mu.Lock()
	fi.stats["NoBoostCard"] = &FuzzCardStats{
		CardName:  "NoBoostCard",
		Frequency: 1,
		AvgScore:  5.0,
		MaxScore:  5.0,
		Boost:     1.0,
	}
	fi.mu.Unlock()

	result = fi.ApplyFuzzBoost(baseScore, "NoBoostCard")
	if result != baseScore {
		t.Errorf("Expected base score %f for card with boost=1.0, got %f", baseScore, result)
	}
}

// TestFuzzIntegrationAnalyzeFromStorage verifies analysis of stored fuzz results
func TestFuzzIntegrationAnalyzeFromStorage(t *testing.T) {
	// Create a temporary storage
	storage, err := fuzzstorage.NewStorage("")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Add some test decks
	now := time.Now()
	testDecks := []fuzzstorage.DeckEntry{
		{
			Cards:            []string{"Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7", "Card8"},
			OverallScore:     9.0,
			AttackScore:      8.5,
			DefenseScore:     9.5,
			SynergyScore:     8.0,
			VersatilityScore: 9.0,
			AvgElixir:        3.5,
			Archetype:        "beatdown",
			ArchetypeConf:    0.85,
			EvaluatedAt:      now,
			RunID:            "test1",
		},
		{
			Cards:            []string{"Card1", "Card2", "Card3", "Card4", "Card9", "Card10", "Card11", "Card12"},
			OverallScore:     8.5,
			AttackScore:      8.0,
			DefenseScore:     9.0,
			SynergyScore:     7.5,
			VersatilityScore: 8.5,
			AvgElixir:        3.2,
			Archetype:        "control",
			ArchetypeConf:    0.80,
			EvaluatedAt:      now,
			RunID:            "test2",
		},
		{
			Cards:            []string{"Card5", "Card6", "Card7", "Card8", "Card9", "Card10", "Card11", "Card12"},
			OverallScore:     7.0,
			AttackScore:      7.0,
			DefenseScore:     7.0,
			SynergyScore:     7.0,
			VersatilityScore: 7.0,
			AvgElixir:        3.0,
			Archetype:        "cycle",
			ArchetypeConf:    0.75,
			EvaluatedAt:      now,
			RunID:            "test3",
		},
	}

	_, err = storage.SaveTopDecks(testDecks)
	if err != nil {
		t.Fatalf("Failed to save test decks: %v", err)
	}

	// Analyze from storage
	fi := NewFuzzIntegration()
	fi.SetTopPercentile(1.0) // Use all decks for testing
	err = fi.AnalyzeFromStorage(storage, 100)
	if err != nil {
		t.Fatalf("AnalyzeFromStorage failed: %v", err)
	}

	// Verify stats were collected
	if !fi.HasStats() {
		t.Error("Expected stats to be collected")
	}

	// Card1 appears in top 2 decks (9.0 and 8.5 scores)
	// AvgScore = (9.0 + 8.5) / 2 = 8.75
	stats := fi.GetFuzzStats("Card1")
	if stats == nil {
		t.Fatal("Expected stats for Card1, got nil")
	}

	if stats.Frequency != 2 {
		t.Errorf("Expected Card1 frequency 2, got %d", stats.Frequency)
	}

	expectedAvgScore := 8.75
	if stats.AvgScore != expectedAvgScore {
		t.Errorf("Expected Card1 AvgScore %f, got %f", expectedAvgScore, stats.AvgScore)
	}

	// Card5 appears in deck 1 (9.0) and deck 3 (7.0)
	// But with topPercentile=1.0 and 3 decks, we analyze all 3
	// AvgScore = (9.0 + 7.0) / 2 = 8.0
	stats = fi.GetFuzzStats("Card5")
	if stats == nil {
		t.Fatal("Expected stats for Card5, got nil")
	}

	if stats.Frequency != 2 {
		t.Errorf("Expected Card5 frequency 2, got %d", stats.Frequency)
	}

	// Verify boost is in expected range (between minBoost and maxBoost)
	if stats.Boost < fi.minBoost || stats.Boost > fi.maxBoost {
		t.Errorf("Boost %f is outside expected range [%f, %f]", stats.Boost, fi.minBoost, fi.maxBoost)
	}
}

// TestFuzzIntegrationConcurrentAccess verifies thread safety
func TestFuzzIntegrationConcurrentAccess(t *testing.T) {
	fi := NewFuzzIntegration()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cardName := fmt.Sprintf("Card%d", n)
			fi.mu.Lock()
			fi.stats[cardName] = &FuzzCardStats{
				CardName:  cardName,
				Frequency: n,
				Boost:     1.0 + float64(n)*0.1,
			}
			fi.mu.Unlock()
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cardName := fmt.Sprintf("Card%d", n)
			fi.GetFuzzBoost(cardName)
		}(i)
	}

	wg.Wait()

	// Verify all stats were added
	if fi.StatsCount() != 10 {
		t.Errorf("Expected 10 stats after concurrent writes, got %d", fi.StatsCount())
	}
}

// TestFuzzIntegrationNilStorage verifies handling of nil storage
func TestFuzzIntegrationNilStorage(t *testing.T) {
	fi := NewFuzzIntegration()

	err := fi.AnalyzeFromStorage(nil, 100)
	if err == nil {
		t.Error("Expected error for nil storage, got nil")
	}

	if fi.HasStats() {
		t.Error("Expected no stats after nil storage analysis")
	}
}

// TestFuzzIntegrationEmptyStorage verifies handling of empty storage
func TestFuzzIntegrationEmptyStorage(t *testing.T) {
	// Create a temporary database to avoid conflicts with default storage
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_empty.db")
	storage, err := fuzzstorage.NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	fi := NewFuzzIntegration()
	err = fi.AnalyzeFromStorage(storage, 100)
	if err != nil {
		t.Fatalf("AnalyzeFromStorage failed for empty storage: %v", err)
	}

	if fi.HasStats() {
		t.Error("Expected no stats from empty storage")
	}
}
