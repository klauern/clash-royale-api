// Package deck provides the improved V2 scoring algorithm for deck evaluation.
//
// This file implements the redesigned scoring system that addresses weaknesses
// in the original algorithm:
//   - Reduced card level dominance (60% vs 120% weight)
//   - Added synergy awareness (20% weight)
//   - Added counter coverage analysis (15% weight)
//   - Added archetype coherence checking (10% weight)
//   - Increased elixir fit importance (25% vs 15% weight)
//
// The V2 algorithm is designed to be used alongside the original scorer during
// a migration period, with the goal of replacing V1 once validated.
//
// Design document: docs/IMPROVED_SCORING_DESIGN.md
// Related tasks: clash-royale-api-bwq8 (design), clash-royale-api-33f5 (research)
package deck

import (
	"math"

	"github.com/klauer/clash-royale-api/go/internal/config"
)

// ScorerV2Weights contains the weight constants for the V2 scoring algorithm.
// These weights are designed to balance card quality with strategic factors.
// All weights should sum to 1.0 for normalization.
const (
	// ScorerV2CardQualityWeight is the weight for individual card quality (reduced from 120%)
	// This reduces level dominance while still valuing card strength
	ScorerV2CardQualityWeight = 0.60

	// ScorerV2SynergyWeight is the weight for card pair synergies
	// Rewards decks with cards that work well together
	ScorerV2SynergyWeight = 0.20

	// ScorerV2CounterWeight is the weight for counter coverage analysis
	// Ensures decks can defend against common threats
	ScorerV2CounterWeight = 0.15

	// ScorerV2ArchetypeWeight is the weight for archetype coherence
	// Penalizes conflicting strategies (e.g., Golem + Hog Rider)
	ScorerV2ArchetypeWeight = 0.10

	// ScorerV2ElixirWeight is the weight for elixir curve fit
	// Increased to better enforce strategy-appropriate curves
	ScorerV2ElixirWeight = 0.25

	// ScorerV2CombatWeight is the weight for combat statistics
	// Factors in actual DPS, HP, range effectiveness
	ScorerV2CombatWeight = 0.20

	// ScorerV2UniquenessWeight is the weight for card uniqueness (anti-meta)
	// Rewards decks with less commonly used cards
	// Disabled by default (0.0), can be enabled via config
	ScorerV2UniquenessWeight = 0.0
)

// CounterCoverageThresholds defines minimum requirements for defensive coverage.
// These thresholds implement the WASTED framework for deck viability.
const (
	// MinAirDefense is the minimum number of air-targeting cards required
	MinAirDefense = 2

	// IdealAirDefense is the optimal number of air-targeting cards
	IdealAirDefense = 3

	// MinTankKillers is the minimum number of high-DPS or % damage cards
	MinTankKillers = 1

	// MinSplash is the minimum number of splash damage sources
	MinSplash = 1

	// MinSwarmSpells is the minimum number of swarm-clearing spells
	MinSwarmSpells = 1

	// MinBuildings is the recommended number of defensive buildings
	MinBuildings = 0

	// IdealBuildings is the optimal number of defensive buildings
	IdealBuildings = 1
)

// AntiSynergyPenalties defines score penalties for conflicting card combinations.
// These penalties are applied to archetype coherence scoring.
const (
	// AntiSynergyConflictingWinConditions penalizes decks with multiple win conditions
	// that have different play patterns (e.g., Golem + Hog Rider)
	AntiSynergyConflictingWinConditions = -0.30

	// AntiSynergyConflictingArchetypes penalizes mixing incompatible archetypes
	// (e.g., Siege + Beatdown)
	AntiSynergyConflictingArchetypes = -0.30

	// AntiSynergyTooManyBuildings penalizes over-reliance on buildings
	AntiSynergyTooManyBuildings = -0.20

	// AntiSynergyTooManySpells penalizes spell-heavy decks with few troops
	AntiSynergyTooManySpells = -0.20

	// AntiSynergyNoWinCondition penalizes decks without a clear win condition
	AntiSynergyNoWinCondition = -0.40
)

