// Package genetic provides genetic algorithm-based deck optimization
// using the eaopt library for evolutionary deck generation.
package genetic

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// DeckGenome represents an individual deck in the genetic algorithm population.
// It implements the eaopt.Genome interface for use with the eaopt library.
type DeckGenome struct {
	// Cards is the list of 8 card names in the deck.
	Cards []string

	// Fitness is the deck's fitness score (higher is better).
	// Calculated by the Evaluate method.
	Fitness float64

	// config holds the genetic algorithm configuration.
	config *GeneticConfig

	// candidates is the pool of cards available for deck building.
	candidates []*deck.CardCandidate

	// strategy is the deck building strategy to use for fitness evaluation.
	strategy deck.Strategy
}

// NewDeckGenome creates a new random deck genome from the available candidates.
// The deck is randomly selected while respecting role constraints.
func NewDeckGenome(candidates []*deck.CardCandidate, strategy deck.Strategy, config *GeneticConfig) (*DeckGenome, error) {
	if len(candidates) < 8 {
		return nil, fmt.Errorf("insufficient candidates: need at least 8 cards, got %d", len(candidates))
	}

	g := &DeckGenome{
		Cards:      make([]string, 8),
		config:     config,
		candidates: candidates,
		strategy:   strategy,
	}

	// Initialize with random valid deck
	if err := g.initializeRandomDeck(); err != nil {
		return nil, fmt.Errorf("failed to initialize random deck: %w", err)
	}

	return g, nil
}

// NewDeckGenomeFromCards creates a genome from a specific set of card names.
// Useful for seeding the initial population with known good decks.
func NewDeckGenomeFromCards(cardNames []string, candidates []*deck.CardCandidate, strategy deck.Strategy, config *GeneticConfig) (*DeckGenome, error) {
	if len(cardNames) != 8 {
		return nil, fmt.Errorf("deck must have exactly 8 cards, got %d", len(cardNames))
	}

	// Validate all cards exist in candidates
	cardMap := make(map[string]*deck.CardCandidate)
	for _, c := range candidates {
		cardMap[c.Name] = c
	}

	for _, name := range cardNames {
		if _, exists := cardMap[name]; !exists {
			return nil, fmt.Errorf("card %q not found in candidates", name)
		}
	}

	g := &DeckGenome{
		Cards:      make([]string, 8),
		config:     config,
		candidates: candidates,
		strategy:   strategy,
	}
	copy(g.Cards, cardNames)

	return g, nil
}

// initializeRandomDeck creates a random valid deck from the candidate pool.
// It ensures the deck has at least one win condition and respects role diversity.
func (g *DeckGenome) initializeRandomDeck() error {
	// Group candidates by role for balanced selection
	byRole := make(map[deck.CardRole][]*deck.CardCandidate)
	for _, c := range g.candidates {
		if c.HasRole() {
			byRole[*c.Role] = append(byRole[*c.Role], c)
		}
	}

	// Simple random selection for now - more sophisticated logic can be added later
	// This is a basic implementation that ensures 8 unique cards
	selected := make(map[string]bool)
	var cards []string

	// Try to include at least one win condition
	if winConditions, ok := byRole[deck.RoleWinCondition]; ok && len(winConditions) > 0 {
		for len(cards) < 1 {
			idx := randomInt(len(winConditions))
			card := winConditions[idx]
			if !selected[card.Name] {
				cards = append(cards, card.Name)
				selected[card.Name] = true
			}
		}
	}

	// Fill remaining slots with random cards
	allCards := make([]*deck.CardCandidate, 0, len(g.candidates))
	for _, c := range g.candidates {
		if !selected[c.Name] {
			allCards = append(allCards, c)
		}
	}

	for len(cards) < 8 && len(allCards) > 0 {
		idx := randomInt(len(allCards))
		card := allCards[idx]
		cards = append(cards, card.Name)
		// Remove selected card
		allCards = append(allCards[:idx], allCards[idx+1:]...)
	}

	if len(cards) != 8 {
		return fmt.Errorf("failed to select 8 unique cards, got %d", len(cards))
	}

	g.Cards = cards
	return nil
}

// Evaluate calculates the fitness of this deck genome.
// Higher fitness indicates a better deck according to the strategy.
//
// This method implements the eaopt.Genome interface requirement.
func (g *DeckGenome) Evaluate() (float64, error) {
	// TODO: Implement fitness calculation using deck scoring system
	// This will integrate with the existing pkg/deck/scorer functionality
	//
	// Fitness factors:
	// 1. Strategy compatibility (from deck.Scoring)
	// 2. Role balance and archetype validity
	// 3. Synergy between cards
	// 4. Average elixir cost appropriateness
	// 5. Evolution potential

	g.Fitness = 0.5 // Placeholder
	return g.Fitness, nil
}

