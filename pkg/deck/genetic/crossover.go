// Package genetic provides genetic algorithm-based deck optimization
// using the eaopt library for evolutionary deck generation.
package genetic

import (
	"fmt"
	"sort"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

const (
	synergyCrossoverThreshold = 0.8
	roleUnknown               = "unknown"
)

// Crossover creates offspring by combining this genome with another.
// It randomly selects between multiple crossover operators and repairs the deck.
//
// This method implements the eaopt.Genome interface requirement.
func (g *DeckGenome) Crossover(other any) (any, error) {
	otherDeck, ok := other.(*DeckGenome)
	if !ok {
		return nil, fmt.Errorf("crossover requires DeckGenome, got %T", other)
	}

	var cards []string
	switch randomInt(3) {
	case 0:
		cards = g.uniformCrossover(otherDeck)
	case 1:
		cards = g.rolePreservingCrossover(otherDeck)
	default:
		cards = g.synergyAwareCrossover(otherDeck)
	}

	offspring := &DeckGenome{
		Cards:      g.repairDeck(cards, otherDeck),
		config:     g.config,
		candidates: g.candidates,
		strategy:   g.strategy,
	}

	return offspring, nil
}

func (g *DeckGenome) uniformCrossover(other *DeckGenome) []string {
	offspring := make([]string, 0, 8)
	for i := range 8 {
		if randomInt(2) == 0 {
			offspring = append(offspring, g.Cards[i])
		} else {
			offspring = append(offspring, other.Cards[i])
		}
	}
	return offspring
}

func (g *DeckGenome) rolePreservingCrossover(other *DeckGenome) []string {
	candidateMap := g.candidateMap()
	parent1Roles := g.cardsByRole(g.Cards, candidateMap)
	parent2Roles := g.cardsByRole(other.Cards, candidateMap)

	roleSet := make(map[string]struct{})
	for role := range parent1Roles {
		roleSet[role] = struct{}{}
	}
	for role := range parent2Roles {
		roleSet[role] = struct{}{}
	}

	offspring := make([]string, 0, 8)
	for role := range roleSet {
		if randomInt(2) == 0 {
			offspring = append(offspring, parent1Roles[role]...)
		} else {
			offspring = append(offspring, parent2Roles[role]...)
		}
	}

	return offspring
}

func (g *DeckGenome) synergyAwareCrossover(other *DeckGenome) []string {
	db := deck.NewSynergyDatabase()
	pairs := append(g.findSynergyPairs(g.Cards, db), g.findSynergyPairs(other.Cards, db)...)

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Score > pairs[j].Score
	})

	offspring := make([]string, 0, 8)
	used := make(map[string]bool)

	for _, pair := range pairs {
		if len(offspring) > 6 {
			break
		}
		if used[pair.Card1] || used[pair.Card2] {
			continue
		}
		offspring = append(offspring, pair.Card1, pair.Card2)
		used[pair.Card1] = true
		used[pair.Card2] = true
	}

	parentPool := append([]string{}, g.Cards...)
	parentPool = append(parentPool, other.Cards...)

	for len(offspring) < 8 && len(parentPool) > 0 {
		idx := randomInt(len(parentPool))
		card := parentPool[idx]
		parentPool = append(parentPool[:idx], parentPool[idx+1:]...)
		if used[card] {
			continue
		}
		offspring = append(offspring, card)
		used[card] = true
	}

	return offspring
}

func (g *DeckGenome) repairDeck(cards []string, parents ...*DeckGenome) []string {
	candidateMap := g.candidateMap()
	used := make(map[string]bool)
	repaired := make([]string, 0, 8)

	addCard := func(name string) {
		if len(repaired) >= 8 {
			return
		}
		if used[name] {
			return
		}
		if _, ok := candidateMap[name]; !ok {
			return
		}
		repaired = append(repaired, name)
		used[name] = true
	}

	for _, card := range cards {
		addCard(card)
	}

	for _, parent := range parents {
		if parent == nil {
			continue
		}
		for _, card := range parent.Cards {
			addCard(card)
		}
	}

	if len(repaired) < 8 {
		remaining := make([]*deck.CardCandidate, 0, len(g.candidates))
		for _, candidate := range g.candidates {
			if !used[candidate.Name] {
				remaining = append(remaining, candidate)
			}
		}

		for len(repaired) < 8 && len(remaining) > 0 {
			idx := randomInt(len(remaining))
			addCard(remaining[idx].Name)
			remaining = append(remaining[:idx], remaining[idx+1:]...)
		}
	}

	return g.ensureWinCondition(repaired, used, candidateMap)
}

func (g *DeckGenome) ensureWinCondition(cards []string, used map[string]bool, candidateMap map[string]*deck.CardCandidate) []string {
	for _, card := range cards {
		if g.isWinCondition(card, candidateMap) {
			return cards
		}
	}

	var winConditions []string
	for _, candidate := range g.candidates {
		if used[candidate.Name] {
			continue
		}
		if candidate.Role != nil && *candidate.Role == deck.RoleWinCondition {
			winConditions = append(winConditions, candidate.Name)
		}
	}
	if len(winConditions) == 0 || len(cards) == 0 {
		return cards
	}

	cards[len(cards)-1] = winConditions[randomInt(len(winConditions))]
	return cards
}

func (g *DeckGenome) isWinCondition(card string, candidateMap map[string]*deck.CardCandidate) bool {
	candidate, ok := candidateMap[card]
	if !ok || candidate.Role == nil {
		return deck.IsWinCondition(card)
	}
	return *candidate.Role == deck.RoleWinCondition
}

func (g *DeckGenome) candidateMap() map[string]*deck.CardCandidate {
	cardMap := make(map[string]*deck.CardCandidate, len(g.candidates))
	for _, c := range g.candidates {
		cardMap[c.Name] = c
	}
	return cardMap
}

func (g *DeckGenome) cardsByRole(cards []string, candidateMap map[string]*deck.CardCandidate) map[string][]string {
	byRole := make(map[string][]string)
	for _, card := range cards {
		roleKey := roleUnknown
		if candidate, ok := candidateMap[card]; ok && candidate.Role != nil {
			roleKey = string(*candidate.Role)
		}
		byRole[roleKey] = append(byRole[roleKey], card)
	}
	return byRole
}

func (g *DeckGenome) findSynergyPairs(cards []string, db *deck.SynergyDatabase) []deck.SynergyPair {
	var pairs []deck.SynergyPair
	for i := range cards {
		for j := i + 1; j < len(cards); j++ {
			score := db.GetSynergy(cards[i], cards[j])
			if score < synergyCrossoverThreshold {
				continue
			}
			if pair := db.GetSynergyPair(cards[i], cards[j]); pair != nil {
				pairs = append(pairs, *pair)
			} else {
				pairs = append(pairs, deck.SynergyPair{
					Card1: cards[i],
					Card2: cards[j],
					Score: score,
				})
			}
		}
	}
	return pairs
}
