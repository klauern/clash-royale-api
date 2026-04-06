package csv

import (
	"fmt"

	"github.com/klauer/clash-royale-api/go/internal/storage"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

const cardsFilename = "cards.csv"

// NewCardsExporter creates a new cards CSV exporter
func NewCardsExporter() *CSVExporter {
	return NewCSVExporter(
		cardsFilename,
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
func cardsExport(dataDir string, data any) error {
	cards, err := assertCSVExportType[[]clashroyale.Card](data)
	if err != nil {
		return err
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
	return writeCSVRows(dataDir, storage.CSVReferenceSubdir, cardsFilename, cardsHeaders(), rows)
}
