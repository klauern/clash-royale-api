// Package storage provides file I/O utilities for persisting Clash Royale data.
// Handles JSON serialization/deserialization with proper formatting and error handling.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteJSON writes data to a JSON file with pretty formatting (2-space indentation)
// Creates parent directories if they don't exist
func WriteJSON(filePath string, data interface{}) error {
	// Ensure parent directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// ReadJSON reads and unmarshals a JSON file into the provided data structure
func ReadJSON(filePath string, data interface{}) error {
	// Read file contents
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Unmarshal JSON
	if err := json.Unmarshal(fileData, data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", filePath, err)
	}

	return nil
}

// FileExists checks if a file exists at the given path
func FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirectoryExists checks if a directory exists at the given path
func DirectoryExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// EnsureDirectory creates a directory and all parent directories if they don't exist
func EnsureDirectory(dirPath string) error {
	if DirectoryExists(dirPath) {
		return nil
	}

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	return nil
}

// ListJSONFiles returns a list of all .json files in a directory
// Returns empty slice if directory doesn't exist
func ListJSONFiles(dirPath string) ([]string, error) {
	if !DirectoryExists(dirPath) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	jsonFiles := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			jsonFiles = append(jsonFiles, filepath.Join(dirPath, entry.Name()))
		}
	}

	return jsonFiles, nil
}

// DeleteFile removes a file if it exists
// Returns nil if file doesn't exist (idempotent)
func DeleteFile(filePath string) error {
	if !FileExists(filePath) {
		return nil
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", filePath, err)
	}

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", src, err)
	}

	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := EnsureDirectory(dstDir); err != nil {
		return err
	}

	// Write to destination
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("failed to write destination file %s: %w", dst, err)
	}

	return nil
}

// MoveFile moves a file from src to dst
func MoveFile(src, dst string) error {
	// Copy first
	if err := CopyFile(src, dst); err != nil {
		return err
	}

	// Delete source
	if err := DeleteFile(src); err != nil {
		return fmt.Errorf("failed to delete source file after copy: %w", err)
	}

	return nil
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	return info.Size(), nil
}
