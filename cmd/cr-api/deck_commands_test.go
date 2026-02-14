package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestLoadDeckCandidatesFromFile_UsesDeckDetail(t *testing.T) {
	t.Helper()

	payload := deckFilePayload{
		Deck: []string{
			"Ice Golem",
			"Princess",
			"Poison",
			"Vines",
			"Goblin Hut",
			"Graveyard",
			"Ice Wizard",
			"Bomber",
		},
		DeckDetail: []deck.CardDetail{
			{Name: "Ice Golem", Elixir: 2, Rarity: "Rare", Role: "cycle"},
			{Name: "Princess", Elixir: 3, Rarity: "Legendary", Role: "support"},
			{Name: "Poison", Elixir: 4, Rarity: "Epic", Role: "spells_big"},
			{Name: "Vines", Elixir: 2, Rarity: "Epic", Role: "spells_small"},
			{Name: "Goblin Hut", Elixir: 4, Rarity: "Rare", Role: "buildings"},
			{Name: "Graveyard", Elixir: 5, Rarity: "Legendary", Role: "win_conditions"},
			{Name: "Ice Wizard", Elixir: 3, Rarity: "Legendary", Role: "support"},
			{Name: "Bomber", Elixir: 2, Rarity: "Common", Role: "cycle"},
		},
	}

	dir := t.TempDir()
	path := dir + "/deck.json"

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("failed to write deck file: %v", err)
	}

	candidates, ok, err := loadDeckCandidatesFromFile(path)
	if err != nil {
		t.Fatalf("loadDeckCandidatesFromFile returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected deck detail to be used")
	}
	if len(candidates) != 8 {
		t.Fatalf("expected 8 candidates, got %d", len(candidates))
	}

	var iceGolemElixir int
	for _, candidate := range candidates {
		if candidate.Name == "Ice Golem" {
			iceGolemElixir = candidate.Elixir
			break
		}
	}
	if iceGolemElixir != 2 {
		t.Fatalf("Ice Golem elixir = %d, want 2", iceGolemElixir)
	}
}

func TestConvertToCardCandidates_UsesRarityLookup(t *testing.T) {
	candidates := convertToCardCandidates([]string{
		"Witch",
		"The Log",
		"Mini P.E.K.K.A",
		"Archer Queen",
		"Skeletons",
	})

	rarityByName := make(map[string]string, len(candidates))
	for _, candidate := range candidates {
		rarityByName[candidate.Name] = candidate.Rarity
	}

	tests := []struct {
		name string
		want string
	}{
		{name: "Witch", want: "Epic"},
		{name: "The Log", want: "Legendary"},
		{name: "Mini P.E.K.K.A", want: "Rare"},
		{name: "Archer Queen", want: "Champion"},
		{name: "Skeletons", want: "Common"},
	}

	for _, tt := range tests {
		got := rarityByName[tt.name]
		if got != tt.want {
			t.Fatalf("%s rarity = %q, want %q", tt.name, got, tt.want)
		}
	}
}
