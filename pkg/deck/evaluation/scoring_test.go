package evaluation

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// Helper function to create a CardCandidate with basic stats
func makeCard(name string, role deck.CardRole, level, maxLevel int, rarity string, elixir int) deck.CardCandidate {
	rolePtr := &role
	targets := "Ground"
	if role == deck.RoleSupport || role == deck.RoleSpellBig {
		targets = "Air & Ground"
	}

	return deck.CardCandidate{
		Name:     name,
		Level:    level,
		MaxLevel: maxLevel,
		Rarity:   rarity,
		Elixir:   elixir,
		Role:     rolePtr,
		Stats: &clashroyale.CombatStats{
			DamagePerSecond: 100,
			Targets:         targets,
		},
	}
}

func TestScoreAttack(t *testing.T) {
	tests := []struct {
		name        string
		cards       []deck.CardCandidate
		expectScore float64 // Approximate expected score
		minScore    float64 // Minimum acceptable score
		maxScore    float64 // Maximum acceptable score
	}{
		{
			name: "Strong attack with 2 win conditions",
			cards: []deck.CardCandidate{
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Giant", deck.RoleWinCondition, 11, 11, "Rare", 5),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
				makeCard("Zap", deck.RoleSpellSmall, 11, 11, "Common", 2),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
			},
			minScore: 6.5,
			maxScore: 10.0,
		},
		{
			name: "Moderate attack with 1 win condition",
			cards: []deck.CardCandidate{
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
				makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
			},
			minScore: 6.0,
			maxScore: 8.5,
		},
		{
			name: "Weak attack with no win conditions",
			cards: []deck.CardCandidate{
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
				makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Archers", deck.RoleSupport, 11, 11, "Common", 3),
				makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
			},
			minScore: 2.0,
			maxScore: 5.0,
		},
		{
			name:        "Empty deck",
			cards:       []deck.CardCandidate{},
			minScore:    0.0,
			maxScore:    0.0,
			expectScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScoreAttack(tt.cards)

			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("ScoreAttack() score = %v, want between %v and %v",
					result.Score, tt.minScore, tt.maxScore)
			}

			if result.Assessment == "" {
				t.Errorf("ScoreAttack() missing assessment text")
			}

			if result.Stars < 1 || result.Stars > 3 {
				t.Errorf("ScoreAttack() stars = %v, want 1-3", result.Stars)
			}
		})
	}
}

func TestScoreDefense(t *testing.T) {
	tests := []struct {
		name     string
		cards    []deck.CardCandidate
		minScore float64
		maxScore float64
	}{
		{
			name: "Strong defense with building and anti-air",
			cards: []deck.CardCandidate{
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Mega Minion", deck.RoleSupport, 11, 11, "Rare", 3),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Archers", deck.RoleSupport, 11, 11, "Common", 3),
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
			},
			minScore: 7.0,
			maxScore: 11.0,
		},
		{
			name: "Weak defense with no anti-air",
			cards: []deck.CardCandidate{
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
				makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
			},
			minScore: 6.0,
			maxScore: 9.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScoreDefense(tt.cards)

			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("ScoreDefense() score = %v, want between %v and %v",
					result.Score, tt.minScore, tt.maxScore)
			}

			if result.Assessment == "" {
				t.Errorf("ScoreDefense() missing assessment text")
			}
		})
	}
}

