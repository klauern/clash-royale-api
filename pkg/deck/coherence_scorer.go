// Package deck provides archetype coherence scoring for strategic deck evaluation.
//
// This file implements dynamic archetype detection and coherence validation based on
// configurable archetype requirements. It replaces hardcoded card checks with a
// data-driven approach that can be extended without code changes.
package deck

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

//go:embed archetype_requirements.json
var defaultArchetypeRequirementsJSON []byte

// ArchetypeRequirementsConfig represents the JSON configuration for archetype validation
type ArchetypeRequirementsConfig struct {
	Version          int                       `json:"version"`
	Description      string                    `json:"description"`
	LastUpdated      string                    `json:"last_updated"`
	Archetypes       map[string]ArchetypeDef   `json:"archetypes"`
	AntiSynergyRules AntiSynergyRules          `json:"anti_synergy_rules"`
	CardCategories   CardCategoryDefinitions   `json:"card_categories"`
}

// ArchetypeDef defines the requirements for a coherent archetype
type ArchetypeDef struct {
	Name                  string              `json:"name"`
	ElixirRange           ElixirRange         `json:"elixir_range"`
	RequiredWinConditions []string            `json:"required_win_conditions"`
	RequiredSupportCount  MinMax              `json:"required_support_count"`
	SupportCategories     []string            `json:"support_categories"`
	MinCycleCards         int                 `json:"min_cycle_cards,omitempty"`
	MaxCycleCards         int                 `json:"max_cycle_cards,omitempty"`
	MinBaitCards          int                 `json:"min_bait_cards,omitempty"`
	MinBigSpells          int                 `json:"min_big_spells,omitempty"`
	MaxBigSpells          int                 `json:"max_big_spells,omitempty"`
	MinBuildings          int                 `json:"min_buildings,omitempty"`
	MaxBuildings          int                 `json:"max_buildings,omitempty"`
	MinDefensiveCards     int                 `json:"min_defensive_cards,omitempty"`
	MaxWinConditions      int                 `json:"max_win_conditions,omitempty"`
	MinFastThreats        int                 `json:"min_fast_threats,omitempty"`
	PreferredCardRoles    map[string]MinMax   `json:"preferred_card_roles"`
}

// MinMax defines minimum and maximum counts
type MinMax struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// AntiSynergyRules defines penalties for conflicting card combinations
type AntiSynergyRules struct {
	ConflictingWinConditions []ConflictingWinConditionRule `json:"conflicting_win_conditions"`
	CompositionViolations    []CompositionViolation        `json:"composition_violations"`
}

// ConflictingWinConditionRule defines incompatible win condition combinations
type ConflictingWinConditionRule struct {
	Name        string   `json:"name"`
	CardsA      []string `json:"cards_a"`
	CardsB      []string `json:"cards_b,omitempty"`
	MaxAllowed  int      `json:"max_allowed,omitempty"`
	Penalty     float64  `json:"penalty"`
	Reason      string   `json:"reason"`
}

// CompositionViolation defines penalties for deck composition issues
type CompositionViolation struct {
	Name              string   `json:"name"`
	Threshold         int      `json:"threshold"`
	MinWinConditions  int      `json:"min_win_conditions,omitempty"`
	MinAirDefense     int      `json:"min_air_defense,omitempty"`
	Penalty           float64  `json:"penalty"`
	Reason            string   `json:"reason"`
}

// CardCategoryDefinitions groups cards by strategic categories
type CardCategoryDefinitions struct {
	CycleCards    []string `json:"cycle_cards"`
	BaitCards     []string `json:"bait_cards"`
	SplashDamage  []string `json:"splash_damage"`
	HighDPS       []string `json:"high_dps"`
	MiniTanks     []string `json:"mini_tanks"`
	ResetCards    []string `json:"reset_cards"`
	BigSpells     []string `json:"big_spells"`
	SmallSpells   []string `json:"small_spells"`
	AirDefense    []string `json:"air_defense"`
	FastThreats   []string `json:"fast_threats"`
}

