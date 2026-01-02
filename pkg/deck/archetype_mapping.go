// Package deck provides archetype-to-strategy mapping functionality.
package deck

import (
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
)

// GetStrategyForArchetype returns the recommended deck builder strategy
// for a given archetype template.
// If the archetype has an explicit preferred strategy, it uses that.
// Otherwise, it infers the strategy from archetype characteristics.
func GetStrategyForArchetype(archetype analysis.DeckArchetypeTemplate) Strategy {
	// If archetype has explicit preferred strategy, use it
	if archetype.PreferredStrategy != "" {
		return Strategy(archetype.PreferredStrategy)
	}

	// Otherwise, infer from archetype characteristics
	return inferStrategyFromArchetype(archetype)
}

// inferStrategyFromArchetype uses heuristics to map archetype to strategy
// based on average elixir cost and win condition type.
func inferStrategyFromArchetype(archetype analysis.DeckArchetypeTemplate) Strategy {
	avgElixir := (archetype.MinElixir + archetype.MaxElixir) / 2

	// Cycle strategy: low elixir (< 3.0)
	if avgElixir < 3.0 {
		return StrategyCycle
	}

	// Control strategy: siege win conditions (X-Bow, Mortar, Graveyard)
	controlWinCons := []string{"X-Bow", "Mortar", "Graveyard"}
	for _, wc := range controlWinCons {
		if strings.Contains(archetype.WinCondition, wc) {
			return StrategyControl
		}
	}

	// Aggro strategy: beatdown win conditions (Golem, Lava Hound, Electro Giant, Giant)
	aggroWinCons := []string{"Golem", "Lava Hound", "Electro Giant", "Giant"}
	for _, wc := range aggroWinCons {
		if strings.Contains(archetype.WinCondition, wc) {
			return StrategyAggro
		}
	}

	// Default to balanced for everything else
	return StrategyBalanced
}

// GetArchetypesForStrategy returns all archetypes that prefer the given strategy.
// This provides reverse mapping from strategy to archetypes.
func GetArchetypesForStrategy(
	strategy Strategy,
	archetypes []analysis.DeckArchetypeTemplate,
) []analysis.DeckArchetypeTemplate {
	var matches []analysis.DeckArchetypeTemplate
	for _, archetype := range archetypes {
		if GetStrategyForArchetype(archetype) == strategy {
			matches = append(matches, archetype)
		}
	}
	return matches
}
