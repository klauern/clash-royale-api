//nolint:funlen,goconst,gocritic,gocyclo // Existing domain logic is unchanged by the structural split.
package evaluation

import (
	"fmt"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// identifyBaitCards groups cards by spell vulnerability.
func identifyBaitCards(cards []deck.CardCandidate) map[string][]string {
	baitGroups := make(map[string][]string)

	// Define bait categories
	logBait := map[string]bool{
		"Goblin Gang": true, "Princess": true, "Dart Goblin": true,
		"Goblin Barrel": true, "Skeleton Barrel": true, "Rascals": true,
	}

	zapBait := map[string]bool{
		"Minion Horde": true, "Skeleton Army": true, "Bats": true,
		"Inferno Dragon": true, "Inferno Tower": true, "Sparky": true,
	}

	arrowsBait := map[string]bool{
		"Minions": true, "Spear Goblins": true, "Princess": true,
		"Dart Goblin": true, "Firecracker": true,
	}

	fireballBait := map[string]bool{
		"Three Musketeers": true, "Wizard": true, "Witch": true,
		"Flying Machine": true, "Elixir Collector": true, "Night Witch": true,
	}

	// Categorize cards
	for _, card := range cards {
		if logBait[card.Name] {
			baitGroups["Log"] = append(baitGroups["Log"], card.Name)
		}
		if zapBait[card.Name] {
			baitGroups["Zap"] = append(baitGroups["Zap"], card.Name)
		}
		if arrowsBait[card.Name] {
			baitGroups["Arrows"] = append(baitGroups["Arrows"], card.Name)
		}
		if fireballBait[card.Name] {
			baitGroups["Fireball"] = append(baitGroups["Fireball"], card.Name)
		}
	}

	return baitGroups
}

// calculateBaitScore computes bait potential (0-10)
func calculateBaitScore(baitGroups map[string][]string, hasGoblinBarrel, hasGoblinDrill bool) float64 {
	// Count total bait cards
	totalBaitCards := 0
	for _, cards := range baitGroups {
		totalBaitCards += len(cards)
	}

	// Count spell groups with 2+ cards (shared counter potential)
	sharedCounterGroups := 0
	for _, cards := range baitGroups {
		if len(cards) >= 2 {
			sharedCounterGroups++
		}
	}

	// Win condition fit
	winConditionFit := 0.0
	if hasGoblinBarrel || hasGoblinDrill {
		winConditionFit = 10.0
	} else if totalBaitCards >= 2 {
		winConditionFit = 6.0
	}

	// Bait card count score (50%) - using tier scoring
	baitCountScore := 0.0
	if totalBaitCards >= 4 {
		baitCountScore = 10.0
	} else if totalBaitCards == 3 {
		baitCountScore = 7.5
	} else if totalBaitCards == 2 {
		baitCountScore = 5.0
	} else if totalBaitCards == 1 {
		baitCountScore = 2.5
	}

	// Shared counter potential (30%) - using tier scoring
	sharedCounterScore := 0.0
	if sharedCounterGroups >= 3 {
		sharedCounterScore = 10.0
	} else if sharedCounterGroups == 2 {
		sharedCounterScore = 7.0
	} else if sharedCounterGroups == 1 {
		sharedCounterScore = 4.0
	}

	// Combine components
	score := (baitCountScore * 0.5) + (sharedCounterScore * 0.3) + (winConditionFit * 0.2)

	return score
}

// BuildBaitAnalysis creates detailed bait analysis
func BuildBaitAnalysis(deckCards []deck.CardCandidate) AnalysisSection {
	// Identify bait cards
	baitGroups := identifyBaitCards(deckCards)

	// Check for bait-friendly win conditions
	hasGoblinBarrel := false
	hasGoblinDrill := false
	for _, card := range deckCards {
		if card.Name == "Goblin Barrel" {
			hasGoblinBarrel = true
		}
		if card.Name == "Goblin Drill" {
			hasGoblinDrill = true
		}
	}

	// Calculate score
	score := calculateBaitScore(baitGroups, hasGoblinBarrel, hasGoblinDrill)
	rating := ScoreToRating(score)

	// Build details array
	details := []string{}

	// List bait groups
	for spell, cards := range baitGroups {
		if len(cards) >= 2 {
			details = append(details, fmt.Sprintf("%s bait units (%d): %s",
				spell, len(cards), strings.Join(cards, ", ")))
		}
	}

	// Find strongest bait chain
	maxSpell := ""
	maxCount := 0
	for spell, cards := range baitGroups {
		if len(cards) > maxCount {
			maxCount = len(cards)
			maxSpell = spell
		}
	}
	if maxCount >= 2 {
		details = append(details, fmt.Sprintf("Strongest bait chain: %s (%d vulnerable cards)",
			maxSpell, maxCount))
		details = append(details, "Mind-game potential: Opponent must choose which threat to spell")
	}

	// Win condition fit
	if hasGoblinBarrel {
		details = append(details, "Win condition fit: Goblin Barrel benefits from bait pressure")
	} else if hasGoblinDrill {
		details = append(details, "Win condition fit: Goblin Drill benefits from bait pressure")
	} else if score < 3.0 {
		details = append(details, "⚠️  Not a bait deck - lacks spell-vulnerable units")
	}

	// Generate summary
	summary := "Moderate bait potential"
	if score >= 7.0 {
		summary = "Excellent spell bait with multiple vulnerable units"
	} else if score < 3.0 {
		summary = "Not a bait-focused deck"
	}

	return AnalysisSection{
		Title:   "Bait Analysis",
		Summary: summary,
		Details: details,
		Score:   score,
		Rating:  rating,
	}
}

// calculateCycleScore computes cycle efficiency (0-10)
