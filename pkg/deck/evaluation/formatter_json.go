package evaluation

import (
	"encoding/json"
	"fmt"
	"time"
)

// FormatJSON formats an EvaluationResult as structured JSON for programmatic use
// Includes all evaluation data with proper nesting and metadata
func FormatJSON(result *EvaluationResult) (string, error) {
	// Create output structure with metadata
	output := map[string]any{
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"evaluation": map[string]any{
			"deck": map[string]any{
				"cards":          result.Deck,
				"average_elixir": result.AvgElixir,
			},
			"overall": map[string]any{
				"score":  result.OverallScore,
				"rating": result.OverallRating,
			},
			"archetype": map[string]any{
				"detected":   result.DetectedArchetype,
				"confidence": result.ArchetypeConfidence,
			},
			"category_scores": map[string]any{
				"attack":       categoryScoreToMap(result.Attack),
				"defense":      categoryScoreToMap(result.Defense),
				"synergy":      categoryScoreToMap(result.Synergy),
				"versatility":  categoryScoreToMap(result.Versatility),
				"f2p_friendly": categoryScoreToMap(result.F2PFriendly),
			},
			"detailed_analysis": map[string]any{
				"defense": analysisSectionToMap(result.DefenseAnalysis),
				"attack":  analysisSectionToMap(result.AttackAnalysis),
				"bait":    analysisSectionToMap(result.BaitAnalysis),
				"cycle":   analysisSectionToMap(result.CycleAnalysis),
				"ladder":  analysisSectionToMap(result.LadderAnalysis),
			},
			"synergy_matrix":  synergyMatrixToMap(result.SynergyMatrix),
			"recommendations": generateRecommendationsJSON(result),
		},
	}

	if result.OverallBreakdown != nil {
		evaluationMap, ok := output["evaluation"].(map[string]any)
		if ok {
			overallMap, ok := evaluationMap["overall"].(map[string]any)
			if ok {
				overallMap["breakdown"] = map[string]any{
					"base_score":           result.OverallBreakdown.BaseScore,
					"contextual_score":     result.OverallBreakdown.ContextualScore,
					"ladder_score":         result.OverallBreakdown.LadderScore,
					"normalized_score":     result.OverallBreakdown.NormalizedScore,
					"final_score":          result.OverallBreakdown.FinalScore,
					"deck_level_ratio":     result.OverallBreakdown.DeckLevelRatio,
					"normalization_factor": result.OverallBreakdown.NormalizationFactor,
				}
			}
		}
	}

	// Marshal to pretty-printed JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonData), nil
}

// categoryScoreToMap converts a CategoryScore to a map for JSON serialization
func categoryScoreToMap(cs CategoryScore) map[string]any {
	return map[string]any{
		"score":      cs.Score,
		"rating":     cs.Rating,
		"assessment": cs.Assessment,
		"stars":      cs.Stars,
	}
}

// analysisSectionToMap converts an AnalysisSection to a map for JSON serialization
func analysisSectionToMap(as AnalysisSection) map[string]any {
	return map[string]any{
		"title":   as.Title,
		"summary": as.Summary,
		"details": as.Details,
		"score":   as.Score,
		"rating":  as.Rating,
	}
}

// synergyMatrixToMap converts a SynergyMatrix to a map for JSON serialization
func synergyMatrixToMap(sm SynergyMatrix) map[string]any {
	// Convert pairs to array of maps
	pairs := make([]map[string]any, len(sm.Pairs))
	for i, pair := range sm.Pairs {
		pairs[i] = map[string]any{
			"card1":       pair.Card1,
			"card2":       pair.Card2,
			"score":       pair.Score,
			"description": pair.Description,
		}
	}

	return map[string]any{
		"total_score":        sm.TotalScore,
		"average_synergy":    sm.AverageSynergy,
		"pair_count":         sm.PairCount,
		"max_possible_pairs": sm.MaxPossiblePairs,
		"synergy_coverage":   sm.SynergyCoverage,
		"pairs":              pairs,
	}
}

// generateRecommendationsJSON generates recommendations as structured JSON
func generateRecommendationsJSON(result *EvaluationResult) []map[string]any {
	recs := generateRecommendations(result)
	jsonRecs := make([]map[string]any, len(recs))

	for i, rec := range recs {
		jsonRecs[i] = map[string]any{
			"priority":    rec.Priority,
			"title":       rec.Title,
			"description": rec.Description,
		}
	}

	return jsonRecs
}
