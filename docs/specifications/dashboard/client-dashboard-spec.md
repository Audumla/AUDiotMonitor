# Hardware Telemetry Platform

## Raspberry Pi 3 Debian Docker Dashboard Node Specification

**Document purpose**: This document defines the detailed technical specification for a Docker-based dashboard node running on a Raspberry Pi 3 under Debian. This node provides the local dashboard/UI tier, metrics collection tier, and optional viewer/kiosk tier for the hardware telemetry platform, while connecting to one or more remote backend data providers.

**Status**: Normative for implementation

**Audience**: Developers implementing the Raspberry Pi deployment, operators installing the platform on Debian-based Pi systems, and developers building Docker images, Compose stacks, and provisioning/configuration automation

---

# 1. Scope

This document defines:

* the Raspberry Pi 3 dashboard-node role for Debian
* the Docker/Compose deployment model
* required and optional services on the Pi
* service boundaries and network model
* configuration parameters needed to connect to remote backend/exporter providers
* persistent storage layout
* provisioning and startup behavior
* kiosk/display considerations
* upgrade, rollback, and recovery considerations specific to the Pi node

This document does not define source-host exporter internals, except where integration points are required.

---

# 2. Goals

The Raspberry Pi 3 Debian deployment MUST provide:

* a self-contained Docker-based dashboard node
* Grafana for local and remote viewing
* Prometheus for scraping one or more remote backend/exporter endpoints
* configuration-driven target definition for remote data providers
* support for panel, ops, and discovery dashboards
* optional kiosk/browser launcher support for the attached panel display
* low operational complexity on a Pi 3
* compatibility with later migration to a more capable dashboard host

The deployment SHOULD:

* remain simple enough for a Pi 3
* avoid excessive container count where unnecessary
* allow future optional services without reworking the whole layout
* support updates through the lifecycle/install system later

---

# 3. Role definition

This spec defines a deployment role named:

* `embedded-pi-dashboard-node`

This role is a specialization of:

* `embedded-pi-node`
* `dashboard-node`

It assumes:

* Raspberry Pi 3 hardware
* Debian-based OS
* Docker Engine and Docker Compose plugin available or installed by setup workflow
* one or more remote exporters/backends accessible over the local network

This role MAY also act as:

* local kiosk display host
* local Grafana server for other LAN viewers
* small central metrics node for a modest number of hosts

---

# 4. Deployment model

## 4.1 Preferred model

The preferred Pi 3 deployment model is:

* Docker Compose with one container per major service
* bind-mounted config and provisioning files
* named volumes for mutable service data
* host-level browser/kiosk process outside containers, unless explicitly containerized later

## 4.2 Why Compose over one monolithic image

The default implementation MUST use **Docker Compose with separate containers**, not a single all-in-one image.

Reasons:

* clearer separation of concerns
* easier upgrades per component
* easier troubleshooting on a low-power device
* less coupling between Grafana, Prometheus, and optional helpers
* easier future migration to non-Pi dashboard host

## 4.3 Monolithic image option

A monolithic image MAY be supported later for appliance-style deployment, but it is NOT the normative v1 deployment model.

If a monolithic image is introduced later, it MUST still preserve the same effective config model and directory layout semantics where practical.

---

# 5. Required services

The Pi dashboard node MUST include these services.

## 5.1 `grafana`

Purpose:

* serve dashboards
* support LAN browser access
* support attached display kiosk mode
* provide dashboard provisioning for panel, ops, and discovery dashboards

Responsibilities:

* load provisioned datasources
* load provisioned dashboard bundles
* persist Grafana DB/state in a volume

## 5.2 `prometheus`

Purpose:

* scrape one or more remote backend/exporter endpoints
* retain short-to-moderate local time-series history
* serve Grafana datasource needs

Responsibilities:

* load scrape targets from generated or mounted config
* maintain local TSDB
* expose local Prometheus web UI for debugging if allowed

