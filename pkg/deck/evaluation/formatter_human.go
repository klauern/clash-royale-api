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
	output.WriteString(formatAlternativeSuggestions(result))
	output.WriteString(formatMissingCards(result))
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
		{"Playability", result.Playability, "🃏"},
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
		maxPairs := min(len(result.SynergyMatrix.Pairs), 5)

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

// formatCopyDeckLink generates a shareable deck link
func formatCopyDeckLink(result *EvaluationResult) string {
	var link strings.Builder

	link.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	link.WriteString("                          SHARE DECK\n")
	link.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	// Generate official Clash Royale copy-deck link
	deckLink := GenerateDeckLink(result.Deck)
	if deckLink.Valid {
		link.WriteString(fmt.Sprintf("📋 Copy Deck to Game:\n   %s\n\n", deckLink.URL))
	} else {
		link.WriteString(fmt.Sprintf("⚠️  Copy Deck Link: Unable to generate (%s)\n\n", deckLink.Error))
	}

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

// formatAlternativeSuggestions formats alternative deck suggestions
func formatAlternativeSuggestions(result *EvaluationResult) string {
	if result.AlternativeSuggestions == nil || len(result.AlternativeSuggestions.Suggestions) == 0 {
		return ""
	}

	var alts strings.Builder

	alts.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	alts.WriteString("                    ALTERNATIVE DECK SUGGESTIONS\n")
	alts.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	alts.WriteString(fmt.Sprintf("Current Deck Score: %.1f/10\n\n", result.AlternativeSuggestions.OriginalScore))

	for i, suggestion := range result.AlternativeSuggestions.Suggestions {
		if i >= 5 {
			break // Show max 5 suggestions
		}

		alts.WriteString(fmt.Sprintf("💡 Alternative #%d: %s\n", i+1, suggestion.Impact))
		alts.WriteString(fmt.Sprintf("   Replace: %s → %s\n", suggestion.OriginalCard, suggestion.ReplacementCard))
		alts.WriteString(fmt.Sprintf("   Score:   %.1f → %.1f (+%.1f)\n", suggestion.OriginalScore, suggestion.NewScore, suggestion.ScoreDelta))
		alts.WriteString(fmt.Sprintf("   Why:     %s\n\n", suggestion.Rationale))
	}

	return alts.String()
}

// formatMissingCards formats missing cards analysis
func formatMissingCards(result *EvaluationResult) string {
	if result.MissingCardsAnalysis == nil {
		return ""
	}

	analysis := result.MissingCardsAnalysis

	// If deck is playable, show a success message
	if analysis.IsPlayable {
		var playable strings.Builder
		playable.WriteString("═══════════════════════════════════════════════════════════════════════\n")
		playable.WriteString("                       CARD AVAILABILITY\n")
		playable.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")
		playable.WriteString("✅ All cards in this deck are available in your collection!\n\n")
		return playable.String()
	}

	var missing strings.Builder

	missing.WriteString("═══════════════════════════════════════════════════════════════════════\n")
	missing.WriteString("                       MISSING CARDS ANALYSIS\n")
	missing.WriteString("═══════════════════════════════════════════════════════════════════════\n\n")

	missing.WriteString(fmt.Sprintf("Deck Status: %d/%d cards available\n", analysis.AvailableCount, len(analysis.Deck)))

	// Count locked vs unlocked missing cards
	lockedCount := 0
	for _, card := range analysis.MissingCards {
		if card.IsLocked {
			lockedCount++
		}
	}

	// Add warning summary
	if lockedCount > 0 {
		missing.WriteString(fmt.Sprintf("⚠️  Warning: %d card(s) locked by arena restrictions\n", lockedCount))
	}
	if analysis.MissingCount-lockedCount > 0 {
		missing.WriteString(fmt.Sprintf("   %d card(s) unlocked but not in collection\n", analysis.MissingCount-lockedCount))
	}
	missing.WriteString("\n")

	for i, card := range analysis.MissingCards {
		missing.WriteString(fmt.Sprintf("❌ %d. %s (%s)\n", i+1, card.Name, card.Rarity))
		missing.WriteString(fmt.Sprintf("   Unlocks: %s (Arena %d)\n", card.UnlockArenaName, card.UnlockArena))

		if card.IsLocked {
			missing.WriteString(fmt.Sprintf("   Status:  🔒 LOCKED - Progress to Arena %d to unlock\n", card.UnlockArena))
		} else {
			missing.WriteString("   Status:  ✓ Unlocked - Available in chests and shop\n")
		}

		if len(card.AlternativeCards) > 0 {
			missing.WriteString(fmt.Sprintf("   Alternatives: %s\n", strings.Join(card.AlternativeCards, ", ")))
		} else {
			missing.WriteString("   Alternatives: None found in your collection\n")
		}

		missing.WriteString("\n")
	}

	return missing.String()
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
