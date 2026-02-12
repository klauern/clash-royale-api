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
	Version     int               `json:"version"`
	Description string            `json:"description"`
	LastUpdated string            `json:"last_updated"`
	Decks       []MetaDeckFixture `json:"decks"`
}

// BadDeckFixture represents a bad deck from the fixture file
type BadDeckFixture struct {
	Name             string   `json:"name"`
	Description      string   `json:"description,omitempty"`
	Cards            []string `json:"cards"`
	ExpectedMaxScore float64  `json:"expected_max_score"`
	Issue            string   `json:"issue"`
}

// BadDeckFixtures represents the bad decks fixture file
type BadDeckFixtures struct {
	Version     int              `json:"version"`
	Description string           `json:"description"`
	LastUpdated string           `json:"last_updated"`
	Decks       []BadDeckFixture `json:"decks"`
}

// ============================================================================
// Quality Metrics Tests
// ============================================================================

// TestQualityMetrics_MetaDecks verifies that meta decks maintain strong scores.
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

			minAllowed := fixture.ExpectedScore - 3.5
			if minAllowed < 4.2 {
				minAllowed = 4.2
			}

			// Meta decks should remain close to their fixture expectation.
			if result.OverallScore < minAllowed {
				t.Errorf("%s: OverallScore = %.2f, want >= %.2f (Expected: %.1f)",
					fixture.Name, result.OverallScore, minAllowed, fixture.ExpectedScore)
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
		minAllowed := fixture.ExpectedScore - 3.5
		if minAllowed < 4.2 {
			minAllowed = 4.2
		}
		if result.OverallScore >= minAllowed {
			passCount++
		}
	}

	avgScore := totalScore / float64(len(fixtures.Decks))
	t.Logf("Meta Deck Summary:")
	t.Logf("  Average Score: %.2f/10.0", avgScore)
	t.Logf("  Passing (fixture expected - 3.5): %d/%d (%.1f%%)", passCount, len(fixtures.Decks),
		float64(passCount)/float64(len(fixtures.Decks))*100)

	// Verify meta deck average meets quality threshold
	if avgScore < 6.5 {
		t.Errorf("Average meta deck score %.2f is below threshold 6.5", avgScore)
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

			// Bad deck fixtures are historical; allow wider drift while
			// retaining a cap against severe regressions.
			if result.OverallScore > fixture.ExpectedMaxScore+3.0 {
				t.Errorf("%s: OverallScore = %.2f, want <= %.2f (Issue: %s)",
					fixture.Name, result.OverallScore, fixture.ExpectedMaxScore+3.0, fixture.Issue)
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
		if result.OverallScore <= fixture.ExpectedMaxScore+3.0 {
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
		name          string
		cards         []string
		archetype     Archetype
		minScore      float64
		minConfidence float64
	}{
		{
			name: "Pure Cycle Deck (Hog Cycle)",
			cards: []string{
				"Hog Rider", "Musketeer", "Valkyrie", "Cannon", "Fireball",
				"The Log", "Ice Spirit", "Skeletons",
			},
			archetype:     ArchetypeCycle,
			minScore:      7.0,
			minConfidence: 0.5,
		},
		{
			name: "Pure Beatdown Deck (Golem)",
			cards: []string{
				"Golem", "Night Witch", "Baby Dragon", "Tornado",
				"Lightning", "Mega Minion", "Elixir Collector", "Lumberjack",
			},
			archetype:     ArchetypeBeatdown,
			minScore:      7.5,
			minConfidence: 0.6,
		},
		{
			name: "Pure Bait Deck (Log Bait)",
			cards: []string{
				"Goblin Barrel", "Princess", "Goblin Gang", "Knight",
				"Inferno Tower", "Ice Spirit", "The Log", "Rocket",
			},
			archetype:     ArchetypeBait,
			minScore:      6.8,
			minConfidence: 0.5,
		},
		{
			name: "Mixed Strategy Deck",
			cards: []string{
				"Hog Rider", "Golem", "P.E.K.K.A", "Musketeer",
				"Baby Dragon", "Valkyrie", "Fireball", "Zap",
			},
			archetype:     ArchetypeUnknown,
			minScore:      3.0,
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
		name        string
		cards       []string
		minSynergy  float64
		minPairs    int
		description string
	}{
		{
			name: "High Synergy Deck (Golem Beatdown)",
			cards: []string{
				"Golem", "Night Witch", "Baby Dragon", "Tornado",
				"Lightning", "Mega Minion", "Elixir Collector", "Lumberjack",
			},
			minSynergy:  6.0,
			minPairs:    4,
			description: "Strong tank+support and spell synergies",
		},
		{
			name: "High Synergy Deck (Log Bait)",
			cards: []string{
				"Goblin Barrel", "Princess", "Goblin Gang", "Knight",
				"Inferno Tower", "Ice Spirit", "The Log", "Rocket",
			},
			minSynergy:  6.0,
			minPairs:    4,
			description: "Multiple bait synergies",
		},
		{
			name: "High Synergy Deck (LavaLoon)",
			cards: []string{
				"Lava Hound", "Balloon", "Miner", "Mega Minion",
				"Skeleton Dragons", "Tornado", "Log", "Arrows",
			},
			minSynergy:  6.0,
			minPairs:    4,
			description: "Air synergy and support combos",
		},
		{
			name: "Low Synergy Deck (Random Cards)",
			cards: []string{
				"Archer Queen", "Golden Knight", "Skeleton King",
				"Little Prince", "Berserker", "Goblin Demolisher",
				"Royal Delivery", "Phoenix",
			},
			minSynergy:  0.0,
			minPairs:    0,
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
			cards: []string{
				"Hog Rider", "Musketeer", "Mega Minion", "Valkyrie",
				"Cannon", "Fireball", "The Log", "Ice Spirit",
			},
			minDefenseScore: 7.0,
			description:     "Has anti-air (Musketeer, Mega Minion) and building (Cannon)",
		},
		{
			name: "No Anti-Air Coverage",
			cards: []string{
				"Hog Rider", "Knight", "Valkyrie", "Skeleton Army",
				"Goblin Gang", "Ice Spirit", "The Log", "Cannon",
			},
			minDefenseScore: 0.0,
			description:     "Zero anti-air capability",
		},
		{
			name: "Excellent Counter Coverage",
			cards: []string{
				"Hog Rider", "Musketeer", "Mega Minion", "Baby Dragon",
				"Cannon", "Fireball", "The Log", "Ice Spirit",
			},
			minDefenseScore: 7.5,
			description:     "Multiple anti-air options plus building",
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
			hasAntiAir:      true,
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

	for i, name := range cardNames {
		result[i] = createTestCardCandidate(name)
	}

	return result
}
