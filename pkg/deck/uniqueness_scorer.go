// Package deck provides uniqueness scoring for anti-meta deck building.
//
// The uniqueness scoring system rewards decks that use less common cards,
// encouraging variety and potentially catching opponents off-guard with
// unexpected card choices.
//
// This is particularly useful for:
//   - Finding anti-meta decks that opponents are less prepared for
//   - Encouraging deck variety in generated recommendations
//   - Discovering hidden gems in your card collection
package deck

import (
	"maps"
	"math"
	"sort"
	"sync"
)

// CardPopularity tracks how frequently cards appear in decks.
// This can be populated from various sources:
//   - API data ( Royale API, etc.)
//   - Fuzzing results
//   - Static estimates based on card rarity/age
//   - User's own battle history
//
// Popularity scores are 0.0-1.0 where:
//   - 1.0 = very common (e.g., Zap, Hog Rider, Mini P.E.K.K.A)
//   - 0.5 = moderately used
//   - 0.0 = very rare (e.g., Rage, Clone, Heal)
type CardPopularity struct {
	mu         sync.RWMutex
	popularity map[string]float64 // card name -> popularity score (0.0-1.0)
	source     PopularitySource
}

// PopularitySource indicates where popularity data came from
type PopularitySource int

const (
	// PopularitySourceStatic uses built-in estimates based on card characteristics
	PopularitySourceStatic PopularitySource = iota
	// PopularitySourceFuzzing derives popularity from local fuzzing results
	PopularitySourceFuzzing
	// PopularitySourceAPIDerived uses external API data
	PopularitySourceAPIDerived
	// PopularitySourceUserHistory uses user's own battle history
	PopularitySourceUserHistory
)

// DefaultCardPopularity contains estimated popularity scores for common cards.
// These are rough estimates based on card usage patterns in the meta.
// Higher values = more commonly seen cards.
var DefaultCardPopularity = map[string]float64{
	// Very common cards (0.8-1.0)
	"Zap":            0.95,
	"The Log":        0.92,
	"Fireball":       0.88,
	"Mini P.E.K.K.A": 0.85,
	"Hog Rider":      0.90,
	"Musketeer":      0.82,
	"Valkyrie":       0.80,
	"Skeletons":      0.85,
	"Ice Spirit":     0.78,
	"Knight":         0.82,
	"Archers":        0.75,
	"Goblin Barrel":  0.88,
	"Inferno Tower":  0.80,
	"Tesla":          0.78,
	"Baby Dragon":    0.82,
	"Mega Minion":    0.75,
	"Golem":          0.72,
	"P.E.K.K.A":      0.70,
	"Electro Wizard": 0.78,
	"Wizard":         0.72,

	// Common cards (0.5-0.8)
	"Royal Giant":      0.68,
	"Giant":            0.65,
	"Balloon":          0.62,
	"Miner":            0.70,
	"Princess":         0.68,
	"Ice Wizard":       0.65,
	"Lumberjack":       0.60,
	"Ram Rider":        0.58,
	"Battle Ram":       0.62,
	"Mortar":           0.55,
	"X-Bow":            0.52,
	"Graveyard":        0.58,
	"Lava Hound":       0.55,
	"Sparky":           0.50,
	"Night Witch":      0.62,
	"Bandit":           0.60,
	"Royal Ghost":      0.55,
	"Fisherman":        0.58,
	"Hunter":           0.52,
	"Flying Machine":   0.48,
	"Zappies":          0.45,
	"Elixir Golem":     0.42,
	"Battle Healer":    0.40,
	"Elixir Golemites": 0.42,
	"Electro Dragon":   0.58,
	"Inferno Dragon":   0.62,
	"Lava Pups":        0.55,
	"Bats":             0.72,
	"Goblins":          0.70,
	"Spear Goblins":    0.68,
	"Minions":          0.65,
	"Minion Horde":     0.55,
	"Skeleton Barrel":  0.52,
	"Goblin Gang":      0.68,
	"Rascals":          0.48,
	"Royal Recruits":   0.45,
	"Barbarians":       0.58,
	"Three Musketeers": 0.42,
	"Dark Prince":      0.55,
	"Prince":           0.52,
	"Executioner":      0.48,
	"Witch":            0.50,
	"Bowler":           0.45,
	"Cannon Cart":      0.42,
	"Mega Knight":      0.72,
	"Ice Golem":        0.65,
	"Giant Snowball":   0.60,
	"Poison":           0.68,
	"Lightning":        0.62,
	"Rocket":           0.58,
	"Arrows":           0.75,
	"Tornado":          0.65,
	"Freeze":           0.48,
	"Mirror":           0.35,
	"Clone":            0.25,
	"Rage":             0.30,
	"Heal Spirit":      0.55,
	"Electro Spirit":   0.62,
	"Fire Spirit":      0.58,
	"Bomber":           0.52,
	"Dart Goblin":      0.48,
	"Goblin Cage":      0.55,
	"Tombstone":        0.58,
	"Cannon":           0.62,
	"Bomb Tower":       0.48,
	"Goblin Hut":       0.42,
	"Barbarian Hut":    0.38,
	"Furnace":          0.52,
	"Elixir Collector": 0.35,
	"Earthquake":       0.45,
	"Barbarian Barrel": 0.72,
	"Royal Delivery":   0.48,

	// Less common cards (0.2-0.5)
	"Wall Breakers":     0.42,
	"Skeleton Dragons":  0.48,
	"Mother Witch":      0.45,
	"Electro Giant":     0.52,
	"Goblin Drill":      0.58,
	"Skeleton King":     0.55,
	"Archer Queen":      0.62,
	"Golden Knight":     0.58,
	"Mighty Miner":      0.55,
	"Monk":              0.48,
	"Phoenix":           0.52,
	"Little Prince":     0.45,
	"Necromancer":       0.38,
	"Void":              0.35,
	"Goblin Curse":      0.32,
	"Suspicious Bush":   0.28,
	"Ogre":              0.25,
	"Goblin Demolisher": 0.22,
	"Goblin Machine":    0.20,
	"Goblinstein":       0.18,
	"Tower Princess":    0.55,
	"Cannoneer":         0.52,
	"Dagger Duchess":    0.48,

	// Very rare cards (0.0-0.2)
	"Heal":     0.15,
	"Guardian": 0.12,
}

