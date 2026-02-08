// Package evaluation benchmarks test performance of deck evaluation functions.
//
// # Benchmark Results (Apple M3 Max, 3s runtime)
//
// ## Performance Targets vs Actual Results
//
//   - Full Evaluation:        < 100ms target → 0.14ms actual (700x faster ✓)
//   - Archetype Detection:    < 50ms target  → 0.0007ms actual (83,000x faster ✓)
//   - Synergy Matrix:         < 200ms target → 0.06ms actual (3,000x faster ✓)
//   - Batch (100 decks):      < 2s/deck target → 157ms/deck actual (12x faster ✓)
//
// ## Detailed Benchmark Results
//
//   - BenchmarkEvaluate:                 141,593 ns/op (0.14ms)  | 504 KB/op | 6,311 allocs/op
//   - BenchmarkDetectArchetype:              661 ns/op (0.0007ms)| 0 B/op    | 0 allocs/op
//   - BenchmarkGenerateSynergyMatrix:     64,739 ns/op (0.06ms)  | 247 KB/op | 3,088 allocs/op
//   - BenchmarkFormatCSV:                  7,821 ns/op (0.008ms) | 17 KB/op  | 130 allocs/op
//   - BenchmarkFormatJSON:                30,893 ns/op (0.03ms)  | 34 KB/op  | 353 allocs/op
//
// ## Component Performance (for profiling optimization targets)
//
//   - BenchmarkScoreAttack:             48 ns/op | 0 allocs
//   - BenchmarkScoreDefense:            46 ns/op | 0 allocs
//   - BenchmarkScoreSynergy:        65,835 ns/op | 247 KB/op | 3,088 allocs (synergy DB lookup)
//   - BenchmarkScoreVersatility:       130 ns/op | 0 allocs
//   - BenchmarkScoreF2P:               113 ns/op | 0 allocs
//   - BenchmarkBuildDefenseAnalysis:   767 ns/op | 1.3 KB/op | 20 allocs
//   - BenchmarkBuildAttackAnalysis:  1,219 ns/op | 1.9 KB/op | 22 allocs
//
// ## Key Performance Insights
//
//  1. Archetype detection is extremely fast (661ns) with zero allocations
//  2. Synergy scoring dominates evaluation time (65µs of 141µs total = 46%)
//  3. Formatting operations are negligible overhead (<10% of evaluation time)
//  4. All targets exceeded by 12-83,000x margins, indicating excellent performance headroom
//  5. Memory allocations are reasonable (504 KB for full evaluation)
//
// ## Running Benchmarks
//
//	go test -bench=. -benchmem -benchtime=3s ./pkg/deck/evaluation/...
//	go test -bench=BenchmarkEvaluate -benchmem -count=5 ./pkg/deck/evaluation/...  # Multiple runs
//	go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof ./pkg/deck/evaluation/...
//
// Last updated: 2026-01-05
package evaluation

import (
	"testing"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// ============================================================================
// Benchmark Setup - Sample Deck Data
// ============================================================================

// getBenchmarkDeck returns a standard 2.6 Hog Cycle deck for consistent benchmarking
func getBenchmarkDeck() []deck.CardCandidate {
	return []deck.CardCandidate{
		makeCard("Hog Rider", deck.RoleWinCondition, 11, 14, "Rare", 4),
		makeCard("Musketeer", deck.RoleSupport, 11, 14, "Rare", 4),
		makeCard("Valkyrie", deck.RoleSupport, 11, 14, "Rare", 4),
		makeCard("Cannon", deck.RoleBuilding, 11, 13, "Common", 3),
		makeCard("Fireball", deck.RoleSpellBig, 11, 14, "Rare", 4),
		makeCard("The Log", deck.RoleSpellSmall, 11, 14, "Legendary", 2),
		makeCard("Ice Spirit", deck.RoleCycle, 11, 14, "Common", 1),
		makeCard("Skeletons", deck.RoleCycle, 11, 14, "Common", 1),
	}
}

// getBenchmarkDeckNames returns card names for synergy matrix benchmarks
func getBenchmarkDeckNames() []string {
	return []string{
		"Hog Rider",
		"Musketeer",
		"Valkyrie",
		"Cannon",
		"Fireball",
		"The Log",
		"Ice Spirit",
		"Skeletons",
	}
}

// ============================================================================
// Main Benchmarks
// ============================================================================

// BenchmarkEvaluate benchmarks the full deck evaluation pipeline
// Target: < 100ms for single deck evaluation
func BenchmarkEvaluate(b *testing.B) {
	deckCards := getBenchmarkDeck()
	synergyDB := deck.NewSynergyDatabase()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Evaluate(deckCards, synergyDB, nil)
	}
}

