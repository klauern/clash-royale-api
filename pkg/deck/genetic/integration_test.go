//go:build integration

package genetic

import (
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// TestGeneticAlgorithmEndToEnd tests the full genetic algorithm workflow
// with a realistic candidate pool and configuration
func TestGeneticAlgorithmEndToEnd(t *testing.T) {
	// Create realistic candidate pool (25+ cards with diverse roles)
	candidates := createRealisticCandidates()

	config := GeneticConfig{
		PopulationSize:         20,
		Generations:            15,
		MutationRate:           0.2,
		CrossoverRate:          0.7,
		MutationIntensity:      0.3,
		EliteCount:             3,
		TournamentSize:         5,
		ParallelEvaluations:    false,
		ConvergenceGenerations: 0,
		TargetFitness:          0,
	}

	optimizer, err := NewGeneticOptimizer(candidates, deck.StrategyBalanced, &config)
	if err != nil {
		t.Fatalf("NewGeneticOptimizer() failed: %v", err)
	}

	// Track progress
	progressCalls := 0
	var lastProgress GeneticProgress
	optimizer.Progress = func(p GeneticProgress) {
		progressCalls++
		lastProgress = p
	}

	start := time.Now()
	result, err := optimizer.Optimize()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Optimize() error = %v", err)
	}

	// Validate result
	if result == nil {
		t.Fatal("Optimize() returned nil result")
	}

	// Should have hall of fame
	if len(result.HallOfFame) == 0 {
		t.Error("Optimize() produced empty hall of fame")
	}

	if len(result.Scores) == 0 {
		t.Error("Optimize() produced empty scores")
	}

	if len(result.HallOfFame) != len(result.Scores) {
		t.Errorf("HallOfFame length %d != Scores length %d", len(result.HallOfFame), len(result.Scores))
	}

	// Validate each deck in hall of fame
	for i, genome := range result.HallOfFame {
		if genome == nil {
			t.Errorf("HallOfFame[%d] is nil", i)
			continue
		}

		if len(genome.Cards) != 8 {
			t.Errorf("HallOfFame[%d] has %d cards, want 8", i, len(genome.Cards))
		}

		// Check unique cards
		cardSet := make(map[string]bool)
		for _, card := range genome.Cards {
			if cardSet[card] {
				t.Errorf("HallOfFame[%d] has duplicate card: %v", i, card)
			}
			cardSet[card] = true
		}
	}

	// Validate fitness scores
	for i, score := range result.Scores {
		if score < 0 || score > 10 {
			t.Errorf("Scores[%d] = %v, want in range [0, 10]", i, score)
		}
	}

	// Scores should be sorted (best first)
	for i := 1; i < len(result.Scores); i++ {
		if result.Scores[i] > result.Scores[i-1] {
			t.Errorf("Scores not sorted: Scores[%d] = %v > Scores[%d] = %v",
				i, result.Scores[i], i-1, result.Scores[i-1])
		}
	}

	// Validate generations
	if result.Generations == 0 {
		t.Error("Optimize() completed with 0 generations")
	}

	if result.Generations > uint(config.Generations) {
		t.Errorf("Generations %d > configured max %d", result.Generations, config.Generations)
	}

	// Validate duration
	if result.Duration == 0 {
		t.Error("Optimize() has zero duration")
	}

	// Progress callback should have been called
	if progressCalls == 0 {
		t.Error("Progress callback never called")
	}

	if lastProgress.Generation == 0 {
		t.Error("Last progress has generation 0")
	}

	// Performance check (should complete in reasonable time)
	if duration > 60*time.Second {
		t.Errorf("Optimization took %v, expected < 60s for small GA run", duration)
	}

	t.Logf("GA completed in %v: %d generations, best fitness %.2f",
		duration, result.Generations, result.Scores[0])
}

