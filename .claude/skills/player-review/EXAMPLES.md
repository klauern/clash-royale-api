# Player Review Skill - Examples

## Example 1: Standard Terminal Review

**User Request:**
> Review player #R8QGUQRCV

**Workflow:**
```bash
# Build binary if needed
task build

# Run review
./bin/cr-api review --tag "#R8QGUQRCV"
```

**Expected Output Structure:**
```
═══════════════════════════════════════════════════════
  PLAYER REVIEW: ZyLogan (#R8QGUQRCV)
═══════════════════════════════════════════════════════

── Profile ─────────────────────────────────────────────
  Level:    13   Trophies: 4753   Best: 5102
  Arena:    Serenity Peak
  Record:   323 W / 267 L  (54.7% win rate)

── Playstyle ───────────────────────────────────────────
  Style:      Fast Cycle
  Aggression: Moderate
  Traits:     spell-heavy, chip-damage
  Avg Elixir: 2.4

── Top Archetype ───────────────────────────────────────
  Name:       Hog Cycle
  Win Con:    Hog Rider
  Viability:  7.4  (Strong)
  Gold to competitive: 32k

── Top Cross-Archetype Upgrade Priorities ──────────────
  1. Hog Rider              Lv 11  cost 50k gold
     Unlocks 4 archetype(s)  (+12.3 viability)
  2. Musketeer              Lv 10  cost 50k gold
     Unlocks 3 archetype(s)  (+8.1 viability)
  3. Ice Golem              Lv 11  cost 20k gold
     Unlocks 2 archetype(s)  (+5.4 viability)

── Current Deck vs Best Recommended ────────────────────
  Overall:    current 0.61  →  recommended 0.79  (Δ +0.18)
  Archetype:  current Hog Cycle       →  recommended Hog Cycle
  Level fit:  current 82%  →  recommended 88%
  Kept (6):   Hog Rider, Musketeer, Cannon, Ice Spirit, Log, Skeletons
  New  (2):   Ice Golem, Fireball

── Budget: Next 20k Gold ───────────────────────────────
  1. Score 0.61 → 0.74  cost 18k gold
     Hog Rider, Musketeer, Ice Golem, Cannon, Ice Spirit, Skeletons, Log, Fireball
  2. Score 0.61 → 0.71  cost 20k gold
     ...
```

**Response to User:**
Provide a concise summary highlighting:
- Current strength (54.7% win rate, Strong tier archetype)
- Most impactful action (upgrade Hog Rider to unlock 4 archetypes)
- Deck gap (current deck is 0.18 score below optimal; 2 card swaps recommended)
- Budget action (18k gold gets to 0.74 score — worth prioritizing)

---

## Example 2: Markdown Output for Sharing

**User Request:**
> Give me a markdown review of player #ABC123 I can paste into Discord

**Workflow:**
```bash
./bin/cr-api review --tag "#ABC123" --format markdown
```

**Expected Output:**
```markdown
# Player Review: PlayerName (#ABC123)

## Profile

| Field | Value |
|-------|-------|
| Level | 13 |
| Trophies | 5200 (best: 5800) |
| Arena | Ultimate Champion |
| Record | 512 W / 398 L (56.3%) |

## Playstyle

| Field | Value |
|-------|-------|
| Style | Beatdown |
| Aggression | Passive |
| Avg Elixir | 4.1 |
| Traits | tank-heavy, spell-support |

## Top Archetype

**Golem Beatdown** (Golem) — 8.2 viability (Elite)

Gold to competitive tier: **0** (already there)

## Cross-Archetype Upgrade Priorities

| # | Card | Level | Gold Cost | Archetypes Unlocked | Viability Gain |
...
```

---

## Example 3: JSON for Scripting

**User Request:**
> Get a JSON review of player #XYZ789 and tell me the deck delta score

**Workflow:**
```bash
# Full JSON output
./bin/cr-api review --tag "#XYZ789" --format json

# Extract just the deck delta
./bin/cr-api review --tag "#XYZ789" --format json | jq '.DeckDelta'
```

**Expected JSON structure:**
```json
{
  "Player": { "Name": "...", "Tag": "#XYZ789", "Trophies": 4200, ... },
  "Playstyle": { "DeckStyle": "...", "AggressionLevel": "...", ... },
  "TopArchetype": { "Name": "...", "WinCondition": "...", "ViabilityScore": 7.1, ... },
  "CrossArchUpgrades": [
    { "CardName": "Giant", "CurrentLevel": 10, "GoldCost": 50000, "ArchetypesUnlocked": 3, "TotalViabilityGain": 9.2 }
  ],
  "DeckDelta": {
    "CurrentScore": 0.52,
    "RecommendedScore": 0.77,
    "ScoreDelta": 0.25,
    "CurrentArchetype": "Splashyard",
    "RecommendedArchetype": "Giant Beatdown",
    "CurrentLevelFit": 0.74,
    "RecommendedLevelFit": 0.85,
    "SharedCards": ["Giant", "Witch", "Baby Dragon"],
    "ReplacedCards": ["Graveyard", "Poison"]
  },
  "BudgetDecks": [ ... ]
}
```

---

## Example 4: Default Player Quick Review

**User Request:**
> Review my account

**Workflow:**
```bash
# Read DEFAULT_PLAYER_TAG from .env
grep DEFAULT_PLAYER_TAG .env
# e.g. DEFAULT_PLAYER_TAG="#R8QGUQRCV"

./bin/cr-api review --tag "#R8QGUQRCV"
```

---

## Example 5: Player With No Current Deck

**User Request:**
> Review player #NEWPLAYER who just started

**Workflow:**
```bash
./bin/cr-api review --tag "#NEWPLAYER"
```

**Expected — Deck Delta Section:**
```
── Current Deck vs Best Recommended ────────────────────
  Deck delta unavailable.
```

The review still completes. All other sections (profile, playstyle, archetype, upgrades, budget) are shown normally. The missing deck delta is non-fatal.

---

## Example 6: Comparing Before/After an Upgrade

**User Request:**
> How much does upgrading Hog Rider change my review?

**Workflow:**
```bash
# Current state
./bin/cr-api review --tag "#R8QGUQRCV"

# Simulate the upgrade then re-review if what-if file is saved
./bin/cr-api what-if --tag "#R8QGUQRCV" --upgrade "Hog Rider:13" --save

# Or just note the CrossArchUpgrades section shows the delta:
# "Unlocks 4 archetype(s)  (+12.3 viability)"
```

The `review` output's CrossArchUpgrades section already quantifies the impact
of each priority upgrade, so a separate `what-if` call is only needed for
detailed before/after deck comparison.

---

## Common Patterns

### For Players Stuck on a Trophy Range
```bash
./bin/cr-api review --tag "<TAG>"
# Focus on: Deck Delta (is current deck sub-optimal?) and Budget Plan
```

### For New Players Optimizing Upgrades
```bash
./bin/cr-api review --tag "<TAG>"
# Focus on: Cross-Archetype Upgrade Priorities — upgrade what unlocks the most paths
```

### For Competitive Players
```bash
./bin/cr-api review --tag "<TAG>" --format markdown
# Share full review with clanmates or coaches
```

### Quick Gut-Check Before Grinding
```bash
./bin/cr-api review --tag "<TAG>"
# Check Deck Delta: if ScoreDelta > 0.15, consider switching decks before grinding
```
