// Package whatif provides what-if analysis functionality for simulating deck changes
// with upgraded cards
package whatif

import (
	"fmt"
	"maps"
	"strconv"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// CardUpgrade represents a single card upgrade in a what-if scenario
type CardUpgrade struct {
	CardName  string
	FromLevel int
	ToLevel   int
	GoldCost  int
}

// WhatIfScenario represents a complete what-if analysis scenario
type WhatIfScenario struct {
	Name          string
	Description   string
	Upgrades      []CardUpgrade
	TotalGold     int
	OriginalDeck  *deck.DeckRecommendation
	SimulatedDeck *deck.DeckRecommendation
	Impact        SimulationImpact
}

// SimulationImpact quantifies the effect of upgrades on deck performance
type SimulationImpact struct {
	DeckScoreDelta       float64
	NewCardsInDeck       []string
	RemovedCards         []string
	ViabilityImprovement float64
	Recommendation       string
}

// WhatIfAnalyzer performs what-if analysis on card upgrades
type WhatIfAnalyzer struct {
	builder *deck.Builder
}

// NewWhatIfAnalyzer creates a new what-if analyzer
func NewWhatIfAnalyzer(builder *deck.Builder) *WhatIfAnalyzer {
	return &WhatIfAnalyzer{
		builder: builder,
	}
}

// ParseCardUpgrade parses a card upgrade specification string
// Format: "CardName:ToLevel" or "CardName:FromLevel:ToLevel"
// Returns a CardUpgrade that can be used with WhatIfAnalyzer
func ParseCardUpgrade(spec string) (CardUpgrade, error) {
	parts := strings.Split(spec, ":")
	if len(parts) < 2 {
		return CardUpgrade{}, fmt.Errorf("invalid upgrade spec format: %s (expected CardName:ToLevel or CardName:FromLevel:ToLevel)", spec)
	}

	cardName := parts[0]
	var fromLevel, toLevel int

	// Parse levels
	if len(parts) == 2 {
		// Format: CardName:ToLevel (from level will be determined from card data)
		toLevel = parseInt(parts[1])
		fromLevel = 0 // Will be inferred from card data
	} else {
		// Format: CardName:FromLevel:ToLevel
		fromLevel = parseInt(parts[1])
		toLevel = parseInt(parts[2])
	}

	if fromLevel < 0 || toLevel < fromLevel {
		return CardUpgrade{}, fmt.Errorf("invalid level range: %d -> %d", fromLevel, toLevel)
	}

	return CardUpgrade{
		CardName:  cardName,
		FromLevel: fromLevel,
		ToLevel:   toLevel,
		GoldCost:  0, // Will be calculated by the analyzer
	}, nil
}

// AnalyzeUpgradePath simulates upgrading specific cards and analyzes the impact
// Takes a map of card levels and applies the specified upgrades to see the effect
func (w *WhatIfAnalyzer) AnalyzeUpgradePath(
	cardLevels map[string]deck.CardLevelData,
	upgrades []CardUpgrade,
) (*WhatIfScenario, error) {
	// Build original deck
	originalAnalysis := deck.CardAnalysis{
		CardLevels: cardLevels,
	}
	originalDeck, err := w.builder.BuildDeckFromAnalysis(originalAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to build original deck: %w", err)
	}

	// Apply upgrades to card levels
	modifiedLevels := w.applyUpgradesToCardLevels(cardLevels, upgrades)

	// Build simulated deck with upgraded cards
	modifiedAnalysis := deck.CardAnalysis{
		CardLevels: modifiedLevels,
	}
	simulatedDeck, err := w.builder.BuildDeckFromAnalysis(modifiedAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to build simulated deck: %w", err)
	}

	// Calculate impact
	impact := w.calculateImpact(originalDeck, simulatedDeck, upgrades)

	// Calculate total gold cost
	totalGold := 0
	for _, u := range upgrades {
		totalGold += u.GoldCost
	}

	scenario := &WhatIfScenario{
		Name:          generateScenarioName(upgrades),
		Upgrades:      upgrades,
		TotalGold:     totalGold,
		OriginalDeck:  originalDeck,
		SimulatedDeck: simulatedDeck,
		Impact:        *impact,
	}

	return scenario, nil
}

// applyUpgradesToCardLevels creates a modified copy of card levels with upgrades applied
func (w *WhatIfAnalyzer) applyUpgradesToCardLevels(
	cardLevels map[string]deck.CardLevelData,
	upgrades []CardUpgrade,
) map[string]deck.CardLevelData {
	// Create a deep copy
	modified := make(map[string]deck.CardLevelData)
	maps.Copy(modified, cardLevels)

	// Apply each upgrade (use index to modify slice in place)
	for i := range upgrades {
		upgrade := &upgrades[i]
		if cardData, exists := modified[upgrade.CardName]; exists {
			// If from level is 0, use current level as from level
			fromLevel := upgrade.FromLevel
			if fromLevel == 0 {
				fromLevel = cardData.Level
				upgrade.FromLevel = fromLevel
			}

			modified[upgrade.CardName] = deck.CardLevelData{
				Level:             upgrade.ToLevel,
				MaxLevel:          cardData.MaxLevel,
				Rarity:            cardData.Rarity,
				Elixir:            cardData.Elixir,
				EvolutionLevel:    cardData.EvolutionLevel,
				MaxEvolutionLevel: cardData.MaxEvolutionLevel,
				ScoreBoost:        cardData.ScoreBoost,
			}

			// Calculate gold cost if not already set
			if upgrade.GoldCost == 0 {
				upgrade.GoldCost = calculateUpgradeCost(fromLevel, upgrade.ToLevel, cardData.Rarity)
			}
		}
	}

	return modified
}

// calculateImpact computes the impact of upgrades on the deck
func (w *WhatIfAnalyzer) calculateImpact(
	original, simulated *deck.DeckRecommendation,
	upgrades []CardUpgrade,
) *SimulationImpact {
	// Calculate deck score delta
	originalScore := calculateDeckScore(original)
	simulatedScore := calculateDeckScore(simulated)
	scoreDelta := simulatedScore - originalScore

	impact := &SimulationImpact{
		DeckScoreDelta: scoreDelta,
	}

	// Find new cards in deck
	originalCards := make(map[string]bool)
	for _, c := range original.Deck {
		originalCards[c] = true
	}

	for _, c := range simulated.Deck {
		if !originalCards[c] {
			impact.NewCardsInDeck = append(impact.NewCardsInDeck, c)
		}
	}

	// Find removed cards
	simulatedCards := make(map[string]bool)
	for _, c := range simulated.Deck {
		simulatedCards[c] = true
	}

	for _, c := range original.Deck {
		if !simulatedCards[c] {
			impact.RemovedCards = append(impact.RemovedCards, c)
		}
	}

	// Calculate viability improvement (percentage)
	if originalScore > 0 {
		impact.ViabilityImprovement = (scoreDelta / originalScore) * 100
	}

	// Generate recommendation
	impact.Recommendation = w.generateRecommendation(impact, upgrades)

	return impact
}

// generateRecommendation creates a human-readable recommendation
func (w *WhatIfAnalyzer) generateRecommendation(impact *SimulationImpact, upgrades []CardUpgrade) string {
	totalGold := 0
	for _, u := range upgrades {
		totalGold += u.GoldCost
	}

	if impact.DeckScoreDelta <= 0 {
		return fmt.Sprintf("These upgrades are not recommended. The simulated deck score decreased by %.2f points.", -impact.DeckScoreDelta)
	}

	if impact.ViabilityImprovement > 10 {
		return fmt.Sprintf("Highly recommended! These upgrades (%d gold) significantly improve your deck viability by %.1f%%.", totalGold, impact.ViabilityImprovement)
	}

	if impact.ViabilityImprovement > 5 {
		return fmt.Sprintf("Recommended. These upgrades (%d gold) moderately improve your deck viability by %.1f%%.", totalGold, impact.ViabilityImprovement)
	}

	return fmt.Sprintf("Minor improvement. These upgrades (%d gold) slightly improve your deck by %.1f%%. Consider prioritizing other upgrades.", totalGold, impact.ViabilityImprovement)
}

// calculateDeckScore computes an overall score for a deck recommendation
func calculateDeckScore(deck *deck.DeckRecommendation) float64 {
	if deck == nil || len(deck.DeckDetail) == 0 {
		return 0
	}

	total := 0.0
	for _, card := range deck.DeckDetail {
		total += card.Score
	}
	return total
}

// Helper functions

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// calculateUpgradeCost returns the gold cost for upgrading a card
// Uses simplified cost calculation based on rarity and level difference
func calculateUpgradeCost(fromLevel, toLevel int, rarity string) int {
	// Gold costs per level by rarity (simplified)
	costPerLevel := map[string]int{
		"Common":    100,
		"Rare":      1000,
		"Epic":      3000,
		"Legendary": 40000,
		"Champion":  50000,
	}

	baseCost := costPerLevel[rarity]
	if baseCost == 0 {
		baseCost = 1000
	}

	return baseCost * (toLevel - fromLevel)
}

func generateScenarioName(upgrades []CardUpgrade) string {
	if len(upgrades) == 1 {
		return fmt.Sprintf("Upgrade %s to Lv%d", upgrades[0].CardName, upgrades[0].ToLevel)
	}

	cardNames := make([]string, len(upgrades))
	for i, u := range upgrades {
		cardNames[i] = u.CardName
	}

	return fmt.Sprintf("Upgrade %d cards: %s", len(upgrades), strings.Join(cardNames, ", "))
}
