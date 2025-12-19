package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

func TestNewManager(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.dataDir != tempDir {
		t.Errorf("dataDir = %s, want %s", manager.dataDir, tempDir)
	}
	if manager.parser == nil {
		t.Error("parser should be initialized")
	}
}

func TestManager_EnsurePlayerDirectories(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	playerTag := "#TEST123"
	err := manager.ensurePlayerDirectories(playerTag)
	if err != nil {
		t.Fatalf("ensurePlayerDirectories failed: %v", err)
	}

	// Check that directories were created
	playerDir := manager.getPlayerEventDir(playerTag)
	dirs := []string{
		playerDir,
		filepath.Join(playerDir, "challenges"),
		filepath.Join(playerDir, "tournaments"),
		filepath.Join(playerDir, "special_events"),
		filepath.Join(playerDir, "aggregated"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory was not created: %s", dir)
		}
	}
}

func TestManager_SaveEventDeck(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create test event deck
	eventDeck := &EventDeck{
		EventID:   "test_event_001",
		PlayerTag: "#TEST123",
		EventName: "Grand Challenge",
		EventType: EventTypeGrandChallenge,
		StartTime: time.Now(),
		Deck: Deck{
			Cards: []CardInDeck{
				{Name: "Knight", Level: 11},
				{Name: "Archers", Level: 10},
				{Name: "Fireball", Level: 9},
				{Name: "Zap", Level: 11},
				{Name: "Giant", Level: 8},
				{Name: "Musketeer", Level: 9},
				{Name: "Valkyrie", Level: 10},
				{Name: "Mini P.E.K.K.A", Level: 9},
			},
			AvgElixir: 3.5,
		},
		Performance: EventPerformance{
			Wins:   5,
			Losses: 2,
		},
		Battles: []BattleRecord{},
	}

	// Save event deck
	err := manager.SaveEventDeck(eventDeck)
	if err != nil {
		t.Fatalf("SaveEventDeck failed: %v", err)
	}

	// Verify file was created
	playerDir := manager.getPlayerEventDir(eventDeck.PlayerTag)
	subdir := filepath.Join(playerDir, "challenges")
	files, err := filepath.Glob(filepath.Join(subdir, "*.json"))
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	if len(files) == 0 {
		t.Error("No files were created")
	}

	// Verify collection file was created
	collectionFile := filepath.Join(playerDir, "collection.json")
	if _, err := os.Stat(collectionFile); os.IsNotExist(err) {
		t.Error("Collection file was not created")
	}

	// Read and verify collection
	data, err := os.ReadFile(collectionFile)
	if err != nil {
		t.Fatalf("Failed to read collection file: %v", err)
	}

	var collection EventDeckCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("Failed to unmarshal collection: %v", err)
	}

	if collection.PlayerTag != eventDeck.PlayerTag {
		t.Errorf("Collection PlayerTag = %s, want %s", collection.PlayerTag, eventDeck.PlayerTag)
	}
	if len(collection.Decks) != 1 {
		t.Errorf("Collection has %d decks, want 1", len(collection.Decks))
	}
}

func TestManager_SaveEventDeck_DifferentTypes(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	tests := []struct {
		name      string
		eventType EventType
		subdir    string
	}{
		{
			name:      "Tournament",
			eventType: EventTypeTournament,
			subdir:    "tournaments",
		},
		{
			name:      "Special Event",
			eventType: EventTypeSpecialEvent,
			subdir:    "special_events",
		},
		{
			name:      "Challenge",
			eventType: EventTypeChallenge,
			subdir:    "challenges",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventDeck := &EventDeck{
				EventID:   "test_" + tt.name,
				PlayerTag: "#TEST123",
				EventName: tt.name,
				EventType: tt.eventType,
				StartTime: time.Now(),
				Deck: Deck{
					Cards:     []CardInDeck{{Name: "Test"}},
					AvgElixir: 3.0,
				},
				Performance: EventPerformance{},
				Battles:     []BattleRecord{},
			}

			err := manager.SaveEventDeck(eventDeck)
			if err != nil {
				t.Fatalf("SaveEventDeck failed: %v", err)
			}

			// Verify file was created in correct subdirectory
			playerDir := manager.getPlayerEventDir(eventDeck.PlayerTag)
			subdir := filepath.Join(playerDir, tt.subdir)
			files, err := filepath.Glob(filepath.Join(subdir, "*.json"))
			if err != nil {
				t.Fatalf("Failed to list files: %v", err)
			}

			found := false
			for _, file := range files {
				if filepath.Base(file) != "collection.json" {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("No event deck file found in %s", tt.subdir)
			}
		})
	}
}

