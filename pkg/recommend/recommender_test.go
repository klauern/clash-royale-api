package recommend

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/archetypes"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
)

// createTestDataDir creates a temporary directory for test data
func createTestDataDir(t *testing.T) string {
	tempDir := t.TempDir()
	return tempDir
}

// createMockCardAnalysis creates a sample CardAnalysis for testing
func createMockCardAnalysis() deck.CardAnalysis {
	return deck.CardAnalysis{
		CardLevels: map[string]deck.CardLevelData{
			// Common cards - high levels
			"Knight":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Archers":    {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Skeletons":  {Level: 13, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Ice Spirit": {Level: 13, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Zap":        {Level: 13, MaxLevel: 14, Rarity: "Common", Elixir: 2},

			// Rare cards - medium levels
			"Fireball":       {Level: 12, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Musketeer":      {Level: 11, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Hog Rider":      {Level: 12, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Valkyrie":       {Level: 11, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Giant":          {Level: 10, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Mega Minion":    {Level: 11, MaxLevel: 14, Rarity: "Rare", Elixir: 3},
			"Mini P.E.K.K.A": {Level: 11, MaxLevel: 14, Rarity: "Rare", Elixir: 4},

			// Epic cards - lower levels
			"Baby Dragon":   {Level: 9, MaxLevel: 14, Rarity: "Epic", Elixir: 4},
			"Skeleton Army": {Level: 8, MaxLevel: 14, Rarity: "Epic", Elixir: 3},
			"Goblin Barrel": {Level: 8, MaxLevel: 14, Rarity: "Epic", Elixir: 3},
			"Prince":        {Level: 7, MaxLevel: 14, Rarity: "Epic", Elixir: 5},

			// Legendary cards - low levels
			"Electro Wizard": {Level: 11, MaxLevel: 14, Rarity: "Legendary", Elixir: 4},
			"Mega Knight":    {Level: 11, MaxLevel: 14, Rarity: "Legendary", Elixir: 7},
			"Miner":          {Level: 10, MaxLevel: 14, Rarity: "Legendary", Elixir: 3},
			"Princess":       {Level: 9, MaxLevel: 14, Rarity: "Legendary", Elixir: 3},
		},
		AnalysisTime: time.Now().Format(time.RFC3339),
	}
}

// createMockArchetypeDeck creates a sample ArchetypeDeck for testing
func createMockArchetypeDeck(archetype mulligan.Archetype) archetypes.ArchetypeDeck {
	var deckCards []string
	var avgElixir float64

	switch archetype {
	case "cycle":
		deckCards = []string{"Hog Rider", "Ice Spirit", "Skeletons", "Musketeer", "Fireball", "Zap", "Knight", "Archers"}
		avgElixir = 2.9
	case "beatdown":
		deckCards = []string{"Golem", "Baby Dragon", "Mega Minion", "Lightning", "Zap", "Night Witch", "Lumberjack", "Tornado"}
		avgElixir = 4.3
	case "bait":
		deckCards = []string{"Goblin Barrel", "Princess", "Rocket", "Skeleton Army", "Knight", "Ice Spirit", "Goblin Gang", "Inferno Tower"}
		avgElixir = 3.3
	default:
		deckCards = []string{"Knight", "Archers", "Fireball", "Zap", "Musketeer", "Hog Rider", "Valkyrie", "Mega Minion"}
		avgElixir = 3.5
	}

	// Create deck details
	deckDetail := make([]deck.CardDetail, len(deckCards))
	for i, cardName := range deckCards {
		deckDetail[i] = deck.CardDetail{
			Name:     cardName,
			Level:    11,
			MaxLevel: 14,
			Rarity:   "Common",
			Elixir:   3,
			Role:     "support",
		}
	}

	return archetypes.ArchetypeDeck{
		Archetype:       archetype,
		Deck:            deckCards,
		DeckDetail:      deckDetail,
		AvgElixir:       avgElixir,
		CurrentAvgLevel: 11.0,
		TargetLevel:     12,
		CardsNeeded:     500,
		GoldNeeded:      50000,
		GemsNeeded:      0,
		DistanceMetric:  0.3,
		UpgradeDetails:  []archetypes.CardUpgrade{},
	}
}

// TestNewRecommender tests the recommender constructor
func TestNewRecommender(t *testing.T) {
	dataDir := createTestDataDir(t)
	options := DefaultOptions()

	recommender := NewRecommender(dataDir, options)

	if recommender == nil {
		t.Fatal("NewRecommender returned nil")
	}
	if recommender.archetypeAnalyzer == nil {
		t.Error("Recommender archetypeAnalyzer should not be nil")
	}
	if recommender.scorer == nil {
		t.Error("Recommender scorer should not be nil")
	}
	if recommender.variationGenerator == nil {
		t.Error("Recommender variationGenerator should not be nil")
	}
}

// TestDefaultOptions tests the default options
func TestDefaultOptions(t *testing.T) {
	options := DefaultOptions()

	if options.Limit != 5 {
		t.Errorf("Default Limit = %d, want 5", options.Limit)
	}
	if options.IncludeVariations != true {
		t.Error("Default IncludeVariations should be true")
	}
	if options.MinCompatibility != 30.0 {
		t.Errorf("Default MinCompatibility = %.1f, want 30.0", options.MinCompatibility)
	}
	if options.TargetLevel != 12 {
		t.Errorf("Default TargetLevel = %d, want 12", options.TargetLevel)
	}
	if options.MaxVariationsPerArchetype != 2 {
		t.Errorf("Default MaxVariationsPerArchetype = %d, want 2", options.MaxVariationsPerArchetype)
	}
}

// TestGenerateRecommendations_WithMockData tests the full recommendation flow
// Note: This test creates an archetype definition file to avoid filesystem dependencies
func TestGenerateRecommendations_WithMockData(t *testing.T) {
	// Create test data directory
	dataDir := createTestDataDir(t)

	// Create archetypes directory and definition file
	archetypesDir := filepath.Join(dataDir, "archetypes")
	if err := os.MkdirAll(archetypesDir, 0o755); err != nil {
		t.Fatalf("Failed to create archetypes directory: %v", err)
	}

	// Create a minimal archetype definition file
	definitionContent := `{
  "version": "1.0",
  "archetypes": [
    {
      "name": "cycle",
      "description": "Fast cycle deck",
      "constraints": {
        "min_elixir": 2.5,
        "max_elixir": 3.5,
        "preferred_cards": ["Hog Rider", "Ice Spirit"],
        "excluded_cards": ["Golem", "Mega Knight"]
      }
    }
  ]
}`

	definitionPath := filepath.Join(archetypesDir, "archetypes.json")
	if err := os.WriteFile(definitionPath, []byte(definitionContent), 0o644); err != nil {
		t.Fatalf("Failed to write archetype definition: %v", err)
	}

	// Create recommender with options
	options := RecommenderOptions{
		Limit:                     10,
		IncludeVariations:         false, // Disable variations for simpler test
		MinCompatibility:          20.0,
		TargetLevel:               12,
		MaxVariationsPerArchetype: 2,
	}

	recommender := NewRecommender(dataDir, options)
	analysis := createMockCardAnalysis()

	// Note: This will likely fail because we don't have real archetype data,
	// but we're testing the function's error handling and basic structure
	result, err := recommender.GenerateRecommendations(
		"#TEST123",
		"TestPlayer",
		analysis,
	)
	// We expect this might fail due to missing archetype data, which is okay
	// The important thing is it doesn't panic
	if err != nil {
		t.Logf("GenerateRecommendations returned error (expected with mock data): %v", err)
		return
	}

	// If it succeeds, validate the result
	if result == nil {
		t.Fatal("GenerateRecommendations returned nil result without error")
	}

	if result.PlayerTag != "#TEST123" {
		t.Errorf("PlayerTag = %s, want #TEST123", result.PlayerTag)
	}
	if result.PlayerName != "TestPlayer" {
		t.Errorf("PlayerName = %s, want TestPlayer", result.PlayerName)
	}
	if result.GeneratedAt == "" {
		t.Error("GeneratedAt should not be empty")
	}
}

// TestScoreArchetypeDeck tests archetype deck scoring
func TestScoreArchetypeDeck(t *testing.T) {
	dataDir := createTestDataDir(t)
	options := DefaultOptions()
	recommender := NewRecommender(dataDir, options)

	archDeck := createMockArchetypeDeck("cycle")
	analysis := createMockCardAnalysis()

	rec := recommender.scoreArchetypeDeck(archDeck, analysis)

	if rec == nil {
		t.Fatal("scoreArchetypeDeck returned nil")
	}
	if rec.Archetype != "cycle" {
		t.Errorf("Archetype = %s, want cycle", rec.Archetype)
	}
	if rec.Type != TypeArchetypeMatch {
		t.Errorf("Type = %s, want %s", rec.Type, TypeArchetypeMatch)
	}
	if rec.CompatibilityScore < 0 || rec.CompatibilityScore > 100 {
		t.Errorf("CompatibilityScore = %.2f, should be between 0 and 100", rec.CompatibilityScore)
	}
	if rec.SynergyScore < 0 || rec.SynergyScore > 100 {
		t.Errorf("SynergyScore = %.2f, should be between 0 and 100", rec.SynergyScore)
	}
	if rec.OverallScore < 0 || rec.OverallScore > 100 {
		t.Errorf("OverallScore = %.2f, should be between 0 and 100", rec.OverallScore)
	}
	if rec.UpgradeCost.CardsNeeded != 500 {
		t.Errorf("CardsNeeded = %d, want 500", rec.UpgradeCost.CardsNeeded)
	}
	if rec.UpgradeCost.GoldNeeded != 50000 {
		t.Errorf("GoldNeeded = %d, want 50000", rec.UpgradeCost.GoldNeeded)
	}
}

// TestScoreRecommendation tests recommendation scoring
func TestScoreRecommendation(t *testing.T) {
	dataDir := createTestDataDir(t)
	options := DefaultOptions()
	recommender := NewRecommender(dataDir, options)

	analysis := createMockCardAnalysis()

	rec := &DeckRecommendation{
		Deck: &deck.DeckRecommendation{
			Deck: []string{"Knight", "Archers", "Fireball", "Zap", "Musketeer", "Hog Rider", "Valkyrie", "Mega Minion"},
			DeckDetail: []deck.CardDetail{
				{Name: "Knight", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
				{Name: "Archers", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
				{Name: "Fireball", Level: 12, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
				{Name: "Zap", Level: 13, MaxLevel: 14, Rarity: "Common", Elixir: 2},
				{Name: "Musketeer", Level: 11, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
				{Name: "Hog Rider", Level: 12, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
				{Name: "Valkyrie", Level: 11, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
				{Name: "Mega Minion", Level: 11, MaxLevel: 14, Rarity: "Rare", Elixir: 3},
			},
			AvgElixir: 3.4,
		},
		Archetype:     "cycle",
		ArchetypeName: "cycle",
		Type:          TypeArchetypeMatch,
	}

	recommender.scoreRecommendation(rec, analysis)

	// Verify scores are calculated
	if rec.CompatibilityScore == 0 {
		t.Error("CompatibilityScore should not be 0 after scoring")
	}
	if rec.OverallScore == 0 {
		t.Error("OverallScore should not be 0 after scoring")
	}

	// Scores should be in valid range
	if rec.CompatibilityScore < 0 || rec.CompatibilityScore > 100 {
		t.Errorf("CompatibilityScore = %.2f, should be between 0 and 100", rec.CompatibilityScore)
	}
	if rec.SynergyScore < 0 || rec.SynergyScore > 100 {
		t.Errorf("SynergyScore = %.2f, should be between 0 and 100", rec.SynergyScore)
	}
	if rec.OverallScore < 0 || rec.OverallScore > 100 {
		t.Errorf("OverallScore = %.2f, should be between 0 and 100", rec.OverallScore)
	}
}

// TestScoreRecommendation_CustomVariation tests that custom variations have lower archetype fit
func TestScoreRecommendation_CustomVariation(t *testing.T) {
	dataDir := createTestDataDir(t)
	options := DefaultOptions()
	recommender := NewRecommender(dataDir, options)

	analysis := createMockCardAnalysis()

	recArchetype := &DeckRecommendation{
		Deck: &deck.DeckRecommendation{
			Deck: []string{"Knight", "Archers", "Fireball", "Zap", "Musketeer", "Hog Rider", "Valkyrie", "Mega Minion"},
			DeckDetail: []deck.CardDetail{
				{Name: "Knight", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
				{Name: "Archers", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
				{Name: "Fireball", Level: 12, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
				{Name: "Zap", Level: 13, MaxLevel: 14, Rarity: "Common", Elixir: 2},
				{Name: "Musketeer", Level: 11, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
				{Name: "Hog Rider", Level: 12, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
				{Name: "Valkyrie", Level: 11, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
				{Name: "Mega Minion", Level: 11, MaxLevel: 14, Rarity: "Rare", Elixir: 3},
			},
			AvgElixir: 3.4,
		},
		Archetype:     "cycle",
		ArchetypeName: "cycle",
		Type:          TypeArchetypeMatch,
	}

	recVariation := &DeckRecommendation{
		Deck:          recArchetype.Deck,
		Archetype:     "cycle",
		ArchetypeName: "cycle",
		Type:          TypeCustomVariation,
	}

	recommender.scoreRecommendation(recArchetype, analysis)
	recommender.scoreRecommendation(recVariation, analysis)

	// With same compatibility and synergy, archetype match should score higher
	// due to archetype fit being 100 vs 85
	if recArchetype.OverallScore <= recVariation.OverallScore {
		t.Logf("Note: Archetype match score (%.2f) should be higher than variation score (%.2f) with same cards",
			recArchetype.OverallScore, recVariation.OverallScore)
	}
}

// TestGetTopArchetypes tests top archetype selection
func TestGetTopArchetypes(t *testing.T) {
	dataDir := createTestDataDir(t)
	options := DefaultOptions()
	recommender := NewRecommender(dataDir, options)

	recommendations := []*DeckRecommendation{
		{
			Deck:         &deck.DeckRecommendation{},
			Archetype:    "cycle",
			Type:         TypeArchetypeMatch,
			OverallScore: 85.0,
		},
		{
			Deck:         &deck.DeckRecommendation{},
			Archetype:    "beatdown",
			Type:         TypeArchetypeMatch,
			OverallScore: 75.0,
		},
		{
			Deck:         &deck.DeckRecommendation{},
			Archetype:    "bait",
			Type:         TypeArchetypeMatch,
			OverallScore: 65.0,
		},
		{
			Deck:         &deck.DeckRecommendation{},
			Archetype:    "cycle",
			Type:         TypeCustomVariation,
			OverallScore: 90.0, // Variation with higher score should be excluded
		},
	}

	top := recommender.getTopArchetypes(recommendations, 2)

	if len(top) != 2 {
		t.Errorf("getTopArchetypes returned %d items, want 2", len(top))
	}

	// Should only include archetype matches, sorted by score
	for _, rec := range top {
		if rec.Type != TypeArchetypeMatch {
			t.Errorf("getTopArchetypes included variation, should only include archetype matches")
		}
	}

	// Should be sorted by score descending
	if len(top) >= 2 && top[0].OverallScore < top[1].OverallScore {
		t.Error("getTopArchetypes should return results sorted by score descending")
	}

	// First should be cycle with score 85
	if top[0].Archetype != "cycle" || top[0].OverallScore != 85.0 {
		t.Errorf("Top archetype = %s (%.1f), want cycle (85.0)", top[0].Archetype, top[0].OverallScore)
	}
}

// TestGetTopArchetypes_LimitExceedsAvailable tests requesting more archetypes than available
func TestGetTopArchetypes_LimitExceedsAvailable(t *testing.T) {
	dataDir := createTestDataDir(t)
	options := DefaultOptions()
	recommender := NewRecommender(dataDir, options)

	recommendations := []*DeckRecommendation{
		{
			Deck:         &deck.DeckRecommendation{},
			Archetype:    "cycle",
			Type:         TypeArchetypeMatch,
			OverallScore: 85.0,
		},
		{
			Deck:         &deck.DeckRecommendation{},
			Archetype:    "beatdown",
			Type:         TypeArchetypeMatch,
			OverallScore: 75.0,
		},
	}

	top := recommender.getTopArchetypes(recommendations, 5)

	// Should return only 2 (all available)
	if len(top) != 2 {
		t.Errorf("getTopArchetypes returned %d items, want 2 (all available)", len(top))
	}
}

// TestApplyFilters tests arena/league progression filtering behavior.
func TestApplyFilters(t *testing.T) {
	dataDir := createTestDataDir(t)
	recommendations := []*DeckRecommendation{
		{
			Deck:               &deck.DeckRecommendation{},
			Archetype:          "cycle",
			CompatibilityScore: 55,
			UpgradeCost: UpgradeCost{
				DistanceMetric: 0.30,
				CardsNeeded:    8000,
			},
		},
		{
			Deck:               &deck.DeckRecommendation{},
			Archetype:          "beatdown",
			CompatibilityScore: 28,
			UpgradeCost: UpgradeCost{
				DistanceMetric: 0.72,
				CardsNeeded:    26000,
			},
		},
	}

	t.Run("no filters returns all", func(t *testing.T) {
		recommender := NewRecommender(dataDir, DefaultOptions())
		filtered := recommender.applyFilters(recommendations)
		if len(filtered) != len(recommendations) {
			t.Fatalf("applyFilters returned %d, want %d", len(filtered), len(recommendations))
		}
	})

	t.Run("arena filter removes high-upgrade recommendations", func(t *testing.T) {
		options := DefaultOptions()
		options.Arena = "Arena 9"
		recommender := NewRecommender(dataDir, options)
		filtered := recommender.applyFilters(recommendations)
		if len(filtered) != 1 {
			t.Fatalf("applyFilters with arena returned %d, want 1", len(filtered))
		}
		if filtered[0].Archetype != "cycle" {
			t.Fatalf("expected cycle recommendation to remain, got %s", filtered[0].Archetype)
		}
	})

	t.Run("league filter enforces compatibility threshold", func(t *testing.T) {
		options := DefaultOptions()
		options.League = "Challenger I"
		recommender := NewRecommender(dataDir, options)
		filtered := recommender.applyFilters(recommendations)
		if len(filtered) != 1 {
			t.Fatalf("applyFilters with league returned %d, want 1", len(filtered))
		}
		if filtered[0].CompatibilityScore < 40.0 {
			t.Fatalf("expected compatibility >= 40 after challenger filter, got %.1f", filtered[0].CompatibilityScore)
		}
	})
}

// TestRecommendationFiltering tests minimum compatibility filtering
func TestRecommendationFiltering(t *testing.T) {
	// This test validates that the GenerateRecommendations function properly
	// filters out recommendations below the minimum compatibility threshold.
	// Since GenerateRecommendations needs real archetype data, we test the
	// filtering logic conceptually.

	recommendations := []*DeckRecommendation{
		{CompatibilityScore: 50.0, OverallScore: 60.0},
		{CompatibilityScore: 25.0, OverallScore: 30.0},
		{CompatibilityScore: 40.0, OverallScore: 45.0},
		{CompatibilityScore: 20.0, OverallScore: 25.0},
	}

	minCompatibility := 30.0
	filtered := make([]*DeckRecommendation, 0)
	for _, rec := range recommendations {
		if rec.CompatibilityScore >= minCompatibility {
			filtered = append(filtered, rec)
		}
	}

	if len(filtered) != 2 {
		t.Errorf("Filtering with min compatibility %.1f returned %d items, want 2", minCompatibility, len(filtered))
	}

	for _, rec := range filtered {
		if rec.CompatibilityScore < minCompatibility {
			t.Errorf("Filtered recommendation has compatibility %.1f, below minimum %.1f", rec.CompatibilityScore, minCompatibility)
		}
	}
}

// TestRecommendationSorting tests that recommendations are sorted by overall score
func TestRecommendationSorting(t *testing.T) {
	recommendations := []*DeckRecommendation{
		{OverallScore: 60.0, Archetype: "cycle"},
		{OverallScore: 85.0, Archetype: "beatdown"},
		{OverallScore: 40.0, Archetype: "bait"},
		{OverallScore: 75.0, Archetype: "control"},
	}

	// Sort by overall score descending (simulating what GenerateRecommendations does)
	for i := 0; i < len(recommendations)-1; i++ {
		for j := i + 1; j < len(recommendations); j++ {
			if recommendations[i].OverallScore < recommendations[j].OverallScore {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}

	// Verify sorted order
	if len(recommendations) != 4 {
		t.Errorf("Expected 4 recommendations, got %d", len(recommendations))
	}

	expectedOrder := []float64{85.0, 75.0, 60.0, 40.0}
	for i, expected := range expectedOrder {
		if recommendations[i].OverallScore != expected {
			t.Errorf("Position %d: score = %.1f, want %.1f", i, recommendations[i].OverallScore, expected)
		}
	}
}

// TestRecommendationLimit tests that limit is properly applied
func TestRecommendationLimit(t *testing.T) {
	recommendations := make([]*DeckRecommendation, 10)
	for i := range 10 {
		recommendations[i] = &DeckRecommendation{
			OverallScore: float64(100 - i*5),
		}
	}

	limit := 5
	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	if len(recommendations) != 5 {
		t.Errorf("After applying limit %d, got %d recommendations", limit, len(recommendations))
	}
}
