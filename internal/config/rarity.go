// Package config provides centralized game configuration constants and mappings.
// This package consolidates rarity-related data that was previously duplicated
// across multiple packages (deck, analysis, scoring).
package config

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Rarity weights for deck scoring (1.0-1.2 scale)
// Higher rarity cards get higher weights since they're harder to level up
var rarityWeights = map[string]float64{
	"Common":    1.0,
	"Rare":      1.05,
	"Epic":      1.1,
	"Legendary": 1.15,
	"Champion":  1.2,
}

// Priority scores for upgrade recommendation analysis (0-80 scale)
// Used in upgrade priority calculations (contributes 20% of total score)
var priorityScores = map[string]float64{
	"Common":    0.0,
	"Rare":      20.0,
	"Epic":      40.0,
	"Legendary": 60.0,
	"Champion":  80.0,
}

// Priority bonuses for deck building (1.0-2.5 scale)
// Used to prioritize selection of higher rarity cards in deck building
// Adjusted from original 1.0-5.0 to prevent over-prioritization of high-rarity cards
var priorityBonuses = map[string]float64{
	"Common":    1.0,
	"Rare":      1.3,
	"Epic":      1.7,
	"Legendary": 2.2,
	"Champion":  2.5,
}

// Maximum levels for each rarity
// All rarities currently have the same max level (16)
var maxLevels = map[string]int{
	"Common":    16,
	"Rare":      16,
	"Epic":      16,
	"Legendary": 16,
	"Champion":  16,
}

// Starting (unlock) levels for each rarity
// These represent the level at which cards of each rarity are unlocked
var startingLevels = map[string]int{
	"Common":    1,
	"Rare":      3,
	"Epic":      6,
	"Legendary": 9,
	"Champion":  11,
}

// cardRarityByNormalizedName stores high-impact card rarity overrides by normalized name.
// Unknown cards intentionally fall back to Common in callers that need a safe default.
var cardRarityByNormalizedName = map[string]string{
	// Champions
	"archerqueen":  "Champion",
	"goldenknight": "Champion",
	"skeletonking": "Champion",
	"mightyminer":  "Champion",
	"monk":         "Champion",
	"littleprince": "Champion",

	// Legendaries
	"princess":      "Legendary",
	"log":           "Legendary",
	"thelog":        "Legendary",
	"miner":         "Legendary",
	"icewizard":     "Legendary",
	"lumberjack":    "Legendary",
	"infernodragon": "Legendary",
	"electrowizard": "Legendary",
	"nightwitch":    "Legendary",
	"bandit":        "Legendary",
	"royalghost":    "Legendary",
	"ramrider":      "Legendary",
	"magicarcher":   "Legendary",
	"fisherman":     "Legendary",
	"motherwitch":   "Legendary",
	"phoenix":       "Legendary",
	"sparky":        "Legendary",
	"lavahound":     "Legendary",
	"megaknight":    "Legendary",

	// Rares
	"minipekka":       "Rare",
	"threemusketeers": "Rare",

	// Epics
	"pekka":           "Epic",
	"golem":           "Epic",
	"balloon":         "Epic",
	"babydragon":      "Epic",
	"prince":          "Epic",
	"darkprince":      "Epic",
	"witch":           "Epic",
	"bowler":          "Epic",
	"executioner":     "Epic",
	"goblindrill":     "Epic",
	"goblinbarrel":    "Epic",
	"guards":          "Epic",
	"skeletonarmy":    "Epic",
	"rage":            "Epic",
	"clone":           "Epic",
	"mirror":          "Epic",
	"freeze":          "Epic",
	"poison":          "Epic",
	"tornado":         "Epic",
	"lightning":       "Epic",
	"xbow":            "Epic",
	"wallbreakers":    "Epic",
	"electrodragon":   "Epic",
	"cannoncart":      "Epic",
	"goblingiant":     "Epic",
	"giantskeleton":   "Epic",
	"hunter":          "Epic",
	"electrogiant":    "Epic",
	"elixirgolem":     "Epic",
}

