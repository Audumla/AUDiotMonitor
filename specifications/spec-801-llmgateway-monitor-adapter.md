# Spec 801 — LLM Gateway Monitor Adapter

**Status:** Draft
**Project:** AUDiotMonitor
**Covers:** Generic manifest-driven hwexp adapter for monitoring AUDiaLLMGateway components;
Prometheus metric definitions; Grafana dashboard layout
**Related:** [AUDiaLLMGateway spec-701](../../AUDia/AUDiaLLMGateway/specifications/components/dashboard/spec-701-gateway-dashboard.md) — gateway control panel and manifest schema

---

## 1. Goals

| Goal | Detail |
| --- | --- |
| Replace hardcoded adapter | Replace the bare `llamaswap` adapter (single port, one metric) with a generic adapter driven by component manifests |
| Zero code for new components | Adding a new gateway component to monitor requires only a YAML manifest file — no Go code |
| Prometheus as single data bus | All component health and metric state flows through Prometheus; Grafana and the gateway control panel both read from there |
| GPU correlation | Gateway model state displayed alongside existing hardware metrics (temperature, VRAM) in the same Grafana instance |

---

## 2. Context — Component Manifests

Component manifests are YAML files that live in the AUDiaLLMGateway project at
`config/project/components/<id>.yaml` (project defaults) and
`config/local/components/<id>.yaml` (user overrides).

The hwexp adapter reads only the `health`, `metrics`, and `connection` sections.
The full manifest schema is defined in AUDiaLLMGateway spec-701 §4.2.

The relevant sections for this adapter are:

```yaml
id: llamaswap

health:
  endpoint: /health        # GET this path; 200 = up
  expect_status: 200
  timeout_s: 3

metrics:
  - id: models_loaded
    endpoint: /v1/models
    extract: ".data | length"    # jq expression on JSON response
    prometheus_name: gateway_llamaswap_models_loaded
    unit: count
    poll_interval_s: 15

  - id: active_model
    endpoint: /v1/models
    extract: '.data[0].id // "none"'
    prometheus_name: gateway_llamaswap_active_model
    unit: label                  # emitted as info metric, not a gauge
    poll_interval_s: 15

  - id: vram_ratio
    endpoint: /metrics
    extract: "llama_kv_cache_usage_ratio"
    prometheus_name: gateway_llamaswap_vram_ratio
    unit: ratio
    poll_interval_s: 30
    source_format: prometheus    # json (default) | prometheus

connection:
  host: "127.0.0.1"
  port: 41080
  auth:
    type: none                   # none | bearer | basic
```

---

## 3. New Adapter — `gateway_manifest`

### 3.1 Replace the Per-Service Adapter

The current `llamaswap` adapter at `hwexp/internal/adapters/llamaswap/adapter.go`
is hardcoded to one endpoint and emits only `model_active: 1.0`. It is removed
and replaced by a single generic adapter.

```text
hwexp/internal/adapters/gateway_manifest/
  adapter.go       ← implements Adapter interface; orchestrates discover + poll
  loader.go        ← reads and merges manifest YAML files
  extractor.go     ← jq-style extraction from JSON and Prometheus text responses
  config.go        ← hwexp.yaml config struct for this adapter
```

The old `hwexp/internal/adapters/llamaswap/` directory is deleted.

### 3.2 How It Works

**`Discover(ctx)`**

1. Reads all `*.yaml` files from `manifest_dir`
2. Merges any matching files from `local_manifest_dir` (local fields override project fields)
3. Skips manifests with `enabled: false`
4. Returns one `DiscoveredDevice` per enabled component:

```go
model.DiscoveredDevice{
    StableID:    "gateway-" + manifest.ID,
    Platform:    "software",
    Source:      "gateway_manifest",
    DeviceClass: "gateway_component",
    Vendor:      "audia",
    Model:       manifest.ID,
    DisplayName: manifest.DisplayName,
}
```

**`Poll(ctx)`**

For each discovered component:

1. GET `connection.host:port + health.endpoint`
   - HTTP status matches `health.expect_status` → emit `component_up = 1.0`
   - Timeout or wrong status → emit `component_up = 0.0`

