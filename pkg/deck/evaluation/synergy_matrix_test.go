package evaluation

import (
	"strings"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestGenerateSynergyMatrix(t *testing.T) {
	tests := []struct {
		name          string
		deckCards     []string
		synergyDB     *deck.SynergyDatabase
		wantNil       bool
		wantPairCount int
	}{
		{
			name:      "nil database",
			deckCards: []string{"Giant", "Witch", "Wizard", "Fireball", "Arrows", "Archers", "Tombstone", "Knight"},
			synergyDB: nil,
			wantNil:   true,
		},
		{
			name:      "empty deck",
			deckCards: []string{},
			synergyDB: deck.NewSynergyDatabase(),
			wantNil:   true,
		},
		{
			name:          "valid deck with synergies",
			deckCards:     []string{"Giant", "Witch", "Wizard", "Fireball", "Arrows", "Archers", "Tombstone", "Knight"},
			synergyDB:     deck.NewSynergyDatabase(),
			wantNil:       false,
			wantPairCount: 0, // Will vary based on database, just check it's non-nil
		},
		{
			name:          "log bait deck",
			deckCards:     []string{"Goblin Barrel", "Princess", "Rocket", "Inferno Tower", "Goblin Gang", "Knight", "Ice Spirit", "The Log"},
			synergyDB:     deck.NewSynergyDatabase(),
			wantNil:       false,
			wantPairCount: 0, // Will check synergies exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSynergyMatrix(tt.deckCards, tt.synergyDB)

			if tt.wantNil {
				if got != nil {
					t.Errorf("GenerateSynergyMatrix() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("GenerateSynergyMatrix() = nil, want non-nil")
			}

			// Verify structure
			if got.MaxPossiblePairs != 28 {
				t.Errorf("MaxPossiblePairs = %d, want 28", got.MaxPossiblePairs)
			}

			if got.AverageSynergy < 0 || got.AverageSynergy > 1.0 {
				t.Errorf("AverageSynergy = %f, want 0.0-1.0", got.AverageSynergy)
			}

			if got.SynergyCoverage < 0 || got.SynergyCoverage > 100 {
				t.Errorf("SynergyCoverage = %f, want 0.0-100.0", got.SynergyCoverage)
			}

			if got.PairCount != len(got.Pairs) {
				t.Errorf("PairCount = %d, but Pairs length = %d", got.PairCount, len(got.Pairs))
			}
		})
	}
}

func TestFormatSynergyMatrixText(t *testing.T) {
	db := deck.NewSynergyDatabase()

	tests := []struct {
		name         string
		deckCards    []string
		wantContains []string
	}{
		{
			name:      "log bait deck",
			deckCards: []string{"Goblin Barrel", "Princess", "Rocket", "Inferno Tower", "Goblin Gang", "Knight", "Ice Spirit", "The Log"},
			wantContains: []string{
				"SYNERGY MATRIX",
				"Total Synergies:",
				"Average Synergy:",
				"Overall Score:",
				"Synergy Grid",
				"Top Synergies:",
			},
		},
		{
			name:      "beatdown deck",
			deckCards: []string{"Golem", "Night Witch", "Baby Dragon", "Lightning", "Tornado", "Mega Minion", "Elixir Collector", "Lumberjack"},
			wantContains: []string{
				"SYNERGY MATRIX",
				"Total Synergies:",
				"Average Synergy:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matrix := GenerateSynergyMatrix(tt.deckCards, db)
			if matrix == nil {
				t.Fatal("GenerateSynergyMatrix() returned nil")
			}

			// Set a dummy total score for display
			matrix.TotalScore = 7.5

			got := FormatSynergyMatrixText(matrix, tt.deckCards, db)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("FormatSynergyMatrixText() missing %q in output", want)
				}
			}

			// Verify grid structure
			lines := strings.Split(got, "\n")
			if len(lines) < 10 {
				t.Errorf("FormatSynergyMatrixText() output too short: %d lines", len(lines))
			}
		})
	}
}

func TestFormatSynergyMatrixText_NilMatrix(t *testing.T) {
	got := FormatSynergyMatrixText(nil, []string{}, nil)
	want := "No synergy matrix available"

	if got != want {
		t.Errorf("FormatSynergyMatrixText(nil) = %q, want %q", got, want)
	}
}

