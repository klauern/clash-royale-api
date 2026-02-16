package leaderboard

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// createTestStorage creates a temporary test database
func createTestStorage(t *testing.T) (*Storage, func()) {
	t.Helper()

	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "leaderboard_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	storage, err := NewStorage("#TEST123")
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		storage.Close()
		os.Setenv("HOME", originalHome)
		os.RemoveAll(tmpDir)
	}

	return storage, cleanup
}

// createTestDeckEntry creates a test deck entry
func createTestDeckEntry(cards []string, overallScore float64) *DeckEntry {
	return &DeckEntry{
		Cards:             cards,
		OverallScore:      overallScore,
		AttackScore:       7.5,
		DefenseScore:      8.0,
		SynergyScore:      7.0,
		VersatilityScore:  6.5,
		F2PScore:          8.5,
		PlayabilityScore:  7.8,
		Archetype:         "beatdown",
		ArchetypeConf:     0.85,
		Strategy:          "balanced",
		AvgElixir:         3.8,
		EvaluatedAt:       time.Now(),
		PlayerTag:         "#TEST123",
		EvaluationVersion: "1.0.0",
	}
}

func TestNewStorage(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	if storage == nil {
		t.Fatal("expected storage to be created")
	}

	// Verify database file exists
	dbPath := storage.GetDBPath()
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("database file does not exist: %s", dbPath)
	}

	// Verify path format
	expectedDir := filepath.Join(os.Getenv("HOME"), ".cr-api", "leaderboards")
	expectedPath := filepath.Join(expectedDir, "TEST123.db")
	if dbPath != expectedPath {
		t.Errorf("expected db path %s, got %s", expectedPath, dbPath)
	}
}

func TestInsertDeck_NewDeck(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	entry := createTestDeckEntry(
		[]string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Ice Spirit"},
		8.5,
	)

	id, isNew, err := storage.InsertDeck(entry)
	if err != nil {
		t.Fatalf("failed to insert deck: %v", err)
	}

	if !isNew {
		t.Error("expected new deck insertion")
	}

	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}

	if entry.ID != id {
		t.Errorf("expected entry.ID to be set to %d, got %d", id, entry.ID)
	}

	// Verify deck hash was computed
	if entry.DeckHash == "" {
		t.Error("expected deck hash to be computed")
	}
}

func TestInsertDeck_Deduplication(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	cards1 := []string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Ice Spirit"}
	cards2 := []string{"Ice Spirit", "Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang"} // Same cards, different order

	entry1 := createTestDeckEntry(cards1, 8.5)
	entry2 := createTestDeckEntry(cards2, 9.0) // Different score

	// Insert first deck
	id1, isNew1, err := storage.InsertDeck(entry1)
	if err != nil {
		t.Fatalf("failed to insert first deck: %v", err)
	}
	if !isNew1 {
		t.Error("expected first insertion to be new")
	}

	// Insert second deck (same cards, different order)
	id2, isNew2, err := storage.InsertDeck(entry2)
	if err != nil {
		t.Fatalf("failed to insert second deck: %v", err)
	}
	if isNew2 {
		t.Error("expected second insertion to be update, not new")
	}

	if id1 != id2 {
		t.Errorf("expected same id for duplicate deck, got %d and %d", id1, id2)
	}

	// Verify updated score
	entries, err := storage.GetTopN(1)
	if err != nil {
		t.Fatalf("failed to query deck: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 deck, got %d", len(entries))
	}
	if entries[0].OverallScore != 9.0 {
		t.Errorf("expected score 9.0, got %f", entries[0].OverallScore)
	}
}

func TestComputeDeckHash_Consistency(t *testing.T) {
	cards1 := []string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Ice Spirit"}
	cards2 := []string{"Ice Spirit", "Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang"}
	cards3 := []string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Skeleton Army"} // Different deck

	hash1 := deck.DeckHash(cards1)
	hash2 := deck.DeckHash(cards2)
	hash3 := deck.DeckHash(cards3)

	if hash1 != hash2 {
		t.Errorf("expected same hash for same cards in different order")
	}

	if hash1 == hash3 {
		t.Errorf("expected different hash for different cards")
	}

	// Verify hash format (64 hex characters for SHA256)
	if len(hash1) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash1))
	}
}

