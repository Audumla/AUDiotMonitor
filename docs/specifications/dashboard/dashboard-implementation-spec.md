# Hardware Telemetry Exporter Platform

## Dashboard Implementation Reference

**Document purpose**: This document is the implementation-level companion to `Dashboard_Data_Contract_Guide.md`. It defines the next level of dashboard detail required for practical Grafana implementation, including canonical PromQL queries, template variable queries, threshold guidance, kiosk/auth provisioning expectations, and the file/folder structure for the dashboard provisioning pack.

**Status**: Normative for dashboard implementation and provisioning

**Audience**: Developers building Grafana dashboards, operators provisioning dashboard nodes, and implementers creating the Pi dashboard deployment pack

---

# 1. Scope

This document defines:

* canonical PromQL query patterns
* recommended query strings for common panels
* Grafana variable query definitions
* threshold and state-color guidance
* kiosk mode provisioning and auth behavior
* dashboard provisioning file/folder layout
* panel implementation guidance for Pi panel, ops, and discovery dashboards

This document does not define final Grafana dashboard JSON exports, but it is intended to be sufficiently detailed to create them directly.

---

# 2. Design principles carried forward

All dashboards MUST:

* use normalized metric families only
* use stable labels only
* avoid raw vendor metric names
* remain portable between embedded Pi mode and central server mode
* tolerate optional metrics being absent
* degrade gracefully if optional components like LLM observer or custom ingest are not present

Dashboard implementations SHOULD:

* prefer `logical_name` for single-sensor selection
* prefer `device_class` + `sensor` for grouped sections
* avoid hard-coding hostnames unless the dashboard is explicitly host-specific

---

# 3. Canonical PromQL conventions

## 3.1 Query design rules

PromQL used in dashboards SHOULD follow these rules:

1. Scope by `host` when a host variable exists.
2. Prefer `logical_name` for exact conceptual sensor selection.
3. Prefer `device_class` and `sensor` when rendering repeated device-class cards.
4. Use `max by (...)`, `avg by (...)`, or `sum by (...)` only where aggregation semantics are intentional.
5. Avoid aggregating away `device_id` on dashboards that need one card per device.
6. Use `rate()`/`increase()` only on counters.
7. Use `clamp_min()` or defensive expressions only when needed to avoid broken visuals.

## 3.2 Canonical host filter fragment

Where a host variable is used, the recommended pattern is:

```promql
{host=~"$host"}
```

If the dashboard is single-host and a fixed host is provisioned, it MAY use:

```promql
{host="sanctum"}
```

## 3.3 Canonical optional metric behavior

If a metric may be absent on some systems, the dashboard SHOULD:

* allow blank/no data state
* avoid hard dependency for dashboard load
* use separate panels or repeated rows only if data exists

---

# 4. Canonical PromQL reference sheet

## 4.1 Host-level queries

### Host uptime

```promql
hw_host_uptime_seconds{host=~"$host"}
```

### CPU package temperature

```promql
max by (host, logical_name) (
  hw_device_temperature_celsius{host=~"$host", device_class="cpu", sensor="package"}
)
```

### RAM used

```promql
hw_device_memory_used_bytes{host=~"$host", device_class="memory"}
```

### RAM total

```promql
hw_device_memory_total_bytes{host=~"$host", device_class="memory"}
```

### RAM usage percent

```promql
100 * (
  hw_device_memory_used_bytes{host=~"$host", device_class="memory"}
/
  hw_device_memory_total_bytes{host=~"$host", device_class="memory"}
)
```

---

## 4.2 GPU queries

### GPU core temperature (one series per device)

```promql
max by (host, device_id, logical_name, model) (
  hw_device_temperature_celsius{host=~"$host", device_class="gpu", sensor="core"}
)
```

### GPU hotspot temperature

```promql
max by (host, device_id, logical_name, model) (
  hw_device_temperature_celsius{host=~"$host", device_class="gpu", sensor="hotspot"}
)
```

### GPU power

```promql
max by (host, device_id, logical_name, model) (
  hw_device_power_watts{host=~"$host", device_class="gpu", sensor="power"}
)
```

If the normalized sensor label instead uses `component="power"`, then the compatible selector is:

```promql
max by (host, device_id, logical_name, model) (
  hw_device_power_watts{host=~"$host", device_class="gpu", component="power"}
)
```

### GPU utilization ratio

```promql
100 * max by (host, device_id, logical_name, model) (
  hw_device_utilization_ratio{host=~"$host", device_class="gpu", sensor="core"}
)
```

### GPU memory used

```promql
max by (host, device_id, logical_name, model) (
  hw_device_memory_used_bytes{host=~"$host", device_class="gpu", component="memory"}
)
```

### GPU memory total

```promql
max by (host, device_id, logical_name, model) (
  hw_device_memory_total_bytes{host=~"$host", device_class="gpu", component="memory"}
)
```

### GPU VRAM usage percent

```promql
100 * (
  max by (host, device_id, logical_name, model) (
    hw_device_memory_used_bytes{host=~"$host", device_class="gpu", component="memory"}
  )
/
  max by (host, device_id, logical_name, model) (
    hw_device_memory_total_bytes{host=~"$host", device_class="gpu", component="memory"}
  )
)
```

### GPU fan speed

```promql
max by (host, device_id, logical_name, model) (
  hw_device_fan_speed_rpm{host=~"$host", device_class="gpu"}
)
```

---

## 4.3 PSU queries

### PSU output power

```promql
max by (host, device_id, logical_name, model) (
  hw_device_power_watts{host=~"$host", device_class="psu", sensor="output"}
)
```

### PSU input power

```promql
max by (host, device_id, logical_name, model) (
  hw_device_power_watts{host=~"$host", device_class="psu", sensor="input"}
)
```

### PSU temperature

```promql
max by (host, device_id, logical_name, model) (
  hw_device_temperature_celsius{host=~"$host", device_class="psu"}
)
```

### PSU fan speed

```promql
max by (host, device_id, logical_name, model) (
  hw_device_fan_speed_rpm{host=~"$host", device_class="psu"}
)
```

### PSU efficiency percent

If directly exposed as ratio:

```promql
100 * max by (host, device_id, logical_name, model) (
  hw_device_utilization_ratio{host=~"$host", device_class="psu", sensor="efficiency"}
)
```

If derived from input/output power:

```promql
100 * (
  max by (host, device_id, logical_name, model) (
    hw_device_power_watts{host=~"$host", device_class="psu", sensor="output"}
  )
/
  max by (host, device_id, logical_name, model) (
    hw_device_power_watts{host=~"$host", device_class="psu", sensor="input"}
  )
)
```

---

## 4.4 Storage and network queries

### Storage temperature

```promql
max by (host, device_id, logical_name, model) (
  hw_device_temperature_celsius{host=~"$host", device_class="storage"}
)
```

### Link speed

```promql
max by (host, device_id, logical_name) (
  hw_device_link_speed_bits_per_second{host=~"$host", device_class="network"}
)
```

### Network throughput

If normalized throughput metrics exist:

```promql
sum by (host) (
  hw_device_bandwidth_bytes_per_second{host=~"$host", device_class="network"}
)
```

---

## 4.5 Exporter health queries

### Adapter failures

```promql
sum by (host, adapter_name) (
  increase(hwexp_adapter_failures_total{host=~"$host"}[15m])
)
```

### Last successful refresh age

```promql
time() - max by (host, adapter_name) (
  hwexp_adapter_last_success_unixtime{host=~"$host"}
)
```

### Mapping failures

```promql
sum by (host) (
  increase(hwexp_mapping_failures_total{host=~"$host"}[15m])
)
```

---

## 4.6 LLM and service queries

These apply only if the optional backend extensions are installed.

### Requests in flight

```promql
sum by (host, logical_name) (
  hw_service_requests_in_flight{host=~"$host", device_class=~"service|llm"}
)
```

### Queue depth

```promql
sum by (host, logical_name) (
  hw_service_queue_depth{host=~"$host", device_class=~"service|llm"}
)
```

### Request duration p95

If histogram support exists later, use histogram_quantile; in v1, if only summary/gauge exists, use the directly normalized family.

Example direct family:

```promql
max by (host, logical_name) (
  hw_service_request_duration_seconds{host=~"$host", device_class=~"service|llm"}
)
```

