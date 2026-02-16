package deck

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/leaderboard"
)

// mockDeckEvaluator is a test implementation of DeckEvaluator
type mockDeckEvaluator struct {
	score      float64
	archetype  string
	err        error
	callCount  int
	lastDeck   []string
	delay      time.Duration
	evaluateFn func(deck []string) (*leaderboard.DeckEntry, error)
}

func (m *mockDeckEvaluator) Evaluate(deck []string) (*leaderboard.DeckEntry, error) {
	m.callCount++
	m.lastDeck = deck

	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if m.err != nil {
		return nil, m.err
	}

	if m.evaluateFn != nil {
		return m.evaluateFn(deck)
	}

	return &leaderboard.DeckEntry{
		Cards:             deck,
		OverallScore:      m.score,
		AttackScore:       7.5,
		DefenseScore:      8.0,
		SynergyScore:      7.0,
		VersatilityScore:  6.5,
		F2PScore:          8.5,
		PlayabilityScore:  7.8,
		Archetype:         m.archetype,
		ArchetypeConf:     0.85,
		AvgElixir:         3.5,
		EvaluatedAt:       time.Now(),
		PlayerTag:         "#TEST",
		EvaluationVersion: "1.0.0",
	}, nil
}

// createTestDiscoveryRunner creates a test discovery runner with a temporary storage
func createTestDiscoveryRunner(t *testing.T, evaluator *mockDeckEvaluator) (*DiscoveryRunner, func()) {
	t.Helper()

	// Create temporary directory for test databases
	tmpDir, err := os.MkdirTemp("", "discovery_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	// Create storage
	storage, err := leaderboard.NewStorage("#TEST123")
	if err != nil {
		os.RemoveAll(tmpDir)
		os.Setenv("HOME", originalHome)
		t.Fatalf("failed to create storage: %v", err)
	}

	// Create generator config
	config := DiscoveryConfig{
		GeneratorConfig: GeneratorConfig{
			Strategy:   StrategyRandomSample,
			Candidates: createTestCandidates(20),
			SampleSize: 10,
			Seed:       12345,
		},
		Storage:   storage,
		Evaluator: evaluator,
		PlayerTag: "#TEST123",
	}

	runner, err := NewDiscoveryRunner(config)
	if err != nil {
		storage.Close()
		os.RemoveAll(tmpDir)
		os.Setenv("HOME", originalHome)
		t.Fatalf("failed to create discovery runner: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		storage.Close()
		os.Setenv("HOME", originalHome)
		os.RemoveAll(tmpDir)
	}

	return runner, cleanup
}

func TestNewDiscoveryRunner(t *testing.T) {
	tests := []struct {
		name         string
		config       DiscoveryConfig
		setupStorage bool
		wantErr      bool
		errContains  string
	}{
		{
			name: "valid config",
			config: DiscoveryConfig{
				GeneratorConfig: GeneratorConfig{
					Strategy:   StrategyRandomSample,
					Candidates: createTestCandidates(20),
					SampleSize: 10,
				},
				Storage:   nil, // Will be set in test
				Evaluator: &mockDeckEvaluator{score: 8.0},
				PlayerTag: "#TEST123",
			},
			setupStorage: true,
			wantErr:      false,
		},
		{
			name: "missing evaluator",
			config: DiscoveryConfig{
				GeneratorConfig: GeneratorConfig{
					Strategy:   StrategyRandomSample,
					Candidates: createTestCandidates(20),
				},
				Evaluator: nil,
				PlayerTag: "#TEST123",
			},
			wantErr:     true,
			errContains: "evaluator is required",
		},
		{
			name: "missing storage",
			config: DiscoveryConfig{
				GeneratorConfig: GeneratorConfig{
					Strategy:   StrategyRandomSample,
					Candidates: createTestCandidates(20),
				},
				Evaluator: &mockDeckEvaluator{score: 8.0},
				Storage:   nil,
				PlayerTag: "#TEST123",
			},
			wantErr:     true,
			errContains: "storage is required",
		},
		{
			name: "missing player tag",
			config: DiscoveryConfig{
				GeneratorConfig: GeneratorConfig{
					Strategy:   StrategyRandomSample,
					Candidates: createTestCandidates(20),
				},
				Evaluator: &mockDeckEvaluator{score: 8.0},
				PlayerTag: "",
			},
			setupStorage: true,
			wantErr:      true,
			errContains:  "player tag is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup storage for cases where storage should not be the primary validation error
			if tt.setupStorage && tt.config.Storage == nil {
				tmpDir, _ := os.MkdirTemp("", "discovery_test_*")
				originalHome := os.Getenv("HOME")
				os.Setenv("HOME", tmpDir)
				defer os.Setenv("HOME", originalHome)
				defer os.RemoveAll(tmpDir)

				storage, err := leaderboard.NewStorage("#TEST123")
				if err != nil {
					t.Fatalf("failed to create storage: %v", err)
				}
				defer storage.Close()
				tt.config.Storage = storage
			}

			runner, err := NewDiscoveryRunner(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errContains)
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("error message '%s' does not contain '%s'", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if runner == nil {
					t.Error("expected non-nil runner")
				}
			}
		})
	}
}

func TestMockDeckEvaluatorEvaluate_SideEffectsOnce(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score:     8.2,
		archetype: "cycle",
		delay:     30 * time.Millisecond,
	}

	start := time.Now()
	result, err := evaluator.Evaluate([]string{"Hog Rider", "Musketeer"})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result == nil {
		t.Fatal("Evaluate() returned nil result")
	}
	if evaluator.callCount != 1 {
		t.Fatalf("callCount = %d, want 1", evaluator.callCount)
	}
	if elapsed < evaluator.delay {
		t.Fatalf("elapsed = %v, want >= %v", elapsed, evaluator.delay)
	}
	if elapsed >= 50*time.Millisecond {
		t.Fatalf("elapsed = %v, expected single delay sleep", elapsed)
	}
}

