// Package deck provides threat analysis for Clash Royale decks.
// This file implements threat detection and analysis to evaluate deck defensive capabilities.
package deck

import (
	"fmt"
	"sort"
)

// ThreatType represents the category of threat a card or archetype poses
type ThreatType string

const (
	ThreatTypeWinCondition ThreatType = "win_condition" // Primary tower damage
	ThreatTypeTank         ThreatType = "tank"          // High HP beatdown
	ThreatTypeAir          ThreatType = "air"           // Air threats
	ThreatTypeSwarm        ThreatType = "swarm"         // Swarm pushes
	ThreatTypeSiege        ThreatType = "siege"         // Building siege
	ThreatTypeSpell        ThreatType = "spell"         // Spell win conditions
	ThreatTypeRush         ThreatType = "rush"          // Fast charge attacks
	ThreatTypeControl      ThreatType = "control"       // Control/stall decks
)

// ThreatDefinition defines a threat that a deck may face
type ThreatDefinition struct {
	Name          string     `json:"name"`
	Type          ThreatType `json:"type"`
	Description   string     `json:"description"`
	MetaRelevance float64    `json:"meta_relevance"` // 0.0 to 1.0, how common in meta
}

// ThreatMatch represents how well a deck can counter a threat
type ThreatMatch struct {
	Threat        ThreatDefinition `json:"threat"`
	CanCounter    bool             `json:"can_counter"`
	Effectiveness float64          `json:"effectiveness"` // 0.0 to 1.0
	CounterCards  []string         `json:"counter_cards"` // Cards in deck that counter this threat
	Gaps          []string         `json:"gaps"`          // What the deck is missing
}

// ThreatAnalysisReport provides comprehensive threat analysis for a deck
type ThreatAnalysisReport struct {
	OverallDefensiveScore float64            `json:"overall_defensive_score"` // 0.0 to 1.0
	MetaThreatMatches     []ThreatMatch      `json:"meta_threat_matches"`     // Analysis of meta-relevant threats
	CriticalGaps          []string           `json:"critical_gaps"`           // Threats deck cannot counter
	StrongDefenses        []ThreatMatch      `json:"strong_defenses"`         // Threats deck counters well
	ThreatBreakdown       map[ThreatType]int `json:"threat_breakdown"`        // Count by threat type
}

// ThreatAnalyzer analyzes threats and deck capabilities against them
type ThreatAnalyzer struct {
	matrix      *CounterMatrix
	metaThreats []ThreatDefinition
}

// NewThreatAnalyzer creates a new threat analyzer with the given counter matrix
func NewThreatAnalyzer(matrix *CounterMatrix) *ThreatAnalyzer {
	analyzer := &ThreatAnalyzer{
		matrix: matrix,
	}

	// Initialize with default meta threats
	analyzer.metaThreats = analyzer.getDefaultMetaThreats()

	return analyzer
}

// getDefaultMetaThreats returns the default list of meta-relevant threats
func (ta *ThreatAnalyzer) getDefaultMetaThreats() []ThreatDefinition {
	return []ThreatDefinition{
		{Name: "Mega Knight", Type: ThreatTypeTank, Description: "High HP jumping tank", MetaRelevance: 0.95},
		{Name: "Balloon", Type: ThreatTypeAir, Description: "Air win condition with death damage", MetaRelevance: 0.85},
		{Name: "Graveyard", Type: ThreatTypeSpell, Description: "Persistent spell win condition", MetaRelevance: 0.80},
		{Name: "Hog Rider", Type: ThreatTypeWinCondition, Description: "Fast cycle win condition", MetaRelevance: 0.90},
		{Name: "Golem", Type: ThreatTypeTank, Description: "Slow beatdown tank", MetaRelevance: 0.75},
		{Name: "Lava Hound", Type: ThreatTypeAir, Description: "Air tank with pup death value", MetaRelevance: 0.80},
		{Name: "Elite Barbarians", Type: ThreatTypeRush, Description: "Fast rush attack", MetaRelevance: 0.70},
		{Name: "X-Bow", Type: ThreatTypeSiege, Description: "Long-range building siege", MetaRelevance: 0.65},
		{Name: "Goblin Barrel", Type: ThreatTypeSpell, Description: "Spell bait win condition", MetaRelevance: 0.75},
		{Name: "Royal Giant", Type: ThreatTypeSiege, Description: "Bridge siege win condition", MetaRelevance: 0.70},
		{Name: "Sparky", Type: ThreatTypeControl, Description: "High damage charge card", MetaRelevance: 0.60},
		{Name: "Prince", Type: ThreatTypeRush, Description: "Charge win condition", MetaRelevance: 0.65},
		{Name: "Three Musketeers", Type: ThreatTypeWinCondition, Description: "Split win condition", MetaRelevance: 0.70},
		{Name: "Miner", Type: ThreatTypeWinCondition, Description: "Flexible anywhere win condition", MetaRelevance: 0.80},
		{Name: "Giant Skeleton", Type: ThreatTypeTank, Description: "Death bomb tank", MetaRelevance: 0.55},
	}
}

