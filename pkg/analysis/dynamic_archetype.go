// Package analysis provides dynamic archetype detection based on player card collections.
// The dynamic archetype detector scores known archetype templates against a player's
// actual card levels and provides personalized upgrade recommendations.
package analysis

import (
	"fmt"
	"sort"
	"time"
)

const (
	viabilityOptimal     = "optimal"
	viabilityCompetitive = "competitive"
	viabilityPlayable    = "playable"
	viabilityBlocked     = "blocked"

	priorityCritical = "critical"
	priorityHigh     = "high"
	priorityMedium   = "medium"
	priorityLow      = "low"

	rarityCommon    = "Common"
	rarityRare      = "Rare"
	rarityEpic      = "Epic"
	rarityLegendary = "Legendary"
	rarityChampion  = "Champion"
)

// Strategy represents a deck building strategy (mirrors deck.Strategy to avoid circular dependency)
type Strategy string

// SynergyPair represents synergy between two cards (mirrors deck.SynergyPair)
type SynergyPair struct {
	Card1       string  `json:"card1"`
	Card2       string  `json:"card2"`
	SynergyType string  `json:"synergy_type"`
	Score       float64 `json:"score"` // 0.0 to 1.0
	Description string  `json:"description"`
}

// SynergyDB is an interface for querying card synergies
type SynergyDB interface {
	GetSynergy(card1, card2 string) float64
	GetSynergyPair(card1, card2 string) *SynergyPair
}

// StrategyConfig contains strategy configuration (simplified to avoid circular dependency)
type StrategyConfig struct {
	TargetElixirMin   float64
	TargetElixirMax   float64
	ArchetypeAffinity map[string]float64
}

// StrategyProvider provides strategy configurations
type StrategyProvider interface {
	GetConfig(strategy Strategy) StrategyConfig
	GetAllStrategies() []Strategy
}

// DetectedArchetype represents a dynamically scored archetype match based on player's collection
type DetectedArchetype struct {
	// Template information
	Name         string `json:"name"`
	WinCondition string `json:"win_condition"`

	// Collection-based scoring
	ViabilityScore float64 `json:"viability_score"` // 0-100
	ViabilityTier  string  `json:"viability_tier"`  // "optimal"/"competitive"/"playable"/"blocked"

	// Card analysis
	WinConditionLevel int      `json:"win_condition_level"`
	WinConditionMax   int      `json:"win_condition_max"`
	SupportCardsAvg   float64  `json:"support_cards_avg"`            // Average level of support cards
	MissingCards      []string `json:"missing_cards,omitempty"`      // Cards player doesn't own
	UnderleveledCards []string `json:"underleveled_cards,omitempty"` // Cards below 50% max level

	// Synergy analysis
	SynergyScore float64       `json:"synergy_score"`           // 0-100
	TopSynergies []SynergyPair `json:"top_synergies,omitempty"` // Top 3 synergies

	// Strategy recommendations (added in Phase 2)
	RecommendedStrategies []StrategyRecommendation `json:"recommended_strategies,omitempty"`

	// Upgrade path
	UpgradePriority   []ArchetypeUpgrade `json:"upgrade_priority,omitempty"`
	GoldToCompetitive int                `json:"gold_to_competitive"` // Gold needed to reach competitive tier

	// Metadata
	AvgElixir      float64 `json:"avg_elixir"`       // Average elixir cost
	OwnedCardCount int     `json:"owned_card_count"` // Number of archetype cards player owns
	TotalCardCount int     `json:"total_card_count"` // Total cards in archetype template
}

// StrategyRecommendation maps an archetype to a deck builder strategy
type StrategyRecommendation struct {
	Strategy           Strategy `json:"strategy"`            // "cycle", "aggro", "control", etc.
	CompatibilityScore float64  `json:"compatibility_score"` // 0-100
	Reason             string   `json:"reason"`              // Human-readable explanation
	ArchetypeAffinity  float64  `json:"archetype_affinity"`  // Affinity score from strategy config
}

