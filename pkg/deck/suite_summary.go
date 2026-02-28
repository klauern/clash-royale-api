package deck

import (
	"errors"
	"fmt"
	"strings"

	"github.com/klauer/clash-royale-api/go/internal/storage"
)

const suiteSummaryVersion = "1.0.0"

var ErrNilRecommendation = errors.New("deck recommendation is required")

// SuiteDeckSummary captures per-deck metadata in suite summary files.
type SuiteDeckSummary struct {
	Strategy  string   `json:"strategy"`
	Variation int      `json:"variation"`
	Cards     []string `json:"cards"`
	AvgElixir float64  `json:"avg_elixir"`
	FilePath  string   `json:"file_path"`
}

// SuiteBuildInfo captures suite generation statistics.
type SuiteBuildInfo struct {
	TotalDecks     int    `json:"total_decks"`
	Successful     int    `json:"successful"`
	Failed         int    `json:"failed"`
	Strategies     int    `json:"strategies"`
	Variations     int    `json:"variations"`
	GenerationTime string `json:"generation_time"`
}

// SuitePlayer identifies the player for a generated suite.
type SuitePlayer struct {
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

// SuiteSummary is the shared output schema for deck suite summaries.
type SuiteSummary struct {
	Version   string             `json:"version"`
	Timestamp string             `json:"timestamp"`
	Player    SuitePlayer        `json:"player"`
	BuildInfo SuiteBuildInfo     `json:"build_info"`
	Decks     []SuiteDeckSummary `json:"decks"`
}

// SuiteDeckPayload is the persisted schema for individual suite decks.
type SuiteDeckPayload struct {
	Deck           []string            `json:"deck"`
	AvgElixir      float64             `json:"avg_elixir"`
	Recommendation *DeckRecommendation `json:"recommendation,omitempty"`
}

// SuiteDeckFilename builds a standard suite deck filename.
func SuiteDeckFilename(timestamp, strategy string, variation int, playerTag string) string {
	return fmt.Sprintf(
		"%s_deck_%s_var%d_%s.json",
		timestamp,
		strategy,
		variation,
		trimTagPrefix(playerTag),
	)
}

// SuiteSummaryFilename builds a standard suite summary filename.
func SuiteSummaryFilename(timestamp, playerTag string) string {
	return fmt.Sprintf("%s_deck_suite_summary_%s.json", timestamp, trimTagPrefix(playerTag))
}

// NewSuiteSummary creates a summary payload with the standard version field.
func NewSuiteSummary(timestamp, playerName, playerTag string, buildInfo SuiteBuildInfo, decks []SuiteDeckSummary) SuiteSummary {
	return SuiteSummary{
		Version:   suiteSummaryVersion,
		Timestamp: timestamp,
		Player: SuitePlayer{
			Name: playerName,
			Tag:  playerTag,
		},
		BuildInfo: buildInfo,
		Decks:     decks,
	}
}

// WriteSuiteDeck writes a suite deck payload to disk.
func WriteSuiteDeck(path string, recommendation *DeckRecommendation) error {
	if recommendation == nil {
		return ErrNilRecommendation
	}

	payload := SuiteDeckPayload{
		Deck:           recommendation.Deck,
		AvgElixir:      recommendation.AvgElixir,
		Recommendation: recommendation,
	}
	return storage.WriteJSON(path, payload)
}

// WriteSuiteSummary writes a suite summary payload to disk.
func WriteSuiteSummary(path string, summary SuiteSummary) error {
	return storage.WriteJSON(path, summary)
}

func trimTagPrefix(playerTag string) string {
	return strings.TrimPrefix(strings.TrimSpace(playerTag), "#")
}
