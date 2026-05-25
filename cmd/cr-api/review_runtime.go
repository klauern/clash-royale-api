package main

import (
	"context"
	"fmt"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/budget"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/urfave/cli/v3"
)

// reviewReport is the view-model assembled by runReview and rendered by review_format.go.
type reviewReport struct {
	Player    *clashroyale.Player
	Playstyle *analysis.PlaystyleAnalysis

	TopArchetype *analysis.DetectedArchetype
	// CrossArchUpgrades contains the top-3 cross-archetype upgrade priorities.
	CrossArchUpgrades []analysis.CardArchetypeImpact

	// BudgetDecks contains decks within the next-20k-gold budget.
	BudgetDecks []*budget.DeckBudgetAnalysis
}

func runReview(ctx context.Context, cmd *cli.Command) (*reviewReport, error) {
	tag := cmd.String("tag")
	apiToken := cmd.String("api-token")
	dataDir := cmd.String("data-dir")
	verbose := cmd.Bool("verbose")

	result, err := loadOnlinePlayerAnalysis(ctx, tag, apiToken, verbose)
	if err != nil {
		return nil, err
	}

	playstyle, err := analysis.AnalyzePlaystyle(result.Player)
	if err != nil {
		return nil, fmt.Errorf("playstyle analysis failed: %w", err)
	}

	archetypeResult, err := detectTopArchetypes(ctx, dataDir, result)
	if err != nil {
		return nil, fmt.Errorf("archetype detection failed: %w", err)
	}

	budgetResult, err := runBudgetFinder(dataDir, result)
	if err != nil {
		return nil, fmt.Errorf("budget analysis failed: %w", err)
	}

	report := &reviewReport{
		Player:    result.Player,
		Playstyle: playstyle,
	}

	if len(archetypeResult.DetectedArchetypes) > 0 {
		top := archetypeResult.DetectedArchetypes[0]
		report.TopArchetype = &top
	}

	const maxCrossArch = 3
	if len(archetypeResult.TopUpgradeImpacts) > maxCrossArch {
		report.CrossArchUpgrades = archetypeResult.TopUpgradeImpacts[:maxCrossArch]
	} else {
		report.CrossArchUpgrades = archetypeResult.TopUpgradeImpacts
	}

	if budgetResult != nil {
		report.BudgetDecks = budgetResult.WithinBudget
	}

	return report, nil
}

func detectTopArchetypes(_ context.Context, dataDir string, result *onlinePlayerAnalysisResult) (*analysis.DynamicArchetypeAnalysis, error) {
	synergyDB := deck.NewSynergyDatabase()
	detector, err := analysis.NewDynamicArchetypeDetector(dataDir, "", &synergyDBAdapter{db: synergyDB}, &deckStrategyProvider{})
	if err != nil {
		return nil, err
	}

	return detector.DetectArchetypes(result.CardAnalysis, analysis.DetectionOptions{
		TopUpgradesPerArch:   3,
		TopCrossArchUpgrades: 10,
		IncludeUpgrades:      true,
	})
}

func runBudgetFinder(dataDir string, result *onlinePlayerAnalysisResult) (*budget.BudgetFinderResult, error) {
	finder := budget.NewFinder(dataDir, budget.BudgetFinderOptions{
		MaxGoldNeeded: 20000,
		SortBy:        budget.SortByROI,
	})
	return finder.FindOptimalDecks(result.DeckCardAnalysis, result.Player.Tag, result.Player.Name)
}