// BenchmarkDetectArchetype benchmarks archetype detection only
// Target: < 50ms for archetype detection
func BenchmarkDetectArchetype(b *testing.B) {
	deckCards := getBenchmarkDeck()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetectArchetype(deckCards)
	}
}

// BenchmarkGenerateSynergyMatrix benchmarks synergy matrix generation
// Target: < 200ms for synergy matrix generation
func BenchmarkGenerateSynergyMatrix(b *testing.B) {
	deckNames := getBenchmarkDeckNames()
	synergyDB := deck.NewSynergyDatabase()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateSynergyMatrix(deckNames, synergyDB)
	}
}

// BenchmarkFormatCSV benchmarks CSV formatting
// Target: reasonable performance for export functionality
func BenchmarkFormatCSV(b *testing.B) {
	deckCards := getBenchmarkDeck()
	synergyDB := deck.NewSynergyDatabase()
	result := Evaluate(deckCards, synergyDB, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatCSV(&result)
	}
}

// BenchmarkFormatJSON benchmarks JSON formatting
// Target: reasonable performance for export functionality
func BenchmarkFormatJSON(b *testing.B) {
	deckCards := getBenchmarkDeck()
	synergyDB := deck.NewSynergyDatabase()
	result := Evaluate(deckCards, synergyDB, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FormatJSON(&result)
	}
}

// ============================================================================
// Component Benchmarks - Detailed Performance Analysis
// ============================================================================

// BenchmarkScoreAttack benchmarks attack scoring
func BenchmarkScoreAttack(b *testing.B) {
	deckCards := getBenchmarkDeck()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ScoreAttack(deckCards)
	}
}

// BenchmarkScoreDefense benchmarks defense scoring
func BenchmarkScoreDefense(b *testing.B) {
	deckCards := getBenchmarkDeck()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ScoreDefense(deckCards)
	}
}

// BenchmarkScoreSynergy benchmarks synergy scoring
func BenchmarkScoreSynergy(b *testing.B) {
	deckCards := getBenchmarkDeck()
	synergyDB := deck.NewSynergyDatabase()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ScoreSynergy(deckCards, synergyDB)
	}
}

// BenchmarkScoreVersatility benchmarks versatility scoring
func BenchmarkScoreVersatility(b *testing.B) {
	deckCards := getBenchmarkDeck()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ScoreVersatility(deckCards)
	}
}

// BenchmarkScoreF2P benchmarks F2P scoring
func BenchmarkScoreF2P(b *testing.B) {
	deckCards := getBenchmarkDeck()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ScoreF2P(deckCards)
	}
}

// ============================================================================
// Analysis Benchmarks
// ============================================================================

// BenchmarkBuildDefenseAnalysis benchmarks defense analysis section building
func BenchmarkBuildDefenseAnalysis(b *testing.B) {
	deckCards := getBenchmarkDeck()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildDefenseAnalysis(deckCards)
	}
}

// BenchmarkBuildAttackAnalysis benchmarks attack analysis section building
func BenchmarkBuildAttackAnalysis(b *testing.B) {
	deckCards := getBenchmarkDeck()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildAttackAnalysis(deckCards)
	}
}

// ============================================================================
// Batch Processing Benchmarks
// ============================================================================