// NormalizeRarity ensures rarity strings are in TitleCase for consistent map lookups.
// It handles case-insensitive input and trims whitespace.
// Returns empty string if input is empty, otherwise returns TitleCase version.
func NormalizeRarity(rarity string) string {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "common":
		return "Common"
	case "rare":
		return "Rare"
	case "epic":
		return "Epic"
	case "legendary":
		return "Legendary"
	case "champion":
		return "Champion"
	default:
		// Trim whitespace first
		trimmed := strings.TrimSpace(rarity)
		// Return empty string if input is empty after trimming
		if len(trimmed) == 0 {
			return ""
		}
		// For unknown rarities, return TitleCase version
		return cases.Title(language.English).String(strings.ToLower(trimmed))
	}
}

// GetRarityWeight returns the scoring weight for deck building (1.0-1.2 scale).
// Higher rarity cards get higher weights since they're harder to level up.
// Returns 1.0 (neutral multiplier) for unknown rarities.
//
// Values:
//   - Common: 1.0
//   - Rare: 1.05
//   - Epic: 1.1
//   - Legendary: 1.15
//   - Champion: 1.2
func GetRarityWeight(rarity string) float64 {
	normalized := NormalizeRarity(rarity)
	if weight, ok := rarityWeights[normalized]; ok {
		return weight
	}
	return 1.0 // Neutral weight for unknown rarities
}

// GetRarityPriorityScore returns the priority score for upgrade recommendations (0-80 scale).
// Used in upgrade priority calculations where it contributes 20% of the total score.
// Returns 0.0 (lowest priority) for unknown rarities.
//
// Values:
//   - Common: 0
//   - Rare: 20
//   - Epic: 40
//   - Legendary: 60
//   - Champion: 80
func GetRarityPriorityScore(rarity string) float64 {
	normalized := NormalizeRarity(rarity)
	if score, ok := priorityScores[normalized]; ok {
		return score
	}
	return 0.0 // Lowest priority for unknown rarities
}

// GetRarityPriorityBonus returns the bonus multiplier for deck building (1.0-2.5 scale).
// Used to prioritize selection of higher rarity cards when building decks.
// Returns 1.0 (neutral multiplier) for unknown rarities.
//
// Note: This serves a different purpose than GetRarityPriorityScore despite similar name.
// This is for deck building card selection, while PriorityScore is for upgrade analysis.
//
// Values:
//   - Common: 1.0
//   - Rare: 1.3
//   - Epic: 1.7
//   - Legendary: 2.2
//   - Champion: 2.5
func GetRarityPriorityBonus(rarity string) float64 {
	normalized := NormalizeRarity(rarity)
	if bonus, ok := priorityBonuses[normalized]; ok {
		return bonus
	}
	return 1.0 // Neutral bonus for unknown rarities
}

// GetMaxLevel returns the maximum level for a rarity (currently 16 for all rarities).
// Returns 0 for unknown rarities to signal invalid input.
func GetMaxLevel(rarity string) int {
	normalized := NormalizeRarity(rarity)
	if maxLevel, ok := maxLevels[normalized]; ok {
		return maxLevel
	}
	return 0 // Return 0 to signal invalid/unknown rarity
}

// GetStartingLevel returns the unlock level for a rarity.
// These represent the level at which cards of each rarity are first unlocked.
// Returns 0 for unknown rarities to signal invalid input.
//
// Values:
//   - Common: 1
//   - Rare: 3
//   - Epic: 6
//   - Legendary: 9
//   - Champion: 11
func GetStartingLevel(rarity string) int {
	normalized := NormalizeRarity(rarity)
	if startingLevel, ok := startingLevels[normalized]; ok {
		return startingLevel
	}
	return 0 // Return 0 to signal invalid/unknown rarity
}

// GetAllRarities returns all valid rarity strings in a consistent order.
// Useful for iteration or validation purposes.
func GetAllRarities() []string {
	return []string{"Common", "Rare", "Epic", "Legendary", "Champion"}
}

// LookupCardRarity resolves a card's rarity from its name.
// Returns (rarity, true) when known and ("", false) when unknown.
func LookupCardRarity(cardName string) (string, bool) {
	normalizedName := normalizeCardNameForLookup(cardName)
	if normalizedName == "" {
		return "", false
	}

	rarity, ok := cardRarityByNormalizedName[normalizedName]
	return rarity, ok
}

func normalizeCardNameForLookup(cardName string) string {
	trimmed := strings.TrimSpace(cardName)
	if trimmed == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(trimmed))
	for _, r := range strings.ToLower(trimmed) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}

	return b.String()
}
