package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/klauer/clash-royale-api/go/internal/playertag"
	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/evaluation"
	"github.com/klauer/clash-royale-api/go/pkg/fuzzstorage"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v3"
)

// deckFuzzListCommand lists saved top decks from storage.
//
//nolint:gocognit,gocyclo,funlen // Command handlers naturally aggregate CLI parsing and orchestration.
func deckFuzzListCommand(ctx context.Context, cmd *cli.Command) error {
	top := cmd.Int("top")
	archetype := cmd.String("archetype")
	minScore := cmd.Float64("min-score")
	maxScore := cmd.Float64("max-score")
	minElixir := cmd.Float64("min-elixir")
	maxElixir := cmd.Float64("max-elixir")
	maxSameArchetype := cmd.Int("max-same-archetype")
	format := cmd.String("format")
	playerTag := cmd.String("tag")
	workers := cmd.Int("workers")
	verbose := cmd.Bool("verbose")

	if workers == 1 && playerTag != "" {
		workers = runtime.NumCPU()
		if verbose {
			fprintf(os.Stderr, "Auto-detected %d CPU cores, using %d workers\n", runtime.NumCPU(), workers)
		}
	}

	storage, err := fuzzstorage.NewStorage("")
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	// Build query options
	queryOpts := fuzzstorage.QueryOptions{
		Limit: top,
	}

	if archetype != "" {
		queryOpts.Archetype = archetype
	}
	if minScore > 0 {
		queryOpts.MinScore = minScore
	}
	if maxScore > 0 {
		queryOpts.MaxScore = maxScore
	}
	if minElixir > 0 {
		queryOpts.MinAvgElixir = minElixir
	}
	if maxElixir > 0 {
		queryOpts.MaxAvgElixir = maxElixir
	}

	// Query decks
	decks, err := storage.Query(queryOpts)
	if err != nil {
		return fmt.Errorf("failed to query decks: %w", err)
	}
	histogram, err := storage.ArchetypeHistogram(queryOpts)
	if err != nil {
		return fmt.Errorf("failed to query archetype histogram: %w", err)
	}

	var theoreticalByID map[int]fuzzstorage.DeckEntry
	if playerTag != "" && len(decks) > 0 {
		apiToken := cmd.String("api-token")
		if apiToken == "" {
			apiToken = os.Getenv("CLASH_ROYALE_API_TOKEN")
		}
		if apiToken == "" {
			return fmt.Errorf("API token is required to load player context (set CLASH_ROYALE_API_TOKEN or use --api-token)")
		}

		client := clashroyale.NewClient(apiToken)
		cleanTag, err := playertag.Sanitize(playerTag)
		if err != nil {
			return fmt.Errorf("invalid player tag %q: %w", playerTag, err)
		}
		player, playerErr := client.GetPlayerWithContext(ctx, cleanTag)
		if playerErr != nil {
			return fmt.Errorf("failed to load player data for %s: %w", playerTag, playerErr)
		}
		playerContext := evaluation.NewPlayerContextFromPlayer(player)

		theoreticalByID = make(map[int]fuzzstorage.DeckEntry, len(decks))
		for _, deck := range decks {
			theoreticalByID[deck.ID] = deck
		}

		decks = reevaluateStoredDecks(decks, player, player.Tag, playerContext, workers, verbose)
		sort.Slice(decks, func(i, j int) bool {
			return decks[i].OverallScore > decks[j].OverallScore
		})

		if verbose {
			printf("Loaded player context for %s (%s)\n", player.Name, player.Tag)
		}
	}
	if maxSameArchetype > 0 {
		decks = limitArchetypeRepetition(decks, maxSameArchetype)
	}

	total, err := storage.Count()
	if err != nil {
		return fmt.Errorf("failed to count decks: %w", err)
	}
	dbPath := storage.GetDBPath()

	fprintf(os.Stderr, "Top decks from: %s\n", dbPath)
	fprintf(os.Stderr, "Showing %d of %d total decks\n\n", len(decks), total)

	// Format output
	switch format {
	case fuzzOutputJSON:
		return formatListResultsJSON(decks, dbPath, total, histogram, theoreticalByID)
	case fuzzOutputCSV:
		return formatListResultsCSV(decks, theoreticalByID)
	case fuzzOutputDetailed:
		return formatListResultsDetailed(decks, dbPath, total, histogram, theoreticalByID)
	default:
		return formatListResultsSummary(decks, dbPath, total, histogram, theoreticalByID)
	}
}

