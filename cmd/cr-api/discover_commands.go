package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/config"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/klauer/clash-royale-api/go/pkg/leaderboard"
	"github.com/urfave/cli/v3"
)

// Simple evaluator implementation
type simpleEvaluator struct {
	synergyDB    *deck.SynergyDatabase
	cardRegistry *clashroyale.CardStatsRegistry
	playerCtx    *evaluation.PlayerContext
	playerTag    string
}

func (e *simpleEvaluator) Evaluate(deckCards []string) (*leaderboard.DeckEntry, error) {
	// Convert card names to candidates
	candidates, err := e.resolveCandidates(deckCards)
	if err != nil {
		return nil, err
	}

	// Score the deck
	attack := evaluation.ScoreAttack(candidates)
	defense := evaluation.ScoreDefense(candidates)
	synergy := evaluation.ScoreSynergy(candidates, e.synergyDB)
	versatility := evaluation.ScoreVersatility(candidates)
	f2p := evaluation.ScoreF2P(candidates)
	playability := evaluation.ScorePlayability(candidates, e.playerCtx)

	// Calculate overall score
	overallScore := (attack.Score + defense.Score + synergy.Score +
		versatility.Score + f2p.Score + playability.Score) / 6.0

	// Detect archetype
	archetypeResult := evaluation.DetectArchetype(candidates)

	// Calculate average elixir
	avgElixir := 0.0
	for _, card := range candidates {
		avgElixir += float64(card.Elixir)
	}
	avgElixir /= float64(len(candidates))

	// Create leaderboard entry
	entry := &leaderboard.DeckEntry{
		Cards:             deckCards,
		OverallScore:      overallScore,
		AttackScore:       attack.Score,
		DefenseScore:      defense.Score,
		SynergyScore:      synergy.Score,
		VersatilityScore:  versatility.Score,
		F2PScore:          f2p.Score,
		PlayabilityScore:  playability.Score,
		Archetype:         string(archetypeResult.Primary),
		ArchetypeConf:     archetypeResult.PrimaryConfidence,
		AvgElixir:         avgElixir,
		EvaluatedAt:       time.Now(),
		PlayerTag:         e.playerTag,
		EvaluationVersion: "1.0",
	}

	return entry, nil
}

