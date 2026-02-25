package csvutil

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

type limitedFailWriter struct {
	remaining int
}

func (w *limitedFailWriter) Write(p []byte) (int, error) {
	if w.remaining <= 0 {
		return 0, errors.New("write failed")
	}

	if len(p) > w.remaining {
		n := w.remaining
		w.remaining = 0
		return n, nil
	}

	w.remaining -= len(p)
	return len(p), nil
}

func TestWriteTo(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	headers := []string{"A", "B"}
	rows := [][]string{
		{"1", "x"},
		{"2", "y"},
	}

	if err := WriteTo(&buf, headers, rows); err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "A,B") {
		t.Fatalf("missing header in CSV output: %q", got)
	}
	if !strings.Contains(got, "1,x") || !strings.Contains(got, "2,y") {
		t.Fatalf("missing row data in CSV output: %q", got)
	}
}

func TestWriteToWriteError(t *testing.T) {
	t.Parallel()

	err := WriteTo(failingWriter{}, []string{"A"}, [][]string{{"1"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteToMidWriteError(t *testing.T) {
	t.Parallel()

	rowPayload := strings.Repeat("x", 5000)
	err := WriteTo(&limitedFailWriter{remaining: 8}, []string{"A"}, [][]string{{rowPayload}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