// MultiCardSynergyBonuses rewards emergent properties of 3+ card combinations.
const (
	// MultiCardArchetypeCore rewards having 3+ cards from the same archetype
	MultiCardArchetypeCore = 0.10

	// MultiCardWinConditionPackage rewards complete win condition support
	MultiCardWinConditionPackage = 0.15

	// MultiCardDefensiveCore rewards having building + support + spell
	MultiCardDefensiveCore = 0.10
)

// StrategyElixirProfile defines the target elixir curve for each strategy.
// Used to calculate elixir fit scores.
type StrategyElixirProfile struct {
	Target float64 // Optimal average elixir
	Min    float64 // Minimum acceptable average
	Max    float64 // Maximum acceptable average
}

// StrategyElixirProfiles maps strategies to their optimal elixir ranges.
// These profiles guide the elixir fit scoring component.
var StrategyElixirProfiles = map[Strategy]StrategyElixirProfile{
	StrategyCycle:    {Target: 2.8, Min: 2.4, Max: 3.2},
	StrategyControl:  {Target: 3.5, Min: 3.0, Max: 4.0},
	StrategyAggro:    {Target: 3.8, Min: 3.2, Max: 4.3},
	StrategyBalanced: {Target: 3.3, Min: 2.8, Max: 3.8},
	StrategySplash:   {Target: 3.5, Min: 3.0, Max: 4.0},
	StrategySpell:    {Target: 3.5, Min: 3.0, Max: 4.0},
	StrategySynergy:  {Target: 3.4, Min: 2.8, Max: 4.0},
}

// ScorerV2Result contains the detailed breakdown of a V2 deck score.
// This allows analysis of which components contribute to the final score.
type ScorerV2Result struct {
	// FinalScore is the weighted sum of all components (0.0-1.0 scale)
	FinalScore float64

	// Component scores (each 0.0-1.0, before weighting)
	CardQualityScore     float64
	SynergyScore         float64
	CounterCoverageScore float64
	ArchetypeScore       float64
	ElixirFitScore       float64
	CombatStatsScore     float64

	// Weighted contributions (component * weight)
	WeightedCardQuality     float64
	WeightedSynergy         float64
	WeightedCounterCoverage float64
	WeightedArchetype       float64
	WeightedElixirFit       float64
	WeightedCombatStats     float64
	WeightedUniqueness      float64

	// Detailed breakdown for analysis
	Details ScorerV2Details

	// Uniqueness scoring details (optional, only populated when enabled)
	UniquenessDetails *UniquenessResult
}

// ScorerV2Details provides granular information about the scoring.
type ScorerV2Details struct {
	// Synergy details
	DetectedSynergies    int
	MaxPossibleSynergies int
	AvgSynergyStrength   float64

	// Counter coverage details
	AirDefenseCount int
	TankKillerCount int
	SplashCount     int
	SwarmSpellCount int
	BuildingCount   int
	CoverageGaps    []string

	// Archetype details
	PrimaryArchetype    string
	ArchetypeConfidence float64
	AntiSynergies       []string
	MultiCardBonuses    []string

	// Elixir details
	AverageElixir  float64
	ElixirVariance float64
	CurveQuality   float64
}

// ScoreDeckV2 calculates the comprehensive V2 score for a deck.
// This is the main entry point for the improved scoring algorithm.
//
// Parameters:
//   - cards: The 8-card deck to evaluate
//   - strategy: The intended strategy (affects elixir fit and archetype scoring)
//   - synergyDB: The synergy database for pair analysis
//
// Returns a detailed result with component breakdowns.
func ScoreDeckV2(cards []CardCandidate, strategy Strategy, synergyDB *SynergyDatabase) ScorerV2Result {
	return ScoreDeckV2WithUniqueness(cards, strategy, synergyDB, nil)
}

