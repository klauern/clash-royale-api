#!/usr/bin/env python3
"""
Analyze player's playstyle based on overall stats and current deck.
Recommend best deck match from the three custom decks.
"""
import json
from pathlib import Path


def analyze_playstyle(player_data: dict) -> dict:
    """Analyze player statistics to determine playstyle."""

    player_info = player_data.get('player_info', {})

    wins = player_info.get('wins', 0)
    losses = player_info.get('losses', 0)
    total_battles = player_info.get('battle_count', 0)
    three_crown_wins = player_info.get('three_crown_wins', 0)

    # Calculate key metrics
    win_rate = (wins / total_battles * 100) if total_battles > 0 else 0
    three_crown_rate = (three_crown_wins / wins * 100) if wins > 0 else 0

    # Analyze current deck
    current_deck_data = player_data.get('current_deck', {})
    current_deck = current_deck_data.get('cards', []) if isinstance(current_deck_data, dict) else []

    if not current_deck and 'cards' in player_data:
        # Try to find currentDeck from raw player data
        with open('data/players/#R8QGUQRCV.json') as f:
            raw_data = json.load(f)
            current_deck = raw_data.get('currentDeck', [])

    deck_avg_elixir = 0
    win_condition = None

    if current_deck:
        total_elixir = sum(card.get('elixirCost', 0) for card in current_deck)
        deck_avg_elixir = total_elixir / len(current_deck) if current_deck else 0

        # Find win condition
        win_conditions = {'Royal Giant', 'Hog Rider', 'Giant', 'Battle Ram', 'Goblin Barrel'}
        for card in current_deck:
            if card.get('name') in win_conditions:
                win_condition = card.get('name')
                break

    # Determine playstyle based on metrics
    playstyle_traits = []

    # 1. Aggression level (based on 3-crown rate)
    if three_crown_rate > 75:
        aggression = "VERY AGGRESSIVE"
        playstyle_traits.append("Goes for tower damage aggressively")
        playstyle_traits.append("Prefers offensive pressure over defensive play")
    elif three_crown_rate > 60:
        aggression = "Aggressive"
        playstyle_traits.append("Balanced offense with strong finishing")
    else:
        aggression = "Defensive/Reactive"
        playstyle_traits.append("Prefers defensive counterplay")

    # 2. Consistency (based on win rate)
    if win_rate > 55:
        consistency = "High"
        playstyle_traits.append("Consistent execution and matchup knowledge")
    elif win_rate > 48:
        consistency = "Balanced"
        playstyle_traits.append("Adapting and learning matchups")
    else:
        consistency = "Learning"
        playstyle_traits.append("Building skills and adapting strategy")

    # 3. Current deck style
    if deck_avg_elixir > 0:
        if deck_avg_elixir < 3.0:
            deck_style = "Ultra-fast cycle"
            playstyle_traits.append("Prefers constant pressure with fast cycle")
        elif deck_avg_elixir < 3.5:
            deck_style = "Fast cycle"
            playstyle_traits.append("Comfortable with aggressive tempo")
        elif deck_avg_elixir < 4.0:
            deck_style = "Balanced"
        else:
            deck_style = "Beatdown/Heavy"
    else:
        deck_style = "Unknown"

    return {
        'wins': wins,
        'losses': losses,
        'total_battles': total_battles,
        'win_rate': round(win_rate, 1),
        'three_crown_wins': three_crown_wins,
        'three_crown_rate': round(three_crown_rate, 1),
        'aggression_level': aggression,
        'consistency': consistency,
        'current_deck_avg_elixir': round(deck_avg_elixir, 2),
        'current_win_condition': win_condition,
        'deck_style': deck_style,
        'playstyle_traits': playstyle_traits
    }


