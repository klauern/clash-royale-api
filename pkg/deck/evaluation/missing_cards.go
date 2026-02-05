// Package evaluation provides comprehensive deck evaluation functionality
package evaluation

import (
	"fmt"
	"sort"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// cardUnlockArenas is a package-level cache of card unlock data
// to avoid rebuilding the map on every function call
var cardUnlockArenas = map[string]int{
	// Training Camp (Arena 0)
	"Knight":         0,
	"Archers":        0,
	"Goblins":        0,
	"Giant":          0,
	"P.E.K.K.A":      0,
	"Minions":        0,
	"Balloon":        0,
	"Witch":          0,
	"Barbarians":     0,
	"Golem":          0,
	"Skeletons":      0,
	"Valkyrie":       0,
	"Skeleton Army":  0,
	"Bomber":         0,
	"Musketeer":      0,
	"Baby Dragon":    0,
	"Prince":         0,
	"Wizard":         0,
	"Mini P.E.K.K.A": 0,
	"Fireball":       0,
	"Arrows":         0,
	"Zap":            0,
	"Cannon":         0,
	"Tesla":          0,

	// Arena 1
	"Spear Goblins":  1,
	"Giant Skeleton": 1,
	"Tombstone":      1,

	// Arena 2
	"Hog Rider":    2,
	"Minion Horde": 2,
	"Rage":         2,
	"Goblin Hut":   2,

	// Arena 3
	"Ice Wizard":    3,
	"Royal Giant":   3,
	"Rocket":        3,
	"Goblin Barrel": 3,

	// Arena 4
	"Guards":      4,
	"Princess":    4,
	"Dark Prince": 4,
	"Freeze":      4,
	"Mirror":      4,
	"Lightning":   4,

	// Arena 5
	"Three Musketeers": 5,
	"Lava Hound":       5,
	"Poison":           5,
	"Elixir Collector": 5,

	// Arena 6
	"Ice Spirit":  6,
	"Fire Spirit": 6,
	"Miner":       6,
	"Sparky":      6,
	"Graveyard":   6,
	"The Log":     6,

	// Arena 7
	"Bowler":         7,
	"Lumberjack":     7,
	"Battle Ram":     7,
	"Inferno Dragon": 7,
	"Tornado":        7,
	"Clone":          7,

	// Arena 8
	"Ice Golem":      8,
	"Mega Minion":    8,
	"Dart Goblin":    8,
	"Goblin Gang":    8,
	"Electro Wizard": 8,
	"Earthquake":     8,

	// Arena 9
	"Elite Barbarians": 9,
	"Hunter":           9,
	"Executioner":      9,
	"Bandit":           9,

	// Arena 10
	"Royal Recruits": 10,
	"Night Witch":    10,
	"Bats":           10,
	"Royal Ghost":    10,

	// Arena 11
	"Ram Rider":        11,
	"Zappies":          11,
	"Rascals":          11,
	"Cannon Cart":      11,
	"Mega Knight":      11,
	"Barbarian Barrel": 11,

	// Arena 12
	"Skeleton Barrel": 12,
	"Flying Machine":  12,
	"Wall Breakers":   12,
	"Royal Hogs":      12,
	"Goblin Giant":    12,
	"Heal Spirit":     12,

	// Arena 13+
	"Fisherman":      13,
	"Magic Archer":   13,
	"Electro Dragon": 13,
	"Firecracker":    13,
	"Giant Snowball": 13,

	// Arena 14+
	"Mighty Miner":   14,
	"Elixir Golem":   14,
	"Battle Healer":  14,
	"Royal Delivery": 14,

	// Arena 15+ (Legendary Arena)
	"Skeleton King":  15,
	"Archer Queen":   15,
	"Golden Knight":  15,
	"Monk":           15,
	"Mother Witch":   15,
	"Electro Spirit": 15,
	"Electro Giant":  15,
	"Phoenix":        15,
}

// MissingCard represents a card that the player doesn't have
type MissingCard struct {
	// Name is the card name
	Name string `json:"name"`

	// Rarity is the card rarity (Common, Rare, Epic, Legendary, Champion)
	Rarity string `json:"rarity"`

	// UnlockArena is the arena where this card unlocks
	UnlockArena int `json:"unlock_arena"`

	// UnlockArenaName is the name of the unlock arena
	UnlockArenaName string `json:"unlock_arena_name"`

	// AlternativeCards are suggested replacements the player owns
	AlternativeCards []string `json:"alternative_cards,omitempty"`

	// IsLocked indicates if the card is not yet unlocked
	IsLocked bool `json:"is_locked"`
}

// MissingCardsAnalysis contains analysis of cards missing from player's collection
type MissingCardsAnalysis struct {
	// Deck is the deck being analyzed
	Deck []string `json:"deck"`

	// MissingCards are cards the player doesn't own
	MissingCards []MissingCard `json:"missing_cards"`

	// MissingCount is the number of missing cards
	MissingCount int `json:"missing_count"`

	// AvailableCount is the number of cards the player owns
	AvailableCount int `json:"available_count"`

	// IsPlayable indicates if the deck is playable (no missing cards)
	IsPlayable bool `json:"is_playable"`

	// SuggestedReplacements maps missing cards to suggested alternatives
	SuggestedReplacements map[string][]string `json:"suggested_replacements,omitempty"`
}

// IdentifyMissingCardsWithContext analyzes which cards in a deck are missing from player's collection
// Uses PlayerContext for arena-aware validation and card ownership checking
// This is the preferred method for context-aware deck evaluation
func IdentifyMissingCardsWithContext(
	deckCards []deck.CardCandidate,
	playerContext *PlayerContext,
) *MissingCardsAnalysis {
	if playerContext == nil {
		// No context - treat as all cards available
		return &MissingCardsAnalysis{
			Deck:                  extractCardNames(deckCards),
			MissingCards:          make([]MissingCard, 0),
			SuggestedReplacements: make(map[string][]string),
			AvailableCount:        len(deckCards),
			IsPlayable:            true,
		}
	}

	// Build simple collection map from PlayerContext
	playerCollection := make(map[string]bool)
	for cardName := range playerContext.Collection {
		playerCollection[cardName] = true
	}

	// Use existing logic with PlayerContext integration
	return identifyMissingCardsInternal(deckCards, playerCollection, playerContext)
}

// IdentifyMissingCards analyzes which cards in a deck are missing from player's collection
// Parameters:
//   - deckCards: The deck to analyze
//   - playerCollection: Map of card name -> owned (true if player has it)
//   - playerArena: Current arena level of the player (0 = all cards unlocked)
//
// Deprecated: Use IdentifyMissingCardsWithContext for arena-aware validation
func IdentifyMissingCards(
	deckCards []deck.CardCandidate,
	playerCollection map[string]bool,
	playerArena int,
) *MissingCardsAnalysis {
	// Create minimal PlayerContext for backward compatibility
	ctx := &PlayerContext{
		ArenaID:    playerArena,
		Collection: make(map[string]CardLevelInfo),
	}

	// Populate collection from the map
	if playerCollection != nil {
		for cardName := range playerCollection {
			ctx.Collection[cardName] = CardLevelInfo{}
		}
	}

	return identifyMissingCardsInternal(deckCards, playerCollection, ctx)
}

// identifyMissingCardsInternal is the core implementation used by both methods
func identifyMissingCardsInternal(
	deckCards []deck.CardCandidate,
	playerCollection map[string]bool,
	playerContext *PlayerContext,
) *MissingCardsAnalysis {
	analysis := &MissingCardsAnalysis{
		Deck:                  make([]string, len(deckCards)),
		MissingCards:          make([]MissingCard, 0),
		SuggestedReplacements: make(map[string][]string),
	}

	// Get deck card names
	for i, card := range deckCards {
		analysis.Deck[i] = card.Name
	}

	// Check each card
	for _, card := range deckCards {
		// Check if player owns this card
		owned := playerCollection != nil && playerCollection[card.Name]

		if owned {
			analysis.AvailableCount++
			continue
		}

		// Card is missing - get details
		unlockArena := getCardUnlockArena(card.Name)

		// Use PlayerContext for arena-aware validation
		isLocked := !playerContext.IsCardUnlockedInArena(card.Name)

		missing := MissingCard{
			Name:            card.Name,
			Rarity:          card.Rarity,
			UnlockArena:     unlockArena,
			UnlockArenaName: getArenaName(unlockArena),
			IsLocked:        isLocked,
		}

		// Find alternative cards the player owns
		alternatives := findOwnedAlternatives(card, playerCollection)
		missing.AlternativeCards = alternatives
		if len(alternatives) > 0 {
			analysis.SuggestedReplacements[card.Name] = alternatives
		}

		analysis.MissingCards = append(analysis.MissingCards, missing)
		analysis.MissingCount++
	}

	// Deck is playable only if no cards are missing
	analysis.IsPlayable = analysis.MissingCount == 0

	// Sort missing cards by unlock arena
	sort.Slice(analysis.MissingCards, func(i, j int) bool {
		return analysis.MissingCards[i].UnlockArena < analysis.MissingCards[j].UnlockArena
	})

	return analysis
}

// findOwnedAlternatives finds replacement cards that the player owns
func findOwnedAlternatives(
	missingCard deck.CardCandidate,
	playerCollection map[string]bool,
) []string {
	alternatives := make([]string, 0)

	// Get similar cards
	similar := getSimilarCards(missingCard)

	// Filter to only cards the player owns
	for _, alt := range similar {
		if playerCollection != nil && playerCollection[alt.Name] {
			alternatives = append(alternatives, alt.Name)
		}
	}

	return alternatives
}

// getCardUnlockArena returns the arena where a card unlocks
func getCardUnlockArena(cardName string) int {
	if arena, exists := cardUnlockArenas[cardName]; exists {
		return arena
	}
	// Default to 0 (available from start)
	return 0
}

// getArenaName returns the name for an arena number
func getArenaName(arenaNum int) string {
	arenaNames := map[int]string{
		0:  "Training Camp",
		1:  "Goblin Stadium",
		2:  "Bone Pit",
		3:  "Barbarian Bowl",
		4:  "P.E.K.K.A's Playhouse",
		5:  "Spell Valley",
		6:  "Builder's Workshop",
		7:  "Royal Arena",
		8:  "Frozen Peak",
		9:  "Jungle Arena",
		10: "Hog Mountain",
		11: "Electro Valley",
		12: "Spooky Town",
		13: "Rascal's Hideout",
		14: "Serenity Peak",
		15: "Legendary Arena",
	}

	if name, exists := arenaNames[arenaNum]; exists {
		return name
	}

	return fmt.Sprintf("Arena %d", arenaNum)
}

// FormatMissingCardsReport creates a human-readable report of missing cards
func FormatMissingCardsReport(analysis *MissingCardsAnalysis) string {
	if analysis.IsPlayable {
		return "✓ All cards in this deck are available in your collection!\n"
	}

	report := fmt.Sprintf("Missing Cards Analysis\n")
	report += fmt.Sprintf("═══════════════════════\n\n")
	report += fmt.Sprintf("Deck Status: %d/%d cards available\n\n", analysis.AvailableCount, len(analysis.Deck))

	for i, missing := range analysis.MissingCards {
		report += fmt.Sprintf("%d. %s (%s)\n", i+1, missing.Name, missing.Rarity)
		report += fmt.Sprintf("   Unlocks: %s (Arena %d)\n", missing.UnlockArenaName, missing.UnlockArena)

		if missing.IsLocked {
			report += fmt.Sprintf("   Status: 🔒 LOCKED - Progress to Arena %d to unlock\n", missing.UnlockArena)
		} else {
			report += fmt.Sprintf("   Status: ✓ Unlocked - Available in chests and shop\n")
		}

		if len(missing.AlternativeCards) > 0 {
			report += fmt.Sprintf("   Alternatives: %s\n", joinCardNames(missing.AlternativeCards))
		} else {
			report += fmt.Sprintf("   Alternatives: None found in your collection\n")
		}

		report += "\n"
	}

	return report
}

// extractCardNames extracts card names from a slice of CardCandidates
func extractCardNames(deckCards []deck.CardCandidate) []string {
	names := make([]string, len(deckCards))
	for i, card := range deckCards {
		names[i] = card.Name
	}
	return names
}

// joinCardNames joins card names with commas
func joinCardNames(cards []string) string {
	if len(cards) == 0 {
		return ""
	}
	if len(cards) == 1 {
		return cards[0]
	}
	if len(cards) == 2 {
		return cards[0] + ", " + cards[1]
	}

	result := ""
	for i, card := range cards {
		if i == len(cards)-1 {
			result += card
		} else {
			result += card + ", "
		}
	}
	return result
}
