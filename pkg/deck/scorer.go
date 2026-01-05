// Package deck provides card scoring functionality for intelligent deck building.
// The scoring algorithm considers card level, rarity, elixir cost, and strategic role.
package deck

import (
	"math"
	"os"
	"sort"
	"strconv"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// Elixir sweet spot is 3-4 elixir (most versatile cards)
// Higher cost cards get penalized slightly, very low cost cards also penalized
// All magic numbers have been moved to internal/config/constants.go

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
// 1. Level Ratio (0-1): Higher level cards score better (uses curve-based calculation if available)
// 2. Rarity Boost (1.0-1.2): Rarer cards get slight boost since they're harder to level
// 3. Elixir Weight (0-1): Cards around 3-4 elixir are optimal
// 4. Role Bonus (+0.05): Cards with defined roles get small bonus
// 5. Evolution Bonus (0-0.15): Cards with higher evolution level get additional score
//
// Evolution Bonus Formula: evolutionBonusWeight * (evolutionLevel / maxEvolutionLevel)
// This provides a linear scaling based on evolution progress.
//
// Score range: typically 0.0 to ~1.65, higher is better
//
// Note: This function uses linear level calculation for backward compatibility when
// card name is not available. For curve-based calculation, use ScoreCardCandidateWithCombat.
func ScoreCardWithEvolution(level, maxLevel int, rarity string, elixir int, role *CardRole, evolutionLevel, maxEvolutionLevel int) float64 {
	// Calculate level ratio using linear calculation (no card name available in this signature)
	levelRatio := calculateLevelRatio("", level, maxLevel, nil)

	// Get rarity boost multiplier
	rarityBoost := config.GetRarityWeight(rarity)

	// Calculate elixir efficiency weight
	// Penalize cards far from optimal elixir cost (3)
	elixirDiff := math.Abs(float64(elixir) - config.ElixirOptimal)
	elixirWeight := 1.0 - (elixirDiff / config.ElixirMaxDiff)

	// Add role bonus if card has defined strategic role
	roleBonus := 0.0
	if role != nil {
		roleBonus = config.RoleBonusValue
	}

	// Calculate evolution level bonus
	evolutionBonus := calculateEvolutionLevelBonus(evolutionLevel, maxEvolutionLevel)

	// Combine all factors into final score
	score := (levelRatio * config.LevelWeightFactor * rarityBoost) +
		(elixirWeight * config.ElixirWeightFactor) +
		roleBonus +
		evolutionBonus

	return score
}

// calculateLevelRatio calculates the level ratio for a card
// Uses curve-based calculation when levelCurve is available and cardName is provided.
// Falls back to linear calculation for backward compatibility.
func calculateLevelRatio(cardName string, level, maxLevel int, levelCurve *LevelCurve) float64 {
	if maxLevel <= 0 {
		return 0.0
	}

	// Use curve-based calculation if available and card name provided
	if levelCurve != nil && cardName != "" {
		return levelCurve.GetRelativeLevelRatio(cardName, level, maxLevel)
	}

	// Fall back to linear calculation
	return float64(level) / float64(maxLevel)
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
	return config.EvolutionBonusWeight * evolutionRatio
}

// ScoreCardCandidate calculates score for a CardCandidate and updates its Score field.
// This function now uses curve-based level calculation when available.
func ScoreCardCandidate(candidate *CardCandidate) float64 {
	score := scoreCardWithEvolutionInternal(
		candidate.Name,
		candidate.Level,
		candidate.MaxLevel,
		candidate.Rarity,
		candidate.Elixir,
		candidate.Role,
		candidate.EvolutionLevel,
		candidate.MaxEvolutionLevel,
		nil, // No level curve available in this context
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
	return config.DefaultCombatWeight
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
	combatScore := (dpsNormalized * config.CombatDPSWeight) +
		(hpNormalized * config.CombatHPWeight) +
		(roleEffectiveness * config.CombatRoleWeight)

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
// This function now includes evolution level bonus in the scoring and uses curve-based level calculation.
// Updates the candidate's Score field and returns the score.
func ScoreCardCandidateWithCombat(candidate *CardCandidate) float64 {
	score := scoreCardWithCombatAndEvolutionInternal(
		candidate.Name,
		candidate.Level,
		candidate.MaxLevel,
		candidate.Rarity,
		candidate.Elixir,
		candidate.Role,
		candidate.Stats,
		candidate.EvolutionLevel,
		candidate.MaxEvolutionLevel,
		nil, // No level curve available in this context
	)
	candidate.Score = score
	return score
}

// scoreCardWithCombatAndEvolutionInternal is the internal implementation that accepts cardName
// for curve-based level calculation
func scoreCardWithCombatAndEvolutionInternal(cardName string, level, maxLevel int, rarity string, elixir int, role *CardRole, stats *clashroyale.CombatStats, evolutionLevel, maxEvolutionLevel int, levelCurve *LevelCurve) float64 {
	// Calculate base score using curve-based level ratio
	baseScore := scoreCardWithEvolutionInternal(cardName, level, maxLevel, rarity, elixir, role, evolutionLevel, maxEvolutionLevel, levelCurve)

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

// scoreCardWithEvolutionInternal is the internal implementation that accepts cardName
// for curve-based level calculation
func scoreCardWithEvolutionInternal(cardName string, level, maxLevel int, rarity string, elixir int, role *CardRole, evolutionLevel, maxEvolutionLevel int, levelCurve *LevelCurve) float64 {
	// Calculate level ratio using curve-based calculation when card name is available
	levelRatio := calculateLevelRatio(cardName, level, maxLevel, levelCurve)

	// Get rarity boost multiplier - use priority bonus for deck building
	rarityBoost := config.GetRarityPriorityBonus(rarity)

	// Calculate elixir efficiency weight
	// Penalize cards far from optimal elixir cost (3)
	elixirDiff := math.Abs(float64(elixir) - config.ElixirOptimal)
	elixirWeight := 1.0 - (elixirDiff / config.ElixirMaxDiff)

	// Add role bonus if card has defined strategic role
	roleBonus := 0.0
	if role != nil {
		roleBonus = config.RoleBonusValue
	}

	// Calculate evolution level bonus
	evolutionBonus := calculateEvolutionLevelBonus(evolutionLevel, maxEvolutionLevel)

	// Combine all factors into final score
	score := (levelRatio * config.LevelWeightFactor * rarityBoost) +
		(elixirWeight * config.ElixirWeightFactor) +
		roleBonus +
		evolutionBonus

	return score
}

// ScoreAllCandidatesWithCombat scores a slice of CardCandidates in place using combat statistics
func ScoreAllCandidatesWithCombat(candidates []CardCandidate) {
	for i := range candidates {
		ScoreCardCandidateWithCombat(&candidates[i])
	}
}

// ScoreCardWithStrategy calculates enhanced score using strategy configuration
// Applies role bonuses and elixir targeting adjustments based on the strategy
// Uses curve-based level calculation when available
func ScoreCardWithStrategy(card *CardCandidate, role *CardRole, strategyConfig StrategyConfig, levelCurve *LevelCurve) float64 {
	// Start with base score using curve-based calculation (via internal functions with card name)
	var baseScore float64
	if card.Stats != nil {
		// Use combat-enhanced scoring if stats available
		baseScore = scoreCardWithCombatAndEvolutionInternal(
			card.Name,
			card.Level,
			card.MaxLevel,
			card.Rarity,
			card.Elixir,
			role,
			card.Stats,
			card.EvolutionLevel,
			card.MaxEvolutionLevel,
			levelCurve,
		)
	} else {
		// Fall back to traditional scoring (but still with curve-based level calculation)
		baseScore = scoreCardWithEvolutionInternal(
			card.Name,
			card.Level,
			card.MaxLevel,
			card.Rarity,
			card.Elixir,
			role,
			card.EvolutionLevel,
			card.MaxEvolutionLevel,
			levelCurve,
		)
	}

	// Apply strategy bonus (additive, level-agnostic)
	// Prefer additive bonuses over legacy multipliers for better differentiation
	strategyBonus := 0.0
	if role != nil {
		// Try additive bonuses first (new system)
		if bonus, exists := strategyConfig.RoleBonuses[*role]; exists {
			strategyBonus = bonus * GetStrategyScaling()
		} else if strategyConfig.RoleMultipliers != nil {
			// Fallback to legacy multiplier system for backward compatibility
			if mult, exists := strategyConfig.RoleMultipliers[*role]; exists {
				if mult > 1.0 {
					strategyBonus = baseScore * (mult - 1.0)
				} else if mult < 1.0 {
					strategyBonus = baseScore * (mult - 1.0) // Negative value
				}
			}
		}
	}

	// Apply archetype affinity bonus (helps on-archetype cards compete with higher-level cards)
	archetypeBonus := 0.0
	if affinityBonus, exists := strategyConfig.ArchetypeAffinity[card.Name]; exists {
		archetypeBonus = affinityBonus * GetStrategyScaling()
	}

	// Apply elixir targeting adjustment
	elixirAdjustment := calculateElixirAdjustment(card.Elixir, strategyConfig)

	// Combine: base score + strategy bonus + archetype bonus + elixir adjustment
	finalScore := baseScore + strategyBonus + archetypeBonus + elixirAdjustment

	return finalScore
}

// calculateElixirAdjustment calculates score adjustment based on strategy's elixir targets
// Penalizes cards outside the target elixir range, especially for cycle strategy
func calculateElixirAdjustment(elixir int, strategyConfig StrategyConfig) float64 {
	elixirCost := float64(elixir)
	targetMin := strategyConfig.TargetElixirMin
	targetMax := strategyConfig.TargetElixirMax

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
	if targetMax <= config.ElixirOptimal && elixirCost > config.ElixirCyclePenaltyThreshold {
		// Extra penalty for cards that would push average too high
		return -0.3 * distance
	}

	// Standard penalty for cards outside target range
	return -0.15 * distance
}
