package main

import (
	"context"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

type fakePlayerClient struct {
	player *clashroyale.Player
	err    error
}

func (f fakePlayerClient) GetPlayerWithContext(_ context.Context, _ string) (*clashroyale.Player, error) {
	return f.player, f.err
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
