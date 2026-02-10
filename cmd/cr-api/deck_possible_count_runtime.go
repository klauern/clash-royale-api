package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/urfave/cli/v3"
)

func deckPossibleCountCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	format := cmd.String("format")
	verbose := cmd.Bool("verbose")
	outputFile := cmd.String("output")

	// Get API token
	apiToken := cmd.String("api-token")
	if apiToken == "" {
		return fmt.Errorf("API token required (set CLASH_ROYALE_API_TOKEN or use --api-token)")
	}

	// Fetch player data
	client := clashroyale.NewClient(apiToken)
	player, err := client.GetPlayer(tag)
	if err != nil {
		return fmt.Errorf("failed to get player data: %w", err)
	}

	// Create deck space calculator
	calculator, err := deck.NewDeckSpaceCalculator(player)
	if err != nil {
		return fmt.Errorf("failed to create calculator: %w", err)
	}

	// Calculate statistics
	stats := calculator.CalculateStats()

	// Format output
	var output string
	switch strings.ToLower(format) {
	case "json":
		output, err = formatPossibleCountJSON(player, stats)
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
	case "csv":
		output = formatPossibleCountCSV(player, stats, verbose)
	case "human":
		fallthrough
	default:
		output = formatPossibleCountHuman(player, stats, verbose)
	}

	// Output to file or stdout
	if outputFile != "" {
		err = os.WriteFile(outputFile, []byte(output), 0o644)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		printf("Results saved to: %s\n", outputFile)
	} else {
		fmt.Print(output)
	}

	return nil
}

func formatPossibleCountHuman(player *clashroyale.Player, stats *deck.DeckSpaceStats, verbose bool) string {
	var buf strings.Builder

	buf.WriteString("╔════════════════════════════════════════════════════════════════════════╗\n")
	buf.WriteString("║               DECK COMBINATION CALCULATOR                              ║\n")
	buf.WriteString("╚════════════════════════════════════════════════════════════════════════╝\n\n")

	buf.WriteString(fmt.Sprintf("Player: %s (Tag: %s)\n", player.Name, player.Tag))
	buf.WriteString(fmt.Sprintf("Total Cards: %d\n\n", stats.TotalCards))

	buf.WriteString("═══ POSSIBLE DECK COMBINATIONS ═══\n\n")

	// Total combinations
	buf.WriteString(fmt.Sprintf("Total Unconstrained:  %s (%s)\n",
		stats.TotalCombinations.String(),
		deck.FormatLargeNumber(stats.TotalCombinations)))

	buf.WriteString(fmt.Sprintf("Valid (With Roles):   %s (%s)\n\n",
		stats.ValidCombinations.String(),
		deck.FormatLargeNumber(stats.ValidCombinations)))

	// Combinations by elixir range
	if len(stats.ByElixirRange) > 0 {
		buf.WriteString("═══ ESTIMATED BY ELIXIR RANGE ═══\n\n")
		w2 := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
		fprintf(w2, "Range\tCombinations\n")
		fprintf(w2, "─────\t────────────\n")

		for _, elixirRange := range deck.StandardElixirRanges {
			if count, exists := stats.ByElixirRange[elixirRange.Label]; exists {
				fprintf(w2, "%s\t%s\n",
					elixirRange.Label,
					deck.FormatLargeNumber(count))
			}
		}
		flushWriter(w2)
		buf.WriteString("\n")
	}

	// Combinations by archetype
	if len(stats.ByArchetype) > 0 {
		buf.WriteString("═══ ESTIMATED BY ARCHETYPE ═══\n\n")
		w3 := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
		fprintf(w3, "Archetype\tCombinations\n")
		fprintf(w3, "─────────\t────────────\n")

		archetypes := []string{"Beatdown", "Control", "Cycle", "Siege", "Bridge Spam", "Bait"}
		for _, archetype := range archetypes {
			if count, exists := stats.ByArchetype[archetype]; exists {
				fprintf(w3, "%s\t%s\n",
					archetype,
					deck.FormatLargeNumber(count))
			}
		}
		flushWriter(w3)
		buf.WriteString("\n")
	}

	if verbose {
		// Cards by role
		buf.WriteString("═══ CARDS BY ROLE ═══\n\n")
		w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
		fprintf(w, "Role\tCount\n")
		fprintf(w, "────\t─────\n")

		// Print in a specific order
		roles := []deck.CardRole{
			deck.RoleWinCondition,
			deck.RoleBuilding,
			deck.RoleSpellBig,
			deck.RoleSpellSmall,
			deck.RoleSupport,
			deck.RoleCycle,
		}
		roleLabels := map[deck.CardRole]string{
			deck.RoleWinCondition: "Win Condition",
			deck.RoleBuilding:     "Building",
			deck.RoleSpellBig:     "Big Spell",
			deck.RoleSpellSmall:   "Small Spell",
			deck.RoleSupport:      "Support",
			deck.RoleCycle:        "Cycle",
		}

		for _, role := range roles {
			count := stats.CardsByRole[role]
			label := roleLabels[role]
			fprintf(w, "%s\t%d\n", label, count)
		}
		flushWriter(w)
		buf.WriteString("\n")
	}

	buf.WriteString("Note: Estimates for elixir ranges and archetypes are approximations.\n")
	buf.WriteString("Default deck composition: 1 win condition, 1 building, 1 big spell,\n")
	buf.WriteString("1 small spell, 2 support, 2 cycle.\n")

	return buf.String()
}

