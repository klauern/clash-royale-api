package research

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/klauer/clash-royale-api/go/pkg/deck/genetic"
)

func namesToCandidates(cardNames []string, cards []deck.CardCandidate) []deck.CardCandidate {
	byName := make(map[string]deck.CardCandidate, len(cards))
	for _, c := range cards {
		byName[c.Name] = c
	}
	out := make([]deck.CardCandidate, 0, len(cardNames))
	for _, name := range cardNames {
		if c, ok := byName[name]; ok {
			out = append(out, c)
		}
	}
	return out
}

func toDeckResult(method string, selected []deck.CardCandidate, m DeckMetrics) DeckResult {
	deckNames := make([]string, 0, len(selected))
	for _, c := range selected {
		deckNames = append(deckNames, c.Name)
	}
	arch := evaluation.DetectArchetype(selected)
	return DeckResult{
		Method:              method,
		Deck:                deckNames,
		Metrics:             m,
		DetectedArchetype:   arch.Primary.String(),
		ArchetypeConfidence: arch.PrimaryConfidence,
	}
}

func runWithTiming(method string, selected []deck.CardCandidate, synergyDB *deck.SynergyDatabase, constraints ConstraintConfig, started time.Time) DeckResult {
	metrics := ScoreDeckComposite(selected, synergyDB, constraints)
	metrics.RuntimeMs = time.Since(started).Milliseconds()
	return toDeckResult(method, selected, metrics)
}

// BaselineMethod wraps the current deck builder behavior.
type BaselineMethod struct {
	Builder *deck.Builder
}

func (m BaselineMethod) Name() string { return MethodBaseline }

func (m BaselineMethod) Build(cards []deck.CardCandidate, cfg MethodConfig) (DeckResult, error) {
	if m.Builder == nil {
		return DeckResult{}, fmt.Errorf("baseline builder is nil")
	}
	analysis := deck.CardAnalysis{CardLevels: make(map[string]deck.CardLevelData, len(cards))}
	for _, c := range cards {
		analysis.CardLevels[c.Name] = deck.CardLevelData{
			Level:             c.Level,
			MaxLevel:          c.MaxLevel,
			Rarity:            c.Rarity,
			Elixir:            c.Elixir,
			EvolutionLevel:    c.EvolutionLevel,
			MaxEvolutionLevel: c.MaxEvolutionLevel,
		}
	}

	started := time.Now()
	constraints, cfgErr := resolveConstraintConfig(cfg.Constraints)
	if cfgErr != nil {
		return DeckResult{}, fmt.Errorf("invalid constraints: %w", cfgErr)
	}
	rec, err := m.Builder.BuildDeckFromAnalysis(analysis)
	if err != nil {
		return DeckResult{}, err
	}
	selected := namesToCandidates(rec.Deck, cards)
	synergyDB := deck.NewSynergyDatabase()
	return runWithTiming(m.Name(), selected, synergyDB, constraints, started), nil
}

// GeneticMethod runs archetype-free genetic optimization.
type GeneticMethod struct{}

func (m GeneticMethod) Name() string { return MethodGenetic }

func (m GeneticMethod) Build(cards []deck.CardCandidate, cfg MethodConfig) (DeckResult, error) {
	started := time.Now()
	constraints, err := resolveConstraintConfig(cfg.Constraints)
	if err != nil {
		return DeckResult{}, fmt.Errorf("invalid constraints: %w", err)
	}
	synergyDB := deck.NewSynergyDatabase()
	candidates := make([]*deck.CardCandidate, 0, len(cards))
	for i := range cards {
		card := cards[i]
		candidates = append(candidates, &card)
	}
	gaCfg := genetic.DefaultGeneticConfig()
	gaCfg.UseArchetypes = false
	gaCfg.PopulationSize = 60
	gaCfg.Generations = 80
	gaCfg.ConvergenceGenerations = 20
	gaCfg.EliteCount = 2

	opt, err := genetic.NewGeneticOptimizer(candidates, deck.StrategyBalanced, &gaCfg)
	if err != nil {
		return DeckResult{}, err
	}
	opt.FitnessFunc = func(deckCards []deck.CardCandidate) (float64, error) {
		return ScoreDeckComposite(deckCards, synergyDB, constraints).Composite, nil
	}
	if cfg.Seed != 0 {
		opt.RNG = rand.New(rand.NewSource(cfg.Seed))
	}

	result, err := opt.Optimize()
	if err != nil {
		return DeckResult{}, err
	}
	if len(result.HallOfFame) == 0 {
		return DeckResult{}, fmt.Errorf("genetic optimization produced no hall-of-fame deck")
	}
	selected := namesToCandidates(result.HallOfFame[0].Cards, cards)
	return runWithTiming(m.Name(), selected, synergyDB, constraints, started), nil
}

// ConstraintMethod uses constructive search + local improvement.
type ConstraintMethod struct{}

