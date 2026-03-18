# AUDiot Monitor — Installation Guide

Deploys a full hardware + OS monitoring stack on any Linux machine using Docker Compose.

**What gets installed:**
| Service | Port | Purpose |
|---------|------|---------|
| hwexp | 9200 | Hardware sensor exporter (CPU/GPU/PSU temps, fans, voltages, power) |
| node-exporter | 9100 | OS metrics (CPU%, memory, disk, network, load) |
| Prometheus | 9090 | Metrics storage |
| Grafana | 3000 | Dashboards |

---

## Prerequisites

- Linux machine (x86_64, arm64, or armv7 — Raspberry Pi included)
- Docker Engine ≥ 24 and Docker Compose v2
- Ports 3000, 9090, 9100, 9200 accessible on your network

### Install Docker (if not already installed)

**Ubuntu / Debian / Raspberry Pi OS:**
```bash
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
# Log out and back in, then verify:
docker run --rm hello-world
```

**OpenSUSE / Fedora:**
```bash
sudo zypper install docker          # OpenSUSE
# or: sudo dnf install docker       # Fedora
sudo systemctl enable --now docker
sudo usermod -aG docker $USER
```

---

## Installation

### 1. Copy files to the target machine

From your development machine, copy the `rpi-dashboard-node` directory:

```bash
# Using scp
scp -r rpi-dashboard-node/ user@<machine-ip>:~/audiot-monitor/

# Or clone the repo directly on the machine
git clone <repo-url> && cd AUDiotMonitor/rpi-dashboard-node
```

### 2. Start the stack

```bash
cd ~/audiot-monitor   # or wherever you copied the files

docker compose up -d
```

That's it. The stack starts immediately. On first run Docker pulls the images (~500 MB total).

### 3. Verify everything is running

```bash
docker compose ps
```

All four services should show `running`. Check individual services:

```bash
# Hardware exporter health
curl http://localhost:9200/healthz

# Sensor metrics (available within ~5 seconds of first start)
curl http://localhost:9200/metrics

# What sensors were discovered
curl http://localhost:9200/debug/discovery

# Check any sensors that didn't get a mapping rule
curl http://localhost:9200/debug/mappings | python3 -m json.tool | grep '"ignored"'

# Prometheus scraping
curl http://localhost:9090/-/healthy

# Node exporter
curl http://localhost:9100/metrics | head -20
```

### 4. Open Grafana

Navigate to `http://<machine-ip>:3000` in a browser.

Default login: **admin / admin** (you'll be prompted to change the password).

Three dashboards are pre-loaded in the **AUDiot** folder:
- **AUDiot System Overview** — comprehensive view: OS stats, all hardware sensors, network, disk
- **AUDiot Panel Display** — compact at-a-glance status board
- **AUDiot Discovery / Operations** — debug views

---

## How Auto-Mapping Works

On first start, hwexp reads every sensor in `/sys/class/hwmon`. Any sensor that doesn't have a manual rule in `mappings.yaml` gets an auto-generated rule within the first 5-second poll cycle.

Auto-generated rules are written to `config/hwexp/mappings.auto.yaml` — inspect this file to see exactly what was found:

```bash
cat config/hwexp/mappings.auto.yaml
```

Rules are grouped by device class and sensor type (e.g., one rule covers all AMD GPU temperature sensors). If you want to customise a rule (rename a sensor, change thresholds), copy it from `mappings.auto.yaml` into `mappings.yaml` and set a higher `priority`. Manual rules always win.

**New hardware added later** is picked up automatically within 5 seconds — no restart needed. If the new hardware introduces a new *type* of sensor that wasn't seen before, a new auto-generated rule is created and written to `mappings.auto.yaml` immediately.

To reload `mappings.yaml` after editing (without restarting the container):
```bash
docker kill --signal=SIGHUP hwexp
```

---

## Open Firewall Ports (if needed)

**firewalld (OpenSUSE / Fedora / RHEL):**
```bash
sudo firewall-cmd --zone=public --add-port=3000/tcp --permanent  # Grafana
sudo firewall-cmd --zone=public --add-port=9090/tcp --permanent  # Prometheus
sudo firewall-cmd --zone=public --add-port=9200/tcp --permanent  # hwexp
sudo firewall-cmd --reload
```

**ufw (Ubuntu / Debian / Raspberry Pi OS):**
```bash
sudo ufw allow 3000/tcp   # Grafana
sudo ufw allow 9090/tcp   # Prometheus
sudo ufw allow 9200/tcp   # hwexp
```

Port 9100 (node-exporter) is internal-only and does not need to be opened externally.

---

## Updates

Pull the latest images and restart:
```bash
docker compose pull
docker compose up -d
```

Your Prometheus data, Grafana dashboards, and auto-generated mappings are stored in named Docker volumes (`prometheus-data`, `grafana-data`) and the local `config/` directory. They survive updates.

---

## Stopping / Removing

```bash
# Stop without removing data
docker compose down

# Stop and remove all data (Prometheus history, Grafana settings)
docker compose down -v
```

---

## Designing Your Own Dashboards

The provisioning configuration has `editable: true` and `allowUiUpdates: true`. You can:

1. Open Grafana → navigate to any dashboard → click the pencil icon to edit
2. Add panels, rearrange, change queries
3. Save — changes are written back to the JSON files in `dashboards/`

All hwexp hardware metrics use the prefix `hw_device_` and always have a `host` label:

| Metric | Unit | What it covers |
|--------|------|---------------|
| `hw_device_temperature_celsius` | °C | CPU, GPU, NVMe, motherboard, PSU temps |
| `hw_device_fan_rpm` | RPM | All fans |
| `hw_device_power_watts` | W | GPU power draw, PSU output power |
| `hw_device_voltage_volts` | V | Rail voltages |
| `hw_device_current_amps` | A | Current draw |
| `hw_device_frequency_hz` | Hz | GPU/CPU clock frequencies |
| `hw_device_energy_joules` | J | Cumulative energy |

Key labels on every `hw_device_*` metric: `host`, `device_class`, `logical_name`, `component`, `sensor`.

Example PromQL queries:
```promql
# All GPU temperatures
hw_device_temperature_celsius{device_class="gpu"}

# CPU package temperature for a specific host
hw_device_temperature_celsius{host="myhost", device_class="cpu", component="thermal"}

# Total power across all PSU outputs
sum(hw_device_power_watts{device_class="psu"})

# Fan speeds above 1000 RPM
hw_device_fan_rpm > 1000
```
