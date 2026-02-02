// Package deck provides defensive capability scoring for Clash Royale decks.
// This file implements counter coverage analysis to ensure decks can defend against common threats.
package deck

import (
	"fmt"
	"sort"
)

// DefensiveCoverageReport represents a comprehensive defensive analysis of a deck
type DefensiveCoverageReport struct {
	OverallScore       float64                  `json:"overall_score"`        // 0.0 to 1.0
	CategoryScores     map[CounterCategory]float64 `json:"category_scores"`     // Score per category
	ThreatAnalysis     []ThreatCoverage          `json:"threat_analysis"`      // Analysis of specific threats
	CoverageGaps       []string                  `json:"coverage_gaps"`        // Missing counter capabilities
	StrongCounters     []Counter                 `json:"strong_counters"`      // Deck's best counters
	RecommendedAdds    []string                  `json:"recommended_adds"`     // Cards to consider adding
}

// DefensiveScorer analyzes deck defensive capabilities
type DefensiveScorer struct {
	matrix *CounterMatrix
}

// NewDefensiveScorer creates a new defensive scorer with the given counter matrix
func NewDefensiveScorer(matrix *CounterMatrix) *DefensiveScorer {
	return &DefensiveScorer{
		matrix: matrix,
	}
}

// CalculateDefensiveCoverage calculates comprehensive defensive coverage for a deck
// This is the main entry point for defensive analysis
func (ds *DefensiveScorer) CalculateDefensiveCoverage(deckCards []string) *DefensiveCoverageReport {
	if len(deckCards) == 0 {
		return &DefensiveCoverageReport{
			OverallScore:   0.0,
			CategoryScores: make(map[CounterCategory]float64),
			CoverageGaps:   []string{"Empty deck"},
		}
	}

	report := &DefensiveCoverageReport{
		CategoryScores: make(map[CounterCategory]float64),
	}

	// Analyze each counter category
	airDefenseScore := ds.scoreCategory(deckCards, CounterAirDefense, 2, 3)
	report.CategoryScores[CounterAirDefense] = airDefenseScore

	tankKillerScore := ds.scoreCategory(deckCards, CounterTankKillers, 1, 2)
	report.CategoryScores[CounterTankKillers] = tankKillerScore

	splashScore := ds.scoreCategory(deckCards, CounterSplashDefense, 1, 2)
	report.CategoryScores[CounterSplashDefense] = splashScore

	swarmClearScore := ds.scoreCategory(deckCards, CounterSwarmClear, 1, 2)
	report.CategoryScores[CounterSwarmClear] = swarmClearScore

	buildingScore := ds.scoreCategory(deckCards, CounterBuildings, 0, 1)
	report.CategoryScores[CounterBuildings] = buildingScore

	// Calculate overall score (weighted average)
	report.OverallScore = ds.calculateOverallScore(report.CategoryScores)

	// Analyze specific common threats
	report.ThreatAnalysis = ds.analyzeCommonThreats(deckCards)

	// Identify coverage gaps
	report.CoverageGaps = ds.identifyGaps(deckCards, report.CategoryScores)

	// Identify strong counters
	report.StrongCounters = ds.identifyStrongCounters(deckCards)

	// Generate recommendations
	report.RecommendedAdds = ds.generateRecommendations(deckCards, report.CoverageGaps)

	return report
}

// scoreCategory scores a single counter category based on deck composition
func (ds *DefensiveScorer) scoreCategory(deckCards []string, category CounterCategory, minRequired, ideal int) float64 {
	count := ds.matrix.CountCardsWithCapability(deckCards, category)

	if count < minRequired {
		// Linear penalty from 0.6 down to 0
		return 0.6 * float64(count) / float64(minRequired)
	}
	if count >= ideal {
		return 1.0
	}
	// Linear between min and ideal
	return 0.6 + (0.4 * float64(count-minRequired) / float64(ideal-minRequired))
}

// calculateOverallScore computes the weighted overall defensive score
func (ds *DefensiveScorer) calculateOverallScore(scores map[CounterCategory]float64) float64 {
	// Weights based on importance of each category
	weights := map[CounterCategory]float64{
		CounterAirDefense:    0.25, // Most critical - can't lose to air
		CounterTankKillers:   0.20, // Need to stop beatdown
		CounterSplashDefense: 0.20, // Swarms are common
		CounterSwarmClear:    0.15, // Spell bait is prevalent
		CounterBuildings:     0.10, // Nice to have but not required
	}

	// Remaining 10% for general coverage
	baseScore := 0.10

	weightedSum := baseScore
	for category, score := range scores {
		if weight, exists := weights[category]; exists {
			weightedSum += score * weight
		}
	}

	return weightedSum
}

