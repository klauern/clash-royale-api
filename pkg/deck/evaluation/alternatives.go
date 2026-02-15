// Package evaluation provides comprehensive deck evaluation functionality
package evaluation

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// AlternativeDeck represents a suggested deck variation
type AlternativeDeck struct {
	// OriginalCard is the card being replaced
	OriginalCard string `json:"original_card"`

	// ReplacementCard is the suggested replacement
	ReplacementCard string `json:"replacement_card"`

	// Rationale explains why this change is suggested
	Rationale string `json:"rationale"`

	// Deck is the complete deck with the replacement
	Deck []string `json:"deck"`

	// OriginalScore is the score of the original deck
	OriginalScore float64 `json:"original_score"`

	// NewScore is the projected score with the replacement
	NewScore float64 `json:"new_score"`

	// ScoreDelta is the change in score (positive = improvement)
	ScoreDelta float64 `json:"score_delta"`

	// Impact describes the expected impact of this change
	Impact string `json:"impact"`
}

// AlternativeSuggestions contains all alternative deck suggestions
type AlternativeSuggestions struct {
	// OriginalDeck is the starting deck
	OriginalDeck []string `json:"original_deck"`

	// OriginalScore is the baseline score
	OriginalScore float64 `json:"original_score"`

	// Suggestions is the list of alternative decks
	Suggestions []AlternativeDeck `json:"suggestions"`

	// TopSuggestion is the best alternative (highest score improvement)
	TopSuggestion *AlternativeDeck `json:"top_suggestion,omitempty"`
}

// GenerateAlternatives creates alternative deck suggestions by swapping 1-2 cards
// Parameters:
//   - deckCards: The current deck cards
//   - synergyDB: Synergy database for scoring
//   - maxSuggestions: Maximum number of suggestions to return (default 5)
//   - playerCards: Optional map of cards the player owns (nil = all cards available)
func GenerateAlternatives(
	deckCards []deck.CardCandidate,
	synergyDB *deck.SynergyDatabase,
	maxSuggestions int,
	playerCards map[string]bool,
) *AlternativeSuggestions {
	if maxSuggestions <= 0 {
		maxSuggestions = 5
	}

	// Get original deck card names
	originalDeck := make([]string, len(deckCards))
	for i, card := range deckCards {
		originalDeck[i] = card.Name
	}

	// Evaluate original deck
	originalEval := Evaluate(deckCards, synergyDB, nil)

	result := &AlternativeSuggestions{
		OriginalDeck:  originalDeck,
		OriginalScore: originalEval.OverallScore,
		Suggestions:   make([]AlternativeDeck, 0, maxSuggestions),
	}

	// Generate single-card replacements
	alternatives := generateSingleCardAlternatives(deckCards, synergyDB, playerCards)

	// Sort by score improvement (descending)
	sort.Slice(alternatives, func(i, j int) bool {
		return alternatives[i].ScoreDelta > alternatives[j].ScoreDelta
	})

	// Take top N suggestions
	count := min(maxSuggestions, len(alternatives))

	result.Suggestions = alternatives[:count]

	// Set top suggestion
	if len(result.Suggestions) > 0 {
		result.TopSuggestion = &result.Suggestions[0]
	}

	return result
}

