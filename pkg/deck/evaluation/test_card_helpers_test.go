package evaluation

import (
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

var (
	testWinConditions = map[string]bool{
		"Hog Rider": true, "Giant": true, "Royal Giant": true, "Golem": true,
		"Lava Hound": true, "P.E.K.K.A": true, "Mega Knight": true, "Balloon": true,
		"X-Bow": true, "Mortar": true, "Miner": true, "Graveyard": true,
		"Goblin Barrel": true, "Goblin Drill": true, "Electro Giant": true,
		"Elite Barbarians": true, "Battle Ram": true, "Ram Rider": true,
		"Wall Breakers": true, "Sparky": true, "Royal Hogs": true,
		"Three Musketeers": true, "Archer Queen": true, "Golden Knight": true,
		"Skeleton King": true, "Little Prince": true, "Phoenix": true,
	}
	testSpellBig = map[string]bool{
		"Fireball": true, "Lightning": true, "Rocket": true, "Poison": true,
		"Freeze": true, "Earthquake": true, "Rage": true, "Tornado": true,
	}
	testSpellSmall = map[string]bool{
		"Zap": true, "The Log": true, "Arrows": true, "Snowball": true,
		"Barbarian Barrel": true, "Giant Snowball": true, "Royal Delivery": true,
	}
	testBuildings = map[string]bool{
		"Cannon": true, "Tesla": true, "Inferno Tower": true, "Bomb Tower": true,
		"X-Bow": true, "Mortar": true, "Elixir Collector": true, "Furnace": true,
		"Goblin Hut": true, "Goblin Cage": true, "Tombstone": true,
	}
	testLegendaries = map[string]bool{
		"Princess": true, "The Log": true, "Miner": true, "Ice Wizard": true,
		"Mega Knight": true, "Night Witch": true, "Lumberjack": true,
		"Electro Wizard": true, "Lava Hound": true, "Sparky": true,
		"Bandit": true, "Battle Ram": true, "Royal Ghost": true,
	}
	testEpics = map[string]bool{
		"Golem": true, "P.E.K.K.A": true, "Balloon": true, "X-Bow": true,
		"Mortar": true, "Graveyard": true, "Freeze": true, "Poison": true,
		"Tornado": true, "Rocket": true, "Lightning": true, "Baby Dragon": true,
		"Prince": true, "Dark Prince": true, "Bowling": true, "Three Musketeers": true,
	}
	testChampions = map[string]bool{
		"Archer Queen": true, "Golden Knight": true, "Skeleton King": true,
		"Little Prince": true, "Mighty Miner": true, "Phoenix": true,
	}
	testElixirMap = map[string]int{
		"Skeletons": 1, "Ice Spirit": 1, "Bats": 1, "Fire Spirit": 1,
		"The Log": 2, "Zap": 2, "Snowball": 2, "Knight": 3,
		"Ice Golem": 2, "Heal Spirit": 1, "Spirit": 1,
		"Musketeer": 4, "Valkyrie": 4, "Mini P.E.K.K.A": 4, "Mega Minion": 3,
		"Hog Rider": 4, "Cannon": 3, "Tesla": 4, "Fireball": 4,
		"Golem": 8, "P.E.K.K.A": 7, "Mega Knight": 7, "Balloon": 5,
		"X-Bow": 6, "Mortar": 4, "Miner": 3, "Graveyard": 5,
		"Lava Hound": 7, "Electro Giant": 8, "Lightning": 6,
		"Rocket": 6, "Poison": 4, "Freeze": 4, "Tornado": 3,
		"Baby Dragon": 4, "Night Witch": 4, "Lumberjack": 4,
		"Electro Wizard": 4, "Bandit": 3, "Battle Ram": 5,
		"Royal Ghost": 3, "Inferno Tower": 5, "Inferno Dragon": 4,
		"Elixir Collector": 6, "Goblin Barrel": 3, "Goblin Gang": 3,
		"Goblin Drill": 4, "Princess": 3, "Arrows": 3,
		"Skeleton Army": 3, "Tombstone": 3, "Bomb Tower": 4,
		"Goblin Cage": 5, "Goblin Hut": 5, "Furnace": 4,
		"Archer Queen": 5, "Golden Knight": 4, "Skeleton King": 4,
		"Little Prince": 3, "Elite Barbarians": 6, "Ram Rider": 5,
		"Royal Giant": 6, "Royal Hogs": 5, "Wall Breakers": 4,
		"Sparky": 6, "Three Musketeers": 9, "Hunter": 4,
		"Witch": 5, "Executioner": 5, "Wizard": 5,
		"Magic Archer": 4, "Dart Goblin": 3, "Spear Goblins": 2,
		"Goblins": 2, "Archers": 3, "Minions": 3,
		"Skeleton Dragons": 4, "Mother Witch": 4, "Dark Prince": 4,
		"Fisherman": 3, "Royal Delivery": 3, "Phoenix": 4,
	}
	testAirTargetingCards = map[string]bool{
		"Archers": true, "Archer Queen": true, "Baby Dragon": true, "Bats": true,
		"Dart Goblin": true, "Electro Wizard": true, "Executioner": true, "Hunter": true,
		"Magic Archer": true, "Mega Minion": true, "Minions": true, "Musketeer": true,
		"Night Witch": true, "Phoenix": true, "Princess": true, "Skeleton Dragons": true,
		"Spear Goblins": true, "Wizard": true, "Witch": true,
	}
)

func determineTestCardRole(name string) deck.CardRole {
	if testWinConditions[name] {
		return deck.RoleWinCondition
	}
	if testSpellBig[name] {
		return deck.RoleSpellBig
	}
	if testSpellSmall[name] {
		return deck.RoleSpellSmall
	}
	if testBuildings[name] {
		return deck.RoleBuilding
	}
	return deck.RoleSupport
}

func determineTestCardRarity(name string) string {
	if testChampions[name] {
		return "Champion"
	}
	if testLegendaries[name] {
		return "Legendary"
	}
	if testEpics[name] {
		return "Epic"
	}
	return "Rare"
}

func determineTestCardElixir(name string) int {
	if elixir, ok := testElixirMap[name]; ok {
		return elixir
	}
	return 4
}

func determineTestCardTargets(name string, role deck.CardRole) string {
	if role == deck.RoleSpellBig || role == deck.RoleSpellSmall {
		return "Ground"
	}
	if testAirTargetingCards[name] {
		return "Air & Ground"
	}
	return "Ground"
}

func createTestCardCandidate(name string) deck.CardCandidate {
	role := determineTestCardRole(name)
	rarity := determineTestCardRarity(name)
	elixir := determineTestCardElixir(name)
	targets := determineTestCardTargets(name, role)

	dps := 90
	switch role {
	case deck.RoleWinCondition:
		dps = 120
	case deck.RoleBuilding:
		dps = 70
	case deck.RoleCycle:
		dps = 60
	case deck.RoleSpellBig, deck.RoleSpellSmall:
		dps = 0
	}

	return deck.CardCandidate{
		Name:     name,
		Level:    11,
		MaxLevel: 14,
		Rarity:   rarity,
		Elixir:   elixir,
		Role:     &role,
		Stats: &clashroyale.CombatStats{
			DamagePerSecond: dps,
			Targets:         targets,
		},
	}
}