// CoherenceResult contains the detailed archetype coherence analysis
type CoherenceResult struct {
	// Detected archetype information
	PrimaryArchetype     string  `json:"primary_archetype"`
	ArchetypeConfidence  float64 `json:"archetype_confidence"` // 0.0-1.0
	CoherenceScore       float64 `json:"coherence_score"`       // 0.0-1.0 overall coherence

	// Elixir analysis
	AverageElixir        float64  `json:"average_elixir"`
	ElixirMatch          bool     `json:"elixir_match"`
	ElixirVariance       float64  `json:"elixir_variance"`

	// Component violations
	Violations          []CoherenceViolation `json:"violations"`
	Bonuses             []CoherenceBonus    `json:"bonuses"`

	// Archetype-specific metrics
	WinConditionCount   int                     `json:"win_condition_count"`
	SupportCount        int                     `json:"support_count"`
	BuildingCount       int                     `json:"building_count"`
	SpellCount          int                     `json:"spell_count"`
	CycleCardCount      int                     `json:"cycle_card_count"`
	BaitCardCount       int                     `json:"bait_card_count"`
	FastThreatCount     int                     `json:"fast_threat_count"`
	AirDefenseCount     int                     `json:"air_defense_count"`

	// Role distribution
	RoleDistribution    map[string]int          `json:"role_distribution"`
}

// CoherenceViolation represents a coherence issue with penalty
type CoherenceViolation struct {
	Type       string  `json:"type"` // "anti_synergy", "composition", "elixir", "missing_cards"
	Severity   float64 `json:"severity"` // 0.0-1.0
	Penalty    float64 `json:"penalty"` // Applied score penalty
	Message    string  `json:"message"`
	Cards      []string `json:"cards,omitempty"`
}

// CoherenceBonus represents a coherence bonus
type CoherenceBonus struct {
	Type     string  `json:"type"`
	Bonus    float64 `json:"bonus"`
	Message  string  `json:"message"`
}

// CoherenceScorer analyzes deck archetype coherence
type CoherenceScorer struct {
	config *ArchetypeRequirementsConfig
	// Cached card category lookups
	cardInCategory map[string]map[string]bool // card -> category -> bool
}

// LoadCoherenceScorer loads archetype requirements from a JSON file.
// If the file cannot be found or read, falls back to embedded defaults.
func LoadCoherenceScorer(dataPath string) (*CoherenceScorer, error) {
	var config ArchetypeRequirementsConfig
	var data []byte
	var err error

	// Try loading from file if path provided
	if dataPath != "" {
		data, err = os.ReadFile(dataPath)
		if err == nil {
			if err := json.Unmarshal(data, &config); err != nil {
				return nil, fmt.Errorf("failed to parse archetype requirements: %w", err)
			}
		}
	}

	// Fall back to embedded defaults
	if len(data) == 0 {
		if err := json.Unmarshal(defaultArchetypeRequirementsJSON, &config); err != nil {
			return nil, fmt.Errorf("failed to parse embedded archetype requirements: %w", err)
		}
	}

	return NewCoherenceScorer(&config), nil
}

// NewCoherenceScorer creates a coherence scorer from configuration
func NewCoherenceScorer(config *ArchetypeRequirementsConfig) *CoherenceScorer {
	cs := &CoherenceScorer{
		config:         config,
		cardInCategory: make(map[string]map[string]bool),
	}
	cs.buildCategoryLookups()
	return cs
}

// buildCategoryLookups creates fast lookup maps for card categories
func (cs *CoherenceScorer) buildCategoryLookups() {
	categories := map[string][]string{
		"cycle":      cs.config.CardCategories.CycleCards,
		"bait":       cs.config.CardCategories.BaitCards,
		"splash":     cs.config.CardCategories.SplashDamage,
		"high_dps":   cs.config.CardCategories.HighDPS,
		"mini_tank":  cs.config.CardCategories.MiniTanks,
		"reset":      cs.config.CardCategories.ResetCards,
		"big_spell":  cs.config.CardCategories.BigSpells,
		"small_spell": cs.config.CardCategories.SmallSpells,
		"air_defense": cs.config.CardCategories.AirDefense,
		"fast_threat": cs.config.CardCategories.FastThreats,
	}

	for category, cards := range categories {
		for _, card := range cards {
			if cs.cardInCategory[card] == nil {
				cs.cardInCategory[card] = make(map[string]bool)
			}
			cs.cardInCategory[card][category] = true
		}
	}
}

