//nolint:funlen,goconst,gocritic,gocognit,gocyclo // Existing domain logic is unchanged by the structural split.
package evaluation

import (
	"fmt"
	"sort"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// isLevelIndependent reports whether a card remains useful when underleveled.
func isLevelIndependent(card deck.CardCandidate) bool {
	// Small spells (utility-based)
	smallSpells := map[string]bool{
		"Log": true, "Zap": true, "Arrows": true, "Snowball": true,
		"Barbarian Barrel": true, "Giant Snowball": true,
	}

	// Cycle cards (cheap utility)
	cycleCards := map[string]bool{
		"Skeletons": true, "Ice Spirit": true, "Ice Golem": true,
		"Heal Spirit": true, "Electro Spirit": true, "Fire Spirit": true,
	}

	// Defensive buildings (utility)
	buildings := map[string]bool{
		"Tesla": true, "Cannon": true, "Bomb Tower": true,
	}

	// Reset cards (utility effect)
	resetCards := map[string]bool{
		"Electro Wizard": true, "Electro Spirit": true,
	}

	return smallSpells[card.Name] || cycleCards[card.Name] ||
		buildings[card.Name] || resetCards[card.Name]
}

// calculateLadderScore combines F2P factors with level-independence (0-10)
func calculateLadderScore(rarityScore, levelIndepScore, upgradeProgress float64) float64 {
	// Rarity distribution (40%)
	// Level-independence (30%)
	// Upgrade progress (20%)
	// Upgrade clarity bonus (10%) - implicit in other factors

	score := (rarityScore * 0.4) + (levelIndepScore * 0.3) + (upgradeProgress * 0.2)

	return score
}

// calculatePlayerLevelMetrics calculates level-based metrics from player context
func calculatePlayerLevelMetrics(deckCards []deck.CardCandidate, playerContext *PlayerContext) (avgCurrentLevel, avgLevelGap float64, maxLevelGap, cardsWithLevels int) {
	totalCurrentLevel := 0
	totalGap := 0

	for _, card := range deckCards {
		if info, exists := playerContext.Collection[card.Name]; exists {
			totalCurrentLevel += info.Level
			gap := info.MaxLevel - info.Level
			totalGap += gap
			cardsWithLevels++

			if gap > maxLevelGap {
				maxLevelGap = gap
			}
		}
	}

	if cardsWithLevels > 0 {
		avgCurrentLevel = float64(totalCurrentLevel) / float64(cardsWithLevels)
		avgLevelGap = float64(totalGap) / float64(cardsWithLevels)
	}

	return avgCurrentLevel, avgLevelGap, maxLevelGap, cardsWithLevels
}

// generateViabilityRating assigns competitive viability rating from score
func generateViabilityRating(competitiveViability float64) string {
	switch {
	case competitiveViability >= 9.0:
		return "Tournament ready"
	case competitiveViability >= 7.0:
		return "Ladder competitive"
	case competitiveViability >= 5.0:
		return "Playable but underleveled"
	case competitiveViability >= 3.0:
		return "Significant disadvantage"
	default:
		return "Not competitive"
	}
}

// collectRarityBreakdown counts cards by rarity
func collectRarityBreakdown(deckCards []deck.CardCandidate) map[string]int {
	rarityCount := map[string]int{
		"Common":    0,
		"Rare":      0,
		"Epic":      0,
		"Legendary": 0,
		"Champion":  0,
	}

	for _, card := range deckCards {
		rarityCount[card.Rarity]++
	}

	return rarityCount
}

// calculateLadderViabilityScore estimates ladder viability from level gaps.
// Uses a steeper penalty curve to reflect competitive breakpoints.
func calculateLadderViabilityScore(avgLevelGap float64, maxLevelGap int) float64 {
	score := 10.0 - avgLevelGap

	if maxLevelGap >= 3 {
		score -= 0.5
	}
	if maxLevelGap >= 5 {
		score -= 1.0
	}
	if maxLevelGap >= 7 {
		score -= 1.5
	}

	if score < 0 {
		score = 0
	}
	if score > 10 {
		score = 10
	}

	return score
}

func clampScoreToTen(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 10 {
		return 10
	}
	return score
}

func calculateDeckLevelRatio(deckCards []deck.CardCandidate, playerContext *PlayerContext) float64 {
	if playerContext == nil || len(deckCards) == 0 {
		return 1.0
	}

	totalRatio := 0.0
	seen := 0
	for _, card := range deckCards {
		info, exists := playerContext.Collection[card.Name]
		if !exists || info.MaxLevel <= 0 {
			continue
		}
		levelRatio := float64(info.Level) / float64(info.MaxLevel)
		if levelRatio < 0 {
			levelRatio = 0
		}
		if levelRatio > 1 {
			levelRatio = 1
		}
		totalRatio += levelRatio
		seen++
	}

	if seen == 0 {
		return 1.0
	}

	return totalRatio / float64(seen)
}

func calculateLevelNormalizationFactor(levelRatio float64) float64 {
	if levelRatio < 0 {
		levelRatio = 0
	}
	if levelRatio > 1 {
		levelRatio = 1
	}

	// Keep the prototype conservative to avoid destabilizing existing score scales.
	return 0.90 + (levelRatio * 0.20)
}

// upgradePriority represents an upgrade recommendation
type upgradePriority struct {
	cardName     string
	currentLevel int
	maxLevel     int
	gap          int
	tier         int
	reason       string
}

// isEvolutionPriority checks if card needs evolution upgrade
func isEvolutionPriority(info CardLevelInfo) (bool, string) {
	if info.MaxEvolutionLevel > 0 && info.EvolutionLevel == 0 {
		return true, "evolvable"
	}
	if info.MaxEvolutionLevel > 0 && info.EvolutionLevel < info.MaxEvolutionLevel {
		return true, "evolution upgrade"
	}
	return false, ""
}

// isSpellCard checks if card is a spell (big or small)
func isSpellCard(card deck.CardCandidate) bool {
	return card.Role != nil && (*card.Role == deck.RoleSpellBig || *card.Role == deck.RoleSpellSmall)
}

// determineUpgradeTierAndReason assigns priority tier and reason based on card characteristics
func determineUpgradeTierAndReason(card deck.CardCandidate, info CardLevelInfo) (tier int, reason string) {
	if isEvolution, evolutionReason := isEvolutionPriority(info); isEvolution {
		return 0, evolutionReason
	}
	if card.Role != nil && *card.Role == deck.RoleWinCondition {
		return 1, "win condition"
	}
	if isSpellCard(card) {
		return 2, "spell breakpoints"
	}
	if card.Stats != nil && card.Stats.DamagePerSecond > 150 {
		return 3, "tank killer"
	}
	return 4, "support"
}

// calculateUpgradePriorities returns sorted upgrade recommendations
func calculateUpgradePriorities(deckCards []deck.CardCandidate, playerContext *PlayerContext) []upgradePriority {
	priorities := []upgradePriority{}

	for _, card := range deckCards {
		info, exists := playerContext.Collection[card.Name]
		if !exists {
			continue // Card not owned
		}

		gap := info.MaxLevel - info.Level
		if gap == 0 {
			continue // Already maxed
		}

		tier, reason := determineUpgradeTierAndReason(card, info)

		priority := upgradePriority{
			cardName:     card.Name,
			currentLevel: info.Level,
			maxLevel:     info.MaxLevel,
			gap:          gap,
			tier:         tier,
			reason:       reason,
		}

		priorities = append(priorities, priority)
	}

	// Sort by tier (ascending), then gap (descending)
	sort.Slice(priorities, func(i, j int) bool {
		if priorities[i].tier != priorities[j].tier {
			return priorities[i].tier < priorities[j].tier
		}
		return priorities[i].gap > priorities[j].gap
	})

	return priorities
}

// BuildLadderAnalysis creates detailed ladder analysis
// If playerContext is provided, uses actual card levels for personalized recommendations
// If playerContext is nil, falls back to generic rarity-based analysis
func BuildLadderAnalysis(deckCards []deck.CardCandidate, playerContext *PlayerContext) AnalysisSection {
	// Get F2P score for rarity assessment
	f2pScore := ScoreF2P(deckCards)

	// Count rarity breakdown using helper
	rarityCount := collectRarityBreakdown(deckCards)

	// Calculate level independence
	levelIndepCards := []deck.CardCandidate{}
	totalLevelRatio := 0.0

	for _, card := range deckCards {
		if isLevelIndependent(card) {
			levelIndepCards = append(levelIndepCards, card)
		}
		totalLevelRatio += card.LevelRatio()
	}

	// Check if we have player context for level-based analysis
	hasPlayerContext := playerContext != nil
	var avgCurrentLevel, avgLevelGap float64
	var maxLevelGap, cardsWithLevels int
	var upgradePriorities []upgradePriority

	if hasPlayerContext {
		// Calculate actual level metrics using helper
		avgCurrentLevel, avgLevelGap, maxLevelGap, cardsWithLevels = calculatePlayerLevelMetrics(deckCards, playerContext)
		upgradePriorities = calculateUpgradePriorities(deckCards, playerContext)
	}

	// Calculate competitive viability (level-based)
	var competitiveViability float64
	var viabilityRating string

	if hasPlayerContext {
		if cardsWithLevels == 0 {
			competitiveViability = 0
			viabilityRating = "Not competitive"
		} else {
			competitiveViability = calculateLadderViabilityScore(avgLevelGap, maxLevelGap)
			viabilityRating = generateViabilityRating(competitiveViability)
		}
	}

	// Calculate ladder score
	var score float64
	var rating Rating

	if hasPlayerContext {
		// Use competitive viability as primary score when player context available
		score = competitiveViability
		rating = ScoreToRating(score)
	} else {
		// Fall back to F2P-based scoring
		levelIndepScore := float64(len(levelIndepCards)) / float64(len(deckCards)) * 10.0
		avgProgress := totalLevelRatio / float64(len(deckCards))
		upgradeProgressScore := avgProgress * 10.0
		score = calculateLadderScore(f2pScore.Score, levelIndepScore, upgradeProgressScore)
		rating = ScoreToRating(score)
	}

	// Build details array
	details := []string{}

	// Rarity breakdown (always shown)
	details = append(details, fmt.Sprintf("Rarity breakdown: %d Commons, %d Rares, %d Epics, %d Legendaries, %d Champions",
		rarityCount["Common"], rarityCount["Rare"], rarityCount["Epic"],
		rarityCount["Legendary"], rarityCount["Champion"]))

	// Level-based analysis (only if player context available)
	if hasPlayerContext {
		// Average deck level
		maxAvgLevel := avgCurrentLevel + avgLevelGap
		details = append(details, fmt.Sprintf("Average deck level: %.1f / %.0f (%.1f level gap)",
			avgCurrentLevel, maxAvgLevel, avgLevelGap))

		// Competitive viability
		details = append(details, fmt.Sprintf("Competitive viability: %s (%.1f/10)",
			viabilityRating, competitiveViability))

		// Upgrade priorities (top 3)
		if len(upgradePriorities) > 0 {
			for i := 0; i < len(upgradePriorities) && i < 3; i++ {
				p := upgradePriorities[i]
				details = append(details, fmt.Sprintf("Upgrade priority %d: %s (%d→%d, %s)",
					i+1, p.cardName, p.currentLevel, p.maxLevel, p.reason))
			}
		}

		// Cards ready for ladder (level 11+)
		readyCount := 0
		for _, card := range deckCards {
			if info, exists := playerContext.Collection[card.Name]; exists {
				if info.Level >= 11 {
					readyCount++
				}
			}
		}
		details = append(details, fmt.Sprintf("Cards ready for ladder: %d/%d (level 11+)", readyCount, len(deckCards)))
	}

	// Level-independent cards (always shown if present)
	if len(levelIndepCards) > 0 {
		details = append(details, fmt.Sprintf("Level-independent cards (%d): %s",
			len(levelIndepCards), buildCardList(levelIndepCards)))
	}

	// F2P assessment (always shown)
	f2pAssessment := "Difficult"
	if f2pScore.Score >= 8.0 {
		f2pAssessment = "Excellent"
	} else if f2pScore.Score >= 6.0 {
		f2pAssessment = "Good"
	}
	reason := ""
	if rarityCount["Legendary"] == 0 && rarityCount["Champion"] == 0 {
		reason = "no legendaries, common-heavy"
	} else if rarityCount["Legendary"]+rarityCount["Champion"] >= 3 {
		reason = "multiple legendaries/champions"
	} else {
		reason = "balanced rarity distribution"
	}
	details = append(details, fmt.Sprintf("F2P assessment: %s - %s", f2pAssessment, reason))

	// Gold efficiency (always shown)
	goldEfficiency := int(f2pScore.Score * 10)
	details = append(details, fmt.Sprintf("Gold efficiency: %d/100 - %s upgrade costs",
		goldEfficiency, f2pAssessment))

	// Generate summary
	summary := "Moderate F2P-friendliness"

	if hasPlayerContext {
		if competitiveViability >= 9.0 {
			summary = "Tournament-ready deck with maxed or near-maxed cards"
		} else if competitiveViability >= 7.0 {
			summary = fmt.Sprintf("Ladder competitive with %.1f average level gap", avgLevelGap)
		} else if competitiveViability >= 5.0 {
			summary = fmt.Sprintf("Playable but underleveled (%.1f level gap)", avgLevelGap)
		} else {
			summary = fmt.Sprintf("Significant level disadvantage (%.1f gap)", avgLevelGap)
		}
	} else {
		// Rarity-based summary (existing logic)
		if f2pScore.Score >= 8.0 {
			summary = "Excellent F2P deck with clear upgrade path"
		} else if f2pScore.Score < 5.0 {
			summary = "Expensive deck requiring significant investment"
		}
	}

	return AnalysisSection{
		Title:   "Ladder Analysis",
		Summary: summary,
		Details: details,
		Score:   score,
		Rating:  rating,
	}
}

// ============================================================================
// Phase 4.5: Evolution Analysis
// ============================================================================

// filterEvolvableCards returns cards that have evolution potential
