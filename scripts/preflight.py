#!/usr/bin/env python3
from __future__ import annotations

import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def run(cmd: list[str], cwd: Path | None = None) -> None:
    print(f"+ {' '.join(cmd)}")
    subprocess.run(cmd, cwd=str(cwd or ROOT), check=True)


def main() -> int:
    try:
        run(
            [
                sys.executable,
                ".github/scripts/validate_mappings.py",
                "hwexp/configs/mappings.yaml",
                "monitoring/collector/config/hwexp/mappings.yaml",
            ]
        )
        run(
            [
                sys.executable,
                ".github/scripts/validate_query_migration_matrix.py",
                "monitoring/dashboard/migrations/query-migration-matrix.yaml",
            ]
        )
        run(["go", "test", "./..."], cwd=ROOT / "hwexp")
        run(["docker", "compose", "-f", "monitoring/collector/docker-compose.yml", "config"])
        run(["docker", "compose", "-f", "monitoring/dashboard/docker-compose.dietpi.yml", "config"])
    except subprocess.CalledProcessError as exc:
        print(f"Preflight failed: {exc}", file=sys.stderr)
        return 1

    print("Preflight passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
