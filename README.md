# AUDiot Monitor

Hardware, OS, and AI telemetry stack for Linux machines. Collects sensor data, system specifications, and service status, exposes them as Prometheus metrics, and visualises them in Grafana.

## Features

- **Automatic sensor discovery** — reads temperatures, fan speeds, voltages, power, and frequencies from `/sys/class/hwmon`.
- **GPU Telemetry** — native support for **AMD** (sysfs) and **NVIDIA** (nvidia-smi) utilization and VRAM tracking.
- **Hardware Inventory** — exports deep system specs: motherboard model, BIOS version, CPU core/thread counts, total RAM, NIC vendor/model.
- **AI/LLM Monitoring** — built-in adapter for **Llamaswap** to monitor local LLM model status.
- **Extensible Architecture** — support for custom collection scripts (Bash, Python, etc.) via the `vendor_exec` plugin system.
- **Auto-mapping** — generates Prometheus metric rules for unmapped sensors automatically.
- **Dashboard Profiles** — pre-built layouts for standard monitors, wide sensor panels (1920×440), and mobile screens.
- **Automatic Scaffolding** — populates your dashboard library on first run while preserving your custom edits.
- **OS metrics** — CPU usage, memory, disk I/O, network, RAPL power, systemd service states via integrated `node_exporter`.
- **Modular Config** — drop `.yaml` files into `/etc/hwexp/conf.d/` for non-destructive configuration.

## Components

| Component | Purpose | Port |
| --------- | ------- | ---- |
| `hwexp` | Go binary — universal hardware and service exporter | `:9200` |
| `node-exporter` | OS-level performance metrics | `:9100` |
| `prometheus` | Metrics storage and query engine | `:9090` |
| `grafana` | Dashboard visualisation | `:3000` |

## Quick install

Full instructions: [`monitoring/INSTALL.md`](monitoring/INSTALL.md)

```bash
# On every machine you want to monitor (collector stack):
cd monitoring/collector && docker compose up -d

# On the machine running Grafana (can be a different machine, e.g. a Raspberry Pi):
PROMETHEUS_URL=http://<collector-ip>:9090 \
HWEXP_URL=http://<collector-ip>:9200 \
docker compose -f monitoring/dashboard/docker-compose.yml up -d
```

Grafana opens at **<http://localhost:3000>** (default login: admin / admin).

## API Endpoints

`hwexp` exposes a self-documenting index page at the root URL:

| Endpoint | Auth | Description |
| -------- | ---- | ----------- |
| `GET /` | open | HTML index — lists all endpoints |
| `GET /metrics` | open | Prometheus text format — scrape target |
| `GET /healthz` | open | Liveness probe |
| `GET /readyz` | open | Readiness probe (200 once first poll completes) |
| `GET /version` | open | Exporter and schema version strings |
| `GET /debug/state` | debug:read | All devices joined with current measurements (JSON) |
| `GET /debug/catalog` | debug:read | Flat list of all normalised measurements |
| `GET /debug/discovery` | debug:read | Discovered device inventory |
| `GET /debug/mappings` | debug:read | Mapping rule decisions per measurement |
| `GET /debug/raw` | debug:read | Raw adapter output (requires `enable_raw_endpoint: true`) |

`debug:read` endpoints require a Bearer token when `auth_mode: bearer_token` is set in config; otherwise they are open.

## Metrics

| Metric | Type | Description |
| ------ | ---- | ----------- |
| `hw_device_temperature_celsius` | Gauge | Temperature sensors |
| `hw_device_utilization_percent` | Gauge | CPU / GPU / LLM compute & memory load |
| `hw_device_info` | Info | Vendor, Model, BIOS version, Driver, NIC model |
| `hw_device_capacity_bytes` | Info | RAM / VRAM total capacity |
| `hw_device_sensor_count` | Info | CPU core / thread counts |
| `hwexp_adapter_refresh_success` | Gauge | Exporter self-health |
| `hwexp_discovered_devices` | Gauge | Number of discovered devices |
| `hwexp_mapping_failures_total` | Counter | Mapping rule failures |

## Extension & Customization

AUDiot is designed to be customized without rebuilding the Docker image:

- **Modular Config**: Drop `.yaml` files into `/etc/hwexp/conf.d/` to enable features without touching the main config.
- **Custom Scripts**: Drop executable scripts into `/etc/hwexp/custom.d/` to collect your own metrics in any language.

## License

MIT
