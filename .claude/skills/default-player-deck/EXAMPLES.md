# Default Player Deck Skill - Examples

## Example 1: Basic Deck Build

**User:** "Build me a deck"

**Skill Actions:**
```bash
# Read DEFAULT_PLAYER_TAG from .env
grep DEFAULT_PLAYER_TAG .env
# Output: DEFAULT_PLAYER_TAG="#R8QGUQRCV"

# Build binary if needed
task build

# Verify player
./bin/cr-api player get --tag "#R8QGUQRCV"

# Build deck with default strategy (balanced)
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy balanced \
  --enable-synergy
```

**Expected Output:**
```
Player: klauer (5600 trophies, Challenger III)
Collection: 102/110 cards
Evolutions: Archers

Recommended Deck (Balanced):
1. Hog Rider (Level 11)
2. Musketeer (Level 10)
3. Cannon (Level 12)
4. Ice Spirit (Level 13)
5. Fireball (Level 11)
6. Skeletons (Level 12)
7. Ice Golem (Level 11)
8. Log (Level 12)

Average Elixir: 3.1

Strategy Notes:
- Classic 2.6 Hog Cycle variant adapted to your levels
- Hog Rider is primary win condition
- Musketeer + Ice Golem for defense
- Cheap cycle cards for quick rotation

Upgrade Recommendations:
1. Hog Rider (11→13) - +15% damage, 50,000 gold
2. Musketeer (10→13) - +20% damage/HP, 50,000 gold
3. Fireball (11→13) - +15% damage, 50,000 gold
```

## Example 2: Cycle Deck Request

**User:** "Build me a cycle deck"

**Skill Actions:**
```bash
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy cycle \
  --max-elixir 3.0 \
  --enable-synergy
```

**Expected Output:**
```
Player: klauer (5600 trophies)

Recommended Deck (Cycle):
1. Miner (Level 11)
2. Wall Breakers (Level 12)
3. Spear Goblins (Level 13)
4. Bats (Level 12)
5. Bomb Tower (Level 11)
6. Poison (Level 11)
7. Log (Level 12)
8. Ice Spirit (Level 13)

Average Elixir: 2.6

Strategy Notes:
- Ultra-fast 2.6 cycle for constant pressure
- Miner + Wall Breakers for chip damage
- Bats and Spear Goblins for cycle and defense
- Bomb Tower for building target and defense

Key Combos:
- Miner + Poison for spell damage
- Wall Breakers at bridge when opponent low on elixir
- Bats to counter single-target troops

Upgrade Recommendations:
1. Miner (11→13) - +15% HP/damage, 50,000 gold
2. Poison (11→13) - +15% damage, 50,000 gold
3. Wall Breakers (12→13) - +10% damage, 20,000 gold
```

## Example 3: Deck with Specific Cards

**User:** "Build me a deck with Hog Rider and Fireball"

**Skill Actions:**
```bash
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy balanced \
  --include-cards "Hog Rider,Fireball" \
  --enable-synergy
```

**Expected Output:**
```
Player: klauer (5600 trophies)

Requested Cards: Hog Rider, Fireball

Recommended Deck:
1. Hog Rider (Level 11) ✓ Requested
2. Fireball (Level 11) ✓ Requested
3. Musketeer (Level 10)
4. Cannon (Level 12)
5. Ice Spirit (Level 13)
6. Skeletons (Level 12)
7. Ice Golem (Level 11)
8. Log (Level 12)

Average Elixir: 3.1

Strategy Notes:
- Hog Rider + Fireball core as requested
- Fireball for spell support and tower damage
- Classic cycle support cards for 3.1 average
- Strong against medium HP troops with Fireball

Upgrade Recommendations:
1. Hog Rider (11→13) - Primary win condition, 50,000 gold
2. Fireball (11→13) - Spell damage scaling, 50,000 gold
3. Musketeer (10→13) - Air defense priority, 50,000 gold
```

## Example 4: Aggro Deck

**User:** "Make me an aggro deck"

**Skill Actions:**
```bash
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy aggro \
  --max-elixir 3.5 \
  --enable-synergy
```

