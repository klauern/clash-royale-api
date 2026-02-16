package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/urfave/cli/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func deckAnalyzeCommand(ctx context.Context, cmd *cli.Command) error {
	cardNames := cmd.StringSlice("cards")

	if len(cardNames) != 8 {
		return fmt.Errorf("exactly 8 cards are required for deck analysis")
	}

	return fmt.Errorf("deck analyze is not implemented yet")
}

//nolint:funlen,gocognit,gocyclo // Command flow complexity scheduled for decomposition in clash-royale-api-1g1r.
func deckOptimizeCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	cardNames := cmd.StringSlice("cards")
	apiToken := cmd.String("api-token")
	dataDir := cmd.String("data-dir")
	verbose := cmd.Bool("verbose")
	maxChanges := cmd.Int("max-changes")
	keepWinCondition := cmd.Bool("keep-win-con")
	exportCSV := cmd.Bool("export-csv")

	if maxChanges > 0 {
		fprintf(os.Stderr, "Warning: --max-changes is not implemented yet and will be ignored (got %d)\n", maxChanges)
	}
	if keepWinCondition {
		fprintf(os.Stderr, "Warning: --keep-win-con is not implemented yet and will be ignored\n")
	}

	if apiToken == "" {
		return fmt.Errorf("API token is required")
	}

	client := clashroyale.NewClient(apiToken)
	var player *clashroyale.Player
	var err error

	// If no cards provided, fetch player's current deck from API
	if len(cardNames) == 0 {
		if verbose {
			printf("Fetching player data for tag: %s\n", tag)
		}

		player, err = client.GetPlayerWithContext(ctx, tag)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}

		if len(player.CurrentDeck) == 0 {
			return fmt.Errorf("player %s has no current deck configured", tag)
		}

		if len(player.CurrentDeck) != 8 {
			return fmt.Errorf("player's current deck has %d cards, expected 8", len(player.CurrentDeck))
		}

		// Extract card names from CurrentDeck
		cardNames = make([]string, len(player.CurrentDeck))
		for i, card := range player.CurrentDeck {
			cardNames[i] = card.Name
		}

		if verbose {
			printf("Using player's current deck: %v\n", cardNames)
		}
	} else if len(cardNames) != 8 {
		return fmt.Errorf("exactly 8 cards are required for optimization")
	}

	// Fetch player data for context
	if player == nil {
		player, err = client.GetPlayerWithContext(ctx, tag)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}
	}

	// Convert card names to CardCandidates
	deckCards := convertToCardCandidates(cardNames)

	// Load synergy database
	synergyDB := deck.NewSynergyDatabase()

	// Create player context
	playerContext := evaluation.NewPlayerContextFromPlayer(player)

	// Evaluate current deck
	if verbose {
		fmt.Println("Evaluating current deck...")
	}
	currentResult := evaluation.Evaluate(deckCards, synergyDB, playerContext)

	// Convert player cards to map for GenerateAlternatives
	playerCardMap := make(map[string]bool)
	for _, card := range player.Cards {
		playerCardMap[card.Name] = true
	}

	// Generate alternative suggestions
	if verbose {
		fmt.Println("Generating optimization suggestions...")
	}
	alternatives := evaluation.GenerateAlternatives(deckCards, synergyDB, 10, playerCardMap)

	// Display current deck analysis
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    DECK OPTIMIZATION REPORT                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	printf("🃏 Current Deck: %s\n", strings.Join(cardNames, " • "))
	printf("📊 Average Elixir: %.2f\n", currentResult.AvgElixir)
	printf("🎯 Archetype: %s (%.0f%% confidence)\n",
		cases.Title(language.English).String(string(currentResult.DetectedArchetype)),
		currentResult.ArchetypeConfidence*100)
	fmt.Println()
	printf("⭐ Current Overall Score: %.1f/10 - %s\n",
		currentResult.OverallScore,
		currentResult.OverallRating)
	fmt.Println()

	// Display current category scores
	fmt.Println("Current Category Scores:")
	printf("  ⚔️  Attack:        %s  %.1f/10 - %s\n",
		formatStars(currentResult.Attack.Stars),
		currentResult.Attack.Score,
		currentResult.Attack.Rating)
	printf("  🛡️  Defense:       %s  %.1f/10 - %s\n",
		formatStars(currentResult.Defense.Stars),
		currentResult.Defense.Score,
		currentResult.Defense.Rating)
	printf("  🔗 Synergy:       %s  %.1f/10 - %s\n",
		formatStars(currentResult.Synergy.Stars),
		currentResult.Synergy.Score,
		currentResult.Synergy.Rating)
	printf("  🎭 Versatility:   %s  %.1f/10 - %s\n",
		formatStars(currentResult.Versatility.Stars),
		currentResult.Versatility.Score,
		currentResult.Versatility.Rating)
	printf("  💰 F2P Friendly:  %s  %.1f/10 - %s\n",
		formatStars(currentResult.F2PFriendly.Stars),
		currentResult.F2PFriendly.Score,
		currentResult.F2PFriendly.Rating)
	fmt.Println()

	// Display optimization suggestions
	if len(alternatives.Suggestions) == 0 {
		fmt.Println("✨ Your deck is already well-optimized! No better alternatives found.")
		return nil
	}

	fmt.Println("═══════════════════════════════════════════════════════════════")
	printf("                OPTIMIZATION SUGGESTIONS (%d found)\n", len(alternatives.Suggestions))
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Display top suggestions
	displayCount := min(len(alternatives.Suggestions), 5)

	for i, alt := range alternatives.Suggestions[:displayCount] {
		printf("Suggestion #%d: %s\n", i+1, alt.Impact)
		fmt.Println("───────────────────────────────────────────────────────────────")
		printf("  Replace: %s  →  %s\n", alt.OriginalCard, alt.ReplacementCard)
		printf("  Score Improvement: %.1f → %.1f (+%.1f)\n",
			alt.OriginalScore, alt.NewScore, alt.ScoreDelta)
		printf("  Rationale: %s\n", alt.Rationale)
		printf("  New Deck: %s\n", strings.Join(alt.Deck, " • "))
		fmt.Println()
	}

	// CSV export if requested
	if exportCSV {
		safeTag := sanitizePathComponent(tag)
		csvPath := filepath.Join(dataDir, fmt.Sprintf("deck-optimize-%s-%d.csv", safeTag, time.Now().Unix()))
		if err := exportOptimizationCSV(csvPath, tag, cardNames, currentResult, *alternatives); err != nil {
			fprintf(os.Stderr, "Warning: Failed to export CSV: %v\n", err)
		} else {
			printf("✓ Optimization results exported to: %s\n", csvPath)
		}
	}

	return nil
}
