package csv

import (
	"fmt"
	"path/filepath"

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
func analysisExport(dataDir string, data interface{}) error {
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
	filePath := filepath.Join(dataDir, "csv", "analysis", exporter.FilenameBase)
	return exporter.writeCSV(filePath, analysisHeaders(), rows)
}

// NewCardLevelsExporter creates a new card levels CSV exporter
func NewCardLevelsExporter() *CSVExporter {
	return NewCSVExporter(
		"card_levels.csv",
		cardLevelsHeaders,
		cardLevelsExport,
	)
}

// cardLevelsHeaders returns the CSV headers for card levels data
func cardLevelsHeaders() []string {
	return []string{
		"Player Tag",
		"Card Name",
		"Card ID",
		"Current Level",
		"Max Level",
		"Level Ratio",
		"Rarity",
		"Elixir Cost",
		"Cards Owned",
		"Cards Needed for Next",
		"Progress to Next %",
		"Is Max Level",
	}
}

// cardLevelsExport exports detailed card levels to CSV
func cardLevelsExport(dataDir string, data interface{}) error {
	cardAnalysis, ok := data.(*analysis.CardAnalysis)
	if !ok {
		return fmt.Errorf("expected CardAnalysis type, got %T", data)
	}

	// Prepare CSV rows
	var rows [][]string

	for _, cardInfo := range cardAnalysis.CardLevels {
		progressToNext := "100.0"
		if !cardInfo.IsMaxLevel {
			progressToNext = fmt.Sprintf("%.1f", cardInfo.ProgressToNext())
		}

		row := []string{
			cardAnalysis.PlayerTag,
			cardInfo.Name,
			fmt.Sprintf("%d", cardInfo.ID),
			fmt.Sprintf("%d", cardInfo.Level),
			fmt.Sprintf("%d", cardInfo.MaxLevel),
			fmt.Sprintf("%.3f", cardInfo.LevelRatio()),
			cardInfo.Rarity,
			fmt.Sprintf("%d", cardInfo.Elixir),
			fmt.Sprintf("%d", cardInfo.CardCount),
			fmt.Sprintf("%d", cardInfo.CardsToNext),
			progressToNext,
			fmt.Sprintf("%t", cardInfo.IsMaxLevel),
		}
		rows = append(rows, row)
	}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "card_levels.csv"}
	filePath := filepath.Join(dataDir, "csv", "analysis", exporter.FilenameBase)
	return exporter.writeCSV(filePath, cardLevelsHeaders(), rows)
}

// NewUpgradePrioritiesExporter creates a new upgrade priorities CSV exporter
func NewUpgradePrioritiesExporter() *CSVExporter {
	return NewCSVExporter(
		"upgrade_priorities.csv",
		upgradePrioritiesHeaders,
		upgradePrioritiesExport,
	)
}

// upgradePrioritiesHeaders returns the CSV headers for upgrade priorities data
func upgradePrioritiesHeaders() []string {
	return []string{
		"Player Tag",
		"Card Name",
		"Rarity",
		"Current Level",
		"Max Level",
		"Cards Owned",
		"Cards Needed",
		"Priority",
		"Priority Score",
		"Ready to Upgrade",
		"Completion %",
		"Reasons",
	}
}

// upgradePrioritiesExport exports upgrade priorities to CSV
func upgradePrioritiesExport(dataDir string, data interface{}) error {
	cardAnalysis, ok := data.(*analysis.CardAnalysis)
	if !ok {
		return fmt.Errorf("expected CardAnalysis type, got %T", data)
	}

	// Prepare CSV rows
	var rows [][]string

	for _, priority := range cardAnalysis.UpgradePriority {
		// Combine reasons into a single string
		reasons := ""
		for i, reason := range priority.Reasons {
			if i > 0 {
				reasons += "; "
			}
			reasons += reason
		}

		row := []string{
			cardAnalysis.PlayerTag,
			priority.CardName,
			priority.Rarity,
			fmt.Sprintf("%d", priority.CurrentLevel),
			fmt.Sprintf("%d", priority.MaxLevel),
			fmt.Sprintf("%d", priority.CardsOwned),
			fmt.Sprintf("%d", priority.CardsNeeded),
			priority.Priority,
			fmt.Sprintf("%.1f", priority.PriorityScore),
			fmt.Sprintf("%t", priority.IsReadyToUpgrade()),
			fmt.Sprintf("%.1f", priority.PercentageComplete()),
			reasons,
		}
		rows = append(rows, row)
	}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "upgrade_priorities.csv"}
	filePath := filepath.Join(dataDir, "csv", "analysis", exporter.FilenameBase)
	return exporter.writeCSV(filePath, upgradePrioritiesHeaders(), rows)
}

// NewRarityBreakdownExporter creates a new rarity breakdown CSV exporter
func NewRarityBreakdownExporter() *CSVExporter {
	return NewCSVExporter(
		"rarity_breakdown.csv",
		rarityBreakdownHeaders,
		rarityBreakdownExport,
	)
}

// rarityBreakdownHeaders returns the CSV headers for rarity breakdown data
func rarityBreakdownHeaders() []string {
	return []string{
		"Player Tag",
		"Rarity",
		"Total Cards",
		"Max Level Cards",
		"Average Level",
		"Average Level Ratio",
		"Cards Near Max",
		"Cards Ready to Upgrade",
	}
}

// rarityBreakdownExport exports rarity breakdown to CSV
func rarityBreakdownExport(dataDir string, data interface{}) error {
	cardAnalysis, ok := data.(*analysis.CardAnalysis)
	if !ok {
		return fmt.Errorf("expected CardAnalysis type, got %T", data)
	}

	// Prepare CSV rows
	var rows [][]string

	for _, stats := range cardAnalysis.RarityBreakdown {
		row := []string{
			cardAnalysis.PlayerTag,
			stats.Rarity,
			fmt.Sprintf("%d", stats.TotalCards),
			fmt.Sprintf("%d", stats.MaxLevelCards),
			fmt.Sprintf("%.1f", stats.AvgLevel),
			fmt.Sprintf("%.3f", stats.AvgLevelRatio),
			fmt.Sprintf("%d", stats.CardsNearMax),
			fmt.Sprintf("%d", stats.CardsReadyUpgrade),
		}
		rows = append(rows, row)
	}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "rarity_breakdown.csv"}
	filePath := filepath.Join(dataDir, "csv", "analysis", exporter.FilenameBase)
	return exporter.writeCSV(filePath, rarityBreakdownHeaders(), rows)
}