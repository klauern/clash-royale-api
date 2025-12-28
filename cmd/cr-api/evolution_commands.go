package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/storage"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/urfave/cli/v3"
)

// addEvolutionCommands adds evolution-related subcommands to the CLI.
func addEvolutionCommands() *cli.Command {
	return &cli.Command{
		Name:  "evolutions",
		Usage: "Evolution tracking commands",
		Commands: []*cli.Command{
			{
				Name:  "shards",
				Usage: "Track evolution shard inventory",
				Commands: []*cli.Command{
					{
						Name:  "list",
						Usage: "List evolution shard counts",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "card",
								Usage: "Filter by card name",
							},
						},
						Action: evolutionShardsListCommand,
					},
					{
						Name:  "set",
						Usage: "Set evolution shard count for a card",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "card",
								Usage:    "Card name",
								Required: true,
							},
							&cli.IntFlag{
								Name:     "count",
								Usage:    "Shard count",
								Required: true,
							},
						},
						Action: evolutionShardsSetCommand,
					},
				},
			},
			{
				Name:  "recommend",
				Usage: "Recommend optimal evolutions based on shards and card levels",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.IntFlag{
						Name:  "top",
						Value: 10,
						Usage: "Number of recommendations to show",
					},
					&cli.BoolFlag{
						Name:  "verbose",
						Usage: "Show detailed reasons for each recommendation",
					},
					&cli.StringFlag{
						Name:  "unlocked-evolutions",
						Usage: "Comma-separated list of cards with unlocked evolutions (overrides UNLOCKED_EVOLUTIONS env var)",
					},
				},
				Action: evolutionRecommendCommand,
			},
		},
	}
}

func evolutionShardsListCommand(ctx context.Context, cmd *cli.Command) error {
	dataDir := cmd.String("data-dir")
	filter := strings.TrimSpace(cmd.String("card"))

	pathBuilder := storage.NewPathBuilder(dataDir)
	inventory, err := storage.LoadEvolutionShardInventory(pathBuilder.GetEvolutionShardsPath())
	if err != nil {
		return fmt.Errorf("failed to load evolution shard inventory: %w", err)
	}

	if len(inventory.Shards) == 0 {
		fmt.Printf("No evolution shard inventory found.\n")
		return nil
	}

	type row struct {
		name  string
		count int
	}
	rows := make([]row, 0, len(inventory.Shards))
	for name, count := range inventory.Shards {
		if filter != "" && !strings.EqualFold(name, filter) {
			continue
		}
		rows = append(rows, row{name: name, count: count})
	}

	if filter != "" && len(rows) == 0 {
		fmt.Printf("No evolution shard count recorded for %s.\n", filter)
		return nil
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].name < rows[j].name
	})

	total := 0
	for _, entry := range rows {
		total += entry.count
	}

	fmt.Printf("Evolution Shards:\n")
	if !inventory.LastUpdated.IsZero() {
		fmt.Printf("Last updated: %s\n", inventory.LastUpdated.Format(time.RFC3339))
	}
	fmt.Printf("Tracked cards: %d\n", len(rows))
	fmt.Printf("Total shards: %d\n\n", total)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Card\tShards\n")
	fmt.Fprintf(w, "----\t------\n")
	for _, entry := range rows {
		fmt.Fprintf(w, "%s\t%d\n", entry.name, entry.count)
	}
	w.Flush()

	return nil
}

func evolutionShardsSetCommand(ctx context.Context, cmd *cli.Command) error {
	dataDir := cmd.String("data-dir")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	cardInput := cmd.String("card")
	count := cmd.Int("count")

	if count < 0 {
		return fmt.Errorf("count must be 0 or greater")
	}

	cards, err := loadStaticCards(dataDir, apiToken, verbose)
	if err != nil {
		return err
	}

	cardIndex := buildCardNameIndex(cards)
	cardName, err := resolveCardName(cardInput, cardIndex)
	if err != nil {
		return err
	}

	pathBuilder := storage.NewPathBuilder(dataDir)
	inventory, err := storage.LoadEvolutionShardInventory(pathBuilder.GetEvolutionShardsPath())
	if err != nil {
		return fmt.Errorf("failed to load evolution shard inventory: %w", err)
	}

	if inventory.Shards == nil {
		inventory.Shards = make(map[string]int)
	}

	inventory.Shards[cardName] = count
	inventory.LastUpdated = time.Now().UTC()

	if err := storage.SaveEvolutionShardInventory(pathBuilder.GetEvolutionShardsPath(), inventory); err != nil {
		return fmt.Errorf("failed to save evolution shard inventory: %w", err)
	}

	fmt.Printf("Set %s evolution shards to %d.\n", cardName, count)
	return nil
}

func buildCardNameIndex(cards []clashroyale.Card) map[string]string {
	index := make(map[string]string, len(cards))
	for _, card := range cards {
		if card.Name == "" {
			continue
		}
		index[strings.ToLower(card.Name)] = card.Name
	}
	return index
}

