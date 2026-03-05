package main

import (
	"context"
	"fmt"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

type playerAPIClient interface {
	GetPlayerWithContext(ctx context.Context, tag string) (*clashroyale.Player, error)
}

var newPlayerAPIClient = func(apiToken string, opts apiClientOptions) (playerAPIClient, error) {
	return requireAPIClientFromToken(apiToken, opts)
}

type onlinePlayerAnalysisResult struct {
	Player           *clashroyale.Player
	CardAnalysis     *analysis.CardAnalysis
	DeckCardAnalysis deck.CardAnalysis
}

func loadOnlinePlayerAnalysis(ctx context.Context, tag, apiToken string, verbose bool) (*onlinePlayerAnalysisResult, error) {
	client, err := newPlayerAPIClient(apiToken, apiClientOptions{offlineAllowed: true})
	if err != nil {
		return nil, err
	}

	player, err := client.GetPlayerWithContext(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	if verbose {
		printf("Player: %s (%s)\n", player.Name, player.Tag)
		printf("Analyzing %d cards...\n", len(player.Cards))
	}

	cardAnalysis, err := analysis.AnalyzeCardCollection(player, analysis.DefaultAnalysisOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to analyze card collection: %w", err)
	}

	return newOnlinePlayerAnalysisResult(player, cardAnalysis), nil
}

func newOnlinePlayerAnalysisResult(player *clashroyale.Player, cardAnalysis *analysis.CardAnalysis) *onlinePlayerAnalysisResult {
	return &onlinePlayerAnalysisResult{
		Player:           player,
		CardAnalysis:     cardAnalysis,
		DeckCardAnalysis: convertToDeckCardAnalysis(cardAnalysis, player),
	}
}
