package archetypes

import (
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
)

// GetArchetypeConstraints returns predefined constraints for all 8 archetypes.
// These constraints guide deck building to match specific playstyle characteristics.
func GetArchetypeConstraints() map[mulligan.Archetype]ArchetypeConstraints {
	return map[mulligan.Archetype]ArchetypeConstraints{
		mulligan.ArchetypeBeatdown: {
			Archetype: mulligan.ArchetypeBeatdown,
			MinElixir: 4.0,
			MaxElixir: 5.5,
			RequiredRoles: map[deck.CardRole]int{
				deck.RoleWinCondition: 1, // Must have heavy tank
				deck.RoleSupport:      2, // Support troops for tank
			},
			PreferredCards: []string{
				"Golem", "Giant", "Lava Hound", "Electro Giant",
				"Baby Dragon", "Night Witch", "Mega Minion", "Lumberjack",
				"Lightning", "Tornado", "Arrows",
			},
			ExcludedCards: []string{
				"X-Bow", "Mortar", "Hog Rider", "Miner", "Goblin Barrel",
			},
			Description: "Heavy tank-based deck with support troops for big pushes",
		},

		mulligan.ArchetypeCycle: {
			Archetype: mulligan.ArchetypeCycle,
			MinElixir: 2.5,
			MaxElixir: 3.5,
			RequiredRoles: map[deck.CardRole]int{
				deck.RoleWinCondition: 1,
				deck.RoleCycle:        3, // Many cheap cards for fast cycling
				deck.RoleSpellSmall:   1,
			},
			PreferredCards: []string{
				"Hog Rider", "Miner", "Skeletons", "Ice Spirit",
				"Ice Golem", "Cannon", "Musketeer", "Log",
				"Fireball", "Electro Spirit", "Bats",
			},
			ExcludedCards: []string{
				"Golem", "Lava Hound", "Giant", "Electro Giant",
				"P.E.K.K.A", "Mega Knight",
			},
			Description: "Fast cycling deck with low elixir cards for constant pressure",
		},

		mulligan.ArchetypeControl: {
			Archetype: mulligan.ArchetypeControl,
			MinElixir: 3.5,
			MaxElixir: 4.5,
			RequiredRoles: map[deck.CardRole]int{
				deck.RoleBuilding:   1, // Defensive building
				deck.RoleSpellBig:   1, // Big spell for control
				deck.RoleSpellSmall: 1,
			},
			PreferredCards: []string{
				"Inferno Tower", "Cannon", "Bomb Tower", "Tesla",
				"Valkyrie", "Wizard", "Musketeer", "Archers",
				"Fireball", "Poison", "Log", "Arrows",
			},
			ExcludedCards: []string{
				"Golem", "Giant", "Lava Hound",
			},
			Description: "Defensive deck with strong reactive plays and spell control",
		},

		mulligan.ArchetypeSiege: {
			Archetype: mulligan.ArchetypeSiege,
			MinElixir: 3.0,
			MaxElixir: 4.0,
			RequiredRoles: map[deck.CardRole]int{
				deck.RoleWinCondition: 1, // Must have X-Bow or Mortar
				deck.RoleBuilding:     1, // Additional defensive building
				deck.RoleCycle:        2, // Cycle cards
			},
			PreferredCards: []string{
				"X-Bow", "Mortar", "Tesla", "Cannon",
				"Archers", "Skeletons", "Ice Spirit", "Knight",
				"Fireball", "Log", "Rocket",
			},
			ExcludedCards: []string{
				"Golem", "Giant", "Lava Hound", "Hog Rider",
				"Mega Knight", "P.E.K.K.A",
			},
			Description: "Siege building-focused deck with strong defensive support",
		},

		mulligan.ArchetypeBridgeSpam: {
			Archetype: mulligan.ArchetypeBridgeSpam,
			MinElixir: 3.0,
			MaxElixir: 4.0,
			RequiredRoles: map[deck.CardRole]int{
				deck.RoleWinCondition: 1,
				deck.RoleSupport:      2,
			},
			PreferredCards: []string{
				"Battle Ram", "Hog Rider", "Bandit", "Royal Ghost",
				"Electro Wizard", "Magic Archer", "Dark Prince",
				"Poison", "Fireball", "Zap", "Log",
			},
			ExcludedCards: []string{
				"Golem", "Lava Hound", "Giant", "X-Bow", "Mortar",
			},
			Description: "Aggressive deck with fast units for quick bridge pressure",
		},

		mulligan.ArchetypeMidrange: {
			Archetype: mulligan.ArchetypeMidrange,
			MinElixir: 3.0,
			MaxElixir: 4.0,
			RequiredRoles: map[deck.CardRole]int{
				deck.RoleWinCondition: 1,
				deck.RoleSpellBig:     1,
				deck.RoleSpellSmall:   1,
			},
			PreferredCards: []string{
				"Hog Rider", "Royal Giant", "Royal Hogs", "Valkyrie",
				"Musketeer", "Mega Minion", "Knight", "Archers",
				"Fireball", "Zap", "Log", "Arrows",
			},
			ExcludedCards: []string{},
			Description:   "Balanced deck with flexible offense and defense options",
		},

		mulligan.ArchetypeSpawndeck: {
			Archetype: mulligan.ArchetypeSpawndeck,
			MinElixir: 3.5,
			MaxElixir: 5.0,
			RequiredRoles: map[deck.CardRole]int{
				deck.RoleBuilding: 2, // Spawn buildings
			},
			PreferredCards: []string{
				"Goblin Hut", "Barbarian Hut", "Furnace", "Tombstone",
				"Graveyard", "Poison", "Fireball", "Zap",
				"Valkyrie", "Baby Dragon", "Mega Minion",
			},
			ExcludedCards: []string{
				"X-Bow", "Mortar", "Hog Rider",
			},
			Description: "Spawn building-based deck with continuous troop generation",
		},

		mulligan.ArchetypeBait: {
			Archetype: mulligan.ArchetypeBait,
			MinElixir: 2.8,
			MaxElixir: 3.8,
			RequiredRoles: map[deck.CardRole]int{
				deck.RoleWinCondition: 1,
				deck.RoleSpellSmall:   1,
			},
			PreferredCards: []string{
				"Goblin Barrel", "Princess", "Goblin Gang", "Skeleton Army",
				"Rocket", "Log", "Zap", "Arrows",
				"Knight", "Ice Spirit", "Inferno Tower",
			},
			ExcludedCards: []string{
				"Golem", "Giant", "Lava Hound", "X-Bow", "Mortar",
			},
			Description: "Spell bait deck with multiple targets for enemy spells",
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
