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
	output := map[string]interface{}{
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"evaluation": map[string]interface{}{
			"deck": map[string]interface{}{
				"cards":          result.Deck,
				"average_elixir": result.AvgElixir,
			},
			"overall": map[string]interface{}{
				"score":  result.OverallScore,
				"rating": result.OverallRating,
			},
			"archetype": map[string]interface{}{
				"detected":   result.DetectedArchetype,
				"confidence": result.ArchetypeConfidence,
			},
			"category_scores": map[string]interface{}{
				"attack":       categoryScoreToMap(result.Attack),
				"defense":      categoryScoreToMap(result.Defense),
				"synergy":      categoryScoreToMap(result.Synergy),
				"versatility":  categoryScoreToMap(result.Versatility),
				"f2p_friendly": categoryScoreToMap(result.F2PFriendly),
			},
			"detailed_analysis": map[string]interface{}{
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

	// Marshal to pretty-printed JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonData), nil
}

// categoryScoreToMap converts a CategoryScore to a map for JSON serialization
func categoryScoreToMap(cs CategoryScore) map[string]interface{} {
	return map[string]interface{}{
		"score":      cs.Score,
		"rating":     cs.Rating,
		"assessment": cs.Assessment,
		"stars":      cs.Stars,
	}
}

// analysisSectionToMap converts an AnalysisSection to a map for JSON serialization
func analysisSectionToMap(as AnalysisSection) map[string]interface{} {
	return map[string]interface{}{
		"title":   as.Title,
		"summary": as.Summary,
		"details": as.Details,
		"score":   as.Score,
		"rating":  as.Rating,
	}
}

// synergyMatrixToMap converts a SynergyMatrix to a map for JSON serialization
func synergyMatrixToMap(sm SynergyMatrix) map[string]interface{} {
	// Convert pairs to array of maps
	pairs := make([]map[string]interface{}, len(sm.Pairs))
	for i, pair := range sm.Pairs {
		pairs[i] = map[string]interface{}{
			"card1":       pair.Card1,
			"card2":       pair.Card2,
			"score":       pair.Score,
			"description": pair.Description,
		}
	}

	return map[string]interface{}{
		"total_score":        sm.TotalScore,
		"average_synergy":    sm.AverageSynergy,
		"pair_count":         sm.PairCount,
		"max_possible_pairs": sm.MaxPossiblePairs,
		"synergy_coverage":   sm.SynergyCoverage,
		"pairs":              pairs,
	}
}

// generateRecommendationsJSON generates recommendations as structured JSON
func generateRecommendationsJSON(result *EvaluationResult) []map[string]interface{} {
	recs := generateRecommendations(result)
	jsonRecs := make([]map[string]interface{}, len(recs))

	for i, rec := range recs {
		jsonRecs[i] = map[string]interface{}{
			"priority":    rec.Priority,
			"title":       rec.Title,
			"description": rec.Description,
		}
	}

	return jsonRecs
}