// analyzeCommonThreats analyzes how well the deck counters common meta threats
func (ds *DefensiveScorer) analyzeCommonThreats(deckCards []string) []ThreatCoverage {
	// List of high-priority threats to check
	priorityThreats := []string{
		"Mega Knight",
		"Balloon",
		"Graveyard",
		"Hog Rider",
		"Golem",
		"Lava Hound",
		"Elite Barbarians",
		"X-Bow",
		"Goblin Barrel",
		"Royal Giant",
	}

	var analyses []ThreatCoverage
	for _, threat := range priorityThreats {
		coverage := ds.matrix.AnalyzeThreatCoverage(deckCards, threat)
		analyses = append(analyses, coverage)
	}

	// Sort by effectiveness (worst first)
	sort.Slice(analyses, func(i, j int) bool {
		return analyses[i].Effectiveness < analyses[j].Effectiveness
	})

	return analyses
}

// identifyGaps identifies missing counter capabilities in the deck
func (ds *DefensiveScorer) identifyGaps(deckCards []string, scores map[CounterCategory]float64) []string {
	var gaps []string

	// Check each category for insufficient coverage
	if scores[CounterAirDefense] < 0.6 {
		gaps = append(gaps, "Insufficient air defense")
	}
	if scores[CounterTankKillers] < 0.6 {
		gaps = append(gaps, "No tank killer")
	}
	if scores[CounterSplashDefense] < 0.6 {
		gaps = append(gaps, "No splash damage")
	}
	if scores[CounterSwarmClear] < 0.6 {
		gaps = append(gaps, "No swarm spell")
	}
	if scores[CounterBuildings] < 0.4 {
		gaps = append(gaps, "No defensive building")
	}

	return gaps
}

// identifyStrongCounters identifies the best counters in the deck
func (ds *DefensiveScorer) identifyStrongCounters(deckCards []string) []Counter {
	var allCounters []Counter

	for _, card := range deckCards {
		// Check what this card counters across all threats
		// We'll look for high-effectiveness counters
		capabilities := ds.matrix.GetCardCapabilities(card)
		for _, cap := range capabilities {
			// Add as a counter with its capability
			allCounters = append(allCounters, Counter{
				Card:          card,
				Effectiveness: 0.8, // Base effectiveness for having the capability
				Reason:        fmt.Sprintf("Provides %s", cap),
			})
		}
	}

	// Sort by effectiveness and take top 5
	sort.Slice(allCounters, func(i, j int) bool {
		return allCounters[i].Effectiveness > allCounters[j].Effectiveness
	})

	if len(allCounters) > 5 {
		allCounters = allCounters[:5]
	}

	return allCounters
}

// buildDeckCardMap creates a lookup map for fast deck membership checking
func buildDeckCardMap(deckCards []string) map[string]bool {
	deckMap := make(map[string]bool)
	for _, card := range deckCards {
		deckMap[card] = true
	}
	return deckMap
}

// getRecommendedCardsForGap returns suggested cards for a specific coverage gap
func getRecommendedCardsForGap(gap string) []string {
	switch gap {
	case "Insufficient air defense":
		return []string{"Musketeer", "Electro Wizard", "Inferno Tower", "Baby Dragon"}
	case "No tank killer":
		return []string{"Inferno Tower", "P.E.K.K.A", "Mini P.E.K.K.A", "Inferno Dragon"}
	case "No splash damage":
		return []string{"Valkyrie", "Baby Dragon", "Wizard", "Executioner"}
	case "No swarm spell":
		return []string{"The Log", "Zap", "Arrows"}
	case "No defensive building":
		return []string{"Cannon", "Tesla", "Inferno Tower", "Bomb Tower"}
	default:
		return nil
	}
}

// findFirstAvailableCard returns the first card from candidates not already in the deck
func findFirstAvailableCard(candidates []string, deckMap map[string]bool) string {
	for _, card := range candidates {
		if !deckMap[card] {
			return card
		}
	}
	return ""
}

// generateRecommendations suggests cards to add based on coverage gaps
func (ds *DefensiveScorer) generateRecommendations(deckCards []string, gaps []string) []string {
	var recommendations []string
	deckMap := buildDeckCardMap(deckCards)

	for _, gap := range gaps {
		candidates := getRecommendedCardsForGap(gap)
		if card := findFirstAvailableCard(candidates, deckMap); card != "" {
			recommendations = append(recommendations, card)
		}
	}

	return recommendations
}

// GetScoreForCategory returns the defensive score for a specific category
func (ds *DefensiveScorer) GetScoreForCategory(deckCards []string, category CounterCategory) float64 {
	switch category {
	case CounterAirDefense:
		return ds.scoreCategory(deckCards, category, 2, 3)
	case CounterTankKillers:
		return ds.scoreCategory(deckCards, category, 1, 2)
	case CounterSplashDefense:
		return ds.scoreCategory(deckCards, category, 1, 2)
	case CounterSwarmClear:
		return ds.scoreCategory(deckCards, category, 1, 2)
	case CounterBuildings:
		return ds.scoreCategory(deckCards, category, 0, 1)
	default:
		return 0.0
	}
}
