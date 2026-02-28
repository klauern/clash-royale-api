package main

import (
	"path/filepath"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func saveSuiteDeckFile(
	outputDir, strategy string,
	variation int,
	playerTag string,
	recommendation *deck.DeckRecommendation,
) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	targetPath := filepath.Join(outputDir, deck.SuiteDeckFilename(timestamp, strategy, variation, playerTag))
	if err := deck.WriteSuiteDeck(targetPath, recommendation); err != nil {
		return "", err
	}
	return targetPath, nil
}
