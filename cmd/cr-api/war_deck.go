package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/klauer/clash-royale-api/go/pkg/analysis"
	"github.com/klauer/clash-royale-api/go/pkg/archetypes"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	deckpkg "github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/mulligan"
	"github.com/urfave/cli/v3"
)

type warDeckCandidate struct {
	Archetype mulligan.Archetype
	Deck      *deckpkg.DeckRecommendation
	Score     float64
}

func deckWarCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	dataDir := cmd.String("data-dir")
	deckCount := cmd.Int("deck-count")
	combatStatsWeight := cmd.Float64("combat-stats-weight")
	disableCombatStats := cmd.Bool("disable-combat-stats")

	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	if deckCount < 1 {
		return fmt.Errorf("deck-count must be at least 1")
	}

	// Configure combat stats weight
	if disableCombatStats {
		os.Setenv("COMBAT_STATS_WEIGHT", "0")
		if verbose {
			fmt.Printf("Combat stats disabled (using traditional scoring only)\n")
		}
	} else if combatStatsWeight >= 0 && combatStatsWeight <= 1.0 {
		os.Setenv("COMBAT_STATS_WEIGHT", fmt.Sprintf("%.2f", combatStatsWeight))
		if verbose {
			fmt.Printf("Combat stats weight set to: %.2f\n", combatStatsWeight)
		}
	}

	client := clashroyale.NewClient(apiToken)

	if verbose {
		fmt.Printf("Building war decks for player %s (%d decks)\n", tag, deckCount)
	}

	player, err := client.GetPlayer(tag)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	if verbose {
		fmt.Printf("Player: %s (%s)\n", player.Name, player.Tag)
		fmt.Printf("Analyzing %d cards...\n", len(player.Cards))
	}

	analysisOptions := analysis.DefaultAnalysisOptions()
	cardAnalysis, err := analysis.AnalyzeCardCollection(player, analysisOptions)
	if err != nil {
		return fmt.Errorf("failed to analyze card collection: %w", err)
	}

	deckAnalysis := deckpkg.ConvertAnalysisForDeckBuilding(cardAnalysis)

	builder := archetypes.NewArchetypeBuilder(dataDir)
	if unlockedEvos := cmd.String("unlocked-evolutions"); unlockedEvos != "" {
		builder.SetUnlockedEvolutions(strings.Split(unlockedEvos, ","))
	}
	if slots := cmd.Int("evolution-slots"); slots > 0 {
		builder.SetEvolutionSlotLimit(slots)
	}

	warDecks, err := buildWarDecks(builder, deckAnalysis, deckCount)
	if err != nil {
		return err
	}

	displayWarDecks(player, warDecks)
	return nil
}

func buildWarDecks(
	builder *archetypes.ArchetypeBuilder,
	analysis deckpkg.CardAnalysis,
	deckCount int,
) ([]warDeckCandidate, error) {
	allArchetypes := archetypes.GetAllArchetypes()
	if deckCount > len(allArchetypes) {
		return nil, fmt.Errorf("deck-count must be at most %d", len(allArchetypes))
	}

	bestScore := -1.0
	bestMinScore := -1.0
	var best []warDeckCandidate

	permuteArchetypes(allArchetypes, deckCount, func(order []mulligan.Archetype) {
		used := make(map[string]bool)
		decks := make([]warDeckCandidate, 0, deckCount)
		totalScore := 0.0
		minScore := math.MaxFloat64

		for _, archetype := range order {
			filtered := filterDeckAnalysis(analysis, used)
			recommendation, err := builder.BuildForArchetype(archetype, filtered)
			if err != nil {
				return
			}

			if hasOverlap(recommendation.Deck, used) {
				return
			}

			score := sumDeckScore(recommendation)
			if score < minScore {
				minScore = score
			}
			totalScore += score
			decks = append(decks, warDeckCandidate{
				Archetype: archetype,
				Deck:      recommendation,
				Score:     score,
			})
			markUsedCards(recommendation.Deck, used)
		}

		if len(decks) != deckCount {
			return
		}

		if totalScore > bestScore || (math.Abs(totalScore-bestScore) < 0.0001 && minScore > bestMinScore) {
			bestScore = totalScore
			bestMinScore = minScore
			best = append([]warDeckCandidate(nil), decks...)
		}
	})

	if best == nil {
		return nil, fmt.Errorf("failed to build %d no-repeat decks from available archetypes", deckCount)
	}

	return best, nil
}

