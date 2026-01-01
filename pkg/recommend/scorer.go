package recommend

import (
	"fmt"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// Scoring weights for overall score calculation have been moved to internal/config/constants.go

// Rarity weights for compatibility scoring
// Common cards are easier to max than rarer cards
var rarityWeights = map[string]float64{
	"Common":    config.GetRarityWeight("Common"),
	"Rare":      config.GetRarityWeight("Rare"),
	"Epic":      config.GetRarityWeight("Epic"),
	"Legendary": config.GetRarityWeight("Legendary"),
	"Champion":  config.GetRarityWeight("Champion"),
}

// Scorer handles scoring of deck recommendations
type Scorer struct {
	synergyDB *deck.SynergyDatabase
}

// NewScorer creates a new scorer with a synergy database
func NewScorer() *Scorer {
	return &Scorer{
		synergyDB: deck.NewSynergyDatabase(),
	}
}

// CalculateCompatibility measures how well player's card levels match a deck (0-100)
// Considers both card level ratios and rarity weights
func (s *Scorer) CalculateCompatibility(deckDetail []deck.CardDetail, playerCards map[string]deck.CardLevelData) float64 {
	if len(deckDetail) == 0 {
		return 0
	}

	totalScore := 0.0
	maxScore := 0.0

	for _, card := range deckDetail {
		maxScore += 1.0

		// Check if player owns this card
		playerCard, exists := playerCards[card.Name]
		if !exists {
			// Card not owned, 0 contribution
			continue
		}

		// Calculate level ratio (0.0 to 1.0)
		if playerCard.MaxLevel == 0 {
			continue
		}
		levelRatio := float64(playerCard.Level) / float64(playerCard.MaxLevel)

		// Get rarity weight
		rarityWeight := rarityWeights[card.Rarity]
		if rarityWeight == 0 {
			rarityWeight = 1.0 // Default weight
		}

		// Card score = level ratio * rarity weight
		cardScore := levelRatio * rarityWeight
		totalScore += cardScore
	}

	// Normalize to 0-100
	if maxScore == 0 {
		return 0
	}
	return (totalScore / maxScore) * 100.0
}

// CalculateSynergy measures card pair synergies within a deck (0-100)
// Uses the existing synergy database to find strong card combinations
func (s *Scorer) CalculateSynergy(deckNames []string) float64 {
	if len(deckNames) == 0 {
		return 0
	}

	analysis := s.synergyDB.AnalyzeDeckSynergy(deckNames)
	return analysis.TotalScore
}

// CalculateOverallScore combines all scoring factors into a single score (0-100)
// Formula: 60% compatibility + 25% synergy + 15% archetype fit
func (s *Scorer) CalculateOverallScore(compatibility, synergy, archetypeFit float64) float64 {
	return (compatibility * config.RecommendationWeightCompatibility) +
		(synergy * config.RecommendationWeightSynergy) +
		(archetypeFit * config.RecommendationWeightArchetypeFit)
}

// GenerateReasons creates human-readable reasons for why a deck is recommended
func (s *Scorer) GenerateReasons(rec *DeckRecommendation) []string {
	reasons := make([]string, 0)

	// Compatibility-based reasons
	if rec.CompatibilityScore >= 80 {
		reasons = append(reasons, "Excellent card level match - your cards are near max level")
	} else if rec.CompatibilityScore >= 60 {
		reasons = append(reasons, "Strong card levels - most cards are well-upgraded")
	} else if rec.CompatibilityScore >= 40 {
		reasons = append(reasons, "Decent card levels - some upgrades recommended")
	} else {
		reasons = append(reasons, "Consider upgrading key cards for better ladder performance")
	}

	// Synergy-based reasons
	if rec.SynergyScore >= 70 {
		reasons = append(reasons, "Excellent card synergies - cards work well together")
	} else if rec.SynergyScore >= 50 {
		reasons = append(reasons, "Good synergy between key cards")
	}

	// Archetype-specific reasons
	switch rec.Archetype {
	case "cycle":
		if rec.Deck.AvgElixir <= 3.0 {
			reasons = append(reasons, "Ultra-low elixir for fast cycling and constant pressure")
		} else if rec.Deck.AvgElixir <= 3.5 {
			reasons = append(reasons, "Low elixir cost supports cycle archetype playstyle")
		}
	case "beatdown":
		if rec.Deck.AvgElixir >= 4.0 {
			reasons = append(reasons, "High elixir beatdown deck - play patiently and build big pushes")
		}
	case "bait":
		reasons = append(reasons, "Bait archetype - threatens multiple spells to create openings")
	case "control":
		reasons = append(reasons, "Control archetype - reactively counters opponent plays")
	}

	// Type-specific reasons
	if rec.Type == TypeCustomVariation {
		reasons = append(reasons, "Custom variation optimized for your card collection")
	} else {
		reasons = append(reasons, fmt.Sprintf("Proven %s archetype with established win conditions", rec.ArchetypeName))
	}

	return reasons
}

// CardLevelData represents the level information for a card in player's collection
// This is an alias to deck.CardLevelData for convenience
type CardLevelData = deck.CardLevelData
