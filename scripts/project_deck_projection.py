#!/usr/bin/env python3
import argparse
import json
import os
import sys
from copy import deepcopy
from datetime import datetime, timezone

UPGRADE_COSTS = {
    "Common": {
        1: 2,
        2: 4,
        3: 10,
        4: 20,
        5: 50,
        6: 100,
        7: 200,
        8: 400,
        9: 800,
        10: 1000,
        11: 2000,
        12: 5000,
        13: 10000,
    },
    "Rare": {
        3: 2,
        4: 4,
        5: 10,
        6: 20,
        7: 50,
        8: 100,
        9: 200,
        10: 400,
        11: 800,
        12: 1000,
        13: 2000,
    },
    "Epic": {
        6: 2,
        7: 4,
        8: 10,
        9: 20,
        10: 50,
        11: 100,
        12: 200,
        13: 400,
    },
    "Legendary": {
        9: 2,
        10: 4,
        11: 10,
        12: 20,
        13: 40,
    },
    "Champion": {
        11: 2,
        12: 4,
        13: 10,
    },
}

MAX_LEVELS = {
    "Common": 14,
    "Rare": 14,
    "Epic": 14,
    "Legendary": 14,
    "Champion": 14,
}

RARITY_ALIASES = {
    "common": "Common",
    "rare": "Rare",
    "epic": "Epic",
    "legendary": "Legendary",
    "champion": "Champion",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Project upgrades onto a saved analysis JSON and emit a new projected analysis file."
    )
    parser.add_argument("--analysis", required=True, help="Path to analysis JSON (from ./bin/cr-api analyze --save)")
    parser.add_argument("--plan", required=True, help="Path to upgrade plan JSON")
    parser.add_argument("--output", help="Output path for projected analysis JSON")
    parser.add_argument("--tag", help="Player tag (used only for printing a cr-api deck build command)")
    parser.add_argument(
        "--unbounded",
        action="store_true",
        help="Ignore wildcard affordability checks and allow any upgrades",
    )
    parser.add_argument("--dry-run", action="store_true", help="Show planned changes without writing output")
    return parser.parse_args()


def load_json(path: str) -> dict:
    with open(path, "r", encoding="utf-8") as handle:
        return json.load(handle)


def write_json(path: str, payload: dict) -> None:
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w", encoding="utf-8") as handle:
        json.dump(payload, handle, indent=2, sort_keys=True)


def normalize_rarity(value: str) -> str:
    if value is None:
        return ""
    key = value.strip().lower()
    return RARITY_ALIASES.get(key, "")


def cards_needed_to_level(rarity: str, current_level: int, target_level: int) -> int:
    if target_level <= current_level:
        return 0
    costs = UPGRADE_COSTS.get(rarity, {})
    total = 0
    for level in range(current_level, target_level):
        if level not in costs:
            raise ValueError(f"Missing upgrade cost for {rarity} at level {level}")
        total += costs[level]
    return total


def cards_needed_for_next(rarity: str, level: int, max_level: int) -> int:
    if level >= max_level:
        return 0
    return UPGRADE_COSTS.get(rarity, {}).get(level, 0)


def resolve_card_name(card_levels: dict, name: str) -> str:
    lookup = {key.lower(): key for key in card_levels.keys()}
    return lookup.get(name.strip().lower(), "")


