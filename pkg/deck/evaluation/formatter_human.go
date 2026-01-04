package evaluation

import (
	"fmt"
	"strings"
)

// FormatHuman formats an EvaluationResult in DeckShop.pro-style human-readable format
// with visual star ratings and comprehensive analysis display
func FormatHuman(result *EvaluationResult) string {
	var output strings.Builder

	// Build all sections
	output.WriteString(formatHeader(result))
	output.WriteString(formatScoringGrid(result))
	output.WriteString(formatDetailedAnalysis(result))
	output.WriteString(formatSynergyMatrix(result))
	output.WriteString(formatCounterAnalysis(result))
	output.WriteString(formatRecommendations(result))
	output.WriteString(formatCopyDeckLink(result))
	output.WriteString(formatFooter(result))

	return output.String()
}

// formatHeader creates the deck overview section with player info and basic stats
func formatHeader(result *EvaluationResult) string {
	var header strings.Builder

	// Top border
	header.WriteString("╔═══════════════════════════════════════════════════════════════════════╗\n")
	header.WriteString("║                        DECK EVALUATION REPORT                         ║\n")
	header.WriteString("╚═══════════════════════════════════════════════════════════════════════╝\n\n")

	// Deck cards
	header.WriteString("🃏 Deck Cards:\n")
	header.WriteString("  " + strings.Join(result.Deck, " • ") + "\n\n")

	// Basic stats
	header.WriteString(fmt.Sprintf("📊 Average Elixir: %.2f\n", result.AvgElixir))
	header.WriteString(fmt.Sprintf("🎯 Archetype: %s (%.0f%% confidence)\n\n",
		strings.Title(string(result.DetectedArchetype)),
		result.ArchetypeConfidence*100))

	// Overall score with large visual display
	header.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	header.WriteString(fmt.Sprintf("                  OVERALL SCORE: %.1f/10 - %s\n",
		result.OverallScore, result.OverallRating))
	header.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	return header.String()
}

// formatScoringGrid creates the 5-category scoring section with stars and assessments
func formatScoringGrid(result *EvaluationResult) string {
	var grid strings.Builder

	grid.WriteString("📈 Category Scores:\n\n")

	// Define categories in display order
	categories := []struct {
		Name  string
		Score CategoryScore
		Icon  string
	}{
		{"Attack", result.Attack, "⚔️"},
		{"Defense", result.Defense, "🛡️"},
		{"Synergy", result.Synergy, "🔗"},
		{"Versatility", result.Versatility, "🎭"},
		{"F2P Friendly", result.F2PFriendly, "💰"},
	}

	// Display each category
	for _, cat := range categories {
		stars := formatStars(cat.Score.Stars)
		grid.WriteString(fmt.Sprintf("  %s %-14s %s  %s\n",
			cat.Icon,
			cat.Name+":",
			stars,
			cat.Score.Rating))
		grid.WriteString(fmt.Sprintf("     Score: %.1f/10 - %s\n\n",
			cat.Score.Score,
			cat.Score.Assessment))
	}

	return grid.String()
}

// formatStars converts star count (0-3) to visual star representation
func formatStars(count int) string {
	filled := "★"
	empty := "☆"

	stars := strings.Repeat(filled, count)
	stars += strings.Repeat(empty, 3-count)

	return stars
}

