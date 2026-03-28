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
	model := buildComparisonRenderModel(names, results)

	var sb strings.Builder
	formatReportRankings(&sb, model)
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
