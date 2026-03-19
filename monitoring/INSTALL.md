# AUDiot Monitor — Installation Guide

The monitoring system is split into two independent Docker Compose stacks that can run on the same machine or on different machines.

| Stack | Purpose | Deploy on |
| ----- | ------- | --------- |
| **Collector** | Scrapes hardware, OS, and AI metrics; stores them in Prometheus | Every machine you want to monitor |
| **Dashboard** | Runs Grafana; queries Prometheus to visualise the metrics | Any machine on the same network (including a Raspberry Pi) |

---

## Prerequisites

- Docker Engine 24+ and the `docker compose` plugin (v2)
- Linux host — the collector requires access to `/sys` and `/proc`

Install Docker on Debian / Ubuntu / Raspberry Pi OS:

```bash
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER   # log out and back in after this
```

---

## Collector Stack

Runs on **every machine you want to monitor**.

### Minimal quick-start (no config files needed)

```bash
mkdir -p ~/audiot/collector && cd ~/audiot/collector

# Label metrics with this machine's hostname
HWEXP_HOST=$(hostname) docker compose \
  -f https://raw.githubusercontent.com/Audumla/AUDiotMonitor/main/monitoring/collector/docker-compose.yml \
  up -d
```

Prometheus is now available at **<http://localhost:9090>** and hwexp at **<http://localhost:9200>**.

Browse to **<http://localhost:9200>** to see a live index of all available API endpoints.

### Full `docker-compose.yml` reference

See [`collector/docker-compose.yml`](collector/docker-compose.yml) for the complete file. Key environment variables:

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `HWEXP_HOST` | `$HOSTNAME` or `localhost` | Label applied to every metric as the `host` dimension |

### Custom config (optional)

Mount a config directory to override settings without rebuilding the image:

```bash
# Directory layout
collector/
  config/
    hwexp/
      hwexp.yaml         # main config overrides
      conf.d/
        llm.yaml         # example: enable llamaswap adapter
      mappings.yaml      # manual metric mapping rules
```

Example `conf.d/llm.yaml` to enable LLM monitoring:

```yaml
adapters:
  llamaswap:
    enabled: true
    settings:
      endpoint: "http://localhost:50099"
```

---

## Dashboard Stack

Runs **once** — typically on a machine with a screen, or on a dedicated display device such as a Raspberry Pi.

### Quick-start

```bash
mkdir -p ~/audiot/dashboard && cd ~/audiot/dashboard

PROMETHEUS_URL=http://<collector-ip>:9090 \
HWEXP_URL=http://<collector-ip>:9200 \
docker compose \
  -f https://raw.githubusercontent.com/Audumla/AUDiotMonitor/main/monitoring/dashboard/docker-compose.yml \
  up -d
```

Grafana opens at **<http://localhost:3000>** (default login: admin / admin).

### Environment variables

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `PROMETHEUS_URL` | `http://localhost:9090` | Prometheus endpoint Grafana queries for time-series data |
| `HWEXP_URL` | `http://localhost:9200` | hwexp endpoint used by the Infinity JSON datasource |
| `GF_ADMIN_PASSWORD` | `admin` | Grafana admin password |
| `GF_AUTH_ANONYMOUS_ENABLED` | `true` | Allow unauthenticated read-only access |
| `GF_AUTH_ANONYMOUS_ORG_ROLE` | `Viewer` | Role granted to anonymous users |
| `SKIP_DASHBOARD_DOWNLOAD` | `false` | Set `true` to skip downloading default dashboards on first run |

### What gets provisioned automatically

On first run `grafana-init` writes two Grafana datasources:

| Datasource | UID | Used by |
| ---------- | --- | ------- |
| Prometheus | *(auto)* | All time-series panels in the standard and debug dashboards |
| Infinity (hwexp) | `infinity-hwexp` | JSON table panels in the **State** dashboard |

Dashboard JSON files are downloaded from GitHub into four profile folders:

| Folder | Dashboards |
| ------ | ---------- |
| `profiles/standard` | System Overview, Dashboard [1080p] |
| `profiles/wide-screens` | Panel [1920×440] |
| `profiles/mobile` | Panel [Portrait] |
| `profiles/debug` | State (Live JSON), Discovery, Operations, Panel |

