# Player Review Skill

## Invocation

Use this skill when the user wants a comprehensive, one-shot analysis of a Clash Royale player.

**Trigger phrases:**
- "review player X"
- "full player analysis"
- "analyze my account"
- "show me everything about player X"
- "player overview"
- "review #R8QGUQRCV"
- "comprehensive review"
- "what should I upgrade / what deck should I run"

## What This Skill Does

Runs a full player review in a single API call using `cr-api review`. The command
orchestrates all major analyzers internally and outputs a consolidated report covering:

1. **Profile snapshot** — level, trophies, arena/league, win rate
2. **Playstyle summary** — deck style, aggression level, traits, avg elixir
3. **Top archetype** — best-fit archetype name, win condition, viability score, gold-to-competitive
4. **Cross-archetype upgrade priorities** — top-3 upgrades that unlock the most archetypes
5. **Deck delta** — current deck vs. top recommended deck (score gap, level fit, swapped cards)
6. **Budget plan** — up to 3 decks achievable within the next 20,000 gold

## Parameters

### Required
- `--tag` / `-p`: Player tag (with or without `#`)

### Optional
- `--format`: Output format — `human` (default), `json`, or `markdown`
  - `human`: Formatted terminal output with section headers
  - `json`: Machine-readable full report (pipe to `jq`, save to file, etc.)
  - `markdown`: GitHub-flavored markdown tables, great for sharing

## Workflow

### 1. Build the binary (if needed)
```bash
task build
```

### 2. Run the review
```bash
./bin/cr-api review --tag "<PLAYER_TAG>"
```

### 3. For structured output
```bash
# Markdown (good for sharing or pasting into notes)
./bin/cr-api review --tag "<PLAYER_TAG>" --format markdown

# JSON (good for piping to jq or saving)
./bin/cr-api review --tag "<PLAYER_TAG>" --format json | jq '.TopArchetype'
```

## Output Sections Explained

### Profile
Basic stats: level, trophies, best trophies, arena, league, win/loss record, win rate.

### Playstyle
Derived from the player's current deck. Identifies:
- **DeckStyle**: e.g., "Fast Cycle", "Beatdown", "Control"
- **AggressionLevel**: Passive / Moderate / Aggressive
- **PlaystyleTraits**: List of descriptors (e.g., "spell-heavy", "swarm-focused")
- **CurrentDeckAvgElixir**: Average elixir of the current 8-card deck

### Top Archetype
The most viable archetype detected from the player's card collection:
- **Name**: Archetype name (e.g., "Hog Cycle", "Golem Beatdown")
- **WinCondition**: Primary win condition card(s)
- **ViabilityScore** / **ViabilityTier**: How close to competitive (e.g., 7.8 / "Strong")
- **GoldToCompetitive**: Gold cost to reach competitive tier

### Cross-Archetype Upgrade Priorities
Top-3 card upgrades ranked by how many archetypes they unlock:
- **CardName**, **CurrentLevel**, **GoldCost**
- **ArchetypesUnlocked**: How many archetypes become viable after the upgrade
- **TotalViabilityGain**: Sum of viability score improvements

### Deck Delta
Compares the player's *current in-game deck* against the *top recommended deck*:
- **CurrentScore** / **RecommendedScore** / **ScoreDelta**: Overall evaluation scores
- **CurrentArchetype** / **RecommendedArchetype**: Archetype classification of each
- **CurrentLevelFit** / **RecommendedLevelFit**: % of cards at or near max level
- **SharedCards**: Cards kept from current deck in the recommendation
- **ReplacedCards**: New cards introduced in the recommendation

> Note: If the player has no current deck set in-game, the delta section shows "unavailable".

### Budget: Next 20k Gold
Up to 3 decks achievable by spending at most 20,000 gold in upgrades:
- **CurrentScore** → **ProjectedScore**: Before/after upgrade impact
- **TotalGoldNeeded**: Exact gold investment required
- Card composition of the budget deck

## Response Format

After running the review, provide:

1. **Highlight the key insight** — What is the most actionable takeaway? (e.g., "You're 2 upgrades away from a competitive Hog Cycle deck")
2. **Profile Summary** — Trophies, win rate, arena
3. **Playstyle and Top Archetype** — What the data says about how the player plays
4. **Top Upgrade Priority** — The single most impactful upgrade (from cross-arch priorities)
5. **Deck Delta** — Whether the current deck is near-optimal or significantly behind
6. **Budget Recommendation** — Best deck achievable with next 20k gold
7. **Next Steps** — Specific, actionable advice (upgrade X, try deck Y, etc.)

## Examples

### Basic Human Review
```bash
./bin/cr-api review --tag "#R8QGUQRCV"
```

### Markdown Output (for sharing)
```bash
./bin/cr-api review --tag "#R8QGUQRCV" --format markdown
```

### JSON for Programmatic Use
```bash
# Full report
./bin/cr-api review --tag "#R8QGUQRCV" --format json

# Extract just the top archetype
./bin/cr-api review --tag "#R8QGUQRCV" --format json | jq '.TopArchetype'

# Extract upgrade priorities
./bin/cr-api review --tag "#R8QGUQRCV" --format json | jq '.CrossArchUpgrades'
```

## When to Use vs. Other Skills

| Goal | Use |
|------|-----|
| One comprehensive overview | `review` (this skill) |
| Build many deck variations | `deck-analysis` skill |
| Quick single deck build | `default-player-deck` skill |
| Simulate a specific upgrade | `./bin/cr-api what-if` |
| Detailed upgrade ranking only | `./bin/cr-api upgrade-impact` |
| Compare specific decks side-by-side | `./bin/cr-api compare` |

## Notes

- The `review` command makes a single API call to fetch player data, then runs all
  analysis locally — it is fast (typically <5 seconds).
- Deck delta is non-fatal: if the player has no current deck, the review still
  completes and shows all other sections.
- Gold amounts are displayed in compact form (e.g., "50k" for 50,000).
- The budget plan uses a 20k gold cap — use `deck budget` directly for a custom cap.
