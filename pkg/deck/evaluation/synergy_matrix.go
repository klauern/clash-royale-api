package evaluation

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// GenerateSynergyMatrix creates a matrix visualization of all card pairings in a deck.
// It scores each pair using the synergy database and identifies top synergies.
func GenerateSynergyMatrix(deckCards []string, synergyDB *deck.SynergyDatabase) *SynergyMatrix {
	if synergyDB == nil || len(deckCards) == 0 {
		return nil
	}

	// Analyze the deck synergies using the database
	analysis := synergyDB.AnalyzeDeckSynergy(deckCards)
	if analysis == nil {
		return nil
	}

	// Calculate all pairs in the deck (C(n,2) where n=8, so 28 pairs)
	maxPairs := (len(deckCards) * (len(deckCards) - 1)) / 2

	// Calculate coverage: percentage of possible pairs that have synergies
	pairCount := 0
	for _, count := range analysis.CategoryScores {
		pairCount += count
	}
	coverage := 0.0
	if maxPairs > 0 {
		coverage = (float64(pairCount) / float64(maxPairs)) * 100.0
	}

	// Create the matrix structure
	matrix := &SynergyMatrix{
		Pairs:            analysis.TopSynergies,
		TotalScore:       0.0, // Will be set by caller from ScoreSynergy
		AverageSynergy:   analysis.AverageScore,
		PairCount:        pairCount,
		MaxPossiblePairs: maxPairs,
		SynergyCoverage:  coverage,
	}

	return matrix
}