// deckFuzzUpdateCommand re-evaluates saved decks with current scoring and updates storage.
//
//nolint:gocognit,gocyclo,funlen // Command handlers naturally aggregate CLI parsing and orchestration.
func deckFuzzUpdateCommand(ctx context.Context, cmd *cli.Command) error {
	playerTag := cmd.String("tag")
	top := cmd.Int("top")
	archetype := cmd.String("archetype")
	minScore := cmd.Float64("min-score")
	maxScore := cmd.Float64("max-score")
	minElixir := cmd.Float64("min-elixir")
	maxElixir := cmd.Float64("max-elixir")
	workers := cmd.Int("workers")
	verbose := cmd.Bool("verbose")

	if workers == 1 {
		workers = runtime.NumCPU()
		if verbose {
			fprintf(os.Stderr, "Auto-detected %d CPU cores, using %d workers\n", runtime.NumCPU(), workers)
		}
	}

	storage, err := fuzzstorage.NewStorage("")
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer closeFile(storage)

	queryOpts := fuzzstorage.QueryOptions{
		Limit: top,
	}
	if archetype != "" {
		queryOpts.Archetype = archetype
	}
	if minScore > 0 {
		queryOpts.MinScore = minScore
	}
	if maxScore > 0 {
		queryOpts.MaxScore = maxScore
	}
	if minElixir > 0 {
		queryOpts.MinAvgElixir = minElixir
	}
	if maxElixir > 0 {
		queryOpts.MaxAvgElixir = maxElixir
	}

	entries, err := storage.Query(queryOpts)
	if err != nil {
		return fmt.Errorf("failed to query decks: %w", err)
	}
	if len(entries) == 0 {
		fmt.Println("No decks found for update.")
		return nil
	}

	var player *clashroyale.Player
	var playerContext *evaluation.PlayerContext
	if playerTag != "" {
		apiToken := cmd.String("api-token")
		if apiToken == "" {
			apiToken = os.Getenv("CLASH_ROYALE_API_TOKEN")
		}
		if apiToken == "" {
			return fmt.Errorf("API token is required to load player context (set CLASH_ROYALE_API_TOKEN or use --api-token)")
		}
		client := clashroyale.NewClient(apiToken)
		var err error
		player, err = client.GetPlayerWithContext(ctx, playerTag)
		if err != nil {
			return fmt.Errorf("failed to load player data for %s: %w", playerTag, err)
		}
		playerContext = evaluation.NewPlayerContextFromPlayer(player)
		if verbose {
			printf("Loaded player context for %s (%s)\n", player.Name, playerTag)
		}
	}

	start := time.Now()
	updatedEntries := reevaluateStoredDecks(entries, player, playerTag, playerContext, workers, verbose)

	updated := 0
	for i := range updatedEntries {
		if err := storage.UpdateDeck(&updatedEntries[i]); err != nil {
			return fmt.Errorf("failed to update deck %d: %w", updatedEntries[i].ID, err)
		}
		updated++
	}

	if verbose {
		fprintf(os.Stderr, "Updated %d decks in %v\n", updated, time.Since(start).Round(time.Millisecond))
		fprintf(os.Stderr, "Database: %s\n", storage.GetDBPath())
	}

	printf("Updated %d saved decks\n", updated)
	return nil
}

type storedDeckWork struct {
	index int
	entry fuzzstorage.DeckEntry
}

type storedDeckResult struct {
	index int
	entry fuzzstorage.DeckEntry
}

func formatScoreTransition(
	theoreticalByID map[int]fuzzstorage.DeckEntry,
	deckID int,
	current float64,
	extract func(fuzzstorage.DeckEntry) float64,
) string {
	if theoreticalByID == nil {
		return fmt.Sprintf("%.2f", current)
	}
	theoretical, ok := theoreticalByID[deckID]
	if !ok {
		return fmt.Sprintf("%.2f", current)
	}
	return fmt.Sprintf("%.2f->%.2f", extract(theoretical), current)
}

