# Deck CLI Contract

Canonical command contract for deck analyze/optimize/recommend/batch surfaces.

## Scope

- `cr-api deck analyze`
- `cr-api deck optimize`
- `cr-api deck recommend`
- `cr-api deck evaluate-batch`

## Canonical Inputs and Behavior

### `cr-api deck analyze`

- Required: `--deck` (8-card dash-separated deck string)
- Optional: `--format` (`human`, `json`; default `human`)
- Behavior: evaluates the provided deck and renders analysis output.

### `cr-api deck optimize`

- Required: `--deck` (8-card dash-separated deck string)
- Optional: `--tag`, `--suggestions`, `--focus`, `--export-csv`
- Optional global flags consumed: `--api-token`, `--data-dir`, `--verbose`
- Behavior:
  - Evaluates the current deck.
  - Generates replacement suggestions.
  - Uses player collection constraints only when both `--tag` and `--api-token` are provided.
  - Applies `--focus` to suggestion ranking.

### `cr-api deck recommend`

- Optional core inputs: `--tag`, `--archetype`, `--count`, `--include-unowned`, `--export-csv`
- Optional environment/context: `--api-token`, `--data-dir`, `--verbose`
- Offline mode:
  - `--from-analysis`
  - `--analysis-dir`
  - `--analysis-file`
- Additional filters:
  - `--arena`
  - `--league`
- Behavior:
  - Online mode requires API token.
  - Offline mode loads analysis data without API access.
  - `--archetype` filters results to matching archetype.
  - `--include-unowned=false` filters out decks containing cards absent from analyzed collection.
  - `--count` controls max results.

### `cr-api deck evaluate-batch`

- Canonical batch evaluator for suite/directory-driven workflows.
- Supports `--from-suite` or `--deck-dir` as primary mutually exclusive inputs.

## Contract Rule

- No declared flag should be read by runtime if missing from command/global flag definitions.
