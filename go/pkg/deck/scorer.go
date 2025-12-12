// Package deck provides card scoring functionality for intelligent deck building.
// The scoring algorithm considers card level, rarity, elixir cost, and strategic role.
package deck

import (
	"math"
	"sort"
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
	// Calculate level ratio (0.0 to 1.0)
	levelRatio := 0.0
	if maxLevel > 0 {
		levelRatio = float64(level) / float64(maxLevel)
	}

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

	// Combine all factors into final score
	score := (levelRatio * levelWeightFactor * rarityBoost) +
		(elixirWeight * elixirWeightFactor) +
		roleBonus

	return score
}

// ScoreCardCandidate calculates score for a CardCandidate and updates its Score field
func ScoreCardCandidate(candidate *CardCandidate) float64 {
	score := ScoreCard(
		candidate.Level,
		candidate.MaxLevel,
		candidate.Rarity,
		candidate.Elixir,
		candidate.Role,
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
