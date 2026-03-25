# Spec 100 — Component Monitor Architecture

**Status:** Draft
**Project:** AUDiotMonitor
**Covers:** Componentization strategy; manifest-driven adapter pattern; cross-project
component registration; Docker image separation; optional runtime dependency build path
**Related:** [spec-801](spec-801-llmgateway-monitor-adapter.md) — LLM gateway adapter (reference implementation)

---

## 1. Problem Statement

The current system is partially monolithic:

| Problem | Impact |
| --- | --- |
| `llamaswap` adapter is hardcoded | Adding a new LLM service requires Go code in this repo |
| No standard way for external projects to contribute monitoring | Every new service needs a change here |
| `wl-gammarelay` must be compiled from source on each target device | Requires Rust toolchain post-deployment; fragile |
| Single Docker image per tier | Kiosk display tools (wl-gammarelay, swayidle) mixed with Grafana concerns |
| Kiosk launched natively; no reproducible build | Display environment differs per device |

**Goal:** Any service — in this repo or a separate project — can be monitored by
dropping a YAML manifest file. No code changes to AUDiotMonitor are required.
Optional runtime tools are built in CI and delivered as Docker image layers or
release artifacts, never compiled on the target device.

---

## 2. Component Manifest Pattern

Defined concretely in [spec-801](spec-801-llmgateway-monitor-adapter.md). Summary:

A **component manifest** is a YAML file that describes how to:
1. Check a service is alive (`health` stanza — HTTP GET, expected status)
2. Collect metrics (`metrics[]` stanza — HTTP GET + jq/Prometheus extraction)
3. Reach the service (`connection` stanza — host, port, optional auth)

The `gateway_manifest` hwexp adapter reads manifests from one or more directories
and polls each service. Zero Go code is needed for new components.

### 2.1 Generalisation Beyond the LLM Gateway

Spec-801 names the adapter `gateway_manifest` and targets LLM backends. The
underlying mechanism is generic: **any HTTP service** that exposes a health
endpoint and at least one JSON or Prometheus-format metrics endpoint can be
described in a manifest.

The refactor renames the adapter `component_manifest` with full backward
compatibility (the `gateway_manifest` name remains as an alias). The manifest
schema is unchanged.

### 2.2 Manifest Directory Convention

```
/etc/hwexp/components/            ← standard install location
  <project>-<component>.yaml      ← e.g. llamaswap-server.yaml, litellm-proxy.yaml

/etc/hwexp/components-local/      ← user overrides (higher priority, git-ignored)
  <component>.yaml
```

External projects install their manifests here via:
- **Same machine, bind mount:** Docker bind-mount `<project>/config/components` → `/etc/hwexp/components/<project>/`
- **Same machine, symlink:** `ln -s /opt/myproject/config/components /etc/hwexp/components/myproject`
- **Remote machine:** Copy manifests to `monitoring/collector/config/hwexp/components/<project>/` in this repo and redeploy the collector stack

Neither approach requires changes to AUDiotMonitor code.

### 2.3 Metric Namespace

Manifest authors own their metric names via `prometheus_name`. The only rule:
all component metrics must begin with a project prefix (e.g. `gateway_`, `llama_`,
`audia_`). The mapping rule `^<prefix>_` passes them through automatically.

---

## 3. Docker Image Separation

### 3.1 Current Images

| Image | Tier | Contents |
| --- | --- | --- |
| `audumla/audiot-hwexp` | Collector | hwexp binary + default configs |
| `audumla/audiot-dashboard` | Dashboard | Grafana + provisioning + kiosk patches |

### 3.2 Target Images

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
device gets a pre-built binary with no compiler required.

### 3.3 wl-gammarelay Build (kiosk-tools image)

New file: `monitoring/dashboard/kiosk/Dockerfile`

