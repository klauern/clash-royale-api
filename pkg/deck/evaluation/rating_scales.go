package evaluation

import (
	"fmt"
	"math"
)

// ScoreToRating converts a numeric score (0-10) to a qualitative Rating
func ScoreToRating(score float64) Rating {
	// Clamp score to 0-10 range
	if score < 0 {
		score = 0
	}
	if score > 10 {
		score = 10
	}

	switch {
	case score >= 9.0:
		return RatingGodly
	case score >= 8.0:
		return RatingAmazing
	case score >= 7.0:
		return RatingGreat
	case score >= 6.0:
		return RatingGood
	case score >= 5.0:
		return RatingDecent
	case score >= 4.0:
		return RatingMediocre
	case score >= 3.0:
		return RatingPoor
	case score >= 2.0:
		return RatingBad
	case score >= 1.0:
		return RatingTerrible
	default:
		return RatingAwful
	}
}

// ScoreToStars converts a numeric score (0-10) to a star rating (1-3)
// - 1 star: 0.0-4.9 (Poor to Awful)
// - 2 stars: 5.0-7.9 (Decent to Great)
// - 3 stars: 8.0-10.0 (Amazing to Godly)
func ScoreToStars(score float64) int {
	// Clamp score to 0-10 range
	if score < 0 {
		score = 0
	}
	if score > 10 {
		score = 10
	}

	switch {
	case score >= 8.0:
		return 3
	case score >= 5.0:
		return 2
	default:
		return 1
	}
}

// RatingColor returns a terminal color code for a rating
// Used for colorful terminal output
func RatingColor(rating Rating) string {
	switch rating {
	case RatingGodly, RatingAmazing:
		return "\033[1;32m" // Bright green
	case RatingGreat, RatingGood:
		return "\033[0;32m" // Green
	case RatingDecent:
		return "\033[0;33m" // Yellow
	case RatingMediocre, RatingPoor:
		return "\033[0;31m" // Red
	case RatingBad, RatingTerrible, RatingAwful:
		return "\033[1;31m" // Bright red
	default:
		return "\033[0m" // Reset
	}
}

// ColorReset returns the terminal color reset code
func ColorReset() string {
	return "\033[0m"
}

// FormatScore formats a score with rating and stars
// Example: "8.5/10 (Amazing) ★★★"
func FormatScore(score float64, includeColor bool) string {
	rating := ScoreToRating(score)
	stars := ScoreToStars(score)
	starsStr := ""
	for range stars {
		starsStr += "★"
	}

	formattedScore := fmt.Sprintf("%.1f", roundToOne(score))

	if includeColor {
		color := RatingColor(rating)
		return fmt.Sprintf("%s%s/10 (%s) %s%s", color, formattedScore, rating.String(), starsStr, ColorReset())
	}

	return fmt.Sprintf("%s/10 (%s) %s", formattedScore, rating.String(), starsStr)
}

// CreateCategoryScore creates a CategoryScore from a numeric score
func CreateCategoryScore(score float64, assessment string) CategoryScore {
	roundedScore := roundToOne(score)
	return CategoryScore{
		Score:      roundedScore,
		Rating:     ScoreToRating(roundedScore),
		Assessment: assessment,
		Stars:      ScoreToStars(roundedScore),
	}
}

// roundToOne rounds a float to 1 decimal place
func roundToOne(f float64) float64 {
	return math.Round(f*10) / 10
}
