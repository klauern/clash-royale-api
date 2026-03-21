package deckhash

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Compute returns the canonical deck hash for deduplication.
//
// Deprecated: prefer DeckHash directly in new call sites.
func Compute(cards []string) string {
	return DeckHash(cards)
}

// LegacyCompute returns the historical deck hash based on a pipe-joined sorted card list.
//
// This is retained only for migration and compatibility checks.
func LegacyCompute(cards []string) string {
	sorted := sortedCardsCopy(cards)
	sum := sha256.Sum256([]byte(strings.Join(sorted, "|")))
	return hex.EncodeToString(sum[:])
}

// IsLegacyHash reports whether hash matches the historical Compute semantics.
func IsLegacyHash(cards []string, hash string) bool {
	return legacyPipeJoinedHash(cards) == strings.ToLower(hash)
}

func legacyPipeJoinedHash(cards []string) string {
	return LegacyCompute(cards)
}
