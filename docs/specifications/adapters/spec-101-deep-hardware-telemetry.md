# Spec 101 — Deep Hardware Telemetry

**Status:** Draft
**Project:** AUDiotMonitor
**Covers:** Native adapters for S.M.A.R.T. storage data, generic CPU power states, and system-wide hardware pressure metrics.
**Related:** [Spec 100](../foundation/spec-100-component-architecture.md)

---

## 1. Goal
Move beyond basic inventory and temperature monitoring to provide "forensic" hardware telemetry using native Go adapters.

---

## 2. Storage Tier (Native `linux_storage` Adapter)

The existing `linux_static` block device discovery will be supplemented by a dedicated storage adapter capable of deep polling.

### 2.1 S.M.A.R.T. & Health Data
The adapter will probe for S.M.A.R.T. capabilities and expose normalized metrics for HDD, SSD, and NVMe drives.

| Metric | Type | Description |
| :--- | :--- | :--- |
| `hw_device_disk_health_status` | Gauge | State-set: `0=OK`, `1=Warning`, `2=Critical`, `3=Failing`. |
| `hw_device_disk_life_remaining_percent` | Gauge | Estimated wear leveling (SSD/NVMe). |
| `hw_device_disk_reallocated_sectors_total` | Counter | Total count of relocated sectors (HDD failure predictor). |
| `hw_device_disk_power_on_hours_total` | Counter | Lifetime power-on hours. |
| `hw_device_disk_errors_total` | Counter | Cumulative sum of media and data integrity errors. |

### 2.2 Discovery & Capability Logic
1.  Identify block devices via `/sys/class/block`.
2.  Check for NVMe vs SCSI/ATA interface.
3.  Use native `ioctl` calls or standard binary wrappers (e.g., `smartctl --json`) to extract attributes.
4.  Dynamically register capabilities (`smart`, `wear_leveling`) based on device support.

---

## 3. Generic System Telemetry (Non-Vendor Specific)

These metrics capture hardware-level behavior that is standard across platforms (ACPI/Kernel standards).

### 3.1 CPU Power Management
| Metric | Type | Description |
| :--- | :--- | :--- |
| `hw_device_cpu_cstate_residency_percent` | Gauge | Time spent in deep sleep states (C1-C10). |
| `hw_device_cpu_throttling_events_total` | Counter | Count of thermal or power limit throttling events. |

### 3.2 Thermal Zones (ACPI)
Expose generic thermal zones from `/sys/class/thermal/thermal_zone*` that aren't captured by specific `hwmon` drivers.
*   Metric: `hw_device_temperature_celsius{device_class="thermal", sensor="zone"}`.

### 3.3 System Pressure (Hardware Context)
Capture hardware-triggered events that impact OS stability.
| Metric | Type | Description |
| :--- | :--- | :--- |
| `hw_device_system_interrupts_total` | Counter | Total hardware interrupt count (CPU pressure). |
| `hw_device_memory_ecc_errors_total` | Counter | Correctable/Uncorrectable errors from EDAC drivers. |

---

## 4. Implementation Strategy

### 4.1 Native over Scripts
While `vendor_exec` remains for user customization, these "Deep" metrics must be built as native Go adapters to ensure:
*   High performance (low overhead polling).
*   Rich metadata injection.
*   Automatic discovery without user configuration.

### 4.2 Modular Adapter Pattern
The `internal/adapters` package will be expanded:
*   `linux_storage`: Handles block/NVMe depth.
*   `linux_system`: Handles generic ACPI, Thermal, and CPU states.
