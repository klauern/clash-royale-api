#!/usr/bin/env python3
"""
Enhanced Clash Royale API client using the official clashroyale package.
Provides comprehensive data collection and analysis capabilities.
"""

import asyncio
import json
import os
import time
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Optional, Union

# Official Clash Royale API wrapper
from clashroyale import OfficialAPI
from dotenv import load_dotenv

# Import CSV exporter and event deck manager
from .csv_exporter import CSVExporter
from .deck_builder import DeckBuilder
from .event_deck_manager import EventDeckManager
from .models.event_deck import EventType


class ClashRoyaleAPI:
    """Enhanced client for the Clash Royale API with data analysis capabilities."""

    def __init__(self, config_path: Optional[str] = None):
        """Initialize the API client with configuration."""
        # Load environment variables
        if config_path:
            env_path = Path(config_path)
        else:
            # Default to project root .env
            env_path = Path(__file__).parent.parent.parent / ".env"

        # Only load if the file exists
        if env_path.exists():
            load_dotenv(env_path)

        self.api_token = os.getenv("CLASH_ROYALE_API_TOKEN")
        if not self.api_token:
            raise ValueError(
                "API token not found. Please set CLASH_ROYALE_API_TOKEN in .env file."
            )

        self.data_dir = Path(os.getenv("DATA_DIR", "data"))
        self.data_dir.mkdir(exist_ok=True)

        # Initialize the official API client with async enabled
        self.client = OfficialAPI(token=self.api_token, is_async=True)

        # Configuration
        self.request_delay = float(os.getenv("REQUEST_DELAY", 1))
        self.max_retries = int(os.getenv("MAX_RETRIES", 3))

        # Rate limiting tracking
        self.last_request_time = 0

        # Initialize CSV exporter
        self.csv_exporter = CSVExporter(self)

        # Initialize event deck manager
        self.event_manager = EventDeckManager(data_dir=str(self.data_dir))
        self.deck_builder = DeckBuilder(self.data_dir)

    async def _rate_limit(self):
        """Implement rate limiting between requests."""
        current_time = time.time()
        time_since_last = current_time - self.last_request_time

        if time_since_last < self.request_delay:
            await asyncio.sleep(self.request_delay - time_since_last)

        self.last_request_time = time.time()

    async def get_all_cards(self) -> dict[str, Any]:
        """Fetch all available cards from the API."""
        print("Fetching all cards...")
        await self._rate_limit()
        result = await self.client.get_all_cards()
        # Convert to dict for consistent format
        if hasattr(result, "to_dict"):
            return result.to_dict()
        elif isinstance(result, list):
            # If it's already a list of cards, wrap it
            return {"items": result}
        else:
            return result

    async def get_player_info(
        self, player_tag: str, include_card_needs: bool = False
    ) -> dict[str, Any]:
        """Fetch comprehensive player information."""
        # Clean player tag
        clean_tag = player_tag.lstrip("#")
        print(f"Fetching player info for {player_tag}...")
        await self._rate_limit()
        result = await self.client.get_player(clean_tag)
        player_data = result.to_dict() if hasattr(result, "to_dict") else result

        # Add card needs information if requested
        if include_card_needs:
            cards = player_data.get("cards", [])
            # Get card reference data
            all_cards_data = await self.get_all_cards()
            card_reference = {
                card.get("name"): card for card in all_cards_data.get("items", [])
            }

            # Enhance each card with cards needed for next level
            for card in cards:
                card_name = card.get("name", "")
                rarity = card_reference.get(card_name, {}).get("rarity", "Common")
                current_level = card.get("level", 0)
                max_level = card.get("maxLevel", 13)

                card["cards_to_next_level"] = (
                    self._calculate_cards_needed(current_level, rarity)
                    if current_level < max_level
                    else 0
                )
                card["next_level"] = (
                    current_level + 1 if current_level < max_level else current_level
                )

        return player_data

    async def get_player_upcoming_chests(self, player_tag: str) -> dict[str, Any]:
        """Fetch upcoming chests for a player."""
        clean_tag = player_tag.lstrip("#")
        print(f"Fetching upcoming chests for {player_tag}...")
        await self._rate_limit()
        result = await self.client.get_player_chests(clean_tag)

        # Handle different response formats
        if hasattr(result, "to_dict"):
            data = result.to_dict()
            return data if isinstance(data, dict) else {"items": data}
        elif isinstance(result, list):
            # If result is already a list, wrap it in a dict with 'items' key
            return {"items": result}
        else:
            # Assume it's already a dict
            return result

    async def get_player_chest_cycle(self, player_tag: str) -> list[dict[str, Any]]:
        """Get the full chest cycle for a player."""
        clean_tag = player_tag.lstrip("#")
        print(f"Fetching chest cycle for {player_tag}...")
        await self._rate_limit()
        result = await self.client.get_player_chests(clean_tag)
        # The chest cycle is included in the chests response
        if hasattr(result, "to_dict"):
            data = result.to_dict()
            return data.get("items", [])
        return []

    async def get_player_card_collection(self, player_tag: str) -> list[dict[str, Any]]:
        """Extract detailed card collection from player data."""
        player_data = await self.get_player_info(player_tag)
        return player_data.get("cards", [])

    async def get_player_battle_log(self, player_tag: str) -> list[dict[str, Any]]:
        """Get recent battles for a player."""
        clean_tag = player_tag.lstrip("#")
        print(f"Fetching battle log for {player_tag}...")
        await self._rate_limit()
        result = await self.client.get_player_battles(clean_tag)
        if hasattr(result, "to_dict"):
            data = result.to_dict()
            return data if isinstance(data, list) else []
        return []

    async def save_data(self, data: Any, filename: str, subdir: Optional[str] = None):
        """Save data to JSON file with timestamp."""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

        if subdir:
            save_dir = self.data_dir / subdir
            save_dir.mkdir(exist_ok=True)
        else:
            save_dir = self.data_dir

        filename = f"{timestamp}_{filename}"
        filepath = save_dir / filename

        with open(filepath, "w") as f:
            json.dump(data, f, indent=2, default=str)

        print(f"Data saved to: {filepath}")
        return filepath

    async def get_complete_player_profile(
        self, player_tag: str, save: bool = True, include_card_analysis: bool = True
    ) -> dict[str, Any]:
        """Get comprehensive player data including all related information."""
        profile = {
            "player_tag": player_tag,
            "fetch_time": datetime.now().isoformat(),
            "player_info": await self.get_player_info(player_tag),
            "upcoming_chests": await self.get_player_upcoming_chests(player_tag),
            "chest_cycle": await self.get_player_chest_cycle(player_tag),
            "battle_log": await self.get_player_battle_log(player_tag),
        }

        # Add card analysis if requested
        if include_card_analysis:
            profile["card_analysis"] = await self.analyze_card_collection(player_tag)

        if save:
            await self.save_data(
                profile, f"player_profile_{player_tag.lstrip('#')}.json", "players"
            )

        return profile

    async def analyze_card_collection(self, player_tag: str) -> dict[str, Any]:
        """Analyze player's card collection with detailed statistics."""
        player_data = await self.get_player_info(player_tag)
        cards = player_data.get("cards", [])

        analysis = {
            "player_tag": player_tag,
            "analysis_time": datetime.now().isoformat(),
            "total_cards": len(cards),
            "card_levels": {},
            "rarity_breakdown": {},
            "max_level_cards": [],
            "cards_needing_upgrade": [],
            "collection_completion": {},
        }

        # Get all cards data for reference
        all_cards_data = await self.get_all_cards()

        # Create card reference dictionary
        card_reference = {
            card.get("name"): card for card in all_cards_data.get("items", [])
        }

        for card in cards:
            card_name = card.get("name", "Unknown")
            card_level = card.get("level", 0)
            card_count = card.get("count", 0)

            # Get card details from reference
            card_details = card_reference.get(card_name, {})
            rarity = card_details.get("rarity", "Unknown")
            max_level = card_details.get("maxLevel", 13)

            analysis["card_levels"][card_name] = {
                "level": card_level,
                "count": card_count,
                "rarity": rarity,
                "max_level": max_level,
                "elixir": card_details.get("elixirCost", 0),
                "cards_to_next_level": self._calculate_cards_needed(card_level, rarity),
            }

            # Rarity breakdown
            if rarity not in analysis["rarity_breakdown"]:
                analysis["rarity_breakdown"][rarity] = {
                    "count": 0,
                    "max_level": 0,
                    "total_possible": 0,
                }

            analysis["rarity_breakdown"][rarity]["count"] += 1

            # Count total cards of each rarity
            if rarity not in analysis["collection_completion"]:
                analysis["collection_completion"][rarity] = {"owned": 0, "total": 0}

            analysis["collection_completion"][rarity]["owned"] += 1

            if card_level >= max_level:
                analysis["max_level_cards"].append(card_name)
                analysis["rarity_breakdown"][rarity]["max_level"] += 1
            elif card_level < max_level - 1:
                priority = self._get_upgrade_priority(rarity, card_level, max_level)
                analysis["cards_needing_upgrade"].append(
                    {
                        "name": card_name,
                        "current_level": card_level,
                        "max_level": max_level,
                        "priority": priority,
                        "cards_needed": self._calculate_cards_needed(
                            card_level, rarity
                        ),
                    }
                )

        # Calculate total possible cards per rarity
        for card in all_cards_data.get("items", []):
            rarity = card.get("rarity", "Unknown")
            if rarity in analysis["collection_completion"]:
                analysis["collection_completion"][rarity]["total"] += 1

        # Sort cards needing upgrade by priority
        analysis["cards_needing_upgrade"].sort(key=lambda x: x["priority"])

        # Calculate collection completion percentages
        for rarity in analysis["collection_completion"]:
            data = analysis["collection_completion"][rarity]
            data["completion_percentage"] = (
                (data["owned"] / data["total"] * 100) if data["total"] > 0 else 0
            )

        return analysis

    async def build_ladder_deck(
        self,
        player_tag: str,
        use_latest_saved: bool = True,
        analysis_path: Optional[Union[str, Path]] = None,
        save: bool = False,
    ) -> dict[str, Any]:
        """
        Build a recommended 1v1 ladder deck from saved or live analysis data.

        If a cached analysis file is provided (or found), it will be used; otherwise
        the method will fetch live card data and analyze it before building the deck.
        """
        analysis = None

        if analysis_path:
            analysis = self.deck_builder.load_analysis(Path(analysis_path))
        elif use_latest_saved:
            analysis = self.deck_builder.load_latest_analysis(
                player_tag, self.data_dir / "analysis"
            )

        if analysis is None:
            analysis = await self.analyze_card_collection(player_tag)

        deck_data = self.deck_builder.build_deck_from_analysis(analysis)

        if save:
            saved_path = self.deck_builder.save_deck(
                deck_data, self.data_dir / "decks", player_tag
            )
            deck_data["saved_to"] = str(saved_path)

        return deck_data

    def _calculate_cards_needed(self, current_level: int, rarity: str) -> int:
        """Calculate cards needed for next level based on rarity."""
        # Base cards needed for each level progression in Clash Royale
        # These values represent the number of cards required to go from one level to the next
        # Level 1 -> 2: first value, Level 2 -> 3: second value, etc.
        cards_needed_per_level = {
            "Common": [2, 4, 10, 20, 50, 100, 200, 400, 800, 1000, 2000, 3000],
            "Rare": [5, 10, 25, 50, 125, 250, 500, 1000, 2000, 2500, 5000, 7500],
            "Epic": [10, 20, 50, 100, 250, 500, 1000, 2000, 4000, 5000, 10000, 15000],
            "Legendary": [
                20,
                40,
                100,
                200,
                500,
                1000,
                2000,
                4000,
                8000,
                10000,
                20000,
                30000,
            ],
            "Champion": [
                15,
                30,
                75,
                150,
                375,
                750,
                1500,
                3000,
                6000,
                7500,
                15000,
                22500,
            ],
        }

        # Get the progression array for this rarity
        progression = cards_needed_per_level.get(
            rarity, cards_needed_per_level["Common"]
        )

        # If already at max level, no cards needed
        if current_level >= len(progression):
            return 0

        # Return the cards needed for the next level
        return progression[current_level] if current_level < len(progression) else 0

    def _get_upgrade_priority(
        self, rarity: str, current_level: int, max_level: int
    ) -> float:
        """Calculate upgrade priority (lower number = higher priority)."""
        rarity_priority = {
            "Common": 1,
            "Rare": 2,
            "Epic": 3,
            "Legendary": 4,
            "Champion": 5,
        }

        level_progress = (max_level - current_level) / max_level
        return rarity_priority.get(rarity, 6) + level_progress

    async def analyze_battle_history(
        self, player_tag: str, battles: Optional[list] = None
    ) -> dict[str, Any]:
        """Analyze recent battle history for insights."""
        if battles is None:
            battles = await self.get_player_battle_log(player_tag)

        analysis = {
            "player_tag": player_tag,
            "analysis_time": datetime.now().isoformat(),
            "total_battles": len(battles),
            "wins": 0,
            "losses": 0,
            "deck_usage": {},
            "battle_types": {},
            "trophy_change": 0,
            "recent_performance": [],
        }

        for battle in battles:
            team = battle.get("team", [])
            if not team:
                continue

            # Check if player won
            player_crowns = team[0].get("crowns", 0)
            opponent_crowns = battle.get("opponent", [{}])[0].get("crowns", 0)
            won = player_crowns > opponent_crowns

            if won:
                analysis["wins"] += 1
            else:
                analysis["losses"] += 1

            # Track trophy changes
            trophy_change = battle.get("trophyChange", 0)
            analysis["trophy_change"] += trophy_change

            # Track deck usage (simplified)
            if won and len(team) > 0:
                deck_key = f"{len(team[0].get('cards', []))}_cards"
                analysis["deck_usage"][deck_key] = (
                    analysis["deck_usage"].get(deck_key, 0) + 1
                )

            # Track battle types
            battle_type = battle.get("type", "Unknown")
            analysis["battle_types"][battle_type] = (
                analysis["battle_types"].get(battle_type, 0) + 1
            )

            # Recent performance (last 10 battles)
            if len(analysis["recent_performance"]) < 10:
                analysis["recent_performance"].append(
                    {
                        "won": won,
                        "trophy_change": trophy_change,
                        "crowns": player_crowns,
                    }
                )

        # Calculate win rate
        analysis["win_rate"] = (
            (analysis["wins"] / analysis["total_battles"] * 100)
            if analysis["total_battles"] > 0
            else 0
        )

        return analysis

    async def close(self):
        """Close the API client session."""
        if hasattr(self.client, "close"):
            await self.client.close()

    async def __aenter__(self):
        """Async context manager entry."""
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        await self.close()

    # CSV Export Methods
    async def export_player_info_csv(
        self, player_tag: str, filepath: Optional[Union[str, Path]] = None
    ) -> Path:
        """Export player information to CSV."""
        return await self.csv_exporter.export_player_info_csv(player_tag, filepath)

    async def export_card_collection_csv(
        self, player_tag: str, filepath: Optional[Union[str, Path]] = None
    ) -> Path:
        """Export player's card collection to CSV."""
        return await self.csv_exporter.export_card_collection_csv(player_tag, filepath)

    async def export_battle_log_csv(
        self,
        player_tag: str,
        limit: int = 100,
        filepath: Optional[Union[str, Path]] = None,
    ) -> Path:
        """Export battle log to CSV."""
        return await self.csv_exporter.export_battle_log_csv(
            player_tag, limit, filepath
        )

    async def export_chest_cycle_csv(
        self, player_tag: str, filepath: Optional[Union[str, Path]] = None
    ) -> Path:
        """Export chest cycle to CSV."""
        return await self.csv_exporter.export_chest_cycle_csv(player_tag, filepath)

    async def export_card_database_csv(
        self, filepath: Optional[Union[str, Path]] = None
    ) -> Path:
        """Export complete card database to CSV."""
        return await self.csv_exporter.export_card_database_csv(filepath)

    async def export_all_data_csv(
        self, player_tag: str, directory: Optional[Union[str, Path]] = None
    ) -> list[Path]:
        """Export all available data types for a player to separate CSV files."""
        return await self.csv_exporter.export_all_data_csv(player_tag, directory)

    # Event Deck Methods
    async def scan_and_import_event_decks(
        self, player_tag: str, days_back: int = 7
    ) -> int:
        """
        Scan battle logs and automatically import event decks.

        Args:
            player_tag: Player's tag
            days_back: Number of days to scan in battle logs

        Returns:
            Number of event decks imported
        """
        if not player_tag.startswith("#"):
            player_tag = f"#{player_tag}"

        print(f"Scanning battle logs for event decks (last {days_back} days)...")

        # Get player battle logs
        battle_logs = await self.get_player_battle_log(player_tag)

        # Filter by date if needed
        if days_back > 0:
            cutoff_date = datetime.now().replace(tzinfo=None) - timedelta(
                days=days_back
            )
            filtered_logs = []
            for battle in battle_logs:
                battle_time_str = battle.get("battleTime", "")
                try:
                    # Parse battle time (format: 20241208T123456.000Z)
                    battle_time = datetime.strptime(
                        battle_time_str.split(".")[0], "%Y%m%dT%H%M%S"
                    )
                    if battle_time >= cutoff_date:
                        filtered_logs.append(battle)
                except Exception:
                    # If we can't parse the time, include it
                    filtered_logs.append(battle)
            battle_logs = filtered_logs

        # Import event decks
        imported_decks = await self.event_manager.import_from_battle_logs(
            battle_logs, player_tag
        )

        print(f"Imported {len(imported_decks)} event decks from battle logs")
        return len(imported_decks)

    async def get_event_decks(
        self,
        player_tag: str,
        event_type: Optional[EventType] = None,
        days_back: Optional[int] = None,
        limit: Optional[int] = None,
    ) -> list[dict[str, Any]]:
        """
        Get event decks for a player.

        Args:
            player_tag: Player's tag
            event_type: Filter by event type (optional)
            days_back: Only get decks from last N days (optional)
            limit: Maximum number of decks to return (optional)

        Returns:
            List of EventDeck objects as dictionaries
        """
        if not player_tag.startswith("#"):
            player_tag = f"#{player_tag}"

        decks = await self.event_manager.get_event_decks(
            player_tag, event_type=event_type, days_back=days_back, limit=limit
        )

        return [deck.dict() for deck in decks]

    async def export_event_decks(
        self,
        player_tag: str,
        filepath: Union[str, Path],
        format: str = "csv",
        event_type: Optional[EventType] = None,
    ) -> Path:
        """
        Export event decks to file.

        Args:
            player_tag: Player's tag
            filepath: Output file path
            format: Export format (csv, json, or deck_builder)
            event_type: Filter by event type (optional)

        Returns:
            Path to the exported file
        """
        if not player_tag.startswith("#"):
            player_tag = f"#{player_tag}"

        filepath = Path(filepath)

        if format.lower() == "csv":
            return await self.event_manager.export_to_csv(
                player_tag, str(filepath), event_type
            )
        elif format.lower() == "json":
            return await self.event_manager.export_to_json(
                player_tag, str(filepath), event_type
            )
        elif format.lower() in ["deck_builder", "decklink"]:
            return await self.event_manager.export_deck_builder_format(
                player_tag, str(filepath), event_type
            )
        else:
            raise ValueError(
                f"Unsupported format: {format}. Use csv, json, or deck_builder"
            )

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
        if not player_tag.startswith("#"):
            player_tag = f"#{player_tag}"

        return await self.event_manager.analyze_event_decks(player_tag, event_type)

    async def get_active_challenges(self) -> list[dict[str, Any]]:
        """
        Get information about currently active challenges.
        Note: This is a placeholder as the official API doesn't provide this endpoint.
        This would need to be implemented with web scraping or a third-party API.

        Returns:
            List of active challenges
        """
        # This is a placeholder implementation
        # In a real scenario, you might:
        # 1. Scrape the official Clash Royale website
        # 2. Use a third-party API like ClashStats
        # 3. Parse in-game client data

        return [
            {
                "name": "Lava Hound Challenge",
                "type": "challenge",
                "max_wins": 12,
                "cost": 100,
                "description": "Win 12 times with decks containing Lava Hound",
            },
            {
                "name": "Draft Challenge",
                "type": "draft_challenge",
                "max_wins": 12,
                "cost": 500,
                "description": "Draft cards before each battle",
            },
        ]