func (e *simpleEvaluator) resolveCandidates(cardNames []string) ([]deck.CardCandidate, error) {
	candidates := make([]deck.CardCandidate, 0, len(cardNames))

	for _, name := range cardNames {
		// Get card stats
		stats := e.cardRegistry.GetStats(name)
		if stats == nil {
			return nil, fmt.Errorf("card not found: %s", name)
		}

		// Get player's level for this card
		levelInfo, hasCard := e.playerCtx.Collection[name]
		if !hasCard {
			return nil, fmt.Errorf("player doesn't own card: %s", name)
		}

		// Get rarity and elixir from card data
		rarity := levelInfo.Rarity
		if rarity == "" {
			rarity = "Common"
		}
		// Get elixir cost from config
		elixir := config.GetCardElixir(name, 0)

		// Determine card role with evolution awareness
		role := config.GetCardRoleWithEvolution(name, levelInfo.EvolutionLevel)

		// Check evolution status
		hasEvolution := false
		evolutionLevel := 0
		if e.playerCtx.UnlockedEvolutions != nil {
			hasEvolution = e.playerCtx.UnlockedEvolutions[name]
			if hasEvolution {
				evolutionLevel = levelInfo.EvolutionLevel
			}
		}

		// Create candidate
		candidate := deck.CardCandidate{
			Name:              name,
			Level:             levelInfo.Level,
			MaxLevel:          levelInfo.MaxLevel,
			Rarity:            rarity,
			Elixir:            elixir,
			Role:              &role,
			Score:             0, // Not used
			HasEvolution:      hasEvolution,
			EvolutionLevel:    evolutionLevel,
			MaxEvolutionLevel: levelInfo.MaxEvolutionLevel,
			Stats:             stats,
		}

		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

// Add discover commands to deck command
func addDiscoverCommands() *cli.Command {
	return &cli.Command{
		Name:  "discover",
		Usage: "Discover optimal deck combinations with resumable evaluation",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a new deck discovery evaluation session",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "strategy",
						Value: string(deck.StrategySmartSample),
						Usage: "Sampling strategy (exhaustive, smart, random, archetype)",
					},
					&cli.IntFlag{
						Name:  "sample-size",
						Value: 1000,
						Usage: "Number of decks to generate (for sampling strategies)",
					},
					&cli.IntFlag{
						Name:  "limit",
						Usage: "Maximum number of decks to evaluate (0 for unlimited)",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show detailed progress",
					},
					&cli.BoolFlag{
						Name:  "background",
						Usage: "Run discovery in background as daemon process",
					},
				},
				Action: deckDiscoverStartCommand,
			},
			{
				Name:  "run",
				Usage: "Run deck discovery evaluation session (alias for start)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "strategy",
						Value: string(deck.StrategySmartSample),
						Usage: "Sampling strategy (exhaustive, smart, random, archetype)",
					},
					&cli.IntFlag{
						Name:  "sample-size",
						Value: 1000,
						Usage: "Number of decks to generate (for sampling strategies)",
					},
					&cli.IntFlag{
						Name:  "limit",
						Usage: "Maximum number of decks to evaluate (0 for unlimited)",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show detailed progress",
					},
					&cli.BoolFlag{
						Name:  "resume",
						Usage: "Resume from last checkpoint if available",
					},
					&cli.BoolFlag{
						Name:  "background",
						Usage: "Run discovery in background as daemon process",
					},
				},
				Action: deckDiscoverRunCommand,
			},
			{
				Name:  "stop",
				Usage: "Stop a running discovery session gracefully",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
				},
				Action: deckDiscoverStopCommand,
			},
			{
				Name:  "resume",
				Usage: "Resume a discovery session from checkpoint",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show detailed progress",
					},
					&cli.BoolFlag{
						Name:  "background",
						Usage: "Run discovery in background as daemon process",
					},
				},
				Action: deckDiscoverResumeCommand,
			},
			{
				Name:  "status",
				Usage: "Show status of discovery session",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
				},
				Action: deckDiscoverStatusCommand,
			},
			{
				Name:  "stats",
				Usage: "Show detailed session statistics",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
				},
				Action: deckDiscoverStatsCommand,
			},
		},
	}
}

func deckDiscoverStatusCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")

	// Check for checkpoint
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	sanitizedTag := strings.TrimPrefix(playerTag, "#")
	checkpointPath := filepath.Join(homeDir, ".cr-api", "discover", fmt.Sprintf("%s.json", sanitizedTag))

	if _, err := os.Stat(checkpointPath); os.IsNotExist(err) {
		printf("No active discovery session found for player #%s\n", playerTag)
		return nil
	}

	// Read checkpoint
	data, err := os.ReadFile(checkpointPath)
	if err != nil {
		return fmt.Errorf("failed to read checkpoint: %w", err)
	}

	// Parse checkpoint
	var checkpoint deck.DiscoveryCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return fmt.Errorf("failed to parse checkpoint: %w", err)
	}

	// Display status
	printf("Discovery Session Status for #%s\n", playerTag)
	printf("Strategy: %s\n", checkpoint.Strategy)
	printf("Evaluated: %d decks\n", checkpoint.Stats.Evaluated)
	printf("Stored: %d decks\n", checkpoint.Stats.Stored)
	printf("Average Score: %.2f\n", checkpoint.Stats.AvgScore)
	printf("Best Score: %.2f\n", checkpoint.Stats.BestScore)
	if len(checkpoint.Stats.BestDeck) > 0 {
		printf("Best Deck: %v\n", checkpoint.Stats.BestDeck)
	}
	printf("Last Updated: %s\n", checkpoint.Timestamp.Format(time.RFC3339))
	printf("\nResume with: cr-api deck discover resume --tag %s\n", playerTag)

	return nil
}