---

## Raspberry Pi (Debian) — Dashboard Stack

A Raspberry Pi 4 or 5 running Debian (bookworm/bullseye) makes an ideal always-on dashboard host. All images used support `linux/arm64` and `linux/arm/v7`.

### RPi-specific `docker-compose.yml`

Save this as `docker-compose.yml` in a working directory on the Pi, then run `docker compose up -d`:

```yaml
# AUDiot Monitor — Dashboard Stack (Raspberry Pi / ARM)
#
# Point at the collector running on your main machine:
#   PROMETHEUS_URL=http://192.168.1.x:9090
#   HWEXP_URL=http://192.168.1.x:9200
#   docker compose up -d
#
# First boot downloads plugins and dashboards — allow 2–3 minutes.

services:

  grafana-init:
    image: busybox
    environment:
      - PROMETHEUS_URL=${PROMETHEUS_URL:-http://localhost:9090}
      - HWEXP_URL=${HWEXP_URL:-http://localhost:9200}
      - SKIP_DASHBOARD_DOWNLOAD=${SKIP_DASHBOARD_DOWNLOAD:-false}
    command:
      - sh
      - -c
      - |
        set -e
        BASE="https://raw.githubusercontent.com/Audumla/AUDiotMonitor/main"

        mkdir -p /provisioning/datasources /provisioning/dashboards \
                 /dashboards/profiles/standard \
                 /dashboards/profiles/wide-screens \
                 /dashboards/profiles/mobile \
                 /dashboards/profiles/debug

        if [ ! -f /provisioning/datasources/prometheus.yaml ]; then
          cat > /provisioning/datasources/prometheus.yaml << EOF
        apiVersion: 1
        datasources:
          - name: Prometheus
            type: prometheus
            access: proxy
            url: ${PROMETHEUS_URL:-http://localhost:9090}
            isDefault: true
            editable: false
        EOF
        fi

        if [ ! -f /provisioning/datasources/infinity.yaml ]; then
          cat > /provisioning/datasources/infinity.yaml << EOF
        apiVersion: 1
        datasources:
          - name: Infinity (hwexp)
            type: yesoreyeram-infinity-datasource
            uid: infinity-hwexp
            access: proxy
            url: ${HWEXP_URL:-http://localhost:9200}
            isDefault: false
            editable: false
            jsonData:
              tlsSkipVerify: true
        EOF
        fi

        if [ ! -f /provisioning/dashboards/dashboards.yaml ]; then
          cat > /provisioning/dashboards/dashboards.yaml << 'EOF'
        apiVersion: 1
        providers:
          - name: 'AUDiot Profiles'
            orgId: 1
            folder: ''
            type: file
            disableDeletion: false
            editable: true
            allowUiUpdates: true
            options:
              path: /var/lib/grafana/dashboards
              foldersFromFilesStructure: true
        EOF
        fi

        if [ "$SKIP_DASHBOARD_DOWNLOAD" != "true" ]; then
          echo "Downloading dashboard profiles..."
          for d in system-overview panel-1080p; do
            [ ! -f "/dashboards/profiles/standard/${d}.json" ] && \
              wget -q -O "/dashboards/profiles/standard/${d}.json" \
                "${BASE}/monitoring/dashboard/dashboards/profiles/standard/${d}.json" || true
          done
          for d in panel-1920x440; do
            [ ! -f "/dashboards/profiles/wide-screens/${d}.json" ] && \
              wget -q -O "/dashboards/profiles/wide-screens/${d}.json" \
                "${BASE}/monitoring/dashboard/dashboards/profiles/wide-screens/${d}.json" || true
          done
          for d in panel-portrait; do
            [ ! -f "/dashboards/profiles/mobile/${d}.json" ] && \
              wget -q -O "/dashboards/profiles/mobile/${d}.json" \
                "${BASE}/monitoring/dashboard/dashboards/profiles/mobile/${d}.json" || true
          done
          for d in discovery-dashboard operations-dashboard panel-dashboard state-dashboard; do
            [ ! -f "/dashboards/profiles/debug/${d}.json" ] && \
              wget -q -O "/dashboards/profiles/debug/${d}.json" \
                "${BASE}/monitoring/dashboard/dashboards/profiles/debug/${d}.json" || true
          done
        fi

        echo "Grafana init complete"
    volumes:
      - grafana-provisioning:/provisioning
      - grafana-dashboards:/dashboards

  grafana:
    image: grafana/grafana:latest   # multi-arch: supports arm64 and arm/v7
    container_name: audiot-grafana
    depends_on:
      grafana-init:
        condition: service_completed_successfully
    ports:
      - "3000:3000"
    volumes:
      - grafana-provisioning:/etc/grafana/provisioning:ro
      - grafana-dashboards:/var/lib/grafana/dashboards:ro
      - grafana-data:/var/lib/grafana
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=${GF_AUTH_ANONYMOUS_ENABLED:-true}
      - GF_AUTH_ANONYMOUS_ORG_ROLE=${GF_AUTH_ANONYMOUS_ORG_ROLE:-Viewer}
      - GF_SECURITY_ADMIN_PASSWORD=${GF_ADMIN_PASSWORD:-admin}
      - GF_INSTALL_PLUGINS=yesoreyeram-infinity-datasource
      # Reduce memory usage on RAM-constrained devices
      - GF_DATABASE_WAL=true
      - GF_ANALYTICS_REPORTING_ENABLED=false
      - GF_ANALYTICS_CHECK_FOR_UPDATES=false
      - GF_ANALYTICS_CHECK_FOR_PLUGIN_UPDATES=false
    # Resource limits — tune for your Pi model:
    #   RPi 5 (4 GB+): remove limits entirely
    #   RPi 4 (4 GB):  mem_limit 512m is comfortable
    #   RPi 4 (2 GB):  mem_limit 384m; close other services
    #   RPi 3 / Zero:  not recommended for Grafana
    deploy:
      resources:
        limits:
          memory: 512m
        reservations:
          memory: 128m
    restart: unless-stopped

volumes:
  grafana-provisioning:
  grafana-dashboards:
  grafana-data:
```