// ScoreDeckV2WithUniqueness calculates the V2 score with optional uniqueness scoring.
//
// Parameters:
//   - cards: The 8-card deck to evaluate
//   - strategy: The intended strategy (affects elixir fit and archetype scoring)
//   - synergyDB: The synergy database for pair analysis
//   - uniquenessScorer: Optional uniqueness scorer (nil to disable)
//
// Returns a detailed result with component breakdowns including uniqueness.
func ScoreDeckV2WithUniqueness(cards []CardCandidate, strategy Strategy, synergyDB *SynergyDatabase, uniquenessScorer *UniquenessScorer) ScorerV2Result {
	if len(cards) == 0 {
		return ScorerV2Result{}
	}

	result := ScorerV2Result{}

	// Calculate individual component scores
	result.CardQualityScore = calculateCardQualityScore(cards)
	result.SynergyScore = calculateSynergyScore(cards, synergyDB, &result.Details)
	result.CounterCoverageScore = calculateCounterCoverageScore(cards, &result.Details)
	result.ArchetypeScore = calculateArchetypeScore(cards, strategy, &result.Details)
	result.ElixirFitScore = calculateElixirFitScore(cards, strategy, &result.Details)
	result.CombatStatsScore = calculateCombatStatsScore(cards)

	// Calculate weighted contributions
	result.WeightedCardQuality = result.CardQualityScore * ScorerV2CardQualityWeight
	result.WeightedSynergy = result.SynergyScore * ScorerV2SynergyWeight
	result.WeightedCounterCoverage = result.CounterCoverageScore * ScorerV2CounterWeight
	result.WeightedArchetype = result.ArchetypeScore * ScorerV2ArchetypeWeight
	result.WeightedElixirFit = result.ElixirFitScore * ScorerV2ElixirWeight
	result.WeightedCombatStats = result.CombatStatsScore * ScorerV2CombatWeight

	// Calculate uniqueness score if scorer provided and enabled
	if uniquenessScorer != nil && uniquenessScorer.config.Enabled {
		cardNames := make([]string, len(cards))
		for i, card := range cards {
			cardNames[i] = card.Name
		}
		uniquenessResult := uniquenessScorer.ScoreDeckWithDetails(cardNames)
		result.WeightedUniqueness = uniquenessResult.WeightedScore
		result.UniquenessDetails = &uniquenessResult
	}

	// Calculate final score
	result.FinalScore = result.WeightedCardQuality +
		result.WeightedSynergy +
		result.WeightedCounterCoverage +
		result.WeightedArchetype +
		result.WeightedElixirFit +
		result.WeightedCombatStats +
		result.WeightedUniqueness

	// Normalize to 0.0-1.0 range
	if result.FinalScore > 1.0 {
		result.FinalScore = 1.0
	}
	if result.FinalScore < 0.0 {
		result.FinalScore = 0.0
	}

	return result
}

// calculateCardQualityScore computes the average card quality.
// Uses reduced level weight (60% vs original 120%) to prevent level dominance.
func calculateCardQualityScore(cards []CardCandidate) float64 {
	if len(cards) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, card := range cards {
		// Use the V2 card scoring with reduced level weight
		cardScore := scoreCardV2(&card)
		totalScore += cardScore
	}

	// Average across all cards
	return totalScore / float64(len(cards))
}

// scoreCardV2 calculates an individual card score with reduced level emphasis.
// Formula: (levelRatio * 0.6 * rarityBoost) + (evolution * 0.5) + roleBonus
func scoreCardV2(card *CardCandidate) float64 {
	// Level component (reduced weight)
	levelRatio := card.LevelRatio()
	rarityBoost := config.GetRarityWeight(card.Rarity)
	levelComponent := levelRatio * 0.6 * rarityBoost

	// Evolution component
	evolutionComponent := 0.0
	if card.MaxEvolutionLevel > 0 && card.EvolutionLevel > 0 {
		evoRatio := float64(card.EvolutionLevel) / float64(card.MaxEvolutionLevel)
		evolutionComponent = evoRatio * 0.5
	}

	// Role bonus
	roleComponent := 0.0
	if card.HasRole() {
		roleComponent = config.RoleBonusValue
	}

	return levelComponent + evolutionComponent + roleComponent
}