// generateSingleCardAlternatives generates alternatives by replacing one card at a time
func generateSingleCardAlternatives(
	deckCards []deck.CardCandidate,
	synergyDB *deck.SynergyDatabase,
	playerCards map[string]bool,
) []AlternativeDeck {
	alternatives := make([]AlternativeDeck, 0)

	// Get original score
	originalEval := Evaluate(deckCards, synergyDB, nil)
	originalScore := originalEval.OverallScore

	// For each card in the deck, try replacing it
	for i, originalCard := range deckCards {
		// Get potential replacements for this card
		replacements := findReplacements(originalCard, deckCards, playerCards)

		for _, replacement := range replacements {
			// Create new deck with replacement
			newDeck := make([]deck.CardCandidate, len(deckCards))
			copy(newDeck, deckCards)
			newDeck[i] = replacement

			// Evaluate new deck
			newEval := Evaluate(newDeck, synergyDB, nil)
			newScore := newEval.OverallScore

			// Only suggest if it improves the score
			scoreDelta := newScore - originalScore
			if scoreDelta <= 0 {
				continue
			}

			// Build rationale
			rationale := buildRationale(originalCard, replacement, originalEval, newEval)

			// Determine impact level
			impact := determineImpact(scoreDelta)

			// Get deck card names
			deckNames := make([]string, len(newDeck))
			for j, card := range newDeck {
				deckNames[j] = card.Name
			}

			alternatives = append(alternatives, AlternativeDeck{
				OriginalCard:    originalCard.Name,
				ReplacementCard: replacement.Name,
				Rationale:       rationale,
				Deck:            deckNames,
				OriginalScore:   originalScore,
				NewScore:        newScore,
				ScoreDelta:      scoreDelta,
				Impact:          impact,
			})
		}
	}

	return alternatives
}

// findReplacements finds suitable replacement cards for a given card
func findReplacements(
	originalCard deck.CardCandidate,
	currentDeck []deck.CardCandidate,
	playerCards map[string]bool,
) []deck.CardCandidate {
	replacements := make([]deck.CardCandidate, 0)

	// Build a set of cards already in the deck to avoid duplicates
	deckCardSet := make(map[string]bool)
	for _, card := range currentDeck {
		deckCardSet[card.Name] = true
	}

	// Get cards of similar role or elixir cost
	candidates := getSimilarCards(originalCard)

	for _, candidate := range candidates {
		// Skip if already in deck
		if deckCardSet[candidate.Name] {
			continue
		}

		// Skip if player doesn't own the card (if playerCards is provided)
		if playerCards != nil && !playerCards[candidate.Name] {
			continue
		}

		replacements = append(replacements, candidate)
	}

	return replacements
}

// getSimilarCards returns cards similar to the given card (same role or elixir cost)
func getSimilarCards(card deck.CardCandidate) []deck.CardCandidate {
	// This is a simplified implementation
	// In a full implementation, this would query a card database
	similar := make([]deck.CardCandidate, 0)

	// For now, we'll return a hardcoded set of common alternatives
	// This should be replaced with actual card database lookup
	commonAlternatives := map[string][]string{
		"Knight":         {"Valkyrie", "Ice Golem", "Dark Prince"},
		"Hog Rider":      {"Ram Rider", "Battle Ram", "Royal Hogs"},
		"Fireball":       {"Poison", "Lightning", "Rocket"},
		"Zap":            {"The Log", "Arrows", "Giant Snowball"},
		"Musketeer":      {"Hunter", "Magic Archer", "Flying Machine"},
		"Mega Minion":    {"Minions", "Bats", "Minion Horde"},
		"Ice Spirit":     {"Fire Spirit", "Heal Spirit", "Electro Spirit"},
		"Tesla":          {"Cannon", "Inferno Tower", "Bomb Tower"},
		"Prince":         {"Dark Prince", "Mini P.E.K.K.A", "Valkyrie"},
		"Goblin Gang":    {"Skeleton Army", "Guards", "Rascals"},
		"Balloon":        {"Lava Hound", "Giant", "Golem"},
		"Wizard":         {"Executioner", "Baby Dragon", "Witch"},
		"Giant":          {"Golem", "Royal Giant", "Goblin Giant"},
		"P.E.K.K.A":      {"Mega Knight", "Golem", "Giant Skeleton"},
		"Electro Wizard": {"Ice Wizard", "Witch", "Mother Witch"},
	}

	// Get alternatives for this card
	altNames, exists := commonAlternatives[card.Name]
	if !exists {
		return similar
	}

	// Create CardCandidate objects for alternatives
	for _, altName := range altNames {
		alt := deck.CardCandidate{
			Name:     altName,
			Level:    card.Level,
			MaxLevel: card.MaxLevel,
			Elixir:   inferElixirForCard(altName),
			Role:     card.Role,
		}
		similar = append(similar, alt)
	}

	return similar
}

