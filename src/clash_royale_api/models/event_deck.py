"""Event deck data models for Clash Royale challenge and tournament tracking."""

from datetime import datetime, timedelta
from enum import Enum
from typing import Any, Optional

from pydantic import BaseModel, Field


class EventType(str, Enum):
    """Types of events in Clash Royale."""

    CHALLENGE = "challenge"
    TOURNAMENT = "tournament"
    SPECIAL_EVENT = "special_event"
    GRAND_CHALLENGE = "grand_challenge"
    CLASSIC_CHALLENGE = "classic_challenge"
    DRAFT_CHALLENGE = "draft_challenge"
    SUDDEN_DEATH = "sudden_death"
    DOUBLE_ELIMINATION = "double_elimination"


class EventProgress(str, Enum):
    """Progress status of an event."""

    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    ELIMINATED = "eliminated"
    FORFEITED = "forfeited"


class CardInDeck(BaseModel):
    """Represents a card in a deck with its level."""

    name: str
    id: int
    level: int
    max_level: int
    rarity: str
    elixir_cost: int


class BattleRecord(BaseModel):
    """Represents a single battle within an event."""

    timestamp: datetime
    opponent_tag: str
    opponent_name: Optional[str] = None
    result: str  # "win" or "loss"
    crowns: int
    opponent_crowns: int
    trophy_change: Optional[int] = None
    battle_mode: Optional[str] = None


class Deck(BaseModel):
    """Represents a deck of 8 cards."""

    cards: list[CardInDeck]
    avg_elixir: float = Field(..., ge=0, le=10)

    class Config:
        json_encoders = {datetime: lambda v: v.isoformat()}


class EventPerformance(BaseModel):
    """Performance metrics for an event."""

    wins: int = Field(0, ge=0)
    losses: int = Field(0, ge=0)
    win_rate: float = Field(0, ge=0, le=1)
    crowns_earned: int = Field(0, ge=0)
    crowns_lost: int = Field(0, ge=0)
    max_wins: Optional[int] = None
    current_streak: int = Field(0, ge=0)
    best_streak: int = Field(0, ge=0)
    progress: EventProgress = EventProgress.IN_PROGRESS


class EventDeck(BaseModel):
    """Complete representation of an event deck with metadata and performance."""

    event_id: str = Field(..., description="Unique identifier for the event")
    player_tag: str = Field(..., description="Player tag (with #)")
    event_name: str = Field(..., description="Name of the event")
    event_type: EventType = Field(..., description="Type of event")
    start_time: datetime = Field(..., description="When the event started")
    end_time: Optional[datetime] = Field(None, description="When the event ended")
    deck: Deck = Field(..., description="The deck used in the event")
    performance: EventPerformance = Field(..., description="Performance metrics")
    battles: list[BattleRecord] = Field(
        default_factory=list, description="All battles in this event"
    )
    event_rules: Optional[dict[str, Any]] = Field(
        None, description="Special rules for the event"
    )
    notes: Optional[str] = Field(None, description="Additional notes about the event")

    class Config:
        json_encoders = {datetime: lambda v: v.isoformat()}

    def add_battle(self, battle: BattleRecord) -> None:
        """Add a battle record and update performance metrics."""
        self.battles.append(battle)

        # Update wins/losses
        if battle.result == "win":
            self.performance.wins += 1
            self.performance.crowns_earned += battle.crowns
            self.performance.current_streak += 1
            self.performance.best_streak = max(
                self.performance.best_streak, self.performance.current_streak
            )
        else:
            self.performance.losses += 1
            self.performance.crowns_lost += battle.opponent_crowns
            self.performance.current_streak = 0

        # Update win rate
        total_battles = self.performance.wins + self.performance.losses
        if total_battles > 0:
            self.performance.win_rate = self.performance.wins / total_battles

        # Update progress based on event type
        self._update_progress()

    def _update_progress(self) -> None:
        """Update progress based on current state."""
        if self.performance.max_wins:
            if self.performance.wins >= self.performance.max_wins:
                self.performance.progress = EventProgress.COMPLETED
            elif self.performance.losses >= 3:  # Typically 3 losses eliminates
                self.performance.progress = EventProgress.ELIMINATED

        # If end_time is set, mark as completed
        if self.end_time and self.end_time < datetime.now():
            if self.performance.progress == EventProgress.IN_PROGRESS:
                self.performance.progress = EventProgress.COMPLETED


class EventDeckCollection(BaseModel):
    """Collection of event decks for a player."""

    player_tag: str
    decks: list[EventDeck] = Field(default_factory=list)
    last_updated: datetime = Field(default_factory=datetime.now)

    class Config:
        json_encoders = {datetime: lambda v: v.isoformat()}

    def add_deck(self, deck: EventDeck) -> None:
        """Add a new event deck to the collection."""
        # Check if event_id already exists
        for i, existing in enumerate(self.decks):
            if existing.event_id == deck.event_id:
                # Update existing deck
                self.decks[i] = deck
                self.last_updated = datetime.now()
                return

        # Add new deck
        self.decks.append(deck)
        self.last_updated = datetime.now()

    def get_decks_by_type(self, event_type: EventType) -> list[EventDeck]:
        """Get all decks of a specific event type."""
        return [deck for deck in self.decks if deck.event_type == event_type]

    def get_recent_decks(self, days: int = 7) -> list[EventDeck]:
        """Get decks from the last N days."""
        cutoff = datetime.now() - timedelta(days=days)
        return [deck for deck in self.decks if deck.start_time >= cutoff]

    def get_best_decks_by_win_rate(self, min_battles: int = 5) -> list[EventDeck]:
        """Get decks sorted by win rate (minimum battles filter)."""
        qualified = [
            deck
            for deck in self.decks
            if (deck.performance.wins + deck.performance.losses) >= min_battles
        ]
        return sorted(qualified, key=lambda d: d.performance.win_rate, reverse=True)
