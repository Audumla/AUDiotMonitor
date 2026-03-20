The monitoring system is split into two independent Docker Compose stacks that can run on the same machine or on different machines.

---

## Remote Deployment

The recommended way to manage the stacks is from a local repo checkout using the `deploy-remote.sh` tool.

**Requirements:**
- Linux or macOS local machine (or Windows with Git Bash / WSL)
- `ssh` and `rsync` installed locally
- SSH access to the target host (e.g., `ssh pi@<rpi-ip>` works)

```bash
cd monitoring
./deploy-remote.sh <host> <collector|dashboard> [target_dir]
```

Default install targets:

- collector: `/opt/docker/services/monitoring`
- dashboard: `/opt/docker/services/dashboard`

Typical example:

```bash
cd monitoring
./deploy-remote.sh buri collector
./deploy-remote.sh brutusview dashboard
```

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

### Minimal quick-start (ephemeral config)

```bash
mkdir -p ~/audiot/collector && cd ~/audiot/collector

# Label metrics with this machine's hostname
HWEXP_HOST=$(hostname) docker compose \
  -f https://raw.githubusercontent.com/Audumla/AUDiotMonitor/main/monitoring/collector/docker-compose.yml \
  up -d
```

Prometheus is now available at **<http://localhost:9090>** and hwexp at **<http://localhost:9200>**.

Browse to **<http://localhost:9200>** to see a live index of all available API endpoints.

### Recommended host-owned layout

This is the preferred install if you want editable Prometheus rules and
persistent system-specific config. The recommended way to deploy is using the
remote deployment tool from your local repo:

```bash
# From your local checkout of the AUDiot repo:
cd monitoring
./deploy-remote.sh <collector-host> collector
```

This installs the collector under:

```text
/opt/docker/services/monitoring/
  docker-compose.yml
  .env
  config/
    hwexp/
      hwexp.yaml
      mappings.yaml
    prometheus/
      prometheus.yml
      rules/
  hwexp/
  prometheus/
```

Common collector operations:

```bash
cd /opt/docker/services/monitoring
./manage-collector.sh validate
./manage-collector.sh verify-metrics
./manage-collector.sh generate-rules --force
./manage-collector.sh restart-prometheus
./manage-collector.sh restart-hwexp
./manage-collector.sh status
```

### Full `docker-compose.yml` reference

See [`collector/docker-compose.yml`](collector/docker-compose.yml) for the complete file. Key environment variables:

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `HWEXP_HOST` | `$HOSTNAME` or `localhost` | Label applied to every metric as the `host` dimension |

### Custom config and recording rules

The collector compose file mounts config directly from the host:

```bash
# Directory layout
collector/
  config/
    hwexp/
      hwexp.yaml         # main config overrides
      mappings.yaml      # manual metric mapping rules
    prometheus/
      prometheus.yml     # scrape + rule_files config
      rules/
        defaults/
          audiot-recording-rules.yml
        custom/
          system.rules.yml
```

Default Prometheus recording rules are shipped in `rules/defaults/` and create
reusable synthetic metrics such as:

- `audiot_gpu_compute_utilization_percent`
- `audiot_gpu_memory_utilization_percent`
- `audiot_gpu_vram_used_bytes`
- `audiot_gpu_vram_capacity_bytes`
- `audiot_gpu_vram_usage_percent`

Host-specific aliases live in `rules/custom/system.rules.yml`. A generator writes
that file once from the detected DRM cards so you get stable labels like
`gpu_index="1"` without hard-coding PCI IDs in dashboards.

Re-generate the custom rules file manually:

```bash
cd /opt/docker/services/monitoring
./manage-collector.sh generate-rules --force
./manage-collector.sh restart-prometheus
```

Example `hwexp.yaml` override to enable LLM monitoring:

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

### RPi / bind-mounted layout

Use the dashboard installer to scaffold a simple host-owned layout. The recommended
way to deploy is using the remote deployment tool from your local repo:

```bash
# From your local checkout of the AUDiot repo:
cd monitoring
./deploy-remote.sh <rpi-ip> dashboard
```

This tool:
1. Syncs the `monitoring/dashboard/` directory to `/opt/docker/services/dashboard` on the RPi.
2. Runs the local installer to set up the layout.

The host-owned layout structure:

```text
/opt/docker/services/dashboard/
  docker-compose.yml
  .env
  config/
    grafana/
      grafana.ini
      provisioning/
    kiosk.env
  dashboards/
    profiles/    # shipped defaults
    custom/      # your host-local dashboards
```

Then set the collector endpoints in `/opt/docker/services/dashboard/.env` or export them
at launch.

### Managing the Dashboard

The dashboard is now managed by a single, powerful script installed globally as `audiot-dashboard`.

Common dashboard operations:

```bash
audiot-dashboard status
audiot-dashboard restart
audiot-dashboard logs
audiot-dashboard set <uid>
audiot-dashboard update
```

### Notes for Raspberry Pi

#### First-boot plugin install

If Grafana fails to load Infinity plugins on a fresh Pi install (common if networking isn't fully up yet), just restart the stack:

```bash
audiot-dashboard restart
```

#### Wayland vs X11

Raspberry Pi OS Bookworm uses **Wayland** (`wayfire` / `labwc`) by default.
X11 utilities like `xrandr` and `xset` do not work directly.

The `audiot-dashboard` installer and kiosk launcher automatically handle the differences between X11 and Wayland for resolution detection and autostart.

### Kiosk mode — auto resolution

The kiosk launcher detects the connected screen resolution and picks a dashboard using `config/kiosk.env`, then restarts Chromium if it exits.

Edit `/opt/docker/services/dashboard/config/kiosk.env` to control:
* `KIOSK_DASHBOARD=` (Force one UID for all screens)
* `KIOSK_IDLE_TIMEOUT=` (Idle timeout in seconds before screen blanks. Useful for touchscreens. Default: always on)
* `KIOSK_DPMS_STANDBY=`, `KIOSK_DPMS_SUSPEND=`, `KIOSK_DPMS_OFF=` (Advanced staggered dimming/sleep states)

To add your own dashboards, drop JSON files into `/opt/docker/services/dashboard/dashboards/custom/` and set the desired UID in `/opt/docker/services/dashboard/config/kiosk.env` using `audiot-dashboard set <uid>`.

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

Dashboards are plain files on the host in `/opt/docker/services/dashboard/dashboards/`.
Default shipped profiles live under `profiles/`; your machine-specific dashboards should go under `custom/`.

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