func resolveCardName(input string, index map[string]string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("card name is required")
	}
	if cardName, ok := index[strings.ToLower(trimmed)]; ok {
		return cardName, nil
	}
	return "", fmt.Errorf("unknown card name %q (run `cr-api cards` to refresh card data)", trimmed)
}

func cacheStaticCards(dataDir string, cards *clashroyale.CardList) error {
	if cards == nil {
		return fmt.Errorf("card database is nil")
	}

	pathBuilder := storage.NewPathBuilder(dataDir)
	return storage.WriteJSON(pathBuilder.GetStaticCardsPath(), cards)
}

func loadStaticCards(dataDir, apiToken string, verbose bool) ([]clashroyale.Card, error) {
	pathBuilder := storage.NewPathBuilder(dataDir)
	cardsPath := pathBuilder.GetStaticCardsPath()

	if storage.FileExists(cardsPath) {
		var cached clashroyale.CardList
		if err := storage.ReadJSON(cardsPath, &cached); err != nil {
			return nil, fmt.Errorf("failed to read cached card database: %w", err)
		}
		if len(cached.Items) > 0 {
			return cached.Items, nil
		}
	}

	if apiToken == "" {
		return nil, fmt.Errorf("card database not cached; run `cr-api cards` or provide --api-token to fetch")
	}

	if verbose {
		fmt.Printf("Fetching card database for validation...\n")
	}

	client := clashroyale.NewClient(apiToken)
	cards, err := client.GetCards()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch card database: %w", err)
	}

	if err := cacheStaticCards(dataDir, cards); err != nil && verbose {
		fmt.Printf("Warning: Failed to cache card database: %v\n", err)
	}

	return cards.Items, nil
}

func evolutionRecommendCommand(ctx context.Context, cmd *cli.Command) error {
	dataDir := cmd.String("data-dir")
	apiToken := cmd.String("api-token")
	verbose := cmd.Bool("verbose")
	playerTag := cmd.String("tag")
	topN := cmd.Int("top")
	unlockedEvolutionsStr := cmd.String("unlocked-evolutions")

	// Parse unlocked evolutions
	var unlockedEvolutions []string
	if unlockedEvolutionsStr != "" {
		unlockedEvolutions = strings.Split(unlockedEvolutionsStr, ",")
		for i := range unlockedEvolutions {
			unlockedEvolutions[i] = strings.TrimSpace(unlockedEvolutions[i])
		}
	} else {
		// Fallback to env var if not specified
		envVar := cmd.String("unlocked-evolutions") // Note: this gets from env if set
		if envVar != "" {
			unlockedEvolutions = strings.Split(envVar, ",")
			for i := range unlockedEvolutions {
				unlockedEvolutions[i] = strings.TrimSpace(unlockedEvolutions[i])
			}
		}
	}

	// Load player data
	client := clashroyale.NewClient(apiToken)
	if verbose {
		fmt.Printf("Fetching player data for %s...\n", playerTag)
	}

	player, err := client.GetPlayer(playerTag)
	if err != nil {
		return fmt.Errorf("failed to fetch player: %w", err)
	}

	// Load shard inventory
	pathBuilder := storage.NewPathBuilder(dataDir)
	shardInventory, err := storage.LoadEvolutionShardInventory(pathBuilder.GetEvolutionShardsPath())
	if err != nil {
		return fmt.Errorf("failed to load shard inventory: %w", err)
	}

	if verbose {
		fmt.Printf("Loaded shard inventory with %d cards tracked.\n", len(shardInventory.Shards))
	}

	// Load static cards for max evolution levels
	cards, err := loadStaticCards(dataDir, apiToken, verbose)
	if err != nil {
		return err
	}

	// Build max evolution level lookup
	maxEvolutionLevels := make(map[string]int)
	for _, card := range cards {
		if card.MaxEvolutionLevel > 0 {
			maxEvolutionLevels[card.Name] = card.MaxEvolutionLevel
		}
	}

	// Build card candidates from player data
	candidates := make([]deck.CardCandidate, 0, len(player.Cards))
	for _, card := range player.Cards {
		candidate := deck.CardCandidate{
			Name:              card.Name,
			Level:             card.Level,
			MaxLevel:          card.MaxLevel,
			Rarity:            card.Rarity,
			Elixir:            card.ElixirCost,
			EvolutionLevel:    card.EvolutionLevel,
			MaxEvolutionLevel: maxEvolutionLevels[card.Name],
		}
		candidates = append(candidates, candidate)
	}

	// Classify candidates
	deck.ClassifyAllCandidates(candidates)

	// Create recommender and get recommendations
	recommender := deck.NewEvolutionRecommender(shardInventory.Shards, unlockedEvolutions)
	recommendations := recommender.Recommend(candidates, topN)

	// Display results
	fmt.Print(deck.FormatRecommendations(recommendations, verbose))

	return nil
}
