// Package analysis provides types and utilities for analyzing Clash Royale card collections,
// calculating upgrade priorities, and tracking player progression.
package analysis

import (
	"encoding/json"
	"time"
)

// CardAnalysis represents a complete analysis of a player's card collection
// with rarity breakdowns, level information, and upgrade recommendations
type CardAnalysis struct {
	PlayerTag       string                   `json:"player_tag"`
	PlayerName      string                   `json:"player_name,omitempty"`
	AnalysisTime    time.Time                `json:"analysis_time"`
	TotalCards      int                      `json:"total_cards"`
	CardLevels      map[string]CardLevelInfo `json:"card_levels"`
	RarityBreakdown map[string]RarityStats   `json:"rarity_breakdown"`
	UpgradePriority []UpgradePriority        `json:"upgrade_priority"`
	Summary         CollectionSummary        `json:"summary"`
}

// CardLevelInfo provides detailed information about a single card's level and upgrade status
type CardLevelInfo struct {
	Name        string `json:"name"`
	ID          int    `json:"id,omitempty"`
	Level       int    `json:"level"`
	MaxLevel    int    `json:"max_level"`
	Rarity      string `json:"rarity"`
	CardCount   int    `json:"card_count"`
	CardsToNext int    `json:"cards_to_next_level"`
	IsMaxLevel  bool   `json:"is_max_level"`
}

// LevelRatio returns the card's level as a ratio of its max level (0.0 to 1.0)
func (cli *CardLevelInfo) LevelRatio() float64 {
	if cli.MaxLevel == 0 {
		return 0
	}
	return float64(cli.Level) / float64(cli.MaxLevel)
}

// ProgressToNext returns the progress toward next level as a percentage (0-100)
func (cli *CardLevelInfo) ProgressToNext() float64 {
	if cli.IsMaxLevel || cli.CardsToNext == 0 {
		return 100.0
	}
	needed := cli.CardsToNext
	current := cli.CardCount
	return (float64(current) / float64(needed)) * 100.0
}

// RarityStats provides aggregate statistics for cards of a specific rarity
type RarityStats struct {
	Rarity           string  `json:"rarity"`
	TotalCards       int     `json:"total_cards"`
	MaxLevelCards    int     `json:"max_level_cards"`
	AvgLevel         float64 `json:"avg_level"`
	AvgLevelRatio    float64 `json:"avg_level_ratio"`
	CardsNearMax     int     `json:"cards_near_max"`      // Within 1-2 levels of max
	CardsReadyUpgrade int    `json:"cards_ready_upgrade"` // Have enough cards to upgrade
}

// UpgradePriority represents a card recommended for upgrade with scoring
type UpgradePriority struct {
	CardName      string  `json:"card_name"`
	Rarity        string  `json:"rarity"`
	CurrentLevel  int     `json:"current_level"`
	MaxLevel      int     `json:"max_level"`
	CardsOwned    int     `json:"cards_owned"`
	CardsNeeded   int     `json:"cards_needed"`
	Priority      string  `json:"priority"`       // "high", "medium", "low"
	PriorityScore float64 `json:"priority_score"` // 0-100
	Reasons       []string `json:"reasons"`
}

// IsReadyToUpgrade returns true if the player has enough cards to upgrade
func (up *UpgradePriority) IsReadyToUpgrade() bool {
	return up.CardsOwned >= up.CardsNeeded
}

// PercentageComplete returns progress toward upgrade as percentage (0-100)
func (up *UpgradePriority) PercentageComplete() float64 {
	if up.CardsNeeded == 0 {
		return 100.0
	}
	return (float64(up.CardsOwned) / float64(up.CardsNeeded)) * 100.0
}

// CollectionSummary provides high-level overview statistics
type CollectionSummary struct {
	TotalCards       int     `json:"total_cards"`
	MaxLevelCards    int     `json:"max_level_cards"`
	UpgradableCards  int     `json:"upgradable_cards"`    // Ready to upgrade now
	AvgCardLevel     float64 `json:"avg_card_level"`
	AvgLevelRatio    float64 `json:"avg_level_ratio"`     // Overall progress (0-1)
	CompletionPercent float64 `json:"completion_percent"` // Max level cards / total
}