// buildRationale creates a rationale for why a replacement is suggested
func buildRationale(
	original deck.CardCandidate,
	replacement deck.CardCandidate,
	originalEval EvaluationResult,
	newEval EvaluationResult,
) string {
	// Analyze what improved
	improvements := make([]string, 0)

	if newEval.Attack.Score > originalEval.Attack.Score {
		improvements = append(improvements, "stronger attack")
	}
	if newEval.Defense.Score > originalEval.Defense.Score {
		improvements = append(improvements, "better defense")
	}
	if newEval.Synergy.Score > originalEval.Synergy.Score {
		improvements = append(improvements, "improved synergy")
	}
	if newEval.Versatility.Score > originalEval.Versatility.Score {
		improvements = append(improvements, "more versatile")
	}

	if len(improvements) == 0 {
		return fmt.Sprintf("Replacing %s with %s improves overall deck performance", original.Name, replacement.Name)
	}

	return fmt.Sprintf("Replacing %s with %s provides %s", original.Name, replacement.Name, joinImprovements(improvements))
}

// joinImprovements joins improvement descriptions with proper grammar
func joinImprovements(improvements []string) string {
	if len(improvements) == 0 {
		return ""
	}
	if len(improvements) == 1 {
		return improvements[0]
	}
	if len(improvements) == 2 {
		return improvements[0] + " and " + improvements[1]
	}

	// For 3+ items, use commas
	var result strings.Builder
	for i, imp := range improvements {
		if i == len(improvements)-1 {
			result.WriteString("and " + imp)
		} else if i > 0 {
			result.WriteString(", " + imp)
		} else {
			result.WriteString(imp)
		}
	}
	return result.String()
}

// determineImpact determines the impact level based on score delta
func determineImpact(scoreDelta float64) string {
	absScore := math.Abs(scoreDelta)

	if absScore >= 2.0 {
		return "Major Improvement"
	} else if absScore >= 1.0 {
		return "Significant Improvement"
	} else if absScore >= 0.5 {
		return "Moderate Improvement"
	} else {
		return "Minor Improvement"
	}
}

// inferElixirForCard infers the elixir cost for a card by name
// This is a simplified version - should use actual card database
func inferElixirForCard(name string) int {
	// Hardcoded elixir costs for common cards
	elixirCosts := map[string]int{
		"Knight":         3,
		"Valkyrie":       4,
		"Ice Golem":      2,
		"Dark Prince":    4,
		"Hog Rider":      4,
		"Ram Rider":      5,
		"Battle Ram":     4,
		"Royal Hogs":     5,
		"Fireball":       4,
		"Poison":         4,
		"Lightning":      6,
		"Rocket":         6,
		"Zap":            2,
		"The Log":        2,
		"Arrows":         3,
		"Giant Snowball": 2,
		"Musketeer":      4,
		"Hunter":         4,
		"Magic Archer":   4,
		"Flying Machine": 4,
		"Mega Minion":    3,
		"Minions":        3,
		"Bats":           2,
		"Minion Horde":   5,
		"Ice Spirit":     1,
		"Fire Spirit":    1,
		"Heal Spirit":    1,
		"Electro Spirit": 1,
		"Tesla":          4,
		"Cannon":         3,
		"Inferno Tower":  5,
		"Bomb Tower":     4,
		"Prince":         5,
		"Mini P.E.K.K.A": 4,
		"Goblin Gang":    3,
		"Skeleton Army":  3,
		"Guards":         3,
		"Rascals":        5,
		"Balloon":        5,
		"Lava Hound":     7,
		"Giant":          5,
		"Golem":          8,
		"Wizard":         5,
		"Executioner":    5,
		"Baby Dragon":    4,
		"Witch":          5,
		"Royal Giant":    6,
		"Goblin Giant":   6,
		"Giant Skeleton": 6,
		"P.E.K.K.A":      7,
		"Mega Knight":    7,
		"Electro Wizard": 4,
		"Ice Wizard":     3,
		"Mother Witch":   4,
	}

	if cost, exists := elixirCosts[name]; exists {
		return cost
	}

	// Default to 3 if unknown
	return 3
}