// NewCardPopularity creates a new popularity tracker with default values
func NewCardPopularity() *CardPopularity {
	cp := &CardPopularity{
		popularity: make(map[string]float64),
		source:     PopularitySourceStatic,
	}

	// Copy default values
	maps.Copy(cp.popularity, DefaultCardPopularity)

	return cp
}

// GetPopularity returns the popularity score for a card (0.0-1.0)
// Returns 0.5 (average) for unknown cards
func (cp *CardPopularity) GetPopularity(cardName string) float64 {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	if popularity, exists := cp.popularity[cardName]; exists {
		return popularity
	}

	// Return average popularity for unknown cards
	return 0.5
}

// SetPopularity sets the popularity for a card
func (cp *CardPopularity) SetPopularity(cardName string, popularity float64) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Clamp to valid range
	if popularity < 0 {
		popularity = 0
	}
	if popularity > 1 {
		popularity = 1
	}

	cp.popularity[cardName] = popularity
}

// UpdateFromFuzzing updates popularity based on fuzzing results
// Cards that appear frequently in high-scoring fuzzed decks are considered popular
func (cp *CardPopularity) UpdateFromFuzzing(decks [][]string, topDeckThreshold float64) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cardCounts := make(map[string]int)
	totalDecks := len(decks)

	if totalDecks == 0 {
		return
	}

	// Count card appearances
	for _, deck := range decks {
		for _, card := range deck {
			cardCounts[card]++
		}
	}

	// Update popularity based on frequency
	for card, count := range cardCounts {
		frequency := float64(count) / float64(totalDecks)
		// Normalize: 8 cards per deck, so max frequency is 1.0 (appears in every deck)
		// Scale to 0.0-1.0 popularity
		cp.popularity[card] = math.Min(frequency*2.0, 1.0) // Multiply by 2 to spread out the range
	}

	cp.source = PopularitySourceFuzzing
}

// GetUniquenessScore returns a uniqueness score (0.0-1.0) where higher = more unique
// This is the inverse of popularity: uniqueness = 1.0 - popularity
func (cp *CardPopularity) GetUniquenessScore(cardName string) float64 {
	popularity := cp.GetPopularity(cardName)
	return 1.0 - popularity
}

// UniquenessScorer calculates uniqueness bonuses for decks
type UniquenessScorer struct {
	popularity *CardPopularity
	config     UniquenessConfig
}

