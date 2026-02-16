package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		data     any
		wantErr  bool
	}{
		{
			name:     "Write simple struct",
			filePath: "/tmp/test_simple.json",
			data: map[string]string{
				"name":  "test",
				"value": "123",
			},
			wantErr: false,
		},
		{
			name:     "Write complex struct",
			filePath: "/tmp/test_complex.json",
			data: &clashroyale.Player{
				Tag:      "#ABC123",
				Name:     "Test Player",
				ExpLevel: 50,
				Trophies: 4000,
			},
			wantErr: false,
		},
		{
			name:     "Write array",
			filePath: "/tmp/test_array.json",
			data:     []string{"item1", "item2", "item3"},
			wantErr:  false,
		},
		{
			name:     "Write nil data",
			filePath: "/tmp/test_nil.json",
			data:     nil,
			wantErr:  false,
		},
		{
			name:     "Write to nested directory",
			filePath: "/tmp/nested/dir/test.json",
			data:     map[string]int{"count": 42},
			wantErr:  false,
		},
		{
			name:     "Invalid path (read-only root)",
			filePath: "/root/test.json",
			data:     map[string]string{"test": "data"},
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Clean up any existing file
			_ = os.Remove(test.filePath)

			err := WriteJSON(test.filePath, test.data)

			if (err != nil) != test.wantErr {
				t.Errorf("WriteJSON() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			if !test.wantErr {
				// Verify file was created
				if _, err := os.Stat(test.filePath); os.IsNotExist(err) {
					t.Errorf("WriteJSON() file was not created")
					return
				}

				// Read and verify content is valid JSON
				fileData, err := os.ReadFile(test.filePath)
				if err != nil {
					t.Errorf("WriteJSON() failed to read created file: %v", err)
					return
				}

				// Check if it's valid JSON
				var temp any
				if err := json.Unmarshal(fileData, &temp); err != nil {
					t.Errorf("WriteJSON() created invalid JSON: %v", err)
				}

				// Clean up
				_ = os.Remove(test.filePath)
			}
		})
	}
}

func TestWriteJSON_PrettyFormatting(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "pretty.json")
	data := map[string]any{
		"nested": map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		"array": []int{1, 2, 3},
	}

	err := WriteJSON(tempFile, data)
	if err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	// Read file content
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Check for 2-space indentation
	contentStr := string(content)
	if !contains(contentStr, "  \"key1\"") {
		t.Error("WriteJSON() should use 2-space indentation")
	}
}

func TestReadJSON(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		data     any
		wantErr  bool
	}{
		{
			name:     "Read valid JSON",
			filePath: "/tmp/test_read.json",
			data:     map[string]string{"test": "value"},
			wantErr:  false,
		},
		{
			name:     "Read non-existent file",
			filePath: "/tmp/nonexistent.json",
			data:     map[string]string{},
			wantErr:  true,
		},
		{
			name:     "Read invalid JSON",
			filePath: "/tmp/invalid.json",
			data:     map[string]string{},
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.name == "Read valid JSON" {
				// Create file with valid JSON
				_ = WriteJSON(test.filePath, map[string]string{"test": "value"})
			} else if test.name == "Read invalid JSON" {
				// Create file with invalid JSON
				_ = os.WriteFile(test.filePath, []byte("{invalid json"), 0o644)
			}

			var result map[string]string
			err := ReadJSON(test.filePath, &result)

			if (err != nil) != test.wantErr {
				t.Errorf("ReadJSON() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			if !test.wantErr && test.name == "Read valid JSON" {
				if result["test"] != "value" {
					t.Errorf("ReadJSON() = %v, want map[test:value]", result)
				}
			}

			// Clean up
			_ = os.Remove(test.filePath)
		})
	}
}

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "exists.txt")
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
	dirPath := tempDir

	// Create existing file
	_ = os.WriteFile(existingFile, []byte("test"), 0o644)

	tests := []struct {
		name     string
		filePath string
		want     bool
	}{
		{
			name:     "Existing file",
			filePath: existingFile,
			want:     true,
		},
		{
			name:     "Non-existent file",
			filePath: nonExistentFile,
			want:     false,
		},
		{
			name:     "Directory path",
			filePath: dirPath,
			want:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := FileExists(test.filePath)
			if result != test.want {
				t.Errorf("FileExists() = %v, want %v", result, test.want)
			}
		})
	}
}

