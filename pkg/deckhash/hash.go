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