// ArchetypeUpgrade suggests a specific card upgrade to improve archetype viability
type ArchetypeUpgrade struct {
	CardName          string  `json:"card_name"`
	CurrentLevel      int     `json:"current_level"`
	TargetLevel       int     `json:"target_level"`
	GoldCost          int     `json:"gold_cost"`
	ImpactOnViability float64 `json:"impact_on_viability"` // Score improvement
	Priority          string  `json:"priority"`            // "critical"/"high"/"medium"/"low"
}

// CardArchetypeImpact shows how upgrading a card affects multiple archetypes
type CardArchetypeImpact struct {
	CardName           string   `json:"card_name"`
	CurrentLevel       int      `json:"current_level"`
	GoldCost           int      `json:"gold_cost"`
	AffectedArchetypes []string `json:"affected_archetypes"`  // Archetype names affected
	TotalViabilityGain float64  `json:"total_viability_gain"` // Sum of viability gains
	ArchetypesUnlocked int      `json:"archetypes_unlocked"`  // Count crossing into competitive tier
}

// DynamicArchetypeAnalysis represents the complete dynamic archetype detection result
type DynamicArchetypeAnalysis struct {
	PlayerTag    string    `json:"player_tag"`
	PlayerName   string    `json:"player_name,omitempty"`
	AnalysisTime time.Time `json:"analysis_time"`

	// All detected archetypes (sorted by viability score descending)
	DetectedArchetypes []DetectedArchetype `json:"detected_archetypes"`

	// Archetypes grouped by viability tier
	OptimalArchetypes     []string `json:"optimal_archetypes"`     // 90-100
	CompetitiveArchetypes []string `json:"competitive_archetypes"` // 75-89
	PlayableArchetypes    []string `json:"playable_archetypes"`    // 60-74
	BlockedArchetypes     []string `json:"blocked_archetypes"`     // 0-59

	// Cross-archetype upgrade recommendations
	TopUpgradeImpacts []CardArchetypeImpact `json:"top_upgrade_impacts,omitempty"`
}

// DetectionOptions configures dynamic archetype detection behavior
type DetectionOptions struct {
	MinViability         float64  `json:"min_viability"`           // Minimum viability score to include (0-100)
	IncludeStrategies    bool     `json:"include_strategies"`      // Include strategy recommendations
	IncludeUpgrades      bool     `json:"include_upgrades"`        // Include upgrade recommendations
	TopUpgradesPerArch   int      `json:"top_upgrades_per_arch"`   // Number of top upgrades per archetype
	TopCrossArchUpgrades int      `json:"top_cross_arch_upgrades"` // Number of cross-archetype upgrades
	ExcludeCards         []string `json:"exclude_cards"`           // Cards to exclude from recommendations
}

// DefaultDetectionOptions returns sensible defaults for archetype detection
func DefaultDetectionOptions() DetectionOptions {
	return DetectionOptions{
		MinViability:         0, // Show all archetypes
		IncludeStrategies:    true,
		IncludeUpgrades:      true,
		TopUpgradesPerArch:   3,
		TopCrossArchUpgrades: 10,
		ExcludeCards:         []string{},
	}
}

// DynamicArchetypeDetector performs dynamic archetype detection
type DynamicArchetypeDetector struct {
	archetypes         []DeckArchetypeTemplate
	synergyDB          SynergyDB
	strategyProvider   StrategyProvider
	dataDir            string
	archetypesFilePath string // Path to custom archetypes file
}

