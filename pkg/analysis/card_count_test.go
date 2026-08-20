package analysis

import (
	"sync"
	"testing"
)

func TestDefaultCardCountConfig(t *testing.T) {
	t.Parallel()

	config := DefaultCardCountConfig()

	// Check defaults
	expected := map[string]int{
		"Common":    19,
		"Rare":      20,
		"Epic":      12,
		"Legendary": 10,
		"Champion":  6,
	}

	for rarity, want := range expected {
		if got := config.GetTotalCards(rarity); got != want {
			t.Errorf("Default %s cards = %d, want %d", rarity, got, want)
		}
	}
}

func TestDefaultCardCountConfigsAreIndependent(t *testing.T) {
	t.Parallel()

	first := DefaultCardCountConfig()
	second := DefaultCardCountConfig()
	first.cardCounts["Common"] = 1

	if got := second.GetTotalCards("Common"); got != 19 {
		t.Fatalf("second config Common cards = %d, want 19", got)
	}
}

func TestCardCountConfigUnknownRarity(t *testing.T) {
	t.Parallel()

	config := DefaultCardCountConfig()

	// Unknown rarity should return 0
	if got := config.GetTotalCards("Unknown"); got != 0 {
		t.Errorf("Unknown rarity = %d, want 0", got)
	}

	if got := config.GetTotalCards(""); got != 0 {
		t.Errorf("Empty rarity = %d, want 0", got)
	}
}

func TestCardCountConfigNilSafety(t *testing.T) {
	t.Parallel()

	var config *CardCountConfig

	// Nil config should return 0 without panicking
	if got := config.GetTotalCards("Common"); got != 0 {
		t.Errorf("Nil config GetTotalCards = %d, want 0", got)
	}
}

func TestCardCountConfigCaseInsensitive(t *testing.T) {
	t.Parallel()

	config := DefaultCardCountConfig()

	// Test case insensitivity through NormalizeRarity
	testCases := []struct {
		input string
		want  int
	}{
		{"common", 19},
		{"COMMON", 19},
		{"Common", 19},
		{"rare", 20},
		{"RARE", 20},
		{"Rare", 20},
	}

	for _, tc := range testCases {
		if got := config.GetTotalCards(tc.input); got != tc.want {
			t.Errorf("GetTotalCards(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestCardCountConfigConcurrent(t *testing.T) {
	t.Parallel()

	config := DefaultCardCountConfig()

	// Verify thread-safety with 100 parallel reads
	const numGoroutines = 100
	var wg sync.WaitGroup

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine reads different rarities
			rarities := []string{"Common", "Rare", "Epic", "Legendary", "Champion"}
			rarity := rarities[id%len(rarities)]

			_ = config.GetTotalCards(rarity)
		}(i)
	}

	wg.Wait()

	// If we got here without data races or panics, the test passed
}