// isCardInCategory checks if a card belongs to a category
func (cs *CoherenceScorer) isCardInCategory(cardName, category string) bool {
	if cats, exists := cs.cardInCategory[cardName]; exists {
		return cats[category]
	}
	return false
}

// AnalyzeCoherence performs a comprehensive archetype coherence analysis
func (cs *CoherenceScorer) AnalyzeCoherence(cards []CardCandidate, strategy Strategy) *CoherenceResult {
	if len(cards) == 0 {
		return &CoherenceResult{CoherenceScore: 0.0}
	}

	result := &CoherenceResult{
		CoherenceScore:    0.8, // Start with good score, apply penalties
		Violations:        []CoherenceViolation{},
		Bonuses:           []CoherenceBonus{},
		RoleDistribution:  make(map[string]int),
	}

	// Analyze card composition
	cs.analyzeComposition(cards, result)

	// Detect primary archetype
	result.PrimaryArchetype, result.ArchetypeConfidence = cs.detectArchetype(cards, result)

	// Calculate elixer metrics
	result.AverageElixir = cs.calculateAverageElixir(cards)

	// Apply anti-synergy penalties
	cs.applyAntiSynergyPenalties(cards, result)

	// Apply composition violations
	cs.applyCompositionViolations(cards, result)

	// Apply archetype-specific validation
	if archetype, exists := cs.config.Archetypes[result.PrimaryArchetype]; exists {
		cs.validateArchetypeRequirements(cards, archetype, result)
	}

	// Calculate elixir match
	cs.validateElixirRange(cards, strategy, result)

	// Final score calculation (apply all penalties)
	for _, violation := range result.Violations {
		result.CoherenceScore -= violation.Penalty
	}

	// Apply bonuses
	for _, bonus := range result.Bonuses {
		result.CoherenceScore += bonus.Bonus
	}

	// Clamp to valid range
	if result.CoherenceScore > 1.0 {
		result.CoherenceScore = 1.0
	}
	if result.CoherenceScore < 0.0 {
		result.CoherenceScore = 0.0
	}

	return result
}

// analyzeComposition counts cards by role and category
func (cs *CoherenceScorer) analyzeComposition(cards []CardCandidate, result *CoherenceResult) {
	for _, card := range cards {
		// Count by role
		if card.Role != nil {
			role := string(*card.Role)
			result.RoleDistribution[role]++

			switch *card.Role {
			case RoleWinCondition:
				result.WinConditionCount++
			case RoleBuilding:
				result.BuildingCount++
			case RoleSpellBig, RoleSpellSmall:
				result.SpellCount++
			case RoleSupport:
				result.SupportCount++
			}
		}

		// Count by category
		if cs.isCardInCategory(card.Name, "cycle") {
			result.CycleCardCount++
		}
		if cs.isCardInCategory(card.Name, "bait") {
			result.BaitCardCount++
		}
		if cs.isCardInCategory(card.Name, "fast_threat") {
			result.FastThreatCount++
		}
		if cs.isCardInCategory(card.Name, "air_defense") {
			result.AirDefenseCount++
		}
	}
}

