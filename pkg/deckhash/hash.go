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