// TestGeneticAlgorithmConvergence tests that GA improves solutions over time
func TestGeneticAlgorithmConvergence(t *testing.T) {
	candidates := createRealisticCandidates()

	config := GeneticConfig{
		PopulationSize:    30,
		Generations:       25,
		MutationRate:      0.2,
		CrossoverRate:     0.7,
		MutationIntensity: 0.3,
		EliteCount:        5,
		TournamentSize:    5,
	}

	optimizer, err := NewGeneticOptimizer(candidates, deck.StrategyBalanced, &config)
	if err != nil {
		t.Fatalf("NewGeneticOptimizer() failed: %v", err)
	}

	// Track fitness over time
	var fitnessHistory []float64
	optimizer.Progress = func(p GeneticProgress) {
		fitnessHistory = append(fitnessHistory, p.BestFitness)
	}

	result, err := optimizer.Optimize()
	if err != nil {
		t.Fatalf("Optimize() error = %v", err)
	}

	// Should have tracked fitness
	if len(fitnessHistory) == 0 {
		t.Fatal("No fitness history recorded")
	}

	// Final best fitness should be at least as good as initial
	initialBest := fitnessHistory[0]
	finalBest := result.Scores[0]

	if finalBest < initialBest-0.5 { // Allow small regression due to randomness
		t.Errorf("GA regressed: final fitness %.2f < initial %.2f", finalBest, initialBest)
	}

	t.Logf("Fitness progression: initial=%.2f, final=%.2f (improvement: %.2f)",
		initialBest, finalBest, finalBest-initialBest)
}

// TestGeneticAlgorithmWithSeedPopulation tests initialization with seed decks
func TestGeneticAlgorithmWithSeedPopulation(t *testing.T) {
	candidates := createRealisticCandidates()

	// Create seed decks
	seed1 := []string{"Giant", "Musketeer", "Wizard", "Fireball", "Arrows", "Knight", "Archers", "Skeletons"}
	seed2 := []string{"Hog Rider", "Ice Spirit", "Ice Golem", "Cannon", "Musketeer", "Fireball", "The Log", "Skeletons"}

	config := GeneticConfig{
		PopulationSize: 15,
		Generations:    10,
		MutationRate:   0.2,
		CrossoverRate:  0.7,
		EliteCount:     2,
		SeedPopulation: [][]string{seed1, seed2},
	}

	optimizer, err := NewGeneticOptimizer(candidates, deck.StrategyBalanced, &config)
	if err != nil {
		t.Fatalf("NewGeneticOptimizer() failed: %v", err)
	}

	result, err := optimizer.Optimize()
	if err != nil {
		t.Fatalf("Optimize() error = %v", err)
	}

	if result == nil || len(result.HallOfFame) == 0 {
		t.Fatal("Optimize() with seed population produced no results")
	}

	// Validate results
	for i, genome := range result.HallOfFame {
		if len(genome.Cards) != 8 {
			t.Errorf("HallOfFame[%d] has %d cards, want 8", i, len(genome.Cards))
		}
	}

	t.Logf("Seed population optimization: best fitness %.2f", result.Scores[0])
}

// TestGeneticAlgorithmIslandModel tests island-based evolution
func TestGeneticAlgorithmIslandModel(t *testing.T) {
	candidates := createRealisticCandidates()

	config := GeneticConfig{
		PopulationSize:    40, // Split across islands
		Generations:       15,
		MutationRate:      0.2,
		CrossoverRate:     0.7,
		EliteCount:        2,
		IslandModel:       true,
		IslandCount:       4,
		MigrationInterval: 5,
		MigrationSize:     3,
	}

	optimizer, err := NewGeneticOptimizer(candidates, deck.StrategyBalanced, &config)
	if err != nil {
		t.Fatalf("NewGeneticOptimizer() failed: %v", err)
	}

	result, err := optimizer.Optimize()
	if err != nil {
		t.Fatalf("Optimize() error = %v", err)
	}

	if result == nil {
		t.Fatal("Island model optimization returned nil")
	}

	if len(result.HallOfFame) == 0 {
		t.Error("Island model produced empty hall of fame")
	}

	// Validate decks
	for i, genome := range result.HallOfFame {
		if genome == nil || len(genome.Cards) != 8 {
			t.Errorf("HallOfFame[%d] invalid", i)
		}
	}

	t.Logf("Island model optimization: %d islands, best fitness %.2f",
		config.IslandCount, result.Scores[0])
}

