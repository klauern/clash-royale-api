package evaluation

import (
	"fmt"
	"strings"
)

// FormatDetailed formats an EvaluationResult with comprehensive explanations
// Includes calculation methodology, verbose descriptions, and debug information
func FormatDetailed(result *EvaluationResult) string {
	var output strings.Builder

	// Header with full metadata
	output.WriteString("╔═══════════════════════════════════════════════════════════════════════╗\n")
	output.WriteString("║               DETAILED DECK EVALUATION REPORT                         ║\n")
	output.WriteString("║                    (Developer Mode)                                   ║\n")
	output.WriteString("╚═══════════════════════════════════════════════════════════════════════╝\n\n")

	// Deck overview with detailed stats
	output.WriteString(formatDetailedHeader(result))

	// Category scores with calculation methodology
	output.WriteString(formatDetailedCategoryScores(result))

	// Detailed analysis with verbose explanations
	output.WriteString(formatDetailedAnalysisSections(result))

	// Synergy matrix with detailed pair explanations
	output.WriteString(formatDetailedSynergyMatrix(result))

	// Missing cards analysis with unlock requirements
	output.WriteString(formatDetailedMissingCards(result))

	// Counter analysis with comprehensive breakdowns
	output.WriteString(formatDetailedCounterAnalysis(result))

	// Recommendations with detailed reasoning
	output.WriteString(formatDetailedRecommendations(result))

	// Debug information
	output.WriteString(formatDebugInformation(result))

	// Footer
	output.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	output.WriteString("cr-api deck evaluation engine • v1.0.0 • Detailed Output Mode\n")
	output.WriteString("═══════════════════════════════════════════════════════════════════════\n")

	return output.String()
}

// formatDetailedHeader creates detailed deck overview
func formatDetailedHeader(result *EvaluationResult) string {
	var header strings.Builder

	header.WriteString("🃏 DECK COMPOSITION\n")
	header.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")
	header.WriteString("Cards: " + strings.Join(result.Deck, " • ") + "\n")
	header.WriteString(fmt.Sprintf("Card Count: %d\n", len(result.Deck)))
	header.WriteString(fmt.Sprintf("Average Elixir Cost: %.2f\n\n", result.AvgElixir))

	header.WriteString("Archetype Detection:\n")
	header.WriteString(fmt.Sprintf("  Detected: %s\n", strings.Title(string(result.DetectedArchetype))))
	header.WriteString(fmt.Sprintf("  Confidence: %.2f%% (%.2f/1.0)\n", result.ArchetypeConfidence*100, result.ArchetypeConfidence))
	header.WriteString("  Method: Pattern matching with weighted signatures\n\n")

	header.WriteString("Overall Evaluation:\n")
	header.WriteString(fmt.Sprintf("  Overall Score: %.2f/10.0\n", result.OverallScore))
	header.WriteString(fmt.Sprintf("  Rating: %s\n", result.OverallRating))
	header.WriteString("  Calculation: Weighted average of 5 category scores\n")
	header.WriteString("  Weights: Attack(25%), Defense(25%), Synergy(20%), Versatility(20%), F2P(10%)\n\n")

	return header.String()
}

