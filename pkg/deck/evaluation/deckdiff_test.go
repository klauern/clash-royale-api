package evaluation

import (
	"sort"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func testPlayer(cards []clashroyale.Card) *clashroyale.Player {
	return &clashroyale.Player{
		Tag:  "#TEST",
		Name: "TestPlayer",
		Arena: clashroyale.Arena{
			ID:   13,
			Name: "Legendary Arena",
		},
		Cards: cards,
	}
}

func makeDiffCard(name string, level, maxLevel int) clashroyale.Card {
	return clashroyale.Card{
		Name:     name,
		Level:    level,
		MaxLevel: maxLevel,
		Rarity:   "Common",
	}
}

func makeDiffCandidate(name string, level int) deck.CardCandidate {
	return deck.CardCandidate{
		Name:     name,
		Level:    level,
		MaxLevel: 14,
		Rarity:   "Common",
		Elixir:   3,
	}
}

func TestDiffDecks_ScoreDeltaConsistency(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	playerCards := []clashroyale.Card{
		makeDiffCard("Giant", 11, 14),
		makeDiffCard("Musketeer", 11, 14),
		makeDiffCard("Knight", 11, 14),
		makeDiffCard("Fireball", 10, 14),
		makeDiffCard("The Log", 11, 14),
		makeDiffCard("Valkyrie", 10, 14),
		makeDiffCard("Mega Minion", 11, 14),
		makeDiffCard("Miner", 9, 14),
		makeDiffCard("Hog Rider", 12, 14),
		makeDiffCard("Ice Golem", 11, 14),
	}
	player := testPlayer(playerCards)

	currentCards := []deck.CardCandidate{
		makeDiffCandidate("Giant", 11),
		makeDiffCandidate("Musketeer", 11),
		makeDiffCandidate("Knight", 11),
		makeDiffCandidate("Fireball", 10),
		makeDiffCandidate("The Log", 11),
		makeDiffCandidate("Valkyrie", 10),
		makeDiffCandidate("Mega Minion", 11),
		makeDiffCandidate("Miner", 9),
	}

	recommendedCards := []deck.CardCandidate{
		makeDiffCandidate("Hog Rider", 12),
		makeDiffCandidate("Musketeer", 11),
		makeDiffCandidate("Knight", 11),
		makeDiffCandidate("Fireball", 10),
		makeDiffCandidate("The Log", 11),
		makeDiffCandidate("Ice Golem", 11),
		makeDiffCandidate("Mega Minion", 11),
		makeDiffCandidate("Miner", 9),
	}

	diff, err := DiffDecks(player, currentCards, recommendedCards, synergyDB)
	if err != nil {
		t.Fatalf("DiffDecks returned error: %v", err)
	}

	// ScoreDelta must equal RecommendedScore - CurrentScore
	want := diff.RecommendedScore - diff.CurrentScore
	if diff.ScoreDelta != want {
		t.Errorf("ScoreDelta = %.6f, want %.6f", diff.ScoreDelta, want)
	}

	// All 6 category deltas must be present
	wantKeys := []string{categoryAttack, categoryDefense, categorySynergy, categoryVersatility, categoryF2P, categoryPlayability}
	for _, key := range wantKeys {
		cd, ok := diff.CategoryDeltas[key]
		if !ok {
			t.Errorf("missing category delta: %s", key)
			continue
		}
		wantDelta := cd.Recommended - cd.Current
		if cd.Delta != wantDelta {
			t.Errorf("category %s: Delta = %.6f, want %.6f", key, cd.Delta, wantDelta)
		}
	}
}

func TestDiffDecks_SharedAndReplacedCards(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()
	player := testPlayer([]clashroyale.Card{
		makeDiffCard("Giant", 11, 14),
		makeDiffCard("Musketeer", 11, 14),
		makeDiffCard("Knight", 11, 14),
		makeDiffCard("Fireball", 10, 14),
		makeDiffCard("The Log", 11, 14),
		makeDiffCard("Valkyrie", 10, 14),
		makeDiffCard("Mega Minion", 11, 14),
		makeDiffCard("Miner", 9, 14),
		makeDiffCard("Hog Rider", 12, 14),
		makeDiffCard("Ice Golem", 11, 14),
	})

	current := []deck.CardCandidate{
		makeDiffCandidate("Giant", 11),
		makeDiffCandidate("Musketeer", 11),
		makeDiffCandidate("Knight", 11),
		makeDiffCandidate("Fireball", 10),
		makeDiffCandidate("The Log", 11),
		makeDiffCandidate("Valkyrie", 10),
		makeDiffCandidate("Mega Minion", 11),
		makeDiffCandidate("Miner", 9),
	}

	// Replace Giant+Valkyrie with Hog Rider+Ice Golem; keep the other 6.
	recommended := []deck.CardCandidate{
		makeDiffCandidate("Hog Rider", 12),
		makeDiffCandidate("Musketeer", 11),
		makeDiffCandidate("Knight", 11),
		makeDiffCandidate("Fireball", 10),
		makeDiffCandidate("The Log", 11),
		makeDiffCandidate("Ice Golem", 11),
		makeDiffCandidate("Mega Minion", 11),
		makeDiffCandidate("Miner", 9),
	}

	diff, err := DiffDecks(player, current, recommended, synergyDB)
	if err != nil {
		t.Fatalf("DiffDecks returned error: %v", err)
	}

	wantShared := []string{"Fireball", "Knight", "Mega Minion", "Miner", "Musketeer", "The Log"}
	wantReplaced := []string{"Hog Rider", "Ice Golem"}

	sort.Strings(diff.SharedCards)
	sort.Strings(diff.ReplacedCards)

	if len(diff.SharedCards) != len(wantShared) {
		t.Errorf("SharedCards = %v, want %v", diff.SharedCards, wantShared)
	} else {
		for i, name := range wantShared {
			if diff.SharedCards[i] != name {
				t.Errorf("SharedCards[%d] = %q, want %q", i, diff.SharedCards[i], name)
			}
		}
	}

	if len(diff.ReplacedCards) != len(wantReplaced) {
		t.Errorf("ReplacedCards = %v, want %v", diff.ReplacedCards, wantReplaced)
	} else {
		for i, name := range wantReplaced {
			if diff.ReplacedCards[i] != name {
				t.Errorf("ReplacedCards[%d] = %q, want %q", i, diff.ReplacedCards[i], name)
			}
		}
	}
}

func TestDiffDecks_IdenticalDecksZeroDelta(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()
	player := testPlayer([]clashroyale.Card{
		makeDiffCard("Hog Rider", 12, 14),
		makeDiffCard("Musketeer", 11, 14),
		makeDiffCard("Knight", 11, 14),
		makeDiffCard("Fireball", 10, 14),
		makeDiffCard("The Log", 11, 14),
		makeDiffCard("Ice Golem", 11, 14),
		makeDiffCard("Mega Minion", 11, 14),
		makeDiffCard("Miner", 9, 14),
	})

	cards := []deck.CardCandidate{
		makeDiffCandidate("Hog Rider", 12),
		makeDiffCandidate("Musketeer", 11),
		makeDiffCandidate("Knight", 11),
		makeDiffCandidate("Fireball", 10),
		makeDiffCandidate("The Log", 11),
		makeDiffCandidate("Ice Golem", 11),
		makeDiffCandidate("Mega Minion", 11),
		makeDiffCandidate("Miner", 9),
	}

	diff, err := DiffDecks(player, cards, cards, synergyDB)
	if err != nil {
		t.Fatalf("DiffDecks returned error: %v", err)
	}

	if diff.ScoreDelta != 0 {
		t.Errorf("identical decks: ScoreDelta = %.4f, want 0", diff.ScoreDelta)
	}
	if len(diff.SharedCards) != 8 {
		t.Errorf("identical decks: SharedCards len = %d, want 8", len(diff.SharedCards))
	}
	if len(diff.ReplacedCards) != 0 {
		t.Errorf("identical decks: ReplacedCards = %v, want empty", diff.ReplacedCards)
	}
}

func TestDiffDecks_LevelFitPresent(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()
	player := testPlayer([]clashroyale.Card{
		makeDiffCard("Giant", 11, 14),
		makeDiffCard("Musketeer", 11, 14),
		makeDiffCard("Knight", 11, 14),
		makeDiffCard("Fireball", 10, 14),
		makeDiffCard("The Log", 11, 14),
		makeDiffCard("Valkyrie", 10, 14),
		makeDiffCard("Mega Minion", 11, 14),
		makeDiffCard("Miner", 9, 14),
	})

	cards := []deck.CardCandidate{
		makeDiffCandidate("Giant", 11),
		makeDiffCandidate("Musketeer", 11),
		makeDiffCandidate("Knight", 11),
		makeDiffCandidate("Fireball", 10),
		makeDiffCandidate("The Log", 11),
		makeDiffCandidate("Valkyrie", 10),
		makeDiffCandidate("Mega Minion", 11),
		makeDiffCandidate("Miner", 9),
	}

	diff, err := DiffDecks(player, cards, cards, synergyDB)
	if err != nil {
		t.Fatalf("DiffDecks returned error: %v", err)
	}

	// Level fit must be non-negative (either from OverallBreakdown or GetAverageLevel)
	if diff.CurrentLevelFit < 0 {
		t.Errorf("CurrentLevelFit = %.4f, want >= 0", diff.CurrentLevelFit)
	}
	if diff.RecommendedLevelFit < 0 {
		t.Errorf("RecommendedLevelFit = %.4f, want >= 0", diff.RecommendedLevelFit)
	}
}