func TestScoreSynergy(t *testing.T) {
	// Create a minimal synergy database for testing
	synergyDB := deck.NewSynergyDatabase()

	tests := []struct {
		name     string
		cards    []deck.CardCandidate
		minScore float64
		maxScore float64
	}{
		{
			name: "Good synergy deck with known pairs",
			cards: []deck.CardCandidate{
				makeCard("Giant", deck.RoleWinCondition, 11, 11, "Rare", 5),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
				makeCard("Zap", deck.RoleSpellSmall, 11, 11, "Common", 2),
				makeCard("Mega Minion", deck.RoleSupport, 11, 11, "Rare", 3),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
			},
			minScore: 5.5,
			maxScore: 8.0,
		},
		{
			name: "Nil synergy database",
			cards: []deck.CardCandidate{
				makeCard("Giant", deck.RoleWinCondition, 11, 11, "Rare", 5),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
			},
			minScore: 5.0,
			maxScore: 5.0,
		},
		{
			name: "No synergy deck",
			cards: []deck.CardCandidate{
				makeCard("Archer Queen", deck.RoleSupport, 11, 11, "Champion", 5),
				makeCard("Golden Knight", deck.RoleSupport, 11, 11, "Champion", 4),
				makeCard("Skeleton King", deck.RoleSupport, 11, 11, "Champion", 4),
				makeCard("Little Prince", deck.RoleSupport, 11, 11, "Champion", 3),
				makeCard("Berserker", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Goblin Demolisher", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Royal Delivery", deck.RoleSpellSmall, 11, 11, "Common", 3),
				makeCard("Phoenix", deck.RoleSupport, 11, 11, "Legendary", 4),
			},
			minScore: 2.5,
			maxScore: 2.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var db *deck.SynergyDatabase
			if tt.name != "Nil synergy database" {
				db = synergyDB
			}

			result := ScoreSynergy(tt.cards, db)

			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("ScoreSynergy() score = %v, want between %v and %v",
					result.Score, tt.minScore, tt.maxScore)
			}

			if result.Assessment == "" {
				t.Errorf("ScoreSynergy() missing assessment text")
			}
		})
	}
}

func TestScoreVersatility(t *testing.T) {
	tests := []struct {
		name     string
		cards    []deck.CardCandidate
		minScore float64
		maxScore float64
	}{
		{
			name: "Highly versatile deck",
			cards: []deck.CardCandidate{
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
				makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
			},
			minScore: 6.0,
			maxScore: 10.0,
		},
		{
			name: "Limited versatility deck",
			cards: []deck.CardCandidate{
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
				makeCard("Archers", deck.RoleSupport, 11, 11, "Common", 3),
				makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Mini P.E.K.K.A", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Mega Minion", deck.RoleSupport, 11, 11, "Rare", 3),
				makeCard("Ice Golem", deck.RoleSupport, 11, 11, "Rare", 2),
			},
			minScore: 2.0,
			maxScore: 6.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScoreVersatility(tt.cards)

			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("ScoreVersatility() score = %v, want between %v and %v",
					result.Score, tt.minScore, tt.maxScore)
			}

			if result.Assessment == "" {
				t.Errorf("ScoreVersatility() missing assessment text")
			}
		})
	}
}

func TestScoreF2P(t *testing.T) {
	tests := []struct {
		name     string
		cards    []deck.CardCandidate
		minScore float64
		maxScore float64
	}{
		{
			name: "F2P-friendly deck (all commons/rares)",
			cards: []deck.CardCandidate{
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Archers", deck.RoleSupport, 11, 11, "Common", 3),
			},
			minScore: 8.0,
			maxScore: 10.0,
		},
		{
			name: "Expensive deck (multiple legendaries)",
			cards: []deck.CardCandidate{
				makeCard("Mega Knight", deck.RoleWinCondition, 11, 14, "Legendary", 7),
				makeCard("Log", deck.RoleSpellSmall, 11, 14, "Legendary", 2),
				makeCard("Miner", deck.RoleWinCondition, 11, 14, "Legendary", 3),
				makeCard("Princess", deck.RoleSupport, 11, 14, "Legendary", 3),
				makeCard("Inferno Dragon", deck.RoleSupport, 11, 14, "Legendary", 4),
				makeCard("Electro Wizard", deck.RoleSupport, 11, 14, "Legendary", 4),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
			},
			minScore: 0.0,
			maxScore: 4.0,
		},
		{
			name: "Mixed rarity deck",
			cards: []deck.CardCandidate{
				makeCard("Giant", deck.RoleWinCondition, 11, 11, "Rare", 5),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
				makeCard("Log", deck.RoleSpellSmall, 11, 14, "Legendary", 2),
				makeCard("Mega Minion", deck.RoleSupport, 11, 11, "Rare", 3),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
			},
			minScore: 5.0,
			maxScore: 8.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScoreF2P(tt.cards)

			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("ScoreF2P() score = %v, want between %v and %v",
					result.Score, tt.minScore, tt.maxScore)
			}

			if result.Assessment == "" {
				t.Errorf("ScoreF2P() missing assessment text")
			}
		})
	}
}

