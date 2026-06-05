package main

import "github.com/urfave/cli/v3"

var (
	combatStatsWeightFlag = &cli.Float64Flag{
		Name:  combatStatsWeightFlagName,
		Value: defaultCombatStatsWeight,
		Usage: combatStatsWeightUsage,
	}
	disableCombatStatsFlag = &cli.BoolFlag{
		Name:  disableCombatStatsFlagName,
		Usage: disableCombatStatsUsage,
	}
)

func combatStatsFlags() []cli.Flag {
	return []cli.Flag{
		combatStatsWeightFlag,
		disableCombatStatsFlag,
	}
}
