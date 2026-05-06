package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// SuiteSummaryWriteError wraps write failures with the target path.
type SuiteSummaryWriteError struct {
	Path string
	Err  error
}

func (e *SuiteSummaryWriteError) Error() string {
	return fmt.Sprintf("write suite summary %q: %v", e.Path, e.Err)
}

func (e *SuiteSummaryWriteError) Unwrap() error {
	return e.Err
}

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

// writeSuiteSummary assembles a SuiteSummary from build results and writes it
// to outputDir, returning the destination path. Centralizes the
// NewSuiteSummary + WriteSuiteSummary pair shared by build/analyze suite
// runtimes.
func writeSuiteSummary(
	outputDir, timestamp, playerName, playerTag string,
	info deck.SuiteBuildInfo,
	summaries []deck.SuiteDeckSummary,
) (string, error) {
	path := filepath.Join(outputDir, deck.SuiteSummaryFilename(timestamp, playerTag))
	summary := deck.NewSuiteSummary(
		time.Now().UTC().Format(time.RFC3339),
		playerName,
		playerTag,
		info,
		summaries,
	)
	if err := deck.WriteSuiteSummary(path, summary); err != nil {
		return "", &SuiteSummaryWriteError{Path: path, Err: err}
	}
	return path, nil
}
