//go:build research

// Package deck provides versatility scoring research prototypes.
// This is RESEARCH CODE, not production-ready implementations.
//
// Research spike: clash-royale-api-5xb
package deck

import (
	"fmt"
)

// VersatilityScorer calculates versatility bonuses for cards and decks.
// This is a research prototype for exploring versatility measurement.
type VersatilityScorer struct {
	// cardDatabase contains pre-calculated versatility data for each card
	cardDatabase map[string]*CardVersatilityData
}

// CardVersatilityData contains all versatility metrics for a single card.
type CardVersatilityData struct {
	CardName string

	// Multi-role data (60% weight)
	PrimaryRole    CardRole
	SecondaryRoles []CardRole
	EvolutionRoles []CardRole
	MultiRoleScore float64

	// Situational flexibility (30% weight)
	OffensiveEffectiveness float64 // 0-1
	DefensiveEffectiveness float64 // 0-1
	HitsAir                bool
	HitsGround             bool
	HasUtility             bool
	VsTank                 bool
	VsSwarm                bool
	VsAir                  bool
	VsBuilding             bool
	SituationalScore       float64

	// Elixir adaptability (10% weight)
	EffectiveEarly bool
	EffectiveMid   bool
	EffectiveLate  bool
	ElixirScore    float64

	// Calculated total
	TotalVersatility float64
}

// DeckVersatilityReport contains comprehensive versatility analysis.
type DeckVersatilityReport struct {
	CardVersatility     []*CardVersatilityScore
	AverageVersatility  float64
	VersatileCards      []string // Cards with versatility > 1.0
	LowVersatilityCards []string // Cards with versatility < 0.5
	RoleDiversity       float64  // Complement to redundancy analysis
}

// CardVersatilityScore combines versatility data with scoring.
type CardVersatilityScore struct {
	CardName         string
	MultiRoleBonus   float64 // 60% weight
	SituationalBonus float64 // 30% weight
	ElixirBonus      float64 // 10% weight
	TotalScore       float64
}

// NewVersatilityScorer creates a new versatility scorer with a card database.
func NewVersatilityScorer(database map[string]*CardVersatilityData) *VersatilityScorer {
	return &VersatilityScorer{
		cardDatabase: database,
	}
}

// CalculateCardVersatility computes the versatility score for a single card.
func (vs *VersatilityScorer) CalculateCardVersatility(cardName string, evolutionLevel int) (*CardVersatilityScore, error) {
	data, exists := vs.cardDatabase[cardName]
	if !exists {
		return nil, fmt.Errorf("card not found in database: %s", cardName)
	}

	// Multi-role bonus (60%)
	multiRoleScore := data.MultiRoleScore
	if evolutionLevel > 0 && len(data.EvolutionRoles) > 0 {
		// Evolution adds additional role versatility
		multiRoleScore += 0.3
	}
	multiRoleBonus := multiRoleScore * 0.6

	// Situational flexibility (30%)
	situationalBonus := data.SituationalScore * 0.3

	// Elixir adaptability (10%)
	elixirBonus := data.ElixirScore * 0.1

	total := multiRoleBonus + situationalBonus + elixirBonus

	return &CardVersatilityScore{
		CardName:         cardName,
		MultiRoleBonus:   multiRoleBonus,
		SituationalBonus: situationalBonus,
		ElixirBonus:      elixirBonus,
		TotalScore:       total,
	}, nil
}

// CalculateDeckVersatility analyzes versatility for an entire deck.
func (vs *VersatilityScorer) CalculateDeckVersatility(deck []Card, evolutionLevels map[string]int) (*DeckVersatilityReport, error) {
	scores := make([]*CardVersatilityScore, len(deck))
	totalVersatility := 0.0
	versatile := []string{}
	lowVersatile := []string{}

	for i, card := range deck {
		evoLevel := 0
		if card.EvolutionLevel > 0 {
			evoLevel = card.EvolutionLevel
		}

		score, err := vs.CalculateCardVersatility(card.Name, evoLevel)
		if err != nil {
			return nil, fmt.Errorf("error calculating versatility for %s: %w", card.Name, err)
		}

		scores[i] = score
		totalVersatility += score.TotalScore

		if score.TotalScore > 1.0 {
			versatile = append(versatile, card.Name)
		}
		if score.TotalScore < 0.5 {
			lowVersatile = append(lowVersatile, card.Name)
		}
	}

	avgVersatility := totalVersatility / float64(len(deck))
	roleDiversity := vs.calculateRoleDiversity(deck)

	return &DeckVersatilityReport{
		CardVersatility:     scores,
		AverageVersatility:  avgVersatility,
		VersatileCards:      versatile,
		LowVersatilityCards: lowVersatile,
		RoleDiversity:       roleDiversity,
	}, nil
}

