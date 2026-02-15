package events

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"sort"
	"strings"
)

func deckHash(cardNames []string) string {
	normalized := normalizeDeckNames(cardNames)
	if len(normalized) == 0 {
		return ""
	}

	sum := sha256.Sum256([]byte(strings.Join(normalized, "|")))
	return hex.EncodeToString(sum[:])[:12]
}

func normalizeDeckNames(cardNames []string) []string {
	normalized := make([]string, 0, len(cardNames))
	for _, name := range cardNames {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, strings.ToLower(trimmed))
	}
	sort.Strings(normalized)
	return normalized
}

func inferDeckArchetype(cardNames []string) string {
	names := normalizeDeckNames(cardNames)
	if len(names) == 0 {
		return ""
	}

	contains := func(target string) bool {
		return slices.Contains(names, target)
	}

	rules := []struct {
		archetype string
		cards     []string
	}{
		{archetype: "siege", cards: []string{"x-bow", "mortar"}},
		{archetype: "bait", cards: []string{"goblin barrel", "princess"}},
		{archetype: "beatdown", cards: []string{"golem", "lava hound", "giant"}},
		{archetype: "cycle", cards: []string{"hog rider"}},
		{archetype: "control", cards: []string{"graveyard", "miner"}},
	}

	for _, rule := range rules {
		if slices.ContainsFunc(rule.cards, contains) {
			return rule.archetype
		}
	}
	return ""
}
