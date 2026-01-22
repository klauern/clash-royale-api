// Package genetic provides genetic algorithm-based deck optimization
// using the eaopt library for evolutionary deck generation.
package genetic

import (
	"math/rand"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestNewGeneticOptimizer(t *testing.T) {
	tests := []struct {
		name        string
		candidates  []*deck.CardCandidate
		strategy    deck.Strategy
		config      *GeneticConfig
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid optimizer with default config",
			candidates: createMockCandidates(10),
			strategy:   deck.StrategyBalanced,
			config:     nil, // Should use default
			wantErr:    false,
		},
		{
			name:       "valid optimizer with custom config",
			candidates: createMockCandidates(10),
			strategy:   deck.StrategyBalanced,
			config: &GeneticConfig{
				PopulationSize: 50,
				Generations:    100,
				MutationRate:   0.2,
				CrossoverRate:  0.7,
				TournamentSize: 5,
			},
			wantErr: false,
		},
		{
			name:        "insufficient candidates",
			candidates:  createMockCandidates(5),
			strategy:    deck.StrategyBalanced,
			config:      nil,
			wantErr:     true,
			errContains: "insufficient candidates",
		},
		{
			name:        "zero candidates",
			candidates:  createMockCandidates(0),
			strategy:    deck.StrategyBalanced,
			config:      nil,
			wantErr:     true,
			errContains: "insufficient candidates",
		},
		{
			name:       "invalid config",
			candidates: createMockCandidates(10),
			strategy:   deck.StrategyBalanced,
			config: &GeneticConfig{
				PopulationSize: -10, // Invalid
				Generations:    100,
			},
			wantErr:     true,
			errContains: "population_size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optimizer, err := NewGeneticOptimizer(tt.candidates, tt.strategy, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewGeneticOptimizer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("NewGeneticOptimizer() error = %v, should contain %v", err, tt.errContains)
				}
			}

			if !tt.wantErr {
				if optimizer == nil {
					t.Fatal("NewGeneticOptimizer() returned nil optimizer without error")
				}
				if optimizer.Config == nil {
					t.Error("NewGeneticOptimizer() optimizer has nil config")
				}
				if len(optimizer.Candidates) != len(tt.candidates) {
					t.Errorf("NewGeneticOptimizer() candidates = %d, want %d", len(optimizer.Candidates), len(tt.candidates))
				}
				if optimizer.Strategy != tt.strategy {
					t.Errorf("NewGeneticOptimizer() strategy = %v, want %v", optimizer.Strategy, tt.strategy)
				}
			}
		})
	}
}

