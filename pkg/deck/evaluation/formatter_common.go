package evaluation

import (
	"fmt"
	"strings"
)

// Recommendation represents a deck improvement suggestion
type Recommendation struct {
	Priority    string // High, Medium, Low
	Title       string
	Description string
}

// formatStars converts star count (0-3) to visual star representation
func formatStars(count int) string {
	filled := "★"
	empty := "☆"

	stars := strings.Repeat(filled, count)
	stars += strings.Repeat(empty, 3-count)

	return stars
}

// formatSynergyStrength converts numeric strength to visual representation
func formatSynergyStrength(strength float64) string {
	switch {
	case strength >= 0.8:
		return "🔥🔥🔥 Excellent"
	case strength >= 0.6:
		return "🔥🔥 Strong"
	case strength >= 0.4:
		return "🔥 Good"
	default:
		return "• Moderate"
	}
}

// getSynergyStrengthLabel converts numeric synergy score to strength label
func getSynergyStrengthLabel(score float64) string {
	switch {
	case score >= 0.8:
		return "Excellent"
	case score >= 0.6:
		return "Strong"
	case score >= 0.4:
		return string(RatingGood)
	default:
		return "Moderate"
	}
}

// deriveStrengths extracts strengths from high-scoring categories
func deriveStrengths(result *EvaluationResult) []string {
	return deriveCategoryMessages(
		result,
		7.0,
		true,
		"Balanced deck with no major standout strengths",
		[]string{
			"Strong offensive potential",
			"Solid defensive capabilities",
			"Excellent card synergies",
			"Highly versatile and adaptable",
			"Easy to upgrade for F2P players",
			"All cards available - fully playable",
		},
	)
}

// deriveWeaknesses extracts weaknesses from low-scoring categories
func deriveWeaknesses(result *EvaluationResult) []string {
	return deriveCategoryMessages(
		result,
		5.0,
		false,
		"No critical weaknesses identified",
		[]string{
			"Weak offensive pressure",
			"Vulnerable defensive structure",
			"Poor card synergies",
			"Limited adaptability to different matchups",
			"Expensive upgrade path",
			"Missing cards - not fully playable",
		},
	)
}

func deriveCategoryMessages(result *EvaluationResult, threshold float64, useGTE bool, fallback string, messages []string) []string {
	scores := []float64{
		result.Attack.Score,
		result.Defense.Score,
		result.Synergy.Score,
		result.Versatility.Score,
		result.F2PFriendly.Score,
		result.Playability.Score,
	}

	out := make([]string, 0, len(messages))
	for i, score := range scores {
		if (useGTE && score >= threshold) || (!useGTE && score < threshold) {
			out = append(out, messages[i])
		}
	}
	if len(out) == 0 {
		out = append(out, fallback)
	}

	return out
}

// generateRecommendations creates improvement suggestions based on scores
func generateRecommendations(result *EvaluationResult) []Recommendation {
	var recs []Recommendation

	// Generate recommendations based on low scores
	if result.Attack.Score < 5.0 {
		recs = append(recs, Recommendation{
			Priority:    "High",
			Title:       "Add Win Conditions",
			Description: "Deck lacks reliable win conditions. Consider adding a strong win condition card like Hog Rider, Giant, or Balloon.",
		})
	}

	if result.Defense.Score < 5.0 {
		recs = append(recs, Recommendation{
			Priority:    "High",
			Title:       "Strengthen Defenses",
			Description: "Deck is vulnerable defensively. Add more anti-air troops or a defensive building like Tesla or Cannon.",
		})
	}

	if result.Synergy.Score < 5.0 {
		recs = append(recs, Recommendation{
			Priority:    "Medium",
			Title:       "Improve Card Synergies",
			Description: "Cards don't work well together. Look for combos like Giant + Witch or Hog + Freeze for better synergy.",
		})
	}

	if result.Versatility.Score < 5.0 {
		recs = append(recs, Recommendation{
			Priority:    "Medium",
			Title:       "Increase Versatility",
			Description: "Deck is too narrow in strategy. Add cards with different roles and elixir costs for more flexibility.",
		})
	}

	if result.F2PFriendly.Score < 5.0 {
		recs = append(recs, Recommendation{
			Priority:    "Low",
			Title:       "Consider F2P Alternatives",
			Description: "This deck is expensive to upgrade. Try swapping legendaries/epics for commons/rares with similar roles.",
		})
	}

	return recs
}

// formatAnalysisSection formats a single detailed analysis section
func formatAnalysisSection(section AnalysisSection) string {
	var output strings.Builder

	// Section header with score
	output.WriteString(fmt.Sprintf("▶ %s (%.1f/10 - %s)\n",
		section.Title,
		section.Score,
		section.Rating))
	output.WriteString("  ────────────────────────────────────────────────────────────────\n")

	// Summary
	output.WriteString(fmt.Sprintf("  %s\n\n", section.Summary))

	// Details as bulleted list
	if len(section.Details) > 0 {
		output.WriteString("  Key Points:\n")
		for _, detail := range section.Details {
			output.WriteString(fmt.Sprintf("  • %s\n", detail))
		}
	}

	output.WriteString("\n")

	return output.String()
}

// escapeCSV escapes a string for CSV format (handles commas, quotes, newlines)
func escapeCSV(s string) string {
	// If string contains comma, quote, or newline, wrap in quotes and escape quotes
	if strings.ContainsAny(s, ",\"\n") {
		// Escape quotes by doubling them
		escaped := strings.ReplaceAll(s, "\"", "\"\"")
		return "\"" + escaped + "\""
	}
	return s
}