// deckDiscoverStartCommand starts a new discovery session
func deckDiscoverStartCommand(ctx context.Context, cmd *cli.Command) error {
	return runDiscoveryCommand(ctx, cmd, false)
}

// deckDiscoverRunCommand runs a discovery session (with optional resume)
func deckDiscoverRunCommand(ctx context.Context, cmd *cli.Command) error {
	resume := cmd.Bool("resume")
	return runDiscoveryCommand(ctx, cmd, resume)
}

// runDiscoveryCommand is the shared implementation for start/run commands
func runDiscoveryCommand(ctx context.Context, cmd *cli.Command, resume bool) error {
	playerTag := cmd.String("tag")
	strategy := deck.GeneratorStrategy(cmd.String("strategy"))
	sampleSize := cmd.Int("sample-size")
	limit := cmd.Int("limit")
	verbose := cmd.Bool("verbose")
	background := cmd.Bool("background")

	// Check for existing checkpoint when starting fresh (not resuming)
	if !resume {
		homeDir, _ := os.UserHomeDir()
		sanitizedTag := strings.TrimPrefix(playerTag, "#")
		checkpointPath := filepath.Join(homeDir, ".cr-api", "discover", fmt.Sprintf("%s.json", sanitizedTag))
		if _, err := os.Stat(checkpointPath); err == nil {
			fprintf(os.Stderr, "Warning: Existing checkpoint found. Use --resume or 'cr-api deck discover resume' to continue.\n")
			fprintf(os.Stderr, "Starting fresh will clear the existing checkpoint.\n")
		}
	}

	// Handle background mode
	if background {
		return runDiscoveryInBackground(ctx, cmd, resume)
	}

	// Get API token
	apiToken := os.Getenv("CLASH_ROYALE_API_TOKEN")
	if apiToken == "" {
		return fmt.Errorf("CLASH_ROYALE_API_TOKEN environment variable required")
	}

	// Fetch player data
	if verbose {
		fprintf(os.Stderr, "Fetching player data for #%s...\n", playerTag)
	}
	client := clashroyale.NewClient(apiToken)
	cleanTag := strings.TrimPrefix(playerTag, "#")
	player, err := client.GetPlayer(cleanTag)
	if err != nil {
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	// Get data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	dataDir := filepath.Join(homeDir, ".cr-api")

	// Load card stats
	statsPath := filepath.Join(dataDir, "cards_stats.json")
	statsRegistry, err := clashroyale.LoadStats(statsPath)
	if err != nil {
		return fmt.Errorf("failed to load card stats: %w", err)
	}

	// Create player context
	playerCtx := evaluation.NewPlayerContextFromPlayer(player)

	// Load synergy database
	synergyDB := deck.NewSynergyDatabase()

	// Create leaderboard storage
	storage, err := leaderboard.NewStorage(playerTag)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}
	defer storage.Close()

	// Build candidates from player collection
	candidates, err := buildGeneticCandidates(player, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to build candidates: %w", err)
	}

	// Create generator config
	genConfig := deck.GeneratorConfig{
		Strategy:   strategy,
		Candidates: candidates,
		SampleSize: sampleSize,
		Constraints: &deck.GeneratorConstraints{
			MinAvgElixir:        2.0,
			MaxAvgElixir:        5.0,
			RequireWinCondition: true,
		},
	}

	// Create evaluator
	evaluator := &simpleEvaluator{
		synergyDB:    synergyDB,
		cardRegistry: statsRegistry,
		playerCtx:    playerCtx,
		playerTag:    playerTag,
	}

	// Create discovery runner
	runner, err := deck.NewDiscoveryRunner(deck.DiscoveryConfig{
		GeneratorConfig: genConfig,
		Storage:         storage,
		Evaluator:       evaluator,
		PlayerTag:       playerTag,
		OnProgress: func(stats deck.DiscoveryStats) {
			if verbose {
				fprintf(os.Stderr, "\r[%s] Evaluated: %d | Stored: %d | Best: %.2f | Avg: %.2f | Rate: %.1f/s",
					stats.Strategy, stats.Evaluated, stats.Stored, stats.BestScore,
					stats.AvgScore, stats.Rate)
			}
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}

	// Resume if requested and checkpoint exists
	if resume && runner.HasCheckpoint() {
		if verbose {
			fprintf(os.Stderr, "Resuming from checkpoint...\n")
		}
		if err := runner.Resume(); err != nil {
			fprintf(os.Stderr, "Warning: failed to resume from checkpoint: %v\n", err)
			fprintf(os.Stderr, "Starting fresh...\n")
		}
	}

	// Set up signal handling for graceful shutdown
	var interrupted atomic.Bool
	var canceler stageCanceler

	interrupts := make(chan os.Signal, 2)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupts)

	go func() {
		<-interrupts
		if interrupted.CompareAndSwap(false, true) {
			fprintf(os.Stderr, "\nInterrupt received; saving checkpoint and stopping...\n")
			canceler.Cancel()
		}
		<-interrupts
		fprintf(os.Stderr, "\nSecond interrupt received; exiting immediately.\n")
		os.Exit(130)
	}()

	// Create context with cancellation
	runCtx, cancelRun := context.WithCancel(ctx)
	canceler.Set(cancelRun)
	defer cancelRun()

	// Run discovery
	if verbose {
		fprintf(os.Stderr, "Starting discovery with strategy: %s\n", strategy)
	}

	// Apply limit if specified
	if limit > 0 {
		runCtx, cancelRun = context.WithCancel(runCtx)
		defer cancelRun()
		go func() {
			stats := runner.GetStats()
			for stats.Evaluated < limit {
				time.Sleep(time.Second)
				stats = runner.GetStats()
			}
			cancelRun()
		}()
	}

	err = runner.Run(runCtx)
	canceler.Clear()

	// Handle result
	if err != nil {
		if errors.Is(err, context.Canceled) {
			// Graceful shutdown - checkpoint already saved
			fprintf(os.Stderr, "\nDiscovery stopped. Checkpoint saved. Use 'cr-api deck discover resume' to continue.\n")
			stats := runner.GetStats()
			fprintf(os.Stderr, "\nProgress: %d decks evaluated, %d stored\n", stats.Evaluated, stats.Stored)
			if len(stats.BestDeck) > 0 {
				fprintf(os.Stderr, "Best deck so far: %.2f - %v\n", stats.BestScore, stats.BestDeck)
			}
			return nil
		}
		return err
	}

	// Discovery completed
	if verbose {
		fprintf(os.Stderr, "\nDiscovery complete!\n")
	}
	stats := runner.GetStats()
	printf("\nFinal Statistics:\n")
	printf("  Evaluated: %d decks\n", stats.Evaluated)
	printf("  Stored: %d decks\n", stats.Stored)
	printf("  Average Score: %.2f\n", stats.AvgScore)
	printf("  Best Score: %.2f\n", stats.BestScore)
	if len(stats.BestDeck) > 0 {
		printf("  Best Deck: %v\n", stats.BestDeck)
	}
	printf("\nView results with: cr-api deck leaderboard show --tag %s\n", playerTag)

	// Clear checkpoint on successful completion
	runner.ClearCheckpoint()

	return nil
}

// deckDiscoverStopCommand stops a running discovery session
func deckDiscoverStopCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	sanitizedTag := strings.TrimPrefix(playerTag, "#")

	// Check for PID file
	pidFile := filepath.Join(homeDir, ".cr-api", "discover", fmt.Sprintf("%s.pid", sanitizedTag))
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		// Check if there's a checkpoint (might be foreground process)
		checkpointPath := filepath.Join(homeDir, ".cr-api", "discover", fmt.Sprintf("%s.json", sanitizedTag))
		if _, err := os.Stat(checkpointPath); os.IsNotExist(err) {
			return fmt.Errorf("no active discovery session found for player #%s", playerTag)
		}
		printf("Note: Only checkpoint found. If a foreground discovery is running, use Ctrl+C to stop it.\n")
		printf("Checkpoint will be saved automatically.\n")
		return nil
	}

	// Read PID
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	var pid int
	_, err = fmt.Sscanf(string(pidData), "%d", &pid)
	if err != nil {
		return fmt.Errorf("failed to parse PID: %w", err)
	}

	// Send SIGTERM for graceful shutdown
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	printf("Stopping discovery session (PID: %d)...\n", pid)
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send stop signal: %w", err)
	}

	printf("Stop signal sent. Discovery will save checkpoint and exit gracefully.\n")
	printf("Use 'cr-api deck discover status --tag %s' to verify checkpoint was saved.\n", playerTag)

	return nil
}

