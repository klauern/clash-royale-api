// Package genetic provides genetic algorithm-based deck optimization
// using the eaopt library for evolutionary deck generation.
package genetic

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/MaxHalford/eaopt"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// GeneticProgress captures progress metrics emitted during evolution.
type GeneticProgress struct {
	Generation  uint
	BestFitness float64
	AvgFitness  float64
	Populations int
}

// GeneticResult captures the final outputs of a genetic optimization run.
type GeneticResult struct {
	HallOfFame  []*DeckGenome
	Scores      []float64
	Generations uint
	Duration    time.Duration
}

// GeneticOptimizer orchestrates genetic algorithm runs for deck optimization.
type GeneticOptimizer struct {
	Config     *GeneticConfig
	Candidates []*deck.CardCandidate
	Strategy   deck.Strategy
	Progress   func(GeneticProgress)
	RNG        *rand.Rand
	// FitnessFunc overrides default genome fitness evaluation when set.
	FitnessFunc func([]deck.CardCandidate) (float64, error)
}

// NewGeneticOptimizer constructs a genetic optimizer with validation.
func NewGeneticOptimizer(candidates []*deck.CardCandidate, strategy deck.Strategy, config *GeneticConfig) (*GeneticOptimizer, error) {
	if len(candidates) < 8 {
		return nil, fmt.Errorf("insufficient candidates: need at least 8 cards, got %d", len(candidates))
	}
	if config == nil {
		cfg := DefaultGeneticConfig()
		config = &cfg
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &GeneticOptimizer{
		Config:     config,
		Candidates: candidates,
		Strategy:   strategy,
	}, nil
}

// Optimize runs the genetic algorithm and returns the hall of fame decks.
func (o *GeneticOptimizer) Optimize() (*GeneticResult, error) {
	if o == nil {
		return nil, fmt.Errorf("optimizer is nil")
	}
	if o.Config == nil {
		return nil, fmt.Errorf("optimizer config is nil")
	}
	if err := o.Config.Validate(); err != nil {
		return nil, err
	}
	if len(o.Candidates) < 8 {
		return nil, fmt.Errorf("insufficient candidates: need at least 8 cards, got %d", len(o.Candidates))
	}

	rng := o.RNG
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	popSize, nPops := o.populationConfig()
	hofSize := uint(1)
	if o.Config.EliteCount > 1 {
		hofSize = uint(o.Config.EliteCount)
	}

	model := elitismModel{
		Selector:  eaopt.SelTournament{NContestants: uint(o.Config.TournamentSize)},
		Elite:     uint(o.Config.EliteCount),
		MutRate:   o.Config.MutationRate,
		CrossRate: o.Config.CrossoverRate,
	}

	var (
		bestScore          = math.Inf(-1)
		lastImprovementGen uint
	)

	gaConfig := eaopt.GAConfig{
		NPops:        nPops,
		PopSize:      popSize,
		NGenerations: uint(o.Config.Generations),
		HofSize:      hofSize,
		Model:        model,
		ParallelEval: o.Config.ParallelEvaluations,
		RNG:          rng,
		Callback: func(ga *eaopt.GA) {
			if o.Progress == nil || ga == nil {
				return
			}
			best, avg := aggregateFitness(ga)
			o.Progress(GeneticProgress{
				Generation:  ga.Generations,
				BestFitness: best,
				AvgFitness:  avg,
				Populations: len(ga.Populations),
			})
		},
		EarlyStop: func(ga *eaopt.GA) bool {
			if ga == nil || len(ga.HallOfFame) == 0 {
				return false
			}
			currentBest := -ga.HallOfFame[0].Fitness
			if currentBest > bestScore {
				bestScore = currentBest
				lastImprovementGen = ga.Generations
			}

			if o.Config.TargetFitness > 0 && currentBest >= o.Config.TargetFitness {
				return true
			}
			if o.Config.ConvergenceGenerations > 0 {
				if ga.Generations >= lastImprovementGen+uint(o.Config.ConvergenceGenerations) {
					return true
				}
			}
			return false
		},
	}

	if o.Config.IslandModel {
		gaConfig.Migrator = eaopt.MigRing{NMigrants: uint(o.Config.MigrationSize)}
		gaConfig.MigFrequency = uint(o.Config.MigrationInterval)
	}

	ga, err := gaConfig.NewGA()
	if err != nil {
		return nil, err
	}

	newGenome := o.genomeFactory()
	if err := ga.Minimize(newGenome); err != nil {
		return nil, err
	}

	hallOfFame, scores := extractHallOfFame(ga)

	return &GeneticResult{
		HallOfFame:  hallOfFame,
		Scores:      scores,
		Generations: ga.Generations,
		Duration:    ga.Age,
	}, nil
}

func (o *GeneticOptimizer) populationConfig() (uint, uint) {
	if o.Config.IslandModel && o.Config.IslandCount > 0 {
		perPop := o.Config.PopulationSize / o.Config.IslandCount
		if perPop < 1 {
			perPop = 1
		}
		return uint(perPop), uint(o.Config.IslandCount)
	}
	return uint(o.Config.PopulationSize), 1
}

func (o *GeneticOptimizer) genomeFactory() func(rng *rand.Rand) eaopt.Genome {
	seeds := o.Config.SeedPopulation
	seedIndex := 0
	return func(rng *rand.Rand) eaopt.Genome {
		if seedIndex < len(seeds) {
			cards := seeds[seedIndex]
			seedIndex++
			if genome, err := NewDeckGenomeFromCards(cards, o.Candidates, o.Strategy, o.Config); err == nil {
				genome.fitnessEvaluator = o.FitnessFunc
				return &eaoptDeckGenome{genome: genome}
			}
		}

		genome, err := NewDeckGenome(o.Candidates, o.Strategy, o.Config)
		if err != nil {
			return &eaoptDeckGenome{genome: &DeckGenome{
				Cards:            []string{},
				Fitness:          0,
				config:           o.Config,
				candidates:       o.Candidates,
				strategy:         o.Strategy,
				fitnessEvaluator: o.FitnessFunc,
			}}
		}
		genome.fitnessEvaluator = o.FitnessFunc
		return &eaoptDeckGenome{genome: genome}
	}
}

type eaoptDeckGenome struct {
	genome *DeckGenome
}

func (g *eaoptDeckGenome) Evaluate() (float64, error) {
	if g == nil || g.genome == nil {
		return 0, fmt.Errorf("nil genome")
	}
	fitness, err := g.genome.Evaluate()
	if err != nil {
		return 0, err
	}
	return -fitness, nil
}

func (g *eaoptDeckGenome) Mutate(rng *rand.Rand) {
	if g == nil || g.genome == nil {
		return
	}
	_ = g.genome.Mutate()
}

func (g *eaoptDeckGenome) Crossover(genome eaopt.Genome, rng *rand.Rand) {
	if g == nil || g.genome == nil {
		return
	}
	other, ok := genome.(*eaoptDeckGenome)
	if !ok || other == nil || other.genome == nil {
		return
	}
	child, err := g.genome.Crossover(other.genome)
	if err != nil {
		return
	}
	if deckChild, ok := child.(*DeckGenome); ok {
		g.genome = deckChild
	}
}

func (g *eaoptDeckGenome) Clone() eaopt.Genome {
	if g == nil || g.genome == nil {
		return &eaoptDeckGenome{genome: nil}
	}
	clone := g.genome.Clone()
	if deckClone, ok := clone.(*DeckGenome); ok {
		return &eaoptDeckGenome{genome: deckClone}
	}
	return &eaoptDeckGenome{genome: nil}
}

type elitismModel struct {
	Selector  eaopt.Selector
	Elite     uint
	MutRate   float64
	CrossRate float64
}

func (mod elitismModel) Apply(pop *eaopt.Population) error {
	if pop == nil || len(pop.Individuals) == 0 {
		return nil
	}

	if mod.Elite > uint(len(pop.Individuals)) {
		mod.Elite = uint(len(pop.Individuals))
	}

	pop.Individuals.SortByFitness()

	var elites eaopt.Individuals
	if mod.Elite > 0 {
		elites = pop.Individuals[:mod.Elite].Clone(pop.RNG)
	}

	offspringCount := uint(len(pop.Individuals)) - mod.Elite
	if offspringCount > 0 {
		offsprings, err := generateOffsprings(offspringCount, pop.Individuals, mod.Selector, mod.CrossRate, pop.RNG)
		if err != nil {
			return err
		}
		if mod.MutRate > 0 {
			offsprings.Mutate(mod.MutRate, pop.RNG)
		}
		copy(pop.Individuals, elites)
		copy(pop.Individuals[mod.Elite:], offsprings)
		return nil
	}

	copy(pop.Individuals, elites)
	return nil
}

func (mod elitismModel) Validate() error {
	if mod.Selector == nil {
		return fmt.Errorf("selector cannot be nil")
	}
	if err := mod.Selector.Validate(); err != nil {
		return err
	}
	if mod.MutRate < 0 || mod.MutRate > 1 {
		return fmt.Errorf("mutation rate must be between 0 and 1, got %f", mod.MutRate)
	}
	if mod.CrossRate < 0 || mod.CrossRate > 1 {
		return fmt.Errorf("crossover rate must be between 0 and 1, got %f", mod.CrossRate)
	}
	return nil
}

func generateOffsprings(n uint, indis eaopt.Individuals, sel eaopt.Selector, crossRate float64,
	rng *rand.Rand,
) (eaopt.Individuals, error) {
	offsprings := make(eaopt.Individuals, n)
	i := 0
	for i < len(offsprings) {
		selected, _, err := sel.Apply(2, indis, rng)
		if err != nil {
			return nil, err
		}
		if rng.Float64() < crossRate {
			selected[0].Crossover(selected[1], rng)
		}
		if i < len(offsprings) {
			offsprings[i] = selected[0]
			i++
		}
		if i < len(offsprings) {
			offsprings[i] = selected[1]
			i++
		}
	}
	return offsprings, nil
}

func aggregateFitness(ga *eaopt.GA) (float64, float64) {
	if ga == nil || len(ga.Populations) == 0 {
		return 0, 0
	}

	best := -ga.HallOfFame[0].Fitness
	sum := 0.0
	count := 0
	for _, pop := range ga.Populations {
		sum += -pop.Individuals.FitAvg()
		count++
	}
	if count == 0 {
		return best, 0
	}
	return best, sum / float64(count)
}

func extractHallOfFame(ga *eaopt.GA) ([]*DeckGenome, []float64) {
	if ga == nil {
		return nil, nil
	}

	hall := make([]*DeckGenome, 0, len(ga.HallOfFame))
	scores := make([]float64, 0, len(ga.HallOfFame))
	for _, indi := range ga.HallOfFame {
		wrapped, ok := indi.Genome.(*eaoptDeckGenome)
		if !ok || wrapped == nil || wrapped.genome == nil {
			continue
		}
		clone := wrapped.genome.Clone()
		deckClone, ok := clone.(*DeckGenome)
		if !ok {
			continue
		}
		hall = append(hall, deckClone)
		scores = append(scores, -indi.Fitness)
	}
	return hall, scores
}