func TestGeneticOptimizerOptimize(t *testing.T) {
	t.Run("basic optimization execution", func(t *testing.T) {
		candidates := createMockCandidates(15)
		config := GeneticConfig{
			PopulationSize:         10,
			Generations:            5,
			MutationRate:           0.2,
			CrossoverRate:          0.7,
			MutationIntensity:      0.3,
			EliteCount:             2,
			TournamentSize:         3,
			ParallelEvaluations:    false,
			ConvergenceGenerations: 0,
			TargetFitness:          0,
		}

		optimizer, err := NewGeneticOptimizer(candidates, deck.StrategyBalanced, &config)
		if err != nil {
			t.Fatalf("NewGeneticOptimizer() failed: %v", err)
		}

		// Use fixed RNG for reproducibility
		optimizer.RNG = rand.New(rand.NewSource(42))

		result, err := optimizer.Optimize()
		if err != nil {
			t.Fatalf("Optimize() error = %v", err)
		}

		if result == nil {
			t.Fatal("Optimize() returned nil result")
		}

		// Validate result structure
		if result.HallOfFame == nil {
			t.Error("Optimize() result has nil HallOfFame")
		}
		if result.Scores == nil {
			t.Error("Optimize() result has nil Scores")
		}
		if len(result.HallOfFame) != len(result.Scores) {
			t.Errorf("Optimize() HallOfFame length %d != Scores length %d", len(result.HallOfFame), len(result.Scores))
		}
		if result.Generations == 0 {
			t.Error("Optimize() result has zero generations")
		}
		if result.Duration == 0 {
			t.Error("Optimize() result has zero duration")
		}
	})

	t.Run("hall of fame contains valid decks", func(t *testing.T) {
		candidates := createMockCandidates(15)
		config := GeneticConfig{
			PopulationSize: 10,
			Generations:    5,
			MutationRate:   0.2,
			CrossoverRate:  0.7,
			EliteCount:     3,
			TournamentSize: 3,
		}

		optimizer, err := NewGeneticOptimizer(candidates, deck.StrategyBalanced, &config)
		if err != nil {
			t.Fatalf("NewGeneticOptimizer() failed: %v", err)
		}

		result, err := optimizer.Optimize()
		if err != nil {
			t.Fatalf("Optimize() error = %v", err)
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

			// Check for unique cards
			cardSet := make(map[string]bool)
			for _, card := range genome.Cards {
				if cardSet[card] {
					t.Errorf("HallOfFame[%d] has duplicate card: %v", i, card)
				}
				cardSet[card] = true
			}
		}
	})

	t.Run("fitness scores in valid range", func(t *testing.T) {
		candidates := createMockCandidates(15)
		config := GeneticConfig{
			PopulationSize: 10,
			Generations:    5,
			MutationRate:   0.2,
			CrossoverRate:  0.7,
			EliteCount:     2,
			TournamentSize: 3,
		}

		optimizer, err := NewGeneticOptimizer(candidates, deck.StrategyBalanced, &config)
		if err != nil {
			t.Fatalf("NewGeneticOptimizer() failed: %v", err)
		}

		result, err := optimizer.Optimize()
		if err != nil {
			t.Fatalf("Optimize() error = %v", err)
		}

		// Validate fitness scores
		for i, score := range result.Scores {
			if score < 0 || score > 10 {
				t.Errorf("Scores[%d] = %v, want in range [0, 10]", i, score)
			}
		}
	})

	t.Run("hall of fame sorted by fitness", func(t *testing.T) {
		candidates := createMockCandidates(15)
		config := GeneticConfig{
			PopulationSize: 10,
			Generations:    5,
			MutationRate:   0.2,
			CrossoverRate:  0.7,
			EliteCount:     3,
			TournamentSize: 3,
		}

		optimizer, err := NewGeneticOptimizer(candidates, deck.StrategyBalanced, &config)
		if err != nil {
			t.Fatalf("NewGeneticOptimizer() failed: %v", err)
		}

		result, err := optimizer.Optimize()
		if err != nil {
			t.Fatalf("Optimize() error = %v", err)
		}

		// Scores should be in descending order (best first)
		for i := 1; i < len(result.Scores); i++ {
			if result.Scores[i] > result.Scores[i-1] {
				t.Errorf("Scores not sorted: Scores[%d] = %v > Scores[%d] = %v",
					i, result.Scores[i], i-1, result.Scores[i-1])
			}
		}
	})

	t.Run("nil optimizer", func(t *testing.T) {
		var optimizer *GeneticOptimizer
		_, err := optimizer.Optimize()
		if err == nil {
			t.Error("Optimize() with nil optimizer should error")
		}
		if !contains(err.Error(), "nil") {
			t.Errorf("Optimize() error should mention nil, got: %v", err)
		}
	})

	t.Run("optimizer with nil config", func(t *testing.T) {
		candidates := createMockCandidates(10)
		optimizer := &GeneticOptimizer{
			Candidates: candidates,
			Strategy:   deck.StrategyBalanced,
			Config:     nil,
		}

		_, err := optimizer.Optimize()
		if err == nil {
			t.Error("Optimize() with nil config should error")
		}
	})

	t.Run("optimizer with insufficient candidates", func(t *testing.T) {
		candidates := createMockCandidates(5)
		config := DefaultGeneticConfig()

		optimizer := &GeneticOptimizer{
			Candidates: candidates,
			Strategy:   deck.StrategyBalanced,
			Config:     &config,
		}

		_, err := optimizer.Optimize()
		if err == nil {
			t.Error("Optimize() with insufficient candidates should error")
		}
		if !contains(err.Error(), "insufficient candidates") {
			t.Errorf("Optimize() error should mention insufficient candidates, got: %v", err)
		}
	})
}

