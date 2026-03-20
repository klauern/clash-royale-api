package deckhash

import (
	"strings"
)

// Compute returns a deterministic hash for a deck card list, independent of card order.
func Compute(cards []string) string {
	sorted := sortedCards(cards)

	data := strings.Join(sorted, "|")
	return sha256HexString(data)
}
