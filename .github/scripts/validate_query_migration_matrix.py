#!/usr/bin/env python3
from __future__ import annotations

import argparse
import sys
from pathlib import Path

import yaml

REQUIRED_KEYS = {"id", "old_query", "new_query", "dashboards"}


def validate(path: Path) -> list[str]:
    errors: list[str] = []
    try:
        data = yaml.safe_load(path.read_text(encoding="utf-8"))
    except Exception as exc:
        return [f"{path}: failed to parse YAML: {exc}"]

    if not isinstance(data, dict):
        return [f"{path}: top-level must be a YAML object"]
    migrations = data.get("migrations")
    if not isinstance(migrations, list):
        return [f"{path}: missing migrations list"]

    seen_ids: set[str] = set()
    for idx, item in enumerate(migrations):
        if not isinstance(item, dict):
            errors.append(f"{path}: migrations[{idx}] is not an object")
            continue
        missing = REQUIRED_KEYS.difference(item.keys())
        if missing:
            errors.append(f"{path}: migrations[{idx}] missing keys: {sorted(missing)}")
            continue
        mid = str(item["id"])
        if mid in seen_ids:
            errors.append(f"{path}: duplicate migration id '{mid}'")
        seen_ids.add(mid)
        dashboards = item.get("dashboards")
        if not isinstance(dashboards, list) or not dashboards:
            errors.append(f"{path}: migration '{mid}' dashboards must be a non-empty list")
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate dashboard query migration matrix")
    parser.add_argument("matrix", help="Path to query migration matrix YAML")
    args = parser.parse_args()

    errors = validate(Path(args.matrix))
    if errors:
        print("Query migration matrix validation failed:")
        for err in errors:
            print(f"- {err}")
        return 1
    print("Query migration matrix validation passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
