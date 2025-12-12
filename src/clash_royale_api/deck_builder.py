"""
Utilities for building Clash Royale decks from saved card analysis data.
"""

from __future__ import annotations

import json
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Any, Iterable, Optional


@dataclass
class CardCandidate:
    """Lightweight container for card data used during deck selection."""

    name: str
    level: int
    max_level: int
    rarity: str
    elixir: int
    role: str | None
    score: float


class DeckBuilder:
    """Builds a balanced 1v1 ladder deck from card analysis data."""

    def __init__(self, data_dir: Optional[Path] = None):
        self.data_dir = data_dir or Path("data")
        self.data_dir.mkdir(exist_ok=True)

        # Simple card role groups to keep the deck balanced
        self.role_groups: dict[str, set[str]] = {
            "win_conditions": {
                "Royal Giant",
                "Hog Rider",
                "Giant",
                "P.E.K.K.A",
                "Giant Skeleton",
                "Goblin Barrel",
                "Mortar",
                "X-Bow",
                "Royal Hogs",
            },
            "buildings": {
                "Cannon",
                "Goblin Cage",
                "Inferno Tower",
                "Bomb Tower",
                "Tombstone",
                "Goblin Hut",
                "Barbarian Hut",
            },
            "spells_big": {"Fireball", "Poison", "Lightning", "Rocket"},
            "spells_small": {
                "Zap",
                "Arrows",
                "Giant Snowball",
                "Barbarian Barrel",
                "Freeze",
                "Log",
            },
            "support": {
                "Archers",
                "Bomber",
                "Musketeer",
                "Wizard",
                "Mega Minion",
                "Valkyrie",
                "Baby Dragon",
                "Skeleton Dragons",
            },
            "cycle": {
                "Knight",
                "Skeletons",
                "Ice Spirit",
                "Electro Spirit",
                "Fire Spirit",
                "Bats",
                "Spear Goblins",
                "Goblin Gang",
                "Minions",
            },
        }

        # Fallback elixir costs for common cards (API data sometimes omits cost)
        self.fallback_elixir: dict[str, int] = {
            "Royal Giant": 6,
            "Hog Rider": 4,
            "Giant": 5,
            "P.E.K.K.A": 7,
            "Giant Skeleton": 6,
            "Goblin Barrel": 3,
            "Mortar": 4,
            "X-Bow": 6,
            "Royal Hogs": 5,
            "Cannon": 3,
            "Goblin Cage": 4,
            "Inferno Tower": 5,
            "Bomb Tower": 4,
            "Tombstone": 3,
            "Goblin Hut": 5,
            "Barbarian Hut": 6,
            "Fireball": 4,
            "Poison": 4,
            "Lightning": 6,
            "Rocket": 6,
            "Zap": 2,
            "Arrows": 3,
            "Giant Snowball": 2,
            "Barbarian Barrel": 2,
            "Freeze": 4,
            "Log": 2,
            "Archers": 3,
            "Bomber": 2,
            "Musketeer": 4,
            "Wizard": 5,
            "Mega Minion": 3,
            "Valkyrie": 4,
            "Baby Dragon": 4,
            "Skeleton Dragons": 4,
            "Knight": 3,
            "Skeletons": 1,
            "Ice Spirit": 1,
            "Electro Spirit": 1,
            "Fire Spirit": 1,
            "Bats": 2,
            "Spear Goblins": 2,
            "Goblin Gang": 3,
            "Minions": 3,
            "Goblin Barrel": 3,
            "Bomber": 2,
        }

        self.rarity_weights = {
            "Common": 1.0,
            "Rare": 1.05,
            "Epic": 1.1,
            "Legendary": 1.15,
            "Champion": 1.2,
        }

    def build_deck_from_analysis(self, analysis: dict[str, Any]) -> dict[str, Any]:
        """
        Build a recommended 1v1 deck from a card analysis payload.

        Returns a dictionary containing deck list, average elixir, and helpful notes.
        """
        card_levels = analysis.get("card_levels", {})
        if not card_levels:
            raise ValueError("Analysis data missing 'card_levels'")

        candidates = [
            self._build_candidate(name, data) for name, data in card_levels.items()
        ]

        deck: list[CardCandidate] = []
        used: set[str] = set()
        notes: list[str] = []

        # Core roles: win condition, building, two spells
        win_condition = self._pick_best("win_conditions", candidates, used)
        if win_condition:
            deck.append(win_condition)
            used.add(win_condition.name)
        else:
            notes.append("No win condition found; selected highest power cards instead.")

        building = self._pick_best("buildings", candidates, used)
        if building:
            deck.append(building)
            used.add(building.name)

        big_spell = self._pick_best("spells_big", candidates, used)
        if big_spell:
            deck.append(big_spell)
            used.add(big_spell.name)

        small_spell = self._pick_best("spells_small", candidates, used)
        if small_spell:
            deck.append(small_spell)
            used.add(small_spell.name)

        # Support backbone
        support_cards = self._pick_many("support", candidates, used, count=2)
        deck.extend(support_cards)
        used.update(card.name for card in support_cards)

        # Cheap cycle fillers to keep average elixir reasonable
        cycle_cards = self._pick_many("cycle", candidates, used, count=2)
        deck.extend(cycle_cards)
        used.update(card.name for card in cycle_cards)

        # Fill any remaining slots with the highest score cards not yet used
        if len(deck) < 8:
            remaining = [
                c for c in sorted(candidates, key=lambda c: c.score, reverse=True)
                if c.name not in used
            ]
            deck.extend(remaining[: 8 - len(deck)])

        deck = deck[:8]

        average_elixir = round(
            sum(card.elixir for card in deck) / len(deck), 2
        ) if deck else 0

        # Add friendly notes to guide the player
        if not building:
            notes.append("No defensive building available; play troops high to kite.")
        if not (big_spell or small_spell):
            notes.append("No spell picked; beware of swarm matchups.")
        if average_elixir > 3.8:
            notes.append("High average elixir; play patiently and build pushes.")
        elif average_elixir < 2.8:
            notes.append("Low average elixir; pressure often and out-cycle counters.")

        return {
            "deck": [card.name for card in deck],
            "deck_detail": [
                {
                    "name": card.name,
                    "level": card.level,
                    "max_level": card.max_level,
                    "rarity": card.rarity,
                    "elixir": card.elixir,
                    "role": card.role,
                    "score": round(card.score, 3),
                }
                for card in deck
            ],
            "average_elixir": average_elixir,
            "analysis_time": analysis.get("analysis_time"),
            "notes": notes,
        }

    def build_deck_from_file(self, analysis_path: Path) -> dict[str, Any]:
        """Convenience wrapper to load analysis data from disk."""
        analysis = self.load_analysis(analysis_path)
        return self.build_deck_from_analysis(analysis)

    def load_analysis(
        self, analysis_path: Path | str, strict: bool = True
    ) -> dict[str, Any]:
        """Load analysis JSON from disk."""
        path = Path(analysis_path)
        if not path.exists():
            if strict:
                raise FileNotFoundError(f"Analysis file not found: {path}")
            return {}
        with open(path, "r") as f:
            return json.load(f)

    def load_latest_analysis(
        self, player_tag: str, analysis_dir: Optional[Path] = None
    ) -> Optional[dict[str, Any]]:
        """Get the most recent analysis JSON for a player, if it exists."""
        analysis_dir = analysis_dir or (self.data_dir / "analysis")
        if not analysis_dir.exists():
            return None

        clean_tag = player_tag.lstrip("#")
        candidates = sorted(
            analysis_dir.glob(f"*analysis*{clean_tag}.json"), reverse=True
        )
        if not candidates:
            return None

        return self.load_analysis(candidates[0], strict=False)

    def save_deck(
        self, deck_data: dict[str, Any], output_dir: Optional[Path], player_tag: str
    ) -> Path:
        """Persist a generated deck to disk for later reference."""
        output_dir = output_dir or (self.data_dir / "decks")
        output_dir.mkdir(parents=True, exist_ok=True)

        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        filename = f"{timestamp}_deck_{player_tag.lstrip('#')}.json"
        path = output_dir / filename
        with open(path, "w") as f:
            json.dump(deck_data, f, indent=2)
        return path

    def _build_candidate(self, name: str, data: dict[str, Any]) -> CardCandidate:
        """Create a CardCandidate with calculated score."""
        level = data.get("level", 0)
        max_level = data.get("max_level", level or 1)
        rarity = data.get("rarity", "Common")
        elixir = self._resolve_elixir(name, data)
        role = self._infer_role(name)
        score = self._score_card(level, max_level, rarity, elixir, role)
        return CardCandidate(
            name=name,
            level=level,
            max_level=max_level,
            rarity=rarity,
            elixir=elixir,
            role=role,
            score=score,
        )

    def _resolve_elixir(self, name: str, data: dict[str, Any]) -> int:
        """Use analysis value when present; otherwise fall back to static costs."""
        elixir = data.get("elixir", 0) or 0
        if elixir > 0:
            return elixir
        return self.fallback_elixir.get(name, 4)

    def _infer_role(self, name: str) -> Optional[str]:
        """Infer the primary role for the card based on name membership."""
        for role, names in self.role_groups.items():
            if name in names:
                return role
        return None

    def _score_card(
        self, level: int, max_level: int, rarity: str, elixir: int, role: Optional[str]
    ) -> float:
        """Score cards by power, rarity, and efficiency."""
        level_ratio = (level / max_level) if max_level else 0
        rarity_boost = self.rarity_weights.get(rarity, 1.0)
        # Encourage cheaper cards slightly to keep cycle tight
        elixir_weight = 1 - max(elixir - 3, 0) / 9
        role_bonus = 0.05 if role else 0
        return (level_ratio * 1.2 * rarity_boost) + (elixir_weight * 0.15) + role_bonus

    def _pick_best(
        self, role: str, candidates: Iterable[CardCandidate], used: set[str]
    ) -> Optional[CardCandidate]:
        """Pick the best-scoring unused card for a role."""
        pool = [
            c for c in candidates if c.name not in used and c.name in self.role_groups[role]
        ]
        if not pool:
            return None
        return sorted(pool, key=lambda c: c.score, reverse=True)[0]

    def _pick_many(
        self,
        role: str,
        candidates: Iterable[CardCandidate],
        used: set[str],
        count: int,
    ) -> list[CardCandidate]:
        """Pick multiple cards for a role, highest score first."""
        pool = [
            c for c in candidates if c.name not in used and c.name in self.role_groups[role]
        ]
        pool.sort(key=lambda c: c.score, reverse=True)
        return pool[:count]
