//nolint:funlen,goconst,gocritic,gocyclo // Existing domain logic is unchanged by the structural split.
package evaluation

import (
	"fmt"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// classifyWinCondition determines the win condition category.
func classifyWinCondition(cardName string) string {
	// Direct damage
	directDamage := map[string]bool{
		"Hog Rider": true, "Giant": true, "Royal Giant": true,
		"Balloon": true, "Golem": true, "Lava Hound": true,
		"Electro Giant": true, "Royal Hogs": true, "Ram Rider": true,
	}

	// Siege
	siege := map[string]bool{
		"X-Bow": true, "Mortar": true,
	}

	// Chip damage
	chip := map[string]bool{
		"Miner": true, "Goblin Barrel": true, "Graveyard": true,
		"Goblin Drill": true, "Wall Breakers": true,
	}

	// Bridge spam
	bridgeSpam := map[string]bool{
		"Battle Ram": true, "P.E.K.K.A": true, "Mega Knight": true,
		"Royal Ghost": true, "Bandit": true, "Ram Rider": true,
	}

	if directDamage[cardName] {
		return "Direct Damage"
	}
	if siege[cardName] {
		return "Siege"
	}
	if chip[cardName] {
		return "Chip Damage"
	}
	if bridgeSpam[cardName] {
		return "Bridge Spam"
	}

	return "Win Condition"
}

// BuildAttackAnalysis creates detailed attack analysis
func BuildAttackAnalysis(deckCards []deck.CardCandidate) AnalysisSection {
	// Get numeric score from existing function
	categoryScore := ScoreAttack(deckCards)

	// Identify offensive elements using helper functions
	winConditions := filterByRole(deckCards, deck.RoleWinCondition)
	bigSpells := filterByRole(deckCards, deck.RoleSpellBig)

	// Build details array
	details := []string{}

	// Win conditions
	if len(winConditions) > 0 {
		category := classifyWinCondition(winConditions[0].Name)
		details = append(details, fmt.Sprintf("Primary win condition: %s (%s)",
			winConditions[0].Name, category))

		if len(winConditions) > 1 {
			category2 := classifyWinCondition(winConditions[1].Name)
			details = append(details, fmt.Sprintf("Secondary win condition: %s (%s)",
				winConditions[1].Name, category2))
		}
	} else {
		details = append(details, "⚠️  No dedicated win condition - may struggle to take towers")
	}

	// Spell damage
	if len(bigSpells) > 0 {
		spellList := buildCardList(bigSpells)
		assessment := "excellent"
		if len(bigSpells) == 1 {
			assessment = "good"
		}
		details = append(details, fmt.Sprintf("Spell damage: %s - %s finishing power",
			spellList, assessment))
	}

	// Bridge spam potential
	bridgeCards := []string{}
	for _, card := range winConditions {
		if classifyWinCondition(card.Name) == "Bridge Spam" {
			bridgeCards = append(bridgeCards, card.Name)
		}
	}
	if len(bridgeCards) > 0 {
		details = append(details, fmt.Sprintf("Bridge spam potential: %s can punish elixir disadvantage",
			strings.Join(bridgeCards, ", ")))
	}

	// Strategic recommendation
	if len(winConditions) > 0 {
		category := classifyWinCondition(winConditions[0].Name)
		switch category {
		case "Direct Damage":
			details = append(details, "Strategy: Apply consistent pressure with direct tower damage")
		case "Siege":
			details = append(details, "Strategy: Establish defensive perimeter and chip tower from range")
		case "Chip Damage":
			details = append(details, "Strategy: Accumulate small amounts of damage over time")
		case "Bridge Spam":
			details = append(details, "Strategy: Capitalize on elixir advantages with fast pushes")
		}
	}

	// Generate summary
	summary := "Strong offensive potential"
	if len(winConditions) == 0 {
		summary = "Lacks dedicated win condition"
	} else if len(winConditions) >= 2 {
		summary = "Versatile offense with multiple win conditions"
	} else if len(bigSpells) >= 2 {
		summary = "Strong offensive pressure with spell support"
	}

	return AnalysisSection{
		Title:   "Attack Analysis",
		Summary: summary,
		Details: details,
		Score:   categoryScore.Score,
		Rating:  categoryScore.Rating,
	}
}

// ============================================================================
// Phase 3: Complex Analysis Builders (Bait & Cycle)
// ============================================================================

// identifyBaitCards groups cards by spell vulnerability