### Notes for Raspberry Pi

#### First-boot plugin install

`GF_INSTALL_PLUGINS` causes Grafana to download and install the Infinity plugin on first start.
On a Pi with a slow SD card this may take 2–3 minutes before the UI is responsive.
Subsequent restarts are fast because the plugin is persisted in the `grafana-data` volume.

### SD card longevity

Prometheus write-ahead logs generate significant I/O. Run the **collector** stack on the machine you're monitoring (typically x86), not on the Pi. The Pi only needs Grafana.

### Kiosk mode — auto resolution

`kiosk.sh` detects the connected screen resolution and automatically opens the best-matching dashboard profile in Chromium fullscreen mode, then restarts Chromium if it exits.

| Resolution | Dashboard |
| ---------- | --------- |
| Width ≥ 1800, height ≤ 500 | Panel [1920×440] — ultra-wide ticker strip |
| Portrait (height > width) or width ≤ 800 | Panel [Portrait] |
| Width ≥ 1920, height ≥ 1000 | Dashboard [1080p] |
| Everything else | System Overview |

**One-shot installer** — run as the display user (not root):

```bash
cd /opt/docker/dashboard

# Point at Grafana (default: http://localhost:3000)
GRAFANA_URL=http://localhost:3000 ./kiosk-install.sh
```

The installer:

1. Installs `chromium-browser` if missing
2. Registers `kiosk.sh` as a **systemd user service** (Debian bookworm) or **XDG autostart entry** (LXDE / other desktops)
3. Starts the kiosk immediately

Override the auto-selected dashboard with `KIOSK_DASHBOARD=audiot-system-overview` if you want a fixed layout regardless of screen size.

**Manual launch** (without installing):

```bash
GRAFANA_URL=http://localhost:3000 ./kiosk.sh
```

---

## Running Collector and Dashboard on the Same Machine

If you want everything on one host (typical for a single server setup):

