package csv

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

func TestNewCardsExporter(t *testing.T) {
	exporter := NewCardsExporter()

	if exporter.Filename() != "cards.csv" {
		t.Errorf("NewCardsExporter() filename = %v, want cards.csv", exporter.Filename())
	}

	headers := exporter.Headers()
	expectedHeaders := []string{
		"ID", "Name", "Elixir Cost", "Type", "Rarity",
		"Max Level", "Max Evolution Level", "Description",
	}

	if len(headers) != len(expectedHeaders) {
		t.Errorf("NewCardsExporter() headers count = %d, want %d", len(headers), len(expectedHeaders))
	}

	for i, h := range expectedHeaders {
		if i >= len(headers) || headers[i] != h {
			t.Errorf("NewCardsExporter() header %d = %v, want %v", i, headers[i], h)
		}
	}
}

func TestCardsExport(t *testing.T) {
	tempDir := t.TempDir()

	// Create mock card data
	mockCards := []clashroyale.Card{
		{
			ID:         28000000,
			Name:       "Knight",
			ElixirCost: 3,
			Type:       "troop",
			Rarity:     "common",
			IconUrls: clashroyale.IconUrls{
				Medium: "https://example.com/knight.png",
			},
		},
		{
			ID:         28000001,
			Name:       "Fireball",
			ElixirCost: 4,
			Type:       "spell",
			Rarity:     "rare",
			IconUrls: clashroyale.IconUrls{
				Medium: "https://example.com/fireball.png",
			},
		},
	}

	exporter := NewCardsExporter()
	err := exporter.Export(tempDir, mockCards)
	if err != nil {
		t.Fatalf("CardsExport() error = %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, "csv", "reference", "cards.csv")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("CardsExport() file was not created at %s", filePath)
		return
	}

	// Read and verify content
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open exported file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	// Check that we have headers and data
	if len(records) < 2 {
		t.Errorf("CardsExport() row count = %d, want at least 2 (header + data)", len(records))
	}

	// Check that the first card ID is correct
	if len(records) > 1 {
		firstCard := records[1]
		if len(firstCard) == 0 || firstCard[0] != "28000000" {
			t.Errorf("CardsExport() first card ID = %v, want 28000000", firstCard[0])
		}
	}
}

func TestCardsExport_EmptyData(t *testing.T) {
	tempDir := t.TempDir()

	emptyCards := []clashroyale.Card{}

	exporter := NewCardsExporter()
	err := exporter.Export(tempDir, emptyCards)
	if err != nil {
		t.Fatalf("CardsExport() with empty data error = %v", err)
	}

	// File should still be created with headers
	filePath := filepath.Join(tempDir, "csv", "reference", "cards.csv")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("CardsExport() with empty data should still create file with headers")
	}
}

func TestCardsExport_InvalidDataType(t *testing.T) {
	tempDir := t.TempDir()

	exporter := NewCardsExporter()
	err := exporter.Export(tempDir, "invalid data")

	// Should not crash, but might not create meaningful data
	if err != nil {
		t.Logf("CardsExport() with invalid data returned error (acceptable): %v", err)
	}
}
