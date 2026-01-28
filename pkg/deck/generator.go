// Package deck provides deck generation with multiple sampling strategies
package deck

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// GeneratorStrategy defines the approach for generating decks
type GeneratorStrategy string

const (
	// StrategyExhaustive iterates through all valid deck combinations
	StrategyExhaustive GeneratorStrategy = "exhaustive"

	// StrategySmartSample prioritizes high-level cards and known synergies
	StrategySmartSample GeneratorStrategy = "smart-sample"

	// StrategyRandomSample generates N random valid decks
	StrategyRandomSample GeneratorStrategy = "random-sample"

	// StrategyArchetypeFocused explores specific archetype space
	StrategyArchetypeFocused GeneratorStrategy = "archetype-focused"

	// StrategyGenetic uses genetic algorithm for evolutionary deck optimization
	StrategyGenetic GeneratorStrategy = "genetic"
)

// GeneratorConfig configures deck generation behavior
type GeneratorConfig struct {
	// Strategy defines the generation approach
	Strategy GeneratorStrategy

	// Candidates are the available cards to choose from
	Candidates []*CardCandidate

	// Composition defines role requirements (optional, uses defaults if nil)
	Composition *RoleComposition

	// Constraints defines deck constraints
	Constraints *GeneratorConstraints

	// Seed for reproducible random generation (0 = random seed)
	Seed int64

	// SampleSize for sampling strategies (smart/random)
	SampleSize int

	// Workers for parallel generation (default: 1)
	Workers int

	// Archetype for archetype-focused strategy
	Archetype string

	// Progress callback for long-running operations
	Progress func(GeneratorProgress)
}

// GeneratorConstraints defines validation rules for generated decks
type GeneratorConstraints struct {
	// MinAvgElixir minimum average elixir cost
	MinAvgElixir float64

	// MaxAvgElixir maximum average elixir cost
	MaxAvgElixir float64

	// IncludeCards forces specific cards in every deck
	IncludeCards []string

	// ExcludeCards removes specific cards from candidate pool
	ExcludeCards []string

	// RequireWinCondition ensures at least one win condition (default: true)
	RequireWinCondition bool

	// MinEvolutionCards minimum evolved cards (0 = no requirement)
	MinEvolutionCards int

	// PreferHighLevel prioritizes higher-level cards in sampling
	PreferHighLevel bool
}

// GeneratorProgress reports generation progress
type GeneratorProgress struct {
	Generated int
	Total     int
	Valid     int
	Invalid   int
}

// GeneratorCheckpoint stores iteration state for resumption
type GeneratorCheckpoint struct {
	// Strategy identifies the generator type
	Strategy GeneratorStrategy

	// Position tracks iteration progress (strategy-specific)
	Position int64

	// Generated tracks total decks generated so far
	Generated int

	// State holds strategy-specific state (e.g., combination indices)
	State map[string]any
}

// DeckIterator provides resumable iteration over generated decks
type DeckIterator interface {
	// Next returns the next generated deck or nil when exhausted
	Next(ctx context.Context) ([]string, error)

	// Checkpoint saves current iteration state
	Checkpoint() *GeneratorCheckpoint

	// Resume restores iteration from checkpoint
	Resume(checkpoint *GeneratorCheckpoint) error

	// Reset restarts iteration from beginning
	Reset()

	// Close releases resources
	Close() error
}

// DeckGenerator orchestrates deck generation with multiple strategies
type DeckGenerator struct {
	config      GeneratorConfig
	candidates  []*CardCandidate
	candidatesByRole map[CardRole][]*CardCandidate
	includeMap  map[string]bool
	excludeMap  map[string]bool
	composition RoleComposition
	rng         *rand.Rand
	mu          sync.RWMutex
}