```bash
cd monitoring/collector && docker compose up -d
cd ../dashboard && PROMETHEUS_URL=http://localhost:9090 HWEXP_URL=http://localhost:9200 docker compose up -d
```

Both stacks use their own Docker networks by default. The dashboard stack reaches the collector via the host's published ports (`9090`, `9200`).

---

## Customization

### Modular configuration (`conf.d`)

Drop `.yaml` files into a `conf.d/` directory and mount it:

```yaml
volumes:
  - ./config/hwexp/conf.d:/etc/hwexp/conf.d:ro
```

Example `conf.d/llm.yaml`:

```yaml
adapters:
  llamaswap:
    enabled: true
    settings:
      endpoint: "http://192.168.1.50:50099"
```

### Custom scripts (`custom.d`)

Place any executable script in a `custom.d/` directory and mount it:

```yaml
volumes:
  - ./custom.d:/etc/hwexp/custom.d:ro
```

Scripts are executed every poll cycle. They must write a JSON array of `RawMeasurement` objects to stdout. Run with `--discover` to emit `DiscoveredDevice` objects instead.

---

## Persisting Dashboard Edits

By default dashboards are stored in a Docker volume (`grafana-dashboards`). To manage them as plain files on the host:

```yaml
volumes:
  - ./my-dashboards:/var/lib/grafana/dashboards   # replace the named volume
```

Set `SKIP_DASHBOARD_DOWNLOAD=true` after the first run to stop the bootstrapper from overwriting local edits with the defaults.

---

## Ports & Firewall

| Service | Port | Needed by |
| ------- | ---- | --------- |
| hwexp | `9200` | Prometheus (scrape), browser (debug UI) |
| node-exporter | `9100` | Prometheus (scrape, host network only) |
| Prometheus | `9090` | Grafana, browser |
| Grafana | `3000` | Browser |

Open the required ports with UFW (Debian/Ubuntu):

```bash
sudo ufw allow 9090/tcp   # Prometheus
sudo ufw allow 9200/tcp   # hwexp
sudo ufw allow 3000/tcp   # Grafana
```

---

## API Reference

Browse to **`http://<collector-ip>:9200`** for a live HTML index of all endpoints.

| Endpoint | Description |
| -------- | ----------- |
| `GET /` | HTML index — lists all endpoints with auth requirements |
| `GET /metrics` | Prometheus text format |
| `GET /healthz` | Liveness probe |
| `GET /readyz` | Readiness probe (200 once first poll completes) |
| `GET /version` | Version and schema strings |
| `GET /debug/state` | All devices joined with current measurements (JSON) — primary feed for Grafana Infinity |
| `GET /debug/catalog` | Flat list of all normalised measurements |
| `GET /debug/discovery` | Discovered device inventory |
| `GET /debug/mappings` | Mapping rule decisions |
| `GET /debug/raw` | Raw adapter output (requires `enable_raw_endpoint: true` in config) |

---

## Metrics Reference

| Metric | Type | Description |
| ------ | ---- | ----------- |
| `hw_device_info` | Info | Vendor, Model, BIOS version, Driver, NIC model |
| `hw_device_temperature_celsius` | Gauge | Thermal sensors |
| `hw_device_utilization_percent` | Gauge | CPU / GPU / LLM load |
| `hw_device_capacity_bytes` | Info | RAM / VRAM total capacity |
| `hw_device_sensor_count` | Info | CPU core / thread counts |
| `hwexp_adapter_refresh_success` | Gauge | `1` if last poll succeeded |
| `hwexp_adapter_refresh_duration_seconds` | Gauge | Duration of last poll cycle |
| `hwexp_adapter_last_success_unixtime` | Gauge | Unix timestamp of last successful poll |
| `hwexp_discovered_devices` | Gauge | Number of discovered devices |
| `hwexp_mapping_failures_total` | Counter | Cumulative mapping failures |
| `node_rapl_*_joules_total` | Counter | Intel RAPL power domains (rate → watts) |
| `node_systemd_unit_state` | Gauge | Systemd unit states (active / inactive / failed) |