func TestDirectoryExists(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentDir := "/tmp/nonexistent_dir_12345"
	filePath := filepath.Join(tempDir, "file.txt")

	// Create a file
	_ = os.WriteFile(filePath, []byte("test"), 0o644)

	tests := []struct {
		name    string
		dirPath string
		want    bool
	}{
		{
			name:    "Existing directory",
			dirPath: tempDir,
			want:    true,
		},
		{
			name:    "Non-existent directory",
			dirPath: nonExistentDir,
			want:    false,
		},
		{
			name:    "File path",
			dirPath: filePath,
			want:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := DirectoryExists(test.dirPath)
			if result != test.want {
				t.Errorf("DirectoryExists() = %v, want %v", result, test.want)
			}
		})
	}
}

func TestEnsureDirectory(t *testing.T) {
	tests := []struct {
		name    string
		dirPath string
		wantErr bool
	}{
		{
			name:    "Create new directory",
			dirPath: filepath.Join(t.TempDir(), "new_dir"),
			wantErr: false,
		},
		{
			name:    "Create nested directory",
			dirPath: filepath.Join(t.TempDir(), "nested", "deep", "dir"),
			wantErr: false,
		},
		{
			name:    "Directory already exists",
			dirPath: t.TempDir(),
			wantErr: false,
		},
		{
			name:    "Invalid path (read-only root)",
			dirPath: "/root/test_dir",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := EnsureDirectory(test.dirPath)

			if (err != nil) != test.wantErr {
				t.Errorf("EnsureDirectory() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			if !test.wantErr && test.name != "Invalid path" {
				if !DirectoryExists(test.dirPath) {
					t.Errorf("EnsureDirectory() directory was not created")
				}
			}
		})
	}
}

func TestListJSONFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tempDir, "file1.json"), []byte("{}"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("text"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "file3.json"), []byte("{}"), 0o644)
	_ = os.WriteFile(filepath.Join(tempDir, "file.json.bak"), []byte("{}"), 0o644)

	// Create subdirectory with JSON
	subDir := filepath.Join(tempDir, "subdir")
	_ = os.Mkdir(subDir, 0o755)
	_ = os.WriteFile(filepath.Join(subDir, "subfile.json"), []byte("{}"), 0o644)

	tests := []struct {
		name    string
		dirPath string
		want    int // Expected number of JSON files
	}{
		{
			name:    "Directory with JSON files",
			dirPath: tempDir,
			want:    2,
		},
		{
			name:    "Non-existent directory",
			dirPath: "/tmp/nonexistent_12345",
			want:    0,
		},
		{
			name:    "Empty directory",
			dirPath: subDir,
			want:    1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			files, err := ListJSONFiles(test.dirPath)
			if err != nil {
				t.Errorf("ListJSONFiles() error = %v", err)
				return
			}

			if len(files) != test.want {
				t.Errorf("ListJSONFiles() = %d files, want %d", len(files), test.want)
			}

			// Verify all returned files are JSON
			for _, file := range files {
				if filepath.Ext(file) != ".json" {
					t.Errorf("ListJSONFiles() returned non-JSON file: %s", file)
				}
			}
		})
	}
}

