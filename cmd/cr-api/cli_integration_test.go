//go:build integration

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/urfave/cli/v3"
)

// TestCLIIntegration performs end-to-end testing of the CLI
func TestCLIIntegration(t *testing.T) {
	// Set test environment
	os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")

	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T, tempDir string)
		wantErr  bool
	}{
		{
			name: "Player command with export",
			args: []string{
				"player",
				"--tag", "TEST12345",
				"--export-csv",
				"--save",
			},
			validate: validatePlayerExport,
			wantErr:  false,
		},
		{
			name: "Cards command with export",
			args: []string{
				"cards",
				"--export-csv",
			},
			validate: validateCardsExport,
			wantErr:  false,
		},
		{
			name: "Analyze command with export",
			args: []string{
				"analyze",
				"--tag", "TEST12345",
				"--export-csv",
				"--save",
			},
			validate: validateAnalysisExport,
			wantErr:  false,
		},
		{
			name: "Deck build command",
			args: []string{
				"deck",
				"build",
				"--tag", "TEST12345",
				"--strategy", "balanced",
				"--save",
				"--export-csv",
			},
			validate: validateDeckBuild,
			wantErr:  false,
		},
		{
			name: "Event scan command",
			args: []string{
				"events",
				"scan",
				"--tag", "TEST12345",
				"--days", "7",
				"--save",
				"--export-csv",
			},
			validate: validateEventScan,
			wantErr:  false,
		},
		{
			name: "Export all command",
			args: []string{
				"export",
				"all",
				"--tag", "TEST12345",
				"--timestamp",
			},
			validate: validateExportAll,
			wantErr:  false,
		},
		{
			name: "Player command without tag",
			args: []string{
				"player",
				"--export-csv",
			},
			wantErr: true,
		},
		{
			name: "Invalid command",
			args: []string{
				"invalid_command",
			},
			wantErr: true,
		},
		{
			name: "Deck build-suite command",
			args: []string{
				"deck",
				"build-suite",
				"--tag", "TEST12345",
				"--strategies", "balanced,aggro",
				"--variations", "2",
				"--save",
			},
			validate: validateDeckSuiteBuild,
			wantErr:  false,
		},
		{
			name: "Deck evaluate-batch command",
			args: []string{
				"deck",
				"evaluate-batch",
				"--tag", "TEST12345",
				"--deck-dir", "data/decks",
				"--format", "json",
				"--save-aggregated",
			},
			validate: validateEvaluateBatch,
			wantErr:  false,
		},
		{
			name: "Deck compare command",
			args: []string{
				"deck",
				"compare",
				"--from-evaluations", "data/evaluations/*.json",
				"--format", "markdown",
				"--report-output", "data/reports/comparison.md",
			},
			validate: validateCompareReport,
			wantErr:  false,
		},
		{
			name: "Deck analyze-suite command",
			args: []string{
				"deck",
				"analyze-suite",
				"--tag", "TEST12345",
				"--strategies", "all",
				"--variations", "1",
				"--top-n", "5",
			},
			validate: validateAnalyzeSuite,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset temp dir for each test
			testDir := t.TempDir()
			os.Setenv("DATA_DIR", testDir)

			// Update command with new test dir
			cmd := createTestCommand(testDir)

			err := cmd.Run(context.Background(), append([]string{"cr-api"}, tt.args...))

			if (err != nil) != tt.wantErr {
				t.Errorf("command error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, testDir)
			}
		})
	}
}

