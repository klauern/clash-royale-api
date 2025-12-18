package deck

import (
	"fmt"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
)

// TargetLevelPolicy defines how to compute target levels for deck projection
type TargetLevelPolicy string

const (
	PolicyMaxAll       TargetLevelPolicy = "max_all"       // All cards to level 14
	PolicyMatchHighest TargetLevelPolicy = "match_highest" // Match highest card in deck
	PolicyTournament   TargetLevelPolicy = "tournament"    // Level 11 standard
	PolicyBudget       TargetLevelPolicy = "budget"        // Minimize upgrade cost
	PolicyCustom       TargetLevelPolicy = "custom"        // User-specified levels
)

// DeckProjection represents a deck with current and projected level states
type DeckProjection struct {
	Deck              []*CardCandidate  `json:"deck"`
	CurrentScore      float64           `json:"current_score"`
	ProjectedScore    float64           `json:"projected_score"`
	TargetLevelPolicy TargetLevelPolicy `json:"target_policy"`
	UpgradePath       *UpgradePath      `json:"upgrade_path"`
	ScoreImprovement  float64           `json:"score_improvement"` // Percentage improvement
	TargetLevels      map[string]int    `json:"target_levels"`     // Card name -> target level
}

// UpgradePath details the cost to reach projected levels
type UpgradePath struct {
	CardUpgrades  []CardUpgrade `json:"card_upgrades"`
	TotalCards    int           `json:"total_cards"`    // Total cards needed across all upgrades
	TotalGold     int           `json:"total_gold"`     // Total gold needed
	EstimatedDays int           `json:"estimated_days"` // Rough estimate based on typical progress
}

// CardUpgrade represents a single card's upgrade requirements
type CardUpgrade struct {
	CardName     string `json:"card_name"`
	Rarity       string `json:"rarity"`
	CurrentLevel int    `json:"current_level"`
	TargetLevel  int    `json:"target_level"`
	CardsNeeded  int    `json:"cards_needed"`
	GoldNeeded   int    `json:"gold_needed"`
}

// ProjectionComparison compares two deck projections
type ProjectionComparison struct {
	DeckA            string  `json:"deck_a"`
	DeckB            string  `json:"deck_b"`
	ScoreDifference  float64 `json:"score_difference"`
	CostDifference   int     `json:"cost_difference"`   // Gold cost difference
	ValueRatio       float64 `json:"value_ratio"`       // Score improvement per 1000 gold
	Recommendation   string  `json:"recommendation"`
}

// NewProjection creates a deck projection from card candidates
func NewProjection(deck []*CardCandidate, policy TargetLevelPolicy, customLevels map[string]int) *DeckProjection {
	if deck == nil || len(deck) == 0 {
		return nil
	}

	// Calculate current score
	currentScore := calculateDeckScore(deck)

	// Determine target levels based on policy
	targetLevels := determineTargetLevels(deck, policy, customLevels)

	// Create projected deck with target levels
	projectedDeck := make([]*CardCandidate, len(deck))
	for i, card := range deck {
		projected := *card // Copy the card
		if targetLevel, exists := targetLevels[card.Name]; exists {
			projected.Level = targetLevel
		}
		projectedDeck[i] = &projected
	}

	// Calculate projected score
	projectedScore := calculateDeckScore(projectedDeck)

	// Calculate score improvement percentage
	scoreImprovement := 0.0
	if currentScore > 0 {
		scoreImprovement = ((projectedScore - currentScore) / currentScore) * 100.0
	}

	projection := &DeckProjection{
		Deck:              deck,
		CurrentScore:      currentScore,
		ProjectedScore:    projectedScore,
		TargetLevelPolicy: policy,
		TargetLevels:      targetLevels,
		ScoreImprovement:  scoreImprovement,
	}

	// Calculate upgrade path
	projection.UpgradePath = projection.CalculateUpgradePath()

	return projection
}

