// Package deck provides integration between deck fuzzing results and intelligent deck building.
package deck

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/fuzzstorage"
)

const (
	// DefaultFuzzScoringWeight is the default weight for fuzz boost in scoring (10%)
	DefaultFuzzScoringWeight = 0.10
	// DefaultFuzzTopPercentile is the default percentile of top decks to analyze (10%)
	DefaultFuzzTopPercentile = 0.10
	// DefaultFuzzMinBoost is the minimum boost a card can receive from fuzz stats
	DefaultFuzzMinBoost = 1.0
	// DefaultFuzzMaxBoost is the maximum boost a card can receive from fuzz stats
	DefaultFuzzMaxBoost = 1.5
)

// FuzzCardStats represents aggregated statistics for a single card from fuzzing results
type FuzzCardStats struct {
	// CardName is the name of the card
	CardName string
	// Frequency is how often this card appears in top fuzz decks
	Frequency int
	// AvgScore is the average overall score of decks containing this card
	AvgScore float64
	// MaxScore is the highest score achieved by a deck containing this card
	MaxScore float64
	// Boost is the calculated boost multiplier for this card (1.0-2.0)
	Boost float64
}

// FuzzIntegration provides methods to analyze fuzz results and apply boosts to card scoring
type FuzzIntegration struct {
	mu            sync.RWMutex
	stats         map[string]*FuzzCardStats // card name -> stats
	topPercentile float64                   // percentile of top decks to analyze
	weight        float64                   // weight for fuzz boost in final score
	minBoost      float64                   // minimum boost multiplier
	maxBoost      float64                   // maximum boost multiplier
}

// NewFuzzIntegration creates a new FuzzIntegration instance with default settings
func NewFuzzIntegration() *FuzzIntegration {
	return &FuzzIntegration{
		stats:         make(map[string]*FuzzCardStats),
		topPercentile: DefaultFuzzTopPercentile,
		weight:        DefaultFuzzScoringWeight,
		minBoost:      DefaultFuzzMinBoost,
		maxBoost:      DefaultFuzzMaxBoost,
	}
}

// SetWeight sets the weight for fuzz boost in scoring (0.0 to 1.0)
func (fi *FuzzIntegration) SetWeight(weight float64) {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	if weight < 0.0 {
		weight = 0.0
	}
	if weight > 1.0 {
		weight = 1.0
	}
	fi.weight = weight
}

// SetTopPercentile sets the percentile of top decks to analyze (0.0 to 1.0)
func (fi *FuzzIntegration) SetTopPercentile(percentile float64) {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	if percentile < 0.01 {
		percentile = 0.01
	}
	if percentile > 1.0 {
		percentile = 1.0
	}
	fi.topPercentile = percentile
}

// SetBoostRange sets the min/max boost multipliers
func (fi *FuzzIntegration) SetBoostRange(min, max float64) {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	fi.minBoost = min
	fi.maxBoost = max
}

