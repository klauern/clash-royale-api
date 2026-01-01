package deck

import (
	"os"
	"strconv"
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

// GetStrategyConfig returns the configuration for a given strategy
func GetStrategyConfig(strategy Strategy) StrategyConfig {
	switch strategy {
	case StrategyAggro:
		// Aggro strategy: 2 win conditions, 0 buildings (pure offense)
		winConditions := 2
		buildings := 0
		support := 3
		cycle := 1

		return StrategyConfig{
			TargetElixirMin: 3.5,
			TargetElixirMax: 4.0,
			RoleMultipliers: map[CardRole]float64{
				RoleWinCondition: 2.0, // Strongly favor win conditions
				RoleSupport:      1.2, // Favor support for offensive pressure
				RoleCycle:        1.0,
				RoleSpellBig:     1.0,
				RoleSpellSmall:   1.0,
				RoleBuilding:     0.3, // Disfavor defensive buildings
			},
			RoleBonuses: map[CardRole]float64{
				RoleWinCondition: 0.40,  // Strong favor
				RoleSupport:      0.15,  // Moderate favor
				RoleCycle:        0.0,   // Neutral
				RoleSpellBig:     0.0,   // Neutral
				RoleSpellSmall:   0.0,   // Neutral
				RoleBuilding:    -0.30,  // Strong disfavor
			},
			ArchetypeAffinity: map[string]float64{
				// Aggressive win conditions (fast, direct damage)
				"Hog Rider":         0.25,
				"Royal Hogs":        0.25,
				"Ram Rider":         0.25,
				"Battle Ram":        0.25,
				"Balloon":           0.25,
				"Elite Barbarians":  0.20,
				"Royal Giant":       0.20,
				// Aggressive support (high damage, offensive pressure)
				"Mini PEKKA":        0.20,
				"Prince":            0.20,
				"Dark Prince":       0.20,
				"Lumberjack":        0.20,
				"Bandit":            0.20,
				"Electro Giant":     0.15,
				"Goblin Giant":      0.15,
				// Offensive spells
				"Rage":              0.15,
				"Clone":             0.15,
			},
			CompositionOverrides: &CompositionOverride{
				WinConditions: &winConditions,
				Buildings:     &buildings,
				Support:       &support,
				Cycle:         &cycle,
			},
		}

	case StrategyControl:
		// Control strategy: 2 buildings, 2 big spells, 0 small spells (defensive grind)
		buildings := 2
		bigSpells := 2
		smallSpells := 0
		support := 2
		cycle := 1

		return StrategyConfig{
			TargetElixirMin: 3.5,
			TargetElixirMax: 4.2,
			RoleMultipliers: map[CardRole]float64{
				RoleBuilding:     2.0, // Strongly favor defensive buildings
				RoleSpellBig:     1.5, // Favor big spells for area control
				RoleSpellSmall:   0.3, // Disfavor small spells
				RoleSupport:      1.0,
				RoleCycle:        0.5, // Disfavor cheap cycle cards
				RoleWinCondition: 0.5, // Disfavor pure offensive win conditions
			},
			RoleBonuses: map[CardRole]float64{
				RoleBuilding:     0.40,  // Strong favor
				RoleSpellBig:     0.25,  // Moderate favor
				RoleSpellSmall:  -0.20,  // Disfavor
				RoleSupport:      0.10,  // Slight favor
				RoleCycle:       -0.15,  // Disfavor
				RoleWinCondition: -0.30, // Strong disfavor
			},
			ArchetypeAffinity: map[string]float64{
				// Defensive/siege win conditions
				"X-Bow":            0.25,
				"Mortar":           0.25,
				"Graveyard":        0.20,
				"Miner":            0.20,
				// Defensive buildings
				"Inferno Tower":    0.25,
				"Tesla":            0.25,
				"Bomb Tower":       0.20,
				"Tombstone":        0.20,
				"Furnace":          0.20,
				// Defensive support (area control, slowing)
				"Ice Wizard":       0.25,
				"Electro Wizard":   0.25,
				"Bowler":           0.20,
				"Executioner":      0.20,
				"Tornado":          0.25,
				"Freeze":           0.15,
				// Area damage spells
				"Poison":           0.20,
				"Earthquake":       0.20,
				"Rocket":           0.15,
			},
			CompositionOverrides: &CompositionOverride{
				Buildings:   &buildings,
				BigSpells:   &bigSpells,
				SmallSpells: &smallSpells,
				Support:     &support,
				Cycle:       &cycle,
			},
		}

	case StrategyCycle:
		// Cycle strategy: 4 cycle cards, 0 big spells (fast rotation)
		cycle := 4
		bigSpells := 0
		support := 1

		return StrategyConfig{
			TargetElixirMin: 2.5,
			TargetElixirMax: 3.0,
			RoleMultipliers: map[CardRole]float64{
				RoleCycle:        2.0, // Strongly favor cycle cards
				RoleSpellSmall:   1.2, // Favor small spells
				RoleWinCondition: 1.0,
				RoleSupport:      1.0,
				RoleBuilding:     1.0,
				RoleSpellBig:     0.3, // Strongly disfavor big spells (high cost)
			},
			RoleBonuses: map[CardRole]float64{
				RoleCycle:        0.35,  // Strong favor
				RoleSpellSmall:   0.15,  // Moderate favor
				RoleWinCondition: 0.10,  // Slight favor
				RoleSupport:      0.0,   // Neutral
				RoleBuilding:     0.0,   // Neutral
				RoleSpellBig:    -0.30,  // Strong disfavor
			},
			ArchetypeAffinity: map[string]float64{
				// Low-cost cycle cards
				"Skeletons":      0.30,
				"Ice Spirit":     0.30,
				"Electro Spirit": 0.30,
				"Fire Spirit":    0.30,
				"Bats":           0.25,
				"Spear Goblins":  0.25,
				// Fast cycle support
				"Archers":        0.20,
				"Knight":         0.20,
				"Ice Golem":      0.20,
				// Low-cost spells
				"Zap":            0.20,
				"Log":            0.20,
				"Snowball":       0.20,
			},
			CompositionOverrides: &CompositionOverride{
				BigSpells: &bigSpells,
				Cycle:     &cycle,
				Support:   &support,
			},
		}

	case StrategySplash:
		// Splash strategy: 3 splash support cards (area damage focus)
		support := 3
		cycle := 1

		return StrategyConfig{
			TargetElixirMin: 3.2,
			TargetElixirMax: 3.8,
			RoleMultipliers: map[CardRole]float64{
				RoleSupport:      2.0, // Strongly favor splash support troops
				RoleSpellBig:     1.2, // Favor big splash spells
				RoleWinCondition: 1.0,
				RoleBuilding:     1.0,
				RoleSpellSmall:   1.0,
				RoleCycle:        0.5, // Disfavor cheap cycle cards
			},
			RoleBonuses: map[CardRole]float64{
				RoleSupport:      0.40,  // Strong favor
				RoleSpellBig:     0.15,  // Moderate favor
				RoleWinCondition: 0.0,   // Neutral
				RoleBuilding:    -0.10,  // Slight disfavor
				RoleSpellSmall:   0.0,   // Neutral
				RoleCycle:       -0.15,  // Disfavor
			},
			ArchetypeAffinity: map[string]float64{
				// Splash support troops
				"Baby Dragon":       0.25,
				"Valkyrie":          0.25,
				"Wizard":            0.25,
				"Executioner":       0.25,
				"Bowler":            0.25,
				"Bomber":            0.25,
				"Skeleton Dragons":  0.20,
				"Fire Spirit":       0.20,
				"Phoenix":           0.20,
				// Splash spells
				"Fireball":          0.15,
				"Poison":            0.15,
				"Lightning":         0.15,
			},
			CompositionOverrides: &CompositionOverride{
				Support: &support,
				Cycle:   &cycle,
			},
		}

	case StrategySpell:
		// Spell strategy has composition overrides (2 big spells, 0 buildings, 1 small spell)
		bigSpells := 2
		buildings := 0
		smallSpells := 1
		support := 3
		cycle := 1

		return StrategyConfig{
			TargetElixirMin: 3.2,
			TargetElixirMax: 3.8,
			RoleMultipliers: map[CardRole]float64{
				RoleSpellBig:     2.0, // Strongly favor big spells
				RoleSpellSmall:   1.5, // Favor small spells
				RoleWinCondition: 1.0,
				RoleSupport:      1.0,
				RoleBuilding:     0.1, // Strongly disfavor buildings (override to 0)
				RoleCycle:        1.0,
			},
			RoleBonuses: map[CardRole]float64{
				RoleSpellBig:     0.40,  // Strong favor
				RoleSpellSmall:   0.25,  // Moderate favor
				RoleWinCondition: 0.10,  // Slight favor
				RoleSupport:      0.10,  // Slight favor
				RoleBuilding:    -0.35,  // Strong disfavor
				RoleCycle:        0.0,   // Neutral
			},
			ArchetypeAffinity: map[string]float64{
				// Direct damage spells
				"Rocket":      0.25,
				"Lightning":   0.25,
				"Fireball":    0.25,
				"Poison":      0.20,
				"Earthquake":  0.20,
				// Small spells
				"Log":         0.20,
				"Zap":         0.20,
				"Snowball":    0.20,
				"Arrows":      0.20,
				// Spell-synergy support
				"Miner":       0.20, // Pairs well with spell chip
				"Goblin Barrel": 0.15, // Spell bait element
			},
			CompositionOverrides: &CompositionOverride{
				BigSpells:   &bigSpells,
				Buildings:   &buildings,
				SmallSpells: &smallSpells,
				Support:     &support,
				Cycle:       &cycle,
			},
		}

	case StrategyBalanced:
		fallthrough
	default:
		return StrategyConfig{
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
			ArchetypeAffinity: map[string]float64{}, // No archetype preference for balanced
		}
	}
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
