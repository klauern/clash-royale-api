// Package deck provides role classification for Clash Royale cards.
// Role classification assigns strategic purposes to cards for intelligent deck building.
//
// DEPRECATED: This package now wraps the single source of truth in internal/config/elixir.go
// to eliminate duplication. The role data (winConditions, buildings, spells, etc.) is now
// maintained in internal/config. New code should use config.CardRole and config.GetCardRole().
package deck

import (
	"github.com/klauer/clash-royale-api/go/internal/config"
)

// ClassifyCard determines the strategic role of a card based on its properties
// Returns a pointer to CardRole, or nil if the card doesn't fit a clear role
func ClassifyCard(cardName string, elixirCost int) *CardRole {
	return ClassifyCardWithEvolution(cardName, elixirCost, 0)
}

// ClassifyCardWithEvolution determines the strategic role of a card considering
// both its base properties and evolution status. When a card is evolved and has
// a specific evolution role override, that role takes precedence over base classification.
func ClassifyCardWithEvolution(cardName string, elixirCost, evolutionLevel int) *CardRole {
	// Use the single source of truth from config
	role := config.GetCardRoleWithEvolution(cardName, evolutionLevel)

	// Fallback: classify by elixir cost if not in config database
	if role == "" {
		if elixirCost <= 2 {
			role = RoleCycle
		} else if elixirCost >= 3 && elixirCost <= 5 {
			role = RoleSupport
		} else if elixirCost >= 6 {
			role = RoleWinCondition
		} else {
			// Unknown elixir cost, no clear role
			return nil
		}
		return &role
	}

	return &role
}

// ClassifyCardCandidate assigns a role to a CardCandidate and updates its Role field.
// Considers the card's evolution level when determining its role.
func ClassifyCardCandidate(candidate *CardCandidate) *CardRole {
	role := ClassifyCardWithEvolution(candidate.Name, candidate.Elixir, candidate.EvolutionLevel)
	candidate.Role = role
	return role
}

// ClassifyAllCandidates assigns roles to all candidates in a slice
func ClassifyAllCandidates(candidates []CardCandidate) {
	for i := range candidates {
		ClassifyCardCandidate(&candidates[i])
	}
}

// IsWinCondition returns true if the card is classified as a win condition
func IsWinCondition(cardName string) bool {
	return config.GetCardRole(cardName) == RoleWinCondition
}

// IsBuilding returns true if the card is a building
func IsBuilding(cardName string) bool {
	return config.GetCardRole(cardName) == RoleBuilding
}

// IsSpell returns true if the card is any type of spell
func IsSpell(cardName string, elixirCost int) bool {
	role := config.GetCardRole(cardName)
	return role == RoleSpellBig || role == RoleSpellSmall
}

// IsCycleCard returns true if the card is a cheap cycle card (1-2 elixir)
func IsCycleCard(cardName string, elixirCost int) bool {
	role := config.GetCardRole(cardName)
	return role == RoleCycle || elixirCost <= 2
}

// HasEvolutionOverride returns true if the card has a special role override when evolved
func HasEvolutionOverride(cardName string, evolutionLevel int) bool {
	if evolutionLevel <= 0 {
		return false
	}
	return config.HasEvolutionOverride(cardName)
}

// GetEvolutionOverrideRole returns the override role for an evolved card, or nil if none
func GetEvolutionOverrideRole(cardName string, evolutionLevel int) *CardRole {
	if evolutionLevel <= 0 {
		return nil
	}
	// Check if the card has an evolution override defined in config
	if !config.HasEvolutionOverride(cardName) {
		return nil
	}
	// Get the evolved role (which includes evolution overrides)
	evolvedRole := config.GetCardRoleWithEvolution(cardName, evolutionLevel)
	if evolvedRole == "" {
		return nil
	}
	return &evolvedRole
}

// GetRoleDescription returns a human-readable description of a card role
func GetRoleDescription(role CardRole) string {
	descriptions := map[CardRole]string{
		RoleWinCondition: "Primary tower-damaging threat",
		RoleBuilding:     "Defensive building or siege structure",
		RoleSpellBig:     "High-damage spell (4+ elixir)",
		RoleSpellSmall:   "Utility spell (2-3 elixir)",
		RoleSupport:      "Mid-cost support troop",
		RoleCycle:        "Cheap cycle card (1-2 elixir)",
	}

	if desc, exists := descriptions[role]; exists {
		return desc
	}

	return "Unknown role"
}

// CountRoles returns a map of role counts from a slice of candidates
func CountRoles(candidates []CardCandidate) map[CardRole]int {
	counts := make(map[CardRole]int)

	for _, candidate := range candidates {
		if candidate.Role != nil {
			counts[*candidate.Role]++
		}
	}

	return counts
}

// HasBalancedRoles checks if a deck has balanced role distribution
// A balanced deck should have: 1-2 win conditions, 1 building, 2 spells, support, and cycle
func HasBalancedRoles(candidates []CardCandidate) bool {
	counts := CountRoles(candidates)

	// Must have at least 1 win condition
	if counts[RoleWinCondition] < 1 {
		return false
	}

	// Should have exactly 1 small spell
	if counts[RoleSpellSmall] != 1 {
		return false
	}

	// Should have at least 1 big spell or 2 total spells
	totalSpells := counts[RoleSpellBig] + counts[RoleSpellSmall]
	if totalSpells < 2 {
		return false
	}

	// Should have some cycle cards (at least 1-2)
	if counts[RoleCycle] < 1 {
		return false
	}

	return true
}
