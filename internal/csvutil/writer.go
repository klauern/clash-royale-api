package csvutil

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

// Write writes CSV headers and rows to filePath, creating parent directories.
func Write(filePath string, headers []string, rows [][]string) (returnErr error) {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil && returnErr == nil {
			returnErr = fmt.Errorf("failed to close file: %w", err)
		}
	}()

	writer := csv.NewWriter(file)
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to flush csv: %w", err)
	}

	return nil
}