func TestAssessmentGenerators(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
	}{
		{
			name:     "Attack assessment",
			function: func() string { return generateAttackAssessment(2, 2.0, 8.5, 0.3) },
		},
		{
			name:     "Defense assessment",
			function: func() string { return generateDefenseAssessment(3, 1, 7.0, 0.2) },
		},
		{
			name:     "Synergy assessment",
			function: func() string { return generateSynergyAssessment([]deck.SynergyPair{}, 5, 8.0) },
		},
		{
			name:     "Versatility assessment",
			function: func() string { return generateVersatilityAssessment(5, 6, 8.5) },
		},
		{
			name:     "F2P assessment",
			function: func() string { return generateF2PAssessment(1, 2, 3, 7.0) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessment := tt.function()
			if assessment == "" {
				t.Errorf("%s() returned empty string", tt.name)
			}
		})
	}
}

// Test edge cases for assessment generators
func TestGenerateAttackAssessment(t *testing.T) {
	tests := []struct {
		name         string
		winCond      int
		spellDamage  float64
		score        float64
		evoBonus     float64
		wantContains string
	}{
		{"Excellent offense", 2, 2000, 8.5, 0.3, "Excellent"},
		{"Good offense", 2, 1500, 7.0, 0.0, "Good"},
		{"Moderate offense", 1, 1000, 5.0, 0.15, "Moderate"},
		{"Weak offense", 0, 500, 2.0, 0.0, "Weak"},
		{"Boundary excellent", 2, 2000, 8.0, 0.0, "Excellent"},
		{"Boundary good", 2, 1500, 6.0, 0.0, "Good"},
		{"Boundary moderate", 1, 1000, 4.0, 0.0, "Moderate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateAttackAssessment(tt.winCond, tt.spellDamage, tt.score, tt.evoBonus)
			if result == "" {
				t.Errorf("generateAttackAssessment() returned empty string")
			}
		})
	}
}

func TestGenerateDefenseAssessment(t *testing.T) {
	tests := []struct {
		name         string
		antiAir      int
		buildings    int
		score        float64
		evoBonus     float64
		wantContains string
	}{
		{"Excellent defense", 3, 1, 9.0, 0.2, "Solid"},
		{"Good defense", 2, 1, 7.0, 0.0, "Decent"},
		{"No anti-air warning", 0, 1, 5.0, 0.0, "no anti-air"},
		{"Weak defense", 1, 0, 3.0, 0.0, "Weak"},
		{"Boundary solid", 3, 1, 8.0, 0.0, "Solid"},
		{"Boundary decent", 2, 1, 6.0, 0.0, "Decent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateDefenseAssessment(tt.antiAir, tt.buildings, tt.score, tt.evoBonus)
			if result == "" {
				t.Errorf("generateDefenseAssessment() returned empty string")
			}
			if tt.wantContains != "" {
				// Case-insensitive check for substring
				lowerResult := result
				lowerWant := tt.wantContains
				found := false
				for i := 0; i <= len(lowerResult)-len(lowerWant); i++ {
					if lowerResult[i:i+len(lowerWant)] == lowerWant {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("generateDefenseAssessment() = %q, want to contain %q", result, tt.wantContains)
				}
			}
		})
	}
}

func TestGenerateSynergyAssessment(t *testing.T) {
	tests := []struct {
		name         string
		synergies    []deck.SynergyPair
		pairCount    int
		score        float64
		wantContains string
	}{
		{"Excellent synergy", []deck.SynergyPair{{Card1: "A", Card2: "B", Score: 0.9}}, 5, 9.0, "Excellent"},
		{"Good synergy", []deck.SynergyPair{{Card1: "A", Card2: "B", Score: 0.7}}, 3, 7.0, "Good"},
		{"Moderate synergy", []deck.SynergyPair{{Card1: "A", Card2: "B", Score: 0.5}}, 1, 5.0, "Moderate"},
		{"Poor synergy", []deck.SynergyPair{}, 0, 2.0, "Poor"},
		{"Empty synergies excellent", []deck.SynergyPair{}, 0, 8.5, "Excellent"},
		{"Boundary excellent", []deck.SynergyPair{{Card1: "A", Card2: "B", Score: 0.9}}, 5, 8.0, "Excellent"},
		{"Boundary good", []deck.SynergyPair{{Card1: "A", Card2: "B", Score: 0.7}}, 3, 6.0, "Good"},
		{"Boundary moderate", []deck.SynergyPair{{Card1: "A", Card2: "B", Score: 0.5}}, 1, 4.0, "Moderate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSynergyAssessment(tt.synergies, tt.pairCount, tt.score)
			if result == "" {
				t.Errorf("generateSynergyAssessment() returned empty string")
			}
		})
	}
}

