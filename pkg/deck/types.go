// Package deck provides types and utilities for building Clash Royale decks
// with intelligent card selection based on player's collection.
package deck

import (
	"fmt"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// CardRole represents the strategic role of a card in a deck
type CardRole string

const (
	// RoleWinCondition represents primary tower-damaging cards
	RoleWinCondition CardRole = "win_conditions"

	// RoleBuilding represents defensive buildings
	RoleBuilding CardRole = "buildings"

	// RoleSpellBig represents high-elixir damage spells
	RoleSpellBig CardRole = "spells_big"

	// RoleSpellSmall represents low-elixir utility spells
	RoleSpellSmall CardRole = "spells_small"

	// RoleSupport represents mid-cost support troops
	RoleSupport CardRole = "support"

	// RoleCycle represents low-cost cycle cards
	RoleCycle CardRole = "cycle"
)

// String returns the string representation of the card role
func (cr CardRole) String() string {
	return string(cr)
}

// Strategy represents a deck building strategy that influences card selection
type Strategy string

const (
	// StrategyBalanced is the default strategy with neutral preferences
	StrategyBalanced Strategy = "balanced"

	// StrategyAggro focuses on high-damage win conditions and offensive play
	StrategyAggro Strategy = "aggro"

	// StrategyControl emphasizes defensive structures and big spells
	StrategyControl Strategy = "control"

	// StrategyCycle builds fast-cycling decks with low elixir costs
	StrategyCycle Strategy = "cycle"

	// StrategySplash focuses on area-of-effect damage cards
	StrategySplash Strategy = "splash"

	// StrategySpell builds spell-heavy decks with multiple big spells
	StrategySpell Strategy = "spell"
)

// ParseStrategy converts a string to a Strategy type with case-insensitive parsing
func ParseStrategy(s string) (Strategy, error) {
	normalized := strings.ToLower(strings.TrimSpace(s))

	switch normalized {
	case "balanced":
		return StrategyBalanced, nil
	case "aggro":
		return StrategyAggro, nil
	case "control":
		return StrategyControl, nil
	case "cycle":
		return StrategyCycle, nil
	case "splash":
		return StrategySplash, nil
	case "spell":
		return StrategySpell, nil
	default:
		return "", &DeckError{
			Code:    "INVALID_STRATEGY",
			Message: fmt.Sprintf("invalid strategy '%s': must be one of [balanced, aggro, control, cycle, splash, spell]", s),
		}
	}
}

// Validate checks if the strategy is one of the valid predefined strategies
func (s Strategy) Validate() error {
	switch s {
	case StrategyBalanced, StrategyAggro, StrategyControl, StrategyCycle, StrategySplash, StrategySpell:
		return nil
	default:
		return &DeckError{
			Code:    "INVALID_STRATEGY",
			Message: fmt.Sprintf("invalid strategy '%s': must be one of [balanced, aggro, control, cycle, splash, spell]", s),
		}
	}
}

// String returns the string representation of the strategy
func (s Strategy) String() string {
	return string(s)
}

// CardCandidate represents a card being considered for deck building
// with its metadata and calculated score
type CardCandidate struct {
	Name              string
	Level             int
	MaxLevel          int
	Rarity            string
	Elixir            int
	Role              *CardRole
	Score             float64
	HasEvolution      bool
	EvolutionPriority int
	EvolutionLevel    int // Current evolution level (0 = no evolution, 1-3 = evolution stages)
	MaxEvolutionLevel int
	Stats             *clashroyale.CombatStats
}

// LevelRatio returns the card's overall progression as a weighted combination
// of card level and evolution level. Card level is weighted 70%, evolution 30%.
// For cards without evolution capability, returns pure card level ratio.
func (cc *CardCandidate) LevelRatio() float64 {
	if cc.MaxLevel == 0 {
		return 0
	}

	cardLevelRatio := float64(cc.Level) / float64(cc.MaxLevel)

	// If card has no evolution capability, return pure card level ratio
	if cc.MaxEvolutionLevel == 0 {
		return cardLevelRatio
	}

	// Calculate evolution level ratio
	evolutionRatio := float64(cc.EvolutionLevel) / float64(cc.MaxEvolutionLevel)

	// Weighted combination: 70% card level, 30% evolution level
	return (cardLevelRatio * 0.7) + (evolutionRatio * 0.3)
}

// HasRole returns true if the candidate has a defined role
func (cc *CardCandidate) HasRole() bool {
	return cc.Role != nil
}

// DeckRecommendation represents a recommended 8-card deck with metadata
type DeckRecommendation struct {
	Deck           []string     `json:"deck"`
	DeckDetail     []CardDetail `json:"deck_detail"`
	AvgElixir      float64      `json:"average_elixir"`
	AnalysisTime   string       `json:"analysis_time,omitempty"`
	Notes          []string     `json:"notes"`
	EvolutionSlots []string     `json:"evolution_slots,omitempty"`
}

// CardDetail provides detailed information about a card in a recommended deck
type CardDetail struct {
	Name              string  `json:"name"`
	Level             int     `json:"level"`
	MaxLevel          int     `json:"max_level"`
	Rarity            string  `json:"rarity"`
	Elixir            int     `json:"elixir"`
	Role              string  `json:"role,omitempty"`
	Score             float64 `json:"score"`
	EvolutionLevel    int     `json:"evolution_level,omitempty"`
	MaxEvolutionLevel int     `json:"max_evolution_level,omitempty"`
}

// CalculateAvgElixir calculates the average elixir cost from card details
func (dr *DeckRecommendation) CalculateAvgElixir() float64 {
	if len(dr.DeckDetail) == 0 {
		return 0
	}

	total := 0
	for _, card := range dr.DeckDetail {
		total += card.Elixir
	}

	return roundToTwo(float64(total) / float64(len(dr.DeckDetail)))
}

// FormatEvolutionBadge returns a formatted evolution badge for a card.
// Examples: "Evo 1", "Evo 2", or "" if no evolution.
func FormatEvolutionBadge(evolutionLevel int) string {
	if evolutionLevel > 0 {
		return fmt.Sprintf("Evo %d", evolutionLevel)
	}
	return ""
}

// AddNote appends a strategic note to the recommendation
func (dr *DeckRecommendation) AddNote(note string) {
	dr.Notes = append(dr.Notes, note)
}

// Validate checks if the deck recommendation is valid
func (dr *DeckRecommendation) Validate() error {
	if len(dr.Deck) != 8 {
		return ErrInvalidDeckSize
	}

	if len(dr.DeckDetail) != 8 {
		return ErrInvalidDeckSize
	}

	if dr.AvgElixir < 0 || dr.AvgElixir > 10 {
		return ErrInvalidAvgElixir
	}

	return nil
}

// Error types for deck building operations
var (
	ErrInvalidDeckSize   = &DeckError{Code: "INVALID_DECK_SIZE", Message: "deck must contain exactly 8 cards"}
	ErrInvalidAvgElixir  = &DeckError{Code: "INVALID_AVG_ELIXIR", Message: "average elixir must be between 0 and 10"}
	ErrNoWinCondition    = &DeckError{Code: "NO_WIN_CONDITION", Message: "deck must have at least one win condition"}
	ErrInsufficientCards = &DeckError{Code: "INSUFFICIENT_CARDS", Message: "insufficient cards available for deck building"}
)

// DeckError represents a deck building error
type DeckError struct {
	Code    string
	Message string
}

func (e *DeckError) Error() string {
	return e.Message
}

// UpgradeRecommendation represents a single card upgrade recommendation
type UpgradeRecommendation struct {
	CardName     string  `json:"card_name"`
	CurrentLevel int     `json:"current_level"`
	TargetLevel  int     `json:"target_level"`
	Rarity       string  `json:"rarity"`
	Elixir       int     `json:"elixir"`
	Role         string  `json:"role,omitempty"`
	ImpactScore  float64 `json:"impact_score"`
	GoldCost     int     `json:"gold_cost"`
	ValuePerGold float64 `json:"value_per_gold"`
	Reason       string  `json:"reason"`
}

// UpgradeRecommendations represents upgrade suggestions for a deck
type UpgradeRecommendations struct {
	PlayerTag       string                  `json:"player_tag,omitempty"`
	DeckName        string                  `json:"deck_name,omitempty"`
	TotalGoldNeeded int                     `json:"total_gold_needed"`
	Recommendations []UpgradeRecommendation `json:"recommendations"`
	GeneratedAt     string                  `json:"generated_at,omitempty"`
}
