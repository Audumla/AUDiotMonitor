# AUDiot Monitor — Installation Guide

The monitoring system is split into two independent Docker Compose stacks.
Deploy each one separately depending on your infrastructure:

| Stack | Directory | Purpose | Deploy on |
| --- | --- | --- | --- |
| **Collector** | `monitoring/collector/` | Scrapes hardware and OS metrics, stores them in Prometheus | Every machine you want to monitor |
| **Dashboard** | `monitoring/dashboard/` | Runs Grafana; queries Prometheus to visualise the metrics | Any machine on the same network |

Both stacks can run on the same host. Alternatively the dashboard can run on a
dedicated machine (e.g. a Raspberry Pi kiosk) while the collector runs on each
machine being monitored.

---

## Prerequisites

- Linux machine (x86_64, arm64, or armv7 — Raspberry Pi included)
- Docker Engine ≥ 24 and Docker Compose v2
- Collector ports: 9090 (Prometheus), 9100 (node-exporter, internal), 9200 (hwexp)
- Dashboard port: 3000 (Grafana)

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

## Collector Stack — `monitoring/collector/`

Runs on every machine you want to monitor. Contains three services:

| Service | Port | What it does |
| --- | --- | --- |
| `hwexp` | 9200 | Reads `/sys/class/hwmon`, exposes hardware sensor metrics |
| `node-exporter` | 9100 | Exposes OS metrics (CPU%, memory, disk, network, load) |
| `prometheus` | 9090 | Scrapes both exporters, stores time-series data |

### Collector `docker-compose.yml`

```yaml
services:

  hwexp:
    image: audumla/audiot-hwexp:latest
    container_name: hwexp
    privileged: true                         # required to read /sys/class/hwmon
    volumes:
      - /sys:/sys:ro
      - /proc:/proc:ro
      - ./config/hwexp:/etc/hwexp:z          # hwexp.yaml + mappings.yaml
    environment:
      - HWEXP_HOST=${HWEXP_HOST:-localhost}   # label applied to all hw_device_* metrics
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

### Collector environment variables (`.env.example`)

```env
# Hostname label applied to all hw_device_* metrics.
# Defaults to the machine's $HOSTNAME if not set.
HWEXP_HOST=myserver
```

### Deploy the collector

```bash
cd monitoring/collector
cp .env.example .env    # optional — set HWEXP_HOST if the default $HOSTNAME is not suitable
docker compose up -d
```

### Verify the collector is running

```bash
docker compose ps

# Hardware metrics — available within ~5 s of first start
curl http://localhost:9200/metrics | head -20

# Full list of sensors discovered on this machine
curl http://localhost:9200/debug/discovery

# Prometheus scrape health
curl http://localhost:9090/-/healthy
```

### Open firewall ports on the collector machine

These ports must be reachable from your dashboard machine.

**firewalld (OpenSUSE / Fedora / RHEL):**

```bash
sudo firewall-cmd --zone=public --add-port=9090/tcp --permanent  # Prometheus
sudo firewall-cmd --zone=public --add-port=9200/tcp --permanent  # hwexp
sudo firewall-cmd --reload
```

**ufw (Ubuntu / Debian / Raspberry Pi OS):**

```bash
sudo ufw allow 9090/tcp   # Prometheus
sudo ufw allow 9200/tcp   # hwexp
```

Port 9100 (node-exporter) is scraped internally by Prometheus and does not need external access.

---

## Dashboard Stack — `monitoring/dashboard/`

Runs Grafana only. Queries any Prometheus instance via the `PROMETHEUS_URL` environment
variable — point it at the collector machine's Prometheus port.

### Dashboard `docker-compose.yml`

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
      - PROMETHEUS_URL=${PROMETHEUS_URL:-http://localhost:9090}

volumes:
  grafana-data:
```

### Dashboard environment variables (`.env.example`)

