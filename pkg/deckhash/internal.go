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

func sha256HexString(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
