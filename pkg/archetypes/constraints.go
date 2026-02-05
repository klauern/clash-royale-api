package archetypes

import (
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
)

// Package-level archetype data for reduced function complexity
var (
	beatdownRoles = map[deck.CardRole]int{
		deck.RoleWinCondition: 1, // Must have heavy tank
		deck.RoleSupport:      2, // Support troops for tank
	}
	beatdownPreferred = []string{
		"Golem", "Giant", "Lava Hound", "Electro Giant",
		"Baby Dragon", "Night Witch", "Mega Minion", "Lumberjack",
		"Lightning", "Tornado", "Arrows",
	}
	beatdownExcluded = []string{
		"X-Bow", "Mortar", "Hog Rider", "Miner", "Goblin Barrel",
	}

	cycleRoles = map[deck.CardRole]int{
		deck.RoleWinCondition: 1,
		deck.RoleCycle:        3, // Many cheap cards for fast cycling
		deck.RoleSpellSmall:   1,
	}
	cyclePreferred = []string{
		"Hog Rider", "Miner", "Skeletons", "Ice Spirit",
		"Ice Golem", "Cannon", "Musketeer", "Log",
		"Fireball", "Electro Spirit", "Bats",
	}
	cycleExcluded = []string{
		"Golem", "Lava Hound", "Giant", "Electro Giant",
		"P.E.K.K.A", "Mega Knight",
	}

	controlRoles = map[deck.CardRole]int{
		deck.RoleBuilding:   1, // Defensive building
		deck.RoleSpellBig:   1, // Big spell for control
		deck.RoleSpellSmall: 1,
	}
	controlPreferred = []string{
		"Inferno Tower", "Cannon", "Bomb Tower", "Tesla",
		"Valkyrie", "Wizard", "Musketeer", "Archers",
		"Fireball", "Poison", "Log", "Arrows",
	}
	controlExcluded = []string{
		"Golem", "Giant", "Lava Hound",
	}

	siegeRoles = map[deck.CardRole]int{
		deck.RoleWinCondition: 1, // Must have X-Bow or Mortar
		deck.RoleBuilding:     1, // Additional defensive building
		deck.RoleCycle:        2, // Cycle cards
	}
	siegePreferred = []string{
		"X-Bow", "Mortar", "Tesla", "Cannon",
		"Archers", "Skeletons", "Ice Spirit", "Knight",
		"Fireball", "Log", "Rocket",
	}
	siegeExcluded = []string{
		"Golem", "Giant", "Lava Hound", "Hog Rider",
		"Mega Knight", "P.E.K.K.A",
	}

	bridgeSpamRoles = map[deck.CardRole]int{
		deck.RoleWinCondition: 1,
		deck.RoleSupport:      2,
	}
	bridgeSpamPreferred = []string{
		"Battle Ram", "Hog Rider", "Bandit", "Royal Ghost",
		"Electro Wizard", "Magic Archer", "Dark Prince",
		"Poison", "Fireball", "Zap", "Log",
	}
	bridgeSpamExcluded = []string{
		"Golem", "Lava Hound", "Giant", "X-Bow", "Mortar",
	}

	midrangeRoles = map[deck.CardRole]int{
		deck.RoleWinCondition: 1,
		deck.RoleSpellBig:     1,
		deck.RoleSpellSmall:   1,
	}
	midrangePreferred = []string{
		"Hog Rider", "Royal Giant", "Royal Hogs", "Valkyrie",
		"Musketeer", "Mega Minion", "Knight", "Archers",
		"Fireball", "Zap", "Log", "Arrows",
	}

	spawndeckRoles = map[deck.CardRole]int{
		deck.RoleBuilding: 2, // Spawn buildings
	}
	spawndeckPreferred = []string{
		"Goblin Hut", "Barbarian Hut", "Furnace", "Tombstone",
		"Graveyard", "Poison", "Fireball", "Zap",
		"Valkyrie", "Baby Dragon", "Mega Minion",
	}
	spawndeckExcluded = []string{
		"X-Bow", "Mortar", "Hog Rider",
	}

	baitRoles = map[deck.CardRole]int{
		deck.RoleWinCondition: 1,
		deck.RoleSpellSmall:   1,
	}
	baitPreferred = []string{
		"Goblin Barrel", "Princess", "Goblin Gang", "Skeleton Army",
		"Rocket", "Log", "Zap", "Arrows",
		"Knight", "Ice Spirit", "Inferno Tower",
	}
	baitExcluded = []string{
		"Golem", "Giant", "Lava Hound", "X-Bow", "Mortar",
	}
)