// formatDetailedCategoryScores formats category scores with methodology
func formatDetailedCategoryScores(result *EvaluationResult) string {
	var scores strings.Builder

	scores.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	scores.WriteString("                     CATEGORY SCORING BREAKDOWN\n")
	scores.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	categories := []struct {
		Name     string
		Score    CategoryScore
		Icon     string
		Method   string
		Criteria string
	}{
		{
			Name:     "Attack",
			Score:    result.Attack,
			Icon:     "⚔️",
			Method:   "Win condition presence, damage potential, tower threat",
			Criteria: "Win conditions (40%), Spell damage (30%), Support troops (30%)",
		},
		{
			Name:     "Defense",
			Score:    result.Defense,
			Icon:     "🛡️",
			Method:   "Anti-air coverage, building presence, defensive troops",
			Criteria: "Buildings (40%), Anti-air (30%), Ground defense (30%)",
		},
		{
			Name:     "Synergy",
			Score:    result.Synergy,
			Icon:     "🔗",
			Method:   "Card pair synergies, combo potential, archetype coherence",
			Criteria: "Synergy pairs (50%), Card coverage (30%), Role diversity (20%)",
		},
		{
			Name:     "Versatility",
			Score:    result.Versatility,
			Icon:     "🎭",
			Method:   "Elixir distribution, card roles, adaptability",
			Criteria: "Role variety (40%), Elixir range (30%), Multi-use cards (30%)",
		},
		{
			Name:     "F2P Friendly",
			Score:    result.F2PFriendly,
			Icon:     "💰",
			Method:   "Rarity distribution, upgrade costs, accessibility",
			Criteria: "Common/Rare ratio (50%), Legendary count (30%), Epic count (20%)",
		},
		{
			Name:     "Playability",
			Score:    result.Playability,
			Icon:     "🃏",
			Method:   "Card availability based on player collection and arena",
			Criteria: "Card ownership (60%), Arena unlock status (40%)",
		},
	}

	for _, cat := range categories {
		scores.WriteString(fmt.Sprintf("%s %s SCORE\n", cat.Icon, strings.ToUpper(cat.Name)))
		scores.WriteString("─────────────────────────────────────────────────────────────────────\n")
		scores.WriteString(fmt.Sprintf("Numeric Score: %.2f/10.0\n", cat.Score.Score))
		scores.WriteString(fmt.Sprintf("Star Rating: %s (%d/3 stars)\n", formatStars(cat.Score.Stars), cat.Score.Stars))
		scores.WriteString(fmt.Sprintf("Qualitative Rating: %s\n", cat.Score.Rating))
		scores.WriteString(fmt.Sprintf("Assessment: %s\n\n", cat.Score.Assessment))
		scores.WriteString("Scoring Methodology:\n")
		scores.WriteString(fmt.Sprintf("  Method: %s\n", cat.Method))
		scores.WriteString(fmt.Sprintf("  Criteria: %s\n\n", cat.Criteria))
	}

	return scores.String()
}

// formatDetailedAnalysisSections formats analysis sections with verbose descriptions
func formatDetailedAnalysisSections(result *EvaluationResult) string {
	var analysis strings.Builder

	analysis.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	analysis.WriteString("                    COMPREHENSIVE DECK ANALYSIS\n")
	analysis.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	sections := []AnalysisSection{
		result.DefenseAnalysis,
		result.AttackAnalysis,
		result.BaitAnalysis,
		result.CycleAnalysis,
		result.LadderAnalysis,
	}

	// Add evolution analysis if present
	if result.EvolutionAnalysis.Title != "" {
		sections = append(sections, result.EvolutionAnalysis)
	}

	for i, section := range sections {
		analysis.WriteString(fmt.Sprintf("%d. %s\n", i+1, strings.ToUpper(section.Title)))
		analysis.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")
		analysis.WriteString(fmt.Sprintf("Score: %.2f/10.0 | Rating: %s\n\n", section.Score, section.Rating))

		analysis.WriteString("Executive Summary:\n")
		analysis.WriteString(fmt.Sprintf("%s\n\n", section.Summary))

		if len(section.Details) > 0 {
			analysis.WriteString("Detailed Observations:\n")
			for j, detail := range section.Details {
				analysis.WriteString(fmt.Sprintf("  %d. %s\n", j+1, detail))
			}
			analysis.WriteString("\n")
		}

		// Add calculation notes
		analysis.WriteString("Scoring Notes:\n")
		analysis.WriteString("  This score is calculated based on card composition, role distribution,\n")
		analysis.WriteString("  and archetype-specific requirements. Higher scores indicate better\n")
		analysis.WriteString("  performance in this specific aspect of deck building.\n\n")
	}

	return analysis.String()
}

