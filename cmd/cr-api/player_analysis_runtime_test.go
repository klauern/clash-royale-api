package main

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

type fakePlayerClient struct {
	player *clashroyale.Player
	err    error
}

func (f fakePlayerClient) GetPlayerWithContext(_ context.Context, _ string) (*clashroyale.Player, error) {
	return f.player, f.err
}

type fakeOfflineAnalysisLoader struct {
	analysis        *deck.CardAnalysis
	loadLatestErr   error
	loadAnalysisErr error
	latestTag       string
	latestDir       string
	analysisPath    string
	latestCalls     int
	analysisCalls   int
}

const offlineTag = "#POFFLINE"

func (f *fakeOfflineAnalysisLoader) LoadLatestAnalysis(tag, analysisDir string) (*deck.CardAnalysis, error) {
	f.latestCalls++
	f.latestTag = tag
	f.latestDir = analysisDir
	if f.loadLatestErr != nil {
		return nil, f.loadLatestErr
	}
	return f.analysis, nil
}

func (f *fakeOfflineAnalysisLoader) LoadAnalysis(path string) (*deck.CardAnalysis, error) {
	f.analysisCalls++
	f.analysisPath = path
	if f.loadAnalysisErr != nil {
		return nil, f.loadAnalysisErr
	}
	return f.analysis, nil
}

func TestLoadOnlinePlayerAnalysisPreservesEvolutionLevels(t *testing.T) {
	originalFactory := newPlayerAPIClient
	t.Cleanup(func() {
		newPlayerAPIClient = originalFactory
	})

	newPlayerAPIClient = func(_ string, _ apiClientOptions) (playerAPIClient, error) {
		return fakePlayerClient{
			player: &clashroyale.Player{
				Tag:  "#PTEST",
				Name: "Test Player",
				Cards: []clashroyale.Card{
					{
						Name:              "Archers",
						Rarity:            "Common",
						Level:             1,
						MaxLevel:          14,
						Count:             1000,
						ElixirCost:        3,
						EvolutionLevel:    1,
						MaxEvolutionLevel: 2,
					},
				},
			},
		}, nil
	}

	result, err := loadOnlinePlayerAnalysis(context.Background(), "PTEST", "token", false)
	if err != nil {
		t.Fatalf("loadOnlinePlayerAnalysis returned error: %v", err)
	}

	card, ok := result.DeckCardAnalysis.CardLevels["Archers"]
	if !ok {
		t.Fatalf("expected Archers in deck analysis, got %#v", result.DeckCardAnalysis.CardLevels)
	}
	if card.EvolutionLevel != 1 {
		t.Fatalf("EvolutionLevel=%d, want 1", card.EvolutionLevel)
	}
	if card.MaxEvolutionLevel != 2 {
		t.Fatalf("MaxEvolutionLevel=%d, want 2", card.MaxEvolutionLevel)
	}
	if result.DeckCardAnalysis.PlayerName != "Test Player" {
		t.Fatalf("PlayerName=%q, want Test Player", result.DeckCardAnalysis.PlayerName)
	}
	if result.DeckCardAnalysis.PlayerTag != "#PTEST" {
		t.Fatalf("PlayerTag=%q, want #PTEST", result.DeckCardAnalysis.PlayerTag)
	}
}

func TestLoadSuitePlayerDataFromAPIUsesSharedDeckAnalysis(t *testing.T) {
	originalFactory := newPlayerAPIClient
	t.Cleanup(func() {
		newPlayerAPIClient = originalFactory
	})

	newPlayerAPIClient = func(_ string, _ apiClientOptions) (playerAPIClient, error) {
		return fakePlayerClient{
			player: &clashroyale.Player{
				Tag:  "#PSUITE",
				Name: "Suite Player",
				Cards: []clashroyale.Card{
					{
						Name:              "Firecracker",
						Rarity:            "Common",
						Level:             1,
						MaxLevel:          14,
						Count:             500,
						ElixirCost:        3,
						EvolutionLevel:    1,
						MaxEvolutionLevel: 1,
					},
				},
			},
		}, nil
	}

	result, err := loadSuitePlayerDataFromAPI(context.Background(), "PSUITE", "token", false)
	if err != nil {
		t.Fatalf("loadSuitePlayerDataFromAPI returned error: %v", err)
	}

	card, ok := result.CardAnalysis.CardLevels["Firecracker"]
	if !ok {
		t.Fatalf("expected Firecracker in suite analysis, got %#v", result.CardAnalysis.CardLevels)
	}
	if card.EvolutionLevel != 1 {
		t.Fatalf("EvolutionLevel=%d, want 1", card.EvolutionLevel)
	}
	if card.MaxEvolutionLevel != 1 {
		t.Fatalf("MaxEvolutionLevel=%d, want 1", card.MaxEvolutionLevel)
	}
	if result.PlayerName != "Suite Player" {
		t.Fatalf("PlayerName=%q, want Suite Player", result.PlayerName)
	}
	if result.PlayerTag != "#PSUITE" {
		t.Fatalf("PlayerTag=%q, want #PSUITE", result.PlayerTag)
	}
}

