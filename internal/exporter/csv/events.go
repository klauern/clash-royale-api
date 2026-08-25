package csv

import (
	"reflect"

	"github.com/klauer/clash-royale-api/go/internal/storage"
	"github.com/klauer/clash-royale-api/go/pkg/events"
)

// NewEventDeckExporter creates a new event deck CSV exporter
func NewEventDeckExporter() *CSVExporter {
	return NewCSVExporter(
		"event_decks.csv",
		eventDeckHeaders,
		eventDeckExport,
	)
}

// eventDeckHeaders returns the CSV headers for event deck data
func eventDeckHeaders() []string {
	return events.EventDeckCSVHeaders()
}

// eventDeckExport exports event deck data to CSV
func eventDeckExport(dataDir string, data any) error {
	collection, ok := data.(*events.EventDeckCollection)
	if !ok {
		return csvTypeMismatchError(reflect.TypeOf((*events.EventDeckCollection)(nil)), data)
	}

	// Prepare CSV rows
	var rows [][]string

	for _, deck := range collection.Decks {
		rows = append(rows, events.EventDeckCSVRow(deck))
	}

	return writeCSVRows(dataDir, storage.CSVEventsSubdir, "event_decks.csv", eventDeckHeaders(), rows)
}
