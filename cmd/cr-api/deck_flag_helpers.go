package main

import "github.com/urfave/cli/v3"

const (
	defaultEvolutionSlots      = 2
	defaultCombatStatsWeight   = 0.25
	defaultSynergyWeight       = 0.15
	defaultUniquenessWeight    = 0.2
	defaultFuzzWeight          = 0.10
	defaultFuzzDeckLimit       = 100
	combatStatsWeightUsage     = "Weight for combat stats in scoring (0.0-1.0, where 0=disabled, 0.25=default, 1.0=combat-only)"
	disableCombatStatsUsage    = "Disable combat stats completely (use traditional scoring only)"
	enableSynergyUsage         = "Enable synergy-based card selection (considers card interactions and combos)"
	synergyWeightUsage         = "Weight for synergy scoring (0.0-1.0, default 0.15 = 15%)"
	preferUniqueUsage          = "Enable uniqueness/anti-meta scoring (prefers less common cards)"
	uniquenessWeightUsage      = "Weight for uniqueness scoring (0.0-0.3, default 0.2 = 20%)"
	avoidArchetypeUsage        = "Archetypes to avoid when building decks (e.g., beatdown, cycle, control, siege, bridge_spam, bait, spawndeck, midrange)"
	fuzzStorageUsage           = "Path to fuzz storage database for data-driven card scoring (default: ~/.cr-api/fuzz_top_decks.db)"
	fuzzWeightUsage            = "Weight for fuzz-based card scoring (0.0-1.0, default 0.10 = 10%)"
	fuzzDeckLimitUsage         = "Number of top fuzz decks to analyze for card stats (default 100)"
	evolutionSlotsDefaultUsage = "Number of evolution slots available (default 2)"
)

func deckEvolutionFlags() []cli.Flag {
	return []cli.Flag{
		unlockedEvolutionsFlag(),
		&cli.IntFlag{Name: "evolution-slots", Value: defaultEvolutionSlots, Usage: evolutionSlotsDefaultUsage},
	}
}

func deckCombatFlags() []cli.Flag {
	return []cli.Flag{
		&cli.Float64Flag{Name: "combat-stats-weight", Value: defaultCombatStatsWeight, Usage: combatStatsWeightUsage},
		&cli.BoolFlag{Name: "disable-combat-stats", Usage: disableCombatStatsUsage},
	}
}

func deckBuilderScoringFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{Name: "enable-synergy", Usage: enableSynergyUsage},
		&cli.Float64Flag{Name: "synergy-weight", Value: defaultSynergyWeight, Usage: synergyWeightUsage},
		&cli.BoolFlag{Name: "prefer-unique", Usage: preferUniqueUsage},
		&cli.Float64Flag{Name: "uniqueness-weight", Value: defaultUniquenessWeight, Usage: uniquenessWeightUsage},
		&cli.StringSliceFlag{Name: "avoid-archetype", Usage: avoidArchetypeUsage},
		&cli.StringFlag{Name: "fuzz-storage", Usage: fuzzStorageUsage},
		&cli.Float64Flag{Name: "fuzz-weight", Value: defaultFuzzWeight, Usage: fuzzWeightUsage},
		&cli.IntFlag{Name: "fuzz-deck-limit", Value: defaultFuzzDeckLimit, Usage: fuzzDeckLimitUsage},
	}
}

func deckSharedBuilderFlags() []cli.Flag {
	flags := deckEvolutionFlags()
	flags = append(flags, deckCombatFlags()...)
	flags = append(flags, deckBuilderScoringFlags()...)
	flags = append(flags, boostedCardLevelFlag())
	return flags
}
