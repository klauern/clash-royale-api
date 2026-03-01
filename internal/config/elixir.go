package config

// Elixir-related constants for deck building and scoring

import "slices"

const (
	// ElixirOptimal is the optimal elixir cost for balanced deck composition
	// Cards around this cost are generally most flexible and efficient
	ElixirOptimal = 3.0

	// ElixirWeightFactor determines the weight of elixir efficiency in card scoring
	// Higher values make elixir cost more influential in card selection
	ElixirWeightFactor = 0.15

	// ElixirMaxDiff is the maximum meaningful difference from optimal elixir
	// Used to normalize elixir weight calculations (0-9 elixir range)
	ElixirMaxDiff = 9.0

	// ElixirCyclePenaltyThreshold is the elixir cost above which cycle decks heavily penalize cards
	// Cycle decks target low average elixir (≤3.0) and need to avoid expensive cards
	ElixirCyclePenaltyThreshold = 4.0
)

// fallbackElixir provides elixir costs for cards when the API doesn't return them.
// This is a comprehensive mapping of common cards to their elixir costs.
// Note: The actual elixir cost should come from the API when available.
var fallbackElixir = map[string]int{
	// Win Conditions (3-7 elixir)
	"Royal Giant":    6,
	"Hog Rider":      4,
	"Giant":          5,
	"Golem":          8,
	"P.E.K.K.A":      7,
	"Giant Skeleton": 6,
	"Goblin Barrel":  3,
	"Mortar":         4,
	"X-Bow":          6,
	"Royal Hogs":     5,
	"Rune Giant":     6,
	"Goblin Giant":   6,

	// Buildings (3-6 elixir)
	"Cannon":        3,
	"Goblin Cage":   4,
	"Inferno Tower": 5,
	"Bomb Tower":    4,
	"Tombstone":     3,
	"Goblin Hut":    5,
	"Barbarian Hut": 6,

	// Big Spells (4-6 elixir)
	"Fireball":  4,
	"Poison":    4,
	"Lightning": 6,
	"Rocket":    6,

	// Small Spells (2-4 elixir)
	"Zap":              2,
	"Arrows":           3,
	"Giant Snowball":   2,
	"Barbarian Barrel": 2,
	"Freeze":           4,
	"Log":              2,
	"The Log":          2,
	"Mirror":           1,
	"Vines":            2,

	// Support (2-5 elixir)
	"Archers":           3,
	"Bomber":            2,
	"Musketeer":         4,
	"Wizard":            5,
	"Mega Minion":       3,
	"Valkyrie":          4,
	"Baby Dragon":       4,
	"Skeleton Dragons":  4,
	"Berserker":         4,
	"Dart Goblin":       3,
	"Goblin Demolisher": 4,
	"Minion Horde":      5,
	"Phoenix":           4,
	"Royal Ghost":       3,
	"Skeleton Barrel":   3,

	// Cycle (1-3 elixir)
	"Knight":         3,
	"Skeletons":      1,
	"Ice Spirit":     1,
	"Electro Spirit": 1,
	"Fire Spirit":    1,
	"Bats":           2,
	"Spear Goblins":  2,
	"Goblin Gang":    3,
	"Minions":        3,
	"Ice Golem":      2,
}

// CardRole represents the strategic role of a card in a deck (string type)
type CardRole string

// String returns the string representation of the card role
func (cr CardRole) String() string {
	return string(cr)
}

const (
	// RoleWinCondition represents primary tower-damaging cards
	RoleWinCondition CardRole = "win_conditions"
	// RoleBuilding represents defensive buildings
	RoleBuilding CardRole = "buildings"
	// RoleSpellBig represents high-elixir damage spells
	RoleSpellBig CardRole = "spells_big"
	// RoleSpellSmall represents low-elixir utility spells
	RoleSpellSmall CardRole = "spells_small"
	// RoleSupport represents mid-cost support troops
	RoleSupport CardRole = "support"
	// RoleCycle represents low-cost cycle cards
	RoleCycle CardRole = "cycle"
)

