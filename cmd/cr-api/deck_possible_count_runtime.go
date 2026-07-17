package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/klauer/clash-royale-api/go/pkg/deck"
	"github.com/urfave/cli/v3"
)

type possibleCountRoleRow struct {
	Role  string `json:"role"`
	Count int    `json:"count"`
}

type possibleCountCountRow struct {
	Label string `json:"label"`
	Count string `json:"count"`
}

var possibleCountRoleOrder = []struct {
	role  deck.CardRole
	label string
}{
	{role: deck.RoleWinCondition, label: "Win Condition"},
	{role: deck.RoleBuilding, label: "Building"},
	{role: deck.RoleSpellBig, label: "Big Spell"},
	{role: deck.RoleSpellSmall, label: "Small Spell"},
	{role: deck.RoleSupport, label: "Support"},
	{role: deck.RoleCycle, label: "Cycle"},
}

var possibleCountArchetypeOrder = []string{
	"Beatdown",
	"Control",
	"Cycle",
	"Siege",
	"Bridge Spam",
	"Bait",
}

func orderedPossibleCountRoleRows(stats *deck.DeckSpaceStats) []possibleCountRoleRow {
	rows := make([]possibleCountRoleRow, 0, len(possibleCountRoleOrder))
	for _, entry := range possibleCountRoleOrder {
		rows = append(rows, possibleCountRoleRow{
			Role:  entry.label,
			Count: stats.CardsByRole[entry.role],
		})
	}

	return rows
}

func orderedPossibleCountElixirRows(stats *deck.DeckSpaceStats) []possibleCountCountRow {
	rows := make([]possibleCountCountRow, 0, len(deck.StandardElixirRanges))
	for _, elixirRange := range deck.StandardElixirRanges {
		count, exists := stats.ByElixirRange[elixirRange.Label]
		if !exists {
			continue
		}

		rows = append(rows, possibleCountCountRow{
			Label: elixirRange.Label,
			Count: deck.FormatLargeNumber(count),
		})
	}

	return rows
}

func orderedPossibleCountArchetypeRows(stats *deck.DeckSpaceStats) []possibleCountCountRow {
	rows := make([]possibleCountCountRow, 0, len(stats.ByArchetype))
	seen := make(map[string]struct{}, len(stats.ByArchetype))

	for _, archetype := range possibleCountArchetypeOrder {
		count, exists := stats.ByArchetype[archetype]
		if !exists {
			continue
		}

		rows = append(rows, possibleCountCountRow{
			Label: archetype,
			Count: deck.FormatLargeNumber(count),
		})
		seen[archetype] = struct{}{}
	}

	extras := make([]string, 0, len(stats.ByArchetype))
	for archetype := range stats.ByArchetype {
		if _, exists := seen[archetype]; exists {
			continue
		}
		extras = append(extras, archetype)
	}
	sort.Strings(extras)

	for _, archetype := range extras {
		rows = append(rows, possibleCountCountRow{
			Label: archetype,
			Count: deck.FormatLargeNumber(stats.ByArchetype[archetype]),
		})
	}

	return rows
}

func writePossibleCountTableHeader(buf *strings.Builder, title, left, right string) *tabwriter.Writer {
	buf.WriteString(title)
	buf.WriteString("\n\n")

	writer := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)
	fprintf(writer, "%s\t%s\n", left, right)
	fprintf(writer, "%s\t%s\n", strings.Repeat("─", len(left)), strings.Repeat("─", len(right)))

	return writer
}

func writePossibleCountElixirSection(buf *strings.Builder, stats *deck.DeckSpaceStats) {
	rows := orderedPossibleCountElixirRows(stats)
	if len(rows) == 0 {
		return
	}

	writer := writePossibleCountTableHeader(buf, "═══ ESTIMATED BY ELIXIR RANGE ═══", "Range", "Combinations")
	for _, row := range rows {
		fprintf(writer, "%s\t%s\n", row.Label, row.Count)
	}
	flushWriter(writer)
	buf.WriteString("\n")
}

func writePossibleCountArchetypeSection(buf *strings.Builder, stats *deck.DeckSpaceStats) {
	rows := orderedPossibleCountArchetypeRows(stats)
	if len(rows) == 0 {
		return
	}

	writer := writePossibleCountTableHeader(buf, "═══ ESTIMATED BY ARCHETYPE ═══", "Archetype", "Combinations")
	for _, row := range rows {
		fprintf(writer, "%s\t%s\n", row.Label, row.Count)
	}
	flushWriter(writer)
	buf.WriteString("\n")
}

func writePossibleCountRoleSection(buf *strings.Builder, stats *deck.DeckSpaceStats) {
	writer := writePossibleCountTableHeader(buf, "═══ CARDS BY ROLE ═══", "Role", "Count")
	for _, row := range orderedPossibleCountRoleRows(stats) {
		fprintf(writer, "%s\t%d\n", row.Role, row.Count)
	}
	flushWriter(writer)
	buf.WriteString("\n")
}

func mapPossibleCountRows(rows []possibleCountCountRow) map[string]string {
	mapped := make(map[string]string, len(rows))
	for _, row := range rows {
		mapped[row.Label] = row.Count
	}

	return mapped
}

