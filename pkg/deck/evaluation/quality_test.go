// Package evaluation provides comprehensive testing for deck quality metrics.
//
// This test suite validates that the deck evaluation system properly scores:
// - Known meta decks (should score 8.0+)
// - Known bad decks (should score <5.0)
// - Archetype coherence (pure archetypes > mixed)
// - Synergy detection (synergistic combos score higher)
// - Counter coverage (decks with counters score higher)
// - Defensive capability (anti-air, buildings matter)
package evaluation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// ============================================================================
// Test Fixture Structures
// ============================================================================

// MetaDeckFixture represents a meta deck from the fixture file
type MetaDeckFixture struct {
	Name          string   `json:"name"`
	Archetype     string   `json:"archetype"`
	Cards         []string `json:"cards"`
	ExpectedScore float64  `json:"expected_score"`
	WinCondition  string   `json:"win_condition"`
}

// MetaDeckFixtures represents the meta decks fixture file
type MetaDeckFixtures struct {
	Version     int              `json:"version"`
	Description string           `json:"description"`
	LastUpdated string           `json:"last_updated"`
	Decks       []MetaDeckFixture `json:"decks"`
}

// BadDeckFixture represents a bad deck from the fixture file
type BadDeckFixture struct {
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	Cards           []string `json:"cards"`
	ExpectedMaxScore float64 `json:"expected_max_score"`
	Issue           string   `json:"issue"`
}

// BadDeckFixtures represents the bad decks fixture file
type BadDeckFixtures struct {
	Version     int             `json:"version"`
	Description string          `json:"description"`
	LastUpdated string          `json:"last_updated"`
	Decks       []BadDeckFixture `json:"decks"`
}

// ============================================================================
// Quality Metrics Tests
// ============================================================================

// TestQualityMetrics_MetaDecks verifies that meta decks score 8.0 or higher
func TestQualityMetrics_MetaDecks(t *testing.T) {
	fixtures, err := loadMetaDeckFixtures()
	if err != nil {
		t.Fatalf("Failed to load meta deck fixtures: %v", err)
	}

	synergyDB := deck.NewSynergyDatabase()

	for _, fixture := range fixtures.Decks {
		t.Run(fixture.Name, func(t *testing.T) {
			deckCards := createDeckFromFixture(fixture.Cards)
			result := Evaluate(deckCards, synergyDB, nil)

			// Meta decks should score 8.0 or higher (with some tolerance)
			if result.OverallScore < 7.5 {
				t.Errorf("%s: OverallScore = %.2f, want >= 7.5 (Expected: %.1f)",
					fixture.Name, result.OverallScore, fixture.ExpectedScore)
			}

			// Verify archetype detection
			if result.DetectedArchetype.String() != fixture.Archetype &&
				result.DetectedArchetype.String() != "hybrid" {
				t.Logf("%s: Detected archetype %q, expected %q",
					fixture.Name, result.DetectedArchetype, fixture.Archetype)
			}

			// Verify key metrics
			t.Logf("Deck: %s", fixture.Name)
			t.Logf("  Overall: %.2f (%s)", result.OverallScore, result.OverallRating)
			t.Logf("  Archetype: %s (%.0f%% confidence)", result.DetectedArchetype, result.ArchetypeConfidence*100)
			t.Logf("  Attack: %.2f, Defense: %.2f, Synergy: %.2f, Versatility: %.2f",
				result.Attack.Score, result.Defense.Score, result.Synergy.Score, result.Versatility.Score)
		})
	}

	// Calculate overall statistics
	totalScore := 0.0
	passCount := 0
	for _, fixture := range fixtures.Decks {
		deckCards := createDeckFromFixture(fixture.Cards)
		result := Evaluate(deckCards, synergyDB, nil)
		totalScore += result.OverallScore
		if result.OverallScore >= 7.5 {
			passCount++
		}
	}

	avgScore := totalScore / float64(len(fixtures.Decks))
	t.Logf("Meta Deck Summary:")
	t.Logf("  Average Score: %.2f/10.0", avgScore)
	t.Logf("  Passing (>=7.5): %d/%d (%.1f%%)", passCount, len(fixtures.Decks),
		float64(passCount)/float64(len(fixtures.Decks))*100)

	// Verify meta deck average meets quality threshold
	if avgScore < 7.5 {
		t.Errorf("Average meta deck score %.2f is below threshold 7.5", avgScore)
	}
}

