package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/urfave/cli/v3"
)

const (
	compareFormatTable    = "table"
	compareFormatHuman    = "human"
	compareFormatJSON     = "json"
	compareFormatCSV      = "csv"
	compareFormatMarkdown = "markdown"
	compareFormatMD       = "md"
)

// addCompareCommands adds deck comparison subcommands to the CLI
func addCompareCommands() *cli.Command {
	return &cli.Command{
		Name:  "compare",
		Usage: "Compare multiple decks side-by-side with detailed analysis",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "decks",
				Aliases:  []string{"d"},
				Usage:    "Decks to compare (format: \"card1-card2-...-card8\", can specify multiple)",
				Required: false,
			},
			&cli.StringSliceFlag{
				Name:  "names",
				Usage: "Custom names for each deck (optional, defaults to Deck #1, Deck #2, etc.)",
			},
			&cli.StringSliceFlag{
				Name:    "from-evaluations",
				Aliases: []string{"e"},
				Usage:   "Load decks from evaluation batch results (JSON files from 'deck evaluate-batch')",
			},
			&cli.IntFlag{
				Name:  "auto-select-top",
				Usage: "Automatically select and compare top N decks by score from evaluation results",
				Value: 0,
			},
			&cli.StringFlag{
				Name:  "format",
				Value: compareFormatTable,
				Usage: "Output format: table, json, csv, markdown",
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Output file path (optional, prints to stdout if not specified)",
			},
			&cli.StringFlag{
				Name:  "report-output",
				Usage: "Generate comprehensive markdown report to file",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show detailed comparison including all analysis sections",
			},
			&cli.BoolFlag{
				Name:  "winrate",
				Usage: "Show predicted win rate comparison (requires meta data)",
			},
		},
		Action: deckCompareCommand,
	}
}

// batchEvalResult matches the structure from deck evaluate-batch
type batchEvalResult struct {
	Name      string                      `json:"name"`
	Strategy  string                      `json:"strategy"`
	Deck      []string                    `json:"deck"`
	Result    evaluation.EvaluationResult `json:"Result"`
	FilePath  string                      `json:"FilePath"`
	Evaluated string                      `json:"Evaluated"`
	Duration  int64                       `json:"Duration"`
}

// deckCompareCommand compares multiple decks side-by-side
func deckCompareCommand(ctx context.Context, cmd *cli.Command) error {
	deckStrings := cmd.StringSlice("decks")
	names := cmd.StringSlice("names")
	evalFiles := cmd.StringSlice("from-evaluations")
	autoSelectTop := cmd.Int("auto-select-top")
	format := cmd.String("format")
	outputFile := cmd.String("output")
	reportOutput := cmd.String("report-output")
	verbose := cmd.Bool("verbose")
	showWinRate := cmd.Bool("winrate")

	if err := validateComparisonInputs(deckStrings, evalFiles); err != nil {
		return err
	}

	deckNames, results, err := loadDecksForComparison(evalFiles, deckStrings, names, autoSelectTop)
	if err != nil {
		return err
	}

	if err := validateLoadedDecks(results); err != nil {
		return err
	}

	formattedOutput, err := formatDeckComparisonOutput(format, deckNames, results, verbose, showWinRate)
	if err != nil {
		return err
	}

	if err := writeComparisonOutput(outputFile, formattedOutput); err != nil {
		return err
	}

	if reportOutput != "" {
		report := generateComparisonReport(deckNames, results)
		if err := os.WriteFile(reportOutput, []byte(report), 0o644); err != nil {
			return fmt.Errorf("failed to write report: %w", err)
		}
		printf("Comprehensive report saved to: %s\n", reportOutput)
	}

	return nil
}

func validateComparisonInputs(deckStrings, evalFiles []string) error {
	if len(deckStrings) == 0 && len(evalFiles) == 0 {
		return fmt.Errorf("must provide either --decks or --from-evaluations")
	}

	if len(deckStrings) > 0 && len(evalFiles) > 0 {
		return fmt.Errorf("cannot use both --decks and --from-evaluations")
	}

	return nil
}

