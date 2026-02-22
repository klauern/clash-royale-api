package main

import (
	"bytes"
	"errors"
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