// AnalyzeFromStorage analyzes stored fuzz results and calculates card statistics
// It queries the storage for top decks and aggregates statistics for each card
func (fi *FuzzIntegration) AnalyzeFromStorage(storage *fuzzstorage.Storage, limit int) error {
	if storage == nil {
		return fmt.Errorf("storage cannot be nil")
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	fi.stats = make(map[string]*FuzzCardStats)

	topDecks, err := storage.GetTopN(limit)
	if err != nil {
		return fmt.Errorf("failed to query top decks: %w", err)
	}

	if len(topDecks) == 0 {
		return nil
	}

	topPercentileDecks := fi.getTopPercentileDecks(topDecks)
	fi.aggregateCardStats(topPercentileDecks)
	fi.calculateBoosts(len(topPercentileDecks))

	return nil
}

// getTopPercentileDecks returns the top percentile of decks based on configured percentile
func (fi *FuzzIntegration) getTopPercentileDecks(topDecks []fuzzstorage.DeckEntry) []fuzzstorage.DeckEntry {
	cutoffIdx := max(int(float64(len(topDecks))*fi.topPercentile), 1)
	return topDecks[:cutoffIdx]
}

// isValidCardName checks if a card name is valid (not a placeholder or test card)
// Returns false for placeholder names like "Card1", "Card2", "Card3", etc.
func isValidCardName(cardName string) bool {
	trimmed := strings.TrimSpace(cardName)
	if trimmed == "" {
		return false
	}

	// Skip test/placeholder card names
	// Common patterns: Card1, Card2, Card3, CardN, card1, card2, etc.
	if after, ok := strings.CutPrefix(strings.ToLower(trimmed), "card"); ok {
		// Check if it ends with a number (pattern like "Card1", "Card123")
		rest := after
		if rest == "" || isNumeric(rest) {
			return false
		}
	}

	// Check if it's a known card by trying to get its elixir cost
	// If it has a valid elixir cost (>= 0), it's a real card
	elixir := config.GetCardElixir(trimmed, -1)
	return elixir >= 0
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// aggregateCardStats aggregates statistics for each card across all provided decks
// Skips invalid/test card names automatically
func (fi *FuzzIntegration) aggregateCardStats(decks []fuzzstorage.DeckEntry) {
	for _, deck := range decks {
		for _, cardName := range deck.Cards {
			normalizedCard := strings.TrimSpace(cardName)
			if normalizedCard == "" {
				continue
			}

			// Skip invalid/test card names
			if !isValidCardName(normalizedCard) {
				continue
			}

			if _, exists := fi.stats[normalizedCard]; !exists {
				fi.stats[normalizedCard] = &FuzzCardStats{
					CardName: normalizedCard,
				}
			}

			stats := fi.stats[normalizedCard]
			stats.Frequency++
			stats.AvgScore += deck.OverallScore
			if deck.OverallScore > stats.MaxScore {
				stats.MaxScore = deck.OverallScore
			}
		}
	}
}

// calculateBoosts calculates the final boost multiplier for all cards
func (fi *FuzzIntegration) calculateBoosts(totalDecks int) {
	for _, stats := range fi.stats {
		if stats.Frequency > 0 {
			stats.AvgScore /= float64(stats.Frequency)
		}
		stats.Boost = fi.calculateBoost(stats.Frequency, stats.AvgScore, totalDecks)
	}
}

// calculateBoost computes a boost multiplier based on card frequency and average score
func (fi *FuzzIntegration) calculateBoost(frequency int, avgScore float64, totalDecks int) float64 {
	// Frequency factor: cards appearing in more top decks get higher boost
	// Normalize to 0-1 range based on appearance in top decks
	frequencyFactor := float64(frequency) / float64(totalDecks)

	// Score factor: cards in higher-scoring decks get higher boost
	// Normalize avgScore assuming max reasonable score is 10.0
	scoreFactor := avgScore / 10.0
	if scoreFactor > 1.0 {
		scoreFactor = 1.0
	}

	// Combined factor (60% frequency, 40% score)
	combinedFactor := (frequencyFactor*0.6 + scoreFactor*0.4)

	// Map to boost range
	boost := fi.minBoost + combinedFactor*(fi.maxBoost-fi.minBoost)

	return boost
}

// GetFuzzBoost returns the fuzz boost multiplier for a specific card
// Returns 1.0 (no boost) if the card has no fuzz statistics
func (fi *FuzzIntegration) GetFuzzBoost(cardName string) float64 {
	fi.mu.RLock()
	defer fi.mu.RUnlock()

	normalizedCard := strings.TrimSpace(cardName)
	if stats, exists := fi.stats[normalizedCard]; exists {
		return stats.Boost
	}
	return 1.0
}

// GetFuzzStats returns the FuzzCardStats for a specific card
// Returns nil if the card has no fuzz statistics
func (fi *FuzzIntegration) GetFuzzStats(cardName string) *FuzzCardStats {
	fi.mu.RLock()
	defer fi.mu.RUnlock()

	normalizedCard := strings.TrimSpace(cardName)
	if stats, exists := fi.stats[normalizedCard]; exists {
		// Return a copy to prevent external modification
		copy := *stats
		return &copy
	}
	return nil
}

// GetAllStats returns a copy of all card statistics
func (fi *FuzzIntegration) GetAllStats() map[string]*FuzzCardStats {
	fi.mu.RLock()
	defer fi.mu.RUnlock()

	result := make(map[string]*FuzzCardStats, len(fi.stats))
	for card, stats := range fi.stats {
		copy := *stats
		result[card] = &copy
	}
	return result
}

// GetTopCards returns the top N cards by fuzz boost, sorted by boost descending
func (fi *FuzzIntegration) GetTopCards(n int) []*FuzzCardStats {
	fi.mu.RLock()
	defer fi.mu.RUnlock()

	// Collect all stats into a slice
	allStats := make([]*FuzzCardStats, 0, len(fi.stats))
	for _, stats := range fi.stats {
		copy := *stats
		allStats = append(allStats, &copy)
	}

	// Sort by boost descending
	sort.Slice(allStats, func(i, j int) bool {
		return allStats[i].Boost > allStats[j].Boost
	})

	// Return top N
	if n > 0 && n < len(allStats) {
		return allStats[:n]
	}
	return allStats
}

// GetWeight returns the current fuzz scoring weight
func (fi *FuzzIntegration) GetWeight() float64 {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	return fi.weight
}

// HasStats returns true if fuzz statistics are available
func (fi *FuzzIntegration) HasStats() bool {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	return len(fi.stats) > 0
}

// StatsCount returns the number of cards with fuzz statistics
func (fi *FuzzIntegration) StatsCount() int {
	fi.mu.RLock()
	defer fi.mu.RUnlock()
	return len(fi.stats)
}

// Clear removes all fuzz statistics
func (fi *FuzzIntegration) Clear() {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	fi.stats = make(map[string]*FuzzCardStats)
}

// ApplyFuzzBoost applies the fuzz boost to a base score, returning the adjusted score
// The fuzz boost contributes according to the configured weight:
// finalScore = baseScore * (1 + (boost - 1) * weight)
// For example, with weight=0.10 and boost=1.5: score *= 1.05
func (fi *FuzzIntegration) ApplyFuzzBoost(baseScore float64, cardName string) float64 {
	boost := fi.GetFuzzBoost(cardName)
	fi.mu.RLock()
	defer fi.mu.RUnlock()

	// Apply weighted boost: the boost effect is scaled by the weight
	// boost=1.5 means 50% increase, with weight=0.10 means 5% actual increase
	boostEffect := (boost - 1.0) * fi.weight
	return baseScore * (1.0 + boostEffect)
}
