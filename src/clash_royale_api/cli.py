#!/usr/bin/env python3
"""
Command line interface for Clash Royale API data collector.
"""

import argparse
import asyncio
import os
import sys
from datetime import datetime
from pathlib import Path

# Add the src directory to the path so we can import our module
sys.path.insert(0, str(Path(__file__).parent.parent))

from clash_royale_api import ClashRoyaleAPI
from clash_royale_api.models.event_deck import EventType


def get_arg_with_env_fallback(args, arg_name, env_var, default=None):
    """Get argument value with environment variable fallback.
    Priority: CLI argument > Environment variable > Default value
    """
    cli_value = getattr(args, arg_name, None)
    if cli_value is not None:
        return cli_value
    return os.getenv(env_var, default)


def load_env_file(config_path=None):
    """Load environment variables from .env file."""
    from dotenv import load_dotenv

    if config_path:
        load_dotenv(config_path)
    else:
        # Load from project root .env by default
        env_path = Path(__file__).parent.parent.parent / ".env"
        if env_path.exists():
            load_dotenv(env_path)


async def main():
    """Main CLI entry point."""
    parser = argparse.ArgumentParser(
        description="Clash Royale API Data Collector",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s --player #PLAYERTAG              # Analyze a specific player
  %(prog)s --player #PLAYERTAG --save       # Analyze and save player data
  %(prog)s --cards                         # Get all available cards
  %(prog)s --player #PLAYERTAG --chests    # Get player's upcoming chests

  Event Deck Operations:
  %(prog)s --player #PLAYERTAG --scan-event-decks      # Scan battle logs for event decks
  %(prog)s --player #PLAYERTAG --export-event-decks   # Export event decks to CSV
  %(prog)s --player #PLAYERTAG --export-deck-builder  # Export in deck builder format
  %(prog)s --player #PLAYERTAG --analyze-events      # Analyze event performance
  %(prog)s --player #PLAYERTAG --sync-recent-events   # Sync recent event activity
        """,
    )

    parser.add_argument("--player", "-p", help="Player tag to analyze (include #)")

    parser.add_argument(
        "--cards",
        "-c",
        action="store_true",
        help="Fetch and display all available cards",
    )

    parser.add_argument(
        "--chests",
        action="store_true",
        help="Include chest cycle information (requires --player)",
    )

    parser.add_argument(
        "--build-ladder-deck",
        action="store_true",
        help="Build a recommended 1v1 ladder deck from your card analysis",
    )

    parser.add_argument(
        "--analysis-file",
        help="Use a saved analysis JSON for deck building (defaults to most recent)",
    )

    parser.add_argument(
        "--save-deck",
        action="store_true",
        help="Save recommended deck output to data/decks",
    )

    parser.add_argument(
        "--save", "-s", action="store_true", help="Save results to files"
    )

    parser.add_argument("--config", help="Path to .env configuration file")

    parser.add_argument(
        "--format",
        "-f",
        choices=["json", "table"],
        default="table",
        help="Output format (default: table)",
    )

    # CSV export arguments
    parser.add_argument(
        "--export-csv", "-csv", action="store_true", help="Export data to CSV format"
    )

    parser.add_argument(
        "--csv-dir",
        default="data/csv",
        help="Output directory for CSV files (default: data/csv)",
    )

    parser.add_argument(
        "--csv-types",
        help="Comma-separated list of CSV types to export: player,cards,battles,chests,all (default: all)",
    )

    parser.add_argument(
        "--battle-limit",
        type=int,
        default=100,
        help="Number of recent battles to export (default: 100)",
    )

    # Event deck arguments
    parser.add_argument(
        "--scan-event-decks",
        action="store_true",
        help="Scan battle logs and automatically import event decks (requires --player)",
    )

    parser.add_argument(
        "--export-event-decks",
        action="store_true",
        help="Export event decks to file (requires --player)",
    )

    parser.add_argument(
        "--export-deck-builder",
        action="store_true",
        help="Export event decks in deck builder format (requires --player)",
    )

    parser.add_argument(
        "--analyze-events",
        action="store_true",
        help="Analyze event deck performance (requires --player)",
    )

    parser.add_argument(
        "--sync-recent-events",
        action="store_true",
        help="Sync recent event activity from battle logs (requires --player)",
    )

    parser.add_argument(
        "--event-type",
        choices=[t.value for t in EventType],
        help="Filter by specific event type",
    )

    parser.add_argument(
        "--days-back",
        type=int,
        default=7,
        help="Number of days to scan for event decks (default: 7)",
    )

    parser.add_argument("--event-output", help="Output file for event deck export")

    args = parser.parse_args()

    # Load environment variables
    config_path = getattr(args, 'config', None) or os.getenv("CONFIG_PATH")
    load_env_file(config_path)

    # Apply environment variable fallbacks
    args.player = get_arg_with_env_fallback(args, 'player', 'DEFAULT_PLAYER_TAG')
    args.format = get_arg_with_env_fallback(args, 'format', 'OUTPUT_FORMAT', 'table')
    args.csv_dir = get_arg_with_env_fallback(args, 'csv_dir', 'CSV_DIR', 'data/csv')
    args.csv_types = get_arg_with_env_fallback(args, 'csv_types', 'CSV_TYPES', 'all')
    args.battle_limit = int(get_arg_with_env_fallback(args, 'battle_limit', 'BATTLE_LIMIT', '100'))
    args.days_back = int(get_arg_with_env_fallback(args, 'days_back', 'DAYS_BACK', '7'))
    args.event_output = get_arg_with_env_fallback(args, 'event_output', 'EVENT_OUTPUT')

    api = None
    try:
        # Initialize API client
        api = ClashRoyaleAPI(args.config)

        try:
            # Handle CSV export
            if args.export_csv:
                exported_files = []

                # Card database export (doesn't require player)
                if args.cards:
                    try:
                        print("\n=== Exporting Card Database ===")
                        filepath = await api.export_card_database_csv()
                        exported_files.append(filepath)
                        print(f"✓ Exported card database to: {filepath}")
                    except Exception as e:
                        print(f"✗ Failed to export card database: {e}")

                # Player-specific CSV export
                if args.player:
                    # Determine which types to export
                    csv_types = args.csv_types or "all"
                    if csv_types == "all":
                        export_types = ["player", "cards", "battles", "chests"]
                    else:
                        export_types = [t.strip() for t in csv_types.split(",")]

                    print(f"\n=== Exporting CSV Data for {args.player} ===")

                    if "player" in export_types:
                        try:
                            filepath = await api.export_player_info_csv(args.player)
                            exported_files.append(filepath)
                            print(f"✓ Exported player info to: {filepath}")
                        except Exception as e:
                            print(f"✗ Failed to export player info: {e}")

                    if "cards" in export_types:
                        try:
                            filepath = await api.export_card_collection_csv(args.player)
                            exported_files.append(filepath)
                            print(f"✓ Exported card collection to: {filepath}")
                        except Exception as e:
                            print(f"✗ Failed to export card collection: {e}")

                    if "battles" in export_types:
                        try:
                            filepath = await api.export_battle_log_csv(
                                args.player, args.battle_limit
                            )
                            exported_files.append(filepath)
                            print(f"✓ Exported battle log to: {filepath}")
                        except Exception as e:
                            print(f"✗ Failed to export battle log: {e}")

                    if "chests" in export_types:
                        try:
                            filepath = await api.export_chest_cycle_csv(args.player)
                            exported_files.append(filepath)
                            print(f"✓ Exported chest cycle to: {filepath}")
                        except Exception as e:
                            print(f"✗ Failed to export chest cycle: {e}")

                if exported_files:
                    print(f"\nSuccessfully exported {len(exported_files)} CSV file(s)")
                    return 0
                else:
                    print("\nError: --export-csv requires either --player or --cards")
                    return 1

            # Handle event deck operations
            if any(
                [
                    args.scan_event_decks,
                    args.export_event_decks,
                    args.export_deck_builder,
                    args.analyze_events,
                    args.sync_recent_events,
                ]
            ):
                if not args.player:
                    print("\nError: Event deck operations require --player or DEFAULT_PLAYER_TAG in .env")
                    return 1

                # Parse event type if provided
                event_type = None
                if args.event_type:
                    event_type = EventType(args.event_type)

                # Scan event decks from battle logs
                if args.scan_event_decks or args.sync_recent_events:
                    try:
                        print("\n=== Scanning Event Decks from Battle Logs ===")
                        days = (
                            args.days_back
                            if args.sync_recent_events
                            else args.days_back
                        )
                        imported_count = await api.scan_and_import_event_decks(
                            args.player, days_back=days
                        )
                        print(f"✓ Successfully imported {imported_count} event decks")
                    except Exception as e:
                        print(f"✗ Failed to scan event decks: {e}")

                # Export event decks
                if args.export_event_decks or args.export_deck_builder:
                    try:
                        print("\n=== Exporting Event Decks ===")
                        output_file = args.event_output
                        if not output_file:
                            timestamp = datetime.now().strftime("%Y%m%d")
                            if args.export_deck_builder:
                                output_file = f"event_decks_{args.player.replace('#', '')}_{timestamp}.decklink"
                            else:
                                output_file = f"event_decks_{args.player.replace('#', '')}_{timestamp}.csv"

                        format_type = (
                            "deck_builder" if args.export_deck_builder else "csv"
                        )
                        filepath = await api.export_event_decks(
                            args.player,
                            output_file,
                            format=format_type,
                            event_type=event_type,
                        )
                        print(f"✓ Exported event decks to: {filepath}")
                    except Exception as e:
                        print(f"✗ Failed to export event decks: {e}")

                # Analyze event decks
                if args.analyze_events:
                    try:
                        print("\n=== Analyzing Event Deck Performance ===")
                        analysis = await api.analyze_event_decks(
                            args.player, event_type=event_type
                        )

                        # Print summary
                        summary = analysis.get("summary", {})
                        print("\n📊 Summary:")
                        print(f"  Total Event Decks: {summary.get('total_decks', 0)}")
                        print(
                            f"  Overall Win Rate: {summary.get('overall_win_rate', 0):.1%}"
                        )
                        print(f"  Total Battles: {summary.get('total_battles', 0)}")
                        print(
                            f"  Average Crowns/Battle: {summary.get('avg_crowns_per_battle', 0):.1f}"
                        )
                        print(
                            f"  Average Deck Elixir: {summary.get('avg_deck_elixir', 0):.1f}"
                        )

                        # Print most used cards
                        card_analysis = analysis.get("card_analysis", {})
                        most_used = card_analysis.get("most_used_cards", [])[:5]
                        if most_used:
                            print("\n🃏 Most Used Cards:")
                            for card, count in most_used:
                                print(f"  {card}: {count} decks")

                        # Print best performing cards
                        best_cards = card_analysis.get("highest_win_rate_cards", [])[:5]
                        if best_cards:
                            print("\n⭐ Highest Win Rate Cards:")
                            for card, win_rate in best_cards:
                                print(f"  {card}: {win_rate:.1%}")

                        # Print top decks
                        top_decks = analysis.get("top_performing_decks", [])[:3]
                        if top_decks:
                            print("\n🏆 Top Performing Decks:")
                            for i, deck in enumerate(top_decks, 1):
                                print(
                                    f"  {i}. {deck['event_name']}: {deck['record']} ({deck['win_rate']})"
                                )

                        # Save analysis to file if requested
                        if args.event_output:
                            import json

                            analysis_file = args.event_output.replace(
                                ".csv", "_analysis.json"
                            ).replace(".decklink", "_analysis.json")
                            with open(analysis_file, "w") as f:
                                json.dump(analysis, f, indent=2)
                            print(f"\n💾 Analysis saved to: {analysis_file}")

                    except Exception as e:
                        print(f"✗ Failed to analyze event decks: {e}")

                return 0

            if args.cards:
                print("\n=== Clash Royale Cards ===")
                cards = await api.get_all_cards()

                if args.format == "table":
                    print(f"Total cards: {len(cards.get('items', []))}")
                    for card in cards.get("items", [])[:10]:  # Show first 10
                        print(
                            f"  {card.get('name', 'Unknown')} - {card.get('rarity', 'Unknown')}"
                        )
                    if len(cards.get("items", [])) > 10:
                        print(f"  ... and {len(cards.get('items', [])) - 10} more")

                if args.save:
                    await api.save_data(cards, "all_cards.json", "static")

            if args.player:
                print(f"\n=== Player Analysis: {args.player} ===")

                # Get player info with card needs included
                player_info = await api.get_player_info(
                    args.player, include_card_needs=True
                )

                if args.format == "table":
                    print(f"Name: {player_info.get('name', 'Unknown')}")
                    print(f"Trophies: {player_info.get('trophies', 0)}")
                    print(f"Level: {player_info.get('expLevel', 0)}")
                    print(
                        f"Arena: {player_info.get('arena', {}).get('name', 'Unknown')}"
                    )

                    cards = player_info.get("cards", [])
                    print(f"\nCard Collection: {len(cards)} cards")

                    # Show cards needing upgrades with detailed info
                    cards_needing_upgrade = []
                    print("\nCards needing upgrade:")
                    print("-" * 80)
                    print(
                        f"{'Card Name':<25} {'Level':<8} {'Count':<8} {'Rarity':<12} {'Needed for Next':<15}"
                    )
                    print("-" * 80)

                    for card in cards:
                        current_level = card.get("level", 0)
                        max_level = card.get("maxLevel", 13)
                        card_count = card.get("count", 0)

                        if current_level < max_level:
                            cards_needing_upgrade.append(card)
                            rarity = card.get("rarity", "Common")
                            cards_needed = api._calculate_cards_needed(
                                current_level, rarity
                            )
                            print(
                                f"{card.get('name', 'Unknown'):<25} {current_level:<8} {card_count:<8} {rarity:<12} {cards_needed:<15}"
                            )

                    if not cards_needing_upgrade:
                        print("All cards are at max level!")
                    else:
                        print(
                            f"\nTotal cards needing upgrade: {len(cards_needing_upgrade)}"
                        )

                        # Show summary by rarity
                        rarity_summary = {}
                        for card in cards_needing_upgrade:
                            rarity = card.get("rarity", "Unknown")
                            if rarity not in rarity_summary:
                                rarity_summary[rarity] = {"count": 0, "total_needed": 0}
                            rarity_summary[rarity]["count"] += 1
                            rarity_summary[rarity][
                                "total_needed"
                            ] += api._calculate_cards_needed(
                                card.get("level", 0), rarity
                            )

                        print("\nUpgrade needs by rarity:")
                        for rarity, data in sorted(rarity_summary.items()):
                            print(
                                f"  {rarity}: {data['count']} cards, {data['total_needed']} total cards needed"
                            )

                if args.build_ladder_deck:
                    try:
                        print("\n=== Recommended 1v1 Deck ===")
                        deck_data = await api.build_ladder_deck(
                            args.player,
                            use_latest_saved=True,
                            analysis_path=args.analysis_file,
                            save=args.save_deck,
                        )
                        print(
                            f"Average Elixir: {deck_data.get('average_elixir', 0):.2f}"
                        )
                        for idx, card in enumerate(deck_data.get("deck_detail", []), 1):
                            print(
                                f"{idx}. {card['name']} (lvl {card['level']}/{card['max_level']}, "
                                f"{card['elixir']} elixir) - {card.get('role', 'flex')}"
                            )
                        notes = deck_data.get("notes", [])
                        if notes:
                            print("\nNotes:")
                            for note in notes:
                                print(f"- {note}")
                        if deck_data.get("saved_to"):
                            print(f"\nSaved to: {deck_data['saved_to']}")
                    except Exception as e:
                        print(f"✗ Failed to build ladder deck: {e}")

                if args.chests:
                    print("\n=== Upcoming Chests ===")
                    chests = await api.get_player_upcoming_chests(args.player)

                    if args.format == "table":
                        upcoming = chests.get("items", [])[:5]
                        for chest in upcoming:
                            print(
                                f"  {chest.get('name', 'Unknown')} - {chest.get('index', 0)} positions"
                            )

                if args.save:
                    # Save enhanced player profile with card analysis included
                    await api.get_complete_player_profile(
                        args.player, save=True
                    )
                    print(
                        f"✓ Player profile saved to: data/players/player_profile_{args.player.lstrip('#')}.json"
                    )

                    # Also save the analysis separately for reference
                    analysis = await api.analyze_card_collection(args.player)
                    await api.save_data(
                        analysis, f"analysis_{args.player.lstrip('#')}.json", "analysis"
                    )
                    print(
                        f"✓ Card analysis saved to: data/analysis/analysis_{args.player.lstrip('#')}.json"
                    )

            if not args.cards and not args.player:
                parser.print_help()
                print("\nError: Please specify --player or --cards")
                return 1

            return 0

        finally:
            # Clean up API client if it was initialized
            if api:
                await api.close()

    except Exception as e:
        print(f"Error: {e}")
        return 1


def cli_entry():
    """Synchronous entry point for the CLI script."""
    exit_code = asyncio.run(main())
    sys.exit(exit_code)


if __name__ == "__main__":
    cli_entry()
