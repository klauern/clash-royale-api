package evaluation

import (
	"strings"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// TestFormatHuman tests the human-readable formatting output
func TestFormatHuman(t *testing.T) {
	result := getTestEvaluationResult()

	output := FormatHuman(result)

	// Verify major sections are present
	expectedSections := []string{
		"DECK EVALUATION REPORT",
		"Category Scores:",
		"DETAILED ANALYSIS",
		"cr-api deck evaluation engine",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("FormatHuman() missing section: %s", section)
		}
	}

	// Verify cards are included
	for _, card := range result.Deck {
		if !strings.Contains(output, card) {
			t.Errorf("FormatHuman() missing card: %s", card)
		}
	}

	// Verify overall score is displayed
	if !strings.Contains(output, "OVERALL SCORE:") {
		t.Error("FormatHuman() missing overall score")
	}
}

// TestFormatHeader tests the header formatting
func TestFormatHeader(t *testing.T) {
	result := getTestEvaluationResult()

	header := formatHeader(result)

	// Verify key elements
	expectedElements := []string{
		"DECK EVALUATION REPORT",
		"Deck Cards:",
		"Average Elixir:",
		"Archetype:",
		"OVERALL SCORE:",
	}

	for _, element := range expectedElements {
		if !strings.Contains(header, element) {
			t.Errorf("formatHeader() missing element: %s", element)
		}
	}

	// Verify archetype is shown
	if !strings.Contains(header, string(result.DetectedArchetype)) && !strings.Contains(header, "Cycle") {
		t.Error("formatHeader() missing archetype")
	}

	// Verify overall rating is shown
	if !strings.Contains(header, result.OverallRating.String()) {
		t.Error("formatHeader() missing overall rating")
	}
}

// TestFormatScoringGrid tests the scoring grid formatting
func TestFormatScoringGrid(t *testing.T) {
	result := getTestEvaluationResult()

	grid := formatScoringGrid(result)

	// Verify all categories are present
	expectedCategories := []string{
		"Attack",
		"Defense",
		"Synergy",
		"Versatility",
		"F2P Friendly",
	}

	for _, category := range expectedCategories {
		if !strings.Contains(grid, category) {
			t.Errorf("formatScoringGrid() missing category: %s", category)
		}
	}

	// Verify stars are shown
	if !strings.Contains(grid, "★") && !strings.Contains(grid, "☆") {
		t.Error("formatScoringGrid() missing star ratings")
	}

	// Verify scores are shown
	if !strings.Contains(grid, "Score:") {
		t.Error("formatScoringGrid() missing scores")
	}
}

// TestFormatStars tests star formatting
func TestFormatStars(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{"No stars", 0, "☆☆☆"},
		{"One star", 1, "★☆☆"},
		{"Two stars", 2, "★★☆"},
		{"Three stars", 3, "★★★"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStars(tt.count)
			if result != tt.expected {
				t.Errorf("formatStars(%d) = %q, want %q", tt.count, result, tt.expected)
			}
		})
	}
}

// TestFormatDetailedAnalysisHuman tests detailed analysis formatting
func TestFormatDetailedAnalysisHuman(t *testing.T) {
	result := getTestEvaluationResult()

	analysis := formatDetailedAnalysis(result)

	// Verify analysis header
	if !strings.Contains(analysis, "DETAILED ANALYSIS") {
		t.Error("formatDetailedAnalysis() missing header")
	}

	// Verify section titles are included
	expectedSections := []string{
		result.DefenseAnalysis.Title,
		result.AttackAnalysis.Title,
		result.BaitAnalysis.Title,
		result.CycleAnalysis.Title,
		result.LadderAnalysis.Title,
	}

	for _, section := range expectedSections {
		if !strings.Contains(analysis, section) {
			t.Errorf("formatDetailedAnalysis() missing section: %s", section)
		}
	}
}

// TestFormatAnalysisSection tests individual analysis section formatting
func TestFormatAnalysisSection(t *testing.T) {
	section := AnalysisSection{
		Title:   "Test Section",
		Score:   7.5,
		Rating:  "Great",
		Summary: "This is a test summary",
		Details: []string{"Detail 1", "Detail 2"},
	}

	output := formatAnalysisSection(section)

	// Verify section components
	if !strings.Contains(output, section.Title) {
		t.Error("formatAnalysisSection() missing title")
	}

	if !strings.Contains(output, section.Summary) {
		t.Error("formatAnalysisSection() missing summary")
	}

	if !strings.Contains(output, "Detail 1") {
		t.Error("formatAnalysisSection() missing detail 1")
	}

	if !strings.Contains(output, "Detail 2") {
		t.Error("formatAnalysisSection() missing detail 2")
	}

	if !strings.Contains(output, "Key Points:") {
		t.Error("formatAnalysisSection() missing Key Points header")
	}
}