//nolint:funlen // Worker orchestration is clearer in one block with channels and progress updates.
func reevaluateStoredDecks(entries []fuzzstorage.DeckEntry, player *clashroyale.Player, playerTag string, playerContext *evaluation.PlayerContext, workers int, verbose bool) []fuzzstorage.DeckEntry {
	if workers <= 1 {
		return reevaluateStoredDecksSequential(entries, player, playerTag, playerContext, verbose)
	}

	results := make([]fuzzstorage.DeckEntry, len(entries))
	workChan := make(chan storedDeckWork, len(entries))
	resultChan := make(chan storedDeckResult, len(entries))
	var wg sync.WaitGroup

	var bar *progressbar.ProgressBar
	if verbose {
		bar = progressbar.NewOptions(len(entries),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("decks"),
			progressbar.OptionOnCompletion(func() {
				fprintln(os.Stderr)
			}),
		)
	}

	for range workers {
		wg.Go(func() {
			synergyDB := deck.NewSynergyDatabase()

			for work := range workChan {
				result := evaluateSingleDeck(work.entry.Cards, player, playerTag, synergyDB, playerContext)
				updated := work.entry
				updated.OverallScore = result.OverallScore
				updated.AttackScore = result.AttackScore
				updated.DefenseScore = result.DefenseScore
				updated.SynergyScore = result.SynergyScore
				updated.VersatilityScore = result.VersatilityScore
				updated.AvgElixir = result.AvgElixir
				updated.Archetype = result.Archetype
				updated.ArchetypeConf = result.ArchetypeConfidence
				updated.EvaluatedAt = result.EvaluatedAt
				resultChan <- storedDeckResult{index: work.index, entry: updated}
			}
		})
	}

	for i, entry := range entries {
		workChan <- storedDeckWork{index: i, entry: entry}
	}
	close(workChan)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		results[result.index] = result.entry
		if verbose && bar != nil {
			if err := bar.Add(1); err != nil {
				fprintf(os.Stderr, "Warning: progress update failed: %v\n", err)
			}
		}
	}

	return results
}

func reevaluateStoredDecksSequential(entries []fuzzstorage.DeckEntry, player *clashroyale.Player, playerTag string, playerContext *evaluation.PlayerContext, verbose bool) []fuzzstorage.DeckEntry {
	results := make([]fuzzstorage.DeckEntry, len(entries))
	synergyDB := deck.NewSynergyDatabase()

	var bar *progressbar.ProgressBar
	if verbose {
		bar = progressbar.NewOptions(len(entries),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("decks"),
			progressbar.OptionOnCompletion(func() {
				fprintln(os.Stderr)
			}),
		)
	}

	for i, entry := range entries {
		result := evaluateSingleDeck(entry.Cards, player, playerTag, synergyDB, playerContext)
		entry.OverallScore = result.OverallScore
		entry.AttackScore = result.AttackScore
		entry.DefenseScore = result.DefenseScore
		entry.SynergyScore = result.SynergyScore
		entry.VersatilityScore = result.VersatilityScore
		entry.AvgElixir = result.AvgElixir
		entry.Archetype = result.Archetype
		entry.ArchetypeConf = result.ArchetypeConfidence
		entry.EvaluatedAt = result.EvaluatedAt
		results[i] = entry

		if verbose && bar != nil {
			if err := bar.Add(1); err != nil {
				fprintf(os.Stderr, "Warning: progress update failed: %v\n", err)
			}
		}
	}

	return results
}

