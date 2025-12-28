// Package budget provides budget-optimized deck finding functionality.
// It analyzes deck variations to find the most competitive decks with minimal upgrade investment.
package budget

import (
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// DeckBudgetAnalysis represents a deck analyzed for budget optimization
type DeckBudgetAnalysis struct {
	// Deck is the deck recommendation being analyzed
	Deck *deck.DeckRecommendation `json:"deck"`

	// Current score of the deck with existing card levels
	CurrentScore float64 `json:"current_score"`

	// Projected score after upgrades (based on target policy)
	ProjectedScore float64 `json:"projected_score"`

	// Total cards needed across all upgrades
	TotalCardsNeeded int `json:"total_cards_needed"`

	// Total gold needed across all upgrades
	TotalGoldNeeded int `json:"total_gold_needed"`

	// ROI (Return on Investment) = ScoreImprovement / CardsNeeded
	// Higher is better - measures efficiency of investment
	ROI float64 `json:"roi"`

	// Cost efficiency score = CurrentScore / (1 + log(TotalCardsNeeded))
	// Higher is better for budget-conscious players
	CostEfficiency float64 `json:"cost_efficiency"`

	// UpgradesNeeded is the number of card upgrades required
	UpgradesNeeded int `json:"upgrades_needed"`

	// CardUpgrades contains details for each card upgrade
	CardUpgrades []CardUpgradeDetail `json:"card_upgrades"`

	// IsQuickWin indicates deck is 1-2 upgrades away from viable
	IsQuickWin bool `json:"is_quick_win"`

	// ViabilityGap measures how far from "viable" (avg level 12+)
	ViabilityGap float64 `json:"viability_gap"`

	// BudgetCategory classifies the deck by cost
	BudgetCategory BudgetCategory `json:"budget_category"`
}

// CardUpgradeDetail contains upgrade information for a single card
type CardUpgradeDetail struct {
	CardName     string  `json:"card_name"`
	CurrentLevel int     `json:"current_level"`
	TargetLevel  int     `json:"target_level"`
	CardsNeeded  int     `json:"cards_needed"`
	GoldNeeded   int     `json:"gold_needed"`
	Priority     float64 `json:"priority"` // Higher = should upgrade first
}

// BudgetCategory classifies decks by upgrade cost
type BudgetCategory string

const (
	// CategoryReady means deck is already competitive (minimal upgrades needed)
	CategoryReady BudgetCategory = "ready"

	// CategoryQuickWin means deck needs 1-2 small upgrades
	CategoryQuickWin BudgetCategory = "quick_win"

	// CategoryMediumInvestment means deck needs moderate investment
	CategoryMediumInvestment BudgetCategory = "medium_investment"

	// CategoryLongTerm means deck requires significant investment
	CategoryLongTerm BudgetCategory = "long_term"
)

// BudgetFinderOptions configures the budget deck finder
type BudgetFinderOptions struct {
	// MaxCardsNeeded filters out decks needing more than this many cards
	// 0 = no limit
	MaxCardsNeeded int `json:"max_cards_needed"`

	// MaxGoldNeeded filters out decks needing more than this much gold
	// 0 = no limit
	MaxGoldNeeded int `json:"max_gold_needed"`

	// TargetAverageLevel is the minimum average level for "viable" deck
	// Default is 12 (competitive for mid-ladder)
	TargetAverageLevel float64 `json:"target_average_level"`

	// QuickWinMaxUpgrades defines max upgrades for "quick win" classification
	// Default is 2
	QuickWinMaxUpgrades int `json:"quick_win_max_upgrades"`

	// QuickWinMaxCards defines max cards needed for "quick win" classification
	// Default is 1000
	QuickWinMaxCards int `json:"quick_win_max_cards"`

	// SortBy determines how to sort results
	SortBy SortCriteria `json:"sort_by"`

	// TopN limits results to top N decks
	// 0 = return all
	TopN int `json:"top_n"`

	// IncludeVariations generates deck variations with card swaps
	IncludeVariations bool `json:"include_variations"`

	// MaxVariations limits the number of variations per base deck
	MaxVariations int `json:"max_variations"`
}

// SortCriteria determines how to sort budget analysis results
type SortCriteria string

const (
	// SortByROI sorts by return on investment (score improvement / cards needed)
	SortByROI SortCriteria = "roi"

	// SortByCostEfficiency sorts by cost efficiency score
	SortByCostEfficiency SortCriteria = "cost_efficiency"

	// SortByTotalCards sorts by total cards needed (ascending)
	SortByTotalCards SortCriteria = "total_cards"

	// SortByTotalGold sorts by total gold needed (ascending)
	SortByTotalGold SortCriteria = "total_gold"

	// SortByCurrentScore sorts by current deck score (descending)
	SortByCurrentScore SortCriteria = "current_score"

	// SortByProjectedScore sorts by projected score (descending)
	SortByProjectedScore SortCriteria = "projected_score"
)

// BudgetFinderResult contains the results of a budget optimization analysis
type BudgetFinderResult struct {
	// PlayerTag is the player being analyzed
	PlayerTag string `json:"player_tag"`

	// PlayerName is the player's display name
	PlayerName string `json:"player_name"`

	// AllDecks contains all analyzed decks sorted by the chosen criteria
	AllDecks []*DeckBudgetAnalysis `json:"all_decks"`

	// BestROIDecks contains decks with best return on investment
	BestROIDecks []*DeckBudgetAnalysis `json:"best_roi_decks"`

	// QuickWins contains decks that are 1-2 upgrades away from viable
	QuickWins []*DeckBudgetAnalysis `json:"quick_wins"`

	// ReadyDecks contains decks that are already competitive
	ReadyDecks []*DeckBudgetAnalysis `json:"ready_decks"`

	// WithinBudget contains decks within the specified budget constraints
	WithinBudget []*DeckBudgetAnalysis `json:"within_budget"`

	// Summary contains aggregate statistics
	Summary BudgetSummary `json:"summary"`
}

// BudgetSummary contains aggregate statistics for the analysis
type BudgetSummary struct {
	TotalDecksAnalyzed  int     `json:"total_decks_analyzed"`
	ReadyDeckCount      int     `json:"ready_deck_count"`
	QuickWinCount       int     `json:"quick_win_count"`
	AverageCardsNeeded  int     `json:"average_cards_needed"`
	AverageGoldNeeded   int     `json:"average_gold_needed"`
	BestROI             float64 `json:"best_roi"`
	BestCostEfficiency  float64 `json:"best_cost_efficiency"`
	LowestCardsNeeded   int     `json:"lowest_cards_needed"`
	PlayerAverageLevel  float64 `json:"player_average_level"`
}

// DefaultOptions returns default options for budget finder
func DefaultOptions() BudgetFinderOptions {
	return BudgetFinderOptions{
		MaxCardsNeeded:      0, // No limit
		MaxGoldNeeded:       0, // No limit
		TargetAverageLevel:  12.0,
		QuickWinMaxUpgrades: 2,
		QuickWinMaxCards:    1000,
		SortBy:              SortByROI,
		TopN:                10,
		IncludeVariations:   false,
		MaxVariations:       5,
	}
}
