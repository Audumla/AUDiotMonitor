# Metric Migration Matrix Template

Use this template for any metric schema migration.

| Area | Old Metric / Query | New Metric / Query | Affected Dashboards | Cutover Release |
| --- | --- | --- | --- | --- |
| GPU VRAM used | `hw_device_capacity_bytes{logical_name=~".*_vram_used_bytes"}` | `hw_device_capacity_bytes{component="memory",sensor="used",memory_kind="vram"}` | `system-overview.json`, `triple-gpu-wide.json` | `TBD` |

Checklist:
1. Update dashboards and alert rules in the same release.
2. Add/adjust regression tests and golden checks.
3. Record cutover release.