// deckDiscoverResumeCommand resumes a discovery session from checkpoint
func deckDiscoverResumeCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")
	verbose := cmd.Bool("verbose")
	background := cmd.Bool("background")

	// Verify checkpoint exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	sanitizedTag := strings.TrimPrefix(playerTag, "#")
	checkpointPath := filepath.Join(homeDir, ".cr-api", "discover", fmt.Sprintf("%s.json", sanitizedTag))

	if _, err := os.Stat(checkpointPath); os.IsNotExist(err) {
		return fmt.Errorf("no checkpoint found for player #%s. Use 'cr-api deck discover start' to begin a new session", playerTag)
	}

	// Read checkpoint to verify it's valid
	data, err := os.ReadFile(checkpointPath)
	if err != nil {
		return fmt.Errorf("failed to read checkpoint: %w", err)
	}

	var checkpoint deck.DiscoveryCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return fmt.Errorf("failed to parse checkpoint: %w", err)
	}

	if verbose {
		fprintf(os.Stderr, "Resuming session from %s\n", checkpoint.Timestamp.Format(time.RFC3339))
		fprintf(os.Stderr, "Previous progress: %d decks evaluated, %d stored\n", checkpoint.Stats.Evaluated, checkpoint.Stats.Stored)
	}

	// Build a synthetic command with resume=true
	if background {
		return runDiscoveryInBackground(ctx, cmd, true)
	}

	// Run in foreground with resume
	return runDiscoveryCommand(ctx, cmd, true)
}

