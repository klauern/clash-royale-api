// Package deck provides card scoring functionality for intelligent deck building.
// The scoring algorithm considers card level, rarity, elixir cost, and strategic role.
package deck

import (
	"math"
	"os"
	"sort"
	"strconv"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// Rarity weights boost scores for higher rarity cards since they're harder to level up
// Common cards (easiest to upgrade) get no boost, Champions get the highest boost
var rarityWeights = map[string]float64{
	"Common":    1.0,
	"Rare":      1.05,
	"Epic":      1.1,
	"Legendary": 1.15,
	"Champion":  1.2,
}

// Elixir sweet spot is 3-4 elixir (most versatile cards)
// Higher cost cards get penalized slightly, very low cost cards also penalized
const (
	elixirOptimal      = 3.0
	elixirWeightFactor = 0.15
	levelWeightFactor  = 1.2
	roleBonusValue     = 0.05

	// Combat stats integration weights
	defaultCombatWeight = 0.25 // 25% combat stats, 75% base scoring
	combatDPSWeight     = 0.4  // 40% of combat score from DPS efficiency
	combatHPWeight      = 0.4  // 40% of combat score from HP efficiency
	combatRoleWeight    = 0.2  // 20% of combat score from role-specific effectiveness

	// Evolution level bonus weights
	// Evolution level provides additional score boost based on how far the card is evolved
	evolutionBonusWeight = 0.15 // Base weight for evolution bonus (0.15 = up to +0.15 score at max evo)
)

// ScoreCard calculates a comprehensive score for a card based on multiple factors:
//
// 1. Level Ratio (0-1): Higher level cards score better
// 2. Rarity Boost (1.0-1.2): Rarer cards get slight boost since they're harder to level
// 3. Elixir Weight (0-1): Cards around 3-4 elixir are optimal
// 4. Role Bonus (+0.05): Cards with defined roles get small bonus
//
// Formula: (levelRatio * 1.2 * rarityBoost) + (elixirWeight * 0.15) + roleBonus
//
// Score range: typically 0.0 to ~1.5, higher is better
func ScoreCard(level, maxLevel int, rarity string, elixir int, role *CardRole) float64 {
	return ScoreCardWithEvolution(level, maxLevel, rarity, elixir, role, 0, 0)
}

// ScoreCardWithEvolution calculates a comprehensive score for a card including evolution level bonus.
//
// Factors considered:
// 1. Level Ratio (0-1): Higher level cards score better
// 2. Rarity Boost (1.0-1.2): Rarer cards get slight boost since they're harder to level
// 3. Elixir Weight (0-1): Cards around 3-4 elixir are optimal
// 4. Role Bonus (+0.05): Cards with defined roles get small bonus
// 5. Evolution Bonus (0-0.15): Cards with higher evolution level get additional score
//
// Evolution Bonus Formula: evolutionBonusWeight * (evolutionLevel / maxEvolutionLevel)
// This provides a linear scaling based on evolution progress.
//
// Score range: typically 0.0 to ~1.65, higher is better
func ScoreCardWithEvolution(level, maxLevel int, rarity string, elixir int, role *CardRole, evolutionLevel, maxEvolutionLevel int) float64 {
	// Calculate level ratio using legacy linear calculation for backward compatibility
	// In Phase 2, this will be replaced with curve-based calculation
	levelRatio := calculateLevelRatio(level, maxLevel)

	// Get rarity boost multiplier
	rarityBoost, exists := rarityWeights[rarity]
	if !exists {
		rarityBoost = 1.0 // Default to Common if unknown rarity
	}

	// Calculate elixir efficiency weight
	// Penalize cards far from optimal elixir cost (3)
	elixirDiff := math.Abs(float64(elixir) - elixirOptimal)
	elixirWeight := 1.0 - (elixirDiff / 9.0) // 9 is max meaningful diff

	// Add role bonus if card has defined strategic role
	roleBonus := 0.0
	if role != nil {
		roleBonus = roleBonusValue
	}

	// Calculate evolution level bonus
	evolutionBonus := calculateEvolutionLevelBonus(evolutionLevel, maxEvolutionLevel)

	// Combine all factors into final score
	score := (levelRatio * levelWeightFactor * rarityBoost) +
		(elixirWeight * elixirWeightFactor) +
		roleBonus +
		evolutionBonus

	return score
}

// calculateLevelRatio calculates the level ratio for a card
// Uses legacy linear calculation for backward compatibility
// TODO(clash-royale-api-gv5): Replace with curve-based calculation in Phase 2
func calculateLevelRatio(level, maxLevel int) float64 {
	levelRatio := 0.0
	if maxLevel > 0 {
		levelRatio = float64(level) / float64(maxLevel)
	}
	return levelRatio
}

// calculateEvolutionLevelBonus calculates the evolution level bonus for a card.
// The bonus is proportional to the evolution progress (evolutionLevel/maxEvolutionLevel).
//
// Formula: evolutionBonusWeight * (evolutionLevel / maxEvolutionLevel)
//
// Returns 0 if card has no evolution capability (maxEvolutionLevel == 0) or no evolution progress.
func calculateEvolutionLevelBonus(evolutionLevel, maxEvolutionLevel int) float64 {
	if maxEvolutionLevel <= 0 || evolutionLevel <= 0 {
		return 0.0
	}

	// Calculate evolution ratio (0.0 to 1.0)
	evolutionRatio := float64(evolutionLevel) / float64(maxEvolutionLevel)

	// Clamp ratio to valid range
	if evolutionRatio > 1.0 {
		evolutionRatio = 1.0
	}

	// Apply evolution bonus weight
	return evolutionBonusWeight * evolutionRatio
}

// ScoreCardCandidate calculates score for a CardCandidate and updates its Score field.
// This function now uses ScoreCardWithEvolution to include evolution level bonus.
func ScoreCardCandidate(candidate *CardCandidate) float64 {
	score := ScoreCardWithEvolution(
		candidate.Level,
		candidate.MaxLevel,
		candidate.Rarity,
		candidate.Elixir,
		candidate.Role,
		candidate.EvolutionLevel,
		candidate.MaxEvolutionLevel,
	)
	candidate.Score = score
	return score
}

// ScoreAllCandidates scores a slice of CardCandidates in place
func ScoreAllCandidates(candidates []CardCandidate) {
	for i := range candidates {
		ScoreCardCandidate(&candidates[i])
	}
}

// SortByScore sorts candidates by score in descending order (highest score first)
func SortByScore(candidates []CardCandidate) {
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})
}

