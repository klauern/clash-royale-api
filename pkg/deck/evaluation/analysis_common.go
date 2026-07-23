//nolint:goconst // Existing shared vocabulary is unchanged by the structural split.
package evaluation

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

const counterMatrixDataDir = "data"

var (
	counterMatrixOnce sync.Once
	counterMatrix     *deck.CounterMatrix
)

func getCounterMatrix() *deck.CounterMatrix {
	counterMatrixOnce.Do(func() {
		counterMatrix = deck.LoadCounterMatrix(counterMatrixDataDir, "")
	})
	return counterMatrix
}

// ============================================================================
// Phase 1: Foundation Helpers - Tier Scoring
// ============================================================================

// ============================================================================
// Phase 1: Foundation Helpers - Validation
// ============================================================================

// hasRole safely checks if card has the specified role
func hasRole(card deck.CardCandidate, role deck.CardRole) bool {
	return card.Role != nil && *card.Role == role
}

// canTargetAir safely checks if card can target air units
func canTargetAir(card deck.CardCandidate) bool {
	return card.Stats != nil &&
		(card.Stats.Targets == "Air" || card.Stats.Targets == "Air & Ground")
}

// getDPS safely retrieves DPS with default fallback
func getDPS(card deck.CardCandidate) float64 {
	if card.Stats != nil {
		return float64(card.Stats.DamagePerSecond)
	}
	return 0.0
}

// ============================================================================
// Phase 1: Foundation Helpers - Card Filtering
// ============================================================================

// filterByRole returns cards matching the specified role
func filterByRole(cards []deck.CardCandidate, role deck.CardRole) []deck.CardCandidate {
	filtered := []deck.CardCandidate{}
	for _, card := range cards {
		if hasRole(card, role) {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// filterByElixir returns cards with cost <= maxCost
func filterByElixir(cards []deck.CardCandidate, maxCost int) []deck.CardCandidate {
	filtered := []deck.CardCandidate{}
	for _, card := range cards {
		if card.Elixir <= maxCost {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// filterByDPS returns cards with DPS > threshold
func filterByDPS(cards []deck.CardCandidate, threshold float64) []deck.CardCandidate {
	filtered := []deck.CardCandidate{}
	for _, card := range cards {
		if getDPS(card) > threshold {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// filterByAirTargeting returns cards that can target air units
func filterByAirTargeting(cards []deck.CardCandidate) []deck.CardCandidate {
	filtered := []deck.CardCandidate{}
	for _, card := range cards {
		if canTargetAir(card) {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// ============================================================================
// Phase 1: Foundation Helpers - Summary Generation
// ============================================================================

// calculateElixirCurve returns distribution of cards across elixir costs
func calculateElixirCurve(cards []deck.CardCandidate) map[int]int {
	curve := make(map[int]int)
	for _, card := range cards {
		curve[card.Elixir]++
	}
	return curve
}

// findShortestCycle returns the sum of 4 cheapest cards and their names
func findShortestCycle(cards []deck.CardCandidate) (int, []string) {
	if len(cards) < 4 {
		return 0, []string{}
	}

	// Sort cards by elixir cost
	sorted := make([]deck.CardCandidate, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Elixir < sorted[j].Elixir
	})

	// Get 4 cheapest cards
	total := 0
	names := []string{}
	for i := range 4 {
		total += sorted[i].Elixir
		names = append(names, sorted[i].Name)
	}

	return total, names
}

// buildCardList formats card names with elixir costs
// Example: "Musketeer (4), Baby Dragon (4), Mega Minion (3)"
func buildCardList(cards []deck.CardCandidate) string {
	if len(cards) == 0 {
		return ""
	}

	parts := make([]string, len(cards))
	for i, card := range cards {
		parts[i] = fmt.Sprintf("%s (%d)", card.Name, card.Elixir)
	}
	return strings.Join(parts, ", ")
}

func extractNonEmptyCardNames(cards []deck.CardCandidate) []string {
	names := make([]string, 0, len(cards))
	for _, card := range cards {
		if card.Name == "" {
			continue
		}
		names = append(names, card.Name)
	}
	return names
}

func getResetRetargetCoverage(cards []deck.CardCandidate) []string {
	matrix := getCounterMatrix()
	return matrix.GetDeckCardsWithCapability(extractNonEmptyCardNames(cards), deck.CounterResetRetarget)
}
