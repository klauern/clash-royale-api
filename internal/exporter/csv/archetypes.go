package csv

import (
	"fmt"
	"strings"

	"github.com/klauer/clash-royale-api/go/internal/storage"
	"github.com/klauer/clash-royale-api/go/pkg/archetypes"
)

const (
	archetypeComparisonFilename     = "archetype_comparison.csv"
	archetypeUpgradeDetailsFilename = "archetype_upgrade_details.csv"
)

// NewArchetypeExporter creates archetype comparison CSV exporter
func NewArchetypeExporter() *CSVExporter {
	return NewCSVExporter(
		archetypeComparisonFilename,
		archetypeHeaders,
		archetypeExport,
	)
}

// archetypeHeaders returns CSV headers for archetype comparison
func archetypeHeaders() []string {
	return []string{
		"Player Tag",
		"Player Name",
		"Target Level",
		"Archetype",
		"Avg Elixir",
		"Current Avg Level",
		"Cards Needed",
		"Gold Needed",
		"Distance Metric",
		"Deck",
	}
}

// archetypeExport exports archetype analysis to CSV
func archetypeExport(dataDir string, data any) error {
	result, err := assertCSVExportType[*archetypes.ArchetypeAnalysisResult](data)
	if err != nil {
		return err
	}

	var rows [][]string

	for _, arch := range result.Archetypes {
		deckStr := formatDeckCSV(arch.Deck)

		row := []string{
			result.PlayerTag,
			result.PlayerName,
			fmt.Sprintf("%d", result.TargetLevel),
			string(arch.Archetype),
			fmt.Sprintf("%.1f", arch.AvgElixir),
			fmt.Sprintf("%.1f", arch.CurrentAvgLevel),
			fmt.Sprintf("%d", arch.CardsNeeded),
			fmt.Sprintf("%d", arch.GoldNeeded),
			fmt.Sprintf("%.3f", arch.DistanceMetric),
			deckStr,
		}
		rows = append(rows, row)
	}

	return writeCSVRows(dataDir, storage.CSVArchetypesSubdir, archetypeComparisonFilename, archetypeHeaders(), rows)
}

// NewArchetypeDetailsExporter creates per-card upgrade details exporter
func NewArchetypeDetailsExporter() *CSVExporter {
	return NewCSVExporter(
		archetypeUpgradeDetailsFilename,
		archetypeDetailsHeaders,
		archetypeDetailsExport,
	)
}

// archetypeDetailsHeaders returns detailed CSV headers
func archetypeDetailsHeaders() []string {
	return []string{
		"Player Tag",
		"Archetype",
		"Card Name",
		"Current Level",
		"Target Level",
		"Level Gap",
		"Rarity",
		"Cards Needed",
		"Gold Needed",
	}
}

// archetypeDetailsExport exports per-card upgrade details
func archetypeDetailsExport(dataDir string, data any) error {
	result, err := assertCSVExportType[*archetypes.ArchetypeAnalysisResult](data)
	if err != nil {
		return err
	}

	var rows [][]string

	for _, arch := range result.Archetypes {
		for _, upgrade := range arch.UpgradeDetails {
			row := []string{
				result.PlayerTag,
				string(arch.Archetype),
				upgrade.CardName,
				fmt.Sprintf("%d", upgrade.CurrentLevel),
				fmt.Sprintf("%d", upgrade.TargetLevel),
				fmt.Sprintf("%d", upgrade.LevelGap),
				upgrade.Rarity,
				fmt.Sprintf("%d", upgrade.CardsNeeded),
				fmt.Sprintf("%d", upgrade.GoldNeeded),
			}
			rows = append(rows, row)
		}
	}

	return writeCSVRows(dataDir, storage.CSVArchetypesSubdir, archetypeUpgradeDetailsFilename, archetypeDetailsHeaders(), rows)
}

// formatDeckCSV formats deck cards as comma-separated string
func formatDeckCSV(deck []string) string {
	if len(deck) == 0 {
		return ""
	}
	return strings.Join(deck, ", ")
}
