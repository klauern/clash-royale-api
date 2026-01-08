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
	if ctx == nil {
		return false
	}
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
// Returns true if the player's arena level is >= the card's unlock arena
// Returns true if ArenaID is 0 (no arena restrictions - training mode)
func (ctx *PlayerContext) IsCardUnlockedInArena(cardName string) bool {
	// No arena restrictions if ArenaID is 0
	if ctx.ArenaID == 0 {
		return true
	}

	// Get the arena where this card unlocks
	unlockArena := getCardUnlockArena(cardName)

	// Card is unlocked if player has reached or passed the unlock arena
	return ctx.ArenaID >= unlockArena
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

// IsEvolutionAvailable checks if a card has evolution potential
// Returns true if the card has an evolution path available (MaxEvolutionLevel > 0)
func (ctx *PlayerContext) IsEvolutionAvailable(cardName string) bool {
	info, exists := ctx.Collection[cardName]
	if !exists {
		return false
	}
	return info.MaxEvolutionLevel > 0
}

// CanEvolve checks if a card is ready to evolve
// Returns true if the player has the card at sufficient level and count for evolution
func (ctx *PlayerContext) CanEvolve(cardName string) bool {
	info, exists := ctx.Collection[cardName]
	if !exists {
		return false
	}
	// Standard evolution requirements: level 10+ and sufficient cards
	return info.Level >= 10 && info.Count >= 5
}

// GetEvolutionProgress returns evolution progress for a card
// Returns current level, max level, current count, required count
func (ctx *PlayerContext) GetEvolutionProgress(cardName string) (currentLevel, maxLevel, currentCount, requiredCount int) {
	info, exists := ctx.Collection[cardName]
	if !exists {
		return 0, 0, 0, 5
	}
	return info.EvolutionLevel, info.MaxEvolutionLevel, info.Count, 5
}

// GetUnlockedEvolutionCards returns list of cards with unlocked evolutions
func (ctx *PlayerContext) GetUnlockedEvolutionCards() []string {
	if ctx == nil {
		return []string{}
	}
	evolved := []string{}
	for cardName := range ctx.UnlockedEvolutions {
		evolved = append(evolved, cardName)
	}
	return evolved
}

// GetEvolvableCards returns list of cards that could evolve but haven't
func (ctx *PlayerContext) GetEvolvableCards() []string {
	if ctx == nil {
		return []string{}
	}
	evolvable := []string{}
	for cardName, info := range ctx.Collection {
		if info.MaxEvolutionLevel > 0 && info.EvolutionLevel == 0 {
			evolvable = append(evolvable, cardName)
		}
	}
	return evolvable
}
