//go:build integration

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/deck"
)

// createTestCheckpoint creates a test checkpoint file
func createTestCheckpoint(t *testing.T, playerTag string, stats deck.DiscoveryStats) string {
	t.Helper()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	sanitizedTag := playerTag
	if sanitizedTag[0] == '#' {
		sanitizedTag = sanitizedTag[1:]
	}
	checkpointDir := filepath.Join(homeDir, ".cr-api", "discover")
	if err := os.MkdirAll(checkpointDir, 0o755); err != nil {
		t.Fatalf("failed to create checkpoint dir: %v", err)
	}

	checkpoint := deck.DiscoveryCheckpoint{
		GeneratorCheckpoint: &deck.GeneratorCheckpoint{
			Strategy:  deck.StrategySmartSample,
			Position:  100,
			Generated: 100,
		},
		Stats:     stats,
		Timestamp: time.Now(),
		PlayerTag: playerTag,
		Strategy:  deck.StrategySmartSample,
	}

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal checkpoint: %v", err)
	}

	checkpointPath := filepath.Join(checkpointDir, sanitizedTag+".json")
	if err := os.WriteFile(checkpointPath, data, 0o644); err != nil {
		t.Fatalf("failed to write checkpoint: %v", err)
	}

	return checkpointPath
}

func TestDeckDiscoverCheckpointPersistence(t *testing.T) {
	tag := "PERSIST"

	// Create initial checkpoint
	initialStats := deck.DiscoveryStats{
		Evaluated: 100,
		Stored:    50,
		AvgScore:  7.5,
		BestScore: 8.0,
		StartTime: time.Now(),
		Elapsed:   5 * time.Minute,
		Strategy:  deck.StrategySmartSample,
		PlayerTag: "#" + tag,
	}

	checkpointPath := createTestCheckpoint(t, tag, initialStats)
	defer os.Remove(checkpointPath)

	// Read and verify checkpoint
	data, err := os.ReadFile(checkpointPath)
	if err != nil {
		t.Fatalf("failed to read checkpoint: %v", err)
	}

	var checkpoint deck.DiscoveryCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		t.Fatalf("failed to unmarshal checkpoint: %v", err)
	}

	// Verify checkpoint data
	if checkpoint.PlayerTag != "#"+tag {
		t.Errorf("expected player tag #%s, got %s", tag, checkpoint.PlayerTag)
	}

	if checkpoint.Stats.Evaluated != 100 {
		t.Errorf("expected 100 evaluated, got %d", checkpoint.Stats.Evaluated)
	}

	if checkpoint.Strategy != deck.StrategySmartSample {
		t.Errorf("expected strategy smart-sample, got %s", checkpoint.Strategy)
	}
}

