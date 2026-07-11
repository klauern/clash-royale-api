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

type offlineDeckPlayerData struct {
	CardAnalysis deck.CardAnalysis
	PlayerName   string
	PlayerTag    string
	Source       string
}

func loadOfflineDeckPlayerData(loader offlineAnalysisLoader, tag, analysisDir, analysisFile, dataDir string) (*offlineDeckPlayerData, error) {
	loadedAnalysis, err := loadOfflineAnalysisFromFlags(loader, tag, dataDir, analysisDir, analysisFile, false)
	if err != nil {
		return nil, err
	}
	if loadedAnalysis == nil {
		return nil, fmt.Errorf("loader returned nil offline analysis result for player %s", tag)
	}

	return &offlineDeckPlayerData{
		CardAnalysis: loadedAnalysis.CardAnalysis,
		PlayerName:   loadedAnalysis.PlayerName,
		PlayerTag:    loadedAnalysis.PlayerTag,
		Source:       loadedAnalysis.Source,
	}, nil
}

func loadOnlinePlayerAnalysis(ctx context.Context, tag, apiToken string, verbose bool) (*onlinePlayerAnalysisResult, error) {
	client, err := newPlayerAPIClient(apiToken, apiClientOptions{offlineAllowed: true})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize player API client: %w", err)
	}

	player, err := client.GetPlayerWithContext(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	if player == nil {
		return nil, fmt.Errorf("failed to get player %q: empty player response", tag)
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

func mapOnlineAnalysisToOfflineDeckPlayerData(result *onlinePlayerAnalysisResult) *offlineDeckPlayerData {
	return &offlineDeckPlayerData{
		CardAnalysis: result.DeckCardAnalysis,
		PlayerName:   result.Player.Name,
		PlayerTag:    result.Player.Tag,
	}
}