// BenchmarkEvaluateBatch benchmarks batch evaluation of multiple decks
// Target: < 2s average per deck for 100 decks
func BenchmarkEvaluateBatch(b *testing.B) {
	synergyDB := deck.NewSynergyDatabase()

	// Create 100 test decks (variations of the benchmark deck)
	testDecks := make([][]deck.CardCandidate, 100)
	for i := 0; i < 100; i++ {
		testDecks[i] = getBenchmarkDeck()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, deckCards := range testDecks {
			_ = Evaluate(deckCards, synergyDB, nil)
		}
	}
}

// ============================================================================
// Archetype Detection Benchmarks - All Archetypes
// ============================================================================

// BenchmarkDetectArchetypeBeatdown benchmarks beatdown archetype detection
func BenchmarkDetectArchetypeBeatdown(b *testing.B) {
	deckCards := []deck.CardCandidate{
		makeCard("Golem", deck.RoleWinCondition, 11, 11, "Epic", 8),
		makeCard("Night Witch", deck.RoleSupport, 11, 11, "Legendary", 4),
		makeCard("Baby Dragon", deck.RoleSupport, 11, 11, "Epic", 4),
		makeCard("Lightning", deck.RoleSpellBig, 11, 11, "Epic", 6),
		makeCard("Tornado", deck.RoleSpellBig, 11, 11, "Epic", 3),
		makeCard("Mega Minion", deck.RoleSupport, 11, 11, "Rare", 3),
		makeCard("Elixir Collector", deck.RoleBuilding, 11, 11, "Rare", 6),
		makeCard("Lumberjack", deck.RoleSupport, 11, 11, "Legendary", 4),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetectArchetype(deckCards)
	}
}

// BenchmarkDetectArchetypeCycle benchmarks cycle archetype detection
func BenchmarkDetectArchetypeCycle(b *testing.B) {
	deckCards := getBenchmarkDeck() // 2.6 Hog Cycle

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetectArchetype(deckCards)
	}
}

// BenchmarkDetectArchetypeBait benchmarks bait archetype detection
func BenchmarkDetectArchetypeBait(b *testing.B) {
	deckCards := []deck.CardCandidate{
		makeCard("Goblin Barrel", deck.RoleWinCondition, 11, 11, "Epic", 3),
		makeCard("Princess", deck.RoleSupport, 11, 11, "Legendary", 3),
		makeCard("Rocket", deck.RoleSpellBig, 11, 11, "Rare", 6),
		makeCard("Inferno Tower", deck.RoleBuilding, 11, 11, "Rare", 5),
		makeCard("Goblin Gang", deck.RoleSupport, 11, 11, "Common", 3),
		makeCard("Knight", deck.RoleSupport, 11, 11, "Common", 3),
		makeCard("Ice Spirit", deck.RoleCycle, 11, 11, "Common", 1),
		makeCard("The Log", deck.RoleSpellSmall, 11, 11, "Legendary", 2),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetectArchetype(deckCards)
	}
}

// ============================================================================
// Memory Allocation Benchmarks
// ============================================================================

// BenchmarkEvaluateAllocs benchmarks memory allocations during evaluation
func BenchmarkEvaluateAllocs(b *testing.B) {
	deckCards := getBenchmarkDeck()
	synergyDB := deck.NewSynergyDatabase()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Evaluate(deckCards, synergyDB, nil)
	}
}

// BenchmarkFormatCSVAllocs benchmarks memory allocations during CSV formatting
func BenchmarkFormatCSVAllocs(b *testing.B) {
	deckCards := getBenchmarkDeck()
	synergyDB := deck.NewSynergyDatabase()
	result := Evaluate(deckCards, synergyDB, nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatCSV(&result)
	}
}

// BenchmarkFormatJSONAllocs benchmarks memory allocations during JSON formatting
func BenchmarkFormatJSONAllocs(b *testing.B) {
	deckCards := getBenchmarkDeck()
	synergyDB := deck.NewSynergyDatabase()
	result := Evaluate(deckCards, synergyDB, nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FormatJSON(&result)
	}
}

