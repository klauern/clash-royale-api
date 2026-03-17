package deckhash

import "testing"

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
