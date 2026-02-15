// Package analysis provides upgrade impact analysis for Clash Royale cards.
// The upgrade impact analyzer determines which single card upgrade has the
// biggest impact on deck viability across all possible decks.
package analysis

import (
	"slices"
	"sort"
	"time"
)

// UpgradeImpactAnalysis represents the complete upgrade impact analysis result
type UpgradeImpactAnalysis struct {
	PlayerTag    string                `json:"player_tag"`
	PlayerName   string                `json:"player_name"`
	AnalysisTime time.Time             `json:"analysis_time"`
	CardImpacts  []CardUpgradeImpact   `json:"card_impacts"`
	KeyCards     []KeyCardInfo         `json:"key_cards"`   // Cards that unlock multiple archetypes
	UnlockTree   []ArchetypeUnlockInfo `json:"unlock_tree"` // How upgrades open new deck options
	TopImpacts   []CardUpgradeImpact   `json:"top_impacts"` // Most impactful upgrades (top N)
	Summary      ImpactSummary         `json:"summary"`
}

// CardUpgradeImpact represents the impact analysis for upgrading a single card
type CardUpgradeImpact struct {
	CardName            string              `json:"card_name"`
	Rarity              string              `json:"rarity"`
	CurrentLevel        int                 `json:"current_level"`
	MaxLevel            int                 `json:"max_level"`
	UpgradedLevel       int                 `json:"upgraded_level"` // Level after +1 upgrade
	Elixir              int                 `json:"elixir"`
	Role                string              `json:"role,omitempty"`
	ViableDecksCount    int                 `json:"viable_decks_count"`           // Number of viable decks containing this card
	AvgScoreImprovement float64             `json:"avg_score_improvement"`        // Average deck score improvement if upgraded
	MaxScoreImprovement float64             `json:"max_score_improvement"`        // Max improvement in any deck
	UnlockPotential     int                 `json:"unlock_potential"`             // Decks that become viable with this upgrade
	ImpactScore         float64             `json:"impact_score"`                 // Overall impact ranking score
	AffectedDecks       []DeckImpactSummary `json:"affected_decks,omitempty"`     // Decks affected by this upgrade
	GoldCost            int                 `json:"gold_cost"`                    // Gold needed for +1 upgrade
	ValuePerGold        float64             `json:"value_per_gold"`               // Impact score per gold spent
	IsKeyCard           bool                `json:"is_key_card"`                  // Unlocks multiple archetypes
	UnlocksArchetypes   []string            `json:"unlocks_archetypes,omitempty"` // Which archetypes this upgrade helps
}

// DeckImpactSummary represents how a card upgrade affects a specific deck
type DeckImpactSummary struct {
	DeckName       string  `json:"deck_name"`
	WinCondition   string  `json:"win_condition"`
	CurrentScore   float64 `json:"current_score"`
	ProjectedScore float64 `json:"projected_score"`
	ScoreDelta     float64 `json:"score_delta"`
	BecomesViable  bool    `json:"becomes_viable"` // True if this upgrade makes the deck viable
	CardRole       string  `json:"card_role,omitempty"`
}

// KeyCardInfo represents a card that is key to unlocking multiple deck archetypes
type KeyCardInfo struct {
	CardName           string   `json:"card_name"`
	Rarity             string   `json:"rarity"`
	CurrentLevel       int      `json:"current_level"`
	UnlockedArchetypes []string `json:"unlocked_archetypes"` // Archetypes enabled by this card
	DeckUnlockCount    int      `json:"deck_unlock_count"`   // Number of decks this card unlocks
	ImpactScore        float64  `json:"impact_score"`
}

// ArchetypeUnlockInfo shows how upgrades open new deck archetypes
type ArchetypeUnlockInfo struct {
	Archetype        string   `json:"archetype"`
	CurrentViability string   `json:"current_viability"` // "viable", "marginal", "blocked"
	UpgradesNeeded   []string `json:"upgrades_needed"`   // Cards that need upgrading to unlock
	PriorityUpgrade  string   `json:"priority_upgrade"`  // Most impactful single upgrade
	EstimatedGold    int      `json:"estimated_gold"`    // Gold to unlock this archetype
}