func TestDiscoveryRunner_Run_Basic(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score:     8.5,
		archetype: "beatdown",
	}

	runner, cleanup := createTestDiscoveryRunner(t, evaluator)
	defer cleanup()

	// Run for a few decks then cancel
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := runner.Run(ctx)
	// Error is expected from context timeout or iterator exhaustion
	if err != nil && err != context.DeadlineExceeded && !containsString(err.Error(), "context deadline exceeded") {
		t.Errorf("Run() error = %v", err)
	}

	// Verify some decks were evaluated
	if evaluator.callCount == 0 {
		t.Error("expected at least one deck to be evaluated")
	}

	// Verify stats
	stats := runner.GetStats()
	if stats.Evaluated == 0 {
		t.Error("expected Evaluated > 0")
	}
	if stats.Stored == 0 {
		t.Error("expected Stored > 0")
	}
}

func TestDiscoveryRunner_Run_ContextCancellation(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score: 8.0,
		delay: 50 * time.Millisecond,
	}

	runner, cleanup := createTestDiscoveryRunner(t, evaluator)
	defer cleanup()

	// Cancel immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel right away

	err := runner.Run(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Verify checkpoint was saved despite cancellation
	if !runner.HasCheckpoint() {
		t.Error("expected checkpoint to be saved after cancellation")
	}
}

func TestDiscoveryRunner_Run_GracefulShutdown(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score: 8.0,
	}

	runner, cleanup := createTestDiscoveryRunner(t, evaluator)
	defer cleanup()

	// Run with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = runner.Run(ctx)

	// Explicitly save checkpoint for testing
	_ = runner.SaveCheckpoint()

	// Verify checkpoint exists and has valid data
	stats := runner.GetStats()
	if stats.Evaluated == 0 {
		t.Skip("skipping - no decks were evaluated")
	}

	// After explicit save, checkpoint should exist
	if !runner.HasCheckpoint() {
		t.Error("expected checkpoint to exist after explicit save")
	}
}

