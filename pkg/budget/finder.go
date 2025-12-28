// Package budget provides budget-optimized deck finding functionality.
package budget

import (
	"fmt"
	"math"
	"sort"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// Finder analyzes decks for budget optimization
type Finder struct {
	builder *deck.Builder
	options BudgetFinderOptions
}

// NewFinder creates a new budget-optimized deck finder
func NewFinder(dataDir string, options BudgetFinderOptions) *Finder {
	return &Finder{
		builder: deck.NewBuilder(dataDir),
		options: options,
	}
}

// SetUnlockedEvolutions updates the unlocked evolutions list
func (f *Finder) SetUnlockedEvolutions(cards []string) {
	f.builder.SetUnlockedEvolutions(cards)
}

// SetEvolutionSlotLimit updates the evolution slot limit
func (f *Finder) SetEvolutionSlotLimit(limit int) {
	f.builder.SetEvolutionSlotLimit(limit)
}

// AnalyzeDeck analyzes a single deck for budget optimization
func (f *Finder) AnalyzeDeck(deckRec *deck.DeckRecommendation, cardLevels map[string]deck.CardLevelData) *DeckBudgetAnalysis {
	if deckRec == nil || len(deckRec.DeckDetail) == 0 {
		return nil
	}

	// Calculate current deck metrics
	totalCardsNeeded := 0
	totalGoldNeeded := 0
	totalCurrentLevel := 0
	totalTargetLevel := 0
	upgradesNeeded := 0
	cardUpgrades := make([]CardUpgradeDetail, 0)

	for _, card := range deckRec.DeckDetail {
		currentLevel := card.Level
		targetLevel := card.MaxLevel
		if targetLevel == 0 {
			targetLevel = 14 // Default max level
		}

		totalCurrentLevel += currentLevel
		totalTargetLevel += targetLevel

		if currentLevel < targetLevel {
			// Calculate cards needed for this upgrade
			cardsNeeded := 0
			for level := currentLevel; level < targetLevel; level++ {
				cardsNeeded += calculateCardsNeededForLevel(level, card.Rarity)
			}

			// Calculate gold needed
			goldNeeded := deck.CalculateGoldNeeded(currentLevel, targetLevel, card.Rarity)

			if cardsNeeded > 0 {
				upgradesNeeded++
				totalCardsNeeded += cardsNeeded
				totalGoldNeeded += goldNeeded

				// Calculate priority based on how close to next level and importance
				priority := calculateUpgradePriority(card, cardsNeeded)

				cardUpgrades = append(cardUpgrades, CardUpgradeDetail{
					CardName:     card.Name,
					CurrentLevel: currentLevel,
					TargetLevel:  targetLevel,
					CardsNeeded:  cardsNeeded,
					GoldNeeded:   goldNeeded,
					Priority:     priority,
				})
			}
		}
	}

	// Sort card upgrades by priority (highest first)
	sort.Slice(cardUpgrades, func(i, j int) bool {
		return cardUpgrades[i].Priority > cardUpgrades[j].Priority
	})

	// Calculate current and projected scores
	currentScore := calculateDeckScore(deckRec.DeckDetail)
	projectedScore := calculateProjectedScore(deckRec.DeckDetail)

	// Calculate average levels
	avgCurrentLevel := float64(totalCurrentLevel) / float64(len(deckRec.DeckDetail))
	_ = float64(totalTargetLevel) / float64(len(deckRec.DeckDetail)) // avgTargetLevel for future use

	// Calculate viability gap (how far from target average level)
	viabilityGap := f.options.TargetAverageLevel - avgCurrentLevel
	if viabilityGap < 0 {
		viabilityGap = 0
	}

	// Calculate ROI (Return on Investment)
	roi := 0.0
	if totalCardsNeeded > 0 {
		scoreImprovement := projectedScore - currentScore
		roi = scoreImprovement / float64(totalCardsNeeded) * 1000 // Scale for readability
	} else if currentScore > 0 {
		roi = math.MaxFloat64 // Already optimal
	}

	// Calculate cost efficiency
	costEfficiency := currentScore
	if totalCardsNeeded > 0 {
		costEfficiency = currentScore / (1 + math.Log10(float64(totalCardsNeeded)))
	}

	// Determine if this is a "quick win"
	isQuickWin := upgradesNeeded <= f.options.QuickWinMaxUpgrades &&
		totalCardsNeeded <= f.options.QuickWinMaxCards &&
		avgCurrentLevel >= f.options.TargetAverageLevel-1.0

	// Determine budget category
	budgetCategory := categorizeDeck(totalCardsNeeded, upgradesNeeded, avgCurrentLevel, f.options)

	return &DeckBudgetAnalysis{
		Deck:             deckRec,
		CurrentScore:     currentScore,
		ProjectedScore:   projectedScore,
		TotalCardsNeeded: totalCardsNeeded,
		TotalGoldNeeded:  totalGoldNeeded,
		ROI:              roi,
		CostEfficiency:   costEfficiency,
		UpgradesNeeded:   upgradesNeeded,
		CardUpgrades:     cardUpgrades,
		IsQuickWin:       isQuickWin,
		ViabilityGap:     viabilityGap,
		BudgetCategory:   budgetCategory,
	}
}

// FindOptimalDecks generates and analyzes multiple deck variations to find budget-optimal options
func (f *Finder) FindOptimalDecks(cardAnalysis deck.CardAnalysis, playerTag, playerName string) (*BudgetFinderResult, error) {
	if len(cardAnalysis.CardLevels) == 0 {
		return nil, fmt.Errorf("no card data available for analysis")
	}

	result := &BudgetFinderResult{
		PlayerTag:    playerTag,
		PlayerName:   playerName,
		AllDecks:     make([]*DeckBudgetAnalysis, 0),
		BestROIDecks: make([]*DeckBudgetAnalysis, 0),
		QuickWins:    make([]*DeckBudgetAnalysis, 0),
		ReadyDecks:   make([]*DeckBudgetAnalysis, 0),
		WithinBudget: make([]*DeckBudgetAnalysis, 0),
	}

	// Generate the primary deck
	primaryDeck, err := f.builder.BuildDeckFromAnalysis(cardAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to build primary deck: %w", err)
	}

	// Analyze primary deck
	primaryAnalysis := f.AnalyzeDeck(primaryDeck, cardAnalysis.CardLevels)
	if primaryAnalysis != nil {
		result.AllDecks = append(result.AllDecks, primaryAnalysis)
	}

	// Generate variations if requested
	if f.options.IncludeVariations {
		variations := f.generateVariations(cardAnalysis, f.options.MaxVariations)
		for _, variation := range variations {
			varAnalysis := f.AnalyzeDeck(variation, cardAnalysis.CardLevels)
			if varAnalysis != nil {
				result.AllDecks = append(result.AllDecks, varAnalysis)
			}
		}
	}

	// Filter and categorize results
	f.categorizeResults(result)

	// Sort results by chosen criteria
	f.sortResults(result)

	// Apply top N limit
	if f.options.TopN > 0 && len(result.AllDecks) > f.options.TopN {
		result.AllDecks = result.AllDecks[:f.options.TopN]
	}

	// Calculate summary statistics
	result.Summary = f.calculateSummary(result, cardAnalysis)

	return result, nil
}

// generateVariations creates deck variations by swapping cards
func (f *Finder) generateVariations(cardAnalysis deck.CardAnalysis, maxVariations int) []*deck.DeckRecommendation {
	variations := make([]*deck.DeckRecommendation, 0, maxVariations)

	// Strategy 1: Prioritize highest-level cards only
	highLevelAnalysis := filterByMinLevel(cardAnalysis, 12)
	if len(highLevelAnalysis.CardLevels) >= 8 {
		if variation, err := f.builder.BuildDeckFromAnalysis(highLevelAnalysis); err == nil {
			variation.Notes = append(variation.Notes, "High-level card priority variation")
			variations = append(variations, variation)
		}
	}

	// Strategy 2: Budget-friendly (prioritize common/rare cards)
	budgetAnalysis := filterByRarity(cardAnalysis, []string{"Common", "Rare"})
	if len(budgetAnalysis.CardLevels) >= 8 {
		if variation, err := f.builder.BuildDeckFromAnalysis(budgetAnalysis); err == nil {
			variation.Notes = append(variation.Notes, "Budget-friendly (Common/Rare focus)")
			variations = append(variations, variation)
		}
	}

	// Strategy 3: Near-max cards only (level 13+)
	nearMaxAnalysis := filterByMinLevel(cardAnalysis, 13)
	if len(nearMaxAnalysis.CardLevels) >= 8 {
		if variation, err := f.builder.BuildDeckFromAnalysis(nearMaxAnalysis); err == nil {
			variation.Notes = append(variation.Notes, "Near-max cards only")
			variations = append(variations, variation)
		}
	}

	// Limit variations
	if len(variations) > maxVariations {
		variations = variations[:maxVariations]
	}

	return variations
}

// categorizeResults categorizes decks into ready, quick win, and within budget
func (f *Finder) categorizeResults(result *BudgetFinderResult) {
	for _, analysis := range result.AllDecks {
		// Check if within budget constraints
		withinBudget := true
		if f.options.MaxCardsNeeded > 0 && analysis.TotalCardsNeeded > f.options.MaxCardsNeeded {
			withinBudget = false
		}
		if f.options.MaxGoldNeeded > 0 && analysis.TotalGoldNeeded > f.options.MaxGoldNeeded {
			withinBudget = false
		}

		if withinBudget {
			result.WithinBudget = append(result.WithinBudget, analysis)
		}

		// Categorize by budget category
		switch analysis.BudgetCategory {
		case CategoryReady:
			result.ReadyDecks = append(result.ReadyDecks, analysis)
		case CategoryQuickWin:
			result.QuickWins = append(result.QuickWins, analysis)
		}

		// Add to best ROI if ROI is positive
		if analysis.ROI > 0 {
			result.BestROIDecks = append(result.BestROIDecks, analysis)
		}
	}

	// Sort best ROI decks by ROI
	sort.Slice(result.BestROIDecks, func(i, j int) bool {
		return result.BestROIDecks[i].ROI > result.BestROIDecks[j].ROI
	})

	// Limit best ROI decks
	if len(result.BestROIDecks) > 5 {
		result.BestROIDecks = result.BestROIDecks[:5]
	}
}

// sortResults sorts all decks by the chosen criteria
func (f *Finder) sortResults(result *BudgetFinderResult) {
	sort.Slice(result.AllDecks, func(i, j int) bool {
		a, b := result.AllDecks[i], result.AllDecks[j]
		switch f.options.SortBy {
		case SortByROI:
			return a.ROI > b.ROI
		case SortByCostEfficiency:
			return a.CostEfficiency > b.CostEfficiency
		case SortByTotalCards:
			return a.TotalCardsNeeded < b.TotalCardsNeeded
		case SortByTotalGold:
			return a.TotalGoldNeeded < b.TotalGoldNeeded
		case SortByCurrentScore:
			return a.CurrentScore > b.CurrentScore
		case SortByProjectedScore:
			return a.ProjectedScore > b.ProjectedScore
		default:
			return a.ROI > b.ROI
		}
	})
}

// calculateSummary computes aggregate statistics
func (f *Finder) calculateSummary(result *BudgetFinderResult, cardAnalysis deck.CardAnalysis) BudgetSummary {
	summary := BudgetSummary{
		TotalDecksAnalyzed: len(result.AllDecks),
		ReadyDeckCount:     len(result.ReadyDecks),
		QuickWinCount:      len(result.QuickWins),
		LowestCardsNeeded:  math.MaxInt32,
	}

	totalCards := 0
	totalGold := 0
	totalLevel := 0
	cardCount := 0

	for _, analysis := range result.AllDecks {
		totalCards += analysis.TotalCardsNeeded
		totalGold += analysis.TotalGoldNeeded

		if analysis.ROI > summary.BestROI {
			summary.BestROI = analysis.ROI
		}
		if analysis.CostEfficiency > summary.BestCostEfficiency {
			summary.BestCostEfficiency = analysis.CostEfficiency
		}
		if analysis.TotalCardsNeeded < summary.LowestCardsNeeded {
			summary.LowestCardsNeeded = analysis.TotalCardsNeeded
		}
	}

	if len(result.AllDecks) > 0 {
		summary.AverageCardsNeeded = totalCards / len(result.AllDecks)
		summary.AverageGoldNeeded = totalGold / len(result.AllDecks)
	}

	// Calculate player's average card level
	for _, card := range cardAnalysis.CardLevels {
		totalLevel += card.Level
		cardCount++
	}
	if cardCount > 0 {
		summary.PlayerAverageLevel = float64(totalLevel) / float64(cardCount)
	}

	if summary.LowestCardsNeeded == math.MaxInt32 {
		summary.LowestCardsNeeded = 0
	}

	return summary
}

// Helper functions

// calculateDeckScore calculates the current score of a deck based on card levels
func calculateDeckScore(cards []deck.CardDetail) float64 {
	if len(cards) == 0 {
		return 0
	}

	totalScore := 0.0
	for _, card := range cards {
		// Score based on level ratio and card's inherent score
		levelRatio := float64(card.Level) / float64(card.MaxLevel)
		totalScore += card.Score * levelRatio
	}

	return totalScore / float64(len(cards))
}

// calculateProjectedScore calculates projected score with all cards at max level
func calculateProjectedScore(cards []deck.CardDetail) float64 {
	if len(cards) == 0 {
		return 0
	}

	totalScore := 0.0
	for _, card := range cards {
		// At max level, level ratio is 1.0
		totalScore += card.Score
	}

	return totalScore / float64(len(cards))
}

// calculateUpgradePriority determines upgrade priority for a card
func calculateUpgradePriority(card deck.CardDetail, cardsNeeded int) float64 {
	// Priority factors:
	// 1. How close to next level (fewer cards = higher priority)
	// 2. Card role importance
	// 3. Current level (higher level cards more impactful)

	// Inverse of cards needed (closer = higher priority)
	closeness := 1.0 / (1.0 + math.Log10(float64(cardsNeeded+1)))

	// Level factor (higher level = more impact)
	levelFactor := float64(card.Level) / float64(card.MaxLevel)

	// Role importance
	roleBonus := 0.0
	switch card.Role {
	case "win_conditions":
		roleBonus = 0.3
	case "spells_big":
		roleBonus = 0.2
	case "buildings":
		roleBonus = 0.15
	case "support":
		roleBonus = 0.1
	}

	return (closeness * 0.5) + (levelFactor * 0.3) + roleBonus
}

// categorizeDeck determines the budget category for a deck
func categorizeDeck(totalCardsNeeded, upgradesNeeded int, avgLevel float64, options BudgetFinderOptions) BudgetCategory {
	// Ready: already at or above target level with minimal upgrades
	if avgLevel >= options.TargetAverageLevel && totalCardsNeeded < 100 {
		return CategoryReady
	}

	// Quick win: just 1-2 upgrades away
	if upgradesNeeded <= options.QuickWinMaxUpgrades && totalCardsNeeded <= options.QuickWinMaxCards {
		return CategoryQuickWin
	}

	// Medium investment: moderate work needed
	if totalCardsNeeded <= 5000 && upgradesNeeded <= 5 {
		return CategoryMediumInvestment
	}

	// Long term: significant investment required
	return CategoryLongTerm
}

// filterByMinLevel creates a copy of analysis with only cards at or above min level
func filterByMinLevel(original deck.CardAnalysis, minLevel int) deck.CardAnalysis {
	filtered := deck.CardAnalysis{
		CardLevels:   make(map[string]deck.CardLevelData),
		AnalysisTime: original.AnalysisTime,
	}

	for name, data := range original.CardLevels {
		if data.Level >= minLevel {
			filtered.CardLevels[name] = data
		}
	}

	return filtered
}

// filterByRarity creates a copy of analysis with only cards of specified rarities
func filterByRarity(original deck.CardAnalysis, rarities []string) deck.CardAnalysis {
	raritySet := make(map[string]bool)
	for _, r := range rarities {
		raritySet[r] = true
	}

	filtered := deck.CardAnalysis{
		CardLevels:   make(map[string]deck.CardLevelData),
		AnalysisTime: original.AnalysisTime,
	}

	for name, data := range original.CardLevels {
		if raritySet[data.Rarity] {
			filtered.CardLevels[name] = data
		}
	}

	return filtered
}

// calculateCardsNeededForLevel returns how many cards are needed to upgrade from currentLevel
// This is a local implementation to avoid circular import with pkg/analysis
func calculateCardsNeededForLevel(currentLevel int, rarity string) int {
	// Upgrade costs by rarity and level (based on Clash Royale card progression)
	upgradeCosts := map[string]map[int]int{
		"Common": {
			1: 2, 2: 4, 3: 10, 4: 20, 5: 50, 6: 100, 7: 200, 8: 400,
			9: 800, 10: 1000, 11: 2000, 12: 3000, 13: 2500, 14: 3500, 15: 5500,
		},
		"Rare": {
			1: 2, 2: 2, 3: 2, 4: 4, 5: 10, 6: 20, 7: 50, 8: 100,
			9: 200, 10: 300, 11: 400, 12: 400, 13: 550, 14: 750, 15: 1000,
		},
		"Epic": {
			1: 2, 2: 2, 3: 2, 4: 2, 5: 2, 6: 2, 7: 4, 8: 10,
			9: 20, 10: 50, 11: 30, 12: 40, 13: 70, 14: 100, 15: 140,
		},
		"Legendary": {
			1: 2, 2: 2, 3: 2, 4: 2, 5: 2, 6: 2, 7: 2, 8: 2,
			9: 2, 10: 4, 11: 10, 12: 20, 13: 10, 14: 12, 15: 15,
		},
		"Champion": {
			1: 2, 11: 2, 12: 4, 13: 8, 14: 10, 15: 12,
		},
	}

	costs, exists := upgradeCosts[rarity]
	if !exists {
		return 0
	}

	cardsNeeded, exists := costs[currentLevel]
	if !exists {
		return 2 // Default fallback
	}

	return cardsNeeded
}
