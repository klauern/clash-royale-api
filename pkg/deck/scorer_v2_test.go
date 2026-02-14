package deck

import (
	"math"
	"testing"
)

func TestScorerV2NormalizationFactor(t *testing.T) {
	tests := []struct {
		name             string
		uniquenessWeight float64
		want             float64
	}{
		{
			name:             "base v2 weights",
			uniquenessWeight: 0.0,
			want:             1.0 / 1.5,
		},
		{
			name:             "base plus uniqueness",
			uniquenessWeight: 0.15,
			want:             1.0 / 1.65,
		},
		{
			name:             "zero total guard",
			uniquenessWeight: -1.5,
			want:             0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scorerV2NormalizationFactor(tt.uniquenessWeight)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Fatalf("scorerV2NormalizationFactor(%v)=%v, want %v", tt.uniquenessWeight, got, tt.want)
			}
		})
	}
}
