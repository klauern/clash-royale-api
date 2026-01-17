// Package deck provides Monte Carlo-style deck fuzzing functionality
// for generating random valid deck combinations from a player's card collection.
package deck

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// FuzzingConfig contains configuration parameters for deck fuzzing
type FuzzingConfig struct {
	// Count is the number of decks to generate (default: 1000)
	Count int
	// Workers is the number of parallel workers (default: 1)
	Workers int
	// Seed is the random seed for reproducibility (0 = random)
	Seed int64
	// IncludeCards are cards that must be included in every generated deck
	IncludeCards []string
	// ExcludeCards are cards that must be excluded from all generated decks
	ExcludeCards []string
	// MinAvgElixir is the minimum average elixir (default: 0.0)
	MinAvgElixir float64
	// MaxAvgElixir is the maximum average elixir (default: 10.0)
	MaxAvgElixir float64
	// MinOverallScore is the minimum overall score threshold
	MinOverallScore float64
	// MinSynergyScore is the minimum synergy score threshold
	MinSynergyScore float64
}

// FuzzingStats tracks metrics during deck generation
type FuzzingStats struct {
	mu               sync.Mutex
	Generated        int
	Success          int
	Failed           int
	SkippedElixir    int
	SkippedInclude   int
	SkippedExclude   int
	SkippedScore     int
	StartTime        time.Time
	GenerationTimes  []time.Duration
}

// FuzzedDeck represents a generated deck with its evaluation results
type FuzzedDeck struct {
	Deck           []string
	AvgElixir      float64
	OverallScore   float64
	AttackScore    float64
	DefenseScore   float64
	SynergyScore   float64
	VersatilityScore float64
	Archetype      string
	GenerationTime time.Duration
}

// RoleComposition defines the required card count for each role
type RoleComposition struct {
	WinConditions int
	Buildings     int
	BigSpells     int
	SmallSpells   int
	Support       int
	Cycle         int
}

// DefaultRoleComposition returns the standard deck composition
func DefaultRoleComposition() *RoleComposition {
	return &RoleComposition{
		WinConditions: 1,
		Buildings:     1,
		BigSpells:     1,
		SmallSpells:   1,
		Support:       2,
		Cycle:         2,
	}
}

// DeckFuzzer handles the generation of random valid deck combinations
type DeckFuzzer struct {
	cardsByRole map[config.CardRole][]CardCandidate
	allCards    []CardCandidate
	config      *FuzzingConfig
	composition *RoleComposition
	rng         *rand.Rand
	stats       *FuzzingStats
	excludeMap  map[string]bool
	includeMap  map[string]bool
}

