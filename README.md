# AUDiot Monitor

Hardware, OS, and AI telemetry stack for Linux machines. Collects sensor data, system specifications, and service status, exposes them as Prometheus metrics, and visualises them in Grafana.

## Features

- **Automatic sensor discovery** — reads temperatures, fan speeds, voltages, power, and frequencies from `/sys/class/hwmon`.
- **GPU Telemetry** — native support for **AMD** (sysfs) and **NVIDIA** (nvidia-smi) utilization and VRAM tracking.
- **Hardware Inventory** — exports deep system specs: Motherboard model, BIOS version, CPU core/thread counts, and total RAM.
- **AI/LLM Monitoring** — built-in adapter for **Llamaswap** to monitor local LLM model status.
- **Extensible Architecture** — support for custom collection scripts (Bash, Python, etc.) via the `vendor_exec` plugin system.
- **Auto-mapping** — generates Prometheus metric rules for unmapped sensors automatically.
- **Dashboard Profiles** — Pre-built layouts for standard monitors, wide sensor panels (1920x440), and mobile screens.
- **Automatic Scaffolding** — Automatically populates your dashboard library on first run while preserving your custom edits.
- **OS metrics** — CPU usage, memory, disk, network via integrated `node_exporter`.
- **Modular Config** — support for `conf.d` directory for easy, non-destructive configuration.

## Components

| Component | Purpose |
| --------- | ------- |
| `hwexp` | Go binary — universal hardware and service exporter (`:9200`) |
| `node-exporter` | OS-level performance metrics (`:9100`) |
| `prometheus` | Metrics storage and query engine (`:9090`) |
| `grafana` | Dashboard visualisation (`:3000`) |

## Quick install

Full instructions: [`monitoring/INSTALL.md`](monitoring/INSTALL.md)

```bash
# On every machine you want to monitor (collector stack):
cd monitoring/collector && docker compose up -d

# On the machine running Grafana (dashboard stack):
PROMETHEUS_URL=http://<collector-ip>:9090 docker compose -f monitoring/dashboard/docker-compose.yml up -d
```

## Metrics

All metrics use the prefix `hw_device_` or `hwexp_`:

| Metric | Type | Purpose |
| ------ | ---- | ------- |
| `hw_device_temperature_celsius` | Gauge | Temperature sensors |
| `hw_device_utilization_percent` | Gauge | CPU/GPU/LLM compute & memory load |
| `hw_device_info` | Info | Metadata (Vendor, Model, BIOS, Driver) |
| `hw_device_capacity_bytes` | Info | Fixed capacities (RAM, VRAM) |
| `hw_device_sensor_count` | Info | Counts (CPU Cores, Threads) |
| `hwexp_adapter_refresh_success`| Self | Health of the exporter itself |

## API Endpoints

- `/metrics` — Prometheus metrics (includes hardware + self-metrics)
- `/readyz` — Returns 200 OK once the first poll cycle is complete
- `/debug/catalog` — JSON summary of all detected and normalized sensors
- `/debug/discovery` — JSON view of raw discovered hardware devices

## Extension & Customization

AUDiot is designed to be customized without rebuilding the Docker image:
- **Modular Config**: Drop `.yaml` files into `/etc/hwexp/conf.d/` to enable features.
- **Custom Scripts**: Drop executable scripts into `/etc/hwexp/custom.d/` to collect your own data.

## License

MIT
