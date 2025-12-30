package deck

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sync"
)

// LevelCurve manages per-card level scaling curves for more accurate deck building.
// It replaces the uniform linear levelRatio with card-specific exponential curves
// based on community research showing approximately 10% growth per level.
type LevelCurve struct {
	config     CardLevelCurvesConfig
	cardCache  map[string]CardLevelConfig
	cacheMutex sync.RWMutex
}

// CardLevelCurvesConfig is the top-level configuration structure
type CardLevelCurvesConfig struct {
	CardLevelCurves map[string]CardLevelConfig `json:"cardLevelCurves"`
}

// CardLevelConfig defines the curve parameters for a specific card
type CardLevelConfig struct {
	// Curve parameters (values are percentages, e.g., 100 = 1.0x)
	BaseScale   float64 `json:"baseScale"`   // Base scaling factor (typically 100)
	GrowthRate  float64 `json:"growthRate"`  // Per-level growth rate (typically 0.10)
	MinScale    float64 `json:"minScale"`    // Minimum multiplier (level 0) - often 0.0
	MaxScale    float64 `json:"maxScale"`    // Maximum multiplier (level 15+) - for clamping

	// Special adjustments
	Type        string  `json:"type"`        // "standard", "spell_duration", "tower", "champion"
	RarityBonus float64 `json:"rarityBonus"` // Additional boost for rarity (0.0 = none, 0.05 = +5%)

	// Significant deviations from standard curve
	LevelOverrides map[string]float64 `json:"levelOverrides"` // Specific level overrides
}

// Default configuration for cards without specific settings
var defaultCardLevelConfig = CardLevelConfig{
	BaseScale:   100.0,
	GrowthRate:  0.10,
	MinScale:    0.0,
	MaxScale:    400.0,
	Type:        "standard",
	RarityBonus: 0.0,
}

// NewLevelCurve creates a new LevelCurve instance with loaded configuration
func NewLevelCurve(configPath string) (*LevelCurve, error) {
	lc := &LevelCurve{
		cardCache: make(map[string]CardLevelConfig),
	}

	if err := lc.loadConfig(configPath); err != nil {
		return nil, fmt.Errorf("failed to load level curve config: %w", err)
	}

	return lc, nil
}

// loadConfig loads the level curves configuration from a JSON file
func (lc *LevelCurve) loadConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If config doesn't exist, use default configuration
			lc.config = CardLevelCurvesConfig{
				CardLevelCurves: make(map[string]CardLevelConfig),
			}
			// Set default entry
			lc.config.CardLevelCurves["_default"] = defaultCardLevelConfig
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, &lc.config); err != nil {
		return fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Ensure default configuration exists
	if _, exists := lc.config.CardLevelCurves["_default"]; !exists {
		lc.config.CardLevelCurves["_default"] = defaultCardLevelConfig
	}

	return nil
}

// getConfig retrieves the configuration for a specific card
func (lc *LevelCurve) getConfig(cardName string) CardLevelConfig {
	// Check cache first
	lc.cacheMutex.RLock()
	if config, exists := lc.cardCache[cardName]; exists {
		lc.cacheMutex.RUnlock()
		return config
	}
	lc.cacheMutex.RUnlock()

	// Get config from loaded configuration
	var config CardLevelConfig
	if cardConfig, exists := lc.config.CardLevelCurves[cardName]; exists {
		config = cardConfig
	} else if cardConfig, exists := lc.config.CardLevelCurves["_default"]; exists {
		config = cardConfig
	} else {
		config = defaultCardLevelConfig
	}

	// Fill in missing fields with defaults
	config = lc.fillMissingFields(config)

	// Cache the result
	lc.cacheMutex.Lock()
	lc.cardCache[cardName] = config
	lc.cacheMutex.Unlock()

	return config
}

// fillMissingFields ensures all required fields have values
func (lc *LevelCurve) fillMissingFields(config CardLevelConfig) CardLevelConfig {
	defaultConfig := defaultCardLevelConfig

	if config.BaseScale == 0 {
		config.BaseScale = defaultConfig.BaseScale
	}
	if config.GrowthRate == 0 {
		config.GrowthRate = defaultConfig.GrowthRate
	}
	if config.Type == "" {
		config.Type = defaultConfig.Type
	}

	return config
}

// GetLevelMultiplier returns the multiplier for a card at a specific level
// The multiplier is a value between 0 and ~4.0, representing how much stronger
// the card is compared to its level 1 stats
func (lc *LevelCurve) GetLevelMultiplier(cardName string, level int) float64 {
	if level <= 0 {
		return 0.0
	}

	config := lc.getConfig(cardName)

	// Check for specific level overrides
	if config.LevelOverrides != nil {
		if levelStr := fmt.Sprintf("%d", level); levelStr != "" {
			if override, exists := config.LevelOverrides[levelStr]; exists {
				// Convert percentage to multiplier (e.g., 212 = 2.12x)
				return override / 100.0
			}
		}
	}

	// Calculate using exponential curve formula
	// multiplier = baseScale * (1 + growthRate)^(level-1) * (1 + rarityBonus)
	baseMultiplier := config.BaseScale * math.Pow(1+config.GrowthRate, float64(level-1))
	adjustedMultiplier := baseMultiplier * (1 + config.RarityBonus)

	// Apply min/max clamping if specified
	if config.MinScale > 0 && adjustedMultiplier < config.MinScale {
		adjustedMultiplier = config.MinScale
	}
	if config.MaxScale > 0 && adjustedMultiplier > config.MaxScale {
		adjustedMultiplier = config.MaxScale
	}

	// Convert percentage to multiplier
	return adjustedMultiplier / 100.0
}

// GetRelativeLevelRatio returns the level ratio compared to max level for a card
// This is the replacement for the simple linear ratio: level / maxLevel
func (lc *LevelCurve) GetRelativeLevelRatio(cardName string, level, maxLevel int) float64 {
	if maxLevel <= 0 {
		return 0.0
	}

	currentMultiplier := lc.GetLevelMultiplier(cardName, level)
	maxMultiplier := lc.GetLevelMultiplier(cardName, maxLevel)

	if maxMultiplier <= 0 {
		return 0.0
	}

	return currentMultiplier / maxMultiplier
}

// ValidateCard validates a card's level curve against known data points
func (lc *LevelCurve) ValidateCard(cardName string, validationPoints map[int]float64) error {
	for level, expectedMultiplier := range validationPoints {
		actualMultiplier := lc.GetLevelMultiplier(cardName, level)

		// Allow 2% tolerance for floating point and rounding differences
		tolerance := expectedMultiplier * 0.02
		if math.Abs(actualMultiplier-expectedMultiplier) > tolerance {
			return fmt.Errorf("level %d: expected multiplier %.3f, got %.3f (diff: %.3f)",
				level, expectedMultiplier, actualMultiplier,
				math.Abs(actualMultiplier-expectedMultiplier))
		}
	}

	return nil
}
