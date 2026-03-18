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
	vm := buildComparisonViewModel(names, results)

	header := []string{"Deck"}
	for _, deck := range vm.Decks {
		header = append(header, deck.Name)
	}
	if err := w.Write(header); err != nil {
		return "", fmt.Errorf("write comparison csv header: %w", err)
	}

	overallRow := []string{"Overall Score"}
	for _, deck := range vm.Decks {
		overallRow = append(overallRow, fmt.Sprintf("%.2f", deck.Result.OverallScore))
	}
	if err := w.Write(overallRow); err != nil {
		return "", fmt.Errorf("write comparison csv overall score row: %w", err)
	}

	elixirRow := []string{"Avg Elixir"}
	for _, deck := range vm.Decks {
		elixirRow = append(elixirRow, fmt.Sprintf("%.2f", deck.Result.AvgElixir))
	}
	if err := w.Write(elixirRow); err != nil {
		return "", fmt.Errorf("write comparison csv avg elixir row: %w", err)
	}

	for _, category := range vm.Categories {
		row := []string{category.Name}
		for _, score := range category.Scores {
			row = append(row, fmt.Sprintf("%.1f", score.Score))
		}
		if err := w.Write(row); err != nil {
			return "", fmt.Errorf("write comparison csv category row %q: %w", category.Name, err)
		}
	}

	archetypeRow := []string{"Archetype"}
	for _, deck := range vm.Decks {
		archetypeRow = append(archetypeRow, string(deck.Result.DetectedArchetype))
	}
	if err := w.Write(archetypeRow); err != nil {
		return "", fmt.Errorf("write comparison csv archetype row: %w", err)
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("flush comparison csv writer: %w", err)
	}
	return sb.String(), nil
}
