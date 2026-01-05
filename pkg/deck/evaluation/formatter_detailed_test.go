package evaluation

import (
	"strings"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// TestFormatDetailed tests the detailed formatting output
func TestFormatDetailed(t *testing.T) {
	result := getTestEvaluationResult()

	output := FormatDetailed(result)

	// Verify major sections are present
	expectedSections := []string{
		"DETAILED DECK EVALUATION REPORT",
		"DECK COMPOSITION",
		"CATEGORY SCORING BREAKDOWN",
		"COMPREHENSIVE DECK ANALYSIS",
		"DEBUG INFORMATION",
		"cr-api deck evaluation engine",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("FormatDetailed() missing section: %s", section)
		}
	}

	// Verify cards are included
	for _, card := range result.Deck {
		if !strings.Contains(output, card) {
			t.Errorf("FormatDetailed() missing card: %s", card)
		}
	}

	// Verify scores are included
	if !strings.Contains(output, "Attack") {
		t.Error("FormatDetailed() missing Attack category")
	}
	if !strings.Contains(output, "Defense") {
		t.Error("FormatDetailed() missing Defense category")
	}
}

// TestFormatDetailedHeader tests the header formatting
func TestFormatDetailedHeader(t *testing.T) {
	result := getTestEvaluationResult()

	header := formatDetailedHeader(result)

	// Verify key elements
	expectedElements := []string{
		"DECK COMPOSITION",
		"Card Count:",
		"Average Elixir Cost:",
		"Archetype Detection:",
		"Overall Evaluation:",
		"Overall Score:",
	}

	for _, element := range expectedElements {
		if !strings.Contains(header, element) {
			t.Errorf("formatDetailedHeader() missing element: %s", element)
		}
	}

	// Verify archetype is included
	if !strings.Contains(header, string(result.DetectedArchetype)) && !strings.Contains(header, "Cycle") {
		t.Error("formatDetailedHeader() missing archetype")
	}
}

// TestFormatDetailedCategoryScores tests category score formatting
func TestFormatDetailedCategoryScores(t *testing.T) {
	result := getTestEvaluationResult()

	scores := formatDetailedCategoryScores(result)

	// Verify all categories are present
	expectedCategories := []string{
		"ATTACK SCORE",
		"DEFENSE SCORE",
		"SYNERGY SCORE",
		"VERSATILITY SCORE",
		"F2P FRIENDLY SCORE",
	}

	for _, category := range expectedCategories {
		if !strings.Contains(scores, category) {
			t.Errorf("formatDetailedCategoryScores() missing category: %s", category)
		}
	}

	// Verify score details
	if !strings.Contains(scores, "Numeric Score:") {
		t.Error("formatDetailedCategoryScores() missing numeric scores")
	}
	if !strings.Contains(scores, "Star Rating:") {
		t.Error("formatDetailedCategoryScores() missing star ratings")
	}
	if !strings.Contains(scores, "Scoring Methodology:") {
		t.Error("formatDetailedCategoryScores() missing methodology")
	}
}

// TestFormatDetailedAnalysisSections tests analysis section formatting
func TestFormatDetailedAnalysisSections(t *testing.T) {
	result := getTestEvaluationResult()

	analysis := formatDetailedAnalysisSections(result)

	// Verify analysis sections are present
	if !strings.Contains(analysis, "COMPREHENSIVE DECK ANALYSIS") {
		t.Error("formatDetailedAnalysisSections() missing header")
	}

	// Verify section titles are included (uppercased in output)
	expectedSections := []string{
		strings.ToUpper(result.DefenseAnalysis.Title),
		strings.ToUpper(result.AttackAnalysis.Title),
		strings.ToUpper(result.BaitAnalysis.Title),
		strings.ToUpper(result.CycleAnalysis.Title),
		strings.ToUpper(result.LadderAnalysis.Title),
	}

	for _, section := range expectedSections {
		if !strings.Contains(analysis, section) {
			t.Errorf("formatDetailedAnalysisSections() missing section: %s", section)
		}
	}

	// Verify scores are shown
	if !strings.Contains(analysis, "Score:") {
		t.Error("formatDetailedAnalysisSections() missing scores")
	}
	if !strings.Contains(analysis, "Executive Summary:") {
		t.Error("formatDetailedAnalysisSections() missing summaries")
	}
}

