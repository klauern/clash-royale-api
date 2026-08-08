package config

// EvolutionProgressBonus calculates a weighted bonus from evolution progress.
// Invalid or missing progress returns zero, and progress above the maximum is
// capped so the result never exceeds weight.
func EvolutionProgressBonus(evolutionLevel, maxEvolutionLevel int, weight float64) float64 {
	if evolutionLevel <= 0 || maxEvolutionLevel <= 0 {
		return 0
	}

	progress := min(float64(evolutionLevel)/float64(maxEvolutionLevel), 1)
	return weight * progress
}
