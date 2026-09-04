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
