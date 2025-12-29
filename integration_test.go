//go:build integration

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/events"
)

// TestAnalysisToDeckBuilderIntegration tests the full flow from card analysis to deck building
func TestAnalysisToDeckBuilderIntegration(t *testing.T) {
	// Create temporary directory for test data
	tempDir := t.TempDir()

	// Create test card data that matches the expected format
	cardLevels := make(map[string]deck.CardLevelData)

	// Hog Rider - win condition
	cardLevels["Hog Rider"] = deck.CardLevelData{
		Level:    8,
		MaxLevel: 13,
		Rarity:   "Rare",
	}

	// Fireball - big spell
	cardLevels["Fireball"] = deck.CardLevelData{
		Level:    7,
		MaxLevel: 11,
		Rarity:   "Rare",
	}

	// Zap - small spell
	cardLevels["Zap"] = deck.CardLevelData{
		Level:    11,
		MaxLevel: 13,
		Rarity:   "Common",
	}

	// Cannon - building
	cardLevels["Cannon"] = deck.CardLevelData{
		Level:    11,
		MaxLevel: 13,
		Rarity:   "Common",
	}

	// Archers - support
	cardLevels["Archers"] = deck.CardLevelData{
		Level:    10,
		MaxLevel: 13,
		Rarity:   "Common",
	}

	// Knight - support/cycle
	cardLevels["Knight"] = deck.CardLevelData{
		Level:    11,
		MaxLevel: 13,
		Rarity:   "Common",
	}

	// Skeletons - cycle
	cardLevels["Skeletons"] = deck.CardLevelData{
		Level:    11,
		MaxLevel: 13,
		Rarity:   "Common",
	}

	// Ice Spirit - cycle
	cardLevels["Ice Spirit"] = deck.CardLevelData{
		Level:    11,
		MaxLevel: 13,
		Rarity:   "Common",
	}

	// Create card analysis in the format expected by deck builder
	analysisData := deck.CardAnalysis{
		CardLevels:   cardLevels,
		AnalysisTime: time.Now().Format("2006-01-02T15:04:05Z"),
	}

	// Save the analysis to JSON file manually
	analysisPath := filepath.Join(tempDir, "test_analysis.json")
	data, err := json.MarshalIndent(analysisData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal analysis: %v", err)
	}

	err = os.WriteFile(analysisPath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to save analysis: %v", err)
	}

	// Create deck builder
	builder := deck.NewBuilder(tempDir)

	// Load the analysis from file
	loadedAnalysis, err := builder.LoadAnalysis(analysisPath)
	if err != nil {
		t.Fatalf("Failed to load analysis: %v", err)
	}

	// Build a deck from the analysis
	recommendation, err := builder.BuildDeckFromAnalysis(*loadedAnalysis)
	if err != nil {
		t.Fatalf("Failed to build deck: %v", err)
	}

	// Verify the recommendation
	if len(recommendation.Deck) != 8 {
		t.Errorf("Expected 8 cards in deck, got %d", len(recommendation.Deck))
	}

	if len(recommendation.DeckDetail) != 8 {
		t.Errorf("Expected 8 card details, got %d", len(recommendation.DeckDetail))
	}

	// Check that we have a win condition
	hasWinCondition := false
	for _, card := range recommendation.DeckDetail {
		if card.Role == string(deck.RoleWinCondition) {
			hasWinCondition = true
			break
		}
	}
	if !hasWinCondition {
		t.Log("Note: No win condition found in available cards")
	}

	// Check average elixir is reasonable
	if recommendation.AvgElixir < 2.0 || recommendation.AvgElixir > 5.0 {
		t.Errorf("Average elixir %.2f seems unreasonable", recommendation.AvgElixir)
	}

	// Save the deck recommendation
	deckPath, err := builder.SaveDeck(recommendation, filepath.Join(tempDir, "decks"), "#INTEGRATION_TEST")
	if err != nil {
		t.Fatalf("Failed to save deck: %v", err)
	}

	// Verify deck file exists and is valid JSON
	data, err = os.ReadFile(deckPath)
	if err != nil {
		t.Fatalf("Failed to read deck file: %v", err)
	}

	var savedDeck deck.DeckRecommendation
	if err := json.Unmarshal(data, &savedDeck); err != nil {
		t.Fatalf("Failed to unmarshal deck JSON: %v", err)
	}

	// Compare saved deck with original
	if len(savedDeck.Deck) != len(recommendation.Deck) {
		t.Error("Saved deck has different number of cards than original")
	}
}

