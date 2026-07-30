package csv

import (
	"fmt"

	"github.com/klauer/clash-royale-api/go/internal/storage"
	"github.com/klauer/clash-royale-api/go/pkg/analysis"
)

// NewAnalysisExporter creates a new analysis CSV exporter
func NewAnalysisExporter() *CSVExporter {
	return NewCSVExporter(
		"card_analysis.csv",
		analysisHeaders,
		analysisExport,
	)
}

// analysisHeaders returns the CSV headers for card analysis data
func analysisHeaders() []string {
	return []string{
		"Player Tag",
		"Player Name",
		"Analysis Time",
		"Total Cards",
		"Max Level Cards",
		"Upgradable Cards",
		"Average Card Level",
		"Completion Percentage",
		"Average Level Ratio",
	}
}

// analysisExport exports card analysis summary to CSV
func analysisExport(dataDir string, data any) error {
	cardAnalysis, ok := data.(*analysis.CardAnalysis)
	if !ok {
		return fmt.Errorf("expected CardAnalysis type, got %T", data)
	}

	// Prepare CSV rows
	rows := [][]string{
		{
			cardAnalysis.PlayerTag,
			cardAnalysis.PlayerName,
			cardAnalysis.AnalysisTime.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%d", cardAnalysis.TotalCards),
			fmt.Sprintf("%d", cardAnalysis.Summary.MaxLevelCards),
			fmt.Sprintf("%d", cardAnalysis.Summary.UpgradableCards),
			fmt.Sprintf("%.2f", cardAnalysis.Summary.AvgCardLevel),
			fmt.Sprintf("%.1f%%", cardAnalysis.Summary.CompletionPercent),
			fmt.Sprintf("%.3f", cardAnalysis.Summary.AvgLevelRatio),
		},
	}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "card_analysis.csv"}
	filePath := exporter.csvFilePath(dataDir, storage.CSVAnalysisSubdir)
	return exporter.writeCSV(filePath, analysisHeaders(), rows)
}
