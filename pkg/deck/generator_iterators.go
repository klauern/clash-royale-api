package deck

import (
	"context"
	"fmt"
	"math"
	"math/rand"
)

// exhaustiveIterator generates all possible valid deck combinations
type exhaustiveIterator struct {
	gen        *DeckGenerator
	indices    []int // Current combination indices
	n          int   // Total candidates
	k          int   // Cards per deck (8)
	done       bool
	generated  int
}

func newExhaustiveIterator(gen *DeckGenerator) *exhaustiveIterator {
	n := len(gen.candidates)
	k := 8

	// Initialize indices for first combination
	indices := make([]int, k)
	for i := range indices {
		indices[i] = i
	}

	return &exhaustiveIterator{
		gen:     gen,
		indices: indices,
		n:       n,
		k:       k,
		done:    n < k,
	}
}

func (it *exhaustiveIterator) Next(ctx context.Context) ([]string, error) {
	if it.done {
		return nil, nil
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Build deck from current indices
	deck := make([]string, it.k)
	for i, idx := range it.indices {
		deck[i] = it.gen.candidates[idx].Name
	}

	// Advance to next combination
	if !it.nextCombination() {
		it.done = true
	}

	it.generated++

	// Validate the deck
	if err := it.gen.validateDeck(deck); err != nil {
		// Skip invalid decks and try next
		return it.Next(ctx)
	}

	return deck, nil
}

func (it *exhaustiveIterator) nextCombination() bool {
	// Find rightmost element that can be incremented
	i := it.k - 1
	for i >= 0 {
		if it.indices[i] < it.n-it.k+i {
			break
		}
		i--
	}

	if i < 0 {
		return false // No more combinations
	}

	// Increment this element and reset elements to its right
	it.indices[i]++
	for j := i + 1; j < it.k; j++ {
		it.indices[j] = it.indices[j-1] + 1
	}

	return true
}

func (it *exhaustiveIterator) Checkpoint() *GeneratorCheckpoint {
	// Convert indices to position
	position := int64(0)
	for i, idx := range it.indices {
		position += int64(idx) << (i * 8) // Simple encoding
	}

	return &GeneratorCheckpoint{
		Strategy:  StrategyExhaustive,
		Position:  position,
		Generated: it.generated,
		State: map[string]any{
			"indices": it.indices,
			"done":    it.done,
		},
	}
}

func (it *exhaustiveIterator) Resume(checkpoint *GeneratorCheckpoint) error {
	if checkpoint.Strategy != StrategyExhaustive {
		return fmt.Errorf("checkpoint strategy mismatch: expected exhaustive, got %s", checkpoint.Strategy)
	}

	if indices, ok := checkpoint.State["indices"].([]int); ok {
		it.indices = indices
		it.generated = checkpoint.Generated
		if done, ok := checkpoint.State["done"].(bool); ok {
			it.done = done
		}
		return nil
	}

	return fmt.Errorf("invalid checkpoint state")
}

func (it *exhaustiveIterator) Reset() {
	for i := range it.indices {
		it.indices[i] = i
	}
	it.done = false
	it.generated = 0
}

func (it *exhaustiveIterator) Close() error {
	return nil
}

// smartSampleIterator generates decks prioritizing high-level cards and synergies
type smartSampleIterator struct {
	gen       *DeckGenerator
	remaining int
	generated int
	rng       *rand.Rand
}

func newSmartSampleIterator(gen *DeckGenerator) *smartSampleIterator {
	sampleSize := gen.config.SampleSize
	if sampleSize <= 0 {
		sampleSize = 1000
	}

	return &smartSampleIterator{
		gen:       gen,
		remaining: sampleSize,
		rng:       rand.New(rand.NewSource(gen.config.Seed)),
	}
}

func (it *smartSampleIterator) Next(ctx context.Context) ([]string, error) {
	if it.remaining <= 0 {
		return nil, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Try up to 100 times to generate a valid deck
	maxAttempts := 100
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Generate using weighted sampling based on card scores
		deck := make([]string, 0, 8)
		used := make(map[string]bool)

		// Add forced include cards first
		for card := range it.gen.includeMap {
			deck = append(deck, card)
			used[card] = true
		}

		// Fill by role with weighted selection
		roleSelections := []struct {
			role  CardRole
			count int
		}{
			{RoleWinCondition, it.gen.composition.WinConditions},
			{RoleBuilding, it.gen.composition.Buildings},
			{RoleSpellBig, it.gen.composition.BigSpells},
			{RoleSpellSmall, it.gen.composition.SmallSpells},
			{RoleSupport, it.gen.composition.Support},
			{RoleCycle, it.gen.composition.Cycle},
		}

		for _, sel := range roleSelections {
			selected := it.selectWeightedCardsFromRole(sel.role, sel.count, used)
			deck = append(deck, selected...)
			// If we couldn't get enough cards from this role, that's ok - we'll fill later
		}

		// Fill remaining slots to reach exactly 8 cards
		for len(deck) < 8 {
			added := it.gen.fillRemainingSlots(len(deck), used)
			if len(added) == 0 {
				break // No more cards available
			}
			deck = append(deck, added...)
		}

		// Validate
		lastErr = it.gen.validateDeck(deck)
		if lastErr == nil {
			it.remaining--
			it.generated++
			return deck, nil
		}
	}

	// Could not generate valid deck after max attempts
	it.remaining--
	it.generated++
	if lastErr != nil {
		return nil, fmt.Errorf("failed to generate valid deck after %d attempts, last error: %w", maxAttempts, lastErr)
	}
	return nil, fmt.Errorf("failed to generate valid deck after %d attempts", maxAttempts)
}

func (it *smartSampleIterator) selectWeightedCardsFromRole(role CardRole, count int, used map[string]bool) []string {
	candidates := it.gen.candidatesByRole[role]
	if len(candidates) == 0 {
		return nil
	}

	// Build weighted selection pool based on scores
	type weightedCard struct {
		name   string
		weight float64
	}

	pool := make([]weightedCard, 0, len(candidates))
	totalWeight := 0.0

	for _, card := range candidates {
		if used[card.Name] {
			continue
		}

		// Weight based on score with exponential emphasis on high scores
		weight := math.Pow(card.Score, 2)
		pool = append(pool, weightedCard{name: card.Name, weight: weight})
		totalWeight += weight
	}

	if len(pool) == 0 {
		return nil
	}

	selected := make([]string, 0, count)

	for i := 0; i < count && len(pool) > 0; i++ {
		// Weighted random selection
		r := it.rng.Float64() * totalWeight
		cumulative := 0.0

		for j, card := range pool {
			cumulative += card.weight
			if cumulative >= r {
				selected = append(selected, card.name)
				used[card.name] = true

				// Remove from pool and update total weight
				totalWeight -= card.weight
				pool = append(pool[:j], pool[j+1:]...)
				break
			}
		}
	}

	return selected
}

func (it *smartSampleIterator) Checkpoint() *GeneratorCheckpoint {
	return &GeneratorCheckpoint{
		Strategy:  StrategySmartSample,
		Position:  int64(it.generated),
		Generated: it.generated,
		State: map[string]any{
			"remaining": it.remaining,
		},
	}
}

func (it *smartSampleIterator) Resume(checkpoint *GeneratorCheckpoint) error {
	if checkpoint.Strategy != StrategySmartSample {
		return fmt.Errorf("checkpoint strategy mismatch")
	}

	it.generated = checkpoint.Generated
	if remaining, ok := checkpoint.State["remaining"].(int); ok {
		it.remaining = remaining
	}
	return nil
}

func (it *smartSampleIterator) Reset() {
	it.remaining = it.gen.config.SampleSize
	it.generated = 0
}

func (it *smartSampleIterator) Close() error {
	return nil
}

// randomSampleIterator generates completely random valid decks
type randomSampleIterator struct {
	gen       *DeckGenerator
	remaining int
	generated int
}

func newRandomSampleIterator(gen *DeckGenerator) *randomSampleIterator {
	sampleSize := gen.config.SampleSize
	if sampleSize <= 0 {
		sampleSize = 1000
	}

	return &randomSampleIterator{
		gen:       gen,
		remaining: sampleSize,
	}
}

func (it *randomSampleIterator) Next(ctx context.Context) ([]string, error) {
	if it.remaining <= 0 {
		return nil, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Try up to 100 times to generate a valid deck
	maxAttempts := 100
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Generate random deck
		deck := make([]string, 0, 8)
		used := make(map[string]bool)

		// Add forced include cards
		for card := range it.gen.includeMap {
			deck = append(deck, card)
			used[card] = true
		}

		// Fill by role with random selection
		roleSelections := []struct {
			role  CardRole
			count int
		}{
			{RoleWinCondition, it.gen.composition.WinConditions},
			{RoleBuilding, it.gen.composition.Buildings},
			{RoleSpellBig, it.gen.composition.BigSpells},
			{RoleSpellSmall, it.gen.composition.SmallSpells},
			{RoleSupport, it.gen.composition.Support},
			{RoleCycle, it.gen.composition.Cycle},
		}

		for _, sel := range roleSelections {
			selected := it.gen.selectRandomCardsFromRole(sel.role, sel.count, used)
			deck = append(deck, selected...)
			// If we couldn't get enough cards from this role, that's ok - we'll fill later
		}

		// Fill remaining slots to reach exactly 8 cards
		for len(deck) < 8 {
			added := it.gen.fillRemainingSlots(len(deck), used)
			if len(added) == 0 {
				break // No more cards available
			}
			deck = append(deck, added...)
		}

		// Validate
		lastErr = it.gen.validateDeck(deck)
		if lastErr == nil {
			it.remaining--
			it.generated++
			return deck, nil
		}
	}

	// Could not generate valid deck after max attempts
	it.remaining--
	it.generated++
	if lastErr != nil {
		return nil, fmt.Errorf("failed to generate valid deck after %d attempts, last error: %w", maxAttempts, lastErr)
	}
	return nil, fmt.Errorf("failed to generate valid deck after %d attempts", maxAttempts)
}

func (it *randomSampleIterator) Checkpoint() *GeneratorCheckpoint {
	return &GeneratorCheckpoint{
		Strategy:  StrategyRandomSample,
		Position:  int64(it.generated),
		Generated: it.generated,
		State: map[string]any{
			"remaining": it.remaining,
		},
	}
}

func (it *randomSampleIterator) Resume(checkpoint *GeneratorCheckpoint) error {
	if checkpoint.Strategy != StrategyRandomSample {
		return fmt.Errorf("checkpoint strategy mismatch")
	}

	it.generated = checkpoint.Generated
	if remaining, ok := checkpoint.State["remaining"].(int); ok {
		it.remaining = remaining
	}
	return nil
}

func (it *randomSampleIterator) Reset() {
	it.remaining = it.gen.config.SampleSize
	it.generated = 0
}

func (it *randomSampleIterator) Close() error {
	return nil
}

// archetypeIterator generates decks focused on a specific archetype
type archetypeIterator struct {
	gen       *DeckGenerator
	remaining int
	generated int
	archetype string
}

func newArchetypeIterator(gen *DeckGenerator) *archetypeIterator {
	sampleSize := gen.config.SampleSize
	if sampleSize <= 0 {
		sampleSize = 1000
	}

	return &archetypeIterator{
		gen:       gen,
		remaining: sampleSize,
		archetype: gen.config.Archetype,
	}
}

func (it *archetypeIterator) Next(ctx context.Context) ([]string, error) {
	if it.remaining <= 0 {
		return nil, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	it.remaining--
	it.generated++

	// TODO: Implement archetype-specific logic
	// For now, delegate to smart sampling
	// Future: integrate with archetypes package for archetype-aware generation

	smartIter := newSmartSampleIterator(it.gen)
	return smartIter.Next(ctx)
}

func (it *archetypeIterator) Checkpoint() *GeneratorCheckpoint {
	return &GeneratorCheckpoint{
		Strategy:  StrategyArchetypeFocused,
		Position:  int64(it.generated),
		Generated: it.generated,
		State: map[string]any{
			"remaining": it.remaining,
			"archetype": it.archetype,
		},
	}
}

func (it *archetypeIterator) Resume(checkpoint *GeneratorCheckpoint) error {
	if checkpoint.Strategy != StrategyArchetypeFocused {
		return fmt.Errorf("checkpoint strategy mismatch")
	}

	it.generated = checkpoint.Generated
	if remaining, ok := checkpoint.State["remaining"].(int); ok {
		it.remaining = remaining
	}
	if archetype, ok := checkpoint.State["archetype"].(string); ok {
		it.archetype = archetype
	}
	return nil
}

func (it *archetypeIterator) Reset() {
	it.remaining = it.gen.config.SampleSize
	it.generated = 0
}

func (it *archetypeIterator) Close() error {
	return nil
}
