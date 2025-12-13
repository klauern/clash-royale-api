package csv

import (
	"fmt"
	"path/filepath"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

// NewBattleLogExporter creates a new battle log CSV exporter
func NewBattleLogExporter() *CSVExporter {
	return NewCSVExporter(
		"battle_log.csv",
		battleLogHeaders,
		battleLogExport,
	)
}

// battleLogHeaders returns the CSV headers for battle log data
func battleLogHeaders() []string {
	return []string{
		"Timestamp",
		"Battle Type",
		"Player Tag",
		"Player Name",
		"Player Starting Trophies",
		"Player Trophy Change",
		"Player Crowns",
		"Opponent Tag",
		"Opponent Name",
		"Opponent Starting Trophies",
		"Opponent Trophy Change",
		"Opponent Crowns",
		"Is Ladder Tournament",
		"Team Size",
		"Deck Average Elixir",
		"Deck Link",
		"Not Counted",
	}
}

// battleLogExport exports battle log data to CSV
func battleLogExport(dataDir string, data interface{}) error {
	battles, ok := data.([]clashroyale.Battle)
	if !ok {
		return fmt.Errorf("expected []Battle type, got %T", data)
	}

	// Prepare CSV rows
	var rows [][]string

	for _, battle := range battles {
		// Extract player and opponent information
		// Assume first team is player's team and first opponent is main opponent
		if len(battle.Team) == 0 || len(battle.Opponent) == 0 {
			continue // Skip battles with missing data
		}

		player := battle.Team[0]
		opponent := battle.Opponent[0]

		// Format deck average elixir
		deckAvgElixir := ""
		if battle.DeckAverage > 0 {
			deckAvgElixir = fmt.Sprintf("%d", battle.DeckAverage)
		}

		// Format deck link
		deckLink := ""
		if battle.GameMode.DeckLink != "" {
			deckLink = battle.GameMode.DeckLink
		}

		// Format deck cards as string
		deckCards := ""
		if len(battle.Deck) > 0 {
			cardNames := make([]string, len(battle.Deck))
			for i, card := range battle.Deck {
				cardNames[i] = fmt.Sprintf("%s (Lv.%d)", card.Name, card.Level)
			}
			deckCards = fmt.Sprintf("%v", cardNames)
		}

		row := []string{
			battle.UTCDate.Format("2006-01-02 15:04:05"),
			battle.Type,
			player.Tag,
			player.Name,
			fmt.Sprintf("%d", player.StartingTrophies),
			fmt.Sprintf("%d", player.TrophyChange),
			fmt.Sprintf("%d", player.Crowns),
			opponent.Tag,
			opponent.Name,
			fmt.Sprintf("%d", opponent.StartingTrophies),
			fmt.Sprintf("%d", opponent.TrophyChange),
			fmt.Sprintf("%d", opponent.Crowns),
			fmt.Sprintf("%t", battle.IsLadderTournament),
			fmt.Sprintf("%d", len(battle.Team)), // Team size
			deckAvgElixir,
			deckLink,
			fmt.Sprintf("%t", battle.GameMode.NotCounted),
		}

		// Add deck cards as an additional column if it exists
		if deckCards != "" {
			// Append deck cards as the last column
			row = append(row, deckCards)
		}

		rows = append(rows, row)
	}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "battle_log.csv"}
	filePath := filepath.Join(dataDir, "csv", "battles", exporter.FilenameBase)
	return exporter.writeCSV(filePath, battleLogHeaders(), rows)
}

// NewBattleSummaryExporter creates a new battle summary CSV exporter
func NewBattleSummaryExporter() *CSVExporter {
	return NewCSVExporter(
		"battle_summary.csv",
		battleSummaryHeaders,
		battleSummaryExport,
	)
}

// battleSummaryHeaders returns the CSV headers for battle summary data
func battleSummaryHeaders() []string {
	return []string{
		"Player Tag",
		"Player Name",
		"Total Battles",
		"Wins",
		"Losses",
		"Win Rate",
		"Total Crown Wins",
		"Total Crown Losses",
		"Net Trophy Change",
		"Average Trophy Change",
		"Ladder Battles",
		"Challenge Battles",
		"Tournament Battles",
		"Best Trophy Result",
		"Worst Trophy Result",
		"Current Win Streak",
		"Best Win Streak",
		"Three Crown Wins",
		"Three Crown Losses",
		"Three Crown Rate",
	}
}

