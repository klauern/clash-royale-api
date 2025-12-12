// Package deck provides role classification for Clash Royale cards.
// Role classification assigns strategic purposes to cards for intelligent deck building.
package deck

import (
	"strings"
)

// Win condition cards - primary tower-damaging threats
var winConditions = map[string]bool{
	"Hog Rider":          true,
	"Royal Giant":        true,
	"Battle Ram":         true,
	"Goblin Barrel":      true,
	"Miner":              true,
	"Giant":              true,
	"Golem":              true,
	"Lava Hound":         true,
	"Balloon":            true,
	"X-Bow":              true,
	"Mortar":             true,
	"P.E.K.K.A":          true,
	"Mega Knight":        true,
	"Electro Giant":      true,
	"Royal Hogs":         true,
	"Ram Rider":          true,
	"Wall Breakers":      true,
	"Graveyard":          true,
	"Sparky":             true,
	"Three Musketeers":   true,
	"Giant Skeleton":     true,
	"Elixir Golem":       true,
}

// Defensive buildings - used for defense, kiting, or siege
var buildings = map[string]bool{
	"Cannon":            true,
	"Tesla":             true,
	"Inferno Tower":     true,
	"Bomb Tower":        true,
	"Goblin Cage":       true,
	"Furnace":           true,
	"Barbarian Hut":     true,
	"Goblin Hut":        true,
	"Tombstone":         true,
	"Elixir Collector":  true,
	"X-Bow":             true,  // Also win condition
	"Mortar":            true,  // Also win condition
}

// Big damage spells - 4+ elixir, high damage
var spellsBig = map[string]bool{
	"Fireball":          true,
	"Poison":            true,
	"Lightning":         true,
	"Rocket":            true,
	"Freeze":            true,
	"Earthquake":        true,
	"Graveyard":         true,  // Also win condition
	"Clone":             true,
	"Rage":              true,
}

// Small utility spells - 2-3 elixir, tactical
var spellsSmall = map[string]bool{
	"Zap":               true,
	"Log":               true,
	"Arrows":            true,
	"Snowball":          true,
	"Tornado":           true,
	"Barbarian Barrel":  true,
	"Giant Snowball":    true,
	"Heal Spirit":       true,
}

// Support troops - mid-cost troops for offense/defense
var supportTroops = map[string]bool{
	"Musketeer":         true,
	"Wizard":            true,
	"Witch":             true,
	"Baby Dragon":       true,
	"Electro Wizard":    true,
	"Ice Wizard":        true,
	"Night Witch":       true,
	"Executioner":       true,
	"Bowler":            true,
	"Dark Prince":       true,
	"Prince":            true,
	"Mini P.E.K.K.A":    true,
	"Lumberjack":        true,
	"Bandit":            true,
	"Magic Archer":      true,
	"Hunter":            true,
	"Skeleton Dragons":  true,
	"Mother Witch":      true,
	"Archer Queen":      true,
	"Golden Knight":     true,
	"Skeleton King":     true,
	"Mighty Miner":      true,
	"Monk":              true,
	"Little Prince":     true,
	"Archers":           true,
	"Knight":            true,
	"Valkyrie":          true,
	"Goblin Gang":       true,
	"Minions":           true,
	"Mega Minion":       true,
	"Guards":            true,
	"Skeleton Army":     true,
	"Goblin Barrel":     true,  // Also win condition
	"Battle Healer":     true,
	"Electro Dragon":    true,
	"Inferno Dragon":    true,
	"Royal Recruits":    true,
	"Cannon Cart":       true,
	"Fisherman":         true,
	"Firecracker":       true,
	"Rascals":           true,
	"Flying Machine":    true,
	"Zappies":           true,
	"Royal Delivery":    true,
	"Barbarians":        true,
	"Elite Barbarians":  true,
}

// Cycle cards - 1-2 elixir cheap troops for fast cycling
var cycleTroops = map[string]bool{
	"Skeletons":         true,
	"Ice Spirit":        true,
	"Fire Spirit":       true,
	"Heal Spirit":       true,
	"Electro Spirit":    true,
	"Spear Goblins":     true,
	"Goblins":           true,
	"Bats":              true,
	"Ice Golem":         true,
	"Larry":             true,  // Skeletons alternate name
}

// ClassifyCard determines the strategic role of a card based on its properties
// Returns a pointer to CardRole, or nil if the card doesn't fit a clear role
func ClassifyCard(cardName string, elixirCost int) *CardRole {
	// Normalize card name for matching (case-insensitive, trim whitespace)
	normalized := strings.TrimSpace(cardName)

	// Check win conditions first (highest priority)
	if winConditions[normalized] {
		role := RoleWinCondition
		return &role
	}

	// Check buildings
	if buildings[normalized] {
		role := RoleBuilding
		return &role
	}

	// Check big spells (4+ elixir)
	if spellsBig[normalized] {
		role := RoleSpellBig
		return &role
	}

	// Check small spells (2-3 elixir)
	if spellsSmall[normalized] {
		role := RoleSpellSmall
		return &role
	}

	// Check cycle cards (1-2 elixir)
	if cycleTroops[normalized] {
		role := RoleCycle
		return &role
	}

	// Check support troops
	if supportTroops[normalized] {
		role := RoleSupport
		return &role
	}

	// Fallback: classify by elixir cost if not in known lists
	if elixirCost <= 2 {
		role := RoleCycle
		return &role
	}

	if elixirCost >= 3 && elixirCost <= 5 {
		role := RoleSupport
		return &role
	}

	// High cost cards (6+) likely win conditions
	if elixirCost >= 6 {
		role := RoleWinCondition
		return &role
	}

	// Unknown card, no clear role
	return nil
}

// ClassifyCardCandidate assigns a role to a CardCandidate and updates its Role field
func ClassifyCardCandidate(candidate *CardCandidate) *CardRole {
	role := ClassifyCard(candidate.Name, candidate.Elixir)
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
	return winConditions[strings.TrimSpace(cardName)]
}

// IsBuilding returns true if the card is a building
func IsBuilding(cardName string) bool {
	return buildings[strings.TrimSpace(cardName)]
}

// IsSpell returns true if the card is any type of spell
func IsSpell(cardName string, elixirCost int) bool {
	normalized := strings.TrimSpace(cardName)
	return spellsBig[normalized] || spellsSmall[normalized]
}

// IsCycleCard returns true if the card is a cheap cycle card (1-2 elixir)
func IsCycleCard(cardName string, elixirCost int) bool {
	normalized := strings.TrimSpace(cardName)
	return cycleTroops[normalized] || elixirCost <= 2
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
