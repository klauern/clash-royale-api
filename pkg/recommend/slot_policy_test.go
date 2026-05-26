package recommend

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

func card(name string, st SlotType) CardWithSlotType {
	return CardWithSlotType{Name: name, SlotType: st}
}

// --- SlotType metadata tests ---

func TestCardSlotTyper_Champion(t *testing.T) {
	typeer := NewCardSlotTyper([]clashroyale.Card{
		{Name: "Archer Queen", Rarity: "Champion"},
		{Name: "Skeleton Army", Rarity: "Common", MaxEvolutionLevel: 1},
		{Name: "Balloon", Rarity: "Epic", MaxEvolutionLevel: 2},
		{Name: "Bowler", Rarity: "Epic", MaxEvolutionLevel: 2},
	})

	cases := []struct {
		name    string
		evolved bool
		want    SlotType
	}{
		{"Archer Queen", false, Champion},
		{"Archer Queen", true, Champion}, // champion regardless of evolution field
		{"Skeleton Army", true, RegularEvo},
		{"Skeleton Army", false, RegularCard},
		{"Balloon", true, ChampionSlotOnlyEvo},
		{"Balloon", false, RegularCard},
		{"Bowler", true, RegularEvo}, // regular Bowler evolved = RegularEvo
		{"Bowler", false, RegularCard},
		{"Unknown", false, RegularCard},
	}

	for _, tc := range cases {
		got := typeer.SlotTypeFor(tc.name, tc.evolved)
		if got != tc.want {
			t.Errorf("SlotTypeFor(%q, evolved=%v) = %v, want %v", tc.name, tc.evolved, got, tc.want)
		}
	}
}

// --- DeckSlotPolicy.Validate tests ---

func TestValidate_LegalDecks(t *testing.T) {
	p := DefaultPolicy()
	cases := []struct {
		name string
		deck []CardWithSlotType
	}{
		{"no special cards", []CardWithSlotType{card("Knight", RegularCard)}},
		{"1 evo only", []CardWithSlotType{card("Skeleton Army", RegularEvo)}},
		{"1 champ only", []CardWithSlotType{card("Archer Queen", Champion)}},
		{"1 champSlotEvo only", []CardWithSlotType{card("Balloon", ChampionSlotOnlyEvo)}},
		{"evo + champion", []CardWithSlotType{
			card("Skeleton Army", RegularEvo),
			card("Archer Queen", Champion),
		}},
		{"evo + champSlotEvo", []CardWithSlotType{
			card("Skeleton Army", RegularEvo),
			card("Balloon", ChampionSlotOnlyEvo),
		}},
		{"2 evos + champion (evo, flex, champ slots)", []CardWithSlotType{
			card("Skeleton Army", RegularEvo),
			card("Witch", RegularEvo),
			card("Archer Queen", Champion),
		}},
		{"champSlotEvo + evo + champ", []CardWithSlotType{
			card("Balloon", ChampionSlotOnlyEvo),
			card("Skeleton Army", RegularEvo),
			card("Archer Queen", Champion),
		}},
	}

	for _, tc := range cases {
		if err := p.Validate(tc.deck); err != nil {
			t.Errorf("Validate(%q): unexpected error: %v", tc.name, err)
		}
	}
}

func TestValidate_IllegalDecks(t *testing.T) {
	p := DefaultPolicy()
	cases := []struct {
		name string
		deck []CardWithSlotType
	}{
		{"2 champSlotEvos", []CardWithSlotType{
			card("Balloon", ChampionSlotOnlyEvo),
			card("Balloon2", ChampionSlotOnlyEvo),
		}},
		{"3 evos", []CardWithSlotType{
			card("SkelArmy", RegularEvo),
			card("Witch", RegularEvo),
			card("Goblin Barrel", RegularEvo),
		}},
		{"2 champions", []CardWithSlotType{
			card("Archer Queen", Champion),
			card("Mighty Miner", Champion),
		}},
		{"2 evos + 2 champions exceeds flex capacity", []CardWithSlotType{
			card("SkelArmy", RegularEvo),
			card("Witch", RegularEvo),
			card("Archer Queen", Champion),
			card("Mighty Miner", Champion),
		}},
	}

	for _, tc := range cases {
		if err := p.Validate(tc.deck); err == nil {
			t.Errorf("Validate(%q): expected error but got nil", tc.name)
		}
	}
}

// --- Combinator tests ---

