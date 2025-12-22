#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.9"
# dependencies = ["pyyaml"]
# ///
import os
import subprocess
import sys
from pathlib import Path


def main(argv: list[str]) -> int:
    if len(argv) < 2:
        print("usage: uv run scripts/quick_validate_skill.py <skill-path> [args...]", file=sys.stderr)
        return 2

    codex_home = os.environ.get("CODEX_HOME")
    if codex_home:
        base = Path(codex_home)
    else:
        base = Path.home() / ".codex"

    validator = base / "skills" / ".system" / "skill-creator" / "scripts" / "quick_validate.py"
    if not validator.exists():
        print(f"error: quick_validate.py not found at {validator}", file=sys.stderr)
        return 1

    cmd = [sys.executable, str(validator)] + argv[1:]
    try:
        subprocess.run(cmd, check=True)
    except subprocess.CalledProcessError as exc:
        return exc.returncode
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
