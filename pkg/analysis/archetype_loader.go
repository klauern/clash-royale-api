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

	// Validate archetypes
	for i, archetype := range config.Archetypes {
		if err := validateArchetype(&archetype); err != nil {
			return nil, fmt.Errorf("archetype #%d (%s) is invalid: %w", i, archetype.Name, err)
		}
	}

	return config.Archetypes, nil
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

	// Pretty-print JSON for easy editing
	data, err := json.MarshalIndent(ArchetypeConfig{
		Version:    1,
		Archetypes: defaultEmbeddedArchetypes(),
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// defaultEmbeddedArchetypes returns the hardcoded default archetypes
// This maintains backward compatibility and provides embedded defaults
func defaultEmbeddedArchetypes() []DeckArchetypeTemplate {
	return []DeckArchetypeTemplate{
		// === Beatdown Archetypes ===
		{
			Name:         "Golem Beatdown",
			WinCondition: "Golem",
			SupportCards: []string{"Baby Dragon", "Night Witch", "Mega Minion", "Lumberjack", "Lightning", "Tornado"},
			MinElixir:    3.5,
			MaxElixir:    4.5,
		},
		{
			Name:         "LavaLoon Beatdown",
			WinCondition: "Lava Hound",
			SupportCards: []string{"Balloon", "Inferno Dragon", "Lumberjack", "Bowler", "Freeze", "Tornado"},
			MinElixir:    3.8,
			MaxElixir:    4.8,
		},
		{
			Name:         "Giant Beatdown",
			WinCondition: "Giant",
			SupportCards: []string{"Witch", "Musketeer", "Prince", "Mega Minion", "Fireball", "Elixir Collector"},
			MinElixir:    3.2,
			MaxElixir:    4.2,
		},
		{
			Name:         "Electro Giant Beatdown",
			WinCondition: "Electro Giant",
			SupportCards: []string{"Little Prince", "Bowler", "Knight", "Archers", "Cannon", "Lightning", "Tornado"},
			MinElixir:    3.8,
			MaxElixir:    5.0,
		},

		// === Cycle Archetypes ===
		{
			Name:         "Hog Rider Cycle",
			WinCondition: "Hog Rider",
			SupportCards: []string{"Skeletons", "Cannon", "Ice Golem", "Musketeer", "Ice Spirit", "Fireball"},
			MinElixir:    2.4,
			MaxElixir:    3.2,
		},
		{
			Name:         "Royal Giant Cycle",
			WinCondition: "Royal Giant",
			SupportCards: []string{"Fireball", "Lightning", "Cannon", "Furnace", "Royal Ghost", "Skeletons"},
			MinElixir:    2.8,
			MaxElixir:    3.6,
		},
		{
			Name:         "Miner Poison Cycle",
			WinCondition: "Miner",
			SupportCards: []string{"Poison", "Electro Wizard", "Valkyrie", "Ice Golem", "Fireball"},
			MinElixir:    2.6,
			MaxElixir:    3.4,
		},
		{
			Name:         "Miner Balloon Cycle",
			WinCondition: "Miner",
			SupportCards: []string{"Balloon", "Musketeer", "Skeletons", "Ice Golem", "Bomb Tower"},
			MinElixir:    2.8,
			MaxElixir:    3.6,
		},
		{
			Name:         "Royal Hogs Cycle",
			WinCondition: "Royal Hogs",
			SupportCards: []string{"Royal Recruits", "Flying Machine", "Goblin Cage", "Fireball", "Arrows"},
			MinElixir:    2.8,
			MaxElixir:    3.6,
		},

		// === Siege Archetypes ===
		{
			Name:         "X-Bow Siege",
			WinCondition: "X-Bow",
			SupportCards: []string{"Tesla", "Archers", "Knight", "Skeletons", "Electro Spirit", "Fireball"},
			MinElixir:    2.8,
			MaxElixir:    3.6,
		},
		{
			Name:         "Mortar Siege",
			WinCondition: "Mortar",
			SupportCards: []string{"Skeleton Barrel", "Cannon Cart", "Royal Chef", "Knight", "Arrows"},
			MinElixir:    2.6,
			MaxElixir:    3.4,
		},

		// === Bridge Spam Archetypes ===
		{
			Name:         "PEKKA Bridge Spam",
			WinCondition: "P.E.K.K.A",
			SupportCards: []string{"Battle Ram", "Bandit", "Royal Ghost", "Electro Wizard", "Minions", "Poison"},
			MinElixir:    3.0,
			MaxElixir:    4.0,
		},
		{
			Name:         "Mega Knight Bridge Spam",
			WinCondition: "Mega Knight",
			SupportCards: []string{"Wall Breakers", "Archer Queen", "Prince", "Bandit", "Goblin Gang"},
			MinElixir:    3.2,
			MaxElixir:    4.2,
		},
		{
			Name:         "Royal Ghost Bridge Spam",
			WinCondition: "Royal Ghost",
			SupportCards: []string{"Battle Ram", "Ice Golem", "Night Witch", "Inferno Dragon", "Guards"},
			MinElixir:    2.8,
			MaxElixir:    3.8,
		},

		// === Bait Archetypes ===
		{
			Name:         "Log Bait",
			WinCondition: "Goblin Barrel",
			SupportCards: []string{"Knight", "Princess", "Goblin Gang", "Ice Spirit", "Inferno Tower", "Rocket"},
			MinElixir:    2.6,
			MaxElixir:    3.4,
		},
		{
			Name:         "Evo Dart Goblin Bait",
			WinCondition: "Goblin Barrel",
			SupportCards: []string{"Dart Goblin", "Knight", "Princess", "Goblin Gang", "Ice Spirit", "Cannon"},
			MinElixir:    2.4,
			MaxElixir:    3.2,
		},
		{
			Name:         "Evo Recruits Barrel Bait",
			WinCondition: "Goblin Barrel",
			SupportCards: []string{"Royal Recruits", "Goblinstein", "Cannon Cart", "Dart Goblin", "Goblin Gang"},
			MinElixir:    2.8,
			MaxElixir:    3.6,
		},

		// === Control/Graveyard Archetypes ===
		{
			Name:         "Graveyard Freeze",
			WinCondition: "Graveyard",
			SupportCards: []string{"Ice Wizard", "Baby Dragon", "Bomb Tower", "Poison", "Tornado", "Knight"},
			MinElixir:    3.0,
			MaxElixir:    4.0,
		},
		{
			Name:         "Splashyard",
			WinCondition: "Graveyard",
			SupportCards: []string{"Knight", "Ice Wizard", "Baby Dragon", "Bomb Tower", "Poison", "Tornado"},
			MinElixir:    2.8,
			MaxElixir:    3.8,
		},
		{
			Name:         "Evo Witch Giant Graveyard",
			WinCondition: "Giant",
			SupportCards: []string{"Graveyard", "Witch", "Giant Snowball", "Bowler", "Minions", "Guards"},
			MinElixir:    3.2,
			MaxElixir:    4.2,
		},

		// === Three Musketeers Archetypes ===
		{
			Name:         "3M Bridge Spam",
			WinCondition: "Three Musketeers",
			SupportCards: []string{"Battle Ram", "Royal Ghost", "Bandit", "Ice Golem", "Elixir Collector"},
			MinElixir:    3.2,
			MaxElixir:    4.2,
		},
		{
			Name:         "Giant Muskets",
			WinCondition: "Giant",
			SupportCards: []string{"Three Musketeers", "Bats", "Battle Ram", "Minion Horde", "Elixir Collector"},
			MinElixir:    3.5,
			MaxElixir:    4.5,
		},

		// === Goblin Drill Archetypes ===
		{
			Name:         "Goblin Drill Evo Valk",
			WinCondition: "Goblin Drill",
			SupportCards: []string{"Valkyrie", "Skeletons", "Magic Archer", "Spear Goblins", "Bomb Tower", "Tornado"},
			MinElixir:    2.8,
			MaxElixir:    3.6,
		},

		// === Skeleton King Archetypes ===
		{
			Name:         "Evo Mortar Miner Skeleton King",
			WinCondition: "Mortar",
			SupportCards: []string{"Skeleton King", "Miner", "Bats", "Cannon Cart", "Goblin Gang"},
			MinElixir:    2.8,
			MaxElixir:    3.6,
		},
		{
			Name:         "Evo Royal Giant Archers Skeleton King",
			WinCondition: "Royal Giant",
			SupportCards: []string{"Skeleton King", "Archers", "Fisherman", "Mother Witch", "Tombstone"},
			MinElixir:    2.8,
			MaxElixir:    3.8,
		},

		// === Double Dragon Archetypes ===
		{
			Name:         "Double Elixir Loon Freeze",
			WinCondition: "Balloon",
			SupportCards: []string{"Electro Dragon", "Inferno Dragon", "Lumberjack", "Bowler", "Freeze", "Tornado"},
			MinElixir:    3.5,
			MaxElixir:    4.5,
		},

		// === Ram Rider Archetypes ===
		{
			Name:         "Golden Knight Royal Hogs",
			WinCondition: "Royal Hogs",
			SupportCards: []string{"Golden Knight", "Royal Recruits", "Zappies", "Flying Machine", "Goblin Cage"},
			MinElixir:    2.8,
			MaxElixir:    3.8,
		},
	}
}