// NewDynamicArchetypeDetector creates a new dynamic archetype detector.
// If archetypesFilePath is empty, uses embedded default archetypes.
// Otherwise loads archetypes from the specified JSON file.
// synergyDB and strategyProvider must be provided by the caller to avoid circular dependencies.
func NewDynamicArchetypeDetector(
	dataDir string,
	archetypesFilePath string,
	synergyDB SynergyDB,
	strategyProvider StrategyProvider,
) (*DynamicArchetypeDetector, error) {
	// Load archetype templates
	archetypes, err := LoadArchetypes(archetypesFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load archetypes: %w", err)
	}

	return &DynamicArchetypeDetector{
		archetypes:         archetypes,
		synergyDB:          synergyDB,
		strategyProvider:   strategyProvider,
		dataDir:            dataDir,
		archetypesFilePath: archetypesFilePath,
	}, nil
}

// DetectArchetypes performs dynamic archetype detection on a player's card collection
func (d *DynamicArchetypeDetector) DetectArchetypes(
	cardAnalysis *CardAnalysis,
	options DetectionOptions,
) (*DynamicArchetypeAnalysis, error) {
	if cardAnalysis == nil {
		return nil, ErrMissingPlayerTag
	}

	// Score each archetype template
	detectedArchetypes := make([]DetectedArchetype, 0, len(d.archetypes))

	for _, template := range d.archetypes {
		detected := d.scoreArchetype(template, cardAnalysis)

		// Apply minimum viability filter
		if detected.ViabilityScore >= options.MinViability {
			detectedArchetypes = append(detectedArchetypes, detected)
		}
	}

	// Sort by viability score descending
	sort.Slice(detectedArchetypes, func(i, j int) bool {
		return detectedArchetypes[i].ViabilityScore > detectedArchetypes[j].ViabilityScore
	})

	// Group by tier
	optimal := []string{}
	competitive := []string{}
	playable := []string{}
	blocked := []string{}

	for _, arch := range detectedArchetypes {
		switch arch.ViabilityTier {
		case viabilityOptimal:
			optimal = append(optimal, arch.Name)
		case viabilityCompetitive:
			competitive = append(competitive, arch.Name)
		case viabilityPlayable:
			playable = append(playable, arch.Name)
		case viabilityBlocked:
			blocked = append(blocked, arch.Name)
		}
	}

	analysis := &DynamicArchetypeAnalysis{
		PlayerTag:             cardAnalysis.PlayerTag,
		PlayerName:            cardAnalysis.PlayerName,
		AnalysisTime:          time.Now(),
		DetectedArchetypes:    detectedArchetypes,
		OptimalArchetypes:     optimal,
		CompetitiveArchetypes: competitive,
		PlayableArchetypes:    playable,
		BlockedArchetypes:     blocked,
	}

	// Add strategy recommendations if requested
	if options.IncludeStrategies {
		d.AddStrategyRecommendations(analysis)
	}

	// Add upgrade recommendations if requested
	if options.IncludeUpgrades {
		d.addUpgradeRecommendations(analysis, cardAnalysis, options)
	}

	return analysis, nil
}

// scoreArchetype calculates viability score for a single archetype template
func (d *DynamicArchetypeDetector) scoreArchetype(
	template DeckArchetypeTemplate,
	analysis *CardAnalysis,
) DetectedArchetype {
	// Calculate viability score components
	winConScore := d.calculateWinConditionScore(template, analysis)
	supportScore := d.calculateSupportScore(template, analysis)
	synergyScore := d.calculateSynergyScore(template, analysis)
	completenessScore := d.calculateCompletenessScore(template, analysis)

	// Weighted viability score (0-100)
	viabilityScore := (winConScore * 0.35) +
		(supportScore * 0.30) +
		(synergyScore * 0.20) +
		(completenessScore * 0.15)

	// Determine tier
	tier := getViabilityTier(viabilityScore)

	// Get win condition info
	winConInfo, exists := analysis.CardLevels[template.WinCondition]
	winConLevel := 0
	winConMax := 14 // Default max
	if exists {
		winConLevel = winConInfo.Level
		winConMax = winConInfo.MaxLevel
	}

	// Identify missing and underleveled cards
	missingCards, underleveledCards := d.identifyProblemCards(template, analysis)

	// Get top synergies
	topSynergies := d.getTopSynergies(template, analysis, 3)

	// Calculate average elixir and card counts
	avgElixir := d.calculateAvgElixir(template, analysis)
	ownedCount, totalCount := d.getCardCounts(template, analysis)

	return DetectedArchetype{
		Name:              template.Name,
		WinCondition:      template.WinCondition,
		ViabilityScore:    viabilityScore,
		ViabilityTier:     tier,
		WinConditionLevel: winConLevel,
		WinConditionMax:   winConMax,
		SupportCardsAvg:   supportScore / 100.0 * float64(winConMax), // Convert back to level
		MissingCards:      missingCards,
		UnderleveledCards: underleveledCards,
		SynergyScore:      synergyScore,
		TopSynergies:      topSynergies,
		AvgElixir:         avgElixir,
		OwnedCardCount:    ownedCount,
		TotalCardCount:    totalCount,
		GoldToCompetitive: 0, // Calculate in upgrade recommendations
	}
}

