# AUDiotMonitor Collector v2

A unified monitoring stack: **Netdata** + **Prometheus** + **llamaswap-exporter** + **Grafana**.

## Stack

| Service | Port | Purpose |
|---------|------|---------|
| Netdata | `:19999` | Real-time hardware/OS metrics (CPU, RAM, GPU, disk, temps, fans, network) |
| Prometheus | `:9090` | Metrics storage — scrapes Netdata + llamaswap-exporter |
| llamaswap-exporter | `:9300` | Custom exporter — scrapes llama-swap API, exposes model metrics to Prometheus |
| Grafana | `:3000` | Dashboard visualization |

## Quick Start

```bash
cd monitoring/collector-v2

# Copy and edit env file
cp .env.example .env

# Start everything
./manage.sh up
```

## URLs

- **Netdata UI** — http://localhost:19999
- **Prometheus** — http://localhost:9090
- **Grafana** — http://localhost:3000 (admin / admin)
- **llamaswap-exporter metrics** — http://localhost:9300/metrics

## Management

```bash
./manage.sh up          # Start all services
./manage.sh down        # Stop all services
./manage.sh restart     # Restart all services
./manage.sh logs [svc]  # View logs (svc optional)
./manage.sh status      # Show container status
./manage.sh validate    # Validate docker-compose.yml
./manage.sh update      # Pull latest images and restart
```

## Configuration

### Llama-swap

Edit `.env` to point at your llama-swap instance:

```bash
LLAMASWAP_URL=http://localhost:41080
LLAMASWAP_POLL_INTERVAL=15
```

### Prometheus Retention

```bash
PROMETHEUS_RETENTION=30d    # 30 days of data
```

### Ports

All ports are configurable via `.env`:

```bash
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
LLAMASWAP_EXPORTER_PORT=9300
```

## Metrics

### Netdata (scraped by Prometheus)

Netdata exposes 1000+ metrics including:
- `netdata_cpu_cpu_percentage_usage` — CPU usage per core
- `netdata_mem_available_bytes` — Available RAM
- `netdata_disk_read_B` / `netdata_disk_write_B` — Disk I/O
- `netdata_systemd_unit_state` — Service states
- `netdata_hdd_temp_celsius` / `netdata_ssd_temp_celsius` — Drive temps
- `netdata_cpu_temp_celsius` — CPU temperature
- `netdata_fan_rpm` — Fan speeds
- `netdata_gpu_*` — GPU metrics (NVIDIA NVML, AMD rocm)
- `netdata_smart_health_status` — Disk SMART health

### llamaswap-exporter

- `llamaswap_models_total` — Total models loaded
- `llamaswap_model_slots{model_id, model_name}` — GPU slots used
- `llamaswap_model_queue_length{model_id, model_name}` — Request queue
- `llamaswap_model_active{model_id, model_name}` — Currently processing
- `llamaswap_model_backend{model_id, model_name, backend}` — Backend type
- `llamaswap_model_backend_port{model_id, model_name}` — Backend port
- `llamaswap_health` — API reachability (1=up, 0=down)
- `llamaswap_scrape_errors_total` — Scrape error count
- `llamaswap_last_scrape_duration_seconds` — Scrape duration

## Dashboard Profiles

Dashboard JSON files go in `./dashboards/`. They are auto-loaded by Grafana via the provisioning config.

## Remote Deployment

The stack deploys to `/srv/docker/monitor/` on the remote server using the project's standard docker service structure:

```
/srv/docker/monitor/
├── compose.yaml              # Main compose file
├── .env.example              # Environment template
├── .env                      # (created from .env.example)
├── config/
│   ├── prometheus/
│   │   └── prometheus.yml    # Prometheus scrape config
│   └── grafana/
│       └── provisioning/     # Grafana datasource/dashboard provisioning
└── llamaswap-exporter/
    ├── app.py                # Flask exporter
    └── Dockerfile            # Build context
```

### Deploy steps

```bash
# 1. Copy local files to remote
scp -r monitoring/collector-v2/* audumla@10.10.100.10:/srv/docker/monitor/

# 2. Create .env from template
ssh audumla@10.10.100.10 "cp /srv/docker/monitor/.env.example /srv/docker/monitor/.env"

# 3. Start using remote dku script
ssh audumla@10.10.100.10 "cd /srv/docker/monitor && dku -u monitor"
```

### Remote management

```bash
# Start/update
ssh audumla@10.10.100.10 "dku monitor"

# Stop
ssh audumla@10.10.100.10 "dks monitor"

# Update images and restart
ssh audumla@10.10.100.10 "dku -u monitor"
```