func TestQuery_TopN(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Insert multiple decks with different scores
	decks := []*DeckEntry{
		createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", "H"}, 9.5),
		createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", "I"}, 8.0),
		createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", "J"}, 7.5),
		createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", "K"}, 6.0),
		createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", "L"}, 5.5),
	}

	for _, deck := range decks {
		if _, _, err := storage.InsertDeck(deck); err != nil {
			t.Fatalf("failed to insert deck: %v", err)
		}
	}

	// Query top 3
	results, err := storage.GetTopN(3)
	if err != nil {
		t.Fatalf("failed to query top N: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Verify descending order
	if results[0].OverallScore != 9.5 {
		t.Errorf("expected first score 9.5, got %f", results[0].OverallScore)
	}
	if results[1].OverallScore != 8.0 {
		t.Errorf("expected second score 8.0, got %f", results[1].OverallScore)
	}
	if results[2].OverallScore != 7.5 {
		t.Errorf("expected third score 7.5, got %f", results[2].OverallScore)
	}
}

func TestQuery_WithFilters(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Insert decks with different archetypes
	deck1 := createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", "H"}, 9.0)
	deck1.Archetype = "beatdown"

	deck2 := createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", "I"}, 8.5)
	deck2.Archetype = "control"

	deck3 := createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", "J"}, 8.0)
	deck3.Archetype = "beatdown"

	for _, deck := range []*DeckEntry{deck1, deck2, deck3} {
		if _, _, err := storage.InsertDeck(deck); err != nil {
			t.Fatalf("failed to insert deck: %v", err)
		}
	}

	// Query beatdown decks only
	opts := QueryOptions{
		Archetype: "beatdown",
		SortBy:    "overall_score",
		SortOrder: "desc",
	}
	results, err := storage.Query(opts)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 beatdown decks, got %d", len(results))
	}

	for _, result := range results {
		if result.Archetype != "beatdown" {
			t.Errorf("expected beatdown archetype, got %s", result.Archetype)
		}
	}
}

func TestQuery_ScoreRange(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Insert decks with various scores
	scores := []float64{9.5, 8.5, 7.5, 6.5, 5.5}
	for i, score := range scores {
		deck := createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", string(rune('H' + i))}, score)
		if _, _, err := storage.InsertDeck(deck); err != nil {
			t.Fatalf("failed to insert deck: %v", err)
		}
	}

	// Query decks with score between 7.0 and 9.0
	opts := QueryOptions{
		MinScore:  7.0,
		MaxScore:  9.0,
		SortBy:    "overall_score",
		SortOrder: "desc",
	}
	results, err := storage.Query(opts)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 decks in range, got %d", len(results))
	}

	// Verify all results are in range
	for _, result := range results {
		if result.OverallScore < 7.0 || result.OverallScore > 9.0 {
			t.Errorf("score %f out of range [7.0, 9.0]", result.OverallScore)
		}
	}
}

func TestQuery_CardFilters(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Insert decks with different card combinations
	deck1 := createTestDeckEntry([]string{"Giant", "Wizard", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Ice Spirit"}, 9.0)
	deck2 := createTestDeckEntry([]string{"Giant", "Witch", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Skeleton Army", "Ice Spirit"}, 8.5)
	deck3 := createTestDeckEntry([]string{"Hog Rider", "Valkyrie", "Mini P.E.K.K.A", "Musketeer", "Arrows", "Fireball", "Goblin Gang", "Ice Spirit"}, 8.0)

	for _, deck := range []*DeckEntry{deck1, deck2, deck3} {
		if _, _, err := storage.InsertDeck(deck); err != nil {
			t.Fatalf("failed to insert deck: %v", err)
		}
	}

	t.Run("RequireAllCards", func(t *testing.T) {
		opts := QueryOptions{
			RequireAllCards: []string{"Giant", "Wizard"},
			SortBy:          "overall_score",
			SortOrder:       "desc",
		}
		results, err := storage.Query(opts)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 deck with Giant and Wizard, got %d", len(results))
		}
		if results[0].OverallScore != 9.0 {
			t.Errorf("expected score 9.0, got %f", results[0].OverallScore)
		}
	})

	t.Run("RequireAnyCards", func(t *testing.T) {
		opts := QueryOptions{
			RequireAnyCards: []string{"Wizard", "Witch"},
			SortBy:          "overall_score",
			SortOrder:       "desc",
		}
		results, err := storage.Query(opts)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("expected 2 decks with Wizard or Witch, got %d", len(results))
		}
	})

	t.Run("ExcludeCards", func(t *testing.T) {
		opts := QueryOptions{
			ExcludeCards: []string{"Giant"},
			SortBy:       "overall_score",
			SortOrder:    "desc",
		}
		results, err := storage.Query(opts)
		if err != nil {
			t.Fatalf("failed to query: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 deck without Giant, got %d", len(results))
		}
		if results[0].OverallScore != 8.0 {
			t.Errorf("expected score 8.0 (Hog Rider deck), got %f", results[0].OverallScore)
		}
	})
}

func TestDeleteDeck(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	entry := createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", "H"}, 8.5)
	id, _, err := storage.InsertDeck(entry)
	if err != nil {
		t.Fatalf("failed to insert deck: %v", err)
	}

	// Delete the deck
	if err := storage.DeleteDeck(id); err != nil {
		t.Fatalf("failed to delete deck: %v", err)
	}

	// Verify it's gone
	count, err := storage.Count()
	if err != nil {
		t.Fatalf("failed to count decks: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 decks after deletion, got %d", count)
	}
}

func TestClear(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Insert multiple decks
	for i := 0; i < 5; i++ {
		entry := createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", string(rune('H' + i))}, 8.0)
		if _, _, err := storage.InsertDeck(entry); err != nil {
			t.Fatalf("failed to insert deck: %v", err)
		}
	}

	// Clear all decks
	if err := storage.Clear(); err != nil {
		t.Fatalf("failed to clear decks: %v", err)
	}

	// Verify all gone
	count, err := storage.Count()
	if err != nil {
		t.Fatalf("failed to count decks: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 decks after clear, got %d", count)
	}
}

func TestGetStats_Empty(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	stats, err := storage.GetStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.PlayerTag != "#TEST123" {
		t.Errorf("expected player tag #TEST123, got %s", stats.PlayerTag)
	}

	if stats.TotalUniqueDecks != 0 {
		t.Errorf("expected 0 unique decks, got %d", stats.TotalUniqueDecks)
	}
}

func TestRecalculateStats(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Insert decks
	scores := []float64{9.5, 8.5, 7.5}
	for i, score := range scores {
		entry := createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", string(rune('H' + i))}, score)
		if _, _, err := storage.InsertDeck(entry); err != nil {
			t.Fatalf("failed to insert deck: %v", err)
		}
	}

	// Recalculate stats
	stats, err := storage.RecalculateStats()
	if err != nil {
		t.Fatalf("failed to recalculate stats: %v", err)
	}

	if stats.TotalUniqueDecks != 3 {
		t.Errorf("expected 3 unique decks, got %d", stats.TotalUniqueDecks)
	}

	if stats.TopScore != 9.5 {
		t.Errorf("expected top score 9.5, got %f", stats.TopScore)
	}

	expectedAvg := (9.5 + 8.5 + 7.5) / 3
	if stats.AvgScore != expectedAvg {
		t.Errorf("expected avg score %f, got %f", expectedAvg, stats.AvgScore)
	}

	// Verify stats were persisted
	loadedStats, err := storage.GetStats()
	if err != nil {
		t.Fatalf("failed to load stats: %v", err)
	}

	if loadedStats.TotalUniqueDecks != stats.TotalUniqueDecks {
		t.Errorf("persisted stats mismatch: expected %d unique decks, got %d",
			stats.TotalUniqueDecks, loadedStats.TotalUniqueDecks)
	}
}

func TestCount(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Initially empty
	count, err := storage.Count()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 decks, got %d", count)
	}

	// Insert decks
	for i := 0; i < 5; i++ {
		entry := createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", string(rune('H' + i))}, 8.0)
		if _, _, err := storage.InsertDeck(entry); err != nil {
			t.Fatalf("failed to insert deck: %v", err)
		}
	}

	// Verify count
	count, err = storage.Count()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 decks, got %d", count)
	}
}

