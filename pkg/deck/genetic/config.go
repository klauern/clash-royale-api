// Package genetic provides genetic algorithm-based deck optimization
// using the eaopt library for evolutionary deck generation.
package genetic

import (
	"fmt"
	"os"
	"strconv"
)

// GeneticConfig defines the parameters for genetic algorithm deck optimization.
// These parameters control the evolution process, population management,
// selection pressure, and genetic operators.
type GeneticConfig struct {
	// PopulationSize is the number of individuals in each generation.
	// Larger populations explore more solutions but require more computation.
	// Recommended: 50-200 for most deck optimization problems.
	PopulationSize int

	// Generations is the maximum number of generations to run.
	// The GA may terminate early if convergence is detected.
	// Recommended: 100-500 generations.
	Generations int

	// MutationRate is the probability that an individual gene mutates.
	// Range: 0.0 to 1.0. Typical values: 0.05-0.2 for deck building.
	MutationRate float64

	// CrossoverRate is the probability that two parents produce offspring via crossover.
	// Range: 0.0 to 1.0. Typical values: 0.6-0.9.
	CrossoverRate float64

	// MutationIntensity controls how radically mutations alter the deck.
	// 0.0 = minimal changes (swap 1-2 cards), 1.0 = major changes (swap 5+ cards)
	MutationIntensity float64

	// EliteCount is the number of top individuals copied unchanged to the next generation.
	// Elitism preserves the best solutions across generations.
	// Recommended: 1-5 individuals.
	EliteCount int

	// TournamentSize is the number of individuals competing in tournament selection.
	// Larger values increase selection pressure (favor better individuals more strongly).
	// Recommended: 3-7 for balanced exploration/exploitation.
	TournamentSize int

	// ParallelEvaluations enables concurrent fitness function evaluation.
	// Significantly speeds up evolution on multi-core systems.
	ParallelEvaluations bool

	// ConvergenceGenerations is the number of generations without improvement
	// before early termination. 0 disables early termination.
	// Recommended: 20-50 generations.
	ConvergenceGenerations int

	// TargetFitness is the fitness threshold for early termination.
	// The GA stops when an individual reaches or exceeds this score.
	// 0 disables fitness-based termination.
	TargetFitness float64

	// IslandModel enables parallel evolution with periodic migration.
	// Multiple populations evolve independently and exchange best individuals.
	IslandModel bool

	// IslandCount is the number of parallel populations when IslandModel is enabled.
	// Recommended: 4-8 islands for most systems.
	IslandCount int

	// MigrationInterval is the number of generations between island migrations.
	// Recommended: 10-20 generations.
	MigrationInterval int

	// MigrationSize is the number of individuals migrating between islands.
	// Recommended: 1-3 individuals.
	MigrationSize int

	// SeedPopulation is an optional initial population to start evolution.
	// If provided, evolution begins from these decks instead of random initialization.
	// Useful for resuming previous runs or warm-starting from known good decks.
	SeedPopulation [][]string

	// UseArchetypes indicates whether to enforce archetype constraints during evolution.
	// When true, generated decks will respect archetype composition rules.
	UseArchetypes bool
}

// DefaultGeneticConfig returns a configuration with sensible defaults
// for Clash Royale deck optimization.
func DefaultGeneticConfig() GeneticConfig {
	return GeneticConfig{
		PopulationSize:         100,
		Generations:            200,
		MutationRate:           0.1,
		CrossoverRate:          0.8,
		MutationIntensity:      0.3,
		EliteCount:             2,
		TournamentSize:         5,
		ParallelEvaluations:    true,
		ConvergenceGenerations: 30,
		TargetFitness:          0,
		IslandModel:            false,
		IslandCount:            4,
		MigrationInterval:      15,
		MigrationSize:          2,
		SeedPopulation:         nil,
		UseArchetypes:          true,
	}
}

const (
	envTrue        = "1"
	envTrueLiteral = "true"
)

// envParser contains functions for parsing environment variables.
type envParser struct {
	config *GeneticConfig
}

// parsePositiveInt parses a positive integer from an environment variable.
func (p *envParser) parsePositiveInt(key string, setter func(int)) {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			setter(i)
		}
	}
}

// parseNonNegativeInt parses a non-negative integer from an environment variable.
func (p *envParser) parseNonNegativeInt(key string, setter func(int)) {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 0 {
			setter(i)
		}
	}
}

// parseFloat01 parses a float in range [0, 1] from an environment variable.
func (p *envParser) parseFloat01(key string, setter func(float64)) {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 && f <= 1 {
			setter(f)
		}
	}
}

// parseNonNegativeFloat parses a non-negative float from an environment variable.
func (p *envParser) parseNonNegativeFloat(key string, setter func(float64)) {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 {
			setter(f)
		}
	}
}

// parseBool parses a boolean from an environment variable.
func (p *envParser) parseBool(key string, setter func(bool)) {
	if v := os.Getenv(key); v != "" {
		setter(v == envTrue || v == envTrueLiteral)
	}
}

