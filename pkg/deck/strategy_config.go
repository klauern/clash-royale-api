package deck

// StrategyConfig defines the parameters for a deck building strategy
type StrategyConfig struct {
	// TargetElixirMin is the minimum target average elixir cost
	TargetElixirMin float64

	// TargetElixirMax is the maximum target average elixir cost
	TargetElixirMax float64

	// RoleMultipliers defines scoring multipliers for each card role
	// A multiplier > 1.0 increases preference, < 1.0 decreases preference
	RoleMultipliers map[CardRole]float64

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
		return StrategyConfig{
			TargetElixirMin: 3.5,
			TargetElixirMax: 4.0,
			RoleMultipliers: map[CardRole]float64{
				RoleWinCondition: 1.3,  // Strongly favor win conditions
				RoleSupport:      1.1,  // Slightly favor support
				RoleBuilding:     1.0,
				RoleSpellBig:     1.0,
				RoleSpellSmall:   1.0,
				RoleCycle:        1.0,
			},
		}

	case StrategyControl:
		return StrategyConfig{
			TargetElixirMin: 3.5,
			TargetElixirMax: 4.2,
			RoleMultipliers: map[CardRole]float64{
				RoleBuilding:     1.3,  // Favor defensive buildings
				RoleSpellBig:     1.2,  // Favor big spells
				RoleWinCondition: 1.0,
				RoleSupport:      1.0,
				RoleSpellSmall:   1.0,
				RoleCycle:        1.0,
			},
		}

	case StrategyCycle:
		return StrategyConfig{
			TargetElixirMin: 2.5,
			TargetElixirMax: 3.0,
			RoleMultipliers: map[CardRole]float64{
				RoleCycle:        1.4,  // Strongly favor cycle cards
				RoleWinCondition: 1.0,
				RoleSupport:      1.0,
				RoleBuilding:     1.0,
				RoleSpellBig:     0.8,  // Slightly disfavor big spells (high cost)
				RoleSpellSmall:   1.1,  // Favor small spells
			},
		}

	case StrategySplash:
		return StrategyConfig{
			TargetElixirMin: 3.2,
			TargetElixirMax: 3.8,
			RoleMultipliers: map[CardRole]float64{
				RoleSupport:      1.3,  // Favor splash support troops
				RoleWinCondition: 1.0,
				RoleBuilding:     1.0,
				RoleSpellBig:     1.0,
				RoleSpellSmall:   1.0,
				RoleCycle:        1.0,
			},
		}

	case StrategySpell:
		// Spell strategy has composition overrides
		bigSpells := 2
		buildings := 0
		smallSpells := 1

		return StrategyConfig{
			TargetElixirMin: 3.2,
			TargetElixirMax: 3.8,
			RoleMultipliers: map[CardRole]float64{
				RoleSpellBig:     1.5,  // Strongly favor big spells
				RoleSpellSmall:   1.2,  // Favor small spells
				RoleWinCondition: 1.0,
				RoleSupport:      1.0,
				RoleBuilding:     0.5,  // Disfavor buildings (override to 0)
				RoleCycle:        1.0,
			},
			CompositionOverrides: &CompositionOverride{
				BigSpells:   &bigSpells,
				Buildings:   &buildings,
				SmallSpells: &smallSpells,
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
		}
	}
}
