// Package evaluation provides comprehensive deck evaluation functionality
// including scoring, analysis, and archetype detection.
package evaluation

import (
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// Archetype represents a detected deck archetype classification
type Archetype string

const (
	// ArchetypeBeatdown represents heavy tank-based decks
	ArchetypeBeatdown Archetype = "beatdown"

	// ArchetypeControl represents defensive decks with spell control
	ArchetypeControl Archetype = "control"

	// ArchetypeCycle represents fast-cycling, low-elixir decks
	ArchetypeCycle Archetype = "cycle"

	// ArchetypeBridge represents aggressive bridge spam decks
	ArchetypeBridge Archetype = "bridge"

	// ArchetypeSiege represents building-based siege decks
	ArchetypeSiege Archetype = "siege"

	// ArchetypeBait represents spell bait decks
	ArchetypeBait Archetype = "bait"

	// ArchetypeGraveyard represents graveyard-focused decks
	ArchetypeGraveyard Archetype = "graveyard"

	// ArchetypeMiner represents miner-focused decks
	ArchetypeMiner Archetype = "miner"

	// ArchetypeHybrid represents decks with multiple win conditions
	ArchetypeHybrid Archetype = "hybrid"

	// ArchetypeUnknown represents decks that don't fit standard archetypes
	ArchetypeUnknown Archetype = "unknown"
)

// String returns the string representation of the archetype
func (a Archetype) String() string {
	return string(a)
}

// Rating represents a qualitative assessment rating
type Rating string

const (
	// RatingGodly represents exceptional performance (9.0-10.0)
	RatingGodly Rating = "Godly!"

	// RatingAmazing represents outstanding performance (8.0-8.9)
	RatingAmazing Rating = "Amazing"

	// RatingGreat represents excellent performance (7.0-7.9)
	RatingGreat Rating = "Great"

	// RatingGood represents good performance (6.0-6.9)
	RatingGood Rating = "Good"

	// RatingDecent represents acceptable performance (5.0-5.9)
	RatingDecent Rating = "Decent"

	// RatingMediocre represents mediocre performance (4.0-4.9)
	RatingMediocre Rating = "Mediocre"

	// RatingPoor represents poor performance (3.0-3.9)
	RatingPoor Rating = "Poor"

	// RatingBad represents bad performance (2.0-2.9)
	RatingBad Rating = "Bad"

	// RatingTerrible represents terrible performance (1.0-1.9)
	RatingTerrible Rating = "Terrible"

	// RatingAwful represents awful performance (0.0-0.9)
	RatingAwful Rating = "Awful"
)

// String returns the string representation of the rating
func (r Rating) String() string {
	return string(r)
}

// CategoryScore represents a score for a specific evaluation category
// with both numeric score and qualitative assessment
type CategoryScore struct {
	// Score is the numeric score (0.0-10.0)
	Score float64 `json:"score"`

	// Rating is the qualitative assessment
	Rating Rating `json:"rating"`

	// Assessment is detailed text explanation
	Assessment string `json:"assessment"`

	// Stars is the 1-3 star visual representation
	Stars int `json:"stars"`
}

// EvaluationResult contains all evaluation data for a deck
type EvaluationResult struct {
	// Core deck information
	Deck      []string `json:"deck"`
	AvgElixir float64  `json:"average_elixir"`

	// Category scores (5 categories)
	Attack       CategoryScore `json:"attack"`
	Defense      CategoryScore `json:"defense"`
	Synergy      CategoryScore `json:"synergy"`
	Versatility  CategoryScore `json:"versatility"`
	F2PFriendly  CategoryScore `json:"f2p_friendly"`

	// Overall score (weighted average)
	OverallScore  float64 `json:"overall_score"`
	OverallRating Rating  `json:"overall_rating"`

	// Archetype detection
	DetectedArchetype Archetype `json:"detected_archetype"`
	ArchetypeConfidence float64  `json:"archetype_confidence"` // 0.0-1.0

	// Detailed analysis sections
	DefenseAnalysis    AnalysisSection `json:"defense_analysis"`
	AttackAnalysis     AnalysisSection `json:"attack_analysis"`
	BaitAnalysis       AnalysisSection `json:"bait_analysis"`
	CycleAnalysis      AnalysisSection `json:"cycle_analysis"`
	LadderAnalysis     AnalysisSection `json:"ladder_analysis"`

	// Synergy matrix
	SynergyMatrix SynergyMatrix `json:"synergy_matrix,omitempty"`
}

// AnalysisSection represents detailed analysis for a specific aspect
type AnalysisSection struct {
	// Title of the analysis section
	Title string `json:"title"`

	// Summary is a brief overview
	Summary string `json:"summary"`

	// Details contains specific observations and insights
	Details []string `json:"details"`

	// Score for this section (0.0-10.0)
	Score float64 `json:"score"`

	// Rating for this section
	Rating Rating `json:"rating"`
}

// SynergyMatrix represents all synergy relationships in a deck
type SynergyMatrix struct {
	// Pairs is all synergy pairs found in the deck
	Pairs []deck.SynergyPair `json:"pairs"`

	// TotalScore is the overall synergy score (0.0-10.0)
	TotalScore float64 `json:"total_score"`

	// AverageSynergy is the average synergy per pair (0.0-1.0)
	AverageSynergy float64 `json:"average_synergy"`

	// PairCount is the number of synergy pairs found
	PairCount int `json:"pair_count"`

	// MaxPossiblePairs is 28 (combinations of 8 cards)
	MaxPossiblePairs int `json:"max_possible_pairs"`

	// SynergyCoverage is percentage of cards with synergies
	SynergyCoverage float64 `json:"synergy_coverage"` // 0.0-100.0
}
