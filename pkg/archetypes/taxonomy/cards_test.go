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

func contains(cards []string, want string) bool {
	for _, card := range cards {
		if card == want {
			return true
		}
	}
	return false
}
