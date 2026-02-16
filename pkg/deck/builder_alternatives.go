// Package deck provides alternative deck building methodologies that don't rely on
// static archetype templates. These builders use synergy, role, and counter data
// to construct viable decks from scratch.
//
// Available builders:
//   - SynergyGraphBuilder: Maximizes synergy weight across card pairs
//   - ConstraintSatisfactionBuilder: Enforces hard/soft constraints
//   - RoleFirstBuilder: Fills roles in priority order
//   - CounterCentricBuilder: Maximizes threat coverage
//   - ArchetypeFreeGABuilder: Genetic algorithm without archetype fitness
//
// Design document: docs/DECK_BUILDER_ALTERNATIVES.md
package deck

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
)

// DeckBuilder defines the common interface for all deck building strategies.
type DeckBuilder interface {
	// Build constructs an 8-card deck from the available candidates.
	// Returns the deck as card names and an error if building fails.
	Build() ([]string, error)

	// Name returns the builder's identifier for logging and selection.
	Name() string
}

// BuilderConfig provides common configuration for all builders.
type BuilderConfig struct {
	// Candidates is the pool of available cards
	Candidates []*CardCandidate

	// SynergyDB provides synergy pair data
	SynergyDB *SynergyDatabase

	// CounterMatrix provides threat/counter relationships
	CounterMatrix *CounterMatrix

	// PreferredElixir target (0 to ignore)
	PreferredElixir float64

	// MaxElixir upper bound (0 to ignore)
	MaxElixir float64

	// RequireWinCondition ensures at least one win condition
	RequireWinCondition bool

	// RequireSpell ensures at least one spell
	RequireSpell bool

	// RequireAirDefense ensures at least one air-targeting card
	RequireAirDefense bool
}

// DefaultBuilderConfig returns sensible defaults for deck building.
func DefaultBuilderConfig(candidates []*CardCandidate, synergyDB *SynergyDatabase) BuilderConfig {
	return BuilderConfig{
		Candidates:          candidates,
		SynergyDB:           synergyDB,
		CounterMatrix:       NewCounterMatrixWithDefaults(),
		PreferredElixir:     3.5,
		MaxElixir:           4.5,
		RequireWinCondition: true,
		RequireSpell:        true,
		RequireAirDefense:   true,
	}
}

// =============================================================================
// 1. Synergy Graph Builder
// =============================================================================

// SynergyGraphBuilder constructs decks by modeling cards as graph nodes
// with synergy pairs as weighted edges. It finds the 8-card subgraph
// with maximum total synergy weight.
type SynergyGraphBuilder struct {
	config BuilderConfig
}

// NewSynergyGraphBuilder creates a new synergy graph builder.
func NewSynergyGraphBuilder(config BuilderConfig) *SynergyGraphBuilder {
	return &SynergyGraphBuilder{config: config}
}

func (b *SynergyGraphBuilder) Name() string {
	return "synergy_graph"
}

// Build constructs a deck by maximizing synergy connections.
// Uses a greedy approximation since finding optimal 8-node subgraph is NP-hard.
//
//nolint:funlen,gocognit,gocyclo // Greedy selection flow intentionally keeps scoring/constraint checks in one pass.
func (b *SynergyGraphBuilder) Build() ([]string, error) {
	if len(b.config.Candidates) < 8 {
		return nil, fmt.Errorf("insufficient candidates: need at least 8, got %d", len(b.config.Candidates))
	}

	if b.config.SynergyDB == nil {
		return nil, fmt.Errorf("synergy database is required")
	}

	// Build synergy graph: card -> {neighbor -> weight}
	synergyGraph := b.buildSynergyGraph()

	// Start with the card that has the highest total synergy potential
	deck := make([]string, 0, 8)
	inDeck := make(map[string]bool)

	// Find best starting card
	startCard := b.findBestStartCard(synergyGraph)
	if startCard == "" {
		// Fallback to first valid candidate
		for _, c := range b.config.Candidates {
			if b.meetsConstraints([]string{c.Name}) {
				startCard = c.Name
				break
			}
		}
	}

	if startCard == "" {
		return nil, fmt.Errorf("no valid starting card found")
	}

	deck = append(deck, startCard)
	inDeck[startCard] = true

	// Greedily add cards that maximize synergy with current deck
	for len(deck) < 8 {
		bestCard := ""
		bestScore := -1.0

		for _, candidate := range b.config.Candidates {
			if inDeck[candidate.Name] {
				continue
			}

			// Calculate synergy with current deck
			synergyScore := b.calculateSynergyWithDeck(candidate.Name, deck, synergyGraph)

			// Check constraints
			testDeck := append(append([]string{}, deck...), candidate.Name)
			if len(testDeck) == 8 && !b.meetsConstraints(testDeck) {
				continue
			}

			// Bonus for meeting unfulfilled constraints early
			constraintBonus := b.constraintBonus(candidate, deck)

			totalScore := synergyScore + constraintBonus
			if totalScore > bestScore {
				bestScore = totalScore
				bestCard = candidate.Name
			}
		}

		if bestCard == "" {
			// Can't find a card - try to meet constraints
			bestCard = b.findConstraintFiller(deck, inDeck)
			if bestCard == "" {
				return nil, fmt.Errorf("failed to build deck: stuck at %d cards", len(deck))
			}
		}

		deck = append(deck, bestCard)
		inDeck[bestCard] = true
	}

	return deck, nil
}

