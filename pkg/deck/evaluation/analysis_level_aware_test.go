package evaluation

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestEvaluateLevelAwareBreakdownPresentWithPlayerContext(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()
	deckCards := testLevelAwareDeckCards()

	playerContext := &PlayerContext{
		Collection: map[string]CardLevelInfo{
			"Hog Rider": {Level: 13, MaxLevel: 15},
			"Musketeer": {Level: 13, MaxLevel: 15},
			"Fireball":  {Level: 12, MaxLevel: 15},
			"The Log":   {Level: 12, MaxLevel: 15},
			"Ice Spirit": {
				Level: 12, MaxLevel: 15,
			},
			"Skeletons": {Level: 12, MaxLevel: 15},
			"Cannon":    {Level: 13, MaxLevel: 15},
			"Ice Golem": {Level: 13, MaxLevel: 15},
		},
	}

	result := Evaluate(deckCards, synergyDB, playerContext)
	if result.OverallBreakdown == nil {
		t.Fatal("expected OverallBreakdown with player context")
	}

	bd := result.OverallBreakdown
	if bd.FinalScore != result.OverallScore {
		t.Fatalf("breakdown final score %.2f must equal overall %.2f", bd.FinalScore, result.OverallScore)
	}
	if bd.NormalizationFactor <= 0 || bd.NormalizationFactor > 1.10 {
		t.Fatalf("unexpected normalization factor %.3f", bd.NormalizationFactor)
	}
	if bd.DeckLevelRatio <= 0 || bd.DeckLevelRatio > 1.0 {
		t.Fatalf("unexpected deck level ratio %.3f", bd.DeckLevelRatio)
	}
}

func TestEvaluateLevelNormalizationDifferentiatesHighAndLowLevels(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()
	deckCards := testLevelAwareDeckCards()

	highContext := &PlayerContext{
		Collection: map[string]CardLevelInfo{
			"Hog Rider": {Level: 15, MaxLevel: 15},
			"Musketeer": {Level: 15, MaxLevel: 15},
			"Fireball":  {Level: 14, MaxLevel: 15},
			"The Log":   {Level: 14, MaxLevel: 15},
			"Ice Spirit": {
				Level: 14, MaxLevel: 15,
			},
			"Skeletons": {Level: 14, MaxLevel: 15},
			"Cannon":    {Level: 15, MaxLevel: 15},
			"Ice Golem": {Level: 15, MaxLevel: 15},
		},
	}
	lowContext := &PlayerContext{
		Collection: map[string]CardLevelInfo{
			"Hog Rider": {Level: 9, MaxLevel: 15},
			"Musketeer": {Level: 9, MaxLevel: 15},
			"Fireball":  {Level: 8, MaxLevel: 15},
			"The Log":   {Level: 8, MaxLevel: 15},
			"Ice Spirit": {
				Level: 8, MaxLevel: 15,
			},
			"Skeletons": {Level: 8, MaxLevel: 15},
			"Cannon":    {Level: 9, MaxLevel: 15},
			"Ice Golem": {Level: 9, MaxLevel: 15},
		},
	}

	highResult := Evaluate(deckCards, synergyDB, highContext)
	lowResult := Evaluate(deckCards, synergyDB, lowContext)

	if highResult.OverallBreakdown == nil || lowResult.OverallBreakdown == nil {
		t.Fatal("expected breakdown for both contextual evaluations")
	}
	if highResult.OverallBreakdown.NormalizationFactor <= lowResult.OverallBreakdown.NormalizationFactor {
		t.Fatalf("expected high-level normalization factor %.3f to exceed low-level factor %.3f",
			highResult.OverallBreakdown.NormalizationFactor, lowResult.OverallBreakdown.NormalizationFactor)
	}
	if highResult.OverallScore <= lowResult.OverallScore {
		t.Fatalf("expected higher-level deck overall score %.2f to exceed lower-level score %.2f",
			highResult.OverallScore, lowResult.OverallScore)
	}
}

func testLevelAwareDeckCards() []deck.CardCandidate {
	baseDeck := []struct {
		name   string
		elixir int
		rarity string
		role   deck.CardRole
	}{
		{name: "Hog Rider", elixir: 4, rarity: "Rare", role: deck.RoleWinCondition},
		{name: "Musketeer", elixir: 4, rarity: "Rare", role: deck.RoleSupport},
		{name: "Fireball", elixir: 4, rarity: "Rare", role: deck.RoleSpellBig},
		{name: "The Log", elixir: 2, rarity: "Legendary", role: deck.RoleSpellSmall},
		{name: "Ice Spirit", elixir: 1, rarity: "Common", role: deck.RoleCycle},
		{name: "Skeletons", elixir: 1, rarity: "Common", role: deck.RoleCycle},
		{name: "Cannon", elixir: 3, rarity: "Common", role: deck.RoleBuilding},
		{name: "Ice Golem", elixir: 2, rarity: "Rare", role: deck.RoleCycle},
	}

	cards := make([]deck.CardCandidate, 0, len(baseDeck))
	for _, item := range baseDeck {
		cards = append(cards, deck.CardCandidate{
			Name:     item.name,
			Elixir:   item.elixir,
			Level:    11,
			MaxLevel: 15,
			Rarity:   item.rarity,
			Role:     ptrRole(item.role),
		})
	}

	return cards
}
