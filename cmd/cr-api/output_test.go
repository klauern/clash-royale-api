package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestWriteCSVRows(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	header := []string{"A", "B"}
	rows := [][]string{
		{"1", "x"},
		{"2", "y"},
	}

	if err := writeCSVDocument(&buf, header, rows); err != nil {
		t.Fatalf("writeCSVRows returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "A,B") {
		t.Fatalf("missing header in CSV output: %q", got)
	}
	if !strings.Contains(got, "1,x") || !strings.Contains(got, "2,y") {
		t.Fatalf("missing row data in CSV output: %q", got)
	}
}

func TestWriteCSVRowsWriteError(t *testing.T) {
	t.Parallel()

	err := writeCSVDocument(failingWriter{}, []string{"A"}, [][]string{{"1"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteTextOutputStdout(t *testing.T) {
	output, err := captureStdout(t, func() error {
		return writeTextOutput("hello world", "", textOutputOptions{})
	})
	if err != nil {
		t.Fatalf("writeTextOutput returned error: %v", err)
	}
	if output != "hello world" {
		t.Fatalf("stdout output = %q, want %q", output, "hello world")
	}
}

func TestWriteTextOutputFileWithSaveMessage(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "out.txt")
	output, err := captureStdout(t, func() error {
		return writeTextOutput("saved content", outputPath, textOutputOptions{
			saveMessage: "Saved to",
		})
	})
	if err != nil {
		t.Fatalf("writeTextOutput returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != "saved content" {
		t.Fatalf("file content = %q, want %q", string(data), "saved content")
	}
	if output != "Saved to: "+outputPath+"\n" {
		t.Fatalf("stdout output = %q", output)
	}
}

func TestWriteTextOutputCreatesParentDirectories(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "nested", "dir", "out.txt")
	if err := writeTextOutput("saved content", outputPath, textOutputOptions{}); err != nil {
		t.Fatalf("writeTextOutput returned error: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != "saved content" {
		t.Fatalf("file content = %q, want %q", string(data), "saved content")
	}
}

func TestWriteTextOutputVerboseOnlySuppressesSaveMessage(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "out.txt")
	output, err := captureStdout(t, func() error {
		return writeTextOutput("saved content", outputPath, textOutputOptions{
			saveMessage: "Saved to",
			verboseOnly: true,
			verbose:     false,
		})
	})
	if err != nil {
		t.Fatalf("writeTextOutput returned error: %v", err)
	}
	if output != "" {
		t.Fatalf("stdout output = %q, want empty string", output)
	}
}