def main() -> int:
    args = parse_args()
    analysis = load_json(args.analysis)

    card_levels = analysis.get("card_levels")
    if not isinstance(card_levels, dict) or not card_levels:
        print("error: analysis JSON missing card_levels map", file=sys.stderr)
        return 1

    plan = load_json(args.plan)
    upgrades = plan.get("upgrades", [])
    if not isinstance(upgrades, list) or not upgrades:
        print("error: plan JSON missing upgrades list", file=sys.stderr)
        return 1

    raw_wildcards = plan.get("wildcards", {})
    wildcards_available = {rarity: 0 for rarity in MAX_LEVELS.keys()}
    for key, value in raw_wildcards.items():
        rarity = normalize_rarity(key)
        if not rarity:
            print(f"error: unknown wildcard rarity '{key}'", file=sys.stderr)
            return 1
        wildcards_available[rarity] = int(value)

    projected = deepcopy(analysis)
    projected_levels = projected.get("card_levels", {})

    upgrades_applied = []
    wildcards_spent = {rarity: 0 for rarity in MAX_LEVELS.keys()}
    wildcards_remaining = deepcopy(wildcards_available)

    for upgrade in upgrades:
        card_name = upgrade.get("card")
        target_level = upgrade.get("target_level")
        if not card_name or target_level is None:
            print("error: upgrade entries require card and target_level", file=sys.stderr)
            return 1

        resolved = resolve_card_name(projected_levels, card_name)
        if not resolved:
            print(f"error: card not found in analysis: {card_name}", file=sys.stderr)
            return 1

        info = projected_levels[resolved]
        current_level = int(info.get("level", 0))
        rarity = normalize_rarity(info.get("rarity", ""))
        if not rarity:
            print(f"error: card {resolved} has unknown rarity", file=sys.stderr)
            return 1

        max_level = int(info.get("max_level", MAX_LEVELS[rarity]))
        if target_level > max_level:
            print(
                f"error: target level {target_level} exceeds max {max_level} for {resolved}",
                file=sys.stderr,
            )
            return 1

        if target_level <= current_level:
            upgrades_applied.append(
                {
                    "card": resolved,
                    "current_level": current_level,
                    "target_level": target_level,
                    "cards_required": 0,
                    "wildcards_spent": 0,
                    "note": "target not above current; no change",
                }
            )
            continue

        card_count = info.get("card_count")
        if card_count is None:
            print(
                f"error: card_count missing for {resolved}. Use analysis from ./bin/cr-api analyze --save",
                file=sys.stderr,
            )
            return 1
        card_count = int(card_count)

        required = cards_needed_to_level(rarity, current_level, target_level)
        available = card_count + wildcards_remaining.get(rarity, 0)
        if not args.unbounded and required > available:
            print(
                f"error: insufficient cards for {resolved} ({rarity}). Need {required}, have {available} (cards + wildcards).",
                file=sys.stderr,
            )
            return 1

        shortfall = max(0, required - card_count)
        if not args.unbounded:
            wildcards_remaining[rarity] -= shortfall
        wildcards_spent[rarity] += shortfall

        new_card_count = max(0, card_count - required)
        info["level"] = target_level
        info["card_count"] = new_card_count
        info["is_max_level"] = target_level >= max_level
        info["cards_to_next_level"] = cards_needed_for_next(rarity, target_level, max_level)

        upgrades_applied.append(
            {
                "card": resolved,
                "current_level": current_level,
                "target_level": target_level,
                "cards_required": required,
                "wildcards_spent": shortfall,
            }
        )

    projection_meta = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "source_analysis": os.path.abspath(args.analysis),
        "plan": os.path.abspath(args.plan),
        "upgrades_applied": upgrades_applied,
        "wildcards_available": wildcards_available,
        "wildcards_spent": wildcards_spent,
        "unbounded": args.unbounded,
    }
    if not args.unbounded:
        projection_meta["wildcards_remaining"] = wildcards_remaining

    projected["projection"] = projection_meta

    output_path = args.output
    if not output_path:
        base_dir = os.path.dirname(os.path.abspath(args.analysis))
        base_name = os.path.basename(args.analysis)
        timestamp = datetime.now(timezone.utc).strftime("%Y%m%d_%H%M%S")
        output_path = os.path.join(base_dir, f"projected_{timestamp}_{base_name}")

    if args.dry_run:
        print(json.dumps(projection_meta, indent=2, sort_keys=True))
        return 0

    write_json(output_path, projected)
    print(f"Projected analysis saved to {output_path}")

    if args.tag:
        print(f"Deck build command: ./bin/cr-api deck build --tag {args.tag} --from-analysis {output_path}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
