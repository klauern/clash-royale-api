package deck

import "testing"

func TestCanonicalDeckKey_OrderInsensitive(t *testing.T) {
	cards1 := []string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Ice Spirit"}
	cards2 := []string{"Ice Spirit", "Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang"}

	key1 := CanonicalDeckKey(cards1)
	key2 := CanonicalDeckKey(cards2)

	if key1 != key2 {
		t.Fatalf("expected same key for same cards in different order")
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
		t.Errorf("expected same hash for same cards in different order")
	}
	if hash1 == hash3 {
		t.Errorf("expected different hash for different cards")
	}
	if len(hash1) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash1))
	}
}
