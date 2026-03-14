package evaluation

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestSharedArchetypeSignalScoring(t *testing.T) {
	t.Run("beatdown core win condition contributes score", func(t *testing.T) {
		score := scoreBeatdown([]deck.CardCandidate{
			{Name: "Golem", Elixir: 8},
			{Name: "Baby Dragon", Elixir: 4},
			{Name: "Night Witch", Elixir: 4},
			{Name: "Tornado", Elixir: 3},
			{Name: "Barbarian Barrel", Elixir: 2},
			{Name: "Mega Minion", Elixir: 3},
			{Name: "Arrows", Elixir: 3},
			{Name: "Lightning", Elixir: 6},
		})
		if score <= 5.0 {
			t.Fatalf("expected beatdown score > 5.0, got %.2f", score)
		}
	})

	t.Run("siege core win condition contributes score", func(t *testing.T) {
		score := scoreSiege([]deck.CardCandidate{
			{Name: "X-Bow", Elixir: 6},
			{Name: "Tesla", Elixir: 4},
			{Name: "Knight", Elixir: 3},
			{Name: "Archers", Elixir: 3},
			{Name: "Skeletons", Elixir: 1},
			{Name: "Ice Spirit", Elixir: 1},
			{Name: "Fireball", Elixir: 4},
			{Name: "The Log", Elixir: 2},
		})
		if score <= 6.0 {
			t.Fatalf("expected siege score > 6.0, got %.2f", score)
		}
	})

	t.Run("bridge spam core win condition contributes score", func(t *testing.T) {
		score := scoreBridgeSpam([]deck.CardCandidate{
			{Name: "Battle Ram", Elixir: 4},
			{Name: "Bandit", Elixir: 3},
			{Name: "Royal Ghost", Elixir: 3},
			{Name: "Electro Wizard", Elixir: 4},
			{Name: "Magic Archer", Elixir: 4},
			{Name: "Poison", Elixir: 4},
			{Name: "Zap", Elixir: 2},
			{Name: "Dark Prince", Elixir: 4},
		})
		if score <= 6.0 {
			t.Fatalf("expected bridge spam score > 6.0, got %.2f", score)
		}
	})
}
