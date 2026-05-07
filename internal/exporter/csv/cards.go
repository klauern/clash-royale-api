package csv

import (
	"fmt"
	"reflect"

	"github.com/klauer/clash-royale-api/go/internal/storage"
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
func cardsExport(dataDir string, data any) error {
	cards, ok := data.([]clashroyale.Card)
	if !ok {
		return csvTypeMismatchError(reflect.TypeOf([]clashroyale.Card(nil)), data)
	}

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

	return writeCSVRows(dataDir, storage.CSVReferenceSubdir, "cards.csv", cardsHeaders(), rows)
}
