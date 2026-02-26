package deck

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/leaderboard"
	"go.uber.org/ratelimit"
)

var (
	// ErrNoCheckpoint is returned when no checkpoint exists
	ErrNoCheckpoint = errors.New("no checkpoint found")

	// ErrInvalidCheckpoint is returned when checkpoint data is invalid
	ErrInvalidCheckpoint = errors.New("invalid checkpoint data")
)

// DiscoveryStats tracks progress and statistics during deck discovery
type DiscoveryStats struct {
	// Total number of decks evaluated
	Evaluated int

	// Total number of decks to evaluate (0 if unknown/unlimited)
	Total int

	// Number of decks added to leaderboard
	Stored int

	// Top 5 deck scores
	TopScores []float64

	// Average score of all evaluated decks
	AvgScore float64

	// Evaluation rate (decks per second)
	Rate float64

	// Estimated time remaining
	ETA time.Duration

	// Current best deck
	BestDeck  []string
	BestScore float64

	// Start time of discovery session
	StartTime time.Time

	// Elapsed time
	Elapsed time.Duration

	// Strategy being used
	Strategy GeneratorStrategy

	// Player tag
	PlayerTag string
}

// DiscoveryCheckpoint stores the state of a discovery session for resumption
type DiscoveryCheckpoint struct {
	// Generator checkpoint
	GeneratorCheckpoint *GeneratorCheckpoint `json:"generator_checkpoint"`

	// Statistics at checkpoint time
	Stats DiscoveryStats `json:"stats"`

	// Timestamp when checkpoint was saved
	Timestamp time.Time `json:"timestamp"`

	// Player tag
	PlayerTag string `json:"player_tag"`

	// Strategy
	Strategy GeneratorStrategy `json:"strategy"`
}

// DeckEvaluator evaluates a deck and returns a leaderboard entry
type DeckEvaluator interface {
	Evaluate(deck []string) (*leaderboard.DeckEntry, error)
}

// DiscoveryRunner orchestrates deck discovery with progress tracking and resumption
type DiscoveryRunner struct {
	generator   *DeckGenerator
	iterator    DeckIterator
	storage     *leaderboard.Storage
	rateLimiter ratelimit.Limiter
	evaluator   DeckEvaluator

	stats      DiscoveryStats
	statsMu    sync.RWMutex
	scoreSum   float64
	scoreCount int

	checkpointDir string
	playerTag     string
	strategy      GeneratorStrategy

	// Progress callback (optional)
	OnProgress func(DiscoveryStats)
}

// DiscoveryConfig configures a discovery runner
type DiscoveryConfig struct {
	// Generator configuration
	GeneratorConfig GeneratorConfig

	// Leaderboard storage
	Storage *leaderboard.Storage

	// Deck evaluator
	Evaluator DeckEvaluator

	// Progress callback (optional)
	OnProgress func(DiscoveryStats)

	// Player tag
	PlayerTag string
}

// NewDiscoveryRunner creates a new discovery runner
func NewDiscoveryRunner(config DiscoveryConfig) (*DiscoveryRunner, error) {
	if config.Evaluator == nil {
		return nil, errors.New("evaluator is required")
	}
	if config.Storage == nil {
		return nil, errors.New("storage is required")
	}
	if config.PlayerTag == "" {
		return nil, errors.New("player tag is required")
	}

	// Create generator
	generator, err := NewDeckGenerator(config.GeneratorConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create generator: %w", err)
	}

	// Create iterator
	iterator, err := generator.Iterator()
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator: %w", err)
	}

	// Get checkpoint directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	sanitizedTag, err := SanitizePlayerTag(config.PlayerTag)
	if err != nil {
		return nil, fmt.Errorf("invalid player tag: %w", err)
	}
	checkpointDir := filepath.Join(homeDir, ".cr-api", "discover")

	runner := &DiscoveryRunner{
		generator:     generator,
		iterator:      iterator,
		storage:       config.Storage,
		rateLimiter:   ratelimit.New(1, ratelimit.Per(time.Second)), // 1 req/sec
		evaluator:     config.Evaluator,
		checkpointDir: checkpointDir,
		playerTag:     sanitizedTag,
		strategy:      config.GeneratorConfig.Strategy,
		OnProgress:    config.OnProgress,
		stats: DiscoveryStats{
			StartTime: time.Now(),
			Strategy:  config.GeneratorConfig.Strategy,
			PlayerTag: "#" + sanitizedTag,
			TopScores: make([]float64, 0, 5),
		},
	}

	// Ensure checkpoint directory exists
	if err := os.MkdirAll(checkpointDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	return runner, nil
}