func TestDeckDiscoverCheckpointValidation(t *testing.T) {
	tag := "VALIDATE"

	// Create comprehensive checkpoint
	stats := deck.DiscoveryStats{
		Evaluated: 500,
		Total:     1000,
		Stored:    250,
		AvgScore:  7.2,
		BestScore: 9.1,
		BestDeck:  []string{"Hog Rider", "Musketeer", "Fireball", "Zap", "Cannon", "Ice Spirit", "Skeletons", "Archers"},
		StartTime: time.Now().Add(-30 * time.Minute),
		Elapsed:   30 * time.Minute,
		Rate:      16.67,
		ETA:       30 * time.Minute,
		Strategy:  deck.StrategySmartSample,
		PlayerTag: "#" + tag,
		TopScores: []float64{9.1, 8.9, 8.7, 8.5, 8.3},
	}

	checkpointPath := createTestCheckpoint(t, tag, stats)
	defer os.Remove(checkpointPath)

	// Read and validate
	data, err := os.ReadFile(checkpointPath)
	if err != nil {
		t.Fatalf("failed to read checkpoint: %v", err)
	}

	var checkpoint deck.DiscoveryCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		t.Fatalf("failed to unmarshal checkpoint: %v", err)
	}

	// Validate all fields
	if checkpoint.Stats.Evaluated != 500 {
		t.Errorf("Expected 500 evaluated, got %d", checkpoint.Stats.Evaluated)
	}

	if checkpoint.Stats.Total != 1000 {
		t.Errorf("Expected 1000 total, got %d", checkpoint.Stats.Total)
	}

	if checkpoint.Stats.Stored != 250 {
		t.Errorf("Expected 250 stored, got %d", checkpoint.Stats.Stored)
	}

	if checkpoint.Stats.AvgScore != 7.2 {
		t.Errorf("Expected avg score 7.2, got %f", checkpoint.Stats.AvgScore)
	}

	if checkpoint.Stats.BestScore != 9.1 {
		t.Errorf("Expected best score 9.1, got %f", checkpoint.Stats.BestScore)
	}

	if len(checkpoint.Stats.BestDeck) != 8 {
		t.Errorf("Expected 8 cards in best deck, got %d", len(checkpoint.Stats.BestDeck))
	}

	if checkpoint.Stats.Rate != 16.67 {
		t.Errorf("Expected rate 16.67, got %f", checkpoint.Stats.Rate)
	}

	if len(checkpoint.Stats.TopScores) != 5 {
		t.Errorf("Expected 5 top scores, got %d", len(checkpoint.Stats.TopScores))
	}
}

func TestDeckDiscoverStrategyConstants(t *testing.T) {
	// Validate all strategy constants are properly defined
	strategies := []struct {
		strategy   deck.GeneratorStrategy
		stringRep  string
		validators []func(deck.GeneratorStrategy) bool
	}{
		{
			strategy:  deck.StrategyExhaustive,
			stringRep: "exhaustive",
			validators: []func(deck.GeneratorStrategy) bool{
				func(s deck.GeneratorStrategy) bool { return string(s) == "exhaustive" },
			},
		},
		{
			strategy:  deck.StrategySmartSample,
			stringRep: "smart-sample",
			validators: []func(deck.GeneratorStrategy) bool{
				func(s deck.GeneratorStrategy) bool { return string(s) == "smart-sample" },
			},
		},
		{
			strategy:  deck.StrategyRandomSample,
			stringRep: "random-sample",
			validators: []func(deck.GeneratorStrategy) bool{
				func(s deck.GeneratorStrategy) bool { return string(s) == "random-sample" },
			},
		},
		{
			strategy:  deck.StrategyArchetypeFocused,
			stringRep: "archetype-focused",
			validators: []func(deck.GeneratorStrategy) bool{
				func(s deck.GeneratorStrategy) bool { return string(s) == "archetype-focused" },
			},
		},
	}

	for _, tc := range strategies {
		t.Run(tc.stringRep, func(t *testing.T) {
			// Verify string representation
			if string(tc.strategy) != tc.stringRep {
				t.Errorf("Strategy string mismatch: got %s, want %s", string(tc.strategy), tc.stringRep)
			}

			// Run validators
			for _, validator := range tc.validators {
				if !validator(tc.strategy) {
					t.Errorf("Validation failed for strategy %s", tc.stringRep)
				}
			}
		})
	}
}

func TestDeckDiscoverStatsDefaults(t *testing.T) {
	stats := deck.DiscoveryStats{}

	// Verify default values
	if stats.Evaluated != 0 {
		t.Errorf("Expected default Evaluated to be 0, got %d", stats.Evaluated)
	}

	if stats.Total != 0 {
		t.Errorf("Expected default Total to be 0, got %d", stats.Total)
	}

	if stats.Stored != 0 {
		t.Errorf("Expected default Stored to be 0, got %d", stats.Stored)
	}

	if stats.AvgScore != 0 {
		t.Errorf("Expected default AvgScore to be 0, got %f", stats.AvgScore)
	}

	if stats.BestScore != 0 {
		t.Errorf("Expected default BestScore to be 0, got %f", stats.BestScore)
	}

	if stats.Rate != 0 {
		t.Errorf("Expected default Rate to be 0, got %f", stats.Rate)
	}

	if len(stats.BestDeck) != 0 {
		t.Errorf("Expected default BestDeck to be empty, got %d cards", len(stats.BestDeck))
	}

	if len(stats.TopScores) != 0 {
		t.Errorf("Expected default TopScores to be empty, got %d scores", len(stats.TopScores))
	}
}

