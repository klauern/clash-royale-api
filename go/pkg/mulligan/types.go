// Package mulligan provides opening hand analysis and strategic recommendations
// for Clash Royale decks based on card composition and matchups.
package mulligan

import "time"

// Archetype represents the general playstyle of a deck
type Archetype string

const (
	ArchetypeBeatdown   Archetype = "beatdown"    // Giant, Golem, Lava Hound decks
	ArchetypeCycle      Archetype = "cycle"       // Hog, Cycle decks
	ArchetypeControl    Archetype = "control"     // Control, spell bait decks
	ArchetypeSiege      Archetype = "siege"       // X-Bow, Mortar decks
	ArchetypeBridgeSpam Archetype = "bridge_spam" // Battle Ram, Bandit decks
	ArchetypeMidrange   Archetype = "midrange"    // Balanced decks
	ArchetypeSpawndeck  Archetype = "spawndeck"   // Goblin Hut, Barbarian Hut
	ArchetypeBait       Archetype = "bait"        // Spell bait decks
)

// Matchup represents opponent deck types for strategic planning
type Matchup struct {
	OpponentType string   `json:"opponent_type"`
	OpeningPlay  string   `json:"opening_play"`
	Reason       string   `json:"reason"`
	Backup       string   `json:"backup"`
	KeyCards     []string `json:"key_cards"`
	DangerLevel  string   `json:"danger_level"` // "low", "medium", "high"
}

// MulliganGuide represents a complete opening hand strategy for a deck
type MulliganGuide struct {
	DeckName          string    `json:"deck_name"`
	DeckCards         []string  `json:"deck_cards"`
	Archetype         Archetype `json:"archetype"`
	GeneralPrinciples []string  `json:"general_principles"`
	Matchups          []Matchup `json:"matchups"`
	NeverOpenWith     []string  `json:"never_open_with"`
	IdealOpenings     []string  `json:"ideal_openings"`
	GeneratedAt       time.Time `json:"generated_at"`
}

// CardRole represents the strategic function of a card in opening plays
type CardRole string

const (
	RoleWinCondition CardRole = "win_condition" // Primary tower damage
	RoleDefensive    CardRole = "defensive"     // Defensive buildings/troops
	RoleCycle        CardRole = "cycle"         // Low cost cycle cards
	RoleSpell        CardRole = "spell"         // Damage/utility spells
	RoleSupport      CardRole = "support"       // Support troops
	RoleBuilding     CardRole = "building"      // Defensive buildings
)

// OpeningCard represents a card with its opening play characteristics
type OpeningCard struct {
	Name         string   `json:"name"`
	Elixir       int      `json:"elixir"`
	Role         CardRole `json:"role"`
	OpeningScore float64  `json:"opening_score"` // Score for opening play viability
	Reasons      []string `json:"reasons"`       // Why this card is good/bad for opening
}

// DeckAnalysis represents the analysis of a deck for mulligan purposes
type DeckAnalysis struct {
	DeckCards      []string      `json:"deck_cards"`
	Archetype      Archetype     `json:"archetype"`
	AvgElixir      float64       `json:"avg_elixir"`
	OpeningCards   []OpeningCard `json:"opening_cards"`
	WinConditions  []string      `json:"win_conditions"`
	DefensiveCards []string      `json:"defensive_cards"`
	CycleCards     []string      `json:"cycle_cards"`
	Spells         []string      `json:"spells"`
	Buildings      []string      `json:"buildings"`
}

// GetRole returns the string representation of the card role
func (cr CardRole) String() string {
	return string(cr)
}

// GetArchetype returns the string representation of the archetype
func (a Archetype) String() string {
	return string(a)
}
