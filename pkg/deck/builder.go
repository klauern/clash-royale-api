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

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// Builder handles the construction of balanced Clash Royale decks
// from player card analysis data.
type Builder struct {
	dataDir            string
	unlockedEvolutions map[string]bool
	evolutionSlotLimit int
	roleGroups         map[CardRole][]string
	fallbackElixir     map[string]int
	rarityWeights      map[string]float64
	statsRegistry      *clashroyale.CardStatsRegistry
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
		roleGroups: map[CardRole][]string{
			RoleWinCondition: {
				"Royal Giant", "Hog Rider", "Giant", "P.E.K.K.A", "Giant Skeleton",
				"Goblin Barrel", "Mortar", "X-Bow", "Royal Hogs",
			},
			RoleBuilding: {
				"Cannon", "Goblin Cage", "Inferno Tower", "Bomb Tower", "Tombstone",
				"Goblin Hut", "Barbarian Hut",
			},
			RoleSpellBig: {
				"Fireball", "Poison", "Lightning", "Rocket",
			},
			RoleSpellSmall: {
				"Zap", "Arrows", "Giant Snowball", "Barbarian Barrel",
				"Freeze", "Log",
			},
			RoleSupport: {
				"Archers", "Bomber", "Musketeer", "Wizard", "Mega Minion",
				"Valkyrie", "Baby Dragon", "Skeleton Dragons",
			},
			RoleCycle: {
				"Knight", "Skeletons", "Ice Spirit", "Electro Spirit",
				"Fire Spirit", "Bats", "Spear Goblins", "Goblin Gang", "Minions",
			},
		},
		fallbackElixir: map[string]int{
			"Royal Giant": 6, "Hog Rider": 4, "Giant": 5, "P.E.K.K.A": 7,
			"Giant Skeleton": 6, "Goblin Barrel": 3, "Mortar": 4, "X-Bow": 6,
			"Royal Hogs": 5, "Cannon": 3, "Goblin Cage": 4, "Inferno Tower": 5,
			"Bomb Tower": 4, "Tombstone": 3, "Goblin Hut": 5, "Barbarian Hut": 6,
			"Fireball": 4, "Poison": 4, "Lightning": 6, "Rocket": 6,
			"Zap": 2, "Arrows": 3, "Giant Snowball": 2, "Barbarian Barrel": 2,
			"Freeze": 4, "Log": 2, "Archers": 3, "Bomber": 2,
			"Musketeer": 4, "Wizard": 5, "Mega Minion": 3, "Valkyrie": 4,
			"Baby Dragon": 4, "Skeleton Dragons": 4, "Knight": 3,
			"Skeletons": 1, "Ice Spirit": 1, "Electro Spirit": 1, "Fire Spirit": 1,
			"Bats": 2, "Spear Goblins": 2, "Goblin Gang": 3, "Minions": 3,
		},
		rarityWeights: map[string]float64{
			"Common":    1.0,
			"Rare":      1.05,
			"Epic":      1.1,
			"Legendary": 1.15,
			"Champion":  1.2,
		},
	}

	// Try to load combat stats
	statsPath := filepath.Join(dataDir, "cards_stats.json")
	if stats, err := clashroyale.LoadStats(statsPath); err == nil {
		builder.statsRegistry = stats
	} else {
		// Log error if needed, for now just silent failure or maybe print to stderr
		// fmt.Fprintf(os.Stderr, "Warning: Failed to load card stats from %s: %v\n", statsPath, err)
	}

	return builder
}

// CardAnalysis represents the input analysis data structure
type CardAnalysis struct {
	CardLevels   map[string]CardLevelData `json:"card_levels"`
	AnalysisTime string                   `json:"analysis_time,omitempty"`
}

// CardLevelData represents card level and metadata from analysis
type CardLevelData struct {
	Level             int    `json:"level"`
	MaxLevel          int    `json:"max_level"`
	Rarity            string `json:"rarity"`
	Elixir            int    `json:"elixir,omitempty"`
	EvolutionLevel    int    `json:"evolution_level,omitempty"`    // Current evolution level (0-3)
	MaxEvolutionLevel int    `json:"max_evolution_level,omitempty"` // Maximum possible evolution level
}

