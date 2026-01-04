// Package evaluation provides comprehensive deck evaluation with meta-aware scoring
package evaluation

import (
	"fmt"
	"math"
	"sort"

	"github.com/klauer/clash-royale-api/go/pkg/events"
)

// MetaAnalyzer provides meta-aware deck evaluation using event tracking data
type MetaAnalyzer struct {
	eventData *events.EventAnalysis
	options   MetaAnalysisOptions
}

// MetaAnalysisOptions configures meta-aware analysis behavior
type MetaAnalysisOptions struct {
	EnableMetaScoring         bool    `json:"enable_meta_scoring"`         // Apply meta-based score adjustments
	MetaWeight                float64 `json:"meta_weight"`                 // Weight of meta score (0.0-1.0)
	TrendWindowDays           int     `json:"trend_window_days"`           // Days to consider for trend analysis
	MinSampleSize             int     `json:"min_sample_size"`             // Minimum battles for reliable stats
	ShowWeakMatchups          bool    `json:"show_weak_matchups"`          // Highlight weak meta matchups
	RecommendMetaAlternatives bool    `json:"recommend_meta_alternatives"` // Suggest meta-aware alternatives
}

// DefaultMetaAnalysisOptions returns sensible defaults for meta analysis
func DefaultMetaAnalysisOptions() MetaAnalysisOptions {
	return MetaAnalysisOptions{
		EnableMetaScoring:         true,
		MetaWeight:                0.15, // 15% weight for meta adjustments
		TrendWindowDays:           30,   // Last 30 days
		MinSampleSize:             10,   // Minimum 10 battles
		ShowWeakMatchups:          true,
		RecommendMetaAlternatives: true,
	}
}

// NewMetaAnalyzer creates a new meta-aware analyzer
func NewMetaAnalyzer(eventData *events.EventAnalysis, options MetaAnalysisOptions) *MetaAnalyzer {
	if options.MetaWeight == 0 {
		options = DefaultMetaAnalysisOptions()
	}

	return &MetaAnalyzer{
		eventData: eventData,
		options:   options,
	}
}

// MetaAdjustment represents a score adjustment based on meta data
type MetaAdjustment struct {
	BaseScore       float64  `json:"base_score"`        // Original score
	MetaScore       float64  `json:"meta_score"`        // Meta-aware score
	Adjustment      float64  `json:"adjustment"`        // Score delta (-100 to +100)
	Factors         []string `json:"factors"`           // Reasons for adjustment
	MetaTier        string   `json:"meta_tier"`         // S/A/B/C/D tier based on meta
	WinRateEstimate float64  `json:"win_rate_estimate"` // Estimated win rate
	Confidence      float64  `json:"confidence"`        // Confidence in meta data (0-1)
}

// CardMetaInfo provides meta information for a single card
type CardMetaInfo struct {
	CardName      string   `json:"card_name"`
	Popularity    float64  `json:"popularity"`     // 0-100, how often card is used
	WinRate       float64  `json:"win_rate"`       // Win rate when this card is in deck
	Trend         string   `json:"trend"`          // "rising", "stable", "falling"
	SampleSize    int      `json:"sample_size"`    // Number of battles for stats
	MetaTier      string   `json:"meta_tier"`      // S/A/B/C/D tier
	IsTrending    bool     `json:"is_trending"`    // Currently trending up
	WeakAgainst   []string `json:"weak_against"`   // Cards/decks this struggles against
	StrongAgainst []string `json:"strong_against"` // Cards/decks this counters
}

// DeckMetaAnalysis provides comprehensive meta analysis for a deck
type DeckMetaAnalysis struct {
	DeckCards           []string             `json:"deck_cards"`
	MetaAdjustment      MetaAdjustment       `json:"meta_adjustment"`
	CardMetaInfo        []CardMetaInfo       `json:"card_meta_info"`
	ArchetypeMatch      string               `json:"archetype_match"`      // Best matching archetype
	ArchetypeWinRate    float64              `json:"archetype_win_rate"`   // Win rate for this archetype
	TrendingCards       int                  `json:"trending_cards"`       // Number of trending cards in deck
	WeakMatchups        []MatchupAnalysis    `json:"weak_matchups"`        // Problematic matchups
	StrongMatchups      []MatchupAnalysis    `json:"strong_matchups"`      // Favorable matchups
	MetaRecommendations []MetaRecommendation `json:"meta_recommendations"` // Suggested improvements
}

