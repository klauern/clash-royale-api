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
				Name:  "run",
				Usage: "Run deck discovery evaluation session",
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
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show detailed progress",
					},
					&cli.BoolFlag{
						Name:  "resume",
						Usage: "Resume from last checkpoint if available",
					},
				},
				Action: deckDiscoverRunCommand,
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
		},
	}
}

func deckDiscoverRunCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")
	strategy := deck.GeneratorStrategy(cmd.String("strategy"))
	sampleSize := cmd.Int("sample-size")
	verbose := cmd.Bool("verbose")
	resume := cmd.Bool("resume")

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

	err = runner.Run(runCtx)
	canceler.Clear()

	// Handle result
	if err != nil {
		if errors.Is(err, context.Canceled) {
			// Graceful shutdown - checkpoint already saved
			fprintf(os.Stderr, "\nDiscovery stopped. Checkpoint saved. Use --resume to continue.\n")
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
	printf("\nView results with: cr-api deck leaderboard query --tag %s\n", playerTag)

	// Clear checkpoint on successful completion
	runner.ClearCheckpoint()

	return nil
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
	printf("\nResume with: cr-api deck discover run --tag %s --resume\n", playerTag)

	return nil
}
