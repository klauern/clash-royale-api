// Package genetic provides genetic algorithm-based deck optimization
// using the eaopt library for evolutionary deck generation.
package genetic

import (
	"os"
	"testing"
)

func TestDefaultGeneticConfig(t *testing.T) {
	config := DefaultGeneticConfig()

	tests := []struct {
		name  string
		check func(GeneticConfig) bool
		want  bool
	}{
		{
			name:  "positive population size",
			check: func(c GeneticConfig) bool { return c.PopulationSize > 0 },
			want:  true,
		},
		{
			name:  "positive generations",
			check: func(c GeneticConfig) bool { return c.Generations > 0 },
			want:  true,
		},
		{
			name:  "mutation rate in valid range",
			check: func(c GeneticConfig) bool { return c.MutationRate >= 0 && c.MutationRate <= 1 },
			want:  true,
		},
		{
			name:  "crossover rate in valid range",
			check: func(c GeneticConfig) bool { return c.CrossoverRate >= 0 && c.CrossoverRate <= 1 },
			want:  true,
		},
		{
			name:  "elite count less than population",
			check: func(c GeneticConfig) bool { return c.EliteCount < c.PopulationSize },
			want:  true,
		},
		{
			name:  "mutation intensity in valid range",
			check: func(c GeneticConfig) bool { return c.MutationIntensity >= 0 && c.MutationIntensity <= 1 },
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.check(config); got != tt.want {
				t.Errorf("DefaultGeneticConfig() check failed")
			}
		})
	}

	// Validate the entire config
	if err := config.Validate(); err != nil {
		t.Errorf("DefaultGeneticConfig() produced invalid config: %v", err)
	}
}

func TestGeneticConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  GeneticConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: GeneticConfig{
				PopulationSize:         100,
				Generations:            200,
				MutationRate:           0.1,
				CrossoverRate:          0.8,
				MutationIntensity:      0.3,
				EliteCount:             2,
				TournamentSize:         5,
				ConvergenceGenerations: 30,
				TargetFitness:          0,
			},
			wantErr: false,
		},
		{
			name: "zero population size",
			config: GeneticConfig{
				PopulationSize: 0,
				Generations:    100,
			},
			wantErr: true,
		},
		{
			name: "negative population size",
			config: GeneticConfig{
				PopulationSize: -10,
				Generations:    100,
			},
			wantErr: true,
		},
		{
			name: "zero generations",
			config: GeneticConfig{
				PopulationSize: 100,
				Generations:    0,
			},
			wantErr: true,
		},
		{
			name: "mutation rate too low",
			config: GeneticConfig{
				PopulationSize: 100,
				Generations:    100,
				MutationRate:   -0.1,
			},
			wantErr: true,
		},
		{
			name: "mutation rate too high",
			config: GeneticConfig{
				PopulationSize: 100,
				Generations:    100,
				MutationRate:   1.5,
			},
			wantErr: true,
		},
		{
			name: "crossover rate too low",
			config: GeneticConfig{
				PopulationSize: 100,
				Generations:    100,
				CrossoverRate:  -0.1,
			},
			wantErr: true,
		},
		{
			name: "crossover rate too high",
			config: GeneticConfig{
				PopulationSize: 100,
				Generations:    100,
				CrossoverRate:  1.5,
			},
			wantErr: true,
		},
		{
			name: "mutation intensity out of range",
			config: GeneticConfig{
				PopulationSize:    100,
				Generations:       100,
				MutationIntensity: 1.5,
			},
			wantErr: true,
		},
		{
			name: "elite count equals population size",
			config: GeneticConfig{
				PopulationSize: 100,
				Generations:    100,
				EliteCount:     100,
			},
			wantErr: true,
		},
		{
			name: "elite count exceeds population size",
			config: GeneticConfig{
				PopulationSize: 100,
				Generations:    100,
				EliteCount:     150,
			},
			wantErr: true,
		},
		{
			name: "zero tournament size",
			config: GeneticConfig{
				PopulationSize: 100,
				Generations:    100,
				TournamentSize: 0,
			},
			wantErr: true,
		},
		{
			name: "tournament size exceeds population",
			config: GeneticConfig{
				PopulationSize: 50,
				Generations:    100,
				TournamentSize: 100,
			},
			wantErr: true,
		},
		{
			name: "negative convergence generations",
			config: GeneticConfig{
				PopulationSize:         100,
				Generations:            100,
				ConvergenceGenerations: -10,
			},
			wantErr: true,
		},
		{
			name: "negative target fitness",
			config: GeneticConfig{
				PopulationSize: 100,
				Generations:    100,
				TargetFitness:  -1.0,
			},
			wantErr: true,
		},
		{
			name: "island model with insufficient islands",
			config: GeneticConfig{
				PopulationSize: 100,
				Generations:    100,
				IslandModel:    true,
				IslandCount:    1,
			},
			wantErr: true,
		},
		{
			name: "island model with zero migration interval",
			config: GeneticConfig{
				PopulationSize:    100,
				Generations:       100,
				IslandModel:       true,
				IslandCount:       4,
				MigrationInterval: 0,
			},
			wantErr: true,
		},
		{
			name: "island model with zero migration size",
			config: GeneticConfig{
				PopulationSize:    100,
				Generations:       100,
				IslandModel:       true,
				IslandCount:       4,
				MigrationInterval: 15,
				MigrationSize:     0,
			},
			wantErr: true,
		},
		{
			name: "island model with migration size too large",
			config: GeneticConfig{
				PopulationSize:    100,
				Generations:       100,
				IslandModel:       true,
				IslandCount:       4,
				MigrationInterval: 15,
				MigrationSize:     30, // Per-island pop is 25, so 30 is too large
			},
			wantErr: true,
		},
		{
			name: "valid island model config",
			config: GeneticConfig{
				PopulationSize:    100,
				Generations:       100,
				TournamentSize:    5,
				IslandModel:       true,
				IslandCount:       4,
				MigrationInterval: 15,
				MigrationSize:     2,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneticConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	envVars := []string{
		"GA_POPULATION_SIZE", "GA_GENERATIONS", "GA_MUTATION_RATE",
		"GA_CROSSOVER_RATE", "GA_MUTATION_INTENSITY", "GA_ELITE_COUNT",
		"GA_TOURNAMENT_SIZE", "GA_PARALLEL_EVALUATIONS",
		"GA_CONVERGENCE_GENERATIONS", "GA_TARGET_FITNESS",
		"GA_ISLAND_MODEL", "GA_ISLAND_COUNT",
		"GA_MIGRATION_INTERVAL", "GA_MIGRATION_SIZE", "GA_USE_ARCHETYPES",
	}
	for _, v := range envVars {
		originalEnv[v] = os.Getenv(v)
		os.Unsetenv(v)
	}
	defer func() {
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	t.Run("uses defaults when no env vars set", func(t *testing.T) {
		config := LoadFromEnv()
		defaultConfig := DefaultGeneticConfig()

		if config.PopulationSize != defaultConfig.PopulationSize {
			t.Errorf("LoadFromEnv() PopulationSize = %v, want %v", config.PopulationSize, defaultConfig.PopulationSize)
		}
		if config.Generations != defaultConfig.Generations {
			t.Errorf("LoadFromEnv() Generations = %v, want %v", config.Generations, defaultConfig.Generations)
		}
	})

	t.Run("overrides population size from env", func(t *testing.T) {
		os.Setenv("GA_POPULATION_SIZE", "250")
		config := LoadFromEnv()

		if config.PopulationSize != 250 {
			t.Errorf("LoadFromEnv() PopulationSize = %v, want 250", config.PopulationSize)
		}
	})

	t.Run("overrides generations from env", func(t *testing.T) {
		os.Setenv("GA_GENERATIONS", "500")
		config := LoadFromEnv()

		if config.Generations != 500 {
			t.Errorf("LoadFromEnv() Generations = %v, want 500", config.Generations)
		}
	})

	t.Run("overrides mutation rate from env", func(t *testing.T) {
		os.Setenv("GA_MUTATION_RATE", "0.15")
		config := LoadFromEnv()

		if config.MutationRate != 0.15 {
			t.Errorf("LoadFromEnv() MutationRate = %v, want 0.15", config.MutationRate)
		}
	})

	t.Run("overrides crossover rate from env", func(t *testing.T) {
		os.Setenv("GA_CROSSOVER_RATE", "0.7")
		config := LoadFromEnv()

		if config.CrossoverRate != 0.7 {
			t.Errorf("LoadFromEnv() CrossoverRate = %v, want 0.7", config.CrossoverRate)
		}
	})

	t.Run("overrides mutation intensity from env", func(t *testing.T) {
		os.Setenv("GA_MUTATION_INTENSITY", "0.5")
		config := LoadFromEnv()

		if config.MutationIntensity != 0.5 {
			t.Errorf("LoadFromEnv() MutationIntensity = %v, want 0.5", config.MutationIntensity)
		}
	})

	t.Run("overrides elite count from env", func(t *testing.T) {
		os.Setenv("GA_ELITE_COUNT", "5")
		config := LoadFromEnv()

		if config.EliteCount != 5 {
			t.Errorf("LoadFromEnv() EliteCount = %v, want 5", config.EliteCount)
		}
	})

	t.Run("overrides tournament size from env", func(t *testing.T) {
		os.Setenv("GA_TOURNAMENT_SIZE", "7")
		config := LoadFromEnv()

		if config.TournamentSize != 7 {
			t.Errorf("LoadFromEnv() TournamentSize = %v, want 7", config.TournamentSize)
		}
	})

	t.Run("overrides parallel evaluations from env (true literal)", func(t *testing.T) {
		os.Setenv("GA_PARALLEL_EVALUATIONS", "true")
		config := LoadFromEnv()

		if !config.ParallelEvaluations {
			t.Errorf("LoadFromEnv() ParallelEvaluations = %v, want true", config.ParallelEvaluations)
		}
	})

	t.Run("overrides parallel evaluations from env (1)", func(t *testing.T) {
		os.Setenv("GA_PARALLEL_EVALUATIONS", "1")
		config := LoadFromEnv()

		if !config.ParallelEvaluations {
			t.Errorf("LoadFromEnv() ParallelEvaluations = %v, want true", config.ParallelEvaluations)
		}
	})

	t.Run("overrides parallel evaluations from env (0)", func(t *testing.T) {
		os.Setenv("GA_PARALLEL_EVALUATIONS", "0")
		config := LoadFromEnv()

		if config.ParallelEvaluations {
			t.Errorf("LoadFromEnv() ParallelEvaluations = %v, want false", config.ParallelEvaluations)
		}
	})

	t.Run("overrides convergence generations from env", func(t *testing.T) {
		os.Setenv("GA_CONVERGENCE_GENERATIONS", "50")
		config := LoadFromEnv()

		if config.ConvergenceGenerations != 50 {
			t.Errorf("LoadFromEnv() ConvergenceGenerations = %v, want 50", config.ConvergenceGenerations)
		}
	})

	t.Run("overrides target fitness from env", func(t *testing.T) {
		os.Setenv("GA_TARGET_FITNESS", "0.95")
		config := LoadFromEnv()

		if config.TargetFitness != 0.95 {
			t.Errorf("LoadFromEnv() TargetFitness = %v, want 0.95", config.TargetFitness)
		}
	})

	t.Run("overrides island model from env", func(t *testing.T) {
		os.Setenv("GA_ISLAND_MODEL", "true")
		os.Setenv("GA_ISLAND_COUNT", "8")
		os.Setenv("GA_MIGRATION_INTERVAL", "20")
		os.Setenv("GA_MIGRATION_SIZE", "3")
		config := LoadFromEnv()

		if !config.IslandModel {
			t.Errorf("LoadFromEnv() IslandModel = %v, want true", config.IslandModel)
		}
		if config.IslandCount != 8 {
			t.Errorf("LoadFromEnv() IslandCount = %v, want 8", config.IslandCount)
		}
		if config.MigrationInterval != 20 {
			t.Errorf("LoadFromEnv() MigrationInterval = %v, want 20", config.MigrationInterval)
		}
		if config.MigrationSize != 3 {
			t.Errorf("LoadFromEnv() MigrationSize = %v, want 3", config.MigrationSize)
		}
	})

	t.Run("ignores invalid env values", func(t *testing.T) {
		os.Setenv("GA_POPULATION_SIZE", "invalid")
		os.Setenv("GA_MUTATION_RATE", "2.5") // Out of range
		config := LoadFromEnv()

		// Should fall back to defaults for invalid values
		if config.PopulationSize <= 0 {
			t.Errorf("LoadFromEnv() should use default for invalid population size")
		}
	})

	t.Run("loaded config is valid", func(t *testing.T) {
		os.Setenv("GA_POPULATION_SIZE", "150")
		os.Setenv("GA_GENERATIONS", "300")
		config := LoadFromEnv()

		if err := config.Validate(); err != nil {
			t.Errorf("LoadFromEnv() produced invalid config: %v", err)
		}
	})
}

func TestGeneticConfigString(t *testing.T) {
	config := GeneticConfig{
		PopulationSize: 100,
		Generations:    200,
		MutationRate:   0.1,
		CrossoverRate:  0.8,
		EliteCount:     2,
		TournamentSize: 5,
	}

	s := config.String()
	if s == "" {
		t.Error("GeneticConfig.String() returned empty string")
	}

	// Check it contains key info
	expectedSubstrings := []string{
		"PopulationSize:100",
		"Generations:200",
		"MutationRate:0.10",
		"CrossoverRate:0.80",
		"EliteCount:2",
		"TournamentSize:5",
	}

	for _, substr := range expectedSubstrings {
		if !contains(s, substr) {
			t.Errorf("GeneticConfig.String() = %v, should contain %v", s, substr)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}
