package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

const (
	fuzzOutputJSON     = "json"
	fuzzOutputCSV      = "csv"
	fuzzOutputDetailed = "detailed"
)

func sortFuzzingResultsImpl(results []FuzzingResult, sortBy string) {
	sort.Slice(results, func(i, j int) bool {
		var iValue, jValue float64

		switch sortBy {
		case "overall":
			iValue = results[i].OverallScore
			jValue = results[j].OverallScore
		case "attack":
			iValue = results[i].AttackScore
			jValue = results[j].AttackScore
		case "defense":
			iValue = results[i].DefenseScore
			jValue = results[j].DefenseScore
		case "synergy":
			iValue = results[i].SynergyScore
			jValue = results[j].SynergyScore
		case "versatility":
			iValue = results[i].VersatilityScore
			jValue = results[j].VersatilityScore
		case "elixir":
			return results[i].AvgElixir < results[j].AvgElixir
		default:
			iValue = results[i].OverallScore
			jValue = results[j].OverallScore
		}

		return iValue > jValue
	})
}

func getTopResultsImpl(results []FuzzingResult, top int) []FuzzingResult {
	if len(results) <= top {
		return results
	}
	return results[:top]
}

func formatFuzzingResultsImpl(
	results []FuzzingResult,
	format string,
	playerName string,
	playerTag string,
	fuzzerConfig *deck.FuzzingConfig,
	mode string,
	generationTime time.Duration,
	stats *deck.FuzzingStats,
	totalFiltered int,
) error {
	switch format {
	case fuzzOutputJSON:
		return formatResultsJSONImpl(results, playerName, playerTag, fuzzerConfig, mode, generationTime, stats, totalFiltered)
	case fuzzOutputCSV:
		return formatResultsCSVImpl(results)
	case fuzzOutputDetailed:
		return formatResultsDetailedImpl(results, playerName, playerTag)
	default:
		return formatResultsSummaryImpl(results, playerName, playerTag, fuzzerConfig, mode, generationTime, stats, totalFiltered)
	}
}

//nolint:funlen // Keeping terminal output formatting in one place improves maintainability.
func formatResultsSummaryImpl(
	results []FuzzingResult,
	playerName string,
	playerTag string,
	fuzzerConfig *deck.FuzzingConfig,
	mode string,
	generationTime time.Duration,
	stats *deck.FuzzingStats,
	totalFiltered int,
) error {
	printf("\nDeck Fuzzing Results for %s (%s)\n", playerName, playerTag)
	printf("Generated %d decks in %v\n", stats.Generated, generationTime.Round(time.Millisecond))
	printf("Configuration:\n")
	if mode != "" {
		printf("  Mode: %s\n", mode)
	}

	if len(fuzzerConfig.IncludeCards) > 0 {
		printf("  Include cards: %s\n", strings.Join(fuzzerConfig.IncludeCards, ", "))
	}
	if len(fuzzerConfig.ExcludeCards) > 0 {
		printf("  Exclude cards: %s\n", strings.Join(fuzzerConfig.ExcludeCards, ", "))
	}
	printf("  Elixir range: %.1f - %.1f\n", fuzzerConfig.MinAvgElixir, fuzzerConfig.MaxAvgElixir)
	if fuzzerConfig.MinOverallScore > 0 {
		printf("  Min overall score: %.1f\n", fuzzerConfig.MinOverallScore)
	}
	if fuzzerConfig.MinSynergyScore > 0 {
		printf("  Min synergy score: %.1f\n", fuzzerConfig.MinSynergyScore)
	}

	printf("\nTop %d Decks (from %d decks passing filters):\n\n", len(results), totalFiltered)

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, ' ', 0)
	fprintln(w, "Rank\tDeck\tOverall\tLadder\tNorm\tAttack\tDefense\tSynergy\tElixir")

	for i, result := range results {
		deckStr := strings.Join(result.Deck, ", ")

		if len(deckStr) > 50 {
			firstLine := strings.Join(result.Deck[:4], ", ")
			fprintf(w, "%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\n",
				i+1,
				firstLine+",",
				result.OverallScore,
				result.LadderScore,
				result.NormalizedScore,
				result.AttackScore,
				result.DefenseScore,
				result.SynergyScore,
				result.AvgElixir,
			)

			secondLine := strings.Join(result.Deck[4:], ", ")
			fprintf(w, "\t%s\n", secondLine)
		} else {
			fprintf(w, "%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\n",
				i+1,
				deckStr,
				result.OverallScore,
				result.LadderScore,
				result.NormalizedScore,
				result.AttackScore,
				result.DefenseScore,
				result.SynergyScore,
				result.AvgElixir,
			)
		}
	}

	flushWriter(w)
	return nil
}