func TestLoadOfflineDeckPlayerDataDefaultsDirAndFallbacks(t *testing.T) {
	loader := &fakeOfflineAnalysisLoader{
		analysis: &deck.CardAnalysis{},
	}

	result, err := loadOfflineDeckPlayerData(loader, offlineTag, "", "", "testdata")
	if err != nil {
		t.Fatalf("loadOfflineDeckPlayerData returned error: %v", err)
	}

	if loader.latestCalls != 1 {
		t.Fatalf("LoadLatestAnalysis calls=%d, want 1", loader.latestCalls)
	}
	if loader.analysisCalls != 0 {
		t.Fatalf("LoadAnalysis calls=%d, want 0", loader.analysisCalls)
	}
	if loader.latestTag != offlineTag {
		t.Fatalf("latest tag=%q, want %q", loader.latestTag, offlineTag)
	}
	expectedDir := filepath.Join("testdata", "analysis")
	if loader.latestDir != expectedDir {
		t.Fatalf("latest dir=%q, want %q", loader.latestDir, expectedDir)
	}
	if result.PlayerName != offlineTag {
		t.Fatalf("PlayerName=%q, want %q", result.PlayerName, offlineTag)
	}
	if result.PlayerTag != offlineTag {
		t.Fatalf("PlayerTag=%q, want %q", result.PlayerTag, offlineTag)
	}
	if result.Source != expectedDir {
		t.Fatalf("Source=%q, want %q", result.Source, expectedDir)
	}
}

func TestLoadOfflineDeckPlayerDataUsesExplicitAnalysisDir(t *testing.T) {
	loader := &fakeOfflineAnalysisLoader{
		analysis: &deck.CardAnalysis{},
	}
	customDir := "/tmp/custom-analysis"

	result, err := loadOfflineDeckPlayerData(loader, offlineTag, customDir, "", "testdata")
	if err != nil {
		t.Fatalf("loadOfflineDeckPlayerData returned error: %v", err)
	}
	if loader.latestCalls != 1 {
		t.Fatalf("LoadLatestAnalysis calls=%d, want 1", loader.latestCalls)
	}
	if loader.latestDir != customDir {
		t.Fatalf("latest dir=%q, want %q", loader.latestDir, customDir)
	}
	if result.Source != customDir {
		t.Fatalf("Source=%q, want %q", result.Source, customDir)
	}
}

func TestLoadOfflineDeckPlayerDataUsesExplicitFile(t *testing.T) {
	loader := &fakeOfflineAnalysisLoader{
		analysis: &deck.CardAnalysis{
			PlayerName: "Offline User",
			PlayerTag:  "#OFF",
		},
	}

	result, err := loadOfflineDeckPlayerData(loader, offlineTag, "", "analysis.json", "testdata")
	if err != nil {
		t.Fatalf("loadOfflineDeckPlayerData returned error: %v", err)
	}

	if loader.analysisCalls != 1 {
		t.Fatalf("LoadAnalysis calls=%d, want 1", loader.analysisCalls)
	}
	if loader.analysisPath != "analysis.json" {
		t.Fatalf("analysis path=%q, want analysis.json", loader.analysisPath)
	}
	if loader.latestCalls != 0 {
		t.Fatalf("LoadLatestAnalysis calls=%d, want 0", loader.latestCalls)
	}
	if result.PlayerName != "Offline User" {
		t.Fatalf("PlayerName=%q, want Offline User", result.PlayerName)
	}
	if result.PlayerTag != "#OFF" {
		t.Fatalf("PlayerTag=%q, want #OFF", result.PlayerTag)
	}
	if result.Source != "analysis.json" {
		t.Fatalf("Source=%q, want analysis.json", result.Source)
	}
}

func TestLoadOfflineDeckPlayerDataWrapsLoadErrors(t *testing.T) {
	sentinel := errors.New("boom")
	loader := &fakeOfflineAnalysisLoader{
		loadLatestErr: sentinel,
	}

	_, err := loadOfflineDeckPlayerData(loader, offlineTag, "", "", "testdata")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected wrapped sentinel error, got %v", err)
	}
	if !strings.Contains(err.Error(), offlineTag) {
		t.Fatalf("expected error to include player tag, got %q", err.Error())
	}
}
