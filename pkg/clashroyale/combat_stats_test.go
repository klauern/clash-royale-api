package clashroyale

import (
	"math"
	"os"
	"strings"
	"testing"
)

func TestLoadStats(t *testing.T) {
	// Create a temporary stats file for testing
	content := []byte(`{
		"stats": {
			"TestCard": {
				"hitpoints": 1000,
				"damage": 200,
				"speed": "Fast"
			}
		}
	}`)

	tmpfile, err := os.CreateTemp("", "stats_test_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test Loading
	registry, err := LoadStats(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadStats failed: %v", err)
	}

	// Test GetStats
	stats := registry.GetStats("TestCard")
	if stats == nil {
		t.Fatal("Expected stats for TestCard, got nil")
	}

	if stats.Hitpoints != 1000 {
		t.Errorf("Expected Hitpoints 1000, got %d", stats.Hitpoints)
	}
	if stats.Speed != "Fast" {
		t.Errorf("Expected Speed Fast, got %s", stats.Speed)
	}

	// Test Missing Card
	if registry.GetStats("MissingCard") != nil {
		t.Error("Expected nil for MissingCard")
	}
}

func TestLoadRealStatsFile(t *testing.T) {
	// Try to locate the real data file relative to this test file
	// This assumes the test is run from the project root or the package directory
	// We'll try a few common paths
	paths := []string{
		"../../data/static/cards_stats.json", // From package dir (go pkg structure)
		"data/static/cards_stats.json",       // From project root
		"../../../data/cards_stats.json",     // Legacy path
		"data/cards_stats.json",              // Legacy path
	}

	var foundPath string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			foundPath = p
			break
		}
	}

	if foundPath == "" {
		t.Skip("Could not find data/cards_stats.json, skipping integration test")
	}

	registry, err := LoadStats(foundPath)
	if err != nil {
		t.Fatalf("Failed to load real stats file: %v", err)
	}

	// Check for a known card
	knight := registry.GetStats("Knight")
	if knight == nil {
		t.Error("Expected to find Knight stats in real data file")
	} else {
		if knight.Hitpoints <= 0 {
			t.Error("Knight should have positive hitpoints")
		}
	}
}