func displayWarDecks(player *clashroyale.Player, warDecks []warDeckCandidate) {
	totalScore := 0.0
	uniqueCards := make(map[string]bool)
	for _, deck := range warDecks {
		totalScore += deck.Score
		for _, card := range deck.Deck.Deck {
			uniqueCards[normalizeCardName(card)] = true
		}
	}

	fmt.Printf("\nWAR DECK SET (NO REPEATS)\n")
	fmt.Printf("========================\n\n")
	fmt.Printf("Player: %s (%s)\n", player.Name, player.Tag)
	fmt.Printf("Decks: %d\n", len(warDecks))
	fmt.Printf("Unique cards: %d\n", len(uniqueCards))
	fmt.Printf("Total score: %.3f\n", totalScore)

	if combatWeight := os.Getenv("COMBAT_STATS_WEIGHT"); combatWeight != "" {
		if combatWeight == "0" {
			fmt.Printf("Scoring: Traditional only (combat stats disabled)\n")
		} else {
			fmt.Printf("Scoring: %.0f%% traditional, %.0f%% combat stats\n",
				(1-mustParseFloat(combatWeight))*100,
				mustParseFloat(combatWeight)*100)
		}
	}

	for i, deck := range warDecks {
		fmt.Printf("\nDeck %d - %s\n", i+1, formatArchetypeName(deck.Archetype))
		fmt.Printf("Average Elixir: %.2f\n", deck.Deck.AvgElixir)
		fmt.Printf("Deck score: %.3f\n", deck.Score)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "#\tCard\tLevel\t\tElixir\tRole\n")
		fmt.Fprintf(w, "-\t----\t-----\t\t------\t----\n")

		for j, card := range deck.Deck.DeckDetail {
			evoBadge := deckpkg.FormatEvolutionBadge(card.EvolutionLevel)
			levelStr := fmt.Sprintf("%d/%d", card.Level, card.MaxLevel)
			if evoBadge != "" {
				levelStr = fmt.Sprintf("%s (%s)", levelStr, evoBadge)
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%s\n",
				j+1,
				card.Name,
				levelStr,
				card.Elixir,
				card.Role)
		}
		w.Flush()

		if len(deck.Deck.Notes) > 0 {
			fmt.Printf("\nNotes:\n")
			for _, note := range deck.Deck.Notes {
				fmt.Printf("- %s\n", note)
			}
		}
	}
}

func permuteArchetypes(
	archetypesList []mulligan.Archetype,
	count int,
	fn func([]mulligan.Archetype),
) {
	used := make([]bool, len(archetypesList))
	current := make([]mulligan.Archetype, 0, count)

	var walk func()
	walk = func() {
		if len(current) == count {
			fn(append([]mulligan.Archetype(nil), current...))
			return
		}
		for i, archetype := range archetypesList {
			if used[i] {
				continue
			}
			used[i] = true
			current = append(current, archetype)
			walk()
			current = current[:len(current)-1]
			used[i] = false
		}
	}

	walk()
}

func filterDeckAnalysis(analysis deckpkg.CardAnalysis, excluded map[string]bool) deckpkg.CardAnalysis {
	filtered := deckpkg.CardAnalysis{
		CardLevels:   make(map[string]deckpkg.CardLevelData),
		AnalysisTime: analysis.AnalysisTime,
	}

	for cardName, cardData := range analysis.CardLevels {
		if excluded[normalizeCardName(cardName)] {
			continue
		}
		filtered.CardLevels[cardName] = cardData
	}

	return filtered
}

func sumDeckScore(rec *deckpkg.DeckRecommendation) float64 {
	total := 0.0
	for _, card := range rec.DeckDetail {
		total += card.Score
	}
	return total
}

func hasOverlap(deckCards []string, used map[string]bool) bool {
	for _, card := range deckCards {
		if used[normalizeCardName(card)] {
			return true
		}
	}
	return false
}

func markUsedCards(deckCards []string, used map[string]bool) {
	for _, card := range deckCards {
		used[normalizeCardName(card)] = true
	}
}

func normalizeCardName(cardName string) string {
	return strings.ToLower(strings.TrimSpace(cardName))
}

func formatArchetypeName(archetype mulligan.Archetype) string {
	parts := strings.Split(archetype.String(), "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