// ImpactSummary provides high-level overview of upgrade impact analysis
type ImpactSummary struct {
	TotalCardsAnalyzed int     `json:"total_cards_analyzed"`
	KeyCardsIdentified int     `json:"key_cards_identified"`
	AvgImpactScore     float64 `json:"avg_impact_score"`
	MaxImpactScore     float64 `json:"max_impact_score"`
	TotalViableDecks   int     `json:"total_viable_decks"`
	PotentialUnlocks   int     `json:"potential_unlocks"` // Decks that could become viable
}

// UpgradeImpactOptions configures the upgrade impact analysis
type UpgradeImpactOptions struct {
	ViabilityThreshold float64  `json:"viability_threshold"` // Minimum deck score to be considered viable
	TopN               int      `json:"top_n"`               // Number of top impacts to return
	IncludeMaxLevel    bool     `json:"include_max_level"`   // Include already-maxed cards
	FocusRarities      []string `json:"focus_rarities"`      // Filter to specific rarities
	ExcludeCards       []string `json:"exclude_cards"`       // Cards to exclude from analysis
	UseCombatStats     bool     `json:"use_combat_stats"`    // Enable combat stats integration (DPS/HP scoring)
}

// DefaultUpgradeImpactOptions returns sensible defaults for upgrade impact analysis
func DefaultUpgradeImpactOptions() UpgradeImpactOptions {
	return UpgradeImpactOptions{
		ViabilityThreshold: 0.75, // 75% of max possible score is "viable"
		TopN:               10,
		IncludeMaxLevel:    false,
		FocusRarities:      []string{},
		ExcludeCards:       []string{},
	}
}

// UpgradeImpactAnalyzer performs upgrade impact analysis
type UpgradeImpactAnalyzer struct {
	archetypes         []DeckArchetypeTemplate
	options            UpgradeImpactOptions
	dataDir            string
	archetypesFilePath string // Path to custom archetypes file (empty if using defaults)
}

// DeckArchetypeTemplate represents a deck archetype for analysis
type DeckArchetypeTemplate struct {
	Name              string   `json:"name"`
	WinCondition      string   `json:"win_condition"`
	RequiredCards     []string `json:"required_cards,omitempty"`
	SupportCards      []string `json:"support_cards,omitempty"`
	MinElixir         float64  `json:"min_elixir"`
	MaxElixir         float64  `json:"max_elixir"`
	Category          string   `json:"category,omitempty"`           // Archetype category (beatdown, cycle, siege, etc.)
	Enabled           bool     `json:"enabled"`                      // Whether this archetype is enabled for analysis
	PreferredStrategy string   `json:"preferred_strategy,omitempty"` // Recommended deck builder strategy
}

// NewUpgradeImpactAnalyzer creates a new upgrade impact analyzer.
// If archetypesFilePath is empty, uses embedded default archetypes.
// Otherwise loads archetypes from the specified JSON file.
func NewUpgradeImpactAnalyzer(dataDir, archetypesFilePath string, options UpgradeImpactOptions) (*UpgradeImpactAnalyzer, error) {
	archetypes, err := LoadArchetypes(archetypesFilePath)
	if err != nil {
		return nil, err
	}

	return &UpgradeImpactAnalyzer{
		archetypes:         archetypes,
		options:            options,
		dataDir:            dataDir,
		archetypesFilePath: archetypesFilePath,
	}, nil
}

// AnalyzeUpgradeImpact performs comprehensive upgrade impact analysis
func (a *UpgradeImpactAnalyzer) AnalyzeUpgradeImpact(cardAnalysis *CardAnalysis) (*UpgradeImpactAnalysis, error) {
	if cardAnalysis == nil {
		return nil, ErrMissingPlayerTag
	}

	// Calculate impact for each card
	cardImpacts := make([]CardUpgradeImpact, 0, len(cardAnalysis.CardLevels))

	for cardName, cardInfo := range cardAnalysis.CardLevels {
		// Skip if max level and not including max level cards
		if cardInfo.IsMaxLevel && !a.options.IncludeMaxLevel {
			continue
		}

		// Skip if in exclude list
		if a.containsCard(a.options.ExcludeCards, cardName) {
			continue
		}

		// Skip if filtering by rarity and this card doesn't match
		if len(a.options.FocusRarities) > 0 && !a.containsRarity(a.options.FocusRarities, cardInfo.Rarity) {
			continue
		}

		impact := a.calculateCardImpact(cardName, cardInfo, cardAnalysis)
		cardImpacts = append(cardImpacts, impact)
	}

	// Sort by impact score (descending)
	sort.Slice(cardImpacts, func(i, j int) bool {
		return cardImpacts[i].ImpactScore > cardImpacts[j].ImpactScore
	})

	// Get top N impacts
	topN := min(a.options.TopN, len(cardImpacts))
	topImpacts := make([]CardUpgradeImpact, topN)
	copy(topImpacts, cardImpacts[:topN])

	// Identify key cards (those that unlock multiple archetypes)
	keyCards := a.identifyKeyCards(cardImpacts)

	// Build unlock tree
	unlockTree := a.buildUnlockTree(cardImpacts, cardAnalysis)

	// Calculate summary
	summary := a.calculateSummary(cardImpacts, keyCards)

	return &UpgradeImpactAnalysis{
		PlayerTag:    cardAnalysis.PlayerTag,
		PlayerName:   cardAnalysis.PlayerName,
		AnalysisTime: time.Now(),
		CardImpacts:  cardImpacts,
		KeyCards:     keyCards,
		UnlockTree:   unlockTree,
		TopImpacts:   topImpacts,
		Summary:      summary,
	}, nil
}