func TestGeneticOptimizerProgressCallback(t *testing.T) {
	candidates := createMockCandidates(15)
	config := GeneticConfig{
		PopulationSize: 10,
		Generations:    3,
		MutationRate:   0.2,
		CrossoverRate:  0.7,
		EliteCount:     2,
		TournamentSize: 3,
	}

	optimizer, err := NewGeneticOptimizer(candidates, deck.StrategyBalanced, &config)
	if err != nil {
		t.Fatalf("NewGeneticOptimizer() failed: %v", err)
	}

	// Track progress callbacks
	callCount := 0
	var progressUpdates []GeneticProgress

	optimizer.Progress = func(p GeneticProgress) {
		callCount++
		progressUpdates = append(progressUpdates, p)
	}

	_, err = optimizer.Optimize()
	if err != nil {
		t.Fatalf("Optimize() error = %v", err)
	}

	// Should have received progress callbacks
	if callCount == 0 {
		t.Error("Progress callback never called")
	}

	// Validate progress updates
	for i, p := range progressUpdates {
		if p.BestFitness < 0 || p.BestFitness > 10 {
			t.Errorf("Progress[%d] BestFitness = %v, want in range [0, 10]", i, p.BestFitness)
		}
		if p.AvgFitness < 0 || p.AvgFitness > 10 {
			t.Errorf("Progress[%d] AvgFitness = %v, want in range [0, 10]", i, p.AvgFitness)
		}
		if p.Populations <= 0 {
			t.Errorf("Progress[%d] Populations = %d, want > 0", i, p.Populations)
		}
	}

	// Generation numbers should increase
	if len(progressUpdates) > 1 {
		for i := 1; i < len(progressUpdates); i++ {
			if progressUpdates[i].Generation < progressUpdates[i-1].Generation {
				t.Errorf("Generation numbers not increasing: %d -> %d",
					progressUpdates[i-1].Generation, progressUpdates[i].Generation)
			}
		}
	}
}

