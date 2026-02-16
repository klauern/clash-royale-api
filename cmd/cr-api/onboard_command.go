package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

// addOnboardCommand adds the onboard command to generate AI agent onboarding instructions.
func addOnboardCommand() *cli.Command {
	return &cli.Command{
		Name:  "onboard",
		Usage: "Generate onboarding instructions",
		Description: "Outputs onboarding instructions in various formats. " +
			"Default: developer/AGENTS.md snippet. Use 'onboard user' for user-facing quickstart.",
		Commands: []*cli.Command{
			addUserOnboardCommand(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Usage: "Output format: markdown (default), json, copilot",
				Value: "markdown",
				Action: func(ctx context.Context, cmd *cli.Command, s string) error {
					// Validate format at parse time
					if s != "" && s != "markdown" && s != "json" && s != "copilot" {
						return fmt.Errorf("invalid format: %q (must be: markdown, json, or copilot)", s)
					}
					return nil
				},
			},
		},
		Action: onboardCommandAction,
	}
}

// addUserOnboardCommand adds the user-facing onboarding subcommand.
func addUserOnboardCommand() *cli.Command {
	return &cli.Command{
		Name:  "user",
		Usage: "Generate user-facing quickstart guide",
		Description: "Outputs a user-facing quickstart guide for analyzing a player. " +
			"Shows the essential first commands to run with a player tag.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Usage: "Output format: markdown (default), json, copilot",
				Value: "markdown",
				Action: func(ctx context.Context, cmd *cli.Command, s string) error {
					if s != "" && s != "markdown" && s != "json" && s != "copilot" {
						return fmt.Errorf("invalid format: %q (must be: markdown, json, or copilot)", s)
					}
					return nil
				},
			},
		},
		Action: userOnboardCommandAction,
	}
}

