package csv

import (
	"fmt"
	"strconv"

	"github.com/klauer/clash-royale-api/go/internal/storage"
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
		"Deck Cards",
	}
}

// battleLogExport exports battle log data to CSV.
func battleLogExport(dataDir string, data any) error {
	battles, ok := data.([]clashroyale.Battle)
	if !ok {
		return fmt.Errorf("expected []Battle type, got %T", data)
	}

	rows := makeBattleLogRows(battles)

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "battle_log.csv"}
	filePath := exporter.csvFilePath(dataDir, storage.CSVBattlesSubdir)
	return exporter.writeCSV(filePath, battleLogHeaders(), rows)
}

func makeBattleLogRows(battles []clashroyale.Battle) [][]string {
	rows := make([][]string, 0, len(battles))
	for _, battle := range battles {
		row, ok := battleLogRow(battle)
		if !ok {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

func battleLogRow(battle clashroyale.Battle) ([]string, bool) {
	if len(battle.Team) == 0 || len(battle.Opponent) == 0 {
		return nil, false
	}
	player := battle.Team[0]
	opponent := battle.Opponent[0]
	row := []string{
		battle.UTCDate.Format("2006-01-02 15:04:05"),
		battle.Type,
		player.Tag,
		player.Name,
		strconv.Itoa(player.StartingTrophies),
		strconv.Itoa(player.TrophyChange),
		strconv.Itoa(player.Crowns),
		opponent.Tag,
		opponent.Name,
		strconv.Itoa(opponent.StartingTrophies),
		strconv.Itoa(opponent.TrophyChange),
		strconv.Itoa(opponent.Crowns),
		strconv.FormatBool(battle.IsLadderTournament),
		strconv.Itoa(len(battle.Team)),
		formatPositiveIntOrEmpty(battle.DeckAverage),
		battle.GameMode.DeckLink,
		strconv.FormatBool(battle.GameMode.NotCounted),
		formatDeckCards(battle.Deck),
	}
	return row, true
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

// battleSummaryExport exports battle summary statistics to CSV.
func battleSummaryExport(dataDir string, data any) error {
	battles, ok := data.([]clashroyale.Battle)
	if !ok {
		return fmt.Errorf("expected []Battle type, got %T", data)
	}

	// If no battles, return early
	if len(battles) == 0 {
		return nil
	}

	stats := summarizeBattles(battles)
	row := stats.toCSVRow()

	// Create exporter and write to file
	exporter := &BaseExporter{FilenameBase: "battle_summary.csv"}
	filePath := exporter.csvFilePath(dataDir, storage.CSVBattlesSubdir)

	// Create a single-row CSV
	rows := [][]string{row}
	return exporter.writeCSV(filePath, battleSummaryHeaders(), rows)
}

type battleSummaryStats struct {
	PlayerTag         string
	PlayerName        string
	TotalBattles      int
	Wins              int
	Losses            int
	TotalCrownsWon    int
	TotalCrownsLost   int
	TotalTrophyChange int
	LadderBattles     int
	ChallengeBattles  int
	TournamentBattles int
	ThreeCrowns       int
	ThreeCrownsLost   int
	CurrentStreak     int
	BestStreak        int
	streak            int
	BestTrophy        int
	WorstTrophy       int
}

func summarizeBattles(battles []clashroyale.Battle) battleSummaryStats {
	stats := battleSummaryStats{}
	for _, battle := range battles {
		stats.addBattle(battle)
	}
	stats.CurrentStreak = stats.streak
	return stats
}

func (s *battleSummaryStats) addBattle(battle clashroyale.Battle) {
	if len(battle.Team) == 0 || len(battle.Opponent) == 0 {
		return
	}
	s.TotalBattles++
	player := battle.Team[0]
	opponent := battle.Opponent[0]
	if s.PlayerTag == "" {
		s.PlayerTag = player.Tag
		s.PlayerName = player.Name
	}

	s.trackResult(player.Crowns, opponent.Crowns)
	s.TotalCrownsWon += player.Crowns
	s.TotalCrownsLost += opponent.Crowns
	s.trackThreeCrowns(player.Crowns, opponent.Crowns)
	s.trackTrophy(player.StartingTrophies + player.TrophyChange)
	s.TotalTrophyChange += player.TrophyChange
	s.categorizeBattle(battle)
}

func (s *battleSummaryStats) trackResult(playerCrowns, opponentCrowns int) {
	if playerCrowns > opponentCrowns {
		s.Wins++
		s.streak++
		if s.streak > s.BestStreak {
			s.BestStreak = s.streak
		}
		return
	}
	s.Losses++
	s.streak = 0
}

func (s *battleSummaryStats) trackThreeCrowns(playerCrowns, opponentCrowns int) {
	if playerCrowns == 3 {
		s.ThreeCrowns++
	}
	if opponentCrowns == 3 {
		s.ThreeCrownsLost++
	}
}

func (s *battleSummaryStats) trackTrophy(result int) {
	if result > s.BestTrophy {
		s.BestTrophy = result
	}
	if result < s.WorstTrophy || s.WorstTrophy == 0 {
		s.WorstTrophy = result
	}
}

func (s *battleSummaryStats) categorizeBattle(battle clashroyale.Battle) {
	if battle.IsLadderTournament {
		s.LadderBattles++
		return
	}
	if battle.Type == "PvP" {
		s.ChallengeBattles++
		return
	}
	if battle.Type == "tournament" {
		s.TournamentBattles++
	}
}

func (s battleSummaryStats) winRate() float64 {
	if s.TotalBattles == 0 {
		return 0
	}
	return float64(s.Wins) / float64(s.TotalBattles)
}

func (s battleSummaryStats) avgTrophyChange() float64 {
	if s.TotalBattles == 0 {
		return 0
	}
	return float64(s.TotalTrophyChange) / float64(s.TotalBattles)
}

func (s battleSummaryStats) threeCrownRate() float64 {
	if s.TotalBattles == 0 {
		return 0
	}
	return float64(s.ThreeCrowns) / float64(s.TotalBattles)
}

func (s battleSummaryStats) toCSVRow() []string {
	return []string{
		s.PlayerTag,
		s.PlayerName,
		strconv.Itoa(s.TotalBattles),
		strconv.Itoa(s.Wins),
		strconv.Itoa(s.Losses),
		fmt.Sprintf("%.2f%%", s.winRate()*100),
		strconv.Itoa(s.TotalCrownsWon),
		strconv.Itoa(s.TotalCrownsLost),
		strconv.Itoa(s.TotalTrophyChange),
		fmt.Sprintf("%.1f", s.avgTrophyChange()),
		strconv.Itoa(s.LadderBattles),
		strconv.Itoa(s.ChallengeBattles),
		strconv.Itoa(s.TournamentBattles),
		strconv.Itoa(s.BestTrophy),
		strconv.Itoa(s.WorstTrophy),
		strconv.Itoa(s.CurrentStreak),
		strconv.Itoa(s.BestStreak),
		strconv.Itoa(s.ThreeCrowns),
		strconv.Itoa(s.ThreeCrownsLost),
		fmt.Sprintf("%.2f%%", s.threeCrownRate()*100),
	}
}

func formatDeckCards(cards []clashroyale.Card) string {
	if len(cards) == 0 {
		return ""
	}
	cardNames := make([]string, len(cards))
	for i, card := range cards {
		cardNames[i] = fmt.Sprintf("%s (Lv.%d)", card.Name, card.Level)
	}
	return fmt.Sprintf("%v", cardNames)
}

func formatPositiveIntOrEmpty(value int) string {
	if value <= 0 {
		return ""
	}
	return strconv.Itoa(value)
}
