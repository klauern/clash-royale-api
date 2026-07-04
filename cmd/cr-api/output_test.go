package main

import (
	"bytes"
	"errors"
	"io"
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

func TestWriteTextOutputToStdout(t *testing.T) {
	stdout := captureStdoutText(t, func() {
		if err := writeTextOutput("hello", "", false, textOutputOptions{}); err != nil {
			t.Fatalf("writeTextOutput returned error: %v", err)
		}
	})

	if stdout != "hello" {
		t.Fatalf("stdout = %q, want %q", stdout, "hello")
	}
}

func TestWriteTextOutputToFile(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "out.txt")
	stdout := captureStdoutText(t, func() {
		if err := writeTextOutput("saved", outputPath, false, textOutputOptions{
			saveMessage: "Saved to: %s",
		}); err != nil {
			t.Fatalf("writeTextOutput returned error: %v", err)
		}
	})

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "saved" {
		t.Fatalf("file contents = %q, want %q", string(data), "saved")
	}
	if stdout != "Saved to: "+outputPath+"\n" {
		t.Fatalf("stdout = %q", stdout)
	}
}

func TestWriteTextOutputVerboseOnlyNotice(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "out.txt")
	stdout := captureStdoutText(t, func() {
		if err := writeTextOutput("saved", outputPath, false, textOutputOptions{
			saveMessage: "Saved to: %s",
			verboseOnly: true,
		}); err != nil {
			t.Fatalf("writeTextOutput returned error: %v", err)
		}
	})
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty output when verbose disabled", stdout)
	}

	stdout = captureStdoutText(t, func() {
		if err := writeTextOutput("saved-again", outputPath, true, textOutputOptions{
			saveMessage: "Saved to: %s",
			verboseOnly: true,
		}); err != nil {
			t.Fatalf("writeTextOutput returned error: %v", err)
		}
	})
	if stdout != "Saved to: "+outputPath+"\n" {
		t.Fatalf("stdout = %q", stdout)
	}
}

func captureStdoutText(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = originalStdout
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	return string(data)
}
