package evaluation

import (
	"math"
	"slices"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// ArchetypeDetectionResult contains the detected archetype with confidence scores
type ArchetypeDetectionResult struct {
	// Primary archetype with highest confidence
	Primary Archetype
	// PrimaryConfidence is 0.0-1.0 confidence score for primary archetype
	PrimaryConfidence float64

	// Secondary archetype for hybrid decks (may be ArchetypeUnknown if not hybrid)
	Secondary Archetype
	// SecondaryConfidence is 0.0-1.0 confidence score for secondary archetype
	SecondaryConfidence float64

	// IsHybrid indicates if this deck has multiple distinct archetypes
	IsHybrid bool
}

// DetectArchetype analyzes a deck and returns the detected archetype with confidence scoring
func DetectArchetype(deckCards []deck.CardCandidate) ArchetypeDetectionResult {
	if len(deckCards) == 0 {
		return ArchetypeDetectionResult{
			Primary:           ArchetypeUnknown,
			PrimaryConfidence: 0.0,
		}
	}

	// Calculate scores for each archetype
	archetypeScores := make(map[Archetype]float64)

	archetypeScores[ArchetypeBeatdown] = scoreBeatdown(deckCards)
	archetypeScores[ArchetypeControl] = scoreControl(deckCards)
	archetypeScores[ArchetypeCycle] = scoreCycle(deckCards)
	archetypeScores[ArchetypeBridge] = scoreBridgeSpam(deckCards)
	archetypeScores[ArchetypeSiege] = scoreSiege(deckCards)
	archetypeScores[ArchetypeBait] = scoreBait(deckCards)
	archetypeScores[ArchetypeGraveyard] = scoreGraveyard(deckCards)
	archetypeScores[ArchetypeMiner] = scoreMiner(deckCards)

	// Find top 2 archetypes
	primary, primaryScore := findTopArchetype(archetypeScores)
	delete(archetypeScores, primary) // Remove primary to find secondary
	secondary, secondaryScore := findTopArchetype(archetypeScores)

	// Normalize scores to confidence (0.0-1.0)
	// A score > 7.0 is considered high confidence
	primaryConfidence := normalizeConfidence(primaryScore)
	secondaryConfidence := normalizeConfidence(secondaryScore)

	// Determine if hybrid: secondary archetype must have >70% of primary score
	// and both must have high confidence (0.7 threshold)
	// Also require significant score gap to avoid marking similar archetypes as hybrid
	scoreGap := primaryScore - secondaryScore
	isHybrid := secondaryConfidence > 0.7 && primaryConfidence > 0.7 &&
		secondaryScore > 0.7*primaryScore && scoreGap < 2.0

	result := ArchetypeDetectionResult{
		Primary:             primary,
		PrimaryConfidence:   primaryConfidence,
		Secondary:           secondary,
		SecondaryConfidence: secondaryConfidence,
		IsHybrid:            isHybrid,
	}

	// If primary confidence is too low, mark as unknown
	if primaryConfidence < 0.3 {
		result.Primary = ArchetypeUnknown
		result.IsHybrid = false
	}

	// Mark as hybrid archetype only if truly hybrid (not related archetypes)
	// Skip hybrid marking for closely related archetype pairs
	if isHybrid && areRelatedArchetypes(primary, secondary) {
		result.IsHybrid = false // Related archetypes, not true hybrid
	}

	if result.IsHybrid {
		result.Primary = ArchetypeHybrid
	}

	return result
}

// isArchetypePair checks if two archetypes match a pair (order-independent)
func isArchetypePair(a1, a2, target1, target2 Archetype) bool {
	return (a1 == target1 && a2 == target2) || (a1 == target2 && a2 == target1)
}

// areRelatedArchetypes returns true if two archetypes are closely related
// and shouldn't be marked as "hybrid" together
func areRelatedArchetypes(a1, a2 Archetype) bool {
	return isArchetypePair(a1, a2, ArchetypeSiege, ArchetypeControl) ||
		isArchetypePair(a1, a2, ArchetypeGraveyard, ArchetypeControl) ||
		isArchetypePair(a1, a2, ArchetypeMiner, ArchetypeCycle) ||
		isArchetypePair(a1, a2, ArchetypeBridge, ArchetypeBeatdown)
}

// normalizeConfidence converts a 0-10 score to 0.0-1.0 confidence
// Uses a sigmoid-like curve for better confidence distribution
func normalizeConfidence(score float64) float64 {
	if score <= 0 {
		return 0.0
	}
	if score >= 10.0 {
		return 1.0
	}

	// Use a smooth scaling: scores 5-8 map to 0.5-0.9 confidence
	// This provides good differentiation in the useful range
	confidence := score / 10.0

	// Apply slight sigmoid curve for better distribution
	confidence = 1.0 / (1.0 + math.Exp(-5.0*(confidence-0.5)))

	return math.Min(1.0, math.Max(0.0, confidence))
}

// findTopArchetype returns the archetype with the highest score
func findTopArchetype(scores map[Archetype]float64) (Archetype, float64) {
	var topArchetype Archetype = ArchetypeUnknown
	topScore := 0.0

	for archetype, score := range scores {
		if score > topScore {
			topScore = score
			topArchetype = archetype
		}
	}

	return topArchetype, topScore
}

// scoreBeatdown scores a deck's fit for beatdown archetype (0-10 scale)
// Beatdown: Heavy tanks + support troops + big spells
func scoreBeatdown(deckCards []deck.CardCandidate) float64 {
	score := 0.0
	heavyTanks := []string{"Golem", "Lava Hound", "Electro Giant", "Giant", "Mega Knight"}
	supportTroops := []string{"Baby Dragon", "Night Witch", "Lumberjack", "Mega Minion", "Witch"}

	// Check for heavy tank win conditions (40% of score)
	tankScore := 0.0
	for _, card := range deckCards {
		if slices.Contains(heavyTanks, card.Name) {
			tankScore = 10.0
		}
	}

	// Check for support troops (30% of score)
	supportCount := 0
	for _, card := range deckCards {
		if slices.Contains(supportTroops, card.Name) {
			supportCount++
		}
	}
	supportScore := float64(supportCount) * 2.5 // Max 10.0 with 4+ supports
	if supportScore > 10.0 {
		supportScore = 10.0
	}

	// Check average elixir (30% of score) - beatdown typically 3.5-4.5
	avgElixir := calculateAvgElixir(deckCards)
	elixirScore := 0.0
	if avgElixir >= 3.5 && avgElixir <= 4.5 {
		elixirScore = 10.0
	} else if avgElixir >= 3.2 && avgElixir <= 5.0 {
		elixirScore = 6.0
	} else if avgElixir >= 3.0 {
		elixirScore = 3.0
	}

	score = (tankScore * 0.4) + (supportScore * 0.3) + (elixirScore * 0.3)
	return score
}

// scoreControl scores a deck's fit for control archetype (0-10 scale)
// Control: Defensive buildings + big spells + defensive troops
func scoreControl(deckCards []deck.CardCandidate) float64 {
	score := 0.0
	controlWinCons := []string{"Graveyard"}
	defensiveBuildings := []string{"Tesla", "Cannon", "Inferno Tower", "Bomb Tower"}
	bigSpells := []string{"Poison", "Fireball", "Lightning", "Rocket"}

	// Check for control win conditions (35% of score)
	winConScore := 0.0
	for _, card := range deckCards {
		if slices.Contains(controlWinCons, card.Name) {
			winConScore = 10.0
		}
	}

	// Count defensive buildings (35% of score)
	buildingCount := 0
	for _, card := range deckCards {
		if card.Role != nil && *card.Role == deck.RoleBuilding {
			buildingCount++
		}
		if slices.Contains(defensiveBuildings, card.Name) {
			buildingCount++
		}
	}
	buildingScore := float64(buildingCount) * 5.0 // Max 10.0 with 2+ buildings
	if buildingScore > 10.0 {
		buildingScore = 10.0
	}

	// Count big spells (30% of score)
	spellCount := 0
	for _, card := range deckCards {
		if slices.Contains(bigSpells, card.Name) {
			spellCount++
		}
	}
	spellScore := float64(spellCount) * 5.0 // Max 10.0 with 2+ spells
	if spellScore > 10.0 {
		spellScore = 10.0
	}

	score = (winConScore * 0.35) + (buildingScore * 0.35) + (spellScore * 0.30)
	return score
}

// scoreCycle scores a deck's fit for cycle archetype (0-10 scale)
// Cycle: Low elixir + fast rotation + cycle cards
func scoreCycle(deckCards []deck.CardCandidate) float64 {
	score := 0.0
	cycleWinCons := []string{"Hog Rider", "Royal Giant", "Royal Hogs"}
	cycleCards := []string{"Skeletons", "Ice Spirit", "Ice Golem", "Electro Spirit"}

	// Check for cycle win conditions (30% of score)
	winConScore := 0.0
	for _, card := range deckCards {
		if slices.Contains(cycleWinCons, card.Name) {
			winConScore = 10.0
		}
	}

	// Count cheap cycle cards 1-2 elixir (40% of score)
	cycleCount := 0
	for _, card := range deckCards {
		if card.Elixir <= 2 {
			cycleCount++
		}
		if slices.Contains(cycleCards, card.Name) {
			cycleCount++
		}
	}
	cycleCardScore := float64(cycleCount) * 2.0 // Max 10.0 with 5+ cheap cards
	if cycleCardScore > 10.0 {
		cycleCardScore = 10.0
	}

	// Check average elixir (30% of score) - cycle typically 2.4-3.2
	avgElixir := calculateAvgElixir(deckCards)
	elixirScore := 0.0
	if avgElixir >= 2.4 && avgElixir <= 3.2 {
		elixirScore = 10.0
	} else if avgElixir >= 2.0 && avgElixir <= 3.5 {
		elixirScore = 6.0
	} else {
		elixirScore = 2.0
	}

	score = (winConScore * 0.3) + (cycleCardScore * 0.4) + (elixirScore * 0.3)
	return score
}

// scoreBridgeSpam scores a deck's fit for bridge spam archetype (0-10 scale)
// Bridge Spam: Fast units + aggressive cards + immediate pressure
func scoreBridgeSpam(deckCards []deck.CardCandidate) float64 {
	score := 0.0
	bridgeWinCons := []string{"P.E.K.K.A", "Mega Knight", "Royal Ghost", "Battle Ram"}
	spamCards := []string{"Bandit", "Royal Ghost", "Battle Ram", "Wall Breakers", "Prince"}

	// Check for bridge spam win conditions (40% of score)
	winConScore := 0.0
	for _, card := range deckCards {
		if slices.Contains(bridgeWinCons, card.Name) {
			winConScore = 10.0
		}
	}

	// Count spam cards (40% of score)
	spamCount := 0
	for _, card := range deckCards {
		if slices.Contains(spamCards, card.Name) {
			spamCount++
		}
	}
	spamScore := float64(spamCount) * 3.0 // Max 10.0 with 3+ spam cards
	if spamScore > 10.0 {
		spamScore = 10.0
	}

	// Check average elixir (20% of score) - bridge spam typically 3.0-4.0
	avgElixir := calculateAvgElixir(deckCards)
	elixirScore := 0.0
	if avgElixir >= 3.0 && avgElixir <= 4.0 {
		elixirScore = 10.0
	} else if avgElixir >= 2.8 && avgElixir <= 4.2 {
		elixirScore = 6.0
	}

	score = (winConScore * 0.4) + (spamScore * 0.4) + (elixirScore * 0.2)
	return score
}

// scoreSiege scores a deck's fit for siege archetype (0-10 scale)
// Siege: X-Bow or Mortar + defensive support
func scoreSiege(deckCards []deck.CardCandidate) float64 {
	score := 0.0
	siegeWinCons := []string{"X-Bow", "Mortar"}
	defensiveCards := []string{"Tesla", "Knight", "Archers", "Cannon"}

	// Check for siege win conditions (60% of score) - critical for siege
	winConScore := 0.0
	for _, card := range deckCards {
		if slices.Contains(siegeWinCons, card.Name) {
			winConScore = 10.0
		}
	}

	// If no siege win condition, this can't be siege
	if winConScore == 0 {
		return 0.0
	}

	// Count defensive support (40% of score)
	defenseCount := 0
	for _, card := range deckCards {
		if slices.Contains(defensiveCards, card.Name) {
			defenseCount++
		}
	}
	defenseScore := float64(defenseCount) * 2.5 // Max 10.0 with 4+ defensive cards
	if defenseScore > 10.0 {
		defenseScore = 10.0
	}

	score = (winConScore * 0.6) + (defenseScore * 0.4)
	return score
}

// scoreBait scores a deck's fit for bait archetype (0-10 scale)
// Bait: Goblin Barrel + spell bait cards + swarm units
func scoreBait(deckCards []deck.CardCandidate) float64 {
	score := 0.0
	baitWinCon := "Goblin Barrel"
	baitCards := []string{"Goblin Gang", "Princess", "Goblin Barrel", "Dart Goblin", "Goblin Drill"}

	// Check for Goblin Barrel (50% of score) - essential for bait
	winConScore := 0.0
	for _, card := range deckCards {
		if card.Name == baitWinCon {
			winConScore = 10.0
			break
		}
	}

	// If no Goblin Barrel, check for Goblin Drill as alternative
	if winConScore == 0 {
		for _, card := range deckCards {
			if card.Name == "Goblin Drill" {
				winConScore = 7.0
				break
			}
		}
	}

	// Count bait cards (50% of score)
	baitCount := 0
	for _, card := range deckCards {
		if slices.Contains(baitCards, card.Name) {
			baitCount++
		}
	}
	baitScore := float64(baitCount) * 2.5 // Max 10.0 with 4+ bait cards
	if baitScore > 10.0 {
		baitScore = 10.0
	}

	score = (winConScore * 0.5) + (baitScore * 0.5)
	return score
}

// scoreGraveyard scores a deck's fit for graveyard archetype (0-10 scale)
// Graveyard: Graveyard + defensive support + freeze/poison
func scoreGraveyard(deckCards []deck.CardCandidate) float64 {
	score := 0.0
	graveyardWinCon := "Graveyard"
	supportCards := []string{"Ice Wizard", "Baby Dragon", "Bowler", "Bomb Tower", "Knight"}
	synergies := []string{"Freeze", "Poison", "Tornado"}

	// Check for Graveyard (50% of score) - essential
	winConScore := 0.0
	for _, card := range deckCards {
		if card.Name == graveyardWinCon {
			winConScore = 10.0
			break
		}
	}

	// If no Graveyard, can't be graveyard archetype
	if winConScore == 0 {
		return 0.0
	}

	// Count support cards (30% of score)
	supportCount := 0
	for _, card := range deckCards {
		if slices.Contains(supportCards, card.Name) {
			supportCount++
		}
	}
	supportScore := float64(supportCount) * 3.0 // Max 10.0 with 3+ supports
	if supportScore > 10.0 {
		supportScore = 10.0
	}

	// Count synergy spells (20% of score)
	synergyCount := 0
	for _, card := range deckCards {
		if slices.Contains(synergies, card.Name) {
			synergyCount++
		}
	}
	synergyScore := float64(synergyCount) * 5.0 // Max 10.0 with 2+ synergies
	if synergyScore > 10.0 {
		synergyScore = 10.0
	}

	score = (winConScore * 0.5) + (supportScore * 0.3) + (synergyScore * 0.2)
	return score
}

// scoreMiner scores a deck's fit for miner archetype (0-10 scale)
// Miner: Miner + poison/cycle support
func scoreMiner(deckCards []deck.CardCandidate) float64 {
	score := 0.0
	minerWinCon := "Miner"
	supportCards := []string{"Poison", "Valkyrie", "Electro Wizard", "Ice Golem"}

	// Check for Miner (60% of score) - essential
	winConScore := 0.0
	for _, card := range deckCards {
		if card.Name == minerWinCon {
			winConScore = 10.0
			break
		}
	}

	// If no Miner, can't be miner archetype
	if winConScore == 0 {
		return 0.0
	}

	// Count support cards (40% of score)
	supportCount := 0
	for _, card := range deckCards {
		if slices.Contains(supportCards, card.Name) {
			supportCount++
		}
	}
	supportScore := float64(supportCount) * 3.0 // Max 10.0 with 3+ supports
	if supportScore > 10.0 {
		supportScore = 10.0
	}

	score = (winConScore * 0.6) + (supportScore * 0.4)
	return score
}

// calculateAvgElixir calculates the average elixir cost of a deck
func calculateAvgElixir(deckCards []deck.CardCandidate) float64 {
	if len(deckCards) == 0 {
		return 0.0
	}

	total := 0
	for _, card := range deckCards {
		total += card.Elixir
	}

	return float64(total) / float64(len(deckCards))
}
