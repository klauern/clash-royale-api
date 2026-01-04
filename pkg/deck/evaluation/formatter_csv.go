package evaluation

import (
	"fmt"
	"strings"
)

// FormatCSV formats an EvaluationResult as multiple CSV tables for spreadsheet analysis
// Returns a multi-section CSV document with headers for each category
func FormatCSV(result *EvaluationResult) string {
	var output strings.Builder

	// Section 1: Overall Summary
	output.WriteString("# DECK EVALUATION SUMMARY\n")
	output.WriteString("Cards," + strings.Join(result.Deck, ";") + "\n")
	output.WriteString(fmt.Sprintf("Average Elixir,%.2f\n", result.AvgElixir))
	output.WriteString(fmt.Sprintf("Overall Score,%.2f\n", result.OverallScore))
	output.WriteString(fmt.Sprintf("Overall Rating,%s\n", result.OverallRating))
	output.WriteString(fmt.Sprintf("Archetype,%s\n", result.DetectedArchetype))
	output.WriteString(fmt.Sprintf("Archetype Confidence,%.2f\n\n", result.ArchetypeConfidence*100))

	// Section 2: Category Scores
	output.WriteString("# CATEGORY SCORES\n")
	output.WriteString("Category,Score,Rating,Stars,Assessment\n")
	output.WriteString(formatCategoryScoreCSV("Attack", result.Attack))
	output.WriteString(formatCategoryScoreCSV("Defense", result.Defense))
	output.WriteString(formatCategoryScoreCSV("Synergy", result.Synergy))
	output.WriteString(formatCategoryScoreCSV("Versatility", result.Versatility))
	output.WriteString(formatCategoryScoreCSV("F2P Friendly", result.F2PFriendly))
	output.WriteString("\n")

	// Section 3: Detailed Analysis
	output.WriteString("# DETAILED ANALYSIS\n")
	output.WriteString("Section,Score,Rating,Summary,Details\n")
	output.WriteString(formatAnalysisSectionCSV(result.DefenseAnalysis))
	output.WriteString(formatAnalysisSectionCSV(result.AttackAnalysis))
	output.WriteString(formatAnalysisSectionCSV(result.BaitAnalysis))
	output.WriteString(formatAnalysisSectionCSV(result.CycleAnalysis))
	output.WriteString(formatAnalysisSectionCSV(result.LadderAnalysis))
	output.WriteString("\n")

	// Section 4: Synergy Matrix
	if result.SynergyMatrix.PairCount > 0 {
		output.WriteString("# SYNERGY MATRIX\n")
		output.WriteString(fmt.Sprintf("Total Synergy Score,%.2f\n", result.SynergyMatrix.TotalScore))
		output.WriteString(fmt.Sprintf("Average Synergy,%.2f\n", result.SynergyMatrix.AverageSynergy))
		output.WriteString(fmt.Sprintf("Pair Count,%d\n", result.SynergyMatrix.PairCount))
		output.WriteString(fmt.Sprintf("Max Possible Pairs,%d\n", result.SynergyMatrix.MaxPossiblePairs))
		output.WriteString(fmt.Sprintf("Synergy Coverage,%.2f%%\n\n", result.SynergyMatrix.SynergyCoverage))

		output.WriteString("# SYNERGY PAIRS\n")
		output.WriteString("Card 1,Card 2,Score,Strength,Description\n")
		for _, pair := range result.SynergyMatrix.Pairs {
			strength := getSynergyStrengthLabel(pair.Score)
			escapedDesc := escapeCSV(pair.Description)
			output.WriteString(fmt.Sprintf("%s,%s,%.2f,%s,%s\n",
				pair.Card1, pair.Card2, pair.Score*100, strength, escapedDesc))
		}
		output.WriteString("\n")
	}

	// Section 5: Recommendations
	recs := generateRecommendations(result)
	if len(recs) > 0 {
		output.WriteString("# RECOMMENDATIONS\n")
		output.WriteString("Priority,Title,Description\n")
		for _, rec := range recs {
			escapedTitle := escapeCSV(rec.Title)
			escapedDesc := escapeCSV(rec.Description)
			output.WriteString(fmt.Sprintf("%s,%s,%s\n", rec.Priority, escapedTitle, escapedDesc))
		}
		output.WriteString("\n")
	}

	// Section 6: Strengths and Weaknesses
	output.WriteString("# STRENGTHS\n")
	strengths := deriveStrengths(result)
	for _, strength := range strengths {
		output.WriteString(escapeCSV(strength) + "\n")
	}
	output.WriteString("\n")

	output.WriteString("# WEAKNESSES\n")
	weaknesses := deriveWeaknesses(result)
	for _, weakness := range weaknesses {
		output.WriteString(escapeCSV(weakness) + "\n")
	}

	return output.String()
}

// formatCategoryScoreCSV formats a single category score as a CSV row
func formatCategoryScoreCSV(category string, score CategoryScore) string {
	escapedAssessment := escapeCSV(score.Assessment)
	return fmt.Sprintf("%s,%.2f,%s,%d,%s\n",
		category, score.Score, score.Rating, score.Stars, escapedAssessment)
}

// formatAnalysisSectionCSV formats a single analysis section as a CSV row
func formatAnalysisSectionCSV(section AnalysisSection) string {
	escapedSummary := escapeCSV(section.Summary)
	escapedDetails := escapeCSV(strings.Join(section.Details, "; "))
	return fmt.Sprintf("%s,%.2f,%s,%s,%s\n",
		section.Title, section.Score, section.Rating, escapedSummary, escapedDetails)
}

// getSynergyStrengthLabel converts numeric synergy score to strength label
func getSynergyStrengthLabel(score float64) string {
	if score >= 0.8 {
		return "Excellent"
	} else if score >= 0.6 {
		return "Strong"
	} else if score >= 0.4 {
		return "Good"
	} else {
		return "Moderate"
	}
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