// userOnboardCommandAction executes the user onboarding command.
func userOnboardCommandAction(ctx context.Context, cmd *cli.Command) error {
	format := cmd.String("format")
	if format == "" {
		format = "markdown"
	}

	switch format {
	case "json":
		return outputUserJSONOnboard()
	case "copilot":
		return outputUserCopilotOnboard()
	case "markdown":
		return outputUserMarkdownOnboard()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

// outputUserMarkdownOnboard outputs a user-facing markdown quickstart guide.
func outputUserMarkdownOnboard() error {
	const template = `# Quickstart: Analyzing a Player

## Player Tag Format

Use your player tag with or without the ` + "`#`" + ` prefix:
- ` + "`cr-api player --tag ABC123`" + `
- ` + "`cr-api player --tag #ABC123`" + `

**Finding your tag:** In-game, tap your Profile → Copy Tag

## First 3 Commands to Run

### 1. Get Player Info
` + "```bash" + `
cr-api player --tag <TAG>
` + "```" + `
**What you get:** Name, level, arena, wins, losses, current deck, and clan info.

### 2. Analyze Collection
` + "```bash" + `
cr-api analyze --tag <TAG>
` + "```" + `
**What you get:** Upgrade priorities, card levels, and rare cards to focus on.

### 3. Analyze Playstyle
` + "```bash" + `
cr-api playstyle --tag <TAG>
` + "```" + `
**What you get:** Archetype preferences, win rates by deck type, and playstyle patterns.

## Next Steps

### Build a Deck
` + "```bash" + `
cr-api deck build --tag <TAG>
` + "```" + `
Builds an optimized 1v1 ladder deck based on your card collection.

### Scan Events
` + "```bash" + `
cr-api events scan --tag <TAG>
` + "```" + `
Scans your battle log for event decks and tracks performance.

### Evolution Recommendations
` + "```bash" + `
cr-api evolutions recommend --tag <TAG>
` + "```" + `
Shows which evolutions you should prioritize based on your deck usage.

## Prerequisites

Make sure your API token is configured:
` + "```bash" + `
# Check if token is set
cr-api player --help
# If you get an API error, run:
task setup
` + "```" + `

## More Commands

- ` + "`cr-api deck war --tag <TAG>`" + ` - Build 4-deck war lineup
- ` + "`cr-api archetypes --tag <TAG>`" + ` - Analyze archetype variety
- ` + "`cr-api what-if --tag <TAG> --card <CARD>`" + ` - Simulate upgrade impact
- ` + "`cr-api --help`" + ` - See all 40+ commands
`
	fmt.Print(template)
	return nil
}

// outputUserJSONOnboard outputs structured JSON for programmatic use.
func outputUserJSONOnboard() error {
	data := map[string]any{
		"player_tag": map[string]any{
			"description": "Use with or without # prefix",
			"examples":    []string{"cr-api player --tag ABC123", "cr-api player --tag #ABC123"},
			"how_to_find": "In-game: Profile → Copy Tag",
		},
		"first_commands": []map[string]any{
			{
				"order":       1,
				"command":     "cr-api player --tag <TAG>",
				"description": "Get player info: name, level, arena, wins, losses, current deck",
			},
			{
				"order":       2,
				"command":     "cr-api analyze --tag <TAG>",
				"description": "Upgrade priorities, card levels, rare cards to focus on",
			},
			{
				"order":       3,
				"command":     "cr-api playstyle --tag <TAG>",
				"description": "Archetype preferences, win rates by deck type, playstyle patterns",
			},
		},
		"next_steps": []map[string]any{
			{
				"command":     "cr-api deck build --tag <TAG>",
				"description": "Build optimized 1v1 ladder deck",
			},
			{
				"command":     "cr-api events scan --tag <TAG>",
				"description": "Scan battle log for event decks",
			},
			{
				"command":     "cr-api evolutions recommend --tag <TAG>",
				"description": "Evolution recommendations based on deck usage",
			},
		},
		"prerequisites": map[string]any{
			"check": "cr-api player --help",
			"setup": "task setup",
		},
		"more_commands": []map[string]any{
			{"command": "cr-api deck war --tag <TAG>", "description": "Build 4-deck war lineup"},
			{"command": "cr-api archetypes --tag <TAG>", "description": "Analyze archetype variety"},
			{"command": "cr-api what-if --tag <TAG> --card <CARD>", "description": "Simulate upgrade impact"},
			{"command": "cr-api --help", "description": "See all 40+ commands"},
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// outputUserCopilotOnboard outputs GitHub Copilot-friendly instructions.
func outputUserCopilotOnboard() error {
	const template = `# Clash Royale API - User Quickstart

## Player Tag Format

Use player tag with or without ` + "`#`" + ` prefix:
- ` + "`cr-api player --tag ABC123`" + `
- ` + "`cr-api player --tag #ABC123`" + `

Find tag in-game: Profile → Copy Tag

## First 3 Commands

**1. Get Player Info:**
` + "```bash" + `
cr-api player --tag <TAG>
` + "```" + `
Returns: name, level, arena, wins, losses, current deck, clan info

**2. Analyze Collection:**
` + "```bash" + `
cr-api analyze --tag <TAG>
` + "```" + `
Returns: upgrade priorities, card levels, rare cards

**3. Analyze Playstyle:**
` + "```bash" + `
cr-api playstyle --tag <TAG>
` + "```" + `
Returns: archetype preferences, win rates by deck type

## Next Steps

- ` + "`cr-api deck build --tag <TAG>`" + ` - Build optimized 1v1 ladder deck
- ` + "`cr-api events scan --tag <TAG>`" + ` - Scan battle log for event decks
- ` + "`cr-api evolutions recommend --tag <TAG>`" + ` - Evolution recommendations

## Prerequisites

API token must be configured. Check with ` + "`cr-api player --help`" + `. Run ` + "`task setup`" + ` if needed.
`
	fmt.Print(template)
	return nil
}

// onboardCommandAction executes the onboard command.
func onboardCommandAction(ctx context.Context, cmd *cli.Command) error {
	format := cmd.String("format")

	// Empty string defaults to markdown
	if format == "" {
		format = "markdown"
	}

	switch format {
	case "json":
		return outputJSONOnboard()
	case "copilot":
		return outputCopilotOnboard()
	case "markdown":
		return outputMarkdownOnboard()
	default:
		// Should not reach here due to flag validation
		return fmt.Errorf("unknown format: %s", format)
	}
}

// outputMarkdownOnboard outputs a lean AGENTS.md snippet (~10-15 lines).
func outputMarkdownOnboard() error {
	const template = `# AGENTS.md

This project uses [bd (beads)](https://github.com/steveyegge/beads) for issue tracking.

## Quick Start
- Setup: ` + "`task setup && task build`" + `
- Analyze: ` + "`./bin/cr-api analyze --tag <TAG>`" + `
- Tasks: ` + "`task`" + ` (shows all available tasks)
- Issues: ` + "`bd ready --json`" + `

## Key Commands
- ` + "`player`" + ` - Get player information
- ` + "`analyze`" + ` - Card collection analysis and upgrade priorities
- ` + "`deck build`" + ` - Build optimized decks
- ` + "`events scan`" + ` - Track and analyze event decks
- ` + "`evolutions recommend`" + ` - Evolution tracking and recommendations
- ` + "`archetypes`" + ` - Analyze archetype variety across playstyles
- ` + "`what-if`" + ` - Simulate card upgrade impact

## Development
- Go 1.26+ required
- Framework: urfave/cli/v3
- Tests: ` + "`task test`" + `
- Lint: ` + "`task lint`" + `
- Use ` + "`cr-api onboard --format copilot`" + ` for GitHub Copilot instructions
`
	fmt.Print(template)
	return nil
}

// outputJSONOnboard outputs structured JSON for programmatic use.
func outputJSONOnboard() error {
	data := map[string]any{
		"project": map[string]any{
			"name":        "clash-royale-api",
			"description": "Go-only Clash Royale API client and analysis tool",
			"language":    "Go",
			"framework":   "urfave/cli/v3",
		},
		"quick_start": map[string]any{
			"setup":   []string{"task setup", "task build"},
			"analyze": "./bin/cr-api analyze --tag <TAG>",
			"tasks":   "task",
			"issues":  "bd ready --json",
		},
		"key_commands": []map[string]any{
			{"name": "player", "description": "Get player information"},
			{"name": "analyze", "description": "Card collection analysis and upgrade priorities"},
			{"name": "deck build", "description": "Build optimized decks"},
			{"name": "events scan", "description": "Track and analyze event decks"},
			{"name": "evolutions recommend", "description": "Evolution tracking and recommendations"},
			{"name": "archetypes", "description": "Analyze archetype variety across playstyles"},
			{"name": "what-if", "description": "Simulate card upgrade impact"},
		},
		"task_runner": map[string]any{
			"tool":    "Taskfile.dev",
			"command": "task",
		},
		"development": map[string]any{
			"test":  "task test",
			"lint":  "task lint",
			"build": "task build",
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// outputCopilotOnboard outputs GitHub Copilot instructions.
func outputCopilotOnboard() error {
	const template = `# Clash Royale API - GitHub Copilot Instructions

## Project Overview

Go-only Clash Royale API client and analysis tool using the official API.

## Quick Start

Setup and build: ` + "`task setup && task build`" + `
Analyze player: ` + "`./bin/cr-api analyze --tag <PLAYER_TAG>`" + `

## Key Commands

**Analysis:**
- ` + "`analyze`" + ` - Card collection analysis and upgrade priorities
- ` + "`playstyle`" + ` - Analyze player's playstyle
- ` + "`archetypes`" + ` - Analyze archetype variety

**Deck Building:**
- ` + "`deck build`" + ` - Build optimized 1v1 ladder deck
- ` + "`deck war`" + ` - Build 4-deck war lineup

**Events:**
- ` + "`events scan`" + ` - Scan battle logs for event decks
- ` + "`events analyze`" + ` - Analyze event deck performance

**Utilities:**
- ` + "`player`" + ` - Get player information
- ` + "`cards`" + ` - Get card database
- ` + "`what-if`" + ` - Simulate card upgrade impact

## Task Runner

Use ` + "`task`" + ` for common operations:
- ` + "`task build`" + ` - Build binaries
- ` + "`task test`" + ` - Run tests
- ` + "`task lint`" + ` - Run linters

## Development

**Framework:** urfave/cli/v3
**Code Organization:**
- cmd/cr-api/ - Main CLI application
- pkg/ - Reusable libraries (clashroyale, analysis, deck, events)
- internal/ - Internal packages (exporters, storage)

**Testing:**
- ` + "`go test ./...`" + ` or ` + "`task test`" + `
- Integration tests: ` + "`go test -tags=integration ./...`" + `

**Issue Tracking:**
- Use bd (beads) for ALL task tracking
- Never use markdown TODO lists
`
	fmt.Print(template)
	return nil
}