**Expected Output:**
```
Player: klauer (5600 trophies)

Recommended Deck (Aggro):
1. Battle Ram (Level 11)
2. Bandit (Level 10)
3. Magic Archer (Level 11)
4. Royal Ghost (Level 10)
5. Barb Barrel (Level 12)
6. Ice Spirit (Level 13)
7. Fireball (Level 11)
8. Log (Level 12)

Average Elixir: 3.3

Strategy Notes:
- Bridge spam style with Battle Ram and Bandit
- Magic Archer for chip and defensive value
- Royal Ghost for surprise attacks and defense
- Multiple threats require constant opponent response

Offense Patterns:
- Battle Ram at bridge when opponent low on elixir
- Bandit to snipe support troops
- Royal Ghost for un-counterable damage
- Magic Archer behind tank for tower chip

Upgrade Recommendations:
1. Battle Ram (11→13) - Win condition priority, 50,000 gold
2. Bandit (10→13) - Dash damage scaling, 50,000 gold
3. Magic Archer (11→13) - Range damage boost, 50,000 gold
```

## Example 5: Control Deck with Ideal Preview

**User:** "Build me a control deck and show me the ideal version"

**Skill Actions:**
```bash
# Build the deck
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy control \
  --enable-synergy

# Show ideal deck
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy control \
  --enable-synergy \
  --ideal-deck
```

**Expected Output:**
```
Player: klauer (5600 trophies)

Recommended Deck (Control):
1. Golem (Level 10)
2. Night Witch (Level 11)
3. Baby Dragon (Level 11)
4. Lumberjack (Level 10)
5. Lightning (Level 10)
6. Tornado (Level 11)
7. Barbarians (Level 12)
8. Zap (Level 13)

Average Elixir: 4.3

Strategy Notes:
- Beatdown control with Golem as tank
- Build elixir advantage for big pushes
- Night Witch + Baby Dragon behind Golem
- Tornado for defensive value and activation

Defensive Strategy:
- Barbarians for ground defense
- Tornado + Baby Dragon for splash
- Lumberjack for high DPS threats

---

Ideal Deck (with upgrades):
1. Golem (Level 13) ⬆️ +3 levels
2. Night Witch (Level 13) ⬆️ +2 levels
3. Baby Dragon (Level 13) ⬆️ +2 levels
4. Lumberjack (Level 13) ⬆️ +3 levels
5. Lightning (Level 13) ⬆️ +3 levels
6. Tornado (Level 13) ⬆️ +2 levels
7. Barbarians (Level 13) ⬆️ +1 level
8. Zap (Level 13) ✓ Maxed

Total Upgrade Cost: 340,000 gold
Expected Trophy Gain: +300-500 with skill
```

## Example 6: Missing DEFAULT_PLAYER_TAG

**User:** "Build me a deck"

**Skill Actions:**
```bash
# Check .env for DEFAULT_PLAYER_TAG
grep DEFAULT_PLAYER_TAG .env || echo "NOT_FOUND"
# Output: NOT_FOUND (or empty value)
```

**Expected Output:**
```
❌ DEFAULT_PLAYER_TAG not configured

To use this skill, add your player tag to .env:

1. Open .env in your editor
2. Add: DEFAULT_PLAYER_TAG="#YOUR_PLAYER_TAG"
3. Save the file

Or run:
echo 'DEFAULT_PLAYER_TAG="#YOUR_TAG"' >> .env

You can find your player tag in Clash Royale:
- Profile → Name → Copy Tag

Alternatively, use the deck-analysis skill with your tag:
"Analyze decks for player #R8QGUQRCV"
```

## Example 7: Exclude Cards

**User:** "Build me a deck without buildings"

**Skill Actions:**
```bash
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy balanced \
  --exclude-cards "Cannon,Tesla,Bomb Tower,Inferno Tower,Mortar,X-Bow,Goblin Cage,Tombstone" \
  --enable-synergy
```

**Expected Output:**
```
Player: klauer (5600 trophies)

Excluded: All building cards

Recommended Deck (No Buildings):
1. Hog Rider (Level 11)
2. Musketeer (Level 10)
3. Ice Golem (Level 11)
4. Ice Spirit (Level 13)
5. Fireball (Level 11)
6. Skeletons (Level 12)
7. Log (Level 12)
8. Valkyrie (Level 11)

Average Elixir: 3.1

Note: No building means weaker to Hog Rider, Balloon, and Golem.
Use Valkyrie and Musketeer for defensive value.

Upgrade Recommendations:
1. Hog Rider (11→13) - 50,000 gold
2. Musketeer (10→13) - 50,000 gold
3. Valkyrie (11→13) - 50,000 gold
```

