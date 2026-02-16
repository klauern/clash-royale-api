package exporter

import (
	"errors"
	"testing"
)

type fakeExporter struct {
	headers  []string
	filename string
	lastDir  string
	lastData any
	err      error
}

func (f *fakeExporter) Export(dataDir string, data any) error {
	f.lastDir = dataDir
	f.lastData = data
	return f.err
}

func (f *fakeExporter) Headers() []string {
	return f.headers
}

func (f *fakeExporter) Filename() string {
	return f.filename
}

func TestManager_RegisterAndGet(t *testing.T) {
	manager := NewManager()
	exporter := &fakeExporter{filename: "test.csv"}

	manager.Register("csv", exporter)

	got, err := manager.Get("csv")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != exporter {
		t.Fatalf("Get() exporter = %v, want %v", got, exporter)
	}
}

func TestManager_Register_DuplicatePanics(t *testing.T) {
	manager := NewManager()
	exporter := &fakeExporter{}

	manager.Register("csv", exporter)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Register() did not panic on duplicate format")
		}
	}()

	manager.Register("csv", exporter)
}

func TestManager_Get_MissingFormat(t *testing.T) {
	manager := NewManager()

	_, err := manager.Get("missing")
	if err == nil {
		t.Fatal("Get() error = nil, want error for missing format")
	}
}

func TestManager_Export(t *testing.T) {
	manager := NewManager()
	exporter := &fakeExporter{}
	manager.Register("csv", exporter)

	err := manager.Export("csv", "/tmp/data", "payload")
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if exporter.lastDir != "/tmp/data" {
		t.Fatalf("Export() dataDir = %v, want %v", exporter.lastDir, "/tmp/data")
	}
	if exporter.lastData != "payload" {
		t.Fatalf("Export() data = %v, want %v", exporter.lastData, "payload")
	}
}

func TestManager_Export_MissingFormat(t *testing.T) {
	manager := NewManager()

	err := manager.Export("missing", "/tmp/data", "payload")
	if err == nil {
		t.Fatal("Export() error = nil, want error for missing format")
	}
}

func TestManager_Export_PropagatesError(t *testing.T) {
	manager := NewManager()
	expectedErr := errors.New("export failed")
	exporter := &fakeExporter{err: expectedErr}
	manager.Register("csv", exporter)

	err := manager.Export("csv", "/tmp/data", "payload")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("Export() error = %v, want %v", err, expectedErr)
	}
}

func TestManager_ListFormats(t *testing.T) {
	manager := NewManager()
	manager.Register("csv", &fakeExporter{})
	manager.Register("json", &fakeExporter{})

	formats := manager.ListFormats()
	if len(formats) != 2 {
		t.Fatalf("ListFormats() len = %d, want %d", len(formats), 2)
	}

	found := map[string]bool{}
	for _, format := range formats {
		found[format] = true
	}
	if !found["csv"] || !found["json"] {
		t.Fatalf("ListFormats() = %v, want csv and json", formats)
	}
}
