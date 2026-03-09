package deckhash

import "testing"

func TestDeckHashOrderInvariant(t *testing.T) {
	cardsA := []string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Ice Spirit"}
	cardsB := []string{"Ice Spirit", "Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang"}
	cardsC := []string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Skeleton Army"}

	hashA := DeckHash(cardsA)
	hashB := DeckHash(cardsB)
	hashC := DeckHash(cardsC)

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

func TestDeckHashAvoidsDelimiterCollision(t *testing.T) {
	cardsA := []string{"a", "b|c"}
	cardsB := []string{"a|b", "c"}

	hashA := DeckHash(cardsA)
	hashB := DeckHash(cardsB)
	if hashA == hashB {
		t.Fatalf("expected different hashes for distinct inputs, got %q", hashA)
	}
}
