package deckhash

import (
	"reflect"
	"testing"
)

func TestComputeOrderIndependent(t *testing.T) {
	cardsA := []string{"hog rider", "musketeer", "fireball", "log"}
	cardsB := []string{"log", "fireball", "musketeer", "hog rider"}

	hashA := Compute(cardsA)
	hashB := Compute(cardsB)

	if hashA != hashB {
		t.Fatalf("expected equal hashes for reordered cards, got %q and %q", hashA, hashB)
	}
}

func TestComputeDifferentDecks(t *testing.T) {
	cardsA := []string{"hog rider", "musketeer", "fireball", "log"}
	cardsB := []string{"hog rider", "musketeer", "earthquake", "log"}

	hashA := Compute(cardsA)
	hashB := Compute(cardsB)

	if hashA == hashB {
		t.Fatalf("expected different hashes for different cards, both %q", hashA)
	}
}

func TestComputeMatchesCanonicalDeckHash(t *testing.T) {
	cards := []string{"a", "b|c"}

	got := Compute(cards)
	want := DeckHash(cards)
	if got != want {
		t.Fatalf("expected Compute to match DeckHash, got %q want %q", got, want)
	}
}

func TestLegacyComputePreservedForMigration(t *testing.T) {
	cardsA := []string{"a", "b|c"}
	cardsB := []string{"a|b", "c"}

	hashA := LegacyCompute(cardsA)
	hashB := LegacyCompute(cardsB)
	if hashA != hashB {
		t.Fatalf("expected legacy hash collision for delimiter-ambiguous inputs")
	}

	if !IsLegacyHash(cardsA, hashA) {
		t.Fatalf("expected IsLegacyHash to match LegacyCompute output")
	}

	if DeckHash(cardsA) == DeckHash(cardsB) {
		t.Fatalf("expected canonical DeckHash to avoid delimiter collision")
	}
}

func TestDeckHashDoesNotMutateInput(t *testing.T) {
	cards := []string{"log", "fireball", "musketeer", "hog rider"}
	original := append([]string(nil), cards...)

	_ = DeckHash(cards)

	if !reflect.DeepEqual(cards, original) {
		t.Fatalf("expected DeckHash to leave input unchanged, got %v want %v", cards, original)
	}
}