## Example 8: Splash Deck

**User:** "Build me a splash deck"

**Skill Actions:**
```bash
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy splash \
  --enable-synergy
```

**Expected Output:**
```
Player: klauer (5600 trophies)

Recommended Deck (Splash):
1. Wizard (Level 10)
2. Valkyrie (Level 11)
3. Baby Dragon (Level 11)
4. Dark Prince (Level 11)
5. Fireball (Level 11)
6. Zap (Level 13)
7. Mini P.E.K.K.A (Level 11)
8. Bats (Level 12)

Average Elixir: 3.6

Strategy Notes:
- Heavy splash damage to counter swarm decks
- Multiple area damage options
- Strong against Graveyard, Goblin Gang, Minion Horde
- Weak to heavy single-target troops (P.E.K.K.A, Mega Knight)

Splash Coverage:
- Wizard: Ground + Air splash
- Valkyrie: 360° ground splash
- Baby Dragon: Air splash tank
- Dark Prince: Charge splash
- Fireball: Spell splash

Upgrade Recommendations:
1. Wizard (10→13) - Splash damage priority, 50,000 gold
2. Valkyrie (11→13) - HP for tankiness, 50,000 gold
3. Baby Dragon (11→13) - Versatile splash, 50,000 gold
```

## Example 9: Synergy-First Deck

**User:** "Build me a deck with the best synergies"

**Skill Actions:**
```bash
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy synergy-first \
  --enable-synergy
```

**Expected Output:**
```
Player: klauer (5600 trophies)

Recommended Deck (Synergy-First):
1. Giant (Level 11)
2. Witch (Level 10)
3. Musketeer (Level 10)
4. Minions (Level 12)
5. Zap (Level 13)
6. Fireball (Level 11)
7. Skeleton Army (Level 11)
8. Cannon (Level 12)

Average Elixir: 3.4

Synergy Analysis:
⭐⭐⭐ Giant + Witch: Tank + splash support
⭐⭐⭐ Witch + Zap: Skeletons survive zap damage
⭐⭐⭐ Giant + Musketeer: Classic beatdown combo
⭐⭐ Musketeer + Cannon: Defensive core
⭐⭐ Minions + Fireball: Spell bait potential

Key Synergies:
- Giant tanks for Witch's skeletons
- Witch behind Giant = constant splash pressure
- Zap protects Witch from small spells
- Minions punish opponent's spell usage

Upgrade Recommendations:
1. Giant (11→13) - Tank HP priority, 50,000 gold
2. Witch (10→13) - Splash damage scaling, 50,000 gold
3. Musketeer (10→13) - DPS increase, 50,000 gold
```

## Example 10: Spell Deck

**User:** "Build me a spell deck"

**Skill Actions:**
```bash
./bin/cr-api deck build \
  --tag "#R8QGUQRCV" \
  --strategy spell \
  --enable-synergy
```

**Expected Output:**
```
Player: klauer (5600 trophies)

Recommended Deck (Spell):
1. Rocket (Level 10)
2. Lightning (Level 10)
3. Fireball (Level 11)
4. Log (Level 12)
5. Zap (Level 13)
6. Tesla (Level 11)
7. Ice Spirit (Level 13)
8. Skeletons (Level 12)

Average Elixir: 3.4

Strategy Notes:
- Spell-cycle win condition
- Rocket + Lightning for tower damage
- Defensive buildings and cheap cards to survive
- Win by spell damage when towers are low

Spell Damage per Rotation:
- Rocket: 493 damage
- Lightning: 446 damage
- Fireball: 325 damage
- Log: 96 damage
- Zap: 64 damage
Total: 1,424 damage per full cycle

Win Condition:
- Survive with Tesla and cheap cards
- Spell cycle when towers are <1,500 HP
- Rocket for positive elixir trades on pumps/troops

Upgrade Recommendations:
1. Rocket (10→13) - Spell damage priority, 50,000 gold
2. Lightning (10→13) - Secondary spell, 50,000 gold
3. Fireball (11→13) - Main spell, 50,000 gold
```