// formatDetailedSynergyMatrix formats synergy matrix with detailed explanations
func formatDetailedSynergyMatrix(result *EvaluationResult) string {
	var matrix strings.Builder

	if result.SynergyMatrix.PairCount == 0 {
		return ""
	}

	matrix.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	matrix.WriteString("                    SYNERGY MATRIX ANALYSIS\n")
	matrix.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	matrix.WriteString("Overall Synergy Metrics:\n")
	matrix.WriteString(fmt.Sprintf("  Total Synergy Score: %.2f/10.0\n", result.SynergyMatrix.TotalScore))
	matrix.WriteString(fmt.Sprintf("  Average Pair Synergy: %.2f%% (%.2f/1.0)\n",
		result.SynergyMatrix.AverageSynergy*100, result.SynergyMatrix.AverageSynergy))
	matrix.WriteString(fmt.Sprintf("  Synergy Pairs Found: %d out of %d possible pairs\n",
		result.SynergyMatrix.PairCount, result.SynergyMatrix.MaxPossiblePairs))
	matrix.WriteString(fmt.Sprintf("  Card Coverage: %.2f%% of cards have synergies\n\n",
		result.SynergyMatrix.SynergyCoverage))

	matrix.WriteString("Methodology:\n")
	matrix.WriteString("  Synergies are identified through predefined combo patterns that enhance\n")
	matrix.WriteString("  offensive potential, defensive capabilities, or strategic versatility.\n")
	matrix.WriteString("  Scores range from 0.0 (no synergy) to 1.0 (perfect synergy).\n\n")

	if len(result.SynergyMatrix.Pairs) > 0 {
		matrix.WriteString("Card-by-Card Synergy Breakdown:\n")
		matrix.WriteString("─────────────────────────────────────────────────────────────────────\n\n")

		for i, pair := range result.SynergyMatrix.Pairs {
			strengthLabel := formatSynergyStrength(pair.Score)
			matrix.WriteString(fmt.Sprintf("Pair #%d: %s ↔ %s\n", i+1, pair.Card1, pair.Card2))
			matrix.WriteString(fmt.Sprintf("  Synergy Strength: %s (%.2f%%, score: %.2f/1.0)\n",
				strengthLabel, pair.Score*100, pair.Score))
			matrix.WriteString(fmt.Sprintf("  Why This Works: %s\n", pair.Description))
			matrix.WriteString("  Impact: This combo enhances deck effectiveness through complementary roles\n")
			matrix.WriteString("          and timing opportunities.\n\n")
		}
	}

	return matrix.String()
}

// formatDetailedMissingCards formats missing cards analysis with detailed unlock requirements
func formatDetailedMissingCards(result *EvaluationResult) string {
	if result.MissingCardsAnalysis == nil {
		return ""
	}

	analysis := result.MissingCardsAnalysis
	var missing strings.Builder

	missing.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	missing.WriteString("                   CARD AVAILABILITY & UNLOCK REQUIREMENTS\n")
	missing.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	// If deck is playable, show success and return
	if analysis.IsPlayable {
		missing.WriteString("✅ DECK STATUS: FULLY PLAYABLE\n")
		missing.WriteString("─────────────────────────────────────────────────────────────────────\n\n")
		missing.WriteString("All 8 cards in this deck are available in your collection!\n")
		missing.WriteString("You can build and play this deck immediately.\n\n")
		return missing.String()
	}

	// Deck not playable - show detailed analysis
	missing.WriteString("📊 DECK AVAILABILITY SUMMARY\n")
	missing.WriteString("─────────────────────────────────────────────────────────────────────\n\n")
	missing.WriteString(fmt.Sprintf("  Cards Available:  %d / %d (%.0f%%)\n",
		analysis.AvailableCount, len(analysis.Deck),
		float64(analysis.AvailableCount)/float64(len(analysis.Deck))*100))
	missing.WriteString(fmt.Sprintf("  Cards Missing:    %d / %d\n\n", analysis.MissingCount, len(analysis.Deck)))

	// Count locked vs unlocked
	lockedCount := 0
	unlockedButMissingCount := 0
	for _, card := range analysis.MissingCards {
		if card.IsLocked {
			lockedCount++
		} else {
			unlockedButMissingCount++
		}
	}

	missing.WriteString("Missing Card Breakdown:\n")
	if lockedCount > 0 {
		missing.WriteString(fmt.Sprintf("  🔒 Arena Locked:  %d card(s) - Requires arena progression\n", lockedCount))
	}
	if unlockedButMissingCount > 0 {
		missing.WriteString(fmt.Sprintf("  ✓  Unlocked:       %d card(s) - Obtainable from chests/shop\n", unlockedButMissingCount))
	}
	missing.WriteString("\n")

	// Detailed card-by-card breakdown
	missing.WriteString("MISSING CARDS - DETAILED ANALYSIS\n")
	missing.WriteString("─────────────────────────────────────────────────────────────────────\n\n")

	for i, card := range analysis.MissingCards {
		missing.WriteString(fmt.Sprintf("CARD #%d: %s\n", i+1, strings.ToUpper(card.Name)))
		missing.WriteString("─────────────────────────────────────────────────────────────────────\n")

		// Availability status
		if card.IsLocked {
			missing.WriteString(fmt.Sprintf("Availability Status: 🔒 LOCKED - Requires Arena %d (%s)\n",
				card.UnlockArena, card.UnlockArenaName))
		} else {
			missing.WriteString(fmt.Sprintf("Availability Status: ✓ UNLOCKED - Available at Arena %d (%s)\n",
				card.UnlockArena, card.UnlockArenaName))
		}

		// Rarity
		missing.WriteString(fmt.Sprintf("Rarity:              %s\n", card.Rarity))

		// Unlock requirements
		missing.WriteString("\nUnlock Requirements:\n")
		if card.IsLocked {
			missing.WriteString(fmt.Sprintf("  • Arena Progression: Progress to Arena %d (%s) to unlock\n",
				card.UnlockArena, card.UnlockArenaName))
			missing.WriteString("  • Once unlocked, card will appear in chests and shop\n")
		} else {
			missing.WriteString("  • Card is already unlocked in your current arena\n")
			missing.WriteString("  • Can be obtained from chests, shop, or trades\n")
		}

		// Suggested alternatives
		if len(card.AlternativeCards) > 0 {
			missing.WriteString("\nSuggested Alternatives (in your collection):\n")
			for j, alt := range card.AlternativeCards {
				missing.WriteString(fmt.Sprintf("  %d. %s - Similar role/elixir cost\n", j+1, alt))
			}
			missing.WriteString("\nNote: These alternatives can replace the missing card while maintaining\n")
			missing.WriteString("      similar deck strategy and elixir balance.\n")
		} else {
			missing.WriteString("\nSuggested Alternatives:\n")
			missing.WriteString("  • No similar cards found in your collection\n")
			missing.WriteString("  • Consider waiting to unlock this card before using this deck\n")
		}

		missing.WriteString("\n")
	}

	// Impact on deck score
	if lockedCount > 0 || unlockedButMissingCount > 0 {
		missing.WriteString("IMPACT ON DECK EVALUATION\n")
		missing.WriteString("─────────────────────────────────────────────────────────────────────\n\n")
		penalty := float64(lockedCount)*2.0 + float64(unlockedButMissingCount)*1.0
		missing.WriteString(fmt.Sprintf("Score Penalty Applied: -%.1f points\n", penalty))
		missing.WriteString("  • Locked cards: -2.0 points each (arena progression required)\n")
		missing.WriteString("  • Unlocked missing cards: -1.0 points each (obtainable immediately)\n\n")
		missing.WriteString("Note: The overall deck score has been reduced to reflect card\n")
		missing.WriteString("      availability limitations. Score will improve as you acquire cards.\n\n")
	}

	return missing.String()
}

