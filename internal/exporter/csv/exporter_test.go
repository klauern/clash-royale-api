package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestBaseExporter_WriteCSV(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		headers  []string
		rows     [][]string
		wantErr  bool
	}{
		{
			name:     "Valid CSV write",
			filePath: "/tmp/test.csv",
			headers:  []string{"ID", "Name", "Value"},
			rows: [][]string{
				{"1", "Test", "100"},
				{"2", "Another", "200"},
			},
			wantErr: false,
		},
		{
			name:     "Empty rows",
			filePath: "/tmp/empty.csv",
			headers:  []string{"ID", "Name"},
			rows:     [][]string{},
			wantErr:  false,
		},
		{
			name:     "Single row",
			filePath: "/tmp/single.csv",
			headers:  []string{"Column"},
			rows:     [][]string{{"Value"}},
			wantErr:  false,
		},
		{
			name:     "Special characters",
			filePath: "/tmp/special.csv",
			headers:  []string{"Name", "Description"},
			rows: [][]string{
				{"Test, with comma", "Line\nBreak"},
				{"\"Quotes\"", "Tabs\tand\nnewlines"},
			},
			wantErr: false,
		},
		{
			name:     "Invalid path (read-only directory)",
			filePath: "/root/test.csv",
			headers:  []string{"ID"},
			rows:     [][]string{{"1"}},
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Clean up any existing file
			_ = os.Remove(test.filePath)

			exporter := &BaseExporter{}
			err := exporter.writeCSV(test.filePath, test.headers, test.rows)

			if (err != nil) != test.wantErr {
				t.Errorf("writeCSV() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			if !test.wantErr {
				// Verify file was created
				if _, err := os.Stat(test.filePath); os.IsNotExist(err) {
					t.Errorf("writeCSV() file was not created")
					return
				}

				// Read and verify content
				file, err := os.Open(test.filePath)
				if err != nil {
					t.Errorf("writeCSV() failed to open created file: %v", err)
					return
				}
				defer file.Close()

				reader := csv.NewReader(file)
				records, err := reader.ReadAll()
				if err != nil {
					t.Errorf("writeCSV() failed to read created file: %v", err)
					return
				}

				// Check headers
				if len(records) == 0 || !equalSlices(records[0], test.headers) {
					t.Errorf("writeCSV() headers = %v, want %v", records, test.headers)
				}

				// Check rows (skip header row)
				if len(records)-1 != len(test.rows) {
					t.Errorf("writeCSV() row count = %d, want %d", len(records)-1, len(test.rows))
				} else {
					for i, row := range test.rows {
						if !equalSlices(records[i+1], row) {
							t.Errorf("writeCSV() row %d = %v, want %v", i, records[i+1], row)
						}
					}
				}

				// Clean up
				_ = os.Remove(test.filePath)
			}
		})
	}
}

func TestBaseExporter_WriteCSV_DirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "nested", "dir", "test.csv")

	exporter := &BaseExporter{}
	err := exporter.writeCSV(nestedPath, []string{"ID"}, [][]string{{"1"}})
	if err != nil {
		t.Errorf("writeCSV() with nested directories failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("writeCSV() file was not created in nested directory")
	}
}

func TestNewCSVExporter(t *testing.T) {
	filename := "test.csv"
	headers := func() []string { return []string{"ID", "Name"} }
	exportFunc := func(dataDir string, data interface{}) error { return nil }

	exporter := NewCSVExporter(filename, headers, exportFunc)

	if exporter.FilenameBase != filename {
		t.Errorf("NewCSVExporter() FilenameBase = %v, want %v", exporter.FilenameBase, filename)
	}

	if exporter.Filename() != filename {
		t.Errorf("NewCSVExporter() Filename() = %v, want %v", exporter.Filename(), filename)
	}

	// Test that Headers function works
	resultHeaders := exporter.Headers()
	if !equalSlices(resultHeaders, headers()) {
		t.Errorf("NewCSVExporter() Headers = %v, want %v", resultHeaders, headers())
	}
}