// UniquenessConfig controls how uniqueness scoring behaves
type UniquenessConfig struct {
	// Enabled turns uniqueness scoring on/off
	Enabled bool

	// Weight is how much uniqueness affects overall score (0.0-0.3 recommended)
	// 0.0 = no uniqueness bonus
	// 0.1 = slight preference for unique cards
	// 0.2 = moderate preference
	// 0.3 = strong preference for anti-meta decks
	Weight float64

	// MinUniquenessThreshold only gives bonuses to cards below this popularity
	// e.g., 0.7 means cards with popularity < 0.7 (uniqueness > 0.3) get bonuses
	MinUniquenessThreshold float64

	// UseGeometricMean uses geometric mean instead of arithmetic mean
	// This rewards decks where ALL cards are unique, not just a few
	UseGeometricMean bool
}

// DefaultUniquenessConfig returns a balanced configuration
func DefaultUniquenessConfig() UniquenessConfig {
	return UniquenessConfig{
		Enabled:                false, // Off by default
		Weight:                 0.15,
		MinUniquenessThreshold: 0.5,   // Only cards with popularity < 0.5 get bonuses
		UseGeometricMean:       false, // Use arithmetic mean by default
	}
}

// NewUniquenessScorer creates a new scorer with default popularity data
func NewUniquenessScorer(config UniquenessConfig) *UniquenessScorer {
	return &UniquenessScorer{
		popularity: NewCardPopularity(),
		config:     config,
	}
}

// NewUniquenessScorerWithPopularity creates a scorer with custom popularity data
func NewUniquenessScorerWithPopularity(popularity *CardPopularity, config UniquenessConfig) *UniquenessScorer {
	return &UniquenessScorer{
		popularity: popularity,
		config:     config,
	}
}

// ScoreDeck calculates the uniqueness score for a deck
// Returns 0.0-1.0 where higher means more unique/anti-meta
func (us *UniquenessScorer) ScoreDeck(cardNames []string) float64 {
	if !us.config.Enabled || len(cardNames) == 0 {
		return 0.0
	}

	uniquenessScores := make([]float64, 0, len(cardNames))

	for _, card := range cardNames {
		uniqueness := us.popularity.GetUniquenessScore(card)

		// Apply threshold - only count cards that meet the uniqueness threshold
		if uniqueness >= us.config.MinUniquenessThreshold {
			uniquenessScores = append(uniquenessScores, uniqueness)
		} else {
			// Card doesn't meet threshold, give it a 0 for this component
			uniquenessScores = append(uniquenessScores, 0)
		}
	}

	if len(uniquenessScores) == 0 {
		return 0.0
	}

	if us.config.UseGeometricMean {
		return calculateGeometricMean(uniquenessScores)
	}

	// Arithmetic mean
	total := 0.0
	for _, score := range uniquenessScores {
		total += score
	}
	return total / float64(len(uniquenessScores))
}

// ScoreDeckWithDetails calculates uniqueness score with detailed breakdown
func (us *UniquenessScorer) ScoreDeckWithDetails(cardNames []string) UniquenessResult {
	result := UniquenessResult{
		CardUniqueness: make(map[string]float64),
	}

	if !us.config.Enabled || len(cardNames) == 0 {
		return result
	}

	scores, totalUniqueness := us.calculateCardScores(cardNames, &result)
	us.updateMinMaxCards(cardNames, &result)

	result.AverageUniqueness = totalUniqueness / float64(len(cardNames))
	result.FinalScore = us.calculateFinalScore(scores, totalUniqueness)
	result.WeightedScore = result.FinalScore * us.config.Weight

	return result
}

// calculateCardScores processes each card and returns scores and total uniqueness
func (us *UniquenessScorer) calculateCardScores(cardNames []string, result *UniquenessResult) ([]float64, float64) {
	scores := make([]float64, 0, len(cardNames))
	var totalUniqueness float64

	for _, card := range cardNames {
		uniqueness := us.popularity.GetUniquenessScore(card)
		result.CardUniqueness[card] = uniqueness

		if uniqueness >= us.config.MinUniquenessThreshold {
			scores = append(scores, uniqueness)
			totalUniqueness += uniqueness
			result.QualifyingCards++
		} else {
			scores = append(scores, 0)
			result.BelowThresholdCards = append(result.BelowThresholdCards, card)
		}
	}

	return scores, totalUniqueness
}