// calculateSynergyScore computes synergy from card pairs and multi-card patterns.
// Uses the 188-pair synergy database plus emergent 3+ card bonuses.
func calculateSynergyScore(cards []CardCandidate, synergyDB *SynergyDatabase, details *ScorerV2Details) float64 {
	if len(cards) < 2 || synergyDB == nil {
		return 0.0
	}

	// Count possible pairs in 8-card deck: C(8,2) = 28
	maxPairs := len(cards) * (len(cards) - 1) / 2
	details.MaxPossibleSynergies = maxPairs

	// Find all synergies between card pairs
	detectedPairs := 0
	totalSynergyStrength := 0.0

	for i := 0; i < len(cards); i++ {
		for j := i + 1; j < len(cards); j++ {
			synergyScore := synergyDB.GetSynergy(cards[i].Name, cards[j].Name)
			if synergyScore > 0 {
				detectedPairs++
				totalSynergyStrength += synergyScore
			}
		}
	}

	details.DetectedSynergies = detectedPairs

	// Calculate average synergy strength
	avgSynergy := 0.0
	if detectedPairs > 0 {
		avgSynergy = totalSynergyStrength / float64(detectedPairs)
	}
	details.AvgSynergyStrength = avgSynergy

	// Coverage: what % of possible pairs have synergies
	coverage := float64(detectedPairs) / float64(maxPairs)

	// Base synergy score: 70% strength, 30% coverage
	synergyScore := (avgSynergy * 0.7) + (coverage * 0.3)

	// Apply multi-card synergy bonuses
	multiCardBonus := calculateMultiCardBonuses(cards, details)

	return synergyScore + multiCardBonus
}

// calculateMultiCardBonuses detects emergent synergies from 3+ card combinations.
func calculateMultiCardBonuses(cards []CardCandidate, details *ScorerV2Details) float64 {
	bonus := 0.0

	roleCounts := countCardsByRole(cards)

	bonus += calculateArchetypeCoreBonus(roleCounts, details)
	bonus += calculateWinConditionPackageBonus(roleCounts, details)
	bonus += calculateDefensiveCoreBonus(roleCounts, details)

	return bonus
}

// roleCounts holds the count of cards by role
type roleCounts struct {
	winCondition int
	building     int
	spell        int
	support      int
	all          map[CardRole]int
}

// countCardsByRole counts cards by their roles
func countCardsByRole(cards []CardCandidate) roleCounts {
	counts := roleCounts{
		all: make(map[CardRole]int),
	}

	for _, card := range cards {
		if card.Role == nil {
			continue
		}

		counts.all[*card.Role]++

		switch *card.Role {
		case RoleWinCondition:
			counts.winCondition++
		case RoleBuilding:
			counts.building++
		case RoleSpellBig, RoleSpellSmall:
			counts.spell++
		case RoleSupport:
			counts.support++
		}
	}

	return counts
}

// calculateArchetypeCoreBonus calculates bonus for having 3+ cards from same strategic group
func calculateArchetypeCoreBonus(counts roleCounts, details *ScorerV2Details) float64 {
	bonus := 0.0

	for role, count := range counts.all {
		if count >= 3 {
			bonus += MultiCardArchetypeCore
			details.MultiCardBonuses = append(details.MultiCardBonuses,
				"Archetype core: 3+ "+string(role)+" cards")
		}
	}

	return bonus
}

// calculateWinConditionPackageBonus calculates bonus for win condition + support + spell combination
func calculateWinConditionPackageBonus(counts roleCounts, details *ScorerV2Details) float64 {
	if counts.winCondition >= 1 && counts.support >= 2 && counts.spell >= 1 {
		details.MultiCardBonuses = append(details.MultiCardBonuses,
			"Complete win condition package")
		return MultiCardWinConditionPackage
	}
	return 0.0
}

// calculateDefensiveCoreBonus calculates bonus for building + support + spell combination
func calculateDefensiveCoreBonus(counts roleCounts, details *ScorerV2Details) float64 {
	if counts.building >= 1 && counts.support >= 1 && counts.spell >= 1 {
		details.MultiCardBonuses = append(details.MultiCardBonuses,
			"Defensive core established")
		return MultiCardDefensiveCore
	}
	return 0.0
}