// LoadFromEnv creates a GeneticConfig by reading from environment variables.
// Any unset variables use the default values.
// Variables:
//
//	GA_POPULATION_SIZE, GA_GENERATIONS, GA_MUTATION_RATE,
//	GA_CROSSOVER_RATE, GA_MUTATION_INTENSITY, GA_ELITE_COUNT,
//	GA_TOURNAMENT_SIZE, GA_PARALLEL_EVALUATIONS, GA_CONVERGENCE_GENERATIONS,
//	GA_TARGET_FITNESS, GA_ISLAND_MODEL, GA_ISLAND_COUNT,
//	GA_MIGRATION_INTERVAL, GA_MIGRATION_SIZE, GA_USE_ARCHETYPES
func LoadFromEnv() GeneticConfig {
	config := DefaultGeneticConfig()
	p := &envParser{config: &config}

	p.parsePositiveInt("GA_POPULATION_SIZE", func(v int) { config.PopulationSize = v })
	p.parsePositiveInt("GA_GENERATIONS", func(v int) { config.Generations = v })
	p.parseFloat01("GA_MUTATION_RATE", func(v float64) { config.MutationRate = v })
	p.parseFloat01("GA_CROSSOVER_RATE", func(v float64) { config.CrossoverRate = v })
	p.parseFloat01("GA_MUTATION_INTENSITY", func(v float64) { config.MutationIntensity = v })
	p.parseNonNegativeInt("GA_ELITE_COUNT", func(v int) { config.EliteCount = v })
	p.parsePositiveInt("GA_TOURNAMENT_SIZE", func(v int) { config.TournamentSize = v })
	p.parseBool("GA_PARALLEL_EVALUATIONS", func(v bool) { config.ParallelEvaluations = v })
	p.parseNonNegativeInt("GA_CONVERGENCE_GENERATIONS", func(v int) { config.ConvergenceGenerations = v })
	p.parseNonNegativeFloat("GA_TARGET_FITNESS", func(v float64) { config.TargetFitness = v })
	p.parseBool("GA_ISLAND_MODEL", func(v bool) { config.IslandModel = v })
	p.parsePositiveInt("GA_ISLAND_COUNT", func(v int) { config.IslandCount = v })
	p.parsePositiveInt("GA_MIGRATION_INTERVAL", func(v int) { config.MigrationInterval = v })
	p.parsePositiveInt("GA_MIGRATION_SIZE", func(v int) { config.MigrationSize = v })
	p.parseBool("GA_USE_ARCHETYPES", func(v bool) { config.UseArchetypes = v })

	return config
}

// Validate checks if the configuration is valid for use.
func (c *GeneticConfig) Validate() error {
	if c.PopulationSize <= 0 {
		return fmt.Errorf("population_size must be positive, got %d", c.PopulationSize)
	}
	if c.Generations <= 0 {
		return fmt.Errorf("generations must be positive, got %d", c.Generations)
	}
	if c.MutationRate < 0 || c.MutationRate > 1 {
		return fmt.Errorf("mutation_rate must be between 0 and 1, got %f", c.MutationRate)
	}
	if c.CrossoverRate < 0 || c.CrossoverRate > 1 {
		return fmt.Errorf("crossover_rate must be between 0 and 1, got %f", c.CrossoverRate)
	}
	if c.MutationIntensity < 0 || c.MutationIntensity > 1 {
		return fmt.Errorf("mutation_intensity must be between 0 and 1, got %f", c.MutationIntensity)
	}
	if c.EliteCount < 0 {
		return fmt.Errorf("elite_count must be non-negative, got %d", c.EliteCount)
	}
	if c.EliteCount >= c.PopulationSize {
		return fmt.Errorf("elite_count (%d) must be less than population_size (%d)", c.EliteCount, c.PopulationSize)
	}
	if c.TournamentSize <= 0 {
		return fmt.Errorf("tournament_size must be positive, got %d", c.TournamentSize)
	}
	if c.TournamentSize > c.PopulationSize {
		return fmt.Errorf("tournament_size (%d) must not exceed population_size (%d)", c.TournamentSize, c.PopulationSize)
	}
	if c.ConvergenceGenerations < 0 {
		return fmt.Errorf("convergence_generations must be non-negative, got %d", c.ConvergenceGenerations)
	}
	if c.TargetFitness < 0 {
		return fmt.Errorf("target_fitness must be non-negative, got %f", c.TargetFitness)
	}
	if c.IslandModel {
		if c.IslandCount <= 1 {
			return fmt.Errorf("island_count must be at least 2 when island_model is enabled, got %d", c.IslandCount)
		}
		if c.MigrationInterval <= 0 {
			return fmt.Errorf("migration_interval must be positive when island_model is enabled, got %d", c.MigrationInterval)
		}
		if c.MigrationSize <= 0 {
			return fmt.Errorf("migration_size must be positive when island_model is enabled, got %d", c.MigrationSize)
		}
		if c.MigrationSize >= c.PopulationSize/c.IslandCount {
			return fmt.Errorf("migration_size (%d) must be less than per-island population (%d)",
				c.MigrationSize, c.PopulationSize/c.IslandCount)
		}
	}
	return nil
}

// String returns a human-readable representation of the configuration.
func (c *GeneticConfig) String() string {
	return fmt.Sprintf(
		"GeneticConfig{PopulationSize:%d, Generations:%d, MutationRate:%.2f, CrossoverRate:%.2f, EliteCount:%d, TournamentSize:%d}",
		c.PopulationSize, c.Generations, c.MutationRate, c.CrossoverRate, c.EliteCount, c.TournamentSize,
	)
}