// NewDeckFuzzer creates a new deck fuzzer from a player's card collection
func NewDeckFuzzer(player *clashroyale.Player, cfg *FuzzingConfig) (*DeckFuzzer, error) {
	if player == nil {
		return nil, fmt.Errorf("player cannot be nil")
	}
	if len(player.Cards) < 8 {
		return nil, fmt.Errorf("player must have at least 8 cards, got %d", len(player.Cards))
	}

	// Set default config values
	if cfg == nil {
		cfg = &FuzzingConfig{}
	}
	if cfg.Count <= 0 {
		cfg.Count = 1000
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}
	if cfg.MinAvgElixir < 0 {
		cfg.MinAvgElixir = 0
	}
	if cfg.MaxAvgElixir <= 0 || cfg.MaxAvgElixir > 10 {
		cfg.MaxAvgElixir = 10
	}
	if cfg.MinOverallScore < 0 {
		cfg.MinOverallScore = 0
	}
	if cfg.MinOverallScore > 10 {
		cfg.MinOverallScore = 10
	}
	if cfg.MinSynergyScore < 0 {
		cfg.MinSynergyScore = 0
	}
	if cfg.MinSynergyScore > 10 {
		cfg.MinSynergyScore = 10
	}

	// Initialize random number generator
	var rng *rand.Rand
	if cfg.Seed != 0 {
		rng = rand.New(rand.NewSource(cfg.Seed))
	} else {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	// Build exclude and include maps
	excludeMap := make(map[string]bool)
	for _, card := range cfg.ExcludeCards {
		excludeMap[strings.TrimSpace(card)] = true
	}

	includeMap := make(map[string]bool)
	for _, card := range cfg.IncludeCards {
		includeMap[strings.TrimSpace(card)] = true
	}

	// Convert player cards to candidates and categorize by role
	cardsByRole := make(map[config.CardRole][]CardCandidate)
	allCards := make([]CardCandidate, 0, len(player.Cards))

	for _, card := range player.Cards {
		cardName := strings.TrimSpace(card.Name)

		// Skip excluded cards
		if excludeMap[cardName] {
			continue
		}

		role := config.GetCardRoleWithEvolution(cardName, card.EvolutionLevel)

		// Calculate level ratio manually
		levelRatio := float64(card.Level) / float64(card.MaxLevel)
		if card.MaxLevel == 0 {
			levelRatio = 0
		}

		candidate := CardCandidate{
			Name:              cardName,
			Level:             card.Level,
			MaxLevel:          card.MaxLevel,
			Rarity:            card.Rarity,
			Elixir:            config.GetCardElixir(cardName, card.ElixirCost),
			Role:              &role,
			Score:             levelRatio * 10, // Simple score based on level
			EvolutionLevel:    card.EvolutionLevel,
			MaxEvolutionLevel: card.MaxEvolutionLevel,
		}

		allCards = append(allCards, candidate)

		// Categorize by role (only if role is defined)
		if role != "" {
			cardsByRole[role] = append(cardsByRole[role], candidate)
		}
	}

	fuzzer := &DeckFuzzer{
		cardsByRole: cardsByRole,
		allCards:    allCards,
		config:      cfg,
		composition: DefaultRoleComposition(),
		rng:         rng,
		stats: &FuzzingStats{
			StartTime:       time.Now(),
			GenerationTimes: make([]time.Duration, 0, cfg.Count),
		},
		excludeMap: excludeMap,
		includeMap: includeMap,
	}

	return fuzzer, nil
}

// GenerateRandomDeck generates a single random valid deck using smart sampling
func (df *DeckFuzzer) GenerateRandomDeck() ([]string, error) {
	const maxRetries = 100

	for attempt := 0; attempt < maxRetries; attempt++ {
		deck, err := df.generateRandomDeckAttempt()
		if err != nil {
			df.recordFailure()
			continue
		}

		df.recordSuccess()
		return deck, nil
	}

	df.recordFailure()
	return nil, fmt.Errorf("failed to generate valid deck after %d attempts", maxRetries)
}

// generateRandomDeckAttempt attempts to generate a single random valid deck
func (df *DeckFuzzer) generateRandomDeckAttempt() ([]string, error) {
	deck := make([]string, 0, 8)
	used := make(map[string]bool)

	// 1. Add include cards first (force-add any --include-cards)
	for cardName := range df.includeMap {
		if !df.isCardAvailable(cardName) {
			return nil, fmt.Errorf("included card not available: %s", cardName)
		}
		deck = append(deck, cardName)
		used[cardName] = true
	}

	// 2. Select cards by role using weighted random sampling
	roleSelections := []struct {
		role  config.CardRole
		count int
	}{
		{config.RoleWinCondition, df.composition.WinConditions},
		{config.RoleBuilding, df.composition.Buildings},
		{config.RoleSpellBig, df.composition.BigSpells},
		{config.RoleSpellSmall, df.composition.SmallSpells},
		{config.RoleSupport, df.composition.Support},
		{config.RoleCycle, df.composition.Cycle},
	}

	for _, selection := range roleSelections {
		cards := df.selectRandomCards(selection.role, selection.count, used)
		if len(cards) < selection.count {
			// Not enough cards of this role, fill with remaining available cards
			remaining := df.fillRemainingSlots(8-len(deck), used)
			cards = append(cards, remaining...)
		}
		for _, card := range cards {
			if !used[card] {
				deck = append(deck, card)
				used[card] = true
			}
		}
	}

	// 3. Fill remaining slots with highest-score available cards
	for len(deck) < 8 {
		remaining := df.getHighestScoreAvailableCards(used, 8-len(deck))
		if len(remaining) == 0 {
			break
		}
		for _, card := range remaining {
			if !used[card] && len(deck) < 8 {
				deck = append(deck, card)
				used[card] = true
			}
		}
	}

	// 4. Validate deck
	if len(deck) != 8 {
		return nil, fmt.Errorf("invalid deck size: %d", len(deck))
	}

	// 5. Validate average elixir
	avgElixir := df.calculateAvgElixir(deck)
	if avgElixir < df.config.MinAvgElixir || avgElixir > df.config.MaxAvgElixir {
		df.stats.SkippedElixir++
		return nil, fmt.Errorf("elixir out of range: %.2f", avgElixir)
	}

	// 6. Validate all include cards are present
	for cardName := range df.includeMap {
		if !used[cardName] {
			df.stats.SkippedInclude++
			return nil, fmt.Errorf("missing include card: %s", cardName)
		}
	}

	// 7. Validate no excluded cards are present
	for _, cardName := range deck {
		if df.excludeMap[cardName] {
			df.stats.SkippedExclude++
			return nil, fmt.Errorf("excluded card present: %s", cardName)
		}
	}

	return deck, nil
}

// selectRandomCards selects random cards from a role using weighted sampling
func (df *DeckFuzzer) selectRandomCards(role config.CardRole, count int, used map[string]bool) []string {
	cards := df.cardsByRole[role]
	if len(cards) == 0 {
		return nil
	}

	// Filter out used cards
	available := make([]CardCandidate, 0, len(cards))
	for _, card := range cards {
		if !used[card.Name] {
			available = append(available, card)
		}
	}

	if len(available) == 0 {
		return nil
	}

	// Weighted random selection based on card scores
	selected := make([]string, 0, count)
	for i := 0; i < count && len(available) > 0; i++ {
		// Calculate total score for weighted selection
		totalScore := 0.0
		for _, card := range available {
			totalScore += card.Score
		}

		// Select random card weighted by score
		r := df.rng.Float64() * totalScore
		cumScore := 0.0
		selectedIdx := -1
		for idx, card := range available {
			cumScore += card.Score
			if r <= cumScore {
				selectedIdx = idx
				break
			}
		}

		// Fallback to random if weighted selection failed
		if selectedIdx == -1 {
			selectedIdx = df.rng.Intn(len(available))
		}

		selected = append(selected, available[selectedIdx].Name)
		used[available[selectedIdx].Name] = true

		// Remove selected card from available
		available = append(available[:selectedIdx], available[selectedIdx+1:]...)
	}

	return selected
}

// fillRemainingSlots fills remaining slots with random available cards
func (df *DeckFuzzer) fillRemainingSlots(count int, used map[string]bool) []string {
	selected := make([]string, 0, count)

	// Get all available cards
	available := make([]CardCandidate, 0)
	for _, card := range df.allCards {
		if !used[card.Name] {
			available = append(available, card)
		}
	}

	// Shuffle and select
	df.rng.Shuffle(len(available), func(i, j int) {
		available[i], available[j] = available[j], available[i]
	})

	for i := 0; i < count && i < len(available); i++ {
		selected = append(selected, available[i].Name)
	}

	return selected
}

// getHighestScoreAvailableCards returns the highest scoring available cards
func (df *DeckFuzzer) getHighestScoreAvailableCards(used map[string]bool, count int) []string {
	available := make([]CardCandidate, 0)
	for _, card := range df.allCards {
		if !used[card.Name] {
			available = append(available, card)
		}
	}

	// Sort by score descending
	for i := 0; i < len(available); i++ {
		for j := i + 1; j < len(available); j++ {
			if available[j].Score > available[i].Score {
				available[i], available[j] = available[j], available[i]
			}
		}
	}

	result := make([]string, 0, count)
	for i := 0; i < count && i < len(available); i++ {
		result = append(result, available[i].Name)
	}

	return result
}

// calculateAvgElixir calculates the average elixir cost of a deck
func (df *DeckFuzzer) calculateAvgElixir(deck []string) float64 {
	if len(deck) == 0 {
		return 0
	}

	total := 0
	for _, cardName := range deck {
		for _, card := range df.allCards {
			if card.Name == cardName {
				total += card.Elixir
				break
			}
		}
	}

	return float64(total) / float64(len(deck))
}

// isCardAvailable checks if a card is in the available card pool
func (df *DeckFuzzer) isCardAvailable(cardName string) bool {
	for _, card := range df.allCards {
		if card.Name == cardName {
			return true
		}
	}
	return false
}

// recordSuccess records a successful deck generation
func (df *DeckFuzzer) recordSuccess() {
	df.stats.mu.Lock()
	df.stats.Generated++
	df.stats.Success++
	df.stats.mu.Unlock()
}

// recordFailure records a failed deck generation attempt
func (df *DeckFuzzer) recordFailure() {
	df.stats.mu.Lock()
	df.stats.Generated++
	df.stats.Failed++
	df.stats.mu.Unlock()
}

// GetStats returns a copy of the current stats
func (df *DeckFuzzer) GetStats() FuzzingStats {
	df.stats.mu.Lock()
	defer df.stats.mu.Unlock()

	return FuzzingStats{
		Generated:       df.stats.Generated,
		Success:         df.stats.Success,
		Failed:          df.stats.Failed,
		SkippedElixir:   df.stats.SkippedElixir,
		SkippedInclude:  df.stats.SkippedInclude,
		SkippedExclude:  df.stats.SkippedExclude,
		SkippedScore:    df.stats.SkippedScore,
		StartTime:       df.stats.StartTime,
		GenerationTimes: append([]time.Duration{}, df.stats.GenerationTimes...),
	}
}

// GenerateDecks generates the specified number of decks
func (df *DeckFuzzer) GenerateDecks(count int) ([][]string, error) {
	decks := make([][]string, 0, count)

	for i := 0; i < count; i++ {
		deck, err := df.GenerateRandomDeck()
		if err != nil {
			// Continue on error, just skip this deck
			continue
		}
		decks = append(decks, deck)
	}

	return decks, nil
}

// GenerateDecksParallel generates decks using parallel workers
func (df *DeckFuzzer) GenerateDecksParallel() ([][]string, error) {
	if df.config.Workers <= 1 {
		return df.GenerateDecks(df.config.Count)
	}

	decks := make([][]string, 0, df.config.Count)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create a worker pool
	workChan := make(chan int, df.config.Count)
	resultChan := make(chan []string, df.config.Count)

	// Start workers
	for w := 0; w < df.config.Workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range workChan {
				deck, err := df.GenerateRandomDeck()
				if err == nil {
					resultChan <- deck
				}
			}
		}()
	}

	// Send work
	go func() {
		for i := 0; i < df.config.Count; i++ {
			workChan <- i
		}
		close(workChan)
	}()

	// Close result channel when workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for deck := range resultChan {
		mu.Lock()
		decks = append(decks, deck)
		mu.Unlock()
	}

	return decks, nil
}

// SetRoleComposition sets a custom role composition
func (df *DeckFuzzer) SetRoleComposition(comp *RoleComposition) {
	if comp != nil {
		df.composition = comp
	}
}

// GetCardsByRole returns the available cards for a specific role
func (df *DeckFuzzer) GetCardsByRole(role config.CardRole) []CardCandidate {
	return df.cardsByRole[role]
}

// GetAllCards returns all available cards
func (df *DeckFuzzer) GetAllCards() []CardCandidate {
	return df.allCards
}

