package main

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExportTargetFileUsesExporterFilename(t *testing.T) {
	t.Parallel()

	got := exportTargetFile("/tmp/analysis", "card_analysis.csv")
	want := filepath.Join("/tmp/analysis", "card_analysis.csv")
	if got != want {
		t.Fatalf("exportTargetFile() = %q, want %q", got, want)
	}
}

func TestExporterFuncDelegatesToExporter(t *testing.T) {
	t.Parallel()

	exporter := &stubCSVExportAdapter{}
	fn := exporterFunc(exporter)

	payload := map[string]string{"tag": "#TEST"}
	if err := fn("/tmp/data", payload); err != nil {
		t.Fatalf("exporterFunc() error = %v", err)
	}
	if exporter.dataDir != "/tmp/data" {
		t.Fatalf("exporter dataDir = %q, want %q", exporter.dataDir, "/tmp/data")
	}
	if !reflect.DeepEqual(exporter.data, payload) {
		t.Fatalf("exporter data = %#v, want %#v", exporter.data, payload)
	}
}

func TestExporterFuncReturnsExporterError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")
	exporter := &stubCSVExportAdapter{err: wantErr}
	fn := exporterFunc(exporter)

	if err := fn("/tmp/data", "payload"); !errors.Is(err, wantErr) {
		t.Fatalf("exporterFunc() error = %v, want %v", err, wantErr)
	}
}

type stubCSVExportAdapter struct {
	dataDir string
	data    any
	err     error
}

func (s *stubCSVExportAdapter) Export(dataDir string, data any) error {
	s.dataDir = dataDir
	s.data = data
	return s.err
}

func (s *stubCSVExportAdapter) Filename() string {
	return "stub.csv"
}

func TestExportTargetFilePreservesBattleLogFilename(t *testing.T) {
	t.Parallel()

	got := exportTargetFile("/tmp/battles", "battle_log.csv")
	want := filepath.Join("/tmp/battles", "battle_log.csv")
	if got != want {
		t.Fatalf("exportTargetFile() = %q, want %q", got, want)
	}
}
