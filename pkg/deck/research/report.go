package research

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauer/clash-royale-api/go/internal/storage"
)

//nolint:gocyclo,funlen // Report sections are intentionally explicit for stable output.
func buildMarkdown(report *BenchmarkReport) string {
	var b bytes.Buffer
	b.WriteString("# Archetype-Free Deck Builder Benchmark\n\n")
	b.WriteString("## Executive Summary\n\n")
	if len(report.Aggregates) > 0 {
		fmt.Fprintf(&b, "Top method: **%s** (mean composite %.3f)\n\n", report.Aggregates[0].Method, report.Aggregates[0].MeanComposite)
	}

	b.WriteString("## Per-Method Aggregate Table\n\n")
	b.WriteString("| Method | Runs | Mean Composite | Median Composite | Mean Runtime (ms) | Constraint Violations |\n")
	b.WriteString("|---|---:|---:|---:|---:|---:|\n")
	for _, a := range report.Aggregates {
		fmt.Fprintf(&b, "| %s | %d | %.3f | %.3f | %.1f | %d |\n", a.Method, a.Runs, a.MeanComposite, a.MedianComposite, a.MeanRuntimeMs, a.ConstraintViolationCount)
	}
	b.WriteString("\n")

	b.WriteString("## Per-Tag Winners\n\n")
	for _, p := range report.PlayerRuns {
		fmt.Fprintf(&b, "- `%s` (%s): %s (%.3f)\n", p.PlayerTag, p.PlayerName, p.Winner, p.WinnerScore)
	}
	b.WriteString("\n")

	b.WriteString("## Per-Tag Outcomes\n\n")
	b.WriteString("| Tag | Method | Composite | Runtime (ms) | Violations |\n")
	b.WriteString("|---|---|---:|---:|---:|\n")
	for _, p := range report.PlayerRuns {
		for _, m := range p.MethodRuns {
			fmt.Fprintf(&b, "| %s | %s | %.3f | %d | %d |\n",
				p.PlayerTag,
				m.Method,
				m.Metrics.Composite,
				m.Metrics.RuntimeMs,
				len(m.Metrics.ConstraintViolations))
		}
	}
	b.WriteString("\n")

	b.WriteString("## Constraint/Failure Analysis\n\n")
	for _, p := range report.PlayerRuns {
		for _, m := range p.MethodRuns {
			if len(m.Metrics.ConstraintViolations) > 0 {
				fmt.Fprintf(&b, "- `%s` / `%s`: %s\n", p.PlayerTag, m.Method, strings.Join(m.Metrics.ConstraintViolations, "; "))
			}
		}
	}
	b.WriteString("\n")

	b.WriteString("## Notable Deck Composition Patterns\n\n")
	for _, p := range report.PlayerRuns {
		if len(p.MethodRuns) == 0 {
			continue
		}
		best := p.MethodRuns[0]
		fmt.Fprintf(&b, "- `%s`: best method `%s`, archetype diagnostic `%s` (%.0f%% conf)\n",
			p.PlayerTag, best.Method, best.DetectedArchetype, best.ArchetypeConfidence*100)
	}
	b.WriteString("\n")

	b.WriteString("## Recommendation\n\n")
	for _, rec := range report.Recommendations {
		fmt.Fprintf(&b, "- %s\n", rec)
	}
	b.WriteString("\n")

	return b.String()
}

// WriteReport writes both benchmark.json and benchmark.md.
func WriteReport(outputDir string, report *BenchmarkReport) (string, string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", "", err
	}
	jsonPath := filepath.Join(outputDir, "benchmark.json")
	mdPath := filepath.Join(outputDir, "benchmark.md")

	if err := storage.WriteJSON(jsonPath, report); err != nil {
		return "", "", fmt.Errorf("failed to write benchmark JSON %s: %w", jsonPath, err)
	}
	if err := os.WriteFile(mdPath, []byte(buildMarkdown(report)), 0o644); err != nil {
		return "", "", fmt.Errorf("failed to write benchmark markdown %s: %w", mdPath, err)
	}
	return jsonPath, mdPath, nil
}