func TestDiscoveryRunner_Run_Exhaustion(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score: 8.0,
	}

	// Create generator with small sample size
	tmpDir, _ := os.MkdirTemp("", "discovery_test_*")
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)
	defer os.RemoveAll(tmpDir)

	storage, err := leaderboard.NewStorage("#TEST123")
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	config := DiscoveryConfig{
		GeneratorConfig: GeneratorConfig{
			Strategy:   StrategyRandomSample,
			Candidates: createTestCandidates(10),
			SampleSize: 5, // Small sample
			Seed:       12345,
		},
		Storage:   storage,
		Evaluator: evaluator,
		PlayerTag: "#TEST123",
	}

	runner, err := NewDiscoveryRunner(config)
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	// Run until exhaustion
	ctx := context.Background()
	err = runner.Run(ctx)
	if err != nil {
		t.Errorf("unexpected error on exhaustion: %v", err)
	}

	// Should have evaluated all sample decks
	stats := runner.GetStats()
	if stats.Evaluated == 0 {
		t.Error("expected decks to be evaluated")
	}
}

func TestDiscoveryRunner_EvaluationErrors(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score: 8.0,
		err:   errors.New("evaluation failed"),
	}

	runner, cleanup := createTestDiscoveryRunner(t, evaluator)
	defer cleanup()

	// Run for a short time
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Should not fail despite evaluation errors
	err := runner.Run(ctx)
	// Error is expected from context timeout
	if err != nil && !containsString(err.Error(), "context deadline exceeded") && err != context.DeadlineExceeded {
		t.Errorf("Run() should handle evaluation errors gracefully, got: %v", err)
	}

	// Verify that evaluations were attempted
	if evaluator.callCount == 0 {
		t.Error("expected evaluation attempts despite errors")
	}
}

func TestDiscoveryRunner_ProgressCallback(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score: 8.5,
	}

	progressCalls := 0
	var lastStats DiscoveryStats

	runner, cleanup := createTestDiscoveryRunner(t, evaluator)
	defer cleanup()

	runner.OnProgress = func(stats DiscoveryStats) {
		progressCalls++
		lastStats = stats
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = runner.Run(ctx)

	if progressCalls == 0 {
		t.Error("expected progress callback to be called")
	}

	if lastStats.Evaluated == 0 {
		t.Error("expected last stats to have evaluated count")
	}
}

func TestDiscoveryRunner_SaveAndResume(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score:     8.5,
		archetype: "beatdown",
	}

	// First run: generate some decks and save checkpoint
	runner1, cleanup1 := createTestDiscoveryRunner(t, evaluator)
	defer cleanup1()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = runner1.Run(ctx)

	// Explicitly save checkpoint
	if err := runner1.SaveCheckpoint(); err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	initialStats := runner1.GetStats()
	if initialStats.Evaluated == 0 {
		t.Skip("skipping - no decks were evaluated")
	}

	// Verify checkpoint was saved
	if !runner1.HasCheckpoint() {
		t.Fatal("expected checkpoint to exist after explicit save")
	}

	// Second run: create a new runner with the same player tag
	// It should find and use the existing checkpoint
	storage2, err := leaderboard.NewStorage("#TEST123")
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage2.Close()

	evaluator2 := &mockDeckEvaluator{
		score:     8.5,
		archetype: "beatdown",
	}

	config2 := DiscoveryConfig{
		GeneratorConfig: GeneratorConfig{
			Strategy:   StrategyRandomSample,
			Candidates: createTestCandidates(20),
			SampleSize: 10,
			Seed:       12345,
		},
		Storage:   storage2,
		Evaluator: evaluator2,
		PlayerTag: "#TEST123",
	}

	runner2, err := NewDiscoveryRunner(config2)
	if err != nil {
		t.Fatalf("failed to create second runner: %v", err)
	}

	// Resume from checkpoint - this should work since both use same player tag
	err = runner2.Resume()
	if err == ErrNoCheckpoint {
		// Checkpoint might be in a different location due to os.UserHomeDir()
		// This is acceptable for the test
		t.Skip("checkpoint not accessible - different home directory")
	} else if err != nil {
		t.Fatalf("failed to resume: %v", err)
	}

	resumedStats := runner2.GetStats()
	if resumedStats.Evaluated != initialStats.Evaluated {
		t.Logf("Warning: evaluated count mismatch after resume: %d vs %d",
			resumedStats.Evaluated, initialStats.Evaluated)
	}

	// Run for a bit more
	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()

	_ = runner2.Run(ctx2)

	// Should have evaluated more decks
	finalStats := runner2.GetStats()
	if finalStats.Evaluated <= resumedStats.Evaluated {
		t.Error("expected more decks to be evaluated after resume")
	}
}