func TestDeckDiscoverCheckpointRoundTrip(t *testing.T) {
	tag := "ROUNDTRIP"

	// Create original checkpoint
	originalStats := deck.DiscoveryStats{
		Evaluated: 123,
		Total:     456,
		Stored:    78,
		AvgScore:  6.5,
		BestScore: 8.9,
		BestDeck:  []string{"Card1", "Card2", "Card3", "Card4", "Card5", "Card6", "Card7", "Card8"},
		StartTime: time.Now(),
		Elapsed:   10 * time.Minute,
		Rate:      12.3,
		ETA:       5 * time.Minute,
		Strategy:  deck.StrategyExhaustive,
		PlayerTag: "#" + tag,
		TopScores: []float64{8.9, 8.7, 8.5, 8.3, 8.1},
	}

	originalCheckpoint := deck.DiscoveryCheckpoint{
		GeneratorCheckpoint: &deck.GeneratorCheckpoint{
			Strategy:  deck.StrategyExhaustive,
			Position:  123,
			Generated: 123,
			State: map[string]interface{}{
				"test_key": "test_value",
			},
		},
		Stats:     originalStats,
		Timestamp: time.Now(),
		PlayerTag: "#" + tag,
		Strategy:  deck.StrategyExhaustive,
	}

	// Write checkpoint
	checkpointPath := createTestCheckpoint(t, tag, originalStats)
	defer os.Remove(checkpointPath)

	// Read back
	data, err := os.ReadFile(checkpointPath)
	if err != nil {
		t.Fatalf("failed to read checkpoint: %v", err)
	}

	var loadedCheckpoint deck.DiscoveryCheckpoint
	if err := json.Unmarshal(data, &loadedCheckpoint); err != nil {
		t.Fatalf("failed to unmarshal checkpoint: %v", err)
	}

	// Verify round-trip integrity
	if loadedCheckpoint.GeneratorCheckpoint.Strategy != originalCheckpoint.GeneratorCheckpoint.Strategy {
		t.Errorf("Strategy mismatch: got %s, want %s",
			loadedCheckpoint.GeneratorCheckpoint.Strategy,
			originalCheckpoint.GeneratorCheckpoint.Strategy)
	}

	if loadedCheckpoint.GeneratorCheckpoint.Generated != originalCheckpoint.GeneratorCheckpoint.Generated {
		t.Errorf("Generated mismatch: got %d, want %d",
			loadedCheckpoint.GeneratorCheckpoint.Generated,
			originalCheckpoint.GeneratorCheckpoint.Generated)
	}

	if loadedCheckpoint.Stats.Evaluated != originalCheckpoint.Stats.Evaluated {
		t.Errorf("Evaluated mismatch: got %d, want %d",
			loadedCheckpoint.Stats.Evaluated,
			originalCheckpoint.Stats.Evaluated)
	}

	if loadedCheckpoint.Stats.BestScore != originalCheckpoint.Stats.BestScore {
		t.Errorf("BestScore mismatch: got %f, want %f",
			loadedCheckpoint.Stats.BestScore,
			originalCheckpoint.Stats.BestScore)
	}

	// Verify deck cards match
	for i, card := range loadedCheckpoint.Stats.BestDeck {
		if card != originalCheckpoint.Stats.BestDeck[i] {
			t.Errorf("BestDeck[%d] mismatch: got %s, want %s", i, card, originalCheckpoint.Stats.BestDeck[i])
		}
	}
}
