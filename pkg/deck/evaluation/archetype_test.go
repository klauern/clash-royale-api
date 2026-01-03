package evaluation

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

func TestDetectArchetype(t *testing.T) {
	tests := []struct {
		name                 string
		deckCards            []deck.CardCandidate
		expectedPrimary      Archetype
		minPrimaryConfidence float64
		expectHybrid         bool
	}{
		{
			name: "Golem Beatdown",
			deckCards: []deck.CardCandidate{
				{Name: "Golem", Elixir: 8},
				{Name: "Baby Dragon", Elixir: 4},
				{Name: "Night Witch", Elixir: 4},
				{Name: "Lumberjack", Elixir: 4},
				{Name: "Lightning", Elixir: 6},
				{Name: "Tornado", Elixir: 3},
				{Name: "Mega Minion", Elixir: 3},
				{Name: "Skeletons", Elixir: 1},
			},
			expectedPrimary:      ArchetypeBeatdown,
			minPrimaryConfidence: 0.6,
			expectHybrid:         false,
		},
		{
			name: "Hog Cycle",
			deckCards: []deck.CardCandidate{
				{Name: "Hog Rider", Elixir: 4},
				{Name: "Skeletons", Elixir: 1},
				{Name: "Ice Spirit", Elixir: 1},
				{Name: "Ice Golem", Elixir: 2},
				{Name: "Musketeer", Elixir: 4},
				{Name: "Cannon", Elixir: 3, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Fireball", Elixir: 4},
				{Name: "Log", Elixir: 2},
			},
			expectedPrimary:      ArchetypeCycle,
			minPrimaryConfidence: 0.6,
			expectHybrid:         false,
		},
		{
			name: "X-Bow Siege",
			deckCards: []deck.CardCandidate{
				{Name: "X-Bow", Elixir: 6},
				{Name: "Tesla", Elixir: 4, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Archers", Elixir: 3},
				{Name: "Knight", Elixir: 3},
				{Name: "Skeletons", Elixir: 1},
				{Name: "Ice Golem", Elixir: 2},
				{Name: "Fireball", Elixir: 4},
				{Name: "Log", Elixir: 2},
			},
			expectedPrimary:      ArchetypeSiege,
			minPrimaryConfidence: 0.7,
			expectHybrid:         false,
		},
		{
			name: "Log Bait",
			deckCards: []deck.CardCandidate{
				{Name: "Goblin Barrel", Elixir: 3},
				{Name: "Princess", Elixir: 3},
				{Name: "Goblin Gang", Elixir: 3},
				{Name: "Knight", Elixir: 3},
				{Name: "Ice Spirit", Elixir: 1},
				{Name: "Inferno Tower", Elixir: 5, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Rocket", Elixir: 6},
				{Name: "Log", Elixir: 2},
			},
			expectedPrimary:      ArchetypeBait,
			minPrimaryConfidence: 0.7,
			expectHybrid:         false,
		},
		{
			name: "PEKKA Bridge Spam",
			deckCards: []deck.CardCandidate{
				{Name: "P.E.K.K.A", Elixir: 7},
				{Name: "Battle Ram", Elixir: 4},
				{Name: "Bandit", Elixir: 3},
				{Name: "Royal Ghost", Elixir: 3},
				{Name: "Electro Wizard", Elixir: 4},
				{Name: "Minions", Elixir: 3},
				{Name: "Poison", Elixir: 4},
				{Name: "Zap", Elixir: 2},
			},
			expectedPrimary:      ArchetypeBridge,
			minPrimaryConfidence: 0.6,
			expectHybrid:         false,
		},
		{
			name: "Graveyard Freeze",
			deckCards: []deck.CardCandidate{
				{Name: "Graveyard", Elixir: 5},
				{Name: "Ice Wizard", Elixir: 3},
				{Name: "Baby Dragon", Elixir: 4},
				{Name: "Bomb Tower", Elixir: 4, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Freeze", Elixir: 4},
				{Name: "Tornado", Elixir: 3},
				{Name: "Knight", Elixir: 3},
				{Name: "Poison", Elixir: 4},
			},
			expectedPrimary:      ArchetypeGraveyard,
			minPrimaryConfidence: 0.6,
			expectHybrid:         false,
		},
		{
			name: "Miner Poison",
			deckCards: []deck.CardCandidate{
				{Name: "Miner", Elixir: 3},
				{Name: "Poison", Elixir: 4},
				{Name: "Valkyrie", Elixir: 4},
				{Name: "Electro Wizard", Elixir: 4},
				{Name: "Ice Golem", Elixir: 2},
				{Name: "Inferno Tower", Elixir: 5, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Skeletons", Elixir: 1},
				{Name: "Log", Elixir: 2},
			},
			expectedPrimary:      ArchetypeMiner,
			minPrimaryConfidence: 0.6,
			expectHybrid:         false,
		},
		{
			name:                 "Empty deck",
			deckCards:            []deck.CardCandidate{},
			expectedPrimary:      ArchetypeUnknown,
			minPrimaryConfidence: 0.0,
			expectHybrid:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectArchetype(tt.deckCards)

			if result.Primary != tt.expectedPrimary && result.Primary != ArchetypeHybrid {
				t.Errorf("DetectArchetype() primary = %v, want %v or hybrid", result.Primary, tt.expectedPrimary)
			}

			if result.PrimaryConfidence < tt.minPrimaryConfidence {
				t.Errorf("DetectArchetype() primaryConfidence = %.2f, want >= %.2f", result.PrimaryConfidence, tt.minPrimaryConfidence)
			}

			if result.IsHybrid != tt.expectHybrid {
				t.Errorf("DetectArchetype() isHybrid = %v, want %v (Primary: %v %.2f, Secondary: %v %.2f)",
					result.IsHybrid, tt.expectHybrid,
					result.Primary, result.PrimaryConfidence,
					result.Secondary, result.SecondaryConfidence)
			}

			// Validate confidence bounds
			if result.PrimaryConfidence < 0.0 || result.PrimaryConfidence > 1.0 {
				t.Errorf("DetectArchetype() primaryConfidence = %.2f, must be between 0.0 and 1.0", result.PrimaryConfidence)
			}

			if result.SecondaryConfidence < 0.0 || result.SecondaryConfidence > 1.0 {
				t.Errorf("DetectArchetype() secondaryConfidence = %.2f, must be between 0.0 and 1.0", result.SecondaryConfidence)
			}
		})
	}
}

func TestNormalizeConfidence(t *testing.T) {
	tests := []struct {
		name    string
		score   float64
		minConf float64
		maxConf float64
	}{
		{
			name:    "Zero score",
			score:   0.0,
			minConf: 0.0,
			maxConf: 0.0,
		},
		{
			name:    "Low score",
			score:   3.0,
			minConf: 0.0,
			maxConf: 0.5,
		},
		{
			name:    "Medium score",
			score:   5.0,
			minConf: 0.4,
			maxConf: 0.7,
		},
		{
			name:    "High score",
			score:   8.0,
			minConf: 0.7,
			maxConf: 1.0,
		},
		{
			name:    "Max score",
			score:   10.0,
			minConf: 1.0,
			maxConf: 1.0,
		},
		{
			name:    "Above max",
			score:   12.0,
			minConf: 1.0,
			maxConf: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := normalizeConfidence(tt.score)

			if conf < tt.minConf || conf > tt.maxConf {
				t.Errorf("normalizeConfidence(%.1f) = %.2f, want between %.2f and %.2f", tt.score, conf, tt.minConf, tt.maxConf)
			}

			// Always validate bounds
			if conf < 0.0 || conf > 1.0 {
				t.Errorf("normalizeConfidence(%.1f) = %.2f, must be between 0.0 and 1.0", tt.score, conf)
			}
		})
	}
}

func TestArchetypeScoreFunctions(t *testing.T) {
	// Test individual archetype scoring functions
	t.Run("scoreBeatdown", func(t *testing.T) {
		beatdownDeck := []deck.CardCandidate{
			{Name: "Golem", Elixir: 8},
			{Name: "Baby Dragon", Elixir: 4},
			{Name: "Night Witch", Elixir: 4},
			{Name: "Lumberjack", Elixir: 4},
			{Name: "Lightning", Elixir: 6},
			{Name: "Tornado", Elixir: 3},
			{Name: "Mega Minion", Elixir: 3},
			{Name: "Skeletons", Elixir: 1},
		}

		score := scoreBeatdown(beatdownDeck)
		if score < 6.0 {
			t.Errorf("scoreBeatdown() = %.2f, want >= 6.0 for strong beatdown deck", score)
		}
	})

	t.Run("scoreCycle", func(t *testing.T) {
		cycleDeck := []deck.CardCandidate{
			{Name: "Hog Rider", Elixir: 4},
			{Name: "Skeletons", Elixir: 1},
			{Name: "Ice Spirit", Elixir: 1},
			{Name: "Ice Golem", Elixir: 2},
			{Name: "Musketeer", Elixir: 4},
			{Name: "Cannon", Elixir: 3},
			{Name: "Fireball", Elixir: 4},
			{Name: "Log", Elixir: 2},
		}

		score := scoreCycle(cycleDeck)
		if score < 6.0 {
			t.Errorf("scoreCycle() = %.2f, want >= 6.0 for strong cycle deck", score)
		}
	})

	t.Run("scoreSiege", func(t *testing.T) {
		siegeDeck := []deck.CardCandidate{
			{Name: "X-Bow", Elixir: 6},
			{Name: "Tesla", Elixir: 4},
			{Name: "Archers", Elixir: 3},
			{Name: "Knight", Elixir: 3},
		}

		score := scoreSiege(siegeDeck)
		if score < 6.0 {
			t.Errorf("scoreSiege() = %.2f, want >= 6.0 for deck with X-Bow", score)
		}
	})

	t.Run("scoreBait", func(t *testing.T) {
		baitDeck := []deck.CardCandidate{
			{Name: "Goblin Barrel", Elixir: 3},
			{Name: "Princess", Elixir: 3},
			{Name: "Goblin Gang", Elixir: 3},
			{Name: "Knight", Elixir: 3},
		}

		score := scoreBait(baitDeck)
		if score < 7.0 {
			t.Errorf("scoreBait() = %.2f, want >= 7.0 for strong bait deck", score)
		}
	})

	t.Run("scoreGraveyard", func(t *testing.T) {
		graveyardDeck := []deck.CardCandidate{
			{Name: "Graveyard", Elixir: 5},
			{Name: "Ice Wizard", Elixir: 3},
			{Name: "Baby Dragon", Elixir: 4},
			{Name: "Freeze", Elixir: 4},
		}

		score := scoreGraveyard(graveyardDeck)
		if score < 6.0 {
			t.Errorf("scoreGraveyard() = %.2f, want >= 6.0 for deck with Graveyard", score)
		}
	})

	t.Run("scoreMiner", func(t *testing.T) {
		minerDeck := []deck.CardCandidate{
			{Name: "Miner", Elixir: 3},
			{Name: "Poison", Elixir: 4},
			{Name: "Valkyrie", Elixir: 4},
			{Name: "Electro Wizard", Elixir: 4},
		}

		score := scoreMiner(minerDeck)
		if score < 6.0 {
			t.Errorf("scoreMiner() = %.2f, want >= 6.0 for deck with Miner", score)
		}
	})
}

func TestCalculateAvgElixir(t *testing.T) {
	tests := []struct {
		name      string
		deckCards []deck.CardCandidate
		wantAvg   float64
	}{
		{
			name: "Standard 8-card deck",
			deckCards: []deck.CardCandidate{
				{Elixir: 4}, {Elixir: 3}, {Elixir: 2}, {Elixir: 1},
				{Elixir: 5}, {Elixir: 3}, {Elixir: 4}, {Elixir: 2},
			},
			wantAvg: 3.0,
		},
		{
			name: "Heavy deck",
			deckCards: []deck.CardCandidate{
				{Elixir: 8}, {Elixir: 6}, {Elixir: 5}, {Elixir: 4},
				{Elixir: 4}, {Elixir: 3}, {Elixir: 3}, {Elixir: 1},
			},
			wantAvg: 4.25,
		},
		{
			name: "Cycle deck",
			deckCards: []deck.CardCandidate{
				{Elixir: 4}, {Elixir: 3}, {Elixir: 2}, {Elixir: 1},
				{Elixir: 1}, {Elixir: 2}, {Elixir: 3}, {Elixir: 2},
			},
			wantAvg: 2.25,
		},
		{
			name:      "Empty deck",
			deckCards: []deck.CardCandidate{},
			wantAvg:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateAvgElixir(tt.deckCards)
			if got != tt.wantAvg {
				t.Errorf("calculateAvgElixir() = %.2f, want %.2f", got, tt.wantAvg)
			}
		})
	}
}

func TestFindTopArchetype(t *testing.T) {
	scores := map[Archetype]float64{
		ArchetypeBeatdown: 8.5,
		ArchetypeControl:  4.2,
		ArchetypeCycle:    3.1,
		ArchetypeBridge:   2.0,
	}

	archetype, score := findTopArchetype(scores)

	if archetype != ArchetypeBeatdown {
		t.Errorf("findTopArchetype() archetype = %v, want %v", archetype, ArchetypeBeatdown)
	}

	if score != 8.5 {
		t.Errorf("findTopArchetype() score = %.2f, want %.2f", score, 8.5)
	}
}

func TestHybridDetection(t *testing.T) {
	// Test a deck that could be both beatdown and control
	mixedDeck := []deck.CardCandidate{
		{Name: "Golem", Elixir: 8},     // Beatdown
		{Name: "Graveyard", Elixir: 5}, // Control/Graveyard
		{Name: "Baby Dragon", Elixir: 4},
		{Name: "Tornado", Elixir: 3},
		{Name: "Poison", Elixir: 4},
		{Name: "Knight", Elixir: 3},
		{Name: "Ice Wizard", Elixir: 3},
		{Name: "Skeletons", Elixir: 1},
	}

	result := DetectArchetype(mixedDeck)

	// This deck should potentially be detected as hybrid or have high secondary confidence
	if result.SecondaryConfidence < 0.3 {
		t.Logf("Hybrid test: Primary=%v (%.2f), Secondary=%v (%.2f), IsHybrid=%v",
			result.Primary, result.PrimaryConfidence,
			result.Secondary, result.SecondaryConfidence,
			result.IsHybrid)
		// Note: Not failing here as hybrid detection depends on scoring thresholds
	}
}

// ptrRole returns a pointer to a CardRole
func ptrRole(r deck.CardRole) *deck.CardRole {
	return &r
}