func TestEnumerateAssignments_BalloonNeverInEvoSlot(t *testing.T) {
	deck := []CardWithSlotType{
		card("Balloon", ChampionSlotOnlyEvo),
		card("Skeleton Army", RegularEvo),
		card("Archer Queen", Champion),
		card("Knight", RegularCard),
	}

	candidates := CollectCandidates(deck)
	policy := DefaultPolicy()
	scorer := func(a SlotAssignment) float64 { return 1.0 }

	assignments := EnumerateAssignments(candidates, policy, scorer, 0)

	for _, a := range assignments {
		if a.EvoSlot != nil && a.EvoSlot.Name == "Balloon" {
			t.Errorf("Balloon appeared in EvoSlot — must never be in evo slot")
		}
	}
	if len(assignments) == 0 {
		t.Fatal("expected at least one valid assignment")
	}
}

func TestEnumerateAssignments_BowlerChampionVsRegular(t *testing.T) {
	// Simulate a deck with both regular Bowler (evo) and Bowler-champion
	// They are separate named cards so they won't collide.
	deck := []CardWithSlotType{
		card("Bowler", RegularEvo), // Epic Bowler evolved
		card("Archer Queen", Champion),
	}

	candidates := CollectCandidates(deck)
	policy := DefaultPolicy()
	scorer := func(a SlotAssignment) float64 { return 1.0 }

	assignments := EnumerateAssignments(candidates, policy, scorer, 0)
	if len(assignments) == 0 {
		t.Fatal("expected assignments for Bowler-evo + champion deck")
	}
	for _, a := range assignments {
		// Bowler as RegularEvo should appear in EvoSlot or FlexSlot, never ChampionSlot
		if a.ChampionSlot != nil && a.ChampionSlot.Name == "Bowler" &&
			a.ChampionSlot.SlotType == RegularEvo {
			t.Errorf("RegularEvo Bowler incorrectly placed in ChampionSlot")
		}
	}
}

func TestEnumerateAssignments_TopN(t *testing.T) {
	deck := []CardWithSlotType{
		card("Skeleton Army", RegularEvo),
		card("Witch", RegularEvo),
		card("Archer Queen", Champion),
	}

	candidates := CollectCandidates(deck)
	policy := DefaultPolicy()
	scorer := func(_ SlotAssignment) float64 { return 1.0 }

	assignments := EnumerateAssignments(candidates, policy, scorer, 3)
	if len(assignments) > 3 {
		t.Errorf("expected at most 3 assignments, got %d", len(assignments))
	}
}

// TestEnumerateAssignments_ZyLoganDeck verifies the third acceptance criterion:
// Balloon surfaces in champion slot and Skeleton Army surfaces in evo slot.
func TestEnumerateAssignments_ZyLoganDeck(t *testing.T) {
	deck := []CardWithSlotType{
		card("Balloon", ChampionSlotOnlyEvo),
		card("Skeleton Army", RegularEvo),
		card("Witch", RegularEvo),
		card("Miner", Champion),
	}

	candidates := CollectCandidates(deck)
	assignments := EnumerateAssignments(candidates, DefaultPolicy(), DefaultSlotScorer, 0)

	if len(assignments) == 0 {
		t.Fatal("expected at least one valid assignment for ZyLogan deck")
	}

	// Balloon must never appear in the evo slot (ChampionSlotOnlyEvo constraint).
	for _, a := range assignments {
		if a.EvoSlot != nil && a.EvoSlot.Name == "Balloon" {
			t.Errorf("Balloon appeared in EvoSlot — violation of ChampionSlotOnlyEvo constraint")
		}
	}

	// Top assignment must have Balloon in champion slot.
	top := assignments[0]
	if top.ChampionSlot == nil || top.ChampionSlot.Name != "Balloon" {
		t.Errorf("top assignment champion slot = %v, want Balloon", top.ChampionSlot)
	}

	// At least one top-scoring assignment must have Skeleton Army in the evo slot.
	hasSkelArmyInEvo := false
	topScore := assignments[0].Score
	for _, a := range assignments {
		if a.Score < topScore {
			break
		}
		if a.EvoSlot != nil && a.EvoSlot.Name == "Skeleton Army" {
			hasSkelArmyInEvo = true
		}
	}
	if !hasSkelArmyInEvo {
		t.Errorf("no top-scoring assignment has Skeleton Army in evo slot")
	}
}

func TestEnumerateAssignments_SortedByScore(t *testing.T) {
	deck := []CardWithSlotType{
		card("Skeleton Army", RegularEvo),
		card("Witch", RegularEvo),
		card("Archer Queen", Champion),
	}

	candidates := CollectCandidates(deck)
	policy := DefaultPolicy()
	i := 0
	scorer := func(_ SlotAssignment) float64 {
		i++
		return float64(100 - i)
	}

	assignments := EnumerateAssignments(candidates, policy, scorer, 0)
	for j := 1; j < len(assignments); j++ {
		if assignments[j].Score > assignments[j-1].Score {
			t.Errorf("assignments not sorted descending at index %d", j)
		}
	}
}
