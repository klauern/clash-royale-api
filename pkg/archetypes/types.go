// Package archetypes provides archetype-specific deck generation and analysis
// for Clash Royale. It generates decks for different playstyles (beatdown, cycle,
// control, etc.) and calculates upgrade costs to reach competitive levels.
package archetypes

import (
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
)

// ArchetypeConstraints defines deck building constraints for a specific archetype.
// These constraints guide the deck builder to create decks that match the
// characteristics of each playstyle.
type ArchetypeConstraints struct {
	Archetype      mulligan.Archetype    `json:"archetype"`
	MinElixir      float64               `json:"min_elixir"`      // Minimum average elixir cost
	MaxElixir      float64               `json:"max_elixir"`      // Maximum average elixir cost
	RequiredRoles  map[deck.CardRole]int `json:"required_roles"`  // Minimum count per role
	PreferredCards []string              `json:"preferred_cards"` // Cards that fit this archetype
	ExcludedCards  []string              `json:"excluded_cards"`  // Cards that don't fit
	Description    string                `json:"description"`
}

// CardUpgrade represents upgrade requirements for a single card.
// It tracks how many cards and gold are needed to upgrade from
// the current level to the target level.
type CardUpgrade struct {
	CardName     string `json:"card_name"`
	CurrentLevel int    `json:"current_level"`
	TargetLevel  int    `json:"target_level"`
	Rarity       string `json:"rarity"`
	CardsNeeded  int    `json:"cards_needed"`  // Cards needed for upgrades
	GoldNeeded   int    `json:"gold_needed"`   // Gold needed for upgrades
	LevelGap     int    `json:"level_gap"`     // targetLevel - currentLevel
}

// ArchetypeDeck represents a generated deck for a specific archetype
// with complete cost analysis and upgrade requirements.
type ArchetypeDeck struct {
	Archetype       mulligan.Archetype `json:"archetype"`
	Deck            []string           `json:"deck"`              // Card names in deck
	DeckDetail      []deck.CardDetail  `json:"deck_detail"`       // Full card details
	AvgElixir       float64            `json:"avg_elixir"`        // Average elixir cost
	CurrentAvgLevel float64            `json:"current_avg_level"` // Current average card level
	TargetLevel     int                `json:"target_level"`      // Target competitive level
	CardsNeeded     int                `json:"cards_needed"`      // Total cards to reach target
	GoldNeeded      int                `json:"gold_needed"`       // Total gold to reach target
	DistanceMetric  float64            `json:"distance_metric"`   // 0.0 (perfect) to 1.0 (far from ideal)
	UpgradeDetails  []CardUpgrade      `json:"upgrade_details"`   // Per-card upgrade breakdown
}

// ArchetypeAnalysisResult contains the complete analysis for all archetypes.
// It includes a deck and cost analysis for each of the 8 defined archetypes.
type ArchetypeAnalysisResult struct {
	PlayerTag    string          `json:"player_tag"`
	PlayerName   string          `json:"player_name"`
	TargetLevel  int             `json:"target_level"`
	Archetypes   []ArchetypeDeck `json:"archetypes"`
	AnalysisTime string          `json:"analysis_time"`
}

// SortBy defines sorting options for archetype comparison
type SortBy string

const (
	// SortByDistance sorts by viability (distance metric, lower is better)
	SortByDistance SortBy = "distance"
	// SortByCardsNeeded sorts by upgrade investment (fewer cards is better)
	SortByCardsNeeded SortBy = "cards_needed"
	// SortByAvgLevel sorts by current deck strength (higher is better)
	SortByAvgLevel SortBy = "avg_level"
)