// battleSummaryExport exports battle summary statistics to CSV
func battleSummaryExport(dataDir string, data interface{}) error {
	battles, ok := data.([]clashroyale.Battle)
	if !ok {
		return fmt.Errorf("expected []Battle type, got %T", data)
	}

	// If no battles, return early
	if len(battles) == 0 {
		return nil
	}

	// Aggregate statistics
	// Use first battle's player as the reference player
	playerTag := battles[0].Team[0].Tag
	playerName := battles[0].Team[0].Name

	totalBattles := len(battles)
	wins := 0
	losses := 0
	totalCrownsWon := 0
	totalCrownsLost := 0
	totalTrophyChange := 0
	ladderBattles := 0
	challengeBattles := 0
	tournamentBattles := 0
	threeCrowns := 0
	threeCrownsLost := 0
	currentStreak := 0
	bestStreak := 0
	streak := 0
	bestTrophy := 0
	worstTrophy := 0

	for _, battle := range battles {
		if len(battle.Team) == 0 || len(battle.Opponent) == 0 {
			continue
		}

		player := battle.Team[0]
		opponent := battle.Opponent[0]

		// Track wins/losses
		if player.Crowns > opponent.Crowns {
			wins++
			streak++
			if streak > bestStreak {
				bestStreak = streak
			}
		} else {
			losses++
			streak = 0
		}

		// Track crowns
		totalCrownsWon += player.Crowns
		totalCrownsLost += opponent.Crowns

		// Track three crowns
		if player.Crowns == 3 {
			threeCrowns++
		}
		if opponent.Crowns == 3 {
			threeCrownsLost++
		}

		// Track trophies
		totalTrophyChange += player.TrophyChange
		if player.StartingTrophies+player.TrophyChange > bestTrophy {
			bestTrophy = player.StartingTrophies + player.TrophyChange
		}
		if player.StartingTrophies+player.TrophyChange < worstTrophy || worstTrophy == 0 {
			worstTrophy = player.StartingTrophies + player.TrophyChange
		}

		// Categorize battles
		if battle.IsLadderTournament {
			ladderBattles++
		} else if battle.Type == "PvP" {
			challengeBattles++
		} else if battle.Type == "tournament" {
			tournamentBattles++
		}
	}

	// Calculate derived stats
	winRate := float64(0)
	if totalBattles > 0 {
		winRate = float64(wins) / float64(totalBattles)
	}

	avgTrophyChange := float64(0)
	if totalBattles > 0 {
		avgTrophyChange = float64(totalTrophyChange) / float64(totalBattles)
	}

	threeCrownRate := float64(0)
	if totalBattles > 0 {
		threeCrownRate = float64(threeCrowns) / float64(totalBattles)
	}

	// Prepare CSV row
	row := []string{
		playerTag,
		playerName,
		fmt.Sprintf("%d", totalBattles),
		fmt.Sprintf("%d", wins),
		fmt.Sprintf("%d", losses),
		fmt.Sprintf("%.2f%%", winRate*100),
		fmt.Sprintf("%d", totalCrownsWon),
		fmt.Sprintf("%d", totalCrownsLost),
		fmt.Sprintf("%d", totalTrophyChange),
		fmt.Sprintf("%.1f", avgTrophyChange),
		fmt.Sprintf("%d", ladderBattles),
		fmt.Sprintf("%d", challengeBattles),
		fmt.Sprintf("%d", tournamentBattles),
		fmt.Sprintf("%d", bestTrophy),
		fmt.Sprintf("%d", worstTrophy),
		fmt.Sprintf("%d", currentStreak),
		fmt.Sprintf("%d", bestStreak),
		fmt.Sprintf("%d", threeCrowns),
		fmt.Sprintf("%d", threeCrownsLost),
		fmt.Sprintf("%.2f%%", threeCrownRate*100),
	}

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "battle_summary.csv"}
	filePath := filepath.Join(dataDir, "csv", "battles", exporter.FilenameBase)

	// Create a single-row CSV
	rows := [][]string{row}
	return exporter.writeCSV(filePath, battleSummaryHeaders(), rows)
}