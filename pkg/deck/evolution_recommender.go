package deck

import (
	"fmt"
	"sort"
)

// EvolutionShardSource provides shard counts for cards.
type EvolutionShardSource interface {
	GetShardCount(cardName string) int
}

// mapShardSource wraps a map[string]int to implement EvolutionShardSource.
type mapShardSource map[string]int

func (m mapShardSource) GetShardCount(cardName string) int {
	if m == nil {
		return 0
	}
	return m[cardName]
}

// EvolutionRecommendation represents a recommended evolution with scoring.
type EvolutionRecommendation struct {
	CardName           string  `json:"card_name"`
	CurrentShards      int     `json:"current_shards"`
	ShardsNeeded       int     `json:"shards_needed"`
	CompletionPercent  float64 `json:"completion_percent"`
	CardLevel          int     `json:"card_level"`
	MaxLevel           int     `json:"max_level"`
	LevelRatio         float64 `json:"level_ratio"`
	Role               string  `json:"role,omitempty"`
	EvolutionMaxLevel  int     `json:"evolution_max_level"`
	RecommendationScore float64 `json:"recommendation_score"`
	Reasons            []string `json:"reasons"`
}

// EvolutionRecommender recommends evolutions based on card levels, shard inventory,
// and strategic impact.
type EvolutionRecommender struct {
	shardSource    EvolutionShardSource
	unlockedEvolutions map[string]bool
	shardsPerEvolution int // Usually 10 for single evolution, may vary
}

// NewEvolutionRecommender creates a new recommender with the given shard source.
func NewEvolutionRecommender(shardSource map[string]int, unlockedEvolutions []string) *EvolutionRecommender {
	unlocked := make(map[string]bool, len(unlockedEvolutions))
	for _, name := range unlockedEvolutions {
		unlocked[name] = true
	}
	return &EvolutionRecommender{
		shardSource:       mapShardSource(shardSource),
		unlockedEvolutions: unlocked,
		shardsPerEvolution: 10, // Standard: 10 shards per evolution level
	}
}

// SetShardsPerEvolution sets the number of shards required per evolution level.
func (r *EvolutionRecommender) SetShardsPerEvolution(count int) {
	r.shardsPerEvolution = count
}

// Recommend generates evolution recommendations from a list of card candidates.
func (r *EvolutionRecommender) Recommend(candidates []CardCandidate, topN int) []EvolutionRecommendation {
	var recommendations []EvolutionRecommendation

	for _, candidate := range candidates {
		// Skip cards without evolution capability
		if candidate.MaxEvolutionLevel <= 0 {
			continue
		}

		// Skip already unlocked evolutions
		if r.unlockedEvolutions[candidate.Name] {
			continue
		}

		rec := r.evaluateCandidate(candidate)
		if rec != nil {
			recommendations = append(recommendations, *rec)
		}
	}

	// Sort by recommendation score (descending)
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].RecommendationScore > recommendations[j].RecommendationScore
	})

	// Return top N
	if topN > 0 && topN < len(recommendations) {
		recommendations = recommendations[:topN]
	}

	return recommendations
}

