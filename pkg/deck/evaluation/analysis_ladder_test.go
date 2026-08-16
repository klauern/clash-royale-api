package evaluation

import "testing"

func TestCalculateLadderScorePreservesTenPointMaximum(t *testing.T) {
	t.Parallel()

	if got := calculateLadderScore(10, 10, 10); got != 10 {
		t.Fatalf("calculateLadderScore() = %.1f, want 10.0", got)
	}
}