// calculateWinConditionScore returns the win condition level ratio × 100 (0-100)
func (d *DynamicArchetypeDetector) calculateWinConditionScore(
	template DeckArchetypeTemplate,
	analysis *CardAnalysis,
) float64 {
	winConInfo, exists := analysis.CardLevels[template.WinCondition]
	if !exists {
		return 0 // Cannot build archetype without win condition
	}

	// Use CardLevelInfo.LevelRatio() for consistency with deck builder
	return winConInfo.LevelRatio() * 100
}

// calculateSupportScore returns the average support card level ratio × 100 (0-100)
func (d *DynamicArchetypeDetector) calculateSupportScore(
	template DeckArchetypeTemplate,
	analysis *CardAnalysis,
) float64 {
	if len(template.SupportCards) == 0 {
		return 100 // No support cards required
	}

	totalRatio := 0.0
	count := 0

	// Average level ratio of OWNED support cards only
	for _, cardName := range template.SupportCards {
		if cardInfo, exists := analysis.CardLevels[cardName]; exists {
			totalRatio += cardInfo.LevelRatio()
			count++
		}
	}

	if count == 0 {
		return 0 // Player owns none of the support cards
	}

	return (totalRatio / float64(count)) * 100
}

// calculateSynergyScore analyzes card synergies using the synergy database (0-100)
func (d *DynamicArchetypeDetector) calculateSynergyScore(
	template DeckArchetypeTemplate,
	analysis *CardAnalysis,
) float64 {
	totalSynergy := 0.0
	pairCount := 0

	// Check synergies between win condition and support cards
	for _, support := range template.SupportCards {
		if _, exists := analysis.CardLevels[support]; exists {
			synergy := d.synergyDB.GetSynergy(template.WinCondition, support)
			if synergy > 0 {
				totalSynergy += synergy
				pairCount++
			}
		}
	}

	// Check synergies among support cards
	ownedSupport := []string{}
	for _, support := range template.SupportCards {
		if _, exists := analysis.CardLevels[support]; exists {
			ownedSupport = append(ownedSupport, support)
		}
	}

	for i := 0; i < len(ownedSupport); i++ {
		for j := i + 1; j < len(ownedSupport); j++ {
			synergy := d.synergyDB.GetSynergy(ownedSupport[i], ownedSupport[j])
			if synergy > 0 {
				totalSynergy += synergy
				pairCount++
			}
		}
	}

	if pairCount == 0 {
		// No known synergies - neutral score
		return 50
	}

	// Average synergy converted to 0-100 scale
	avgSynergy := totalSynergy / float64(pairCount)
	return avgSynergy * 100
}

