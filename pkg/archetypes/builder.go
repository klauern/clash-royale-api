package archetypes

import (
	"fmt"
	"strings"

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
// It uses archetype-specific deck composition and prioritizes preferred cards.
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

	// Build deck using archetype-specific composition
	recommendation, err := ab.buildArchetypeSpecificDeck(archetype, filteredAnalysis, constraints)
	if err != nil {
		// Try with relaxed constraints if initial build fails
		relaxedAnalysis := ab.filterCandidatesRelaxed(analysis, constraints)
		recommendation, err = ab.buildArchetypeSpecificDeck(archetype, relaxedAnalysis, constraints)
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

// filterCandidatesWithPreference filters cards and boosts preferred cards for archetype.
func (ab *ArchetypeBuilder) filterCandidatesWithPreference(
	analysis deck.CardAnalysis,
	constraints ArchetypeConstraints,
) deck.CardAnalysis {
	filtered := deck.CardAnalysis{
		CardLevels:   make(map[string]deck.CardLevelData),
		AnalysisTime: analysis.AnalysisTime,
	}

	// Create excluded cards and preferred cards maps for fast lookup
	excludedMap := make(map[string]bool)
	for _, cardName := range constraints.ExcludedCards {
		excludedMap[cardName] = true
	}

	preferredMap := make(map[string]bool)
	for _, cardName := range constraints.PreferredCards {
		preferredMap[cardName] = true
	}

	// Filter out excluded cards and boost preferred cards
	for cardName, cardData := range analysis.CardLevels {
		if !excludedMap[cardName] {
			// Boost score for preferred cards to increase selection priority
			// (using ScoreBoost instead of Level to preserve correct display values)
			if preferredMap[cardName] {
				cardData.ScoreBoost = 0.2 // 20% score boost for archetype-preferred cards
			}
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

// buildArchetypeSpecificDeck builds a deck using archetype-specific composition templates.
func (ab *ArchetypeBuilder) buildArchetypeSpecificDeck(
	_ mulligan.Archetype, // archetype info is embedded in constraints
	analysis deck.CardAnalysis,
	constraints ArchetypeConstraints,
) (*deck.DeckRecommendation, error) {
	// Apply preferred card boosting
	weightedAnalysis := ab.filterCandidatesWithPreference(analysis, constraints)

	// Use base builder for initial deck
	recommendation, err := ab.baseBuilder.BuildDeckFromAnalysis(weightedAnalysis)
	if err != nil {
		return nil, err
	}

	// Adjust deck composition based on archetype requirements
	recommendation = ab.adjustDeckForArchetype(recommendation, constraints)

	return recommendation, nil
}

// adjustDeckForArchetype modifies the deck to better match archetype characteristics.
func (ab *ArchetypeBuilder) adjustDeckForArchetype(
	recommendation *deck.DeckRecommendation,
	constraints ArchetypeConstraints,
) *deck.DeckRecommendation {
	// Check elixir range and suggest adjustments if needed
	if recommendation.AvgElixir < constraints.MinElixir {
		recommendation.AddNote(fmt.Sprintf("Low elixir for %s (target: %.1f-%.1f)",
			constraints.Archetype, constraints.MinElixir, constraints.MaxElixir))
	} else if recommendation.AvgElixir > constraints.MaxElixir {
		recommendation.AddNote(fmt.Sprintf("High elixir for %s (target: %.1f-%.1f)",
			constraints.Archetype, constraints.MinElixir, constraints.MaxElixir))
	}

	// Add archetype-specific notes
	if len(constraints.PreferredCards) > 0 {
		maxShow := min(3, len(constraints.PreferredCards))
		recommendation.AddNote(fmt.Sprintf("Preferred for %s: %s",
			constraints.Archetype, strings.Join(constraints.PreferredCards[:maxShow], ", ")))
	}

	return recommendation
}

// SetUnlockedEvolutions updates the unlocked evolutions list for the base builder.
func (ab *ArchetypeBuilder) SetUnlockedEvolutions(cards []string) {
	ab.baseBuilder.SetUnlockedEvolutions(cards)
}

// SetEvolutionSlotLimit updates the evolution slot limit for the base builder.
func (ab *ArchetypeBuilder) SetEvolutionSlotLimit(limit int) {
	ab.baseBuilder.SetEvolutionSlotLimit(limit)
}