// roleGroups maps card roles to their representative cards.
// This is used to classify cards by their strategic function in a deck.
// Migrated from pkg/deck/role_classifier.go to provide comprehensive coverage.
var roleGroups = map[CardRole][]string{
	RoleWinCondition: {
		"Hog Rider", "Royal Giant", "Battle Ram", "Goblin Barrel", "Miner",
		"Giant", "Golem", "Lava Hound", "Balloon", "X-Bow", "Mortar",
		"P.E.K.K.A", "Mega Knight", "Electro Giant", "Royal Hogs", "Ram Rider",
		"Wall Breakers", "Graveyard", "Sparky", "Three Musketeers",
		"Giant Skeleton", "Elixir Golem",
		"Rune Giant", "Goblin Giant",
	},
	RoleBuilding: {
		"Cannon", "Tesla", "Inferno Tower", "Bomb Tower", "Goblin Cage",
		"Furnace", "Barbarian Hut", "Goblin Hut", "Tombstone",
		"Elixir Collector",
		// Note: X-Bow and Mortar also in RoleWinCondition (dual role)
	},
	RoleSpellBig: {
		"Fireball", "Poison", "Lightning", "Rocket", "Freeze", "Earthquake",
		"Clone", "Rage",
		// Note: Graveyard also in RoleWinCondition (dual role)
	},
	RoleSpellSmall: {
		"Zap", "Log", "Arrows", "Snowball", "Tornado", "Barbarian Barrel",
		"Giant Snowball", "Heal Spirit",
		"Mirror", "Vines",
	},
	RoleSupport: {
		"Musketeer", "Wizard", "Witch", "Baby Dragon", "Electro Wizard",
		"Ice Wizard", "Night Witch", "Executioner", "Bowler", "Dark Prince",
		"Prince", "Mini P.E.K.K.A", "Lumberjack", "Bandit", "Magic Archer",
		"Hunter", "Skeleton Dragons", "Mother Witch",
		// Champions
		"Archer Queen", "Golden Knight", "Skeleton King", "Mighty Miner",
		"Monk", "Little Prince",
		// Common support troops
		"Princess", "Archers", "Knight", "Valkyrie", "Goblin Gang", "Minions",
		"Mega Minion", "Guards", "Skeleton Army",
		// Additional support
		"Battle Healer", "Electro Dragon", "Inferno Dragon", "Royal Recruits",
		"Cannon Cart", "Fisherman", "Firecracker", "Rascals", "Flying Machine",
		"Zappies", "Royal Delivery", "Barbarians", "Elite Barbarians",
		// Note: Goblin Barrel also in RoleWinCondition (dual role)
		"Berserker", "Dart Goblin", "Goblin Demolisher", "Minion Horde",
		"Phoenix", "Royal Ghost", "Skeleton Barrel",
	},
	RoleCycle: {
		"Skeletons", "Ice Spirit", "Fire Spirit", "Heal Spirit",
		"Electro Spirit", "Spear Goblins", "Goblins", "Bats", "Ice Golem",
		// Note: Knight, Valkyrie, Goblin Gang, Minions moved to RoleSupport
		"Bomber",
	},
}

// evolutionRoleOverrides defines cards whose role changes when evolved.
// When a card is evolved, check this map first before using roleGroups.
var evolutionRoleOverrides = map[string]CardRole{
	"Valkyrie":    RoleSupport,      // Evolved: whirlwind pull makes it a control support card
	"Knight":      RoleSupport,      // Evolved: clone on death increases defensive value
	"Royal Giant": RoleWinCondition, // Evolved: anti-pushback improves win condition reliability
	"Barbarian":   RoleSupport,      // Evolved: 3 spawned barbarians act as support swarm
	"Witch":       RoleSupport,      // Evolved: faster skeleton spawn enhances support
	"Golem":       RoleWinCondition, // Evolved: golemites spawn on death strengthens win condition
}

var roleDescriptions = map[CardRole]string{
	RoleWinCondition: "Primary tower-damaging threat",
	RoleBuilding:     "Defensive building or siege structure",
	RoleSpellBig:     "High-damage spell (4+ elixir)",
	RoleSpellSmall:   "Utility spell (2-3 elixir)",
	RoleSupport:      "Mid-cost support troop",
	RoleCycle:        "Cheap cycle card (1-2 elixir)",
}

// GetCardElixir returns the elixir cost for a card.
// It first checks the API-provided elixir cost, then falls back to the static mapping.
// Returns 4 (default fallback) if the card is not found in either source.
func GetCardElixir(cardName string, apiElixir int) int {
	// If API provides elixir cost, use it
	if apiElixir > 0 {
		return apiElixir
	}

	// Fall back to static mapping
	if cost, exists := fallbackElixir[cardName]; exists {
		return cost
	}

	// Default fallback for unknown cards
	return 4
}

// GetCardRole returns the role for a given card name.
// Returns empty CardRole ("") if the card is not found in any role group.
// For evolution-aware role classification, use GetCardRoleWithEvolution.
func GetCardRole(cardName string) CardRole {
	return GetCardRoleWithEvolution(cardName, 0)
}

// GetCardRoleWithEvolution returns the role for a given card name, considering evolution level.
// When evolutionLevel > 0, checks evolutionRoleOverrides first before roleGroups.
// Returns empty CardRole ("") if the card is not found in any role group.
func GetCardRoleWithEvolution(cardName string, evolutionLevel int) CardRole {
	// Check evolution overrides first if evolved
	if evolutionLevel > 0 {
		if role, exists := evolutionRoleOverrides[cardName]; exists {
			return role
		}
	}

	// Check standard role groups
	for role, cards := range roleGroups {
		if slices.Contains(cards, cardName) {
			return role
		}
	}
	return ""
}

// GetRoleCards returns the list of cards for a given role.
// Returns nil if the role doesn't exist.
func GetRoleCards(role CardRole) []string {
	if cards, exists := roleGroups[role]; exists {
		return cards
	}
	return nil
}

// GetRoleDescription returns a human-readable description for a card role.
// Unknown roles return "Unknown role".
func GetRoleDescription(role CardRole) string {
	if desc, exists := roleDescriptions[role]; exists {
		return desc
	}
	return "Unknown role"
}

// HasEvolutionOverride returns true if the card has an evolution role override defined.
// This checks if the card exists in the evolutionRoleOverrides map, regardless of
// whether the override role differs from the base role.
func HasEvolutionOverride(cardName string) bool {
	_, exists := evolutionRoleOverrides[cardName]
	return exists
}