// deckDiscoverStatsCommand shows detailed session statistics
func deckDiscoverStatsCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	sanitizedTag := strings.TrimPrefix(playerTag, "#")

	// Check for checkpoint
	checkpointPath := filepath.Join(homeDir, ".cr-api", "discover", fmt.Sprintf("%s.json", sanitizedTag))

	if _, err := os.Stat(checkpointPath); os.IsNotExist(err) {
		return fmt.Errorf("no discovery session found for player #%s", playerTag)
	}

	// Read checkpoint
	data, err := os.ReadFile(checkpointPath)
	if err != nil {
		return fmt.Errorf("failed to read checkpoint: %w", err)
	}

	var checkpoint deck.DiscoveryCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return fmt.Errorf("failed to parse checkpoint: %w", err)
	}

	// Display detailed statistics
	printf("\n╔════════════════════════════════════════════════════════════════════╗\n")
	printf("║                    DISCOVERY SESSION STATS                         ║\n")
	printf("╚════════════════════════════════════════════════════════════════════╝\n\n")

	printf("Player: #%s\n", playerTag)
	printf("Strategy: %s\n", checkpoint.Strategy)
	printf("Last Updated: %s\n\n", checkpoint.Timestamp.Format("2006-01-02 15:04:05"))

	printf("Progress:\n")
	printf("  Evaluated: %d decks\n", checkpoint.Stats.Evaluated)
	if checkpoint.Stats.Total > 0 {
		printf("  Total: %d decks\n", checkpoint.Stats.Total)
		pct := float64(checkpoint.Stats.Evaluated) / float64(checkpoint.Stats.Total) * 100
		printf("  Complete: %.1f%%\n", pct)
	}
	printf("  Stored: %d decks in leaderboard\n", checkpoint.Stats.Stored)
	printf("\n")

	printf("Performance:\n")
	printf("  Elapsed: %v\n", checkpoint.Stats.Elapsed.Round(time.Second))
	if checkpoint.Stats.Rate > 0 {
		printf("  Rate: %.2f decks/sec\n", checkpoint.Stats.Rate)
	}
	if checkpoint.Stats.ETA > 0 {
		printf("  ETA: %v\n", checkpoint.Stats.ETA.Round(time.Second))
	}
	printf("\n")

	printf("Scores:\n")
	printf("  Average: %.2f\n", checkpoint.Stats.AvgScore)
	printf("  Best: %.2f\n", checkpoint.Stats.BestScore)
	if len(checkpoint.Stats.BestDeck) > 0 {
		printf("  Best Deck: %v\n", checkpoint.Stats.BestDeck)
	}
	if len(checkpoint.Stats.TopScores) > 0 {
		printf("  Top 5 Scores: ")
		for i, score := range checkpoint.Stats.TopScores {
			if i > 0 {
				printf(", ")
			}
			printf("%.2f", score)
		}
		printf("\n")
	}
	printf("\n")

	printf("Actions:\n")
	printf("  Resume: cr-api deck discover resume --tag %s\n", playerTag)
	printf("  View leaderboard: cr-api deck leaderboard show --tag %s\n", playerTag)

	return nil
}

