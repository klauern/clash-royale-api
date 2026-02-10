package evaluation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

const (
	overallWeightAttack      = 0.23
	overallWeightDefense     = 0.22
	overallWeightSynergy     = 0.21
	overallWeightVersatility = 0.14
	overallWeightF2P         = 0.10
	overallWeightPlayability = 0.10
)

// ============================================================================
// Phase 1: Foundation Helpers - Tier Scoring
// ============================================================================

// scoreTierThresholds applies tiered scoring based on value thresholds
// Returns the score for the first threshold met (thresholds checked in order)
func scoreTierThresholds(value float64, thresholds, scores []float64) float64 {
	for i, threshold := range thresholds {
		if value >= threshold {
			return scores[i]
		}
	}
	return scores[len(scores)-1]
}

// ============================================================================
// Phase 1: Foundation Helpers - Validation
// ============================================================================

// hasRole safely checks if card has the specified role
func hasRole(card deck.CardCandidate, role deck.CardRole) bool {
	return card.Role != nil && *card.Role == role
}

// canTargetAir safely checks if card can target air units
func canTargetAir(card deck.CardCandidate) bool {
	return card.Stats != nil &&
		(card.Stats.Targets == "Air" || card.Stats.Targets == "Air & Ground")
}

// getDPS safely retrieves DPS with default fallback
func getDPS(card deck.CardCandidate) float64 {
	if card.Stats != nil {
		return float64(card.Stats.DamagePerSecond)
	}
	return 0.0
}

// ============================================================================
// Phase 1: Foundation Helpers - Card Filtering
// ============================================================================

