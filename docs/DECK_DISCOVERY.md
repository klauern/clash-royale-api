# Deck Discovery & Leaderboard System

The Clash Royale API includes a powerful deck discovery system that systematically explores your deck space to find optimal combinations through intelligent evaluation and persistent storage.

## Overview

The deck discovery system provides:

- **Multiple Sampling Strategies**: Smart sampling, random exploration, archetype-focused generation
- **Resumable Sessions**: Start/stop discovery at any time with automatic checkpointing
- **Background Execution**: Run long discovery sessions as daemon processes
- **Persistent Leaderboard**: SQLite-based storage with deduplication and advanced querying
- **Progress Tracking**: Real-time statistics, ETA, and rate monitoring
- **Comprehensive Evaluation**: Multi-category scoring (attack, defense, synergy, versatility, F2P, playability)

## Architecture

### Components

```
┌─────────────────────┐      ┌──────────────────────┐
│  Deck Generator     │─────▶│  Discovery Runner    │
│  - Strategies       │      │  - Progress tracking │
│  - Constraints      │      │  - Checkpointing     │
│  - Iterators        │      │  - Rate limiting     │
└─────────────────────┘      └──────────────────────┘
                                      │
                                      ▼
                              ┌──────────────────────┐
                              │  Deck Evaluator      │
                              │  - Multi-category    │
                              │  - Player context    │
                              │  - Scoring           │
                              └──────────────────────┘
                                      │
                                      ▼
                              ┌──────────────────────┐
                              │  Leaderboard Storage │
                              │  - SQLite database   │
                              │  - Deduplication     │
                              │  - Query system      │
                              └──────────────────────┘
```

### Key Files

| File | Purpose |
|------|---------|
| `cmd/cr-api/discover_commands.go` | CLI commands for discovery sessions |
| `cmd/cr-api/leaderboard_commands.go` | CLI commands for leaderboard queries |
| `pkg/deck/discovery_runner.go` | Core discovery orchestration |
| `pkg/deck/generator.go` | Deck generation strategies |
| `pkg/leaderboard/storage.go` | Persistent SQLite storage |

## Discovery Commands

### Start a Discovery Session

```bash
# Basic discovery with smart sampling (default)
cr-api deck discover start --tag PLAYERTAG

# Specify sampling strategy
cr-api deck discover start --tag PLAYERTAG --strategy random-sample

# Set sample size for non-exhaustive strategies
cr-api deck discover start --tag PLAYERTAG --sample-size 5000

# Limit maximum decks to evaluate
cr-api deck discover start --tag PLAYERTAG --limit 10000

# Verbose progress output
cr-api deck discover start --tag PLAYERTAG --verbose

# Background mode (daemon process)
cr-api deck discover start --tag PLAYERTAG --background
```

**Start Command Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag`, `-p` | string | *required* | Player tag (without #) |
| `--strategy` | string | `smart-sample` | Sampling strategy |
| `--sample-size` | int | `1000` | Number of decks to generate (sampling strategies) |
| `--limit` | int | `0` | Maximum decks to evaluate (0 = unlimited) |
| `--verbose`, `-v` | bool | `false` | Show detailed progress |
| `--background` | bool | `false` | Run as daemon process |

### Resume a Discovery Session

```bash
# Resume from last checkpoint
cr-api deck discover resume --tag PLAYERTAG

# Resume with verbose output
cr-api deck discover resume --tag PLAYERTAG --verbose

