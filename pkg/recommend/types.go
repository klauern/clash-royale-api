// Package recommend provides deck recommendation functionality that combines
// archetype-based matching with custom variations optimized for player's card collection.
package recommend

import (
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
)

// RecommendationType distinguishes between meta archetype decks and custom variations
type RecommendationType string

const (
	// TypeArchetypeMatch represents a direct match to a meta archetype deck
	TypeArchetypeMatch RecommendationType = "archetype_match"

	// TypeCustomVariation represents a player-optimized variation with card swaps
	TypeCustomVariation RecommendationType = "custom_variation"
)

// UpgradeCost summarizes upgrade requirements for a recommended deck
type UpgradeCost struct {
	CardsNeeded    int     `json:"cards_needed"`
	GoldNeeded     int     `json:"gold_needed"`
	GemsNeeded     int     `json:"gems_needed"`
	DistanceMetric float64 `json:"distance_metric"` // 0.0 = perfect match, 1.0 = far from target
}

// DeckRecommendation represents a single deck recommendation with scoring and analysis
type DeckRecommendation struct {
	// Deck is the recommended deck composition
	Deck *deck.DeckRecommendation `json:"deck"`

	// Archetype is the playstyle this deck represents
	Archetype mulligan.Archetype `json:"archetype"`

	// ArchetypeName is a human-readable archetype name
	ArchetypeName string `json:"archetype_name"`

	// CompatibilityScore (0-100) measures how well player's card levels match this deck
	// Based on card level ratios and rarity weights
	CompatibilityScore float64 `json:"compatibility_score"`

	// SynergyScore (0-100) measures card pair synergies within the deck
	// Uses the synergy database to find strong card combinations
	SynergyScore float64 `json:"synergy_score"`

	// OverallScore (0-100) is the weighted combination of all factors
	// Formula: 60% compatibility + 25% synergy + 15% archetype fit
	OverallScore float64 `json:"overall_score"`

	// Type indicates whether this is a direct archetype match or custom variation
	Type RecommendationType `json:"type"`

	// Reasons explains why this deck is recommended
	Reasons []string `json:"reasons"`

	// UpgradeCost summarizes resources needed to make this deck competitive
	UpgradeCost UpgradeCost `json:"upgrade_cost,omitempty"`
}

// RecommendationResult contains all recommendations for a player
type RecommendationResult struct {
	// PlayerTag is the player being analyzed
	PlayerTag string `json:"player_tag"`

	// PlayerName is the player's display name
	PlayerName string `json:"player_name"`

	// Recommendations is the list of recommended decks, sorted by overall score
	Recommendations []*DeckRecommendation `json:"recommendations"`

	// TopArchetype is the best-matching archetype for this player
	TopArchetype string `json:"top_archetype"`

	// ArenaFilter if specified, filters recommendations to appropriate arena
	ArenaFilter string `json:"arena_filter,omitempty"`

	// LeagueFilter if specified, filters recommendations to appropriate league
	LeagueFilter string `json:"league_filter,omitempty"`

	// GeneratedAt is when these recommendations were created
	GeneratedAt string `json:"generated_at"`
}

// RecommenderOptions configures the recommendation engine
type RecommenderOptions struct {
	// Limit is the maximum number of recommendations to return (default: 5)
	Limit int

	// Arena filters recommendations to cards appropriate for this arena
	Arena string

	// League filters recommendations to cards appropriate for this league
	League string

	// IncludeVariations generates custom variations in addition to archetype matches (default: true)
	IncludeVariations bool

	// MinCompatibility is the minimum compatibility score to include a recommendation (default: 30.0)
	MinCompatibility float64

	// TargetLevel is the target card level for upgrade cost calculations (default: 12)
	TargetLevel int

	// MaxVariationsPerArchetype limits custom variations generated per archetype (default: 2)
	MaxVariationsPerArchetype int
}

// DefaultOptions returns default recommender options
func DefaultOptions() RecommenderOptions {
	return RecommenderOptions{
		Limit:                     5,
		Arena:                     "",
		League:                    "",
		IncludeVariations:         true,
		MinCompatibility:          30.0,
		TargetLevel:               12,
		MaxVariationsPerArchetype: 2,
	}
}
