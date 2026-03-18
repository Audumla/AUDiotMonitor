# AUDiot Monitor — Installation Guide

Two independent Docker Compose stacks — deploy them separately depending on your setup:

| Stack | Directory | What it runs | Deploy on |
|-------|-----------|--------------|-----------|
| **server** | `dashboard/server/` | hwexp + node-exporter + Prometheus | Every machine you want to monitor |
| **display** | `dashboard/display/` | Grafana | Any machine on the same network |

Both stacks can run on the same host, or the display stack can run on a dedicated machine (e.g. a Raspberry Pi kiosk).

---

## Prerequisites

- Linux machine (x86_64, arm64, or armv7 — Raspberry Pi included)
- Docker Engine ≥ 24 and Docker Compose v2
- Required ports: server uses 9090, 9100, 9200; display uses 3000

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

## Deploy: Server Stack (monitored machine)

Copy `dashboard/server/` to the target machine, then:

```bash
cd dashboard/server
docker compose up -d
```

### Verify

```bash
docker compose ps

# Hardware exporter metrics (available within ~5 s of first start)
curl http://localhost:9200/metrics | head -20

# What sensors were auto-discovered
curl http://localhost:9200/debug/discovery

# Prometheus health
curl http://localhost:9090/-/healthy
```

### Open firewall ports (if needed)

**firewalld (OpenSUSE / Fedora / RHEL):**
```bash
sudo firewall-cmd --zone=public --add-port=9090/tcp --permanent  # Prometheus
sudo firewall-cmd --zone=public --add-port=9200/tcp --permanent  # hwexp
sudo firewall-cmd --reload
```

**ufw (Ubuntu / Debian / Raspberry Pi OS):**
```bash
sudo ufw allow 9090/tcp
sudo ufw allow 9200/tcp
```

Port 9100 (node-exporter) is internal to the server stack and does not need external access.

---

## Deploy: Display Stack (Grafana)

Copy `dashboard/display/` to the display machine, then set `PROMETHEUS_URL` to point at your server:

```bash
cd dashboard/display

# Point at your server (replace with the server's IP or hostname)
export PROMETHEUS_URL=http://192.168.1.10:9090

docker compose up -d
```

Or create a `.env` file in `dashboard/display/`:

```
PROMETHEUS_URL=http://192.168.1.10:9090
GF_ADMIN_PASSWORD=changeme
```

Then run `docker compose up -d`.

### Open Grafana

Navigate to `http://<display-host>:3000`.

Default login: **admin / admin** (you'll be prompted to change the password).

Three dashboards are pre-loaded in the **AUDiot** folder:
- **AUDiot System Overview** — comprehensive view: OS stats, all hardware sensors, network, disk
- **AUDiot Panel Display** — compact at-a-glance status board
- **AUDiot Discovery / Operations** — debug views

### Open firewall ports (if needed)

```bash
sudo firewall-cmd --zone=public --add-port=3000/tcp --permanent && sudo firewall-cmd --reload
# or
sudo ufw allow 3000/tcp
```

---

## How Auto-Mapping Works

On first start, hwexp reads every sensor in `/sys/class/hwmon`. Any sensor that doesn't have a manual rule in `config/hwexp/mappings.yaml` gets an auto-generated rule within the first 5-second poll cycle.

Auto-generated rules are written to `config/hwexp/mappings.auto.yaml`:

```bash
cat dashboard/server/config/hwexp/mappings.auto.yaml
```

Rules are grouped by device class and sensor type (e.g., one rule covers all AMD GPU temperature sensors). To customise a rule, copy it from `mappings.auto.yaml` into `mappings.yaml` and set a higher `priority`. Manual rules always override auto-generated ones.

To reload `mappings.yaml` after editing (no container restart needed):
```bash
docker kill --signal=SIGHUP hwexp
```

---

## Updates

```bash
# Server
cd dashboard/server && docker compose pull && docker compose up -d

# Display
cd dashboard/display && docker compose pull && docker compose up -d
```

Prometheus data and Grafana settings are stored in named Docker volumes and survive updates.

---

## Stopping / Removing

```bash
# Stop without removing data
docker compose down

# Stop and remove all data (Prometheus history, Grafana settings)
docker compose down -v
```

---

## Designing Dashboards

The provisioning configuration has `allowUiUpdates: true`. Edit dashboards in the Grafana UI, then export the JSON and save it back into `display/dashboards/`.

All hwexp hardware metrics use the prefix `hw_device_` and have a `host` label:

| Metric | Unit | What it covers |
|--------|------|----------------|
| `hw_device_temperature_celsius` | °C | CPU, GPU, NVMe, motherboard, PSU temps |
| `hw_device_fan_rpm` | RPM | All fans |
| `hw_device_power_watts` | W | GPU power draw, PSU output |
| `hw_device_voltage_volts` | V | Rail voltages |
| `hw_device_current_amps` | A | Current draw |
| `hw_device_frequency_hz` | Hz | GPU/CPU clock frequencies |
| `hw_device_energy_joules` | J | Cumulative energy |

Key labels: `host`, `device_class`, `logical_name`, `component`, `sensor`.

Example PromQL:
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