// TestEventTrackingIntegration tests the full event tracking flow
func TestEventTrackingIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// Create event manager
	manager := events.NewManager(tempDir)

	// Create test event deck
	eventDeck := &events.EventDeck{
		EventID:   "test_integration_event_123",
		PlayerTag: "#INTEGRATION_TEST",
		EventName: "Test Challenge",
		EventType: events.EventTypeChallenge,
		StartTime: time.Now().Add(-2 * time.Hour),
		Deck: events.Deck{
			Cards: []events.CardInDeck{
				{
					Name:       "Hog Rider",
					ID:         26000000,
					Level:      8,
					MaxLevel:   13,
					Rarity:     "Rare",
					ElixirCost: 4,
				},
				{
					Name:       "Fireball",
					ID:         28000004,
					Level:      7,
					MaxLevel:   11,
					Rarity:     "Rare",
					ElixirCost: 4,
				},
				// Add more cards to make 8 total
				{
					Name:       "Zap",
					ID:         28000000,
					Level:      11,
					MaxLevel:   13,
					Rarity:     "Common",
					ElixirCost: 2,
				},
				{
					Name:       "Cannon",
					ID:         23000000,
					Level:      11,
					MaxLevel:   13,
					Rarity:     "Common",
					ElixirCost: 3,
				},
				{
					Name:       "Archers",
					ID:         27000002,
					Level:      10,
					MaxLevel:   13,
					Rarity:     "Common",
					ElixirCost: 3,
				},
				{
					Name:       "Knight",
					ID:         26000001,
					Level:      11,
					MaxLevel:   13,
					Rarity:     "Common",
					ElixirCost: 3,
				},
				{
					Name:       "Skeletons",
					ID:         27000000,
					Level:      11,
					MaxLevel:   13,
					Rarity:     "Common",
					ElixirCost: 1,
				},
				{
					Name:       "Ice Spirit",
					ID:         27000001,
					Level:      11,
					MaxLevel:   13,
					Rarity:     "Common",
					ElixirCost: 1,
				},
			},
		},
		Performance: events.EventPerformance{
			Wins:          9,
			Losses:        3,
			MaxWins:       &[]int{12}[0],
			CurrentStreak: 0,
			BestStreak:    4,
		},
	}

	// Calculate initial metrics
	eventDeck.Deck.CalculateAvgElixir()
	eventDeck.Performance.CalculateWinRate()
	eventDeck.Performance.UpdateProgress()

	// Add some battle records
	eventDeck.AddBattle(events.BattleRecord{
		Timestamp:      time.Now().Add(-90 * time.Minute),
		OpponentTag:    "#OPPONENT1",
		OpponentName:   "TestPlayer1",
		Result:         "win",
		Crowns:         3,
		OpponentCrowns: 1,
	})

	eventDeck.AddBattle(events.BattleRecord{
		Timestamp:      time.Now().Add(-60 * time.Minute),
		OpponentTag:    "#OPPONENT2",
		OpponentName:   "TestPlayer2",
		Result:         "loss",
		Crowns:         1,
		OpponentCrowns: 3,
	})

	// Store the event deck
	err := manager.SaveEventDeck(eventDeck)
	if err != nil {
		t.Fatalf("Failed to store event deck: %v", err)
	}

	// Retrieve the player's event deck collection
	collection, err := manager.GetPlayerEventDecks("#INTEGRATION_TEST")
	if err != nil {
		t.Fatalf("Failed to get player event decks: %v", err)
	}

	// Verify collection
	if collection.PlayerTag != "#INTEGRATION_TEST" {
		t.Errorf("Expected player tag #INTEGRATION_TEST, got %s", collection.PlayerTag)
	}

	if len(collection.Decks) != 1 {
		t.Fatalf("Expected 1 deck in collection, got %d", len(collection.Decks))
	}

	// Retrieve the specific event deck
	retrievedDeck, err := manager.GetEventDeck(eventDeck.EventID, "#INTEGRATION_TEST")
	if err != nil {
		t.Fatalf("Failed to get specific event deck: %v", err)
	}

	// Verify retrieved deck matches original
	if retrievedDeck.EventID != eventDeck.EventID {
		t.Error("Retrieved deck has different event ID")
	}

	if len(retrievedDeck.Battles) != 2 {
		t.Errorf("Expected 2 battles, got %d", len(retrievedDeck.Battles))
	}

	// Test collection queries
	recentDecks := collection.GetRecentDecks(7) // Last 7 days
	if len(recentDecks) != 1 {
		t.Error("Should have 1 recent deck")
	}

	challengeDecks := collection.GetDecksByType(events.EventTypeChallenge)
	if len(challengeDecks) != 1 {
		t.Error("Should have 1 challenge deck")
	}
}