// calculateCardImpact calculates the upgrade impact for a single card
func (a *UpgradeImpactAnalyzer) calculateCardImpact(
	cardName string,
	cardInfo CardLevelInfo,
	cardAnalysis *CardAnalysis,
) CardUpgradeImpact {
	// Calculate upgraded level (capped at max)
	upgradedLevel := min(cardInfo.Level+1, cardInfo.MaxLevel)

	// Calculate current and upgraded card scores (evolution-aware)
	currentScore := a.scoreCard(cardName, cardInfo.Level, cardInfo.MaxLevel, cardInfo.Rarity, cardInfo.Elixir, cardInfo.EvolutionLevel, cardInfo.MaxEvolutionLevel)
	upgradedScore := a.scoreCard(cardName, upgradedLevel, cardInfo.MaxLevel, cardInfo.Rarity, cardInfo.Elixir, cardInfo.EvolutionLevel, cardInfo.MaxEvolutionLevel)
	scoreDelta := upgradedScore - currentScore

	// Analyze affected decks
	affectedDecks := a.analyzeAffectedDecks(cardName, cardInfo, cardAnalysis)

	viableDeckCount := 0
	unlockPotential := 0
	avgImprovement := 0.0
	maxImprovement := 0.0

	for _, deckImpact := range affectedDecks {
		if deckImpact.ScoreDelta > maxImprovement {
			maxImprovement = deckImpact.ScoreDelta
		}
		avgImprovement += deckImpact.ScoreDelta

		if deckImpact.ProjectedScore >= a.options.ViabilityThreshold {
			viableDeckCount++
		}

		if deckImpact.BecomesViable {
			unlockPotential++
		}
	}

	if len(affectedDecks) > 0 {
		avgImprovement /= float64(len(affectedDecks))
	}

	// Calculate gold cost for upgrade
	goldCost := a.getGoldForSingleUpgrade(cardInfo.Level, cardInfo.Rarity)

	// Calculate impact score
	// Formula: weighted combination of:
	// - Score delta (direct impact on card strength) (30%)
	// - Average deck improvement (30%)
	// - Unlock potential (25%)
	// - Role importance (15%)
	roleImportance := a.getRoleImportance(cardName)

	impactScore := (scoreDelta * 30.0 * 100) +
		(avgImprovement * 30.0 * 100) +
		(float64(unlockPotential) * 25.0 / float64(max(len(a.archetypes), 1)) * 100) +
		(roleImportance * 15.0)

	// Calculate value per gold
	valuePerGold := 0.0
	if goldCost > 0 {
		valuePerGold = impactScore / float64(goldCost) * 1000 // Per 1000 gold
	}

	// Determine role
	role := a.inferRole(cardName)

	// Check if this is a key card (unlocks multiple archetypes)
	unlocksArchetypes := a.getUnlockedArchetypes(cardName, cardInfo, cardAnalysis)
	isKeyCard := len(unlocksArchetypes) >= 2

	return CardUpgradeImpact{
		CardName:            cardName,
		Rarity:              cardInfo.Rarity,
		CurrentLevel:        cardInfo.Level,
		MaxLevel:            cardInfo.MaxLevel,
		UpgradedLevel:       upgradedLevel,
		Elixir:              cardInfo.Elixir,
		Role:                role,
		ViableDecksCount:    viableDeckCount,
		AvgScoreImprovement: avgImprovement,
		MaxScoreImprovement: maxImprovement,
		UnlockPotential:     unlockPotential,
		ImpactScore:         impactScore,
		AffectedDecks:       affectedDecks,
		GoldCost:            goldCost,
		ValuePerGold:        valuePerGold,
		IsKeyCard:           isKeyCard,
		UnlocksArchetypes:   unlocksArchetypes,
	}
}

