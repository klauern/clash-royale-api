package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

type offlineAnalysisLoadResult struct {
	CardAnalysis deck.CardAnalysis
	PlayerName   string
	PlayerTag    string
}

func loadOfflineAnalysisFromFlags(
	builder *deck.Builder,
	playerTag string,
	dataDir string,
	analysisDir string,
	analysisFile string,
	verbose bool,
) (*offlineAnalysisLoadResult, error) {
	if builder == nil {
		return nil, fmt.Errorf("deck builder is required")
	}

	resolvedAnalysisDir := resolveAnalysisDir(dataDir, analysisDir)
	loadedAnalysis, err := loadDeckCardAnalysis(builder, playerTag, resolvedAnalysisDir, analysisFile, verbose)
	if err != nil {
		return nil, err
	}

	playerName := loadedAnalysis.PlayerName
	if playerName == "" {
		playerName = playerTag
	}

	resolvedTag := loadedAnalysis.PlayerTag
	if resolvedTag == "" {
		resolvedTag = playerTag
	}

	return &offlineAnalysisLoadResult{
		CardAnalysis: *loadedAnalysis,
		PlayerName:   playerName,
		PlayerTag:    resolvedTag,
	}, nil
}

func resolveAnalysisDir(dataDir, analysisDir string) string {
	if analysisDir != "" {
		return analysisDir
	}

	return filepath.Join(dataDir, "analysis")
}

func loadDeckCardAnalysis(
	builder *deck.Builder,
	playerTag string,
	analysisDir string,
	analysisFile string,
	verbose bool,
) (*deck.CardAnalysis, error) {
	if analysisFile != "" {
		loadedAnalysis, err := builder.LoadAnalysis(analysisFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load analysis from --analysis-file %q: %w", analysisFile, err)
		}
		if verbose {
			printf("Loaded analysis from: %s\n", analysisFile)
		}
		return loadedAnalysis, nil
	}

	normalizedTag := strings.TrimSpace(playerTag)
	if normalizedTag == "" {
		return nil, fmt.Errorf("--tag is required when using --analysis-dir without --analysis-file")
	}

	loadedAnalysis, err := builder.LoadLatestAnalysis(normalizedTag, analysisDir)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to load latest analysis for player %s from --analysis-dir %q: %w",
			normalizedTag,
			analysisDir,
			err,
		)
	}
	if verbose {
		printf("Loaded latest analysis from: %s\n", analysisDir)
	}
	return loadedAnalysis, nil
}
