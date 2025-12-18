package mulligan

// initializeCardDatabase creates a database of cards with their mulligan properties
func initializeCardDatabase() map[string]CardInfo {
	return map[string]CardInfo{
		// Win Conditions
		"Hog Rider": {
			Name:         "Hog Rider",
			Elixir:       4,
			Type:         "troop",
			Role:         RoleWinCondition,
			OpeningScore: 0.8,
		},
		"Battle Ram": {
			Name:         "Battle Ram",
			Elixir:       4,
			Type:         "troop",
			Role:         RoleWinCondition,
			OpeningScore: 0.6,
		},
		"Goblin Barrel": {
			Name:         "Goblin Barrel",
			Elixir:       3,
			Type:         "spell",
			Role:         RoleWinCondition,
			OpeningScore: 0.2, // Never open with barrel
		},
		"Mortar": {
			Name:         "Mortar",
			Elixir:       4,
			Type:         "building",
			Role:         RoleWinCondition,
			OpeningScore: 0.1, // Never open with siege
		},
		"X-Bow": {
			Name:         "X-Bow",
			Elixir:       6,
			Type:         "building",
			Role:         RoleWinCondition,
			OpeningScore: 0.1,
		},
		"Rocket": {
			Name:         "Rocket",
			Elixir:       6,
			Type:         "spell",
			Role:         RoleWinCondition,
			OpeningScore: 0.1,
		},
		"Lightning": {
			Name:         "Lightning",
			Elixir:       6,
			Type:         "spell",
			Role:         RoleSpell,
			OpeningScore: 0.2,
		},
		"Fireball": {
			Name:         "Fireball",
			Elixir:       4,
			Type:         "spell",
			Role:         RoleSpell,
			OpeningScore: 0.3,
		},
		"Poison": {
			Name:         "Poison",
			Elixir:       4,
			Type:         "spell",
			Role:         RoleSpell,
			OpeningScore: 0.3,
		},

		// Big Tanks (Beatdown)
		"Golem": {
			Name:         "Golem",
			Elixir:       8,
			Type:         "troop",
			Role:         RoleWinCondition,
			OpeningScore: 0.0,
		},
		"Giant": {
			Name:         "Giant",
			Elixir:       5,
			Type:         "troop",
			Role:         RoleWinCondition,
			OpeningScore: 0.0,
		},
		"Lava Hound": {
			Name:         "Lava Hound",
			Elixir:       7,
			Type:         "troop",
			Role:         RoleWinCondition,
			OpeningScore: 0.0,
		},
		"Royal Giant": {
			Name:         "Royal Giant",
			Elixir:       6,
			Type:         "troop",
			Role:         RoleWinCondition,
			OpeningScore: 0.1,
		},
		"Baby Dragon": {
			Name:         "Baby Dragon",
			Elixir:       4,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.6,
		},
		"Skeleton Dragons": {
			Name:         "Skeleton Dragons",
			Elixir:       4,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.7,
		},

		// Defensive Buildings
		"Cannon": {
			Name:         "Cannon",
			Elixir:       3,
			Type:         "building",
			Role:         RoleDefensive,
			OpeningScore: 0.4, // Only reactive
		},
		"Tesla": {
			Name:         "Tesla",
			Elixir:       4,
			Type:         "building",
			Role:         RoleDefensive,
			OpeningScore: 0.4,
		},
		"Inferno Tower": {
			Name:         "Inferno Tower",
			Elixir:       5,
			Type:         "building",
			Role:         RoleDefensive,
			OpeningScore: 0.3,
		},
		"Bomb Tower": {
			Name:         "Bomb Tower",
			Elixir:       4,
			Type:         "building",
			Role:         RoleDefensive,
			OpeningScore: 0.4,
		},
		"Furnace": {
			Name:         "Furnace",
			Elixir:       4,
			Type:         "building",
			Role:         RoleDefensive,
			OpeningScore: 0.5,
		},
		"Goblin Cage": {
			Name:         "Goblin Cage",
			Elixir:       3,
			Type:         "building",
			Role:         RoleDefensive,
			OpeningScore: 0.6,
		},
		"Tombstone": {
			Name:         "Tombstone",
			Elixir:       3,
			Type:         "building",
			Role:         RoleDefensive,
			OpeningScore: 0.6,
		},

		// Spawn Buildings
		"Goblin Hut": {
			Name:         "Goblin Hut",
			Elixir:       5,
			Type:         "building",
			Role:         RoleDefensive,
			OpeningScore: 0.5,
		},
		"Barbarian Hut": {
			Name:         "Barbarian Hut",
			Elixir:       7,
			Type:         "building",
			Role:         RoleDefensive,
			OpeningScore: 0.3,
		},
		"Elixir Collector": {
			Name:         "Elixir Collector",
			Elixir:       6,
			Type:         "building",
			Role:         RoleSupport,
			OpeningScore: 0.7,
		},

		// Cycle Cards (Low Elixir)
		"Skeletons": {
			Name:         "Skeletons",
			Elixir:       1,
			Type:         "troop",
			Role:         RoleCycle,
			OpeningScore: 0.9,
		},
		"Ice Spirit": {
			Name:         "Ice Spirit",
			Elixir:       1,
			Type:         "troop",
			Role:         RoleCycle,
			OpeningScore: 0.9,
		},
		"Fire Spirit": {
			Name:         "Fire Spirit",
			Elixir:       1,
			Type:         "troop",
			Role:         RoleCycle,
			OpeningScore: 0.9,
		},
		"Zap": {
			Name:         "Zap",
			Elixir:       2,
			Type:         "spell",
			Role:         RoleSpell,
			OpeningScore: 0.7,
		},
		"Log": {
			Name:         "The Log",
			Elixir:       2,
			Type:         "spell",
			Role:         RoleSpell,
			OpeningScore: 0.7,
		},
		"Snowball": {
			Name:         "Snowball",
			Elixir:       2,
			Type:         "spell",
			Role:         RoleSpell,
			OpeningScore: 0.7,
		},
		"Arrows": {
			Name:         "Arrows",
			Elixir:       3,
			Type:         "spell",
			Role:         RoleSpell,
			OpeningScore: 0.6,
		},

		// Small Troops
		"Goblins": {
			Name:         "Goblins",
			Elixir:       2,
			Type:         "troop",
			Role:         RoleCycle,
			OpeningScore: 0.8,
		},
		"Bats": {
			Name:         "Bats",
			Elixir:       2,
			Type:         "troop",
			Role:         RoleCycle,
			OpeningScore: 0.8,
		},
		"Minions": {
			Name:         "Minions",
			Elixir:       3,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.7,
		},
		"Minion Horde": {
			Name:         "Minion Horde",
			Elixir:       5,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.4,
		},
		"Goblin Gang": {
			Name:         "Goblin Gang",
			Elixir:       3,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.7,
		},
		"Archers": {
			Name:         "Archers",
			Elixir:       3,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.8,
		},
		"Knight": {
			Name:         "Knight",
			Elixir:       3,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.8,
		},
		"Bandit": {
			Name:         "Bandit",
			Elixir:       3,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.7,
		},
		"Wall Breakers": {
			Name:         "Wall Breakers",
			Elixir:       2,
			Type:         "troop",
			Role:         RoleCycle,
			OpeningScore: 0.6,
		},

		// Medium Troops
		"Valkyrie": {
			Name:         "Valkyrie",
			Elixir:       4,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.6,
		},
		"Hunter": {
			Name:         "Hunter",
			Elixir:       4,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.6,
		},
		"Musketeer": {
			Name:         "Musketeer",
			Elixir:       4,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.5,
		},
		"Electro Wizard": {
			Name:         "Electro Wizard",
			Elixir:       4,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.6,
		},
		"Dark Prince": {
			Name:         "Dark Prince",
			Elixir:       4,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.6,
		},
		"Princess": {
			Name:         "Princess",
			Elixir:       3,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.5,
		},

		// Big Troops
		"Barbarians": {
			Name:         "Barbarians",
			Elixir:       5,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.4,
		},
		"PEKKA": {
			Name:         "PEKKA",
			Elixir:       7,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.2,
		},
		"Bowler": {
			Name:         "Bowler",
			Elixir:       5,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.4,
		},
		"Executioner": {
			Name:         "Executioner",
			Elixir:       5,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.4,
		},
		"Witch": {
			Name:         "Witch",
			Elixir:       5,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.4,
		},
		"Lumberjack": {
			Name:         "Lumberjack",
			Elixir:       4,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.6,
		},

		// Spells
		"Rage": {
			Name:         "Rage",
			Elixir:       2,
			Type:         "spell",
			Role:         RoleSpell,
			OpeningScore: 0.3,
		},
		"Freeze": {
			Name:         "Freeze",
			Elixir:       4,
			Type:         "spell",
			Role:         RoleSpell,
			OpeningScore: 0.3,
		},
		"Giant Snowball": {
			Name:         "Giant Snowball",
			Elixir:       2,
			Type:         "spell",
			Role:         RoleSpell,
			OpeningScore: 0.6,
		},
		"Barbarian Barrel": {
			Name:         "Barbarian Barrel",
			Elixir:       2,
			Type:         "spell",
			Role:         RoleSpell,
			OpeningScore: 0.6,
		},

		// Champion Cards
		"Phoenix": {
			Name:         "Phoenix",
			Elixir:       7,
			Type:         "troop",
			Role:         RoleWinCondition,
			OpeningScore: 0.2,
		},
		"Monk": {
			Name:         "Monk",
			Elixir:       5,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.4,
		},
		"Architect": {
			Name:         "Architect",
			Elixir:       4,
			Type:         "troop",
			Role:         RoleSupport,
			OpeningScore: 0.5,
		},
	}
}
