// Package analysis provides archetype loading functionality for dynamic deck templates.
// This allows users to customize archetypes via external JSON configuration files.
package analysis

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed config/archetypes.json
var defaultArchetypesJSON []byte

// ArchetypeConfig represents the JSON configuration structure for deck archetypes
type ArchetypeConfig struct {
	Version    int                     `json:"version"`
	Archetypes []DeckArchetypeTemplate `json:"archetypes"`
}

// LoadArchetypes loads deck archetypes from a JSON file.
// If the file cannot be found or loaded, falls back to embedded default archetypes.
// This allows users to customize archetypes while maintaining backward compatibility.
func LoadArchetypes(configPath string) ([]DeckArchetypeTemplate, error) {
	// Try to load from file
	if configPath != "" {
		archetypes, err := loadArchetypesFromFile(configPath)
		if err == nil {
			return archetypes, nil
		}
		// Return error only if explicitly provided file fails
		return nil, fmt.Errorf("failed to load archetypes from %s: %w", configPath, err)
	}

	// Fall back to embedded defaults
	return loadDefaultArchetypes()
}

// loadArchetypesFromFile loads archetypes from a JSON file
func loadArchetypesFromFile(path string) ([]DeckArchetypeTemplate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return parseArchetypesJSON(data)
}

// loadDefaultArchetypes loads embedded default archetypes
func loadDefaultArchetypes() ([]DeckArchetypeTemplate, error) {
	return parseArchetypesJSON(defaultArchetypesJSON)
}

// parseArchetypesJSON parses archetype data from JSON bytes
func parseArchetypesJSON(data []byte) ([]DeckArchetypeTemplate, error) {
	var config ArchetypeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(config.Archetypes) == 0 {
		return nil, fmt.Errorf("no archetypes found in configuration")
	}

	// Validate and filter archetypes
	enabledArchetypes := make([]DeckArchetypeTemplate, 0, len(config.Archetypes))
	for i, archetype := range config.Archetypes {
		if err := validateArchetype(&archetype); err != nil {
			return nil, fmt.Errorf("archetype #%d (%s) is invalid: %w", i, archetype.Name, err)
		}

		// Only include enabled archetypes (default to true if not specified for backward compatibility)
		if archetype.Enabled {
			enabledArchetypes = append(enabledArchetypes, archetype)
		}
	}

	if len(enabledArchetypes) == 0 {
		return nil, fmt.Errorf("no enabled archetypes found in configuration")
	}

	return enabledArchetypes, nil
}

// validateArchetype validates that an archetype has all required fields
func validateArchetype(archetype *DeckArchetypeTemplate) error {
	if archetype.Name == "" {
		return fmt.Errorf("name is required")
	}
	if archetype.WinCondition == "" {
		return fmt.Errorf("win_condition is required")
	}
	if archetype.MinElixir < 0 || archetype.MinElixir > 10 {
		return fmt.Errorf("min_elixir must be between 0 and 10, got %.1f", archetype.MinElixir)
	}
	if archetype.MaxElixir < 0 || archetype.MaxElixir > 10 {
		return fmt.Errorf("max_elixir must be between 0 and 10, got %.1f", archetype.MaxElixir)
	}
	if archetype.MinElixir > archetype.MaxElixir {
		return fmt.Errorf("min_elixir (%.1f) cannot be greater than max_elixir (%.1f)",
			archetype.MinElixir, archetype.MaxElixir)
	}
	return nil
}

// SaveDefaultArchetypes saves the default embedded archetypes to a file
// This allows users to extract the default configuration as a starting point for customization
func SaveDefaultArchetypes(destPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the embedded JSON data directly
	if err := os.WriteFile(destPath, defaultArchetypesJSON, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
