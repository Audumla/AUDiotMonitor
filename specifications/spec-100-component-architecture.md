# Spec 100 — Component Monitor Architecture

**Status:** Draft
**Project:** AUDiotMonitor
**Covers:** System boundary definition; manifest-driven software component monitoring;
external project data integration via Prometheus; Docker image separation; CI build path
for optional runtime tools
**Related:** [spec-801](spec-801-llmgateway-monitor-adapter.md) — LLM gateway manifest adapter (reference implementation)

---

## 1. System Boundary

AUDiotMonitor has two distinct concerns that must not blur:

| Concern | Owner | Extension point |
| --- | --- | --- |
| **Hardware metrics** — temperatures, GPU state, VRAM, fans, etc. | This repo | New hwexp adapters added here |
| **External service data** — LLM model state, application health, any non-hardware metric | External project | Prometheus `/metrics` endpoint; this repo only adds a scrape config |

**Hardware adapters are always internal.** Any new hardware source (new GPU vendor,
new sensor type, new platform) gets a Go adapter in `hwexp/internal/adapters/`.
That work happens in this repo; there is no external adapter plugin system.

**External project data is always decoupled.** An external project exposes a
standard Prometheus `/metrics` endpoint. AUDiotMonitor's Prometheus scrapes it.
No hwexp code changes are ever required for external data.

---

## 2. How External Projects Expose Data

An external project that wants its metrics in Grafana:

1. Runs a Prometheus exporter on a known port (any language, any framework)
2. Adds a scrape entry to the collector's `prometheus.yml`

```yaml
# monitoring/collector/config/prometheus/prometheus.yml
scrape_configs:
  - job_name: 'llama-swap'
    static_configs:
      - targets: ['host.docker.internal:9400']   # or the actual host:port

  - job_name: 'my-other-service'
    static_configs:
      - targets: ['host.docker.internal:9401']
```

That is the entire integration. No hwexp changes. No manifest files. No adapter
code. Grafana queries the metrics by name directly.

This is the **primary integration path for external projects.**

---

## 3. Component Manifests — Software Services Correlated With Hardware

The component manifest adapter (spec-801) serves a narrower purpose than general
external data integration: it is for **software services that need to be displayed
alongside and correlated with hardware metrics** in the same Grafana panels.

Examples:

- LLM model loaded → shown next to GPU temperature and VRAM usage
- Inference backend health → shown in the same dashboard row as GPU utilisation

For this use case, hwexp polls the service via a manifest YAML file and emits
measurements through its normal normalization pipeline. This allows the same
device/label model to be applied and enables Grafana queries that join hardware
and software metrics using shared labels.

A manifest is the right tool when:

- The service does not already expose Prometheus metrics, OR
- You need the data correlated with a specific hardware device via hwexp labels

A Prometheus scrape is the right tool when:

- The service already exports Prometheus metrics, OR
- The data does not need to be joined with hardware device labels

### 3.1 Manifest Directory

```text
monitoring/collector/config/hwexp/components/
  llamaswap.yaml        ← llama-swap reference manifest
  <service>.yaml        ← add one file per service to correlate with hardware
```

These files are part of this repo. Any service that needs hardware correlation
gets a manifest added here. Services that only need visibility in Grafana use
the Prometheus scrape path instead.

---

## 4. Docker Image Separation

### 4.1 Current Images

| Image | Tier | Contents |
| --- | --- | --- |
| `audumla/audiot-hwexp` | Collector | hwexp binary + default configs |
| `audumla/audiot-dashboard` | Dashboard | Grafana + provisioning + kiosk patches |

### 4.2 Target Images

| Image | Tier | Contents | Change |
| --- | --- | --- | --- |
| `audumla/audiot-hwexp` | Collector | hwexp binary + component_manifest adapter | Updated |
| `audumla/audiot-dashboard` | Dashboard | Grafana + provisioning + kiosk patches | Unchanged |
| `audumla/audiot-kiosk-tools` | Kiosk host | wl-gammarelay + helper binaries | **New** |

`audiot-kiosk-tools` is a thin image used solely as a source of pre-built binaries.
The DietPi bootstrap script extracts binaries from it at setup time:

```bash
docker run --rm audumla/audiot-kiosk-tools \
    cat /usr/local/bin/wl-gammarelay > /usr/local/bin/wl-gammarelay
chmod +x /usr/local/bin/wl-gammarelay
```

This keeps the Rust toolchain and build environment entirely in CI. The target
device receives a pre-built binary with no compiler required.

### 4.3 wl-gammarelay Build (kiosk-tools image)

Dockerfile at `monitoring/dashboard/kiosk/Dockerfile` — multi-stage Rust build,
thin Debian runtime. CI publishes `audumla/audiot-kiosk-tools:latest` (and
versioned tags) on every merge to `main`.

---

## 5. Refactor Scope (branch: `refactor/component-architecture`)

### Phase A — Component Manifest Adapter (hwexp)

Implements spec-801. Replaces the hardcoded `llamaswap` Go adapter with a
manifest-driven adapter that polls any HTTP service described in a YAML file.

| Task | File(s) |
| --- | --- |
| Implement `component_manifest` adapter | `hwexp/internal/adapters/component_manifest/` |
| Loader, extractor (JSON + Prometheus text), config | Same package |
| Register adapter in main | `cmd/hwexp/main.go` |
| Add mapping rules | `monitoring/collector/config/hwexp/mappings.yaml` |
| Update hwexp.yaml | `monitoring/collector/config/hwexp/hwexp.yaml` |
| Disable old `llamaswap` adapter (config flag, not deleted) | Same |

### Phase B — Manifest Distribution

| Task | File(s) |
| --- | --- |
| llamaswap reference manifest | `monitoring/collector/config/hwexp/components/llamaswap.yaml` ✓ |
| Bind-mount in collector Docker compose | `monitoring/collector/docker-compose.yml` |

### Phase C — Docker Build (kiosk-tools)

| Task | File(s) |
| --- | --- |
| kiosk-tools Dockerfile | `monitoring/dashboard/kiosk/Dockerfile` ✓ |
| CI workflow | `.github/workflows/kiosk-tools.yml` |
| DietPi bootstrap extracts binary | `monitoring/dashboard/setup/dietpi-setup.sh` |

### Phase D — Grafana Dashboard

| Task | File(s) |
| --- | --- |
| LLM Gateway dashboard | `monitoring/dashboard/dashboards/llmgateway.json` |
| Provisioning entry | `monitoring/dashboard/config/grafana/` |

---

## 6. Delivery Order

```text
Phase A  →  Phase B  →  Phase D
  (adapter)   (manifests)  (grafana)

Phase C  (independent — kiosk-tools docker build)
```

---

## 7. Constraints

| Constraint | Detail |
| --- | --- |
| Hardware adapters stay internal | No external plugin system for hwexp adapters |
| External data via Prometheus | External projects own their exporters; this repo adds only a scrape entry |
| No core hwexp changes | component_manifest uses existing Adapter interface |
| Backward compatible | Old llamaswap adapter disabled by config, not deleted |
| No secrets in manifests | Token env var name only; actual value never in the manifest |
| arm64 + amd64 | kiosk-tools image multi-platform build |
| Idempotent bootstrap | Re-running DietPi setup does not break existing install |
