package csv

import (
	"fmt"
	"path/filepath"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// NewCardsExporter creates a new cards CSV exporter
func NewCardsExporter() *CSVExporter {
	return NewCSVExporter(
		"cards.csv",
		cardsHeaders,
		cardsExport,
	)
}

// cardsHeaders returns the CSV headers for card data
func cardsHeaders() []string {
	return []string{
		"ID",
		"Name",
		"Elixir Cost",
		"Type",
		"Rarity",
		"Max Level",
		"Max Evolution Level",
		"Description",
	}
}

// cardsExport exports card data to CSV
func cardsExport(dataDir string, data interface{}) error {
	cards, ok := data.([]clashroyale.Card)
	if !ok {
		return fmt.Errorf("expected []Card type, got %T", data)
	}

	// Prepare CSV rows
	var rows [][]string
	for _, card := range cards {
		row := []string{
			fmt.Sprintf("%d", card.ID),
			card.Name,
			fmt.Sprintf("%d", card.ElixirCost),
			card.Type,
			card.Rarity,
			fmt.Sprintf("%d", card.MaxLevel),
			fmt.Sprintf("%d", card.MaxEvolutionLevel),
			card.Description,
		}
		rows = append(rows, row)
	}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "cards.csv"}
	filePath := filepath.Join(dataDir, "csv", "reference", exporter.FilenameBase)
	return exporter.writeCSV(filePath, cardsHeaders(), rows)
}