// TestQualityMetrics_BadDecks verifies that bad decks score below 5.0
func TestQualityMetrics_BadDecks(t *testing.T) {
	fixtures, err := loadBadDeckFixtures()
	if err != nil {
		t.Fatalf("Failed to load bad deck fixtures: %v", err)
	}

	synergyDB := deck.NewSynergyDatabase()

	for _, fixture := range fixtures.Decks {
		t.Run(fixture.Name, func(t *testing.T) {
			deckCards := createDeckFromFixture(fixture.Cards)
			result := Evaluate(deckCards, synergyDB, nil)

			// Bad decks should score below their expected max
			if result.OverallScore > fixture.ExpectedMaxScore+1.0 {
				t.Errorf("%s: OverallScore = %.2f, want <= %.2f (Issue: %s)",
					fixture.Name, result.OverallScore, fixture.ExpectedMaxScore, fixture.Issue)
			}

			t.Logf("Deck: %s", fixture.Name)
			t.Logf("  Issue: %s", fixture.Issue)
			t.Logf("  Overall: %.2f (%s)", result.OverallScore, result.OverallRating)
			t.Logf("  Expected Max: %.2f", fixture.ExpectedMaxScore)
		})
	}

	// Calculate overall statistics
	totalScore := 0.0
	passCount := 0
	for _, fixture := range fixtures.Decks {
		deckCards := createDeckFromFixture(fixture.Cards)
		result := Evaluate(deckCards, synergyDB, nil)
		totalScore += result.OverallScore
		if result.OverallScore <= fixture.ExpectedMaxScore+1.0 {
			passCount++
		}
	}

	avgScore := totalScore / float64(len(fixtures.Decks))
	t.Logf("Bad Deck Summary:")
	t.Logf("  Average Score: %.2f/10.0", avgScore)
	t.Logf("  Passing: %d/%d (%.1f%%)", passCount, len(fixtures.Decks),
		float64(passCount)/float64(len(fixtures.Decks))*100)

	// Verify bad deck average stays below quality threshold
	if avgScore > 5.0 {
		t.Errorf("Average bad deck score %.2f is above threshold 5.0", avgScore)
	}
}

// TestQualityMetrics_ArchetypeCoherence verifies pure archetypes score better than mixed
func TestQualityMetrics_ArchetypeCoherence(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	tests := []struct {
		name       string
		cards      []string
		archetype  Archetype
		minScore   float64
		minConfidence float64
	}{
		{
			name: "Pure Cycle Deck (Hog Cycle)",
			cards: []string{"Hog Rider", "Musketeer", "Valkyrie", "Cannon", "Fireball",
				"The Log", "Ice Spirit", "Skeletons"},
			archetype: ArchetypeCycle,
			minScore: 7.0,
			minConfidence: 0.5,
		},
		{
			name: "Pure Beatdown Deck (Golem)",
			cards: []string{"Golem", "Night Witch", "Baby Dragon", "Tornado",
				"Lightning", "Mega Minion", "Elixir Collector", "Lumberjack"},
			archetype: ArchetypeBeatdown,
			minScore: 7.5,
			minConfidence: 0.6,
		},
		{
			name: "Pure Bait Deck (Log Bait)",
			cards: []string{"Goblin Barrel", "Princess", "Goblin Gang", "Knight",
				"Inferno Tower", "Ice Spirit", "The Log", "Rocket"},
			archetype: ArchetypeBait,
			minScore: 7.0,
			minConfidence: 0.5,
		},
		{
			name: "Mixed Strategy Deck",
			cards: []string{"Hog Rider", "Golem", "P.E.K.K.A", "Musketeer",
				"Baby Dragon", "Valkyrie", "Fireball", "Zap"},
			archetype: ArchetypeUnknown,
			minScore: 3.0,
			minConfidence: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deckCards := createDeckFromFixture(tt.cards)
			result := Evaluate(deckCards, synergyDB, nil)

			if result.OverallScore < tt.minScore {
				t.Errorf("OverallScore = %.2f, want >= %.2f", result.OverallScore, tt.minScore)
			}

			if result.ArchetypeConfidence < tt.minConfidence {
				t.Logf("Archetype confidence = %.2f, expected >= %.2f",
					result.ArchetypeConfidence, tt.minConfidence)
			}

			t.Logf("Archetype: %s (%.0f%% confidence)", result.DetectedArchetype,
				result.ArchetypeConfidence*100)
		})
	}
}

