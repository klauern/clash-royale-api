package main

import "github.com/urfave/cli/v3"

const defaultPlayerTagFlagUsage = "Player tag (without #)"

func playerTagFlag(required bool) *cli.StringFlag {
	return playerTagFlagWithUsage(required, defaultPlayerTagFlagUsage)
}

func playerTagFlagWithUsage(required bool, usage string) *cli.StringFlag {
	if usage == "" {
		usage = defaultPlayerTagFlagUsage
	}

	return &cli.StringFlag{
		Name:     "tag",
		Aliases:  []string{"p"},
		Usage:    usage,
		Required: required,
	}
}
