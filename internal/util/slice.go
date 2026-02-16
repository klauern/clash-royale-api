package util

// FilterSlice returns a new slice containing only elements that satisfy the predicate.
// This consolidates duplicate filtering patterns across the codebase.
func FilterSlice[T any](slice []T, predicate func(T) bool) []T {
	var filtered []T
	for _, item := range slice {
		if predicate(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// CalcAvgElixir calculates the average elixir cost from a slice of items.
// The getElixir function extracts the elixir cost from each item.
// Returns 0 if the slice is empty.
func CalcAvgElixir[T any](items []T, getElixir func(T) int) float64 {
	if len(items) == 0 {
		return 0
	}

	total := 0
	for _, item := range items {
		total += getElixir(item)
	}

	return float64(total) / float64(len(items))
}