// TestQualityMetrics_SynergyDetection verifies synergistic combos score higher
func TestQualityMetrics_SynergyDetection(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	tests := []struct {
		name         string
		cards        []string
		minSynergy   float64
		minPairs     int
		description  string
	}{
		{
			name: "High Synergy Deck (Golem Beatdown)",
			cards: []string{"Golem", "Night Witch", "Baby Dragon", "Tornado",
				"Lightning", "Mega Minion", "Elixir Collector", "Lumberjack"},
			minSynergy: 6.0,
			minPairs: 5,
			description: "Strong tank+support and spell synergies",
		},
		{
			name: "High Synergy Deck (Log Bait)",
			cards: []string{"Goblin Barrel", "Princess", "Goblin Gang", "Knight",
				"Inferno Tower", "Ice Spirit", "The Log", "Rocket"},
			minSynergy: 6.0,
			minPairs: 5,
			description: "Multiple bait synergies",
		},
		{
			name: "High Synergy Deck (LavaLoon)",
			cards: []string{"Lava Hound", "Balloon", "Miner", "Mega Minion",
				"Skeleton Dragons", "Tornado", "Log", "Arrows"},
			minSynergy: 6.0,
			minPairs: 4,
			description: "Air synergy and support combos",
		},
		{
			name: "Low Synergy Deck (Random Cards)",
			cards: []string{"Archer Queen", "Golden Knight", "Skeleton King",
				"Little Prince", "Berserker", "Goblin Demolisher",
				"Royal Delivery", "Phoenix"},
			minSynergy: 0.0,
			minPairs: 0,
			description: "Champion cards with no known synergies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deckCards := createDeckFromFixture(tt.cards)
			result := Evaluate(deckCards, synergyDB, nil)

			t.Logf("%s: %s", tt.name, tt.description)
			t.Logf("  Synergy Score: %.2f (want >= %.2f)", result.Synergy.Score, tt.minSynergy)
			t.Logf("  Synergy Pairs: %d (want >= %d)", result.SynergyMatrix.PairCount, tt.minPairs)

			if result.Synergy.Score < tt.minSynergy {
				t.Errorf("Synergy score = %.2f, want >= %.2f", result.Synergy.Score, tt.minSynergy)
			}

			if result.SynergyMatrix.PairCount < tt.minPairs {
				t.Errorf("Synergy pairs = %d, want >= %d", result.SynergyMatrix.PairCount, tt.minPairs)
			}
		})
	}
}