// TestGeneticAlgorithmEarlyStoppingTarget tests early stopping with target fitness
func TestGeneticAlgorithmEarlyStoppingTarget(t *testing.T) {
	candidates := createRealisticCandidates()

	config := GeneticConfig{
		PopulationSize: 20,
		Generations:    100, // High number
		MutationRate:   0.2,
		CrossoverRate:  0.7,
		EliteCount:     3,
		TargetFitness:  6.0, // Stop when fitness reaches 6.0
	}

	optimizer, err := NewGeneticOptimizer(candidates, deck.StrategyBalanced, &config)
	if err != nil {
		t.Fatalf("NewGeneticOptimizer() failed: %v", err)
	}

	start := time.Now()
	result, err := optimizer.Optimize()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Optimize() error = %v", err)
	}

	if result.Generations == uint(config.Generations) {
		t.Logf("Note: Ran all %d generations (may not have reached target)", result.Generations)
	} else {
		t.Logf("Early stopped at generation %d (target: %.1f, achieved: %.2f)",
			result.Generations, config.TargetFitness, result.Scores[0])
	}

	// Should complete faster than running all generations
	if duration > 90*time.Second {
		t.Errorf("Optimization took %v, expected faster with target fitness", duration)
	}
}

// TestGeneticAlgorithmEarlyStoppingConvergence tests early stopping with convergence
func TestGeneticAlgorithmEarlyStoppingConvergence(t *testing.T) {
	candidates := createRealisticCandidates()

	config := GeneticConfig{
		PopulationSize:         15,
		Generations:            100, // High number
		MutationRate:           0.1, // Low mutation for faster convergence
		CrossoverRate:          0.9,
		EliteCount:             5,
		ConvergenceGenerations: 15, // Stop if no improvement for 15 generations
	}

	optimizer, err := NewGeneticOptimizer(candidates, deck.StrategyBalanced, &config)
	if err != nil {
		t.Fatalf("NewGeneticOptimizer() failed: %v", err)
	}

	result, err := optimizer.Optimize()
	if err != nil {
		t.Fatalf("Optimize() error = %v", err)
	}

	if result.Generations == uint(config.Generations) {
		t.Logf("Note: Ran all %d generations (population may not have converged)", result.Generations)
	} else {
		t.Logf("Converged at generation %d (convergence threshold: %d)",
			result.Generations, config.ConvergenceGenerations)
	}
}

// TestGeneticAlgorithmDifferentStrategies tests GA with different strategies
func TestGeneticAlgorithmDifferentStrategies(t *testing.T) {
	candidates := createRealisticCandidates()

	strategies := []deck.Strategy{
		deck.StrategyBalanced,
		deck.StrategyAggro,
		deck.StrategyControl,
	}

	config := GeneticConfig{
		PopulationSize: 15,
		Generations:    10,
		MutationRate:   0.2,
		CrossoverRate:  0.7,
		EliteCount:     2,
	}

	for _, strategy := range strategies {
		t.Run(string(strategy), func(t *testing.T) {
			optimizer, err := NewGeneticOptimizer(candidates, strategy, &config)
			if err != nil {
				t.Fatalf("NewGeneticOptimizer(%v) failed: %v", strategy, err)
			}

			result, err := optimizer.Optimize()
			if err != nil {
				t.Fatalf("Optimize() with %v error = %v", strategy, err)
			}

			if result == nil || len(result.HallOfFame) == 0 {
				t.Errorf("Optimize() with %v produced no results", strategy)
			}

			// Validate decks
			for i, genome := range result.HallOfFame {
				if len(genome.Cards) != 8 {
					t.Errorf("Strategy %v HallOfFame[%d] has %d cards", strategy, i, len(genome.Cards))
				}
			}

			t.Logf("Strategy %v: best fitness %.2f", strategy, result.Scores[0])
		})
	}
}