// TestCLIWorkflow tests a complete user workflow
func TestCLIWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")
	os.Setenv("DATA_DIR", tempDir)

	playerTag := "WORKFLOW123"

	// Step 1: Analyze player and export all data
	t.Run("Step1_AnalyzeAndExport", func(t *testing.T) {
		cmd := createTestCommand(tempDir)
		args := []string{
			"export",
			"all",
			"--tag", playerTag,
			"--timestamp",
		}

		err := cmd.Run(context.Background(), append([]string{"cr-api"}, args...))
		if err != nil {
			t.Errorf("Export all failed: %v", err)
		}

		// Verify exports were created
		verifyWorkflowExports(t, tempDir, playerTag)
	})

	// Step 2: Build a deck
	t.Run("Step2_BuildDeck", func(t *testing.T) {
		cmd := createTestCommand(tempDir)
		args := []string{
			"deck",
			"build",
			"--tag", playerTag,
			"--strategy", "aggro",
			"--max-elixir", "3.5",
			"--save",
			"--export-csv",
		}

		err := cmd.Run(context.Background(), append([]string{"cr-api"}, args...))
		if err != nil {
			t.Errorf("Deck build failed: %v", err)
		}
	})

	// Step 3: Scan for events
	t.Run("Step3_ScanEvents", func(t *testing.T) {
		cmd := createTestCommand(tempDir)
		args := []string{
			"events",
			"scan",
			"--tag", playerTag,
			"--days", "3",
			"--event-types", "challenge",
			"--save",
		}

		err := cmd.Run(context.Background(), append([]string{"cr-api"}, args...))
		if err != nil {
			t.Errorf("Event scan failed: %v", err)
		}
	})

	// Step 4: Export specific data types
	t.Run("Step4_ExportSpecific", func(t *testing.T) {
		cmd := createTestCommand(tempDir)
		args := []string{
			"export",
			"analysis",
			"--tag", playerTag,
			"--types", "priorities,rarity",
			"--min-priority-score", "40.0",
		}

		err := cmd.Run(context.Background(), append([]string{"cr-api"}, args...))
		if err != nil {
			t.Errorf("Analysis export failed: %v", err)
		}
	})
}

// TestDeckSuiteWorkflow tests the complete deck analysis suite workflow
func TestDeckSuiteWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")
	os.Setenv("DATA_DIR", tempDir)

	playerTag := "SUITE123"

	// Step 1: Build deck suite with multiple strategies
	t.Run("Step1_BuildDeckSuite", func(t *testing.T) {
		cmd := createTestCommand(tempDir)
		args := []string{
			"deck",
			"build-suite",
			"--tag", playerTag,
			"--strategies", "balanced,aggro,control",
			"--variations", "2",
			"--min-elixir", "2.5",
			"--max-elixir", "4.5",
			"--save",
		}

		err := cmd.Run(context.Background(), append([]string{"cr-api"}, args...))
		if err != nil {
			t.Errorf("Build suite failed: %v", err)
		}

		validateDeckSuiteBuild(t, tempDir)
	})

	// Step 2: Evaluate all built decks
	t.Run("Step2_EvaluateBatch", func(t *testing.T) {
		cmd := createTestCommand(tempDir)
		args := []string{
			"deck",
			"evaluate-batch",
			"--tag", playerTag,
			"--from-suite", filepath.Join(tempDir, "decks", "*_suite_summary_*.json"),
			"--format", "json",
			"--sort-by", "overall",
			"--save-aggregated",
		}

		err := cmd.Run(context.Background(), append([]string{"cr-api"}, args...))
		if err != nil {
			t.Errorf("Evaluate batch failed: %v", err)
		}

		validateEvaluateBatch(t, tempDir)
	})

	// Step 3: Compare top performers
	t.Run("Step3_CompareTopDecks", func(t *testing.T) {
		cmd := createTestCommand(tempDir)
		args := []string{
			"deck",
			"compare",
			"--from-evaluations", filepath.Join(tempDir, "evaluations", "*.json"),
			"--auto-select-top", "5",
			"--format", "markdown",
			"--report-output", filepath.Join(tempDir, "reports", "analysis_report.md"),
		}

		err := cmd.Run(context.Background(), append([]string{"cr-api"}, args...))
		if err != nil {
			t.Errorf("Compare decks failed: %v", err)
		}

		validateCompareReport(t, tempDir)
	})

	// Step 4: Full analyze-suite command (all-in-one)
	t.Run("Step4_AnalyzeSuite_AllInOne", func(t *testing.T) {
		cmd := createTestCommand(tempDir)
		args := []string{
			"deck",
			"analyze-suite",
			"--tag", playerTag,
			"--strategies", "all",
			"--variations", "1",
			"--top-n", "5",
			"--output-dir", filepath.Join(tempDir, "full_analysis"),
		}

		err := cmd.Run(context.Background(), append([]string{"cr-api"}, args...))
		if err != nil {
			t.Errorf("Analyze suite failed: %v", err)
		}

		// Validate complete output structure
		validateAnalyzeSuite(t, filepath.Join(tempDir, "full_analysis"))
	})
}

