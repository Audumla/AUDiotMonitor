#!/usr/bin/env python3
from __future__ import annotations

import argparse
import sys
from pathlib import Path

import yaml

ALLOWED_METRIC_TYPES = {"gauge", "counter"}


def validate_file(path: Path) -> list[str]:
    errors: list[str] = []
    try:
        data = yaml.safe_load(path.read_text(encoding="utf-8"))
    except Exception as exc:  # pragma: no cover - surfaced as CI error
        return [f"{path}: failed to parse YAML: {exc}"]

    if not isinstance(data, dict):
        return [f"{path}: expected YAML object at top level"]

    rules = data.get("rules")
    if not isinstance(rules, list):
        return [f"{path}: expected 'rules' list"]

    for idx, rule in enumerate(rules):
        if not isinstance(rule, dict):
            errors.append(f"{path}: rules[{idx}] is not an object")
            continue
        rid = rule.get("id", f"index:{idx}")
        normalize = rule.get("normalize")
        if not isinstance(normalize, dict):
            errors.append(f"{path}: rule '{rid}' missing normalize object")
            continue
        mtype = normalize.get("metric_type")
        if mtype not in ALLOWED_METRIC_TYPES:
            errors.append(
                f"{path}: rule '{rid}' has invalid metric_type '{mtype}' "
                f"(allowed: {sorted(ALLOWED_METRIC_TYPES)})"
            )
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate mapping YAML contracts")
    parser.add_argument("files", nargs="+", help="mapping YAML files to validate")
    args = parser.parse_args()

    all_errors: list[str] = []
    for raw in args.files:
        all_errors.extend(validate_file(Path(raw)))

    if all_errors:
        print("Mapping validation failed:")
        for err in all_errors:
            print(f"- {err}")
        return 1

    print("Mapping validation passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
