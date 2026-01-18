# Data Directory

This directory contains generated output from CLI commands. **All files except the examples are temporary and can be safely deleted.**

## Tracked Files (Keep These)
- `upgrade_plan_example.json` - Example upgrade plan configuration
- `evolution_shards.example.json` - Example evolution shards configuration
- `static/cards_stats.json` - Static card statistics

## Generated Directories (Temporary)
- `analysis/` - Player analysis JSON files (`analyze --save`)
- `decks/` - Generated deck files (`deck build --save`)
- `csv/` - CSV exports (`export` commands)
- `evaluations/` - Deck evaluation results
- `reports/` - Analysis reports
- `players/` - Cached player data
- `static/cards.json` - Auto-cached card database

## Cleanup
Run `task clean-data` to remove all generated files while preserving examples.

## Retention Policy
Generated files are not tracked in git. Clean up periodically to avoid clutter:
- Analysis files: Keep if actively developing, otherwise delete
- Deck files: Regenerate as needed
- CSV exports: Archive important ones, delete the rest
- Cache files: Safe to delete, will regenerate on next command
