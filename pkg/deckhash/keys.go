package deckhash

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

// CanonicalDeckKey returns a deterministic key for a deck independent of card order.
func CanonicalDeckKey(cards []string) string {
	sorted := make([]string, len(cards))
	copy(sorted, cards)
	sort.Strings(sorted)

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
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