// TestErrorHandling tests various error conditions
func TestErrorHandling(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		args       []string
		setup      func()
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "Missing API token",
			args: []string{
				"player",
				"--tag", "TEST123",
			},
			setup: func() {
				os.Unsetenv("CLASH_ROYALE_API_TOKEN")
			},
			wantErr:    true,
			wantErrMsg: "API token is required",
		},
		{
			name: "Invalid player tag",
			args: []string{
				"player",
				"--tag", "", // Empty tag
			},
			setup: func() {
				os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")
			},
			wantErr: true,
		},
		{
			name: "Invalid deck strategy",
			args: []string{
				"deck",
				"build",
				"--tag", "TEST123",
				"--strategy", "invalid_strategy",
			},
			setup: func() {
				os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")
			},
			wantErr: true,
		},
		{
			name: "Invalid export type",
			args: []string{
				"export",
				"player",
				"--tag", "TEST123",
				"--types", "invalid_type",
			},
			setup: func() {
				os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")
			},
			wantErr: false, // Should warn but not fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			cmd := createTestCommand(tempDir)
			err := cmd.Run(context.Background(), append([]string{"cr-api"}, tt.args...))

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrMsg != "" {
				if err == nil || err.Error() != tt.wantErrMsg {
					t.Errorf("error message = %v, want %v", err, tt.wantErrMsg)
				}
			}
		})
	}
}

// TestPerformanceAndConcurrency tests CLI performance and concurrent usage
func TestPerformanceAndConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tempDir := t.TempDir()
	os.Setenv("CLASH_ROYALE_API_TOKEN", "test_token")
	os.Setenv("DATA_DIR", tempDir)

	// Test command execution time
	t.Run("PerformanceTest", func(t *testing.T) {
		cmd := createTestCommand(tempDir)
		args := []string{
			"player",
			"--tag", "PERF123",
			"--export-csv",
		}

		start := time.Now()
		err := cmd.Run(context.Background(), append([]string{"cr-api"}, args...))
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Command failed: %v", err)
		}

		// Command should complete within reasonable time (5 seconds for test)
		if duration > 5*time.Second {
			t.Errorf("Command took too long: %v", duration)
		}

		t.Logf("Command completed in: %v", duration)
	})

	// Test concurrent command execution
	t.Run("ConcurrencyTest", func(t *testing.T) {
		const numConcurrent = 5

		results := make(chan error, numConcurrent)

		for i := 0; i < numConcurrent; i++ {
			go func(id int) {
				cmd := createTestCommand(tempDir)
				args := []string{
					"cards",
					"--export-csv",
				}

				err := cmd.Run(context.Background(), append([]string{"cr-api"}, args...))
				results <- err
			}(i)
		}

		// Wait for all commands to complete
		for i := 0; i < numConcurrent; i++ {
			err := <-results
			if err != nil {
				t.Errorf("Concurrent command %d failed: %v", i, err)
			}
		}
	})
}

// Helper functions
func createTestCommand(dataDir string) *cli.Command {
	return &cli.Command{
		Name:  "cr-api",
		Usage: "Clash Royale API client and analysis tool",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "api-token",
				Aliases: []string{"t"},
				Sources: cli.EnvVars("CLASH_ROYALE_API_TOKEN"),
			},
			&cli.StringFlag{
				Name:    "data-dir",
				Aliases: []string{"d"},
				Value:   dataDir,
				Sources: cli.EnvVars("DATA_DIR"),
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
			},
		},
		Commands: []*cli.Command{
			addDeckCommands(),
			addEvolutionCommands(),
			addEventCommands(),
			addExportCommands(),
			{
				Name:  "player",
				Usage: "Get player information",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "chests",
						Usage: "Show upcoming chests",
					},
					&cli.BoolFlag{
						Name:  "save",
						Usage: "Save player data to file",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export player data to CSV",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					// Mock implementation for testing
					return mockPlayerCommand(cmd)
				},
			},
			{
				Name:  "cards",
				Usage: "Get card database",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export card database to CSV",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					// Mock implementation for testing
					return mockCardsCommand(cmd)
				},
			},
			{
				Name:  "analyze",
				Usage: "Analyze player card collection and upgrade priorities",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag",
						Aliases:  []string{"p"},
						Usage:    "Player tag (without #)",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "include-max-level",
						Usage: "Include max level cards in analysis",
					},
					&cli.Float64Flag{
						Name:  "min-priority-score",
						Value: 30.0,
						Usage: "Minimum priority score for upgrade recommendations",
					},
					&cli.StringSliceFlag{
						Name:  "focus-rarities",
						Usage: "Focus on specific rarities (Common, Rare, Epic, Legendary, Champion)",
					},
					&cli.StringSliceFlag{
						Name:  "exclude-cards",
						Usage: "Exclude specific cards from recommendations",
					},
					&cli.BoolFlag{
						Name:  "prioritize-win-cons",
						Value: true,
						Usage: "Boost priority for win condition cards",
					},
					&cli.IntFlag{
						Name:  "top-n",
						Value: 15,
						Usage: "Show top N upgrade priorities",
					},
					&cli.BoolFlag{
						Name:  "save",
						Usage: "Save analysis to JSON file",
					},
					&cli.BoolFlag{
						Name:  "export-csv",
						Usage: "Export analysis to CSV",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					// Mock implementation for testing
					return mockAnalyzeCommand(cmd)
				},
			},
		},
	}
}