// Mutate applies random mutations to the deck genome.
// The mutation intensity determines how many cards are replaced.
//
// This method implements the eaopt.Genome interface requirement.
func (g *DeckGenome) Mutate() error {
	// Calculate number of cards to mutate based on intensity
	numToMutate := int(float64(8) * g.config.MutationIntensity)
	if numToMutate < 1 {
		numToMutate = 1
	}

	// Select random positions to mutate
	positions := make(map[int]bool)
	for len(positions) < numToMutate {
		pos := randomInt(8)
		positions[pos] = true
	}

	// Build map of current cards
	currentCards := make(map[string]bool)
	for _, c := range g.Cards {
		currentCards[c] = true
	}

	// Find replacement candidates
	for pos := range positions {
		oldCard := g.Cards[pos]
		delete(currentCards, oldCard)

		// Find a card not currently in deck and not the old card
		replaced := false
		for _, candidate := range g.candidates {
			if !currentCards[candidate.Name] && candidate.Name != oldCard {
				g.Cards[pos] = candidate.Name
				currentCards[candidate.Name] = true
				replaced = true
				break
			}
		}

		// If we couldn't find a different card (e.g., all candidates are in deck),
		// put the old card back
		if !replaced {
			g.Cards[pos] = oldCard
			currentCards[oldCard] = true
		}
	}

	// Reset fitness after mutation
	g.Fitness = 0
	return nil
}

// Clone creates a deep copy of this genome.
//
// This method implements the eaopt.Genome interface requirement.
func (g *DeckGenome) Clone() interface{} {
	cards := make([]string, 8)
	copy(cards, g.Cards)

	return &DeckGenome{
		Cards:      cards,
		Fitness:    g.Fitness,
		config:     g.config,
		candidates: g.candidates,
		strategy:   g.strategy,
	}
}

// Crossover creates offspring by combining this genome with another.
// Uses uniform crossover - each card position has 50% chance from each parent.
//
// This method implements the eaopt.Genome interface requirement.
func (g *DeckGenome) Crossover(other interface{}) (interface{}, error) {
	otherDeck, ok := other.(*DeckGenome)
	if !ok {
		return nil, fmt.Errorf("crossover requires DeckGenome, got %T", other)
	}

	offspring := &DeckGenome{
		Cards:      make([]string, 8),
		config:     g.config,
		candidates: g.candidates,
		strategy:   g.strategy,
	}

	// Uniform crossover with duplicate resolution
	parent1Cards := make(map[string]bool)
	for _, c := range g.Cards {
		parent1Cards[c] = true
	}

	parent2Cards := make(map[string]bool)
	for _, c := range otherDeck.Cards {
		parent2Cards[c] = true
	}

	usedCards := make(map[string]bool)

	for i := 0; i < 8; i++ {
		// Randomly choose from either parent
		useParent1 := randomInt(2) == 0

		var card string
		if useParent1 {
			card = g.Cards[i]
			if usedCards[card] {
				// Card already used, try other parent
				card = otherDeck.Cards[i]
			}
		} else {
			card = otherDeck.Cards[i]
			if usedCards[card] {
				card = g.Cards[i]
			}
		}

		// If still duplicate, pick from remaining cards
		if usedCards[card] {
			card = g.findUnusedCard(parent1Cards, parent2Cards, usedCards)
		}

		offspring.Cards[i] = card
		usedCards[card] = true
	}

	return offspring, nil
}

// findUnusedCard finds a card not yet used, preferring cards from either parent.
func (g *DeckGenome) findUnusedCard(parent1, parent2, used map[string]bool) string {
	// First try parent1 cards
	for _, c := range g.Cards {
		if !used[c] {
			return c
		}
	}
	// Then try parent2 cards
	for _, c := range g.Cards {
		if !used[c] {
			return c
		}
	}
	// Finally, pick from all candidates
	for _, c := range g.candidates {
		if !used[c.Name] {
			return c.Name
		}
	}
	// Fallback (shouldn't happen with valid input)
	return ""
}

// String returns a string representation of the deck.
func (g *DeckGenome) String() string {
	return fmt.Sprintf("Deck{%s, Fitness:%.4f}", strings.Join(g.Cards, ", "), g.Fitness)
}

// randomInt returns a random integer in [0, n).
func randomInt(n int) int {
	return rand.IntN(n)
}