// calculateCompletenessScore returns the percentage of archetype cards the player owns (0-100)
func (d *DynamicArchetypeDetector) calculateCompletenessScore(
	template DeckArchetypeTemplate,
	analysis *CardAnalysis,
) float64 {
	totalCards := 1 + len(template.SupportCards) // Win condition + support
	if len(template.RequiredCards) > 0 {
		totalCards += len(template.RequiredCards)
	}

	ownedCards := 0
	if _, exists := analysis.CardLevels[template.WinCondition]; exists {
		ownedCards++
	}

	for _, card := range template.SupportCards {
		if _, exists := analysis.CardLevels[card]; exists {
			ownedCards++
		}
	}

	for _, card := range template.RequiredCards {
		if _, exists := analysis.CardLevels[card]; exists {
			ownedCards++
		}
	}

	return (float64(ownedCards) / float64(totalCards)) * 100
}

// getViabilityTier returns the tier name for a viability score
func getViabilityTier(score float64) string {
	switch {
	case score >= 90:
		return "optimal"
	case score >= 75:
		return "competitive"
	case score >= 60:
		return "playable"
	default:
		return "blocked"
	}
}

// identifyProblemCards finds missing and underleveled cards in an archetype
func (d *DynamicArchetypeDetector) identifyProblemCards(
	template DeckArchetypeTemplate,
	analysis *CardAnalysis,
) (missing, underleveled []string) {
	allCards := append([]string{template.WinCondition}, template.SupportCards...)
	allCards = append(allCards, template.RequiredCards...)

	for _, cardName := range allCards {
		cardInfo, exists := analysis.CardLevels[cardName]
		if !exists {
			missing = append(missing, cardName)
		} else if cardInfo.LevelRatio() < 0.5 { // Below 50% max level
			underleveled = append(underleveled, cardName)
		}
	}

	return missing, underleveled
}

// getTopSynergies returns the top N synergies in an archetype
func (d *DynamicArchetypeDetector) getTopSynergies(
	template DeckArchetypeTemplate,
	analysis *CardAnalysis,
	topN int,
) []SynergyPair {
	synergies := []SynergyPair{}

	// Collect all synergies between owned cards
	allCards := []string{template.WinCondition}
	for _, card := range template.SupportCards {
		if _, exists := analysis.CardLevels[card]; exists {
			allCards = append(allCards, card)
		}
	}

	for i := 0; i < len(allCards); i++ {
		for j := i + 1; j < len(allCards); j++ {
			pair := d.synergyDB.GetSynergyPair(allCards[i], allCards[j])
			if pair != nil {
				synergies = append(synergies, *pair)
			}
		}
	}

	// Sort by score descending
	sort.Slice(synergies, func(i, j int) bool {
		return synergies[i].Score > synergies[j].Score
	})

	// Return top N
	if len(synergies) > topN {
		return synergies[:topN]
	}
	return synergies
}

