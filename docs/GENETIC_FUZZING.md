# Genetic Algorithm Deck Optimization

The genetic algorithm (GA) deck optimization feature uses evolutionary computation to discover optimal deck combinations from a player's card collection. Unlike Monte Carlo fuzzing which evaluates random decks, the genetic algorithm iteratively evolves and improves deck solutions over multiple generations.

## Overview

Genetic algorithm optimization works by:
1. **Population Initialization**: Creates an initial population of random valid decks
2. **Evolution**: Iteratively improves decks through selection, crossover (breeding), and mutation
3. **Fitness Evaluation**: Scores each deck using the comprehensive evaluation system
4. **Elite Preservation**: Keeps the best decks across generations
5. **Convergence**: Continues until reaching a solution threshold or generation limit

This approach is particularly effective when you want to find the absolute best deck from a large card collection, as the GA explores the solution space more intelligently than random sampling.

## Usage

### Basic Usage

```bash
# Run genetic optimization with default settings
./bin/cr-api deck fuzz --mode genetic --tag R8QGUQRCV
```

Default configuration:
- Population: 100 decks per generation
- Generations: 100
- Mutation rate: 20%
- Crossover rate: 70%

### Quick Optimization

```bash
# Faster optimization for initial exploration
./bin/cr-api deck fuzz --mode genetic --tag R8QGUQRCV \
  --ga-population 50 \
  --ga-generations 30
```

### Thorough Search

```bash
# More thorough search for best possible decks
./bin/cr-api deck fuzz --mode genetic --tag R8QGUQRCV \
  --ga-population 200 \
  --ga-generations 300 \
  --ga-elite-count 20
```

### Island Model for Diversity

```bash
# Use island model to maintain population diversity
./bin/cr-api deck fuzz --mode genetic --tag R8QGUQRCV \
  --ga-island-model \
  --ga-island-count 4 \
  --ga-migration-interval 10 \
  --ga-migration-size 5
```

### With Early Stopping

```bash
# Stop when target fitness reached or convergence detected
./bin/cr-api deck fuzz --mode genetic --tag R8QGUQRCV \
  --ga-target-fitness 9.0 \
  --ga-convergence-generations 25
```

### Advanced Configuration

```bash
# Fine-tune GA parameters for specific needs
./bin/cr-api deck fuzz --mode genetic --tag R8QGUQRCV \
  --ga-population 150 \
  --ga-generations 200 \
  --ga-mutation-rate 0.25 \
  --ga-crossover-rate 0.75 \
  --ga-mutation-intensity 0.4 \
  --ga-elite-count 15 \
  --ga-tournament-size 7 \
  --ga-parallel-eval \
  --verbose
```

## Command Options

