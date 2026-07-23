//nolint:funlen,goconst,gocyclo // Existing domain logic is unchanged by the structural split.
package evaluation

import (
	"fmt"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// BuildDefenseAnalysis creates detailed defense analysis.
func BuildDefenseAnalysis(deckCards []deck.CardCandidate) AnalysisSection {
	// Get numeric score from existing function
	categoryScore := ScoreDefense(deckCards)

	// Count and identify defensive elements using helper functions
	airTargeters := filterByAirTargeting(deckCards)
	buildings := filterByRole(deckCards, deck.RoleBuilding)
	tankKillers := filterByDPS(deckCards, 150.0)
	resetRetargetCards := getResetRetargetCoverage(deckCards)

	// Investment cards (high elixir win conditions)
	investments := []deck.CardCandidate{}
	winConditions := filterByRole(deckCards, deck.RoleWinCondition)
	for _, card := range winConditions {
		if card.Elixir >= 6 {
			investments = append(investments, card)
		}
	}

	// Build details array
	details := []string{}

	// Anti-air coverage
	if len(airTargeters) > 0 {
		details = append(details, fmt.Sprintf("Anti-air units (%d): %s",
			len(airTargeters), buildCardList(airTargeters)))
	} else {
		details = append(details, "⚠️  No anti-air units - vulnerable to aerial threats")
	}

	// Defensive buildings
	if len(buildings) > 0 {
		details = append(details, fmt.Sprintf("Defensive buildings: %s", buildCardList(buildings)))
	} else {
		details = append(details, "⚠️  No defensive buildings - vulnerable to bridge spam")
	}

	// Tank killers
	if len(tankKillers) > 0 {
		details = append(details, fmt.Sprintf("Tank killers: %s provides strong ground defense", tankKillers[0].Name))
	}

	// Reset/retarget coverage
	if len(resetRetargetCards) > 0 {
		details = append(details, fmt.Sprintf("Reset/retarget tools: %s",
			strings.Join(resetRetargetCards, ", ")))
	} else {
		details = append(details, "⚠️  No reset/retarget tools - vulnerable to Inferno Tower/Dragon, Sparky, and charging units")
	}

	// Investment protection
	if len(investments) > 0 {
		details = append(details, fmt.Sprintf("⚠️  %s (%d elixir) needs defensive support",
			investments[0].Name, investments[0].Elixir))
	}

	// Generate summary using helper function
	airCount := float64(len(airTargeters))
	buildingCount := float64(len(buildings))

	summary := "Solid defensive capabilities"
	switch {
	case airCount == 0:
		summary = "No anti-air coverage - vulnerable to aerial threats"
	case airCount < 2:
		summary = "Weak anti-air coverage"
	case len(resetRetargetCards) == 0:
		summary = "Solid base defense but lacks reset/retarget protection"
	case buildingCount == 0:
		summary = "Good anti-air but lacks defensive buildings"
	case airCount >= 3 && buildingCount >= 1:
		summary = "Excellent defensive coverage with strong anti-air and buildings"
	}

	return AnalysisSection{
		Title:   "Defense Analysis",
		Summary: summary,
		Details: details,
		Score:   categoryScore.Score,
		Rating:  categoryScore.Rating,
	}
}

// classifyWinCondition determines win condition category