// scoreCard calculates card score with evolution awareness.
// This provides more accurate scoring than the legacy calculateCardScore by considering evolution levels.
func (a *UpgradeImpactAnalyzer) scoreCard(cardName string, level, maxLevel int, rarity string, elixir, evolutionLevel, maxEvolutionLevel int) float64 {
	//  Base score calculation (same as before)
	baseScore := a.fallbackScore(level, maxLevel, rarity, elixir)

	// Add evolution bonus if applicable
	evolutionBonus := a.calculateEvolutionBonus(evolutionLevel, maxEvolutionLevel)

	return baseScore + evolutionBonus
}

// fallbackScore provides basic scoring when scorer unavailable.
// This preserves the original calculateCardScore logic for backward compatibility.
func (a *UpgradeImpactAnalyzer) fallbackScore(level, maxLevel int, rarity string, elixir int) float64 {
	if maxLevel == 0 {
		maxLevel = 14
	}

	levelRatio := float64(level) / float64(maxLevel)

	// Rarity boost
	rarityBoost := 1.0
	switch rarity {
	case "Rare":
		rarityBoost = 1.05
	case "Epic":
		rarityBoost = 1.1
	case "Legendary":
		rarityBoost = 1.15
	case "Champion":
		rarityBoost = 1.2
	}

	// Elixir efficiency
	elixirWeight := 1.0 - float64(max(elixir-3, 0))/9.0

	return (levelRatio * 1.2 * rarityBoost) + (elixirWeight * 0.15)
}

// calculateEvolutionBonus calculates the evolution level bonus for a card.
// The bonus is proportional to the evolution progress (evolutionLevel/maxEvolutionLevel).
// This matches the deck builder's evolution bonus calculation for consistency.
func (a *UpgradeImpactAnalyzer) calculateEvolutionBonus(evolutionLevel, maxEvolutionLevel int) float64 {
	if maxEvolutionLevel <= 0 || evolutionLevel <= 0 {
		return 0.0
	}

	// Calculate evolution ratio (0.0 to 1.0)
	evolutionRatio := float64(evolutionLevel) / float64(maxEvolutionLevel)

	// Clamp ratio to valid range
	if evolutionRatio > 1.0 {
		evolutionRatio = 1.0
	}

	// Evolution bonus weight (matches deck builder: 0.15 max bonus)
	const evolutionBonusWeight = 0.15

	// Apply evolution bonus weight
	return evolutionBonusWeight * evolutionRatio
}

// analyzeAffectedDecks analyzes how upgrading a card affects various deck archetypes
func (a *UpgradeImpactAnalyzer) analyzeAffectedDecks(
	cardName string,
	cardInfo CardLevelInfo,
	cardAnalysis *CardAnalysis,
) []DeckImpactSummary {
	affected := make([]DeckImpactSummary, 0)

	for _, archetype := range a.archetypes {
		// Check if this card is relevant to this archetype
		isRelevant := cardName == archetype.WinCondition
		if slices.Contains(archetype.SupportCards, cardName) {
			isRelevant = true
		}

		if !isRelevant {
			continue
		}

		// Check if win condition exists in collection
		winConInfo, hasWinCon := cardAnalysis.CardLevels[archetype.WinCondition]
		if !hasWinCon {
			continue
		}

		// Calculate current archetype score
		currentScore := a.calculateArchetypeScore(archetype, cardAnalysis)

		// Calculate projected score with upgrade
		projectedScore := a.calculateArchetypeScoreWithUpgrade(archetype, cardAnalysis, cardName, cardInfo.Level+1)

		scoreDelta := projectedScore - currentScore
		currentlyViable := currentScore >= a.options.ViabilityThreshold
		becomesViable := !currentlyViable && projectedScore >= a.options.ViabilityThreshold

		// Determine card role in this deck
		cardRole := "support"
		if cardName == archetype.WinCondition {
			cardRole = "win_condition"
		}

		affected = append(affected, DeckImpactSummary{
			DeckName:       archetype.Name,
			WinCondition:   archetype.WinCondition,
			CurrentScore:   currentScore,
			ProjectedScore: projectedScore,
			ScoreDelta:     scoreDelta,
			BecomesViable:  becomesViable,
			CardRole:       cardRole,
		})

		// Also consider if this card IS the win condition
		if cardName == archetype.WinCondition && winConInfo.Level < winConInfo.MaxLevel {
			// Already added above
		}
	}

	return affected
}