// TestFormatDetailedSynergyMatrix tests synergy matrix formatting
func TestFormatDetailedSynergyMatrix(t *testing.T) {
	result := getTestEvaluationResult()

	matrix := formatDetailedSynergyMatrix(result)

	// Verify synergy matrix sections
	if !strings.Contains(matrix, "SYNERGY MATRIX ANALYSIS") {
		t.Error("formatDetailedSynergyMatrix() missing header")
	}

	if !strings.Contains(matrix, "Total Synergy Score:") {
		t.Error("formatDetailedSynergyMatrix() missing total score")
	}

	if !strings.Contains(matrix, "Synergy Pairs Found:") {
		t.Error("formatDetailedSynergyMatrix() missing pair count")
	}

	// Check for synergy pair details if pairs exist
	if result.SynergyMatrix.PairCount > 0 {
		if !strings.Contains(matrix, "Card-by-Card Synergy Breakdown:") {
			t.Error("formatDetailedSynergyMatrix() missing synergy pairs")
		}
	}
}

// TestFormatDetailedSynergyMatrixNoPairs tests empty synergy matrix
func TestFormatDetailedSynergyMatrixNoPairs(t *testing.T) {
	result := getTestEvaluationResult()
	result.SynergyMatrix.PairCount = 0
	result.SynergyMatrix.Pairs = []deck.SynergyPair{}

	matrix := formatDetailedSynergyMatrix(result)

	// Should return empty string when no pairs
	if matrix != "" {
		t.Error("formatDetailedSynergyMatrix() should return empty string when no pairs")
	}
}

// TestFormatDetailedCounterAnalysis tests counter analysis formatting
func TestFormatDetailedCounterAnalysis(t *testing.T) {
	result := getTestEvaluationResult()

	counter := formatDetailedCounterAnalysis(result)

	// Verify counter analysis sections
	if !strings.Contains(counter, "MATCHUP ANALYSIS: STRENGTHS & WEAKNESSES") {
		t.Error("formatDetailedCounterAnalysis() missing header")
	}

	if !strings.Contains(counter, "DECK STRENGTHS") {
		t.Error("formatDetailedCounterAnalysis() missing strengths section")
	}

	if !strings.Contains(counter, "DECK WEAKNESSES") {
		t.Error("formatDetailedCounterAnalysis() missing weaknesses section")
	}
}

// TestFormatDetailedRecommendations tests recommendations formatting
func TestFormatDetailedRecommendations(t *testing.T) {
	// Test with a weak deck that needs recommendations
	result := getTestEvaluationResult()
	result.Attack.Score = 3.0
	result.Defense.Score = 4.0

	recs := formatDetailedRecommendations(result)

	// Verify recommendations sections
	if !strings.Contains(recs, "IMPROVEMENT RECOMMENDATIONS") {
		t.Error("formatDetailedRecommendations() missing header")
	}

	if !strings.Contains(recs, "Priority-Ordered Action Items:") {
		t.Error("formatDetailedRecommendations() missing action items header")
	}

	// Should have recommendations for low scores
	if !strings.Contains(recs, "Recommendation #") {
		t.Error("formatDetailedRecommendations() missing recommendations")
	}
}

// TestFormatDetailedRecommendationsNone tests no recommendations case
func TestFormatDetailedRecommendationsNone(t *testing.T) {
	// Test with a strong deck that needs no recommendations
	result := getTestEvaluationResult()
	result.Attack.Score = 8.0
	result.Defense.Score = 8.0
	result.Synergy.Score = 8.0
	result.Versatility.Score = 8.0
	result.F2PFriendly.Score = 8.0

	recs := formatDetailedRecommendations(result)

	// Should show congratulations message
	if !strings.Contains(recs, "CONGRATULATIONS") {
		t.Error("formatDetailedRecommendations() should show congratulations for strong deck")
	}
}