// NewDeckGenerator creates a new generator with the given configuration
func NewDeckGenerator(config GeneratorConfig) (*DeckGenerator, error) {
	if len(config.Candidates) == 0 {
		return nil, ErrInsufficientCards
	}

	if config.Constraints == nil {
		config.Constraints = &GeneratorConstraints{
			MinAvgElixir:        2.0,
			MaxAvgElixir:        5.0,
			RequireWinCondition: true,
		}
	}

	// Set default composition if not provided
	composition := *DefaultRoleComposition()
	if config.Composition != nil {
		composition = *config.Composition
	}

	// Initialize RNG with seed
	seed := config.Seed
	if seed == 0 {
		seed = rand.Int63()
	}

	gen := &DeckGenerator{
		config:      config,
		composition: composition,
		rng:         rand.New(rand.NewSource(seed)),
		includeMap:  make(map[string]bool),
		excludeMap:  make(map[string]bool),
		candidatesByRole: make(map[CardRole][]*CardCandidate),
	}

	// Build include/exclude maps
	for _, card := range config.Constraints.IncludeCards {
		gen.includeMap[strings.TrimSpace(card)] = true
	}
	for _, card := range config.Constraints.ExcludeCards {
		gen.excludeMap[strings.TrimSpace(card)] = true
	}

	// Filter and organize candidates
	gen.candidates = make([]*CardCandidate, 0, len(config.Candidates))
	for _, card := range config.Candidates {
		// Skip excluded cards
		if gen.excludeMap[card.Name] {
			continue
		}

		gen.candidates = append(gen.candidates, card)

		// Organize by role
		if card.Role != nil {
			role := *card.Role
			gen.candidatesByRole[role] = append(gen.candidatesByRole[role], card)
		}
	}

	if len(gen.candidates) < 8 {
		return nil, ErrInsufficientCards
	}

	// Sort candidates by score (descending) for efficient sampling
	sort.Slice(gen.candidates, func(i, j int) bool {
		return gen.candidates[i].Score > gen.candidates[j].Score
	})

	return gen, nil
}

// Iterator creates a DeckIterator for the configured strategy
func (g *DeckGenerator) Iterator() (DeckIterator, error) {
	switch g.config.Strategy {
	case StrategyExhaustive:
		return newExhaustiveIterator(g), nil
	case StrategySmartSample:
		return newSmartSampleIterator(g), nil
	case StrategyRandomSample:
		return newRandomSampleIterator(g), nil
	case StrategyArchetypeFocused:
		return newArchetypeIterator(g), nil
	case StrategyGenetic:
		return newGeneticIterator(g), nil
	default:
		return nil, fmt.Errorf("unsupported strategy: %s", g.config.Strategy)
	}
}

// GenerateOne generates a single deck using the configured strategy
func (g *DeckGenerator) GenerateOne(ctx context.Context) ([]string, error) {
	iterator, err := g.Iterator()
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	return iterator.Next(ctx)
}

// Generate generates multiple decks using the configured strategy
func (g *DeckGenerator) Generate(ctx context.Context, count int) ([][]string, error) {
	iterator, err := g.Iterator()
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	decks := make([][]string, 0, count)
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return decks, ctx.Err()
		default:
		}

		deck, err := iterator.Next(ctx)
		if err != nil {
			return decks, err
		}
		if deck == nil {
			break // Iterator exhausted
		}

		decks = append(decks, deck)

		// Report progress
		if g.config.Progress != nil {
			g.config.Progress(GeneratorProgress{
				Generated: i + 1,
				Total:     count,
				Valid:     len(decks),
			})
		}
	}

	return decks, nil
}