// MatchupAnalysis analyzes how a deck performs against another deck/archetype
type MatchupAnalysis struct {
	OpponentDeck   string   `json:"opponent_deck"`   // Opponent archetype or deck name
	WinRate        float64  `json:"win_rate"`        // Expected win rate
	Confidence     float64  `json:"confidence"`      // Statistical confidence
	StrengthFactor float64  `json:"strength_factor"` // Positive = favorable, negative = unfavorable
	KeyCards       []string `json:"key_cards"`       // Cards that determine this matchup
	Notes          string   `json:"notes"`           // Additional context
}

// MetaRecommendation suggests deck improvements based on meta data
type MetaRecommendation struct {
	Type            string  `json:"type"`             // "replace", "add", "upgrade"
	CardName        string  `json:"card_name"`        // Card to modify
	Reason          string  `json:"reason"`           // Why this change is recommended
	ExpectedImpact  float64 `json:"expected_impact"`  // Expected score improvement
	AlternativeCard string  `json:"alternative_card"` // Suggested replacement (if applicable)
}

// AnalyzeDeckWithMeta performs meta-aware analysis of a deck
func (ma *MetaAnalyzer) AnalyzeDeckWithMeta(deckCards []string, baseScore float64) *DeckMetaAnalysis {
	if ma.eventData == nil {
		// Return empty analysis if no event data available
		return &DeckMetaAnalysis{
			DeckCards: deckCards,
			MetaAdjustment: MetaAdjustment{
				BaseScore:  baseScore,
				MetaScore:  baseScore,
				Adjustment: 0,
				Confidence: 0,
			},
			CardMetaInfo:        []CardMetaInfo{},
			MetaRecommendations: []MetaRecommendation{},
		}
	}

	// Analyze each card's meta performance
	cardMeta := ma.analyzeCardMeta(deckCards)

	// Calculate meta adjustment
	adjustment := ma.calculateMetaAdjustment(deckCards, cardMeta, baseScore)

	// Identify archetype match
	archetypeMatch := ma.identifyArchetype(deckCards)
	archetypeWinRate := ma.getArchetypeWinRate(archetypeMatch)

	// Count trending cards
	trendingCount := ma.countTrendingCards(cardMeta)

	// Analyze matchups (simplified - would need more data for full analysis)
	weakMatchups := ma.identifyWeakMatchups(deckCards, cardMeta)
	strongMatchups := ma.identifyStrongMatchups(deckCards, cardMeta)

	// Generate recommendations
	recommendations := ma.generateRecommendations(deckCards, cardMeta, adjustment)

	return &DeckMetaAnalysis{
		DeckCards:           deckCards,
		MetaAdjustment:      adjustment,
		CardMetaInfo:        cardMeta,
		ArchetypeMatch:      archetypeMatch,
		ArchetypeWinRate:    archetypeWinRate,
		TrendingCards:       trendingCount,
		WeakMatchups:        weakMatchups,
		StrongMatchups:      strongMatchups,
		MetaRecommendations: recommendations,
	}
}

// analyzeCardMeta extracts meta information for each card in the deck
func (ma *MetaAnalyzer) analyzeCardMeta(deckCards []string) []CardMetaInfo {
	cardMeta := make([]CardMetaInfo, 0, len(deckCards))

	// Build lookup maps from event data
	cardPopularity := ma.buildCardPopularityMap()
	cardWinRates := ma.buildCardWinRateMap()

	for _, cardName := range deckCards {
		popularity := cardPopularity[cardName]
		winRate := cardWinRates[cardName]

		// Determine meta tier based on win rate
		metaTier := calculateMetaTier(winRate)

		// Determine trend (simplified - would need historical data)
		trend := "stable"
		isTrending := false
		if winRate > 0.55 {
			trend = "rising"
			isTrending = true
		} else if winRate < 0.45 {
			trend = "falling"
		}

		metaInfo := CardMetaInfo{
			CardName:   cardName,
			Popularity: popularity,
			WinRate:    winRate,
			Trend:      trend,
			SampleSize: ma.estimateSampleSize(cardName),
			MetaTier:   metaTier,
			IsTrending: isTrending,
			// WeakAgainst/StrongAgainst would need matchup data
		}

		cardMeta = append(cardMeta, metaInfo)
	}

	return cardMeta
}

// calculateMetaTier converts win rate to tier (S/A/B/C/D)
func calculateMetaTier(winRate float64) string {
	if winRate >= 0.60 {
		return "S" // Overpowered
	} else if winRate >= 0.55 {
		return "A" // Strong
	} else if winRate >= 0.50 {
		return "B" // Balanced
	} else if winRate >= 0.45 {
		return "C" // Below average
	}
	return "D" // Weak
}