// buildSynergyGraph creates adjacency map from synergy database.
func (b *SynergyGraphBuilder) buildSynergyGraph() map[string]map[string]float64 {
	graph := make(map[string]map[string]float64)

	for _, pair := range b.config.SynergyDB.Pairs {
		if graph[pair.Card1] == nil {
			graph[pair.Card1] = make(map[string]float64)
		}
		if graph[pair.Card2] == nil {
			graph[pair.Card2] = make(map[string]float64)
		}
		graph[pair.Card1][pair.Card2] = pair.Score
		graph[pair.Card2][pair.Card1] = pair.Score
	}

	return graph
}

// findBestStartCard finds the card with highest total synergy potential.
func (b *SynergyGraphBuilder) findBestStartCard(graph map[string]map[string]float64) string {
	bestCard := ""
	bestTotal := -1.0

	candidateNames := make(map[string]bool)
	for _, c := range b.config.Candidates {
		candidateNames[c.Name] = true
	}

	for _, c := range b.config.Candidates {
		total := 0.0
		for neighbor, weight := range graph[c.Name] {
			// Only count synergies with available candidates
			if candidateNames[neighbor] {
				total += weight
			}
		}

		// Prefer win conditions as starting cards
		if c.Role != nil && *c.Role == RoleWinCondition {
			total += 0.5
		}

		if total > bestTotal {
			bestTotal = total
			bestCard = c.Name
		}
	}

	return bestCard
}

// calculateSynergyWithDeck sums synergy weights with all current deck cards.
func (b *SynergyGraphBuilder) calculateSynergyWithDeck(cardName string, deck []string, graph map[string]map[string]float64) float64 {
	total := 0.0
	for _, deckCard := range deck {
		if weight, exists := graph[cardName][deckCard]; exists {
			total += weight
		}
	}
	return total
}

// constraintBonus rewards cards that help fulfill constraints.
//
//nolint:gocognit,gocyclo // Constraint checks are explicit for readability over decomposition.
func (b *SynergyGraphBuilder) constraintBonus(candidate *CardCandidate, deck []string) float64 {
	bonus := 0.0

	// Check if deck needs a win condition
	if b.config.RequireWinCondition {
		hasWinCon := false
		for _, name := range deck {
			if c := b.findCandidate(name); c != nil && c.Role != nil && *c.Role == RoleWinCondition {
				hasWinCon = true
				break
			}
		}
		if !hasWinCon && candidate.Role != nil && *candidate.Role == RoleWinCondition {
			bonus += 0.3
		}
	}

	// Check if deck needs a spell
	if b.config.RequireSpell {
		hasSpell := false
		for _, name := range deck {
			if c := b.findCandidate(name); c != nil && c.Role != nil && (*c.Role == RoleSpellBig || *c.Role == RoleSpellSmall) {
				hasSpell = true
				break
			}
		}
		if !hasSpell && candidate.Role != nil && (*candidate.Role == RoleSpellBig || *candidate.Role == RoleSpellSmall) {
			bonus += 0.2
		}
	}

	// Check air defense
	if b.config.RequireAirDefense {
		hasAir := false
		for _, name := range deck {
			if c := b.findCandidate(name); c != nil && c.Stats != nil && (c.Stats.Targets == targetsAir || c.Stats.Targets == targetsAirAndGround) {
				hasAir = true
				break
			}
		}
		if !hasAir && candidate.Stats != nil && (candidate.Stats.Targets == targetsAir || candidate.Stats.Targets == targetsAirAndGround) {
			bonus += 0.2
		}
	}

	return bonus
}

// meetsConstraints checks if a complete deck satisfies all hard constraints.
//
//nolint:gocyclo // Constraint gating is intentionally branch-heavy.
func (b *SynergyGraphBuilder) meetsConstraints(deck []string) bool {
	hasWinCon := !b.config.RequireWinCondition
	hasSpell := !b.config.RequireSpell
	hasAirDef := !b.config.RequireAirDefense
	totalElixir := 0

	for _, name := range deck {
		c := b.findCandidate(name)
		if c == nil {
			continue
		}

		totalElixir += c.Elixir

		if c.Role != nil && *c.Role == RoleWinCondition {
			hasWinCon = true
		}
		if c.Role != nil && (*c.Role == RoleSpellBig || *c.Role == RoleSpellSmall) {
			hasSpell = true
		}
		if c.Stats != nil && (c.Stats.Targets == targetsAir || c.Stats.Targets == targetsAirAndGround) {
			hasAirDef = true
		}
	}

	// Check elixir constraint
	if b.config.MaxElixir > 0 {
		avgElixir := float64(totalElixir) / float64(len(deck))
		if avgElixir > b.config.MaxElixir {
			return false
		}
	}

	return hasWinCon && hasSpell && hasAirDef
}