2. For each entry in `metrics[]`:
   - GET `connection.host:port + metric.endpoint`
   - Apply `extract` expression to the response body (see §3.3)
   - Emit `RawMeasurement` with `RawName = metric.prometheus_name`, `RawValue = extracted`

### 3.3 Value Extraction

The `extract` field uses a minimal jq-compatible expression evaluated against
the response body. Two response formats are supported:

**`source_format: json`** (default) — response body is JSON; expression is a jq path:

```
".data | length"           → integer length of .data array
'.data[0].id // "none"'   → first model id, or "none" if empty
".status"                  → string field (emitted as label metric)
```

**`source_format: prometheus`** — response body is Prometheus text format;
`extract` is a metric name to read the first matching value from:

```
"llama_kv_cache_usage_ratio"   → reads the gauge value by name
```

Metrics with `unit: label` are emitted as Prometheus info metrics
(`{metric_name}_info{value="..."} 1`) rather than gauges.

### 3.4 Error Handling

- If a component's health endpoint is unreachable: emit `component_up = 0.0`, skip its metrics
- If a metric endpoint returns non-200: log warning, omit that measurement (no stale value written)
- If extraction fails: log warning with raw response snippet, omit measurement
- Manifests with parse errors at load time: log error, skip that manifest, continue with others

---

## 4. hwexp Configuration