// buildCardPopularityMap creates a map of card -> popularity score (0-100)
func (ma *MetaAnalyzer) buildCardPopularityMap() map[string]float64 {
	popularity := make(map[string]float64)

	if ma.eventData == nil || len(ma.eventData.CardAnalysis.MostUsedCards) == 0 {
		return popularity
	}

	// Find max usage for normalization
	maxUsage := 0
	for _, card := range ma.eventData.CardAnalysis.MostUsedCards {
		if card.Count > maxUsage {
			maxUsage = card.Count
		}
	}

	// Normalize to 0-100 scale
	for _, card := range ma.eventData.CardAnalysis.MostUsedCards {
		if maxUsage > 0 {
			popularity[card.CardName] = float64(card.Count) / float64(maxUsage) * 100
		}
	}

	return popularity
}

// buildCardWinRateMap creates a map of card -> win rate
func (ma *MetaAnalyzer) buildCardWinRateMap() map[string]float64 {
	winRates := make(map[string]float64)

	if ma.eventData == nil {
		return winRates
	}

	// Combine card usage and win rate data
	for _, card := range ma.eventData.CardAnalysis.HighestWinRateCards {
		winRates[card.CardName] = card.WinRate
	}

	return winRates
}

// estimateSampleSize estimates the number of battles for a card's stats
func (ma *MetaAnalyzer) estimateSampleSize(cardName string) int {
	if ma.eventData == nil {
		return 0
	}

	// Simplified - use event summary as proxy
	return ma.eventData.Summary.TotalBattles
}

// calculateMetaAdjustment computes the meta-based score adjustment
func (ma *MetaAnalyzer) calculateMetaAdjustment(deckCards []string, cardMeta []CardMetaInfo, baseScore float64) MetaAdjustment {
	if !ma.options.EnableMetaScoring || len(cardMeta) == 0 {
		return MetaAdjustment{
			BaseScore:  baseScore,
			MetaScore:  baseScore,
			Adjustment: 0,
			Confidence: 0,
		}
	}

	// Calculate average card win rate
	totalWinRate := 0.0
	validCards := 0
	var factors []string

	for _, meta := range cardMeta {
		if meta.WinRate > 0 {
			totalWinRate += meta.WinRate
			validCards++
		}
	}

	avgWinRate := 0.50 // Default
	if validCards > 0 {
		avgWinRate = totalWinRate / float64(validCards)
	}

	// Calculate adjustment based on win rate
	// Win rate > 50% increases score, < 50% decreases
	winRateDelta := avgWinRate - 0.50
	adjustment := winRateDelta * 200 // Scale: 10% win rate difference = 20 point adjustment

	// Apply meta weight
	adjustment = adjustment * ma.options.MetaWeight

	// Add bonus for trending cards
	trendingBonus := 0.0
	for _, meta := range cardMeta {
		if meta.IsTrending {
			trendingBonus += 2.0
			factors = append(factors, fmt.Sprintf("Trending card: %s", meta.CardName))
		}
	}
	adjustment += trendingBonus

	// Determine meta tier
	metaTier := calculateMetaTier(avgWinRate)

	// Calculate confidence based on sample size
	confidence := ma.calculateConfidence(validCards)

	// Calculate final meta score
	metaScore := baseScore + adjustment

	// Clamp score to valid range
	if metaScore > 100 {
		metaScore = 100
	} else if metaScore < 0 {
		metaScore = 0
	}

	// Add factor descriptions
	if len(factors) == 0 {
		if adjustment > 0 {
			factors = append(factors, "Above average meta win rate")
		} else if adjustment < 0 {
			factors = append(factors, "Below average meta win rate")
		} else {
			factors = append(factors, "Neutral meta performance")
		}
	}

	return MetaAdjustment{
		BaseScore:       baseScore,
		MetaScore:       metaScore,
		Adjustment:      adjustment,
		Factors:         factors,
		MetaTier:        metaTier,
		WinRateEstimate: avgWinRate,
		Confidence:      confidence,
	}
}

// calculateConfidence determines how confident we are in the meta data (0-1)
func (ma *MetaAnalyzer) calculateConfidence(dataPoints int) float64 {
	if ma.eventData == nil {
		return 0
	}

	// More data = higher confidence
	totalBattles := ma.eventData.Summary.TotalBattles
	if totalBattles < ma.options.MinSampleSize {
		return 0.0
	}

	// Logarithmic scaling: more data gives diminishing confidence returns
	confidence := math.Min(1.0, math.Log10(float64(totalBattles))/math.Log10(100))
	return confidence
}