// calculateRoleDiversity measures how well the deck covers different roles.
// This is the inverse of redundancy - higher diversity = lower redundancy.
func (vs *VersatilityScorer) calculateRoleDiversity(deck []Card) float64 {
	roleCounts := make(map[CardRole]int)
	for _, card := range deck {
		if data, exists := vs.cardDatabase[card.Name]; exists {
			roleCounts[data.PrimaryRole]++
			for _, role := range data.SecondaryRoles {
				roleCounts[role]++
			}
		}
	}

	// Diversity = number of unique roles represented / maximum possible roles (6)
	uniqueRoles := len(roleCounts)
	maxRoles := 6 // win_condition, building, spell_big, spell_small, support, cycle

	return float64(uniqueRoles) / float64(maxRoles)
}

// MultiRoleVersatility calculates the multi-role component of versatility.
func (vs *VersatilityScorer) MultiRoleVersatility(primary CardRole, secondary, evolution []CardRole) float64 {
	roleCount := 1 + len(secondary) // Count primary + secondary roles
	baseScore := 1.0 + float64(roleCount-1)*0.6

	// Evolution bonus
	if len(evolution) > 0 {
		baseScore += 0.3
	}

	return baseScore
}

// SituationalFlexibility calculates the situational component of versatility.
func (vs *VersatilityScorer) SituationalFlexibility(data *CardVersatilityData) float64 {
	// Component 1: Offense + Defense balance (15%)
	offenseDefenseScore := (data.OffensiveEffectiveness + data.DefensiveEffectiveness) / 2.0
	offenseDefenseBonus := offenseDefenseScore * 0.15

	// Component 2: Target type coverage (10%)
	targetScore := 0.0
	if data.HitsAir && data.HitsGround {
		targetScore += 0.5
	}
	if data.HasUtility {
		targetScore += 0.3
	}
	targetBonus := targetScore * 0.10

	// Component 3: Response variety (5%)
	responseCount := 0
	if data.VsTank {
		responseCount++
	}
	if data.VsSwarm {
		responseCount++
	}
	if data.VsAir {
		responseCount++
	}
	if data.VsBuilding {
		responseCount++
	}
	responseBonus := float64(responseCount) / 4.0 * 0.05

	return offenseDefenseBonus + targetBonus + responseBonus
}

// ElixirAdaptability calculates the elixir flexibility component.
func (vs *VersatilityScorer) ElixirAdaptability(data *CardVersatilityData) float64 {
	viableStages := 0
	if data.EffectiveEarly {
		viableStages++
	}
	if data.EffectiveMid {
		viableStages++
	}
	if data.EffectiveLate {
		viableStages++
	}

	baseScore := float64(viableStages) / 3.0

	// Penalty for single-stage cards
	if viableStages == 1 {
		baseScore *= 0.5
	}

	return baseScore * 0.10
}

// CalculateVersatilityForAo8 computes the Versatility score for the ao8 epic's 5-category system.
// Formula: (1.0 - redundancyPenalty) * 0.6 + versatilityBonus * 0.3 + elixirFlexibility * 0.1
func (vs *VersatilityScorer) CalculateVersatilityForAo8(deck []Card, redundancyPenalty float64) (float64, error) {
	report, err := vs.CalculateDeckVersatility(deck, make(map[string]int))
	if err != nil {
		return 0.0, err
	}

	// Component 1: Role diversity (inverse of redundancy)
	roleDiversity := (1.0 - redundancyPenalty) * 0.6

	// Component 2: Multi-role versatility bonus
	versatilityBonus := report.AverageVersatility * 0.3

	// Component 3: Elixir flexibility (use avg ellixir adaptability from cards)
	elixirFlexibility := report.AverageVersatility * 0.1

	return roleDiversity + versatilityBonus + elixirFlexibility, nil
}

// String returns a human-readable representation of the versatility score.
func (cs *CardVersatilityScore) String() string {
	return fmt.Sprintf("CardVersatility{%s: Multi=%.2f, Sit=%.2f, Elix=%.2f, Total=%.2f}",
		cs.CardName, cs.MultiRoleBonus, cs.SituationalBonus, cs.ElixirBonus, cs.TotalScore)
}

// String returns a human-readable representation of the deck report.
func (dvr *DeckVersatilityReport) String() string {
	return fmt.Sprintf("DeckVersatility{Avg=%.2f, RoleDiv=%.2f, Versatile=%d, LowVers=%d}",
		dvr.AverageVersatility, dvr.RoleDiversity, len(dvr.VersatileCards), len(dvr.LowVersatilityCards))
}

// Card represents a simplified card model for research purposes.
// In production, use the existing Card type from pkg/clashroyale.
type Card struct {
	Name           string
	ElixirCost     int
	EvolutionLevel int
	// Additional fields as needed
}

