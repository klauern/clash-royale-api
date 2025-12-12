#!/usr/bin/env python3
"""
Custom deck builder for aggressive fast-cycle playstyles.
Creates 3 deck variations with different win conditions.
"""
import json
from pathlib import Path
from datetime import datetime


def load_player_data(player_file: Path) -> dict:
    """Load player data from saved JSON."""
    with open(player_file, 'r') as f:
        return json.load(f)


def load_analysis_data(analysis_file: Path) -> dict:
    """Load card analysis data."""
    with open(analysis_file, 'r') as f:
        return json.load(f)


def get_card_info(card_name: str, player_data: dict) -> dict:
    """Get card level and elixir cost from player data."""
    for card in player_data.get('cards', []):
        if card['name'] == card_name:
            return {
                'name': card['name'],
                'level': card['level'],
                'maxLevel': card.get('maxLevel', 13),
                'elixir': card.get('elixirCost', 0),
                'rarity': card.get('rarity', 'common')
            }
    return None


def calculate_avg_elixir(deck: list[dict]) -> float:
    """Calculate average elixir cost."""
    total = sum(card['elixir'] for card in deck if card)
    return round(total / len(deck), 2) if deck else 0


def build_hog_cycle_deck(player_data: dict) -> dict:
    """Build aggressive Hog Rider cycle deck."""
    deck_cards = [
        'Hog Rider',      # Win condition (4 elixir)
        'Cannon',         # Defensive building (3 elixir)
        'Fireball',       # Big spell (4 elixir) - REQUIRED
        'Zap',            # Small spell (2 elixir) - REQUIRED
        'Musketeer',      # Ranged support (4 elixir)
        'Knight',         # Tank/defensive cycle (3 elixir)
        'Fire Spirit',    # Cycle (1 elixir)
        'Skeletons'       # Ultra-cheap cycle (1 elixir)
    ]

    deck = [get_card_info(card, player_data) for card in deck_cards]
    deck = [c for c in deck if c]  # Remove None values

    return {
        'deck_name': 'Hog Cycle (Fast Aggression)',
        'win_condition': 'Hog Rider',
        'deck': [c['name'] for c in deck],
        'deck_detail': deck,
        'average_elixir': calculate_avg_elixir(deck),
        'strategy': 'Constant pressure with Hog Rider. Defend efficiently with Cannon and Knight, then counter-push. Out-cycle opponent counters with cheap cards. Use Fireball for pumps/buildings.',
        'strengths': [
            'Very low elixir cost enables constant cycling',
            'Hog Rider applies consistent tower pressure',
            'Strong defensive synergy with Cannon + Musketeer',
            'Can out-cycle most counters'
        ],
        'weaknesses': [
            'Hog Rider is underleveled (6) compared to opponents at your trophy range',
            'Vulnerable to heavy swarm if Fireball/Zap are out of rotation',
            'Requires precise timing and elixir management'
        ],
        'playstyle': 'Aggressive - Fast Cycle',
        'created_at': datetime.now().isoformat()
    }


def build_battle_ram_deck(player_data: dict) -> dict:
    """Build Battle Ram cycle deck."""
    deck_cards = [
        'Battle Ram',         # Win condition (4 elixir)
        'Goblin Cage',        # Defensive building with synergy (4 elixir)
        'Fireball',           # Big spell (4 elixir) - REQUIRED
        'Zap',                # Small spell (2 elixir) - REQUIRED
        'Musketeer',          # Ranged support (4 elixir)
        'Goblin Gang',        # Defensive cycle (3 elixir) - HIGHEST LEVEL
        'Fire Spirit',        # Cycle (1 elixir)
        'Bats'                # Cycle (2 elixir)
    ]

    deck = [get_card_info(card, player_data) for card in deck_cards]
    deck = [c for c in deck if c]

    return {
        'deck_name': 'Battle Ram Cycle (Balanced Aggression)',
        'win_condition': 'Battle Ram',
        'deck': [c['name'] for c in deck],
        'deck_detail': deck,
        'average_elixir': calculate_avg_elixir(deck),
        'strategy': 'Build small pushes with Battle Ram. Use Goblin Cage defensively - the Brawler that spawns can join your counter-push. Level 7 Battle Ram is better leveled than Hog Rider.',
        'strengths': [
            'Battle Ram at level 7 is well-leveled for your trophy range',
            'Goblin Cage provides excellent defensive value + spawns Brawler',
            'Goblin Gang (level 11) is your strongest card',
            'Good spell coverage with Fireball/Zap'
        ],
        'weaknesses': [
            'Slightly higher elixir than pure cycle decks',
            'Battle Ram can be countered before reaching tower',
            'Requires support troops to protect the ram'
        ],
        'playstyle': 'Aggressive - Balanced Cycle',
        'created_at': datetime.now().isoformat()
    }


