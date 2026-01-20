// Package genetic provides genetic algorithm-based deck optimization
// using the eaopt library for evolutionary deck generation.
package genetic

import (
	"fmt"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

const (
	synergyMutationThreshold = 0.6
)

// Mutate applies random mutations to the deck genome.
// The mutation intensity determines how many cards are replaced.
//
// This method implements the eaopt.Genome interface requirement.
func (g *DeckGenome) Mutate() error {
	if g.config == nil {
		return fmt.Errorf("mutation requires config")
	}

	numToMutate := int(float64(8) * g.config.MutationIntensity)
	if numToMutate < 1 {
		numToMutate = 1
	}

	positions := g.pickMutationPositions(numToMutate)
	used := g.currentCardSet()

	for _, pos := range positions {
		oldCard := g.Cards[pos]
		delete(used, oldCard)

		var replacement string
		switch randomInt(5) {
		case 0:
			replacement = g.singleCardSwap(used)
		case 1:
			replacement = g.roleBasedSwap(oldCard, used)
		case 2:
			replacement = g.synergyGuidedSwap(oldCard, used)
		case 3:
			replacement = g.evolutionAwareSwap(oldCard, used)
		default:
			replacement = g.mixedMutationSwap(oldCard, used)
		}

		if replacement == "" || replacement == oldCard {
			g.Cards[pos] = oldCard
			used[oldCard] = true
			continue
		}

		g.Cards[pos] = replacement
		used[replacement] = true
	}

	g.Cards = g.repairDeck(g.Cards, g)
	g.Fitness = 0
	return nil
}

func (g *DeckGenome) pickMutationPositions(count int) []int {
	positions := make(map[int]struct{})
	for len(positions) < count {
		positions[randomInt(8)] = struct{}{}
	}

	result := make([]int, 0, len(positions))
	for pos := range positions {
		result = append(result, pos)
	}
	return result
}

func (g *DeckGenome) currentCardSet() map[string]bool {
	used := make(map[string]bool, 8)
	for _, card := range g.Cards {
		used[card] = true
	}
	return used
}

func (g *DeckGenome) singleCardSwap(used map[string]bool) string {
	var options []string
	for _, candidate := range g.candidates {
		if !used[candidate.Name] {
			options = append(options, candidate.Name)
		}
	}
	if len(options) == 0 {
		return ""
	}
	return options[randomInt(len(options))]
}

func (g *DeckGenome) roleBasedSwap(oldCard string, used map[string]bool) string {
	candidateMap := g.candidateMap()
	var targetRole *deck.CardRole
	if candidate, ok := candidateMap[oldCard]; ok {
		targetRole = candidate.Role
	}

	var options []string
	for _, candidate := range g.candidates {
		if used[candidate.Name] {
			continue
		}
		if targetRole == nil {
			options = append(options, candidate.Name)
			continue
		}
		if candidate.Role != nil && *candidate.Role == *targetRole {
			options = append(options, candidate.Name)
		}
	}

	if len(options) == 0 {
		return g.singleCardSwap(used)
	}
	return options[randomInt(len(options))]
}

func (g *DeckGenome) synergyGuidedSwap(oldCard string, used map[string]bool) string {
	db := deck.NewSynergyDatabase()
	baselineScore := g.scoreSynergyWithDeck(oldCard, db)

	bestCandidate := ""
	bestScore := baselineScore
	for _, candidate := range g.candidates {
		if used[candidate.Name] {
			continue
		}
		score := g.scoreSynergyWithDeck(candidate.Name, db)
		if score >= bestScore+synergyMutationThreshold {
			bestScore = score
			bestCandidate = candidate.Name
		}
	}

	if bestCandidate == "" {
		return g.roleBasedSwap(oldCard, used)
	}
	return bestCandidate
}

func (g *DeckGenome) evolutionAwareSwap(oldCard string, used map[string]bool) string {
	var evolved []string
	var normal []string

	for _, candidate := range g.candidates {
		if used[candidate.Name] {
			continue
		}
		if candidate.HasEvolution || candidate.EvolutionLevel > 0 {
			evolved = append(evolved, candidate.Name)
		} else {
			normal = append(normal, candidate.Name)
		}
	}

	if len(evolved) > 0 && randomInt(100) < 70 {
		return evolved[randomInt(len(evolved))]
	}
	if len(normal) > 0 {
		return normal[randomInt(len(normal))]
	}
	return g.singleCardSwap(used)
}

func (g *DeckGenome) mixedMutationSwap(oldCard string, used map[string]bool) string {
	switch randomInt(3) {
	case 0:
		return g.roleBasedSwap(oldCard, used)
	case 1:
		return g.synergyGuidedSwap(oldCard, used)
	default:
		return g.evolutionAwareSwap(oldCard, used)
	}
}

func (g *DeckGenome) scoreSynergyWithDeck(card string, db *deck.SynergyDatabase) float64 {
	total := 0.0
	count := 0
	for _, other := range g.Cards {
		if other == card {
			continue
		}
		total += db.GetSynergy(card, other)
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}
