package evaluation

import (
	"strings"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestCountAirTargeters(t *testing.T) {
	tests := []struct {
		name          string
		cards         []deck.CardCandidate
		expectedCount int
	}{
		{
			name: "Multiple air targeters",
			cards: []deck.CardCandidate{
				makeCardWithTargets("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4, "Air & Ground", 100),
				makeCardWithTargets("Baby Dragon", deck.RoleSupport, 11, 11, "Epic", 4, "Air & Ground", 80),
				makeCardWithTargets("Knight", deck.RoleSupport, 11, 11, "Common", 3, "Ground", 90),
			},
			expectedCount: 2,
		},
		{
			name: "No air targeters",
			cards: []deck.CardCandidate{
				makeCardWithTargets("Knight", deck.RoleSupport, 11, 11, "Common", 3, "Ground", 90),
				makeCardWithTargets("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4, "Ground", 85),
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countAirTargeters(tt.cards)
			if len(result) != tt.expectedCount {
				t.Errorf("countAirTargeters() = %v cards, want %v", len(result), tt.expectedCount)
			}
		})
	}
}

func TestCalculateElixirCurve(t *testing.T) {
	cards := []deck.CardCandidate{
		makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
		makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
		makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
		makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
	}

	curve := calculateElixirCurve(cards)

	if curve[1] != 2 {
		t.Errorf("Expected 2 cards at 1 elixir, got %v", curve[1])
	}
	if curve[4] != 2 {
		t.Errorf("Expected 2 cards at 4 elixir, got %v", curve[4])
	}
}

func TestFindShortestCycle(t *testing.T) {
	tests := []struct {
		name          string
		cards         []deck.CardCandidate
		expectedTotal int
		expectedLen   int
	}{
		{
			name: "Normal deck",
			cards: []deck.CardCandidate{
				makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
				makeCard("Ice Golem", deck.RoleCycle, 11, 11, "Rare", 2),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
			},
			expectedTotal: 6, // 1+1+2+2
			expectedLen:   4,
		},
		{
			name:          "Insufficient cards",
			cards:         []deck.CardCandidate{makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3)},
			expectedTotal: 0,
			expectedLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, names := findShortestCycle(tt.cards)
			if total != tt.expectedTotal {
				t.Errorf("findShortestCycle() total = %v, want %v", total, tt.expectedTotal)
			}
			if len(names) != tt.expectedLen {
				t.Errorf("findShortestCycle() names length = %v, want %v", len(names), tt.expectedLen)
			}
		})
	}
}

func TestBuildCardList(t *testing.T) {
	cards := []deck.CardCandidate{
		makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
		makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
	}

	result := buildCardList(cards)

	if !strings.Contains(result, "Musketeer (4)") {
		t.Errorf("buildCardList() missing Musketeer (4)")
	}
	if !strings.Contains(result, "Fireball (4)") {
		t.Errorf("buildCardList() missing Fireball (4)")
	}
}

func TestCalculateDeckAvgElixir(t *testing.T) {
	cards := []deck.CardCandidate{
		makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
		makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
		makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
		makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
	}

	avg := calculateDeckAvgElixir(cards)
	expected := 2.5 // (1+1+4+4) / 4

	if avg != expected {
		t.Errorf("calculateDeckAvgElixir() = %v, want %v", avg, expected)
	}
}

// ============================================================================
// Defense Analysis Tests
// ============================================================================

func TestBuildDefenseAnalysis(t *testing.T) {
	tests := []struct {
		name            string
		cards           []deck.CardCandidate
		minScore        float64
		maxScore        float64
		expectInSummary string
		expectInDetails string
	}{
		{
			name: "Strong defense with anti-air and buildings",
			cards: []deck.CardCandidate{
				makeCardWithTargets("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4, "Air & Ground", 100),
				makeCardWithTargets("Baby Dragon", deck.RoleSupport, 11, 11, "Epic", 4, "Air & Ground", 80),
				makeCardWithTargets("Mega Minion", deck.RoleSupport, 11, 11, "Rare", 3, "Air & Ground", 120),
				makeCard("Tesla", deck.RoleBuilding, 11, 11, "Common", 4),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
			},
			minScore:        7.0,
			maxScore:        10.0,
			expectInSummary: "anti-air",
			expectInDetails: "Anti-air units",
		},
		{
			name: "Weak defense - no anti-air",
			cards: []deck.CardCandidate{
				makeCardWithTargets("Knight", deck.RoleSupport, 11, 11, "Common", 3, "Ground", 90),
				makeCardWithTargets("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4, "Ground", 85),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
				makeCardWithTargets("Earthquake", deck.RoleSpellBig, 11, 11, "Rare", 3, "Buildings", 100), // Ground spell, not anti-air
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
			},
			minScore:        3.0,
			maxScore:        6.0,
			expectInSummary: "",
			expectInDetails: "No anti-air", // Should now actually say "No anti-air units"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDefenseAnalysis(tt.cards)

			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("BuildDefenseAnalysis() score = %v, want between %v and %v",
					result.Score, tt.minScore, tt.maxScore)
			}

			if result.Title != "Defense Analysis" {
				t.Errorf("BuildDefenseAnalysis() title = %v, want 'Defense Analysis'", result.Title)
			}

			if result.Summary == "" {
				t.Errorf("BuildDefenseAnalysis() missing summary")
			}

			if tt.expectInSummary != "" && !strings.Contains(strings.ToLower(result.Summary), tt.expectInSummary) {
				t.Errorf("BuildDefenseAnalysis() summary missing '%v'", tt.expectInSummary)
			}

			if len(result.Details) == 0 {
				t.Errorf("BuildDefenseAnalysis() has no details")
			}

			detailsStr := strings.Join(result.Details, " ")
			if !strings.Contains(detailsStr, tt.expectInDetails) {
				t.Errorf("BuildDefenseAnalysis() details missing '%v'", tt.expectInDetails)
			}
		})
	}
}

// ============================================================================
// Attack Analysis Tests
// ============================================================================

func TestBuildAttackAnalysis(t *testing.T) {
	tests := []struct {
		name            string
		cards           []deck.CardCandidate
		minScore        float64
		maxScore        float64
		expectInDetails string
	}{
		{
			name: "Strong attack with multiple win conditions",
			cards: []deck.CardCandidate{
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Giant", deck.RoleWinCondition, 11, 11, "Rare", 5),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
				makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
			},
			minScore:        7.0,
			maxScore:        10.0,
			expectInDetails: "Primary win condition",
		},
		{
			name: "No win condition",
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
			minScore:        2.0,
			maxScore:        5.0,
			expectInDetails: "No dedicated win condition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildAttackAnalysis(tt.cards)

			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("BuildAttackAnalysis() score = %v, want between %v and %v",
					result.Score, tt.minScore, tt.maxScore)
			}

			if result.Title != "Attack Analysis" {
				t.Errorf("BuildAttackAnalysis() title = %v, want 'Attack Analysis'", result.Title)
			}

			detailsStr := strings.Join(result.Details, " ")
			if !strings.Contains(detailsStr, tt.expectInDetails) {
				t.Errorf("BuildAttackAnalysis() details missing '%v'", tt.expectInDetails)
			}
		})
	}
}