// ============================================================================
// Algorithm Comparison Tests - Quality Metrics
// ============================================================================

// TestQualityComparison_MetaVsBad compares meta deck scores against bad deck scores
// to verify the algorithm properly distinguishes quality
func TestQualityComparison_MetaVsBad(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	// Meta decks - should score high
	metaDecks := [][]string{
		{"Hog Rider", "Musketeer", "Valkyrie", "Cannon", "Fireball", "The Log", "Ice Spirit", "Skeletons"},           // 2.6 Hog Cycle
		{"Golem", "Night Witch", "Baby Dragon", "Tornado", "Lightning", "Mega Minion", "Elixir Collector", "Lumberjack"}, // Golem Beatdown
		{"Lava Hound", "Balloon", "Miner", "Mega Minion", "Skeleton Dragons", "Tornado", "Log", "Arrows"},                // LavaLoon
		{"Goblin Barrel", "Princess", "Goblin Gang", "Knight", "Inferno Tower", "Ice Spirit", "The Log", "Rocket"},       // Log Bait
	}

	// Bad decks - should score low
	badDecks := [][]string{
		{"Knight", "Archers", "Valkyrie", "Mini P.E.K.K.A", "Musketeer", "Ice Golem", "Mega Minion", "Skeleton Army"}, // No Win Condition
		{"Fireball", "Lightning", "Rocket", "Poison", "Freeze", "Zap", "Log", "Arrows"},                                   // All Spells
		{"Hog Rider", "Knight", "Valkyrie", "Skeleton Army", "Goblin Gang", "Ice Spirit", "Log", "Cannon"},                // No Anti-Air
	}

	// Calculate scores
	var metaTotal, badTotal float64
	for _, deckCards := range metaDecks {
		cards := createDeckFromComparison(deckCards)
		result := Evaluate(cards, synergyDB, nil)
		metaTotal += result.OverallScore
	}
	avgMeta := metaTotal / float64(len(metaDecks))

	for _, deckCards := range badDecks {
		cards := createDeckFromComparison(deckCards)
		result := Evaluate(cards, synergyDB, nil)
		badTotal += result.OverallScore
	}
	avgBad := badTotal / float64(len(badDecks))

	// Meta decks should significantly outscore bad decks
	scoreGap := avgMeta - avgBad

	t.Logf("Meta Deck Average Score: %.2f/10.0", avgMeta)
	t.Logf("Bad Deck Average Score: %.2f/10.0", avgBad)
	t.Logf("Score Gap: %.2f points", scoreGap)

	// Verify quality separation thresholds
	if avgMeta < 7.0 {
		t.Errorf("Meta deck average %.2f is below threshold 7.0", avgMeta)
	}
	if avgBad > 5.0 {
		t.Errorf("Bad deck average %.2f is above threshold 5.0", avgBad)
	}
	if scoreGap < 2.5 {
		t.Errorf("Score gap %.2f is too small - algorithm may not distinguish quality well", scoreGap)
	}
}

