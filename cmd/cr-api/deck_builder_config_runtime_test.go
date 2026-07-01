package main

import (
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
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

func TestApplyExcludeFilterMutatesAnalysisInPlace(t *testing.T) {
	t.Parallel()

	cardAnalysis := deck.CardAnalysis{
		CardLevels: map[string]deck.CardLevelData{
			"Archers":  {Level: 14},
			"Fireball": {Level: 14},
			"Knight":   {Level: 14},
		},
	}

	applyExcludeFilter(&cardAnalysis, []string{" archers ", "FIREBALL"})

	if len(cardAnalysis.CardLevels) != 1 {
		t.Fatalf("expected 1 card after exclusions, got %d", len(cardAnalysis.CardLevels))
	}
	if _, ok := cardAnalysis.CardLevels["Archers"]; ok {
		t.Fatalf("expected Archers to be excluded")
	}
	if _, ok := cardAnalysis.CardLevels["Fireball"]; ok {
		t.Fatalf("expected Fireball to be excluded")
	}
	if _, ok := cardAnalysis.CardLevels["Knight"]; !ok {
		t.Fatalf("expected Knight to remain")
	}
}

func TestApplyCardExclusionsReturnsFilteredCopy(t *testing.T) {
	t.Parallel()

	original := deck.CardAnalysis{
		CardLevels: map[string]deck.CardLevelData{
			"Archers":  {Level: 14},
			"Fireball": {Level: 14},
			"Knight":   {Level: 14},
		},
	}

	filtered := applyCardExclusions(original, []string{"fireball"})

	if len(filtered.CardLevels) != 2 {
		t.Fatalf("expected 2 cards after exclusions, got %d", len(filtered.CardLevels))
	}
	if _, ok := filtered.CardLevels["Fireball"]; ok {
		t.Fatalf("expected Fireball to be excluded from filtered copy")
	}
	if len(original.CardLevels) != 3 {
		t.Fatalf("expected original card levels to remain unchanged, got %d", len(original.CardLevels))
	}
	if _, ok := original.CardLevels["Fireball"]; !ok {
		t.Fatalf("expected original card levels to retain Fireball")
	}
}
