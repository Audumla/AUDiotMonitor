#!/usr/bin/env python3
"""Generate host-specific Prometheus recording rules for detected GPUs.

This script creates a rules file that adds stable, indexed aliases for GPUs
(e.g., gpu_index="1") based on their detected PCI slot. This allows dashboards
to be built without hard-coding PCI device IDs.
"""

from __future__ import annotations

import argparse
import os
from pathlib import Path


DEFAULT_OUTPUT = Path(
    os.environ.get(
        "INSTALL_DIR",
        "/opt/docker/collector",
    )
) / "config/prometheus/rules/custom/system.rules.yml"
DRM_PATH = Path("/sys/class/drm")


def pci_slot_for(card_device: Path) -> str | None:
    uevent = card_device / "device" / "uevent"
    if not uevent.exists():
        return None
    for line in uevent.read_text().splitlines():
        if line.startswith("PCI_SLOT_NAME="):
            return f"pci-{line.split('=', 1)[1].strip()}"
    return None


def discover_gpu_ids() -> list[str]:
    gpu_ids: list[str] = []
    for card in sorted(DRM_PATH.glob("card[0-9]*")):
        if not (card / "device").exists():
            continue
        slot = pci_slot_for(card)
        if slot and slot not in gpu_ids:
            gpu_ids.append(slot)
    return gpu_ids


def build_rules(gpu_ids: list[str]) -> str:
    lines = [
        "groups:",
        "  - name: audiot_system_gpu_aliases",
        "    interval: 15s",
        "    rules:",
    ]
    if not gpu_ids:
        lines.extend(
            [
                "      - record: audiot_system_gpu_count",
                "        expr: vector(0)",
            ]
        )
        return "
".join(lines) + "
"

    # This group adds a stable 'gpu_index' label to existing raw metrics.
    # It allows dashboards to query, for example, 'audiot_system_gpu_utilization{gpu_index="1"}'
    # without needing to know the underlying PCI device_id.
    for idx, gpu_id in enumerate(gpu_ids, start=1):
        gpu_name = f"gpu{idx}"
        for record, source_metric in [
            ("audiot_system_gpu_compute_utilization_percent", "hw_device_utilization_percent"),
            ("audiot_system_gpu_memory_utilization_percent", "hw_device_utilization_percent"),
        ]:
            lines.extend(
                [
                    f"      - record: {record}",
                    f"        expr: {source_metric}{{device_id='{gpu_id}'}}",
                    "        labels:",
                    f'          gpu_index: "{idx}"',
                    f'          gpu_name: "{gpu_name}"',
                ]
            )
    return "
".join(lines) + "
"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", default=str(DEFAULT_OUTPUT))
    parser.add_argument("--force", action="store_true")
    args = parser.parse_args()

    output = Path(args.output)
    if output.exists() and not args.force:
        print(f"[rules] Keeping existing custom rules: {output}")
        return 0

    output.parent.mkdir(parents=True, exist_ok=True)
    gpu_ids = discover_gpu_ids()
    output.write_text(build_rules(gpu_ids))
    print(f"[rules] Wrote {output} for {len(gpu_ids)} GPU(s)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

