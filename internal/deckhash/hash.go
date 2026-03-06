package deckhash

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

// FromCards computes a deterministic SHA256 hash for a deck card list.
// Card order does not affect the output.
func FromCards(cards []string) string {
	sorted := make([]string, len(cards))
	copy(sorted, cards)
	sort.Strings(sorted)

	sum := sha256.Sum256([]byte(strings.Join(sorted, "|")))
	return fmt.Sprintf("%x", sum)
}
