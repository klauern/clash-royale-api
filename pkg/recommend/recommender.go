package recommend

import (
	"fmt"
	"sort"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/archetypes"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// Recommender generates deck recommendations combining archetype matches with custom variations
type Recommender struct {
	archetypeAnalyzer *archetypes.Analyzer
	scorer            *Scorer
	variationGenerator *VariationGenerator
	options           RecommenderOptions
}

// NewRecommender creates a new deck recommender
func NewRecommender(dataDir string, options RecommenderOptions) *Recommender {
	return &Recommender{
		archetypeAnalyzer: archetypes.NewAnalyzer(dataDir),
		scorer:            NewScorer(),
		variationGenerator: NewVariationGenerator(),
		options:           options,
	}
}

// GenerateRecommendations creates deck recommendations for a player
func (r *Recommender) GenerateRecommendations(
	playerTag string,
	playerName string,
	analysis deck.CardAnalysis,
) (*RecommendationResult, error) {
	// 1. Generate archetype decks for all 8 archetypes
	archetypeResult, err := r.archetypeAnalyzer.AnalyzeArchetypes(
		playerTag,
		playerName,
		analysis,
		r.options.TargetLevel,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze archetypes: %w", err)
	}

	// 2. Convert archetype decks to recommendations and score them
	recommendations := make([]*DeckRecommendation, 0)
	for _, archDeck := range archetypeResult.Archetypes {
		rec := r.scoreArchetypeDeck(archDeck, analysis)
		recommendations = append(recommendations, rec)
	}

	// 3. Generate custom variations for top archetypes
	if r.options.IncludeVariations {
		topArchetypes := r.getTopArchetypes(recommendations, 3)
		for _, archRec := range topArchetypes {
			variations := r.variationGenerator.GenerateVariations(
				archRec.Deck,
				archRec.Archetype,
				analysis,
				r.options.MaxVariationsPerArchetype,
			)

			// Score each variation
			for _, variation := range variations {
				r.scoreRecommendation(variation, analysis)
				recommendations = append(recommendations, variation)
			}
		}
	}

	// 4. Apply filters (arena/league)
	recommendations = r.applyFilters(recommendations)

	// 5. Sort by overall score (descending)
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].OverallScore > recommendations[j].OverallScore
	})

	// 6. Apply minimum compatibility filter
	filtered := make([]*DeckRecommendation, 0)
	for _, rec := range recommendations {
		if rec.CompatibilityScore >= r.options.MinCompatibility {
			filtered = append(filtered, rec)
		}
	}
	recommendations = filtered

	// 7. Apply limit
	if r.options.Limit > 0 && len(recommendations) > r.options.Limit {
		recommendations = recommendations[:r.options.Limit]
	}

	// 8. Determine top archetype
	topArchetype := ""
	if len(recommendations) > 0 {
		topArchetype = recommendations[0].ArchetypeName
	}

	// 9. Generate reasons for all recommendations
	for _, rec := range recommendations {
		if len(rec.Reasons) == 0 {
			rec.Reasons = r.scorer.GenerateReasons(rec)
		}
	}

	return &RecommendationResult{
		PlayerTag:       playerTag,
		PlayerName:      playerName,
		Recommendations: recommendations,
		TopArchetype:    topArchetype,
		ArenaFilter:     r.options.Arena,
		LeagueFilter:    r.options.League,
		GeneratedAt:     time.Now().Format(time.RFC3339),
	}, nil
}

// scoreArchetypeDeck converts an archetype deck to a recommendation with scores
func (r *Recommender) scoreArchetypeDeck(
	archDeck archetypes.ArchetypeDeck,
	analysis deck.CardAnalysis,
) *DeckRecommendation {
	// Create deck recommendation from archetype deck
	deckRec := &deck.DeckRecommendation{
		Deck:       archDeck.Deck,
		DeckDetail: archDeck.DeckDetail,
		AvgElixir:  archDeck.AvgElixir,
		Notes:      []string{},
	}

	rec := &DeckRecommendation{
		Deck:          deckRec,
		Archetype:     archDeck.Archetype,
		ArchetypeName: string(archDeck.Archetype),
		Type:          TypeArchetypeMatch,
		UpgradeCost: UpgradeCost{
			CardsNeeded:    archDeck.CardsNeeded,
			GoldNeeded:     archDeck.GoldNeeded,
			GemsNeeded:     archDeck.GemsNeeded,
			DistanceMetric: archDeck.DistanceMetric,
		},
	}

	r.scoreRecommendation(rec, analysis)

	return rec
}

// scoreRecommendation calculates all scores for a recommendation
func (r *Recommender) scoreRecommendation(
	rec *DeckRecommendation,
	analysis deck.CardAnalysis,
) {
	// Calculate scores directly using analysis.CardLevels (map[string]deck.CardLevelData)
	rec.CompatibilityScore = r.scorer.CalculateCompatibility(rec.Deck.DeckDetail, analysis.CardLevels)
	rec.SynergyScore = r.scorer.CalculateSynergy(rec.Deck.Deck)

	// Archetype fit is based on how well deck matches archetype constraints
	// For archetype matches, this is high. For variations, it's slightly lower.
	archetypeFit := 100.0
	if rec.Type == TypeCustomVariation {
		archetypeFit = 85.0 // Slightly lower for variations
	}

	rec.OverallScore = r.scorer.CalculateOverallScore(
		rec.CompatibilityScore,
		rec.SynergyScore,
		archetypeFit,
	)
}

// getTopArchetypes returns the top N recommendations by overall score
func (r *Recommender) getTopArchetypes(recommendations []*DeckRecommendation, n int) []*DeckRecommendation {
	// Sort by overall score (descending)
	sorted := make([]*DeckRecommendation, len(recommendations))
	copy(sorted, recommendations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].OverallScore > sorted[j].OverallScore
	})

	// Filter to archetype matches only (no variations)
	archetypeOnly := make([]*DeckRecommendation, 0)
	for _, rec := range sorted {
		if rec.Type == TypeArchetypeMatch {
			archetypeOnly = append(archetypeOnly, rec)
		}
	}

	// Limit to N
	if len(archetypeOnly) > n {
		archetypeOnly = archetypeOnly[:n]
	}

	return archetypeOnly
}

// applyFilters applies arena and league filters to recommendations
func (r *Recommender) applyFilters(recommendations []*DeckRecommendation) []*DeckRecommendation {
	// TODO: Implement arena/league filtering
	// For now, just return all recommendations
	// Future implementation could:
	// - Filter out cards not available at certain arenas
	// - Prioritize decks that work well in specific leagues

	return recommendations
}