func TestManager_GetEventDecks(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create test event decks with different properties
	baseTime := time.Now()
	decks := []*EventDeck{
		{
			EventID:   "deck_1",
			PlayerTag: "#TEST123",
			EventName: "Grand Challenge 1",
			EventType: EventTypeGrandChallenge,
			StartTime: baseTime,
			Deck: Deck{
				Cards:     []CardInDeck{{Name: "Knight"}},
				AvgElixir: 3.0,
			},
			Performance: EventPerformance{},
			Battles:     []BattleRecord{},
		},
		{
			EventID:   "deck_2",
			PlayerTag: "#TEST123",
			EventName: "Classic Challenge 1",
			EventType: EventTypeClassicChallenge,
			StartTime: baseTime.Add(-2 * 24 * time.Hour), // 2 days ago
			Deck: Deck{
				Cards:     []CardInDeck{{Name: "Archers"}},
				AvgElixir: 3.5,
			},
			Performance: EventPerformance{},
			Battles:     []BattleRecord{},
		},
		{
			EventID:   "deck_3",
			PlayerTag: "#TEST123",
			EventName: "Tournament 1",
			EventType: EventTypeTournament,
			StartTime: baseTime.Add(-10 * 24 * time.Hour), // 10 days ago
			Deck: Deck{
				Cards:     []CardInDeck{{Name: "Giant"}},
				AvgElixir: 4.0,
			},
			Performance: EventPerformance{},
			Battles:     []BattleRecord{},
		},
	}

	// Save all decks
	for _, deck := range decks {
		if err := manager.SaveEventDeck(deck); err != nil {
			t.Fatalf("Failed to save deck: %v", err)
		}
	}

	t.Run("Get all decks", func(t *testing.T) {
		retrieved, err := manager.GetEventDecks("#TEST123", nil)
		if err != nil {
			t.Fatalf("GetEventDecks failed: %v", err)
		}

		if len(retrieved) != 3 {
			t.Errorf("GetEventDecks returned %d decks, want 3", len(retrieved))
		}

		// Verify sorting (newest first)
		if len(retrieved) >= 2 {
			if retrieved[0].StartTime.Before(retrieved[1].StartTime) {
				t.Error("Decks are not sorted by time (newest first)")
			}
		}
	})

	t.Run("Filter by event type", func(t *testing.T) {
		eventType := EventTypeTournament
		opts := &GetEventDeckOptions{
			EventType: &eventType,
		}

		retrieved, err := manager.GetEventDecks("#TEST123", opts)
		if err != nil {
			t.Fatalf("GetEventDecks failed: %v", err)
		}

		if len(retrieved) != 1 {
			t.Errorf("GetEventDecks returned %d decks, want 1", len(retrieved))
		}
		if len(retrieved) > 0 && retrieved[0].EventType != EventTypeTournament {
			t.Error("Returned deck is not a tournament")
		}
	})

	t.Run("Filter by days back", func(t *testing.T) {
		daysBack := 5
		opts := &GetEventDeckOptions{
			DaysBack: &daysBack,
		}

		retrieved, err := manager.GetEventDecks("#TEST123", opts)
		if err != nil {
			t.Fatalf("GetEventDecks failed: %v", err)
		}

		// Should get deck_1 and deck_2 (not deck_3 which is 10 days old)
		if len(retrieved) != 2 {
			t.Errorf("GetEventDecks returned %d decks, want 2", len(retrieved))
		}
	})

	t.Run("Apply limit", func(t *testing.T) {
		limit := 2
		opts := &GetEventDeckOptions{
			Limit: &limit,
		}

		retrieved, err := manager.GetEventDecks("#TEST123", opts)
		if err != nil {
			t.Fatalf("GetEventDecks failed: %v", err)
		}

		if len(retrieved) != 2 {
			t.Errorf("GetEventDecks returned %d decks, want 2 (limit applied)", len(retrieved))
		}
	})
}

func TestManager_ImportFromBattleLogs(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create mock battle logs
	baseTime := time.Now()
	battleLogs := []clashroyale.Battle{
		{
			UTCDate:            baseTime,
			GameMode:           clashroyale.GameMode{Name: "Grand Challenge"},
			IsLadderTournament: false,
			Team: []clashroyale.BattleTeam{
				{
					Crowns: 3,
					Cards: []clashroyale.Card{
						{Name: "Knight", Level: 11, ElixirCost: 3, MaxLevel: 14},
						{Name: "Archers", Level: 10, ElixirCost: 3, MaxLevel: 14},
						{Name: "Fireball", Level: 9, ElixirCost: 4, MaxLevel: 14},
						{Name: "Zap", Level: 11, ElixirCost: 2, MaxLevel: 14},
						{Name: "Giant", Level: 8, ElixirCost: 5, MaxLevel: 14},
						{Name: "Musketeer", Level: 9, ElixirCost: 4, MaxLevel: 14},
						{Name: "Valkyrie", Level: 10, ElixirCost: 4, MaxLevel: 14},
						{Name: "Mini P.E.K.K.A", Level: 9, ElixirCost: 4, MaxLevel: 14},
					},
				},
			},
			Opponent: []clashroyale.BattleTeam{
				{Crowns: 1},
			},
		},
		{
			UTCDate:            baseTime.Add(5 * time.Minute),
			GameMode:           clashroyale.GameMode{Name: "Grand Challenge"},
			IsLadderTournament: false,
			Team: []clashroyale.BattleTeam{
				{
					Crowns: 2,
					Cards: []clashroyale.Card{
						{Name: "Knight", Level: 11, ElixirCost: 3, MaxLevel: 14},
						{Name: "Archers", Level: 10, ElixirCost: 3, MaxLevel: 14},
						{Name: "Fireball", Level: 9, ElixirCost: 4, MaxLevel: 14},
						{Name: "Zap", Level: 11, ElixirCost: 2, MaxLevel: 14},
						{Name: "Giant", Level: 8, ElixirCost: 5, MaxLevel: 14},
						{Name: "Musketeer", Level: 9, ElixirCost: 4, MaxLevel: 14},
						{Name: "Valkyrie", Level: 10, ElixirCost: 4, MaxLevel: 14},
						{Name: "Mini P.E.K.K.A", Level: 9, ElixirCost: 4, MaxLevel: 14},
					},
				},
			},
			Opponent: []clashroyale.BattleTeam{
				{Crowns: 3},
			},
		},
	}

	imported, err := manager.ImportFromBattleLogs(battleLogs, "#TEST123")
	if err != nil {
		t.Fatalf("ImportFromBattleLogs failed: %v", err)
	}

	if len(imported) == 0 {
		t.Error("No decks were imported")
	}

	// Verify decks were saved
	retrieved, err := manager.GetEventDecks("#TEST123", nil)
	if err != nil {
		t.Fatalf("GetEventDecks failed: %v", err)
	}

	if len(retrieved) == 0 {
		t.Error("No decks found after import")
	}
}

