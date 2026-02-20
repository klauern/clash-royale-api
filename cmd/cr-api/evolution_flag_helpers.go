package main

import (
	"strings"

	"github.com/urfave/cli/v3"
)

const (
	unlockedEvolutionsFlagName  = "unlocked-evolutions"
	unlockedEvolutionsEnvVar    = "UNLOCKED_EVOLUTIONS"
	unlockedEvolutionsFlagUsage = "Comma-separated list of cards with unlocked evolutions (overrides UNLOCKED_EVOLUTIONS env var)"
)

func unlockedEvolutionsFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:    unlockedEvolutionsFlagName,
		Usage:   unlockedEvolutionsFlagUsage,
		Sources: cli.EnvVars(unlockedEvolutionsEnvVar),
	}
}

func unlockedEvolutionsFromCommand(cmd *cli.Command) []string {
	return parseUnlockedEvolutionsCSV(cmd.String(unlockedEvolutionsFlagName))
}

func parseUnlockedEvolutionsCSV(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	parts := strings.Split(trimmed, ",")
	parsed := make([]string, 0, len(parts))
	for _, part := range parts {
		card := strings.TrimSpace(part)
		if card == "" {
			continue
		}
		parsed = append(parsed, card)
	}

	if len(parsed) == 0 {
		return nil
	}
	return parsed
}
