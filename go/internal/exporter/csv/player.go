package csv

import (
	"fmt"
	"path/filepath"

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
		"Experience Level",
		"Experience Points",
		"Trophies",
		"Best Trophies",
		"Wins",
		"Losses",
		"Total Battles",
		"Three Crown Wins",
		"Challenge Wins",
		"Tournament Wins",
		"Role",
		"Clan Tag",
		"Clan Name",
		"Arena ID",
		"Arena Name",
		"League ID",
		"League Name",
		"Donations",
		"Star Points",
		"Created At",
	}
}

// playerExport exports player data to CSV
func playerExport(dataDir string, data interface{}) error {
	player, ok := data.(*clashroyale.Player)
	if !ok {
		return fmt.Errorf("expected Player type, got %T", data)
	}

	// Prepare CSV rows
	rows := [][]string{
		{
			player.Tag,
			player.Name,
			fmt.Sprintf("%d", player.ExpLevel),
			fmt.Sprintf("%d", player.ExpPoints),
			fmt.Sprintf("%d", player.Trophies),
			fmt.Sprintf("%d", player.BestTrophies),
			fmt.Sprintf("%d", player.Wins),
			fmt.Sprintf("%d", player.Losses),
			fmt.Sprintf("%d", player.BattleCount),
			fmt.Sprintf("%d", player.ThreeCrownWins),
			fmt.Sprintf("%d", player.ChallengeWins),
			fmt.Sprintf("%d", player.TournamentWins),
			player.Role,
			func() string {
				if player.Clan != nil {
					return player.Clan.Tag
				}
				return ""
			}(),
			func() string {
				if player.Clan != nil {
					return player.Clan.Name
				}
				return ""
			}(),
			fmt.Sprintf("%d", player.Arena.ID),
			player.Arena.Name,
			fmt.Sprintf("%d", player.League.ID),
			player.League.Name,
			fmt.Sprintf("%d", player.Donations),
			fmt.Sprintf("%d", player.StarPoints),
			player.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "players.csv"}
	filePath := filepath.Join(dataDir, "csv", "players", exporter.FilenameBase)
	return exporter.writeCSV(filePath, playerHeaders(), rows)
}