// formatDetailedAnalysis creates the detailed analysis sections
func formatDetailedAnalysis(result *EvaluationResult) string {
	var analysis strings.Builder

	analysis.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	analysis.WriteString("                         DETAILED ANALYSIS\n")
	analysis.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	// Define sections in display order
	sections := []AnalysisSection{
		result.DefenseAnalysis,
		result.AttackAnalysis,
		result.BaitAnalysis,
		result.CycleAnalysis,
		result.LadderAnalysis,
	}

	for _, section := range sections {
		analysis.WriteString(formatAnalysisSection(section))
	}

	return analysis.String()
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

// formatSynergyMatrix formats the synergy matrix with attack/defense grids
func formatSynergyMatrix(result *EvaluationResult) string {
	var matrix strings.Builder

	if result.SynergyMatrix.PairCount == 0 {
		return ""
	}

	matrix.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	matrix.WriteString("                          SYNERGY MATRIX\n")
	matrix.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	// Overall synergy stats
	matrix.WriteString(fmt.Sprintf("Total Synergy Score: %.1f/10\n", result.SynergyMatrix.TotalScore))
	matrix.WriteString(fmt.Sprintf("Synergy Pairs Found: %d/%d\n",
		result.SynergyMatrix.PairCount,
		result.SynergyMatrix.MaxPossiblePairs))
	matrix.WriteString(fmt.Sprintf("Card Coverage: %.1f%%\n\n", result.SynergyMatrix.SynergyCoverage))

	// Top synergy pairs
	if len(result.SynergyMatrix.Pairs) > 0 {
		matrix.WriteString("🔥 Top Synergy Pairs:\n\n")

		// Show top 5 pairs
		maxPairs := 5
		if len(result.SynergyMatrix.Pairs) < maxPairs {
			maxPairs = len(result.SynergyMatrix.Pairs)
		}

		for i := 0; i < maxPairs; i++ {
			pair := result.SynergyMatrix.Pairs[i]
			strength := formatSynergyStrength(pair.Score)
			matrix.WriteString(fmt.Sprintf("  %d. %s ↔ %s\n",
				i+1,
				pair.Card1,
				pair.Card2))
			matrix.WriteString(fmt.Sprintf("     %s (%.1f%%) - %s\n\n",
				strength,
				pair.Score*100,
				pair.Description))
		}
	}

	return matrix.String()
}

// formatSynergyStrength converts numeric strength to visual representation
func formatSynergyStrength(strength float64) string {
	if strength >= 0.8 {
		return "🔥🔥🔥 Excellent"
	} else if strength >= 0.6 {
		return "🔥🔥 Strong"
	} else if strength >= 0.4 {
		return "🔥 Good"
	} else {
		return "• Moderate"
	}
}

// formatCounterAnalysis formats strengths and weaknesses
func formatCounterAnalysis(result *EvaluationResult) string {
	var counter strings.Builder

	counter.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	counter.WriteString("                        COUNTER ANALYSIS\n")
	counter.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	// Strengths (derived from high-scoring categories)
	counter.WriteString("✅ Strengths:\n")
	strengths := deriveStrengths(result)
	for _, strength := range strengths {
		counter.WriteString(fmt.Sprintf("  • %s\n", strength))
	}
	counter.WriteString("\n")

	// Weaknesses (derived from low-scoring categories)
	counter.WriteString("⚠️  Weaknesses:\n")
	weaknesses := deriveWeaknesses(result)
	for _, weakness := range weaknesses {
		counter.WriteString(fmt.Sprintf("  • %s\n", weakness))
	}
	counter.WriteString("\n")

	return counter.String()
}

// deriveStrengths extracts strengths from high-scoring categories
func deriveStrengths(result *EvaluationResult) []string {
	var strengths []string

	if result.Attack.Score >= 7.0 {
		strengths = append(strengths, "Strong offensive potential")
	}
	if result.Defense.Score >= 7.0 {
		strengths = append(strengths, "Solid defensive capabilities")
	}
	if result.Synergy.Score >= 7.0 {
		strengths = append(strengths, "Excellent card synergies")
	}
	if result.Versatility.Score >= 7.0 {
		strengths = append(strengths, "Highly versatile and adaptable")
	}
	if result.F2PFriendly.Score >= 7.0 {
		strengths = append(strengths, "Easy to upgrade for F2P players")
	}

	if len(strengths) == 0 {
		strengths = append(strengths, "Balanced deck with no major standout strengths")
	}

	return strengths
}

// deriveWeaknesses extracts weaknesses from low-scoring categories
func deriveWeaknesses(result *EvaluationResult) []string {
	var weaknesses []string

	if result.Attack.Score < 5.0 {
		weaknesses = append(weaknesses, "Weak offensive pressure")
	}
	if result.Defense.Score < 5.0 {
		weaknesses = append(weaknesses, "Vulnerable defensive structure")
	}
	if result.Synergy.Score < 5.0 {
		weaknesses = append(weaknesses, "Poor card synergies")
	}
	if result.Versatility.Score < 5.0 {
		weaknesses = append(weaknesses, "Limited adaptability to different matchups")
	}
	if result.F2PFriendly.Score < 5.0 {
		weaknesses = append(weaknesses, "Expensive upgrade path")
	}

	if len(weaknesses) == 0 {
		weaknesses = append(weaknesses, "No critical weaknesses identified")
	}

	return weaknesses
}

// formatRecommendations formats improvement suggestions with priorities
func formatRecommendations(result *EvaluationResult) string {
	var recs strings.Builder

	recs.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	recs.WriteString("                        RECOMMENDATIONS\n")
	recs.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	recommendations := generateRecommendations(result)

	if len(recommendations) == 0 {
		recs.WriteString("  ✨ This deck is well-balanced! No major improvements needed.\n\n")
		return recs.String()
	}

	for i, rec := range recommendations {
		priority := "🔴"
		if rec.Priority == "Medium" {
			priority = "🟡"
		} else if rec.Priority == "Low" {
			priority = "🟢"
		}

		recs.WriteString(fmt.Sprintf("%s %d. %s\n", priority, i+1, rec.Title))
		recs.WriteString(fmt.Sprintf("   %s\n\n", rec.Description))
	}

	return recs.String()
}

// Recommendation represents a deck improvement suggestion
type Recommendation struct {
	Priority    string // High, Medium, Low
	Title       string
	Description string
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

// formatCopyDeckLink generates a shareable deck link
func formatCopyDeckLink(result *EvaluationResult) string {
	var link strings.Builder

	link.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	link.WriteString("                          SHARE DECK\n")
	link.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	// Generate RoyaleAPI deck link format
	// Format: https://royaleapi.com/decks/stats/card1,card2,card3,card4,card5,card6,card7,card8
	deckString := strings.Join(result.Deck, ",")
	// Convert to lowercase and replace spaces with dashes for URL
	deckString = strings.ToLower(strings.ReplaceAll(deckString, " ", "-"))

	url := fmt.Sprintf("https://royaleapi.com/decks/stats/%s", deckString)

	link.WriteString(fmt.Sprintf("🔗 RoyaleAPI Link:\n   %s\n\n", url))

	// Also provide DeckShop.pro link format
	// Format: https://www.deckshop.pro/check/?deck=Card1-Card2-Card3-Card4-Card5-Card6-Card7-Card8
	deckshopString := strings.Join(result.Deck, "-")
	deckshopString = strings.ReplaceAll(deckshopString, " ", "-")
	deckshopURL := fmt.Sprintf("https://www.deckshop.pro/check/?deck=%s", deckshopString)

	link.WriteString(fmt.Sprintf("🔗 DeckShop.pro Link:\n   %s\n\n", deckshopURL))

	return link.String()
}

// formatFooter creates the footer with performance metrics and version info
func formatFooter(result *EvaluationResult) string {
	var footer strings.Builder

	footer.WriteString("═══════════════════════════════════════════════════════════════════════\n")

	// Version and meta info
	footer.WriteString("cr-api deck evaluation engine • v1.0.0\n")
	footer.WriteString("Powered by Clash Royale API data and synergy analysis\n")

	footer.WriteString("═══════════════════════════════════════════════════════════════════════\n")

	return footer.String()
}
