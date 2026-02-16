package main

import "testing"

func TestParseBoostedCardLevelEntry(t *testing.T) {
	tests := []struct {
		name      string
		entry     string
		wantName  string
		wantLevel int
		wantErr   bool
	}{
		{
			name:      "valid",
			entry:     "Knight:14",
			wantName:  "Knight",
			wantLevel: 14,
		},
		{
			name:      "trimmed",
			entry:     "  Knight  :  15  ",
			wantName:  "Knight",
			wantLevel: 15,
		},
		{
			name:    "non-numeric suffix rejected",
			entry:   "Knight:10abc",
			wantErr: true,
		},
		{
			name:    "out of range rejected",
			entry:   "Knight:17",
			wantErr: true,
		},
		{
			name:    "empty rejected",
			entry:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotLevel, err := parseBoostedCardLevelEntry(tt.entry)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotName != tt.wantName {
				t.Fatalf("name = %q, want %q", gotName, tt.wantName)
			}
			if gotLevel != tt.wantLevel {
				t.Fatalf("level = %d, want %d", gotLevel, tt.wantLevel)
			}
		})
	}
}
