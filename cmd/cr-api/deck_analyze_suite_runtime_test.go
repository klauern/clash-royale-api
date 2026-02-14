package main

import (
	"strings"
	"testing"
)

func TestCalculateCompareCount(t *testing.T) {
	tests := []struct {
		name        string
		resultCount int
		topN        int
		wantCount   int
		wantErr     string
	}{
		{
			name:        "requires two results",
			resultCount: 1,
			topN:        5,
			wantErr:     "need at least 2 evaluated decks to compare, got 1",
		},
		{
			name:        "caps to available result count",
			resultCount: 3,
			topN:        5,
			wantCount:   3,
		},
		{
			name:        "minimum compare count",
			resultCount: 4,
			topN:        1,
			wantCount:   2,
		},
		{
			name:        "maximum compare count",
			resultCount: 10,
			topN:        8,
			wantCount:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateCompareCount(tt.resultCount, tt.topN)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantCount {
				t.Fatalf("calculateCompareCount(%d, %d)=%d, want %d", tt.resultCount, tt.topN, got, tt.wantCount)
			}
		})
	}
}