func TestCSVExporter_Export(t *testing.T) {
	called := false
	var receivedDataDir string
	var receivedData interface{}

	exportFunc := func(dataDir string, data interface{}) error {
		called = true
		receivedDataDir = dataDir
		receivedData = data
		return nil
	}

	exporter := NewCSVExporter("test.csv", func() []string { return []string{} }, exportFunc)

	err := exporter.Export("/test/dir", "test data")
	if err != nil {
		t.Errorf("Export() error = %v", err)
	}

	if !called {
		t.Error("Export() did not call the export function")
	}

	if receivedDataDir != "/test/dir" {
		t.Errorf("Export() dataDir = %v, want %v", receivedDataDir, "/test/dir")
	}

	if receivedData != "test data" {
		t.Errorf("Export() data = %v, want %v", receivedData, "test data")
	}
}

func TestCSVExporter_Export_Error(t *testing.T) {
	expectedErr := fmt.Errorf("test error")
	exportFunc := func(dataDir string, data interface{}) error {
		return expectedErr
	}

	exporter := NewCSVExporter("test.csv", func() []string { return []string{} }, exportFunc)

	err := exporter.Export("/test/dir", "test data")

	if err != expectedErr {
		t.Errorf("Export() error = %v, want %v", err, expectedErr)
	}
}

// Helper function to compare string slices
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Benchmark tests
func BenchmarkWriteCSV(b *testing.B) {
	exporter := &BaseExporter{}
	headers := []string{"ID", "Name", "Value", "Description"}

	// Create test data
	rows := make([][]string, 100)
	for i := 0; i < 100; i++ {
		rows[i] = []string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("Name%d", i),
			fmt.Sprintf("%d", i*10),
			fmt.Sprintf("Description for item %d with some longer text", i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filePath := fmt.Sprintf("/tmp/bench_%d.csv", i)
		err := exporter.writeCSV(filePath, headers, rows)
		if err != nil {
			b.Fatalf("writeCSV failed: %v", err)
		}
		os.Remove(filePath) // Clean up
	}
}

func BenchmarkNewCSVExporter(b *testing.B) {
	headers := func() []string { return []string{"ID", "Name"} }
	exportFunc := func(dataDir string, data interface{}) error { return nil }

	for i := 0; i < b.N; i++ {
		_ = NewCSVExporter(fmt.Sprintf("test_%d.csv", i), headers, exportFunc)
	}
}

// Test error conditions
func TestWriteCSV_FilePermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping test when running as root")
	}

	exporter := &BaseExporter{}

	// Try to write to a directory that doesn't exist and can't be created
	err := exporter.writeCSV("/proc/readonly/test.csv", []string{"ID"}, [][]string{{"1"}})

	if err == nil {
		t.Error("writeCSV() should have failed with permission error")
	}
}

func TestWriteCSV_EmptyData(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "empty.csv")
	exporter := &BaseExporter{}

	// Write with headers but no rows
	err := exporter.writeCSV(tempFile, []string{"Header1", "Header2"}, [][]string{})
	if err != nil {
		t.Errorf("writeCSV() with empty rows failed: %v", err)
	}

	// Verify file exists and has only headers
	file, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("Expected 1 record (headers only), got %d", len(records))
	}
}

func TestWriteCSV_LargeDataset(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "large.csv")
	exporter := &BaseExporter{}

	// Create a large dataset
	headers := []string{"ID", "Name", "Value"}
	rows := make([][]string, 10000)
	for i := 0; i < 10000; i++ {
		rows[i] = []string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("Item%d", i),
			fmt.Sprintf("%.2f", float64(i)*1.5),
		}
	}

	err := exporter.writeCSV(tempFile, headers, rows)
	if err != nil {
		t.Errorf("writeCSV() with large dataset failed: %v", err)
	}

	// Verify file size is reasonable
	stat, err := os.Stat(tempFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if stat.Size() == 0 {
		t.Error("writeCSV() created empty file")
	}
}
