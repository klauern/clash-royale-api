"""Manager for saving, retrieving, and analyzing event decks."""

import csv
import json
import logging
from collections import defaultdict
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Optional

from .battle_log_parser import BattleLogParser
from .models.event_deck import EventDeck, EventDeckCollection, EventType

logger = logging.getLogger(__name__)


class EventDeckManager:
    """Manages event deck storage, retrieval, and analysis."""

    def __init__(self, data_dir: str = "./data"):
        """Initialize the manager with a data directory."""
        self.data_dir = Path(data_dir)
        self.event_decks_dir = self.data_dir / "event_decks"
        self.parser = BattleLogParser()

        # Create directories if they don't exist
        self.event_decks_dir.mkdir(parents=True, exist_ok=True)

    def get_player_event_dir(self, player_tag: str) -> Path:
        """Get the directory for a player's event decks."""
        # Remove # from tag for directory name
        clean_tag = player_tag.lstrip("#")
        player_dir = self.event_decks_dir / clean_tag
        player_dir.mkdir(exist_ok=True)

        # Create subdirectories
        (player_dir / "challenges").mkdir(exist_ok=True)
        (player_dir / "tournaments").mkdir(exist_ok=True)
        (player_dir / "special_events").mkdir(exist_ok=True)
        (player_dir / "aggregated").mkdir(exist_ok=True)

        return player_dir

    async def save_event_deck(self, event_deck: EventDeck) -> None:
        """
        Save an event deck to the file system.

        Args:
            event_deck: The EventDeck object to save
        """
        player_dir = self.get_player_event_dir(event_deck.player_tag)

        # Determine subdirectory based on event type
        if event_deck.event_type == EventType.TOURNAMENT:
            subdir = player_dir / "tournaments"
        elif event_deck.event_type == EventType.SPECIAL_EVENT:
            subdir = player_dir / "special_events"
        else:
            subdir = player_dir / "challenges"

        # Generate filename
        timestamp = event_deck.start_time.strftime("%Y-%m-%d")
        event_name = event_deck.event_name.lower().replace(" ", "_").replace("/", "_")
        filename = f"{timestamp}_{event_name}.json"
        filepath = subdir / filename

        # Save to file
        try:
            with open(filepath, "w") as f:
                json.dump(event_deck.dict(), f, indent=2, default=str)
            logger.info(f"Saved event deck to {filepath}")
        except Exception as e:
            logger.error(f"Failed to save event deck: {e}")
            raise

        # Also update the collection file
        await self._update_collection_file(event_deck)

    async def _update_collection_file(self, event_deck: EventDeck) -> None:
        """Update the player's event deck collection file."""
        player_dir = self.get_player_event_dir(event_deck.player_tag)
        collection_file = player_dir / "collection.json"

        # Load existing collection
        collection = None
        if collection_file.exists():
            try:
                with open(collection_file) as f:
                    data = json.load(f)
                    collection = EventDeckCollection(**data)
            except Exception as e:
                logger.warning(f"Could not load collection file: {e}")

        # Create new collection if needed
        if collection is None:
            collection = EventDeckCollection(player_tag=event_deck.player_tag)

        # Add the deck
        collection.add_deck(event_deck)

        # Save collection
        with open(collection_file, "w") as f:
            json.dump(collection.dict(), f, indent=2, default=str)

    async def get_event_decks(
        self,
        player_tag: str,
        event_type: Optional[EventType] = None,
        days_back: Optional[int] = None,
        limit: Optional[int] = None,
    ) -> list[EventDeck]:
        """
        Retrieve event decks for a player.

        Args:
            player_tag: Player's tag
            event_type: Filter by event type (optional)
            days_back: Only get decks from last N days (optional)
            limit: Maximum number of decks to return (optional)

        Returns:
            List of EventDeck objects
        """
        player_dir = self.get_player_event_dir(player_tag)
        decks = []

        # Determine which subdirectories to search
        subdirs = []
        if event_type:
            if event_type == EventType.TOURNAMENT:
                subdirs = [player_dir / "tournaments"]
            elif event_type == EventType.SPECIAL_EVENT:
                subdirs = [player_dir / "special_events"]
            else:
                subdirs = [player_dir / "challenges"]
        else:
            subdirs = [
                player_dir / "challenges",
                player_dir / "tournaments",
                player_dir / "special_events",
            ]

        # Load decks from subdirectories
        for subdir in subdirs:
            if not subdir.exists():
                continue

            for file_path in subdir.glob("*.json"):
                if file_path.name == "collection.json":
                    continue  # Skip collection file

                try:
                    with open(file_path) as f:
                        data = json.load(f)
                        deck = EventDeck(**data)

                        # Apply filters
                        if days_back:
                            cutoff = datetime.now() - timedelta(days=days_back)
                            if deck.start_time < cutoff:
                                continue

                        decks.append(deck)

                except Exception as e:
                    logger.warning(f"Could not load deck from {file_path}: {e}")
                    continue

        # Sort by start time (newest first)
        decks.sort(key=lambda d: d.start_time, reverse=True)

        # Apply limit
        if limit:
            decks = decks[:limit]

        return decks

    async def import_from_battle_logs(
        self, battle_logs: list[dict], player_tag: str
    ) -> list[EventDeck]:
        """
        Import event decks from battle logs.

        Args:
            battle_logs: Battle log data from API
            player_tag: Player's tag

        Returns:
            List of newly imported EventDeck objects
        """
        # Parse battle logs
        event_decks = await self.parser.parse_battle_logs(battle_logs, player_tag)

        # Save each event deck
        imported = []
        for deck in event_decks:
            try:
                await self.save_event_deck(deck)
                imported.append(deck)
                logger.info(
                    f"Imported event deck: {deck.event_name} on {deck.start_time}"
                )
            except Exception as e:
                logger.error(f"Failed to import event deck: {e}")

        return imported

    async def export_to_csv(
        self, player_tag: str, filepath: str, event_type: Optional[EventType] = None
    ) -> None:
        """
        Export event decks to CSV format.

        Args:
            player_tag: Player's tag
            filepath: Output CSV file path
            event_type: Filter by event type (optional)
        """
        decks = await self.get_event_decks(player_tag, event_type=event_type)

        with open(filepath, "w", newline="", encoding="utf-8") as csvfile:
            fieldnames = [
                "event_id",
                "event_name",
                "event_type",
                "start_time",
                "end_time",
                "card_1",
                "card_2",
                "card_3",
                "card_4",
                "card_5",
                "card_6",
                "card_7",
                "card_8",
                "avg_elixir",
                "wins",
                "losses",
                "win_rate",
                "crowns_earned",
                "max_wins",
                "best_streak",
                "progress",
                "notes",
            ]

            writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
            writer.writeheader()

            for deck in decks:
                row = {
                    "event_id": deck.event_id,
                    "event_name": deck.event_name,
                    "event_type": deck.event_type.value,
                    "start_time": (
                        deck.start_time.isoformat() if deck.start_time else ""
                    ),
                    "end_time": deck.end_time.isoformat() if deck.end_time else "",
                    "avg_elixir": deck.deck.avg_elixir,
                    "wins": deck.performance.wins,
                    "losses": deck.performance.losses,
                    "win_rate": f"{deck.performance.win_rate:.2%}",
                    "crowns_earned": deck.performance.crowns_earned,
                    "max_wins": deck.performance.max_wins,
                    "best_streak": deck.performance.best_streak,
                    "progress": deck.performance.progress.value,
                    "notes": deck.notes or "",
                }

                # Add cards
                for i, card in enumerate(deck.deck.cards, 1):
                    row[f"card_{i}"] = card.name

                writer.writerow(row)

        logger.info(f"Exported {len(decks)} event decks to {filepath}")

    async def export_to_json(
        self, player_tag: str, filepath: str, event_type: Optional[EventType] = None
    ) -> None:
        """
        Export event decks to JSON format.

        Args:
            player_tag: Player's tag
            filepath: Output JSON file path
            event_type: Filter by event type (optional)
        """
        decks = await self.get_event_decks(player_tag, event_type=event_type)

        export_data = {
            "player_tag": player_tag,
            "export_time": datetime.now().isoformat(),
            "filter": {"event_type": event_type.value if event_type else None},
            "total_decks": len(decks),
            "decks": [deck.dict() for deck in decks],
        }

        with open(filepath, "w", encoding="utf-8") as f:
            json.dump(export_data, f, indent=2, default=str)

        logger.info(f"Exported {len(decks)} event decks to {filepath}")

    async def export_deck_builder_format(
        self, player_tag: str, filepath: str, event_type: Optional[EventType] = None
    ) -> None:
        """
        Export decks in format compatible with deck builders.

        Args:
            player_tag: Player's tag
            filepath: Output file path
            event_type: Filter by event type (optional)
        """
        decks = await self.get_event_decks(player_tag, event_type=event_type)

        # Create deck builder format
        deck_lines = []
        for deck in decks:
            # Comment line with event info
            comment = f"# {deck.event_name} - {deck.start_time.strftime('%Y-%m-%d')} - W{deck.performance.wins}L{deck.performance.losses}"
            deck_lines.append(comment)

            # Deck line with cards
            card_names = [card.name for card in deck.deck.cards]
            deck_line = " ".join(card_names)
            deck_lines.append(deck_line)

            # Empty line for separation
            deck_lines.append("")

        # Write to file
        with open(filepath, "w", encoding="utf-8") as f:
            f.write("\n".join(deck_lines))

        logger.info(f"Exported {len(decks)} decks in deck builder format to {filepath}")

    async def analyze_event_decks(
        self, player_tag: str, event_type: Optional[EventType] = None
    ) -> dict[str, Any]:
        """
        Analyze event decks and return insights.

        Args:
            player_tag: Player's tag
            event_type: Filter by event type (optional)

        Returns:
            Dictionary with analysis results
        """
        decks = await self.get_event_decks(player_tag, event_type=event_type)

        if not decks:
            return {"error": "No event decks found"}

        # Card usage statistics
        card_usage = {}
        card_wins = {}
        card_losses = {}

        # Deck statistics
        total_battles = sum(d.performance.wins + d.performance.losses for d in decks)
        total_wins = sum(d.performance.wins for d in decks)
        total_losses = sum(d.performance.losses for d in decks)
        total_crowns = sum(d.performance.crowns_earned for d in decks)

        # Event type breakdown
        event_type_stats = defaultdict(lambda: {"count": 0, "wins": 0, "losses": 0})

        # Best performing decks
        performing_decks = sorted(
            [d for d in decks if d.performance.wins + d.performance.losses >= 3],
            key=lambda d: d.performance.win_rate,
            reverse=True,
        )[:5]

        # Analyze each deck
        for deck in decks:
            # Update event type stats
            event_type_stats[deck.event_type.value]["count"] += 1
            event_type_stats[deck.event_type.value]["wins"] += deck.performance.wins
            event_type_stats[deck.event_type.value]["losses"] += deck.performance.losses

            # Analyze card usage
            for card in deck.deck.cards:
                card_usage[card.name] = card_usage.get(card.name, 0) + 1
                card_wins[card.name] = (
                    card_wins.get(card.name, 0) + deck.performance.wins
                )
                card_losses[card.name] = (
                    card_losses.get(card.name, 0) + deck.performance.losses
                )

        # Calculate card win rates
        card_win_rates = {}
        for card in card_usage:
            total_card_wins = card_wins.get(card, 0)
            total_card_losses = card_losses.get(card, 0)
            if total_card_wins + total_card_losses > 0:
                card_win_rates[card] = total_card_wins / (
                    total_card_wins + total_card_losses
                )

        # Sort cards by usage and win rate
        most_used_cards = sorted(card_usage.items(), key=lambda x: x[1], reverse=True)[
            :10
        ]
        best_cards = sorted(card_win_rates.items(), key=lambda x: x[1], reverse=True)[
            :10
        ]

        # Average elixir analysis
        avg_elixirs = [d.deck.avg_elixir for d in decks]
        avg_elixir_mean = sum(avg_elixirs) / len(avg_elixirs)
        avg_elixir_by_win_rate = {
            "low_elixir": {
                "range": "0-3.0",
                "decks": len([d for d in decks if d.deck.avg_elixir <= 3.0]),
                "avg_win_rate": sum(
                    d.performance.win_rate for d in decks if d.deck.avg_elixir <= 3.0
                )
                / max(1, len([d for d in decks if d.deck.avg_elixir <= 3.0])),
            },
            "mid_elixir": {
                "range": "3.1-4.0",
                "decks": len([d for d in decks if 3.0 < d.deck.avg_elixir <= 4.0]),
                "avg_win_rate": sum(
                    d.performance.win_rate
                    for d in decks
                    if 3.0 < d.deck.avg_elixir <= 4.0
                )
                / max(1, len([d for d in decks if 3.0 < d.deck.avg_elixir <= 4.0])),
            },
            "high_elixir": {
                "range": "4.1+",
                "decks": len([d for d in decks if d.deck.avg_elixir > 4.0]),
                "avg_win_rate": sum(
                    d.performance.win_rate for d in decks if d.deck.avg_elixir > 4.0
                )
                / max(1, len([d for d in decks if d.deck.avg_elixir > 4.0])),
            },
        }

        return {
            "summary": {
                "total_decks": len(decks),
                "total_battles": total_battles,
                "overall_win_rate": total_wins / max(1, total_wins + total_losses),
                "avg_crowns_per_battle": total_crowns / max(1, total_battles),
                "avg_deck_elixir": avg_elixir_mean,
            },
            "card_analysis": {
                "most_used_cards": most_used_cards,
                "highest_win_rate_cards": best_cards,
                "total_unique_cards": len(card_usage),
            },
            "elixir_analysis": avg_elixir_by_win_rate,
            "event_type_breakdown": dict(event_type_stats),
            "top_performing_decks": [
                {
                    "event_name": d.event_name,
                    "event_type": d.event_type.value,
                    "win_rate": f"{d.performance.win_rate:.2%}",
                    "record": f"{d.performance.wins}W-{d.performance.losses}L",
                    "deck": [card.name for card in d.deck.cards],
                    "avg_elixir": d.deck.avg_elixir,
                }
                for d in performing_decks
            ],
        }
