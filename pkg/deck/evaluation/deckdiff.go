package evaluation

import (
	"sort"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// CategoryDelta captures the score delta for a single evaluation category.
type CategoryDelta struct {
	Current     float64 `json:"current"`
	Recommended float64 `json:"recommended"`
	Delta       float64 `json:"delta"`
}

// DeckDiff holds the result of comparing a current deck against a recommended deck.
type DeckDiff struct {
	CurrentScore         float64                  `json:"current_score"`
	RecommendedScore     float64                  `json:"recommended_score"`
	ScoreDelta           float64                  `json:"score_delta"`
	CategoryDeltas       map[string]CategoryDelta `json:"category_deltas"`
	CurrentArchetype     Archetype                `json:"current_archetype"`
	RecommendedArchetype Archetype                `json:"recommended_archetype"`
	// CurrentLevelFit is DeckLevelRatio from the evaluation breakdown (0.0–1.0).
	CurrentLevelFit     float64  `json:"current_level_fit"`
	RecommendedLevelFit float64  `json:"recommended_level_fit"`
	SharedCards         []string `json:"shared_cards"`
	ReplacedCards       []string `json:"replaced_cards"`
}

// DiffDecks evaluates both decks under the same player context and returns the diff.
// currentCards and recommendedCards must carry level/evolution data so that
// evaluation.Evaluate can produce accurate level-fit scoring.
func DiffDecks(
	player *clashroyale.Player,
	currentCards, recommendedCards []deck.CardCandidate,
	synergyDB *deck.SynergyDatabase,
) (*DeckDiff, error) {
	playerCtx := NewPlayerContextFromPlayer(player)

	currentResult := Evaluate(currentCards, synergyDB, playerCtx)
	recommendedResult := Evaluate(recommendedCards, synergyDB, playerCtx)

	currentNames := cardNameSet(currentCards)
	recommendedNames := cardNameSet(recommendedCards)

	return &DeckDiff{
		CurrentScore:         currentResult.OverallScore,
		RecommendedScore:     recommendedResult.OverallScore,
		ScoreDelta:           recommendedResult.OverallScore - currentResult.OverallScore,
		CategoryDeltas:       buildCategoryDeltas(currentResult, recommendedResult),
		CurrentArchetype:     currentResult.DetectedArchetype,
		RecommendedArchetype: recommendedResult.DetectedArchetype,
		CurrentLevelFit:      levelFit(currentResult, playerCtx, currentCards),
		RecommendedLevelFit:  levelFit(recommendedResult, playerCtx, recommendedCards),
		SharedCards:          sortedIntersection(currentNames, recommendedNames),
		ReplacedCards:        sortedDifference(recommendedNames, currentNames),
	}, nil
}

const (
	categoryAttack      = "attack"
	categoryDefense     = "defense"
	categorySynergy     = "synergy"
	categoryVersatility = "versatility"
	categoryF2P         = "f2p_friendly"
	categoryPlayability = "playability"
)

func buildCategoryDeltas(cur, rec EvaluationResult) map[string]CategoryDelta {
	mk := func(c, r CategoryScore) CategoryDelta {
		return CategoryDelta{Current: c.Score, Recommended: r.Score, Delta: r.Score - c.Score}
	}
	return map[string]CategoryDelta{
		categoryAttack:      mk(cur.Attack, rec.Attack),
		categoryDefense:     mk(cur.Defense, rec.Defense),
		categorySynergy:     mk(cur.Synergy, rec.Synergy),
		categoryVersatility: mk(cur.Versatility, rec.Versatility),
		categoryF2P:         mk(cur.F2PFriendly, rec.F2PFriendly),
		categoryPlayability: mk(cur.Playability, rec.Playability),
	}
}

func levelFit(result EvaluationResult, ctx *PlayerContext, cards []deck.CardCandidate) float64 {
	if result.OverallBreakdown != nil {
		return result.OverallBreakdown.DeckLevelRatio
	}
	if ctx == nil || len(cards) == 0 {
		return 0
	}
	names := make([]string, len(cards))
	for i, c := range cards {
		names[i] = c.Name
	}
	return ctx.GetAverageLevel(names)
}

func cardNameSet(cards []deck.CardCandidate) map[string]struct{} {
	s := make(map[string]struct{}, len(cards))
	for _, c := range cards {
		s[c.Name] = struct{}{}
	}
	return s
}

func sortedIntersection(a, b map[string]struct{}) []string {
	var result []string
	for name := range a {
		if _, ok := b[name]; ok {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}

func sortedDifference(a, b map[string]struct{}) []string {
	var result []string
	for name := range a {
		if _, ok := b[name]; ok {
			continue
		}
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}