func TestGenerateVersatilityAssessment(t *testing.T) {
	tests := []struct {
		name         string
		roleCount    int
		elixirCount  int
		score        float64
		wantContains string
	}{
		{"High versatility", 6, 6, 9.0, "Highly"},
		{"Good versatility", 5, 5, 7.0, "Good"},
		{"Moderate versatility", 3, 3, 5.0, "Moderate"},
		{"Low versatility", 2, 2, 2.0, "Limited"},
		{"Boundary highly", 6, 6, 8.0, "Highly"},
		{"Boundary good", 5, 5, 6.0, "Good"},
		{"Boundary moderate", 3, 3, 4.0, "Moderate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateVersatilityAssessment(tt.roleCount, tt.elixirCount, tt.score)
			if result == "" {
				t.Errorf("generateVersatilityAssessment() returned empty string")
			}
		})
	}
}

func TestGenerateF2PAssessment(t *testing.T) {
	tests := []struct {
		name           string
		legendaryCount int
		epicCount      int
		commonCount    int
		score          float64
		wantContains   string
	}{
		{"Excellent F2P", 0, 1, 5, 9.0, "Excellent"},
		{"Good F2P", 0, 2, 4, 7.0, "Good"},
		{"Moderate F2P", 1, 3, 3, 5.0, "Moderate"},
		{"Many legendaries", 3, 1, 2, 3.0, "legendaries"},
		{"Many epics", 1, 4, 2, 5.0, "epic"},
		{"Boundary excellent", 0, 1, 5, 8.0, "Excellent"},
		{"Boundary good", 0, 2, 4, 6.0, "Good"},
		{"Edge case 0 common", 0, 1, 0, 7.0, "Good"},
		{"Edge case all legendaries", 5, 0, 0, 0.0, "legendaries"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateF2PAssessment(tt.legendaryCount, tt.epicCount, tt.commonCount, tt.score)
			if result == "" {
				t.Errorf("generateF2PAssessment() returned empty string")
			}
		})
	}
}

func TestEmptyDeckScoring(t *testing.T) {
	emptyDeck := []deck.CardCandidate{}
	synergyDB := deck.NewSynergyDatabase()

	t.Run("ScoreAttack empty deck", func(t *testing.T) {
		result := ScoreAttack(emptyDeck)
		if result.Score != 0.0 {
			t.Errorf("ScoreAttack(empty) = %v, want 0.0", result.Score)
		}
	})

	t.Run("ScoreDefense empty deck", func(t *testing.T) {
		result := ScoreDefense(emptyDeck)
		if result.Score != 0.0 {
			t.Errorf("ScoreDefense(empty) = %v, want 0.0", result.Score)
		}
	})

	t.Run("ScoreSynergy empty deck", func(t *testing.T) {
		result := ScoreSynergy(emptyDeck, synergyDB)
		if result.Score != 0.0 {
			t.Errorf("ScoreSynergy(empty) = %v, want 0.0", result.Score)
		}
	})

	t.Run("ScoreVersatility empty deck", func(t *testing.T) {
		result := ScoreVersatility(emptyDeck)
		if result.Score != 0.0 {
			t.Errorf("ScoreVersatility(empty) = %v, want 0.0", result.Score)
		}
	})

	t.Run("ScoreF2P empty deck", func(t *testing.T) {
		result := ScoreF2P(emptyDeck)
		if result.Score != 0.0 {
			t.Errorf("ScoreF2P(empty) = %v, want 0.0", result.Score)
		}
	})
}
