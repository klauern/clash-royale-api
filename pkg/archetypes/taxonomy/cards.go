package taxonomy

import "strings"

var (
	// Shared canonical card names used in multiple archetype detectors.
	BeatdownCoreWinConditions = []string{"Golem", "Giant", "Lava Hound"}
	SiegeWinConditions        = []string{"X-Bow", "Mortar"}
	BridgeSpamCoreWinConds    = []string{"Battle Ram"}

	// Mulligan-specific substring signals.
	BeatdownCoreSignals     = []string{"golem", "giant", "lava hound"}
	SiegeWinConditionTokens = []string{"x-bow", "xbow", "mortar"}
	BridgeSpamSignals       = []string{"battle ram", "hog rider"}
)

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
