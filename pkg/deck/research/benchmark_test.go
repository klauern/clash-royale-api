package research

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestBenchmarkRunnerEndToEndWithSmallPool(t *testing.T) {
	builder := deck.NewBuilder("data")
	runner := BenchmarkRunner{Builder: builder}

	players := []PlayerInput{
		{Tag: "#A", Name: "PlayerA", Candidates: testPool()},
	}

	report, err := runner.Run(BenchmarkConfig{
		Seed:    42,
		TopN:    1,
		Methods: []string{MethodBaseline, MethodConstraint, MethodRoleFirst},
		DataDir: "data",
	}, players)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(report.PlayerRuns) != 1 {
		t.Fatalf("expected 1 player run, got %d", len(report.PlayerRuns))
	}
	if len(report.Aggregates) != 3 {
		t.Fatalf("expected 3 aggregates, got %d", len(report.Aggregates))
	}
}

func TestWriteReportCreatesJSONAndMarkdown(t *testing.T) {
	report := &BenchmarkReport{
		Methods: []string{MethodBaseline},
		PlayerRuns: []PlayerResult{
			{PlayerTag: "#A", PlayerName: "A", Winner: MethodBaseline, WinnerScore: 0.7},
		},
		Aggregates:      []BenchmarkAggregate{{Method: MethodBaseline, Runs: 1, MeanComposite: 0.7, MedianComposite: 0.7}},
		Recommendations: []string{"test recommendation"},
	}
	dir := t.TempDir()
	jsonPath, mdPath, err := WriteReport(dir, report)
	if err != nil {
		t.Fatalf("write report failed: %v", err)
	}
	if _, err := os.Stat(jsonPath); err != nil {
		t.Fatalf("json report missing: %v", err)
	}
	if _, err := os.Stat(mdPath); err != nil {
		t.Fatalf("markdown report missing: %v", err)
	}
	if filepath.Base(jsonPath) != "benchmark.json" {
		t.Fatalf("unexpected json name: %s", jsonPath)
	}
}
