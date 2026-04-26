package main

import (
	"context"
	"fmt"
	"path/filepath"

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

type offlineAnalysisLoader interface {
	LoadLatestAnalysis(tag, analysisDir string) (*deck.CardAnalysis, error)
	LoadAnalysis(path string) (*deck.CardAnalysis, error)
}

type offlineDeckPlayerData struct {
	CardAnalysis deck.CardAnalysis
	PlayerName   string
	PlayerTag    string
	Source       string
}

func loadOfflineDeckPlayerData(loader offlineAnalysisLoader, tag, analysisDir, analysisFile, dataDir string) (*offlineDeckPlayerData, error) {
	if analysisDir == "" {
		analysisDir = filepath.Join(dataDir, "analysis")
	}

	var loadedAnalysis *deck.CardAnalysis
	var err error

	if analysisFile != "" {
		loadedAnalysis, err = loader.LoadAnalysis(analysisFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load analysis file %s: %w", analysisFile, err)
		}
	} else {
		loadedAnalysis, err = loader.LoadLatestAnalysis(tag, analysisDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load analysis for player %s from %s: %w", tag, analysisDir, err)
		}
	}
	if loadedAnalysis == nil {
		if analysisFile != "" {
			return nil, fmt.Errorf("loader returned nil analysis for player %s from file %s", tag, analysisFile)
		}
		return nil, fmt.Errorf("loader returned nil analysis for player %s from dir %s", tag, analysisDir)
	}

	playerName := loadedAnalysis.PlayerName
	if playerName == "" {
		playerName = tag
	}

	playerTag := loadedAnalysis.PlayerTag
	if playerTag == "" {
		playerTag = tag
	}

	source := analysisDir
	if analysisFile != "" {
		source = analysisFile
	}

	return &offlineDeckPlayerData{
		CardAnalysis: *loadedAnalysis,
		PlayerName:   playerName,
		PlayerTag:    playerTag,
		Source:       source,
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