// formatDetailedCounterAnalysis formats counter analysis with comprehensive breakdowns
func formatDetailedCounterAnalysis(result *EvaluationResult) string {
	var counter strings.Builder

	counter.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	counter.WriteString("                   MATCHUP ANALYSIS: STRENGTHS & WEAKNESSES\n")
	counter.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	// Strengths
	counter.WriteString("✅ DECK STRENGTHS\n")
	counter.WriteString("─────────────────────────────────────────────────────────────────────\n")
	strengths := deriveStrengths(result)
	for i, strength := range strengths {
		counter.WriteString(fmt.Sprintf("%d. %s\n", i+1, strength))
		counter.WriteString("   Impact: High-scoring categories indicate your deck excels in these areas,\n")
		counter.WriteString("           giving you advantages in specific matchup types.\n\n")
	}

	// Weaknesses
	counter.WriteString("⚠️  DECK WEAKNESSES\n")
	counter.WriteString("─────────────────────────────────────────────────────────────────────\n")
	weaknesses := deriveWeaknesses(result)
	for i, weakness := range weaknesses {
		counter.WriteString(fmt.Sprintf("%d. %s\n", i+1, weakness))
		counter.WriteString("   Impact: Low-scoring categories represent vulnerabilities that skilled\n")
		counter.WriteString("           opponents can exploit. Consider addressing these gaps.\n\n")
	}

	counter.WriteString("Analysis Note:\n")
	counter.WriteString("Strengths (score >= 7.0) and weaknesses (score < 5.0) are identified\n")
	counter.WriteString("automatically based on category scores. Balanced decks (5.0-6.9) have\n")
	counter.WriteString("moderate performance across categories.\n\n")

	return counter.String()
}

