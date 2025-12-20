package csv

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/archetypes"
)

// NewArchetypeExporter creates archetype comparison CSV exporter
func NewArchetypeExporter() *CSVExporter {
	return NewCSVExporter(
		"archetype_comparison.csv",
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
func archetypeExport(dataDir string, data interface{}) error {
	result, ok := data.(*archetypes.ArchetypeAnalysisResult)
	if !ok {
		return fmt.Errorf("expected ArchetypeAnalysisResult, got %T", data)
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

	exporter := &BaseExporter{FilenameBase: "archetype_comparison.csv"}
	filePath := filepath.Join(dataDir, "csv", "archetypes", exporter.FilenameBase)
	return exporter.writeCSV(filePath, archetypeHeaders(), rows)
}

// NewArchetypeDetailsExporter creates per-card upgrade details exporter
func NewArchetypeDetailsExporter() *CSVExporter {
	return NewCSVExporter(
		"archetype_upgrade_details.csv",
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
func archetypeDetailsExport(dataDir string, data interface{}) error {
	result, ok := data.(*archetypes.ArchetypeAnalysisResult)
	if !ok {
		return fmt.Errorf("expected ArchetypeAnalysisResult, got %T", data)
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

	exporter := &BaseExporter{FilenameBase: "archetype_upgrade_details.csv"}
	filePath := filepath.Join(dataDir, "csv", "archetypes", exporter.FilenameBase)
	return exporter.writeCSV(filePath, archetypeDetailsHeaders(), rows)
}

// formatDeckCSV formats deck cards as comma-separated string
func formatDeckCSV(deck []string) string {
	if len(deck) == 0 {
		return ""
	}
	return strings.Join(deck, ", ")
}
