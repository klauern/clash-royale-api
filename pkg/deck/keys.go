package deck

import "github.com/klauer/clash-royale-api/go/pkg/deckhash"

// CanonicalDeckKey returns a deterministic key for a deck independent of card order.
func CanonicalDeckKey(cards []string) string {
	return deckhash.CanonicalDeckKey(cards)
}

// DeckHash returns a SHA256 hash of the canonical deck key for deduplication.
func DeckHash(cards []string) string {
	return deckhash.DeckHash(cards)
}
