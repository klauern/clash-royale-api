package csv

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
)

func TestNewPlayerExporter(t *testing.T) {
	exporter := NewPlayerExporter()

	if exporter.Filename() != "players.csv" {
		t.Errorf("NewPlayerExporter() filename = %v, want players.csv", exporter.Filename())
	}

	headers := exporter.Headers()
	// Just check that we have headers (they're many)
	if len(headers) < 40 {
		t.Errorf("NewPlayerExporter() headers count = %d, want at least 40", len(headers))
	}
}

func TestPlayerExport(t *testing.T) {
	tempDir := t.TempDir()

	// Create mock player data
	mockPlayer := &clashroyale.Player{
		Tag:            "#ABC123",
		Name:           "Test Player",
		ExpLevel:       50,
		Trophies:       4000,
		BestTrophies:   4500,
		Wins:           2000,
		Losses:         1500,
		BattleCount:    3500,
		ThreeCrownWins: 800,
		ChallengeWins:  50,
		TournamentWins: 10,
		Clan: &clashroyale.Clan{
			Tag:  "#CLAN123",
			Name: "Test Clan",
		},
		Arena: clashroyale.Arena{
			ID:   54000000,
			Name: "Legendary Arena",
		},
		League: clashroyale.League{
			ID:   29000022,
			Name: "Legendary League",
		},
		StarPoints:     5000,
		Donations:      1000,
		TotalDonations: 5000,
		CreatedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	exporter := NewPlayerExporter()
	err := exporter.Export(tempDir, mockPlayer)
	if err != nil {
		t.Fatalf("PlayerExport() error = %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, "csv", "players", "players.csv")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("PlayerExport() file was not created at %s", filePath)
		return
	}

	// Read and verify content
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open exported file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	// Should have header + 1 data row
	if len(records) != 2 {
		t.Errorf("PlayerExport() row count = %d, want 2", len(records))
	}

	// Check that we have data and the first field is the tag
	if len(records) > 1 {
		playerData := records[1]
		if len(playerData) == 0 {
			t.Error("PlayerExport() player data is empty")
		} else if playerData[0] != "#ABC123" {
			t.Errorf("PlayerExport() player tag = %v, want #ABC123", playerData[0])
		}
	}
}

func TestPlayerExport_NoClan(t *testing.T) {
	tempDir := t.TempDir()

	// Create player without clan
	mockPlayer := &clashroyale.Player{
		Tag:          "#NOCLAN",
		Name:         "Lonely Player",
		ExpLevel:     30,
		Trophies:     3000,
		BestTrophies: 3200,
		Wins:         500,
		Losses:       400,
		BattleCount:  900,
		// No clan data
		Arena: clashroyale.Arena{
			ID:   54000000,
			Name: "Legendary Arena",
		},
		League: clashroyale.League{
			ID:   29000022,
			Name: "Legendary League",
		},
		CreatedAt: time.Date(2021, 6, 15, 12, 30, 0, 0, time.UTC),
	}

	exporter := NewPlayerExporter()
	err := exporter.Export(tempDir, mockPlayer)
	if err != nil {
		t.Fatalf("PlayerExport() no clan error = %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tempDir, "csv", "players", "players.csv")
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open exported file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("PlayerExport() row count = %d, want 2", len(records))
	}

	// Player should exist even without clan
	if len(records) > 1 {
		playerData := records[1]
		if len(playerData) == 0 {
			t.Error("PlayerExport() player data is empty")
		} else if playerData[0] != "#NOCLAN" {
			t.Errorf("PlayerExport() player tag = %v, want #NOCLAN", playerData[0])
		}
	}
}

func TestNewPlayerCardsExporter(t *testing.T) {
	exporter := NewPlayerCardsExporter()

	if exporter.Filename() != "player_cards.csv" {
		t.Errorf("NewPlayerCardsExporter() filename = %v, want player_cards.csv", exporter.Filename())
	}

	headers := exporter.Headers()
	if len(headers) == 0 {
		t.Error("NewPlayerCardsExporter() should have headers")
	}
}

func TestPlayerCardsExport(t *testing.T) {
	tempDir := t.TempDir()

	// Create mock player with cards
	mockPlayer := &clashroyale.Player{
		Tag:  "#CARDS123",
		Name: "Card Collector",
		Cards: []clashroyale.Card{
			{
				ID:         28000000,
				Name:       "Knight",
				ElixirCost: 3,
				Type:       "troop",
				Rarity:     "common",
				Count:      5,
				Level:      11,
			},
			{
				ID:         28000001,
				Name:       "Fireball",
				ElixirCost: 4,
				Type:       "spell",
				Rarity:     "rare",
				Count:      2,
				Level:      8,
			},
		},
	}

	exporter := NewPlayerCardsExporter()
	err := exporter.Export(tempDir, mockPlayer)
	if err != nil {
		t.Fatalf("PlayerCardsExport() error = %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, "csv", "players", "player_cards.csv")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("PlayerCardsExport() file was not created at %s", filePath)
		return
	}

	// Read and verify content
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open exported file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}

	// Should have header + 2 data rows
	if len(records) != 3 {
		t.Errorf("PlayerCardsExport() row count = %d, want 3", len(records))
	}
}

func TestPlayerExport_InvalidDataType(t *testing.T) {
	tempDir := t.TempDir()

	exporter := NewPlayerExporter()
	err := exporter.Export(tempDir, "not a player")

	// Should not crash
	if err == nil {
		t.Log("PlayerExport() with invalid data did not return error")
	}
}