// validateDeck checks if a deck meets all constraints
func (g *DeckGenerator) validateDeck(deck []string) error {
	if len(deck) != 8 {
		return ErrInvalidDeckSize
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, card := range deck {
		if seen[card] {
			return fmt.Errorf("duplicate card: %s", card)
		}
		seen[card] = true
	}

	// Build candidate map for quick lookup
	candidateMap := make(map[string]*CardCandidate)
	for _, c := range g.candidates {
		candidateMap[c.Name] = c
	}

	// Calculate average elixir and check evolution count
	totalElixir := 0
	evolutionCount := 0
	hasWinCondition := false

	for _, cardName := range deck {
		card, ok := candidateMap[cardName]
		if !ok {
			return fmt.Errorf("card not found in candidates: %s", cardName)
		}

		totalElixir += card.Elixir

		if card.HasEvolution || card.EvolutionLevel > 0 {
			evolutionCount++
		}

		if card.Role != nil && *card.Role == RoleWinCondition {
			hasWinCondition = true
		}
	}

	avgElixir := float64(totalElixir) / 8.0

	// Validate elixir range
	if avgElixir < g.config.Constraints.MinAvgElixir {
		return fmt.Errorf("average elixir %.2f below minimum %.2f", avgElixir, g.config.Constraints.MinAvgElixir)
	}
	if avgElixir > g.config.Constraints.MaxAvgElixir {
		return fmt.Errorf("average elixir %.2f above maximum %.2f", avgElixir, g.config.Constraints.MaxAvgElixir)
	}

	// Validate win condition requirement
	if g.config.Constraints.RequireWinCondition && !hasWinCondition {
		return ErrNoWinCondition
	}

	// Validate evolution requirement
	if evolutionCount < g.config.Constraints.MinEvolutionCards {
		return fmt.Errorf("only %d evolution cards, need at least %d", evolutionCount, g.config.Constraints.MinEvolutionCards)
	}

	// Validate include cards are present
	for card := range g.includeMap {
		if !seen[card] {
			return fmt.Errorf("required card missing: %s", card)
		}
	}

	return nil
}

// selectRandomCardsFromRole selects N random cards from a specific role
func (g *DeckGenerator) selectRandomCardsFromRole(role CardRole, count int, used map[string]bool) []string {
	candidates := g.candidatesByRole[role]
	if len(candidates) == 0 {
		return nil
	}

	selected := make([]string, 0, count)
	shuffled := make([]*CardCandidate, len(candidates))
	copy(shuffled, candidates)

	// Fisher-Yates shuffle
	g.mu.Lock()
	for i := len(shuffled) - 1; i > 0; i-- {
		j := g.rng.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	g.mu.Unlock()

	for _, card := range shuffled {
		if used[card.Name] {
			continue
		}
		selected = append(selected, card.Name)
		used[card.Name] = true
		if len(selected) >= count {
			break
		}
	}

	return selected
}

// selectBestCardsFromRole selects top N cards from a role by score
func (g *DeckGenerator) selectBestCardsFromRole(role CardRole, count int, used map[string]bool) []string {
	candidates := g.candidatesByRole[role]
	if len(candidates) == 0 {
		return nil
	}

	// Candidates are already sorted by score
	selected := make([]string, 0, count)
	for _, card := range candidates {
		if used[card.Name] {
			continue
		}
		selected = append(selected, card.Name)
		used[card.Name] = true
		if len(selected) >= count {
			break
		}
	}

	return selected
}

// fillRemainingSlots fills remaining deck slots with highest-scoring available cards
func (g *DeckGenerator) fillRemainingSlots(deckSize int, used map[string]bool) []string {
	remaining := make([]string, 0, 8-deckSize)
	for _, card := range g.candidates {
		if used[card.Name] {
			continue
		}
		remaining = append(remaining, card.Name)
		used[card.Name] = true
		if len(remaining)+deckSize >= 8 {
			break
		}
	}
	return remaining
}

// NewGeneratorConfigFromPlayer creates a generator config from player data
func NewGeneratorConfigFromPlayer(player *clashroyale.Player, strategy GeneratorStrategy) (*GeneratorConfig, error) {
	// Convert player cards to candidates (using existing scoring logic)
	candidates := make([]*CardCandidate, 0, len(player.Cards))
	for i := range player.Cards {
		card := player.Cards[i]
		candidate := &CardCandidate{
			Name:              card.Name,
			Level:             card.Level,
			MaxLevel:          card.MaxLevel,
			Elixir:            0, // Would need card metadata
			HasEvolution:      card.MaxEvolutionLevel > 0,
			EvolutionLevel:    card.EvolutionLevel,
			MaxEvolutionLevel: card.MaxEvolutionLevel,
		}
		candidates = append(candidates, candidate)
	}

	return &GeneratorConfig{
		Strategy:   strategy,
		Candidates: candidates,
		Constraints: &GeneratorConstraints{
			MinAvgElixir:        2.0,
			MaxAvgElixir:        5.0,
			RequireWinCondition: true,
		},
		SampleSize: 1000,
		Workers:    1,
	}, nil
}