// filterByRole returns cards matching the specified role
func filterByRole(cards []deck.CardCandidate, role deck.CardRole) []deck.CardCandidate {
	filtered := []deck.CardCandidate{}
	for _, card := range cards {
		if hasRole(card, role) {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// filterByElixir returns cards with cost <= maxCost
func filterByElixir(cards []deck.CardCandidate, maxCost int) []deck.CardCandidate {
	filtered := []deck.CardCandidate{}
	for _, card := range cards {
		if card.Elixir <= maxCost {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// filterByDPS returns cards with DPS > threshold
func filterByDPS(cards []deck.CardCandidate, threshold float64) []deck.CardCandidate {
	filtered := []deck.CardCandidate{}
	for _, card := range cards {
		if getDPS(card) > threshold {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// filterByAirTargeting returns cards that can target air units
func filterByAirTargeting(cards []deck.CardCandidate) []deck.CardCandidate {
	filtered := []deck.CardCandidate{}
	for _, card := range cards {
		if canTargetAir(card) {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// ============================================================================
// Phase 1: Foundation Helpers - Summary Generation
// ============================================================================

// generateSummaryFromScore returns summary text based on score thresholds
func generateSummaryFromScore(score float64, summaries map[float64]string) string {
	// Sort thresholds descending
	type threshold struct {
		text  string
		score float64
	}
	thresholds := []threshold{}
	for minScore, text := range summaries {
		thresholds = append(thresholds, threshold{text, minScore})
	}
	sort.Slice(thresholds, func(i, j int) bool {
		return thresholds[i].score > thresholds[j].score
	})

	// Find first matching threshold
	for _, t := range thresholds {
		if score >= t.score {
			return t.text
		}
	}
	return thresholds[len(thresholds)-1].text
}

// ============================================================================
// Phase 1: Foundation Helpers - Legacy Functions
// ============================================================================

// countAirTargeters returns cards that can target air units
// Deprecated: Use filterByAirTargeting instead
func countAirTargeters(cards []deck.CardCandidate) []deck.CardCandidate {
	return filterByAirTargeting(cards)
}

// calculateElixirCurve returns distribution of cards across elixir costs
func calculateElixirCurve(cards []deck.CardCandidate) map[int]int {
	curve := make(map[int]int)
	for _, card := range cards {
		curve[card.Elixir]++
	}
	return curve
}

// findShortestCycle returns the sum of 4 cheapest cards and their names
func findShortestCycle(cards []deck.CardCandidate) (int, []string) {
	if len(cards) < 4 {
		return 0, []string{}
	}

	// Sort cards by elixir cost
	sorted := make([]deck.CardCandidate, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Elixir < sorted[j].Elixir
	})

	// Get 4 cheapest cards
	total := 0
	names := []string{}
	for i := 0; i < 4; i++ {
		total += sorted[i].Elixir
		names = append(names, sorted[i].Name)
	}

	return total, names
}

// buildCardList formats card names with elixir costs
// Example: "Musketeer (4), Baby Dragon (4), Mega Minion (3)"
func buildCardList(cards []deck.CardCandidate) string {
	if len(cards) == 0 {
		return ""
	}

	parts := make([]string, len(cards))
	for i, card := range cards {
		parts[i] = fmt.Sprintf("%s (%d)", card.Name, card.Elixir)
	}
	return strings.Join(parts, ", ")
}

// calculateDeckAvgElixir calculates average elixir cost of deck
// Note: Similar function exists in archetype.go but kept separate to avoid circular dependency
func calculateDeckAvgElixir(cards []deck.CardCandidate) float64 {
	if len(cards) == 0 {
		return 0.0
	}

	total := 0
	for _, card := range cards {
		total += card.Elixir
	}

	return float64(total) / float64(len(cards))
}

// ============================================================================
// Phase 2: Simple Analysis Builders (Defense & Attack)
// ============================================================================

// BuildDefenseAnalysis creates detailed defense analysis
func BuildDefenseAnalysis(deckCards []deck.CardCandidate) AnalysisSection {
	// Get numeric score from existing function
	categoryScore := ScoreDefense(deckCards)

	// Count and identify defensive elements using helper functions
	airTargeters := filterByAirTargeting(deckCards)
	buildings := filterByRole(deckCards, deck.RoleBuilding)
	tankKillers := filterByDPS(deckCards, 150.0)

	// Investment cards (high elixir win conditions)
	investments := []deck.CardCandidate{}
	winConditions := filterByRole(deckCards, deck.RoleWinCondition)
	for _, card := range winConditions {
		if card.Elixir >= 6 {
			investments = append(investments, card)
		}
	}

	// Build details array
	details := []string{}

	// Anti-air coverage
	if len(airTargeters) > 0 {
		details = append(details, fmt.Sprintf("Anti-air units (%d): %s",
			len(airTargeters), buildCardList(airTargeters)))
	} else {
		details = append(details, "⚠️  No anti-air units - vulnerable to aerial threats")
	}

	// Defensive buildings
	if len(buildings) > 0 {
		details = append(details, fmt.Sprintf("Defensive buildings: %s", buildCardList(buildings)))
	} else {
		details = append(details, "⚠️  No defensive buildings - vulnerable to bridge spam")
	}

	// Tank killers
	if len(tankKillers) > 0 {
		details = append(details, fmt.Sprintf("Tank killers: %s provides strong ground defense", tankKillers[0].Name))
	}

	// Investment protection
	if len(investments) > 0 {
		details = append(details, fmt.Sprintf("⚠️  %s (%d elixir) needs defensive support",
			investments[0].Name, investments[0].Elixir))
	}

	// Generate summary using helper function
	airCount := float64(len(airTargeters))
	buildingCount := float64(len(buildings))

	summary := "Solid defensive capabilities"
	if airCount == 0 {
		summary = "No anti-air coverage - vulnerable to aerial threats"
	} else if airCount < 2 {
		summary = "Weak anti-air coverage"
	} else if buildingCount == 0 {
		summary = "Good anti-air but lacks defensive buildings"
	} else if airCount >= 3 && buildingCount >= 1 {
		summary = "Excellent defensive coverage with strong anti-air and buildings"
	}

	return AnalysisSection{
		Title:   "Defense Analysis",
		Summary: summary,
		Details: details,
		Score:   categoryScore.Score,
		Rating:  categoryScore.Rating,
	}
}

// classifyWinCondition determines win condition category
func classifyWinCondition(cardName string) string {
	// Direct damage
	directDamage := map[string]bool{
		"Hog Rider": true, "Giant": true, "Royal Giant": true,
		"Balloon": true, "Golem": true, "Lava Hound": true,
		"Electro Giant": true, "Royal Hogs": true, "Ram Rider": true,
	}

	// Siege
	siege := map[string]bool{
		"X-Bow": true, "Mortar": true,
	}

	// Chip damage
	chip := map[string]bool{
		"Miner": true, "Goblin Barrel": true, "Graveyard": true,
		"Goblin Drill": true, "Wall Breakers": true,
	}

	// Bridge spam
	bridgeSpam := map[string]bool{
		"Battle Ram": true, "P.E.K.K.A": true, "Mega Knight": true,
		"Royal Ghost": true, "Bandit": true, "Ram Rider": true,
	}

	if directDamage[cardName] {
		return "Direct Damage"
	}
	if siege[cardName] {
		return "Siege"
	}
	if chip[cardName] {
		return "Chip Damage"
	}
	if bridgeSpam[cardName] {
		return "Bridge Spam"
	}

	return "Win Condition"
}

// BuildAttackAnalysis creates detailed attack analysis
func BuildAttackAnalysis(deckCards []deck.CardCandidate) AnalysisSection {
	// Get numeric score from existing function
	categoryScore := ScoreAttack(deckCards)

	// Identify offensive elements using helper functions
	winConditions := filterByRole(deckCards, deck.RoleWinCondition)
	bigSpells := filterByRole(deckCards, deck.RoleSpellBig)

	// Build details array
	details := []string{}

	// Win conditions
	if len(winConditions) > 0 {
		category := classifyWinCondition(winConditions[0].Name)
		details = append(details, fmt.Sprintf("Primary win condition: %s (%s)",
			winConditions[0].Name, category))

		if len(winConditions) > 1 {
			category2 := classifyWinCondition(winConditions[1].Name)
			details = append(details, fmt.Sprintf("Secondary win condition: %s (%s)",
				winConditions[1].Name, category2))
		}
	} else {
		details = append(details, "⚠️  No dedicated win condition - may struggle to take towers")
	}

	// Spell damage
	if len(bigSpells) > 0 {
		spellList := buildCardList(bigSpells)
		assessment := "excellent"
		if len(bigSpells) == 1 {
			assessment = "good"
		}
		details = append(details, fmt.Sprintf("Spell damage: %s - %s finishing power",
			spellList, assessment))
	}

	// Bridge spam potential
	bridgeCards := []string{}
	for _, card := range winConditions {
		if classifyWinCondition(card.Name) == "Bridge Spam" {
			bridgeCards = append(bridgeCards, card.Name)
		}
	}
	if len(bridgeCards) > 0 {
		details = append(details, fmt.Sprintf("Bridge spam potential: %s can punish elixir disadvantage",
			strings.Join(bridgeCards, ", ")))
	}

	// Strategic recommendation
	if len(winConditions) > 0 {
		category := classifyWinCondition(winConditions[0].Name)
		switch category {
		case "Direct Damage":
			details = append(details, "Strategy: Apply consistent pressure with direct tower damage")
		case "Siege":
			details = append(details, "Strategy: Establish defensive perimeter and chip tower from range")
		case "Chip Damage":
			details = append(details, "Strategy: Accumulate small amounts of damage over time")
		case "Bridge Spam":
			details = append(details, "Strategy: Capitalize on elixir advantages with fast pushes")
		}
	}

	// Generate summary
	summary := "Strong offensive potential"
	if len(winConditions) == 0 {
		summary = "Lacks dedicated win condition"
	} else if len(winConditions) >= 2 {
		summary = "Versatile offense with multiple win conditions"
	} else if len(bigSpells) >= 2 {
		summary = "Strong offensive pressure with spell support"
	}

	return AnalysisSection{
		Title:   "Attack Analysis",
		Summary: summary,
		Details: details,
		Score:   categoryScore.Score,
		Rating:  categoryScore.Rating,
	}
}

// ============================================================================
// Phase 3: Complex Analysis Builders (Bait & Cycle)
// ============================================================================

// identifyBaitCards groups cards by spell vulnerability
func identifyBaitCards(cards []deck.CardCandidate) map[string][]string {
	baitGroups := make(map[string][]string)

	// Define bait categories
	logBait := map[string]bool{
		"Goblin Gang": true, "Princess": true, "Dart Goblin": true,
		"Goblin Barrel": true, "Skeleton Barrel": true, "Rascals": true,
	}

	zapBait := map[string]bool{
		"Minion Horde": true, "Skeleton Army": true, "Bats": true,
		"Inferno Dragon": true, "Inferno Tower": true, "Sparky": true,
	}

	arrowsBait := map[string]bool{
		"Minions": true, "Spear Goblins": true, "Princess": true,
		"Dart Goblin": true, "Firecracker": true,
	}

	fireballBait := map[string]bool{
		"Three Musketeers": true, "Wizard": true, "Witch": true,
		"Flying Machine": true, "Elixir Collector": true, "Night Witch": true,
	}

	// Categorize cards
	for _, card := range cards {
		if logBait[card.Name] {
			baitGroups["Log"] = append(baitGroups["Log"], card.Name)
		}
		if zapBait[card.Name] {
			baitGroups["Zap"] = append(baitGroups["Zap"], card.Name)
		}
		if arrowsBait[card.Name] {
			baitGroups["Arrows"] = append(baitGroups["Arrows"], card.Name)
		}
		if fireballBait[card.Name] {
			baitGroups["Fireball"] = append(baitGroups["Fireball"], card.Name)
		}
	}

	return baitGroups
}

// calculateBaitScore computes bait potential (0-10)
func calculateBaitScore(baitGroups map[string][]string, hasGoblinBarrel, hasGoblinDrill bool) float64 {
	// Count total bait cards
	totalBaitCards := 0
	for _, cards := range baitGroups {
		totalBaitCards += len(cards)
	}

	// Count spell groups with 2+ cards (shared counter potential)
	sharedCounterGroups := 0
	for _, cards := range baitGroups {
		if len(cards) >= 2 {
			sharedCounterGroups++
		}
	}

	// Win condition fit
	winConditionFit := 0.0
	if hasGoblinBarrel || hasGoblinDrill {
		winConditionFit = 10.0
	} else if totalBaitCards >= 2 {
		winConditionFit = 6.0
	}

	// Bait card count score (50%) - using tier scoring
	baitCountScore := 0.0
	if totalBaitCards >= 4 {
		baitCountScore = 10.0
	} else if totalBaitCards == 3 {
		baitCountScore = 7.5
	} else if totalBaitCards == 2 {
		baitCountScore = 5.0
	} else if totalBaitCards == 1 {
		baitCountScore = 2.5
	}

	// Shared counter potential (30%) - using tier scoring
	sharedCounterScore := 0.0
	if sharedCounterGroups >= 3 {
		sharedCounterScore = 10.0
	} else if sharedCounterGroups == 2 {
		sharedCounterScore = 7.0
	} else if sharedCounterGroups == 1 {
		sharedCounterScore = 4.0
	}

	// Combine components
	score := (baitCountScore * 0.5) + (sharedCounterScore * 0.3) + (winConditionFit * 0.2)

	return score
}

// BuildBaitAnalysis creates detailed bait analysis
func BuildBaitAnalysis(deckCards []deck.CardCandidate) AnalysisSection {
	// Identify bait cards
	baitGroups := identifyBaitCards(deckCards)

	// Check for bait-friendly win conditions
	hasGoblinBarrel := false
	hasGoblinDrill := false
	for _, card := range deckCards {
		if card.Name == "Goblin Barrel" {
			hasGoblinBarrel = true
		}
		if card.Name == "Goblin Drill" {
			hasGoblinDrill = true
		}
	}

	// Calculate score
	score := calculateBaitScore(baitGroups, hasGoblinBarrel, hasGoblinDrill)
	rating := ScoreToRating(score)

	// Build details array
	details := []string{}

	// List bait groups
	for spell, cards := range baitGroups {
		if len(cards) >= 2 {
			details = append(details, fmt.Sprintf("%s bait units (%d): %s",
				spell, len(cards), strings.Join(cards, ", ")))
		}
	}

	// Find strongest bait chain
	maxSpell := ""
	maxCount := 0
	for spell, cards := range baitGroups {
		if len(cards) > maxCount {
			maxCount = len(cards)
			maxSpell = spell
		}
	}
	if maxCount >= 2 {
		details = append(details, fmt.Sprintf("Strongest bait chain: %s (%d vulnerable cards)",
			maxSpell, maxCount))
		details = append(details, "Mind-game potential: Opponent must choose which threat to spell")
	}

	// Win condition fit
	if hasGoblinBarrel {
		details = append(details, "Win condition fit: Goblin Barrel benefits from bait pressure")
	} else if hasGoblinDrill {
		details = append(details, "Win condition fit: Goblin Drill benefits from bait pressure")
	} else if score < 3.0 {
		details = append(details, "⚠️  Not a bait deck - lacks spell-vulnerable units")
	}

	// Generate summary
	summary := "Moderate bait potential"
	if score >= 7.0 {
		summary = "Excellent spell bait with multiple vulnerable units"
	} else if score < 3.0 {
		summary = "Not a bait-focused deck"
	}

	return AnalysisSection{
		Title:   "Bait Analysis",
		Summary: summary,
		Details: details,
		Score:   score,
		Rating:  rating,
	}
}

// calculateCycleScore computes cycle efficiency (0-10)
func calculateCycleScore(avgElixir float64, lowCostCount, shortestCycle int) float64 {
	// Cycle speed score (40%)
	cycleSpeedScore := 0.0
	if avgElixir < 3.0 {
		cycleSpeedScore = 10.0
	} else if avgElixir < 3.3 {
		cycleSpeedScore = 9.0
	} else if avgElixir < 3.6 {
		cycleSpeedScore = 7.0
	} else if avgElixir < 4.0 {
		cycleSpeedScore = 5.0
	} else {
		cycleSpeedScore = 3.0
	}

	// Low-cost card count score (35%)
	lowCostScore := 0.0
	if lowCostCount >= 4 {
		lowCostScore = 10.0
	} else if lowCostCount == 3 {
		lowCostScore = 7.0
	} else if lowCostCount == 2 {
		lowCostScore = 4.0
	} else if lowCostCount == 1 {
		lowCostScore = 2.0
	}

	// Shortest cycle score (25%)
	shortestCycleScore := 0.0
	if shortestCycle <= 6 {
		shortestCycleScore = 10.0
	} else if shortestCycle <= 8 {
		shortestCycleScore = 7.0
	} else if shortestCycle <= 10 {
		shortestCycleScore = 4.0
	} else {
		shortestCycleScore = 2.0
	}

	// Combine components
	score := (cycleSpeedScore * 0.4) + (lowCostScore * 0.35) + (shortestCycleScore * 0.25)

	return score
}

// BuildCycleAnalysis creates detailed cycle analysis
func BuildCycleAnalysis(deckCards []deck.CardCandidate) AnalysisSection {
	// Calculate cycle metrics
	avgElixir := calculateDeckAvgElixir(deckCards)
	shortestCycle, _ := findShortestCycle(deckCards)
	elixirCurve := calculateElixirCurve(deckCards)

	// Count low-cost cards (≤ 2 elixir) using helper
	lowCostCards := filterByElixir(deckCards, 2)
	lowCostCount := len(lowCostCards)

	// Calculate score
	score := calculateCycleScore(avgElixir, lowCostCount, shortestCycle)
	rating := ScoreToRating(score)

	// Build details array
	details := []string{}

	// Average elixir
	cycleType := "Slow"
	if avgElixir < 3.0 {
		cycleType = "Fast"
	} else if avgElixir < 3.6 {
		cycleType = "Medium"
	}
	details = append(details, fmt.Sprintf("Average elixir: %.1f (%s Cycle)", avgElixir, cycleType))

	// Cycle cards
	if lowCostCount > 0 {
		details = append(details, fmt.Sprintf("Cycle cards (%d): %s",
			lowCostCount, buildCardList(lowCostCards)))
	}

	// Shortest 4-card cycle
	cycleAssessment := "poor rotation"
	if shortestCycle <= 6 {
		cycleAssessment = "excellent rotation"
	} else if shortestCycle <= 8 {
		cycleAssessment = "good rotation"
	}
	details = append(details, fmt.Sprintf("Shortest 4-card cycle: %d elixir (%s)",
		shortestCycle, cycleAssessment))

	// Rotation estimate (find win condition) using helper
	winConditions := filterByRole(deckCards, deck.RoleWinCondition)
	if len(winConditions) > 0 {
		winCondition := winConditions[0].Name
		rotationTime := int(avgElixir * 3.5) // Rough estimate
		details = append(details, fmt.Sprintf("Rotation estimate: Can return to %s in ~%d seconds",
			winCondition, rotationTime))
	}

	// Elixir curve distribution
	curveStr := ""
	for cost := 1; cost <= 8; cost++ {
		if count, ok := elixirCurve[cost]; ok && count > 0 {
			if curveStr != "" {
				curveStr += ", "
			}
			curveStr += fmt.Sprintf("%d-cost (%d)", cost, count)
		}
	}
	if curveStr != "" {
		details = append(details, fmt.Sprintf("Elixir curve: %s", curveStr))
	}

	// Tempo description
	if avgElixir < 3.2 {
		details = append(details, "Tempo: Constant pressure through rapid cycling")
	} else if avgElixir >= 4.0 {
		details = append(details, "Tempo: Slower build-up with larger pushes")
	}

	// Generate summary
	summary := "Medium cycle speed"
	if avgElixir < 3.0 {
		summary = "Fast cycle deck with excellent rotation speed"
	} else if avgElixir >= 4.0 {
		summary = "Slow cycle - focuses on larger pushes"
	}

	return AnalysisSection{
		Title:   "Cycle Analysis",
		Summary: summary,
		Details: details,
		Score:   score,
		Rating:  rating,
	}
}

// ============================================================================
// Phase 4: Ladder Analysis
// ============================================================================

// isLevelIndependent determines if card is effective when underleveled
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

	score := (rarityScore * 0.4) + (levelIndepScore * 0.3) + (upgradeProgress * 0.2) + (rarityScore * 0.1)

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
	if competitiveViability >= 9.0 {
		return "Tournament ready"
	} else if competitiveViability >= 7.0 {
		return "Ladder competitive"
	} else if competitiveViability >= 5.0 {
		return "Playable but underleveled"
	} else if competitiveViability >= 3.0 {
		return "Significant disadvantage"
	}
	return "Not competitive"
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
		rating = Rating(viabilityRating)
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
		details = append(details, fmt.Sprintf("Cards ready for ladder: %d/8 (level 11+)", readyCount))
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
func filterEvolvableCards(deckCards []deck.CardCandidate) (evolvable, evolved []deck.CardCandidate) {
	for _, card := range deckCards {
		if card.MaxEvolutionLevel > 0 {
			evolvable = append(evolvable, card)
			if card.EvolutionLevel > 0 {
				evolved = append(evolved, card)
			}
		}
	}
	return evolvable, evolved
}

// calculateEvolutionPotential calculates evolution score (0-10)
func calculateEvolutionPotential(evolvableCards, evolvedCards []deck.CardCandidate) float64 {
	if len(evolvableCards) == 0 {
		return 0.0
	}

	// Base score: percentage of evolvable cards that are evolved
	evolutionRatio := float64(len(evolvedCards)) / float64(len(evolvableCards))
	score := evolutionRatio * 10.0

	// Add bonus for multiple evolved cards
	if len(evolvedCards) >= 2 {
		score += 1.0
	}
	if len(evolvedCards) >= 3 {
		score += 0.5
	}

	// Cap at 10
	if score > 10.0 {
		score = 10.0
	}

	return score
}

// addEvolutionProgressDetails adds player-specific evolution details
func addEvolutionProgressDetails(details []string, evolvableCards, evolvedCards []deck.CardCandidate, playerContext *PlayerContext) []string {
	if playerContext == nil {
		return details
	}

	unlockedEvolutions := playerContext.GetUnlockedEvolutionCards()
	details = append(details, fmt.Sprintf("Your unlocked evolutions: %d cards", len(unlockedEvolutions)))

	// Check which deck cards can evolve
	readyToEvolve := []string{}
	for _, card := range evolvableCards {
		if card.EvolutionLevel == 0 && playerContext.CanEvolve(card.Name) {
			readyToEvolve = append(readyToEvolve, card.Name)
		}
	}
	if len(readyToEvolve) > 0 {
		details = append(details, fmt.Sprintf("Ready to evolve: %s", strings.Join(readyToEvolve, ", ")))
	}

	// Show evolution progress for key cards
	if len(evolvedCards) > 0 {
		details = append(details, "Evolution progress:")
		for _, card := range evolvedCards {
			currentLevel, maxLevel, currentCount, requiredCount := playerContext.GetEvolutionProgress(card.Name)
			details = append(details, fmt.Sprintf("  %s: Level %d/%d, %d/%d cards",
				card.Name, currentLevel, maxLevel, currentCount, requiredCount))
		}
	}

	return details
}

// BuildEvolutionAnalysis creates detailed evolution analysis
// If playerContext is provided, shows player's evolution status
// If playerContext is nil, shows generic evolution potential for the deck
func BuildEvolutionAnalysis(deckCards []deck.CardCandidate, playerContext *PlayerContext) AnalysisSection {
	details := []string{}
	var score float64
	var rating Rating
	var summary string

	// Identify evolvable cards using helper
	evolvableInDeck, evolvedInDeck := filterEvolvableCards(deckCards)

	// Calculate evolution score (0-10)
	if len(evolvableInDeck) == 0 {
		score = 0.0
		summary = "No evolvable cards in deck"
		details = append(details, "This deck contains no cards with evolution potential")
	} else {
		// Calculate score using helper
		score = calculateEvolutionPotential(evolvableInDeck, evolvedInDeck)

		// Generate summary
		if len(evolvedInDeck) == 0 {
			summary = fmt.Sprintf("Deck has %d evolvable card(s) but none evolved", len(evolvableInDeck))
		} else if len(evolvedInDeck) == 1 {
			summary = fmt.Sprintf("Deck has 1 evolved card out of %d evolvable", len(evolvableInDeck))
		} else {
			summary = fmt.Sprintf("Deck has %d evolved cards out of %d evolvable", len(evolvedInDeck), len(evolvableInDeck))
		}

		// List evolvable cards
		if len(evolvableInDeck) > 0 {
			cardNames := []string{}
			for _, card := range evolvableInDeck {
				if card.EvolutionLevel > 0 {
					cardNames = append(cardNames, fmt.Sprintf("%s (Evo.%d/%d)", card.Name, card.EvolutionLevel, card.MaxEvolutionLevel))
				} else {
					cardNames = append(cardNames, fmt.Sprintf("%s (unevolved)", card.Name))
				}
			}
			details = append(details, fmt.Sprintf("Evolvable cards (%d): %s", len(evolvableInDeck), strings.Join(cardNames, ", ")))
		}

		// Add player-specific evolution details using helper
		details = addEvolutionProgressDetails(details, evolvableInDeck, evolvedInDeck, playerContext)

		// Evolution slot strategy
		if len(evolvedInDeck) > 2 {
			details = append(details, "⚠️  More than 2 evolved cards - prioritize best 2 for active slots")
		}

		// Evolution impact assessment
		if score >= 7.0 {
			details = append(details, "Evolution impact: Strong - evolutions significantly boost deck power")
		} else if score >= 4.0 {
			details = append(details, "Evolution impact: Moderate - some evolution synergy present")
		} else {
			details = append(details, "Evolution impact: Low - consider evolving key cards")
		}
	}

	rating = ScoreToRating(score)

	return AnalysisSection{
		Title:   "Evolution Analysis",
		Summary: summary,
		Details: details,
		Score:   score,
		Rating:  rating,
	}
}

// ============================================================================
// Phase 5: Main Orchestrator
// ============================================================================

// applyCriticalFlawPenalties applies additional penalties for critical compositional flaws
// that make a deck fundamentally unviable, beyond what category scores capture
func applyCriticalFlawPenalties(baseScore float64, deckCards []deck.CardCandidate) float64 {
	score := baseScore

	// Check for critical attack flaws
	winConditionCount := 0
	spellCount := 0
	bigSpellCount := 0
	for _, card := range deckCards {
		if card.Role != nil {
			if *card.Role == deck.RoleWinCondition {
				winConditionCount++
			}
			if *card.Role == deck.RoleSpellBig {
				bigSpellCount++
			}
			if *card.Role == deck.RoleSpellBig || *card.Role == deck.RoleSpellSmall {
				spellCount++
			}
		}
	}

	// Penalty for no win condition: -2.0 points (critical flaw)
	if winConditionCount == 0 {
		score -= 2.0
	}

	// Penalty for no spells: -1.5 points (severe limitation)
	if spellCount == 0 {
		score -= 1.5
	}

	// Check for critical defense flaws
	antiAirCount := 0
	for _, card := range deckCards {
		if card.Stats != nil && (card.Stats.Targets == "Air" || card.Stats.Targets == "Air & Ground") {
			antiAirCount++
		}
	}

	// Penalty for no anti-air: -2.0 points (critical vulnerability)
	if antiAirCount == 0 {
		score -= 2.0
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	return score
}

// Evaluate performs comprehensive deck evaluation with all scoring and analysis
// If playerContext is provided, evaluation will include player-specific context such as:
// - Card levels from player's collection
// - Arena-specific card availability
// - Evolution unlock status
func Evaluate(deckCards []deck.CardCandidate, synergyDB *deck.SynergyDatabase, playerContext *PlayerContext) EvaluationResult {
	// Extract deck card names
	deckNames := make([]string, len(deckCards))
	for i, card := range deckCards {
		deckNames[i] = card.Name
	}

	// Calculate average elixir
	avgElixir := calculateDeckAvgElixir(deckCards)

	// Phase 1: Category Scoring
	attackScore := ScoreAttack(deckCards)
	defenseScore := ScoreDefense(deckCards)
	synergyScore := ScoreSynergy(deckCards, synergyDB)
	versatilityScore := ScoreVersatility(deckCards)
	f2pScore := ScoreF2P(deckCards)
	playabilityScore := ScorePlayability(deckCards, playerContext)

	// Phase 2: Archetype Detection
	archetypeResult := DetectArchetype(deckCards)

	// Phase 3: Build Analysis Sections
	defenseAnalysis := BuildDefenseAnalysis(deckCards)
	attackAnalysis := BuildAttackAnalysis(deckCards)
	baitAnalysis := BuildBaitAnalysis(deckCards)
	cycleAnalysis := BuildCycleAnalysis(deckCards)
	ladderAnalysis := BuildLadderAnalysis(deckCards, playerContext)
	evolutionAnalysis := BuildEvolutionAnalysis(deckCards, playerContext)

	// Phase 4: Calculate Overall Score (weighted average)
	// Weights: Attack 23%, Defense 22%, Synergy 21%, Versatility 14%, F2P 10%, Playability 10%
	// Balanced emphasis on attack/defense/synergy fundamentals
	// Critical flaws are separately penalized via applyCriticalFlawPenalties
	// When player context is available, replace Playability with ladder viability at the same weight.
	overallScore := (attackScore.Score * overallWeightAttack) +
		(defenseScore.Score * overallWeightDefense) +
		(synergyScore.Score * overallWeightSynergy) +
		(versatilityScore.Score * overallWeightVersatility) +
		(f2pScore.Score * overallWeightF2P) +
		(playabilityScore.Score * overallWeightPlayability)

	if playerContext != nil {
		overallScore = overallScore - (playabilityScore.Score * overallWeightPlayability) + (ladderAnalysis.Score * overallWeightPlayability)
	}

	// Apply penalties for critical compositional flaws
	// These are severe enough to warrant direct overall score penalties
	overallScore = applyCriticalFlawPenalties(overallScore, deckCards)

	overallRating := ScoreToRating(overallScore)

	// Build synergy matrix (if database provided)
	var synergyMatrix SynergyMatrix
	if synergyDB != nil {
		synergyAnalysis := synergyDB.AnalyzeDeckSynergy(deckNames)
		if synergyAnalysis != nil {
			maxPairs := (len(deckNames) * (len(deckNames) - 1)) / 2
			pairCount := 0
			for _, count := range synergyAnalysis.CategoryScores {
				pairCount += count
			}
			coverage := 0.0
			if maxPairs > 0 {
				coverage = float64(pairCount) / float64(maxPairs) * 100.0
			}

			synergyMatrix = SynergyMatrix{
				Pairs:            synergyAnalysis.TopSynergies,
				TotalScore:       synergyScore.Score,
				AverageSynergy:   synergyAnalysis.AverageScore,
				PairCount:        pairCount,
				MaxPossiblePairs: maxPairs,
				SynergyCoverage:  coverage,
			}
		}
	}

	// Analyze missing cards if player context provided
	var missingCardsAnalysis *MissingCardsAnalysis
	if playerContext != nil {
		missingCardsAnalysis = IdentifyMissingCardsWithContext(deckCards, playerContext)

		// Apply arena-based score penalties for locked cards
		if missingCardsAnalysis != nil && missingCardsAnalysis.MissingCount > 0 {
			lockedCount := 0
			for _, card := range missingCardsAnalysis.MissingCards {
				if card.IsLocked {
					lockedCount++
				}
			}

			// Penalty: -2 points per locked card, -1 point per unlocked but missing card
			penalty := float64(lockedCount)*2.0 + float64(missingCardsAnalysis.MissingCount-lockedCount)*1.0
			overallScore -= penalty

			// Ensure score doesn't go below 0
			if overallScore < 0 {
				overallScore = 0
			}

			// Recalculate overall rating with penalty applied
			overallRating = ScoreToRating(overallScore)
		}
	}

	// Assemble complete result
	return EvaluationResult{
		Deck:      deckNames,
		AvgElixir: avgElixir,

		Attack:      attackScore,
		Defense:     defenseScore,
		Synergy:     synergyScore,
		Versatility: versatilityScore,
		F2PFriendly: f2pScore,
		Playability: playabilityScore,

		OverallScore:  overallScore,
		OverallRating: overallRating,

		DetectedArchetype:   archetypeResult.Primary,
		ArchetypeConfidence: archetypeResult.PrimaryConfidence,

		DefenseAnalysis:   defenseAnalysis,
		AttackAnalysis:    attackAnalysis,
		BaitAnalysis:      baitAnalysis,
		CycleAnalysis:     cycleAnalysis,
		LadderAnalysis:    ladderAnalysis,
		EvolutionAnalysis: evolutionAnalysis,

		SynergyMatrix:        synergyMatrix,
		MissingCardsAnalysis: missingCardsAnalysis,
	}
}
