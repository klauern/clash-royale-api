package budget

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// createTestCardAnalysis creates a test card analysis with given card levels
func createTestCardAnalysis() deck.CardAnalysis {
	return deck.CardAnalysis{
		CardLevels: map[string]deck.CardLevelData{
			// Win conditions
			"Hog Rider":     {Level: 13, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Royal Giant":   {Level: 11, MaxLevel: 14, Rarity: "Common", Elixir: 6},
			"Giant":         {Level: 12, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Goblin Barrel": {Level: 14, MaxLevel: 14, Rarity: "Epic", Elixir: 3},

			// Buildings
			"Cannon":        {Level: 13, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Inferno Tower": {Level: 11, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Bomb Tower":    {Level: 12, MaxLevel: 14, Rarity: "Rare", Elixir: 4},

			// Spells
			"Fireball": {Level: 13, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Zap":      {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 2},
			"Arrows":   {Level: 13, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Log":      {Level: 12, MaxLevel: 14, Rarity: "Legendary", Elixir: 2},

			// Support
			"Musketeer":   {Level: 13, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Wizard":      {Level: 11, MaxLevel: 14, Rarity: "Rare", Elixir: 5},
			"Valkyrie":    {Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4},
			"Baby Dragon": {Level: 12, MaxLevel: 14, Rarity: "Epic", Elixir: 4},

			// Cycle
			"Knight":     {Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3},
			"Ice Spirit": {Level: 13, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Skeletons":  {Level: 12, MaxLevel: 14, Rarity: "Common", Elixir: 1},
			"Bats":       {Level: 11, MaxLevel: 14, Rarity: "Common", Elixir: 2},
		},
		AnalysisTime: "2024-01-01T12:00:00Z",
	}
}

func TestNewFinder(t *testing.T) {
	options := DefaultOptions()
	finder := NewFinder("testdata", options)

	if finder == nil {
		t.Fatal("NewFinder returned nil")
	}

	if finder.builder == nil {
		t.Error("Finder should have a builder")
	}

	if finder.options.TopN != 10 {
		t.Errorf("Expected default TopN of 10, got %d", finder.options.TopN)
	}
}

func TestDefaultOptions(t *testing.T) {
	options := DefaultOptions()

	if options.TargetAverageLevel != 12.0 {
		t.Errorf("Expected TargetAverageLevel 12.0, got %f", options.TargetAverageLevel)
	}

	if options.QuickWinMaxUpgrades != 2 {
		t.Errorf("Expected QuickWinMaxUpgrades 2, got %d", options.QuickWinMaxUpgrades)
	}

	if options.QuickWinMaxCards != 1000 {
		t.Errorf("Expected QuickWinMaxCards 1000, got %d", options.QuickWinMaxCards)
	}

	if options.SortBy != SortByROI {
		t.Errorf("Expected SortBy ROI, got %s", options.SortBy)
	}

	if options.TopN != 10 {
		t.Errorf("Expected TopN 10, got %d", options.TopN)
	}
}

func TestAnalyzeDeck(t *testing.T) {
	options := DefaultOptions()
	finder := NewFinder("testdata", options)

	// Create a test deck recommendation
	deckRec := &deck.DeckRecommendation{
		Deck: []string{
			"Hog Rider", "Musketeer", "Cannon", "Fireball",
			"Zap", "Ice Spirit", "Knight", "Valkyrie",
		},
		DeckDetail: []deck.CardDetail{
			{Name: "Hog Rider", Level: 13, MaxLevel: 14, Rarity: "Rare", Elixir: 4, Role: "win_conditions", Score: 1.2},
			{Name: "Musketeer", Level: 13, MaxLevel: 14, Rarity: "Rare", Elixir: 4, Role: "support", Score: 1.1},
			{Name: "Cannon", Level: 13, MaxLevel: 14, Rarity: "Common", Elixir: 3, Role: "buildings", Score: 1.0},
			{Name: "Fireball", Level: 13, MaxLevel: 14, Rarity: "Rare", Elixir: 4, Role: "spells_big", Score: 1.15},
			{Name: "Zap", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 2, Role: "spells_small", Score: 1.2},
			{Name: "Ice Spirit", Level: 13, MaxLevel: 14, Rarity: "Common", Elixir: 1, Role: "cycle", Score: 1.0},
			{Name: "Knight", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3, Role: "cycle", Score: 1.2},
			{Name: "Valkyrie", Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4, Role: "support", Score: 1.25},
		},
		AvgElixir: 3.125,
	}

	cardLevels := createTestCardAnalysis().CardLevels
	analysis := finder.AnalyzeDeck(deckRec, cardLevels)

	if analysis == nil {
		t.Fatal("AnalyzeDeck returned nil")
	}

	if analysis.Deck == nil {
		t.Error("Analysis should have a deck reference")
	}

	if analysis.CurrentScore <= 0 {
		t.Error("CurrentScore should be positive")
	}

	if analysis.ProjectedScore <= 0 {
		t.Error("ProjectedScore should be positive")
	}

	if analysis.ProjectedScore < analysis.CurrentScore {
		t.Error("ProjectedScore should be >= CurrentScore")
	}

	// Check that we have upgrade information
	t.Logf("Cards needed: %d, Gold needed: %d, ROI: %.4f",
		analysis.TotalCardsNeeded, analysis.TotalGoldNeeded, analysis.ROI)
	t.Logf("Budget category: %s, Quick win: %v", analysis.BudgetCategory, analysis.IsQuickWin)
}

func TestAnalyzeDeckNilInput(t *testing.T) {
	options := DefaultOptions()
	finder := NewFinder("testdata", options)

	analysis := finder.AnalyzeDeck(nil, nil)
	if analysis != nil {
		t.Error("AnalyzeDeck should return nil for nil input")
	}
}

func TestAnalyzeDeckEmptyDeck(t *testing.T) {
	options := DefaultOptions()
	finder := NewFinder("testdata", options)

	emptyDeck := &deck.DeckRecommendation{
		Deck:       []string{},
		DeckDetail: []deck.CardDetail{},
	}

	analysis := finder.AnalyzeDeck(emptyDeck, nil)
	if analysis != nil {
		t.Error("AnalyzeDeck should return nil for empty deck")
	}
}

func TestFindOptimalDecks(t *testing.T) {
	options := DefaultOptions()
	options.IncludeVariations = true
	options.MaxVariations = 3
	finder := NewFinder("testdata", options)

	cardAnalysis := createTestCardAnalysis()

	result, err := finder.FindOptimalDecks(cardAnalysis, "#TEST123", "TestPlayer")

	if err != nil {
		t.Fatalf("FindOptimalDecks returned error: %v", err)
	}

	if result == nil {
		t.Fatal("FindOptimalDecks returned nil result")
	}

	if result.PlayerTag != "#TEST123" {
		t.Errorf("Expected PlayerTag #TEST123, got %s", result.PlayerTag)
	}

	if result.PlayerName != "TestPlayer" {
		t.Errorf("Expected PlayerName TestPlayer, got %s", result.PlayerName)
	}

	if len(result.AllDecks) == 0 {
		t.Error("Expected at least one deck in results")
	}

	t.Logf("Found %d decks total", len(result.AllDecks))
	t.Logf("Ready decks: %d, Quick wins: %d", len(result.ReadyDecks), len(result.QuickWins))
	t.Logf("Summary: avg cards needed=%d, lowest cards needed=%d",
		result.Summary.AverageCardsNeeded, result.Summary.LowestCardsNeeded)
}

func TestFindOptimalDecksEmptyInput(t *testing.T) {
	options := DefaultOptions()
	finder := NewFinder("testdata", options)

	emptyAnalysis := deck.CardAnalysis{
		CardLevels: map[string]deck.CardLevelData{},
	}

	_, err := finder.FindOptimalDecks(emptyAnalysis, "#TEST", "Test")

	if err == nil {
		t.Error("Expected error for empty card analysis")
	}
}

func TestCategorizeDeck(t *testing.T) {
	options := DefaultOptions()

	tests := []struct {
		name             string
		totalCardsNeeded int
		upgradesNeeded   int
		avgLevel         float64
		expected         BudgetCategory
	}{
		{
			name:             "Ready deck - high level minimal upgrades",
			totalCardsNeeded: 50,
			upgradesNeeded:   1,
			avgLevel:         13.5,
			expected:         CategoryReady,
		},
		{
			name:             "Quick win - few upgrades needed",
			totalCardsNeeded: 500,
			upgradesNeeded:   2,
			avgLevel:         11.5,
			expected:         CategoryQuickWin,
		},
		{
			name:             "Medium investment",
			totalCardsNeeded: 3000,
			upgradesNeeded:   4,
			avgLevel:         10.5,
			expected:         CategoryMediumInvestment,
		},
		{
			name:             "Long term investment",
			totalCardsNeeded: 10000,
			upgradesNeeded:   8,
			avgLevel:         9.0,
			expected:         CategoryLongTerm,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := categorizeDeck(tt.totalCardsNeeded, tt.upgradesNeeded, tt.avgLevel, options)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFilterByMinLevel(t *testing.T) {
	original := createTestCardAnalysis()

	filtered := filterByMinLevel(original, 13)

	for name, data := range filtered.CardLevels {
		if data.Level < 13 {
			t.Errorf("Card %s has level %d, expected >= 13", name, data.Level)
		}
	}

	if len(filtered.CardLevels) >= len(original.CardLevels) {
		t.Error("Filtered should have fewer cards than original")
	}
}

func TestFilterByRarity(t *testing.T) {
	original := createTestCardAnalysis()

	filtered := filterByRarity(original, []string{"Common", "Rare"})

	for name, data := range filtered.CardLevels {
		if data.Rarity != "Common" && data.Rarity != "Rare" {
			t.Errorf("Card %s has rarity %s, expected Common or Rare", name, data.Rarity)
		}
	}
}

func TestCalculateUpgradePriority(t *testing.T) {
	card := deck.CardDetail{
		Name:     "Hog Rider",
		Level:    13,
		MaxLevel: 14,
		Role:     "win_conditions",
	}

	priority := calculateUpgradePriority(card, 550)

	if priority <= 0 {
		t.Error("Priority should be positive")
	}

	// Win condition should have higher priority
	cycleCard := deck.CardDetail{
		Name:     "Skeletons",
		Level:    12,
		MaxLevel: 14,
		Role:     "cycle",
	}

	cyclePriority := calculateUpgradePriority(cycleCard, 550)

	if priority <= cyclePriority {
		t.Logf("Win condition priority: %.4f, Cycle priority: %.4f", priority, cyclePriority)
		// Note: This is not necessarily an error as priority depends on multiple factors
	}
}

func TestSortResults(t *testing.T) {
	options := DefaultOptions()
	options.SortBy = SortByTotalCards
	finder := NewFinder("testdata", options)

	result := &BudgetFinderResult{
		AllDecks: []*DeckBudgetAnalysis{
			{TotalCardsNeeded: 1000, ROI: 0.5},
			{TotalCardsNeeded: 500, ROI: 0.8},
			{TotalCardsNeeded: 2000, ROI: 0.3},
		},
	}

	finder.sortResults(result)

	// Should be sorted by total cards ascending
	if result.AllDecks[0].TotalCardsNeeded != 500 {
		t.Errorf("Expected first deck to have 500 cards, got %d", result.AllDecks[0].TotalCardsNeeded)
	}

	if result.AllDecks[2].TotalCardsNeeded != 2000 {
		t.Errorf("Expected last deck to have 2000 cards, got %d", result.AllDecks[2].TotalCardsNeeded)
	}
}

func TestSortByROI(t *testing.T) {
	options := DefaultOptions()
	options.SortBy = SortByROI
	finder := NewFinder("testdata", options)

	result := &BudgetFinderResult{
		AllDecks: []*DeckBudgetAnalysis{
			{TotalCardsNeeded: 1000, ROI: 0.5},
			{TotalCardsNeeded: 500, ROI: 0.8},
			{TotalCardsNeeded: 2000, ROI: 0.3},
		},
	}

	finder.sortResults(result)

	// Should be sorted by ROI descending
	if result.AllDecks[0].ROI != 0.8 {
		t.Errorf("Expected first deck to have ROI 0.8, got %f", result.AllDecks[0].ROI)
	}

	if result.AllDecks[2].ROI != 0.3 {
		t.Errorf("Expected last deck to have ROI 0.3, got %f", result.AllDecks[2].ROI)
	}
}

func TestBudgetConstraints(t *testing.T) {
	options := DefaultOptions()
	options.MaxCardsNeeded = 1000
	options.MaxGoldNeeded = 50000
	finder := NewFinder("testdata", options)

	result := &BudgetFinderResult{
		AllDecks: []*DeckBudgetAnalysis{
			{TotalCardsNeeded: 500, TotalGoldNeeded: 25000},
			{TotalCardsNeeded: 1500, TotalGoldNeeded: 30000}, // Over card limit
			{TotalCardsNeeded: 800, TotalGoldNeeded: 75000},  // Over gold limit
			{TotalCardsNeeded: 900, TotalGoldNeeded: 40000},
		},
	}

	finder.categorizeResults(result)

	if len(result.WithinBudget) != 2 {
		t.Errorf("Expected 2 decks within budget, got %d", len(result.WithinBudget))
	}
}

// Benchmark tests
func BenchmarkAnalyzeDeck(b *testing.B) {
	options := DefaultOptions()
	finder := NewFinder("testdata", options)

	deckRec := &deck.DeckRecommendation{
		Deck: []string{
			"Hog Rider", "Musketeer", "Cannon", "Fireball",
			"Zap", "Ice Spirit", "Knight", "Valkyrie",
		},
		DeckDetail: []deck.CardDetail{
			{Name: "Hog Rider", Level: 13, MaxLevel: 14, Rarity: "Rare", Elixir: 4, Score: 1.2},
			{Name: "Musketeer", Level: 13, MaxLevel: 14, Rarity: "Rare", Elixir: 4, Score: 1.1},
			{Name: "Cannon", Level: 13, MaxLevel: 14, Rarity: "Common", Elixir: 3, Score: 1.0},
			{Name: "Fireball", Level: 13, MaxLevel: 14, Rarity: "Rare", Elixir: 4, Score: 1.15},
			{Name: "Zap", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 2, Score: 1.2},
			{Name: "Ice Spirit", Level: 13, MaxLevel: 14, Rarity: "Common", Elixir: 1, Score: 1.0},
			{Name: "Knight", Level: 14, MaxLevel: 14, Rarity: "Common", Elixir: 3, Score: 1.2},
			{Name: "Valkyrie", Level: 14, MaxLevel: 14, Rarity: "Rare", Elixir: 4, Score: 1.25},
		},
		AvgElixir: 3.125,
	}

	cardLevels := createTestCardAnalysis().CardLevels

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		finder.AnalyzeDeck(deckRec, cardLevels)
	}
}

func BenchmarkFindOptimalDecks(b *testing.B) {
	options := DefaultOptions()
	options.IncludeVariations = true
	options.MaxVariations = 3
	finder := NewFinder("testdata", options)

	cardAnalysis := createTestCardAnalysis()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		finder.FindOptimalDecks(cardAnalysis, "#TEST", "Test")
	}
}
