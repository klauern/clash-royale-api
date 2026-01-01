package deck

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// StrategyConfig defines the parameters for a deck building strategy
type StrategyConfig struct {
	// TargetElixirMin is the minimum target average elixir cost
	TargetElixirMin float64

	// TargetElixirMax is the maximum target average elixir cost
	TargetElixirMax float64

	// RoleMultipliers defines scoring multipliers for each card role
	// A multiplier > 1.0 increases preference, < 1.0 decreases preference
	// DEPRECATED: Use RoleBonuses instead for better strategy differentiation
	RoleMultipliers map[CardRole]float64

	// RoleBonuses provides additive score adjustments for each card role
	// Positive values (0 to +0.5) increase preference, negative values (-0.5 to 0) decrease preference
	// This is level-agnostic, unlike multipliers which scale with card level
	// Additive bonuses allow on-strategy cards to compete regardless of level
	RoleBonuses map[CardRole]float64

	// ArchetypeAffinity provides extra bonuses for specific cards that naturally fit the strategy
	// This helps archetype-appropriate cards compete with higher-level off-archetype cards
	// Key: card name, Value: bonus score (typically +0.15 to +0.30)
	ArchetypeAffinity map[string]float64

	// CompositionOverrides allows forcing specific role counts
	// nil map means use default composition logic
	CompositionOverrides *CompositionOverride
}

// CompositionOverride specifies forced counts for specific roles
type CompositionOverride struct {
	WinConditions *int // Pointer allows nil = use default
	Buildings     *int
	BigSpells     *int
	SmallSpells   *int
	Support       *int
	Cycle         *int
}

// strategyConfigJSON is the JSON representation of a single strategy
type strategyConfigJSON struct {
	Name               string             `json:"name"`
	Description        string             `json:"description"`
	TargetElixirMin    float64            `json:"target_elixir_min"`
	TargetElixirMax    float64            `json:"target_elixir_max"`
	RoleMultipliers    map[string]float64 `json:"role_multipliers"`
	RoleBonuses        map[string]float64 `json:"role_bonuses"`
	ArchetypeAffinity  map[string]float64 `json:"archetype_affinity"`
	CompositionOverrides *compositionOverrideJSON `json:"composition_overrides,omitempty"`
}

// compositionOverrideJSON is the JSON representation of composition overrides
type compositionOverrideJSON struct {
	WinConditions *int `json:"win_conditions,omitempty"`
	Buildings     *int `json:"buildings,omitempty"`
	BigSpells     *int `json:"big_spells,omitempty"`
	SmallSpells   *int `json:"small_spells,omitempty"`
	Support       *int `json:"support,omitempty"`
	Cycle         *int `json:"cycle,omitempty"`
}

// strategiesFile is the root JSON structure
type strategiesFile struct {
	Strategies map[string]strategyConfigJSON `json:"strategies"`
}

var (
	// strategyCache holds the loaded strategies
	strategyCache     map[string]StrategyConfig
	strategyCacheMu   sync.RWMutex
	strategyCacheOnce sync.Once
)

// roleFromString converts a JSON string role to a CardRole
func roleFromString(s string) (CardRole, error) {
	switch strings.ToLower(strings.ReplaceAll(s, "_", "")) {
	case "wincondition", "win_condition":
		return RoleWinCondition, nil
	case "building":
		return RoleBuilding, nil
	case "spellbig", "spell_big":
		return RoleSpellBig, nil
	case "spellsmall", "spell_small":
		return RoleSpellSmall, nil
	case "support":
		return RoleSupport, nil
	case "cycle":
		return RoleCycle, nil
	default:
		return "", fmt.Errorf("unknown role: %s", s)
	}
}