Add to `/etc/hwexp/hwexp.yaml` (or the collector's `config/hwexp/hwexp.yaml`):

```yaml
adapters:
  gateway_manifest:
    enabled: true
    # Directory containing component manifest YAML files.
    # On the same machine as the gateway: symlink or copy of
    #   <gateway-root>/config/project/components/
    manifest_dir: "/etc/hwexp/gateway-components"
    # Optional: user override manifests merged on top of manifest_dir
    local_manifest_dir: "/etc/hwexp/gateway-components-local"
    poll_interval: 15s
    # Optional: override connection.host for all manifests.
    # Useful when monitoring a gateway on a different machine.
    host_override: ""   # empty = use each manifest's connection.host
```

Remove (or disable) the old adapter entry:

```yaml
adapters:
  llamaswap:
    enabled: false   # replaced by gateway_manifest
```

---

## 5. Prometheus Metric Names

The adapter emits `RawMeasurement` objects; the hwexp mapper translates them to
Prometheus metric names using rules in `mappings.yaml`.

Add to `monitoring/collector/config/hwexp/mappings.yaml`:

```yaml
# Component up/down (1 = healthy, 0 = down or unreachable)
- match:
    source: gateway_manifest
    raw_name: component_up
  metric: gateway_component_up
  labels:
    component: "{{device_id}}"

# All other gateway metrics: prometheus_name from manifest passes through directly.
# The manifest author controls the metric name via the prometheus_name field.
- match:
    source: gateway_manifest
    raw_name: "^gateway_"
  metric: "{{raw_name}}"
  labels:
    component: "{{device_id}}"
```

### 5.1 Bundled Metrics Reference

The following metrics are emitted by the default bundled manifests in
AUDiaLLMGateway `config/project/components/`:

| Metric | Type | Labels | Source manifest |
| --- | --- | --- | --- |
| `gateway_component_up` | gauge | `component` | all |
| `gateway_llamaswap_models_loaded` | gauge | `component` | llamaswap |
| `gateway_llamaswap_active_model` | info | `component`, `value` | llamaswap |
| `gateway_llamaswap_vram_ratio` | gauge | `component` | llamaswap |
| `gateway_litellm_requests_total` | counter | `component` | litellm |
| `gateway_litellm_error_rate` | gauge | `component` | litellm |
| `gateway_vllm_gpu_cache_usage` | gauge | `component` | vllm |
| `gateway_vllm_requests_running` | gauge | `component` | vllm |

New components added via user manifests in `config/local/components/` emit whatever
`prometheus_name` values they declare. The mapping rule `^gateway_` catches all of them
automatically — no `mappings.yaml` changes needed for new components.

---

## 6. Grafana Dashboard

New dashboard `AUDia LLM Gateway` provisioned as JSON in
`monitoring/dashboard/dashboards/llmgateway.json`.

Grafana is the **primary display** for all gateway monitoring. Historical trends,
alerting, and GPU correlation all live here. The gateway control panel (`/dashboard/`)
reads current state from Prometheus but defers all visualization to Grafana.

### 6.1 Dashboard Rows

#### Row: Component Health

- Stat panel per component — `gateway_component_up{component="<id>"}` coloured green/red
- Timeline panel — health state history over the last 24 h, one band per component

#### Row: Active Models

- Table panel — `gateway_model_loaded` joined with `gateway_llamaswap_vram_ratio`;
  columns: model name, backend, VRAM ratio, time since loaded
- Stat panel — current value of `gateway_llamaswap_active_model` info label

#### Row: GPU Correlation

- Graph — `hw_device_temperature_celsius{device_class="gpu"}` overlaid with
  `gateway_model_loaded` state-change annotations
- Graph — `gateway_llamaswap_vram_ratio` vs GPU VRAM headroom from
  `hw_device_capacity_bytes{type="vram"}`

#### Row: LiteLLM API *(shown only if litellm metrics are non-zero)*

- Graph — `rate(gateway_litellm_requests_total[5m])`
- Stat — `gateway_litellm_error_rate`

#### Row: vLLM *(hidden when `gateway_component_up{component="vllm"} == 0`)*

- Graph — `gateway_vllm_gpu_cache_usage` over time
- Stat — `gateway_vllm_requests_running`

### 6.2 Dashboard Variables

- `$component` — multi-select, populated from label values of `gateway_component_up`
- Standard Grafana time range picker

### 6.3 Extensibility

New components added via manifests emit metrics with names defined in their own
`prometheus_name` fields. Adding panels for a new component only requires writing
a new Grafana panel referencing those metric names — no adapter changes required.

---

## 7. Manifest Distribution

The manifests live in the AUDiaLLMGateway project and need to be available to
hwexp on the monitoring host. Two options:

**Same machine:** Symlink or bind-mount the gateway component directory:

```bash
ln -s /opt/audia-gateway/config/project/components /etc/hwexp/gateway-components
ln -s /opt/audia-gateway/config/local/components   /etc/hwexp/gateway-components-local
```

**Different machine:** Copy manifests to the monitoring host. Since manifests
only contain health/metric endpoint paths (no secrets), they can be committed
to the AUDiotMonitor repo directly under
`monitoring/collector/config/hwexp/gateway-components/` and deployed alongside
the collector stack.

---

## 8. Delivery Checklist

- [ ] Delete `hwexp/internal/adapters/llamaswap/`
- [ ] Implement `hwexp/internal/adapters/gateway_manifest/adapter.go`
- [ ] Implement `hwexp/internal/adapters/gateway_manifest/loader.go`
- [ ] Implement `hwexp/internal/adapters/gateway_manifest/extractor.go` (JSON + Prometheus text)
- [ ] Implement `hwexp/internal/adapters/gateway_manifest/config.go`
- [ ] Register adapter in `cmd/hwexp/main.go`
- [ ] Add mapping rules to `monitoring/collector/config/hwexp/mappings.yaml`
- [ ] Copy or symlink bundled gateway manifests to collector config
- [ ] Create `monitoring/dashboard/dashboards/llmgateway.json` provisioning file
- [ ] Update `monitoring/collector/config/hwexp/hwexp.yaml` with `gateway_manifest` stanza
- [ ] Smoke test: `python -m py_compile` / `go build ./...`
- [ ] Verify metrics appear in Prometheus at `http://localhost:9090`
- [ ] Verify dashboard loads in Grafana without errors

**Outcome:** All AUDiaLLMGateway component health and model state flows into Prometheus
and is visible in Grafana. No changes to the AUDiaLLMGateway project are required.

---

## 9. Constraints

| Constraint | Detail |
| --- | --- |
| Read-only | This adapter emits metrics only; it has no write path |
| No secrets in manifests | `connection.auth.token_env` names an env var; the token itself is never in the manifest file |
| Additive schema | New manifest fields must have defaults so existing manifests remain valid without changes |
| No hwexp core changes | The adapter uses the existing `Adapter` interface; no engine or mapper changes needed |
