package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

// BaseExporter provides common functionality for CSV exporters
type BaseExporter struct {
	FilenameBase string
}

// writeCSV writes data to a CSV file
func (e *BaseExporter) writeCSV(filePath string, headers []string, rows [][]string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	// Write rows
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// CSVExporter wraps a BaseExporter to implement the exporter.Exporter interface
type CSVExporter struct {
	BaseExporter
	Headers    func() []string
	Filename   func() string
	ExportFunc func(string, interface{}) error
}

// Export implements the exporter.Exporter interface
func (e *CSVExporter) Export(dataDir string, data interface{}) error {
	return e.ExportFunc(dataDir, data)
}

// NewCSVExporter creates a new CSV exporter with the given functions
func NewCSVExporter(filename string, headers func() []string, exportFunc func(string, interface{}) error) *CSVExporter {
	return &CSVExporter{
		BaseExporter: BaseExporter{
			FilenameBase: filename,
		},
		Headers:    headers,
		Filename:   func() string { return filename },
		ExportFunc: exportFunc,
	}
}