func deckPossibleCountCommand(ctx context.Context, cmd *cli.Command) error {
	tag := cmd.String("tag")
	format := cmd.String("format")
	verbose := cmd.Bool("verbose")
	outputFile := cmd.String("output")

	client, err := requireAPIClient(cmd, apiClientOptions{
		missingToken: "API token required (set CLASH_ROYALE_API_TOKEN or use --api-token)",
	})
	if err != nil {
		return err
	}

	player, err := client.GetPlayerWithContext(ctx, tag)
	if err != nil {
		return fmt.Errorf("failed to get player data: %w", err)
	}

	calculator, err := deck.NewDeckSpaceCalculator(player)
	if err != nil {
		return fmt.Errorf("failed to create calculator: %w", err)
	}

	stats := calculator.CalculateStats()

	var output string
	switch strings.ToLower(format) {
	case storageFormatJSON:
		output, err = formatPossibleCountJSON(player, stats)
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
	case batchFormatCSV:
		output = formatPossibleCountCSV(player, stats, verbose)
	case batchFormatHuman:
		fallthrough
	default:
		output = formatPossibleCountHuman(player, stats, verbose)
	}

	return writeTextOutput(output, outputFile, textOutputOptions{
		saveMessage: "Results saved to",
	})
}

func formatPossibleCountHuman(player *clashroyale.Player, stats *deck.DeckSpaceStats, verbose bool) string {
	var buf strings.Builder

	buf.WriteString("╔════════════════════════════════════════════════════════════════════════╗\n")
	buf.WriteString("║                    DECK COMBINATION CALCULATOR                       ║\n")
	buf.WriteString("╚════════════════════════════════════════════════════════════════════════╝\n\n")

	buf.WriteString(fmt.Sprintf("Player: %s (Tag: %s)\n", player.Name, player.Tag))
	buf.WriteString(fmt.Sprintf("Total Cards: %d\n\n", stats.TotalCards))

	buf.WriteString("═══ POSSIBLE DECK COMBINATIONS ═══\n\n")
	buf.WriteString(fmt.Sprintf("Total Unconstrained: %s (%s)\n",
		stats.TotalCombinations.String(),
		deck.FormatLargeNumber(stats.TotalCombinations)))
	buf.WriteString(fmt.Sprintf("Valid (With Roles): %s (%s)\n\n",
		stats.ValidCombinations.String(),
		deck.FormatLargeNumber(stats.ValidCombinations)))

	writePossibleCountElixirSection(&buf, stats)
	writePossibleCountArchetypeSection(&buf, stats)

	if verbose {
		writePossibleCountRoleSection(&buf, stats)
	}

	buf.WriteString("Note: Estimates for elixir ranges and archetypes are approximations.\n")
	buf.WriteString("Default deck composition: 1 win condition, 1 building, 1 big spell,\n")
	buf.WriteString("1 small spell, 2 support, 2 cycle.\n")

	return buf.String()
}

func formatPossibleCountJSON(player *clashroyale.Player, stats *deck.DeckSpaceStats) (string, error) {
	roleRows := orderedPossibleCountRoleRows(stats)
	elixirRows := orderedPossibleCountElixirRows(stats)
	archetypeRows := orderedPossibleCountArchetypeRows(stats)

	output := map[string]any{
		"player":                   map[string]string{"tag": player.Tag, "name": player.Name},
		"total_cards":              stats.TotalCards,
		"total_combinations":       stats.TotalCombinations.String(),
		"valid_combinations":       stats.ValidCombinations.String(),
		"total_combinations_human": deck.FormatLargeNumber(stats.TotalCombinations),
		"valid_combinations_human": deck.FormatLargeNumber(stats.ValidCombinations),
		"cards_by_role":            stats.CardsByRole,
		"cards_by_role_ordered":    roleRows,
		"by_elixir_range":          mapPossibleCountRows(elixirRows),
		"by_elixir_range_ordered":  elixirRows,
		"by_archetype":             mapPossibleCountRows(archetypeRows),
		"by_archetype_ordered":     archetypeRows,
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func formatPossibleCountCSV(player *clashroyale.Player, stats *deck.DeckSpaceStats, verbose bool) string {
	var buf strings.Builder

	buf.WriteString("Metric,Value\n")
	buf.WriteString(fmt.Sprintf("Player Name,%s\n", player.Name))
	buf.WriteString(fmt.Sprintf("Player Tag,%s\n", player.Tag))
	buf.WriteString(fmt.Sprintf("Total Cards,%d\n", stats.TotalCards))
	buf.WriteString(fmt.Sprintf("Total Combinations,%s\n", stats.TotalCombinations.String()))
	buf.WriteString(fmt.Sprintf("Valid Combinations,%s\n", stats.ValidCombinations.String()))
	buf.WriteString(fmt.Sprintf("Total Combinations (Formatted),%s\n", deck.FormatLargeNumber(stats.TotalCombinations)))
	buf.WriteString(fmt.Sprintf("Valid Combinations (Formatted),%s\n\n", deck.FormatLargeNumber(stats.ValidCombinations)))

	if verbose {
		buf.WriteString("Role,Card Count\n")
		for _, row := range orderedPossibleCountRoleRows(stats) {
			buf.WriteString(fmt.Sprintf("%s,%d\n", row.Role, row.Count))
		}
	}

	return buf.String()
}