// identifyArchetype identifies the archetype that best matches this deck
func (ma *MetaAnalyzer) identifyArchetype(deckCards []string) string {
	// This is a simplified implementation
	// In production, would use the archetype package for proper matching

	// For now, return a generic archetype based on elixir and card types
	avgElixir := 0.0
	for range deckCards {
		// Simplified elixir estimation
		avgElixir += 3.5 // Average elixir
	}
	avgElixir /= float64(len(deckCards))

	if avgElixir < 3.0 {
		return "Cycle"
	} else if avgElixir < 4.0 {
		return "Midrange"
	}
	return "Control"
}

// getArchetypeWinRate returns the win rate for a given archetype
func (ma *MetaAnalyzer) getArchetypeWinRate(archetype string) float64 {
	// Simplified - would use event data to calculate actual archetype win rates
	if ma.eventData != nil {
		return ma.eventData.Summary.OverallWinRate
	}
	return 0.50
}

// countTrendingCards counts how many cards in the deck are trending
func (ma *MetaAnalyzer) countTrendingCards(cardMeta []CardMetaInfo) int {
	count := 0
	for _, meta := range cardMeta {
		if meta.IsTrending {
			count++
		}
	}
	return count
}

// identifyWeakMatchups identifies matchups where this deck struggles
func (ma *MetaAnalyzer) identifyWeakMatchups(deckCards []string, cardMeta []CardMetaInfo) []MatchupAnalysis {
	if !ma.options.ShowWeakMatchups {
		return nil
	}

	// Find cards with low win rates
	var weakMatchups []MatchupAnalysis

	for _, meta := range cardMeta {
		if meta.WinRate > 0 && meta.WinRate < 0.45 {
			// This card has a below-average win rate
			weakMatchups = append(weakMatchups, MatchupAnalysis{
				OpponentDeck:   fmt.Sprintf("Decks countering %s", meta.CardName),
				WinRate:        meta.WinRate,
				Confidence:     0.5,
				StrengthFactor: -(0.5 - meta.WinRate) * 100, // Negative = unfavorable
				KeyCards:       []string{meta.CardName},
				Notes:          fmt.Sprintf("%s has a below-average win rate (%.1f%%)", meta.CardName, meta.WinRate*100),
			})
		}
	}

	return weakMatchups
}

// identifyStrongMatchups identifies matchups where this deck excels
func (ma *MetaAnalyzer) identifyStrongMatchups(deckCards []string, cardMeta []CardMetaInfo) []MatchupAnalysis {
	// Find cards with high win rates
	var strongMatchups []MatchupAnalysis

	for _, meta := range cardMeta {
		if meta.WinRate > 0.55 {
			// This card has an above-average win rate
			strongMatchups = append(strongMatchups, MatchupAnalysis{
				OpponentDeck:   fmt.Sprintf("Decks weak to %s", meta.CardName),
				WinRate:        meta.WinRate,
				Confidence:     0.5,
				StrengthFactor: (meta.WinRate - 0.5) * 100, // Positive = favorable
				KeyCards:       []string{meta.CardName},
				Notes:          fmt.Sprintf("%s has an above-average win rate (%.1f%%)", meta.CardName, meta.WinRate*100),
			})
		}
	}

	return strongMatchups
}

// generateRecommendations suggests deck improvements based on meta data
func (ma *MetaAnalyzer) generateRecommendations(deckCards []string, cardMeta []CardMetaInfo, adjustment MetaAdjustment) []MetaRecommendation {
	if !ma.options.RecommendMetaAlternatives {
		return nil
	}

	var recommendations []MetaRecommendation

	// Suggest replacing low-win-rate cards
	for _, meta := range cardMeta {
		if meta.WinRate > 0 && meta.WinRate < 0.45 {
			// Find trending alternatives from event data
			alternative := ma.findTrendingAlternative(meta.CardName)

			rec := MetaRecommendation{
				Type:            "replace",
				CardName:        meta.CardName,
				Reason:          fmt.Sprintf("Low win rate (%.1f%%)", meta.WinRate*100),
				ExpectedImpact:  (0.50 - meta.WinRate) * 100, // Potential improvement
				AlternativeCard: alternative,
			}
			recommendations = append(recommendations, rec)
		}
	}

	// Sort by expected impact (highest first)
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].ExpectedImpact > recommendations[j].ExpectedImpact
	})

	// Limit to top 5 recommendations
	if len(recommendations) > 5 {
		recommendations = recommendations[:5]
	}

	return recommendations
}

