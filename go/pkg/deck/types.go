// Package deck provides types and utilities for building Clash Royale decks
// with intelligent card selection based on player's collection.
package deck

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
	MaxEvolutionLevel int
}

// LevelRatio returns the card's level as a ratio of its max level
func (cc *CardCandidate) LevelRatio() float64 {
	if cc.MaxLevel == 0 {
		return 0
	}
	return float64(cc.Level) / float64(cc.MaxLevel)
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
	Name     string  `json:"name"`
	Level    int     `json:"level"`
	MaxLevel int     `json:"max_level"`
	Rarity   string  `json:"rarity"`
	Elixir   int     `json:"elixir"`
	Role     string  `json:"role,omitempty"`
	Score    float64 `json:"score"`
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

	return float64(total) / float64(len(dr.DeckDetail))
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