func TestFormatSynergyMatrixJSON(t *testing.T) {
	db := deck.NewSynergyDatabase()
	deckCards := []string{"Giant", "Witch", "Wizard", "Fireball", "Arrows", "Archers", "Tombstone", "Knight"}

	matrix := GenerateSynergyMatrix(deckCards, db)
	if matrix == nil {
		t.Fatal("GenerateSynergyMatrix() returned nil")
	}
	matrix.TotalScore = 6.5

	got := FormatSynergyMatrixJSON(matrix)

	if got == nil {
		t.Fatal("FormatSynergyMatrixJSON() = nil")
	}

	// Check required fields
	requiredFields := []string{"pairs", "total_score", "average_synergy", "pair_count", "max_possible_pairs", "synergy_coverage"}
	for _, field := range requiredFields {
		if _, ok := got[field]; !ok {
			t.Errorf("FormatSynergyMatrixJSON() missing field %q", field)
		}
	}

	// Verify types and ranges
	if totalScore, ok := got["total_score"].(float64); ok {
		if totalScore != 6.5 {
			t.Errorf("total_score = %f, want 6.5", totalScore)
		}
	} else {
		t.Error("total_score is not float64")
	}

	if maxPairs, ok := got["max_possible_pairs"].(int); ok {
		if maxPairs != 28 {
			t.Errorf("max_possible_pairs = %d, want 28", maxPairs)
		}
	} else {
		t.Error("max_possible_pairs is not int")
	}
}

func TestFormatSynergyMatrixJSON_NilMatrix(t *testing.T) {
	got := FormatSynergyMatrixJSON(nil)
	if got != nil {
		t.Errorf("FormatSynergyMatrixJSON(nil) = %v, want nil", got)
	}
}

func TestGenerateTopSynergyNarrative(t *testing.T) {
	db := deck.NewSynergyDatabase()

	tests := []struct {
		name         string
		deckCards    []string
		wantMinLines int
	}{
		{
			name:         "log bait deck",
			deckCards:    []string{"Goblin Barrel", "Princess", "Rocket", "Inferno Tower", "Goblin Gang", "Knight", "Ice Spirit", "The Log"},
			wantMinLines: 2, // At least top synergy + overall
		},
		{
			name:         "beatdown deck",
			deckCards:    []string{"Golem", "Night Witch", "Baby Dragon", "Lightning", "Tornado", "Mega Minion", "Elixir Collector", "Lumberjack"},
			wantMinLines: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matrix := GenerateSynergyMatrix(tt.deckCards, db)
			if matrix == nil {
				t.Fatal("GenerateSynergyMatrix() returned nil")
			}

			got := GenerateTopSynergyNarrative(matrix)

			if len(got) < tt.wantMinLines {
				t.Errorf("GenerateTopSynergyNarrative() returned %d lines, want at least %d", len(got), tt.wantMinLines)
			}

			// Check that narratives contain expected elements
			for i, narrative := range got {
				if narrative == "" {
					t.Errorf("narrative[%d] is empty", i)
				}
			}

			// Last line should be overall assessment
			lastLine := got[len(got)-1]
			if !strings.Contains(lastLine, "Overall") {
				t.Errorf("Last narrative line should contain 'Overall', got %q", lastLine)
			}
		})
	}
}

func TestGenerateTopSynergyNarrative_NilMatrix(t *testing.T) {
	got := GenerateTopSynergyNarrative(nil)

	if len(got) != 1 {
		t.Errorf("GenerateTopSynergyNarrative(nil) returned %d lines, want 1", len(got))
	}

	want := "No synergies found in this deck"
	if got[0] != want {
		t.Errorf("GenerateTopSynergyNarrative(nil)[0] = %q, want %q", got[0], want)
	}
}

