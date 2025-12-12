#!/usr/bin/env python3
"""
Generate detailed mulligan guide (opening plays) for each deck.
"""
import json
from pathlib import Path


def generate_mulligan_guide():
    """Create mulligan guide for all three decks."""

    guides = []

    # Goblin Barrel Bait Mulligan Guide
    barrel_mulligan = {
        'deck_name': 'Goblin Barrel Bait',
        'general_principles': [
            'Never open with Goblin Barrel (waste of surprise)',
            'Scout opponent\'s spell first with Goblin Gang or Skeleton Dragons',
            'Play cheapest defensive card in back (Fire Spirit or Archers)',
            'Save Fireball for their pumps or buildings'
        ],
        'matchups': [
            {
                'opponent_type': 'Beatdown (Giant, Golem, Royal Giant)',
                'opening_play': 'Goblin Gang at bridge',
                'reason': 'Apply early pressure, force them to defend instead of building pushes',
                'backup': 'If they ignore it, follow with Skeleton Dragons opposite lane',
                'key_cards': ['Cannon (for kiting)', 'Skeleton Dragons (air defense)']
            },
            {
                'opponent_type': 'Hog Cycle / Fast Cycle',
                'opening_play': 'Cannon in center (4 tiles from river)',
                'reason': 'Defensive building ready for first Hog push',
                'backup': 'Fire Spirit in back to start cycling',
                'key_cards': ['Cannon', 'Goblin Gang (swarm Hog)']
            },
            {
                'opponent_type': 'Bridge Spam (Battle Ram, Bandit)',
                'opening_play': 'Archers in back',
                'reason': 'Ranged pressure, can defend both lanes',
                'backup': 'Goblin Gang for quick swarm defense',
                'key_cards': ['Goblin Gang', 'Skeleton Dragons']
            },
            {
                'opponent_type': 'Control/Spell Bait',
                'opening_play': 'Fire Spirit in back',
                'reason': 'Cheap cycle, don\'t commit until you see their cards',
                'backup': 'Mirror their play - if they play tank, you play support',
                'key_cards': ['Save Fireball for Princess/Swarms']
            },
            {
                'opponent_type': 'X-Bow / Mortar (Siege)',
                'opening_play': 'Goblin Gang at bridge immediately',
                'reason': 'Prevent them from locking siege building on your tower',
                'backup': 'Skeleton Dragons same lane to destroy siege building',
                'key_cards': ['Goblin Barrel opposite lane when they place siege']
            },
            {
                'opponent_type': 'Unknown (Scout First)',
                'opening_play': 'Archers in back OR Fire Spirit in back',
                'reason': 'Safe plays that give you elixir advantage and card rotation info',
                'backup': 'React to their first play defensively',
                'key_cards': ['Hold Zap and Fireball until you identify threats']
            }
        ],
        'never_open_with': [
            'Goblin Barrel - telegraphs your win condition',
            'Fireball - waste of 4 elixir with no value',
            'Zap - gives no information and wastes cycle',
            'Cannon - only play reactively to their push'
        ]
    }

    # Hog Cycle Mulligan Guide
    hog_mulligan = {
        'deck_name': 'Hog Cycle (Fast Aggression)',
        'general_principles': [
            'Open with Skeletons or Fire Spirit in back to start fast cycle',
            'Hog + Fire Spirit (5 elixir) is your main offensive combo',
            'Always have Cannon available for defense',
            'Cycle to Musketeer for air defense quickly'
        ],
        'matchups': [
            {
                'opponent_type': 'Beatdown (Giant, Golem)',
                'opening_play': 'Hog Rider at bridge',
                'reason': 'Force them to defend early, prevent big elixir investments',
                'backup': 'Cycle cheap cards (Skeletons, Fire Spirit) to get back to Hog',
                'key_cards': ['Cannon (kite tank)', 'Musketeer (DPS on tank)']
            },
            {
                'opponent_type': 'Hog Mirror / Fast Cycle',
                'opening_play': 'Cannon pre-emptively in center',
                'reason': 'Ready for their first Hog, avoid tower damage',
                'backup': 'Knight in back to prepare defense',
                'key_cards': ['Cannon every rotation', 'Skeletons to surround Hog']
            },
            {
                'opponent_type': 'Bridge Spam',
                'opening_play': 'Skeletons in back',
                'reason': 'Cheap cycle, ready to defend sudden rush',
                'backup': 'Knight or Cannon for immediate defense',
                'key_cards': ['Knight (tank spam)', 'Fire Spirit (splash)']
            },
            {
                'opponent_type': 'Siege (X-Bow, Mortar)',
                'opening_play': 'Hog Rider opposite lane immediately',
                'reason': 'Force them to defend, can\'t build siege safely',
                'backup': 'Knight to tank their siege building',
                'key_cards': ['Fireball their siege building', 'Hog opposite lane pressure']
            },
            {
                'opponent_type': 'Control Decks',
                'opening_play': 'Fire Spirit in back',
                'reason': 'Cheap cycle, see their archetype',
                'backup': 'Knight in back for stable defense',
                'key_cards': ['Save Fireball for their key defensive cards']
            },
            {
                'opponent_type': 'Unknown',
                'opening_play': 'Skeletons in back OR Knight in back',
                'reason': 'Safe, cheap cycle or stable tank',
                'backup': 'React with Cannon if they push',
                'key_cards': ['Keep Hog + Fire Spirit ready for counter-push']
            }
        ],
        'never_open_with': [
            'Fireball - no value, wastes elixir',
            'Zap - meaningless cycle',
            'Musketeer - too expensive for opening, vulnerable to spell',
            'Cannon - only defensive, reactive play'
        ]
    }

    # Battle Ram Cycle Mulligan Guide
    ram_mulligan = {
        'deck_name': 'Battle Ram Cycle (Balanced Aggression)',
        'general_principles': [
            'Battle Ram works best with support troops behind it',
            'Goblin Cage is both defense AND offense (Brawler spawns)',
            'Build small pushes rather than solo Ram',
            'Use Bats for quick air defense and DPS'
        ],
        'matchups': [
            {
                'opponent_type': 'Beatdown',
                'opening_play': 'Goblin Cage in center proactively',
                'reason': 'Ready to defend, spawns Brawler for counter-push',
                'backup': 'Battle Ram opposite lane when they commit',
                'key_cards': ['Goblin Cage', 'Musketeer for DPS']
            },
            {
                'opponent_type': 'Hog / Fast Cycle',
                'opening_play': 'Goblin Gang at bridge',
                'reason': 'Early pressure with your strongest card (level 11)',
                'backup': 'Goblin Cage ready for their Hog',
                'key_cards': ['Goblin Cage', 'Fire Spirit']
            },
            {
                'opponent_type': 'Bridge Spam',
                'opening_play': 'Musketeer in back',
                'reason': 'Ranged support ready to defend both lanes',
                'backup': 'Goblin Gang for swarm defense',
                'key_cards': ['Goblin Gang', 'Zap for resets']
            },
            {
                'opponent_type': 'Siege',
                'opening_play': 'Battle Ram at bridge immediately',
                'reason': 'Prevent siege lock, force defensive play',
                'backup': 'Goblin Gang to finish siege building',
                'key_cards': ['Battle Ram pressure', 'Fireball on siege']
            },
            {
                'opponent_type': 'Control',
                'opening_play': 'Fire Spirit in back',
                'reason': 'Cheap cycle, prepare for counter-push',
                'backup': 'Goblin Cage when safe',
                'key_cards': ['Save Fireball for their defensive buildings']
            },
            {
                'opponent_type': 'Unknown',
                'opening_play': 'Bats in back OR Fire Spirit in back',
                'reason': 'Safe, cheap cycle cards',
                'backup': 'Goblin Cage if they push, Battle Ram if they play defensive',
                'key_cards': ['Keep Zap for swarms defending your Ram']
            }
        ],
        'never_open_with': [
            'Battle Ram alone - easily countered without support',
            'Fireball - no value',
            'Zap - wastes cycle',
            'Musketeer - vulnerable to spell value'
        ]
    }

    guides = [barrel_mulligan, hog_mulligan, ram_mulligan]

    # Print guide
    print("="*80)
    print("🎯 MULLIGAN GUIDE - OPENING PLAYS FOR ALL DECKS")
    print("="*80)
    print()

    for guide in guides:
        print(f"\n{'='*80}")
        print(f"🃏 {guide['deck_name'].upper()}")
        print(f"{'='*80}\n")

        print("📋 General Principles:")
        for principle in guide['general_principles']:
            print(f"   • {principle}")
        print()

        print("🎮 Matchup-Specific Openings:\n")

        for i, matchup in enumerate(guide['matchups'], 1):
            print(f"{i}. VS {matchup['opponent_type']}")
            print(f"   ▶ Opening Play: {matchup['opening_play']}")
            print(f"   ▶ Why: {matchup['reason']}")
            print(f"   ▶ Backup: {matchup['backup']}")
            print(f"   ▶ Key Cards to Cycle To: {', '.join(matchup['key_cards'])}")
            print()

        print("❌ Never Open With:")
        for never in guide['never_open_with']:
            print(f"   ✗ {never}")
        print()

    # Save to JSON
    output = {
        'mulligan_guides': guides,
        'created_at': '2025-12-10'
    }

    output_file = Path('data/analysis/mulligan_guides.json')
    with open(output_file, 'w') as f:
        json.dump(output, f, indent=2)

    print(f"\nMulligan guides saved to: {output_file}")

    return guides


if __name__ == '__main__':
    generate_mulligan_guide()
