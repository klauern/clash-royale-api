package main

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

func TestBuildAnalysisCardLevelsFromPlayer(t *testing.T) {
	player := &clashroyale.Player{
		Name: "Tester",
		Tag:  "#TAG123",
		Cards: []clashroyale.Card{
			{
				ID:                26000000,
				Name:              "Knight",
				Level:             14,
				MaxLevel:          14,
				EvolutionLevel:    1,
				MaxEvolutionLevel: 3,
				Rarity:            "Common",
				ElixirCost:        3,
				Count:             5000,
			},
		},
	}

	cardLevels := buildAnalysisCardLevelsFromPlayer(player)
	info, ok := cardLevels["Knight"]
	if !ok {
		t.Fatalf("expected Knight card level info to be present")
	}
	if info.Name != "Knight" || info.ID != 26000000 {
		t.Fatalf("unexpected card identity: %+v", info)
	}
	if info.Level != 14 || info.MaxLevel != 14 || !info.IsMaxLevel {
		t.Fatalf("unexpected level mapping: %+v", info)
	}
	if info.EvolutionLevel != 1 || info.MaxEvolutionLevel != 3 {
		t.Fatalf("unexpected evolution mapping: %+v", info)
	}
	if info.Rarity != "Common" || info.Elixir != 3 || info.CardCount != 5000 {
		t.Fatalf("unexpected metadata mapping: %+v", info)
	}
	if info.CardsToNext != 0 {
		t.Fatalf("expected CardsToNext to default to 0, got %d", info.CardsToNext)
	}
}

func TestConvertPlayerToAnalysisVariantsStayAligned(t *testing.T) {
	player := &clashroyale.Player{
		Name: "Tester",
		Tag:  "#TAG123",
		Cards: []clashroyale.Card{
			{
				Name:              "Knight",
				Level:             14,
				MaxLevel:          14,
				EvolutionLevel:    1,
				MaxEvolutionLevel: 3,
				Rarity:            "Common",
				ElixirCost:        3,
			},
		},
	}

	deckAnalysis := convertPlayerToAnalysis(player)
	packageAnalysis := convertPlayerToAnalysisPackage(player)

	deckInfo, ok := deckAnalysis.CardLevels["Knight"]
	if !ok {
		t.Fatalf("expected deck analysis to include Knight")
	}
	packageInfo, ok := packageAnalysis.CardLevels["Knight"]
	if !ok {
		t.Fatalf("expected package analysis to include Knight")
	}
	if deckInfo.Level != packageInfo.Level || deckInfo.MaxLevel != packageInfo.MaxLevel {
		t.Fatalf("expected shared level mapping, got deck=%+v package=%+v", deckInfo, packageInfo)
	}
	if deckInfo.Rarity != packageInfo.Rarity || deckInfo.Elixir != packageInfo.Elixir {
		t.Fatalf("expected shared metadata mapping, got deck=%+v package=%+v", deckInfo, packageInfo)
	}
	if deckInfo.EvolutionLevel != packageInfo.EvolutionLevel || deckInfo.MaxEvolutionLevel != packageInfo.MaxEvolutionLevel {
		t.Fatalf("expected shared evolution mapping, got deck=%+v package=%+v", deckInfo, packageInfo)
	}
	if packageAnalysis.TotalCards != 1 {
		t.Fatalf("expected TotalCards to match player cards, got %d", packageAnalysis.TotalCards)
	}
}
