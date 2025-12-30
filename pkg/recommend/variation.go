package recommend

import (
	"fmt"
	"sort"

	"github.com/klauer/clash-royale-api/go/pkg/archetypes"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
)

// VariationGenerator creates custom deck variations by swapping weak cards
type VariationGenerator struct {
	archetypeConstraints map[mulligan.Archetype]archetypes.ArchetypeConstraints
}

// NewVariationGenerator creates a new variation generator
func NewVariationGenerator() *VariationGenerator {
	return &VariationGenerator{
		archetypeConstraints: archetypes.GetArchetypeConstraints(),
	}
}

// WeakCard identifies a card that could be replaced
type WeakCard struct {
	Name       string
	LevelRatio float64
	Role       string
	CardIndex  int // Index in deck
}

// GenerateVariations creates custom deck variations by swapping weak cards
// for better alternatives from the player's collection
func (vg *VariationGenerator) GenerateVariations(
	baseDeck *deck.DeckRecommendation,
	archetype mulligan.Archetype,
	analysis deck.CardAnalysis,
	maxVariations int,
) []*DeckRecommendation {
	variations := make([]*DeckRecommendation, 0)

	// Get archetype constraints
	constraints, exists := vg.archetypeConstraints[archetype]
	if !exists {
		return variations // No constraints, can't generate variations
	}

	// Find weak cards in the base deck
	weakCards := vg.findWeakCards(baseDeck, analysis, 2)
	if len(weakCards) == 0 {
		return variations // No weak cards to replace
	}

	// Generate variations by swapping weak cards
	for i, weakCard := range weakCards {
		if i >= maxVariations {
			break
		}

		// Find better alternatives
		alternatives := vg.findBetterAlternatives(weakCard, baseDeck, constraints, analysis)
		if len(alternatives) == 0 {
			continue // No better alternatives found
		}

		// Create variation with best alternative
		bestAlt := alternatives[0]
		variation := vg.createVariation(baseDeck, weakCard, bestAlt, archetype)
		variations = append(variations, variation)
	}

	return variations
}

// findWeakCards identifies the weakest cards in a deck based on level ratio
func (vg *VariationGenerator) findWeakCards(
	baseDeck *deck.DeckRecommendation,
	analysis deck.CardAnalysis,
	maxCount int,
) []WeakCard {
	weakCards := make([]WeakCard, 0)

	// Calculate level ratio for each card
	for i, card := range baseDeck.DeckDetail {
		playerCard, exists := analysis.CardLevels[card.Name]
		if !exists {
			// Card not owned, it's definitely weak
			weakCards = append(weakCards, WeakCard{
				Name:       card.Name,
				LevelRatio: 0.0,
				Role:       card.Role,
				CardIndex:  i,
			})
			continue
		}

		levelRatio := float64(playerCard.Level) / float64(playerCard.MaxLevel)

		// Consider cards weak if below 50% level
		if levelRatio < 0.5 {
			weakCards = append(weakCards, WeakCard{
				Name:       card.Name,
				LevelRatio: levelRatio,
				Role:       card.Role,
				CardIndex:  i,
			})
		}
	}

	// Sort by level ratio (ascending - weakest first)
	sort.Slice(weakCards, func(i, j int) bool {
		return weakCards[i].LevelRatio < weakCards[j].LevelRatio
	})

	// Limit to max count
	if len(weakCards) > maxCount {
		weakCards = weakCards[:maxCount]
	}

	return weakCards
}

// CardAlternative represents a potential replacement card
type CardAlternative struct {
	Name       string
	LevelRatio float64
	Role       string
	Score      float64
}

