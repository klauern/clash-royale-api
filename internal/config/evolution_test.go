package config

import (
	"math"
	"testing"
)

func TestEvolutionProgressBonus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		level    int
		maxLevel int
		weight   float64
		want     float64
	}{
		{name: "no progress", level: 0, maxLevel: 3, weight: 0.15, want: 0},
		{name: "invalid maximum", level: 1, maxLevel: 0, weight: 0.15, want: 0},
		{name: "partial progress", level: 1, maxLevel: 3, weight: 0.15, want: 0.05},
		{name: "full progress", level: 3, maxLevel: 3, weight: 0.15, want: 0.15},
		{name: "over maximum", level: 4, maxLevel: 3, weight: 0.15, want: 0.15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := EvolutionProgressBonus(tt.level, tt.maxLevel, tt.weight); math.Abs(got-tt.want) > 1e-12 {
				t.Fatalf("EvolutionProgressBonus(%d, %d, %v) = %v, want %v", tt.level, tt.maxLevel, tt.weight, got, tt.want)
			}
		})
	}
}
