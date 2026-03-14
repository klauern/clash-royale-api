package taxonomy

import "testing"

func TestContainsAnySubstringFold(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		signals []string
		want    bool
	}{
		{
			name:    "matches mixed case signal",
			values:  []string{"Mega Knight", "Battle Ram"},
			signals: []string{"battle ram"},
			want:    true,
		},
		{
			name:    "no matching signal",
			values:  []string{"Skeletons", "Archers"},
			signals: []string{"golem"},
			want:    false,
		},
		{
			name:    "empty values",
			values:  []string{},
			signals: []string{"golem"},
			want:    false,
		},
		{
			name:    "empty signals",
			values:  []string{"Golem"},
			signals: []string{},
			want:    false,
		},
		{
			name:    "partial substring match",
			values:  []string{"Mega Knight Giant Skeleton"},
			signals: []string{"giant"},
			want:    true,
		},
		{
			name:    "multiple matching signals",
			values:  []string{"X-Bow", "Mortar"},
			signals: []string{"x-bow", "mortar"},
			want:    true,
		},
		{
			name:    "signal matching remains case-insensitive",
			values:  []string{"Battle Ram"},
			signals: []string{"Battle Ram"},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsAnySubstringFold(tt.values, tt.signals)
			if got != tt.want {
				t.Fatalf("ContainsAnySubstringFold() = %v, want %v", got, tt.want)
			}
		})
	}
}
