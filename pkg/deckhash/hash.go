package deckhash

import (
	"strings"
)

// Compute returns a deterministic hash for a deck card list, independent of card order.
func Compute(cards []string) string {
	return DeckHash(cards)
}

// LegacyCompute returns the historical deck hash based on a pipe-joined sorted card list.
//
// This is retained only for migration and compatibility checks.
func LegacyCompute(cards []string) string {
	data := strings.Join(sortedCards(cards), "|")
	return sha256HexString([]byte(data))
}

// IsLegacyHash reports whether hash matches the historical Compute semantics.
func IsLegacyHash(cards []string, hash string) bool {
	return LegacyCompute(cards) == strings.ToLower(hash)
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