// detectArchetype determines the most likely archetype for a deck
func (cs *CoherenceScorer) detectArchetype(cards []CardCandidate, result *CoherenceResult) (string, float64) {
	cardNames := make(map[string]bool)
	for _, card := range cards {
		cardNames[card.Name] = true
	}

	// Score each archetype by how well the deck matches
	bestArchetype := ""
	bestScore := 0.0

	for archetypeName, archetype := range cs.config.Archetypes {
		score := 0.0
		maxScore := 0.0

		// Check for required win conditions
		wcMatch := false
		for _, wc := range archetype.RequiredWinConditions {
			if cardNames[wc] {
				wcMatch = true
				break
			}
		}
		if wcMatch {
			score += 2.0
		}
		maxScore += 2.0

		// Check elixir range
		avgElixir := result.AverageElixir
		if avgElixir >= archetype.ElixirRange.Min && avgElixir <= archetype.ElixirRange.Max {
			score += 1.5
		} else {
			// Partial credit for being close
			diff := 0.0
			if avgElixir < archetype.ElixirRange.Min {
				diff = archetype.ElixirRange.Min - avgElixir
			} else {
				diff = avgElixir - archetype.ElixirRange.Max
			}
			if diff < 0.5 {
				score += 0.5
			}
		}
		maxScore += 1.5

		// Check support count
		if result.SupportCount >= archetype.RequiredSupportCount.Min &&
			result.SupportCount <= archetype.RequiredSupportCount.Max {
			score += 1.0
		}
		maxScore += 1.0

		// Check cycle cards (if relevant)
		if archetype.MaxCycleCards > 0 && result.CycleCardCount <= archetype.MaxCycleCards {
			score += 0.5
		}
		maxScore += 0.5

		// Normalize score
		if maxScore > 0 {
			confidence := score / maxScore
			if confidence > bestScore {
				bestScore = confidence
				bestArchetype = archetypeName
			}
		}
	}

	// Default to cycle if no archetype matched well
	if bestArchetype == "" || bestScore < 0.3 {
		bestArchetype = "cycle"
		bestScore = 0.3
	}

	return bestArchetype, bestScore
}

// applyAntiSynergyPenalties checks for conflicting card combinations
func (cs *CoherenceScorer) applyAntiSynergyPenalties(cards []CardCandidate, result *CoherenceResult) {
	cardNames := make(map[string]bool)
	for _, card := range cards {
		cardNames[card.Name] = true
	}

	for _, rule := range cs.config.AntiSynergyRules.ConflictingWinConditions {
		// Check for cards from group A
		hasA := false
		cardsA := []string{}
		for _, card := range rule.CardsA {
			if cardNames[card] {
				hasA = true
				cardsA = append(cardsA, card)
			}
		}

		// Check max allowed for same-group conflicts
		if rule.MaxAllowed > 0 && len(cardsA) > rule.MaxAllowed {
			result.Violations = append(result.Violations, CoherenceViolation{
				Type:     "anti_synergy",
				Severity: rule.Penalty,
				Penalty:  rule.Penalty,
				Message:  rule.Reason,
				Cards:    cardsA,
			})
			continue
		}

		// Check for conflicts between group A and B
		if len(rule.CardsB) > 0 {
			hasB := false
			cardsB := []string{}
			for _, card := range rule.CardsB {
				if cardNames[card] {
					hasB = true
					cardsB = append(cardsB, card)
				}
			}

			if hasA && hasB {
				result.Violations = append(result.Violations, CoherenceViolation{
					Type:     "anti_synergy",
					Severity: rule.Penalty,
					Penalty:  rule.Penalty,
					Message:  rule.Reason,
					Cards:    append(cardsA, cardsB...),
				})
			}
		}
	}
}

// applyCompositionViolations checks for general composition issues
func (cs *CoherenceScorer) applyCompositionViolations(cards []CardCandidate, result *CoherenceResult) {
	for _, rule := range cs.config.AntiSynergyRules.CompositionViolations {
		violation := false

		switch rule.Name {
		case "Too Many Buildings":
			violation = result.BuildingCount > rule.Threshold
		case "Too Many Spells":
			violation = result.SpellCount > rule.Threshold
		case "No Win Condition":
			violation = result.WinConditionCount < rule.MinWinConditions
		case "No Air Defense":
			violation = result.AirDefenseCount < rule.MinAirDefense
		}

		if violation {
			result.Violations = append(result.Violations, CoherenceViolation{
				Type:     "composition",
				Severity: rule.Penalty,
				Penalty:  rule.Penalty,
				Message:  rule.Reason,
			})
		}
	}
}