// calculateAvgElixir calculates the average elixir cost for owned archetype cards
func (d *DynamicArchetypeDetector) calculateAvgElixir(
	template DeckArchetypeTemplate,
	analysis *CardAnalysis,
) float64 {
	totalElixir := 0
	count := 0

	allCards := append([]string{template.WinCondition}, template.SupportCards...)
	for _, cardName := range allCards {
		if cardInfo, exists := analysis.CardLevels[cardName]; exists {
			totalElixir += cardInfo.Elixir
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return float64(totalElixir) / float64(count)
}

// getCardCounts returns the number of owned cards vs total cards in the archetype
func (d *DynamicArchetypeDetector) getCardCounts(
	template DeckArchetypeTemplate,
	analysis *CardAnalysis,
) (owned, total int) {
	allCards := append([]string{template.WinCondition}, template.SupportCards...)
	allCards = append(allCards, template.RequiredCards...)

	total = len(allCards)
	owned = 0

	for _, cardName := range allCards {
		if _, exists := analysis.CardLevels[cardName]; exists {
			owned++
		}
	}

	return owned, total
}

// addUpgradeRecommendations calculates upgrade recommendations for improving archetype viability
func (d *DynamicArchetypeDetector) addUpgradeRecommendations(
	analysis *DynamicArchetypeAnalysis,
	cardAnalysis *CardAnalysis,
	options DetectionOptions,
) {
	// Calculate cross-archetype upgrade impacts
	cardImpacts := d.calculateCrossArchetypeImpacts(analysis, cardAnalysis)

	// Sort by total viability gain descending
	sort.Slice(cardImpacts, func(i, j int) bool {
		return cardImpacts[i].TotalViabilityGain > cardImpacts[j].TotalViabilityGain
	})

	// Take top N
	if len(cardImpacts) > options.TopCrossArchUpgrades {
		cardImpacts = cardImpacts[:options.TopCrossArchUpgrades]
	}

	analysis.TopUpgradeImpacts = cardImpacts

	// Add per-archetype upgrade recommendations
	for i := range analysis.DetectedArchetypes {
		arch := &analysis.DetectedArchetypes[i]
		arch.UpgradePriority = d.calculateArchetypeUpgrades(arch, cardAnalysis, options.TopUpgradesPerArch)
		arch.GoldToCompetitive = d.calculateGoldToCompetitive(arch)
	}
}

// calculateCrossArchetypeImpacts calculates how upgrading each card affects multiple archetypes
func (d *DynamicArchetypeDetector) calculateCrossArchetypeImpacts(
	analysis *DynamicArchetypeAnalysis,
	cardAnalysis *CardAnalysis,
) []CardArchetypeImpact {
	// Map of card name -> impact
	impactMap := make(map[string]*CardArchetypeImpact)

	// For each archetype, check which cards appear in it
	for _, arch := range analysis.DetectedArchetypes {
		// Find the archetype template
		var template *DeckArchetypeTemplate
		for i := range d.archetypes {
			if d.archetypes[i].Name == arch.Name {
				template = &d.archetypes[i]
				break
			}
		}
		if template == nil {
			continue
		}

		// Check all cards in the archetype
		allCards := append([]string{template.WinCondition}, template.SupportCards...)
		for _, cardName := range allCards {
			cardInfo, exists := cardAnalysis.CardLevels[cardName]
			if !exists || cardInfo.IsMaxLevel {
				continue
			}

			// Initialize impact if not exists
			if impactMap[cardName] == nil {
				impactMap[cardName] = &CardArchetypeImpact{
					CardName:           cardName,
					CurrentLevel:       cardInfo.Level,
					GoldCost:           estimateGoldCost(cardInfo), // TODO: implement proper gold cost
					AffectedArchetypes: []string{},
					TotalViabilityGain: 0,
					ArchetypesUnlocked: 0,
				}
			}

			// Add this archetype to affected list
			impact := impactMap[cardName]
			impact.AffectedArchetypes = append(impact.AffectedArchetypes, arch.Name)

			// Estimate viability gain (simplified - could be more sophisticated)
			viabilityGain := estimateViabilityGain(cardInfo, arch.ViabilityScore)
			impact.TotalViabilityGain += viabilityGain

			// Check if upgrade would unlock archetype (cross into competitive tier)
			if arch.ViabilityTier == "playable" && arch.ViabilityScore+viabilityGain >= 75 {
				impact.ArchetypesUnlocked++
			} else if arch.ViabilityTier == "blocked" && arch.ViabilityScore+viabilityGain >= 60 {
				impact.ArchetypesUnlocked++
			}
		}
	}

	// Convert map to slice
	impacts := make([]CardArchetypeImpact, 0, len(impactMap))
	for _, impact := range impactMap {
		impacts = append(impacts, *impact)
	}

	return impacts
}

// calculateArchetypeUpgrades returns top upgrade recommendations for a specific archetype
func (d *DynamicArchetypeDetector) calculateArchetypeUpgrades(
	arch *DetectedArchetype,
	cardAnalysis *CardAnalysis,
	topN int,
) []ArchetypeUpgrade {
	upgrades := []ArchetypeUpgrade{}

	// Find the template
	var template *DeckArchetypeTemplate
	for i := range d.archetypes {
		if d.archetypes[i].Name == arch.Name {
			template = &d.archetypes[i]
			break
		}
	}
	if template == nil {
		return upgrades
	}

	// Evaluate each card upgrade
	allCards := append([]string{template.WinCondition}, template.SupportCards...)
	for _, cardName := range allCards {
		cardInfo, exists := cardAnalysis.CardLevels[cardName]
		if !exists || cardInfo.IsMaxLevel {
			continue
		}

		impact := estimateViabilityGain(cardInfo, arch.ViabilityScore)
		priority := getUpgradePriority(impact, cardName == template.WinCondition)

		upgrades = append(upgrades, ArchetypeUpgrade{
			CardName:          cardName,
			CurrentLevel:      cardInfo.Level,
			TargetLevel:       cardInfo.Level + 1,
			GoldCost:          estimateGoldCost(cardInfo),
			ImpactOnViability: impact,
			Priority:          priority,
		})
	}

	// Sort by impact descending
	sort.Slice(upgrades, func(i, j int) bool {
		return upgrades[i].ImpactOnViability > upgrades[j].ImpactOnViability
	})

	if len(upgrades) > topN {
		return upgrades[:topN]
	}
	return upgrades
}

// calculateGoldToCompetitive estimates gold needed to reach competitive tier
func (d *DynamicArchetypeDetector) calculateGoldToCompetitive(arch *DetectedArchetype) int {
	if arch.ViabilityTier == viabilityCompetitive || arch.ViabilityTier == viabilityOptimal {
		return 0 // Already competitive
	}

	// Simplified estimate based on upgrade priority
	totalGold := 0
	for _, upgrade := range arch.UpgradePriority {
		if upgrade.Priority == priorityCritical || upgrade.Priority == priorityHigh {
			totalGold += upgrade.GoldCost
		}
	}

	return totalGold
}

// Helper functions

// estimateGoldCost estimates the gold cost for a +1 upgrade (simplified)
func estimateGoldCost(cardInfo CardLevelInfo) int {
	// Simplified gold cost estimation
	// TODO: Use actual upgrade costs from configuration
	switch cardInfo.Rarity {
	case rarityCommon:
		return 2000 + (cardInfo.Level * 200)
	case rarityRare:
		return 4000 + (cardInfo.Level * 400)
	case rarityEpic:
		return 8000 + (cardInfo.Level * 800)
	case rarityLegendary, rarityChampion:
		return 20000 + (cardInfo.Level * 2000)
	default:
		return 5000
	}
}

// estimateViabilityGain estimates how much a +1 upgrade improves viability
func estimateViabilityGain(cardInfo CardLevelInfo, currentViability float64) float64 {
	// Higher impact if card is far from max and archetype is close to tier boundary
	levelGap := float64(cardInfo.MaxLevel - cardInfo.Level)
	maxImpact := 10.0 // Max ~10 point gain per upgrade

	// Scale impact based on how far from max level
	impact := maxImpact * (levelGap / float64(cardInfo.MaxLevel))

	// Boost if near tier boundary (59->60, 74->75, 89->90)
	if currentViability >= 58 && currentViability < 60 {
		impact *= 1.5 // Crossing into playable
	} else if currentViability >= 73 && currentViability < 75 {
		impact *= 1.5 // Crossing into competitive
	} else if currentViability >= 88 && currentViability < 90 {
		impact *= 1.3 // Crossing into optimal
	}

	return impact
}

// getUpgradePriority determines upgrade priority level
func getUpgradePriority(impact float64, isWinCondition bool) string {
	// Win conditions always get priority boost
	if isWinCondition {
		if impact >= 8 {
			return priorityCritical
		}
		if impact >= 5 {
			return priorityHigh
		}
		return priorityMedium
	}

	// Support cards
	if impact >= 10 {
		return priorityCritical
	}
	if impact >= 7 {
		return priorityHigh
	}
	if impact >= 4 {
		return priorityMedium
	}
	return priorityLow
}
