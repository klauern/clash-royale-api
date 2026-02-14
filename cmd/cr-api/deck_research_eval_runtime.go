package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/klauer/clash-royale-api/go/pkg/deck/research"
	"github.com/urfave/cli/v3"
)

var defaultResearchTags = []string{"R8QGUQRCV", "2P0GYQJ", "8VCGL8CG", "9Y9RRPQ", "LYR0U0Q"}

func parseMethodsList(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return []string{research.MethodBaseline, research.MethodGenetic, research.MethodConstraint, research.MethodRoleFirst}, nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]bool)
	valid := map[string]bool{
		research.MethodBaseline:   true,
		research.MethodGenetic:    true,
		research.MethodConstraint: true,
		research.MethodRoleFirst:  true,
	}
	for _, p := range parts {
		m := strings.ToLower(strings.TrimSpace(p))
		if m == "" || seen[m] {
			continue
		}
		if !valid[m] {
			return nil, fmt.Errorf("unsupported method %q (valid: baseline, genetic, constraint, role-first)", m)
		}
		seen[m] = true
		out = append(out, m)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid methods provided")
	}
	return out, nil
}

func parseTags(input []string) []string {
	if len(input) == 0 {
		return append([]string(nil), defaultResearchTags...)
	}
	out := make([]string, 0, len(input))
	seen := make(map[string]bool)
	for _, t := range input {
		clean := strings.TrimSpace(strings.TrimPrefix(t, "#"))
		if clean == "" || seen[clean] {
			continue
		}
		seen[clean] = true
		out = append(out, clean)
	}
	return out
}

//nolint:gocyclo,funlen // Command wiring intentionally keeps validation/fetch/run/output in one flow.
func deckResearchEvalCommand(_ context.Context, cmd *cli.Command) error {
	tags := parseTags(cmd.StringSlice("tags"))
	if len(tags) == 0 {
		return fmt.Errorf("at least one valid tag is required")
	}

	methods, err := parseMethodsList(cmd.String("methods"))
	if err != nil {
		return err
	}

	apiToken := cmd.String("api-token")
	if apiToken == "" {
		apiToken = os.Getenv("CLASH_ROYALE_API_TOKEN")
	}
	if apiToken == "" {
		return fmt.Errorf("API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag")
	}

	dataDir := cmd.String("data-dir")
	statsPath := filepath.Join(dataDir, "cards_stats.json")
	stats, statsErr := clashroyale.LoadStats(statsPath)
	if statsErr != nil {
		fprintf(os.Stderr, "Warning: failed to load combat stats at %s: %v\n", statsPath, statsErr)
	}

	client := clashroyale.NewClient(apiToken)
	players := make([]research.PlayerInput, 0, len(tags))
	for _, tag := range tags {
		player, getErr := client.GetPlayer(tag)
		if getErr != nil {
			return fmt.Errorf("failed to fetch player %s: %w", tag, getErr)
		}
		candidates := research.BuildCandidatesFromPlayer(player, stats)
		if len(candidates) < 8 {
			return fmt.Errorf("player %s has only %d candidate cards", tag, len(candidates))
		}
		players = append(players, research.PlayerInput{
			Tag:        "#" + strings.TrimPrefix(player.Tag, "#"),
			Name:       player.Name,
			Candidates: candidates,
		})
	}

	builder := deck.NewBuilder(dataDir)
	runner := research.BenchmarkRunner{Builder: builder}
	report, runErr := runner.Run(research.BenchmarkConfig{
		Tags:      tags,
		Seed:      int64(cmd.Int("seed")),
		TopN:      cmd.Int("top"),
		Methods:   methods,
		OutputDir: cmd.String("output-dir"),
		DataDir:   dataDir,
	}, players)
	if runErr != nil {
		return runErr
	}

	jsonPath, mdPath, writeErr := research.WriteReport(cmd.String("output-dir"), report)
	if writeErr != nil {
		return writeErr
	}

	printf("Research benchmark completed.\n")
	printf("Methods: %s\n", strings.Join(methods, ", "))
	printf("Players: %d\n", len(players))
	printf("JSON: %s\n", jsonPath)
	printf("Markdown: %s\n", mdPath)
	return nil
}
