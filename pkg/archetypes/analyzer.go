package archetypes

import (
	"fmt"
	"sort"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
)

// Analyzer performs archetype variety analysis.
// It generates decks for all 8 archetypes and calculates upgrade costs.
type Analyzer struct {
	builder *ArchetypeBuilder
}

// NewAnalyzer creates a new archetype analyzer
func NewAnalyzer(dataDir string) *Analyzer {
	return &Analyzer{
		builder: NewArchetypeBuilder(dataDir),
	}
}

// AnalyzeArchetypes generates decks for all archetypes and calculates costs.
// It returns a complete analysis result with decks sorted by the specified criteria.
func (a *Analyzer) AnalyzeArchetypes(
	playerTag string,
	playerName string,
	analysis deck.CardAnalysis,
	targetLevel int,
) (*ArchetypeAnalysisResult, error) {
	result := &ArchetypeAnalysisResult{
		PlayerTag:    playerTag,
		PlayerName:   playerName,
		TargetLevel:  targetLevel,
		Archetypes:   make([]ArchetypeDeck, 0, 8),
		AnalysisTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Generate deck for each archetype
	for _, archetype := range GetAllArchetypes() {
		archetypeDeck, err := a.buildArchetypeDeck(archetype, analysis, targetLevel)
		if err != nil {
			// Log error but continue with other archetypes
			// Some archetypes may not be buildable with limited collections
			continue
		}

		result.Archetypes = append(result.Archetypes, *archetypeDeck)
	}

	if len(result.Archetypes) == 0 {
		return nil, fmt.Errorf("failed to generate any viable archetype decks - collection may be too limited")
	}

	return result, nil
}

// buildArchetypeDeck builds and analyzes a single archetype
func (a *Analyzer) buildArchetypeDeck(
	archetype mulligan.Archetype,
	analysis deck.CardAnalysis,
	targetLevel int,
) (*ArchetypeDeck, error) {
	// Build deck for archetype
	recommendation, err := a.builder.BuildForArchetype(archetype, analysis)
	if err != nil {
		return nil, err
	}

	// Calculate current average level
	currentAvgLevel := calculateAvgLevel(recommendation.DeckDetail)

	// Calculate upgrade costs
	upgrades, totalCards, totalGold := CalculateUpgradeCosts(recommendation, targetLevel)

	// Calculate distance metric
	distance := CalculateDistanceMetric(recommendation, targetLevel)

	return &ArchetypeDeck{
		Archetype:       archetype,
		Deck:            recommendation.Deck,
		DeckDetail:      recommendation.DeckDetail,
		AvgElixir:       recommendation.AvgElixir,
		CurrentAvgLevel: currentAvgLevel,
		TargetLevel:     targetLevel,
		CardsNeeded:     totalCards,
		GoldNeeded:      totalGold,
		DistanceMetric:  distance,
		UpgradeDetails:  upgrades,
	}, nil
}

// SortBy sorts the archetype analysis by the specified criteria
func (result *ArchetypeAnalysisResult) SortBy(sortBy SortBy) {
	switch sortBy {
	case SortByDistance:
		// Sort by distance (lower is better)
		sort.Slice(result.Archetypes, func(i, j int) bool {
			return result.Archetypes[i].DistanceMetric < result.Archetypes[j].DistanceMetric
		})
	case SortByCardsNeeded:
		// Sort by upgrade cost (fewer cards is better)
		sort.Slice(result.Archetypes, func(i, j int) bool {
			return result.Archetypes[i].CardsNeeded < result.Archetypes[j].CardsNeeded
		})
	case SortByAvgLevel:
		// Sort by current strength (higher level is better)
		sort.Slice(result.Archetypes, func(i, j int) bool {
			return result.Archetypes[i].CurrentAvgLevel > result.Archetypes[j].CurrentAvgLevel
		})
	}
}