// TestFullWorkflowIntegration tests the complete workflow from raw data to exported CSV
func TestFullWorkflowIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()

	// Step 1: Create and save card analysis
	cardLevels := make(map[string]deck.CardLevelData)
	cardNames := []string{
		"Hog Rider", "Fireball", "Zap", "Cannon",
		"Archers", "Knight", "Skeletons", "Ice Spirit",
		"Musketeer", "Valkyrie",
	}

	for i, name := range cardNames {
		cardLevels[name] = deck.CardLevelData{
			Level:    8 + i%4,
			MaxLevel: 13,
			Rarity:   []string{"Common", "Rare", "Epic"}[i%3],
		}
	}

	// Create card analysis in the format expected by deck builder
	analysisData := deck.CardAnalysis{
		CardLevels:   cardLevels,
		AnalysisTime: time.Now().Format("2006-01-02T15:04:05Z"),
	}

	// Save the analysis to JSON file manually
	analysisPath := filepath.Join(tempDir, "workflow_analysis.json")
	data, err := json.MarshalIndent(analysisData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal analysis: %v", err)
	}

	err = os.WriteFile(analysisPath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to save analysis: %v", err)
	}

	// Step 2: Build a deck from the analysis
	builder := deck.NewBuilder(tempDir)
	loadedAnalysis, err := builder.LoadAnalysis(analysisPath)
	if err != nil {
		t.Fatalf("Failed to load analysis: %v", err)
	}

	recommendation, err := builder.BuildDeckFromAnalysis(*loadedAnalysis)
	if err != nil {
		t.Fatalf("Failed to build deck: %v", err)
	}

	deckPath, err := builder.SaveDeck(recommendation, filepath.Join(tempDir, "decks"), "#WORKFLOW_TEST")
	if err != nil {
		t.Fatalf("Failed to save deck: %v", err)
	}

	// Step 3: Create and store event data
	manager := events.NewManager(tempDir)
	eventDeck := &events.EventDeck{
		EventID:   "workflow_test_event",
		PlayerTag: "#WORKFLOW_TEST",
		EventName: "Integration Test Challenge",
		EventType: events.EventTypeChallenge,
		StartTime: time.Now().Add(-time.Hour),
		Deck:      events.Deck{Cards: make([]events.CardInDeck, 8)},
	}

	// Copy cards from recommendation to event deck
	for i, cardName := range recommendation.Deck {
		if i < 8 {
			eventDeck.Deck.Cards[i] = events.CardInDeck{
				Name:       cardName,
				Level:      recommendation.DeckDetail[i].Level,
				MaxLevel:   recommendation.DeckDetail[i].MaxLevel,
				Rarity:     recommendation.DeckDetail[i].Rarity,
				ElixirCost: recommendation.DeckDetail[i].Elixir,
			}
		}
	}

	err = manager.SaveEventDeck(eventDeck)
	if err != nil {
		t.Fatalf("Failed to store event deck: %v", err)
	}

	// Step 4: Verify all files were created correctly
	files := []string{analysisPath, deckPath}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", file)
		}
	}

	// Check that event deck file was created in correct location
	// Since getPlayerEventDir is private, we construct the path ourselves
	playerDir := filepath.Join(tempDir, "event_decks", strings.TrimPrefix("#WORKFLOW_TEST", "#"))
	eventFiles, err := filepath.Glob(filepath.Join(playerDir, "*.json"))
	if err != nil || len(eventFiles) == 0 {
		t.Error("No event deck files found in player directory")
	}

	// Step 5: Load and verify data integrity
	// Load and verify analysis
	savedAnalysis, err := builder.LoadAnalysis(analysisPath)
	if err != nil {
		t.Fatalf("Failed to reload analysis: %v", err)
	}

	if len(savedAnalysis.CardLevels) != len(cardLevels) {
		t.Error("Analysis card count mismatch after reload")
	}

	// Load and verify deck
	savedDeck, err := builder.LoadDeckFromFile(deckPath)
	if err != nil {
		t.Fatalf("Failed to reload deck: %v", err)
	}

	if savedDeck.AvgElixir != recommendation.AvgElixir {
		t.Error("Deck average elixir mismatch after reload")
	}

	// Load and verify event deck
	savedEventDeck, err := manager.GetEventDeck(eventDeck.EventID, "#WORKFLOW_TEST")
	if err != nil {
		t.Fatalf("Failed to reload event deck: %v", err)
	}

	if savedEventDeck.EventName != eventDeck.EventName {
		t.Error("Event deck name mismatch after reload")
	}
}

