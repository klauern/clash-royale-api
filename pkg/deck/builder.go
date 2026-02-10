// Package deck provides intelligent deck building functionality for Clash Royale
package deck

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/internal/util"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// Builder handles the construction of balanced Clash Royale decks
// from player card analysis data.
type Builder struct {
	dataDir                  string
	unlockedEvolutions       map[string]bool
	evolutionSlotLimit       int
	statsRegistry            *clashroyale.CardStatsRegistry
	strategy                 Strategy
	strategyConfig           StrategyConfig
	levelCurve               *LevelCurve
	synergyDB                *SynergyDatabase
	synergyEnabled           bool
	synergyWeight            float64
	synergyCache             map[string]float64 // Cache for synergy lookups: "card1|card2" -> score
	uniquenessEnabled        bool
	uniquenessWeight         float64
	uniquenessScorer         *UniquenessScorer
	avoidArchetypes          []string                  // Archetypes to avoid when building decks
	archetypeAvoidanceScorer *ArchetypeAvoidanceScorer // Scorer for archetype avoidance
	includeCards             []string                  // Cards to force into the deck
	excludeCards             []string                  // Cards to exclude from consideration
	fuzzIntegration          *FuzzIntegration          // Fuzz stats integration for data-driven card scoring
}

// NewBuilder creates a new deck builder instance
func NewBuilder(dataDir string) *Builder {
	if dataDir == "" {
		dataDir = "data"
	}

	// Parse UNLOCKED_EVOLUTIONS environment variable
	unlockedEvos := make(map[string]bool)
	if envEvos := os.Getenv("UNLOCKED_EVOLUTIONS"); envEvos != "" {
		for _, card := range strings.Split(envEvos, ",") {
			cardName := strings.TrimSpace(card)
			if cardName != "" {
				unlockedEvos[cardName] = true
			}
		}
	}

	builder := &Builder{
		dataDir:            dataDir,
		unlockedEvolutions: unlockedEvos,
		evolutionSlotLimit: 2,
		synergyCache:       make(map[string]float64),
	}

	// Try to load combat stats
	statsPath := filepath.Join(dataDir, "cards_stats.json")
	if stats, err := clashroyale.LoadStats(statsPath); err == nil {
		builder.statsRegistry = stats
	} else {
		// Log error if needed, for now just silent failure or maybe print to stderr
		// fmt.Fprintf(os.Stderr, "Warning: Failed to load card stats from %s: %v\n", statsPath, err)
	}

	// Initialize level curve framework
	levelCurvePath := "config/card_level_curves.json"
	if lc, err := NewLevelCurve(levelCurvePath); err == nil {
		builder.levelCurve = lc
	} else {
		// Log error if needed, for now just silent failure
		// fmt.Fprintf(os.Stderr, "Warning: Failed to load level curves from %s: %v\n", levelCurvePath, err)
	}

	// Initialize with balanced strategy by default
	builder.strategy = StrategyBalanced
	builder.strategyConfig = GetStrategyConfig(StrategyBalanced)

	// Initialize synergy database
	builder.synergyDB = NewSynergyDatabase()
	builder.synergyEnabled = false // Disabled by default, enabled via CLI flag
	builder.synergyWeight = 0.15   // Default: 15% of total score from synergy

	// Initialize uniqueness scorer with default configuration
	uniquenessConfig := UniquenessConfig{
		Enabled:                false, // Disabled by default, enabled via CLI flag
		Weight:                 0.2,   // Default: 20% weight when enabled
		MinUniquenessThreshold: 0.3,   // Only bonus for cards below 30% popularity
		UseGeometricMean:       false, // Use arithmetic mean for deck-level uniqueness
	}
	builder.uniquenessScorer = NewUniquenessScorer(uniquenessConfig)
	builder.uniquenessEnabled = false // Disabled by default
	builder.uniquenessWeight = 0.2    // Default: 20% of total score from uniqueness

	return builder
}

// CardAnalysis represents the input analysis data structure
type CardAnalysis struct {
	CardLevels   map[string]CardLevelData `json:"card_levels"`
	AnalysisTime string                   `json:"analysis_time,omitempty"`
	PlayerName   string                   `json:"player_name,omitempty"`
	PlayerTag    string                   `json:"player_tag,omitempty"`
}

// CardLevelData represents card level and metadata from analysis
type CardLevelData struct {
	Level             int     `json:"level"`
	MaxLevel          int     `json:"max_level"`
	Rarity            string  `json:"rarity"`
	Elixir            int     `json:"elixir,omitempty"`
	EvolutionLevel    int     `json:"evolution_level,omitempty"`     // Current evolution level (0-3)
	MaxEvolutionLevel int     `json:"max_evolution_level,omitempty"` // Maximum possible evolution level
	ScoreBoost        float64 `json:"score_boost,omitempty"`         // Boost for preferred cards (doesn't affect display level)
}

// BuildDeckFromAnalysis creates a deck recommendation from card analysis data
func (b *Builder) BuildDeckFromAnalysis(analysis CardAnalysis) (*DeckRecommendation, error) {
	if len(analysis.CardLevels) == 0 {
		return nil, fmt.Errorf("analysis data missing 'card_levels'")
	}

	b.clearSynergyCache()
	candidates := b.buildCandidates(analysis.CardLevels)
	candidates = b.filterExcludedCards(candidates)

	deck := make([]*CardCandidate, 0)
	used := make(map[string]bool)

	deck, used = b.addIncludedCards(deck, candidates, used)
	deck, used, notes := b.selectCardsByRole(deck, candidates, used)
	deck = b.fillRemainingSlots(deck, candidates, used)

	evolutionSlots := b.selectEvolutionSlots(deck)
	recommendation := b.buildRecommendationDetails(deck, analysis.AnalysisTime, evolutionSlots, notes)
	b.finalizeRecommendation(recommendation)

	return recommendation, nil
}

