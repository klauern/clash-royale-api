# Configuration Files

This directory contains JSON configuration files for customizing various aspects of the Clash Royale API client.

## archetypes.json

The `archetypes.json` file defines deck archetypes used for upgrade impact analysis. You can customize this file to add, remove, or modify archetypes.

### Schema

```json
{
  "version": 1,
  "archetypes": [
    {
      "name": "Archetype Name",
      "win_condition": "Card Name",
      "support_cards": ["Card1", "Card2", "Card3"],
      "min_elixir": 2.5,
      "max_elixir": 3.5,
      "category": "beatdown|cycle|siege|bridge_spam|bait|control",
      "enabled": true,
      "preferred_strategy": "aggro|cycle|control|balanced"
    }
  ]
}
```

### Fields

- **name** (required): Human-readable archetype name
- **win_condition** (required): Primary win condition card
- **support_cards** (optional): List of supporting cards commonly used in this archetype
- **min_elixir** (required): Minimum average elixir cost (0.0-10.0)
- **max_elixir** (required): Maximum average elixir cost (0.0-10.0)
- **category** (optional): Archetype category for grouping and filtering
  - `beatdown`: Heavy tank-based decks
  - `cycle`: Fast, low-elixir decks
  - `siege`: X-Bow/Mortar decks
  - `bridge_spam`: Aggressive pressure decks
  - `bait`: Spell bait decks
  - `control`: Defensive, counter-push decks
- **enabled** (required): Whether to include this archetype in analysis (true/false)
- **preferred_strategy** (optional): Recommended deck builder strategy

### Usage

#### Using the default archetypes

By default, the CLI uses embedded archetypes. No configuration file is needed.

```bash
./bin/cr-api upgrade-impact --tag PLAYERTAG
```

#### Using custom archetypes

Create a custom `archetypes.json` file and pass it with the `--archetypes-file` flag:

```bash
./bin/cr-api upgrade-impact --tag PLAYERTAG --archetypes-file ./config/archetypes.json
```

#### Disabling specific archetypes

To disable an archetype without deleting it, set `"enabled": false`:

```json
{
  "name": "Golem Beatdown",
  "win_condition": "Golem",
  "support_cards": ["Baby Dragon", "Night Witch"],
  "min_elixir": 3.5,
  "max_elixir": 4.5,
  "category": "beatdown",
  "enabled": false
}
```

#### Adding custom archetypes

Add new archetype objects to the `archetypes` array following the schema above.

### Benefits

1. **User Customization**: Tailor archetypes to your playstyle or meta
2. **Community Sharing**: Share archetype databases with other players
3. **Easy Updates**: Modify archetypes without recompiling the code
4. **Focused Analysis**: Enable/disable archetypes for targeted upgrade recommendations

### Validation

The loader validates that:
- Required fields are present (`name`, `win_condition`, `min_elixir`, `max_elixir`, `enabled`)
- Elixir costs are in valid range (0.0-10.0)
- `min_elixir` ≤ `max_elixir`
- At least one archetype is enabled

Invalid configurations will result in an error with details about what's wrong.
