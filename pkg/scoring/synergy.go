// Package scoring provides implementations of the Scorer interface for
// various card scoring algorithms.
package scoring

import (
	"sync"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// SynergyScorer implements the Scorer interface for card synergy scoring.
// It evaluates how well a card synergizes with cards already selected
// for the deck. The synergy bonus is the average of all pairwise synergy
// scores between the candidate and cards currently in the deck.
//
// # Scoring Formula
//
//	synergyScore = Σ(synergy(candidate, deckCard)) / count
//
// Where:
//   - synergy(candidate, deckCard) is the pairwise synergy score (0.0 to 1.0)
//   - count is the number of cards in the current deck with synergies
//
// The final score returned is the synergy bonus, typically ranging from
// 0.0 (no synergies) to ~1.0 (perfect synergies with all deck cards).
//
// # Thread Safety
//
// SynergyScorer is thread-safe for concurrent use. The underlying
// SynergyDatabase is not modified during scoring operations.
type SynergyScorer struct {
	// synergyDB provides access to card pair synergy data.
	// If nil, scorer returns 0 for all cards.
	synergyDB SynergyDatabase

	// synergyWeight determines how much synergy affects the final score.
	// Default 0.15 (15% of total score from synergy).
	synergyWeight float64

	// mu protects concurrent access to synergyDB
	mu sync.RWMutex
}

// SynergyScorerConfig configures a SynergyScorer with custom parameters.
type SynergyScorerConfig struct {
	// SynergyDatabase provides access to card pair synergy data.
	SynergyDatabase SynergyDatabase

	// SynergyWeight determines how much synergy affects the score (default 0.15).
	SynergyWeight float64
}

// NewSynergyScorer creates a new SynergyScorer with default parameters.
//
// Example:
//
//	scorer := NewSynergyScorer(synergyDB)
//	score := scorer.Score(candidate, config)
func NewSynergyScorer(synergyDB SynergyDatabase) *SynergyScorer {
	return &SynergyScorer{
		synergyDB:     synergyDB,
		synergyWeight: 0.15,
	}
}

// NewSynergyScorerWithConfig creates a new SynergyScorer with custom configuration.
//
// Example:
//
//	config := SynergyScorerConfig{
//	    SynergyDatabase: synergyDB,
//	    SynergyWeight: 0.25,  // Higher synergy emphasis
//	}
//	scorer := NewSynergyScorerWithConfig(config)
func NewSynergyScorerWithConfig(config SynergyScorerConfig) *SynergyScorer {
	synergyWeight := config.SynergyWeight
	if synergyWeight == 0 {
		synergyWeight = 0.15
	}

	return &SynergyScorer{
		synergyDB:     config.SynergyDatabase,
		synergyWeight: synergyWeight,
	}
}

// Score calculates the synergy bonus for a card candidate.
//
// The bonus is the average of all pairwise synergy scores between the
// candidate and cards currently in the deck (from config.CurrentDeck).
//
// Returns 0 if:
//   - No synergy database is available
//   - Current deck is empty
//   - No synergies exist with current deck cards
//
// Returns a score in the range 0.0 to ~1.0, higher is better.
func (s *SynergyScorer) Score(candidate CardCandidate, config ScoringConfig) float64 {
	// Use config's synergy database if provided, otherwise use scorer's default
	synergyDB := config.SynergyDatabase
	if synergyDB == nil {
		s.mu.RLock()
		synergyDB = s.synergyDB
		s.mu.RUnlock()
	}

	if synergyDB == nil {
		return 0.0 // No synergy database available
	}

	// Get current deck from config
	currentDeck := config.CurrentDeck
	if len(currentDeck) == 0 {
		return 0.0 // No cards in deck yet, no synergy bonus
	}

	// Calculate average synergy with current deck
	totalSynergy := 0.0
	synergyCount := 0

	for _, deckCard := range currentDeck {
		if synergyScore := synergyDB.GetSynergy(candidate.Name, deckCard.Name); synergyScore > 0 {
			totalSynergy += synergyScore
			synergyCount++
		}
	}

	// Return average synergy (0.0 if no synergies found)
	if synergyCount == 0 {
		return 0.0
	}

	return (totalSynergy / float64(synergyCount)) * s.synergyWeight
}

// SetSynergyDatabase updates the synergy database.
// This method is thread-safe.
func (s *SynergyScorer) SetSynergyDatabase(db SynergyDatabase) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.synergyDB = db
}

// GetSynergyDatabase returns the current synergy database.
// This method is thread-safe.
func (s *SynergyScorer) GetSynergyDatabase() SynergyDatabase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.synergyDB
}

// SetSynergyWeight updates the synergy weight.
func (s *SynergyScorer) SetSynergyWeight(weight float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.synergyWeight = weight
}

// GetSynergyWeight returns the current synergy weight.
func (s *SynergyScorer) GetSynergyWeight() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.synergyWeight
}

// CalculateSynergyScore calculates the synergy bonus for a card with the current deck.
// This is a convenience method for external use.
func (s *SynergyScorer) CalculateSynergyScore(cardName string, currentDeck []CardCandidate) float64 {
	if len(currentDeck) == 0 {
		return 0.0
	}

	s.mu.RLock()
	synergyDB := s.synergyDB
	s.mu.RUnlock()

	if synergyDB == nil {
		return 0.0
	}

	totalSynergy := 0.0
	synergyCount := 0

	for _, deckCard := range currentDeck {
		if synergyScore := synergyDB.GetSynergy(cardName, deckCard.Name); synergyScore > 0 {
			totalSynergy += synergyScore
			synergyCount++
		}
	}

	if synergyCount == 0 {
		return 0.0
	}

	return (totalSynergy / float64(synergyCount)) * s.synergyWeight
}

// DeckSynergyDatabase adapts deck.SynergyDatabase to the scoring.SynergyDatabase interface.
// This allows the existing deck.SynergyDatabase to be used with SynergyScorer.
type DeckSynergyDatabase struct {
	db *deck.SynergyDatabase
}

// NewDeckSynergyDatabase creates a new adapter for deck.SynergyDatabase.
func NewDeckSynergyDatabase(db *deck.SynergyDatabase) *DeckSynergyDatabase {
	return &DeckSynergyDatabase{db: db}
}

// GetSynergy returns the synergy score between two cards.
func (a *DeckSynergyDatabase) GetSynergy(card1, card2 string) float64 {
	if a.db == nil {
		return 0.0
	}
	return a.db.GetSynergy(card1, card2)
}

// AnalyzeDeckSynergy returns comprehensive synergy analysis for a deck.
func (a *DeckSynergyDatabase) AnalyzeDeckSynergy(deckNames []string) *deck.DeckSynergyAnalysis {
	if a.db == nil {
		return &deck.DeckSynergyAnalysis{}
	}
	return a.db.AnalyzeDeckSynergy(deckNames)
}
