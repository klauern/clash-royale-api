package taxonomy

import "strings"

// Shared archetype card groups consumed by both scoring and constraints.
var (
	beatdownHeavyTanks = []string{"Golem", "Lava Hound", "Electro Giant", "Giant", "Mega Knight"}
	beatdownSupport    = []string{"Baby Dragon", "Night Witch", "Lumberjack", "Mega Minion", "Witch"}

	controlDefensiveBuildings = []string{"Tesla", "Cannon", "Inferno Tower", "Bomb Tower"}
	controlBigSpells          = []string{"Poison", "Fireball", "Lightning", "Rocket"}

	cycleWinConditions = []string{"Hog Rider", "Royal Giant", "Royal Hogs"}
	cycleCoreCards     = []string{"Skeletons", "Ice Spirit", "Ice Golem", "Electro Spirit"}

	// Shared canonical card names used in multiple archetype detectors.
	BeatdownCoreWinConditions = []string{"Golem", "Giant", "Lava Hound"}
	SiegeWinConditions        = []string{"X-Bow", "Mortar"}
	BridgeSpamCoreWinConds    = []string{"Battle Ram"}

	// Mulligan-specific substring signals.
	BeatdownCoreSignals     = []string{"golem", "giant", "lava hound"}
	SiegeWinConditionTokens = []string{"x-bow", "xbow", "mortar"}
	BridgeSpamSignals       = []string{"battle ram", "hog rider"}
)

// BeatdownHeavyTanks returns the shared heavy tank card group.
func BeatdownHeavyTanks() []string {
	return Clone(beatdownHeavyTanks)
}

// BeatdownSupport returns the shared beatdown support troop group.
func BeatdownSupport() []string {
	return Clone(beatdownSupport)
}

// ControlDefensiveBuildings returns the shared defensive building group.
func ControlDefensiveBuildings() []string {
	return Clone(controlDefensiveBuildings)
}

// ControlBigSpells returns the shared control big spell group.
func ControlBigSpells() []string {
	return Clone(controlBigSpells)
}

// CycleWinConditions returns the shared cycle win condition group.
func CycleWinConditions() []string {
	return Clone(cycleWinConditions)
}

// CycleCoreCards returns the shared low-cost cycle core group.
func CycleCoreCards() []string {
	return Clone(cycleCoreCards)
}

// Clone returns a defensive copy of cards for callers that need mutable slices.
func Clone(cards []string) []string {
	out := make([]string, len(cards))
	copy(out, cards)
	return out
}

// Merge concatenates multiple card groups into one list.
// It preserves order and does not deduplicate cards.
func Merge(groups ...[]string) []string {
	total := 0
	for _, g := range groups {
		total += len(g)
	}
	out := make([]string, 0, total)
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

// ContainsAnySubstringFold returns true when any value contains any signal
// using case-insensitive matching.
func ContainsAnySubstringFold(values, signals []string) bool {
	for _, value := range values {
		lowerValue := strings.ToLower(value)
		for _, signal := range signals {
			if strings.Contains(lowerValue, strings.ToLower(signal)) {
				return true
			}
		}
	}
	return false
}
