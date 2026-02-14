// Package genetic provides genetic algorithm-based deck optimization
// using the eaopt library for evolutionary deck generation.
package genetic

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
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

	// fitnessEvaluator overrides default Evaluate behavior when set.
	fitnessEvaluator func([]deck.CardCandidate) (float64, error)
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

// getCardCandidates converts the genome's card names to CardCandidate instances.
// It looks up each card name in the candidates pool.
func (g *DeckGenome) getCardCandidates() []deck.CardCandidate {
	cardMap := make(map[string]*deck.CardCandidate)
	for _, c := range g.candidates {
		cardMap[c.Name] = c
	}

	result := make([]deck.CardCandidate, 0, 8)
	for _, name := range g.Cards {
		if c, ok := cardMap[name]; ok {
			result = append(result, *c)
		}
	}
	return result
}

// Evaluate calculates the fitness of this deck genome.
// Higher fitness indicates a better deck according to the strategy.
//
// This method implements the eaopt.Genome interface requirement.
func (g *DeckGenome) Evaluate() (float64, error) {
	// Convert card names to CardCandidate slice
	deckCards := g.getCardCandidates()
	if len(deckCards) != 8 {
		return 0, fmt.Errorf("failed to resolve all cards: got %d, want 8", len(deckCards))
	}

	if cached, ok := getCachedFitness(g.Cards); ok {
		g.Fitness = cached
		return g.Fitness, nil
	}

	if g.fitnessEvaluator != nil {
		fitness, err := g.fitnessEvaluator(deckCards)
		if err != nil {
			return 0, err
		}
		g.Fitness = fitness
		storeCachedFitness(g.Cards, g.Fitness)
		return g.Fitness, nil
	}

	// Create synergy database for evaluation
	synergyDB := deck.NewSynergyDatabase()

	// Run full deck evaluation (no player context for genetic algorithm)
	result := evaluation.Evaluate(deckCards, synergyDB, nil)

	// Use OverallScore (0-10 scale) as fitness
	g.Fitness = result.OverallScore
	storeCachedFitness(g.Cards, g.Fitness)

	return g.Fitness, nil
}

// Clone creates a deep copy of this genome.
//
// This method implements the eaopt.Genome interface requirement.
func (g *DeckGenome) Clone() interface{} {
	cards := make([]string, 8)
	copy(cards, g.Cards)

	return &DeckGenome{
		Cards:            cards,
		Fitness:          g.Fitness,
		config:           g.config,
		candidates:       g.candidates,
		strategy:         g.strategy,
		fitnessEvaluator: g.fitnessEvaluator,
	}
}

// String returns a string representation of the deck.
func (g *DeckGenome) String() string {
	return fmt.Sprintf("Deck{%s, Fitness:%.4f}", strings.Join(g.Cards, ", "), g.Fitness)
}

// randomInt returns a random integer in [0, n).
func randomInt(n int) int {
	return rand.IntN(n)
}
