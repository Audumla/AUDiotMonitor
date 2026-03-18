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

## Server Stack — `dashboard/server/docker-compose.yml`

Runs on every machine you want to monitor. Exposes hardware and OS metrics to Prometheus.

```yaml
services:

  hwexp:
    image: audumla/audiot-hwexp:latest
    container_name: hwexp
    privileged: true                        # required to read /sys/class/hwmon
    volumes:
      - /sys:/sys:ro
      - /proc:/proc:ro
      - ./config/hwexp:/etc/hwexp:z         # hwexp.yaml + mappings.yaml config
    environment:
      - HWEXP_HOST=${HOSTNAME:-localhost}   # label applied to all hw_device_* metrics
    ports:
      - "9200:9200"
    restart: unless-stopped

  node-exporter:
    image: prom/node-exporter:latest
    container_name: node-exporter
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    command:
      - '--path.procfs=/host/proc'
      - '--path.rootfs=/rootfs'
      - '--path.sysfs=/host/sys'
      - '--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)'
    ports:
      - "9100:9100"
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    volumes:
      - ./config/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro,z
      - prometheus-data:/prometheus
    ports:
      - "9090:9090"
    restart: unless-stopped

volumes:
  prometheus-data:
```

**`.env.example` (copy to `.env` and customise):**

```env
# Hostname label on all hw_device_* metrics — defaults to $HOSTNAME
HWEXP_HOST=myserver
```

### Deploy server

```bash
cd dashboard/server
cp .env.example .env          # optional — edit HWEXP_HOST if needed
docker compose up -d
```

### Verify

```bash
docker compose ps

# Hardware metrics (available within ~5 s of first start)
curl http://localhost:9200/metrics | head -20

# Sensors discovered on this machine
curl http://localhost:9200/debug/discovery

# Prometheus health check
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

## Display Stack — `dashboard/display/docker-compose.yml`

Runs Grafana only. Can be on the same machine as the server or anywhere else on the network.

```yaml
services:

  grafana:
    image: grafana/grafana:latest
    container_name: audiot-grafana
    ports:
      - "3000:3000"
    volumes:
      - ./config/grafana/provisioning:/etc/grafana/provisioning:ro,z
      - ./config/grafana/grafana.ini:/etc/grafana/grafana.ini:ro,z
      - ./dashboards:/var/lib/grafana/dashboards:ro,z
      - grafana-data:/var/lib/grafana
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=${GF_AUTH_ANONYMOUS_ENABLED:-true}
      - GF_AUTH_ANONYMOUS_ORG_ROLE=${GF_AUTH_ANONYMOUS_ORG_ROLE:-Viewer}
      - GF_SECURITY_ADMIN_PASSWORD=${GF_ADMIN_PASSWORD:-admin}
      - PROMETHEUS_URL=${PROMETHEUS_URL:-http://localhost:9090}   # point at your server

volumes:
  grafana-data:
```

**`.env.example` (copy to `.env` and customise):**

```env
# URL of Prometheus on the server stack — change to the server's IP or hostname
PROMETHEUS_URL=http://192.168.1.10:9090

# Grafana admin password (default: admin — change in production)
GF_ADMIN_PASSWORD=admin

# Anonymous read-only access — set to false to require login for all users
GF_AUTH_ANONYMOUS_ENABLED=true
GF_AUTH_ANONYMOUS_ORG_ROLE=Viewer
```

### Deploy display

```bash
cd dashboard/display
cp .env.example .env
# Edit .env — set PROMETHEUS_URL to your server's IP
docker compose up -d
```

### Open Grafana

Navigate to `http://<display-host>:3000`.

Default login: **admin / admin** (you'll be prompted to change the password on first login).

Three dashboards are pre-loaded in the **AUDiot** folder:
- **AUDiot System Overview** — comprehensive view: OS stats, all hardware sensors, network, disk
- **AUDiot Panel Display** — compact at-a-glance status board
- **AUDiot Discovery / Operations** — debug views

Use the **host** template variable at the top of each dashboard to switch between monitored machines.

### Open firewall ports (if needed)

```bash
sudo firewall-cmd --zone=public --add-port=3000/tcp --permanent && sudo firewall-cmd --reload
# or
sudo ufw allow 3000/tcp
```

---

## Common Scenarios

### Single machine (all-in-one)

```bash
cd dashboard/server  && docker compose up -d
cd dashboard/display && docker compose up -d   # PROMETHEUS_URL defaults to localhost:9090
```

### Dedicated display (RPi kiosk + multiple servers)

```text
Server A (192.168.1.10) — runs dashboard/server/
Server B (192.168.1.11) — runs dashboard/server/
RPi display            — runs dashboard/display/ → PROMETHEUS_URL=http://192.168.1.10:9090
```

*To scrape multiple servers from one Prometheus, add extra scrape targets to `server/config/prometheus/prometheus.yml`.*

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
cd dashboard/server  && docker compose pull && docker compose up -d
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