// formatListResultsSummary formats list results in summary format
func formatListResultsSummary(
	decks []fuzzstorage.DeckEntry,
	dbPath string,
	total int,
	histogram map[string]int,
	theoreticalByID map[int]fuzzstorage.DeckEntry,
) error {
	printf("Saved Top Decks\n")
	printf("Database: %s\n", dbPath)
	printf("Total decks: %d\n\n", total)

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, ' ', 0)
	fprintln(w, "Rank\tDeck\tOverall\tAttack\tDefense\tSynergy\tElixir\tArchetype")

	for i, deck := range decks {
		deckStr := strings.Join(deck.Cards, ", ")
		overall := formatScoreTransition(theoreticalByID, deck.ID, deck.OverallScore, func(entry fuzzstorage.DeckEntry) float64 { return entry.OverallScore })
		attack := formatScoreTransition(theoreticalByID, deck.ID, deck.AttackScore, func(entry fuzzstorage.DeckEntry) float64 { return entry.AttackScore })
		defense := formatScoreTransition(theoreticalByID, deck.ID, deck.DefenseScore, func(entry fuzzstorage.DeckEntry) float64 { return entry.DefenseScore })
		synergy := formatScoreTransition(theoreticalByID, deck.ID, deck.SynergyScore, func(entry fuzzstorage.DeckEntry) float64 { return entry.SynergyScore })
		if len(deckStr) > 50 {
			firstLine := strings.Join(deck.Cards[:4], ", ")
			fprintf(w, "%d\t%s,\t%s\t%s\t%s\t%s\t%.2f\t%s\n",
				i+1, firstLine, overall, attack, defense, synergy, deck.AvgElixir, deck.Archetype)
			secondLine := strings.Join(deck.Cards[4:], ", ")
			fprintf(w, "\t%s\n", secondLine)
		} else {
			fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%.2f\t%s\n",
				i+1, deckStr, overall, attack, defense, synergy, deck.AvgElixir, deck.Archetype)
		}
	}

	flushWriter(w)

	if len(histogram) > 0 {
		printf("\nArchetype Histogram (matching query):\n")
		printArchetypeHistogram(histogram)
	}
	return nil
}