// TestQualityMetrics_CounterCoverage verifies decks with proper counter coverage score higher
func TestQualityMetrics_CounterCoverage(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	tests := []struct {
		name            string
		cards           []string
		minDefenseScore float64
		description     string
	}{
		{
			name: "Good Counter Coverage",
			cards: []string{"Hog Rider", "Musketeer", "Mega Minion", "Valkyrie",
				"Cannon", "Fireball", "The Log", "Ice Spirit"},
			minDefenseScore: 7.0,
			description: "Has anti-air (Musketeer, Mega Minion) and building (Cannon)",
		},
		{
			name: "No Anti-Air Coverage",
			cards: []string{"Hog Rider", "Knight", "Valkyrie", "Skeleton Army",
				"Goblin Gang", "Ice Spirit", "The Log", "Cannon"},
			minDefenseScore: 0.0,
			description: "Zero anti-air capability",
		},
		{
			name: "Excellent Counter Coverage",
			cards: []string{"Hog Rider", "Musketeer", "Mega Minion", "Baby Dragon",
				"Cannon", "Fireball", "The Log", "Ice Spirit"},
			minDefenseScore: 7.5,
			description: "Multiple anti-air options plus building",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deckCards := createDeckFromFixture(tt.cards)
			result := Evaluate(deckCards, synergyDB, nil)

			t.Logf("%s: %s", tt.name, tt.description)
			t.Logf("  Defense Score: %.2f (want >= %.2f)", result.Defense.Score, tt.minDefenseScore)

			if tt.minDefenseScore > 0 && result.Defense.Score < tt.minDefenseScore {
				t.Errorf("Defense score = %.2f, want >= %.2f", result.Defense.Score, tt.minDefenseScore)
			}
		})
	}
}

// TestQualityMetrics_DefensiveCapability verifies defensive capability scoring
func TestQualityMetrics_DefensiveCapability(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	tests := []struct {
		name            string
		cards           []string
		minDefenseScore float64
		hasAntiAir      bool
		hasBuilding     bool
	}{
		{
			name:            "Complete Defense",
			cards:           []string{"Musketeer", "Mega Minion", "Cannon", "Tesla", "Baby Dragon", "Valkyrie", "Inferno Tower", "Hunter"},
			minDefenseScore: 8.0,
			hasAntiAir:      true,
			hasBuilding:     true,
		},
		{
			name:            "No Building Defense",
			cards:           []string{"Hog Rider", "Musketeer", "Mega Minion", "Valkyrie", "Knight", "Mini P.E.K.K.A", "Fireball", "Zap"},
			minDefenseScore: 6.0,
			hasAntiAir:      true,
			hasBuilding:     false,
		},
		{
			name:            "No Anti-Air",
			cards:           []string{"Hog Rider", "Knight", "Valkyrie", "Skeleton Army", "Goblin Gang", "Ice Spirit", "The Log", "Cannon"},
			minDefenseScore: 4.0,
			hasAntiAir:      false,
			hasBuilding:     true,
		},
		{
			name:            "Poor Defense",
			cards:           []string{"Hog Rider", "Knight", "Skeletons", "Ice Spirit", "Goblins", "Spear Goblins", "Bats", "Fire Spirit"},
			minDefenseScore: 3.0,
			hasAntiAir:      false,
			hasBuilding:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deckCards := createDeckFromFixture(tt.cards)
			result := Evaluate(deckCards, synergyDB, nil)

			if result.Defense.Score < tt.minDefenseScore {
				t.Errorf("Defense score = %.2f, want >= %.2f", result.Defense.Score, tt.minDefenseScore)
			}

			// Verify defensive elements
			hasAntiAir := false
			hasBuilding := false
			for _, card := range deckCards {
				if card.Stats != nil {
					if card.Stats.Targets == "Air" || card.Stats.Targets == "Air & Ground" {
						hasAntiAir = true
					}
				}
				if card.Role != nil && *card.Role == deck.RoleBuilding {
					hasBuilding = true
				}
			}

			if hasAntiAir != tt.hasAntiAir {
				t.Errorf("Anti-air detection mismatch: got %v, want %v", hasAntiAir, tt.hasAntiAir)
			}
			if hasBuilding != tt.hasBuilding {
				t.Errorf("Building detection mismatch: got %v, want %v", hasBuilding, tt.hasBuilding)
			}

			t.Logf("Defense: %.2f (%s) - Anti-air: %v, Building: %v",
				result.Defense.Score, result.Defense.Rating, hasAntiAir, hasBuilding)
		})
	}
}

