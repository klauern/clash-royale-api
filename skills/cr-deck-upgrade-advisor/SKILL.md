---
name: cr-deck-upgrade-advisor
description: Deck building and upgrade projection guidance for this Clash Royale Go repo. Use when asked to build best decks (1v1 or war lineups), suggest upgrades or archetype shifts, or simulate "within reach" decks using ./bin/cr-api, ./bin/deckbuilder, or a projected analysis file.
---

# CR Deck Upgrade Advisor

## Overview

Guide deck building and upgrade suggestions using the repo binaries and player card data. Focus on multi-option 1v1 decks, 4-deck war lineups with no repeats, and upgrade projections tied to affordability.

## Inputs to Collect

- Player tag (with or without #)
- Goal: best 1v1 options, war lineup, upgrade path, or archetype shift
- Constraints: strategy, min/max elixir, include/exclude cards, unlocked evolutions, evolution slots, combat stats weight
- Budget: upgrades to simulate and wildcard counts by rarity (if affordability matters)

## Quick Start (Live API Deck Build)

Use this when the user wants the best current deck from their live collection.

```bash
./bin/cr-api deck build --tag <TAG> --strategy balanced --data-dir ./data --save
```

Adjust with:
- `--strategy` (balanced, aggro, control, cycle, splash, spell)
- `--min-elixir` / `--max-elixir`
- `--include-cards` / `--exclude-cards`
- `--unlocked-evolutions` / `--evolution-slots`
- `--combat-stats-weight` or `--disable-combat-stats`

## Task: Multiple 1v1 Options

1) Build 3-5 decks with different strategies or constraints.
2) Present each deck with average elixir and a short rationale.
3) Highlight any cards that repeat across options (staples).

Example command set:
```bash
./bin/cr-api deck build --tag <TAG> --strategy balanced --data-dir ./data --save
./bin/cr-api deck build --tag <TAG> --strategy cycle --min-elixir 2.6 --max-elixir 3.2 --data-dir ./data --save
./bin/cr-api deck build --tag <TAG> --strategy control --min-elixir 3.4 --max-elixir 4.2 --data-dir ./data --save
```

If the user wants a specific archetype, force the win condition via `--include-cards` and then build 2-3 variants by tweaking elixir or exclusion lists.

## Task: War Lineup (4 Decks, 32 Unique Cards)

Use the dedicated `deck war` command to automatically build an optimal war lineup with no repeated cards:

```bash
./bin/cr-api deck war --tag <TAG>
```

The command:
1. Evaluates all 8 available archetypes (Beatdown, Cycle, Control, Siege, BridgeSpam, Midrange, Spawndeck, Bait)
2. Finds the optimal 4-archetype combination that maximizes total deck quality
3. Ensures zero card overlap across all decks
4. Returns all 4 decks with card details, elixir costs, and a unique card count

### Options

- `--deck-count N`: Build N decks instead of 4 (useful for 3-deck formats)
- `--unlocked-evolutions "Card1,Card2"`: Specify which evolutions you've unlocked
- `--evolution-slots N`: Override evolution slot limit (default 2)
- `--combat-stats-weight 0.5`: Adjust combat stats influence (0.0-1.0)
- `--disable-combat-stats`: Use traditional scoring only
- `--verbose`: Show detailed progress

### Example Output

```
WAR DECK SET (NO REPEATS)
========================

Player: YourName (#TAG)
Decks: 4
Unique cards: 32
Total score: 42.567

Deck 1 - Beatdown
Average Elixir: 3.88
...
```

### Manual Approach (Legacy)

For more control over specific archetypes, you can still build decks manually with `--exclude-cards`:

```bash
./bin/cr-api deck build --tag <TAG> --strategy balanced --save
./bin/cr-api deck build --tag <TAG> --strategy cycle --exclude-cards "Card1,Card2,..."
./bin/cr-api deck build --tag <TAG> --strategy control --exclude-cards "..."
./bin/cr-api deck build --tag <TAG> --strategy aggro --exclude-cards "..."
```

## Task: Upgrade Suggestions and "Within Reach" Projections

Use this when the user wants to know which upgrades unlock a new archetype or strengthen a current deck.

1) Generate an analysis file:
```bash
./bin/cr-api analyze --tag <TAG> --save --data-dir ./data
```
2) Use the projection helper to apply a small upgrade plan and budget (wildcards by rarity).
```bash
./scripts/project_deck_projection.py --analysis ./data/analysis/<analysis>.json --plan ./data/upgrade_plan.json --tag <TAG>
```
3) Build a deck from the projected analysis using `./bin/deckbuilder --analysis-file`.
4) Compare current vs projected decks and call out the minimal upgrades that unlock the better option.

If wildcard counts are unknown, ask for a rough budget by rarity or skip affordability checks and label the projection as \"unbounded\" (use `--unbounded`).

## References

- Read `DECK_BUILDER.md` for deck roles and scoring details.
- Read `DECK_STRATEGIES.md` for strategy definitions and tuning.
- Use `cmd/cr-api/deck_commands.go` for exact flag names and defaults.
