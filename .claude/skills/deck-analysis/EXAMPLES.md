# Deck Analysis Skill - Examples

## Example 1: Basic Comprehensive Analysis

**User Request:**
> Analyze decks for player #R8QGUQRCV

**Workflow:**
```bash
# 1. Build binary
task build

# 2. Verify player
./bin/cr-api player get --tag "#R8QGUQRCV"

# 3. Run comprehensive analysis
./bin/cr-api deck analyze-suite \
  --tag "#R8QGUQRCV" \
  --strategies all \
  --variations 3 \
  --verbose
```

**Output:**
- 18 decks generated (6 strategies × 3 variations)
- Top deck: Cycle bait (6.69/10.0)
- Report saved to `data/analysis/reports/`

---

## Example 2: Evolution-Focused Analysis

**User Request:**
> Build strong 1v1 decks with my Archers evolution for #R8QGUQRCV

**Workflow:**
```bash
# 1. Build binary
task build

# 2. Comprehensive analysis
./bin/cr-api deck analyze-suite \
  --tag "#R8QGUQRCV" \
  --strategies all \
  --variations 3 \
  --verbose

# 3. Targeted Archers cycle deck
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy cycle \
  --include-cards "Archers" \
  --max-elixir 3.2 \
  --enable-synergy

# 4. Targeted Archers control deck
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy control \
  --include-cards "Archers" \
  --enable-synergy
```

**Output:**
- General analysis: 18 decks
- Archers cycle deck: 2.50 avg elixir with Archers evolution
- Archers control deck: 3.75 avg elixir with defensive focus
- Upgrade recommendations prioritizing Princess and Electro Wizard

---

## Example 3: Low-Elixir Cycle Focus

**User Request:**
> Build fast cycle decks under 3.0 elixir for #ABC123

**Workflow:**
```bash
./bin/cr-api deck analyze-suite \
  --tag "#ABC123" \
  --strategies cycle \
  --max-elixir 3.0 \
  --variations 5 \
  --verbose
```

**Expected Output:**
- 5 ultra-fast cycle deck variations
- All decks under 3.0 average elixir
- Optimized for chip damage and out-cycling opponents

---

## Example 4: Specific Card Requirements

**User Request:**
> Build aggro decks with Hog Rider and Valkyrie for #XYZ789

**Workflow:**
```bash
./bin/cr-api deck build \
  --tag "#XYZ789" \
  --strategy aggro \
  --include-cards "Hog Rider" \
  --include-cards "Valkyrie" \
  --enable-synergy
```

**Expected Output:**
- Single optimized aggro deck
- Must include Hog Rider and Valkyrie
- Remaining 6 cards selected for synergy and aggro strategy

---

## Example 5: Multi-Strategy Comparison

**User Request:**
> Compare cycle, control, and aggro decks for #DEF456

**Workflow:**
```bash
./bin/cr-api deck analyze-suite \
  --tag "#DEF456" \
  --strategies "cycle,control,aggro" \
  --variations 4 \
  --top-n 6 \
  --verbose
```

**Expected Output:**
- 12 decks total (3 strategies × 4 variations)
- Top 6 decks compared in final report
- Strategy-specific recommendations

---

## Example 6: High-Level Competition Decks

**User Request:**
> Build competitive decks avoiding commons and rares for #GHI789

**Workflow:**
```bash
./bin/cr-api deck analyze-suite \
  --tag "#GHI789" \
  --strategies all \
  --min-elixir 3.5 \
  --variations 2 \
  --verbose
```

**Note:** The tool doesn't have a direct rarity filter, but higher elixir constraints typically favor Epic/Legendary cards.

---

## Example 7: Graveyard-Focused Control

**User Request:**
> Build Graveyard control decks for #JKL012

**Workflow:**
```bash
./bin/cr-api deck build \
  --tag "#JKL012" \
  --strategy control \
  --include-cards "Graveyard" \
  --min-elixir 3.5 \
  --enable-synergy
```

**Expected Output:**
- Control deck centered around Graveyard win condition
- Likely includes Poison, defensive buildings, and support units
- Strategic notes on Graveyard placement and chip damage strategy

---

## Example 8: Offline Analysis (No API Calls)

**User Request:**
> Re-analyze my decks from yesterday's data

**Workflow:**
```bash
./bin/cr-api deck analyze-suite \
  --from-analysis \
  --analysis-dir data/analysis \
  --top-n 10
```

**Use Case:**
- Analyze existing data without API calls
- Re-run comparisons with different top-N settings
- Faster iterations during development

---

## Example Response Format

After running analysis, Claude provides:

### Player Profile
```
Player: ZyLogan (#R8QGUQRCV)
- Trophies: 4753 (Arena: Serenity Peak)
- Win Rate: 54.7% (323W/267L)
- Card Collection: 98 cards
- Unlocked Evolution: Archers
```

### Top Recommendations
```
🥇 Top 3: Cycle Bait Decks (6.69/10.0)
   - Average Elixir: 2.38
   - Cards: Goblin Barrel, Tombstone, Vines, Princess,
            Ice Golem, Spear Goblins, Fire Spirit, Bats
   - Strengths: Defense (7.9), Synergy (7.2)
   - Weaknesses: Attack (3.4)
```

### Targeted Builds
```
Archers Cycle Deck (2.50 avg elixir)
- Evolution: Archers★
- Upgrade Priority: Princess (11→12) - 50k gold
```

### File Locations
```
- Full report: data/analysis/reports/20260118_183306_deck_analysis_report_R8QGUQRCV.md
- Raw data: data/analysis/decks/20260118_183306_deck_suite_summary_R8QGUQRCV.json
```

---

## Common Patterns

### For New Players (Low Card Levels)
```bash
--strategies balanced,cycle \
--max-elixir 3.5 \
--prioritize-upgrades
```

### For F2P Players
```bash
--strategies all \
--enable-synergy \
--variations 3
# Then check F2P Friendly scores in output
```

### For Specific Archetype Testing
```bash
--strategy <archetype> \
--variations 5 \
--enable-synergy
```

### For Tournament Preparation
```bash
--strategies all \
--min-elixir 3.0 \
--max-elixir 4.5 \
--variations 3 \
--top-n 10
```
