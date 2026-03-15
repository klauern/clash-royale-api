package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

func formatComparisonJSON(names []string, results []evaluation.EvaluationResult) (string, error) {
	comparison := map[string]any{
		"decks":   names,
		"results": results,
	}

	data, err := json.MarshalIndent(comparison, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data) + "\n", nil
}

//nolint:gocyclo,staticcheck // CSV row construction is intentionally explicit and stable.
func formatComparisonCSV(names []string, results []evaluation.EvaluationResult) string {
	var sb strings.Builder
	w := csv.NewWriter(&sb)

	header := append([]string{"Deck"}, names...)
	_ = w.Write(header)

	overallRow := []string{"Overall Score"}
	for _, r := range results {
		overallRow = append(overallRow, fmt.Sprintf("%.2f", r.OverallScore))
	}
	_ = w.Write(overallRow)

	elixirRow := []string{"Avg Elixir"}
	for _, r := range results {
		elixirRow = append(elixirRow, fmt.Sprintf("%.2f", r.AvgElixir))
	}
	_ = w.Write(elixirRow)

	for _, cat := range getEvaluationCategories() {
		row := []string{cat.name}
		for _, r := range results {
			row = append(row, fmt.Sprintf("%.1f", cat.get(r).Score))
		}
		_ = w.Write(row)
	}

	archetypeRow := []string{"Archetype"}
	for _, r := range results {
		archetypeRow = append(archetypeRow, string(r.DetectedArchetype))
	}
	_ = w.Write(archetypeRow)

	w.Flush()
	return sb.String()
}
