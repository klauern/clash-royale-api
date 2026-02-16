package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/urfave/cli/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	optimizeFocusBalanced = "balanced"
	optimizeDefaultTag    = "deck"
)

func deckAnalyzeCommand(ctx context.Context, cmd *cli.Command) error {
	_ = ctx
	deckString := cmd.String("deck")
	format := strings.ToLower(strings.TrimSpace(cmd.String("format")))
	cardNames := parseDeckString(deckString)

	if len(cardNames) != 8 {
		return fmt.Errorf("exactly 8 cards are required for deck analysis")
	}

	deckCards := convertToCardCandidates(cardNames)
	result := evaluation.Evaluate(deckCards, deck.NewSynergyDatabase(), nil)

	switch format {
	case "", batchFormatHuman:
		fmt.Print(evaluation.FormatHuman(&result))
	case batchFormatJSON:
		jsonOutput, err := evaluation.FormatJSON(&result)
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Print(jsonOutput)
	default:
		return fmt.Errorf("unknown format: %s (supported: human, json)", format)
	}

	return nil
}

//nolint:funlen,gocognit,gocyclo // Command flow complexity scheduled for decomposition in clash-royale-api-1g1r.
func deckOptimizeCommand(ctx context.Context, cmd *cli.Command) error {
	deckString := cmd.String("deck")
	tag := cmd.String("tag")
	cardNames := parseDeckString(deckString)
	apiToken := cmd.String("api-token")
	dataDir := cmd.String("data-dir")
	verbose := cmd.Bool("verbose")
	suggestions := cmd.Int("suggestions")
	focus := strings.ToLower(strings.TrimSpace(cmd.String("focus")))
	exportCSV := cmd.Bool("export-csv")

	if len(cardNames) != 8 {
		return fmt.Errorf("exactly 8 cards are required for optimization")
	}

	if suggestions <= 0 {
		return fmt.Errorf("--suggestions must be >= 1")
	}

	if focus == "" {
		focus = optimizeFocusBalanced
	}
	if focus != optimizeFocusBalanced && focus != batchSortAttack && focus != batchSortDefense && focus != batchSortSynergy {
		return fmt.Errorf("invalid --focus value %q (supported: balanced, attack, defense, synergy)", focus)
	}

	var player *clashroyale.Player
	var playerContext *evaluation.PlayerContext
	var playerCardMap map[string]bool
	if tag != "" && apiToken != "" {
		client := clashroyale.NewClient(apiToken)
		if verbose {
			printf("Fetching player context for tag: %s\n", tag)
		}
		loadedPlayer, err := client.GetPlayerWithContext(ctx, tag)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}
		player = loadedPlayer
		playerContext = evaluation.NewPlayerContextFromPlayer(player)
		playerCardMap = make(map[string]bool, len(player.Cards))
		for _, card := range player.Cards {
			playerCardMap[card.Name] = true
		}
	} else if tag != "" && apiToken == "" && verbose {
		fprintf(os.Stderr, "Warning: --tag provided without API token; proceeding without collection-aware filtering\n")
	}

	// Convert card names to CardCandidates
	deckCards := convertToCardCandidates(cardNames)

	// Load synergy database
	synergyDB := deck.NewSynergyDatabase()

	// Evaluate current deck
	if verbose {
		fmt.Println("Evaluating current deck...")
	}
	currentResult := evaluation.Evaluate(deckCards, synergyDB, playerContext)

	// Generate alternative suggestions
	if verbose {
		fmt.Println("Generating optimization suggestions...")
	}
	alternatives := evaluation.GenerateAlternatives(deckCards, synergyDB, suggestions*3, playerCardMap)
	alternatives.Suggestions = prioritizeOptimizeSuggestions(alternatives.Suggestions, synergyDB, focus)
	if len(alternatives.Suggestions) > suggestions {
		alternatives.Suggestions = alternatives.Suggestions[:suggestions]
	}

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
	printf("                OPTIMIZATION SUGGESTIONS (%d found, focus=%s)\n", len(alternatives.Suggestions), focus)
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	for i, alt := range alternatives.Suggestions {
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
		csvTag := tag
		if csvTag == "" {
			csvTag = optimizeDefaultTag
		}
		safeTag := sanitizePathComponent(csvTag)
		csvPath := filepath.Join(dataDir, fmt.Sprintf("deck-optimize-%s-%d.csv", safeTag, time.Now().Unix()))
		if err := exportOptimizationCSV(csvPath, tag, cardNames, currentResult, *alternatives); err != nil {
			fprintf(os.Stderr, "Warning: Failed to export CSV: %v\n", err)
		} else {
			printf("✓ Optimization results exported to: %s\n", csvPath)
		}
	}

	return nil
}

func prioritizeOptimizeSuggestions(
	suggestions []evaluation.AlternativeDeck,
	synergyDB *deck.SynergyDatabase,
	focus string,
) []evaluation.AlternativeDeck {
	if len(suggestions) < 2 || focus == optimizeFocusBalanced {
		return suggestions
	}

	type focusedSuggestion struct {
		suggestion evaluation.AlternativeDeck
		focusScore float64
	}

	ranked := make([]focusedSuggestion, 0, len(suggestions))
	for _, suggestion := range suggestions {
		result := evaluation.Evaluate(convertToCardCandidates(suggestion.Deck), synergyDB, nil)
		score := result.OverallScore
		switch focus {
		case batchSortAttack:
			score = result.Attack.Score
		case batchSortDefense:
			score = result.Defense.Score
		case batchSortSynergy:
			score = result.Synergy.Score
		}
		ranked = append(ranked, focusedSuggestion{
			suggestion: suggestion,
			focusScore: score,
		})
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].focusScore == ranked[j].focusScore {
			return ranked[i].suggestion.ScoreDelta > ranked[j].suggestion.ScoreDelta
		}
		return ranked[i].focusScore > ranked[j].focusScore
	})

	sortedSuggestions := make([]evaluation.AlternativeDeck, 0, len(ranked))
	for _, rankedSuggestion := range ranked {
		sortedSuggestions = append(sortedSuggestions, rankedSuggestion.suggestion)
	}
	return sortedSuggestions
}