```dockerfile
# ── Stage 1: build wl-gammarelay ──────────────────────────────────────────────
FROM rust:1-slim-bookworm AS wl-gammarelay-builder
RUN apt-get update && apt-get install -y --no-install-recommends \
        libdbus-1-dev pkg-config \
    && rm -rf /var/lib/apt/lists/*
RUN cargo install wl-gammarelay --locked

# ── Stage 2: minimal runtime image ───────────────────────────────────────────
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
        libdbus-1-3 \
    && rm -rf /var/lib/apt/lists/*
COPY --from=wl-gammarelay-builder /usr/local/cargo/bin/wl-gammarelay /usr/local/bin/
# Additional kiosk host tools can be added here in future stages.
CMD ["/usr/local/bin/wl-gammarelay"]
```

CI publishes `audumla/audiot-kiosk-tools:latest` (and versioned tags) on every
merge to `main`. The DietPi bootstrap installer pulls this image to extract binaries.

---

## 4. Refactor Scope (branch: `refactor/component-architecture`)

### Phase A — Manifest Adapter (hwexp)

| Task | File(s) |
| --- | --- |
| Rename `gateway_manifest` → `component_manifest` (alias kept) | `hwexp/internal/adapters/component_manifest/` |
| Implement adapter, loader, extractor, config | spec-801 §3 |
| Register adapter in main | `cmd/hwexp/main.go` |
| Add mapping rules | `monitoring/collector/config/hwexp/mappings.yaml` |
| Update hwexp.yaml | `monitoring/collector/config/hwexp/hwexp.yaml` |
| Disable old `llamaswap` adapter | Same |

### Phase B — Manifest Distribution

| Task | File(s) |
| --- | --- |
| Add `components/` directory to collector config | `monitoring/collector/config/hwexp/components/` |
| Add llamaswap manifest (reference implementation) | `monitoring/collector/config/hwexp/components/llamaswap.yaml` |
| Document bind-mount pattern for same-machine projects | README + spec-801 §7 |

### Phase C — Docker Build (kiosk-tools)

| Task | File(s) |
| --- | --- |
| Create kiosk-tools Dockerfile | `monitoring/dashboard/kiosk/Dockerfile` |
| Add CI workflow for kiosk-tools image | `.github/workflows/kiosk-tools.yml` (or equivalent) |
| Update DietPi bootstrap to extract binary | `monitoring/dashboard/setup/dietpi-setup.sh` |
| Update kiosk.sh comment (wl-gammarelay installed via kiosk-tools) | `monitoring/dashboard/kiosk.sh` |

### Phase D — Grafana Dashboard

| Task | File(s) |
| --- | --- |
| Create LLM Gateway dashboard | `monitoring/dashboard/dashboards/llmgateway.json` |
| Add provisioning entry | `monitoring/dashboard/config/grafana/` |

---

## 5. External Project Integration — llama-swap Example

The user's local llama-swap instance at `h:/development/tools/llama-swap` can be
monitored by:

1. Creating `monitoring/collector/config/hwexp/components/llamaswap.yaml` in this
   repo with the appropriate `connection.host`, `health`, and `metrics` stanzas.

2. On the monitoring host, the collector Docker container bind-mounts this directory
   to `/etc/hwexp/components/` — no changes to AUDiotMonitor code needed.

3. Grafana panels reference `gateway_llamaswap_*` (or whatever `prometheus_name`
   values the manifest declares) — no adapter changes needed.

When llama-swap adds new endpoints, only the manifest YAML changes. New components
(litellm, vLLM, etc.) require only a new YAML file.

---

## 6. Delivery Order

```
Phase A  →  Phase B  →  Phase C  →  Phase D
  (adapter)   (manifests)  (docker)    (grafana)
```

Phases A and C are independent and can be developed in parallel.
Phase B requires Phase A. Phase D requires Phase B.

---

## 7. Constraints

| Constraint | Detail |
| --- | --- |
| No core hwexp changes | component_manifest uses existing Adapter interface |
| Backward compatible | llamaswap adapter remains available during transition; disabled by config not deleted |
| No secrets in manifests | Token env var names only; never actual values |
| arm64 support | kiosk-tools image built for linux/arm64 (RPi target) and linux/amd64 |
| Idempotent bootstrap | Re-running DietPi setup does not break existing install |
