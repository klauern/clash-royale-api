package main

import (
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func testDeckSpaceStats() *deck.DeckSpaceStats {
	return &deck.DeckSpaceStats{
		TotalCards:        42,
		CardsByRole:       map[deck.CardRole]int{deck.RoleCycle: 6, deck.RoleSupport: 5, deck.RoleSpellSmall: 4, deck.RoleSpellBig: 3, deck.RoleBuilding: 2, deck.RoleWinCondition: 1},
		TotalCombinations: big.NewInt(123456),
		ValidCombinations: big.NewInt(4567),
		ByElixirRange:     map[string]*big.Int{"Medium (3.5-4.0)": big.NewInt(400), "Fast (2.5-3.0)": big.NewInt(200)},
		ByArchetype:       map[string]*big.Int{"Control": big.NewInt(300), "Bait": big.NewInt(250), "Zap Bait": big.NewInt(150)},
	}
}

func TestOrderedPossibleCountArchetypeRows(t *testing.T) {
	rows := orderedPossibleCountArchetypeRows(testDeckSpaceStats())
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	want := []string{"Control", "Bait", "Zap Bait"}
	for i, label := range want {
		if rows[i].Label != label {
			t.Fatalf("row %d label = %q, want %q", i, rows[i].Label, label)
		}
	}
}

func TestFormatPossibleCountCSVVerboseUsesSharedRoleOrder(t *testing.T) {
	output := formatPossibleCountCSV(&clashroyale.Player{Name: "Test", Tag: "#TAG"}, testDeckSpaceStats(), true)

	wantOrder := []string{
		"Win Condition,1",
		"Building,2",
		"Big Spell,3",
		"Small Spell,4",
		"Support,5",
		"Cycle,6",
	}

	lastIndex := -1
	for _, fragment := range wantOrder {
		index := strings.Index(output, fragment)
		if index == -1 {
			t.Fatalf("output missing %q:\n%s", fragment, output)
		}
		if index <= lastIndex {
			t.Fatalf("fragment %q appeared out of order", fragment)
		}
		lastIndex = index
	}
}

func TestFormatPossibleCountJSONIncludesOrderedSections(t *testing.T) {
	output, err := formatPossibleCountJSON(&clashroyale.Player{Name: "Test", Tag: "#TAG"}, testDeckSpaceStats())
	if err != nil {
		t.Fatalf("formatPossibleCountJSON returned error: %v", err)
	}

	var parsed struct {
		ByElixirRange        map[string]string       `json:"by_elixir_range"`
		ByElixirRangeOrdered []possibleCountCountRow `json:"by_elixir_range_ordered"`
		ByArchetype          map[string]string       `json:"by_archetype"`
		ByArchetypeOrdered   []possibleCountCountRow `json:"by_archetype_ordered"`
		CardsByRoleOrdered   []possibleCountRoleRow  `json:"cards_by_role_ordered"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if parsed.ByElixirRange["Fast (2.5-3.0)"] != "200" {
		t.Fatalf("legacy elixir map missing expected value: %#v", parsed.ByElixirRange)
	}
	if len(parsed.ByElixirRangeOrdered) != 2 || parsed.ByElixirRangeOrdered[0].Label != "Fast (2.5-3.0)" || parsed.ByElixirRangeOrdered[1].Label != "Medium (3.5-4.0)" {
		t.Fatalf("unexpected ordered elixir rows: %#v", parsed.ByElixirRangeOrdered)
	}
	if len(parsed.ByArchetypeOrdered) != 3 || parsed.ByArchetypeOrdered[0].Label != "Control" || parsed.ByArchetypeOrdered[1].Label != "Bait" || parsed.ByArchetypeOrdered[2].Label != "Zap Bait" {
		t.Fatalf("unexpected ordered archetype rows: %#v", parsed.ByArchetypeOrdered)
	}
	if len(parsed.CardsByRoleOrdered) != 6 || parsed.CardsByRoleOrdered[0].Role != "Win Condition" || parsed.CardsByRoleOrdered[5].Role != "Cycle" {
		t.Fatalf("unexpected ordered role rows: %#v", parsed.CardsByRoleOrdered)
	}
}
