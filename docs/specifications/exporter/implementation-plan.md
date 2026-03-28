# Detailed Implementation Plan — Component & Hardware Extension

This document breaks down the implementation of Spec 100 and Spec 101 into actionable technical tasks.

---

## Stage 1: Foundation & Hardware Extension

**Goal**: Robust script-based hardware extensions and initial HTTP manifest support.

### 1.1 Formalize `vendor_exec` Adapter

- **Location**: `hwexp/internal/adapters/vendor_exec/`
- **Tasks**:
  - [ ] Add per-script `PollTimeout` (default 5s) using `exec.CommandContext` with a child deadline — `CommandContext` is already used for cancellation, but no per-script deadline is set.
  - [ ] Add support for `source_format: prometheus` (currently only JSON output is parsed).
  - [ ] Improve error logging: log the specific script name and exit code on failure (currently `if err != nil { continue }` swallows all errors silently).
  - [ ] Update `hwexp.yaml` config struct to expose `scripts_dir` and `timeout` via `AdapterConfig.Settings`.

### 1.2 Implement `gateway_manifest` Adapter (New)

- **Location**: `hwexp/internal/adapters/gateway_manifest/`
- **Note**: This is a new adapter built from scratch. The `llamaswap` stub adapter has been removed. The example manifest at `monitoring/collector/config/hwexp/components/llamaswap.yaml.example` shows the target schema.
- **Tasks**:
  - [ ] Implement `model.go`: Define `ComponentManifest` struct (id, display_name, connection, health, metrics, hardware_correlation).
  - [ ] Implement `loader.go`: Shallow merge logic — load from `/etc/hwexp/components/`, then apply overrides from `/etc/hwexp/local/components/` by id. Lists (`metrics`, `actions`) merged by `id` field.
  - [ ] Implement `resolver.go`: `${VAR:-default}` variable expansion for all manifest string fields, resolved from the hwexp process environment.
  - [ ] Implement `extractor.go`: JQ-style path extraction for JSON responses; Prometheus text-format parsing for `source_format: prometheus`.
  - [ ] Implement `adapter.go`: `Discover()` scans manifests and emits `gateway_component_up` and `gateway_component_info`; `Poll()` performs HTTP GET against each metric endpoint.

### 1.3 Standard Mappings

- **File**: `monitoring/collector/config/hwexp/mappings.yaml`
- **Tasks**:
  - [ ] Add `^gateway_` regex pass-through rule so manifest metrics reach Prometheus without being dropped.
  - [ ] Add explicit `gateway_component_up` mapping with description.

---

## Stage 2: Generalized Manifests & Hot Reload

**Goal**: Expansion to non-HTTP sources and zero-restart configuration.

### 2.1 Source Generalization

- **Tasks**:
  - [ ] Add `source_type: exec` to manifest schema (runs a local command and parses stdout).
  - [ ] Add `source_type: file` to manifest schema (reads a file from disk).
  - [ ] Unify extraction logic so JQ/Prometheus parsing in `extractor.go` works across all source types.

### 2.2 Hot Reloading

- **Location**: `hwexp/internal/adapters/gateway_manifest/`
- **Cadence**: 15 seconds (per llmgateway-monitor-adapter §2.2).
- **Tasks**:
  - [ ] Implement a timer-based re-scan (15s) of the components directories inside the adapter's background loop.
  - [ ] On each re-scan, reload and re-merge manifests; atomically replace the active manifest list so in-flight polls are unaffected.

---

## Stage 3: Hardware Correlation (The Join Layer)

**Goal**: Native binding of software metrics to physical devices.

### 3.0 Shared Device Index (Prerequisite)

- **Location**: `hwexp/internal/engine/`
- **Note**: The label-join in Stage 3.2 requires a queryable index of discovered hardware. This does not currently exist and must be added before Stage 3.2.
- **Tasks**:
  - [ ] Add a `DeviceIndex` type (e.g., `map[string]*model.DiscoveredDevice` keyed by stable ID and by PCI slot) to the engine.
  - [ ] Populate the index after each `Discover()` cycle across all hardware adapters.
  - [ ] Expose a read-only lookup: `LookupByPCISlot(slot string) (*model.DiscoveredDevice, bool)`.