// calculateArchetypeScore calculates the viability score for an archetype
func (a *UpgradeImpactAnalyzer) calculateArchetypeScore(
	archetype DeckArchetypeTemplate,
	cardAnalysis *CardAnalysis,
) float64 {
	// Win condition contributes 40% of score
	winConScore := 0.0
	if cardInfo, exists := cardAnalysis.CardLevels[archetype.WinCondition]; exists {
		winConScore = a.scoreCard(archetype.WinCondition, cardInfo.Level, cardInfo.MaxLevel, cardInfo.Rarity, cardInfo.Elixir, cardInfo.EvolutionLevel, cardInfo.MaxEvolutionLevel)
	}

	// Support cards contribute 60% of score
	supportScore := 0.0
	supportCount := 0
	for _, supportCard := range archetype.SupportCards {
		if cardInfo, exists := cardAnalysis.CardLevels[supportCard]; exists {
			supportScore += a.scoreCard(supportCard, cardInfo.Level, cardInfo.MaxLevel, cardInfo.Rarity, cardInfo.Elixir, cardInfo.EvolutionLevel, cardInfo.MaxEvolutionLevel)
			supportCount++
		}
	}
	if supportCount > 0 {
		supportScore /= float64(supportCount)
	}

	return (winConScore * 0.4) + (supportScore * 0.6)
}

// calculateArchetypeScoreWithUpgrade calculates archetype score with one card upgraded
func (a *UpgradeImpactAnalyzer) calculateArchetypeScoreWithUpgrade(
	archetype DeckArchetypeTemplate,
	cardAnalysis *CardAnalysis,
	upgradedCard string,
	newLevel int,
) float64 {
	// Win condition contributes 40% of score
	winConScore := 0.0
	if cardInfo, exists := cardAnalysis.CardLevels[archetype.WinCondition]; exists {
		level := cardInfo.Level
		if archetype.WinCondition == upgradedCard {
			level = min(newLevel, cardInfo.MaxLevel)
		}
		winConScore = a.scoreCard(archetype.WinCondition, level, cardInfo.MaxLevel, cardInfo.Rarity, cardInfo.Elixir, cardInfo.EvolutionLevel, cardInfo.MaxEvolutionLevel)
	}

	// Support cards contribute 60% of score
	supportScore := 0.0
	supportCount := 0
	for _, supportCard := range archetype.SupportCards {
		if cardInfo, exists := cardAnalysis.CardLevels[supportCard]; exists {
			level := cardInfo.Level
			if supportCard == upgradedCard {
				level = min(newLevel, cardInfo.MaxLevel)
			}
			supportScore += a.scoreCard(supportCard, level, cardInfo.MaxLevel, cardInfo.Rarity, cardInfo.Elixir, cardInfo.EvolutionLevel, cardInfo.MaxEvolutionLevel)
			supportCount++
		}
	}
	if supportCount > 0 {
		supportScore /= float64(supportCount)
	}

	return (winConScore * 0.4) + (supportScore * 0.6)
}

// getUnlockedArchetypes returns archetypes that this card upgrade would help unlock
func (a *UpgradeImpactAnalyzer) getUnlockedArchetypes(
	cardName string,
	cardInfo CardLevelInfo,
	cardAnalysis *CardAnalysis,
) []string {
	unlocked := make([]string, 0)

	for _, archetype := range a.archetypes {
		// Only consider if this card is the win condition or a key support
		isRelevant := cardName == archetype.WinCondition
		if slices.Contains(archetype.SupportCards, cardName) {
			isRelevant = true
		}

		if !isRelevant {
			continue
		}

		currentScore := a.calculateArchetypeScore(archetype, cardAnalysis)
		projectedScore := a.calculateArchetypeScoreWithUpgrade(archetype, cardAnalysis, cardName, cardInfo.Level+1)

		// If this upgrade pushes the archetype above viability threshold
		if currentScore < a.options.ViabilityThreshold && projectedScore >= a.options.ViabilityThreshold {
			unlocked = append(unlocked, archetype.Name)
		}
	}

	return unlocked
}