// Run executes the discovery process
//
//nolint:gocognit,gocyclo // Runner lifecycle intentionally keeps explicit checkpoint/error paths.
func (r *DiscoveryRunner) Run(ctx context.Context) (err error) {
	defer func() {
		if closeErr := r.iterator.Close(); closeErr != nil {
			// Log close error but prioritize any existing error
			if err == nil {
				err = fmt.Errorf("failed to close iterator: %w", closeErr)
			} else {
				fmt.Fprintf(os.Stderr, "warning: failed to close iterator: %v\n", closeErr)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			// Save checkpoint before exiting
			if err := r.SaveCheckpoint(); err != nil {
				return fmt.Errorf("failed to save checkpoint: %w", err)
			}
			return ctx.Err()

		default:
			// Apply rate limiting
			r.rateLimiter.Take()

			// Get next deck
			deck, err := r.iterator.Next(ctx)
			if err != nil {
				return fmt.Errorf("iterator error: %w", err)
			}

			// Check if exhausted
			if deck == nil {
				// Save final checkpoint
				if err := r.SaveCheckpoint(); err != nil {
					return fmt.Errorf("failed to save final checkpoint: %w", err)
				}
				return nil
			}

			// Evaluate deck
			if err := r.evaluateDeck(deck); err != nil {
				// Log error but continue
				fmt.Fprintf(os.Stderr, "Warning: failed to evaluate deck: %v\n", err)
				continue
			}

			// Update stats
			r.updateStats()

			// Call progress callback
			if r.OnProgress != nil {
				r.statsMu.RLock()
				statsCopy := r.stats
				r.statsMu.RUnlock()
				r.OnProgress(statsCopy)
			}
		}
	}
}

// evaluateDeck evaluates a deck and stores it in the leaderboard
func (r *DiscoveryRunner) evaluateDeck(deck []string) error {
	// Evaluate deck using the provided evaluator
	entry, err := r.evaluator.Evaluate(deck)
	if err != nil {
		return fmt.Errorf("failed to evaluate deck: %w", err)
	}

	// Add strategy metadata
	entry.Strategy = string(r.strategy)

	// Store in leaderboard
	_, _, err = r.storage.InsertDeck(entry)
	if err != nil {
		return fmt.Errorf("failed to store deck: %w", err)
	}

	// Update internal stats
	r.statsMu.Lock()
	r.scoreSum += entry.OverallScore
	r.scoreCount++
	r.stats.Stored++

	// Update best deck if necessary
	if entry.OverallScore > r.stats.BestScore {
		r.stats.BestScore = entry.OverallScore
		r.stats.BestDeck = deck
	}
	r.statsMu.Unlock()

	return nil
}

// updateStats updates discovery statistics
func (r *DiscoveryRunner) updateStats() {
	r.statsMu.Lock()
	defer r.statsMu.Unlock()

	r.stats.Evaluated++
	r.stats.Elapsed = time.Since(r.stats.StartTime)

	// Calculate average score
	if r.scoreCount > 0 {
		r.stats.AvgScore = r.scoreSum / float64(r.scoreCount)
	}

	// Calculate rate
	if r.stats.Elapsed.Seconds() > 0 {
		r.stats.Rate = float64(r.stats.Evaluated) / r.stats.Elapsed.Seconds()
	}

	// Calculate ETA
	if r.stats.Total > 0 && r.stats.Rate > 0 {
		remaining := r.stats.Total - r.stats.Evaluated
		r.stats.ETA = time.Duration(float64(remaining)/r.stats.Rate) * time.Second
	}

	// Update top scores from leaderboard
	topDecks, err := r.storage.Query(leaderboard.QueryOptions{
		SortBy:    "overall_score",
		SortOrder: "desc",
		Limit:     5,
	})
	if err == nil && len(topDecks) > 0 {
		r.stats.TopScores = make([]float64, len(topDecks))
		for i, deck := range topDecks {
			r.stats.TopScores[i] = deck.OverallScore
		}
	}
}

// GetStats returns a copy of current statistics
func (r *DiscoveryRunner) GetStats() DiscoveryStats {
	r.statsMu.RLock()
	defer r.statsMu.RUnlock()
	statsCopy := r.stats
	return statsCopy
}

// SaveCheckpoint saves the current state for resumption
func (r *DiscoveryRunner) SaveCheckpoint() error {
	// Get generator checkpoint
	genCheckpoint := r.iterator.Checkpoint()

	// Get current stats
	r.statsMu.RLock()
	stats := r.stats
	r.statsMu.RUnlock()

	// Create discovery checkpoint
	checkpoint := DiscoveryCheckpoint{
		GeneratorCheckpoint: genCheckpoint,
		Stats:               stats,
		Timestamp:           time.Now(),
		PlayerTag:           r.playerTag,
		Strategy:            r.strategy,
	}

	return SaveDiscoveryCheckpoint(r.getCheckpointPath(), checkpoint)
}

// Resume resumes from a saved checkpoint
func (r *DiscoveryRunner) Resume() error {
	checkpoint, err := LoadDiscoveryCheckpoint(r.getCheckpointPath())
	if err != nil {
		return err
	}

	// Resume iterator
	if err := r.iterator.Resume(checkpoint.GeneratorCheckpoint); err != nil {
		return fmt.Errorf("failed to resume iterator: %w", err)
	}

	// Restore stats
	r.statsMu.Lock()
	r.stats = checkpoint.Stats
	r.stats.StartTime = time.Now() // Reset start time for rate calculation
	r.scoreSum = r.stats.AvgScore * float64(r.stats.Evaluated)
	r.scoreCount = r.stats.Evaluated
	r.statsMu.Unlock()

	return nil
}

// HasCheckpoint checks if a checkpoint exists
func (r *DiscoveryRunner) HasCheckpoint() bool {
	checkpointPath := r.getCheckpointPath()
	_, err := os.Stat(checkpointPath)
	return err == nil
}

// ClearCheckpoint removes the checkpoint file
func (r *DiscoveryRunner) ClearCheckpoint() error {
	checkpointPath := r.getCheckpointPath()
	err := os.Remove(checkpointPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove checkpoint: %w", err)
	}
	return nil
}

// getCheckpointPath returns the path to the checkpoint file
func (r *DiscoveryRunner) getCheckpointPath() string {
	return DiscoveryCheckpointPath(r.checkpointDir, r.playerTag)
}

// GetStatusSummary returns a human-readable status summary
func (r *DiscoveryRunner) GetStatusSummary() string {
	r.statsMu.RLock()
	defer r.statsMu.RUnlock()

	var sb strings.Builder
	fmt.Fprintf(&sb, "Discovery Status for %s\n", r.stats.PlayerTag)
	fmt.Fprintf(&sb, "Strategy: %s\n", r.stats.Strategy)
	fmt.Fprintf(&sb, "Progress: %d decks evaluated", r.stats.Evaluated)
	if r.stats.Total > 0 {
		fmt.Fprintf(&sb, " / %d total (%.1f%%)", r.stats.Total,
			float64(r.stats.Evaluated)/float64(r.stats.Total)*100)
	}
	fmt.Fprintf(&sb, "\n")
	fmt.Fprintf(&sb, "Stored: %d decks in leaderboard\n", r.stats.Stored)
	fmt.Fprintf(&sb, "Rate: %.2f decks/sec\n", r.stats.Rate)
	if r.stats.ETA > 0 {
		fmt.Fprintf(&sb, "ETA: %v\n", r.stats.ETA.Round(time.Second))
	}
	fmt.Fprintf(&sb, "Average Score: %.2f\n", r.stats.AvgScore)
	if len(r.stats.BestDeck) > 0 {
		fmt.Fprintf(&sb, "Best Deck: %.2f - %v\n", r.stats.BestScore, r.stats.BestDeck)
	}
	if len(r.stats.TopScores) > 0 {
		fmt.Fprintf(&sb, "Top 5 Scores: ")
		for i, score := range r.stats.TopScores {
			if i > 0 {
				fmt.Fprintf(&sb, ", ")
			}
			fmt.Fprintf(&sb, "%.2f", score)
		}
		fmt.Fprintf(&sb, "\n")
	}

	return sb.String()
}
