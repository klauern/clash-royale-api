package archetypes

import (
	"fmt"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
)

// ArchetypeBuilder builds decks constrained to specific archetypes.
// It wraps the base deck.Builder and applies archetype-specific filters
// before generating decks.
type ArchetypeBuilder struct {
	baseBuilder *deck.Builder
	constraints map[mulligan.Archetype]ArchetypeConstraints
}

// NewArchetypeBuilder creates a new archetype-aware deck builder
func NewArchetypeBuilder(dataDir string) *ArchetypeBuilder {
	return &ArchetypeBuilder{
		baseBuilder: deck.NewBuilder(dataDir),
		constraints: GetArchetypeConstraints(),
	}
}

// BuildForArchetype builds a deck matching the specified archetype characteristics.
// It filters the card collection based on archetype constraints, then uses the
// base builder to create an optimal deck from the filtered candidates.
func (ab *ArchetypeBuilder) BuildForArchetype(
	archetype mulligan.Archetype,
	analysis deck.CardAnalysis,
) (*deck.DeckRecommendation, error) {
	constraints, exists := ab.constraints[archetype]
	if !exists {
		return nil, fmt.Errorf("unknown archetype: %s", archetype)
	}

	// Filter candidates based on archetype constraints
	filteredAnalysis := ab.filterCandidates(analysis, constraints)

	// Use base builder to create deck from filtered candidates
	recommendation, err := ab.baseBuilder.BuildDeckFromAnalysis(filteredAnalysis)
	if err != nil {
		// Try with relaxed constraints if initial build fails
		relaxedAnalysis := ab.filterCandidatesRelaxed(analysis, constraints)
		recommendation, err = ab.baseBuilder.BuildDeckFromAnalysis(relaxedAnalysis)
		if err != nil {
			return nil, fmt.Errorf("failed to build %s deck: %w", archetype, err)
		}
	}

	// Validate that deck matches archetype characteristics
	if !ab.validateArchetype(recommendation, constraints) {
		// Deck doesn't perfectly match, but it's the best we can do
		// Add a note about the constraint mismatch
		note := fmt.Sprintf("Note: Limited collection - deck may not perfectly match %s archetype", archetype)
		recommendation.AddNote(note)
	}

	return recommendation, nil
}

// filterCandidates filters cards based on archetype constraints.
// It removes excluded cards from the analysis to guide deck building.
func (ab *ArchetypeBuilder) filterCandidates(
	analysis deck.CardAnalysis,
	constraints ArchetypeConstraints,
) deck.CardAnalysis {
	filtered := deck.CardAnalysis{
		CardLevels:   make(map[string]deck.CardLevelData),
		AnalysisTime: analysis.AnalysisTime,
	}

	// Create excluded cards map for fast lookup
	excludedMap := make(map[string]bool)
	for _, cardName := range constraints.ExcludedCards {
		excludedMap[cardName] = true
	}

	// Filter out excluded cards
	for cardName, cardData := range analysis.CardLevels {
		if !excludedMap[cardName] {
			filtered.CardLevels[cardName] = cardData
		}
	}

	return filtered
}

// filterCandidatesRelaxed applies more lenient filtering for limited collections.
// It only excludes cards that are strongly incompatible with the archetype.
func (ab *ArchetypeBuilder) filterCandidatesRelaxed(
	analysis deck.CardAnalysis,
	constraints ArchetypeConstraints,
) deck.CardAnalysis {
	// For relaxed mode, we only exclude the most incompatible cards
	// (e.g., don't put X-Bow in beatdown, don't put Golem in siege)
	relaxedExcluded := ab.getStronglyIncompatibleCards(constraints.Archetype)

	filtered := deck.CardAnalysis{
		CardLevels:   make(map[string]deck.CardLevelData),
		AnalysisTime: analysis.AnalysisTime,
	}

	// Create excluded cards map for fast lookup
	excludedMap := make(map[string]bool)
	for _, cardName := range relaxedExcluded {
		excludedMap[cardName] = true
	}

	// Filter out only strongly excluded cards
	for cardName, cardData := range analysis.CardLevels {
		if !excludedMap[cardName] {
			filtered.CardLevels[cardName] = cardData
		}
	}

	return filtered
}

// getStronglyIncompatibleCards returns cards that are fundamentally incompatible
// with an archetype (e.g., heavy tanks in cycle decks, siege weapons in beatdown)
func (ab *ArchetypeBuilder) getStronglyIncompatibleCards(archetype mulligan.Archetype) []string {
	incompatible := map[mulligan.Archetype][]string{
		mulligan.ArchetypeBeatdown: {"X-Bow", "Mortar"},
		mulligan.ArchetypeCycle:    {"Golem", "Lava Hound", "Electro Giant"},
		mulligan.ArchetypeSiege:    {"Golem", "Giant", "Lava Hound", "Mega Knight"},
		// Other archetypes are more flexible
	}

	if cards, exists := incompatible[archetype]; exists {
		return cards
	}
	return []string{}
}

// validateArchetype checks if a deck matches archetype characteristics.
// It validates elixir range and basic role requirements.
func (ab *ArchetypeBuilder) validateArchetype(
	rec *deck.DeckRecommendation,
	constraints ArchetypeConstraints,
) bool {
	// Check elixir range
	if rec.AvgElixir < constraints.MinElixir || rec.AvgElixir > constraints.MaxElixir {
		// Allow some tolerance (±0.3 elixir)
		tolerance := 0.3
		if rec.AvgElixir < constraints.MinElixir-tolerance ||
			rec.AvgElixir > constraints.MaxElixir+tolerance {
			return false
		}
	}

	// Check required role counts
	roleCounts := make(map[deck.CardRole]int)
	for _, card := range rec.DeckDetail {
		if card.Role != "" {
			role := deck.CardRole(card.Role)
			roleCounts[role]++
		}
	}

	// Validate minimum role requirements
	for role, minCount := range constraints.RequiredRoles {
		if roleCounts[role] < minCount {
			return false
		}
	}

	return true
}
