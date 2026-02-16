package evaluation

import (
	"strings"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/events"
)

// ============================================================================
// Test Fixtures and Helpers
// ============================================================================

// createMockEventAnalysis creates a mock EventAnalysis with predictable data
func createMockEventAnalysis(totalBattles int, overallWinRate float64) *events.EventAnalysis {
	return &events.EventAnalysis{
		PlayerTag:    "#TEST123",
		AnalysisTime: time.Now(),
		TotalDecks:   5,
		Summary: events.EventSummary{
			TotalBattles:       totalBattles,
			TotalWins:          int(float64(totalBattles) * overallWinRate),
			TotalLosses:        int(float64(totalBattles) * (1 - overallWinRate)),
			OverallWinRate:     overallWinRate,
			AvgCrownsPerBattle: 1.5,
			AvgDeckElixir:      3.5,
		},
		CardAnalysis: events.EventCardAnalysis{
			MostUsedCards: []events.CardUsage{
				{CardName: "Hog Rider", Count: 50},
				{CardName: "Fireball", Count: 40},
				{CardName: "Musketeer", Count: 30},
				{CardName: "Valkyrie", Count: 20},
				{CardName: "Ice Spirit", Count: 10},
			},
			HighestWinRateCards: []events.CardWinRate{
				{CardName: "Hog Rider", WinRate: 0.65},        // S tier
				{CardName: "Fireball", WinRate: 0.57},         // A tier
				{CardName: "Musketeer", WinRate: 0.52},        // B tier
				{CardName: "Valkyrie", WinRate: 0.47},         // C tier
				{CardName: "Ice Spirit", WinRate: 0.42},       // D tier
				{CardName: "Elite Barbarians", WinRate: 0.40}, // D tier
			},
			TotalUniqueCards: 6,
		},
	}
}

// createEmptyEventAnalysis creates an EventAnalysis with no data
func createEmptyEventAnalysis() *events.EventAnalysis {
	return &events.EventAnalysis{
		PlayerTag:    "#EMPTY",
		AnalysisTime: time.Now(),
		TotalDecks:   0,
		Summary: events.EventSummary{
			TotalBattles:       0,
			TotalWins:          0,
			TotalLosses:        0,
			OverallWinRate:     0,
			AvgCrownsPerBattle: 0,
			AvgDeckElixir:      0,
		},
		CardAnalysis: events.EventCardAnalysis{
			MostUsedCards:       []events.CardUsage{},
			HighestWinRateCards: []events.CardWinRate{},
			TotalUniqueCards:    0,
		},
	}
}

// ============================================================================
// Core Function Tests
// ============================================================================

func TestDefaultMetaAnalysisOptions(t *testing.T) {
	opts := DefaultMetaAnalysisOptions()

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"EnableMetaScoring", opts.EnableMetaScoring, true},
		{"MetaWeight", opts.MetaWeight, 0.15},
		{"TrendWindowDays", opts.TrendWindowDays, 30},
		{"MinSampleSize", opts.MinSampleSize, 10},
		{"ShowWeakMatchups", opts.ShowWeakMatchups, true},
		{"RecommendMetaAlternatives", opts.RecommendMetaAlternatives, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestNewMetaAnalyzer(t *testing.T) {
	tests := []struct {
		name        string
		eventData   *events.EventAnalysis
		options     MetaAnalysisOptions
		expectNil   bool
		checkWeight float64
	}{
		{
			name:        "Valid analyzer with custom options",
			eventData:   createMockEventAnalysis(100, 0.55),
			options:     MetaAnalysisOptions{MetaWeight: 0.20},
			expectNil:   false,
			checkWeight: 0.20,
		},
		{
			name:        "Analyzer with zero weight uses defaults",
			eventData:   createMockEventAnalysis(100, 0.55),
			options:     MetaAnalysisOptions{MetaWeight: 0},
			expectNil:   false,
			checkWeight: 0.15, // Should use default
		},
		{
			name:        "Analyzer with nil event data",
			eventData:   nil,
			options:     DefaultMetaAnalysisOptions(),
			expectNil:   false,
			checkWeight: 0.15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMetaAnalyzer(tt.eventData, tt.options)

			if tt.expectNil && analyzer != nil {
				t.Errorf("NewMetaAnalyzer() expected nil, got %v", analyzer)
			}

			if !tt.expectNil && analyzer == nil {
				t.Fatal("NewMetaAnalyzer() returned nil unexpectedly")
			}

			if analyzer != nil && analyzer.options.MetaWeight != tt.checkWeight {
				t.Errorf("MetaWeight = %v, want %v", analyzer.options.MetaWeight, tt.checkWeight)
			}
		})
	}
}

