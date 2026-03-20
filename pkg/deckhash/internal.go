package deckhash

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

func sortedCards(cards []string) []string {
	sorted := make([]string, len(cards))
	copy(sorted, cards)
	sort.Strings(sorted)
	return sorted
}

func sha256HexString(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}