func TestGeneticOptimizerEarlyStopping(t *testing.T) {
	t.Run("target fitness early stop", func(t *testing.T) {
		candidates := createMockCandidates(15)
		config := GeneticConfig{
			PopulationSize: 20,
			Generations:    100, // High number
			MutationRate:   0.2,
			CrossoverRate:  0.7,
			EliteCount:     2,
			TournamentSize: 5,
			TargetFitness:  5.0, // Stop when fitness reaches 5.0
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

		// Should stop early if target fitness reached
		if result.Generations == uint(config.Generations) {
			t.Logf("Note: Ran all %d generations (may not have reached target fitness)", result.Generations)
		}

		// Should complete reasonably fast (target fitness should trigger early stop)
		if duration > 30*time.Second {
			t.Errorf("Optimize() took %v, expected faster with target fitness", duration)
		}
	})

	t.Run("convergence early stop", func(t *testing.T) {
		candidates := createMockCandidates(15)
		config := GeneticConfig{
			PopulationSize:         10,
			Generations:            100, // High number
			MutationRate:           0.1, // Low mutation for faster convergence
			CrossoverRate:          0.9,
			EliteCount:             3,
			TournamentSize:         5,
			ConvergenceGenerations: 10, // Stop if no improvement for 10 generations
		}

		optimizer, err := NewGeneticOptimizer(candidates, deck.StrategyBalanced, &config)
		if err != nil {
			t.Fatalf("NewGeneticOptimizer() failed: %v", err)
		}

		result, err := optimizer.Optimize()
		if err != nil {
			t.Fatalf("Optimize() error = %v", err)
		}

		// Should stop early due to convergence
		if result.Generations == uint(config.Generations) {
			t.Logf("Note: Ran all %d generations (population may not have converged)", result.Generations)
		}
	})
}

func TestGeneticOptimizerIslandModel(t *testing.T) {
	candidates := createMockCandidates(20)
	config := GeneticConfig{
		PopulationSize:    40, // Will be split across islands
		Generations:       10,
		MutationRate:      0.2,
		CrossoverRate:     0.7,
		EliteCount:        2,
		TournamentSize:    5,
		IslandModel:       true,
		IslandCount:       4,
		MigrationInterval: 3,
		MigrationSize:     2,
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
		t.Fatal("Optimize() returned nil result")
	}

	// Validate result contains valid decks
	if len(result.HallOfFame) == 0 {
		t.Error("Island model optimization produced empty hall of fame")
	}

	for i, genome := range result.HallOfFame {
		if genome == nil || len(genome.Cards) != 8 {
			t.Errorf("HallOfFame[%d] invalid", i)
		}
	}
}

func TestGeneticOptimizerSeedPopulation(t *testing.T) {
	candidates := createMockCandidates(15)

	// Create seed decks
	seed1 := []string{"Card0", "Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7"}
	seed2 := []string{"Card7", "Card8", "Card9", "Card10", "Card11", "Card12", "Card13", "Card14"}

	config := GeneticConfig{
		PopulationSize: 10,
		Generations:    5,
		MutationRate:   0.2,
		CrossoverRate:  0.7,
		EliteCount:     2,
		TournamentSize: 3,
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

	// Seed population should influence results (hall of fame should contain valid decks)
	for i, genome := range result.HallOfFame {
		if len(genome.Cards) != 8 {
			t.Errorf("HallOfFame[%d] has %d cards, want 8", i, len(genome.Cards))
		}
	}
}

func TestGeneticOptimizerPopulationConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         GeneticConfig
		wantPopSize    uint
		wantPopCount   uint
	}{
		{
			name: "single population (no island model)",
			config: GeneticConfig{
				PopulationSize: 100,
				IslandModel:    false,
			},
			wantPopSize:  100,
			wantPopCount: 1,
		},
		{
			name: "island model with even split",
			config: GeneticConfig{
				PopulationSize: 100,
				IslandModel:    true,
				IslandCount:    4,
			},
			wantPopSize:  25, // 100 / 4
			wantPopCount: 4,
		},
		{
			name: "island model with uneven split",
			config: GeneticConfig{
				PopulationSize: 100,
				IslandModel:    true,
				IslandCount:    3,
			},
			wantPopSize:  33, // 100 / 3 = 33
			wantPopCount: 3,
		},
		{
			name: "island model with small population",
			config: GeneticConfig{
				PopulationSize: 2,
				IslandModel:    true,
				IslandCount:    5,
			},
			wantPopSize:  1, // Minimum 1 per island
			wantPopCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optimizer := &GeneticOptimizer{
				Config: &tt.config,
			}

			popSize, popCount := optimizer.populationConfig()

			if popSize != tt.wantPopSize {
				t.Errorf("populationConfig() popSize = %d, want %d", popSize, tt.wantPopSize)
			}
			if popCount != tt.wantPopCount {
				t.Errorf("populationConfig() popCount = %d, want %d", popCount, tt.wantPopCount)
			}
		})
	}
}

func TestGeneticOptimizerPerformance(t *testing.T) {
	candidates := createMockCandidates(20)
	config := GeneticConfig{
		PopulationSize:      20,
		Generations:         10,
		MutationRate:        0.2,
		CrossoverRate:       0.7,
		EliteCount:          3,
		TournamentSize:      5,
		ParallelEvaluations: false,
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

	// Should complete in reasonable time (small GA run)
	if duration > 30*time.Second {
		t.Errorf("Optimize() took %v, expected < 30s for small GA run", duration)
	}

	// Result duration should match actual duration (roughly)
	if result.Duration > duration+time.Second {
		t.Errorf("Result.Duration = %v > actual duration %v", result.Duration, duration)
	}

	t.Logf("Optimization completed in %v for %d generations", duration, result.Generations)
}
