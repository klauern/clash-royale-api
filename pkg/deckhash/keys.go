package deckhash

import (
	"strconv"
	"strings"
)

// CanonicalDeckKey returns a deterministic key for a deck independent of card order.
func CanonicalDeckKey(cards []string) string {
	sorted := sortedCards(cards)

	var b strings.Builder
	for _, card := range sorted {
		// Length-prefix each card to avoid delimiter-collision ambiguity.
		b.WriteString(strconv.Itoa(len(card)))
		b.WriteByte(':')
		b.WriteString(card)
		b.WriteByte('|')
	}

	return b.String()
}

// DeckHash returns a SHA256 hash of the canonical deck key for deduplication.
func DeckHash(cards []string) string {
	key := CanonicalDeckKey(cards)
	return sha256HexString([]byte(key))
}