### Tokens per second

```promql
sum by (host, logical_name) (
  hw_llm_tokens_per_second{host=~"$host", device_class="llm"}
)
```

### Context tokens

```promql
max by (host, logical_name) (
  hw_llm_context_tokens{host=~"$host", device_class="llm"}
)
```

### Active model loaded

If represented as an info/state metric:

```promql
hw_llm_model_loaded{host=~"$host", device_class="llm"}
```

---

# 5. Grafana variable queries

All dashboards SHOULD use Grafana template variables where appropriate.

## 5.1 `host`

Type:

* Query

Query:

```promql
label_values(hw_host_uptime_seconds, host)
```

Fallback if host uptime metric is absent:

```promql
label_values(hw_device_temperature_celsius, host)
```

Multi-value:

* enabled on ops/discovery dashboards
* optional on panel dashboards

Include All option:

* enabled on ops/discovery dashboards
* usually disabled on panel dashboard

---

## 5.2 `device_class`

Type:

* Query

Query:

```promql
label_values(hw_device_temperature_celsius{host=~"$host"}, device_class)
```

Discovery dashboards MAY use a broader metric family set if needed.

---

## 5.3 `logical_name`

Type:

* Query

Recommended query:

```promql
label_values(hw_device_temperature_celsius{host=~"$host"}, logical_name)
```

For generalized dashboards, a broader metric family may be needed if not all logical names emit temperature metrics. In that case use a representative family per dashboard scope.

---

## 5.4 `vendor`

Type:

* Query

Query:

```promql
label_values(hw_device_temperature_celsius{host=~"$host"}, vendor)
```

---

## 5.5 `model`

Type:

* Query

Query:

```promql
label_values(hw_device_temperature_celsius{host=~"$host"}, model)
```

---

## 5.6 `source`

Type:

* Query

Query:

```promql
label_values(hw_device_temperature_celsius{host=~"$host"}, source)
```

This variable SHOULD be used mainly on discovery/debug dashboards.

---

# 6. Thresholds and alert-color guidance

These thresholds define visual consistency for dashboard coloring. They are **dashboard thresholds**, not necessarily alerting rules.

## 6.1 General threshold model

Three states SHOULD be used where applicable:

* Normal
* Warning
* Critical

Color choices are implementation-defined, but SHOULD be semantically consistent across dashboards.

---

## 6.2 GPU temperature thresholds

### Core temperature (°C)

* Normal: `< 75`
* Warning: `75 – < 90`
* Critical: `>= 90`

### Hotspot temperature (°C)

* Normal: `< 90`
* Warning: `90 – < 105`
* Critical: `>= 105`

These are reasonable dashboard defaults for modern GPUs, but MUST remain configurable if operator-specific tuning is needed.

---

## 6.3 CPU package temperature thresholds

### CPU package temperature (°C)

* Normal: `< 75`
* Warning: `75 – < 90`
* Critical: `>= 90`

---

## 6.4 PSU temperature thresholds

### PSU internal temperature (°C)

* Normal: `< 50`
* Warning: `50 – < 65`
* Critical: `>= 65`

Because PSU thermal behavior varies, these SHOULD be treated as conservative defaults.

---

## 6.5 Fan speed thresholds

Fan RPM does not have universal absolute thresholds. Preferred visual behavior:

* no warning/critical by RPM alone unless device-specific thresholds are known
* use low-RPM warning only when fan is expected to be active

If a simple threshold is needed for always-active cooling fans:

* Warning: `< 500 rpm`
* Critical: `< 200 rpm`

These SHOULD be disabled for zero-RPM fan-stop capable devices unless context is available.

---

## 6.6 GPU power thresholds

Absolute GPU power thresholds vary widely by model. Default dashboard guidance:

* do not apply warning/critical purely on watts unless model-specific power envelopes are available
* prefer utilization/temperature color guidance instead

If model-aware limits later exist, thresholds MAY be expressed as percentage-of-expected-max.

---

## 6.7 PSU efficiency thresholds

If efficiency is directly available or derived:

* Normal: `>= 88%`
* Warning: `>= 80% and < 88%`
* Critical: `< 80%`

These are dashboard heuristics and SHOULD be configurable.