// formatDetailedRecommendations formats recommendations with detailed reasoning
func formatDetailedRecommendations(result *EvaluationResult) string {
	var recs strings.Builder

	recs.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	recs.WriteString("                    IMPROVEMENT RECOMMENDATIONS\n")
	recs.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	recommendations := generateRecommendations(result)

	if len(recommendations) == 0 {
		recs.WriteString("✨ CONGRATULATIONS!\n")
		recs.WriteString("─────────────────────────────────────────────────────────────────────\n")
		recs.WriteString("This deck shows no critical weaknesses. All category scores are within\n")
		recs.WriteString("acceptable ranges (>= 5.0/10). Continue refining your strategy and card\n")
		recs.WriteString("levels to maximize competitive potential.\n\n")
		return recs.String()
	}

	recs.WriteString("Priority-Ordered Action Items:\n")
	recs.WriteString("─────────────────────────────────────────────────────────────────────\n\n")

	for i, rec := range recommendations {
		priorityIcon := "🔴"
		priorityLabel := "HIGH PRIORITY"
		if rec.Priority == "Medium" {
			priorityIcon = "🟡"
			priorityLabel = "MEDIUM PRIORITY"
		} else if rec.Priority == "Low" {
			priorityIcon = "🟢"
			priorityLabel = "LOW PRIORITY"
		}

		recs.WriteString(fmt.Sprintf("%s Recommendation #%d: %s [%s]\n", priorityIcon, i+1, rec.Title, priorityLabel))
		recs.WriteString("─────────────────────────────────────────────────────────────────────\n")
		recs.WriteString(fmt.Sprintf("Issue: %s\n\n", rec.Description))
		recs.WriteString("Why This Matters:\n")

		switch rec.Priority {
		case "High":
			recs.WriteString("  Critical weakness that significantly impacts deck performance. Address\n")
			recs.WriteString("  this immediately to improve competitive viability and matchup success.\n")
		case "Medium":
			recs.WriteString("  Moderate weakness that may cause difficulties in certain matchups.\n")
			recs.WriteString("  Address after resolving high-priority issues.\n")
		case "Low":
			recs.WriteString("  Minor optimization opportunity. Consider addressing if resources permit,\n")
			recs.WriteString("  but not critical to deck functionality.\n")
		}

		recs.WriteString("\n")
	}

	return recs.String()
}

// formatDebugInformation adds debug/developer information
func formatDebugInformation(result *EvaluationResult) string {
	var debug strings.Builder

	debug.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	debug.WriteString("                      DEBUG INFORMATION\n")
	debug.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	debug.WriteString("Evaluation Metadata:\n")
	debug.WriteString(fmt.Sprintf("  Engine Version: 1.0.0\n"))
	debug.WriteString(fmt.Sprintf("  Total Cards Evaluated: %d\n", len(result.Deck)))
	debug.WriteString(fmt.Sprintf("  Synergy Pairs Detected: %d/%d (%.1f%%)\n",
		result.SynergyMatrix.PairCount,
		result.SynergyMatrix.MaxPossiblePairs,
		float64(result.SynergyMatrix.PairCount)/float64(result.SynergyMatrix.MaxPossiblePairs)*100))

	debug.WriteString("\nScore Calculation Details:\n")
	debug.WriteString(fmt.Sprintf("  Attack Score: %.2f (Weight: 0.25)\n", result.Attack.Score))
	debug.WriteString(fmt.Sprintf("  Defense Score: %.2f (Weight: 0.25)\n", result.Defense.Score))
	debug.WriteString(fmt.Sprintf("  Synergy Score: %.2f (Weight: 0.20)\n", result.Synergy.Score))
	debug.WriteString(fmt.Sprintf("  Versatility Score: %.2f (Weight: 0.20)\n", result.Versatility.Score))
	debug.WriteString(fmt.Sprintf("  F2P Friendly Score: %.2f (Weight: 0.10)\n", result.F2PFriendly.Score))
	debug.WriteString(fmt.Sprintf("  Weighted Sum: %.2f\n", result.OverallScore))

	debug.WriteString("\nRating Scale Reference:\n")
	debug.WriteString("  9.0-10.0: Godly!\n")
	debug.WriteString("  8.0-8.9:  Amazing\n")
	debug.WriteString("  7.0-7.9:  Great\n")
	debug.WriteString("  6.0-6.9:  Good\n")
	debug.WriteString("  5.0-5.9:  Decent\n")
	debug.WriteString("  4.0-4.9:  Mediocre\n")
	debug.WriteString("  3.0-3.9:  Poor\n")
	debug.WriteString("  2.0-2.9:  Bad\n")
	debug.WriteString("  1.0-1.9:  Terrible\n")
	debug.WriteString("  0.0-0.9:  Awful\n\n")

	return debug.String()
}