// FilterByMinScore returns only candidates with score >= minScore
func FilterByMinScore(candidates []CardCandidate, minScore float64) []CardCandidate {
	filtered := make([]CardCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Score >= minScore {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

// GetTopN returns the top N highest-scoring candidates
// If N > len(candidates), returns all candidates
func GetTopN(candidates []CardCandidate, n int) []CardCandidate {
	// Sort first to ensure we get top scores
	SortByScore(candidates)

	if n >= len(candidates) {
		return candidates
	}
	return candidates[:n]
}

// FilterByRole returns only candidates with specified role
func FilterByRole(candidates []CardCandidate, role CardRole) []CardCandidate {
	filtered := make([]CardCandidate, 0)
	for _, candidate := range candidates {
		if candidate.Role != nil && *candidate.Role == role {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

// FilterByElixirRange returns candidates within specified elixir cost range (inclusive)
func FilterByElixirRange(candidates []CardCandidate, minElixir, maxElixir int) []CardCandidate {
	filtered := make([]CardCandidate, 0)
	for _, candidate := range candidates {
		if candidate.Elixir >= minElixir && candidate.Elixir <= maxElixir {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

// FilterByRarity returns candidates of specified rarity
func FilterByRarity(candidates []CardCandidate, rarity string) []CardCandidate {
	filtered := make([]CardCandidate, 0)
	for _, candidate := range candidates {
		if candidate.Rarity == rarity {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

// ExcludeCards removes cards with specified names from candidates
func ExcludeCards(candidates []CardCandidate, excludeNames []string) []CardCandidate {
	excludeMap := make(map[string]bool)
	for _, name := range excludeNames {
		excludeMap[name] = true
	}

	filtered := make([]CardCandidate, 0)
	for _, candidate := range candidates {
		if !excludeMap[candidate.Name] {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

// CalculateAvgElixir calculates average elixir cost from a slice of candidates
func CalculateAvgElixir(candidates []CardCandidate) float64 {
	if len(candidates) == 0 {
		return 0.0
	}

	total := 0
	for _, candidate := range candidates {
		total += candidate.Elixir
	}

	return float64(total) / float64(len(candidates))
}

// GetLevelDistribution returns count of cards at each level
func GetLevelDistribution(candidates []CardCandidate) map[int]int {
	distribution := make(map[int]int)
	for _, candidate := range candidates {
		distribution[candidate.Level]++
	}
	return distribution
}

// GetRarityDistribution returns count of cards by rarity
func GetRarityDistribution(candidates []CardCandidate) map[string]int {
	distribution := make(map[string]int)
	for _, candidate := range candidates {
		distribution[candidate.Rarity]++
	}
	return distribution
}

// getCombatWeight returns the combat stats weight from environment or default
func getCombatWeight() float64 {
	if weightStr := os.Getenv("COMBAT_STATS_WEIGHT"); weightStr != "" {
		if weight, err := strconv.ParseFloat(weightStr, 64); err == nil {
			// Clamp to reasonable range (0.0 to 1.0)
			if weight < 0 {
				return 0
			}
			if weight > 1 {
				return 1
			}
			return weight
		}
	}
	return defaultCombatWeight
}

// roleToString converts CardRole to string for combat stats integration
func roleToString(role *CardRole) string {
	if role == nil {
		return ""
	}
	switch *role {
	case RoleWinCondition:
		return "wincondition"
	case RoleBuilding:
		return "building"
	case RoleSupport:
		return "support"
	case RoleSpellBig:
		return "spell"
	case RoleSpellSmall:
		return "spell" // Treat both spell types as "spell" for combat analysis
	case RoleCycle:
		return "cycle"
	default:
		return ""
	}
}

// calculateCombatScore calculates combat effectiveness score from stats
func calculateCombatScore(stats *clashroyale.CombatStats, elixir int, role string) float64 {
	if stats == nil {
		return 0 // No combat stats available
	}

	// Calculate efficiency metrics
	dpsEfficiency := stats.DPSPerElixir(elixir)
	hpEfficiency := stats.HPPerElixir(elixir)
	roleEffectiveness := stats.RoleSpecificEffectiveness(role)

	// Normalize efficiency scores to 0-1 range
	// These normalization values are approximate and can be tuned
	dpsNormalized := math.Min(dpsEfficiency/50.0, 1.0) // ~50 DPS/elixir as excellent
	hpNormalized := math.Min(hpEfficiency/400.0, 1.0)  // ~400 HP/elixir as excellent

	// Combine combat factors with weights
	combatScore := (dpsNormalized * combatDPSWeight) +
		(hpNormalized * combatHPWeight) +
		(roleEffectiveness * combatRoleWeight)

	return math.Max(0, math.Min(1, combatScore)) // Clamp to 0-1 range
}

// ScoreCardWithCombat calculates enhanced score using combat statistics
// Combines traditional scoring with combat effectiveness analysis
// For backward compatibility, this does not include evolution level bonus.
// Use ScoreCardWithCombatAndEvolution for full scoring with evolution support.
func ScoreCardWithCombat(level, maxLevel int, rarity string, elixir int, role *CardRole, stats *clashroyale.CombatStats) float64 {
	return ScoreCardWithCombatAndEvolution(level, maxLevel, rarity, elixir, role, stats, 0, 0)
}

// ScoreCardWithCombatAndEvolution calculates enhanced score using combat statistics and evolution level.
// Combines traditional scoring with combat effectiveness analysis and evolution bonus.
func ScoreCardWithCombatAndEvolution(level, maxLevel int, rarity string, elixir int, role *CardRole, stats *clashroyale.CombatStats, evolutionLevel, maxEvolutionLevel int) float64 {
	// Calculate base score using evolution-aware algorithm
	baseScore := ScoreCardWithEvolution(level, maxLevel, rarity, elixir, role, evolutionLevel, maxEvolutionLevel)

	// Get combat weight from environment (default 0.25)
	combatWeight := getCombatWeight()

	if combatWeight == 0 || stats == nil {
		// Combat stats disabled or not available, return base score only
		return baseScore
	}

	// Calculate combat score
	roleStr := roleToString(role)
	combatScore := calculateCombatScore(stats, elixir, roleStr)

	// Combine base score and combat score
	// Final Score = (Base Score × (1 - combatWeight)) + (Combat Score × combatWeight)
	finalScore := (baseScore * (1 - combatWeight)) + (combatScore * combatWeight)

	return finalScore
}

// ScoreCardCandidateWithCombat calculates enhanced score for a CardCandidate using combat statistics.
// This function now includes evolution level bonus in the scoring.
// Updates the candidate's Score field and returns the score.
func ScoreCardCandidateWithCombat(candidate *CardCandidate) float64 {
	score := ScoreCardWithCombatAndEvolution(
		candidate.Level,
		candidate.MaxLevel,
		candidate.Rarity,
		candidate.Elixir,
		candidate.Role,
		candidate.Stats,
		candidate.EvolutionLevel,
		candidate.MaxEvolutionLevel,
	)
	candidate.Score = score
	return score
}

// ScoreAllCandidatesWithCombat scores a slice of CardCandidates in place using combat statistics
func ScoreAllCandidatesWithCombat(candidates []CardCandidate) {
	for i := range candidates {
		ScoreCardCandidateWithCombat(&candidates[i])
	}
}

// ScoreCardWithStrategy calculates enhanced score using strategy configuration
// Applies role multipliers and elixir targeting adjustments based on the strategy
func ScoreCardWithStrategy(card *CardCandidate, role *CardRole, config StrategyConfig) float64 {
	// Start with base score (using existing scoring logic)
	var baseScore float64
	if card.Stats != nil {
		// Use combat-enhanced scoring if stats available
		baseScore = ScoreCardWithCombatAndEvolution(
			card.Level,
			card.MaxLevel,
			card.Rarity,
			card.Elixir,
			role,
			card.Stats,
			card.EvolutionLevel,
			card.MaxEvolutionLevel,
		)
	} else {
		// Fall back to traditional scoring
		baseScore = ScoreCardWithEvolution(
			card.Level,
			card.MaxLevel,
			card.Rarity,
			card.Elixir,
			role,
			card.EvolutionLevel,
			card.MaxEvolutionLevel,
		)
	}

	// Apply role multiplier from strategy config
	roleMultiplier := 1.0
	if role != nil {
		if mult, exists := config.RoleMultipliers[*role]; exists {
			roleMultiplier = mult
		}
	}

	// Apply elixir targeting adjustment
	elixirAdjustment := calculateElixirAdjustment(card.Elixir, config)

	// Combine: base score × role multiplier + elixir adjustment
	finalScore := (baseScore * roleMultiplier) + elixirAdjustment

	return finalScore
}

// calculateElixirAdjustment calculates score adjustment based on strategy's elixir targets
// Penalizes cards outside the target elixir range, especially for cycle strategy
func calculateElixirAdjustment(elixir int, config StrategyConfig) float64 {
	elixirCost := float64(elixir)
	targetMin := config.TargetElixirMin
	targetMax := config.TargetElixirMax

	// If within target range, no penalty
	if elixirCost >= targetMin && elixirCost <= targetMax {
		return 0.0
	}

	// Calculate distance from target range
	var distance float64
	if elixirCost < targetMin {
		distance = targetMin - elixirCost
	} else {
		distance = elixirCost - targetMax
	}

	// For cycle strategy (low elixir target), heavily penalize high-cost cards
	// This is particularly important to keep the deck fast
	if targetMax <= 3.0 && elixirCost > 4.0 {
		// Extra penalty for cards that would push average too high
		return -0.3 * distance
	}

	// Standard penalty for cards outside target range
	return -0.15 * distance
}