def recommend_deck(playstyle: dict) -> dict:
    """Recommend best deck based on playstyle analysis."""

    # Load the three custom decks
    deck_files = [
        'data/decks/20251210_233235_hog_cycle_fast_aggression_R8QGUQRCV.json',
        'data/decks/20251210_233235_battle_ram_cycle_balanced_aggression_R8QGUQRCV.json',
        'data/decks/20251210_233235_goblin_barrel_bait_high-level_cards_R8QGUQRCV.json'
    ]

    decks = []
    for deck_file in deck_files:
        try:
            with open(deck_file) as f:
                decks.append(json.load(f))
        except FileNotFoundError:
            continue

    # Score each deck based on playstyle
    deck_scores = []

    for deck in decks:
        score = 0
        reasons = []

        # Factor 1: Three-crown aggression match
        if playstyle['three_crown_rate'] > 75:
            # Very aggressive player
            if deck['average_elixir'] < 2.9:
                score += 40
                reasons.append(f"Ultra-low elixir ({deck['average_elixir']}) matches your very aggressive playstyle")
            elif deck['average_elixir'] < 3.1:
                score += 30
                reasons.append(f"Low elixir ({deck['average_elixir']}) suits aggressive play")

        # Factor 2: Card level quality
        avg_card_level = sum(c['level'] / c['maxLevel'] for c in deck['deck_detail']) / len(deck['deck_detail'])
        score += int(avg_card_level * 30)
        reasons.append(f"Card level ratio: {avg_card_level:.1%}")

        # Factor 3: Win condition level
        win_con_card = next((c for c in deck['deck_detail'] if c['name'] == deck['win_condition']), None)
        if win_con_card:
            win_con_ratio = win_con_card['level'] / win_con_card['maxLevel']
            if win_con_ratio > 0.6:
                score += 20
                reasons.append(f"{deck['win_condition']} level is competitive ({win_con_card['level']})")
            elif win_con_ratio > 0.5:
                score += 10
                reasons.append(f"{deck['win_condition']} level is decent ({win_con_card['level']})")
            else:
                score += 0
                reasons.append(f"⚠️ {deck['win_condition']} is underleveled ({win_con_card['level']})")

        # Factor 4: Current deck similarity (familiarity bonus)
        if playstyle['current_win_condition'] == deck['win_condition']:
            score += 10
            reasons.append(f"Already using {deck['win_condition']} - familiar playstyle")

        deck_scores.append({
            'deck': deck,
            'score': score,
            'reasons': reasons
        })

    # Sort by score
    deck_scores.sort(key=lambda x: x['score'], reverse=True)

    return {
        'recommended': deck_scores[0] if deck_scores else None,
        'all_scores': deck_scores
    }


def main():
    """Run playstyle analysis and recommendation."""

    # Load player profile
    profile_file = Path('data/players/20251210_233440_player_profile_R8QGUQRCV.json')
    with open(profile_file) as f:
        player_data = json.load(f)

    # Analyze playstyle
    playstyle = analyze_playstyle(player_data)

    print("="*70)
    print("🎮 PLAYSTYLE ANALYSIS FOR ZYLOGAN")
    print("="*70)
    print()

    print(f"📊 Overall Statistics:")
    print(f"   Total Battles: {playstyle['total_battles']}")
    print(f"   Record: {playstyle['wins']}W - {playstyle['losses']}L")
    print(f"   Win Rate: {playstyle['win_rate']}%")
    print(f"   Three-Crown Wins: {playstyle['three_crown_wins']} ({playstyle['three_crown_rate']}% of wins)")
    print()

    print(f"⚡ Playstyle Profile:")
    print(f"   Aggression Level: {playstyle['aggression_level']}")
    print(f"   Consistency: {playstyle['consistency']}")
    print(f"   Current Deck Style: {playstyle['deck_style']}")
    if playstyle['current_win_condition']:
        print(f"   Current Win Condition: {playstyle['current_win_condition']}")
        print(f"   Current Average Elixir: {playstyle['current_deck_avg_elixir']}")
    print()

    print(f"🎯 Key Traits:")
    for trait in playstyle['playstyle_traits']:
        print(f"   • {trait}")
    print()

    print("="*70)
    print()

    # Get deck recommendation
    recommendation = recommend_deck(playstyle)

    if recommendation['recommended']:
        print("🏆 RECOMMENDED DECK")
        print("="*70)

        top_deck = recommendation['recommended']['deck']
        print(f"Deck: {top_deck['deck_name']}")
        print(f"Win Condition: {top_deck['win_condition']}")
        print(f"Average Elixir: {top_deck['average_elixir']}")
        print(f"Match Score: {recommendation['recommended']['score']}/100")
        print()

        print("Why this deck:")
        for reason in recommendation['recommended']['reasons']:
            print(f"  ✓ {reason}")
        print()

        print("Cards:")
        for card in top_deck['deck_detail']:
            level_pct = card['level'] / card['maxLevel'] * 100
            level_indicator = "✓" if level_pct > 60 else "○"
            print(f"  {level_indicator} {card['name']:<20} Lvl {card['level']}/{card['maxLevel']} ({card['elixir']} elixir)")
        print()

        print(f"Strategy: {top_deck['strategy']}")
        print()

        print("="*70)
        print()

        # Show other options
        print("📋 OTHER DECK OPTIONS")
        print("="*70)
        for i, ranked_deck in enumerate(recommendation['all_scores'][1:], 2):
            deck = ranked_deck['deck']
            print(f"\n#{i}: {deck['deck_name']}")
            print(f"   Score: {ranked_deck['score']}/100")
            print(f"   Average Elixir: {deck['average_elixir']}")
            print(f"   Top reason: {ranked_deck['reasons'][0]}")

    print()
    print("="*70)
    print()

    # Save analysis
    output = {
        'playstyle_analysis': playstyle,
        'deck_recommendation': recommendation,
        'analysis_time': player_data.get('fetch_time')
    }

    output_file = Path('data/analysis/playstyle_analysis_R8QGUQRCV.json')
    with open(output_file, 'w') as f:
        json.dump(output, f, indent=2)

    print(f"Analysis saved to: {output_file}")


if __name__ == '__main__':
    main()