---

## 6.8 RAM usage thresholds

RAM usage percentage:

* Normal: `< 80%`
* Warning: `80 – < 92%`
* Critical: `>= 92%`

---

## 6.9 LLM/service thresholds

### Requests in flight

No universal thresholds. Dashboard SHOULD default to neutral display unless operator-configured.

### Queue depth

Suggested generic thresholds:

* Normal: `< 2`
* Warning: `2 – < 5`
* Critical: `>= 5`

### Tokens/sec

Lower tokens/sec is not universally bad, so default threshold coloring SHOULD usually be disabled unless tied to a service-specific SLA.

### Request duration

Operator-defined preferred. Generic dashboard defaults MAY be:

* Normal: `< 2s`
* Warning: `2 – < 10s`
* Critical: `>= 10s`

---

# 7. Panel implementation guidance by dashboard class

## 7.1 Panel dashboard

### Intended rendering model

The panel dashboard SHOULD use:

* Stat panels for headline metrics
* Stat panels with sparkline enabled where lightweight and helpful
* Minimal repeated rows
* Minimal template variable UI exposure

### GPU region

Recommended repeated panel or repeated row grouped by GPU device.

Per GPU card SHOULD include:

* core temp
* hotspot temp if available
* power
* utilization
* VRAM used or VRAM %
* fan RPM if available

### PSU region

Single PSU card SHOULD include:

* output power
* temp
* fan speed
* efficiency if present

### Host region

Single host card SHOULD include:

* CPU temp
* RAM %
* uptime
* total network throughput or another compact host-health metric

### Optional service/LLM region

Only render if matching metrics exist.

---

## 7.2 Operations dashboard

The operations dashboard SHOULD include:

* templated host variable
* repeated or grouped device-class sections
* more time-series panels
* exporter health row
* optional multi-host comparison panels

Useful panel types:

* Time series
* Stat
* Table
* Bar gauge

---

## 7.3 Discovery dashboard

The discovery dashboard SHOULD combine:

* Prometheus-backed views
* JSON/debug endpoint driven views where supported by tooling or companion panels
* operator-oriented tables/lists

At minimum, discovery SHOULD expose:

* discovered device count
* unknown device count
* mapping failure count
* raw vs normalized coverage summary

---

# 8. Kiosk mode provisioning and auth

## 8.1 Browser launch model

For the Pi display, the recommended kiosk launch model is host-side browser startup, not containerized GUI.

Recommended command pattern:

```bash
chromium-browser --kiosk --app="http://127.0.0.1:3000/d/<dashboard_uid>/<slug>?kiosk"
```

Alternative using standard Chromium path on Debian-based Pi systems MAY be:

```bash
/usr/bin/chromium --kiosk --app="http://127.0.0.1:3000/d/<dashboard_uid>/<slug>?kiosk"
```

## 8.2 systemd or autostart provisioning

The deployment pack SHOULD support one of:

* systemd user service
* desktop autostart entry
* lightweight session launch script

Recommended first implementation for Debian Pi desktop sessions:

* a user-level autostart or systemd user service

## 8.3 Grafana authentication for kiosk

Two supported patterns are recommended.

### Pattern A: local anonymous access for kiosk only

Grafana config:

* enable anonymous auth
* restrict anonymous org role to Viewer
* keep Grafana otherwise LAN-reachable with admin auth for configuration

This is the simplest kiosk pattern.

Recommended only on trusted LANs.

### Pattern B: pre-authenticated session/cookie model

More secure, but more operationally complex.
This MAY be deferred.

## 8.4 Recommended v1 choice

For the Pi appliance use case, the recommended v1 model is:

* `GF_AUTH_ANONYMOUS_ENABLED=true`
* anonymous org role = Viewer
* kiosk browser points to localhost Grafana URL

If Grafana is exposed to LAN, operators SHOULD consider whether anonymous viewer access is acceptable for their environment.

## 8.5 API keys / service accounts

Grafana API keys or service accounts are NOT the preferred v1 mechanism for browser kiosk login.
They MAY be used for automation, but SHOULD NOT be the primary browser-auth pattern.

---

# 9. Dashboard Provisioning Pack folder structure

The following folder structure is the normative reference for the provisioning pack.