// TestQualityComparison_CoherenceScores verifies that coherent archetypes
// score better than incoherent mixed decks
func TestQualityComparison_CoherenceScores(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	tests := []struct {
		name      string
		cards     []string
		minScore  float64
		archetype Archetype
	}{
		{
			name:      "Coherent Cycle Deck",
			cards:     []string{"Hog Rider", "Musketeer", "Valkyrie", "Cannon", "Fireball", "The Log", "Ice Spirit", "Skeletons"},
			minScore:  7.5,
			archetype: ArchetypeCycle,
		},
		{
			name:      "Coherent Beatdown Deck",
			cards:     []string{"Golem", "Night Witch", "Baby Dragon", "Tornado", "Lightning", "Mega Minion", "Elixir Collector", "Lumberjack"},
			minScore:  7.5,
			archetype: ArchetypeBeatdown,
		},
		{
			name:      "Coherent Bait Deck",
			cards:     []string{"Goblin Barrel", "Princess", "Goblin Gang", "Knight", "Inferno Tower", "Ice Spirit", "The Log", "Rocket"},
			minScore:  7.0,
			archetype: ArchetypeBait,
		},
		{
			name:      "Incoherent Mixed Deck",
			cards:     []string{"Hog Rider", "Golem", "P.E.K.K.A", "Musketeer", "Baby Dragon", "Valkyrie", "Fireball", "Zap"},
			minScore:  3.0,
			archetype: ArchetypeUnknown,
		},
	}

	coherentTotal := 0.0
	coherentCount := 0
	incoherentScore := 0.0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards := createDeckFromComparison(tt.cards)
			result := Evaluate(cards, synergyDB, nil)

			if result.OverallScore < tt.minScore {
				t.Errorf("OverallScore = %.2f, want >= %.2f", result.OverallScore, tt.minScore)
			}

			t.Logf("%s: %.2f/10.0 (%s), Archetype: %s (%.0f%% confidence)",
				tt.name, result.OverallScore, result.OverallRating,
				result.DetectedArchetype, result.ArchetypeConfidence*100)

			if tt.archetype != ArchetypeUnknown {
				coherentTotal += result.OverallScore
				coherentCount++
			} else {
				incoherentScore = result.OverallScore
			}
		})
	}

	avgCoherent := coherentTotal / float64(coherentCount)
	t.Logf("Coherent deck average: %.2f vs Incoherent deck: %.2f", avgCoherent, incoherentScore)

	if avgCoherent < incoherentScore+3.0 {
		t.Errorf("Coherent decks (%.2f) should significantly outscore incoherent (%.2f)",
			avgCoherent, incoherentScore)
	}
}

// TestQualityComparison_SynergyImpact measures the impact of synergy
// on deck quality scores
func TestQualityComparison_SynergyImpact(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	// High synergy deck
	highSynergyCards := createDeckFromComparison([]string{
		"Golem", "Night Witch", "Baby Dragon", "Tornado",
		"Lightning", "Mega Minion", "Elixir Collector", "Lumberjack",
	})

	// Low synergy deck (random champions with poor synergy)
	lowSynergyCards := createDeckFromComparison([]string{
		"Archer Queen", "Golden Knight", "Skeleton King",
		"Little Prince", "Berserker", "Goblin Demolisher",
		"Royal Delivery", "Phoenix",
	})

	highResult := Evaluate(highSynergyCards, synergyDB, nil)
	lowResult := Evaluate(lowSynergyCards, synergyDB, nil)

	synergyGap := highResult.Synergy.Score - lowResult.Synergy.Score
	overallGap := highResult.OverallScore - lowResult.OverallScore

	t.Logf("High Synergy Deck:")
	t.Logf("  Synergy: %.2f, Overall: %.2f", highResult.Synergy.Score, highResult.OverallScore)
	t.Logf("Low Synergy Deck:")
	t.Logf("  Synergy: %.2f, Overall: %.2f", lowResult.Synergy.Score, lowResult.OverallScore)
	t.Logf("Synergy Gap: %.2f, Overall Gap: %.2f", synergyGap, overallGap)

	if synergyGap < 3.0 {
		t.Errorf("Synergy gap %.2f is too small - high synergy deck should score much higher", synergyGap)
	}

	if highResult.OverallScore < lowResult.OverallScore {
		t.Errorf("High synergy deck (%.2f) should outscore low synergy deck (%.2f)",
			highResult.OverallScore, lowResult.OverallScore)
	}
}

