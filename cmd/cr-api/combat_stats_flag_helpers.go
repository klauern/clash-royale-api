package main

import "github.com/urfave/cli/v3"

const (
	combatStatsWeightFlagName   = "combat-stats-weight"
	disableCombatStatsFlagName  = "disable-combat-stats"
	combatStatsWeightFlagUsage  = "Weight for combat stats in scoring (0.0-1.0, where 0=disabled, 0.25=default, 1.0=combat-only)"
	disableCombatStatsFlagUsage = "Disable combat stats completely (use traditional scoring only)"
	defaultCombatStatsWeight    = 0.25
)

func combatStatsFlags() []cli.Flag {
	return []cli.Flag{
		&cli.Float64Flag{
			Name:  combatStatsWeightFlagName,
			Value: defaultCombatStatsWeight,
			Usage: combatStatsWeightFlagUsage,
		},
		&cli.BoolFlag{
			Name:  disableCombatStatsFlagName,
			Usage: disableCombatStatsFlagUsage,
		},
	}
}