// validateArchetypeRequirements checks archetype-specific requirements
func (cs *CoherenceScorer) validateArchetypeRequirements(cards []CardCandidate, archetype ArchetypeDef, result *CoherenceResult) {
	// Validate win condition count
	if result.WinConditionCount < archetype.PreferredCardRoles["win_condition"].Min {
		result.Violations = append(result.Violations, CoherenceViolation{
			Type:     "missing_cards",
			Severity: 0.15,
			Penalty:  0.15,
			Message:  fmt.Sprintf("%s archetype needs at least %d win condition(s)", archetype.Name, archetype.PreferredCardRoles["win_condition"].Min),
		})
	}

	// Validate cycle cards for beatdown (shouldn't have many)
	if archetype.MaxCycleCards > 0 && result.CycleCardCount > archetype.MaxCycleCards {
		result.Violations = append(result.Violations, CoherenceViolation{
			Type:     "composition",
			Severity: 0.10,
			Penalty:  0.10,
			Message:  fmt.Sprintf("%s archetype should have max %d cycle cards, has %d", archetype.Name, archetype.MaxCycleCards, result.CycleCardCount),
		})
	}

	// Validate building count for bridge spam
	if archetype.MaxBuildings > 0 && result.BuildingCount > archetype.MaxBuildings {
		result.Violations = append(result.Violations, CoherenceViolation{
			Type:     "composition",
			Severity: 0.10,
			Penalty:  0.10,
			Message:  fmt.Sprintf("%s archetype should have max %d building(s), has %d", archetype.Name, archetype.MaxBuildings, result.BuildingCount),
		})
	}

	// Validate fast threats for bridge spam
	if archetype.MinFastThreats > 0 && result.FastThreatCount < archetype.MinFastThreats {
		result.Violations = append(result.Violations, CoherenceViolation{
			Type:     "missing_cards",
			Severity: 0.10,
			Penalty:  0.10,
			Message:  fmt.Sprintf("%s archetype needs at least %d fast threat(s)", archetype.Name, archetype.MinFastThreats),
		})
	}

	// Bonus for meeting archetype requirements
	if result.WinConditionCount >= 1 && result.SupportCount >= 2 {
		result.Bonuses = append(result.Bonuses, CoherenceBonus{
			Type:    "archetype_core",
			Bonus:   0.05,
			Message: fmt.Sprintf("Solid %s core established", archetype.Name),
		})
	}
}

// validateElixirRange checks if elixir curve matches archetype
func (cs *CoherenceScorer) validateElixirRange(cards []CardCandidate, strategy Strategy, result *CoherenceResult) {
	// Calculate variance
	if len(cards) == 0 {
		return
	}

	variance := 0.0
	for _, card := range cards {
		diff := float64(card.Elixir) - result.AverageElixir
		variance += diff * diff
	}
	variance /= float64(len(cards))
	result.ElixirVariance = variance

	// Check against strategy profile
	profile, exists := StrategyElixirProfiles[strategy]
	if !exists {
		profile = StrategyElixirProfiles[StrategyBalanced]
	}

	result.ElixirMatch = result.AverageElixir >= profile.Min && result.AverageElixir <= profile.Max

	if !result.ElixirMatch {
		result.Violations = append(result.Violations, CoherenceViolation{
			Type:     "elixir",
			Severity: 0.10,
			Penalty:  0.10,
			Message:  fmt.Sprintf("Average elixir %.1f doesn't match %s strategy (target: %.1f-%.1f)", result.AverageElixir, strategy, profile.Min, profile.Max),
		})
	}
}

// calculateAverageElixir computes the average elixir cost
func (cs *CoherenceScorer) calculateAverageElixir(cards []CardCandidate) float64 {
	if len(cards) == 0 {
		return 0.0
	}

	total := 0
	for _, card := range cards {
		total += card.Elixir
	}
	return float64(total) / float64(len(cards))
}

// GetCoherenceScore returns just the coherence score for quick use
func (cs *CoherenceScorer) GetCoherenceScore(cards []CardCandidate, strategy Strategy) float64 {
	result := cs.AnalyzeCoherence(cards, strategy)
	return result.CoherenceScore
}