// determineTargetLevels determines target level for each card based on policy
func determineTargetLevels(deck []*CardCandidate, policy TargetLevelPolicy, customLevels map[string]int) map[string]int {
	targetLevels := make(map[string]int)

	switch policy {
	case PolicyMaxAll:
		// All cards to 14
		for _, card := range deck {
			targetLevels[card.Name] = 14
		}

	case PolicyMatchHighest:
		// Find highest level in deck
		highestLevel := 0
		for _, card := range deck {
			if card.Level > highestLevel {
				highestLevel = card.Level
			}
		}
		// Set all cards to match highest
		for _, card := range deck {
			targetLevels[card.Name] = highestLevel
		}

	case PolicyTournament:
		// Tournament standard is level 11
		for _, card := range deck {
			targetLevels[card.Name] = 11
		}

	case PolicyBudget:
		// Minimize cost: just one level up for each card
		for _, card := range deck {
			nextLevel := card.Level + 1
			if nextLevel > card.MaxLevel {
				nextLevel = card.MaxLevel
			}
			targetLevels[card.Name] = nextLevel
		}

	case PolicyCustom:
		// Use provided custom levels
		if customLevels != nil {
			for _, card := range deck {
				if level, exists := customLevels[card.Name]; exists {
					targetLevels[card.Name] = level
				} else {
					// Default to current level if not specified
					targetLevels[card.Name] = card.Level
				}
			}
		}

	default:
		// Default to PolicyMatchHighest
		return determineTargetLevels(deck, PolicyMatchHighest, nil)
	}

	return targetLevels
}

// CalculateUpgradePath computes upgrade requirements for the projection
func (p *DeckProjection) CalculateUpgradePath() *UpgradePath {
	if p == nil || p.Deck == nil {
		return nil
	}

	cardUpgrades := make([]CardUpgrade, 0, len(p.Deck))
	totalCards := 0
	totalGold := 0

	for _, card := range p.Deck {
		targetLevel, exists := p.TargetLevels[card.Name]
		if !exists {
			targetLevel = card.Level // No upgrade needed
		}

		if targetLevel <= card.Level {
			continue // Already at or above target
		}

		// Calculate cards needed
		cardsNeeded := analysis.CalculateTotalCardsToMax(card.Level, card.Rarity)
		if targetLevel < card.MaxLevel {
			// Partial upgrade - need to recalculate
			cardsNeeded = 0
			for level := card.Level; level < targetLevel; level++ {
				cardsNeeded += analysis.CalculateCardsNeeded(level, card.Rarity)
			}
		}

		// Calculate gold needed
		goldNeeded := CalculateGoldNeeded(card.Level, targetLevel, card.Rarity)

		upgrade := CardUpgrade{
			CardName:     card.Name,
			Rarity:       card.Rarity,
			CurrentLevel: card.Level,
			TargetLevel:  targetLevel,
			CardsNeeded:  cardsNeeded,
			GoldNeeded:   goldNeeded,
		}

		cardUpgrades = append(cardUpgrades, upgrade)
		totalCards += cardsNeeded
		totalGold += goldNeeded
	}

	// Rough estimate: assume 500 cards/day from chests/donations
	estimatedDays := totalCards / 500
	if estimatedDays == 0 && totalCards > 0 {
		estimatedDays = 1
	}

	return &UpgradePath{
		CardUpgrades:  cardUpgrades,
		TotalCards:    totalCards,
		TotalGold:     totalGold,
		EstimatedDays: estimatedDays,
	}
}

// ScoreAtLevel calculates hypothetical deck score if all cards were at given level
func (p *DeckProjection) ScoreAtLevel(level int) float64 {
	if p == nil || p.Deck == nil {
		return 0.0
	}

	// Create hypothetical deck with all cards at specified level
	hypotheticalDeck := make([]*CardCandidate, len(p.Deck))
	for i, card := range p.Deck {
		hypothetical := *card // Copy
		hypothetical.Level = level
		if level > card.MaxLevel {
			hypothetical.Level = card.MaxLevel
		}
		hypotheticalDeck[i] = &hypothetical
	}

	return calculateDeckScore(hypotheticalDeck)
}

