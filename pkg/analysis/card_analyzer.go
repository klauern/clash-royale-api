package analysis

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// AnalyzeCardCollection transforms player card data into comprehensive analysis
func AnalyzeCardCollection(player *clashroyale.Player, options AnalysisOptions) (*CardAnalysis, error) {
	if player == nil {
		return nil, errors.New("player cannot be nil")
	}

	// Convert API cards to UpgradeInfo slice
	upgradeInfos := convertCardsToUpgradeInfos(player.Cards)

	// Apply options filtering
	upgradeInfos = filterUpgradeInfos(upgradeInfos, options)

	// Build CardLevels map
	cardLevels := buildCardLevelsMap(upgradeInfos)

	// Calculate rarity breakdown
	rarityBreakdown := calculateRarityBreakdown(upgradeInfos)

	// Generate upgrade priorities
	upgradePriorities := generateUpgradePriorities(upgradeInfos, options)

	// Create analysis
	analysis := &CardAnalysis{
		PlayerTag:       player.Tag,
		PlayerName:      player.Name,
		AnalysisTime:    time.Now(),
		TotalCards:      len(cardLevels),
		CardLevels:      cardLevels,
		RarityBreakdown: rarityBreakdown,
		UpgradePriority: upgradePriorities,
	}

	// Populate max level cards list
	analysis.MaxLevelCards = make([]string, 0)
	for name, card := range analysis.CardLevels {
		if card.IsMaxLevel {
			analysis.MaxLevelCards = append(analysis.MaxLevelCards, name)
		}
	}

	// Calculate summary
	analysis.CalculateSummary()

	// Validate the analysis
	if err := analysis.Validate(); err != nil {
		return nil, fmt.Errorf("analysis validation failed: %w", err)
	}

	return analysis, nil
}

// convertCardsToUpgradeInfos converts API cards to UpgradeInfo slice
func convertCardsToUpgradeInfos(cards []clashroyale.Card) []UpgradeInfo {
	infos := make([]UpgradeInfo, 0, len(cards))
	for _, card := range cards {
		// Calculate absolute level based on rarity starting level
		// API provides relative levels (e.g., Epic starts at 1? No, usually 6-relative offset)
		// Usually API Level + StartingLevel - 1 = Absolute Level
		// e.g. Common 1 + 1 - 1 = 1
		// e.g. Epic 1 + 6 - 1 = 6 (if API level 1 is start)
		// Let's assume API returns relative level starting from 1 for the card's rarity
		
		rarity := NormalizeRarity(card.Rarity)
		startingLevel := GetStartingLevel(rarity)
		
		// If API level is 0, it might mean level 1? Usually API uses 1-based levels matching the in-game display for that rarity relative to 1
		// Actually for non-commons, if API says level 6 for Epic, and Epic starts at 6, is it Absolute 6?
		// User observed: Witch (Epic) API Level 6 corresponds to a card that needs 50 to upgrade (Level 10->11) or 30/100 (Level 11->12)?
		// If Skeleton Army (Epic) API Level 5 corresponds to Level 10 (Cost 50), then offset is:
		// 10 - 5 = 5.
		// Epic Starting Level is 6. So 6 - 1 = 5.
		// Formula: AbsLevel = APILevel + StartingLevel - 1
		
		absLevel := card.Level + startingLevel - 1
		
		// Adjust max level similarly
		// If MaxLevel is 11 (from API for Epic), AbsMax = 11 + 6 - 1 = 16
		absMaxLevel := card.MaxLevel + startingLevel - 1
		if card.MaxLevel == 0 {
			absMaxLevel = 0 // Let calculator decide default
		}

		info := CalculateUpgradeInfo(
			card.Name,
			card.Rarity,
			card.ElixirCost,
			absLevel,
			card.Count,
			card.EvolutionLevel,
			card.MaxEvolutionLevel,
			absMaxLevel,
		)
		infos = append(infos, info)
	}
	return infos
}

// buildCardLevelsMap builds CardLevelInfo map from UpgradeInfo slice
func buildCardLevelsMap(infos []UpgradeInfo) map[string]CardLevelInfo {
	cardLevels := make(map[string]CardLevelInfo)
	for _, info := range infos {
		cardLevels[info.CardName] = CardLevelInfo{
			Name:              info.CardName,
			ID:                0, // Not available in UpgradeInfo
			Level:             info.CurrentLevel,
			MaxLevel:          info.MaxLevel,
			EvolutionLevel:    info.EvolutionLevel,
			Rarity:            info.Rarity,
			Elixir:            info.ElixirCost,
			CardCount:         info.CardsOwned,
			CardsToNext:       info.CardsToNextLevel,
			IsMaxLevel:        info.IsMaxLevel,
			MaxEvolutionLevel: info.MaxEvolutionLevel,
		}
	}
	return cardLevels
}

