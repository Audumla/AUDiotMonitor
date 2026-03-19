# AUDiot (hwexp) Specification

This document outlines the architecture, taxonomy, and implementation plan for the AUDiot hardware telemetry platform.

---

## 10. External Interfaces

### 10.1 Exporter HTTP API
Implemented endpoints:
- `/metrics`: Prometheus scrape endpoint (Hardware + AI + Self-metrics).
- `/healthz`: Overall process liveness.
- `/readyz`: Readiness for scraping (Returns 200 OK once the first poll cycle is complete).
- `/version`: Version and platform information.
- `/debug/discovery`: JSON describing discovered devices and source metadata.
- `/debug/mappings`: JSON describing active mapping decisions.
- `/debug/catalog`: JSON summary of the normalized sensor catalog.
- `/debug/raw`: Raw sensor data (requires `debug.enable_raw_endpoint` in config).

---

## 13. Taxonomy

### 13.1 Top-level Device Classes
`host`, `cpu`, `gpu`, `psu`, `memory`, `motherboard`, `fan`, `storage`, `network`, `llm`, `sensor`, `unknown`.

### 13.3 Sensor Types
`temperature`, `power`, `voltage`, `current`, `fan_speed`, `utilization`, `frequency`, `energy`, `capacity`, `cores`, `threads`, `status`.

---

## 14. Metric Naming Standard

### 14.1 Base Metric Families
- `hw_device_temperature_celsius`: Thermal sensors.
- `hw_device_utilization_percent`: Compute/Memory load (0-100).
- `hw_device_info`: Metadata info metric (labels: vendor, model, driver, bios, revision).
- `hw_device_capacity_bytes`: Fixed capacities (RAM, VRAM).
- `hw_device_sensor_count`: Hardware counts (Cores, Threads).
- `hw_device_fan_rpm`: Cooling speeds.
- `hw_device_power_watts`: Electrical power.

---

## 18. Platform Adapters

### 18.1 Linux / Universal
- `linux_hwmon`: Kernel hwmon subsystem.
- `linux_gpu`: AMD (sysfs) and NVIDIA (nvidia-smi) telemetry.
- `linux_static`: System inventory (DMI, /proc).
- `llamaswap`: Local LLM model monitoring.
- `vendor_exec`: Custom external scripts/plugins.

---

## 28. Observability of the Exporter Itself
The exporter tracks its own health via:
- `hwexp_adapter_refresh_duration_seconds`
- `hwexp_adapter_refresh_success` (1.0 = Success)
- `hwexp_adapter_last_success_unixtime`
- `hwexp_discovered_devices`
- `hwexp_mapping_failures_total`