// findTrendingAlternative finds a trending card that could replace the given card
func (ma *MetaAnalyzer) findTrendingAlternative(cardName string) string {
	if ma.eventData == nil || len(ma.eventData.CardAnalysis.HighestWinRateCards) == 0 {
		return "Unknown"
	}

	// Return the highest win rate card as a suggestion
	topCard := ma.eventData.CardAnalysis.HighestWinRateCards[0]
	if topCard.CardName != cardName {
		return topCard.CardName
	}

	// Return second best if first is the same card
	if len(ma.eventData.CardAnalysis.HighestWinRateCards) > 1 {
		return ma.eventData.CardAnalysis.HighestWinRateCards[1].CardName
	}

	return "Unknown"
}

// ApplyMetaAdjustment applies meta adjustment to a base score
func ApplyMetaAdjustment(baseScore float64, metaAdjustment MetaAdjustment) float64 {
	if !metaAdjustment.IsZero() {
		return metaAdjustment.MetaScore
	}
	return baseScore
}

// IsZero checks if MetaAdjustment is empty/uninitialized
func (ma MetaAdjustment) IsZero() bool {
	return ma.MetaScore == 0 && ma.Adjustment == 0 && ma.BaseScore == 0
}

// FormatMetaAnalysis formats a meta analysis for human-readable output
func FormatMetaAnalysis(analysis *DeckMetaAnalysis) string {
	if analysis == nil {
		return ""
	}

	output := fmt.Sprintf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	output += fmt.Sprintf("║                       META ANALYSIS                                 ║\n")
	output += fmt.Sprintf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	// Score adjustment
	adj := analysis.MetaAdjustment
	output += fmt.Sprintf("Score Adjustment:\n")
	output += fmt.Sprintf("──────────────────\n")
	output += fmt.Sprintf("Base Score:  %.2f\n", adj.BaseScore)
	output += fmt.Sprintf("Meta Score:  %.2f (%+.2f)\n", adj.MetaScore, adj.Adjustment)
	output += fmt.Sprintf("Meta Tier:   %s\n", adj.MetaTier)
	output += fmt.Sprintf("Confidence: %.0f%%\n\n", adj.Confidence*100)

	if len(adj.Factors) > 0 {
		output += fmt.Sprintf("Factors:\n")
		for _, factor := range adj.Factors {
			output += fmt.Sprintf("  • %s\n", factor)
		}
		output += "\n"
	}

	// Card meta information
	if len(analysis.CardMetaInfo) > 0 {
		output += fmt.Sprintf("Card Meta Data:\n")
		output += fmt.Sprintf("───────────────\n")
		for _, meta := range analysis.CardMetaInfo {
			trendIcon := "•"
			if meta.IsTrending {
				trendIcon = "📈"
			}
			output += fmt.Sprintf("  %s %s: Tier %s (%.1f%% WR, %.0f%% popularity)\n",
				trendIcon, meta.CardName, meta.MetaTier, meta.WinRate*100, meta.Popularity)
		}
		output += "\n"
	}

	// Archetype
	output += fmt.Sprintf("Archetype: %s (%.1f%% WR)\n", analysis.ArchetypeMatch, analysis.ArchetypeWinRate*100)
	output += fmt.Sprintf("Trending Cards: %d/8\n\n", analysis.TrendingCards)

	// Matchups
	if len(analysis.WeakMatchups) > 0 {
		output += fmt.Sprintf("Weak Matchups:\n")
		output += fmt.Sprintf("──────────────\n")
		for _, matchup := range analysis.WeakMatchups {
			output += fmt.Sprintf("  ⚠ %s: %.1f%% WR\n", matchup.OpponentDeck, matchup.WinRate*100)
		}
		output += "\n"
	}

	if len(analysis.StrongMatchups) > 0 {
		output += fmt.Sprintf("Strong Matchups:\n")
		output += fmt.Sprintf("────────────────\n")
		for _, matchup := range analysis.StrongMatchups {
			output += fmt.Sprintf("  ✓ %s: %.1f%% WR\n", matchup.OpponentDeck, matchup.WinRate*100)
		}
		output += "\n"
	}

	// Recommendations
	if len(analysis.MetaRecommendations) > 0 {
		output += fmt.Sprintf("Meta-Aware Recommendations:\n")
		output += fmt.Sprintf("─────────────────────────\n")
		for i, rec := range analysis.MetaRecommendations {
			output += fmt.Sprintf("%d. %s %s", i+1, rec.Type, rec.CardName)
			if rec.AlternativeCard != "" {
				output += fmt.Sprintf(" → %s", rec.AlternativeCard)
			}
			output += fmt.Sprintf("\n   Reason: %s (+%.1f expected)\n", rec.Reason, rec.ExpectedImpact)
		}
	}

	return output
}
