package taxonomy

import "testing"

func TestSharedGroupsContainExpectedCards(t *testing.T) {
	tests := []struct {
		name  string
		cards []string
		want  string
	}{
		{name: "beatdown heavy tanks", cards: BeatdownHeavyTanks(), want: "Golem"},
		{name: "beatdown support", cards: BeatdownSupport(), want: "Baby Dragon"},
		{name: "control buildings", cards: ControlDefensiveBuildings(), want: "Tesla"},
		{name: "control big spells", cards: ControlBigSpells(), want: "Poison"},
		{name: "cycle win conditions", cards: CycleWinConditions(), want: "Hog Rider"},
		{name: "cycle core cards", cards: CycleCoreCards(), want: "Skeletons"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !contains(tt.cards, tt.want) {
				t.Fatalf("expected %q to be present", tt.want)
			}
		})
	}
}

func TestCloneReturnsIndependentSlice(t *testing.T) {
	original := []string{"A", "B"}
	cloned := Clone(original)
	cloned[0] = "changed"

	if original[0] == "changed" {
		t.Fatalf("clone mutated original slice")
	}
}

func TestMergeConcatenatesGroups(t *testing.T) {
	merged := Merge([]string{"A"}, []string{"B", "C"})
	if len(merged) != 3 || merged[0] != "A" || merged[1] != "B" || merged[2] != "C" {
		t.Fatalf("unexpected merge result: %#v", merged)
	}
}

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

func contains(cards []string, want string) bool {
	for _, card := range cards {
		if card == want {
			return true
		}
	}
	return false
}
