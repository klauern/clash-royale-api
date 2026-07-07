package main

import (
	"fmt"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

const deckCardCount = 8

func parseDeckStringWithLabel(deckStr, label string) ([]string, error) {
	cardNames := parseDeckString(deckStr)
	if len(cardNames) == deckCardCount {
		return cardNames, nil
	}

	if label == "" {
		label = optimizeDefaultTag
	}

	return nil, fmt.Errorf("%s must contain exactly %d cards, got %d", label, deckCardCount, len(cardNames))
}

func loadDeckCardsFromInput(deckString, fromAnalysis string) ([]string, error) {
	if deckString != "" {
		return parseDeckStringWithLabel(deckString, optimizeDefaultTag)
	}

	deckCardNames, err := loadDeckFromAnalysis(fromAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to load deck from analysis: %w", err)
	}

	return deckCardNames, nil
}

func newDefaultDeckCandidate(name string) deck.CardCandidate {
	return deck.CardCandidate{
		Name:     name,
		Level:    11,
		MaxLevel: 15,
		Rarity:   inferRarity(name),
		Elixir:   config.GetCardElixir(name, 0),
		Role:     inferRole(name),
		Stats:    inferStats(name),
	}
}