// identifyKeyCards identifies cards that unlock multiple deck archetypes
func (a *UpgradeImpactAnalyzer) identifyKeyCards(impacts []CardUpgradeImpact) []KeyCardInfo {
	keyCards := make([]KeyCardInfo, 0)

	for _, impact := range impacts {
		if impact.IsKeyCard || len(impact.UnlocksArchetypes) >= 2 {
			keyCards = append(keyCards, KeyCardInfo{
				CardName:           impact.CardName,
				Rarity:             impact.Rarity,
				CurrentLevel:       impact.CurrentLevel,
				UnlockedArchetypes: impact.UnlocksArchetypes,
				DeckUnlockCount:    impact.UnlockPotential,
				ImpactScore:        impact.ImpactScore,
			})
		}
	}

	// Sort by deck unlock count
	sort.Slice(keyCards, func(i, j int) bool {
		return keyCards[i].DeckUnlockCount > keyCards[j].DeckUnlockCount
	})

	return keyCards
}

// buildUnlockTree builds a tree showing how upgrades open new deck options
func (a *UpgradeImpactAnalyzer) buildUnlockTree(
	impacts []CardUpgradeImpact,
	cardAnalysis *CardAnalysis,
) []ArchetypeUnlockInfo {
	unlockTree := make([]ArchetypeUnlockInfo, 0, len(a.archetypes))

	// For each archetype, determine current viability and upgrades needed
	for _, archetype := range a.archetypes {
		// Check if win condition exists
		cardInfo, exists := cardAnalysis.CardLevels[archetype.WinCondition]
		if !exists {
			unlockTree = append(unlockTree, ArchetypeUnlockInfo{
				Archetype:        archetype.Name,
				CurrentViability: "blocked",
				UpgradesNeeded:   []string{archetype.WinCondition + " (not owned)"},
				PriorityUpgrade:  archetype.WinCondition,
				EstimatedGold:    0,
			})
			continue
		}

		// Determine viability based on archetype score
		currentScore := a.calculateArchetypeScore(archetype, cardAnalysis)
		viability := "viable"
		if currentScore < 0.5 {
			viability = "blocked"
		} else if currentScore < a.options.ViabilityThreshold {
			viability = "marginal"
		}

		// Find upgrades needed
		upgradesNeeded := make([]string, 0)
		priorityUpgrade := ""
		maxImpact := 0.0
		totalGold := 0

		// Check win condition upgrade impact
		if cardInfo.Level < cardInfo.MaxLevel {
			winConScore := a.calculateArchetypeScoreWithUpgrade(archetype, cardAnalysis, archetype.WinCondition, cardInfo.Level+1)
			delta := winConScore - currentScore
			if delta > maxImpact {
				maxImpact = delta
				priorityUpgrade = archetype.WinCondition
			}
			if delta > 0.01 {
				upgradesNeeded = append(upgradesNeeded, archetype.WinCondition)
				totalGold += a.getGoldForSingleUpgrade(cardInfo.Level, cardInfo.Rarity)
			}
		}

		// Check support card upgrade impacts
		for _, supportCard := range archetype.SupportCards {
			if supportInfo, exists := cardAnalysis.CardLevels[supportCard]; exists && supportInfo.Level < supportInfo.MaxLevel {
				supportScore := a.calculateArchetypeScoreWithUpgrade(archetype, cardAnalysis, supportCard, supportInfo.Level+1)
				delta := supportScore - currentScore
				if delta > maxImpact {
					maxImpact = delta
					priorityUpgrade = supportCard
				}
				if delta > 0.01 {
					upgradesNeeded = append(upgradesNeeded, supportCard)
					totalGold += a.getGoldForSingleUpgrade(supportInfo.Level, supportInfo.Rarity)
				}
			}
		}

		// Limit to top 3 upgrades
		if len(upgradesNeeded) > 3 {
			upgradesNeeded = upgradesNeeded[:3]
		}

		unlockTree = append(unlockTree, ArchetypeUnlockInfo{
			Archetype:        archetype.Name,
			CurrentViability: viability,
			UpgradesNeeded:   upgradesNeeded,
			PriorityUpgrade:  priorityUpgrade,
			EstimatedGold:    totalGold,
		})
	}

	return unlockTree
}