// Mock command implementations for testing
func mockPlayerCommand(cmd *cli.Command) error {
	// Mock player data creation
	dataDir := cmd.String("data-dir")

	if cmd.Bool("export-csv") || cmd.Bool("save") {
		playersDir := filepath.Join(dataDir, "csv", "players")
		os.MkdirAll(playersDir, 0o755)

		// Create mock CSV file
		csvFile := filepath.Join(playersDir, "players.csv")
		os.WriteFile(csvFile, []byte(mockPlayerCSV), 0o644)

		// Create mock JSON file
		if cmd.Bool("save") {
			jsonFile := filepath.Join(dataDir, "players", "TEST12345.json")
			os.MkdirAll(filepath.Dir(jsonFile), 0o755)
			os.WriteFile(jsonFile, []byte(mockPlayerJSON), 0o644)
		}
	}

	return nil
}

func mockCardsCommand(cmd *cli.Command) error {
	if cmd.Bool("export-csv") {
		dataDir := cmd.String("data-dir")
		refDir := filepath.Join(dataDir, "csv", "reference")
		os.MkdirAll(refDir, 0o755)

		csvFile := filepath.Join(refDir, "cards.csv")
		os.WriteFile(csvFile, []byte(mockCardsCSV), 0o644)
	}

	return nil
}

func mockAnalyzeCommand(cmd *cli.Command) error {
	dataDir := cmd.String("data-dir")

	if cmd.Bool("export-csv") {
		analysisDir := filepath.Join(dataDir, "csv", "analysis")
		os.MkdirAll(analysisDir, 0o755)

		csvFile := filepath.Join(analysisDir, "card_analysis.csv")
		os.WriteFile(csvFile, []byte(mockAnalysisCSV), 0o644)
	}

	if cmd.Bool("save") {
		analysisDir := filepath.Join(dataDir, "analysis")
		os.MkdirAll(analysisDir, 0o755)

		jsonFile := filepath.Join(analysisDir, "TEST12345.json")
		os.WriteFile(jsonFile, []byte(mockAnalysisJSON), 0o644)
	}

	return nil
}

// Validation functions
func validatePlayerExport(t *testing.T, tempDir string) {
	// Check for player CSV
	csvPath := filepath.Join(tempDir, "csv", "players", "players.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Errorf("Player CSV file not created: %s", csvPath)
	}

	// Check for player JSON
	jsonPath := filepath.Join(tempDir, "players", "TEST12345.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Errorf("Player JSON file not created: %s", jsonPath)
	}
}

func validateCardsExport(t *testing.T, tempDir string) {
	csvPath := filepath.Join(tempDir, "csv", "reference", "cards.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Errorf("Cards CSV file not created: %s", csvPath)
	}
}

func validateAnalysisExport(t *testing.T, tempDir string) {
	csvPath := filepath.Join(tempDir, "csv", "analysis", "card_analysis.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Errorf("Analysis CSV file not created: %s", csvPath)
	}

	jsonPath := filepath.Join(tempDir, "analysis", "TEST12345.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Errorf("Analysis JSON file not created: %s", jsonPath)
	}
}

func validateDeckBuild(t *testing.T, tempDir string) {
	decksDir := filepath.Join(tempDir, "decks")
	if _, err := os.Stat(decksDir); os.IsNotExist(err) {
		t.Errorf("Decks directory not created")
	}
}