```env
# URL of Prometheus running in the collector stack.
# Change this to the IP or hostname of the machine running monitoring/collector/.
PROMETHEUS_URL=http://192.168.1.10:9090

# Grafana admin password (default: admin — change in production).
GF_ADMIN_PASSWORD=admin

# Anonymous read-only access.
# Set GF_AUTH_ANONYMOUS_ENABLED=false to require login for all users.
GF_AUTH_ANONYMOUS_ENABLED=true
GF_AUTH_ANONYMOUS_ORG_ROLE=Viewer
```

### Deploy the dashboard

```bash
cd monitoring/dashboard
cp .env.example .env
# Edit .env — set PROMETHEUS_URL to the collector machine's IP or hostname
docker compose up -d
```

### Open Grafana

Navigate to `http://<dashboard-host>:3000`.

Default login: **admin / admin** — you will be prompted to change the password on first login.

Three dashboards are pre-loaded in the **AUDiot** folder. Use the **host** variable at the top
of each dashboard to switch between monitored machines:

| Dashboard | Purpose |
| --- | --- |
| AUDiot System Overview | Comprehensive view — OS stats, all hardware sensors, network, disk |
| AUDiot Panel Display | Compact at-a-glance status board |
| AUDiot Discovery / Operations | Debug view for sensor bring-up and mapping inspection |

### Open firewall ports on the dashboard machine

```bash
# firewalld
sudo firewall-cmd --zone=public --add-port=3000/tcp --permanent && sudo firewall-cmd --reload
# ufw
sudo ufw allow 3000/tcp
```

---

## Deployment Scenarios

### Same machine — collector and dashboard together

```bash
cd monitoring/collector  && docker compose up -d
cd monitoring/dashboard  && docker compose up -d
# PROMETHEUS_URL defaults to http://localhost:9090
```

### Separate machines — dashboard querying multiple collectors

```text
PC-A (192.168.1.10) ── monitoring/collector/ ──┐
PC-B (192.168.1.11) ── monitoring/collector/ ──┤──► RPi ── monitoring/dashboard/
Server (192.168.1.12) ─ monitoring/collector/ ──┘       PROMETHEUS_URL=http://192.168.1.10:9090
```

To scrape all machines from a single Prometheus, add extra scrape targets to
`collector/config/prometheus/prometheus.yml`. The **host** variable in Grafana
will then let you filter and compare machines.

---

## How Auto-Mapping Works

On first start, `hwexp` reads every sensor in `/sys/class/hwmon`. Sensors without a
manual rule in `config/hwexp/mappings.yaml` receive an auto-generated rule within the first
5-second poll cycle. No manual configuration is needed for a fresh machine.

Auto-generated rules are written to `collector/config/hwexp/mappings.auto.yaml`:

```bash
cat monitoring/collector/config/hwexp/mappings.auto.yaml
```

Rules are grouped by device class and sensor type (e.g., one rule covers all AMD GPU temperature
sensors). To override a rule — rename a sensor, add thresholds, change labels — copy the rule
from `mappings.auto.yaml` into `mappings.yaml` and give it a higher `priority`. Manual rules
always take precedence over auto-generated ones.

To reload `mappings.yaml` after editing, without restarting the container:

```bash
docker kill --signal=SIGHUP hwexp
```

---

## Updates

```bash
cd monitoring/collector  && docker compose pull && docker compose up -d
cd monitoring/dashboard  && docker compose pull && docker compose up -d
```

Prometheus data and Grafana settings are stored in named Docker volumes and survive updates.

---

## Stopping / Removing

```bash
# Stop containers without removing stored data
docker compose down

# Stop and permanently remove all stored data (Prometheus history, Grafana settings)
docker compose down -v
```

---

## Designing Dashboards

The Grafana provisioning config sets `allowUiUpdates: true`. Edit any dashboard in the Grafana
UI, then export the JSON and save it back into `monitoring/dashboard/dashboards/` to make the
changes permanent.

All `hwexp` hardware metrics use the prefix `hw_device_` and carry a `host` label:

| Metric | Unit | What it covers |
|--------|------|----------------|
| `hw_device_temperature_celsius` | °C | CPU, GPU, NVMe, motherboard, PSU temperatures |
| `hw_device_fan_rpm` | RPM | All fans |
| `hw_device_power_watts` | W | GPU power draw, PSU output |
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