### General Fuzz Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--mode` | string | random | Fuzzing mode: `random` or `genetic` |
| `--tag`, `-p` | string | (required) | Player tag (without #) |
| `--top` | int | 10 | Number of top decks to display |
| `--sort-by` | string | overall | Sort criteria (overall, attack, defense, synergy) |
| `--format` | string | summary | Output format (summary, json, csv, detailed) |
| `--output-dir` | string | - | Directory to save results |
| `--verbose`, `-v` | bool | false | Show detailed progress |

### Genetic Algorithm Flags

#### Population and Generations

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ga-population` | int | 100 | Population size per generation (50-500 recommended) |
| `--ga-generations` | int | 100 | Maximum number of generations (50-500 recommended) |

**Guidelines:**
- Larger populations explore more diversity but take longer
- More generations allow more refinement but have diminishing returns
- Start with defaults, increase for better results or decrease for speed

#### Mutation and Crossover

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ga-mutation-rate` | float | 0.2 | Probability of mutation (0.0-1.0) |
| `--ga-crossover-rate` | float | 0.7 | Probability of crossover (0.0-1.0) |
| `--ga-mutation-intensity` | float | 0.3 | Mutation intensity: cards changed (0.0-1.0) |

**Guidelines:**
- Mutation rate 0.1-0.3: Typical range, higher for more exploration
- Crossover rate 0.6-0.9: High crossover exploits good solutions
- Mutation intensity 0.2-0.5: Higher values change more cards per mutation

#### Elite Preservation

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ga-elite-count` | int | 10 | Best decks to preserve per generation (1-20) |
| `--ga-tournament-size` | int | 5 | Tournament selection size (3-10) |

**Guidelines:**
- Elite count: 5-15% of population size recommended
- Tournament size: Larger values increase selection pressure

#### Early Stopping

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ga-convergence-generations` | int | 0 | Stop if no improvement for N generations (0=disabled) |
| `--ga-target-fitness` | float | 0.0 | Stop when fitness reaches this value (0=disabled) |

**Guidelines:**
- Convergence: 20-50 generations typical stopping threshold
- Target fitness: Set to 8.5-9.5 for high-quality decks

#### Island Model

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ga-island-model` | bool | false | Enable island model (parallel populations) |
| `--ga-island-count` | int | 4 | Number of islands (2-8 recommended) |
| `--ga-migration-interval` | int | 10 | Generations between migrations |
| `--ga-migration-size` | int | 5 | Decks to migrate per interval |

**Guidelines:**
- Islands maintain diversity and explore different solutions in parallel
- Migration interval: 5-20 generations
- Migration size: 1-10% of per-island population

#### Performance

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ga-parallel-eval` | bool | false | Enable parallel fitness evaluation |

**Guidelines:**
- Parallel evaluation speeds up fitness computation
- Most beneficial with large populations (100+)

## Configuration Details

### Population and Generations

The population size and generation count control the breadth and depth of the evolutionary search:

- **Small (Pop: 50, Gen: 30)**: Fast, good for initial exploration, may miss optimal solutions
- **Medium (Pop: 100, Gen: 100)**: Balanced, suitable for most use cases (default)
- **Large (Pop: 200, Gen: 300)**: Thorough, best for finding absolute optimal decks, slower

### Mutation and Crossover

The GA uses five mutation strategies, randomly selected during mutation:

1. **Single Card Swap**: Random card replacement
2. **Role-Based Swap**: Replace with same role card (preserves deck structure)
3. **Synergy-Guided Swap**: Replace with high-synergy card
4. **Evolution-Aware Swap**: Prioritize evolved cards (70% weight)
5. **Mixed Mutation**: Combines role, synergy, and evolution awareness

Crossover strategies (randomly selected):

1. **Uniform Crossover**: Random gene selection from each parent
2. **Role-Preserving Crossover**: Preserves role composition from parents
3. **Synergy-Aware Crossover**: Builds offspring from high-synergy pairs (threshold: 0.8)

### Elite Preservation

Elite preservation ensures the best solutions are never lost:

- **Elite Count**: Top N decks copied unchanged to next generation
- **Tournament Selection**: Randomly select K individuals, choose best for breeding

Higher elite counts provide stability but reduce exploration. Tournament size controls selection pressure - larger tournaments favor better individuals more strongly.

### Island Model

The island model runs multiple parallel populations (islands) that occasionally exchange solutions:

- **Islands**: Independent populations that evolve separately
- **Migration**: Best individuals from each island migrate to others periodically
- **Diversity**: Islands explore different areas of solution space

Benefits:
- Maintains diversity longer than single population
- Finds multiple high-quality solutions
- Can escape local optima more effectively

### Early Stopping

Early stopping prevents wasted computation when optimization has converged:

- **Convergence-based**: Stop if best fitness hasn't improved for N generations
- **Target-based**: Stop when fitness reaches specified threshold

Example: With `--ga-convergence-generations 25`, if the best deck doesn't improve for 25 consecutive generations, optimization stops early.

## Algorithm Details

### How Genetic Algorithms Work

1. **Initialization**: Create random population of valid 8-card decks
2. **Evaluation**: Score each deck using comprehensive evaluation system (fitness: 0-10 scale)
3. **Selection**: Choose parents for breeding using tournament selection
4. **Crossover**: Combine parent decks to create offspring (respects crossover rate)
5. **Mutation**: Randomly modify offspring decks (respects mutation rate and intensity)
6. **Elite Preservation**: Copy top decks to next generation unchanged
7. **Replacement**: New generation replaces old, repeat from step 2

### Fitness Function

The fitness function uses the same comprehensive evaluation system as deck building:

- **Attack** (0-10): Offensive power and win conditions
- **Defense** (0-10): Defensive capability and counters
- **Synergy** (0-10): Card interactions and combinations
- **Versatility** (0-10): Adaptability and consistency
- **F2P** (0-10): Card level and upgrade requirements

Overall score (0-10) is weighted average based on selected strategy (balanced, aggro, control, etc.).

### Repair Mechanisms

The GA automatically repairs invalid decks:

- **Duplicate Removal**: Ensures 8 unique cards
- **Win Condition Enforcement**: Guarantees at least one win condition card
- **Candidate Validation**: Only includes cards from player's collection

### Convergence Behavior

Typical convergence pattern:
- **Early Generations (1-20)**: Rapid fitness improvement as GA finds good building blocks
- **Middle Generations (20-60)**: Slower improvement as GA refines solutions
- **Late Generations (60+)**: Minimal improvement, population has converged

Use `--verbose` to see generation-by-generation progress.

## Examples

### Find Best Possible Decks

```bash
./bin/cr-api deck fuzz --mode genetic --tag R8QGUQRCV \
  --ga-population 200 \
  --ga-generations 300 \
  --ga-elite-count 20 \
  --ga-convergence-generations 30 \
  --top 20 \
  --verbose
```

Expected time: 2-5 minutes
Output: Top 20 best decks from extensive search

### Quick Deck Ideas

```bash
./bin/cr-api deck fuzz --mode genetic --tag R8QGUQRCV \
  --ga-population 50 \
  --ga-generations 30 \
  --top 5
```

Expected time: 20-40 seconds
Output: Top 5 decks from quick optimization

### Maximize Synergy

```bash
./bin/cr-api deck fuzz --mode genetic --tag R8QGUQRCV \
  --ga-population 100 \
  --ga-generations 150 \
  --sort-by synergy \
  --top 10
```

Expected time: 1-2 minutes
Output: Top 10 decks sorted by synergy score

### Island Model for Multiple Solutions

```bash
./bin/cr-api deck fuzz --mode genetic --tag R8QGUQRCV \
  --ga-island-model \
  --ga-island-count 4 \
  --ga-population 120 \
  --ga-generations 100 \
  --ga-migration-interval 15 \
  --top 20
```

Expected time: 2-4 minutes
Output: Top 20 decks from diverse island populations

### Save Results to JSON

```bash
./bin/cr-api deck fuzz --mode genetic --tag R8QGUQRCV \
  --ga-population 150 \
  --ga-generations 200 \
  --format json \
  --output-dir data/ga-results \
  --top 50
```

Output: JSON file with top 50 decks and full GA metadata

## Performance

### Timing

Approximate times for genetic optimization (depends on hardware and card pool size):

| Configuration | Time | Use Case |
|---------------|------|----------|
| Pop: 50, Gen: 30 | 20-40s | Quick exploration |
| Pop: 100, Gen: 100 | 1-2 min | Standard optimization (default) |
| Pop: 200, Gen: 300 | 3-6 min | Thorough search |
| Island model (4 islands) | +20-40% | Diverse solutions |

### Memory Usage

Memory usage scales with population size:
- Small (50): ~10 MB
- Medium (100): ~20 MB
- Large (200): ~40 MB

### Parallelization

Use `--ga-parallel-eval` to enable parallel fitness evaluation:
- Speeds up evaluation by 2-4x on multi-core systems
- Most beneficial with populations ≥100
- No impact on solution quality

## Troubleshooting

### Optimization Takes Too Long

**Solution**: Reduce population or generations:
```bash
--ga-population 50 --ga-generations 50
```

Or enable early stopping:
```bash
--ga-convergence-generations 20
```

### Solutions Not Improving

**Possible causes**:
1. **Already converged**: GA found local optimum
2. **Insufficient diversity**: Population too small or mutation rate too low

**Solutions**:
- Increase mutation rate: `--ga-mutation-rate 0.3`
- Use island model: `--ga-island-model`
- Increase population: `--ga-population 200`

### All Decks Look Similar

**Cause**: Over-convergence to single solution

**Solution**: Use island model for diversity:
```bash
--ga-island-model --ga-island-count 4
```

### Fitness Scores Lower Than Expected

**Possible causes**:
1. **Limited card pool**: Few cards restrict deck quality
2. **Early stopping**: Not enough generations to optimize
3. **Low mutation intensity**: Insufficient exploration

**Solutions**:
- Increase generations: `--ga-generations 200`
- Increase mutation intensity: `--ga-mutation-intensity 0.4`
- Check card collection has diverse roles and synergies

## Comparison with Monte Carlo Fuzzing

| Aspect | Genetic Algorithm | Monte Carlo Fuzzing |
|--------|------------------|---------------------|
| **Approach** | Evolutionary, iterative improvement | Random sampling |
| **Speed** | Slower (seconds to minutes) | Faster (sub-second to seconds) |
| **Quality** | Generally finds better solutions | May miss optimal solutions |
| **Diversity** | Converges to similar solutions | More diverse results |
| **Use Case** | Finding best possible deck | Quick exploration, many options |
| **Reproducibility** | Stochastic (varies with seed) | Fully random or seeded |

**When to use Genetic Algorithm:**
- You want the absolute best deck from your collection
- You have time for longer optimization (1-5 minutes)
- You're willing to trade speed for solution quality
- You want to see evolutionary improvement over time

**When to use Monte Carlo:**
- You want quick results (< 30 seconds)
- You want to see many diverse deck options
- You're exploring what's possible with specific constraints
- You want to generate large numbers of decks for analysis

## See Also

- [DECK_FUZZING.md](DECK_FUZZING.md) - Monte Carlo deck fuzzing
- [DECK_BUILDER.md](DECK_BUILDER.md) - Manual deck building
- [DECK_ANALYSIS_SUITE.md](DECK_ANALYSIS_SUITE.md) - Comprehensive deck evaluation
- [CLI_REFERENCE.md](CLI_REFERENCE.md) - All deck commands
