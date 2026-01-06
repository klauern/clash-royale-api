package evaluation

import (
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// PlayerContext contains player-specific data for context-aware deck evaluation
// This includes arena level, card collection, levels, and evolution status
type PlayerContext struct {
	// Arena information
	Arena     *clashroyale.Arena
	ArenaID   int
	ArenaName string

	// Card collection: map of card name -> level data
	Collection map[string]CardLevelInfo

	// Evolution data: which cards have evolutions unlocked
	UnlockedEvolutions map[string]bool

	// Player metadata
	PlayerTag  string
	PlayerName string
}

// CardLevelInfo stores level information for a single card
type CardLevelInfo struct {
	Level             int
	MaxLevel          int
	EvolutionLevel    int
	MaxEvolutionLevel int
	Rarity            string
	Count             int // Number of cards owned
}

// NewPlayerContextFromPlayer builds a PlayerContext from a Player API response
// This is the primary way to create player context for deck evaluation
func NewPlayerContextFromPlayer(player *clashroyale.Player) *PlayerContext {
	if player == nil {
		return nil
	}

	ctx := &PlayerContext{
		Arena:              &player.Arena,
		ArenaID:            player.Arena.ID,
		ArenaName:          player.Arena.Name,
		Collection:         make(map[string]CardLevelInfo),
		UnlockedEvolutions: make(map[string]bool),
		PlayerTag:          player.Tag,
		PlayerName:         player.Name,
	}

	// Build card collection map
	for _, card := range player.Cards {
		ctx.Collection[card.Name] = CardLevelInfo{
			Level:             card.Level,
			MaxLevel:          card.MaxLevel,
			EvolutionLevel:    card.EvolutionLevel,
			MaxEvolutionLevel: card.MaxEvolutionLevel,
			Rarity:            card.Rarity,
			Count:             card.Count,
		}

		// Track unlocked evolutions
		if card.EvolutionLevel > 0 {
			ctx.UnlockedEvolutions[card.Name] = true
		}
	}

	return ctx
}

// GetCardLevel returns the level of a card in the player's collection
// Returns 0 if the card is not in the collection
func (ctx *PlayerContext) GetCardLevel(cardName string) int {
	if info, exists := ctx.Collection[cardName]; exists {
		return info.Level
	}
	return 0
}

// HasCard checks if the player owns a specific card
func (ctx *PlayerContext) HasCard(cardName string) bool {
	_, exists := ctx.Collection[cardName]
	return exists
}

// HasEvolution checks if a card has an evolution unlocked
func (ctx *PlayerContext) HasEvolution(cardName string) bool {
	return ctx.UnlockedEvolutions[cardName]
}

// GetAverageLevel calculates the average card level for a list of cards
// Only includes cards that exist in the player's collection
func (ctx *PlayerContext) GetAverageLevel(cardNames []string) float64 {
	if len(cardNames) == 0 {
		return 0.0
	}

	total := 0
	count := 0
	for _, name := range cardNames {
		if info, exists := ctx.Collection[name]; exists {
			total += info.Level
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return float64(total) / float64(count)
}

// IsCardUnlockedInArena checks if a card is available in the player's current arena
// This requires card unlock data which would typically come from game constants
// For now, this is a placeholder that always returns true
// TODO: Implement arena unlock checking in future task (Task 3.3.2)
func (ctx *PlayerContext) IsCardUnlockedInArena(cardName string) bool {
	// Placeholder implementation
	// Will be implemented in Task 3.3.2: Arena-aware card unlock validation
	return true
}

// CalculateUpgradeGap calculates how many levels a deck is below max
// Useful for ladder readiness assessment
func (ctx *PlayerContext) CalculateUpgradeGap(deckCards []deck.CardCandidate) int {
	totalGap := 0
	for _, card := range deckCards {
		if info, exists := ctx.Collection[card.Name]; exists {
			gap := info.MaxLevel - info.Level
			totalGap += gap
		}
	}
	return totalGap
}

// GetRarity returns the rarity of a card from the player's collection
func (ctx *PlayerContext) GetRarity(cardName string) string {
	if info, exists := ctx.Collection[cardName]; exists {
		return info.Rarity
	}
	return ""
}
