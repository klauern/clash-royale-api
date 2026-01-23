// Package genetic provides genetic algorithm-based deck optimization
// using the eaopt library for evolutionary deck generation.
package genetic

import (
	"sort"
	"strings"
	"sync"
)

var fitnessCache sync.Map

func fitnessCacheKey(cards []string) string {
	if len(cards) == 0 {
		return ""
	}
	sorted := make([]string, len(cards))
	copy(sorted, cards)
	sort.Strings(sorted)
	return strings.Join(sorted, "|")
}

func getCachedFitness(cards []string) (float64, bool) {
	key := fitnessCacheKey(cards)
	if key == "" {
		return 0, false
	}
	if value, ok := fitnessCache.Load(key); ok {
		if fitness, ok := value.(float64); ok {
			return fitness, true
		}
	}
	return 0, false
}

func storeCachedFitness(cards []string, fitness float64) {
	key := fitnessCacheKey(cards)
	if key == "" {
		return
	}
	fitnessCache.Store(key, fitness)
}