func TestDeleteFile(t *testing.T) {
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "exists.txt")
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

	// Create existing file
	_ = os.WriteFile(existingFile, []byte("test"), 0o644)

	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "Delete existing file",
			filePath: existingFile,
			wantErr:  false,
		},
		{
			name:     "Delete non-existent file",
			filePath: nonExistentFile,
			wantErr:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := DeleteFile(test.filePath)

			if (err != nil) != test.wantErr {
				t.Errorf("DeleteFile() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			// File should not exist after deletion
			if FileExists(test.filePath) {
				t.Errorf("DeleteFile() file still exists")
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "dest.txt")
	nestedDst := filepath.Join(tempDir, "nested", "dest.txt")

	// Create source file
	content := []byte("test content for copying")
	_ = os.WriteFile(srcFile, content, 0o644)

	tests := []struct {
		name    string
		src     string
		dst     string
		wantErr bool
	}{
		{
			name:    "Copy existing file",
			src:     srcFile,
			dst:     dstFile,
			wantErr: false,
		},
		{
			name:    "Copy to nested directory",
			src:     srcFile,
			dst:     nestedDst,
			wantErr: false,
		},
		{
			name:    "Copy non-existent source",
			src:     "/tmp/nonexistent_12345.txt",
			dst:     dstFile,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Clean up destination if it exists
			_ = os.Remove(test.dst)

			err := CopyFile(test.src, test.dst)

			if (err != nil) != test.wantErr {
				t.Errorf("CopyFile() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			if !test.wantErr {
				if !FileExists(test.dst) {
					t.Errorf("CopyFile() destination file was not created")
					return
				}

				// Verify content
				dstContent, err := os.ReadFile(test.dst)
				if err != nil {
					t.Errorf("CopyFile() failed to read destination: %v", err)
					return
				}

				if string(dstContent) != string(content) {
					t.Errorf("CopyFile() content mismatch")
				}
			}
		})
	}
}

func TestMoveFile(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "dest.txt")

	// Create source file
	content := []byte("test content for moving")
	_ = os.WriteFile(srcFile, content, 0o644)

	err := MoveFile(srcFile, dstFile)
	if err != nil {
		t.Fatalf("MoveFile() error = %v", err)
	}

	// Source should not exist
	if FileExists(srcFile) {
		t.Error("MoveFile() source file still exists")
	}

	// Destination should exist
	if !FileExists(dstFile) {
		t.Error("MoveFile() destination file was not created")
	}

	// Verify content
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Errorf("MoveFile() failed to read destination: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Error("MoveFile() content mismatch")
	}
}

func TestGetFileSize(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Create file with known content
	content := []byte("test content for size")
	_ = os.WriteFile(testFile, content, 0o644)

	size, err := GetFileSize(testFile)
	if err != nil {
		t.Errorf("GetFileSize() error = %v", err)
	}

	expectedSize := int64(len(content))
	if size != expectedSize {
		t.Errorf("GetFileSize() = %d, want %d", size, expectedSize)
	}

	// Test non-existent file
	_, err = GetFileSize("/tmp/nonexistent_12345.txt")
	if err == nil {
		t.Error("GetFileSize() should return error for non-existent file")
	}
}

func TestWriteReadRoundTrip(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "roundtrip.json")

	// Original data
	original := &clashroyale.Player{
		Tag:          "#ROUNDTRIP",
		Name:         "Round Trip Test",
		ExpLevel:     75,
		Trophies:     5000,
		BestTrophies: 5500,
		Wins:         3000,
		Losses:       2000,
		CreatedAt:    time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Write to JSON
	err := WriteJSON(tempFile, original)
	if err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	// Read from JSON
	var loaded clashroyale.Player
	err = ReadJSON(tempFile, &loaded)
	if err != nil {
		t.Fatalf("ReadJSON() error = %v", err)
	}

	// Compare
	if loaded.Tag != original.Tag {
		t.Errorf("Round trip tag mismatch: got %v, want %v", loaded.Tag, original.Tag)
	}
	if loaded.Name != original.Name {
		t.Errorf("Round trip name mismatch: got %v, want %v", loaded.Name, original.Name)
	}
	if loaded.ExpLevel != original.ExpLevel {
		t.Errorf("Round trip expLevel mismatch: got %v, want %v", loaded.ExpLevel, original.ExpLevel)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(hasPrefix(s, substr) || hasSuffix(s, substr) || indexOf(s, substr) >= 0))
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Benchmark tests
func BenchmarkWriteJSON(b *testing.B) {
	data := map[string]any{
		"players": make([]map[string]any, 100),
		"metadata": map[string]string{
			"version": "1.0",
			"created": time.Now().Format(time.RFC3339),
		},
	}

	// Populate players
	for i := range 100 {
		data["players"].([]map[string]any)[i] = map[string]any{
			"id":    i,
			"name":  fmt.Sprintf("Player%d", i),
			"score": i * 100,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filePath := fmt.Sprintf("/tmp/bench_write_%d.json", i)
		err := WriteJSON(filePath, data)
		if err != nil {
			b.Fatalf("WriteJSON failed: %v", err)
		}
		os.Remove(filePath)
	}
}

func BenchmarkReadJSON(b *testing.B) {
	// Create a test file
	data := map[string]any{
		"numbers": make([]int, 1000),
	}
	for i := range 1000 {
		data["numbers"].([]int)[i] = i
	}

	testFile := "/tmp/bench_read_test.json"
	_ = WriteJSON(testFile, data)
	defer os.Remove(testFile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]any
		err := ReadJSON(testFile, &result)
		if err != nil {
			b.Fatalf("ReadJSON failed: %v", err)
		}
	}
}