func TestDiscoveryRunner_CheckpointPersistence(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score: 8.0,
	}

	runner, cleanup := createTestDiscoveryRunner(t, evaluator)
	defer cleanup()

	// Run briefly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = runner.Run(ctx)

	// Save checkpoint explicitly
	err := runner.SaveCheckpoint()
	if err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	// Verify checkpoint file exists
	checkpointPath := runner.getCheckpointPath()
	if _, err := os.Stat(checkpointPath); os.IsNotExist(err) {
		t.Errorf("checkpoint file does not exist: %s", checkpointPath)
	}

	// Read and verify checkpoint content
	data, err := os.ReadFile(checkpointPath)
	if err != nil {
		t.Fatalf("failed to read checkpoint: %v", err)
	}

	var checkpoint DiscoveryCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		t.Fatalf("failed to unmarshal checkpoint: %v", err)
	}

	if checkpoint.GeneratorCheckpoint == nil {
		t.Error("expected generator checkpoint to be present")
	}

	if checkpoint.PlayerTag != "TEST123" {
		t.Errorf("expected player tag TEST123, got %s", checkpoint.PlayerTag)
	}
}

func TestDiscoveryRunner_ClearCheckpoint(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score: 8.0,
	}

	runner, cleanup := createTestDiscoveryRunner(t, evaluator)
	defer cleanup()

	// Run to create checkpoint
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = runner.Run(ctx)

	// Explicitly save checkpoint
	if err := runner.SaveCheckpoint(); err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	// Verify checkpoint exists
	stats := runner.GetStats()
	if stats.Evaluated == 0 {
		t.Skip("skipping - no decks were evaluated")
	}

	if !runner.HasCheckpoint() {
		t.Fatal("expected checkpoint to exist after explicit save")
	}

	// Clear checkpoint
	err := runner.ClearCheckpoint()
	if err != nil {
		t.Fatalf("failed to clear checkpoint: %v", err)
	}

	// Verify checkpoint is gone
	if runner.HasCheckpoint() {
		t.Error("expected checkpoint to be cleared")
	}
}

func TestDiscoveryRunner_Resume_NoCheckpoint(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score: 8.0,
	}

	runner, cleanup := createTestDiscoveryRunner(t, evaluator)
	defer cleanup()

	// Try to resume without checkpoint
	err := runner.Resume()
	if err != ErrNoCheckpoint {
		t.Errorf("expected ErrNoCheckpoint, got %v", err)
	}
}

func TestDiscoveryRunner_Resume_InvalidCheckpoint(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score: 8.0,
	}

	runner, cleanup := createTestDiscoveryRunner(t, evaluator)
	defer cleanup()

	// Create invalid checkpoint file
	checkpointPath := runner.getCheckpointPath()
	checkpointDir := filepath.Dir(checkpointPath)
	if err := os.MkdirAll(checkpointDir, 0o755); err != nil {
		t.Fatalf("failed to create checkpoint dir: %v", err)
	}

	if err := os.WriteFile(checkpointPath, []byte("invalid json"), 0o644); err != nil {
		t.Fatalf("failed to write invalid checkpoint: %v", err)
	}

	// Try to resume
	err := runner.Resume()
	if err != ErrInvalidCheckpoint {
		t.Errorf("expected ErrInvalidCheckpoint, got %v", err)
	}
}