## 5.3 `provisioner-init` or equivalent bootstrap step

Purpose:

* render or validate config files before services start
* prepare environment-specific target files and dashboard provisioning

This MAY be implemented as:

* a one-shot init container
* a host-side generation script run before `docker compose up`
* a lifecycle-manager-generated config phase

Normative requirement:

* some bootstrap mechanism MUST exist to render runtime config from operator parameters

---

# 6. Optional services

These services MAY be supported, but are not mandatory for v1.

## 6.1 `grafana-image-renderer`

Optional.

Use only if image rendering is actually needed.
On a Pi 3, this SHOULD be omitted by default due to resource constraints.

## 6.2 `nginx` or reverse proxy

Optional.

Use only if needed for:

* TLS termination
* auth fronting
* path routing
* nicer local hostnames

By default, this SHOULD be omitted on Pi 3 unless specific requirements justify it.

## 6.3 `watchdog-sidecar`

Optional.

Could later provide:

* health monitoring
* Compose stack restart heuristics
* display/session monitoring

Not required for v1.

## 6.4 `dashboard-config-generator`

Optional dedicated service.

This may later generate:

* Prometheus target files
* Grafana provisioning
* panel-specific dashboard variants

For v1, a host-side script or one-shot init container is acceptable.

---

# 7. Non-containerized host responsibilities

The following responsibilities SHOULD remain on the host OS rather than inside containers in v1.

## 7.1 Kiosk browser session

The browser used to display Grafana on the attached panel SHOULD run on the host OS.

Reasons:

* easier access to local display/graphics stack
* simpler startup on Raspberry Pi desktop/lightweight session
* avoids extra container GUI complexity

## 7.2 Docker and Compose runtime

Docker Engine and Docker Compose plugin live on the host OS.

## 7.3 Optional setup helper scripts

Host-side helper scripts MAY be used for:

* first-time environment file generation
* kiosk autostart setup
* display resolution helpers
* config validation before Compose startup

---

# 8. Network model

## 8.1 Local service access

By default, the Pi node MUST expose:

* Grafana HTTP port on LAN or localhost depending on configuration
* Prometheus HTTP port on LAN or localhost depending on configuration

Default recommended exposure on trusted LAN:

* Grafana exposed on LAN
* Prometheus exposed on localhost or LAN depending on operator preference

## 8.2 Remote backend/exporter connectivity

The Pi node MUST be able to connect to one or more remote exporter/backend endpoints on the LAN.

Examples:

* source-host exporter at `http://10.10.100.10:9200/metrics`
* node_exporter at `http://10.10.100.10:9100/metrics`
* custom ingest exporter at `http://10.10.100.10:9300/metrics`
* LLM observer exporter at `http://10.10.100.10:9400/metrics`

## 8.3 Prometheus scrape model

Prometheus on the Pi MUST pull metrics from configured remote targets.
It MUST NOT require push from remote exporters.

## 8.4 Docker networking

The Compose stack SHOULD use a dedicated bridge network for internal service communication.

Suggested network:

* `hwexp_net`

Services within the stack communicate internally via service names.
External/LAN access is provided through published ports.

---

# 9. Configuration model

The Pi node MUST be configurable through a combination of:

* `.env` file or equivalent environment file
* mounted static config templates
* rendered runtime config files

## 9.1 Required operator parameters

At minimum, the deployment MUST support these parameters.

### Identity and environment

* `HWEXP_NODE_NAME`
* `HWEXP_SITE`
* `HWEXP_ENVIRONMENT`
* `TZ`

### Grafana

* `GRAFANA_HTTP_PORT`
* `GRAFANA_ADMIN_USER`
* `GRAFANA_ADMIN_PASSWORD`
* `GRAFANA_ALLOW_ANONYMOUS` (optional, default false)

### Prometheus

* `PROMETHEUS_HTTP_PORT`
* `PROMETHEUS_RETENTION_TIME`
* `PROMETHEUS_SCRAPE_INTERVAL`
* `PROMETHEUS_EVALUATION_INTERVAL`

