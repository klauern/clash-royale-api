package evaluation

import (
	"strings"
	"testing"
)

func TestScoreToRating(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected Rating
	}{
		{"Godly upper bound", 10.0, RatingGodly},
		{"Godly lower bound", 9.0, RatingGodly},
		{"Amazing upper bound", 8.9, RatingAmazing},
		{"Amazing lower bound", 8.0, RatingAmazing},
		{"Great upper bound", 7.9, RatingGreat},
		{"Great lower bound", 7.0, RatingGreat},
		{"Good upper bound", 6.9, RatingGood},
		{"Good lower bound", 6.0, RatingGood},
		{"Decent upper bound", 5.9, RatingDecent},
		{"Decent lower bound", 5.0, RatingDecent},
		{"Mediocre upper bound", 4.9, RatingMediocre},
		{"Mediocre lower bound", 4.0, RatingMediocre},
		{"Poor upper bound", 3.9, RatingPoor},
		{"Poor lower bound", 3.0, RatingPoor},
		{"Bad upper bound", 2.9, RatingBad},
		{"Bad lower bound", 2.0, RatingBad},
		{"Terrible upper bound", 1.9, RatingTerrible},
		{"Terrible lower bound", 1.0, RatingTerrible},
		{"Awful upper bound", 0.9, RatingAwful},
		{"Awful lower bound", 0.0, RatingAwful},
		{"Above max clamped", 15.0, RatingGodly},
		{"Below min clamped", -5.0, RatingAwful},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScoreToRating(tt.score)
			if result != tt.expected {
				t.Errorf("ScoreToRating(%v) = %v, want %v", tt.score, result, tt.expected)
			}
		})
	}
}

func TestScoreToStars(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected int
	}{
		{"3 stars - max", 10.0, 3},
		{"3 stars - lower bound", 8.0, 3},
		{"2 stars - upper bound", 7.9, 2},
		{"2 stars - lower bound", 5.0, 2},
		{"1 star - upper bound", 4.9, 1},
		{"1 star - lower bound", 0.0, 1},
		{"Above max clamped", 15.0, 3},
		{"Below min clamped", -5.0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScoreToStars(tt.score)
			if result != tt.expected {
				t.Errorf("ScoreToStars(%v) = %v, want %v", tt.score, result, tt.expected)
			}
		})
	}
}

func TestRatingColor(t *testing.T) {
	tests := []struct {
		name     string
		rating   Rating
		expected string
	}{
		{"Godly color", RatingGodly, "\033[1;32m"},
		{"Amazing color", RatingAmazing, "\033[1;32m"},
		{"Great color", RatingGreat, "\033[0;32m"},
		{"Good color", RatingGood, "\033[0;32m"},
		{"Decent color", RatingDecent, "\033[0;33m"},
		{"Mediocre color", RatingMediocre, "\033[0;31m"},
		{"Poor color", RatingPoor, "\033[0;31m"},
		{"Bad color", RatingBad, "\033[1;31m"},
		{"Terrible color", RatingTerrible, "\033[1;31m"},
		{"Awful color", RatingAwful, "\033[1;31m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RatingColor(tt.rating)
			if result != tt.expected {
				t.Errorf("RatingColor(%v) = %v, want %v", tt.rating, result, tt.expected)
			}
		})
	}
}

func TestColorReset(t *testing.T) {
	expected := "\033[0m"
	result := ColorReset()
	if result != expected {
		t.Errorf("ColorReset() = %v, want %v", result, expected)
	}
}

func TestFormatScore(t *testing.T) {
	tests := []struct {
		name         string
		score        float64
		includeColor bool
		contains     []string // Parts that should be in the output
	}{
		{
			name:         "High score without color",
			score:        8.7,
			includeColor: false,
			contains:     []string{"8.7/10", "Amazing", "★★★"},
		},
		{
			name:         "Medium score without color",
			score:        6.3,
			includeColor: false,
			contains:     []string{"6.3/10", "Good", "★★"},
		},
		{
			name:         "Low score without color",
			score:        2.5,
			includeColor: false,
			contains:     []string{"2.5/10", "Bad", "★"},
		},
		{
			name:         "Score with color",
			score:        9.2,
			includeColor: true,
			contains:     []string{"9.2/10", "Godly", "★★★", "\033["},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatScore(tt.score, tt.includeColor)
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("FormatScore(%v, %v) = %v, should contain %v", tt.score, tt.includeColor, result, substr)
				}
			}
		})
	}
}

func TestCreateCategoryScore(t *testing.T) {
	tests := []struct {
		name           string
		score          float64
		assessment     string
		expectedRating Rating
		expectedStars  int
	}{
		{
			name:           "High category score",
			score:          8.5,
			assessment:     "Excellent defense",
			expectedRating: RatingAmazing,
			expectedStars:  3,
		},
		{
			name:           "Medium category score",
			score:          6.2,
			assessment:     "Good attack options",
			expectedRating: RatingGood,
			expectedStars:  2,
		},
		{
			name:           "Low category score",
			score:          3.8,
			assessment:     "Poor synergy",
			expectedRating: RatingPoor,
			expectedStars:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateCategoryScore(tt.score, tt.assessment)

			if result.Rating != tt.expectedRating {
				t.Errorf("CreateCategoryScore(%v, %v).Rating = %v, want %v",
					tt.score, tt.assessment, result.Rating, tt.expectedRating)
			}

			if result.Stars != tt.expectedStars {
				t.Errorf("CreateCategoryScore(%v, %v).Stars = %v, want %v",
					tt.score, tt.assessment, result.Stars, tt.expectedStars)
			}

			if result.Assessment != tt.assessment {
				t.Errorf("CreateCategoryScore(%v, %v).Assessment = %v, want %v",
					tt.score, tt.assessment, result.Assessment, tt.assessment)
			}

			// Check that score is rounded to 1 decimal
			if result.Score != roundToOne(tt.score) {
				t.Errorf("CreateCategoryScore(%v, %v).Score = %v, want %v",
					tt.score, tt.assessment, result.Score, roundToOne(tt.score))
			}
		})
	}
}

func TestRoundToOne(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"Round up", 8.67, 8.7},
		{"Round down", 8.62, 8.6},
		{"Already rounded", 8.5, 8.5},
		{"Round to integer", 8.04, 8.0},
		{"Exact half rounds up", 8.55, 8.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roundToOne(tt.input)
			if result != tt.expected {
				t.Errorf("roundToOne(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestRatingString tests the String() method on Rating type
func TestRatingString(t *testing.T) {
	tests := []struct {
		rating   Rating
		expected string
	}{
		{RatingGodly, "Godly!"},
		{RatingAmazing, "Amazing"},
		{RatingGreat, "Great"},
		{RatingGood, "Good"},
		{RatingDecent, "Decent"},
		{RatingMediocre, "Mediocre"},
		{RatingPoor, "Poor"},
		{RatingBad, "Bad"},
		{RatingTerrible, "Terrible"},
		{RatingAwful, "Awful"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.rating.String()
			if result != tt.expected {
				t.Errorf("Rating.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}