func formatPossibleCountJSON(player *clashroyale.Player, stats *deck.DeckSpaceStats) (string, error) {
	output := map[string]any{
		"player": map[string]string{
			"tag":  player.Tag,
			"name": player.Name,
		},
		"total_cards":              stats.TotalCards,
		"total_combinations":       stats.TotalCombinations.String(),
		"valid_combinations":       stats.ValidCombinations.String(),
		"total_combinations_human": deck.FormatLargeNumber(stats.TotalCombinations),
		"valid_combinations_human": deck.FormatLargeNumber(stats.ValidCombinations),
		"cards_by_role":            stats.CardsByRole,
	}

	// Add elixir ranges
	elixirRanges := make(map[string]string)
	for label, count := range stats.ByElixirRange {
		elixirRanges[label] = deck.FormatLargeNumber(count)
	}
	output["by_elixir_range"] = elixirRanges

	// Add archetypes
	archetypes := make(map[string]string)
	for archetype, count := range stats.ByArchetype {
		archetypes[archetype] = deck.FormatLargeNumber(count)
	}
	output["by_archetype"] = archetypes

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func formatPossibleCountCSV(player *clashroyale.Player, stats *deck.DeckSpaceStats, verbose bool) string {
	var buf strings.Builder

	// Header
	buf.WriteString("Metric,Value\n")

	// Basic info
	buf.WriteString(fmt.Sprintf("Player Name,%s\n", player.Name))
	buf.WriteString(fmt.Sprintf("Player Tag,%s\n", player.Tag))
	buf.WriteString(fmt.Sprintf("Total Cards,%d\n", stats.TotalCards))
	buf.WriteString(fmt.Sprintf("Total Combinations,%s\n", stats.TotalCombinations.String()))
	buf.WriteString(fmt.Sprintf("Valid Combinations,%s\n", stats.ValidCombinations.String()))
	buf.WriteString(fmt.Sprintf("Total Combinations (Formatted),%s\n", deck.FormatLargeNumber(stats.TotalCombinations)))
	buf.WriteString(fmt.Sprintf("Valid Combinations (Formatted),%s\n\n", deck.FormatLargeNumber(stats.ValidCombinations)))

	if verbose {
		// Cards by role
		buf.WriteString("Role,Card Count\n")
		roles := []deck.CardRole{
			deck.RoleWinCondition,
			deck.RoleBuilding,
			deck.RoleSpellBig,
			deck.RoleSpellSmall,
			deck.RoleSupport,
			deck.RoleCycle,
		}
		roleLabels := map[deck.CardRole]string{
			deck.RoleWinCondition: "Win Condition",
			deck.RoleBuilding:     "Building",
			deck.RoleSpellBig:     "Big Spell",
			deck.RoleSpellSmall:   "Small Spell",
			deck.RoleSupport:      "Support",
			deck.RoleCycle:        "Cycle",
		}

		for _, role := range roles {
			count := stats.CardsByRole[role]
			label := roleLabels[role]
			buf.WriteString(fmt.Sprintf("%s,%d\n", label, count))
		}
	}

	return buf.String()
}