// GetArchetypeConstraints returns predefined constraints for all 8 archetypes.
// These constraints guide deck building to match specific playstyle characteristics.
func GetArchetypeConstraints() map[mulligan.Archetype]ArchetypeConstraints {
	return map[mulligan.Archetype]ArchetypeConstraints{
		mulligan.ArchetypeBeatdown: {
			Archetype:      mulligan.ArchetypeBeatdown,
			MinElixir:      4.0,
			MaxElixir:      5.5,
			RequiredRoles:  beatdownRoles,
			PreferredCards: beatdownPreferred,
			ExcludedCards:  beatdownExcluded,
			Description:    "Heavy tank-based deck with support troops for big pushes",
		},
		mulligan.ArchetypeCycle: {
			Archetype:      mulligan.ArchetypeCycle,
			MinElixir:      2.5,
			MaxElixir:      3.5,
			RequiredRoles:  cycleRoles,
			PreferredCards: cyclePreferred,
			ExcludedCards:  cycleExcluded,
			Description:    "Fast cycling deck with low elixir cards for constant pressure",
		},
		mulligan.ArchetypeControl: {
			Archetype:      mulligan.ArchetypeControl,
			MinElixir:      3.5,
			MaxElixir:      4.5,
			RequiredRoles:  controlRoles,
			PreferredCards: controlPreferred,
			ExcludedCards:  controlExcluded,
			Description:    "Defensive deck with strong reactive plays and spell control",
		},
		mulligan.ArchetypeSiege: {
			Archetype:      mulligan.ArchetypeSiege,
			MinElixir:      3.0,
			MaxElixir:      4.0,
			RequiredRoles:  siegeRoles,
			PreferredCards: siegePreferred,
			ExcludedCards:  siegeExcluded,
			Description:    "Siege building-focused deck with strong defensive support",
		},
		mulligan.ArchetypeBridgeSpam: {
			Archetype:      mulligan.ArchetypeBridgeSpam,
			MinElixir:      3.0,
			MaxElixir:      4.0,
			RequiredRoles:  bridgeSpamRoles,
			PreferredCards: bridgeSpamPreferred,
			ExcludedCards:  bridgeSpamExcluded,
			Description:    "Aggressive deck with fast units for quick bridge pressure",
		},
		mulligan.ArchetypeMidrange: {
			Archetype:      mulligan.ArchetypeMidrange,
			MinElixir:      3.0,
			MaxElixir:      4.0,
			RequiredRoles:  midrangeRoles,
			PreferredCards: midrangePreferred,
			ExcludedCards:  []string{},
			Description:    "Balanced deck with flexible offense and defense options",
		},
		mulligan.ArchetypeSpawndeck: {
			Archetype:      mulligan.ArchetypeSpawndeck,
			MinElixir:      3.5,
			MaxElixir:      5.0,
			RequiredRoles:  spawndeckRoles,
			PreferredCards: spawndeckPreferred,
			ExcludedCards:  spawndeckExcluded,
			Description:    "Spawn building-based deck with continuous troop generation",
		},
		mulligan.ArchetypeBait: {
			Archetype:      mulligan.ArchetypeBait,
			MinElixir:      2.8,
			MaxElixir:      3.8,
			RequiredRoles:  baitRoles,
			PreferredCards: baitPreferred,
			ExcludedCards:  baitExcluded,
			Description:    "Spell bait deck with multiple targets for enemy spells",
		},
	}
}

// GetAllArchetypes returns an ordered list of all archetypes for iteration.
func GetAllArchetypes() []mulligan.Archetype {
	return []mulligan.Archetype{
		mulligan.ArchetypeBeatdown,
		mulligan.ArchetypeCycle,
		mulligan.ArchetypeControl,
		mulligan.ArchetypeSiege,
		mulligan.ArchetypeBridgeSpam,
		mulligan.ArchetypeMidrange,
		mulligan.ArchetypeSpawndeck,
		mulligan.ArchetypeBait,
	}
}