func formatResultsJSONImpl(
	results []FuzzingResult,
	playerName string,
	playerTag string,
	fuzzerConfig *deck.FuzzingConfig,
	mode string,
	generationTime time.Duration,
	stats *deck.FuzzingStats,
	totalFiltered int,
) error {
	output := map[string]any{
		"player_name":             playerName,
		"player_tag":              playerTag,
		"generated":               stats.Generated,
		"success":                 stats.Success,
		"failed":                  stats.Failed,
		"filtered":                totalFiltered,
		"returned":                len(results),
		"generation_time_seconds": generationTime.Seconds(),
		"config": map[string]any{
			"mode":              mode,
			"count":             fuzzerConfig.Count,
			"workers":           fuzzerConfig.Workers,
			"include_cards":     fuzzerConfig.IncludeCards,
			"exclude_cards":     fuzzerConfig.ExcludeCards,
			"min_avg_elixir":    fuzzerConfig.MinAvgElixir,
			"max_avg_elixir":    fuzzerConfig.MaxAvgElixir,
			"min_overall_score": fuzzerConfig.MinOverallScore,
			"min_synergy_score": fuzzerConfig.MinSynergyScore,
		},
		"results": results,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func formatResultsCSVImpl(results []FuzzingResult) error {
	header := []string{"Rank", "Deck", "Overall", "Contextual", "Ladder", "Normalized", "LevelRatio", "NormFactor", "Attack", "Defense", "Synergy", "Versatility", "AvgElixir", "Archetype"}
	rows := make([][]string, 0, len(results))
	for i, result := range results {
		deckStr := strings.Join(result.Deck, ", ")
		rows = append(rows, []string{
			strconv.Itoa(i + 1),
			deckStr,
			fmt.Sprintf("%.2f", result.OverallScore),
			fmt.Sprintf("%.2f", result.ContextualScore),
			fmt.Sprintf("%.2f", result.LadderScore),
			fmt.Sprintf("%.2f", result.NormalizedScore),
			fmt.Sprintf("%.3f", result.DeckLevelRatio),
			fmt.Sprintf("%.3f", result.NormalizationFactor),
			fmt.Sprintf("%.2f", result.AttackScore),
			fmt.Sprintf("%.2f", result.DefenseScore),
			fmt.Sprintf("%.2f", result.SynergyScore),
			fmt.Sprintf("%.2f", result.VersatilityScore),
			fmt.Sprintf("%.2f", result.AvgElixir),
			result.Archetype,
		})
	}
	return writeCSVDocument(os.Stdout, header, rows)
}

func formatResultsDetailedImpl(
	results []FuzzingResult,
	playerName string,
	playerTag string,
) error {
	printf("\nDeck Fuzzing Results for %s (%s)\n", playerName, playerTag)
	printf("\nTop %d Decks:\n\n", len(results))

	for i, result := range results {
		printf("=== Deck %d ===\n", i+1)
		printf("Cards: %s\n", strings.Join(result.Deck, ", "))
		printf("Overall: %.2f | Attack: %.2f | Defense: %.2f | Synergy: %.2f | Versatility: %.2f\n",
			result.OverallScore, result.AttackScore, result.DefenseScore, result.SynergyScore, result.VersatilityScore)
		printf("Contextual: %.2f | Ladder: %.2f | Normalized: %.2f\n",
			result.ContextualScore, result.LadderScore, result.NormalizedScore)
		printf("Level Ratio: %.3f | Normalization Factor: %.3f\n",
			result.DeckLevelRatio, result.NormalizationFactor)
		printf("Avg Elixir: %.2f | Archetype: %s (%.0f%% confidence)\n",
			result.AvgElixir, result.Archetype, result.ArchetypeConfidence*100)
		printf("Evaluated: %s\n\n", result.EvaluatedAt.Format(time.RFC3339))
	}

	return nil
}

func saveResultsToFileImpl(results []FuzzingResult, outputDir, format, playerTag string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102_150405")
	cleanTag := strings.TrimPrefix(playerTag, "#")
	var filename string

	switch format {
	case fuzzOutputJSON:
		filename = fmt.Sprintf("fuzz_%s_%s.json", cleanTag, timestamp)
	case fuzzOutputCSV:
		filename = fmt.Sprintf("fuzz_%s_%s.csv", cleanTag, timestamp)
	default:
		filename = fmt.Sprintf("fuzz_%s_%s.txt", cleanTag, timestamp)
	}

	outputPath := filepath.Join(outputDir, filename)

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer closeFile(file)

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	os.Stdout = file
	os.Stderr = file
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	switch format {
	case fuzzOutputJSON:
		config := &deck.FuzzingConfig{}
		stats := &deck.FuzzingStats{}
		return formatResultsJSONImpl(results, cleanTag, playerTag, config, "unknown", 0, stats, len(results))
	case fuzzOutputCSV:
		return formatResultsCSVImpl(results)
	default:
		return formatResultsSummaryImpl(results, cleanTag, playerTag, &deck.FuzzingConfig{}, "unknown", 0, &deck.FuzzingStats{}, len(results))
	}
}
