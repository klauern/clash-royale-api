package deckhash

import "testing"

func TestFromCards(t *testing.T) {
	cardsA := []string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Ice Spirit"}
	cardsB := []string{"Ice Spirit", "Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang"}
	cardsC := []string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Skeleton Army"}

	hashA := FromCards(cardsA)
	hashB := FromCards(cardsB)
	hashC := FromCards(cardsC)

	if hashA != hashB {
		t.Fatalf("expected same hash for equivalent decks, got %q and %q", hashA, hashB)
	}

	if hashA == hashC {
		t.Fatalf("expected different hash for different decks, got %q", hashA)
	}

	if len(hashA) != 64 {
		t.Fatalf("expected SHA256 hex length 64, got %d", len(hashA))
	}
}
