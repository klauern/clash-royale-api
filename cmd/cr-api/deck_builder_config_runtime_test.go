package main

import (
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

func TestConvertToDeckCardAnalysisPreservesEvolutionFields(t *testing.T) {
	t.Parallel()

	player := &clashroyale.Player{
		Name: "Test Player",
		Tag:  "#TEST",
	}
	cardAnalysis := &analysis.CardAnalysis{
		AnalysisTime: time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC),
		CardLevels: map[string]analysis.CardLevelInfo{
			"Archers": {
				Level:             14,
				MaxLevel:          15,
				Rarity:            "common",
				Elixir:            3,
				EvolutionLevel:    1,
				MaxEvolutionLevel: 2,
			},
		},
	}

	got := convertToDeckCardAnalysis(cardAnalysis, player)
	card, ok := got.CardLevels["Archers"]
	if !ok {
		t.Fatalf("converted analysis missing card entry")
	}
	if card.EvolutionLevel != 1 {
		t.Fatalf("EvolutionLevel mismatch: got %d, want 1", card.EvolutionLevel)
	}
	if card.MaxEvolutionLevel != 2 {
		t.Fatalf("MaxEvolutionLevel mismatch: got %d, want 2", card.MaxEvolutionLevel)
	}
	if got.PlayerName != player.Name || got.PlayerTag != player.Tag {
		t.Fatalf("player metadata mismatch: got %q/%q", got.PlayerName, got.PlayerTag)
	}
	if got.AnalysisTime != "2026-03-02T12:00:00Z" {
		t.Fatalf("analysis time mismatch: got %q", got.AnalysisTime)
	}
}
