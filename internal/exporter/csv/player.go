package csv

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// NewPlayerExporter creates a new player CSV exporter
func NewPlayerExporter() *CSVExporter {
	return NewCSVExporter(
		"players.csv",
		playerHeaders,
		playerExport,
	)
}

// playerHeaders returns the CSV headers for player data
func playerHeaders() []string {
	return []string{
		"Tag",
		"Name",
		"Name Set",
		"Experience Level",
		"Experience Points",
		"Trophies",
		"Best Trophies",
		"Wins",
		"Losses",
		"Battle Count",
		"Win Rate",
		"Three Crown Wins",
		"Three Crown Rate",
		"Challenge Wins",
		"Challenge Max Wins",
		"Tournament Wins",
		"Tournament Battle Count",
		"Total Donations",
		"Challenge Cards Won",
		"Player Level",
		"Player Experience",
		"Role",
		"Clan Tag",
		"Clan Name",
		"Clan Score",
		"Clan Donations",
		"Clan Badge ID",
		"Clan Type",
		"Clan Members",
		"Clan Required Trophies",
		"Arena ID",
		"Arena Name",
		"Arena Trophy Limit",
		"League ID",
		"League Name",
		"Donations",
		"Star Points",
		"Total Cards",
		"Cards Collection",
		"Current Deck",
		"Created At",
	}
}

// playerExport exports player data to CSV.
func playerExport(dataDir string, data any) error {
	player, ok := data.(*clashroyale.Player)
	if !ok {
		return fmt.Errorf("expected Player type, got %T", data)
	}
	rows := [][]string{playerCSVRow(player)}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "players.csv"}
	filePath := filepath.Join(dataDir, "csv", "players", exporter.FilenameBase)
	return exporter.writeCSV(filePath, playerHeaders(), rows)
}

func playerCSVRow(player *clashroyale.Player) []string {
	winRate := ratio(player.Wins, player.BattleCount)
	threeCrownRate := ratio(player.ThreeCrownWins, player.BattleCount)
	return []string{
		player.Tag,
		player.Name,
		strconv.FormatBool(player.NameSet),
		strconv.Itoa(player.ExpLevel),
		strconv.Itoa(player.ExpPoints),
		strconv.Itoa(player.Trophies),
		strconv.Itoa(player.BestTrophies),
		strconv.Itoa(player.Wins),
		strconv.Itoa(player.Losses),
		strconv.Itoa(player.BattleCount),
		fmt.Sprintf("%.2f%%", winRate*100),
		strconv.Itoa(player.ThreeCrownWins),
		fmt.Sprintf("%.2f%%", threeCrownRate*100),
		strconv.Itoa(player.ChallengeWins),
		strconv.Itoa(player.ChallengeMaxWins),
		strconv.Itoa(player.TournamentWins),
		strconv.Itoa(player.TournamentBattleCount),
		strconv.Itoa(player.TotalDonations),
		strconv.Itoa(player.ChallengeCardsWon),
		strconv.Itoa(player.Level),
		strconv.Itoa(player.Experience),
		player.Role,
		clanString(player.Clan, func(c *clashroyale.Clan) string { return c.Tag }),
		clanString(player.Clan, func(c *clashroyale.Clan) string { return c.Name }),
		clanString(player.Clan, func(c *clashroyale.Clan) string { return strconv.Itoa(c.ClanScore) }),
		clanString(player.Clan, func(c *clashroyale.Clan) string { return strconv.Itoa(c.Donations) }),
		clanString(player.Clan, func(c *clashroyale.Clan) string { return strconv.Itoa(c.BadgeID) }),
		clanString(player.Clan, func(c *clashroyale.Clan) string { return c.Type }),
		clanString(player.Clan, func(c *clashroyale.Clan) string { return strconv.Itoa(c.Members) }),
		clanString(player.Clan, func(c *clashroyale.Clan) string { return strconv.Itoa(c.RequiredTrophies) }),
		strconv.Itoa(player.Arena.ID),
		player.Arena.Name,
		strconv.Itoa(player.Arena.TrophyLimit),
		strconv.Itoa(player.League.ID),
		player.League.Name,
		strconv.Itoa(player.Donations),
		strconv.Itoa(player.StarPoints),
		strconv.Itoa(len(player.Cards)),
		formatDeckCards(player.Cards),
		formatDeckCards(player.CurrentDeck),
		player.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func clanString(clan *clashroyale.Clan, extract func(*clashroyale.Clan) string) string {
	if clan == nil {
		return ""
	}
	return extract(clan)
}

func ratio(value, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(value) / float64(total)
}

// NewPlayerCardsExporter creates a new player cards CSV exporter
func NewPlayerCardsExporter() *CSVExporter {
	return NewCSVExporter(
		"player_cards.csv",
		playerCardsHeaders,
		playerCardsExport,
	)
}

// playerCardsHeaders returns the CSV headers for player cards data
func playerCardsHeaders() []string {
	return []string{
		"Player Tag",
		"Player Name",
		"Card ID",
		"Card Name",
		"Card Level",
		"Max Level",
		"Evolution Level",
		"Max Evolution Level",
		"Card Count",
		"Elixir Cost",
		"Card Type",
		"Rarity",
		"Icon URL",
	}
}

// playerCardsExport exports detailed player card information to CSV
func playerCardsExport(dataDir string, data any) error {
	player, ok := data.(*clashroyale.Player)
	if !ok {
		return fmt.Errorf("expected Player type, got %T", data)
	}

	// Prepare CSV rows
	var rows [][]string

	for _, card := range player.Cards {
		row := []string{
			player.Tag,
			player.Name,
			fmt.Sprintf("%d", card.ID),
			card.Name,
			fmt.Sprintf("%d", card.Level),
			fmt.Sprintf("%d", card.MaxLevel),
			fmt.Sprintf("%d", card.EvolutionLevel),
			fmt.Sprintf("%d", card.MaxEvolutionLevel),
			fmt.Sprintf("%d", card.Count),
			fmt.Sprintf("%d", card.ElixirCost),
			card.Type,
			card.Rarity,
			"", // Icon URL not available in current struct
		}
		rows = append(rows, row)
	}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "player_cards.csv"}
	filePath := filepath.Join(dataDir, "csv", "players", exporter.FilenameBase)
	return exporter.writeCSV(filePath, playerCardsHeaders(), rows)
}