// TestFormatSynergyMatrixHuman tests synergy matrix formatting
func TestFormatSynergyMatrixHuman(t *testing.T) {
	result := getTestEvaluationResult()

	matrix := formatSynergyMatrix(result)

	// Verify synergy matrix sections
	if !strings.Contains(matrix, "SYNERGY MATRIX") {
		t.Error("formatSynergyMatrix() missing header")
	}

	if !strings.Contains(matrix, "Total Synergy Score:") {
		t.Error("formatSynergyMatrix() missing total score")
	}

	if !strings.Contains(matrix, "Synergy Pairs Found:") {
		t.Error("formatSynergyMatrix() missing pair count")
	}

	// Check for top pairs
	if result.SynergyMatrix.PairCount > 0 {
		if !strings.Contains(matrix, "Top Synergy Pairs:") {
			t.Error("formatSynergyMatrix() missing top pairs section")
		}
	}
}

// TestFormatSynergyMatrixNoPairsHuman tests empty synergy matrix
func TestFormatSynergyMatrixNoPairsHuman(t *testing.T) {
	result := getTestEvaluationResult()
	result.SynergyMatrix.PairCount = 0
	result.SynergyMatrix.Pairs = []deck.SynergyPair{}

	matrix := formatSynergyMatrix(result)

	// Should return empty string when no pairs
	if matrix != "" {
		t.Error("formatSynergyMatrix() should return empty string when no pairs")
	}
}

// TestFormatSynergyStrength tests synergy strength formatting
func TestFormatSynergyStrength(t *testing.T) {
	tests := []struct {
		name     string
		strength float64
		expected string
	}{
		{"Excellent synergy", 0.9, "🔥🔥🔥 Excellent"},
		{"Strong synergy", 0.7, "🔥🔥 Strong"},
		{"Good synergy", 0.5, "🔥 Good"},
		{"Moderate synergy", 0.3, "• Moderate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSynergyStrength(tt.strength)
			if result != tt.expected {
				t.Errorf("formatSynergyStrength(%.1f) = %q, want %q", tt.strength, result, tt.expected)
			}
		})
	}
}

// TestFormatCounterAnalysisHuman tests counter analysis formatting
func TestFormatCounterAnalysisHuman(t *testing.T) {
	result := getTestEvaluationResult()

	counter := formatCounterAnalysis(result)

	// Verify counter analysis sections
	if !strings.Contains(counter, "COUNTER ANALYSIS") {
		t.Error("formatCounterAnalysis() missing header")
	}

	if !strings.Contains(counter, "Strengths:") {
		t.Error("formatCounterAnalysis() missing strengths section")
	}

	if !strings.Contains(counter, "Weaknesses:") {
		t.Error("formatCounterAnalysis() missing weaknesses section")
	}
}

// TestDeriveStrengths tests strength derivation
func TestDeriveStrengths(t *testing.T) {
	tests := []struct {
		name     string
		result   *EvaluationResult
		minCount int
	}{
		{
			name: "Multiple strengths",
			result: &EvaluationResult{
				Attack:      CategoryScore{Score: 8.0},
				Defense:     CategoryScore{Score: 8.0},
				Synergy:     CategoryScore{Score: 7.5},
				Versatility: CategoryScore{Score: 6.0},
				F2PFriendly: CategoryScore{Score: 7.0},
			},
			minCount: 3,
		},
		{
			name: "No strengths",
			result: &EvaluationResult{
				Attack:      CategoryScore{Score: 5.0},
				Defense:     CategoryScore{Score: 5.0},
				Synergy:     CategoryScore{Score: 5.0},
				Versatility: CategoryScore{Score: 5.0},
				F2PFriendly: CategoryScore{Score: 5.0},
			},
			minCount: 1, // Should have "balanced deck" message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strengths := deriveStrengths(tt.result)
			if len(strengths) < tt.minCount {
				t.Errorf("deriveStrengths() = %d strengths, want at least %d", len(strengths), tt.minCount)
			}
		})
	}
}