func (m ConstraintMethod) Name() string { return MethodConstraint }

func (m ConstraintMethod) Build(cards []deck.CardCandidate, cfg MethodConfig) (DeckResult, error) {
	if len(cards) < 8 {
		return DeckResult{}, fmt.Errorf("need at least 8 cards, got %d", len(cards))
	}
	constraints, err := resolveConstraintConfig(cfg.Constraints)
	if err != nil {
		return DeckResult{}, fmt.Errorf("invalid constraints: %w", err)
	}
	started := time.Now()
	synergyDB := deck.NewSynergyDatabase()
	rng := rand.New(rand.NewSource(cfg.Seed + 17))

	best := make([]deck.CardCandidate, 0, 8)
	bestScore := -1.0

	iterations := 300
	if cfg.TopN > 1 {
		iterations += cfg.TopN * 40
	}

	failures := 0
	for i := 0; i < iterations; i++ {
		candidate, buildErr := constructConstraintDeck(cards, constraints, rng)
		if buildErr != nil {
			failures++
			continue
		}
		candidate = improveDeckLocally(candidate, cards, synergyDB, constraints, rng, 80)
		score := ScoreDeckComposite(candidate, synergyDB, constraints)
		if score.Composite > bestScore {
			bestScore = score.Composite
			best = candidate
		}
	}

	if len(best) == 0 {
		return DeckResult{}, fmt.Errorf("constraint search could not generate valid deck after %d attempts (%d failed constructions); relax hard minima or increase candidate pool", iterations, failures)
	}
	return runWithTiming(m.Name(), best, synergyDB, constraints, started), nil
}

//nolint:gocognit,gocyclo // Constraint construction keeps hard-requirement steps explicit.
func constructConstraintDeck(pool []deck.CardCandidate, constraints ConstraintConfig, rng *rand.Rand) ([]deck.CardCandidate, error) {
	used := make(map[string]bool)
	selected := make([]deck.CardCandidate, 0, 8)
	hard := constraints.Hard

	pickAny := func(filter func(deck.CardCandidate) bool, min int) {
		for j := 0; j < min; j++ {
			idxs := make([]int, 0)
			for i, c := range pool {
				if used[c.Name] {
					continue
				}
				if filter(c) {
					idxs = append(idxs, i)
				}
			}
			if len(idxs) == 0 {
				return
			}
			pick := idxs[rng.Intn(len(idxs))]
			selected = append(selected, pool[pick])
			used[pool[pick].Name] = true
		}
	}

	pickAny(func(c deck.CardCandidate) bool { return c.Role != nil && *c.Role == deck.RoleWinCondition }, hard.MinWinConditions)
	pickAny(func(c deck.CardCandidate) bool {
		return c.Role != nil && (*c.Role == deck.RoleSpellBig || *c.Role == deck.RoleSpellSmall)
	}, hard.MinSpells)
	pickAny(canTargetAir, hard.MinAirDefense)
	pickAny(isTankKiller, hard.MinTankKillers)

	for len(selected) < 8 {
		idx := rng.Intn(len(pool))
		if used[pool[idx].Name] {
			continue
		}
		used[pool[idx].Name] = true
		selected = append(selected, pool[idx])
	}

	report := ValidateConstraints(selected, constraints)
	if !report.IsValid() {
		return nil, fmt.Errorf("generated deck failed constraints")
	}
	return selected, nil
}