// ============================================================================
// Bait Analysis Tests
// ============================================================================

func TestIdentifyBaitCards(t *testing.T) {
	cards := []deck.CardCandidate{
		makeCard("Goblin Gang", deck.RoleSupport, 11, 11, "Common", 3),
		makeCard("Princess", deck.RoleSupport, 11, 11, "Legendary", 3),
		makeCard("Goblin Barrel", deck.RoleWinCondition, 11, 11, "Epic", 3),
		makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
	}

	baitGroups := identifyBaitCards(cards)

	if len(baitGroups["Log"]) != 3 {
		t.Errorf("identifyBaitCards() Log bait = %v cards, want 3", len(baitGroups["Log"]))
	}
}

func TestCalculateBaitScore(t *testing.T) {
	tests := []struct {
		name            string
		baitGroups      map[string][]string
		hasGoblinBarrel bool
		hasGoblinDrill  bool
		minScore        float64
		maxScore        float64
	}{
		{
			name: "Strong bait with Goblin Barrel",
			baitGroups: map[string][]string{
				"Log": {"Goblin Gang", "Princess", "Goblin Barrel"},
				"Zap": {"Bats", "Inferno Tower"},
			},
			hasGoblinBarrel: true,
			minScore:        7.0,
			maxScore:        10.0,
		},
		{
			name:            "No bait cards",
			baitGroups:      map[string][]string{},
			hasGoblinBarrel: false,
			minScore:        0.0,
			maxScore:        2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateBaitScore(tt.baitGroups, tt.hasGoblinBarrel, tt.hasGoblinDrill)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("calculateBaitScore() = %v, want between %v and %v",
					score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestBuildBaitAnalysis(t *testing.T) {
	tests := []struct {
		name     string
		cards    []deck.CardCandidate
		minScore float64
		maxScore float64
	}{
		{
			name: "Log bait deck",
			cards: []deck.CardCandidate{
				makeCard("Goblin Barrel", deck.RoleWinCondition, 11, 11, "Epic", 3),
				makeCard("Goblin Gang", deck.RoleSupport, 11, 11, "Common", 3),
				makeCard("Princess", deck.RoleSupport, 11, 11, "Legendary", 3),
				makeCard("Dart Goblin", deck.RoleSupport, 11, 11, "Rare", 3),
				makeCard("Inferno Tower", deck.RoleBuilding, 11, 11, "Rare", 5),
				makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
				makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
				makeCard("Rocket", deck.RoleSpellBig, 11, 11, "Rare", 6),
			},
			minScore: 7.0,
			maxScore: 10.0,
		},
		{
			name: "Non-bait deck",
			cards: []deck.CardCandidate{
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
				makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Ice Golem", deck.RoleCycle, 11, 11, "Rare", 2),
			},
			minScore: 0.0,
			maxScore: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildBaitAnalysis(tt.cards)

			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("BuildBaitAnalysis() score = %v, want between %v and %v",
					result.Score, tt.minScore, tt.maxScore)
			}

			if result.Title != "Bait Analysis" {
				t.Errorf("BuildBaitAnalysis() title = %v, want 'Bait Analysis'", result.Title)
			}
		})
	}
}

// ============================================================================
// Cycle Analysis Tests
// ============================================================================

func TestCalculateCycleScore(t *testing.T) {
	tests := []struct {
		name          string
		avgElixir     float64
		lowCostCount  int
		shortestCycle int
		minScore      float64
		maxScore      float64
	}{
		{
			name:          "Fast cycle deck",
			avgElixir:     2.8,
			lowCostCount:  4,
			shortestCycle: 6,
			minScore:      8.0,
			maxScore:      10.0,
		},
		{
			name:          "Slow beatdown deck",
			avgElixir:     4.2,
			lowCostCount:  1,
			shortestCycle: 12,
			minScore:      2.0,
			maxScore:      5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateCycleScore(tt.avgElixir, tt.lowCostCount, tt.shortestCycle)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("calculateCycleScore() = %v, want between %v and %v",
					score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestBuildCycleAnalysis(t *testing.T) {
	tests := []struct {
		name            string
		cards           []deck.CardCandidate
		minScore        float64
		maxScore        float64
		expectInDetails string
	}{
		{
			name: "Hog Cycle deck",
			cards: []deck.CardCandidate{
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
				makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Ice Golem", deck.RoleCycle, 11, 11, "Rare", 2),
			},
			minScore:        7.0,
			maxScore:        10.0,
			expectInDetails: "Fast",
		},
		{
			name: "Beatdown deck",
			cards: []deck.CardCandidate{
				makeCard("Golem", deck.RoleWinCondition, 11, 11, "Epic", 8),
				makeCard("Baby Dragon", deck.RoleSupport, 11, 11, "Epic", 4),
				makeCard("Night Witch", deck.RoleSupport, 11, 11, "Legendary", 4),
				makeCard("Lightning", deck.RoleSpellBig, 11, 11, "Epic", 6),
				makeCard("Mega Minion", deck.RoleSupport, 11, 11, "Rare", 3),
				makeCard("Tornado", deck.RoleSpellSmall, 11, 11, "Epic", 3),
				makeCard("Lumberjack", deck.RoleSupport, 11, 11, "Legendary", 4),
				makeCard("Barbarian Hut", deck.RoleBuilding, 11, 11, "Rare", 7),
			},
			minScore:        1.5, // Adjusted - very slow deck
			maxScore:        5.0,
			expectInDetails: "Slow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildCycleAnalysis(tt.cards)

			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("BuildCycleAnalysis() score = %v, want between %v and %v",
					result.Score, tt.minScore, tt.maxScore)
			}

			if result.Title != "Cycle Analysis" {
				t.Errorf("BuildCycleAnalysis() title = %v, want 'Cycle Analysis'", result.Title)
			}

			detailsStr := strings.Join(result.Details, " ")
			if !strings.Contains(detailsStr, tt.expectInDetails) {
				t.Errorf("BuildCycleAnalysis() details missing '%v'", tt.expectInDetails)
			}
		})
	}
}

// ============================================================================
// Ladder Analysis Tests
// ============================================================================

func TestIsLevelIndependent(t *testing.T) {
	tests := []struct {
		name     string
		card     deck.CardCandidate
		expected bool
	}{
		{
			name:     "Log - level independent spell",
			card:     makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
			expected: true,
		},
		{
			name:     "Ice Spirit - level independent cycle",
			card:     makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
			expected: true,
		},
		{
			name:     "Hog Rider - level dependent",
			card:     makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLevelIndependent(tt.card)
			if result != tt.expected {
				t.Errorf("isLevelIndependent(%v) = %v, want %v", tt.card.Name, result, tt.expected)
			}
		})
	}
}

func TestCalculateLadderScore(t *testing.T) {
	tests := []struct {
		name            string
		rarityScore     float64
		levelIndepScore float64
		upgradeProgress float64
		minScore        float64
		maxScore        float64
	}{
		{
			name:            "F2P friendly deck",
			rarityScore:     9.0,
			levelIndepScore: 8.0,
			upgradeProgress: 7.0,
			minScore:        7.5,
			maxScore:        9.0,
		},
		{
			name:            "Expensive deck",
			rarityScore:     3.0,
			levelIndepScore: 2.0,
			upgradeProgress: 4.0,
			minScore:        2.0,
			maxScore:        4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateLadderScore(tt.rarityScore, tt.levelIndepScore, tt.upgradeProgress)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("calculateLadderScore() = %v, want between %v and %v",
					score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestBuildLadderAnalysis(t *testing.T) {
	tests := []struct {
		name            string
		cards           []deck.CardCandidate
		minScore        float64
		maxScore        float64
		expectInDetails string
	}{
		{
			name: "F2P friendly deck",
			cards: []deck.CardCandidate{
				makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
				makeCard("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Valkyrie", deck.RoleSupport, 11, 11, "Rare", 4),
				makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
				makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
				makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
				makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
				makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
			},
			minScore:        6.0,
			maxScore:        9.0,
			expectInDetails: "F2P assessment",
		},
		{
			name: "Expensive legendary deck",
			cards: []deck.CardCandidate{
				makeCard("Mega Knight", deck.RoleWinCondition, 11, 11, "Legendary", 7),
				makeCard("Inferno Dragon", deck.RoleSupport, 11, 11, "Legendary", 4),
				makeCard("Electro Wizard", deck.RoleSupport, 11, 11, "Legendary", 4),
				makeCard("Princess", deck.RoleSupport, 11, 11, "Legendary", 3),
				makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
				makeCard("Miner", deck.RoleWinCondition, 11, 11, "Legendary", 3),
				makeCard("Magic Archer", deck.RoleSupport, 11, 11, "Legendary", 4),
				makeCard("Bandit", deck.RoleSupport, 11, 11, "Legendary", 3),
			},
			minScore:        2.0,
			maxScore:        5.0,
			expectInDetails: "multiple legendaries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildLadderAnalysis(tt.cards)

			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("BuildLadderAnalysis() score = %v, want between %v and %v",
					result.Score, tt.minScore, tt.maxScore)
			}

			if result.Title != "Ladder Analysis" {
				t.Errorf("BuildLadderAnalysis() title = %v, want 'Ladder Analysis'", result.Title)
			}

			detailsStr := strings.Join(result.Details, " ")
			if !strings.Contains(detailsStr, tt.expectInDetails) {
				t.Errorf("BuildLadderAnalysis() details missing '%v'", tt.expectInDetails)
			}
		})
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestEvaluate(t *testing.T) {
	// Create a simple Hog Cycle deck
	hogCycleDeck := []deck.CardCandidate{
		makeCard("Hog Rider", deck.RoleWinCondition, 11, 11, "Rare", 4),
		makeCardWithTargets("Musketeer", deck.RoleSupport, 11, 11, "Rare", 4, "Air & Ground", 100),
		makeCard("Cannon", deck.RoleBuilding, 11, 11, "Common", 3),
		makeCard("Fireball", deck.RoleSpellBig, 11, 11, "Rare", 4),
		makeCard("Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
		makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
		makeCard("Skeletons", deck.RoleCycle, 11, 11, "Common", 1),
		makeCard("Ice Golem", deck.RoleCycle, 11, 11, "Rare", 2),
	}

	result := Evaluate(hogCycleDeck, nil)

	// Check basic structure
	if len(result.Deck) != 8 {
		t.Errorf("Evaluate() deck length = %v, want 8", len(result.Deck))
	}

	if result.AvgElixir < 2.0 || result.AvgElixir > 4.0 {
		t.Errorf("Evaluate() avg elixir = %v, want between 2.0 and 4.0", result.AvgElixir)
	}

	// Check category scores are present
	if result.Attack.Score == 0 {
		t.Errorf("Evaluate() missing attack score")
	}
	if result.Defense.Score == 0 {
		t.Errorf("Evaluate() missing defense score")
	}
	if result.Synergy.Score == 0 {
		t.Errorf("Evaluate() missing synergy score")
	}
	if result.Versatility.Score == 0 {
		t.Errorf("Evaluate() missing versatility score")
	}
	if result.F2PFriendly.Score == 0 {
		t.Errorf("Evaluate() missing F2P score")
	}

	// Check overall score is calculated
	if result.OverallScore == 0 {
		t.Errorf("Evaluate() missing overall score")
	}

	// Check archetype is detected
	if result.DetectedArchetype == ArchetypeUnknown {
		t.Errorf("Evaluate() failed to detect archetype")
	}

	// Check all 5 analysis sections are present
	if result.DefenseAnalysis.Title == "" {
		t.Errorf("Evaluate() missing defense analysis")
	}
	if result.AttackAnalysis.Title == "" {
		t.Errorf("Evaluate() missing attack analysis")
	}
	if result.BaitAnalysis.Title == "" {
		t.Errorf("Evaluate() missing bait analysis")
	}
	if result.CycleAnalysis.Title == "" {
		t.Errorf("Evaluate() missing cycle analysis")
	}
	if result.LadderAnalysis.Title == "" {
		t.Errorf("Evaluate() missing ladder analysis")
	}

	// Check analysis sections have details
	if len(result.DefenseAnalysis.Details) == 0 {
		t.Errorf("Evaluate() defense analysis has no details")
	}
	if len(result.AttackAnalysis.Details) == 0 {
		t.Errorf("Evaluate() attack analysis has no details")
	}
	if len(result.BaitAnalysis.Details) == 0 {
		t.Errorf("Evaluate() bait analysis has no details")
	}
	if len(result.CycleAnalysis.Details) == 0 {
		t.Errorf("Evaluate() cycle analysis has no details")
	}
	if len(result.LadderAnalysis.Details) == 0 {
		t.Errorf("Evaluate() ladder analysis has no details")
	}
}

// Helper to create card with custom targets
func makeCardWithTargets(name string, role deck.CardRole, level, maxLevel int, rarity string, elixir int, targets string, dps int) deck.CardCandidate {
	rolePtr := &role
	return deck.CardCandidate{
		Name:     name,
		Level:    level,
		MaxLevel: maxLevel,
		Rarity:   rarity,
		Elixir:   elixir,
		Role:     rolePtr,
		Stats: &clashroyale.CombatStats{
			DamagePerSecond: dps,
			Targets:         targets,
		},
	}
}
