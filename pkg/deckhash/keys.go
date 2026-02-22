package deckhash

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

// CanonicalDeckKey returns a deterministic key for a deck independent of card order.
func CanonicalDeckKey(cards []string) string {
	sorted := make([]string, len(cards))
	copy(sorted, cards)
	sort.Strings(sorted)
	return strings.Join(sorted, "|")
}

// DeckHash returns a SHA256 hash of the canonical deck key for deduplication.
func DeckHash(cards []string) string {
	key := CanonicalDeckKey(cards)
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash)
}