// FormatSynergyMatrixText generates a human-readable text representation of the synergy matrix.
// It creates a grid showing synergy scores between all card pairs.
func FormatSynergyMatrixText(matrix *SynergyMatrix, deckCards []string, synergyDB *deck.SynergyDatabase) string {
	if matrix == nil || len(deckCards) == 0 {
		return "No synergy matrix available"
	}

	var sb strings.Builder

	// Header
	sb.WriteString("═══════════════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("  SYNERGY MATRIX - Card Pair Analysis\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════════════\n\n")

	// Summary stats
	sb.WriteString(fmt.Sprintf("Total Synergies: %d/%d pairs (%.1f%% coverage)\n",
		matrix.PairCount, matrix.MaxPossiblePairs, matrix.SynergyCoverage))
	sb.WriteString(fmt.Sprintf("Average Synergy: %.2f/1.00\n", matrix.AverageSynergy))
	sb.WriteString(fmt.Sprintf("Overall Score: %.1f/10.0\n\n", matrix.TotalScore))

	// Create synergy lookup map for fast access
	synergyMap := make(map[string]float64)
	for _, pair := range matrix.Pairs {
		key1 := pair.Card1 + "|" + pair.Card2
		key2 := pair.Card2 + "|" + pair.Card1
		synergyMap[key1] = pair.Score
		synergyMap[key2] = pair.Score
	}

	// Generate matrix grid
	sb.WriteString("Synergy Grid (scores 0.0-1.0):\n\n")

	// Use tabwriter for alignment
	tw := new(tabwriter.Writer)
	tw.Init(&sb, 0, 8, 2, ' ', 0)
	var writeErr error
	writef := func(format string, args ...any) {
		if writeErr != nil {
			return
		}
		if _, err := fmt.Fprintf(tw, format, args...); err != nil {
			writeErr = err
		}
	}

	// Header row with card names (abbreviated to 10 chars max)
	writef("Card\t")
	for _, card := range deckCards {
		abbrev := abbreviateCardName(card)
		writef("%s\t", abbrev)
	}
	writef("\n")

	// Separator
	writef("────\t")
	for range deckCards {
		writef("──────\t")
	}
	writef("\n")

	// Data rows
	for i, card1 := range deckCards {
		abbrev1 := abbreviateCardName(card1)
		writef("%s\t", abbrev1)

		for j, card2 := range deckCards {
			if i == j {
				// Same card - show dash
				writef("  -   \t")
			} else {
				// Look up synergy score
				key := card1 + "|" + card2
				score := synergyMap[key]
				if score > 0 {
					writef("%.2f\t", score)
				} else {
					writef("  .   \t")
				}
			}
		}
		writef("\n")
	}

	if err := tw.Flush(); err != nil && writeErr == nil {
		writeErr = err
	}
	if writeErr != nil {
		return fmt.Sprintf("Failed to format synergy matrix: %v", writeErr)
	}
	sb.WriteString("\n")

	// Top synergies section
	if len(matrix.Pairs) > 0 {
		sb.WriteString("Top Synergies:\n")
		sb.WriteString("──────────────\n")

		// Sort pairs by score descending
		sortedPairs := make([]deck.SynergyPair, len(matrix.Pairs))
		copy(sortedPairs, matrix.Pairs)
		sort.Slice(sortedPairs, func(i, j int) bool {
			return sortedPairs[i].Score > sortedPairs[j].Score
		})

		// Show top 5 synergies with narrative
		displayCount := 5
		if len(sortedPairs) < displayCount {
			displayCount = len(sortedPairs)
		}

		for i := 0; i < displayCount; i++ {
			pair := sortedPairs[i]
			rating := formatSynergyRating(pair.Score)
			sb.WriteString(fmt.Sprintf("%d. %s + %s (%.2f) - %s\n",
				i+1,
				pair.Card1,
				pair.Card2,
				pair.Score,
				rating))
			if pair.Description != "" {
				sb.WriteString(fmt.Sprintf("   → %s\n", pair.Description))
			}
		}
		sb.WriteString("\n")
	}

	// Category breakdown
	if synergyDB != nil {
		categoryCount := make(map[deck.SynergyCategory]int)
		for _, pair := range matrix.Pairs {
			categoryCount[pair.SynergyType]++
		}

		if len(categoryCount) > 0 {
			sb.WriteString("Synergy Categories:\n")
			sb.WriteString("───────────────────\n")

			// Sort categories by count
			type catCount struct {
				category deck.SynergyCategory
				count    int
			}
			var categories []catCount
			for cat, count := range categoryCount {
				categories = append(categories, catCount{cat, count})
			}
			sort.Slice(categories, func(i, j int) bool {
				return categories[i].count > categories[j].count
			})

			for _, cc := range categories {
				sb.WriteString(fmt.Sprintf("  • %s: %d pairs\n",
					deck.GetCategoryDescription(cc.category),
					cc.count))
			}
		}
	}

	return sb.String()
}

// FormatSynergyMatrixJSON returns a JSON-compatible map representation
func FormatSynergyMatrixJSON(matrix *SynergyMatrix) map[string]interface{} {
	if matrix == nil {
		return nil
	}

	// Convert pairs to simple format
	pairs := make([]map[string]interface{}, len(matrix.Pairs))
	for i, pair := range matrix.Pairs {
		pairs[i] = map[string]interface{}{
			"card1":       pair.Card1,
			"card2":       pair.Card2,
			"type":        string(pair.SynergyType),
			"score":       pair.Score,
			"description": pair.Description,
		}
	}

	return map[string]interface{}{
		"pairs":              pairs,
		"total_score":        matrix.TotalScore,
		"average_synergy":    matrix.AverageSynergy,
		"pair_count":         matrix.PairCount,
		"max_possible_pairs": matrix.MaxPossiblePairs,
		"synergy_coverage":   matrix.SynergyCoverage,
	}
}

// GenerateTopSynergyNarrative creates human-readable commentary for top synergies
func GenerateTopSynergyNarrative(matrix *SynergyMatrix) []string {
	if matrix == nil || len(matrix.Pairs) == 0 {
		return []string{"No synergies found in this deck"}
	}

	var narratives []string

	// Sort pairs by score
	sortedPairs := make([]deck.SynergyPair, len(matrix.Pairs))
	copy(sortedPairs, matrix.Pairs)
	sort.Slice(sortedPairs, func(i, j int) bool {
		return sortedPairs[i].Score > sortedPairs[j].Score
	})

	// Generate narrative for top 3
	topCount := 3
	if len(sortedPairs) < topCount {
		topCount = len(sortedPairs)
	}

	for i := 0; i < topCount; i++ {
		pair := sortedPairs[i]
		rating := formatSynergyRating(pair.Score)
		narrative := fmt.Sprintf("%s and %s have %s synergy (%.2f)",
			pair.Card1, pair.Card2, rating, pair.Score)
		if pair.Description != "" {
			narrative += fmt.Sprintf(": %s", pair.Description)
		}
		narratives = append(narratives, narrative)
	}

	// Overall assessment
	avgRating := formatSynergyRating(matrix.AverageSynergy)
	overallNarrative := fmt.Sprintf("Overall, this deck has %s synergy with %d/%d card pairs working together",
		avgRating, matrix.PairCount, matrix.MaxPossiblePairs)
	narratives = append(narratives, overallNarrative)

	return narratives
}

// abbreviateCardName shortens card names for matrix display
func abbreviateCardName(name string) string {
	maxLen := 10
	if len(name) <= maxLen {
		return name
	}

	// Try to abbreviate intelligently
	// Remove "The" prefix
	name = strings.TrimPrefix(name, "The ")

	// If still too long, truncate with ellipsis
	if len(name) > maxLen {
		return name[:maxLen-1] + "…"
	}

	return name
}

// formatSynergyRating converts numeric score to qualitative rating
func formatSynergyRating(score float64) string {
	switch {
	case score >= 0.9:
		return "exceptional"
	case score >= 0.8:
		return "excellent"
	case score >= 0.7:
		return "strong"
	case score >= 0.6:
		return "good"
	case score >= 0.5:
		return "moderate"
	case score >= 0.3:
		return "weak"
	default:
		return "minimal"
	}
}
