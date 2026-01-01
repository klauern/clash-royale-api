package config

// Elixir-related constants for deck building and scoring

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
	"Royal Giant":     6,
	"Hog Rider":       4,
	"Giant":           5,
	"P.E.K.K.A":       7,
	"Giant Skeleton":  6,
	"Goblin Barrel":   3,
	"Mortar":          4,
	"X-Bow":           6,
	"Royal Hogs":      5,

	// Buildings (3-6 elixir)
	"Cannon":         3,
	"Goblin Cage":    4,
	"Inferno Tower":  5,
	"Bomb Tower":     4,
	"Tombstone":      3,
	"Goblin Hut":     5,
	"Barbarian Hut":  6,

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

	// Support (2-5 elixir)
	"Archers":          3,
	"Bomber":           2,
	"Musketeer":        4,
	"Wizard":           5,
	"Mega Minion":      3,
	"Valkyrie":         4,
	"Baby Dragon":      4,
	"Skeleton Dragons": 4,

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
}

// CardRole represents the strategic role of a card in a deck (string type)
type CardRole string

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
var roleGroups = map[CardRole][]string{
	RoleWinCondition: {
		"Royal Giant", "Hog Rider", "Giant", "P.E.K.K.A", "Giant Skeleton",
		"Goblin Barrel", "Mortar", "X-Bow", "Royal Hogs",
	},
	RoleBuilding: {
		"Cannon", "Goblin Cage", "Inferno Tower", "Bomb Tower", "Tombstone",
		"Goblin Hut", "Barbarian Hut",
	},
	RoleSpellBig: {
		"Fireball", "Poison", "Lightning", "Rocket",
	},
	RoleSpellSmall: {
		"Zap", "Arrows", "Giant Snowball", "Barbarian Barrel",
		"Freeze", "Log",
	},
	RoleSupport: {
		"Archers", "Bomber", "Musketeer", "Wizard", "Mega Minion",
		"Valkyrie", "Baby Dragon", "Skeleton Dragons",
	},
	RoleCycle: {
		"Knight", "Skeletons", "Ice Spirit", "Electro Spirit",
		"Fire Spirit", "Bats", "Spear Goblins", "Goblin Gang", "Minions",
	},
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
func GetCardRole(cardName string) CardRole {
	for role, cards := range roleGroups {
		for _, card := range cards {
			if card == cardName {
				return role
			}
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