// runDiscoveryInBackground runs the discovery session as a background daemon
func runDiscoveryInBackground(ctx context.Context, cmd *cli.Command, resume bool) error {
	playerTag := cmd.String("tag")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	sanitizedTag := strings.TrimPrefix(playerTag, "#")

	// Check for already running process
	pidFile := filepath.Join(homeDir, ".cr-api", "discover", fmt.Sprintf("%s.pid", sanitizedTag))
	if _, err := os.Stat(pidFile); err == nil {
		pidData, _ := os.ReadFile(pidFile)
		var pid int
		fmt.Sscanf(string(pidData), "%d", &pid)
		process, err := os.FindProcess(pid)
		if err == nil {
			// Try to signal the process to check if it's alive
			if err := process.Signal(syscall.Signal(0)); err == nil {
				return fmt.Errorf("discovery already running in background (PID: %d). Use 'stop' command first", pid)
			}
		}
		// Process is dead, clean up PID file
		os.Remove(pidFile)
	}

	// Ensure PID directory exists
	pidDir := filepath.Dir(pidFile)
	if err := os.MkdirAll(pidDir, 0o755); err != nil {
		return fmt.Errorf("failed to create PID directory: %w", err)
	}

	// Get the executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Build command arguments
	args := []string{"deck", "discover", "run"}
	if resume {
		args = append(args, "--resume")
	}

	// Copy all relevant flags
	flags := []struct {
		name  string
		value string
	}{
		{"tag", cmd.String("tag")},
		{"strategy", cmd.String("strategy")},
		{"sample-size", fmt.Sprintf("%d", cmd.Int("sample-size"))},
	}

	for _, flag := range flags {
		if flag.value != "" {
			args = append(args, fmt.Sprintf("--%s=%s", flag.name, flag.value))
		}
	}

	if cmd.Bool("verbose") {
		args = append(args, "--verbose")
	}

	// Redirect output to log file
	logFile := filepath.Join(homeDir, ".cr-api", "discover", fmt.Sprintf("%s.log", sanitizedTag))
	logHandle, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer logHandle.Close()

	// Set up attributes for background process
	attr := &os.ProcAttr{
		Files: []*os.File{nil, logHandle, logHandle}, // stdin, stdout, stderr
	}

	// Start the process
	process, err := os.StartProcess(execPath, args, attr)
	if err != nil {
		return fmt.Errorf("failed to start background process: %w", err)
	}

	// Write PID file
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", process.Pid)), 0o644); err != nil {
		process.Kill()
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	printf("Discovery started in background (PID: %d)\n", process.Pid)
	printf("Log file: %s\n", logFile)
	printf("\nCommands:\n")
	printf("  Status: cr-api deck discover status --tag %s\n", playerTag)
	printf("  Stop: cr-api deck discover stop --tag %s\n", playerTag)
	printf("  Stats: cr-api deck discover stats --tag %s\n", playerTag)

	return nil
}