// updateMinMaxCards tracks the most and least unique cards
func (us *UniquenessScorer) updateMinMaxCards(cardNames []string, result *UniquenessResult) {
	for _, card := range cardNames {
		uniqueness := result.CardUniqueness[card]
		popularity := us.popularity.GetPopularity(card)

		if result.MostUniqueCard == "" || uniqueness > result.CardUniqueness[result.MostUniqueCard] {
			result.MostUniqueCard = card
		}
		if result.LeastUniqueCard == "" || popularity > us.popularity.GetPopularity(result.LeastUniqueCard) {
			result.LeastUniqueCard = card
		}
	}
}

// calculateFinalScore computes the final score based on configuration
func (us *UniquenessScorer) calculateFinalScore(scores []float64, totalUniqueness float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}

	if us.config.UseGeometricMean {
		return calculateGeometricMean(scores)
	}

	return totalUniqueness / float64(len(scores))
}

// UniquenessResult contains detailed uniqueness scoring information
type UniquenessResult struct {
	FinalScore          float64
	WeightedScore       float64
	AverageUniqueness   float64
	CardUniqueness      map[string]float64
	QualifyingCards     int
	BelowThresholdCards []string
	MostUniqueCard      string
	LeastUniqueCard     string
}

// GetPopularity returns the underlying popularity tracker (for advanced use)
func (us *UniquenessScorer) GetPopularity() *CardPopularity {
	return us.popularity
}

// UpdateConfig updates the scorer configuration
func (us *UniquenessScorer) UpdateConfig(config UniquenessConfig) {
	us.config = config
}

// GetConfig returns the current configuration
func (us *UniquenessScorer) GetConfig() UniquenessConfig {
	return us.config
}

// calculateGeometricMean calculates the geometric mean of a slice
// This rewards consistent uniqueness across all cards
func calculateGeometricMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	// Use log-space to avoid underflow
	sumLogs := 0.0
	for _, v := range values {
		if v <= 0 {
			return 0.0 // Any zero makes geometric mean zero
		}
		sumLogs += math.Log(v)
	}

	return math.Exp(sumLogs / float64(len(values)))
}

// GetMostCommonCards returns the N most common cards
func (cp *CardPopularity) GetMostCommonCards(n int) []CardPopularityEntry {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	entries := make([]CardPopularityEntry, 0, len(cp.popularity))
	for card, popularity := range cp.popularity {
		entries = append(entries, CardPopularityEntry{
			CardName:   card,
			Popularity: popularity,
		})
	}

	// Sort by popularity descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Popularity > entries[j].Popularity
	})

	if n > len(entries) {
		n = len(entries)
	}

	return entries[:n]
}

// GetMostUniqueCards returns the N most unique (least common) cards
func (cp *CardPopularity) GetMostUniqueCards(n int) []CardPopularityEntry {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	entries := make([]CardPopularityEntry, 0, len(cp.popularity))
	for card, popularity := range cp.popularity {
		entries = append(entries, CardPopularityEntry{
			CardName:   card,
			Popularity: popularity,
			Uniqueness: 1.0 - popularity,
		})
	}

	// Sort by popularity ascending (most unique first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Popularity < entries[j].Popularity
	})

	if n > len(entries) {
		n = len(entries)
	}

	return entries[:n]
}

// CardPopularityEntry represents a card's popularity data
type CardPopularityEntry struct {
	CardName   string
	Popularity float64
	Uniqueness float64
}

// GetCardUniquenessTier returns a qualitative tier for a card's uniqueness
func GetCardUniquenessTier(uniqueness float64) string {
	switch {
	case uniqueness >= 0.8:
		return "Very Unique"
	case uniqueness >= 0.6:
		return "Unique"
	case uniqueness >= 0.4:
		return "Moderate"
	case uniqueness >= 0.2:
		return "Common"
	default:
		return "Very Common"
	}
}

// GlobalUniquenessScorer is a package-level scorer for convenience
var GlobalUniquenessScorer = NewUniquenessScorer(DefaultUniquenessConfig())

// EnableUniquenessScoring enables the global uniqueness scorer
func EnableUniquenessScoring(weight float64) {
	config := DefaultUniquenessConfig()
	config.Enabled = true
	config.Weight = weight
	GlobalUniquenessScorer.UpdateConfig(config)
}

// DisableUniquenessScoring disables the global uniqueness scorer
func DisableUniquenessScoring() {
	config := DefaultUniquenessConfig()
	config.Enabled = false
	GlobalUniquenessScorer.UpdateConfig(config)
}
