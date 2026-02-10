package main

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

func exportOptimizationCSV(
	path string,
	playerTag string,
	currentDeck []string,
	currentResult evaluation.EvaluationResult,
	alternatives evaluation.AlternativeSuggestions,
) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer closeFile(file)

	// Write header
	fprintln(file, "# DECK OPTIMIZATION RESULTS")
	fprintf(file, "Player Tag,%s\n", playerTag)
	fprintf(file, "Current Deck,%s\n", strings.Join(currentDeck, ";"))
	fprintf(file, "Current Score,%.2f\n", currentResult.OverallScore)
	fprintf(file, "Archetype,%s\n", currentResult.DetectedArchetype)
	fprintln(file, "")

	// Write category scores
	fprintln(file, "# CURRENT CATEGORY SCORES")
	fprintln(file, "Category,Score,Rating,Stars")
	fprintf(file, "Attack,%.2f,%s,%d\n",
		currentResult.Attack.Score,
		currentResult.Attack.Rating,
		currentResult.Attack.Stars)
	fprintf(file, "Defense,%.2f,%s,%d\n",
		currentResult.Defense.Score,
		currentResult.Defense.Rating,
		currentResult.Defense.Stars)
	fprintf(file, "Synergy,%.2f,%s,%d\n",
		currentResult.Synergy.Score,
		currentResult.Synergy.Rating,
		currentResult.Synergy.Stars)
	fprintf(file, "Versatility,%.2f,%s,%d\n",
		currentResult.Versatility.Score,
		currentResult.Versatility.Rating,
		currentResult.Versatility.Stars)
	fprintf(file, "F2P Friendly,%.2f,%s,%d\n",
		currentResult.F2PFriendly.Score,
		currentResult.F2PFriendly.Rating,
		currentResult.F2PFriendly.Stars)
	fprintln(file, "")

	// Write optimization suggestions
	fprintln(file, "# OPTIMIZATION SUGGESTIONS")
	fprintln(file, "Rank,Original Card,Replacement Card,Score Before,Score After,Improvement,Impact,Rationale,New Deck")
	for i, alt := range alternatives.Suggestions {
		fprintf(file, "%d,%s,%s,%.2f,%.2f,%.2f,%s,\"%s\",%s\n",
			i+1,
			alt.OriginalCard,
			alt.ReplacementCard,
			alt.OriginalScore,
			alt.NewScore,
			alt.ScoreDelta,
			alt.Impact,
			alt.Rationale,
			strings.Join(alt.Deck, ";"))
	}

	return nil
}

// sortEvaluationResults sorts batch evaluation results by the specified criteria
func sortEvaluationResults[T any](results []T, sortBy string) {
	if len(results) < 2 {
		return
	}

	type resultInterface interface {
		GetResult() evaluation.EvaluationResult
	}

	// Type assertion helper
	getResult := func(r T) evaluation.EvaluationResult {
		switch v := any(r).(type) {
		case resultInterface:
			return v.GetResult()
		default:
			rv := reflect.ValueOf(r)
			if rv.Kind() == reflect.Pointer {
				rv = rv.Elem()
			}
			if rv.IsValid() && rv.Kind() == reflect.Struct {
				field := rv.FieldByName("Result")
				if field.IsValid() && field.Type() == reflect.TypeOf(evaluation.EvaluationResult{}) {
					return field.Interface().(evaluation.EvaluationResult)
				}
			}
			return evaluation.EvaluationResult{}
		}
	}

	// Get the comparison function for the sort criteria
	less := getSortLessFunc(getResult, strings.ToLower(sortBy))
	sort.Slice(results, func(i, j int) bool { return less(results[i], results[j]) })
}

// Comparator function types for evaluation results
type evaluationComparator func(a, b evaluation.EvaluationResult) bool

