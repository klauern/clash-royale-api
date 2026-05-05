package analysis

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadDeckFromFile_EnrichesMissingMetadata reproduces the bug filed under
// clash-royale-api-eb6: deck files written by pkg/deck.DeckRecommendation
// have no deck_name/win_condition/strategy, leaving those fields blank in
// playstyle's "RECOMMENDED DECK" output.
func TestLoadDeckFromFile_EnrichesMissingMetadata(t *testing.T) {
	raw := `{
		"deck": ["Ram Rider","Goblin Hut","Poison","Vines","Princess","Ice Wizard","Ice Golem","Bomber"],
		"deck_detail": [
			{"name":"Ram Rider","level":10,"max_level":16,"rarity":"Legendary","elixir":5,"role":"win_conditions","score":1.5},
			{"name":"Goblin Hut","level":11,"max_level":16,"rarity":"Rare","elixir":4,"role":"buildings","score":1.1},
			{"name":"Poison","level":10,"max_level":16,"rarity":"Epic","elixir":4,"role":"spells_big","score":1.4},
			{"name":"Vines","level":11,"max_level":16,"rarity":"Epic","elixir":3,"role":"spells_small","score":1.5},
			{"name":"Princess","level":12,"max_level":16,"rarity":"Legendary","elixir":3,"role":"support","score":2.0},
			{"name":"Ice Wizard","level":10,"max_level":16,"rarity":"Legendary","elixir":3,"role":"support","score":1.7},
			{"name":"Ice Golem","level":9,"max_level":16,"rarity":"Rare","elixir":2,"role":"cycle","score":0.8},
			{"name":"Bomber","level":11,"max_level":16,"rarity":"Common","elixir":2,"role":"cycle","score":0.8}
		],
		"average_elixir": 3.25,
		"analysis_time": "2026-01-18T19:44:06-06:00",
		"notes": []
	}`

	dir := t.TempDir()
	path := filepath.Join(dir, "20260118_194406_deck_TEST.json")
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatalf("setup write: %v", err)
	}

	deck, err := loadDeckFromFile(path)
	if err != nil {
		t.Fatalf("loadDeckFromFile: %v", err)
	}

	if deck.WinCondition != "Ram Rider" {
		t.Errorf("WinCondition: want %q, got %q", "Ram Rider", deck.WinCondition)
	}
	if deck.DeckName == "" {
		t.Error("DeckName should not be empty after enrichment")
	}
	if deck.Strategy == "" {
		t.Error("Strategy should not be empty after enrichment")
	}
	if len(deck.DeckDetail) != 8 {
		t.Errorf("DeckDetail: want 8 cards, got %d", len(deck.DeckDetail))
	}
	if len(deck.Cards) != 8 {
		t.Errorf("Cards: expected hydrated from DeckDetail (8), got %d", len(deck.Cards))
	}
	if got := deck.AverageElixir; got != 3.25 {
		t.Errorf("AverageElixir: want 3.25, got %v", got)
	}
}

func TestEnrichDeckMetadata_PreservesExistingFields(t *testing.T) {
	deck := &DeckAnalysis{
		DeckName:      "Custom Name",
		WinCondition:  "Hog Rider",
		Strategy:      "Custom strategy text",
		AverageElixir: 2.9,
	}
	enrichDeckMetadata(deck, nil, []byte(`{}`))
	if deck.DeckName != "Custom Name" {
		t.Errorf("DeckName overwritten: %q", deck.DeckName)
	}
	if deck.WinCondition != "Hog Rider" {
		t.Errorf("WinCondition overwritten: %q", deck.WinCondition)
	}
	if deck.Strategy != "Custom strategy text" {
		t.Errorf("Strategy overwritten: %q", deck.Strategy)
	}
}

func TestClassifyDeckStyle(t *testing.T) {
	tests := []struct {
		avg  float64
		want string
	}{
		{2.5, "Cycle"},
		{3.2, "Control"},
		{3.8, "Midrange"},
		{4.5, "Beatdown"},
	}
	for _, tt := range tests {
		if got := classifyDeckStyle(tt.avg); got != tt.want {
			t.Errorf("classifyDeckStyle(%v): want %q, got %q", tt.avg, tt.want, got)
		}
	}
}
