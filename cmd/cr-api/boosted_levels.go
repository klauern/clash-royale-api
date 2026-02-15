package main

import (
	"fmt"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
)

func parseBoostedCardLevels(entries []string) (map[string]int, error) {
	overrides := make(map[string]int)
	for _, entry := range entries {
		name, level, err := parseBoostedCardLevelEntry(entry)
		if err != nil {
			return nil, err
		}
		overrides[name] = level
	}
	return overrides, nil
}

func parseBoostedCardLevelEntry(entry string) (string, int, error) {
	raw := strings.TrimSpace(entry)
	if raw == "" {
		return "", 0, fmt.Errorf("invalid --boosted-card-level value: empty string")
	}

	idx := strings.LastIndex(raw, ":")
	if idx <= 0 || idx == len(raw)-1 {
		return "", 0, fmt.Errorf("invalid --boosted-card-level value %q (expected CardName:Level)", entry)
	}

	cardName := strings.TrimSpace(raw[:idx])
	levelPart := strings.TrimSpace(raw[idx+1:])
	if cardName == "" {
		return "", 0, fmt.Errorf("invalid --boosted-card-level value %q (missing card name)", entry)
	}

	var level int
	if _, err := fmt.Sscanf(levelPart, "%d", &level); err != nil {
		return "", 0, fmt.Errorf("invalid --boosted-card-level value %q (level must be integer)", entry)
	}
	if level < 1 || level > 16 {
		return "", 0, fmt.Errorf("invalid --boosted-card-level value %q (level must be 1-16)", entry)
	}

	return cardName, level, nil
}

func applyBoostedLevelsToCardAnalysis(analysis *deck.CardAnalysis, overrides map[string]int) {
	if analysis == nil || len(overrides) == 0 {
		return
	}
	for cardName, boostedLevel := range overrides {
		matchedKey, ok := matchCardKey(analysis.CardLevels, cardName)
		if !ok {
			continue
		}
		card := analysis.CardLevels[matchedKey]
		card.Level = clampBoostLevel(boostedLevel, card.MaxLevel)
		analysis.CardLevels[matchedKey] = card
	}
}

func applyBoostedLevelsToPlayerContext(ctx *evaluation.PlayerContext, overrides map[string]int) {
	if ctx == nil || len(overrides) == 0 {
		return
	}
	for cardName, boostedLevel := range overrides {
		matchedKey, ok := matchCardKey(ctx.Collection, cardName)
		if !ok {
			continue
		}
		info := ctx.Collection[matchedKey]
		info.Level = clampBoostLevel(boostedLevel, info.MaxLevel)
		ctx.Collection[matchedKey] = info
	}
}

func applyBoostedLevelsToDeckCandidates(cards []deck.CardCandidate, overrides map[string]int) {
	if len(cards) == 0 || len(overrides) == 0 {
		return
	}
	for i := range cards {
		overrideLevel, ok := findOverrideLevel(overrides, cards[i].Name)
		if !ok {
			continue
		}
		cards[i].Level = clampBoostLevel(overrideLevel, cards[i].MaxLevel)
	}
}

func findOverrideLevel(overrides map[string]int, name string) (int, bool) {
	for k, level := range overrides {
		if strings.EqualFold(strings.TrimSpace(k), strings.TrimSpace(name)) {
			return level, true
		}
	}
	return 0, false
}

func matchCardKey[T any](m map[string]T, target string) (string, bool) {
	for key := range m {
		if strings.EqualFold(strings.TrimSpace(key), strings.TrimSpace(target)) {
			return key, true
		}
	}
	return "", false
}

func clampBoostLevel(level, maxLevel int) int {
	if maxLevel > 0 && level > maxLevel {
		return maxLevel
	}
	return level
}