// findCandidate looks up a candidate by name.
func (b *SynergyGraphBuilder) findCandidate(name string) *CardCandidate {
	for _, c := range b.config.Candidates {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// findConstraintFiller finds a card that helps meet unmet constraints.
func (b *SynergyGraphBuilder) findConstraintFiller(deck []string, inDeck map[string]bool) string {
	for _, c := range b.config.Candidates {
		if inDeck[c.Name] {
			continue
		}
		testDeck := append(append([]string{}, deck...), c.Name)
		if len(testDeck) < 8 || b.meetsConstraints(testDeck) {
			return c.Name
		}
	}
	return ""
}

// =============================================================================
// 2. Constraint Satisfaction Builder
// =============================================================================

// ConstraintSatisfactionBuilder uses hard and soft constraints to build decks.
// Hard constraints must be satisfied (win con, spells, air defense).
// Soft constraints are optimized (synergy, elixir curve, counter coverage).
type ConstraintSatisfactionBuilder struct {
	config BuilderConfig
}

// NewConstraintSatisfactionBuilder creates a new CSP-style builder.
func NewConstraintSatisfactionBuilder(config BuilderConfig) *ConstraintSatisfactionBuilder {
	return &ConstraintSatisfactionBuilder{config: config}
}

func (b *ConstraintSatisfactionBuilder) Name() string {
	return "constraint_satisfaction"
}

// SlotConstraint defines requirements for a deck slot.
type SlotConstraint struct {
	Name         string
	Required     bool
	AllowedRoles []CardRole
	Filter       func(*CardCandidate) bool
}

// Build constructs a deck by filling slots according to constraints.
//
//nolint:gocyclo // Slot-constraint orchestration is inherently branchy.
func (b *ConstraintSatisfactionBuilder) Build() ([]string, error) {
	if len(b.config.Candidates) < 8 {
		return nil, fmt.Errorf("insufficient candidates: need at least 8, got %d", len(b.config.Candidates))
	}

	// Define slot constraints
	slots := []SlotConstraint{
		{Name: "WinCondition", Required: true, AllowedRoles: []CardRole{RoleWinCondition}},
		{Name: "BigSpell", Required: true, AllowedRoles: []CardRole{RoleSpellBig}},
		{Name: "SmallSpell", Required: true, AllowedRoles: []CardRole{RoleSpellSmall}},
		{Name: "AirDefense", Required: true, Filter: func(c *CardCandidate) bool {
			return c.Stats != nil && (c.Stats.Targets == targetsAir || c.Stats.Targets == targetsAirAndGround)
		}},
		{Name: "Support1", Required: false, AllowedRoles: []CardRole{RoleSupport}},
		{Name: "Support2", Required: false, AllowedRoles: []CardRole{RoleSupport}},
		{Name: "Flex1", Required: false},
		{Name: "Flex2", Required: false},
	}

	deck := make([]string, 0, 8)
	used := make(map[string]bool)

	// Fill required slots first
	for _, slot := range slots {
		if !slot.Required {
			continue
		}

		best := b.findBestForSlot(slot, deck, used)
		if best == "" {
			return nil, fmt.Errorf("cannot satisfy slot: %s", slot.Name)
		}
		deck = append(deck, best)
		used[best] = true
	}

	// Fill optional slots
	for _, slot := range slots {
		if slot.Required || len(deck) >= 8 {
			continue
		}

		best := b.findBestForSlot(slot, deck, used)
		if best != "" {
			deck = append(deck, best)
			used[best] = true
		}
	}

	// Fill remaining slots with highest-scoring cards
	for len(deck) < 8 {
		best := b.findBestRemaining(deck, used)
		if best == "" {
			return nil, fmt.Errorf("cannot fill deck: stuck at %d cards", len(deck))
		}
		deck = append(deck, best)
		used[best] = true
	}

	return deck, nil
}

// findBestForSlot finds the best card for a slot constraint.
//
//nolint:gocognit,gocyclo // Slot matching combines role and predicate checks in one scan for performance.
func (b *ConstraintSatisfactionBuilder) findBestForSlot(slot SlotConstraint, deck []string, used map[string]bool) string {
	type scoredCard struct {
		name  string
		score float64
	}

	var candidates []scoredCard

	for _, c := range b.config.Candidates {
		if used[c.Name] {
			continue
		}

		// Check role constraint
		if len(slot.AllowedRoles) > 0 {
			if c.Role == nil {
				continue
			}
			roleMatch := false
			for _, r := range slot.AllowedRoles {
				if *c.Role == r {
					roleMatch = true
					break
				}
			}
			if !roleMatch {
				continue
			}
		}

		// Check filter constraint
		if slot.Filter != nil && !slot.Filter(c) {
			continue
		}

		// Score card based on soft constraints
		score := b.scoreCard(c, deck)
		candidates = append(candidates, scoredCard{name: c.Name, score: score})
	}

	if len(candidates) == 0 {
		return ""
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	return candidates[0].name
}

// findBestRemaining finds the best card not yet in deck.
func (b *ConstraintSatisfactionBuilder) findBestRemaining(deck []string, used map[string]bool) string {
	type scoredCard struct {
		name  string
		score float64
	}

	var candidates []scoredCard

	for _, c := range b.config.Candidates {
		if used[c.Name] {
			continue
		}
		score := b.scoreCard(c, deck)
		candidates = append(candidates, scoredCard{name: c.Name, score: score})
	}

	if len(candidates) == 0 {
		return ""
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	return candidates[0].name
}

// scoreCard evaluates a card based on soft constraints.
func (b *ConstraintSatisfactionBuilder) scoreCard(c *CardCandidate, deck []string) float64 {
	score := 0.0

	// Level/quality component (30%)
	score += c.LevelRatio() * 0.3

	// Synergy with current deck (40%)
	if b.config.SynergyDB != nil {
		synergyScore := 0.0
		for _, deckCard := range deck {
			synergyScore += b.config.SynergyDB.GetSynergy(c.Name, deckCard)
		}
		score += synergyScore * 0.4
	}

	// Elixir fit (15%)
	if b.config.PreferredElixir > 0 {
		elixirDiff := math.Abs(float64(c.Elixir) - b.config.PreferredElixir)
		elixirScore := 1.0 - (elixirDiff / 5.0)
		if elixirScore < 0 {
			elixirScore = 0
		}
		score += elixirScore * 0.15
	}

	// Counter coverage bonus (15%)
	if b.config.CounterMatrix != nil {
		capabilities := b.config.CounterMatrix.GetCardCapabilities(c.Name)
		score += float64(len(capabilities)) * 0.05
	}

	return score
}

// =============================================================================
// 3. Archetype-Free GA Fitness
// =============================================================================

// ArchetypeFreeScorer provides a fitness function for genetic algorithms
// that does not rely on archetype coherence.
type ArchetypeFreeScorer struct {
	SynergyDB     *SynergyDatabase
	CounterMatrix *CounterMatrix
}

// NewArchetypeFreeScorer creates a scorer without archetype dependencies.
func NewArchetypeFreeScorer(synergyDB *SynergyDatabase, counterMatrix *CounterMatrix) *ArchetypeFreeScorer {
	return &ArchetypeFreeScorer{
		SynergyDB:     synergyDB,
		CounterMatrix: counterMatrix,
	}
}

// Score calculates fitness without archetype component.
// Formula: 0.35×Synergy + 0.25×Coverage + 0.25×CardQuality + 0.15×ElixirFit
func (s *ArchetypeFreeScorer) Score(cards []CardCandidate) float64 {
	if len(cards) == 0 {
		return 0
	}

	// Synergy component (35%)
	synergyScore := s.calculateSynergy(cards)

	// Counter coverage component (25%)
	coverageScore := s.calculateCoverage(cards)

	// Card quality component (25%)
	qualityScore := s.calculateQuality(cards)

	// Elixir fit component (15%)
	elixirScore := s.calculateElixirFit(cards)

	return (synergyScore * 0.35) + (coverageScore * 0.25) + (qualityScore * 0.25) + (elixirScore * 0.15)
}

// calculateSynergy computes synergy score from pair database.
func (s *ArchetypeFreeScorer) calculateSynergy(cards []CardCandidate) float64 {
	if s.SynergyDB == nil || len(cards) < 2 {
		return 0.5
	}

	totalSynergy := 0.0
	pairCount := 0

	for i := range cards {
		for j := i + 1; j < len(cards); j++ {
			synergy := s.SynergyDB.GetSynergy(cards[i].Name, cards[j].Name)
			if synergy > 0 {
				totalSynergy += synergy
				pairCount++
			}
		}
	}

	// Scale: 0 pairs = 0.3, 5+ high-quality pairs = 1.0
	if pairCount == 0 {
		return 0.3
	}

	avgSynergy := totalSynergy / float64(pairCount)
	// Coverage bonus: having many synergy pairs is good
	coverageBonus := math.Min(float64(pairCount)/10.0, 0.3)

	return math.Min(avgSynergy+coverageBonus, 1.0)
}

// calculateCoverage evaluates defensive coverage without archetype bias.
//
//nolint:gocyclo // Coverage rubric requires explicit category branching.
func (s *ArchetypeFreeScorer) calculateCoverage(cards []CardCandidate) float64 {
	// Check for key capabilities
	hasWinCon := false
	hasSpell := false
	airDefCount := 0
	splashCount := 0

	for _, c := range cards {
		if c.Role != nil {
			switch *c.Role {
			case RoleWinCondition:
				hasWinCon = true
			case RoleSpellBig, RoleSpellSmall:
				hasSpell = true
			}
		}
		if c.Stats != nil && (c.Stats.Targets == "Air" || c.Stats.Targets == "Air & Ground") {
			airDefCount++
		}
		if c.Stats != nil && c.Stats.Radius > 0 {
			splashCount++
		}
	}

	score := 0.0

	// Win condition required
	if hasWinCon {
		score += 0.25
	}

	// Spell required
	if hasSpell {
		score += 0.15
	}

	// Air defense (min 2 ideal)
	if airDefCount >= 2 {
		score += 0.30
	} else if airDefCount == 1 {
		score += 0.15
	}

	// Splash damage (min 1 ideal)
	if splashCount >= 1 {
		score += 0.20
	}

	// Base coverage
	score += 0.10

	return math.Min(score, 1.0)
}

// calculateQuality computes average card quality.
func (s *ArchetypeFreeScorer) calculateQuality(cards []CardCandidate) float64 {
	if len(cards) == 0 {
		return 0
	}

	total := 0.0
	for _, c := range cards {
		total += c.LevelRatio()
	}

	return total / float64(len(cards))
}

// calculateElixirFit evaluates elixir curve.
func (s *ArchetypeFreeScorer) calculateElixirFit(cards []CardCandidate) float64 {
	if len(cards) == 0 {
		return 0
	}

	totalElixir := 0
	for _, c := range cards {
		totalElixir += c.Elixir
	}
	avgElixir := float64(totalElixir) / float64(len(cards))

	// Optimal range: 3.0-3.8
	if avgElixir >= 3.0 && avgElixir <= 3.8 {
		return 1.0
	}

	// Penalty for being outside range
	if avgElixir < 3.0 {
		return 1.0 - ((3.0 - avgElixir) / 2.0)
	}
	return 1.0 - ((avgElixir - 3.8) / 2.0)
}

// =============================================================================
// 4. Role-First Composition Builder
// =============================================================================

// RoleFirstBuilder fills deck slots by role priority.
// Order: WinCon -> AirDefense -> Splash -> TankKiller -> Spells -> Support -> Cycle
type RoleFirstBuilder struct {
	config BuilderConfig
}

// NewRoleFirstBuilder creates a role-priority builder.
func NewRoleFirstBuilder(config BuilderConfig) *RoleFirstBuilder {
	return &RoleFirstBuilder{config: config}
}

func (b *RoleFirstBuilder) Name() string {
	return "role_first"
}

// roleSlot defines a slot with role requirements.
type roleSlot struct {
	name     string
	roles    []CardRole
	filter   func(*CardCandidate) bool
	priority int
}

// Build fills slots in priority order.
//
//nolint:gocyclo // Slot fallback logic is intentionally explicit.
func (b *RoleFirstBuilder) Build() ([]string, error) {
	if len(b.config.Candidates) < 8 {
		return nil, fmt.Errorf("insufficient candidates: need at least 8, got %d", len(b.config.Candidates))
	}

	// Define slot priorities
	slots := []roleSlot{
		{name: "WinCondition", roles: []CardRole{RoleWinCondition}, priority: 1},
		{name: "AirDefense", filter: func(c *CardCandidate) bool {
			return c.Stats != nil && (c.Stats.Targets == "Air" || c.Stats.Targets == "Air & Ground")
		}, priority: 2},
		{name: "BigSpell", roles: []CardRole{RoleSpellBig}, priority: 3},
		{name: "SmallSpell", roles: []CardRole{RoleSpellSmall}, priority: 4},
		{name: "TankKiller", filter: func(c *CardCandidate) bool {
			// High DPS cards
			return c.Stats != nil && c.Stats.DamagePerSecond >= 150
		}, priority: 5},
		{name: "Splash", filter: func(c *CardCandidate) bool {
			return c.Stats != nil && c.Stats.Radius > 0
		}, priority: 6},
		{name: "Support1", roles: []CardRole{RoleSupport}, priority: 7},
		{name: "Flex", priority: 8},
	}

	deck := make([]string, 0, 8)
	used := make(map[string]bool)

	for _, slot := range slots {
		if len(deck) >= 8 {
			break
		}

		best := b.selectForSlot(slot, deck, used)
		if best == "" {
			// Slot is optional if we can't fill it
			continue
		}

		deck = append(deck, best)
		used[best] = true
	}

	// Fill remaining with best available
	for len(deck) < 8 {
		best := b.selectBestRemaining(deck, used)
		if best == "" {
			return nil, fmt.Errorf("cannot complete deck: stuck at %d cards", len(deck))
		}
		deck = append(deck, best)
		used[best] = true
	}

	return deck, nil
}

// selectForSlot finds the best card for a role slot.
//
//nolint:gocyclo // Mixed role/filter matching requires multiple early exits.
func (b *RoleFirstBuilder) selectForSlot(slot roleSlot, deck []string, used map[string]bool) string {
	type scored struct {
		name  string
		score float64
	}

	var matches []scored

	for _, c := range b.config.Candidates {
		if used[c.Name] {
			continue
		}

		// Check role match
		roleMatch := len(slot.roles) == 0
		for _, r := range slot.roles {
			if c.Role != nil && *c.Role == r {
				roleMatch = true
				break
			}
		}

		// Check filter match
		filterMatch := slot.filter == nil || slot.filter(c)

		if !roleMatch && !filterMatch {
			continue
		}

		// Score: level ratio * synergy with deck
		score := c.LevelRatio()
		if b.config.SynergyDB != nil {
			for _, deckCard := range deck {
				score += b.config.SynergyDB.GetSynergy(c.Name, deckCard) * 0.5
			}
		}

		matches = append(matches, scored{name: c.Name, score: score})
	}

	if len(matches) == 0 {
		return ""
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	return matches[0].name
}

// selectBestRemaining picks the best unused card.
func (b *RoleFirstBuilder) selectBestRemaining(deck []string, used map[string]bool) string {
	type scored struct {
		name  string
		score float64
	}

	var candidates []scored

	for _, c := range b.config.Candidates {
		if used[c.Name] {
			continue
		}

		score := c.LevelRatio()
		if b.config.SynergyDB != nil {
			for _, deckCard := range deck {
				score += b.config.SynergyDB.GetSynergy(c.Name, deckCard) * 0.3
			}
		}

		candidates = append(candidates, scored{name: c.Name, score: score})
	}

	if len(candidates) == 0 {
		return ""
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	return candidates[0].name
}

// =============================================================================
// 5. Counter-Centric Builder
// =============================================================================

// CounterCentricBuilder greedily adds cards that cover uncovered threat categories.
// Prioritizes defensive coverage while ensuring a viable offense.
type CounterCentricBuilder struct {
	config BuilderConfig
}

// NewCounterCentricBuilder creates a counter-focused builder.
func NewCounterCentricBuilder(config BuilderConfig) *CounterCentricBuilder {
	return &CounterCentricBuilder{config: config}
}

func (b *CounterCentricBuilder) Name() string {
	return "counter_centric"
}

// Build constructs a deck maximizing threat coverage.
func (b *CounterCentricBuilder) Build() ([]string, error) {
	if len(b.config.Candidates) < 8 {
		return nil, fmt.Errorf("insufficient candidates: need at least 8, got %d", len(b.config.Candidates))
	}

	if b.config.CounterMatrix == nil {
		return nil, fmt.Errorf("counter matrix is required")
	}

	deck := make([]string, 0, 8)
	used := make(map[string]bool)

	// Track covered categories
	coveredCategories := make(map[CounterCategory]int)

	// Priority categories to cover
	priorityCategories := []CounterCategory{
		CounterAirDefense,
		CounterTankKillers,
		CounterSplashDefense,
		CounterSwarmClear,
		CounterBuildings,
	}

	// First, ensure we have a win condition
	winCon := b.selectBestWinCondition(used)
	if winCon != "" {
		deck = append(deck, winCon)
		used[winCon] = true
		b.updateCoverage(winCon, coveredCategories)
	}

	// Then fill to maximize coverage
	for len(deck) < 8 {
		best := b.selectBestCoverage(deck, used, coveredCategories, priorityCategories)
		if best == "" {
			// Fallback to any available card
			best = b.selectAnyAvailable(used)
			if best == "" {
				return nil, fmt.Errorf("cannot complete deck: stuck at %d cards", len(deck))
			}
		}

		deck = append(deck, best)
		used[best] = true
		b.updateCoverage(best, coveredCategories)
	}

	return deck, nil
}

// selectBestWinCondition finds a high-quality win condition.
func (b *CounterCentricBuilder) selectBestWinCondition(used map[string]bool) string {
	type scored struct {
		name  string
		score float64
	}

	var winCons []scored

	for _, c := range b.config.Candidates {
		if used[c.Name] {
			continue
		}
		if c.Role == nil || *c.Role != RoleWinCondition {
			continue
		}

		// Score by level and counter capabilities
		score := c.LevelRatio()
		caps := b.config.CounterMatrix.GetCardCapabilities(c.Name)
		score += float64(len(caps)) * 0.1

		winCons = append(winCons, scored{name: c.Name, score: score})
	}

	if len(winCons) == 0 {
		return ""
	}

	sort.Slice(winCons, func(i, j int) bool {
		return winCons[i].score > winCons[j].score
	})

	return winCons[0].name
}

// selectBestCoverage finds the card that covers the most uncovered categories.
//
//nolint:funlen,gocognit,gocyclo // Coverage scoring intentionally keeps priority/redundancy logic localized.
func (b *CounterCentricBuilder) selectBestCoverage(deck []string, used map[string]bool, covered map[CounterCategory]int, priorities []CounterCategory) string {
	type scored struct {
		name  string
		score float64
	}

	var candidates []scored

	for _, c := range b.config.Candidates {
		if used[c.Name] {
			continue
		}

		score := 0.0

		// Score based on coverage of uncovered categories
		caps := b.config.CounterMatrix.GetCardCapabilities(c.Name)
		for _, cap := range caps {
			// Higher score for uncovered priority categories
			for priority, pc := range priorities {
				if cap == pc {
					current := covered[cap]
					if current == 0 {
						score += 1.0 - float64(priority)*0.1 // Higher priority = higher score
					} else if current == 1 && (cap == CounterAirDefense || cap == CounterSplashDefense) {
						score += 0.3 // Bonus for redundancy in critical categories
					}
				}
			}
		}

		// Bonus for spells (needed for deck completeness)
		if c.Role != nil && (*c.Role == RoleSpellBig || *c.Role == RoleSpellSmall) {
			hasSpell := false
			for _, d := range deck {
				dc := b.findCandidate(d)
				if dc != nil && dc.Role != nil && (*dc.Role == RoleSpellBig || *dc.Role == RoleSpellSmall) {
					hasSpell = true
					break
				}
			}
			if !hasSpell {
				score += 0.5
			}
		}

		// Level quality bonus
		score += c.LevelRatio() * 0.2

		// Synergy bonus
		if b.config.SynergyDB != nil {
			for _, d := range deck {
				score += b.config.SynergyDB.GetSynergy(c.Name, d) * 0.1
			}
		}

		if score > 0 {
			candidates = append(candidates, scored{name: c.Name, score: score})
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	return candidates[0].name
}

// updateCoverage adds a card's capabilities to the coverage map.
func (b *CounterCentricBuilder) updateCoverage(cardName string, covered map[CounterCategory]int) {
	caps := b.config.CounterMatrix.GetCardCapabilities(cardName)
	for _, cap := range caps {
		covered[cap]++
	}
}

// selectAnyAvailable returns any unused card.
func (b *CounterCentricBuilder) selectAnyAvailable(used map[string]bool) string {
	for _, c := range b.config.Candidates {
		if !used[c.Name] {
			return c.Name
		}
	}
	return ""
}

// findCandidate looks up a candidate by name.
func (b *CounterCentricBuilder) findCandidate(name string) *CardCandidate {
	for _, c := range b.config.Candidates {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// =============================================================================
// 6. Meta-Learning Co-occurrence Builder
// =============================================================================

// CoOccurrenceMatrix stores P(card_B | card_A) probabilities.
type CoOccurrenceMatrix struct {
	// matrix[cardA][cardB] = probability that cardB appears when cardA is in deck
	matrix map[string]map[string]float64
}

// NewCoOccurrenceMatrix creates an empty co-occurrence matrix.
func NewCoOccurrenceMatrix() *CoOccurrenceMatrix {
	return &CoOccurrenceMatrix{
		matrix: make(map[string]map[string]float64),
	}
}

// LearnFromDecks builds the co-occurrence matrix from sample decks.
func (m *CoOccurrenceMatrix) LearnFromDecks(decks [][]string) {
	// Count co-occurrences
	cardCount := make(map[string]int)
	pairCount := make(map[string]map[string]int)

	for _, deck := range decks {
		// Track which cards appeared together
		for _, cardA := range deck {
			cardCount[cardA]++
			if pairCount[cardA] == nil {
				pairCount[cardA] = make(map[string]int)
			}
			for _, cardB := range deck {
				if cardA != cardB {
					pairCount[cardA][cardB]++
				}
			}
		}
	}

	// Convert to probabilities
	for cardA, pairs := range pairCount {
		if m.matrix[cardA] == nil {
			m.matrix[cardA] = make(map[string]float64)
		}
		totalA := cardCount[cardA]
		for cardB, count := range pairs {
			m.matrix[cardA][cardB] = float64(count) / float64(totalA)
		}
	}
}

// GetProbability returns P(cardB | cardA).
func (m *CoOccurrenceMatrix) GetProbability(cardA, cardB string) float64 {
	if pairs, exists := m.matrix[cardA]; exists {
		if prob, exists := pairs[cardB]; exists {
			return prob
		}
	}
	return 0.0
}

// MetaLearningBuilder uses co-occurrence data to build decks.
type MetaLearningBuilder struct {
	config           BuilderConfig
	coOccurrenceData *CoOccurrenceMatrix
}

// NewMetaLearningBuilder creates a builder using learned co-occurrence.
func NewMetaLearningBuilder(config BuilderConfig, coOccurrence *CoOccurrenceMatrix) *MetaLearningBuilder {
	return &MetaLearningBuilder{
		config:           config,
		coOccurrenceData: coOccurrence,
	}
}

func (b *MetaLearningBuilder) Name() string {
	return "meta_learning"
}

// Build constructs a deck using co-occurrence probabilities.
func (b *MetaLearningBuilder) Build() ([]string, error) {
	if len(b.config.Candidates) < 8 {
		return nil, fmt.Errorf("insufficient candidates: need at least 8, got %d", len(b.config.Candidates))
	}

	deck := make([]string, 0, 8)
	used := make(map[string]bool)

	// Start with a win condition
	startCard := b.selectStartCard(used)
	if startCard == "" {
		return nil, fmt.Errorf("no valid starting card")
	}
	deck = append(deck, startCard)
	used[startCard] = true

	// Add cards based on co-occurrence probability
	for len(deck) < 8 {
		best := b.selectBestCoOccurrence(deck, used)
		if best == "" {
			// Fallback
			best = b.selectAnyValid(used)
			if best == "" {
				return nil, fmt.Errorf("cannot complete deck")
			}
		}
		deck = append(deck, best)
		used[best] = true
	}

	return deck, nil
}

// selectStartCard finds a good starting win condition.
func (b *MetaLearningBuilder) selectStartCard(used map[string]bool) string {
	for _, c := range b.config.Candidates {
		if used[c.Name] {
			continue
		}
		if c.Role != nil && *c.Role == RoleWinCondition {
			return c.Name
		}
	}
	// Fallback to any card
	if len(b.config.Candidates) > 0 {
		return b.config.Candidates[rand.IntN(len(b.config.Candidates))].Name
	}
	return ""
}

// selectBestCoOccurrence finds the card with highest co-occurrence with deck.
func (b *MetaLearningBuilder) selectBestCoOccurrence(deck []string, used map[string]bool) string {
	type scored struct {
		name  string
		score float64
	}

	var candidates []scored

	for _, c := range b.config.Candidates {
		if used[c.Name] {
			continue
		}

		// Sum co-occurrence probabilities with all deck cards
		score := 0.0
		if b.coOccurrenceData != nil {
			for _, deckCard := range deck {
				score += b.coOccurrenceData.GetProbability(deckCard, c.Name)
			}
		}

		// Add level quality
		score += c.LevelRatio() * 0.1

		// Add synergy
		if b.config.SynergyDB != nil {
			for _, deckCard := range deck {
				score += b.config.SynergyDB.GetSynergy(c.Name, deckCard) * 0.05
			}
		}

		candidates = append(candidates, scored{name: c.Name, score: score})
	}

	if len(candidates) == 0 {
		return ""
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	return candidates[0].name
}

// selectAnyValid returns any unused card.
func (b *MetaLearningBuilder) selectAnyValid(used map[string]bool) string {
	for _, c := range b.config.Candidates {
		if !used[c.Name] {
			return c.Name
		}
	}
	return ""
}

// =============================================================================
// Builder Factory
// =============================================================================

// BuilderType identifies the deck building strategy.
type BuilderType string

const (
	BuilderSynergyGraph           BuilderType = "synergy_graph"
	BuilderConstraintSatisfaction BuilderType = "constraint_satisfaction"
	BuilderRoleFirst              BuilderType = "role_first"
	BuilderCounterCentric         BuilderType = "counter_centric"
	BuilderMetaLearning           BuilderType = "meta_learning"
)

// CreateBuilder creates a deck builder of the specified type.
//
//nolint:ireturn // Factory intentionally returns the common DeckBuilder interface.
func CreateBuilder(builderType BuilderType, config BuilderConfig) (DeckBuilder, error) {
	switch builderType {
	case BuilderSynergyGraph:
		if config.SynergyDB == nil {
			return nil, fmt.Errorf("synergy database required for synergy graph builder")
		}
		return NewSynergyGraphBuilder(config), nil

	case BuilderConstraintSatisfaction:
		return NewConstraintSatisfactionBuilder(config), nil

	case BuilderRoleFirst:
		return NewRoleFirstBuilder(config), nil

	case BuilderCounterCentric:
		if config.CounterMatrix == nil {
			return nil, fmt.Errorf("counter matrix required for counter-centric builder")
		}
		return NewCounterCentricBuilder(config), nil

	case BuilderMetaLearning:
		return NewMetaLearningBuilder(config, NewCoOccurrenceMatrix()), nil

	default:
		return nil, fmt.Errorf("unknown builder type: %s", builderType)
	}
}

// BuildMultiple runs multiple builders and returns all results.
func BuildMultiple(config BuilderConfig, builderTypes ...BuilderType) map[BuilderType][]string {
	results := make(map[BuilderType][]string)

	for _, bt := range builderTypes {
		builder, err := CreateBuilder(bt, config)
		if err != nil {
			continue
		}

		deck, err := builder.Build()
		if err != nil {
			continue
		}

		results[bt] = deck
	}

	return results
}
