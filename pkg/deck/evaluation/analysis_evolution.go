//nolint:gocritic // Existing domain logic is unchanged by the structural split.
package evaluation

import (
	"fmt"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// filterEvolvableCards partitions evolvable cards from those already evolved.
func filterEvolvableCards(deckCards []deck.CardCandidate) (evolvable, evolved []deck.CardCandidate) {
	for _, card := range deckCards {
		if card.MaxEvolutionLevel > 0 {
			evolvable = append(evolvable, card)
			if card.EvolutionLevel > 0 {
				evolved = append(evolved, card)
			}
		}
	}
	return evolvable, evolved
}

// calculateEvolutionPotential calculates evolution score (0-10)
func calculateEvolutionPotential(evolvableCards, evolvedCards []deck.CardCandidate) float64 {
	if len(evolvableCards) == 0 {
		return 0.0
	}

	// Base score: percentage of evolvable cards that are evolved
	evolutionRatio := float64(len(evolvedCards)) / float64(len(evolvableCards))
	score := evolutionRatio * 10.0

	// Add bonus for multiple evolved cards
	if len(evolvedCards) >= 2 {
		score += 1.0
	}
	if len(evolvedCards) >= 3 {
		score += 0.5
	}

	// Cap at 10
	if score > 10.0 {
		score = 10.0
	}

	return score
}

// addEvolutionProgressDetails adds player-specific evolution details
func addEvolutionProgressDetails(details []string, evolvableCards, evolvedCards []deck.CardCandidate, playerContext *PlayerContext) []string {
	if playerContext == nil {
		return details
	}

	unlockedEvolutions := playerContext.GetUnlockedEvolutionCards()
	details = append(details, fmt.Sprintf("Your unlocked evolutions: %d cards", len(unlockedEvolutions)))

	// Check which deck cards can evolve
	readyToEvolve := []string{}
	for _, card := range evolvableCards {
		if card.EvolutionLevel == 0 && playerContext.CanEvolve(card.Name) {
			readyToEvolve = append(readyToEvolve, card.Name)
		}
	}
	if len(readyToEvolve) > 0 {
		details = append(details, fmt.Sprintf("Ready to evolve: %s", strings.Join(readyToEvolve, ", ")))
	}

	// Show evolution progress for key cards
	if len(evolvedCards) > 0 {
		details = append(details, "Evolution progress:")
		for _, card := range evolvedCards {
			currentLevel, maxLevel, currentCount, requiredCount := playerContext.GetEvolutionProgress(card.Name)
			details = append(details, fmt.Sprintf("  %s: Level %d/%d, %d/%d cards",
				card.Name, currentLevel, maxLevel, currentCount, requiredCount))
		}
	}

	return details
}

// BuildEvolutionAnalysis creates detailed evolution analysis
// If playerContext is provided, shows player's evolution status
// If playerContext is nil, shows generic evolution potential for the deck
func BuildEvolutionAnalysis(deckCards []deck.CardCandidate, playerContext *PlayerContext) AnalysisSection {
	details := []string{}
	var score float64
	var rating Rating
	var summary string

	// Identify evolvable cards using helper
	evolvableInDeck, evolvedInDeck := filterEvolvableCards(deckCards)

	// Calculate evolution score (0-10)
	if len(evolvableInDeck) == 0 {
		score = 0.0
		summary = "No evolvable cards in deck"
		details = append(details, "This deck contains no cards with evolution potential")
	} else {
		// Calculate score using helper
		score = calculateEvolutionPotential(evolvableInDeck, evolvedInDeck)

		// Generate summary
		if len(evolvedInDeck) == 0 {
			summary = fmt.Sprintf("Deck has %d evolvable card(s) but none evolved", len(evolvableInDeck))
		} else if len(evolvedInDeck) == 1 {
			summary = fmt.Sprintf("Deck has 1 evolved card out of %d evolvable", len(evolvableInDeck))
		} else {
			summary = fmt.Sprintf("Deck has %d evolved cards out of %d evolvable", len(evolvedInDeck), len(evolvableInDeck))
		}

		// List evolvable cards
		if len(evolvableInDeck) > 0 {
			cardNames := []string{}
			for _, card := range evolvableInDeck {
				if card.EvolutionLevel > 0 {
					cardNames = append(cardNames, fmt.Sprintf("%s (Evo.%d/%d)", card.Name, card.EvolutionLevel, card.MaxEvolutionLevel))
				} else {
					cardNames = append(cardNames, fmt.Sprintf("%s (unevolved)", card.Name))
				}
			}
			details = append(details, fmt.Sprintf("Evolvable cards (%d): %s", len(evolvableInDeck), strings.Join(cardNames, ", ")))
		}

		// Add player-specific evolution details using helper
		details = addEvolutionProgressDetails(details, evolvableInDeck, evolvedInDeck, playerContext)

		// Evolution slot strategy
		if len(evolvedInDeck) > 2 {
			details = append(details, "⚠️  More than 2 evolved cards - prioritize best 2 for active slots")
		}

		// Evolution impact assessment
		if score >= 7.0 {
			details = append(details, "Evolution impact: Strong - evolutions significantly boost deck power")
		} else if score >= 4.0 {
			details = append(details, "Evolution impact: Moderate - some evolution synergy present")
		} else {
			details = append(details, "Evolution impact: Low - consider evolving key cards")
		}
	}

	rating = ScoreToRating(score)

	return AnalysisSection{
		Title:   "Evolution Analysis",
		Summary: summary,
		Details: details,
		Score:   score,
		Rating:  rating,
	}
}

// ============================================================================
// Phase 5: Main Orchestrator
// ============================================================================

// applyCriticalFlawPenalties applies additional penalties for critical compositional flaws
// that make a deck fundamentally unviable, beyond what category scores capture
//
//nolint:gocognit,gocyclo // Domain penalty matrix is explicit to keep balancing transparent.
