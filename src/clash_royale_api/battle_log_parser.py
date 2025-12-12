"""Parse battle logs to detect and extract event decks."""

import logging
from datetime import datetime, timedelta
from typing import Any, Optional

from .models.event_deck import (
    BattleRecord,
    CardInDeck,
    Deck,
    EventDeck,
    EventPerformance,
    EventType,
)

logger = logging.getLogger(__name__)


class BattleLogParser:
    """Parses battle logs to identify event decks and extract performance data."""

    # Battle modes that indicate special events
    EVENT_BATTLE_MODES = {
        "Challenge": EventType.CHALLENGE,
        "Grand Challenge": EventType.GRAND_CHALLENGE,
        "Classic Challenge": EventType.CLASSIC_CHALLENGE,
        "Draft Challenge": EventType.DRAFT_CHALLENGE,
        "Tournament": EventType.TOURNAMENT,
        "Special Event": EventType.SPECIAL_EVENT,
        "Sudden Death": EventType.SUDDEN_DEATH,
        "Double Elimination": EventType.DOUBLE_ELIMINATION,
    }

    # Event name patterns to detect special events
    EVENT_PATTERNS = {
        "lava": "Lava Hound Challenge",
        "hog": "Hog Rider Challenge",
        "mortar": "Mortar Challenge",
        "graveyard": "Graveyard Challenge",
        "ram rage": "Ram Rage Challenge",
        "sparky": "Sparky Challenge",
        "electro": "Electro Challenge",
        "skeleton": "Skeleton Army Challenge",
        "bandit": "Bandit Challenge",
        "night witch": "Night Witch Challenge",
        "royale": "Clash Royale Championship",
        "worlds": "World Championship",
        "ccgs": "Clash Championship Series",
    }

    def __init__(self):
        self.card_cache = {}  # Cache for card data

    async def parse_battle_logs(
        self, battle_logs: list[dict[str, Any]], player_tag: str
    ) -> list[EventDeck]:
        """
        Parse battle logs and extract event decks.

        Args:
            battle_logs: List of battle log entries from the API
            player_tag: Player's tag for identification

        Returns:
            List of EventDeck objects extracted from the battle logs
        """
        # Group battles by potential events
        event_groups = self._group_battles_by_event(battle_logs, player_tag)

        # Convert groups to EventDeck objects
        event_decks = []
        for event_id, (event_data, battles) in event_groups.items():
            try:
                event_deck = await self._create_event_deck(
                    event_data, battles, player_tag
                )
                if event_deck:
                    event_decks.append(event_deck)
            except Exception as e:
                logger.error(f"Error creating event deck for {event_id}: {e}")
                continue

        return event_decks

    def _group_battles_by_event(
        self, battle_logs: list[dict[str, Any]], player_tag: str
    ) -> dict[str, tuple[dict, list]]:
        """
        Group battles into events based on timing and mode.

        Returns:
            Dict mapping event_id to (event_data, [battles])
        """
        event_groups = {}

        # Sort battles by time
        sorted_battles = sorted(
            [b for b in battle_logs if self._is_event_battle(b)],
            key=lambda x: x.get("battleTime", ""),
        )

        current_event = None
        current_battles = []
        last_battle_time = None

        for battle in sorted_battles:
            battle_time = self._parse_battle_time(battle.get("battleTime", ""))

            # Check if this starts a new event
            if self._is_new_event(battle, current_event, battle_time, last_battle_time):
                # Save previous event if exists
                if current_event and current_battles:
                    event_id = self._generate_event_id(
                        current_event, current_battles[0]
                    )
                    event_groups[event_id] = (current_event, current_battles)

                # Start new event
                current_event = self._extract_event_data(battle)
                current_battles = [battle]
            else:
                # Continue current event
                current_battles.append(battle)

            last_battle_time = battle_time

        # Save last event
        if current_event and current_battles:
            event_id = self._generate_event_id(current_event, current_battles[0])
            event_groups[event_id] = (current_event, current_battles)

        return event_groups

    def _is_event_battle(self, battle: dict[str, Any]) -> bool:
        """Check if a battle is part of an event."""
        # Check battle mode
        mode = battle.get("gameMode", {}).get("name", "").lower()
        for event_mode in self.EVENT_BATTLE_MODES:
            if event_mode.lower() in mode:
                return True

        # Check for special event indicators
        team = battle.get("team", [])
        if len(team) > 0:
            for player in team:
                cards = player.get("cards", [])
                # If player has no cards selected, might be a draft or special mode
                if not cards:
                    return True

        # Check if it's not a regular ladder battle
        if not battle.get("isLadderBattle", True):
            return True

        return False

    def _is_new_event(
        self,
        battle: dict,
        current_event: Optional[dict],
        battle_time: datetime,
        last_battle_time: Optional[datetime],
    ) -> bool:
        """Determine if this battle starts a new event."""
        if not current_event:
            return True

        # Check time gap (events typically don't have large gaps)
        if last_battle_time and (battle_time - last_battle_time) > timedelta(hours=1):
            return True

        # Check if event type changed
        new_event_data = self._extract_event_data(battle)
        if new_event_data.get("event_type") != current_event.get("event_type"):
            return True

        # Check deck composition (significant change might indicate new event)
        team = battle.get("team", [])
        if team:
            current_deck = self._extract_deck_from_battle(battle)
            if current_event.get("deck_cards") != current_deck:
                # Check if it's just a level upgrade
                if not self._is_same_deck_different_levels(
                    current_event.get("deck_cards", []), current_deck
                ):
                    return True

        return False

    def _extract_event_data(self, battle: dict[str, Any]) -> dict[str, Any]:
        """Extract event information from a battle."""
        mode_name = battle.get("gameMode", {}).get("name", "")

        # Determine event type
        event_type = EventType.CHALLENGE  # Default
        for pattern, type_ in self.EVENT_BATTLE_MODES.items():
            if pattern.lower() in mode_name.lower():
                event_type = type_
                break

        # Try to detect specific event name
        event_name = mode_name
        mode_lower = mode_name.lower()
        for pattern, name in self.EVENT_PATTERNS.items():
            if pattern in mode_lower:
                event_name = name
                break

        return {
            "event_type": event_type,
            "event_name": event_name,
            "battle_mode": mode_name,
            "deck_cards": self._extract_deck_from_battle(battle),
        }

    def _extract_deck_from_battle(self, battle: dict[str, Any]) -> list[str]:
        """Extract card names from battle team data."""
        team = battle.get("team", [])
        if not team:
            return []

        # Get first player's cards
        player = team[0]
        cards = player.get("cards", [])

        # Sort by slot position if available
        if all("slot" in card for card in cards):
            cards.sort(key=lambda x: x.get("slot", 0))

        return [card.get("name", "") for card in cards]

    def _is_same_deck_different_levels(
        self, deck1: list[str], deck2: list[str]
    ) -> bool:
        """Check if two decks have the same cards (ignoring levels)."""
        if len(deck1) != len(deck2):
            return False

        # Sort both decks
        sorted1 = sorted(deck1)
        sorted2 = sorted(deck2)

        return sorted1 == sorted2

    def _parse_battle_time(self, battle_time: str) -> datetime:
        """Parse battle time string from API response."""
        # Format: "20241208T123456.000Z"
        try:
            # Remove milliseconds and Z
            clean_time = battle_time.split(".")[0]
            return datetime.strptime(clean_time, "%Y%m%dT%H%M%S")
        except Exception:
            logger.warning(f"Could not parse battle time: {battle_time}")
            return datetime.now()

    def _generate_event_id(self, event_data: dict, battle: dict) -> str:
        """Generate a unique event ID."""
        event_name = event_data.get("event_name", "unknown").lower().replace(" ", "_")
        battle_time = self._parse_battle_time(battle.get("battleTime", ""))
        date_str = battle_time.strftime("%Y%m%d")

        # Add a hash of the deck to make it unique
        deck_cards = event_data.get("deck_cards", [])
        deck_hash = hash(tuple(sorted(deck_cards))) if deck_cards else 0

        return f"{event_name}_{date_str}_{abs(deck_hash) % 1000}"

    async def _create_event_deck(
        self, event_data: dict, battles: list[dict], player_tag: str
    ) -> Optional[EventDeck]:
        """Create an EventDeck object from grouped battles."""
        if not battles:
            return None

        # Extract detailed card information from first battle
        first_battle = battles[0]
        team = first_battle.get("team", [])
        if not team:
            logger.warning("No team data found in battle")
            return None

        player_data = team[0]
        cards_data = player_data.get("cards", [])
        if len(cards_data) != 8:
            logger.warning(f"Expected 8 cards, found {len(cards_data)}")
            return None

        # Create CardInDeck objects
        deck_cards = []
        total_elixir = 0
        for card_data in cards_data:
            card = CardInDeck(
                name=card_data.get("name", ""),
                id=card_data.get("id", 0),
                level=card_data.get("level", 0),
                max_level=card_data.get("maxLevel", 13),
                rarity=card_data.get("rarity", "common"),
                elixir_cost=card_data.get("elixirCost", 0),
            )
            deck_cards.append(card)
            total_elixir += card.elixir_cost

        # Create deck object
        deck = Deck(cards=deck_cards, avg_elixir=total_elixir / 8)

        # Create event performance object
        performance = EventPerformance()

        # Create EventDeck
        start_time = self._parse_battle_time(battles[0].get("battleTime", ""))
        event_deck = EventDeck(
            event_id=self._generate_event_id(event_data, battles[0]),
            player_tag=player_tag,
            event_name=event_data.get("event_name", ""),
            event_type=event_data.get("event_type", EventType.CHALLENGE),
            start_time=start_time,
            end_time=None,  # Will be set later if needed
            deck=deck,
            performance=performance,
            event_rules={
                "battle_mode": event_data.get("battle_mode"),
                "is_ladder": False,
            },
        )

        # Add all battles
        for battle in battles:
            battle_record = self._create_battle_record(battle, player_tag)
            if battle_record:
                event_deck.add_battle(battle_record)

        # Set max_wins based on event type
        event_type = event_data.get("event_type", EventType.CHALLENGE)
        if event_type == EventType.GRAND_CHALLENGE:
            event_deck.performance.max_wins = 10
        elif event_type == EventType.CLASSIC_CHALLENGE:
            event_deck.performance.max_wins = 12
        elif event_type == EventType.DRAFT_CHALLENGE:
            event_deck.performance.max_wins = 12
        else:
            event_deck.performance.max_wins = 20  # Default for special events

        # Update progress
        event_deck._update_progress()

        return event_deck

    def _create_battle_record(
        self, battle: dict, player_tag: str
    ) -> Optional[BattleRecord]:
        """Create a BattleRecord from battle data."""
        team = battle.get("team", [])
        opponent = battle.get("opponent", [])

        if not team or not opponent:
            return None

        # Determine win/loss
        team_crowns = team[0].get("crowns", 0)
        opponent_crowns = opponent[0].get("crowns", 0)
        result = "win" if team_crowns > opponent_crowns else "loss"

        # Get trophy change
        trophy_change = battle.get("team", [{}])[0].get("trophyChange")

        return BattleRecord(
            timestamp=self._parse_battle_time(battle.get("battleTime", "")),
            opponent_tag=opponent[0].get("tag", ""),
            opponent_name=opponent[0].get("name"),
            result=result,
            crowns=team_crowns,
            opponent_crowns=opponent_crowns,
            trophy_change=trophy_change,
            battle_mode=battle.get("gameMode", {}).get("name"),
        )