def build_goblin_barrel_bait_deck(player_data: dict) -> dict:
    """Build spell bait deck using highest level cards."""
    deck_cards = [
        'Goblin Barrel',      # Win condition (3 elixir)
        'Cannon',             # Defensive building (3 elixir)
        'Fireball',           # Big spell (4 elixir) - REQUIRED
        'Zap',                # Small spell (2 elixir) - REQUIRED
        'Goblin Gang',        # Bait + Defense (3 elixir) - HIGHEST LEVEL
        'Skeleton Dragons',   # Splash support (4 elixir) - LEVEL 11
        'Archers',            # Ranged support (3 elixir) - LEVEL 10
        'Fire Spirit'         # Cycle (1 elixir) - LEVEL 10
    ]

    deck = [get_card_info(card, player_data) for card in deck_cards]
    deck = [c for c in deck if c]

    return {
        'deck_name': 'Goblin Barrel Bait (High-Level Cards)',
        'win_condition': 'Goblin Barrel',
        'deck': [c['name'] for c in deck],
        'deck_detail': deck,
        'average_elixir': calculate_avg_elixir(deck),
        'strategy': 'Bait out opponent spells with Goblin Gang and Skeleton Dragons, then punish with Goblin Barrel. Uses your highest-level cards for maximum defensive strength.',
        'strengths': [
            'Uses your 4 highest-level cards (Goblin Gang 11, Skeleton Dragons 11, Archers 10, Fire Spirit 10)',
            'Strong spell bait synergy forces difficult decisions',
            'Very low elixir for constant pressure',
            'Excellent defensive card quality'
        ],
        'weaknesses': [
            'Goblin Barrel is only level 4 (Epic, harder to upgrade)',
            'Heavily spell-dependent - if Barrel is hard-countered, win condition struggles',
            'Requires precise prediction and timing',
            'Vulnerable to heavy spell decks (Lightning, Rocket)'
        ],
        'playstyle': 'Aggressive - Spell Bait',
        'created_at': datetime.now().isoformat()
    }


def save_deck(deck_data: dict, output_dir: Path, player_tag: str):
    """Save deck to JSON file."""
    output_dir.mkdir(parents=True, exist_ok=True)
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    deck_name_slug = deck_data['deck_name'].lower().replace(' ', '_').replace('(', '').replace(')', '')
    filename = f"{timestamp}_{deck_name_slug}_{player_tag}.json"
    path = output_dir / filename

    with open(path, 'w') as f:
        json.dump(deck_data, f, indent=2)

    return path


def main():
    """Generate all custom decks."""
    player_file = Path('data/players/#R8QGUQRCV.json')
    analysis_file = Path('data/analysis/20251208_174559_analysis_R8QGUQRCV.json')
    output_dir = Path('data/decks')

    player_data = load_player_data(player_file)
    player_tag = player_data.get('tag', '').lstrip('#')

    print("=== Building Custom Aggressive Decks for ZyLogan ===\n")

    # Build three deck variations
    decks = [
        build_hog_cycle_deck(player_data),
        build_battle_ram_deck(player_data),
        build_goblin_barrel_bait_deck(player_data)
    ]

    for deck in decks:
        print(f"✓ {deck['deck_name']}")
        print(f"  Win Condition: {deck['win_condition']}")
        print(f"  Average Elixir: {deck['average_elixir']}")
        print(f"  Cards ({len(deck['deck_detail'])}/8):")
        for card in deck['deck_detail']:
            print(f"    - {card['name']} (Lvl {card['level']}, {card['elixir']} elixir)")

        # Save deck
        saved_path = save_deck(deck, output_dir, player_tag)
        print(f"  Saved to: {saved_path}\n")

    print("=== Deck Analysis Summary ===\n")

    for deck in decks:
        print(f"## {deck['deck_name']}")
        print(f"Average Elixir: {deck['average_elixir']}")
        print(f"Strategy: {deck['strategy']}")
        print(f"Playstyle: {deck['playstyle']}\n")

        print("Strengths:")
        for strength in deck['strengths']:
            print(f"  ✓ {strength}")

        print("\nWeaknesses:")
        for weakness in deck['weaknesses']:
            print(f"  ✗ {weakness}")

        print("\n" + "="*60 + "\n")


if __name__ == '__main__':
    main()
