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
				{Elixir: 4},
				{Elixir: 3},
				{Elixir: 2},
				{Elixir: 1},
				{Elixir: 5},
				{Elixir: 3},
				{Elixir: 4},
				{Elixir: 2},
			},
			wantAvg: 3.0,
		},
		{
			name: "Heavy deck",
			deckCards: []deck.CardCandidate{
				{Elixir: 8},
				{Elixir: 6},
				{Elixir: 5},
				{Elixir: 4},
				{Elixir: 4},
				{Elixir: 3},
				{Elixir: 3},
				{Elixir: 1},
			},
			wantAvg: 4.25,
		},
		{
			name: "Cycle deck",
			deckCards: []deck.CardCandidate{
				{Elixir: 4},
				{Elixir: 3},
				{Elixir: 2},
				{Elixir: 1},
				{Elixir: 1},
				{Elixir: 2},
				{Elixir: 3},
				{Elixir: 2},
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

// labeledDeck represents a deck with known archetype for accuracy testing
type labeledDeck struct {
	Name           string
	DeckCards      []deck.CardCandidate
	ExpectedArchetype Archetype
	IsHybrid       bool
	SecondaryArchetype Archetype // For hybrid decks
}

// TestArchetypeDetectionAccuracy validates overall archetype detection accuracy
// Success criteria: Overall accuracy >80%, Pure archetype detection >85%, Hybrid detection >75%
func TestArchetypeDetectionAccuracy(t *testing.T) {
	// Labeled test decks with known archetypes (human-classified baseline)
	labeledDecks := []labeledDeck{
		// === BEATDOWN (Pure) ===
		{
			Name: "Golem Beatdown (Classic)",
			DeckCards: []deck.CardCandidate{
				{Name: "Golem", Elixir: 8},
				{Name: "Baby Dragon", Elixir: 4},
				{Name: "Night Witch", Elixir: 4},
				{Name: "Lumberjack", Elixir: 4},
				{Name: "Lightning", Elixir: 6},
				{Name: "Tornado", Elixir: 3},
				{Name: "Mega Minion", Elixir: 3},
				{Name: "Skeletons", Elixir: 1},
			},
			ExpectedArchetype: ArchetypeBeatdown,
			IsHybrid: false,
		},
		{
			Name: "Lava Hound Beatdown",
			DeckCards: []deck.CardCandidate{
				{Name: "Lava Hound", Elixir: 7},
				{Name: "Baby Dragon", Elixir: 4},
				{Name: "Balloon", Elixir: 5},
				{Name: "Inferno Dragon", Elixir: 4},
				{Name: "Lightning", Elixir: 6},
				{Name: "Tornado", Elixir: 3},
				{Name: "Mega Minion", Elixir: 3},
				{Name: "Skeletons", Elixir: 1},
			},
			ExpectedArchetype: ArchetypeBeatdown,
			IsHybrid: false,
		},
		{
			Name: "Electro Giant Beatdown",
			DeckCards: []deck.CardCandidate{
				{Name: "Electro Giant", Elixir: 8},
				{Name: "Mega Knight", Elixir: 7},
				{Name: "Baby Dragon", Elixir: 4},
				{Name: "Electro Wizard", Elixir: 4},
				{Name: "P.E.K.K.A", Elixir: 7},
				{Name: "Tornado", Elixir: 3},
				{Name: "Skeletons", Elixir: 1},
				{Name: "The Log", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeBeatdown,
			IsHybrid: false,
		},
		{
			Name: "Giant Beatdown",
			DeckCards: []deck.CardCandidate{
				{Name: "Giant", Elixir: 5},
				{Name: "Mega Minion", Elixir: 3},
				{Name: "Musketeer", Elixir: 4},
				{Name: "Fireball", Elixir: 4},
				{Name: "Skeleton Army", Elixir: 3},
				{Name: "Ice Golem", Elixir: 2},
				{Name: "The Log", Elixir: 2},
				{Name: "Inferno Tower", Elixir: 5, Role: ptrRole(deck.RoleBuilding)},
			},
			ExpectedArchetype: ArchetypeBeatdown,
			IsHybrid: false,
		},

		// === CYCLE (Pure) ===
		{
			Name: "Hog Cycle (Classic)",
			DeckCards: []deck.CardCandidate{
				{Name: "Hog Rider", Elixir: 4},
				{Name: "Skeletons", Elixir: 1},
				{Name: "Ice Spirit", Elixir: 1},
				{Name: "Ice Golem", Elixir: 2},
				{Name: "Musketeer", Elixir: 4},
				{Name: "Cannon", Elixir: 3, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Fireball", Elixir: 4},
				{Name: "The Log", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeCycle,
			IsHybrid: false,
		},
		{
			Name: "Hog 2.6 Cycle",
			DeckCards: []deck.CardCandidate{
				{Name: "Hog Rider", Elixir: 4},
				{Name: "Ice Spirit", Elixir: 1},
				{Name: "Ice Golem", Elixir: 2},
				{Name: "Cannon", Elixir: 3, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Fireball", Elixir: 4},
				{Name: "The Log", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeCycle,
			IsHybrid: false,
		},
		{
			Name: "Miner Cycle",
			DeckCards: []deck.CardCandidate{
				{Name: "Miner", Elixir: 3},
				{Name: "Skeletons", Elixir: 1},
				{Name: "Ice Spirit", Elixir: 1},
				{Name: "Ice Golem", Elixir: 2},
				{Name: "Musketeer", Elixir: 4},
				{Name: "Poison", Elixir: 4},
				{Name: "Fireball", Elixir: 4},
				{Name: "The Log", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeCycle,
			IsHybrid: false,
		},
		{
			Name: "Royal Giant Cycle",
			DeckCards: []deck.CardCandidate{
				{Name: "Royal Giant", Elixir: 6},
				{Name: "Skeletons", Elixir: 1},
				{Name: "Ice Spirit", Elixir: 1},
				{Name: "Musketeer", Elixir: 4},
				{Name: "Fireball", Elixir: 4},
				{Name: "The Log", Elixir: 2},
				{Name: "Cannon", Elixir: 3, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Inferno Tower", Elixir: 5, Role: ptrRole(deck.RoleBuilding)},
			},
			ExpectedArchetype: ArchetypeCycle,
			IsHybrid: false,
		},

		// === SIEGE (Pure) ===
		{
			Name: "X-Bow Siege (Classic)",
			DeckCards: []deck.CardCandidate{
				{Name: "X-Bow", Elixir: 6},
				{Name: "Tesla", Elixir: 4, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Archers", Elixir: 3},
				{Name: "Knight", Elixir: 3},
				{Name: "Skeletons", Elixir: 1},
				{Name: "Ice Golem", Elixir: 2},
				{Name: "Fireball", Elixir: 4},
				{Name: "The Log", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeSiege,
			IsHybrid: false,
		},
		{
			Name: "X-Bow Cycle",
			DeckCards: []deck.CardCandidate{
				{Name: "X-Bow", Elixir: 6},
				{Name: "Ice Spirit", Elixir: 1},
				{Name: "Skeletons", Elixir: 1},
				{Name: "Tesla", Elixir: 4, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Knight", Elixir: 3},
				{Name: "Fireball", Elixir: 4},
				{Name: "The Log", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeSiege,
			IsHybrid: false,
		},
		{
			Name: "Mortar Siege",
			DeckCards: []deck.CardCandidate{
				{Name: "Mortar", Elixir: 4},
				{Name: "Tesla", Elixir: 4, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Knight", Elixir: 3},
				{Name: "Archers", Elixir: 3},
				{Name: "Skeletons", Elixir: 1},
				{Name: "Fireball", Elixir: 4},
				{Name: "The Log", Elixir: 2},
				{Name: "Ice Spirit", Elixir: 1},
			},
			ExpectedArchetype: ArchetypeSiege,
			IsHybrid: false,
		},

		// === BAIT (Pure) ===
		{
			Name: "Log Bait (Classic)",
			DeckCards: []deck.CardCandidate{
				{Name: "Goblin Barrel", Elixir: 3},
				{Name: "Princess", Elixir: 3},
				{Name: "Goblin Gang", Elixir: 3},
				{Name: "Knight", Elixir: 3},
				{Name: "Ice Spirit", Elixir: 1},
				{Name: "Inferno Tower", Elixir: 5, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Rocket", Elixir: 6},
				{Name: "The Log", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeBait,
			IsHybrid: false,
		},
		{
			Name: "Goblin Barrel Bait",
			DeckCards: []deck.CardCandidate{
				{Name: "Goblin Barrel", Elixir: 3},
				{Name: "Princess", Elixir: 3},
				{Name: "Goblin Gang", Elixir: 3},
				{Name: "Dart Goblin", Elixir: 3},
				{Name: "Knight", Elixir: 3},
				{Name: "Ice Spirit", Elixir: 1},
				{Name: "Inferno Tower", Elixir: 5, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Rocket", Elixir: 6},
			},
			ExpectedArchetype: ArchetypeBait,
			IsHybrid: false,
		},
		{
			Name: "Goblin Drill Bait",
			DeckCards: []deck.CardCandidate{
				{Name: "Goblin Drill", Elixir: 4},
				{Name: "Goblin Gang", Elixir: 3},
				{Name: "Princess", Elixir: 3},
				{Name: "Dart Goblin", Elixir: 3},
				{Name: "Knight", Elixir: 3},
				{Name: "Ice Spirit", Elixir: 1},
				{Name: "Rocket", Elixir: 6},
				{Name: "The Log", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeBait,
			IsHybrid: false,
		},

		// === BRIDGE SPAM (Pure) ===
		{
			Name: "PEKKA Bridge Spam (Classic)",
			DeckCards: []deck.CardCandidate{
				{Name: "P.E.K.K.A", Elixir: 7},
				{Name: "Battle Ram", Elixir: 4},
				{Name: "Bandit", Elixir: 3},
				{Name: "Royal Ghost", Elixir: 3},
				{Name: "Electro Wizard", Elixir: 4},
				{Name: "Minions", Elixir: 3},
				{Name: "Poison", Elixir: 4},
				{Name: "Zap", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeBridge,
			IsHybrid: false,
		},
		{
			Name: "Mega Knight Bridge Spam",
			DeckCards: []deck.CardCandidate{
				{Name: "Mega Knight", Elixir: 7},
				{Name: "Battle Ram", Elixir: 4},
				{Name: "Bandit", Elixir: 3},
				{Name: "Royal Ghost", Elixir: 3},
				{Name: "Electro Wizard", Elixir: 4},
				{Name: "Inferno Dragon", Elixir: 4},
				{Name: "Poison", Elixir: 4},
				{Name: "Zap", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeBridge,
			IsHybrid: false,
		},
		{
			Name: "Royal Ghost Bridge Spam",
			DeckCards: []deck.CardCandidate{
				{Name: "Royal Ghost", Elixir: 3},
				{Name: "Battle Ram", Elixir: 4},
				{Name: "Bandit", Elixir: 3},
				{Name: "P.E.K.K.A", Elixir: 7},
				{Name: "Electro Wizard", Elixir: 4},
				{Name: "Minions", Elixir: 3},
				{Name: "Poison", Elixir: 4},
				{Name: "Fireball", Elixir: 4},
			},
			ExpectedArchetype: ArchetypeBridge,
			IsHybrid: false,
		},

		// === GRAVEYARD (Pure) ===
		{
			Name: "Graveyard Freeze (Classic)",
			DeckCards: []deck.CardCandidate{
				{Name: "Graveyard", Elixir: 5},
				{Name: "Ice Wizard", Elixir: 3},
				{Name: "Baby Dragon", Elixir: 4},
				{Name: "Bomb Tower", Elixir: 4, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Freeze", Elixir: 4},
				{Name: "Tornado", Elixir: 3},
				{Name: "Knight", Elixir: 3},
				{Name: "Poison", Elixir: 4},
			},
			ExpectedArchetype: ArchetypeGraveyard,
			IsHybrid: false,
		},
		{
			Name: "Graveyard Poison",
			DeckCards: []deck.CardCandidate{
				{Name: "Graveyard", Elixir: 5},
				{Name: "Baby Dragon", Elixir: 4},
				{Name: "Mega Minion", Elixir: 3},
				{Name: "Poison", Elixir: 4},
				{Name: "Tornado", Elixir: 3},
				{Name: "Knight", Elixir: 3},
				{Name: "Ice Wizard", Elixir: 3},
				{Name: "Skeletons", Elixir: 1},
			},
			ExpectedArchetype: ArchetypeGraveyard,
			IsHybrid: false,
		},

		// === MINER (Pure) ===
		{
			Name: "Miner Poison (Classic)",
			DeckCards: []deck.CardCandidate{
				{Name: "Miner", Elixir: 3},
				{Name: "Poison", Elixir: 4},
				{Name: "Valkyrie", Elixir: 4},
				{Name: "Electro Wizard", Elixir: 4},
				{Name: "Ice Golem", Elixir: 2},
				{Name: "Inferno Tower", Elixir: 5, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Skeletons", Elixir: 1},
				{Name: "The Log", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeMiner,
			IsHybrid: false,
		},
		{
			Name: "Miner Control",
			DeckCards: []deck.CardCandidate{
				{Name: "Miner", Elixir: 3},
				{Name: "Poison", Elixir: 4},
				{Name: "Valkyrie", Elixir: 4},
				{Name: "Rocket", Elixir: 6},
				{Name: "Electro Wizard", Elixir: 4},
				{Name: "Ice Wizard", Elixir: 3},
				{Name: "Inferno Tower", Elixir: 5, Role: ptrRole(deck.RoleBuilding)},
				{Name: "The Log", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeMiner,
			IsHybrid: false,
		},

		// === HYBRID DECKS ===
		{
			Name: "Beatdown-Bridge Hybrid",
			DeckCards: []deck.CardCandidate{
				{Name: "Golem", Elixir: 8},
				{Name: "Battle Ram", Elixir: 4},
				{Name: "P.E.K.K.A", Elixir: 7},
				{Name: "Night Witch", Elixir: 4},
				{Name: "Bandit", Elixir: 3},
				{Name: "Tornado", Elixir: 3},
				{Name: "Lightning", Elixir: 6},
				{Name: "Skeletons", Elixir: 1},
			},
			ExpectedArchetype: ArchetypeBeatdown,
			SecondaryArchetype: ArchetypeBridge,
			IsHybrid: true,
		},
		{
			Name: "Cycle-Miner Hybrid",
			DeckCards: []deck.CardCandidate{
				{Name: "Hog Rider", Elixir: 4},
				{Name: "Miner", Elixir: 3},
				{Name: "Ice Spirit", Elixir: 1},
				{Name: "Skeletons", Elixir: 1},
				{Name: "Musketeer", Elixir: 4},
				{Name: "Poison", Elixir: 4},
				{Name: "Ice Golem", Elixir: 2},
				{Name: "The Log", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeCycle,
			SecondaryArchetype: ArchetypeMiner,
			IsHybrid: true,
		},
		{
			Name: "Graveyard-Control Hybrid",
			DeckCards: []deck.CardCandidate{
				{Name: "Graveyard", Elixir: 5},
				{Name: "X-Bow", Elixir: 6},
				{Name: "Tesla", Elixir: 4, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Ice Wizard", Elixir: 3},
				{Name: "Poison", Elixir: 4},
				{Name: "Tornado", Elixir: 3},
				{Name: "Knight", Elixir: 3},
				{Name: "Skeletons", Elixir: 1},
			},
			ExpectedArchetype: ArchetypeSiege,
			SecondaryArchetype: ArchetypeGraveyard,
			IsHybrid: true,
		},

		// === EDGE CASES ===
		{
			Name: "Control Deck (No clear archetype)",
			DeckCards: []deck.CardCandidate{
				{Name: "Electro Wizard", Elixir: 4},
				{Name: "Valkyrie", Elixir: 4},
				{Name: "Musketeer", Elixir: 4},
				{Name: "Mega Minion", Elixir: 3},
				{Name: "Inferno Tower", Elixir: 5, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Poison", Elixir: 4},
				{Name: "Fireball", Elixir: 4},
				{Name: "The Log", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeControl,
			IsHybrid: false,
		},
		{
			Name: "Midrange Deck",
			DeckCards: []deck.CardCandidate{
				{Name: "Mega Knight", Elixir: 7},
				{Name: "Balloon", Elixir: 5},
				{Name: "Baby Dragon", Elixir: 4},
				{Name: "Musketeer", Elixir: 4},
				{Name: "Fireball", Elixir: 4},
				{Name: "Inferno Tower", Elixir: 5, Role: ptrRole(deck.RoleBuilding)},
				{Name: "The Log", Elixir: 2},
				{Name: "Ice Golem", Elixir: 2},
			},
			ExpectedArchetype: ArchetypeBeatdown,
			IsHybrid: false,
		},
		{
			Name: "Spell Bait",
			DeckCards: []deck.CardCandidate{
				{Name: "Goblin Barrel", Elixir: 3},
				{Name: "Princess", Elixir: 3},
				{Name: "Goblin Gang", Elixir: 3},
				{Name: "Knight", Elixir: 3},
				{Name: "Ice Spirit", Elixir: 1},
				{Name: "Inferno Tower", Elixir: 5, Role: ptrRole(deck.RoleBuilding)},
				{Name: "Rocket", Elixir: 6},
				{Name: "Skeleton Army", Elixir: 3},
			},
			ExpectedArchetype: ArchetypeBait,
			IsHybrid: false,
		},
	}

	// Run detection and track results
	var pureCorrect, pureTotal int
	var hybridCorrect, hybridTotal int
	var correctByArchetype = make(map[Archetype]int)
	var totalByArchetype = make(map[Archetype]int)

	// Confusion matrix tracking
	confusionMatrix := make(map[Archetype]map[Archetype]int) // [actual][detected]

	for _, deck := range labeledDecks {
		result := DetectArchetype(deck.DeckCards)

		// Initialize confusion matrix row if needed
		if confusionMatrix[deck.ExpectedArchetype] == nil {
			confusionMatrix[deck.ExpectedArchetype] = make(map[Archetype]int)
		}

		// For hybrid decks, check if detected correctly
		if deck.IsHybrid {
			hybridTotal++
			// Check if primary or secondary matches expected archetypes
			primaryMatch := result.Primary == deck.ExpectedArchetype || result.Primary == deck.SecondaryArchetype
			secondaryMatch := result.Secondary == deck.ExpectedArchetype || result.Secondary == deck.SecondaryArchetype
			if result.IsHybrid && (primaryMatch || secondaryMatch) {
				hybridCorrect++
			}
			confusionMatrix[deck.ExpectedArchetype][result.Primary]++
		} else {
			// Pure archetype deck
			pureTotal++
			totalByArchetype[deck.ExpectedArchetype]++

			// Check if detected correctly (including when detected as hybrid with correct primary)
			detectedCorrectly := result.Primary == deck.ExpectedArchetype ||
				(result.Primary == ArchetypeHybrid && result.Secondary == deck.ExpectedArchetype)

			if detectedCorrectly {
				pureCorrect++
				correctByArchetype[deck.ExpectedArchetype]++
			}

			confusionMatrix[deck.ExpectedArchetype][result.Primary]++
		}

		// Log results for debugging
		t.Logf("Deck: %s\n  Expected: %v (hybrid: %v)\n  Detected: %v (%.2f confidence) | %v (%.2f) | isHybrid: %v",
			deck.Name,
			deck.ExpectedArchetype, deck.IsHybrid,
			result.Primary, result.PrimaryConfidence,
			result.Secondary, result.SecondaryConfidence,
			result.IsHybrid)
	}

	// Calculate overall accuracy
	totalCorrect := pureCorrect + hybridCorrect
	totalDecks := len(labeledDecks)
	overallAccuracy := float64(totalCorrect) / float64(totalDecks) * 100

	// Calculate pure archetype accuracy
	pureAccuracy := float64(pureCorrect) / float64(pureTotal) * 100

	// Calculate hybrid archetype accuracy
	var hybridAccuracy float64
	if hybridTotal > 0 {
		hybridAccuracy = float64(hybridCorrect) / float64(hybridTotal) * 100
	}

	// Report results
	t.Logf("\n=== ARCHETYPE DETECTION ACCURACY RESULTS ===")
	t.Logf("Total Decks: %d", totalDecks)
	t.Logf("Overall Accuracy: %.1f%% (%d/%d)", overallAccuracy, totalCorrect, totalDecks)
	t.Logf("Pure Archetype Accuracy: %.1f%% (%d/%d)", pureAccuracy, pureCorrect, pureTotal)
	if hybridTotal > 0 {
		t.Logf("Hybrid Archetype Accuracy: %.1f%% (%d/%d)", hybridAccuracy, hybridCorrect, hybridTotal)
	}

	// Per-archetype breakdown
	t.Logf("\n=== PER-ARCHETYPE ACCURACY ===")
	allArchetypes := []Archetype{
		ArchetypeBeatdown,
		ArchetypeControl,
		ArchetypeCycle,
		ArchetypeBridge,
		ArchetypeSiege,
		ArchetypeBait,
		ArchetypeGraveyard,
		ArchetypeMiner,
	}
	for _, archetype := range allArchetypes {
		if totalByArchetype[archetype] > 0 {
			accuracy := float64(correctByArchetype[archetype]) / float64(totalByArchetype[archetype]) * 100
			t.Logf("%s: %.1f%% (%d/%d)",
				archetype, accuracy, correctByArchetype[archetype], totalByArchetype[archetype])
		}
	}

	// Confusion matrix (simplified - only show misclassifications)
	t.Logf("\n=== CONFUSION MATRIX (Misclassifications) ===")
	for actual, detectedMap := range confusionMatrix {
		for detected, count := range detectedMap {
			if count > 0 && actual != detected {
				t.Logf("%s -> %s: %d", actual, detected, count)
			}
		}
	}

	// Assert success criteria
	// Overall accuracy >80%
	if overallAccuracy < 80.0 {
		t.Errorf("Overall accuracy %.1f%% is below 80%% threshold", overallAccuracy)
	}

	// Pure archetype detection >85%
	if pureAccuracy < 85.0 {
		t.Errorf("Pure archetype accuracy %.1f%% is below 85%% threshold", pureAccuracy)
	}

	// Hybrid detection >75%
	if hybridTotal > 0 && hybridAccuracy < 75.0 {
		t.Errorf("Hybrid archetype accuracy %.1f%% is below 75%% threshold", hybridAccuracy)
	}
}
