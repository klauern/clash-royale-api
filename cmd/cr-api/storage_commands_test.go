package main

import "testing"

func TestHumanReadableBytes(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{999, "999 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
	}

	for _, tt := range tests {
		got := humanReadableBytes(tt.size)
		if got != tt.want {
			t.Errorf("humanReadableBytes(%d) = %q, want %q", tt.size, got, tt.want)
		}
	}
}