### 3.1 Hardware Claim Fields in Manifest Model

- **Location**: `hwexp/internal/model/`
- **Tasks**:
  - [ ] Add `HardwareCorrelation` struct to `ComponentManifest` model.
  - [ ] Define matching fields: `device_class`, `pci_slot`, `logical_name`.

### 3.2 Engine Label Injection

- **Location**: `hwexp/internal/engine/`
- **Prerequisite**: Stage 3.0 (DeviceIndex must exist).
- **Tasks**:
  - [ ] After polling a manifest component, check if `hardware_correlation` is set.
  - [ ] Look up the corresponding `DiscoveredDevice` from the `DeviceIndex` using `pci_slot` or `device_class`.
  - [ ] If a match is found, inject `device_id`, `vendor`, and `model` labels from the hardware device into all raw measurements for that component.

---

## Stage 4: Ecosystem & Correlation Dashboards

**Goal**: Finalize correlation dashboards using the hardware-software join labels.

### 4.1 Kiosk Toolchain

- **Status**: **Complete** — `kiosk-tools` is now a build target in `monitoring/dashboard/Dockerfile`. `dietpi-setup.sh` extracts binaries from the dashboard image.
- No further implementation required.

### 4.2 Correlation Dashboards

- **Prerequisite**: Stage 3 (joined labels must exist in Prometheus).
- **Tasks**:
  - [ ] Create `monitoring/dashboard/provisioning/dashboards/llmgateway.json` using joined labels (e.g., graph vLLM KV-cache usage alongside GPU temperature, linked via `device_id`).

---

## Stage 5: Deep Hardware Telemetry (Storage)

**Goal**: S.M.A.R.T. and forensic disk analysis.

### 5.1 `linux_storage` Adapter

- **Location**: `hwexp/internal/adapters/linux_storage/`
- **Note**: CGO is disabled in current builds (`CGO_ENABLED=0`). Native ioctl calls require CGO. **Start with the `smartctl --json` wrapper** for broad compatibility; defer native NVMe ioctl to a follow-on task if CGO is enabled.
- **Tasks**:
  - [ ] Implement `smartctl --json` wrapper: discover block devices via `/sys/class/block`, detect NVMe vs SATA/SAS interface, invoke `smartctl --json -a /dev/X` and parse output.
  - [ ] Normalize to standard metrics: `hw_device_disk_health_status`, `hw_device_disk_life_remaining_percent`, `hw_device_disk_reallocated_sectors_total`, `hw_device_disk_power_on_hours_total`, `hw_device_disk_errors_total`.
  - [ ] Dynamically register capabilities (`smart`, `wear_leveling`) based on device support.
  - [ ] (Optional/future) Implement native NVMe Smart Log parsing via `/dev/nvmeX` ioctl — requires enabling CGO.

---

## Stage 6: Generic System & ACPI

**Goal**: Exposing deep kernel/ACPI signals.

### 6.1 `linux_system` Adapter

- **Location**: `hwexp/internal/adapters/linux_system/`
- **Tasks**:
  - [ ] **C-States**: Parse `/sys/devices/system/cpu/cpu*/cpuidle/state*/residency` to calculate per-state residency percentage (`hw_device_cpu_cstate_residency_percent`).
  - [ ] **Thermal Zones**: Read `/sys/class/thermal/thermal_zone*` for generic ACPI zones not covered by `linux_hwmon`.
  - [ ] **Interrupts**: Implement a delta-calculator for `/proc/interrupts` (`hw_device_system_interrupts_total`).
  - [ ] **EDAC**: Scan `/sys/devices/system/edac/mc/mc*/` for correctable/uncorrectable ECC memory error counts (`hw_device_memory_ecc_errors_total`).