# Resume in background
cr-api deck discover resume --tag PLAYERTAG --background
```

**Resume Command Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag`, `-p` | string | *required* | Player tag (without #) |
| `--verbose`, `-v` | bool | `false` | Show detailed progress |
| `--background` | bool | `false` | Run as daemon process |

### Check Discovery Status

```bash
# Quick status check
cr-api deck discover status --tag PLAYERTAG

# Output example:
# Discovery Session Status for #PLAYERTAG
# Strategy: smart-sample
# Evaluated: 1523 decks
# Stored: 892 decks
# Average Score: 7.45
# Best Score: 8.92
# Best Deck: [Hog Rider, Fireball, Ice Spirit, Log, Musketeer, Cannon, Knight, Archers]
# Last Updated: 2024-01-15T14:23:45Z
#
# Resume with: cr-api deck discover resume --tag PLAYERTAG
```

### View Detailed Statistics

```bash
# Full statistics breakdown
cr-api deck discover stats --tag PLAYERTAG

# Output example:
# ╔════════════════════════════════════════════════════════════════════╗
# ║                    DISCOVERY SESSION STATS                         ║
# ╚════════════════════════════════════════════════════════════════════╝
#
# Player: #PLAYERTAG
# Strategy: smart-sample
# Last Updated: 2024-01-15 14:23:45
#
# Progress:
#   Evaluated: 1523 decks
#   Total: 5000 decks
#   Complete: 30.5%
#   Stored: 892 decks in leaderboard
#
# Performance:
#   Elapsed: 25m32s
#   Rate: 1.00 decks/sec
#   ETA: 58m12s
#
# Scores:
#   Average: 7.45
#   Best: 8.92
#   Best Deck: [Hog Rider, Fireball, Ice Spirit, Log, Musketeer, Cannon, Knight, Archers]
#   Top 5 Scores: 8.92, 8.87, 8.85, 8.81, 8.79
#
# Actions:
#   Resume: cr-api deck discover resume --tag PLAYERTAG
#   View leaderboard: cr-api deck leaderboard show --tag PLAYERTAG
```

### Stop a Running Discovery

```bash
# Gracefully stop background discovery
cr-api deck discover stop --tag PLAYERTAG

# Output:
# Stopping discovery session (PID: 12345)...
# Stop signal sent. Discovery will save checkpoint and exit gracefully.
# Use 'cr-api deck discover status --tag PLAYERTAG' to verify checkpoint was saved.
```

**Note**: For foreground discovery, use `Ctrl+C` to stop. The system will automatically save a checkpoint.

## Sampling Strategies

### Smart Sample (Default)

Intelligently prioritizes high-level cards and known synergies for efficient exploration.

```bash
cr-api deck discover start --tag PLAYERTAG --strategy smart-sample --sample-size 5000
```

**Best for**: Finding competitive decks quickly when you have overleveled cards.

**How it works**:
- Weights card selection by level and rarity
- Prioritizes cards with strong synergy potential
- Generates diverse but high-quality combinations

### Random Sample

Explores the deck space through pure random sampling for unbiased discovery.

```bash
cr-api deck discover start --tag PLAYERTAG --strategy random-sample --sample-size 10000
```

**Best for**: Comprehensive exploration when you want to discover unexpected combinations.

**How it works**:
- Uniform random selection from your collection
- No bias toward specific cards or strategies
- Maximum diversity in generated decks

### Archetype Focused

Explores decks within a specific archetype (beatdown, control, cycle, etc.).

```bash
cr-api deck discover start --tag PLAYERTAG --strategy archetype-focused --archetype beatdown
```

**Best for**: Deep dive into a specific playstyle.

**How it works**:
- Generates decks matching archetype patterns
- Enforces archetype-specific constraints
- Explores variations within that archetype

### Exhaustive

Evaluates every possible valid deck combination (warning: can be extremely slow).

```bash
cr-api deck discover start --tag PLAYERTAG --strategy exhaustive
```

**Best for**: Small collections or when you want absolute certainty.

**Warning**: With 80+ cards, this can generate millions of combinations. Use `--limit` to cap evaluation.

## Leaderboard Commands

### Show Top Decks

```bash
# Show top 10 decks (default)
cr-api deck leaderboard show --tag PLAYERTAG

# Show top 20 decks
cr-api deck leaderboard show --tag PLAYERTAG --top 20

# Different output formats
cr-api deck leaderboard show --tag PLAYERTAG --format detailed
cr-api deck leaderboard show --tag PLAYERTAG --format json
cr-api deck leaderboard show --tag PLAYERTAG --format csv

# Save to file
cr-api deck leaderboard show --tag PLAYERTAG --output top_decks.json
```

**Output Formats:**

| Format | Description |
|--------|-------------|
| `summary` (default) | Compact table with rank, score, archetype, elixir, cards |
| `detailed` | Full breakdown with category scores and card list |
| `json` | Machine-readable JSON for programmatic use |
| `csv` | Spreadsheet-compatible CSV export |

**Show Command Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag`, `-p` | string | *required* | Player tag (without #) |
| `--top`, `-n` | int | `10` | Number of top decks to display |
| `--format` | string | `summary` | Output format: summary, detailed, json, csv |
| `--output` | string | stdout | Output file path (optional) |

### Query with Filters

```bash
# Filter by archetype
cr-api deck leaderboard filter --tag PLAYERTAG --archetype cycle --top 10

# Filter by score range
cr-api deck leaderboard filter --tag PLAYERTAG --min-score 8.0 --max-score 9.0

# Filter by elixir range
cr-api deck leaderboard filter --tag PLAYERTAG --min-elixir 2.5 --max-elixir 3.2

# Filter by strategy
cr-api deck leaderboard filter --tag PLAYERTAG --strategy smart-sample

# Require specific cards (ALL must be present)
cr-api deck leaderboard filter --tag PLAYERTAG --require-all "Hog Rider,Log"

# Require specific cards (ANY must be present)
cr-api deck leaderboard filter --tag PLAYERTAG --require-any "Hog Rider,Royal Giant"

# Exclude specific cards
cr-api deck leaderboard filter --tag PLAYERTAG --exclude "Elite Barbarians"

# Sort options
cr-api deck leaderboard filter --tag PLAYERTAG --sort-by attack_score --order desc

# Pagination
cr-api deck leaderboard filter --tag PLAYERTAG --limit 20 --offset 40

# Combine multiple filters
cr-api deck leaderboard filter \
  --tag PLAYERTAG \
  --archetype beatdown \
  --min-elixir 3.5 \
  --max-elixir 4.2 \
  --require-all "Giant" \
  --sort-by overall_score \
  --top 15
```

**Filter Command Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag`, `-p` | string | *required* | Player tag (without #) |
| `--archetype` | string | - | Filter by archetype |
| `--min-score` | float | - | Minimum overall score (0-10) |
| `--max-score` | float | - | Maximum overall score (0-10) |
| `--min-elixir` | float | - | Minimum average elixir |
| `--max-elixir` | float | - | Maximum average elixir |
| `--strategy` | string | - | Filter by generation strategy |
| `--require-all` | strings | - | Decks must contain ALL cards |
| `--require-any` | strings | - | Decks must contain ANY cards |
| `--exclude` | strings | - | Exclude decks with ANY cards |
| `--sort-by` | string | `overall_score` | Sort field |
| `--order` | string | `desc` | Sort order: asc, desc |
| `--limit`, `-n` | int | `10` | Maximum results |
| `--offset` | int | `0` | Results to skip (pagination) |
| `--format` | string | `summary` | Output format |
| `--output` | string | stdout | Output file path |

**Sort Fields:**
- `overall_score` (default)
- `attack_score`
- `defense_score`
- `synergy_score`
- `versatility_score`
- `f2p_score`
- `playability_score`
- `avg_elixir`

### View Statistics

```bash
# Basic stats
cr-api deck leaderboard stats --tag PLAYERTAG

# Include archetype distribution
cr-api deck leaderboard stats --tag PLAYERTAG --archetypes

# Output example:
# ╔════════════════════════════════════════════════════════════════════╗
# ║                      LEADERBOARD STATS                             ║
# ╚════════════════════════════════════════════════════════════════════╝
#
# Player: #PLAYERTAG
# Total Decks Stored: 892
# Total Evaluated: 1523
# Last Updated: 2024-01-15 14:23:45
#
# Score Statistics:
#   Top Score: 8.92
#   Avg Score: 7.45
#
# Database Size: 2.34 MB
#
# Archetype Distribution:
# ════════════════════════════
# Archetype      Count    Percentage
# ---------      -----    ----------
# Cycle          245      27.5%
# Beatdown       198      22.2%
# Control        167      18.7%
# Bridge Spam    134      15.0%
# Spellbait      98       11.0%
# Other          50       5.6%
```

### Export Leaderboard

```bash
# Export to JSON
cr-api deck leaderboard export --tag PLAYERTAG --format json --output all_decks.json

# Export to CSV
cr-api deck leaderboard export --tag PLAYERTAG --format csv --output decks.csv

# Exported 892 decks to all_decks.json
```

### Clear Leaderboard

```bash
# Clear with confirmation prompt
cr-api deck leaderboard clear --tag PLAYERTAG

# Skip confirmation (use with caution)
cr-api deck leaderboard clear --tag PLAYERTAG --confirm

# This will delete all 892 decks from the leaderboard for player #PLAYERTAG
# Are you sure? (y/N): y
# Cleared 892 decks from leaderboard for player #PLAYERTAG
```

## Background Mode

Run long discovery sessions as daemon processes that continue after you close the terminal.

### Starting Background Discovery

```bash
# Start in background
cr-api deck discover start --tag PLAYERTAG --background

# Output:
# Discovery started in background (PID: 12345)
# Log file: /Users/username/.cr-api/discover/PLAYERTAG.log
#
# Commands:
#   Status: cr-api deck discover status --tag PLAYERTAG
#   Stop: cr-api deck discover stop --tag PLAYERTAG
#   Stats: cr-api deck discover stats --tag PLAYERTAG
```

### Monitoring Background Sessions

```bash
# Check status
cr-api deck discover status --tag PLAYERTAG

# View detailed stats
cr-api deck discover stats --tag PLAYERTAG

# Tail the log file
tail -f ~/.cr-api/discover/PLAYERTAG.log
```

### Stopping Background Sessions

```bash
# Graceful shutdown
cr-api deck discover stop --tag PLAYERTAG
```

## Data Storage

### File Locations

```
~/.cr-api/
├── discover/
│   ├── PLAYERTAG.json     # Checkpoint file
│   ├── PLAYERTAG.pid      # Background process PID
│   └── PLAYERTAG.log      # Background process log
└── leaderboards/
    └── PLAYERTAG.db       # SQLite database
```

### Database Schema

**Decks Table:**
```sql
CREATE TABLE decks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    deck_hash TEXT NOT NULL UNIQUE,           -- SHA256 for deduplication
    cards TEXT NOT NULL,                      -- JSON array of card names
    overall_score REAL NOT NULL,
    attack_score REAL NOT NULL,
    defense_score REAL NOT NULL,
    synergy_score REAL NOT NULL,
    versatility_score REAL NOT NULL,
    f2p_score REAL NOT NULL,
    playability_score REAL NOT NULL,
    archetype TEXT NOT NULL,
    archetype_conf REAL NOT NULL,
    strategy TEXT,
    avg_elixir REAL NOT NULL,
    evaluated_at DATETIME NOT NULL,
    player_tag TEXT NOT NULL,
    evaluation_version TEXT NOT NULL
);
```

**Indexes for Performance:**
- `idx_overall_score`: Fast top-N queries
- `idx_archetype`: Archetype filtering
- `idx_strategy`: Strategy filtering
- `idx_archetype_score`: Combined archetype + score queries
- `idx_strategy_score`: Combined strategy + score queries

### Checkpoint Format

```json
{
  "generator_checkpoint": {
    "strategy": "smart-sample",
    "position": 1234,
    "generated": 5000,
    "state": { /* strategy-specific state */ }
  },
  "stats": {
    "evaluated": 1523,
    "total": 5000,
    "stored": 892,
    "top_scores": [8.92, 8.87, 8.85, 8.81, 8.79],
    "avg_score": 7.45,
    "rate": 1.0,
    "eta": 3492000000000,
    "best_deck": ["Hog Rider", "Fireball", ...],
    "best_score": 8.92,
    "start_time": "2024-01-15T13:50:00Z",
    "elapsed": 1532000000000,
    "strategy": "smart-sample",
    "player_tag": "PLAYERTAG"
  },
  "timestamp": "2024-01-15T14:23:45Z",
  "player_tag": "PLAYERTAG",
  "strategy": "smart-sample"
}
```

## Complete Workflow Example

### 1. Initial Discovery

```bash
# Start a smart sample discovery
cr-api deck discover start --tag R8QGUQRCV \
  --strategy smart-sample \
  --sample-size 5000 \
  --verbose
```

### 2. Monitor Progress

```bash
# In another terminal, check status
cr-api deck discover stats --tag R8QGUQRCV
```

### 3. Pause and Resume

```bash
# Stop the discovery (Ctrl+C or stop command)
cr-api deck discover stop --tag R8QGUQRCV

# Later, resume from checkpoint
cr-api deck discover resume --tag R8QGUQRCV --verbose
```

### 4. Query Results

```bash
# View top decks
cr-api deck leaderboard show --tag R8QGUQRCV --top 20 --format detailed

# Filter for cycle decks
cr-api deck leaderboard filter \
  --tag R8QGUQRCV \
  --archetype cycle \
  --min-elixir 2.5 \
  --max-elixir 3.0 \
  --top 10

# Export for analysis
cr-api deck leaderboard export \
  --tag R8QGUQRCV \
  --format csv \
  --output cycle_decks.csv
```

### 5. Analyze Archetypes

```bash
# View archetype distribution
cr-api deck leaderboard stats --tag R8QGUQRCV --archetypes

# Deep dive into beatdown archetype
cr-api deck leaderboard filter \
  --tag R8QGUQRCV \
  --archetype beatdown \
  --format detailed \
  --top 15
```

## Performance Considerations

### Rate Limiting

The discovery system is rate-limited to **1 deck per second** to respect API limits and ensure stability.

### Evaluation Speed

| Strategy | Throughput | Best For |
|----------|-----------|----------|
| Smart Sample | ~1 deck/sec | Quick competitive results |
| Random Sample | ~1 deck/sec | Unbiased exploration |
| Archetype Focused | ~1 deck/sec | Targeted archetype search |
| Exhaustive | Varies | Small collections only |

### Storage Growth

- **Per deck**: ~500 bytes (including indexes)
- **10,000 decks**: ~5 MB
- **100,000 decks**: ~50 MB

### Checkpoint Overhead

Checkpoints are saved:
- On `Ctrl+C` / SIGTERM
- Every context cancellation
- On graceful shutdown

Checkpoint size: ~1-2 KB

## Best Practices

1. **Start with Smart Sample**: Get competitive results quickly
2. **Use Background Mode**: For sessions longer than a few minutes
3. **Check Status Regularly**: Monitor progress and ETA
4. **Resume Interrupted Sessions**: Never lose progress
5. **Filter Results**: Use leaderboard filters to find specific decks
6. **Export for Analysis**: Use JSON/CSV for external analysis
7. **Clear Periodically**: Remove old data to manage database size
8. **Use Limit for Testing**: Set `--limit` to test before full runs

## Troubleshooting

### "No checkpoint found"

```bash
# Start a fresh discovery
cr-api deck discover start --tag PLAYERTAG
```

### "Discovery already running"

```bash
# Stop existing background process
cr-api deck discover stop --tag PLAYERTAG

# Then start new discovery
cr-api deck discover start --tag PLAYERTAG
```

### Empty leaderboard results

```bash
# Check discovery status
cr-api deck discover stats --tag PLAYERTAG

# If no decks stored, discovery may not have run yet
# Start discovery first
cr-api deck discover start --tag PLAYERTAG
```

### Database locked

```bash
# Ensure no other discovery is running
# Only one process can write to a leaderboard at a time

# Check for background process
cr-api deck discover stop --tag PLAYERTAG
```

## Advanced Usage

### Custom Evaluation

The discovery system uses a pluggable `DeckEvaluator` interface. See `pkg/deck/discovery_runner.go` for implementation details.

### Custom Strategies

Add new strategies by extending `GeneratorStrategy` in `pkg/deck/generator.go`.

### Direct API Usage

```go
import (
    "github.com/klauer/clash-royale-api/go/pkg/deck"
    "github.com/klauer/clash-royale-api/go/pkg/leaderboard"
)

// Create storage
storage, err := leaderboard.NewStorage("PLAYERTAG")
defer storage.Close()

// Create generator
generator, err := deck.NewDeckGenerator(deck.GeneratorConfig{
    Strategy:   deck.StrategySmartSample,
    Candidates: candidates,
    SampleSize: 1000,
})

// Create discovery runner
runner, err := deck.NewDiscoveryRunner(deck.DiscoveryConfig{
    GeneratorConfig: generator.Config(),
    Storage:         storage,
    Evaluator:       &myEvaluator{},
    PlayerTag:       "PLAYERTAG",
})

// Run discovery
err = runner.Run(context.Background())
```

## Related Documentation

- [DECK_BUILDER.md](DECK_BUILDER.md) - Deck building strategies and algorithms
- [DECK_ANALYSIS_SUITE.md](DECK_ANALYSIS_SUITE.md) - Comprehensive deck analysis workflows
- [CLI_REFERENCE.md](CLI_REFERENCE.md) - Complete CLI command reference
- [TESTING.md](TESTING.md) - Testing deck discovery components