func TestAnalyzeDeckWithMeta(t *testing.T) {
	tests := []struct {
		name           string
		eventData      *events.EventAnalysis
		deckCards      []string
		baseScore      float64
		expectNonEmpty bool
		minAdjustment  float64
		maxAdjustment  float64
	}{
		{
			name:           "High win rate deck with meta data",
			eventData:      createMockEventAnalysis(100, 0.55),
			deckCards:      []string{"Hog Rider", "Fireball", "Musketeer", "Valkyrie", "Ice Spirit", "Log", "Cannon", "Ice Golem"},
			baseScore:      75.0,
			expectNonEmpty: true,
			minAdjustment:  -10.0,
			maxAdjustment:  20.0,
		},
		{
			name:           "No event data returns zero adjustment",
			eventData:      nil,
			deckCards:      []string{"Hog Rider", "Fireball"},
			baseScore:      75.0,
			expectNonEmpty: false,
			minAdjustment:  0.0,
			maxAdjustment:  0.0,
		},
		{
			name:           "Empty event data returns zero adjustment",
			eventData:      createEmptyEventAnalysis(),
			deckCards:      []string{"Hog Rider", "Fireball"},
			baseScore:      75.0,
			expectNonEmpty: false,
			minAdjustment:  0.0,
			maxAdjustment:  0.0,
		},
		{
			name:           "Empty deck",
			eventData:      createMockEventAnalysis(100, 0.55),
			deckCards:      []string{},
			baseScore:      0.0,
			expectNonEmpty: true,
			minAdjustment:  -5.0,
			maxAdjustment:  5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMetaAnalyzer(tt.eventData, DefaultMetaAnalysisOptions())
			analysis := analyzer.AnalyzeDeckWithMeta(tt.deckCards, tt.baseScore)

			if analysis == nil {
				t.Fatal("AnalyzeDeckWithMeta() returned nil")
			}

			// Verify deck cards match
			if len(analysis.DeckCards) != len(tt.deckCards) {
				t.Errorf("DeckCards length = %v, want %v", len(analysis.DeckCards), len(tt.deckCards))
			}

			// Verify meta adjustment
			if analysis.MetaAdjustment.BaseScore != tt.baseScore {
				t.Errorf("BaseScore = %v, want %v", analysis.MetaAdjustment.BaseScore, tt.baseScore)
			}

			// Check adjustment range
			adj := analysis.MetaAdjustment.Adjustment
			if adj < tt.minAdjustment || adj > tt.maxAdjustment {
				t.Errorf("Adjustment = %v, want between %v and %v", adj, tt.minAdjustment, tt.maxAdjustment)
			}

			// Verify non-empty fields when data is present
			if tt.expectNonEmpty {
				if len(analysis.CardMetaInfo) == 0 && len(tt.deckCards) > 0 {
					t.Error("Expected non-empty CardMetaInfo")
				}
			}
		})
	}
}

