package main

import (
	"strings"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

func TestFormatReportRankingsSortsByOverallScore(t *testing.T) {
	names := []string{"Deck A", "Deck B", "Deck C"}
	results := []evaluation.EvaluationResult{
		{OverallScore: 6.5, OverallRating: "Good"},
		{OverallScore: 8.9, OverallRating: "Excellent"},
		{OverallScore: 7.2, OverallRating: "Strong"},
	}
	vm := buildComparisonViewModel(names, results)

	var sb strings.Builder
	formatReportRankings(&sb, vm)
	report := sb.String()

	first := strings.Index(report, "🥇 **Deck B**")
	second := strings.Index(report, "🥈 **Deck C**")
	third := strings.Index(report, "🥉 **Deck A**")
	if first == -1 || second == -1 || third == -1 {
		t.Fatalf("expected sorted ranking lines in report, got:\n%s", report)
	}
	if !(first < second && second < third) {
		t.Fatalf("expected medal ordering by descending score, got:\n%s", report)
	}
}

func TestBuildComparisonViewModelInvalidInputReturnsExplicitEmptyModel(t *testing.T) {
	vm := buildComparisonViewModel(nil, nil)
	if vm.BestOverallIndex != -1 {
		t.Fatalf("expected BestOverallIndex=-1 for empty input, got %d", vm.BestOverallIndex)
	}
	if len(vm.Decks) != 0 || len(vm.Categories) != 0 || len(vm.RankedDecks) != 0 {
		t.Fatalf("expected empty slices for empty input, got decks=%d categories=%d ranked=%d", len(vm.Decks), len(vm.Categories), len(vm.RankedDecks))
	}

	vm = buildComparisonViewModel([]string{"A"}, nil)
	if vm.BestOverallIndex != -1 {
		t.Fatalf("expected BestOverallIndex=-1 for mismatched input, got %d", vm.BestOverallIndex)
	}
	if len(vm.Decks) != 0 || len(vm.Categories) != 0 || len(vm.RankedDecks) != 0 {
		t.Fatalf("expected empty slices for mismatched input, got decks=%d categories=%d ranked=%d", len(vm.Decks), len(vm.Categories), len(vm.RankedDecks))
	}
}

func TestCompareFormattersHandleEmptyInput(t *testing.T) {
	report := generateComparisonReport(nil, nil)
	if !strings.Contains(report, "No decks to compare.") {
		t.Fatalf("expected empty-report notice, got:\n%s", report)
	}

	table := formatComparisonTable(nil, nil, false, false)
	if !strings.Contains(table, "No decks to compare.") {
		t.Fatalf("expected empty-table notice, got:\n%s", table)
	}

	markdown := formatComparisonMarkdown(nil, nil, false)
	if !strings.Contains(markdown, "*Comparing 0 decks*") {
		t.Fatalf("expected markdown to report zero decks, got:\n%s", markdown)
	}
}