// Built-in comparators for common sort criteria
var (
	compareByAttack       evaluationComparator = func(a, b evaluation.EvaluationResult) bool { return a.Attack.Score > b.Attack.Score }
	compareByDefense      evaluationComparator = func(a, b evaluation.EvaluationResult) bool { return a.Defense.Score > b.Defense.Score }
	compareBySynergy      evaluationComparator = func(a, b evaluation.EvaluationResult) bool { return a.Synergy.Score > b.Synergy.Score }
	compareByVersatility  evaluationComparator = func(a, b evaluation.EvaluationResult) bool { return a.Versatility.Score > b.Versatility.Score }
	compareByF2PFriendly  evaluationComparator = func(a, b evaluation.EvaluationResult) bool { return a.F2PFriendly.Score > b.F2PFriendly.Score }
	compareByPlayability  evaluationComparator = func(a, b evaluation.EvaluationResult) bool { return a.Playability.Score > b.Playability.Score }
	compareByElixir       evaluationComparator = func(a, b evaluation.EvaluationResult) bool { return a.AvgElixir < b.AvgElixir }
	compareByOverallScore evaluationComparator = func(a, b evaluation.EvaluationResult) bool { return a.OverallScore > b.OverallScore }
)

// getSortLessFunc returns a comparison function for the given sort criteria.
func getSortLessFunc[T any](getResult func(T) evaluation.EvaluationResult, sortBy string) func(T, T) bool {
	comparator := getComparatorForCriteria(sortBy)
	return func(a, b T) bool {
		return comparator(getResult(a), getResult(b))
	}
}

// getComparatorForCriteria returns the appropriate comparator function for the sort criteria
func getComparatorForCriteria(sortBy string) evaluationComparator {
	switch sortBy {
	case "attack":
		return compareByAttack
	case "defense":
		return compareByDefense
	case "synergy":
		return compareBySynergy
	case "versatility":
		return compareByVersatility
	case "f2p", "f2p-friendly":
		return compareByF2PFriendly
	case "playability":
		return compareByPlayability
	case "elixir":
		return compareByElixir
	default: // "overall"
		return compareByOverallScore
	}
}

// ============================================================================
// Reflection Helper Functions - Type Field Extraction
// ============================================================================

// extractName extracts the Name field from a generic struct type using reflection
func extractName[T any](r T) string {
	v := reflect.ValueOf(r)
	if v.Kind() == reflect.Struct {
		if field := v.FieldByName("Name"); field.IsValid() && field.Kind() == reflect.String {
			return field.String()
		}
	}
	return ""
}

// extractStrategy extracts the Strategy field from a generic struct type using reflection
func extractStrategy[T any](r T) string {
	v := reflect.ValueOf(r)
	if v.Kind() == reflect.Struct {
		if field := v.FieldByName("Strategy"); field.IsValid() && field.Kind() == reflect.String {
			return field.String()
		}
	}
	return ""
}

// extractDeck extracts the Deck field from a generic struct type using reflection
func extractDeck[T any](r T) []string {
	v := reflect.ValueOf(r)
	if v.Kind() == reflect.Struct {
		if field := v.FieldByName("Deck"); field.IsValid() && field.Kind() == reflect.Slice {
			if deck, ok := field.Interface().([]string); ok {
				return deck
			}
		}
	}
	return nil
}

// extractResult extracts the Result field from a generic struct type using reflection
func extractResult[T any](r T) evaluation.EvaluationResult {
	v := reflect.ValueOf(r)
	if v.Kind() == reflect.Struct {
		if field := v.FieldByName("Result"); field.IsValid() {
			if result, ok := field.Interface().(evaluation.EvaluationResult); ok {
				return result
			}
		}
	}
	return evaluation.EvaluationResult{}
}

// ============================================================================
// Formatting Helper Functions - Text Utilities
// ============================================================================

