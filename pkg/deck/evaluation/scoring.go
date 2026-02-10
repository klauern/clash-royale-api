package evaluation

import (
	"fmt"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// attackMetrics holds extracted metrics from deck cards for attack scoring
type attackMetrics struct {
	winConditionCount   int
	winConditionQuality float64
	spellDamage         float64
	totalDamage         float64
	evolutionBonus      float64
}

// extractAttackMetrics collects attack-related metrics from deck cards
func extractAttackMetrics(deckCards []deck.CardCandidate) attackMetrics {
	metrics := attackMetrics{}

	for _, card := range deckCards {
		// Check if card is a win condition
		if card.Role != nil && *card.Role == deck.RoleWinCondition {
			metrics.winConditionCount++
			// Quality based on level and stats
			metrics.winConditionQuality += card.LevelRatio()
		}

		// Big spells contribute to attack potential
		if card.Role != nil && *card.Role == deck.RoleSpellBig {
			metrics.spellDamage += card.LevelRatio()
		}

		// Calculate overall damage potential using combat stats
		if card.Stats != nil {
			damageContribution := float64(card.Stats.DamagePerSecond) * card.LevelRatio()
			metrics.totalDamage += damageContribution
		}

		// Calculate evolution bonus
		if card.EvolutionLevel > 0 {
			// Evolution bonus: 15% per evolution level
			evoBonus := 0.15 * float64(card.EvolutionLevel)
			metrics.evolutionBonus += evoBonus
		}
	}

	return metrics
}

// scoreWinConditions calculates the win condition score component (0-10 scale)
func scoreWinConditions(count int, quality float64) float64 {
	if count >= 2 {
		return 10.0 // Multiple win conditions = excellent
	} else if count == 1 {
		return 7.5 + (quality * 2.5) // Single win condition with quality bonus
	}
	return 0.0 // No dedicated win condition - critical flaw
}

// scoreSpellDamage calculates the spell damage score component (0-10 scale)
func scoreSpellDamage(spellDamage float64) float64 {
	if spellDamage >= 2.0 {
		return 10.0 // Strong spell damage
	} else if spellDamage >= 1.0 {
		return 6.5 + (spellDamage * 2.0)
	} else if spellDamage > 0 {
		return 2.5 + (spellDamage * 4.0)
	}
	return 0.0 // No spells = critical flaw
}

// scoreDamagePotential calculates the overall damage score component (0-10 scale)
func scoreDamagePotential(avgDamage float64) float64 {
	if avgDamage >= 200 {
		return 10.0
	} else if avgDamage >= 100 {
		return 5.5 + ((avgDamage - 100) / 100 * 4.5)
	}
	return (avgDamage / 100) * 5.5
}

// ScoreAttack calculates the attack score for a deck (0-10 scale)
// Considers win conditions, damage potential, spell damage, and evolution bonuses
func ScoreAttack(deckCards []deck.CardCandidate) CategoryScore {
	if len(deckCards) == 0 {
		return CreateCategoryScore(0, "No cards in deck")
	}

	// Extract metrics from deck cards
	metrics := extractAttackMetrics(deckCards)

	// Score components (0-10 scale)
	// Win condition presence (40% of score)
	winConditionScore := scoreWinConditions(metrics.winConditionCount, metrics.winConditionQuality)

	// Spell damage potential (30% of score)
	spellScore := scoreSpellDamage(metrics.spellDamage)

	// Overall damage potential (30% of score)
	avgDamage := metrics.totalDamage / float64(len(deckCards))
	damageScore := scoreDamagePotential(avgDamage)

	// Combine components with weights
	score := (winConditionScore * 0.4) + (spellScore * 0.3) + (damageScore * 0.3)

	// Add evolution bonus (up to +1.5 points)
	score += metrics.evolutionBonus
	if score > 10.0 {
		score = 10.0
	}

	// Generate assessment text
	assessment := generateAttackAssessment(metrics.winConditionCount, metrics.spellDamage, score, metrics.evolutionBonus)

	return CreateCategoryScore(score, assessment)
}

// defenseMetrics holds extracted metrics from deck cards for defense scoring
type defenseMetrics struct {
	antiAirCount   int
	buildingCount  int
	supportCount   int
	defenseQuality float64
	evolutionBonus float64
}

// extractDefenseMetrics collects defense-related metrics from deck cards
//
//nolint:gocognit,gocyclo // Defensive metric extraction keeps explicit weighted branches.
func extractDefenseMetrics(deckCards []deck.CardCandidate) defenseMetrics {
	metrics := defenseMetrics{}

	for _, card := range deckCards {
		// Count buildings (defensive structures)
		if card.Role != nil && *card.Role == deck.RoleBuilding {
			metrics.buildingCount++
			metrics.defenseQuality += card.LevelRatio()
		}

		// Count support troops (defensive utility)
		if card.Role != nil && *card.Role == deck.RoleSupport {
			metrics.supportCount++
			metrics.defenseQuality += card.LevelRatio() * 0.5
		}

		// Check anti-air capability using combat stats
		if card.Stats != nil && (card.Stats.Targets == "Air" || card.Stats.Targets == "Air & Ground") {
			metrics.antiAirCount++
		}

		// Calculate evolution bonus for defensive cards
		if card.EvolutionLevel > 0 {
			// Buildings and support get larger evolution bonus
			if card.Role != nil && (*card.Role == deck.RoleBuilding || *card.Role == deck.RoleSupport) {
				metrics.evolutionBonus += 0.20 * float64(card.EvolutionLevel)
			} else if card.Stats != nil && (card.Stats.Targets == "Air" || card.Stats.Targets == "Air & Ground") {
				// Anti-air cards get evolution bonus
				metrics.evolutionBonus += 0.15 * float64(card.EvolutionLevel)
			}
		}
	}

	return metrics
}

// scoreAntiAir calculates the anti-air coverage score component (0-10 scale)
func scoreAntiAir(count int) float64 {
	switch {
	case count >= 4:
		return 10.0 // Excellent air defense
	case count >= 3:
		return 8.5
	case count >= 2:
		return 5.5
	case count == 1:
		return 3.0
	default:
		return 0.0 // No anti-air = critical weakness
	}
}

// scoreBuildings calculates the building presence score component (0-10 scale)
func scoreBuildings(count int, quality float64) float64 {
	if count >= 2 {
		return 10.0
	} else if count == 1 {
		return 6.0 + (quality * 2.0)
	}
	return 5.0 // No building is acceptable for aggressive archetypes
}

// scoreSupport calculates the support troop score component (0-10 scale)
func scoreSupport(count int) float64 {
	switch {
	case count >= 4:
		return 10.0
	case count >= 3:
		return 7.0
	case count >= 2:
		return 5.0
	default:
		return 2.0
	}
}

// ScoreDefense calculates the defense score for a deck (0-10 scale)
// Considers anti-air capability, defensive buildings, support troops, and evolution bonuses
func ScoreDefense(deckCards []deck.CardCandidate) CategoryScore {
	if len(deckCards) == 0 {
		return CreateCategoryScore(0, "No cards in deck")
	}

	// Extract metrics from deck cards
	metrics := extractDefenseMetrics(deckCards)

	// Score components (0-10 scale)
	// Anti-air coverage (40% of score)
	antiAirScore := scoreAntiAir(metrics.antiAirCount)

	// Defensive building presence (30% of score)
	buildingScore := scoreBuildings(metrics.buildingCount, metrics.defenseQuality)

	// Support troop presence (30% of score)
	supportScore := scoreSupport(metrics.supportCount)

	// Combine components with weights
	score := (antiAirScore * 0.4) + (buildingScore * 0.3) + (supportScore * 0.3)

	// Add evolution bonus (up to +2.0 points for defense)
	score += metrics.evolutionBonus
	if score > 10.0 {
		score = 10.0
	}

	// Generate assessment text
	assessment := generateDefenseAssessment(metrics.antiAirCount, metrics.buildingCount, score, metrics.evolutionBonus)

	return CreateCategoryScore(score, assessment)
}

// ScoreSynergy calculates the synergy score for a deck (0-10 scale)
// Uses the 188-pair synergy database to find card interactions
func ScoreSynergy(deckCards []deck.CardCandidate, synergyDB *deck.SynergyDatabase) CategoryScore {
	if len(deckCards) == 0 {
		return CreateCategoryScore(0, "No cards in deck")
	}

	if synergyDB == nil {
		// Fallback if no synergy database provided
		return CreateCategoryScore(5.0, "Synergy database not available")
	}

	// Extract card names
	cardNames := make([]string, len(deckCards))
	for i, card := range deckCards {
		cardNames[i] = card.Name
	}

	// Calculate synergy using existing analysis
	analysis := synergyDB.AnalyzeDeckSynergy(cardNames)

	pairCount := 0
	for _, count := range analysis.CategoryScores {
		pairCount += count
	}
	maxPairs := (len(deckCards) * (len(deckCards) - 1)) / 2
	coverage := 0.0
	if maxPairs > 0 {
		coverage = float64(pairCount) / float64(maxPairs)
	}

	// Blend average synergy strength with coverage for a 0-10 scale score.
	// Average synergy drives most of the score, coverage adds context.
	score := ((analysis.AverageScore * 0.75) + (coverage * 0.25)) * 10.0

	// Baseline for decks with no detected synergies.
	if pairCount == 0 {
		score = 2.5
	}

	// Cap at 10.0
	if score > 10.0 {
		score = 10.0
	}

	// Generate assessment based on synergy pairs found
	assessment := generateSynergyAssessment(analysis.TopSynergies, pairCount, score)

	return CreateCategoryScore(score, assessment)
}

// versatilityMetrics holds extracted metrics from deck cards for versatility scoring
type versatilityMetrics struct {
	roleCount     map[deck.CardRole]int
	elixirVariety map[int]bool
	targetsAir    int
	targetsGround int
}

// collectRoleStats collects role, elixir, and targeting statistics from deck cards
func collectRoleStats(deckCards []deck.CardCandidate) versatilityMetrics {
	metrics := versatilityMetrics{
		roleCount:     make(map[deck.CardRole]int),
		elixirVariety: make(map[int]bool),
	}

	for _, card := range deckCards {
		// Count roles
		if card.Role != nil {
			metrics.roleCount[*card.Role]++
		}

		// Track elixir variety
		metrics.elixirVariety[card.Elixir] = true

		// Track targeting capabilities
		if card.Stats != nil {
			if card.Stats.Targets == "Air" || card.Stats.Targets == "Air & Ground" {
				metrics.targetsAir++
			}
			if card.Stats.Targets == "Ground" || card.Stats.Targets == "Air & Ground" || card.Stats.Targets == "Buildings" {
				metrics.targetsGround++
			}
		}
	}

	return metrics
}

// scoreRoleDiversity calculates the role diversity score component (0-10 scale)
func scoreRoleDiversity(uniqueRoles int) float64 {
	if uniqueRoles >= 5 {
		return 10.0 // Excellent role diversity
	} else if uniqueRoles >= 4 {
		return 7.5
	} else if uniqueRoles >= 3 {
		return 5.5
	}
	return float64(uniqueRoles) * 2.0
}

// scoreElixirVariety calculates the elixir variety score component (0-10 scale)
func scoreElixirVariety(elixirDiversity int) float64 {
	if elixirDiversity >= 6 {
		return 10.0
	}
	return float64(elixirDiversity) * 1.5
}

// scoreTargetCoverage calculates the targeting coverage score component (0-10 scale)
func scoreTargetCoverage(targetsAir, targetsGround int) float64 {
	if targetsAir >= 3 && targetsGround >= 6 {
		return 10.0 // Excellent coverage
	} else if targetsAir >= 2 && targetsGround >= 5 {
		return 7.0
	}
	return (float64(targetsAir) + float64(targetsGround)) * 0.5
}

// ScoreVersatility calculates the versatility score for a deck (0-10 scale)
// Considers threat coverage and adaptability to different opponents
func ScoreVersatility(deckCards []deck.CardCandidate) CategoryScore {
	if len(deckCards) == 0 {
		return CreateCategoryScore(0, "No cards in deck")
	}

	// Collect versatility metrics
	metrics := collectRoleStats(deckCards)

	// Score components (0-10 scale)
	// Role diversity (40% of score)
	roleScore := scoreRoleDiversity(len(metrics.roleCount))

	// Elixir variety (30% of score)
	elixirScore := scoreElixirVariety(len(metrics.elixirVariety))

	// Target coverage (30% of score)
	targetScore := scoreTargetCoverage(metrics.targetsAir, metrics.targetsGround)

	// Combine components with weights
	score := (roleScore * 0.4) + (elixirScore * 0.3) + (targetScore * 0.3)

	// Generate assessment text
	assessment := generateVersatilityAssessment(len(metrics.roleCount), len(metrics.elixirVariety), score)

	return CreateCategoryScore(score, assessment)
}

// ScoreF2P calculates the F2P-friendliness score for a deck (0-10 scale)
// Considers card rarity distribution and upgrade path difficulty
func ScoreF2P(deckCards []deck.CardCandidate) CategoryScore {
	if len(deckCards) == 0 {
		return CreateCategoryScore(0, "No cards in deck")
	}

	rarityCount := make(map[string]int)
	totalUpgradeRatio := 0.0
	commonCount := 0
	rareCount := 0
	epicCount := 0
	legendaryCount := 0
	championCount := 0

	for _, card := range deckCards {
		rarityCount[card.Rarity]++
		totalUpgradeRatio += card.LevelRatio()

		switch card.Rarity {
		case "Common":
			commonCount++
		case "Rare":
			rareCount++
		case "Epic":
			epicCount++
		case "Legendary":
			legendaryCount++
		case "Champion":
			championCount++
		}
	}

	// Score components (0-10 scale)
	var score float64

	// Rarity distribution (60% of score)
	rarityScore := 0.0

	// Penalize expensive rarities heavily
	// Champions are hardest to upgrade (4x penalty)
	// Legendaries are hard (3x penalty)
	// Epics are moderate (2x penalty)
	rarityPenalty := float64(championCount)*4.0 + float64(legendaryCount)*3.0 + float64(epicCount)*2.0

	if rarityPenalty == 0 {
		rarityScore = 10.0 // All commons/rares = perfect for F2P
	} else if rarityPenalty <= 2.0 {
		rarityScore = 10.0 - rarityPenalty*2.0
	} else if rarityPenalty <= 4.0 {
		rarityScore = 6.0 - (rarityPenalty-2.0)*1.5
	} else {
		rarityScore = 3.0 - (rarityPenalty-4.0)*0.5
		if rarityScore < 0 {
			rarityScore = 0
		}
	}

	// Upgrade progress (40% of score)
	avgUpgradeRatio := totalUpgradeRatio / float64(len(deckCards))
	upgradeScore := avgUpgradeRatio * 10.0

	// Combine components with weights
	score = (rarityScore * 0.6) + (upgradeScore * 0.4)

	// Generate assessment text
	assessment := generateF2PAssessment(legendaryCount, epicCount, commonCount, score)

	return CreateCategoryScore(score, assessment)
}

// Assessment text generators

func generateAttackAssessment(winConditions int, spellDamage, score, evolutionBonus float64) string {
	evoText := ""
	if evolutionBonus > 0 {
		evoText = fmt.Sprintf(" (+%.1f evolution bonus)", evolutionBonus)
	}

	if score >= 8.0 {
		return fmt.Sprintf("Excellent offensive potential with strong win conditions%s", evoText)
	} else if score >= 6.0 {
		return fmt.Sprintf("Good attack capabilities with decent win conditions%s", evoText)
	} else if score >= 4.0 {
		return fmt.Sprintf("Moderate offensive pressure, could use stronger win conditions%s", evoText)
	} else {
		return "Weak attack potential, lacks reliable win conditions"
	}
}

func generateDefenseAssessment(antiAir, buildings int, score, evolutionBonus float64) string {
	evoText := ""
	if evolutionBonus > 0 {
		evoText = fmt.Sprintf(" (+%.1f evolution bonus)", evolutionBonus)
	}

	if score >= 8.0 {
		return fmt.Sprintf("Solid defensive setup with good air coverage%s", evoText)
	} else if score >= 6.0 {
		return fmt.Sprintf("Decent defensive capabilities%s", evoText)
	} else if antiAir == 0 {
		return "Critical weakness: no anti-air defense"
	} else {
		return "Weak defensive structure, vulnerable to heavy pushes"
	}
}

func generateSynergyAssessment(topSynergies []deck.SynergyPair, pairCount int, score float64) string {
	if score >= 8.0 {
		return "Excellent card synergies with multiple strong combinations"
	} else if score >= 6.0 {
		return "Good synergy between cards"
	} else if score >= 4.0 {
		return "Moderate synergy, some cards work well together"
	} else if score == 2.5 && pairCount == 0 {
		return "No special synergies between cards"
	} else {
		return "Poor synergy, cards don't complement each other well"
	}
}

func generateVersatilityAssessment(roleCount, elixirCount int, score float64) string {
	if score >= 8.0 {
		return "Highly versatile deck with diverse roles and elixir costs"
	} else if score >= 6.0 {
		return "Good versatility with decent role coverage"
	} else if score >= 4.0 {
		return "Moderate versatility, some gaps in role coverage"
	} else {
		return "Limited versatility, narrow strategy focus"
	}
}

func generateF2PAssessment(legendaryCount, epicCount, commonCount int, score float64) string {
	if score >= 8.0 {
		return "Excellent for F2P players with easy upgrade path"
	} else if score >= 6.0 {
		return "Good for F2P with manageable upgrade costs"
	} else if legendaryCount >= 3 {
		return "Difficult for F2P due to multiple legendaries"
	} else if epicCount >= 4 {
		return "Challenging for F2P with many epic cards"
	} else {
		return "Moderate F2P friendliness, some expensive upgrades"
	}
}

// ScorePlayability calculates how playable a deck is based on card availability (0-10 scale)
// Requires PlayerContext to determine which cards are owned
// Returns perfect score (10.0) if no player context (assume all cards available)
func ScorePlayability(deckCards []deck.CardCandidate, playerContext *PlayerContext) CategoryScore {
	if len(deckCards) == 0 {
		return CreateCategoryScore(0, "No cards in deck")
	}

	// No player context - assume deck is fully playable
	if playerContext == nil {
		return CreateCategoryScore(10.0, "Player context not available - assuming all cards owned")
	}

	// Identify missing cards
	analysis := IdentifyMissingCardsWithContext(deckCards, playerContext)

	// Calculate playability score based on card availability
	// Base score: percentage of cards owned
	ownedRatio := float64(analysis.AvailableCount) / float64(len(deckCards))
	baseScore := ownedRatio * 10.0

	// Apply penalties for missing cards
	// Locked cards (arena restrictions) are more severe than unlocked missing cards
	lockedPenalty := 0.0
	unlockedPenalty := 0.0

	for _, card := range analysis.MissingCards {
		if card.IsLocked {
			// -1.5 points per locked card (harder to obtain)
			lockedPenalty += 1.5
		} else {
			// -0.5 points per unlocked missing card (can be obtained)
			unlockedPenalty += 0.5
		}
	}

	score := baseScore - lockedPenalty - unlockedPenalty

	// Ensure score is within 0-10 range
	if score > 10.0 {
		score = 10.0
	}
	if score < 0 {
		score = 0
	}

	// Generate assessment
	assessment := generatePlayabilityAssessment(analysis, score)

	return CreateCategoryScore(score, assessment)
}

// generatePlayabilityAssessment creates playability assessment text
func generatePlayabilityAssessment(analysis *MissingCardsAnalysis, score float64) string {
	if analysis.IsPlayable {
		return "All cards available - deck is fully playable"
	}

	lockedCount := 0
	unlockedCount := 0
	for _, card := range analysis.MissingCards {
		if card.IsLocked {
			lockedCount++
		} else {
			unlockedCount++
		}
	}

	if score >= 8.0 {
		return fmt.Sprintf("Mostly playable - only %d card(s) missing", analysis.MissingCount)
	} else if score >= 5.0 {
		if lockedCount > 0 {
			return fmt.Sprintf("Partially playable - %d card(s) locked by arena, %d obtainable", lockedCount, unlockedCount)
		}
		return fmt.Sprintf("Partially playable - %d card(s) need to be obtained", unlockedCount)
	} else if lockedCount > 0 {
		return fmt.Sprintf("Not playable - %d card(s) locked by arena progression", lockedCount)
	} else {
		return fmt.Sprintf("Not playable - %d card(s) missing from collection", analysis.MissingCount)
	}
}