// findBetterAlternatives finds cards that are better than the weak card
// and maintain archetype constraints
func (vg *VariationGenerator) findBetterAlternatives(
	weakCard WeakCard,
	baseDeck *deck.DeckRecommendation,
	constraints archetypes.ArchetypeConstraints,
	analysis deck.CardAnalysis,
) []CardAlternative {
	alternatives := make([]CardAlternative, 0)

	// Create excluded cards map (base deck + archetype excluded)
	excludedMap := make(map[string]bool)
	for _, cardName := range baseDeck.Deck {
		excludedMap[cardName] = true
	}
	for _, cardName := range constraints.ExcludedCards {
		excludedMap[cardName] = true
	}

	// Find better cards from player's collection
	for cardName, cardData := range analysis.CardLevels {
		// Skip if already in deck or excluded
		if excludedMap[cardName] {
			continue
		}

		levelRatio := float64(cardData.Level) / float64(cardData.MaxLevel)

		// Must be significantly better than weak card
		if levelRatio <= weakCard.LevelRatio+0.1 {
			continue
		}

		// Check elixir range
		if cardData.Elixir > 0 {
			newAvgElixir := vg.calculateNewAvgElixir(baseDeck, weakCard.CardIndex, cardData.Elixir)
			if newAvgElixir < constraints.MinElixir || newAvgElixir > constraints.MaxElixir {
				continue // Would violate elixir constraints
			}
		}

		// Calculate score for this alternative
		score := levelRatio

		// Bonus for preferred cards
		if vg.isPreferred(cardName, constraints.PreferredCards) {
			score += 0.2
		}

		// Bonus for same role (maintains deck balance)
		// Classify the card to get its role
		cardRole := deck.ClassifyCard(cardName, cardData.Elixir)
		if weakCard.Role != "" && cardRole != nil && string(*cardRole) == weakCard.Role {
			score += 0.1
		}

		// Convert role to string for storage
		roleStr := ""
		if cardRole != nil {
			roleStr = string(*cardRole)
		}

		alternatives = append(alternatives, CardAlternative{
			Name:       cardName,
			LevelRatio: levelRatio,
			Role:       roleStr,
			Score:      score,
		})
	}

	// Sort by score (descending)
	sort.Slice(alternatives, func(i, j int) bool {
		return alternatives[i].Score > alternatives[j].Score
	})

	return alternatives
}

// calculateNewAvgElixir calculates what the average elixir would be after a swap
func (vg *VariationGenerator) calculateNewAvgElixir(
	baseDeck *deck.DeckRecommendation,
	cardIndex int,
	newCardElixir int,
) float64 {
	totalElixir := 0
	for i, card := range baseDeck.DeckDetail {
		if i == cardIndex {
			totalElixir += newCardElixir
		} else {
			totalElixir += card.Elixir
		}
	}
	return float64(totalElixir) / 8.0
}

// isPreferred checks if a card is in the preferred list
func (vg *VariationGenerator) isPreferred(cardName string, preferredCards []string) bool {
	for _, preferred := range preferredCards {
		if cardName == preferred {
			return true
		}
	}
	return false
}

// createVariation creates a new deck recommendation with the card swap
func (vg *VariationGenerator) createVariation(
	baseDeck *deck.DeckRecommendation,
	weakCard WeakCard,
	alternative CardAlternative,
	archetype mulligan.Archetype,
) *DeckRecommendation {
	// Deep copy the base deck
	newDeck := &deck.DeckRecommendation{
		Deck:           make([]string, 8),
		DeckDetail:     make([]deck.CardDetail, 8),
		AvgElixir:      baseDeck.AvgElixir,
		AnalysisTime:   baseDeck.AnalysisTime,
		Notes:          append([]string{}, baseDeck.Notes...),
		EvolutionSlots: append([]string{}, baseDeck.EvolutionSlots...),
	}

	copy(newDeck.Deck, baseDeck.Deck)
	copy(newDeck.DeckDetail, baseDeck.DeckDetail)

	// Swap the card
	newDeck.Deck[weakCard.CardIndex] = alternative.Name

	// Update card detail (we only have limited info)
	newDeck.DeckDetail[weakCard.CardIndex] = deck.CardDetail{
		Name: alternative.Name,
		Role: alternative.Role,
	}

	// Recalculate average elixir
	totalElixir := 0
	for _, card := range newDeck.DeckDetail {
		totalElixir += card.Elixir
	}
	newDeck.AvgElixir = float64(totalElixir) / 8.0

	// Add note about the variation
	note := fmt.Sprintf("Custom variation: %s → %s (better card level)", weakCard.Name, alternative.Name)
	newDeck.Notes = append(newDeck.Notes, note)

	return &DeckRecommendation{
		Deck:          newDeck,
		Archetype:     archetype,
		ArchetypeName: string(archetype),
		Type:          TypeCustomVariation,
	}
}