// calculateCounterCoverageScore evaluates defensive capabilities against threats.
// Implements the WASTED framework: Win condition, Air, Splash, Tank killer, Elixir, Defense.
//
//nolint:funlen,gocognit,gocyclo // Counter coverage scoring uses explicit per-threat branching.
func calculateCounterCoverageScore(cards []CardCandidate, details *ScorerV2Details) float64 {
	// Count defensive capabilities
	airDefense := 0
	tankKillers := 0
	splash := 0
	swarmSpells := 0
	buildings := 0

	for _, card := range cards {
		// Air defense: cards that target air
		if card.Stats != nil && (card.Stats.Targets == "Air" || card.Stats.Targets == "Air & Ground") {
			airDefense++
		}

		// Tank killers: high DPS or % damage
		if card.Stats != nil && card.Stats.DamagePerSecond >= 200 {
			tankKillers++
		}
		// Include Inferno cards (percentage damage)
		if card.Name == "Inferno Dragon" || card.Name == "Inferno Tower" {
			tankKillers++
		}

		// Splash damage
		if card.Stats != nil && card.Stats.Radius > 0 {
			splash++
		}
		// Known splash cards
		if card.Name == "Valkyrie" || card.Name == "Baby Dragon" || card.Name == "Dark Prince" {
			splash++
		}

		// Swarm-clearing spells
		if card.Role != nil && *card.Role == RoleSpellSmall {
			swarmSpells++
		}

		// Buildings
		if card.Role != nil && *card.Role == RoleBuilding {
			buildings++
		}
	}

	// Store counts in details
	details.AirDefenseCount = airDefense
	details.TankKillerCount = tankKillers
	details.SplashCount = splash
	details.SwarmSpellCount = swarmSpells
	details.BuildingCount = buildings

	// Calculate coverage scores for each category
	airScore := calculateCoverageScore(airDefense, MinAirDefense, IdealAirDefense)
	tankScore := calculateCoverageScore(tankKillers, MinTankKillers, MinTankKillers+1)
	splashScore := calculateCoverageScore(splash, MinSplash, MinSplash+1)
	swarmScore := calculateCoverageScore(swarmSpells, MinSwarmSpells, MinSwarmSpells+1)
	buildingScore := calculateCoverageScore(buildings, MinBuildings, IdealBuildings)

	// Track gaps
	if airDefense < MinAirDefense {
		details.CoverageGaps = append(details.CoverageGaps, "Insufficient air defense")
	}
	if tankKillers < MinTankKillers {
		details.CoverageGaps = append(details.CoverageGaps, "No tank killer")
	}
	if splash < MinSplash {
		details.CoverageGaps = append(details.CoverageGaps, "No splash damage")
	}
	if swarmSpells < MinSwarmSpells {
		details.CoverageGaps = append(details.CoverageGaps, "No swarm spell")
	}

	// Weighted sum (weights from design document)
	coverageScore := (airScore * 0.25) +
		(tankScore * 0.20) +
		(splashScore * 0.20) +
		(swarmScore * 0.15) +
		(buildingScore * 0.10) +
		(0.10) // Big spell baseline (assumed present in most decks)

	return coverageScore
}

