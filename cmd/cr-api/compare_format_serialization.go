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
func formatComparisonCSV(names []string, results []evaluation.EvaluationResult) (string, error) {
	var sb strings.Builder
	w := csv.NewWriter(&sb)

	header := append([]string{"Deck"}, names...)
	if err := w.Write(header); err != nil {
		return "", err
	}

	overallRow := []string{"Overall Score"}
	for _, r := range results {
		overallRow = append(overallRow, fmt.Sprintf("%.2f", r.OverallScore))
	}
	if err := w.Write(overallRow); err != nil {
		return "", err
	}

	elixirRow := []string{"Avg Elixir"}
	for _, r := range results {
		elixirRow = append(elixirRow, fmt.Sprintf("%.2f", r.AvgElixir))
	}
	if err := w.Write(elixirRow); err != nil {
		return "", err
	}

	for _, cat := range getEvaluationCategories() {
		row := []string{cat.name}
		for _, r := range results {
			row = append(row, fmt.Sprintf("%.1f", cat.get(r).Score))
		}
		if err := w.Write(row); err != nil {
			return "", err
		}
	}

	archetypeRow := []string{"Archetype"}
	for _, r := range results {
		archetypeRow = append(archetypeRow, string(r.DetectedArchetype))
	}
	if err := w.Write(archetypeRow); err != nil {
		return "", err
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return sb.String(), nil
}