// TestFormatDebugInformation tests debug info formatting
func TestFormatDebugInformation(t *testing.T) {
	result := getTestEvaluationResult()

	debug := formatDebugInformation(result)

	// Verify debug sections
	if !strings.Contains(debug, "DEBUG INFORMATION") {
		t.Error("formatDebugInformation() missing header")
	}

	if !strings.Contains(debug, "Evaluation Metadata:") {
		t.Error("formatDebugInformation() missing metadata")
	}

	if !strings.Contains(debug, "Score Calculation Details:") {
		t.Error("formatDebugInformation() missing calculation details")
	}

	if !strings.Contains(debug, "Rating Scale Reference:") {
		t.Error("formatDebugInformation() missing rating scale")
	}

	// Verify score breakdown is included
	expectedScores := []string{
		"Attack Score:",
		"Defense Score:",
		"Synergy Score:",
		"Versatility Score:",
		"F2P Friendly Score:",
	}

	for _, score := range expectedScores {
		if !strings.Contains(debug, score) {
			t.Errorf("formatDebugInformation() missing score: %s", score)
		}
	}
}

// getTestEvaluationResult creates a test evaluation result
func getTestEvaluationResult() *EvaluationResult {
	return &EvaluationResult{
		Deck: []string{
			"Hog Rider",
			"Musketeer",
			"Fireball",
			"The Log",
			"Ice Spirit",
			"Skeletons",
			"Cannon",
			"Ice Golem",
		},
		AvgElixir:           2.6,
		DetectedArchetype:   ArchetypeCycle,
		ArchetypeConfidence: 0.85,
		OverallScore:        7.5,
		OverallRating:       "Great",
		Attack:              CategoryScore{Score: 7.0, Stars: 2, Rating: "Good", Assessment: "Solid offensive potential"},
		Defense:             CategoryScore{Score: 8.0, Stars: 3, Rating: "Great", Assessment: "Strong defensive capabilities"},
		Synergy:             CategoryScore{Score: 7.5, Stars: 2, Rating: "Good", Assessment: "Good card synergies"},
		Versatility:         CategoryScore{Score: 6.5, Stars: 2, Rating: "Good", Assessment: "Moderate versatility"},
		F2PFriendly:         CategoryScore{Score: 8.5, Stars: 3, Rating: "Great", Assessment: "Very F2P friendly"},
		DefenseAnalysis:     AnalysisSection{Title: "Defense Analysis", Score: 8.0, Rating: "Great", Summary: "Strong defense", Details: []string{"Good anti-air", "Solid building"}},
		AttackAnalysis:      AnalysisSection{Title: "Attack Analysis", Score: 7.0, Rating: "Good", Summary: "Decent attack", Details: []string{"Single win condition"}},
		BaitAnalysis:        AnalysisSection{Title: "Bait Analysis", Score: 6.0, Rating: "Good", Summary: "Some bait potential", Details: []string{}},
		CycleAnalysis:       AnalysisSection{Title: "Cycle Analysis", Score: 9.0, Rating: "Amazing", Summary: "Excellent cycle", Details: []string{"Very fast cycle", "Low elixir cost"}},
		LadderAnalysis:      AnalysisSection{Title: "Ladder Analysis", Score: 8.0, Rating: "Great", Summary: "Great for ladder", Details: []string{"F2P friendly", "Effective at all levels"}},
		SynergyMatrix: SynergyMatrix{
			TotalScore:       7.5,
			AverageSynergy:   0.75,
			PairCount:        5,
			MaxPossiblePairs: 28,
			SynergyCoverage:  62.5,
			Pairs: []deck.SynergyPair{
				{Card1: "Hog Rider", Card2: "Ice Golem", Score: 0.8, Description: "Ice Golem tanks for Hog Rider"},
				{Card1: "Musketeer", Card2: "Ice Golem", Score: 0.7, Description: "Ice Golem protects Musketeer"},
			},
		},
	}
}