// loadStrategyConfigFile loads strategies from the JSON file
func loadStrategyConfigFile() (map[string]StrategyConfig, error) {
	// Get config file path from environment or use default
	configPath := os.Getenv("STRATEGIES_CONFIG_PATH")
	if configPath == "" {
		// Default to config/strategies.json relative to current working directory
		configPath = "config/strategies.json"
	}

	// Try absolute path first, then relative to working directory
	if !filepath.IsAbs(configPath) {
		// Check if running from the project root
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Try relative to the deck package (for tests)
			configPath = filepath.Join("..", "..", configPath)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read strategies config from %s: %w", configPath, err)
	}

	var file strategiesFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("failed to parse strategies config: %w", err)
	}

	strategies := make(map[string]StrategyConfig)
	for key, jsonConfig := range file.Strategies {
		config := StrategyConfig{
			TargetElixirMin:   jsonConfig.TargetElixirMin,
			TargetElixirMax:   jsonConfig.TargetElixirMax,
			RoleMultipliers:   make(map[CardRole]float64),
			RoleBonuses:       make(map[CardRole]float64),
			ArchetypeAffinity: jsonConfig.ArchetypeAffinity,
		}

		// Convert role multipliers
		for roleStr, val := range jsonConfig.RoleMultipliers {
			role, err := roleFromString(roleStr)
			if err != nil {
				return nil, fmt.Errorf("invalid role multiplier %q in strategy %q: %w", roleStr, key, err)
			}
			config.RoleMultipliers[role] = val
		}

		// Convert role bonuses
		for roleStr, val := range jsonConfig.RoleBonuses {
			role, err := roleFromString(roleStr)
			if err != nil {
				return nil, fmt.Errorf("invalid role bonus %q in strategy %q: %w", roleStr, key, err)
			}
			config.RoleBonuses[role] = val
		}

		// Convert composition overrides if present
		if jsonConfig.CompositionOverrides != nil {
			config.CompositionOverrides = &CompositionOverride{
				WinConditions: jsonConfig.CompositionOverrides.WinConditions,
				Buildings:     jsonConfig.CompositionOverrides.Buildings,
				BigSpells:     jsonConfig.CompositionOverrides.BigSpells,
				SmallSpells:   jsonConfig.CompositionOverrides.SmallSpells,
				Support:       jsonConfig.CompositionOverrides.Support,
				Cycle:         jsonConfig.CompositionOverrides.Cycle,
			}
		}

		// Validate the strategy
		if err := validateStrategyConfig(&config); err != nil {
			return nil, fmt.Errorf("invalid strategy %q: %w", key, err)
		}

		strategies[key] = config
	}

	return strategies, nil
}

// validateStrategyConfig validates a strategy configuration
func validateStrategyConfig(config *StrategyConfig) error {
	if config.TargetElixirMin < 0 || config.TargetElixirMin > 10 {
		return fmt.Errorf("target_elixir_min must be between 0 and 10, got %f", config.TargetElixirMin)
	}
	if config.TargetElixirMax < 0 || config.TargetElixirMax > 10 {
		return fmt.Errorf("target_elixir_max must be between 0 and 10, got %f", config.TargetElixirMax)
	}
	if config.TargetElixirMin > config.TargetElixirMax {
		return fmt.Errorf("target_elixir_min (%f) must be <= target_elixir_max (%f)", config.TargetElixirMin, config.TargetElixirMax)
	}
	return nil
}

// loadStrategyCache loads the strategy cache (thread-safe singleton)
func loadStrategyCache() map[string]StrategyConfig {
	strategyCacheOnce.Do(func() {
		cache, err := loadStrategyConfigFile()
		if err != nil {
			// If loading fails, log the error and provide default balanced strategy
			fmt.Fprintf(os.Stderr, "Warning: failed to load strategies config: %v\n", err)
			fmt.Fprintf(os.Stderr, "Using default balanced strategy only.\n")
			cache = map[string]StrategyConfig{
				"balanced": {
					TargetElixirMin: 3.0,
					TargetElixirMax: 3.5,
					RoleMultipliers: map[CardRole]float64{
						RoleWinCondition: 1.0,
						RoleBuilding:     1.0,
						RoleSpellBig:     1.0,
						RoleSpellSmall:   1.0,
						RoleSupport:      1.0,
						RoleCycle:        1.0,
					},
					RoleBonuses: map[CardRole]float64{
						RoleWinCondition: 0.0,
						RoleBuilding:     0.0,
						RoleSpellBig:     0.0,
						RoleSpellSmall:   0.0,
						RoleSupport:      0.0,
						RoleCycle:        0.0,
					},
					ArchetypeAffinity: map[string]float64{},
				},
			}
		}
		strategyCacheMu.Lock()
		strategyCache = cache
		strategyCacheMu.Unlock()
	})
	return strategyCache
}

// ReloadStrategyConfig reloads the strategy configuration from disk
// This enables hot-reloading for development without restarting the application
func ReloadStrategyConfig() error {
	cache, err := loadStrategyConfigFile()
	if err != nil {
		return err
	}
	strategyCacheMu.Lock()
	strategyCache = cache
	strategyCacheMu.Unlock()
	return nil
}

// GetStrategyConfig returns the configuration for a given strategy
func GetStrategyConfig(strategy Strategy) StrategyConfig {
	cache := loadStrategyCache()

	strategyCacheMu.RLock()
	defer strategyCacheMu.RUnlock()

	config, exists := cache[string(strategy)]
	if !exists {
		// Return balanced strategy as fallback
		config = cache["balanced"]
	}
	return config
}

// GetStrategyScaling returns global strategy bonus scaling from environment variable.
// STRATEGY_BONUS_SCALE=1.0 (default), 0.0 to disable, 2.0 for extreme differentiation.
// This allows runtime tuning of strategy effectiveness without code changes.
func GetStrategyScaling() float64 {
	if scaleStr := os.Getenv("STRATEGY_BONUS_SCALE"); scaleStr != "" {
		if scale, err := strconv.ParseFloat(scaleStr, 64); err == nil {
			// Clamp to reasonable range (0.0 to 2.0)
			if scale < 0 {
				return 0
			}
			if scale > 2.0 {
				return 2.0
			}
			return scale
		}
	}
	return 1.0 // Default scaling
}
