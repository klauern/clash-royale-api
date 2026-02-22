package recommend

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/archetypes"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// Recommender generates deck recommendations combining archetype matches with custom variations
type Recommender struct {
	archetypeAnalyzer  *archetypes.Analyzer
	scorer             *Scorer
	variationGenerator *VariationGenerator
	options            RecommenderOptions
}

// NewRecommender creates a new deck recommender
func NewRecommender(dataDir string, options RecommenderOptions) *Recommender {
	return &Recommender{
		archetypeAnalyzer:  archetypes.NewAnalyzer(dataDir),
		scorer:             NewScorer(),
		variationGenerator: NewVariationGenerator(),
		options:            options,
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
	profile, enabled := progressionFilterProfile(r.options.Arena, r.options.League)
	if !enabled {
		return recommendations
	}

	filtered := make([]*DeckRecommendation, 0, len(recommendations))
	for _, rec := range recommendations {
		if rec.CompatibilityScore < profile.minCompatibility {
			continue
		}
		if rec.UpgradeCost.DistanceMetric > profile.maxDistanceMetric {
			continue
		}
		if profile.maxCardsNeeded > 0 && rec.UpgradeCost.CardsNeeded > profile.maxCardsNeeded {
			continue
		}
		filtered = append(filtered, rec)
	}

	return filtered
}

type progressionFilterConfig struct {
	minCompatibility  float64
	maxDistanceMetric float64
	maxCardsNeeded    int
}

func progressionFilterProfile(arena, league string) (progressionFilterConfig, bool) {
	arena = strings.ToLower(strings.TrimSpace(arena))
	league = strings.ToLower(strings.TrimSpace(league))
	if arena == "" && league == "" {
		return progressionFilterConfig{}, false
	}

	// Conservative defaults when filters are requested but stage parsing is unknown.
	config := progressionFilterConfig{
		minCompatibility:  30.0,
		maxDistanceMetric: 0.65,
		maxCardsNeeded:    22000,
	}

	if level, ok := parseProgressionLevel(arena); ok {
		switch {
		case level <= 10:
			config.minCompatibility = 45.0
			config.maxDistanceMetric = 0.45
			config.maxCardsNeeded = 12000
		case level <= 14:
			config.minCompatibility = 35.0
			config.maxDistanceMetric = 0.55
			config.maxCardsNeeded = 18000
		}
	}

	switch {
	case strings.Contains(league, "challenger"):
		config.minCompatibility = maxFloat(config.minCompatibility, 40.0)
		config.maxDistanceMetric = minFloat(config.maxDistanceMetric, 0.55)
		config.maxCardsNeeded = minIntPositive(config.maxCardsNeeded, 18000)
	case strings.Contains(league, "master"):
		config.minCompatibility = maxFloat(config.minCompatibility, 35.0)
		config.maxDistanceMetric = minFloat(config.maxDistanceMetric, 0.60)
	case strings.Contains(league, "champion"), strings.Contains(league, "ultimate"):
		// High leagues can support broader recommendations.
		config.minCompatibility = minFloat(config.minCompatibility, 25.0)
		config.maxDistanceMetric = maxFloat(config.maxDistanceMetric, 0.75)
		config.maxCardsNeeded = 0
	}

	return config, true
}

func parseProgressionLevel(value string) (int, bool) {
	start := -1
	for i, r := range value {
		if r >= '0' && r <= '9' {
			start = i
			break
		}
	}
	if start == -1 {
		return 0, false
	}

	level := 0
	for _, r := range value[start:] {
		if r < '0' || r > '9' {
			break
		}
		level = (level * 10) + int(r-'0')
	}
	return level, true
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func minIntPositive(a, b int) int {
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}
