package fuzzstorage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestUpdateDeck(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "fuzz_test.db")
	storage, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	entry := &DeckEntry{
		Cards:            []string{"Knight", "Archers", "Fireball", "Zap", "Cannon", "Hog Rider", "Ice Spirit", "Skeletons"},
		OverallScore:     5.0,
		AttackScore:      5.0,
		DefenseScore:     5.0,
		SynergyScore:     5.0,
		VersatilityScore: 5.0,
		AvgElixir:        3.0,
		Archetype:        "cycle",
		ArchetypeConf:    0.5,
		EvaluatedAt:      time.Now().Add(-time.Hour),
		RunID:            "seed",
	}

	id, isNew, err := storage.InsertDeck(entry)
	if err != nil {
		t.Fatalf("failed to insert deck: %v", err)
	}
	if !isNew {
		t.Fatalf("expected new deck insert")
	}

	updatedAt := time.Now()
	entry.ID = id
	entry.OverallScore = 7.5
	entry.AttackScore = 7.0
	entry.DefenseScore = 6.5
	entry.SynergyScore = 7.2
	entry.VersatilityScore = 6.8
	entry.AvgElixir = 3.1
	entry.Archetype = "fast_cycle"
	entry.ArchetypeConf = 0.9
	entry.EvaluatedAt = updatedAt
	entry.RunID = "update"

	if err := storage.UpdateDeck(entry); err != nil {
		t.Fatalf("failed to update deck: %v", err)
	}

	entries, err := storage.Query(QueryOptions{Limit: 10})
	if err != nil {
		t.Fatalf("failed to query decks: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 deck entry, got %d", len(entries))
	}

	got := entries[0]
	if got.ID != id {
		t.Fatalf("expected ID %d, got %d", id, got.ID)
	}
	if got.OverallScore != entry.OverallScore {
		t.Errorf("expected overall score %.2f, got %.2f", entry.OverallScore, got.OverallScore)
	}
	if got.Archetype != entry.Archetype {
		t.Errorf("expected archetype %q, got %q", entry.Archetype, got.Archetype)
	}
	if got.RunID != entry.RunID {
		t.Errorf("expected run ID %q, got %q", entry.RunID, got.RunID)
	}
	if got.EvaluatedAt.IsZero() {
		t.Errorf("expected evaluated_at to be set")
	}
}
