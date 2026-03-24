package deckhash

import (
	"strings"
)

// Compute returns a deterministic hash for a deck card list, independent of card order.
func Compute(cards []string) string {
	data := strings.Join(sortedCards(cards), "|")
	return sha256HexString([]byte(data))
}