### Remote backend target definition

One or more of the following models MUST be supported:

* static target list file
* environment-variable target list
* rendered YAML target file from template

At minimum, support MUST exist for:

* `REMOTE_EXPORTER_TARGETS_FILE`

### Kiosk/display parameters

* `ENABLE_KIOSK`
* `KIOSK_GRAFANA_URL`
* `KIOSK_BROWSER`
* `KIOSK_FULLSCREEN`

## 9.2 Recommended remote target file model

The preferred v1 model is a mounted YAML or JSON file describing scrape targets.

Example logical file:

```yaml
remote_targets:
  - job_name: hwexp_exporters
    targets:
      - http://10.10.100.10:9200/metrics
      - http://10.10.100.11:9200/metrics
  - job_name: node_exporters
    targets:
      - http://10.10.100.10:9100/metrics
```

A config-generation step MUST transform this into valid Prometheus scrape config.

## 9.3 Optional auth parameters

If remote exporters require auth later, config MUST allow for:

* bearer token file mounts
* basic auth credentials if explicitly supported
* TLS CA/cert configuration if needed

These MAY be deferred for v1 if the LAN deployment assumes open scrape endpoints.

---

# 10. Persistent storage model

The Pi node MUST distinguish between:

* immutable application/config templates
* mutable service data
* operator-editable runtime config

## 10.1 Required persistence areas

### Grafana data

Must persist:

* Grafana DB/state
* user settings if not fully stateless

Suggested volume:

* `grafana_data`

### Prometheus data

Must persist:

* TSDB data

Suggested volume:

* `prometheus_data`

### Runtime config/provisioning

Must persist in host-mounted directory:

* Prometheus rendered config
* Grafana provisioning files
* dashboard bundles or mounted dashboard directories
* target definition files
* operator overrides

Suggested host path structure:

```text
/opt/hwexp/rpi-dashboard/
  .env
  compose/
  config/
    prometheus/
    grafana/
    targets/
  dashboards/
    panel/
    ops/
    discovery/
  scripts/
  data/
```

---

# 11. Grafana provisioning model

## 11.1 Datasource provisioning

Grafana MUST be provisioned with Prometheus as a datasource automatically.

The datasource definition MUST be mounted/read from provisioning files, not created manually in the UI for baseline functionality.

## 11.2 Dashboard provisioning

Grafana MUST support provisioned dashboards for:

* panel dashboard
* ops dashboard
* discovery dashboard

## 11.3 Dashboard bundles

Dashboard bundles SHOULD be organized by folder or directory:

* `panel`
* `ops`
* `discovery`

## 11.4 Pi-specific panel defaults

The panel dashboard SHOULD have a Pi/panel-optimized default version for 1920×440 display use.

---

# 12. Prometheus configuration model

## 12.1 Prometheus config generation

Prometheus config SHOULD be generated from:

* base template
* operator parameters
* remote target list file

This avoids hard-coding targets in the Compose file.

## 12.2 Required scrape jobs

At minimum, Prometheus MUST support jobs for:

* `hwexp_exporter`
* `node_exporter` (optional but recommended)
* `custom_ingest` if present remotely
* `llm_observer` if present remotely
* self-scrape for Prometheus itself

## 12.3 Resource-conscious defaults

Given Pi 3 constraints, default Prometheus values SHOULD be conservative.

Recommended starting defaults:

* scrape interval: `15s`
* evaluation interval: `15s`
* retention: `3d` to `7d` depending on expected target count and storage

The exact shipped defaults MAY vary, but the default spec MUST remain conservative for Pi 3.

---

# 13. Kiosk/display model

## 13.1 Kiosk approach

The kiosk/display layer SHOULD launch a browser on the host OS pointing to local Grafana.

Example target URL pattern:

* `http://127.0.0.1:${GRAFANA_HTTP_PORT}/d/<dashboard_id>/<slug>?kiosk`