func TestAbbreviateCardName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "short name",
			input: "Giant",
			want:  "Giant",
		},
		{
			name:  "exactly 10 chars",
			input: "Exact Name",
			want:  "Exact Name",
		},
		{
			name:  "long name with The prefix",
			input: "The Royal Giant",
			want:  "Royal Gia…",
		},
		{
			name:  "long name without The",
			input: "Electro Wizard",
			want:  "Electro W…",
		},
		{
			name:  "very long name",
			input: "Three Musketeers",
			want:  "Three Mus…",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := abbreviateCardName(tt.input)
			if got != tt.want {
				t.Errorf("abbreviateCardName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatSynergyRating(t *testing.T) {
	tests := []struct {
		score float64
		want  string
	}{
		{0.95, "exceptional"},
		{0.9, "exceptional"},
		{0.85, "excellent"},
		{0.8, "excellent"},
		{0.75, "strong"},
		{0.7, "strong"},
		{0.65, "good"},
		{0.6, "good"},
		{0.55, "moderate"},
		{0.5, "moderate"},
		{0.4, "weak"},
		{0.3, "weak"},
		{0.2, "minimal"},
		{0.0, "minimal"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatSynergyRating(tt.score)
			if got != tt.want {
				t.Errorf("formatSynergyRating(%f) = %q, want %q", tt.score, got, tt.want)
			}
		})
	}
}

func TestSynergyMatrixPerformance(t *testing.T) {
	db := deck.NewSynergyDatabase()
	deckCards := []string{
		"Golem", "Night Witch", "Baby Dragon", "Lightning",
		"Tornado", "Mega Minion", "Elixir Collector", "Lumberjack",
	}

	// Run 100 iterations to get average performance
	iterations := 100
	start := time.Now()

	for range iterations {
		matrix := GenerateSynergyMatrix(deckCards, db)
		if matrix == nil {
			t.Fatal("GenerateSynergyMatrix() returned nil")
		}

		_ = FormatSynergyMatrixText(matrix, deckCards, db)
		_ = FormatSynergyMatrixJSON(matrix)
		_ = GenerateTopSynergyNarrative(matrix)
	}

	elapsed := time.Since(start)
	avgTime := elapsed / time.Duration(iterations)

	// Should be well under 200ms per operation
	maxAllowed := 200 * time.Millisecond
	if avgTime > maxAllowed {
		t.Errorf("Average operation time %v exceeds limit %v", avgTime, maxAllowed)
	}

	t.Logf("Average operation time: %v (limit: %v)", avgTime, maxAllowed)
}

func TestSynergyMatrixIntegration(t *testing.T) {
	// Test with a real deck that has known synergies
	db := deck.NewSynergyDatabase()
	deckCards := []string{
		"Golem", "Night Witch", "Baby Dragon", "Lightning",
		"Tornado", "Mega Minion", "Elixir Collector", "Lumberjack",
	}

	// Generate matrix
	matrix := GenerateSynergyMatrix(deckCards, db)
	if matrix == nil {
		t.Fatal("GenerateSynergyMatrix() returned nil")
	}

	// Should have some synergies
	if matrix.PairCount == 0 {
		t.Error("Expected some synergies in Golem deck, got 0")
	}

	// Check that Golem + Night Witch synergy exists (known strong pair)
	foundGolemNW := false
	for _, pair := range matrix.Pairs {
		if (pair.Card1 == "Golem" && pair.Card2 == "Night Witch") ||
			(pair.Card1 == "Night Witch" && pair.Card2 == "Golem") {
			foundGolemNW = true
			if pair.Score < 0.8 {
				t.Errorf("Golem + Night Witch synergy score = %f, want >= 0.8", pair.Score)
			}
		}
	}

	if !foundGolemNW {
		t.Error("Expected to find Golem + Night Witch synergy in top pairs")
	}

	// Verify text formatting produces valid output
	text := FormatSynergyMatrixText(matrix, deckCards, db)
	if len(text) < 100 {
		t.Errorf("FormatSynergyMatrixText() output too short: %d chars", len(text))
	}

	// Verify JSON formatting
	jsonData := FormatSynergyMatrixJSON(matrix)
	if jsonData == nil {
		t.Error("FormatSynergyMatrixJSON() returned nil")
	}

	// Verify narrative generation
	narratives := GenerateTopSynergyNarrative(matrix)
	if len(narratives) == 0 {
		t.Error("GenerateTopSynergyNarrative() returned empty array")
	}
}
