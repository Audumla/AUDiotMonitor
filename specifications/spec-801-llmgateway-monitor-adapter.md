# Spec 801 — LLM Gateway Manifest Adapter

**Status:** Finalized
**Date:** 2026-03-26
**Project:** AUDiotMonitor
**Covers:** Manifest-driven hwexp adapter for monitoring AUDiaLLMGateway components; Two-tier discovery; Hardware correlation; Hot reloading.
**Related:** [Spec 100](spec-100-component-architecture.md), AUDiaLLMGateway [spec-701](../../AUDia/AUDiaLLMGateway/specifications/components/dashboard/spec-701-gateway-dashboard.md)

---

## 1. Goal
Provide a generic, manifest-driven adapter that enables zero-code monitoring of software components while allowing deep correlation with physical hardware (GPUs).

---

## 2. Manifest Discovery & Lifecycle

### 2.1 Search Paths
The adapter MUST search for `.yaml` manifests in:
1.  **Project Path:** `/etc/hwexp/components/`
2.  **Local Overrides:** `/etc/hwexp/local/components/`

### 2.2 Merge Logic
- **Shallow Merge:** Top-level keys in local manifests override project manifests.
- **List Merging:** Items in `metrics` and `actions` lists are merged by their `id`.
- **Hot Reload:** The adapter MUST re-scan these directories every **15 seconds**.

### 2.3 Variable Resolution
All strings in the manifest MUST support `${VAR:-default}` syntax, resolved from the environment where `hwexp` is running.

---

## 3. Two-Tier Discovery Pattern (llama-swap)

The adapter MUST support a dynamic discovery flow to handle ephemeral or scaling backends.

### 3.1 Tier 1: Router Discovery
1.  Query the router endpoint (e.g., `http://${LLAMA_SWAP_HOST}:${LLAMA_SWAP_PORT}/v1/models`).
2.  Identify active models and their associated activity indicators (`requests_processing`).

### 3.2 Tier 2: Conditional Scrape
1.  **Scrape Decision:** Only scrape the backend `/metrics` (Prometheus format) if `requests_processing > 0`.
2.  **Label Injection:** Inject `model_name="<id>"` into all metrics extracted from the backend.
3.  **Aggregation:** Emit a summary metric:
    `gateway_llamacpp_prompt_tokens_total{model_name="active"}` (Sum of all active instances).

---

## 4. Hardware Correlation Contract

When `hardware_correlation` is defined, `hwexp` MUST perform a label-join between the software service and the physical device.

### 4.1 Schema
```yaml
hardware_correlation:
  device_class: gpu
  pci_slot: "${GPU_PCI_SLOT_0}"
```

### 4.2 Engine Join Logic
1.  Match the `pci_slot` against devices discovered by the `linux_gpu` adapter.
2.  Automatically inject `device_id` and `gpu_name` labels into all metrics produced by this component.

---

## 5. Universal Metrics

The adapter MUST emit the following standard metrics for every component:

| Metric | Type | Description |
| :--- | :--- | :--- |
| `gateway_component_up` | Gauge | `1` if health check passes, else `0`. |
| `gateway_component_info` | Info | Metadata about the component (version, etc). |

---

## 6. Security & Isolation
- **Permissions:** Read-only access to endpoints.
- **Authentication:** Support `Bearer` tokens via `token_env` (referencing an environment variable).
- **Network:** Resolve hostnames within the Docker internal network (`llm-gateway`, etc).

---

## 7. Implementation Tasks

| Task | Detail |
| :--- | :--- |
| **Adapter Core** | `hwexp/internal/adapters/gateway_manifest/` |
| **Variable Resolver** | Implement `${VAR:-default}` expansion. |
| **Discovery Engine** | Implement Tier 1/Tier 2 logic. |
| **Correlation Layer** | Update `internal/engine` to handle label-joins. |
| **Mapping Rules** | Add `^gateway_.*$` catch-all to `mappings.yaml` for passthrough families. |
