// Package deck provides intelligent deck building functionality for Clash Royale
package deck

import (
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
)

// ConvertAnalysisForDeckBuilding converts analysis.CardAnalysis to the deck builder's CardAnalysis format
// This allows the deck builder to work with output from the card collection analyzer
func ConvertAnalysisForDeckBuilding(analysisData *analysis.CardAnalysis) CardAnalysis {
	cardLevels := make(map[string]CardLevelData)

	for cardName, cardInfo := range analysisData.CardLevels {
		cardLevels[cardName] = CardLevelData{
			Level:             cardInfo.Level,
			MaxLevel:          cardInfo.MaxLevel,
			Rarity:            cardInfo.Rarity,
			Elixir:            cardInfo.Elixir,
			MaxEvolutionLevel: cardInfo.MaxEvolutionLevel,
		}
	}

	return CardAnalysis{
		CardLevels:   cardLevels,
		AnalysisTime: analysisData.AnalysisTime.Format(time.RFC3339),
	}
}

// BuildDeckFromPlayerAnalysis is a convenience function that converts and builds a deck in one call
// This is the recommended way to build decks from player analysis data
func (b *Builder) BuildDeckFromPlayerAnalysis(analysisData *analysis.CardAnalysis) (*DeckRecommendation, error) {
	deckAnalysis := ConvertAnalysisForDeckBuilding(analysisData)
	return b.BuildDeckFromAnalysis(deckAnalysis)
}
