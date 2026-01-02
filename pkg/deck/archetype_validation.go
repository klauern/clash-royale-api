// Package deck provides validation for archetype-strategy alignment.
package deck

import (
	"fmt"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
)

// ValidationWarning represents an archetype-strategy alignment issue
type ValidationWarning struct {
	Archetype string
	Strategy  string
	Issue     string
}

// ValidateArchetypeStrategyAlignment checks if strategy affinity maps
// align with archetype preferred strategies.
// Returns a list of warnings for misalignments.
func ValidateArchetypeStrategyAlignment(
	archetypes []analysis.DeckArchetypeTemplate,
	strategies map[Strategy]StrategyConfig,
) []ValidationWarning {
	var warnings []ValidationWarning

	for _, archetype := range archetypes {
		strategy := GetStrategyForArchetype(archetype)
		config, exists := strategies[strategy]

		if !exists {
			warnings = append(warnings, ValidationWarning{
				Archetype: archetype.Name,
				Strategy:  string(strategy),
				Issue:     fmt.Sprintf("Strategy '%s' not found in configuration", strategy),
			})
			continue
		}

		// Check if win condition appears in strategy's affinity map
		if config.ArchetypeAffinity != nil {
			if _, exists := config.ArchetypeAffinity[archetype.WinCondition]; !exists {
				warnings = append(warnings, ValidationWarning{
					Archetype: archetype.Name,
					Strategy:  string(strategy),
					Issue:     fmt.Sprintf("Win condition '%s' not in affinity map", archetype.WinCondition),
				})
			}
		}

		// Check if any support cards appear in strategy's affinity map
		if config.ArchetypeAffinity != nil && len(archetype.SupportCards) > 0 {
			foundSupport := false
			for _, supportCard := range archetype.SupportCards {
				if _, exists := config.ArchetypeAffinity[supportCard]; exists {
					foundSupport = true
					break
				}
			}
			if !foundSupport {
				warnings = append(warnings, ValidationWarning{
					Archetype: archetype.Name,
					Strategy:  string(strategy),
					Issue:     "No support cards found in strategy affinity map",
				})
			}
		}
	}

	return warnings
}
