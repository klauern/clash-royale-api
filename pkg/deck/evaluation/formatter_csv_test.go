package evaluation

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestFormatCSV(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()
	deckCards := []deck.CardCandidate{
		makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
		makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
		makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
		makeCard("Zap", deck.RoleSpellSmall, 11, 11, "Common", 2),
		makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
		makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
		makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
		makeCard("Ice Golem", deck.RoleCycle, 11, 11, "Rare", 2),
	}

	result := Evaluate(deckCards, synergyDB)

	csv := FormatCSV(&result)

	if csv == "" {
		t.Errorf("FormatCSV() returned empty string")
	}

	// Check for expected sections
	expectedSections := []string{
		"# DECK EVALUATION SUMMARY",
		"# CATEGORY SCORES",
		"# DETAILED ANALYSIS",
		"# STRENGTHS",
		"# WEAKNESSES",
	}

	for _, section := range expectedSections {
		if !contains(csv, section) {
			t.Errorf("FormatCSV() missing section %q", section)
		}
	}

	// Check for expected fields
	expectedFields := []string{
		"Average Elixir",
		"Overall Score",
		"Category,Score,Rating,Stars,Assessment",
	}

	for _, field := range expectedFields {
		if !contains(csv, field) {
			t.Errorf("FormatCSV() missing field %q", field)
		}
	}
}

func TestFormatCategoryScoreCSV(t *testing.T) {
	score := CategoryScore{
		Score:      7.5,
		Rating:     "Good",
		Stars:      3,
		Assessment: "Solid performance",
	}

	result := formatCategoryScoreCSV("Attack", score)

	if result == "" {
		t.Errorf("formatCategoryScoreCSV() returned empty string")
	}

	// Check format
	expectedFields := []string{"Attack", "7.50", "Good", "3", "Solid performance"}
	for _, field := range expectedFields {
		if !contains(result, field) {
			t.Errorf("formatCategoryScoreCSV() missing expected field %q", field)
		}
	}
}

func TestFormatAnalysisSectionCSV(t *testing.T) {
	section := AnalysisSection{
		Title:   "Defense Analysis",
		Score:   8.0,
		Rating:  "Strong",
		Summary: "Good defensive capabilities",
		Details: []string{"Strong anti-air", "Good buildings"},
	}

	result := formatAnalysisSectionCSV(section)

	if result == "" {
		t.Errorf("formatAnalysisSectionCSV() returned empty string")
	}

	// Check format
	expectedFields := []string{"Defense Analysis", "8.00", "Strong", "Good defensive capabilities"}
	for _, field := range expectedFields {
		if !contains(result, field) {
			t.Errorf("formatAnalysisSectionCSV() missing expected field %q", field)
		}
	}
}

func TestGetSynergyStrengthLabel(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{0.9, "Excellent"},
		{0.8, "Excellent"},
		{0.7, "Strong"},
		{0.6, "Strong"},
		{0.5, "Good"},
		{0.4, "Good"},
		{0.3, "Moderate"},
		{0.0, "Moderate"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getSynergyStrengthLabel(tt.score)
			if result != tt.expected {
				t.Errorf("getSynergyStrengthLabel(%.1f) = %q, want %q",
					tt.score, result, tt.expected)
			}
		})
	}
}

func TestEscapeCSV(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No special chars",
			input:    "Simple text",
			expected: "Simple text",
		},
		{
			name:     "Contains comma",
			input:    "Hello, world",
			expected: "\"Hello, world\"",
		},
		{
			name:     "Contains quote",
			input:    `Say "hello"`,
			expected: `"Say ""hello"""`,
		},
		{
			name:     "Contains newline",
			input:    "Line 1\nLine 2",
			expected: "\"Line 1\nLine 2\"",
		},
		{
			name:     "Contains multiple special chars",
			input:    `Value "with", comma`,
			expected: `"Value ""with"", comma"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeCSV(tt.input)
			if result != tt.expected {
				t.Errorf("escapeCSV(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
