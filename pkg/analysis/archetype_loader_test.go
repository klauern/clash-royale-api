package analysis

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadArchetypes_EmbeddedDefaults(t *testing.T) {
	// Test loading embedded defaults (empty path)
	archetypes, err := LoadArchetypes("")
	if err != nil {
		t.Fatalf("LoadArchetypes with empty path failed: %v", err)
	}

	if len(archetypes) != 27 {
		t.Errorf("Expected 27 default archetypes, got %d", len(archetypes))
	}

	// Verify first archetype
	if archetypes[0].Name != "Golem Beatdown" {
		t.Errorf("Expected first archetype to be 'Golem Beatdown', got '%s'", archetypes[0].Name)
	}

	if archetypes[0].WinCondition != "Golem" {
		t.Errorf("Expected WinCondition 'Golem', got '%s'", archetypes[0].WinCondition)
	}
}

func TestLoadArchetypes_CustomFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom_archetypes.json")

	configJSON := `{
		"version": 1,
		"archetypes": [
			{
				"name": "Custom Beatdown",
				"win_condition": "Golem",
				"support_cards": ["Baby Dragon", "Mega Minion"],
				"min_elixir": 3.0,
				"max_elixir": 4.0
			}
		]
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0o644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load from custom file
	archetypes, err := LoadArchetypes(configPath)
	if err != nil {
		t.Fatalf("LoadArchetypes from file failed: %v", err)
	}

	if len(archetypes) != 1 {
		t.Errorf("Expected 1 custom archetype, got %d", len(archetypes))
	}

	if archetypes[0].Name != "Custom Beatdown" {
		t.Errorf("Expected archetype 'Custom Beatdown', got '%s'", archetypes[0].Name)
	}
}

func TestLoadArchetypes_InvalidFile(t *testing.T) {
	// Test with non-existent file
	_, err := LoadArchetypes("/nonexistent/path/archetypes.json")
	if err == nil {
		t.Error("Expected error when loading non-existent file, got nil")
	}
}

func TestLoadArchetypes_InvalidJSON(t *testing.T) {
	// Create a file with invalid JSON
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(configPath, []byte("{invalid json"), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadArchetypes(configPath)
	if err == nil {
		t.Error("Expected error when loading invalid JSON, got nil")
	}
}

func TestValidateArchetype(t *testing.T) {
	tests := []struct {
		name      string
		archetype DeckArchetypeTemplate
		wantErr   bool
	}{
		{
			name: "Valid archetype",
			archetype: DeckArchetypeTemplate{
				Name:         "Test Deck",
				WinCondition: "Golem",
				MinElixir:    3.0,
				MaxElixir:    4.0,
			},
			wantErr: false,
		},
		{
			name: "Missing name",
			archetype: DeckArchetypeTemplate{
				WinCondition: "Golem",
				MinElixir:    3.0,
				MaxElixir:    4.0,
			},
			wantErr: true,
		},
		{
			name: "Missing win condition",
			archetype: DeckArchetypeTemplate{
				Name:      "Test Deck",
				MinElixir: 3.0,
				MaxElixir: 4.0,
			},
			wantErr: true,
		},
		{
			name: "Min elixir > Max elixir",
			archetype: DeckArchetypeTemplate{
				Name:         "Test Deck",
				WinCondition: "Golem",
				MinElixir:    5.0,
				MaxElixir:    3.0,
			},
			wantErr: true,
		},
		{
			name: "Negative elixir",
			archetype: DeckArchetypeTemplate{
				Name:         "Test Deck",
				WinCondition: "Golem",
				MinElixir:    -1.0,
				MaxElixir:    4.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArchetype(&tt.archetype)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateArchetype() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveDefaultArchetypes(t *testing.T) {
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "exported_archetypes.json")

	if err := SaveDefaultArchetypes(destPath); err != nil {
		t.Fatalf("SaveDefaultArchetypes failed: %v", err)
	}

	// Verify the file was created
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	// Verify it contains expected content
	if len(data) == 0 {
		t.Error("Exported file is empty")
	}

	// Verify we can load it back
	archetypes, err := LoadArchetypes(destPath)
	if err != nil {
		t.Fatalf("Failed to load exported archetypes: %v", err)
	}

	if len(archetypes) != 27 {
		t.Errorf("Expected 27 archetypes in exported file, got %d", len(archetypes))
	}
}

func TestLoadArchetypes_EmptyConfig(t *testing.T) {
	// Test with empty archetype list
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "empty.json")

	configJSON := `{"version": 1, "archetypes": []}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadArchetypes(configPath)
	if err == nil {
		t.Error("Expected error for empty archetype list, got nil")
	}
}
