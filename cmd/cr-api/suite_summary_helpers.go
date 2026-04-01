package main

import "github.com/klauer/clash-royale-api/go/pkg/deck"

func newSuiteDeckSummary(strategy string, variation int, cards []string, avgElixir float64, filePath string) deck.SuiteDeckSummary {
	return deck.SuiteDeckSummary{
		Strategy:  strategy,
		Variation: variation,
		Cards:     cards,
		AvgElixir: avgElixir,
		FilePath:  filePath,
	}
}