// TestErrorHandlingIntegration tests error handling across components
func TestErrorHandlingIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// Test 1: Invalid analysis data
	builder := deck.NewBuilder(tempDir)
	invalidAnalysis := deck.CardAnalysis{
		CardLevels: map[string]deck.CardLevelData{}, // Empty card levels
	}

	_, err := builder.BuildDeckFromAnalysis(invalidAnalysis)
	if err == nil {
		t.Error("Expected error when building deck with empty analysis")
	}

	// Test 2: Non-existent analysis file
	_, err = builder.LoadAnalysis("non_existent_file.json")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}

	// Test 3: Invalid event deck
	manager := events.NewManager(tempDir)
	invalidEventDeck := &events.EventDeck{
		EventID:   "invalid_test",
		PlayerTag: "#TEST",
		Deck:      events.Deck{Cards: make([]events.CardInDeck, 7)}, // Invalid deck size
	}

	err = invalidEventDeck.Deck.Validate()
	if err == nil {
		t.Error("Expected error when validating deck with wrong size")
	}

	// Test 4: Non-existent event deck
	_, err = manager.GetEventDeck("non_existent_event", "#TEST")
	if err == nil {
		t.Error("Expected error when getting non-existent event deck")
	}

	// Test 5: Invalid card elixir costs
	invalidCards := []events.CardInDeck{
		{Name: "Test Card", ElixirCost: 11}, // Invalid elixir cost
	}
	invalidDeck2 := &events.Deck{Cards: invalidCards}

	err = invalidDeck2.Validate()
	if err == nil {
		t.Error("Expected error when validating deck with invalid elixir cost")
	}
}

// BenchmarkFullWorkflow benchmarks the complete workflow performance
func BenchmarkFullWorkflow(b *testing.B) {
	tempDir := b.TempDir()

	// Prepare test data once
	cardData := make(map[string]deck.CardLevelData)
	cardNames := []string{
		"Hog Rider", "Fireball", "Zap", "Cannon",
		"Archers", "Knight", "Skeletons", "Ice Spirit",
		"Musketeer", "Valkyrie",
	}

	for i, name := range cardNames {
		cardData[name] = deck.CardLevelData{
			Level:    8 + i%4,
			MaxLevel: 13,
			Rarity:   []string{"Common", "Rare", "Epic"}[i%3],
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create analysis
		analysisData := deck.CardAnalysis{
			CardLevels:   cardData,
			AnalysisTime: time.Now().Format("2006-01-02T15:04:05Z"),
		}

		// Save to file
		analysisPath := filepath.Join(tempDir, fmt.Sprintf("bench_analysis_%d.json", i))
		data, err := json.MarshalIndent(analysisData, "", "  ")
		if err != nil {
			b.Fatal(err)
		}

		err = os.WriteFile(analysisPath, data, 0644)
		if err != nil {
			b.Fatal(err)
		}

		// Build deck
		builder := deck.NewBuilder(tempDir)
		loadedAnalysis, err := builder.LoadAnalysis(analysisPath)
		if err != nil {
			b.Fatal(err)
		}

		_, err = builder.BuildDeckFromAnalysis(*loadedAnalysis)
		if err != nil {
			b.Fatal(err)
		}

		// Clean up file to avoid accumulation during benchmark
		os.Remove(analysisPath)
	}
}
