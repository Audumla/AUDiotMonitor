# AUDiot Monitor

Hardware and OS telemetry stack for Linux machines. Collects sensor data from the kernel's hwmon subsystem, exposes it as Prometheus metrics, and visualises it in Grafana.

## What it does

- **Automatic sensor discovery** — reads every sensor in `/sys/class/hwmon` on startup (temperatures, fan speeds, voltages, power, current, frequencies)
- **Auto-mapping** — generates Prometheus metric rules for unmapped sensors on first run, no manual config needed
- **OS metrics** — CPU usage, memory, disk, network and load via node_exporter
- **Grafana dashboards** — pre-built dashboards provisioned automatically on deploy
- **Multi-arch Docker image** — runs on x86_64, arm64 and Raspberry Pi (armv7)

## Components

| Component | Purpose |
| --------- | ------- |
| `hwexp` | Go binary — reads hwmon sysfs, exposes metrics on `:9200/metrics` |
| `node-exporter` | OS-level metrics (CPU%, memory, disk, network) |
| `prometheus` | Metrics storage and query engine |
| `grafana` | Dashboard visualisation |

## Quick install

Full instructions: [`rpi-dashboard-node/INSTALL.md`](rpi-dashboard-node/INSTALL.md)

```bash
# Copy rpi-dashboard-node/ to target machine, then:
docker compose up -d
```

Grafana at `http://<host>:3000` — default login `admin / admin`.

## Docker Hub

```text
audumla/audiot-hwexp:latest
```

Multi-arch: `linux/amd64`, `linux/arm64`, `linux/arm/v7`

## Repository layout

```text
hwexp/                      Go source for the hardware exporter
  cmd/hwexp/                Binary entrypoint
  internal/
    adapters/hwmon/         Linux hwmon sysfs adapter
    automapper/             Auto-generates metric rules for unmapped sensors
    engine/                 Poll and discovery loop
    mapper/                 Rule-based metric normalisation
    httpapi/                Prometheus metrics + debug endpoints
  packaging/                .deb / .rpm packaging (nfpm)
  tests/integration/        Install and smoke tests

rpi-dashboard-node/         Docker Compose stack for deployment
  config/
    hwexp/                  hwexp config and mapping rules
    prometheus/             Prometheus scrape config
    grafana/                Grafana provisioning
  dashboards/               Pre-built Grafana dashboard JSON
  INSTALL.md                Step-by-step install guide
```

## Metrics

All hardware metrics use the prefix `hw_device_` with a `host` label:

| Metric | Unit |
| ------ | ---- |
| `hw_device_temperature_celsius` | °C |
| `hw_device_fan_rpm` | RPM |
| `hw_device_power_watts` | W |
| `hw_device_voltage_volts` | V |
| `hw_device_current_amps` | A |
| `hw_device_frequency_hz` | Hz |

## CI / CD

- **PRs** — build, vet, unit tests, integration tests (deb/rpm install, hwmon smoke test)
- **Nightly** — full test suite
- **Release** — `hwexp-v*` tag → builds packages + pushes multi-arch Docker image to Docker Hub

## License

MIT