// AnalyzeDeck performs comprehensive threat analysis on a deck
func (ta *ThreatAnalyzer) AnalyzeDeck(deckCards []string) *ThreatAnalysisReport {
	if len(deckCards) == 0 {
		return &ThreatAnalysisReport{
			OverallDefensiveScore: 0.0,
			CriticalGaps:          []string{"Empty deck"},
			ThreatBreakdown:       make(map[ThreatType]int),
		}
	}

	report := &ThreatAnalysisReport{
		MetaThreatMatches: make([]ThreatMatch, 0),
		ThreatBreakdown:   make(map[ThreatType]int),
	}

	// Analyze each meta threat
	for _, threat := range ta.metaThreats {
		match := ta.analyzeThreat(deckCards, threat)
		report.MetaThreatMatches = append(report.MetaThreatMatches, match)

		// Track threat types
		report.ThreatBreakdown[threat.Type]++
	}

	// Calculate overall defensive score
	report.OverallDefensiveScore = ta.calculateOverallScore(report.MetaThreatMatches)

	// Identify critical gaps (cannot counter meta-relevant threats)
	report.CriticalGaps = ta.identifyCriticalGaps(report.MetaThreatMatches)

	// Identify strong defenses
	report.StrongDefenses = ta.identifyStrongDefenses(report.MetaThreatMatches)

	return report
}

// analyzeThreat analyzes how well a deck can counter a specific threat
func (ta *ThreatAnalyzer) analyzeThreat(deckCards []string, threat ThreatDefinition) ThreatMatch {
	coverage := ta.matrix.AnalyzeThreatCoverage(deckCards, threat.Name)

	match := ThreatMatch{
		Threat:        threat,
		CanCounter:    coverage.CanCounter,
		Effectiveness: coverage.Effectiveness,
		CounterCards:  make([]string, 0),
		Gaps:          make([]string, 0),
	}

	// Extract counter card names
	for _, counter := range coverage.DeckCounters {
		match.CounterCards = append(match.CounterCards, counter.Card)
	}

	// Extract missing counter names
	for _, missing := range coverage.MissingCounters {
		match.Gaps = append(match.Gaps, missing.Card)
	}

	return match
}

// calculateOverallScore computes the weighted defensive score across all threats
func (ta *ThreatAnalyzer) calculateOverallScore(matches []ThreatMatch) float64 {
	if len(matches) == 0 {
		return 0.0
	}

	totalWeightedScore := 0.0
	totalWeight := 0.0

	for _, match := range matches {
		// Weight by meta relevance
		weight := match.Threat.MetaRelevance
		totalWeight += weight
		totalWeightedScore += match.Effectiveness * weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalWeightedScore / totalWeight
}

// identifyCriticalGaps identifies threats the deck cannot counter
func (ta *ThreatAnalyzer) identifyCriticalGaps(matches []ThreatMatch) []string {
	var gaps []string

	for _, match := range matches {
		// Only consider meta-relevant threats (relevance >= 0.7)
		if match.Threat.MetaRelevance >= 0.7 && !match.CanCounter {
			gaps = append(gaps, fmt.Sprintf("%s (%s)", match.Threat.Name, match.Threat.Type))
		}
	}

	// Also include low-effectiveness counters against high-meta threats
	for _, match := range matches {
		if match.Threat.MetaRelevance >= 0.85 && match.Effectiveness < 0.5 {
			gaps = append(gaps, fmt.Sprintf("%s (weak counter)", match.Threat.Name))
		}
	}

	return gaps
}

// identifyStrongDefenses identifies threats the deck counters very well
func (ta *ThreatAnalyzer) identifyStrongDefenses(matches []ThreatMatch) []ThreatMatch {
	var strong []ThreatMatch

	for _, match := range matches {
		// Strong defense: effectiveness >= 0.8 against meta-relevant threats
		if match.Threat.MetaRelevance >= 0.7 && match.Effectiveness >= 0.8 {
			strong = append(strong, match)
		}
	}

	// Sort by effectiveness (highest first)
	sort.Slice(strong, func(i, j int) bool {
		return strong[i].Effectiveness > strong[j].Effectiveness
	})

	return strong
}

// GetMatchForThreat returns the threat match for a specific threat name
func (ta *ThreatAnalyzer) GetMatchForThreat(deckCards []string, threatName string) *ThreatMatch {
	for _, threat := range ta.metaThreats {
		if threat.Name == threatName {
			match := ta.analyzeThreat(deckCards, threat)
			return &match
		}
	}
	return nil
}

// GetThreatsByType returns all threats of a specific type
func (ta *ThreatAnalyzer) GetThreatsByType(threatType ThreatType) []ThreatDefinition {
	var threats []ThreatDefinition
	for _, threat := range ta.metaThreats {
		if threat.Type == threatType {
			threats = append(threats, threat)
		}
	}
	return threats
}

// AddCustomThreat adds a custom threat to the analyzer
func (ta *ThreatAnalyzer) AddCustomThreat(threat ThreatDefinition) {
	ta.metaThreats = append(ta.metaThreats, threat)
}

// SetMetaThreats replaces the meta threats list
func (ta *ThreatAnalyzer) SetMetaThreats(threats []ThreatDefinition) {
	ta.metaThreats = threats
}

// GetMetaThreats returns the current list of meta threats
func (ta *ThreatAnalyzer) GetMetaThreats() []ThreatDefinition {
	return ta.metaThreats
}
