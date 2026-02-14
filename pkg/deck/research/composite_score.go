package research

import (
	"math"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

const (
	weightSynergy     = 0.30
	weightCoverage    = 0.25
	weightRoleFit     = 0.20
	weightElixirFit   = 0.15
	weightCardQuality = 0.10
)

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func canTargetAir(c deck.CardCandidate) bool {
	return c.Stats != nil && (c.Stats.Targets == "Air" || c.Stats.Targets == "Air & Ground")
}

func isTankKiller(c deck.CardCandidate) bool {
	if c.Stats != nil && c.Stats.DamagePerSecond >= 150 {
		return true
	}
	return c.Name == "Inferno Dragon" || c.Name == "Inferno Tower" || c.Name == "Mini P.E.K.K.A" || c.Name == "P.E.K.K.A"
}

func isSplash(c deck.CardCandidate) bool {
	if c.Stats != nil && c.Stats.Radius > 0 {
		return true
	}
	return c.Name == "Baby Dragon" || c.Name == "Wizard" || c.Name == "Executioner" || c.Name == "Bowler"
}

func avgElixir(cards []deck.CardCandidate) float64 {
	if len(cards) == 0 {
		return 0
	}
	total := 0
	for _, c := range cards {
		total += c.Elixir
	}
	return float64(total) / float64(len(cards))
}

//nolint:gocyclo // Coverage components are intentionally explicit and weighted.
func coverageScore(cards []deck.CardCandidate) float64 {
	if len(cards) == 0 {
		return 0
	}
	winCons, spells, air, tanks, splash := 0, 0, 0, 0, 0
	for _, c := range cards {
		if c.Role != nil && *c.Role == deck.RoleWinCondition {
			winCons++
		}
		if c.Role != nil && (*c.Role == deck.RoleSpellBig || *c.Role == deck.RoleSpellSmall) {
			spells++
		}
		if canTargetAir(c) {
			air++
		}
		if isTankKiller(c) {
			tanks++
		}
		if isSplash(c) {
			splash++
		}
	}

	winComponent := clamp01(float64(winCons) / 1.0)
	spellComponent := clamp01(float64(spells) / 1.0)
	airComponent := clamp01(float64(air) / 2.0)
	tankComponent := clamp01(float64(tanks) / 1.0)
	splashComponent := clamp01(float64(splash) / 1.0)

	return (winComponent * 0.20) + (spellComponent * 0.20) + (airComponent * 0.25) + (tankComponent * 0.20) + (splashComponent * 0.15)
}

func roleFitScore(cards []deck.CardCandidate) float64 {
	if len(cards) == 0 {
		return 0
	}
	counts := map[deck.CardRole]int{}
	for _, c := range cards {
		if c.Role != nil {
			counts[*c.Role]++
		}
	}

	// Target profile: 1 wincon, 1-2 spells, 2 support/cycle mix, >=1 air answer via role or stats.
	winFit := 1.0 - math.Abs(float64(counts[deck.RoleWinCondition]-1))*0.5
	spellCount := counts[deck.RoleSpellBig] + counts[deck.RoleSpellSmall]
	spellFit := 1.0
	if spellCount < 1 {
		spellFit = 0
	} else if spellCount > 3 {
		spellFit = 0.5
	}
	supportFit := clamp01(float64(counts[deck.RoleSupport]+counts[deck.RoleCycle]) / 3.0)

	airCount := 0
	for _, c := range cards {
		if canTargetAir(c) {
			airCount++
		}
	}
	airFit := clamp01(float64(airCount) / 2.0)

	return clamp01(winFit*0.30 + spellFit*0.25 + supportFit*0.25 + airFit*0.20)
}

func elixirFitScore(cards []deck.CardCandidate) float64 {
	if len(cards) == 0 {
		return 0
	}
	avg := avgElixir(cards)
	// Target 3.3, score drops to 0 at +/-2.0.
	return clamp01(1.0 - (math.Abs(avg-3.3) / 2.0))
}

func cardQualityScore(cards []deck.CardCandidate) float64 {
	if len(cards) == 0 {
		return 0
	}
	total := 0.0
	for _, c := range cards {
		total += c.LevelRatio()
	}
	return clamp01(total / float64(len(cards)))
}

func synergyScore(cards []deck.CardCandidate, synergyDB *deck.SynergyDatabase) float64 {
	if len(cards) == 0 || synergyDB == nil {
		return 0
	}
	names := make([]string, 0, len(cards))
	for _, c := range cards {
		names = append(names, c.Name)
	}
	analysis := synergyDB.AnalyzeDeckSynergy(names)
	if analysis == nil {
		return 0
	}
	pairCount := 0
	for _, n := range analysis.CategoryScores {
		pairCount += n
	}
	maxPairs := (len(cards) * (len(cards) - 1)) / 2
	coverage := 0.0
	if maxPairs > 0 {
		coverage = float64(pairCount) / float64(maxPairs)
	}
	return clamp01((analysis.AverageScore * 0.75) + (coverage * 0.25))
}

// ScoreDeckComposite returns the archetype-free metric breakdown.
func ScoreDeckComposite(cards []deck.CardCandidate, synergyDB *deck.SynergyDatabase) DeckMetrics {
	synergy := synergyScore(cards, synergyDB)
	coverage := coverageScore(cards)
	roleFit := roleFitScore(cards)
	elixirFit := elixirFitScore(cards)
	quality := cardQualityScore(cards)
	composite := (weightSynergy * synergy) +
		(weightCoverage * coverage) +
		(weightRoleFit * roleFit) +
		(weightElixirFit * elixirFit) +
		(weightCardQuality * quality)

	report := ValidateConstraints(cards)

	return DeckMetrics{
		Composite:            clamp01(composite),
		Synergy:              synergy,
		Coverage:             coverage,
		RoleFit:              roleFit,
		ElixirFit:            elixirFit,
		CardQuality:          quality,
		ConstraintViolations: report.Violations,
	}
}
