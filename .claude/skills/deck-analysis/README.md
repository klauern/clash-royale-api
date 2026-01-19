# Deck Analysis Skill for Clash Royale

A Claude Code skill that generates comprehensive deck recommendations based on your card collection.

## Quick Start

Simply ask Claude to analyze decks:

```
Analyze decks for player #R8QGUQRCV
Build me strong decks with my Archers evolution
Recommend 1v1 ladder decks for my account
```

## What You Get

### Comprehensive Analysis
- **18+ deck variations** across 6 strategies (balanced, aggro, control, cycle, splash, spell)
- **Detailed scoring** on attack, defense, synergy, versatility, and F2P friendliness
- **Ranked recommendations** with archetype identification
- **Upgrade priorities** with gold costs and impact analysis

### Strategy-Specific Builds
- **Cycle decks**: Ultra-fast rotation (2.3-3.0 elixir average)
- **Control decks**: Defensive powerhouses (3.5-4.5 elixir)
- **Aggro decks**: High-pressure offensive builds
- **Spell decks**: Spell-cycling win conditions
- **And more**: Balanced and splash-focused variations

### Evolution Support
- Automatically incorporates your unlocked evolutions
- Builds targeted decks around evolution cards
- Optimizes evolution slot allocation

## How It Works

1. **Fetches your player data** from Clash Royale API
2. **Generates deck variations** using advanced building algorithms
3. **Evaluates each deck** across multiple dimensions
4. **Ranks and compares** top performers
5. **Provides upgrade recommendations** for maximum improvement

## Output Files

All analysis results are saved to `data/analysis/`:

- `decks/` - Deck suite summaries (JSON)
- `evaluations/` - Detailed scoring data (JSON)
- `reports/` - Human-readable analysis reports (Markdown)

## Advanced Usage

### Specify Strategies
```
Build cycle and control decks for #ABC123
```

### Include Specific Cards
```
Build a deck with Hog Rider and Valkyrie for #XYZ789
```

### Elixir Constraints
```
Build low-elixir decks (max 3.0) for #DEF456
```

### With Evolutions
```
Analyze decks with my unlocked Archers evolution for #R8QGUQRCV
```

## Understanding Scores

### Overall Score (0-10)
Combined rating across all dimensions:
- **8-10**: Excellent - Tournament-ready
- **7-8**: Great - Strong ladder deck
- **6-7**: Good - Solid choice
- **5-6**: Mediocre - Needs work
- **<5**: Weak - Major issues

### Score Components
- **Attack**: Offensive pressure and win condition strength
- **Defense**: Defensive capabilities and versatility
- **Synergy**: Card interactions and combo potential
- **Versatility**: Role coverage and adaptability
- **F2P Friendly**: Upgrade costs and rarity distribution

## Tips

1. **Start with comprehensive analysis** - Run with `--strategies all` to see all options
2. **Consider your playstyle** - Aggressive? Try aggro. Defensive? Go control.
3. **Factor in card levels** - Higher-level cards score better
4. **Follow upgrade recommendations** - Prioritize cards with high value/gold ratios
5. **Try multiple strategies** - Different decks work better against different opponents

## Requirements

- Valid Clash Royale API token in `.env`
- Player tag (find in-game or via clan search)
- Built `cr-api` binary (`task build`)

## Troubleshooting

**"flag provided but not defined"** - Check flag names match current CLI version

**"Player not found"** - Verify player tag format (with or without #)

**"No decks generated"** - Check elixir constraints and card filters

**Empty results** - Ensure player has sufficient cards unlocked

## Related Commands

- `./bin/cr-api player get --tag <TAG>` - View player stats
- `./bin/cr-api deck build` - Build single deck with custom constraints
- `./bin/cr-api deck analyze-suite` - Full analysis workflow
- `task build` - Rebuild binary after code changes