func TestManager_GetCollection(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	t.Run("Empty collection", func(t *testing.T) {
		collection, err := manager.GetCollection("#TEST123")
		if err != nil {
			t.Fatalf("GetCollection failed: %v", err)
		}

		if collection == nil {
			t.Fatal("GetCollection returned nil")
		}
		if len(collection.Decks) != 0 {
			t.Errorf("Empty collection should have 0 decks, got %d", len(collection.Decks))
		}
	})

	t.Run("Collection with decks", func(t *testing.T) {
		// Save a deck first
		eventDeck := &EventDeck{
			EventID:     "test_collection",
			PlayerTag:   "#TEST123",
			EventName:   "Test Event",
			EventType:   EventTypeChallenge,
			StartTime:   time.Now(),
			Deck:        Deck{Cards: []CardInDeck{{Name: "Test"}}, AvgElixir: 3.0},
			Performance: EventPerformance{},
			Battles:     []BattleRecord{},
		}

		if err := manager.SaveEventDeck(eventDeck); err != nil {
			t.Fatalf("SaveEventDeck failed: %v", err)
		}

		// Get collection
		collection, err := manager.GetCollection("#TEST123")
		if err != nil {
			t.Fatalf("GetCollection failed: %v", err)
		}

		if len(collection.Decks) != 1 {
			t.Errorf("Collection should have 1 deck, got %d", len(collection.Decks))
		}
		if collection.PlayerTag != "#TEST123" {
			t.Errorf("PlayerTag = %s, want #TEST123", collection.PlayerTag)
		}
	})
}

func TestManager_DeleteEventDeck(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Create and save an event deck
	eventDeck := &EventDeck{
		EventID:     "delete_test",
		PlayerTag:   "#TEST123",
		EventName:   "Delete Test",
		EventType:   EventTypeChallenge,
		StartTime:   time.Now(),
		Deck:        Deck{Cards: []CardInDeck{{Name: "Test"}}, AvgElixir: 3.0},
		Performance: EventPerformance{},
		Battles:     []BattleRecord{},
	}

	if err := manager.SaveEventDeck(eventDeck); err != nil {
		t.Fatalf("SaveEventDeck failed: %v", err)
	}

	// Verify deck exists
	decks, err := manager.GetEventDecks("#TEST123", nil)
	if err != nil {
		t.Fatalf("GetEventDecks failed: %v", err)
	}
	if len(decks) == 0 {
		t.Fatal("Deck was not saved")
	}

	// Delete the deck
	err = manager.DeleteEventDeck("#TEST123", "delete_test")
	if err != nil {
		t.Fatalf("DeleteEventDeck failed: %v", err)
	}

	// Verify deck was deleted
	decks, err = manager.GetEventDecks("#TEST123", nil)
	if err != nil {
		t.Fatalf("GetEventDecks failed: %v", err)
	}
	if len(decks) != 0 {
		t.Errorf("Deck was not deleted, still found %d decks", len(decks))
	}

	// Verify collection was updated
	collection, err := manager.GetCollection("#TEST123")
	if err != nil {
		t.Fatalf("GetCollection failed: %v", err)
	}
	if len(collection.Decks) != 0 {
		t.Errorf("Collection still has %d decks after deletion", len(collection.Decks))
	}
}

func TestManager_DeleteEventDeck_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Try to delete non-existent deck
	err := manager.DeleteEventDeck("#TEST123", "nonexistent")
	if err == nil {
		t.Error("DeleteEventDeck should return error for non-existent deck")
	}
}
