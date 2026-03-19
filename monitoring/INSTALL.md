# AUDiot Monitor — Installation Guide

The monitoring system is split into two independent Docker Compose stacks.

| Stack | Purpose | Deploy on |
| --- | --- | --- |
| **Collector** | Scrapes hardware, OS, and AI metrics; stores them in Prometheus | Every machine you want to monitor |
| **Dashboard** | Runs Grafana; queries Prometheus to visualise the metrics | Any machine on the same network |

---

## Quick start — no repo clone needed

Install Docker, then run:

```bash
curl -fsSL https://raw.githubusercontent.com/Audumla/AUDiotMonitor/main/deploy.sh | bash
```

Grafana at **<http://localhost:3000>** (admin / admin).

---

## Collector Stack — `monitoring/collector/`

Runs on every machine you want to monitor.

### Collector `docker-compose.yml`

```yaml
services:
  hwexp:
    image: audumla/audiot-hwexp:latest
    container_name: hwexp
    privileged: true
    volumes:
      - /sys:/sys:ro
      - /proc:/proc:ro
      - ./config/hwexp:/etc/hwexp:z
      - ./config/hwexp/conf.d:/etc/hwexp/conf.d:ro    # Modular configs
      - ./custom.d:/etc/hwexp/custom.d:ro             # Custom scripts
    environment:
      - HWEXP_HOST=${HWEXP_HOST:-localhost}
    ports:
      - "9200:9200"
    restart: unless-stopped

  node-exporter:
    image: prom/node-exporter:latest
    # ... (rest of node-exporter)

  prometheus:
    image: prom/prometheus:latest
    # ... (rest of prometheus)
```

---

## Customization & Extensions

AUDiot is highly extensible without rebuilding the Docker image.

### 1. Modular Configuration (`conf.d`)
Mount YAML files into `/etc/hwexp/conf.d/` to override settings or enable new adapters.
Example `conf.d/llm.yaml`:
```yaml
adapters:
  llamaswap:
    enabled: true
    settings:
      endpoint: "http://192.168.1.50:50099"
```

### 2. Custom Scripts (`custom.d`)
Place any executable script in `/etc/hwexp/custom.d/`. It will be run every poll cycle.
Scripts must output JSON in the `RawMeasurement` format.
To support discovery, scripts should return a list of `DiscoveredDevice` when run with `--discover`.

---

## AI & LLM Monitoring

To monitor local LLM models running in **Llamaswap**:
1. Ensure Llamaswap is running with the OpenAI-compatible API enabled.
2. Enable the `llamaswap` adapter in your config (see example above).
3. Models will automatically appear in the **AI & LLM Services** section of the dashboard.

---

## Dashboards & Profiles

Dashboards are organized into **Profiles** based on screen size and use-case. These are automatically provisioned into matching folders in Grafana.

| Profile Folder | Contents | Target |
| --- | --- | --- |
| **Standard** | `System Overview`, `Dashboard [1080p]` | Desktop monitors. |
| **Wide-Screens** | `Panel [1920x440]` | Ultra-wide internal displays. |
| **Mobile** | `Panel [Portrait]` | Vertical side-screens/mobile. |
| **Debug** | `Discovery`, `Operations`, `Legacy Panel` | Troubleshooting & Bring-up. |

### Persisting Edits
By default, the dashboard stack uses Docker volumes to store your changes. To manage dashboards as files on your host machine:
1. Map a local directory to `/var/lib/grafana/dashboards` in `docker-compose.yml`.
2. On first run, AUDiot will populate this directory with the default profile folders.
3. Any changes you make to these JSON files (or via the Grafana UI) will be persisted locally.

**Optional: Skip Downloads**
If you want to prevent the bootstrapper from downloading default dashboards (e.g. if you already have your own set), set `SKIP_DASHBOARD_DOWNLOAD=true` in your environment.

---

## Verified Ports & Firewall

| Component | Port | Usage | Access |
| --- | --- | --- | --- |
| **Prometheus** | 9090 | Metrics API | From Dashboard machine |
| **hwexp** | 9200 | Hardware metrics | From Prometheus (localhost or external) |
| **Grafana** | 3000 | Web UI | From your browser |

**UFW (Ubuntu):** `sudo ufw allow 9090, 9200, 3000/tcp`

---

## Metrics Reference

| Metric | Purpose |
|--------|---------|
| `hw_device_info` | Metadata info metric (Vendor, Model, BIOS, etc) |
| `hw_device_temperature_celsius` | Thermal sensors |
| `hw_device_utilization_percent` | CPU/GPU/LLM load |
| `hw_device_capacity_bytes` | RAM/VRAM total capacity |
| `hw_device_sensor_count` | CPU Cores/Threads count |
| `hwexp_adapter_refresh_success` | Exporter health |