// evaluateCandidate scores a single card candidate for evolution recommendation.
func (r *EvolutionRecommender) evaluateCandidate(candidate CardCandidate) *EvolutionRecommendation {
	cardName := candidate.Name
	currentShards := r.shardSource.GetShardCount(cardName)
	shardsNeeded := r.shardsPerEvolution

	levelRatio := candidate.LevelRatio()

	var reasons []string
	score := 0.0

	// 1. Level ratio bonus (0-40 points)
	// Higher level = more valuable evolution
	levelBonus := levelRatio * 40.0
	score += levelBonus
	if levelRatio >= 0.9 {
		reasons = append(reasons, "Near max level, evolution will be very effective")
	} else if levelRatio >= 0.7 {
		reasons = append(reasons, "High level, evolution will be effective")
	}

	// 2. Shard progress bonus (0-30 points)
	// Having shards already makes it more actionable
	completionPercent := float64(currentShards) / float64(shardsNeeded) * 100.0
	if completionPercent > 100 {
		completionPercent = 100
	}
	shardBonus := completionPercent / 100.0 * 30.0
	score += shardBonus
	if currentShards >= shardsNeeded {
		reasons = append(reasons, fmt.Sprintf("Already have enough shards (%d/%d)", currentShards, shardsNeeded))
	} else if currentShards >= shardsNeeded/2 {
		reasons = append(reasons, fmt.Sprintf("Halfway to required shards (%d/%d)", currentShards, shardsNeeded))
	}

	// 3. Role priority bonus (0-20 points)
	if candidate.Role != nil {
		roleBonus := r.rolePriorityScore(*candidate.Role)
		score += roleBonus
		if roleBonus > 15 {
			reasons = append(reasons, fmt.Sprintf("High-priority role: %s", GetRoleDescription(*candidate.Role)))
		}
	}

	// 4. Evolution level bonus (0-10 points)
	// Multi-evolution cards get slight bonus for more potential
	if candidate.MaxEvolutionLevel > 1 {
		evoBonus := float64(candidate.MaxEvolutionLevel-1) * 5.0
		score += evoBonus
		reasons = append(reasons, fmt.Sprintf("Multi-evolution card (%d levels)", candidate.MaxEvolutionLevel))
	}

	roleStr := ""
	if candidate.Role != nil {
		roleStr = GetRoleDescription(*candidate.Role)
	}

	return &EvolutionRecommendation{
		CardName:            cardName,
		CurrentShards:       currentShards,
		ShardsNeeded:        shardsNeeded,
		CompletionPercent:   completionPercent,
		CardLevel:           candidate.Level,
		MaxLevel:            candidate.MaxLevel,
		LevelRatio:          levelRatio,
		Role:                roleStr,
		EvolutionMaxLevel:   candidate.MaxEvolutionLevel,
		RecommendationScore: score,
		Reasons:             reasons,
	}
}

// rolePriorityScore returns a priority score for a card role (0-20).
func (r *EvolutionRecommender) rolePriorityScore(role CardRole) float64 {
	// Role priorities for evolution recommendations
	// Win conditions and high-impact support get highest priority
	switch role {
	case RoleWinCondition:
		return 20.0
	case RoleSupport:
		return 15.0
	case RoleSpellBig:
		return 12.0
	case RoleSpellSmall:
		return 8.0
	case RoleBuilding:
		return 10.0
	case RoleCycle:
		return 5.0
	default:
		return 10.0
	}
}

// FormatRecommendations formats recommendations for display.
func FormatRecommendations(recs []EvolutionRecommendation, showReasons bool) string {
	if len(recs) == 0 {
		return "No evolution recommendations available."
	}

	output := fmt.Sprintf("Evolution Recommendations (Top %d):\n", len(recs))
	output += "─────────────────────────────────────────────────────────────\n"

	for i, rec := range recs {
		output += fmt.Sprintf("\n%d. %s (Score: %.1f)\n", i+1, rec.CardName, rec.RecommendationScore)
		output += fmt.Sprintf("   Progress: %d/%d shards (%.0f%%)\n",
			rec.CurrentShards, rec.ShardsNeeded, rec.CompletionPercent)
		output += fmt.Sprintf("   Card Level: %d/%d (%.0f%%)\n",
			rec.CardLevel, rec.MaxLevel, rec.LevelRatio*100)
		if rec.Role != "" {
			output += fmt.Sprintf("   Role: %s\n", rec.Role)
		}
		if rec.EvolutionMaxLevel > 1 {
			output += fmt.Sprintf("   Evolution: %d levels available\n", rec.EvolutionMaxLevel)
		}
		if showReasons && len(rec.Reasons) > 0 {
			output += "   Reasons:\n"
			for _, reason := range rec.Reasons {
				output += fmt.Sprintf("   • %s\n", reason)
			}
		}
	}

	return output
}