// TestDeriveWeaknesses tests weakness derivation
func TestDeriveWeaknesses(t *testing.T) {
	tests := []struct {
		name     string
		result   *EvaluationResult
		minCount int
	}{
		{
			name: "Multiple weaknesses",
			result: &EvaluationResult{
				Attack:      CategoryScore{Score: 3.0},
				Defense:     CategoryScore{Score: 4.0},
				Synergy:     CategoryScore{Score: 4.5},
				Versatility: CategoryScore{Score: 6.0},
				F2PFriendly: CategoryScore{Score: 7.0},
			},
			minCount: 3,
		},
		{
			name: "No weaknesses",
			result: &EvaluationResult{
				Attack:      CategoryScore{Score: 7.0},
				Defense:     CategoryScore{Score: 7.0},
				Synergy:     CategoryScore{Score: 7.0},
				Versatility: CategoryScore{Score: 7.0},
				F2PFriendly: CategoryScore{Score: 7.0},
			},
			minCount: 1, // Should have "no critical weaknesses" message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weaknesses := deriveWeaknesses(tt.result)
			if len(weaknesses) < tt.minCount {
				t.Errorf("deriveWeaknesses() = %d weaknesses, want at least %d", len(weaknesses), tt.minCount)
			}
		})
	}
}

// TestFormatRecommendationsHuman tests recommendations formatting
func TestFormatRecommendationsHuman(t *testing.T) {
	// Test with a weak deck that needs recommendations
	result := getTestEvaluationResult()
	result.Attack.Score = 3.0
	result.Defense.Score = 4.0

	recs := formatRecommendations(result)

	// Verify recommendations sections
	if !strings.Contains(recs, "RECOMMENDATIONS") {
		t.Error("formatRecommendations() missing header")
	}

	// Should have recommendations for low scores
	if !strings.Contains(recs, "1.") {
		t.Error("formatRecommendations() missing recommendations")
	}
}

// TestFormatRecommendationsNoneHuman tests no recommendations case
func TestFormatRecommendationsNoneHuman(t *testing.T) {
	// Test with a strong deck that needs no recommendations
	result := getTestEvaluationResult()
	result.Attack.Score = 8.0
	result.Defense.Score = 8.0
	result.Synergy.Score = 8.0
	result.Versatility.Score = 8.0
	result.F2PFriendly.Score = 8.0

	recs := formatRecommendations(result)

	// Should show well-balanced message
	if !strings.Contains(recs, "well-balanced") {
		t.Error("formatRecommendations() should show well-balanced message for strong deck")
	}
}

// TestGenerateRecommendationsFormatter tests recommendation generation for formatter
func TestGenerateRecommendationsFormatter(t *testing.T) {
	tests := []struct {
		name         string
		result       *EvaluationResult
		expectedRecs int
	}{
		{
			name: "Multiple weak categories",
			result: &EvaluationResult{
				Attack:      CategoryScore{Score: 3.0},
				Defense:     CategoryScore{Score: 4.0},
				Synergy:     CategoryScore{Score: 4.5},
				Versatility: CategoryScore{Score: 4.8},
				F2PFriendly: CategoryScore{Score: 4.9},
			},
			expectedRecs: 5,
		},
		{
			name: "Strong deck",
			result: &EvaluationResult{
				Attack:      CategoryScore{Score: 8.0},
				Defense:     CategoryScore{Score: 8.0},
				Synergy:     CategoryScore{Score: 7.0},
				Versatility: CategoryScore{Score: 7.0},
				F2PFriendly: CategoryScore{Score: 6.0},
			},
			expectedRecs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recs := generateRecommendations(tt.result)
			if len(recs) != tt.expectedRecs {
				t.Errorf("generateRecommendations() = %d recommendations, want %d", len(recs), tt.expectedRecs)
			}
		})
	}
}

// TestFormatCopyDeckLink tests deck link formatting
func TestFormatCopyDeckLink(t *testing.T) {
	result := getTestEvaluationResult()

	link := formatCopyDeckLink(result)

	// Verify link sections
	if !strings.Contains(link, "SHARE DECK") {
		t.Error("formatCopyDeckLink() missing header")
	}

	if !strings.Contains(link, "RoyaleAPI Link:") {
		t.Error("formatCopyDeckLink() missing RoyaleAPI link")
	}

	if !strings.Contains(link, "DeckShop.pro Link:") {
		t.Error("formatCopyDeckLink() missing DeckShop.pro link")
	}

	// Verify URLs are properly formatted
	if !strings.Contains(link, "https://royaleapi.com/decks/stats/") {
		t.Error("formatCopyDeckLink() missing RoyaleAPI URL")
	}

	if !strings.Contains(link, "https://www.deckshop.pro/check/?deck=") {
		t.Error("formatCopyDeckLink() missing DeckShop.pro URL")
	}
}