// formatListResultsJSON formats list results in JSON format
func formatListResultsJSON(
	decks []fuzzstorage.DeckEntry,
	dbPath string,
	total int,
	histogram map[string]int,
	theoreticalByID map[int]fuzzstorage.DeckEntry,
) error {
	results := make([]map[string]any, 0, len(decks))
	for _, deck := range decks {
		result := map[string]any{
			"id":                deck.ID,
			"cards":             deck.Cards,
			"overall_score":     deck.OverallScore,
			"attack_score":      deck.AttackScore,
			"defense_score":     deck.DefenseScore,
			"synergy_score":     deck.SynergyScore,
			"versatility_score": deck.VersatilityScore,
			"avg_elixir":        deck.AvgElixir,
			"archetype":         deck.Archetype,
			"archetype_conf":    deck.ArchetypeConf,
			"evaluated_at":      deck.EvaluatedAt,
		}
		if theoreticalByID != nil {
			if theoretical, ok := theoreticalByID[deck.ID]; ok {
				result["stored_overall_score"] = theoretical.OverallScore
				result["stored_attack_score"] = theoretical.AttackScore
				result["stored_defense_score"] = theoretical.DefenseScore
				result["stored_synergy_score"] = theoretical.SynergyScore
			}
		}
		results = append(results, result)
	}

	output := map[string]any{
		"database":            dbPath,
		"total":               total,
		"returned":            len(decks),
		"results":             results,
		"archetype_histogram": histogram,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// formatListResultsCSV formats list results in CSV format
func formatListResultsCSV(decks []fuzzstorage.DeckEntry, theoreticalByID map[int]fuzzstorage.DeckEntry) error {
	header := []string{"Rank", "Deck", "Overall", "Attack", "Defense", "Synergy", "Versatility", "AvgElixir", "Archetype"}
	if theoreticalByID != nil {
		header = []string{
			"Rank", "Deck",
			"StoredOverall", "PlayerOverall",
			"StoredAttack", "PlayerAttack",
			"StoredDefense", "PlayerDefense",
			"StoredSynergy", "PlayerSynergy",
			"Versatility", "AvgElixir", "Archetype",
		}
	}
	rows := make([][]string, 0, len(decks))
	for i, deck := range decks {
		deckStr := strings.Join(deck.Cards, ", ")
		row := []string{
			strconv.Itoa(i + 1),
			deckStr,
		}
		if theoreticalByID != nil {
			theoretical := theoreticalByID[deck.ID]
			row = append(row,
				fmt.Sprintf("%.2f", theoretical.OverallScore),
				fmt.Sprintf("%.2f", deck.OverallScore),
				fmt.Sprintf("%.2f", theoretical.AttackScore),
				fmt.Sprintf("%.2f", deck.AttackScore),
				fmt.Sprintf("%.2f", theoretical.DefenseScore),
				fmt.Sprintf("%.2f", deck.DefenseScore),
				fmt.Sprintf("%.2f", theoretical.SynergyScore),
				fmt.Sprintf("%.2f", deck.SynergyScore),
				fmt.Sprintf("%.2f", deck.VersatilityScore),
				fmt.Sprintf("%.2f", deck.AvgElixir),
				deck.Archetype,
			)
		} else {
			row = append(row,
				fmt.Sprintf("%.2f", deck.OverallScore),
				fmt.Sprintf("%.2f", deck.AttackScore),
				fmt.Sprintf("%.2f", deck.DefenseScore),
				fmt.Sprintf("%.2f", deck.SynergyScore),
				fmt.Sprintf("%.2f", deck.VersatilityScore),
				fmt.Sprintf("%.2f", deck.AvgElixir),
				deck.Archetype,
			)
		}
		rows = append(rows, row)
	}
	return writeCSVDocument(os.Stdout, header, rows)
}

// formatListResultsDetailed formats list results in detailed format
func formatListResultsDetailed(
	decks []fuzzstorage.DeckEntry,
	dbPath string,
	total int,
	histogram map[string]int,
	theoreticalByID map[int]fuzzstorage.DeckEntry,
) error {
	printf("Saved Top Decks\n")
	printf("Database: %s\n", dbPath)
	printf("Total decks: %d\n\n", total)

	for i, deck := range decks {
		printf("=== Deck %d ===\n", i+1)
		printf("Cards: %s\n", strings.Join(deck.Cards, ", "))
		if theoreticalByID != nil {
			if theoretical, ok := theoreticalByID[deck.ID]; ok {
				printf("Overall: %.2f -> %.2f | Attack: %.2f -> %.2f | Defense: %.2f -> %.2f | Synergy: %.2f -> %.2f | Versatility: %.2f\n",
					theoretical.OverallScore, deck.OverallScore,
					theoretical.AttackScore, deck.AttackScore,
					theoretical.DefenseScore, deck.DefenseScore,
					theoretical.SynergyScore, deck.SynergyScore,
					deck.VersatilityScore,
				)
			} else {
				printf("Overall: %.2f | Attack: %.2f | Defense: %.2f | Synergy: %.2f | Versatility: %.2f\n",
					deck.OverallScore, deck.AttackScore, deck.DefenseScore, deck.SynergyScore, deck.VersatilityScore)
			}
		} else {
			printf("Overall: %.2f | Attack: %.2f | Defense: %.2f | Synergy: %.2f | Versatility: %.2f\n",
				deck.OverallScore, deck.AttackScore, deck.DefenseScore, deck.SynergyScore, deck.VersatilityScore)
		}
		printf("Avg Elixir: %.2f | Archetype: %s (%.0f%% confidence)\n",
			deck.AvgElixir, deck.Archetype, deck.ArchetypeConf*100)
		printf("Evaluated: %s\n\n", deck.EvaluatedAt.Format(time.RFC3339))
	}

	if len(histogram) > 0 {
		printf("Archetype Histogram (matching query):\n")
		printArchetypeHistogram(histogram)
	}

	return nil
}

func printArchetypeHistogram(histogram map[string]int) {
	type entry struct {
		archetype string
		count     int
	}

	entries := make([]entry, 0, len(histogram))
	for archetype, count := range histogram {
		entries = append(entries, entry{archetype: archetype, count: count})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count == entries[j].count {
			return entries[i].archetype < entries[j].archetype
		}
		return entries[i].count > entries[j].count
	})

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, ' ', 0)
	fprintln(w, "Archetype\tCount")
	fprintln(w, "---------\t-----")
	for _, e := range entries {
		fprintf(w, "%s\t%d\n", e.archetype, e.count)
	}
	flushWriter(w)
}

func limitArchetypeRepetition(decks []fuzzstorage.DeckEntry, maxPerArchetype int) []fuzzstorage.DeckEntry {
	if maxPerArchetype <= 0 {
		return decks
	}

	counts := make(map[string]int, len(decks))
	filtered := make([]fuzzstorage.DeckEntry, 0, len(decks))
	for _, deck := range decks {
		if counts[deck.Archetype] >= maxPerArchetype {
			continue
		}
		counts[deck.Archetype]++
		filtered = append(filtered, deck)
	}
	return filtered
}
