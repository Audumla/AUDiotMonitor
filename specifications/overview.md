# AUDiot (hwexp) Specification

This document outlines the architecture, taxonomy, and implementation plan for the AUDiot hardware telemetry platform.

---

## Table of Contents

1. [Purpose](#1-purpose)
2. [Scope](#2-scope)
3. [System Roles](#3-system-roles)
4. [Supported Deployment Modes](#4-supported-deployment-modes)
5. [Recommended v1 Architecture](#5-recommended-v1-architecture)
6. [Build vs Buy Strategy](#6-build-vs-buy-strategy)
7. [Product Decomposition](#7-product-decomposition)
8. [Design Principles](#8-design-principles)
9. [Technology Decisions](#9-technology-decisions)
10. [External Interfaces](#10-external-interfaces)
11. [Exporter Runtime Behavior](#11-exporter-runtime-behavior)
12. [Discovery Model](#12-discovery-model)
13. [Taxonomy](#13-taxonomy)
14. [Metric Naming Standard](#14-metric-naming-standard)
15. [Label Schema](#15-label-schema)
16. [Mapping and Normalization Pipeline](#16-mapping-and-normalization-pipeline)
17. [Adapter Framework](#17-adapter-framework)
18. [Platform Adapters](#18-platform-adapters)
19. [Source Precedence Rules](#19-source-precedence-rules)
20. [Configuration Schema](#20-configuration-schema)
21. [Dashboard Platform Spec](#21-dashboard-platform-spec)
22. [Deployment Specs](#22-deployment-specs)
23. [Packaging and Portability](#23-packaging-and-portability)
24. [Security Model](#24-security-model)
25. [Test Harness Specification](#25-test-harness-specification)
26. [CI/CD Specification](#26-cicd-specification)
27. [Versioning and Compatibility](#27-versioning-and-compatibility)
28. [Observability of the Exporter Itself](#28-observability-of-the-exporter-itself)
29. [V1 Implementation Phases](#29-v1-implementation-phases)
30. [Repository Layout](#30-repository-layout)
31. [Concrete First Deliverables](#31-concrete-first-deliverables)
32. [Final Recommendation](#32-final-recommendation)

---

## 1. Purpose

Build a portable hardware telemetry platform with these outcomes:
- **Source machines** expose normalized hardware telemetry over HTTP.
- A **central dashboard stack** scrapes and stores that telemetry.
- **Dashboards** can be viewed from:
  - The Raspberry Pi screen in kiosk mode.
  - Desktop browsers.
  - Laptops.
  - Other LAN viewers.
- The exporter is designed to support:
  - Current Linux hardware.
  - Future unknown hardware.
  - Future Windows and macOS sources.
- The naming, classification, and mapping remain stable even when raw hardware sources differ.

## 2. Scope

### In Scope
- Linux-first hardware exporter.
- Normalized taxonomy and naming model.
- Mapping rules from raw sensor names to stable logical names.
- Test harness with fixtures and golden outputs.
- Two supported deployment modes:
  - Multi-host monitoring.
  - Pi display support.
- Support for existing exporters where appropriate.
- Future adapter hooks for Windows and macOS.

### Out of Scope for v1
- Cloud services / SaaS dependencies.
- Vendor lock-in.
- Remote write pipelines.
- Mobile apps.
- Alerting workflows beyond basic health checks.
- Custom dashboard editor UI.

## 3. System Roles

There are three primary product roles.

### 3.1 Source Host
A monitored machine that owns the real sensors.
- **Examples:** Linux workstation with GPUs/PSU, future Windows gaming PC, future macOS machine.
- **Responsibilities:**
  - Run standard exporters where useful.
  - Run custom normalization exporter.
  - Expose metrics locally on HTTP.

### 3.2 Dashboard Node
The machine that runs the backend stack (Prometheus and Grafana).
- **Responsibilities:**
  - Scrape source hosts.
  - Store time series.
  - Serve dashboards.
  - Support one or many viewers.

### 3.3 Viewer
Any browser client used to access the dashboards.
- **Examples:** Raspberry Pi screen (kiosk), desktop/laptop browsers, tablets.
- **Note:** Grafana dashboards are shared and viewed from multiple browsers. Kiosk mode hides the chrome for dedicated displays.

## 4. Supported Deployment Modes

The system must support both modes below with the same exporter schema, dashboard definitions, and taxonomy.

### 4.1 Mode A: Embedded Pi Server Mode
The Raspberry Pi runs the entire stack: Prometheus, Grafana, and the kiosk browser. Source machines run exporters only.
- **Use Case:** Small self-contained appliance, low host/viewer count, modest dashboards.
- **Pros:** Simple, local, easy first deployment.
- **Cons:** Pi handles both server and display tasks; less headroom for growth.

### 4.2 Mode B: Central Server Mode
A non-Pi machine runs Prometheus and Grafana. The Raspberry Pi runs browser-only kiosk session.
- **Use Case:** Multiple viewers, rich dashboards, long-term retention, high host count.
- **Pros:** Better performance, Pi is only a viewer, easier to scale.
- **Cons:** One extra machine involved.

### 4.3 Requirement
The platform must be deployment-location independent:
- Moving Prometheus/Grafana must not require exporter redesign.
- Moving the Pi from server mode to viewer-only mode must not require dashboard redesign.

## 5. Recommended v1 Architecture

For v1, support both modes, but design around this logical model:

```text
[Source Host Exporters] --> [Prometheus] --> [Grafana] --> [Viewers]
                                              |--> Pi kiosk
                                              |--> Desktop browser
                                              |--> Laptop browser
```

Prometheus scrapes HTTP metrics endpoints from targets.

## 6. Build vs Buy Strategy

Do not build every collector from scratch.

### Use existing components for:
- **Generic host metrics:** `node_exporter`
- **AMD GPU metrics:** AMD SMI exporter or direct AMDSMI integration.
- **NVIDIA GPU metrics:** DCGM exporter where needed.
- **Dashboarding:** Grafana.
- **Storage/Query:** Prometheus.

### Build custom components for:
- Discovery across heterogeneous hardware.
- Normalized naming and taxonomy.
- Mapping raw sensor names to stable logical names.
- Stitching data from multiple sources into one coherent catalog.
- Future cross-platform adapter framework.
- Test harness and fixture capture.

## 7. Product Decomposition

Build two logical products.

### 7.1 Product A: Hardware Exporter Platform
Runs on source machines.
- **Responsibilities:** Discover devices, poll sensors, normalize metrics, classify devices/sensors, expose `/metrics`, provide debug endpoints, support mapping overrides, provide test fixture capture.

### 7.2 Product B: Dashboard Platform
Runs on the dashboard node.
- **Responsibilities:** Scrape exporters, retain data, render dashboards, serve multiple viewers, support kiosk and full desktop layouts.

## 8. Design Principles

- **Local-first:** No cloud required.
- **Schema-first:** Dashboard stability depends on metric stability.
- **Unknown-hardware tolerant:** Do not drop unknown sensors; map them as "unknown".
- **Cross-platform by interface:** Linux first, but model must be extensible to Windows/macOS.
- **Low-cardinality discipline:** Labels must remain queryable at scale.
- **Partial failure tolerance:** One broken adapter must not collapse all metrics.
- **Debuggable normalization:** Every mapping decision must be inspectable.

## 9. Technology Decisions

### 9.1 Exporter Language: Go
- **Reason:** Strong fit for exporter daemons, first-class Prometheus client support, easy static binaries, good cross-platform packaging.

### 9.2 Data Format: Prometheus exposition format
- Use `/metrics` endpoint and follow OpenMetrics-friendly naming.

### 9.3 Dashboard Stack: Prometheus + Grafana
- Grafana is used for visualization and supports kiosk mode natively.

## 10. External Interfaces

### 10.1 Exporter HTTP API
Required endpoints:
- `/metrics`: Prometheus scrape endpoint.
- `/healthz`: Overall process liveness.
- `/readyz`: Readiness for scraping (fails if config is invalid or pipeline is not ready).
- `/version`: Version information.
- `/debug/discovery`: JSON describing discovered devices and source metadata.
- `/debug/mappings`: JSON describing how raw measurements are mapped to logical metrics.
- `/debug/catalog`: JSON describing the normalized sensor catalog.
- `/debug/raw`: (Optional) Raw sensor data, off by default.

## 11. Exporter Runtime Behavior

### 11.1 Refresh Model
The exporter maintains an internal loop:
1. Discover devices.
2. Poll adapters.
3. Normalize values.
4. Replace current metric snapshot atomically.
*This ensures predictable scrape performance.*

### 11.2 Intervals
- **Refresh Interval:** Default 5s (configurable).
- **Discovery Interval:** Default 60s (configurable).

### 11.3 Failure Behavior
- If an adapter fails, keep previous successful values for a grace period.
- Expose adapter failure metrics and log reasons.

### 11.4 Self-Metrics
Must expose:
- `hwexp_adapter_refresh_duration_seconds`
- `hwexp_adapter_refresh_success`
- `hwexp_adapter_last_success_unixtime`
- `hwexp_discovered_devices`
- `hwexp_mapping_failures_total`
- `hwexp_scrape_requests_total`
- `hwexp_config_reload_success`

## 12. Discovery Model

Discovery is independent of polling.

### 12.1 Discovery Outputs
Fields for each device: `stable_id`, `platform`, `source`, `device_class`, `device_subclass`, `vendor`, `model`, `raw_identifiers`, `driver`, `bus`, `location`, `capabilities`, `display_name` (optional).

### 12.2 Stable Identity Rules
Order of preference:
1. Vendor/device serial (if stable).
2. PCI address.
3. USB stable path + VID/PID.
4. Canonical OS identifier.
5. Fallback synthesized ID (cached locally).

### 12.3 Unknown Devices
Emitted with `device_class="unknown"`, raw metadata, and minimal safe labels.

## 13. Taxonomy

Canonical classification model.

### 13.1 Top-level Device Classes
`host`, `cpu`, `gpu`, `psu`, `memory`, `motherboard`, `fan`, `pump`, `storage`, `network`, `usb`, `ups`, `battery`, `display`, `sensor_hub`, `unknown`.

### 13.2 Components
`core`, `memory`, `vrm`, `hotspot`, `thermal`, `fan`, `rail`, `input`, `output`, `pcie`, `controller`, `link`, `smart`, `wear`.

### 13.3 Sensor Types
`temperature`, `power`, `voltage`, `current`, `fan_speed`, `utilization`, `capacity`, `usage`, `throughput`, `link_speed`, `energy`, `status`, `uptime`.

### 13.4 Friendly Logical Naming Layer
Used for config and dashboards (e.g., `gpu0_core_temp`, `psu_output_power`).

## 14. Metric Naming Standard

Follow Prometheus guidance: `snake_case`, include units, use labels for dimensions.

### 14.1 Base Metric Families
Examples: `hw_device_temperature_celsius`, `hw_device_power_watts`, `hw_device_fan_speed_rpm`, `hw_device_utilization_ratio`, `hw_device_status`.

### 14.2 Host-level Families
Examples: `hw_host_uptime_seconds`, `hw_host_power_estimate_watts`.

## 15. Label Schema

### 15.1 Required Labels
`host`, `platform`, `source`, `device_class`, `device_id`, `logical_name`, `sensor`.

### 15.2 Optional/Strongly Recommended
`vendor`, `model`, `component`, `driver`, `bus`, `gpu`, `fan`, `rail`.

### 15.3 Forbidden Patterns
No raw timestamps, unbounded text, or labels starting with `__`.

## 16. Mapping and Normalization Pipeline

Pipeline: **Discover Source Object** -> **Classify Device** -> **Classify Sensor** -> **Match Rule** -> **Convert Units** -> **Assign Logical Name** -> **Emit Metric**.

### 16.1 Mapping Rule Priorities
1. Explicit user override.
2. Device-specific built-in rule.
3. Vendor/model built-in rule.
4. Class-level fallback.
5. Unknown fallback.

### 16.2 Example Rule (YAML)
```yaml
mappings:
  - id: corsair_psu_temp
    match:
      platform: linux
      source: linux_hwmon
      vendor: corsair
      model_regex: "HX1500i"
      raw_name_regex: "temp([0-9]+)_input"
    normalize:
      metric: hw_device_temperature_celsius
      device_class: psu
      component: thermal
      sensor_template: "internal_${1}"
      logical_name_template: "psu_internal_temp_${1}"
      unit_scale: 0.001
```

## 17. Adapter Framework

### 17.1 DiscoveryAdapter
```text
Discover(ctx) -> []DiscoveredDevice
```

### 17.2 PollAdapter
```text
Poll(ctx, device) -> []RawMeasurement
```

### 17.3 BridgeAdapter
```text
ScrapeAndBridge(ctx, endpoint) -> []RawMeasurement
```

## 18. Platform Adapters

### 18.1 Linux (v1)
- `linux_hwmon`
- `linux_vendor_exec`
- `linux_gpu_vendor`
- `linux_node_bridge`

### 18.2 Windows/macOS (v2 Target)
- `windows_wmi`
- `windows_pdh`
- `darwin_iokit`
- `darwin_smc`

## 19. Source Precedence Rules
1. Native vendor/library source (Highest).
2. OS-native structured source.
3. Specialized exporter bridge.
4. Generic exporter bridge.
5. CLI-derived source (Lowest).

## 20. Configuration Schema

### 20.1 Main Config Example
```yaml
server:
  listen_address: "0.0.0.0:9200"
  refresh_interval: 5s
identity:
  host: sanctum
  platform: linux
adapters:
  linux_hwmon:
    enabled: true
mapping:
  rules_file: /etc/hwexp/mappings.yaml
```

## 21. Dashboard Platform Spec

### 21.1 Prometheus
Responsible for scraping targets and retaining time series.

### 21.2 Grafana
Responsible for rendering dashboards in full or kiosk mode.

### 21.3 Required Layouts
- **A. Panel Dashboard:** Optimized for 1920×440 Pi display.
- **B. Operations Dashboard:** For full desktop browsers.
- **C. Discovery Dashboard:** For hardware bring-up and debugging.

## 22. Deployment Specs
*For detailed installation flows and profiles, refer to `installation_spec.md`.*

### 22.1 Mode A (Embedded)
Pi runs everything. Suitable for small setups.

### 22.2 Mode B (Central)
Central server runs the stack; Pi is a kiosk viewer. Preferred for growth.

### 22.3 Migration
Moving from A to B must only require updating target config and URLs; no exporter changes.

## 23. Packaging and Portability
*For detailed packaging and installation specifics, including component binaries, service definitions, and deployment profiles, refer to `installation_spec.md`.*

- **Exporter:** Linux tarball, `.deb`, `.rpm`, `systemd` units. (Future: Windows MSI, macOS PKG).
- **Dashboard:** Native install or containerized (Docker).

## 24. Security Model
- **Network:** Bind to LAN/Host only; no default internet exposure.
- **Privilege:** Minimum privileges; document any need for root.
- **Data:** Do not expose sensitive info (full serials, user metadata) by default.

## 25. Test Harness Specification (Mandatory)

### 25.1 Layers
- **Unit Tests:** Logic, matching, scaling.
- **Fixture Tests:** Replay captured hardware snapshots (hwmon, etc.).
- **Golden Tests:** Assert exact normalized outputs.
- **Contract Tests:** Validate schema, names, and labels.
- **Integration Tests:** Scrape real/test instances.

### 25.2 Fixture Capture Utility
`hwexp-fixture-capture` tool to record real hardware states for replay in CI.

## 26. CI/CD Specification
- **Build Matrix:** Linux (amd64/arm64), Windows (amd64), macOS (amd64/arm64).
- **Stages:** Lint -> Unit -> Fixture Replay -> Golden -> Contract -> Integration -> Package.

## 27. Versioning and Compatibility
Maintain versions for exporter, metric schema, and mapping rules. Increment version for any breaking change.

## 28. Observability of the Exporter Itself
Monitor: Up/Down status, adapter failures, refresh duration, discovery changes, mapping failures.

## 29. V1 Implementation Phases
- **Phase 0:** Spec Lock.
- **Phase 1:** Linux Core MVP (Go daemon, hwmon, mapping engine).
- **Phase 2:** GPU Integration.
- **Phase 3:** Dashboard Platform (Prometheus/Grafana).
- **Phase 4:** Bridge and Future Hooks.
- **Phase 5:** Packaging and Migration. (Detailed in `installation_spec.md`)

## 30. Repository Layout
```text
hwexp/
  cmd/
    hwexp/
    hwexp-fixture-capture/
  internal/
    core/
    model/
    taxonomy/
    mapper/
    adapters/
      linux/
      windows/
      darwin/
    bridges/
    httpapi/
    selfmetrics/
    config/
  pkg/
    schema/
  configs/
    examples/
  dashboards/
    grafana/
      panel/
      ops/
      discovery/
  deploy/
    pi-embedded/
    central-server/
  packaging/
    systemd/
    deb/
    rpm/
  tests/
    fixtures/
    golden/
    contracts/
    integration/
```

## 31. Concrete First Deliverables
- ADR (Architecture Decision Record).
- Metric Schema Document.
- Mapping Rules Schema.
- Fixture Capture Tool.
- Linux MVP Exporter.
- Dashboard Starter Pack.

## 32. Final Recommendation

The design of a custom normalized exporter combined with a standard Prometheus/Grafana stack is highly sensible. It provides a stable "telemetry language" that abstracts away hardware differences, ensuring that dashboards remain functional even as source hardware or operating systems evolve. The support for both embedded and central deployment modes offers a clear path from a single-machine hobbyist setup to a multi-node professional monitoring solution.