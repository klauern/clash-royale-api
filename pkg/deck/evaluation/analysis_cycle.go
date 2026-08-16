//nolint:funlen,goconst,gocritic,gocyclo // Existing domain logic is unchanged by the structural split.
package evaluation

import (
	"fmt"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// calculateCycleScore computes cycle efficiency on a 0-10 scale.
func calculateCycleScore(avgElixir float64, lowCostCount, shortestCycle int) float64 {
	// Cycle speed score (40%)
	cycleSpeedScore := 0.0
	if avgElixir < 3.0 {
		cycleSpeedScore = 10.0
	} else if avgElixir < 3.3 {
		cycleSpeedScore = 9.0
	} else if avgElixir < 3.6 {
		cycleSpeedScore = 7.0
	} else if avgElixir < 4.0 {
		cycleSpeedScore = 5.0
	} else {
		cycleSpeedScore = 3.0
	}

	// Low-cost card count score (35%)
	lowCostScore := 0.0
	if lowCostCount >= 4 {
		lowCostScore = 10.0
	} else if lowCostCount == 3 {
		lowCostScore = 7.0
	} else if lowCostCount == 2 {
		lowCostScore = 4.0
	} else if lowCostCount == 1 {
		lowCostScore = 2.0
	}

	// Shortest cycle score (25%)
	shortestCycleScore := 0.0
	if shortestCycle <= 6 {
		shortestCycleScore = 10.0
	} else if shortestCycle <= 8 {
		shortestCycleScore = 7.0
	} else if shortestCycle <= 10 {
		shortestCycleScore = 4.0
	} else {
		shortestCycleScore = 2.0
	}

	// Combine components
	score := (cycleSpeedScore * 0.4) + (lowCostScore * 0.35) + (shortestCycleScore * 0.25)

	return score
}

// BuildCycleAnalysis creates detailed cycle analysis
func BuildCycleAnalysis(deckCards []deck.CardCandidate) AnalysisSection {
	// Calculate cycle metrics
	avgElixir := calculateAvgElixir(deckCards)
	shortestCycle, _ := findShortestCycle(deckCards)
	elixirCurve := calculateElixirCurve(deckCards)

	// Count low-cost cards (≤ 2 elixir) using helper
	lowCostCards := filterByElixir(deckCards, 2)
	lowCostCount := len(lowCostCards)

	// Calculate score
	score := calculateCycleScore(avgElixir, lowCostCount, shortestCycle)
	rating := ScoreToRating(score)

	// Build details array
	details := []string{}

	// Average elixir
	cycleType := "Slow"
	if avgElixir < 3.0 {
		cycleType = "Fast"
	} else if avgElixir < 3.6 {
		cycleType = "Medium"
	}
	details = append(details, fmt.Sprintf("Average elixir: %.1f (%s Cycle)", avgElixir, cycleType))

	// Cycle cards
	if lowCostCount > 0 {
		details = append(details, fmt.Sprintf("Cycle cards (%d): %s",
			lowCostCount, buildCardList(lowCostCards)))
	}

	// Shortest 4-card cycle
	cycleAssessment := "poor rotation"
	if shortestCycle <= 6 {
		cycleAssessment = "excellent rotation"
	} else if shortestCycle <= 8 {
		cycleAssessment = "good rotation"
	}
	details = append(details, fmt.Sprintf("Shortest 4-card cycle: %d elixir (%s)",
		shortestCycle, cycleAssessment))

	// Rotation estimate (find win condition) using helper
	winConditions := filterByRole(deckCards, deck.RoleWinCondition)
	if len(winConditions) > 0 {
		winCondition := winConditions[0].Name
		rotationTime := int(avgElixir * 3.5) // Rough estimate
		details = append(details, fmt.Sprintf("Rotation estimate: Can return to %s in ~%d seconds",
			winCondition, rotationTime))
	}

	// Elixir curve distribution
	curveStr := ""
	for cost := 1; cost <= 8; cost++ {
		if count, ok := elixirCurve[cost]; ok && count > 0 {
			if curveStr != "" {
				curveStr += ", "
			}
			curveStr += fmt.Sprintf("%d-cost (%d)", cost, count)
		}
	}
	if curveStr != "" {
		details = append(details, fmt.Sprintf("Elixir curve: %s", curveStr))
	}

	// Tempo description
	if avgElixir < 3.2 {
		details = append(details, "Tempo: Constant pressure through rapid cycling")
	} else if avgElixir >= 4.0 {
		details = append(details, "Tempo: Slower build-up with larger pushes")
	}

	// Generate summary
	summary := "Medium cycle speed"
	if avgElixir < 3.0 {
		summary = "Fast cycle deck with excellent rotation speed"
	} else if avgElixir >= 4.0 {
		summary = "Slow cycle - focuses on larger pushes"
	}

	return AnalysisSection{
		Title:   "Cycle Analysis",
		Summary: summary,
		Details: details,
		Score:   score,
		Rating:  rating,
	}
}

// ============================================================================
// Phase 4: Ladder Analysis
// ============================================================================

// isLevelIndependent determines if card is effective when underleveled
