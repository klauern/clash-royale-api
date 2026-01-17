# Deck Fuzzing

The deck fuzzing feature provides Monte Carlo-style random deck generation and evaluation to discover top-performing decks from a player's card collection.

## Overview

Deck fuzzing generates hundreds or thousands of random valid deck combinations from your available cards, evaluates each deck using the comprehensive scoring system, and returns the top performers based on your criteria.

## Usage

### Basic Usage

```bash
# Generate 1000 random decks and show top 10
./bin/cr-api deck fuzz --tag R8QGUQRCV
```

### With Constraints

```bash
# Generate 5000 decks with specific constraints
./bin/cr-api deck fuzz --tag R8QGUQRCV \
  --count 5000 \
  --include-cards "Royal Giant" \
  --max-elixir 3.5 \
  --top 20 \
  --sort-by synergy
```

### Parallel Processing

```bash
# Use 4 parallel workers for faster generation
./bin/cr-api deck fuzz --tag R8QGUQRCV \
  --count 10000 \
  --workers 4 \
  --verbose
```

### Output to File

```bash
# Save results to JSON file
./bin/cr-api deck fuzz --tag R8QGUQRCV \
  --count 1000 \
  --format json \
  --output-dir data/fuzz-results
```

### Offline Mode (from existing analysis)

```bash
# Use existing analysis file instead of API
./bin/cr-api deck fuzz --tag R8QGUQRCV \
  --from-analysis \
  --analysis-dir data/analysis \
  --count 1000
```

## Command Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag`, `-p` | string | (required) | Player tag (without #) |
| `--count` | int | 1000 | Number of random decks to generate |
| `--workers` | int | 1 | Number of parallel workers |
| `--include-cards` | string[] | - | Cards that must be in every deck |
| `--exclude-cards` | string[] | - | Cards to exclude from all decks |
| `--min-elixir` | float | 0.0 | Minimum average elixir |
| `--max-elixir` | float | 10.0 | Maximum average elixir |
| `--min-overall` | float | 0.0 | Minimum overall score (0-10) |
| `--min-synergy` | float | 0.0 | Minimum synergy score (0-10) |
| `--top` | int | 10 | Number of top decks to display |
| `--sort-by` | string | overall | Sort criteria (overall, attack, defense, synergy, versatility, elixir) |
| `--format` | string | summary | Output format (summary, json, csv, detailed) |
| `--output-dir` | string | - | Directory to save results |
| `--verbose`, `-v` | bool | false | Show detailed progress |
| `--from-analysis` | bool | false | Load from analysis file (offline) |
| `--analysis-file` | string | - | Specific analysis file path |
| `--analysis-dir` | string | - | Directory of analysis files |
| `--seed` | int | 0 | Random seed (0 = random) |
| `--storage` | string | - | Path to persistent storage database |

## Output Formats

### Summary (default)

Human-readable table with top decks:

```
Deck Fuzzing Results for PlayerName (#R8QGUQRCV)
Generated 1000 random decks in 2.5s
Configuration:
  Elixir range: 0.0 - 10.0

Top 10 Decks (from 850 decks passing filters):

Rank  Deck                                              Overall  Attack  Defense  Synergy  Elixir
----  ------------------------------------------------- ------- ------- ------- ------- -------
1     Hog Rider, Ice Spirit, Skeletons, Archers, ...    8.45    8.20    8.10    8.90    2.85
2     Royal Giant, Zap, Cannon, Archers, ...           8.32    8.50    7.90    8.45    3.10
...
```

### JSON

Machine-readable format with full results:

```json
{
  "player_name": "PlayerName",
  "player_tag": "#R8QGUQRCV",
  "generated": 1000,
  "success": 950,
  "failed": 50,
  "filtered": 850,
  "returned": 10,
  "generation_time_seconds": 2.5,
  "config": {
    "count": 1000,
    "workers": 1,
    "include_cards": ["Royal Giant"],
    "exclude_cards": [],
    "min_avg_elixir": 0.0,
    "max_avg_elixir": 10.0
  },
  "results": [...]
}
```

### CSV

Tabular format for spreadsheet analysis:

```csv
Rank,Deck,Overall,Attack,Defense,Synergy,Versatility,AvgElixir,Archetype
1,"Hog Rider, Ice Spirit, ...",8.45,8.20,8.10,8.90,7.80,2.85,cycle
2,"Royal Giant, Zap, ...",8.32,8.50,7.90,8.45,7.65,3.10,control
```

