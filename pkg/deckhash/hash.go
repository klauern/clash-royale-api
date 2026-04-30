package deckhash

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

// Compute returns a deterministic hash for a deck card list, independent of card order.
func Compute(cards []string) string {
	sorted := make([]string, len(cards))
	copy(sorted, cards)
	sort.Strings(sorted)

	data := strings.Join(sorted, "|")
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// ComputeCanonical normalizes card names (trim/lowercase), sorts, and hashes.
func ComputeCanonical(cards []string) string {
	normalized := make([]string, 0, len(cards))
	for _, card := range cards {
		name := strings.ToLower(strings.TrimSpace(card))
		if name == "" {
			continue
		}
		normalized = append(normalized, name)
	}
	if len(normalized) == 0 {
		return ""
	}

	sort.Strings(normalized)
	data := strings.Join(normalized, "|")
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// ComputeCanonicalShort returns a deterministic short hash for normalized decks.
func ComputeCanonicalShort(cards []string, length int) string {
	full := ComputeCanonical(cards)
	if full == "" {
		return ""
	}
	if length <= 0 || length >= len(full) {
		return full
	}
	return full[:length]
}
