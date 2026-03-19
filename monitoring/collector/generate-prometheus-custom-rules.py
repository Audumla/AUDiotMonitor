#!/usr/bin/env python3
"""Generate host-specific Prometheus recording rules for detected GPUs.

Writes a custom rules file only when one does not already exist unless --force
is supplied. The generated rules add stable gpu_index / gpu_name labels for the
locally detected DRM cards so dashboards can avoid hard-coded PCI IDs.
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
    uevent = card_device / "uevent"
    if not uevent.exists():
        return None
    for line in uevent.read_text().splitlines():
        if line.startswith("PCI_SLOT_NAME="):
            return f"pci-{line.split('=', 1)[1].strip()}"
    return None


def discover_gpu_ids() -> list[str]:
    gpu_ids: list[str] = []
    for card in sorted(DRM_PATH.glob("card[0-9]*")):
        device = card / "device"
        if not device.exists():
            continue
        slot = pci_slot_for(device)
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
        return "\n".join(lines) + "\n"

    for idx, gpu_id in enumerate(gpu_ids, start=1):
        gpu_name = f"gpu{idx}"
        for record, expr in [
            (
                "audiot_system_gpu_compute_utilization_percent",
                f'hw_device_utilization_percent{{device_class="gpu", component="compute", sensor="utilization", device_id="{gpu_id}"}}',
            ),
            (
                "audiot_system_gpu_memory_utilization_percent",
                f'hw_device_utilization_percent{{device_class="gpu", component="memory", sensor="utilization", device_id="{gpu_id}"}}',
            ),
            (
                "audiot_system_gpu_vram_usage_percent",
                f'audiot_gpu_vram_usage_percent{{device_id="{gpu_id}"}}',
            ),
            (
                "audiot_system_gpu_vram_used_bytes",
                f'audiot_gpu_vram_used_bytes{{device_id="{gpu_id}"}}',
            ),
            (
                "audiot_system_gpu_vram_capacity_bytes",
                f'audiot_gpu_vram_capacity_bytes{{device_id="{gpu_id}"}}',
            ),
        ]:
            lines.extend(
                [
                    f"      - record: {record}",
                    f"        expr: {expr}",
                    "        labels:",
                    f'          gpu_index: "{idx}"',
                    f'          gpu_name: "{gpu_name}"',
                    f'          source_device_id: "{gpu_id}"',
                ]
            )
    return "\n".join(lines) + "\n"


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