// calculateSummary calculates summary statistics
func (a *UpgradeImpactAnalyzer) calculateSummary(
	impacts []CardUpgradeImpact,
	keyCards []KeyCardInfo,
) ImpactSummary {
	if len(impacts) == 0 {
		return ImpactSummary{}
	}

	totalScore := 0.0
	maxScore := 0.0
	totalViable := 0
	totalUnlocks := 0

	for _, impact := range impacts {
		totalScore += impact.ImpactScore
		if impact.ImpactScore > maxScore {
			maxScore = impact.ImpactScore
		}
		totalViable += impact.ViableDecksCount
		totalUnlocks += impact.UnlockPotential
	}

	return ImpactSummary{
		TotalCardsAnalyzed: len(impacts),
		KeyCardsIdentified: len(keyCards),
		AvgImpactScore:     totalScore / float64(len(impacts)),
		MaxImpactScore:     maxScore,
		TotalViableDecks:   totalViable,
		PotentialUnlocks:   totalUnlocks,
	}
}

// Helper functions

func (a *UpgradeImpactAnalyzer) inferRole(cardName string) string {
	// Expanded win conditions list for 28+ archetypes
	winConditions := []string{
		"Royal Giant", "Hog Rider", "Giant", "P.E.K.K.A", "Giant Skeleton",
		"Goblin Barrel", "Mortar", "X-Bow", "Royal Hogs",
		"Golem", "Lava Hound", "Electro Giant", "Balloon",
		"Miner", "Graveyard", "Three Musketeers", "Mega Knight",
		"Royal Ghost", "Goblin Drill", "Skeleton King",
	}
	if slices.Contains(winConditions, cardName) {
		return "win_conditions"
	}

	buildings := []string{
		"Cannon", "Goblin Cage", "Inferno Tower", "Bomb Tower", "Tombstone",
		"Goblin Hut", "Barbarian Hut", "Tesla", "Furnace",
	}
	if slices.Contains(buildings, cardName) {
		return "buildings"
	}

	bigSpells := []string{"Fireball", "Poison", "Lightning", "Rocket", "Earthquake"}
	if slices.Contains(bigSpells, cardName) {
		return "spells_big"
	}

	smallSpells := []string{"Zap", "Arrows", "Giant Snowball", "Barbarian Barrel", "Freeze", "Log", "Tornado"}
	if slices.Contains(smallSpells, cardName) {
		return "spells_small"
	}

	return "support"
}

func (a *UpgradeImpactAnalyzer) getRoleImportance(cardName string) float64 {
	role := a.inferRole(cardName)
	switch role {
	case "win_conditions":
		return 1.0 // Win conditions are most important
	case "buildings":
		return 0.7
	case "spells_big":
		return 0.6
	case "spells_small":
		return 0.5
	default:
		return 0.4
	}
}

func (a *UpgradeImpactAnalyzer) containsCard(cards []string, cardName string) bool {
	return slices.Contains(cards, cardName)
}

func (a *UpgradeImpactAnalyzer) containsRarity(rarities []string, rarity string) bool {
	return slices.Contains(rarities, rarity)
}

// getGoldForSingleUpgrade returns gold needed for a single level upgrade
func (a *UpgradeImpactAnalyzer) getGoldForSingleUpgrade(currentLevel int, rarity string) int {
	goldCosts := map[string]map[int]int{
		"Common": {
			1: 5, 2: 20, 3: 50, 4: 150, 5: 400, 6: 1000, 7: 2000,
			8: 4000, 9: 8000, 10: 20000, 11: 50000, 12: 100000, 13: 100000,
		},
		"Rare": {
			3: 50, 4: 150, 5: 400, 6: 1000, 7: 2000,
			8: 4000, 9: 8000, 10: 20000, 11: 50000, 12: 100000, 13: 100000,
		},
		"Epic": {
			6: 400, 7: 2000, 8: 4000, 9: 8000, 10: 20000, 11: 50000, 12: 100000, 13: 100000,
		},
		"Legendary": {
			9: 5000, 10: 20000, 11: 50000, 12: 100000, 13: 100000,
		},
		"Champion": {
			11: 50000, 12: 100000, 13: 100000,
		},
	}

	costs, exists := goldCosts[rarity]
	if !exists {
		return 0
	}

	goldNeeded, exists := costs[currentLevel]
	if !exists {
		return 0
	}

	return goldNeeded
}
