package events

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/storage"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// Manager manages event deck storage, retrieval, and analysis
type Manager struct {
	dataDir       string
	eventDecksDir string
	parser        *Parser
}

// NewManager creates a new event deck manager
func NewManager(dataDir string) *Manager {
	eventDecksDir := filepath.Join(dataDir, "event_decks")

	return &Manager{
		dataDir:       dataDir,
		eventDecksDir: eventDecksDir,
		parser:        NewParser(),
	}
}

// getPlayerEventDir returns the directory for a player's event decks
func (m *Manager) getPlayerEventDir(playerTag string) string {
	// Remove # from tag for directory name
	cleanTag := strings.TrimPrefix(playerTag, "#")
	return filepath.Join(m.eventDecksDir, cleanTag)
}

// ensurePlayerDirectories creates all necessary subdirectories for a player
func (m *Manager) ensurePlayerDirectories(playerTag string) error {
	playerDir := m.getPlayerEventDir(playerTag)

	// Create subdirectories
	dirs := []string{
		playerDir,
		filepath.Join(playerDir, "challenges"),
		filepath.Join(playerDir, "tournaments"),
		filepath.Join(playerDir, "special_events"),
		filepath.Join(playerDir, "aggregated"),
	}

	for _, dir := range dirs {
		if err := storage.EnsureDirectory(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// SaveEventDeck saves an event deck to the file system
func (m *Manager) SaveEventDeck(eventDeck *EventDeck) error {
	if err := m.ensurePlayerDirectories(eventDeck.PlayerTag); err != nil {
		return err
	}

	playerDir := m.getPlayerEventDir(eventDeck.PlayerTag)

	// Determine subdirectory based on event type
	var subdir string
	switch eventDeck.EventType {
	case EventTypeTournament:
		subdir = filepath.Join(playerDir, "tournaments")
	case EventTypeSpecialEvent:
		subdir = filepath.Join(playerDir, "special_events")
	default:
		subdir = filepath.Join(playerDir, "challenges")
	}

	// Generate filename
	timestamp := eventDeck.StartTime.Format("2006-01-02")
	eventName := strings.ToLower(eventDeck.EventName)
	eventName = strings.ReplaceAll(eventName, " ", "_")
	eventName = strings.ReplaceAll(eventName, "/", "_")
	filename := fmt.Sprintf("%s_%s.json", timestamp, eventName)
	filePath := filepath.Join(subdir, filename)

	if err := storage.WriteJSON(filePath, eventDeck); err != nil {
		return fmt.Errorf("failed to write event deck file: %w", err)
	}

	// Update collection file
	if err := m.updateCollectionFile(eventDeck); err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	return nil
}

// updateCollectionFile updates the player's event deck collection file
func (m *Manager) updateCollectionFile(eventDeck *EventDeck) error {
	playerDir := m.getPlayerEventDir(eventDeck.PlayerTag)
	collectionFile := filepath.Join(playerDir, "collection.json")

	// Load existing collection
	var collection EventDeckCollection
	if data, err := os.ReadFile(collectionFile); err == nil {
		// File exists, unmarshal it
		if err := json.Unmarshal(data, &collection); err != nil {
			// If unmarshal fails, create new collection
			collection = EventDeckCollection{
				PlayerTag:   eventDeck.PlayerTag,
				LastUpdated: time.Now(),
			}
		}
	} else {
		// File doesn't exist, create new collection
		collection = EventDeckCollection{
			PlayerTag:   eventDeck.PlayerTag,
			LastUpdated: time.Now(),
		}
	}

	// Add the deck
	collection.AddDeck(*eventDeck)

	if err := storage.WriteJSON(collectionFile, collection); err != nil {
		return fmt.Errorf("failed to write collection file: %w", err)
	}

	return nil
}

// GetEventDeckOptions configures event deck retrieval
type GetEventDeckOptions struct {
	EventType *EventType // Filter by event type
	DaysBack  *int       // Only get decks from last N days
	Limit     *int       // Maximum number of decks to return
}

// GetEventDecks retrieves event decks for a player
func (m *Manager) GetEventDecks(playerTag string, opts *GetEventDeckOptions) ([]EventDeck, error) {
	if opts == nil {
		opts = &GetEventDeckOptions{}
	}

	playerDir := m.getPlayerEventDir(playerTag)
	decks := make([]EventDeck, 0)

	// Determine which subdirectories to search
	var subdirs []string
	if opts.EventType != nil {
		switch *opts.EventType {
		case EventTypeTournament:
			subdirs = []string{filepath.Join(playerDir, "tournaments")}
		case EventTypeSpecialEvent:
			subdirs = []string{filepath.Join(playerDir, "special_events")}
		default:
			subdirs = []string{filepath.Join(playerDir, "challenges")}
		}
	} else {
		subdirs = []string{
			filepath.Join(playerDir, "challenges"),
			filepath.Join(playerDir, "tournaments"),
			filepath.Join(playerDir, "special_events"),
		}
	}

	// Calculate cutoff time if days_back is specified
	var cutoff time.Time
	if opts.DaysBack != nil {
		cutoff = time.Now().AddDate(0, 0, -*opts.DaysBack)
	}

	// Load decks from subdirectories
	for _, subdir := range subdirs {
		if _, err := os.Stat(subdir); os.IsNotExist(err) {
			continue
		}

		files, err := filepath.Glob(filepath.Join(subdir, "*.json"))
		if err != nil {
			continue
		}

		for _, filePath := range files {
			// Skip collection file
			if filepath.Base(filePath) == "collection.json" {
				continue
			}

			// Read file
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			// Unmarshal deck
			var deck EventDeck
			if err := json.Unmarshal(data, &deck); err != nil {
				continue
			}

			// Apply filters
			if opts.DaysBack != nil && deck.StartTime.Before(cutoff) {
				continue
			}

			decks = append(decks, deck)
		}
	}

	// Sort by start time (newest first)
	sort.Slice(decks, func(i, j int) bool {
		return decks[i].StartTime.After(decks[j].StartTime)
	})

	// Apply limit
	if opts.Limit != nil && len(decks) > *opts.Limit {
		decks = decks[:*opts.Limit]
	}

	return decks, nil
}

// ImportFromBattleLogs imports event decks from battle logs
func (m *Manager) ImportFromBattleLogs(battleLogs []clashroyale.Battle, playerTag string) ([]EventDeck, error) {
	// Parse battle logs
	eventDecks, err := m.parser.ParseBattleLogs(battleLogs, playerTag)
	if err != nil {
		return nil, fmt.Errorf("failed to parse battle logs: %w", err)
	}

	// Save each event deck
	imported := make([]EventDeck, 0, len(eventDecks))
	for _, deck := range eventDecks {
		if err := m.SaveEventDeck(&deck); err != nil {
			// Log error but continue with other decks
			continue
		}
		imported = append(imported, deck)
	}

	return imported, nil
}

// GetCollection loads the player's event deck collection
func (m *Manager) GetCollection(playerTag string) (*EventDeckCollection, error) {
	playerDir := m.getPlayerEventDir(playerTag)
	collectionFile := filepath.Join(playerDir, "collection.json")

	data, err := os.ReadFile(collectionFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty collection if file doesn't exist
			return &EventDeckCollection{
				PlayerTag:   playerTag,
				Decks:       []EventDeck{},
				LastUpdated: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to read collection file: %w", err)
	}

	var collection EventDeckCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("failed to unmarshal collection: %w", err)
	}

	return &collection, nil
}

// Helper functions for Manager methods
func getSubdirectoryForEventType(eventType EventType, playerDir string) string {
	var subdirName string
	switch eventType {
	case EventTypeTournament:
		subdirName = "tournaments"
	case EventTypeSpecialEvent:
		subdirName = "special_events"
	default:
		subdirName = "challenges"
	}
	return filepath.Join(playerDir, subdirName)
}

func findDeckInCollection(collection *EventDeckCollection, eventID string) *EventDeck {
	for i := range collection.Decks {
		if collection.Decks[i].EventID == eventID {
			return &collection.Decks[i]
		}
	}
	return nil
}

func findDeckFileInDirectory(subdir, eventID string) (string, error) {
	files, err := filepath.Glob(filepath.Join(subdir, "*.json"))
	if err != nil {
		return "", fmt.Errorf("failed to list files: %w", err)
	}

	for _, filePath := range files {
		if filepath.Base(filePath) == "collection.json" {
			continue
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var deck EventDeck
		if err := json.Unmarshal(data, &deck); err != nil {
			continue
		}

		if deck.EventID == eventID {
			return filePath, nil
		}
	}

	return "", fmt.Errorf("event deck file not found: %s", eventID)
}

func removeEventDeckFromCollection(collection *EventDeckCollection, eventID string) {
	newDecks := make([]EventDeck, 0, len(collection.Decks)-1)
	for i := range collection.Decks {
		if collection.Decks[i].EventID != eventID {
			newDecks = append(newDecks, collection.Decks[i])
		}
	}
	collection.Decks = newDecks
	collection.LastUpdated = time.Now()
}

func persistUpdatedCollection(playerDir string, collection *EventDeckCollection) error {
	collectionFile := filepath.Join(playerDir, "collection.json")
	collectionData, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal collection: %w", err)
	}

	if err := os.WriteFile(collectionFile, collectionData, 0o644); err != nil {
		return fmt.Errorf("failed to write collection file: %w", err)
	}

	return nil
}

// DeleteEventDeck deletes an event deck by ID
func (m *Manager) DeleteEventDeck(playerTag, eventID string) error {
	collection, err := m.GetCollection(playerTag)
	if err != nil {
		return err
	}

	targetDeck := findDeckInCollection(collection, eventID)
	if targetDeck == nil {
		return fmt.Errorf("event deck not found: %s", eventID)
	}

	playerDir := m.getPlayerEventDir(playerTag)
	subdir := getSubdirectoryForEventType(targetDeck.EventType, playerDir)

	filePath, err := findDeckFileInDirectory(subdir, eventID)
	if err != nil {
		return err
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	removeEventDeckFromCollection(collection, eventID)

	if err := persistUpdatedCollection(playerDir, collection); err != nil {
		return err
	}

	return nil
}

// GetEventDeck retrieves a specific event deck by ID for a player
func (m *Manager) GetEventDeck(eventID, playerTag string) (*EventDeck, error) {
	// Get collection to find the deck
	collection, err := m.GetCollection(playerTag)
	if err != nil {
		return nil, err
	}

	// Find the deck in collection
	for _, deck := range collection.Decks {
		if deck.EventID == eventID {
			// Return a copy to avoid modifying the collection
			deckCopy := deck
			return &deckCopy, nil
		}
	}

	return nil, fmt.Errorf("event deck not found: %s", eventID)
}

// GetPlayerEventDecks retrieves all event decks for a player
func (m *Manager) GetPlayerEventDecks(playerTag string) (*EventDeckCollection, error) {
	return m.GetCollection(playerTag)
}