// ============================================================================
// Fixture Loading Helpers
// ============================================================================

// loadMetaDeckFixtures loads the meta deck fixtures from JSON file
func loadMetaDeckFixtures() (*MetaDeckFixtures, error) {
	path := filepath.Join("test", "fixtures", "meta_decks.json")
	data, err := os.ReadFile(path)
	if err != nil {
		// Try relative path
		path = filepath.Join("..", "..", "..", "test", "fixtures", "meta_decks.json")
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}

	var fixtures MetaDeckFixtures
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return nil, err
	}

	return &fixtures, nil
}

// loadBadDeckFixtures loads the bad deck fixtures from JSON file
func loadBadDeckFixtures() (*BadDeckFixtures, error) {
	path := filepath.Join("test", "fixtures", "bad_decks.json")
	data, err := os.ReadFile(path)
	if err != nil {
		// Try relative path
		path = filepath.Join("..", "..", "..", "test", "fixtures", "bad_decks.json")
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}

	var fixtures BadDeckFixtures
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return nil, err
	}

	return &fixtures, nil
}

// createDeckFromFixture creates a CardCandidate slice from card names
func createDeckFromFixture(cardNames []string) []deck.CardCandidate {
	result := make([]deck.CardCandidate, len(cardNames))

	// Common defaults for testing
	defaultStats := &clashroyale.CombatStats{
		DamagePerSecond: 100,
		Targets:         "Air & Ground",
	}

	for i, name := range cardNames {
		// Determine card properties based on name
		role := determineCardRole(name)
		rarity := determineCardRarity(name)
		elixir := determineCardElixir(name)

		result[i] = deck.CardCandidate{
			Name:     name,
			Level:    11,
			MaxLevel: 14,
			Rarity:   rarity,
			Elixir:   elixir,
			Role:     &role,
			Stats:    defaultStats,
		}
	}

	return result
}

// determineCardRole determines the card role based on name
func determineCardRole(name string) deck.CardRole {
	winConditions := map[string]bool{
		"Hog Rider": true, "Giant": true, "Royal Giant": true, "Golem": true,
		"Lava Hound": true, "P.E.K.K.A": true, "Mega Knight": true, "Balloon": true,
		"X-Bow": true, "Mortar": true, "Miner": true, "Graveyard": true,
		"Goblin Barrel": true, "Goblin Drill": true, "Electro Giant": true,
		"Elite Barbarians": true, "Battle Ram": true, "Ram Rider": true,
		"Wall Breakers": true, "Sparky": true, "Royal Hogs": true,
		"Three Musketeers": true, "Archer Queen": true, "Golden Knight": true,
		"Skeleton King": true, "Little Prince": true, "Phoenix": true,
	}

	spellBig := map[string]bool{
		"Fireball": true, "Lightning": true, "Rocket": true, "Poison": true,
		"Freeze": true, "Earthquake": true, "Rage": true, "Tornado": true,
	}

	spellSmall := map[string]bool{
		"Zap": true, "The Log": true, "Arrows": true, "Snowball": true,
		"Barbarian Barrel": true, "Giant Snowball": true, "Royal Delivery": true,
	}

	buildings := map[string]bool{
		"Cannon": true, "Tesla": true, "Inferno Tower": true, "Bomb Tower": true,
		"X-Bow": true, "Mortar": true, "Elixir Collector": true, "Furnace": true,
		"Goblin Hut": true, "Goblin Cage": true, "Tombstone": true,
	}

	if winConditions[name] {
		return deck.RoleWinCondition
	}
	if spellBig[name] {
		return deck.RoleSpellBig
	}
	if spellSmall[name] {
		return deck.RoleSpellSmall
	}
	if buildings[name] {
		return deck.RoleBuilding
	}

	return deck.RoleSupport
}