## 13.2 Browser choice

The deployment SHOULD support configurable browser choice.

Recommended options:

* Chromium
* Firefox ESR if available and suitable

## 13.3 Autostart

The Pi deployment SHOULD include an optional kiosk autostart helper.

This MAY be implemented later via:

* desktop autostart entry
* systemd user service
* lightweight X session start script

This is outside the core Docker stack but MUST be allowed for by config.

---

# 14. Docker Compose requirements

## 14.1 Compose file

A Compose file MUST be provided.

Canonical file name recommendation:

* `docker-compose.yml`

## 14.2 Required service definitions

The Compose file MUST define at least:

* `grafana`
* `prometheus`

Optional:

* `provisioner-init`

## 14.3 Image architecture compatibility

All images selected for the Pi deployment MUST support ARMv7 / Raspberry Pi 3-compatible architecture, or the deployment instructions MUST clearly declare a stronger requirement if that changes.

## 14.4 Restart policy

The required services SHOULD use restart policies suitable for appliance-style operation.

Recommended default:

* `unless-stopped`

## 14.5 Health checks

Service health checks SHOULD be defined where practical.

At minimum:

* Grafana HTTP health check or port check
* Prometheus readiness/HTTP check

---

# 15. Example Compose structure

The exact file content may be implemented separately, but the logical structure SHOULD resemble:

```yaml
services:
  prometheus:
    image: <prometheus-image>
    container_name: hwexp-prometheus
    restart: unless-stopped
    networks: [hwexp_net]
    ports:
      - "${PROMETHEUS_HTTP_PORT}:9090"
    volumes:
      - prometheus_data:/prometheus
      - ./config/prometheus:/etc/prometheus
      - ./config/targets:/etc/hwexp-targets

  grafana:
    image: <grafana-image>
    container_name: hwexp-grafana
    restart: unless-stopped
    networks: [hwexp_net]
    ports:
      - "${GRAFANA_HTTP_PORT}:3000"
    environment:
      - GF_SECURITY_ADMIN_USER=${GRAFANA_ADMIN_USER}
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
    volumes:
      - grafana_data:/var/lib/grafana
      - ./config/grafana/provisioning:/etc/grafana/provisioning
      - ./dashboards:/var/lib/grafana/dashboards

networks:
  hwexp_net:

volumes:
  grafana_data:
  prometheus_data:
```

This is illustrative only. Final implementation artifacts may refine paths and environment handling.

---

# 16. Bootstrap/init model

## 16.1 Purpose

The deployment needs a bootstrap phase to:

* prepare `.env`
* render Prometheus scrape config from remote target definitions
* validate Grafana provisioning files
* ensure required directories exist

## 16.2 Acceptable implementations

Any one of these is acceptable for v1:

* host-side setup script that runs before Compose startup
* one-shot init container run explicitly
* lifecycle-tool-generated config phase

## 16.3 Validation duties

Bootstrap MUST validate:

* required parameters present
* target file syntactically valid
* generated Prometheus config syntactically valid
* Grafana provisioning paths exist

If validation fails, `docker compose up` SHOULD NOT be the first place failure is discovered.

---

# 17. Security model for the Pi node

## 17.1 Default trust model

This deployment is intended primarily for trusted LAN environments.

## 17.2 Defaults

Default behavior SHOULD be:

* Grafana exposed on LAN with authenticated access
* Prometheus exposed on localhost or LAN depending on operator preference
* no raw exporter secrets baked into images
* secret values provided via env files or mounted secret files

## 17.3 Admin credentials

Grafana admin credentials MUST be configurable externally and MUST NOT be hard-coded into the image.

## 17.4 Future hardening

The design SHOULD allow later support for:

* reverse proxy TLS termination
* non-default Grafana auth hardening
* secret files rather than plaintext `.env` for sensitive values

---

# 18. Resource constraints and Pi 3 guidance