// calculateCoverageScore computes a 0-1 score based on count vs thresholds.
func calculateCoverageScore(count, minRequired, ideal int) float64 {
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

// calculateArchetypeScore evaluates strategic coherence and detects anti-synergies.
//
//nolint:funlen,gocognit,gocyclo // Archetype scoring keeps explicit strategy rules for transparency.
func calculateArchetypeScore(cards []CardCandidate, strategy Strategy, details *ScorerV2Details) float64 {
	// Count cards by role for archetype analysis
	winConditions := 0
	buildings := 0
	spells := 0

	// Track specific cards for anti-synergy detection
	hasGolem := false
	hasHogRider := false
	hasXBow := false
	hasMortar := false
	hasGiant := false
	hasLavaHound := false

	for _, card := range cards {
		if card.Role != nil {
			switch *card.Role {
			case RoleWinCondition:
				winConditions++
			case RoleBuilding:
				buildings++
			case RoleSpellBig, RoleSpellSmall:
				spells++
			}
		}

		// Check for specific cards
		//nolint:goconst // Card names are domain-specific values
		switch card.Name {
		case "Golem":
			hasGolem = true
		case "Hog Rider":
			hasHogRider = true
		case "X-Bow":
			hasXBow = true
		case "Mortar":
			hasMortar = true
		case "Giant":
			hasGiant = true
		case "Lava Hound":
			hasLavaHound = true
		}
	}

	// Base archetype match score
	archetypeScore := 0.8 // Start with good score, apply penalties

	// Detect anti-synergies
	penalties := 0.0

	// Conflicting win conditions
	if (hasGolem || hasLavaHound) && hasHogRider {
		penalties += -AntiSynergyConflictingWinConditions // Add positive value to penalties
		details.AntiSynergies = append(details.AntiSynergies,
			"Beatdown (Golem/Lava) conflicts with Cycle (Hog)")
	}

	if (hasXBow || hasMortar) && (hasGolem || hasGiant || hasLavaHound) {
		penalties += -AntiSynergyConflictingArchetypes
		details.AntiSynergies = append(details.AntiSynergies,
			"Siege (X-Bow/Mortar) conflicts with Beatdown")
	}

	// Too many buildings
	if buildings > 2 {
		penalties += -AntiSynergyTooManyBuildings
		details.AntiSynergies = append(details.AntiSynergies,
			"Too many buildings (>2)")
	}

	// Too many spells
	if spells > 4 {
		penalties += -AntiSynergyTooManySpells
		details.AntiSynergies = append(details.AntiSynergies,
			"Too many spells (>4)")
	}

	// No win condition
	if winConditions == 0 {
		penalties += -AntiSynergyNoWinCondition
		details.AntiSynergies = append(details.AntiSynergies,
			"No win condition")
	}

	// Apply penalties
	archetypeScore -= penalties

	// Strategy-specific bonuses
	if strategy == StrategyCycle && hasHogRider {
		archetypeScore += 0.05
	}
	if strategy == StrategyAggro && (hasGolem || hasGiant || hasLavaHound) {
		archetypeScore += 0.05
	}

	// Clamp to valid range
	if archetypeScore > 1.0 {
		archetypeScore = 1.0
	}
	if archetypeScore < 0.0 {
		archetypeScore = 0.0
	}

	return archetypeScore
}

// calculateElixirFitScore evaluates how well the elixir curve matches the strategy.
func calculateElixirFitScore(cards []CardCandidate, strategy Strategy, details *ScorerV2Details) float64 {
	if len(cards) == 0 {
		return 0.0
	}

	// Calculate average elixir
	totalElixir := 0
	elixirCounts := make(map[int]int) // Track distribution

	for _, card := range cards {
		totalElixir += card.Elixir
		elixirCounts[card.Elixir]++
	}

	avgElixir := float64(totalElixir) / float64(len(cards))
	details.AverageElixir = avgElixir

	// Calculate variance
	variance := 0.0
	for _, card := range cards {
		diff := float64(card.Elixir) - avgElixir
		variance += diff * diff
	}
	variance /= float64(len(cards))
	details.ElixirVariance = variance

	// Get strategy profile
	profile, exists := StrategyElixirProfiles[strategy]
	if !exists {
		profile = StrategyElixirProfiles[StrategyBalanced]
	}

	// Calculate strategy match
	strategyMatch := 1.0
	if avgElixir < profile.Min {
		strategyMatch = 1.0 - ((profile.Min - avgElixir) / 1.5)
	} else if avgElixir > profile.Max {
		strategyMatch = 1.0 - ((avgElixir - profile.Max) / 1.5)
	}
	if strategyMatch < 0.0 {
		strategyMatch = 0.0
	}

	// Calculate curve quality based on distribution
	curveQuality := calculateCurveQuality(elixirCounts, len(cards))
	details.CurveQuality = curveQuality

	// Combine: 60% strategy match, 40% curve quality
	return (strategyMatch * 0.6) + (curveQuality * 0.4)
}

// calculateCurveQuality evaluates the distribution of elixir costs.
func calculateCurveQuality(elixirCounts map[int]int, totalCards int) float64 {
	if totalCards == 0 {
		return 0.0
	}

	// Ideal distribution for most decks:
	// 1-2 elixir: 2-3 cards (cheap cycle/response)
	// 3-4 elixir: 3-4 cards (core cards)
	// 5+ elixir: 1-2 cards (heavy hitters)

	cheap := elixirCounts[1] + elixirCounts[2]
	medium := elixirCounts[3] + elixirCounts[4]
	heavy := countHeavyCards(elixirCounts)

	cheapScore := scoreCheapCards(cheap)
	mediumScore := scoreMediumCards(medium)
	heavyScore := scoreHeavyCards(heavy)

	// Weighted average
	return (cheapScore * 0.3) + (mediumScore * 0.5) + (heavyScore * 0.2)
}

// countHeavyCards counts cards with elixir cost >= 5
func countHeavyCards(elixirCounts map[int]int) int {
	heavy := 0
	for cost, count := range elixirCounts {
		if cost >= 5 {
			heavy += count
		}
	}
	return heavy
}

// scoreCheapCards scores the cheap card bucket (1-2 elixir, ideal: 2-3 cards)
func scoreCheapCards(count int) float64 {
	if count >= 2 && count <= 3 {
		return 1.0
	}
	if count < 2 {
		return float64(count) / 2.0
	}
	return 1.0 - (float64(count-3) * 0.2)
}

// scoreMediumCards scores the medium card bucket (3-4 elixir, ideal: 3-4 cards)
func scoreMediumCards(count int) float64 {
	if count >= 3 && count <= 4 {
		return 1.0
	}
	if count < 3 {
		return float64(count) / 3.0
	}
	return 1.0 - (float64(count-4) * 0.2)
}

// scoreHeavyCards scores the heavy card bucket (5+ elixir, ideal: 1-2 cards)
func scoreHeavyCards(count int) float64 {
	if count >= 1 && count <= 2 {
		return 1.0
	}
	if count == 0 {
		return 0.7 // No heavy hitter is okay but not ideal
	}
	return 1.0 - (float64(count-2) * 0.3)
}

// calculateCombatStatsScore evaluates actual card effectiveness.
func calculateCombatStatsScore(cards []CardCandidate) float64 {
	if len(cards) == 0 {
		return 0.0
	}

	totalScore := 0.0
	cardsWithStats := 0

	for _, card := range cards {
		if card.Stats == nil {
			continue
		}
		cardsWithStats++

		// DPS efficiency (capped at 50 DPS per elixir)
		dpsPerElixir := float64(card.Stats.DamagePerSecond) / float64(max(card.Elixir, 1))
		dpsScore := math.Min(dpsPerElixir/50.0, 1.0)

		// HP efficiency (capped at 400 HP per elixir)
		hpPerElixir := float64(card.Stats.Hitpoints) / float64(max(card.Elixir, 1))
		hpScore := math.Min(hpPerElixir/400.0, 1.0)

		// Target coverage
		targetScore := 0.5
		switch card.Stats.Targets {
		case "Air & Ground":
			targetScore = 1.0
		case "Air", "Ground":
			targetScore = 0.7
		}

		// Range effectiveness (normalized)
		rangeScore := math.Min(card.Stats.Range/6.0, 1.0)

		// Combine with weights
		cardScore := (dpsScore * 0.35) + (hpScore * 0.35) + (targetScore * 0.15) + (rangeScore * 0.15)
		totalScore += cardScore
	}

	if cardsWithStats == 0 {
		return 0.5 // Neutral score if no stats available
	}

	return totalScore / float64(cardsWithStats)
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ScoreDeckV2Simple provides a simplified interface for V2 scoring.
// Returns just the final score for quick comparisons.
func ScoreDeckV2Simple(cards []CardCandidate, strategy Strategy, synergyDB *SynergyDatabase) float64 {
	result := ScoreDeckV2(cards, strategy, synergyDB)
	return result.FinalScore
}

// CompareScorers runs both V1 and V2 scoring and returns comparison metrics.
// Useful for validation and migration testing.
func CompareScorers(cards []CardCandidate, strategy Strategy, synergyDB *SynergyDatabase) map[string]interface{} {
	// Calculate V1 score (average of individual card scores)
	v1Score := 0.0
	for _, card := range cards {
		v1Score += card.Score
	}
	if len(cards) > 0 {
		v1Score /= float64(len(cards))
	}

	// Calculate V2 score
	v2Result := ScoreDeckV2(cards, strategy, synergyDB)

	return map[string]interface{}{
		"v1_score":              v1Score,
		"v2_score":              v2Result.FinalScore,
		"difference":            v2Result.FinalScore - v1Score,
		"percent_change":        ((v2Result.FinalScore - v1Score) / v1Score) * 100,
		"v2_card_quality":       v2Result.CardQualityScore,
		"v2_synergy":            v2Result.SynergyScore,
		"v2_counter_coverage":   v2Result.CounterCoverageScore,
		"v2_archetype":          v2Result.ArchetypeScore,
		"v2_elixir_fit":         v2Result.ElixirFitScore,
		"v2_combat_stats":       v2Result.CombatStatsScore,
		"v2_coverage_gaps":      v2Result.Details.CoverageGaps,
		"v2_anti_synergies":     v2Result.Details.AntiSynergies,
		"v2_multi_card_bonuses": v2Result.Details.MultiCardBonuses,
	}
}

// ScoreDeckV2WithCounterAnalysis provides enhanced counter coverage analysis using the CounterMatrix.
// This function extends the basic V2 scoring with detailed threat-based counter analysis.
//
// Parameters:
//   - cards: The 8-card deck to evaluate
//   - strategy: The intended strategy
//   - synergyDB: The synergy database for pair analysis
//   - counterMatrix: The counter matrix for threat-based analysis (can be nil for basic analysis)
//
// Returns the V2 result plus detailed defensive coverage information.
func ScoreDeckV2WithCounterAnalysis(cards []CardCandidate, strategy Strategy, synergyDB *SynergyDatabase, counterMatrix *CounterMatrix) map[string]interface{} {
	// Get standard V2 result
	v2Result := ScoreDeckV2(cards, strategy, synergyDB)

	result := map[string]interface{}{
		"v2_score":            v2Result.FinalScore,
		"v2_counter_coverage": v2Result.CounterCoverageScore,
		"v2_details":          v2Result.Details,
	}

	// If counter matrix is provided, add detailed threat analysis
	if counterMatrix != nil {
		// Extract card names
		cardNames := make([]string, len(cards))
		for i, card := range cards {
			cardNames[i] = card.Name
		}

		// Create defensive scorer
		defensiveScorer := NewDefensiveScorer(counterMatrix)
		coverageReport := defensiveScorer.CalculateDefensiveCoverage(cardNames)

		// Create threat analyzer
		threatAnalyzer := NewThreatAnalyzer(counterMatrix)
		threatReport := threatAnalyzer.AnalyzeDeck(cardNames)

		// Add detailed counter analysis to result
		result["defensive_coverage"] = coverageReport
		result["threat_analysis"] = threatReport
		result["counter_matrix_available"] = true
	} else {
		result["counter_matrix_available"] = false
		result["defensive_coverage"] = nil
		result["threat_analysis"] = nil
	}

	return result
}

// CalculateCounterCoverageWithMatrix uses the CounterMatrix to calculate detailed counter coverage.
// This provides more sophisticated analysis than the basic calculateCounterCoverageScore.
//
// Parameters:
//   - cards: The deck cards as CardCandidate
//   - counterMatrix: The counter matrix with threat relationships
//
// Returns a detailed defensive coverage report.
func CalculateCounterCoverageWithMatrix(cards []CardCandidate, counterMatrix *CounterMatrix) *DefensiveCoverageReport {
	if counterMatrix == nil {
		return &DefensiveCoverageReport{
			OverallScore: 0.0,
			CoverageGaps: []string{"No counter matrix provided"},
		}
	}

	// Extract card names
	cardNames := make([]string, len(cards))
	for i, card := range cards {
		cardNames[i] = card.Name
	}

	// Create defensive scorer and calculate coverage
	defensiveScorer := NewDefensiveScorer(counterMatrix)
	return defensiveScorer.CalculateDefensiveCoverage(cardNames)
}
