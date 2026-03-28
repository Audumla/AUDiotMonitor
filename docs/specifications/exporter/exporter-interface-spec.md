# Hardware Telemetry Exporter Platform

## Exporter Interface Specification

**Document purpose**: This document defines the code-level interfaces, package boundaries, lifecycle hooks, concurrency model, and implementation contracts for the exporter runtime.

---

# 10. Mapping engine behavior

## 10.6 Auto-Mapping Logic
The engine MUST support dynamic rule inference for unmapped sensors (Auto-Map). 
- **Trigger**: Any measurement that fails to match a manual or built-in rule.
- **Persistence**: Inferred rules SHOULD be persisted to an external YAML file (e.g., `mappings.auto.yaml`) to survive restarts.
- **Priority**: Auto-generated rules MUST have the lowest priority (`1`) to allow easy manual overrides.

---

# 11. Metrics exposition model

## 11.1 `/metrics`
Exposed metric categories:
- **Normalized Metrics**: `hw_device_*` (e.g., `hw_device_temperature_celsius`).
- **Metadata Metrics**: `hw_device_info` (Exposes vendor, model, bios as labels).
- **Self-Metrics**: `hwexp_*` (See Section 11.4).

## 11.4 Self-Metrics
Required implementation:
- `hwexp_adapter_refresh_duration_seconds`: Gauge.
- `hwexp_adapter_refresh_success`: Gauge (1.0 = success).
- `hwexp_adapter_last_success_unixtime`: Gauge.
- `hwexp_discovered_devices`: Gauge.
- `hwexp_mapping_failures_total`: Counter.

---

# 12. HTTP API implementation rules

## 12.1 Routing
Implemented routes:
- `/metrics`: Prometheus metrics.
- `/version`: Version and platform info.
- `/healthz`: Liveness check.
- `/readyz`: Readiness check (Returns 200 OK only after the first poll cycle is complete).
- `/debug/discovery`: JSON view of raw discovered hardware.
- `/debug/catalog`: JSON summary of normalized sensor catalog.
- `/debug/mappings`: JSON view of mapping decisions.
- `/debug/raw`: Raw sensor data (disabled by default).

---

# 15. Extension model

## 15.1 Modular Configuration (`conf.d`)
The exporter MUST support merging configuration from a `conf.d/` directory relative to the main config file. 
- All `.yaml` files in this directory MUST be parsed and merged into the active configuration at startup.

## 15.2 External Scripts (`vendor_exec`)
A specialized adapter MUST support executing external binaries or scripts.
- **Location**: `/etc/hwexp/custom.d/`.
- **Protocol**: Scripts MUST return JSON matching the `RawMeasurement` schema on stdout.
- **Discovery**: When run with `--discover`, scripts MUST return a `DiscoveryResult` JSON.

## 15.3 LLM Monitoring
Built-in support for **Llamaswap** (or OpenAI-compatible APIs) is provided via a dedicated adapter.
- It treats loaded models as virtual devices and tracks their status and performance.