// BuildDeckFromAnalysis creates a deck recommendation from card analysis data
func (b *Builder) BuildDeckFromAnalysis(analysis CardAnalysis) (*DeckRecommendation, error) {
	if len(analysis.CardLevels) == 0 {
		return nil, fmt.Errorf("analysis data missing 'card_levels'")
	}

	// Convert analysis data to candidates
	candidates := b.buildCandidates(analysis.CardLevels)

	deck := make([]*CardCandidate, 0)
	used := make(map[string]bool)
	notes := make([]string, 0)

	// Core roles: win condition, building, two spells
	if winCondition := b.pickBest(RoleWinCondition, candidates, used); winCondition != nil {
		deck = append(deck, winCondition)
		used[winCondition.Name] = true
	} else {
		notes = append(notes, "No win condition found; selected highest power cards instead.")
	}

	if building := b.pickBest(RoleBuilding, candidates, used); building != nil {
		deck = append(deck, building)
		used[building.Name] = true
	}

	if bigSpell := b.pickBest(RoleSpellBig, candidates, used); bigSpell != nil {
		deck = append(deck, bigSpell)
		used[bigSpell.Name] = true
	}

	if smallSpell := b.pickBest(RoleSpellSmall, candidates, used); smallSpell != nil {
		deck = append(deck, smallSpell)
		used[smallSpell.Name] = true
	}

	// Support backbone (2 cards)
	supportCards := b.pickMany(RoleSupport, candidates, used, 2)
	deck = append(deck, supportCards...)
	for _, card := range supportCards {
		used[card.Name] = true
	}

	// Cheap cycle fillers (2 cards)
	cycleCards := b.pickMany(RoleCycle, candidates, used, 2)
	deck = append(deck, cycleCards...)
	for _, card := range cycleCards {
		used[card.Name] = true
	}

	// Fill remaining slots with highest score cards
	if len(deck) < 8 {
		remaining := b.getHighestScoreCards(candidates, used, 8-len(deck))
		deck = append(deck, remaining...)
	}

	// Ensure exactly 8 cards
	deck = deck[:8]

	// Select evolution slots based on role priority
	evolutionSlots := b.selectEvolutionSlots(deck)

	// Create recommendation
	recommendation := &DeckRecommendation{
		Deck:           make([]string, 8),
		DeckDetail:     make([]CardDetail, 8),
		AvgElixir:      b.calculateAvgElixir(deck),
		AnalysisTime:   analysis.AnalysisTime,
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
			Name:     card.Name,
			Level:    card.Level,
			MaxLevel: card.MaxLevel,
			Rarity:   card.Rarity,
			Elixir:   card.Elixir,
			Role:     roleStr,
			Score:    roundToThree(card.Score),
		}
	}

	// Add strategic notes
	b.addStrategicNotes(recommendation)

	// Add evolution slot note if applicable
	if len(recommendation.EvolutionSlots) > 0 {
		slotNote := fmt.Sprintf("Evolution slots: %s", strings.Join(recommendation.EvolutionSlots, ", "))
		recommendation.AddNote(slotNote)
	}

	return recommendation, nil
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
func (b *Builder) LoadLatestAnalysis(playerTag string, analysisDir string) (*CardAnalysis, error) {
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
func (b *Builder) SaveDeck(deckData *DeckRecommendation, outputDir string, playerTag string) (string, error) {
	if outputDir == "" {
		outputDir = filepath.Join(b.dataDir, "decks")
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
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

	if err := os.WriteFile(path, data, 0644); err != nil {
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

	// Use combat-enhanced scoring if stats are available
	var score float64
	if stats != nil {
		// Calculate custom evolution bonus and add to combat-enhanced score
		evolutionBonus := b.calculateEvolutionBonus(name, level, maxLevel, data.MaxEvolutionLevel)
		combatScore := ScoreCardWithCombat(level, maxLevel, rarity, elixir, role, stats)
		score = combatScore + evolutionBonus
	} else {
		// Fall back to traditional scoring if no combat stats available
		score = b.scoreCard(name, level, maxLevel, rarity, elixir, role, data.MaxEvolutionLevel)
	}

	return &CardCandidate{
		Name:              name,
		Level:             level,
		MaxLevel:          maxLevel,
		Rarity:            rarity,
		Elixir:            elixir,
		Role:              role,
		Score:             score,
		HasEvolution:      hasEvolution,
		EvolutionPriority: evoPriority,
		EvolutionLevel:    data.EvolutionLevel,
		MaxEvolutionLevel: data.MaxEvolutionLevel,
		Stats:             stats,
	}
}

func (b *Builder) resolveElixir(name string, data CardLevelData) int {
	if data.Elixir > 0 {
		return data.Elixir
	}
	if elixir, exists := b.fallbackElixir[name]; exists {
		return elixir
	}
	return 4 // Default fallback
}

func (b *Builder) inferRole(name string) *CardRole {
	for role, names := range b.roleGroups {
		for _, cardName := range names {
			if cardName == name {
				return &role
			}
		}
	}
	return nil
}

func (b *Builder) scoreCard(name string, level, maxLevel int, rarity string, elixir int, role *CardRole, maxEvolutionLevel int) float64 {
	levelRatio := float64(level) / float64(maxLevel)
	rarityBoost := b.rarityWeights[rarity]
	if rarityBoost == 0 {
		rarityBoost = 1.0
	}

	// Encourage cheaper cards slightly to keep cycle tight
	elixirWeight := 1.0 - float64(max(elixir-3, 0))/9.0

	roleBonus := 0.05
	if role == nil {
		roleBonus = 0
	}

	// Use level-scaled evolution bonus
	evolutionBonus := b.calculateEvolutionBonus(name, level, maxLevel, maxEvolutionLevel)

	return (levelRatio * 1.2 * rarityBoost) + (elixirWeight * 0.15) + roleBonus + evolutionBonus
}

// calculateEvolutionBonus returns level-scaled evolution bonus
// Formula: baseBonus * (level/maxLevel)^1.5 * (1 + 0.2*(maxEvoLevel-1))
// This rewards using higher-level cards and accounts for multi-evolution cards
func (b *Builder) calculateEvolutionBonus(cardName string, level, maxLevel, maxEvoLevel int) float64 {
	// Check if evolution is unlocked
	if !b.unlockedEvolutions[cardName] || maxEvoLevel == 0 {
		return 0.0
	}

	const baseBonus = 0.25
	levelRatio := float64(level) / float64(maxLevel)
	scaledRatio := math.Pow(levelRatio, 1.5)

	// Bonus multiplier for multi-evolution cards (e.g., Knight with evo level 3)
	evoMultiplier := 1.0 + (0.2 * float64(maxEvoLevel-1))

	return baseBonus * scaledRatio * evoMultiplier
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

func (b *Builder) pickBest(role CardRole, candidates []*CardCandidate, used map[string]bool) *CardCandidate {
	roleCards, exists := b.roleGroups[role]
	if !exists {
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

	// Return highest scoring card
	sort.Slice(pool, func(i, j int) bool {
		return pool[i].Score > pool[j].Score
	})

	return pool[0]
}

func (b *Builder) pickMany(role CardRole, candidates []*CardCandidate, used map[string]bool, count int) []*CardCandidate {
	roleCards, exists := b.roleGroups[role]
	if !exists {
		return nil
	}

	var pool []*CardCandidate
	for _, candidate := range candidates {
		if !used[candidate.Name] && b.contains(roleCards, candidate.Name) {
			pool = append(pool, candidate)
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

func (b *Builder) getHighestScoreCards(candidates []*CardCandidate, used map[string]bool, count int) []*CardCandidate {
	var pool []*CardCandidate
	for _, candidate := range candidates {
		if !used[candidate.Name] {
			pool = append(pool, candidate)
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
	if len(deck) == 0 {
		return 0
	}

	total := 0
	for _, card := range deck {
		total += card.Elixir
	}

	return roundToTwo(float64(total) / float64(len(deck)))
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