// TestFormatAlternativeSuggestions tests alternative suggestions formatting
func TestFormatAlternativeSuggestions(t *testing.T) {
	result := getTestEvaluationResult()
	result.AlternativeSuggestions = &AlternativeSuggestions{
		OriginalScore: 7.0,
		Suggestions: []AlternativeDeck{
			{
				OriginalCard:    "Ice Spirit",
				ReplacementCard: "Electro Spirit",
				Impact:          "Better stun effect",
				Rationale:       "Provides additional stun utility",
				OriginalScore:   7.0,
				NewScore:        7.5,
				ScoreDelta:      0.5,
			},
		},
	}

	alts := formatAlternativeSuggestions(result)

	// Verify alternative suggestions sections
	if !strings.Contains(alts, "ALTERNATIVE DECK SUGGESTIONS") {
		t.Error("formatAlternativeSuggestions() missing header")
	}

	if !strings.Contains(alts, "Current Deck Score:") {
		t.Error("formatAlternativeSuggestions() missing current score")
	}

	if !strings.Contains(alts, "Alternative #1:") {
		t.Error("formatAlternativeSuggestions() missing alternative")
	}
}

// TestFormatAlternativeSuggestionsNone tests no alternatives case
func TestFormatAlternativeSuggestionsNone(t *testing.T) {
	result := getTestEvaluationResult()
	result.AlternativeSuggestions = nil

	alts := formatAlternativeSuggestions(result)

	// Should return empty string when no alternatives
	if alts != "" {
		t.Error("formatAlternativeSuggestions() should return empty string when no alternatives")
	}
}

// TestFormatMissingCards tests missing cards formatting
func TestFormatMissingCards(t *testing.T) {
	result := getTestEvaluationResult()
	result.MissingCardsAnalysis = &MissingCardsAnalysis{
		Deck:           result.Deck,
		IsPlayable:     false,
		AvailableCount: 6,
		MissingCards: []MissingCard{
			{
				Name:             "Musketeer",
				Rarity:           "Rare",
				UnlockArena:      3,
				UnlockArenaName:  "Arena 3",
				IsLocked:         false,
				AlternativeCards: []string{"Wizard", "Electro Wizard"},
			},
		},
	}

	missing := formatMissingCards(result)

	// Verify missing cards sections
	if !strings.Contains(missing, "MISSING CARDS ANALYSIS") {
		t.Error("formatMissingCards() missing header")
	}

	if !strings.Contains(missing, "Deck Status:") {
		t.Error("formatMissingCards() missing deck status")
	}

	if !strings.Contains(missing, "Musketeer") {
		t.Error("formatMissingCards() missing card name")
	}

	if !strings.Contains(missing, "Alternatives:") {
		t.Error("formatMissingCards() missing alternatives")
	}
}

// TestFormatMissingCardsPlayable tests playable deck case
func TestFormatMissingCardsPlayable(t *testing.T) {
	result := getTestEvaluationResult()
	result.MissingCardsAnalysis = &MissingCardsAnalysis{
		Deck:           result.Deck,
		IsPlayable:     true,
		AvailableCount: 8,
		MissingCards:   []MissingCard{},
	}

	missing := formatMissingCards(result)

	// Should show success message
	if !strings.Contains(missing, "All cards in this deck are available") {
		t.Error("formatMissingCards() should show success message for playable deck")
	}
}

// TestFormatFooter tests footer formatting
func TestFormatFooter(t *testing.T) {
	result := getTestEvaluationResult()

	footer := formatFooter(result)

	// Verify footer elements
	if !strings.Contains(footer, "cr-api deck evaluation engine") {
		t.Error("formatFooter() missing engine name")
	}

	if !strings.Contains(footer, "v1.0.0") {
		t.Error("formatFooter() missing version")
	}

	if !strings.Contains(footer, "Clash Royale API") {
		t.Error("formatFooter() missing API reference")
	}
}