### Detailed

Full evaluation output for each top deck:

```
Deck Fuzzing Results for PlayerName (#R8QGUQRCV)

Top 10 Decks:

=== Deck 1 ===
Cards: Hog Rider, Ice Spirit, Skeletons, Archers, Cannon, Log, Musketeer, Valkyrie
Overall: 8.45 | Attack: 8.20 | Defense: 8.10 | Synergy: 8.90 | Versatility: 7.80
Avg Elixir: 2.85 | Archetype: cycle (85% confidence)
Evaluated: 2024-01-15T10:30:00Z

=== Deck 2 ===
...
```

## Algorithm

The fuzzer uses intelligent random sampling to generate valid decks:

1. **Role-based Selection**: Selects cards by strategic role:
   - 1 Win Condition (e.g., Hog Rider, Royal Giant)
   - 1 Building (e.g., Cannon, Tesla)
   - 1 Big Spell (e.g., Fireball, Poison)
   - 1 Small Spell (e.g., Zap, Log)
   - 2 Support troops (e.g., Musketeer, Valkyrie)
   - 2 Cycle cards (e.g., Skeletons, Ice Spirit)

2. **Weighted Random Sampling**: Cards are selected randomly with bias toward higher-level cards

3. **Constraint Validation**: Each generated deck is validated for:
   - Exactly 8 cards
   - Average elixir within range
   - All include cards present
   - No excluded cards present

4. **Retry Logic**: Failed generations are retried up to 100 times

## Performance

- **Single Worker**: ~500-1000 decks/second
- **Parallel Workers**: Near-linear scaling (4 workers ≈ 4x speed)
- **Memory**: ~100MB for 10,000 deck generation

## Examples

### Find Best Royal Giant Decks

```bash
./bin/cr-api deck fuzz --tag R8QGUQRCV \
  --include-cards "Royal Giant" \
  --count 5000 \
  --top 20 \
  --sort-by overall
```

### Find Fast Cycle Decks

```bash
./bin/cr-api deck fuzz --tag R8QGUQRCV \
  --max-elixir 2.8 \
  --count 3000 \
  --top 15 \
  --sort-by elixir
```

### Find High Synergy Decks

```bash
./bin/cr-api deck fuzz --tag R8QGUQRCV \
  --min-synergy 8.0 \
  --count 5000 \
  --top 20 \
  --sort-by synergy
```

### Reproducible Results

```bash
# Use the same seed for reproducible results
./bin/cr-api deck fuzz --tag R8QGUQRCV \
  --seed 12345 \
  --count 1000
```

### Save All Results for Analysis

```bash
./bin/cr-api deck fuzz --tag R8QGUQRCV \
  --count 10000 \
  --format json \
  --output-dir data/fuzz-results
```

## Troubleshooting

### "No decks passed score filters"

Lower your score thresholds:

```bash
./bin/cr-api deck fuzz --tag R8QGUQRCV \
  --min-overall 5.0 \
  --min-synergy 5.0
```

### "Failed to generate deck"

Your card collection may be too limited. Try:
- Reducing elixir constraints
- Removing include card requirements
- Checking you have at least 8 cards

### Slow Generation

Use parallel workers:

```bash
./bin/cr-api deck fuzz --tag R8QGUQRCV \
  --workers 4 \
  --count 10000
```

### API Rate Limiting

Use offline mode with existing analysis:

```bash
./bin/cr-api deck fuzz --tag R8QGUQRCV \
  --from-analysis \
  --analysis-dir data/analysis
```

## Integration with Storage

The fuzz command can save evaluated decks to persistent storage:

```bash
./bin/cr-api deck fuzz --tag R8QGUQRCV \
  --storage data/decks.db \
  --count 10000
```

This allows you to:
- Build a database of evaluated decks over time
- Query for best decks across multiple runs
- Track deck performance trends

## See Also

- [Deck Building](DECK_BUILDER.md) - Intelligent deck building
- [Deck Evaluation](CLI_REFERENCE.md#deck-evaluate) - Single deck evaluation
- [Deck Analysis Suite](DECK_ANALYSIS_SUITE.md) - Multi-strategy deck generation