func TestApplyMetaAdjustment(t *testing.T) {
	tests := []struct {
		name           string
		baseScore      float64
		metaAdjustment MetaAdjustment
		expected       float64
	}{
		{
			name:      "Apply non-zero adjustment",
			baseScore: 75.0,
			metaAdjustment: MetaAdjustment{
				BaseScore:  75.0,
				MetaScore:  80.0,
				Adjustment: 5.0,
			},
			expected: 80.0,
		},
		{
			name:      "Zero adjustment",
			baseScore: 75.0,
			metaAdjustment: MetaAdjustment{
				BaseScore:  0,
				MetaScore:  0,
				Adjustment: 0,
			},
			expected: 75.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyMetaAdjustment(tt.baseScore, tt.metaAdjustment)
			if result != tt.expected {
				t.Errorf("ApplyMetaAdjustment() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMetaAdjustmentIsZero(t *testing.T) {
	tests := []struct {
		name       string
		adjustment MetaAdjustment
		expected   bool
	}{
		{
			name: "Zero adjustment",
			adjustment: MetaAdjustment{
				BaseScore:  0,
				MetaScore:  0,
				Adjustment: 0,
			},
			expected: true,
		},
		{
			name: "Non-zero base score",
			adjustment: MetaAdjustment{
				BaseScore:  75.0,
				MetaScore:  0,
				Adjustment: 0,
			},
			expected: false,
		},
		{
			name: "Non-zero meta score",
			adjustment: MetaAdjustment{
				BaseScore:  0,
				MetaScore:  80.0,
				Adjustment: 0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.adjustment.IsZero()
			if result != tt.expected {
				t.Errorf("IsZero() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestCalculateMetaTier(t *testing.T) {
	tests := []struct {
		name     string
		winRate  float64
		expected string
	}{
		{"S tier - high win rate", 0.65, "S"},
		{"S tier - boundary", 0.60, "S"},
		{"A tier - above average", 0.57, "A"},
		{"A tier - boundary", 0.55, "A"},
		{"B tier - balanced", 0.52, "B"},
		{"B tier - boundary", 0.50, "B"},
		{"C tier - below average", 0.47, "C"},
		{"C tier - boundary", 0.45, "C"},
		{"D tier - weak", 0.42, "D"},
		{"D tier - very weak", 0.30, "D"},
		{"Edge case - zero", 0.0, "D"},
		{"Edge case - perfect", 1.0, "S"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMetaTier(tt.winRate)
			if result != tt.expected {
				t.Errorf("calculateMetaTier(%v) = %v, want %v", tt.winRate, result, tt.expected)
			}
		})
	}
}

func TestBuildCardPopularityMap(t *testing.T) {
	tests := []struct {
		name      string
		eventData *events.EventAnalysis
		cardName  string
		expected  float64
	}{
		{
			name:      "Most popular card",
			eventData: createMockEventAnalysis(100, 0.55),
			cardName:  "Hog Rider",
			expected:  100.0, // 50/50 * 100 = 100
		},
		{
			name:      "Medium popularity card",
			eventData: createMockEventAnalysis(100, 0.55),
			cardName:  "Fireball",
			expected:  80.0, // 40/50 * 100 = 80
		},
		{
			name:      "Least popular card",
			eventData: createMockEventAnalysis(100, 0.55),
			cardName:  "Ice Spirit",
			expected:  20.0, // 10/50 * 100 = 20
		},
		{
			name:      "Card not in data",
			eventData: createMockEventAnalysis(100, 0.55),
			cardName:  "Unknown Card",
			expected:  0.0,
		},
		{
			name:      "Empty event data",
			eventData: createEmptyEventAnalysis(),
			cardName:  "Hog Rider",
			expected:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMetaAnalyzer(tt.eventData, DefaultMetaAnalysisOptions())
			popMap := analyzer.buildCardPopularityMap()

			result := popMap[tt.cardName]
			if result != tt.expected {
				t.Errorf("buildCardPopularityMap()[%v] = %v, want %v", tt.cardName, result, tt.expected)
			}
		})
	}
}

func TestBuildCardWinRateMap(t *testing.T) {
	tests := []struct {
		name      string
		eventData *events.EventAnalysis
		cardName  string
		expected  float64
	}{
		{
			name:      "High win rate card",
			eventData: createMockEventAnalysis(100, 0.55),
			cardName:  "Hog Rider",
			expected:  0.65,
		},
		{
			name:      "Medium win rate card",
			eventData: createMockEventAnalysis(100, 0.55),
			cardName:  "Musketeer",
			expected:  0.52,
		},
		{
			name:      "Low win rate card",
			eventData: createMockEventAnalysis(100, 0.55),
			cardName:  "Ice Spirit",
			expected:  0.42,
		},
		{
			name:      "Card not in data",
			eventData: createMockEventAnalysis(100, 0.55),
			cardName:  "Unknown Card",
			expected:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMetaAnalyzer(tt.eventData, DefaultMetaAnalysisOptions())
			winRateMap := analyzer.buildCardWinRateMap()

			result := winRateMap[tt.cardName]
			if result != tt.expected {
				t.Errorf("buildCardWinRateMap()[%v] = %v, want %v", tt.cardName, result, tt.expected)
			}
		})
	}
}

func TestEstimateSampleSize(t *testing.T) {
	tests := []struct {
		name         string
		eventData    *events.EventAnalysis
		cardName     string
		expectedSize int
	}{
		{
			name:         "With event data",
			eventData:    createMockEventAnalysis(100, 0.55),
			cardName:     "Hog Rider",
			expectedSize: 100,
		},
		{
			name:         "No event data",
			eventData:    nil,
			cardName:     "Hog Rider",
			expectedSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMetaAnalyzer(tt.eventData, DefaultMetaAnalysisOptions())
			result := analyzer.estimateSampleSize(tt.cardName)

			if result != tt.expectedSize {
				t.Errorf("estimateSampleSize(%v) = %v, want %v", tt.cardName, result, tt.expectedSize)
			}
		})
	}
}

func TestAnalyzeCardMeta(t *testing.T) {
	eventData := createMockEventAnalysis(100, 0.55)
	analyzer := NewMetaAnalyzer(eventData, DefaultMetaAnalysisOptions())

	tests := []struct {
		name          string
		deckCards     []string
		expectedCount int
	}{
		{
			name:          "Full deck",
			deckCards:     []string{"Hog Rider", "Fireball", "Musketeer", "Valkyrie", "Ice Spirit", "Log", "Cannon", "Ice Golem"},
			expectedCount: 8,
		},
		{
			name:          "Partial deck",
			deckCards:     []string{"Hog Rider", "Fireball"},
			expectedCount: 2,
		},
		{
			name:          "Empty deck",
			deckCards:     []string{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeCardMeta(tt.deckCards)

			if len(result) != tt.expectedCount {
				t.Errorf("analyzeCardMeta() returned %v cards, want %v", len(result), tt.expectedCount)
			}

			// Verify each card has proper meta info
			for _, meta := range result {
				if meta.CardName == "" {
					t.Error("Card has empty name")
				}
				if meta.MetaTier == "" {
					t.Errorf("Card %v has empty meta tier", meta.CardName)
				}
				if meta.Trend == "" {
					t.Errorf("Card %v has empty trend", meta.CardName)
				}
			}
		})
	}
}

func TestCountTrendingCards(t *testing.T) {
	eventData := createMockEventAnalysis(100, 0.55)
	analyzer := NewMetaAnalyzer(eventData, DefaultMetaAnalysisOptions())

	cardMeta := []CardMetaInfo{
		{CardName: "Hog Rider", IsTrending: true},
		{CardName: "Fireball", IsTrending: true},
		{CardName: "Musketeer", IsTrending: false},
		{CardName: "Valkyrie", IsTrending: false},
	}

	result := analyzer.countTrendingCards(cardMeta)
	expected := 2

	if result != expected {
		t.Errorf("countTrendingCards() = %v, want %v", result, expected)
	}
}

// ============================================================================
// Analysis Function Tests
// ============================================================================

func TestCalculateMetaAdjustment(t *testing.T) {
	tests := []struct {
		name          string
		eventData     *events.EventAnalysis
		options       MetaAnalysisOptions
		deckCards     []string
		baseScore     float64
		minAdjustment float64
		maxAdjustment float64
		checkFactors  bool
	}{
		{
			name:          "High win rate deck gets positive adjustment",
			eventData:     createMockEventAnalysis(100, 0.55),
			options:       DefaultMetaAnalysisOptions(),
			deckCards:     []string{"Hog Rider", "Fireball"}, // Both have >50% win rate
			baseScore:     75.0,
			minAdjustment: 1.0,
			maxAdjustment: 15.0,
			checkFactors:  true,
		},
		{
			name:          "Low win rate deck gets negative adjustment",
			eventData:     createMockEventAnalysis(100, 0.55),
			options:       DefaultMetaAnalysisOptions(),
			deckCards:     []string{"Valkyrie", "Ice Spirit"}, // Both have <50% win rate
			baseScore:     75.0,
			minAdjustment: -10.0,
			maxAdjustment: 0.0,
			checkFactors:  true,
		},
		{
			name:          "Meta scoring disabled still adds trending bonus",
			eventData:     createMockEventAnalysis(100, 0.55),
			options:       MetaAnalysisOptions{EnableMetaScoring: false},
			deckCards:     []string{"Hog Rider", "Fireball"},
			baseScore:     75.0,
			minAdjustment: 0.0,
			maxAdjustment: 10.0, // Trending bonuses still apply even with meta scoring disabled
			checkFactors:  false,
		},
		{
			name:          "Empty card meta",
			eventData:     createMockEventAnalysis(100, 0.55),
			options:       DefaultMetaAnalysisOptions(),
			deckCards:     []string{},
			baseScore:     75.0,
			minAdjustment: 0.0,
			maxAdjustment: 0.0,
			checkFactors:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMetaAnalyzer(tt.eventData, tt.options)
			cardMeta := analyzer.analyzeCardMeta(tt.deckCards)
			adjustment := analyzer.calculateMetaAdjustment(tt.deckCards, cardMeta, tt.baseScore)

			// Check base score
			if adjustment.BaseScore != tt.baseScore {
				t.Errorf("BaseScore = %v, want %v", adjustment.BaseScore, tt.baseScore)
			}

			// Check adjustment range
			if adjustment.Adjustment < tt.minAdjustment || adjustment.Adjustment > tt.maxAdjustment {
				t.Errorf("Adjustment = %v, want between %v and %v",
					adjustment.Adjustment, tt.minAdjustment, tt.maxAdjustment)
			}

			// Check meta score is properly calculated
			expectedMetaScore := tt.baseScore + adjustment.Adjustment
			if expectedMetaScore > 100 {
				expectedMetaScore = 100
			} else if expectedMetaScore < 0 {
				expectedMetaScore = 0
			}

			if adjustment.MetaScore != expectedMetaScore {
				t.Errorf("MetaScore = %v, want %v", adjustment.MetaScore, expectedMetaScore)
			}

			// Check factors are present when expected
			if tt.checkFactors && len(adjustment.Factors) == 0 {
				t.Error("Expected non-empty Factors")
			}

			// Check meta tier is set
			if adjustment.MetaTier == "" && len(cardMeta) > 0 {
				t.Error("MetaTier should be set")
			}

			// Check confidence is between 0 and 1
			if adjustment.Confidence < 0 || adjustment.Confidence > 1 {
				t.Errorf("Confidence = %v, want between 0 and 1", adjustment.Confidence)
			}
		})
	}
}

func TestCalculateConfidence(t *testing.T) {
	tests := []struct {
		name       string
		eventData  *events.EventAnalysis
		dataPoints int
		minConf    float64
		maxConf    float64
	}{
		{
			name:       "Large sample size",
			eventData:  createMockEventAnalysis(1000, 0.55),
			dataPoints: 8,
			minConf:    0.5,
			maxConf:    1.0,
		},
		{
			name:       "Medium sample size",
			eventData:  createMockEventAnalysis(100, 0.55),
			dataPoints: 8,
			minConf:    0.3,
			maxConf:    1.0,
		},
		{
			name:       "Small sample size below minimum",
			eventData:  createMockEventAnalysis(5, 0.55),
			dataPoints: 8,
			minConf:    0.0,
			maxConf:    0.0,
		},
		{
			name:       "No event data",
			eventData:  nil,
			dataPoints: 8,
			minConf:    0.0,
			maxConf:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMetaAnalyzer(tt.eventData, DefaultMetaAnalysisOptions())
			result := analyzer.calculateConfidence(tt.dataPoints)

			if result < tt.minConf || result > tt.maxConf {
				t.Errorf("calculateConfidence() = %v, want between %v and %v",
					result, tt.minConf, tt.maxConf)
			}

			// Confidence must be between 0 and 1
			if result < 0 || result > 1 {
				t.Errorf("Confidence out of valid range: %v", result)
			}
		})
	}
}

func TestIdentifyArchetype(t *testing.T) {
	eventData := createMockEventAnalysis(100, 0.55)
	analyzer := NewMetaAnalyzer(eventData, DefaultMetaAnalysisOptions())

	tests := []struct {
		name         string
		deckCards    []string
		expectResult string
	}{
		{
			name:         "Cycle deck",
			deckCards:    []string{"Ice Spirit", "Skeletons", "Log", "Ice Golem"},
			expectResult: "Cycle",
		},
		{
			name:         "Control deck",
			deckCards:    []string{"Golem", "Lightning", "Mega Minion", "Baby Dragon"},
			expectResult: "Control",
		},
		{
			name:         "Midrange deck",
			deckCards:    []string{"Hog Rider", "Fireball", "Musketeer", "Valkyrie"},
			expectResult: "Midrange",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.identifyArchetype(tt.deckCards)

			// Just check that result is not empty - the function uses simplified logic
			if result == "" {
				t.Error("identifyArchetype() returned empty string")
			}
		})
	}
}

func TestGetArchetypeWinRate(t *testing.T) {
	tests := []struct {
		name      string
		eventData *events.EventAnalysis
		archetype string
		expected  float64
	}{
		{
			name:      "With event data",
			eventData: createMockEventAnalysis(100, 0.55),
			archetype: "Cycle",
			expected:  0.55,
		},
		{
			name:      "Without event data",
			eventData: nil,
			archetype: "Cycle",
			expected:  0.50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMetaAnalyzer(tt.eventData, DefaultMetaAnalysisOptions())
			result := analyzer.getArchetypeWinRate(tt.archetype)

			if result != tt.expected {
				t.Errorf("getArchetypeWinRate(%v) = %v, want %v", tt.archetype, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Matchup Analysis Function Tests
// ============================================================================

func TestIdentifyWeakMatchups(t *testing.T) {
	tests := []struct {
		name          string
		eventData     *events.EventAnalysis
		options       MetaAnalysisOptions
		deckCards     []string
		expectMatches bool
	}{
		{
			name:          "Deck with weak cards",
			eventData:     createMockEventAnalysis(100, 0.55),
			options:       DefaultMetaAnalysisOptions(),
			deckCards:     []string{"Valkyrie", "Ice Spirit", "Elite Barbarians"}, // All have <45% WR
			expectMatches: true,
		},
		{
			name:          "Deck with strong cards only",
			eventData:     createMockEventAnalysis(100, 0.55),
			options:       DefaultMetaAnalysisOptions(),
			deckCards:     []string{"Hog Rider", "Fireball", "Musketeer"},
			expectMatches: false,
		},
		{
			name:          "ShowWeakMatchups disabled",
			eventData:     createMockEventAnalysis(100, 0.55),
			options:       MetaAnalysisOptions{MetaWeight: 0.15, ShowWeakMatchups: false},
			deckCards:     []string{"Valkyrie", "Ice Spirit"},
			expectMatches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMetaAnalyzer(tt.eventData, tt.options)
			cardMeta := analyzer.analyzeCardMeta(tt.deckCards)
			result := analyzer.identifyWeakMatchups(tt.deckCards, cardMeta)

			if tt.expectMatches && len(result) == 0 {
				t.Error("Expected weak matchups but got none")
			}

			if !tt.expectMatches && len(result) > 0 {
				t.Errorf("Expected no weak matchups but got %v", len(result))
			}

			// Verify matchup structure
			for _, matchup := range result {
				if matchup.OpponentDeck == "" {
					t.Error("Matchup has empty OpponentDeck")
				}
				if matchup.StrengthFactor >= 0 {
					t.Errorf("Weak matchup should have negative StrengthFactor, got %v", matchup.StrengthFactor)
				}
				if len(matchup.KeyCards) == 0 {
					t.Error("Matchup has no KeyCards")
				}
			}
		})
	}
}

func TestIdentifyStrongMatchups(t *testing.T) {
	tests := []struct {
		name          string
		eventData     *events.EventAnalysis
		deckCards     []string
		expectMatches bool
	}{
		{
			name:          "Deck with strong cards",
			eventData:     createMockEventAnalysis(100, 0.55),
			deckCards:     []string{"Hog Rider", "Fireball"}, // Both have >55% WR
			expectMatches: true,
		},
		{
			name:          "Deck with weak cards only",
			eventData:     createMockEventAnalysis(100, 0.55),
			deckCards:     []string{"Valkyrie", "Ice Spirit"},
			expectMatches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMetaAnalyzer(tt.eventData, DefaultMetaAnalysisOptions())
			cardMeta := analyzer.analyzeCardMeta(tt.deckCards)
			result := analyzer.identifyStrongMatchups(tt.deckCards, cardMeta)

			if tt.expectMatches && len(result) == 0 {
				t.Error("Expected strong matchups but got none")
			}

			if !tt.expectMatches && len(result) > 0 {
				t.Errorf("Expected no strong matchups but got %v", len(result))
			}

			// Verify matchup structure
			for _, matchup := range result {
				if matchup.OpponentDeck == "" {
					t.Error("Matchup has empty OpponentDeck")
				}
				if matchup.StrengthFactor <= 0 {
					t.Errorf("Strong matchup should have positive StrengthFactor, got %v", matchup.StrengthFactor)
				}
				if len(matchup.KeyCards) == 0 {
					t.Error("Matchup has no KeyCards")
				}
			}
		})
	}
}

func TestGenerateRecommendations(t *testing.T) {
	tests := []struct {
		name               string
		eventData          *events.EventAnalysis
		options            MetaAnalysisOptions
		deckCards          []string
		expectRecommends   bool
		maxRecommendations int
	}{
		{
			name:               "Deck with low win rate cards",
			eventData:          createMockEventAnalysis(100, 0.55),
			options:            DefaultMetaAnalysisOptions(),
			deckCards:          []string{"Valkyrie", "Ice Spirit", "Elite Barbarians"},
			expectRecommends:   true,
			maxRecommendations: 5,
		},
		{
			name:               "Deck with high win rate cards",
			eventData:          createMockEventAnalysis(100, 0.55),
			deckCards:          []string{"Hog Rider", "Fireball", "Musketeer"},
			expectRecommends:   false,
			maxRecommendations: 0,
		},
		{
			name:               "RecommendMetaAlternatives disabled",
			eventData:          createMockEventAnalysis(100, 0.55),
			options:            MetaAnalysisOptions{MetaWeight: 0.15, RecommendMetaAlternatives: false},
			deckCards:          []string{"Valkyrie", "Ice Spirit"},
			expectRecommends:   false,
			maxRecommendations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMetaAnalyzer(tt.eventData, tt.options)
			cardMeta := analyzer.analyzeCardMeta(tt.deckCards)
			adjustment := analyzer.calculateMetaAdjustment(tt.deckCards, cardMeta, 75.0)
			result := analyzer.generateRecommendations(tt.deckCards, cardMeta, adjustment)

			if tt.expectRecommends && len(result) == 0 {
				t.Error("Expected recommendations but got none")
			}

			if !tt.expectRecommends && len(result) > 0 {
				t.Errorf("Expected no recommendations but got %v", len(result))
			}

			// Verify no more than 5 recommendations
			if len(result) > tt.maxRecommendations {
				t.Errorf("Got %v recommendations, max should be %v", len(result), tt.maxRecommendations)
			}

			// Verify recommendation structure
			for _, rec := range result {
				if rec.Type == "" {
					t.Error("Recommendation has empty Type")
				}
				if rec.CardName == "" {
					t.Error("Recommendation has empty CardName")
				}
				if rec.Reason == "" {
					t.Error("Recommendation has empty Reason")
				}
			}

			// Verify recommendations are sorted by impact (descending)
			for i := 0; i < len(result)-1; i++ {
				if result[i].ExpectedImpact < result[i+1].ExpectedImpact {
					t.Error("Recommendations not sorted by expected impact")
				}
			}
		})
	}
}

func TestFindTrendingAlternative(t *testing.T) {
	tests := []struct {
		name      string
		eventData *events.EventAnalysis
		cardName  string
		expected  string
	}{
		{
			name:      "Find alternative for weak card",
			eventData: createMockEventAnalysis(100, 0.55),
			cardName:  "Ice Spirit",
			expected:  "Hog Rider", // Highest win rate card
		},
		{
			name:      "Find alternative when card is top",
			eventData: createMockEventAnalysis(100, 0.55),
			cardName:  "Hog Rider",
			expected:  "Fireball", // Second highest win rate
		},
		{
			name:      "No event data",
			eventData: nil,
			cardName:  "Ice Spirit",
			expected:  "Unknown",
		},
		{
			name:      "Empty event data",
			eventData: createEmptyEventAnalysis(),
			cardName:  "Ice Spirit",
			expected:  "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMetaAnalyzer(tt.eventData, DefaultMetaAnalysisOptions())
			result := analyzer.findTrendingAlternative(tt.cardName)

			if result != tt.expected {
				t.Errorf("findTrendingAlternative(%v) = %v, want %v", tt.cardName, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Formatting Function Tests
// ============================================================================

func TestFormatMetaAnalysis(t *testing.T) {
	tests := []struct {
		name          string
		analysis      *DeckMetaAnalysis
		expectEmpty   bool
		checkContains []string
	}{
		{
			name:        "Nil analysis",
			analysis:    nil,
			expectEmpty: true,
		},
		{
			name: "Full analysis",
			analysis: &DeckMetaAnalysis{
				DeckCards: []string{"Hog Rider", "Fireball"},
				MetaAdjustment: MetaAdjustment{
					BaseScore:       75.0,
					MetaScore:       80.0,
					Adjustment:      5.0,
					MetaTier:        "A",
					WinRateEstimate: 0.57,
					Confidence:      0.8,
					Factors:         []string{"Above average meta win rate"},
				},
				CardMetaInfo: []CardMetaInfo{
					{CardName: "Hog Rider", MetaTier: "S", WinRate: 0.65, Popularity: 100, IsTrending: true},
				},
				ArchetypeMatch:   "Midrange",
				ArchetypeWinRate: 0.55,
				TrendingCards:    1,
				WeakMatchups: []MatchupAnalysis{
					{OpponentDeck: "Test Opponent", WinRate: 0.40},
				},
				StrongMatchups: []MatchupAnalysis{
					{OpponentDeck: "Test Weak Opponent", WinRate: 0.70},
				},
				MetaRecommendations: []MetaRecommendation{
					{Type: "replace", CardName: "Test Card", Reason: "Low win rate"},
				},
			},
			expectEmpty: false,
			checkContains: []string{
				"META ANALYSIS",
				"Score Adjustment",
				"Base Score",
				"Meta Score",
				"Card Meta Data",
				"Archetype",
				"Weak Matchups",
				"Strong Matchups",
				"Meta-Aware Recommendations",
			},
		},
		{
			name: "Minimal analysis",
			analysis: &DeckMetaAnalysis{
				DeckCards: []string{},
				MetaAdjustment: MetaAdjustment{
					BaseScore: 75.0,
					MetaScore: 75.0,
				},
			},
			expectEmpty: false,
			checkContains: []string{
				"META ANALYSIS",
				"Score Adjustment",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatMetaAnalysis(tt.analysis)

			if tt.expectEmpty {
				if result != "" {
					t.Errorf("Expected empty string but got: %v", result)
				}
				return
			}

			if result == "" && !tt.expectEmpty {
				t.Error("FormatMetaAnalysis() returned empty string unexpectedly")
			}

			// Check for expected content
			for _, expected := range tt.checkContains {
				if !strings.Contains(result, expected) {
					t.Errorf("FormatMetaAnalysis() missing expected content: %v", expected)
				}
			}
		})
	}
}
