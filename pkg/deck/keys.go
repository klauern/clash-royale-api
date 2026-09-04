package deck

import "github.com/klauer/clash-royale-api/go/pkg/deckhash"

// CanonicalDeckKey returns a deterministic key for a deck independent of card order.
func CanonicalDeckKey(cards []string) string {
	return deckhash.CanonicalDeckKey(cards)
}