// filterExcludedCards removes excluded cards from the candidate pool
func (b *Builder) filterExcludedCards(candidates []*CardCandidate) []*CardCandidate {
	if len(b.excludeCards) == 0 {
		return candidates
	}

	excludeMap := make(map[string]bool)
	for _, card := range b.excludeCards {
		excludeMap[card] = true
	}
	filtered := make([]*CardCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if !excludeMap[candidate.Name] {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

// addIncludedCards adds force-included cards to deck and marks them as used
func (b *Builder) addIncludedCards(deck, candidates []*CardCandidate, used map[string]bool) ([]*CardCandidate, map[string]bool) {
	if len(b.includeCards) == 0 {
		return deck, used
	}

	for _, cardName := range b.includeCards {
		// Find the card in candidates
		for _, candidate := range candidates {
			if candidate.Name == cardName {
				deck = append(deck, candidate)
				used[candidate.Name] = true
				break
			}
		}
	}
	return deck, used
}

// selectCardsByRole selects cards for each role with composition override support
// getOverrideCount returns the override count for a role, or the default if not overridden
//
//nolint:gocyclo // Role override matrix requires explicit precedence rules.
func (b *Builder) getOverrideCount(role CardRole, defaultCount int) int {
	if b.strategyConfig.CompositionOverrides == nil {
		return defaultCount
	}

	override := b.strategyConfig.CompositionOverrides
	switch role {
	case RoleWinCondition:
		if override.WinConditions != nil {
			return *override.WinConditions
		}
	case RoleBuilding:
		if override.Buildings != nil {
			return *override.Buildings
		}
	case RoleSpellBig:
		if override.BigSpells != nil {
			return *override.BigSpells
		}
	case RoleSpellSmall:
		if override.SmallSpells != nil {
			return *override.SmallSpells
		}
	case RoleSupport:
		if override.Support != nil {
			return *override.Support
		}
	case RoleCycle:
		if override.Cycle != nil {
			return *override.Cycle
		}
	}
	return defaultCount
}

// selectCardsForRole selects cards of a specific role using pickBest and updates deck/used
func (b *Builder) selectCardsForRole(role CardRole, count int, deck, candidates []*CardCandidate, used map[string]bool, trackMissing bool) ([]*CardCandidate, map[string]bool, []string) {
	notes := make([]string, 0)

	for i := 0; i < count; i++ {
		if card := b.pickBest(role, candidates, used, deck); card != nil {
			deck = append(deck, card)
			used[card.Name] = true
		} else if i == 0 && trackMissing {
			notes = append(notes, "No win condition found; selected highest power cards instead.")
		}
	}

	return deck, used, notes
}

// selectMultipleCardsForRole selects multiple cards of a role using pickMany and updates deck/used
func (b *Builder) selectMultipleCardsForRole(role CardRole, count int, deck, candidates []*CardCandidate, used map[string]bool) ([]*CardCandidate, map[string]bool) {
	cards := b.pickMany(role, candidates, used, count, deck)
	deck = append(deck, cards...)
	for _, card := range cards {
		used[card.Name] = true
	}
	return deck, used
}

// Returns notes for missing win conditions
func (b *Builder) selectCardsByRole(deck, candidates []*CardCandidate, used map[string]bool) ([]*CardCandidate, map[string]bool, []string) {
	notes := make([]string, 0)

	// Core roles: win condition, building, two spells
	// Use override counts if specified, otherwise use defaults
	winConditionCount := b.getOverrideCount(RoleWinCondition, 1)
	deck, used, winConditionNotes := b.selectCardsForRole(RoleWinCondition, winConditionCount, deck, candidates, used, true)
	notes = append(notes, winConditionNotes...)

	buildingCount := b.getOverrideCount(RoleBuilding, 1)
	deck, used, _ = b.selectCardsForRole(RoleBuilding, buildingCount, deck, candidates, used, false)

	bigSpellCount := b.getOverrideCount(RoleSpellBig, 1)
	deck, used, _ = b.selectCardsForRole(RoleSpellBig, bigSpellCount, deck, candidates, used, false)

	smallSpellCount := b.getOverrideCount(RoleSpellSmall, 1)
	deck, used, _ = b.selectCardsForRole(RoleSpellSmall, smallSpellCount, deck, candidates, used, false)

	// Support backbone (2 cards, or override count if specified)
	supportCount := b.getOverrideCount(RoleSupport, 2)
	deck, used = b.selectMultipleCardsForRole(RoleSupport, supportCount, deck, candidates, used)

	// Cheap cycle fillers (2 cards, or override count if specified)
	cycleCount := b.getOverrideCount(RoleCycle, 2)
	deck, used = b.selectMultipleCardsForRole(RoleCycle, cycleCount, deck, candidates, used)

	return deck, used, notes
}

// fillRemainingSlots fills remaining deck slots (up to 8) with highest-scoring unused cards
func (b *Builder) fillRemainingSlots(deck, candidates []*CardCandidate, used map[string]bool) []*CardCandidate {
	if len(deck) < 8 {
		remaining := b.getHighestScoreCards(candidates, used, 8-len(deck), deck)
		deck = append(deck, remaining...)
	}
	// Ensure exactly 8 cards
	return deck[:8]
}

// buildRecommendationDetails builds the DeckRecommendation struct and populates card details
func (b *Builder) buildRecommendationDetails(deck []*CardCandidate, analysisTime string, evolutionSlots, notes []string) *DeckRecommendation {
	recommendation := &DeckRecommendation{
		Deck:           make([]string, 8),
		DeckDetail:     make([]CardDetail, 8),
		AvgElixir:      b.calculateAvgElixir(deck),
		AnalysisTime:   analysisTime,
		Notes:          notes,
		EvolutionSlots: evolutionSlots,
	}

	for i, card := range deck {
		recommendation.Deck[i] = card.Name
		roleStr := ""
		if card.Role != nil {
			roleStr = string(*card.Role)
		}
		recommendation.DeckDetail[i] = CardDetail{
			Name:              card.Name,
			Level:             card.Level,
			MaxLevel:          card.MaxLevel,
			Rarity:            card.Rarity,
			Elixir:            card.Elixir,
			Role:              roleStr,
			Score:             roundToThree(card.Score),
			EvolutionLevel:    card.EvolutionLevel,
			MaxEvolutionLevel: card.MaxEvolutionLevel,
		}
	}

	return recommendation
}

// finalizeRecommendation adds strategic notes and evolution slot note to recommendation
func (b *Builder) finalizeRecommendation(recommendation *DeckRecommendation) {
	// Add strategic notes
	b.addStrategicNotes(recommendation)

	// Add evolution slot note if applicable
	if len(recommendation.EvolutionSlots) > 0 {
		slotNote := fmt.Sprintf("Evolution slots: %s", strings.Join(recommendation.EvolutionSlots, ", "))
		recommendation.AddNote(slotNote)
	}
}

// BuildDeckFromFile loads analysis from a file and builds a deck
func (b *Builder) BuildDeckFromFile(analysisPath string) (*DeckRecommendation, error) {
	analysis, err := b.LoadAnalysis(analysisPath)
	if err != nil {
		return nil, err
	}
	return b.BuildDeckFromAnalysis(*analysis)
}

// LoadAnalysis loads card analysis data from a JSON file
func (b *Builder) LoadAnalysis(analysisPath string) (*CardAnalysis, error) {
	data, err := os.ReadFile(analysisPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read analysis file: %w", err)
	}

	var analysis CardAnalysis
	if err := json.Unmarshal(data, &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis JSON: %w", err)
	}

	return &analysis, nil
}

// LoadLatestAnalysis loads the most recent analysis for a player
func (b *Builder) LoadLatestAnalysis(playerTag, analysisDir string) (*CardAnalysis, error) {
	if analysisDir == "" {
		analysisDir = filepath.Join(b.dataDir, "analysis")
	}

	if _, err := os.Stat(analysisDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("analysis directory does not exist: %s", analysisDir)
	}

	cleanTag := strings.TrimPrefix(playerTag, "#")
	pattern := fmt.Sprintf("*analysis*%s.json", cleanTag)

	matches, err := filepath.Glob(filepath.Join(analysisDir, pattern))
	if err != nil {
		return nil, fmt.Errorf("failed to glob analysis files: %w", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no analysis files found for player %s", playerTag)
	}

	// Sort by modification time (newest first)
	sort.Slice(matches, func(i, j int) bool {
		infoI, _ := os.Stat(matches[i])
		infoJ, _ := os.Stat(matches[j])
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return b.LoadAnalysis(matches[0])
}

// SaveDeck persists a deck recommendation to disk
func (b *Builder) SaveDeck(deckData *DeckRecommendation, outputDir, playerTag string) (string, error) {
	if outputDir == "" {
		outputDir = filepath.Join(b.dataDir, "decks")
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	cleanTag := strings.TrimPrefix(playerTag, "#")
	filename := fmt.Sprintf("%s_deck_%s.json", timestamp, cleanTag)
	path := filepath.Join(outputDir, filename)

	data, err := json.MarshalIndent(deckData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal deck data: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("failed to write deck file: %w", err)
	}

	return path, nil
}

// Helper methods

func (b *Builder) buildCandidates(cardLevels map[string]CardLevelData) []*CardCandidate {
	candidates := make([]*CardCandidate, 0, len(cardLevels))

	for name, data := range cardLevels {
		candidate := b.buildCandidate(name, data)
		candidates = append(candidates, candidate)
	}

	return candidates
}

func (b *Builder) buildCandidate(name string, data CardLevelData) *CardCandidate {
	level := data.Level
	maxLevel := data.MaxLevel
	if maxLevel == 0 {
		maxLevel = 1
	}
	rarity := data.Rarity
	if rarity == "" {
		rarity = "Common"
	}

	elixir := b.resolveElixir(name, data)
	role := b.inferRole(name)

	// Add evolution tracking
	hasEvolution := data.MaxEvolutionLevel > 0 && b.unlockedEvolutions[name]
	evoPriority := b.getEvolutionPriority(role)

	var stats *clashroyale.CombatStats
	if b.statsRegistry != nil {
		stats = b.statsRegistry.GetStats(name)
	}

	// Create candidate first (score will be calculated next)
	candidate := &CardCandidate{
		Name:              name,
		Level:             level,
		MaxLevel:          maxLevel,
		Rarity:            rarity,
		Elixir:            elixir,
		Role:              role,
		Score:             0, // Will be calculated below
		HasEvolution:      hasEvolution,
		EvolutionPriority: evoPriority,
		EvolutionLevel:    data.EvolutionLevel,
		MaxEvolutionLevel: data.MaxEvolutionLevel,
		Stats:             stats,
	}

	// Calculate score using strategy-aware scoring
	score := ScoreCardWithStrategy(candidate, role, b.strategyConfig, b.levelCurve)

	// Apply archetype-preferred card boost (if any)
	if data.ScoreBoost > 0 {
		score *= (1 + data.ScoreBoost)
	}

	// Apply fuzz boost if available
	if b.fuzzIntegration != nil && b.fuzzIntegration.HasStats() {
		score = b.fuzzIntegration.ApplyFuzzBoost(score, name)
	}

	candidate.Score = score
	return candidate
}

func (b *Builder) resolveElixir(name string, data CardLevelData) int {
	// Use config package to resolve elixir cost
	return config.GetCardElixir(name, data.Elixir)
}

func (b *Builder) inferRole(name string) *CardRole {
	// Use config package to get card role
	configRole := config.GetCardRole(name)
	if configRole != "" {
		// Convert config.CardRole to deck.CardRole (both are string types)
		deckRole := CardRole(configRole)
		return &deckRole
	}
	return nil
}

func (b *Builder) scoreCard(name string, level, maxLevel int, rarity string, elixir int, role *CardRole, maxEvolutionLevel int) float64 {
	levelRatio := float64(level) / float64(maxLevel)
	rarityBoost := config.GetRarityPriorityBonus(rarity) // Use priority bonus (4.0x for Legendary) instead of weight (1.15x)

	// Encourage cheaper cards slightly to keep cycle tight
	elixirWeight := 1.0 - float64(max(elixir-3, 0))/9.0

	roleBonus := config.RoleBonusValue
	if role == nil {
		roleBonus = 0
	}

	// Use level-scaled evolution bonus
	evolutionBonus := b.calculateEvolutionBonus(name, level, maxLevel, maxEvolutionLevel)

	return (levelRatio * config.LevelWeightFactor * rarityBoost) + (elixirWeight * config.ElixirWeightFactor) + roleBonus + evolutionBonus
}

// calculateEvolutionBonus returns level-scaled evolution bonus
// Formula: EvolutionBaseBonus * (level/maxLevel)^1.5 * (1 + 0.2*(maxEvoLevel-1))
// Additional bonus for cards with evolution-specific role overrides
// This rewards using higher-level cards and accounts for multi-evolution cards
func (b *Builder) calculateEvolutionBonus(cardName string, level, maxLevel, maxEvoLevel int) float64 {
	// Check if evolution is unlocked
	if !b.unlockedEvolutions[cardName] || maxEvoLevel == 0 {
		return 0.0
	}

	levelRatio := float64(level) / float64(maxLevel)
	scaledRatio := math.Pow(levelRatio, 1.5)

	// Bonus multiplier for multi-evolution cards (e.g., Knight with evo level 3)
	evoMultiplier := 1.0 + (0.2 * float64(maxEvoLevel-1))

	bonus := config.EvolutionBaseBonus * scaledRatio * evoMultiplier

	// Additional bonus for cards with evolution-specific role overrides
	// These cards have their strategic role changed by evolution, making them
	// more valuable in deck building
	if HasEvolutionOverride(cardName, 1) {
		// Add 10% extra bonus for evolution role overrides
		bonus *= 1.1
	}

	return bonus
}

// getEvolutionPriority returns priority value for role (lower = higher priority)
// Priority: Win Conditions (1) > Buildings (2) > Big Spells (3) > Support (4) > Small Spells (5) > Cycle (6)
func (b *Builder) getEvolutionPriority(role *CardRole) int {
	if role == nil {
		return 100 // Lowest priority for cards without roles
	}

	priorityMap := map[CardRole]int{
		RoleWinCondition: 1,
		RoleBuilding:     2,
		RoleSpellBig:     3,
		RoleSupport:      4,
		RoleSpellSmall:   5,
		RoleCycle:        6,
	}

	if priority, exists := priorityMap[*role]; exists {
		return priority
	}
	return 100
}

// selectEvolutionSlots chooses which cards use evolution slots based on role priority and card score
// When 3+ evolved cards exist in deck, selects top N (default 2) by role priority, then score
func (b *Builder) selectEvolutionSlots(deck []*CardCandidate) []string {
	// Filter to evolved cards only
	evolved := make([]*CardCandidate, 0)
	for _, card := range deck {
		if card.HasEvolution {
			evolved = append(evolved, card)
		}
	}

	// If we have <= slot limit, all evolved cards get slots
	if len(evolved) <= b.evolutionSlotLimit {
		slots := make([]string, len(evolved))
		for i, card := range evolved {
			slots[i] = card.Name
		}
		return slots
	}

	// Sort by priority (ASC - lower number = higher priority), then by score (DESC - higher is better)
	sort.Slice(evolved, func(i, j int) bool {
		if evolved[i].EvolutionPriority != evolved[j].EvolutionPriority {
			return evolved[i].EvolutionPriority < evolved[j].EvolutionPriority
		}
		return evolved[i].Score > evolved[j].Score
	})

	// Select top N by slot limit
	slots := make([]string, b.evolutionSlotLimit)
	for i := 0; i < b.evolutionSlotLimit; i++ {
		slots[i] = evolved[i].Name
	}
	return slots
}

// calculateSynergyScore computes the synergy bonus for a card based on its synergies
// with cards already in the deck. Returns a score between 0.0 and 1.0 where:
// - 0.0 means no synergies with current deck
// - 1.0 means perfect synergies (average score of 1.0 with all deck cards)
//
// The score is calculated as the average synergy score across all pairs between
// the candidate card and cards currently in the deck.
//
// Time Complexity:
// - Without cache: O(n) database lookups per call where n = deck size
// - With cache (warm): O(1) per cached pair lookup
// - During deck building: Each pair is queried once from DB, then cached
//
// Space Complexity: O(c²) where c is the number of unique cards in the pool
// (cache stores all pairwise combinations queried during deck building)
//
// The memoization cache is cleared at the start of each BuildDeckFromAnalysis call
// to prevent unbounded memory growth.
func (b *Builder) calculateSynergyScore(cardName string, deck []*CardCandidate) float64 {
	if b.synergyDB == nil || len(deck) == 0 {
		return 0.0
	}

	totalSynergy := 0.0
	synergyCount := 0

	// Check synergy with each card in the current deck
	for _, deckCard := range deck {
		if synergyScore := b.getCachedSynergy(cardName, deckCard.Name); synergyScore > 0 {
			totalSynergy += synergyScore
			synergyCount++
		}
	}

	// Return average synergy score (0.0 if no synergies found)
	if synergyCount == 0 {
		return 0.0
	}

	return totalSynergy / float64(synergyCount)
}

// getCachedSynergy retrieves a synergy score from cache or queries the database.
//
// The cache key is ordered (alphabetically) to ensure card1+card2 and card2+card1
// use the same cache entry. This is important because synergy is symmetric:
// the synergy between "Giant" and "Witch" is the same as between "Witch" and "Giant".
//
// Performance: O(1) for cache hits, O(1) for database query (cached internally by
// the synergy database). The main benefit is avoiding repeated map lookups in the
// underlying synergy database during deck building.
func (b *Builder) getCachedSynergy(card1, card2 string) float64 {
	// Create ordered cache key to ensure card1+card2 == card2+card1
	var key string
	if card1 < card2 {
		key = card1 + "|" + card2
	} else {
		key = card2 + "|" + card1
	}

	// Check cache first - O(1) map lookup
	if score, exists := b.synergyCache[key]; exists {
		return score
	}

	// Cache miss: query from database and cache the result
	score := b.synergyDB.GetSynergy(card1, card2)
	b.synergyCache[key] = score
	return score
}

// clearSynergyCache clears the memoization cache. Should be called between
// deck builds to prevent unbounded cache growth.
//
// During a typical deck build with 100 candidate cards and 8 deck slots:
// - Maximum unique pairs: C(100, 2) = 4,950
// - Actual pairs queried: ~100 * 8 = 800 (each candidate checked against current deck)
// - Cache size after build: ~800 entries
//
// By clearing between builds, we ensure the cache doesn't grow indefinitely
// when the builder is reused for multiple deck recommendations.
func (b *Builder) clearSynergyCache() {
	b.synergyCache = make(map[string]float64)
}

func (b *Builder) pickBest(role CardRole, candidates []*CardCandidate, used map[string]bool, currentDeck []*CardCandidate) *CardCandidate {
	// Convert deck.CardRole to config.CardRole
	roleCards := config.GetRoleCards(config.CardRole(role))
	if roleCards == nil {
		return nil
	}

	var pool []*CardCandidate
	for _, candidate := range candidates {
		if !used[candidate.Name] && b.contains(roleCards, candidate.Name) {
			pool = append(pool, candidate)
		}
	}

	if len(pool) == 0 {
		return nil
	}

	// Apply synergy bonuses if enabled
	if b.synergyEnabled && len(currentDeck) > 0 {
		for _, candidate := range pool {
			synergyBonus := b.calculateSynergyScore(candidate.Name, currentDeck)
			candidate.Score += synergyBonus * b.synergyWeight
		}
	}

	// Apply uniqueness bonuses if enabled
	if b.uniquenessEnabled && b.uniquenessScorer != nil {
		for _, candidate := range pool {
			// Build hypothetical deck with this candidate
			deckCardNames := make([]string, 0, len(currentDeck)+1)
			for _, card := range currentDeck {
				deckCardNames = append(deckCardNames, card.Name)
			}
			deckCardNames = append(deckCardNames, candidate.Name)

			// Calculate uniqueness score for this hypothetical deck
			uniquenessScore := b.uniquenessScorer.ScoreDeck(deckCardNames)
			candidate.Score += uniquenessScore * b.uniquenessWeight
		}
	}

	// Apply archetype avoidance penalties if configured
	if b.archetypeAvoidanceScorer != nil && b.archetypeAvoidanceScorer.IsEnabled() {
		for _, candidate := range pool {
			penalty := b.archetypeAvoidanceScorer.ScoreCard(candidate.Name)
			candidate.Score += penalty
		}
	}

	// Return highest scoring card
	sort.Slice(pool, func(i, j int) bool {
		return pool[i].Score > pool[j].Score
	})

	return pool[0]
}

func (b *Builder) pickMany(role CardRole, candidates []*CardCandidate, used map[string]bool, count int, currentDeck []*CardCandidate) []*CardCandidate {
	// Convert deck.CardRole to config.CardRole
	roleCards := config.GetRoleCards(config.CardRole(role))
	if roleCards == nil {
		return nil
	}

	var pool []*CardCandidate
	for _, candidate := range candidates {
		if !used[candidate.Name] && b.contains(roleCards, candidate.Name) {
			pool = append(pool, candidate)
		}
	}

	// Apply synergy bonuses if enabled
	if b.synergyEnabled && len(currentDeck) > 0 {
		for _, candidate := range pool {
			synergyBonus := b.calculateSynergyScore(candidate.Name, currentDeck)
			candidate.Score += synergyBonus * b.synergyWeight
		}
	}

	// Apply uniqueness bonuses if enabled
	if b.uniquenessEnabled && b.uniquenessScorer != nil {
		for _, candidate := range pool {
			// Build hypothetical deck with this candidate
			deckCardNames := make([]string, 0, len(currentDeck)+1)
			for _, card := range currentDeck {
				deckCardNames = append(deckCardNames, card.Name)
			}
			deckCardNames = append(deckCardNames, candidate.Name)

			// Calculate uniqueness score for this hypothetical deck
			uniquenessScore := b.uniquenessScorer.ScoreDeck(deckCardNames)
			candidate.Score += uniquenessScore * b.uniquenessWeight
		}
	}

	// Apply archetype avoidance penalties if configured
	if b.archetypeAvoidanceScorer != nil && b.archetypeAvoidanceScorer.IsEnabled() {
		for _, candidate := range pool {
			penalty := b.archetypeAvoidanceScorer.ScoreCard(candidate.Name)
			candidate.Score += penalty
		}
	}

	sort.Slice(pool, func(i, j int) bool {
		return pool[i].Score > pool[j].Score
	})

	if len(pool) < count {
		return pool
	}

	return pool[:count]
}

//nolint:gocyclo // Selection logic intentionally branches on availability/constraints.
func (b *Builder) getHighestScoreCards(candidates []*CardCandidate, used map[string]bool, count int, currentDeck []*CardCandidate) []*CardCandidate {
	var pool []*CardCandidate
	for _, candidate := range candidates {
		if !used[candidate.Name] {
			pool = append(pool, candidate)
		}
	}

	// Apply synergy bonuses if enabled
	if b.synergyEnabled && len(currentDeck) > 0 {
		for _, candidate := range pool {
			synergyBonus := b.calculateSynergyScore(candidate.Name, currentDeck)
			candidate.Score += synergyBonus * b.synergyWeight
		}
	}

	// Apply uniqueness bonuses if enabled
	if b.uniquenessEnabled && b.uniquenessScorer != nil {
		for _, candidate := range pool {
			// Build hypothetical deck with this candidate
			deckCardNames := make([]string, 0, len(currentDeck)+1)
			for _, card := range currentDeck {
				deckCardNames = append(deckCardNames, card.Name)
			}
			deckCardNames = append(deckCardNames, candidate.Name)

			// Calculate uniqueness score for this hypothetical deck
			uniquenessScore := b.uniquenessScorer.ScoreDeck(deckCardNames)
			candidate.Score += uniquenessScore * b.uniquenessWeight
		}
	}

	// Apply archetype avoidance penalties if configured
	if b.archetypeAvoidanceScorer != nil && b.archetypeAvoidanceScorer.IsEnabled() {
		for _, candidate := range pool {
			penalty := b.archetypeAvoidanceScorer.ScoreCard(candidate.Name)
			candidate.Score += penalty
		}
	}

	sort.Slice(pool, func(i, j int) bool {
		return pool[i].Score > pool[j].Score
	})

	if len(pool) < count {
		return pool
	}

	return pool[:count]
}

func (b *Builder) calculateAvgElixir(deck []*CardCandidate) float64 {
	avg := util.CalcAvgElixir(deck, func(card *CardCandidate) int {
		return card.Elixir
	})
	return roundToTwo(avg)
}

func (b *Builder) addStrategicNotes(recommendation *DeckRecommendation) {
	hasBuilding := false
	hasSpell := false

	for _, card := range recommendation.DeckDetail {
		if card.Role == string(RoleBuilding) {
			hasBuilding = true
		}
		if card.Role == string(RoleSpellBig) || card.Role == string(RoleSpellSmall) {
			hasSpell = true
		}
	}

	if !hasBuilding {
		recommendation.AddNote("No defensive building available; play troops high to kite.")
	}
	if !hasSpell {
		recommendation.AddNote("No spell picked; beware of swarm matchups.")
	}
	if recommendation.AvgElixir > 3.8 {
		recommendation.AddNote("High average elixir; play patiently and build pushes.")
	} else if recommendation.AvgElixir < 2.8 {
		recommendation.AddNote("Low average elixir; pressure often and out-cycle counters.")
	}
}

// Utility functions

func (b *Builder) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func roundToTwo(value float64) float64 {
	return float64(int(value*100)) / 100
}

func roundToThree(value float64) float64 {
	return float64(int(value*1000)) / 1000
}

// LoadDeckFromFile loads a deck recommendation from a JSON file
func (b *Builder) LoadDeckFromFile(deckPath string) (*DeckRecommendation, error) {
	data, err := os.ReadFile(deckPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read deck file: %w", err)
	}

	var recommendation DeckRecommendation
	if err := json.Unmarshal(data, &recommendation); err != nil {
		return nil, fmt.Errorf("failed to parse deck JSON: %w", err)
	}

	return &recommendation, nil
}

// SetUnlockedEvolutions updates the unlocked evolutions list
// This allows runtime override of the UNLOCKED_EVOLUTIONS environment variable
func (b *Builder) SetUnlockedEvolutions(cards []string) {
	b.unlockedEvolutions = make(map[string]bool)
	for _, card := range cards {
		cardName := strings.TrimSpace(card)
		if cardName != "" {
			b.unlockedEvolutions[cardName] = true
		}
	}
}

// SetEvolutionSlotLimit updates the evolution slot limit
// This allows runtime override of the default 2-slot limit
func (b *Builder) SetEvolutionSlotLimit(limit int) {
	if limit > 0 {
		b.evolutionSlotLimit = limit
	}
}

// SetStrategy updates the builder's strategy and loads the corresponding configuration
func (b *Builder) SetStrategy(strategy Strategy) error {
	if err := strategy.Validate(); err != nil {
		return err
	}

	b.strategy = strategy
	b.strategyConfig = GetStrategyConfig(strategy)
	return nil
}

// SetSynergyEnabled enables or disables synergy scoring in deck building
func (b *Builder) SetSynergyEnabled(enabled bool) {
	b.synergyEnabled = enabled
}

// SetSynergyWeight sets the weight for synergy scoring (0.0 to 1.0)
// Default is 0.15 (15% of total score from synergy)
// Higher values give more importance to synergies
func (b *Builder) SetSynergyWeight(weight float64) {
	if weight < 0.0 {
		weight = 0.0
	}
	if weight > 1.0 {
		weight = 1.0
	}
	b.synergyWeight = weight
}

// SetUniquenessEnabled enables or disables uniqueness/anti-meta scoring
func (b *Builder) SetUniquenessEnabled(enabled bool) {
	b.uniquenessEnabled = enabled
}

// SetUniquenessWeight sets the weight for uniqueness scoring (0.0 to 0.3)
// Default is 0.2 (20% of total score from uniqueness)
// Higher values give more preference to less common/anti-meta cards
func (b *Builder) SetUniquenessWeight(weight float64) {
	if weight < 0.0 {
		weight = 0.0
	}
	if weight > 0.3 {
		weight = 0.3 // Cap at 0.3 to prevent over-prioritizing uniqueness
	}
	b.uniquenessWeight = weight
}

// SetIncludeCards sets the cards that must be included in the deck
// These cards will be forced into the deck if they're available in the collection
func (b *Builder) SetIncludeCards(cards []string) {
	b.includeCards = cards
}

// SetExcludeCards sets the cards that should be excluded from consideration
// These cards will never be selected for the deck
func (b *Builder) SetExcludeCards(cards []string) {
	b.excludeCards = cards
}

// SetAvoidArchetypes sets the archetypes to avoid when building decks
// Cards strongly associated with these archetypes will receive score penalties
func (b *Builder) SetAvoidArchetypes(archetypes []string) {
	b.avoidArchetypes = archetypes
	b.archetypeAvoidanceScorer = NewArchetypeAvoidanceScorer(archetypes)
}

// SetFuzzIntegration sets the fuzz integration instance for data-driven card scoring
// Pass nil to disable fuzz integration
func (b *Builder) SetFuzzIntegration(fi *FuzzIntegration) {
	b.fuzzIntegration = fi
}

// GetUpgradeRecommendations generates upgrade recommendations for a deck.
// Analyzes the card collection and identifies cards whose upgrades would have
// the most impact on deck performance.
func (b *Builder) GetUpgradeRecommendations(analysis CardAnalysis, deck *DeckRecommendation, topN int) (*UpgradeRecommendations, error) {
	if deck == nil {
		return nil, fmt.Errorf("deck recommendation is required")
	}

	if topN <= 0 {
		topN = 5 // Default to top 5 recommendations
	}

	// Collect upgrade candidates from the deck
	candidates := b.getUpgradeCandidates(deck, analysis.CardLevels)

	// Sort by impact score (descending)
	sort.Slice(candidates, func(i, j int) bool {
		// First sort by impact score
		if candidates[i].ImpactScore != candidates[j].ImpactScore {
			return candidates[i].ImpactScore > candidates[j].ImpactScore
		}
		// Then by value per gold for efficiency
		return candidates[i].ValuePerGold > candidates[j].ValuePerGold
	})

	// Limit to top N
	if len(candidates) > topN {
		candidates = candidates[:topN]
	}

	// Calculate total gold needed
	totalGold := 0
	for _, c := range candidates {
		totalGold += c.GoldCost
	}

	recommendations := &UpgradeRecommendations{
		DeckName:        strings.Join(deck.Deck[:min(3, len(deck.Deck))], ", ") + "...",
		TotalGoldNeeded: totalGold,
		Recommendations: candidates,
		GeneratedAt:     time.Now().Format(time.RFC3339),
	}

	return recommendations, nil
}

// getUpgradeCandidates analyzes deck cards and generates upgrade recommendations
func (b *Builder) getUpgradeCandidates(deck *DeckRecommendation, cardLevels map[string]CardLevelData) []UpgradeRecommendation {
	candidates := make([]UpgradeRecommendation, 0)

	for _, card := range deck.DeckDetail {
		// Skip if card is not in collection (shouldn't happen)
		_, exists := cardLevels[card.Name]
		if !exists {
			continue
		}

		// Skip if already at max level
		if card.Level >= card.MaxLevel {
			continue
		}

		// Calculate current and upgraded scores
		role := b.inferRole(card.Name)
		currentScore := ScoreCardWithStrategy(
			&CardCandidate{
				Name:              card.Name,
				Level:             card.Level,
				MaxLevel:          card.MaxLevel,
				Rarity:            card.Rarity,
				Elixir:            card.Elixir,
				Role:              role,
				EvolutionLevel:    card.EvolutionLevel,
				MaxEvolutionLevel: card.MaxEvolutionLevel,
				Stats:             b.getStatsForCard(card.Name),
			},
			role,
			b.strategyConfig,
			b.levelCurve,
		)

		targetLevel := card.Level + 1
		if targetLevel > card.MaxLevel {
			targetLevel = card.MaxLevel
		}

		upgradedScore := ScoreCardWithStrategy(
			&CardCandidate{
				Name:              card.Name,
				Level:             targetLevel,
				MaxLevel:          card.MaxLevel,
				Rarity:            card.Rarity,
				Elixir:            card.Elixir,
				Role:              role,
				EvolutionLevel:    card.EvolutionLevel,
				MaxEvolutionLevel: card.MaxEvolutionLevel,
				Stats:             b.getStatsForCard(card.Name),
			},
			role,
			b.strategyConfig,
			b.levelCurve,
		)

		scoreDelta := upgradedScore - currentScore

		// Calculate gold cost for this upgrade
		goldCost := b.getGoldCost(card.Level, card.Rarity)

		// Calculate impact score:
		// - Score delta contribution (70%)
		// - Role importance (20%)
		// - Rarity bonus (10% - rarer cards get priority)
		roleImportance := b.getRoleImportance(role)
		rarityBonus := config.GetRarityPriorityBonus(card.Rarity)

		impactScore := (scoreDelta * 1000) + // Scale up for meaningful comparison
			(roleImportance * 20) +
			(rarityBonus * 10)

		// Value per gold (impact per 1000 gold)
		valuePerGold := 0.0
		if goldCost > 0 {
			valuePerGold = impactScore / float64(goldCost) * 1000
		}

		// Generate reason for the recommendation
		reason := b.generateUpgradeReason(card, scoreDelta, role)

		candidates = append(candidates, UpgradeRecommendation{
			CardName:     card.Name,
			CurrentLevel: card.Level,
			TargetLevel:  targetLevel,
			Rarity:       card.Rarity,
			Elixir:       card.Elixir,
			Role:         getRoleString(role),
			ImpactScore:  impactScore,
			GoldCost:     goldCost,
			ValuePerGold: valuePerGold,
			Reason:       reason,
		})
	}

	return candidates
}

// getStatsForCard returns combat stats for a card if available
func (b *Builder) getStatsForCard(cardName string) *clashroyale.CombatStats {
	if b.statsRegistry != nil {
		return b.statsRegistry.GetStats(cardName)
	}
	return nil
}

// getGoldCost returns the gold needed to upgrade a card from its current level
func (b *Builder) getGoldCost(currentLevel int, rarity string) int {
	return config.GetGoldCost(currentLevel, rarity)
}

// getRoleImportance returns a score representing how important a card role is
// Uses strategy-specific role bonuses to prioritize upgrades that fit the strategy
func (b *Builder) getRoleImportance(role *CardRole) float64 {
	if role == nil {
		return 0.4
	}

	// Use strategy-specific role bonuses for upgrade prioritization
	// Strategy bonuses range from -0.5 to +0.5
	// Convert to importance score (0.0 to 1.5 range)
	baseImportance := 0.5
	if b.strategyConfig.RoleBonuses != nil {
		bonus, exists := b.strategyConfig.RoleBonuses[*role]
		if exists {
			// Convert bonus (-0.5 to +0.5) to importance (0.0 to 1.5)
			// Positive bonuses increase importance, negative bonuses decrease it
			return baseImportance + bonus*2.0
		}
	}

	// Fallback to default importance if no strategy config
	switch *role {
	case RoleWinCondition:
		return 1.0 // Most important
	case RoleBuilding:
		return 0.7
	case RoleSpellBig:
		return 0.6
	case RoleSupport:
		return 0.5
	case RoleSpellSmall:
		return 0.4
	case RoleCycle:
		return 0.3
	default:
		return 0.4
	}
}

// getRoleString converts CardRole pointer to string
func getRoleString(role *CardRole) string {
	if role == nil {
		return ""
	}
	return string(*role)
}

// generateUpgradeReason creates a human-readable reason for the upgrade recommendation
func (b *Builder) generateUpgradeReason(card CardDetail, scoreDelta float64, role *CardRole) string {
	roleStr := "card"
	if role != nil {
		roleStr = string(*role)
	}

	if scoreDelta > 0.05 {
		return fmt.Sprintf("Significant score boost (+%.3f) for this key %s", scoreDelta, roleStr)
	} else if scoreDelta > 0.02 {
		return fmt.Sprintf("Moderate score boost (+%.3f) for this %s", scoreDelta, roleStr)
	} else {
		return fmt.Sprintf("Minor improvement (+%.3f) for this %s", scoreDelta, roleStr)
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
