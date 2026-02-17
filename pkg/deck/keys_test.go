package deck

import "testing"

func TestCanonicalDeckKey_OrderInsensitive(t *testing.T) {
	cards1 := []string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Ice Spirit"}
	cards2 := []string{"Ice Spirit", "Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang"}
	cards3 := []string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Skeleton Army"}

	key1 := CanonicalDeckKey(cards1)
	key2 := CanonicalDeckKey(cards2)
	key3 := CanonicalDeckKey(cards3)

	if key1 != key2 {
		t.Fatalf("expected same key for same cards in different order")
	}
	if key1 == key3 {
		t.Fatalf("expected different key for different cards")
	}
}

func TestDeckHash_OrderInsensitive(t *testing.T) {
	cards1 := []string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Ice Spirit"}
	cards2 := []string{"Ice Spirit", "Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang"}
	cards3 := []string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Skeleton Army"}

	hash1 := DeckHash(cards1)
	hash2 := DeckHash(cards2)
	hash3 := DeckHash(cards3)

	if hash1 != hash2 {
		t.Fatalf("expected same hash for same cards in different order")
	}
	if hash1 == hash3 {
		t.Fatalf("expected different hash for different cards")
	}
	if len(hash1) != 64 {
		t.Fatalf("expected hash length 64, got %d", len(hash1))
	}
}