// ExampleCardDatabase provides sample versatility data for common cards.
// In production, this would be loaded from a JSON file or database.
func ExampleCardDatabase() map[string]*CardVersatilityData {
	return map[string]*CardVersatilityData{
		"Valkyrie": {
			CardName:               "Valkyrie",
			PrimaryRole:            RoleSupport,
			SecondaryRoles:         []CardRole{RoleSupport}, // Anti-swarm
			EvolutionRoles:         []CardRole{RoleSupport}, // Dash utility
			MultiRoleScore:         1.3,
			OffensiveEffectiveness: 0.8,
			DefensiveEffectiveness: 0.9,
			HitsAir:                false,
			HitsGround:             true,
			HasUtility:             true, // Dash
			VsTank:                 true,
			VsSwarm:                true,
			VsAir:                  false,
			VsBuilding:             true,
			SituationalScore:       0.25,
			EffectiveEarly:         true,
			EffectiveMid:           true,
			EffectiveLate:          true,
			ElixirScore:            0.10,
			TotalVersatility:       1.65,
		},
		"Electro Wizard": {
			CardName:               "Electro Wizard",
			PrimaryRole:            RoleSupport,
			SecondaryRoles:         []CardRole{RoleSupport}, // Air defense
			EvolutionRoles:         []CardRole{},
			MultiRoleScore:         1.5,
			OffensiveEffectiveness: 0.8,
			DefensiveEffectiveness: 0.8,
			HitsAir:                true,
			HitsGround:             true,
			HasUtility:             true, // Stun
			VsTank:                 false,
			VsSwarm:                false,
			VsAir:                  true,
			VsBuilding:             true,
			SituationalScore:       0.30,
			EffectiveEarly:         true,
			EffectiveMid:           true,
			EffectiveLate:          true,
			ElixirScore:            0.08,
			TotalVersatility:       1.88,
		},
		"Baby Dragon": {
			CardName:               "Baby Dragon",
			PrimaryRole:            RoleSupport,
			SecondaryRoles:         []CardRole{RoleSupport}, // Anti-air + anti-swarm
			EvolutionRoles:         []CardRole{},
			MultiRoleScore:         1.3,
			OffensiveEffectiveness: 0.7,
			DefensiveEffectiveness: 0.8,
			HitsAir:                true,
			HitsGround:             true,
			HasUtility:             false,
			VsTank:                 false,
			VsSwarm:                true,
			VsAir:                  true,
			VsBuilding:             true,
			SituationalScore:       0.30,
			EffectiveEarly:         true,
			EffectiveMid:           true,
			EffectiveLate:          true,
			ElixirScore:            0.08,
			TotalVersatility:       1.68,
		},
		"Knight": {
			CardName:               "Knight",
			PrimaryRole:            RoleCycle,
			SecondaryRoles:         []CardRole{},
			EvolutionRoles:         []CardRole{RoleSupport}, // Dash adds versatility
			MultiRoleScore:         1.0,                     // Base
			OffensiveEffectiveness: 0.6,
			DefensiveEffectiveness: 0.7,
			HitsAir:                false,
			HitsGround:             true,
			HasUtility:             false,
			VsTank:                 true,
			VsSwarm:                false,
			VsAir:                  false,
			VsBuilding:             false,
			SituationalScore:       0.15,
			EffectiveEarly:         true,
			EffectiveMid:           true,
			EffectiveLate:          true,
			ElixirScore:            0.10,
			TotalVersatility:       1.25, // Base 1.0, with evo becomes 1.43
		},
		"Giant": {
			CardName:               "Giant",
			PrimaryRole:            RoleWinCondition,
			SecondaryRoles:         []CardRole{},
			EvolutionRoles:         []CardRole{},
			MultiRoleScore:         1.0, // Single role
			OffensiveEffectiveness: 0.9,
			DefensiveEffectiveness: 0.3, // Can tank but not defensive
			HitsAir:                false,
			HitsGround:             true,
			HasUtility:             false,
			VsTank:                 false,
			VsSwarm:                false,
			VsAir:                  false,
			VsBuilding:             true,
			SituationalScore:       0.15,
			EffectiveEarly:         false, // Weak without support
			EffectiveMid:           true,
			EffectiveLate:          true,
			ElixirScore:            0.05,
			TotalVersatility:       1.20,
		},
	}
}

// ExampleUsage demonstrates how to use the versatility scorer.
func ExampleVersatilityUsage() {
	database := ExampleCardDatabase()
	scorer := NewVersatilityScorer(database)

	// Example deck
	deck := []Card{
		{Name: "Valkyrie", ElixirCost: 4, EvolutionLevel: 0},
		{Name: "Electro Wizard", ElixirCost: 4, EvolutionLevel: 0},
		{Name: "Baby Dragon", ElixirCost: 4, EvolutionLevel: 0},
		{Name: "Knight", ElixirCost: 3, EvolutionLevel: 0},
		// ... more cards
	}

	report, err := scorer.CalculateDeckVersatility(deck, make(map[string]int))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(report.String())

	// Calculate for ao8 epic's 5-category system
	ao8Score, err := scorer.CalculateVersatilityForAo8(deck, 0.1) // 10% redundancy penalty
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Ao8 Versatility Score: %.2f / 1.0\n", ao8Score)
}
