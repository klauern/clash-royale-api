package main

import (
	"context"
	"fmt"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/budget"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/klauer/clash-royale-api/go/pkg/recommend"
	"github.com/urfave/cli/v3"
)

// reviewReport is the view-model assembled by runReview and rendered by review_format.go.
type reviewReport struct {
	Player    *clashroyale.Player
	Playstyle *analysis.PlaystyleAnalysis

	TopArchetype *analysis.DetectedArchetype
	// CrossArchUpgrades contains the top-3 cross-archetype upgrade priorities.
	CrossArchUpgrades []analysis.CardArchetypeImpact

	// DeckDelta compares the player's current deck against the top recommended deck.
	DeckDelta *evaluation.DeckDiff

	// BudgetDecks contains decks within the next-20k-gold budget.
	BudgetDecks []*budget.DeckBudgetAnalysis

	// SlotAssignments lists the top valid (evo/champion/flex) slot assignments for the current deck.
	// Empty when the deck contains no evolution or champion cards.
	SlotAssignments []recommend.SlotAssignment
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

	delta, err := runDeckDiff(dataDir, result)
	if err != nil {
		// Non-fatal: report without delta rather than failing the whole review.
		printf("warning: deck delta unavailable: %v\n", err)
	}
	report.DeckDelta = delta

	report.SlotAssignments = runSlotAssignments(result)

	return report, nil
}

// runSlotAssignments enumerates the top valid evo/champion/flex slot assignments
// for the player's current deck using DefaultSlotScorer.
func runSlotAssignments(result *onlinePlayerAnalysisResult) []recommend.SlotAssignment {
	typeer := recommend.NewCardSlotTyperFromDeckAnalysis(result.DeckCardAnalysis)
	cards := make([]recommend.CardDetailLike, len(result.Player.CurrentDeck))
	for i, c := range result.Player.CurrentDeck {
		cards[i] = recommend.CardDetailLike{CardName: c.Name, Evolved: c.EvolutionLevel > 0}
	}
	classified := typeer.ClassifyCards(cards)
	candidates := recommend.CollectCandidates(classified)
	return recommend.EnumerateAssignments(candidates, recommend.DefaultPolicy(), recommend.DefaultSlotScorer, 3)
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

func runDeckDiff(dataDir string, result *onlinePlayerAnalysisResult) (*evaluation.DeckDiff, error) {
	if len(result.Player.CurrentDeck) == 0 {
		return nil, fmt.Errorf("player has no current deck")
	}

	rec := recommend.NewRecommender(dataDir, recommend.DefaultOptions())
	recResult, err := rec.GenerateRecommendations(result.Player.Tag, result.Player.Name, result.DeckCardAnalysis)
	if err != nil {
		return nil, fmt.Errorf("recommendation failed: %w", err)
	}
	if len(recResult.Recommendations) == 0 || recResult.Recommendations[0].Deck == nil {
		return nil, fmt.Errorf("no deck recommendations available")
	}

	currentCards := playerDeckToCandidates(result.Player.CurrentDeck)
	recommendedCards := deckDetailToCandidates(recResult.Recommendations[0].Deck.DeckDetail)

	synergyDB := deck.NewSynergyDatabase()
	return evaluation.DiffDecks(result.Player, currentCards, recommendedCards, synergyDB)
}

func makeCandidate(name, rarity string, level, maxLevel, elixir, evolutionLevel, maxEvolutionLevel int) deck.CardCandidate {
	return deck.CardCandidate{
		Name:              name,
		Level:             level,
		MaxLevel:          maxLevel,
		Rarity:            rarity,
		Elixir:            elixir,
		Role:              inferRole(name),
		Stats:             inferStats(name),
		HasEvolution:      evolutionLevel > 0,
		EvolutionLevel:    evolutionLevel,
		MaxEvolutionLevel: maxEvolutionLevel,
	}
}

func playerDeckToCandidates(cards []clashroyale.Card) []deck.CardCandidate {
	candidates := make([]deck.CardCandidate, len(cards))
	for i, c := range cards {
		candidates[i] = makeCandidate(c.Name, c.Rarity, c.Level, c.MaxLevel, c.ElixirCost, c.EvolutionLevel, c.MaxEvolutionLevel)
	}
	return candidates
}

func deckDetailToCandidates(details []deck.CardDetail) []deck.CardCandidate {
	candidates := make([]deck.CardCandidate, len(details))
	for i, d := range details {
		candidates[i] = makeCandidate(d.Name, d.Rarity, d.Level, d.MaxLevel, d.Elixir, d.EvolutionLevel, d.MaxEvolutionLevel)
	}
	return candidates
}
