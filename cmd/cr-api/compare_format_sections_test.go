package main

import (
	"strings"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

func sampleCompareResults() ([]string, []evaluation.EvaluationResult) {
	names := []string{"Deck One", "Deck Two Long Name"}
	deckOneScores := []evaluation.CategoryScore{
		{Score: 8.1, Rating: "Strong", Assessment: "Great pressure", Stars: 2},
		{Score: 7.2, Rating: "Good", Assessment: "Solid defense", Stars: 2},
		{Score: 7.4, Rating: "Good", Assessment: "Reliable pairs", Stars: 2},
		{Score: 7.0, Rating: "Good", Assessment: "Flexible", Stars: 2},
		{Score: 8.4, Rating: "Great", Assessment: "Cheap upgrades", Stars: 3},
		{Score: 8.6, Rating: "Great", Assessment: "Fully leveled", Stars: 3},
	}
	deckTwoScores := []evaluation.CategoryScore{
		{Score: 7.0, Rating: "Good", Assessment: "Heavy pushes", Stars: 2},
		{Score: 6.5, Rating: "Okay", Assessment: "Can be outcycled", Stars: 1},
		{Score: 7.2, Rating: "Good", Assessment: "Good combos", Stars: 2},
		{Score: 6.0, Rating: "Okay", Assessment: "Matchup dependent", Stars: 1},
		{Score: 5.8, Rating: "Mediocre", Assessment: "Expensive core", Stars: 1},
		{Score: 4.8, Rating: "Weak", Assessment: "Underleveled cards", Stars: 0},
	}
	results := []evaluation.EvaluationResult{
		makeResult(
			[]string{"Knight", "Musketeer", "Hog Rider", "Fireball", "Cannon", "Ice Spirit", "Skeletons", "The Log"},
			7.8, "Strong", 3.1, "Cycle",
			deckOneScores[0], deckOneScores[1], deckOneScores[2],
			deckOneScores[3], deckOneScores[4], deckOneScores[5],
		),
		makeResult(
			[]string{"Golem", "Night Witch", "Baby Dragon", "Tornado", "Lightning", "Lumberjack", "Barbarian Barrel", "Electro Dragon"},
			6.9, "Good", 4.3, "Beatdown",
			deckTwoScores[0], deckTwoScores[1], deckTwoScores[2],
			deckTwoScores[3], deckTwoScores[4], deckTwoScores[5],
		),
	}

	return names, results
}

func makeResult(
	deck []string,
	overallScore float64,
	overallRating evaluation.Rating,
	avgElixir float64,
	archetype evaluation.Archetype,
	attack, defense, synergy, versatility, f2p, playability evaluation.CategoryScore,
) evaluation.EvaluationResult {
	return evaluation.EvaluationResult{
		Deck:              deck,
		OverallScore:      overallScore,
		OverallRating:     overallRating,
		AvgElixir:         avgElixir,
		DetectedArchetype: archetype,
		Attack:            attack,
		Defense:           defense,
		Synergy:           synergy,
		Versatility:       versatility,
		F2PFriendly:       f2p,
		Playability:       playability,
	}
}

func TestFormatTableCategoryScoresSection(t *testing.T) {
	names, results := sampleCompareResults()
	var sb strings.Builder
	vm := buildComparisonViewModel(names, results)

	formatTableCategoryScoresSection(&sb, vm)
	output := sb.String()

	for _, want := range []string{"CATEGORY SCORES", "Deck One", "Attack", "F2P Friendly", "Playability"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got:\n%s", want, output)
		}
	}
}

func TestFormatMarkdownBestInCategorySection(t *testing.T) {
	names, results := sampleCompareResults()
	var sb strings.Builder
	vm := buildComparisonViewModel(names, results)

	formatMarkdownBestInCategorySection(&sb, vm)
	output := sb.String()

	if !strings.Contains(output, "**Attack**: Deck One (8.10)") {
		t.Fatalf("expected Deck One to win attack category, got:\n%s", output)
	}
	if !strings.Contains(output, "**F2P Friendly**: Deck One (8.40)") {
		t.Fatalf("expected Deck One to win f2p category, got:\n%s", output)
	}
	if !strings.Contains(output, "**Playability**: Deck One (8.60)") {
		t.Fatalf("expected Deck One to win playability category, got:\n%s", output)
	}
}

func TestFormatReportDetailedScoreComparison(t *testing.T) {
	names, results := sampleCompareResults()
	var sb strings.Builder
	vm := buildComparisonViewModel(names, results)

	formatReportDetailedScoreComparison(&sb, vm)
	output := sb.String()

	for _, want := range []string{"Detailed Score Comparison", "| **Overall** |", "| Avg Elixir |", "| Archetype |", "Deck Two Long Name"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output, got:\n%s", want, output)
		}
	}
}
