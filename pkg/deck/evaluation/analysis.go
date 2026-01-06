package evaluation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// ============================================================================
// Phase 1: Foundation Helpers
// ============================================================================

// countAirTargeters returns cards that can target air units
func countAirTargeters(cards []deck.CardCandidate) []deck.CardCandidate {
	airTargeters := []deck.CardCandidate{}
	for _, card := range cards {
		if card.Stats != nil {
			targets := card.Stats.Targets
			if targets == "Air" || targets == "Air & Ground" {
				airTargeters = append(airTargeters, card)
			}
		}
	}
	return airTargeters
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

	// Count and identify defensive elements
	airTargeters := countAirTargeters(deckCards)
	buildings := []deck.CardCandidate{}
	tankKillers := []deck.CardCandidate{}
	investments := []deck.CardCandidate{}

	for _, card := range deckCards {
		// Defensive buildings
		if card.Role != nil && *card.Role == deck.RoleBuilding {
			buildings = append(buildings, card)
		}

		// Tank killers (high DPS > 150)
		if card.Stats != nil && card.Stats.DamagePerSecond > 150 {
			tankKillers = append(tankKillers, card)
		}

		// Investment cards (high elixir win conditions)
		if card.Role != nil && *card.Role == deck.RoleWinCondition && card.Elixir >= 6 {
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

	// Generate summary
	summary := "Solid defensive capabilities"
	if len(airTargeters) == 0 {
		summary = "No anti-air coverage - vulnerable to aerial threats"
	} else if len(airTargeters) < 2 {
		summary = "Weak anti-air coverage"
	} else if len(buildings) == 0 {
		summary = "Good anti-air but lacks defensive buildings"
	} else if len(airTargeters) >= 3 && len(buildings) >= 1 {
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

	// Identify offensive elements
	winConditions := []deck.CardCandidate{}
	bigSpells := []deck.CardCandidate{}

	for _, card := range deckCards {
		if card.Role != nil && *card.Role == deck.RoleWinCondition {
			winConditions = append(winConditions, card)
		}
		if card.Role != nil && *card.Role == deck.RoleSpellBig {
			bigSpells = append(bigSpells, card)
		}
	}

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

	// Bait card count score (50%)
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

	// Shared counter potential (30%)
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

	// Count low-cost cards (≤ 2 elixir)
	lowCostCount := 0
	lowCostCards := []deck.CardCandidate{}
	for _, card := range deckCards {
		if card.Elixir <= 2 {
			lowCostCount++
			lowCostCards = append(lowCostCards, card)
		}
	}

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

	// Rotation estimate (find win condition)
	winCondition := ""
	for _, card := range deckCards {
		if card.Role != nil && *card.Role == deck.RoleWinCondition {
			winCondition = card.Name
			break
		}
	}
	if winCondition != "" {
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

// BuildLadderAnalysis creates detailed ladder analysis
func BuildLadderAnalysis(deckCards []deck.CardCandidate) AnalysisSection {
	// Get F2P score for rarity assessment
	f2pScore := ScoreF2P(deckCards)

	// Count rarity breakdown
	rarityCount := map[string]int{
		"Common":    0,
		"Rare":      0,
		"Epic":      0,
		"Legendary": 0,
		"Champion":  0,
	}

	levelIndepCards := []deck.CardCandidate{}
	totalLevelRatio := 0.0

	for _, card := range deckCards {
		rarityCount[card.Rarity]++
		if isLevelIndependent(card) {
			levelIndepCards = append(levelIndepCards, card)
		}
		totalLevelRatio += card.LevelRatio()
	}

	// Calculate level-independence score
	levelIndepScore := float64(len(levelIndepCards)) / float64(len(deckCards)) * 10.0

	// Calculate average upgrade progress
	avgProgress := totalLevelRatio / float64(len(deckCards))
	upgradeProgressScore := avgProgress * 10.0

	// Calculate ladder score
	score := calculateLadderScore(f2pScore.Score, levelIndepScore, upgradeProgressScore)
	rating := ScoreToRating(score)

	// Build details array
	details := []string{}

	// Rarity breakdown
	details = append(details, fmt.Sprintf("Rarity breakdown: %d Commons, %d Rares, %d Epics, %d Legendaries, %d Champions",
		rarityCount["Common"], rarityCount["Rare"], rarityCount["Epic"],
		rarityCount["Legendary"], rarityCount["Champion"]))

	// Level-independent cards
	if len(levelIndepCards) > 0 {
		details = append(details, fmt.Sprintf("Level-independent cards (%d): %s",
			len(levelIndepCards), buildCardList(levelIndepCards)))
	}

	// Upgrade priority (top 3 most impactful)
	priorities := []string{}
	for _, card := range deckCards {
		if card.Role != nil && *card.Role == deck.RoleWinCondition {
			priorities = append(priorities, fmt.Sprintf("%s (win condition)", card.Name))
		}
	}
	for _, card := range deckCards {
		if len(priorities) >= 3 {
			break
		}
		if card.Role != nil && *card.Role == deck.RoleSupport && card.Rarity != "Common" {
			priorities = append(priorities, fmt.Sprintf("%s (versatile %s)", card.Name, strings.ToLower(card.Rarity)))
		}
	}
	if len(priorities) > 0 {
		priorityStr := ""
		for i, p := range priorities {
			if i < 3 {
				priorityStr += fmt.Sprintf("%d) %s", i+1, p)
				if i < len(priorities)-1 && i < 2 {
					priorityStr += ", "
				}
			}
		}
		details = append(details, fmt.Sprintf("Upgrade priority: %s", priorityStr))
	}

	// Overleveling impact
	for _, card := range deckCards {
		if card.Role != nil && *card.Role == deck.RoleSpellBig {
			details = append(details, fmt.Sprintf("Overleveling impact: %s breakpoints critical vs. support troops",
				card.Name))
			break
		}
	}

	// F2P assessment
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

	// Gold efficiency
	goldEfficiency := int(f2pScore.Score * 10)
	details = append(details, fmt.Sprintf("Gold efficiency: %d/100 - %s upgrade costs",
		goldEfficiency, f2pAssessment))

	// Generate summary
	summary := "Moderate F2P-friendliness"
	if f2pScore.Score >= 8.0 {
		summary = "Excellent F2P deck with clear upgrade path"
	} else if f2pScore.Score < 5.0 {
		summary = "Expensive deck requiring significant investment"
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
// Phase 5: Main Orchestrator
// ============================================================================

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

	// Phase 2: Archetype Detection
	archetypeResult := DetectArchetype(deckCards)

	// Phase 3: Build Analysis Sections
	defenseAnalysis := BuildDefenseAnalysis(deckCards)
	attackAnalysis := BuildAttackAnalysis(deckCards)
	baitAnalysis := BuildBaitAnalysis(deckCards)
	cycleAnalysis := BuildCycleAnalysis(deckCards)
	ladderAnalysis := BuildLadderAnalysis(deckCards)

	// Phase 4: Calculate Overall Score (weighted average)
	// Weights: Attack 20%, Defense 20%, Synergy 25%, Versatility 20%, F2P 15%
	overallScore := (attackScore.Score * 0.20) +
		(defenseScore.Score * 0.20) +
		(synergyScore.Score * 0.25) +
		(versatilityScore.Score * 0.20) +
		(f2pScore.Score * 0.15)

	overallRating := ScoreToRating(overallScore)

	// Build synergy matrix (if database provided)
	var synergyMatrix SynergyMatrix
	if synergyDB != nil {
		synergyAnalysis := synergyDB.AnalyzeDeckSynergy(deckNames)
		if synergyAnalysis != nil && len(synergyAnalysis.TopSynergies) > 0 {
			maxPairs := 28 // C(8,2)
			coverage := float64(len(synergyAnalysis.TopSynergies)) / float64(maxPairs) * 100.0

			synergyMatrix = SynergyMatrix{
				Pairs:            synergyAnalysis.TopSynergies,
				TotalScore:       synergyScore.Score,
				AverageSynergy:   synergyAnalysis.AverageScore,
				PairCount:        len(synergyAnalysis.TopSynergies),
				MaxPossiblePairs: maxPairs,
				SynergyCoverage:  coverage,
			}
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

		OverallScore:  overallScore,
		OverallRating: overallRating,

		DetectedArchetype:   archetypeResult.Primary,
		ArchetypeConfidence: archetypeResult.PrimaryConfidence,

		DefenseAnalysis: defenseAnalysis,
		AttackAnalysis:  attackAnalysis,
		BaitAnalysis:    baitAnalysis,
		CycleAnalysis:   cycleAnalysis,
		LadderAnalysis:  ladderAnalysis,

		SynergyMatrix: synergyMatrix,
	}
}