// AnalysisOptions configures card collection analysis behavior
type AnalysisOptions struct {
	IncludeMaxLevel     bool     `json:"include_max_level"`      // Include max level cards in analysis
	MinPriorityScore    float64  `json:"min_priority_score"`     // Minimum score for upgrade priority list
	FocusRarities       []string `json:"focus_rarities"`         // Filter to specific rarities
	ExcludeCards        []string `json:"exclude_cards"`          // Cards to exclude from recommendations
	PrioritizeWinCons   bool     `json:"prioritize_win_cons"`    // Boost priority for win condition cards
	TopN                int      `json:"top_n"`                  // Return only top N upgrade priorities
}

// DefaultAnalysisOptions returns sensible defaults for card analysis
func DefaultAnalysisOptions() AnalysisOptions {
	return AnalysisOptions{
		IncludeMaxLevel:  false,
		MinPriorityScore: 30.0,
		FocusRarities:    []string{},
		ExcludeCards:     []string{},
		PrioritizeWinCons: true,
		TopN:             15,
	}
}

// CalculateSummary computes collection summary from card level info
func (ca *CardAnalysis) CalculateSummary() {
	if len(ca.CardLevels) == 0 {
		ca.Summary = CollectionSummary{}
		return
	}

	totalLevel := 0
	totalLevelRatio := 0.0
	maxLevelCount := 0
	upgradableCount := 0

	for _, card := range ca.CardLevels {
		totalLevel += card.Level
		totalLevelRatio += card.LevelRatio()

		if card.IsMaxLevel {
			maxLevelCount++
		}

		// Card is upgradable if it has enough cards for next level
		if !card.IsMaxLevel && card.CardCount >= card.CardsToNext {
			upgradableCount++
		}
	}

	cardCount := len(ca.CardLevels)
	ca.Summary = CollectionSummary{
		TotalCards:        cardCount,
		MaxLevelCards:     maxLevelCount,
		UpgradableCards:   upgradableCount,
		AvgCardLevel:      float64(totalLevel) / float64(cardCount),
		AvgLevelRatio:     totalLevelRatio / float64(cardCount),
		CompletionPercent: (float64(maxLevelCount) / float64(cardCount)) * 100.0,
	}
}

// Validate checks if the card analysis is valid
func (ca *CardAnalysis) Validate() error {
	if ca.PlayerTag == "" {
		return ErrMissingPlayerTag
	}

	if ca.TotalCards < 0 {
		return ErrInvalidCardCount
	}

	if len(ca.CardLevels) != ca.TotalCards {
		return ErrCardCountMismatch
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling for time handling
func (ca CardAnalysis) MarshalJSON() ([]byte, error) {
	type Alias CardAnalysis
	return json.Marshal(&struct {
		AnalysisTime string `json:"analysis_time"`
		*Alias
	}{
		AnalysisTime: ca.AnalysisTime.Format(time.RFC3339),
		Alias:        (*Alias)(&ca),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for time handling
func (ca *CardAnalysis) UnmarshalJSON(data []byte) error {
	type Alias CardAnalysis
	aux := &struct {
		AnalysisTime string `json:"analysis_time"`
		*Alias
	}{
		Alias: (*Alias)(ca),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var err error
	ca.AnalysisTime, err = time.Parse(time.RFC3339, aux.AnalysisTime)
	if err != nil {
		return err
	}

	return nil
}

// Error types for analysis operations
var (
	ErrMissingPlayerTag   = &AnalysisError{Code: "MISSING_PLAYER_TAG", Message: "player tag is required"}
	ErrInvalidCardCount   = &AnalysisError{Code: "INVALID_CARD_COUNT", Message: "card count cannot be negative"}
	ErrCardCountMismatch  = &AnalysisError{Code: "CARD_COUNT_MISMATCH", Message: "card levels count does not match total cards"}
	ErrNoUpgradePriorities = &AnalysisError{Code: "NO_UPGRADE_PRIORITIES", Message: "no upgrade priorities found"}
	ErrInvalidPriorityScore = &AnalysisError{Code: "INVALID_PRIORITY_SCORE", Message: "priority score must be between 0 and 100"}
)

// AnalysisError represents an analysis-related error
type AnalysisError struct {
	Code    string
	Message string
}

func (e *AnalysisError) Error() string {
	return e.Message
}
