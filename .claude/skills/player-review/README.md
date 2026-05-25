# Player Review Skill

One-stop comprehensive analysis for a Clash Royale player. Runs all major analyzers in a single command.

## Quick Start

Ask Claude to review a player:

```
Review player #R8QGUQRCV
Give me a full analysis of my account
What should I upgrade for player #ABC123?
Show me everything about player #XYZ789
```

## What You Get

### Single Command, Full Picture
The `review` command orchestrates six analyses internally:

| Section | What It Shows |
|---------|---------------|
| Profile | Level, trophies, arena, win rate |
| Playstyle | Deck style, aggression, avg elixir |
| Top Archetype | Best-fit archetype + viability score |
| Upgrade Priorities | Top-3 upgrades that unlock the most archetypes |
| Deck Delta | Gap between current deck and best recommendation |
| Budget Plan | Best decks achievable within 20k gold |

### Three Output Formats

- **human** (default) — Readable terminal output with section headers
- **markdown** — GitHub-flavored tables, great for sharing
- **json** — Full machine-readable report, pipe to `jq`

## Key Insight: Deck Delta

The deck delta section is uniquely useful — it tells you *how close your current deck
is to the best recommendation*, including which cards to swap and whether the gap is
primarily a level issue or an archetype mismatch.

## Usage

```bash
# Terminal output
./bin/cr-api review --tag "#R8QGUQRCV"

# Markdown (share or paste into notes)
./bin/cr-api review --tag "#R8QGUQRCV" --format markdown

# JSON (for further processing)
./bin/cr-api review --tag "#R8QGUQRCV" --format json | jq '.TopArchetype'
```

## Requirements

- Valid Clash Royale API token in `.env` or `$CLASH_ROYALE_API_TOKEN`
- Player tag (find in-game under your profile)
- Built `cr-api` binary (`task build`)

## When to Use This vs. Other Skills

| Need | Skill / Command |
|------|----------------|
| Comprehensive one-shot overview | **player-review** (this skill) |
| Build 18+ deck variations | `deck-analysis` skill |
| Quick single deck for your account | `default-player-deck` skill |
| Compare specific decks | `./bin/cr-api compare` |
| Simulate upgrading a card | `./bin/cr-api what-if` |