// TestGeneticAlgorithmPerformance tests performance characteristics
func TestGeneticAlgorithmPerformance(t *testing.T) {
	candidates := createRealisticCandidates()

	tests := []struct {
		name       string
		config     GeneticConfig
		maxTime    time.Duration
	}{
		{
			name: "quick optimization",
			config: GeneticConfig{
				PopulationSize: 10,
				Generations:    5,
				MutationRate:   0.2,
				CrossoverRate:  0.7,
				EliteCount:     2,
			},
			maxTime: 15 * time.Second,
		},
		{
			name: "medium optimization",
			config: GeneticConfig{
				PopulationSize: 30,
				Generations:    15,
				MutationRate:   0.2,
				CrossoverRate:  0.7,
				EliteCount:     3,
			},
			maxTime: 45 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optimizer, err := NewGeneticOptimizer(candidates, deck.StrategyBalanced, &tt.config)
			if err != nil {
				t.Fatalf("NewGeneticOptimizer() failed: %v", err)
			}

			start := time.Now()
			result, err := optimizer.Optimize()
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("Optimize() error = %v", err)
			}

			if duration > tt.maxTime {
				t.Errorf("Optimization took %v, expected < %v", duration, tt.maxTime)
			}

			t.Logf("%s: %d generations in %v (%.2f gen/s), best fitness %.2f",
				tt.name, result.Generations, duration,
				float64(result.Generations)/duration.Seconds(), result.Scores[0])
		})
	}
}

// Helper function to create realistic card candidates for testing
func createRealisticCandidates() []*deck.CardCandidate {
	cards := []struct {
		name   string
		elixir int
		role   config.CardRole
		rarity string
	}{
		{"Giant", 5, config.RoleWinCondition, "Rare"},
		{"Hog Rider", 4, config.RoleWinCondition, "Rare"},
		{"Golem", 8, config.RoleWinCondition, "Legendary"},
		{"Royal Giant", 6, config.RoleWinCondition, "Common"},
		{"X-Bow", 6, config.RoleWinCondition, "Epic"},

		{"Musketeer", 4, config.RoleSupport, "Rare"},
		{"Wizard", 5, config.RoleSupport, "Rare"},
		{"Baby Dragon", 4, config.RoleSupport, "Epic"},
		{"Electro Wizard", 4, config.RoleSupport, "Legendary"},
		{"Mega Minion", 3, config.RoleSupport, "Epic"},
		{"Archers", 3, config.RoleSupport, "Common"},
		{"Night Witch", 4, config.RoleSupport, "Legendary"},

		{"Cannon", 3, config.RoleBuilding, "Common"},
		{"Tesla", 4, config.RoleBuilding, "Epic"},
		{"Inferno Tower", 5, config.RoleBuilding, "Rare"},

		{"Fireball", 4, config.RoleSpellBig, "Rare"},
		{"Lightning", 6, config.RoleSpellBig, "Epic"},
		{"Rocket", 6, config.RoleSpellBig, "Rare"},

		{"The Log", 2, config.RoleSpellSmall, "Legendary"},
		{"Arrows", 3, config.RoleSpellSmall, "Common"},
		{"Zap", 2, config.RoleSpellSmall, "Common"},

		{"Skeletons", 1, config.RoleCycle, "Common"},
		{"Ice Spirit", 1, config.RoleCycle, "Common"},
		{"Ice Golem", 2, config.RoleCycle, "Rare"},
		{"Knight", 3, config.RoleCycle, "Common"},
		{"Goblin Gang", 3, config.RoleCycle, "Common"},
	}

	candidates := make([]*deck.CardCandidate, len(cards))
	for i, card := range cards {
		candidates[i] = &deck.CardCandidate{
			Name:     card.name,
			Level:    11,
			MaxLevel: 14,
			Rarity:   card.rarity,
			Elixir:   card.elixir,
			Role:     &card.role,
			Score:    float64(i) / 10.0,
		}
	}

	return candidates
}