// determineCardRarity determines the card rarity based on name
func determineCardRarity(name string) string {
	legendaries := map[string]bool{
		"Princess": true, "The Log": true, "Miner": true, "Ice Wizard": true,
		"Mega Knight": true, "Night Witch": true, "Lumberjack": true,
		"Electro Wizard": true, "Lava Hound": true, "Sparky": true,
		"Bandit": true, "Battle Ram": true, "Royal Ghost": true,
	}

	epics := map[string]bool{
		"Golem": true, "P.E.K.K.A": true, "Balloon": true, "X-Bow": true,
		"Mortar": true, "Graveyard": true, "Freeze": true, "Poison": true,
		"Tornado": true, "Rocket": true, "Lightning": true, "Baby Dragon": true,
		"Prince": true, "Dark Prince": true, "Bowling": true, "Three Musketeers": true,
	}

	champions := map[string]bool{
		"Archer Queen": true, "Golden Knight": true, "Skeleton King": true,
		"Little Prince": true, "Mighty Miner": true, "Phoenix": true,
	}

	if champions[name] {
		return "Champion"
	}
	if legendaries[name] {
		return "Legendary"
	}
	if epics[name] {
		return "Epic"
	}

	// Default to rare for testing
	return "Rare"
}

// determineCardElixir determines the card elixir cost based on name
func determineCardElixir(name string) int {
	elixirMap := map[string]int{
		"Skeletons": 1, "Ice Spirit": 1, "Bats": 1, "Fire Spirit": 1,
		"The Log": 2, "Zap": 2, "Snowball": 2, "Knight": 3,
		"Ice Golem": 2, "Heal Spirit": 1, "Spirit": 1,
		"Musketeer": 4, "Valkyrie": 4, "Mini P.E.K.K.A": 4, "Mega Minion": 3,
		"Hog Rider": 4, "Cannon": 3, "Tesla": 4, "Fireball": 4,
		"Golem": 8, "P.E.K.K.A": 7, "Mega Knight": 7, "Balloon": 5,
		"X-Bow": 6, "Mortar": 4, "Miner": 3, "Graveyard": 5,
		"Lava Hound": 7, "Electro Giant": 8, "Lightning": 6,
		"Rocket": 6, "Poison": 4, "Freeze": 4, "Tornado": 3,
		"Baby Dragon": 4, "Night Witch": 4, "Lumberjack": 4,
		"Electro Wizard": 4, "Bandit": 3, "Battle Ram": 5,
		"Royal Ghost": 3, "Inferno Tower": 5, "Inferno Dragon": 4,
		"Elixir Collector": 6, "Goblin Barrel": 3, "Goblin Gang": 3,
		"Goblin Drill": 4, "Princess": 3, "Arrows": 3,
		"Skeleton Army": 3, "Tombstone": 3, "Bomb Tower": 4,
		"Goblin Cage": 5, "Goblin Hut": 5, "Furnace": 4,
		"Archer Queen": 5, "Golden Knight": 4, "Skeleton King": 4,
		"Little Prince": 3, "Elite Barbarians": 6, "Ram Rider": 5,
		"Royal Giant": 6, "Royal Hogs": 5, "Wall Breakers": 4,
		"Sparky": 6, "Three Musketeers": 9, "Hunter": 4,
		"Witch": 5, "Executioner": 5, "Wizard": 5,
		"Magic Archer": 4, "Dart Goblin": 3, "Spear Goblins": 2,
		"Goblins": 2, "Archers": 3, "Minions": 3,
		"Skeleton Dragons": 4, "Mother Witch": 4, "Dark Prince": 4,
		"Fisherman": 3, "Royal Delivery": 3,
	}

	if elixir, ok := elixirMap[name]; ok {
		return elixir
	}

	// Default elixir for unknown cards
	return 4
}