```text
monitoring/
  collector/                          # deploy on every machine being monitored
    ...
  dashboard/                          # deploy on the Grafana host
    docker-compose.yml                # Includes grafana-init bootstrapper
    config/
      grafana/
        provisioning/
          dashboards/
            dashboards.yaml           # Configured for recursive folder loading
    dashboards/
      profiles/                       # Profile-based organization
        standard/                     # Desktop / 1080p profiles
        wide-screens/                 # Internal panels (1920x440)
        mobile/                       # Portrait / mobile layouts
        debug/                        # Troubleshooting and bring-up
```

## 9.4 Scaffolding and Bootstrapping
The dashboard stack MUST include an initialization service (`grafana-init`) that:
1.  Creates the `profiles/` directory structure on the host if it does not exist.
2.  Populates the directories with default dashboard JSON files from the repository.
3.  Avoids overwriting existing files to ensure user customizations are persisted.

## 9.5 Optional Dashboard Downloads
The initialization service SHOULD support an environment variable `SKIP_DASHBOARD_DOWNLOAD`.
- If set to `true`, the service MUST skip the download phase of dashboard JSON files.
- The service MUST still ensure the directory structure and provisioning configuration files are created.

---

# 10. Concrete provisioning expectations

## 10.2 Profile-Based Providers
Grafana MUST provision dashboard folders automatically based on the directory structure under `/var/lib/grafana/dashboards/profiles/`.
- The provider MUST use `foldersFromFilesStructure: true`.
- Each subfolder (e.g., `standard`, `wide-screens`) MUST appear as a matching folder in the Grafana UI.

## 10.3 Stable dashboard UID guidance

Each provisioned dashboard SHOULD have a stable UID so kiosk URLs remain stable across updates.

Examples:

* `hwexp-panel-main`
* `hwexp-ops-main`
* `hwexp-discovery-main`

---

# 11. Example panel definitions by intent

## 11.1 Panel: GPU core temp stat

* Type: Stat
* Query: GPU core temperature query from section 4.2
* Unit: `celsius`
* Thresholds: GPU core temp thresholds
* Display name: derived from `${__field.labels.logical_name}` or transformed display name

## 11.2 Panel: GPU VRAM usage percent

* Type: Stat or Bar gauge
* Query: GPU VRAM usage percent query
* Unit: `percent (0-100)`
* Thresholds:

  * Normal `< 80`
  * Warning `80–<92`
  * Critical `>=92`

## 11.3 Panel: PSU output power

* Type: Stat
* Query: PSU output power
* Unit: `watt`
* Thresholds: usually neutral unless operator-specific limits are defined

## 11.4 Panel: Exporter mapping failures

* Type: Stat
* Query: mapping failures query
* Unit: none
* Thresholds:

  * Normal `0`
  * Warning `>0`
  * Critical if sustained or rising, implementation-specific

## 11.5 Panel: LLM queue depth

* Type: Stat
* Query: queue depth query
* Unit: none
* Thresholds:

  * Normal `<2`
  * Warning `2–<5`
  * Critical `>=5`

---

# 12. No-data and absent-metric behavior

Dashboards MUST handle absent data gracefully.

Recommended Grafana behavior:

* show `No data` rather than error where possible
* optional panels MAY be hidden by provisioning variants later, but in baseline they may simply render empty
* panel dashboard SHOULD avoid overly noisy error states for absent optional metrics

---

# 13. Implementation/testing checklist

Before dashboards are considered implementation-ready, the following SHOULD be verified:

* host variable populates correctly
* panel dashboard renders on 1920×440 layout without overlap
* GPU panels repeat correctly per discovered GPU
* missing hotspot/fan/PSU efficiency metrics do not break rendering
* exporter health panels resolve and show expected self-metrics
* optional LLM panels stay empty or hidden when no LLM metrics exist
* kiosk URL remains stable across reprovisioning
* anonymous/local viewer auth behavior matches intended kiosk mode

---

# 14. Immediate next implementation artifact

After this document, the next highest-value artifact is a **Dashboard Provisioning Pack** containing actual Grafana provisioning files, starter dashboards, `.env.example`, and kiosk helper scripts for the Pi deployment.



