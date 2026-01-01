package analysis

import (
	"fmt"
	"math"
	"sort"
)

// recommendStrategiesForArchetype analyzes which deck builder strategies are compatible
// with a detected archetype and returns recommendations sorted by compatibility
func (d *DynamicArchetypeDetector) recommendStrategiesForArchetype(
	arch *DetectedArchetype,
	template DeckArchetypeTemplate,
) []StrategyRecommendation {
	// Get all strategies from the provider
	allStrategies := d.strategyProvider.GetAllStrategies()

	recommendations := []StrategyRecommendation{}

	for _, strategy := range allStrategies {
		config := d.strategyProvider.GetConfig(strategy)

		// Calculate archetype affinity score
		affinityScore := d.calculateStrategyAffinity(template, config)

		// Calculate elixir range fit
		elixirFit := d.calculateElixirFit(arch.AvgElixir, config)

		// Calculate overall compatibility (0-100)
		compatibility := (affinityScore * 0.7) + (elixirFit * 0.3)

		// Only include strategies with reasonable compatibility
		if compatibility >= 40 {
			reason := d.generateStrategyReason(template, strategy, compatibility)
			recommendations = append(recommendations, StrategyRecommendation{
				Strategy:           strategy,
				CompatibilityScore: compatibility,
				Reason:             reason,
				ArchetypeAffinity:  affinityScore,
			})
		}
	}

	// Sort by compatibility score descending
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].CompatibilityScore > recommendations[j].CompatibilityScore
	})

	return recommendations
}

// calculateStrategyAffinity calculates how well an archetype aligns with a strategy's
// preferred cards, based on the ArchetypeAffinity map in the strategy config
func (d *DynamicArchetypeDetector) calculateStrategyAffinity(
	template DeckArchetypeTemplate,
	config StrategyConfig,
) float64 {
	if config.ArchetypeAffinity == nil || len(config.ArchetypeAffinity) == 0 {
		return 50 // Neutral if no affinity defined
	}

	totalAffinity := 0.0
	count := 0

	// Check win condition affinity
	if bonus, exists := config.ArchetypeAffinity[template.WinCondition]; exists {
		totalAffinity += bonus * 100 // Convert 0.25 → 25
		count++
	}

	// Check support card affinities
	for _, supportCard := range template.SupportCards {
		if bonus, exists := config.ArchetypeAffinity[supportCard]; exists {
			totalAffinity += bonus * 100
			count++
		}
	}

	// Check required card affinities
	for _, requiredCard := range template.RequiredCards {
		if bonus, exists := config.ArchetypeAffinity[requiredCard]; exists {
			totalAffinity += bonus * 100
			count++
		}
	}

	if count == 0 {
		return 30 // Low affinity if no matches
	}

	// Average affinity (0-100 scale)
	avgAffinity := totalAffinity / float64(count)

	// Ensure within 0-100 range
	if avgAffinity > 100 {
		avgAffinity = 100
	}

	return avgAffinity
}

// calculateElixirFit calculates how well the archetype's elixir cost fits the strategy's
// target elixir range (0-100 scale)
func (d *DynamicArchetypeDetector) calculateElixirFit(
	archetypeElixir float64,
	config StrategyConfig,
) float64 {
	// Handle case where archetype has no elixir data
	if archetypeElixir == 0 {
		return 50 // Neutral
	}

	minElixir := config.TargetElixirMin
	maxElixir := config.TargetElixirMax

	// Perfect fit: within range
	if archetypeElixir >= minElixir && archetypeElixir <= maxElixir {
		// Score based on how centered it is within range
		rangeCenter := (minElixir + maxElixir) / 2
		distanceFromCenter := math.Abs(archetypeElixir - rangeCenter)
		rangeSize := maxElixir - minElixir

		// Closer to center = higher score
		centeredness := 1.0 - (distanceFromCenter / (rangeSize / 2))
		return 80 + (centeredness * 20) // 80-100
	}

	// Outside range: penalize based on distance
	var distance float64
	if archetypeElixir < minElixir {
		distance = minElixir - archetypeElixir
	} else {
		distance = archetypeElixir - maxElixir
	}

	// Penalty: -15 points per 0.5 elixir away, minimum 20
	penalty := distance * 30 // 0.5 elixir = 15 points
	fit := 80 - penalty

	if fit < 20 {
		fit = 20 // Minimum 20% fit
	}

	return fit
}

// generateStrategyReason creates a human-readable explanation for why a strategy
// is recommended for an archetype
func (d *DynamicArchetypeDetector) generateStrategyReason(
	template DeckArchetypeTemplate,
	strategy Strategy,
	compatibility float64,
) string {
	config := d.strategyProvider.GetConfig(strategy)

	// Check which cards have affinity
	affinityCards := []string{}
	if _, exists := config.ArchetypeAffinity[template.WinCondition]; exists {
		affinityCards = append(affinityCards, template.WinCondition)
	}

	for _, card := range template.SupportCards {
		if _, exists := config.ArchetypeAffinity[card]; exists {
			affinityCards = append(affinityCards, card)
			if len(affinityCards) >= 3 {
				break // Limit to 3 examples
			}
		}
	}

	// Generate reason based on compatibility level and affinity
	var reason string

	switch {
	case compatibility >= 85:
		reason = "Excellent match"
	case compatibility >= 70:
		reason = "Strong compatibility"
	case compatibility >= 55:
		reason = "Good fit"
	default:
		reason = "Moderate compatibility"
	}

	// Add specific details
	if len(affinityCards) > 0 {
		reason += fmt.Sprintf(": %s naturally fits %s strategy", template.WinCondition, strategy)
	}

	// Add elixir range info
	if template.MinElixir > 0 && template.MaxElixir > 0 {
		avgElixir := (template.MinElixir + template.MaxElixir) / 2
		if avgElixir >= config.TargetElixirMin && avgElixir <= config.TargetElixirMax {
			reason += fmt.Sprintf(" (%.1f elixir matches target range)", avgElixir)
		}
	}

	return reason
}

// AddStrategyRecommendations enhances a DynamicArchetypeAnalysis with strategy recommendations
func (d *DynamicArchetypeDetector) AddStrategyRecommendations(analysis *DynamicArchetypeAnalysis) {
	for i := range analysis.DetectedArchetypes {
		arch := &analysis.DetectedArchetypes[i]

		// Find the corresponding template
		var template *DeckArchetypeTemplate
		for j := range d.archetypes {
			if d.archetypes[j].Name == arch.Name {
				template = &d.archetypes[j]
				break
			}
		}

		if template != nil {
			arch.RecommendedStrategies = d.recommendStrategiesForArchetype(arch, *template)
		}
	}
}
