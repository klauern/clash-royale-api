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
		Usage: "Generate AI agent onboarding instructions",
		Description: "Outputs onboarding instructions for AI agents in various formats. " +
			"Default: lean AGENTS.md snippet (~10 lines). Use --format json for structured " +
			"data or --format copilot for GitHub Copilot instructions.",
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
- Go 1.22+ required
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
	data := map[string]interface{}{
		"project": map[string]interface{}{
			"name":        "clash-royale-api",
			"description": "Go-only Clash Royale API client and analysis tool",
			"language":    "Go",
			"framework":   "urfave/cli/v3",
		},
		"quick_start": map[string]interface{}{
			"setup":   []string{"task setup", "task build"},
			"analyze": "./bin/cr-api analyze --tag <TAG>",
			"tasks":   "task",
			"issues":  "bd ready --json",
		},
		"key_commands": []map[string]interface{}{
			{"name": "player", "description": "Get player information"},
			{"name": "analyze", "description": "Card collection analysis and upgrade priorities"},
			{"name": "deck build", "description": "Build optimized decks"},
			{"name": "events scan", "description": "Track and analyze event decks"},
			{"name": "evolutions recommend", "description": "Evolution tracking and recommendations"},
			{"name": "archetypes", "description": "Analyze archetype variety across playstyles"},
			{"name": "what-if", "description": "Simulate card upgrade impact"},
		},
		"task_runner": map[string]interface{}{
			"tool":    "Taskfile.dev",
			"command": "task",
		},
		"development": map[string]interface{}{
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
