# Next Stage Update — Metrics Schema Model Proposal

**Status:** Proposed (Deferred; not part of current baseline)  
**Date:** 2026-03-28  
**Project:** AUDiotMonitor  
**Covers:** Extensible telemetry schema model for raw, canonical, and derived metrics; label contracts; compatibility strategy  
**Related:** [detailed-spec.md](../exporter/detailed-spec.md), [mapping-rule-spec.md](../exporter/mapping-rule-spec.md), [json-schema-spec.md](../exporter/json-schema-spec.md), [dashboard-data-spec.md](../dashboard/dashboard-data-spec.md)

> This document is a next-stage planning artifact only. It does not modify current production schema contracts unless explicitly promoted in a future stage.

---

## 1. Goal / Problem Statement

Define a stable, extensible metric model so adapter changes do not break dashboards and recording rules.

The model MUST separate:
- source/raw telemetry
- canonical semantic telemetry
- derived dashboard-friendly telemetry

---

## 2. Scope and Boundaries

This spec defines:
- canonical metric layer names and labels
- normalization expectations from raw adapters
- derived metric conventions for dashboard use
- compatibility/versioning rules

This spec does not define:
- full dashboard JSON implementation
- adapter-specific parsing internals

---

## 3. Contracts (Inputs / Outputs)

### 3.1 Three-Layer Metric Model

1. `raw_*` layer (optional export for debug/diagnostics)
- source-specific naming
- unstable by source
- not intended for dashboards

2. `hw_*` layer (canonical contract)
- stable semantic families
- strict label contract
- primary integration surface for rules and dashboards

3. `audiot_*` layer (derived/recording)
- rollups, ratios, and display-oriented signals
- calculated from canonical `hw_*`
- no direct adapter writes

### 3.2 Canonical Family Groups

- Thermal: `hw_device_temperature_celsius`
- Utilization: `hw_device_utilization_percent`
- Capacity/usage bytes: `hw_device_capacity_bytes`
- Power/Energy: `hw_device_power_watts`, `hw_device_energy_joules`
- Inventory/info: `hw_device_info`
- Health/errors: `hw_device_*_errors_total`, `hw_device_*_status`

### 3.3 Required Labels (Canonical)

Every `hw_*` metric MUST include:
- `host`
- `platform`
- `source`
- `device_class`
- `device_id`
- `logical_name`
- `component`
- `sensor`

Optional labels MUST be additive only (for example `memory_kind`, `bus`, `vendor`, `model`).

### 3.4 Memory Telemetry Canonicalization

Memory telemetry SHOULD use:
- `component="memory"`
- `sensor` representing semantic dimension (`used`, `capacity`, `usage`)
- `memory_kind` label with controlled values:
  - `vram`
  - `gtt`
  - `ram`
  - `swap`
  - `cache`

Example canonical series:
- `hw_device_capacity_bytes{component="memory",sensor="used",memory_kind="vram",...}`
- `hw_device_capacity_bytes{component="memory",sensor="capacity",memory_kind="vram",...}`
- `hw_device_utilization_percent{component="memory",sensor="usage",memory_kind="vram",...}`

---

## 4. Configuration and Examples

### 4.1 Example Mapping Rule (VRAM Used)

```yaml
- id: "gpu_vram_used_bytes"
  priority: 100
  match:
    platform: "linux"
    device_class: "gpu"
    raw_name_regex: "^vram_used_bytes$"
  normalize:
    metric_family: "hw_device_capacity_bytes"
    metric_type: "gauge"
    logical_name_template: "${logical_device_name}_vram_used_bytes"
    device_class: "gpu"
    component: "memory"
    sensor: "used"
    labels:
      memory_kind: "vram"
```

### 4.2 Example Derived Recording Rule

```yaml
- record: audiot_gpu_vram_usage_percent
  expr: |
    (
      sum by (device_id) (hw_device_capacity_bytes{device_class="gpu", component="memory", sensor="used", memory_kind="vram"})
      /
      sum by (device_id) (hw_device_capacity_bytes{device_class="gpu", component="memory", sensor="capacity", memory_kind="vram"})
    ) * 100
```

### 4.3 Example Query Surface

Dashboard and alerting queries SHOULD use:
- `audiot_*` for display-ready aggregates
- `hw_*` for raw semantic details

They SHOULD NOT use adapter raw names directly.

---

## 5. Error Handling and Edge Cases

- If one canonical memory dimension is missing (`used` without `capacity`), emit available series and omit derived ratio.
- Unknown raw names MUST map to `ignored` unless automapper is explicitly enabled.
- Invalid metric type declarations in mappings MUST fail validation before scrape exposure.
- Adapter-specific metrics that cannot fit canonical families MAY be emitted as scoped `hw_device_custom_*` families with explicit `component` and `sensor`.

---

## 6. Test and Validation Requirements

- Unit tests for mapper matching semantics (`raw_name`, `raw_name_regex`, label expansion).
- Contract tests asserting required labels on every `hw_*` family.
- Integration tests ensuring critical dashboard inputs exist:
  - `vram used`
  - `vram capacity`
  - `vram usage percent`
- Prometheus scrape validation test that rejects invalid exposition types.
- Golden query tests for common dashboard PromQL expressions.

---

## 7. Implementation Notes / Rollout

### 7.1 Versioning

- `metric_schema_version` MUST increment on breaking label/family changes.
- Additive labels/families are minor-compatible.

### 7.2 Compatibility Strategy

- Introduce compatibility recording rules when renaming labels or families.
- Keep old derived names for at least one minor release cycle.
- Add migration notes in changelog and dashboard spec.

### 7.3 Recommended Migration Sequence

1. Define canonical family/label target in mapping rules.
2. Add compatibility recording rules.
3. Update dashboards to canonical/derived queries.
4. Remove compatibility rules after deprecation window.

### 7.4 Dashboard Migration Requirements (Mandatory)

Schema migration is incomplete until dashboards are migrated.

Required dashboard work:
1. Build a query migration matrix (`old query` -> `new query` -> `dashboard file`).
2. Update dashboards in `monitoring/dashboard/dashboards/` to canonical/derived model.
3. Keep compatibility recording rules active during the migration window.
4. Add dashboard golden checks for critical panels (especially multi-GPU VRAM panels).
5. Remove deprecated query paths only after dashboard verification passes in CI.

Minimum migration matrix entries:
- VRAM used: legacy `logical_name=~".*_vram_used_bytes"` -> canonical `sensor="used",memory_kind="vram"`
- VRAM capacity: legacy `logical_name=~".*_vram_capacity_bytes"` -> canonical `sensor="capacity",memory_kind="vram"`
- VRAM percent: legacy ratio query -> `audiot_gpu_vram_usage_percent` recording rule