// SimulateUpgrade shows impact of upgrading a single card
func (p *DeckProjection) SimulateUpgrade(cardName string, newLevel int) *DeckProjection {
	if p == nil || p.Deck == nil {
		return nil
	}

	// Create modified deck with one card upgraded
	simulatedDeck := make([]*CardCandidate, len(p.Deck))
	for i, card := range p.Deck {
		simulated := *card // Copy
		if card.Name == cardName {
			simulated.Level = newLevel
			if newLevel > card.MaxLevel {
				simulated.Level = card.MaxLevel
			}
		}
		simulatedDeck[i] = &simulated
	}

	// Create new projection with same policy
	return NewProjection(simulatedDeck, p.TargetLevelPolicy, p.TargetLevels)
}

// CompareProjections compares two deck projections
func CompareProjections(a, b *DeckProjection) *ProjectionComparison {
	if a == nil || b == nil {
		return nil
	}

	deckAName := getDeckName(a.Deck)
	deckBName := getDeckName(b.Deck)

	scoreDiff := a.ProjectedScore - b.ProjectedScore
	costDiff := 0
	if a.UpgradePath != nil && b.UpgradePath != nil {
		costDiff = a.UpgradePath.TotalGold - b.UpgradePath.TotalGold
	}

	// Calculate value ratio: score improvement per 1000 gold
	valueA := 0.0
	valueB := 0.0
	if a.UpgradePath != nil && a.UpgradePath.TotalGold > 0 {
		valueA = (a.ScoreImprovement / float64(a.UpgradePath.TotalGold)) * 1000.0
	}
	if b.UpgradePath != nil && b.UpgradePath.TotalGold > 0 {
		valueB = (b.ScoreImprovement / float64(b.UpgradePath.TotalGold)) * 1000.0
	}

	// Make recommendation
	recommendation := "Equal value"
	if scoreDiff > 0.5 && costDiff < 0 {
		recommendation = fmt.Sprintf("%s is stronger and cheaper", deckAName)
	} else if scoreDiff > 0.5 {
		recommendation = fmt.Sprintf("%s is stronger (worth the extra cost)", deckAName)
	} else if scoreDiff < -0.5 && costDiff > 0 {
		recommendation = fmt.Sprintf("%s is stronger and cheaper", deckBName)
	} else if scoreDiff < -0.5 {
		recommendation = fmt.Sprintf("%s is stronger (worth the extra cost)", deckBName)
	} else if valueA > valueB {
		recommendation = fmt.Sprintf("%s has better value per gold spent", deckAName)
	} else if valueB > valueA {
		recommendation = fmt.Sprintf("%s has better value per gold spent", deckBName)
	}

	return &ProjectionComparison{
		DeckA:           deckAName,
		DeckB:           deckBName,
		ScoreDifference: scoreDiff,
		CostDifference:  costDiff,
		ValueRatio:      valueA - valueB,
		Recommendation:  recommendation,
	}
}

// calculateDeckScore calculates overall score for a deck
func calculateDeckScore(deck []*CardCandidate) float64 {
	if len(deck) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, card := range deck {
		totalScore += card.Score
	}

	// Return average score
	return totalScore / float64(len(deck))
}

// getDeckName creates a simple name for a deck based on its win condition
func getDeckName(deck []*CardCandidate) string {
	if len(deck) == 0 {
		return "Empty Deck"
	}

	// Find win condition (role = win_condition)
	for _, card := range deck {
		if card.Role != nil && *card.Role == RoleWinCondition {
			return fmt.Sprintf("%s Deck", card.Name)
		}
	}

	// Fallback to first card
	return fmt.Sprintf("%s Deck", deck[0].Name)
}