func validateEventScan(t *testing.T, tempDir string) {
	eventDir := filepath.Join(tempDir, "event_decks")
	if _, err := os.Stat(eventDir); os.IsNotExist(err) {
		t.Errorf("Event decks directory not created")
	}
}

func validateDeckSuiteBuild(t *testing.T, tempDir string) {
	decksDir := filepath.Join(tempDir, "decks")
	if _, err := os.Stat(decksDir); os.IsNotExist(err) {
		t.Errorf("Decks directory not created for suite")
		return
	}

	// Check for suite summary file
	files, err := os.ReadDir(decksDir)
	if err != nil {
		t.Errorf("Failed to read decks directory: %v", err)
		return
	}

	foundSummary := false
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			if filepath.Base(file.Name()) != "" {
				foundSummary = true
				break
			}
		}
	}

	if !foundSummary {
		t.Error("Suite summary JSON file not created")
	}
}

func validateEvaluateBatch(t *testing.T, tempDir string) {
	evalsDir := filepath.Join(tempDir, "evaluations")
	if _, err := os.Stat(evalsDir); os.IsNotExist(err) {
		t.Errorf("Evaluations directory not created")
		return
	}

	// Check for evaluation results file
	files, err := os.ReadDir(evalsDir)
	if err != nil {
		t.Errorf("Failed to read evaluations directory: %v", err)
		return
	}

	foundResults := false
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			foundResults = true
			break
		}
	}

	if !foundResults {
		t.Error("Evaluation results JSON file not created")
	}
}

func validateCompareReport(t *testing.T, tempDir string) {
	reportsDir := filepath.Join(tempDir, "reports")
	if _, err := os.Stat(reportsDir); os.IsNotExist(err) {
		t.Errorf("Reports directory not created")
		return
	}

	// Check for markdown report
	files, err := os.ReadDir(reportsDir)
	if err != nil {
		t.Errorf("Failed to read reports directory: %v", err)
		return
	}

	foundReport := false
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".md" {
			foundReport = true
			break
		}
	}

	if !foundReport {
		t.Error("Markdown report file not created")
	}
}

func validateAnalyzeSuite(t *testing.T, tempDir string) {
	// Check for all three output directories
	dirs := []string{"decks", "evaluations", "reports"}

	for _, dir := range dirs {
		dirPath := filepath.Join(tempDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("Analyze-suite directory not created: %s", dir)
		}
	}

	// Validate suite summary in decks/
	validateDeckSuiteBuild(t, tempDir)

	// Validate evaluation results in evaluations/
	validateEvaluateBatch(t, tempDir)

	// Validate markdown report in reports/
	validateCompareReport(t, tempDir)
}

func validateExportAll(t *testing.T, tempDir string) {
	// Check for various export directories
	dirs := []string{
		"csv/players",
		"csv/reference",
		"csv/analysis",
		"csv/battles",
		"csv/events",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(tempDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Export directory not created: %s", fullPath)
		}
	}
}

func verifyWorkflowExports(t *testing.T, tempDir, playerTag string) {
	// Verify all expected exports exist
	expectedFiles := []string{
		filepath.Join("csv", "players", "players.csv"),
		filepath.Join("csv", "reference", "cards.csv"),
		filepath.Join("csv", "analysis", "card_analysis.csv"),
		filepath.Join("players", playerTag+".json"),
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(tempDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Workflow export missing: %s", fullPath)
		}
	}
}

// Mock data constants
const mockPlayerCSV = `Tag,Name,Experience Level,Trophies,Wins,Losses
TEST12345,Test Player,50,4000,2000,1500
`

const mockPlayerJSON = `{
	"tag": "TEST12345",
	"name": "Test Player",
	"expLevel": 50,
	"trophies": 4000,
	"wins": 2000,
	"losses": 1500
}
`

const mockCardsCSV = `ID,Name,Elixir Cost,Type,Rarity
1,Arrows,3,Spell,Common
2,Fireball,4,Spell,Rare
`

const mockAnalysisJSON = `{
	"player_tag": "TEST12345",
	"total_cards": 108,
	"max_level_cards": 45,
	"upgradable_cards": 12
}
`

const mockAnalysisCSV = `Player Tag,Total Cards,Max Level Cards,Upgradable Cards
TEST12345,108,45,12
`