// calculateRarityBreakdown calculates rarity statistics from UpgradeInfo slice
func calculateRarityBreakdown(infos []UpgradeInfo) map[string]RarityStats {
	// Group cards by rarity
	rarityGroups := make(map[string][]UpgradeInfo)
	for _, info := range infos {
		rarityGroups[info.Rarity] = append(rarityGroups[info.Rarity], info)
	}

	breakdown := make(map[string]RarityStats)
	for rarity, cards := range rarityGroups {
		totalLevel := 0
		totalLevelRatio := 0.0
		maxLevelCount := 0
		cardsNearMax := 0
		cardsReadyUpgrade := 0

		for _, card := range cards {
			totalLevel += card.CurrentLevel
			totalLevelRatio += float64(card.CurrentLevel) / float64(card.MaxLevel)

			if card.IsMaxLevel {
				maxLevelCount++
			}

			// Cards within 1-2 levels of max
			if card.MaxLevel-card.CurrentLevel <= 2 {
				cardsNearMax++
			}

			// Cards ready to upgrade now
			if card.CanUpgradeNow {
				cardsReadyUpgrade++
			}
		}

		cardCount := len(cards)
		breakdown[rarity] = RarityStats{
			Rarity:            rarity,
			TotalCards:        cardCount,
			TotalPossible:     totalCardsPerRarity[rarity],
			MaxLevelCards:     maxLevelCount,
			AvgLevel:          float64(totalLevel) / float64(cardCount),
			AvgLevelRatio:     totalLevelRatio / float64(cardCount),
			CardsNearMax:      cardsNearMax,
			CardsReadyUpgrade: cardsReadyUpgrade,
		}
	}

	return breakdown
}

// generateUpgradePriorities generates upgrade priority list from UpgradeInfo slice
func generateUpgradePriorities(infos []UpgradeInfo, options AnalysisOptions) []UpgradePriority {
	// Filter out max level cards if not included
	filteredInfos := make([]UpgradeInfo, 0, len(infos))
	for _, info := range infos {
		if info.IsMaxLevel && !options.IncludeMaxLevel {
			continue
		}
		filteredInfos = append(filteredInfos, info)
	}

	// Get upgrade priorities using existing calculator
	priorityInfos := GetUpgradePriorities(filteredInfos, options.MinPriorityScore, options.TopN)

	// Convert to UpgradePriority type
	priorities := make([]UpgradePriority, 0, len(priorityInfos))
	for _, info := range priorityInfos {
		priority := UpgradePriority{
			CardName:      info.CardName,
			Rarity:        info.Rarity,
			CurrentLevel:  info.CurrentLevel,
			MaxLevel:      info.MaxLevel,
			CardsOwned:    info.CardsOwned,
			CardsRequired: info.CardsToNextLevel,
			CardsNeeded:   info.CardsRemaining,
			PriorityScore: CalculatePriorityScore(info),
			Reasons:       calculatePriorityReasons(info),
		}

		// Determine priority level based on score
		switch {
		case priority.PriorityScore >= 70.0:
			priority.Priority = "high"
		case priority.PriorityScore >= 40.0:
			priority.Priority = "medium"
		default:
			priority.Priority = "low"
		}

		// Apply focus rarities filter
		if len(options.FocusRarities) > 0 && !contains(options.FocusRarities, info.Rarity) {
			continue
		}

		// Apply exclude cards filter
		if contains(options.ExcludeCards, info.CardName) {
			continue
		}

		priorities = append(priorities, priority)
	}

	return priorities
}

// calculatePriorityReasons generates human-readable reasons for upgrade priority
func calculatePriorityReasons(info UpgradeInfo) []string {
	reasons := make([]string, 0)

	if info.CanUpgradeNow {
		reasons = append(reasons, "Ready to upgrade now")
	}

	if info.ProgressPercent >= 90.0 {
		reasons = append(reasons, "Nearly enough cards")
	} else if info.ProgressPercent >= 50.0 {
		reasons = append(reasons, "Good progress")
	}

	if info.Rarity == "Legendary" || info.Rarity == "Champion" {
		reasons = append(reasons, "Rare card type")
	}

	if info.CurrentLevel >= 12 {
		reasons = append(reasons, "High level")
	}

	return reasons
}

// filterUpgradeInfos applies options filtering to UpgradeInfo slice
func filterUpgradeInfos(infos []UpgradeInfo, options AnalysisOptions) []UpgradeInfo {
	if len(options.FocusRarities) == 0 && len(options.ExcludeCards) == 0 {
		return infos
	}

	filtered := make([]UpgradeInfo, 0, len(infos))
	for _, info := range infos {
		// Apply focus rarities filter
		if len(options.FocusRarities) > 0 && !contains(options.FocusRarities, info.Rarity) {
			continue
		}

		// Apply exclude cards filter
		if contains(options.ExcludeCards, info.CardName) {
			continue
		}

		filtered = append(filtered, info)
	}
	return filtered
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, value) {
			return true
		}
	}
	return false
}