func improveDeckLocally(deckIn, pool []deck.CardCandidate, synergyDB *deck.SynergyDatabase, constraints ConstraintConfig, rng *rand.Rand, steps int) []deck.CardCandidate {
	best := make([]deck.CardCandidate, len(deckIn))
	copy(best, deckIn)
	bestScore := ScoreDeckComposite(best, synergyDB, constraints).Composite

	for i := 0; i < steps; i++ {
		mut := make([]deck.CardCandidate, len(best))
		copy(mut, best)
		replaceIdx := rng.Intn(len(mut))
		candidate := pool[rng.Intn(len(pool))]
		exists := false
		for _, c := range mut {
			if c.Name == candidate.Name {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		mut[replaceIdx] = candidate
		if !ValidateConstraints(mut, constraints).IsValid() {
			continue
		}
		s := ScoreDeckComposite(mut, synergyDB, constraints).Composite
		if s > bestScore {
			bestScore = s
			best = mut
		}
	}
	return best
}

// RoleFirstMethod fills slots by explicit roles.
type RoleFirstMethod struct{}

func (m RoleFirstMethod) Name() string { return MethodRoleFirst }

func (m RoleFirstMethod) Build(cards []deck.CardCandidate, cfg MethodConfig) (DeckResult, error) {
	if len(cards) < 8 {
		return DeckResult{}, fmt.Errorf("need at least 8 cards, got %d", len(cards))
	}
	constraints, err := resolveConstraintConfig(cfg.Constraints)
	if err != nil {
		return DeckResult{}, fmt.Errorf("invalid constraints: %w", err)
	}
	started := time.Now()
	synergyDB := deck.NewSynergyDatabase()
	rng := rand.New(rand.NewSource(cfg.Seed + 91))

	selected := make([]deck.CardCandidate, 0, 8)
	used := make(map[string]bool)

	slots := []func(deck.CardCandidate) bool{
		func(c deck.CardCandidate) bool { return c.Role != nil && *c.Role == deck.RoleWinCondition },
		isTankKiller,
		isSplash,
		canTargetAir,
		func(c deck.CardCandidate) bool { return c.Role != nil && *c.Role == deck.RoleSpellBig },
		func(c deck.CardCandidate) bool { return c.Role != nil && *c.Role == deck.RoleSpellSmall },
		nil,
		nil,
	}

	for _, slotFilter := range slots {
		best, ok := pickBestForSlot(cards, selected, used, slotFilter, synergyDB, constraints, rng)
		if !ok {
			return DeckResult{}, fmt.Errorf("unable to fill role-first slot")
		}
		selected = append(selected, best)
		used[best.Name] = true
	}

	return runWithTiming(m.Name(), selected, synergyDB, constraints, started), nil
}

//nolint:gocognit,gocyclo // Slot scoring combines multiple weighted factors intentionally.
func pickBestForSlot(pool, selected []deck.CardCandidate, used map[string]bool, filter func(deck.CardCandidate) bool, synergyDB *deck.SynergyDatabase, constraints ConstraintConfig, rng *rand.Rand) (deck.CardCandidate, bool) {
	type scored struct {
		card  deck.CardCandidate
		score float64
	}
	candidates := make([]scored, 0)
	for _, c := range pool {
		if used[c.Name] {
			continue
		}
		if filter != nil && !filter(c) {
			continue
		}
		testDeck := make([]deck.CardCandidate, 0, len(selected)+1)
		testDeck = append(testDeck, selected...)
		testDeck = append(testDeck, c)

		incSynergy := ScoreDeckComposite(testDeck, synergyDB, constraints).Synergy
		coverageGain := coverageScore(testDeck)
		baseQuality := c.LevelRatio()
		elixirPenalty := 0.0
		if c.Elixir > 5 {
			elixirPenalty = 0.04 * float64(c.Elixir-5)
		}
		roleUniqPenalty := 0.0
		if c.Role != nil {
			roleCount := 0
			for _, s := range selected {
				if s.Role != nil && *s.Role == *c.Role {
					roleCount++
				}
			}
			roleUniqPenalty = float64(roleCount) * 0.03
		}
		score := baseQuality + (incSynergy * 0.35) + (coverageGain * 0.35) - elixirPenalty - roleUniqPenalty
		candidates = append(candidates, scored{card: c, score: score})
	}

	if len(candidates) == 0 && filter != nil {
		return pickBestForSlot(pool, selected, used, nil, synergyDB, constraints, rng)
	}
	if len(candidates) == 0 {
		return deck.CardCandidate{}, false
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
	top := candidates[0:1]
	if len(candidates) > 1 && (candidates[0].score-candidates[1].score) < 0.03 {
		top = candidates[0:2]
	}
	pick := top[rng.Intn(len(top))]
	return pick.card, true
}

//nolint:ireturn // Factory intentionally returns behavior via Method interface.
func methodByName(name string, builder *deck.Builder) (Method, error) {
	switch name {
	case MethodBaseline:
		return BaselineMethod{Builder: builder}, nil
	case MethodGenetic:
		return GeneticMethod{}, nil
	case MethodConstraint:
		return ConstraintMethod{}, nil
	case MethodRoleFirst:
		return RoleFirstMethod{}, nil
	default:
		return nil, fmt.Errorf("unknown method: %s", name)
	}
}

// BuildCandidatesFromPlayer converts a player collection to scored candidates.
func BuildCandidatesFromPlayer(player *clashroyale.Player, stats *clashroyale.CardStatsRegistry) []deck.CardCandidate {
	out := make([]deck.CardCandidate, 0, len(player.Cards))
	for _, card := range player.Cards {
		name := card.Name
		role := config.GetCardRoleWithEvolution(name, card.EvolutionLevel)
		candidate := deck.CardCandidate{
			Name:              name,
			Level:             card.Level,
			MaxLevel:          card.MaxLevel,
			Rarity:            card.Rarity,
			Elixir:            config.GetCardElixir(name, card.ElixirCost),
			Role:              &role,
			EvolutionLevel:    card.EvolutionLevel,
			MaxEvolutionLevel: card.MaxEvolutionLevel,
		}
		if stats != nil {
			candidate.Stats = stats.GetStats(name)
		}
		out = append(out, candidate)
	}
	return out
}
