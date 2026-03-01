package main

import "testing"

func TestExportTargetFileUsesExporterFilename(t *testing.T) {
	t.Parallel()

	got := exportTargetFile("/tmp/analysis", "card_analysis.csv")
	want := "/tmp/analysis/card_analysis.csv"
	if got != want {
		t.Fatalf("exportTargetFile() = %q, want %q", got, want)
	}
}

func TestExportTargetFilePreservesBattleLogFilename(t *testing.T) {
	t.Parallel()

	got := exportTargetFile("/tmp/battles", "battle_log.csv")
	want := "/tmp/battles/battle_log.csv"
	if got != want {
		t.Fatalf("exportTargetFile() = %q, want %q", got, want)
	}
}
