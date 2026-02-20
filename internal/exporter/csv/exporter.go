package csv

import (
	"path/filepath"

	"github.com/klauer/clash-royale-api/go/internal/csvutil"
	"github.com/klauer/clash-royale-api/go/internal/storage"
)

// BaseExporter provides common functionality for CSV exporters
type BaseExporter struct {
	FilenameBase string
}

// writeCSV writes data to a CSV file
func (e *BaseExporter) writeCSV(filePath string, headers []string, rows [][]string) (returnErr error) {
	return csvutil.Write(filePath, headers, rows)
}

func (e *BaseExporter) csvFilePath(dataDir, subdir string) string {
	pathBuilder := storage.NewPathBuilder(dataDir)
	return filepath.Join(pathBuilder.GetCSVDir(), subdir, e.FilenameBase)
}

// CSVExporter wraps a BaseExporter to implement the exporter.Exporter interface
type CSVExporter struct {
	BaseExporter
	Headers    func() []string
	Filename   func() string
	ExportFunc func(string, any) error
}

// Export implements the exporter.Exporter interface
func (e *CSVExporter) Export(dataDir string, data any) error {
	return e.ExportFunc(dataDir, data)
}

// NewCSVExporter creates a new CSV exporter with the given functions
func NewCSVExporter(filename string, headers func() []string, exportFunc func(string, any) error) *CSVExporter {
	return &CSVExporter{
		BaseExporter: BaseExporter{
			FilenameBase: filename,
		},
		Headers:    headers,
		Filename:   func() string { return filename },
		ExportFunc: exportFunc,
	}
}