func loadDecksForComparison(evalFiles, deckStrings, names []string, autoSelectTop int) ([]string, []evaluation.EvaluationResult, error) {
	if len(evalFiles) > 0 {
		deckNames, results, err := loadDecksFromEvaluations(evalFiles, autoSelectTop)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load evaluations: %w", err)
		}
		return deckNames, results, nil
	}

	if len(deckStrings) < 2 {
		return nil, nil, fmt.Errorf("at least 2 decks are required for comparison, got %d", len(deckStrings))
	}

	if len(deckStrings) > 5 {
		return nil, nil, fmt.Errorf("maximum 5 decks can be compared at once, got %d", len(deckStrings))
	}

	deckCards := make([][]deck.CardCandidate, len(deckStrings))
	deckNames := make([]string, len(deckStrings))

	for i, deckStr := range deckStrings {
		cardNames := parseDeckString(deckStr)
		if len(cardNames) != 8 {
			return nil, nil, fmt.Errorf("deck #%d must contain exactly 8 cards, got %d", i+1, len(cardNames))
		}
		deckCards[i] = convertToCardCandidates(cardNames)

		if i < len(names) && names[i] != "" {
			deckNames[i] = names[i]
		} else {
			deckNames[i] = fmt.Sprintf("Deck #%d", i+1)
		}
	}

	synergyDB := deck.NewSynergyDatabase()
	results := make([]evaluation.EvaluationResult, len(deckCards))
	for i, cards := range deckCards {
		results[i] = evaluation.Evaluate(cards, synergyDB, nil)
	}

	return deckNames, results, nil
}

func validateLoadedDecks(results []evaluation.EvaluationResult) error {
	if len(results) < 2 {
		return fmt.Errorf("at least 2 decks are required for comparison, got %d", len(results))
	}

	if len(results) > 5 {
		return fmt.Errorf("maximum 5 decks can be compared at once, got %d (use --auto-select-top to limit)", len(results))
	}

	return nil
}

func formatDeckComparisonOutput(format string, deckNames []string, results []evaluation.EvaluationResult, verbose, showWinRate bool) (string, error) {
	switch strings.ToLower(format) {
	case compareFormatTable, compareFormatHuman:
		return formatComparisonTable(deckNames, results, verbose, showWinRate), nil
	case compareFormatJSON:
		output, err := formatComparisonJSON(deckNames, results)
		if err != nil {
			return "", fmt.Errorf("failed to format JSON: %w", err)
		}
		return output, nil
	case compareFormatCSV:
		return formatComparisonCSV(deckNames, results)
	case compareFormatMarkdown, compareFormatMD:
		return formatComparisonMarkdown(deckNames, results, verbose), nil
	default:
		return "", fmt.Errorf("unknown format: %s (supported: table, json, csv, markdown)", format)
	}
}

func writeComparisonOutput(outputFile, content string) error {
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		printf("Comparison saved to: %s\n", outputFile)
	} else {
		fmt.Print(content)
	}
	return nil
}

func loadDecksFromEvaluations(evalFiles []string, autoSelectTop int) ([]string, []evaluation.EvaluationResult, error) {
	type batchResultFile struct {
		Version        string `json:"version"`
		Timestamp      string `json:"timestamp"`
		EvaluationInfo struct {
			TotalDecks int    `json:"total_decks"`
			Evaluated  int    `json:"evaluated"`
			SortBy     string `json:"sort_by"`
		} `json:"evaluation_info"`
		Results []batchEvalResult `json:"results"`
	}

	var allResults []batchEvalResult

	for _, evalFile := range evalFiles {
		matches, err := filepath.Glob(evalFile)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid pattern %s: %w", evalFile, err)
		}

		if len(matches) == 0 {
			matches = []string{evalFile}
		}

		for _, file := range matches {
			data, err := os.ReadFile(file)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read %s: %w", file, err)
			}

			var batchData batchResultFile
			if err := json.Unmarshal(data, &batchData); err != nil {
				return nil, nil, fmt.Errorf("failed to parse %s: %w", file, err)
			}

			allResults = append(allResults, batchData.Results...)
		}
	}

	if len(allResults) == 0 {
		return nil, nil, fmt.Errorf("no decks found in evaluation files")
	}

	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Result.OverallScore > allResults[j].Result.OverallScore
	})

	if autoSelectTop > 0 && len(allResults) > autoSelectTop {
		allResults = allResults[:autoSelectTop]
	}

	names := make([]string, len(allResults))
	results := make([]evaluation.EvaluationResult, len(allResults))

	for i, r := range allResults {
		names[i] = r.Name
		results[i] = r.Result
	}

	return names, results, nil
}
