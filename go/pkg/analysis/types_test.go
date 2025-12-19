package analysis

import (
	"math"
	"testing"
)

func TestCardLevelInfo_LevelRatio(t *testing.T) {
	tests := []struct {
		name string
		info CardLevelInfo
		want float64
	}{
		{
			name: "no evolution",
			info: CardLevelInfo{Level: 7, MaxLevel: 14},
			want: 0.5,
		},
		{
			name: "evolution available but not max level",
			info: CardLevelInfo{Level: 10, MaxLevel: 14, EvolutionLevel: 1, MaxEvolutionLevel: 3},
			want: 10.0 / 14.0,
		},
		{
			name: "max level with evolution progress",
			info: CardLevelInfo{Level: 14, MaxLevel: 14, EvolutionLevel: 1, MaxEvolutionLevel: 3},
			want: 15.0 / 17.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.LevelRatio()
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("LevelRatio() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCardLevelInfo_ProgressToNext(t *testing.T) {
	tests := []struct {
		name string
		info CardLevelInfo
		want float64
	}{
		{
			name: "cards toward next level",
			info: CardLevelInfo{CardCount: 40, CardsToNext: 100, IsMaxLevel: false},
			want: 40.0,
		},
		{
			name: "max level without evolution",
			info: CardLevelInfo{IsMaxLevel: true},
			want: 100.0,
		},
		{
			name: "max level with evolution progress",
			info: CardLevelInfo{IsMaxLevel: true, EvolutionLevel: 1, MaxEvolutionLevel: 3},
			want: (1.0 / 3.0) * 100.0,
		},
		{
			name: "evolution over max caps to 100",
			info: CardLevelInfo{IsMaxLevel: true, EvolutionLevel: 4, MaxEvolutionLevel: 3},
			want: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.ProgressToNext()
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("ProgressToNext() = %v, want %v", got, tt.want)
			}
		})
	}
}
