package analysis

import "testing"

func TestUpdateCardCounts(t *testing.T) {
	// Test with sample card data
	cards := []CardInfo{
		NewCardAdapter("Common"),
		NewCardAdapter("Common"),
		NewCardAdapter("Rare"),
		NewCardAdapter("Rare"),
		NewCardAdapter("Epic"),
		NewCardAdapter("Legendary"),
	}

	// Update counts
	UpdateCardCounts(cards)

	// Check counts
	if got := GetTotalCardsByRarity("Common"); got != 2 {
		t.Errorf("Common cards = %d, want 2", got)
	}

	if got := GetTotalCardsByRarity("Rare"); got != 2 {
		t.Errorf("Rare cards = %d, want 2", got)
	}

	if got := GetTotalCardsByRarity("Epic"); got != 1 {
		t.Errorf("Epic cards = %d, want 1", got)
	}

	if got := GetTotalCardsByRarity("Legendary"); got != 1 {
		t.Errorf("Legendary cards = %d, want 1", got)
	}

	// Champion should be fallback (0 provided in test, should use default)
	if got := GetTotalCardsByRarity("Champion"); got != 6 {
		t.Errorf("Champion cards = %d, want 6 (fallback)", got)
	}
}

func TestNewCardAdapter(t *testing.T) {
	card := NewCardAdapter("Epic")
	if got := card.GetRarity(); got != "Epic" {
		t.Errorf("Card rarity = %s, want Epic", got)
	}
}