// truncateWithEllipsis truncates a string to maxLen and adds "..." if truncated
func truncateWithEllipsis(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

// formatScoreWithRating formats a score and rating in a consistent format
//
//nolint:unused // Shared formatter candidate retained for follow-up display refactor.
func formatScoreWithRating(score float64, rating string) string {
	return fmt.Sprintf("%.2f (%s)", score, rating)
}

// ============================================================================
// Batch Formatting Functions
// ============================================================================

// formatEvaluationBatchSummary formats batch evaluation results as a human-readable summary
func formatEvaluationBatchSummary[T any](results []T, totalDecks int, totalTime time.Duration, sortBy, playerName, playerTag string) string {
	var buf strings.Builder

	writeSummaryHeader(&buf, playerName, playerTag)
	writeSummaryStats(&buf, totalDecks, len(results), totalTime, sortBy)
	writeSummaryTable(&buf, results)

	return buf.String()
}

// writeSummaryHeader writes the header section for batch summary
func writeSummaryHeader(buf *strings.Builder, playerName, playerTag string) {
	buf.WriteString("╔═══════════════════════════════════════════════════════════════════════════════╗\n")
	buf.WriteString("║                        BATCH DECK EVALUATION RESULTS                          ║\n")
	buf.WriteString("╚═══════════════════════════════════════════════════════════════════════════════╝\n\n")

	if playerName != "" || playerTag != "" {
		buf.WriteString(fmt.Sprintf("Player: %s (%s)\n", playerName, playerTag))
	}
}

// writeSummaryStats writes the statistics section for batch summary
func writeSummaryStats(buf *strings.Builder, totalDecks, evaluatedCount int, totalTime time.Duration, sortBy string) {
	fmt.Fprintf(buf, "Total Decks: %d | Evaluated: %d | Sorted by: %s\n", totalDecks, evaluatedCount, sortBy)
	fmt.Fprintf(buf, "Total Time: %v | Avg: %v\n\n", totalTime, totalTime/time.Duration(max(evaluatedCount, 1)))
}

// writeSummaryTable writes the results table for batch summary
func writeSummaryTable[T any](buf *strings.Builder, results []T) {
	buf.WriteString("┌─────┬──────────────────────────────┬─────────┬────────┬────────┬────────┬──────────────┐\n")
	buf.WriteString("│ Rank│ Deck Name                    │ Overall │ Attack │ Defense│ Synergy│ Archetype    │\n")
	buf.WriteString("├─────┼──────────────────────────────┼─────────┼────────┼────────┼────────┼──────────────┤\n")

	for i, r := range results {
		name := extractName(r)
		if name == "" {
			continue
		}
		name = truncateWithEllipsis(name, 28)
		result := extractResult(r)
		archetype := truncateWithEllipsis(string(result.DetectedArchetype), 12)

		buf.WriteString(fmt.Sprintf("│ %3d │ %-28s │  %5.2f  │  %5.2f │  %5.2f │  %5.2f │ %-12s │\n",
			i+1, name,
			result.OverallScore,
			result.Attack.Score,
			result.Defense.Score,
			result.Synergy.Score,
			archetype))
	}

	buf.WriteString("└─────┴──────────────────────────────┴─────────┴────────┴────────┴────────┴──────────────┘\n")
}

// formatEvaluationBatchCSV formats batch evaluation results as CSV
func formatEvaluationBatchCSV[T any](results []T) string {
	var buf strings.Builder

	writeCSVHeader(&buf)
	writeCSVRows(&buf, results)

	return buf.String()
}

// writeCSVHeader writes the CSV header row
func writeCSVHeader(buf *strings.Builder) {
	buf.WriteString("Rank,Name,Strategy,Overall,Attack,Defense,Synergy,Versatility,F2P,Playability,Archetype,Avg_Elixir,Deck\n")
}

// writeCSVRows writes CSV data rows for evaluation results
func writeCSVRows[T any](buf *strings.Builder, results []T) {
	for i, r := range results {
		name := extractName(r)
		if name == "" {
			continue
		}
		strategy := extractStrategy(r)
		deck := extractDeck(r)
		result := extractResult(r)

		buf.WriteString(fmt.Sprintf("%d,%s,%s,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%s,%.2f,\"%s\"\n",
			i+1,
			name,
			strategy,
			result.OverallScore,
			result.Attack.Score,
			result.Defense.Score,
			result.Synergy.Score,
			result.Versatility.Score,
			result.F2PFriendly.Score,
			result.Playability.Score,
			result.DetectedArchetype,
			result.AvgElixir,
			strings.Join(deck, " - ")))
	}
}

// formatEvaluationBatchDetailed formats batch evaluation results with detailed analysis
func formatEvaluationBatchDetailed[T any](results []T, playerName, playerTag string) string {
	var buf strings.Builder

	writeDetailedHeader(&buf, playerName, playerTag)
	writeDetailedResults(&buf, results)

	return buf.String()
}

// writeDetailedHeader writes the header section for detailed batch results
func writeDetailedHeader(buf *strings.Builder, playerName, playerTag string) {
	buf.WriteString("╔═══════════════════════════════════════════════════════════════════════════════╗\n")
	buf.WriteString("║                    DETAILED BATCH EVALUATION RESULTS                          ║\n")
	buf.WriteString("╚═══════════════════════════════════════════════════════════════════════════════╝\n\n")

	if playerName != "" || playerTag != "" {
		buf.WriteString(fmt.Sprintf("Player: %s (%s)\n\n", playerName, playerTag))
	}
}

// writeDetailedResults writes detailed evaluation for each deck
func writeDetailedResults[T any](buf *strings.Builder, results []T) {
	for i, r := range results {
		name := extractName(r)
		if name == "" {
			continue
		}
		strategy := extractStrategy(r)
		deck := extractDeck(r)
		result := extractResult(r)

		writeDeckHeader(buf, i+1, name)
		writeDeckInfo(buf, strategy, deck, result)
		writeDeckScores(buf, result)
		writeDeckAssessments(buf, result)

		buf.WriteString("\n" + strings.Repeat("─", 80) + "\n\n")
	}
}

// writeDeckHeader writes the deck number and name header
func writeDeckHeader(buf *strings.Builder, deckNum int, name string) {
	fmt.Fprintf(buf, "═══════════════════ DECK #%d: %s ═══════════════════\n\n", deckNum, name)
}

// writeDeckInfo writes basic deck information (strategy, cards, elixir, archetype)
func writeDeckInfo(buf *strings.Builder, strategy string, deck []string, result evaluation.EvaluationResult) {
	if strategy != "" && strategy != "unknown" {
		fmt.Fprintf(buf, "Strategy: %s\n", strategy)
	}
	fmt.Fprintf(buf, "Deck: %s\n", strings.Join(deck, " - "))
	fmt.Fprintf(buf, "Avg Elixir: %.2f\n", result.AvgElixir)
	fmt.Fprintf(buf, "Archetype: %s (%.1f%% confidence)\n\n", result.DetectedArchetype, result.ArchetypeConfidence*100)
}

// writeDeckScores writes all category scores for a deck
func writeDeckScores(buf *strings.Builder, result evaluation.EvaluationResult) {
	buf.WriteString("SCORES:\n")
	fmt.Fprintf(buf, "  Overall:     %.2f (%s)\n", result.OverallScore, result.OverallRating)
	fmt.Fprintf(buf, "  Attack:      %.2f (%s)\n", result.Attack.Score, result.Attack.Rating)
	fmt.Fprintf(buf, "  Defense:     %.2f (%s)\n", result.Defense.Score, result.Defense.Rating)
	fmt.Fprintf(buf, "  Synergy:     %.2f (%s)\n", result.Synergy.Score, result.Synergy.Rating)
	fmt.Fprintf(buf, "  Versatility: %.2f (%s)\n", result.Versatility.Score, result.Versatility.Rating)
	fmt.Fprintf(buf, "  F2P:         %.2f (%s)\n", result.F2PFriendly.Score, result.F2PFriendly.Rating)
	fmt.Fprintf(buf, "  Playability: %.2f (%s)\n\n", result.Playability.Score, result.Playability.Rating)
}

// writeDeckAssessments writes key assessments for attack, defense, and synergy
func writeDeckAssessments(buf *strings.Builder, result evaluation.EvaluationResult) {
	if result.Attack.Assessment != "" {
		fmt.Fprintf(buf, "Attack: %s\n", result.Attack.Assessment)
	}
	if result.Defense.Assessment != "" {
		fmt.Fprintf(buf, "Defense: %s\n", result.Defense.Assessment)
	}
	if result.Synergy.Assessment != "" {
		fmt.Fprintf(buf, "Synergy: %s\n", result.Synergy.Assessment)
	}
}
