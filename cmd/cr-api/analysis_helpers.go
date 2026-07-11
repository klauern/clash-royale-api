package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

type offlineAnalysisLoader interface {
	LoadLatestAnalysis(tag, analysisDir string) (*deck.CardAnalysis, error)
	LoadAnalysis(path string) (*deck.CardAnalysis, error)
}

type offlineAnalysisLoadResult struct {
	CardAnalysis deck.CardAnalysis
	PlayerName   string
	PlayerTag    string
	Source       string
}

func loadOfflineAnalysisFromFlags(
	loader offlineAnalysisLoader,
	playerTag string,
	dataDir string,
	analysisDir string,
	analysisFile string,
	verbose bool,
) (*offlineAnalysisLoadResult, error) {
	if loader == nil {
		return nil, fmt.Errorf("offline analysis loader is required")
	}

	resolvedAnalysisDir := resolveAnalysisDir(dataDir, analysisDir)
	loadedAnalysis, err := loadDeckCardAnalysis(loader, playerTag, resolvedAnalysisDir, analysisFile, verbose)
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
		Source:       offlineAnalysisSource(analysisFile, resolvedAnalysisDir),
	}, nil
}

func resolveAnalysisDir(dataDir, analysisDir string) string {
	if analysisDir != "" {
		return analysisDir
	}

	return filepath.Join(dataDir, "analysis")
}

func loadDeckCardAnalysis(
	loader offlineAnalysisLoader,
	playerTag string,
	analysisDir string,
	analysisFile string,
	verbose bool,
) (*deck.CardAnalysis, error) {
	if analysisFile != "" {
		loadedAnalysis, err := loader.LoadAnalysis(analysisFile)
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

	loadedAnalysis, err := loader.LoadLatestAnalysis(normalizedTag, analysisDir)
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

func offlineAnalysisSource(analysisFile, analysisDir string) string {
	if analysisFile != "" {
		return analysisFile
	}

	return analysisDir
}