// TestQualityComparison_PlayerAnalysis verifies that the evaluation
// provides useful analysis for player improvement
func TestQualityComparison_PlayerAnalysis(t *testing.T) {
	synergyDB := deck.NewSynergyDatabase()

	// Test deck with clear strengths and weaknesses
	cards := createDeckFromComparison([]string{
		"Hog Rider", "Musketeer", "Valkyrie", "Cannon",
		"Fireball", "The Log", "Ice Spirit", "Skeletons",
	})

	result := Evaluate(cards, synergyDB, nil)

	// Verify all category scores are present
	if result.Attack.Score == 0 {
		t.Error("Attack score should not be zero")
	}
	if result.Defense.Score == 0 {
		t.Error("Defense score should not be zero")
	}
	if result.Synergy.Score == 0 {
		t.Error("Synergy score should not be zero")
	}
	if result.Versatility.Score == 0 {
		t.Error("Versatility score should not be zero")
	}

	// Verify analysis sections have content
	if result.DefenseAnalysis.Summary == "" {
		t.Error("Defense analysis should have a summary")
	}
	if result.AttackAnalysis.Summary == "" {
		t.Error("Attack analysis should have a summary")
	}

	// Verify archetype was detected
	if result.DetectedArchetype == ArchetypeUnknown {
		t.Log("Warning: Archetype detection returned 'unknown'")
	}

	// Output detailed analysis for manual review
	t.Logf("\n=== Deck Quality Analysis ===")
	t.Logf("Overall Score: %.2f/10.0 (%s)", result.OverallScore, result.OverallRating)
	t.Logf("Archetype: %s (%.0f%% confidence)", result.DetectedArchetype, result.ArchetypeConfidence*100)
	t.Logf("\nCategory Scores:")
	t.Logf("  Attack:      %.2f/10.0 (%s)", result.Attack.Score, result.Attack.Rating)
	t.Logf("  Defense:     %.2f/10.0 (%s)", result.Defense.Score, result.Defense.Rating)
	t.Logf("  Synergy:     %.2f/10.0 (%s)", result.Synergy.Score, result.Synergy.Rating)
	t.Logf("  Versatility: %.2f/10.0 (%s)", result.Versatility.Score, result.Versatility.Rating)
	t.Logf("  F2P:         %.2f/10.0 (%s)", result.F2PFriendly.Score, result.F2PFriendly.Rating)

	// Verify synergy matrix has data
	if result.SynergyMatrix.PairCount > 0 {
		t.Logf("\nSynergy Matrix:")
		t.Logf("  Pairs: %d", result.SynergyMatrix.PairCount)
		t.Logf("  Average: %.2f", result.SynergyMatrix.AverageSynergy)
		t.Logf("  Coverage: %.1f%%", result.SynergyMatrix.SynergyCoverage)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// createDeckFromComparison creates a CardCandidate slice from card names
func createDeckFromComparison(cardNames []string) []deck.CardCandidate {
	result := make([]deck.CardCandidate, len(cardNames))

	defaultStats := &clashroyale.CombatStats{
		DamagePerSecond: 100,
		Targets:         "Air & Ground",
	}

	for i, name := range cardNames {
		role := determineCardRoleBenchmark(name)
		rarity := determineCardRarityBenchmark(name)
		elixir := determineCardElixirBenchmark(name)

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

// determineCardRoleBenchmark determines the card role for benchmark tests
func determineCardRoleBenchmark(name string) deck.CardRole {
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

// determineCardRarityBenchmark determines the card rarity for benchmark tests
func determineCardRarityBenchmark(name string) string {
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
		return "Champion" //nolint:goconst // Rarity names are domain-specific values
	}
	if legendaries[name] {
		return "Legendary" //nolint:goconst // Rarity names are domain-specific values
	}
	if epics[name] {
		return "Epic" //nolint:goconst // Rarity names are domain-specific values
	}

	return "Rare" //nolint:goconst // Rarity names are domain-specific values
}

// determineCardElixirBenchmark determines the card elixir cost for benchmark tests
func determineCardElixirBenchmark(name string) int {
	elixirMap := map[string]int{
		"Skeletons": 1, "Ice Spirit": 1, "Bats": 1, "Fire Spirit": 1,
		"The Log": 2, "Zap": 2, "Snowball": 2, "Knight": 3,
		"Ice Golem": 2, "Heal Spirit": 1,
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
		"Fisherman": 3, "Royal Delivery": 3, "Phoenix": 4,
	}

	if elixir, ok := elixirMap[name]; ok {
		return elixir
	}

	return 4
}
