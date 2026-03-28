# Spec 100 — Component Monitor Architecture

**Status:** Draft
**Project:** AUDiotMonitor
**Covers:** System boundary definition; manifest-driven software component monitoring;
external project data integration via Prometheus; Docker image separation; CI build path
for optional runtime tools
**Related:** [spec-801](spec-801-llmgateway-monitor-adapter.md), [spec-101](spec-101-deep-hardware-telemetry.md)

---

## 1. System Boundary

AUDiotMonitor has two distinct concerns that must not blur:

| Concern | Owner | Extension point |
| --- | --- | --- |
| **Hardware metrics** | This repo | `vendor_exec` (scripts) or new Go adapters |
| **Service data** | External project | Prometheus `/metrics` or `component_manifest` |

### 1.1 Hardware Extensions (`vendor_exec`)
While core hardware adapters (e.g., `amdgpu`, `hwmon`) are built into the Go binary, users can extend hardware support without Go code using the `vendor_exec` adapter. This adapter executes external scripts (Python, Bash) from `/etc/hwexp/custom.d/` and parses their standard output as raw metrics.

### 1.2 Software Extensions (`component_manifest`)
Software services that need to be correlated with hardware use the manifest-driven adapter. This allows mapping service-specific health and performance data into the `hw_device_*` standard with common labels.

---

## 2. How External Projects Expose Data

The primary integration path for external projects is a standard Prometheus scrape. This requires no changes to `hwexp` and is suitable for metrics that do not need to be joined with hardware-specific device labels in Grafana.

```yaml
# monitoring/collector/config/prometheus/prometheus.yml
scrape_configs:
  - job_name: 'external-app'
    static_configs:
      - targets: ['host.docker.internal:9400']
```

---

## 3. Component Manifests & Hardware Correlation

A manifest is required when software metrics must be displayed alongside and **correlated** with specific hardware devices (e.g., matching an LLM model to the GPU it is actually occupying).

### 3.1 Hardware Correlation Contract
To enable seamless correlation, the manifest schema supports `hardware_correlation`. When a manifest "claims" a hardware ID, `hwexp` automatically injects matching labels from the hardware discovery layer.

```yaml
id: vllm-instance-0
hardware_correlation:
  device_class: gpu
  pci_slot: "0000:03:00.0"  # hwexp will inject device_id="pci-0000:03:00.0"
```

### 3.2 Hot Reloading
The `component_manifest` adapter re-scans its manifest directories (`/etc/hwexp/components/`) on a regular cadence (default 15s, per spec-801 §2.2). This allows adding new monitored components to a live system without restarting the collector container.

---

## 4. Docker Image Separation

| Image | Tier | Contents |
| --- | --- | --- |
| `audumla/audiot-hwexp` | Collector | hwexp binary + adapters |
| `audumla/audiot-dashboard` | Dashboard | Grafana + provisioning + kiosk binaries (wl-gammarelay extracted via `kiosk-tools` build target) |

---

## 5. Implementation Stages (Refactor Scope)

### Stage 1: Foundation & Hardware Extension
*   **Formalize `vendor_exec`**: Ensure the existing Go adapter for external scripts is robust and documented.
*   **Initial `gateway_manifest`**: Implement the HTTP-based manifest adapter from spec-801.
*   **Standard Mappings**: Add `^gateway_` rules to `mappings.yaml`.

### Stage 2: Generalized Manifests & Hot Reload
*   **Source Generalization**: Expand manifests to support `source_type: exec` and `source_type: file` in addition to `http`.
*   **Directory Watcher**: Implement the re-scan logic for `/etc/hwexp/components/`.
*   **Loader Logic**: Implement manifest merging (local overrides project).

### Stage 3: Hardware Correlation (The Join Layer)
*   **Label Injection**: Update the engine to perform label-joins between `DiscoveredDevice` (hardware) and `component_manifest` metrics.
*   **Binding Contract**: Implement the `hardware_correlation` manifest fields.

### Stage 4: Ecosystem & Correlation Dashboards

*   **Kiosk Build**: Complete — `kiosk-tools` is a build target in `monitoring/dashboard/Dockerfile`; `dietpi-setup.sh` extracts binaries from the dashboard image.
*   **Global Dashboard**: Finalize the `llmgateway.json` dashboard with hardware-software correlation panels (requires Stage 3).

### Stage 5: Deep Hardware Telemetry (Storage)
*   **Native `linux_storage` Adapter**: Implement deep polling for S.M.A.R.T. and wear leveling (HDD/SSD/NVMe).
*   **Dynamic Capabilities**: Auto-detect feature support (e.g., NVMe error logs vs ATA SMART).

### Stage 6: Generic System & ACPI
*   **Native `linux_system` Adapter**: Implement generic ACPI thermal zones and CPU C-state residency metrics.
*   **Forensic Metrics**: Add system interrupts and EDAC (ECC RAM) error tracking.

---

## 6. Constraints

| Constraint | Detail |
| --- | --- |
| Hardware Correlation | Correlation is performed at the engine level via shared labels |
| Backward compatible | Stage 1 implementation must not break existing config files |
| No secrets in manifests | Tokens are referenced via environment variable names |
| Idempotent bootstrap | Re-running DietPi setup does not break existing install |