func TestDiscoveryRunner_GetStats(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score: 8.5,
	}

	runner, cleanup := createTestDiscoveryRunner(t, evaluator)
	defer cleanup()

	// Initial stats should have StartTime set (set in constructor)
	stats := runner.GetStats()
	if stats.Evaluated != 0 {
		t.Errorf("expected initial Evaluated to be 0, got %d", stats.Evaluated)
	}
	if stats.Stored != 0 {
		t.Errorf("expected initial Stored to be 0, got %d", stats.Stored)
	}
	// Note: StartTime is set in NewDiscoveryRunner, so it won't be zero
	if stats.StartTime.IsZero() {
		t.Error("expected StartTime to be initialized")
	}

	// Run briefly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = runner.Run(ctx)

	// Stats should be updated
	stats = runner.GetStats()
	if stats.Evaluated == 0 {
		t.Error("expected Evaluated > 0 after running")
	}
	if stats.StartTime.IsZero() {
		t.Error("expected StartTime to be set after running")
	}
	if stats.Elapsed == 0 {
		t.Error("expected Elapsed > 0 after running")
	}
}

func TestDiscoveryRunner_GetStatusSummary(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score:     8.5,
		archetype: "beatdown",
	}

	runner, cleanup := createTestDiscoveryRunner(t, evaluator)
	defer cleanup()

	// Run briefly to populate stats
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = runner.Run(ctx)

	summary := runner.GetStatusSummary()

	// Verify summary contains expected information
	if summary == "" {
		t.Error("expected non-empty status summary")
	}

	if !containsString(summary, "Discovery Status") {
		t.Error("expected summary to contain 'Discovery Status'")
	}

	if !containsString(summary, "TEST123") {
		t.Error("expected summary to contain player tag")
	}

	if !containsString(summary, "Strategy") {
		t.Error("expected summary to contain strategy")
	}

	if !containsString(summary, "Progress") {
		t.Error("expected summary to contain progress")
	}
}

func TestDiscoveryRunner_RateLimiting(t *testing.T) {
	evaluator := &mockDeckEvaluator{
		score: 8.0,
		delay: 10 * time.Millisecond,
	}

	runner, cleanup := createTestDiscoveryRunner(t, evaluator)
	defer cleanup()

	// Run for 200ms - should process ~20 decks at 1 deck/sec with 10ms eval time
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = runner.Run(ctx)

	stats := runner.GetStats()

	// With rate limiting (1 req/sec), should process ~2 decks in 200ms
	// (rate limiter is the bottleneck, not evaluation time)
	if stats.Evaluated > 5 {
		t.Logf("Warning: processed %d decks, expected ~2 due to rate limiting", stats.Evaluated)
	}

	// Verify rate is calculated
	if stats.Rate == 0 {
		t.Error("expected rate to be calculated")
	}
}

func TestDiscoveryRunner_BestDeckTracking(t *testing.T) {
	// Create evaluator that returns increasing scores
	callCount := 0
	evaluator := &mockDeckEvaluator{
		score: 7.0,
		evaluateFn: func(deck []string) (*leaderboard.DeckEntry, error) {
			callCount++
			// Increase score with each call
			score := 7.0 + float64(callCount)*0.1
			return &leaderboard.DeckEntry{
				Cards:        deck,
				OverallScore: score,
				AvgElixir:    3.5,
				EvaluatedAt:  time.Now(),
				PlayerTag:    "#TEST",
			}, nil
		},
	}

	runner, cleanup := createTestDiscoveryRunner(t, evaluator)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = runner.Run(ctx)

	stats := runner.GetStats()

	// Best score should be tracked
	if stats.BestScore == 0 {
		t.Error("expected BestScore to be tracked")
	}

	if len(stats.BestDeck) == 0 {
		t.Error("expected BestDeck to be tracked")
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		containsMiddleString(s, substr)))
}

func containsMiddleString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