func TestDPSPerElixir(t *testing.T) {
	tests := []struct {
		name      string
		stats     CombatStats
		elixir    int
		wantScore float64
	}{
		{
			name: "High DPS card",
			stats: CombatStats{
				DamagePerSecond: 100,
			},
			elixir:    4,
			wantScore: 25.0,
		},
		{
			name: "Zero elixir",
			stats: CombatStats{
				DamagePerSecond: 100,
			},
			elixir:    0,
			wantScore: 0,
		},
		{
			name:      "Zero DPS",
			stats:     CombatStats{},
			elixir:    3,
			wantScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := tt.stats.DPSPerElixir(tt.elixir)
			if score != tt.wantScore {
				t.Errorf("DPSPerElixir() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestHPPerElixir(t *testing.T) {
	tests := []struct {
		name      string
		stats     CombatStats
		elixir    int
		wantScore float64
	}{
		{
			name: "Tank card",
			stats: CombatStats{
				Hitpoints: 2000,
			},
			elixir:    5,
			wantScore: 400.0,
		},
		{
			name: "Zero elixir",
			stats: CombatStats{
				Hitpoints: 1000,
			},
			elixir:    0,
			wantScore: 0,
		},
		{
			name:      "Zero HP",
			stats:     CombatStats{},
			elixir:    3,
			wantScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := tt.stats.HPPerElixir(tt.elixir)
			if score != tt.wantScore {
				t.Errorf("HPPerElixir() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestRangeEffectiveness(t *testing.T) {
	tests := []struct {
		name      string
		stats     CombatStats
		wantScore float64
	}{
		{
			name: "Long range unit",
			stats: CombatStats{
				Range: 7.0,
			},
			wantScore: 1.0,
		},
		{
			name: "Medium range unit",
			stats: CombatStats{
				Range: 3.5,
			},
			wantScore: 0.5,
		},
		{
			name: "Melee unit",
			stats: CombatStats{
				Range: 0,
			},
			wantScore: 0,
		},
		{
			name: "Very long range (clamped)",
			stats: CombatStats{
				Range: 10.0,
			},
			wantScore: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := tt.stats.RangeEffectiveness()
			if score != tt.wantScore {
				t.Errorf("RangeEffectiveness() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestTargetCoverage(t *testing.T) {
	tests := []struct {
		name      string
		stats     CombatStats
		wantScore float64
	}{
		{
			name: "Air & Ground",
			stats: CombatStats{
				Targets: "Air & Ground",
			},
			wantScore: 1.0,
		},
		{
			name: "Ground only",
			stats: CombatStats{
				Targets: "Ground",
			},
			wantScore: 0.7,
		},
		{
			name: "Air only",
			stats: CombatStats{
				Targets: "Air",
			},
			wantScore: 0.6,
		},
		{
			name: "Buildings",
			stats: CombatStats{
				Targets: "Buildings",
			},
			wantScore: 0.4,
		},
		{
			name: "Unknown targets",
			stats: CombatStats{
				Targets: "Unknown",
			},
			wantScore: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := tt.stats.TargetCoverage()
			if score != tt.wantScore {
				t.Errorf("TargetCoverage() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestSpeedEffectiveness(t *testing.T) {
	tests := []struct {
		name      string
		stats     CombatStats
		wantScore float64
	}{
		{
			name: "Very Fast",
			stats: CombatStats{
				Speed: "Very Fast",
			},
			wantScore: 1.0,
		},
		{
			name: "Fast",
			stats: CombatStats{
				Speed: "Fast",
			},
			wantScore: 0.8,
		},
		{
			name: "Medium",
			stats: CombatStats{
				Speed: "Medium",
			},
			wantScore: 0.6,
		},
		{
			name: "Slow",
			stats: CombatStats{
				Speed: "Slow",
			},
			wantScore: 0.4,
		},
		{
			name: "Unknown speed",
			stats: CombatStats{
				Speed: "Unknown",
			},
			wantScore: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := tt.stats.SpeedEffectiveness()
			if score != tt.wantScore {
				t.Errorf("SpeedEffectiveness() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestRoleSpecificEffectiveness(t *testing.T) {
	tests := []struct {
		name      string
		stats     CombatStats
		role      string
		wantScore float64
	}{
		{
			name: "Win Condition with high HP and damage",
			stats: CombatStats{
				Hitpoints: 3000,
				Damage:    500,
			},
			role:      "wincondition",
			wantScore: 1.0,
		},
		{
			name: "Building with HP and lifetime",
			stats: CombatStats{
				Hitpoints: 2000,
				Lifetime:  60,
			},
			role:      "building",
			wantScore: 1.0,
		},
		{
			name: "Spell with damage and radius",
			stats: CombatStats{
				Damage: 600,
				Radius: 5,
			},
			role:      "spell",
			wantScore: 1.0,
		},
		{
			name: "Support with range and coverage",
			stats: CombatStats{
				Range:   7,
				Targets: "Air & Ground",
			},
			role:      "support",
			wantScore: 1.0,
		},
		{
			name: "Empty role",
			stats: CombatStats{
				Hitpoints: 1000,
				Damage:    100,
			},
			role:      "",
			wantScore: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := tt.stats.RoleSpecificEffectiveness(tt.role)
			if score != tt.wantScore {
				t.Errorf("RoleSpecificEffectiveness() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestCombatEffectiveness(t *testing.T) {
	tests := []struct {
		name      string
		stats     CombatStats
		elixir    int
		wantScore float64
	}{
		{
			name: "Balanced combat unit",
			stats: CombatStats{
				Hitpoints:       2000,
				DamagePerSecond: 100,
				Range:           5,
				Targets:         "Air & Ground",
				Speed:           "Medium",
			},
			elixir:    5,
			wantScore: 0.7171428571428571, // Calculated manually: DPS:100/5=20(0.4), HP:2000/5=400(1.0), Range:5/7=0.714(0.714), Targets:1.0(1.0), Speed:0.6(0.6) = 0.4*0.3+1.0*0.25+0.714*0.15+1.0*0.15+0.6*0.15 = 0.717
		},
		{
			name: "Zero elixir",
			stats: CombatStats{
				Hitpoints: 1000,
				Damage:    100,
			},
			elixir:    0,
			wantScore: 0,
		},
		{
			name:      "Nil stats",
			stats:     CombatStats{},
			elixir:    3,
			wantScore: 0.15, // Range 0 + Targets 0.5 + Speed 0.5 = 0.15 (since other factors are 0)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := tt.stats.CombatEffectiveness(tt.elixir)
			// Allow small tolerance for floating point calculations
			if math.Abs(score-tt.wantScore) > 0.01 {
				t.Errorf("CombatEffectiveness() = %v, want %v", score, tt.wantScore)
			}
		})
	}
}

func TestStatsSummary(t *testing.T) {
	tests := []struct {
		name      string
		stats     CombatStats
		wantParts []string
	}{
		{
			name: "Complete stats",
			stats: CombatStats{
				Hitpoints:       1000,
				Damage:          200,
				DamagePerSecond: 100,
				Range:           5.5,
				Targets:         "Air & Ground",
			},
			wantParts: []string{"HP: 1000", "DMG: 200", "DPS: 100", "Range: 5.5", "Targets: Air & Ground"},
		},
		{
			name: "Minimal stats",
			stats: CombatStats{
				Hitpoints: 500,
			},
			wantParts: []string{"HP: 500"},
		},
		{
			name:      "Empty stats",
			stats:     CombatStats{},
			wantParts: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.stats.StatsSummary()
			for _, part := range tt.wantParts {
				if !strings.Contains(summary, part) {
					t.Errorf("StatsSummary() missing %s. Got: %s", part, summary)
				}
			}
		})
	}
}
