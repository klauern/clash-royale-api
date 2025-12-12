#!/usr/bin/env python3
"""
CSV export functionality for Clash Royale data.
Provides structured CSV generation for player information, card collections, battle logs, and more.
"""

import re
from datetime import datetime
from pathlib import Path
from typing import Any, Optional, Union

import pandas as pd

from .event_deck_manager import EventDeckManager
from .models.event_deck import EventType


class CSVExporter:
    """Handles CSV export functionality for Clash Royale data."""

    def __init__(self, api_client, output_dir: Optional[Union[str, Path]] = None):
        """Initialize the CSV exporter with an API client and output directory."""
        self.api = api_client
        # Default to data/csv within the project, not outside it
        if output_dir:
            self.output_dir = Path(output_dir)
        else:
            # Use the API client's data_dir if available, otherwise use a local data/csv
            if hasattr(api_client, "data_dir"):
                self.output_dir = Path(api_client.data_dir) / "csv"
            else:
                self.output_dir = Path("data/csv")
        self.output_dir.mkdir(parents=True, exist_ok=True)

        # Create subdirectories
        (self.output_dir / "players").mkdir(exist_ok=True)
        (self.output_dir / "clans").mkdir(exist_ok=True)
        (self.output_dir / "analysis").mkdir(exist_ok=True)
        (self.output_dir / "reference").mkdir(exist_ok=True)
        (self.output_dir / "events").mkdir(exist_ok=True)

        # Cache for card reference data
        self._card_reference_cache = None

        # Initialize event deck manager
        self.event_manager = EventDeckManager(data_dir=str(self.output_dir.parent))

    def _sanitize_for_csv(self, value: Any) -> str:
        """Sanitize a value for CSV compatibility."""
        if value is None:
            return ""
        if isinstance(value, (list, dict)):
            value = str(value)
        # Remove newlines and tabs, replace with spaces
        value = str(value).replace("\n", " ").replace("\r", " ").replace("\t", " ")
        # Remove any control characters
        value = re.sub(r"[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]", "", value)
        return value

    def _flatten_dict(self, data: dict, prefix: str = "", separator: str = "_") -> dict:
        """Recursively flatten nested dictionaries."""
        items = []
        for key, value in data.items():
            new_key = f"{prefix}{separator}{key}" if prefix else key
            if isinstance(value, dict):
                items.extend(self._flatten_dict(value, new_key, separator).items())
            else:
                items.append((new_key, self._sanitize_for_csv(value)))
        return dict(items)

    def _expand_array_fields(self, data: dict, field_configs: dict) -> dict:
        """Expand arrays into multiple columns."""
        result = data.copy()

        for field, config in field_configs.items():
            if field in data and isinstance(data[field], list):
                items = data[field]
                max_items = config.get("max_items", len(items))
                field_prefix = config.get("prefix", field)

                for i in range(max_items):
                    if i < len(items):
                        item = items[i]
                        if isinstance(item, dict):
                            # For dict items, expand with nested keys
                            for sub_key, sub_value in item.items():
                                result[f"{field_prefix}_{i+1}_{sub_key}"] = (
                                    self._sanitize_for_csv(sub_value)
                                )
                        else:
                            result[f"{field_prefix}_{i+1}"] = self._sanitize_for_csv(
                                item
                            )
                    else:
                        # Fill missing items with empty values
                        result[f"{field_prefix}_{i+1}"] = ""

        return result

    async def _get_card_reference(self) -> dict[str, dict]:
        """Get cached card reference data."""
        if self._card_reference_cache is None:
            self._card_reference_cache = await self.api.get_all_cards()
            # Convert to a dict keyed by card name for easier lookup
            card_dict = {}
            for card in self._card_reference_cache.get("items", []):
                card_dict[card["name"]] = card
            self._card_reference_cache = card_dict

        return self._card_reference_cache

    def _generate_filename(
        self, data_type: str, tag: Optional[str] = None, timestamp: Optional[str] = None
    ) -> Path:
        """Generate a filename for a CSV export."""
        timestamp = timestamp or datetime.now().strftime("%Y-%m-%d")
        tag_clean = tag.replace("#", "").replace("/", "_") if tag else "all"

        if tag:
            filename = f"{data_type}_{tag_clean}_{timestamp}.csv"
        else:
            filename = f"{data_type}_{timestamp}.csv"

        return self.output_dir / filename

    async def export_player_info_csv(
        self, player_tag: str, filepath: Optional[Union[str, Path]] = None
    ) -> Path:
        """Export player information to CSV."""
        # Ensure player tag starts with #
        if not player_tag.startswith("#"):
            player_tag = f"#{player_tag}"

        # Get player data with card needs
        player_data = await self.api.get_player_info(
            player_tag, include_card_needs=True
        )

        # Prepare CSV data
        csv_data = {}

        # Basic player info
        csv_data["player_tag"] = player_data.get("tag", "")
        csv_data["name"] = self._sanitize_for_csv(player_data.get("name", ""))
        csv_data["exp_level"] = player_data.get("expLevel", 0)
        csv_data["trophies"] = player_data.get("trophies", 0)
        csv_data["best_trophies"] = player_data.get("bestTrophies", 0)
        csv_data["wins"] = player_data.get("wins", 0)
        csv_data["losses"] = player_data.get("losses", 0)
        csv_data["battle_count"] = player_data.get("battleCount", 0)
        csv_data["three_crown_wins"] = player_data.get("threeCrownWins", 0)

        # Calculate win rate
        wins = player_data.get("wins", 0)
        losses = player_data.get("losses", 0)
        csv_data["win_rate"] = (
            round(wins / max(wins + losses, 1) * 100, 2) if (wins + losses) > 0 else 0
        )

        # Clan information
        clan = player_data.get("clan", {})
        csv_data["clan_tag"] = clan.get("tag", "")
        csv_data["clan_name"] = self._sanitize_for_csv(clan.get("name", ""))
        csv_data["clan_role"] = clan.get("role", "")

        # Arena information
        arena = player_data.get("arena", {})
        csv_data["arena_name"] = self._sanitize_for_csv(arena.get("name", ""))
        csv_data["arena_id"] = arena.get("id", 0)

        # Current deck
        current_deck = player_data.get("currentDeck", [])
        deck_elixir = 0
        card_ref = await self._get_card_reference()

        for i, card in enumerate(current_deck[:8], 1):
            card_name = card.get("name", "")
            csv_data[f"current_deck_{i}"] = card_name
            # Calculate deck average elixir
            if card_name in card_ref:
                deck_elixir += card_ref[card_name].get("elixirCost", 0)

        csv_data["deck_avg_elixir"] = (
            round(deck_elixir / len(current_deck), 2) if current_deck else 0
        )
        csv_data["fetch_time"] = datetime.now().isoformat()

        # Write to CSV
        if filepath is None:
            filepath = (
                self.output_dir
                / "players"
                / f"player_info_{player_tag.replace('#', '')}_{datetime.now().strftime('%Y-%m-%d')}.csv"
            )
        else:
            filepath = Path(filepath)
            filepath.parent.mkdir(parents=True, exist_ok=True)

        df = pd.DataFrame([csv_data])
        df.to_csv(filepath, index=False)

        return filepath

    async def export_card_collection_csv(
        self, player_tag: str, filepath: Optional[Union[str, Path]] = None
    ) -> Path:
        """Export player's card collection to CSV."""
        if not player_tag.startswith("#"):
            player_tag = f"#{player_tag}"

        # Get player data and card reference
        player_data = await self.api.get_player_info(player_tag)
        card_ref = await self._get_card_reference()

        # Process card collection
        cards = player_data.get("cards", [])
        csv_data = []

        for card in cards:
            card_name = card.get("name", "")
            card_info = card_ref.get(card_name, {})

            row = {
                "player_tag": player_tag,
                "card_name": card_name,
                "level": card.get("level", 0),
                "count": card.get("count", 0),
                "max_level": card.get("maxLevel", 0),
                "rarity": card_info.get("rarity", ""),
                "elixir_cost": card_info.get("elixirCost", 0),
                "card_type": card_info.get("type", ""),
                "arena_required": card_info.get("arena", 0),
                "icon_url_medium": card.get("iconUrls", {}).get("medium", ""),
                "icon_url_large": card.get("iconUrls", {}).get("large", ""),
            }

            # Calculate cards needed for next level
            current_level = card.get("level", 0)
            max_level = card.get("maxLevel", 0)
            rarity = card_info.get("rarity", "Common")

            if current_level < max_level:
                # Use the API's calculation method
                cards_needed = self.api._calculate_cards_needed(current_level, rarity)
                row["cards_to_next_level"] = cards_needed
                row["is_max_level"] = False
                row["next_level"] = current_level + 1
            else:
                row["cards_to_next_level"] = 0
                row["is_max_level"] = True
                row["next_level"] = current_level

            csv_data.append(row)

        # Write to CSV
        if filepath is None:
            filepath = (
                self.output_dir
                / "players"
                / f"card_collection_{player_tag.replace('#', '')}_{datetime.now().strftime('%Y-%m-%d')}.csv"
            )
        else:
            filepath = Path(filepath)
            filepath.parent.mkdir(parents=True, exist_ok=True)

        df = pd.DataFrame(csv_data)
        df.to_csv(filepath, index=False)

        return filepath

    async def export_battle_log_csv(
        self,
        player_tag: str,
        limit: int = 100,
        filepath: Optional[Union[str, Path]] = None,
    ) -> Path:
        """Export battle log to CSV."""
        if not player_tag.startswith("#"):
            player_tag = f"#{player_tag}"

        # Get battle log
        battles = await self.api.get_player_battle_log(player_tag)

        # Limit battles if requested
        if limit and limit > 0:
            battles = battles[:limit]

        csv_data = []

        for battle in battles:
            row = {
                "player_tag": player_tag,
                "battle_time_utc": battle.get("battleTime", ""),
                "game_mode": self._sanitize_for_csv(
                    battle.get("gameMode", {}).get("name", "")
                ),
                "is_ladder_tournament": battle.get("isLadderTournament", False),
            }

            # Team information
            team = battle.get("team", [])
            if team:
                team_data = team[0]  # Get first team member
                row["team_crowns"] = team_data.get("crowns", 0)
                row["team_trophies_before"] = team_data.get("startingTrophies", 0)
                row["team_trophy_change"] = team_data.get("trophyChange", 0)

                # Team cards
                team_cards = team_data.get("cards", [])
                for i, card in enumerate(team_cards[:8], 1):
                    row[f"team_card_{i}"] = card.get("name", "")
                    row[f"team_card_{i}_level"] = card.get("level", 0)

            # Opponent information
            opponent = battle.get("opponent", [])
            if opponent:
                opponent_data = opponent[0]
                row["opponent_crowns"] = opponent_data.get("crowns", 0)
                row["opponent_name"] = self._sanitize_for_csv(
                    opponent_data.get("name", "")
                )
                row["opponent_tag"] = opponent_data.get("tag", "")
                row["opponent_trophies"] = opponent_data.get("startingTrophies", 0)

                # Opponent cards (if available)
                opponent_cards = opponent_data.get("cards", [])
                for i, card in enumerate(opponent_cards[:8], 1):
                    row[f"opponent_card_{i}"] = card.get("name", "")
                    row[f"opponent_card_{i}_level"] = card.get("level", 0)

            # Determine result
            if team and opponent:
                row["result"] = (
                    "win"
                    if team[0].get("crowns", 0) > opponent[0].get("crowns", 0)
                    else "loss"
                )

            csv_data.append(row)

        # Write to CSV
        if filepath is None:
            filepath = (
                self.output_dir
                / "players"
                / f"battle_log_{player_tag.replace('#', '')}_{datetime.now().strftime('%Y-%m-%d')}.csv"
            )
        else:
            filepath = Path(filepath)
            filepath.parent.mkdir(parents=True, exist_ok=True)

        df = pd.DataFrame(csv_data)
        df.to_csv(filepath, index=False)

        return filepath

    async def export_chest_cycle_csv(
        self, player_tag: str, filepath: Optional[Union[str, Path]] = None
    ) -> Path:
        """Export chest cycle to CSV."""
        if not player_tag.startswith("#"):
            player_tag = f"#{player_tag}"

        # Get chest cycle data
        chest_cycle = await self.api.get_player_chest_cycle(player_tag)

        csv_data = []

        for i, chest in enumerate(chest_cycle, 1):
            row = {
                "player_tag": player_tag,
                "chest_position": i,
                "chest_name": chest.get("name", ""),
                "time_to_open_hours": (
                    chest.get("timeToOpen", 0) / 3600 if chest.get("timeToOpen") else 0
                ),
            }

            # Classify chest rarity
            chest_name = chest.get("name", "").lower()
            if "magical" in chest_name:
                row["chest_rarity"] = "Magical"
            elif "super magical" in chest_name:
                row["chest_rarity"] = "Super Magical"
            elif "legendary" in chest_name:
                row["chest_rarity"] = "Legendary"
            elif "epic" in chest_name:
                row["chest_rarity"] = "Epic"
            elif "gold" in chest_name:
                row["chest_rarity"] = "Gold"
            else:
                row["chest_rarity"] = "Silver"

            csv_data.append(row)

        # Write to CSV
        if filepath is None:
            filepath = (
                self.output_dir
                / "players"
                / f"chest_cycle_{player_tag.replace('#', '')}_{datetime.now().strftime('%Y-%m-%d')}.csv"
            )
        else:
            filepath = Path(filepath)
            filepath.parent.mkdir(parents=True, exist_ok=True)

        df = pd.DataFrame(csv_data)
        df.to_csv(filepath, index=False)

        return filepath

    async def export_card_database_csv(
        self, filepath: Optional[Union[str, Path]] = None
    ) -> Path:
        """Export complete card database to CSV (reference file)."""
        # Get all cards
        all_cards = await self.api.get_all_cards()

        csv_data = []

        for card in all_cards.get("items", []):
            row = {
                "card_id": card.get("id", 0),
                "name": self._sanitize_for_csv(card.get("name", "")),
                "rarity": card.get("rarity", ""),
                "type": card.get("type", ""),
                "elixir_cost": card.get("elixirCost", 0),
                "arena_required": card.get("arena", 0),
                "max_level": card.get("maxLevel", 0),
                "description": self._sanitize_for_csv(card.get("description", "")),
                "icon_url_medium": card.get("iconUrls", {}).get("medium", ""),
                "icon_url_large": card.get("iconUrls", {}).get("large", ""),
            }

            csv_data.append(row)

        # Write to CSV
        if filepath is None:
            filepath = (
                self.output_dir
                / "reference"
                / f"card_database_{datetime.now().strftime('%Y-%m-%d')}.csv"
            )
        else:
            filepath = Path(filepath)
            filepath.parent.mkdir(parents=True, exist_ok=True)

        df = pd.DataFrame(csv_data)
        df.to_csv(filepath, index=False)

        return filepath

    async def export_all_data_csv(
        self, player_tag: str, directory: Optional[Union[str, Path]] = None
    ) -> list[Path]:
        """Export all available data types for a player to separate CSV files."""
        if not player_tag.startswith("#"):
            player_tag = f"#{player_tag}"

        # Use provided directory or default
        base_dir = Path(directory) if directory else self.output_dir / "players"
        base_dir.mkdir(parents=True, exist_ok=True)

        exported_files = []

        try:
            # Export player info
            filepath = await self.export_player_info_csv(
                player_tag,
                base_dir
                / f"player_info_{player_tag.replace('#', '')}_{datetime.now().strftime('%Y-%m-%d')}.csv",
            )
            exported_files.append(filepath)
        except Exception as e:
            print(f"Failed to export player info: {e}")

        try:
            # Export card collection
            filepath = await self.export_card_collection_csv(
                player_tag,
                base_dir
                / f"card_collection_{player_tag.replace('#', '')}_{datetime.now().strftime('%Y-%m-%d')}.csv",
            )
            exported_files.append(filepath)
        except Exception as e:
            print(f"Failed to export card collection: {e}")

        try:
            # Export battle log
            filepath = await self.export_battle_log_csv(
                player_tag,
                100,
                base_dir
                / f"battle_log_{player_tag.replace('#', '')}_{datetime.now().strftime('%Y-%m-%d')}.csv",
            )
            exported_files.append(filepath)
        except Exception as e:
            print(f"Failed to export battle log: {e}")

        try:
            # Export chest cycle
            filepath = await self.export_chest_cycle_csv(
                player_tag,
                base_dir
                / f"chest_cycle_{player_tag.replace('#', '')}_{datetime.now().strftime('%Y-%m-%d')}.csv",
            )
            exported_files.append(filepath)
        except Exception as e:
            print(f"Failed to export chest cycle: {e}")

        try:
            # Export event decks
            filepath = await self.export_event_decks_csv(
                player_tag,
                base_dir
                / f"event_decks_{player_tag.replace('#', '')}_{datetime.now().strftime('%Y-%m-%d')}.csv",
            )
            exported_files.append(filepath)
        except Exception as e:
            print(f"Failed to export event decks: {e}")

        try:
            # Export event analysis
            filepath = await self.export_event_analysis_csv(
                player_tag,
                base_dir
                / f"event_analysis_{player_tag.replace('#', '')}_{datetime.now().strftime('%Y-%m-%d')}.csv",
            )
            exported_files.append(filepath)
        except Exception as e:
            print(f"Failed to export event analysis: {e}")

        return exported_files

    async def export_event_decks_csv(
        self,
        player_tag: str,
        event_type: Optional[EventType] = None,
        filepath: Optional[Union[str, Path]] = None,
    ) -> Path:
        """
        Export event decks to CSV format.

        Args:
            player_tag: Player's tag
            event_type: Filter by event type (optional)
            filepath: Output file path (optional)

        Returns:
            Path to the generated CSV file
        """
        if not player_tag.startswith("#"):
            player_tag = f"#{player_tag}"

        # Get event decks from manager
        decks = await self.event_manager.get_event_decks(
            player_tag, event_type=event_type
        )

        # Prepare CSV data
        csv_data = []
        for deck in decks:
            row = {
                "player_tag": player_tag,
                "event_id": deck.event_id,
                "event_name": self._sanitize_for_csv(deck.event_name),
                "event_type": deck.event_type.value,
                "start_time": deck.start_time.isoformat() if deck.start_time else "",
                "end_time": deck.end_time.isoformat() if deck.end_time else "",
                "avg_elixir": round(deck.deck.avg_elixir, 2),
                "wins": deck.performance.wins,
                "losses": deck.performance.losses,
                "win_rate": round(deck.performance.win_rate * 100, 2),
                "crowns_earned": deck.performance.crowns_earned,
                "crowns_lost": deck.performance.crowns_lost,
                "max_wins": deck.performance.max_wins,
                "current_streak": deck.performance.current_streak,
                "best_streak": deck.performance.best_streak,
                "progress": deck.performance.progress.value,
                "notes": self._sanitize_for_csv(deck.notes) if deck.notes else "",
            }

            # Add cards (8 cards max)
            for i, card in enumerate(deck.deck.cards, 1):
                row[f"card_{i}"] = self._sanitize_for_csv(card.name)
                row[f"card_{i}_level"] = card.level
                row[f"card_{i}_rarity"] = card.rarity

            # Fill missing card slots
            for i in range(len(deck.deck.cards) + 1, 9):
                row[f"card_{i}"] = ""
                row[f"card_{i}_level"] = ""
                row[f"card_{i}_rarity"] = ""

            csv_data.append(row)

        # Write to CSV
        if filepath is None:
            event_type_str = f"_{event_type.value}" if event_type else ""
            filepath = (
                self.output_dir
                / "events"
                / f"event_decks_{player_tag.replace('#', '')}{event_type_str}_{datetime.now().strftime('%Y-%m-%d')}.csv"
            )
        else:
            filepath = Path(filepath)
            filepath.parent.mkdir(parents=True, exist_ok=True)

        df = pd.DataFrame(csv_data)
        df.to_csv(filepath, index=False)

        return filepath

    async def export_event_analysis_csv(
        self,
        player_tag: str,
        event_type: Optional[EventType] = None,
        filepath: Optional[Union[str, Path]] = None,
    ) -> Path:
        """
        Export event deck analysis to CSV.

        Args:
            player_tag: Player's tag
            event_type: Filter by event type (optional)
            filepath: Output file path (optional)

        Returns:
            Path to the generated CSV file
        """
        if not player_tag.startswith("#"):
            player_tag = f"#{player_tag}"

        # Get analysis from event manager
        analysis = await self.event_manager.analyze_event_decks(
            player_tag, event_type=event_type
        )

        # Prepare CSV data
        csv_data = []

        # Summary statistics
        summary = analysis.get("summary", {})
        csv_data.append(
            {
                "metric": "Total Event Decks",
                "value": summary.get("total_decks", 0),
                "category": "summary",
            }
        )
        csv_data.append(
            {
                "metric": "Total Battles",
                "value": summary.get("total_battles", 0),
                "category": "summary",
            }
        )
        csv_data.append(
            {
                "metric": "Overall Win Rate (%)",
                "value": round(summary.get("overall_win_rate", 0) * 100, 2),
                "category": "summary",
            }
        )
        csv_data.append(
            {
                "metric": "Average Crowns per Battle",
                "value": round(summary.get("avg_crowns_per_battle", 0), 2),
                "category": "summary",
            }
        )
        csv_data.append(
            {
                "metric": "Average Deck Elixir",
                "value": round(summary.get("avg_deck_elixir", 0), 2),
                "category": "summary",
            }
        )

        # Card usage
        card_analysis = analysis.get("card_analysis", {})
        most_used = card_analysis.get("most_used_cards", [])
        for i, (card_name, count) in enumerate(most_used[:10], 1):
            csv_data.append(
                {
                    "metric": f"Most Used Card #{i}",
                    "value": f"{card_name} ({count} decks)",
                    "category": "card_usage",
                }
            )

        # Best performing cards
        best_cards = card_analysis.get("highest_win_rate_cards", [])
        for i, (card_name, win_rate) in enumerate(best_cards[:10], 1):
            csv_data.append(
                {
                    "metric": f"Highest Win Rate Card #{i}",
                    "value": f"{card_name} ({round(win_rate * 100, 2)}%)",
                    "category": "card_performance",
                }
            )

        # Elixir analysis
        elixir_analysis = analysis.get("elixir_analysis", {})
        for elixir_range, stats in elixir_analysis.items():
            csv_data.append(
                {
                    "metric": f'{elixir_range.replace("_", " ").title()} Decks',
                    "value": f"{stats.get('decks', 0)} decks ({round(stats.get('avg_win_rate', 0) * 100, 2)}% WR)",
                    "category": "elixir_analysis",
                }
            )

        # Top performing decks
        top_decks = analysis.get("top_performing_decks", [])
        for i, deck in enumerate(top_decks[:5], 1):
            deck_str = " | ".join(deck["deck"][:4])  # Show first 4 cards
            csv_data.append(
                {
                    "metric": f"Top Deck #{i}",
                    "value": f"{deck['event_name']}: {deck['record']} ({deck_str}...)",
                    "category": "top_decks",
                }
            )

        # Write to CSV
        if filepath is None:
            event_type_str = f"_{event_type.value}" if event_type else ""
            filepath = (
                self.output_dir
                / "events"
                / f"event_analysis_{player_tag.replace('#', '')}{event_type_str}_{datetime.now().strftime('%Y-%m-%d')}.csv"
            )
        else:
            filepath = Path(filepath)
            filepath.parent.mkdir(parents=True, exist_ok=True)

        df = pd.DataFrame(csv_data)
        df.to_csv(filepath, index=False)

        return filepath
