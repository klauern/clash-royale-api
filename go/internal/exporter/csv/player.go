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

// playerExport exports player data to CSV
func playerExport(dataDir string, data interface{}) error {
	player, ok := data.(*clashroyale.Player)
	if !ok {
		return fmt.Errorf("expected Player type, got %T", data)
	}

	// Calculate derived statistics
	winRate := float64(0)
	if player.BattleCount > 0 {
		winRate = float64(player.Wins) / float64(player.BattleCount)
	}

	threeCrownRate := float64(0)
	if player.BattleCount > 0 {
		threeCrownRate = float64(player.ThreeCrownWins) / float64(player.BattleCount)
	}

	// Format cards collection
	cardsCollection := ""
	if len(player.Cards) > 0 {
		cardNames := make([]string, len(player.Cards))
		for i, card := range player.Cards {
			cardNames[i] = fmt.Sprintf("%s (Lv.%d)", card.Name, card.Level)
		}
		cardsCollection = fmt.Sprintf("%v", cardNames)
	}

	// Format current deck
	currentDeck := ""
	if len(player.CurrentDeck) > 0 {
		cardNames := make([]string, len(player.CurrentDeck))
		for i, card := range player.CurrentDeck {
			cardNames[i] = fmt.Sprintf("%s (Lv.%d)", card.Name, card.Level)
		}
		currentDeck = fmt.Sprintf("%v", cardNames)
	}

	// Prepare CSV rows
	rows := [][]string{
		{
			player.Tag,
			player.Name,
			fmt.Sprintf("%t", player.NameSet),
			fmt.Sprintf("%d", player.ExpLevel),
			fmt.Sprintf("%d", player.ExpPoints),
			fmt.Sprintf("%d", player.Trophies),
			fmt.Sprintf("%d", player.BestTrophies),
			fmt.Sprintf("%d", player.Wins),
			fmt.Sprintf("%d", player.Losses),
			fmt.Sprintf("%d", player.BattleCount),
			fmt.Sprintf("%.2f%%", winRate*100),
			fmt.Sprintf("%d", player.ThreeCrownWins),
			fmt.Sprintf("%.2f%%", threeCrownRate*100),
			fmt.Sprintf("%d", player.ChallengeWins),
			fmt.Sprintf("%d", player.ChallengeMaxWins),
			fmt.Sprintf("%d", player.TournamentWins),
			fmt.Sprintf("%d", player.TournamentBattleCount),
			fmt.Sprintf("%d", player.TotalDonations),
			fmt.Sprintf("%d", player.ChallengeCardsWon),
			fmt.Sprintf("%d", player.Level),
			fmt.Sprintf("%d", player.Experience),
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
			func() string {
				if player.Clan != nil {
					return fmt.Sprintf("%d", player.Clan.ClanScore)
				}
				return ""
			}(),
			func() string {
				if player.Clan != nil {
					return fmt.Sprintf("%d", player.Clan.Donations)
				}
				return ""
			}(),
			func() string {
				if player.Clan != nil {
					return fmt.Sprintf("%d", player.Clan.BadgeID)
				}
				return ""
			}(),
			func() string {
				if player.Clan != nil {
					return player.Clan.Type
				}
				return ""
			}(),
			func() string {
				if player.Clan != nil {
					return fmt.Sprintf("%d", player.Clan.Members)
				}
				return ""
			}(),
			func() string {
				if player.Clan != nil {
					return fmt.Sprintf("%d", player.Clan.RequiredTrophies)
				}
				return ""
			}(),
			fmt.Sprintf("%d", player.Arena.ID),
			player.Arena.Name,
			fmt.Sprintf("%d", player.Arena.TrophyLimit),
			fmt.Sprintf("%d", player.League.ID),
			player.League.Name,
			fmt.Sprintf("%d", player.Donations),
			fmt.Sprintf("%d", player.StarPoints),
			fmt.Sprintf("%d", len(player.Cards)),
			cardsCollection,
			currentDeck,
			player.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "players.csv"}
	filePath := filepath.Join(dataDir, "csv", "players", exporter.FilenameBase)
	return exporter.writeCSV(filePath, playerHeaders(), rows)
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
		"Card Count",
		"Elixir Cost",
		"Card Type",
		"Rarity",
		"Icon URL",
	}
}

// playerCardsExport exports detailed player card information to CSV
func playerCardsExport(dataDir string, data interface{}) error {
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