## 18.1 General rule

The spec MUST assume Raspberry Pi 3 is resource-constrained.

## 18.2 Required design implications

The deployment SHOULD:

* keep container count low
* avoid unnecessary heavy sidecars
* use conservative Prometheus retention defaults
* avoid image-rendering extras by default
* prefer simple dashboard provisioning

## 18.3 Dashboard guidance

The Pi-attached panel dashboard SHOULD remain lightweight.
The richer ops/discovery dashboards may still be hosted by Grafana, but should mainly be consumed from more capable clients when needed.

---

# 19. Update and lifecycle considerations

## 19.1 Compatibility with lifecycle manager

This deployment spec MUST be compatible with the lifecycle/install system defined elsewhere.

The lifecycle system SHOULD later be able to:

* install Docker/Compose prerequisites if needed
* place Compose files and config
* place dashboard bundles
* manage updates for dashboard node artifacts
* preserve data volumes across updates

## 19.2 Update boundaries

Updates SHOULD distinguish between:

* service image updates
* dashboard bundle updates
* provisioning/config updates
* kiosk helper updates

## 19.3 Rollback considerations

Rollback SHOULD preserve:

* Grafana data volume
* Prometheus data volume where compatible
* previous Compose/config bundle versions where lifecycle tooling supports it

---

# 20. Failure and recovery expectations

## 20.1 Recoverable cases

The deployment SHOULD recover cleanly from:

* container restart
* Pi reboot
* temporary network loss to remote exporters
* Grafana restart
* Prometheus restart

## 20.2 Remote backend unreachable

If one or more remote exporters are unreachable:

* Prometheus continues running
* Grafana continues running
* unavailable targets are visible through Prometheus/Grafana health views

## 20.3 Config errors

Invalid generated config SHOULD be detected during bootstrap validation before normal service startup.

---

# 21. Required deliverables for implementation

The implementation following this spec MUST produce:

* a Pi 3 Debian setup directory structure
* a Docker Compose file
* a `.env.example`
* a remote target definition example
* Prometheus base template/config
* Grafana datasource provisioning
* Grafana dashboard provisioning structure
* optional host-side kiosk helper script or documented kiosk setup path
* README or runbook for first-time deployment

---

# 22. Recommended file layout for implementation

A recommended implementation layout is:

```text
monitoring/
  INSTALL.md
  collector/                          # deploy on every machine being monitored
    docker-compose.yml                # hwexp + node-exporter + Prometheus
    .env.example
    config/
      hwexp/
        hwexp.yaml
        mappings.yaml
        mappings.auto.yaml            # auto-generated on first run, not committed
      prometheus/
        prometheus.yml
  dashboard/                          # deploy on the Grafana host (anywhere on the network)
    docker-compose.yml                # Grafana only — PROMETHEUS_URL points at collector
    .env.example
    config/
      grafana/
        grafana.ini
        provisioning/
          datasources/
            prometheus.yaml
          dashboards/
            dashboards.yaml
    dashboards/
      panel/
      ops/
      discovery/
  scripts/
    start-kiosk.sh
```

---

# 23. First-time deployment flow

The recommended first-time deployment flow MUST be:

1. install Debian on Pi 3
2. install Docker Engine and Compose plugin
3. deploy the Pi dashboard-node bundle directory
4. copy `.env.example` to `.env`
5. set Grafana credentials and local settings
6. define remote backend/exporter targets
7. run config preparation/validation
8. run `docker compose up -d`
9. verify Grafana and Prometheus locally
10. configure optional kiosk/browser launch

---

# 24. Immediate follow-on document

After this spec, the next highest-value implementation document is a **Pi Dashboard Node Deployment Pack**, containing:

* concrete Docker Compose files
* `.env.example`
* Prometheus template files
* Grafana provisioning files
* sample remote target file
* first-time setup/runbook

That pack will turn this deployment spec into directly usable artifacts for the Pi 3 Debian installation.