func TestGetByArchetype(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Insert decks with different archetypes
	archetypes := []string{"beatdown", "control", "beatdown", "cycle", "beatdown"}
	for i, archetype := range archetypes {
		entry := createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", string(rune('H' + i))}, 8.0+float64(i)*0.1)
		entry.Archetype = archetype
		if _, _, err := storage.InsertDeck(entry); err != nil {
			t.Fatalf("failed to insert deck: %v", err)
		}
	}

	// Get beatdown decks
	results, err := storage.GetByArchetype("beatdown", 10)
	if err != nil {
		t.Fatalf("failed to get by archetype: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 beatdown decks, got %d", len(results))
	}

	for _, result := range results {
		if result.Archetype != "beatdown" {
			t.Errorf("expected beatdown archetype, got %s", result.Archetype)
		}
	}
}

func TestQuery_Pagination(t *testing.T) {
	storage, cleanup := createTestStorage(t)
	defer cleanup()

	// Insert 10 decks
	for i := 0; i < 10; i++ {
		entry := createTestDeckEntry([]string{"A", "B", "C", "D", "E", "F", "G", string(rune('H' + i))}, float64(10-i))
		if _, _, err := storage.InsertDeck(entry); err != nil {
			t.Fatalf("failed to insert deck: %v", err)
		}
	}

	// First page (3 items)
	page1, err := storage.Query(QueryOptions{
		Limit:     3,
		Offset:    0,
		SortBy:    "overall_score",
		SortOrder: "desc",
	})
	if err != nil {
		t.Fatalf("failed to query page 1: %v", err)
	}
	if len(page1) != 3 {
		t.Fatalf("expected 3 items in page 1, got %d", len(page1))
	}

	// Second page (3 items)
	page2, err := storage.Query(QueryOptions{
		Limit:     3,
		Offset:    3,
		SortBy:    "overall_score",
		SortOrder: "desc",
	})
	if err != nil {
		t.Fatalf("failed to query page 2: %v", err)
	}
	if len(page2) != 3 {
		t.Fatalf("expected 3 items in page 2, got %d", len(page2))
	}

	// Verify no overlap
	if page1[2].OverallScore <= page2[0].OverallScore {
		t.Errorf("pages overlap: page1 last=%f, page2 first=%f",
			page1[2].OverallScore, page2[0].OverallScore)
	}
}
